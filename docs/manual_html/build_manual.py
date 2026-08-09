# -*- coding: utf-8 -*-
import sys, os
sys.path.insert(0, os.path.dirname(__file__))
from shared import *

VERSION = "2.2.0"
DATE = "Julho de 2026"
AUTHOR = "Gustavo Meireles — Assessoria de Investimentos XP"

pages = []

# Cover
pages.append(cover(
    "Ferramentas de Assessoria", "Manual do Usuário",
    "Como configurar o app e usar cada funcionalidade",
    VERSION, DATE, AUTHOR,
))

# TOC
pages.append(toc([
    ("01", "Primeiros passos", "instalar, base de clientes, pasta de relatórios"),
    ("02", "Aba Rentabilidade", "lista de clientes, mensagem, modelo (agora com 12 meses), WhatsApp, imagem do gráfico"),
    ("03", "Aba E-mails de Ordem", "produtos, operações, Erro Operacional, R$ automático, modos padronizado e livre"),
    ("04", "Tabela de Deságio", "agora com ordenação por nome/posição/deságio e linha de soma"),
    ("05", "Editor de PDF", "texto, desenho, tarjas, imagens e assinatura direto no PDF"),
    ("06", "Imagens em PDF", "novo — junte várias imagens num único PDF, na ordem que quiser"),
    ("07", "Calculadora", "os 9 cartões + Calculadora de Renda Fixa"),
    ("08", "Comparadora de Renda Fixa", ""),
    ("09", "Calculadora Previdenciária", ""),
    ("10", "Compromissada", ""),
    ("11", "Planejamento de Aposentadoria", ""),
    ("12", "Configurações", "cada opção explicada"),
    ("13", "Problemas comuns e onde ficam os dados", ""),
]))

# ---------------------------------------------------------------- 01
b = sec_open("01", "Primeiros passos")
b += sub("1.1 Instalar e abrir")
b += p(f'O app é um único arquivo, {chip("FerramentasAssessoria.exe")} — não tem instalador nem dependências. '
       f'Copie para onde preferir (por exemplo, {chip("C:\\\\Ferramentas")}) e crie um atalho na área de '
       'trabalho, ou clique com o botão direito no ícone da barra de tarefas com o app aberto e escolha '
       '"Fixar na barra de tarefas".')
b += p('O app trabalha com <b>instância única</b>: se você der dois cliques no .exe com ele já aberto, '
       'nenhuma segunda janela abre — a que já existe é trazida para a frente.')
b += box_note("Atenção · primeira execução",
    '"Aplicativo não reconhecido" na primeira execução. Por ser um programa próprio, sem assinatura digital '
    'paga, o Windows SmartScreen pode alertar na primeira vez. Clique em "Mais informações" → "Executar assim '
    'mesmo". Acontece só uma vez por máquina.')
b += sub("1.2 Carregar a base de clientes")
b += p('A base de clientes é o coração da integração entre as abas: é ela quem dá nome aos códigos de conta '
       'extraídos dos relatórios, preenche automaticamente o nome no gerador de e-mails e fornece o telefone '
       'usado pelo botão "Enviar WhatsApp".')
b += steps([
    'Monte uma planilha com três colunas: <b>código da conta XP</b>, <b>nome do cliente</b> e '
    '<b>telefone</b> (opcional, mas necessário para o WhatsApp).',
    'Salve como CSV (no Excel: Arquivo → Salvar como → tipo "CSV").',
    'No app, clique em "Base de clientes" (ícone de pessoas, no rodapé da barra lateral) e selecione o arquivo.',
])
b += box_example("Exemplo do conteúdo do arquivo", code_example([
    "codigo;nome;telefone",
    "100001;CLIENTE EXEMPLO 01;11999990001",
    "100002;CLIENTE EXEMPLO 02;11999990002",
]))
b += p('O separador (<code>;</code> ou <code>,</code>) é detectado automaticamente, os nomes das colunas podem '
       'variar e vir em qualquer ordem, e o telefone aceita qualquer formato — sem DDI, o app assume Brasil (+55). '
       'O caminho fica salvo: nas próximas aberturas a base carrega sozinha.')
b += sub("1.3 Escolher a pasta de relatórios")
b += p('Na aba Rentabilidade, clique em "Escolher pasta com os PDFs" e aponte para a pasta onde você baixa os '
       'relatórios XPerformance.')
b += steps([
    'O app varre só os PDFs novos da pasta (o código da conta é lido de dentro do arquivo — o nome do arquivo não importa).',
    'Extrai automaticamente código, ganhos, rentabilidade, %CDI, data de referência e patrimônio.',
    f'Guarda tudo num banquinho ({chip("rentabilidades.db")}) dentro da própria pasta — reabrir o app não reprocessa o que já foi lido.',
])
pages.append(page(b))

# ---------------------------------------------------------------- 02
b = sec_open("02", "Aba Rentabilidade")
b += shot("rentabilidade_previa_12m.jpg",
    "Prévia da mensagem: além do mês e do ano, agora também mostra os últimos 12 meses.")
b += sub("2.1 A lista de clientes")
b += p('Todos os clientes da base aparecem, ordenados por patrimônio (maior primeiro). Cliente sem relatório '
       'na pasta (ou cujo PDF falhou na leitura) aparece em branco — assim você enxerga rapidamente quem falta. '
       'A busca filtra por nome ou código.')
b += sub("2.2 Prévia e ações")
b += table(["Ação", "O que faz"], [
    ["Copiar mensagem", "Copia o texto pronto e marca o cliente como COPIADO."],
    ["Enviar WhatsApp", "Abre o WhatsApp com a mensagem já preenchida — você revisa e aperta enviar lá."],
    ["Copiar imagem", "Copia o gráfico de rentabilidade recortado do PDF (área ajustável em Configurações → Recorte da imagem)."],
])
b += sub("2.3 O modelo da mensagem — agora com 12 meses")
b += p('A aba "Modelo" mostra a legenda com os placeholders clicáveis: clique num selo para inserir no texto. '
       'Desde esta versão, além das colunas <b>Mês</b> e <b>Ano</b>, existe uma terceira coluna <b>12M</b> — os '
       'últimos 12 meses corridos, lidos direto da linha "12M" do relatório XP (a mesma tabela de onde já vinham '
       'Mês e Ano).')
b += shot("rentabilidade_modelo_12m.jpg", "Legenda de placeholders com a nova coluna 12M ao lado de Mês e Ano.")
b += table(["Placeholder", "Substituído por"], [
    [chip("_Rent"), "Ganho em R$ no mês"],
    [chip("_RentA"), "Ganho em R$ no ano"],
    [chip("_Rent12M"), "Ganho em R$ nos últimos 12 meses"],
    [chip("_Perc"), "Rentabilidade % no mês"],
    [chip("_PercA"), "Rentabilidade % no ano"],
    [chip("_Perc12M"), "Rentabilidade % nos últimos 12 meses"],
    [chip("_CDI") + " / " + chip("_CDIA") + " / " + chip("_CDI12M"), "% do CDI no mês / ano / 12 meses"],
    [chip("_Nome"), "Nome do cliente"],
])
b += p('Modelos já personalizados continuam funcionando exatamente como antes — o placeholder de 12 meses só fica '
       'disponível para quem quiser inserir. Clique em "Salvar modelo" para gravar as mudanças (fica em '
       f'{chip("modelo_mensagem.txt")}, dentro da pasta de relatórios).')
pages.append(page(b))

# ---------------------------------------------------------------- 03
b = sec_open("03", "Aba E-mails de Ordem")
b += shot("emails_erro_operacional.jpg",
    "Modelo \u201cErro Operacional\u201d: campos próprios, sem tipo de movimentação, e limitado a uma ordem por e-mail.")
b += sub("3.1 Identificando o cliente e o produto")
b += p('Digite o código do cliente — o nome preenche sozinho se estiver na base. Escolha produto e tipo de '
       'operação; a lista cobre Renda Fixa, Tesouro Direto, Fundos, COE, Ofertas, Subscrição, Societárias, '
       'Clubes, Ações/Opções/Termo, Movimentação, Compromissadas, Carteira Automatizada, Resgate Prev e, desde '
       'esta versão, <b>Erro Operacional</b>.')
b += box_note("Novo · Erro Operacional",
    'Modelo para reportar um erro operacional interno (não é uma ordem de cliente): pede código do erro, '
    'assessor, responsável pelo erro, datas do incidente e de identificação, descrição e valor. Como não tem '
    '"tipo de movimentação" (só existe uma opção), e como compliance exige aprovação individual por incidente, '
    'esse modelo aceita <b>só uma ocorrência por e-mail</b> — o botão "Adicionar operação" fica desabilitado '
    'automaticamente ao escolher esse produto.')
b += sub("3.2 Valores em R$ — prefixo automático")
b += p('Nos campos que só aceitam um valor monetário (Valor a ser aplicado, Preço, Valor limite etc.), agora '
       'não é preciso digitar "R$" — o app coloca sozinho na hora de montar o texto do e-mail. Basta digitar o '
       'número, por exemplo <code>1.000,00</code>. Campos que aceitam outras opções além de um valor em reais '
       '(como "Quantidade" ou "Resgate Total") continuam como texto livre, sem prefixo automático.')
b += sub("3.3 Várias operações e geração do e-mail")
b += p('"+ Adicionar operação" inclui mais de uma no mesmo e-mail (mesmo produto/tipo, regra de compliance do '
       'modo padronizado). Em Configurações dá para alternar para o modo <b>livre</b>, que permite misturar '
       'produtos diferentes num mesmo e-mail (comportamento antigo). "Gerar e-mail" monta o texto padronizado; '
       'revise e use "Copiar texto".')
pages.append(page(b))

# ---------------------------------------------------------------- 04
b = sec_open("04", "Tabela de Deságio")
b += shot("desagio_ordem_soma.jpg",
    "Ordenação por Nome, Tamanho da posição ou Ágio/Deságio, e a linha de total opcional.")
b += p('Adicione linhas com nome do título, valor atual e valor de saída — o deságio em R$ e % calcula sozinho, '
       'com prévia da tabela pronta para virar imagem ("Copiar imagem" ou "Salvar imagem").')
b += sub("4.1 Ordenar por outros critérios")
b += p('Um seletor no cabeçalho da tabela escolhe o critério de ordenação:')
b += table(["Critério", "Ordena por"], [
    ["Nome", "Ordem alfabética (A→Z por padrão)"],
    ["Tamanho da posição", "Valor Atual, do maior para o menor por padrão"],
    ["Ágio/Deságio", "% de ágio/deságio, do maior ágio para o maior deságio por padrão (comportamento original)"],
])
b += p('O botão ao lado inverte a direção a qualquer momento (ex.: "Ágio → Deságio" vira "Deságio → Ágio"). '
       'Linhas sem valor completo sempre vão para o final, em qualquer critério.')
b += sub("4.2 Somar os valores")
b += p('O botão "Mostrar soma" acrescenta uma linha de total em negrito, embaixo da tabela, com a soma do Valor '
       'Atual, do Valor de Saída e do Deságio em R$ de todos os títulos — e o deságio % do total (ponderado pelo '
       'valor de cada título, não a média simples das porcentagens individuais). A linha de total entra na '
       'imagem exportada normalmente.')
pages.append(page(b))

# ---------------------------------------------------------------- 05 (NOVO)
b = sec_open("05", "Editor de PDF")
b += p('Ferramenta para anotar um PDF (ex.: um relatório ou contrato) diretamente na tela: desenhar, escrever, '
       'tarjar informação, colar imagens ou sua assinatura salva — sem precisar de outro programa.')
b += shot("editor_pdf_caneta_preta.jpg",
    "Ferramenta Caneta com a cor padrão preta, navegação de páginas e botão Desfazer na barra de ferramentas.")
b += sub("5.1 Ferramentas")
b += table(["Ferramenta", "O que faz"], [
    ["Caneta", "Desenho livre — cor e espessura ajustáveis na barra de ferramentas."],
    ["Texto", "Caixa de texto arrastável; clique fora ou no ✓ para aplicar na página."],
    ["Tarja branca", "Cobre uma área retangular com branco sólido, para ocultar informação."],
    ["Imagem", "Cola uma imagem do computador, redimensionável antes de aplicar."],
    ["Assinatura", "Cola direto a assinatura salva em Configurações → Assinatura."],
    ["Marca X", "Um \u201c✕\u201d preto fixo, do tamanho certo pra marcar um campo, no clique."],
])
b += sub("5.2 O que mudou nesta versão")
b += box_note("Cor padrão agora é preta",
    'A caneta e o texto nasciam em vermelho; agora começam em preto, mais discreto pra a maioria dos usos '
    '(o seletor de cor continua livre para trocar quando quiser).')
b += box_note("Zoom não reseta mais ao trocar de página",
    'Antes, dar zoom numa página e ir pra próxima voltava o zoom ao padrão. Agora o zoom escolhido permanece '
    'ao navegar entre as páginas de um mesmo PDF — só reajusta automaticamente na primeira página de um PDF novo.')
b += box_note("Desfazer com Ctrl+Z",
    'Além do botão "Desfazer" na barra de ferramentas, o atalho Ctrl+Z (ou Cmd+Z) agora desfaz o último traço, '
    'texto ou imagem — funciona sempre que a aba do Editor de PDF estiver visível.')
b += p('Ao terminar, "Salvar como..." monta um PDF novo com todas as edições já achatadas nas páginas — o '
       'arquivo original nunca é alterado.')
pages.append(page(b))

# ---------------------------------------------------------------- 06 (NOVO)
b = sec_open("06", "Imagens em PDF")
b += p('Ferramenta nova: transforma quantas imagens você quiser (fotos, comprovantes, prints, documentos '
       'escaneados) em um único arquivo PDF, uma imagem por página, na ordem que você escolher.')
b += shot("imgpdf_vazio.jpg", "Tela inicial da aba Imagens em PDF, antes de adicionar qualquer imagem.")
b += sub("6.1 Como usar")
b += steps([
    'Clique em "Adicionar imagens" e escolha uma ou mais fotos/imagens do computador. Pode clicar de novo pra '
    'ir somando mais — cada seleção soma no fim da lista, não substitui a anterior.',
    'Reordene arrastando o item pela alcinha (⣿) à esquerda, ou pelos botões ↑/↓ à direita de cada linha.',
    'Remova uma imagem individual com o ícone de lixeira, se precisar tirar alguma do lote.',
    'Clique em "Gerar PDF" — o app monta o arquivo (uma página por imagem, na ordem final da lista) e abre o '
    'diálogo para você escolher onde salvar.',
])
b += shot("imgpdf_lista.jpg", "Imagens adicionadas em lotes diferentes, prontas para reordenar e virar um PDF.")
b += box_note("Imagens com transparência (PNG)",
    'Uma imagem com fundo transparente ganha fundo branco automaticamente ao virar página de PDF — igual ao '
    'Editor de PDF, PDF não tem canal de transparência.')
pages.append(page(b))

# ---------------------------------------------------------------- 07 Calculadora
b = sec_open("07", "Calculadora")
b += p('Nove cartões agrupados em três categorias, mais a Calculadora de Renda Fixa — tudo calcula em tempo '
       'real conforme os campos são preenchidos, sem precisar apertar nenhum botão "calcular". "Limpar" reseta '
       'todos os cartões de uma vez.')
b += sub("7.1 Planejamento de metas")
b += table(["Cartão", "Pergunta que responde"], [
    ["Tempo para meta", "Quanto tempo até atingir um valor objetivo, dado o aporte mensal e a rentabilidade."],
    ["Aporte necessário", "Quanto aplicar por mês para atingir a meta no prazo desejado."],
    ["Rentabilidade necessária", "Que rentabilidade buscar para atingir a meta com o aporte e prazo dados."],
])
b += sub("7.2 Projeção de valores")
b += table(["Cartão", "Pergunta que responde"], [
    ["Rentabilidade futura", "Quanto um valor renderá até uma data futura, numa taxa dada."],
    ["Rendimento futuro", "Quanto de rendimento (R$) um valor gera até uma data futura."],
    ["Cobertura do FGC", "Se um valor está dentro do limite garantido pelo FGC."],
])
b += sub("7.3 Conversão de taxas e Calculadora de Renda Fixa")
b += p('Cálculo da Taxa Real (desconta a inflação), Composição de Taxas e Transformação de Taxas (mensal ↔ '
       'anual etc.) completam o grupo de conversões. A Calculadora de Renda Fixa monta um título completo '
       '(emissor, tipo CDB/LCI/LCA, datas pelo calendário ANBIMA, taxa contratada, valor aplicado) e destaca o '
       'valor líquido no vencimento, com o IR e o passo a passo do cálculo disponíveis num "Ver detalhes".')
pages.append(page(b))

# ---------------------------------------------------------------- 08 Comparadora
b = sec_open("08", "Comparadora de Renda Fixa")
b += p('Compara dois títulos lado a lado (pré, pós % CDI, IPCA+, isentos ou não), mostrando qual rende mais '
       'líquido e destacando o vencedor métrica a métrica: alíquota de IR de cada um, valor líquido e '
       'rentabilidade líquida a.a./a.m.')
b += box_note("Calendário ANBIMA", 'Os prazos usam o calendário de dias úteis ANBIMA — a mesma referência usada pela mesa de operações.')
pages.append(page(b))

# ---------------------------------------------------------------- 09 Previdenciária
b = sec_open("09", "Calculadora Previdenciária")
b += p('Compara declaração Simplificada × Completa × Completa com 12% em PGBL a partir da renda tributável e '
       'deduções informadas. A tabela do IR usada (2022 ou 2026) é escolhida em Configurações → Tabela '
       'previdenciária.')
pages.append(page(b))

# ---------------------------------------------------------------- 10 Compromissada
b = sec_open("10", "Compromissada")
b += p('Compara Conta Remunerada × CDB × Compromissada dia a dia entre duas datas, usando o calendário de dias '
       'úteis ANBIMA. O resultado tem um gráfico de rendimento por dia útil, o dia em que o CDB passa a valer '
       'mais a pena (considerando o IOF regressivo dos primeiros 30 dias) e uma tabela diária completa.')
b += box_note("Visão cliente × visão assessor",
    'O Modo apresentação (Configurações) controla o que aparece na Compromissada: <b>desligado</b>, mostra o '
    'SPREAD da plataforma e a comissão do escritório (visão completa, uso interno); <b>ligado</b>, esconde os '
    'dois — seguro para compartilhar a tela com o cliente. Não existe mais uma escolha separada de "Visão"; é '
    'sempre o mesmo interruptor do Modo apresentação (ver seção 12.5).')
pages.append(page(b))

# ---------------------------------------------------------------- 11 Aposentadoria
b = sec_open("11", "Planejamento de Aposentadoria")
b += p('A partir da idade atual, idade desejada, expectativa de vida, renda mensal pretendida, aplicações já '
       'feitas e premissas de rentabilidade, calcula o aporte mensal necessário. O resultado compara dois '
       'cenários — <b>Consumo de Patrimônio</b> (com sucessão) e <b>Viver de Renda</b> (preservando o '
       'patrimônio) — cada um com seus próprios valores de aporte e renda projetada.')
pages.append(page(b))

# ---------------------------------------------------------------- 12 Configurações
b = sec_open("12", "Configurações")
b += p('A janela de Configurações (engrenagem na barra lateral) tem dez abas.')
b += sub("12.1 E-mail")
b += p('Escolhe o modo do gerador de e-mails: <b>padronizado</b> (recomendado — um produto por e-mail, modelo '
       'específico, regra de compliance) ou <b>livre</b> (mistura produtos, comportamento antigo).')
b += sub("12.2 Temas")
b += p('Tema claro/escuro, cor de destaque (Esmeralda, Teal, Aqua, Verde petróleo...) e fonte da interface.')
b += sub("12.3 Tabela previdenciária")
b += p('Escolhe a tabela de IR (2022 ou 2026) usada pela Calculadora Previdenciária.')
b += sub("12.4 Rentabilidade")
b += p('Liga o Modo Festas — troca a mensagem padrão por um modelo sazonal (ex.: Natal) para todos os clientes '
       'da base, com ou sem relatório processado.')
b += sub("12.5 Modo apresentação")
b += p('Um único interruptor que substitui as antigas abas separadas "Visão" e "Modo apresentação".')
b += shot("config_modo_apresentacao.jpg", "Configurações → Modo apresentação: um único Desligado/Ligado, sem escolha separada de Visão.")
b += table(["Opção", "O que faz"], [
    ["Desligado", "Tudo normal — todas as abas aparecem na sidebar, com SPREAD e comissão visíveis na Compromissada."],
    ["Ligado", "Esconde Rentabilidade, E-mails de Ordem, Tabela de Deságio, Editor de PDF, Imagens em PDF e Base de "
               "clientes da sidebar; some com as notificações de status/falha; esconde o SPREAD e a comissão do "
               "escritório; muda a Compromissada pra visão cliente."],
])
b += sub("12.6 Recorte da imagem")
b += p('Ajuste fino da área do gráfico copiada pelo "Copiar imagem" da aba Rentabilidade, com prévia sobre um '
       'relatório real. "Restaurar padrão" volta ao recorte original.')
b += sub("12.7 Assinatura")
b += p('Cadastra uma ou mais imagens de assinatura; a marcada como ativa é a que o botão "Assinatura" do Editor '
       'de PDF cola na página.')
b += sub("12.8 Ordem da barra lateral")
b += p('Arraste para reordenar as ferramentas na barra lateral, ou use o botão "-" para esconder alguma.')
b += sub("12.9 Sobre")
b += p('Versão instalada e informações de build.')
b += sub("12.10 Arquivos com falha")
b += p('Lista os PDFs que não puderam ser lidos na última varredura, com o motivo exato de cada falha.')
pages.append(page(b))

# ---------------------------------------------------------------- 13 Problemas comuns
b = sec_open("13", "Problemas comuns e onde ficam os dados")
b += table(["Sintoma", "Causa / solução"], [
    ['"Aplicativo não reconhecido" ao abrir', 'SmartScreen para apps sem assinatura — "Mais informações" → "Executar assim mesmo" (só na 1ª vez)'],
    ["Cliente em branco na lista", "Não há PDF daquela conta na pasta, ou o PDF veio quebrado da XP (confira em Configurações → Arquivos com falha)"],
    ["Primeira abertura numa máquina demora alguns segundos", "O motor de PDF compila na primeira vez e fica em cache — as próximas são rápidas"],
    ['"Enviar WhatsApp" reclama de telefone', "O cliente não tem telefone na base de clientes — adicione a coluna e recarregue a base"],
    ["Não consigo exportar o CSV", "O arquivo está aberto no Excel — feche e exporte de novo"],
    ["Quero recomeçar do zero numa pasta", 'Botão "Limpar" (ou apague o rentabilidades.db da pasta)'],
])
b += sub("Onde ficam os dados")
b += table(["O quê", "Onde"], [
    ["Preferências", chip("%APPDATA%\\RentabilidadeXP\\config.json")],
    ["Dados extraídos dos relatórios", chip("rentabilidades.db") + " (dentro da pasta de PDFs)"],
    ["Modelo de mensagem", chip("modelo_mensagem.txt") + " (dentro da pasta de PDFs)"],
    ["Log", chip("%LOCALAPPDATA%\\RentabilidadeXP\\app.log")],
])
b += box_note("Backup", 'Fazer backup é simples: copie o config.json e a pasta de relatórios (com o .db e o .txt juntos) — não há nenhum outro lugar com dado do app.')
pages.append(page(b, closing={"text": f"FERRAMENTAS DE ASSESSORIA — MANUAL DO USUÁRIO — V{VERSION}"}))

html = doc("Ferramentas de Assessoria — Manual do Usuário", pages)
outpath = os.path.join(os.path.dirname(__file__), "MANUAL.html")
with open(outpath, "w", encoding="utf-8") as f:
    f.write(html)
print("wrote", outpath, len(html), "bytes,", len(pages), "pages")
