// cloud/mapleapps-backend/internal/maplefile/service/file/list_recent_files.go
package file

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// RecentFileResponseDTO represents a recent file in the response
type RecentFileResponseDTO struct {
	ID                            gocql.UUID            `json:"id"`
	CollectionID                  gocql.UUID            `json:"collection_id"`
	OwnerID                       gocql.UUID            `json:"owner_id"`
	EncryptedMetadata             string                `json:"encrypted_metadata"`
	EncryptedFileKey              keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion             string                `json:"encryption_version"`
	EncryptedHash                 string                `json:"encrypted_hash"`
	EncryptedFileSizeInBytes      int64                 `json:"encrypted_file_size_in_bytes"`
	EncryptedThumbnailSizeInBytes int64                 `json:"encrypted_thumbnail_size_in_bytes"`
	CreatedAt                     string                `json:"created_at"`
	ModifiedAt                    string                `json:"modified_at"`
	Version                       uint64                `json:"version"`
	State                         string                `json:"state"`
}

// ListRecentFilesResponseDTO represents the response for listing recent files
type ListRecentFilesResponseDTO struct {
	Files      []RecentFileResponseDTO `json:"files"`
	NextCursor *string                 `json:"next_cursor,omitempty"`
	HasMore    bool                    `json:"has_more"`
	TotalCount int                     `json:"total_count"`
}

type ListRecentFilesService interface {
	Execute(ctx context.Context, cursor *string, limit int64) (*ListRecentFilesResponseDTO, error)
}

type listRecentFilesServiceImpl struct {
	config                 *config.Configuration
	logger                 *zap.Logger
	listRecentFilesUseCase uc_filemetadata.ListRecentFilesUseCase
}

func NewListRecentFilesService(
	config *config.Configuration,
	logger *zap.Logger,
	listRecentFilesUseCase uc_filemetadata.ListRecentFilesUseCase,
) ListRecentFilesService {
	logger = logger.Named("ListRecentFilesService")
	return &listRecentFilesServiceImpl{
		config:                 config,
		logger:                 logger,
		listRecentFilesUseCase: listRecentFilesUseCase,
	}
}

func (svc *listRecentFilesServiceImpl) Execute(ctx context.Context, cursor *string, limit int64) (*ListRecentFilesResponseDTO, error) {
	//
	// STEP 1: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 2: Parse cursor if provided
	//
	var parsedCursor *dom_file.RecentFilesCursor
	if cursor != nil && *cursor != "" {
		// Decode base64 cursor
		cursorBytes, err := base64.StdEncoding.DecodeString(*cursor)
		if err != nil {
			svc.logger.Error("Failed to decode cursor",
				zap.String("cursor", *cursor),
				zap.Error(err))
			return nil, httperror.NewForBadRequestWithSingleField("cursor", "Invalid cursor format")
		}

		// Parse JSON cursor
		var cursorData dom_file.RecentFilesCursor
		if err := json.Unmarshal(cursorBytes, &cursorData); err != nil {
			svc.logger.Error("Failed to parse cursor",
				zap.String("cursor", *cursor),
				zap.Error(err))
			return nil, httperror.NewForBadRequestWithSingleField("cursor", "Invalid cursor format")
		}
		parsedCursor = &cursorData
	}

	//
	// STEP 3: Set default limit if not provided
	//
	if limit <= 0 {
		limit = 30 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	svc.logger.Debug("Processing recent files request",
		zap.Any("user_id", userID),
		zap.Int64("limit", limit),
		zap.Any("cursor", parsedCursor))

	//
	// STEP 4: Call use case to get recent files
	//
	response, err := svc.listRecentFilesUseCase.Execute(ctx, userID, parsedCursor, limit)
	if err != nil {
		svc.logger.Error("Failed to get recent files",
			zap.Any("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Convert domain response to service DTO
	//
	files := make([]RecentFileResponseDTO, len(response.Files))
	for i, file := range response.Files {
		// Deserialize encrypted file key
		var encryptedFileKey keys.EncryptedFileKey
		if err := json.Unmarshal([]byte(file.EncryptedFileKey), &encryptedFileKey); err != nil {
			svc.logger.Warn("Failed to deserialize encrypted file key for file",
				zap.String("file_id", file.ID.String()),
				zap.Error(err))
			// Continue with empty key rather than failing entirely
		}

		files[i] = RecentFileResponseDTO{
			ID:                            file.ID,
			CollectionID:                  file.CollectionID,
			OwnerID:                       file.OwnerID,
			EncryptedMetadata:             file.EncryptedMetadata,
			EncryptedFileKey:              encryptedFileKey,
			EncryptionVersion:             file.EncryptionVersion,
			EncryptedHash:                 file.EncryptedHash,
			EncryptedFileSizeInBytes:      file.EncryptedFileSizeInBytes,
			EncryptedThumbnailSizeInBytes: file.EncryptedThumbnailSizeInBytes,
			CreatedAt:                     file.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			ModifiedAt:                    file.ModifiedAt.Format("2006-01-02T15:04:05Z07:00"),
			Version:                       file.Version,
			State:                         file.State,
		}
	}

	//
	// STEP 6: Encode next cursor if present
	//
	var encodedNextCursor *string
	if response.NextCursor != nil {
		cursorBytes, err := json.Marshal(response.NextCursor)
		if err != nil {
			svc.logger.Error("Failed to marshal next cursor",
				zap.Any("cursor", response.NextCursor),
				zap.Error(err))
		} else {
			cursorStr := base64.StdEncoding.EncodeToString(cursorBytes)
			encodedNextCursor = &cursorStr
		}
	}

	//
	// STEP 7: Prepare response
	//
	serviceResponse := &ListRecentFilesResponseDTO{
		Files:      files,
		NextCursor: encodedNextCursor,
		HasMore:    response.HasMore,
		TotalCount: len(files),
	}

	svc.logger.Info("Successfully served recent files",
		zap.Any("user_id", userID),
		zap.Int("files_count", len(files)),
		zap.Bool("has_more", response.HasMore),
		zap.Any("next_cursor", encodedNextCursor))

	return serviceResponse, nil
}
