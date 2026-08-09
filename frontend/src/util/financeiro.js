// Fórmulas de juros compostos usadas na aba "Calculadora Financeira".
// Convenção: taxa sempre como fração decimal (1% = 0.01); n em número de
// períodos (o significado do período — mês ou ano — depende de cada
// calculadora, ver frontend/src/tabs/calculadora.js).

// Valor futuro de um aporte inicial (pv) + aportes periódicos constantes
// (pmt), à taxa i por período, ao longo de n períodos.
export function valorFuturo(pv, pmt, i, n) {
    if (i === 0) return pv + pmt * n;
    const fator = Math.pow(1 + i, n);
    return pv * fator + (pmt * (fator - 1)) / i;
}

// Número de períodos necessários pra ir de pv (+ aportes pmt por período)
// até fv, à taxa i por período. Forma fechada via logaritmo (sem pmt seria
// só ln(fv/pv)/ln(1+i); com pmt, isola-se (1+i)^n antes de aplicar o log).
// Devolve null se os parâmetros não têm solução (ex.: fv já foi
// ultrapassado, ou não há crescimento possível).
export function periodosParaValorFuturo(pv, pmt, i, fv) {
    if (i === 0) {
        if (pmt === 0) return null;
        const n = (fv - pv) / pmt;
        return n > 0 ? n : null;
    }
    const base = pv + pmt / i;
    if (base === 0) return null;
    const x = (fv + pmt / i) / base;
    if (x <= 0) return null;
    const n = Math.log(x) / Math.log(1 + i);
    return n > 0 ? n : null;
}

// Aporte periódico necessário pra ir de pv até fv em n períodos, à taxa i.
export function aporteParaValorFuturo(pv, i, n, fv) {
    const fator = Math.pow(1 + i, n);
    if (i === 0 || fator === 1) return n !== 0 ? (fv - pv) / n : null;
    return ((fv - pv * fator) * i) / (fator - 1);
}

// Taxa por período necessária pra ir de pv (+ aportes pmt) até fv em n
// períodos. Sem forma fechada quando pmt != 0 — resolvido por bisseção.
// Devolve null se não houver raiz no intervalo pesquisado.
export function taxaParaValorFuturo(pv, pmt, n, fv) {
    let lo = -0.9999;
    let hi = 10; // -99,99% a 1000% por período — faixa generosa o bastante
    const f = (i) => valorFuturo(pv, pmt, i, n) - fv;
    let fLo = f(lo);
    let fHi = f(hi);
    if (!Number.isFinite(fLo) || !Number.isFinite(fHi) || fLo * fHi > 0) return null;
    for (let iter = 0; iter < 100; iter++) {
        const meio = (lo + hi) / 2;
        const fMeio = f(meio);
        if (Math.abs(fMeio) < 1e-9) return meio;
        if (fLo * fMeio < 0) {
            hi = meio;
            fHi = fMeio;
        } else {
            lo = meio;
            fLo = fMeio;
        }
    }
    return (lo + hi) / 2;
}

// Converte uma taxa acumulada em `deN` períodos-base pra taxa equivalente
// em `paraN` períodos-base (juros compostos). Ex.: taxa anual -> mensal é
// taxaEquivalente(taxaAno, 12, 1); mensal -> anual é
// taxaEquivalente(taxaMes, 1, 12).
export function taxaEquivalente(taxa, deN, paraN) {
    return Math.pow(1 + taxa, paraN / deN) - 1;
}

// Taxa real a partir da taxa nominal e da inflação no mesmo período
// (equação de Fisher).
export function taxaReal(taxaNominal, inflacao) {
    return (1 + taxaNominal) / (1 + inflacao) - 1;
}

// Alíquota de IR regressivo (renda fixa), por prazo em meses. Limiares em
// meses (não em dias corridos) — confirmado com o usuário que 24 meses já
// cai na faixa de 15%.
export function aliquotaIR(prazoMeses) {
    if (prazoMeses <= 6) return 0.225;
    if (prazoMeses <= 12) return 0.2;
    if (prazoMeses < 24) return 0.175;
    return 0.15;
}

// Alíquota de IR regressivo por dias corridos (usada no Comparador de Renda
// Fixa, que trabalha com datas de aplicação/vencimento em vez de prazo em
// meses). Faixas oficiais: até 180 dias 22,5%, até 360 dias 20%, até 720
// dias 17,5%, acima disso 15%.
export function aliquotaIRDias(diasCorridos) {
    if (diasCorridos <= 180) return 0.225;
    if (diasCorridos <= 360) return 0.2;
    if (diasCorridos <= 720) return 0.175;
    return 0.15;
}
