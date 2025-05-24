// monorepo/native/desktop/maplefile-cli/internal/repo/file/check_by_id.go
package file

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *fileRepository) CheckIfExistsByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
	//TODO: Impl.
	return false, errors.New("not implemented")
}
