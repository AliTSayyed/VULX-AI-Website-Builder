package cache

/*
all cache supporitng operations using redis db
*/
import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(cfg config.Redis) *RedisClient {
	dbName, err := strconv.Atoi(cfg.Name)
	if err != nil {
		panic(fmt.Errorf("invalid redis cache name: %w", err))
	}

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:6379", cfg.Host),
		DB:   dbName,
	})

	return &RedisClient{
		Client: client,
	}
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	result, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return "", domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to get cahce key %s: %w", key, err))
	}
	return result, nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err := r.Client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to set cahce key %s: %w", key, err))
	}
	return nil
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		return domain.NewError(domain.ErrorTypeInternal, fmt.Errorf("failed to delete cache key %s: %w", key, err))
	}
	return nil
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	count := r.Client.Exists(ctx, key).Val()
	return count > 0, nil
}
