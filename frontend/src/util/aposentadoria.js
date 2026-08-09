// Planejamento de aposentadoria: quanto contribuir por mês pra se
// aposentar com a renda desejada, em dois cenários — consumindo o
// patrimônio ao longo da aposentadoria (com ou sem deixar herança) ou
// vivendo só do rendimento sem nunca consumir o principal (perpetuidade).
//
// Os dois cenários usam o mesmo prazo de acumulação (idade de aposentadoria
// − idade atual) pra calcular a parcela mensal necessária — a planilha de
// referência usava prazos diferentes em cada cenário, o que foi confirmado
// com o usuário como uma inconsistência a corrigir, não replicar.

import { valorFuturo, aporteParaValorFuturo } from "./financeiro.js";

// Valor presente, na data de aposentadoria, de `rendaMensal` sacada por
// `nConsumoMeses`, terminando com `sucessao` de saldo (valor presente de
// uma anuidade + valor presente de um valor futuro).
function patrimonioParaConsumo(rendaMensal, i, nConsumoMeses, sucessao) {
    if (nConsumoMeses <= 0) return null;
    if (i === 0) return rendaMensal * nConsumoMeses + sucessao;
    const fator = Math.pow(1 + i, nConsumoMeses);
    return (rendaMensal * (1 - 1 / fator)) / i + sucessao / fator;
}

// Patrimônio cujo rendimento mensal sozinho já cobre `rendaMensal`, pra
// sempre — não depende de quantos meses dura a aposentadoria.
function patrimonioParaPreservacao(rendaMensal, i) {
    if (i <= 0) return null; // sem taxa real positiva não há perpetuidade sustentável
    return rendaMensal / i;
}

// Parcela mensal pra acumular `alvo` em `nAcumMeses` partindo de `pv0` já
// aplicado. Nunca negativa — patrimônio já suficiente vira "R$ 0" (ver
// mensagem no chamador), não um valor negativo sem sentido.
function parcelaNecessaria(pv0, i, nAcumMeses, alvo) {
    if (alvo === null || nAcumMeses <= 0) return null;
    const pmt = aporteParaValorFuturo(pv0, i, nAcumMeses, alvo);
    return pmt === null ? null : Math.max(0, pmt);
}

// Evolução do patrimônio ano a ano — projeção REAL (não a meta teórica):
// acumulação (0..nAcumMeses/12) usando o aporte mensal que o cliente
// informou que já faz hoje, seguida da fase de aposentadoria (saques de
// `rendaMensal`) partindo do patrimônio que esse aporte real de fato vai
// gerar até lá. Se o aporte real for maior que a parcela mínima necessária,
// o patrimônio dura mais que `nConsumoMeses` (ou até cresce, no cenário de
// preservação); se for menor, esgota antes — os dois são resultados válidos
// e é isso que o gráfico deve mostrar, não a meta batendo certinho sempre.
// nAcumMeses/nConsumoMeses são sempre múltiplos de 12 (diferenças de idade
// em anos inteiros), então a amostragem anual cai exatamente nos limites de
// cada fase. `anoRelativo` é a idade absoluta do cliente (idadeAtual + anos
// decorridos), não anos restantes — o gráfico mostra a linha da vida toda,
// não só o horizonte daqui pra frente.
function serieAnual(pv0, aporteMensalReal, i, nAcumMeses, rendaMensal, nConsumoMeses, idadeAtual) {
    const pontos = [];
    const anosAcum = nAcumMeses / 12;
    const anosConsumo = nConsumoMeses / 12;
    for (let ano = 0; ano <= anosAcum; ano++) {
        pontos.push({ anoRelativo: idadeAtual + ano, fase: "acumulacao", patrimonio: valorFuturo(pv0, aporteMensalReal, i, ano * 12) });
    }
    const patrimonioNaAposentadoria = valorFuturo(pv0, aporteMensalReal, i, nAcumMeses);
    for (let ano = 1; ano <= anosConsumo; ano++) {
        const patrimonio = valorFuturo(patrimonioNaAposentadoria, -rendaMensal, i, ano * 12);
        pontos.push({ anoRelativo: idadeAtual + anosAcum + ano, fase: "consumo", patrimonio: Math.max(0, patrimonio) });
    }
    return pontos;
}

// calcularAposentadoria(params) -> { taxaMensal, nAcumMeses, nConsumoMeses,
//   rendaNecessaria, consumo, preservacao }
// `consumo`/`preservacao` são null quando os parâmetros não permitem
// calcular aquele cenário (ver casos de borda abaixo); cada um, quando
// presente, tem { patrimonioNecessario, parcelaMensal, serieAnual }.
export function calcularAposentadoria(params) {
    const {
        idadeAtual,
        idadeAposentadoria,
        expectativaVida,
        aplicacoesFinanceiras,
        aplicacaoMensal,
        rendaDesejada,
        rendaINSS,
        outrasFontes,
        taxaAnual,
        patrimonioSucessao,
    } = params;

    const taxaMensal = Math.pow(1 + taxaAnual, 1 / 12) - 1;
    const nAcumMeses = Math.round((idadeAposentadoria - idadeAtual) * 12);
    const nConsumoMeses = Math.round((expectativaVida - idadeAposentadoria) * 12);
    const rendaNecessaria = Math.max(0, rendaDesejada - rendaINSS - outrasFontes);

    // Sem prazo de acumulação não dá pra calcular parcela nenhuma —
    // devolve só os números derivados, os dois cenários ficam null (o
    // chamador mostra o aviso).
    if (nAcumMeses <= 0) {
        return { taxaMensal, nAcumMeses, nConsumoMeses, rendaNecessaria, consumo: null, preservacao: null };
    }

    let consumo = null;
    if (nConsumoMeses > 0) {
        const patrimonioNecessario = patrimonioParaConsumo(rendaNecessaria, taxaMensal, nConsumoMeses, patrimonioSucessao);
        const parcelaMensal = parcelaNecessaria(aplicacoesFinanceiras, taxaMensal, nAcumMeses, patrimonioNecessario);
        consumo = {
            patrimonioNecessario,
            parcelaMensal,
            serieAnual: serieAnual(aplicacoesFinanceiras, aplicacaoMensal, taxaMensal, nAcumMeses, rendaNecessaria, nConsumoMeses, idadeAtual),
        };
    }

    let preservacao = null;
    const patrimonioPreservacao = patrimonioParaPreservacao(rendaNecessaria, taxaMensal);
    if (patrimonioPreservacao !== null) {
        const parcelaMensal = parcelaNecessaria(aplicacoesFinanceiras, taxaMensal, nAcumMeses, patrimonioPreservacao);
        // O patrimônio de preservação fica achatado por definição (o saque
        // é exatamente o rendimento sustentável) — o comprimento da fase de
        // "consumo" no gráfico é só estético, usa o mesmo da aposentadoria
        // por consumo quando disponível, ou 30 anos como padrão razoável.
        const nConsumoGrafico = nConsumoMeses > 0 ? nConsumoMeses : 360;
        preservacao = {
            patrimonioNecessario: patrimonioPreservacao,
            parcelaMensal,
            serieAnual: serieAnual(aplicacoesFinanceiras, aplicacaoMensal, taxaMensal, nAcumMeses, rendaNecessaria, nConsumoGrafico, idadeAtual),
        };
    }

    return { taxaMensal, nAcumMeses, nConsumoMeses, rendaNecessaria, consumo, preservacao };
}
