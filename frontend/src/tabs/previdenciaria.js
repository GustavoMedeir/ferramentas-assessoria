import { state } from "../state.js";
import { el, clear, montarToolbarLimpar } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import { parseNumeroPtBR, formatarReais, formatarMilharEnquantoDigita } from "../util/numeros.js";
import { calcularPrevidenciaria, TABELAS } from "../util/previdenciaria.js";

// Inputs na mesma ordem do painel "Premissas" da planilha original.
const CAMPOS = [
    { chave: "rendaBrutaMes", label: "Renda bruta (mês)", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "gastoSaudeMes", label: "Gasto com saúde (mês)", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "qtdDependentes", label: "Quantidade de dependentes", unidade: "numero", placeholder: "0" },
    { chave: "gastoEducPropriaMes", label: "Gasto com educação própria (mês)", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "gastoEducDependentesMes", label: "Gasto com educação dos dependentes (mês)", unidade: "reais", placeholder: "R$ 0,00" },
    { chave: "contribuicaoPgblAno", label: "Contribuição no PGBL (total no ano)", unidade: "reais", placeholder: "R$ 0,00" },
];

// Linhas da tabela comparativa (rótulo + chave em calcularPrevidenciaria().linhas).
// `grupoAntes` insere uma linha divisória (só rótulo, sem valores) antes da
// linha — agrupa visualmente "o que foi deduzido" separado do "resultado".
const LINHAS_TABELA = [
    { chave: "rendaAnual", label: "Renda Anual" },
    { chave: "inss", label: "Dedução INSS", grupoAntes: "Deduções aplicadas" },
    { chave: "saude", label: "Dedução Saúde" },
    { chave: "eduPropria", label: "Dedução Educação própria" },
    { chave: "eduDependentes", label: "Dedução Educação dependentes" },
    { chave: "dependentes", label: "Dedução Dependentes" },
    { chave: "pgbl", label: "Dedução PGBL" },
    { chave: "deducaoSimplificada", label: "Dedução Simplificada" },
    { chave: "baseTributavel", label: "Base Tributável", total: true, grupoAntes: "Resultado" },
    { chave: "irFonte", label: "IR na Fonte" },
    { chave: "irDevido", label: "IR Devido", total: true },
    { chave: "restituicao", label: "Restituição / a pagar", total: true },
];

const COLUNAS = [
    { chave: "simplificada", label: "Simplificada" },
    { chave: "completa", label: "Completa" },
    { chave: "completaPgbl12", label: "Completa c/ 12% PGBL", recomendada: true },
];

let refs = {};

function chaveEstado(chave) {
    return `prev.${chave}`;
}

function anoTabelaAtual() {
    return state.prefs.tabelaPrevidenciaria === "2022" ? 2022 : 2026;
}

function formatarSeReais(texto) {
    if (!texto) return "";
    const valor = parseNumeroPtBR(texto);
    return valor !== null ? formatarReais(valor) : texto;
}

export function mount(container, ctx) {
    // clear() esvazia o container e reseta o scroll pro topo (mesmo problema
    // documentado em tabs/emails.js:renderBlocos) — preserva a posição pra
    // não "puxar pra cima" a cada remontagem (ex.: Confirmar em Configurações).
    const scrollAnterior = container.scrollTop;
    clear(container);
    refs = { outputs: {} };

    container.appendChild(montarToolbarLimpar(state.previdenciaria, () => mount(container, ctx)));

    const painelInputs = el("div", { class: "ccard prev-inputs" });
    painelInputs.appendChild(el("h4", { text: "Premissas" }));

    for (const campo of CAMPOS) {
        const linhaEl = el("div", { class: "crow" });
        linhaEl.appendChild(el("span", { class: "clbl", text: campo.label }));

        const input = el("input", { class: "cinput", type: "text", placeholder: campo.placeholder });
        const valorSalvo = state.previdenciaria[chaveEstado(campo.chave)] || "";
        input.value = campo.unidade === "reais" ? formatarSeReais(valorSalvo) : valorSalvo;
        input.addEventListener("input", () => {
            if (campo.unidade === "reais") {
                const distanciaDoFim = input.value.length - input.selectionStart;
                input.value = formatarMilharEnquantoDigita(input.value);
                const pos = Math.max(0, input.value.length - distanciaDoFim);
                input.setSelectionRange(pos, pos);
            }
            state.previdenciaria[chaveEstado(campo.chave)] = input.value;
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
                    state.previdenciaria[chaveEstado(campo.chave)] = input.value;
                }
            });
        }
        linhaEl.appendChild(el("div", { class: "cwrap" }, [input]));
        painelInputs.appendChild(linhaEl);
    }

    refs.diagPgbl = el("div", { class: "prev-diag" });
    painelInputs.appendChild(refs.diagPgbl);

    const painelResultado = el("div", { class: "prev-resultado" });

    const tabela = el("table", { class: "dtable prev-table" });
    const thead = el("thead", {}, [
        el(
            "tr",
            {},
            [el("th", { text: "Cálculos" })].concat(
                COLUNAS.map((c) => {
                    const th = el("th", {});
                    th.appendChild(document.createTextNode(c.label));
                    if (c.recomendada) th.appendChild(el("span", { class: "th-tag", text: "Recomendada" }));
                    return th;
                })
            )
        ),
    ]);
    const tbody = el("tbody");
    for (const linha of LINHAS_TABELA) {
        if (linha.grupoAntes) {
            tbody.appendChild(el("tr", { class: "grp" }, [el("td", { colspan: String(COLUNAS.length + 1), text: linha.grupoAntes })]));
        }
        const tr = el("tr", { class: linha.total ? "prev-total" : "" });
        tr.appendChild(el("td", { class: "prev-lbl", text: linha.label }));
        refs.outputs[linha.chave] = {};
        for (const coluna of COLUNAS) {
            const td = el("td", { class: "prev-val" });
            refs.outputs[linha.chave][coluna.chave] = td;
            tr.appendChild(td);
        }
        tbody.appendChild(tr);
    }
    tabela.appendChild(thead);
    tabela.appendChild(tbody);
    const tabelaWrap = el("div", { class: "comp-table-wrap" }, [tabela]);
    painelResultado.appendChild(tabelaWrap);

    refs.diagEconomia = el("div", { class: "prev-diag prev-diag-economia" });
    painelResultado.appendChild(refs.diagEconomia);

    refs.avisoTabela = el("div", { class: "warn prev-aviso" });
    refs.avisoTabela.insertAdjacentHTML("beforeend", icons.iconAlerta);
    refs.avisoTexto = el("span");
    refs.avisoTabela.appendChild(refs.avisoTexto);
    painelResultado.appendChild(refs.avisoTabela);

    container.appendChild(el("div", { class: "prev" }, [painelInputs, painelResultado]));

    atualizar();
    container.scrollTop = scrollAnterior;
}

function atualizar() {
    const valores = {};
    for (const campo of CAMPOS) {
        const texto = state.previdenciaria[chaveEstado(campo.chave)] || "";
        valores[campo.chave] = parseNumeroPtBR(texto) ?? 0;
    }

    const resultado = calcularPrevidenciaria(valores, anoTabelaAtual());

    for (const linha of LINHAS_TABELA) {
        for (const coluna of COLUNAS) {
            const valor = resultado.linhas[linha.chave][coluna.chave];
            refs.outputs[linha.chave][coluna.chave].textContent = formatarReais(valor);
        }
    }

    refs.diagPgbl.textContent = `Limite dedutível de PGBL: ${formatarReais(resultado.limitePgbl)}. ${resultado.diagnosticoPgbl}`;
    refs.diagEconomia.textContent = resultado.diagnosticoEconomia;

    if (resultado.tabela.aviso) {
        refs.avisoTexto.textContent = resultado.tabela.aviso;
        refs.avisoTabela.style.display = "flex";
    } else {
        refs.avisoTabela.style.display = "none";
    }
}
