package sigvalidator

import (
	"crypto/x509"
	"encoding/asn1"
	"strings"
)

// OIDs do padrão ICP-Brasil (DOC-ICP-04 — Atribuição de OID na ICP-Brasil)
// pros dados de identidade embutidos no otherName da extensão
// SubjectAlternativeName. O crypto/x509 do Go não expõe otherName (só
// popula DNSNames/EmailAddresses/etc a partir da SAN), por isso o parsing
// abaixo é manual via encoding/asn1 sobre a extensão bruta.
var (
	oidSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
	// OID_PF_DADOS_TITULAR: NASCIMENTO(8) + CPF(11) + NIS(11) + RG(15) +
	// ORGAO_EXPEDIDOR(5), concatenados sem separador numa única string.
	// Layout documentado publicamente (DOC-ICP-04); só o campo CPF
	// (posições 8:19) é extraído aqui — é o único que este feature precisa.
	oidICPBrasilPF = asn1.ObjectIdentifier{2, 16, 76, 1, 3, 1}
	// OID_PJ_CNPJ: contém só o CNPJ da pessoa jurídica titular.
	oidICPBrasilPJCNPJ = asn1.ObjectIdentifier{2, 16, 76, 1, 3, 3}
)

// extrairIdentidade lê nome e CPF/CNPJ do signatário a partir do
// certificado. Nunca dá panic em formato inesperado — na pior hipótese
// devolve TipoPessoa "desconhecido" e CPF/CNPJ vazios, sem abortar a
// validação por causa disso (o layout exato do otherName pode variar
// entre ACs; isso é tratado como best-effort, não como requisito rígido).
func extrairIdentidade(cert *x509.Certificate) (s Signatario) {
	defer func() {
		if recover() != nil {
			s = Signatario{TipoPessoa: "desconhecido"}
		}
	}()

	s.Nome = cert.Subject.CommonName
	s.TipoPessoa = "desconhecido"

	if pf, ok := lerOutroNomeICPBrasil(cert, oidICPBrasilPF); ok && len(pf) >= 19 {
		if cpf := somenteDigitos(pf[8:19]); len(cpf) == 11 {
			s.CPF = cpf
			s.TipoPessoa = "fisica"
		}
	}
	if pj, ok := lerOutroNomeICPBrasil(cert, oidICPBrasilPJCNPJ); ok {
		if cnpj := somenteDigitos(pj); len(cnpj) == 14 {
			s.CNPJ = cnpj
			s.TipoPessoa = "juridica"
		}
	}
	return s
}

func somenteDigitos(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// lerOutroNomeICPBrasil busca, na extensão SubjectAlternativeName do
// certificado, um otherName cujo type-id seja oid, e devolve o conteúdo
// textual bruto (quem chama decide como fatiar os campos internos).
//
// Decodificação manual porque GeneralName é um CHOICE com tags de
// contexto: "otherName [0]" é IMPLICIT no módulo do RFC 5280 (o tag
// SEQUENCE de OtherName vira direto o tag de contexto [0]), mas o campo
// "value [0]" dentro dele é EXPLICIT — por isso são duas camadas de
// unwrap (uma pro otherName em si, outra pro valor).
func lerOutroNomeICPBrasil(cert *x509.Certificate, oid asn1.ObjectIdentifier) (valor string, ok bool) {
	var extValue []byte
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(oidSubjectAltName) {
			extValue = ext.Value
			break
		}
	}
	if extValue == nil {
		return "", false
	}

	var nomes []asn1.RawValue
	if _, err := asn1.Unmarshal(extValue, &nomes); err != nil {
		return "", false
	}

	for _, nome := range nomes {
		if nome.Class != asn1.ClassContextSpecific || nome.Tag != 0 {
			continue // não é a alternativa otherName da CHOICE
		}

		var tipoID asn1.ObjectIdentifier
		rest, err := asn1.Unmarshal(nome.Bytes, &tipoID)
		if err != nil || !tipoID.Equal(oid) {
			continue
		}

		var explicito asn1.RawValue
		if _, err := asn1.Unmarshal(rest, &explicito); err != nil {
			continue
		}
		var interno asn1.RawValue
		if _, err := asn1.Unmarshal(explicito.Bytes, &interno); err != nil {
			continue
		}
		if interno.IsCompound {
			continue // esperado um tipo primitivo (string) — formato inesperado, ignora
		}
		return string(interno.Bytes), true
	}
	return "", false
}
