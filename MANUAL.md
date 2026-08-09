# Manual do Usuário — Ferramentas de Assessoria

> Como configurar o app e usar cada funcionalidade.
> Para a documentação técnica (arquitetura e build), veja [DOCUMENTACAO.md](DOCUMENTACAO.md).

## Primeiros passos

### 1. Instalar / abrir

O app é um único arquivo: `FerramentasAssessoria.exe`. Não tem instalador — basta copiar
para onde preferir (ex.: `C:\Ferramentas`) e criar um atalho na área de trabalho ou barra
de tarefas. Abrir o .exe uma segunda vez não cria outra janela: ele só traz a que já está
aberta para a frente.

> **Aviso do Windows Defender/SmartScreen na primeira execução**: por ser um programa novo
> e sem assinatura digital paga, o Windows pode mostrar "aplicativo não reconhecido".
> Clique em **"Mais informações" → "Executar assim mesmo"**. Isso acontece só na primeira
> vez em cada máquina. O executável carrega os dados do autor e versão nas propriedades do
> arquivo (botão direito → Propriedades → Detalhes).

### 2. Carregar a base de clientes

Botão **"Base de clientes"** no rodapé da barra lateral. Selecione um CSV com as colunas:

```
codigo;nome;telefone
3419078;ALDETE LOURDES FAVERO;27999990000
```

- Separador `;` ou `,` — detectado automaticamente.
- Os nomes das colunas podem variar ("Código", "Conta", "Nome", "Telefone", "Celular",
  "WhatsApp"...) e em qualquer ordem — o app reconhece pelo cabeçalho.
- A coluna de telefone é opcional; sem ela o botão "Enviar WhatsApp" não funciona para o
  cliente (o resto funciona normalmente).
- Telefone pode estar em qualquer formato — só os dígitos importam. Sem DDI, o app assume
  Brasil (+55).

O caminho fica salvo: nas próximas aberturas a base carrega sozinha.

### 3. Escolher a pasta de relatórios

Na aba **Rentabilidade**, clique em "Escolher pasta com os PDFs" e aponte para a pasta
onde você baixa os relatórios XPerformance. O app:

- lê cada PDF novo e extrai os dados automaticamente (o código da conta é lido de dentro
  do arquivo — o nome do arquivo não importa);
- guarda o resultado num banquinho (`rentabilidades.db`) dentro da própria pasta —
  reabrir o app não reprocessa o que já foi lido;
- cria um `modelo_mensagem.txt` com o modelo padrão da mensagem.

A pasta também fica salva para as próximas aberturas.

## Aba Rentabilidade

**Lista de clientes (esquerda)** — todos os clientes da base, ordenados por patrimônio
(maior primeiro). Cada linha mostra nome, código, o ganho do mês e o selo GERADO/COPIADO:

- Cliente **sem relatório na pasta** (ou cujo PDF falhou na leitura) aparece **em branco**
  (um traço no lugar do valor) — assim você enxerga rapidamente quem falta.
- PDF cujo código não está na base aparece no fim da lista, só com o código.
- Se houver dois PDFs para a mesma conta, vale o de data de referência mais recente.
- A busca filtra por nome ou código.

**Painel direito** — duas visões:

- **Prévia**: a mensagem montada para o cliente selecionado, com os valores destacados.
  - **Copiar mensagem** — copia o texto pronto e marca o cliente como COPIADO.
  - **Enviar WhatsApp** — abre o WhatsApp com a mensagem já preenchida na conversa do
    cliente (você ainda revisa e aperta enviar lá).
  - **Copiar imagem** — copia o gráfico de rentabilidade recortado do PDF (área
    configurável em Configurações → Recorte da imagem).
- **Modelo**: edite o texto da mensagem usando os placeholders (clique num selo para
  inserir): `_Rent`/`_RentA` (ganho R$ mês/ano), `_Perc`/`_PercA` (rentabilidade %),
  `_CDI`/`_CDIA` (% do CDI), `_Nome`. **Salvar modelo** grava para as próximas vezes.

**Barra superior** — Atualizar (relê a pasta), Exportar planilha (CSV que o Excel abre
direto), Limpar (apaga o banco da pasta; os PDFs ficam e são relidos do zero).

Arquivos que falharem na leitura aparecem em **Configurações → Arquivos com falha**, com o
motivo. (Acontece com PDFs corrompidos ou relatórios que a XP gerou sem os dados.)

## Aba E-mails de Ordem

1. Digite o **código do cliente** — o nome preenche sozinho se estiver na base.
2. Escolha **produto** e **tipo de operação** (Renda Fixa, Tesouro Direto, Fundos, COE,
   Ofertas, Subscrição, Societárias, Clubes, Ações/Opções/Termo, Movimentação,
   Compromissadas, Carteira Automatizada...).
3. Preencha os campos da operação. **+ Adicionar operação** inclui mais de uma no mesmo
   e-mail (mesmo produto/tipo, regra de compliance).
4. **Gerar e-mail** monta o texto padronizado; revise/edite e **Copiar texto**.

Em Configurações dá para alternar entre modo **padronizado** (um produto por e-mail, modelo
específico) e **livre** (mistura produtos — modelo antigo).

## Tabela de Deságio

Adicione linhas com **nome do título**, **valor atual** e **valor de saída** — o deságio em
R$ e % calcula sozinho, com prévia da tabela pronta. **Copiar imagem** ou **Salvar imagem**
para mandar ao cliente.

## Calculadora

Cartões agrupados em três categorias + Calculadora de Renda Fixa:

- **Planejamento de metas**: quanto tempo para atingir um valor; quanto aplicar por mês;
  qual rentabilidade buscar.
- **Projeção de valores**: rentabilidade futura; rendimento futuro; cobertura do FGC.
- **Conversão de taxas**: taxa real (desconto da inflação); composição de taxas;
  transformação de taxas (mensal ↔ anual etc.).
- **Calculadora de Renda Fixa**: um título completo (indexador, taxa, prazo, IR) com o
  valor líquido no vencimento em destaque e os detalhes do cálculo num toggle.

## Comparadora

Dois títulos de renda fixa lado a lado (pré, pós % CDI, IPCA+, isentos ou não), com
veredito de qual rende mais líquido e a tabela métrica a métrica destacando o vencedor.

## Calculadora Previdenciária

Compara declaração **Simplificada × Completa × Completa com 12% em PGBL** a partir da renda
tributável e deduções. A tabela do IR usada (2022 ou 2026) é escolhida em Configurações.

## Compromissada

Compara **Conta Remunerada × CDB × Compromissada** dia a dia entre duas datas, usando o
calendário de dias úteis ANBIMA. A **Visão** (Configurações) alterna entre a tela completa
(assessor) e uma versão enxuta para compartilhar com o cliente.

## Planejamento de Aposentadoria

A partir da idade atual, idade desejada, renda mensal pretendida e premissas de taxa,
calcula o aporte mensal necessário para sustentar a renda na aposentadoria.

## Configurações (engrenagem na barra lateral)

- **Tema** claro/escuro e **cor de destaque** (esmeralda, teal, aqua, petróleo...).
- **Fonte** da interface.
- **Modo do e-mail de ordem**: padronizado (recomendado) ou livre.
- **Tabela previdenciária**: 2022 ou 2026.
- **Visão** (Compromissada): cliente ou assessor.
- **Modo apresentação**: esconde Rentabilidade, E-mails, Deságio, Base de clientes e as
  notificações — seguro para compartilhar a tela com o cliente.
- **Recorte da imagem**: ajuste fino da área do gráfico copiada pelo "Copiar imagem" da
  aba Rentabilidade, com prévia sobre um relatório real.
- **Arquivos com falha**: lista dos PDFs que não puderam ser lidos na última varredura.

## Problemas comuns

| Sintoma | Causa/solução |
|---|---|
| "Aplicativo não reconhecido" ao abrir | SmartScreen para apps sem assinatura — "Mais informações" → "Executar assim mesmo" (só na 1ª vez) |
| Cliente em branco na lista | Não há PDF daquela conta na pasta, ou o PDF veio quebrado da XP (confira em Configurações → Arquivos com falha) |
| Primeira abertura numa máquina demora alguns segundos ao processar | O motor de PDF compila na primeira vez e fica em cache — as próximas são rápidas |
| "Enviar WhatsApp" reclama de telefone | O cliente não tem telefone na base de clientes — adicione a coluna e recarregue a base |
| Não consigo exportar o CSV | O arquivo está aberto no Excel — feche e exporte de novo |
| Quero recomeçar do zero numa pasta | Botão "Limpar" (ou apague o `rentabilidades.db` da pasta) |
| Onde ficam meus dados? | Config: `%APPDATA%\RentabilidadeXP\config.json` · Banco: `rentabilidades.db` na pasta dos PDFs · Log: `%LOCALAPPDATA%\RentabilidadeXP\app.log` |
