// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/repo/user/delete.go
package user

import (
	"context"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (impl userStorerImpl) DeleteByID(ctx context.Context, id gocql.UUID) error {
	_, err := impl.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		impl.Logger.Error("database failed deletion error",
			zap.Any("error", err))
		return err
	}
	return nil
}

func (impl userStorerImpl) DeleteByEmail(ctx context.Context, email string) error {
	_, err := impl.Collection.DeleteOne(ctx, bson.M{"email": email})
	if err != nil {
		impl.Logger.Error("database failed deletion error",
			zap.Any("error", err))
		return err
	}
	return nil
}
