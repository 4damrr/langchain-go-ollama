package rag

import (
	"context"
	"errors"
)

type Service struct {
	retriever *Retriever
	generator *Generator
}

func NewService(
	retriever *Retriever,
	generator *Generator,
) (*Service, error) {
	if retriever == nil {
		return nil, errors.New("retriever is required")
	}
	if generator == nil {
		return nil, errors.New("generator is required")
	}
	return &Service{
		retriever: retriever,
		generator: generator,
	}, nil
}

func (s *Service) Ask(ctx context.Context, query string, k int) (string, error) {
	// Retrieve relevant chunk
	docs, err := s.retriever.Retrieve(ctx, SearchRequest{
		Query: query,
		K:     k,
	})
	if err != nil {
		return "", err
	}

	// Generate the answer
	return s.generator.Generate(ctx, query, docs)
}
