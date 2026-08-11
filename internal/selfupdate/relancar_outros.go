//go:build !windows

package selfupdate

import (
	"os/exec"
)

// relancarComEspera fora do Windows: o app só é distribuído pra Windows,
// mas manter o pacote compilando em qualquer plataforma facilita rodar
// `go test ./...` e as ferramentas de análise em outro sistema.
func relancarComEspera(caminhoNovo, caminhoAtual string) error {
	return exec.Command(caminhoNovo).Start()
}

// ExecutarSeAjudanteDeRelancamento fora do Windows: o mecanismo de
// auto-atualização é Windows-only (ver checarAtualizacao em app.go), então
// esta execução nunca é a instância ajudante — sempre false. Existe só pra
// main() poder chamar a função sem `//go:build` espalhado por lá.
func ExecutarSeAjudanteDeRelancamento(args []string) bool {
	return false
}
