// native/desktop/maplefile-cli/internal/usecase/fileupload/encrypt_file.go
package fileupload

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// EncryptFileUseCase handles file encryption
type EncryptFileUseCase interface {
	Execute(ctx context.Context, filePath string, fileKey []byte) (encryptedData []byte, hash string, err error)
}

type encryptFileUseCase struct {
	logger *zap.Logger
}

func NewEncryptFileUseCase(logger *zap.Logger) EncryptFileUseCase {
	return &encryptFileUseCase{
		logger: logger,
	}
}

func (uc *encryptFileUseCase) Execute(ctx context.Context, filePath string, fileKey []byte) ([]byte, string, error) {
	// Read file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", errors.NewAppError("failed to open file", err)
	}
	defer file.Close()

	// Read all content (for now - should use streaming for large files)
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, "", errors.NewAppError("failed to read file", err)
	}

	// Encrypt content
	encryptedData, err := crypto.EncryptWithSecretBox(content, fileKey)
	if err != nil {
		return nil, "", errors.NewAppError("failed to encrypt file", err)
	}

	// Combine nonce and ciphertext
	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)

	// Calculate hash of encrypted data
	hasher := sha256.New()
	hasher.Write(combined)
	hash := hex.EncodeToString(hasher.Sum(nil))

	return combined, hash, nil
}
