package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"rcoi/internal/utils"
)

// Ключи контекста
type ContextKey string

const (
	UserEmailKey ContextKey = "user_email"
	UserRoleKey  ContextKey = "user_role"
)

// AuthMiddleware проверяет JWT токен и добавляет email и роль в контекст
func AuthMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Отсутствует токен", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			claims, err := utils.ParseToken(tokenString, logger)
			if err != nil {
				http.Error(w, "Неверный или просроченный токен", http.StatusUnauthorized)
				return
			}

			email, ok := claims["email"].(string)
			if !ok {
				http.Error(w, "Некорректный email в токене", http.StatusUnauthorized)
				return
			}

			role, ok := claims["role"].(string)
			if !ok {
				http.Error(w, "Некорректная роль в токене", http.StatusUnauthorized)
				return
			}

			// Используем функции-хелперы для добавления данных в контекст
			ctx := SetEmailToContext(r.Context(), email)
			ctx = SetRoleToContext(ctx, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleMiddleware проверяет роль пользователя
func RoleMiddleware(allowedRole string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := GetRoleFromContext(r.Context())
			if !ok {
				http.Error(w, "Роль не найдена в токене", http.StatusUnauthorized)
				return
			}

			if role != allowedRole {
				http.Error(w, "Доступ запрещён", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
