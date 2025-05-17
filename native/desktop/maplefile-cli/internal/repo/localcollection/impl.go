// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/impl.go
package collection

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// localcollectionRepository implements the localcollection.CollectionRepository interface
type localcollectionRepository struct {
	logger         *zap.Logger
	configService  config.ConfigService
	userRepo       user.Repository
	tokenRefresher auth.TokenRefresher
	httpClient     *http.Client
}

// NewCollectionRepository creates a new repository for localcollection operations
func NewLocalCollectionRepository(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenRefresher auth.TokenRefresher,
) localcollection.LocalCollectionRepository {
	return &localcollectionRepository{
		logger:         logger,
		configService:  configService,
		userRepo:       userRepo,
		tokenRefresher: tokenRefresher,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}
