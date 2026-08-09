package clientdb

import (
	"os"
	"path/filepath"
	"testing"
)

func escrever(t *testing.T, conteudo string) string {
	t.Helper()
	dir := t.TempDir()
	caminho := filepath.Join(dir, "base.csv")
	if err := os.WriteFile(caminho, []byte(conteudo), 0644); err != nil {
		t.Fatalf("escrever arquivo de teste: %v", err)
	}
	return caminho
}

func TestCarregarBaseClientesComCabecalhoPontoEVirgula(t *testing.T) {
	caminho := escrever(t, "Código;Nome\n4312514;Gustavo Teste\n9999999;Outra Pessoa\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["4312514"].Nome != "Gustavo Teste" {
		t.Errorf("base[4312514].Nome = %q, esperado Gustavo Teste", base["4312514"].Nome)
	}
	if base["9999999"].Nome != "Outra Pessoa" {
		t.Errorf("base[9999999].Nome = %q, esperado Outra Pessoa", base["9999999"].Nome)
	}
}

func TestCarregarBaseClientesComVirgula(t *testing.T) {
	caminho := escrever(t, "conta,nome\n111,Fulano\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["111"].Nome != "Fulano" {
		t.Errorf("base[111].Nome = %q, esperado Fulano", base["111"].Nome)
	}
}

func TestCarregarBaseClientesColunasInvertidas(t *testing.T) {
	caminho := escrever(t, "nome;codigo\nCiclana;222\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["222"].Nome != "Ciclana" {
		t.Errorf("base[222].Nome = %q, esperado Ciclana", base["222"].Nome)
	}
}

func TestCarregarBaseClientesSemCabecalhoReconhecido(t *testing.T) {
	caminho := escrever(t, "333;Beltrano\n444;Sicrano\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if len(base) != 2 || base["333"].Nome != "Beltrano" || base["444"].Nome != "Sicrano" {
		t.Errorf("base = %+v, esperado {333:Beltrano 444:Sicrano}", base)
	}
}

func TestCarregarBaseClientesLinhasEmBrancoIgnoradas(t *testing.T) {
	caminho := escrever(t, "codigo;nome\n555;Fulano\n\n\n666;Ciclano\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if len(base) != 2 {
		t.Errorf("esperava 2 registros, veio %d: %+v", len(base), base)
	}
}

func TestCarregarBaseClientesComBOM(t *testing.T) {
	conteudo := "\uFEFFcodigo;nome\n777;Com BOM\n"
	caminho := escrever(t, conteudo)
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["777"].Nome != "Com BOM" {
		t.Errorf("base[777].Nome = %q, esperado 'Com BOM' (BOM não foi descartado do cabeçalho?)", base["777"].Nome)
	}
}

func TestCarregarBaseClientesComTelefone(t *testing.T) {
	caminho := escrever(t, "codigo;nome;telefone\n888;Fulano;(11) 91234-5678\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["888"].Telefone != "(11) 91234-5678" {
		t.Errorf("base[888].Telefone = %q, esperado '(11) 91234-5678'", base["888"].Telefone)
	}
}

func TestCarregarBaseClientesComCelularOuWhatsapp(t *testing.T) {
	caminho := escrever(t, "codigo,nome,celular\n1,Fulano,11999998888\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["1"].Telefone != "11999998888" {
		t.Errorf("base[1].Telefone = %q, esperado 11999998888", base["1"].Telefone)
	}
}

func TestCarregarBaseClientesSemColunaTelefone(t *testing.T) {
	caminho := escrever(t, "codigo;nome\n2;Fulano\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["2"].Telefone != "" {
		t.Errorf("base[2].Telefone = %q, esperado vazio", base["2"].Telefone)
	}
}

func TestCarregarBaseClientesComEmail(t *testing.T) {
	caminho := escrever(t, "codigo;nome;email\n3;Fulano;fulano@exemplo.com\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["3"].Email != "fulano@exemplo.com" {
		t.Errorf("base[3].Email = %q, esperado fulano@exemplo.com", base["3"].Email)
	}
}

func TestCarregarBaseClientesComEmailHifenizado(t *testing.T) {
	caminho := escrever(t, "codigo;nome;E-mail\n4;Fulano;fulano@exemplo.com\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["4"].Email != "fulano@exemplo.com" {
		t.Errorf("base[4].Email = %q, esperado fulano@exemplo.com (cabeçalho 'E-mail' não reconhecido?)", base["4"].Email)
	}
}

func TestCarregarBaseClientesSemColunaEmail(t *testing.T) {
	caminho := escrever(t, "codigo;nome\n5;Fulano\n")
	base, err := CarregarBaseClientes(caminho)
	if err != nil {
		t.Fatalf("CarregarBaseClientes: %v", err)
	}
	if base["5"].Email != "" {
		t.Errorf("base[5].Email = %q, esperado vazio", base["5"].Email)
	}
}

func TestNormalizarCabecalho(t *testing.T) {
	casos := map[string]string{
		"Código": "codigo",
		"CÓDIGO": "codigo",
		" Nome ": "nome",
		"conta":  "conta",
		"E-mail": "e-mail",
	}
	for entrada, esperado := range casos {
		if got := normalizarCabecalho(entrada); got != esperado {
			t.Errorf("normalizarCabecalho(%q) = %q, esperado %q", entrada, got, esperado)
		}
	}
}
