//go:build windows

// Package outlookmail abre um rascunho no Outlook Classic (desktop, via
// automação COM) já endereçado e com o corpo preenchido, mantendo a
// assinatura padrão configurada pelo próprio usuário no Outlook. Nunca
// envia — só chama Display(), deixando o rascunho pronto pra revisão.
package outlookmail

import (
	"fmt"
	"strings"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// AbrirRascunho abre o rascunho. remetente é o e-mail da conta Outlook a
// usar (Configurações > E-mail) — vazio deixa o Outlook escolher sozinho,
// o que é ambíguo quando há mais de uma conta logada.
func AbrirRascunho(destinatario, assunto, corpo, remetente string) error {
	if err := ole.CoInitialize(0); err != nil {
		return fmt.Errorf("iniciar COM: %w", err)
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return fmt.Errorf("Outlook Classic não encontrado — é necessário tê-lo instalado: %w", err)
	}
	defer unknown.Release()

	outlookApp, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("comunicar com o Outlook: %w", err)
	}
	defer outlookApp.Release()

	mailVariant, err := oleutil.CallMethod(outlookApp, "CreateItem", 0) // 0 = olMailItem
	if err != nil {
		return fmt.Errorf("criar e-mail no Outlook: %w", err)
	}
	mail := mailVariant.ToIDispatch()
	defer mail.Release()

	if _, err := oleutil.PutProperty(mail, "To", destinatario); err != nil {
		return fmt.Errorf("definir destinatário: %w", err)
	}
	if _, err := oleutil.PutProperty(mail, "Subject", assunto); err != nil {
		return fmt.Errorf("definir assunto: %w", err)
	}

	// SendUsingAccount ANTES do Display(): cada conta pode ter sua própria
	// assinatura padrão (Outlook > Opções > E-mail > Assinaturas > "Escolher
	// assinatura padrão" é por conta) — setar a conta antes garante que a
	// assinatura auto-inserida no Display() abaixo seja a da conta certa,
	// não a da conta padrão geral do Outlook.
	if remetente != "" {
		conta, err := encontrarConta(outlookApp, remetente)
		if err != nil {
			return err
		}
		defer conta.Release()
		if _, err := oleutil.PutProperty(mail, "SendUsingAccount", conta); err != nil {
			return fmt.Errorf("definir conta remetente: %w", err)
		}
	}

	// Display() ANTES de mexer no corpo: é o que faz o Outlook se
	// comportar como se o usuário tivesse clicado em "Novo E-mail" na UI
	// normal, inserindo a assinatura padrão. Setar o corpo antes do
	// Display() pula esse mecanismo e o rascunho fica sem assinatura.
	if _, err := oleutil.CallMethod(mail, "Display", false); err != nil {
		return fmt.Errorf("exibir o rascunho: %w", err)
	}

	// Nesse ponto o HTMLBody só tem a assinatura (corpo estava vazio antes
	// do Display) — lê de volta e prepend do nosso texto na frente dela.
	assinaturaVariant, err := oleutil.GetProperty(mail, "HTMLBody")
	if err != nil {
		return fmt.Errorf("ler assinatura padrão: %w", err)
	}
	novoCorpo := textoParaHTML(corpo) + assinaturaVariant.ToString()
	if _, err := oleutil.PutProperty(mail, "HTMLBody", novoCorpo); err != nil {
		return fmt.Errorf("preencher corpo do e-mail: %w", err)
	}

	return nil
}

// encontrarConta procura, entre as contas logadas no Outlook
// (Session.Accounts), a que tem o SmtpAddress igual a remetente. Erro
// explícito (não fallback silencioso) se não achar — o ponto inteiro dessa
// função é eliminar a ambiguidade de conta, não mascará-la.
func encontrarConta(outlookApp *ole.IDispatch, remetente string) (*ole.IDispatch, error) {
	sessionVar, err := oleutil.GetProperty(outlookApp, "Session")
	if err != nil {
		return nil, fmt.Errorf("acessar sessão do Outlook: %w", err)
	}
	session := sessionVar.ToIDispatch()
	defer session.Release()

	accountsVar, err := oleutil.GetProperty(session, "Accounts")
	if err != nil {
		return nil, fmt.Errorf("listar contas do Outlook: %w", err)
	}
	accounts := accountsVar.ToIDispatch()
	defer accounts.Release()

	countVar, err := oleutil.GetProperty(accounts, "Count")
	if err != nil {
		return nil, fmt.Errorf("contar contas do Outlook: %w", err)
	}
	count := int(countVar.Value().(int32))

	for i := 1; i <= count; i++ {
		itemVar, err := oleutil.GetProperty(accounts, "Item", i)
		if err != nil {
			continue
		}
		conta := itemVar.ToIDispatch()
		smtpVar, err := oleutil.GetProperty(conta, "SmtpAddress")
		if err == nil && strings.EqualFold(smtpVar.ToString(), remetente) {
			return conta, nil // não libera — quem chamou usa e libera depois
		}
		conta.Release()
	}
	return nil, fmt.Errorf("conta %q não encontrada entre as contas logadas no Outlook — confira o e-mail remetente em Configurações > E-mail", remetente)
}

// textoParaHTML escapa o texto puro e troca quebras de linha por <br> —
// necessário porque escrevemos em HTMLBody (pra preservar a assinatura em
// HTML), não em Body (texto puro).
func textoParaHTML(texto string) string {
	texto = strings.ReplaceAll(texto, "&", "&amp;")
	texto = strings.ReplaceAll(texto, "<", "&lt;")
	texto = strings.ReplaceAll(texto, ">", "&gt;")
	texto = strings.ReplaceAll(texto, "\n", "<br>")
	return `<div style="font-family:Calibri,Arial,sans-serif;font-size:11pt">` + texto + "</div>"
}
