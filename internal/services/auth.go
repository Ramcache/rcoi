package services

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"rcoi/internal/repositories"
)

type AuthService interface {
	RegisterUser(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, string, error)
	RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error)
	Logout(ctx context.Context, email string) error
}

type authService struct {
	repo   repositories.UserRepository
	logger *zap.Logger
}

type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(repo repositories.UserRepository, logger *zap.Logger) AuthService {
	return &authService{repo: repo, logger: logger}
}

var cachedSecretKey []byte
var once sync.Once

func GetSecretKey(logger *zap.Logger) ([]byte, error) {
	var err error
	once.Do(func() {
		key := os.Getenv("JWT_SECRET")
		if key == "" {
			logger.Fatal("Ошибка: переменная JWT_SECRET не установлена")
			err = errors.New("переменная окружения JWT_SECRET отсутствует")
			return
		}
		cachedSecretKey = []byte(key)
	})
	return cachedSecretKey, err
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func checkPassword(hashedPwd, plainPwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd)) == nil
}

func generateToken(email, role string, expiry time.Duration, logger *zap.Logger) (string, error) {
	key, err := GetSecretKey(logger)
	if err != nil {
		return "", err
	}

	claims := Claims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(key)
}

func parseToken(tokenString string, logger *zap.Logger) (jwt.MapClaims, error) {
	key, err := GetSecretKey(logger)
	if err != nil {
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

func (s *authService) RegisterUser(ctx context.Context, email, password string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		s.logger.Error("Ошибка хеширования пароля", zap.Error(err))
		return err
	}
	return s.repo.CreateUser(ctx, email, hashedPassword)
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("Пользователь не найден", zap.String("email", email), zap.Error(err))
		return "", "", errors.New("пользователь не найден")
	}

	if !checkPassword(user.Password, password) {
		s.logger.Warn("Неверный пароль", zap.String("email", email))
		return "", "", errors.New("неверный пароль")
	}

	accessToken, err := generateToken(email, user.Role, time.Hour, s.logger)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateToken(email, user.Role, time.Hour*24*7, s.logger)
	if err != nil {
		return "", "", err
	}

	err = s.repo.UpdateRefreshToken(ctx, email, refreshToken)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error) {
	claims, err := parseToken(oldRefreshToken, s.logger)
	if err != nil {
		return "", "", errors.New("неверный или просроченный refresh-токен")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", "", errors.New("неверный токен")
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("пользователь не найден")
	}

	if user.RefreshToken != oldRefreshToken {
		s.logger.Warn("Несоответствие refresh-токена", zap.String("email", email))
		return "", "", errors.New("refresh-токен не совпадает")
	}

	newAccessToken, err := generateToken(email, user.Role, time.Hour, s.logger)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := generateToken(email, user.Role, time.Hour*24*7, s.logger)
	if err != nil {
		return "", "", err
	}

	err = s.repo.UpdateRefreshToken(ctx, email, newRefreshToken)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, email string) error {
	return s.repo.UpdateRefreshToken(ctx, email, "")
}
