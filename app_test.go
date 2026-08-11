package main

import (
	"testing"
	"time"
)

func TestDuracaoAteProximaChecagem(t *testing.T) {
	local := time.Local
	casos := []struct {
		nome     string
		agora    time.Time
		esperado time.Time // instante exato em que a checagem deve disparar
	}{
		{
			nome:     "antes das 18h: dispara mais tarde no mesmo dia",
			agora:    time.Date(2026, 8, 11, 9, 30, 0, 0, local),
			esperado: time.Date(2026, 8, 11, 18, 0, 0, 0, local),
		},
		{
			nome:     "depois das 18h: dispara no dia seguinte",
			agora:    time.Date(2026, 8, 11, 20, 0, 0, 0, local),
			esperado: time.Date(2026, 8, 12, 18, 0, 0, 0, local),
		},
		{
			nome:     "exatamente 18h: não dispara de novo agora, só amanhã",
			agora:    time.Date(2026, 8, 11, 18, 0, 0, 0, local),
			esperado: time.Date(2026, 8, 12, 18, 0, 0, 0, local),
		},
		{
			nome:     "vira o mês",
			agora:    time.Date(2026, 8, 31, 19, 0, 0, 0, local),
			esperado: time.Date(2026, 9, 1, 18, 0, 0, 0, local),
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			d := duracaoAteProximaChecagem(c.agora)
			disparoEm := c.agora.Add(d)
			if !disparoEm.Equal(c.esperado) {
				t.Errorf("agora=%v: disparo calculado em %v, esperado %v", c.agora, disparoEm, c.esperado)
			}
			if d <= 0 {
				t.Errorf("duração não pode ser <= 0, veio %v", d)
			}
		})
	}
}
