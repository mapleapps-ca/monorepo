// native/desktop/maplefile-cli/internal/repo/recoverydto/impl.go
package recoverydto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
)

// recoveryDTORepository implements the recovery.RecoveryDTORepository interface for cloud API calls
type recoveryDTORepository struct {
	logger          *zap.Logger
	configService   config.ConfigService
	tokenRepository dom_authdto.TokenDTORepository
	httpClient      *http.Client
}

// NewRecoveryDTORepository creates a new repository for recovery cloud operations
func NewRecoveryDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository dom_authdto.TokenDTORepository,
) recovery.RecoveryDTORepository {
	logger = logger.Named("RecoveryDTORepository")
	return &recoveryDTORepository{
		logger:          logger,
		configService:   configService,
		tokenRepository: tokenRepository,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}
