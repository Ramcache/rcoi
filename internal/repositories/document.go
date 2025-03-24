package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"rcoi/internal/models"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, id int) (*models.Document, error)
	GetAll(ctx context.Context) ([]*models.Document, error)
	Delete(ctx context.Context, id int) error
}

type documentRepo struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) DocumentRepository {
	return &documentRepo{db: db}
}

func (r *documentRepo) Create(ctx context.Context, doc *models.Document) error {
	query := `INSERT INTO documents (title, filename) VALUES ($1, $2) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, doc.Title, doc.Filename).Scan(&doc.ID, &doc.CreatedAt)
}

func (r *documentRepo) GetByID(ctx context.Context, id int) (*models.Document, error) {
	doc := &models.Document{}
	query := `SELECT id, title, filename, created_at FROM documents WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&doc.ID, &doc.Title, &doc.Filename, &doc.CreatedAt)
	return doc, err
}

func (r *documentRepo) GetAll(ctx context.Context) ([]*models.Document, error) {
	query := `SELECT id, title, filename, created_at FROM documents ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.ID, &d.Title, &d.Filename, &d.CreatedAt); err != nil {
			return nil, err
		}
		docs = append(docs, &d)
	}
	return docs, nil
}

func (r *documentRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM documents WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
