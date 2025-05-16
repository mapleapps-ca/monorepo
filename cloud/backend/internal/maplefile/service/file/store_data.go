// cloud/backend/internal/maplefile/service/file/store_data.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type StoreFileDataRequestDTO struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}

type StoreFileDataResponseDTO struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	StoragePath string `json:"storage_path"`
}

type StoreFileDataService interface {
	Execute(sessCtx context.Context, req *StoreFileDataRequestDTO) (*StoreFileDataResponseDTO, error)
}

type storeFileDataServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewStoreFileDataService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) StoreFileDataService {
	return &storeFileDataServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *storeFileDataServiceImpl) Execute(sessCtx context.Context, req *StoreFileDataRequestDTO) (*StoreFileDataResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File data is required")
	}

	if req.ID == "" {
		svc.logger.Warn("Empty file ID")
		return nil, httperror.NewForBadRequestWithSingleField("id", "File ID is required")
	}

	if len(req.Data) == 0 {
		svc.logger.Warn("Empty file data")
		return nil, httperror.NewForBadRequestWithSingleField("data", "File data is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := sessCtx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Retrieve existing file
	//
	file, err := svc.fileRepo.Get(req.ID)
	if err != nil {
		svc.logger.Error("Failed to get file",
			zap.Any("error", err),
			zap.String("file_id", req.ID))
		return nil, err
	}

	if file == nil {
		svc.logger.Debug("File not found",
			zap.String("file_id", req.ID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	//
	// STEP 4: Check if user has rights to upload data for this file
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(
		file.CollectionID,
		userID.Hex(),
		dom_collection.CollectionPermissionReadWrite,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.String("collection_id", file.CollectionID),
			zap.String("user_id", userID.Hex()))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file data upload attempt",
			zap.String("user_id", userID.Hex()),
			zap.String("file_id", req.ID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to upload data for this file")
	}

	//
	// STEP 5: Store encrypted data
	//
	err = svc.fileRepo.StoreEncryptedData(req.ID, req.Data)
	if err != nil {
		svc.logger.Error("Failed to store file data",
			zap.Any("error", err),
			zap.String("file_id", req.ID))
		return nil, err
	}

	// Refresh file data to get updated storage path
	updatedFile, err := svc.fileRepo.Get(req.ID)
	if err != nil {
		svc.logger.Error("Failed to get updated file",
			zap.Any("error", err),
			zap.String("file_id", req.ID))
		return nil, err
	}

	svc.logger.Info("📄⬆️☁️ File data stored successfully",
		zap.String("file_id", req.ID),
		zap.String("storage_path", updatedFile.StoragePath),
		zap.Int64("size", updatedFile.EncryptedSize))

	return &StoreFileDataResponseDTO{
		Success:     true,
		Message:     "File data stored successfully",
		StoragePath: updatedFile.StoragePath,
	}, nil
}
