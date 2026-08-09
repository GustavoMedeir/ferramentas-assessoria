package icpbrasil

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// gerarCertPEMTeste cria um certificado autoassinado sintético (não é um
// certificado ICP-Brasil real — só serve pra exercitar o parsing/pool sem
// depender de rede ou de um arquivo de assinatura real).
func gerarCertPEMTeste(t *testing.T, cn string) []byte {
	t.Helper()
	chave, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gerar chave: %v", err)
	}
	modelo := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, modelo, modelo, &chave.PublicKey, chave)
	if err != nil {
		t.Fatalf("criar certificado de teste: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

// iniciarServidorFakeITI sobe um httptest.Server que imita o repositório do
// ITI: todas as gerações de AC-Raiz devolvem raizPEM, e o bundle zip contém
// uma única entrada com subPEM. Redireciona urlRaizFmt/urlBundle pra ele e
// devolve uma função de limpeza que restaura os valores originais.
func iniciarServidorFakeITI(t *testing.T, raizPEM, subPEM []byte) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for _, v := range geracoesRaiz {
		v := v
		mux.HandleFunc("/RAIZ/ICP-Brasilv"+v+".crt", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(raizPEM)
		})
	}
	mux.HandleFunc("/ACcompactado.zip", func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		f, err := zw.Create("AC-Subordinada-Teste.crt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := f.Write(subPEM); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := zw.Close(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(buf.Bytes())
	})
	srv := httptest.NewServer(mux)

	urlRaizAntigo, urlBundleAntigo := urlRaizFmt, urlBundle
	urlRaizFmt = srv.URL + "/RAIZ/ICP-Brasilv%s.crt"
	urlBundle = srv.URL + "/ACcompactado.zip"
	t.Cleanup(func() {
		urlRaizFmt, urlBundle = urlRaizAntigo, urlBundleAntigo
		srv.Close()
	})
	return srv
}
