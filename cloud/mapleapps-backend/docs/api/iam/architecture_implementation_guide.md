# Detailed Library Implementation Guide

## Libraries Used in Current Implementation

### Go Backend Libraries

```go
// Core Encryption Libraries
import (
    // Symmetric encryption (ChaCha20-Poly1305)
    "golang.org/x/crypto/chacha20poly1305"

    // Asymmetric encryption (NaCl box)
    "golang.org/x/crypto/nacl/box"

    // Key derivation
    "golang.org/x/crypto/argon2"

    // BIP39 mnemonic generation
    "github.com/tyler-smith/go-bip39"

    // Random number generation
    "crypto/rand"

    // Hashing
    "crypto/sha256"
)
```

### JavaScript/React.js Recommended Libraries

```javascript
// Package.json dependencies
{
  "dependencies": {
    "libsodium-wrappers-sumo": "^0.7.13",  // Main crypto library
    "@scure/bip39": "^1.2.1",               // BIP39 mnemonics
    "argon2-browser": "^1.18.0",            // Argon2 in browser
    "@noble/ciphers": "^0.4.1"              // Alternative for ChaCha20
  }
}
```

## Detailed Encryption Service Implementations

### 1. Token Encryption Service

**Go Implementation (`token_encryption.go`):**

```go
// Uses NaCl box (X25519 + XSalsa20-Poly1305)
import "golang.org/x/crypto/nacl/box"

func EncryptTokens(accessToken, refreshToken string, publicKey []byte) {
    // 1. Generate ephemeral keypair
    ephemeralPubKey, ephemeralPrivKey, _ := box.GenerateKey(rand.Reader)

    // 2. Generate 24-byte nonce for NaCl box
    nonce := make([]byte, 24)
    rand.Read(nonce)

    // 3. Encrypt using box.Seal
    encryptedAccessToken := box.Seal(nil, []byte(accessToken), &nonce, &recipientPubKey, ephemeralPrivKey)

    // 4. Output format:
    // [32 bytes ephemeral pubkey][24 bytes nonce][encrypted data]
}
```

**JavaScript/React.js Implementation:**

```javascript
import _sodium from 'libsodium-wrappers-sumo';

async function decryptToken(encryptedTokenBase64, publicKey, privateKey) {
    await _sodium.ready;
    const sodium = _sodium;

    // Decode from base64
    const encryptedData = sodium.from_base64(encryptedTokenBase64, sodium.base64_variants.ORIGINAL);

    // Extract components
    const ephemeralPubKey = encryptedData.slice(0, 32);
    const nonce = encryptedData.slice(32, 56);
    const ciphertext = encryptedData.slice(56);

    // Decrypt using crypto_box_open_easy
    const decrypted = sodium.crypto_box_open_easy(
        ciphertext,
        nonce,
        ephemeralPubKey,
        privateKey
    );

    return sodium.to_string(decrypted);
}
```

### 2. Challenge Encryption (Sealed Box)

**Go Implementation:**

```go
import "golang.org/x/crypto/nacl/box"

// Server-side challenge encryption
func encryptChallenge(challenge []byte, userPublicKey []byte) ([]byte, error) {
    var pubKeyArray [32]byte
    copy(pubKeyArray[:], userPublicKey)

    // SealAnonymous creates ephemeral keypair internally
    sealed, err := box.SealAnonymous(nil, challenge, &pubKeyArray, rand.Reader)
    // Output: [32 bytes ephemeral pubkey][ciphertext with MAC]
    return sealed, err
}
```

**JavaScript/React.js Implementation:**

```javascript
async function decryptChallenge(sealedChallengeBase64, publicKey, privateKey) {
    await _sodium.ready;
    const sodium = _sodium;

    const sealedChallenge = sodium.from_base64(
        sealedChallengeBase64,
        sodium.base64_variants.ORIGINAL
    );

    // crypto_box_seal_open handles the ephemeral key extraction
    const challenge = sodium.crypto_box_seal_open(
        sealedChallenge,
        publicKey,
        privateKey
    );

    // Return as base64 URL encoded for API
    return sodium.to_base64(challenge, sodium.base64_variants.URLSAFE_NO_PADDING);
}
```

### 3. Symmetric Key Encryption (ChaCha20-Poly1305)

**Go Implementation:**

```go
import "golang.org/x/crypto/chacha20poly1305"

func encryptWithChaCha20(data, key []byte) ([]byte, error) {
    cipher, err := chacha20poly1305.New(key)
    if err != nil {
        return nil, err
    }

    // Generate 12-byte nonce
    nonce := make([]byte, 12)
    rand.Read(nonce)

    // Encrypt
    ciphertext := cipher.Seal(nil, nonce, data, nil)

    // Return nonce + ciphertext
    return append(nonce, ciphertext...), nil
}
```

**JavaScript/React.js Implementation:**

```javascript
async function encryptWithChaCha20(data, key) {
    await _sodium.ready;
    const sodium = _sodium;

    // Generate nonce
    const nonce = sodium.randombytes_buf(sodium.crypto_aead_chacha20poly1305_IETF_NPUBBYTES);

    // Encrypt
    const ciphertext = sodium.crypto_aead_chacha20poly1305_ietf_encrypt(
        data,
        null,  // no additional data
        null,  // no secret nonce
        nonce,
        key
    );

    // Combine nonce + ciphertext
    const combined = new Uint8Array(nonce.length + ciphertext.length);
    combined.set(nonce);
    combined.set(ciphertext, nonce.length);

    return sodium.to_base64(combined, sodium.base64_variants.URLSAFE_NO_PADDING);
}
```

### 4. Key Derivation (Argon2ID)

**Go Implementation:**

```go
import "golang.org/x/crypto/argon2"

func deriveKey(password string, salt []byte) []byte {
    return argon2.IDKey(
        []byte(password),
        salt,
        4,        // time cost
        65536,    // memory in KB (64 MB)
        1,        // parallelism
        32        // key length
    )
}
```

**JavaScript/React.js Implementation:**

```javascript
import { argon2id } from 'argon2-browser';

async function deriveKey(password, saltBase64) {
    const salt = sodium.from_base64(saltBase64, sodium.base64_variants.URLSAFE_NO_PADDING);

    const key = await argon2id({
        password: password,
        salt: salt,
        iterations: 4,
        memory: 65536,  // 64 MB
        parallelism: 1,
        hashLength: 32
    });

    return key.hash;  // Uint8Array
}
```

## API Endpoints and Encryption Usage

### Registration Endpoint
**`POST /iam/api/v1/register`**

| Operation | Library | Algorithm | Purpose |
|-----------|---------|-----------|---------|
| Salt Generation | `crypto/rand` | Random bytes | Generate 16-byte salt |
| KEK Derivation | `argon2` | Argon2ID | Derive key from password |
| Master Key Encryption | `chacha20poly1305` | ChaCha20-Poly1305 | Encrypt with KEK |
| Private Key Encryption | `chacha20poly1305` | ChaCha20-Poly1305 | Encrypt with Master Key |
| Recovery Key Encryption | `chacha20poly1305` | ChaCha20-Poly1305 | Encrypt with Master Key |
| Verification ID | `go-bip39` | SHA256 + BIP39 | Generate from public key |

### Login Flow Endpoints

#### `POST /iam/api/v1/verify-ott`
| Operation | Library | Algorithm | Purpose |
|-----------|---------|-----------|---------|
| Challenge Generation | `crypto/rand` | Random bytes | 32-byte challenge |
| Challenge Encryption | `nacl/box` | X25519 + XSalsa20-Poly1305 | Sealed box encryption |

#### `POST /iam/api/v1/complete-login`
| Operation | Library | Algorithm | Purpose |
|-----------|---------|-----------|---------|
| Challenge Verification | Native Go | Byte comparison | Verify decrypted challenge |
| Token Generation | `jwt` | HMAC-SHA256 | Create JWT tokens |
| Token Encryption | `nacl/box` | X25519 + XSalsa20-Poly1305 | Encrypt tokens with user's public key |

### Token Refresh Endpoint
**`POST /iam/api/v1/token/refresh`**

| Operation | Library | Algorithm | Purpose |
|-----------|---------|-----------|---------|
| Token Validation | `jwt` | HMAC-SHA256 | Validate refresh token |
| New Token Generation | `jwt` | HMAC-SHA256 | Generate new tokens |
| Token Encryption | `nacl/box` | X25519 + XSalsa20-Poly1305 | Encrypt with user's public key |

### Recovery Flow Endpoints

#### `POST /iam/api/v1/recovery/initiate`
| Operation | Library | Algorithm | Purpose |
|-----------|---------|-----------|---------|
| Challenge Generation | `crypto/rand` | Random bytes | 32-byte challenge |
| Challenge Encryption | `nacl/box` | X25519 + XSalsa20-Poly1305 | Sealed box with public key |

#### `POST /iam/api/v1/recovery/complete`
| Operation | Library | Algorithm | Purpose |
|-----------|---------|-----------|---------|
| New Salt Generation | `crypto/rand` | Random bytes | 16-byte salt |
| Key Re-encryption | `chacha20poly1305` | ChaCha20-Poly1305 | All keys with new password |

## Complete React.js Implementation Example

```javascript
// crypto-service.js
import _sodium from 'libsodium-wrappers-sumo';
import { argon2id } from 'argon2-browser';
import { generateMnemonic } from '@scure/bip39';
import { wordlist } from '@scure/bip39/wordlists/english';

class CryptoService {
    constructor() {
        this.sodium = null;
    }

    async initialize() {
        await _sodium.ready;
        this.sodium = _sodium;
    }

    // Generate key pair for new user
    generateKeyPair() {
        const keypair = this.sodium.crypto_box_keypair();
        return {
            publicKey: keypair.publicKey,
            privateKey: keypair.privateKey
        };
    }

    // Derive KEK from password
    async deriveKEK(password, salt) {
        const key = await argon2id({
            password: password,
            salt: salt,
            iterations: 4,
            memory: 65536,
            parallelism: 1,
            hashLength: 32
        });
        return key.hash;
    }

    // Encrypt data with ChaCha20-Poly1305
    encryptSymmetric(data, key) {
        const nonce = this.sodium.randombytes_buf(
            this.sodium.crypto_aead_chacha20poly1305_IETF_NPUBBYTES
        );

        const ciphertext = this.sodium.crypto_aead_chacha20poly1305_ietf_encrypt(
            data, null, null, nonce, key
        );

        // Combine nonce + ciphertext
        const combined = new Uint8Array(nonce.length + ciphertext.length);
        combined.set(nonce);
        combined.set(ciphertext, nonce.length);

        return combined;
    }

    // Decrypt data with ChaCha20-Poly1305
    decryptSymmetric(encryptedData, key) {
        const nonce = encryptedData.slice(0, 12);
        const ciphertext = encryptedData.slice(12);

        return this.sodium.crypto_aead_chacha20poly1305_ietf_decrypt(
            null, ciphertext, null, nonce, key
        );
    }

    // Decrypt sealed box challenge
    decryptSealedBox(sealedBox, publicKey, privateKey) {
        return this.sodium.crypto_box_seal_open(
            sealedBox, publicKey, privateKey
        );
    }

    // Generate verification ID from public key
    generateVerificationID(publicKey) {
        const hash = this.sodium.crypto_hash_sha256(publicKey);
        return generateMnemonic(wordlist, hash);
    }

    // Base64 encoding utilities
    toBase64Url(data) {
        return this.sodium.to_base64(
            data,
            this.sodium.base64_variants.URLSAFE_NO_PADDING
        );
    }

    fromBase64(data, variant = 'ORIGINAL') {
        const variantMap = {
            'ORIGINAL': this.sodium.base64_variants.ORIGINAL,
            'URLSAFE': this.sodium.base64_variants.URLSAFE_NO_PADDING
        };
        return this.sodium.from_base64(data, variantMap[variant]);
    }
}

export default new CryptoService();
```

## Library Installation Commands

### Go Backend
```bash
go get golang.org/x/crypto/chacha20poly1305
go get golang.org/x/crypto/nacl/box
go get golang.org/x/crypto/argon2
go get github.com/tyler-smith/go-bip39
```

### React.js Frontend
```bash
npm install libsodium-wrappers-sumo argon2-browser @scure/bip39
# or
yarn add libsodium-wrappers-sumo argon2-browser @scure/bip39
```

### React Native
```bash
npm install react-native-sodium
# iOS
cd ios && pod install
```

## Performance Considerations

1. **libsodium-wrappers-sumo** includes all algorithms but is ~1MB
2. For production, consider using **libsodium-wrappers** (smaller) with selective algorithm loading
3. **argon2-browser** uses WebAssembly for better performance
4. Consider using Web Workers for heavy crypto operations in browsers
