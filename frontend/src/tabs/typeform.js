import { SalvarTypeform, PreencherTypeform } from "../../wailsjs/go/main/App.js";
import { EventsOn } from "../../wailsjs/runtime/runtime.js";
import { state } from "../state.js";
import { el, clear, btn, montarToolbarLimpar } from "../ui/components.js";
import * as icons from "../ui/icons.js";
import { parseNumeroPtBR, formatarReais, formatarMilharEnquantoDigita } from "../util/numeros.js";

// ctx da montagem mais recente — guardado à parte porque o listener de
// progresso abaixo é registrado uma única vez no carregamento do módulo
// (não a cada mount(), que também é chamado de novo pelo botão "Limpar" —
// registrar ali duplicaria o listener a cada clique).
let ctxApp = null;
EventsOn("typeform:progresso", (feitos, total, pergunta) => {
    ctxApp?.setStatus(`Preenchendo Typeform (${feitos}/${total}): ${pergunta}`);
});

// Perguntas do Typeform de coleta de dados do cliente, na mesma ordem e
// numeração do formulário original (perguntas 2 a 46 — a numeração começa
// em 2 porque a 1 é o cabeçalho institucional do Typeform, fora do escopo
// desta aba). Cada seção vira um cartão; cada pergunta guarda a resposta em
// state.typeform["typ.<chave>"].
const SECOES = [
    {
        titulo: "Dados Pessoais do Cliente",
        perguntas: [
            { chave: "3", texto: "Nome completo do cliente:", tipo: "texto" },
            { chave: "4", texto: "Data de nascimento:", tipo: "data" },
            {
                chave: "5",
                texto: "Estado civil:",
                tipo: "radio",
                // Opções iguais às do formulário real (múltipla escolha,
                // não texto livre) — o preenchimento automático casa a
                // resposta salva com essas opções por texto, então precisam
                // bater com a redação exata usada lá.
                opcoes: ["Solteiro(a)", "Casado(a) / União estável", "Divorciado(a)", "Viúvo(a)"],
            },
            {
                chave: "6",
                texto: "Regime de bens adotado no matrimônio:",
                tipo: "radio",
                opcoes: [
                    "Comunhão Parcial de Bens",
                    "Comunhão Universal de Bens",
                    "Separação Total de Bens",
                    "Participação Final nos Aquestos",
                ],
            },
        ],
    },
    {
        titulo: "Estrutura Familiar",
        perguntas: [
            {
                chave: "7",
                texto: "Possui dependentes financeiros? (Filhos, cônjuge sem renda, pais. Pode marcar mais de uma opção.)",
                tipo: "checkbox",
                // Opções iguais às do formulário real — ver comentário na
                // pergunta 5 (Estado civil) sobre por que isso importa.
                opcoes: ["Não", "Filhos menores", "Cônjuge sem renda própria", "Pais / outros"],
            },
            {
                chave: "8",
                texto: "Quantos anos faltam até o dependente mais novo se tornar financeiramente independente?",
                tipo: "radio",
                opcoes: ["Menos de 5 anos", "Entre 5 e 10 anos", "Nunca (dependente permanente)", "Não se aplica"],
            },
            { chave: "9", texto: "É o principal provedor financeiro da família?", tipo: "radio", opcoes: ["Sim", "Não"] },
        ],
    },
    {
        titulo: "Renda e Situação Profissional",
        perguntas: [
            {
                chave: "10",
                texto: "Profissão/Ocupação principal:",
                tipo: "radio",
                opcoes: [
                    "Empresário / Sócio",
                    "Profissional liberal (PJ)",
                    "Assalariado CLT",
                    "Servidor público",
                    "Investidor / Rentista",
                    "Aposentado(a)",
                ],
            },
            {
                chave: "11",
                texto: "Forma de remuneração principal:",
                tipo: "radio",
                opcoes: ["Salário CLT", "Pró-labore", "Distribuição de lucros", "Honorários / RPA", "Dividendos / aluguéis"],
            },
            { chave: "12", texto: "Qual a sua renda mensal líquida (individual)?", tipo: "valor" },
            { chave: "13", texto: "Possui renda extra ou esporádica?", tipo: "radio", opcoes: ["Sim", "Não"] },
            // Pergunta condicional do formulário real (só aparece se a
            // anterior for "Sim") — não existia na aba, causava parada no
            // preenchimento automático.
            { chave: "13a", texto: "Qual o valor da sua renda extra/esporádica?", tipo: "valor" },
            { chave: "14", texto: "Despesa mensal total estimada:", tipo: "valor" },
            {
                chave: "15",
                texto: "Declaração de Imposto de Renda (Pessoa Física):",
                tipo: "radio",
                opcoes: ["Simplificada", "Completa", "Isento / Não declaro"],
            },
            {
                chave: "16",
                texto: "Regime tributário da empresa (se aplicável):",
                tipo: "radio",
                opcoes: ["Simples Nacional", "Lucro Presumido", "Lucro Real", "Não se aplica"],
            },
        ],
    },
    {
        titulo: "Patrimônio e Investimentos",
        perguntas: [
            { chave: "17", texto: "Valor total em aplicações financeiras líquidas (exceto previdência e imóveis):", tipo: "valor" },
            { chave: "18", texto: "Possui previdência privada ativa?", tipo: "radio", opcoes: ["Sim", "Não"] },
            // Condicional (só aparece se a anterior for "Sim") — mesma
            // situação da 13a/20a, acima.
            { chave: "18a", texto: "Qual o valor da sua previdência privada ativa?", tipo: "valor" },
            { chave: "19", texto: "Patrimônio em imóveis (valor de mercado estimado):", tipo: "valor" },
            {
                chave: "20",
                texto: "Possui investimentos internacionais?",
                tipo: "radio",
                opcoes: ["Não", "Sim, via conta PF no exterior", "Sim, via Offshore/BVI"],
            },
            // Pergunta condicional do formulário real (só aparece se a
            // anterior não for "Não") — mesma situação da 13a, acima.
            { chave: "20a", texto: "Qual o valor total dos investimentos internacionais?", tipo: "valor" },
            { chave: "21", texto: "Você vê valor em possuir investimentos dolarizados?", tipo: "radio", opcoes: ["Sim", "Não"] },
            { chave: "22", texto: "Possui investimentos via Pessoa Jurídica?", tipo: "radio", opcoes: ["Sim", "Não"] },
            { chave: "23", texto: "Possui dívidas ou financiamentos ativos? (Individual)", tipo: "radio", opcoes: ["Sim", "Não"] },
            {
                chave: "24",
                texto:
                    "Por quantos meses conseguiria manter seu padrão de vida utilizando apenas suas reservas, sem depender da renda?",
                tipo: "radio",
                opcoes: ["Menos de 6 meses", "Entre 6 meses e 2 anos", "Entre 2 e 5 anos", "Mais de 5 anos", "Indefinidamente"],
                nota: "Indicador calculado dividindo o patrimônio financeiro líquido pelas despesas mensais.",
            },
        ],
    },
    {
        titulo: "Proteção Patrimonial",
        perguntas: [
            {
                chave: "25",
                texto: "Possui seguro de vida contratado?",
                tipo: "radio",
                opcoes: [
                    "Não possuo",
                    "Sim, possuo proteção completa para minha renda e em caso de falecimento",
                    "Sim, possuo cobertura apenas em caso de falecimento",
                    "Sim, possuo, mas preciso revisar minha apólice",
                ],
            },
        ],
    },
    {
        titulo: "Objetivos Patrimoniais — Aquisição de Imóvel",
        perguntas: [
            { chave: "26", texto: "Pretende fazer alguma aquisição imobiliária nos próximos 1 a 7 anos?", tipo: "radio", opcoes: ["Sim", "Não"] },
            { chave: "27", texto: "Qual o valor estimado do imóvel?", tipo: "valor" },
            { chave: "28", texto: "Qual o prazo para aquisição?", tipo: "radio", opcoes: ["3 a 12 meses", "1 a 3 anos", "3 a 7 anos"] },
        ],
    },
    {
        titulo: "Objetivos Patrimoniais — Aquisição de Automóvel",
        perguntas: [
            { chave: "29", texto: "Pretende comprar ou trocar de automóvel nos próximos 2 a 5 anos?", tipo: "radio", opcoes: ["Sim", "Não"] },
            { chave: "30", texto: "Valor estimado:", tipo: "valor" },
            { chave: "31", texto: "Prazo para aquisição:", tipo: "radio", opcoes: ["3 a 12 meses", "1 a 3 anos", "3 a 7 anos"] },
        ],
    },
    {
        titulo: "Planejamento Patrimonial e Sucessório",
        perguntas: [
            { chave: "32", texto: "Possui holding patrimonial constituída?", tipo: "radio", opcoes: ["Sim", "Não"] },
            { chave: "33", texto: "Possui testamento formalizado?", tipo: "radio", opcoes: ["Sim", "Não"] },
            {
                chave: "34",
                texto: "O tema sucessão patrimonial e planejamento de herança é relevante para você hoje?",
                tipo: "radio",
                opcoes: [
                    "Quero entender melhor o tema",
                    "É relevante, mas ainda não tomei ação",
                    "Não é prioridade agora",
                    "Já estou estruturando",
                ],
            },
        ],
    },
    {
        titulo: "Perfil do Investidor",
        perguntas: [
            {
                chave: "35",
                texto: "Como você se define como investidor?",
                tipo: "radio",
                opcoes: [
                    "Conservador – priorizo segurança",
                    "Moderado – equilíbrio entre risco e retorno",
                    "Agressivo – foco na multiplicação do patrimônio",
                ],
            },
            {
                chave: "36",
                texto: "Qual seu nível de conhecimento sobre mercado financeiro?",
                tipo: "radio",
                opcoes: ["Iniciante", "Intermediário", "Experiente", "Profissional"],
            },
            { chave: "37", texto: "Qual o valor necessário para sua reserva de emergência/liquidez imediata?", tipo: "valor" },
            {
                chave: "38",
                texto: "Qual benchmark de retorno você considera ideal para um período de 1 ano?",
                tipo: "radio",
                opcoes: ["CDI", "CDI + 2% a.a.", "IPCA + 4% a.a.", "Entre 12% e 15% a.a.", "Acima de 15% a.a."],
            },
        ],
    },
    {
        titulo: "Objetivos Financeiros",
        perguntas: [
            {
                chave: "39",
                texto: "Possui algum objetivo financeiro de curto ou médio prazo? (Viagens, cirurgias, educação, etc.)",
                tipo: "radio",
                opcoes: ["Sim", "Não"],
            },
            {
                chave: "40",
                texto: "Qual é esse objetivo?",
                tipo: "radio",
                opcoes: [
                    "Compra de imóvel",
                    "Educação dos filhos",
                    "Viagem/Experiência",
                    "Abertura ou expansão de negócio",
                    "Acumulação de patrimônio",
                    "Outro",
                ],
                comOutro: true,
            },
            { chave: "41", texto: "Valor necessário para atingir esse objetivo:", tipo: "valor" },
            { chave: "42", texto: "Em quantos anos deseja atingir esse objetivo?", tipo: "anos" },
        ],
    },
    {
        titulo: "Aposentadoria",
        perguntas: [
            {
                chave: "43",
                texto: "Já pensou em sua aposentadoria ou independência financeira?",
                tipo: "radio",
                opcoes: ["Ainda não pensei", "Já penso, mas não planejei", "Tenho um plano informal", "Tenho um plano estruturado"],
            },
            { chave: "44", texto: "Com que idade deseja se aposentar ou atingir independência financeira?", tipo: "anos" },
            { chave: "45", texto: "Qual renda mensal deseja ter na aposentadoria?", tipo: "valor" },
        ],
    },
];

// Pergunta usada como nome do cliente pro nome do arquivo salvo.
const CHAVE_NOME = "typ.3";

function chaveEstado(pergunta) {
    return `typ.${pergunta.chave}`;
}

function formatarSeReais(texto) {
    if (!texto) return "";
    const valor = parseNumeroPtBR(texto);
    return valor !== null ? formatarReais(valor) : texto;
}

function linhaBase(pergunta) {
    const linha = el("div", { class: "tf-pergunta" });
    linha.appendChild(el("div", { class: "tf-rotulo", text: `${pergunta.chave}. ${pergunta.texto}` }));
    if (pergunta.nota) linha.appendChild(el("div", { class: "tf-nota", text: pergunta.nota }));
    return linha;
}

function montarCampoSimples(pergunta) {
    const linha = linhaBase(pergunta);
    const k = chaveEstado(pergunta);
    const input = el("input", {
        class: "tf-input",
        type: pergunta.tipo === "data" ? "date" : "text",
        placeholder: pergunta.tipo === "valor" ? "R$ 0,00" : "",
    });
    const valorSalvo = state.typeform[k] || "";
    input.value = pergunta.tipo === "valor" ? formatarSeReais(valorSalvo) : valorSalvo;

    input.addEventListener("input", () => {
        if (pergunta.tipo === "valor") {
            const distanciaDoFim = input.value.length - input.selectionStart;
            input.value = formatarMilharEnquantoDigita(input.value);
            const pos = Math.max(0, input.value.length - distanciaDoFim);
            input.setSelectionRange(pos, pos);
        }
        state.typeform[k] = input.value;
    });
    if (pergunta.tipo === "valor") {
        input.addEventListener("focus", () => {
            input.value = input.value.replace(/^R\$\s*/, "");
        });
        input.addEventListener("blur", () => {
            const valor = parseNumeroPtBR(input.value);
            if (valor !== null) {
                input.value = formatarReais(valor);
                state.typeform[k] = input.value;
            }
        });
    }

    const wrap = el("div", { class: "tf-campo" }, [input]);
    if (pergunta.tipo === "anos") wrap.appendChild(el("span", { class: "csuffix", text: "anos" }));
    linha.appendChild(wrap);
    return linha;
}

// Radio (escolha única) — clique de novo na opção já marcada desmarca. Se
// `comOutro`, a opção "Outro" revela um campo de texto extra pra
// especificar (usado só na pergunta 40).
function montarCampoRadio(pergunta) {
    const linha = linhaBase(pergunta);
    const k = chaveEstado(pergunta);
    const opts = el("div", { class: "tf-opts" });

    const outroInput = el("input", { class: "tf-input", type: "text", placeholder: "Especifique..." });
    outroInput.value = state.typeform[`${k}.outro`] || "";
    outroInput.addEventListener("input", () => {
        state.typeform[`${k}.outro`] = outroInput.value;
    });
    const outroWrap = el("div", { class: "tf-campo tf-outro" }, [outroInput]);

    function render() {
        clear(opts);
        const selecionado = state.typeform[k];
        for (const opcao of pergunta.opcoes) {
            const botao = el("button", { class: `tf-opt${selecionado === opcao ? " on" : ""}`, type: "button", text: opcao });
            botao.addEventListener("click", () => {
                state.typeform[k] = selecionado === opcao ? null : opcao;
                render();
            });
            opts.appendChild(botao);
        }
        outroWrap.hidden = !(pergunta.comOutro && state.typeform[k] === "Outro");
    }
    render();

    linha.appendChild(opts);
    if (pergunta.comOutro) linha.appendChild(outroWrap);
    return linha;
}

// Checkbox (múltipla escolha) — cada clique liga/desliga a opção
// independente das outras.
function montarCampoCheckbox(pergunta) {
    const linha = linhaBase(pergunta);
    const k = chaveEstado(pergunta);
    const opts = el("div", { class: "tf-opts" });

    function render() {
        clear(opts);
        const selecionados = state.typeform[k] || [];
        for (const opcao of pergunta.opcoes) {
            const ligado = selecionados.includes(opcao);
            const botao = el("button", { class: `tf-opt${ligado ? " on" : ""}`, type: "button", text: opcao });
            botao.addEventListener("click", () => {
                const atual = state.typeform[k] || [];
                state.typeform[k] = atual.includes(opcao) ? atual.filter((o) => o !== opcao) : [...atual, opcao];
                render();
            });
            opts.appendChild(botao);
        }
    }
    render();

    linha.appendChild(opts);
    return linha;
}

function montarPergunta(pergunta) {
    if (pergunta.tipo === "radio") return montarCampoRadio(pergunta);
    if (pergunta.tipo === "checkbox") return montarCampoCheckbox(pergunta);
    return montarCampoSimples(pergunta);
}

function respostaFormatada(pergunta) {
    const k = chaveEstado(pergunta);
    if (pergunta.tipo === "checkbox") {
        const selecionados = state.typeform[k] || [];
        return selecionados.length ? selecionados.join(", ") : "—";
    }
    const valor = state.typeform[k];
    if (!valor) return "—";
    if (pergunta.tipo === "data") {
        const [ano, mes, dia] = String(valor).split("-");
        return dia && mes && ano ? `${dia}/${mes}/${ano}` : String(valor);
    }
    if (pergunta.comOutro && valor === "Outro") {
        const outro = (state.typeform[`${k}.outro`] || "").trim();
        return outro ? `Outro (${outro})` : "Outro";
    }
    return String(valor).trim() || "—";
}

function montarTexto() {
    const linhas = ["TYPEFORM DO CLIENTE", `Gerado em ${new Date().toLocaleString("pt-BR")}`, ""];
    for (const secao of SECOES) {
        linhas.push(secao.titulo.toUpperCase());
        for (const pergunta of secao.perguntas) {
            linhas.push("");
            linhas.push(`${pergunta.chave}. ${pergunta.texto}`);
            linhas.push(respostaFormatada(pergunta));
        }
        linhas.push("");
    }
    return linhas.join("\n");
}

async function salvar(ctx) {
    const nome = (state.typeform[CHAVE_NOME] || "").trim();
    if (!nome) {
        ctx.setStatus("Preencha o nome do cliente (pergunta 3) antes de salvar.");
        return;
    }
    try {
        const caminho = await SalvarTypeform(nome, montarTexto());
        ctx.setStatus("Respostas salvas: " + caminho);
    } catch (e) {
        ctx.setStatus("Erro ao salvar respostas: " + e);
    }
}

// Monta a lista {Pergunta, Valor} enviada pro preenchimento automático —
// mesmo texto de pergunta e mesma formatação de resposta usados no .txt
// salvo (respostaFormatada), pra não duplicar essa lógica. Perguntas sem
// resposta ("—") ficam de fora: nada pra preencher nelas mesmo.
//
// As duas primeiras perguntas do Typeform ("Nome do assessor responsável" e
// "E-mail do assessor") não têm campo nenhum nessa aba — são o mesmo em
// toda reunião, então vêm de Configurações > E-mail (SalvarDadosAssessor),
// não de state.typeform.
function respostasParaBot() {
    const itens = [];
    if (state.prefs?.assessorNome) {
        itens.push({ Pergunta: "Nome do assessor responsável", Valor: state.prefs.assessorNome });
    }
    if (state.prefs?.assessorEmail) {
        itens.push({ Pergunta: "E-mail do assessor (para receber o relatório)", Valor: state.prefs.assessorEmail });
    }
    for (const secao of SECOES) {
        for (const pergunta of secao.perguntas) {
            const valor = respostaFormatada(pergunta);
            if (valor && valor !== "—") {
                itens.push({ Pergunta: pergunta.texto, Valor: valor });
            }
        }
    }
    return itens;
}

// Abre o Typeform real no Edge e preenche com o que já foi respondido
// aqui — ver internal/typeformbot no backend pra como o casamento e o
// preenchimento acontecem. Sempre para antes do fim (nunca envia
// sozinho); o navegador fica aberto pra o assessor terminar e enviar.
async function enviarParaTypeform(ctx) {
    const nome = (state.typeform[CHAVE_NOME] || "").trim();
    if (!nome) {
        ctx.setStatus("Preencha o nome do cliente (pergunta 3) antes de enviar pro Typeform.");
        return;
    }
    if (!state.prefs?.assessorNome || !state.prefs?.assessorEmail) {
        ctx.setStatus("Configure nome e e-mail do assessor em Configurações > E-mail antes de enviar pro Typeform.");
        return;
    }
    const itens = respostasParaBot();
    ctx.setStatus("Abrindo o Typeform no Edge...");
    try {
        await PreencherTypeform(itens);
        ctx.setStatus("Preenchimento concluído — confira e envie você mesmo no navegador.");
    } catch (e) {
        // A parada esperada (sem correspondência, ou última resposta
        // disponível) já vem como mensagem pronta pra mostrar — ver
        // ErroParado em internal/typeformbot. Um erro de verdade (Edge não
        // encontrado, formulário não abriu) também aparece aqui, só que
        // sem o prefixo "parei em".
        const texto = String(e);
        ctx.setStatus("Typeform: " + texto.split("\n")[0]);
        // Mensagens longas (ex.: a de página em branco, que explica como
        // testar a rede) não cabem na barra de status — essas vão num
        // alerta, senão o assessor não chega a ler justamente a parte que
        // diz o que fazer.
        if (texto.includes("\n")) alert(texto);
    }
}

export function mount(container, ctx) {
    clear(container);
    ctxApp = ctx;

    const toolbar = montarToolbarLimpar(state.typeform, () => mount(container, ctx));
    toolbar.appendChild(btn("Salvar respostas (.txt)", { classe: "pri", icon: icons.iconSalvar, onClick: () => salvar(ctx) }));
    toolbar.appendChild(btn("Enviar para o Typeform", { icon: icons.iconEnviar, onClick: () => enviarParaTypeform(ctx) }));
    container.appendChild(toolbar);

    for (const secao of SECOES) {
        const card = el("div", { class: "tf-secao" });
        card.appendChild(el("h4", { text: secao.titulo }));
        for (const pergunta of secao.perguntas) card.appendChild(montarPergunta(pergunta));
        container.appendChild(card);
    }
}
