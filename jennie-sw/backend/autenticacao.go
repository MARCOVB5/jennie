package main

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

type Credenciais struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RespostaLogin struct {
	Sucesso  bool   `json:"sucesso"`
	Mensagem string `json:"mensagem"`
	Token    string `json:"token"`
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	var resposta RespostaLogin
	var creds Credenciais

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		resposta = RespostaLogin{Sucesso: false, Mensagem: "Erro ao ler dados"}
		json.NewEncoder(w).Encode(resposta)
		return
	}

	cmd := exec.Command("bash", "./verificar_login.sh", creds.Username, creds.Password)
	saidaBytes, err := cmd.Output()

	if err != nil {
		resposta = RespostaLogin{Sucesso: false, Mensagem: "Erro na verificação"}
		json.NewEncoder(w).Encode(resposta)
		return
	}

	resultado := strings.TrimSpace(string(saidaBytes))

	if resultado == "sucesso" {
		resposta = RespostaLogin{Sucesso: true, Mensagem: "Login efetuado com sucesso!", Token: "token_sessao_" + creds.Username}
	} else {
		resposta = RespostaLogin{Sucesso: false, Mensagem: "Usuário ou senha incorretos!"}
	}

	json.NewEncoder(w).Encode(resposta)
}
