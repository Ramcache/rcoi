package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
	"rcoi/internal/services"
	"strconv"
)

type DocumentHandler struct {
	service services.DocumentService
	logger  *zap.Logger
}

func NewDocumentHandler(service services.DocumentService, logger *zap.Logger) *DocumentHandler {
	return &DocumentHandler{service: service, logger: logger}
}

// Загрузка документа
func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Файл не найден", http.StatusBadRequest)
		return
	}
	defer file.Close()

	doc, err := h.service.UploadDocument(r.Context(), title, file, fileHeader)
	if err != nil {
		h.logger.Error("Ошибка загрузки файла", zap.Error(err))
		http.Error(w, "Ошибка загрузки файла", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

// Список документов
func (h *DocumentHandler) GetAllDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := h.service.GetAllDocuments(r.Context())
	if err != nil {
		http.Error(w, "Ошибка получения документов", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(docs)
}

// Скачивание документа по ID
func (h *DocumentHandler) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Некорректный ID документа", http.StatusBadRequest)
		return
	}

	doc, err := h.service.GetDocumentByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Документ не найден", http.StatusNotFound)
		return
	}

	filePath := filepath.Join("uploads", doc.Filename)
	w.Header().Set("Content-Disposition", "attachment; filename="+doc.Filename)
	http.ServeFile(w, r, filePath)
}

// Удаление документа по ID
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Некорректный ID документа", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteDocument(r.Context(), id)
	if err != nil {
		h.logger.Error("Ошибка удаления документа", zap.Error(err))
		http.Error(w, "Ошибка удаления документа", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
