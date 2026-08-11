//go:build windows

package shortcut

import (
	"os"
	"path/filepath"
	"testing"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// criarAtalho cria um .lnk de verdade em caminhoLnk apontando pra alvo —
// usa a mesma automação COM que o código de produção, então também serve
// de confirmação de que os nomes de método/propriedade (CreateShortcut,
// TargetPath, Save) estão certos — só é possível estar errado em tempo de
// execução, já que oleutil resolve tudo por nome (late binding), sem
// checagem do compilador.
func criarAtalho(t *testing.T, caminhoLnk, alvo string) {
	t.Helper()
	if err := ole.CoInitialize(0); err != nil {
		t.Fatalf("iniciar COM: %v", err)
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		t.Fatalf("criar WScript.Shell: %v", err)
	}
	defer unknown.Release()
	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		t.Fatalf("comunicar com WScript.Shell: %v", err)
	}
	defer shell.Release()

	linkVariant, err := oleutil.CallMethod(shell, "CreateShortcut", caminhoLnk)
	if err != nil {
		t.Fatalf("CreateShortcut: %v", err)
	}
	link := linkVariant.ToIDispatch()
	defer link.Release()
	if _, err := oleutil.PutProperty(link, "TargetPath", alvo); err != nil {
		t.Fatalf("definir TargetPath: %v", err)
	}
	if _, err := oleutil.CallMethod(link, "Save"); err != nil {
		t.Fatalf("salvar atalho: %v", err)
	}
}

func lerAlvo(t *testing.T, caminhoLnk string) string {
	t.Helper()
	if err := ole.CoInitialize(0); err != nil {
		t.Fatalf("iniciar COM: %v", err)
	}
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		t.Fatalf("criar WScript.Shell: %v", err)
	}
	defer unknown.Release()
	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		t.Fatalf("comunicar com WScript.Shell: %v", err)
	}
	defer shell.Release()

	linkVariant, err := oleutil.CallMethod(shell, "CreateShortcut", caminhoLnk)
	if err != nil {
		t.Fatalf("abrir atalho existente: %v", err)
	}
	link := linkVariant.ToIDispatch()
	defer link.Release()
	alvoVariant, err := oleutil.GetProperty(link, "TargetPath")
	if err != nil {
		t.Fatalf("ler TargetPath: %v", err)
	}
	return alvoVariant.ToString()
}

func TestRetargetarEmReapontaAtalhoQueBate(t *testing.T) {
	dir := t.TempDir()
	alvoAntigo := filepath.Join(dir, "app-antigo.exe")
	alvoNovo := filepath.Join(dir, "app-novo.exe")
	// O Shortcut do WScript.Shell normaliza/rejeita TargetPath de arquivo
	// que não existe — os dois alvos precisam existir de verdade, igual no
	// caso real (o executável atual sempre existe; o novo só é reapontado
	// depois de confirmado presente, ver relancar_windows.go).
	if err := os.WriteFile(alvoAntigo, []byte("x"), 0644); err != nil {
		t.Fatalf("preparar alvo antigo: %v", err)
	}
	if err := os.WriteFile(alvoNovo, []byte("x"), 0644); err != nil {
		t.Fatalf("preparar alvo novo: %v", err)
	}
	caminhoLnk := filepath.Join(dir, "Ferramentas de Assessoria.lnk")
	criarAtalho(t, caminhoLnk, alvoAntigo)

	n, err := retargetarEm([]string{dir}, alvoAntigo, alvoNovo)
	if err != nil {
		t.Fatalf("retargetarEm: %v", err)
	}
	if n != 1 {
		t.Errorf("atualizados = %d, esperado 1", n)
	}
	if got := lerAlvo(t, caminhoLnk); got != alvoNovo {
		t.Errorf("alvo do atalho depois = %q, esperado %q", got, alvoNovo)
	}
}

func TestRetargetarEmIgnoraAtalhoQueNaoBate(t *testing.T) {
	dir := t.TempDir()
	alvoDeOutraCoisa := filepath.Join(dir, "outro-programa.exe")
	alvoAntigo := filepath.Join(dir, "app-antigo.exe")
	alvoNovo := filepath.Join(dir, "app-novo.exe")
	caminhoLnk := filepath.Join(dir, "Outro Programa.lnk")
	criarAtalho(t, caminhoLnk, alvoDeOutraCoisa)

	n, err := retargetarEm([]string{dir}, alvoAntigo, alvoNovo)
	if err != nil {
		t.Fatalf("retargetarEm: %v", err)
	}
	if n != 0 {
		t.Errorf("atualizados = %d, esperado 0 (atalho de outro programa não deveria ser mexido)", n)
	}
	if got := lerAlvo(t, caminhoLnk); got != alvoDeOutraCoisa {
		t.Errorf("alvo do atalho de outro programa foi alterado: %q", got)
	}
}

func TestRetargetarEmPastaSemAtalhos(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "nao-e-atalho.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("preparar arquivo: %v", err)
	}

	n, err := retargetarEm([]string{dir}, "C:\\qualquer.exe", "C:\\outro.exe")
	if err != nil {
		t.Fatalf("retargetarEm: %v", err)
	}
	if n != 0 {
		t.Errorf("atualizados = %d, esperado 0", n)
	}
}

func TestRetargetarEmPastaInexistenteNaoFalha(t *testing.T) {
	n, err := retargetarEm([]string{filepath.Join(t.TempDir(), "nao-existe")}, "C:\\a.exe", "C:\\b.exe")
	if err != nil {
		t.Fatalf("pasta inexistente não deveria causar erro, veio: %v", err)
	}
	if n != 0 {
		t.Errorf("atualizados = %d, esperado 0", n)
	}
}
