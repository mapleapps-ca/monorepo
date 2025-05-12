package bannedipaddress

import (
	"context"
	"math/big"

	"go.uber.org/zap"

	dom_banip "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/bannedipaddress"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (impl bannedIPAddressImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*dom_banip.BannedIPAddress, error) {
	filter := bson.M{"_id": id}

	var result dom_banip.BannedIPAddress
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return nil, nil
		}
		impl.Logger.Error("database get by user transaction id error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}

func (impl bannedIPAddressImpl) GetByNonce(ctx context.Context, nonce *big.Int) (*dom_banip.BannedIPAddress, error) {
	filter := bson.M{"transaction.nonce_bytes": nonce.Bytes()}

	var result dom_banip.BannedIPAddress
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return nil, nil
		}
		impl.Logger.Error("database get by user transaction nonce error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}
