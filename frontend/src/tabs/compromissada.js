import { state } from "../state.js";
import { el, clear, montarToolbarLimpar } from "../ui/components.js";
import {
    parseNumeroPtBR,
    formatarReais,
    formatarPercentual,
    formatarMilharEnquantoDigita,
    formatarPercentualEnquantoDigita,
} from "../util/numeros.js";
import { parseDataInput, formatarDataPtBR } from "../util/feriados.js";
import { calcularCompromissada } from "../util/compromissada.js";

// Campos do painel "Premissas", na mesma ordem da planilha de referência.
// unidade "data" usa <input type="date"> — o valor bate direto com o
// formato esperado por parseDataInput ("YYYY-MM-DD"), sem máscara.
const CAMPOS = [
    { chave: "financeiro", label: "Financeiro", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "diasUteisAno", label: "Dias úteis no ano", unidade: "numero", placeholder: "252" },
    { chave: "selic", label: "SELIC atual", unidade: "percentual", placeholder: "0,00" },
    { chave: "dataInicio", label: "Data início", unidade: "data" },
    { chave: "txCR", label: "Taxa Conta Remunerada (% do CDI)", unidade: "percentual", placeholder: "0,00" },
    { chave: "txCDB", label: "Taxa CDB (% do CDI)", unidade: "percentual", placeholder: "0,00" },
    { chave: "txComp", label: "Taxa Compromissada (% do CDI)", unidade: "percentual", placeholder: "0,00" },
    { chave: "horizonte", label: "Horizonte (dias úteis)", unidade: "numero", placeholder: "20" },
    { chave: "spread", label: "SPREAD Plataforma (% do CDI)", unidade: "percentual", placeholder: "0,00", soAssessor: true },
];

const SERIES = [
    { chave: "contaRemunerada", label: "Conta Remunerada", cor: "#0d9488" },
    { chave: "cdb", label: "CDB", cor: "#2563eb" },
    { chave: "compromissada", label: "Compromissada", cor: "#d97706" },
    { chave: "comissao", label: "Comissão Escritório", cor: "#9333ea", soAssessor: true },
];

let refs = {};
let dadosAtuais = { cdiDiario: 0, linhas: [], primeiroDiaCDBGanha: null };
let seriesOcultas = new Set();
let resizeObserver = null;
let layoutAtual = null; // {padL, larguraPlot, linhas} do último desenho — usado pelo hover pra achar o D.U. sob o cursor
let hoverIndex = null;

function chaveEstado(chave) {
    return `comp.${chave}`;
}

function visaoAssessor() {
    return state.prefs.visao === "assessor";
}

function seriesParaVisao() {
    return SERIES.filter((s) => !s.soAssessor || visaoAssessor());
}

function seriesAtivas() {
    return seriesParaVisao().filter((s) => !seriesOcultas.has(s.chave));
}

function formatarSeReais(texto) {
    if (!texto) return "";
    const valor = parseNumeroPtBR(texto);
    return valor !== null ? formatarReais(valor) : texto;
}

function formatarFracao(valor, casas = 8) {
    if (!Number.isFinite(valor)) return "";
    return valor.toLocaleString("pt-BR", { minimumFractionDigits: casas, maximumFractionDigits: casas });
}

function formatarReaisCompacto(valor) {
    if (Math.abs(valor) >= 1000) {
        return `R$ ${(valor / 1000).toLocaleString("pt-BR", { maximumFractionDigits: 1 })} mil`;
    }
    return formatarReais(valor);
}

// Só roda na primeira vez que a aba é aberta (state.compromissada ainda
// vazio) — evita sobrescrever o que o usuário já digitou em re-mounts
// disparados por troca de tema/visão.
function inicializarDefaults() {
    if (Object.keys(state.compromissada).length > 0) return;
    state.compromissada[chaveEstado("diasUteisAno")] = "252";
    state.compromissada[chaveEstado("horizonte")] = "20";
    const hoje = new Date();
    const iso = `${hoje.getFullYear()}-${String(hoje.getMonth() + 1).padStart(2, "0")}-${String(hoje.getDate()).padStart(2, "0")}`;
    state.compromissada[chaveEstado("dataInicio")] = iso;
}

export function mount(container, ctx) {
    // clear() esvazia o container e reseta o scroll pro topo (mesmo problema
    // documentado em tabs/emails.js:renderBlocos) — preserva a posição pra
    // não "puxar pra cima" a cada remontagem (ex.: Confirmar em Configurações).
    const scrollAnterior = container.scrollTop;
    clear(container);
    refs = { outputs: {} };
    if (resizeObserver) {
        resizeObserver.disconnect();
        resizeObserver = null;
    }
    seriesOcultas = new Set();
    layoutAtual = null;
    hoverIndex = null;
    inicializarDefaults();

    container.appendChild(montarToolbarLimpar(state.compromissada, () => mount(container, ctx)));

    // ---- painel esquerdo: premissas ----
    const painelInputs = el("div", { class: "ccard comp-inputs" });
    painelInputs.appendChild(el("h4", { text: "Premissas" }));

    for (const campo of CAMPOS) {
        if (campo.soAssessor && !visaoAssessor()) continue;

        const linhaEl = el("div", { class: "crow" });
        linhaEl.appendChild(el("span", { class: "clbl", text: campo.label }));

        let input;
        if (campo.unidade === "data") {
            input = el("input", { class: "cinput", type: "date" });
            input.value = state.compromissada[chaveEstado(campo.chave)] || "";
            input.addEventListener("input", () => {
                state.compromissada[chaveEstado(campo.chave)] = input.value;
                atualizar();
            });
        } else {
            input = el("input", { class: "cinput", type: "text", placeholder: campo.placeholder || "" });
            const valorSalvo = state.compromissada[chaveEstado(campo.chave)] || "";
            input.value = campo.unidade === "reais" ? formatarSeReais(valorSalvo) : valorSalvo;
            input.addEventListener("input", () => {
                if (campo.unidade === "reais" || campo.unidade === "percentual") {
                    const distanciaDoFim = input.value.length - input.selectionStart;
                    input.value =
                        campo.unidade === "reais"
                            ? formatarMilharEnquantoDigita(input.value)
                            : formatarPercentualEnquantoDigita(input.value);
                    const pos = Math.max(0, input.value.length - distanciaDoFim);
                    input.setSelectionRange(pos, pos);
                }
                state.compromissada[chaveEstado(campo.chave)] = input.value;
                atualizar();
            });
            if (campo.unidade === "reais") {
                input.addEventListener("focus", () => {
                    input.value = input.value.replace(/^R\$\s*/, "");
                });
                input.addEventListener("blur", () => {
                    const valor = parseNumeroPtBR(input.value);
                    if (valor !== null) {
                        input.value = formatarReais(valor);
                        state.compromissada[chaveEstado(campo.chave)] = input.value;
                    }
                });
            }
        }

        linhaEl.appendChild(el("div", { class: "cwrap" }, [input]));
        painelInputs.appendChild(linhaEl);
    }

    refs.diagCdi = el("div", { class: "prev-diag" });
    painelInputs.appendChild(refs.diagCdi);

    // ---- painel direito: gráfico + destaques + tabela ----
    const painelResultado = el("div", { class: "comp-resultado" });

    const chartCard = el("div", { class: "ccard chart-card" });
    chartCard.appendChild(el("h4", { text: "Rendimento por dia útil" }));
    refs.legenda = el("div", { class: "chart-legend" });
    chartCard.appendChild(refs.legenda);
    refs.canvasWrap = el("div", { class: "chart-canvas-wrap" });
    refs.canvas = el("canvas");
    refs.canvasWrap.appendChild(refs.canvas);
    refs.tooltip = el("div", { class: "chart-tooltip" });
    refs.canvasWrap.appendChild(refs.tooltip);
    chartCard.appendChild(refs.canvasWrap);
    painelResultado.appendChild(chartCard);

    refs.canvas.addEventListener("mousemove", (e) => {
        if (!layoutAtual || !layoutAtual.linhas.length) return;
        const rect = refs.canvas.getBoundingClientRect();
        const xCss = e.clientX - rect.left;
        const idx = indicePorX(xCss);
        if (idx !== hoverIndex) {
            hoverIndex = idx;
            desenharGrafico();
        }
        atualizarTooltip(xCss);
    });
    refs.canvas.addEventListener("mouseleave", () => {
        if (hoverIndex !== null) {
            hoverIndex = null;
            desenharGrafico();
        }
        refs.tooltip.style.display = "none";
    });

    const cardDestaque = el("div", { class: "ccard comp-destaque" });
    cardDestaque.appendChild(el("h4", { text: "Dia em que o CDB fica à frente" }));
    refs.destaqueTexto = el("div", { class: "comp-destaque-valor" });
    cardDestaque.appendChild(refs.destaqueTexto);
    painelResultado.appendChild(cardDestaque);

    const tabelaWrap = el("div", { class: "table-card comp-table-wrap" });
    const tabela = el("table", { class: "dtable prev-table" });
    const colunas = ["D.U.", "Data", "IOF", "Conta Remunerada", "CDB", "Compromissada", "Equiv. do CDB"];
    if (visaoAssessor()) colunas.push("Comissão Escritório");
    tabela.appendChild(el("thead", {}, [el("tr", {}, colunas.map((c) => el("th", { text: c })))]));
    refs.tbody = el("tbody");
    tabela.appendChild(refs.tbody);
    tabelaWrap.appendChild(tabela);
    painelResultado.appendChild(tabelaWrap);

    container.appendChild(el("div", { class: "comp" }, [painelInputs, painelResultado]));

    resizeObserver = new ResizeObserver(() => desenharGrafico());
    resizeObserver.observe(refs.canvasWrap);

    atualizar();
    container.scrollTop = scrollAnterior;
}

function atualizar() {
    const v = {};
    for (const campo of CAMPOS) {
        const texto = state.compromissada[chaveEstado(campo.chave)] || "";
        if (campo.unidade === "data") {
            v[campo.chave] = parseDataInput(texto);
        } else {
            v[campo.chave] = parseNumeroPtBR(texto) ?? 0;
        }
    }

    dadosAtuais = calcularCompromissada({
        financeiro: v.financeiro,
        diasUteisAno: v.diasUteisAno,
        selic: v.selic / 100,
        dataInicioTs: v.dataInicio,
        txCR: v.txCR / 100,
        txCDB: v.txCDB / 100,
        txComp: v.txComp / 100,
        horizonte: Math.max(1, Math.min(2520, Math.round(v.horizonte || 20))),
        spread: visaoAssessor() ? v.spread / 100 : 0,
    });

    refs.diagCdi.textContent = `CDI diário: ${formatarFracao(dadosAtuais.cdiDiario)}`;
    renderLegenda();
    renderDestaque();
    renderTabela();
    desenharGrafico();
}

function renderLegenda() {
    if (!refs.legenda) return;
    clear(refs.legenda);
    for (const s of seriesParaVisao()) {
        const item = el("button", { class: `legend-item${seriesOcultas.has(s.chave) ? " off" : ""}`, type: "button" });
        const swatch = el("span", { class: "legend-swatch" });
        swatch.style.background = s.cor;
        item.appendChild(swatch);
        item.appendChild(el("span", { text: s.label }));
        item.addEventListener("click", () => {
            if (seriesOcultas.has(s.chave)) seriesOcultas.delete(s.chave);
            else seriesOcultas.add(s.chave);
            item.classList.toggle("off");
            desenharGrafico();
        });
        refs.legenda.appendChild(item);
    }
}

function renderDestaque() {
    if (!refs.destaqueTexto) return;
    refs.destaqueTexto.classList.remove("pos");
    if (!dadosAtuais.linhas.length) {
        refs.destaqueTexto.textContent = "Preencha os campos para calcular.";
        return;
    }
    if (dadosAtuais.primeiroDiaCDBGanha === null) {
        refs.destaqueTexto.textContent = "O CDB não ultrapassa a Compromissada dentro do horizonte analisado.";
        return;
    }
    const n = dadosAtuais.primeiroDiaCDBGanha;
    const linha = dadosAtuais.linhas[n - 1];
    refs.destaqueTexto.textContent = `${n}º dia útil (${formatarDataPtBR(linha.dataTs)})`;
    refs.destaqueTexto.classList.add("pos");
}

function renderTabela() {
    if (!refs.tbody) return;
    clear(refs.tbody);
    const assessor = visaoAssessor();
    for (const linha of dadosAtuais.linhas) {
        const tr = el("tr", { class: linha.du === dadosAtuais.primeiroDiaCDBGanha ? "comp-row-cruzamento" : "" });
        tr.appendChild(el("td", { text: String(linha.du) }));
        tr.appendChild(el("td", { text: formatarDataPtBR(linha.dataTs) }));
        tr.appendChild(el("td", { class: "prev-val", text: formatarPercentual(linha.iof * 100) }));
        tr.appendChild(el("td", { class: "prev-val", text: formatarReais(linha.contaRemunerada) }));
        tr.appendChild(el("td", { class: "prev-val", text: formatarReais(linha.cdb) }));
        tr.appendChild(el("td", { class: "prev-val", text: formatarReais(linha.compromissada) }));
        tr.appendChild(
            el("td", { class: "prev-val", text: linha.equivCDB !== null ? formatarPercentual(linha.equivCDB * 100) : "—" })
        );
        if (assessor) tr.appendChild(el("td", { class: "prev-val", text: formatarReais(linha.comissao) }));
        refs.tbody.appendChild(tr);
    }
}

// Índice do ponto (dentro de dadosAtuais.linhas) mais próximo de uma
// posição X em pixels CSS do canvas, usando o layout do último desenho.
function indicePorX(xCss) {
    const { padL, larguraPlot, linhas } = layoutAtual;
    if (linhas.length <= 1) return 0;
    const frac = (xCss - padL) / larguraPlot;
    const idx = Math.round(frac * (linhas.length - 1));
    return Math.min(linhas.length - 1, Math.max(0, idx));
}

function atualizarTooltip(xCss) {
    if (hoverIndex === null || !dadosAtuais.linhas[hoverIndex]) {
        refs.tooltip.style.display = "none";
        return;
    }
    const linha = dadosAtuais.linhas[hoverIndex];
    const ativas = seriesAtivas();

    clear(refs.tooltip);
    refs.tooltip.appendChild(el("div", { class: "ctt-titulo", text: `D.U. ${linha.du} — ${formatarDataPtBR(linha.dataTs)}` }));
    for (const s of ativas) {
        const linhaTooltip = el("div", { class: "ctt-row" });
        const dot = el("span", { class: "ctt-dot" });
        dot.style.background = s.cor;
        linhaTooltip.appendChild(dot);
        linhaTooltip.appendChild(el("span", { class: "ctt-label", text: s.label }));
        linhaTooltip.appendChild(el("span", { class: "ctt-valor", text: formatarReais(linha[s.chave]) }));
        refs.tooltip.appendChild(linhaTooltip);
    }

    refs.tooltip.style.display = "block";
    const larguraWrap = refs.canvasWrap.clientWidth;
    const larguraTooltip = refs.tooltip.offsetWidth;
    let esquerda = xCss + 14;
    if (esquerda + larguraTooltip > larguraWrap) esquerda = xCss - larguraTooltip - 14;
    refs.tooltip.style.left = `${Math.max(0, esquerda)}px`;
    refs.tooltip.style.top = "8px";
}

function desenharGrafico() {
    if (!refs.canvas || !refs.canvasWrap) return;
    const ctx2d = refs.canvas.getContext("2d");
    const larguraCss = Math.max(refs.canvasWrap.clientWidth, 200);
    const alturaCss = 280;
    const dpr = window.devicePixelRatio || 1;
    refs.canvas.width = larguraCss * dpr;
    refs.canvas.height = alturaCss * dpr;
    refs.canvas.style.width = larguraCss + "px";
    refs.canvas.style.height = alturaCss + "px";
    ctx2d.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx2d.clearRect(0, 0, larguraCss, alturaCss);

    const appEl = document.getElementById("app");
    const estilo = getComputedStyle(appEl);
    const corInk2 = estilo.getPropertyValue("--ink-2").trim() || "#5b716b";
    const corLine2 = estilo.getPropertyValue("--line-2").trim() || "#e8f0ed";

    const padL = 66;
    const padR = 16;
    const padT = 12;
    const padB = 26;
    const larguraPlot = larguraCss - padL - padR;
    const alturaPlot = alturaCss - padT - padB;

    ctx2d.strokeStyle = corLine2;
    ctx2d.lineWidth = 1;

    const linhas = dadosAtuais.linhas;
    const ativas = seriesAtivas();

    if (!linhas.length || !ativas.length) {
        layoutAtual = null;
        refs.tooltip.style.display = "none";
        ctx2d.beginPath();
        ctx2d.moveTo(padL, padT);
        ctx2d.lineTo(padL, padT + alturaPlot);
        ctx2d.lineTo(padL + larguraPlot, padT + alturaPlot);
        ctx2d.stroke();
        return;
    }

    layoutAtual = { padL, larguraPlot, linhas };
    if (hoverIndex !== null && hoverIndex >= linhas.length) hoverIndex = null;

    let maxY = 0;
    for (const linha of linhas) for (const s of ativas) maxY = Math.max(maxY, linha[s.chave]);
    if (maxY <= 0) maxY = 1;

    const escalaX = (n) => padL + ((n - 1) / Math.max(linhas.length - 1, 1)) * larguraPlot;
    const escalaY = (val) => padT + alturaPlot - (val / maxY) * alturaPlot;

    ctx2d.font = "11px 'Plus Jakarta Sans', sans-serif";
    ctx2d.fillStyle = corInk2;
    const NGRID = 4;
    for (let i = 0; i <= NGRID; i++) {
        const val = (maxY / NGRID) * i;
        const y = escalaY(val);
        ctx2d.strokeStyle = corLine2;
        ctx2d.beginPath();
        ctx2d.moveTo(padL, y);
        ctx2d.lineTo(padL + larguraPlot, y);
        ctx2d.stroke();
        ctx2d.textAlign = "right";
        ctx2d.textBaseline = "middle";
        ctx2d.fillText(formatarReaisCompacto(val), padL - 8, y);
    }

    ctx2d.textAlign = "center";
    ctx2d.textBaseline = "top";
    const passo = Math.max(1, Math.ceil(linhas.length / 8));
    for (let i = 0; i < linhas.length; i += passo) {
        const x = escalaX(linhas[i].du);
        ctx2d.fillText(String(linhas[i].du), x, padT + alturaPlot + 6);
    }

    for (const s of ativas) {
        ctx2d.strokeStyle = s.cor;
        ctx2d.lineWidth = 2;
        ctx2d.beginPath();
        linhas.forEach((linha, idx) => {
            const x = escalaX(linha.du);
            const y = escalaY(linha[s.chave]);
            if (idx === 0) ctx2d.moveTo(x, y);
            else ctx2d.lineTo(x, y);
        });
        ctx2d.stroke();
    }

    if (dadosAtuais.primeiroDiaCDBGanha) {
        const x = escalaX(dadosAtuais.primeiroDiaCDBGanha);
        ctx2d.save();
        ctx2d.setLineDash([4, 4]);
        ctx2d.strokeStyle = corInk2;
        ctx2d.beginPath();
        ctx2d.moveTo(x, padT);
        ctx2d.lineTo(x, padT + alturaPlot);
        ctx2d.stroke();
        ctx2d.restore();
    }

    if (hoverIndex !== null) {
        const linhaHover = linhas[hoverIndex];
        const xHover = escalaX(linhaHover.du);
        ctx2d.strokeStyle = corInk2;
        ctx2d.lineWidth = 1;
        ctx2d.beginPath();
        ctx2d.moveTo(xHover, padT);
        ctx2d.lineTo(xHover, padT + alturaPlot);
        ctx2d.stroke();

        for (const s of ativas) {
            const y = escalaY(linhaHover[s.chave]);
            ctx2d.beginPath();
            ctx2d.fillStyle = s.cor;
            ctx2d.arc(xHover, y, 3.5, 0, Math.PI * 2);
            ctx2d.fill();
        }
    }
}
