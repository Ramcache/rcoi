package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"rcoi/internal/models"
	"rcoi/internal/services"
)

type NewsHandler struct {
	service services.NewsService
	logger  *zap.Logger
}

func NewNewsHandler(service services.NewsService, logger *zap.Logger) *NewsHandler {
	return &NewsHandler{service: service, logger: logger}
}

func (h *NewsHandler) CreateNews(w http.ResponseWriter, r *http.Request) {
	var news models.News
	if err := json.NewDecoder(r.Body).Decode(&news); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateNews(r.Context(), &news); err != nil {
		http.Error(w, "Ошибка создания новости", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(news)
}

func (h *NewsHandler) GetNewsByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	news, err := h.service.GetNewsByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Новость не найдена", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(news)
}

func (h *NewsHandler) GetAllNews(w http.ResponseWriter, r *http.Request) {
	newsList, err := h.service.GetAllNews(r.Context())
	if err != nil {
		http.Error(w, "Ошибка получения новостей", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newsList)
}

func (h *NewsHandler) UpdateNews(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var news models.News
	if err := json.NewDecoder(r.Body).Decode(&news); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}
	news.ID = id

	if err := h.service.UpdateNews(r.Context(), &news); err != nil {
		http.Error(w, "Ошибка обновления новости", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(news)
}

func (h *NewsHandler) DeleteNews(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteNews(r.Context(), id); err != nil {
		http.Error(w, "Ошибка удаления новости", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
