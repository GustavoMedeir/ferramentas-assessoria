//go:build !windows

// Fora do Windows não existe o Outlook Classic nem a automação COM que o
// arquivo principal usa. O pacote continua compilando (a versão macOS
// depende disso) e a função devolve um erro explicativo — que na prática
// não chega a ser exibido, porque a interface esconde o botão "Abrir no
// Outlook" quando não está no Windows (ver Plataforma() em app.go e o uso
// em tabs/emails.js). O erro existe como rede de segurança, caso alguém
// chame o binding por outro caminho.
package outlookmail

import "fmt"

func AbrirRascunho(destinatario, assunto, corpo, remetente string) error {
	return fmt.Errorf("abrir rascunho no Outlook só está disponível no Windows — use o botão \"Copiar texto\" e cole no seu e-mail")
}
