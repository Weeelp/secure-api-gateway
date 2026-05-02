// запоминает какие токены уже были использованы
package cache

import (
	"context"

	"secure-api-gateway/internal/config"
	"secure-api-gateway/internal/logger"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	jwtEngine *redis.Client
	rlEngine  *redis.Client
	ipEngine  *redis.Client
}

func NewRedis(cfg *config.Config) *Redis {
	ctx := context.Background()

	jwtClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.JwtDB,
	})

	rlClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.RateLimitDB,
	})

	blClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.BlacklistDB,
	})

	if err := jwtClient.Ping(ctx).Err(); err != nil {
		logger.Log.Error("Ошибка jWT Redis: ", "error", err)
		return nil
	}
	if err := rlClient.Ping(ctx).Err(); err != nil {
		logger.Log.Error("Ошибка RateLimit Redis: ", "error", err)
		return nil
	}
	if err := blClient.Ping(ctx).Err(); err != nil {
		logger.Log.Error("Ошибка BlClient Redis: ", "error", err)
		return nil
	}

	logger.Log.Info("Redis подключен успешно!")
	return &Redis{
		jwtEngine: jwtClient,
		rlEngine:  rlClient,
		ipEngine:  blClient,
	}
}

func (r *Redis) CloseRedis() {
	if r.jwtEngine != nil {
		r.jwtEngine.Close()
	}
	if r.rlEngine != nil {
		r.rlEngine.Close()
	}
	if r.ipEngine != nil {
		r.ipEngine.Close()
	}
}

func (r *Redis) GetJwtEngine() *redis.Client {
	return r.jwtEngine
}

func (r *Redis) GetRlEngine() *redis.Client {
	return r.rlEngine
}

func (r *Redis) GetIPEngine() *redis.Client {
	return r.ipEngine
}
