# Ferramentas de Assessoria

Aplicativo desktop para Windows que reúne, numa única janela, 13 ferramentas de apoio à
rotina de assessoria de investimentos: leitura de relatórios de rentabilidade, geração de
e-mails de ordem, calculadoras financeiras e utilitários de PDF.

Tudo roda **localmente na máquina do usuário** — nenhum dado de cliente é enviado a servidor
nenhum.

> **Aviso**: software independente, desenvolvido em caráter pessoal e sem fins comerciais.
> **Não** foi desenvolvido, homologado, mantido ou divulgado pelo time de tecnologia da XP
> Investimentos. O uso é por conta e risco do usuário.

---

## Instalação

1. Baixe o `.exe` mais recente em **[Releases](../../releases/latest)**.
2. Copie para onde preferir (ex.: `C:\Ferramentas`) e crie um atalho.

Não tem instalador nem dependências: é um arquivo único. Na primeira execução o Windows
SmartScreen pode alertar "aplicativo não reconhecido" (o app não tem assinatura digital
paga) — basta **"Mais informações" → "Executar assim mesmo"**, uma vez por máquina.

**Requisitos**: Windows 10 ou 11. O WebView2 (usado para a interface) já vem com o Windows
atualizado.

## Atualizações automáticas

O app se atualiza sozinho: alguns segundos depois de abrir, ele verifica em segundo plano se
há uma release mais nova aqui no repositório. Havendo, aparece **"Atualização disponível"** na
barra lateral com as notas da versão; ao confirmar, o app baixa o novo executável, valida o
checksum SHA-256, se substitui e reabre — sem instalador, sem permissão de administrador e
sem perder configurações. Se algo falhar no meio, ele continua na versão atual (a troca só
acontece depois do download ser validado).

Também é possível checar na hora em **Configurações → Sobre**, onde fica a versão instalada.

> O mecanismo existe a partir da **v1.00.03**. Quem estiver com um build anterior precisa
> baixar o `.exe` manualmente uma última vez — daí em diante as atualizações são automáticas.

## As ferramentas

| Ferramenta | O que faz |
|---|---|
| **Rentabilidade** | Lê os relatórios XPerformance (PDF) de uma pasta, extrai rentabilidade do mês/ano/12 meses e monta a mensagem de cada cliente, com envio por WhatsApp |
| **E-mails de Ordem** | Gera e-mails padronizados de ordem para 16 produtos, com abertura de rascunho direto no Outlook |
| **Tabela de Deságio** | Compara valor atual × valor de saída dos títulos e exporta a tabela como imagem |
| **Apresentação** | Exibe a apresentação institucional (PDF ou HTML) em modo slides, sem barras de ferramentas na tela |
| **Typeform** | Preenche o diagnóstico financeiro durante a reunião e transporta as respostas para o formulário online |
| **Calculadora** | 9 cartões de cálculo financeiro (metas, projeções, conversão de taxas) mais a Calculadora de Renda Fixa |
| **Comparadora de Renda Fixa** | Dois títulos lado a lado, com veredito de qual rende mais líquido |
| **Calculadora Previdenciária** | Simplificada × Completa × Completa com 12% em PGBL, com todas as deduções e limites |
| **Compromissada** | Conta Remunerada × CDB × Compromissada dia útil a dia útil, considerando IOF regressivo |
| **Planejamento de Aposentadoria** | Aporte mensal necessário em dois cenários: consumo de patrimônio e viver de renda |
| **Editor de PDF** | Texto, desenho, tarjas, imagens e assinatura direto sobre o PDF |
| **Imagens em PDF** | Junta várias imagens num único PDF, na ordem escolhida |
| **Validar Assinatura** | Confere assinaturas digitais ICP-Brasil (PAdES e CAdES), offline |

A barra lateral é personalizável: dá para reordenar e esconder ferramentas. O **Modo
apresentação** oculta de uma vez as abas de uso interno e os dados sensíveis (spread,
comissão), deixando a tela segura para compartilhar com o cliente.

## Como funciona por dentro

**Interface** em HTML/CSS/JS (vanilla, sem framework), rodando numa janela WebView2 nativa;
**backend em Go**, compilado num único executável via [Wails v2](https://wails.io). Sem cgo
— não precisa de compilador C nem de DLL externa.

| Área | Como é resolvido |
|---|---|
| Leitura de PDF | [go-pdfium](https://github.com/klippa-app/go-pdfium) — PDFium compilado para WASM e executado via [wazero](https://wazero.io) |
| Banco de dados | SQLite em Go puro (`modernc.org/sqlite`), um arquivo por pasta de relatórios |
| Assinaturas digitais | Validação própria de CAdES/PAdES, com a cadeia ICP-Brasil embutida e atualizável |
| Rascunhos de e-mail | Automação COM do Outlook Classic — preserva a assinatura configurada pelo usuário e nunca envia sozinho |
| Preenchimento do Typeform | Automação do Microsoft Edge via [chromedp](https://github.com/chromedp/chromedp), lendo cada tela ao vivo |
| Atualização automática | Releases do GitHub + [minio/selfupdate](https://github.com/minio/selfupdate), com verificação de checksum |

**Onde ficam os dados**: preferências e assinaturas em `%APPDATA%\RentabilidadeXP`; banco,
modelos de mensagem e Typeforms salvos ficam na própria pasta de relatórios escolhida pelo
usuário; log em `%LOCALAPPDATA%\RentabilidadeXP\app.log`.

As duas únicas conexões de rede que o app faz são para **buscar atualizações do próprio app**
e para **baixar a cadeia de certificados públicos da ICP-Brasil**.

## Desenvolvimento

```bash
wails dev            # desenvolvimento com hot reload
go test ./...        # testes
wails build -ldflags "-X main.Version=X.YY.ZZ"   # build de produção
```

O `-ldflags` é obrigatório num build de release: sem ele o binário fica com
`Version = "dev"` e a checagem de atualização é pulada silenciosamente.

**Versionamento** `MAJOR.MINOR.PATCH` (ex.: `1.01.00`): o 3º número sobe em correção de bug,
o 2º em funcionalidade nova (zerando o 3º) e o 1º em mudança grande (zerando os dois).

---

© 2026 Gustavo De Medeiros — uso interno, sem distribuição comercial.
