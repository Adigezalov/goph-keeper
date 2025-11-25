package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Adigezalov/goph-keeper/internal/config"
	"github.com/Adigezalov/goph-keeper/internal/email"
	"github.com/Adigezalov/goph-keeper/internal/emailworker"
	"github.com/Adigezalov/goph-keeper/internal/health"
	"github.com/Adigezalov/goph-keeper/internal/localization"
	"github.com/Adigezalov/goph-keeper/internal/logger"
	"github.com/Adigezalov/goph-keeper/internal/middleware"
	"github.com/Adigezalov/goph-keeper/internal/realtime"
	"github.com/Adigezalov/goph-keeper/internal/repositories"
	"github.com/Adigezalov/goph-keeper/internal/secret"
	"github.com/Adigezalov/goph-keeper/internal/tokens"
	"github.com/Adigezalov/goph-keeper/internal/user"
	"github.com/Adigezalov/goph-keeper/internal/verification"

	_ "github.com/Adigezalov/goph-keeper/docs"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Goph-Keeper API
// @version 1.0
// @description API сервер для менеджера паролей Goph-Keeper
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@goph-keeper.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {token}

func main() {
	logger.Info("Запуск приложения")

	cfg := config.NewConfig()
	logger.Infof("JWT Secret длина: %d символов", len(cfg.JWTSecret))
	logger.Infof("Access Token TTL: %v", cfg.AccessTokenTTL)
	logger.Infof("Refresh Token TTL: %v", cfg.RefreshTokenTTL)
	logger.Infof("Verification Code TTL: %v", cfg.VerificationCodeTTL)
	logger.Infof("SMTP Host: %s:%s", cfg.SMTPHost, cfg.SMTPPort)

	dbRepo, err := repositories.NewDatabaseRepository(cfg.DatabaseURI)
	if err != nil {
		logger.Warnf("Предупреждение: Не удалось подключиться к базе данных: %v", err)
		dbRepo = nil
	}

	router := mux.NewRouter()

	router.Use(middleware.LoggingMiddleware)
	router.Use(localization.LanguageMiddleware)

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isWebSocket := r.Header.Get("Upgrade") == "websocket"

			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "http://localhost:3000"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Session-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if isWebSocket {
				next.ServeHTTP(w, r)
				return
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	api := router.PathPrefix("/api").Subrouter()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if dbRepo != nil {
		tokenRepo := tokens.NewDatabaseRepository(dbRepo.GetDB())
		tokenService := tokens.NewService(tokenRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

		authMiddleware := middleware.NewAuthMiddleware(tokenService)

		realtimeHub := realtime.NewHub()
		realtimeService := realtime.NewService(realtimeHub)
		realtimeHandler := realtime.NewHandler(realtimeHub, tokenService)

		api.HandleFunc("/v1/realtime", realtimeHandler.HandleWebSocket)

		healthService := health.NewService()
		healthHandler := health.NewHandler(healthService)
		healthRoutes := api.PathPrefix("/").Subrouter()

		healthRoutes.HandleFunc("/v1/health", authMiddleware.RequireAuth(healthHandler.Check)).Methods("PATCH")

		// Инициализация email сервиса и воркера
		smtpService := email.NewService(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.SMTPFrom,
		)

		// Создаем email worker с очередью на 100 задач, 3 попытки, задержка 5 секунд
		emailWorker := emailworker.NewWorker(smtpService, 100, 3, 5*time.Second)
		emailWorker.Start()
		defer emailWorker.Stop()

		// Создаем сервис для интеграции с user service
		emailService := emailworker.NewEmailWorkerService(emailWorker)

		verificationRepo := verification.NewRepository(dbRepo.GetDBX())
		userRepo := user.NewDatabaseRepository(dbRepo.GetDB())
		userService := user.NewService(
			userRepo,
			tokenService,
			emailService,
			verificationRepo,
			cfg.VerificationCodeTTL,
		)
		userHandler := user.NewHandler(userService, cfg.RefreshTokenTTL)
		userRoutes := api.PathPrefix("/v1/user").Subrouter()

		userRoutes.HandleFunc("/register", userHandler.Register).Methods("POST")
		userRoutes.HandleFunc("/verify-email", userHandler.VerifyEmail).Methods("POST")
		userRoutes.HandleFunc("/resend-code", userHandler.ResendCode).Methods("POST")
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

	// Swagger UI
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Infof("Сервер запущен на %s", cfg.ServerAddress)
		logger.Infof("База данных: %s", cfg.DatabaseURI)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	<-sigChan
	logger.Info("Получен сигнал завершения, останавливаем сервер...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Ошибка при остановке сервера: %v", err)
	} else {
		logger.Info("Сервер успешно остановлен")
	}
}
