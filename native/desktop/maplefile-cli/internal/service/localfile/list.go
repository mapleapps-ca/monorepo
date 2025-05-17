// internal/service/localfile/list.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// ListOutput represents the result of listing local files
type ListOutput struct {
	Files []*localfile.LocalFile `json:"files"`
	Count int                    `json:"count"`
}

// ListService defines the interface for listing local files
type ListService interface {
	ListAll(ctx context.Context) (*ListOutput, error)
	ListByCollection(ctx context.Context, collectionID string) (*ListOutput, error)
	ListModifiedLocally(ctx context.Context) (*ListOutput, error)
	Search(ctx context.Context, nameContains string, mimeType string) (*ListOutput, error)
}

// listService implements the ListService interface
type listService struct {
	logger           *zap.Logger
	listFilesUseCase uc.ListLocalFilesUseCase
}

// NewListService creates a new service for listing local files
func NewListService(
	logger *zap.Logger,
	listFilesUseCase uc.ListLocalFilesUseCase,
) ListService {
	return &listService{
		logger:           logger,
		listFilesUseCase: listFilesUseCase,
	}
}

// ListAll lists all local files
func (s *listService) ListAll(ctx context.Context) (*ListOutput, error) {
	// Get all files
	files, err := s.listFilesUseCase.Execute(ctx, localfile.LocalFileFilter{})
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		Files: files,
		Count: len(files),
	}, nil
}

// ListByCollection lists local files in a collection
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

// ListModifiedLocally lists locally modified files
func (s *listService) ListModifiedLocally(ctx context.Context) (*ListOutput, error) {
	// Get modified files
	files, err := s.listFilesUseCase.ModifiedLocally(ctx)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		Files: files,
		Count: len(files),
	}, nil
}

// Search searches for files by name and mime type
func (s *listService) Search(ctx context.Context, nameContains string, mimeType string) (*ListOutput, error) {
	// Search for files
	files, err := s.listFilesUseCase.Search(ctx, nameContains, mimeType)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		Files: files,
		Count: len(files),
	}, nil
}
