package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Data struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
}

type MessageResponse struct {
	Mensagem string `json:"mensagem"`
}

var storage = []Data{
	{Nome: "Exemplo", Email: "exemplo@dominio.com"},
}

// Mutex para proteger o acesso ao 'storage' em ambiente concorrente
var mu sync.Mutex

func main() {
	port := ":8080"

	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/api/dados", apiDadosHandler)

	fmt.Printf("Servidor escutando em http://127.0.0.1%s\n", port)

	log.Fatal(http.ListenAndServe(port, nil))
}

// ---------------- HANDLERS ----------------

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido.", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Bem-vindo ao Servidor JSON em Go!"))
}

// apiDadosHandler lida com as requisições GET e POST
func apiDadosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	default:
		// Retorna 405 Method Not Allowed para outros métodos
		http.Error(w, "Método não suportado para esta API.", http.StatusMethodNotAllowed)
	}
}

// handleGet transforma a slice 'storage' em JSON e envia a resposta
func handleGet(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Cria um codificador JSON que escreve diretamente na resposta HTTP
	if err := json.NewEncoder(w).Encode(storage); err != nil {
		log.Printf("Erro ao codificar resposta JSON: %v", err)
		http.Error(w, "Erro interno ao gerar JSON", http.StatusInternalServerError)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	var receivedData Data

	/*
	 * 1. Decodificação (Unmarshal)
	 * Limita o tamanho do corpo para evitar ataques de DOS
	 */
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limite

	// Cria um decodificador JSON que lê diretamente do corpo da requisição
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Opção de segurança: falha se houver campos desconhecidos

	if err := decoder.Decode(&receivedData); err != nil {
		if strings.HasPrefix(err.Error(), "json: unknown field") {
			http.Error(w, fmt.Sprintf("Corpo da requisição inválido: %v", err), http.StatusBadRequest)
		} else if err == io.EOF {
			http.Error(w, "O corpo da requisição não pode estar vazio", http.StatusBadRequest)
		} else {
			http.Error(w, fmt.Sprintf("Erro ao processar JSON: %v", err), http.StatusBadRequest)
		}
		return
	}

	// 2. Armazenamento (Thread-safe)
	mu.Lock()
	storage = append(storage, receivedData)
	mu.Unlock()

	// 3. Resposta de Confirmação (Codificação)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := MessageResponse{
		Mensagem: fmt.Sprintf("Dados de %s recebidos com sucesso!", receivedData.Nome),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Erro ao codificar resposta de sucesso: %v", err)
	}
}
