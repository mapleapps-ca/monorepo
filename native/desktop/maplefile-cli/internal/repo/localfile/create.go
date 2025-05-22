// monorepo/native/desktop/maplefile-cli/internal/repo/localfile/create.go
package localfile

import (
	"context"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// Create creates a new local file
func (r *localFileRepository) Create(ctx context.Context, file *localfile.LocalFile) error {
	r.logger.Debug("Creating new local file",
		zap.String("remoteID", file.RemoteID.Hex()))

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
		file.SyncStatus = localfile.SyncStatusLocalOnly
	}

	// Save the file metadata to storage
	return r.Save(ctx, file)
}
