// internal/service/localfile/list.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// ListInput represents the input for listing files by collection
type ListInput struct {
	CollectionID string `json:"collection_id"`
}

// ListOutput represents the result of listing files by collection
type ListOutput struct {
	Files []*dom_file.File `json:"files"`
	Count int              `json:"count"`
}

// ListService defines the interface for listing local files by collection
type ListService interface {
	ListByCollection(ctx context.Context, input *ListInput) (*ListOutput, error)
}

// listService implements the ListService interface
type listService struct {
	logger                       *zap.Logger
	listFilesByCollectionUseCase file.ListFilesByCollectionUseCase
}

// NewListService creates a new service for listing local files by collection
func NewListService(
	logger *zap.Logger,
	listFilesByCollectionUseCase file.ListFilesByCollectionUseCase,
) ListService {
	logger = logger.Named("ListService")
	return &listService{
		logger:                       logger,
		listFilesByCollectionUseCase: listFilesByCollectionUseCase,
	}
}

// ListByCollection handles the listing of local files within a specific collection
func (s *listService) ListByCollection(ctx context.Context, input *ListInput) (*ListOutput, error) {
	//
	// STEP 1: Validate inputs
	//
	if input == nil {
		s.logger.Error("❌ Input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.CollectionID == "" {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	//
	// STEP 2: Convert collection ID string to ObjectID
	//
	collectionObjectID, err := primitive.ObjectIDFromHex(input.CollectionID)
	if err != nil {
		s.logger.Error("❌ Invalid collection ID format",
			zap.String("collectionID", input.CollectionID),
			zap.Error(err))
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	//
	// STEP 3: Execute the use case to list files by collection
	//
	s.logger.Debug("🔍 Listing files by collection",
		zap.String("collectionID", input.CollectionID))

	files, err := s.listFilesByCollectionUseCase.Execute(ctx, collectionObjectID)
	if err != nil {
		s.logger.Error("❌ Failed to list files by collection",
			zap.String("collectionID", input.CollectionID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to list files by collection", err)
	}

	//
	// STEP 4: Return structured output
	//
	s.logger.Info("✅ Successfully listed files by collection",
		zap.String("collectionID", input.CollectionID),
		zap.Int("fileCount", len(files)))

	return &ListOutput{
		Files: files,
		Count: len(files),
	}, nil
}
