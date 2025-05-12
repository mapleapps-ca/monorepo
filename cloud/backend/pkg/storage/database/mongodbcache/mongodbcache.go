package mongodbcache

import (
	"context"
	"time"

	"github.com/faabiosr/cachego"
	"github.com/faabiosr/cachego/mongo"
	mongo_client "go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	c "github.com/mapleapps-ca/monorepo/cloud/backend/config"
)

type Cacher interface {
	Shutdown(context.Context)
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte) error
	SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error
	Delete(ctx context.Context, key string) error
}

type cacheImpl struct {
	config CacheConfigurationProvider
	Client cachego.Cache
	Logger *zap.Logger
}

func NewProviderWithCustomConfig(
	config CacheConfigurationProvider,
	logger *zap.Logger,
	dbClient *mongo_client.Client,
) Cacher {
	logger.Debug("cache initializing...")

	cc := dbClient.Database(config.GetDatabaseName()).Collection("caches")

	c := mongo.New(cc)

	logger.Debug("cache initialized with mongodb as backend")
	return &cacheImpl{
		config: config,
		Client: c,
		Logger: logger,
	}
}

func NewProvider(
	appCfg *c.Configuration, //TODO: REPAIR!
	logger *zap.Logger,
	dbClient *mongo_client.Client,
) Cacher {
	logger.Debug("cache initializing...")

	cc := dbClient.Database(appCfg.DB.IAMName).Collection("caches")

	c := mongo.New(cc)

	logger.Debug("cache initialized with mongodb as backend")
	return &cacheImpl{
		Client: c,
		Logger: logger,
	}
}

func (s *cacheImpl) Shutdown(context.Context) {
	// Do nothing...
}

func (s *cacheImpl) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := s.Client.Fetch(key)
	if err != nil {
		s.Logger.Error("cache get failed", zap.Any("error", err))
		return nil, err
	}
	return []byte(val), nil
}

func (s *cacheImpl) Set(ctx context.Context, key string, val []byte) error {
	err := s.Client.Save(key, string(val), 0)
	if err != nil {
		s.Logger.Error("cache set failed", zap.Any("error", err))
		return err
	}
	return nil
}

func (s *cacheImpl) SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error {
	err := s.Client.Save(key, string(val), expiry)
	if err != nil {
		s.Logger.Error("cache set with expiry failed", zap.Any("error", err))
		return err
	}
	return nil
}

func (s *cacheImpl) Delete(ctx context.Context, key string) error {
	err := s.Client.Delete(key)
	if err != nil {
		s.Logger.Error("cache delete failed", zap.Any("error", err))
		return err
	}
	return nil
}
