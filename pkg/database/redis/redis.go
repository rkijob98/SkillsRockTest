package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"task-manager/pkg/config"
	"task-manager/pkg/logger"
	"time"
)

type Client struct {
	client *redis.Client
	ttl    time.Duration
}

func New(cfg *config.Config) *Client {
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		logger.Get().Fatal("Failed to parse Redis URL",
			zap.Error(err),
			zap.String("url", cfg.Redis.URL),
		)
	}

	client := redis.NewClient(opt)

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Get().Fatal("Failed to connect to Redis",
			zap.Error(err),
			zap.String("address", opt.Addr),
			zap.Int("db", opt.DB),
		)
	}

	logger.Get().Info("Redis connected successfully",
		zap.String("address", opt.Addr),
		zap.Int("db", opt.DB),
		zap.Duration("default_ttl", cfg.Redis.TTL),
	)

	return &Client{
		client: client,
		ttl:    cfg.Redis.TTL,
	}
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *Client) Set(ctx context.Context, key string, value []byte) error {
	return c.client.Set(ctx, key, value, c.ttl).Err()
}

func (c *Client) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.SetEx(ctx, key, value, ttl).Err()
}

func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *Client) Close() error {
	return c.client.Close()
}
