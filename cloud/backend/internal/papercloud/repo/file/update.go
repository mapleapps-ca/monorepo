// cloud/backend/internal/papercloud/repo/file/update.go
package file

import (
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
)

// Update implements the FileRepository.Update method
func (repo *fileRepositoryImpl) Update(file *dom_file.File) error {
	return repo.metadata.Update(file)
}
