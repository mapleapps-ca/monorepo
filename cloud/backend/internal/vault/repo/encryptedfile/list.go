// cloud/backend/internal/vault/repo/encryptedfile/mongodb_repo.go
package encryptedfile

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
)

// ListByUserID lists all encrypted files for a user
func (repo *encryptedFileRepository) ListByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
) ([]*domain.EncryptedFile, error) {
	// Define query options for sorting by creation time
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	// Execute the query
	cursor, err := repo.collection.Find(
		ctx,
		bson.M{"user_id": userID},
		findOptions,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to list encrypted files: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode the results
	var files []*domain.EncryptedFile
	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("failed to decode encrypted files: %w", err)
	}

	return files, nil
}
