package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"rcoi/internal/services"
)

func ParseToken(tokenString string, logger *zap.Logger) (jwt.MapClaims, error) {
	key, err := services.GetSecretKey(logger)
	if err != nil {
		logger.Error("Ошибка получения секретного ключа", zap.Error(err))
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи")
		}
		return key, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("неверный или просроченный токен")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("неверные claims")
	}

	return claims, nil
}
