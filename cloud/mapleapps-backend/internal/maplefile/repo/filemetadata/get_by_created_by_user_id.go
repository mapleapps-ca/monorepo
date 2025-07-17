// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/get_by_created_by_user_id.go
package filemetadata

import (
	"fmt"

	"github.com/gocql/gocql"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl *fileMetadataRepositoryImpl) GetByCreatedByUserID(createdByUserID gocql.UUID) ([]*dom_file.File, error) {
	var fileIDs []gocql.UUID

	query := `SELECT file_id FROM mapleapps.maplefile_files_by_created_by_user_id_with_desc_created_at_and_asc_file_id
		WHERE created_by_user_id = ?`

	iter := impl.Session.Query(query, createdByUserID).Iter()

	var fileID gocql.UUID
	for iter.Scan(&fileID) {
		fileIDs = append(fileIDs, fileID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get files by creator: %w", err)
	}

	return impl.loadMultipleFiles(fileIDs)
}
