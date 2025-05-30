// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondsharingdto/impl.go
package collectiondsharingdto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
)

// collectionDTORepository implements the collection.RemoteCollectionRepository interface
type collectionSharingDTORepository struct {
	logger          *zap.Logger
	configService   config.ConfigService
	tokenRepository auth.TokenRepository
	httpClient      *http.Client
}

// NewCollectionSharingDTORepository creates a new repository for collection sharing operations
func NewCollectionSharingDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository auth.TokenRepository,
) collectionsharingdto.CollectionSharingDTORepository {
	logger = logger.Named("CollectionSharingDTORepository")
	return &collectionSharingDTORepository{
		logger:          logger,
		configService:   configService,
		tokenRepository: tokenRepository,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}
