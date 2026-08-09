package sigvalidator

import (
	"context"
	"os"
	"testing"

	"rentabilidade/internal/icpbrasil"
)

// pdfAssinadoRealParaTeste é um PDF assinado de verdade, fornecido durante
// o desenvolvimento desta feature — não faz parte do repositório (caminho
// fora da árvore do projeto), então o teste pula sozinho se não existir
// (ex.: rodando em outra máquina/CI). Serve pra validar o pipeline
// ByteRange/Contents/pkcs7.Parse contra um arquivo real, não só sintético.
const pdfAssinadoRealParaTeste = `C:\Users\gutom\.claude\uploads\ee4753bf-d6b3-4192-80ac-f11a55e656f6\efa93291-9.pdf`

func TestValidarPAdESArquivoReal(t *testing.T) {
	if _, err := os.Stat(pdfAssinadoRealParaTeste); err != nil {
		t.Skipf("arquivo de teste real não encontrado nesta máquina: %v", err)
	}

	store, err := icpbrasil.NovoStore(t.TempDir())
	if err != nil {
		t.Fatalf("NovoStore: %v", err)
	}

	r := validarPAdES(context.Background(), pdfAssinadoRealParaTeste, store)

	for _, v := range r.Verificacoes {
		t.Logf("Verificacao: %s passou=%v detalhe=%q", v.Nome, v.Passou, v.Detalhe)
	}

	// Extração de identidade (CPF via otherName da SAN) e a integridade
	// criptográfica são fatos verificáveis contra este arquivo real,
	// independente da cadeia de confiança disponível na máquina que roda o
	// teste — checados com precisão.
	if r.Signatario.CPF != "12442567733" {
		t.Errorf("CPF: esperado 12442567733, veio %q", r.Signatario.CPF)
	}
	if r.Signatario.Nome != "GUSTAVO DE MEDEIROS MEIRELES" {
		t.Errorf("Nome: esperado GUSTAVO DE MEDEIROS MEIRELES, veio %q", r.Signatario.Nome)
	}
	if r.ACEmissora != "AC Final do Governo Federal do Brasil v1" {
		t.Errorf("ACEmissora: esperado AC Final do Governo Federal do Brasil v1, veio %q", r.ACEmissora)
	}
	if !r.TemDataAssinatura {
		t.Error("esperava TemDataAssinatura true")
	}

	achouIntegridade := false
	for _, v := range r.Verificacoes {
		if v.Nome == "Integridade e assinatura" {
			achouIntegridade = true
			if !v.Passou {
				t.Errorf("integridade/assinatura deveria passar (arquivo real, não adulterado): %s", v.Detalhe)
			}
		}
	}
	if !achouIntegridade {
		t.Error("esperava um item de verificação \"Integridade e assinatura\" na lista")
	}

	// O emissor "AC Final do Governo Federal do Brasil v1" não consta nos
	// bundles públicos do repositório do ITI no momento em que este teste
	// foi escrito (confirmado contra o bundle "ativas" de 176 certs e o
	// "cadeia completa" de 336 — nenhum dos dois tem essa AC) — então o
	// resultado correto E ESPERADO com a cadeia local atual é
	// ACNaoReconhecida, não Valida. Isso não é falha do pipeline: é a
	// cadeia de confiança realmente não reconhecendo essa AC específica.
	// Se um bundle futuro do ITI passar a incluir essa AC, este teste vai
	// começar a falhar aqui — sinal de que a asserção precisa ser
	// atualizada pra EstadoValida.
	if r.Estado != EstadoACNaoReconhecida {
		t.Errorf("Estado: esperado ACNaoReconhecida (AC emissora fora dos bundles públicos do ITI conhecidos), veio %s — Motivo: %s", r.Estado, r.Motivo)
	}
}
