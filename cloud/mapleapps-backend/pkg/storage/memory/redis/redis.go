package redis

import (
	"context"
	"errors"
	"time"

	c "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Cacher interface {
	Shutdown(ctx context.Context)
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte) error
	SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error
	Delete(ctx context.Context, key string) error
}

type cache struct {
	Client *redis.Client
	Logger *zap.Logger
}

func NewCache(cfg *c.Configuration, logger *zap.Logger) Cacher {
	logger = logger.Named("Redis Memory Storage")

	opt, err := redis.ParseURL(cfg.Cache.URI)
	if err != nil {
		logger.Fatal("failed parsing Redis URL", zap.Error(err))
	}
	rdb := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err = rdb.Ping(ctx).Result(); err != nil {
		logger.Fatal("failed connecting to Redis", zap.Error(err))
	}

	return &cache{
		Client: rdb,
		Logger: logger,
	}
}

func (s *cache) Shutdown(ctx context.Context) {
	s.Logger.Info("shutting down Redis cache...")
	s.Client.Close()
}

func (s *cache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := s.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return []byte(val), err
}

func (s *cache) Set(ctx context.Context, key string, val []byte) error {
	return s.Client.Set(ctx, key, val, 0).Err()
}

func (s *cache) SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error {
	return s.Client.Set(ctx, key, val, expiry).Err()
}

func (s *cache) Delete(ctx context.Context, key string) error {
	return s.Client.Del(ctx, key).Err()
}
