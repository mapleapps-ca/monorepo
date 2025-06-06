// cloud/backend/internal/maplefile/repo/filemetadata/create.go
package filemetadata

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

// Create a new file metadata entry
func (impl fileMetadataRepositoryImpl) Create(file *dom_file.File) error {
	ctx := context.Background()

	// Validate file ID
	if file.ID.IsZero() {
		file.ID = primitive.NewObjectID()
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
			zap.Any("id", file.ID))
		return err
	}

	return nil
}

// CreateMany inserts multiple file metadata entries
func (impl fileMetadataRepositoryImpl) CreateMany(files []*dom_file.File) error {
	if len(files) == 0 {
		return nil
	}

	ctx := context.Background()
	now := time.Now()

	// Prepare documents for insertion
	documents := make([]interface{}, len(files))
	for i, file := range files {
		// Validate file ID
		if file.ID.IsZero() {
			file.ID = primitive.NewObjectID()
		}

		// Set creation time if not set
		if file.CreatedAt.IsZero() {
			file.CreatedAt = now
		}

		// Set modification time to creation time
		file.ModifiedAt = file.CreatedAt

		documents[i] = file
	}

	// Insert all documents
	_, err := impl.Collection.InsertMany(ctx, documents)
	if err != nil {
		impl.Logger.Error("database failed batch create files error",
			zap.Any("error", err),
			zap.Int("count", len(files)))
		return err
	}

	return nil
}
