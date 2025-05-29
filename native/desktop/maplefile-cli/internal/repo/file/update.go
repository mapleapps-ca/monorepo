// native/desktop/maplefile-cli/internal/repo/file/update.go
package file

import (
	"context"

	"go.uber.org/zap"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) Update(ctx context.Context, file *dom_file.File) error {
	r.logger.Debug("💾 Updating file in local storage",
		zap.String("fileID", file.ID.Hex()),
		zap.String("fileName", file.Name))

	// Use the save method which handles serialization and storage
	return r.Save(ctx, file)
}
