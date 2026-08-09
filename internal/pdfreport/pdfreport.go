// Package pdfreport extrai os dados de rentabilidade dos relatórios XP em
// PDF (a tabela "RESUMO DE INFORMAÇÕES DA CARTEIRA" tem uma linha por
// período, sempre no formato:
//
//	MÊS -R$ 8,51 -0,12% -39,38% -R$ 1.094,22
//	ANO R$ 649,69 3,46% 48,23% -R$ 1.708,41
//	12M R$ 610,93 2,61% 19,05% R$ 6.961,55
//
// (Período, Ganho Financeiro, Rentabilidade, %CDI, Movimentações) — é a
// única fonte que tem R$, % e %CDI juntos, por isso usamos ela pros três
// períodos (existe também uma linha "24M", mas não é lida).
package pdfreport

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/tetratelabs/wazero"
)

// compilationCacheDir é onde o wazero guarda o binário WASM do PDFium já
// compilado, entre uma execução do app e outra.
func compilationCacheDir() (string, error) {
	dir, err := os.UserCacheDir() // %LocalAppData% no Windows
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "RentabilidadeXP", "pdfium-wasm-cache"), nil
}

// Rentabilidade e %CDI aceitam ponto de milhar (não só a vírgula decimal):
// quando o CDI do período é muito baixo, a rentabilidade dividida por ele
// "explode" pra %CDI de 4+ dígitos (ex.: "-1.993,75%" — confirmado num
// relatório real, conta com CDI do mês próximo de zero).
func padraoLinhaResumo(rotulo string) *regexp.Regexp {
	return regexp.MustCompile(rotulo + `\s+(-?R\$\s*[\d.,]+)\s+(-?[\d.,]+)%\s+(-?[\d.,]+)%\s+-?R\$\s*[\d.,]+`)
}

var (
	linhaMesRe = padraoLinhaResumo(`M[ÊE]S`)
	linhaAnoRe = padraoLinhaResumo(`ANO`)
	linha12MRe = padraoLinhaResumo(`12M`)

	dataRefRe = regexp.MustCompile(`(?i)Data de refer[êe]ncia:?\s*(\d{2}/\d{2}/\d{4})`)

	// Card "PATRIMÔNIO TOTAL BRUTO:" no topo da seção "Evolução
	// Patrimonial" — mesma página do gráfico de rentabilidade
	// (PaginaGraficoRentabilidade). Usado só pra ordenar a lista de
	// clientes por patrimônio, não é exibido em lugar nenhum.
	patrimonioRe = regexp.MustCompile(`(?i)PATRIM[ÔO]NIO TOTAL BRUTO:\s*(R\$\s*[\d.,]+)`)
)

// ParseNumeroPtBR aceita "R$ 1.234,56", "-1,08%" ou "2.434,91" (pt-BR: ponto
// separa milhar, vírgula separa decimal) -> float64. Serve tanto pra R$
// quanto pra %.
func ParseNumeroPtBR(texto string) (float64, error) {
	texto = strings.TrimSpace(texto)
	negativo := strings.HasPrefix(texto, "-")
	numero := strings.TrimPrefix(texto, "-")
	numero = strings.ReplaceAll(numero, "R$", "")
	numero = strings.ReplaceAll(numero, "%", "")
	numero = strings.TrimSpace(numero)
	numero = strings.ReplaceAll(numero, ".", "")
	numero = strings.ReplaceAll(numero, ",", ".")
	valor, err := strconv.ParseFloat(numero, 64)
	if err != nil {
		return 0, err
	}
	if negativo {
		valor = -valor
	}
	return valor, nil
}

// magnitudePtBR devolve abs(valor) formatado como "1.234,56" (sem sinal —
// quem chama decide onde colocar o "-" e o sufixo/prefixo).
func magnitudePtBR(valor float64) string {
	if valor < 0 {
		valor = -valor
	}
	inteiro := int64(valor)
	frac := int64((valor-float64(inteiro))*100 + 0.5)
	if frac >= 100 {
		inteiro++
		frac -= 100
	}
	milhar := formatarMilhar(inteiro)
	return fmt.Sprintf("%s,%02d", milhar, frac)
}

func formatarMilhar(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var partes []string
	for len(s) > 3 {
		partes = append([]string{s[len(s)-3:]}, partes...)
		s = s[:len(s)-3]
	}
	partes = append([]string{s}, partes...)
	return strings.Join(partes, ".")
}

// FormatarReais formata um valor como "R$ 1.234,56" ou "-R$ 1.234,56".
func FormatarReais(valor float64) string {
	sinal := ""
	if valor < 0 {
		sinal = "-"
	}
	return fmt.Sprintf("%sR$ %s", sinal, magnitudePtBR(valor))
}

// FormatarPercentual formata um valor como "1.234,56%" ou "-1.234,56%".
func FormatarPercentual(valor float64) string {
	sinal := ""
	if valor < 0 {
		sinal = "-"
	}
	return fmt.Sprintf("%s%s%%", sinal, magnitudePtBR(valor))
}

// Dados são os valores extraídos de um relatório.
type Dados struct {
	GanhoMes             float64
	RentMesPct           float64
	CDIMesPct            float64
	GanhoAno             float64
	RentAnoPct           float64
	CDIAnoPct            float64
	Ganho12M             float64
	Rent12MPct           float64
	CDI12MPct            float64
	DataReferencia       string  // "dd/mm/aaaa", ou "" se não encontrada
	PatrimonioTotalBruto float64 // usado só pra ordenar a lista de clientes; 0 se não encontrado
}

func extrairPeriodo(texto string, padrao *regexp.Regexp) (ganho, rentPct, cdiPct float64, err error) {
	m := padrao.FindStringSubmatch(texto)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("linha de resumo não encontrada no PDF")
	}
	if ganho, err = ParseNumeroPtBR(m[1]); err != nil {
		return 0, 0, 0, err
	}
	if rentPct, err = ParseNumeroPtBR(m[2]); err != nil {
		return 0, 0, 0, err
	}
	if cdiPct, err = ParseNumeroPtBR(m[3]); err != nil {
		return 0, 0, 0, err
	}
	return ganho, rentPct, cdiPct, nil
}

// extrairDadosDoTexto aplica os regexes sobre o texto já extraído do PDF.
// Separado de ExtrairDados para poder ser testado sem abrir PDFium.
func extrairDadosDoTexto(texto string) (Dados, error) {
	var dados Dados

	ganhoMes, rentMes, cdiMes, err := extrairPeriodo(texto, linhaMesRe)
	if err != nil {
		return Dados{}, fmt.Errorf("período 'mes': %w", err)
	}
	dados.GanhoMes, dados.RentMesPct, dados.CDIMesPct = ganhoMes, rentMes, cdiMes

	ganhoAno, rentAno, cdiAno, err := extrairPeriodo(texto, linhaAnoRe)
	if err != nil {
		return Dados{}, fmt.Errorf("período 'ano': %w", err)
	}
	dados.GanhoAno, dados.RentAnoPct, dados.CDIAnoPct = ganhoAno, rentAno, cdiAno

	// 12M é opcional: clientes com menos de 12 meses de casa não têm essa
	// linha no relatório — sua ausência não deve impedir o processamento
	// (mesma lógica do patrimônio total abaixo).
	if ganho12M, rent12M, cdi12M, err := extrairPeriodo(texto, linha12MRe); err == nil {
		dados.Ganho12M, dados.Rent12MPct, dados.CDI12MPct = ganho12M, rent12M, cdi12M
	}

	if m := dataRefRe.FindStringSubmatch(texto); m != nil {
		dados.DataReferencia = m[1]
	}

	// Patrimônio é opcional: só usado pra ordenar a lista de clientes, sua
	// ausência não deve impedir o processamento do relatório.
	if m := patrimonioRe.FindStringSubmatch(texto); m != nil {
		if valor, err := ParseNumeroPtBR(m[1]); err == nil {
			dados.PatrimonioTotalBruto = valor
		}
	}

	return dados, nil
}

// Extractor mantém uma instância de PDFium (via WASM/wazero, sem cgo) viva
// entre chamadas — abrir uma instância nova a cada PDF é caro; reabrir uma
// pasta com uma centena de relatórios seria lento.
type Extractor struct {
	pool     pdfium.Pool
	instance pdfium.Pdfium
}

// NewExtractor inicializa o motor PDFium. Deve ser chamado uma vez no
// startup do app e fechado com Close() no shutdown.
//
// Sem cache de compilação, o wazero recompila o binário WASM do PDFium do
// zero toda vez que o app abre — isso sozinho levava ~3,4s (medido) e
// deixava a interface "travada" na abertura. Com WithCompilationCache
// apontando pra uma pasta persistente, só a primeira execução paga esse
// custo; as próximas reaproveitam o cache em disco e inicializam quase
// instantaneamente.
func NewExtractor() (*Extractor, error) {
	// O .exe é uma aplicação GUI (sem console); nesse caso o Windows não dá
	// ao processo um handle válido de stdout/stderr. Passar Stdout/Stderr
	// no Config abaixo não é suficiente — o go-pdfium referencia os.Stdout
	// diretamente num ponto interno (fora do Config, um listener de debug
	// que fica vestigial no código deles) e o wazero também acessa o
	// stdout real do processo ao montar o filesystem do WASM. Por isso
	// trocamos o os.Stdout/os.Stderr *do processo* por um destino válido
	// (NUL) só durante a inicialização — cobre qualquer referência interna
	// à variável global, não só a que passamos explicitamente no Config.
	// O .exe é uma aplicação GUI (sem console); nesse caso o Windows não dá
	// ao processo um handle válido de stdout/stderr. Passar Stdout/Stderr
	// no Config abaixo só cobre a saída do próprio módulo WASM (o guest) —
	// o go-pdfium ainda referencia os.Stdout diretamente num ponto interno
	// (fora do Config). A correção de verdade é global, em main.go
	// (redirecionarStdioParaArquivo), que troca os.Stdout/os.Stderr do
	// processo por um arquivo de log antes de qualquer coisa rodar —
	// então quando este código executa, os.Stdout já é válido.
	runtimeConfig := wazero.NewRuntimeConfig()
	if cacheDir, err := compilationCacheDir(); err == nil {
		if cache, err := wazero.NewCompilationCacheWithDir(cacheDir); err == nil {
			runtimeConfig = runtimeConfig.WithCompilationCache(cache)
		}
	}

	pool, err := webassembly.Init(webassembly.Config{
		MinIdle:       1,
		MaxIdle:       1,
		MaxTotal:      1,
		RuntimeConfig: runtimeConfig,
		Stdout:        io.Discard, // saída do próprio módulo WASM (o guest), não usada
		Stderr:        io.Discard,
	})
	if err != nil {
		return nil, fmt.Errorf("inicializar PDFium: %w", err)
	}
	instance, err := pool.GetInstance(30 * time.Second)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("obter instância PDFium: %w", err)
	}
	return &Extractor{pool: pool, instance: instance}, nil
}

// Close libera os recursos do motor PDFium.
func (e *Extractor) Close() {
	if e == nil {
		return
	}
	if e.instance != nil {
		e.instance.Close()
	}
	if e.pool != nil {
		e.pool.Close()
	}
}

// ExtrairTexto lê todas as páginas do PDF e devolve o texto concatenado.
func (e *Extractor) ExtrairTexto(caminho string) (string, error) {
	fileData, err := os.ReadFile(caminho)
	if err != nil {
		return "", err
	}

	doc, err := e.instance.OpenDocument(&requests.OpenDocument{File: &fileData})
	if err != nil {
		return "", fmt.Errorf("abrir PDF: %w", err)
	}
	defer e.instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	pageCount, err := e.instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return "", fmt.Errorf("contar páginas: %w", err)
	}

	var sb strings.Builder
	for i := range pageCount.PageCount {
		texto, err := e.instance.GetPageText(&requests.GetPageText{
			Page: requests.Page{
				ByIndex: &requests.PageByIndex{
					Document: doc.Document,
					Index:    i,
				},
			},
		})
		if err != nil {
			return "", fmt.Errorf("extrair texto da página %d: %w", i, err)
		}
		sb.WriteString(texto.Text)
		sb.WriteString("\n")
	}
	// PDFium devolve espaço não separável (U+00A0) em vez de espaço comum
	// em vários pontos do relatório (ex.: "R$ 8,51") — o \s do regexp
	// do Go (RE2) só cobre espaços ASCII, então sem essa normalização os
	// padrões de extração não batem.
	texto := strings.ReplaceAll(sb.String(), " ", " ")
	return texto, nil
}

// ExtrairDados lê o PDF em caminho e extrai os dados de rentabilidade.
func (e *Extractor) ExtrairDados(caminho string) (Dados, error) {
	texto, err := e.ExtrairTexto(caminho)
	if err != nil {
		return Dados{}, err
	}
	return extrairDadosDoTexto(texto)
}

// ContarPaginas devolve o número de páginas do PDF em caminho — usado pelo
// modo apresentação (slideshow) da aba Apresentação, que trata cada página
// como um slide.
func (e *Extractor) ContarPaginas(caminho string) (int, error) {
	fileData, err := os.ReadFile(caminho)
	if err != nil {
		return 0, err
	}

	doc, err := e.instance.OpenDocument(&requests.OpenDocument{File: &fileData})
	if err != nil {
		return 0, fmt.Errorf("abrir PDF: %w", err)
	}
	defer e.instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	pageCount, err := e.instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return 0, fmt.Errorf("contar páginas: %w", err)
	}
	return pageCount.PageCount, nil
}

// RenderizarPagina renderiza a página (0-indexed) do PDF em caminho, na
// resolução dpi indicada, já codificada em PNG. Genérica (qualquer página,
// sem recorte) — usada pelo modo apresentação; RenderizarGraficoRentabilidade
// continua sendo a versão específica (página fixa + recorte) do gráfico de
// rentabilidade.
func (e *Extractor) RenderizarPagina(caminho string, indice, dpi int) ([]byte, error) {
	fileData, err := os.ReadFile(caminho)
	if err != nil {
		return nil, err
	}

	doc, err := e.instance.OpenDocument(&requests.OpenDocument{File: &fileData})
	if err != nil {
		return nil, fmt.Errorf("abrir PDF: %w", err)
	}
	defer e.instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	resultado, err := e.instance.RenderPageInDPI(&requests.RenderPageInDPI{
		Page: requests.Page{
			ByIndex: &requests.PageByIndex{
				Document: doc.Document,
				Index:    indice,
			},
		},
		DPI: dpi,
	})
	if err != nil {
		return nil, fmt.Errorf("renderizar página %d: %w", indice, err)
	}
	defer resultado.Cleanup()

	origem := resultado.Result.Image
	copia := image.NewRGBA(origem.Bounds())
	draw.Draw(copia, copia.Bounds(), origem, origem.Bounds().Min, draw.Src)

	var buf bytes.Buffer
	if err := png.Encode(&buf, copia); err != nil {
		return nil, fmt.Errorf("codificar PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// PaginaGraficoRentabilidade é a página (0-indexed) do relatório XP onde
// fica o gráfico "Rentabilidade" com a evolução histórica — é sempre a
// mesma posição no modelo de relatório usado hoje.
const PaginaGraficoRentabilidade = 1

// RecorteGrafico delimita, em frações da página renderizada (0 a 1,
// independente do DPI usado), a área do gráfico "Rentabilidade" a recortar.
type RecorteGrafico struct{ X0, Y0, X1, Y1 float64 }

// RecorteGraficoRentabilidadePadrao é a área sem a tabela "Referências (%)"
// ao lado — medido visualmente contra um relatório real e confirmado com o
// usuário. Usado quando o usuário não personalizou o recorte nas
// Configurações.
var RecorteGraficoRentabilidadePadrao = RecorteGrafico{0.045, 0.175, 0.645, 0.545}

// renderizarPaginaGrafico decodifica, a partir do PDF em caminho, a página
// inteira (sem recorte) onde fica o gráfico de rentabilidade. A imagem
// devolvida é copiada pra memória própria do Go (não a estrutura nativa do
// PDFium) porque resultado.Cleanup() roda antes desta função retornar ao
// chamador — sem a cópia, os pixels seriam lidos depois de já liberados.
func (e *Extractor) renderizarPaginaGrafico(caminho string) (image.Image, error) {
	fileData, err := os.ReadFile(caminho)
	if err != nil {
		return nil, err
	}

	doc, err := e.instance.OpenDocument(&requests.OpenDocument{File: &fileData})
	if err != nil {
		return nil, fmt.Errorf("abrir PDF: %w", err)
	}
	defer e.instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	resultado, err := e.instance.RenderPageInDPI(&requests.RenderPageInDPI{
		Page: requests.Page{
			ByIndex: &requests.PageByIndex{
				Document: doc.Document,
				Index:    PaginaGraficoRentabilidade,
			},
		},
		DPI: 200,
	})
	if err != nil {
		return nil, fmt.Errorf("renderizar página: %w", err)
	}
	defer resultado.Cleanup()

	origem := resultado.Result.Image
	copia := image.NewRGBA(origem.Bounds())
	draw.Draw(copia, copia.Bounds(), origem, origem.Bounds().Min, draw.Src)
	return copia, nil
}

// RenderizarPaginaGraficoCompleta devolve a página inteira (sem recorte)
// já codificada em PNG — usada pela tela de configuração do recorte, pra o
// usuário ver onde a área selecionada cai dentro da página real.
func (e *Extractor) RenderizarPaginaGraficoCompleta(caminho string) ([]byte, error) {
	pagina, err := e.renderizarPaginaGrafico(caminho)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, pagina); err != nil {
		return nil, fmt.Errorf("codificar PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// RenderizarGraficoRentabilidade renderiza e recorta, a partir do PDF em
// caminho, a imagem do gráfico de rentabilidade histórica segundo
// `recorte` (frações da página), pronta pra copiar/compartilhar. Devolve
// os bytes já codificados em PNG.
func (e *Extractor) RenderizarGraficoRentabilidade(caminho string, recorte RecorteGrafico) ([]byte, error) {
	pagina, err := e.renderizarPaginaGrafico(caminho)
	if err != nil {
		return nil, err
	}

	b := pagina.Bounds()
	largura, altura := float64(b.Dx()), float64(b.Dy())
	recorteRect := image.Rect(
		int(recorte.X0*largura), int(recorte.Y0*altura),
		int(recorte.X1*largura), int(recorte.Y1*altura),
	)

	dst := image.NewRGBA(image.Rect(0, 0, recorteRect.Dx(), recorteRect.Dy()))
	draw.Draw(dst, dst.Bounds(), pagina, recorteRect.Min, draw.Src)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, fmt.Errorf("codificar PNG: %w", err)
	}
	return buf.Bytes(), nil
}
