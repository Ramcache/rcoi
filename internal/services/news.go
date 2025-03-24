package services

import (
	"context"
	"go.uber.org/zap"
	"rcoi/internal/models"
	"rcoi/internal/repositories"
)

type NewsService interface {
	CreateNews(ctx context.Context, news *models.News) error
	GetNewsByID(ctx context.Context, id int) (*models.News, error)
	GetAllNews(ctx context.Context) ([]*models.News, error)
	UpdateNews(ctx context.Context, news *models.News) error
	DeleteNews(ctx context.Context, id int) error
}

type newsService struct {
	repo   repositories.NewsRepository
	logger *zap.Logger
}

func NewNewsService(repo repositories.NewsRepository, logger *zap.Logger) NewsService {
	return &newsService{repo: repo, logger: logger}
}

func (s *newsService) CreateNews(ctx context.Context, news *models.News) error {
	return s.repo.Create(ctx, news)
}

func (s *newsService) GetNewsByID(ctx context.Context, id int) (*models.News, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *newsService) GetAllNews(ctx context.Context) ([]*models.News, error) {
	return s.repo.GetAll(ctx)
}

func (s *newsService) UpdateNews(ctx context.Context, news *models.News) error {
	return s.repo.Update(ctx, news)
}

func (s *newsService) DeleteNews(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
