package pdfreport

import (
	"math"
	"os"
	"testing"
)

func TestParseNumeroPtBR(t *testing.T) {
	casos := map[string]float64{
		"R$ 1.234,56": 1234.56,
		"-1,08%":      -1.08,
		"2.434,91":    2434.91,
		"-R$ 8,51":    -8.51,
		"0,56":        0.56,
	}
	for entrada, esperado := range casos {
		valor, err := ParseNumeroPtBR(entrada)
		if err != nil {
			t.Fatalf("ParseNumeroPtBR(%q): %v", entrada, err)
		}
		if math.Abs(valor-esperado) > 1e-9 {
			t.Errorf("ParseNumeroPtBR(%q) = %v, esperado %v", entrada, valor, esperado)
		}
	}
}

func TestFormatarReais(t *testing.T) {
	casos := map[float64]string{
		1234.56:  "R$ 1.234,56",
		-1234.56: "-R$ 1.234,56",
		8.51:     "R$ 8,51",
		-8.51:    "-R$ 8,51",
		0:        "R$ 0,00",
		681.41:   "R$ 681,41",
		-3558.43: "-R$ 3.558,43",
	}
	for valor, esperado := range casos {
		got := FormatarReais(valor)
		if got != esperado {
			t.Errorf("FormatarReais(%v) = %q, esperado %q", valor, got, esperado)
		}
	}
}

func TestFormatarPercentual(t *testing.T) {
	casos := map[float64]string{
		0.56:  "0,56%",
		-1.08: "-1,08%",
		54.79: "54,79%",
		3.75:  "3,75%",
	}
	for valor, esperado := range casos {
		got := FormatarPercentual(valor)
		if got != esperado {
			t.Errorf("FormatarPercentual(%v) = %q, esperado %q", valor, got, esperado)
		}
	}
}

func TestExtrairDadosDoTexto(t *testing.T) {
	texto := `Relatório de
Investimentos
Conta Assessor Data de Referência
4312514 Digital Gustavo Meireles 08/07/2026
Gerado em 8 de julho de 2026 às 09:28
PATRIMÔNIO TOTAL BRUTO:
R$ 31.476,53
R E S U M O D E I N FOR M AÇÕE S DA C A RT E I R A
Período Ganho Financeiro Rentabilidade %CDI Movimentações
MÊS -R$ 8,51 -0,12% -39,38% -R$ 1.094,22
ANO R$ 649,69 3,46% 48,23% -R$ 1.708,41
12M R$ 610,93 2,61% 19,05% R$ 6.961,55
Relatório informativo de performance não destinado a fins fiscais Data de referência: 08/07/2026`

	dados, err := extrairDadosDoTexto(texto)
	if err != nil {
		t.Fatalf("extrairDadosDoTexto: %v", err)
	}

	esperado := Dados{
		GanhoMes: -8.51, RentMesPct: -0.12, CDIMesPct: -39.38,
		GanhoAno: 649.69, RentAnoPct: 3.46, CDIAnoPct: 48.23,
		Ganho12M: 610.93, Rent12MPct: 2.61, CDI12MPct: 19.05,
		DataReferencia:       "08/07/2026",
		PatrimonioTotalBruto: 31476.53,
	}
	if dados != esperado {
		t.Errorf("extrairDadosDoTexto = %+v, esperado %+v", dados, esperado)
	}
}

// TestExtrairDadosDoTextoCDIExtremo cobre o caso real de um cliente cujo
// CDI do período foi próximo de zero: a rentabilidade dividida por ele
// "explode" pra um %CDI de 4+ dígitos, com ponto de milhar ("-1.993,75%")
// em vez do formato usual de 2-3 dígitos só com vírgula decimal.
func TestExtrairDadosDoTextoCDIExtremo(t *testing.T) {
	texto := `R E S U M O D E I N FOR M AÇÕE S DA C A RT E I R A
Período Ganho Financeiro Rentabilidade %CDI Movimentações
MÊS -R$ 2.204,28 -24,23% -1.993,75% R$ 0,00
ANO -R$ 3.368,92 -32,82% -403,03% R$ 0,00
12M -R$ 495,13 -6,72% -45,65% R$ 0,00`

	dados, err := extrairDadosDoTexto(texto)
	if err != nil {
		t.Fatalf("extrairDadosDoTexto: %v", err)
	}
	if dados.CDIMesPct != -1993.75 {
		t.Errorf("CDIMesPct = %v, esperado -1993.75", dados.CDIMesPct)
	}
	if dados.GanhoMes != -2204.28 || dados.RentMesPct != -24.23 {
		t.Errorf("GanhoMes/RentMesPct = %v/%v, esperado -2204.28/-24.23", dados.GanhoMes, dados.RentMesPct)
	}
}

func TestExtrairDadosDoTextoSem12M(t *testing.T) {
	texto := `R E S U M O D E I N FOR M AÇÕE S DA C A RT E I R A
Período Ganho Financeiro Rentabilidade %CDI Movimentações
MÊS -R$ 8,51 -0,12% -39,38% -R$ 1.094,22
ANO R$ 649,69 3,46% 48,23% -R$ 1.708,41`

	dados, err := extrairDadosDoTexto(texto)
	if err != nil {
		t.Fatalf("extrairDadosDoTexto: %v", err)
	}
	if dados.Ganho12M != 0 || dados.Rent12MPct != 0 || dados.CDI12MPct != 0 {
		t.Errorf("campos 12M = %+v, esperados zerados (sem linha 12M no texto — cliente com menos de 12 meses de casa)", dados)
	}
}

func TestExtrairDadosDoTextoSemPatrimonio(t *testing.T) {
	texto := `R E S U M O D E I N FOR M AÇÕE S DA C A RT E I R A
Período Ganho Financeiro Rentabilidade %CDI Movimentações
MÊS -R$ 8,51 -0,12% -39,38% -R$ 1.094,22
ANO R$ 649,69 3,46% 48,23% -R$ 1.708,41
12M R$ 610,93 2,61% 19,05% R$ 6.961,55`

	dados, err := extrairDadosDoTexto(texto)
	if err != nil {
		t.Fatalf("extrairDadosDoTexto: %v", err)
	}
	if dados.PatrimonioTotalBruto != 0 {
		t.Errorf("PatrimonioTotalBruto = %v, esperado 0 (sem card no texto)", dados.PatrimonioTotalBruto)
	}
}

// TestExtractorComPDFReal roda o parser completo (PDFium + regex) contra um
// PDF real de relatório XP. Aponte PDFREPORT_TEST_PDF para o arquivo antes
// de rodar; o teste é pulado se a variável não estiver definida.
func TestExtractorComPDFReal(t *testing.T) {
	caminho := os.Getenv("PDFREPORT_TEST_PDF")
	if caminho == "" {
		t.Skip("defina PDFREPORT_TEST_PDF com o caminho de um PDF de relatório XP para rodar este teste")
	}

	extractor, err := NewExtractor()
	if err != nil {
		t.Fatalf("NewExtractor: %v", err)
	}
	defer extractor.Close()

	dados, err := extractor.ExtrairDados(caminho)
	if err != nil {
		t.Fatalf("ExtrairDados: %v", err)
	}
	t.Logf("dados extraídos: %+v", dados)

	if dados.DataReferencia == "" {
		t.Error("data de referência não encontrada")
	}
	if dados.GanhoMes == 0 && dados.RentMesPct == 0 {
		t.Error("dados do mês parecem vazios")
	}
}

func TestRenderizarGraficoRentabilidadeComPDFReal(t *testing.T) {
	caminho := os.Getenv("PDFREPORT_TEST_PDF")
	if caminho == "" {
		t.Skip("defina PDFREPORT_TEST_PDF com o caminho de um PDF de relatório XP para rodar este teste")
	}

	extractor, err := NewExtractor()
	if err != nil {
		t.Fatalf("NewExtractor: %v", err)
	}
	defer extractor.Close()

	png, err := extractor.RenderizarGraficoRentabilidade(caminho, RecorteGraficoRentabilidadePadrao)
	if err != nil {
		t.Fatalf("RenderizarGraficoRentabilidade: %v", err)
	}
	if len(png) < 1000 {
		t.Fatalf("PNG suspeito pequeno: %d bytes", len(png))
	}
	// Assinatura PNG: 89 50 4E 47 0D 0A 1A 0A
	assinatura := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range assinatura {
		if png[i] != b {
			t.Fatalf("assinatura PNG inválida no byte %d: got %x, want %x", i, png[i], b)
		}
	}
	t.Logf("PNG gerado: %d bytes", len(png))
}
