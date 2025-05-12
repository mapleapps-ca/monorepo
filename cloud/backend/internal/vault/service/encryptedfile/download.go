// cloud/backend/internal/vault/service/encryptedfile/download.go
package encryptedfile

import (
	"context"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/usecase/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// DownloadEncryptedFileService defines operations for downloading encrypted file content
type DownloadEncryptedFileService interface {
	Execute(ctx context.Context, id primitive.ObjectID) (io.ReadCloser, error)
}

type downloadEncryptedFileServiceImpl struct {
	config          *config.Configuration
	logger          *zap.Logger
	getByIDUseCase  encryptedfile.GetEncryptedFileByIDUseCase
	downloadUseCase encryptedfile.DownloadEncryptedFileUseCase
}

// NewDownloadEncryptedFileService creates a new instance of the service
func NewDownloadEncryptedFileService(
	config *config.Configuration,
	logger *zap.Logger,
	getByIDUseCase encryptedfile.GetEncryptedFileByIDUseCase,
	downloadUseCase encryptedfile.DownloadEncryptedFileUseCase,
) DownloadEncryptedFileService {
	return &downloadEncryptedFileServiceImpl{
		config:          config,
		logger:          logger.With(zap.String("component", "download-encrypted-file-service")),
		getByIDUseCase:  getByIDUseCase,
		downloadUseCase: downloadUseCase,
	}
}

// Execute downloads the encrypted content of a file after verifying ownership
func (s *downloadEncryptedFileServiceImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (io.ReadCloser, error) {
	// First get the file to verify ownership
	file, err := s.getByIDUseCase.Execute(ctx, id)
	if err != nil {
		return nil, err
	}

	if file == nil {
		return nil, httperror.NewForBadRequestWithSingleField("id", "File not found")
	}

	// Verify that the authenticated user has access to this file
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if ok && !userID.IsZero() && file.UserID != userID {
		s.logger.Warn("Unauthorized file download attempt",
			zap.String("file_id", id.Hex()),
			zap.String("file_owner", file.UserID.Hex()),
			zap.String("requester", userID.Hex()),
		)
		return nil, httperror.NewForForbiddenWithSingleField("message", "You do not have permission to download this file")
	}

	// Download the file using the use case
	return s.downloadUseCase.Execute(ctx, id)
}
