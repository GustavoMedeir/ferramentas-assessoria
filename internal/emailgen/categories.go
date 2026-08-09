package emailgen

import "fmt"

// Categories — Gerador de E-mails de Ordem: modelos (Renda Fixa, Tesouro
// Direto, Fundos, COE, Ofertas Públicas, Subscrição, Operações
// Societárias, Clubes, Ações/Opções/Termo, Movimentação de Recursos,
// Compromissadas, Carteira Automatizada). Cada categoria tem: grupo
// (produto), rótulo (operação), campos do formulário, corpo do e-mail
// (função de v -> texto), anexo opcional e frase de fechamento. Porte 1:1
// das linhas 391-927 de rentabilidade_app_v2.py.
var Categories = []Category{

	// ---------------- RENDA FIXA ----------------
	// Material de compliance (Res. CVM 178/23): Aplicação e Aplicação
	// Cotada viraram um modelo só (idem Resgate), com Taxa de Negociação,
	// Carência e PU no lugar das taxas mínima/máxima separadas.
	{
		ID: "rf-aplicacao", Group: "Renda Fixa", Label: "Aplicação", Tipo: "aplicacao",
		IntroFrase: "aplicação",
		Fields: []Field{
			{Key: "ativo", Label: "Ativo", Placeholder: "Nome do ativo"},
			{Key: "emissor", Label: "Emissor", Placeholder: "Nome do emissor"},
			{Key: "taxaNegociacao", Label: "Taxa de Negociação", Placeholder: "Taxa cotada ou taxa mínima que o cliente aceita comprar"},
			{Key: "carencia", Label: "Carência", Placeholder: "Data, nº de dias a partir da aplicação ou No vencimento"},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "Data ou nº de dias a partir da aplicação"},
			{Key: "valorLimite", Label: "Valor limite a ser aplicado", Placeholder: "Valor financeiro limite ou valor cotado"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ativo: %s;\nEmissor: %s;\nTaxa de Negociação: %s;\nCarência: %s;\nVencimento: %s;\nValor limite a ser aplicado: %s.",
				v["ativo"], v["emissor"], v["taxaNegociacao"], v["carencia"], v["vencimento"], v["valorLimite"])
		},
		Fechamento: "Aguardo confirmação para realizar a aplicação.",
	},
	{
		ID: "rf-resgate", Group: "Renda Fixa", Label: "Resgate", Tipo: "resgate",
		IntroFrase: "o resgate",
		Fields: []Field{
			{Key: "ativo", Label: "Ativo", Placeholder: "Nome do ativo"},
			{Key: "emissor", Label: "Emissor", Placeholder: "Nome do emissor"},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "MÊS/ANO"},
			{Key: "pu", Label: "Preço Unitário do Título (PU)", Placeholder: "PU da venda negociado no mercado na data cotada"},
			{Key: "valorOuQtd", Label: "Valor a ser resgatado", Placeholder: "R$ ..., Quantidade: xxx ou Resgate Total"},
			{Key: "valorCotado", Label: "Valor financeiro cotado", Placeholder: "Ex: 1.000,00 (só se for resgate cotado)", Optional: true, Moeda: true},
		},
		Body: func(v map[string]string) string {
			texto := fmt.Sprintf("Ativo: %s;\nEmissor: %s;\nVencimento: %s;\nPreço Unitário do Título (\"PU\"): %s;\nValor a ser resgatado: %s",
				v["ativo"], v["emissor"], v["vencimento"], v["pu"], v["valorOuQtd"])
			if v["valorCotado"] != "" {
				texto += fmt.Sprintf(";\nValor financeiro cotado: %s", v["valorCotado"])
			}
			return texto + "."
		},
		Fechamento: "Aguardo confirmação para realizar o resgate.",
	},

	// ---------------- TESOURO DIRETO ----------------
	{
		ID: "td-aplicacao", Group: "Tesouro Direto", Label: "Aplicação", Tipo: "aplicacao",
		IntroFrase: "a aplicação no Tesouro Direto",
		Fields: []Field{
			{Key: "titulo", Label: "Título", Placeholder: "Nome do título"},
			{Key: "rentabilidade", Label: "Rentabilidade ao ano", Placeholder: "Taxa ou indexador"},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "Data de vencimento"},
			{Key: "pagamentoJuros", Label: "Pagamento de juros", Placeholder: "No vencimento ou Cupons semestrais"},
			{Key: "valorOuQtd", Label: "Valor a ser aplicado ou Quantidade", Placeholder: "R$ ... ou Quantidade: xxx"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Título: %s;\nRentabilidade ao ano: %s;\nVencimento: %s;\nPagamento de juros: %s;\nValor a ser aplicado: %s.",
				v["titulo"], v["rentabilidade"], v["vencimento"], v["pagamentoJuros"], v["valorOuQtd"])
		},
		Fechamento: "Aguardo confirmação para realizar a aplicação.",
	},
	{
		ID: "td-resgate", Group: "Tesouro Direto", Label: "Resgate", Tipo: "resgate",
		IntroFrase: "o resgate no Tesouro Direto",
		Fields: []Field{
			{Key: "titulo", Label: "Título", Placeholder: "Nome do título"},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "Data de vencimento"},
			{Key: "valorOuQtd", Label: "Valor a ser resgatado", Placeholder: "R$ ..., Quantidade: xxx ou Resgate Total"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Título: %s;\nVencimento: %s;\nValor a ser resgatado: %s.",
				v["titulo"], v["vencimento"], v["valorOuQtd"])
		},
		Fechamento: "Aguardo confirmação para realizar o resgate.",
	},

	// ---------------- FUNDOS ----------------
	{
		ID: "fundo-aplicacao", Group: "Fundos de Investimento", Label: "Aplicação", Tipo: "aplicacao",
		IntroFrase: "a aplicação no Fundo de Investimentos",
		Fields: []Field{
			{Key: "fundo", Label: "Fundo", Placeholder: "Nome do fundo"},
			{Key: "valor", Label: "Valor a ser aplicado", Placeholder: "Ex: 1.000,00", Moeda: true},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Fundo: %s;\nValor a ser aplicado: %s.", v["fundo"], v["valor"])
		},
		Fechamento: "Aguardo confirmação para realizar a aplicação.",
	},
	{
		ID: "fundo-resgate", Group: "Fundos de Investimento", Label: "Resgate", Tipo: "resgate",
		IntroFrase: "o resgate do Fundo de Investimentos",
		Fields: []Field{
			{Key: "fundo", Label: "Fundo", Placeholder: "Nome do fundo"},
			{Key: "valor", Label: "Valor a ser resgatado", Placeholder: "R$ ... ou Resgate Total"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Fundo: %s;\nValor a ser resgatado: %s.", v["fundo"], v["valor"])
		},
		Fechamento: "Aguardo confirmação para realizar o resgate.",
	},

	// ---------------- COE ----------------
	{
		ID: "coe-reserva", Group: "COE", Label: "Reserva de Compra", Tipo: "outro",
		IntroFrase: "a reserva de compra do COE",
		Fields: []Field{
			{Key: "coe", Label: "COE", Placeholder: "Nome do COE"},
			{Key: "emissor", Label: "Emissor", Placeholder: "Emissor do COE"},
			{Key: "tipo", Label: "Tipo", Type: "select", Options: []string{"Valor Nominal Protegido", "Em Risco"}},
			{Key: "quantidade", Label: "Quantidade a ser reservada", Placeholder: "xxx"},
			{Key: "valor", Label: "Valor da reserva", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "xx/xx/xxxx"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("COE: %s;\nEmissor: %s;\nTipo: %s;\nQuantidade: %s;\nValor: %s;\nVencimento: %s.",
				v["coe"], v["emissor"], v["tipo"], v["quantidade"], v["valor"], v["vencimento"])
		},
		Fechamento: "Aguardo confirmação para realizar a reserva.",
	},
	{
		ID: "coe-venda", Group: "COE", Label: "Venda", Tipo: "outro",
		IntroFrase: "a venda do COE",
		Fields: []Field{
			{Key: "coe", Label: "COE", Placeholder: "Nome do COE"},
			{Key: "emissor", Label: "Emissor", Placeholder: "Emissor do COE"},
			{Key: "tipo", Label: "Tipo", Type: "select", Options: []string{"Valor Nominal Protegido", "Em Risco"}},
			{Key: "quantidade", Label: "Quantidade a ser resgatada", Placeholder: "xxx"},
			{Key: "valor", Label: "Valor bruto do resgate", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "xx/xx/xxxx"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("COE: %s;\nEmissor: %s;\nTipo: %s;\nQuantidade: %s;\nValor: %s;\nVencimento: %s.",
				v["coe"], v["emissor"], v["tipo"], v["quantidade"], v["valor"], v["vencimento"])
		},
		Fechamento: "Aguardo confirmação para realizar a venda.",
	},

	// ---------------- OFERTAS ----------------
	{
		ID: "oferta-publica-rf", Group: "Ofertas Públicas", Label: "Oferta Pública RF - Reserva", Tipo: "outro",
		IntroFrase: "a reserva na Oferta Pública",
		Fields: []Field{
			{Key: "ativo", Label: "Ativo", Placeholder: "Nome do ativo"},
			{Key: "emissor", Label: "Emissor", Placeholder: "Nome do emissor"},
			{Key: "taxa", Label: "Taxa mínima", Placeholder: "Taxa mínima"},
			{Key: "quantidade", Label: "Quantidade a ser reservada", Placeholder: "se aplicável"},
			{Key: "valor", Label: "Valor da reserva", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "carencia", Label: "Carência", Placeholder: "xx/xx/xxxx ou No vencimento"},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "xx/xx/xxxx"},
			{Key: "vinculada", Label: "Cliente pessoa vinculada", Type: "select", Options: []string{"Não", "Sim"}},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ativo: %s;\nEmissor: %s;\nTaxa: %s;\nQuantidade: %s;\nValor: %s;\nCarência: %s;\nVencimento: %s.\nCliente pessoa vinculada: %s.",
				v["ativo"], v["emissor"], v["taxa"], v["quantidade"], v["valor"], v["carencia"], v["vencimento"], v["vinculada"])
		},
		Anexo:      "Enviar em anexo ao e-mail o Prospecto oficial em PDF da Oferta Pública.",
		Fechamento: "Aguardo confirmação para realizar a reserva.",
	},
	{
		ID: "ipo-reserva", Group: "Ofertas Públicas", Label: "IPO - Reserva", Tipo: "outro",
		IntroFrase: "a reserva na Oferta Pública",
		Fields: []Field{
			{Key: "oferta", Label: "Oferta", Placeholder: "Nome da oferta"},
			{Key: "valor", Label: "Valor da reserva", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "preco", Label: "Preço limite (se desejado)", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "vinculada", Label: "Cliente pessoa vinculada", Type: "select", Options: []string{"Não", "Sim"}},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Oferta: %s;\nValor: %s;\nPreço: %s.\nCliente pessoa vinculada: %s.",
				v["oferta"], v["valor"], v["preco"], v["vinculada"])
		},
		Anexo:      "Enviar em anexo ao e-mail o Prospecto oficial em PDF da Oferta Pública.",
		Fechamento: "Aguardo confirmação para realizar a reserva.",
	},
	{
		ID: "oferta-restrita", Group: "Ofertas Públicas", Label: "Oferta Restrita - Reserva", Tipo: "outro",
		IntroFrase: "a reserva na Oferta Restrita",
		Fields: []Field{
			{Key: "oferta", Label: "Oferta", Placeholder: "Nome da oferta"},
			{Key: "quantidade", Label: "Quantidade a ser reservada", Placeholder: "xxx"},
			{Key: "preco", Label: "Preço limite (se desejado)", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "vinculada", Label: "Cliente pessoa vinculada", Type: "select", Options: []string{"Não", "Sim"}},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Oferta: %s;\nQuantidade: %s;\nPreço: %s.\nCliente pessoa vinculada: %s.",
				v["oferta"], v["quantidade"], v["preco"], v["vinculada"])
		},
		Anexo:      "Enviar em anexo ao e-mail o documento oficial de \"Fato Relevante\" em PDF da Oferta.",
		Fechamento: "Aguardo confirmação para realizar a reserva.",
	},
	{
		ID: "opa-venda", Group: "Ofertas Públicas", Label: "OPA - Venda", Tipo: "outro",
		IntroFrase: "a venda na OPA",
		Fields: []Field{
			{Key: "nomeOpa", Label: "Nome da OPA", Placeholder: "Nome da OPA"},
			{Key: "ativo", Label: "Ativo", Placeholder: "Código ou nome do papel"},
			{Key: "quantidade", Label: "Quantidade a ser vendida", Placeholder: "xxx"},
			{Key: "preco", Label: "Preço mínimo do ativo", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "vinculada", Label: "Cliente pessoa vinculada", Type: "select", Options: []string{"Não", "Sim"}},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Nome: %s;\nAtivo: %s;\nQuantidade: %s;\nPreço: %s.\nCliente pessoa vinculada: %s.",
				v["nomeOpa"], v["ativo"], v["quantidade"], v["preco"], v["vinculada"])
		},
		Anexo:      "Enviar em anexo ao e-mail o Edital oficial da OPA em PDF.",
		Fechamento: "Aguardo confirmação para realizar a operação.",
	},

	// ---------------- SUBSCRIÇÃO ----------------
	{
		ID: "subscricao-exercicio", Group: "Subscrição", Label: "Exercício", Tipo: "outro",
		IntroFrase: "o exercício do direito de subscrição",
		Fields: []Field{
			{Key: "acao", Label: "Ação a ser subscrita", Placeholder: "Código da ação"},
			{Key: "codigoDireito", Label: "Código do direito de subscrição", Placeholder: "Código do direito"},
			{Key: "preco", Label: "Preço unitário", Placeholder: "Ex: 1.000,00 por ação", Moeda: true},
			{Key: "quantidade", Label: "Quantidade a ser subscrita", Placeholder: "xxx"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ação: %s;\nCódigo do direito de subscrição: %s;\nAo Preço unitário de: %s;\nQuantidade: %s.",
				v["acao"], v["codigoDireito"], v["preco"], v["quantidade"])
		},
		Anexo:      "Enviar em anexo ao e-mail o documento oficial em PDF de Aviso aos Acionistas.",
		Fechamento: "Aguardo confirmação para realizar a operação.",
	},
	{
		ID: "subscricao-negociacao", Group: "Subscrição", Label: "Negociação de Direitos", Tipo: "outro",
		IntroFrase: "a negociação de direitos de subscrição",
		Fields: []Field{
			{Key: "codigo", Label: "Código do direito de subscrição", Placeholder: "Código"},
			{Key: "natureza", Label: "Natureza", Type: "select", Options: []string{"Compra", "Venda"}},
			{Key: "preco", Label: "Preço limite", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "quantidade", Label: "Quantidade a ser negociada", Placeholder: "xxx"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Código: %s;\nNatureza: %s;\nPreço: %s;\nQuantidade: %s.",
				v["codigo"], v["natureza"], v["preco"], v["quantidade"])
		},
		Fechamento: "Aguardo confirmação para realizar a operação.",
	},

	// ---------------- OUTRAS OPERAÇÕES SOCIETÁRIAS ----------------
	{
		ID: "conversao-voluntaria", Group: "Operações Societárias", Label: "Conversão Voluntária", Tipo: "outro",
		IntroFrase: "a conversão",
		Fields: []Field{
			{Key: "papelConvertido", Label: "Papel a ser convertido", Placeholder: "Código do ativo"},
			{Key: "qtdConvertida", Label: "Quantidade a ser convertida", Placeholder: "xxx"},
			{Key: "papelApos", Label: "Papel após conversão", Placeholder: "Código do ativo"},
			{Key: "qtdApos", Label: "Quantidade após conversão", Placeholder: "xxx"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Papel a ser convertido: %s;\nQuantidade a ser convertida: %s;\nPapel após conversão: %s;\nQuantidade após conversão: %s.",
				v["papelConvertido"], v["qtdConvertida"], v["papelApos"], v["qtdApos"])
		},
		Fechamento: "Aguardo confirmação para prosseguir com a conversão.",
	},
	{
		ID: "dissidencia", Group: "Operações Societárias", Label: "Dissidência", Tipo: "outro",
		IntroFrase: "a dissidência",
		Fields: []Field{
			{Key: "acao", Label: "Ação a ser entregue", Placeholder: "Código do ativo"},
			{Key: "quantidade", Label: "Quantidade", Placeholder: "xxx"},
			{Key: "valorReembolso", Label: "Valor de reembolso unitário", Placeholder: "Ex: 1.000,00 por ação", Moeda: true},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ação a ser entregue: %s;\nQuantidade: %s;\nAo valor de reembolso unitário de: %s.",
				v["acao"], v["quantidade"], v["valorReembolso"])
		},
		Fechamento: "Aguardo confirmação para prosseguir com a operação.",
	},

	// ---------------- CLUBES ----------------
	{
		ID: "clube-aplicacao", Group: "Clubes de Investimento", Label: "Aplicação", Tipo: "aplicacao",
		IntroFrase: "a aplicação no clube de investimentos",
		Fields: []Field{
			{Key: "nomeClube", Label: "Nome do Clube", Placeholder: "Nome"},
			{Key: "codigoClube", Label: "Código do Clube", Placeholder: "Código"},
			{Key: "valor", Label: "Valor a ser aplicado", Placeholder: "Ex: 1.000,00", Moeda: true},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Nome do Clube: %s;\nCódigo: %s;\nValor a ser aplicado: %s.",
				v["nomeClube"], v["codigoClube"], v["valor"])
		},
		Fechamento: "Aguardo confirmação para realizar a aplicação.",
	},
	{
		ID: "clube-resgate", Group: "Clubes de Investimento", Label: "Resgate", Tipo: "resgate",
		IntroFrase: "o resgate no clube de investimentos",
		Fields: []Field{
			{Key: "nomeClube", Label: "Nome do Clube", Placeholder: "Nome"},
			{Key: "codigoClube", Label: "Código do Clube", Placeholder: "Código"},
			{Key: "valor", Label: "Valor do resgate", Placeholder: "Ex: 1.000,00", Moeda: true},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Nome do Clube: %s;\nCódigo: %s;\nValor do resgate: %s.",
				v["nomeClube"], v["codigoClube"], v["valor"])
		},
		Fechamento: "Aguardo confirmação para realizar o resgate.",
	},

	// ---------------- RENDA VARIÁVEL ----------------
	{
		ID: "acoes-compra", Group: "Ações, Futuros e FIIs", Label: "Compra", Tipo: "outro",
		IntroFrase: "a compra",
		Fields: []Field{
			{Key: "ativo", Label: "Ativo", Placeholder: "Código do ativo"},
			{Key: "quantidade", Label: "Quantidade", Placeholder: "xxxx"},
			{Key: "preco", Label: "Preço", Placeholder: "Preço ou 'Ordem a mercado'"},
			{Key: "validade", Label: "Validade da ordem", Placeholder: "xx/xx/xxxx ou 'Até Cancelar' (só se não for a mercado)"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ativo: %s;\nQuantidade: %s;\nPreço: %s;\nValidade da ordem: %s.",
				v["ativo"], v["quantidade"], v["preco"], v["validade"])
		},
		Fechamento: "Aguardo confirmação para realizar a compra.",
	},
	{
		ID: "acoes-venda", Group: "Ações, Futuros e FIIs", Label: "Venda", Tipo: "outro",
		IntroFrase: "a venda",
		Fields: []Field{
			{Key: "ativo", Label: "Ativo", Placeholder: "Código do ativo"},
			{Key: "quantidade", Label: "Quantidade", Placeholder: "xxxx"},
			{Key: "preco", Label: "Preço", Placeholder: "Preço ou 'Ordem a mercado'"},
			{Key: "validade", Label: "Validade da ordem", Placeholder: "xx/xx/xxxx ou 'Até Cancelar' (só se não for a mercado)"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ativo: %s;\nQuantidade: %s;\nPreço: %s;\nValidade da ordem: %s.",
				v["ativo"], v["quantidade"], v["preco"], v["validade"])
		},
		Fechamento: "Aguardo confirmação para realizar a venda.",
	},
	{
		ID: "opcoes-compra", Group: "Opções", Label: "Compra", Tipo: "outro",
		IntroFrase: "a compra de opção",
		Fields: []Field{
			{Key: "ativoRef", Label: "Ativo Referência", Placeholder: "Código do ativo"},
			{Key: "opcao", Label: "Opção", Placeholder: "Código da opção"},
			{Key: "tipo", Label: "Tipo", Type: "select", Options: []string{"Call", "Put"}},
			{Key: "strike", Label: "Preço de Exercício", Placeholder: "Preço de strike"},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "xx/xx/xxxx"},
			{Key: "quantidade", Label: "Quantidade", Placeholder: "xxxx"},
			{Key: "preco", Label: "Preço", Placeholder: "Preço ou 'Ordem a mercado'"},
			{Key: "validade", Label: "Validade da ordem", Placeholder: "xx/xx/xxxx ou 'Até Cancelar'"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ativo Referência: %s;\nOpção: %s;\nTipo: %s;\nPreço de Exercício: %s;\nVencimento: %s;\nQuantidade: %s;\nPreço: %s;\nValidade da ordem: %s.",
				v["ativoRef"], v["opcao"], v["tipo"], v["strike"], v["vencimento"], v["quantidade"], v["preco"], v["validade"])
		},
		Fechamento: "Aguardo confirmação para realizar a compra.",
	},
	{
		ID: "opcoes-venda", Group: "Opções", Label: "Venda", Tipo: "outro",
		IntroFrase: "a venda de opção",
		Fields: []Field{
			{Key: "ativoRef", Label: "Ativo Referência", Placeholder: "Código do ativo"},
			{Key: "opcao", Label: "Opção", Placeholder: "Código da opção"},
			{Key: "tipo", Label: "Tipo", Type: "select", Options: []string{"Call", "Put"}},
			{Key: "strike", Label: "Preço de Exercício", Placeholder: "Preço de strike"},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "xx/xx/xxxx"},
			{Key: "quantidade", Label: "Quantidade", Placeholder: "xxxx"},
			{Key: "preco", Label: "Preço", Placeholder: "Preço ou 'Ordem a mercado'"},
			{Key: "validade", Label: "Validade da ordem", Placeholder: "xx/xx/xxxx ou 'Até Cancelar'"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ativo Referência: %s;\nOpção: %s;\nTipo: %s;\nPreço de Exercício: %s;\nVencimento: %s;\nQuantidade: %s;\nPreço: %s;\nValidade da ordem: %s.",
				v["ativoRef"], v["opcao"], v["tipo"], v["strike"], v["vencimento"], v["quantidade"], v["preco"], v["validade"])
		},
		Fechamento: "Aguardo confirmação para realizar a venda.",
	},
	{
		ID: "termo-compra", Group: "Operações a Termo", Label: "Compra a Termo", Tipo: "outro",
		IntroFrase: "a compra a Termo",
		Fields: []Field{
			{Key: "ativo", Label: "Ativo", Placeholder: "Código do ativo"},
			{Key: "quantidade", Label: "Quantidade", Placeholder: "xxxx"},
			{Key: "preco", Label: "Preço", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "xx/xx/xxxx ou nº de dias"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ativo: %s;\nQuantidade: %s;\nPreço: %s;\nVencimento: %s.",
				v["ativo"], v["quantidade"], v["preco"], v["vencimento"])
		},
		Fechamento: "Aguardo confirmação para realizar a compra.",
	},
	{
		ID: "termo-venda", Group: "Operações a Termo", Label: "Venda a Termo", Tipo: "outro",
		IntroFrase: "a venda a Termo",
		Fields: []Field{
			{Key: "ativo", Label: "Ativo", Placeholder: "Código do ativo"},
			{Key: "quantidade", Label: "Quantidade", Placeholder: "xxxx"},
			{Key: "preco", Label: "Preço", Placeholder: "Ex: 1.000,00", Moeda: true},
			{Key: "vencimento", Label: "Vencimento", Placeholder: "xx/xx/xxxx ou nº de dias"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Ativo: %s;\nQuantidade: %s;\nPreço: %s;\nVencimento: %s.",
				v["ativo"], v["quantidade"], v["preco"], v["vencimento"])
		},
		Fechamento: "Aguardo confirmação para realizar a venda.",
	},

	// ---------------- MOVIMENTAÇÃO ----------------
	{
		ID: "retirada-recursos", Group: "Movimentação de Recursos", Label: "Retirada de Recursos", Tipo: "outro",
		IntroFrase: "a retirada de recursos, com destino à conta abaixo, de acordo com seu cadastro,",
		Fields: []Field{
			{Key: "banco", Label: "Banco", Placeholder: "Nome ou número do banco"},
			{Key: "agencia", Label: "Agência", Placeholder: "Número da agência"},
			{Key: "conta", Label: "Conta", Placeholder: "Número da conta"},
			{Key: "valor", Label: "Valor a ser transferido", Placeholder: "Ex: 1.000,00", Moeda: true},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Banco: %s;\nAgência: %s;\nConta: %s;\nValor a ser transferido: %s;\nTipo de transferência: TED.",
				v["banco"], v["agencia"], v["conta"], v["valor"])
		},
		Fechamento: "Aguardo confirmação para realizar a transferência.",
	},

	// ---------------- COMPROMISSADAS ----------------
	{
		ID: "compromissada-aplicacao", Group: "Compromissadas", Label: "Aplicação", Tipo: "aplicacao",
		IntroFrase: "a aplicação em operação Compromissada (autorização do saldo disponível às 17h, pelo prazo de 180 dias)",
		Fields: []Field{
			{Key: "taxaMinima", Label: "Taxa mínima de recompra", Placeholder: "Ex: XX% CDI"},
			{Key: "valorLimite", Label: "Valor limite a ser aplicado", Placeholder: "Ex: 1.000,00", Moeda: true},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Operação: Compromissada XP Investimentos;\nTipo: Compra com compromisso de revenda;\nAtivo objeto: Lastro disponível no momento da aplicação, em atenção ao determinado pelo CMN (Conselho Monetário Nacional);\nTaxa mínima de recompra: %s;\nLiquidez: Diária;\nVencimento: de 1 a 22 dias úteis;\nValor limite a ser aplicado: %s.",
				v["taxaMinima"], v["valorLimite"])
		},
		Fechamento: "Aguardo confirmação para realizar a aplicação.",
	},
	{
		ID: "compromissada-resgate", Group: "Compromissadas", Label: "Resgate", Tipo: "resgate",
		IntroFrase: "o resgate",
		Fields: []Field{
			{Key: "dataRecompra", Label: "Data de Recompra", Placeholder: "xx/xx/xxxx"},
			{Key: "valor", Label: "Valor a ser resgatado", Placeholder: "Ex: 1.000,00", Moeda: true},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Operação: Compromissada XP Investimentos;\nTipo: Revenda;\nData de Recompra: %s;\nValor a ser resgatado: %s.",
				v["dataRecompra"], v["valor"])
		},
		Fechamento: "Aguardo confirmação para realizar o resgate.",
	},

	// ---------------- CARTEIRA AUTOMATIZADA ----------------
	{
		ID: "carteira-nova", Group: "Carteira Automatizada", Label: "Nova Carteira", Tipo: "outro",
		IntroFrase: "a entrada na Carteira Automatizada",
		Fields: []Field{
			{Key: "casaAnalise", Label: "Casa de análise homologada", Placeholder: "Empresa que recomendou a carteira"},
			{Key: "dataExecucao", Label: "Data da execução", Placeholder: "xx/xx/xxxx"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Casa de análise: %s;\nData da execução: %s (na autorização recebida após o fechamento do mercado será processada no próximo dia útil, e somente se possível atender nas condições autorizadas);\nPreço: \"à mercado\" - as execuções ocorrerão ao longo do dia (preço final será encaminhado em novo e-mail ao final do pregão);\nQuantidades: execuções ao longo do dia (quantidades serão conhecidas a partir dos preços, e encaminhadas ao final do dia).\nAbaixo a Carteira de ações com o indicativo dos percentuais recomendados pela casa de análise [incluir tabela com a carteira sugerida].",
				v["casaAnalise"], v["dataExecucao"])
		},
		Anexo:      "Atenção assessor: é obrigatório o envio do PDF da carteira recomendada, tanto para entrada como para o rebalanceamento.",
		Fechamento: "Aguardo confirmação para fazer as compras indicadas acima.",
	},
	{
		ID: "carteira-rebalanceamento", Group: "Carteira Automatizada", Label: "Troca e Rebalanceamento", Tipo: "outro",
		IntroFrase: "o rebalanceamento da Carteira Automatizada",
		Fields: []Field{
			{Key: "casaAnalise", Label: "Casa de análise homologada", Placeholder: "Empresa que recomendou a carteira"},
			{Key: "dataExecucao", Label: "Data da execução", Placeholder: "xx/xx/xxxx"},
		},
		Body: func(v map[string]string) string {
			return fmt.Sprintf("Casa de análise: %s;\nData da execução: %s (a autorização recebida após o fechamento do mercado será processada no próximo dia útil, e somente se possível atender nas condições autorizadas);\nPreço: \"à mercado\" - as execuções ocorrerão ao longo do dia (preço final será encaminhado em novo e-mail ao final do pregão);\nQuantidades: execuções ao longo do dia (quantidades serão conhecidas a partir dos preços, e encaminhadas ao final do dia).\nAbaixo a Carteira atual versus a carteira sugerida [incluir tabela indicando a carteira atual e a sugerida].",
				v["casaAnalise"], v["dataExecucao"])
		},
		Anexo:      "Atenção assessor: é obrigatório o envio do PDF da carteira recomendada, tanto para entrada como para o rebalanceamento.",
		Fechamento: "Aguardo confirmação para fazer as compras indicadas acima.",
	},

	// ---------------- RESGATE DE PREVIDÊNCIA ----------------
	// Modelos de texto completo (saudação a fechamento próprios), aceitos
	// pela XP apenas nesse formato. Só existem no modo padronizado e não
	// combinam com outras operações no mesmo e-mail.
	{
		ID: "prev-pgbl", Group: "Resgate Prev", Label: "PGBL", Tipo: "resgate",
		SoPadronizado: true,
		Fields: []Field{
			{Key: "aliquota", Label: "Alíquota", Placeholder: "Ex: 15%"},
			{Key: "data", Label: "Data do resgate", Placeholder: "xx/xx/xx"},
		},
		EmailCompleto: func(codigo, nome string, v map[string]string) string {
			return fmt.Sprintf("\"Prezado(a) %s, tudo bem?\n\n"+
				"Para seguirmos com sua alocação gostaria de pedir uma validação formal para garantir que estamos seguindo nossa Política de Investimentos, portanto, peço o seu \"de acordo\" nesse e-mail.\n\n"+
				"O resgate de um PGBL na alíquota %s incide uma carga tributária muito onerosa, especialmente por conta da tributação ser sobre o valor total do patrimônio, e não no rendimento.\n\n"+
				"Por esses fatos, recomendamos que seja a última fonte de liquidez em uma eventual necessidade de capital.\n\n"+
				"Queremos garantir que todas as possibilidades de liquidez foram estudadas, dentro e fora da XP.\n\n"+
				"Posto esses fatos, gostaríamos de reafirmar sua ciência em relação ao resgate de previdência realizado em %s.\n\n"+
				"Tenho o seu \"de acordo\"?\"",
				nome, v["aliquota"], v["data"])
		},
	},
	{
		ID: "prev-vgbl", Group: "Resgate Prev", Label: "VGBL", Tipo: "resgate",
		SoPadronizado: true,
		Fields: []Field{
			{Key: "aliquota", Label: "Alíquota", Placeholder: "Ex: 15%"},
			{Key: "data", Label: "Data do resgate", Placeholder: "xx/xx/xx"},
		},
		EmailCompleto: func(codigo, nome string, v map[string]string) string {
			return fmt.Sprintf("\"Prezado(a) %s, tudo bem?\n\n"+
				"Para seguirmos com sua alocação gostaria de pedir uma validação formal para garantir que estamos seguindo nossa Política de Investimentos, portanto, peço o seu \"de acordo\" nesse e-mail.\n\n"+
				"O resgate de um VGBL na alíquota %s incide uma carga tributária muito onerosa.\n\n"+
				"Por esses fatos, recomendamos que seja a última fonte de liquidez em uma eventual necessidade de capital.\n\n"+
				"Queremos garantir que todas as possibilidades de liquidez foram estudadas, dentro e fora da XP.\n\n"+
				"Posto esses fatos, gostaríamos de reafirmar sua ciência em relação ao resgate de previdência realizado em %s.\n\n"+
				"Tenho o seu \"de acordo\"?\"",
				nome, v["aliquota"], v["data"])
		},
	},

	// ---------------- ERRO OPERACIONAL ----------------
	// Reporte de erro operacional pra aprovação prévia + operacionalização
	// do time XP — texto e saudação próprios, sem o "tipo de movimentação"
	// dos outros grupos (um só Label, igual Movimentação de Recursos) e
	// restrito a uma operação por e-mail (garantido pelo próprio mecanismo
	// EmailCompleto, ver MontarTextoEmailPadronizado).
	{
		ID: "erro-operacional", Group: "Erro Operacional", Label: "Erro Operacional", Tipo: "outro",
		SoPadronizado: true,
		Fields: []Field{
			{Key: "codigoErro", Label: "Código do erro", Placeholder: "Código do erro"},
			{Key: "nomeAssessor", Label: "Nome do assessor", Placeholder: "Nome do assessor"},
			{Key: "codigoResponsavel", Label: "Código do assessor responsável", Placeholder: "Código do assessor"},
			{Key: "nomeResponsavel", Label: "Nome do assessor responsável", Placeholder: "Nome do assessor"},
			{Key: "dataIncidente", Label: "Data do incidente", Placeholder: "xx/xx/xxxx"},
			{Key: "dataIdentificacao", Label: "Data de identificação", Placeholder: "xx/xx/xxxx"},
			{Key: "descricao", Label: "Descrição", Placeholder: "Descrição do erro"},
			{Key: "valor", Label: "Valor", Placeholder: "Ex: 1.000,00", Moeda: true},
		},
		EmailCompleto: func(codigo, nome string, v map[string]string) string {
			return fmt.Sprintf("\"Prezados(as), bom dia!\n\n"+
				"Segue abaixo a descrição do erro operacional de código %s para aprovação prévia e, posteriormente, para a operacionalização do time XP.\n\n"+
				"Cliente: %s - %s\n"+
				"Assessor: %s\n"+
				"Responsável pelo erro: %s - %s\n"+
				"Data do incidente: %s\n"+
				"Identificação do incidente: %s\n\n"+
				"Descrição: %s\n\n"+
				"Valor: %s\"",
				v["codigoErro"], codigo, nome, v["nomeAssessor"], v["codigoResponsavel"], v["nomeResponsavel"],
				v["dataIncidente"], v["dataIdentificacao"], v["descricao"], v["valor"])
		},
	},
}
