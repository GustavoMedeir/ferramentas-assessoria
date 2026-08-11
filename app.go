package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"rentabilidade/internal/assinaturas"
	"rentabilidade/internal/clientdb"
	"rentabilidade/internal/emailgen"
	"rentabilidade/internal/icpbrasil"
	"rentabilidade/internal/outlookmail"
	"rentabilidade/internal/pdfedit"
	"rentabilidade/internal/pdfreport"
	"rentabilidade/internal/rentabilidade"
	"rentabilidade/internal/selfupdate"
	"rentabilidade/internal/sigvalidator"
	"rentabilidade/internal/typeformbot"
	"rentabilidade/internal/whatsapp"
)

// App mantém o estado do app entre chamadas do frontend: a pasta de PDFs
// escolhida, o banco SQLite aberto para ela, o motor de PDF (PDFium, vivo
// entre chamadas) e a base de clientes carregada.
type App struct {
	ctx context.Context

	// O motor de PDF (PDFium) é caro pra inicializar (alguns segundos na
	// 1ª vez que roda numa máquina — depois fica cacheado, ver
	// pdfreport.NewExtractor) e só é usado pela aba Rentabilidade. Em vez
	// de pagar esse custo sempre na abertura do app — o que travava a UI
	// logo de cara, mesmo pra quem só ia usar o Gerador de E-mails —
	// inicializamos sob demanda, na primeira chamada a
	// ProcessarPastaAtual. extractorOnce garante que isso acontece uma
	// única vez.
	extractorOnce sync.Once
	extractor     *pdfreport.Extractor
	extractorErr  error

	pasta            string
	modeloPath       string
	modeloFestasPath string
	db               *sql.DB
	clientDB         map[string]string // codigo -> nome (contrato já exposto ao frontend)
	telefones        map[string]string // codigo -> telefone (só usado por EnviarWhatsApp/EnviarWhatsAppFestas)
	emails           map[string]string // codigo -> e-mail (exposto ao frontend, diferente de telefones — usado pra pré-preencher o destinatário em AbrirEmailNoOutlook)

	processando atomic.Bool // trava reentrância de ProcessarPastaAtual

	icpStore *icpbrasil.Store // cadeia de confiança ICP-Brasil (validação de assinatura)

	// atualizacaoDisponivel guarda a release encontrada pela última checagem
	// (automática no startup ou manual via VerificarAtualizacao), pra
	// AplicarAtualizacao não depender do frontend devolver as URLs de volta.
	// nil = nenhuma atualização pendente.
	atualizacaoDisponivel atomic.Pointer[selfupdate.ReleaseInfo]
}

// NewApp cria a struct do app.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Carregar do disco/snapshot embutido é rápido (sem rede) — só a
	// checagem de atualização é que roda em background, sem travar o
	// startup da UI (mesmo cuidado de lazy-init do garantirExtractor acima,
	// aplicado aqui a uma inicialização que já é barata por si só).
	store, err := icpbrasil.NovoStore(pastaICPBrasil())
	if err != nil {
		log.Println("icpbrasil: não foi possível inicializar a cadeia local:", err)
		return
	}
	a.icpStore = store
	a.icpStore.IniciarChecagemPeriodica(ctx)

	// Checagem de atualização: roda uma vez, em background, alguns segundos
	// depois do startup (dá tempo da UI e do carregamento de config
	// terminarem primeiro — não é urgente). Version só é "dev" num build
	// local sem o ldflag de release (ver comentário em main.go); nesse caso
	// não há release correspondente pra comparar, então nem tenta.
	if Version != "dev" {
		go func() {
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return
			}
			if dto := a.checarAtualizacao(ctx); dto.Disponivel {
				runtime.EventsEmit(a.ctx, "atualizacao:disponivel", dto)
			}
		}()

		// A checagem acima roda só uma vez, ao abrir — não ajuda quem deixa o
		// app aberto o dia inteiro sem reiniciar (comum: assessor abre de
		// manhã e só fecha à noite). Esta outra roda em paralelo e repete
		// todo dia às horaChecagemDiaria, enquanto o app continuar aberto.
		go a.loopChecagemDiaria(ctx)
	}
}

// horaChecagemDiaria é a hora (0-23, horário local da máquina) da checagem
// automática recorrente — ver loopChecagemDiaria.
const horaChecagemDiaria = 18

// loopChecagemDiaria verifica atualização todo dia às horaChecagemDiaria,
// enquanto o app estiver aberto. Recalcula a duração até a próxima checagem
// a cada volta (em vez de um time.Ticker de 24h fixas) pra não acumular
// deriva em mudança de horário de verão ou fuso.
func (a *App) loopChecagemDiaria(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(duracaoAteProximaChecagem(time.Now())):
		}
		if dto := a.checarAtualizacao(ctx); dto.Disponivel {
			runtime.EventsEmit(a.ctx, "atualizacao:disponivel", dto)
		}
	}
}

// duracaoAteProximaChecagem calcula quanto falta até horaChecagemDiaria:00
// (horário local) — hoje, se esse horário ainda não passou; amanhã, caso
// contrário (inclusive se `agora` for exatamente esse horário, pra não
// disparar duas vezes seguidas no instante exato).
func duracaoAteProximaChecagem(agora time.Time) time.Duration {
	proxima := time.Date(agora.Year(), agora.Month(), agora.Day(), horaChecagemDiaria, 0, 0, 0, agora.Location())
	if !proxima.After(agora) {
		proxima = proxima.AddDate(0, 0, 1)
	}
	return proxima.Sub(agora)
}

func (a *App) shutdown(_ context.Context) {
	if a.db != nil {
		a.db.Close()
	}
	a.extractor.Close()
}

// aoAbrirSegundaInstancia é chamado (na instância original) quando o usuário
// tenta abrir o .exe de novo — só traz a janela existente pra frente, em vez
// de deixar duas instâncias disputarem o mesmo banco SQLite.
func (a *App) aoAbrirSegundaInstancia(_ options.SecondInstanceData) {
	runtime.WindowUnminimise(a.ctx)
	runtime.Show(a.ctx)
}

// garantirExtractor inicializa o motor de PDF sob demanda (na primeira
// chamada de qualquer binding que precise dele — ver comentário em
// extractorOnce). Idempotente: chamadas depois da primeira só reusam o
// motor já pronto.
func (a *App) garantirExtractor() error {
	a.extractorOnce.Do(func() {
		a.extractor, a.extractorErr = pdfreport.NewExtractor()
	})
	if a.extractor == nil {
		return fmt.Errorf("motor de leitura de PDF não inicializado: %w", a.extractorErr)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Rentabilidade
// ---------------------------------------------------------------------------

// EstadoInicial é chamado uma vez no boot do frontend: recupera a pasta e a
// base de clientes salvas em config.json (se ainda existirem em disco) sem
// processar PDFs ainda — isso fica a cargo de ProcessarPastaAtual, chamado
// pelo frontend logo em seguida.
func (a *App) EstadoInicial() (InicioDTO, error) {
	cfg := carregarConfig()
	dto := InicioDTO{
		ClientDB:     map[string]string{},
		ClientEmails: map[string]string{},
		Prefs: PreferenciasDTO{
			Tema:                 cfg.Tema,
			Acento:               cfg.Acento,
			ModoEmail:            cfg.ModoEmail,
			TabelaPrevidenciaria: cfg.TabelaPrevidenciaria,
			Visao:                cfg.Visao,
			Fonte:                cfg.Fonte,
			ModoApresentacao:     cfg.ModoApresentacao,
			ModoFestas:           cfg.ModoFestas,
			EmailRemetente:       cfg.EmailRemetente,
			AssessorNome:         cfg.AssessorNome,
			AssessorEmail:        cfg.AssessorEmail,
			OrdemNav:             cfg.OrdemNav,
			OrdemNavOcultos:      cfg.OrdemNavOcultos,
			TemApresentacao:      cfg.Apresentacao != "",
			RecortePersonalizado: cfg.RecortePersonalizado,
			RecorteX0:            cfg.RecorteX0,
			RecorteY0:            cfg.RecorteY0,
			RecorteX1:            cfg.RecorteX1,
			RecorteY1:            cfg.RecorteY1,
			RecortePadraoX0:      pdfreport.RecorteGraficoRentabilidadePadrao.X0,
			RecortePadraoY0:      pdfreport.RecorteGraficoRentabilidadePadrao.Y0,
			RecortePadraoX1:      pdfreport.RecorteGraficoRentabilidadePadrao.X1,
			RecortePadraoY1:      pdfreport.RecorteGraficoRentabilidadePadrao.Y1,
		},
	}

	if cfg.Pasta != "" {
		if info, err := os.Stat(cfg.Pasta); err == nil && info.IsDir() {
			if err := a.definirPasta(cfg.Pasta); err != nil {
				return InicioDTO{}, err
			}
			modelo, err := a.CarregarModelo()
			if err != nil {
				return InicioDTO{}, err
			}
			modeloFestas, err := a.CarregarModeloFestas()
			if err != nil {
				return InicioDTO{}, err
			}
			dto.TemPasta = true
			dto.Pasta = a.pasta
			dto.Modelo = modelo
			dto.ModeloFestas = modeloFestas
		}
	}

	if cfg.BaseClientes != "" {
		if _, err := os.Stat(cfg.BaseClientes); err == nil {
			if base, err := clientdb.CarregarBaseClientes(cfg.BaseClientes); err == nil {
				a.clientDB, a.telefones, a.emails = separarNomesTelefonesEmails(base)
				dto.ClientDB = a.clientDB
				dto.ClientEmails = a.emails
			}
		}
	}

	return dto, nil
}

// separarNomesTelefonesEmails quebra {codigo: Cliente} em três mapas
// simples — o frontend conhece o de nomes e o de e-mails (usado pra
// pré-preencher o destinatário em AbrirEmailNoOutlook); o de telefones
// fica só no backend, usado por EnviarWhatsApp.
func separarNomesTelefonesEmails(base map[string]clientdb.Cliente) (nomes, telefones, emails map[string]string) {
	nomes = make(map[string]string, len(base))
	telefones = make(map[string]string, len(base))
	emails = make(map[string]string, len(base))
	for codigo, c := range base {
		nomes[codigo] = c.Nome
		if c.Telefone != "" {
			telefones[codigo] = c.Telefone
		}
		if c.Email != "" {
			emails[codigo] = c.Email
		}
	}
	return nomes, telefones, emails
}

// definirPasta abre (ou cria) o banco da pasta e garante que
// modelo_mensagem.txt e modelo_festas.txt existam.
func (a *App) definirPasta(pasta string) error {
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}

	db, err := rentabilidade.PrepararBanco(filepath.Join(pasta, "rentabilidades.db"))
	if err != nil {
		return fmt.Errorf("preparar banco: %w", err)
	}

	modeloPath := filepath.Join(pasta, "modelo_mensagem.txt")
	if _, err := os.Stat(modeloPath); os.IsNotExist(err) {
		if err := os.WriteFile(modeloPath, []byte(rentabilidade.ModeloPadrao), 0644); err != nil {
			db.Close()
			return fmt.Errorf("criar modelo_mensagem.txt: %w", err)
		}
	}

	modeloFestasPath := filepath.Join(pasta, "modelo_festas.txt")
	if _, err := os.Stat(modeloFestasPath); os.IsNotExist(err) {
		if err := os.WriteFile(modeloFestasPath, []byte(rentabilidade.ModeloFestasPadrao), 0644); err != nil {
			db.Close()
			return fmt.Errorf("criar modelo_festas.txt: %w", err)
		}
	}

	a.pasta = pasta
	a.modeloPath = modeloPath
	a.modeloFestasPath = modeloFestasPath
	a.db = db
	return nil
}

// EscolherPasta abre o diálogo nativo de seleção de pasta. Usado tanto para
// o primeiro uso (estado vazio na tela) quanto pelo botão "Trocar pasta".
func (a *App) EscolherPasta() (PastaDTO, error) {
	escolhida, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecione a pasta com os PDFs",
	})
	if err != nil {
		return PastaDTO{}, err
	}
	if escolhida == "" {
		return PastaDTO{}, nil // usuário cancelou o diálogo
	}

	if err := a.definirPasta(escolhida); err != nil {
		return PastaDTO{}, err
	}

	cfg := carregarConfig()
	cfg.Pasta = escolhida
	if err := salvarConfig(cfg); err != nil {
		return PastaDTO{}, err
	}

	modelo, err := a.CarregarModelo()
	if err != nil {
		return PastaDTO{}, err
	}
	modeloFestas, err := a.CarregarModeloFestas()
	if err != nil {
		return PastaDTO{}, err
	}
	return PastaDTO{Pasta: a.pasta, Modelo: modelo, ModeloFestas: modeloFestas}, nil
}

// ProcessarPastaAtual varre a pasta ativa por PDFs novos, emitindo o evento
// "processamento:progresso" a cada arquivo, e devolve a lista atualizada de
// registros.
func (a *App) ProcessarPastaAtual() (ProcessamentoDTO, error) {
	if a.pasta == "" || a.db == nil {
		return ProcessamentoDTO{}, fmt.Errorf("nenhuma pasta selecionada")
	}

	if a.extractor == nil {
		runtime.EventsEmit(a.ctx, "processamento:progresso", 0, 0) // sinaliza "preparando" enquanto o motor de PDF inicia (só na 1ª chamada)
	}
	if err := a.garantirExtractor(); err != nil {
		return ProcessamentoDTO{}, err
	}
	if !a.processando.CompareAndSwap(false, true) {
		return ProcessamentoDTO{}, fmt.Errorf("já há um processamento em andamento")
	}
	defer a.processando.Store(false)

	sucesso, falhas, err := rentabilidade.ProcessarPasta(a.pasta, a.db, a.extractor, func(feitos, total int) {
		runtime.EventsEmit(a.ctx, "processamento:progresso", feitos, total)
	})
	if err != nil {
		return ProcessamentoDTO{}, err
	}

	clientes, err := a.clientesDTO()
	if err != nil {
		return ProcessamentoDTO{}, err
	}

	return ProcessamentoDTO{
		Sucesso:  sucesso,
		Falhas:   paraFalhasDTO(falhas),
		Clientes: clientes,
	}, nil
}

// clientesDTO monta a lista de clientes da aba Rentabilidade (base de
// clientes + registros já processados, ver rentabilidade.ListarClientes),
// já convertida pro formato "wire". Reaproveitado por toda operação que
// muda a lista (processar pasta, copiar mensagem, enviar WhatsApp, limpar
// tudo).
func (a *App) clientesDTO() ([]ClienteRentabilidadeDTO, error) {
	registros, err := rentabilidade.ListarRegistros(a.db)
	if err != nil {
		return nil, err
	}
	festasEnviados, err := rentabilidade.ListarFestasEnviados(a.db)
	if err != nil {
		return nil, err
	}
	return paraClientesDTO(rentabilidade.ListarClientes(registros, a.clientDB, festasEnviados)), nil
}

// CarregarModelo lê o modelo de mensagem da pasta ativa.
func (a *App) CarregarModelo() (string, error) {
	if a.modeloPath == "" {
		return rentabilidade.ModeloPadrao, nil
	}
	dados, err := os.ReadFile(a.modeloPath)
	if os.IsNotExist(err) {
		return rentabilidade.ModeloPadrao, nil
	}
	if err != nil {
		return "", err
	}
	return string(dados), nil
}

// SalvarModelo grava o modelo de mensagem editado.
func (a *App) SalvarModelo(texto string) error {
	if a.modeloPath == "" {
		return fmt.Errorf("nenhuma pasta selecionada")
	}
	return os.WriteFile(a.modeloPath, []byte(texto), 0644)
}

// CarregarModeloFestas lê o modelo de mensagem de festas da pasta ativa
// (Modo Festas, ver Configurações).
func (a *App) CarregarModeloFestas() (string, error) {
	if a.modeloFestasPath == "" {
		return rentabilidade.ModeloFestasPadrao, nil
	}
	dados, err := os.ReadFile(a.modeloFestasPath)
	if os.IsNotExist(err) {
		return rentabilidade.ModeloFestasPadrao, nil
	}
	if err != nil {
		return "", err
	}
	return string(dados), nil
}

// SalvarModeloFestas grava o modelo de mensagem de festas editado.
func (a *App) SalvarModeloFestas(texto string) error {
	if a.modeloFestasPath == "" {
		return fmt.Errorf("nenhuma pasta selecionada")
	}
	return os.WriteFile(a.modeloFestasPath, []byte(texto), 0644)
}

// CopiarMensagem monta a mensagem final (template + valores do registro),
// copia pra área de transferência e marca o registro como copiado.
func (a *App) CopiarMensagem(arquivo, template string) (ProcessamentoDTO, error) {
	if a.db == nil {
		return ProcessamentoDTO{}, fmt.Errorf("nenhuma pasta selecionada")
	}

	registros, err := rentabilidade.ListarRegistros(a.db)
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	var alvo *rentabilidade.Registro
	for i := range registros {
		if registros[i].Arquivo == arquivo {
			alvo = &registros[i]
			break
		}
	}
	if alvo == nil {
		return ProcessamentoDTO{}, fmt.Errorf("registro não encontrado: %s", arquivo)
	}

	nome := a.clientDB[alvo.Codigo]
	if nome == "" {
		nome = "[Nome do cliente]"
	}
	texto := rentabilidade.MontarMensagem(template, *alvo, nome)
	if err := runtime.ClipboardSetText(a.ctx, texto); err != nil {
		return ProcessamentoDTO{}, fmt.Errorf("copiar para a área de transferência: %w", err)
	}
	if err := rentabilidade.MarcarCopiado(a.db, arquivo); err != nil {
		return ProcessamentoDTO{}, err
	}

	clientes, err := a.clientesDTO()
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	return ProcessamentoDTO{Clientes: clientes}, nil
}

// EnviarWhatsApp monta a mensagem final (mesma lógica de CopiarMensagem) e
// abre o WhatsApp (Desktop ou Web) já com ela preenchida na conversa do
// cliente — quem aperta "Enviar" lá dentro é o assessor, o app só
// pré-preenche. Erra se o código do cliente não tiver telefone cadastrado
// na base de clientes carregada.
func (a *App) EnviarWhatsApp(arquivo, template string) (ProcessamentoDTO, error) {
	if a.db == nil {
		return ProcessamentoDTO{}, fmt.Errorf("nenhuma pasta selecionada")
	}

	registros, err := rentabilidade.ListarRegistros(a.db)
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	var alvo *rentabilidade.Registro
	for i := range registros {
		if registros[i].Arquivo == arquivo {
			alvo = &registros[i]
			break
		}
	}
	if alvo == nil {
		return ProcessamentoDTO{}, fmt.Errorf("registro não encontrado: %s", arquivo)
	}

	telefone := a.telefones[alvo.Codigo]
	if telefone == "" {
		return ProcessamentoDTO{}, fmt.Errorf("cliente %s não tem telefone cadastrado na base de clientes", alvo.Codigo)
	}

	nome := a.clientDB[alvo.Codigo]
	if nome == "" {
		nome = "[Nome do cliente]"
	}
	texto := rentabilidade.MontarMensagem(template, *alvo, nome)

	link, err := whatsapp.MontarLink(telefone, texto)
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	runtime.BrowserOpenURL(a.ctx, link)
	if err := rentabilidade.MarcarCopiado(a.db, arquivo); err != nil {
		return ProcessamentoDTO{}, err
	}

	clientes, err := a.clientesDTO()
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	return ProcessamentoDTO{Clientes: clientes}, nil
}

// CopiarMensagemFestas monta a mensagem de festas (template + nome, sem
// depender de relatório processado — ver rentabilidade.MontarMensagemFestas),
// copia pra área de transferência e marca o cliente como já tendo recebido a
// mensagem nesta leva (badge ENVIADO).
func (a *App) CopiarMensagemFestas(codigo, template string) (ProcessamentoDTO, error) {
	if a.db == nil {
		return ProcessamentoDTO{}, fmt.Errorf("nenhuma pasta selecionada")
	}

	nome := a.clientDB[codigo]
	if nome == "" {
		nome = "[Nome do cliente]"
	}
	texto := rentabilidade.MontarMensagemFestas(template, nome)
	if err := runtime.ClipboardSetText(a.ctx, texto); err != nil {
		return ProcessamentoDTO{}, fmt.Errorf("copiar para a área de transferência: %w", err)
	}
	if err := rentabilidade.MarcarFestasEnviado(a.db, codigo); err != nil {
		return ProcessamentoDTO{}, err
	}

	clientes, err := a.clientesDTO()
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	return ProcessamentoDTO{Clientes: clientes}, nil
}

// EnviarWhatsAppFestas é a versão Modo Festas de EnviarWhatsApp: mesma ideia
// (abre o WhatsApp com a mensagem pronta), mas monta o texto a partir do
// modelo de festas + nome, sem precisar de relatório processado.
func (a *App) EnviarWhatsAppFestas(codigo, template string) (ProcessamentoDTO, error) {
	if a.db == nil {
		return ProcessamentoDTO{}, fmt.Errorf("nenhuma pasta selecionada")
	}

	telefone := a.telefones[codigo]
	if telefone == "" {
		return ProcessamentoDTO{}, fmt.Errorf("cliente %s não tem telefone cadastrado na base de clientes", codigo)
	}

	nome := a.clientDB[codigo]
	if nome == "" {
		nome = "[Nome do cliente]"
	}
	texto := rentabilidade.MontarMensagemFestas(template, nome)

	link, err := whatsapp.MontarLink(telefone, texto)
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	runtime.BrowserOpenURL(a.ctx, link)
	if err := rentabilidade.MarcarFestasEnviado(a.db, codigo); err != nil {
		return ProcessamentoDTO{}, err
	}

	clientes, err := a.clientesDTO()
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	return ProcessamentoDTO{Clientes: clientes}, nil
}

// LimparFestasEnviados apaga o registro de quem já recebeu a mensagem de
// festas — usado antes de começar uma nova leva (Natal do ano seguinte,
// por exemplo). Não mexe nos registros de rentabilidade.
func (a *App) LimparFestasEnviados() (ProcessamentoDTO, error) {
	if a.db == nil {
		return ProcessamentoDTO{}, fmt.Errorf("nenhuma pasta selecionada")
	}
	if err := rentabilidade.LimparFestasEnviados(a.db); err != nil {
		return ProcessamentoDTO{}, err
	}
	clientes, err := a.clientesDTO()
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	return ProcessamentoDTO{Clientes: clientes}, nil
}

// ObterImagemGrafico renderiza e recorta o gráfico de rentabilidade
// histórica do PDF do cliente (sempre a mesma página/posição no modelo de
// relatório) e devolve como PNG em base64 — o frontend copia pra área de
// transferência via Clipboard API do navegador (a API de clipboard nativa
// do Wails só lida com texto).
func (a *App) ObterImagemGrafico(arquivo string) (string, error) {
	if a.pasta == "" {
		return "", fmt.Errorf("nenhuma pasta selecionada")
	}
	if err := a.garantirExtractor(); err != nil {
		return "", err
	}

	png, err := a.extractor.RenderizarGraficoRentabilidade(filepath.Join(a.pasta, arquivo), recorteGraficoConfigurado())
	if err != nil {
		return "", fmt.Errorf("gerar imagem do gráfico: %w", err)
	}
	return base64.StdEncoding.EncodeToString(png), nil
}

// recorteGraficoConfigurado devolve o recorte personalizado salvo em
// config.json, ou o padrão do pacote pdfreport se o usuário nunca
// configurou um recorte próprio nas Configurações.
func recorteGraficoConfigurado() pdfreport.RecorteGrafico {
	cfg := carregarConfig()
	if !cfg.RecortePersonalizado {
		return pdfreport.RecorteGraficoRentabilidadePadrao
	}
	return pdfreport.RecorteGrafico{X0: cfg.RecorteX0, Y0: cfg.RecorteY0, X1: cfg.RecorteX1, Y1: cfg.RecorteY1}
}

// ObterPreviaPaginaGraficoRentabilidade devolve a página inteira (sem
// recorte) do relatório em base64 — usada pela tela de configuração do
// recorte, pra o usuário ver e ajustar a área selecionada sobre a página
// real.
func (a *App) ObterPreviaPaginaGraficoRentabilidade(arquivo string) (string, error) {
	if a.pasta == "" {
		return "", fmt.Errorf("nenhuma pasta selecionada")
	}
	if err := a.garantirExtractor(); err != nil {
		return "", err
	}

	png, err := a.extractor.RenderizarPaginaGraficoCompleta(filepath.Join(a.pasta, arquivo))
	if err != nil {
		return "", fmt.Errorf("gerar prévia da página: %w", err)
	}
	return base64.StdEncoding.EncodeToString(png), nil
}

// SalvarRecorteGraficoRentabilidade persiste o recorte personalizado
// (frações 0 a 1 da página) usado por "Copiar imagem" na aba Rentabilidade.
func (a *App) SalvarRecorteGraficoRentabilidade(x0, y0, x1, y1 float64) error {
	cfg := carregarConfig()
	cfg.RecortePersonalizado = true
	cfg.RecorteX0, cfg.RecorteY0, cfg.RecorteX1, cfg.RecorteY1 = x0, y0, x1, y1
	return salvarConfig(cfg)
}

// RestaurarRecorteGraficoPadrao descarta o recorte personalizado, voltando
// a usar pdfreport.RecorteGraficoRentabilidadePadrao.
func (a *App) RestaurarRecorteGraficoPadrao() error {
	cfg := carregarConfig()
	cfg.RecortePersonalizado = false
	return salvarConfig(cfg)
}

// LimparTudo apaga todos os registros da pasta ativa (confirmação fica a
// cargo do frontend). Os PDFs continuam no disco e são reprocessados do
// zero na próxima leitura.
func (a *App) LimparTudo() (ProcessamentoDTO, error) {
	if a.db == nil {
		return ProcessamentoDTO{}, fmt.Errorf("nenhuma pasta selecionada")
	}
	if err := rentabilidade.LimparBanco(a.db); err != nil {
		return ProcessamentoDTO{}, err
	}
	clientes, err := a.clientesDTO()
	if err != nil {
		return ProcessamentoDTO{}, err
	}
	return ProcessamentoDTO{Clientes: clientes}, nil
}

// ExportarCSV abre o diálogo nativo de salvar e grava a planilha com todos
// os registros da pasta ativa (ignora qualquer filtro de busca aplicado na
// UI). Devolve "" (sem erro) se o usuário cancelar o diálogo.
func (a *App) ExportarCSV() (string, error) {
	if a.db == nil {
		return "", fmt.Errorf("nenhuma pasta selecionada")
	}
	registros, err := rentabilidade.ListarRegistros(a.db)
	if err != nil {
		return "", err
	}
	if len(registros) == 0 {
		return "", fmt.Errorf("não há clientes processados para exportar")
	}

	caminho, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Salvar planilha",
		DefaultDirectory: a.pasta,
		DefaultFilename:  "rentabilidades.csv",
		Filters: []runtime.FileFilter{
			{DisplayName: "Planilha CSV (Excel)", Pattern: "*.csv"},
		},
	})
	if err != nil {
		return "", err
	}
	if caminho == "" {
		return "", nil // usuário cancelou o diálogo
	}

	if err := rentabilidade.ExportarCSV(caminho, registros); err != nil {
		// Caso comum: o próprio CSV aberto no Excel, que tranca o arquivo.
		return "", fmt.Errorf("não foi possível salvar o arquivo. Se ele estiver aberto no Excel, feche e tente de novo.\n\n(%w)", err)
	}
	return caminho, nil
}

// CarregarBaseClientes abre o diálogo nativo de abrir arquivo, lê o CSV e
// salva o caminho escolhido na config. Devolve nil (sem erro) se o usuário
// cancelar o diálogo. A lista de clientes da aba Rentabilidade já vem
// reconstruída com a nova base (o join cliente↔registro muda quando a base
// muda) — nil quando ainda não há pasta selecionada.
func (a *App) CarregarBaseClientes() (*BaseClientesDTO, error) {
	caminho, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar base de clientes (CSV com colunas código e nome)",
		Filters: []runtime.FileFilter{
			{DisplayName: "CSV", Pattern: "*.csv"},
		},
	})
	if err != nil {
		return nil, err
	}
	if caminho == "" {
		return nil, nil // usuário cancelou o diálogo
	}

	base, err := clientdb.CarregarBaseClientes(caminho)
	if err != nil {
		return nil, fmt.Errorf("não foi possível ler o arquivo.\n\n(%w)", err)
	}
	a.clientDB, a.telefones, a.emails = separarNomesTelefonesEmails(base)

	cfg := carregarConfig()
	cfg.BaseClientes = caminho
	if err := salvarConfig(cfg); err != nil {
		return nil, err
	}

	dto := &BaseClientesDTO{ClientDB: a.clientDB, ClientEmails: a.emails}
	if a.db != nil {
		clientes, err := a.clientesDTO()
		if err != nil {
			return nil, err
		}
		dto.Clientes = clientes
	}
	return dto, nil
}

// ---------------------------------------------------------------------------
// Gerador de E-mails de Ordem
// ---------------------------------------------------------------------------

// CategoriasEmail devolve o catálogo completo (produtos, categorias, campos
// e o texto de Operações Estruturadas) — dado estático, buscado uma vez no
// boot do frontend.
func (a *App) CategoriasEmail() CatalogoEmailDTO {
	return catalogoEmailDTO()
}

// GerarEmail monta o texto do e-mail de ordem a partir de 1+ operações já
// preenchidas. modo é "padronizado" (regra de compliance: um produto por
// e-mail, com o modelo específico do produto) ou "livre" (comportamento
// antigo, permite misturar produtos).
func (a *App) GerarEmail(codigo, nome string, itens []ItemEmailEntrada, modo string) (string, error) {
	if len(itens) == 0 {
		return "", fmt.Errorf("adicione ao menos uma operação")
	}
	if codigo == "" {
		codigo = "[Código do cliente]"
	}
	if nome == "" {
		nome = "[Nome do cliente]"
	}

	var itensResolvidos []emailgen.Item
	for _, item := range itens {
		cat := emailgen.CategoriaPorGrupoLabel(item.Group, item.Label)
		if cat == nil {
			continue
		}
		if modo == "livre" && cat.SoPadronizado {
			return "", fmt.Errorf("%s - %s só está disponível no modo padronizado (um produto por e-mail)", cat.Group, cat.Label)
		}
		itensResolvidos = append(itensResolvidos, emailgen.Item{
			Cat:     cat,
			Valores: cat.ResolveValores(item.Valores),
		})
	}
	if len(itensResolvidos) == 0 {
		return "", fmt.Errorf("selecione ao menos uma operação válida")
	}

	if modo == "livre" {
		return emailgen.MontarTextoEmail(codigo, nome, itensResolvidos), nil
	}
	return emailgen.MontarTextoEmailPadronizado(codigo, nome, itensResolvidos)
}

// SalvarPreferencias persiste tema, acento, modo de e-mail, tabela
// previdenciária, visão (cliente/assessor), fonte da interface, modo
// apresentação, modo festas, ordem da barra lateral e ferramentas escondidas
// dela em config.json.
func (a *App) SalvarPreferencias(tema, acento, modoEmail, tabelaPrevidenciaria, visao, fonte string, modoApresentacao, modoFestas bool, ordemNav, ordemNavOcultos []string) error {
	cfg := carregarConfig()
	cfg.Tema, cfg.Acento, cfg.ModoEmail, cfg.TabelaPrevidenciaria, cfg.Visao, cfg.Fonte = tema, acento, modoEmail, tabelaPrevidenciaria, visao, fonte
	cfg.ModoApresentacao = modoApresentacao
	cfg.ModoFestas = modoFestas
	cfg.OrdemNav = ordemNav
	cfg.OrdemNavOcultos = ordemNavOcultos
	return salvarConfig(cfg)
}

// CopiarTexto copia um texto arbitrário pra área de transferência — usado
// pelo botão "Copiar texto" da aba de e-mails (o texto vem do textarea, que
// é editável pelo usuário antes de copiar).
func (a *App) CopiarTexto(texto string) error {
	return runtime.ClipboardSetText(a.ctx, texto)
}

// SalvarEmailRemetente persiste qual conta do Outlook (e-mail) deve ser
// usada como remetente pelos rascunhos abertos via AbrirEmailNoOutlook —
// evita ambiguidade quando o Outlook tem mais de uma conta logada.
func (a *App) SalvarEmailRemetente(email string) error {
	cfg := carregarConfig()
	cfg.EmailRemetente = strings.TrimSpace(email)
	return salvarConfig(cfg)
}

// SalvarDadosAssessor persiste o nome e o e-mail do assessor usados pra
// responder automaticamente as duas primeiras perguntas do Typeform
// ("Nome do assessor responsável" e "E-mail do assessor") no preenchimento
// automático (ver PreencherTypeform) — sem isso, a automação sempre parava
// logo na primeira pergunta por falta de resposta salva pra ela.
func (a *App) SalvarDadosAssessor(nome, email string) error {
	cfg := carregarConfig()
	cfg.AssessorNome = strings.TrimSpace(nome)
	cfg.AssessorEmail = strings.TrimSpace(email)
	return salvarConfig(cfg)
}

// AbrirEmailNoOutlook abre o texto já gerado como rascunho no Outlook
// Classic (desktop), endereçado ao destinatário informado, com a
// assinatura padrão do próprio Outlook — não envia, só deixa pronto pra
// revisão. Requer Outlook Classic instalado (registrado via COM). Usa
// cfg.EmailRemetente (Configurações > E-mail) pra escolher a conta certa
// quando o Outlook tem mais de uma logada — vazio deixa o Outlook decidir
// sozinho.
func (a *App) AbrirEmailNoOutlook(destinatario, assunto, corpo string) error {
	if strings.TrimSpace(destinatario) == "" {
		return fmt.Errorf("informe o e-mail do destinatário")
	}
	cfg := carregarConfig()
	return outlookmail.AbrirRascunho(destinatario, assunto, corpo, cfg.EmailRemetente)
}

// ---------------------------------------------------------------------------
// Tabela de Deságio
// ---------------------------------------------------------------------------

// SalvarImagemPNG recebe uma imagem PNG codificada em base64 (gerada no
// frontend via <canvas>.toDataURL, sem o prefixo "data:image/png;base64,")
// e abre o diálogo nativo de salvar. Devolve "" (sem erro) se o usuário
// cancelar o diálogo.
func (a *App) SalvarImagemPNG(base64PNG string) (string, error) {
	dados, err := base64.StdEncoding.DecodeString(base64PNG)
	if err != nil {
		return "", fmt.Errorf("decodificar imagem: %w", err)
	}

	caminho, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Salvar tabela de deságio",
		DefaultFilename: "tabela_desagio.png",
		Filters: []runtime.FileFilter{
			{DisplayName: "Imagem PNG", Pattern: "*.png"},
		},
	})
	if err != nil {
		return "", err
	}
	if caminho == "" {
		return "", nil // usuário cancelou o diálogo
	}

	if err := os.WriteFile(caminho, dados, 0644); err != nil {
		return "", fmt.Errorf("não foi possível salvar a imagem.\n\n(%w)", err)
	}
	return caminho, nil
}

// ---------------------------------------------------------------------------
// Apresentação institucional
// ---------------------------------------------------------------------------

// Tetos de tamanho por tipo de arquivo da apresentação. HTML autocontido
// (CSS + imagens em base64) cabe folgado em 25 MB; PDF (que pode ter páginas
// cheias de imagens em alta resolução) ganha um teto maior. Os tetos só
// evitam carregar por engano um arquivo gigante que travaria a webview ou
// deixaria a chamada IPC (o PDF vai em base64 no JSON) lenta demais.
const (
	limiteApresentacaoHTMLBytes = 25 << 20 // 25 MB
	limiteApresentacaoPDFBytes  = 60 << 20 // 60 MB
)

// mensagemFormatoOfficeNaoSuportado orienta o usuário a exportar
// PowerPoint/ODP para PDF antes de carregar — este app não converte esses
// formatos automaticamente (decisão consciente: evita depender de instalar
// um conversor externo como o LibreOffice na máquina do usuário).
const mensagemFormatoOfficeNaoSuportado = "Arquivos do PowerPoint (.pptx/.ppsx) e OpenDocument (.odp) ainda não são lidos diretamente. " +
	"Abra o arquivo no PowerPoint, LibreOffice Impress ou Google Slides e use \"Salvar como\" (ou \"Exportar\") " +
	"escolhendo PDF — depois selecione o PDF gerado aqui."

// lerApresentacao lê o arquivo (HTML ou PDF) do caminho salvo em
// config.json. Não é erro de chamada quando não há arquivo escolhido,
// quando o arquivo sumiu/está ilegível, ou quando o formato não é
// suportado — tudo isso vira o campo Erro do DTO, pra a aba mostrar um
// aviso amigável em vez de quebrar. Caminho vazio = nenhum arquivo
// escolhido.
func lerApresentacao(caminho string) ApresentacaoDTO {
	if caminho == "" {
		return ApresentacaoDTO{}
	}
	dto := ApresentacaoDTO{Caminho: caminho}
	info, err := os.Stat(caminho)
	if err != nil {
		dto.Erro = "O arquivo de apresentação não foi encontrado. Ele pode ter sido movido, renomeado ou excluído. Escolha o arquivo de novo."
		return dto
	}

	switch strings.ToLower(filepath.Ext(caminho)) {
	case ".html", ".htm":
		if info.Size() > limiteApresentacaoHTMLBytes {
			dto.Erro = "O arquivo de apresentação é grande demais (acima de 25 MB). Reduza o tamanho das imagens embutidas."
			return dto
		}
		dados, err := os.ReadFile(caminho)
		if err != nil {
			dto.Erro = "Não foi possível ler o arquivo de apresentação: " + err.Error()
			return dto
		}
		dto.Tipo = "html"
		dto.HTML = string(dados)

	case ".pdf":
		if info.Size() > limiteApresentacaoPDFBytes {
			dto.Erro = "O arquivo PDF é grande demais (acima de 60 MB)."
			return dto
		}
		dados, err := os.ReadFile(caminho)
		if err != nil {
			dto.Erro = "Não foi possível ler o arquivo de apresentação: " + err.Error()
			return dto
		}
		dto.Tipo = "pdf"
		dto.PDFBase64 = base64.StdEncoding.EncodeToString(dados)

	case ".pptx", ".ppsx", ".odp":
		dto.Erro = mensagemFormatoOfficeNaoSuportado

	default:
		dto.Erro = "Formato não suportado. Use um arquivo .html, .htm ou .pdf."
	}

	return dto
}

// CarregarApresentacao lê o HTML da apresentação atualmente salva em
// config.json (chamada quando a aba Apresentação é aberta ou recarregada).
func (a *App) CarregarApresentacao() ApresentacaoDTO {
	return lerApresentacao(carregarConfig().Apresentacao)
}

// EscolherApresentacao abre o diálogo nativo pra o usuário apontar um arquivo
// HTML, salva o caminho em config.json e já devolve o conteúdo lido. Devolve
// um DTO com Caminho vazio (sem erro) se o usuário cancelar o diálogo — a aba
// mantém o que já estava.
func (a *App) EscolherApresentacao() (ApresentacaoDTO, error) {
	caminho, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar apresentação (HTML ou PDF)",
		Filters: []runtime.FileFilter{
			{DisplayName: "Apresentação (HTML, PDF, PowerPoint, ODP)", Pattern: "*.html;*.htm;*.pdf;*.pptx;*.ppsx;*.odp"},
		},
	})
	if err != nil {
		return ApresentacaoDTO{}, err
	}
	if caminho == "" {
		return ApresentacaoDTO{}, nil // usuário cancelou o diálogo
	}

	cfg := carregarConfig()
	cfg.Apresentacao = caminho
	if err := salvarConfig(cfg); err != nil {
		return ApresentacaoDTO{}, err
	}
	return lerApresentacao(caminho), nil
}

// RemoverApresentacao esquece o arquivo escolhido (o arquivo em si continua no
// disco). A aba volta ao estado "nenhuma apresentação".
func (a *App) RemoverApresentacao() error {
	cfg := carregarConfig()
	cfg.Apresentacao = ""
	return salvarConfig(cfg)
}

// dpiApresentacaoSlide é a resolução usada pra renderizar cada página do PDF
// como imagem de slide no modo apresentação — mesmo valor usado em
// RenderizarGraficoRentabilidade, boa o bastante pra tela cheia sem gerar
// imagens grandes demais.
const dpiApresentacaoSlide = 200

// ApresentacaoContarPaginas devolve o número de páginas do PDF em caminho —
// chamado ao entrar no modo apresentação (slideshow) da aba Apresentação,
// pra saber até onde a navegação por slides pode avançar.
func (a *App) ApresentacaoContarPaginas(caminho string) (int, error) {
	if err := a.garantirExtractor(); err != nil {
		return 0, err
	}
	return a.extractor.ContarPaginas(caminho)
}

// ApresentacaoRenderizarPagina renderiza uma página do PDF em caminho como
// imagem PNG (base64) — cada chamada é um slide do modo apresentação
// (slideshow) da aba Apresentação, sem a barra de ferramentas do
// visualizador nativo de PDF.
func (a *App) ApresentacaoRenderizarPagina(caminho string, indice int) (string, error) {
	if err := a.garantirExtractor(); err != nil {
		return "", err
	}
	png, err := a.extractor.RenderizarPagina(caminho, indice, dpiApresentacaoSlide)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(png), nil
}

// ---------------------------------------------------------------------------
// Typeform do Cliente
// ---------------------------------------------------------------------------

// caracteresInvalidosNomeArquivo são proibidos (ou problemáticos) em nomes
// de arquivo no Windows.
var caracteresInvalidosNomeArquivo = strings.NewReplacer(
	"\\", "-", "/", "-", ":", "-", "*", "", "?", "", "\"", "", "<", "", ">", "", "|", "-",
)

// nomeArquivoTypeform monta o nome do .txt salvo por SalvarTypeform a partir
// do nome do cliente (pergunta 3 do formulário) + timestamp — o timestamp
// evita sobrescrever um typeform antigo do mesmo cliente numa reunião nova.
func nomeArquivoTypeform(nomeCliente string) string {
	nome := strings.TrimSpace(nomeCliente)
	if nome == "" {
		nome = "Cliente"
	}
	nome = caracteresInvalidosNomeArquivo.Replace(nome)
	return fmt.Sprintf("Typeform - %s - %s.txt", nome, time.Now().Format("2006-01-02_1504"))
}

// SalvarTypeform grava o texto já formatado pelo frontend (aba Typeform) num
// .txt dentro da subpasta "Typeforms" da pasta ativa de Rentabilidade,
// criando a subpasta se ainda não existir. Devolve o caminho salvo.
func (a *App) SalvarTypeform(nomeCliente, texto string) (string, error) {
	if a.pasta == "" {
		return "", fmt.Errorf("nenhuma pasta selecionada")
	}

	destino := filepath.Join(a.pasta, "Typeforms")
	if err := os.MkdirAll(destino, 0755); err != nil {
		return "", fmt.Errorf("criar pasta Typeforms: %w", err)
	}

	caminho := filepath.Join(destino, nomeArquivoTypeform(nomeCliente))
	if err := os.WriteFile(caminho, []byte(texto), 0644); err != nil {
		return "", fmt.Errorf("não foi possível salvar o arquivo.\n\n(%w)", err)
	}
	return caminho, nil
}

// typeformURLPadrao é o formulário online de diagnóstico financeiro que a
// aba Typeform replica localmente (ver frontend/src/tabs/typeform.js).
const typeformURLPadrao = "https://myn161a88d3.typeform.com/to/GwdhwAMm"

// PreencherTypeform abre o Typeform real no Edge (janela visível) e
// preenche as telas com os itens já respondidos na aba Typeform, casando
// cada tela pelo texto da pergunta (o formulário real muda de redação e tem
// perguntas condicionais — ver internal/typeformbot). Emite
// "typeform:progresso" a cada pergunta preenchida. O preenchimento sempre
// para antes do fim (pergunta sem correspondência confiável, ou a última
// resposta disponível) — o navegador continua aberto pra o assessor
// terminar e enviar manualmente; nunca envia sozinho. O erro devolvido
// nesse caso (*typeformbot.ErroParado, mas cruza a fronteira Wails como
// texto) já é uma mensagem pensada pra aparecer pro usuário, não uma falha
// a esconder.
func (a *App) PreencherTypeform(itens []ItemRespostaTypeformDTO) error {
	respostas := make([]typeformbot.Resposta, len(itens))
	for i, item := range itens {
		respostas[i] = typeformbot.Resposta{Pergunta: item.Pergunta, Valor: item.Valor}
	}

	return typeformbot.Preencher(typeformURLPadrao, respostas, func(feitos, total int, pergunta string) {
		runtime.EventsEmit(a.ctx, "typeform:progresso", feitos, total, pergunta)
	})
}

// ---------------------------------------------------------------------------
// Editor de PDF
// ---------------------------------------------------------------------------

// EscolherPDFParaEditar abre o diálogo nativo de abrir arquivo. Devolve ""
// (sem erro) se o usuário cancelar. Não lê nem processa o arquivo — a aba
// usa ApresentacaoContarPaginas/ApresentacaoRenderizarPagina (já genéricas,
// qualquer PDF) pra abrir e renderizar as páginas.
func (a *App) EscolherPDFParaEditar() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar PDF para editar",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF", Pattern: "*.pdf"},
		},
	})
}

// nomeSugeridoPDFEditado deriva "<nome original>_Edited.pdf" do caminho do
// PDF aberto — nome sugerido no diálogo de salvar (o usuário pode trocar).
func nomeSugeridoPDFEditado(caminhoOriginal string) string {
	base := filepath.Base(caminhoOriginal)
	semExt := strings.TrimSuffix(base, filepath.Ext(base))
	return semExt + "_Edited.pdf"
}

// pastaDownloads devolve a pasta de Downloads do usuário (~/Downloads) — é
// só a pasta sugerida no diálogo de salvar; o usuário pode escolher outra.
func pastaDownloads() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Downloads"), nil
}

// PaginaPDFEditadaDTO é uma página já achatada (a página original
// renderizada + o que o usuário desenhou/escreveu por cima, tudo achatado
// no <canvas> do frontend) enviada pra SalvarPDFEditado — JPEGBase64 sem o
// prefixo "data:image/jpeg;base64,". LarguraPt/AlturaPt são o tamanho da
// página em pontos (1/72 polegada), pra manter a proporção original.
type PaginaPDFEditadaDTO struct {
	JPEGBase64 string
	LarguraPt  float64
	AlturaPt   float64
}

// SalvarPDFEditado monta um PDF novo a partir de paginas (uma imagem por
// página, na ordem final) e abre o diálogo nativo de salvar, sugerindo a
// pasta Downloads e o nome "<original>_Edited.pdf" — o PDF original em
// caminhoOriginal nunca é alterado. Devolve "" (sem erro) se o usuário
// cancelar o diálogo.
func (a *App) SalvarPDFEditado(paginas []PaginaPDFEditadaDTO, caminhoOriginal string) (string, error) {
	if len(paginas) == 0 {
		return "", fmt.Errorf("nenhuma página para salvar")
	}

	entradas := make([]pdfedit.Pagina, len(paginas))
	for i, p := range paginas {
		jpegBytes, err := base64.StdEncoding.DecodeString(p.JPEGBase64)
		if err != nil {
			return "", fmt.Errorf("decodificar página %d: %w", i+1, err)
		}
		entradas[i] = pdfedit.Pagina{JPEG: jpegBytes, LarguraPt: p.LarguraPt, AlturaPt: p.AlturaPt}
	}

	dados, err := pdfedit.Montar(entradas)
	if err != nil {
		return "", err
	}

	downloads, _ := pastaDownloads() // best-effort: se falhar, o diálogo só abre sem pasta sugerida
	caminho, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Salvar PDF editado",
		DefaultDirectory: downloads,
		DefaultFilename:  nomeSugeridoPDFEditado(caminhoOriginal),
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return "", err
	}
	if caminho == "" {
		return "", nil // usuário cancelou o diálogo
	}

	if err := os.WriteFile(caminho, dados, 0644); err != nil {
		return "", fmt.Errorf("não foi possível salvar o arquivo.\n\n(%w)", err)
	}
	return caminho, nil
}

// CriarPDFDeImagens monta um PDF novo a partir de imagens (uma por página,
// na ordem final escolhida pelo usuário na aba Imagens em PDF) e abre o
// diálogo nativo de salvar, sugerindo a pasta Downloads. Mesmo mecanismo de
// SalvarPDFEditado (reaproveita pdfedit.Montar e o mesmo DTO de página),
// só sem um PDF original de onde partir. Devolve "" (sem erro) se o
// usuário cancelar o diálogo.
func (a *App) CriarPDFDeImagens(paginas []PaginaPDFEditadaDTO) (string, error) {
	if len(paginas) == 0 {
		return "", fmt.Errorf("nenhuma imagem para converter")
	}

	entradas := make([]pdfedit.Pagina, len(paginas))
	for i, p := range paginas {
		jpegBytes, err := base64.StdEncoding.DecodeString(p.JPEGBase64)
		if err != nil {
			return "", fmt.Errorf("decodificar imagem %d: %w", i+1, err)
		}
		entradas[i] = pdfedit.Pagina{JPEG: jpegBytes, LarguraPt: p.LarguraPt, AlturaPt: p.AlturaPt}
	}

	dados, err := pdfedit.Montar(entradas)
	if err != nil {
		return "", err
	}

	downloads, _ := pastaDownloads() // best-effort: se falhar, o diálogo só abre sem pasta sugerida
	caminho, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Salvar PDF de imagens",
		DefaultDirectory: downloads,
		DefaultFilename:  "imagens_" + time.Now().Format("2006-01-02") + ".pdf",
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return "", err
	}
	if caminho == "" {
		return "", nil // usuário cancelou o diálogo
	}

	if err := os.WriteFile(caminho, dados, 0644); err != nil {
		return "", fmt.Errorf("não foi possível salvar o arquivo.\n\n(%w)", err)
	}
	return caminho, nil
}

// ---------------------------------------------------------------------------
// Assinaturas
// ---------------------------------------------------------------------------

// pastaAssinaturas devolve (criando se preciso) a pasta onde as imagens de
// assinatura ficam guardadas — ao lado do config.json, então não depende
// de onde o app está instalado nem exige admin.
func pastaAssinaturas() (string, error) {
	caminhoConfig, err := configPath()
	if err != nil {
		return "", err
	}
	pasta := filepath.Join(filepath.Dir(caminhoConfig), "assinaturas")
	if err := os.MkdirAll(pasta, 0755); err != nil {
		return "", fmt.Errorf("criar pasta de assinaturas: %w", err)
	}
	return pasta, nil
}

// ListarAssinaturas devolve todas as assinaturas salvas, marcando qual é a
// ativa (ver Configurações → Assinatura).
func (a *App) ListarAssinaturas() ([]AssinaturaDTO, error) {
	pasta, err := pastaAssinaturas()
	if err != nil {
		return nil, err
	}
	lista, err := assinaturas.Listar(pasta)
	if err != nil {
		return nil, err
	}
	ativa := carregarConfig().AssinaturaAtiva
	dto := make([]AssinaturaDTO, len(lista))
	for i, s := range lista {
		dto[i] = AssinaturaDTO{Nome: s.Nome, Base64: s.Base64, Ativa: s.Nome == ativa}
	}
	return dto, nil
}

// AdicionarAssinatura abre o diálogo nativo de abrir arquivo, copia a
// imagem escolhida pra pasta de assinaturas e já a marca como ativa (fica
// pronta pra uso imediato na ferramenta Imagem do Editor de PDF). Devolve
// uma entrada com Nome vazio (sem erro) se o usuário cancelar o diálogo.
func (a *App) AdicionarAssinatura() (AssinaturaDTO, error) {
	caminho, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar imagem de assinatura",
		Filters: []runtime.FileFilter{
			{DisplayName: "Imagem", Pattern: "*.png;*.jpg;*.jpeg"},
		},
	})
	if err != nil {
		return AssinaturaDTO{}, err
	}
	if caminho == "" {
		return AssinaturaDTO{}, nil // usuário cancelou o diálogo
	}

	pasta, err := pastaAssinaturas()
	if err != nil {
		return AssinaturaDTO{}, err
	}
	nome, err := assinaturas.Adicionar(pasta, caminho)
	if err != nil {
		return AssinaturaDTO{}, err
	}

	cfg := carregarConfig()
	cfg.AssinaturaAtiva = nome
	if err := salvarConfig(cfg); err != nil {
		return AssinaturaDTO{}, err
	}

	dados, err := os.ReadFile(filepath.Join(pasta, nome))
	if err != nil {
		return AssinaturaDTO{}, err
	}
	return AssinaturaDTO{Nome: nome, Base64: base64.StdEncoding.EncodeToString(dados), Ativa: true}, nil
}

// SelecionarAssinaturaAtiva troca qual assinatura salva está ativa.
func (a *App) SelecionarAssinaturaAtiva(nome string) error {
	cfg := carregarConfig()
	cfg.AssinaturaAtiva = nome
	return salvarConfig(cfg)
}

// RemoverAssinatura apaga a assinatura nome. Se era a ativa, limpa
// AssinaturaAtiva — a ferramenta Imagem do Editor de PDF volta a pedir
// escolha manual de arquivo até uma nova assinatura ficar ativa.
func (a *App) RemoverAssinatura(nome string) error {
	pasta, err := pastaAssinaturas()
	if err != nil {
		return err
	}
	if err := assinaturas.Remover(pasta, nome); err != nil {
		return err
	}

	cfg := carregarConfig()
	if cfg.AssinaturaAtiva == nome {
		cfg.AssinaturaAtiva = ""
		return salvarConfig(cfg)
	}
	return nil
}

// ObterAssinaturaAtiva devolve a assinatura ativa (ou nil se nenhuma) —
// usado pela ferramenta "Imagem" do Editor de PDF.
func (a *App) ObterAssinaturaAtiva() (*AssinaturaDTO, error) {
	cfg := carregarConfig()
	if cfg.AssinaturaAtiva == "" {
		return nil, nil
	}
	pasta, err := pastaAssinaturas()
	if err != nil {
		return nil, err
	}
	dados, err := os.ReadFile(filepath.Join(pasta, cfg.AssinaturaAtiva))
	if err != nil {
		return nil, nil // arquivo sumiu (ex.: apagado por fora) — trata como "nenhuma ativa", não como erro
	}
	return &AssinaturaDTO{Nome: cfg.AssinaturaAtiva, Base64: base64.StdEncoding.EncodeToString(dados), Ativa: true}, nil
}

// ---------------------------------------------------------------------------
// Validação de Assinatura ICP-Brasil
// ---------------------------------------------------------------------------
//
// Verificação inteiramente local/offline (o certificado do signatário viaja
// dentro da própria assinatura CMS/PKCS#7) — nenhuma chamada de rede em
// tempo de validação, só pra manter a cadeia de certificados ICP-Brasil
// atualizada (dado público, ver internal/icpbrasil).
//
// Restrição importante, não relaxar: nenhum resultado de validação é
// persistido em disco (nada de SQLite/JSON/log com nome/CPF/hash/caminho —
// um novo documento sobrescreve o resultado anterior só em memória). O
// arquivo do cliente também nunca é copiado pra pasta temporária: é lido
// direto do caminho escolhido no diálogo.

// pastaICPBrasil devolve (tentando criar, best-effort) a pasta onde a
// cadeia de certificados ICP-Brasil fica guardada — ao lado do
// config.json, mesmo padrão de pastaAssinaturas(). Nunca retorna erro: é
// dado público sem restrição de privacidade, e icpbrasil.NovoStore já
// degrada sozinho pro snapshot embutido no binário quando a pasta está
// vazia/inacessível.
func pastaICPBrasil() string {
	caminhoConfig, err := configPath()
	if err != nil {
		return ""
	}
	pasta := filepath.Join(filepath.Dir(caminhoConfig), "icp-cadeia")
	_ = os.MkdirAll(pasta, 0755)
	return pasta
}

// EscolherArquivosAssinatura abre o diálogo nativo de seleção múltipla —
// o usuário escolhe um PDF assinado (PAdES), ou o .p7s/.p7m e, se a
// assinatura for destacada, o documento original junto (Ctrl+click nos
// dois).
func (a *App) EscolherArquivosAssinatura() ([]string, error) {
	caminhos, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Selecionar assinatura digital (.pdf, .p7s ou .p7m)",
		Filters: []runtime.FileFilter{
			{DisplayName: "Assinatura digital", Pattern: "*.pdf;*.p7s;*.p7m"},
			{DisplayName: "Todos os arquivos", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	return caminhos, nil // slice vazio = usuário cancelou o diálogo
}

// ValidarAssinaturaICP valida uma assinatura ICP-Brasil (CAdES ou PAdES) a
// partir dos caminhos escolhidos em EscolherArquivosAssinatura. error só é
// usado pra falha dura (nenhum arquivo, contexto cancelado) — todo o resto
// (formato não suportado, cadeia não reconhecida, revogação inconclusiva
// etc.) vira ResultadoValidacaoDTO.Estado, seguindo a mesma convenção de
// ApresentacaoDTO.Erro no resto do app.
func (a *App) ValidarAssinaturaICP(caminhos []string) (ResultadoValidacaoDTO, error) {
	if len(caminhos) == 0 {
		return ResultadoValidacaoDTO{}, fmt.Errorf("nenhum arquivo selecionado")
	}

	_, caminhoP7S, caminhoConteudo, err := sigvalidator.DetectarFormato(caminhos)
	if err != nil {
		return ResultadoValidacaoDTO{}, err
	}

	resultado, err := sigvalidator.Validate(a.ctx, sigvalidator.Input{
		CaminhoAssinatura: caminhoP7S,
		CaminhoConteudo:   caminhoConteudo,
	}, a.icpStore)
	if err != nil {
		return ResultadoValidacaoDTO{}, err
	}

	// Nunca logar nome/CPF/hash/caminho — só estado + tempo de processamento.
	log.Println("validação de assinatura ICP-Brasil:", resultado.Estado, resultado.TempoProcessamento)

	return paraResultadoValidacaoDTO(resultado), nil
}

// ObterInfoCadeiaICP devolve a data/origem da cadeia ICP-Brasil em uso
// localmente — mostrado no rodapé da aba de validação.
func (a *App) ObterInfoCadeiaICP() (InfoCadeiaDTO, error) {
	if a.icpStore == nil {
		return InfoCadeiaDTO{}, fmt.Errorf("cadeia ICP-Brasil não inicializada")
	}
	return paraInfoCadeiaDTO(a.icpStore.Info()), nil
}

// AtualizarCadeiaICP força uma atualização síncrona da cadeia a partir dos
// servidores do ITI (botão "Atualizar" na UI). Falha de rede vira
// InfoCadeiaDTO.Erro, não Go error — a cadeia local continua valendo
// normalmente mesmo se isso falhar.
func (a *App) AtualizarCadeiaICP() (InfoCadeiaDTO, error) {
	if a.icpStore == nil {
		return InfoCadeiaDTO{}, fmt.Errorf("cadeia ICP-Brasil não inicializada")
	}
	info, err := a.icpStore.AtualizarAgora(a.ctx)
	dto := paraInfoCadeiaDTO(info)
	if err != nil {
		dto.Erro = "Não foi possível atualizar a cadeia agora (" + err.Error() + "). A cadeia local em uso continua valendo normalmente."
	}
	return dto, nil
}

// ---------------------------------------------------------------------------
// Atualização automática
// ---------------------------------------------------------------------------
//
// O app checa releases publicadas em internal/selfupdate.Repositorio (ver
// aquele arquivo pra configurar o repositório). Achou uma versão mais nova:
// o evento "atualizacao:disponivel" avisa o frontend (checagem automática
// no startup) e/ou VerificarAtualizacao devolve o mesmo resultado (checagem
// manual, botão em Configurações → Sobre). Em qualquer um dos dois casos, a
// release fica guardada em a.atualizacaoDisponivel — AplicarAtualizacao usa
// esse cache, então o frontend não precisa saber URLs de asset, só pedir
// pra aplicar a última encontrada.

// VersaoAtual devolve a versão deste build (ver comentário em main.go —
// "dev" fora de um build de release, quando a checagem de atualização
// também fica desligada).
func (a *App) VersaoAtual() string {
	return Version
}

// Plataforma devolve o sistema em que o app está rodando ("windows" ou
// "darwin"). O frontend usa isso pra esconder o que não existe no macOS:
// a aba Typeform (depende do Edge nos caminhos do Windows) e o botão
// "Abrir no Outlook" (automação COM). Esconder é melhor que deixar o botão
// lá e falhar no clique.
func (a *App) Plataforma() string {
	return goruntime.GOOS
}

// VerificarAtualizacao dispara uma checagem manual (botão "Verificar
// agora" em Configurações → Sobre). Falha de rede vira AtualizacaoDTO.Erro,
// não Go error — mesma convenção de AtualizarCadeiaICP, checagem de
// atualização não é algo que deveria assustar o assessor com uma tela de
// erro só porque a internet caiu num momento ruim.
func (a *App) VerificarAtualizacao() AtualizacaoDTO {
	return a.checarAtualizacao(a.ctx)
}

func (a *App) checarAtualizacao(ctx context.Context) AtualizacaoDTO {
	if Version == "dev" {
		return AtualizacaoDTO{Erro: "Build de desenvolvimento — sem versão de release pra comparar."}
	}
	// A atualização automática é só do Windows, e a trava é DELIBERADA:
	//
	//  1. selfupdate.UltimaVersao escolhe o asset pelo sufixo ".exe" — no
	//     macOS ele baixaria o binário do Windows e sobrescreveria o app
	//     com um executável que não roda.
	//  2. No macOS o programa vive dentro de um pacote .app assinado (ainda
	//     que com assinatura ad-hoc, obrigatória no Apple Silicon). Trocar
	//     só o binário de dentro invalida a assinatura e o sistema passa a
	//     recusar a abertura — quebraria o app do assessor de vez.
	//
	// Enquanto a atualização de pacote .app não estiver implementada, o Mac
	// avisa que existe versão nova e manda baixar à mão.
	if goruntime.GOOS != "windows" {
		return AtualizacaoDTO{
			Erro: "No macOS a atualização é manual: baixe a versão mais recente em " +
				"https://github.com/" + selfupdate.Repositorio + "/releases/latest",
		}
	}

	info, encontrado, err := selfupdate.UltimaVersao(ctx)
	if err != nil {
		return AtualizacaoDTO{Erro: "Não foi possível checar atualizações agora (" + err.Error() + ")."}
	}
	if !encontrado || !selfupdate.MaisNova(Version, info.Versao) {
		a.atualizacaoDisponivel.Store(nil)
		return AtualizacaoDTO{}
	}

	a.atualizacaoDisponivel.Store(&info)
	return AtualizacaoDTO{Disponivel: true, Versao: info.Versao, Notas: info.Notas}
}

// AplicarAtualizacao baixa e instala a release encontrada pela última
// checagem (automática ou via VerificarAtualizacao) e, se tudo der certo,
// relança o app já atualizado e encerra esta instância — o assessor só vê
// a janela fechar e reabrir. Se falhar antes de mexer no binário (download,
// checksum), devolve o erro e o app continua rodando normalmente na versão
// atual, sem risco pro que já estava aberto.
func (a *App) AplicarAtualizacao() error {
	info := a.atualizacaoDisponivel.Load()
	if info == nil {
		return fmt.Errorf("nenhuma atualização disponível — verifique novamente antes de aplicar")
	}

	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Minute)
	defer cancel()

	runtime.EventsEmit(a.ctx, "atualizacao:aplicando")
	if err := selfupdate.BaixarEAplicar(ctx, *info); err != nil {
		return err
	}

	// O binário em disco já é o novo — mesmo que o relançamento automático
	// falhe por algum motivo (ex. antivírus segurando o arquivo por um
	// instante), a próxima vez que o assessor abrir o .exe manualmente já
	// vem atualizado. Por isso só loga, não retorna erro pro frontend.
	if err := selfupdate.Relancar(); err != nil {
		log.Println("selfupdate: binário atualizado, mas falha ao relançar automaticamente:", err)
	}
	runtime.Quit(a.ctx)
	return nil
}
