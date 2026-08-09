package icpbrasil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBaixarCadeiaCompleta(t *testing.T) {
	raizPEM := gerarCertPEMTeste(t, "AC Raiz Teste")
	subPEM := gerarCertPEMTeste(t, "AC Subordinada Teste")
	iniciarServidorFakeITI(t, raizPEM, subPEM)

	pemBytes, n, err := baixarCadeiaCompleta(context.Background())
	if err != nil {
		t.Fatalf("baixarCadeiaCompleta: %v", err)
	}
	if esperado := len(geracoesRaiz) + 1; n != esperado {
		t.Errorf("esperado %d certificados, veio %d", esperado, n)
	}

	pool, certs, ok := montarPool(pemBytes)
	if !ok {
		t.Fatal("montarPool não conseguiu decodificar o bundle baixado")
	}
	if pool == nil {
		t.Error("pool não deveria ser nil")
	}
	if len(certs) != n {
		t.Errorf("montarPool encontrou %d certificados, baixarCadeiaCompleta contou %d", len(certs), n)
	}
}

func TestBaixarCadeiaCompletaFalhaDeRede(t *testing.T) {
	urlRaizAntigo, urlBundleAntigo := urlRaizFmt, urlBundle
	urlRaizFmt = "http://127.0.0.1:1/RAIZ/ICP-Brasilv%s.crt" // porta inválida, falha de conexão garantida
	urlBundle = "http://127.0.0.1:1/ACcompactado.zip"
	defer func() { urlRaizFmt, urlBundle = urlRaizAntigo, urlBundleAntigo }()

	_, _, err := baixarCadeiaCompleta(context.Background())
	if err == nil {
		t.Fatal("esperava erro quando o servidor é inacessível, veio nil")
	}
}

func TestBaixarCadeiaCompletaStatusNaoOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	urlRaizAntigo, urlBundleAntigo := urlRaizFmt, urlBundle
	urlRaizFmt = srv.URL + "/RAIZ/ICP-Brasilv%s.crt"
	urlBundle = srv.URL + "/ACcompactado.zip"
	defer func() { urlRaizFmt, urlBundle = urlRaizAntigo, urlBundleAntigo }()

	_, _, err := baixarCadeiaCompleta(context.Background())
	if err == nil {
		t.Fatal("esperava erro quando o servidor devolve 404, veio nil")
	}
}
