// monorepo/cloud/backend/internal/maplefile/service/collection/move_collection.go
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

type MoveCollectionRequestDTO struct {
	CollectionID        gocql.UUID   `json:"collection_id"`
	NewParentID         gocql.UUID   `json:"new_parent_id"`
	UpdatedAncestors    []gocql.UUID `json:"updated_ancestors"`
	UpdatedPathSegments []string     `json:"updated_path_segments"`
}

type MoveCollectionResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MoveCollectionService interface {
	Execute(ctx context.Context, req *MoveCollectionRequestDTO) (*MoveCollectionResponseDTO, error)
}

type moveCollectionServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewMoveCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) MoveCollectionService {
	logger = logger.Named("MoveCollectionService")
	return &moveCollectionServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *moveCollectionServiceImpl) Execute(ctx context.Context, req *MoveCollectionRequestDTO) (*MoveCollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Move details are required")
	}

	e := make(map[string]string)
	if req.CollectionID.String() == "" {
		e["collection_id"] = "Collection ID is required"
	}
	if req.NewParentID.String() == "" {
		e["new_parent_id"] = "New parent ID is required"
	}
	if len(req.UpdatedAncestors) == 0 {
		e["updated_ancestors"] = "Updated ancestors are required"
	}
	if len(req.UpdatedPathSegments) == 0 {
		e["updated_path_segments"] = "Updated path segments are required"
	}

	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
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
	// STEP 3: Check if user has write access to the collection
	//
	hasAccess, err := svc.repo.CheckAccess(ctx, req.CollectionID, userID, dom_collection.CollectionPermissionReadWrite)
	if err != nil {
		svc.logger.Error("Failed to check access",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized collection move attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to move this collection")
	}

	//
	// STEP 4: Check if user has write access to the new parent
	//
	hasParentAccess, err := svc.repo.CheckAccess(ctx, req.NewParentID, userID, dom_collection.CollectionPermissionReadWrite)
	if err != nil {
		svc.logger.Error("Failed to check access to new parent",
			zap.Any("error", err),
			zap.Any("new_parent_id", req.NewParentID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasParentAccess {
		svc.logger.Warn("Unauthorized destination parent access",
			zap.Any("user_id", userID),
			zap.Any("new_parent_id", req.NewParentID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to move to this destination")
	}

	//
	// STEP 5: Move the collection
	//
	err = svc.repo.MoveCollection(
		ctx,
		req.CollectionID,
		req.NewParentID,
		req.UpdatedAncestors,
		req.UpdatedPathSegments,
	)
	if err != nil {
		svc.logger.Error("Failed to move collection",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID),
			zap.Any("new_parent_id", req.NewParentID))
		return nil, err
	}

	svc.logger.Info("Collection moved successfully",
		zap.Any("collection_id", req.CollectionID),
		zap.Any("new_parent_id", req.NewParentID))

	return &MoveCollectionResponseDTO{
		Success: true,
		Message: "Collection moved successfully",
	}, nil
}
