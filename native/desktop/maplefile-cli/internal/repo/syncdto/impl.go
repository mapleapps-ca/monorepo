// native/desktop/maplefile-cli/internal/repo/syncdto/impl.go
package syncdto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/syncdto"
)

// syncDTORepository implements the syncdto.SyncDTORepository interface
type syncDTORepository struct {
	logger          *zap.Logger
	configService   config.ConfigService
	tokenRepository dom_authdto.TokenDTORepository
	httpClient      *http.Client
}

// NewSyncDTORepository creates a new repository for syncdto operations
func NewSyncDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository dom_authdto.TokenDTORepository,
) syncdto.SyncDTORepository {
	logger = logger.Named("SyncDTORepository")
	return &syncDTORepository{
		logger:          logger,
		configService:   configService,
		tokenRepository: tokenRepository,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}
