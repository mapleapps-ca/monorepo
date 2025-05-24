// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/impl.go
package collectiondto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

// collectionDTORepository implements the collection.RemoteCollectionRepository interface
type collectionDTORepository struct {
	logger          *zap.Logger
	configService   config.ConfigService
	tokenRepository auth.TokenRepository
	httpClient      *http.Client
}

// NewCollectionDTORepository creates a new repository for collection operations
func NewCollectionDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository auth.TokenRepository,
) collectiondto.CollectionDTORepository {
	return &collectionDTORepository{
		logger:          logger,
		configService:   configService,
		tokenRepository: tokenRepository,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}
