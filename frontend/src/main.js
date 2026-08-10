import {
    EstadoInicial,
    EscolherPasta,
    ProcessarPastaAtual,
    LimparTudo,
    LimparFestasEnviados,
    ExportarCSV,
    CarregarBaseClientes,
    CategoriasEmail,
    SalvarPreferencias,
    SalvarEmailRemetente,
    SalvarDadosAssessor,
    ListarAssinaturas,
    AdicionarAssinatura,
    SelecionarAssinaturaAtiva,
    RemoverAssinatura,
    Plataforma,
    VersaoAtual,
    VerificarAtualizacao,
    AplicarAtualizacao,
} from "../wailsjs/go/main/App.js";
import { EventsOn } from "../wailsjs/runtime/runtime.js";

import { state } from "./state.js";
import { el, clear, btn } from "./ui/components.js";
import * as icons from "./ui/icons.js";
import brandIconUrl from "./assets/brand-icon.png";
import * as rentabilidadeTab from "./tabs/rentabilidade.js";
import * as emailsTab from "./tabs/emails.js";
import * as desagioTab from "./tabs/desagio.js";
import * as calculadoraTab from "./tabs/calculadora.js";
import * as previdenciariaTab from "./tabs/previdenciaria.js";
import * as compromissadaTab from "./tabs/compromissada.js";
import * as comparadoraTab from "./tabs/comparadora.js";
import * as aposentadoriaTab from "./tabs/aposentadoria.js";
import * as editorPdfTab from "./tabs/editorPdf.js";
import * as imagensPdfTab from "./tabs/imagensPdf.js";
import * as apresentacaoTab from "./tabs/apresentacao.js";
import * as typeformTab from "./tabs/typeform.js";
import * as validacaoAssinaturaTab from "./tabs/validacaoAssinatura.js";
import { montarAbaRecorteImagem } from "./tabs/recorteImagem.js";

const appRoot = document.getElementById("app");
const dom = {};
let secaoAtual = "rent";

const TITULOS = {
    rent: ["Rentabilidade", "Selecione um cliente para revisar e copiar a mensagem"],
    mail: ["E-mails de Ordem", "Monte as operações e gere o e-mail do cliente"],
    des: ["Tabela de Deságio", "Compare valor atual e valor de saída dos títulos"],
    calc: ["Calculadora Financeira", "Projeções de valores e conversão de taxas"],
    cmp: ["Comparadora de Renda Fixa", "Compare dois ativos lado a lado e veja o melhor líquido"],
    prev: ["Calculadora Previdenciária", "Compare declaração Simplificada, Completa e Completa com 12% em PGBL"],
    comp: ["Compromissada", "Compare Conta Remunerada, CDB e Compromissada dia a dia"],
    apos: ["Planejamento de Aposentadoria", "Calcule quanto contribuir por mês para se aposentar com a renda desejada"],
    pdfedit: ["Editor de PDF", "Adicione texto, desenhos, tarjas e imagens direto no PDF"],
    imgpdf: ["Imagens em PDF", "Escolha e ordene imagens para juntar num único PDF"],
    apres: ["Apresentação", ""],
    typ: ["Typeform", "Preencha o formulário do cliente durante a reunião"],
    icp: ["Validação de Assinatura", "Confira a autenticidade de uma assinatura digital ICP-Brasil — verificação local, offline"],
};

function setStatus(texto) {
    dom.statusTexto.textContent = texto;
}

function mostrarProgresso(visivel) {
    dom.progresso.hidden = !visivel;
    if (visivel) {
        dom.progressoLabel.textContent = "Processando arquivos...";
        dom.progressoBar.value = 0;
        dom.progressoBar.max = 1;
    }
}

function atualizarAcoesRent() {
    dom.chipPasta.textContent = state.pasta || "Nenhuma pasta selecionada";
    const temPasta = !!state.pasta;
    dom.btnAtualizar.disabled = !temPasta;
    dom.btnExportar.disabled = !temPasta;
    dom.btnLimpar.disabled = !temPasta;
    dom.btnLimparFestas.disabled = !temPasta;
    dom.btnLimparFestas.style.display = state.prefs.modoFestas ? "" : "none";
}

async function processarPastaAtual() {
    mostrarProgresso(true);
    try {
        const resultado = await ProcessarPastaAtual();
        state.clientes = resultado.Clientes;
        state.falhasProcessamento = resultado.Falhas || [];
        if (resultado.Falhas && resultado.Falhas.length) {
            setStatus(`${resultado.Falhas.length} arquivo(s) com falha ao processar (veja em Configurações).`);
            console.warn("Falhas ao processar PDFs:", resultado.Falhas);
        } else {
            setStatus(`Pronto · ${resultado.Clientes.length} cliente(s) carregado(s).`);
        }
    } catch (e) {
        setStatus("Erro ao processar pasta: " + e);
    } finally {
        mostrarProgresso(false);
    }
    rentabilidadeTab.renderLista();
    rentabilidadeTab.renderPreview();
}

async function escolherPasta() {
    try {
        const resultado = await EscolherPasta();
        if (!resultado || !resultado.Pasta) return; // usuário cancelou
        state.pasta = resultado.Pasta;
        state.modelo = resultado.Modelo;
        state.modeloFestas = resultado.ModeloFestas;
        state.arquivoSelecionado = null;
        state.clienteSelecionado = null;
        atualizarAcoesRent();
        rentabilidadeTab.mount(dom.secaoRent, ctx);
        await processarPastaAtual();
    } catch (e) {
        setStatus("Erro ao escolher pasta: " + e);
    }
}

async function limparTudo() {
    if (!state.pasta) return;
    const aviso =
        "Isso apaga todos os códigos e status (GERADO/COPIADO) salvos dessa pasta. Os PDFs continuam no " +
        "disco e são reprocessados do zero na próxima leitura. Continuar?";
    if (!confirm(aviso)) return;
    try {
        const resultado = await LimparTudo();
        state.clientes = resultado.Clientes;
        setStatus("Lista limpa.");
    } catch (e) {
        setStatus("Erro ao limpar: " + e);
    }
    rentabilidadeTab.renderLista();
    rentabilidadeTab.renderPreview();
}

async function limparFestasEnviados() {
    if (!state.pasta) return;
    if (!confirm("Isso apaga o selo ENVIADO de todo mundo (Modo Festas). Use antes de começar uma nova leva. Continuar?")) return;
    try {
        const resultado = await LimparFestasEnviados();
        state.clientes = resultado.Clientes;
        setStatus("Envios de Festas zerados.");
    } catch (e) {
        setStatus("Erro ao limpar envios de Festas: " + e);
    }
    rentabilidadeTab.renderLista();
    rentabilidadeTab.renderPreview();
}

async function exportarCSV() {
    try {
        const caminho = await ExportarCSV();
        if (caminho) setStatus(`Planilha exportada: ${caminho}`);
    } catch (e) {
        setStatus("Erro ao exportar CSV: " + e);
    }
}

async function carregarBaseClientes() {
    try {
        const resultado = await CarregarBaseClientes();
        if (!resultado) return; // usuário cancelou
        state.clientDB = resultado.ClientDB;
        state.clientEmails = resultado.ClientEmails || {};
        state.clientes = resultado.Clientes || [];
        setStatus(`Base de clientes carregada: ${Object.keys(resultado.ClientDB).length} cliente(s).`);
    } catch (e) {
        setStatus("Erro ao carregar base de clientes: " + e);
        return;
    }
    rentabilidadeTab.renderLista();
    emailsTab.refreshLookup();
}

const ctx = { setStatus, escolherPasta };

// Definição fixa de cada ferramenta da barra lateral (rótulo + ícone). A
// ordem de exibição é separada (ver ORDEM_NAV_PADRAO/ordemNavAtual), pra dar
// pra personalizar em Configurações → Ordem da barra lateral sem duplicar
// essa definição.
const ITENS_NAV = {
    rent: { label: "Rentabilidade", icon: icons.iconVelocimetro },
    mail: { label: "E-mails de Ordem", icon: icons.iconDocumento },
    des: { label: "Tabela de Deságio", icon: icons.iconDesagio },
    apres: { label: "Apresentação", icon: icons.iconApresentacao },
    typ: { label: "Typeform", icon: icons.iconFormulario },
    calc: { label: "Calculadora", icon: icons.iconCalculadora },
    cmp: { label: "Comparadora", icon: icons.iconComparadora },
    prev: { label: "Calculadora Previdenciária", icon: icons.iconCofrinho },
    comp: { label: "Compromissada", icon: icons.iconCompromissada },
    apos: { label: "Planejamento de Aposentadoria", icon: icons.iconMeta },
    pdfedit: { label: "Editor de PDF", icon: icons.iconEditarPDF },
    imgpdf: { label: "Imagens em PDF", icon: icons.iconImagensPDF },
    icp: { label: "Validar Assinatura", icon: icons.iconValidarAssinatura },
};
const ORDEM_NAV_PADRAO = [
    "rent", "mail", "des", "apres", "typ", "calc", "cmp", "prev", "comp", "apos", "pdfedit", "imgpdf", "icp",
];

// ordemNavAtual() resolve a ordem persistida (state.prefs.ordemNav) contra
// ITENS_NAV: IDs desconhecidos (aba removida numa atualização) são
// descartados; IDs novos que não estavam salvos entram no fim, na posição
// que teriam na ordem padrão — assim uma ferramenta nova sempre aparece,
// mesmo pra quem já personalizou a ordem antes dela existir.
// Ferramentas que só existem no Windows, porque dependem de programas de
// lá: o Typeform é preenchido controlando o Microsoft Edge nos caminhos de
// instalação do Windows. No macOS a aba nem entra na lista — some da barra
// lateral e de Configurações → Ordem da barra lateral.
const ABAS_SO_WINDOWS = ["typ"];

function abaDisponivelNaPlataforma(id) {
    return state.plataforma === "windows" || !ABAS_SO_WINDOWS.includes(id);
}

function ordemNavAtual() {
    const salva = state.prefs.ordemNav && state.prefs.ordemNav.length ? state.prefs.ordemNav : ORDEM_NAV_PADRAO;
    const validos = salva.filter((id) => ITENS_NAV[id] && abaDisponivelNaPlataforma(id));
    const faltando = ORDEM_NAV_PADRAO.filter((id) => !validos.includes(id) && abaDisponivelNaPlataforma(id));
    return [...validos, ...faltando];
}

// Ferramenta escondida da barra lateral pelo usuário (botão "-" em
// Configurações → Ordem da barra lateral) — separado do Modo apresentação,
// que esconde um conjunto fixo (ver ABAS_OCULTAS_APRESENTACAO).
function navItemOculto(id) {
    return (state.prefs.ordemNavOcultos || []).includes(id);
}

function montarSidebar() {
    const side = el("div", { class: "side" });

    const marca = el("div", { class: "brand" }, [
        el("img", { class: "brand-mark", src: brandIconUrl, alt: "Ferramentas de Assessoria" }),
        el("div", { class: "brand-txt" }, [el("div", { class: "brand-name", text: "Ferramentas de Assessoria" })]),
    ]);
    side.appendChild(marca);
    side.appendChild(el("div", { class: "navlbl", text: "Ferramentas" }));

    const nav = el("div", { class: "nav" });
    dom.navButtons = {};
    for (const id of ordemNavAtual()) {
        const item = ITENS_NAV[id];
        const botao = el("button", { class: "navitem" });
        botao.insertAdjacentHTML("beforeend", item.icon);
        botao.appendChild(el("span", { class: "t", text: item.label }));
        botao.addEventListener("click", () => ativarSecao(id));
        botao.style.display = navItemOculto(id) ? "none" : "";
        dom.navButtons[id] = botao;
        nav.appendChild(botao);
    }
    side.appendChild(nav);

    const rodapeSide = el("div", { class: "side-foot" });
    dom.botaoBase = el("button", { class: "navitem" });
    dom.botaoBase.insertAdjacentHTML("beforeend", icons.iconPessoa);
    dom.botaoBase.appendChild(el("span", { class: "t", text: "Base de clientes" }));
    dom.botaoBase.addEventListener("click", carregarBaseClientes);
    rodapeSide.appendChild(dom.botaoBase);

    const botaoConfig = el("button", { class: "navitem" });
    botaoConfig.insertAdjacentHTML("beforeend", icons.iconConfigNav);
    botaoConfig.appendChild(el("span", { class: "t", text: "Configurações" }));
    botaoConfig.addEventListener("click", abrirConfiguracoes);
    rodapeSide.appendChild(botaoConfig);

    // Só aparece quando EventsOn("atualizacao:disponivel") (ou "Verificar
    // agora" em Configurações → Sobre) encontra uma versão nova — ver
    // atualizarIndicadorAtualizacao(). Escondido por padrão porque a
    // checagem no startup é assíncrona e só termina alguns segundos depois
    // do app abrir (ver comentário em app.go:startup).
    dom.botaoAtualizacao = el("button", { class: "navitem navitem-atualizacao" });
    dom.botaoAtualizacao.insertAdjacentHTML("beforeend", icons.iconAtualizar);
    dom.botaoAtualizacao.appendChild(el("span", { class: "t", text: "Atualização disponível" }));
    dom.botaoAtualizacao.style.display = "none";
    dom.botaoAtualizacao.addEventListener("click", abrirModalAtualizacao);
    rodapeSide.appendChild(dom.botaoAtualizacao);

    side.appendChild(rodapeSide);

    return side;
}

function atualizarIndicadorAtualizacao() {
    if (dom.botaoAtualizacao) {
        dom.botaoAtualizacao.style.display = state.atualizacao.disponivel ? "" : "none";
    }
}

// Modal de confirmação da atualização — mesma estrutura visual de
// abrirConfiguracoes (.modal-fundo/.modal-caixa), mas sem X de fechar
// enquanto a atualização está em andamento: uma vez clicado "Atualizar
// agora", o app vai fechar e reabrir sozinho (ver AplicarAtualizacao em
// app.go), não faz sentido deixar cancelar no meio.
function abrirModalAtualizacao() {
    const { versao, notas } = state.atualizacao;

    const titulo = el("h3", { class: "cfg-h", text: `Nova versão ${versao} disponível` });
    const corpoNotas = el("p", { class: "cfg-placeholder atualizacao-notas", text: notas || "Sem notas de versão." });
    const aviso = el("p", { class: "cfg-sub", text: "" });

    const botaoAtualizar = btn("Atualizar agora", { classe: "pri", icon: icons.iconAtualizar });
    const botaoDepois = btn("Depois", { classe: "ghost" });

    botaoDepois.addEventListener("click", () => fundo.remove());
    botaoAtualizar.addEventListener("click", async () => {
        botaoAtualizar.disabled = true;
        botaoDepois.disabled = true;
        aviso.textContent = "Baixando e instalando a atualização... o app vai fechar e abrir sozinho em instantes.";
        try {
            await AplicarAtualizacao();
            // Sucesso = o processo atual está prestes a encerrar (runtime.Quit
            // do lado Go) — não há mais nada a fazer aqui.
        } catch (e) {
            aviso.textContent = "Não foi possível atualizar agora: " + e;
            botaoAtualizar.disabled = false;
            botaoDepois.disabled = false;
        }
    });

    const acoes = el("div", { class: "atualizacao-acoes" }, [botaoDepois, botaoAtualizar]);
    const caixa = el("div", { class: "modal-caixa atualizacao-caixa" }, [titulo, corpoNotas, aviso, acoes]);
    const fundo = el("div", { class: "modal-fundo" });
    fundo.appendChild(caixa);
    appRoot.appendChild(fundo);
}

// Reconstrói a barra lateral inteira (chamado ao confirmar uma nova ordem
// em Configurações → Ordem da barra lateral) e reaplica o que dependia dos
// botões antigos: aba ativa destacada e abas escondidas no modo apresentação.
function remontarSidebar() {
    const novo = montarSidebar();
    dom.side.replaceWith(novo);
    dom.side = novo;
    ativarSecao(secaoAtual);
    aplicarModoApresentacao();
}

// ---------------------------------------------------------------------------
// Tema / acento
// ---------------------------------------------------------------------------

function aplicarTema() {
    const classes = ["app", `acc-${state.prefs.acento}`];
    if (state.prefs.tema === "escuro") classes.push("dark");
    appRoot.className = classes.join(" ");
    appRoot.style.setProperty("--font-ui", state.prefs.fonte);
}

async function salvarPrefs() {
    try {
        await SalvarPreferencias(
            state.prefs.tema,
            state.prefs.acento,
            state.prefs.modoEmail,
            state.prefs.tabelaPrevidenciaria,
            state.prefs.visao,
            state.prefs.fonte,
            state.prefs.modoApresentacao,
            state.prefs.modoFestas,
            state.prefs.ordemNav,
            state.prefs.ordemNavOcultos
        );
    } catch (e) {
        setStatus("Erro ao salvar preferências: " + e);
    }
}

// ---------------------------------------------------------------------------
// Modo apresentação
// ---------------------------------------------------------------------------

// Abas escondidas no modo apresentação — informação interna (pasta de PDFs,
// e-mails de ordem, deságio, editor de PDF, imagens em PDF) que não deve
// aparecer numa tela compartilhada com o cliente.
const ABAS_OCULTAS_APRESENTACAO = ["rent", "mail", "des", "pdfedit", "imgpdf"];

function aplicarModoApresentacao() {
    const oculto = state.prefs.modoApresentacao;
    for (const id of ABAS_OCULTAS_APRESENTACAO) {
        // Some com o ocultamento manual do usuário (botão "-"): desligar o
        // modo apresentação não deve reaparecer um item escondido à parte.
        dom.navButtons[id].style.display = oculto || navItemOculto(id) ? "none" : "";
    }
    dom.botaoBase.style.display = oculto ? "none" : "";
    // Notificações internas (status de processamento, falhas de PDF) não
    // devem aparecer numa tela compartilhada com o cliente. O progresso
    // usa style.display (não o atributo hidden, que mostrarProgresso já
    // controla pra mostrar/esconder durante o processamento em si) — assim
    // os dois controles não brigam entre si.
    dom.statusbar.style.display = oculto ? "none" : "";
    dom.progresso.style.display = oculto ? "none" : "";
    // Se a aba ativa ficou escondida (modo apresentação ligado agora, ou já
    // ligado no boot com "rent" como aba inicial padrão), navega pra
    // Apresentação quando há um arquivo carregado — é a abertura institucional
    // pro cliente; sem arquivo, cai na Compromissada (ferramenta voltada pra
    // tela compartilhada).
    if (oculto && ABAS_OCULTAS_APRESENTACAO.includes(secaoAtual)) {
        ativarSecao(state.prefs.temApresentacao ? "apres" : "comp");
    }
}

const TEMAS = [
    { id: "claro", label: "Claro" },
    { id: "escuro", label: "Escuro" },
];

const ACENTOS = [
    { id: "esmeralda", label: "Esmeralda", cor: "#059669" },
    { id: "teal", label: "Teal", cor: "#0d9488" },
    { id: "aqua", label: "Aqua", cor: "#06b6d4" },
    { id: "petroleo", label: "Verde petróleo", cor: "#0f3d3e" },
];

// Valor CSS font-family (data-font) + rótulo exibido. A lista replica as
// fontes do Google Fonts carregadas em index.html.
const FONTES = [
    { valor: "'Plus Jakarta Sans', system-ui, sans-serif", label: "Plus Jakarta Sans" },
    { valor: "'Inter', sans-serif", label: "Inter" },
    { valor: "'Roboto', sans-serif", label: "Roboto" },
    { valor: "'Open Sans', sans-serif", label: "Open Sans" },
    { valor: "'Lato', sans-serif", label: "Lato" },
    { valor: "'Montserrat', sans-serif", label: "Montserrat" },
    { valor: "'Poppins', sans-serif", label: "Poppins" },
    { valor: "'Nunito', sans-serif", label: "Nunito" },
    { valor: "'Raleway', sans-serif", label: "Raleway" },
    { valor: "'Work Sans', sans-serif", label: "Work Sans" },
    { valor: "'Source Sans 3', sans-serif", label: "Source Sans 3" },
    { valor: "'Manrope', sans-serif", label: "Manrope" },
    { valor: "'DM Sans', sans-serif", label: "DM Sans" },
    { valor: "'Rubik', sans-serif", label: "Rubik" },
    { valor: "'Mulish', sans-serif", label: "Mulish" },
    { valor: "'Figtree', sans-serif", label: "Figtree" },
    { valor: "'Outfit', sans-serif", label: "Outfit" },
    { valor: "'Sora', sans-serif", label: "Sora" },
    { valor: "'Space Grotesk', sans-serif", label: "Space Grotesk" },
    { valor: "'Jost', sans-serif", label: "Jost" },
    { valor: "'Karla', sans-serif", label: "Karla" },
];

// Variáveis de CSS que o modo Beta deixa escolher individualmente, cada
// uma com o rótulo mostrado ao lado do seletor de cor.
const VARIAVEIS_BETA = [
    { var: "--sidebar", label: "Barra lateral" },
    { var: "--primary", label: "Acento" },
    { var: "--bg", label: "Fundo" },
    { var: "--surface", label: "Cards / superfície" },
    { var: "--ink", label: "Texto" },
    { var: "--ink-2", label: "Texto secundário" },
    { var: "--pos", label: "Positivo" },
    { var: "--line", label: "Bordas" },
];

// "Personalização avançada (Beta)": sobrepõe, por cima do tema/acento
// confirmados, uma cor escolhida à mão pra cada variável — aplicada direto
// (sem esperar o Confirmar de cima) como estilo inline em #app, que só
// desaparece quando o usuário clica "Redefinir cores". Não é persistida
// entre sessões (é a mesma escolha de design do mockup original: um ajuste
// de sessão, não uma preferência salva).
function montarGrupoPersonalizacaoBeta() {
    const grp = el("div", { class: "cfg-grp" });
    const cabecalho = el("div", { class: "cfg-grp-lbl beta-head" });
    cabecalho.appendChild(document.createTextNode("Personalização avançada"));
    cabecalho.appendChild(el("span", { class: "beta-badge", text: "Beta" }));
    grp.appendChild(cabecalho);

    const linhaToggle = el("div", { class: "beta-toggle-row" });
    const texto = el("div", { class: "beta-toggle-txt" });
    texto.appendChild(el("b", { text: "Escolher a cor de cada elemento" }));
    texto.appendChild(el("br"));
    texto.appendChild(
        document.createTextNode("Ajuste manualmente cada parte da interface. Sobrepõe o tema e a cor de acento até você redefinir.")
    );
    linhaToggle.appendChild(texto);
    const botaoSwitch = el("button", { class: "beta-switch", type: "button", "aria-label": "Ativar modo Beta" });
    linhaToggle.appendChild(botaoSwitch);
    grp.appendChild(linhaToggle);

    const painel = el("div", { class: "beta-panel" });
    for (const v of VARIAVEIS_BETA) {
        const linha = el("div", { class: "beta-color" });
        const valorAtual = getComputedStyle(appRoot).getPropertyValue(v.var).trim() || "#000000";
        const input = el("input", { type: "color", value: valorAtual });
        input.dataset.var = v.var;
        input.addEventListener("input", () => {
            appRoot.style.setProperty(v.var, input.value);
            if (v.var === "--primary") appRoot.style.setProperty("--primary-600", input.value);
        });
        linha.appendChild(input);
        linha.appendChild(el("span", { class: "beta-color-lbl", text: v.label }));
        painel.appendChild(linha);
    }

    const acoes = el("div", { class: "beta-actions" });
    const botaoRedefinir = el("button", { class: "btn-limpar", type: "button" });
    botaoRedefinir.insertAdjacentHTML("beforeend", icons.iconRedefinir);
    botaoRedefinir.appendChild(document.createTextNode("Redefinir cores"));
    botaoRedefinir.addEventListener("click", () => {
        for (const v of VARIAVEIS_BETA) appRoot.style.removeProperty(v.var);
        appRoot.style.removeProperty("--primary-600");
        for (const input of painel.querySelectorAll("input[type=color]")) {
            input.value = getComputedStyle(appRoot).getPropertyValue(input.dataset.var).trim();
        }
    });
    acoes.appendChild(botaoRedefinir);
    painel.appendChild(acoes);
    grp.appendChild(painel);

    // Se já existe alguma cor personalizada ativa (o usuário ligou o modo
    // Beta antes, mudou de aba e voltou em Configurações), abre o painel já
    // expandido em vez de esconder um ajuste que já está valendo.
    const jaPersonalizado = VARIAVEIS_BETA.some((v) => appRoot.style.getPropertyValue(v.var));
    painel.hidden = !jaPersonalizado;
    botaoSwitch.classList.toggle("on", jaPersonalizado);

    botaoSwitch.addEventListener("click", () => {
        const ligado = botaoSwitch.classList.toggle("on");
        painel.hidden = !ligado;
    });

    return grp;
}

function montarAbaTemas() {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "Temas" }));
    body.appendChild(el("p", { class: "cfg-sub", text: "Escolha o tema, a cor de acento e a fonte do aplicativo." }));

    // Toda mudança em Configurações só aplica ao confirmar — evita que um
    // clique acidental já mude tema/acento/fonte (ou qualquer outra
    // preferência) na hora.
    let temaEscolhido = state.prefs.tema;
    let acentoEscolhido = state.prefs.acento;
    let fonteEscolhida = state.prefs.fonte;
    const botaoConfirmar = btn("Confirmar", { classe: "pri" });
    const atualizarBotao = () => {
        botaoConfirmar.disabled =
            temaEscolhido === state.prefs.tema && acentoEscolhido === state.prefs.acento && fonteEscolhida === state.prefs.fonte;
    };

    const grpTema = el("div", { class: "cfg-grp" });
    grpTema.appendChild(el("div", { class: "cfg-grp-lbl", text: "Tema" }));
    const optsTema = el("div", { class: "theme-opts" });
    for (const t of TEMAS) {
        const opt = el("div", { class: `theme-opt${temaEscolhido === t.id ? " on" : ""}` });
        opt.appendChild(el("div", { class: `theme-opt-swatch ${t.id}` }));
        opt.appendChild(el("div", { class: "theme-opt-lbl", text: t.label }));
        opt.addEventListener("click", () => {
            temaEscolhido = t.id;
            optsTema.querySelectorAll(".theme-opt").forEach((n) => n.classList.remove("on"));
            opt.classList.add("on");
            atualizarBotao();
        });
        optsTema.appendChild(opt);
    }
    grpTema.appendChild(optsTema);
    body.appendChild(grpTema);

    const grpAcento = el("div", { class: "cfg-grp" });
    grpAcento.appendChild(el("div", { class: "cfg-grp-lbl", text: "Cor de acento" }));
    const optsAcento = el("div", { class: "acc-opts" });
    for (const a of ACENTOS) {
        const opt = el("div", { class: `acc-opt${acentoEscolhido === a.id ? " on" : ""}` });
        const swatch = el("div", { class: "acc-opt-swatch" });
        swatch.style.background = a.cor;
        opt.appendChild(swatch);
        opt.appendChild(el("div", { class: "acc-opt-lbl", text: a.label }));
        opt.addEventListener("click", () => {
            acentoEscolhido = a.id;
            optsAcento.querySelectorAll(".acc-opt").forEach((n) => n.classList.remove("on"));
            opt.classList.add("on");
            atualizarBotao();
        });
        optsAcento.appendChild(opt);
    }
    grpAcento.appendChild(optsAcento);
    body.appendChild(grpAcento);

    const grpFonte = el("div", { class: "cfg-grp" });
    grpFonte.appendChild(el("div", { class: "cfg-grp-lbl", text: "Fonte da interface" }));
    const picker = el("div", { class: "font-picker" });
    const lista = el("div", { class: "font-list" });
    const preview = el("div", { class: "font-preview" });
    const previewSub = el("div", { class: "font-preview-sub" });
    previewSub.appendChild(document.createTextNode("Em junho sua carteira teve uma rentabilidade de "));
    previewSub.appendChild(el("b", { text: "R$ 305,91" }));
    previewSub.appendChild(document.createTextNode(", equivalente a 2,14% no mês ("));
    previewSub.appendChild(el("b", { text: "118%" }));
    previewSub.appendChild(document.createTextNode(" do CDI)."));
    preview.appendChild(el("div", { class: "font-preview-eyebrow", text: "Pré-visualização" }));
    preview.appendChild(el("div", { class: "font-preview-title", text: "Rentabilidade" }));
    preview.appendChild(previewSub);
    preview.appendChild(
        el("div", { class: "font-preview-num" }, ["R$ 2.480,10 ", el("span", { text: "acumulado no ano" })])
    );

    const atualizarPreview = () => {
        preview.style.fontFamily = fonteEscolhida;
    };
    for (const f of FONTES) {
        const linha = el("button", { class: `font-row${fonteEscolhida === f.valor ? " on" : ""}`, type: "button" });
        linha.style.fontFamily = f.valor;
        linha.appendChild(document.createTextNode(f.label));
        if (f.valor === state.prefs.fonte) linha.appendChild(el("span", { class: "font-row-tag", text: "Atual" }));
        linha.addEventListener("click", () => {
            fonteEscolhida = f.valor;
            lista.querySelectorAll(".font-row").forEach((n) => n.classList.remove("on"));
            linha.classList.add("on");
            atualizarPreview();
            atualizarBotao();
        });
        lista.appendChild(linha);
    }
    atualizarPreview();
    picker.appendChild(lista);
    picker.appendChild(preview);
    grpFonte.appendChild(picker);
    body.appendChild(grpFonte);

    body.appendChild(montarGrupoPersonalizacaoBeta());

    atualizarBotao();
    botaoConfirmar.addEventListener("click", () => {
        state.prefs.tema = temaEscolhido;
        state.prefs.acento = acentoEscolhido;
        state.prefs.fonte = fonteEscolhida;
        aplicarTema();
        salvarPrefs();
        // O canvas dos gráficos não lê variáveis CSS — precisa redesenhar
        // com as cores do tema novo.
        compromissadaTab.mount(dom.secaoComp, ctx);
        aposentadoriaTab.mount(dom.secaoApos, ctx);
        atualizarBotao();
    });
    body.appendChild(el("div", { class: "cfg-confirm-row" }, [botaoConfirmar]));

    return body;
}

const MODOS_EMAIL = [
    { id: "padronizado", label: "Padronizado (recomendado)", desc: "Cada e-mail tem só um produto e tipo de movimentação, seguindo os modelos de compliance." },
    { id: "livre", label: "Livre", desc: "Cada operação escolhe seu próprio produto, permitindo misturar vários num só e-mail." },
];

function montarAbaEmail() {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "E-mail" }));
    body.appendChild(el("p", { class: "cfg-sub", text: "Escolha como o gerador de e-mails de ordem se comporta." }));

    // Troca de modo não é em tempo real (ao contrário do tema): só aplica
    // quando confirmada, porque zera as operações em andamento na aba
    // E-mails — clique acidental não pode custar o que já foi preenchido.
    let escolhido = state.prefs.modoEmail || "padronizado";

    const grp = el("div", { class: "cfg-grp" });
    grp.appendChild(el("div", { class: "cfg-grp-lbl", text: "Modo de geração" }));
    const lista = el("div", { class: "theme-opts" });
    const botaoConfirmar = btn("Confirmar", { classe: "pri" });
    for (const m of MODOS_EMAIL) {
        const opt = el("div", { class: `theme-opt${escolhido === m.id ? " on" : ""}` });
        opt.appendChild(el("div", { class: "theme-opt-lbl", text: m.label }));
        opt.appendChild(el("div", { class: "cfg-placeholder", text: m.desc }));
        opt.addEventListener("click", () => {
            escolhido = m.id;
            lista.querySelectorAll(".theme-opt").forEach((n) => n.classList.remove("on"));
            opt.classList.add("on");
            botaoConfirmar.disabled = escolhido === state.prefs.modoEmail;
        });
        lista.appendChild(opt);
    }
    grp.appendChild(lista);
    body.appendChild(grp);

    botaoConfirmar.disabled = true;
    botaoConfirmar.addEventListener("click", () => {
        state.prefs.modoEmail = escolhido;
        salvarPrefs();
        state.emailProduto = null;
        state.emailTipo = null;
        state.blocosEmail = [];
        emailsTab.mount(dom.secaoMail, ctx);
        botaoConfirmar.disabled = true;
    });
    body.appendChild(el("div", { class: "cfg-confirm-row" }, [botaoConfirmar]));

    // Remetente não zera nada em andamento (ao contrário do modo, acima) —
    // salva direto ao clicar, sem exigir confirmação com aviso de perda.
    const grpRemetente = el("div", { class: "cfg-grp" });
    grpRemetente.appendChild(el("div", { class: "cfg-grp-lbl", text: "E-mail remetente (Outlook)" }));
    grpRemetente.appendChild(el("div", {
        class: "cfg-placeholder",
        text: "Se você tem mais de uma conta logada no Outlook, informe aqui qual deve ser usada como remetente — evita que o Outlook escolha a conta errada sozinho.",
    }));
    const inputRemetente = el("input", { class: "input", type: "email", value: state.prefs.emailRemetente || "" });
    const botaoSalvarRemetente = btn("Salvar", { classe: "pri" });
    botaoSalvarRemetente.disabled = true;
    inputRemetente.addEventListener("input", () => {
        botaoSalvarRemetente.disabled = inputRemetente.value.trim() === (state.prefs.emailRemetente || "");
    });
    botaoSalvarRemetente.addEventListener("click", async () => {
        const email = inputRemetente.value.trim();
        try {
            await SalvarEmailRemetente(email);
            state.prefs.emailRemetente = email;
            botaoSalvarRemetente.disabled = true;
        } catch (e) {
            alert("Não foi possível salvar o e-mail remetente.\n\n" + e);
        }
    });
    grpRemetente.appendChild(el("div", { class: "cfg-confirm-row" }, [inputRemetente, botaoSalvarRemetente]));
    body.appendChild(grpRemetente);

    // Nome/e-mail do assessor: respondem sozinhas as duas primeiras
    // perguntas do Typeform ("Nome do assessor responsável" e "E-mail do
    // assessor") no preenchimento automático (aba Typeform, botão "Enviar
    // para o Typeform") — sem isso a automação sempre parava ali por falta
    // de resposta salva. Mesmo padrão de salvar direto do remetente acima.
    const grpAssessor = el("div", { class: "cfg-grp" });
    grpAssessor.appendChild(el("div", { class: "cfg-grp-lbl", text: "Dados do assessor (Typeform)" }));
    grpAssessor.appendChild(el("div", {
        class: "cfg-placeholder",
        text: "Usados pra responder automaticamente as duas primeiras perguntas do Typeform ao clicar em \"Enviar para o Typeform\" na aba Typeform.",
    }));
    const inputAssessorNome = el("input", { class: "input", type: "text", placeholder: "Nome do assessor", value: state.prefs.assessorNome || "" });
    const inputAssessorEmail = el("input", { class: "input", type: "email", placeholder: "E-mail do assessor", value: state.prefs.assessorEmail || "" });
    const botaoSalvarAssessor = btn("Salvar", { classe: "pri" });
    botaoSalvarAssessor.disabled = true;
    const atualizarBotaoAssessor = () => {
        botaoSalvarAssessor.disabled =
            inputAssessorNome.value.trim() === (state.prefs.assessorNome || "") &&
            inputAssessorEmail.value.trim() === (state.prefs.assessorEmail || "");
    };
    inputAssessorNome.addEventListener("input", atualizarBotaoAssessor);
    inputAssessorEmail.addEventListener("input", atualizarBotaoAssessor);
    botaoSalvarAssessor.addEventListener("click", async () => {
        const nome = inputAssessorNome.value.trim();
        const email = inputAssessorEmail.value.trim();
        try {
            await SalvarDadosAssessor(nome, email);
            state.prefs.assessorNome = nome;
            state.prefs.assessorEmail = email;
            botaoSalvarAssessor.disabled = true;
        } catch (e) {
            alert("Não foi possível salvar os dados do assessor.\n\n" + e);
        }
    });
    grpAssessor.appendChild(el("div", { class: "cfg-confirm-row" }, [inputAssessorNome, inputAssessorEmail, botaoSalvarAssessor]));
    body.appendChild(grpAssessor);

    return body;
}

const TABELAS_PREVIDENCIARIAS = [
    { id: "2026", label: "2026", desc: "Faixas de IRPF/INSS vigentes e o redutor de IR criado pela Lei 15.270/2025." },
    { id: "2022", label: "2022", desc: "Valores referentes às regras vigentes a partir do ano de 2022." },
];

function montarAbaTabelaPrevidenciaria() {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "Tabela previdenciária" }));
    body.appendChild(el("p", { class: "cfg-sub", text: "Escolha as tabelas de IRPF/INSS usadas na Calculadora Previdenciária." }));

    let escolhida = state.prefs.tabelaPrevidenciaria;

    const grp = el("div", { class: "cfg-grp" });
    grp.appendChild(el("div", { class: "cfg-grp-lbl", text: "Ano de referência" }));
    const lista = el("div", { class: "theme-opts" });
    const botaoConfirmar = btn("Confirmar", { classe: "pri" });
    for (const t of TABELAS_PREVIDENCIARIAS) {
        const opt = el("div", { class: `theme-opt${escolhida === t.id ? " on" : ""}` });
        opt.appendChild(el("div", { class: "theme-opt-lbl", text: t.label }));
        opt.appendChild(el("div", { class: "cfg-placeholder", text: t.desc }));
        opt.addEventListener("click", () => {
            escolhida = t.id;
            lista.querySelectorAll(".theme-opt").forEach((n) => n.classList.remove("on"));
            opt.classList.add("on");
            botaoConfirmar.disabled = escolhida === state.prefs.tabelaPrevidenciaria;
        });
        lista.appendChild(opt);
    }
    grp.appendChild(lista);
    body.appendChild(grp);

    botaoConfirmar.disabled = true;
    botaoConfirmar.addEventListener("click", () => {
        state.prefs.tabelaPrevidenciaria = escolhida;
        salvarPrefs();
        previdenciariaTab.mount(dom.secaoPrev, ctx);
        botaoConfirmar.disabled = true;
    });
    body.appendChild(el("div", { class: "cfg-confirm-row" }, [botaoConfirmar]));

    return body;
}

// montarAbaAssinatura difere das outras abas de Configurações por precisar
// buscar dados assíncronos (ListarAssinaturas) — devolve o corpo já
// montado (com "Carregando...") e preenche a grade depois que a chamada
// volta, em vez de esperar antes de devolver o nó.
function montarAbaAssinatura() {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "Assinatura" }));
    body.appendChild(
        el("p", {
            class: "cfg-sub",
            text:
                'Cadastre uma ou mais assinaturas por imagem e escolha qual fica ativa — a ferramenta "Imagem" do ' +
                "Editor de PDF carimba a ativa direto, sem pedir o arquivo toda vez.",
        })
    );

    const grade = el("div", { class: "assinatura-grade" });
    body.appendChild(grade);

    async function carregarLista() {
        clear(grade);
        grade.appendChild(el("p", { class: "cfg-placeholder", text: "Carregando..." }));
        let lista;
        try {
            lista = await ListarAssinaturas();
        } catch (e) {
            clear(grade);
            grade.appendChild(el("p", { class: "cfg-placeholder", text: "Erro ao carregar assinaturas: " + e }));
            return;
        }
        clear(grade);
        if (!lista || !lista.length) {
            grade.appendChild(el("p", { class: "cfg-placeholder", text: "Nenhuma assinatura cadastrada ainda." }));
            return;
        }
        for (const assinatura of lista) {
            grade.appendChild(montarCartaoAssinatura(assinatura, carregarLista));
        }
    }

    const botaoAdicionar = btn("Adicionar assinatura", {
        classe: "pri",
        icon: icons.iconMais,
        onClick: async () => {
            botaoAdicionar.disabled = true;
            try {
                const nova = await AdicionarAssinatura();
                if (nova && nova.Nome) await carregarLista();
            } catch (e) {
                alert("Não foi possível adicionar a assinatura.\n\n" + e);
            } finally {
                botaoAdicionar.disabled = false;
            }
        },
    });
    body.appendChild(el("div", { class: "cfg-confirm-row" }, [botaoAdicionar]));

    carregarLista();
    return body;
}

function montarCartaoAssinatura(assinatura, recarregar) {
    const cartao = el("div", { class: `assinatura-cartao${assinatura.Ativa ? " on" : ""}` });

    const img = el("img", { class: "assinatura-imagem" });
    img.src = "data:image;base64," + assinatura.Base64; // o navegador reconhece o formato pelos bytes, não precisa acertar o mime
    cartao.appendChild(img);

    cartao.appendChild(el("div", { class: "assinatura-nome", text: assinatura.Nome }));
    if (assinatura.Ativa) cartao.appendChild(el("div", { class: "assinatura-badge", text: "Ativa" }));

    cartao.addEventListener("click", async () => {
        if (assinatura.Ativa) return;
        await SelecionarAssinaturaAtiva(assinatura.Nome);
        await recarregar();
    });

    const botaoRemover = el("button", { class: "assinatura-remover", type: "button", title: "Remover", text: "×" });
    botaoRemover.addEventListener("click", async (e) => {
        e.stopPropagation();
        if (!confirm(`Remover a assinatura "${assinatura.Nome}"?`)) return;
        await RemoverAssinatura(assinatura.Nome);
        await recarregar();
    });
    cartao.appendChild(botaoRemover);

    return cartao;
}

const OPCOES_APRESENTACAO = [
    { id: false, label: "Desligado", desc: "Tudo normal — todas as abas aparecem na sidebar, com SPREAD e comissão visíveis na Compromissada." },
    {
        id: true,
        label: "Ligado",
        desc: 'Esconde "Rentabilidade", "E-mails de Ordem", "Tabela de Deságio" e "Base de clientes" da sidebar, some com as notificações de status/falha, esconde o SPREAD e a comissão do escritório e muda a aba Compromissada pra Visão Cliente — seguro pra compartilhar a tela com o cliente.',
    },
];

// Não existe mais uma escolha independente entre "visão cliente" e "visão
// assessor" — visao (usada por tabs/compromissada.js pra esconder
// SPREAD/comissão) é 100% derivada do Modo apresentação: ligado = cliente,
// desligado = assessor. Um único botão, sem escolha extra pro usuário.
function montarAbaApresentacao() {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "Modo apresentação" }));
    body.appendChild(el("p", { class: "cfg-sub", text: "Simplifica a sidebar pra compartilhar a tela com o cliente." }));

    let escolhido = state.prefs.modoApresentacao;

    const grp = el("div", { class: "cfg-grp" });
    grp.appendChild(el("div", { class: "cfg-grp-lbl", text: "Modo apresentação" }));
    const lista = el("div", { class: "theme-opts" });
    const botaoConfirmar = btn("Confirmar", { classe: "pri" });
    for (const o of OPCOES_APRESENTACAO) {
        const opt = el("div", { class: `theme-opt${escolhido === o.id ? " on" : ""}` });
        opt.appendChild(el("div", { class: "theme-opt-lbl", text: o.label }));
        opt.appendChild(el("div", { class: "cfg-placeholder", text: o.desc }));
        opt.addEventListener("click", () => {
            escolhido = o.id;
            lista.querySelectorAll(".theme-opt").forEach((n) => n.classList.remove("on"));
            opt.classList.add("on");
            botaoConfirmar.disabled = escolhido === state.prefs.modoApresentacao;
        });
        lista.appendChild(opt);
    }
    grp.appendChild(lista);
    body.appendChild(grp);

    botaoConfirmar.disabled = true;
    botaoConfirmar.addEventListener("click", () => {
        state.prefs.modoApresentacao = escolhido;
        state.prefs.visao = escolhido ? "cliente" : "assessor";
        salvarPrefs();
        aplicarModoApresentacao();
        compromissadaTab.mount(dom.secaoComp, ctx);
        botaoConfirmar.disabled = true;
    });
    body.appendChild(el("div", { class: "cfg-confirm-row" }, [botaoConfirmar]));

    return body;
}

const OPCOES_FESTAS = [
    { id: false, label: "Desligado", desc: "Comportamento normal — só clientes com relatório processado ganham mensagem na aba Rentabilidade." },
    {
        id: true,
        label: "Ligado",
        desc:
            'Troca a aba Rentabilidade pro modelo de festas (Modelo → "modelo_festas.txt", só com o placeholder _Nome) e libera "Copiar mensagem"/"Enviar WhatsApp" pra todo cliente da base, com ou sem relatório. Cada envio marca um selo ENVIADO na lista (limpa em "Limpar envios de Festas", no menu "⋯" da aba Rentabilidade).',
    },
];

function montarAbaRentabilidadeConfig() {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "Rentabilidade" }));
    body.appendChild(el("p", { class: "cfg-sub", text: "Opções da aba Rentabilidade." }));

    let escolhido = state.prefs.modoFestas;

    const grp = el("div", { class: "cfg-grp" });
    grp.appendChild(el("div", { class: "cfg-grp-lbl", text: "Festas" }));
    const lista = el("div", { class: "theme-opts" });
    const botaoConfirmar = btn("Confirmar", { classe: "pri" });
    for (const o of OPCOES_FESTAS) {
        const opt = el("div", { class: `theme-opt${escolhido === o.id ? " on" : ""}` });
        opt.appendChild(el("div", { class: "theme-opt-lbl", text: o.label }));
        opt.appendChild(el("div", { class: "cfg-placeholder", text: o.desc }));
        opt.addEventListener("click", () => {
            escolhido = o.id;
            lista.querySelectorAll(".theme-opt").forEach((n) => n.classList.remove("on"));
            opt.classList.add("on");
            botaoConfirmar.disabled = escolhido === state.prefs.modoFestas;
        });
        lista.appendChild(opt);
    }
    grp.appendChild(lista);
    body.appendChild(grp);

    botaoConfirmar.disabled = true;
    botaoConfirmar.addEventListener("click", () => {
        state.prefs.modoFestas = escolhido;
        salvarPrefs();
        atualizarAcoesRent();
        rentabilidadeTab.mount(dom.secaoRent, ctx);
        botaoConfirmar.disabled = true;
    });
    body.appendChild(el("div", { class: "cfg-confirm-row" }, [botaoConfirmar]));

    return body;
}

function montarAbaOrdemNav() {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "Ordem da barra lateral" }));
    body.appendChild(
        el("p", {
            class: "cfg-sub",
            text: 'Arraste para reordenar as ferramentas na barra lateral. Use o botão "-" pra esconder uma ferramenta.',
        })
    );

    let ordemLocal = [...ordemNavAtual()];
    let ocultosLocal = new Set(state.prefs.ordemNavOcultos || []);
    let indiceArrastado = null;
    const lista = el("div", { class: "ordem-lista" });
    const botaoConfirmar = btn("Confirmar", { classe: "pri" });

    function atualizarBotao() {
        const ordemMudou = ordemLocal.join() !== ordemNavAtual().join();
        const ocultosSalvos = [...(state.prefs.ordemNavOcultos || [])].sort().join();
        const ocultosMudaram = [...ocultosLocal].sort().join() !== ocultosSalvos;
        botaoConfirmar.disabled = !ordemMudou && !ocultosMudaram;
    }

    function mover(indice, delta) {
        const alvo = indice + delta;
        if (alvo < 0 || alvo >= ordemLocal.length) return;
        [ordemLocal[indice], ordemLocal[alvo]] = [ordemLocal[alvo], ordemLocal[indice]];
        renderLista();
        atualizarBotao();
    }

    function moverPara(origem, destino) {
        if (origem === destino) return;
        const [id] = ordemLocal.splice(origem, 1);
        ordemLocal.splice(destino, 0, id);
        renderLista();
        atualizarBotao();
    }

    function alternarOculto(id) {
        if (ocultosLocal.has(id)) ocultosLocal.delete(id);
        else ocultosLocal.add(id);
        renderLista();
        atualizarBotao();
    }

    function renderLista() {
        // clear() esvazia o container e reseta o scroll pro topo (mesmo problema
        // documentado em tabs/emails.js:renderBlocos) — preserva a posição do
        // painel de Configurações pra não "puxar pra cima" a cada mover/esconder.
        const scrollContainer = lista.closest(".cfg-body");
        const scrollAnterior = scrollContainer ? scrollContainer.scrollTop : 0;
        clear(lista);
        ordemLocal.forEach((id, i) => {
            const item = ITENS_NAV[id];
            const oculto = ocultosLocal.has(id);
            const linha = el("div", { class: "ordem-item" + (oculto ? " oculto" : ""), draggable: "true" });

            const alca = el("span", { class: "ordem-alca" });
            alca.insertAdjacentHTML("beforeend", icons.iconArrastar);
            linha.appendChild(alca);

            linha.insertAdjacentHTML("beforeend", item.icon);
            linha.appendChild(el("span", { class: "ordem-lbl", text: item.label }));

            const botaoToggle = el("button", {
                class: "ordem-toggle" + (oculto ? "" : " on"),
                type: "button",
                "aria-label": oculto ? `Mostrar ${item.label} na barra lateral` : `Esconder ${item.label} da barra lateral`,
            });
            botaoToggle.insertAdjacentHTML("beforeend", icons.iconMenos);
            botaoToggle.addEventListener("click", () => alternarOculto(id));

            const botaoCima = el("button", { class: "ordem-seta up", type: "button", "aria-label": `Mover ${item.label} para cima` });
            botaoCima.insertAdjacentHTML("beforeend", icons.iconChevronBaixo);
            botaoCima.disabled = i === 0;
            botaoCima.addEventListener("click", () => mover(i, -1));

            const botaoBaixo = el("button", { class: "ordem-seta", type: "button", "aria-label": `Mover ${item.label} para baixo` });
            botaoBaixo.insertAdjacentHTML("beforeend", icons.iconChevronBaixo);
            botaoBaixo.disabled = i === ordemLocal.length - 1;
            botaoBaixo.addEventListener("click", () => mover(i, 1));

            linha.appendChild(el("div", { class: "ordem-acoes" }, [botaoToggle, botaoCima, botaoBaixo]));

            linha.addEventListener("dragstart", (e) => {
                indiceArrastado = i;
                linha.classList.add("arrastando");
                e.dataTransfer.effectAllowed = "move";
            });
            linha.addEventListener("dragend", () => {
                indiceArrastado = null;
                linha.classList.remove("arrastando");
            });
            linha.addEventListener("dragover", (e) => e.preventDefault());
            linha.addEventListener("drop", (e) => {
                e.preventDefault();
                if (indiceArrastado === null) return;
                moverPara(indiceArrastado, i);
            });

            lista.appendChild(linha);
        });
        if (scrollContainer) scrollContainer.scrollTop = scrollAnterior;
    }
    renderLista();
    body.appendChild(lista);

    atualizarBotao();
    botaoConfirmar.addEventListener("click", () => {
        state.prefs.ordemNav = [...ordemLocal];
        state.prefs.ordemNavOcultos = [...ocultosLocal];
        salvarPrefs();
        remontarSidebar();
        atualizarBotao();
    });
    body.appendChild(el("div", { class: "cfg-confirm-row" }, [botaoConfirmar]));

    return body;
}

function montarAbaSobre() {
    const body = el("div", {});

    body.appendChild(
        el("div", { style: "display:flex;align-items:center;gap:11px;margin-bottom:3px;" }, [
            el("img", { src: brandIconUrl, alt: "Ferramentas de Assessoria", style: "width:38px;height:38px;border-radius:11px;flex:none;" }),
            el("h3", { class: "cfg-h", style: "margin:0;", text: "Ferramentas de Assessoria" }),
        ])
    );
    body.appendChild(el("p", { class: "cfg-sub", style: "margin:0;", text: `Versão ${state.versaoApp || "desconhecida"}` }));
    body.appendChild(el("p", { class: "cfg-sub", style: "margin:3px 0 0;", text: "Desenvolvido por Gustavo De Medeiros" }));

    const resultado = el("p", { class: "cfg-sub" });
    const botaoVerificar = btn("Verificar atualizações agora", { icon: icons.iconAtualizar });
    botaoVerificar.addEventListener("click", async () => {
        botaoVerificar.disabled = true;
        resultado.textContent = "Verificando...";
        try {
            const dto = await VerificarAtualizacao();
            if (dto.Erro) {
                resultado.textContent = dto.Erro;
            } else if (dto.Disponivel) {
                state.atualizacao = { disponivel: true, versao: dto.Versao, notas: dto.Notas };
                atualizarIndicadorAtualizacao();
                resultado.textContent = `Versão ${dto.Versao} disponível — use o botão "Atualização disponível" na barra lateral.`;
            } else {
                resultado.textContent = "Você já está na versão mais recente.";
            }
        } catch (e) {
            resultado.textContent = "Não foi possível verificar agora: " + e;
        }
        botaoVerificar.disabled = false;
    });
    body.appendChild(el("div", { style: "margin-top:13px;" }, [botaoVerificar]));
    body.appendChild(resultado);

    const grupos = [
        [
            "Descrição",
            "Ferramenta de apoio à rotina de assessoria, com recursos de geração de e-mails, processamento de documentos e validação de assinaturas digitais. Todo o processamento é executado localmente na máquina do usuário.",
        ],
        [
            "Suporte e sugestões",
            "Relatos de bug e pedidos de novas funcionalidades podem ser enviados diretamente ao desenvolvedor. As correções são feitas conforme disponibilidade, fora do horário dedicado às atividades de assessor comercial.",
        ],
        [
            "Aviso legal",
            "Software independente, desenvolvido em caráter pessoal e sem fins comerciais. Não foi desenvolvido, homologado, mantido ou divulgado pelo time de tecnologia da XP Investimentos. O uso é por conta e risco do usuário.",
        ],
    ];
    grupos.forEach(([titulo, texto], i) => {
        const grp = el("div", { class: "cfg-grp", style: i === 0 ? "margin-top:22px;" : "" });
        grp.appendChild(el("div", { class: "cfg-grp-lbl", text: titulo }));
        grp.appendChild(el("p", { class: "cfg-placeholder", style: "margin:0;", text: texto }));
        body.appendChild(grp);
    });

    body.appendChild(
        el("p", {
            class: "cfg-sub",
            style: "margin:0; padding-top:14px; border-top:1px solid var(--line-2);",
            text: "© 2026 Gustavo De Medeiros — Uso interno, sem distribuição.",
        })
    );

    return body;
}

function montarAbaFalhas() {
    const body = el("div", {});
    body.appendChild(el("h3", { class: "cfg-h", text: "Arquivos com falha" }));
    body.appendChild(
        el("p", { class: "cfg-sub", text: "PDFs que não puderam ser processados na última leitura da pasta de Rentabilidade." })
    );

    if (!state.falhasProcessamento.length) {
        body.appendChild(el("p", { class: "cfg-placeholder", text: "Nenhuma falha registrada na última leitura." }));
        return body;
    }

    const lista = el("div", { class: "falha-lista" });
    for (const falha of state.falhasProcessamento) {
        const item = el("div", { class: "falha-item" });
        item.insertAdjacentHTML("beforeend", icons.iconAlerta);
        item.appendChild(
            el("div", {}, [
                el("div", { class: "falha-arquivo", text: falha.Arquivo }),
                el("div", { class: "falha-erro", text: falha.Erro }),
            ])
        );
        lista.appendChild(item);
    }
    body.appendChild(lista);

    return body;
}

function abrirConfiguracoes() {
    const abas = [
        { id: "email", label: "E-mail", icon: icons.iconEmail, montar: montarAbaEmail },
        { id: "temas", label: "Temas", icon: icons.iconConfig, montar: montarAbaTemas },
        { id: "prev", label: "Tabela previdenciária", icon: icons.iconPrevidencia, montar: montarAbaTabelaPrevidenciaria },
        { id: "rentabilidade", label: "Rentabilidade", icon: icons.iconFesta, montar: montarAbaRentabilidadeConfig },
        { id: "apresentacao", label: "Modo apresentação", icon: icons.iconApresentacao, montar: montarAbaApresentacao },
        { id: "recorte", label: "Recorte da imagem", icon: icons.iconImagem, montar: montarAbaRecorteImagem },
        { id: "assinatura", label: "Assinatura", icon: icons.iconAssinatura, montar: montarAbaAssinatura },
        { id: "ordemNav", label: "Ordem da barra lateral", icon: icons.iconLista, montar: montarAbaOrdemNav },
        { id: "sobre", label: "Sobre", icon: icons.iconInfo, montar: montarAbaSobre },
        { id: "falhas", label: "Arquivos com falha", icon: icons.iconAlerta, montar: montarAbaFalhas },
    ];
    let abaAtual = "email";

    const nav = el("div", { class: "cfg-nav" });
    const corpo = el("div", { class: "cfg-body" });
    const navButtons = {};

    function renderAba() {
        // clear() esvazia o container e reseta o scroll pro topo (mesmo problema
        // documentado em tabs/emails.js:renderBlocos) — preserva a posição pra
        // não "puxar pra cima" a cada clique dentro de Configurações.
        const scrollAnterior = corpo.scrollTop;
        clear(corpo);
        for (const [id, botao] of Object.entries(navButtons)) botao.classList.toggle("on", id === abaAtual);
        corpo.appendChild(abas.find((a) => a.id === abaAtual).montar(ctx));
        corpo.scrollTop = scrollAnterior;
    }

    for (const aba of abas) {
        const botao = el("button", { class: "cfg-nav-item" });
        botao.insertAdjacentHTML("beforeend", aba.icon);
        botao.appendChild(el("span", { text: aba.label }));
        botao.addEventListener("click", () => {
            abaAtual = aba.id;
            renderAba();
        });
        navButtons[aba.id] = botao;
        nav.appendChild(botao);
    }

    const caixa = el("div", { class: "modal-caixa cfg-caixa" }, [nav, corpo]);
    const fundo = el("div", { class: "modal-fundo" });

    const botaoFechar = el("button", { class: "cfg-close" });
    botaoFechar.insertAdjacentHTML("beforeend", icons.iconFechar);
    botaoFechar.addEventListener("click", () => fundo.remove());
    caixa.appendChild(botaoFechar);

    fundo.appendChild(caixa);
    fundo.addEventListener("click", (e) => {
        if (e.target === fundo) fundo.remove();
    });
    // Anexado dentro de #app (não em document.body): as variáveis de tema
    // (--surface, --line, ...) são definidas na própria .app e só chegam a
    // descendentes dela — fora disso o popup fica com fundo transparente.
    appRoot.appendChild(fundo);

    renderAba();
}

function montarHeader() {
    dom.hdrTitulo = el("h1", { class: "hdr-t" });
    dom.hdrSubtitulo = el("div", { class: "hdr-sub" });
    const hdrL = el("div", { class: "hdr-l" }, [dom.hdrTitulo, dom.hdrSubtitulo]);

    dom.chipPasta = el("span", { class: "path" });
    const chip = el("span", { class: "chip soft" });
    chip.insertAdjacentHTML("beforeend", icons.iconPasta);
    chip.appendChild(dom.chipPasta);

    dom.btnAtualizar = btn("Atualizar", { icon: icons.iconAtualizar, onClick: processarPastaAtual });

    dom.btnExportar = el("button", { class: "menu-item" });
    dom.btnExportar.insertAdjacentHTML("beforeend", icons.iconExportar);
    dom.btnExportar.appendChild(document.createTextNode("Exportar CSV"));
    dom.btnExportar.addEventListener("click", () => {
        menu.hidden = true;
        exportarCSV();
    });

    const btnTrocarPasta = el("button", { class: "menu-item" });
    btnTrocarPasta.insertAdjacentHTML("beforeend", icons.iconPasta);
    btnTrocarPasta.appendChild(document.createTextNode("Trocar pasta"));
    btnTrocarPasta.addEventListener("click", () => {
        menu.hidden = true;
        escolherPasta();
    });

    // Só aparece com o Modo Festas ligado (ver atualizarAcoesRent) — não faz
    // sentido fora dele, já que "Limpar tudo" ao lado cuida dos registros de
    // rentabilidade normais.
    dom.btnLimparFestas = el("button", { class: "menu-item" });
    dom.btnLimparFestas.insertAdjacentHTML("beforeend", icons.iconLixeira);
    dom.btnLimparFestas.appendChild(document.createTextNode("Limpar envios de Festas"));
    dom.btnLimparFestas.addEventListener("click", () => {
        menu.hidden = true;
        limparFestasEnviados();
    });

    dom.btnLimpar = el("button", { class: "menu-item danger" });
    dom.btnLimpar.insertAdjacentHTML("beforeend", icons.iconLixeira);
    dom.btnLimpar.appendChild(document.createTextNode("Limpar tudo"));
    dom.btnLimpar.addEventListener("click", () => {
        menu.hidden = true;
        limparTudo();
    });

    const menu = el("div", { class: "menu" }, [
        dom.btnExportar,
        btnTrocarPasta,
        el("div", { class: "menu-sep" }),
        dom.btnLimparFestas,
        dom.btnLimpar,
    ]);
    menu.hidden = true;
    menu.addEventListener("click", (e) => e.stopPropagation());

    const botaoMais = el("button", { class: "btn icon", title: "Mais ações" });
    botaoMais.insertAdjacentHTML("beforeend", icons.iconMaisPontos);
    botaoMais.addEventListener("click", (e) => {
        e.stopPropagation();
        menu.hidden = !menu.hidden;
    });
    document.addEventListener("click", () => {
        menu.hidden = true;
    });

    const menuWrap = el("div", { class: "menu-wrap" }, [botaoMais, menu]);

    dom.hdrR = el("div", { class: "hdr-r" }, [chip, dom.btnAtualizar, menuWrap]);

    return el("div", { class: "hdr" }, [hdrL, dom.hdrR]);
}

function montarProgresso() {
    dom.progressoLabel = el("div", { class: "rotulo" });
    dom.progressoBar = el("progress");
    const wrap = el("div", { class: "barra-progresso" }, [dom.progressoLabel, dom.progressoBar]);
    wrap.hidden = true;
    dom.progresso = wrap;
    return wrap;
}

function montarConteudo() {
    dom.secaoRent = el("div", { class: "rent" });
    dom.secaoMail = el("div", { class: "mail" });
    dom.secaoDes = el("div", { class: "des" });
    dom.secaoCalc = el("div", { class: "calc" });
    dom.secaoCmp = el("div", { class: "cmpx-secao" });
    dom.secaoPrev = el("div", { class: "prev-secao" });
    dom.secaoComp = el("div", { class: "comp-secao" });
    dom.secaoApos = el("div", { class: "comp-secao" });
    dom.secaoPdfEdit = el("div", { class: "apres-secao" });
    dom.secaoImgPdf = el("div", { class: "apres-secao" });
    dom.secaoApres = el("div", { class: "apres-secao" });
    dom.secaoTyp = el("div", { class: "comp-secao" });
    dom.secaoIcp = el("div", { class: "icp-secao" });
    return el("div", { class: "content" }, [
        dom.secaoRent,
        dom.secaoMail,
        dom.secaoDes,
        dom.secaoCalc,
        dom.secaoCmp,
        dom.secaoPrev,
        dom.secaoComp,
        dom.secaoApos,
        dom.secaoPdfEdit,
        dom.secaoImgPdf,
        dom.secaoApres,
        dom.secaoTyp,
        dom.secaoIcp,
    ]);
}

function ativarSecao(id) {
    secaoAtual = id;
    for (const [chave, botao] of Object.entries(dom.navButtons)) botao.classList.toggle("on", chave === id);

    dom.secaoRent.style.display = id === "rent" ? "flex" : "none";
    dom.secaoMail.style.display = id === "mail" ? "flex" : "none";
    dom.secaoDes.style.display = id === "des" ? "flex" : "none";
    dom.secaoCalc.style.display = id === "calc" ? "flex" : "none";
    dom.secaoCmp.style.display = id === "cmp" ? "flex" : "none";
    dom.secaoPrev.style.display = id === "prev" ? "flex" : "none";
    dom.secaoComp.style.display = id === "comp" ? "flex" : "none";
    dom.secaoApos.style.display = id === "apos" ? "flex" : "none";
    dom.secaoPdfEdit.style.display = id === "pdfedit" ? "flex" : "none";
    dom.secaoImgPdf.style.display = id === "imgpdf" ? "flex" : "none";
    dom.secaoApres.style.display = id === "apres" ? "flex" : "none";
    dom.secaoTyp.style.display = id === "typ" ? "flex" : "none";
    dom.secaoIcp.style.display = id === "icp" ? "flex" : "none";
    dom.hdrR.style.display = id === "rent" ? "flex" : "none";

    const [titulo, subtitulo] = TITULOS[id];
    dom.hdrTitulo.textContent = titulo;
    dom.hdrSubtitulo.textContent = subtitulo;
    dom.hdrSubtitulo.style.display = subtitulo ? "" : "none";
}

async function main() {
    clear(appRoot);
    aplicarTema();

    // Antes de montar a barra lateral: é a plataforma que decide se a aba
    // Typeform entra na lista (ela depende do Microsoft Edge instalado nos
    // caminhos do Windows, que não existem no macOS).
    try {
        state.plataforma = await Plataforma();
    } catch {
        state.plataforma = "windows"; // esconde o mínimo possível se a chamada falhar
    }

    dom.side = montarSidebar();
    appRoot.appendChild(dom.side);

    const main = el("div", { class: "main" });
    main.appendChild(montarHeader());
    main.appendChild(montarProgresso());
    main.appendChild(montarConteudo());

    dom.statusTexto = el("span", { text: "Iniciando..." });
    dom.statusbar = el("div", { class: "statusbar" }, [el("span", { class: "dot" }), dom.statusTexto]);
    main.appendChild(dom.statusbar);

    appRoot.appendChild(main);
    ativarSecao("rent");

    EventsOn("processamento:progresso", (feitos, total) => {
        if (total === 0) {
            // Sinal especial: o motor de leitura de PDF ainda está
            // inicializando (só acontece na primeira vez que o app roda
            // numa máquina — depois disso fica cacheado e é quase
            // instantâneo). Barra indeterminada em vez de "0/0".
            dom.progressoBar.removeAttribute("value");
            dom.progressoLabel.textContent = "Preparando leitor de PDF (só na primeira vez)...";
            return;
        }
        dom.progressoBar.max = total;
        dom.progressoBar.value = feitos;
        dom.progressoLabel.textContent = `Processando arquivos (${feitos}/${total})...`;
    });

    EventsOn("atualizacao:disponivel", (dto) => {
        state.atualizacao = { disponivel: true, versao: dto.Versao, notas: dto.Notas };
        atualizarIndicadorAtualizacao();
    });

    try {
        state.versaoApp = await VersaoAtual();
    } catch {
        state.versaoApp = "";
    }

    let inicio;
    try {
        inicio = await EstadoInicial();
    } catch (e) {
        setStatus("Erro ao iniciar: " + e);
        inicio = { TemPasta: false, ClientDB: {} };
    }
    state.pasta = inicio.TemPasta ? inicio.Pasta : null;
    state.modelo = inicio.Modelo || "";
    state.modeloFestas = inicio.ModeloFestas || "";
    state.clientDB = inicio.ClientDB || {};
    state.clientEmails = inicio.ClientEmails || {};
    state.prefs = {
        tema: inicio.Prefs?.Tema || "claro",
        acento: inicio.Prefs?.Acento || "esmeralda",
        modoEmail: inicio.Prefs?.ModoEmail || "padronizado",
        tabelaPrevidenciaria: inicio.Prefs?.TabelaPrevidenciaria || "2026",
        // visao não é mais escolhida à parte — é 100% derivada do Modo
        // apresentação (ver montarAbaApresentacao), então é recalculada aqui
        // a cada boot pra corrigir prefs salvas antes dessa mudança (onde as
        // duas podiam estar dessincronizadas).
        visao: (inicio.Prefs?.ModoApresentacao || false) ? "cliente" : "assessor",
        fonte: inicio.Prefs?.Fonte || "'Plus Jakarta Sans', system-ui, sans-serif",
        modoApresentacao: inicio.Prefs?.ModoApresentacao || false,
        modoFestas: inicio.Prefs?.ModoFestas || false,
        emailRemetente: inicio.Prefs?.EmailRemetente || "",
        assessorNome: inicio.Prefs?.AssessorNome || "",
        assessorEmail: inicio.Prefs?.AssessorEmail || "",
        ordemNav: inicio.Prefs?.OrdemNav || [],
        ordemNavOcultos: inicio.Prefs?.OrdemNavOcultos || [],
        temApresentacao: inicio.Prefs?.TemApresentacao || false,
        recortePersonalizado: inicio.Prefs?.RecortePersonalizado || false,
        recorteX0: inicio.Prefs?.RecorteX0 || 0,
        recorteY0: inicio.Prefs?.RecorteY0 || 0,
        recorteX1: inicio.Prefs?.RecorteX1 || 0,
        recorteY1: inicio.Prefs?.RecorteY1 || 0,
        recortePadraoX0: inicio.Prefs?.RecortePadraoX0 || 0,
        recortePadraoY0: inicio.Prefs?.RecortePadraoY0 || 0,
        recortePadraoX1: inicio.Prefs?.RecortePadraoX1 || 0,
        recortePadraoY1: inicio.Prefs?.RecortePadraoY1 || 0,
    };
    aplicarTema();
    atualizarAcoesRent();
    aplicarModoApresentacao();

    try {
        state.catalogoEmail = await CategoriasEmail();
    } catch (e) {
        setStatus("Erro ao carregar catálogo de e-mails: " + e);
        state.catalogoEmail = { Produtos: [], Categorias: [], InfoEstruturadas: "" };
    }

    rentabilidadeTab.mount(dom.secaoRent, ctx);
    emailsTab.mount(dom.secaoMail, ctx);
    desagioTab.mount(dom.secaoDes, ctx);
    calculadoraTab.mount(dom.secaoCalc, ctx);
    comparadoraTab.mount(dom.secaoCmp, ctx);
    previdenciariaTab.mount(dom.secaoPrev, ctx);
    compromissadaTab.mount(dom.secaoComp, ctx);
    aposentadoriaTab.mount(dom.secaoApos, ctx);
    editorPdfTab.mount(dom.secaoPdfEdit);
    imagensPdfTab.mount(dom.secaoImgPdf);
    apresentacaoTab.mount(dom.secaoApres, ctx);
    typeformTab.mount(dom.secaoTyp, ctx);
    validacaoAssinaturaTab.mount(dom.secaoIcp, ctx);

    if (state.pasta) {
        await processarPastaAtual();
    } else {
        setStatus("Pronto.");
    }
}

main();
