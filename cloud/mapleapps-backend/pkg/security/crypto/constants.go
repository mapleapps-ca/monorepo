package crypto

// Constants to ensure compatibility between Go and JavaScript
const (
	// Key sizes
	MasterKeySize        = 32 // 256-bit
	KeyEncryptionKeySize = 32
	CollectionKeySize    = 32
	FileKeySize          = 32
	RecoveryKeySize      = 32

	// ChaCha20-Poly1305 constants (updated from XSalsa20-Poly1305)
	NonceSize         = 12 // ChaCha20-Poly1305 nonce size (changed from 24)
	PublicKeySize     = 32
	PrivateKeySize    = 32
	SealedBoxOverhead = 16

	// Legacy naming for backward compatibility
	SecretBoxNonceSize = NonceSize

	// Argon2 parameters - must match between platforms
	Argon2IDAlgorithm = "argon2id"
	Argon2MemLimit    = 67108864 // 64 MB
	Argon2OpsLimit    = 4
	Argon2Parallelism = 1
	Argon2KeySize     = 32
	Argon2SaltSize    = 16

	// Encryption algorithm identifiers
	ChaCha20Poly1305Algorithm = "chacha20poly1305" // Primary algorithm
	XSalsa20Poly1305Algorithm = "xsalsa20poly1305" // Legacy algorithm (deprecated)
)
