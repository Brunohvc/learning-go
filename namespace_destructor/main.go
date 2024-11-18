package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"namespace_destructor/clone_data"
	"namespace_destructor/delete_data"
	"namespace_destructor/get_data"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

const (
	projectDevId  = "dev-clinicorp"
	projectProdId = "clinicorp-solution"
	maxTables     = 4
	colorGreen    = "\033[32m"
	colorRed      = "\033[31m"
	colorReset    = "\033[0m"
)

func main() {
	LoadConfig()
	// startProcessToDeleteNamespaces()
	generateData()
	// startProcessToCloneData()
	// subscriber, err := api.GetSubscriberStatus("ortocoi")
	// if err != nil {
	// 	log.Fatalf("Erro ao obter o status do assinante: %v", err)
	// }
	// fmt.Printf("Status do assinante: %+v\n", subscriber)
}

func startProcessToCloneData() {
	clone_data.CloneData(context.Background(), projectProdId, projectDevId, "CLINICORP_DEFAULTS")
}

func generateData() {
	// Cria o client do Datastore uma vez e o reutiliza
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectDevId)
	if err != nil {
		log.Fatalf("Falha ao criar o cliente do Datastore: %v", err)
	}
	defer client.Close()

	// Lista todos os namespaces e salva em um arquivo
	get_data.ListNamespaces(ctx, client)
	// get_data.ListBackupNamespaces(ctx, client)
	// get_data.CalculateTotalStorageFromFile(ctx, client, "backup.txt")
	// get_data.CheckImagesFromThumb(ctx, client, "laura.br.sp.sertaozinho")
}

func startProcessToDeleteNamespaces() {
	// Carrega os namespaces do arquivo namespaces.txt
	namespaces, err := loadNamespacesFromFile("namespaces.txt")
	if err != nil {
		log.Fatalf("Falha ao carregar namespaces do arquivo: %v", err)
	}

	// Carrega os namespaces seguros do arquivo safeNamespaces.txt
	safeNamespaces, err := loadNamespacesFromFile("safeNamespaces.txt")
	if err != nil {
		log.Fatalf("Falha ao carregar namespaces seguros do arquivo: %v", err)
	}

	// Se não houver namespaces, exibe uma mensagem e termina o processo
	if len(namespaces) == 0 {
		fmt.Printf("%sNenhum namespace encontrado para destruição.%s\n", colorRed, colorReset)
		time.Sleep(10 * time.Second) // Pausa antes de reiniciar
		return
	}

	// Cria o client do Datastore uma vez e o reutiliza
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectDevId)
	if err != nil {
		log.Fatalf("Falha ao criar o cliente do Datastore: %v", err)
	}
	defer client.Close()

	// Executa a deleção para cada namespace de forma sequencial
	for _, namespace := range namespaces {
		// Ignora namespaces que estão na lista de namespaces seguros
		if isNamespaceSafe(namespace, safeNamespaces) {
			fmt.Printf("%sNamespace %s está na lista de safeNamespaces e será ignorado.%s\n", colorGreen, namespace, colorReset)
			if err := removeNamespaceFromFile("namespaces.txt", namespace); err != nil {
				fmt.Printf("Erro ao remover o namespace %s do arquivo: %v\n", namespace, err)
			}
			continue
		}

		// Verifica se o namespace possui tabelas antes de iniciar o processo
		allKinds, err := get_data.ListKinds(ctx, client, namespace)
		if err != nil {
			fmt.Printf("Erro ao listar kinds para o namespace %s: %v\n", namespace, err)
			continue
		}

		if len(allKinds) == 0 {
			fmt.Printf("%sNenhuma tabela encontrada no namespace %s. Removendo do arquivo.%s\n", colorGreen, namespace, colorReset)
			if err := removeNamespaceFromFile("namespaces.txt", namespace); err != nil {
				fmt.Printf("Erro ao remover o namespace %s do arquivo: %v\n", namespace, err)
			}
			continue
		}

		fmt.Printf("Iniciando processo para o namespace: %s\n", namespace)
		startDeleteData(ctx, client, namespace, allKinds)
	}

	// Ao finalizar o processamento, reinicia o main
	fmt.Println("Processamento completo para todos os namespaces. Reiniciando o processo...")
	time.Sleep(10 * time.Second)     // Pausa antes de reiniciar
	startProcessToDeleteNamespaces() // Reinicia o processo
}

// Função para carregar namespaces a partir de um arquivo
func loadNamespacesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir o arquivo %s: %v", filename, err)
	}
	defer file.Close()

	var namespaces []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		namespace := scanner.Text()
		if namespace != "" {
			namespaces = append(namespaces, namespace)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo %s: %v", filename, err)
	}

	return namespaces, nil
}

// Função para verificar se um namespace está na lista de namespaces seguros
func isNamespaceSafe(namespace string, safeNamespaces []string) bool {
	for _, safeNamespace := range safeNamespaces {
		if namespace == safeNamespace {
			return true
		}
	}
	return false
}

// Função para remover um namespace do arquivo namespaces.txt
func removeNamespaceFromFile(filename, namespace string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("falha ao ler o arquivo %s: %v", filename, err)
	}

	// Remove o namespace da lista
	lines := strings.Split(string(file), "\n")
	var newLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != namespace {
			newLines = append(newLines, line)
		}
	}

	// Escreve de volta no arquivo
	err = os.WriteFile(filename, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("falha ao escrever no arquivo %s: %v", filename, err)
	}

	return nil
}

func startDeleteData(ctx context.Context, client *datastore.Client, namespace string, allKinds []string) {
	var kinds []delete_data.KindInfo
	for _, kind := range allKinds {
		kinds = append(kinds, delete_data.KindInfo{Kind: kind})
	}

	// Filtra kinds sem underscores "__"
	var kindsWithoutUnderscore []delete_data.KindInfo
	for _, kind := range kinds {
		if !(kind.Kind[:2] == "__" && kind.Kind[len(kind.Kind)-2:] == "__") {
			kindsWithoutUnderscore = append(kindsWithoutUnderscore, kind)
		}
	}

	fmt.Printf("Iniciando deleção de dados para o namespace %s...\n", namespace)

	// Executa a deleção das tabelas com limite de 4 simultâneas
	delete_data.DeleteData(ctx, client, kindsWithoutUnderscore, namespace, maxTables)

	fmt.Printf("Processo de deleção completo para o namespace %s.\n", namespace)
}
