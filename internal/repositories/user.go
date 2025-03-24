package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"rcoi/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, password string) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateRefreshToken(ctx context.Context, email, refreshToken string) error
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) CreateUser(ctx context.Context, email, password string) error {
	_, err := r.db.Exec(ctx, "INSERT INTO users (email, password) VALUES ($1, $2)", email, password)
	return err
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	var refreshToken sql.NullString

	err := r.db.QueryRow(ctx, "SELECT id, email, password, refresh_token, role FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Email, &user.Password, &refreshToken, &user.Role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	if refreshToken.Valid {
		user.RefreshToken = refreshToken.String
	} else {
		user.RefreshToken = ""
	}

	return &user, nil
}

func (r *userRepo) UpdateRefreshToken(ctx context.Context, email, refreshToken string) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET refresh_token = $1 WHERE email = $2", refreshToken, email)
	return err
}
