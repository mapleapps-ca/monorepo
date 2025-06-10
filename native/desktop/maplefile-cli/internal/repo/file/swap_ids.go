// native/desktop/maplefile-cli/internal/repo/file/swap_ids.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

func (r *fileRepository) SwapIDs(ctx context.Context, oldID gocql.UUID, newID gocql.UUID) error {
	r.logger.Debug("Swapping file IDs",
		zap.String("oldID", oldID.Hex()),
		zap.String("newID", newID.Hex()))

	// Step 1: Validate inputs
	if oldID.IsZero() {
		r.logger.Error("Old ID cannot be zero")
		return errors.NewAppError("old ID cannot be zero", nil)
	}
	if newID.IsZero() {
		r.logger.Error("New ID cannot be zero")
		return errors.NewAppError("new ID cannot be zero", nil)
	}
	if oldID == newID {
		r.logger.Error("Old ID and new ID cannot be the same",
			zap.String("ID", oldID.Hex()))
		return errors.NewAppError("old ID and new ID cannot be the same", nil)
	}

	// Step 2: Confirm old record exists with old ID in the database
	oldExists, err := r.CheckIfExistsByID(ctx, oldID)
	if err != nil {
		r.logger.Error("Failed to check if old record exists",
			zap.String("oldID", oldID.Hex()),
			zap.Error(err))
		return errors.NewAppError("failed to check if old record exists", err)
	}
	if !oldExists {
		r.logger.Error("Old record does not exist",
			zap.String("oldID", oldID.Hex()))
		return errors.NewAppError("old record does not exist", nil)
	}

	// Step 3: Confirm no record exists with new ID in the database
	newExists, err := r.CheckIfExistsByID(ctx, newID)
	if err != nil {
		r.logger.Error("Failed to check if new record exists",
			zap.String("newID", newID.Hex()),
			zap.Error(err))
		return errors.NewAppError("failed to check if new record exists", err)
	}
	if newExists {
		r.logger.Error("New record already exists - cannot swap to existing ID",
			zap.String("newID", newID.Hex()))
		return errors.NewAppError("new record already exists", nil)
	}

	// Step 4: Get the old record to copy
	oldFile, err := r.Get(ctx, oldID)
	if err != nil {
		r.logger.Error("Failed to get old record",
			zap.String("oldID", oldID.Hex()),
			zap.Error(err))
		return errors.NewAppError("failed to get old record", err)
	}
	if oldFile == nil {
		r.logger.Error("Old record not found despite existence check passing",
			zap.String("oldID", oldID.Hex()))
		return errors.NewAppError("old record not found", nil)
	}

	// Step 5: Copy old record into new record struct, updating the ID with the new ID
	newFile := *oldFile // Create a shallow copy of the file
	newFile.ID = newID  // Update the ID to the new ID

	// Step 6: Use transaction to ensure atomicity of the swap operation
	if err := r.OpenTransaction(); err != nil {
		r.logger.Error("Failed to open transaction for ID swap", zap.Error(err))
		return errors.NewAppError("failed to open transaction", err)
	}

	// Step 7: Delete old record using the old ID value from the database
	if err := r.Delete(ctx, oldID); err != nil {
		r.DiscardTransaction()
		r.logger.Error("Failed to delete old record during ID swap",
			zap.String("oldID", oldID.Hex()),
			zap.Error(err))
		return errors.NewAppError("failed to delete old record", err)
	}

	// Step 8: Create new record using the new ID as its `id` value in the database
	if err := r.Create(ctx, &newFile); err != nil {
		r.DiscardTransaction()
		r.logger.Error("Failed to create new record during ID swap",
			zap.String("newID", newID.Hex()),
			zap.Error(err))
		return errors.NewAppError("failed to create new record", err)
	}

	// Step 9: Commit the transaction to make the swap permanent
	if err := r.CommitTransaction(); err != nil {
		r.logger.Error("Failed to commit transaction for ID swap", zap.Error(err))
		return errors.NewAppError("failed to commit transaction", err)
	}

	r.logger.Info("Successfully swapped file IDs",
		zap.String("oldID", oldID.Hex()),
		zap.String("newID", newID.Hex()),
		zap.String("fileName", newFile.Name))

	return nil
}
