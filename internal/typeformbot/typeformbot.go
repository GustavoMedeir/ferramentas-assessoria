// Package typeformbot preenche o Typeform online de diagnóstico financeiro
// com as respostas já coletadas na aba Typeform do app, controlando o
// Microsoft Edge instalado no Windows (mesmo motor do WebView2 que o app já
// exige pra rodar — sem baixar nenhum navegador extra).
//
// O formulário real muda com o tempo (perguntas reformuladas, opções
// reescritas, perguntas condicionais novas) — por isso a leitura de cada
// tela é sempre ao vivo (tipo de campo, texto da pergunta e opções vêm do
// DOM na hora, nunca de uma cópia estática guardada no app) e o casamento
// entre a pergunta exibida e a resposta salva é por texto aproximado, não
// por posição fixa. Quando uma tela não bate com confiança suficiente, ou é
// a última resposta que temos, o preenchimento para ali — o navegador fica
// aberto e visível pra o assessor terminar e enviar manualmente. O envio
// final nunca é automático.
package typeformbot

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"rentabilidade/internal/pdfreport"
)

// Resposta é uma pergunta (mesmo texto exibido no Typeform) e a resposta já
// formatada como aparece no .txt salvo pela aba (ver respostaFormatada em
// typeform.js) — o preenchimento decide como aplicar esse texto de acordo
// com o tipo de campo encontrado ao vivo na tela, não com base num tipo
// vindo do app.
type Resposta struct {
	Pergunta string
	Valor    string
}

// EventoProgresso é chamado a cada pergunta preenchida com sucesso.
type EventoProgresso func(feitos, total int, pergunta string)

// ErroParado é devolvido quando o preenchimento para antes do fim — tela
// sem resposta correspondente, tela sem confiança suficiente, ou a última
// resposta disponível (nunca envia sozinho). O navegador continua aberto na
// tela em questão.
type ErroParado struct {
	Pergunta string
	Motivo   string
}

func (e *ErroParado) Error() string {
	return fmt.Sprintf("parei em %q: %s", e.Pergunta, e.Motivo)
}

// limiteTelas evita loop infinito se o formulário real tiver muito mais
// perguntas do que o esperado (ou alguma condição travar a detecção de
// tela sempre na mesma).
const limiteTelas = 200

// localizarEdge tenta achar o executável do Microsoft Edge instalado no
// Windows nos caminhos padrão de instalação — o mesmo motor Chromium do
// WebView2 que o app já exige pra rodar, então não precisa baixar nenhum
// navegador extra.
func localizarEdge() (string, error) {
	candidatos := []string{
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
	}
	// Instalação por usuário (sem admin) fica em %LOCALAPPDATA% — não é o
	// padrão, mas acontece em máquina onde o Edge foi reinstalado sem
	// privilégio de administrador, e sem isso o app falhava dizendo que o
	// Edge não existe numa máquina onde ele existe.
	if local, err := os.UserCacheDir(); err == nil {
		candidatos = append(candidatos, filepath.Join(local, `Microsoft\Edge\Application\msedge.exe`))
	}
	for _, c := range candidatos {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("Microsoft Edge não encontrado nos caminhos padrão de instalação (ele já vem com o Windows — se foi removido, reinstale)")
}

// aguardarConteudo espera a página realmente ter conteúdo utilizável (algum
// botão visível ou algum bloco de pergunta), em vez de confiar num sleep
// fixo.
//
// Existe por causa de um bug relatado em campo: em algumas máquinas o Edge
// abria numa página em branco e o robô, que só esperava 450ms e não checava
// nada, lia um DOM vazio, não achava botão nenhum pra clicar e encerrava
// reportando SUCESSO — o app dizia "preenchimento concluído" com a tela
// vazia na frente do assessor. Agora a ausência de conteúdo é detectada e
// vira erro explicativo (ver o uso em Preencher).
func aguardarConteudo(ctx context.Context, limite time.Duration) (bool, error) {
	const sonda = `(function(){
	var visiveis = Array.from(document.querySelectorAll('button')).filter(function(b){
		var r = b.getBoundingClientRect(); return r.width > 0 && r.height > 0;
	});
	return visiveis.length > 0 || document.querySelectorAll('[data-qa*="blocktype-"]').length > 0;
})()`
	fim := time.Now().Add(limite)
	for {
		var temConteudo bool
		if err := chromedp.Run(ctx, chromedp.Evaluate(sonda, &temConteudo)); err != nil {
			return false, err
		}
		if temConteudo {
			return true, nil
		}
		if time.Now().After(fim) {
			return false, nil
		}
		if err := chromedp.Run(ctx, chromedp.Sleep(250*time.Millisecond)); err != nil {
			return false, err
		}
	}
}

// Preencher abre urlFormulario no Edge (janela visível), percorre as telas
// casando a pergunta exibida com uma resposta em respostas por texto
// aproximado, preenche/seleciona e avança. Para (sem erro nenhum de
// verdade, ver ErroParado) na primeira tela sem correspondência confiável
// ou ao usar a última resposta disponível — nesses dois casos o navegador
// fica aberto na tela pra o assessor continuar e enviar manualmente. Não
// fecha o navegador ao terminar (com ou sem erro): é intencional, o
// assessor precisa da janela aberta pra revisar e enviar.
func Preencher(urlFormulario string, respostas []Resposta, progresso EventoProgresso) error {
	edge, err := localizarEdge()
	if err != nil {
		return err
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(edge),
		chromedp.Flag("headless", false),
		chromedp.Flag("start-maximized", true),
		chromedp.Flag("disable-extensions", true),
	)
	// Sem cancel por defer aqui de propósito: a janela precisa continuar
	// aberta depois que esta função retornar, com ou sem erro — ver
	// comentário do pacote e do doc desta função.
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx)

	ctxComPrazo, cancelPrazo := context.WithTimeout(ctx, 20*time.Minute)
	defer cancelPrazo()

	log.Println("typeform: abrindo", urlFormulario, "no Edge em", edge)
	if err := chromedp.Run(ctxComPrazo, chromedp.Navigate(urlFormulario)); err != nil {
		return fmt.Errorf("abrir formulário no navegador: %w", err)
	}

	// Traz a aba controlada pra frente: se a janela do Edge tiver mais de
	// uma aba (perfil restaurando sessão, página inicial corporativa), a
	// automação pode estar dirigindo uma aba que não é a visível — o
	// assessor veria uma tela parada enquanto o robô trabalha em outra.
	if err := chromedp.Run(ctxComPrazo, chromedp.ActionFunc(func(c context.Context) error {
		return page.BringToFront().Do(c)
	})); err != nil {
		log.Println("typeform: não foi possível trazer a aba pra frente (seguindo mesmo assim):", err)
	}

	// A página em branco relatada em campo morre aqui, com mensagem útil,
	// em vez de virar um "concluído" silencioso lá na frente.
	temConteudo, err := aguardarConteudo(ctxComPrazo, 30*time.Second)
	if err != nil {
		return fmt.Errorf("verificar se o formulário carregou: %w", err)
	}
	if !temConteudo {
		var urlAtual string
		_ = chromedp.Run(ctxComPrazo, chromedp.Location(&urlAtual))
		log.Println("typeform: página sem conteúdo depois de 30s; url atual =", urlAtual)
		return fmt.Errorf(
			"o formulário não carregou nesta máquina — a página ficou em branco (endereço atual: %s).\n\n"+
				"Isso costuma ser bloqueio de rede: teste abrir %s no Edge manualmente. "+
				"Se pedir login de proxy ou não abrir, é a rede/política da máquina que está barrando.",
			urlAtual, urlFormulario)
	}

	usadas := make([]bool, len(respostas))
	total := len(respostas)
	feitos := 0
	// Limita quantas vezes o clique genérico (tela de boas-vindas) pode
	// rodar — sem isso, uma tela intermediária de um tipo que não sabemos
	// preencher (ex.: pergunta de "opinião"/escala, não implementada) faz o
	// loop tentar clicar em qualquer botão visível indefinidamente até
	// limiteTelas, sem nunca progredir de verdade (bug real observado: 200
	// iterações de ~1s cada, quase 3 minutos, sem sair do lugar).
	tentativasCliqueGenerico := 0
	const maxTentativasCliqueGenerico = 3

	for i := 0; i < limiteTelas; i++ {
		if err := aguardarCarregamento(ctxComPrazo); err != nil {
			return fmt.Errorf("aguardar formulário carregar: %w", err)
		}

		tela, err := lerTelaAtual(ctxComPrazo)
		if err != nil {
			return fmt.Errorf("ler tela atual do formulário: %w", err)
		}
		if !tela.Found {
			// Tentativa de tela de boas-vindas ("Começar") ou fim do
			// formulário (tela de agradecimento): tenta um clique
			// genérico algumas vezes e olha de novo; se continuar sem
			// achar pergunta reconhecida, considera que chegou ao fim (ou
			// que é um tipo de pergunta que ainda não sabemos preencher).
			if tentativasCliqueGenerico < maxTentativasCliqueGenerico && cliqueGenerico(ctxComPrazo) {
				tentativasCliqueGenerico++
				continue
			}
			// Chegar aqui sem ter preenchido NADA não é "fim do
			// formulário": é o robô não ter entendido a tela nenhuma vez.
			// Reportar sucesso nesse caso era o bug relatado em campo — o
			// app dizia "preenchimento concluído" com a tela intocada na
			// frente do assessor, sem nenhuma pista do que houve.
			if feitos == 0 {
				log.Println("typeform: nenhuma pergunta reconhecida na primeira tela útil — encerrando com aviso")
				return &ErroParado{
					Pergunta: "",
					Motivo: "não consegui reconhecer nenhuma pergunta do formulário — ele pode ter mudado de formato, " +
						"ou a página não terminou de carregar nesta máquina. Preencha manualmente desta vez",
				}
			}
			// Com perguntas já preenchidas, cair aqui é o fim normal do
			// formulário (tela de agradecimento).
			log.Println("typeform: fim do formulário depois de", feitos, "pergunta(s)")
			return nil
		}
		tentativasCliqueGenerico = 0

		indice, pontuacao := melhorCorrespondencia(tela.Titulo, respostas, usadas)
		if indice == -1 || pontuacao < 0.4 {
			return &ErroParado{
				Pergunta: tela.Titulo,
				Motivo:   "não encontrei uma resposta salva com confiança suficiente pra essa pergunta",
			}
		}

		ultimaDisponivel := contarNaoUsadas(usadas) == 1
		if err := preencherTela(ctxComPrazo, tela, respostas[indice].Valor); err != nil {
			return fmt.Errorf("preencher %q: %w", tela.Titulo, err)
		}
		usadas[indice] = true
		feitos++
		if progresso != nil {
			progresso(feitos, total, tela.Titulo)
		}

		if ultimaDisponivel {
			return &ErroParado{
				Pergunta: tela.Titulo,
				Motivo:   "essa era a última resposta que eu tinha salva — confira o restante e envie você mesmo",
			}
		}

		if err := avancarTela(ctxComPrazo, tela); err != nil {
			return fmt.Errorf("avançar depois de %q: %w", tela.Titulo, err)
		}
	}

	return &ErroParado{Pergunta: "", Motivo: "o formulário tem mais telas do que o esperado — parei por segurança"}
}

func contarNaoUsadas(usadas []bool) int {
	n := 0
	for _, u := range usadas {
		if !u {
			n++
		}
	}
	return n
}

// aguardarCarregamento dá um respiro pra animação de transição entre telas
// do Typeform terminar antes de ler o DOM de novo.
func aguardarCarregamento(ctx context.Context) error {
	return chromedp.Run(ctx, chromedp.Sleep(450*time.Millisecond))
}

// cliqueGenerico procura qualquer botão visível na página (usado só pra
// passar da tela de boas-vindas, que não tem os data-qa de pergunta) e
// clica nele. Devolve false se não achou nenhum — sinal de que realmente
// chegou ao fim do formulário, não que é só uma tela de boas-vindas.
func cliqueGenerico(ctx context.Context) bool {
	const script = `
(function() {
	var botoes = Array.from(document.querySelectorAll('button'));
	var visivel = botoes.find(function(b) {
		var r = b.getBoundingClientRect();
		return r.width > 0 && r.height > 0;
	});
	if (!visivel) return false;
	visivel.click();
	return true;
})()`
	var clicou bool
	if err := chromedp.Run(ctx, chromedp.Evaluate(script, &clicou)); err != nil {
		return false
	}
	if clicou {
		chromedp.Run(ctx, chromedp.Sleep(450*time.Millisecond))
	}
	return clicou
}

// ---------------------------------------------------------------------------
// Leitura da tela atual
// ---------------------------------------------------------------------------

// telaAtual é o resultado de scriptMarcarEExtrair: a pergunta mais próxima
// do centro da tela (o Typeform mantém telas vizinhas montadas no DOM
// durante a transição de rolagem, por isso não basta pegar "a primeira").
type telaAtual struct {
	Found        bool     `json:"found"`
	Tipo         string   `json:"tipo"` // short_text | email | number | date | multiple_choice | outro
	Titulo       string   `json:"titulo"`
	MultiSelecao bool     `json:"multiSelecao"` // só relevante quando Tipo == multiple_choice
	Opcoes       []string `json:"opcoes"`
}

// marcaAlvo é o atributo que scriptMarcarEExtrair grava no bloco da
// pergunta ativa, pra as ações seguintes (preencher, clicar) mirarem nele
// com um seletor CSS estável, sem depender do id com UUID que o Typeform
// gera pra cada instância de pergunta.
const marcaAlvo = `data-tf-alvo`

const scriptMarcarEExtrair = `
(function() {
	document.querySelectorAll('[` + marcaAlvo + `]').forEach(function(el) {
		el.removeAttribute('` + marcaAlvo + `');
	});
	// Só considera blocos de tipo que sabemos preencher — telas de
	// boas-vindas/agradecimento ("blocktype-statement") não têm
	// block-title nem input, e cairiam como "pergunta sem correspondência"
	// por engano em vez de serem tratadas como início/fim do formulário.
	var tiposConhecidos = ['short_text', 'email', 'number', 'date', 'multiple_choice', 'yes_no'];
	var blocos = Array.from(document.querySelectorAll('[data-qa*="blocktype-"]')).filter(function(b) {
		var dataQa = (b.getAttribute('data-qa') || '').split(' ')[0].replace('blocktype-', '');
		return tiposConhecidos.indexOf(dataQa) !== -1;
	});
	if (blocos.length === 0) return { found: false };

	// As perguntas ficam todas empilhadas no mesmo retângulo (o Typeform
	// não desloca verticalmente entre telas) e a transição é só um
	// crossfade de opacidade — por isso "mais visível" (maior opacity), não
	// "mais perto do centro", é o critério certo pra achar a pergunta ativa.
	var melhor = null, melhorOpacidade = -1;
	blocos.forEach(function(b) {
		var r = b.getBoundingClientRect();
		if (r.height === 0) return;
		var opacidade = parseFloat(getComputedStyle(b).opacity);
		if (isNaN(opacidade)) opacidade = 1;
		if (opacidade > melhorOpacidade) { melhorOpacidade = opacidade; melhor = b; }
	});
	if (!melhor || melhorOpacidade < 0.5) return { found: false };

	melhor.setAttribute('` + marcaAlvo + `', '1');
	var dataQa = melhor.getAttribute('data-qa') || '';
	var tipo = dataQa.split(' ')[0].replace('blocktype-', '');

	function texto(el) { return el ? (el.innerText || '').trim() : ''; }
	// Igual ao filtro de blocktype acima: o Typeform concatena o data-qa
	// com um sufixo de tema ("block-title deep-purple-block-title"), por
	// isso "contém" (*=) em vez de igualdade exata.
	var titulo = texto(melhor.querySelector('[data-qa*="block-title"]'));

	var botoesCheck = melhor.querySelectorAll('button[role="checkbox"]');
	var multi = botoesCheck.length > 0;
	var botoes = multi ? botoesCheck : melhor.querySelectorAll('button[role="radio"]');
	var opcoes = [];
	botoes.forEach(function(btn) {
		var t = texto(btn);
		var linhas = t.split('\n');
		opcoes.push((linhas.length > 1 ? linhas.slice(1).join(' ') : t).trim());
	});

	return { found: true, tipo: tipo, titulo: titulo, multiSelecao: multi, opcoes: opcoes };
})()`

func lerTelaAtual(ctx context.Context) (*telaAtual, error) {
	var res telaAtual
	if err := chromedp.Run(ctx, chromedp.Evaluate(scriptMarcarEExtrair, &res)); err != nil {
		return nil, err
	}
	return &res, nil
}

// ---------------------------------------------------------------------------
// Casamento pergunta exibida <-> resposta salva
// ---------------------------------------------------------------------------

// removedorAcentos troca as letras acentuadas comuns do pt-BR pela versão
// sem acento — evita depender de um pacote extra só pra normalizar texto
// nesse conjunto pequeno e conhecido de caracteres.
var removedorAcentos = strings.NewReplacer(
	"á", "a", "à", "a", "ã", "a", "â", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "õ", "o", "ô", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n",
)

func normalizar(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = removedorAcentos.Replace(s)
	var sb strings.Builder
	anteriorEspaco := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			sb.WriteRune(r)
			anteriorEspaco = false
		default:
			if !anteriorEspaco {
				sb.WriteByte(' ')
				anteriorEspaco = true
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

// pontuarSimilaridade mede sobreposição de palavras entre dois textos
// normalizados (interseção sobre união dos conjuntos de palavras) — simples
// de propósito: as perguntas mudam de redação com pequenos ajustes
// (plural, pontuação, uma palavra a mais), não viram outra pergunta do
// zero, então não precisa de nada mais sofisticado que isso.
func pontuarSimilaridade(a, b string) float64 {
	palavrasA := strings.Fields(normalizar(a))
	palavrasB := strings.Fields(normalizar(b))
	if len(palavrasA) == 0 || len(palavrasB) == 0 {
		return 0
	}
	conjA := make(map[string]bool, len(palavrasA))
	for _, p := range palavrasA {
		conjA[p] = true
	}
	conjB := make(map[string]bool, len(palavrasB))
	for _, p := range palavrasB {
		conjB[p] = true
	}
	intersecao := 0
	for p := range conjA {
		if conjB[p] {
			intersecao++
		}
	}
	uniao := len(conjA) + len(conjB) - intersecao
	if uniao == 0 {
		return 0
	}
	return float64(intersecao) / float64(uniao)
}

// melhorCorrespondencia acha, entre as respostas ainda não usadas, a que
// tem o texto de pergunta mais parecido com tituloTela. Devolve índice -1
// se respostas estiver vazio.
func melhorCorrespondencia(tituloTela string, respostas []Resposta, usadas []bool) (indice int, pontuacao float64) {
	indice = -1
	for i, r := range respostas {
		if usadas[i] {
			continue
		}
		p := pontuarSimilaridade(tituloTela, r.Pergunta)
		if p > pontuacao {
			pontuacao = p
			indice = i
		}
	}
	return indice, pontuacao
}

// ---------------------------------------------------------------------------
// Preenchimento por tipo de campo (detectado ao vivo, ver telaAtual.Tipo)
// ---------------------------------------------------------------------------

func preencherTela(ctx context.Context, tela *telaAtual, valor string) error {
	valor = strings.TrimSpace(valor)
	if valor == "" || valor == "—" {
		return nil // sem resposta salva pra essa pergunta — deixa em branco, avança
	}

	switch tela.Tipo {
	case "short_text", "email":
		return preencherTexto(ctx, valor)
	case "number":
		return preencherNumero(ctx, valor)
	case "date":
		return preencherData(ctx, valor)
	case "multiple_choice", "yes_no":
		// yes_no (perguntas Sim/Não — um blocktype à parte no Typeform, não
		// uma variação de multiple_choice) usa por baixo exatamente a mesma
		// estrutura: role="radiogroup" com button[role="radio"] com o texto
		// "S\nSim" / "N\nNão" (só troca a letra-atalho "A/B" por "S/N") —
		// mesmo preenchimento serve pros dois.
		return preencherMultiplaEscolha(ctx, tela, valor)
	default:
		return fmt.Errorf("tipo de campo não suportado: %q", tela.Tipo)
	}
}

func preencherTexto(ctx context.Context, valor string) error {
	sel := fmt.Sprintf(`[%s="1"] input, [%s="1"] textarea`, marcaAlvo, marcaAlvo)
	return chromedp.Run(ctx,
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.Click(sel, chromedp.ByQuery),
		chromedp.SendKeys(sel, valor, chromedp.ByQuery),
	)
}

// preencherNumero limpa a formatação pt-BR (R$, ponto de milhar, vírgula
// decimal) — o app guarda "R$ 10.000,00" mas o campo nativo do Typeform
// (input type=number) só aceita ponto decimal. Reaproveita
// pdfreport.ParseNumeroPtBR (mesma função usada pra ler os relatórios XP),
// evitando duplicar essa lógica.
func preencherNumero(ctx context.Context, valor string) error {
	numero, err := pdfreport.ParseNumeroPtBR(valor)
	if err != nil {
		// Não é um valor numérico reconhecível (ex.: "anos" já vem como
		// "35", ParseNumeroPtBR aceita isso também) — tenta mandar como
		// veio, o próprio campo rejeita se não servir.
		numero = 0
		if v, err2 := strconv.ParseFloat(strings.TrimSpace(valor), 64); err2 == nil {
			numero = v
		} else {
			return fmt.Errorf("valor %q não é numérico: %w", valor, err)
		}
	}
	texto := strconv.FormatFloat(numero, 'f', -1, 64)
	sel := fmt.Sprintf(`[%s="1"] input`, marcaAlvo)
	return chromedp.Run(ctx,
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.Click(sel, chromedp.ByQuery),
		chromedp.SendKeys(sel, texto, chromedp.ByQuery),
	)
}

// preencherData espera valor em "dd/mm/aaaa" (mesmo formato que
// respostaFormatada devolve em typeform.js pra perguntas tipo "data").
func preencherData(ctx context.Context, valor string) error {
	partes := strings.Split(valor, "/")
	if len(partes) != 3 {
		return fmt.Errorf("data %q não está no formato dd/mm/aaaa", valor)
	}
	dia, mes, ano := partes[0], partes[1], partes[2]
	selDia := fmt.Sprintf(`[%s="1"] input[placeholder="DD"]`, marcaAlvo)
	selMes := fmt.Sprintf(`[%s="1"] input[placeholder="MM"]`, marcaAlvo)
	selAno := fmt.Sprintf(`[%s="1"] input[placeholder="AAAA"]`, marcaAlvo)
	return chromedp.Run(ctx,
		chromedp.WaitVisible(selDia, chromedp.ByQuery),
		chromedp.Click(selDia, chromedp.ByQuery),
		chromedp.SendKeys(selDia, dia, chromedp.ByQuery),
		chromedp.Click(selMes, chromedp.ByQuery),
		chromedp.SendKeys(selMes, mes, chromedp.ByQuery),
		chromedp.Click(selAno, chromedp.ByQuery),
		chromedp.SendKeys(selAno, ano, chromedp.ByQuery),
	)
}

// preencherMultiplaEscolha casa cada pedaço de valor (uma pergunta de
// seleção múltipla salva as respostas separadas por ", " — ver
// respostaFormatada em typeform.js) com a opção visível mais parecida e
// clica nela. Seleção única (radio) avança sozinha ao clicar; seleção
// múltipla (checkbox) precisa do clique em OK depois (feito por
// avancarTela).
func preencherMultiplaEscolha(ctx context.Context, tela *telaAtual, valor string) error {
	if len(tela.Opcoes) == 0 {
		return fmt.Errorf("pergunta de múltipla escolha sem opções detectadas na tela")
	}

	partes := []string{valor}
	if tela.MultiSelecao {
		partes = strings.Split(valor, ",")
		for i := range partes {
			partes[i] = strings.TrimSpace(partes[i])
		}
	}

	papel := "radio"
	if tela.MultiSelecao {
		papel = "checkbox"
	}
	sel := fmt.Sprintf(`[%s="1"] button[role="%s"]`, marcaAlvo, papel)

	var nos []*cdp.Node
	if err := chromedp.Run(ctx, chromedp.Nodes(sel, &nos, chromedp.ByQueryAll)); err != nil {
		return err
	}
	if len(nos) != len(tela.Opcoes) {
		return fmt.Errorf("detectei %d opções mas achei %d botões — a tela pode ter mudado", len(tela.Opcoes), len(nos))
	}

	for _, parte := range partes {
		if parte == "" {
			continue
		}
		indiceOpcao, pontuacao := -1, 0.0
		for i, opcao := range tela.Opcoes {
			p := pontuarSimilaridade(parte, opcao)
			if p > pontuacao {
				pontuacao = p
				indiceOpcao = i
			}
		}
		if indiceOpcao == -1 || pontuacao < 0.3 {
			return fmt.Errorf("não achei uma opção parecida com %q entre: %s", parte, strings.Join(tela.Opcoes, " | "))
		}
		if err := chromedp.Run(ctx, chromedp.MouseClickNode(nos[indiceOpcao])); err != nil {
			return err
		}
		// Seleção única clicada = a própria pergunta dispara a transição
		// pra próxima tela (mesma animação de crossfade do clique em OK) —
		// precisa do mesmo tempo de espera, senão a leitura seguinte pega a
		// pergunta ainda saindo. Seleção múltipla só precisa de um respiro
		// curto entre marcar cada opção (não navega sozinha).
		espera := 200 * time.Millisecond
		if !tela.MultiSelecao {
			espera = 700 * time.Millisecond
		}
		if err := chromedp.Run(ctx, chromedp.Sleep(espera)); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Avançar pra próxima tela
// ---------------------------------------------------------------------------

// avancarTela clica em OK — necessário pra todo tipo de campo, exceto
// seleção única (multiple_choice sem múltipla seleção, e yes_no — que é
// sempre seleção única), que já avança sozinha ao clicar na opção (ver
// preencherMultiplaEscolha). Clicar OK ali de novo não tem efeito (a tela
// já mudou), então simplificamos: só pula o clique nesse caso específico.
func avancarTela(ctx context.Context, tela *telaAtual) error {
	if (tela.Tipo == "multiple_choice" || tela.Tipo == "yes_no") && !tela.MultiSelecao {
		return nil
	}
	sel := fmt.Sprintf(`[%s="1"] [data-qa*="ok-button-visible"]`, marcaAlvo)
	// O sleep depois do clique é necessário: o Typeform anima a saída da
	// pergunta atual e a entrada da próxima (as duas ficam montadas e
	// parcialmente visíveis no meio da transição) — ler o DOM cedo demais
	// pega a pergunta errada.
	return chromedp.Run(ctx,
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.Click(sel, chromedp.ByQuery),
		chromedp.Sleep(700*time.Millisecond),
	)
}
