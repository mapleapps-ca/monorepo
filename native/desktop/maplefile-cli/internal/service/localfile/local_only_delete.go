// internal/service/localfile/local_only_delete.go
package localfile

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	dom_tx "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// LocalOnlyDeleteInput represents the input for deleting local file by id
type LocalOnlyDeleteInput struct {
	ID gocql.UUID `json:"id"`
}

// LocalOnlyDeleteService defines the interface for LocalOnlyDeleteService for deleting local files by id
type LocalOnlyDeleteService interface {
	Execute(ctx context.Context, input *LocalOnlyDeleteInput) error
}

// localOnlyDeleteService implements the LocalOnlyDeleteService interface
type localOnlyDeleteService struct {
	logger                    *zap.Logger
	transactionManager        dom_tx.Manager
	getFileUseCase            file.GetFileUseCase
	deleteFileMetadataUseCase file.DeleteFileUseCase
	deleteFileDataUseCase     localfile.DeleteFileUseCase
}

// NewLocalOnlyDeleteService creates a new service for deleting a local file by id
func NewLocalOnlyDeleteService(
	logger *zap.Logger,
	transactionManager dom_tx.Manager,
	getFileUseCase file.GetFileUseCase,
	deleteFileMetadataUseCase file.DeleteFileUseCase,
	deleteFileDataUseCase localfile.DeleteFileUseCase,
) LocalOnlyDeleteService {
	logger = logger.Named("LocalOnlyDeleteService")
	return &localOnlyDeleteService{
		logger:                    logger,
		transactionManager:        transactionManager,
		getFileUseCase:            getFileUseCase,
		deleteFileMetadataUseCase: deleteFileMetadataUseCase,
		deleteFileDataUseCase:     deleteFileDataUseCase,
	}
}

// LocalOnlyDeleteByCollection handles the deletion of a local file by id
func (s *localOnlyDeleteService) Execute(ctx context.Context, input *LocalOnlyDeleteInput) error {
	//
	// STEP 1: Validate inputs
	//
	if input == nil {
		s.logger.Error("❌ input is required")
		return errors.NewAppError("input is required", nil)
	}
	if input.ID.String() == "" {
		s.logger.Error("❌ ID is required")
		return errors.NewAppError("ID is required", nil)
	}

	//
	// STEP 2: Get related records
	//

	fileMetadata, err := s.getFileUseCase.Execute(ctx, input.ID)
	if err != nil {
		s.logger.Error("❌ failed to get file metadata by ID",
			zap.Any("id", input.ID),
			zap.Error(err))
		return errors.NewAppError("failed to get file metadata by ID", err)
	}
	if fileMetadata == nil {
		err := fmt.Errorf("file metadata does not exist for ID: %v\n", input.ID)
		s.logger.Error("❌ failed to get file metadata by ID",
			zap.Any("id", input.ID),
			zap.Error(err))
		return errors.NewAppError("failed local only delete of file", err)
	}
	if fileMetadata.SyncStatus != dom_file.SyncStatusLocalOnly {
		err := fmt.Errorf("file can only be in %v sync state for deletions", "local-only")
		return errors.NewAppError("failed local only delete of file", err)
	}

	//
	// STEP 3: Begin transaction
	//

	if err := s.transactionManager.Begin(); err != nil {
		s.logger.Error("❌ failed to begin transaction", zap.Error(err))
		return errors.NewAppError("failed to begin transaction", err)
	}

	//
	// STEP 4: Execute the deletion of metadata and file.
	//

	if err := s.deleteFileMetadataUseCase.Execute(ctx, fileMetadata.ID); err != nil {
		s.logger.Error("❌ failed deleting metadata",
			zap.Any("collectionID", input.ID),
			zap.Error(err))
		s.transactionManager.Rollback()
		return errors.NewAppError("failed deleting file metadata", err)
	}
	if err := s.deleteFileDataUseCase.Execute(ctx, fileMetadata.FilePath); err != nil {
		s.logger.Error("⚠️ failed deleting file data",
			zap.Any("collectionID", input.ID),
			zap.Error(err))
		// We might get errors here in case the user deleted the file in the directory, so just skip error handling and proceed.
	}

	//
	// STEP 5: Commit transaction and return method output.
	//
	if err := s.transactionManager.Commit(); err != nil {
		s.logger.Error("❌ failed to commit transaction", zap.Error(err))
		s.transactionManager.Rollback()
		return errors.NewAppError("failed to commit transaction", err)
	}
	return nil
}
