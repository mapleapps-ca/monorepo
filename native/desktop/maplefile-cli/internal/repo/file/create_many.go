// monorepo/native/desktop/maplefile-cli/internal/repo/file/create_many.go
package file

import (
	"context"
	"errors"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) CreateMany(ctx context.Context, files []*dom_file.File) error {
	//TODO: IMPL.
	return errors.New("not implemented")
}
