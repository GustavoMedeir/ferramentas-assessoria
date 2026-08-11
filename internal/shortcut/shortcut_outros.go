//go:build !windows

package shortcut

// RetargetarNoDesktop fora do Windows: atalhos .lnk e automação COM não
// existem fora do Windows, e a atualização automática já é Windows-only
// (ver checarAtualizacao em app.go) — mantido só pra internal/selfupdate
// compilar em qualquer plataforma (facilita `go test ./...`/análise
// estática num Mac, por exemplo).
func RetargetarNoDesktop(alvoAntigo, alvoNovo string) (int, error) {
	return 0, nil
}
