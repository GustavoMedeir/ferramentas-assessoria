// Calendário de dias úteis Brasil/ANBIMA, calculado em vez de embutir uma
// tabela de datas (a fonte original — a planilha de Compromissada — traz uma
// lista pronta de 2001 a 2078, mas reproduzi-la aqui exigiria manter um
// arquivo de dados; o algoritmo abaixo bate com ela nos anos testados).
// Toda a aritmética é em timestamps UTC (Date.UTC/getUTCDay) — usar Date
// local aqui causaria off-by-one de fuso horário na leitura de
// <input type="date">, que sempre entrega "YYYY-MM-DD" em UTC.

const DIA_MS = 86400000;

// Domingo de Páscoa (algoritmo de Meeus/Jones/Butcher), retorna timestamp UTC.
function pascoa(ano) {
    const a = ano % 19;
    const b = Math.floor(ano / 100);
    const c = ano % 100;
    const d = Math.floor(b / 4);
    const e = b % 4;
    const f = Math.floor((b + 8) / 25);
    const g = Math.floor((b - f + 1) / 3);
    const h = (19 * a + b - d - g + 15) % 30;
    const i = Math.floor(c / 4);
    const k = c % 4;
    const l = (32 + 2 * e + 2 * i - h - k) % 7;
    const m = Math.floor((a + 11 * h + 22 * l) / 451);
    const mes = Math.floor((h + l - 7 * m + 114) / 31); // 3 = março, 4 = abril
    const dia = ((h + l - 7 * m + 114) % 31) + 1;
    return Date.UTC(ano, mes - 1, dia);
}

const cacheFeriados = new Map();

// Feriados nacionais ANBIMA (bancos fechados) para um ano — Set de timestamps UTC.
export function feriadosDoAno(ano) {
    if (cacheFeriados.has(ano)) return cacheFeriados.get(ano);

    const pascoaTs = pascoa(ano);
    const datas = [
        Date.UTC(ano, 0, 1), // Confraternização Universal
        pascoaTs - 48 * DIA_MS, // Carnaval (segunda)
        pascoaTs - 47 * DIA_MS, // Carnaval (terça)
        pascoaTs - 2 * DIA_MS, // Sexta-feira Santa
        Date.UTC(ano, 3, 21), // Tiradentes
        Date.UTC(ano, 4, 1), // Dia do Trabalho
        pascoaTs + 60 * DIA_MS, // Corpus Christi
        Date.UTC(ano, 8, 7), // Independência
        Date.UTC(ano, 9, 12), // N. Sra. Aparecida
        Date.UTC(ano, 10, 2), // Finados
        Date.UTC(ano, 10, 15), // Proclamação da República
        Date.UTC(ano, 11, 25), // Natal
    ];
    if (ano >= 2024) datas.push(Date.UTC(ano, 10, 20)); // Consciência Negra (feriado nacional a partir de 2024)

    const set = new Set(datas);
    cacheFeriados.set(ano, set);
    return set;
}

export function ehDiaUtil(ts) {
    const diaSemana = new Date(ts).getUTCDay(); // 0 = domingo, 6 = sábado
    if (diaSemana === 0 || diaSemana === 6) return false;
    const ano = new Date(ts).getUTCFullYear();
    return !feriadosDoAno(ano).has(ts);
}

// Primeiro dia útil estritamente após `ts`.
export function proximoDiaUtil(ts) {
    let candidato = ts + DIA_MS;
    while (!ehDiaUtil(candidato)) candidato += DIA_MS;
    return candidato;
}

// Os `n` primeiros dias úteis estritamente após `inicioTs` (índice 0 = 1º D.U.).
export function serieDiasUteis(inicioTs, n) {
    const dias = [];
    let atual = inicioTs;
    for (let i = 0; i < n; i++) {
        atual = proximoDiaUtil(atual);
        dias.push(atual);
    }
    return dias;
}

// Conta dias úteis no intervalo (inicioTs, fimTs] — exclui o início, inclui
// o fim (convenção de mercado: dia da aplicação não conta, dia do
// vencimento conta). Devolve 0 se fimTs <= inicioTs.
export function diasUteisEntre(inicioTs, fimTs) {
    if (fimTs <= inicioTs) return 0;
    let count = 0;
    let atual = inicioTs;
    while (atual < fimTs) {
        atual += DIA_MS;
        if (ehDiaUtil(atual)) count++;
    }
    return count;
}

// "YYYY-MM-DD" (valor de <input type="date">) -> timestamp UTC. Devolve
// null se vazio/inválido.
export function parseDataInput(str) {
    if (!str) return null;
    const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(str);
    if (!m) return null;
    return Date.UTC(Number(m[1]), Number(m[2]) - 1, Number(m[3]));
}

export function formatarDataPtBR(ts) {
    if (ts === null || ts === undefined) return "";
    const d = new Date(ts);
    const dia = String(d.getUTCDate()).padStart(2, "0");
    const mes = String(d.getUTCMonth() + 1).padStart(2, "0");
    return `${dia}/${mes}/${d.getUTCFullYear()}`;
}
