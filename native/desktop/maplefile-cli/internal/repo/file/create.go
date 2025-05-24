// monorepo/native/desktop/maplefile-cli/internal/repo/file/create.go
package file

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

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

	// Save to local storage
	return r.Save(ctx, file)
}
