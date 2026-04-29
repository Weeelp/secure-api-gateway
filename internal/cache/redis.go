// запоминает какие токены уже были использованы
package cache

import (
	"context"

	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	engine *redis.Client
}

func NewRedis(cfg *config.Config) *Redis {
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Проверяем подключение
	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Log.Error("Ошибка подключения к Redis: ", "error", err)
		return nil
	}

	logger.Log.Info("Redis подключен успешно!")
	return &Redis{
		engine: client,
	}
}

func (r *Redis) CloseRedis() {
	if r.engine != nil {
		r.engine.Close()
	}
}

func (r *Redis) GetEngine() *redis.Client {
	return r.engine
}
