// Package rentabilidade cuida do banco de relatórios processados (um SQLite
// por pasta), da varredura de PDFs novos, da mensagem de WhatsApp a partir
// de um modelo com placeholders, e da exportação para CSV/Excel.
package rentabilidade

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	_ "modernc.org/sqlite"

	"rentabilidade/internal/pdfreport"
)

// ModeloPadrao é o texto inicial do modelo de mensagem, gravado em
// modelo_mensagem.txt na primeira vez que uma pasta é aberta.
const ModeloPadrao = "Rentabilidade do mês: _Rent (_Perc, _CDI% do CDI).\nNo ano: _RentA (_PercA, _CDIA% do CDI)."

// ModeloFestasPadrao é o texto inicial do modelo de mensagem de festas,
// gravado em modelo_festas.txt na primeira vez que uma pasta é aberta —
// não tem placeholders financeiros (_Rent, _Perc, ...) porque o Modo
// Festas também vale pra clientes sem relatório processado.
const ModeloFestasPadrao = "Feliz Natal, _Nome! Desejo a você e sua família um Natal repleto de paz e alegria, e um próspero Ano Novo."

// Registro é uma linha da tabela rentabilidades.
type Registro struct {
	Arquivo        string
	Codigo         string // código da conta, lido do nome do arquivo (ver codigoDoNomeArquivo)
	DataReferencia string // "dd/mm/aaaa" ou ""
	GanhoMesReais  float64
	GanhoAnoReais  float64
	RentMesPct     float64
	RentAnoPct     float64
	CDIMesPct      float64
	CDIAnoPct      float64
	Ganho12MReais  float64
	Rent12MPct     float64
	CDI12MPct      float64
	Patrimonio     float64 // usado só pra ordenar a lista de clientes
	Copiado        bool
}

// ---------------------------------------------------------------------------
// Banco
// ---------------------------------------------------------------------------

// colunasMigracao são colunas adicionadas depois da primeira versão do
// banco. CREATE TABLE já cobre bancos novos; pra bancos já existentes, cada
// ALTER roda uma vez e falha silenciosamente se a coluna já existir.
var colunasMigracao = []string{
	"copiado INTEGER NOT NULL DEFAULT 0",
	"ganho_ano_reais REAL NOT NULL DEFAULT 0",
	"rent_mes_pct REAL NOT NULL DEFAULT 0",
	"rent_ano_pct REAL NOT NULL DEFAULT 0",
	"cdi_mes_pct REAL NOT NULL DEFAULT 0",
	"cdi_ano_pct REAL NOT NULL DEFAULT 0",
	"codigo TEXT NOT NULL DEFAULT ''",
	"patrimonio REAL NOT NULL DEFAULT 0",
	"ganho_12m_reais REAL NOT NULL DEFAULT 0",
	"rent_12m_pct REAL NOT NULL DEFAULT 0",
	"cdi_12m_pct REAL NOT NULL DEFAULT 0",
}

// PrepararBanco abre (criando se preciso) o SQLite em dbPath e garante o
// schema atual.
func PrepararBanco(dbPath string) (*sql.DB, error) {
	// busy_timeout: se outro processo estiver com o banco travado (ex.: um
	// visualizador de SQLite aberto no arquivo), espera até 5s em vez de
	// falhar na hora com "database is locked".
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}
	db.SetMaxOpenConns(1) // modernc.org/sqlite: uma conexão evita "database is locked"

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS rentabilidades (
			arquivo TEXT PRIMARY KEY,
			data_referencia TEXT,
			ganho_mes_reais REAL NOT NULL,
			processado_em TEXT DEFAULT (datetime('now', 'localtime'))
		)
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("criar tabela: %w", err)
	}

	for _, coluna := range colunasMigracao {
		// Erro esperado e ignorado quando a coluna já existe (SQLite não
		// tem "ADD COLUMN IF NOT EXISTS").
		db.Exec("ALTER TABLE rentabilidades ADD COLUMN " + coluna)
	}

	// festas_enviados é independente de rentabilidades (chave por código do
	// cliente, não por arquivo) porque o Modo Festas manda mensagem pra
	// clientes com ou sem PDF processado.
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS festas_enviados (
			codigo TEXT PRIMARY KEY,
			enviado_em TEXT DEFAULT (datetime('now', 'localtime'))
		)
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("criar tabela festas_enviados: %w", err)
	}

	return db, nil
}

// Falha descreve um PDF que não pôde ser processado.
type Falha struct {
	Arquivo string
	Erro    string
}

// codigoArquivoRe casa a primeira sequência de 4+ dígitos no nome do
// arquivo. Nos relatórios baixados do XPerformance ("XPerformance -
// <código> - Ref.dd.mm.pdf") é sempre o código da conta — a data de
// referência no fim usa números de 1-2 dígitos, curtos demais pra bater
// aqui, então não há ambiguidade mesmo sem ancorar no formato exato do
// prefixo/sufixo.
var codigoArquivoRe = regexp.MustCompile(`\d{4,}`)

// codigoDoNomeArquivo extrai o código da conta do nome do arquivo, não do
// conteúdo do PDF: o nome já vem com o código na hora que o relatório é
// baixado, e ler dali é bem mais confiável do que casar um padrão de texto
// que varia de layout pra layout dentro do PDF (era a causa de falsos
// "código da conta não encontrado" em relatórios reais).
func codigoDoNomeArquivo(nomeArquivo string) (string, error) {
	codigo := codigoArquivoRe.FindString(nomeArquivo)
	if codigo == "" {
		return "", fmt.Errorf(`código da conta não encontrado no nome do arquivo (esperado algo como "XPerformance - 1234567 - Ref.dd.mm.pdf")`)
	}
	return codigo, nil
}

// ProcessarPasta lê os PDFs novos da pasta (pula os que já estão no banco) e
// atualiza o banco. Se progresso não for nil, é chamado como
// progresso(feitos, total) a cada arquivo — alimenta a barra de progresso na
// UI.
func ProcessarPasta(pasta string, db *sql.DB, extractor *pdfreport.Extractor, progresso func(feitos, total int)) (sucesso int, falhas []Falha, err error) {
	rows, err := db.Query("SELECT arquivo, codigo FROM rentabilidades")
	if err != nil {
		return 0, nil, fmt.Errorf("consultar processados: %w", err)
	}
	jaProcessados := map[string]bool{}
	for rows.Next() {
		var arquivo, codigo string
		if err := rows.Scan(&arquivo, &codigo); err != nil {
			rows.Close()
			return 0, nil, err
		}
		// Linhas gravadas antes da migração que adicionou "codigo" ficam
		// com "" nessa coluna — tratamos como não processadas pra
		// reprocessar e preencher o campo automaticamente (upsert abaixo),
		// sem exigir que o usuário limpe o banco e perca o status COPIADO.
		if codigo != "" {
			jaProcessados[arquivo] = true
		}
	}
	rows.Close()

	entradas, err := filepath.Glob(filepath.Join(pasta, "*.pdf"))
	if err != nil {
		return 0, nil, fmt.Errorf("listar PDFs: %w", err)
	}
	sort.Strings(entradas)

	var pdfsNovos []string
	for _, caminho := range entradas {
		if !jaProcessados[filepath.Base(caminho)] {
			pdfsNovos = append(pdfsNovos, caminho)
		}
	}

	for i, caminho := range pdfsNovos {
		nome := filepath.Base(caminho)
		codigo, errCodigo := codigoDoNomeArquivo(nome)
		if errCodigo != nil {
			falhas = append(falhas, Falha{Arquivo: nome, Erro: errCodigo.Error()})
		} else if dados, err := extractor.ExtrairDados(caminho); err != nil {
			falhas = append(falhas, Falha{Arquivo: nome, Erro: err.Error()})
		} else {
			// Upsert (não INSERT puro): cobre tanto arquivo novo quanto
			// reprocessamento de uma linha legada com "codigo" vazio (ver
			// comentário acima) — atualiza os campos extraídos sem mexer
			// em copiado/processado_em.
			_, execErr := db.Exec(`
				INSERT INTO rentabilidades
					(arquivo, data_referencia, ganho_mes_reais, ganho_ano_reais,
					 rent_mes_pct, rent_ano_pct, cdi_mes_pct, cdi_ano_pct, codigo, patrimonio,
					 ganho_12m_reais, rent_12m_pct, cdi_12m_pct)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(arquivo) DO UPDATE SET
					data_referencia=excluded.data_referencia,
					ganho_mes_reais=excluded.ganho_mes_reais,
					ganho_ano_reais=excluded.ganho_ano_reais,
					rent_mes_pct=excluded.rent_mes_pct,
					rent_ano_pct=excluded.rent_ano_pct,
					cdi_mes_pct=excluded.cdi_mes_pct,
					cdi_ano_pct=excluded.cdi_ano_pct,
					codigo=excluded.codigo,
					patrimonio=excluded.patrimonio,
					ganho_12m_reais=excluded.ganho_12m_reais,
					rent_12m_pct=excluded.rent_12m_pct,
					cdi_12m_pct=excluded.cdi_12m_pct
			`, nome, dados.DataReferencia, dados.GanhoMes, dados.GanhoAno,
				dados.RentMesPct, dados.RentAnoPct, dados.CDIMesPct, dados.CDIAnoPct,
				codigo, dados.PatrimonioTotalBruto,
				dados.Ganho12M, dados.Rent12MPct, dados.CDI12MPct)
			if execErr != nil {
				falhas = append(falhas, Falha{Arquivo: nome, Erro: execErr.Error()})
			} else {
				sucesso++
			}
		}
		if progresso != nil {
			progresso(i+1, len(pdfsNovos))
		}
	}

	return sucesso, falhas, nil
}

// LimparBanco apaga todos os registros (os PDFs continuam no disco e são
// reprocessados do zero na próxima leitura).
func LimparBanco(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM rentabilidades")
	return err
}

// MarcarCopiado marca um registro como copiado (badge COPIADO na lista).
func MarcarCopiado(db *sql.DB, arquivo string) error {
	_, err := db.Exec("UPDATE rentabilidades SET copiado = 1 WHERE arquivo = ?", arquivo)
	return err
}

// MarcarFestasEnviado registra que a mensagem de festas já foi
// copiada/enviada pro código dado (badge ENVIADO na lista, Modo Festas).
func MarcarFestasEnviado(db *sql.DB, codigo string) error {
	_, err := db.Exec(`
		INSERT INTO festas_enviados (codigo) VALUES (?)
		ON CONFLICT(codigo) DO UPDATE SET enviado_em = excluded.enviado_em
	`, codigo)
	return err
}

// ListarFestasEnviados devolve o conjunto de códigos que já receberam a
// mensagem de festas nesta leva (ver MarcarFestasEnviado / LimparFestasEnviados).
func ListarFestasEnviados(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT codigo FROM festas_enviados")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	enviados := map[string]bool{}
	for rows.Next() {
		var codigo string
		if err := rows.Scan(&codigo); err != nil {
			return nil, err
		}
		enviados[codigo] = true
	}
	return enviados, rows.Err()
}

// LimparFestasEnviados apaga o registro de quem já recebeu a mensagem de
// festas — usado antes de começar uma nova leva (ex.: Natal do ano
// seguinte). Não mexe na tabela rentabilidades.
func LimparFestasEnviados(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM festas_enviados")
	return err
}

// ListarRegistros devolve os registros ordenados por data de referência
// (mais antigo primeiro; sem data vai para o fim). Quando duas leituras da
// pasta encontram arquivos diferentes com o mesmo código de conta (ex.:
// relatório baixado de novo, arquivo renomeado), só o de Data de Referência
// mais recente é devolvido — o(s) mais antigo(s) fica(m) de fora por
// completo, inclusive da exportação CSV (que usa esta mesma função).
func ListarRegistros(db *sql.DB) ([]Registro, error) {
	rows, err := db.Query(`
		SELECT arquivo, data_referencia, ganho_mes_reais, ganho_ano_reais,
		       rent_mes_pct, rent_ano_pct, cdi_mes_pct, cdi_ano_pct, copiado,
		       codigo, patrimonio, ganho_12m_reais, rent_12m_pct, cdi_12m_pct
		FROM rentabilidades
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Registro
	for rows.Next() {
		var r Registro
		var dataReferencia sql.NullString
		var copiado int
		if err := rows.Scan(
			&r.Arquivo, &dataReferencia, &r.GanhoMesReais, &r.GanhoAnoReais,
			&r.RentMesPct, &r.RentAnoPct, &r.CDIMesPct, &r.CDIAnoPct, &copiado,
			&r.Codigo, &r.Patrimonio, &r.Ganho12MReais, &r.Rent12MPct, &r.CDI12MPct,
		); err != nil {
			return nil, err
		}
		r.DataReferencia = dataReferencia.String
		r.Copiado = copiado != 0
		todos = append(todos, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	registros := deduplicarPorCodigo(todos)

	sort.SliceStable(registros, func(i, j int) bool {
		return chaveOrdenacao(registros[i]) < chaveOrdenacao(registros[j])
	})
	return registros, nil
}

// deduplicarPorCodigo mantém, pra cada código de conta, só o registro de
// Data de Referência mais recente (empate: fica o último visto — não
// importa qual). Registros sem código (banco de antes desta migração,
// reprocessados automaticamente na próxima leitura da pasta — ver
// ProcessarPasta) passam direto, sem entrar nessa redução.
func deduplicarPorCodigo(todos []Registro) []Registro {
	melhorPorCodigo := map[string]Registro{}
	var semCodigo []Registro
	for _, r := range todos {
		if r.Codigo == "" {
			semCodigo = append(semCodigo, r)
			continue
		}
		atual, existe := melhorPorCodigo[r.Codigo]
		if !existe || chaveData(r.DataReferencia) >= chaveData(atual.DataReferencia) {
			melhorPorCodigo[r.Codigo] = r
		}
	}

	out := semCodigo
	for _, r := range melhorPorCodigo {
		out = append(out, r)
	}
	return out
}

// chaveData converte "dd/mm/aaaa" numa chave comparável (ano, mês, dia) —
// não dá pra ordenar/comparar direto por string. Data ausente ou mal
// formada vai para o fim.
func chaveData(dataReferencia string) string {
	if dataReferencia != "" {
		partes := strings.Split(dataReferencia, "/")
		if len(partes) == 3 {
			dia, errD := strconv.Atoi(partes[0])
			mes, errM := strconv.Atoi(partes[1])
			ano, errA := strconv.Atoi(partes[2])
			if errD == nil && errM == nil && errA == nil {
				return fmt.Sprintf("%04d-%02d-%02d", ano, mes, dia)
			}
		}
	}
	return "9999-99-99"
}

// chaveOrdenacao converte um registro numa chave comparável por data (mais
// antigo primeiro), desempatando por nome de arquivo.
func chaveOrdenacao(r Registro) string {
	return chaveData(r.DataReferencia) + "-" + r.Arquivo
}

// ---------------------------------------------------------------------------
// Lista de clientes (aba Rentabilidade)
// ---------------------------------------------------------------------------

// ClienteRentabilidade é uma linha da lista de clientes da aba
// Rentabilidade: todo código da base de clientes vira uma linha, mesmo sem
// relatório processado (Registro fica nil nesse caso, ou quando o arquivo
// deu erro no processamento) — a linha aparece "em branco".
type ClienteRentabilidade struct {
	Codigo        string
	Nome          string // "" quando o código veio de um PDF sem cliente correspondente na base
	Registro      *Registro
	FestasEnviado bool // Modo Festas: já recebeu a mensagem de festas nesta leva
}

// ListarClientes junta a base de clientes (código->nome, mesmo mapa
// exposto ao frontend) com os registros já processados (ver
// ListarRegistros — um por código, sem duplicatas): cada código da base
// vira uma linha, com o registro correspondente anexado quando existe;
// códigos de registros que não estão na base entram como linhas extras no
// fim, sem nome.
//
// Ordenação final: (a) clientes com registro, por Patrimonio decrescente
// (maior primeiro); (b) clientes da base sem registro (sem PDF, ou PDF com
// falha), por Nome alfabético; (c) registros sem cliente reconhecido na
// base, por Codigo.
func ListarClientes(registros []Registro, nomes map[string]string, festasEnviados map[string]bool) []ClienteRentabilidade {
	porCodigo := make(map[string]Registro, len(registros))
	for _, r := range registros {
		porCodigo[r.Codigo] = r
	}

	comCliente := make([]ClienteRentabilidade, 0, len(nomes))
	usados := map[string]bool{}
	for codigo, nome := range nomes {
		item := ClienteRentabilidade{Codigo: codigo, Nome: nome, FestasEnviado: festasEnviados[codigo]}
		if r, ok := porCodigo[codigo]; ok {
			rCopia := r
			item.Registro = &rCopia
			usados[codigo] = true
		}
		comCliente = append(comCliente, item)
	}

	var semCliente []ClienteRentabilidade
	for codigo, r := range porCodigo {
		if codigo == "" || usados[codigo] {
			continue
		}
		rCopia := r
		semCliente = append(semCliente, ClienteRentabilidade{Codigo: codigo, Registro: &rCopia, FestasEnviado: festasEnviados[codigo]})
	}

	sort.SliceStable(comCliente, func(i, j int) bool {
		a, b := comCliente[i], comCliente[j]
		aTem, bTem := a.Registro != nil, b.Registro != nil
		if aTem != bTem {
			return aTem // com registro vem antes de sem registro
		}
		if aTem {
			return a.Registro.Patrimonio > b.Registro.Patrimonio // maior primeiro
		}
		return strings.ToLower(a.Nome) < strings.ToLower(b.Nome)
	})

	sort.SliceStable(semCliente, func(i, j int) bool {
		return semCliente[i].Codigo < semCliente[j].Codigo
	})

	return append(comCliente, semCliente...)
}

// ---------------------------------------------------------------------------
// Placeholders do modelo de mensagem
// ---------------------------------------------------------------------------

// placeholders na ordem em que a regex deve tentar casar: "_Rent" é prefixo
// de "_RentA" e "_Rent12M" (idem "_Perc"/"_CDI", e agora "_Nome"/"_NomeM")
// — por isso vai do mais longo pro mais curto, senão "_Rent12M"/"_RentA"
// virariam "<valor>12M"/"<valor>A", e "_NomeM" viraria "<nome>M".
var placeholders = []string{
	"_RentA", "_Rent12M", "_Rent",
	"_PercA", "_Perc12M", "_Perc",
	"_CDIA", "_CDI12M", "_CDI",
	"_NomeM", "_Nome",
}

var placeholderRe = regexp.MustCompile(strings.Join(func() []string {
	escaped := make([]string, len(placeholders))
	for i, p := range placeholders {
		escaped[i] = regexp.QuoteMeta(p)
	}
	return escaped
}(), "|"))

// ValoresPlaceholder devolve o valor formatado de cada placeholder pro
// registro dado. nome vem da base de clientes (carregada à parte, por
// código da conta) — pode vir vazio se o cliente não estiver na base.
// "_NomeM" é o primeiro nome capitalizado (ver PrimeiroNomeCapitalizado) —
// útil pra mensagens mais informais, sem precisar reescrever o modelo pra
// cada cliente.
func ValoresPlaceholder(r Registro, nome string) map[string]string {
	return map[string]string{
		"_Rent":    pdfreport.FormatarReais(r.GanhoMesReais),
		"_RentA":   pdfreport.FormatarReais(r.GanhoAnoReais),
		"_Rent12M": pdfreport.FormatarReais(r.Ganho12MReais),
		"_Perc":    pdfreport.FormatarPercentual(r.RentMesPct),
		"_PercA":   pdfreport.FormatarPercentual(r.RentAnoPct),
		"_Perc12M": pdfreport.FormatarPercentual(r.Rent12MPct),
		"_CDI":     pdfreport.FormatarPercentual(r.CDIMesPct),
		"_CDIA":    pdfreport.FormatarPercentual(r.CDIAnoPct),
		"_CDI12M":  pdfreport.FormatarPercentual(r.CDI12MPct),
		"_Nome":    nome,
		"_NomeM":   PrimeiroNomeCapitalizado(nome),
	}
}

// MontarMensagem substitui os placeholders do modelo pelos valores do
// registro e o nome do cliente.
func MontarMensagem(template string, r Registro, nome string) string {
	valores := ValoresPlaceholder(r, nome)
	return placeholderRe.ReplaceAllStringFunc(template, func(m string) string {
		return valores[m]
	})
}

// MontarMensagemFestas substitui "_Nome"/"_NomeM" do modelo — usado pelo
// Modo Festas, que também vale pra clientes sem relatório processado (sem
// dados financeiros pra preencher os outros placeholders). "_NomeM" precisa
// ser trocado ANTES de "_Nome" — é prefixo dele, e strings.ReplaceAll na
// ordem errada faria "_NomeM" virar "<nome>M".
func MontarMensagemFestas(template, nome string) string {
	texto := strings.ReplaceAll(template, "_NomeM", PrimeiroNomeCapitalizado(nome))
	return strings.ReplaceAll(texto, "_Nome", nome)
}

// PrimeiroNomeCapitalizado devolve só o primeiro nome, com a primeira letra
// maiúscula e o resto minúsculo (ex.: "JOAO DA SILVA" -> "Joao") — usado
// pela preferência "Só o primeiro nome" (Configurações > Rentabilidade). O
// cadastro costuma vir todo em maiúsculas do CSV da base de clientes (ver
// internal/clientdb), sem nenhuma normalização de capitalização.
func PrimeiroNomeCapitalizado(nome string) string {
	campos := strings.Fields(nome)
	if len(campos) == 0 {
		return nome
	}
	letras := []rune(strings.ToLower(campos[0]))
	letras[0] = unicode.ToUpper(letras[0])
	return string(letras)
}

// ---------------------------------------------------------------------------
// Exportação para Excel (CSV)
// ---------------------------------------------------------------------------

var colunasExportacao = []string{
	"Código", "Ganho Mês (R$)", "Rentabilidade Mês (%)", "% CDI Mês",
	"Ganho Ano (R$)", "Rentabilidade Ano (%)", "% CDI Ano",
	"Ganho 12M (R$)", "Rentabilidade 12M (%)", "% CDI 12M",
}

// numeroExcel formata como "1234,56" (vírgula decimal, sem R$/% e sem ponto
// de milhar) — é assim que o Excel pt-BR reconhece a célula como número.
func numeroExcel(valor float64) string {
	return strings.Replace(strconv.FormatFloat(valor, 'f', 2, 64), ".", ",", 1)
}

func linhasExportacao(registros []Registro) [][]string {
	linhas := [][]string{colunasExportacao}
	for _, r := range registros {
		linhas = append(linhas, []string{
			r.Codigo,
			numeroExcel(r.GanhoMesReais),
			numeroExcel(r.RentMesPct),
			numeroExcel(r.CDIMesPct),
			numeroExcel(r.GanhoAnoReais),
			numeroExcel(r.RentAnoPct),
			numeroExcel(r.CDIAnoPct),
			numeroExcel(r.Ganho12MReais),
			numeroExcel(r.Rent12MPct),
			numeroExcel(r.CDI12MPct),
		})
	}
	return linhas
}

// ExportarCSV grava a planilha de rentabilidades em caminho. Separador ";"
// e UTF-8 com BOM: é a combinação que o Excel brasileiro abre com duplo
// clique já separado em colunas e com acentos corretos (com "," ele joga
// tudo numa coluna só, porque "," aqui é o separador decimal).
func ExportarCSV(caminho string, registros []Registro) error {
	f, err := os.Create(caminho)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM UTF-8
	w := csv.NewWriter(f)
	w.Comma = ';'
	if err := w.WriteAll(linhasExportacao(registros)); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
