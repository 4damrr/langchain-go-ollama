package rag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

type Generator struct {
	llm llms.Model
}

func NewGenerator(llm llms.Model) (*Generator, error) {
	if llm == nil {
		return nil, errors.New("llms model is nil")
	}
	return &Generator{
		llm: llm,
	}, nil
}

func (g *Generator) Generate(
	ctx context.Context,
	query string,
	docs []VectorDocument,
) (string, error) {

	contextText := buildContext(docs)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: `You are a helpful assistant.

Use ONLY the provided context to answer the question.
If the answer is not in the context, say "I don't know".`,
				},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{
					Text: fmt.Sprintf(`
Context:
%s

Question:
%s
`, contextText, query),
				},
			},
		},
	}

	resp, err := g.llm.GenerateContent(ctx, messages)
	if err != nil {
		return "", err
	}

	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return resp.Choices[0].Content, nil
}

func buildContext(docs []VectorDocument) string {
	var sb strings.Builder

	for _, d := range docs {
		sb.WriteString(d.Content)
		sb.WriteString("\n---\n")
	}

	return sb.String()
}
