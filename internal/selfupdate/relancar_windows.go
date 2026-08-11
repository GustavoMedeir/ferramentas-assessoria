//go:build windows

package selfupdate

import (
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"rentabilidade/internal/shortcut"
)

// segundosEspera é quanto o relançador aguarda antes de agir. Precisa
// cobrir o encerramento da instância atual (que só começa depois que
// Relancar retorna) e a liberação do SingleInstanceLock — 4s é folgado pra
// um app que fecha em menos de 1s. É também, de propósito, tempo de sobra
// pra um antivírus com varredura de reputação assíncrona já ter agido
// sobre o executável recém-baixado (ver garantirExecutavelValido) antes da
// gente confiar nele.
const segundosEspera = 4 * time.Second

// creationFlags: CREATE_NO_WINDOW. Sem efeito prático aqui — o próprio
// executável já é GUI subsystem e nunca abre console — mas inofensivo
// manter como reforço.
const creationFlags = 0x08000000

// FlagRelancador é o argumento reconhecido em main() pra saber que esta
// execução não é o app normal, e sim a instância ajudante criada por
// relancarComEspera — ver ExecutarSeAjudanteDeRelancamento.
const FlagRelancador = "--pos-atualizacao"

// tamanhoMinimoValido é um piso de sanidade pro executável (a build atual
// fica na casa de 30MB) — só pra distinguir "arquivo de verdade" de um
// arquivo vazio/truncado (ex.: antivírus esvaziou o arquivo em vez de
// apagar), sem precisar saber o tamanho exato esperado.
const tamanhoMinimoValido = 5 << 20 // 5MB

// relancarComEspera sobe o processo ajudante — deliberadamente uma SEGUNDA
// CÓPIA do executável ATUAL (não o novo, recém-baixado e ainda não
// "provado" — ver comentário de ExecutarSeAjudanteDeRelancamento) — com
// FlagRelancador e os dois caminhos envolvidos. main() reconhece a flag e
// desvia pra esperar + decidir + relançar sem nunca inicializar o
// Wails/janela.
func relancarComEspera(caminhoNovo, caminhoAtual string) error {
	cmd := exec.Command(caminhoAtual, FlagRelancador, caminhoNovo, caminhoAtual)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: creationFlags,
	}
	return cmd.Start()
}

// ExecutarSeAjudanteDeRelancamento verifica se esta execução do binário é a
// instância ajudante criada por relancarComEspera (reconhecida por
// FlagRelancador em args). Se for, devolve true — o chamador (main()) deve
// encerrar imediatamente, sem inicializar o Wails — depois de:
//  1. Esperar segundosEspera.
//  2. Conferir se o executável novo (baixado em Downloads por
//     BaixarEAplicar) sobreviveu — ver garantirExecutavelValido. Esta
//     instância ajudante roda a partir do executável ANTIGO de propósito:
//     ele já está rodando com sucesso agora mesmo (é literalmente o
//     processo que acabou de servir o app), enquanto o novo ainda não
//     teve chance de ser "aprovado" por nenhum antivírus com varredura
//     assíncrona — subir o ajudante a partir dele arriscaria o próprio
//     relançador ser barrado.
//  3. Se sobreviveu: reaponta os atalhos da área de trabalho pro novo
//     executável (ver internal/shortcut) e abre ele.
//  4. Se não: loga o ocorrido e abre o executável ANTIGO em vez disso — ele
//     nunca foi tocado, então continua garantidamente válido. O atalho não
//     é mexido nesse caso (já aponta pro antigo, que é o que vai abrir).
func ExecutarSeAjudanteDeRelancamento(args []string) bool {
	if len(args) < 4 || args[1] != FlagRelancador {
		return false
	}
	caminhoNovo := args[2]
	caminhoAtual := args[3]
	time.Sleep(segundosEspera)

	caminhoParaAbrir := caminhoAtual
	if executavelValido(caminhoNovo) {
		caminhoParaAbrir = caminhoNovo
		if n, err := shortcut.RetargetarNoDesktop(caminhoAtual, caminhoNovo); err != nil {
			log.Println("selfupdate: não foi possível atualizar o atalho da área de trabalho:", err)
		} else if n > 0 {
			log.Printf("selfupdate: %d atalho(s) da área de trabalho reapontados pra %s", n, caminhoNovo)
		}
	} else {
		log.Println("selfupdate: executável baixado", caminhoNovo, "sumiu ou ficou corrompido antes de reabrir (provável interferência de antivírus) — voltando pra versão anterior:", caminhoAtual)
	}

	if err := exec.Command(caminhoParaAbrir).Start(); err != nil {
		log.Println("selfupdate: relançador não conseguiu abrir", caminhoParaAbrir, "-", err)
	}
	return true
}

// executavelValido confere se caminho existe e tem um tamanho plausível —
// ver comentário de tamanhoMinimoValido e de BaixarEAplicar (o cenário que
// isso cobre: antivírus com varredura de reputação assíncrona apagando o
// executável novo alguns segundos depois do download já ter terminado com
// sucesso).
func executavelValido(caminho string) bool {
	info, err := os.Stat(caminho)
	return err == nil && info.Size() > tamanhoMinimoValido
}
