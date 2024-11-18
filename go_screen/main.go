package main

import (
	"fmt"
	"log"

	"github.com/zserge/lorca"
)

func main() {
	// Tenta criar a interface Lorca e captura qualquer erro
	ui, err := lorca.New("http://localhost:3000", "", 800, 600)
	if err != nil {
		log.Fatalf("Erro ao iniciar o Lorca: %v", err)
	}
	defer ui.Close()

	// Loga uma mensagem para confirmar que a janela foi aberta
	fmt.Println("Janela aberta com sucesso!")

	<-ui.Done()
}
