package rag

import (
	"context"
	"errors"
	"langchain-go-ollama/internal/documents"
	"time"

	"github.com/google/uuid"
)

type IngestionService struct {
	embedder Embedder
	store    Store
	docRepo  documents.DocumentRepository
}

func NewIngestionService(store Store, embedder Embedder, docRepo documents.DocumentRepository) (*IngestionService, error) {
	if store == nil {
		return nil, errors.New("store is nil")
	}
	if embedder == nil {
		return nil, errors.New("embedder is nil")
	}
	if docRepo == nil {
		return nil, errors.New("docRepo is nil")
	}
	return &IngestionService{
		embedder: embedder,
		store:    store,
		docRepo:  docRepo,
	}, nil
}

func (s *IngestionService) Ingest(
	ctx context.Context,
	req IngestionRequest,
) ([]VectorDocument, error) {
	// insert the document (source of truth)
	docs := make([]VectorDocument, 0, len(req.Content))
	document := documents.Document{
		ID:        uuid.New(),
		Name:      req.Name,
		Type:      "raw",
		Content:   req.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.docRepo.Create(ctx, &document); err != nil {
		return nil, err
	}

	// chunk content
	chunks := chunkText(req.Content, 500)

	// embedd the content
	vectors, err := s.embedder.EmbedBatch(ctx, chunks)
	if err != nil {
		return nil, err
	}

	// create collection
	collection := DocumentCollection{
		UUID:      uuid.New(),
		Name:      time.Now().String(),
		CMetadata: nil,
	}

	s.store.InsertCollection(ctx, collection)

	for i, t := range chunks {
		docs = append(docs, VectorDocument{
			CollectionID: collection.UUID,
			Content:      t,
			Vector:       vectors[i],
			Metadata:     map[string]interface{}{},
		})
	}

	return s.store.Insert(ctx, docs)
}

func chunkText(text string, size int) []string {
	var chunks []string

	runes := []rune(text) // safe for UTF-8

	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}

		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

type IngestionRequest struct {
	Content string `json:"content"`
	Name    string `json:"name"`
}
