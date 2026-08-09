package sigvalidator

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DetectarFormato decide, a partir dos caminhos escolhidos no diálogo de
// arquivo, se é um caso de CAdES destacado (par .p7s/.p7m + conteúdo),
// CAdES anexado (só o .p7s/.p7m, conteúdo embutido), ou PAdES (um PDF só).
// Não abre nenhum arquivo — é decisão só por extensão/contagem, pura e sem
// I/O, pra ser fácil de testar sem depender de nenhuma assinatura real.
//
// A confirmação definitiva de "anexado tem mesmo o conteúdo embutido" fica
// por conta de quem chama (validarCAdES olha se p7.Content veio populado
// depois do parse de verdade).
func DetectarFormato(caminhos []string) (formato Formato, caminhoP7S, caminhoConteudo string, err error) {
	switch len(caminhos) {
	case 0:
		return FormatoDesconhecido, "", "", fmt.Errorf("nenhum arquivo selecionado")

	case 1:
		c := caminhos[0]
		switch {
		case ehAssinatura(c):
			return FormatoCAdESAnexado, c, "", nil
		case ehPDF(c):
			return FormatoPAdES, c, "", nil
		default:
			return FormatoDesconhecido, "", "", nil
		}

	case 2:
		a, b := caminhos[0], caminhos[1]
		switch {
		case ehAssinatura(a) && !ehAssinatura(b):
			return FormatoCAdESDestacado, a, b, nil
		case ehAssinatura(b) && !ehAssinatura(a):
			return FormatoCAdESDestacado, b, a, nil
		default:
			// dois arquivos de assinatura, ou nenhum dos dois — não dá pra
			// saber qual é o par certo.
			return FormatoDesconhecido, "", "", nil
		}

	default:
		return FormatoDesconhecido, "", "", nil
	}
}

func ehAssinatura(caminho string) bool {
	ext := strings.ToLower(filepath.Ext(caminho))
	return ext == ".p7s" || ext == ".p7m"
}

func ehPDF(caminho string) bool {
	return strings.ToLower(filepath.Ext(caminho)) == ".pdf"
}
