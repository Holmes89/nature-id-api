package connection

import (
	"github.com/go-redis/redis/v8"
	"os"
)

func getEnv(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}


type RedisConfig struct {
	URL string
	Username string
	Password string
}

func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		URL:     getEnv("REDIS_URL", "localhost:6379"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
}

func NewRedisClient(config RedisConfig) *redis.Client{
	return redis.NewClient(&redis.Options{
		Addr:               config.URL,
		Username:           config.Username,
		Password:           config.Password,
	})
}

func NewRedisClientDefault() *redis.Client {
	return NewRedisClient(LoadRedisConfig())
}