package services

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"rcoi/internal/models"
	"rcoi/internal/repositories"
	"time"
)

type DocumentService interface {
	UploadDocument(ctx context.Context, title string, file multipart.File, fileHeader *multipart.FileHeader) (*models.Document, error)
	GetDocumentByID(ctx context.Context, id int) (*models.Document, error)
	GetAllDocuments(ctx context.Context) ([]*models.Document, error)
	DeleteDocument(ctx context.Context, id int) error
}

type documentService struct {
	repo repositories.DocumentRepository
}

func NewDocumentService(repo repositories.DocumentRepository) DocumentService {
	return &documentService{repo: repo}
}

func (s *documentService) UploadDocument(ctx context.Context, title string, file multipart.File, fileHeader *multipart.FileHeader) (*models.Document, error) {
	filename := time.Now().Format("20060102150405") + "_" + filepath.Base(fileHeader.Filename)
	path := filepath.Join("uploads", filename)

	dst, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		return nil, err
	}

	doc := &models.Document{
		Title:    title,
		Filename: filename,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *documentService) GetDocumentByID(ctx context.Context, id int) (*models.Document, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *documentService) GetAllDocuments(ctx context.Context) ([]*models.Document, error) {
	return s.repo.GetAll(ctx)
}

func (s *documentService) DeleteDocument(ctx context.Context, id int) error {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	os.Remove(filepath.Join("uploads", doc.Filename))
	return s.repo.Delete(ctx, id)
}
