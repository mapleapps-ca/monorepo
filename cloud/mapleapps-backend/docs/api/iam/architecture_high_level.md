# MapleApps E2EE Authentication Architecture

## Overview

MapleApps implements end-to-end encryption (E2EE) for all user data, ensuring that the server never has access to unencrypted user content. This document details the cryptographic algorithms and protocols used in the authentication system.

## Cryptographic Primitives

### Algorithms Used

| Purpose | Algorithm | Key Size | Nonce Size | Notes |
|---------|-----------|----------|------------|-------|
| Symmetric Encryption | ChaCha20-Poly1305 | 256 bits (32 bytes) | 96 bits (12 bytes) | Primary encryption algorithm |
| Asymmetric Encryption | NaCl box (X25519 + XSalsa20-Poly1305) | 256 bits (32 bytes) | 192 bits (24 bytes) | For key exchange and sealed boxes |
| Key Derivation (Password) | Argon2ID | 256 bits output | 128 bits salt (16 bytes) | For deriving KEK from password |
| Key Derivation (Legacy) | PBKDF2-SHA256 | 256 bits output | N/A | For compatibility (if needed) |
| Verification ID | BIP39 Mnemonic | N/A | N/A | Human-readable public key representation |

### Key Types and Sizes

```
Master Key (MK):              32 bytes (256 bits)
Key Encryption Key (KEK):     32 bytes (256 bits)
Public Key:                   32 bytes (256 bits)
Private Key:                  32 bytes (256 bits)
Recovery Key:                 32 bytes (256 bits)
Collection Key:               32 bytes (256 bits)
File Key:                     32 bytes (256 bits)
```

## Registration Flow

### 1. Client-Side Key Generation

```javascript
// 1. Generate Master Key (MK)
const masterKey = crypto.getRandomValues(new Uint8Array(32));

// 2. Generate Key Pair for asymmetric encryption
const keyPair = await crypto.subtle.generateKey(
  { name: "X25519" },
  true,
  ["deriveKey"]
);

// 3. Generate Recovery Key
const recoveryKey = crypto.getRandomValues(new Uint8Array(32));

// 4. Generate salt for Argon2ID
const salt = crypto.getRandomValues(new Uint8Array(16));
```

### 2. Key Derivation from Password

```javascript
// Derive Key Encryption Key (KEK) using Argon2ID
const kek = await argon2id({
  password: userPassword,
  salt: salt,
  iterations: 4,          // Time cost
  memory: 65536,          // 64 MB
  parallelism: 1,
  keyLength: 32
});
```

**Argon2ID Parameters:**
- Algorithm: `argon2id`
- Memory: 64 MB (67108864 bytes)
- Iterations: 4
- Parallelism: 1
- Salt Length: 16 bytes
- Output Length: 32 bytes

### 3. Encryption of Keys

```javascript
// All encryption uses ChaCha20-Poly1305 with 12-byte nonce

// 1. Encrypt Master Key with KEK
const encryptedMasterKey = encrypt_ChaCha20Poly1305(masterKey, kek);
// Output format: nonce (12 bytes) || ciphertext

// 2. Encrypt Private Key with Master Key
const encryptedPrivateKey = encrypt_ChaCha20Poly1305(privateKey, masterKey);

// 3. Encrypt Recovery Key with Master Key
const encryptedRecoveryKey = encrypt_ChaCha20Poly1305(recoveryKey, masterKey);

// 4. Encrypt Master Key with Recovery Key (for recovery)
const masterKeyWithRecoveryKey = encrypt_ChaCha20Poly1305(masterKey, recoveryKey);
```

### 4. Verification ID Generation

```javascript
// Generate deterministic Verification ID from public key
const verificationID = generateVerificationID(publicKey);

// Algorithm:
// 1. SHA256 hash of public key
// 2. Use hash as entropy for BIP39 mnemonic generation
// 3. Result: 24-word mnemonic phrase
```

### 5. Registration Request Format

```json
{
  "beta_access_code": "string",
  "first_name": "string",
  "last_name": "string",
  "email": "string",
  "phone": "string",
  "country": "string",
  "timezone": "string",
  "agree_terms_of_service": true,
  "module": 1,  // 1 = MapleFile, 2 = PaperCloud

  // E2EE fields (all base64 URL encoded)
  "salt": "base64url_encoded_salt",
  "publicKey": "base64url_encoded_public_key",
  "encryptedMasterKey": "base64url_encoded(nonce + ciphertext)",
  "encryptedPrivateKey": "base64url_encoded(nonce + ciphertext)",
  "encryptedRecoveryKey": "base64url_encoded(nonce + ciphertext)",
  "masterKeyEncryptedWithRecoveryKey": "base64url_encoded(nonce + ciphertext)",
  "verificationID": "24_word_mnemonic_phrase"
}
```

## Login Flow (3-Step E2EE Process)

### Step 1: Request One-Time Token (OTT)

```json
POST /iam/api/v1/request-ott
{
  "email": "user@example.com"
}
```

Server generates 6-digit OTT and sends via email.

### Step 2: Verify OTT and Receive Challenge

```json
POST /iam/api/v1/verify-ott
{
  "email": "user@example.com",
  "ott": "123456"
}
```

**Server Response:**
```json
{
  "salt": "base64_encoded_salt",
  "kdf_params": {
    "algorithm": "argon2id",
    "iterations": 4,
    "memory": 65536,
    "parallelism": 1
  },
  "publicKey": "base64_encoded_public_key",
  "encryptedMasterKey": "base64_encoded(nonce + ciphertext)",
  "encryptedPrivateKey": "base64_encoded(nonce + ciphertext)",
  "encryptedChallenge": "base64_encoded_sealed_challenge",
  "challengeId": "uuid"
}
```

**Challenge Encryption:**
```go
// Server generates 32-byte random challenge
challenge := generateRandom(32)

// Encrypt using NaCl box seal (anonymous encryption)
sealedChallenge := box.SealAnonymous(challenge, userPublicKey)
// Output format: ephemeral_public_key (32 bytes) || ciphertext
```

### Step 3: Complete Login

**Client-Side Challenge Decryption:**
```javascript
// 1. Derive KEK from password
const kek = await deriveKey(password, salt, kdfParams);

// 2. Decrypt master key
const masterKey = decrypt_ChaCha20Poly1305(encryptedMasterKey, kek);

// 3. Decrypt private key
const privateKey = decrypt_ChaCha20Poly1305(encryptedPrivateKey, masterKey);

// 4. Decrypt challenge using NaCl box seal open
const challenge = crypto_box_seal_open(sealedChallenge, publicKey, privateKey);
```

**Complete Login Request:**
```json
POST /iam/api/v1/complete-login
{
  "email": "user@example.com",
  "challengeId": "uuid",
  "decryptedData": "base64url_encoded_challenge"
}
```

**Server Response with Encrypted Tokens:**
```json
{
  "encrypted_access_token": "base64_encoded_encrypted_token",
  "encrypted_refresh_token": "base64_encoded_encrypted_token",
  "token_nonce": "base64_encoded_nonce",
  "access_token_expiry_date": "2024-01-15T11:00:00Z",
  "refresh_token_expiry_date": "2024-01-29T10:30:00Z"
}
```

### Token Encryption Details

Tokens are encrypted using NaCl box with the user's public key:

```go
// Server-side token encryption
// 1. Generate ephemeral keypair
ephemeralPubKey, ephemeralPrivKey := box.GenerateKey()

// 2. Generate nonce (24 bytes for NaCl box)
nonce := generateRandom(24)

// 3. Encrypt tokens separately
encryptedAccessToken := box.Seal(accessToken, nonce, userPublicKey, ephemeralPrivKey)
encryptedRefreshToken := box.Seal(refreshToken, nonce, userPublicKey, ephemeralPrivKey)

// 4. Combined format for each token:
// ephemeral_public_key (32) || nonce (24) || encrypted_token
```

## Token Refresh Flow

```json
POST /iam/api/v1/token/refresh
{
  "value": "encrypted_refresh_token"
}
```

The server validates the refresh token and returns new encrypted tokens using the same encryption method as login.

## Account Recovery Flow (3-Step Process)

### Step 1: Initiate Recovery

```json
POST /iam/api/v1/recovery/initiate
{
  "email": "user@example.com",
  "method": "recovery_key"
}
```

Server generates and encrypts a challenge with user's public key.

### Step 2: Verify Recovery

```json
POST /iam/api/v1/recovery/verify
{
  "session_id": "uuid",
  "decrypted_challenge": "base64url_encoded_challenge"
}
```

Client must decrypt challenge using recovery key to prove ownership.

**Response includes:**
```json
{
  "recovery_token": "base64_encoded_token",
  "master_key_encrypted_with_recovery_key": "base64_encoded_data"
}
```

### Step 3: Complete Recovery

```json
POST /iam/api/v1/recovery/complete
{
  "recovery_token": "token_from_step_2",
  "new_salt": "base64url_encoded_new_salt",
  "new_encrypted_master_key": "base64url_encoded(nonce + ciphertext)",
  "new_encrypted_private_key": "base64url_encoded(nonce + ciphertext)",
  "new_encrypted_recovery_key": "base64url_encoded(nonce + ciphertext)",
  "new_master_key_encrypted_with_recovery_key": "base64url_encoded(nonce + ciphertext)"
}
```

## Encoding Standards

### Base64 Variants

- **Transport (JSON)**: Base64 URL encoding without padding (`base64.RawURLEncoding`)
- **Storage**: Standard Base64 encoding (`base64.StdEncoding`)

### Data Format Conventions

1. **Encrypted Data Format**: `nonce || ciphertext`
   - ChaCha20-Poly1305: 12-byte nonce + ciphertext
   - NaCl box: 24-byte nonce + ciphertext

2. **Sealed Box Format**: `ephemeral_public_key || ciphertext`
   - 32-byte ephemeral public key + encrypted data

3. **Token Format**: `ephemeral_public_key || nonce || ciphertext`
   - 32-byte key + 24-byte nonce + encrypted token

## Implementation Libraries

### JavaScript/TypeScript
```javascript
// Recommended libraries
import * as sodium from 'libsodium-wrappers-sumo';
import { argon2id } from 'argon2-browser';
import * as bip39 from '@scure/bip39';

// For ChaCha20-Poly1305
sodium.crypto_aead_chacha20poly1305_ietf_encrypt()
sodium.crypto_aead_chacha20poly1305_ietf_decrypt()

// For NaCl box (asymmetric)
sodium.crypto_box_seal()
sodium.crypto_box_seal_open()
```

### Go
```go
import (
    "golang.org/x/crypto/chacha20poly1305"  // ChaCha20-Poly1305
    "golang.org/x/crypto/nacl/box"          // NaCl box
    "golang.org/x/crypto/argon2"            // Argon2ID
    "github.com/tyler-smith/go-bip39"       // BIP39 mnemonics
)
```

### React Native
```javascript
// Use react-native-sodium or react-native-libsodium
import { NaCl } from 'react-native-sodium';
```

## Security Considerations

1. **Zero-Knowledge Architecture**: Server never sees plaintext passwords or unencrypted keys
2. **Forward Secrecy**: Each login generates new session tokens
3. **Key Rotation**: Support for key version tracking and rotation
4. **Rate Limiting**:
   - OTT: Valid for 10 minutes
   - Challenge: Valid for 5 minutes
   - Recovery attempts: Max 5 per 15 minutes

## Migration Notes

- The system migrated from XSalsa20-Poly1305 to ChaCha20-Poly1305 for symmetric encryption
- Asymmetric encryption still uses NaCl box (X25519 + XSalsa20-Poly1305) for compatibility
- Key version tracking allows for gradual migration of encrypted data
