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

// CreateNews godoc
// @Summary Создание новости
// @Description Создаёт новую новость
// @Tags news
// @Accept json
// @Produce json
// @Param news body models.News true "Данные новости"
// @Success 201 {object} models.News
// @Failure 400 "Неверный формат запроса"
// @Failure 500 "Ошибка создания новости"
// @Router /api/news [post]
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

// GetNewsByID godoc
// @Summary Получение новости по ID
// @Description Возвращает новость по указанному ID
// @Tags news
// @Produce json
// @Param id path int true "ID новости"
// @Success 200 {object} models.News
// @Failure 400 "Некорректный ID"
// @Failure 404 "Новость не найдена"
// @Router /api/news/{id} [get]
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

// GetAllNews godoc
// @Summary Получение списка всех новостей
// @Description Возвращает список всех новостей
// @Tags news
// @Produce json
// @Success 200 {array} models.News
// @Failure 500 "Ошибка получения новостей"
// @Router /api/news [get]
func (h *NewsHandler) GetAllNews(w http.ResponseWriter, r *http.Request) {
	newsList, err := h.service.GetAllNews(r.Context())
	if err != nil {
		http.Error(w, "Ошибка получения новостей", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newsList)
}

// UpdateNews godoc
// @Summary Обновление новости по ID
// @Description Обновляет новость по указанному ID
// @Tags news
// @Accept json
// @Produce json
// @Param id path int true "ID новости"
// @Param news body models.News true "Обновляемые данные новости"
// @Success 200 {object} models.News
// @Failure 400 "Некорректный ID или неверный формат запроса"
// @Failure 500 "Ошибка обновления новости"
// @Router /api/news/{id} [put]
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

// DeleteNews godoc
// @Summary Удаление новости по ID
// @Description Удаляет новость по указанному ID
// @Tags news
// @Param id path int true "ID новости"
// @Success 204 "Новость удалена"
// @Failure 400 "Некорректный ID"
// @Failure 500 "Ошибка удаления новости"
// @Router /api/news/{id} [delete]
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
