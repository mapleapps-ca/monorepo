// internal/service/collectioncrypto/constants.go
package collectioncrypto

const (
	ErrInvalidPassword         = "invalid password - unable to decrypt master key"
	ErrCollectionKeyMissing    = "collection has no encrypted key"
	ErrCollectionKeyInvalid    = "collection encrypted key is invalid or corrupted"
	ErrUserNotCollectionMember = "user is not a member of this collection"
	ErrEncryptionFailed        = "encryption operation failed"
	ErrDecryptionFailed        = "decryption operation failed"
)
