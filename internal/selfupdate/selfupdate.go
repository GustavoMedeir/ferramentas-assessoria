// Package selfupdate checa releases publicadas num repositório público do
// GitHub e, quando há uma versão mais nova, baixa o novo executável e
// substitui o binário em uso — sem instalador, sem intervenção manual do
// assessor além de clicar em "Atualizar agora" no app.
package selfupdate

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/minio/selfupdate"
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
	for _, a := range rel.Assets {
		switch {
		case strings.HasSuffix(a.Name, ".exe"):
			info.URLExe = a.BrowserDownloadURL
		case strings.HasSuffix(a.Name, ".sha256"):
			info.URLChecksum = a.BrowserDownloadURL
		}
	}
	if info.URLExe == "" || info.URLChecksum == "" {
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

// BaixarEAplicar baixa o .exe e o checksum descritos em info, confere um
// contra o outro e substitui o executável em uso pelo novo. Se falhar
// depois de já ter mexido no binário, tenta reverter sozinho (comportamento
// da lib) — o chamador só precisa decidir o que dizer ao usuário.
func BaixarEAplicar(ctx context.Context, info ReleaseInfo) error {
	checksum, err := baixarChecksum(ctx, info.URLChecksum)
	if err != nil {
		return fmt.Errorf("baixar checksum: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, info.URLExe, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "FerramentasAssessoria-Updater")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("baixar novo executável: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("baixar novo executável: status %d", resp.StatusCode)
	}

	err = selfupdate.Apply(io.LimitReader(resp.Body, limiteTamanhoExe), selfupdate.Options{
		Checksum: checksum,
	})
	if err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("atualização falhou e não foi possível restaurar a versão anterior — reinstale manualmente: %w", rerr)
		}
		return fmt.Errorf("aplicar atualização: %w", err)
	}
	return nil
}

// Relancar inicia uma nova instância do executável atual — depois de
// BaixarEAplicar, isso já é o binário novo em disco — como processo
// separado e desanexado. O chamador ainda precisa encerrar o processo
// corrente (ex. runtime.Quit do Wails) logo em seguida: com
// SingleInstanceLock ativo, a nova instância só consegue abrir a janela
// dela depois que a antiga soltar o lock.
func Relancar() error {
	caminho, err := os.Executable()
	if err != nil {
		return err
	}
	return exec.Command(caminho, os.Args[1:]...).Start()
}

// baixarChecksum lê o conteúdo do asset .sha256 — convencionalmente o hash
// hexadecimal, opcionalmente seguido de espaço e nome do arquivo (formato
// do `sha256sum`) — e devolve só os bytes do hash, prontos pra
// selfupdate.Options.Checksum.
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
