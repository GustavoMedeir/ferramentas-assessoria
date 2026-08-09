package icpbrasil

import (
	"context"
	"crypto/x509"
	"embed"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

//go:embed snapshot/icpbrasil_snapshot.pem snapshot/metadata.json
var embutido embed.FS

const (
	nomeArquivoPEM  = "icpbrasil_snapshot.pem"
	nomeArquivoMeta = "metadata.json"

	// intervaloChecagem é de quanto em quanto tempo IniciarChecagemPeriodica
	// acorda pra decidir se vale a pena tentar atualizar — não é a
	// frequência real de download, essa é limitada por idadeMaxima.
	intervaloChecagem = 24 * time.Hour
	// idadeMaxima é "no máximo 1x/mês" do pedido original: só tenta baixar
	// de novo quando a cadeia em uso já teria mais de 30 dias.
	idadeMaxima = 30 * 24 * time.Hour
)

// Store guarda a cadeia de confiança ICP-Brasil em memória, com troca segura
// enquanto uma goroutine de atualização em background pode estar escrevendo
// (primeira goroutine do projeto — em outros pacotes só há sync.Once/
// sync/atomic.Bool pra lazy-init e reentrância, nunca um worker de fundo).
type Store struct {
	pool  atomic.Pointer[x509.CertPool]
	certs atomic.Pointer[[]*x509.Certificate] // CertPool não permite enumerar; precisamos achar a AC emissora de um certificado pra checar revogação
	info  atomic.Pointer[Info]

	pasta       string
	atualizando atomic.Bool // trava reentrância, mesmo padrão de a.processando em app.go
}

// NovoStore carrega a cadeia salva em pastaDados; se não existir ou estiver
// corrompida, cai pro snapshot embutido no binário (nunca falha "vazio" —
// o app tem que funcionar já na primeira execução, offline).
func NovoStore(pastaDados string) (*Store, error) {
	s := &Store{pasta: pastaDados}

	if pemBytes, atualizadoEm, err := carregarDeDisco(pastaDados); err == nil {
		if pool, certs, ok := montarPool(pemBytes); ok {
			s.aplicar(pool, certs, Info{AtualizadoEm: atualizadoEm, Origem: "baixada", NumCertificados: len(certs)})
			return s, nil
		}
	}

	pemBytes, err := embutido.ReadFile("snapshot/" + nomeArquivoPEM)
	if err != nil {
		return nil, fmt.Errorf("snapshot embutido ilegível: %w", err)
	}
	pool, certs, ok := montarPool(pemBytes)
	if !ok {
		return nil, errors.New("snapshot embutido não contém nenhum certificado válido")
	}
	atualizadoEm := time.Time{}
	if metaBytes, err := embutido.ReadFile("snapshot/" + nomeArquivoMeta); err == nil {
		var m metadataArquivo
		if json.Unmarshal(metaBytes, &m) == nil {
			if t, err := time.Parse(time.RFC3339, m.AtualizadoEm); err == nil {
				atualizadoEm = t
			}
		}
	}
	s.aplicar(pool, certs, Info{AtualizadoEm: atualizadoEm, Origem: "embutida", NumCertificados: len(certs)})
	return s, nil
}

func (s *Store) aplicar(pool *x509.CertPool, certs []*x509.Certificate, info Info) {
	s.pool.Store(pool)
	s.certs.Store(&certs)
	s.info.Store(&info)
}

// Pool devolve o conjunto de certificados de confiança pra checagem de
// cadeia (pkcs7.VerifyWithChain). Contém tanto certificados da AC-Raiz
// quanto das ACs subordinadas — não é um problema misturar os dois num
// único CertPool "raiz": o crypto/x509 aceita qualquer certificado do pool
// como âncora de confiança válida, esteja ele autoassinado ou não.
func (s *Store) Pool() *x509.CertPool {
	return s.pool.Load()
}

// Certificados devolve todos os certificados carregados — usado pra
// localizar a AC emissora de um certificado de signatário (nome pro DTO,
// checagem de revogação).
func (s *Store) Certificados() []*x509.Certificate {
	if p := s.certs.Load(); p != nil {
		return *p
	}
	return nil
}

// Info devolve a data/origem da cadeia atualmente em uso.
func (s *Store) Info() Info {
	if i := s.info.Load(); i != nil {
		return *i
	}
	return Info{}
}

// AtualizarAgora força uma atualização síncrona da cadeia a partir dos
// servidores do ITI. Chamada explícita do usuário (botão "Atualizar" na
// UI) ou pela checagem periódica em background.
func (s *Store) AtualizarAgora(ctx context.Context) (Info, error) {
	if !s.atualizando.CompareAndSwap(false, true) {
		return s.Info(), errors.New("uma atualização da cadeia ICP-Brasil já está em andamento")
	}
	defer s.atualizando.Store(false)

	pemBytes, _, err := baixarCadeiaCompleta(ctx)
	if err != nil {
		return s.Info(), err
	}
	pool, certs, ok := montarPool(pemBytes)
	if !ok {
		return s.Info(), errors.New("a cadeia baixada não contém nenhum certificado válido")
	}

	agora := time.Now().UTC()
	if s.pasta != "" {
		// Melhor esforço: falha ao persistir não invalida a atualização que
		// já está válida em memória pro resto da sessão.
		_ = salvarEmDisco(s.pasta, pemBytes, agora, len(certs))
	}

	info := Info{AtualizadoEm: agora, Origem: "baixada", NumCertificados: len(certs)}
	s.aplicar(pool, certs, info)
	return info, nil
}

// IniciarChecagemPeriodica dispara, em background, a checagem de "a cadeia
// local já passou de 30 dias?" — nunca bloqueia quem chamou, e qualquer
// falha de rede fica só registrada em Info (a validação continua com a
// cadeia local normalmente). Encerra sozinha quando ctx é cancelado
// (shutdown do Wails).
func (s *Store) IniciarChecagemPeriodica(ctx context.Context) {
	go s.loopChecagem(ctx)
}

func (s *Store) loopChecagem(ctx context.Context) {
	s.talvezAtualizar(ctx)

	ticker := time.NewTicker(intervaloChecagem)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.talvezAtualizar(ctx)
		}
	}
}

func (s *Store) talvezAtualizar(ctx context.Context) {
	if time.Since(s.Info().AtualizadoEm) < idadeMaxima {
		return
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	_, _ = s.AtualizarAgora(ctxTimeout)
}

func carregarDeDisco(pasta string) (pemBytes []byte, atualizadoEm time.Time, err error) {
	if pasta == "" {
		return nil, time.Time{}, errors.New("pasta de dados não configurada")
	}
	pemBytes, err = os.ReadFile(filepath.Join(pasta, nomeArquivoPEM))
	if err != nil {
		return nil, time.Time{}, err
	}
	if raw, err := os.ReadFile(filepath.Join(pasta, nomeArquivoMeta)); err == nil {
		var m metadataArquivo
		if json.Unmarshal(raw, &m) == nil {
			if t, err := time.Parse(time.RFC3339, m.AtualizadoEm); err == nil {
				atualizadoEm = t
			}
		}
	}
	return pemBytes, atualizadoEm, nil
}

func salvarEmDisco(pasta string, pemBytes []byte, atualizadoEm time.Time, numCertificados int) error {
	if err := os.MkdirAll(pasta, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(pasta, nomeArquivoPEM), pemBytes, 0644); err != nil {
		return err
	}
	meta := metadataArquivo{
		AtualizadoEm:    atualizadoEm.Format(time.RFC3339),
		NumCertificados: numCertificados,
	}
	raw, err := json.MarshalIndent(meta, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(pasta, nomeArquivoMeta), raw, 0644)
}

// montarPool decodifica um bundle PEM com um ou mais certificados. Um bloco
// individual corrompido não derruba os demais — só é ignorado.
func montarPool(pemBytes []byte) (*x509.CertPool, []*x509.Certificate, bool) {
	pool := x509.NewCertPool()
	var certs []*x509.Certificate
	rest := pemBytes
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		pool.AddCert(cert)
		certs = append(certs, cert)
	}
	return pool, certs, len(certs) > 0
}
