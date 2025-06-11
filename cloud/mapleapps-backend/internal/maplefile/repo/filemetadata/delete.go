// cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/delete.go
package filemetadata

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl *fileMetadataRepositoryImpl) SoftDelete(id gocql.UUID) error {
	file, err := impl.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get file for soft delete: %w", err)
	}

	if file == nil {
		return fmt.Errorf("file not found")
	}

	// Validate state transition
	if err := dom_file.IsValidStateTransition(file.State, dom_file.FileStateDeleted); err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	// Update file state
	file.State = dom_file.FileStateDeleted
	file.ModifiedAt = time.Now()
	file.Version++
	file.TombstoneVersion = file.Version
	file.TombstoneExpiry = time.Now().Add(30 * 24 * time.Hour) // 30 days

	return impl.Update(file)
}

func (impl *fileMetadataRepositoryImpl) SoftDeleteMany(ids []gocql.UUID) error {
	for _, id := range ids {
		if err := impl.SoftDelete(id); err != nil {
			impl.Logger.Warn("failed to soft delete file",
				zap.String("file_id", id.String()),
				zap.Error(err))
		}
	}
	return nil
}

func (impl *fileMetadataRepositoryImpl) HardDelete(id gocql.UUID) error {
	file, err := impl.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get file for hard delete: %w", err)
	}

	if file == nil {
		return fmt.Errorf("file not found")
	}

	batch := impl.Session.NewBatch(gocql.LoggedBatch)

	// 1. Delete from main table
	batch.Query(`DELETE FROM mapleapps.maplefile_files_by_id WHERE id = ?`, id)

	// 2. Delete from collection table
	batch.Query(`DELETE FROM mapleapps.maplefile_files_by_collection_id_with_desc_modified_at_and_asc_file_id
		WHERE collection_id = ? AND modified_at = ? AND file_id = ?`,
		file.CollectionID, file.ModifiedAt, id)

	// 3. Delete from owner table
	batch.Query(`DELETE FROM mapleapps.maplefile_files_by_owner_id_with_desc_modified_at_and_asc_file_id
		WHERE owner_id = ? AND modified_at = ? AND file_id = ?`,
		file.OwnerID, file.ModifiedAt, id)

	// 4. Delete from created_by table
	batch.Query(`DELETE FROM mapleapps.maplefile_files_by_created_by_user_id_with_desc_created_at_and_asc_file_id
		WHERE created_by_user_id = ? AND created_at = ? AND file_id = ?`,
		file.CreatedByUserID, file.CreatedAt, id)

	// 5. Delete from user sync table
	batch.Query(`DELETE FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
		WHERE user_id = ? AND modified_at = ? AND file_id = ?`,
		file.OwnerID, file.ModifiedAt, id)

	// Execute batch
	if err := impl.Session.ExecuteBatch(batch); err != nil {
		impl.Logger.Error("failed to hard delete file",
			zap.String("file_id", id.String()),
			zap.Error(err))
		return fmt.Errorf("failed to hard delete file: %w", err)
	}

	impl.Logger.Info("file hard deleted successfully",
		zap.String("file_id", id.String()))

	return nil
}

func (impl *fileMetadataRepositoryImpl) HardDeleteMany(ids []gocql.UUID) error {
	for _, id := range ids {
		if err := impl.HardDelete(id); err != nil {
			impl.Logger.Warn("failed to hard delete file",
				zap.String("file_id", id.String()),
				zap.Error(err))
		}
	}
	return nil
}
