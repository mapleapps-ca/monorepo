// github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/repo/federateduser/check.go
package federateduser

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (impl userStorerImpl) CheckIfExistsByID(ctx context.Context, id primitive.ObjectID) (bool, error) {
	filter := bson.M{"_id": id}
	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check if exists by ID error", zap.Any("error", err))
		return false, err
	}
	return count >= 1, nil
}

func (impl userStorerImpl) CheckIfExistsByEmail(ctx context.Context, email string) (bool, error) {
	filter := bson.M{"email": email}
	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check if exists by email error", zap.Any("error", err))
		return false, err
	}
	return count >= 1, nil
}
