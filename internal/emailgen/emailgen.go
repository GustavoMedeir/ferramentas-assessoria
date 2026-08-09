// Package emailgen monta o texto padronizado de e-mail de ordem (Renda
// Fixa, Tesouro Direto, Fundos, COE, Ofertas, Subscrição, Societárias,
// Clubes, Ações/Opções/Termo, Movimentação, Compromissadas, Carteira
// Automatizada), com uma ou mais operações por e-mail. Porte de
// rentabilidade_app_v2.py.
package emailgen

import (
	"fmt"
	"strings"
)

// Field é um campo de formulário de uma categoria.
type Field struct {
	Key         string
	Label       string
	Placeholder string   // usado quando Type == ""
	Type        string   // "" (texto) ou "select"
	Options     []string // só quando Type == "select"

	// Optional: quando true e o campo fica vazio, ResolveValores devolve
	// "" (em vez do fallback "[Label]") e o Body da categoria omite a
	// linha correspondente.
	Optional bool

	// Moeda: quando true, o campo só aceita um valor monetário puro (nunca
	// quantidade, taxa ou texto livre) — ResolveValores prefixa "R$ "
	// automaticamente se o usuário não tiver digitado (evita ter que
	// escrever "R$" na mão toda vez).
	Moeda bool
}

// Category é um modelo de operação (produto + tipo de operação).
type Category struct {
	ID    string
	Group string // "produto" no app original
	Label string // "operação" no app original

	// Tipo ("aplicacao" | "resgate" | "outro") é metadado herdado do app
	// original — hoje não é lido por nenhuma lógica além da própria
	// definição da categoria (nem intro, nem corpo, nem agrupamento).
	Tipo string

	IntroFrase string
	Fields     []Field
	Body       func(v map[string]string) string
	Anexo      string // "" se não houver aviso de anexo
	Fechamento string

	// SoPadronizado: a categoria só existe no modo padronizado (um produto
	// por e-mail) — o modo livre não a oferece nem aceita.
	SoPadronizado bool

	// EmailCompleto: quando definido, a categoria tem um texto de e-mail
	// próprio do início ao fim (saudação, corpo e fechamento), sem passar
	// pela montagem padrão intro/corpo/fechamento. Body fica nil.
	EmailCompleto func(codigo, nome string, v map[string]string) string
}

// Produtos lista os grupos na ordem em que aparecem em Categories (dita a
// ordem do combobox de produto).
var Produtos []string

// RVGroups são os grupos de Renda Variável: os únicos que usam a frase de
// abertura específica da operação; todo o resto usa a abertura genérica +
// cabeçalho "Produto - Operação" no corpo.
var RVGroups = map[string]bool{
	"Ações, Futuros e FIIs": true,
	"Opções":                true,
	"Operações a Termo":     true,
}

func init() {
	vistos := map[string]bool{}
	for _, c := range Categories {
		if !vistos[c.Group] {
			vistos[c.Group] = true
			Produtos = append(Produtos, c.Group)
		}
	}
}

// InfoEstruturadas é o texto mostrado no link "Preciso registrar uma
// Operação Estruturada".
const InfoEstruturadas = "Os modelos de ordem para Operações Estruturadas foram elaborados separadamente, devido " +
	"à complexidade das operações, aos riscos envolvidos e às suas especificidades. Além disso, " +
	"os modelos de operações estruturadas ofertadas estão dentro do cotizador e são enviados de " +
	"forma automatizada ao cliente via push.\n\n" +
	"Para qualquer operação diferente das suportadas pelo cotizador, contate o time de Operações " +
	"Estruturadas pelo HUB do assessor (menu Renda Variável > Produtos Estruturados), ou pelos " +
	"e-mails controle.produtosestruturados@xpi.com.br / auditoriadeordens@xpi.com.br.\n\n" +
	"Cautela redobrada é recomendada no registro de ordem e na realização desse tipo de operação. " +
	"Informe sempre o cliente de forma clara sobre toda a estrutura da operação e os riscos " +
	"envolvidos."

// CategoriaPorGrupoLabel busca a categoria por grupo (produto) + rótulo
// (operação). Devolve nil se não encontrar.
func CategoriaPorGrupoLabel(grupo, label string) *Category {
	for i := range Categories {
		if Categories[i].Group == grupo && Categories[i].Label == label {
			return &Categories[i]
		}
	}
	return nil
}

// ResolveValores aplica o fallback "[Label]" pros campos vazios (mesma
// regra do app original: v[campo] = valor ou f"[{campo['label']}]").
func (c *Category) ResolveValores(brutos map[string]string) map[string]string {
	resolvido := make(map[string]string, len(c.Fields))
	for _, f := range c.Fields {
		v := strings.TrimSpace(brutos[f.Key])
		if v == "" {
			if !f.Optional {
				v = "[" + f.Label + "]"
			}
		} else if f.Moeda && !strings.HasPrefix(strings.ToUpper(v), "R$") {
			v = "R$ " + v
		}
		resolvido[f.Key] = v
	}
	return resolvido
}

// Item é uma operação já com os valores resolvidos, pronta pra entrar no
// e-mail.
type Item struct {
	Cat     *Category
	Valores map[string]string // já passado por Cat.ResolveValores
}

// MontarTextoEmail monta o texto final do e-mail de ordem a partir de 1+
// operações. Porte de montar_texto_email.
func MontarTextoEmail(codigo, nome string, itens []Item) string {
	var intro, corpo, fechamento string

	if len(itens) == 1 {
		cat, v := itens[0].Cat, itens[0].Valores
		if RVGroups[cat.Group] {
			introFrase := cat.IntroFrase
			if introFrase == "" {
				introFrase = "a operação"
			}
			intro = fmt.Sprintf("Conforme conversado, gostaria de realizar a ordem para %s abaixo na conta XP %s:", introFrase, codigo)
			corpo = cat.Body(v)
		} else {
			intro = fmt.Sprintf("Conforme conversado, gostaria de realizar a ordem para a operação abaixo na conta XP %s:", codigo)
			corpo = fmt.Sprintf("%s - %s\n%s", cat.Group, cat.Label, cat.Body(v))
		}
		fechamento = cat.Fechamento
	} else {
		intro = fmt.Sprintf("Conforme conversado, gostaria de realizar a ordem para as operações abaixo na conta XP %s:", codigo)
		partes := make([]string, len(itens))
		for i, item := range itens {
			partes[i] = fmt.Sprintf("%d) %s - %s\n%s", i+1, item.Cat.Group, item.Cat.Label, item.Cat.Body(item.Valores))
		}
		corpo = strings.Join(partes, "\n\n")
		fechamento = "Aguardo confirmação para realizar as operações acima."
	}

	return fmt.Sprintf("Prezado(a) %s,\n\n%s\n\n%s\n\n%s", nome, intro, corpo, fechamento)
}

// MontarTextoEmailPadronizado monta o e-mail no modo padronizado (regra de
// compliance: todas as operações do e-mail são do mesmo produto e tipo de
// movimentação). Sempre usa a frase de abertura específica da categoria —
// no modo padronizado ela nunca é ambígua, já que só há um produto.
func MontarTextoEmailPadronizado(codigo, nome string, itens []Item) (string, error) {
	cat := itens[0].Cat
	for _, item := range itens[1:] {
		if item.Cat != cat {
			return "", fmt.Errorf("no modo padronizado todas as operações do e-mail devem ser do mesmo produto e tipo (%s - %s)", cat.Group, cat.Label)
		}
	}

	if cat.EmailCompleto != nil {
		if len(itens) > 1 {
			return "", fmt.Errorf("%s - %s permite apenas uma operação por e-mail", cat.Group, cat.Label)
		}
		return cat.EmailCompleto(codigo, nome, itens[0].Valores), nil
	}

	introFrase := cat.IntroFrase
	if introFrase == "" {
		introFrase = "a operação"
	}

	var intro, corpo, fechamento string
	if len(itens) == 1 {
		intro = fmt.Sprintf("Conforme conversado, gostaria de realizar a ordem para %s abaixo na conta XP %s:", introFrase, codigo)
		corpo = cat.Body(itens[0].Valores)
		fechamento = cat.Fechamento
	} else {
		intro = fmt.Sprintf("Conforme conversado, gostaria de realizar as ordens para %s abaixo na conta XP %s:", introFrase, codigo)
		partes := make([]string, len(itens))
		for i, item := range itens {
			partes[i] = fmt.Sprintf("%d)\n%s", i+1, cat.Body(item.Valores))
		}
		corpo = strings.Join(partes, "\n\n")
		fechamento = "Aguardo confirmação para realizar as operações acima."
	}

	return fmt.Sprintf("Prezado(a) %s,\n\n%s\n\n%s\n\n%s", nome, intro, corpo, fechamento), nil
}
