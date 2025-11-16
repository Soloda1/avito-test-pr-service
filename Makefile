# Makefile для проекта avito-test-pr-service

APP_NAME=pr-service
MIGRATOR_NAME=pr-migrator
BINARY_DIR=bin
GO_VERSION=1.24.2
DOCKER_IMAGE=avito-pr-service:latest
DOCKER_COMPOSE_FILE=docker-compose.yml
PSQL_CONTAINER=pr-service-db
DB_USER=postgres
DB_NAME=prservice
DB_PORT=5432

# Путь к main файлам
SERVER_MAIN=./cmd/server/main.go
MIGRATOR_MAIN=./cmd/migrate/main.go

# Флаги
LDFLAGS=-s -w
TEST_FLAGS=-count=1
RACE_FLAGS=-race

.PHONY: help check-go-version fmt build run migrate-up migrate-down up down restart logs db-shell psql test test-race coverage clean

help:
	@echo "Доступные цели:"
	@echo "  check-go-version    - Проверить установленную версию Go"
	@echo "  fmt                 - Форматирование, go vet и go mod tidy"
	@echo "  build               - Сборка бинарника сервера"
	@echo "  run                 - Запуск сервера локально (go run)"
	@echo "  migrate-up          - Применить миграции (go run мигратора)"
	@echo "  migrate-down        - Откатить миграции (go run мигратора)"
	@echo "  up                  - Запуск docker-compose инфраструктуры"
	@echo "  down                - Остановка docker-compose инфраструктуры"
	@echo "  restart             - Перезапуск контейнера сервиса"
	@echo "  logs                - Живые логи сервиса"
	@echo "  db-shell            - Shell в контейнер базы данных"
	@echo "  psql                - psql подключение к БД"
	@echo "  test                - Запуск тестов"
	@echo "  test-race           - Тесты с -race"
	@echo "  coverage            - Отчёт покрытия (HTML)"
	@echo "  clean               - Очистка бинарников и кешей"

check-go-version:
	@echo "🔍 Проверка версии Go..."
	@go version | grep -q "go$(GO_VERSION)" || (echo "❌ Требуется Go $(GO_VERSION)" && exit 1)
	@echo "✅ Go $(GO_VERSION) найден"

fmt: check-go-version
	@echo "🧹 gofmt + go fmt + go vet + go mod tidy"
	@gofmt -s -w .
	@go fmt ./...
	@go vet ./...
	@go mod tidy
	@echo "✅ fmt/vet/tidy завершены"

build: check-go-version
	@echo "🔨 Сборка сервера..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/$(APP_NAME) -ldflags "$(LDFLAGS)" $(SERVER_MAIN)
	@echo "✅ Бинарник: $(BINARY_DIR)/$(APP_NAME)"

run: check-go-version
	@echo "🚀 Запуск сервера (go run)..."
	@go run $(SERVER_MAIN)

migrate-up: check-go-version
	@echo "🚀 Применение миграций..."
	@go run $(MIGRATOR_MAIN) -command up

migrate-down: check-go-version
	@echo "🔄 Откат миграций..."
	@go run $(MIGRATOR_MAIN) -command down

up:
	@echo "🚀 docker-compose up -d"
	@docker compose -f $(DOCKER_COMPOSE_FILE) up -d --build
	@until docker exec $(PSQL_CONTAINER) pg_isready -U $(DB_USER) -p $(DB_PORT); do \
    		echo "⏳ Ждем готовности Postgres..."; \
    		sleep 1; \
    	done
down:
	@echo "🛑 docker-compose down"
	@docker compose -f $(DOCKER_COMPOSE_FILE) down

restart:
	@echo "🔄 Перезапуск контейнера сервиса..."
	@docker compose -f $(DOCKER_COMPOSE_FILE) restart pr-service

logs:
	@echo "📄 Логи сервиса... (Ctrl+C для выхода)"
	@docker compose -f $(DOCKER_COMPOSE_FILE) logs -f pr-service

db-shell:
	@echo "🐚 Вход в контейнер базы данных..."
	@docker exec -it $(PSQL_CONTAINER) sh

psql:
	@echo "💾 Подключение psql..."
	@docker exec -it $(PSQL_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

test:
	@echo "Запуск всех тестов (unit + integration)..."
	go test ./... -v -count=1

test-integration:
	@echo "Запуск интеграционных тестов..."
	go test ./internal/tests/integration -v -count=1

test-race: check-go-version
	@echo "🧪 Запуск тестов (race)..."
	@go test $(TEST_FLAGS) $(RACE_FLAGS) ./...

coverage: check-go-version
	@echo "🧪 Покрытие..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep -E 'total'
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ coverage.html готов"

clean:
	@echo "🧹 Очистка..."
	@go clean -cache -testcache -modcache
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "✅ Очистка завершена"
