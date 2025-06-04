// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/impl.go
package filedto

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// fileDTORepository implements the FileDTORepository interface
type fileDTORepository struct {
	logger        *zap.Logger
	configService config.ConfigService
	tokenRepo     dom_authdto.TokenRepository
	httpClient    *http.Client
}

// NewFileDTORepository creates a new repository for cloud file DTO operations
func NewFileDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepo dom_authdto.TokenRepository,
) filedto.FileDTORepository {
	logger = logger.Named("FileDTORepository")
	return &fileDTORepository{
		logger:        logger.With(zap.String("repository", "filedto")),
		configService: configService,
		tokenRepo:     tokenRepo,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}
