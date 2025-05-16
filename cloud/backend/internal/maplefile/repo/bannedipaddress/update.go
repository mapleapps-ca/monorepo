package bannedipaddress

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"

	dom_banip "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/bannedipaddress"
)

func (impl bannedIPAddressImpl) UpdateByID(ctx context.Context, m *dom_banip.BannedIPAddress) error {
	filter := bson.M{"_id": m.ID}

	update := bson.M{ // DEVELOPERS NOTE: https://stackoverflow.com/a/60946010
		"$set": m,
	}

	// execute the UpdateOne() function to update the first matching document
	_, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database update user transaction by id error", zap.Any("error", err))
		return err
	}

	// // display the number of documents updated
	// impl.Logger.Debug("number of documents updated", zap.Int64("modified_count", result.ModifiedCount))

	return nil
}
