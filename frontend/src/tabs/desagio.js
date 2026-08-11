import { state, novoBlocoId } from "../state.js";
import { el, clear, btn } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import {
    parseNumeroPtBR,
    formatarReais,
    formatarPercentual,
    formatarMilharEnquantoDigita,
    formatarPercentualEnquantoDigita,
} from "../util/numeros.js";
import { SalvarImagemPNG } from "../../wailsjs/go/main/App.js";

let refs = {};
let ctxApp = null;
let criterioOrdem = "desagio"; // "nome" | "valorAtual" | "desagio"
let ordemAsc = false; // true: crescente (A→Z / menor→maior); false: decrescente
let mostrarSoma = false;
// ids de state.blocosDesagio na ordem exibida na tabela — só recalculada
// nos gatilhos explícitos abaixo (nunca a cada tecla), pra a linha não
// "pular" embaixo do cursor enquanto o campo que a está editando ainda
// está focado. Ver recalcularOrdemSeMudou().
let ordemVisivel = [];

const COLUNAS_FIXAS = [
    { chave: "titulo", nome: "Título" },
    { chave: "valorAtual", nome: "Valor Atual" },
    { chave: "valorSaida", nome: "Valor de Saída" },
    { chave: "desagioReais", nome: "Deságio (R$)" },
    { chave: "desagioPct", nome: "Deságio (%)" },
];
const COR_CABECALHO = "#047857"; // --primary-600 do acento Esmeralda (canvas de exportação não lê variável CSS)

// Formatação de uma coluna extra — "texto" não mexe no que foi digitado;
// "reais"/"percentual" reusam a mesma máscara pt-BR (milhar ao digitar,
// "R$"/"%" no lugar certo) já usada nos campos de calculadora do app (ver
// frontend/src/util/rendaFixa.js), só que aplicada a uma célula da tabela.
const TIPOS_COLUNA = [
    { id: "texto", label: "T" },
    { id: "reais", label: "$" },
    { id: "percentual", label: "%" },
];

const CRITERIOS_ORDEM = [
    { id: "nome", label: "Nome" },
    { id: "valorAtual", label: "Tamanho da posição" },
    { id: "desagio", label: "Ágio/Deságio" },
];
// Direção padrão ao trocar de critério — cada um tem o sentido mais
// intuitivo como ponto de partida (nome começa A→Z; posição e
// ágio/deságio começam do maior pro menor).
const ORDEM_ASC_PADRAO = { nome: true, valorAtual: false, desagio: false };
const ROTULOS_ORDEM = {
    nome: { asc: "A → Z", desc: "Z → A" },
    valorAtual: { asc: "Menor → Maior", desc: "Maior → Menor" },
    desagio: { asc: "Deságio → Ágio", desc: "Ágio → Deságio" },
};

export function mount(container, ctx) {
    clear(container);
    refs = {};
    ctxApp = ctx;
    criterioOrdem = "desagio";
    ordemAsc = false;
    mostrarSoma = false;

    const painel = el("div", { class: "des-r" });

    refs.selectOrdem = el("select", { class: "select" });
    for (const c of CRITERIOS_ORDEM) {
        const opt = el("option", { value: c.id, text: c.label });
        if (c.id === criterioOrdem) opt.selected = true;
        refs.selectOrdem.appendChild(opt);
    }
    refs.selectOrdem.addEventListener("change", () => {
        criterioOrdem = refs.selectOrdem.value;
        ordemAsc = ORDEM_ASC_PADRAO[criterioOrdem];
        atualizarLabelOrdem();
        recalcularOrdem();
        renderCorpo();
    });
    refs.btnOrdem = btn("Ágio → Deságio", { icon: icons.iconOrdenar, onClick: alternarOrdem });
    refs.btnSoma = btn("Mostrar soma", { onClick: alternarSoma });
    const controles = el("div", { class: "des-controles" }, [refs.selectOrdem, refs.btnOrdem, refs.btnSoma]);

    painel.appendChild(el("div", { class: "mail-r-h" }, [el("h3", { text: "Tabela de deságio" }), controles]));

    refs.tableWrap = el("div", { class: "table-card des-table-wrap" });
    refs.table = el("table", { class: "dtable des-table" });
    refs.thead = el("thead");
    refs.tbody = el("tbody");
    refs.table.appendChild(refs.thead);
    refs.table.appendChild(refs.tbody);
    refs.tableWrap.appendChild(refs.table);
    painel.appendChild(refs.tableWrap);

    painel.appendChild(
        el("div", { class: "des-foot" }, [
            btn("Copiar imagem", { classe: "pri", icon: icons.iconImagem, onClick: copiarImagem }),
            btn("Salvar imagem", { icon: icons.iconSalvar, onClick: salvarImagem }),
        ])
    );

    container.appendChild(painel);

    if (state.blocosDesagio.length === 0) {
        state.blocosDesagio.push(novaLinha());
    }
    atualizarLabelOrdem();
    atualizarLabelSoma();
    recalcularOrdem();
    renderCabecalho();
    renderCorpo();
}

function novaLinha() {
    return { id: novoBlocoId(), titulo: "", valorAtual: "", valorSaida: "", extras: {} };
}

function atualizarLabelOrdem() {
    if (!refs.btnOrdem) return;
    const rotulos = ROTULOS_ORDEM[criterioOrdem];
    refs.btnOrdem.lastChild.textContent = ordemAsc ? rotulos.asc : rotulos.desc;
}

function alternarOrdem() {
    ordemAsc = !ordemAsc;
    atualizarLabelOrdem();
    recalcularOrdem();
    renderCorpo();
}

function atualizarLabelSoma() {
    if (!refs.btnSoma) return;
    refs.btnSoma.lastChild.textContent = mostrarSoma ? "Ocultar soma" : "Mostrar soma";
}

function alternarSoma() {
    mostrarSoma = !mostrarSoma;
    atualizarLabelSoma();
    renderCorpo();
}

// ---------------------------------------------------------------------------
// Layout de colunas: as 5 fixas mantêm ordem relativa entre si; cada coluna
// extra (state.colunasExtrasDesagio) se insere logo depois da chave guardada
// em `apos` — é assim que o botão "+" de cada cabeçalho escolhe onde a nova
// coluna nasce, sem precisar de um array de ordem separado pra mexer.
// ---------------------------------------------------------------------------
function layoutColunas() {
    const ordem = COLUNAS_FIXAS.map((c) => c.chave);
    const restantes = state.colunasExtrasDesagio.slice();
    let progresso = true;
    while (restantes.length && progresso) {
        progresso = false;
        for (let i = restantes.length - 1; i >= 0; i--) {
            const idx = ordem.indexOf(restantes[i].apos);
            if (idx !== -1) {
                ordem.splice(idx + 1, 0, restantes[i].id);
                restantes.splice(i, 1);
                progresso = true;
            }
        }
    }
    for (const r of restantes) ordem.push(r.id); // apos órfão (não deveria acontecer) — manda pro fim
    return ordem;
}

function colunaInfo(chave) {
    const fixa = COLUNAS_FIXAS.find((c) => c.chave === chave);
    if (fixa) return fixa;
    return state.colunasExtrasDesagio.find((c) => c.id === chave);
}

// Deságio = diferença entre o valor de saída e o valor atual, em R$ e em %
// relativo ao valor atual — negativo quando o valor de saída é menor que o
// atual (deságio de verdade), positivo quando é maior (ágio).
function camposCalculados(linha) {
    const valorAtual = parseNumeroPtBR(linha.valorAtual);
    const valorSaida = parseNumeroPtBR(linha.valorSaida);
    let desagioReais = null;
    let desagioPct = null;
    if (valorAtual !== null && valorSaida !== null) {
        desagioReais = valorSaida - valorAtual;
        desagioPct = valorAtual !== 0 ? (desagioReais / valorAtual) * 100 : null;
    }
    return { valorAtual, valorSaida, desagioReais, desagioPct };
}

function compararLinhas(a, b) {
    const ca = camposCalculados(a);
    const cb = camposCalculados(b);
    let cmp;
    if (criterioOrdem === "nome") {
        cmp = (a.titulo || "").localeCompare(b.titulo || "", "pt-BR", { sensitivity: "base" });
    } else if (criterioOrdem === "valorAtual") {
        if (ca.valorAtual === null && cb.valorAtual === null) cmp = 0;
        else if (ca.valorAtual === null) cmp = 1;
        else if (cb.valorAtual === null) cmp = -1;
        else cmp = ca.valorAtual - cb.valorAtual;
    } else {
        if (ca.desagioPct === null && cb.desagioPct === null) cmp = 0;
        else if (ca.desagioPct === null) cmp = 1;
        else if (cb.desagioPct === null) cmp = -1;
        else cmp = ca.desagioPct - cb.desagioPct;
    }
    return ordemAsc ? cmp : -cmp;
}

function recalcularOrdem() {
    const copia = state.blocosDesagio.slice().sort(compararLinhas);
    ordemVisivel = copia.map((l) => l.id);
}

// Chamada no blur de qualquer campo de uma linha — só refaz o corpo da
// tabela se a nova ordenação realmente mudar a posição de alguma linha,
// pra não perder o scroll/estado à toa numa edição que não afeta o critério
// de ordenação corrente.
function recalcularOrdemSeMudou() {
    const antes = ordemVisivel.join(",");
    recalcularOrdem();
    if (ordemVisivel.join(",") !== antes) renderCorpo();
}

function totais() {
    let va = 0;
    let vs = 0;
    for (const linha of state.blocosDesagio) {
        const c = camposCalculados(linha);
        // Título sem os dois valores preenchidos ainda não tem deságio
        // calculado (camposCalculados devolve null) — fica de fora da soma
        // em vez de entrar como valorSaida=0, que inflaria o total como se
        // aquele título tivesse perdido 100% do valor.
        if (c.valorAtual === null || c.valorSaida === null) continue;
        va += c.valorAtual;
        vs += c.valorSaida;
    }
    const dr = vs - va;
    const dp = va !== 0 ? (dr / va) * 100 : null;
    return { valorAtual: va, valorSaida: vs, desagioReais: dr, desagioPct: dp };
}

// ---------------------------------------------------------------------------
// Cabeçalho
// ---------------------------------------------------------------------------
function iconeCabecalho(icone, titulo, onClick, classeExtra = "") {
    const botao = el("button", { class: `des-th-icon ${classeExtra}`.trim(), title: titulo, type: "button" });
    botao.insertAdjacentHTML("beforeend", icone);
    botao.addEventListener("click", onClick);
    return botao;
}

function renderCabecalho() {
    const tr = el("tr");
    for (const chave of layoutColunas()) {
        const info = colunaInfo(chave);
        const th = el("th");
        const linhaTh = el("div", { class: "des-th-row" });

        if (info.apos !== undefined) {
            const input = el("input", { class: "des-colname-input", type: "text", placeholder: "Nome da coluna" });
            input.value = info.nome;
            input.addEventListener("input", () => {
                info.nome = input.value;
            });
            linhaTh.appendChild(input);

            const selectTipo = el("select", { class: "des-th-tipo", title: "Formato dos valores desta coluna" });
            for (const t of TIPOS_COLUNA) {
                const opt = el("option", { value: t.id, text: t.label });
                if (t.id === info.tipo) opt.selected = true;
                selectTipo.appendChild(opt);
            }
            selectTipo.addEventListener("change", () => {
                info.tipo = selectTipo.value;
                renderCorpo();
            });
            linhaTh.appendChild(selectTipo);

            linhaTh.appendChild(iconeCabecalho(icons.iconFechar, "Remover coluna", () => removerColuna(info.id), "des-th-icon-rm"));
        } else {
            linhaTh.appendChild(el("span", { class: "des-th-name", text: info.nome }));
        }

        linhaTh.appendChild(iconeCabecalho(icons.iconMais, "Adicionar coluna depois desta", () => adicionarColuna(chave)));
        th.appendChild(linhaTh);
        tr.appendChild(th);
    }
    tr.appendChild(el("th", { class: "des-th-acoes" })); // coluna da lixeira de linha
    clear(refs.thead);
    refs.thead.appendChild(tr);
}

function adicionarColuna(aposChave) {
    const id = "extra" + novoBlocoId();
    state.colunasExtrasDesagio.push({ id, nome: "Nova coluna", apos: aposChave, tipo: "texto" });
    renderCabecalho();
    renderCorpo();
    const posicao = layoutColunas().indexOf(id);
    const th = refs.thead.querySelector(`tr th:nth-child(${posicao + 1})`);
    const input = th?.querySelector(".des-colname-input");
    if (input) {
        input.focus();
        input.select();
    }
}

function removerColuna(id) {
    const col = state.colunasExtrasDesagio.find((c) => c.id === id);
    if (!col) return;
    const temDado = state.blocosDesagio.some((l) => (l.extras[id] || "").trim() !== "");
    if (temDado && !confirm(`Remover a coluna "${col.nome}"? Isso apaga o valor dela em todas as linhas.`)) return;
    state.colunasExtrasDesagio = state.colunasExtrasDesagio.filter((c) => c.id !== id);
    for (const linha of state.blocosDesagio) delete linha.extras[id];
    renderCabecalho();
    renderCorpo();
}

// ---------------------------------------------------------------------------
// Corpo
// ---------------------------------------------------------------------------
function celulaEditavel(valor, placeholder, aoDigitar, aoSair, tipo = "texto") {
    const input = el("input", { class: "des-cell-input", type: "text", placeholder: placeholder || "" });
    input.value = valor || "";

    input.addEventListener("input", () => {
        if (tipo === "reais" || tipo === "percentual") {
            // Mesma técnica de máscara-ao-digitar usada nos campos de
            // calculadora (ver util/rendaFixa.js): reformata só a parte
            // digitada, preservando a posição do cursor em relação ao fim.
            const distanciaDoFim = input.value.length - input.selectionStart;
            input.value = tipo === "reais" ? formatarMilharEnquantoDigita(input.value) : formatarPercentualEnquantoDigita(input.value);
            const pos = Math.max(0, input.value.length - distanciaDoFim);
            input.setSelectionRange(pos, pos);
        }
        aoDigitar(input.value);
    });
    if (tipo === "reais") {
        input.addEventListener("focus", () => {
            input.value = input.value.replace(/^R\$\s*/, "");
        });
    }
    input.addEventListener("blur", () => {
        if (tipo === "reais" || tipo === "percentual") {
            const numero = parseNumeroPtBR(input.value);
            if (numero !== null) {
                input.value = tipo === "reais" ? formatarReais(numero) : formatarPercentual(numero);
                aoDigitar(input.value);
            }
        }
        aoSair();
    });
    return el("td", {}, [input]);
}

function celulaTexto(texto, campo) {
    const td = el("td", { class: "cell-text", text: texto });
    td.dataset.field = campo;
    return td;
}

function renderLinha(linha) {
    const tr = el("tr");
    tr.dataset.id = linha.id;

    for (const chave of layoutColunas()) {
        if (chave === "titulo") {
            tr.appendChild(
                celulaEditavel(
                    linha.titulo,
                    "Nome do título",
                    (v) => {
                        linha.titulo = v;
                    },
                    recalcularOrdemSeMudou
                )
            );
        } else if (chave === "valorAtual") {
            tr.appendChild(
                celulaEditavel(
                    linha.valorAtual,
                    "Ex: 1.234,56",
                    (v) => {
                        linha.valorAtual = v;
                        atualizarComputadas(linha);
                    },
                    recalcularOrdemSeMudou,
                    "reais"
                )
            );
        } else if (chave === "valorSaida") {
            tr.appendChild(
                celulaEditavel(
                    linha.valorSaida,
                    "Ex: 1.000,00",
                    (v) => {
                        linha.valorSaida = v;
                        atualizarComputadas(linha);
                    },
                    recalcularOrdemSeMudou,
                    "reais"
                )
            );
        } else if (chave === "desagioReais" || chave === "desagioPct") {
            const c = camposCalculados(linha);
            tr.appendChild(celulaTexto(chave === "desagioReais" ? formatarValorOuTraco(c.desagioReais, formatarReais) : formatarValorOuTraco(c.desagioPct, formatarPercentual), chave));
        } else {
            const tipoColuna = colunaInfo(chave).tipo;
            const placeholder = tipoColuna === "reais" ? "Ex: 1.234,56" : tipoColuna === "percentual" ? "Ex: 12,34" : "—";
            tr.appendChild(
                celulaEditavel(
                    linha.extras[chave],
                    placeholder,
                    (v) => {
                        linha.extras[chave] = v;
                    },
                    recalcularOrdemSeMudou,
                    tipoColuna
                )
            );
        }
    }

    const tdAcao = el("td", { class: "des-rowrm" });
    const botaoRemover = el("button", { class: "des-icon-btn", title: "Remover título", type: "button" });
    botaoRemover.insertAdjacentHTML("beforeend", icons.iconLixeira);
    botaoRemover.addEventListener("click", () => removerLinha(linha.id));
    tdAcao.appendChild(botaoRemover);
    tr.appendChild(tdAcao);

    return tr;
}

function formatarValorOuTraco(valor, formatador) {
    return valor !== null ? formatador(valor) : "—";
}

// Texto de uma célula de coluna extra pra exportação — colunas "reais"/
// "percentual" já chegam formatadas na célula (reformatadas no blur, ver
// celulaEditavel), mas reformata de novo aqui por segurança (ex.: campo
// ainda focado, sem ter passado pelo blur, no instante do export).
function textoExtraFormatado(linha, chave) {
    const bruto = linha.extras[chave]?.trim();
    if (!bruto) return "—";
    const tipo = colunaInfo(chave).tipo;
    if (tipo === "reais" || tipo === "percentual") {
        const numero = parseNumeroPtBR(bruto);
        if (numero !== null) return tipo === "reais" ? formatarReais(numero) : formatarPercentual(numero);
    }
    return bruto;
}

function linhaTotal() {
    const tr = el("tr", { class: "des-total" });
    const t = totais();
    for (const chave of layoutColunas()) {
        let texto = "";
        if (chave === "titulo") texto = "Total";
        else if (chave === "valorAtual") texto = formatarReais(t.valorAtual);
        else if (chave === "valorSaida") texto = formatarReais(t.valorSaida);
        else if (chave === "desagioReais") texto = formatarReais(t.desagioReais);
        else if (chave === "desagioPct") texto = formatarValorOuTraco(t.desagioPct, formatarPercentual);
        const td = el("td", { class: "cell-text", text: texto });
        td.dataset.totalField = chave;
        tr.appendChild(td);
    }
    tr.appendChild(el("td"));
    return tr;
}

function renderCorpo() {
    clear(refs.tbody);
    for (const id of ordemVisivel) {
        const linha = state.blocosDesagio.find((l) => l.id === id);
        if (linha) refs.tbody.appendChild(renderLinha(linha));
    }
    if (mostrarSoma && state.blocosDesagio.length > 0) refs.tbody.appendChild(linhaTotal());

    const trAdd = el("tr", { class: "des-addrow" });
    const tdAdd = el("td");
    tdAdd.colSpan = layoutColunas().length + 1;
    const botaoAdd = el("button", { type: "button" });
    botaoAdd.insertAdjacentHTML("beforeend", icons.iconMais);
    botaoAdd.appendChild(document.createTextNode("Adicionar título"));
    botaoAdd.addEventListener("click", adicionarLinha);
    tdAdd.appendChild(botaoAdd);
    trAdd.appendChild(tdAdd);
    refs.tbody.appendChild(trAdd);
}

// Atualiza só as células computadas da linha editada (e a linha de Total,
// se visível) sem refazer o resto da tabela — mantém o foco e o cursor no
// campo enquanto o usuário digita.
function atualizarComputadas(linha) {
    const tr = refs.tbody.querySelector(`tr[data-id="${linha.id}"]`);
    if (tr) {
        const c = camposCalculados(linha);
        const tdReais = tr.querySelector('[data-field="desagioReais"]');
        if (tdReais) tdReais.textContent = formatarValorOuTraco(c.desagioReais, formatarReais);
        const tdPct = tr.querySelector('[data-field="desagioPct"]');
        if (tdPct) tdPct.textContent = formatarValorOuTraco(c.desagioPct, formatarPercentual);
    }
    if (mostrarSoma) {
        const trTotal = refs.tbody.querySelector("tr.des-total");
        if (trTotal) {
            const t = totais();
            trTotal.querySelector('[data-total-field="valorAtual"]').textContent = formatarReais(t.valorAtual);
            trTotal.querySelector('[data-total-field="valorSaida"]').textContent = formatarReais(t.valorSaida);
            trTotal.querySelector('[data-total-field="desagioReais"]').textContent = formatarReais(t.desagioReais);
            trTotal.querySelector('[data-total-field="desagioPct"]').textContent = formatarValorOuTraco(t.desagioPct, formatarPercentual);
        }
    }
}

function adicionarLinha() {
    state.blocosDesagio.push(novaLinha());
    recalcularOrdem();
    renderCorpo();
}

function removerLinha(id) {
    state.blocosDesagio = state.blocosDesagio.filter((b) => b.id !== id);
    recalcularOrdem();
    renderCorpo();
}

// ---------------------------------------------------------------------------
// Exportação — desenha um canvas na hora, fiel à ordem/colunas visíveis
// naquele momento na tabela, só pra gerar o PNG de Copiar/Salvar imagem.
// ---------------------------------------------------------------------------
function desenharCanvasExport() {
    const ordem = layoutColunas();
    const nomes = ordem.map((chave) => colunaInfo(chave).nome);
    const linhasTexto = ordemVisivel.map((id) => {
        const linha = state.blocosDesagio.find((l) => l.id === id);
        const c = camposCalculados(linha);
        return ordem.map((chave) => {
            if (chave === "titulo") return linha.titulo?.trim() || "—";
            if (chave === "valorAtual") return formatarValorOuTraco(c.valorAtual, formatarReais);
            if (chave === "valorSaida") return formatarValorOuTraco(c.valorSaida, formatarReais);
            if (chave === "desagioReais") return formatarValorOuTraco(c.desagioReais, formatarReais);
            if (chave === "desagioPct") return formatarValorOuTraco(c.desagioPct, formatarPercentual);
            return textoExtraFormatado(linha, chave);
        });
    });

    let indiceLinhaTotal = -1;
    if (mostrarSoma && state.blocosDesagio.length > 0) {
        const t = totais();
        linhasTexto.push(
            ordem.map((chave) => {
                if (chave === "titulo") return "Total";
                if (chave === "valorAtual") return formatarReais(t.valorAtual);
                if (chave === "valorSaida") return formatarReais(t.valorSaida);
                if (chave === "desagioReais") return formatarReais(t.desagioReais);
                if (chave === "desagioPct") return formatarValorOuTraco(t.desagioPct, formatarPercentual);
                return "";
            })
        );
        indiceLinhaTotal = linhasTexto.length - 1;
    }

    const canvas = document.createElement("canvas");
    const ctx2d = canvas.getContext("2d");

    const fonteCabecalho = "bold 14px 'Plus Jakarta Sans', sans-serif";
    const fonteCelula = "13px 'Plus Jakarta Sans', sans-serif";
    const padX = 16;
    const alturaLinha = 38;
    const alturaCabecalho = 44;

    ctx2d.font = fonteCabecalho;
    const larguras = nomes.map((n) => ctx2d.measureText(n).width + padX * 2);
    ctx2d.font = fonteCelula;
    for (const valores of linhasTexto) {
        valores.forEach((v, i) => {
            larguras[i] = Math.max(larguras[i], ctx2d.measureText(v).width + padX * 2);
        });
    }

    const largura = larguras.reduce((a, b) => a + b, 0);
    const altura = alturaCabecalho + Math.max(linhasTexto.length, 1) * alturaLinha;

    const dpr = window.devicePixelRatio || 1;
    canvas.width = largura * dpr;
    canvas.height = altura * dpr;
    ctx2d.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx2d.textBaseline = "middle";

    // fundo branco — a imagem é pra ser compartilhada (WhatsApp/e-mail), não
    // segue o tema do app.
    ctx2d.fillStyle = "#ffffff";
    ctx2d.fillRect(0, 0, largura, altura);

    ctx2d.fillStyle = COR_CABECALHO;
    ctx2d.fillRect(0, 0, largura, alturaCabecalho);
    ctx2d.fillStyle = "#ffffff";
    ctx2d.font = fonteCabecalho;
    let x = 0;
    nomes.forEach((n, i) => {
        ctx2d.fillText(n, x + padX, alturaCabecalho / 2);
        x += larguras[i];
    });

    linhasTexto.forEach((valores, rowIdx) => {
        const ehTotal = rowIdx === indiceLinhaTotal;
        const y = alturaCabecalho + rowIdx * alturaLinha;
        ctx2d.fillStyle = ehTotal ? "#e8f0ed" : rowIdx % 2 === 0 ? "#f4f9f7" : "#ffffff";
        ctx2d.fillRect(0, y, largura, alturaLinha);

        ctx2d.font = ehTotal ? fonteCabecalho : fonteCelula;
        ctx2d.fillStyle = "#12302b";
        let cx = 0;
        valores.forEach((v, i) => {
            ctx2d.fillText(v, cx + padX, y + alturaLinha / 2);
            cx += larguras[i];
        });
    });

    ctx2d.strokeStyle = "#e8f0ed";
    ctx2d.lineWidth = 1;
    let bx = 0;
    for (let i = 0; i < larguras.length; i++) {
        bx += larguras[i];
        ctx2d.beginPath();
        ctx2d.moveTo(bx + 0.5, 0);
        ctx2d.lineTo(bx + 0.5, altura);
        ctx2d.stroke();
    }
    for (let rowIdx = 0; rowIdx <= linhasTexto.length; rowIdx++) {
        const y = alturaCabecalho + rowIdx * alturaLinha;
        ctx2d.beginPath();
        ctx2d.moveTo(0, y + 0.5);
        ctx2d.lineTo(largura, y + 0.5);
        ctx2d.stroke();
    }

    if (indiceLinhaTotal >= 0) {
        const yTotal = alturaCabecalho + indiceLinhaTotal * alturaLinha;
        ctx2d.strokeStyle = COR_CABECALHO;
        ctx2d.lineWidth = 2;
        ctx2d.beginPath();
        ctx2d.moveTo(0, yTotal + 1);
        ctx2d.lineTo(largura, yTotal + 1);
        ctx2d.stroke();
    }

    return canvas;
}

async function copiarImagem() {
    const canvas = desenharCanvasExport();
    canvas.toBlob(async (blob) => {
        try {
            await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
            ctxApp?.setStatus("Imagem copiada para a área de transferência.");
        } catch (e) {
            alert("Não foi possível copiar a imagem.\n\n" + e);
        }
    }, "image/png");
}

async function salvarImagem() {
    const canvas = desenharCanvasExport();
    const dataUrl = canvas.toDataURL("image/png");
    const base64 = dataUrl.split(",")[1];
    try {
        const caminho = await SalvarImagemPNG(base64);
        if (caminho) ctxApp?.setStatus(`Imagem salva: ${caminho}`);
    } catch (e) {
        alert("Não foi possível salvar a imagem.\n\n" + e);
    }
}
