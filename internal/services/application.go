package services

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"rcoi/internal/models"
	"rcoi/internal/repositories"
)

type ApplicationService interface {
	CreateApplication(ctx context.Context, app *models.Application, file multipart.File, fileHeader *multipart.FileHeader) error
	GetApplicationByID(ctx context.Context, id int) (*models.Application, error)
	GetAllApplications(ctx context.Context) ([]*models.Application, error)
	UpdateApplication(ctx context.Context, app *models.Application) error
	DeleteApplication(ctx context.Context, id int) error
}

type applicationService struct {
	repo repositories.ApplicationRepository
}

func NewApplicationService(repo repositories.ApplicationRepository) ApplicationService {
	return &applicationService{repo: repo}
}

// CreateApplication позволяет загружать файл или сохранять URL
func (s *applicationService) CreateApplication(ctx context.Context, app *models.Application, file multipart.File, fileHeader *multipart.FileHeader) error {
	if file != nil && fileHeader != nil {
		filename := time.Now().Format("20060102150405") + "_" + filepath.Base(fileHeader.Filename)
		path := filepath.Join("uploads", filename)

		dst, err := os.Create(path)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err = io.Copy(dst, file); err != nil {
			return err
		}
		app.Filename = filename
	}

	return s.repo.Create(ctx, app)
}

func (s *applicationService) GetApplicationByID(ctx context.Context, id int) (*models.Application, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *applicationService) GetAllApplications(ctx context.Context) ([]*models.Application, error) {
	return s.repo.GetAll(ctx)
}

func (s *applicationService) UpdateApplication(ctx context.Context, app *models.Application) error {
	return s.repo.Update(ctx, app)
}

func (s *applicationService) DeleteApplication(ctx context.Context, id int) error {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if app.Filename != "" {
		_ = os.Remove(filepath.Join("uploads", app.Filename))
	}

	return s.repo.Delete(ctx, id)
}
