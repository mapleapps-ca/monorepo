// monorepo/native/desktop/maplefile-cli/internal/repo/remotecollection/impl.go
package remotecollection

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// collectionRepository implements the collection.RemoteCollectionRepository interface
type collectionRepository struct {
	logger         *zap.Logger
	configService  config.ConfigService
	userRepo       user.Repository
	tokenRefresher auth.TokenRefresher
	httpClient     *http.Client
}

// NewRemoteCollectionRepository creates a new repository for collection operations
func NewRemoteCollectionRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenRefresher auth.TokenRefresher,
) collection.RemoteCollectionRepository {
	return &collectionRepository{
		logger:         logger,
		configService:  configService,
		userRepo:       userRepo,
		tokenRefresher: tokenRefresher,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}
