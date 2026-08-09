import { state } from "../state.js";
import { el, clear, montarToolbarLimpar } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import {
    parseNumeroPtBR,
    formatarReais,
    formatarPercentual,
    formatarNumero,
    formatarMilharEnquantoDigita,
    formatarPercentualEnquantoDigita,
} from "../util/numeros.js";
import {
    valorFuturo,
    periodosParaValorFuturo,
    aporteParaValorFuturo,
    taxaParaValorFuturo,
    taxaEquivalente,
    taxaReal,
    aliquotaIR,
} from "../util/financeiro.js";
import { renderCalculadoraRendaFixa } from "./calcRendaFixa.js";

// Cada cartão é config-driven: `linhas` na ordem exata de exibição (isso já
// resolve os cartões "bidirecionais" (composição/transformação de taxas),
// que intercalam input/output/input/output em vez de agrupar todos os
// inputs primeiro). `total: true` numa linha de output marca a transição
// visual entre os campos de entrada e o resultado final (linha com borda
// em cima) — só a última "leva" de resultados de cada cartão carrega isso.
// `calcular(v)` recebe os valores atuais dos campos `tipo:"input"` (já
// convertidos pra número, 0 quando vazio) e devolve
// {chaveDoOutput: valorNumerico}.
const CARTOES = [
    {
        id: "tempo",
        titulo: "Quanto tempo para atingir certo valor?",
        linhas: [
            { chave: "aporteInicial", label: "Aporte inicial", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "aportesMensais", label: "Aportes mensais", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "valorAlmejado", label: "Valor almejado", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "rendimento", label: "Rendimento", tipo: "input", unidade: "percentual", placeholder: "0,00", sufixoTexto: "ao mês" },
            { chave: "tempoMeses", label: "Tempo necessário", tipo: "output", unidade: "numero", sufixoTexto: "meses", total: true },
            { chave: "tempoAnos", label: "", tipo: "output", unidade: "numero", sufixoTexto: "anos" },
        ],
        calcular(v) {
            const n = periodosParaValorFuturo(v.aporteInicial, v.aportesMensais, v.rendimento / 100, v.valorAlmejado);
            if (n === null) return {};
            return { tempoMeses: n, tempoAnos: n / 12 };
        },
    },
    {
        id: "aporte",
        titulo: "Quanto preciso aplicar por mês para atingir X em Y prazo?",
        linhas: [
            { chave: "aporteInicial", label: "Aporte inicial", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "valorAlmejado", label: "Valor almejado", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "rendimento", label: "Rendimento", tipo: "input", unidade: "percentual", placeholder: "0,00", sufixoTexto: "ao mês" },
            { chave: "tempoDesejado", label: "Tempo desejado", tipo: "input", unidade: "numero", placeholder: "0", sufixoTexto: "meses" },
            { chave: "aportesMensais", label: "Aportes mensais", tipo: "output", unidade: "reais", total: true },
        ],
        calcular(v) {
            if (v.tempoDesejado === 0) return {};
            const pmt = aporteParaValorFuturo(v.aporteInicial, v.rendimento / 100, v.tempoDesejado, v.valorAlmejado);
            if (pmt === null || !Number.isFinite(pmt)) return {};
            return { aportesMensais: pmt };
        },
    },
    {
        id: "rentabilidadeBuscar",
        titulo: "Qual rentabilidade tenho que buscar?",
        linhas: [
            { chave: "aporteInicial", label: "Aporte inicial", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "aportesAnuais", label: "Aportes anuais", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "valorAlmejado", label: "Valor almejado", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "tempoDesejado", label: "Tempo desejado", tipo: "input", unidade: "numero", placeholder: "0", sufixoTexto: "anos" },
            { chave: "taxaAno", label: "Taxa", tipo: "output", unidade: "percentual", sufixoTexto: "ao ano", total: true },
            { chave: "taxaMes", label: "", tipo: "output", unidade: "percentual", sufixoTexto: "ao mês" },
        ],
        calcular(v) {
            if (v.valorAlmejado === 0 || v.tempoDesejado === 0) return {};
            const i = taxaParaValorFuturo(v.aporteInicial, v.aportesAnuais, v.tempoDesejado, v.valorAlmejado);
            if (i === null) return {};
            return { taxaAno: i * 100, taxaMes: taxaEquivalente(i, 12, 1) * 100 };
        },
    },
    {
        id: "rentabilidadeFutura",
        titulo: "Qual rentabilidade futura?",
        linhas: [
            { chave: "aplicacao", label: "Aplicação", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "taxaMes", label: "Taxa ao mês", tipo: "input", unidade: "percentual", placeholder: "0,00" },
            { chave: "prazoMeses", label: "Prazo", tipo: "input", unidade: "numero", placeholder: "0", sufixoTexto: "meses" },
            { chave: "valorBruto", label: "Valor bruto", tipo: "output", unidade: "reais", total: true },
            { chave: "valorLiquido", label: "Valor líquido", tipo: "output", unidade: "reais" },
        ],
        calcular(v) {
            const bruto = valorFuturo(v.aplicacao, 0, v.taxaMes / 100, v.prazoMeses);
            const ganho = bruto - v.aplicacao;
            const ir = aliquotaIR(v.prazoMeses);
            return { valorBruto: bruto, valorLiquido: bruto - ganho * ir };
        },
    },
    {
        id: "rendimentoFuturo",
        titulo: "Quanto vou ter de rendimento no futuro?",
        linhas: [
            { chave: "aporteInicial", label: "Aporte inicial", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "aportesMensais", label: "Aportes mensais", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "rendimento", label: "Rendimento", tipo: "input", unidade: "percentual", placeholder: "0,00", sufixoTexto: "ao mês" },
            { chave: "tempo", label: "Tempo", tipo: "input", unidade: "numero", placeholder: "0", sufixoTexto: "meses" },
            { chave: "valorFuturoCalc", label: "Valor futuro", tipo: "output", unidade: "reais", total: true },
        ],
        calcular(v) {
            return { valorFuturoCalc: valorFuturo(v.aporteInicial, v.aportesMensais, v.rendimento / 100, v.tempo) };
        },
    },
    {
        id: "fgc",
        titulo: "Cobertura do FGC",
        linhas: [
            { chave: "taxaAno", label: "Taxa contratada", tipo: "input", unidade: "percentual", placeholder: "0,00", sufixoTexto: "ao ano" },
            { chave: "valorAplicado", label: "Valor a aplicar", tipo: "input", unidade: "reais", placeholder: "R$ 0,00" },
            { chave: "prazoMeses", label: "Prazo", tipo: "input", unidade: "numero", placeholder: "0", sufixoTexto: "meses", switchPeriodo: true },
            { chave: "valorFinal", label: "Valor no vencimento", tipo: "output", unidade: "reais", total: true },
            { chave: "statusFgc", label: "Limite FGC (R$ 250 mil)", tipo: "output", unidade: "texto" },
        ],
        calcular(v) {
            const taxaMes = taxaEquivalente(v.taxaAno / 100, 12, 1);
            const final = valorFuturo(v.valorAplicado, 0, taxaMes, v.prazoMeses);
            return { valorFinal: final, statusFgc: final > 250000 ? "Over FGC" : "OK" };
        },
    },
    {
        id: "taxaReal",
        titulo: "Cálculo da Taxa Real",
        linhas: [
            { chave: "taxaAno", label: "Taxa ao ano", tipo: "input", unidade: "percentual", placeholder: "0,00" },
            { chave: "inflacao", label: "Inflação (12 meses)", tipo: "input", unidade: "percentual", placeholder: "0,00" },
            { chave: "taxaRealAno", label: "Taxa real ao ano", tipo: "output", unidade: "percentual", total: true },
            { chave: "taxaRealMes", label: "Taxa real ao mês", tipo: "output", unidade: "percentual" },
        ],
        calcular(v) {
            const real = taxaReal(v.taxaAno / 100, v.inflacao / 100);
            return { taxaRealAno: real * 100, taxaRealMes: taxaEquivalente(real, 12, 1) * 100 };
        },
    },
    {
        id: "composicaoTaxas",
        titulo: "Composição de Taxas",
        linhas: [
            { chave: "taxaAnoParaPeriodo", label: "Taxa ao ano", tipo: "input", unidade: "percentual", placeholder: "0,00" },
            { chave: "taxaNoPeriodoCalc", label: "Taxa no período", tipo: "output", unidade: "percentual" },
            { chave: "periodoAnos", label: "Período (anos)", tipo: "input", unidade: "numero", placeholder: "0" },
            { chave: "taxaPeriodoParaAno", label: "Taxa no período", tipo: "input", unidade: "percentual", placeholder: "0,00" },
            { chave: "taxaAnoCalc", label: "Taxa ao ano", tipo: "output", unidade: "percentual", total: true },
        ],
        calcular(v) {
            if (v.periodoAnos <= 0) return {};
            return {
                taxaNoPeriodoCalc: taxaEquivalente(v.taxaAnoParaPeriodo / 100, 1, v.periodoAnos) * 100,
                taxaAnoCalc: taxaEquivalente(v.taxaPeriodoParaAno / 100, v.periodoAnos, 1) * 100,
            };
        },
    },
    {
        id: "transformacaoTaxas",
        titulo: "Transformação de Taxas",
        linhas: [
            { chave: "taxaAnoParaMes", label: "Taxa ao ano", tipo: "input", unidade: "percentual", placeholder: "0,00" },
            { chave: "taxaMesCalc", label: "Taxa ao mês", tipo: "output", unidade: "percentual" },
            { chave: "taxaMesParaAno", label: "Taxa ao mês", tipo: "input", unidade: "percentual", placeholder: "0,00" },
            { chave: "taxaAnoCalc", label: "Taxa ao ano", tipo: "output", unidade: "percentual", total: true },
        ],
        calcular(v) {
            return {
                taxaMesCalc: taxaEquivalente(v.taxaAnoParaMes / 100, 12, 1) * 100,
                taxaAnoCalc: taxaEquivalente(v.taxaMesParaAno / 100, 1, 12) * 100,
            };
        },
    },
];

const CATEGORIAS = [
    { titulo: "Planejamento de metas", icon: icons.iconRelogio, ids: ["tempo", "aporte", "rentabilidadeBuscar"] },
    { titulo: "Projeção de valores", icon: icons.iconRentabilidade, ids: ["rentabilidadeFutura", "rendimentoFuturo", "fgc"] },
    { titulo: "Conversão de taxas", icon: icons.iconLista, ids: ["taxaReal", "composicaoTaxas", "transformacaoTaxas"] },
    { titulo: "Calculadora de Renda Fixa", icon: icons.iconTabela, render: renderCalculadoraRendaFixa },
];

let refs = {};

export function mount(container, ctx) {
    clear(container);
    refs = {};

    // Limpa os cartões desta aba (CARTOES) e a Calculadora de Renda Fixa
    // (state.calcRF), que vive na mesma tela mas guarda estado à parte.
    container.appendChild(montarToolbarLimpar([state.calculadora, state.calcRF], () => mount(container, ctx)));

    const cartaoPorId = Object.fromEntries(CARTOES.map((c) => [c.id, c]));

    for (const categoria of CATEGORIAS) {
        const catH = el("div", { class: "cat-h" });
        catH.insertAdjacentHTML("beforeend", categoria.icon);
        catH.appendChild(document.createTextNode(categoria.titulo));
        catH.appendChild(el("span", { class: "line" }));

        // Categorias com `render()` fogem do formato input/output linha-a-
        // linha do CARTOES (ex.: Comparador de Renda Fixa, com campos de
        // texto/select/data e layout pareado) — montam seu próprio corpo.
        let corpo;
        if (categoria.render) {
            corpo = categoria.render();
        } else {
            corpo = el("div", { class: "calc-grid" });
            for (const id of categoria.ids) {
                corpo.appendChild(renderCartao(cartaoPorId[id]));
            }
        }

        container.appendChild(el("div", { class: "cat" }, [catH, corpo]));
    }
}

function chaveEstado(cartaoId, chave) {
    return `${cartaoId}.${chave}`;
}

// Campos com switchPeriodo (ex.: "Prazo" da calculadora de FGC) guardam o
// número digitado sempre em meses (mesma unidade que calcular() espera);
// só a unidade escolhida (M/A) muda como ele é digitado/exibido.
function unidadePeriodo(cartaoId, chave) {
    return state.calculadora[`${chaveEstado(cartaoId, chave)}.unidade`] || "M";
}

function formatarValor(valor, unidade, sufixo) {
    if (valor === undefined || valor === null || !Number.isFinite(valor)) return "";
    const texto = unidade === "reais" ? formatarReais(valor) : unidade === "percentual" ? formatarPercentual(valor) : formatarNumero(valor);
    return sufixo ? `${texto} ${sufixo}` : texto;
}

// Formata um valor salvo (de state.calculadora) como "R$ ..." se der pra
// interpretar como número; devolve o texto original se não der (campo
// vazio, ou ainda em edição quando a aba foi montada de novo).
function formatarSeReais(texto) {
    if (!texto) return "";
    const valor = parseNumeroPtBR(texto);
    return valor !== null ? formatarReais(valor) : texto;
}

function renderCartao(cartao) {
    const card = el("div", { class: "ccard" });
    card.appendChild(el("h4", { text: cartao.titulo }));
    refs[cartao.id] = { outputs: {} };

    const atualizar = () => {
        const valores = {};
        for (const linha of cartao.linhas) {
            if (linha.tipo !== "input") continue;
            const texto = state.calculadora[chaveEstado(cartao.id, linha.chave)] || "";
            let valor = parseNumeroPtBR(texto) ?? 0;
            if (linha.switchPeriodo && unidadePeriodo(cartao.id, linha.chave) === "A") valor *= 12;
            valores[linha.chave] = valor;
        }
        const saida = cartao.calcular(valores);
        for (const linha of cartao.linhas) {
            if (linha.tipo !== "output") continue;
            const spanEl = refs[cartao.id].outputs[linha.chave];
            if (linha.unidade === "texto") {
                const valor = saida[linha.chave] || "";
                spanEl.textContent = valor;
                spanEl.classList.toggle("neg", valor === "Over FGC");
                spanEl.classList.toggle("pos", valor === "OK");
            } else {
                spanEl.textContent = formatarValor(saida[linha.chave], linha.unidade, linha.sufixoTexto);
            }
        }
    };

    for (const linha of cartao.linhas) {
        const linhaEl = el("div", { class: `crow${linha.tipo === "output" && linha.total ? " total" : ""}` });
        const labelEl = el("span", { class: "clbl", text: linha.label });
        let labelWrap = null;
        if (linha.switchPeriodo) {
            labelWrap = el("div", { class: "clbl-wrap" }, [labelEl]);
            linhaEl.appendChild(labelWrap);
        } else {
            linhaEl.appendChild(labelEl);
        }

        if (linha.tipo === "input") {
            const input = el("input", { class: "cinput", type: "text", placeholder: linha.placeholder || "" });
            const valorSalvo = state.calculadora[chaveEstado(cartao.id, linha.chave)] || "";
            input.value = linha.unidade === "reais" ? formatarSeReais(valorSalvo) : valorSalvo;
            input.addEventListener("input", () => {
                if (linha.unidade === "reais" || linha.unidade === "percentual") {
                    const distanciaDoFim = input.value.length - input.selectionStart;
                    input.value =
                        linha.unidade === "reais" ? formatarMilharEnquantoDigita(input.value) : formatarPercentualEnquantoDigita(input.value);
                    const pos = Math.max(0, input.value.length - distanciaDoFim);
                    input.setSelectionRange(pos, pos);
                }
                state.calculadora[chaveEstado(cartao.id, linha.chave)] = input.value;
                atualizar();
            });
            if (linha.unidade === "reais") {
                // Ao focar (pra editar), tira o "R$ " do início — mais fácil
                // de selecionar/apagar os dígitos. Ao sair do campo,
                // reformata como "R$ 1.234,56" se der pra interpretar como
                // número.
                input.addEventListener("focus", () => {
                    input.value = input.value.replace(/^R\$\s*/, "");
                });
                input.addEventListener("blur", () => {
                    const valor = parseNumeroPtBR(input.value);
                    if (valor !== null) {
                        input.value = formatarReais(valor);
                        state.calculadora[chaveEstado(cartao.id, linha.chave)] = input.value;
                    }
                });
            }
            const wrap = el("div", { class: "cwrap" }, [input]);
            let sufixoEl = null;
            if (linha.switchPeriodo) {
                sufixoEl = el("span", { class: "csuffix", text: unidadePeriodo(cartao.id, linha.chave) === "A" ? "anos" : "meses" });
                wrap.appendChild(sufixoEl);
                const switchEl = el("div", { class: "period-switch" });
                labelWrap.appendChild(switchEl);
                const botaoM = el("button", { type: "button", class: "period-btn", text: "M" });
                const botaoA = el("button", { type: "button", class: "period-btn", text: "A" });
                const marcarAtivo = () => {
                    const atual = unidadePeriodo(cartao.id, linha.chave);
                    botaoM.classList.toggle("on", atual === "M");
                    botaoA.classList.toggle("on", atual === "A");
                    sufixoEl.textContent = atual === "A" ? "anos" : "meses";
                };
                const selecionar = (u) => {
                    const atual = unidadePeriodo(cartao.id, linha.chave);
                    if (atual === u) return;
                    // Converte o número já digitado pra unidade nova, em vez
                    // de só reinterpretar os mesmos dígitos (24 meses -> 2
                    // anos, não 24 anos).
                    const valorAtual = parseNumeroPtBR(input.value);
                    if (valorAtual !== null) {
                        const convertido = u === "A" ? valorAtual / 12 : valorAtual * 12;
                        const textoConvertido = (Math.round(convertido * 100) / 100).toString().replace(".", ",");
                        input.value = textoConvertido;
                        state.calculadora[chaveEstado(cartao.id, linha.chave)] = textoConvertido;
                    }
                    state.calculadora[`${chaveEstado(cartao.id, linha.chave)}.unidade`] = u;
                    marcarAtivo();
                    atualizar();
                };
                botaoM.addEventListener("click", () => selecionar("M"));
                botaoA.addEventListener("click", () => selecionar("A"));
                marcarAtivo();
                switchEl.appendChild(botaoM);
                switchEl.appendChild(botaoA);
            } else if (linha.sufixoTexto) {
                wrap.appendChild(el("span", { class: "csuffix", text: linha.sufixoTexto }));
            }
            linhaEl.appendChild(wrap);
        } else {
            const saida = el("span", { class: "cout" });
            refs[cartao.id].outputs[linha.chave] = saida;
            linhaEl.appendChild(saida);
        }

        card.appendChild(linhaEl);
    }

    atualizar();
    return card;
}
