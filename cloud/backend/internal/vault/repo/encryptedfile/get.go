// cloud/backend/internal/vault/repo/encryptedfile/mongodb_repo.go
package encryptedfile

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
)

// GetByID retrieves an encrypted file by its ID
func (repo *encryptedFileRepository) GetByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*domain.EncryptedFile, error) {
	var file domain.EncryptedFile

	err := repo.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&file)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Not found
		}
		// Check for no documents error using string comparison since the constant might differ in v2
		if err.Error() == "mongo: no documents in result" {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get encrypted file: %w", err)
	}

	return &file, nil
}

// GetByFileID retrieves an encrypted file by user ID and file ID
func (repo *encryptedFileRepository) GetByFileID(
	ctx context.Context,
	userID primitive.ObjectID,
	fileID string,
) (*domain.EncryptedFile, error) {
	var file domain.EncryptedFile

	err := repo.collection.FindOne(
		ctx,
		bson.M{
			"user_id": userID,
			"file_id": fileID,
		},
	).Decode(&file)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Not found - returns nil, nil which is correct
		}
		// Check for no documents error using string comparison since the constant might differ in v2
		if err.Error() == "mongo: no documents in result" {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get encrypted file: %w", err)
	}

	return &file, nil
}
