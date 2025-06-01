// monorepo/native/desktop/maplefile-cli/internal/repo/publiclookupdto/impl.go
package publiclookupdto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/publiclookupdto"
)

// publiclookupDTORepository implements the collection.PublicLookupDTORepository interface
type publicLookupDTORepository struct {
	logger          *zap.Logger
	configService   config.ConfigService
	tokenRepository auth.TokenRepository
	httpClient      *http.Client
}

// NewPublicLookupDTORepository creates a new repository for collection operations
func NewPublicLookupDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository auth.TokenRepository,
) publiclookupdto.PublicLookupDTORepository {
	logger = logger.Named("PublicLookupDTORepository")
	return &publicLookupDTORepository{
		logger:          logger,
		configService:   configService,
		tokenRepository: tokenRepository,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}
