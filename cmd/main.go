package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rcoi/config"
	_ "rcoi/docs"
	"rcoi/internal/handlers"
	"rcoi/internal/middleware"
	"rcoi/internal/repositories"
	"rcoi/internal/services"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}
	defer cfg.Close()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer logger.Sync()

	logger.Info("Сервис запущен")

	userRepo := repositories.NewUserRepository(cfg.DB)
	authService := services.NewAuthService(userRepo, logger)
	authHandler := handlers.NewAuthHandler(authService, logger)

	newsRepo := repositories.NewNewsRepository(cfg.DB)
	newsService := services.NewNewsService(newsRepo, logger)
	newsHandler := handlers.NewNewsHandler(newsService, logger)

	docRepo := repositories.NewDocumentRepository(cfg.DB)
	docService := services.NewDocumentService(docRepo)
	docHandler := handlers.NewDocumentHandler(docService, logger)

	appRepo := repositories.NewApplicationRepository(cfg.DB)
	appService := services.NewApplicationService(appRepo)
	appHandler := handlers.NewApplicationHandler(appService, logger)

	r := mux.NewRouter()

	// Открытые маршруты (без middleware)
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/refresh", authHandler.Refresh).Methods("POST")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Защищённые маршруты (JWT middleware)
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(logger))

	protected.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		email := r.Context().Value(middleware.UserEmailKey).(string)
		w.Write([]byte("Добро пожаловать, " + email + "!"))
	}).Methods("GET")

	// Новости
	protected.HandleFunc("/news", newsHandler.CreateNews).Methods("POST")
	protected.HandleFunc("/news", newsHandler.GetAllNews).Methods("GET")
	protected.HandleFunc("/news/{id}", newsHandler.GetNewsByID).Methods("GET")
	protected.HandleFunc("/news/{id}", newsHandler.UpdateNews).Methods("PUT")
	protected.HandleFunc("/news/{id}", newsHandler.DeleteNews).Methods("DELETE")

	// Документы
	protected.HandleFunc("/documents", docHandler.UploadDocument).Methods("POST")
	protected.HandleFunc("/documents", docHandler.GetAllDocuments).Methods("GET")
	protected.HandleFunc("/documents/{id}", docHandler.DownloadDocument).Methods("GET")
	protected.HandleFunc("/documents/{id}", docHandler.DeleteDocument).Methods("DELETE")

	// Приложения
	protected.HandleFunc("/applications", appHandler.CreateApplication).Methods("POST")
	protected.HandleFunc("/applications", appHandler.GetAllApplications).Methods("GET")
	protected.HandleFunc("/applications/{id}", appHandler.GetApplicationByID).Methods("GET")
	protected.HandleFunc("/applications/{id}", appHandler.UpdateApplication).Methods("PUT")
	protected.HandleFunc("/applications/{id}", appHandler.DeleteApplication).Methods("DELETE")

	// Маршруты для администраторов
	adminRoute := protected.PathPrefix("/admin").Subrouter()
	adminRoute.Use(middleware.RoleMiddleware("admin"))
	adminRoute.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Добро пожаловать в админ-панель!"))
	}).Methods("GET")

	protected.HandleFunc("/logout", authHandler.Logout).Methods("POST")

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(r)

	server := &http.Server{Addr: ":8080", Handler: handler}

	go func() {
		log.Println("Сервер API запущен на порту http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Ошибка запуска сервера:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	log.Println("Сервер остановлен")
}
