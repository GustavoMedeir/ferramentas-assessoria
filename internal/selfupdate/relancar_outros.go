//go:build !windows

package selfupdate

import (
	"os"
	"os/exec"
)

// relancarComEspera fora do Windows: o app só é distribuído pra Windows,
// mas manter o pacote compilando em qualquer plataforma facilita rodar
// `go test ./...` e as ferramentas de análise em outro sistema.
func relancarComEspera(caminho string) error {
	return exec.Command(caminho, os.Args[1:]...).Start()
}
