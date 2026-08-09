package icpbrasil

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNovoStoreFallbackEmbutido(t *testing.T) {
	dir := t.TempDir()
	s, err := NovoStore(filepath.Join(dir, "nao-existe"))
	if err != nil {
		t.Fatalf("NovoStore: %v", err)
	}
	info := s.Info()
	if info.Origem != "embutida" {
		t.Errorf("esperado origem embutida, veio %q", info.Origem)
	}
	if info.NumCertificados == 0 {
		t.Error("esperado ao menos 1 certificado no snapshot embutido")
	}
	if s.Pool() == nil {
		t.Error("Pool() não deveria ser nil")
	}
	if len(s.Certificados()) != info.NumCertificados {
		t.Errorf("Certificados() devolveu %d, Info.NumCertificados diz %d", len(s.Certificados()), info.NumCertificados)
	}
}

func TestStoreRoundtripDisco(t *testing.T) {
	dir := t.TempDir()
	pemBytes := gerarCertPEMTeste(t, "AC Teste Disco")
	agora := time.Now().UTC().Truncate(time.Second)
	if err := salvarEmDisco(dir, pemBytes, agora, 1); err != nil {
		t.Fatalf("salvarEmDisco: %v", err)
	}

	s, err := NovoStore(dir)
	if err != nil {
		t.Fatalf("NovoStore: %v", err)
	}
	info := s.Info()
	if info.Origem != "baixada" {
		t.Errorf("esperado origem baixada (carregada do disco), veio %q", info.Origem)
	}
	if info.NumCertificados != 1 {
		t.Errorf("esperado 1 certificado, veio %d", info.NumCertificados)
	}
	if !info.AtualizadoEm.Equal(agora) {
		t.Errorf("esperado AtualizadoEm %v, veio %v", agora, info.AtualizadoEm)
	}
}

func TestStoreAtualizarAgoraPersisteEDisponibilizaPool(t *testing.T) {
	dir := t.TempDir()
	s, err := NovoStore(dir)
	if err != nil {
		t.Fatalf("NovoStore: %v", err)
	}
	antesInfo := s.Info()
	if antesInfo.Origem != "embutida" {
		t.Fatalf("pré-condição: esperava origem embutida antes de atualizar, veio %q", antesInfo.Origem)
	}

	raizPEM := gerarCertPEMTeste(t, "AC Raiz Atualizada")
	subPEM := gerarCertPEMTeste(t, "AC Subordinada Atualizada")
	iniciarServidorFakeITI(t, raizPEM, subPEM)

	depoisInfo, err := s.AtualizarAgora(context.Background())
	if err != nil {
		t.Fatalf("AtualizarAgora: %v", err)
	}
	if depoisInfo.Origem != "baixada" {
		t.Errorf("esperado origem baixada após AtualizarAgora, veio %q", depoisInfo.Origem)
	}
	if esperado := len(geracoesRaiz) + 1; depoisInfo.NumCertificados != esperado {
		t.Errorf("esperado %d certificados, veio %d", esperado, depoisInfo.NumCertificados)
	}
	if s.Info().AtualizadoEm != depoisInfo.AtualizadoEm {
		t.Error("Store.Info() não refletiu a atualização")
	}

	// Uma segunda instância criada a partir da mesma pasta deve carregar o
	// que foi persistido em disco pela primeira.
	s2, err := NovoStore(dir)
	if err != nil {
		t.Fatalf("NovoStore (segunda instância): %v", err)
	}
	if s2.Info().NumCertificados != depoisInfo.NumCertificados {
		t.Errorf("segunda instância não carregou o que foi persistido: %d != %d", s2.Info().NumCertificados, depoisInfo.NumCertificados)
	}
}

// TestStoreLeituraConcorrenteDuranteAtualizacao existe pra ser rodado com
// -race: várias goroutines lendo Pool()/Certificados()/Info() enquanto uma
// atualização em background troca o conteúdo do Store.
func TestStoreLeituraConcorrenteDuranteAtualizacao(t *testing.T) {
	dir := t.TempDir()
	s, err := NovoStore(dir)
	if err != nil {
		t.Fatalf("NovoStore: %v", err)
	}

	raizPEM := gerarCertPEMTeste(t, "AC Raiz Race")
	subPEM := gerarCertPEMTeste(t, "AC Subordinada Race")
	iniciarServidorFakeITI(t, raizPEM, subPEM)

	var wg sync.WaitGroup
	pararLeitura := make(chan struct{})

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-pararLeitura:
					return
				default:
					_ = s.Pool()
					_ = s.Certificados()
					_ = s.Info()
				}
			}
		}()
	}

	if _, err := s.AtualizarAgora(context.Background()); err != nil {
		t.Errorf("AtualizarAgora: %v", err)
	}
	close(pararLeitura)
	wg.Wait()
}
