// Motor de cálculo da Calculadora Previdenciária — porta fiel da planilha
// "SimuladorPGBL2022v33.xlsx" (abas Simulador + Tabelas) enviada pelo
// usuário: compara declaração Simplificada vs. Completa vs. Completa com
// 12% em PGBL, em cima da renda anual (mês × 12 + 1/3, considerando 13º e
// 1/3 de férias).
//
// Duas tabelas de referência, alternáveis pelo usuário em Configurações:
// - "2022": valores exatamente como na planilha original (fonte primária).
// - "2026": IRPF/INSS vigentes (Lei 15.191/2025 e Lei 15.270/2025),
//   levantados via pesquisa pública em jul/2026 — não a partir do texto
//   oficial da lei. Os campos marcados "ASSUNÇÃO" abaixo não têm confirmação
//   de terem mudado; foram mantidos iguais a 2022 por falta de indício em
//   contrário. Ver aviso exibido na aba quando esse modo está ativo.

// lookupProgressivo replica VLOOKUP(valor, faixas, coluna, 1) do Excel:
// acha a maior faixa cujo `ini` é <= valor, e aplica valor*aliq - parcela
// (fórmula-padrão de "parcela a deduzir" da tabela progressiva brasileira).
export function lookupProgressivo(valor, faixas) {
    let faixa = faixas[0];
    for (const f of faixas) {
        if (valor >= f.ini) faixa = f;
        else break;
    }
    return valor * faixa.aliq - faixa.parcela;
}

const IRPF_2022 = [
    { ini: 0, aliq: 0, parcela: 0 },
    { ini: 22847.77, aliq: 0.075, parcela: 1713.58 },
    { ini: 33919.81, aliq: 0.15, parcela: 4257.57 },
    { ini: 45012.61, aliq: 0.225, parcela: 7633.51 },
    { ini: 55976.17, aliq: 0.275, parcela: 10432.32 },
];

const INSS_2022 = [
    { ini: 0, aliq: 0.075, parcela: 0 },
    { ini: 14544.12, aliq: 0.09, parcela: 218.16 },
    { ini: 29128.32, aliq: 0.12, parcela: 1092.01 },
    { ini: 43692.48, aliq: 0.14, parcela: 1965.86 },
];

const IRPF_2026 = [
    { ini: 0, aliq: 0, parcela: 0 },
    { ini: 29145.6, aliq: 0.075, parcela: 2185.92 },
    { ini: 33919.8, aliq: 0.15, parcela: 4729.92 },
    { ini: 45012.6, aliq: 0.225, parcela: 8105.88 },
    { ini: 55976.16, aliq: 0.275, parcela: 10904.76 },
];

const INSS_2026 = [
    { ini: 0, aliq: 0.075, parcela: 0 },
    { ini: 19452.0, aliq: 0.09, parcela: 291.78 },
    { ini: 34834.08, aliq: 0.12, parcela: 1336.8024 },
    { ini: 52251.24, aliq: 0.14, parcela: 2381.8272 },
];

// Redutor de IR criado pela Lei 15.270/2025: zera o IR apurado para renda
// anual até R$60.000 (R$5.000/mês) e reduz linearmente até zerar o desconto
// em R$88.200 (R$7.350/mês). Fórmula anualizada a partir da mensal oficial
// (978,62 - 0,133145×rendimento mensal), conferida contra exemplo numérico
// oficial (salário R$5.800 + INSS R$613,51 → IR devido R$311,17).
function redutorAnual2026(rendaAnual) {
    if (rendaAnual <= 60000) return Infinity;
    if (rendaAnual > 88200) return 0;
    return 11743.44 - 0.133145 * rendaAnual;
}

export const TABELAS = {
    2022: {
        label: "2022 (planilha original)",
        irpfAnual: IRPF_2022,
        inssAnual: INSS_2022,
        tetoInssAnual: 7087.22 * 12,
        maxInssAnual: 7087.22 * 0.14 * 12 - 1965.86,
        tetoEducacaoAnual: 3561.5,
        tetoDependenteAnual: 2275.08,
        tetoSimplificadaAnual: 16754.34,
        redutor: null,
        aviso: null,
    },
    2026: {
        label: "2026 (Lei 15.191/2025 e 15.270/2025)",
        irpfAnual: IRPF_2026,
        inssAnual: INSS_2026,
        tetoInssAnual: 8475.55 * 12,
        maxInssAnual: 8475.55 * 0.14 * 12 - 2381.8272,
        tetoEducacaoAnual: 3561.5, // ASSUNÇÃO: sem indício de reajuste, mantido de 2022
        tetoDependenteAnual: 2275.08, // confirmado igual (R$189,59/mês × 12)
        tetoSimplificadaAnual: 16754.34, // ASSUNÇÃO: sem indício de reajuste, mantido de 2022
        redutor: redutorAnual2026,
        aviso:
            "Tabela 2026 levantada a partir de fontes públicas (não do texto oficial da lei) em jul/2026. " +
            "Os tetos de educação e de dedução simplificada anual não têm confirmação de terem mudado e foram " +
            "mantidos iguais a 2022. Valide antes de usar com cliente.",
    },
};

// irComRedutor aplica a tabela progressiva e, se a tabela tiver redutor
// (2026), desconta-o do IR apurado sem deixar o resultado negativo.
function irComRedutor(baseCalculo, rendaAnual, tabela) {
    const bruto = lookupProgressivo(Math.max(baseCalculo, 0), tabela.irpfAnual);
    if (!tabela.redutor) return bruto;
    return Math.max(0, bruto - tabela.redutor(rendaAnual));
}

function calcInssAnual(rendaAnual, tabela) {
    if (rendaAnual > tabela.tetoInssAnual) return tabela.maxInssAnual;
    return lookupProgressivo(rendaAnual, tabela.inssAnual);
}

// calcularPrevidenciaria replica linha a linha o Simulador da planilha.
// `inputs` = { rendaBrutaMes, gastoSaudeMes, qtdDependentes,
// gastoEducPropriaMes, gastoEducDependentesMes, contribuicaoPgblAno }
// (todos já convertidos pra número, 0 quando vazio). `anoTabela` é "2022"
// ou "2026" (chave de TABELAS).
export function calcularPrevidenciaria(inputs, anoTabela) {
    const tabela = TABELAS[anoTabela] || TABELAS[2026];
    const { rendaBrutaMes, gastoSaudeMes, qtdDependentes, gastoEducPropriaMes, gastoEducDependentesMes, contribuicaoPgblAno } = inputs;

    const rendaAnual = rendaBrutaMes * (12 + 1 / 3);
    const inssAnual = calcInssAnual(rendaAnual, tabela);
    const saudeAnual = gastoSaudeMes * 12;
    const eduPropriaAnual = Math.min(gastoEducPropriaMes * 12, tabela.tetoEducacaoAnual);
    const eduDependentesAnual = Math.min(gastoEducDependentesMes * 12, tabela.tetoEducacaoAnual * qtdDependentes);
    const dependentesAnual = qtdDependentes * tabela.tetoDependenteAnual;
    const limitePgbl = rendaAnual * 0.12;

    const col = {
        simplificada: { inss: 0, saude: 0, eduPropria: 0, eduDependentes: 0, dependentes: 0, pgbl: 0 },
        completa: { inss: inssAnual, saude: saudeAnual, eduPropria: eduPropriaAnual, eduDependentes: eduDependentesAnual, dependentes: dependentesAnual, pgbl: Math.min(contribuicaoPgblAno, limitePgbl) },
        completaPgbl12: { inss: inssAnual, saude: saudeAnual, eduPropria: eduPropriaAnual, eduDependentes: eduDependentesAnual, dependentes: dependentesAnual, pgbl: limitePgbl },
    };

    const deducaoSimplificada = Math.min(rendaAnual * 0.2, tabela.tetoSimplificadaAnual);

    const baseSimplificada = rendaAnual - deducaoSimplificada;
    const baseCompleta = Math.max(rendaAnual - (col.completa.inss + col.completa.saude + col.completa.eduPropria + col.completa.eduDependentes + col.completa.dependentes + col.completa.pgbl), 0);
    const baseCompletaPgbl12 = Math.max(rendaAnual - (col.completaPgbl12.inss + col.completaPgbl12.saude + col.completaPgbl12.eduPropria + col.completaPgbl12.eduDependentes + col.completaPgbl12.dependentes + col.completaPgbl12.pgbl), 0);

    const baseIrFonte = rendaAnual - inssAnual;
    const irFonte = irComRedutor(baseIrFonte, rendaAnual, tabela); // igual nas 3 colunas (mesma fórmula na planilha)

    const irDevidoSimplificada = irComRedutor(baseSimplificada, rendaAnual, tabela);
    const irDevidoCompleta = irComRedutor(baseCompleta, rendaAnual, tabela);
    const irDevidoCompletaPgbl12 = irComRedutor(baseCompletaPgbl12, rendaAnual, tabela);

    const linhas = {
        rendaAnual: { simplificada: rendaAnual, completa: rendaAnual, completaPgbl12: rendaAnual },
        inss: { simplificada: col.simplificada.inss, completa: col.completa.inss, completaPgbl12: col.completaPgbl12.inss },
        saude: { simplificada: col.simplificada.saude, completa: col.completa.saude, completaPgbl12: col.completaPgbl12.saude },
        eduPropria: { simplificada: col.simplificada.eduPropria, completa: col.completa.eduPropria, completaPgbl12: col.completaPgbl12.eduPropria },
        eduDependentes: { simplificada: col.simplificada.eduDependentes, completa: col.completa.eduDependentes, completaPgbl12: col.completaPgbl12.eduDependentes },
        dependentes: { simplificada: col.simplificada.dependentes, completa: col.completa.dependentes, completaPgbl12: col.completaPgbl12.dependentes },
        pgbl: { simplificada: col.simplificada.pgbl, completa: col.completa.pgbl, completaPgbl12: col.completaPgbl12.pgbl },
        deducaoSimplificada: { simplificada: deducaoSimplificada, completa: 0, completaPgbl12: 0 },
        baseTributavel: { simplificada: baseSimplificada, completa: baseCompleta, completaPgbl12: baseCompletaPgbl12 },
        irFonte: { simplificada: irFonte, completa: irFonte, completaPgbl12: irFonte },
        irDevido: { simplificada: irDevidoSimplificada, completa: irDevidoCompleta, completaPgbl12: irDevidoCompletaPgbl12 },
        restituicao: {
            simplificada: irFonte - irDevidoSimplificada,
            completa: irFonte - irDevidoCompleta,
            completaPgbl12: irFonte - irDevidoCompletaPgbl12,
        },
    };

    return {
        linhas,
        limitePgbl,
        contribuicaoPgblAno,
        diagnosticoPgbl: diagnosticoPgbl(contribuicaoPgblAno, limitePgbl),
        diagnosticoEconomia: diagnosticoEconomia(irDevidoSimplificada, irDevidoCompleta, irDevidoCompletaPgbl12),
        tabela,
    };
}

function diagnosticoPgbl(contribuicao, limite) {
    const diferenca = Math.abs(contribuicao - limite);
    if (contribuicao < limite) {
        return `Ainda falta R$ ${formatarBR(diferenca)} para atingir o limite dedutível de PGBL.`;
    }
    if (contribuicao > limite) {
        return `Limite dedutível de PGBL ultrapassado em R$ ${formatarBR(diferenca)}.`;
    }
    return "Contribuição de PGBL exatamente no limite dedutível.";
}

function diagnosticoEconomia(irSimplificada, irCompleta, irCompletaPgbl12) {
    const difCompleta = irSimplificada - irCompleta;
    const difCompletaPgbl = irSimplificada - irCompletaPgbl12;
    if (difCompleta > 0) {
        let texto = `Economia de R$ ${formatarBR(difCompleta)} no ano caso seja feita a declaração completa.`;
        if (irCompleta !== irCompletaPgbl12) {
            texto += ` Com a contribuição de PGBL no limite de 12%, a economia sobe para R$ ${formatarBR(difCompletaPgbl)}.`;
        }
        return texto;
    }
    if (difCompleta === 0) {
        return "Sem diferença entre os regimes: recomenda-se a declaração simplificada, com eventual contribuição em VGBL.";
    }
    if (difCompletaPgbl > 0) {
        return `Economia de R$ ${formatarBR(difCompletaPgbl)} no ano caso seja feita a declaração completa com 12% em PGBL.`;
    }
    return `Economia de R$ ${formatarBR(-difCompleta)} no ano caso seja feita a declaração simplificada.`;
}

function formatarBR(valor) {
    return Math.abs(valor).toLocaleString("pt-BR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
