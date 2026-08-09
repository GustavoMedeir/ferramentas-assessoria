import { state, novoBlocoId } from "../state.js";
import { el, clear, btn } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import { parseNumeroPtBR, formatarReais, formatarPercentual } from "../util/numeros.js";
import { SalvarImagemPNG } from "../../wailsjs/go/main/App.js";

let refs = {};
let ctxApp = null;
let criterioOrdem = "desagio"; // "nome" | "valorAtual" | "desagio"
let ordemAsc = false; // true: crescente (A→Z / menor→maior); false: decrescente
let mostrarSoma = false;

const COLUNAS = ["Título", "Valor Atual", "Valor de Saída", "Deságio (R$)", "Deságio (%)"];
const COR_CABECALHO = "#047857"; // --primary-600 do acento Esmeralda (canvas não lê variável CSS)

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

    // ---- painel esquerdo: linhas ----
    const esquerdo = el("div", { class: "des-l" });
    esquerdo.appendChild(el("div", { class: "ops-h" }, [el("span", { text: "Títulos" })]));
    refs.listaLinhas = el("div", { class: "ops" });
    esquerdo.appendChild(refs.listaLinhas);
    esquerdo.appendChild(
        el("div", { class: "mail-foot" }, [btn("Adicionar título", { icon: icons.iconMais, onClick: adicionarLinha })])
    );

    // ---- painel direito: prévia da imagem ----
    const direito = el("div", { class: "des-r" });

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
        desenharTabela();
    });
    refs.btnOrdem = btn("Ágio → Deságio", { icon: icons.iconOrdenar, onClick: alternarOrdem });
    refs.btnSoma = btn("Mostrar soma", { onClick: alternarSoma });
    const controles = el("div", { class: "des-controles" }, [refs.selectOrdem, refs.btnOrdem, refs.btnSoma]);

    direito.appendChild(el("div", { class: "mail-r-h" }, [el("h3", { text: "Tabela de deságio" }), controles]));
    refs.canvasWrap = el("div", { class: "table-card" });
    refs.canvas = el("canvas");
    refs.canvasWrap.appendChild(refs.canvas);
    direito.appendChild(refs.canvasWrap);
    direito.appendChild(
        el("div", { class: "des-foot" }, [
            btn("Copiar imagem", { classe: "pri", icon: icons.iconImagem, onClick: copiarImagem }),
            btn("Salvar imagem", { icon: icons.iconSalvar, onClick: salvarImagem }),
        ])
    );

    container.appendChild(esquerdo);
    container.appendChild(direito);

    if (state.blocosDesagio.length === 0) {
        state.blocosDesagio.push(novaLinha());
    }
    renderLinhas();
    desenharTabela();
}

function novaLinha() {
    return { id: novoBlocoId(), titulo: "", valorAtual: "", valorSaida: "" };
}

function adicionarLinha() {
    state.blocosDesagio.push(novaLinha());
    renderLinhas();
    desenharTabela();
}

function atualizarLabelOrdem() {
    if (!refs.btnOrdem) return;
    const rotulos = ROTULOS_ORDEM[criterioOrdem];
    refs.btnOrdem.lastChild.textContent = ordemAsc ? rotulos.asc : rotulos.desc;
}

function alternarOrdem() {
    ordemAsc = !ordemAsc;
    atualizarLabelOrdem();
    desenharTabela();
}

function atualizarLabelSoma() {
    if (!refs.btnSoma) return;
    refs.btnSoma.lastChild.textContent = mostrarSoma ? "Ocultar soma" : "Mostrar soma";
}

function alternarSoma() {
    mostrarSoma = !mostrarSoma;
    atualizarLabelSoma();
    desenharTabela();
}

function renderLinhas() {
    if (!refs.listaLinhas) return;
    clear(refs.listaLinhas);
    state.blocosDesagio.forEach((linha, idx) => {
        refs.listaLinhas.appendChild(renderLinha(linha, idx));
    });
}

function campoComLabel(rotulo, valorInicial, placeholder, aoDigitar) {
    const grupo = el("div", {});
    grupo.appendChild(el("div", { class: "field-lbl", text: rotulo }));
    const input = el("input", { class: "input", type: "text", placeholder });
    input.value = valorInicial;
    input.addEventListener("input", () => aoDigitar(input.value));
    grupo.appendChild(input);
    return grupo;
}

function renderLinha(linha, idx) {
    const wrapper = el("div", { class: "op" });
    wrapper.appendChild(el("div", { class: "op-h" }, [el("span", { class: "op-num", text: String(idx + 1) }), "Título"]));

    wrapper.appendChild(
        campoComLabel("Nome do título", linha.titulo, "Nome do título", (v) => {
            linha.titulo = v;
            desenharTabela();
        })
    );
    wrapper.appendChild(
        campoComLabel("Valor atual", linha.valorAtual, "Ex: 1.234,56", (v) => {
            linha.valorAtual = v;
            desenharTabela();
        })
    );
    wrapper.appendChild(
        campoComLabel("Valor de saída", linha.valorSaida, "Ex: 1.000,00", (v) => {
            linha.valorSaida = v;
            desenharTabela();
        })
    );

    const btnRemover = btn("Remover", {
        classe: "danger",
        onClick: () => {
            state.blocosDesagio = state.blocosDesagio.filter((b) => b.id !== linha.id);
            renderLinhas();
            desenharTabela();
        },
    });
    wrapper.appendChild(el("div", { class: "op-foot" }, [btnRemover]));

    return wrapper;
}

// Deságio = diferença entre o valor de saída e o valor atual, em R$ e em %
// relativo ao valor atual — negativo quando o valor de saída é menor que o
// atual (deságio de verdade), positivo quando é maior (ágio).
//
// A tabela (canvas) sai ordenada pelo critério/direção escolhidos
// (criterioOrdem/ordemAsc, ver compararLinhas) — só a saída; a lista
// editável à esquerda (renderLinhas, que itera state.blocosDesagio direto)
// mantém a ordem de cadastro, senão a linha "pularia" de lugar enquanto o
// usuário digita. Linhas sem valor numérico no critério atual ficam por
// último (nome sempre tem valor, então só se aplica a valorAtual/desagio).
function compararLinhas(a, b) {
    let cmp;
    if (criterioOrdem === "nome") {
        cmp = a.titulo.localeCompare(b.titulo, "pt-BR", { sensitivity: "base" });
    } else if (criterioOrdem === "valorAtual") {
        if (a.valorAtual === null && b.valorAtual === null) return 0;
        if (a.valorAtual === null) return 1;
        if (b.valorAtual === null) return -1;
        cmp = a.valorAtual - b.valorAtual;
    } else {
        if (a.desagioPct === null && b.desagioPct === null) return 0;
        if (a.desagioPct === null) return 1;
        if (b.desagioPct === null) return -1;
        cmp = a.desagioPct - b.desagioPct;
    }
    return ordemAsc ? cmp : -cmp;
}

function calcularLinhas() {
    const linhas = state.blocosDesagio.map((linha) => {
        const valorAtual = parseNumeroPtBR(linha.valorAtual);
        const valorSaida = parseNumeroPtBR(linha.valorSaida);
        let desagioReais = null;
        let desagioPct = null;
        if (valorAtual !== null && valorSaida !== null) {
            desagioReais = valorSaida - valorAtual;
            desagioPct = valorAtual !== 0 ? (desagioReais / valorAtual) * 100 : null;
        }
        return { titulo: linha.titulo?.trim() || "—", valorAtual, valorSaida, desagioReais, desagioPct };
    });

    return linhas.sort(compararLinhas);
}

function desenharTabela() {
    if (!refs.canvas) return;
    const linhas = calcularLinhas();
    const ctx2d = refs.canvas.getContext("2d");

    const fonteCabecalho = "bold 14px 'Plus Jakarta Sans', sans-serif";
    const fonteCelula = "13px 'Plus Jakarta Sans', sans-serif";
    const padX = 16;
    const alturaLinha = 38;
    const alturaCabecalho = 44;

    const linhasTexto = linhas.map((l) => [
        l.titulo,
        l.valorAtual !== null ? formatarReais(l.valorAtual) : "—",
        l.valorSaida !== null ? formatarReais(l.valorSaida) : "—",
        l.desagioReais !== null ? formatarReais(l.desagioReais) : "—",
        l.desagioPct !== null ? formatarPercentual(l.desagioPct) : "—",
    ]);

    // Linha de total opcional (mostrarSoma) — soma as 3 colunas em R$; o %
    // do total é ponderado (soma dos deságios / soma dos valores atuais),
    // não a média simples das % de cada linha.
    let indiceLinhaTotal = -1;
    if (mostrarSoma && linhas.length > 0) {
        const totalValorAtual = linhas.reduce((s, l) => s + (l.valorAtual ?? 0), 0);
        const totalValorSaida = linhas.reduce((s, l) => s + (l.valorSaida ?? 0), 0);
        const totalDesagioReais = linhas.reduce((s, l) => s + (l.desagioReais ?? 0), 0);
        const totalDesagioPct = totalValorAtual !== 0 ? (totalDesagioReais / totalValorAtual) * 100 : null;
        linhasTexto.push([
            "Total",
            formatarReais(totalValorAtual),
            formatarReais(totalValorSaida),
            formatarReais(totalDesagioReais),
            totalDesagioPct !== null ? formatarPercentual(totalDesagioPct) : "—",
        ]);
        indiceLinhaTotal = linhasTexto.length - 1;
    }

    // medir largura de cada coluna (cabeçalho + maior célula)
    ctx2d.font = fonteCabecalho;
    const larguras = COLUNAS.map((c) => ctx2d.measureText(c).width + padX * 2);
    ctx2d.font = fonteCelula;
    for (const valores of linhasTexto) {
        valores.forEach((v, i) => {
            larguras[i] = Math.max(larguras[i], ctx2d.measureText(v).width + padX * 2);
        });
    }

    const largura = larguras.reduce((a, b) => a + b, 0);
    const altura = alturaCabecalho + Math.max(linhasTexto.length, 1) * alturaLinha;

    const dpr = window.devicePixelRatio || 1;
    refs.canvas.width = largura * dpr;
    refs.canvas.height = altura * dpr;
    refs.canvas.style.width = largura + "px";
    refs.canvas.style.height = altura + "px";
    ctx2d.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx2d.textBaseline = "middle";

    // fundo branco — a imagem é pra ser compartilhada (WhatsApp/e-mail), não
    // segue o tema do app.
    ctx2d.fillStyle = "#ffffff";
    ctx2d.fillRect(0, 0, largura, altura);

    // cabeçalho
    ctx2d.fillStyle = COR_CABECALHO;
    ctx2d.fillRect(0, 0, largura, alturaCabecalho);
    ctx2d.fillStyle = "#ffffff";
    ctx2d.font = fonteCabecalho;
    let x = 0;
    COLUNAS.forEach((c, i) => {
        ctx2d.fillText(c, x + padX, alturaCabecalho / 2);
        x += larguras[i];
    });

    // linhas de dados (a linha de total, se houver, sai em negrito e com
    // fundo destacado, igual ao cabeçalho da tabela)
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

    // grade
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

    // traço mais grosso separando a linha de total das linhas normais
    if (indiceLinhaTotal >= 0) {
        const yTotal = alturaCabecalho + indiceLinhaTotal * alturaLinha;
        ctx2d.strokeStyle = COR_CABECALHO;
        ctx2d.lineWidth = 2;
        ctx2d.beginPath();
        ctx2d.moveTo(0, yTotal + 1);
        ctx2d.lineTo(largura, yTotal + 1);
        ctx2d.stroke();
    }
}

async function copiarImagem() {
    if (!refs.canvas) return;
    refs.canvas.toBlob(async (blob) => {
        try {
            await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
            ctxApp?.setStatus("Imagem copiada para a área de transferência.");
        } catch (e) {
            alert("Não foi possível copiar a imagem.\n\n" + e);
        }
    }, "image/png");
}

async function salvarImagem() {
    if (!refs.canvas) return;
    const dataUrl = refs.canvas.toDataURL("image/png");
    const base64 = dataUrl.split(",")[1];
    try {
        const caminho = await SalvarImagemPNG(base64);
        if (caminho) ctxApp?.setStatus(`Imagem salva: ${caminho}`);
    } catch (e) {
        alert("Não foi possível salvar a imagem.\n\n" + e);
    }
}
