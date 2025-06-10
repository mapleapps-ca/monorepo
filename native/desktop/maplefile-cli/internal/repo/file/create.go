// monorepo/native/desktop/maplefile-cli/internal/repo/file/create.go
package file

import (
	"context"
	"time"

	"github.com/gocql/gocql"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) Create(ctx context.Context, file *dom_file.File) error {
	// Ensure file has an ID
	if file.ID.String() == "" {
		file.ID = gocql.TimeUUID()
	}

	// Set timestamps
	now := time.Now()
	if file.CreatedAt.String() == "" {
		file.CreatedAt = now
	}
	file.ModifiedAt = now

	// Save to local storage
	return r.Save(ctx, file)
}
