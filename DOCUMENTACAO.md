# Documentação Técnica — Ferramentas de Assessoria

> Como o app funciona por dentro e como ele foi construído.
> Para o manual de uso (configuração e funcionalidades), veja [MANUAL.md](MANUAL.md).

## Visão geral

Aplicativo desktop para Windows de apoio à assessoria de investimentos XP. Reúne, numa
única janela, ferramentas que antes viviam em planilhas e num app Python/tkinter:

| Aba | O que faz |
|---|---|
| Rentabilidade | Lê relatórios XPerformance (PDF), extrai os dados de rentabilidade e monta mensagens de WhatsApp por cliente |
| E-mails de Ordem | Gera e-mails padronizados de ordem (31 categorias de operação) |
| Tabela de Deságio | Compara valor atual × valor de saída de títulos e exporta imagem |
| Calculadora | 9 cartões de cálculo financeiro + Calculadora de Renda Fixa |
| Comparadora | Dois títulos de renda fixa lado a lado, veredito do melhor líquido |
| Calculadora Previdenciária | Simplificada × Completa × Completa com 12% em PGBL |
| Compromissada | Conta Remunerada × CDB × Compromissada dia a dia (calendário ANBIMA) |
| Planejamento de Aposentadoria | Aporte mensal necessário para a renda desejada |

## Stack

- **Go 1.x + Wails v2** — backend em Go, interface em HTML/CSS/JS rodando numa janela
  WebView2 nativa do Windows. Escolhido para permitir redesenhar a UI sem tocar no backend.
- **Frontend vanilla JS** (sem framework) — módulos ES, build via Vite. Helpers próprios de
  DOM em `frontend/src/ui/components.js`.
- **PDF**: [go-pdfium](https://github.com/klippa-app/go-pdfium) — PDFium compilado para
  WASM, executado via **wazero**. Zero cgo: não precisa de compilador C nem DLL externa.
- **Banco**: SQLite via `modernc.org/sqlite` (Go puro, sem cgo). Um arquivo
  `rentabilidades.db` por pasta de relatórios.
- **Config**: JSON em `%APPDATA%\RentabilidadeXP\config.json` (mesmo caminho e chaves do
  app Python original, para migração transparente).
- **Log**: `%LOCALAPPDATA%\RentabilidadeXP\app.log` (stdout/stderr do processo são
  redirecionados para lá no boot — ver abaixo por quê).

## Estrutura do projeto

```
rentabilidade/
├── main.go            # bootstrap Wails, janela, instância única, redirecionamento de stdio
├── app.go             # struct App: estado + todos os bindings expostos ao frontend
├── dto.go             # structs "wire" (JSON) trocadas com o frontend
├── config.go          # load/save do config.json
├── internal/
│   ├── pdfreport/     # extração de dados dos PDFs XPerformance (PDFium + regex)
│   ├── rentabilidade/ # banco SQLite, varredura de pasta, dedup, mensagem, CSV
│   ├── clientdb/      # parser do CSV da base de clientes
│   ├── emailgen/      # catálogo de 31 categorias + montagem dos e-mails de ordem
│   ├── selfupdate/    # checagem/aplicação de atualização via releases do GitHub
│   └── whatsapp/      # montagem do link wa.me
├── frontend/
│   ├── src/main.js    # boot, sidebar, navegação, configurações
│   ├── src/state.js   # estado singleton compartilhado entre as abas
│   ├── src/tabs/      # um módulo por aba
│   ├── src/ui/        # components.js (helpers DOM), icons.js, theme.css
│   └── src/util/      # matemática financeira, feriados ANBIMA, números pt-BR
│   └── wailsjs/       # bindings gerados pelo Wails (não editar à mão)
└── build/
    ├── appicon.png    # ícone 1024 (identidade "F raiada"/sunburst)
    └── windows/icon.ico  # ícone multi-tamanho embutido no .exe
```

## Fluxo principal (aba Rentabilidade)

1. **Boot**: o frontend chama `EstadoInicial()` — recupera pasta/base salvas no config e
   devolve preferências. Se há pasta salva, chama `ProcessarPastaAtual()` em seguida.
2. **Varredura** (`internal/rentabilidade.ProcessarPasta`): lista `*.pdf` da pasta, pula os
   já processados (presentes no banco **com** código preenchido) e extrai os dados dos
   novos via PDFium. Progresso é emitido pelo evento `processamento:progresso`.
3. **Extração** (`internal/pdfreport`): o texto de todas as páginas é concatenado e
   normalizado (PDFium devolve espaços U+00A0 — viram espaço comum) e regexes capturam:
   - **Código da conta** — campo "Conta" da capa (obrigatório; sem ele o arquivo vira falha);
   - Linhas MÊS/ANO do "Resumo de Informações da Carteira" (ganho R$, rentabilidade %, %CDI);
   - Data de referência (rodapé);
   - **Patrimônio total bruto** (opcional — usado só para ordenar a lista).
4. **Dedup**: se dois arquivos têm o mesmo código de conta, só o de data de referência mais
   recente sobrevive (`deduplicarPorCodigo`) — o antigo fica fora da lista **e** do CSV.
5. **Lista de clientes** (`ListarClientes`): junta a base de clientes (CSV código→nome) com
   os registros processados. Todo cliente da base vira uma linha — quem não tem PDF (ou o
   PDF falhou) aparece em branco. Ordem: com relatório por patrimônio decrescente → sem
   relatório por nome → PDFs de código desconhecido por código.
6. **Mensagem**: template com placeholders (`_Rent`, `_RentA`, `_Perc`, `_PercA`, `_CDI`,
   `_CDIA`, `_Nome`) salvo em `modelo_mensagem.txt` na pasta dos PDFs. "Copiar mensagem"
   substitui os valores e marca o registro como COPIADO; "Enviar WhatsApp" abre o wa.me
   com a mensagem preenchida (o telefone vem da base de clientes).

## Decisões e armadilhas documentadas

Estas custaram debugging real — não "simplificar" sem entender:

- **`redirecionarStdioParaArquivo()` (main.go) é essencial**: app GUI no Windows recebe
  handles de stdout/stderr inválidos; wazero/go-pdfium acessam `os.Stdout` diretamente e
  quebram sem isso ("The handle is invalid").
- **Cache de compilação do WASM** (`pdfreport.compilationCacheDir`): sem ele, o wazero
  recompila o PDFium a cada abertura (~3,4s de janela travada). Com o cache em
  `%LOCALAPPDATA%\RentabilidadeXP\pdfium-wasm-cache`, só a primeira execução paga o custo.
- **PDFium inicializa sob demanda** (`App.garantirExtractor` + `sync.Once`), não no boot —
  quem só usa o gerador de e-mails não paga o custo do motor de PDF.
- **Espaços U+00A0 no texto extraído**: o `\s` do RE2 do Go não casa com eles; a
  normalização em `ExtrairTexto` é obrigatória antes de qualquer regex.
- **A ordem do texto extraído ≠ ordem visual do PDF**: sempre inspecionar o texto real
  (via `ExtrairTexto`) antes de escrever um regex novo — não confie no layout do PDF.
- **`el()` e eventos**: `el("button", {onClick: ...})` registra o listener com
  `k.slice(2).toLowerCase()` — qualquer handler `onX` novo depende desse lowercase.
- **Instância única** (`SingleInstanceLock` em main.go): duas instâncias disputariam o
  mesmo SQLite ("database is locked" + UI desatualizada). Abrir o .exe de novo só traz a
  janela existente para frente.
- **Migração de banco**: colunas novas entram via `colunasMigracao` (ALTER TABLE que falha
  em silêncio se a coluna existe). Linhas antigas sem `codigo` são reprocessadas
  automaticamente na varredura seguinte (upsert preserva o status COPIADO).
- **Literal BOM em fonte Go**: usar o escape `﻿`, nunca o caractere literal.
- **Placeholders**: a lista vai do mais longo pro mais curto (`_RentA` antes de `_Rent`) —
  ordem errada corrompe a substituição. Duplicada conscientemente em Go e JS
  (`components.js`), mudou num lado → replicar no outro.

## Build

```bash
cd rentabilidade
wails build          # gera build/bin/FerramentasAssessoria.exe
wails dev            # desenvolvimento com hot reload
go test ./...        # testes (pdfreport, rentabilidade, clientdb, emailgen, whatsapp, selfupdate)
```

Testes de integração com PDF real: `PDFREPORT_TEST_PDF=<caminho do pdf> go test ./internal/... -v`.

Metadados de versão do .exe (empresa, produto, versão, copyright) vêm do bloco `info` do
`wails.json` — manter atualizado a cada release; além de aparecer nas propriedades do
arquivo, reduz a chance de falso positivo de antivírus (executável "anônimo" é mais
suspeito para a heurística).

## Atualização automática (internal/selfupdate)

O app checa, em background, se há uma release mais nova publicada num repositório público
do GitHub (`internal/selfupdate.Repositorio` — **hoje um placeholder, `SEU_USUARIO/...`,
precisa ser trocado pelo repo real antes da primeira release com esse mecanismo**). Achou
versão mais nova: aparece um item "Atualização disponível" na barra lateral; ao confirmar,
o app baixa o novo `.exe`, confere o checksum SHA256, substitui o binário em uso
(`github.com/minio/selfupdate`, cuida de "arquivo em uso" no Windows e faz rollback
automático se algo falhar no meio) e reabre sozinho. Tudo dentro do app — o assessor nunca
precisa ir ao GitHub manualmente.

**Passo a passo de cada release:**

1. Bump da versão em `wails.json` (`info.productVersion`).
2. `wails build -ldflags "-X main.Version=X.Y.Z"` — **sem o `-ldflags`, o binário fica com
   `Version = "dev"` e a checagem de atualização é pulada silenciosamente** (ver
   `main.go`). O valor tem que bater com a tag da release, sem o prefixo `v` (a tag em si
   leva `v`, ex. tag `v2.2.0` com `-X main.Version=2.2.0` — `selfupdate.MaisNova` ignora um
   `v` opcional dos dois lados na comparação).
3. Gerar o checksum do `.exe` recém-buildado (`certutil -hashfile caminho\FerramentasAssessoria.exe SHA256`
   no Windows, ou `sha256sum` se disponível) e salvar como `<nome-do-exe>.sha256`.
4. `git tag vX.Y.Z && git push --tags`.
5. `gh release create vX.Y.Z build/bin/FerramentasAssessoria.exe build/bin/FerramentasAssessoria.exe.sha256 --notes "..."`
   — as notas da release são o que aparece pro assessor como "o que mudou" no app.

**Armadilha ao trocar de mecanismo**: uma versão publicada *sem* esse código embutido nunca
vai se auto-atualizar (ela não tem a lógica de checagem). A primeira release com
self-update ainda precisa ser distribuída manualmente pra quem já usa o app — a partir
dela, tudo é automático.

## Identidade visual

Ícone "F raiada (sunburst)": leque de 9 raios dourados convergindo num ponto de luz, sobre
verde-petróleo escuro com cantos arredondados. Gerado programaticamente (Go, sem
dependências) em todos os tamanhos — o gerador renderiza cada tamanho nativamente (16 a
1024 px, com supersampling) em vez de reduzir uma imagem grande, para os raios finos não
virarem ruído nos ícones pequenos. Saídas: `build/appicon.png` (1024) e
`build/windows/icon.ico` (256/128/64/48/40/32/24/16), embutidos no .exe pelo `wails build`.
