// Package sigvalidator valida assinaturas digitais ICP-Brasil localmente,
// sem depender de nenhuma API externa em tempo de validação — a assinatura
// CMS/PKCS#7 já carrega o certificado do signatário, então basta ter a
// cadeia de confiança ICP-Brasil (pacote irmão internal/icpbrasil) pra
// verificar tudo offline.
//
// Suporta CAdES (.p7s/.p7m, PKCS#7/CMS destacado ou anexado — cades.go) e
// PAdES (assinatura embutida em PDF — pades.go). Os dois formatos
// compartilham o mesmo núcleo de verificação (validarCMS, em cades.go) —
// só a extração do *pkcs7.PKCS7 muda entre eles.
package sigvalidator

import (
	"context"
	"time"

	"rentabilidade/internal/icpbrasil"
)

// Estado é o resultado de uma validação. Nunca é um booleano — cada motivo
// de "não deu pra confirmar válida" é um estado próprio, com sua própria
// mensagem, porque as implicações pro assessor são bem diferentes (ex:
// ACNaoReconhecida pode só significar cadeia local desatualizada, não é
// indício de fraude; ErroRevogacao é falha de checagem, não reprovação).
type Estado int

const (
	EstadoIndefinido Estado = iota
	EstadoValida
	EstadoInvalida
	EstadoACNaoReconhecida
	EstadoCertificadoExpirado
	EstadoFormatoNaoSuportado
	EstadoErroRevogacao
)

func (e Estado) String() string {
	switch e {
	case EstadoValida:
		return "Valida"
	case EstadoInvalida:
		return "Invalida"
	case EstadoACNaoReconhecida:
		return "ACNaoReconhecida"
	case EstadoCertificadoExpirado:
		return "CertificadoExpirado"
	case EstadoFormatoNaoSuportado:
		return "FormatoNaoSuportado"
	case EstadoErroRevogacao:
		return "ErroRevogacao"
	default:
		return "Indefinido"
	}
}

// Formato identifica o tipo de assinatura detectado nos arquivos escolhidos.
type Formato int

const (
	FormatoDesconhecido Formato = iota
	FormatoCAdESDestacado
	FormatoCAdESAnexado
	FormatoPAdES // reconhecido (é um PDF), mas ainda não suportado nesta fase
)

func (f Formato) String() string {
	switch f {
	case FormatoCAdESDestacado:
		return "CAdESDestacado"
	case FormatoCAdESAnexado:
		return "CAdESAnexado"
	case FormatoPAdES:
		return "PAdES"
	default:
		return "Desconhecido"
	}
}

// Signatario são os dados de identidade extraídos do certificado.
type Signatario struct {
	Nome       string
	CPF        string
	CNPJ       string
	TipoPessoa string // "fisica" | "juridica" | "desconhecido"
}

// Verificacao é um item individual da checklist mostrada na UI (ex:
// "Integridade", "Cadeia de confiança", "Revogação").
type Verificacao struct {
	Nome    string
	Passou  bool
	Detalhe string
}

// Resultado é o resultado completo de uma validação.
type Resultado struct {
	Estado Estado
	Motivo string // preenchido sempre que Estado != EstadoValida

	Formato Formato

	Signatario Signatario
	ACEmissora string

	DataAssinatura    time.Time
	TemDataAssinatura bool

	// TemCarimboTempo indica só presença/ausência de um atributo de
	// carimbo de tempo (CAdES-T) na assinatura — validar o carimbo de
	// verdade (cadeia da própria ACT, TSTInfo aninhado) fica fora de
	// escopo nesta fase, ver comentário em timestamp.go.
	TemCarimboTempo bool

	Verificacoes []Verificacao

	TempoProcessamento time.Duration
}

// Input são os arquivos escolhidos pelo usuário pra uma validação.
type Input struct {
	CaminhoAssinatura string // .p7s ou .p7m
	CaminhoConteudo   string // vazio quando a assinatura é anexada (conteúdo embutido no próprio arquivo)
}

// Validate nunca usa error pra reportar um estado de negócio — error é
// reservado pra falha dura de verdade (contexto cancelado, por exemplo).
// Todo o resto (arquivo ilegível, assinatura corrompida, PDF ainda não
// suportado, cadeia desconhecida, etc.) vira Resultado.Estado, seguindo a
// mesma convenção de ApresentacaoDTO.Erro no resto do app: falha esperada
// vira campo pra UI mostrar aviso amigável, não quebra a chamada.
func Validate(ctx context.Context, in Input, trust *icpbrasil.Store) (Resultado, error) {
	inicio := time.Now()

	caminhos := []string{in.CaminhoAssinatura}
	if in.CaminhoConteudo != "" {
		caminhos = append(caminhos, in.CaminhoConteudo)
	}
	formato, caminhoP7S, caminhoConteudo, err := DetectarFormato(caminhos)
	if err != nil {
		return Resultado{}, err
	}

	var r Resultado
	switch formato {
	case FormatoCAdESDestacado, FormatoCAdESAnexado:
		r = validarCAdES(ctx, caminhoP7S, caminhoConteudo, trust)
	case FormatoPAdES:
		// caminhoP7S carrega o caminho do PDF nesse caso (ver
		// DetectarFormato, caso de 1 arquivo só) — nome mantido genérico
		// porque é o mesmo campo usado pro caminho do .p7s/.p7m no CAdES.
		r = validarPAdES(ctx, caminhoP7S, trust)
	default:
		r = Resultado{
			Estado: EstadoFormatoNaoSuportado,
			Motivo: "Não foi possível reconhecer o formato dos arquivos selecionados. Selecione um arquivo .p7s/.p7m (e, se a assinatura for destacada, o documento original junto).",
		}
	}
	r.Formato = formato
	r.TempoProcessamento = time.Since(inicio)
	return r, nil
}
