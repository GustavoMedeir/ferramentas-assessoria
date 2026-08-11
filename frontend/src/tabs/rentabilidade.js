import { state } from "../state.js";
import { el, clear, btn, phButton, renderMensagemComPills, renderMensagemFestasComPills } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import { base64ParaBlob } from "../util/imagem.js";
import {
    CopiarMensagem,
    CopiarMensagemFestas,
    EnviarWhatsApp,
    EnviarWhatsAppFestas,
    ObterImagemGrafico,
    SalvarModelo,
    SalvarModeloFestas,
} from "../../wailsjs/go/main/App.js";

const LEGENDA = [
    ["Ganho em R$", "_Rent", "_RentA", "_Rent12M"],
    ["Rentabilidade %", "_Perc", "_PercA", "_Perc12M"],
    ["% do CDI", "_CDI", "_CDIA", "_CDI12M"],
];

let refs = {};
let ctxApp = null;
let rentView = "previa"; // "previa" | "modelo" — persiste entre remontagens da seção

export function mount(container, ctx) {
    // clear() esvazia o container e reseta o scroll pro topo (mesmo problema
    // documentado em tabs/emails.js:renderBlocos) — preserva a posição da
    // lista de clientes e da prévia entre remontagens (ex.: toggle do Modo
    // Festas em Configurações), já que o container em si (.rent) não rola.
    const scrollListaAnterior = container.querySelector(".clients")?.scrollTop ?? 0;
    const scrollPreviewAnterior = container.querySelector(".preview")?.scrollTop ?? 0;
    clear(container);
    refs = {};
    ctxApp = ctx;

    if (!state.pasta) {
        container.appendChild(renderEstadoVazio(ctx));
        return;
    }

    // ---- painel esquerdo: lista de clientes ----
    const painelLista = el("div", { class: "rlist" });
    refs.contagem = el("span", { class: "count" });
    painelLista.appendChild(el("div", { class: "rlist-h" }, [el("h3", { text: "Clientes" }), refs.contagem]));

    refs.busca = el("input", { type: "text", placeholder: "Buscar nome ou número..." });
    refs.busca.value = state.filtro;
    refs.busca.addEventListener("input", () => {
        state.filtro = refs.busca.value;
        renderLista();
    });
    const caixaBusca = el("div", { class: "search" });
    caixaBusca.insertAdjacentHTML("beforeend", icons.iconBusca);
    caixaBusca.appendChild(refs.busca);
    painelLista.appendChild(caixaBusca);

    refs.lista = el("div", { class: "clients" });
    painelLista.appendChild(refs.lista);

    // ---- painel direito: prévia / modelo ----
    const painelDireito = el("div", { class: "rpanel" });

    const segPrevia = el("button", { class: "seg-b", text: "Prévia" });
    const segModelo = el("button", { class: "seg-b", text: "Modelo" });
    segPrevia.addEventListener("click", () => {
        rentView = "previa";
        renderCorpo();
    });
    segModelo.addEventListener("click", () => {
        rentView = "modelo";
        renderCorpo();
    });
    refs.segPrevia = segPrevia;
    refs.segModelo = segModelo;

    refs.who = el("span", { class: "who" });
    refs.btnWhatsApp = btn("Enviar WhatsApp", { classe: "ghost", icon: icons.iconWhatsApp, onClick: enviarWhatsApp });
    const quem = el("div", { class: "who-wrap" }, [refs.who, refs.btnWhatsApp]);
    painelDireito.appendChild(el("div", { class: "rpanel-h" }, [el("div", { class: "seg" }, [segPrevia, segModelo]), quem]));

    refs.corpo = el("div", { class: "rpanel-body" });
    painelDireito.appendChild(refs.corpo);

    container.appendChild(painelLista);
    container.appendChild(painelDireito);

    renderLista();
    renderCorpo();
    if (refs.lista) refs.lista.scrollTop = scrollListaAnterior;
    if (refs.preview) refs.preview.scrollTop = scrollPreviewAnterior;
}

function renderEstadoVazio(ctx) {
    return el("div", { class: "estado-vazio" }, [
        el("div", { text: "Nenhuma pasta selecionada." }),
        btn("Escolher pasta com os PDFs", { classe: "pri", icon: icons.iconPasta, onClick: ctx.escolherPasta }),
    ]);
}

function clientesFiltrados() {
    const filtro = state.filtro.trim().toLowerCase();
    if (!filtro) return state.clientes;
    return state.clientes.filter((c) => {
        const codigo = (c.Codigo || "").toLowerCase();
        const nome = (c.Nome || "").toLowerCase();
        return codigo.includes(filtro) || nome.includes(filtro);
    });
}

function selecionarCliente(c) {
    state.clienteSelecionado = c ? c.Codigo : null;
    state.arquivoSelecionado = c && c.Registro ? c.Registro.Arquivo : null;
}

export function renderLista() {
    if (!refs.lista) return;
    // clear() esvazia o container e reseta o scroll pro topo (mesmo problema
    // documentado em mount() acima e em tabs/emails.js:renderBlocos) —
    // preserva a posição da lista entre re-renders (busca, seleção, ações).
    const scrollAnterior = refs.lista.scrollTop;
    clear(refs.lista);

    const visiveis = clientesFiltrados();
    if (state.clienteSelecionado && !visiveis.some((c) => c.Codigo === state.clienteSelecionado)) {
        selecionarCliente(visiveis.length ? visiveis[0] : null);
    } else if (!state.clienteSelecionado && visiveis.length) {
        selecionarCliente(visiveis[0]);
    }

    if (!visiveis.length) {
        refs.lista.appendChild(el("div", { class: "lista-vazia", text: "Nenhum cliente encontrado." }));
    }

    for (const c of visiveis) {
        const selecionado = c.Codigo === state.clienteSelecionado;
        const linha = el("button", {
            class: `client${selecionado ? " sel" : ""}`,
            onClick: () => {
                selecionarCliente(c);
                renderLista();
                renderPreview();
            },
        });

        const ident = el("span", { class: "cid" });
        if (c.Nome) {
            ident.appendChild(el("span", { class: "cname", text: c.Nome }));
            ident.appendChild(el("span", { class: "ccode mono", text: c.Codigo }));
        } else {
            ident.appendChild(el("span", { class: "ccode-solo mono", text: c.Codigo }));
        }

        const direita = el("span", { class: "cright" });
        if (state.prefs.modoFestas) {
            if (c.FestasEnviado) {
                direita.appendChild(el("span", { class: "badge copied", text: "ENVIADO" }));
            } else {
                direita.appendChild(el("span", { class: "cval cval-vazio", text: "—" }));
            }
        } else if (c.Registro) {
            const badge = el("span", {
                class: `badge${c.Registro.Copiado ? " copied" : ""}`,
                text: c.Registro.Copiado ? "COPIADO" : "GERADO",
            });
            const valor = el("span", { class: `cval${c.Registro.GanhoMesReais < 0 ? " neg" : ""}`, text: c.Registro.RentFmt });
            direita.appendChild(badge);
            direita.appendChild(valor);
        } else {
            direita.appendChild(el("span", { class: "cval cval-vazio", text: "—" }));
        }
        linha.appendChild(ident);
        linha.appendChild(direita);
        refs.lista.appendChild(linha);
    }

    if (refs.contagem) refs.contagem.textContent = `${visiveis.length} de ${state.clientes.length}`;
    refs.lista.scrollTop = scrollAnterior;
}

function renderCorpo() {
    if (!refs.corpo) return;
    refs.segPrevia.classList.toggle("on", rentView === "previa");
    refs.segModelo.classList.toggle("on", rentView === "modelo");
    clear(refs.corpo);
    delete refs.preview;
    delete refs.textarea;

    if (rentView === "previa") {
        montarCorpoPrevia(refs.corpo);
    } else {
        montarCorpoModelo(refs.corpo);
    }
    renderPreview();
}

function montarCorpoPrevia(corpo) {
    refs.preview = el("div", { class: "card preview" });
    corpo.appendChild(refs.preview);

    const botoes = [];
    // "Copiar imagem" recorta o gráfico do PDF de rentabilidade — não existe
    // no Modo Festas, que também vale pra clientes sem relatório processado.
    if (!state.prefs.modoFestas) {
        refs.btnCopiarImagem = btn("Copiar imagem", { icon: icons.iconImagem, onClick: copiarImagemGrafico });
        botoes.push(refs.btnCopiarImagem);
    }
    refs.btnCopiar = btn("Copiar mensagem", { classe: "pri", icon: icons.iconCopiar, onClick: copiarMensagem });
    botoes.push(refs.btnCopiar);
    corpo.appendChild(el("div", { class: "rfoot" }, botoes));
}

function montarCorpoModelo(corpo) {
    if (state.prefs.modoFestas) {
        montarCorpoModeloFestas(corpo);
        return;
    }

    const legenda = el("div", { class: "legend" });
    legenda.appendChild(el("span", {}));
    legenda.appendChild(el("span", { class: "col-h", text: "Mês" }));
    legenda.appendChild(el("span", { class: "col-h", text: "Ano" }));
    legenda.appendChild(el("span", { class: "col-h", text: "12M" }));
    for (const [rotulo, mes, ano, doze] of LEGENDA) {
        legenda.appendChild(el("span", { class: "rlbl", text: rotulo }));
        legenda.appendChild(phButton(mes, () => inserirPlaceholder(mes)));
        legenda.appendChild(phButton(ano, () => inserirPlaceholder(ano)));
        legenda.appendChild(phButton(doze, () => inserirPlaceholder(doze)));
    }
    legenda.appendChild(el("span", { class: "rlbl", text: "Nome do cliente" }));
    legenda.appendChild(phButton("_Nome", () => inserirPlaceholder("_Nome")));
    legenda.appendChild(phButton("_NomeM", () => inserirPlaceholder("_NomeM")));
    legenda.appendChild(el("span", {}));
    corpo.appendChild(legenda);

    refs.textarea = el("textarea", { class: "editor" });
    refs.textarea.value = state.modelo;
    refs.textarea.addEventListener("input", () => {
        state.modelo = refs.textarea.value;
    });
    corpo.appendChild(refs.textarea);

    const btnSalvar = btn("Salvar modelo", {
        classe: "pri",
        icon: icons.iconSalvar,
        onClick: async () => {
            try {
                await SalvarModelo(state.modelo);
                ctxApp?.setStatus("Modelo salvo.");
            } catch (e) {
                ctxApp?.setStatus("Erro ao salvar modelo: " + e);
            }
        },
    });
    corpo.appendChild(el("div", { class: "rfoot" }, [btnSalvar]));
}

// Modelo do Modo Festas: só existem os placeholders _Nome/_NomeM (sem dados
// financeiros, já que também vale pra clientes sem relatório processado).
function montarCorpoModeloFestas(corpo) {
    const legenda = el("div", { class: "legend" });
    legenda.appendChild(el("span", { class: "rlbl", text: "Nome do cliente" }));
    legenda.appendChild(phButton("_Nome", () => inserirPlaceholder("_Nome")));
    legenda.appendChild(phButton("_NomeM", () => inserirPlaceholder("_NomeM")));
    legenda.appendChild(el("span", {}));
    corpo.appendChild(legenda);

    refs.textarea = el("textarea", { class: "editor" });
    refs.textarea.value = state.modeloFestas;
    refs.textarea.addEventListener("input", () => {
        state.modeloFestas = refs.textarea.value;
    });
    corpo.appendChild(refs.textarea);

    const btnSalvar = btn("Salvar modelo", {
        classe: "pri",
        icon: icons.iconSalvar,
        onClick: async () => {
            try {
                await SalvarModeloFestas(state.modeloFestas);
                ctxApp?.setStatus("Modelo de festas salvo.");
            } catch (e) {
                ctxApp?.setStatus("Erro ao salvar modelo de festas: " + e);
            }
        },
    });
    corpo.appendChild(el("div", { class: "rfoot" }, [btnSalvar]));
}

// Chamado de fora (main.js) sempre que a lista de registros muda (processar
// pasta, copiar mensagem, limpar tudo). Atualiza o rótulo do cliente
// selecionado (visível nas duas views) e o conteúdo da prévia, quando ela
// está montada.
export function renderPreview() {
    const cliente = state.clientes.find((c) => c.Codigo === state.clienteSelecionado) || null;
    const registro = cliente?.Registro || null;

    if (refs.who) {
        refs.who.textContent = cliente ? (cliente.Nome ? `${cliente.Nome} (${cliente.Codigo})` : `cliente ${cliente.Codigo}`) : "";
    }

    // No Modo Festas a ação vale pra todo cliente da base, com ou sem
    // relatório — a mensagem de festas não depende de dados financeiros.
    const modoFestas = state.prefs.modoFestas;
    const temAcao = modoFestas ? !!cliente : !!registro;
    if (refs.btnWhatsApp) refs.btnWhatsApp.disabled = !temAcao;
    if (refs.btnCopiarImagem) refs.btnCopiarImagem.disabled = !temAcao;
    if (refs.btnCopiar) refs.btnCopiar.disabled = !temAcao;

    if (!refs.preview) return; // só existe quando rentView === "previa"
    if (!cliente) {
        clear(refs.preview);
        refs.preview.textContent = "Selecione um cliente na lista.";
        return;
    }
    if (modoFestas) {
        renderMensagemFestasComPills(refs.preview, state.modeloFestas, cliente.Nome);
        return;
    }
    if (!registro) {
        clear(refs.preview);
        refs.preview.textContent = "Nenhum relatório encontrado para este cliente.";
        return;
    }
    renderMensagemComPills(refs.preview, state.modelo, registro, cliente.Nome);
}

function inserirPlaceholder(placeholder) {
    const textarea = refs.textarea;
    if (!textarea) return;

    const inicio = textarea.selectionStart ?? textarea.value.length;
    const fim = textarea.selectionEnd ?? textarea.value.length;
    const novoValor = textarea.value.slice(0, inicio) + placeholder + textarea.value.slice(fim);

    textarea.value = novoValor;
    if (state.prefs.modoFestas) {
        state.modeloFestas = novoValor;
    } else {
        state.modelo = novoValor;
    }

    const novaPosicao = inicio + placeholder.length;
    textarea.focus();
    textarea.setSelectionRange(novaPosicao, novaPosicao);
}

async function copiarMensagem() {
    if (state.prefs.modoFestas) {
        if (!state.clienteSelecionado) return;
        try {
            const resultado = await CopiarMensagemFestas(state.clienteSelecionado, state.modeloFestas);
            state.clientes = resultado.Clientes;
            renderLista();
        } catch (e) {
            alert("Não foi possível copiar a mensagem.\n\n" + e);
        }
        return;
    }
    if (!state.arquivoSelecionado) return;
    try {
        const resultado = await CopiarMensagem(state.arquivoSelecionado, state.modelo);
        state.clientes = resultado.Clientes;
        renderLista();
    } catch (e) {
        alert("Não foi possível copiar a mensagem.\n\n" + e);
    }
}

async function enviarWhatsApp() {
    if (state.prefs.modoFestas) {
        if (!state.clienteSelecionado) return;
        try {
            const resultado = await EnviarWhatsAppFestas(state.clienteSelecionado, state.modeloFestas);
            state.clientes = resultado.Clientes;
            renderLista();
        } catch (e) {
            alert("Não foi possível abrir o WhatsApp.\n\n" + e);
        }
        return;
    }
    if (!state.arquivoSelecionado) return;
    try {
        const resultado = await EnviarWhatsApp(state.arquivoSelecionado, state.modelo);
        state.clientes = resultado.Clientes;
        renderLista();
    } catch (e) {
        alert("Não foi possível abrir o WhatsApp.\n\n" + e);
    }
}

async function copiarImagemGrafico() {
    if (!state.arquivoSelecionado) return;
    try {
        const base64 = await ObterImagemGrafico(state.arquivoSelecionado);
        const blob = base64ParaBlob(base64, "image/png");
        await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
        ctxApp?.setStatus("Imagem do gráfico copiada para a área de transferência.");
    } catch (e) {
        alert("Não foi possível copiar a imagem do gráfico.\n\n" + e);
    }
}
