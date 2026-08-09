// Helpers de DOM reutilizados pelas seções da UI.

import * as icons from "./icons.js";

export function el(tag, attrs = {}, children = []) {
    const node = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs)) {
        if (k === "class") node.className = v;
        else if (k === "text") node.textContent = v;
        else if (k.startsWith("on") && typeof v === "function") node.addEventListener(k.slice(2).toLowerCase(), v);
        else if (v !== undefined && v !== null) node.setAttribute(k, v);
    }
    for (const child of [].concat(children)) {
        if (child === null || child === undefined) continue;
        node.appendChild(typeof child === "string" ? document.createTextNode(child) : child);
    }
    return node;
}

export function clear(node) {
    while (node.firstChild) node.removeChild(node.firstChild);
}

// Botão padrão da UI (classe .btn). `classe` aceita "pri" | "ghost" |
// "danger" (combináveis com espaço), `icon` é markup SVG (ver ui/icons.js).
export function btn(texto, { classe = "", icon = null, onClick = null, disabled = false } = {}) {
    const button = document.createElement("button");
    button.className = `btn ${classe}`.trim();
    if (icon) button.insertAdjacentHTML("beforeend", icon);
    if (texto) button.appendChild(document.createTextNode(texto));
    if (disabled) button.disabled = true;
    if (onClick) button.addEventListener("click", onClick);
    return button;
}

// Barra com o botão "Limpar" usado no topo das telas de calculadora
// (Calculadora, Comparadora, Previdenciária, Compromissada, Aposentadoria).
// Zera todas as chaves de `bags` (um ou mais objetos de state.js) e chama
// `remontar()` — cada aba reaproveita seu próprio mount() como remontagem,
// que já sabe reaplicar os padrões (ex.: data de hoje) quando o bag some.
export function montarToolbarLimpar(bags, remontar) {
    const botao = el("button", { class: "btn-limpar", type: "button" });
    botao.insertAdjacentHTML("beforeend", icons.iconBorracha);
    botao.appendChild(document.createTextNode("Limpar"));
    botao.addEventListener("click", () => {
        for (const bag of [].concat(bags)) {
            for (const chave of Object.keys(bag)) delete bag[chave];
        }
        remontar();
    });
    return el("div", { class: "calc-toolbar" }, [botao]);
}

// Selo clicável tipo "_Rent" na legenda do modelo (classe .ph).
export function phButton(texto, onClick, { wide = false } = {}) {
    const button = el("button", { class: `ph${wide ? " wide" : ""}`, text: texto, type: "button" });
    if (onClick) button.addEventListener("click", onClick);
    return button;
}

// ---------------------------------------------------------------------------
// Destaque dos placeholders na prévia da mensagem — única lógica duplicada
// entre Go e JS do projeto (pequena e estável, 9 placeholders financeiros).
// Fonte da verdade da lista/ordem: internal/rentabilidade/rentabilidade.go
// (var placeholders). Os valores já vêm formatados em pt-BR pelo backend
// (RegistroDTO.RentFmt/RentAFmt/...), então o JS só precisa fatiar o texto.
// ---------------------------------------------------------------------------

const PLACEHOLDER_ORDEM = [
    "_RentA", "_Rent12M", "_Rent",
    "_PercA", "_Perc12M", "_Perc",
    "_CDIA", "_CDI12M", "_CDI",
    "_Nome",
];
const PLACEHOLDER_RE = new RegExp(PLACEHOLDER_ORDEM.map((p) => p.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("|"), "g");

// _Nome não vem pré-formatado no RegistroDTO (o nome pode ser carregado
// numa base de clientes depois dos registros já existirem, e não queremos
// que fique desatualizado) — por isso é passado à parte, resolvido na hora
// pelo chamador a partir de state.clientDB[registro.Codigo].
function valoresPlaceholder(registro, nome) {
    return {
        _Rent: registro.RentFmt,
        _RentA: registro.RentAFmt,
        _Rent12M: registro.Rent12MFmt,
        _Perc: registro.PercFmt,
        _PercA: registro.PercAFmt,
        _Perc12M: registro.Perc12MFmt,
        _CDI: registro.CDIFmt,
        _CDIA: registro.CDIAFmt,
        _CDI12M: registro.CDI12MFmt,
        _Nome: nome || "[Nome do cliente]",
    };
}

// Renderiza o template dentro de `container`, substituindo cada placeholder
// por um <span class="pill-val"> com o valor formatado. Se `registro` for
// null, renderiza o texto puro (sem destaque).
export function renderMensagemComPills(container, template, registro, nome) {
    clear(container);
    if (!registro) {
        container.textContent = template;
        return;
    }
    const valores = valoresPlaceholder(registro, nome);
    let posicao = 0;
    PLACEHOLDER_RE.lastIndex = 0;
    let match;
    while ((match = PLACEHOLDER_RE.exec(template)) !== null) {
        if (match.index > posicao) {
            container.appendChild(document.createTextNode(template.slice(posicao, match.index)));
        }
        container.appendChild(el("span", { class: "pill-val", text: valores[match[0]] }));
        posicao = match.index + match[0].length;
    }
    if (posicao < template.length) {
        container.appendChild(document.createTextNode(template.slice(posicao)));
    }
}

const PLACEHOLDER_NOME_RE = /_Nome/g;

// Versão do destaque de placeholders pro Modo Festas (aba Rentabilidade):
// só o "_Nome" existe nesse modelo, sem depender de um Registro processado
// (ao contrário de renderMensagemComPills, que precisa dos valores
// financeiros do PDF).
export function renderMensagemFestasComPills(container, template, nome) {
    clear(container);
    const valor = nome || "[Nome do cliente]";
    let posicao = 0;
    PLACEHOLDER_NOME_RE.lastIndex = 0;
    let match;
    while ((match = PLACEHOLDER_NOME_RE.exec(template)) !== null) {
        if (match.index > posicao) {
            container.appendChild(document.createTextNode(template.slice(posicao, match.index)));
        }
        container.appendChild(el("span", { class: "pill-val", text: valor }));
        posicao = match.index + match[0].length;
    }
    if (posicao < template.length) {
        container.appendChild(document.createTextNode(template.slice(posicao)));
    }
}
