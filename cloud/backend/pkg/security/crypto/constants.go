package crypto

// Constants to ensure compatibility between Go and JavaScript
const (
	// Key sizes
	MasterKeySize        = 32 // 256-bit
	KeyEncryptionKeySize = 32
	CollectionKeySize    = 32
	FileKeySize          = 32
	RecoveryKeySize      = 32

	// Sodium/NaCl constants
	NonceSize         = 24
	PublicKeySize     = 32
	PrivateKeySize    = 32
	SealedBoxOverhead = 16

	// Argon2 parameters - must match between platforms
	Argon2IDAlgorithm = "argon2id"
	Argon2MemLimit    = 67108864 // 64 MB
	Argon2OpsLimit    = 4
	Argon2Parallelism = 1
	Argon2KeySize     = 32
	Argon2SaltSize    = 16
)
