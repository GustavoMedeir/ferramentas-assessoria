// Aba "Apresentação": exibe a apresentação institucional do escritório a
// partir de um arquivo HTML autocontido (CSS e imagens embutidos) ou PDF
// escolhido pelo usuário. O caminho fica salvo em config.json e o conteúdo é
// lido do disco toda vez que a aba abre ou o usuário clica "Recarregar" —
// trocar a apresentação é só apontar pra outro arquivo, sem recompilar o app.
//
// HTML roda dentro de um <iframe sandbox srcdoc=...>: fica totalmente
// isolado do CSS/JS do app (o tema de um não vaza pro outro), e o srcdoc
// dispensa servir o arquivo — o conteúdo já vem inteiro do backend. Por isso
// a recomendação de HTML autocontido: assets externos (imagens em pastas
// soltas) não carregam num srcdoc; imagens em base64 e CSS inline, sim. Pra
// leitura normal, PDF usa o mesmo esquema (data URI num iframe), mostrando o
// visualizador nativo do WebView2 (zoom/busca/paginação de graça).
//
// Só que o visualizador nativo tem sua própria barra de ferramentas, que não
// dá pra tirar — ruim pra apresentar pro cliente. Por isso "Tela cheia" num
// PDF entra num modo apresentação (slideshow) próprio: cada página do PDF é
// renderizada como imagem (via o motor PDFium que o app já embute, o mesmo
// usado em Rentabilidade) e mostrada em tela cheia, sem chrome nenhum — um
// clique avança o slide. .pptx/.ppsx/.odp não são convertidos
// automaticamente (evita depender de instalar um conversor externo como o
// LibreOffice); o backend devolve uma orientação pra exportar como PDF antes
// (ver mensagemFormatoOfficeNaoSuportado em app.go).

import { state } from "../state.js";
import { el, clear, btn } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import {
    CarregarApresentacao,
    EscolherApresentacao,
    ApresentacaoContarPaginas,
    ApresentacaoRenderizarPagina,
} from "../../wailsjs/go/main/App.js";

let refs = {};
let ctxApp = null;
let dtoAtual = null;

export function mount(container, ctx) {
    clear(container);
    refs = {};
    ctxApp = ctx;

    const wrap = el("div", { class: "apres" });
    refs.wrap = wrap;
    container.appendChild(wrap);

    carregarERenderizar();
}

async function carregarERenderizar() {
    try {
        const dto = await CarregarApresentacao();
        renderizar(dto);
    } catch (e) {
        renderizar({ Caminho: "", Erro: "Erro ao carregar a apresentação: " + e });
    }
}

function renderizar(dto) {
    clear(refs.wrap);
    dtoAtual = dto;

    // Mantém o flag usado pelo Modo apresentação (main.js) em dia dentro da
    // sessão — mesmo critério do boot: há apresentação quando um arquivo foi
    // escolhido (config.Apresentacao != "").
    if (state.prefs) state.prefs.temApresentacao = !!dto.Caminho;

    // O conteúdo pode ter mudado desde a última leitura (usuário editou o
    // PDF e clicou Recarregar) — o cache de páginas do slideshow não pode
    // sobreviver a isso.
    if (dto.Tipo === "pdf") cachePaginas = new Map();

    // Erro (arquivo sumiu, formato não suportado, exceção na chamada, etc.)
    // tem prioridade sobre "sem arquivo": uma falha na chamada de rede/IPC
    // devolve Caminho "" só porque nunca chegou a ler nada, e precisa
    // aparecer como erro, não como "nenhuma apresentação" silencioso.
    if (dto.Erro) {
        refs.wrap.appendChild(estadoErro(dto));
        return;
    }
    // Sem arquivo escolhido ainda.
    if (!dto.Caminho) {
        refs.wrap.appendChild(estadoVazio());
        return;
    }
    // Tudo certo: barra de ações + apresentação no iframe.
    refs.wrap.appendChild(montarBarra());

    const palco = el("div", { class: "apres-palco" });
    const frame = el("iframe", { class: "apres-frame", title: "Apresentação" });
    if (dto.Tipo === "pdf") {
        // Visualizador nativo do WebView2 — não precisa (nem deve) de
        // sandbox: não é HTML/JS de terceiros rodando na página, é o plugin
        // de PDF interno do navegador. Bom pra leitura/revisão normal; a
        // apresentação em si (tela cheia) usa o slideshow próprio abaixo.
        frame.src = "data:application/pdf;base64," + dto.PDFBase64;
    } else {
        // O HTML é do próprio usuário (arquivo que ele escolheu), então
        // liberamos scripts pra apresentações com navegação/animação em JS.
        // Continua num sandbox: sem acesso ao app nem à navegação do topo.
        frame.setAttribute("sandbox", "allow-scripts allow-popups allow-forms allow-modals allow-same-origin");
        frame.srcdoc = dto.HTML;
    }
    refs.frame = frame;
    palco.appendChild(frame);
    refs.wrap.appendChild(palco);
}

function montarBarra() {
    const acoes = el("div", { class: "apres-acoes" }, [
        btn("Tela cheia", { classe: "pri", icon: icons.iconTelaCheia, onClick: abrirTelaCheia }),
        btn("Recarregar", { icon: icons.iconAtualizar, onClick: carregarERenderizar }),
        btn("Trocar arquivo", { icon: icons.iconPasta, onClick: escolherArquivo }),
    ]);

    return el("div", { class: "apres-barra" }, [acoes]);
}

function estadoVazio() {
    const caixa = el("div", { class: "apres-vazio" });
    const icone = el("div", { class: "apres-vazio-icone" });
    icone.insertAdjacentHTML("beforeend", icons.iconApresentacao);
    caixa.appendChild(icone);
    caixa.appendChild(el("h3", { class: "apres-vazio-t", text: "Nenhuma apresentação carregada" }));
    caixa.appendChild(
        el("p", {
            class: "apres-vazio-sub",
            text:
                "Escolha um arquivo HTML autocontido (com o CSS e as imagens embutidos) ou um PDF com a " +
                "apresentação institucional do escritório. Você pode trocá-lo quando quiser — é só apontar para " +
                "outro arquivo, sem precisar mexer no programa. Apresentações em PowerPoint (.pptx/.ppsx) ou " +
                "OpenDocument (.odp) precisam ser exportadas como PDF antes (\"Salvar como\" > PDF).",
        })
    );
    caixa.appendChild(
        el("div", { class: "apres-vazio-acao" }, [
            btn("Escolher apresentação", { classe: "pri", icon: icons.iconPasta, onClick: escolherArquivo }),
        ])
    );
    return caixa;
}

function estadoErro(dto) {
    const caixa = el("div", { class: "apres-vazio" });
    const icone = el("div", { class: "apres-vazio-icone erro" });
    icone.insertAdjacentHTML("beforeend", icons.iconAlerta);
    caixa.appendChild(icone);
    caixa.appendChild(el("h3", { class: "apres-vazio-t", text: "Não foi possível abrir a apresentação" }));
    caixa.appendChild(el("p", { class: "apres-vazio-sub", text: dto.Erro }));
    if (dto.Caminho) {
        caixa.appendChild(el("p", { class: "apres-vazio-cam", text: dto.Caminho, title: dto.Caminho }));
    }
    caixa.appendChild(
        el("div", { class: "apres-vazio-acao" }, [
            btn("Escolher outro arquivo", { classe: "pri", icon: icons.iconPasta, onClick: escolherArquivo }),
            btn("Recarregar", { icon: icons.iconAtualizar, onClick: carregarERenderizar }),
        ])
    );
    return caixa;
}

async function escolherArquivo() {
    try {
        const dto = await EscolherApresentacao();
        if (!dto || !dto.Caminho) return; // usuário cancelou o diálogo
        renderizar(dto);
        ctxApp?.setStatus?.("Apresentação carregada.");
    } catch (e) {
        alert("Não foi possível carregar a apresentação.\n\n" + e);
    }
}

function abrirTelaCheia() {
    if (dtoAtual?.Tipo === "pdf") {
        abrirApresentacaoPDF();
        return;
    }
    if (!refs.frame) return;
    // Coloca o próprio iframe em tela cheia — a apresentação (HTML) decide
    // seu próprio formato de slide/rolagem, ver comentário de topo.
    const alvo = refs.frame;
    const pedir = alvo.requestFullscreen || alvo.webkitRequestFullscreen;
    if (pedir) {
        pedir.call(alvo).catch(() => {});
    }
}

// ---------------------------------------------------------------------------
// Modo apresentação (slideshow) para PDF
// ---------------------------------------------------------------------------
// Cada página do PDF vira uma imagem (renderizada sob demanda pelo backend,
// via PDFium) mostrada em tela cheia sem nenhuma barra de ferramentas — um
// clique avança o slide, setas do teclado andam nos dois sentidos, Esc sai.
// As imagens já renderizadas ficam em cache (por índice de página) só
// durante a sessão de apresentação atual; a cache é descartada sempre que a
// apresentação é recarregada (ver renderizar()), pra nunca mostrar um slide
// desatualizado depois de editar o arquivo.

let slideshowEl = null;
let slideshowImgEl = null;
let slideshowLoadingEl = null;
let cachePaginas = new Map();
let paginaAtual = 0;
let totalPaginasAtual = 0;

function garantirSlideshow() {
    if (slideshowEl) return slideshowEl;

    slideshowImgEl = el("img", { class: "apres-slideshow-img", alt: "Slide", draggable: "false" });
    slideshowLoadingEl = el("div", { class: "apres-slideshow-loading", text: "Carregando..." });
    slideshowLoadingEl.hidden = true;

    slideshowEl = el("div", { class: "apres-slideshow" }, [slideshowImgEl, slideshowLoadingEl]);
    slideshowEl.addEventListener("click", () => proximaPagina());
    // Anexado direto no <body> (não dentro da aba): garante um overlay de
    // tela cheia de verdade, sem disputar z-index/overflow com o layout do
    // app (ver .content { overflow: hidden } em theme.css).
    document.body.appendChild(slideshowEl);
    return slideshowEl;
}

function obterPagina(indice) {
    if (cachePaginas.has(indice)) return cachePaginas.get(indice);
    const promessa = ApresentacaoRenderizarPagina(dtoAtual.Caminho, indice).then((b64) => "data:image/png;base64," + b64);
    cachePaginas.set(indice, promessa);
    return promessa;
}

function prefetchPagina(indice) {
    if (indice < 0 || indice >= totalPaginasAtual || cachePaginas.has(indice)) return;
    obterPagina(indice).catch(() => cachePaginas.delete(indice));
}

async function mostrarPagina(indice) {
    paginaAtual = indice;
    slideshowLoadingEl.hidden = false;
    try {
        slideshowImgEl.src = await obterPagina(indice);
    } catch (e) {
        slideshowImgEl.removeAttribute("src");
    } finally {
        slideshowLoadingEl.hidden = true;
    }
    prefetchPagina(indice + 1);
}

function proximaPagina() {
    if (paginaAtual + 1 < totalPaginasAtual) mostrarPagina(paginaAtual + 1);
}

function paginaAnterior() {
    if (paginaAtual - 1 >= 0) mostrarPagina(paginaAtual - 1);
}

function aoTeclaApresentacao(e) {
    if (e.key === "ArrowRight" || e.key === "PageDown" || e.key === " ") {
        e.preventDefault();
        proximaPagina();
    } else if (e.key === "ArrowLeft" || e.key === "PageUp") {
        e.preventDefault();
        paginaAnterior();
    } else if (e.key === "Escape") {
        if (document.fullscreenElement === slideshowEl) {
            document.exitFullscreen?.().catch(() => {});
        } else {
            fecharApresentacaoPDF();
        }
    }
}

function aoMudarFullscreen() {
    // Sai do modo apresentação sempre que o navegador deixa de estar em
    // tela cheia nesse elemento — cobre tanto o Esc nativo quanto qualquer
    // outro jeito do usuário sair (ex.: botão da própria janela do SO).
    if (document.fullscreenElement !== slideshowEl) {
        fecharApresentacaoPDF();
    }
}

function fecharApresentacaoPDF() {
    if (!slideshowEl) return;
    slideshowEl.classList.remove("ativo");
    document.removeEventListener("keydown", aoTeclaApresentacao);
    document.removeEventListener("fullscreenchange", aoMudarFullscreen);
}

async function abrirApresentacaoPDF() {
    if (!dtoAtual || dtoAtual.Tipo !== "pdf") return;
    const cont = garantirSlideshow();

    cont.classList.add("ativo");
    document.addEventListener("keydown", aoTeclaApresentacao);
    document.addEventListener("fullscreenchange", aoMudarFullscreen);

    const pedir = cont.requestFullscreen || cont.webkitRequestFullscreen;
    if (pedir) pedir.call(cont).catch(() => {});

    try {
        totalPaginasAtual = await ApresentacaoContarPaginas(dtoAtual.Caminho);
    } catch (e) {
        totalPaginasAtual = 1; // segue mostrando ao menos a 1ª página, melhor que travar
    }
    await mostrarPagina(0);
}
