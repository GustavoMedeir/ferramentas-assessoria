// Aba de Configurações que deixa o usuário ajustar a área da página
// recortada pelo botão "Copiar imagem" (aba Rentabilidade): campos
// numéricos + um retângulo arrastável/redimensionável sobre uma prévia real
// da página do relatório — mover o corpo desloca o recorte inteiro, as 4
// alças nos cantos redimensionam.

import { state } from "../state.js";
import { el, clear, btn } from "../ui/components.js";
import { parseNumeroPtBR, formatarPercentual, formatarPercentualEnquantoDigita } from "../util/numeros.js";
import {
    ObterPreviaPaginaGraficoRentabilidade,
    SalvarRecorteGraficoRentabilidade,
    RestaurarRecorteGraficoPadrao,
} from "../../wailsjs/go/main/App.js";

const CAMPOS = [
    { chave: "x0", label: "Esquerda" },
    { chave: "y0", label: "Topo" },
    { chave: "x1", label: "Direita" },
    { chave: "y1", label: "Baixo" },
];

// Tamanho mínimo do recorte (fração da página) — evita uma caixa
// degenerada (largura ou altura ~0) ao arrastar uma alça até o canto oposto.
const TAMANHO_MIN = 0.02;

function clamp(v, lo, hi) {
    return Math.min(hi, Math.max(lo, v));
}

function recorteAtual() {
    const p = state.prefs;
    return p.recortePersonalizado
        ? { x0: p.recorteX0, y0: p.recorteY0, x1: p.recorteX1, y1: p.recorteY1 }
        : { x0: p.recortePadraoX0, y0: p.recortePadraoY0, x1: p.recortePadraoX1, y1: p.recortePadraoY1 };
}

function camposIguais(a, b) {
    return CAMPOS.every((c) => Math.abs(a[c.chave] - b[c.chave]) < 1e-9);
}

// Reforça, só no campo que acabou de ser editado, que o recorte não cruza a
// borda oposta (chamado no blur do campo — durante a digitação o valor não
// é corrigido, pra não brigar com o cursor).
function normalizarRecorte(r, campoEditado) {
    let { x0, y0, x1, y1 } = r;
    if (campoEditado === "x0") x0 = Math.min(x0, x1 - TAMANHO_MIN);
    if (campoEditado === "x1") x1 = Math.max(x1, x0 + TAMANHO_MIN);
    if (campoEditado === "y0") y0 = Math.min(y0, y1 - TAMANHO_MIN);
    if (campoEditado === "y1") y1 = Math.max(y1, y0 + TAMANHO_MIN);
    return { x0: clamp(x0, 0, 1), y0: clamp(y0, 0, 1), x1: clamp(x1, 0, 1), y1: clamp(y1, 0, 1) };
}

function cantoParaPonto(r, canto) {
    return {
        x: canto === "nw" || canto === "sw" ? r.x0 : r.x1,
        y: canto === "nw" || canto === "ne" ? r.y0 : r.y1,
    };
}

// Monta a prévia da página + o retângulo interativo. Devolve um objeto com
// `atualizar(recorte)` pra sincronizar a caixa quando o recorte muda por
// fora (ex.: o usuário editou um campo numérico).
function montarPreview(container, base64, recorteInicial, { onMudar }) {
    const wrap = el("div", { class: "recorte-imgwrap" });
    const img = el("img", { class: "recorte-img" });
    img.src = `data:image/png;base64,${base64}`;
    wrap.appendChild(img);

    const caixa = el("div", { class: "recorte-caixa" });
    const handles = {};
    for (const canto of ["nw", "ne", "se", "sw"]) {
        handles[canto] = el("div", { class: `recorte-handle recorte-handle-${canto}` });
        caixa.appendChild(handles[canto]);
    }
    wrap.appendChild(caixa);
    container.appendChild(wrap);

    let atual = { ...recorteInicial };

    function posicionar() {
        caixa.style.left = `${atual.x0 * 100}%`;
        caixa.style.top = `${atual.y0 * 100}%`;
        caixa.style.width = `${(atual.x1 - atual.x0) * 100}%`;
        caixa.style.height = `${(atual.y1 - atual.y0) * 100}%`;
    }
    posicionar();

    function fracaoDoEvento(e) {
        const rect = wrap.getBoundingClientRect();
        return {
            fx: clamp((e.clientX - rect.left) / rect.width, 0, 1),
            fy: clamp((e.clientY - rect.top) / rect.height, 0, 1),
        };
    }

    // Arrastar o corpo da caixa move o recorte inteiro (mesmo tamanho).
    caixa.addEventListener("mousedown", (e) => {
        if (e.target !== caixa) return; // clique numa alça é resize, não move
        e.preventDefault();
        const inicioMouse = fracaoDoEvento(e);
        const inicioRect = { ...atual };
        const largura = inicioRect.x1 - inicioRect.x0;
        const altura = inicioRect.y1 - inicioRect.y0;

        function mover(ev) {
            const { fx, fy } = fracaoDoEvento(ev);
            const x0 = clamp(inicioRect.x0 + (fx - inicioMouse.fx), 0, 1 - largura);
            const y0 = clamp(inicioRect.y0 + (fy - inicioMouse.fy), 0, 1 - altura);
            atual = { x0, y0, x1: x0 + largura, y1: y0 + altura };
            posicionar();
        }
        function soltar() {
            document.removeEventListener("mousemove", mover);
            document.removeEventListener("mouseup", soltar);
            onMudar(atual);
        }
        document.addEventListener("mousemove", mover);
        document.addEventListener("mouseup", soltar);
    });

    // Arrastar uma alça redimensiona a partir do canto oposto (fixo) — usar
    // min/max entre o canto fixo e o mouse trata sozinho o caso de arrastar
    // "passando" do canto oposto (a caixa vira do outro lado, sem código
    // extra pra detectar a inversão).
    const OPOSTO = { nw: "se", ne: "sw", se: "nw", sw: "ne" };
    for (const canto of ["nw", "ne", "se", "sw"]) {
        handles[canto].addEventListener("mousedown", (e) => {
            e.preventDefault();
            e.stopPropagation();
            const fixo = cantoParaPonto(atual, OPOSTO[canto]);

            function mover(ev) {
                const { fx, fy } = fracaoDoEvento(ev);
                const x0 = Math.min(fixo.x, fx);
                const x1 = Math.max(fixo.x, fx);
                const y0 = Math.min(fixo.y, fy);
                const y1 = Math.max(fixo.y, fy);
                atual = {
                    x0: clamp(Math.min(x0, x1 - TAMANHO_MIN), 0, 1),
                    y0: clamp(Math.min(y0, y1 - TAMANHO_MIN), 0, 1),
                    x1: clamp(Math.max(x1, x0 + TAMANHO_MIN), 0, 1),
                    y1: clamp(Math.max(y1, y0 + TAMANHO_MIN), 0, 1),
                };
                posicionar();
            }
            function soltar() {
                document.removeEventListener("mousemove", mover);
                document.removeEventListener("mouseup", soltar);
                onMudar(atual);
            }
            document.addEventListener("mousemove", mover);
            document.addEventListener("mouseup", soltar);
        });
    }

    return {
        atualizar(novoRecorte) {
            atual = { ...novoRecorte };
            posicionar();
        },
    };
}

export function montarAbaRecorteImagem(ctx) {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "Recorte da imagem" }));
    body.appendChild(
        el("p", { class: "cfg-sub", text: 'Ajuste a área da página copiada pelo botão "Copiar imagem" na aba Rentabilidade.' })
    );

    if (!state.arquivoSelecionado) {
        body.appendChild(
            el("p", {
                class: "cfg-placeholder",
                text: "Selecione um cliente na aba Rentabilidade primeiro, pra carregar uma prévia real do relatório.",
            })
        );
        return body;
    }

    let recorte = recorteAtual();
    let recorteSalvo = { ...recorte };
    let overlay = null;

    const previewWrap = el("div", { class: "recorte-preview-wrap" });
    previewWrap.appendChild(el("p", { class: "cfg-placeholder", text: "Carregando prévia..." }));
    body.appendChild(previewWrap);

    const camposEls = {};
    const grpCampos = el("div", { class: "recorte-campos" });
    for (const campo of CAMPOS) {
        const linhaEl = el("div", { class: "crow" });
        linhaEl.appendChild(el("span", { class: "clbl", text: campo.label }));
        const input = el("input", { class: "cinput", type: "text" });
        input.value = formatarPercentual(recorte[campo.chave] * 100);
        input.addEventListener("input", () => {
            const distanciaDoFim = input.value.length - input.selectionStart;
            input.value = formatarPercentualEnquantoDigita(input.value);
            const pos = Math.max(0, input.value.length - distanciaDoFim);
            input.setSelectionRange(pos, pos);

            const valor = parseNumeroPtBR(input.value);
            if (valor !== null) {
                recorte = { ...recorte, [campo.chave]: clamp(valor / 100, 0, 1) };
                overlay?.atualizar(recorte);
                atualizarBotoes();
            }
        });
        input.addEventListener("blur", () => {
            recorte = normalizarRecorte(recorte, campo.chave);
            atualizarCampos();
            overlay?.atualizar(recorte);
            atualizarBotoes();
        });
        camposEls[campo.chave] = input;
        linhaEl.appendChild(el("div", { class: "cwrap" }, [input]));
        grpCampos.appendChild(linhaEl);
    }
    body.appendChild(grpCampos);

    const botaoConfirmar = btn("Confirmar", { classe: "pri" });
    const botaoRestaurar = btn("Restaurar padrão", { classe: "ghost" });

    function atualizarCampos() {
        for (const campo of CAMPOS) camposEls[campo.chave].value = formatarPercentual(recorte[campo.chave] * 100);
    }

    function atualizarBotoes() {
        botaoConfirmar.disabled = camposIguais(recorte, recorteSalvo);
    }
    atualizarBotoes();

    botaoConfirmar.addEventListener("click", async () => {
        try {
            await SalvarRecorteGraficoRentabilidade(recorte.x0, recorte.y0, recorte.x1, recorte.y1);
            state.prefs.recortePersonalizado = true;
            state.prefs.recorteX0 = recorte.x0;
            state.prefs.recorteY0 = recorte.y0;
            state.prefs.recorteX1 = recorte.x1;
            state.prefs.recorteY1 = recorte.y1;
            recorteSalvo = { ...recorte };
            atualizarBotoes();
            ctx?.setStatus("Recorte da imagem salvo.");
        } catch (e) {
            ctx?.setStatus("Erro ao salvar recorte: " + e);
        }
    });

    botaoRestaurar.addEventListener("click", async () => {
        try {
            await RestaurarRecorteGraficoPadrao();
            state.prefs.recortePersonalizado = false;
            recorte = {
                x0: state.prefs.recortePadraoX0,
                y0: state.prefs.recortePadraoY0,
                x1: state.prefs.recortePadraoX1,
                y1: state.prefs.recortePadraoY1,
            };
            recorteSalvo = { ...recorte };
            atualizarCampos();
            overlay?.atualizar(recorte);
            atualizarBotoes();
            ctx?.setStatus("Recorte restaurado ao padrão.");
        } catch (e) {
            ctx?.setStatus("Erro ao restaurar recorte: " + e);
        }
    });

    body.appendChild(el("div", { class: "cfg-confirm-row" }, [botaoRestaurar, botaoConfirmar]));

    ObterPreviaPaginaGraficoRentabilidade(state.arquivoSelecionado)
        .then((base64) => {
            clear(previewWrap);
            overlay = montarPreview(previewWrap, base64, recorte, {
                onMudar(novoRecorte) {
                    recorte = novoRecorte;
                    atualizarCampos();
                    atualizarBotoes();
                },
            });
        })
        .catch((e) => {
            clear(previewWrap);
            previewWrap.appendChild(el("p", { class: "cfg-placeholder", text: "Não foi possível carregar a prévia: " + e }));
        });

    return body;
}
