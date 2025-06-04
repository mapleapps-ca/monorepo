// internal/service/authdto/logout.go
package authdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	uc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/authdto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	uc_localfile "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	uc_syncstate "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/syncstate"
)

// LogoutService provides high-level functionality for user logout
type LogoutService interface {
	Logout(ctx context.Context) error
}

// logoutService implements the LogoutService interface
type logoutService struct {
	logger                       *zap.Logger
	configService                config.ConfigService
	transactionManager           transaction.Manager
	logoutUseCase                uc_authdto.LogoutUseCase
	listCollectionsUseCase       uc_collection.ListCollectionsUseCase
	deleteCollectionUseCase      uc_collection.DeleteCollectionUseCase
	listFilesByCollectionUseCase uc_file.ListFilesByCollectionUseCase
	deleteFileUseCase            uc_file.DeleteFileUseCase
	deleteLocalFileUseCase       uc_localfile.DeleteFileUseCase
	resetSyncStateUseCase        uc_syncstate.ResetSyncStateUseCase
}

// NewLogoutService creates a new logout service
func NewLogoutService(
	logger *zap.Logger,
	configService config.ConfigService,
	transactionManager transaction.Manager,
	logoutUseCase uc_authdto.LogoutUseCase,
	listCollectionsUseCase uc_collection.ListCollectionsUseCase,
	deleteCollectionUseCase uc_collection.DeleteCollectionUseCase,
	listFilesByCollectionUseCase uc_file.ListFilesByCollectionUseCase,
	deleteFileUseCase uc_file.DeleteFileUseCase,
	deleteLocalFileUseCase uc_localfile.DeleteFileUseCase,
	resetSyncStateUseCase uc_syncstate.ResetSyncStateUseCase,
) LogoutService {
	logger = logger.Named("LogoutService")
	return &logoutService{
		logger:                       logger,
		configService:                configService,
		transactionManager:           transactionManager,
		logoutUseCase:                logoutUseCase,
		listCollectionsUseCase:       listCollectionsUseCase,
		deleteCollectionUseCase:      deleteCollectionUseCase,
		listFilesByCollectionUseCase: listFilesByCollectionUseCase,
		deleteFileUseCase:            deleteFileUseCase,
		deleteLocalFileUseCase:       deleteLocalFileUseCase,
		resetSyncStateUseCase:        resetSyncStateUseCase,
	}
}

// Logout handles the entire flow of user logout including complete local data cleanup
func (s *logoutService) Logout(ctx context.Context) error {
	// Check if user is currently logged in
	credentials, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return errors.NewAppError("failed to get current user credentials", err)
	}

	if credentials == nil || credentials.Email == "" {
		return errors.NewAppError("no user is currently logged in", nil)
	}

	currentUserEmail := credentials.Email
	s.logger.Info("🚪 Processing logout request with complete data cleanup", zap.String("email", currentUserEmail))

	// Begin transaction for atomic cleanup
	if err := s.transactionManager.Begin(); err != nil {
		s.logger.Error("❌ Failed to begin transaction for logout cleanup", zap.Error(err))
		return errors.NewAppError("failed to begin transaction for logout cleanup", err)
	}

	// Ensure transaction cleanup on any exit
	defer func() {
		if s.transactionManager.IsInTransaction() {
			s.transactionManager.Rollback()
		}
	}()

	//
	// STEP 1: Delete all local file data and metadata
	//
	s.logger.Info("🗑️  Step 1: Cleaning up local files")
	if err := s.deleteAllLocalFiles(ctx); err != nil {
		s.logger.Error("❌ Failed to delete local files during logout", zap.Error(err))
		return errors.NewAppError("failed to delete local files during logout", err)
	}

	//
	// STEP 2: Delete all local collections
	//
	s.logger.Info("🗑️  Step 2: Cleaning up local collections")
	if err := s.deleteAllLocalCollections(ctx); err != nil {
		s.logger.Error("❌ Failed to delete local collections during logout", zap.Error(err))
		return errors.NewAppError("failed to delete local collections during logout", err)
	}

	//
	// STEP 3: Reset sync state
	//
	s.logger.Info("🔄 Step 3: Resetting sync state")
	if err := s.resetSyncStateUseCase.Execute(ctx); err != nil {
		s.logger.Error("❌ Failed to reset sync state during logout", zap.Error(err))
		return errors.NewAppError("failed to reset sync state during logout", err)
	}

	//
	// STEP 4: Clear user credentials (using the simple use case)
	//
	s.logger.Info("🔑 Step 4: Clearing user credentials")
	if err := s.logoutUseCase.Logout(ctx); err != nil {
		s.logger.Error("❌ Failed to clear user credentials during logout", zap.Error(err))
		return errors.NewAppError("failed to clear user credentials during logout", err)
	}

	// Commit transaction
	if err := s.transactionManager.Commit(); err != nil {
		s.logger.Error("❌ Failed to commit logout transaction", zap.Error(err))
		return errors.NewAppError("failed to commit logout transaction", err)
	}

	s.logger.Info("✅ Logout completed successfully with complete data cleanup", zap.String("email", currentUserEmail))

	return nil
}

// deleteAllLocalFiles deletes all local file data and metadata
func (s *logoutService) deleteAllLocalFiles(ctx context.Context) error {
	// Get all collections to find their files
	collections, err := s.listCollectionsUseCase.ListRoots(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to list collections for file cleanup", zap.Error(err))
		return err
	}

	fileCount := 0
	deletedFileDataCount := 0
	deletedMetadataCount := 0

	for _, coll := range collections {
		s.logger.Debug("🔍 Processing files in collection",
			zap.String("collectionID", coll.ID.Hex()),
			zap.String("collectionName", coll.Name))

		// Get all files in this collection
		files, err := s.listFilesByCollectionUseCase.Execute(ctx, coll.ID)
		if err != nil {
			s.logger.Warn("⚠️  Failed to list files in collection, continuing",
				zap.String("collectionID", coll.ID.Hex()),
				zap.Error(err))
			continue
		}

		fileCount += len(files)

		for _, file := range files {
			s.logger.Debug("🗑️  Deleting file",
				zap.String("fileID", file.ID.Hex()),
				zap.String("fileName", file.Name))

			// Delete file data from disk (if it exists)
			if file.FilePath != "" {
				if err := s.deleteLocalFileUseCase.Execute(ctx, file.FilePath); err != nil {
					s.logger.Warn("⚠️  Failed to delete file data, continuing",
						zap.String("filePath", file.FilePath),
						zap.Error(err))
				} else {
					deletedFileDataCount++
				}
			}

			// Delete encrypted file data from disk (if it exists)
			if file.EncryptedFilePath != "" {
				if err := s.deleteLocalFileUseCase.Execute(ctx, file.EncryptedFilePath); err != nil {
					s.logger.Warn("⚠️  Failed to delete encrypted file data, continuing",
						zap.String("encryptedFilePath", file.EncryptedFilePath),
						zap.Error(err))
				}
			}

			// Delete thumbnail data from disk (if it exists)
			if file.ThumbnailPath != "" {
				if err := s.deleteLocalFileUseCase.Execute(ctx, file.ThumbnailPath); err != nil {
					s.logger.Warn("⚠️  Failed to delete thumbnail data, continuing",
						zap.String("thumbnailPath", file.ThumbnailPath),
						zap.Error(err))
				}
			}

			// Delete encrypted thumbnail data from disk (if it exists)
			if file.EncryptedThumbnailPath != "" {
				if err := s.deleteLocalFileUseCase.Execute(ctx, file.EncryptedThumbnailPath); err != nil {
					s.logger.Warn("⚠️  Failed to delete encrypted thumbnail data, continuing",
						zap.String("encryptedThumbnailPath", file.EncryptedThumbnailPath),
						zap.Error(err))
				}
			}

			// Delete file metadata from database
			if err := s.deleteFileUseCase.Execute(ctx, file.ID); err != nil {
				s.logger.Warn("⚠️  Failed to delete file metadata, continuing",
					zap.String("fileID", file.ID.Hex()),
					zap.Error(err))
			} else {
				deletedMetadataCount++
			}
		}
	}

	s.logger.Info("✅ Completed file cleanup",
		zap.Int("totalFiles", fileCount),
		zap.Int("deletedFileData", deletedFileDataCount),
		zap.Int("deletedMetadata", deletedMetadataCount))

	return nil
}

// deleteAllLocalCollections deletes all local collection metadata
func (s *logoutService) deleteAllLocalCollections(ctx context.Context) error {
	// Get all collections including all states
	filter := collection.GetAllStatesFilter()
	collections, err := s.listCollectionsUseCase.Execute(ctx, filter)
	if err != nil {
		s.logger.Error("❌ Failed to list all collections for cleanup", zap.Error(err))
		return err
	}

	collectionCount := len(collections)
	deletedCount := 0

	for _, coll := range collections {
		s.logger.Debug("🗑️  Deleting collection",
			zap.String("collectionID", coll.ID.Hex()),
			zap.String("collectionName", coll.Name))

		if err := s.deleteCollectionUseCase.Execute(ctx, coll.ID); err != nil {
			s.logger.Warn("⚠️  Failed to delete collection, continuing",
				zap.String("collectionID", coll.ID.Hex()),
				zap.Error(err))
		} else {
			deletedCount++
		}
	}

	s.logger.Info("✅ Completed collection cleanup",
		zap.Int("totalCollections", collectionCount),
		zap.Int("deletedCollections", deletedCount))

	return nil
}
