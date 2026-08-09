// Aba "Comparadora de Renda Fixa": dois títulos lado a lado, com um
// veredito de qual rende mais líquido e uma tabela-matriz comparando
// métrica a métrica (coluna vencedora destacada). Reaproveita o cálculo e
// os campos de util/rendaFixa.js — a mesma lógica usada na Calculadora de
// Renda Fixa (título único, dentro da aba Calculadora).

import { state } from "../state.js";
import { el, clear, montarToolbarLimpar } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import { formatarReais, formatarPercentual } from "../util/numeros.js";
import { formatarDataPtBR, parseDataInput } from "../util/feriados.js";
import { CAMPOS_RF, calcularTituloRF, lerValoresRF, renderCampoRF } from "../util/rendaFixa.js";

const METRICAS = [
    { chave: "tipoVenc", label: "Tipo · vencimento" },
    { chave: "taxaAno", label: "Taxa contratada (a.a.)" },
    { chave: "valorAplicado", label: "Valor aplicado" },
    { chave: "aliquotaIR", label: "Alíquota de IR" },
    { chave: "valorLiquido", label: "Valor líquido" },
    { chave: "rentLiquidaAno", label: "Rentabilidade líquida (a.a.)", destaque: true },
    { chave: "rentLiquidaMes", label: "Rentabilidade líquida (a.m.)" },
];

let refs = {};

export function mount(container) {
    clear(container);
    refs = {};

    container.appendChild(montarToolbarLimpar(state.comparadorRF, () => mount(container)));

    const wrap = el("div", { class: "cmpx" });

    const painelInputs = el("div", { class: "cmpx-inputs" });
    painelInputs.appendChild(renderCartaoInputs("a", "Emissor A"));
    painelInputs.appendChild(renderCartaoInputs("b", "Emissor B"));
    wrap.appendChild(painelInputs);

    const painelResultado = el("div", { class: "cmpx-result" });
    const cardResultado = el("div", { class: "ccard cmp2" });

    refs.veredito = el("div", { class: "cmp2-verdict" });
    refs.veredito.insertAdjacentHTML("beforeend", icons.iconEstrela);
    refs.veredictoTexto = el("span", { class: "cmp2-verdict-txt" });
    refs.veredito.appendChild(refs.veredictoTexto);
    cardResultado.appendChild(refs.veredito);

    const tableWrap = el("div", { class: "comp-table-wrap" });
    const tabela = el("table", { class: "cmp2-tbl" });
    refs.thead = el("thead");
    refs.tbody = el("tbody");
    tabela.appendChild(refs.thead);
    tabela.appendChild(refs.tbody);
    tableWrap.appendChild(tabela);
    cardResultado.appendChild(tableWrap);

    painelResultado.appendChild(cardResultado);
    painelResultado.appendChild(
        el("div", {
            class: "prev-diag",
            text: "Dias úteis consideram o calendário ANBIMA atual (inclui Consciência Negra a partir de 2024).",
        })
    );
    wrap.appendChild(painelResultado);

    container.appendChild(wrap);
    atualizar();
}

function renderCartaoInputs(prefixo, titulo) {
    const card = el("div", { class: "ccard" });
    card.appendChild(el("h4", { text: titulo }));
    for (const campo of CAMPOS_RF) {
        card.appendChild(renderCampoRF(state.comparadorRF, `cmp.${prefixo}.${campo.chave}`, campo, atualizar));
    }
    return card;
}

function nomeEmissor(prefixo, v) {
    return v.emissor?.trim() || (prefixo === "a" ? "Emissor A" : "Emissor B");
}

function valorMetrica(chave, v, r) {
    if (!r) return "—";
    switch (chave) {
        case "tipoVenc": {
            const dataVenc = parseDataInput(v.dataVencimento);
            return `${v.tipoAtivo} · ${dataVenc !== null ? formatarDataPtBR(dataVenc) : "—"}`;
        }
        case "taxaAno":
            return formatarPercentual(v.taxaAno);
        case "valorAplicado":
            return formatarReais(v.valorAplicado);
        case "aliquotaIR":
            return r.isento ? "Isento" : formatarPercentual(r.aliquota * 100);
        case "valorLiquido":
            return formatarReais(r.valorLiquido);
        case "rentLiquidaAno":
            return r.rentLiquidaAno !== null ? formatarPercentual(r.rentLiquidaAno * 100) : "—";
        case "rentLiquidaMes":
            return r.rentLiquidaMes !== null ? formatarPercentual(r.rentLiquidaMes * 100) : "—";
        default:
            return "—";
    }
}

function atualizar() {
    const va = lerValoresRF(state.comparadorRF, "cmp.a");
    const vb = lerValoresRF(state.comparadorRF, "cmp.b");
    const ra = calcularTituloRF(va);
    const rb = calcularTituloRF(vb);

    const nomeA = nomeEmissor("a", va);
    const nomeB = nomeEmissor("b", vb);

    let vencedor = null;
    if (ra?.rentLiquidaAno != null && rb?.rentLiquidaAno != null) {
        if (ra.rentLiquidaAno > rb.rentLiquidaAno) vencedor = "a";
        else if (rb.rentLiquidaAno > ra.rentLiquidaAno) vencedor = "b";
    }

    clear(refs.veredictoTexto);
    if (vencedor) {
        const vencedorNome = vencedor === "a" ? nomeA : nomeB;
        const rVencedor = vencedor === "a" ? ra : rb;
        const rPerdedor = vencedor === "a" ? rb : ra;
        const deltaPP = (rVencedor.rentLiquidaAno - rPerdedor.rentLiquidaAno) * 100;
        refs.veredictoTexto.appendChild(document.createTextNode("Melhor título: "));
        refs.veredictoTexto.appendChild(el("b", { text: vencedorNome }));
        refs.veredictoTexto.appendChild(document.createTextNode(" — rende "));
        refs.veredictoTexto.appendChild(el("b", { text: `${formatarPercentual(deltaPP)} p.p. a.a.` }));
        refs.veredictoTexto.appendChild(document.createTextNode(" a mais no líquido."));
        refs.veredito.style.display = "flex";
    } else if (ra && rb) {
        refs.veredictoTexto.textContent = "Os dois títulos empatam na rentabilidade líquida.";
        refs.veredito.style.display = "flex";
    } else {
        refs.veredito.style.display = "none";
    }

    clear(refs.thead);
    const trh = el("tr");
    trh.appendChild(el("th", { text: "Métrica" }));
    const thA = el("th", { class: vencedor === "a" ? "win" : "" });
    thA.appendChild(document.createTextNode(nomeA));
    if (vencedor === "a") thA.appendChild(el("span", { class: "cmp2-col-tag", text: "Melhor" }));
    trh.appendChild(thA);
    const thB = el("th", { class: vencedor === "b" ? "win" : "" });
    thB.appendChild(document.createTextNode(nomeB));
    if (vencedor === "b") thB.appendChild(el("span", { class: "cmp2-col-tag", text: "Melhor" }));
    trh.appendChild(thB);
    refs.thead.appendChild(trh);

    clear(refs.tbody);
    for (const m of METRICAS) {
        const tr = el("tr", { class: m.destaque ? "hi" : "" });
        tr.appendChild(el("td", { text: m.label }));
        tr.appendChild(el("td", { class: vencedor === "a" ? "win" : "", text: valorMetrica(m.chave, va, ra) }));
        tr.appendChild(el("td", { class: vencedor === "b" ? "win" : "", text: valorMetrica(m.chave, vb, rb) }));
        refs.tbody.appendChild(tr);
    }
}
