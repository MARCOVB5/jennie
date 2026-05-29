package main

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

// Transforma os dados do bash em jso n
// Servidor web já embutido no go

// Para refinar o texto que o bash retornar

type DadosSistema struct {
	Temperatura      string `json:"temperatura"`
	RamUsada         string `json:"ram_usada"`
	RamTotal         string `json:"ram_total"`
	DiscoPorcentagem string `json:"disco_porcentagem"`
}

// Função que vai ser chamado quando a interface web fazer a requisção. w é a resposta e R é o que a interface web enviou
func handleStatus(w http.ResponseWriter, r *http.Request) {
	// Libera comunicação
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Chama o bash e trata os dados
	cmd := exec.Command("bash", "./coleta_dados.sh")
	saidaBytes, err := cmd.Output()

	resposta := DadosSistema{
		Temperatura: "Erro", RamUsada: "0", RamTotal: "0", DiscoPorcentagem: "0%",
	}

	if err == nil {
		// Refina resposta do bash
		textoArquivo := strings.TrimSpace(string(saidaBytes))
		dados := strings.Split(textoArquivo, ",")

		if len(dados) == 4 {
			resposta.Temperatura = dados[0] + "°C"
			resposta.RamUsada = dados[1] + " MB"
			resposta.RamTotal = dados[2] + " MB"
			resposta.DiscoPorcentagem = dados[3]
		}
	}

	// Retorna para a interface web
	json.NewEncoder(w).Encode(resposta)
}
