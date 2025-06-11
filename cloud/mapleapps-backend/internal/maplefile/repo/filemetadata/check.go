// cloud/mapleapps-backend/internal/maplefile/repo/filemetadata/check.go
package filemetadata

import (
	"fmt"

	"github.com/gocql/gocql"
)

func (impl *fileMetadataRepositoryImpl) CheckIfExistsByID(id gocql.UUID) (bool, error) {
	var count int

	query := `SELECT COUNT(*) FROM mapleapps.maplefile_files_by_id WHERE id = ?`

	if err := impl.Session.Query(query, id).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return count > 0, nil
}

func (impl *fileMetadataRepositoryImpl) CheckIfUserHasAccess(fileID gocql.UUID, userID gocql.UUID) (bool, error) {
	// Check if user has access via the user sync table
	var count int

	query := `SELECT COUNT(*) FROM mapleapps.maplefile_files_by_user_id_with_desc_modified_at_and_asc_file_id
		WHERE user_id = ? AND file_id = ? LIMIT 1 ALLOW FILTERING`

	err := impl.Session.Query(query, userID, fileID).Scan(&count)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file access: %w", err)
	}

	return count > 0, nil
}
