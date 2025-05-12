// cloud/backend/internal/vault/service/encryptedfile/getbyid.go
package encryptedfile

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/usecase/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// GetEncryptedFileByIDService defines operations for retrieving an encrypted file by ID
type GetEncryptedFileByIDService interface {
	Execute(ctx context.Context, id primitive.ObjectID) (*domain.EncryptedFile, error)
}

type getEncryptedFileByIDServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	getByIDUseCase encryptedfile.GetEncryptedFileByIDUseCase
}

// NewGetEncryptedFileByIDService creates a new instance of the service
func NewGetEncryptedFileByIDService(
	config *config.Configuration,
	logger *zap.Logger,
	getByIDUseCase encryptedfile.GetEncryptedFileByIDUseCase,
) GetEncryptedFileByIDService {
	return &getEncryptedFileByIDServiceImpl{
		config:         config,
		logger:         logger.With(zap.String("component", "get-encrypted-file-by-id-service")),
		getByIDUseCase: getByIDUseCase,
	}
}

// Execute retrieves an encrypted file by its ID and verifies ownership
func (s *getEncryptedFileByIDServiceImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (*domain.EncryptedFile, error) {
	// Validate inputs (moved from usecase to service)
	if id.IsZero() {
		return nil, httperror.NewForBadRequestWithSingleField("id", "File ID cannot be empty")
	}

	// Get the file using the use case
	file, err := s.getByIDUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get encrypted file",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get encrypted file: %w", err)
	}

	if file == nil {
		return nil, httperror.NewForBadRequestWithSingleField("id", "File not found")
	}

	// Verify that the authenticated user has access to this file
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if ok && !userID.IsZero() && file.UserID != userID {
		s.logger.Warn("Unauthorized file access attempt",
			zap.String("file_id", id.Hex()),
			zap.String("file_owner", file.UserID.Hex()),
			zap.String("requester", userID.Hex()),
		)
		return nil, httperror.NewForForbiddenWithSingleField("message", "You do not have permission to access this file")
	}

	return file, nil
}
