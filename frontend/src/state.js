// Estado em memória compartilhado entre as abas do app. Um único objeto
// singleton, sem framework reativo — os módulos de cada aba chamam as
// funções `render*` depois de mudar o estado.

export const state = {
    pasta: null, // string | null
    modelo: "",
    modeloFestas: "", // modelo_festas.txt da pasta ativa — usado quando prefs.modoFestas está ligado
    clientes: [], // ClienteRentabilidadeDTO[] — todo cliente da base, com Registro anexado (ou null)
    clientDB: {}, // {codigo: nome}
    clientEmails: {}, // {codigo: email}
    clienteSelecionado: null, // Codigo do cliente selecionado na lista (mesmo quando ele está em branco)
    arquivoSelecionado: null, // nome do arquivo do Registro do cliente selecionado, ou null se em branco
    filtro: "",
    falhasProcessamento: [], // FalhaDTO[] — do último ProcessarPastaAtual, visível em Configurações
    catalogoEmail: null, // CatalogoEmailDTO, carregado uma vez
    // Produto + tipo de movimentação valem pro e-mail inteiro (regra de
    // compliance: um e-mail só pode ter operações do mesmo produto/tipo).
    emailProduto: null, // string | null — CategoriaDTO.Group
    emailTipo: null, // string | null — CategoriaDTO.Label
    blocosEmail: [], // {id, valores: {}} — todas do mesmo produto/tipo acima
    blocosDesagio: [], // {id, titulo, valorAtual, valorSaida}
    calculadora: {}, // {"cartaoId.campo": "texto digitado"} — persiste ao trocar de aba
    previdenciaria: {}, // {"prev.campo": "texto digitado"} — persiste ao trocar de aba
    compromissada: {}, // {"comp.campo": "texto digitado"} — persiste ao trocar de aba
    aposentadoria: {}, // {"apos.campo": "texto digitado"} — persiste ao trocar de aba
    comparadorRF: {}, // {"cmp.a/b.campo": "texto digitado"} — persiste ao trocar de aba (aba Comparadora)
    calcRF: {}, // {"calcrf.campo": "texto digitado"} — persiste ao trocar de aba (Calculadora de Renda Fixa)
    typeform: {}, // {"typ.<numero>": resposta} — persiste ao trocar de aba (aba Typeform)
    versaoApp: "", // preenchido no boot via VersaoAtual() — "dev" fora de um build de release (ver main.go)
    plataforma: "windows", // "windows" | "darwin" — decide o que existe na interface (ver ABAS_SO_WINDOWS em main.js)
    atualizacao: { disponivel: false, versao: "", notas: "" }, // ver EventsOn("atualizacao:disponivel") em main.js
    prefs: {
        tema: "claro",
        acento: "esmeralda",
        modoEmail: "padronizado",
        tabelaPrevidenciaria: "2026",
        visao: "cliente",
        fonte: "'Plus Jakarta Sans', system-ui, sans-serif",
        modoApresentacao: false,
        // Modo Festas: a aba Rentabilidade passa a mandar a mensagem de
        // modelo_festas.txt (só placeholder _Nome) pra todo cliente da
        // base, com ou sem relatório processado.
        modoFestas: false,
        // E-mail da conta Outlook usada como remetente pelos rascunhos
        // abertos via AbrirEmailNoOutlook — vazio deixa o Outlook decidir
        // sozinho (ambíguo com mais de uma conta logada).
        emailRemetente: "",
        // Ordem das ferramentas na barra lateral (IDs de aba). Vazio = ordem
        // padrão (ver ORDEM_NAV_PADRAO em main.js).
        ordemNav: [],
        // IDs de ferramentas escondidas da barra lateral (botão "-" em
        // Configurações → Ordem da barra lateral). Vazio = nenhuma escondida.
        ordemNavOcultos: [],
        // Recorte da imagem de "Copiar imagem" (aba Rentabilidade) — salvo à
        // parte via SalvarRecorteGraficoRentabilidade, não por salvarPrefs().
        recortePersonalizado: false,
        recorteX0: 0,
        recorteY0: 0,
        recorteX1: 0,
        recorteY1: 0,
        recortePadraoX0: 0,
        recortePadraoY0: 0,
        recortePadraoX1: 0,
        recortePadraoY1: 0,
    }, // aplicado no #app, persistido via SalvarPreferencias
};

let proximoBlocoId = 1;
export function novoBlocoId() {
    return proximoBlocoId++;
}
