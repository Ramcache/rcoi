package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rcoi/internal/middleware"
	"rcoi/internal/services"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

type AuthHandler struct {
	service services.AuthService
	logger  *zap.Logger
}

func NewAuthHandler(service services.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{service: service, logger: logger}
}

func validatePassword(password string) string {
	var (
		lengthRegex    = regexp.MustCompile(`^.{8,}$`)
		upperCaseRegex = regexp.MustCompile(`[A-Z]`)
		lowerCaseRegex = regexp.MustCompile(`[a-z]`)
		englishRegex   = regexp.MustCompile(`^[A-Za-z\d]+$`)
	)

	var errors []string

	if !lengthRegex.MatchString(password) {
		errors = append(errors, "Пароль должен быть не менее 8 символов")
	}
	if !upperCaseRegex.MatchString(password) {
		errors = append(errors, "Пароль должен содержать хотя бы одну заглавную букву")
	}
	if !lowerCaseRegex.MatchString(password) {
		errors = append(errors, "Пароль должен содержать хотя бы одну строчную букву")
	}
	if !englishRegex.MatchString(password) {
		errors = append(errors, "Пароль должен содержать только латинские буквы и цифры")
	}

	if len(errors) > 0 {
		return "Ошибка пароля: " + formatErrors(errors)
	}

	return ""
}

func formatErrors(errors []string) string {
	result := ""
	for i, err := range errors {
		if i > 0 {
			result += "; "
		}
		result += err
	}
	return result
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Ошибка декодирования JSON", zap.Error(err))
		http.Error(w, `{"error": "Неверный формат запроса"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "Email и пароль обязательны"}`, http.StatusBadRequest)
		return
	}

	if errMsg := validatePassword(req.Password); errMsg != "" {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, errMsg), http.StatusBadRequest)
		return
	}

	err := h.service.RegisterUser(r.Context(), req.Email, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			h.logger.Warn("Попытка регистрации с уже существующим email", zap.String("email", req.Email))
			http.Error(w, `{"error": "Email уже используется"}`, http.StatusConflict)
			return
		}

		h.logger.Error("Ошибка регистрации пользователя", zap.Error(err))
		http.Error(w, `{"error": "Ошибка регистрации"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Пользователь успешно зарегистрирован",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Ошибка декодирования JSON", zap.Error(err))
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Ошибка входа", zap.Error(err))
		http.Error(w, "Ошибка входа", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		MaxAge:   604800,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})

}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh-токен отсутствует", http.StatusUnauthorized)
		return
	}

	newAccessToken, newRefreshToken, err := h.service.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, "Ошибка обновления токена", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		MaxAge:   604800,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})

}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok {
		h.logger.Warn("Ошибка выхода: email не найден в контексте")
		http.Error(w, "Ошибка выхода", http.StatusUnauthorized)
		return
	}

	if err := h.service.Logout(r.Context(), email); err != nil {
		h.logger.Error("Ошибка выхода", zap.Error(err))
		http.Error(w, "Ошибка выхода", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}
