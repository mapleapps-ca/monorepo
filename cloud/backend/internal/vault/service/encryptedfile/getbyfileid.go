// cloud/backend/internal/vault/service/encryptedfile/getbyfileid.go
package encryptedfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/usecase/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// GetEncryptedFileByFileIDService defines operations for retrieving an encrypted file by file ID
type GetEncryptedFileByFileIDService interface {
	Execute(ctx context.Context, userID primitive.ObjectID, fileID string) (*domain.EncryptedFile, error)
}

type getEncryptedFileByFileIDServiceImpl struct {
	config             *config.Configuration
	logger             *zap.Logger
	getByFileIDUseCase encryptedfile.GetEncryptedFileByFileIDUseCase
}

// NewGetEncryptedFileByFileIDService creates a new instance of the service
func NewGetEncryptedFileByFileIDService(
	config *config.Configuration,
	logger *zap.Logger,
	getByFileIDUseCase encryptedfile.GetEncryptedFileByFileIDUseCase,
) GetEncryptedFileByFileIDService {
	return &getEncryptedFileByFileIDServiceImpl{
		config:             config,
		logger:             logger.With(zap.String("component", "get-encrypted-file-by-file-id-service")),
		getByFileIDUseCase: getByFileIDUseCase,
	}
}

// Execute retrieves an encrypted file by user ID and file ID
func (s *getEncryptedFileByFileIDServiceImpl) Execute(
	ctx context.Context,
	userID primitive.ObjectID,
	fileID string,
) (*domain.EncryptedFile, error) {
	// Extract authenticated user ID from context if not provided
	if userID.IsZero() {
		contextUserID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
		if !ok || contextUserID.IsZero() {
			s.logger.Error("User ID not provided and not found in context")
			return nil, httperror.NewForBadRequestWithSingleField("user_id", "User ID is required")
		}
		userID = contextUserID
	}

	// Verify that the authenticated user has permission to access files for this user ID
	contextUserID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if ok && !contextUserID.IsZero() && contextUserID != userID {
		s.logger.Warn("Unauthorized file access attempt for another user's files",
			zap.String("requested_user_id", userID.Hex()),
			zap.String("authenticated_user_id", contextUserID.Hex()),
		)
		return nil, httperror.NewForForbiddenWithSingleField("message", "You do not have permission to access files for this user")
	}

	// Get the file using the use case
	return s.getByFileIDUseCase.Execute(ctx, userID, fileID)
}
