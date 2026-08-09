// Package whatsapp monta o link "wa.me" usado pelo botão "Enviar WhatsApp"
// da aba Rentabilidade — abre o WhatsApp (Desktop ou Web, o que estiver
// instalado) com a mensagem já preenchida na conversa do cliente; quem
// aperta "Enviar" lá dentro é o assessor, não o app.
package whatsapp

import (
	"fmt"
	"net/url"
	"strings"
)

// MontarLink monta a URL "https://wa.me/<telefone>?text=<mensagem>". Aceita
// telefone em qualquer formatação (parênteses, espaços, traços, "+55" etc.)
// — só os dígitos importam. Números de 10 ou 11 dígitos (DDD + número, sem
// código do país) são tratados como brasileiros e ganham o prefixo 55
// automaticamente; números que já vierem com código do país (12+ dígitos)
// são usados como estão.
func MontarLink(telefone, mensagem string) (string, error) {
	digitos := apenasDigitos(telefone)
	if len(digitos) < 8 {
		return "", fmt.Errorf("telefone inválido: %q", telefone)
	}
	if len(digitos) == 10 || len(digitos) == 11 {
		digitos = "55" + digitos
	}

	v := url.Values{}
	v.Set("text", mensagem)
	return fmt.Sprintf("https://wa.me/%s?%s", digitos, v.Encode()), nil
}

func apenasDigitos(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
