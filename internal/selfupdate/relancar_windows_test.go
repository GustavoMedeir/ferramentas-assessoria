//go:build windows

package selfupdate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecutavelValido(t *testing.T) {
	dir := t.TempDir()

	inexistente := filepath.Join(dir, "nao-existe.exe")
	if executavelValido(inexistente) {
		t.Error("arquivo inexistente não deveria ser considerado válido")
	}

	pequeno := filepath.Join(dir, "truncado.exe")
	if err := os.WriteFile(pequeno, make([]byte, 1024), 0755); err != nil {
		t.Fatalf("preparar arquivo pequeno: %v", err)
	}
	if executavelValido(pequeno) {
		t.Error("arquivo abaixo do piso de tamanho não deveria ser considerado válido")
	}

	valido := filepath.Join(dir, "app.exe")
	if err := os.WriteFile(valido, make([]byte, tamanhoMinimoValido+1024), 0755); err != nil {
		t.Fatalf("preparar arquivo válido: %v", err)
	}
	if !executavelValido(valido) {
		t.Error("arquivo presente e acima do piso de tamanho deveria ser considerado válido")
	}
}
