package llm

import (
	"context"
	"log"

	"github.com/tmc/langchaingo/llms/ollama"
)

type OllamaLLM struct {
	client *ollama.LLM
}

type LLM interface {
	Ask(ctx context.Context, prompt string) (string, error)
}

func NewOllamaLLM() (*OllamaLLM, error) {
	llm, err := ollama.New(ollama.WithModel("qwen3.5:cloud"))
	if err != nil {
		log.Fatal(err)
	}
	return &OllamaLLM{
		client: llm,
	}, nil
}

func (o *OllamaLLM) Ask(ctx context.Context, prompt string) (string, error) {
	return o.client.Call(ctx, prompt)
}
