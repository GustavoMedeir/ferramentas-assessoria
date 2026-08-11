//go:build windows

// Package shortcut reaponta atalhos (.lnk) da área de trabalho que apontam
// pra um executável antigo, fazendo-os apontar pro novo — usado pelo
// relançador da atualização automática (ver internal/selfupdate) depois
// que a nova versão é baixada em Downloads: sem isso, o atalho que o
// assessor criou manualmente (ver README) ficaria apontando pro executável
// antigo pra sempre.
//
// Usa a automação COM WScript.Shell (a mesma técnica de scripts .vbs/
// PowerShell clássicos pra ler/editar atalhos), no mesmo padrão de
// automação leve (late-bound, via oleutil) já usado em
// internal/outlookmail pro Outlook.
package shortcut

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// RetargetarNoDesktop procura, entre os atalhos (.lnk) da área de trabalho
// do usuário e da área de trabalho pública (%PUBLIC%\Desktop — usada por
// atalhos "para todos os usuários"), os que apontam pra alvoAntigo e
// reaponta pra alvoNovo. Devolve quantos atalhos foram atualizados.
//
// Não mexe em atalhos no Menu Iniciar ou fixados na barra de tarefas —
// cobrem casos mais raros nesse app (instalação manual, sem instalador) e
// cada um tem seu próprio formato/risco de automação; a área de trabalho é
// onde o README instrui o assessor a criar o atalho.
func RetargetarNoDesktop(alvoAntigo, alvoNovo string) (int, error) {
	pastas, err := pastasDesktop()
	if err != nil {
		return 0, err
	}
	return retargetarEm(pastas, alvoAntigo, alvoNovo)
}

// retargetarEm faz o trabalho de verdade, recebendo a lista de pastas a
// vasculhar — separado de RetargetarNoDesktop só pra poder testar a
// automação COM de verdade (criar um .lnk, reapontar, conferir) contra uma
// pasta temporária, sem mexer na área de trabalho de quem roda os testes.
func retargetarEm(pastas []string, alvoAntigo, alvoNovo string) (int, error) {
	if err := ole.CoInitialize(0); err != nil {
		return 0, fmt.Errorf("iniciar COM: %w", err)
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return 0, fmt.Errorf("criar WScript.Shell: %w", err)
	}
	defer unknown.Release()
	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return 0, fmt.Errorf("comunicar com WScript.Shell: %w", err)
	}
	defer shell.Release()

	atualizados := 0
	for _, pasta := range pastas {
		entradas, err := os.ReadDir(pasta)
		if err != nil {
			continue // pasta pode não existir (ex.: sem área pública) — segue pra próxima
		}
		for _, entrada := range entradas {
			if entrada.IsDir() || !strings.EqualFold(filepath.Ext(entrada.Name()), ".lnk") {
				continue
			}
			ok, err := retargetarUm(shell, filepath.Join(pasta, entrada.Name()), alvoAntigo, alvoNovo)
			if err != nil {
				continue // um atalho problemático não pode travar os outros
			}
			if ok {
				atualizados++
			}
		}
	}
	return atualizados, nil
}

// retargetarUm abre um único atalho existente (CreateShortcut num .lnk que
// já existe carrega as propriedades dele, só grava algo no disco se Save()
// for chamado) e, se o alvo bater com alvoAntigo, reaponta pra alvoNovo.
func retargetarUm(shell *ole.IDispatch, caminhoLnk, alvoAntigo, alvoNovo string) (bool, error) {
	linkVariant, err := oleutil.CallMethod(shell, "CreateShortcut", caminhoLnk)
	if err != nil {
		return false, err
	}
	link := linkVariant.ToIDispatch()
	defer link.Release()

	alvoAtualVariant, err := oleutil.GetProperty(link, "TargetPath")
	if err != nil {
		return false, err
	}
	if !strings.EqualFold(alvoAtualVariant.ToString(), alvoAntigo) {
		return false, nil
	}

	if _, err := oleutil.PutProperty(link, "TargetPath", alvoNovo); err != nil {
		return false, err
	}
	if _, err := oleutil.CallMethod(link, "Save"); err != nil {
		return false, err
	}
	return true, nil
}

func pastasDesktop() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("localizar pasta do usuário: %w", err)
	}
	pastas := []string{filepath.Join(home, "Desktop")}
	if publico := os.Getenv("PUBLIC"); publico != "" {
		pastas = append(pastas, filepath.Join(publico, "Desktop"))
	}
	return pastas, nil
}
