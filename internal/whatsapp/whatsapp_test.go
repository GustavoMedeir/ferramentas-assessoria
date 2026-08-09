package whatsapp

import (
	"net/url"
	"strings"
	"testing"
)

func TestMontarLinkComDDDSemCodigoDoPais(t *testing.T) {
	link, err := MontarLink("(11) 91234-5678", "olá")
	if err != nil {
		t.Fatalf("MontarLink: %v", err)
	}
	if !strings.HasPrefix(link, "https://wa.me/5511912345678?") {
		t.Errorf("link = %q, esperado prefixo https://wa.me/5511912345678?", link)
	}
}

func TestMontarLinkJaComCodigoDoPais(t *testing.T) {
	link, err := MontarLink("+55 11 91234-5678", "olá")
	if err != nil {
		t.Fatalf("MontarLink: %v", err)
	}
	if !strings.HasPrefix(link, "https://wa.me/5511912345678?") {
		t.Errorf("link = %q, esperado prefixo https://wa.me/5511912345678?", link)
	}
}

func TestMontarLinkTelefoneFixoComDDD(t *testing.T) {
	link, err := MontarLink("11 3123-4567", "olá")
	if err != nil {
		t.Fatalf("MontarLink: %v", err)
	}
	if !strings.HasPrefix(link, "https://wa.me/551131234567?") {
		t.Errorf("link = %q, esperado prefixo https://wa.me/551131234567?", link)
	}
}

func TestMontarLinkMensagemVaiCodificadaNoQuery(t *testing.T) {
	link, err := MontarLink("11912345678", "Olá, tudo bem?")
	if err != nil {
		t.Fatalf("MontarLink: %v", err)
	}
	idx := strings.Index(link, "?")
	if idx == -1 {
		t.Fatalf("link sem query string: %q", link)
	}
	valores, err := url.ParseQuery(link[idx+1:])
	if err != nil {
		t.Fatalf("ParseQuery: %v", err)
	}
	if valores.Get("text") != "Olá, tudo bem?" {
		t.Errorf("text = %q, esperado 'Olá, tudo bem?'", valores.Get("text"))
	}
}

func TestMontarLinkTelefoneInvalido(t *testing.T) {
	for _, telefone := range []string{"", "123", "abc"} {
		if _, err := MontarLink(telefone, "olá"); err == nil {
			t.Errorf("MontarLink(%q, ...) esperava erro, não deu", telefone)
		}
	}
}
