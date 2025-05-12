// cloud/backend/internal/vault/service/encryptedfile/getdownloadurl.go
package encryptedfile

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/usecase/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// GetEncryptedFileDownloadURLService defines operations for generating a download URL
type GetEncryptedFileDownloadURLService interface {
	Execute(ctx context.Context, id primitive.ObjectID, expiryDuration time.Duration) (string, error)
}

type getEncryptedFileDownloadURLServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	getByIDUseCase        encryptedfile.GetEncryptedFileByIDUseCase
	getDownloadURLUseCase encryptedfile.GetEncryptedFileDownloadURLUseCase
}

// NewGetEncryptedFileDownloadURLService creates a new instance of the service
func NewGetEncryptedFileDownloadURLService(
	config *config.Configuration,
	logger *zap.Logger,
	getByIDUseCase encryptedfile.GetEncryptedFileByIDUseCase,
	getDownloadURLUseCase encryptedfile.GetEncryptedFileDownloadURLUseCase,
) GetEncryptedFileDownloadURLService {
	return &getEncryptedFileDownloadURLServiceImpl{
		config:                config,
		logger:                logger.With(zap.String("component", "get-encrypted-file-download-url-service")),
		getByIDUseCase:        getByIDUseCase,
		getDownloadURLUseCase: getDownloadURLUseCase,
	}
}

// Execute generates a presigned URL for direct download after verifying ownership
func (s *getEncryptedFileDownloadURLServiceImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
	expiryDuration time.Duration,
) (string, error) {
	// First get the file to verify ownership
	file, err := s.getByIDUseCase.Execute(ctx, id)
	if err != nil {
		return "", err
	}

	if file == nil {
		return "", httperror.NewForBadRequestWithSingleField("id", "File not found")
	}

	// Verify that the authenticated user has access to this file
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if ok && !userID.IsZero() && file.UserID != userID {
		s.logger.Warn("Unauthorized file URL generation attempt",
			zap.String("file_id", id.Hex()),
			zap.String("file_owner", file.UserID.Hex()),
			zap.String("requester", userID.Hex()),
		)
		return "", httperror.NewForForbiddenWithSingleField("message", "You do not have permission to get a download URL for this file")
	}

	// Get the download URL using the use case
	return s.getDownloadURLUseCase.Execute(ctx, id, expiryDuration)
}
