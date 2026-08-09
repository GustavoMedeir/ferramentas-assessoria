package icpbrasil

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// URLs oficiais do repositório do ITI. Variáveis (não const) pra poderem
// ser trocadas por um httptest.Server nos testes.
var (
	urlRaizFmt = "https://acraiz.icpbrasil.gov.br/credenciadas/RAIZ/ICP-Brasilv%s.crt"
	urlBundle  = "https://acraiz.icpbrasil.gov.br/credenciadas/CertificadosAC-ICP-Brasil/ACcompactado.zip"
)

// geracoesRaiz são as gerações da AC-Raiz ainda válidas hoje (a v1/v2/v3
// já expiraram e não constam mais no repositório do ITI). Certificados
// assinados sob uma geração antiga continuam existindo até expirarem, por
// isso a cadeia precisa ter todas as gerações ativas, não só a mais nova.
var geracoesRaiz = []string{"4", "5", "6", "7", "10", "11", "12", "13"}

// limiteTamanhoResposta é um teto de sanidade por download — nenhum dos
// arquivos reais passa de poucos MB; evita um servidor comprometido/
// redirecionado prender memória indefinidamente.
const limiteTamanhoResposta = 8 << 20 // 8MB

// baixarCadeiaCompleta baixa todos os certificados de AC-Raiz + o bundle de
// ACs subordinadas e devolve tudo concatenado como um único bundle PEM.
func baixarCadeiaCompleta(ctx context.Context) ([]byte, int, error) {
	var buf bytes.Buffer

	for _, v := range geracoesRaiz {
		url := fmt.Sprintf(urlRaizFmt, v)
		dados, err := baixar(ctx, url)
		if err != nil {
			return nil, 0, fmt.Errorf("baixar AC-Raiz v%s: %w", v, err)
		}
		escreverComQuebraDeLinha(&buf, dados)
	}

	zipDados, err := baixar(ctx, urlBundle)
	if err != nil {
		return nil, 0, fmt.Errorf("baixar bundle de ACs subordinadas: %w", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipDados), int64(len(zipDados)))
	if err != nil {
		return nil, 0, fmt.Errorf("bundle de ACs subordinadas não é um zip válido: %w", err)
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		dados, err := lerEntradaZip(f)
		if err != nil {
			continue // uma entrada corrompida não derruba o download inteiro
		}
		escreverComQuebraDeLinha(&buf, dados)
	}

	n := bytes.Count(buf.Bytes(), []byte("BEGIN CERTIFICATE"))
	if n == 0 {
		return nil, 0, fmt.Errorf("nenhum certificado encontrado nos arquivos baixados")
	}
	return buf.Bytes(), n, nil
}

func lerEntradaZip(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(io.LimitReader(rc, limiteTamanhoResposta))
}

func escreverComQuebraDeLinha(buf *bytes.Buffer, dados []byte) {
	buf.Write(dados)
	if len(dados) > 0 && dados[len(dados)-1] != '\n' {
		buf.WriteByte('\n')
	}
}

func baixar(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limiteTamanhoResposta))
}
