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
	"github.com/Adigezalov/goph-keeper/internal/repositories"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/Adigezalov/goph-keeper/internal/user"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Запуск приложения")

	// Загружаем конфигурацию
	cfg := config.NewConfig()

	// Создаем подключение к базе данных
	dbRepo, err := repositories.NewDatabaseRepository(cfg.DatabaseURI)
	if err != nil {
		log.Printf("Предупреждение: Не удалось подключиться к базе данных: %v", err)
		dbRepo = nil // Продолжаем работу без БД
	}

	// Создаем роутер
	router := mux.NewRouter()

	// Настраиваем маршруты
	api := router.PathPrefix("/api").Subrouter()

	// Health check endpoint (публичный)
	healthService := health.NewService()
	healthHandler := health.NewHandler(healthService)
	api.HandleFunc("/v1/health", healthHandler.Check).Methods("PATCH")

	log.Println("Зарегистрированы публичные health check маршруты")

	// Контекст для управления жизненным циклом приложения
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if dbRepo != nil {
		// Создаем репозитории
		userRepo := user.NewDatabaseRepository(dbRepo.GetDB())
		tokenRepo := tokens.NewDatabaseRepository(dbRepo.GetDB())

		// Создаем сервисы
		tokenService := tokens.NewService(tokenRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
		userService := user.NewService(userRepo, tokenService)

		// Создаем middleware
		authMiddleware := middleware.NewAuthMiddleware(tokenService)

		// Создаем handlers
		userHandler := user.NewHandler(userService, cfg.RefreshTokenTTL)

		// Пользовательские маршруты
		userRoutes := api.PathPrefix("/v1/user").Subrouter()
		userRoutes.HandleFunc("/register", userHandler.Register).Methods("POST")
		userRoutes.HandleFunc("/login", userHandler.Login).Methods("POST")
		userRoutes.HandleFunc("/refresh", userHandler.Refresh).Methods("GET")
		userRoutes.HandleFunc("/logout", userHandler.Logout).Methods("GET")
		userRoutes.HandleFunc("/logout-all", authMiddleware.RequireAuth(userHandler.LogoutAll)).Methods("GET")

		log.Println("Зарегистрированы пользовательские и защищенные маршруты")
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
