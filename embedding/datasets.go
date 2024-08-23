package embedding

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/copilot-extensions/rag-extension/copilot"
)

func Create(ctx context.Context, integrationID, apiToken string, content string) ([]float32, error) {
	resp, err := copilot.Embeddings(ctx, integrationID, apiToken, &copilot.EmbeddingsRequest{
		Model: copilot.ModelEmbeddings,
		Input: []string{content},
	})

	if err != nil {
		return nil, fmt.Errorf("error fetching embeddings: %w", err)
	}

	for _, data := range resp.Data {
		return data.Embedding, nil
	}

	return nil, fmt.Errorf("no embeddings found in response")
}

type Dataset struct {
	Embedding []float32
	Filename  string
}

func GenerateDatasets(integrationID, apiToken string, filenames []string) ([]*Dataset, error) {
	datasets := make([]*Dataset, len(filenames))
	for i, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("error reading in file %s: %w", filename, err)
		}

		fileContent, err := io.ReadAll(file)

		embedding, err := Create(context.Background(), integrationID, apiToken, string(fileContent))
		if err != nil {
			return nil, fmt.Errorf("error creating embedding for file %s: %w", filename, err)
		}

		datasets[i] = &Dataset{
			Embedding: embedding,
			Filename:  filename,
		}
	}

	return datasets, nil
}

func FindBestDataset(datasets []*Dataset, target []float32) (*Dataset, error) {
	var bestDataset *Dataset
	var bestScore float32

	var targetMagnitude float32
	for i := 0; i < len(target); i++ {
		targetMagnitude += target[i] * target[i]
	}

	for _, dataset := range datasets {
		// Score similarity using Cosine Similarity
		if len(target) != len(dataset.Embedding) {
			return nil, fmt.Errorf("embeddings are different length, cannot compare")
		}

		var docMagnitude, dotProduct float32
		for i := 0; i < len(target); i++ {
			docMagnitude += dataset.Embedding[i] * dataset.Embedding[i]
			dotProduct += target[i] * dataset.Embedding[i]
		}

		dotProduct /= float32(math.Sqrt(float64(targetMagnitude)) * math.Sqrt(float64(docMagnitude)))
		if dotProduct > bestScore {
			bestDataset = dataset
			bestScore = dotProduct
		}
	}

	return bestDataset, nil
}
