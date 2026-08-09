import {
    EscolherArquivosAssinatura,
    ValidarAssinaturaICP,
    ObterInfoCadeiaICP,
    AtualizarCadeiaICP,
} from "../../wailsjs/go/main/App.js";
import { BrowserOpenURL } from "../../wailsjs/runtime/runtime.js";
import { el, clear, btn } from "../ui/components.js";
import * as icons from "../ui/icons.js";

let refs = {};
let ctxApp = null;

const FORMATOS = {
    CAdESDestacado: "Assinatura CAdES destacada (.p7s/.p7m + documento original)",
    CAdESAnexado: "Assinatura CAdES anexada (.p7s/.p7m)",
    PAdES: "PDF assinado (PAdES)",
    Desconhecido: "Formato não reconhecido",
};

// Estados neutros (ACNaoReconhecida, ErroRevogacao, FormatoNaoSuportado) não
// são reprovações — são incerteza ou limitação da ferramenta. Só Valida é
// verde e só Invalida/CertificadoExpirado são vermelho.
const ESTADOS = {
    Valida: { classe: "valida", label: "Assinatura válida", icon: icons.iconCheck },
    Invalida: { classe: "invalida", label: "Assinatura inválida", icon: icons.iconAlerta },
    CertificadoExpirado: { classe: "invalida", label: "Certificado expirado", icon: icons.iconAlerta },
    ACNaoReconhecida: { classe: "neutro", label: "AC emissora não reconhecida na cadeia local", icon: icons.iconInfo },
    ErroRevogacao: { classe: "neutro", label: "Não foi possível checar revogação", icon: icons.iconInfo },
    FormatoNaoSuportado: { classe: "neutro", label: "Não foi possível validar este arquivo", icon: icons.iconInfo },
};

export function mount(container, ctx) {
    clear(container);
    refs = {};
    ctxApp = ctx;

    // O resultado da validação (nome, CPF etc. do signatário) vive só no
    // DOM enquanto a aba está montada — nunca localStorage/sessionStorage
    // (o WebView2 persiste os dois em disco). Selecionar um novo arquivo
    // sobrescreve o card de resultado; não existe histórico.
    container.appendChild(
        el("div", { class: "card icp-intro" }, [
            el("p", {
                text: "A verificação é feita localmente neste computador, sem enviar o documento a nenhum servidor — o certificado do signatário já vem embutido na própria assinatura digital.",
            }),
        ])
    );

    container.appendChild(
        el("div", { class: "card icp-selecao" }, [
            el("p", { class: "field-lbl", text: "Assinatura digital" }),
            el("p", {
                text: "Selecione um PDF assinado, ou o arquivo .p7s/.p7m — se a assinatura for destacada, selecione também o documento original junto (Ctrl+clique nos dois, no mesmo diálogo).",
            }),
            btn("Selecionar arquivo(s)...", { classe: "pri", icon: icons.iconPasta, onClick: selecionarArquivos }),
        ])
    );

    refs.carregando = el("div", { class: "icp-carregando" }, [
        el("div", { class: "spinner" }),
        el("span", { text: "Validando assinatura..." }),
    ]);
    refs.carregando.hidden = true;
    container.appendChild(refs.carregando);

    refs.resultadoWrap = el("div", { class: "icp-resultado-wrap" });
    refs.resultadoWrap.hidden = true;
    container.appendChild(refs.resultadoWrap);

    refs.rodape = el("div", { class: "icp-rodape" });
    container.appendChild(refs.rodape);

    carregarInfoCadeia();
}

function elIcon(svg, classe) {
    const span = el("span", { class: classe || "icp-icon" });
    span.insertAdjacentHTML("beforeend", svg);
    return span;
}

async function selecionarArquivos() {
    let caminhos;
    try {
        caminhos = await EscolherArquivosAssinatura();
    } catch (e) {
        alert("Não foi possível abrir o diálogo de seleção.\n\n" + e);
        return;
    }
    if (!caminhos || caminhos.length === 0) return; // usuário cancelou o diálogo

    refs.resultadoWrap.hidden = true;
    mostrarCarregando(true);
    try {
        const resultado = await ValidarAssinaturaICP(caminhos);
        renderResultado(resultado);
    } catch (e) {
        renderErro(String(e));
    } finally {
        mostrarCarregando(false);
    }
}

function mostrarCarregando(visivel) {
    if (refs.carregando) refs.carregando.hidden = !visivel;
}

function renderErro(mensagem) {
    clear(refs.resultadoWrap);
    refs.resultadoWrap.hidden = false;
    refs.resultadoWrap.appendChild(
        el("div", { class: "card icp-resultado" }, [
            el("div", { class: "icp-estado invalida" }, [
                elIcon(icons.iconAlerta),
                el("div", { class: "icp-estado-txt" }, [
                    el("div", { class: "icp-estado-titulo", text: "Não foi possível validar" }),
                    el("div", { class: "icp-estado-motivo", text: mensagem }),
                ]),
            ]),
        ])
    );
}

function renderResultado(r) {
    clear(refs.resultadoWrap);
    refs.resultadoWrap.hidden = false;

    const cfg = ESTADOS[r.Estado] || ESTADOS.FormatoNaoSuportado;

    const txtBanner = [el("div", { class: "icp-estado-titulo", text: cfg.label })];
    if (r.Motivo) txtBanner.push(el("div", { class: "icp-estado-motivo", text: r.Motivo }));

    const card = el("div", { class: "card icp-resultado" }, [
        el("div", { class: `icp-estado ${cfg.classe}` }, [elIcon(cfg.icon), el("div", { class: "icp-estado-txt" }, txtBanner)]),
    ]);

    if (r.Estado !== "FormatoNaoSuportado") {
        card.appendChild(montarGrid(r));
    }
    if (r.Verificacoes && r.Verificacoes.length) {
        card.appendChild(montarChecklist(r.Verificacoes));
    }

    refs.resultadoWrap.appendChild(card);
    refs.resultadoWrap.appendChild(montarLinkOficial());
}

function linhaCampo(rotulo, valorNode) {
    return el("div", { class: "crow" }, [el("span", { class: "clbl", text: rotulo }), valorNode]);
}

function textoCampo(texto) {
    return el("span", { class: "cout", text: texto || "—" });
}

function montarGrid(r) {
    const grid = el("div", { class: "icp-grid" });
    grid.appendChild(linhaCampo("Documento", textoCampo(FORMATOS[r.Formato] || r.Formato)));
    grid.appendChild(linhaCampo("Signatário", textoCampo(r.NomeSignatario)));
    if (r.CPF) grid.appendChild(linhaCampo("CPF", montarCPFReveal(r.CPF)));
    if (r.CNPJ) grid.appendChild(linhaCampo("CNPJ", textoCampo(formatarCNPJ(r.CNPJ))));
    grid.appendChild(linhaCampo("AC emissora", textoCampo(r.ACEmissora)));
    grid.appendChild(linhaCampo("Data da assinatura", textoCampo(r.DataAssinatura || "não informada")));
    grid.appendChild(linhaCampo("Carimbo de tempo", textoCampo(r.TemCarimboTempo ? "Sim" : "Não")));
    return grid;
}

function formatarCPFMascarado(cpf) {
    if (!cpf || cpf.length !== 11) return cpf || "—";
    return `•••.${cpf.slice(3, 6)}.•••-${cpf.slice(9, 11)}`;
}

function formatarCPFCompleto(cpf) {
    if (!cpf || cpf.length !== 11) return cpf || "—";
    return `${cpf.slice(0, 3)}.${cpf.slice(3, 6)}.${cpf.slice(6, 9)}-${cpf.slice(9, 11)}`;
}

function formatarCNPJ(cnpj) {
    if (!cnpj || cnpj.length !== 14) return cnpj || "—";
    return `${cnpj.slice(0, 2)}.${cnpj.slice(2, 5)}.${cnpj.slice(5, 8)}/${cnpj.slice(8, 12)}-${cnpj.slice(12, 14)}`;
}

function montarCPFReveal(cpf) {
    const wrap = el("span", { class: "icp-cpf" });
    const texto = el("span", { class: "cout", text: formatarCPFMascarado(cpf) });
    const botao = el("button", { class: "icp-cpf-btn", title: "Mostrar CPF completo" });
    botao.insertAdjacentHTML("beforeend", icons.iconOlho);
    let revelado = false;
    botao.addEventListener("click", () => {
        revelado = !revelado;
        texto.textContent = revelado ? formatarCPFCompleto(cpf) : formatarCPFMascarado(cpf);
        botao.title = revelado ? "Ocultar CPF" : "Mostrar CPF completo";
    });
    wrap.appendChild(texto);
    wrap.appendChild(botao);
    return wrap;
}

function montarChecklist(verificacoes) {
    const lista = el("div", { class: "icp-checklist" });
    for (const v of verificacoes) {
        const item = el("div", { class: "falha-item" }, [
            elIcon(v.Passou ? icons.iconCheck : icons.iconAlerta),
            el("div", {}, [el("div", { class: "falha-arquivo", text: v.Nome }), el("div", { class: "falha-erro", text: v.Detalhe })]),
        ]);
        if (v.Passou) item.classList.add("icp-check-ok");
        lista.appendChild(item);
    }
    return lista;
}

function montarLinkOficial() {
    return el("div", { class: "icp-oficial" }, [
        el("p", {
            text: "Esta é uma conferência interna. Para um documento com valor probatório, obtenha o relatório oficial do ITI:",
        }),
        btn("Abrir validar.iti.gov.br", { classe: "ghost", onClick: () => BrowserOpenURL("https://validar.iti.gov.br") }),
    ]);
}

async function carregarInfoCadeia() {
    try {
        renderRodape(await ObterInfoCadeiaICP());
    } catch (e) {
        renderRodape(null);
    }
}

function renderRodape(info) {
    if (!refs.rodape) return;
    clear(refs.rodape);
    if (!info) {
        refs.rodape.appendChild(el("span", { class: "icp-rodape-txt", text: "Cadeia ICP-Brasil indisponível." }));
        return;
    }
    const origemTxt = info.Origem === "embutida" ? "embutida no app" : "baixada do ITI";
    const texto = info.AtualizadoEm
        ? `Cadeia ICP-Brasil ${origemTxt}, de ${info.AtualizadoEm} — ${info.NumCertificados} certificados.`
        : `Cadeia ICP-Brasil — ${info.NumCertificados} certificados.`;
    refs.rodape.appendChild(el("span", { class: "icp-rodape-txt", text: texto }));
    refs.rodape.appendChild(btn("Atualizar cadeia", { classe: "ghost", onClick: atualizarCadeia }));
    if (info.Erro) {
        refs.rodape.appendChild(el("span", { class: "icp-rodape-erro", text: info.Erro }));
    }
}

async function atualizarCadeia() {
    ctxApp?.setStatus("Atualizando cadeia ICP-Brasil...");
    try {
        const info = await AtualizarCadeiaICP();
        renderRodape(info);
        ctxApp?.setStatus(info.Erro ? "Não foi possível atualizar a cadeia agora." : "Cadeia ICP-Brasil atualizada.");
    } catch (e) {
        ctxApp?.setStatus("Erro ao atualizar cadeia: " + e);
    }
}
