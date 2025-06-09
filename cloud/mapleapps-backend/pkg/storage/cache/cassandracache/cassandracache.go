package cassandracache

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

type Cacher interface {
	Shutdown()
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte) error
	SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error
	Delete(ctx context.Context, key string) error
	PurgeExpired(ctx context.Context) error
}

type cache struct {
	Session *gocql.Session
	Logger  *zap.Logger
}

func NewCache(session *gocql.Session, logger *zap.Logger) Cacher {
	logger = logger.Named("CassandraCache")
	logger.Info("cassandra cache initialized")
	return &cache{
		Session: session,
		Logger:  logger,
	}
}

func (s *cache) Shutdown() {
	s.Logger.Info("cassandra cache shutting down...")
	s.Session.Close()
}

func (s *cache) Get(ctx context.Context, key string) ([]byte, error) {
	var value []byte
	query := `SELECT value FROM pkg_cache_by_key_with_asc_expire_at WHERE key=? AND expires_at > toTimestamp(now()) LIMIT 1`
	err := s.Session.Query(query, key).WithContext(ctx).Consistency(gocql.LocalQuorum).Scan(&value)
	if err == gocql.ErrNotFound {
		return nil, nil
	}
	return value, err
}

func (s *cache) Set(ctx context.Context, key string, val []byte) error {
	return s.Session.Query(`INSERT INTO pkg_cache_by_key_with_asc_expire_at (key, expires_at, value) VALUES (?, toTimestamp(now()) + 86400000, ?)`,
		key, val).WithContext(ctx).Consistency(gocql.LocalQuorum).Exec()
}

func (s *cache) SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error {
	return s.Session.Query(`INSERT INTO pkg_cache_by_key_with_asc_expire_at (key, expires_at, value) VALUES (?, toTimestamp(now()) + ?, ?)`,
		key, int64(expiry/time.Millisecond), val).WithContext(ctx).Consistency(gocql.LocalQuorum).Exec()
}

func (s *cache) Delete(ctx context.Context, key string) error {
	return s.Session.Query(`DELETE FROM pkg_cache_by_key_with_asc_expire_at WHERE key=?`, key).
		WithContext(ctx).Consistency(gocql.LocalQuorum).Exec()
}

func (s *cache) PurgeExpired(ctx context.Context) error {
	return s.Session.Query(`DELETE FROM pkg_cache_by_key_with_asc_expire_at WHERE expires_at < toTimestamp(now())`).
		WithContext(ctx).Consistency(gocql.LocalQuorum).Exec()
}
