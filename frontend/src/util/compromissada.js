// Cálculo da aba "Compromissada": rendimento dia a dia de Conta Remunerada,
// CDB e operação Compromissada, replicando a planilha de referência
// (CalculadoraCompromissada). Taxas sempre como fração decimal (10% = 0.10).

import { serieDiasUteis } from "./feriados.js";

const DIA_MS = 86400000;

// IOF regressivo por dia corrido decorrido (índice = dias corridos, 1..29);
// a partir de 30 dias corridos o IOF zera.
const IOF_POR_DIA = [
    null, // índice 0 não usado
    0.96, 0.93, 0.9, 0.86, 0.83, 0.8, 0.76, 0.73, 0.7, 0.66,
    0.63, 0.6, 0.56, 0.53, 0.5, 0.46, 0.43, 0.4, 0.36, 0.33,
    0.3, 0.26, 0.23, 0.2, 0.16, 0.13, 0.1, 0.06, 0.03,
];

function iofPorDiasCorridos(diasCorridos) {
    if (diasCorridos >= 30) return 0;
    return IOF_POR_DIA[diasCorridos] ?? 0;
}

const ALIQUOTA_IR_CURTO_PRAZO = 0.225; // IR regressivo de até 180 dias — todo o horizonte da aba fica nessa faixa

// calcularCompromissada({ financeiro, diasUteisAno, selic, dataInicioTs,
//   txCR, txCDB, txComp, horizonte, spread })
// -> { cdiDiario, linhas: [{du, dataTs, diasCorridos, iof, contaRemunerada,
//        cdb, compromissada, equivCDB, comissao}], primeiroDiaCDBGanha }
export function calcularCompromissada(params) {
    const { financeiro, diasUteisAno, selic, dataInicioTs, txCR, txCDB, txComp, horizonte, spread } = params;
    const diasAno = diasUteisAno || 252;
    const cdiDiario = Math.round((Math.pow(1 + (selic - 0.001), 1 / diasAno) - 1) * 1e8) / 1e8;

    if (!dataInicioTs || horizonte <= 0) {
        return { cdiDiario, linhas: [], primeiroDiaCDBGanha: null };
    }

    const datas = serieDiasUteis(dataInicioTs, horizonte);

    const linhas = [];
    let primeiroDiaCDBGanha = null;

    for (let idx = 0; idx < datas.length; idx++) {
        const n = idx + 1;
        const dataTs = datas[idx];
        const diasCorridos = Math.round((dataTs - dataInicioTs) / DIA_MS);
        const iof = iofPorDiasCorridos(diasCorridos);

        const contaRemunerada = (Math.pow(1 + txCR * cdiDiario, n) - 1) * financeiro * (1 - iof) * (1 - ALIQUOTA_IR_CURTO_PRAZO);
        const cdb = (Math.pow(1 + txCDB * cdiDiario, n) - 1) * financeiro * (1 - iof) * (1 - ALIQUOTA_IR_CURTO_PRAZO);
        const compromissada = (Math.pow(1 + txComp * cdiDiario, n) - 1) * financeiro * (1 - ALIQUOTA_IR_CURTO_PRAZO);
        const equivCDB = iof < 1 ? txComp / (1 - iof) : null;
        const comissao = (Math.pow(1 + selic - 0.001, n / diasAno) - 1) * (spread || 0) * financeiro;

        if (primeiroDiaCDBGanha === null && cdb >= compromissada) primeiroDiaCDBGanha = n;

        linhas.push({ du: n, dataTs, diasCorridos, iof, contaRemunerada, cdb, compromissada, equivCDB, comissao });
    }

    return { cdiDiario, linhas, primeiroDiaCDBGanha };
}
