.PHONY: dev dev-backend dev-frontend build

# Запустить backend локально (polling + HTTP API на :8080)
dev-backend:
	cd backend && go run ./cmd/bot/

# Запустить frontend локально (Vite dev server на :5173, proxy -> :8080)
dev-frontend:
	cd frontend && npm run dev

# Запустить оба одновременно (требует macOS или Linux с поддержкой & в sh)
dev:
	@echo "Запускаю backend на :8080 и frontend на :5173..."
	@cd backend && go run ./cmd/bot/ &
	@cd frontend && npm run dev

# Собрать frontend для деплоя (VITE_API_URL должен быть выставлен)
build-frontend:
	cd frontend && npm run build

# Собрать backend бинарь
build-backend:
	cd backend && go build -o ../bin/bot ./cmd/bot/
