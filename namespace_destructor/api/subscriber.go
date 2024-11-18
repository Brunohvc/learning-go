package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Função pública para buscar o status de um assinante e converter a resposta para a struct
func GetSubscriberStatus(subscriberUuId string) (SubscriberGetSubscriberStatus, error) {
	// Obtenha a api_key da variável de ambiente
	apiKey := os.Getenv("API_KEY_CS")
	if apiKey == "" {
		log.Fatal("API_KEY_CS não configurada")
	}

	// Crie o objeto JSON a ser enviado
	requestBody := map[string]string{
		"api_key":        apiKey,
		"SubscriberUuId": subscriberUuId,
	}

	// Chame o método Post e obtenha o JSON de resposta
	responseBody, err := Post("https://cs.clinicorp.tech/api/adm/subscriber/get_subscriber_status", requestBody)
	if err != nil {
		return SubscriberGetSubscriberStatus{}, fmt.Errorf("erro ao fazer a requisição POST: %v", err)
	}

	// Decodifique a resposta JSON para a struct
	var subscriberStatus SubscriberGetSubscriberStatus
	err = json.Unmarshal(responseBody, &subscriberStatus)
	if err != nil {
		return SubscriberGetSubscriberStatus{}, fmt.Errorf("erro ao decodificar a resposta JSON: %v", err)
	}

	return subscriberStatus, nil
}
