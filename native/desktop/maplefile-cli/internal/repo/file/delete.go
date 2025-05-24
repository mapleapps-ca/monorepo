// monorepo/native/desktop/maplefile-cli/internal/repo/file/delete.go
package file

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *fileRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	//TODO: IMPL.
	return errors.New("not implemented")
}
