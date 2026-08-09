// Package icpbrasil mantém localmente a cadeia de confiança pública da
// ICP-Brasil (certificados da AC-Raiz e das ACs subordinadas), usada pelo
// pacote sigvalidator pra verificar assinaturas digitais sem depender de
// nenhuma API externa em tempo de validação.
//
// A cadeia é dado público (os mesmos arquivos que qualquer um baixa do
// repositório do ITI) — por isso, ao contrário de resultados de validação,
// ela pode ficar em disco normalmente.
package icpbrasil

import "time"

// Info descreve a cadeia de confiança ICP-Brasil atualmente em uso.
type Info struct {
	AtualizadoEm    time.Time
	Origem          string // "embutida" | "baixada"
	NumCertificados int
}

type metadataArquivo struct {
	AtualizadoEm    string `json:"AtualizadoEm"`
	NumCertificados int    `json:"NumCertificados"`
}
