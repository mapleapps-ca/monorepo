// cloud/maplefileapps-backend/pkg/storage/cache/cassandracache/cassandaracache.go
package cassandracache

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

type CassandraCacher interface {
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

func NewCassandraCacher(session *gocql.Session, logger *zap.Logger) CassandraCacher {
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
	var expiresAt time.Time

	query := `SELECT value, expires_at FROM pkg_cache_by_key_with_asc_expire_at WHERE key=?`
	err := s.Session.Query(query, key).WithContext(ctx).Consistency(gocql.LocalQuorum).Scan(&value, &expiresAt)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Check if expired in application code
	if time.Now().After(expiresAt) {
		// Entry is expired, delete it and return nil
		_ = s.Delete(ctx, key) // Clean up expired entry
		return nil, nil
	}

	return value, nil
}

func (s *cache) Set(ctx context.Context, key string, val []byte) error {
	expiresAt := time.Now().Add(24 * time.Hour) // Default 24 hour expiry
	return s.Session.Query(`INSERT INTO pkg_cache_by_key_with_asc_expire_at (key, expires_at, value) VALUES (?, ?, ?)`,
		key, expiresAt, val).WithContext(ctx).Consistency(gocql.LocalQuorum).Exec()
}

func (s *cache) SetWithExpiry(ctx context.Context, key string, val []byte, expiry time.Duration) error {
	expiresAt := time.Now().Add(expiry)
	return s.Session.Query(`INSERT INTO pkg_cache_by_key_with_asc_expire_at (key, expires_at, value) VALUES (?, ?, ?)`,
		key, expiresAt, val).WithContext(ctx).Consistency(gocql.LocalQuorum).Exec()
}

func (s *cache) Delete(ctx context.Context, key string) error {
	return s.Session.Query(`DELETE FROM pkg_cache_by_key_with_asc_expire_at WHERE key=?`,
		key).WithContext(ctx).Consistency(gocql.LocalQuorum).Exec()
}

func (s *cache) PurgeExpired(ctx context.Context) error {
	now := time.Now()

	// Thanks to the index on expires_at, this query is efficient
	iter := s.Session.Query(`SELECT key FROM pkg_cache_by_key_with_asc_expire_at WHERE expires_at < ? ALLOW FILTERING`,
		now).WithContext(ctx).Iter()

	var expiredKeys []string
	var key string
	for iter.Scan(&key) {
		expiredKeys = append(expiredKeys, key)
	}

	if err := iter.Close(); err != nil {
		return err
	}

	// Delete expired keys in batch
	if len(expiredKeys) > 0 {
		batch := s.Session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
		for _, expiredKey := range expiredKeys {
			batch.Query(`DELETE FROM pkg_cache_by_key_with_asc_expire_at WHERE key=?`, expiredKey)
		}
		return s.Session.ExecuteBatch(batch)
	}

	return nil
}
