package distributedmutex

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Adapter provides interface for abstracting distributedmutex generation.
type Adapter interface {
	Acquire(ctx context.Context, key string)
	Acquiref(ctx context.Context, format string, a ...any)
	Release(ctx context.Context, key string)
	Releasef(ctx context.Context, format string, a ...any)
}

type distributedLockerAdapter struct {
	Logger        *zap.Logger
	Redis         redis.UniversalClient
	Locker        *redislock.Client
	LockInstances map[string]*redislock.Lock
	Mutex         *sync.Mutex // Add a mutex for synchronization with goroutines
}

// NewAdapter constructor that returns the default DistributedLocker generator.
func NewAdapter(loggerp *zap.Logger, redisClient redis.UniversalClient) Adapter {
	loggerp.Debug("distributed mutex starting and connecting...")

	// Create a new lock client.
	locker := redislock.New(redisClient)

	loggerp.Debug("distributed mutex initialized")

	return distributedLockerAdapter{
		Logger:        loggerp,
		Redis:         redisClient,
		Locker:        locker,
		LockInstances: make(map[string]*redislock.Lock, 0),
		Mutex:         &sync.Mutex{}, // Initialize the mutex
	}
}

// Acquire function blocks the current thread if the lock key is currently locked.
func (a distributedLockerAdapter) Acquire(ctx context.Context, k string) {
	startDT := time.Now()
	a.Logger.Debug(fmt.Sprintf("locking for key: %v", k))

	// Retry every 250ms, for up-to 20x
	backoff := redislock.LimitRetry(redislock.LinearBackoff(250*time.Millisecond), 20)

	// Obtain lock with retry
	lock, err := a.Locker.Obtain(ctx, k, time.Minute, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err == redislock.ErrNotObtained {
		nowDT := time.Now()
		diff := nowDT.Sub(startDT)
		a.Logger.Error("could not obtain lock",
			zap.String("key", k),
			zap.Time("start_dt", startDT),
			zap.Time("now_dt", nowDT),
			zap.Any("duration_in_minutes", diff.Minutes()))
		return
	} else if err != nil {
		a.Logger.Error("failed obtaining lock",
			zap.String("key", k),
			zap.Any("error", err),
		)
		return
	}

	// DEVELOPERS NOTE:
	// The `map` datastructure in Golang is not concurrently safe, therefore we
	// need to use mutex to coordinate access of our `LockInstances` map
	// resource between all the goroutines.
	a.Mutex.Lock()
	defer a.Mutex.Unlock()

	if a.LockInstances != nil { // Defensive code.
		a.LockInstances[k] = lock
	}
}

// Acquiref function blocks the current thread if the lock key is currently locked.
func (u distributedLockerAdapter) Acquiref(ctx context.Context, format string, a ...any) {
	k := fmt.Sprintf(format, a...)
	u.Acquire(ctx, k)
	return
}

// Release function blocks the current thread if the lock key is currently locked.
func (a distributedLockerAdapter) Release(ctx context.Context, k string) {
	a.Logger.Debug(fmt.Sprintf("unlocking for key: %v", k))

	lockInstance, ok := a.LockInstances[k]
	if ok {
		defer lockInstance.Release(ctx)
	} else {
		a.Logger.Error("could not obtain to unlock", zap.String("key", k))
	}
	return
}

// Releasef
func (u distributedLockerAdapter) Releasef(ctx context.Context, format string, a ...any) {
	k := fmt.Sprintf(format, a...) //TODO: https://github.com/bsm/redislock/blob/main/README.md
	u.Release(ctx, k)
	return
}
