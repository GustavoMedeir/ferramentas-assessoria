package selfupdate

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaisNova(t *testing.T) {
	casos := []struct {
		atual, remota string
		esperado      bool
	}{
		{"2.1.0", "2.2.0", true},
		{"v2.1.0", "v2.2.0", true},
		{"2.1.0", "2.1.0", false},
		{"2.2.0", "2.1.0", false},
		{"2.1.0", "3.0.0", true},
		{"2.1.9", "2.2.0", true},
		{"2.1.0", "2.1.1", true},
		{"dev", "2.1.0", true}, // "dev" vira 0.0.0 — mas app.go nunca chama MaisNova quando Version == "dev"
	}
	for _, c := range casos {
		if got := MaisNova(c.atual, c.remota); got != c.esperado {
			t.Errorf("MaisNova(%q, %q) = %v, esperado %v", c.atual, c.remota, got, c.esperado)
		}
	}
}

func servidorFakeGitHub(t *testing.T, resp releaseGitHub) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("requisição sem User-Agent — a API do GitHub recusaria isso de verdade")
		}
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)

	antigo := apiBaseURL
	apiBaseURL = srv.URL
	t.Cleanup(func() { apiBaseURL = antigo })
}

func TestUltimaVersaoComAssetsCompletos(t *testing.T) {
	servidorFakeGitHub(t, releaseGitHub{
		TagName: "v2.2.0",
		Body:    "Corrige bug X.",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "FerramentasAssessoria.exe", BrowserDownloadURL: "https://example.com/app.exe"},
			{Name: "FerramentasAssessoria.exe.sha256", BrowserDownloadURL: "https://example.com/app.exe.sha256"},
		},
	})

	info, encontrado, err := UltimaVersao(context.Background())
	if err != nil {
		t.Fatalf("UltimaVersao: %v", err)
	}
	if !encontrado {
		t.Fatal("esperava encontrado=true com os dois assets presentes")
	}
	if info.Versao != "v2.2.0" || info.Notas != "Corrige bug X." {
		t.Errorf("info inesperado: %+v", info)
	}
	if info.URLExe != "https://example.com/app.exe" || info.URLChecksum != "https://example.com/app.exe.sha256" {
		t.Errorf("URLs de asset inesperadas: %+v", info)
	}
}

// Desde que a release passou a levar o pacote do macOS junto, ela tem dois
// arquivos .sha256. O do Windows precisa casar com o .exe — pegar o do Mac
// faria a conferência falhar e a atualização morrer como "arquivo
// corrompido" pra todo mundo.
func TestUltimaVersaoEscolheOChecksumDoExeQuandoHaAssetDoMac(t *testing.T) {
	servidorFakeGitHub(t, releaseGitHub{
		TagName: "v1.02.00",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			// O do Mac vem ANTES de propósito: com a varredura antiga, que
			// aceitava qualquer .sha256, ele venceria.
			{Name: "FerramentasAssessoria-1.02.00-macOS.zip", BrowserDownloadURL: "https://example.com/mac.zip"},
			{Name: "FerramentasAssessoria-1.02.00-macOS.zip.sha256", BrowserDownloadURL: "https://example.com/mac.zip.sha256"},
			{Name: "FerramentasAssessoria-1.02.00.exe", BrowserDownloadURL: "https://example.com/app.exe"},
			{Name: "FerramentasAssessoria-1.02.00.exe.sha256", BrowserDownloadURL: "https://example.com/app.exe.sha256"},
		},
	})

	info, encontrado, err := UltimaVersao(context.Background())
	if err != nil {
		t.Fatalf("UltimaVersao: %v", err)
	}
	if !encontrado {
		t.Fatal("esperava encontrado=true")
	}
	if info.URLExe != "https://example.com/app.exe" {
		t.Errorf("URLExe = %q, esperado o .exe do Windows", info.URLExe)
	}
	if info.URLChecksum != "https://example.com/app.exe.sha256" {
		t.Errorf("URLChecksum = %q, esperado o checksum DO EXE (não o do zip do Mac)", info.URLChecksum)
	}
}

func TestUltimaVersaoSemAssetDeChecksum(t *testing.T) {
	// Release publicada pela metade (ex.: upload do .exe terminou mas o do
	// .sha256 ainda não) não deve ser oferecida como atualização.
	servidorFakeGitHub(t, releaseGitHub{
		TagName: "v2.2.0",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "FerramentasAssessoria.exe", BrowserDownloadURL: "https://example.com/app.exe"},
		},
	})

	_, encontrado, err := UltimaVersao(context.Background())
	if err != nil {
		t.Fatalf("UltimaVersao: %v", err)
	}
	if encontrado {
		t.Error("esperava encontrado=false sem o asset .sha256")
	}
}

func TestUltimaVersaoStatusNaoOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	antigo := apiBaseURL
	apiBaseURL = srv.URL
	defer func() { apiBaseURL = antigo }()

	_, _, err := UltimaVersao(context.Background())
	if err == nil {
		t.Fatal("esperava erro com status 404")
	}
}

// O sha256sum do runner Windows escreve em modo binário: "hash *arquivo"
// (com asterisco), diferente do "hash  arquivo" de duas casas do shasum do
// macOS. Os dois precisam funcionar — é o CI que gera esses arquivos.
func TestBaixarChecksumFormatoBinarioDoSha256sum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("c8cb9c9763d446551c5bed892aa287d0b73238d25bf26216d4a723b8c5abc6f8 *FerramentasAssessoria-1.02.00.exe\n"))
	}))
	defer srv.Close()

	dados, err := baixarChecksum(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("baixarChecksum: %v", err)
	}
	if len(dados) != 32 {
		t.Fatalf("esperado hash de 32 bytes, veio %d — o asterisco do modo binário provavelmente entrou no hash", len(dados))
	}
	if got := hex.EncodeToString(dados); got != "c8cb9c9763d446551c5bed892aa287d0b73238d25bf26216d4a723b8c5abc6f8" {
		t.Errorf("hash decodificado = %q", got)
	}
}

func TestBaixarChecksum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// formato do `sha256sum`: hash + espaço + nome do arquivo
		w.Write([]byte("a3f5b1c2d4e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7081920a1b2c3d4e5f6  app.exe\n"))
	}))
	defer srv.Close()

	dados, err := baixarChecksum(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("baixarChecksum: %v", err)
	}
	if len(dados) != 32 { // sha256 = 32 bytes
		t.Errorf("esperado hash de 32 bytes, veio %d", len(dados))
	}
}
