// native/desktop/maplefile-cli/internal/repo/file/get_by_collection.go
package file

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"

	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

func (r *fileRepository) GetByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*dom_file.File, error) {
	//TODO: Impl.
	return nil, errors.New("not implemented")
}
