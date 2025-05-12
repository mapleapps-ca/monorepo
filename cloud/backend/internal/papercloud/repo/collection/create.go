// github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/repo/collection/create.go
package collection

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
)

func (impl collectionRepositoryImpl) Create(collection *dom_collection.Collection) error {
	ctx := context.Background()

	// Validate collection ID
	if collection.ID == "" {
		collection.ID = uuid.New().String()
	}

	// Validate owner ID
	if collection.OwnerID == "" {
		impl.Logger.Error("owner ID is required but not provided")
		return errors.New("owner ID is required")
	}

	// Set creation time if not set
	if collection.CreatedAt.IsZero() {
		collection.CreatedAt = time.Now()
	}

	// Set update time to match creation time
	collection.UpdatedAt = collection.CreatedAt

	// Initialize empty members array if not set
	if collection.Members == nil {
		collection.Members = []dom_collection.CollectionMembership{}
	}

	// Insert collection document
	_, err := impl.Collection.InsertOne(ctx, collection)

	if err != nil {
		impl.Logger.Error("database failed create collection error",
			zap.Any("error", err),
			zap.String("id", collection.ID))
		return err
	}

	return nil
}
