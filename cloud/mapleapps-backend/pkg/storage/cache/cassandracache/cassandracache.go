package cassandracache



import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/faabiosr/cachego"
	"github.com/gocql/gocql"

	c "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
)

type Cacher interface {
	Shutdown()
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte) error
	SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error
	Delete(ctx context.Context, key string) error
}

type cache struct {
	Client cachego.Cache
	Logger *slog.Logger
}

func NewCache(cfg *c.Conf, logger *slog.Logger, session *gocql.Session) Cacher {
	logger.Debug("cassandra based cache initializing...")

	query := `
		CREATE TABLE IF NOT EXISTS mothership.cache_by_date (
			session_id TIMEUUID,
			expires_at TIMESTAMP,
			value BLOB,
			PRIMARY KEY ((expires_at), session_id)
		);
	`
	if err := session.Query(query).Exec(); err != nil {
		logger.Error("Failed creating `cache_by_date` table if DNE", slog.Any("err", err))
		log.Fatal(err)
	}

	logger.Debug("cassandra based cache initialized")
	return &cache{
		Logger: logger,
	}
}

func (s *cache) Shutdown() {
	// Do nothing...
}

func (s *cache) Get(ctx context.Context, key string) ([]byte, error) {
	// val, err := s.Client.Fetch(key)
	// if err != nil {
	// 	s.Logger.Error("cache get failed", slog.Any("error", err))
	// 	return nil, err
	// }
	// return []byte(val), nil
	return nil, nil
}

func (s *cache) Set(ctx context.Context, key string, val []byte) error {
	// err := s.Client.Save(key, string(val), 0)
	// if err != nil {
	// 	s.Logger.Error("cache set failed", slog.Any("error", err))
	// 	return err
	// }
	// return nil

	return nil
}

func (s *cache) SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error {
	// err := s.Client.Save(key, string(val), expiry)
	// if err != nil {
	// 	s.Logger.Error("cache set with expiry failed", slog.Any("error", err))
	// 	return err
	// }
	// return nil
	return nil
}

func (s *cache) Delete(ctx context.Context, key string) error {
	// err := s.Client.Delete(key)
	// if err != nil {
	// 	s.Logger.Error("cache delete failed", slog.Any("error", err))
	// 	return err
	// }
	// return nil
	return nil
}
