// native/desktop/maplefile-cli/internal/repo/file/get_by_ids.go
package file

import (
	"context"
	"errors"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *fileRepository) GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*dom_file.File, error) {
	//TODO: Impl.
	return nil, errors.New("not implemented")
}
