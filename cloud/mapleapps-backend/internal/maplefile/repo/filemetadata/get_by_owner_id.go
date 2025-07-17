// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/get_by_owner_id.go
package filemetadata

import (
	"fmt"

	"github.com/gocql/gocql"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl *fileMetadataRepositoryImpl) GetByOwnerID(ownerID gocql.UUID) ([]*dom_file.File, error) {
	var fileIDs []gocql.UUID

	query := `SELECT file_id FROM mapleapps.maplefile_files_by_owner_id_with_desc_modified_at_and_asc_file_id
		WHERE owner_id = ?`

	iter := impl.Session.Query(query, ownerID).Iter()

	var fileID gocql.UUID
	for iter.Scan(&fileID) {
		fileIDs = append(fileIDs, fileID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get files by owner: %w", err)
	}

	return impl.loadMultipleFiles(fileIDs)
}
