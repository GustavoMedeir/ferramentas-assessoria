package sigvalidator

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/digitorus/pdf"
	"github.com/smallstep/pkcs7"

	"rentabilidade/internal/icpbrasil"
)

// validarPAdES extrai a assinatura embutida num PDF (o dicionário com
// /Filter /Adobe.PPKLite, /Contents e /ByteRange) e delega a verificação
// pra validarCMS — o mesmo núcleo usado pelo CAdES (cades.go). A diferença
// entre os dois formatos termina assim que o *pkcs7.PKCS7 está montado e
// com Content populado.
//
// github.com/digitorus/pdf é uma lib de baixa visibilidade lendo um
// formato binário complexo vindo de fora (arquivo do cliente) — por isso
// a função inteira fica atrás de um recover(), igual ao já usado em
// identidade.go pro parsing de otherName.
func validarPAdES(ctx context.Context, caminhoPDF string, trust *icpbrasil.Store) (resultado Resultado) {
	defer func() {
		if r := recover(); r != nil {
			resultado = Resultado{
				Estado: EstadoFormatoNaoSuportado,
				Motivo: fmt.Sprintf("Não foi possível interpretar este PDF (%v).", r),
			}
		}
	}()

	arquivo, err := os.Open(caminhoPDF)
	if err != nil {
		return Resultado{Estado: EstadoFormatoNaoSuportado, Motivo: "Não foi possível ler o arquivo PDF selecionado."}
	}
	defer arquivo.Close()

	info, err := arquivo.Stat()
	if err != nil {
		return Resultado{Estado: EstadoFormatoNaoSuportado, Motivo: "Não foi possível ler o arquivo PDF selecionado."}
	}

	rdr, err := pdf.NewReader(arquivo, info.Size())
	if err != nil {
		return Resultado{
			Estado: EstadoFormatoNaoSuportado,
			Motivo: "O arquivo selecionado não é um PDF reconhecível — pode estar corrompido ou protegido por senha.",
		}
	}

	assinaturas := localizarAssinaturasPAdES(rdr)
	switch len(assinaturas) {
	case 0:
		return Resultado{Estado: EstadoFormatoNaoSuportado, Motivo: "Não foi encontrada nenhuma assinatura digital neste PDF."}
	case 1:
		// segue abaixo
	default:
		return Resultado{
			Estado: EstadoInvalida,
			Motivo: "Este PDF tem mais de uma assinatura digital (múltiplos signatários) — não suportado nesta versão.",
		}
	}

	v := assinaturas[0]

	// O /Contents é BER com comprimento indefinido (a assinatura CMS não
	// "sabe" de antemão quantos bytes vai ocupar quando o placeholder é
	// reservado no PDF) — o ber2der interno do pkcs7.Parse já sabe parar
	// nos marcadores de fim-de-conteúdo corretos, então os zeros de
	// padding que sobram depois (o Adobe reserva mais espaço do que a
	// assinatura de fato usa) são ignorados sozinhos, sem precisar aparar
	// nada aqui.
	p7, err := pkcs7.Parse([]byte(v.Key("Contents").RawString()))
	if err != nil {
		return Resultado{Estado: EstadoFormatoNaoSuportado, Motivo: "A assinatura embutida neste PDF não pôde ser interpretada."}
	}

	if err := preencherConteudoPorByteRange(v, arquivo, p7); err != nil {
		return Resultado{Estado: EstadoFormatoNaoSuportado, Motivo: "Não foi possível processar o intervalo assinado (ByteRange) deste PDF."}
	}

	return validarCMS(ctx, p7, trust)
}

// localizarAssinaturasPAdES caminha pelas referências cruzadas do PDF
// procurando dicionários de assinatura (Filter Adobe.PPKLite é o valor
// padrão usado por toda ferramenta de assinatura PAdES/Adobe).
func localizarAssinaturasPAdES(rdr *pdf.Reader) []pdf.Value {
	var achadas []pdf.Value
	for _, x := range rdr.Xref() {
		v := rdr.Resolve(x.Ptr(), x.Ptr())
		if v.Key("Filter").Name() == "Adobe.PPKLite" {
			achadas = append(achadas, v)
		}
	}
	return achadas
}

// preencherConteudoPorByteRange lê, do arquivo original, só as faixas de
// bytes indicadas em /ByteRange (o documento inteiro exceto o próprio
// valor de /Contents) e usa isso como o conteúdo assinado — é assim que o
// PAdES evita re-hashear o PDF inteiro incluindo o placeholder da
// assinatura.
func preencherConteudoPorByteRange(v pdf.Value, arquivo io.ReaderAt, p7 *pkcs7.PKCS7) error {
	faixas := v.Key("ByteRange")
	if faixas.Len() == 0 || faixas.Len()%2 != 0 {
		return fmt.Errorf("ByteRange ausente ou com formato inesperado")
	}
	for i := 0; i < faixas.Len(); i += 2 {
		inicio := faixas.Index(i).Int64()
		tamanho := faixas.Index(i + 1).Int64()
		conteudo, err := io.ReadAll(io.NewSectionReader(arquivo, inicio, tamanho))
		if err != nil {
			return fmt.Errorf("faixa de bytes %d-%d: %w", inicio, tamanho, err)
		}
		p7.Content = append(p7.Content, conteudo...)
	}
	return nil
}
