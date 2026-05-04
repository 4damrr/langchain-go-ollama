package rag

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/ollama"
)

type OllamaEmbedder struct {
	client *ollama.LLM
}

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, text []string) ([][]float32, error)
}

func NewOllamaEmbedder(model string) (*OllamaEmbedder, error) {
	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		log.Fatal(err)
	}
	return &OllamaEmbedder{
		client: llm,
	}, nil
}
	
func (o *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	vectors, err := o.client.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(vectors) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}

	return vectors[0], nil
}

func (o *OllamaEmbedder) EmbedBatch(ctx context.Context, text []string) ([][]float32, error) {
	return o.client.CreateEmbedding(ctx, text)
}

type UserInput struct {
	Query []string `json:"query"`
}
