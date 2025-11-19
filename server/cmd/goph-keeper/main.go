package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/config"
	"github.com/Adigezalov/goph-keeper/internal/health"
	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/Adigezalov/goph-keeper/internal/realtime"
	"github.com/Adigezalov/goph-keeper/internal/repositories"
	"github.com/Adigezalov/goph-keeper/internal/secret"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/Adigezalov/goph-keeper/internal/user"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Запуск приложения")

	// Загружаем конфигурацию
	cfg := config.NewConfig()
	log.Printf("JWT Secret длина: %d символов", len(cfg.JWTSecret))
	log.Printf("Access Token TTL: %v", cfg.AccessTokenTTL)
	log.Printf("Refresh Token TTL: %v", cfg.RefreshTokenTTL)

	// Создаем подключение к базе данных
	dbRepo, err := repositories.NewDatabaseRepository(cfg.DatabaseURI)
	if err != nil {
		log.Printf("Предупреждение: Не удалось подключиться к базе данных: %v", err)
		dbRepo = nil // Продолжаем работу без БД
	}

	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "http://localhost:3000"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Обработка preflight запросов
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	api := router.PathPrefix("/api").Subrouter()

	// Контекст для управления жизненным циклом приложения
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if dbRepo != nil {
		tokenRepo := tokens.NewDatabaseRepository(dbRepo.GetDB())
		tokenService := tokens.NewService(tokenRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

		authMiddleware := middleware.NewAuthMiddleware(tokenService)

		realtimeHub := realtime.NewHub()
		realtimeService := realtime.NewService(realtimeHub)
		realtimeHandler := realtime.NewHandler(realtimeHub, tokenService)
		realtimeRoutes := api.PathPrefix("v1/realtime").Subrouter()

		realtimeRoutes.HandleFunc("", realtimeHandler.HandleWebSocket)

		healthService := health.NewService()
		healthHandler := health.NewHandler(healthService)
		healthRoutes := api.PathPrefix("/").Subrouter()

		healthRoutes.HandleFunc("/v1/health", authMiddleware.RequireAuth(healthHandler.Check)).Methods("PATCH")

		userRepo := user.NewDatabaseRepository(dbRepo.GetDB())
		userService := user.NewService(userRepo, tokenService)
		userHandler := user.NewHandler(userService, cfg.RefreshTokenTTL)
		userRoutes := api.PathPrefix("/v1/user").Subrouter()

		userRoutes.HandleFunc("/register", userHandler.Register).Methods("POST")
		userRoutes.HandleFunc("/login", userHandler.Login).Methods("POST")
		userRoutes.HandleFunc("/refresh", userHandler.Refresh).Methods("GET")
		userRoutes.HandleFunc("/logout", userHandler.Logout).Methods("GET")
		userRoutes.HandleFunc("/logout-all", authMiddleware.RequireAuth(userHandler.LogoutAll)).Methods("GET")

		secretRepo := secret.NewDatabaseRepository(dbRepo.GetDB())
		secretService := secret.NewService(secretRepo)
		secretService.SetRealtimeService(realtimeService)
		secretHandler := secret.NewHandler(secretService)
		secretRoutes := api.PathPrefix("/v1/secrets").Subrouter()

		secretRoutes.HandleFunc("", authMiddleware.RequireAuth(secretHandler.GetAll)).Methods("GET")
		secretRoutes.HandleFunc("", authMiddleware.RequireAuth(secretHandler.Create)).Methods("POST")
		secretRoutes.HandleFunc("/sync", authMiddleware.RequireAuth(secretHandler.Sync)).Methods("GET")
		secretRoutes.HandleFunc("/{id}", authMiddleware.RequireAuth(secretHandler.Get)).Methods("GET")
		secretRoutes.HandleFunc("/{id}", authMiddleware.RequireAuth(secretHandler.Update)).Methods("PUT")
		secretRoutes.HandleFunc("/{id}", authMiddleware.RequireAuth(secretHandler.Delete)).Methods("DELETE")

		secretRoutes.HandleFunc("/chunks/init", authMiddleware.RequireAuth(secretHandler.InitChunkedUpload)).Methods("POST")
		secretRoutes.HandleFunc("/{id}/chunks", authMiddleware.RequireAuth(secretHandler.UploadChunk)).Methods("POST")
		secretRoutes.HandleFunc("/{id}/chunks/finalize", authMiddleware.RequireAuth(secretHandler.FinalizeChunkedUpload)).Methods("POST")
		secretRoutes.HandleFunc("/{id}/chunks/{chunkIndex}", authMiddleware.RequireAuth(secretHandler.DownloadChunk)).Methods("GET")
	}

	// Настраиваем graceful shutdown
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	// Канал для получения сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Сервер запущен на %s", cfg.ServerAddress)
		log.Printf("База данных: %s", cfg.DatabaseURI)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ждем сигнал завершения
	<-sigChan
	log.Println("Получен сигнал завершения, останавливаем сервер...")

	// Останавливаем accrual worker
	cancel()

	// Останавливаем HTTP сервер
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	} else {
		log.Println("Сервер успешно остановлен")
	}
}
