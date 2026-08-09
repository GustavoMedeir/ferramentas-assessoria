package sigvalidator

import (
	"encoding/asn1"

	"github.com/smallstep/pkcs7"
)

// oidCarimboTempo é o id-aa-signatureTimeStampToken (RFC 3161 / CAdES-T) —
// atributo NÃO assinado que carrega o token de carimbo de tempo de uma
// Autoridade de Carimbo do Tempo (ACT).
var oidCarimboTempo = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 14}

// temCarimboTempo só detecta presença/ausência do atributo — validar o
// carimbo de verdade exigiria decodificar um CMS SignedData aninhado (o
// próprio TimeStampToken é, ele mesmo, uma assinatura CMS envolvendo um
// TSTInfo) e checar a cadeia de confiança da ACT até uma raiz, o que fica
// fora de escopo nesta fase (ver plano de implementação, risco R2).
//
// Ausência de carimbo não é motivo pra invalidar a assinatura — é só
// informação a mais mostrada na UI.
func temCarimboTempo(p7 *pkcs7.PKCS7) bool {
	if len(p7.Signers) == 0 {
		return false
	}
	for _, a := range p7.Signers[0].UnauthenticatedAttributes {
		if a.Type.Equal(oidCarimboTempo) {
			return true
		}
	}
	return false
}
