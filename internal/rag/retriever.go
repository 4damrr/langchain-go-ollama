package rag

import (
	"context"
	"errors"
)

type Retriever struct {
	store    Store
	embedder Embedder
}

func NewRetriever(store Store, embedder Embedder) (*Retriever, error) {
	if store == nil {
		return nil, errors.New("store is nil")
	}
	if embedder == nil {
		return nil, errors.New("embedder is nil")
	}
	return &Retriever{
		store:    store,
		embedder: embedder,
	}, nil
}

func (r *Retriever) Retrieve(
	ctx context.Context,
	request SearchRequest,
) ([]VectorDocument, error) {

	// 1. embed query
	vec, err := r.embedder.Embed(ctx, request.Query)
	if err != nil {
		return nil, err
	}

	// 2. search
	return r.store.SimilaritySearch(ctx, vec, request.K)
}

type SearchRequest struct {
	Query string `json:"query"`
	K     int    `json:"k"`
	//DocumentId string `json:"document_id"`
}
