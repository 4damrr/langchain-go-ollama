package rag

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type Store interface {
	//SimilaritySearch(vector []float32, k int) ([]Document, error)
	Insert(ctx context.Context, docs []VectorDocument) ([]VectorDocument, error)
	InsertCollection(ctx context.Context, col DocumentCollection) (uuid.UUID, error)

	SimilaritySearch(ctx context.Context, queryVector []float32, k int) ([]VectorDocument, error)
}

type PGVectorStore struct {
	pool *pgxpool.Pool
}

func NewPGVectorStore(pool *pgxpool.Pool) (*PGVectorStore, error) {
	if pool == nil {
		return nil, errors.New("pgxpool is nil")
	}
	return &PGVectorStore{
		pool: pool,
	}, nil
}

func (s *PGVectorStore) InsertCollection(ctx context.Context, col DocumentCollection) (uuid.UUID, error) {

	// generate UUID if empty
	if col.UUID == uuid.Nil {
		col.UUID = uuid.New()
	}

	// marshal metadata
	metaBytes, err := json.Marshal(col.CMetadata)
	if err != nil {
		return uuid.Nil, err
	}

	query := `
		INSERT INTO langchain_pg_collection (uuid, name, cmetadata)
		VALUES ($1, $2, $3)
	`

	_, err = s.pool.Exec(
		ctx,
		query,
		col.UUID,
		col.Name,
		metaBytes,
	)
	if err != nil {
		return uuid.Nil, err
	}

	return col.UUID, nil
}

func (s *PGVectorStore) Insert(ctx context.Context, docs []VectorDocument) ([]VectorDocument, error) {
	query := `
		INSERT INTO langchain_pg_embedding
		(collection_id, embedding, document, cmetadata, custom_id, uuid)
		VALUES ($1, $2, $3, $4, $5, gen_random_uuid())
	`

	batch := &pgx.Batch{}

	for _, d := range docs {
		metaJSON, err := json.Marshal(d.Metadata)
		if err != nil {
			return nil, err
		}

		batch.Queue(
			query,
			d.CollectionID,
			pgvector.NewVector(d.Vector),
			d.Content,
			metaJSON,
			d.CustomID,
		)
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()

	return docs, br.Close()
}

func (s *PGVectorStore) Close() {
	s.pool.Close()
}

type VectorDocument struct {
	CollectionID uuid.UUID              // uuid (required)
	Content      string                 // document
	Vector       []float32              // embedding
	Metadata     map[string]interface{} // cmetadata (jsonb)
	CustomID     *string                // optional external id
}

type DocumentCollection struct {
	UUID      uuid.UUID
	Name      string
	CMetadata map[string]interface{}
}

type Query struct {
	Text string
}

type InsertRequest struct {
	CollectionID string                `json:"collection_id"`
	Documents    []InsertDocumentInput `json:"documents"`
}

type InsertDocumentInput struct {
	Content  string                 `json:"content"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
	CustomID *string                `json:"custom_id,omitempty"`
}

type InsertResponse struct {
	Inserted int `json:"inserted"`
}

// using cosine distance <=>
func (s *PGVectorStore) SimilaritySearch(
	ctx context.Context,
	queryVector []float32,
	k int,
) ([]VectorDocument, error) {

	query := `
		SELECT collection_id, document, embedding, cmetadata
		FROM langchain_pg_embedding
		ORDER BY embedding <=> $1
		LIMIT $2
	`

	rows, err := s.pool.Query(ctx, query, pgvector.NewVector(queryVector), k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []VectorDocument

	for rows.Next() {
		var doc VectorDocument
		var metaBytes []byte
		var vec pgvector.Vector

		err := rows.Scan(
			&doc.CollectionID,
			&doc.Content,
			&vec,
			&metaBytes,
		)
		if err != nil {
			return nil, err
		}

		doc.Vector = vec.Slice()

		if metaBytes != nil {
			_ = json.Unmarshal(metaBytes, &doc.Metadata)
		}

		results = append(results, doc)
	}

	return results, nil
}
