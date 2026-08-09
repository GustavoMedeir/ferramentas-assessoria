// Package clientdb lê a base de clientes (CSV com colunas de código, nome e,
// opcionalmente, telefone e e-mail, em qualquer ordem, separadas por ";" ou
// ",") — compartilhada pelas abas Rentabilidade e Gerador de E-mails de
// Ordem.
package clientdb

import (
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Cliente é o que a base de clientes sabe sobre um código: nome e,
// opcionalmente, telefone (usado pra montar o link do WhatsApp) e e-mail
// (usado pra pré-preencher o destinatário no Gerador de E-mails de Ordem).
type Cliente struct {
	Nome     string
	Telefone string
	Email    string
}

// normalizarCabecalho remove acentos (via NFD + filtro de marcas
// combinantes) e baixa a caixa, pra reconhecer "Código", "codigo", "CÓDIGO"
// etc. como a mesma coluna.
func normalizarCabecalho(texto string) string {
	texto = strings.ToLower(strings.TrimSpace(texto))
	decomposto := norm.NFD.String(texto)
	var sb strings.Builder
	for _, r := range decomposto {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// CarregarBaseClientes lê um CSV com colunas de código, nome e (opcional)
// telefone, e devolve {codigo: Cliente}. Se não reconhecer um cabeçalho com
// as palavras de código/nome, assume que a coluna 0 é o código e a 1 é o
// nome, tratando a primeira linha como dado — nesse caso não há coluna de
// telefone.
func CarregarBaseClientes(caminho string) (map[string]Cliente, error) {
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return nil, err
	}
	// Python usa utf-8-sig — descarta o BOM UTF-8 se presente.
	texto := strings.TrimPrefix(string(dados), "\uFEFF")

	var linhas []string
	for _, l := range strings.Split(texto, "\n") {
		l = strings.TrimRight(l, "\r")
		if strings.TrimSpace(l) != "" {
			linhas = append(linhas, l)
		}
	}
	if len(linhas) == 0 {
		return map[string]Cliente{}, nil
	}

	delim := ","
	if strings.Count(linhas[0], ";") > strings.Count(linhas[0], ",") {
		delim = ";"
	}

	cabecalho := strings.Split(linhas[0], delim)
	idxCodigo, idxNome, idxTelefone, idxEmail := -1, -1, -1, -1
	for i, c := range cabecalho {
		nc := normalizarCabecalho(c)
		if idxCodigo == -1 && (strings.Contains(nc, "codigo") || strings.Contains(nc, "conta")) {
			idxCodigo = i
		}
		if idxNome == -1 && strings.Contains(nc, "nome") {
			idxNome = i
		}
		if idxTelefone == -1 && (strings.Contains(nc, "telefone") || strings.Contains(nc, "celular") || strings.Contains(nc, "whatsapp") || strings.Contains(nc, "fone")) {
			idxTelefone = i
		}
		// "mail" pega "email" e "e-mail" (o hífen não é marca combinante,
		// sobrevive à normalização) sem colidir com nome/codigo/telefone.
		if idxEmail == -1 && strings.Contains(nc, "mail") {
			idxEmail = i
		}
	}

	inicio := 1
	if idxCodigo == -1 || idxNome == -1 {
		idxCodigo, idxNome, idxTelefone, idxEmail, inicio = 0, 1, -1, -1, 0
	}

	maxIdx := idxCodigo
	if idxNome > maxIdx {
		maxIdx = idxNome
	}
	if idxTelefone > maxIdx {
		maxIdx = idxTelefone
	}
	if idxEmail > maxIdx {
		maxIdx = idxEmail
	}

	base := map[string]Cliente{}
	for _, linha := range linhas[inicio:] {
		colunas := strings.Split(linha, delim)
		if len(colunas) <= maxIdx {
			continue
		}
		codigo := strings.TrimSpace(colunas[idxCodigo])
		if codigo == "" {
			continue
		}
		cliente := Cliente{Nome: strings.TrimSpace(colunas[idxNome])}
		if idxTelefone != -1 {
			cliente.Telefone = strings.TrimSpace(colunas[idxTelefone])
		}
		if idxEmail != -1 {
			cliente.Email = strings.TrimSpace(colunas[idxEmail])
		}
		base[codigo] = cliente
	}
	return base, nil
}
