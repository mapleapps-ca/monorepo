// cloud/backend/internal/maplefile/service/collection/archive.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ArchiveCollectionRequestDTO struct {
	ID gocql.UUID `json:"id"`
}

type ArchiveCollectionResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ArchiveCollectionService interface {
	Execute(ctx context.Context, req *ArchiveCollectionRequestDTO) (*ArchiveCollectionResponseDTO, error)
}

type archiveCollectionServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewArchiveCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) ArchiveCollectionService {
	logger = logger.Named("ArchiveCollectionService")
	return &archiveCollectionServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *archiveCollectionServiceImpl) Execute(ctx context.Context, req *ArchiveCollectionRequestDTO) (*ArchiveCollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Collection ID is required")
	}

	if req.ID.String() == "" {
		svc.logger.Warn("Empty collection ID")
		return nil, httperror.NewForBadRequestWithSingleField("id", "Collection ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Retrieve existing collection (including non-active states for archiving)
	//
	collection, err := svc.repo.GetWithAnyState(ctx, req.ID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.Any("collection_id", req.ID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.Any("collection_id", req.ID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 4: Check if user has rights to archive this collection
	//
	if collection.OwnerID != userID {
		svc.logger.Warn("Unauthorized collection archive attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.ID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "Only the collection owner can archive a collection")
	}

	//
	// STEP 5: Validate state transition
	//
	err = dom_collection.IsValidStateTransition(collection.State, dom_collection.CollectionStateArchived)
	if err != nil {
		svc.logger.Warn("Invalid state transition for collection archive",
			zap.Any("collection_id", req.ID),
			zap.String("current_state", collection.State),
			zap.String("target_state", dom_collection.CollectionStateArchived),
			zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("state", err.Error())
	}

	//
	// STEP 6: Archive the collection
	//
	collection.State = dom_collection.CollectionStateArchived
	collection.Version++ // Update mutation means we increment version.
	collection.ModifiedAt = time.Now()
	collection.ModifiedByUserID = userID
	err = svc.repo.Update(ctx, collection)
	if err != nil {
		svc.logger.Error("Failed to archive collection",
			zap.Any("error", err),
			zap.Any("collection_id", req.ID))
		return nil, err
	}

	svc.logger.Info("Collection archived successfully",
		zap.Any("collection_id", req.ID),
		zap.Any("user_id", userID))

	return &ArchiveCollectionResponseDTO{
		Success: true,
		Message: "Collection archived successfully",
	}, nil
}
