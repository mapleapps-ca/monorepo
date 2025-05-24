// monorepo/native/desktop/maplefile-cli/internal/repo/file/check_by_permission.go
package file

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *fileRepository) CheckIfUserHasAccess(ctx context.Context, fileID primitive.ObjectID, userID primitive.ObjectID) (bool, error) {
	//TODO: Impl.
	return false, errors.New("not implemented")
}
