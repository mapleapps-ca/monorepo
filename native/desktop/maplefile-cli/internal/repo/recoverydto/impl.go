// native/desktop/maplefile-cli/internal/repo/recoverydto/impl.go
package recoverydto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recoverydto"
)

// recoveryDTORepository implements the recoverydto.RecoveryDTORepository interface for cloud API calls
type recoveryDTORepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	httpClient    *http.Client
}

// NewRecoveryDTORepository creates a new repository for recovery cloud operations
func NewRecoveryDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
) recoverydto.RecoveryDTORepository {
	logger = logger.Named("RecoveryDTORepository")
	return &recoveryDTORepository{
		logger:        logger,
		configService: configService,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}
