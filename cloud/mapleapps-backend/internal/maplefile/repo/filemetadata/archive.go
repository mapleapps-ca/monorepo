// cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/archive.go
package filemetadata

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl *fileMetadataRepositoryImpl) Archive(id gocql.UUID) error {
	file, err := impl.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get file for archive: %w", err)
	}

	if file == nil {
		return fmt.Errorf("file not found")
	}

	// Validate state transition
	if err := dom_file.IsValidStateTransition(file.State, dom_file.FileStateArchived); err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	// Update file state
	file.State = dom_file.FileStateArchived
	file.ModifiedAt = time.Now()
	file.Version++

	return impl.Update(file)
}
