import { state, novoBlocoId } from "../state.js";
import { el, clear, btn } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import { GerarEmail, CopiarTexto, AbrirEmailNoOutlook } from "../../wailsjs/go/main/App.js";

let refs = {};

// "padronizado" (padrão, regra de compliance: um produto por e-mail, com
// os selects de Produto/Tipo abaixo do cliente) ou "livre" (comportamento
// antigo: cada operação escolhe seu próprio produto/tipo, pode misturar).
// Trocado pela aba Configurações > E-mail (ver main.js).
function modoAtual() {
    return state.prefs.modoEmail === "livre" ? "livre" : "padronizado";
}

export function mount(container, ctx) {
    clear(container);
    refs = {};
    const modo = modoAtual();

    // ---- painel esquerdo ----
    const esquerdo = el("div", { class: "mail-l" });

    esquerdo.appendChild(el("div", { class: "field-lbl", text: "Código do cliente (conta XP)" }));
    refs.codigo = el("input", { class: "input mono", type: "text" });
    refs.codigo.addEventListener("input", lookupCliente);
    esquerdo.appendChild(refs.codigo);

    esquerdo.appendChild(el("div", { class: "field-lbl", text: "Nome do cliente" }));
    refs.nome = el("input", { class: "input", type: "text" });
    esquerdo.appendChild(refs.nome);

    refs.statusLookup = el("div", { class: "lookup" });
    esquerdo.appendChild(refs.statusLookup);

    esquerdo.appendChild(el("div", { class: "field-lbl", text: "E-mail do cliente" }));
    refs.email = el("input", { class: "input", type: "email" });
    esquerdo.appendChild(refs.email);

    refs.statusLookupEmail = el("div", { class: "lookup" });
    esquerdo.appendChild(refs.statusLookupEmail);

    if (modo === "padronizado") {
        // Produto + tipo de movimentação valem pro e-mail inteiro (regra de
        // compliance: um e-mail só pode ter operações do mesmo produto/tipo).
        esquerdo.appendChild(el("div", { class: "field-lbl", text: "Produto e tipo de movimentação" }));
        refs.selectProduto = el("select", { class: "select" });
        refs.selectTipo = el("select", { class: "select" });
        refs.selectProduto.addEventListener("change", () => {
            state.emailProduto = refs.selectProduto.value;
            const categorias = categoriasDoProduto(state.emailProduto, modo);
            state.emailTipo = categorias[0]?.Label || null;
            state.blocosEmail = state.emailTipo ? [novoBloco()] : [];
            renderTipos();
            renderBlocos();
        });
        refs.selectTipo.addEventListener("change", () => {
            state.emailTipo = refs.selectTipo.value;
            state.blocosEmail = [novoBloco()];
            renderBlocos();
        });
        esquerdo.appendChild(el("div", { class: "op-selects" }, [refs.selectProduto, refs.selectTipo]));
    }

    esquerdo.appendChild(el("div", { class: "ops-h" }, [el("span", { text: "Operações" })]));
    refs.listaOps = el("div", { class: "ops" });
    esquerdo.appendChild(refs.listaOps);

    const rodapeEsquerdo = el("div", { class: "mail-foot" });
    refs.btnAdicionar = btn("Adicionar operação", { icon: icons.iconMais, onClick: adicionarBloco });
    rodapeEsquerdo.appendChild(refs.btnAdicionar);
    rodapeEsquerdo.appendChild(btn("Gerar e-mail", { classe: "pri", onClick: gerarEmail }));
    esquerdo.appendChild(rodapeEsquerdo);

    const linkEstruturadas = el("button", {
        class: "link-estr",
        text: "Preciso registrar uma Operação Estruturada",
        onClick: mostrarInfoEstruturadas,
    });
    esquerdo.appendChild(linkEstruturadas);

    // ---- painel direito ----
    const direito = el("div", { class: "mail-r" });
    direito.appendChild(
        el("div", { class: "mail-r-h" }, [
            el("h3", { text: "E-mail gerado" }),
            el("div", { class: "mail-r-actions" }, [
                btn("Copiar texto", { classe: "pri", icon: icons.iconCopiar, onClick: copiarTexto }),
                // O rascunho no Outlook usa automação COM, que só existe no
                // Windows — no macOS o botão nem aparece, em vez de aparecer
                // e falhar no clique. Lá o fluxo é "Copiar texto" e colar no
                // cliente de e-mail.
                state.plataforma === "windows"
                    ? btn("Abrir no Outlook", { classe: "pri", icon: icons.iconEmail, onClick: abrirNoOutlook })
                    : null,
            ]),
        ])
    );
    refs.textoGerado = el("textarea", { class: "email-out" });
    direito.appendChild(refs.textoGerado);

    container.appendChild(esquerdo);
    container.appendChild(direito);

    if (modo === "padronizado") {
        // Aba de e-mails começa com o primeiro produto/tipo do catálogo e
        // uma operação em branco pronta pra preencher — só na primeira
        // montagem (ou depois de trocar de modo, que zera o estado).
        if (!state.emailProduto && produtosDisponiveis(modo).length > 0) {
            state.emailProduto = produtosDisponiveis(modo)[0];
            state.emailTipo = categoriasDoProduto(state.emailProduto, modo)[0]?.Label || null;
        }
        if (state.blocosEmail.length === 0 && state.emailTipo) {
            state.blocosEmail.push(novoBloco());
        }
        renderProdutos();
        renderTipos();
    } else if (state.blocosEmail.length === 0 && produtosDisponiveis(modo).length > 0) {
        state.blocosEmail.push(novoBloco(produtosDisponiveis(modo)[0], modo));
    }

    renderBlocos();
}

function produtosDisponiveis(modo) {
    const produtos = (state.catalogoEmail && state.catalogoEmail.Produtos) || [];
    if (modo !== "livre") return produtos;
    const cats = state.catalogoEmail?.Categorias || [];
    // No modo livre, produtos "só padronizado" (ex.: Resgate Prev) não
    // aparecem — eles exigem produto único por e-mail.
    return produtos.filter((p) => cats.some((c) => c.Group === p && !c.SoPadronizado));
}

function categoriasDoProduto(produto, modo) {
    const cats = (state.catalogoEmail?.Categorias || []).filter((c) => c.Group === produto);
    return modo === "livre" ? cats.filter((c) => !c.SoPadronizado) : cats;
}

function categoriaPorGrupoLabel(produto, label) {
    return (state.catalogoEmail?.Categorias || []).find((c) => c.Group === produto && c.Label === label);
}

function categoriaAtual() {
    return categoriaPorGrupoLabel(state.emailProduto, state.emailTipo);
}

function novoBloco(produtoInicial) {
    if (modoAtual() === "livre") {
        const produto = produtoInicial || produtosDisponiveis("livre")[0];
        const categorias = categoriasDoProduto(produto, "livre");
        return { id: novoBlocoId(), group: produto, label: categorias[0]?.Label || "", valoresBrutos: {} };
    }
    return { id: novoBlocoId(), valoresBrutos: {} };
}

function renderProdutos() {
    clear(refs.selectProduto);
    for (const p of produtosDisponiveis("padronizado")) {
        const opt = el("option", { value: p, text: p });
        if (p === state.emailProduto) opt.selected = true;
        refs.selectProduto.appendChild(opt);
    }
}

function renderTipos() {
    clear(refs.selectTipo);
    for (const c of categoriasDoProduto(state.emailProduto, "padronizado")) {
        const opt = el("option", { value: c.Label, text: c.Label });
        if (c.Label === state.emailTipo) opt.selected = true;
        refs.selectTipo.appendChild(opt);
    }
}

function adicionarBloco() {
    const modo = modoAtual();
    if (modo === "padronizado") {
        const cat = categoriaAtual();
        if (!cat) return;
        if (cat.OperacaoUnica && state.blocosEmail.length >= 1) return; // ex.: Resgate Prev — 1 operação por e-mail
        state.blocosEmail.push(novoBloco());
    } else {
        if (!produtosDisponiveis("livre").length) return;
        state.blocosEmail.push(novoBloco());
    }
    renderBlocos();
}

function renderBlocos() {
    if (!refs.listaOps) return;
    // clear() esvazia o container e reseta o scroll pro topo — preserva a
    // posição pra não "puxar pra cima" a cada troca de produto/tipo num
    // e-mail com várias operações.
    const scrollAnterior = refs.listaOps.scrollTop;
    clear(refs.listaOps);
    const modo = modoAtual();
    const catPadronizado = modo === "padronizado" ? categoriaAtual() : null;

    if (modo === "padronizado") {
        refs.btnAdicionar.disabled = !catPadronizado || (catPadronizado.OperacaoUnica && state.blocosEmail.length >= 1);
    } else {
        refs.btnAdicionar.disabled = !produtosDisponiveis("livre").length;
    }

    state.blocosEmail.forEach((bloco, idx) => {
        refs.listaOps.appendChild(renderBloco(bloco, idx, modo, catPadronizado));
    });
    refs.listaOps.scrollTop = scrollAnterior;
}

function renderBloco(bloco, idx, modo, catPadronizado) {
    const wrapper = el("div", { class: "op" });

    if (modo === "livre") {
        wrapper.appendChild(
            el("div", { class: "op-h" }, [el("span", { class: "op-num", text: String(idx + 1) }), bloco.group || ""])
        );

        const selectProduto = el("select", { class: "select" });
        for (const p of produtosDisponiveis("livre")) {
            const opt = el("option", { value: p, text: p });
            if (p === bloco.group) opt.selected = true;
            selectProduto.appendChild(opt);
        }
        selectProduto.addEventListener("change", () => {
            bloco.group = selectProduto.value;
            const categorias = categoriasDoProduto(bloco.group, "livre");
            bloco.label = categorias[0]?.Label || "";
            bloco.valoresBrutos = {};
            renderBlocos();
        });

        const selectOperacao = el("select", { class: "select" });
        for (const c of categoriasDoProduto(bloco.group, "livre")) {
            const opt = el("option", { value: c.Label, text: c.Label });
            if (c.Label === bloco.label) opt.selected = true;
            selectOperacao.appendChild(opt);
        }
        selectOperacao.addEventListener("change", () => {
            bloco.label = selectOperacao.value;
            bloco.valoresBrutos = {};
            renderBlocos();
        });

        wrapper.appendChild(el("div", { class: "op-selects" }, [selectProduto, selectOperacao]));
    } else {
        wrapper.appendChild(
            el("div", { class: "op-h" }, [el("span", { class: "op-num", text: String(idx + 1) }), catPadronizado?.Label || ""])
        );
    }

    const cat = modo === "livre" ? categoriaPorGrupoLabel(bloco.group, bloco.label) : catPadronizado;
    if (cat) {
        for (const campo of cat.Fields) {
            wrapper.appendChild(el("div", { class: "field-lbl", text: campo.Label }));
            if (campo.Type === "select") {
                const opcoes = campo.Options || [];
                const sel = el("select", { class: "select" });
                for (const opt of opcoes) sel.appendChild(el("option", { value: opt, text: opt }));
                sel.value = bloco.valoresBrutos[campo.Key] || opcoes[0] || "";
                bloco.valoresBrutos[campo.Key] = sel.value;
                sel.addEventListener("change", () => {
                    bloco.valoresBrutos[campo.Key] = sel.value;
                });
                wrapper.appendChild(sel);
            } else {
                const input = el("input", { class: "input", type: "text", placeholder: campo.Placeholder || "" });
                input.value = bloco.valoresBrutos[campo.Key] || "";
                input.addEventListener("input", () => {
                    bloco.valoresBrutos[campo.Key] = input.value;
                });
                wrapper.appendChild(input);
            }
        }
        if (cat.Anexo) {
            const aviso = el("div", { class: "warn" });
            aviso.insertAdjacentHTML("beforeend", icons.iconAlerta);
            aviso.appendChild(el("span", { text: cat.Anexo }));
            wrapper.appendChild(aviso);
        }
    }

    if (modo === "livre" || state.blocosEmail.length > 1) {
        const btnRemover = btn("Remover", { classe: "danger", onClick: () => {
            state.blocosEmail = state.blocosEmail.filter((b) => b.id !== bloco.id);
            renderBlocos();
        } });
        wrapper.appendChild(el("div", { class: "op-foot" }, [btnRemover]));
    }

    return wrapper;
}

function lookupCliente() {
    const codigo = refs.codigo.value.trim();
    if (!codigo) {
        refs.statusLookup.textContent = "";
        refs.statusLookup.className = "lookup";
        refs.statusLookupEmail.textContent = "";
        refs.statusLookupEmail.className = "lookup";
        return;
    }
    const nome = state.clientDB[codigo];
    if (nome) {
        refs.nome.value = nome;
        refs.statusLookup.className = "lookup ok";
        refs.statusLookup.innerHTML = icons.iconCheck;
        refs.statusLookup.appendChild(document.createTextNode("Cliente encontrado na base."));
    } else {
        refs.statusLookup.className = "lookup";
        refs.statusLookup.textContent = "Cliente não encontrado — informe o nome manualmente.";
    }

    const email = state.clientEmails[codigo];
    if (email) {
        refs.email.value = email;
        refs.statusLookupEmail.className = "lookup ok";
        refs.statusLookupEmail.innerHTML = icons.iconCheck;
        refs.statusLookupEmail.appendChild(document.createTextNode("E-mail encontrado na base."));
    } else {
        refs.statusLookupEmail.className = "lookup";
        refs.statusLookupEmail.textContent = "E-mail não encontrado — informe manualmente.";
    }
}

// Chamado de fora (main.js) depois de carregar uma nova base de clientes,
// pra reagir a um código já digitado no campo.
export function refreshLookup() {
    if (refs.codigo) lookupCliente();
}

async function gerarEmail() {
    if (!state.blocosEmail.length) {
        alert("Adicione ao menos uma operação.");
        return;
    }
    const modo = modoAtual();
    const itens =
        modo === "padronizado"
            ? state.blocosEmail.map((b) => ({ Group: state.emailProduto, Label: state.emailTipo, Valores: b.valoresBrutos }))
            : state.blocosEmail.map((b) => ({ Group: b.group, Label: b.label, Valores: b.valoresBrutos }));
    try {
        const texto = await GerarEmail(refs.codigo.value.trim(), refs.nome.value.trim(), itens, modo);
        refs.textoGerado.value = texto;
    } catch (e) {
        alert("Não foi possível gerar o e-mail.\n\n" + e);
    }
}

async function copiarTexto() {
    const texto = refs.textoGerado.value;
    if (!texto) return;
    try {
        await CopiarTexto(texto);
    } catch (e) {
        alert("Não foi possível copiar.\n\n" + e);
    }
}

async function abrirNoOutlook() {
    const texto = refs.textoGerado.value;
    if (!texto) return;
    const destinatario = refs.email.value.trim();
    if (!destinatario) {
        alert("Informe o e-mail do destinatário.");
        return;
    }
    const nome = refs.nome.value.trim() || "[Nome do cliente]";
    const codigo = refs.codigo.value.trim() || "[Código do cliente]";
    const produto = modoAtual() === "padronizado" ? state.emailProduto : "Ordem";
    const assunto = `${produto} - ${nome} (${codigo})`;
    try {
        await AbrirEmailNoOutlook(destinatario, assunto, texto);
    } catch (e) {
        alert("Não foi possível abrir o e-mail no Outlook.\n\n" + e);
    }
}

function mostrarInfoEstruturadas() {
    const caixa = el("div", { class: "modal-caixa" });
    caixa.appendChild(el("div", { text: state.catalogoEmail?.InfoEstruturadas || "" }));
    const fundo = el("div", { class: "modal-fundo" });
    caixa.appendChild(el("div", { class: "rodape-modal" }, [btn("Fechar", { classe: "pri", onClick: () => fundo.remove() })]));
    fundo.appendChild(caixa);
    fundo.addEventListener("click", (e) => {
        if (e.target === fundo) fundo.remove();
    });
    // Anexado dentro de #app (não em document.body): as variáveis de tema
    // só chegam a descendentes de .app — fora disso o popup fica com fundo
    // transparente.
    document.getElementById("app").appendChild(fundo);
}
