// native/desktop/maplefile-cli/internal/repo/syncdto/impl.go
package syncdto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// syncRepository implements the syncdto.SyncDTORepository interface
type syncRepository struct {
	logger          *zap.Logger
	configService   config.ConfigService
	tokenRepository auth.TokenRepository
	httpClient      *http.Client
}

// NewSyncDTORepository creates a new repository for syncdto operations
func NewSyncDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository auth.TokenRepository,
) syncdto.SyncDTORepository {
	return &syncRepository{
		logger:          logger,
		configService:   configService,
		tokenRepository: tokenRepository,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}
