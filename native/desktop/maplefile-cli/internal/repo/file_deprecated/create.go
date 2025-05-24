// monorepo/native/desktop/maplefile-cli/internal/repo/file/create.go
package file

import (
	"context"
	"time"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Create creates a new local file
func (r *fileRepository) Create(ctx context.Context, file *dom_file.File) error {
	// Ensure file has an ID
	if file.ID.IsZero() {
		file.ID = primitive.NewObjectID()
	}

	// Set timestamps
	now := time.Now()
	if file.CreatedAt.IsZero() {
		file.CreatedAt = now
	}
	file.ModifiedAt = now

	// Set as local only initially if not set
	if file.SyncStatus == 0 {
		file.SyncStatus = dom_file.SyncStatusLocalOnly
	}

	// Save the file metadata to storage
	return r.Save(ctx, file)
}
