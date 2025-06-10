// monorepo/native/desktop/maplefile-cli/internal/repo/file/check_by_permission.go
package file

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

func (r *fileRepository) CheckIfUserHasAccess(ctx context.Context, fileID gocql.UUID, userID gocql.UUID) (bool, error) {
	r.logger.Debug("Checking user access to file",
		zap.String("fileID", fileID.Hex()),
		zap.String("userID", userID.Hex()))

	// Get the file first
	file, err := r.Get(ctx, fileID)
	if err != nil {
		r.logger.Error("Failed to get file for permission check",
			zap.String("fileID", fileID.Hex()),
			zap.Error(err))
		return false, err
	}

	// File doesn't exist
	if file == nil {
		r.logger.Warn("File not found for permission check", zap.String("fileID", fileID.Hex()))
		return false, nil
	}

	// Check if user is the owner
	hasAccess := file.OwnerID == userID

	r.logger.Debug("User access check completed",
		zap.String("fileID", fileID.Hex()),
		zap.String("userID", userID.Hex()),
		zap.Bool("hasAccess", hasAccess))

	return hasAccess, nil
}
