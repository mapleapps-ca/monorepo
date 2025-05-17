// internal/repo/remotefile/impl.go
package remotefile

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// remoteFileRepository implements remotefile.RemoteFileRepository
type remoteFileRepository struct {
	logger         *zap.Logger
	configService  config.ConfigService
	userRepo       user.Repository
	tokenRefresher auth.TokenRefresher
	httpClient     *http.Client
}

// NewRemoteFileRepository creates a new repository for remote file operations
func NewRemoteFileRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenRefresher auth.TokenRefresher,
) remotefile.RemoteFileRepository {
	return &remoteFileRepository{
		logger:         logger,
		configService:  configService,
		userRepo:       userRepo,
		tokenRefresher: tokenRefresher,
		httpClient:     &http.Client{Timeout: 60 * time.Second}, // Longer timeout for file operations
	}
}
