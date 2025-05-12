// cloud/backend/internal/vault/service/encryptedfile/delete.go
package encryptedfile

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/usecase/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// DeleteEncryptedFileService defines operations for deleting an encrypted file
type DeleteEncryptedFileService interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
}

type deleteEncryptedFileServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	getByIDUseCase encryptedfile.GetEncryptedFileByIDUseCase
	deleteUseCase  encryptedfile.DeleteEncryptedFileUseCase
}

// NewDeleteEncryptedFileService creates a new instance of the service
func NewDeleteEncryptedFileService(
	config *config.Configuration,
	logger *zap.Logger,
	getByIDUseCase encryptedfile.GetEncryptedFileByIDUseCase,
	deleteUseCase encryptedfile.DeleteEncryptedFileUseCase,
) DeleteEncryptedFileService {
	return &deleteEncryptedFileServiceImpl{
		config:         config,
		logger:         logger.With(zap.String("component", "delete-encrypted-file-service")),
		getByIDUseCase: getByIDUseCase,
		deleteUseCase:  deleteUseCase,
	}
}

// Execute deletes an encrypted file after verifying ownership
func (s *deleteEncryptedFileServiceImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	// Validate inputs (moved from usecase to service)
	if id.IsZero() {
		return httperror.NewForBadRequestWithSingleField("id", "File ID cannot be empty")
	}

	// First get the file to verify ownership
	file, err := s.getByIDUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get file for deletion",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get file for deletion: %w", err)
	}

	if file == nil {
		return httperror.NewForBadRequestWithSingleField("id", "File not found")
	}

	// Verify that the authenticated user has access to this file
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if ok && !userID.IsZero() && file.UserID != userID {
		s.logger.Warn("Unauthorized file deletion attempt",
			zap.String("file_id", id.Hex()),
			zap.String("file_owner", file.UserID.Hex()),
			zap.String("requester", userID.Hex()),
		)
		return httperror.NewForForbiddenWithSingleField("message", "You do not have permission to delete this file")
	}

	// Delete the file using the use case
	err = s.deleteUseCase.Execute(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete encrypted file",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete encrypted file: %w", err)
	}

	s.logger.Info("Successfully deleted encrypted file",
		zap.String("id", id.Hex()),
		zap.String("userID", file.UserID.Hex()),
		zap.String("fileID", file.FileID),
	)

	return nil
}
