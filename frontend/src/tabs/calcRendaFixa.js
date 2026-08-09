// "Calculadora de Renda Fixa": um único título, com o resultado em
// destaque (valor líquido no vencimento) e os detalhes do cálculo por trás
// de um toggle. Plugado como categoria "render()" custom dentro da aba
// Calculadora (ver tabs/calculadora.js). Para comparar dois títulos lado a
// lado, ver a aba própria "Comparadora" (tabs/comparadora.js).

import { state } from "../state.js";
import { el, clear } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import { formatarReais, formatarPercentual } from "../util/numeros.js";
import { CAMPOS_RF, calcularTituloRF, lerValoresRF, renderCampoRF } from "../util/rendaFixa.js";

const PREFIXO = "calcrf";

const LINHAS_DETALHE = [
    { chave: "diasCorridos", label: "Nº de dias corridos" },
    { chave: "diasUteis", label: "Nº de dias úteis" },
    { chave: "aliquotaIR", label: "Alíquota de IR" },
    { chave: "rentBrutaMes", label: "Rentabilidade bruta ao mês" },
    { chave: "rentLiquidaMes", label: "Rentabilidade líquida ao mês" },
];

let refs = {};

export function renderCalculadoraRendaFixa() {
    refs = { outputsDetalhe: {} };
    const grid = el("div", { class: "calc-grid rf" });

    const cardDados = el("div", { class: "ccard" });
    cardDados.appendChild(el("h4", { text: "Dados do título" }));
    for (const campo of CAMPOS_RF) {
        cardDados.appendChild(renderCampoRF(state.calcRF, `${PREFIXO}.${campo.chave}`, campo, atualizar));
    }
    grid.appendChild(cardDados);

    const cardResultado = el("div", { class: "ccard rf-res" });
    cardResultado.appendChild(el("h4", { text: "Resultado" }));

    const hero = el("div", { class: "rf-hero" });
    hero.appendChild(el("span", { class: "rf-hero-lbl", text: "Valor líquido no vencimento" }));
    refs.heroVal = el("span", { class: "rf-hero-val" });
    refs.heroSub = el("span", { class: "rf-hero-sub" });
    hero.appendChild(refs.heroVal);
    hero.appendChild(refs.heroSub);
    cardResultado.appendChild(hero);

    const rowBruto = el("div", { class: "crow" });
    rowBruto.appendChild(el("span", { class: "clbl", text: "Valor bruto" }));
    refs.valorBruto = el("span", { class: "cout" });
    rowBruto.appendChild(refs.valorBruto);
    cardResultado.appendChild(rowBruto);

    const rowIR = el("div", { class: "crow" });
    refs.irLbl = el("span", { class: "clbl" });
    rowIR.appendChild(refs.irLbl);
    refs.irVal = el("span", { class: "cout neg" });
    rowIR.appendChild(refs.irVal);
    cardResultado.appendChild(rowIR);

    const rowBrutaAno = el("div", { class: "crow" });
    rowBrutaAno.appendChild(el("span", { class: "clbl", text: "Rentabilidade bruta (a.a.)" }));
    refs.rentBrutaAno = el("span", { class: "cout" });
    rowBrutaAno.appendChild(refs.rentBrutaAno);
    cardResultado.appendChild(rowBrutaAno);

    refs.toggle = el("button", { class: "detail-toggle", type: "button" });
    refs.toggle.insertAdjacentHTML("beforeend", icons.iconChevronBaixo);
    refs.toggleTexto = document.createTextNode("Ver detalhes do cálculo");
    refs.toggle.appendChild(refs.toggleTexto);
    refs.toggle.addEventListener("click", () => {
        const aberto = !refs.detalhe.hidden;
        refs.detalhe.hidden = aberto;
        refs.toggle.classList.toggle("open", !aberto);
        refs.toggleTexto.textContent = aberto ? "Ver detalhes do cálculo" : "Ocultar detalhes";
    });
    cardResultado.appendChild(refs.toggle);

    refs.detalhe = el("div", { class: "rf-detail" });
    refs.detalhe.hidden = true;
    for (const linha of LINHAS_DETALHE) {
        const linhaEl = el("div", { class: "crow" });
        linhaEl.appendChild(el("span", { class: "clbl", text: linha.label }));
        const saida = el("span", { class: "cout" });
        refs.outputsDetalhe[linha.chave] = saida;
        linhaEl.appendChild(saida);
        refs.detalhe.appendChild(linhaEl);
    }
    cardResultado.appendChild(refs.detalhe);

    grid.appendChild(cardResultado);

    atualizar();
    return grid;
}

function atualizar() {
    const v = lerValoresRF(state.calcRF, PREFIXO);
    const r = calcularTituloRF(v);

    if (!r) {
        refs.heroVal.textContent = "—";
        clear(refs.heroSub);
        refs.heroSub.textContent = "Preencha os dados do título (com vencimento após a aplicação) pra calcular.";
        refs.valorBruto.textContent = "";
        refs.irLbl.textContent = "IR a ser pago";
        refs.irVal.textContent = "";
        refs.rentBrutaAno.textContent = "";
        for (const linha of Object.values(refs.outputsDetalhe)) linha.textContent = "";
        return;
    }

    refs.heroVal.textContent = formatarReais(r.valorLiquido);
    clear(refs.heroSub);
    if (r.rentLiquidaAno !== null) {
        refs.heroSub.appendChild(document.createTextNode("Rentabilidade líquida de "));
        refs.heroSub.appendChild(el("b", { text: `${formatarPercentual(r.rentLiquidaAno * 100)} a.a.` }));
        refs.heroSub.appendChild(document.createTextNode(` · ${formatarPercentual(r.rentLiquidaMes * 100)} a.m.`));
    }

    refs.valorBruto.textContent = formatarReais(r.valorBruto);
    refs.irLbl.textContent = `IR a ser pago (${r.isento ? "Isento" : formatarPercentual(r.aliquota * 100)})`;
    refs.irVal.textContent = `− ${formatarReais(r.irPagar)}`;
    refs.rentBrutaAno.textContent = formatarPercentual(v.taxaAno);

    refs.outputsDetalhe.diasCorridos.textContent = String(r.diasCorridos);
    refs.outputsDetalhe.diasUteis.textContent = String(r.diasUteis);
    refs.outputsDetalhe.aliquotaIR.textContent = r.isento ? "Isento" : formatarPercentual(r.aliquota * 100);
    refs.outputsDetalhe.rentBrutaMes.textContent = formatarPercentual(r.rentBrutaMes * 100);
    refs.outputsDetalhe.rentLiquidaMes.textContent = r.rentLiquidaMes !== null ? formatarPercentual(r.rentLiquidaMes * 100) : "";
}
