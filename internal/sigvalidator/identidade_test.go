package sigvalidator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"testing"
	"time"
)

// certificadoComOtherName monta, à mão, um certificado X.509 autoassinado
// com uma extensão SubjectAlternativeName contendo um único otherName
// (type-id oid, valor UTF8String texto) — reproduz a estrutura ASN.1 real
// (otherName IMPLICIT [0], value EXPLICIT [0] ANY) sem depender de nenhum
// certificado ICP-Brasil real, só pra validar o parser em identidade.go.
func certificadoComOtherName(t *testing.T, cn string, oid asn1.ObjectIdentifier, texto string) *x509.Certificate {
	t.Helper()

	innerTLV, err := asn1.MarshalWithParams(texto, "utf8")
	if err != nil {
		t.Fatalf("marshal do valor interno: %v", err)
	}
	explicitTLV, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: innerTLV})
	if err != nil {
		t.Fatalf("marshal do wrapper explicit: %v", err)
	}
	oidTLV, err := asn1.Marshal(oid)
	if err != nil {
		t.Fatalf("marshal do OID: %v", err)
	}
	otherNameContent := append(append([]byte{}, oidTLV...), explicitTLV...)
	otherNameRaw := asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: otherNameContent}

	sanValue, err := asn1.Marshal([]asn1.RawValue{otherNameRaw})
	if err != nil {
		t.Fatalf("marshal da SAN: %v", err)
	}

	chave, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gerar chave: %v", err)
	}
	modelo := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		ExtraExtensions: []pkix.Extension{
			{Id: oidSubjectAltName, Critical: false, Value: sanValue},
		},
	}
	der, err := x509.CreateCertificate(rand.Reader, modelo, modelo, &chave.PublicKey, chave)
	if err != nil {
		t.Fatalf("criar certificado: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parsear certificado: %v", err)
	}
	return cert
}

func TestExtrairIdentidadePessoaFisica(t *testing.T) {
	// Layout DOC-ICP-04 (OID_PF_DADOS_TITULAR): NASCIMENTO(8)+CPF(11)+NIS(11)+RG(15)+ORGAO(5)
	valor := "17121953" + "58765136012" + "42586038731" + "000000000000000" + "SSPRS"
	cert := certificadoComOtherName(t, "Fulano de Tal", oidICPBrasilPF, valor)

	s := extrairIdentidade(cert)
	if s.TipoPessoa != "fisica" {
		t.Errorf("esperado TipoPessoa fisica, veio %q", s.TipoPessoa)
	}
	if s.CPF != "58765136012" {
		t.Errorf("esperado CPF 58765136012, veio %q", s.CPF)
	}
	if s.Nome != "Fulano de Tal" {
		t.Errorf("esperado Nome do CommonName, veio %q", s.Nome)
	}
}

func TestExtrairIdentidadePessoaJuridica(t *testing.T) {
	cert := certificadoComOtherName(t, "Empresa Teste LTDA:12345678000199", oidICPBrasilPJCNPJ, "12345678000199")

	s := extrairIdentidade(cert)
	if s.TipoPessoa != "juridica" {
		t.Errorf("esperado TipoPessoa juridica, veio %q", s.TipoPessoa)
	}
	if s.CNPJ != "12345678000199" {
		t.Errorf("esperado CNPJ 12345678000199, veio %q", s.CNPJ)
	}
}

func TestExtrairIdentidadeSemSAN(t *testing.T) {
	chave, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gerar chave: %v", err)
	}
	modelo := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Sem SAN"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, modelo, modelo, &chave.PublicKey, chave)
	if err != nil {
		t.Fatalf("criar certificado: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parsear certificado: %v", err)
	}

	s := extrairIdentidade(cert)
	if s.TipoPessoa != "desconhecido" {
		t.Errorf("esperado TipoPessoa desconhecido sem SAN, veio %q", s.TipoPessoa)
	}
	if s.CPF != "" || s.CNPJ != "" {
		t.Errorf("esperado CPF/CNPJ vazios sem SAN, veio CPF=%q CNPJ=%q", s.CPF, s.CNPJ)
	}
	if s.Nome != "Sem SAN" {
		t.Errorf("Nome deveria vir do CommonName mesmo sem SAN, veio %q", s.Nome)
	}
}

func TestExtrairIdentidadeCPFCurtoDemaisNaoQuebra(t *testing.T) {
	// valor mais curto que o esperado (formato inesperado/corrompido) —
	// nunca deve dar panic, só degradar pra "não achou".
	cert := certificadoComOtherName(t, "Curto", oidICPBrasilPF, "123")

	s := extrairIdentidade(cert)
	if s.CPF != "" {
		t.Errorf("esperado CPF vazio pra valor curto demais, veio %q", s.CPF)
	}
	if s.TipoPessoa != "desconhecido" {
		t.Errorf("esperado TipoPessoa desconhecido, veio %q", s.TipoPessoa)
	}
}
