package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"rentabilidade/internal/selfupdate"
)

//go:embed all:frontend/dist
var assets embed.FS

// Version é a versão do app em runtime, usada por internal/selfupdate para
// decidir se há uma release mais nova no GitHub. Fica "dev" por padrão (ex.:
// rodando via `wails dev` ou um build local sem esse ldflag) — nesse caso a
// checagem de atualização é pulada (ver app.go:startup), pra não sugerir
// "atualizar" um binário que não corresponde a nenhuma release publicada.
// Em produção é preenchida no build com:
//
//	wails build -ldflags "-X main.Version=1.00.02"
//
// (o valor deve bater com productVersion em wails.json — é o que vira a tag
// da release no GitHub, ex. v1.00.02). Esquema de versão MAJOR.MINOR.PATCH
// — regra completa e exemplos em DOCUMENTACAO.md, seção "Atualização
// automática".
var Version = "dev"

func main() {
	// Esta execução pode ser a instância ajudante que a própria atualização
	// automática sobe pra relançar o app (ver
	// internal/selfupdate/relancar_windows.go) — nesse caso ela só espera e
	// reabre o executável de verdade, sem inicializar o Wails/janela nem
	// tocar no log.
	if selfupdate.ExecutarSeAjudanteDeRelancamento(os.Args) {
		return
	}

	redirecionarStdioParaArquivo()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "XP Assessor - Ferramentas de Assessoria",
		Width:     1280,
		Height:    820,
		MinWidth:  960,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0xee, G: 0xf3, B: 0xf1, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		// Garante uma única instância do app: abrir o .exe de novo só traz a
		// janela já aberta pra frente. Sem isso, duas instâncias disputam o
		// mesmo rentabilidades.db (SQLite) e a segunda pode ficar com dados
		// desatualizados ou dar "database is locked" — cenário real observado
		// em teste (um build antigo esquecido aberto ao lado do novo).
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "ferramentas-assessoria-xp-gutomeireles",
			OnSecondInstanceLaunch: app.aoAbrirSegundaInstancia,
		},
		// O WebView2 guarda o zoom por conta do usuário e o restaura na
		// abertura seguinte: um Ctrl+scroll acidental (ou o zoom herdado do
		// Edge da máquina) fazia o app abrir com a interface esticada ou
		// minúscula, sem o usuário entender o motivo. Fixar o ZoomFactor
		// aqui faz toda abertura começar no mesmo tamanho, em qualquer
		// máquina. IsZoomControlEnabled fica ligado de propósito — quem
		// quiser dar zoom pontualmente continua podendo; o valor só não
		// sobrevive ao fechamento.
		Windows: &windows.Options{
			ZoomFactor:           0.85,
			IsZoomControlEnabled: true,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Println("Error:", err.Error())
	}
}

// redirecionarStdioParaArquivo troca os.Stdout/os.Stderr do processo por um
// arquivo de log persistente.
//
// Isso é essencial (não só "boa prática"): este é um app GUI, e no Windows
// um processo GUI subsystem aberto sem console (duplo-clique, atalho) recebe
// handles de stdout/stderr inválidos. Bibliotecas de terceiros que acessam
// os.Stdout/os.Stderr diretamente por fora de qualquer parâmetro de
// configuração — como o motor de PDF (wazero/go-pdfium, ver
// internal/pdfreport) — quebram com "GetFileType ...: The handle is
// invalid" nesse cenário. Redirecionar sempre, incondicionalmente, logo no
// início do main() evita ter que detectar corretamente essa condição (nada
// garante que só o PDFium acesse os.Stdout — qualquer dependência futura
// pode fazer o mesmo), e de brinde vira um log persistente pra depurar
// problemas relatados pelo usuário.
func redirecionarStdioParaArquivo() {
	dir, err := os.UserCacheDir() // %LocalAppData% no Windows
	if err != nil {
		dir = os.TempDir()
	}
	logDir := filepath.Join(dir, "RentabilidadeXP")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}

	// Abre em modo APPEND, não O_TRUNC: com truncate, o log de uma falha
	// morria no instante em que o assessor reabria o app pra tentar de
	// novo — que é exatamente o que ele faz antes de pedir ajuda. Ficamos
	// sem o registro justo do problema que queríamos investigar.
	//
	// Pra o arquivo não crescer pra sempre, ele é zerado quando passa de
	// limiteLogBytes: as sessões recentes cabem de sobra nesse tamanho.
	caminhoLog := filepath.Join(logDir, "app.log")
	const limiteLogBytes = 2 << 20 // 2MB
	if info, err := os.Stat(caminhoLog); err == nil && info.Size() > limiteLogBytes {
		os.Remove(caminhoLog)
	}

	arquivo, err := os.OpenFile(caminhoLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	os.Stdout = arquivo
	os.Stderr = arquivo
	log.SetOutput(arquivo)

	// Com o log acumulando várias sessões, é preciso saber onde cada
	// abertura começa e qual versão a gerou.
	log.Printf("=== Ferramentas de Assessoria %s — sessão iniciada ===", Version)
}
