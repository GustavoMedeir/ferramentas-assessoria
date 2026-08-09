import { state } from "../state.js";
import { el, clear, montarToolbarLimpar } from "../ui/components.js";
import {
    parseNumeroPtBR,
    formatarReais,
    formatarPercentual,
    formatarNumero,
    formatarMilharEnquantoDigita,
    formatarPercentualEnquantoDigita,
} from "../util/numeros.js";
import { calcularAposentadoria } from "../util/aposentadoria.js";

// Campos do painel "Premissas", na mesma ordem do print de referência.
const CAMPOS = [
    { chave: "dataCalculo", label: "Data do cálculo", unidade: "data" },
    { chave: "idadeAtual", label: "Idade Atual", unidade: "numero", placeholder: "0", sufixoTexto: "anos" },
    { chave: "idadeAposentadoria", label: "Idade de Aposentadoria", unidade: "numero", placeholder: "0", sufixoTexto: "anos" },
    { chave: "expectativaVida", label: "Expectativa de Vida", unidade: "numero", placeholder: "0", sufixoTexto: "anos" },
    { chave: "aplicacoesFinanceiras", label: "Aplicações Financeiras", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "aplicacaoMensal", label: "Aplicação Mensal (atual)", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "rendaDesejada", label: "Renda Desejada na Aposentadoria", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "rendaINSS", label: "Renda Projetada do INSS", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "outrasFontes", label: "Outras fontes de renda", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "taxaAnual", label: "Taxa Real Líquida (a.a.)", unidade: "percentual", placeholder: "0,00" },
    { chave: "patrimonioSucessao", label: "Patrimônio de Sucessão desejado", unidade: "reais", placeholder: "R$ 0,00" },
];

const CENARIOS = [
    {
        id: "consumo",
        titulo: "Consumo de Patrimônio (com Sucessão)",
        cor: "#2563eb",
        avisoSemHorizonte: "Expectativa de vida precisa ser depois da idade de aposentadoria pra calcular este cenário.",
    },
    {
        id: "preservacao",
        titulo: "Viver de Renda (Preservando Patrimônio)",
        cor: "#0d9488",
        avisoSemHorizonte: "Taxa real líquida precisa ser maior que 0% pra sustentar uma renda perpétua.",
    },
];

let refs = {};
let resizeObservers = [];

function chaveEstado(chave) {
    return `apos.${chave}`;
}

function formatarSeReais(texto) {
    if (!texto) return "";
    const valor = parseNumeroPtBR(texto);
    return valor !== null ? formatarReais(valor) : texto;
}

function inicializarDefaults() {
    if (Object.keys(state.aposentadoria).length > 0) return;
    const hoje = new Date();
    const iso = `${hoje.getFullYear()}-${String(hoje.getMonth() + 1).padStart(2, "0")}-${String(hoje.getDate()).padStart(2, "0")}`;
    state.aposentadoria[chaveEstado("dataCalculo")] = iso;
    state.aposentadoria[chaveEstado("taxaAnual")] = "8,00";
}

export function mount(container) {
    // clear() esvazia o container e reseta o scroll pro topo (mesmo problema
    // documentado em tabs/emails.js:renderBlocos) — preserva a posição pra
    // não "puxar pra cima" a cada remontagem (ex.: Confirmar em Configurações).
    const scrollAnterior = container.scrollTop;
    clear(container);
    refs = { cenarios: {} };
    for (const obs of resizeObservers) obs.disconnect();
    resizeObservers = [];
    inicializarDefaults();

    container.appendChild(montarToolbarLimpar(state.aposentadoria, () => mount(container)));

    const painelInputs = el("div", { class: "ccard comp-inputs" });
    painelInputs.appendChild(el("h4", { text: "Premissas" }));

    for (const campo of CAMPOS) {
        const linhaEl = el("div", { class: "crow" });
        linhaEl.appendChild(el("span", { class: "clbl", text: campo.label }));

        let input;
        if (campo.unidade === "data") {
            input = el("input", { class: "cinput", type: "date" });
            input.value = state.aposentadoria[chaveEstado(campo.chave)] || "";
            input.addEventListener("input", () => {
                state.aposentadoria[chaveEstado(campo.chave)] = input.value;
                atualizar();
            });
        } else {
            input = el("input", { class: "cinput", type: "text", placeholder: campo.placeholder || "" });
            const valorSalvo = state.aposentadoria[chaveEstado(campo.chave)] || "";
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
                state.aposentadoria[chaveEstado(campo.chave)] = input.value;
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
                        state.aposentadoria[chaveEstado(campo.chave)] = input.value;
                    }
                });
            }
        }

        const wrap = el("div", { class: "cwrap" }, [input]);
        if (campo.sufixoTexto) wrap.appendChild(el("span", { class: "csuffix", text: campo.sufixoTexto }));
        linhaEl.appendChild(wrap);
        painelInputs.appendChild(linhaEl);
    }

    refs.diag = el("div", { class: "prev-diag" });
    painelInputs.appendChild(refs.diag);

    const painelResultado = el("div", { class: "comp-resultado" });
    const grid = el("div", { class: "cmp-grid" });
    for (const cenario of CENARIOS) {
        grid.appendChild(criarCardCenario(cenario));
    }
    painelResultado.appendChild(grid);

    container.appendChild(el("div", { class: "comp" }, [painelInputs, painelResultado]));

    atualizar();
    container.scrollTop = scrollAnterior;
}

function criarCardCenario(cenario) {
    const card = el("div", { class: "ccard" });
    card.appendChild(el("h4", { text: cenario.titulo }));

    const rowPatrimonio = el("div", { class: "crow total" });
    rowPatrimonio.appendChild(el("span", { class: "clbl", text: "Patrimônio necessário na aposentadoria" }));
    const outPatrimonio = el("span", { class: "cout" });
    rowPatrimonio.appendChild(outPatrimonio);
    card.appendChild(rowPatrimonio);

    const rowParcela = el("div", { class: "crow" });
    rowParcela.appendChild(el("span", { class: "clbl", text: "Parcela mensal necessária" }));
    const outParcela = el("span", { class: "cout" });
    rowParcela.appendChild(outParcela);
    card.appendChild(rowParcela);

    const rowAtual = el("div", { class: "crow" });
    rowAtual.appendChild(el("span", { class: "clbl", text: "Sua aplicação mensal atual" }));
    const outAtual = el("span", { class: "cout" });
    rowAtual.appendChild(outAtual);
    card.appendChild(rowAtual);

    const rowDelta = el("div", { class: "crow total" });
    rowDelta.appendChild(el("span", { class: "clbl", text: "Diferença" }));
    const outDelta = el("span", { class: "cout" });
    rowDelta.appendChild(outDelta);
    card.appendChild(rowDelta);

    const aviso = el("p", { class: "cfg-placeholder apos-aviso" });
    aviso.style.display = "none";
    card.appendChild(aviso);

    const chartWrap = el("div", { class: "chart-canvas-wrap apos-chart" });
    const canvas = el("canvas");
    chartWrap.appendChild(canvas);
    card.appendChild(chartWrap);

    const observer = new ResizeObserver(() => {
        const r = refs.cenarios[cenario.id];
        desenharGrafico(cenario, r.dados, r.idadeAtual, r.expectativaVida);
    });
    observer.observe(chartWrap);
    resizeObservers.push(observer);

    refs.cenarios[cenario.id] = { card, outPatrimonio, outParcela, outAtual, outDelta, aviso, chartWrap, canvas, dados: null };
    return card;
}

function atualizar() {
    const v = {};
    for (const campo of CAMPOS) {
        const texto = state.aposentadoria[chaveEstado(campo.chave)] || "";
        if (campo.unidade === "data") continue;
        v[campo.chave] = campo.unidade === "percentual" ? (parseNumeroPtBR(texto) ?? 0) / 100 : parseNumeroPtBR(texto) ?? 0;
    }

    const resultado = calcularAposentadoria({
        idadeAtual: v.idadeAtual,
        idadeAposentadoria: v.idadeAposentadoria,
        expectativaVida: v.expectativaVida,
        aplicacoesFinanceiras: v.aplicacoesFinanceiras,
        aplicacaoMensal: v.aplicacaoMensal,
        rendaDesejada: v.rendaDesejada,
        rendaINSS: v.rendaINSS,
        outrasFontes: v.outrasFontes,
        taxaAnual: v.taxaAnual,
        patrimonioSucessao: v.patrimonioSucessao,
    });

    const anosAcum = resultado.nAcumMeses / 12;
    const anosConsumo = resultado.nConsumoMeses / 12;
    if (resultado.nAcumMeses <= 0) {
        refs.diag.textContent = "A idade de aposentadoria precisa ser depois da idade atual pra calcular.";
    } else {
        refs.diag.textContent =
            `Horizonte de acumulação: ${formatarNumero(anosAcum, 0)} anos · ` +
            `Horizonte de aposentadoria: ${anosConsumo > 0 ? formatarNumero(anosConsumo, 0) + " anos" : "—"} · ` +
            `Taxa real líquida mensal: ${formatarPercentual(resultado.taxaMensal * 100)} · ` +
            `Renda mensal necessária do patrimônio: ${formatarReais(resultado.rendaNecessaria)}`;
    }

    for (const cenario of CENARIOS) {
        const dados = resultado[cenario.id];
        const r = refs.cenarios[cenario.id];
        r.dados = dados;
        r.idadeAtual = v.idadeAtual;
        r.expectativaVida = v.expectativaVida;
        if (!dados) {
            r.outPatrimonio.textContent = "—";
            r.outParcela.textContent = "—";
            r.outAtual.textContent = "";
            r.outDelta.textContent = "";
            r.outDelta.classList.remove("pos", "neg");
            r.aviso.textContent = resultado.nAcumMeses <= 0 ? "Ajuste as idades pra calcular." : cenario.avisoSemHorizonte;
            r.aviso.style.display = "block";
            desenharGrafico(cenario, null, v.idadeAtual, v.expectativaVida);
            continue;
        }

        r.aviso.style.display = "none";
        r.outPatrimonio.textContent = formatarReais(dados.patrimonioNecessario);
        r.outParcela.textContent = formatarReais(dados.parcelaMensal);
        r.outAtual.textContent = formatarReais(v.aplicacaoMensal);
        const delta = v.aplicacaoMensal - dados.parcelaMensal;
        if (dados.parcelaMensal <= 0) {
            r.outDelta.textContent = "Patrimônio atual já é suficiente";
            r.outDelta.classList.add("pos");
            r.outDelta.classList.remove("neg");
        } else {
            r.outDelta.textContent = `${delta >= 0 ? "Sobram " : "Faltam "}${formatarReais(Math.abs(delta))}/mês`;
            r.outDelta.classList.toggle("pos", delta >= 0);
            r.outDelta.classList.toggle("neg", delta < 0);
        }
        desenharGrafico(cenario, dados, v.idadeAtual, v.expectativaVida);
    }
}

function desenharGrafico(cenario, dados, idadeAtual, expectativaVida) {
    const refCenario = refs.cenarios?.[cenario.id];
    if (!refCenario || !refCenario.canvas || !refCenario.chartWrap) return;
    const ctx2d = refCenario.canvas.getContext("2d");
    const larguraCss = Math.max(refCenario.chartWrap.clientWidth, 200);
    const alturaCss = 200;
    const dpr = window.devicePixelRatio || 1;
    refCenario.canvas.width = larguraCss * dpr;
    refCenario.canvas.height = alturaCss * dpr;
    refCenario.canvas.style.width = larguraCss + "px";
    refCenario.canvas.style.height = alturaCss + "px";
    ctx2d.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx2d.clearRect(0, 0, larguraCss, alturaCss);

    const appEl = document.getElementById("app");
    const estilo = getComputedStyle(appEl);
    const corInk2 = estilo.getPropertyValue("--ink-2").trim() || "#5b716b";
    const corLine2 = estilo.getPropertyValue("--line-2").trim() || "#e8f0ed";

    const padL = 58;
    const padR = 12;
    const padT = 10;
    const padB = 22;
    const larguraPlot = larguraCss - padL - padR;
    const alturaPlot = alturaCss - padT - padB;

    const pontos = dados?.serieAnual;
    if (!pontos || pontos.length < 2) {
        ctx2d.strokeStyle = corLine2;
        ctx2d.beginPath();
        ctx2d.moveTo(padL, padT);
        ctx2d.lineTo(padL, padT + alturaPlot);
        ctx2d.lineTo(padL + larguraPlot, padT + alturaPlot);
        ctx2d.stroke();
        return;
    }

    const anoMax = expectativaVida;
    const anoAposentadoria = pontos.find((p) => p.fase === "consumo")?.anoRelativo ?? null;
    const primeiroAnoConsumo = anoAposentadoria !== null ? anoAposentadoria - 1 : null;
    let maxY = 0;
    for (const p of pontos) maxY = Math.max(maxY, p.patrimonio);
    if (maxY <= 0) maxY = 1;

    const escalaX = (ano) => padL + (ano / anoMax) * larguraPlot;
    const escalaY = (val) => padT + alturaPlot - (val / maxY) * alturaPlot;

    ctx2d.font = "10.5px 'Plus Jakarta Sans', sans-serif";
    ctx2d.fillStyle = corInk2;
    const NGRID = 3;
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
        const compacto = Math.abs(val) >= 1000 ? `R$ ${(val / 1000).toLocaleString("pt-BR", { maximumFractionDigits: 0 })} mil` : formatarReais(val);
        ctx2d.fillText(compacto, padL - 6, y);
    }

    if (primeiroAnoConsumo !== null) {
        const x = escalaX(primeiroAnoConsumo);
        ctx2d.save();
        ctx2d.setLineDash([4, 4]);
        ctx2d.strokeStyle = corInk2;
        ctx2d.beginPath();
        ctx2d.moveTo(x, padT);
        ctx2d.lineTo(x, padT + alturaPlot);
        ctx2d.stroke();
        ctx2d.restore();
    }

    ctx2d.strokeStyle = cenario.cor;
    ctx2d.lineWidth = 2.5;
    ctx2d.beginPath();
    pontos.forEach((p, idx) => {
        const x = escalaX(p.anoRelativo);
        const y = escalaY(p.patrimonio);
        if (idx === 0) ctx2d.moveTo(x, y);
        else ctx2d.lineTo(x, y);
    });
    ctx2d.stroke();

    ctx2d.textAlign = "center";
    ctx2d.textBaseline = "top";
    ctx2d.fillText("hoje", escalaX(idadeAtual), padT + alturaPlot + 5);
    if (primeiroAnoConsumo !== null) ctx2d.fillText("aposentadoria", escalaX(primeiroAnoConsumo), padT + alturaPlot + 5);
    ctx2d.fillText(`${anoMax} anos`, escalaX(anoMax), padT + alturaPlot + 5);
}
