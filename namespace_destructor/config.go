package main

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadConfig() {
	// Carrega as variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}
}
