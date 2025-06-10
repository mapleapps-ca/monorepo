// internal/service/collectioncrypto/multidecrypt.go
package collectioncrypto

import (
	"context"

	"go.uber.org/zap"

	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

func (s *collectionDecryptionService) ExecuteDecryptMultipleCollectionKeys(
	ctx context.Context,
	user *dom_user.User,
	collections []*dom_collection.Collection,
	password string,
) (map[string][]byte, error) {
	s.logger.Debug("🔐 Starting batch collection key decryption",
		zap.String("userID", user.ID.String()),
		zap.Int("collectionCount", len(collections)))

	// Decrypt master key once for efficiency
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		return nil, NewCryptoError("derive_key", err, user.ID.String())
	}
	defer crypto.ClearBytes(keyEncryptionKey)

	masterKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return nil, NewCryptoError("decrypt_master_key", err, user.ID.String())
	}
	defer crypto.ClearBytes(masterKey)

	// Decrypt all collection keys
	results := make(map[string][]byte)
	for _, collection := range collections {
		if collection.EncryptedCollectionKey == nil {
			s.logger.Warn("⚠️ Skipping collection with no encrypted key",
				zap.String("collectionID", collection.ID.String()))
			continue
		}

		collectionKey, err := crypto.DecryptWithSecretBox(
			collection.EncryptedCollectionKey.Ciphertext,
			collection.EncryptedCollectionKey.Nonce,
			masterKey,
		)
		if err != nil {
			s.logger.Warn("⚠️ Failed to decrypt collection key",
				zap.String("collectionID", collection.ID.String()),
				zap.Error(err))
			continue
		}

		results[collection.ID.String()] = collectionKey
	}

	s.logger.Debug("✅ Successfully completed batch collection key decryption",
		zap.String("userID", user.ID.String()),
		zap.Int("successfulDecryptions", len(results)))

	return results, nil
}
