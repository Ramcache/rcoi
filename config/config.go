package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Config struct {
	DB *pgxpool.Pool
}

var (
	configInstance *Config
	once           sync.Once
)

func LoadConfig() (*Config, error) {
	var err error

	once.Do(func() {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("⚠️ Внимание: .env файл не загружен")
		}

		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			log.Println("❌ Ошибка: переменная окружения DATABASE_URL не установлена")
		}

		dbPool, dbErr := pgxpool.New(context.Background(), dbURL)
		if dbErr != nil {
			err = fmt.Errorf("ошибка подключения к БД: %w", dbErr)
			return
		}

		configInstance = &Config{DB: dbPool}
	})

	return configInstance, err
}

func (c *Config) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
