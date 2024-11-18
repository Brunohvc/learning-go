package delete_data

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

const (
	limit      = 1000
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

var (
	totalDeletions = 0
)

// KindInfo contém informações sobre o kind e a propriedade a ser filtrada.
type KindInfo struct {
	Kind   string
	Prop   string
	Filter interface{}
}

// DeleteData realiza a deleção com um limite de `maxTables` tabelas simultâneas.
func DeleteData(ctx context.Context, client *datastore.Client, kinds []KindInfo, namespace string, maxTables int) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxTables)
	totalDeletionsNamespace := 0

	for _, kind := range kinds {
		// Busca a propriedade de ordenação para cada kind (tabela)
		prop, err := getPropOrdering(ctx, client, namespace, kind.Kind)
		if err != nil {
			fmt.Printf("Erro ao buscar propriedade para kind %s: %v\n", kind.Kind, err)
			continue
		}
		kind.Prop = prop

		sem <- struct{}{} // Reserva um slot
		wg.Add(1)

		go func(kind KindInfo) {
			defer wg.Done()
			defer func() { <-sem }() // Libera o slot ao finalizar

			count := processEntities(ctx, client, kind, namespace)
			totalDeletionsNamespace += count
			fmt.Printf("Deleção de %d registros concluída para o kind %s no namespace %s.\n", count, kind.Kind, namespace)
		}(kind)
	}

	wg.Wait()
	fmt.Printf("Processo de deleção completo para o %snamespace %s%s. Total de registros deletados: %s%d%s\n", colorGreen, namespace, colorReset, colorRed, totalDeletionsNamespace, colorReset)
}

func processEntities(ctx context.Context, client *datastore.Client, kind KindInfo, namespace string) int {
	var totalCount int
	var diffDeletions int
	var cursor *datastore.Cursor

	for {
		query := datastore.NewQuery(kind.Kind).Namespace(namespace).KeysOnly().Limit(limit)
		if kind.Prop != "" {
			query = query.Order(kind.Prop)
		}

		if cursor != nil {
			query = query.Start(*cursor)
		}

		it := client.Run(ctx, query)
		keys := []*datastore.Key{}
		count := 0

		for {
			key, err := it.Next(nil)
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("Erro ao iterar: %v", err)
				time.Sleep(100 * time.Millisecond)
				break
			}
			keys = append(keys, key)
			count++
		}

		if count == 0 {
			// Não há mais registros a serem processados
			break
		}

		if err := client.DeleteMulti(ctx, keys); err != nil {
			log.Printf("Falha ao deletar registros: %v", err)
		} else {
			totalCount += count
		}

		if totalCount-diffDeletions >= 10000 {
			diffDeletions = totalCount
			fmt.Printf("Deleção de %d registros concluída para o kind %s no namespace %s.\n", totalCount, kind.Kind, namespace)
		}

		nextCursor, err := it.Cursor()
		if err != nil || count < limit {
			// Se houve erro ao obter o cursor ou se a quantidade de registros processados for menor que o limite, assume-se que não há mais registros
			break
		}
		cursor = &nextCursor
	}

	totalDeletions += totalCount

	return totalCount
}

// getPropOrdering busca a propriedade de ordenação
func getPropOrdering(ctx context.Context, client *datastore.Client, namespace, kind string) (string, error) {
	query := datastore.NewQuery(kind).Namespace(namespace).Limit(1)
	var entity datastore.PropertyList // Usa PropertyList para carregar as propriedades da entidade
	it := client.Run(ctx, query)

	_, err := it.Next(&entity)
	if err == iterator.Done {
		return "", nil // Se não há entidade, retorna string vazia
	}
	if err != nil {
		return "", fmt.Errorf("falha ao obter entidade para o kind %s: %v", kind, err)
	}

	// Inspeciona as propriedades da entidade para decidir a propriedade de ordenação
	// Neste exemplo, assume-se que "created_at" é a propriedade preferida
	for _, prop := range entity {
		if prop.Name == "created_at" {
			return "created_at", nil
		}
	}

	// Se "created_at" não existir, retorna o primeiro campo encontrado como fallback
	if len(entity) > 0 {
		return entity[0].Name, nil
	}

	return "", nil
}
