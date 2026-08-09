// Réplica mínima, em JS, das funções de parse/formatação pt-BR que existem
// em Go (internal/pdfreport/pdfreport.go: ParseNumeroPtBR/FormatarReais/
// FormatarPercentual). Necessária aqui porque a tabela de deságio recalcula
// a cada tecla digitada — um round-trip pro backend a cada keystroke seria
// perceptível (o app já levou uma queixa de lentidão de interface antes).
// Se a regra de formatação mudar num lado, replique no outro.

// parseNumeroPtBR aceita "R$ 1.234,56", "-1,08%" ou "2.434,91" -> number, ou
// null se não for um número válido.
export function parseNumeroPtBR(texto) {
    texto = (texto || "").trim();
    if (!texto) return null;
    const negativo = texto.startsWith("-");
    let numero = texto.replace(/^-/, "").replace(/R\$/g, "").replace(/%/g, "").trim();
    numero = numero.replace(/\./g, "").replace(",", ".");
    if (numero === "") return null;
    const valor = parseFloat(numero);
    if (Number.isNaN(valor)) return null;
    return negativo ? -valor : valor;
}

function magnitudePtBR(valor) {
    return Math.abs(valor).toLocaleString("pt-BR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

export function formatarReais(valor) {
    if (valor === null || valor === undefined || Number.isNaN(valor)) return "";
    return `${valor < 0 ? "-" : ""}R$ ${magnitudePtBR(valor)}`;
}

export function formatarPercentual(valor) {
    if (valor === null || valor === undefined || Number.isNaN(valor)) return "";
    return `${valor < 0 ? "-" : ""}${magnitudePtBR(valor)}%`;
}

export function formatarNumero(valor, casas = 2) {
    if (valor === null || valor === undefined || Number.isNaN(valor)) return "";
    return valor.toLocaleString("pt-BR", { minimumFractionDigits: casas, maximumFractionDigits: casas });
}

// Reformata a parte inteira de um valor em edição com "." a cada 3 dígitos
// (1234 -> 1.234), sem mexer na parte decimal digitada — usado no evento
// "input" dos campos de reais pra já mostrar o separador de milhar enquanto
// o usuário digita, em vez de só no blur.
export function formatarMilharEnquantoDigita(texto) {
    const idxVirgula = texto.indexOf(",");
    let inteiro = (idxVirgula === -1 ? texto : texto.slice(0, idxVirgula)).replace(/\D/g, "");
    if (inteiro === "") return idxVirgula === -1 ? "" : ",";
    inteiro = inteiro.replace(/\B(?=(\d{3})+(?!\d))/g, ".");
    if (idxVirgula === -1) return inteiro;
    const decimais = texto.slice(idxVirgula + 1).replace(/\D/g, "").slice(0, 2);
    return `${inteiro},${decimais}`;
}

// Acrescenta "%" no fim enquanto o usuário digita um campo percentual (sem
// mexer no resto do texto) — mesma ideia de formatarMilharEnquantoDigita,
// só que pro sinal de porcentagem em vez do separador de milhar.
export function formatarPercentualEnquantoDigita(texto) {
    const semSinal = texto.replace(/%/g, "");
    return semSinal === "" ? "" : `${semSinal}%`;
}
