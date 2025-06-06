// cloud/backend/internal/maplefile/service/collection/find_by_parent.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type FindByParentRequestDTO struct {
	ParentID gocql.UUID `json:"parent_id"`
}

type FindCollectionsByParentService interface {
	Execute(ctx context.Context, req *FindByParentRequestDTO) (*CollectionsResponseDTO, error)
}

type findCollectionsByParentServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewFindCollectionsByParentService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) FindCollectionsByParentService {
	logger = logger.Named("FindCollectionsByParentService")
	return &findCollectionsByParentServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *findCollectionsByParentServiceImpl) Execute(ctx context.Context, req *FindByParentRequestDTO) (*CollectionsResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Parent ID is required")
	}

	if req.ParentID.String() == "" {
		svc.logger.Warn("Empty parent ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("parent_id", "Parent ID is required")
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
	// STEP 3: Check if user has access to the parent collection
	//
	hasAccess, err := svc.repo.CheckAccess(ctx, req.ParentID, userID, dom_collection.CollectionPermissionReadOnly)
	if err != nil {
		svc.logger.Error("Failed to check access",
			zap.Any("error", err),
			zap.Any("parent_id", req.ParentID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized parent collection access attempt",
			zap.Any("user_id", userID),
			zap.Any("parent_id", req.ParentID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this parent collection")
	}

	//
	// STEP 4: Find collections by parent
	//
	collections, err := svc.repo.FindByParent(ctx, req.ParentID)
	if err != nil {
		svc.logger.Error("Failed to find collections by parent",
			zap.Any("error", err),
			zap.Any("parent_id", req.ParentID))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	response := &CollectionsResponseDTO{
		Collections: make([]*CollectionResponseDTO, len(collections)),
	}

	for i, collection := range collections {
		response.Collections[i] = mapCollectionToDTO(collection)
	}

	svc.logger.Debug("Found collections by parent",
		zap.Int("count", len(collections)),
		zap.Any("parent_id", req.ParentID))

	return response, nil
}
