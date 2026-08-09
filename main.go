package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
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

	arquivo, err := os.OpenFile(filepath.Join(logDir, "app.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}

	os.Stdout = arquivo
	os.Stderr = arquivo
	log.SetOutput(arquivo)
}
