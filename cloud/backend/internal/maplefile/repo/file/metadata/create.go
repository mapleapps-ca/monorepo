// cloud/backend/internal/maplefile/repo/file/metadata/create.go
package metadata

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

// Create a new file metadata entry
func (impl fileMetadataRepositoryImpl) Create(file *dom_file.File) error {
	ctx := context.Background()

	// Validate file ID
	if file.ID == "" {
		file.ID = uuid.New().String()
	}

	// Set creation time if not set
	if file.CreatedAt.IsZero() {
		file.CreatedAt = time.Now()
	}

	// Set modification time to creation time
	file.ModifiedAt = file.CreatedAt

	// Insert file document
	_, err := impl.Collection.InsertOne(ctx, file)
	if err != nil {
		impl.Logger.Error("database failed create file error",
			zap.Any("error", err),
			zap.String("id", file.ID))
		return err
	}

	return nil
}
