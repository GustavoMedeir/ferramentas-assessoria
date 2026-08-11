package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestNomeArquivoVersionado(t *testing.T) {
	casos := map[string]string{
		"v1.04.06": "FerramentasAssessoria-1.04.06.exe",
		"1.04.06":  "FerramentasAssessoria-1.04.06.exe", // "v" é opcional, mesmo padrão de MaisNova
	}
	for versao, esperado := range casos {
		if got := nomeArquivoVersionado(versao); got != esperado {
			t.Errorf("nomeArquivoVersionado(%q) = %q, esperado %q", versao, got, esperado)
		}
	}
}

// TestBaixarEGravar cobre o caminho novo (baixa em Downloads com nome
// versionado, nunca sobrescreve o executável em uso) — substitui os testes
// antigos que exercitavam selfupdate.Apply() (biblioteca removida junto
// com a troca in-place pela troca via Downloads, ver selfupdate.go).
func TestBaixarEGravar(t *testing.T) {
	conteudoExe := []byte("conteudo de mentira do executavel novo")
	soma := sha256.Sum256(conteudoExe)

	mux := http.NewServeMux()
	mux.HandleFunc("/app.exe", func(w http.ResponseWriter, r *http.Request) { w.Write(conteudoExe) })
	mux.HandleFunc("/app.exe.sha256", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(hex.EncodeToString(soma[:]) + "  FerramentasAssessoria-1.04.06.exe\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	info := ReleaseInfo{Versao: "v1.04.06", URLExe: srv.URL + "/app.exe", URLChecksum: srv.URL + "/app.exe.sha256"}

	caminho, err := baixarEGravar(context.Background(), info, dir)
	if err != nil {
		t.Fatalf("baixarEGravar: %v", err)
	}
	esperado := filepath.Join(dir, "FerramentasAssessoria-1.04.06.exe")
	if caminho != esperado {
		t.Errorf("caminho = %q, esperado %q", caminho, esperado)
	}
	dados, err := os.ReadFile(caminho)
	if err != nil {
		t.Fatalf("ler arquivo baixado: %v", err)
	}
	if string(dados) != string(conteudoExe) {
		t.Error("conteúdo gravado não bate com o baixado")
	}
}

func TestBaixarEGravarRejeitaChecksumDivergente(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/app.exe", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("conteudo real")) })
	mux.HandleFunc("/app.exe.sha256", func(w http.ResponseWriter, r *http.Request) {
		// hash de outra coisa qualquer, não bate com o conteúdo acima
		soma := sha256.Sum256([]byte("outra coisa"))
		w.Write([]byte(hex.EncodeToString(soma[:])))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	info := ReleaseInfo{Versao: "v1.04.06", URLExe: srv.URL + "/app.exe", URLChecksum: srv.URL + "/app.exe.sha256"}

	if _, err := baixarEGravar(context.Background(), info, dir); err == nil {
		t.Fatal("esperava erro de checksum divergente")
	}
	if entradas, _ := os.ReadDir(dir); len(entradas) != 0 {
		t.Error("não deveria ter gravado nada no destino quando o checksum não bate")
	}
}

func TestLimparVersoesAntigas(t *testing.T) {
	dir := t.TempDir()
	emUso := filepath.Join(dir, "FerramentasAssessoria-1.04.06.exe")
	antiga := filepath.Join(dir, "FerramentasAssessoria-1.04.05.exe")
	outroArquivo := filepath.Join(dir, "nao-mexe.txt")

	for _, caminho := range []string{emUso, antiga, outroArquivo} {
		if err := os.WriteFile(caminho, []byte("x"), 0644); err != nil {
			t.Fatalf("preparar %s: %v", caminho, err)
		}
	}

	limparVersoesAntigas(dir, emUso)

	if _, err := os.Stat(emUso); err != nil {
		t.Error("executável em uso não deveria ter sido apagado")
	}
	if _, err := os.Stat(antiga); !os.IsNotExist(err) {
		t.Error("versão antiga deveria ter sido apagada")
	}
	if _, err := os.Stat(outroArquivo); err != nil {
		t.Error("arquivo sem o prefixo do app não deveria ter sido tocado")
	}
}
