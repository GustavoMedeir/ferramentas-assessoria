# Modelos de E-mail de Ordem

Gerado a partir do app — cada bloco abaixo é uma categoria de operação.
Os textos entre `{chaves}` são os campos que o usuário preenche no formulário;
mantenha as `{chaves}` exatamente como estão (é onde o valor digitado entra).
Quando terminar de editar, me devolva o arquivo que eu aplico as mudanças no app.

---

## Renda Fixa

### Aplicação

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a aplicação

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Nome do ativo)
- `{emissor}` — Emissor (ex.: Nome do emissor)
- `{taxaMinima}` — Taxa de rentabilidade mínima (ex.: Ex: CDI + 2%)
- `{vencimento}` — Vencimento (ex.: Data e/ou nº de dias a partir da aplicação)
- `{valorLimite}` — Valor limite a ser aplicado (ex.: R$ ...)

**Corpo do e-mail:**

```
Ativo: {ativo};
Emissor: {emissor};
Taxa de rentabilidade mínima: {taxaMinima};
Vencimento: {vencimento};
Valor limite a ser aplicado: {valorLimite}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a aplicação.

---

### Aplicação Cotada

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a aplicação

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Nome do ativo)
- `{emissor}` — Emissor (ex.: Nome do emissor)
- `{taxa}` — Taxa (ex.: Taxa de rentabilidade)
- `{vencimento}` — Vencimento (ex.: Data e/ou nº de dias)
- `{valorAplicado}` — Valor a ser aplicado (cotado) (ex.: R$ ...)

**Corpo do e-mail:**

```
Ativo: {ativo};
Emissor: {emissor};
Taxa: {taxa};
Vencimento: {vencimento};
Valor a ser aplicado: {valorAplicado}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a aplicação.

---

### Resgate

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** o resgate

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Nome do ativo)
- `{emissor}` — Emissor (ex.: Nome do emissor)
- `{vencimento}` — Vencimento (ex.: MÊS/ANO)
- `{taxaMaxima}` — Taxa máxima (ex.: Taxa máxima que o cliente aceita vender o título)
- `{valorOuQtd}` — Valor a ser resgatado ou Quantidade (ex.: R$ ... ou Quantidade: xxx)

**Corpo do e-mail:**

```
Ativo: {ativo};
Emissor: {emissor};
Vencimento: {vencimento};
Taxa máxima: {taxaMaxima};
Valor a ser resgatado: {valorOuQtd}.
```

**Frase de fechamento:** Aguardo confirmação para realizar o resgate.

---

### Resgate Cotado

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** o resgate

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Nome do ativo)
- `{emissor}` — Emissor (ex.: Nome do emissor)
- `{vencimento}` — Vencimento (ex.: MÊS/ANO)
- `{taxa}` — Taxa (ex.: Taxa de rentabilidade)
- `{quantidade}` — Quantidade a ser resgatada (ex.: xxx)
- `{valor}` — Valor (cotado) (ex.: R$ ...)

**Corpo do e-mail:**

```
Ativo: {ativo};
Emissor: {emissor};
Vencimento: {vencimento};
Taxa: {taxa};
Quantidade a ser resgatada: {quantidade};
Valor: {valor}.
```

**Frase de fechamento:** Aguardo confirmação para realizar o resgate.

---

## Tesouro Direto

### Aplicação

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a aplicação no Tesouro Direto

**Campos do formulário:**

- `{titulo}` — Título (ex.: Nome do título)
- `{tipo}` — Tipo (opções: Pré-fixado, Pós-fixado)
- `{rentabilidade}` — Rentabilidade ao ano (ex.: Taxa ou indexador)
- `{vencimento}` — Vencimento (ex.: Data de vencimento)
- `{pagamentoJuros}` — Pagamento de juros (ex.: No vencimento ou Cupons semestrais)
- `{valorOuQtd}` — Valor a ser aplicado ou Quantidade (ex.: R$ ... ou Quantidade: xxx)

**Corpo do e-mail:**

```
Título: {titulo};
Tipo: {tipo};
Rentabilidade ao ano: {rentabilidade};
Vencimento: {vencimento};
Pagamento de juros: {pagamentoJuros};
Valor a ser aplicado: {valorOuQtd}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a aplicação.

---

### Resgate

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** o resgate no Tesouro Direto

**Campos do formulário:**

- `{titulo}` — Título (ex.: Nome do título)
- `{vencimento}` — Vencimento (ex.: Data de vencimento)
- `{quantidade}` — Quantidade a ser resgatada (ex.: xxx)
- `{valor}` — Valor financeiro (ex.: R$ ...)

**Corpo do e-mail:**

```
Título: {titulo};
Vencimento: {vencimento};
Quantidade a ser resgatada: {quantidade};
Valor financeiro: {valor}.
```

**Frase de fechamento:** Aguardo confirmação para realizar o resgate.

---

## Fundos de Investimento

### Aplicação

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a aplicação no Fundo de Investimentos

**Campos do formulário:**

- `{fundo}` — Fundo (ex.: Nome do fundo)
- `{tipo}` — Tipo (ex.: Internacional, Renda Fixa, Multimercados, Ações, Cambial...)
- `{valor}` — Valor a ser aplicado (ex.: R$ ...)
- `{cotizacao}` — Cotização (resgate) (ex.: D+x)
- `{liquidacao}` — Liquidação financeira (resgate) (ex.: D+y após a data de cotização)

**Corpo do e-mail:**

```
Fundo: {fundo};
Tipo: {tipo};
Valor a ser aplicado: {valor}.
Seguem informações sobre a liquidez do fundo referido para resgate:
- Cotização: {cotizacao};
- Liquidação financeira: {liquidacao}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a aplicação.

---

### Resgate

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** o resgate do Fundo de Investimentos

**Campos do formulário:**

- `{fundo}` — Fundo (ex.: Nome do fundo)
- `{valor}` — Valor a ser resgatado (ex.: R$ ... ou Resgate Total)
- `{cotizacao}` — Cotização (ex.: D+x)
- `{liquidacao}` — Liquidação financeira (ex.: D+y após a data de cotização)

**Corpo do e-mail:**

```
Fundo: {fundo};
Valor a ser resgatado: {valor};
Cotização: {cotizacao};
Liquidação Financeira: {liquidacao}.
```

**Frase de fechamento:** Aguardo confirmação para realizar o resgate.

---

## COE

### Reserva de Compra

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a reserva de compra do COE

**Campos do formulário:**

- `{coe}` — COE (ex.: Nome do COE)
- `{emissor}` — Emissor (ex.: Emissor do COE)
- `{tipo}` — Tipo (opções: Valor Nominal Protegido, Em Risco)
- `{quantidade}` — Quantidade a ser reservada (ex.: xxx)
- `{valor}` — Valor da reserva (ex.: R$ ...)
- `{vencimento}` — Vencimento (ex.: xx/xx/xxxx)

**Corpo do e-mail:**

```
COE: {coe};
Emissor: {emissor};
Tipo: {tipo};
Quantidade: {quantidade};
Valor: {valor};
Vencimento: {vencimento}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a reserva.

---

### Venda

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a venda do COE

**Campos do formulário:**

- `{coe}` — COE (ex.: Nome do COE)
- `{emissor}` — Emissor (ex.: Emissor do COE)
- `{tipo}` — Tipo (opções: Valor Nominal Protegido, Em Risco)
- `{quantidade}` — Quantidade a ser resgatada (ex.: xxx)
- `{valor}` — Valor bruto do resgate (ex.: R$ ...)
- `{vencimento}` — Vencimento (ex.: xx/xx/xxxx)

**Corpo do e-mail:**

```
COE: {coe};
Emissor: {emissor};
Tipo: {tipo};
Quantidade: {quantidade};
Valor: {valor};
Vencimento: {vencimento}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a venda.

---

## Ofertas Públicas

### Oferta Pública RF - Reserva

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a reserva na Oferta Pública

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Nome do ativo)
- `{emissor}` — Emissor (ex.: Nome do emissor)
- `{taxa}` — Taxa mínima (ex.: Taxa mínima)
- `{quantidade}` — Quantidade a ser reservada (ex.: se aplicável)
- `{valor}` — Valor da reserva (ex.: R$ ...)
- `{carencia}` — Carência (ex.: xx/xx/xxxx ou No vencimento)
- `{vencimento}` — Vencimento (ex.: xx/xx/xxxx)

**Corpo do e-mail:**

```
Ativo: {ativo};
Emissor: {emissor};
Taxa: {taxa};
Quantidade: {quantidade};
Valor: {valor};
Carência: {carencia};
Vencimento: {vencimento}.
```

**Aviso de anexo:** Enviar em anexo ao e-mail o Prospecto oficial em PDF da Oferta Pública.

**Frase de fechamento:** Aguardo confirmação para realizar a reserva.

---

### IPO - Reserva

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a reserva na Oferta Pública

**Campos do formulário:**

- `{oferta}` — Oferta (ex.: Nome da oferta)
- `{valor}` — Valor da reserva (ex.: R$ ...)
- `{preco}` — Preço limite (se desejado) (ex.: R$ ...)
- `{vinculada}` — Cliente pessoa vinculada (opções: Não, Sim)

**Corpo do e-mail:**

```
Oferta: {oferta};
Valor: {valor};
Preço: {preco}.
Cliente pessoa vinculada: {vinculada}.
```

**Aviso de anexo:** Enviar em anexo ao e-mail o Prospecto oficial em PDF da Oferta Pública.

**Frase de fechamento:** Aguardo confirmação para realizar a reserva.

---

### Oferta Restrita - Reserva

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a reserva na Oferta Restrita

**Campos do formulário:**

- `{oferta}` — Oferta (ex.: Nome da oferta)
- `{quantidade}` — Quantidade a ser reservada (ex.: xxx)
- `{preco}` — Preço limite (se desejado) (ex.: R$ ...)
- `{vinculada}` — Cliente pessoa vinculada (opções: Não, Sim)

**Corpo do e-mail:**

```
Oferta: {oferta};
Quantidade: {quantidade};
Preço: {preco}.
Cliente pessoa vinculada: {vinculada}.
```

**Aviso de anexo:** Enviar em anexo ao e-mail o documento oficial de "Fato Relevante" em PDF da Oferta.

**Frase de fechamento:** Aguardo confirmação para realizar a reserva.

---

### OPA - Venda

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a venda na OPA

**Campos do formulário:**

- `{nomeOpa}` — Nome da OPA (ex.: Nome da OPA)
- `{ativo}` — Ativo (ex.: Código ou nome do papel)
- `{quantidade}` — Quantidade a ser vendida (ex.: xxx)
- `{preco}` — Preço mínimo do ativo (ex.: R$ ...)
- `{vinculada}` — Cliente pessoa vinculada (opções: Não, Sim)

**Corpo do e-mail:**

```
Nome: {nomeOpa};
Ativo: {ativo};
Quantidade: {quantidade};
Preço: {preco}.
Cliente pessoa vinculada: {vinculada}.
```

**Aviso de anexo:** Enviar em anexo ao e-mail o Edital oficial da OPA em PDF.

**Frase de fechamento:** Aguardo confirmação para realizar a operação.

---

## Subscrição

### Exercício

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** o exercício do direito de subscrição

**Campos do formulário:**

- `{codigo}` — Código da ação a ser subscrita (ex.: Código)
- `{preco}` — Preço unitário (ex.: R$ ... por papel)
- `{quantidade}` — Quantidade a ser subscrita (ex.: xxx)

**Corpo do e-mail:**

```
Código: {codigo};
Ao Preço unitário de: {preco};
Quantidade: {quantidade}.
```

**Aviso de anexo:** Enviar em anexo ao e-mail o documento oficial em PDF de Aviso aos Acionistas.

**Frase de fechamento:** Aguardo confirmação para realizar a operação.

---

### Negociação de Direitos

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a negociação de direitos de subscrição

**Campos do formulário:**

- `{codigo}` — Código do direito de subscrição (ex.: Código)
- `{natureza}` — Natureza (opções: Compra, Venda)
- `{preco}` — Preço limite (ex.: R$ ...)
- `{quantidade}` — Quantidade a ser negociada (ex.: xxx)

**Corpo do e-mail:**

```
Código: {codigo};
Natureza: {natureza};
Preço: {preco};
Quantidade: {quantidade}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a operação.

---

## Operações Societárias

### Conversão Voluntária

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a conversão

**Campos do formulário:**

- `{papelConvertido}` — Papel a ser convertido (ex.: Código do ativo)
- `{qtdConvertida}` — Quantidade a ser convertida (ex.: xxx)
- `{papelApos}` — Papel após conversão (ex.: Código do ativo)
- `{qtdApos}` — Quantidade após conversão (ex.: xxx)

**Corpo do e-mail:**

```
Papel a ser convertido: {papelConvertido};
Quantidade a ser convertida: {qtdConvertida};
Papel após conversão: {papelApos};
Quantidade após conversão: {qtdApos}.
```

**Frase de fechamento:** Aguardo confirmação para prosseguir com a conversão.

---

### Dissidência

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a dissidência

**Campos do formulário:**

- `{acao}` — Ação a ser entregue (ex.: Código do ativo)
- `{quantidade}` — Quantidade (ex.: xxx)
- `{valorReembolso}` — Valor de reembolso unitário (ex.: R$ ... por ação)

**Corpo do e-mail:**

```
Ação a ser entregue: {acao};
Quantidade: {quantidade};
Ao valor de reembolso unitário de: {valorReembolso}.
```

**Frase de fechamento:** Aguardo confirmação para prosseguir com a operação.

---

## Clubes de Investimento

### Aplicação

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a aplicação no clube de investimentos

**Campos do formulário:**

- `{nomeClube}` — Nome do Clube (ex.: Nome)
- `{codigoClube}` — Código do Clube (ex.: Código)
- `{valor}` — Valor a ser aplicado (ex.: R$ ...)

**Corpo do e-mail:**

```
Nome do Clube: {nomeClube};
Código: {codigoClube};
Valor a ser aplicado: {valor}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a aplicação.

---

### Resgate

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** o resgate no clube de investimentos

**Campos do formulário:**

- `{nomeClube}` — Nome do Clube (ex.: Nome)
- `{codigoClube}` — Código do Clube (ex.: Código)
- `{valor}` — Valor do resgate (ex.: R$ ...)

**Corpo do e-mail:**

```
Nome do Clube: {nomeClube};
Código: {codigoClube};
Valor do resgate: {valor}.
```

**Frase de fechamento:** Aguardo confirmação para realizar o resgate.

---

## Ações, Futuros e FIIs

### Compra

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a compra

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Código do ativo)
- `{quantidade}` — Quantidade (ex.: xxxx)
- `{preco}` — Preço (ex.: Preço ou 'Ordem a mercado')
- `{validade}` — Validade da ordem (ex.: xx/xx/xxxx ou 'Até Cancelar' (só se não for a mercado))

**Corpo do e-mail:**

```
Ativo: {ativo};
Quantidade: {quantidade};
Preço: {preco};
Validade da ordem: {validade}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a compra.

---

### Venda

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a venda

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Código do ativo)
- `{quantidade}` — Quantidade (ex.: xxxx)
- `{preco}` — Preço (ex.: Preço ou 'Ordem a mercado')
- `{validade}` — Validade da ordem (ex.: xx/xx/xxxx ou 'Até Cancelar' (só se não for a mercado))

**Corpo do e-mail:**

```
Ativo: {ativo};
Quantidade: {quantidade};
Preço: {preco};
Validade da ordem: {validade}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a venda.

---

## Opções

### Compra

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a compra de opção

**Campos do formulário:**

- `{ativoRef}` — Ativo Referência (ex.: Código do ativo)
- `{opcao}` — Opção (ex.: Código da opção)
- `{tipo}` — Tipo (opções: Call, Put)
- `{strike}` — Preço de Exercício (ex.: Preço de strike)
- `{vencimento}` — Vencimento (ex.: xx/xx/xxxx)
- `{quantidade}` — Quantidade (ex.: xxxx)
- `{preco}` — Preço (ex.: Preço ou 'Ordem a mercado')
- `{validade}` — Validade da ordem (ex.: xx/xx/xxxx ou 'Até Cancelar')

**Corpo do e-mail:**

```
Ativo Referência: {ativoRef};
Opção: {opcao};
Tipo: {tipo};
Preço de Exercício: {strike};
Vencimento: {vencimento};
Quantidade: {quantidade};
Preço: {preco};
Validade da ordem: {validade}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a compra.

---

### Venda

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a venda de opção

**Campos do formulário:**

- `{ativoRef}` — Ativo Referência (ex.: Código do ativo)
- `{opcao}` — Opção (ex.: Código da opção)
- `{tipo}` — Tipo (opções: Call, Put)
- `{strike}` — Preço de Exercício (ex.: Preço de strike)
- `{vencimento}` — Vencimento (ex.: xx/xx/xxxx)
- `{quantidade}` — Quantidade (ex.: xxxx)
- `{preco}` — Preço (ex.: Preço ou 'Ordem a mercado')
- `{validade}` — Validade da ordem (ex.: xx/xx/xxxx ou 'Até Cancelar')

**Corpo do e-mail:**

```
Ativo Referência: {ativoRef};
Opção: {opcao};
Tipo: {tipo};
Preço de Exercício: {strike};
Vencimento: {vencimento};
Quantidade: {quantidade};
Preço: {preco};
Validade da ordem: {validade}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a venda.

---

## Operações a Termo

### Compra a Termo

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a compra a Termo

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Código do ativo)
- `{quantidade}` — Quantidade (ex.: xxxx)
- `{preco}` — Preço (ex.: R$ ...)
- `{taxa}` — Taxa (exata ou máxima aceita) (ex.: Taxa de juros exata ou taxa máxima aceita pelo cliente)
- `{vencimento}` — Vencimento (ex.: xx/xx/xxxx ou nº de dias)

**Corpo do e-mail:**

```
Ativo: {ativo};
Quantidade: {quantidade};
Preço: {preco};
Taxa: {taxa};
Vencimento: {vencimento}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a compra.

---

### Venda a Termo

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a venda a Termo

**Campos do formulário:**

- `{ativo}` — Ativo (ex.: Código do ativo)
- `{quantidade}` — Quantidade (ex.: xxxx)
- `{preco}` — Preço (ex.: R$ ...)
- `{taxa}` — Taxa (exata ou mínima aceita) (ex.: Taxa de juros exata ou taxa mínima aceita pelo cliente)
- `{vencimento}` — Vencimento (ex.: xx/xx/xxxx ou nº de dias)

**Corpo do e-mail:**

```
Ativo: {ativo};
Quantidade: {quantidade};
Preço: {preco};
Taxa: {taxa};
Vencimento: {vencimento}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a venda.

---

## Movimentação de Recursos

### Retirada de Recursos

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a retirada de recursos, com destino à conta abaixo, de acordo com seu cadastro,

**Campos do formulário:**

- `{banco}` — Banco (ex.: Nome ou número do banco)
- `{agencia}` — Agência (ex.: Número da agência)
- `{conta}` — Conta (ex.: Número da conta)
- `{valor}` — Valor a ser transferido (ex.: R$ ...)

**Corpo do e-mail:**

```
Banco: {banco};
Agência: {agencia};
Conta: {conta};
Valor a ser transferido: {valor};
Tipo de transferência: TED.
```

**Frase de fechamento:** Aguardo confirmação para realizar a transferência.

---

## Compromissadas

### Aplicação

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a aplicação em operação Compromissada (autorização do saldo disponível às 17h, pelo prazo de 180 dias)

**Campos do formulário:**

- `{taxaMinima}` — Taxa mínima de recompra (ex.: Ex: XX% CDI)
- `{valorLimite}` — Valor limite a ser aplicado (ex.: R$ ...)

**Corpo do e-mail:**

```
Operação: Compromissada XP Investimentos;
Taxa mínima de recompra: {taxaMinima};
Valor limite a ser aplicado: {valorLimite}.
```

**Frase de fechamento:** Aguardo confirmação para realizar a aplicação.

---

### Resgate

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** o resgate

**Campos do formulário:**

- `{dataRecompra}` — Data de Recompra (ex.: xx/xx/xxxx)
- `{valor}` — Valor a ser resgatado (ex.: R$ ...)

**Corpo do e-mail:**

```
Operação: Compromissada XP Investimentos;
Tipo: Revenda;
Data de Recompra: {dataRecompra};
Valor a ser resgatado: {valor}.
```

**Frase de fechamento:** Aguardo confirmação para realizar o resgate.

---

## Carteira Automatizada

### Nova Carteira

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** a entrada na Carteira Automatizada

**Campos do formulário:**

- `{casaAnalise}` — Casa de análise homologada (ex.: Empresa que recomendou a carteira)
- `{dataExecucao}` — Data da execução (ex.: xx/xx/xxxx)

**Corpo do e-mail:**

```
Casa de análise: {casaAnalise};
Data da execução: {dataExecucao} (na autorização recebida após o fechamento do mercado será processada no próximo dia útil, e somente se possível atender nas condições autorizadas);
Preço: "à mercado" - as execuções ocorrerão ao longo do dia (preço final será encaminhado em novo e-mail ao final do pregão);
Quantidades: execuções ao longo do dia (quantidades serão conhecidas a partir dos preços, e encaminhadas ao final do dia).
Abaixo a Carteira de ações com o indicativo dos percentuais recomendados pela casa de análise [incluir tabela com a carteira sugerida].
```

**Aviso de anexo:** Atenção assessor: é obrigatório o envio do PDF da carteira recomendada, tanto para entrada como para o rebalanceamento.

**Frase de fechamento:** Aguardo confirmação para fazer as compras indicadas acima.

---

### Troca e Rebalanceamento

**Frase de abertura (só usada quando é a única operação do e-mail, em Ações/Opções/Termo):** o rebalanceamento da Carteira Automatizada

**Campos do formulário:**

- `{casaAnalise}` — Casa de análise homologada (ex.: Empresa que recomendou a carteira)
- `{dataExecucao}` — Data da execução (ex.: xx/xx/xxxx)

**Corpo do e-mail:**

```
Casa de análise: {casaAnalise};
Data da execução: {dataExecucao} (a autorização recebida após o fechamento do mercado será processada no próximo dia útil, e somente se possível atender nas condições autorizadas);
Preço: "à mercado" - as execuções ocorrerão ao longo do dia (preço final será encaminhado em novo e-mail ao final do pregão);
Quantidades: execuções ao longo do dia (quantidades serão conhecidas a partir dos preços, e encaminhadas ao final do dia).
Abaixo a Carteira atual versus a carteira sugerida [incluir tabela indicando a carteira atual e a sugerida].
```

**Aviso de anexo:** Atenção assessor: é obrigatório o envio do PDF da carteira recomendada, tanto para entrada como para o rebalanceamento.

**Frase de fechamento:** Aguardo confirmação para fazer as compras indicadas acima.

---

## Texto do link "Preciso registrar uma Operação Estruturada"

```
Os modelos de ordem para Operações Estruturadas foram elaborados separadamente, devido à complexidade das operações, aos riscos envolvidos e às suas especificidades. Além disso, os modelos de operações estruturadas ofertadas estão dentro do cotizador e são enviados de forma automatizada ao cliente via push.

Para qualquer operação diferente das suportadas pelo cotizador, contate o time de Operações Estruturadas pelo HUB do assessor (menu Renda Variável > Produtos Estruturados), ou pelos e-mails controle.produtosestruturados@xpi.com.br / auditoriadeordens@xpi.com.br.

Cautela redobrada é recomendada no registro de ordem e na realização desse tipo de operação. Informe sempre o cliente de forma clara sobre toda a estrutura da operação e os riscos envolvidos.
```
