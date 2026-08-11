//go:build windows

package selfupdate

import (
	"os/exec"
	"syscall"
	"time"
)

// segundosEspera é quanto o relançador aguarda antes de reabrir o app.
// Precisa cobrir o encerramento da instância atual (que só começa depois
// que Relancar retorna) e a liberação do SingleInstanceLock. 4s é folgado
// pra um app que fecha em menos de 1s, e o usuário mal percebe.
const segundosEspera = 4 * time.Second

// creationFlags: CREATE_NO_WINDOW. Sem efeito prático aqui — o próprio
// executável já é GUI subsystem e nunca abre console — mas inofensivo
// manter como reforço.
const creationFlags = 0x08000000

// FlagRelancador é o argumento reconhecido em main() pra saber que esta
// execução não é o app normal, e sim a instância ajudante criada por
// relancarComEspera — ver ExecutarSeAjudanteDeRelancamento.
const FlagRelancador = "--pos-atualizacao"

// relancarComEspera reabre o app subindo uma SEGUNDA CÓPIA do próprio
// executável recém-atualizado, com FlagRelancador — main() reconhece a flag
// (ExecutarSeAjudanteDeRelancamento) e desvia pra esperar + relançar sem
// nunca inicializar o Wails/janela.
//
// Isto substitui uma versão anterior que fazia isso via `cmd.exe /C ping -n
// 5 127.0.0.1 & start "" "caminho"`, oculto (CREATE_NO_WINDOW). Isoladamente
// cada peça é comum em apps sem instalador, mas juntas — executável não
// assinado, que baixa a própria atualização, se substitui no disco e se
// relança através de um cmd.exe escondido com uma técnica clássica de
// espera disfarçada (ping como sleep) — batem com o padrão que
// antivírus/EDR heurísticos usam pra reconhecer dropper/malware
// auto-atualizável. Houve relato em campo do antivírus apagando o
// executável (e, por tabela, quebrando o atalho que apontava pra ele) na
// sequência de uma atualização. Reexecutar a si mesmo não elimina o "se
// substitui no disco" (inerente a um updater sem instalador), mas tira o
// cmd.exe/ping da equação — um dos sinais mais específicos.
func relancarComEspera(caminho string) error {
	cmd := exec.Command(caminho, FlagRelancador, caminho)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: creationFlags,
	}
	return cmd.Start()
}

// ExecutarSeAjudanteDeRelancamento verifica se esta execução do binário é a
// instância ajudante criada por relancarComEspera (reconhecida por
// FlagRelancador em args). Se for: espera a instância antiga soltar o
// SingleInstanceLock, sobe o app de verdade e devolve true — o chamador
// (main()) deve encerrar imediatamente, sem inicializar o Wails.
func ExecutarSeAjudanteDeRelancamento(args []string) bool {
	if len(args) < 3 || args[1] != FlagRelancador {
		return false
	}
	caminho := args[2]
	time.Sleep(segundosEspera)
	exec.Command(caminho).Start()
	return true
}
