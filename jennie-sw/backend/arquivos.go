package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type InfoArquivo struct {
	Nome     string `json:"nome"`
	Tamanho  int64  `json:"tamanho"` // Em bytes
	IsFolder bool   `json:"is_folder"`
	Data     string `json:"data"`
}

func handleListarArquivos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, `{"erro": "Acesso negado"}`, http.StatusUnauthorized)
		return
	}

	// Extrai o nome do usuário do Token (tira o "token_sessao_")
	usuario := strings.TrimPrefix(token, "token_sessao_")
	if usuario == token {
		http.Error(w, `{"erro": "Token inválido"}`, http.StatusUnauthorized)
		return
	}

	subCaminho := r.URL.Query().Get("path")
	if subCaminho == "" {
		subCaminho = "/"
	}

	// Trava de segurança
	if strings.Contains(subCaminho, "..") {
		http.Error(w, `{"erro": "Caminho inválido ou acesso negado"}`, http.StatusForbidden)
		return
	}

	// Monta o caminho dos arquivos
	caminhoBase := fmt.Sprintf("/mnt/nas/vertigo/%s", usuario)
	caminhoPasta := filepath.Join(caminhoBase, subCaminho)

	// Faz o Linux ler o que tem dentro da pasta
	entradas, err := os.ReadDir(caminhoPasta)

	// Se a pasta ainda não existir, devolve uma lista vazia em vez de erro
	var arquivos []InfoArquivo
	if err != nil {
		arquivos = []InfoArquivo{}
		json.NewEncoder(w).Encode(arquivos)
		return
	}

	// Varre arquivo por arquivo e anota as informações
	for _, entrada := range entradas {
		info, err := entrada.Info()
		if err != nil {
			continue // Se der erro ao ler um arquivo específico, pula pro próximo
		}

		arquivos = append(arquivos, InfoArquivo{
			Nome:     entrada.Name(),
			Tamanho:  info.Size(),
			IsFolder: entrada.IsDir(),
			Data:     info.ModTime().Format("02/01/2006 15:04"), // Formata a data pro padrão brasileiro
		})
	}

	// Garante que o JSON não vá nulo se a pasta estiver vazia
	if arquivos == nil {
		arquivos = []InfoArquivo{}
	}

	// Envia a lista pronta para o navegador
	json.NewEncoder(w).Encode(arquivos)
}

func handleDownloadArquivo(w http.ResponseWriter, r *http.Request) {
	// Recebe o token através de url
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Acesso negado", http.StatusUnauthorized)
		return
	}

	// Captura o usuario pelo token
	usuario := strings.TrimPrefix(token, "token_sessao_")
	if usuario == token {
		http.Error(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	// Recebe o caminho da pasta através da url]
	subCaminho := r.URL.Query().Get("path")
	if subCaminho == "" {
		http.Error(w, "Caminho do arquivo não especificado", http.StatusBadRequest)
		return
	}

	// Verifica se não estão enviando ".." para acessar pastas anteriores
	if strings.Contains(subCaminho, "..") {
		http.Error(w, "Caminho inválido ou acesso negado", http.StatusForbidden)
		return
	}

	// Monta o caminho do arquivo no NAS
	caminhoBase := fmt.Sprintf("/mnt/nas/vertigo/%s", usuario)
	caminhoArquivo := filepath.Join(caminhoBase, subCaminho)

	// Verifica se o caminho realmente existe e se não é uma pasta
	info, err := os.Stat(caminhoArquivo)
	if err != nil || info.IsDir() {
		http.Error(w, "Arquivo não encontrado", http.StatusNotFound)
		return
	}

	// O filepath.Base pega apenas o final do arquivo para enviar para o usuário
	// Naveggador vai salvar abrindo uma janela, ao invés de já baixar tudo de uma vez
	nomeArquivo := filepath.Base(caminhoArquivo)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", nomeArquivo))

	// O go vai ler o disco e soltando o arquivo aos poucos, para não sobrecarregar a ram
	http.ServeFile(w, r, caminhoArquivo)
}

func handleUploadArquivo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	// Verifica o token
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, `{"erro": "Acesso negado"}`, http.StatusUnauthorized)
		return
	}

	// Captura usuario
	usuario := strings.TrimPrefix(token, "token_sessao_")
	if usuario == token {
		http.Error(w, `{"erro": "Token inválido"}`, http.StatusUnauthorized)
		return
	}

	// Limita o processamento da ram a 10mb,
	// O resto vai ser processado direto pro disco, sob demanda
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, `{"erro": "Arquivo grande demais ou corrompido"}`, http.StatusBadRequest)
		return
	}

	// Pega o arquivo enviado pelo frontend
	file, handler, err := r.FormFile("arquivo")
	if err != nil {
		http.Error(w, `{"erro": "Nenhum arquivo encontrado na requisição"}`, http.StatusBadRequest)
		return
	}

	defer file.Close()

	// Pega em qual pasta o arquivo vai ser salvo
	subCaminho := r.URL.Query().Get("path")
	if subCaminho == "" {
		subCaminho = "/"
	}

	// Proibe ".." para evitar o acesso a pastas anteriores
	if strings.Contains(subCaminho, "..") {
		http.Error(w, `{"erro": "Caminho inválido"}`, http.StatusForbidden)
		return
	}

	// Caminho para criar o arquivo
	caminhoBase := fmt.Sprintf("/mnt/nas/vertigo/%s", usuario)
	caminhoArquivo := filepath.Join(caminhoBase, subCaminho, handler.Filename)

	// Cria um arquivo vazio no HD
	arquivoDestino, err := os.Create(caminhoArquivo)
	if err != nil {
		http.Error(w, `{"erro": "Falha ao criar arquivo no disco"}`, http.StatusInternalServerError)
		return
	}
	defer arquivoDestino.Close()

	// Conecta o arquivo de rede direto no arquivo do HD
	_, err = io.Copy(arquivoDestino, file)
	if err != nil {
		http.Error(w, `{"erro": "Falha ao gravar os dados do arquivo"}`, http.StatusInternalServerError)
		return
	}

	// Resposta de sucesso
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"sucesso": true, "mensagem": "Upload finalizado!"}`))
}
