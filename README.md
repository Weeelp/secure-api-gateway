# Secure API Gateway

Этот проект — шлюз для микросервисной архитектуры. Помимо проксирования, он распределяет нагрузку и выполняет функции комплексной проверки безопасности.

## Ключевые возможности

*   Smart Load Balancing: Балансировщик Round Robin равномерно распределяет запросы между указанными бэкендами.
*   Decentralized Health Checks: Каждый бэкенд имеет свою фоновую горутину-чекер. При падении сервера шлюз мгновенно перестает слать туда трафик.
*   JWT Security & Replay Protection: Валидация токенов по секретному ключу и Issuer. Каждый токен (JTI) регистрируется в Redis для защиты от повторного использования.
*   Multi-Level Rate Limiting: Лимиты запросов для пользователей, хранящиеся в отдельном блоке Redis.
*   Dynamic IP Blacklist: Блокировка нежелательных IP-адресов через выделенную базу Redis.
*   Payload Security: Контроль размера тела запроса для предотвращения перегрузки серверов.
*   Logging: логирование с RequestID для сквозного отслеживания запросов.

## 🚀 Быстрый старт
1. # Вариант 1
```bash
docker-compose up --build
```

2. # Вариант 2
### 1. Настройка окружения
Создайте файл `.env` в корне проекта:

```env
# Порт самого шлюза
PORT=8080

# Список бэкендов (указывать через запятую без пробелов)
BACKEND_URL=http://localhost:9090,http://localhost:9091

# Секрет для JWT (обязательно)
JWT_SECRET="YOUR_KEY"

# Конфигурация Redis (3 независимых блока)
REDIS_ADDR=localhost:6379
REDIS_PASS=
JWT_DB=0
RATE_LIMIT_DB=1
BLACKLIST_DB=2
```

### 2. Сборка и запуск
```bash
go run cmd/main.go
```

### 3. Проверка работоспособности
```bash
curl -i http://localhost:8080/health
```

## 🛠 Технологический стек

*   Language: Go (Golang)
*   Storage: Redis
*   Patterns: Middleware, Singleton, Factory
*   Concurrency: `sync.RWMutex`, `sync/atomic`, `goroutines` и `context`
*   Logger: Логирование на базе context и logrus

---
*<small> В процессе разработки ни один мьютекс не привел к дедлоку. </small>*
