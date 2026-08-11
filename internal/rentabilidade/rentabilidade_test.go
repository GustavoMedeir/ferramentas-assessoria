package rentabilidade

import (
	"os"
	"path/filepath"
	"testing"

	"rentabilidade/internal/pdfreport"
)

func abrirBancoTemp(t *testing.T) (string, func() error) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "rentabilidades.db")
	db, err := PrepararBanco(dbPath)
	if err != nil {
		t.Fatalf("PrepararBanco: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return dbPath, db.Close
}

func TestPrepararBancoIdempotente(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "rentabilidades.db")

	db1, err := PrepararBanco(dbPath)
	if err != nil {
		t.Fatalf("primeira PrepararBanco: %v", err)
	}
	db1.Close()

	// Reabrir o mesmo banco não deve falhar mesmo com as colunas de
	// migração já existindo.
	db2, err := PrepararBanco(dbPath)
	if err != nil {
		t.Fatalf("segunda PrepararBanco: %v", err)
	}
	defer db2.Close()

	registros, err := ListarRegistros(db2)
	if err != nil {
		t.Fatalf("ListarRegistros: %v", err)
	}
	if len(registros) != 0 {
		t.Errorf("esperava banco vazio, veio %d registros", len(registros))
	}
}

func TestMontarMensagem(t *testing.T) {
	r := Registro{
		Arquivo:       "XPerformance - 4312514 - Ref.08.07.pdf",
		GanhoMesReais: -8.51, RentMesPct: -0.12, CDIMesPct: -39.38,
		GanhoAnoReais: 649.69, RentAnoPct: 3.46, CDIAnoPct: 48.23,
	}
	msg := MontarMensagem(ModeloPadrao, r, "Gustavo Teste")
	// O modelo padrão tem "_CDI% do CDI" — como _CDI já formata com "%"
	// embutido (FormatarPercentual), o "%" literal do modelo soma, dando
	// "%%" no resultado. Mesmo comportamento do app original em Python.
	esperado := "Rentabilidade do mês: -R$ 8,51 (-0,12%, -39,38%% do CDI).\nNo ano: R$ 649,69 (3,46%, 48,23%% do CDI)."
	if msg != esperado {
		t.Errorf("MontarMensagem =\n%q\nesperado:\n%q", msg, esperado)
	}

	comNome := MontarMensagem("Olá _Nome, tudo bem?", r, "Gustavo Teste")
	if comNome != "Olá Gustavo Teste, tudo bem?" {
		t.Errorf("MontarMensagem com _Nome = %q, esperado inserir o nome", comNome)
	}

	// _NomeM é prefixo-sensível: "_Nome" não pode "roubar" o casamento e
	// deixar um "M" sobrando.
	comPrimeiroNome := MontarMensagem("Oi _NomeM, e _Nome também!", r, "GUSTAVO TESTE SILVA")
	if comPrimeiroNome != "Oi Gustavo, e GUSTAVO TESTE SILVA também!" {
		t.Errorf("MontarMensagem com _NomeM = %q, esperado primeiro nome capitalizado sem sobrar \"M\"", comPrimeiroNome)
	}
}

func TestMontarMensagemFestas(t *testing.T) {
	msg := MontarMensagemFestas("Feliz Natal, _NomeM! (assinado: _Nome)", "MARIA DA SILVA")
	esperado := "Feliz Natal, Maria! (assinado: MARIA DA SILVA)"
	if msg != esperado {
		t.Errorf("MontarMensagemFestas = %q, esperado %q", msg, esperado)
	}
}

func TestListarRegistrosOrdenacao(t *testing.T) {
	dir := t.TempDir()
	db, err := PrepararBanco(filepath.Join(dir, "db.sqlite"))
	if err != nil {
		t.Fatalf("PrepararBanco: %v", err)
	}
	defer db.Close()

	inserir := func(arquivo, data string) {
		_, err := db.Exec(`INSERT INTO rentabilidades (arquivo, data_referencia, ganho_mes_reais) VALUES (?, ?, 0)`, arquivo, data)
		if err != nil {
			t.Fatalf("inserir %s: %v", arquivo, err)
		}
	}
	inserir("c-sem-data.pdf", "")
	inserir("a-recente.pdf", "08/07/2026")
	inserir("b-antigo.pdf", "26/06/2026")

	registros, err := ListarRegistros(db)
	if err != nil {
		t.Fatalf("ListarRegistros: %v", err)
	}
	if len(registros) != 3 {
		t.Fatalf("esperava 3 registros, veio %d", len(registros))
	}
	ordem := []string{registros[0].Arquivo, registros[1].Arquivo, registros[2].Arquivo}
	esperado := []string{"b-antigo.pdf", "a-recente.pdf", "c-sem-data.pdf"}
	for i := range ordem {
		if ordem[i] != esperado[i] {
			t.Errorf("ordem[%d] = %q, esperado %q (ordem completa: %v)", i, ordem[i], esperado[i], ordem)
		}
	}
}

func TestListarRegistrosDeduplicaPorCodigo(t *testing.T) {
	dir := t.TempDir()
	db, err := PrepararBanco(filepath.Join(dir, "db.sqlite"))
	if err != nil {
		t.Fatalf("PrepararBanco: %v", err)
	}
	defer db.Close()

	inserir := func(arquivo, codigo, data string) {
		_, err := db.Exec(`INSERT INTO rentabilidades (arquivo, data_referencia, ganho_mes_reais, codigo) VALUES (?, ?, 0, ?)`, arquivo, data, codigo)
		if err != nil {
			t.Fatalf("inserir %s: %v", arquivo, err)
		}
	}
	// Mesmo código (4312514), duas datas de referência diferentes — só a
	// mais recente deve sobreviver, a outra some por completo.
	inserir("antigo.pdf", "4312514", "26/06/2026")
	inserir("recente.pdf", "4312514", "08/07/2026")
	// Código diferente, não deve ser afetado pela dedup.
	inserir("outro-cliente.pdf", "9999999", "08/07/2026")

	registros, err := ListarRegistros(db)
	if err != nil {
		t.Fatalf("ListarRegistros: %v", err)
	}
	if len(registros) != 2 {
		t.Fatalf("esperava 2 registros (1 descartado por duplicata), veio %d: %+v", len(registros), registros)
	}
	for _, r := range registros {
		if r.Arquivo == "antigo.pdf" {
			t.Errorf("registro antigo.pdf deveria ter sido descartado pela dedup, mas apareceu: %+v", r)
		}
	}
}

func TestListarClientes(t *testing.T) {
	registros := []Registro{
		{Arquivo: "a.pdf", Codigo: "111", Patrimonio: 5000},
		{Arquivo: "b.pdf", Codigo: "222", Patrimonio: 50000},
		{Arquivo: "c.pdf", Codigo: "999", Patrimonio: 1000}, // código não está na base
	}
	nomes := map[string]string{
		"111": "Zeca",    // com registro, menor patrimônio
		"222": "Ana",     // com registro, maior patrimônio
		"333": "Bruno",   // sem registro (sem PDF na pasta)
		"444": "Alberto", // sem registro (sem PDF na pasta)
	}

	clientes := ListarClientes(registros, nomes, nil)
	if len(clientes) != 5 {
		t.Fatalf("esperava 5 clientes, veio %d: %+v", len(clientes), clientes)
	}

	var codigos []string
	for _, c := range clientes {
		codigos = append(codigos, c.Codigo)
	}
	esperado := []string{"222", "111", "444", "333", "999"}
	for i := range esperado {
		if codigos[i] != esperado[i] {
			t.Errorf("ordem[%d] = %q, esperado %q (ordem completa: %v)", i, codigos[i], esperado[i], codigos)
		}
	}

	if clientes[0].Registro == nil || clientes[0].Nome != "Ana" {
		t.Errorf("primeiro cliente deveria ser Ana com registro anexado, veio %+v", clientes[0])
	}
	if clientes[4].Nome != "" || clientes[4].Registro == nil {
		t.Errorf("último cliente deveria ser o código 999 sem nome mas com registro, veio %+v", clientes[4])
	}
}

func TestExportarCSV(t *testing.T) {
	dir := t.TempDir()
	caminho := filepath.Join(dir, "saida.csv")
	registros := []Registro{
		{Arquivo: "XPerformance - 4312514 - Ref.08.07.pdf", Codigo: "4312514", GanhoMesReais: -8.51, RentMesPct: -0.12, CDIMesPct: -39.38, GanhoAnoReais: 649.69, RentAnoPct: 3.46, CDIAnoPct: 48.23},
	}
	if err := ExportarCSV(caminho, registros); err != nil {
		t.Fatalf("ExportarCSV: %v", err)
	}
	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		t.Fatalf("ler CSV: %v", err)
	}
	if conteudo[0] != 0xEF || conteudo[1] != 0xBB || conteudo[2] != 0xBF {
		t.Error("CSV sem BOM UTF-8")
	}
	texto := string(conteudo[3:])
	if !contains(texto, "4312514;-8,51;-0,12;-39,38;649,69;3,46;48,23") {
		t.Errorf("linha de dados inesperada:\n%s", texto)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

// TestProcessarPastaComPDFReal roda o fluxo completo (extração real via
// PDFium + gravação no SQLite) contra um PDF de relatório XP. Aponte
// PDFREPORT_TEST_PDF antes de rodar; pulado se a variável não estiver
// definida.
func TestProcessarPastaComPDFReal(t *testing.T) {
	origem := os.Getenv("PDFREPORT_TEST_PDF")
	if origem == "" {
		t.Skip("defina PDFREPORT_TEST_PDF com o caminho de um PDF de relatório XP para rodar este teste")
	}

	pasta := t.TempDir()
	dados, err := os.ReadFile(origem)
	if err != nil {
		t.Fatalf("ler PDF de origem: %v", err)
	}
	destino := filepath.Join(pasta, "XPerformance - 4312514 - Gustavo - Ref.08.07.pdf")
	if err := os.WriteFile(destino, dados, 0644); err != nil {
		t.Fatalf("copiar PDF: %v", err)
	}

	db, err := PrepararBanco(filepath.Join(pasta, "rentabilidades.db"))
	if err != nil {
		t.Fatalf("PrepararBanco: %v", err)
	}
	defer db.Close()

	extractor, err := pdfreport.NewExtractor()
	if err != nil {
		t.Fatalf("NewExtractor: %v", err)
	}
	defer extractor.Close()

	var chamadas int
	sucesso, falhas, err := ProcessarPasta(pasta, db, extractor, func(feitos, total int) { chamadas++ })
	if err != nil {
		t.Fatalf("ProcessarPasta: %v", err)
	}
	if len(falhas) != 0 {
		t.Fatalf("falhas inesperadas: %+v", falhas)
	}
	if sucesso != 1 {
		t.Fatalf("sucesso = %d, esperado 1", sucesso)
	}
	if chamadas == 0 {
		t.Error("callback de progresso nunca foi chamado")
	}

	// Rodar de novo não deve reprocessar (pula os já processados).
	sucesso2, _, err := ProcessarPasta(pasta, db, extractor, nil)
	if err != nil {
		t.Fatalf("segunda ProcessarPasta: %v", err)
	}
	if sucesso2 != 0 {
		t.Errorf("segunda passada reprocessou %d arquivo(s), esperado 0", sucesso2)
	}

	registros, err := ListarRegistros(db)
	if err != nil {
		t.Fatalf("ListarRegistros: %v", err)
	}
	if len(registros) != 1 {
		t.Fatalf("esperava 1 registro, veio %d", len(registros))
	}
	r := registros[0]
	if r.DataReferencia != "08/07/2026" {
		t.Errorf("DataReferencia = %q, esperado 08/07/2026", r.DataReferencia)
	}
	// O código vem do nome do arquivo (não mais do conteúdo do PDF — ver
	// codigoDoNomeArquivo), então é determinístico a partir do nome usado
	// acima ao copiar o PDF de teste pra "destino".
	if r.Codigo != "4312514" {
		t.Errorf("Codigo = %q, esperado 4312514 (extraído do nome do arquivo)", r.Codigo)
	}

	if err := MarcarCopiado(db, r.Arquivo); err != nil {
		t.Fatalf("MarcarCopiado: %v", err)
	}
	registros, _ = ListarRegistros(db)
	if !registros[0].Copiado {
		t.Error("registro deveria estar marcado como copiado")
	}

	if err := LimparBanco(db); err != nil {
		t.Fatalf("LimparBanco: %v", err)
	}
	registros, _ = ListarRegistros(db)
	if len(registros) != 0 {
		t.Errorf("esperava banco vazio após LimparBanco, veio %d", len(registros))
	}
}

func TestPrimeiroNomeCapitalizado(t *testing.T) {
	casos := []struct {
		nome, esperado string
	}{
		{"JOAO DA SILVA SANTOS", "Joao"},
		{"maria clara", "Maria"},
		{"ÁLVARO", "Álvaro"},
		{"  josé   ", "José"},
		{"ana", "Ana"},
		{"", ""},
		{"   ", "   "}, // nada pra extrair — devolve como veio
	}
	for _, c := range casos {
		if got := PrimeiroNomeCapitalizado(c.nome); got != c.esperado {
			t.Errorf("PrimeiroNomeCapitalizado(%q) = %q, esperado %q", c.nome, got, c.esperado)
		}
	}
}
