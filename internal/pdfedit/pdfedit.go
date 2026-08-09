// Package pdfedit monta um novo PDF a partir de uma lista de imagens (uma
// por página) — usado pela aba Editor de PDF. Cada página editada (a
// página original renderizada + o que o usuário desenhou/escreveu por
// cima, tudo achatado no frontend via <canvas>) vira uma página do PDF
// final. Não depende do PDFium: a leitura/renderização do PDF de origem já
// é feita pelo motor existente (pdfreport.Extractor); este pacote só cuida
// de escrever o resultado.
package pdfedit

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
)

// mmPorPonto converte pontos PDF (1/72 polegada) para milímetros, unidade
// que o fpdf usa internamente.
const mmPorPonto = 25.4 / 72.0

// Pagina é uma página já achatada em JPEG, com o tamanho original (em
// pontos) preservado — assim o PDF final mantém a proporção e o tamanho de
// página do PDF de origem.
type Pagina struct {
	JPEG      []byte
	LarguraPt float64
	AlturaPt  float64
}

// Montar cria um novo PDF com uma página por item de paginas, na ordem
// dada, e devolve os bytes prontos pra salvar em disco.
func Montar(paginas []Pagina) ([]byte, error) {
	if len(paginas) == 0 {
		return nil, fmt.Errorf("nenhuma página para montar")
	}

	primeira := paginas[0]
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size:           fpdf.SizeType{Wd: primeira.LarguraPt * mmPorPonto, Ht: primeira.AlturaPt * mmPorPonto},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	for i, pagina := range paginas {
		if len(pagina.JPEG) == 0 {
			return nil, fmt.Errorf("página %d sem conteúdo", i+1)
		}
		larguraMM := pagina.LarguraPt * mmPorPonto
		alturaMM := pagina.AlturaPt * mmPorPonto
		pdf.AddPageFormat("P", fpdf.SizeType{Wd: larguraMM, Ht: alturaMM})

		nome := fmt.Sprintf("pagina%d", i)
		opt := fpdf.ImageOptions{ImageType: "JPG", ReadDpi: false}
		pdf.RegisterImageOptionsReader(nome, opt, bytes.NewReader(pagina.JPEG))
		pdf.ImageOptions(nome, 0, 0, larguraMM, alturaMM, false, opt, 0, "")
	}

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("montar PDF: %w", err)
	}

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, fmt.Errorf("gerar PDF: %w", err)
	}
	return out.Bytes(), nil
}
