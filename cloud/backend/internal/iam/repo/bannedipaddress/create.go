package bannedipaddress

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	dom_banip "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/bannedipaddress"
)

func (impl bannedIPAddressImpl) Create(ctx context.Context, u *dom_banip.BannedIPAddress) error {
	// DEVELOPER NOTES:
	// According to mongodb documentaiton:
	//     Non-existent Databases and Collections
	//     If the necessary database and collection don't exist when you perform a write operation, the server implicitly creates them.
	//     Source: https://www.mongodb.com/docs/drivers/go/current/usage-examples/insertOne/

	if u.ID == primitive.NilObjectID {
		u.ID = primitive.NewObjectID()
		impl.Logger.Warn("database insert user transaction not included id value, created id now.", zap.Any("id", u.ID))
	}

	_, err := impl.Collection.InsertOne(ctx, u)

	// check for errors in the insertion
	if err != nil {
		impl.Logger.Error("database failed create error",
			zap.Any("error", err))
		return err
	}

	return nil
}
