package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"rcoi/internal/models"
)

type NewsRepository interface {
	Create(ctx context.Context, news *models.News) error
	GetByID(ctx context.Context, id int) (*models.News, error)
	GetAll(ctx context.Context) ([]*models.News, error)
	Update(ctx context.Context, news *models.News) error
	Delete(ctx context.Context, id int) error
}

type newsRepo struct {
	db *pgxpool.Pool
}

func NewNewsRepository(db *pgxpool.Pool) NewsRepository {
	return &newsRepo{db: db}
}

func (r *newsRepo) Create(ctx context.Context, news *models.News) error {
	query := `INSERT INTO news (title, content) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, news.Title, news.Content).Scan(&news.ID, &news.CreatedAt, &news.UpdatedAt)
}

func (r *newsRepo) GetByID(ctx context.Context, id int) (*models.News, error) {
	news := &models.News{}
	query := `SELECT id, title, content, created_at, updated_at FROM news WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&news.ID, &news.Title, &news.Content, &news.CreatedAt, &news.UpdatedAt)
	return news, err
}

func (r *newsRepo) GetAll(ctx context.Context) ([]*models.News, error) {
	query := `SELECT id, title, content, created_at, updated_at FROM news ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newsList []*models.News
	for rows.Next() {
		var n models.News
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		newsList = append(newsList, &n)
	}
	return newsList, nil
}

func (r *newsRepo) Update(ctx context.Context, news *models.News) error {
	query := `UPDATE news SET title = $1, content = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, news.Title, news.Content, news.ID)
	return err
}

func (r *newsRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM news WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
