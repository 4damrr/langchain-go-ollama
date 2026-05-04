package documents

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*Document, error)
	List(ctx context.Context, limit, offset int) ([]Document, error)
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresDocumentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDocumentRepository(pool *pgxpool.Pool) (*PostgresDocumentRepository, error) {
	if pool == nil {
		return nil, errors.New("postgres repository pool is nil")
	}
	return &PostgresDocumentRepository{pool: pool}, nil
}

func (r *PostgresDocumentRepository) Create(ctx context.Context, doc *Document) error {
	query := `
		INSERT INTO documents (id, name, type, content)
		VALUES ($1, $2, $3, $4)
	`

	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		doc.ID,
		doc.Name,
		doc.Type,
		doc.Content,
	)

	return err
}

func (r *PostgresDocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	query := `
		SELECT id, name, type, content, created_at, updated_at
		FROM documents
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	var doc Document
	err := row.Scan(
		&doc.ID,
		&doc.Name,
		&doc.Type,
		&doc.Content,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (r *PostgresDocumentRepository) List(ctx context.Context, limit, offset int) ([]Document, error) {
	query := `
		SELECT id, name, type, content, created_at, updated_at
		FROM documents
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document

	for rows.Next() {
		var doc Document
		err := rows.Scan(
			&doc.ID,
			&doc.Name,
			&doc.Type,
			&doc.Content,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

func (r *PostgresDocumentRepository) Update(ctx context.Context, doc *Document) error {
	query := `
		UPDATE documents
		SET name = $1,
		    type = $2,
		    content = $3,
		    updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.pool.Exec(ctx, query,
		doc.Name,
		doc.Type,
		doc.Content,
		doc.ID,
	)

	return err
}

func (r *PostgresDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM documents WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}
