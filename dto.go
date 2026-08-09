package main

import (
	"rentabilidade/internal/emailgen"
	"rentabilidade/internal/icpbrasil"
	"rentabilidade/internal/pdfreport"
	"rentabilidade/internal/rentabilidade"
	"rentabilidade/internal/sigvalidator"
)

// RegistroDTO é a versão "wire" (JSON-safe) de rentabilidade.Registro,
// exposta ao frontend. Traz os 6 campos de rentabilidade já formatados em
// pt-BR (fonte única de verdade: pdfreport.FormatarReais/FormatarPercentual)
// pra não duplicar essa lógica em JS.
type RegistroDTO struct {
	Arquivo        string
	Codigo         string
	DataReferencia string
	GanhoMesReais  float64
	GanhoAnoReais  float64
	RentMesPct     float64
	RentAnoPct     float64
	CDIMesPct      float64
	CDIAnoPct      float64
	Ganho12MReais  float64
	Rent12MPct     float64
	CDI12MPct      float64
	Patrimonio     float64
	Copiado        bool
	RentFmt        string
	RentAFmt       string
	Rent12MFmt     string
	PercFmt        string
	PercAFmt       string
	Perc12MFmt     string
	CDIFmt         string
	CDIAFmt        string
	CDI12MFmt      string
}

// AtualizacaoDTO é o que o frontend recebe de VerificarAtualizacao e do
// evento "atualizacao:disponivel" — ver seção "Atualização automática" em
// app.go. Erro preenchido = falha de rede/API, não "sem atualização":
// Disponivel=false sem Erro é o caso normal de "já está na última versão".
type AtualizacaoDTO struct {
	Disponivel bool
	Versao     string
	Notas      string
	Erro       string
}

func paraRegistroDTO(r rentabilidade.Registro) RegistroDTO {
	return RegistroDTO{
		Arquivo:        r.Arquivo,
		Codigo:         r.Codigo,
		DataReferencia: r.DataReferencia,
		GanhoMesReais:  r.GanhoMesReais,
		GanhoAnoReais:  r.GanhoAnoReais,
		RentMesPct:     r.RentMesPct,
		RentAnoPct:     r.RentAnoPct,
		CDIMesPct:      r.CDIMesPct,
		CDIAnoPct:      r.CDIAnoPct,
		Ganho12MReais:  r.Ganho12MReais,
		Rent12MPct:     r.Rent12MPct,
		CDI12MPct:      r.CDI12MPct,
		Patrimonio:     r.Patrimonio,
		Copiado:        r.Copiado,
		RentFmt:        pdfreport.FormatarReais(r.GanhoMesReais),
		RentAFmt:       pdfreport.FormatarReais(r.GanhoAnoReais),
		Rent12MFmt:     pdfreport.FormatarReais(r.Ganho12MReais),
		PercFmt:        pdfreport.FormatarPercentual(r.RentMesPct),
		PercAFmt:       pdfreport.FormatarPercentual(r.RentAnoPct),
		Perc12MFmt:     pdfreport.FormatarPercentual(r.Rent12MPct),
		CDIFmt:         pdfreport.FormatarPercentual(r.CDIMesPct),
		CDIAFmt:        pdfreport.FormatarPercentual(r.CDIAnoPct),
		CDI12MFmt:      pdfreport.FormatarPercentual(r.CDI12MPct),
	}
}

// ClienteRentabilidadeDTO é uma linha da lista de clientes da aba
// Rentabilidade: todo cliente da base aparece, mesmo sem relatório
// processado (Registro fica nil nesse caso, ou quando o arquivo deu erro
// — a linha chega "em branco" no frontend).
type ClienteRentabilidadeDTO struct {
	Codigo        string
	Nome          string
	Registro      *RegistroDTO
	FestasEnviado bool // Modo Festas: já recebeu a mensagem de festas nesta leva
}

func paraClienteDTO(c rentabilidade.ClienteRentabilidade) ClienteRentabilidadeDTO {
	out := ClienteRentabilidadeDTO{Codigo: c.Codigo, Nome: c.Nome, FestasEnviado: c.FestasEnviado}
	if c.Registro != nil {
		dto := paraRegistroDTO(*c.Registro)
		out.Registro = &dto
	}
	return out
}

func paraClientesDTO(cs []rentabilidade.ClienteRentabilidade) []ClienteRentabilidadeDTO {
	out := make([]ClienteRentabilidadeDTO, len(cs))
	for i, c := range cs {
		out[i] = paraClienteDTO(c)
	}
	return out
}

// InicioDTO é devolvido por App.EstadoInicial, no boot do frontend.
type InicioDTO struct {
	TemPasta     bool
	Pasta        string
	Modelo       string
	ModeloFestas string
	ClientDB     map[string]string
	ClientEmails map[string]string
	Prefs        PreferenciasDTO
}

// PreferenciasDTO carrega as preferências persistidas em config.json.
// Valores vazios significam "use o padrão" (o frontend resolve).
type PreferenciasDTO struct {
	Tema                 string
	Acento               string
	ModoEmail            string
	TabelaPrevidenciaria string
	Visao                string
	Fonte                string
	ModoApresentacao     bool
	ModoFestas           bool
	EmailRemetente       string
	AssessorNome         string
	AssessorEmail        string
	OrdemNav             []string
	OrdemNavOcultos      []string
	// TemApresentacao indica se há um arquivo HTML de apresentação salvo em
	// config.json — usado no boot pra o Modo apresentação abrir direto na aba
	// Apresentação (em vez de Compromissada) quando há algo pra mostrar.
	TemApresentacao      bool
	RecortePersonalizado bool
	RecorteX0            float64
	RecorteY0            float64
	RecorteX1            float64
	RecorteY1            float64
	// Recorte padrão (pdfreport.RecorteGraficoRentabilidadePadrao) — exposto
	// pra tela de configuração pré-preencher o retângulo mesmo quando o
	// usuário nunca personalizou nada, sem duplicar a constante em JS.
	RecortePadraoX0 float64
	RecortePadraoY0 float64
	RecortePadraoX1 float64
	RecortePadraoY1 float64
}

// PastaDTO é devolvido ao escolher/trocar de pasta.
type PastaDTO struct {
	Pasta        string
	Modelo       string
	ModeloFestas string
}

// ApresentacaoDTO carrega a apresentação institucional da aba Apresentação:
// o caminho do arquivo escolhido (vazio = nenhum ainda) e o conteúdo lido do
// disco na hora. Tipo diz ao frontend qual campo usar ("html" -> HTML,
// "pdf" -> PDFBase64) — arquivos .pdf são exibidos no visualizador nativo do
// WebView2 via data URI, sem nenhuma conversão. Erro fica em Erro pra o
// frontend distinguir "nenhum arquivo" de "arquivo sumiu/ilegível" ou
// "formato não suportado" sem tratar isso como falha da chamada inteira.
type ApresentacaoDTO struct {
	Caminho   string
	Tipo      string // "html" | "pdf"
	HTML      string
	PDFBase64 string
	Erro      string
}

// FalhaDTO descreve um PDF que não pôde ser processado.
type FalhaDTO struct {
	Arquivo string
	Erro    string
}

// ProcessamentoDTO é devolvido por toda operação que muda a lista de
// clientes — o frontend nunca precisa fazer uma chamada extra de
// listagem.
type ProcessamentoDTO struct {
	Sucesso  int
	Falhas   []FalhaDTO
	Clientes []ClienteRentabilidadeDTO
}

// BaseClientesDTO é devolvido por App.CarregarBaseClientes: a base de
// clientes recém-carregada (nome por código) e a lista de clientes da aba
// Rentabilidade já reconstruída com ela — o join cliente↔registro muda
// quando a base muda, então precisa ser recalculado no backend (evita
// duplicar essa lógica em JS).
type BaseClientesDTO struct {
	ClientDB     map[string]string
	ClientEmails map[string]string
	Clientes     []ClienteRentabilidadeDTO
}

func paraFalhasDTO(falhas []rentabilidade.Falha) []FalhaDTO {
	out := make([]FalhaDTO, len(falhas))
	for i, f := range falhas {
		out[i] = FalhaDTO{Arquivo: f.Arquivo, Erro: f.Erro}
	}
	return out
}

// CampoDTO é a versão wire de emailgen.Field.
type CampoDTO struct {
	Key         string
	Label       string
	Placeholder string
	Type        string
	Options     []string `json:",omitempty"`
}

// CategoriaDTO é a versão wire de emailgen.Category, sem o campo Body
// (função Go, não serializa em JSON).
type CategoriaDTO struct {
	Group         string
	Label         string
	IntroFrase    string
	Anexo         string
	SoPadronizado bool
	OperacaoUnica bool // true: e-mail admite só 1 operação (ex.: Resgate Prev)
	Fields        []CampoDTO
}

// CatalogoEmailDTO é o catálogo completo de categorias — dado estático,
// buscado uma vez pelo frontend e cacheado lá.
type CatalogoEmailDTO struct {
	Produtos         []string
	Categorias       []CategoriaDTO
	InfoEstruturadas string
}

func catalogoEmailDTO() CatalogoEmailDTO {
	categorias := make([]CategoriaDTO, len(emailgen.Categories))
	for i, c := range emailgen.Categories {
		fields := make([]CampoDTO, len(c.Fields))
		for j, f := range c.Fields {
			fields[j] = CampoDTO{
				Key:         f.Key,
				Label:       f.Label,
				Placeholder: f.Placeholder,
				Type:        f.Type,
				Options:     f.Options,
			}
		}
		categorias[i] = CategoriaDTO{
			Group:         c.Group,
			Label:         c.Label,
			IntroFrase:    c.IntroFrase,
			Anexo:         c.Anexo,
			SoPadronizado: c.SoPadronizado,
			OperacaoUnica: c.EmailCompleto != nil,
			Fields:        fields,
		}
	}
	return CatalogoEmailDTO{
		Produtos:         emailgen.Produtos,
		Categorias:       categorias,
		InfoEstruturadas: emailgen.InfoEstruturadas,
	}
}

// AssinaturaDTO é uma imagem de assinatura salva (aba Assinatura em
// Configurações), pronta pro frontend exibir/usar.
type AssinaturaDTO struct {
	Nome   string
	Base64 string
	Ativa  bool
}

// ItemEmailEntrada é uma operação enviada pelo frontend pra montar o
// e-mail: identifica a categoria por grupo+label e traz os valores brutos
// dos campos (ainda sem o fallback "[Label]" — isso é resolvido no backend).
type ItemEmailEntrada struct {
	Group   string
	Label   string
	Valores map[string]string
}

// VerificacaoDTO é um item da checklist mostrada no card de resultado da
// validação de assinatura (ex.: "Integridade e assinatura", "Cadeia de
// confiança ICP-Brasil", "Revogação", "Carimbo de tempo").
type VerificacaoDTO struct {
	Nome    string
	Passou  bool
	Detalhe string
}

// ResultadoValidacaoDTO é o resultado de uma validação de assinatura
// ICP-Brasil. Estado nunca é um booleano solto — cada motivo de "não deu
// pra confirmar válida" é um valor próprio, porque a mensagem certa pro
// assessor muda bastante entre eles (ver sigvalidator.Estado).
//
// Estado: "Valida" | "Invalida" | "ACNaoReconhecida" | "CertificadoExpirado"
// | "FormatoNaoSuportado" | "ErroRevogacao"
type ResultadoValidacaoDTO struct {
	Estado string
	Motivo string

	Formato string // "CAdESDestacado" | "CAdESAnexado" | "PAdES" | "Desconhecido"

	NomeSignatario string
	CPF            string
	CNPJ           string
	ACEmissora     string

	DataAssinatura  string // formatada pt-BR; vazio se ausente
	TemCarimboTempo bool

	Verificacoes         []VerificacaoDTO
	TempoProcessamentoMs int64
}

func paraResultadoValidacaoDTO(r sigvalidator.Resultado) ResultadoValidacaoDTO {
	dto := ResultadoValidacaoDTO{
		Estado:               r.Estado.String(),
		Motivo:               r.Motivo,
		Formato:              r.Formato.String(),
		NomeSignatario:       r.Signatario.Nome,
		CPF:                  r.Signatario.CPF,
		CNPJ:                 r.Signatario.CNPJ,
		ACEmissora:           r.ACEmissora,
		TemCarimboTempo:      r.TemCarimboTempo,
		TempoProcessamentoMs: r.TempoProcessamento.Milliseconds(),
	}
	if r.TemDataAssinatura {
		dto.DataAssinatura = r.DataAssinatura.Local().Format("02/01/2006 15:04")
	}
	dto.Verificacoes = make([]VerificacaoDTO, len(r.Verificacoes))
	for i, v := range r.Verificacoes {
		dto.Verificacoes[i] = VerificacaoDTO{Nome: v.Nome, Passou: v.Passou, Detalhe: v.Detalhe}
	}
	return dto
}

// InfoCadeiaDTO descreve a cadeia de certificados ICP-Brasil em uso
// localmente — mostrada no rodapé da aba de validação de assinatura.
type InfoCadeiaDTO struct {
	AtualizadoEm    string // formatada pt-BR; vazio se desconhecida
	Origem          string // "embutida" | "baixada"
	NumCertificados int
	Erro            string // preenchido só quando uma atualização manual falha — estado, não Go error
}

func paraInfoCadeiaDTO(i icpbrasil.Info) InfoCadeiaDTO {
	dto := InfoCadeiaDTO{Origem: i.Origem, NumCertificados: i.NumCertificados}
	if !i.AtualizadoEm.IsZero() {
		dto.AtualizadoEm = i.AtualizadoEm.Local().Format("02/01/2006 15:04")
	}
	return dto
}

// ItemRespostaTypeformDTO é uma pergunta (mesmo texto exibido no Typeform) e
// a resposta já formatada, exatamente como o frontend monta pro .txt salvo
// (ver respostaFormatada em typeform.js) — usado por PreencherTypeform.
type ItemRespostaTypeformDTO struct {
	Pergunta string
	Valor    string
}
