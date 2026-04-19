// запоминает какие токены уже были использованы
package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() {
	// Получаем URL из переменных окружения или используем дефолтный
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379" // Дефолтный адрес
	}

	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Если не настроен через ENV
		Password: "",           // Пароль, если есть
		DB:       0,            // База данных
	})

	// Проверяем подключение
	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Ошибка подключения к Redis: %v\n", err)
		// В продакшене лучше exit(1), но здесь просто выводим ошибку
		return
	}

	RedisClient = client
	fmt.Println("Redis подключен успешно!")
}

func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}
