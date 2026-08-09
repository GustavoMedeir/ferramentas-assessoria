// Lógica e campos de entrada compartilhados entre a Calculadora de Renda
// Fixa (título único, dentro da aba Calculadora — ver tabs/calcRendaFixa.js)
// e a Comparadora de Renda Fixa (dois títulos lado a lado, aba própria —
// ver tabs/comparadora.js). Mantido num só lugar pra não duplicar a
// matemática nem o parsing/máscara dos campos entre as duas telas.

import { el } from "../ui/components.js";
import {
    parseNumeroPtBR,
    formatarReais,
    formatarMilharEnquantoDigita,
    formatarPercentualEnquantoDigita,
} from "./numeros.js";
import { parseDataInput, diasUteisEntre } from "./feriados.js";
import { aliquotaIRDias } from "./financeiro.js";

const DIA_MS = 86400000;

export const TIPOS_ATIVO = [
    { id: "CDB", label: "CDB" },
    { id: "LCI", label: "LCI" },
    { id: "LCA", label: "LCA" },
];

export const CAMPOS_RF = [
    { chave: "emissor", label: "Emissor", tipo: "texto", placeholder: "Nome do emissor" },
    { chave: "tipoAtivo", label: "Tipo do ativo", tipo: "select" },
    { chave: "dataAplicacao", label: "Data da aplicação", tipo: "data" },
    { chave: "dataVencimento", label: "Data de vencimento", tipo: "data" },
    { chave: "taxaAno", label: "Taxa contratada (a.a.)", tipo: "percentual", placeholder: "0,00", sufixoTexto: "%" },
    { chave: "valorAplicado", label: "Valor aplicado", tipo: "reais", placeholder: "R$ 0,00" },
];

function formatarSeReais(texto) {
    if (!texto) return "";
    const valor = parseNumeroPtBR(texto);
    return valor !== null ? formatarReais(valor) : texto;
}

// Guard du <= 0 evita potência negativa/infinita quando vencimento é
// anterior (ou igual) à aplicação, ou quando as datas não formam nenhum dia
// útil entre si. Devolve null nesses casos — o chamador decide como avisar.
export function calcularTituloRF(v) {
    const dataApl = parseDataInput(v.dataAplicacao);
    const dataVenc = parseDataInput(v.dataVencimento);
    if (dataApl === null || dataVenc === null || dataVenc <= dataApl) return null;

    const diasCorridos = Math.round((dataVenc - dataApl) / DIA_MS);
    const diasUteis = diasUteisEntre(dataApl, dataVenc);
    if (diasUteis <= 0) return null;

    const isento = v.tipoAtivo === "LCI" || v.tipoAtivo === "LCA";
    const aliquota = isento ? 0 : aliquotaIRDias(diasCorridos);
    const taxa = v.taxaAno / 100;
    const aplicado = v.valorAplicado;

    const valorBruto = aplicado * Math.pow(1 + taxa, diasUteis / 252);
    const irPagar = aliquota * (valorBruto - aplicado);
    const valorLiquido = valorBruto - irPagar;
    const rentLiquidaAno = aplicado > 0 ? Math.pow(valorLiquido / aplicado, 252 / diasUteis) - 1 : null;
    const rentBrutaMes = Math.pow(1 + taxa, 1 / 12) - 1;
    const rentLiquidaMes = rentLiquidaAno !== null ? Math.pow(1 + rentLiquidaAno, 1 / 12) - 1 : null;

    return { diasCorridos, diasUteis, aliquota, isento, valorBruto, irPagar, valorLiquido, rentLiquidaAno, rentBrutaMes, rentLiquidaMes };
}

// Lê os CAMPOS_RF salvos em `bag` (state.calcRF ou state.comparadorRF) sob
// as chaves `${prefixoChave}.<campo>`, já convertidos pro tipo que
// calcularTituloRF espera.
export function lerValoresRF(bag, prefixoChave) {
    const v = {};
    for (const campo of CAMPOS_RF) {
        const texto = bag[`${prefixoChave}.${campo.chave}`] || "";
        if (campo.tipo === "select") v[campo.chave] = texto || TIPOS_ATIVO[0].id;
        else if (campo.tipo === "reais" || campo.tipo === "percentual") v[campo.chave] = parseNumeroPtBR(texto) ?? 0;
        else v[campo.chave] = texto;
    }
    return v;
}

// Renderiza uma linha .crow para um único campo de CAMPOS_RF, lendo/
// gravando em `bag[chave]` (chave já resolvida pelo chamador, ex.:
// "cmp.a.emissor" ou "calcrf.emissor") e chamando `aoMudar()` a cada
// alteração pra recalcular/re-renderizar a tela.
export function renderCampoRF(bag, chave, campo, aoMudar) {
    const linhaEl = el("div", { class: "crow" });
    linhaEl.appendChild(el("span", { class: "clbl", text: campo.label }));

    let input;
    if (campo.tipo === "select") {
        input = el("select", { class: "cselect" });
        for (const op of TIPOS_ATIVO) input.appendChild(el("option", { value: op.id, text: op.label }));
        input.value = bag[chave] || TIPOS_ATIVO[0].id;
        input.addEventListener("change", () => {
            bag[chave] = input.value;
            aoMudar();
        });
    } else if (campo.tipo === "data") {
        input = el("input", { class: "cinput", type: "date" });
        input.value = bag[chave] || "";
        input.addEventListener("input", () => {
            bag[chave] = input.value;
            aoMudar();
        });
    } else if (campo.tipo === "texto") {
        input = el("input", { class: "cinput", type: "text", placeholder: campo.placeholder || "" });
        input.style.textAlign = "left";
        input.value = bag[chave] || "";
        input.addEventListener("input", () => {
            bag[chave] = input.value;
            aoMudar();
        });
    } else {
        input = el("input", { class: "cinput", type: "text", placeholder: campo.placeholder || "" });
        const valorSalvo = bag[chave] || "";
        input.value = campo.tipo === "reais" ? formatarSeReais(valorSalvo) : valorSalvo;
        input.addEventListener("input", () => {
            const distanciaDoFim = input.value.length - input.selectionStart;
            input.value =
                campo.tipo === "reais" ? formatarMilharEnquantoDigita(input.value) : formatarPercentualEnquantoDigita(input.value);
            const pos = Math.max(0, input.value.length - distanciaDoFim);
            input.setSelectionRange(pos, pos);
            bag[chave] = input.value;
            aoMudar();
        });
        if (campo.tipo === "reais") {
            input.addEventListener("focus", () => {
                input.value = input.value.replace(/^R\$\s*/, "");
            });
            input.addEventListener("blur", () => {
                const valor = parseNumeroPtBR(input.value);
                if (valor !== null) {
                    input.value = formatarReais(valor);
                    bag[chave] = input.value;
                }
            });
        }
    }

    const wrap = campo.sufixoTexto ? el("div", { class: "cwrap" }, [input, el("span", { class: "csuffix", text: campo.sufixoTexto })]) : el("div", { class: "cwrap" }, [input]);
    linhaEl.appendChild(wrap);
    return linhaEl;
}
