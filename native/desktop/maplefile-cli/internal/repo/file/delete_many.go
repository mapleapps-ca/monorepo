// monorepo/native/desktop/maplefile-cli/internal/repo/file/delete_many.go
package file

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *fileRepository) DeleteMany(ctx context.Context, ids []primitive.ObjectID) error {
	//TODO: IMPL.
	return errors.New("not implemented")
}
