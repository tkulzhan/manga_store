package databases

import (
	"context"
	"manga_store/internal/helpers"
	"manga_store/internal/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() {
	redisAddr := helpers.GetEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := helpers.GetEnv("REDIS_PASSWORD", "")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis: " + err.Error())
	} else {
		logger.Info("Connected to Redis")
	}
}

func Redis() *redis.Client {
	return redisClient
}

func CloseRedis() error {
	return redisClient.Close()
}
