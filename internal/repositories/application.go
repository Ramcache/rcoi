package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"rcoi/internal/models"
)

type ApplicationRepository interface {
	Create(ctx context.Context, app *models.Application) error
	GetByID(ctx context.Context, id int) (*models.Application, error)
	GetAll(ctx context.Context) ([]*models.Application, error)
	Update(ctx context.Context, app *models.Application) error
	Delete(ctx context.Context, id int) error
}

type applicationRepo struct {
	db *pgxpool.Pool
}

func NewApplicationRepository(db *pgxpool.Pool) ApplicationRepository {
	return &applicationRepo{db: db}
}

func (r *applicationRepo) Create(ctx context.Context, app *models.Application) error {
	query := `
		INSERT INTO applications (title, description, filename, url) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, app.Title, app.Description, app.Filename, app.URL).
		Scan(&app.ID, &app.CreatedAt)
}

func (r *applicationRepo) GetByID(ctx context.Context, id int) (*models.Application, error) {
	app := &models.Application{}
	query := `
		SELECT id, title, description, filename, url, created_at 
		FROM applications 
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).
		Scan(&app.ID, &app.Title, &app.Description, &app.Filename, &app.URL, &app.CreatedAt)
	return app, err
}

func (r *applicationRepo) GetAll(ctx context.Context) ([]*models.Application, error) {
	query := `
		SELECT id, title, description, filename, url, created_at 
		FROM applications 
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*models.Application
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.Filename, &a.URL, &a.CreatedAt); err != nil {
			return nil, err
		}
		apps = append(apps, &a)
	}
	return apps, nil
}

func (r *applicationRepo) Update(ctx context.Context, app *models.Application) error {
	query := `
		UPDATE applications 
		SET title = $1, description = $2, filename = $3, url = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, app.Title, app.Description, app.Filename, app.URL, app.ID)
	return err
}

func (r *applicationRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM applications WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
