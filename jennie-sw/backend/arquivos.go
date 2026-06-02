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

// Função para fazer a configuração do CORS
func configurarCors(w http.ResponseWriter, r *http.Request, metodos string) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", metodos+", OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return false
	}

	return true
}

// Função para verificar se o token e o usuário existem e estão corretos
func verificarTokenUsuario(w http.ResponseWriter, r *http.Request) (string, bool) {
	// Tenta primeiro pegar do cabeçalho
	token := r.Header.Get("Authorization")

	// Se não achar no cabeçalho, procurar na url
	if token == "" {
		token = r.URL.Query().Get("token")
	}

	// Erro, token não existe
	if token == "" {
		http.Error(w, `{"erro": "Acesso negado"}`, http.StatusUnauthorized)
		return "", false
	}

	usuario := strings.TrimPrefix(token, "token_sessao_")
	if usuario == token || usuario == "" {
		http.Error(w, `{"erro": "Token inválido"}`, http.StatusUnauthorized)
		return "", false
	}

	return usuario, true
}

// Método para capturar o subCaminho através da URL e verifica se está acontecendo uma tentativa de voltar pastas
func capturaVerificaSubCaminho(w http.ResponseWriter, r *http.Request) (string, bool) {
	subCaminho := r.URL.Query().Get("path")
	if subCaminho == "" {
		subCaminho = "/"
	}

	// Trava de segurança para não voltar pastas
	if strings.Contains(subCaminho, "..") {
		http.Error(w, `{"erro": "Caminho inválido ou acesso negado"}`, http.StatusForbidden)
		return "", false
	}

	return subCaminho, true
}

func handleListarArquivos(w http.ResponseWriter, r *http.Request) {
	// Configuração CORS
	if !configurarCors(w, r, "GET") {
		return
	}

	// Verifica se o token e o usuario existem
	usuario, ok := verificarTokenUsuario(w, r)
	if !ok {
		return
	}

	// Pega e verifica o subCaminho
	subCaminho, ok := capturaVerificaSubCaminho(w, r)
	if !ok {
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
	// Verifica se o token e o usuario existem
	usuario, ok := verificarTokenUsuario(w, r)
	if !ok {
		return
	}

	// Pega e verifica o subCaminho
	subCaminho, ok := capturaVerificaSubCaminho(w, r)
	if !ok {
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
	// Configuração CORS
	if !configurarCors(w, r, "POST") {
		return
	}

	// Verifica se o token e o usuario existem
	usuario, ok := verificarTokenUsuario(w, r)
	if !ok {
		return
	}

	// Pega e verifica o subCaminho
	subCaminho, ok := capturaVerificaSubCaminho(w, r)
	if !ok {
		return
	}

	// Cria um leitor da rede direto para o disco, sempre precisar passar pela RAM
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, `{"erro": "Falha ao iniciar leitura do upload"}`, http.StatusBadRequest)
		return
	}

	// Loop para ler as partes da requisição HTTP
	for {
		parte, err := reader.NextPart()
		if err == io.EOF {
			break // Chegou no fim do upload
		}
		if err != nil {
			http.Error(w, `{"erro": "Erro durante a transferência"}`, http.StatusInternalServerError)
			return
		}

		// Se a parte atual for um arquivo, nós a salvamos
		if parte.FileName() != "" {
			caminhoBase := fmt.Sprintf("/mnt/nas/vertigo/%s", usuario)
			caminhoArquivo := filepath.Join(caminhoBase, subCaminho, parte.FileName())

			// Cria o arquivo final no disco
			arquivoDestino, err := os.Create(caminhoArquivo)
			if err != nil {
				http.Error(w, `{"erro": "Falha ao criar arquivo no disco"}`, http.StatusInternalServerError)
				return
			}

			// io.Copy puxa os bytes da parte (rede) e joga no arquivoDestino (HD)
			// Ele faz isso em blocos de 32KB, minimizando o uso de ram.
			_, err = io.Copy(arquivoDestino, parte)

			// Fecha o arquivo manualmente após a cópia
			arquivoDestino.Close()

			if err != nil {
				http.Error(w, `{"erro": "Falha ao gravar os dados no disco"}`, http.StatusInternalServerError)
				return
			}

			break
		}
	}

	// Resposta de sucesso
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"sucesso": true, "mensagem": "Upload finalizado!"}`))
}

func handleExcluirArquivo(w http.ResponseWriter, r *http.Request) {
	// Configuração CORS
	if !configurarCors(w, r, "DELETE") {
		return
	}

	// Verifica se o token e o usuario existem
	usuario, ok := verificarTokenUsuario(w, r)
	if !ok {
		return
	}

	// Pega e verifica o subCaminho
	subCaminho, ok := capturaVerificaSubCaminho(w, r)
	if !ok {
		return
	}

	// Impede de excluir a raiz
	if subCaminho == "/" {
		http.Error(w, `{"erro": "Ação proibida: Não é possível excluir o diretório raiz"}`, http.StatusForbidden)
		return
	}

	// Diretorio da partição do usuário
	caminhoBase := fmt.Sprintf("/mnt/nas/vertigo/%s", usuario)
	caminhoCompleto := filepath.Join(caminhoBase, subCaminho)

	// Remove o arquivo ou pasta
	err := os.RemoveAll(caminhoCompleto)
	if err != nil {
		http.Error(w, `{"erro": "Falha ao excluir o arquivo ou pasta"}`, http.StatusInternalServerError)
		return
	}

	// Retorna ok para o frontend
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"sucesso": true, "mensagem": "Excluído com sucesso!"}`))
}

func handleCriarPasta(w http.ResponseWriter, r *http.Request) {
	// Configuração CORS
	if !configurarCors(w, r, "POST") {
		return
	}

	// Verifica se o token e o usuario existem
	usuario, ok := verificarTokenUsuario(w, r)
	if !ok {
		return
	}

	// Pega e verifica o subCaminho
	subCaminho, ok := capturaVerificaSubCaminho(w, r)
	if !ok {
		return
	}

	// Monta o caminho para acessar o servidor
	caminhoBase := fmt.Sprintf("/mnt/nas/vertigo/%s", usuario)
	caminhoCompleto := filepath.Join(caminhoBase, subCaminho)

	// Cria a nova pasta
	err := os.MkdirAll(caminhoCompleto, 0755)
	if err != nil {
		http.Error(w, `{"erro": "Falha ao criar a pasta no disco"}`, http.StatusInternalServerError)
		return
	}

	// Retorna para o front end
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"sucesso": true, "mensagem": "Pasta criada com sucesso!"}`))
}

// Método para renomear arquivo
func handleRenomearArquivo(w http.ResponseWriter, r *http.Request) {
	// Configuração CORS
	if !configurarCors(w, r, "POST") {
		return
	}

	// Verifica token
	usuario, ok := verificarTokenUsuario(w, r)
	if !ok {
		return
	}

	// Pega o subcaminho e fazer verificaçãoo para não voltar pastas
	subCaminho, ok := capturaVerificaSubCaminho(w, r)
	if !ok {
		return
	}

	// Verificação para não deixar renomear a pasta raiz
	if subCaminho == "/" {
		http.Error(w, `{"erro": "Não é possível renomear o diretório raiz"}`, http.StatusForbidden)
		return
	}

	// Pega o novo nome desejado via URL
	novoNome := r.URL.Query().Get("novoNome")

	// Verificações no novo nome
	if novoNome == "" || strings.Contains(novoNome, "..") || strings.Contains(novoNome, "/") {
		http.Error(w, `{"erro": "Novo nome inválido"}`, http.StatusBadRequest)
		return
	}

	// Monta o caminho completo
	caminhoBase := fmt.Sprintf("/mnt/nas/vertigo/%s", usuario)
	caminhoAntigo := filepath.Join(caminhoBase, subCaminho)

	// Pega só a pasta onde o arquivo está, e junta com o novo nome
	diretorioAtual := filepath.Dir(caminhoAntigo)
	caminhoNovo := filepath.Join(diretorioAtual, novoNome)

	// Renomeia / Move
	err := os.Rename(caminhoAntigo, caminhoNovo)
	if err != nil {
		http.Error(w, `{"erro": "Falha ao renomear arquivo ou pasta"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"sucesso": true, "mensagem": "Renomeado com sucesso!"}`))
}

// Método para fazer a visualização de um arquivo
func handleVisualizarArquivo(w http.ResponseWriter, r *http.Request) {
	// Verifica token
	usuario, ok := verificarTokenUsuario(w, r)
	if !ok {
		return
	}

	// Pega o subcaminho e fazer verificaçãoo para não voltar pastas
	subCaminho, ok := capturaVerificaSubCaminho(w, r)
	if !ok {
		return
	}

	// Monta o caminho do arquivo
	caminhoBase := fmt.Sprintf("/mnt/nas/vertigo/%s", usuario)
	caminhoArquivo := filepath.Join(caminhoBase, subCaminho)

	// Verifica se existe
	info, err := os.Stat(caminhoArquivo)
	if err != nil || info.IsDir() {
		http.Error(w, "Arquivo não encontrado", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, caminhoArquivo)
}
