// Package assinaturas gerencia os arquivos de imagem das assinaturas
// pré-salvas (aba Assinatura em Configurações) — usadas pela ferramenta
// "Imagem" do Editor de PDF pra carimbar a assinatura ativa sem precisar
// escolher o arquivo toda vez. Não sabe onde a pasta fica nem lê
// config.json — isso é responsabilidade de quem chama (ver app.go).
package assinaturas

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Assinatura é uma imagem de assinatura salva, com o conteúdo já em
// base64 (pronto pro frontend exibir a miniatura).
type Assinatura struct {
	Nome   string
	Base64 string
}

// Listar devolve todas as assinaturas salvas em pasta. Pasta inexistente
// não é erro — só significa "nenhuma assinatura ainda".
func Listar(pasta string) ([]Assinatura, error) {
	entradas, err := os.ReadDir(pasta)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listar assinaturas: %w", err)
	}

	var lista []Assinatura
	for _, entrada := range entradas {
		if entrada.IsDir() {
			continue
		}
		dados, err := os.ReadFile(filepath.Join(pasta, entrada.Name()))
		if err != nil {
			continue // um arquivo ilegível não deve travar a lista inteira
		}
		lista = append(lista, Assinatura{Nome: entrada.Name(), Base64: base64.StdEncoding.EncodeToString(dados)})
	}
	return lista, nil
}

// Adicionar copia o arquivo em origem pra dentro de pasta (criando-a se
// preciso), com deduplicação de nome (ex.: "assinatura(1).png" se
// "assinatura.png" já existir), e devolve o nome final salvo.
func Adicionar(pasta, origem string) (string, error) {
	if err := os.MkdirAll(pasta, 0755); err != nil {
		return "", fmt.Errorf("criar pasta de assinaturas: %w", err)
	}

	nome := nomeDisponivel(pasta, filepath.Base(origem))
	entrada, err := os.Open(origem)
	if err != nil {
		return "", fmt.Errorf("abrir arquivo: %w", err)
	}
	defer entrada.Close()

	saida, err := os.Create(filepath.Join(pasta, nome))
	if err != nil {
		return "", fmt.Errorf("salvar assinatura: %w", err)
	}
	defer saida.Close()

	if _, err := io.Copy(saida, entrada); err != nil {
		return "", fmt.Errorf("copiar assinatura: %w", err)
	}
	return nome, nil
}

// nomeDisponivel devolve nome, ou "base(1).ext", "base(2).ext" etc. se já
// existir um arquivo com esse nome em pasta.
func nomeDisponivel(pasta, nome string) string {
	if _, err := os.Stat(filepath.Join(pasta, nome)); os.IsNotExist(err) {
		return nome
	}

	ext := filepath.Ext(nome)
	base := strings.TrimSuffix(nome, ext)
	for i := 1; ; i++ {
		candidato := fmt.Sprintf("%s(%d)%s", base, i, ext)
		if _, err := os.Stat(filepath.Join(pasta, candidato)); os.IsNotExist(err) {
			return candidato
		}
	}
}

// Remover apaga o arquivo nome de dentro de pasta. Já não existir não é
// erro (idempotente).
func Remover(pasta, nome string) error {
	if err := os.Remove(filepath.Join(pasta, nome)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remover assinatura: %w", err)
	}
	return nil
}
