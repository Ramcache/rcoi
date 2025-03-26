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

// UploadDocument godoc
// @Summary Загрузка документа
// @Description Загрузка документа на сервер
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Название документа"
// @Param file formData file true "Файл документа"
// @Success 201 {object} object
// @Failure 400 "Файл не найден"
// @Failure 500 "Ошибка загрузки файла"
// @Router /api/documents [post]
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

// GetAllDocuments godoc
// @Summary Получение списка всех документов
// @Description Возвращает список всех загруженных документов
// @Tags documents
// @Produce json
// @Success 200 {array} object
// @Failure 500 "Ошибка получения документов"
// @Router /api/documents [get]
func (h *DocumentHandler) GetAllDocuments(w http.ResponseWriter, r *http.Request) {
	docs, err := h.service.GetAllDocuments(r.Context())
	if err != nil {
		http.Error(w, "Ошибка получения документов", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(docs)
}

// DownloadDocument godoc
// @Summary Скачивание документа по ID
// @Description Скачивание документа по его ID
// @Tags documents
// @Param id path int true "ID документа"
// @Success 200 "Файл для скачивания"
// @Failure 400 "Некорректный ID документа"
// @Failure 404 "Документ не найден"
// @Router /api/documents/{id} [get]
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

// DeleteDocument godoc
// @Summary Удаление документа по ID
// @Description Удаляет документ по указанному ID
// @Tags documents
// @Param id path int true "ID документа"
// @Success 204 "Документ удален"
// @Failure 400 "Некорректный ID документа"
// @Failure 500 "Ошибка удаления документа"
// @Router /api/documents/{id} [delete]
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
