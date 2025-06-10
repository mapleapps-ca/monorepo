// internal/usecase/file/swap_ids.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// SwapIDsUseCase defines the interface for swapping IDs of a local file
type SwapIDsUseCase interface {
	Execute(ctx context.Context, oldID gocql.UUID, newID gocql.UUID) error
}

// swapIDsUseCase implements the SwapIDsUseCase interface
type swapIDsUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewSwapIDsUseCase creates a new use case for swapping IDs of local files
func NewSwapIDsUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) SwapIDsUseCase {
	logger = logger.Named("SwapIDsUseCase")
	return &swapIDsUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute swaps IDs of a local file
func (uc *swapIDsUseCase) Execute(ctx context.Context, oldID gocql.UUID, newID gocql.UUID) error {
	// Validate inputs
	if oldID.String() == "" {
		return errors.NewAppError("file old ID is required", nil)
	}
	if newID.String() == "" {
		return errors.NewAppError("file new ID is required", nil)
	}

	// Swap IDs of the file
	err := uc.repository.SwapIDs(ctx, oldID, newID)
	if err != nil {
		return errors.NewAppError("failed to swap IDs of local file", err)
	}

	return nil
}
