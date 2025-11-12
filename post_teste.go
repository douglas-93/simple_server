package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// O número total de requisições que queremos enviar ao servidor.
const totalRequests = 1000

type Data struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
}

func sendPostRequest(id int, url string, wg *sync.WaitGroup) {
	defer wg.Done()

	payload := Data{
		Nome:  fmt.Sprintf("Usuario Teste #%d", id),
		Email: fmt.Sprintf("teste%d@carga.com", id),
	}

	// Serializa a struct para um buffer de bytes JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Worker %04d] Erro ao serializar JSON: %v", id, err)
		return
	}

	// 2. Cria e envia a requisição POST
	resp, err := http.Post(
		url,
		"application/json",
		bytes.NewBuffer(jsonPayload),
	)

	if err != nil {
		// Loga erros de rede ou conexão
		log.Printf("[Worker %d] Erro na requisição HTTP: %v", id, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("[Worker %d] Falha: Status %d", id, resp.StatusCode)
	} else {
		fmt.Printf("[Worker %d] Sucesso: Status %d\n", id, resp.StatusCode)
	}
}

func main() {
	serverURL := "http://localhost:8080/api/dados"

	var wg sync.WaitGroup
	startTime := time.Now()

	fmt.Printf("#### Iniciando Teste de Carga: %d Requisições Concorrentes ####\n", totalRequests)

	for i := 1; i <= totalRequests; i++ {
		wg.Add(1)
		go sendPostRequest(i, serverURL, &wg)
	}

	wg.Wait()

	duration := time.Since(startTime)
	fmt.Printf("\n#### Teste Concluído ####\n")
	fmt.Printf("Total de Requisições: %d\n", totalRequests)
	fmt.Printf("Tempo Total Decorrido: %s\n", duration.Round(time.Millisecond))
}
