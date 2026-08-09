package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// appConfig persiste a pasta de PDFs e a base de clientes escolhidas, no
// mesmo caminho e com as mesmas chaves do app original em Python
// (%APPDATA%\RentabilidadeXP\config.json), pra quem já usa o app não perder
// a configuração ao trocar de binário.
type appConfig struct {
	Pasta                string `json:"pasta,omitempty"`
	BaseClientes         string `json:"base_clientes,omitempty"`
	Tema                 string `json:"tema,omitempty"`                  // "claro" (padrão) | "escuro"
	Acento               string `json:"acento,omitempty"`                // "esmeralda" (padrão) | "teal" | "aqua"
	ModoEmail            string `json:"modo_email,omitempty"`            // "padronizado" (padrão) | "livre"
	TabelaPrevidenciaria string `json:"tabela_previdenciaria,omitempty"` // "2026" (padrão) | "2022"
	Visao                string `json:"visao,omitempty"`                 // "cliente" (padrão) | "assessor"
	Fonte                string `json:"fonte,omitempty"`                 // valor CSS font-family; vazio = padrão (Plus Jakarta Sans)
	ModoApresentacao     bool   `json:"modo_apresentacao,omitempty"`     // esconde Rentabilidade/E-mails/Deságio e força Visão Cliente na Compromissada
	ModoFestas           bool   `json:"modo_festas,omitempty"`           // aba Rentabilidade manda mensagem de festas (modelo_festas.txt) pra todo cliente, com ou sem relatório
	Apresentacao         string `json:"apresentacao,omitempty"`          // caminho do arquivo HTML autocontido exibido na aba Apresentação
	AssinaturaAtiva      string `json:"assinatura_ativa,omitempty"`      // nome do arquivo ativo dentro da pasta de assinaturas (ver internal/assinaturas)
	EmailRemetente       string `json:"email_remetente,omitempty"`       // e-mail da conta Outlook usada como remetente nos rascunhos abertos por AbrirEmailNoOutlook; vazio = deixa o Outlook escolher sozinho
	AssessorNome         string `json:"assessor_nome,omitempty"`         // nome do assessor — responde "Nome do assessor responsável" no preenchimento automático do Typeform
	AssessorEmail        string `json:"assessor_email,omitempty"`        // e-mail do assessor — responde "E-mail do assessor" no preenchimento automático do Typeform

	// OrdemNav é a ordem escolhida das ferramentas na barra lateral (IDs de
	// aba, ex.: "rent", "mail", ...). Vazio = ordem padrão do frontend. IDs
	// desconhecidos (aba removida numa atualização) são ignorados; IDs novos
	// que não estejam na lista salva entram no fim, na ordem padrão — ver
	// ordemNavAtual() em main.js.
	OrdemNav []string `json:"ordem_nav,omitempty"`

	// OrdemNavOcultos são os IDs de ferramentas escondidas da barra lateral
	// (botão "-" em Configurações → Ordem da barra lateral). Vazio = nenhuma
	// escondida — ver navItemOculto() em main.js.
	OrdemNavOcultos []string `json:"ordem_nav_ocultos,omitempty"`

	// Recorte personalizado da imagem copiada em "Copiar imagem" (aba
	// Rentabilidade). RecortePersonalizado distingue "nunca configurado"
	// (usa pdfreport.RecorteGraficoRentabilidadePadrao) de "configurado com
	// X0/Y0 = 0" — omitempty nos floats não bastaria pra essa distinção.
	RecortePersonalizado bool    `json:"recorte_personalizado,omitempty"`
	RecorteX0            float64 `json:"recorte_x0,omitempty"`
	RecorteY0            float64 `json:"recorte_y0,omitempty"`
	RecorteX1            float64 `json:"recorte_x1,omitempty"`
	RecorteY1            float64 `json:"recorte_y1,omitempty"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir() // %AppData% no Windows
	if err != nil {
		dir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, "RentabilidadeXP", "config.json"), nil
}

func carregarConfig() appConfig {
	caminho, err := configPath()
	if err != nil {
		return appConfig{}
	}
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return appConfig{}
	}
	var cfg appConfig
	if err := json.Unmarshal(dados, &cfg); err != nil {
		return appConfig{}
	}
	return cfg
}

func salvarConfig(cfg appConfig) error {
	caminho, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(caminho), 0755); err != nil {
		return err
	}
	dados, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(caminho, dados, 0644)
}
