package emailgen

import (
	"strings"
	"testing"
)

func TestCategoriaPorGrupoLabel(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Renda Fixa", "Aplicação")
	if cat == nil {
		t.Fatal("esperava achar 'Renda Fixa'/'Aplicação'")
	}
	if cat.ID != "rf-aplicacao" {
		t.Errorf("cat.ID = %q, esperado rf-aplicacao", cat.ID)
	}

	if got := CategoriaPorGrupoLabel("Produto Inexistente", "Operação Inexistente"); got != nil {
		t.Errorf("esperava nil pra categoria inexistente, veio %+v", got)
	}
}

func TestResolveValoresFallback(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Renda Fixa", "Aplicação")
	brutos := map[string]string{
		"ativo":   "CDB Banco X",
		"emissor": "  ", // só espaço -> conta como vazio
	}
	resolvido := cat.ResolveValores(brutos)

	if resolvido["ativo"] != "CDB Banco X" {
		t.Errorf("ativo = %q, esperado 'CDB Banco X'", resolvido["ativo"])
	}
	if resolvido["emissor"] != "[Emissor]" {
		t.Errorf("emissor = %q, esperado '[Emissor]'", resolvido["emissor"])
	}
	if resolvido["taxaNegociacao"] != "[Taxa de Negociação]" {
		t.Errorf("taxaNegociacao = %q, esperado fallback com o label", resolvido["taxaNegociacao"])
	}
}

func TestResolveValoresOpcionalVazioFicaVazio(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Renda Fixa", "Resgate")
	resolvido := cat.ResolveValores(map[string]string{"ativo": "CDB X"})

	if resolvido["valorCotado"] != "" {
		t.Errorf("valorCotado (opcional, vazio) = %q, esperado string vazia", resolvido["valorCotado"])
	}
	if texto := cat.Body(resolvido); strings.Contains(texto, "Valor financeiro cotado") {
		t.Errorf("linha opcional vazia não deveria aparecer no corpo:\n%s", texto)
	}

	resolvido = cat.ResolveValores(map[string]string{"valorCotado": "R$ 10.000,00"})
	if texto := cat.Body(resolvido); !strings.Contains(texto, "Valor financeiro cotado: R$ 10.000,00") {
		t.Errorf("linha opcional preenchida deveria aparecer no corpo:\n%s", texto)
	}
}

func TestMontarTextoEmailUmItemRV(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Ações, Futuros e FIIs", "Compra")
	v := cat.ResolveValores(map[string]string{
		"ativo": "PETR4", "quantidade": "100", "preco": "R$ 30,00", "validade": "Até Cancelar",
	})
	texto := MontarTextoEmail("4312514", "Gustavo", []Item{{Cat: cat, Valores: v}})

	if !strings.Contains(texto, "gostaria de realizar a ordem para a compra abaixo na conta XP 4312514") {
		t.Errorf("intro RV não bateu:\n%s", texto)
	}
	if strings.Contains(texto, "Ações, Futuros e FIIs - Compra") {
		t.Errorf("grupo RV não deveria ter cabeçalho 'Grupo - Label' no corpo:\n%s", texto)
	}
	if !strings.Contains(texto, "Ativo: PETR4;") {
		t.Errorf("corpo não bateu:\n%s", texto)
	}
	if !strings.Contains(texto, "Aguardo confirmação para realizar a compra.") {
		t.Errorf("fechamento não bateu:\n%s", texto)
	}
}

func TestMontarTextoEmailUmItemNaoRV(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Renda Fixa", "Aplicação")
	v := cat.ResolveValores(map[string]string{"ativo": "CDB X"})
	texto := MontarTextoEmail("111", "Cliente", []Item{{Cat: cat, Valores: v}})

	if !strings.Contains(texto, "gostaria de realizar a ordem para a operação abaixo na conta XP 111") {
		t.Errorf("intro genérica não bateu:\n%s", texto)
	}
	if !strings.Contains(texto, "Renda Fixa - Aplicação\nAtivo: CDB X;") {
		t.Errorf("cabeçalho 'Grupo - Label' + corpo não bateu:\n%s", texto)
	}
}

func TestMontarTextoEmailMultiplosItens(t *testing.T) {
	cat1 := CategoriaPorGrupoLabel("Renda Fixa", "Aplicação")
	v1 := cat1.ResolveValores(map[string]string{"ativo": "CDB X"})
	cat2 := CategoriaPorGrupoLabel("Tesouro Direto", "Resgate")
	v2 := cat2.ResolveValores(map[string]string{"titulo": "Tesouro Selic"})

	texto := MontarTextoEmail("222", "Cliente", []Item{
		{Cat: cat1, Valores: v1},
		{Cat: cat2, Valores: v2},
	})

	if !strings.Contains(texto, "gostaria de realizar a ordem para as operações abaixo na conta XP 222") {
		t.Errorf("intro múltipla não bateu:\n%s", texto)
	}
	if !strings.Contains(texto, "1) Renda Fixa - Aplicação") {
		t.Errorf("numeração do item 1 não bateu:\n%s", texto)
	}
	if !strings.Contains(texto, "2) Tesouro Direto - Resgate") {
		t.Errorf("numeração do item 2 não bateu:\n%s", texto)
	}
	if !strings.Contains(texto, "Aguardo confirmação para realizar as operações acima.") {
		t.Errorf("fechamento agregado não bateu:\n%s", texto)
	}
}

// TestTodasCategoriasBodyNaoQuebra percorre as ~30 categorias chamando Body
// com um mapa fabricado (uma entrada por Field.Key) — pega "chave
// inexistente no mapa" (que em Go simplesmente formata como string vazia,
// não panica, mas o teste ainda serve pra garantir que nenhum Body
// referencia uma chave fora da lista de Fields declarada).
func TestTodasCategoriasBodyNaoQuebra(t *testing.T) {
	if len(Categories) < 25 {
		t.Fatalf("esperava ~30 categorias, veio %d — porte incompleto?", len(Categories))
	}
	for _, cat := range Categories {
		v := make(map[string]string, len(cat.Fields))
		for _, f := range cat.Fields {
			v[f.Key] = "[valor de teste: " + f.Key + "]"
		}
		var texto string
		if cat.EmailCompleto != nil {
			texto = cat.EmailCompleto("000", "Cliente", v)
		} else {
			texto = cat.Body(v)
		}
		if texto == "" {
			t.Errorf("categoria %s (%s/%s): Body devolveu string vazia", cat.ID, cat.Group, cat.Label)
		}
		for _, f := range cat.Fields {
			marcador := "[valor de teste: " + f.Key + "]"
			if !strings.Contains(texto, marcador) {
				t.Errorf("categoria %s (%s/%s): Body não usa o campo %q (esperava encontrar %q no texto)", cat.ID, cat.Group, cat.Label, f.Key, marcador)
			}
		}
	}
}

func TestMontarTextoEmailNaoEnvolveEmAspas(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Renda Fixa", "Aplicação")
	v := cat.ResolveValores(map[string]string{"ativo": "CDB X"})
	texto := MontarTextoEmail("111", "Cliente", []Item{{Cat: cat, Valores: v}})
	if strings.HasPrefix(texto, "\"") || strings.HasSuffix(texto, "\"") {
		t.Errorf("texto não deveria vir entre aspas literais:\n%s", texto)
	}
	if !strings.HasPrefix(texto, "Prezado(a) Cliente,") {
		t.Errorf("texto deveria começar com a saudação:\n%s", texto)
	}
}

func TestMontarTextoEmailPadronizadoNaoEnvolveEmAspas(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Renda Fixa", "Aplicação")
	v := cat.ResolveValores(map[string]string{"ativo": "CDB X"})
	texto, err := MontarTextoEmailPadronizado("111", "Cliente", []Item{{Cat: cat, Valores: v}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(texto, "\"") || strings.HasSuffix(texto, "\"") {
		t.Errorf("texto não deveria vir entre aspas literais:\n%s", texto)
	}
	if !strings.HasPrefix(texto, "Prezado(a) Cliente,") {
		t.Errorf("texto deveria começar com a saudação:\n%s", texto)
	}
}

func TestMontarTextoEmailPadronizadoUmItem(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Fundos de Investimento", "Aplicação")
	v := cat.ResolveValores(map[string]string{"fundo": "Fundo Y", "valor": "R$ 1.000,00"})
	texto, err := MontarTextoEmailPadronizado("333", "Cliente", []Item{{Cat: cat, Valores: v}})
	if err != nil {
		t.Fatal(err)
	}
	// No modo padronizado a intro específica do produto vale pra todos os
	// grupos, não só Renda Variável.
	if !strings.Contains(texto, "a ordem para a aplicação no Fundo de Investimentos abaixo na conta XP 333") {
		t.Errorf("intro específica do produto não bateu:\n%s", texto)
	}
	if strings.Contains(texto, "Fundos de Investimento - Aplicação") {
		t.Errorf("modo padronizado não deveria ter cabeçalho 'Grupo - Label':\n%s", texto)
	}
}

func TestMontarTextoEmailPadronizadoMultiplosMesmoProduto(t *testing.T) {
	cat := CategoriaPorGrupoLabel("Renda Fixa", "Aplicação")
	v1 := cat.ResolveValores(map[string]string{"ativo": "CDB X"})
	v2 := cat.ResolveValores(map[string]string{"ativo": "CDB Y"})
	texto, err := MontarTextoEmailPadronizado("444", "Cliente", []Item{{Cat: cat, Valores: v1}, {Cat: cat, Valores: v2}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(texto, "1)\nAtivo: CDB X;") || !strings.Contains(texto, "2)\nAtivo: CDB Y;") {
		t.Errorf("numeração das operações não bateu:\n%s", texto)
	}
	if !strings.Contains(texto, "Aguardo confirmação para realizar as operações acima.") {
		t.Errorf("fechamento agregado não bateu:\n%s", texto)
	}
}

func TestMontarTextoEmailPadronizadoRejeitaProdutosMisturados(t *testing.T) {
	cat1 := CategoriaPorGrupoLabel("Renda Fixa", "Aplicação")
	cat2 := CategoriaPorGrupoLabel("Tesouro Direto", "Resgate")
	_, err := MontarTextoEmailPadronizado("555", "Cliente", []Item{
		{Cat: cat1, Valores: cat1.ResolveValores(nil)},
		{Cat: cat2, Valores: cat2.ResolveValores(nil)},
	})
	if err == nil {
		t.Fatal("esperava erro ao misturar produtos no modo padronizado")
	}
}

func TestMontarTextoEmailPadronizadoResgatePrev(t *testing.T) {
	for _, label := range []string{"PGBL", "VGBL"} {
		cat := CategoriaPorGrupoLabel("Resgate Prev", label)
		if cat == nil {
			t.Fatalf("esperava achar 'Resgate Prev'/%q", label)
		}
		if !cat.SoPadronizado {
			t.Errorf("%s deveria ser SoPadronizado", label)
		}
		v := cat.ResolveValores(map[string]string{"aliquota": "15%", "data": "01/07/26"})
		texto, err := MontarTextoEmailPadronizado("666", "Gustavo", []Item{{Cat: cat, Valores: v}})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(texto, "Prezado(a) Gustavo, tudo bem?") {
			t.Errorf("%s: saudação própria não bateu:\n%s", label, texto)
		}
		if !strings.Contains(texto, "O resgate de um "+label+" na alíquota 15%") {
			t.Errorf("%s: corpo não bateu:\n%s", label, texto)
		}
		if !strings.Contains(texto, "resgate de previdência realizado em 01/07/26") {
			t.Errorf("%s: data não bateu:\n%s", label, texto)
		}

		// Duas operações num e-mail de Resgate Prev não são permitidas.
		if _, err := MontarTextoEmailPadronizado("666", "Gustavo", []Item{{Cat: cat, Valores: v}, {Cat: cat, Valores: v}}); err == nil {
			t.Errorf("%s: esperava erro com mais de uma operação", label)
		}
	}
}

func TestProdutosOrdemDePrimeiraOcorrencia(t *testing.T) {
	if len(Produtos) == 0 {
		t.Fatal("Produtos não foi preenchido")
	}
	if Produtos[0] != "Renda Fixa" {
		t.Errorf("Produtos[0] = %q, esperado 'Renda Fixa' (primeiro grupo em Categories)", Produtos[0])
	}
	vistos := map[string]bool{}
	for _, p := range Produtos {
		if vistos[p] {
			t.Errorf("produto %q duplicado em Produtos", p)
		}
		vistos[p] = true
	}
}
