package sigvalidator

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/ocsp"
)

const timeoutRevogacao = 10 * time.Second

// checarRevogacao verifica se o certificado do signatário foi revogado, via
// OCSP (quando disponível) e LCR (a partir da extensão CRL Distribution
// Points do próprio certificado — nenhuma LCR fica guardada localmente,
// é baixada na hora). Nunca panica: qualquer falha de rede/parsing vira
// inconclusivo=true, nunca é interpretada como "não revogado".
//
// Conhecimento de domínio: a maioria das ACs da ICP-Brasil só publica LCR,
// sem responder OCSP ativo — cair pra LCR (ou ficar inconclusivo por falta
// de qualquer um dos dois) é o caminho comum, não uma exceção rara.
func checarRevogacao(ctx context.Context, cert, issuer *x509.Certificate) (revogado bool, dataRevogacao time.Time, inconclusivo bool, detalhe string) {
	if issuer == nil {
		return false, time.Time{}, true, "não foi possível localizar o certificado da AC emissora na cadeia local pra conferir OCSP/LCR"
	}

	if confirmado, rev, data, err := checarOCSP(ctx, cert, issuer); err == nil && confirmado {
		if rev {
			return true, data, false, "certificado revogado (confirmado via OCSP)"
		}
		return false, time.Time{}, false, "não revogado (confirmado via OCSP)"
	}

	if len(cert.CRLDistributionPoints) == 0 {
		return false, time.Time{}, true, "certificado não informa endereço de LCR e o OCSP não respondeu"
	}

	var ultimoErro error
	for _, url := range cert.CRLDistributionPoints {
		rev, data, err := checarCRL(ctx, url, cert, issuer)
		if err != nil {
			ultimoErro = err
			continue
		}
		if rev {
			return true, data, false, "certificado revogado (confirmado via LCR)"
		}
		return false, time.Time{}, false, "não revogado (confirmado via LCR)"
	}
	msg := "não foi possível baixar/validar a LCR do emissor"
	if ultimoErro != nil {
		msg += ": " + ultimoErro.Error()
	}
	return false, time.Time{}, true, msg
}

// checarOCSP devolve confirmado=true quando algum responder OCSP deu uma
// resposta assinada e conclusiva (Good ou Revoked). err != nil significa
// "não deu pra usar OCSP" (sem endereço, timeout, resposta malformada) —
// quem chama deve cair pra LCR nesse caso, sem tratar como revogação.
func checarOCSP(ctx context.Context, cert, issuer *x509.Certificate) (confirmado, revogado bool, dataRevogacao time.Time, err error) {
	if len(cert.OCSPServer) == 0 {
		return false, false, time.Time{}, fmt.Errorf("certificado não informa endereço de responder OCSP")
	}
	reqBytes, err := ocsp.CreateRequest(cert, issuer, nil)
	if err != nil {
		return false, false, time.Time{}, err
	}
	var ultimoErro error
	for _, url := range cert.OCSPServer {
		resp, err := enviarOCSP(ctx, url, reqBytes, cert, issuer)
		if err != nil {
			ultimoErro = err
			continue
		}
		switch resp.Status {
		case ocsp.Good:
			return true, false, time.Time{}, nil
		case ocsp.Revoked:
			return true, true, resp.RevokedAt, nil
		default: // ocsp.Unknown — responder existe mas não conhece o certificado
			continue
		}
	}
	if ultimoErro == nil {
		ultimoErro = fmt.Errorf("nenhum responder OCSP deu uma resposta conclusiva")
	}
	return false, false, time.Time{}, ultimoErro
}

func enviarOCSP(ctx context.Context, url string, reqBytes []byte, cert, issuer *x509.Certificate) (*ocsp.Response, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeoutRevogacao)
	defer cancel()
	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodPost, url, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/ocsp-request")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	return ocsp.ParseResponseForCert(body, cert, issuer)
}

func checarCRL(ctx context.Context, url string, cert, issuer *x509.Certificate) (revogado bool, dataRevogacao time.Time, err error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, timeoutRevogacao)
	defer cancel()
	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, url, nil)
	if err != nil {
		return false, time.Time{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, time.Time{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, time.Time{}, fmt.Errorf("%s: status %d", url, resp.StatusCode)
	}
	dados, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return false, time.Time{}, err
	}
	lista, err := x509.ParseRevocationList(dados)
	if err != nil {
		return false, time.Time{}, err
	}
	if err := lista.CheckSignatureFrom(issuer); err != nil {
		return false, time.Time{}, fmt.Errorf("assinatura da LCR não confere com a AC emissora: %w", err)
	}
	for _, entrada := range lista.RevokedCertificateEntries {
		if entrada.SerialNumber.Cmp(cert.SerialNumber) == 0 {
			return true, entrada.RevocationTime, nil
		}
	}
	return false, time.Time{}, nil
}
