import {
    EscolherPDFParaEditar,
    SalvarPDFEditado,
    ApresentacaoContarPaginas,
    ApresentacaoRenderizarPagina,
    ObterAssinaturaAtiva,
} from "../../wailsjs/go/main/App.js";
import { el, clear, btn } from "../ui/components.js";
import * as icons from "../ui/icons.js";

// Precisa bater com dpiApresentacaoSlide em app.go — é a resolução fixa em
// que ApresentacaoRenderizarPagina desenha cada página.
const DPI_RENDER = 200;

const ZOOM_MIN = 0.2;
const ZOOM_MAX = 4;

let refs = {};
let caminhoAtual = null;
let totalPaginas = 0;
let indiceAtual = 0;
let paginas = []; // por índice: { imagemBase, dataUrl, larguraPx, alturaPx, larguraPt, alturaPt, undo: [] }

let ferramenta = "caneta"; // "caneta" | "texto" | "tarja" | "imagem" | "x"
let cor = "#000000";
let espessura = 4;
let desenhando = false;
let pontoInicial = null;
let ultimoPonto = null;
let imagemPendente = null;
let zoomAtual = 1;
let zoomAjustado = false; // já rodou o fit-to-width nesse PDF? (só a 1ª página deve auto-ajustar)
let flutuante = null; // objeto de texto/imagem ainda não achatado no canvas (ver criarFlutuante*)
let handlerMouseUpGlobal = null;
let handlerKeydownGlobal = null;
let containerAtual = null;

export function mount(container) {
    // clear() esvazia o container e reseta o scroll pro topo (mesmo problema
    // documentado em tabs/emails.js:renderBlocos) — preserva a posição pra
    // não "puxar pra cima" a cada remontagem (ex.: Confirmar em Configurações).
    const scrollAnterior = container.scrollTop;
    clear(container);
    refs = {};
    containerAtual = container;
    caminhoAtual = null;
    totalPaginas = 0;
    indiceAtual = 0;
    paginas = [];
    ferramenta = "caneta";
    desenhando = false;
    imagemPendente = null;
    flutuante = null;
    zoomAtual = 1;
    zoomAjustado = false;

    refs.vazio = el("div", { class: "estado-vazio" }, [
        el("div", { text: "Nenhum PDF aberto." }),
        btn("Abrir PDF", { classe: "pri", icon: icons.iconPasta, onClick: abrirPDF }),
    ]);
    container.appendChild(refs.vazio);

    refs.editor = el("div", { class: "pdfedit" });
    refs.editor.style.display = "none";
    montarToolbar();
    refs.editor.appendChild(refs.toolbar);

    refs.canvasWrap = el("div", { class: "pdfedit-canvas-wrap" });
    refs.canvas = el("canvas", { class: "pdfedit-canvas" });
    refs.canvasSombra = el("canvas");
    refs.canvasSombra.style.display = "none";
    refs.canvas.addEventListener("mousedown", aoMouseDown);
    refs.canvas.addEventListener("mousemove", aoMouseMove);
    refs.canvasWrap.addEventListener("wheel", aoRodaMouse, { passive: false });
    refs.canvasWrap.appendChild(refs.canvas);
    refs.canvasWrap.appendChild(refs.canvasSombra);
    refs.editor.appendChild(refs.canvasWrap);

    refs.status = el("p", { class: "cfg-sub pdfedit-status" });
    refs.editor.appendChild(refs.status);

    container.appendChild(refs.editor);
    container.scrollTop = scrollAnterior;

    // Um mousedown na página pode arrastar e soltar o botão fora do
    // <canvas> — um listener só no canvas perderia esse mouseup e deixaria
    // "desenhando" travado em true. Ouve no window; remove o listener
    // anterior antes de registrar de novo (mount() roda de novo se o
    // usuário sair e voltar pra aba).
    if (handlerMouseUpGlobal) window.removeEventListener("mouseup", handlerMouseUpGlobal);
    handlerMouseUpGlobal = aoMouseUp;
    window.addEventListener("mouseup", handlerMouseUpGlobal);

    // Ctrl+Z desfaz — só quando essa aba está de fato visível (senão o
    // atalho dispararia de qualquer outra tela). offsetParent é null
    // sempre que o elemento (ou um ancestral) está com display:none, então
    // reflete a aba ativa sem precisar main.js avisar essa função.
    if (handlerKeydownGlobal) window.removeEventListener("keydown", handlerKeydownGlobal);
    handlerKeydownGlobal = (e) => {
        if (!(e.ctrlKey || e.metaKey) || e.key.toLowerCase() !== "z") return;
        if (!containerAtual || containerAtual.offsetParent === null) return;
        e.preventDefault();
        desfazer();
    };
    window.addEventListener("keydown", handlerKeydownGlobal);
}

function montarToolbar() {
    refs.toolbar = el("div", { class: "pdfedit-toolbar" });

    const grupoFerramentas = el("div", { class: "pdfedit-grupo" });
    const definicoesFerramentas = [
        { id: "caneta", label: "Caneta", icon: icons.iconCaneta },
        { id: "texto", label: "Texto", icon: icons.iconTexto },
        { id: "tarja", label: "Tarja branca", icon: icons.iconTarja },
        { id: "imagem", label: "Imagem", icon: icons.iconImagem },
        { id: "assinatura", label: "Assinatura", icon: icons.iconAssinatura },
        { id: "x", label: "Marca X", icon: icons.iconMarcaX },
    ];
    refs.botoesFerramenta = {};
    for (const def of definicoesFerramentas) {
        const botao = el("button", { class: "pdfedit-ferr-botao", title: def.label, type: "button" });
        botao.insertAdjacentHTML("beforeend", def.icon);
        botao.addEventListener("click", () => {
            finalizarFlutuante();
            ferramenta = def.id;
            atualizarBotoesFerramenta();
            if (def.id === "imagem") escolherImagem();
            if (def.id === "assinatura") usarAssinaturaAtiva();
        });
        refs.botoesFerramenta[def.id] = botao;
        grupoFerramentas.appendChild(botao);
    }
    refs.toolbar.appendChild(grupoFerramentas);

    refs.inputCor = el("input", { type: "color", class: "pdfedit-cor", title: "Cor" });
    refs.inputCor.value = cor;
    refs.inputCor.addEventListener("input", () => {
        cor = refs.inputCor.value;
        if (flutuante && flutuante.tipo === "texto") flutuante.area.style.color = cor;
    });
    refs.toolbar.appendChild(refs.inputCor);

    refs.inputEspessura = el("input", { type: "range", min: "1", max: "20", class: "pdfedit-espessura", title: "Espessura" });
    refs.inputEspessura.value = String(espessura);
    refs.inputEspessura.addEventListener("input", () => {
        espessura = Number(refs.inputEspessura.value);
        if (flutuante && flutuante.tipo === "texto") flutuante.area.style.fontSize = tamanhoFontePorEspessura() + "px";
    });
    refs.toolbar.appendChild(refs.inputEspessura);

    refs.toolbar.appendChild(btn("Desfazer", { icon: icons.iconDesfazer, onClick: desfazer }));

    refs.navPaginas = el("div", { class: "pdfedit-nav" });
    refs.toolbar.appendChild(refs.navPaginas);

    refs.toolbar.appendChild(el("div", { class: "pdfedit-espaco" }));

    refs.toolbar.appendChild(btn("Trocar PDF", { onClick: abrirPDF }));
    refs.btnSalvar = btn("Salvar como...", { classe: "pri", icon: icons.iconSalvar, onClick: salvar });
    refs.toolbar.appendChild(refs.btnSalvar);

    atualizarBotoesFerramenta();
}

function atualizarBotoesFerramenta() {
    for (const [id, botao] of Object.entries(refs.botoesFerramenta)) {
        botao.classList.toggle("on", id === ferramenta);
    }
}

function atualizarNavegacaoPaginas() {
    clear(refs.navPaginas);
    if (totalPaginas <= 1) return;
    const btnAnt = el("button", { class: "btn ghost pdfedit-nav-seta", text: "‹", type: "button" });
    btnAnt.disabled = indiceAtual === 0;
    btnAnt.addEventListener("click", () => irParaPagina(indiceAtual - 1));
    const rotulo = el("span", { class: "pdfedit-nav-label", text: `Página ${indiceAtual + 1} de ${totalPaginas}` });
    const btnProx = el("button", { class: "btn ghost pdfedit-nav-seta", text: "›", type: "button" });
    btnProx.disabled = indiceAtual === totalPaginas - 1;
    btnProx.addEventListener("click", () => irParaPagina(indiceAtual + 1));
    refs.navPaginas.appendChild(btnAnt);
    refs.navPaginas.appendChild(rotulo);
    refs.navPaginas.appendChild(btnProx);
}

function carregarImagem(src) {
    return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => resolve(img);
        img.onerror = () => reject(new Error("não foi possível carregar a imagem"));
        img.src = src;
    });
}

async function abrirPDF() {
    let caminho;
    try {
        caminho = await EscolherPDFParaEditar();
    } catch (e) {
        refs.status.textContent = "Erro ao abrir arquivo: " + e;
        return;
    }
    if (!caminho) return; // usuário cancelou o diálogo

    finalizarFlutuante();
    refs.status.textContent = "Carregando PDF...";
    try {
        totalPaginas = await ApresentacaoContarPaginas(caminho);
        caminhoAtual = caminho;
        paginas = new Array(totalPaginas).fill(null);
        indiceAtual = 0;
        await irParaPagina(0);
        refs.vazio.style.display = "none";
        refs.editor.style.display = "flex";
        refs.status.textContent = "";
    } catch (e) {
        refs.status.textContent = "Erro ao abrir PDF: " + e;
        caminhoAtual = null;
    }
}

async function carregarPaginaSeNecessario(indice) {
    if (paginas[indice]) return;
    const png = await ApresentacaoRenderizarPagina(caminhoAtual, indice);
    const imagemBase = await carregarImagem("data:image/png;base64," + png);
    paginas[indice] = {
        imagemBase,
        dataUrl: null, // só existe depois da 1ª edição confirmada nessa página
        larguraPx: imagemBase.naturalWidth,
        alturaPx: imagemBase.naturalHeight,
        larguraPt: (imagemBase.naturalWidth * 72) / DPI_RENDER,
        alturaPt: (imagemBase.naturalHeight * 72) / DPI_RENDER,
        undo: [],
    };
}

async function irParaPagina(indice) {
    finalizarFlutuante();
    refs.status.textContent = "Carregando página...";
    try {
        await carregarPaginaSeNecessario(indice);
    } catch (e) {
        refs.status.textContent = "Erro ao carregar página: " + e;
        return;
    }
    indiceAtual = indice;
    const p = paginas[indice];
    refs.canvas.width = p.larguraPx;
    refs.canvas.height = p.alturaPx;
    refs.canvasSombra.width = p.larguraPx;
    refs.canvasSombra.height = p.alturaPx;

    const ctx = refs.canvas.getContext("2d");
    const imagemAtual = p.dataUrl ? await carregarImagem(p.dataUrl) : p.imagemBase;
    ctx.clearRect(0, 0, refs.canvas.width, refs.canvas.height);
    ctx.drawImage(imagemAtual, 0, 0);

    // Zoom inicial: ajusta a página pra caber na largura visível do wrap
    // (a página renderizada a 200 DPI costuma ser bem maior que isso) — só
    // na primeira página que o usuário abre desse PDF. Trocar de página
    // depois preserva o zoom que o usuário já tiver ajustado (roda do
    // mouse, aoRodaMouse).
    if (!zoomAjustado) {
        const larguraDisponivel = refs.canvasWrap.clientWidth - 32;
        zoomAtual = larguraDisponivel > 0 ? Math.min(1, larguraDisponivel / p.larguraPx) : 1;
        zoomAjustado = true;
    }
    aplicarZoom();

    atualizarNavegacaoPaginas();
    refs.status.textContent = "";
}

function aplicarZoom() {
    const p = paginas[indiceAtual];
    if (!p) return;
    refs.canvas.style.width = p.larguraPx * zoomAtual + "px";
    refs.canvas.style.height = "auto";
}

// aoRodaMouse restringe o zoom (Ctrl+roda) à área do PDF — sem isso, o
// WebView2 trata Ctrl+roda como zoom nativo da janela inteira. Rolagem sem
// Ctrl não é interceptada: continua sendo o scroll normal do wrap.
function aoRodaMouse(e) {
    if (!e.ctrlKey) return;
    if (!paginas[indiceAtual]) return;
    e.preventDefault();
    const fator = e.deltaY < 0 ? 1.1 : 1 / 1.1;
    zoomAtual = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, zoomAtual * fator));
    aplicarZoom();
}

function commitarEdicao() {
    paginas[indiceAtual].dataUrl = refs.canvas.toDataURL("image/png");
}

function salvarUndo() {
    const p = paginas[indiceAtual];
    p.undo.push(refs.canvas.toDataURL("image/png"));
    if (p.undo.length > 20) p.undo.shift(); // limita memória: só as últimas 20 ações por página
}

async function desfazer() {
    finalizarFlutuante();
    const p = paginas[indiceAtual];
    if (!p || !p.undo.length) return;
    const anterior = p.undo.pop();
    const img = await carregarImagem(anterior);
    const ctx = refs.canvas.getContext("2d");
    ctx.clearRect(0, 0, refs.canvas.width, refs.canvas.height);
    ctx.drawImage(img, 0, 0);
    p.dataUrl = anterior;
}

function coordenadasCanvas(e) {
    const escala = refs.canvas.width / refs.canvas.clientWidth;
    return { x: e.offsetX * escala, y: e.offsetY * escala };
}

function aoMouseDown(e) {
    if (!paginas[indiceAtual]) return;

    if (flutuante) {
        finalizarFlutuante();
        return; // esse clique só confirma o flutuante aberto; não inicia outra ação
    }

    const { x, y } = coordenadasCanvas(e);

    if (ferramenta === "texto") {
        criarFlutuanteTexto(x, y);
        return;
    }
    if (ferramenta === "imagem" || ferramenta === "assinatura") {
        if (imagemPendente) criarFlutuanteImagem(x, y);
        return;
    }
    if (ferramenta === "x") {
        desenharMarcaX(x, y);
        return;
    }

    salvarUndo();
    desenhando = true;
    pontoInicial = { x, y };
    ultimoPonto = { x, y };
    if (ferramenta === "tarja") {
        const ctxSombra = refs.canvasSombra.getContext("2d");
        ctxSombra.clearRect(0, 0, refs.canvasSombra.width, refs.canvasSombra.height);
        ctxSombra.drawImage(refs.canvas, 0, 0);
    }
}

function aoMouseMove(e) {
    if (!desenhando) return;
    const { x, y } = coordenadasCanvas(e);
    const ctx = refs.canvas.getContext("2d");

    if (ferramenta === "caneta") {
        ctx.strokeStyle = cor;
        ctx.lineWidth = espessura;
        ctx.lineCap = "round";
        ctx.lineJoin = "round";
        ctx.beginPath();
        ctx.moveTo(ultimoPonto.x, ultimoPonto.y);
        ctx.lineTo(x, y);
        ctx.stroke();
        ultimoPonto = { x, y };
    } else if (ferramenta === "tarja") {
        ctx.clearRect(0, 0, refs.canvas.width, refs.canvas.height);
        ctx.drawImage(refs.canvasSombra, 0, 0);
        ctx.fillStyle = "#ffffff";
        const rx = Math.min(pontoInicial.x, x);
        const ry = Math.min(pontoInicial.y, y);
        ctx.fillRect(rx, ry, Math.abs(x - pontoInicial.x), Math.abs(y - pontoInicial.y));
    }
}

function aoMouseUp() {
    if (!desenhando) return;
    desenhando = false;
    commitarEdicao();
}

// escalaTelaAtual devolve quantos pixels de tela correspondem a 1 pixel do
// canvas no zoom atual — usado pra converter tamanhos/posições entre o
// espaço de desenho (canvas) e o espaço visual (elementos flutuantes).
function escalaTelaAtual() {
    return refs.canvas.clientWidth / refs.canvas.width;
}

// posicionarFlutuante converte uma posição em pixels do canvas (espaço de
// desenho) pra pixels de tela dentro de canvasWrap, considerando o zoom
// atual — canvas.offsetLeft/Top já são relativos ao wrap (que é o
// offsetParent, position:relative).
function posicionarFlutuante(elemento, xCanvas, yCanvas) {
    const escalaTela = escalaTelaAtual();
    elemento.style.left = refs.canvas.offsetLeft + xCanvas * escalaTela + "px";
    elemento.style.top = refs.canvas.offsetTop + yCanvas * escalaTela + "px";
}

// ativarArrasto liga o arrasto de um "pegador" (a alça) a uma função de
// callback que recebe o delta em pixels do CANVAS (já convertido do delta
// de tela pelo zoom atual) — funciona em qualquer nível de zoom. Devolve
// uma função de limpeza, chamada quando o flutuante é finalizado.
function ativarArrasto(pegador, aoMover) {
    let arrastando = false;
    let ultimoX = 0;
    let ultimoY = 0;

    function aoDescer(e) {
        e.preventDefault();
        e.stopPropagation();
        arrastando = true;
        ultimoX = e.clientX;
        ultimoY = e.clientY;
    }
    function aoMoverPonteiro(e) {
        if (!arrastando) return;
        const escalaTela = escalaTelaAtual();
        const dx = (e.clientX - ultimoX) / escalaTela;
        const dy = (e.clientY - ultimoY) / escalaTela;
        ultimoX = e.clientX;
        ultimoY = e.clientY;
        aoMover(dx, dy);
    }
    function aoSoltar() {
        arrastando = false;
    }

    pegador.addEventListener("mousedown", aoDescer);
    window.addEventListener("mousemove", aoMoverPonteiro);
    window.addEventListener("mouseup", aoSoltar);
    return () => {
        window.removeEventListener("mousemove", aoMoverPonteiro);
        window.removeEventListener("mouseup", aoSoltar);
    };
}

// tamanhoFontePorEspessura deriva o tamanho da fonte do texto a partir do
// mesmo controle "Espessura" da toolbar (continua editável por ali) — só
// desloca a base pra que o valor padrão da espessura (4) resulte em 24px,
// em vez dos 16px que a fórmula antiga (espessura * 4) dava.
function tamanhoFontePorEspessura() {
    return Math.max(14, 8 + espessura * 4);
}

// criarFlutuanteTexto cria um quadro de texto editável (multi-linha),
// arrastável, sobre a página — só vira conteúdo real do PDF quando
// confirmado (ver finalizarFlutuante). Só um flutuante existe por vez.
function criarFlutuanteTexto(x, y) {
    finalizarFlutuante();

    const tamanhoFonte = tamanhoFontePorEspessura();
    const caixa = el("div", { class: "pdfedit-flutuante" });
    caixa.addEventListener("mousedown", (e) => e.stopPropagation());

    const alca = el("div", { class: "pdfedit-flutuante-alca", title: "Arraste para mover" });
    const area = el("textarea", { class: "pdfedit-flutuante-input" });
    area.style.color = cor;
    area.style.fontSize = tamanhoFonte + "px";
    const acoes = montarAcoesFlutuante();

    caixa.appendChild(alca);
    caixa.appendChild(area);
    caixa.appendChild(acoes);
    refs.canvasWrap.appendChild(caixa);
    posicionarFlutuante(caixa, x, y);

    // A alça de arrastar fica ACIMA do textarea dentro da caixa — sem
    // contar essa altura, o texto final desenha na posição da alça, não na
    // posição em que o textarea de fato aparece na tela (sobe ao
    // confirmar). alca.offsetHeight já reflete o tamanho renderizado
    // (em pixels de tela); convertido pra pixels do canvas antes de guardar.
    const deslocamentoY = alca.offsetHeight / escalaTelaAtual();

    flutuante = { tipo: "texto", elemento: caixa, area, x, y, deslocamentoY, tamanhoFonte, cor };
    flutuante.pararArrasto = ativarArrasto(alca, (dx, dy) => {
        flutuante.x += dx;
        flutuante.y += dy;
        posicionarFlutuante(caixa, flutuante.x, flutuante.y);
    });

    area.addEventListener("keydown", (e) => {
        if (e.key === "Escape") cancelarFlutuante();
        e.stopPropagation();
    });
    area.focus();
}

// criarFlutuanteImagem mesma ideia do texto, com a prévia da imagem
// escolhida (ver escolherImagem) no lugar do textarea.
function criarFlutuanteImagem(x, y) {
    finalizarFlutuante();

    const larguraAlvo = Math.min(imagemPendente.naturalWidth, refs.canvas.width * 0.3);
    const alturaAlvo = larguraAlvo * (imagemPendente.naturalHeight / imagemPendente.naturalWidth);
    const xInicial = x - larguraAlvo / 2;
    const yInicial = y - alturaAlvo / 2;

    const caixa = el("div", { class: "pdfedit-flutuante" });
    caixa.addEventListener("mousedown", (e) => e.stopPropagation());

    const alca = el("div", { class: "pdfedit-flutuante-alca", title: "Arraste para mover" });
    const preview = el("img", { class: "pdfedit-flutuante-imagem" });
    preview.src = imagemPendente.src;
    const acoes = montarAcoesFlutuante();

    caixa.appendChild(alca);
    caixa.appendChild(preview);
    caixa.appendChild(acoes);
    refs.canvasWrap.appendChild(caixa);
    posicionarFlutuante(caixa, xInicial, yInicial);
    aplicarEscalaFlutuante(caixa, xInicial, yInicial, larguraAlvo, alturaAlvo);

    // Mesmo ajuste do texto: a alça fica acima da prévia da imagem dentro
    // da caixa, então o desenho final precisa descontar essa altura pra não
    // "subir" em relação a onde a prévia realmente aparecia na tela.
    const deslocamentoY = alca.offsetHeight / escalaTelaAtual();

    flutuante = {
        tipo: "imagem",
        elemento: caixa,
        imagem: imagemPendente,
        x: xInicial,
        y: yInicial,
        deslocamentoY,
        largura: larguraAlvo,
        altura: alturaAlvo,
    };
    imagemPendente = null;
    refs.status.textContent = "Arraste pra reposicionar e clique em ✓ pra aplicar.";

    flutuante.pararArrasto = ativarArrasto(alca, (dx, dy) => {
        flutuante.x += dx;
        flutuante.y += dy;
        posicionarFlutuante(caixa, flutuante.x, flutuante.y);
    });
}

// aplicarEscalaFlutuante redimensiona a prévia (em tela) de acordo com o
// zoom atual — a largura/altura guardadas em flutuante são sempre em
// pixels do canvas, não de tela.
function aplicarEscalaFlutuante(caixa, xCanvas, yCanvas, larguraPx, alturaPx) {
    const escalaTela = escalaTelaAtual();
    const preview = caixa.querySelector(".pdfedit-flutuante-imagem");
    if (preview) {
        preview.style.width = larguraPx * escalaTela + "px";
        preview.style.height = alturaPx * escalaTela + "px";
    }
}

// TAMANHO_MARCA_X é o lado (em pixels do canvas) da caixa onde a marca "X"
// é desenhada — fixo, não depende do zoom nem da espessura escolhida.
const TAMANHO_MARCA_X = 10;

// desenharMarcaX aplica a marca "✕" direto no clique — ao contrário de
// texto/imagem, não fica flutuante/arrastável: sempre preta, sempre
// 10×10px, sempre centrada exatamente no ponto clicado.
function desenharMarcaX(x, y) {
    salvarUndo();
    const meio = TAMANHO_MARCA_X / 2;
    const ctx = refs.canvas.getContext("2d");
    ctx.strokeStyle = "#000000";
    ctx.lineWidth = 1.5;
    ctx.lineCap = "round";
    ctx.beginPath();
    ctx.moveTo(x - meio, y - meio);
    ctx.lineTo(x + meio, y + meio);
    ctx.moveTo(x + meio, y - meio);
    ctx.lineTo(x - meio, y + meio);
    ctx.stroke();
    commitarEdicao();
}

function montarAcoesFlutuante() {
    const acoes = el("div", { class: "pdfedit-flutuante-acoes" });
    const botaoOk = el("button", { class: "pdfedit-flutuante-botao pdfedit-flutuante-ok", type: "button", text: "✓", title: "Aplicar" });
    const botaoCancelar = el("button", {
        class: "pdfedit-flutuante-botao pdfedit-flutuante-cancelar",
        type: "button",
        text: "✕",
        title: "Cancelar",
    });
    botaoOk.addEventListener("mousedown", (e) => e.stopPropagation());
    botaoCancelar.addEventListener("mousedown", (e) => e.stopPropagation());
    botaoOk.addEventListener("click", () => finalizarFlutuante());
    botaoCancelar.addEventListener("click", () => cancelarFlutuante());
    acoes.appendChild(botaoOk);
    acoes.appendChild(botaoCancelar);
    return acoes;
}

// finalizarFlutuante achata o flutuante atual (se houver) na posição em
// que ele está agora e empilha uma única entrada de undo pra toda a
// operação (criar + arrastar conta como uma ação só).
function finalizarFlutuante() {
    if (!flutuante) return;
    const atual = flutuante;
    flutuante = null;
    if (atual.pararArrasto) atual.pararArrasto();

    if (atual.tipo === "texto") {
        const texto = atual.area.value.trim();
        atual.elemento.remove();
        if (!texto) return;
        salvarUndo();
        const ctx = refs.canvas.getContext("2d");
        ctx.fillStyle = atual.cor;
        ctx.font = `${atual.tamanhoFonte}px sans-serif`;
        ctx.textBaseline = "top";
        const yTexto = atual.y + atual.deslocamentoY;
        texto.split("\n").forEach((linha, i) => ctx.fillText(linha, atual.x, yTexto + i * atual.tamanhoFonte * 1.2));
        commitarEdicao();
    } else if (atual.tipo === "imagem") {
        atual.elemento.remove();
        salvarUndo();
        const ctx = refs.canvas.getContext("2d");
        ctx.drawImage(atual.imagem, atual.x, atual.y + atual.deslocamentoY, atual.largura, atual.altura);
        commitarEdicao();
    }
    refs.status.textContent = "";
}

// cancelarFlutuante descarta o flutuante atual sem desenhar nada.
function cancelarFlutuante() {
    if (!flutuante) return;
    if (flutuante.pararArrasto) flutuante.pararArrasto();
    flutuante.elemento.remove();
    flutuante = null;
    refs.status.textContent = "";
}

// usarAssinaturaAtiva carrega direto a assinatura marcada como ativa em
// Configurações → Assinatura (botão dedicado na toolbar, separado da
// ferramenta genérica "Imagem"). Se ainda não houver nenhuma cadastrada,
// só avisa — não abre o seletor de arquivo (isso é papel da ferramenta
// Imagem, não dessa).
async function usarAssinaturaAtiva() {
    let assinatura;
    try {
        assinatura = await ObterAssinaturaAtiva();
    } catch (e) {
        refs.status.textContent = "Erro ao carregar assinatura: " + e;
        return;
    }
    if (!assinatura) {
        refs.status.textContent = "Nenhuma assinatura cadastrada — adicione uma em Configurações → Assinatura.";
        return;
    }
    try {
        imagemPendente = await carregarImagem("data:image;base64," + assinatura.Base64);
        refs.status.textContent = "Assinatura carregada — clique na página para posicionar.";
    } catch (e) {
        refs.status.textContent = "Erro ao carregar assinatura: " + e;
    }
}

function escolherImagem() {
    const input = el("input", { type: "file", accept: "image/*" });
    input.addEventListener("change", () => {
        const arquivo = input.files[0];
        if (!arquivo) return;
        const leitor = new FileReader();
        leitor.onload = () => {
            carregarImagem(leitor.result)
                .then((img) => {
                    imagemPendente = img;
                    refs.status.textContent = "Clique na página para posicionar a imagem.";
                })
                .catch((e) => {
                    refs.status.textContent = "Erro ao carregar imagem: " + e;
                });
        };
        leitor.readAsDataURL(arquivo);
    });
    input.click();
}

// paraJPEGBase64 achata a página (imagem base + edições) num JPEG — o
// backend espera JPEG (ver internal/pdfedit), o PNG só é usado pra manter
// qualidade máxima durante a edição em tela.
async function paraJPEGBase64(p) {
    const imagem = p.dataUrl ? await carregarImagem(p.dataUrl) : p.imagemBase;
    const c = document.createElement("canvas");
    c.width = p.larguraPx;
    c.height = p.alturaPx;
    const ctx = c.getContext("2d");
    ctx.fillStyle = "#ffffff"; // JPEG não tem canal alfa
    ctx.fillRect(0, 0, c.width, c.height);
    ctx.drawImage(imagem, 0, 0);
    return c.toDataURL("image/jpeg", 0.92).split(",")[1];
}

async function salvar() {
    if (!caminhoAtual) return;
    finalizarFlutuante();
    refs.btnSalvar.disabled = true;
    refs.status.textContent = "Salvando...";
    try {
        for (let i = 0; i < totalPaginas; i++) await carregarPaginaSeNecessario(i);

        const paginasDTO = [];
        for (let i = 0; i < totalPaginas; i++) {
            const p = paginas[i];
            paginasDTO.push({
                JPEGBase64: await paraJPEGBase64(p),
                LarguraPt: p.larguraPt,
                AlturaPt: p.alturaPt,
            });
        }

        const caminhoSalvo = await SalvarPDFEditado(paginasDTO, caminhoAtual);
        refs.status.textContent = caminhoSalvo ? `Salvo em: ${caminhoSalvo}` : "";
    } catch (e) {
        refs.status.textContent = "Erro ao salvar: " + e;
    } finally {
        if (refs.btnSalvar) refs.btnSalvar.disabled = false;
    }
}
