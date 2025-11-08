.PHONY: help migrate-up migrate-down migrate-status build-server build-migrate build run-server run-migrate-up run-migrate-down run-migrate-status clean test deps setup start migrate-up-bin migrate-down-bin migrate-status-bin fmt vet lint

# Переменные
SERVER_DIR := server
MIGRATE_DIR := $(SERVER_DIR)/cmd/migrate
SERVER_CMD_DIR := $(SERVER_DIR)/cmd/goph-keeper
MIGRATIONS_DIR := $(SERVER_DIR)/migrations
BIN_DIR := bin

# Переменные окружения (можно переопределить)
DATABASE_URI ?= postgres://user:password@localhost:5432/keeper?sslmode=disable
RUN_ADDRESS ?= :8080
JWT_SECRET ?= your-secret-key-change-in-production

help: ## Показать справку по командам
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

# Миграции
migrate-up: ## Применить все миграции
	@echo "Применение миграций..."
	@cd $(SERVER_DIR) && go run cmd/migrate/main.go -command=up -path=migrations -database="$(DATABASE_URI)"

migrate-down: ## Откатить последнюю миграцию
	@echo "Откат последней миграции..."
	@cd $(SERVER_DIR) && go run cmd/migrate/main.go -command=down -path=migrations -database="$(DATABASE_URI)"

migrate-status: ## Показать статус миграций
	@echo "Статус миграций:"
	@cd $(SERVER_DIR) && go run cmd/migrate/main.go -command=status -path=migrations -database="$(DATABASE_URI)"

# Сборка
build-server: ## Собрать сервер
	@echo "Сборка сервера..."
	@mkdir -p $(BIN_DIR)
	@cd $(SERVER_DIR) && go build -o ../$(BIN_DIR)/goph-keeper cmd/goph-keeper/main.go
	@echo "Сервер собран: $(BIN_DIR)/goph-keeper"

build-migrate: ## Собрать утилиту миграций
	@echo "Сборка утилиты миграций..."
	@mkdir -p $(BIN_DIR)
	@cd $(SERVER_DIR) && go build -o ../$(BIN_DIR)/migrate cmd/migrate/main.go
	@echo "Утилита миграций собрана: $(BIN_DIR)/migrate"

build: build-server build-migrate ## Собрать все (сервер и утилиту миграций)

# Запуск
run-server: ## Запустить сервер
	@echo "Запуск сервера..."
	@cd $(SERVER_DIR) && DATABASE_URI="$(DATABASE_URI)" RUN_ADDRESS="$(RUN_ADDRESS)" JWT_SECRET="$(JWT_SECRET)" go run cmd/goph-keeper/main.go

run-migrate-up: ## Запустить миграции (up)
	@cd $(SERVER_DIR) && DATABASE_URI="$(DATABASE_URI)" go run cmd/migrate/main.go -command=up -path=migrations

run-migrate-down: ## Запустить миграции (down)
	@cd $(SERVER_DIR) && DATABASE_URI="$(DATABASE_URI)" go run cmd/migrate/main.go -command=down -path=migrations

run-migrate-status: ## Запустить миграции (status)
	@cd $(SERVER_DIR) && DATABASE_URI="$(DATABASE_URI)" go run cmd/migrate/main.go -command=status -path=migrations

# Утилиты
clean: ## Удалить собранные бинарники
	@echo "Удаление бинарников..."
	@rm -rf $(BIN_DIR)
	@echo "Бинарники удалены"

test: ## Запустить тесты
	@echo "Запуск тестов..."
	@cd $(SERVER_DIR) && go test ./...

fmt: ## Форматировать код
	@echo "Форматирование кода..."
	@cd $(SERVER_DIR) && go fmt ./...

vet: ## Проверить код с помощью go vet
	@echo "Проверка кода..."
	@cd $(SERVER_DIR) && go vet ./...

lint: fmt vet ## Форматировать и проверить код

deps: ## Установить зависимости
	@echo "Установка зависимостей..."
	@cd $(SERVER_DIR) && go mod download && go mod tidy

# Комбинации
setup: deps migrate-up ## Установить зависимости и применить миграции

start: build-server ## Собрать и запустить сервер из бинарника
	@echo "Запуск собранного сервера..."
	@DATABASE_URI="$(DATABASE_URI)" RUN_ADDRESS="$(RUN_ADDRESS)" JWT_SECRET="$(JWT_SECRET)" ./$(BIN_DIR)/goph-keeper

# Использование собранных бинарников
migrate-up-bin: build-migrate ## Применить миграции используя собранный бинарник
	@echo "Применение миграций (через бинарник)..."
	@DATABASE_URI="$(DATABASE_URI)" ./$(BIN_DIR)/migrate -command=up -path=$(SERVER_DIR)/migrations

migrate-down-bin: build-migrate ## Откатить миграции используя собранный бинарник
	@echo "Откат миграций (через бинарник)..."
	@DATABASE_URI="$(DATABASE_URI)" ./$(BIN_DIR)/migrate -command=down -path=$(SERVER_DIR)/migrations

migrate-status-bin: build-migrate ## Показать статус миграций используя собранный бинарник
	@echo "Статус миграций (через бинарник)..."
	@DATABASE_URI="$(DATABASE_URI)" ./$(BIN_DIR)/migrate -command=status -path=$(SERVER_DIR)/migrations

