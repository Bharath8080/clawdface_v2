package vectors

import (
	"backend/configs"
	"context"
	"log"

	"github.com/pinecone-io/go-pinecone/v4/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

func CreateIndex(input []*pinecone.IntegratedRecord, namespace string, host string) bool {
	ctx := context.Background()

	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: configs.GetEnv("API_PINECONE_TOKEN"),
	})
	if err != nil {
		log.Fatalf("Failed to create Client: %v", err)
		return false
	}

	// To get the unique host for an index,
	// see https://docs.pinecone.io/guides/manage-data/target-an-index
	idxConnection, err := pc.Index(pinecone.NewIndexConnParams{
		Host:      host,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
		return false
	}

	err = idxConnection.UpsertRecords(ctx, input)
	if err != nil {
		log.Fatalf("Failed to upsert vectors: %v", err)
		return false
	}
	return true
}

func DeletePineconeNamespace(host, namespace string) error {
	ctx := context.Background()

	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: configs.GetEnv("API_PINECONE_TOKEN"),
	})
	if err != nil {
		log.Fatalf("Failed to create Client: %v", err)
		return err
	}

	// To get the unique host for an index,
	// see https://docs.pinecone.io/guides/manage-data/target-an-index
	idxConnection, err := pc.Index(pinecone.NewIndexConnParams{Host: host})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
		return err
	}

	err = idxConnection.DeleteNamespace(ctx, namespace)
	if err != nil {
		log.Fatalf("Failed to delete pinecone namespace: %v", err)
		return err
	}
	return nil
}

func DeletePineconeDocument(indexName, namespace, metadataValue string) error {
	ctx := context.Background()

	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: configs.GetEnv("API_PINECONE_TOKEN"),
	})
	if err != nil {
		log.Fatalf("Failed to create Client: %v", err)
		return err
	}

	// To get the unique host for an index,
	// see https://docs.pinecone.io/guides/manage-data/target-an-index
	idxConnection, err := pc.Index(pinecone.NewIndexConnParams{
		Host:      indexName,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
		return err
	}

	filter := map[string]*structpb.Value{
		"url": structpb.NewStringValue(metadataValue),
	}

	// Delete vectors matching the filter
	err = idxConnection.DeleteVectorsByFilter(ctx, &pinecone.MetadataFilter{
		Fields: filter,
	})

	if err != nil {
		log.Fatalf("Failed to delete vectors by metadata filter: %v", err)
		return err
	}

	log.Println("✅ Successfully deleted vectors with url" + metadataValue)
	return nil
}
