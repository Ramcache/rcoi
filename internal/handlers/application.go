package handlers

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"rcoi/internal/models"
	"rcoi/internal/services"
)

type ApplicationHandler struct {
	service services.ApplicationService
	logger  *zap.Logger
}

func NewApplicationHandler(service services.ApplicationService, logger *zap.Logger) *ApplicationHandler {
	return &ApplicationHandler{service: service, logger: logger}
}

func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	description := r.FormValue("description")
	url := r.FormValue("url")

	var file multipart.File
	var fileHeader *multipart.FileHeader
	var err error

	if url == "" {
		file, fileHeader, err = r.FormFile("file")
		if err != nil {
			http.Error(w, "Файл не найден", http.StatusBadRequest)
			return
		}
		defer file.Close()
	}

	app := &models.Application{
		Title:       title,
		Description: description,
		URL:         url,
	}

	err = h.service.CreateApplication(r.Context(), app, file, fileHeader)
	if err != nil {
		h.logger.Error("Ошибка создания приложения", zap.Error(err))
		http.Error(w, "Ошибка создания приложения", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

func (h *ApplicationHandler) GetAllApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := h.service.GetAllApplications(r.Context())
	if err != nil {
		http.Error(w, "Ошибка получения приложений", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(apps)
}

func (h *ApplicationHandler) GetApplicationByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Некорректный ID приложения", http.StatusBadRequest)
		return
	}

	app, err := h.service.GetApplicationByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Приложение не найдено", http.StatusNotFound)
		return
	}

	if app.URL != "" {
		json.NewEncoder(w).Encode(app)
		return
	}

	filePath := filepath.Join("uploads", app.Filename)
	w.Header().Set("Content-Disposition", "attachment; filename="+app.Filename)
	http.ServeFile(w, r, filePath)
}

func (h *ApplicationHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var app models.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}
	app.ID = id

	if err := h.service.UpdateApplication(r.Context(), &app); err != nil {
		http.Error(w, "Ошибка обновления приложения", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(app)
}

func (h *ApplicationHandler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Некорректный ID приложения", http.StatusBadRequest)
		return
	}

	app, err := h.service.GetApplicationByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Приложение не найдено", http.StatusNotFound)
		return
	}

	if app.Filename != "" {
		os.Remove(filepath.Join("uploads", app.Filename))
	}

	err = h.service.DeleteApplication(r.Context(), id)
	if err != nil {
		h.logger.Error("Ошибка удаления приложения", zap.Error(err))
		http.Error(w, "Ошибка удаления приложения", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
