// cloud/backend/internal/vault/domain/encryptedfile/interface.go
package encryptedfile

import (
	"context"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository defines the operations for encrypted file storage
type Repository interface {
	// Core CRUD operations
	Create(ctx context.Context, file *EncryptedFile, encryptedContent io.Reader) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*EncryptedFile, error)
	GetByFileID(ctx context.Context, userID primitive.ObjectID, fileID string) (*EncryptedFile, error)
	UpdateByID(ctx context.Context, file *EncryptedFile, encryptedContent io.Reader) error
	DeleteByID(ctx context.Context, id primitive.ObjectID) error

	// List files for a user
	ListByUserID(ctx context.Context, userID primitive.ObjectID) ([]*EncryptedFile, error)
}
