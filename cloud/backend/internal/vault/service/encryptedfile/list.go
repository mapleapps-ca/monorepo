// cloud/backend/internal/vault/service/encryptedfile/list.go
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

// ListEncryptedFilesService defines operations for listing all encrypted files for a user
type ListEncryptedFilesService interface {
	Execute(ctx context.Context, userID primitive.ObjectID) ([]*domain.EncryptedFile, error)
}

type listEncryptedFilesServiceImpl struct {
	config      *config.Configuration
	logger      *zap.Logger
	listUseCase encryptedfile.ListEncryptedFilesUseCase
}

// NewListEncryptedFilesService creates a new instance of the service
func NewListEncryptedFilesService(
	config *config.Configuration,
	logger *zap.Logger,
	listUseCase encryptedfile.ListEncryptedFilesUseCase,
) ListEncryptedFilesService {
	return &listEncryptedFilesServiceImpl{
		config:      config,
		logger:      logger.With(zap.String("component", "list-encrypted-files-service")),
		listUseCase: listUseCase,
	}
}

// Execute lists all encrypted files for a user
func (s *listEncryptedFilesServiceImpl) Execute(
	ctx context.Context,
	userID primitive.ObjectID,
) ([]*domain.EncryptedFile, error) {
	// Extract authenticated user ID from context if not provided
	if userID.IsZero() {
		contextUserID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
		if !ok || contextUserID.IsZero() {
			s.logger.Error("User ID not provided and not found in context")
			return nil, httperror.NewForBadRequestWithSingleField("user_id", "User ID is required")
		}
		userID = contextUserID
	}

	// Verify that the authenticated user has permission to list files for this user ID
	contextUserID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if ok && !contextUserID.IsZero() && contextUserID != userID {
		s.logger.Warn("Unauthorized attempt to list another user's files",
			zap.String("requested_user_id", userID.Hex()),
			zap.String("authenticated_user_id", contextUserID.Hex()),
		)
		return nil, httperror.NewForForbiddenWithSingleField("message", "You do not have permission to list files for this user")
	}

	// List the files using the use case
	return s.listUseCase.Execute(ctx, userID)
}
