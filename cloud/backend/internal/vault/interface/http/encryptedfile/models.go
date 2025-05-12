// cloud/backend/internal/vault/interface/http/encryptedfile/models.go
package encryptedfile

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileResponse represents file metadata returned in HTTP responses
type FileResponse struct {
	ID                primitive.ObjectID `json:"id"`
	UserID            primitive.ObjectID `json:"user_id"`
	FileID            string             `json:"file_id"`
	EncryptedMetadata string             `json:"encrypted_metadata"`
	EncryptionVersion string             `json:"encryption_version"`
	EncryptedHash     string             `json:"encrypted_hash"`
	CreatedAt         time.Time          `json:"created_at"`
	ModifiedAt        time.Time          `json:"modified_at"`
}

// FilesListResponse represents a list of file metadata
type FilesListResponse struct {
	Files []FileResponse `json:"files"`
}

// FileURLResponse represents a presigned download URL for a file
type FileURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}
