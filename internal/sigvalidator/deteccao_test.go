package sigvalidator

import "testing"

func TestDetectarFormato(t *testing.T) {
	casos := []struct {
		nome            string
		caminhos        []string
		formato         Formato
		caminhoP7S      string
		caminhoConteudo string
		esperaErro      bool
	}{
		{
			nome:       "nenhum arquivo",
			caminhos:   nil,
			esperaErro: true,
		},
		{
			nome:       "só .p7s — candidato a anexado",
			caminhos:   []string{"contrato.p7s"},
			formato:    FormatoCAdESAnexado,
			caminhoP7S: "contrato.p7s",
		},
		{
			nome:       "só .p7m — candidato a anexado",
			caminhos:   []string{"nota.p7m"},
			formato:    FormatoCAdESAnexado,
			caminhoP7S: "nota.p7m",
		},
		{
			nome:       "só .pdf — PAdES",
			caminhos:   []string{"documento.pdf"},
			formato:    FormatoPAdES,
			caminhoP7S: "documento.pdf", // Validate() usa esse valor como o caminho do PDF no caso PAdES (ver sigvalidator.go)
		},
		{
			nome:     "extensão desconhecida sozinha",
			caminhos: []string{"documento.docx"},
			formato:  FormatoDesconhecido,
		},
		{
			nome:            ".p7s + conteúdo — destacado",
			caminhos:        []string{"contrato.p7s", "contrato.pdf"},
			formato:         FormatoCAdESDestacado,
			caminhoP7S:      "contrato.p7s",
			caminhoConteudo: "contrato.pdf",
		},
		{
			nome:            "conteúdo + .p7m — destacado (ordem trocada)",
			caminhos:        []string{"contrato.pdf", "contrato.p7m"},
			formato:         FormatoCAdESDestacado,
			caminhoP7S:      "contrato.p7m",
			caminhoConteudo: "contrato.pdf",
		},
		{
			nome:     "dois .p7s — não dá pra saber qual é o par",
			caminhos: []string{"a.p7s", "b.p7s"},
			formato:  FormatoDesconhecido,
		},
		{
			nome:     "dois arquivos, nenhum é assinatura",
			caminhos: []string{"a.pdf", "b.pdf"},
			formato:  FormatoDesconhecido,
		},
		{
			nome:     "mais de dois arquivos",
			caminhos: []string{"a.p7s", "b.pdf", "c.pdf"},
			formato:  FormatoDesconhecido,
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			formato, p7s, conteudo, err := DetectarFormato(c.caminhos)
			if c.esperaErro {
				if err == nil {
					t.Fatal("esperava erro, veio nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if formato != c.formato {
				t.Errorf("formato: esperado %v, veio %v", c.formato, formato)
			}
			if p7s != c.caminhoP7S {
				t.Errorf("caminhoP7S: esperado %q, veio %q", c.caminhoP7S, p7s)
			}
			if conteudo != c.caminhoConteudo {
				t.Errorf("caminhoConteudo: esperado %q, veio %q", c.caminhoConteudo, conteudo)
			}
		})
	}
}
