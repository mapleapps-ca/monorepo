// native/desktop/maplefile-cli/internal/repo/medto/impl.go
package medto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/medto"
)

// meDTORepository implements the medto.MeDTORepository interface
type meDTORepository struct {
	logger          *zap.Logger
	configService   config.ConfigService
	tokenRepository dom_authdto.TokenDTORepository
	httpClient      *http.Client
}

// NewMeDTORepository creates a new repository for me operations
func NewMeDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository dom_authdto.TokenDTORepository,
) medto.MeDTORepository {
	logger = logger.Named("MeDTORepository")
	return &meDTORepository{
		logger:          logger,
		configService:   configService,
		tokenRepository: tokenRepository,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}
