package twotiercache

import (
	"context"
	"time"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/cassandracache"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/memory/redis"
	"go.uber.org/zap"
)

type TwoTierCacher interface {
	Shutdown(ctx context.Context)
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte) error
	SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error
	Delete(ctx context.Context, key string) error
	PurgeExpired(ctx context.Context) error
}

// twoTierCacheImpl: clean 2-layer (read-through write-through) cache
//
// L1: Redis (fast, in-memory)
// L2: Cassandra (persistent)
//
// On Get: check Redis → then Cassandra → if found in Cassandra → populate Redis
// On Set: write to both
// On SetWithExpiry: write to both with expiry
// On Delete: remove from both
type twoTierCacheImpl struct {
	RedisCache     redis.Cacher
	CassandraCache cassandracache.CassandraCacher
	Logger         *zap.Logger
}

func NewTwoTierCache(redisCache redis.Cacher, cassandraCache cassandracache.CassandraCacher, logger *zap.Logger) TwoTierCacher {
	logger = logger.Named("TwoTierCache")
	return &twoTierCacheImpl{
		RedisCache:     redisCache,
		CassandraCache: cassandraCache,
		Logger:         logger,
	}
}

func (c *twoTierCacheImpl) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.RedisCache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		c.Logger.Debug("cache hit from Redis", zap.String("key", key))
		return val, nil
	}

	val, err = c.CassandraCache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		c.Logger.Debug("cache hit from Cassandra, writing back to Redis", zap.String("key", key))
		_ = c.RedisCache.Set(ctx, key, val)
	}
	return val, nil
}

func (c *twoTierCacheImpl) Set(ctx context.Context, key string, val []byte) error {
	if err := c.RedisCache.Set(ctx, key, val); err != nil {
		return err
	}
	if err := c.CassandraCache.Set(ctx, key, val); err != nil {
		return err
	}
	return nil
}

func (c *twoTierCacheImpl) SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error {
	if err := c.RedisCache.SetWithExpiry(ctx, key, val, expiry); err != nil {
		return err
	}
	if err := c.CassandraCache.SetWithExpiry(ctx, key, val, expiry); err != nil {
		return err
	}
	return nil
}

func (c *twoTierCacheImpl) Delete(ctx context.Context, key string) error {
	if err := c.RedisCache.Delete(ctx, key); err != nil {
		return err
	}
	if err := c.CassandraCache.Delete(ctx, key); err != nil {
		return err
	}
	return nil
}

func (c *twoTierCacheImpl) PurgeExpired(ctx context.Context) error {
	return c.CassandraCache.PurgeExpired(ctx)
}

func (c *twoTierCacheImpl) Shutdown(ctx context.Context) {
	c.Logger.Info("two-tier cache shutting down...")
	c.RedisCache.Shutdown(ctx)
	c.CassandraCache.Shutdown()
	c.Logger.Info("two-tier cache shutdown complete")
}
