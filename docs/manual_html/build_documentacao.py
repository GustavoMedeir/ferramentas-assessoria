# -*- coding: utf-8 -*-
import sys, os
sys.path.insert(0, os.path.dirname(__file__))
from shared import *

VERSION = "2.2.0"
DATE = "Julho de 2026"
AUTHOR = "Gustavo Meireles — Assessoria de Investimentos XP"

pages = []

pages.append(cover(
    "Ferramentas de Assessoria", "Documentação Técnica",
    "Arquitetura, stack e decisões de implementação",
    VERSION, DATE, AUTHOR,
))

pages.append(toc([
    ("1", "Visão geral", "o que o app faz"),
    ("2", "Stack tecnológica", ""),
    ("3", "Estrutura do projeto", ""),
    ("4", "Fluxo principal — aba Rentabilidade", ""),
    ("5", "Decisões e armadilhas documentadas", ""),
    ("6", "Build e testes", ""),
    ("7", "Identidade visual", ""),
]))

# ---------------------------------------------------------------- 1
b = sec_open("1", "Visão geral")
b += p('Aplicativo desktop Windows que substitui a ferramenta antiga em Python — um conjunto de utilitários '
       'para assessores de investimento: leitura automatizada de relatórios XP, mensagens de rentabilidade, '
       'geração de e-mails de ordem, calculadoras financeiras e edição de PDF.')
b += table(["Aba", "O que faz"], [
    ["Rentabilidade", "Lê os relatórios XPerformance da pasta, monta a mensagem de rentabilidade (mês/ano/12 meses) por cliente."],
    ["E-mails de Ordem", "Monta o texto padronizado de ordem por produto (Renda Fixa, COE, Erro Operacional...)."],
    ["Tabela de Deságio", "Calcula e desenha a tabela de ágio/deságio de títulos como imagem."],
    ["Editor de PDF", "Anota (desenho, texto, tarja, imagem, assinatura) um PDF existente."],
    ["Imagens em PDF", "Junta N imagens escolhidas pelo usuário num único PDF novo."],
    ["Apresentação", "Modo slideshow de um PDF/apresentação carregada, tela cheia."],
    ["Typeform", "Formulário de coleta de dados do cliente durante a reunião."],
    ["Calculadora / Comparadora / Previdenciária / Compromissada / Aposentadoria", "Calculadoras financeiras independentes, sem dependência dos relatórios."],
])
b += shot("rentabilidade_previa_12m.jpg", "A aba Rentabilidade: lista de clientes à esquerda, mensagem pronta à direita (dados ilustrativos).")
pages.append(page(b))

# ---------------------------------------------------------------- 2
b = sec_open("2", "Stack tecnológica")
b += table(["Camada", "Tecnologia", "Por quê"], [
    ["Backend", "Go + Wails v2", "binário nativo único, sem runtime externo pro usuário instalar"],
    ["Frontend", "JavaScript vanilla (sem framework)", "app pequeno o bastante pra não precisar de build tooling extra"],
    ["Leitura de PDF", "go-pdfium (WASM via wazero)", "extrai texto/renderiza páginas sem depender de cgo"],
    ["Escrita de PDF", "github.com/go-pdf/fpdf", "monta PDF novo a partir de imagens (Editor de PDF e Imagens em PDF)"],
    ["Banco de dados", "SQLite (modernc.org/sqlite)", "um arquivo por pasta de relatórios, sem servidor, sem cgo"],
    ["Configuração", "JSON em %APPDATA%", "simples de inspecionar e fazer backup"],
])
b += box_note("Por que sem cgo?", 'cgo exige um compilador C configurado na máquina que builda o app — go-pdfium (WASM) e modernc.org/sqlite (Go puro) evitam essa dependência, então "go build" funciona em qualquer máquina com Go instalado, sem toolchain C.')
pages.append(page(b))

# ---------------------------------------------------------------- 3
b = sec_open("3", "Estrutura do projeto")
b += code_light([
    "rentabilidade/",
    "├── main.go                    # entrypoint, redireciona stdio, inicia o Wails",
    "├── app.go                     # métodos expostos ao frontend (bindings Wails)",
    "├── dto.go                     # structs \"wire\" (JSON-safe) expostas ao frontend",
    "├── config.go                  # struct de preferências + leitura/escrita do config.json",
    "├── internal/",
    "│   ├── pdfreport/              # extração de dados dos relatórios XP (regex sobre o texto do PDF)",
    "│   ├── rentabilidade/          # banco SQLite, placeholders da mensagem, exportação CSV",
    "│   ├── clientdb/               # leitura da base de clientes (CSV)",
    "│   ├── emailgen/                # catálogo de categorias, monta o texto de e-mail de ordem",
    "│   ├── pdfedit/                 # monta um PDF novo a partir de páginas/imagens (fpdf)",
    "│   ├── assinaturas/             # imagens de assinatura salvas (Configurações → Assinatura)",
    "│   └── whatsapp/                # monta o link wa.me a partir do telefone",
    "├── frontend/",
    "│   ├── src/",
    "│   │   ├── main.js              # monta a sidebar, o roteamento entre abas, Configurações",
    "│   │   ├── state.js             # estado em memória compartilhado entre as abas",
    "│   │   ├── tabs/                 # um módulo por aba (mount(container, ctx))",
    "│   │   └── ui/                   # componentes (el/btn/...), ícones SVG, theme.css",
    "│   └── wailsjs/                  # bindings gerados (não editar à mão)",
    "└── build/",
    "    ├── appicon.png               # ícone-fonte, 1024×1024",
    "    └── windows/icon.ico          # gerado a partir do appicon.png",
])
pages.append(page(b))

# ---------------------------------------------------------------- 4
b = sec_open("4", "Fluxo principal — aba Rentabilidade")
b += flow([
    ("1. Boot", "EstadoInicial() lê config + reabre última pasta", False),
    ("2. Varredura", "lista PDFs novos na pasta (glob)", False),
    ("3. Extração", "pdfreport: regex sobre o texto do PDF", False),
    ("4. Dedup", "por código de conta, fica o mais recente", True),
    ("5. Junta com a base", "código → nome/telefone do CSV", True),
    ("6. Ordena por patrimônio", "maior primeiro", False),
    ("7. Mensagem", "placeholders → texto final", False),
])
b += sub("4.1 Extração dos dados (internal/pdfreport)")
b += p('O texto extraído do PDF passa por uma normalização (PDFium devolve U+00A0 em vez de espaço comum em '
       'vários pontos do relatório — sem isso os regexes não batem).')
ul_list = "<ul class='plain'>" + "".join(f"<li>{x}</li>" for x in [
    "Código da conta é obrigatório — sem ele o arquivo vira uma falha de processamento.",
    "As linhas <b>MÊS</b>, <b>ANO</b> e <b>12M</b> da tabela \"RESUMO DE INFORMAÇÕES DA CARTEIRA\" são lidas "
    "pelo mesmo padrão de regex genérico (<code>padraoLinhaResumo(rotulo)</code>), só trocando o rótulo — as "
    "três são obrigatórias (arquivo sem alguma delas vira falha de processamento).",
    "Data de referência é opcional (usada pra desempate quando há PDFs duplicados do mesmo código).",
    "Patrimônio total bruto é opcional — usado só pra ordenar a lista de clientes.",
]) + "</ul>"
b += ul_list
b += sub("4.2 Deduplicação por código")
b += p('Quando duas leituras da pasta encontram arquivos diferentes com o mesmo código de conta (relatório '
       'baixado de novo, arquivo renomeado), só o de Data de Referência mais recente é devolvido pela lista de '
       'registros — o(s) mais antigo(s) fica(m) de fora por completo, inclusive da exportação CSV.')
b += sub("4.3 A lista de clientes (ListarClientes)")
b += table(["Situação", "Onde aparece"], [
    ["Código da base com relatório processado", "Linha normal, com os dados do relatório"],
    ["Código da base sem relatório (ou PDF com falha)", "Linha em branco, ordenada por nome"],
    ["Código de um relatório sem cliente correspondente na base", "Linha extra no fim, só com o código"],
])
b += sub("4.4 Mensagem final")
b += p(f'{chip("modelo_mensagem.txt")} (um por pasta) guarda o texto com os placeholders. '
       f'{chip("ValoresPlaceholder")} resolve cada um: {chip("_Rent")}, {chip("_RentA")}, {chip("_Rent12M")}, '
       f'{chip("_Perc")}, {chip("_PercA")}, {chip("_Perc12M")}, {chip("_CDI")}, {chip("_CDIA")}, '
       f'{chip("_CDI12M")}, {chip("_Nome")} — os valores em R$/% já vêm formatados em pt-BR pelo backend '
       f'({chip("pdfreport.FormatarReais")}/{chip("FormatarPercentual")}), fonte única de verdade reaproveitada '
       'também pelo frontend (RegistroDTO carrega os campos já formatados, pra não duplicar a lógica em JS). '
       '"Copiar mensagem"/"Enviar WhatsApp" usam esse texto final; o WhatsApp abre via link wa.me.')
pages.append(page(b))

# ---------------------------------------------------------------- 5
b = sec_open("5", "Decisões e armadilhas documentadas")
b += p('Registro do "porquê" por trás de escolhas não óbvias — pra não redescobrir o mesmo problema duas vezes.')
b += gcard("crit", "Redirecionamento de stdio é obrigatório",
    'O .exe é uma aplicação GUI (sem console) — nesse caso o Windows não dá ao processo um handle válido de '
    'stdout/stderr. Bibliotecas que escrevem direto em os.Stdout (go-pdfium, wazero) travam/erram sem isso. '
    f'{chip("main.go")} redireciona os.Stdout/os.Stderr pra um arquivo de log antes de qualquer outra coisa rodar.')
b += gcard("info", "Cache de compilação do WASM",
    f'Sem {chip("WithCompilationCache")}, o wazero recompila o binário WASM do PDFium do zero toda vez que o '
    'app abre (~3,4s medido, interface "travada" na abertura). Com cache persistente em disco, só a primeira '
    'execução em cada máquina paga esse custo.')
b += gcard("info", "PDFium inicializa sob demanda",
    'O motor de PDF só é criado quando a primeira operação que precisa dele acontece (varrer uma pasta, abrir '
    'o Editor de PDF) — abrir o app sem nenhuma pasta configurada não paga esse custo de inicialização.')
b += gcard("crit", "A ordem do texto extraído do PDF não é a ordem visual da página",
    'PDFium devolve os "runs" de texto na ordem em que foram desenhados no PDF, que raramente bate com a '
    'ordem de leitura visual (colunas, tabelas). Os regexes de extração dependem de rótulos únicos e formato '
    'fixo (ver 4.1) justamente por isso — não dá pra confiar em posição/coordenada.')
b += gcard("info", "Instância única do aplicativo",
    'Um segundo clique no .exe traz a janela existente pra frente em vez de abrir outra — evita duas cópias '
    'do programa disputando o mesmo arquivo SQLite ao mesmo tempo.')
b += gcard("info", "Migração de banco sem perda de dados",
    f'Colunas novas em {chip("rentabilidades.db")} (ex.: as 3 do 12M) entram via {chip("ALTER TABLE")} '
    'idempotente na abertura do banco — erro de "coluna já existe" é esperado e ignorado, então bancos '
    'antigos ganham a coluna nova sem perder o que já tinha.')
b += gcard("info", '"R$" automático nos campos de valor (emailgen.Field.Moeda)',
    f'Campos de e-mail marcados {chip("Moeda: true")} recebem o prefixo "R$ " automaticamente em '
    f'{chip("ResolveValores")} se o usuário não tiver digitado — só quando o campo não está vazio (não mexe no '
    'fallback "[Label]" nem em campo opcional vazio, e não duplica "R$" se o usuário já tiver digitado). Só '
    'existe pra campos que são exclusivamente um valor monetário — campos que aceitam quantidade ou texto '
    'livre (ex.: "Resgate Total") ficam de fora de propósito.')
b += gcard("info", "EmailCompleto agora recebe o código da conta",
    f'{chip("Category.EmailCompleto")} (usado por Resgate Prev e Erro Operacional — modelos com texto 100% '
    'próprio, sem a introdução/fechamento padrão) só recebia o nome do cliente. O modelo "Erro Operacional" '
    'precisa também do código da conta na linha "Cliente: código - nome", então a assinatura passou a ser '
    f'{chip("func(codigo, nome string, v map[string]string) string")} — e de quebra, usar '
    f'{chip("EmailCompleto")} já garante "uma ordem por e-mail" de graça (mesma trava que Resgate Prev usa), '
    'sem precisar de nenhuma flag nova no catálogo.')
pages.append(page(b))

# ---------------------------------------------------------------- 6
b = sec_open("6", "Build e testes")
b += code_dark([
    "cd rentabilidade",
    "wails build              # gera build/bin/FerramentasAssessoria.exe",
    "wails dev                # roda com hot-reload do frontend",
    "go test ./...             # testes de todos os pacotes Go",
])
b += p('Os testes de integração contra um PDF real (extração completa, incluindo 12M) são pulados por padrão '
       '— apontar a variável de ambiente roda o fluxo completo:')
b += code_dark([
    "PDFREPORT_TEST_PDF=<caminho do PDF> go test ./internal/... -v",
])
b += p(f'A versão em {chip("wails.json")} (campo "info") fica gravada nas propriedades do .exe — é o que '
       'aparece em Configurações → Sobre e no botão direito → Propriedades → Detalhes do Windows, o que também '
       'ajuda a passar confiança em máquinas onde o SmartScreen desconhece o binário.')
pages.append(page(b))

# ---------------------------------------------------------------- 7
b = sec_open("7", "Identidade visual")
b += p('O ícone é um "raiado" (sunburst): um leque de raios dourados convergindo para um ponto de luz, sobre '
       'um quadrado arredondado verde-pinho — remete tanto a um facho de luz guiando o assessor quanto ao '
       'gráfico de rentabilidade "para cima e para a direita".')
b += '<div style="text-align:center;margin:22px 0 8px;"><img src="images/appicon.png" style="width:160px;margin:0 auto;border-radius:36px;box-shadow:0 16px 40px rgba(13,46,38,.18);"></div>'
b += '<div class="cap">Ícone em 1024×1024 — a mesma imagem-base de todos os tamanhos usados no Windows.</div>'
b += p('Todos os tamanhos (16 a 1024px, incluindo o .ico multi-resolução do Windows e a versão de 160px da '
       'barra lateral) são gerados programaticamente a partir dessa única imagem-fonte, com supersampling — '
       'sem depender de nenhuma ferramenta de design externa.')
b += "<ul class='plain'>" + "".join(f"<li>{esc(x)}</li>" for x in [
    "build/appicon.png — fonte única, 1024×1024",
    "build/windows/icon.ico — multi-resolução, gerado a partir do appicon.png",
    "Versão de 160px — usada no cabeçalho expandido da barra lateral",
]) + "</ul>"
pages.append(page(b, closing={"plain": True, "text": f"Ferramentas de Assessoria — Documentação Técnica — v{VERSION}"}))

html = doc("Ferramentas de Assessoria — Documentação Técnica", pages)
outpath = os.path.join(os.path.dirname(__file__), "DOCUMENTACAO.html")
with open(outpath, "w", encoding="utf-8") as f:
    f.write(html)
print("wrote", outpath, len(html), "bytes,", len(pages), "pages")
