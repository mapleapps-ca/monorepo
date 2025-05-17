// internal/service/remotefile/list.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// ListOutput represents the result of listing remote files
type ListOutput struct {
	Files []*remotefile.RemoteFile `json:"files"`
	Count int                      `json:"count"`
}

// ListService defines the interface for listing remote files
type ListService interface {
	ListAll(ctx context.Context) (*ListOutput, error)
	ListByCollection(ctx context.Context, collectionID string) (*ListOutput, error)
}

// listService implements the ListService interface
type listService struct {
	logger           *zap.Logger
	listFilesUseCase uc.ListRemoteFilesUseCase
}

// NewListService creates a new service for listing remote files
func NewListService(
	logger *zap.Logger,
	listFilesUseCase uc.ListRemoteFilesUseCase,
) ListService {
	return &listService{
		logger:           logger,
		listFilesUseCase: listFilesUseCase,
	}
}

// ListAll lists all remote files
func (s *listService) ListAll(ctx context.Context) (*ListOutput, error) {
	// Get all files
	files, err := s.listFilesUseCase.Execute(ctx, remotefile.RemoteFileFilter{})
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		Files: files,
		Count: len(files),
	}, nil
}

// ListByCollection lists remote files in a collection
func (s *listService) ListByCollection(ctx context.Context, collectionID string) (*ListOutput, error) {
	// Validate inputs
	if collectionID == "" {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Convert collection ID to ObjectID
	colID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Get files in collection
	files, err := s.listFilesUseCase.ByCollection(ctx, colID)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		Files: files,
		Count: len(files),
	}, nil
}
