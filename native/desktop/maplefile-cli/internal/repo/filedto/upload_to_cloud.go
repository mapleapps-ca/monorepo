// monorepo/native/desktop/maplefile-cli/internal/repo/filedto/upload_to_cloud.go
package filedto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
)

// UploadToCloud is a convenience method that combines all three steps of the upload process
func (r *fileDTORepository) UploadToCloud(ctx context.Context, file *filedto.FileDTO, fileData []byte, thumbnailData []byte) (*primitive.ObjectID, error) {
	r.logger.Debug("Starting complete file upload process",
		zap.String("fileID", file.ID.Hex()),
		zap.Int("fileDataSize", len(fileData)),
		zap.Int("thumbnailDataSize", len(thumbnailData)))

	if file == nil {
		return nil, errors.NewAppError("file DTO is required", nil)
	}

	if len(fileData) == 0 {
		return nil, errors.NewAppError("file data is required", nil)
	}

	// Check if this is a new file (no ID) or an update to existing file
	if file.ID.IsZero() {
		// New file - perform three-step upload
		return r.performThreeStepUpload(ctx, file, fileData, thumbnailData)
	} else {
		// Existing file - update via presigned URL
		return r.updateExistingFile(ctx, file, fileData, thumbnailData)
	}
}

// performThreeStepUpload handles the complete three-step upload process for new files
func (r *fileDTORepository) performThreeStepUpload(ctx context.Context, file *filedto.FileDTO, fileData []byte, thumbnailData []byte) (*primitive.ObjectID, error) {
	r.logger.Debug("Performing three-step upload for new file")

	// Step 1: Create pending file
	createRequest := &filedto.CreatePendingFileRequest{
		CollectionID:                 file.CollectionID,
		EncryptedMetadata:            file.EncryptedMetadata,
		EncryptedFileKey:             convertToEncryptedFileKey(file.EncryptedFileKey),
		EncryptionVersion:            file.EncryptionVersion,
		EncryptedHash:                file.EncryptedHash,
		ExpectedFileSizeInBytes:      int64(len(fileData)),
		ExpectedThumbnailSizeInBytes: int64(len(thumbnailData)),
	}

	createResponse, err := r.CreatePendingFileInCloud(ctx, createRequest)
	if err != nil {
		return nil, errors.NewAppError("failed to create pending file", err)
	}

	fileID := createResponse.File.ID
	r.logger.Debug("Created pending file", zap.String("fileID", fileID.Hex()))

	// Step 2: Upload file content
	err = r.UploadFileToCloud(ctx, createResponse.PresignedUploadURL, fileData)
	if err != nil {
		return &fileID, errors.NewAppError("failed to upload file content", err)
	}

	// Step 2b: Upload thumbnail if provided
	thumbnailUploaded := false
	if len(thumbnailData) > 0 && createResponse.PresignedThumbnailURL != "" {
		err = r.UploadThumbnailToCloud(ctx, createResponse.PresignedThumbnailURL, thumbnailData)
		if err != nil {
			r.logger.Warn("Failed to upload thumbnail, continuing without it", zap.Error(err))
		} else {
			thumbnailUploaded = true
		}
	}

	// Step 3: Complete the upload
	completeRequest := &filedto.CompleteFileUploadRequest{
		ActualFileSizeInBytes:      int64(len(fileData)),
		ActualThumbnailSizeInBytes: int64(len(thumbnailData)),
		UploadConfirmed:            true,
		ThumbnailUploadConfirmed:   thumbnailUploaded,
	}

	completeResponse, err := r.CompleteFileUploadInCloud(ctx, fileID, completeRequest)
	if err != nil {
		return &fileID, errors.NewAppError("failed to complete file upload", err)
	}

	r.logger.Info("Successfully completed three-step upload",
		zap.String("fileID", fileID.Hex()),
		zap.Bool("uploadVerified", completeResponse.UploadVerified))

	return &fileID, nil
}

// updateExistingFile handles updating an existing file by getting new presigned URLs
func (r *fileDTORepository) updateExistingFile(ctx context.Context, file *filedto.FileDTO, fileData []byte, thumbnailData []byte) (*primitive.ObjectID, error) {
	r.logger.Debug("Updating existing file", zap.String("fileID", file.ID.Hex()))

	// Get new presigned URLs for the existing file
	urlRequest := &filedto.GetPresignedUploadURLRequest{
		URLDuration: 0, // Use default duration
	}

	urlResponse, err := r.GetPresignedUploadURLFromCloud(ctx, file.ID, urlRequest)
	if err != nil {
		return &file.ID, errors.NewAppError("failed to get presigned upload URLs", err)
	}

	// Upload file content
	err = r.UploadFileToCloud(ctx, urlResponse.PresignedUploadURL, fileData)
	if err != nil {
		return &file.ID, errors.NewAppError("failed to upload updated file content", err)
	}

	// Upload thumbnail if provided
	thumbnailUploaded := false
	if len(thumbnailData) > 0 && urlResponse.PresignedThumbnailURL != "" {
		err = r.UploadThumbnailToCloud(ctx, urlResponse.PresignedThumbnailURL, thumbnailData)
		if err != nil {
			r.logger.Warn("Failed to upload updated thumbnail, continuing without it", zap.Error(err))
		} else {
			thumbnailUploaded = true
		}
	}

	// Complete the upload to verify the update
	completeRequest := &filedto.CompleteFileUploadRequest{
		ActualFileSizeInBytes:      int64(len(fileData)),
		ActualThumbnailSizeInBytes: int64(len(thumbnailData)),
		UploadConfirmed:            true,
		ThumbnailUploadConfirmed:   thumbnailUploaded,
	}

	_, err = r.CompleteFileUploadInCloud(ctx, file.ID, completeRequest)
	if err != nil {
		return &file.ID, errors.NewAppError("failed to complete file update", err)
	}

	r.logger.Info("Successfully updated existing file", zap.String("fileID", file.ID.Hex()))

	return &file.ID, nil
}

// convertToEncryptedFileKey converts from the domain keys format to the DTO format
func convertToEncryptedFileKey(domainKey interface{}) filedto.EncryptedFileKey {
	// This is a placeholder conversion. In practice, you would need to properly
	// convert from your domain's EncryptedFileKey format to the DTO format.
	// The exact implementation depends on your domain key structure.

	// For now, assuming the domain key has Ciphertext and Nonce fields
	if key, ok := domainKey.(map[string]interface{}); ok {
		if ciphertext, ok := key["ciphertext"].([]byte); ok {
			if nonce, ok := key["nonce"].([]byte); ok {
				return filedto.EncryptedFileKey{
					Ciphertext: ciphertext,
					Nonce:      nonce,
				}
			}
		}
	}

	// Return empty key if conversion fails
	return filedto.EncryptedFileKey{}
}
