package sigvalidator

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"os"
	"time"

	"github.com/smallstep/pkcs7"

	"rentabilidade/internal/icpbrasil"
)

// validarCAdES lê o .p7s/.p7m (destacado ou anexado), monta o
// *pkcs7.PKCS7 com Content populado, e delega o resto pra validarCMS —
// o núcleo de verificação é o mesmo do PAdES (pades.go).
func validarCAdES(ctx context.Context, caminhoP7S, caminhoConteudo string, trust *icpbrasil.Store) Resultado {
	assinaturaBytes, err := os.ReadFile(caminhoP7S)
	if err != nil {
		return Resultado{Estado: EstadoFormatoNaoSuportado, Motivo: "Não foi possível ler o arquivo de assinatura selecionado."}
	}

	p7, err := pkcs7.Parse(assinaturaBytes)
	if err != nil {
		return Resultado{Estado: EstadoFormatoNaoSuportado, Motivo: "O arquivo selecionado não é uma assinatura CAdES/PKCS#7 reconhecível."}
	}

	if caminhoConteudo != "" {
		conteudo, err := os.ReadFile(caminhoConteudo)
		if err != nil {
			return Resultado{Estado: EstadoFormatoNaoSuportado, Motivo: "Não foi possível ler o documento original selecionado."}
		}
		p7.Content = conteudo
	} else if len(p7.Content) == 0 {
		return Resultado{
			Estado: EstadoFormatoNaoSuportado,
			Motivo: "Esta é uma assinatura destacada (o conteúdo não está embutido no arquivo) — selecione também o documento original junto com o .p7s/.p7m.",
		}
	}

	return validarCMS(ctx, p7, trust)
}

// validarCMS é o núcleo compartilhado entre CAdES (cades.go) e PAdES
// (pades.go): a partir de um *pkcs7.PKCS7 já parseado e com Content
// populado (do jeito que for — arquivo separado, embutido, ou faixas de
// bytes de um PDF via /ByteRange), roda a verificação em estágios
// (integridade+assinatura -> cadeia de confiança -> revogação), mais
// extração de identidade e detecção de carimbo de tempo.
//
// Cada estágio só roda se o anterior passou — isso é o que já garante o
// requisito de nunca reportar Valida quando alguma verificação não pôde
// ser concluída (Resultado.Estado == EstadoValida só no fim, quando todos
// os estágios anteriores confirmaram sucesso).
func validarCMS(ctx context.Context, p7 *pkcs7.PKCS7, trust *icpbrasil.Store) Resultado {
	signer := p7.GetOnlySigner()
	if signer == nil {
		return Resultado{
			Estado: EstadoInvalida,
			Motivo: "A assinatura tem mais de um signatário (co-assinatura) — não suportado nesta versão.",
		}
	}

	r := Resultado{
		Signatario: extrairIdentidade(signer),
		ACEmissora: signer.Issuer.CommonName,
	}
	var signingTime time.Time
	if err := p7.UnmarshalSignedAttribute(pkcs7.OIDAttributeSigningTime, &signingTime); err == nil {
		r.DataAssinatura = signingTime
		r.TemDataAssinatura = true
	}
	r.TemCarimboTempo = temCarimboTempo(p7)

	// Estágio 1: integridade do conteúdo + assinatura criptográfica, sem
	// checar cadeia ainda (truststore nil = p7.Verify() só confere que o
	// hash do conteúdo bate e que a assinatura foi feita com a chave
	// privada correspondente ao certificado embutido).
	if err := p7.Verify(); err != nil {
		r.Estado, r.Motivo = classificarErroIntegridade(err)
		r.Verificacoes = []Verificacao{{Nome: "Integridade e assinatura", Passou: false, Detalhe: r.Motivo}}
		return r
	}
	r.Verificacoes = append(r.Verificacoes, Verificacao{Nome: "Integridade e assinatura", Passou: true, Detalhe: "o conteúdo não foi alterado após a assinatura"})

	// Estágio 2: cadeia de confiança até a ICP-Brasil.
	//
	// Verificação manual (em vez de p7.VerifyWithChain) de propósito: a lib
	// envolve o erro do crypto/x509 com fmt.Errorf("...: %v", err) — %v, não
	// %w — o que quebra errors.As e faz todo erro de cadeia cair no caso
	// genérico em vez de ser classificado como ACNaoReconhecida/
	// CertificadoExpirado (confirmado testando contra um PDF assinado
	// real). Verificar direto com signer.Verify devolve o erro do x509 sem
	// esse embrulho.
	var pool *x509.CertPool
	if trust != nil {
		pool = trust.Pool()
	}
	horaVerificacao := time.Now().UTC()
	if r.TemDataAssinatura {
		horaVerificacao = r.DataAssinatura
	}
	intermediarios := x509.NewCertPool()
	for _, c := range p7.Certificates {
		intermediarios.AddCert(c)
	}
	if _, err := signer.Verify(x509.VerifyOptions{
		Roots:         pool,
		Intermediates: intermediarios,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		CurrentTime:   horaVerificacao,
	}); err != nil {
		estado, motivo := classificarErroCadeia(err)
		r.Estado, r.Motivo = estado, motivo
		r.Verificacoes = append(r.Verificacoes, Verificacao{Nome: "Cadeia de confiança ICP-Brasil", Passou: false, Detalhe: motivo})
		return r
	}
	r.Verificacoes = append(r.Verificacoes, Verificacao{Nome: "Cadeia de confiança ICP-Brasil", Passou: true, Detalhe: "certificado encadeia até uma AC reconhecida da ICP-Brasil"})

	// Estágio 3: revogação.
	var certificadosConhecidos []*x509.Certificate
	certificadosConhecidos = append(certificadosConhecidos, p7.Certificates...)
	if trust != nil {
		certificadosConhecidos = append(certificadosConhecidos, trust.Certificados()...)
	}
	issuer := localizarEmissor(signer, certificadosConhecidos)

	revogado, _, inconclusivo, detalheRevogacao := checarRevogacao(ctx, signer, issuer)
	switch {
	case revogado:
		r.Estado = EstadoInvalida
		r.Motivo = "O certificado do signatário foi revogado pela AC emissora."
		r.Verificacoes = append(r.Verificacoes, Verificacao{Nome: "Revogação", Passou: false, Detalhe: detalheRevogacao})
		return r
	case inconclusivo:
		r.Estado = EstadoErroRevogacao
		r.Motivo = "Não foi possível confirmar se o certificado está revogado (" + detalheRevogacao + "). Isso não significa que a assinatura é inválida — só que essa checagem específica não pôde ser concluída agora."
		r.Verificacoes = append(r.Verificacoes, Verificacao{Nome: "Revogação", Passou: false, Detalhe: detalheRevogacao})
		return r
	}
	r.Verificacoes = append(r.Verificacoes, Verificacao{Nome: "Revogação", Passou: true, Detalhe: detalheRevogacao})

	r.Estado = EstadoValida
	r.Verificacoes = append(r.Verificacoes, Verificacao{Nome: "Carimbo de tempo", Passou: r.TemCarimboTempo, Detalhe: detalheCarimbo(r.TemCarimboTempo)})
	return r
}

func detalheCarimbo(presente bool) string {
	if presente {
		return "assinatura tem carimbo de tempo de uma Autoridade de Carimbo do Tempo"
	}
	return "assinatura não tem carimbo de tempo — a data de assinatura, se presente, é autodeclarada"
}

func classificarErroIntegridade(err error) (Estado, string) {
	var digestErr *pkcs7.MessageDigestMismatchError
	if errors.As(err, &digestErr) {
		return EstadoInvalida, "o conteúdo do documento foi alterado após a assinatura (o hash não confere)."
	}
	var signingTimeErr *pkcs7.SigningTimeNotValidError
	if errors.As(err, &signingTimeErr) {
		return EstadoInvalida, "a data de assinatura declarada está fora do período de validade do certificado."
	}
	return EstadoInvalida, "a assinatura criptográfica não pôde ser verificada: " + err.Error()
}

func classificarErroCadeia(err error) (Estado, string) {
	var unknownAuth x509.UnknownAuthorityError
	if errors.As(err, &unknownAuth) {
		return EstadoACNaoReconhecida, "a Autoridade Certificadora que emitiu este certificado não consta na cadeia ICP-Brasil local. Isso pode significar apenas que a cadeia local está desatualizada (uma AC nova foi credenciada recentemente) — não é, por si só, indício de fraude. Tente atualizar a cadeia e validar de novo."
	}
	var certInvalid x509.CertificateInvalidError
	if errors.As(err, &certInvalid) {
		if certInvalid.Reason == x509.Expired {
			return EstadoCertificadoExpirado, "o certificado do signatário está fora do período de validade."
		}
		return EstadoInvalida, "o certificado do signatário é inválido: " + certInvalid.Error()
	}
	return EstadoInvalida, "não foi possível confirmar a cadeia de confiança: " + err.Error()
}

// localizarEmissor procura, entre os certificados embutidos na própria
// assinatura e os da cadeia ICP-Brasil local, aquele cujo Subject bate com
// o Issuer do certificado do signatário. Usado só pra checagem de
// revogação (precisa da chave pública da AC emissora pra validar a
// assinatura da LCR/resposta OCSP) — o nome da AC emissora mostrado na UI
// vem direto de signer.Issuer.CommonName, sem depender disso.
func localizarEmissor(signer *x509.Certificate, candidatos []*x509.Certificate) *x509.Certificate {
	for _, c := range candidatos {
		if bytes.Equal(c.RawSubject, signer.RawIssuer) {
			return c
		}
	}
	return nil
}
