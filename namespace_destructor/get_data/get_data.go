package get_data

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// fetchNamespaces realiza a query no Datastore para listar namespaces em lotes de 100 usando cursores.
func fetchNamespaces(ctx context.Context, client *datastore.Client) ([]string, error) {
	var namespaces []string
	var cursor *datastore.Cursor
	count := 0
	for {
		query := datastore.NewQuery("__namespace__").KeysOnly().Limit(100)
		if cursor != nil {
			query = query.Start(*cursor)
		}

		it := client.Run(ctx, query)
		batchNamespaces := []string{}

		for {
			key, err := it.Next(nil)
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("falha ao listar namespaces: %v", err)
			}

			// Adiciona o nome do namespace à lista, verificando se não é um namespace vazio (namespace padrão)
			if key.Name != "" {
				batchNamespaces = append(batchNamespaces, key.Name)
			}
		}

		count += len(batchNamespaces)
		fmt.Printf("Namespaces listados: %d\n", count)

		if len(batchNamespaces) == 0 {
			// Se não há mais registros no batch atual, finaliza a busca
			break
		}

		namespaces = append(namespaces, batchNamespaces...)

		// Obtém o próximo cursor
		nextCursor, err := it.Cursor()
		if err != nil {
			return nil, fmt.Errorf("falha ao obter cursor: %v", err)
		}
		if nextCursor.String() == "" {
			// Se o cursor não é válido, a busca foi finalizada
			break
		}
		cursor = &nextCursor
	}

	fmt.Printf("Total de namespaces listados: %d\n", count)
	return namespaces, nil
}

// ListKinds lista todos os kinds em um namespace especificado, chamando a função fetchKinds.
func ListKinds(ctx context.Context, client *datastore.Client, namespace string) ([]string, error) {
	fmt.Printf("Listando kinds no namespace: %s...\n", namespace)
	kinds, err := fetchKinds(ctx, client, namespace)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Kinds listados no namespace %s: %v\n", namespace, kinds)
	return kinds, nil
}

// fetchKinds realiza a query no Datastore para listar todos os kinds em um namespace,
// ignorando aqueles que começam e terminam com "__".
func fetchKinds(ctx context.Context, client *datastore.Client, namespace string) ([]string, error) {
	query := datastore.NewQuery("__kind__").Namespace(namespace).KeysOnly()
	var kinds []string
	it := client.Run(ctx, query)

	for {
		key, err := it.Next(nil)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("falha ao listar kinds: %v", err)
		}

		// Ignora kinds que começam e terminam com "__"
		if strings.HasPrefix(key.Name, "__") && strings.HasSuffix(key.Name, "__") {
			continue
		}
		kinds = append(kinds, key.Name)
	}
	return kinds, nil
}

// ListNamespaces lista todos os namespaces no Datastore e salva em um arquivo "todos.txt".
func ListNamespaces(ctx context.Context, client *datastore.Client) error {
	fmt.Println("Listando namespaces...")
	namespaces, err := fetchNamespaces(ctx, client)
	if err != nil {
		return err
	}

	// Abre o arquivo para escrita
	fmt.Println("Salvando namespaces em 'todos.txt'...")
	file, err := os.Create("todos.txt")
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo: %v", err)
	}
	defer file.Close()

	// Escreve cada namespace em uma linha do arquivo
	for _, namespace := range namespaces {
		_, err := file.WriteString(namespace + "\n")
		if err != nil {
			return fmt.Errorf("falha ao escrever no arquivo: %v", err)
		}
	}

	fmt.Println("Namespaces salvos em 'todos.txt'")
	return nil
}

// ListBackupNamespaces lista todos os namespaces que contêm a palavra "backup" em seu nome
// e salva em um arquivo "backup.txt".
func ListBackupNamespaces(ctx context.Context, client *datastore.Client) error {
	fmt.Println("Listando namespaces com 'backup'...")
	namespaces, err := fetchNamespaces(ctx, client)
	if err != nil {
		return err
	}

	var backupNamespaces []string
	fmt.Println("Salvando namespaces com 'backup' em 'backup.txt'...")
	for _, namespace := range namespaces {
		// Adiciona apenas namespaces que contenham "backup" em seu nome
		if strings.Contains(strings.ToLower(namespace), "backup") {
			backupNamespaces = append(backupNamespaces, namespace)
		}
	}

	// Abre o arquivo para escrita
	file, err := os.Create("backup.txt")
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo: %v", err)
	}
	defer file.Close()

	// Escreve cada namespace em uma linha do arquivo
	for _, namespace := range backupNamespaces {
		_, err := file.WriteString(namespace + "\n")
		if err != nil {
			return fmt.Errorf("falha ao escrever no arquivo: %v", err)
		}
	}

	fmt.Println("Namespaces com 'backup' salvos em 'backup.txt'")
	return nil
}

// CalculateTotalStorageFromFile calcula o total de armazenamento em GBs dos namespaces
// listados em um arquivo e imprime o total usando processamento paralelo.
func CalculateTotalStorageFromFile(ctx context.Context, client *datastore.Client, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("falha ao abrir o arquivo %s: %v", filename, err)
	}
	defer file.Close()

	var wg sync.WaitGroup
	namespaceCh := make(chan string, 100)
	resultCh := make(chan float64, 100)

	// Lê os namespaces do arquivo e os envia para o canal
	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			namespace := scanner.Text()
			if namespace != "" {
				namespaceCh <- namespace
			}
		}
		close(namespaceCh)
	}()

	// Processa até 100 namespaces simultaneamente
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for namespace := range namespaceCh {
				storageGB, err := calculateNamespaceStorage(ctx, client, namespace)
				if err != nil {
					fmt.Printf("Erro ao calcular o armazenamento do namespace %s: %v\n", namespace, err)
					continue
				}
				fmt.Printf("Namespace: %s - Armazenamento: %.2f GB\n", namespace, storageGB)
				resultCh <- storageGB
			}
		}()
	}

	// Fecha o canal de resultados após todos os goroutines terminarem
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Soma o armazenamento total
	var totalStorageGB float64
	for storage := range resultCh {
		totalStorageGB += storage
	}

	fmt.Printf("Armazenamento total de todos os namespaces: %.2f GB\n", totalStorageGB)
	return nil
}

// calculateNamespaceStorage calcula o armazenamento em GBs de um namespace usando __Stat_Ns_Total__,
// incluindo o armazenamento de índices.
func calculateNamespaceStorage(ctx context.Context, client *datastore.Client, namespace string) (float64, error) {
	query := datastore.NewQuery("__Stat_Ns_Total__").Namespace(namespace).Limit(1)
	var stats []datastore.PropertyList

	_, err := client.GetAll(ctx, query, &stats)
	if err != nil {
		return 0, fmt.Errorf("falha ao consultar __Stat_Ns_Total__ para o namespace %s: %v", namespace, err)
	}

	if len(stats) == 0 {
		return 0, fmt.Errorf("nenhuma estatística encontrada para o namespace %s", namespace)
	}

	// Extrai os campos necessários diretamente
	var totalBytes int64
	for _, prop := range stats[0] {
		switch prop.Name {
		case "entity_bytes", "builtin_index_bytes", "composite_index_bytes":
			if val, ok := prop.Value.(int64); ok {
				totalBytes += val
			}
		}
	}

	if totalBytes == 0 {
		return 0, fmt.Errorf("não foi possível encontrar os campos de armazenamento para o namespace %s", namespace)
	}

	// Converte bytes para gigabytes (1 GB = 1,073,741,824 bytes)
	storageGB := float64(totalBytes) / (1024 * 1024 * 1024)
	return storageGB, nil
}

func GetSubscriberBucketName(ctx context.Context, client *datastore.Client, namespace string) (string, error) {
	query := datastore.NewQuery("Global_SubscriberNamespace").Namespace("").
		FilterField("Namespace", "=", namespace).Limit(1)

	var results []datastore.PropertyList

	_, err := client.GetAll(ctx, query, &results)
	if err != nil {
		fmt.Printf("Erro ao buscar SubscriberBucketName: %v\n", err)
		return "", fmt.Errorf("falha ao buscar SubscriberBucketName: %v", err)
	}

	if len(results) == 0 {
		fmt.Printf("Nenhum registro encontrado para o namespace %s\n", namespace)
		return "", fmt.Errorf("nenhum registro encontrado para o namespace %s", namespace)
	}

	// Verifica se o campo SubscriberBucketName existe e é uma string
	for _, prop := range results[0] {
		if prop.Name == "SubscriberBucketName" {
			if bucketName, ok := prop.Value.(string); ok {
				return bucketName, nil
			}
		}
	}

	return "", fmt.Errorf("campo SubscriberBucketName não encontrado")
}

// GetFileCreationDateFromBucket busca a data de criação de um arquivo dentro do bucket no Google Cloud Storage.
func GetFileCreationDateFromBucket(ctx context.Context, bucketName, fileName string) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("falha ao criar o cliente do Google Cloud Storage: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	object := bucket.Object(fileName)
	attrs, err := object.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("falha ao obter atributos do arquivo %s no bucket %s: %v", fileName, bucketName, err)
	}

	// Formata a data de criação no formato YYYYMMDD
	creationDate := attrs.Created.Format("20060102")
	return creationDate, nil
}

func GetPersonName(ctx context.Context, client *datastore.Client, namespace string, id string) (string, error) {
	var key *datastore.Key

	// Tenta converter o ID para int64
	if intID, err := strconv.ParseInt(id, 10, 64); err == nil {
		// Se a conversão for bem-sucedida, use int64
		key = datastore.IDKey("Person", intID, nil)
	} else {
		// Se a conversão falhar, use a string como chave
		key = datastore.NameKey("Person", id, nil)
	}

	key.Namespace = namespace

	query := datastore.NewQuery("Person").Namespace(namespace).FilterField("__key__", "=", key).Limit(1)

	var results []datastore.PropertyList

	_, err := client.GetAll(ctx, query, &results)
	if err != nil {
		return "", fmt.Errorf("falha ao buscar o registro do kind 'Person' com id %s: %v", id, err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("nenhum registro encontrado para o id %s no kind 'Person'", id)
	}

	for _, prop := range results[0] {
		if prop.Name == "Name" {
			if name, ok := prop.Value.(string); ok {
				return name, nil
			}
		}
	}

	return "", fmt.Errorf("campo Name não encontrado")
}

// CheckImagesFromThumb verifica imagens que possuem ImageViewBox com altura e largura de 250 e salva as informações em um arquivo.
func CheckImagesFromThumb(ctx context.Context, client *datastore.Client, namespace string) error {
	bucketName, err := GetSubscriberBucketName(ctx, client, namespace)
	count := 0
	countDiff := 0
	if err != nil {
		return err
	}

	fmt.Printf("BucketName: %s\n", bucketName)

	query := datastore.NewQuery("Picture").Namespace(namespace).KeysOnly()
	var records []string
	var mu sync.Mutex // Protege o acesso à variável compartilhada `records`

	// Define a quantidade de goroutines e um canal para gerenciar os resultados
	const numWorkers = 10
	resultsCh := make(chan []string, numWorkers)

	var wg sync.WaitGroup

	// Função de processamento para cada worker
	processBatch := func(batch []*datastore.Key) {
		defer wg.Done()
		localRecords := []string{}

		for _, key := range batch {
			var entity datastore.PropertyList
			if err := client.Get(ctx, key, &entity); err != nil {
				fmt.Printf("Erro ao buscar a entidade: %v\n", err)
				continue
			}

			var fileName, kind, kindId string

			count++
			if count-countDiff >= 100 {
				fmt.Printf("Total de imagens %d listadas: %d\n", count, len(records))
				countDiff = count
			}
			for _, prop := range entity {
				if prop.Name == "ImageViewBox" {
					if nestedEntity, ok := prop.Value.(*datastore.Entity); ok {
						viewBoxMap := make(map[string]interface{})
						for _, nestedProp := range nestedEntity.Properties {
							viewBoxMap[nestedProp.Name] = nestedProp.Value
						}

						// Verifica se height e width são iguais a 250
						if viewBoxMap["height"] == int64(250) && viewBoxMap["width"] == int64(250) {
							for _, prop := range entity {
								if prop.Name == "Kind" {
									if val, ok := prop.Value.(string); ok {
										kind = val
									}
								} else if prop.Name == "KindId" {
									if val, ok := prop.Value.(int64); ok {
										kindId = strconv.FormatInt(val, 10)
									}
								} else if prop.Name == "FileName" {
									if val, ok := prop.Value.(string); ok {
										fileName = val
									}
								}
							}

							dateInserted, err := GetFileCreationDateFromBucket(ctx, bucketName, fileName)
							if err != nil {
								fmt.Printf("Erro ao obter a data de criação para o arquivo %s: %v\n", fileName, err)
								continue
							}

							keyString := key.Name
							if keyString == "" {
								keyString = strconv.FormatInt(key.ID, 10)
							}

							record := fmt.Sprintf("%s-%s-%s-%s", dateInserted, keyString, kind, kindId)

							// Se o kind for "Person", busca a propriedade Name e adiciona ao registro
							if kind == "Person" && kindId != "" {
								name, err := GetPersonName(ctx, client, namespace, kindId)
								if err != nil {
									fmt.Printf("Erro ao obter o Name do kind 'Person' com id %s: %v\n", kindId, err)
									continue
								}
								record = fmt.Sprintf("%s-%s", record, name)
							}

							localRecords = append(localRecords, record)
						}
					}
				}
			}
		}

		// Envia os registros processados para o canal
		resultsCh <- localRecords
	}

	// Executa a busca em lotes de chaves e distribui para as goroutines
	it := client.Run(ctx, query)
	for {
		var batch []*datastore.Key
		for i := 0; i < 100; i++ {
			key, err := it.Next(nil)
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("falha ao iterar sobre chaves: %v", err)
			}
			batch = append(batch, key)
		}

		if len(batch) == 0 {
			break
		}

		wg.Add(1)
		go processBatch(batch)
	}

	// Aguarda o término de todas as goroutines
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Coleta os resultados das goroutines
	for res := range resultsCh {
		mu.Lock()
		records = append(records, res...)
		mu.Unlock()
	}

	// Ordena os registros pela data no início da string
	sort.Slice(records, func(i, j int) bool {
		return records[i] < records[j]
	})

	// Cria a pasta `check_thumb_images` se não existir
	if _, err := os.Stat("check_thumb_images"); os.IsNotExist(err) {
		err = os.Mkdir("check_thumb_images", os.ModePerm)
		if err != nil {
			return fmt.Errorf("falha ao criar a pasta 'check_thumb_images': %v", err)
		}
	}

	// Define o caminho do arquivo com base no namespace
	filePath := fmt.Sprintf("check_thumb_images/%s.txt", namespace)

	// Escreve os registros ordenados em um arquivo
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("falha ao criar o arquivo %s: %v", filePath, err)
	}
	defer file.Close()

	for _, record := range records {
		if _, err := file.WriteString(record + "\n"); err != nil {
			return fmt.Errorf("falha ao escrever no arquivo %s: %v", filePath, err)
		}
	}

	fmt.Printf("Total de imagens %d listadas: %d\n", count, len(records))
	fmt.Printf("Arquivo salvo em: %s\n", filePath)
	return nil
}
