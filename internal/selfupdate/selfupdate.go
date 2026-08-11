// Package selfupdate checa releases publicadas num repositório público do
// GitHub e, quando há uma versão mais nova, baixa o novo executável e
// relança o app a partir dele — sem instalador, sem intervenção manual do
// assessor além de clicar em "Atualizar agora" no app.
package selfupdate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Repositorio é o "owner/repo" público no GitHub onde as releases (o .exe
// mais o .sha256 correspondente) são publicadas — ver seção "Atualização
// automática" em DOCUMENTACAO.md pro passo a passo de release.
const Repositorio = "GustavoMedeir/ferramentas-assessoria"

// apiBaseURL é var (não const) só pra poder ser trocada por um
// httptest.Server nos testes — mesmo padrão de internal/icpbrasil/download.go.
var apiBaseURL = "https://api.github.com"

// limiteTamanhoExe é um teto de sanidade pro download do instalável — o
// build atual do app fica na casa de 15-30MB; nada real deveria chegar
// perto de 200MB, e isso evita prender memória indefinidamente se a URL
// vier de um asset errado ou um redirecionamento comprometido.
const limiteTamanhoExe = 200 << 20 // 200MB

// limiteTamanhoAPI é o teto pra resposta da API do GitHub (JSON de release)
// e pro arquivo de checksum — ambos são pequenos, isso é só um cinto de
// segurança contra resposta inesperada.
const limiteTamanhoAPI = 4 << 20 // 4MB

const timeoutAPI = 15 * time.Second

// ReleaseInfo descreve a release mais recente encontrada no GitHub, já com
// as URLs dos assets que interessam (o .exe e o .sha256 dele).
type ReleaseInfo struct {
	Versao      string // tag da release, ex. "v2.2.0"
	Notas       string // corpo da release no GitHub — vira o changelog mostrado no app
	URLExe      string
	URLChecksum string
}

// releaseGitHub é só o subconjunto do JSON de /releases/latest que
// interessa — o resto (autor, reações, etc.) é ignorado no unmarshal.
type releaseGitHub struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// UltimaVersao consulta a release mais recente do repositório e devolve
// suas informações. encontrado volta false se a release existe mas não tem
// os dois assets esperados (.exe e .sha256) — sinal de uma release
// publicada pela metade, que não deve disparar oferta de atualização.
func UltimaVersao(ctx context.Context) (info ReleaseInfo, encontrado bool, err error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", apiBaseURL, Repositorio)

	ctx, cancel := context.WithTimeout(ctx, timeoutAPI)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ReleaseInfo{}, false, err
	}
	// A API do GitHub recusa requisições sem User-Agent.
	req.Header.Set("User-Agent", "FerramentasAssessoria-Updater")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ReleaseInfo{}, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ReleaseInfo{}, false, fmt.Errorf("GitHub releases/latest: status %d", resp.StatusCode)
	}

	var rel releaseGitHub
	if err := json.NewDecoder(io.LimitReader(resp.Body, limiteTamanhoAPI)).Decode(&rel); err != nil {
		return ReleaseInfo{}, false, fmt.Errorf("resposta do GitHub não é o JSON esperado: %w", err)
	}

	info = ReleaseInfo{Versao: rel.TagName, Notas: rel.Body}

	// Acha primeiro o executável do Windows. Desde que a release passou a
	// levar também o pacote do macOS, ela tem MAIS DE UM arquivo .sha256
	// — um do .exe e outro do .zip do Mac. Pegar "qualquer .sha256", como
	// era feito antes, dava chance de casar o executável do Windows com o
	// checksum do Mac: a conferência falharia e a atualização morreria com
	// erro de arquivo corrompido.
	var nomeExe string
	for _, a := range rel.Assets {
		if strings.HasSuffix(a.Name, ".exe") {
			nomeExe = a.Name
			info.URLExe = a.BrowserDownloadURL
			break
		}
	}
	if nomeExe == "" {
		return info, false, nil
	}

	// Só serve o checksum DAQUELE executável: "<nome do exe>.sha256".
	for _, a := range rel.Assets {
		if a.Name == nomeExe+".sha256" {
			info.URLChecksum = a.BrowserDownloadURL
			break
		}
	}
	if info.URLChecksum == "" {
		return info, false, nil
	}
	return info, true, nil
}

// MaisNova compara duas versões no formato "vX.Y.Z" (o "v" é opcional) e
// diz se remota é mais nova que atual. Comparação puramente numérica,
// parte a parte — suficiente pro esquema de versão do projeto (sem
// pré-release/build metadata). Qualquer parte não numérica é tratada como
// 0, então uma versão mal formada nunca aparenta ser "mais nova" por engano.
func MaisNova(atual, remota string) bool {
	a := partesVersao(atual)
	r := partesVersao(remota)
	for i := 0; i < 3; i++ {
		if r[i] != a[i] {
			return r[i] > a[i]
		}
	}
	return false
}

func partesVersao(v string) [3]int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	campos := strings.SplitN(v, ".", 3)
	var partes [3]int
	for i := 0; i < len(campos) && i < 3; i++ {
		n, err := strconv.Atoi(strings.TrimSpace(campos[i]))
		if err != nil {
			continue
		}
		partes[i] = n
	}
	return partes
}

// prefixoArquivo nomeia os executáveis baixados em Downloads — mesmo nome
// base dos assets publicados pelo workflow de release (ver
// .github/workflows/release.yml), então nomeArquivoVersionado("v1.04.06")
// bate exatamente com o nome do asset "FerramentasAssessoria-1.04.06.exe".
const prefixoArquivo = "FerramentasAssessoria-"

// nomeArquivoVersionado monta o nome do executável baixado a partir da tag
// da release, ex.: "v1.04.06" -> "FerramentasAssessoria-1.04.06.exe".
func nomeArquivoVersionado(versao string) string {
	return prefixoArquivo + strings.TrimPrefix(versao, "v") + ".exe"
}

// pastaDownloads localiza a pasta Downloads do usuário (%USERPROFILE%\Downloads
// no Windows) — é onde o executável novo é baixado a cada atualização, ver
// BaixarEAplicar.
func pastaDownloads() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Downloads"), nil
}

// BaixarEAplicar baixa o .exe descrito em info pra dentro de Downloads, com
// nome versionado (nunca sobrescreve o executável em uso — cada atualização
// grava um arquivo novo), confere o checksum e devolve o caminho completo
// do arquivo baixado, pronto pra Relancar.
//
// Isto substitui uma versão anterior que aplicava a atualização IN PLACE
// (biblioteca minio/selfupdate: renomeava o executável em uso, movia o
// novo pro lugar dele). Cada peça funcionava, mas "um processo reescreve o
// próprio binário no disco" é, sozinho, um padrão que antivírus/EDR
// heurísticos associam a dropper/malware auto-atualizável — e a troca em 2
// passos (renomear o atual, mover o novo pro lugar) tinha uma janela onde
// uma interferência externa (ex. antivírus apagando o arquivo novo — às
// vezes de forma ASSÍNCRONA, segundos depois do Apply() já ter retornado
// sucesso) podia deixar nenhum executável no caminho esperado: o atalho
// que apontava pra lá ficava quebrado, e não havia como voltar atrás,
// porque o binário antigo já tinha sido consumido pela troca.
//
// Baixando pra um arquivo NOVO em vez de substituir o que está rodando,
// esse problema desaparece por construção: o executável atual nunca é
// tocado, então mesmo que o arquivo novo seja apagado por um antivírus
// segundos depois, ele continua aí, intacto, como fallback óbvio (ver
// relancar_windows.go). O preço é que o atalho da área de trabalho
// precisa ser reapontado pro arquivo novo a cada atualização — ver
// internal/shortcut, chamado pelo relançador só depois de confirmar que o
// arquivo novo sobreviveu.
func BaixarEAplicar(ctx context.Context, info ReleaseInfo) (string, error) {
	pasta, err := pastaDownloads()
	if err != nil {
		return "", fmt.Errorf("localizar pasta Downloads: %w", err)
	}
	return baixarEGravar(ctx, info, pasta)
}

func baixarEGravar(ctx context.Context, info ReleaseInfo, pastaDestino string) (string, error) {
	checksum, err := baixarChecksum(ctx, info.URLChecksum)
	if err != nil {
		return "", fmt.Errorf("baixar checksum: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, info.URLExe, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "FerramentasAssessoria-Updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("baixar novo executável: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("baixar novo executável: status %d", resp.StatusCode)
	}

	dados, err := io.ReadAll(io.LimitReader(resp.Body, limiteTamanhoExe))
	if err != nil {
		return "", fmt.Errorf("baixar novo executável: %w", err)
	}

	soma := sha256.Sum256(dados)
	if !bytes.Equal(soma[:], checksum) {
		return "", fmt.Errorf("executável baixado com checksum divergente — download incompleto ou corrompido, tente de novo")
	}

	if err := os.MkdirAll(pastaDestino, 0755); err != nil {
		return "", fmt.Errorf("preparar pasta de destino: %w", err)
	}
	caminho := filepath.Join(pastaDestino, nomeArquivoVersionado(info.Versao))
	if err := os.WriteFile(caminho, dados, 0755); err != nil {
		return "", fmt.Errorf("gravar executável baixado: %w", err)
	}
	return caminho, nil
}

// LimparVersoesAntigas apaga, da pasta Downloads, executáveis baixados por
// atualizações anteriores (mesmo prefixo de nome — ver
// nomeArquivoVersionado), preservando o que está em uso agora
// (os.Executable()). Chamada em segundo plano no startup do app (ver
// app.go) — sem isso, Downloads acumularia um .exe (~30MB) a cada
// atualização, pra sempre.
func LimparVersoesAntigas() {
	pasta, err := pastaDownloads()
	if err != nil {
		return
	}
	emUso, err := os.Executable()
	if err != nil {
		return
	}
	limparVersoesAntigas(pasta, emUso)
}

func limparVersoesAntigas(pasta, emUso string) {
	entradas, err := os.ReadDir(pasta)
	if err != nil {
		return
	}
	for _, entrada := range entradas {
		if entrada.IsDir() || !strings.HasPrefix(entrada.Name(), prefixoArquivo) {
			continue
		}
		caminho := filepath.Join(pasta, entrada.Name())
		if strings.EqualFold(caminho, emUso) {
			continue
		}
		_ = os.Remove(caminho) // best-effort — se estiver em uso por outro motivo, tenta de novo na próxima
	}
}

// Relancar reabre o app a partir de caminhoNovo (o executável recém-baixado
// em Downloads — ver BaixarEAplicar), ESPERANDO a instância atual soltar o
// SingleInstanceLock antes de subir a nova, e guardando o caminho do
// executável ATUAL como fallback (ver relancar_windows.go:
// ExecutarSeAjudanteDeRelancamento) — ele nunca é tocado por esta função,
// então continua válido mesmo se caminhoNovo tiver desaparecido quando o
// relançador for agir.
//
// A espera não é detalhe: o app roda com SingleInstanceLock (ver main.go).
// Subir a nova instância imediatamente fazia a nova detectar a antiga ainda
// viva, mandar o "traga a janela pra frente" pra ela e ENCERRAR. Em
// seguida a antiga também encerrava (o chamador chama runtime.Quit logo
// depois), e o resultado era o app sumir da tela e não voltar (bug
// relatado em campo, mais de uma vez).
//
// Por isso quem relança é um processo intermediário — o próprio executável
// ATUAL, reexecutado com uma flag interna reconhecida em main() — que
// espera alguns segundos e só então decide qual executável abrir de
// verdade. Nasce sem janela de console pra não piscar um prompt preto na
// cara do usuário.
func Relancar(caminhoNovo string) error {
	caminhoAtual, err := os.Executable()
	if err != nil {
		return err
	}
	return relancarComEspera(caminhoNovo, caminhoAtual)
}

// baixarChecksum lê o conteúdo do asset .sha256 — convencionalmente o hash
// hexadecimal, opcionalmente seguido de espaço e nome do arquivo (formato
// do `sha256sum`) — e devolve só os bytes do hash.
func baixarChecksum(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FerramentasAssessoria-Updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	dados, err := io.ReadAll(io.LimitReader(resp.Body, limiteTamanhoAPI))
	if err != nil {
		return nil, err
	}
	hexHash, _, _ := strings.Cut(strings.TrimSpace(string(dados)), " ")
	return hex.DecodeString(hexHash)
}
