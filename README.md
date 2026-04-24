# secure-api-gateway
## Быстрый старт
1. **Создать .env файл в корне проекта:**

PORT=8080
BACKEND_URL=http://localhost:9090

1. **Запустить приложение:**

go run cmd/main.go

1. **Тест шлюза:**

curl -i http://localhost:8080/health
