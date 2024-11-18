package clone_data

import (
	"context"
	"fmt"
	"log"
	"namespace_destructor/get_data"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

// CloneData copia todas as tabelas e registros do namespace do projeto de origem para o projeto de destino.
func CloneData(ctx context.Context, sourceProjectID, destProjectID, namespace string) error {
	sourceClient, err := datastore.NewClient(ctx, sourceProjectID)
	if err != nil {
		return fmt.Errorf("falha ao criar cliente do projeto de origem: %v", err)
	}
	defer sourceClient.Close()

	destClient, err := datastore.NewClient(ctx, destProjectID)
	if err != nil {
		return fmt.Errorf("falha ao criar cliente do projeto de destino: %v", err)
	}
	defer destClient.Close()

	// Lista todas as tabelas no namespace
	kinds, err := get_data.ListKinds(ctx, sourceClient, namespace)
	if err != nil {
		return fmt.Errorf("falha ao listar kinds: %v", err)
	}

	for _, kind := range kinds {
		fmt.Printf("Clonando registros da tabela %s...\n", kind)
		if err := cloneKindData(ctx, sourceClient, destClient, kind, namespace); err != nil {
			log.Printf("Falha ao clonar registros para a tabela %s: %v", kind, err)
		} else {
			fmt.Printf("Clonagem de %s concluída.\n", kind)
		}
	}

	fmt.Println("Processo de clonagem completo.")
	return nil
}

func cloneKindData(ctx context.Context, sourceClient, destClient *datastore.Client, kind, namespace string) error {
	query := datastore.NewQuery(kind).Namespace(namespace)
	it := sourceClient.Run(ctx, query)

	var entities []datastore.PropertyList
	var keys []*datastore.Key
	batchSize := 500

	for {
		var entity datastore.PropertyList
		key, err := it.Next(&entity)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("falha ao iterar registros: %v", err)
		}

		// Cria a chave com o namespace especificado
		if key.ID != 0 {
			keys = append(keys, &datastore.Key{
				Kind:      kind,
				ID:        key.ID,
				Namespace: namespace,
			})
		} else {
			keys = append(keys, &datastore.Key{
				Kind:      kind,
				Name:      key.Name,
				Namespace: namespace,
			})
		}

		entities = append(entities, entity)

		// Processa o batch se o tamanho for alcançado
		if len(entities) == batchSize {
			if err := putEntities(ctx, destClient, keys, entities, namespace); err != nil {
				return err
			}
			keys = nil
			entities = nil
		}
	}

	// Insere qualquer entidade restante
	if len(entities) > 0 {
		if err := putEntities(ctx, destClient, keys, entities, namespace); err != nil {
			return err
		}
	}

	return nil
}

func putEntities(ctx context.Context, client *datastore.Client, keys []*datastore.Key, entities []datastore.PropertyList, namespace string) error {
	// Define o namespace nas entidades ao inseri-las no projeto de destino
	for i := range keys {
		keys[i].Namespace = namespace
	}

	_, err := client.PutMulti(ctx, keys, entities)
	if err != nil {
		return fmt.Errorf("falha ao inserir registros no projeto de destino: %v", err)
	}
	return nil
}
