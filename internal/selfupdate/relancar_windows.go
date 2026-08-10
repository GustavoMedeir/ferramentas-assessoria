//go:build windows

package selfupdate

import (
	"os/exec"
	"strconv"
	"syscall"
)

// segundosEspera é quanto o relançador aguarda antes de reabrir o app.
// Precisa cobrir o encerramento da instância atual (que só começa depois
// que Relancar retorna) e a liberação do SingleInstanceLock. 4s é folgado
// pra um app que fecha em menos de 1s, e o usuário mal percebe.
const segundosEspera = 4

// creationFlags: CREATE_NO_WINDOW — o cmd.exe intermediário não pode piscar
// uma janela de console preta na tela do usuário.
const creationFlags = 0x08000000

// relancarComEspera dispara um cmd.exe que espera e depois abre o app,
// desanexado deste processo (por isso `start`, e por isso ele sobrevive ao
// nosso encerramento).
//
// O "ping -n" é o jeito clássico de esperar em batch sem depender do
// `timeout`, que exige um console interativo — e aqui não há console
// nenhum (ver creationFlags).
//
// A linha vai em SysProcAttr.CmdLine, crua, em vez de exec.Command(...):
// o Go monta a linha de comando escapando aspas no formato que o
// CommandLineToArgvW espera, e o cmd.exe NÃO usa essa convenção. O
// resultado eram as aspas do `start "" "caminho"` chegarem quebradas e o
// relançamento simplesmente não acontecer (visto em teste).
func relancarComEspera(caminho string) error {
	linha := `cmd.exe /C ping -n ` + strconv.Itoa(segundosEspera+1) +
		` 127.0.0.1 >NUL & start "" "` + caminho + `"`

	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: creationFlags,
		CmdLine:       linha,
	}
	return cmd.Start()
}
