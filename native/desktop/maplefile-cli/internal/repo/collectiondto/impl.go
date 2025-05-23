// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/impl.go
package collectiondto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// collectionDTORepository implements the collection.RemoteCollectionRepository interface
type collectionDTORepository struct {
	logger         *zap.Logger
	configService  config.ConfigService
	userRepo       user.Repository
	tokenRefresher auth.TokenRefresher
	httpClient     *http.Client
}

// NewCollectionDTORepository creates a new repository for collection operations
func NewCollectionDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenRefresher auth.TokenRefresher,
) collectiondto.CollectionDTORepository {
	return &collectionDTORepository{
		logger:         logger,
		configService:  configService,
		userRepo:       userRepo,
		tokenRefresher: tokenRefresher,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}
