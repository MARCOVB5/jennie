package main

import (
	"net/http"
)

func main() {
	// Chama a função que está no arquivo autenticacao.go, rota do login
	http.HandleFunc("/api/login", handleLogin)

	// Chama a função que está no arquivo monitoramento.go, rota do dashboard
	http.HandleFunc("/api/status", handleStatus)

	// Chama a função que está no arquivo arquivos.go, rota de ver arquivos
	http.HandleFunc("/api/arquivos", handleListarArquivos)

	// Chama a função que está no arquivo arquivos.go, rota para baixar arquivos
	http.HandleFunc("/api/download", handleDownloadArquivo)

	// Chama a função que está no arquivo arquivos.go, rota para fazer upload de arquivos
	http.HandleFunc("/api/upload", handleUploadArquivo)

	// Chama a função que está no arquivo arquivos.go, rota para deletar arquivos
	http.HandleFunc("/api/excluir", handleExcluirArquivo)

	// Chama a função que está no arquivo arquivos.go, rota para criar nova pasta
	http.HandleFunc("/api/criarpasta", handleCriarPasta)

	// Chama a função que está no arquivo arquivos.go, rota para renomear pasta
	http.HandleFunc("/api/renomear", handleRenomearArquivo)

	// Chama a função que está no arquivo arquivos.go, rota para visualizar arquivo pasta
	http.HandleFunc("/api/visualizar", handleVisualizarArquivo)

	// Configura porta
	http.ListenAndServe(":8080", nil)
}
