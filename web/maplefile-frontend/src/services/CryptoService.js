// web/maplefile-frontend/src/services/CryptoService.js
// src/services/CryptoService.js - LibSodium Implementation
class CryptoService {
  constructor() {
    this.sodium = null;
    this.isReady = false;
    console.log("Initializing CryptoService with libsodium-wrappers-sumo");
  }

  async init() {
    if (this.isReady && this.sodium) {
      return;
    }

    try {
      console.log("Loading libsodium-wrappers-sumo...");

      // Dynamic import with proper error handling
      const sodiumModule = await import("libsodium-wrappers-sumo");
      const sodium = sodiumModule.default || sodiumModule;

      console.log("LibSodium module loaded, waiting for ready...");
      await sodium.ready;

      this.sodium = sodium;
      this.isReady = true;

      console.log("✅ LibSodium initialized successfully");
      console.log("Available methods:", {
        randombytes_buf: typeof sodium.randombytes_buf,
        crypto_box_keypair: typeof sodium.crypto_box_keypair,
        crypto_pwhash: typeof sodium.crypto_pwhash,
        crypto_secretbox_easy: typeof sodium.crypto_secretbox_easy,
        crypto_box_easy: typeof sodium.crypto_box_easy,
        crypto_pwhash_ALG_ARGON2ID: typeof sodium.crypto_pwhash_ALG_ARGON2ID, // Add this line to log
      });
    } catch (error) {
      console.error("❌ Failed to initialize LibSodium:", error);
      throw new Error(`LibSodium initialization failed: ${error.message}`);
    }
  }

  async ensureReady() {
    if (!this.isReady) {
      await this.init();
    }

    if (!this.sodium) {
      throw new Error("LibSodium not available");
    }
  }

  // Generate random bytes using LibSodium
  generateRandomBytes(length) {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }
    return this.sodium.randombytes_buf(length);
  }

  // Generate salt (16 bytes for Argon2ID - MUST match backend Argon2SaltSize)
  generateSalt() {
    return this.generateRandomBytes(16); // Changed from 32 to 16 to match backend
  }

  // Generate master key (32 bytes for ChaCha20-Poly1305)
  generateMasterKey() {
    return this.generateRandomBytes(32);
  }

  // Generate recovery key
  generateRecoveryKey() {
    return this.generateRandomBytes(32);
  }

  // Generate X25519 keypair using LibSodium
  generateKeyPair() {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }

    console.log("Generating X25519 keypair...");
    const keyPair = this.sodium.crypto_box_keypair();

    return {
      publicKey: keyPair.publicKey,
      privateKey: keyPair.privateKey,
    };
  }

  // Argon2ID key derivation (LibSodium native) - Use same parameters as backend
  async deriveKeyFromPassword(password, salt) {
    await this.ensureReady();

    console.log(
      "deriveKeyFromPassword - Checking if Argon2ID algorithm is available:",
      this.sodium && this.sodium.crypto_pwhash_ALG_ARGON2ID13 !== undefined,
    );
    if (this.sodium && !this.sodium.crypto_pwhash_ALG_ARGON2ID13) {
      console.error("Argon2ID13 algorithm not supported by libsodium");
      throw new Error("Argon2ID13 algorithm not supported by libsodium");
    }

    // Validate salt length matches backend expectation
    if (salt.length !== 16) {
      throw new Error(`Invalid salt length: expected 16, got ${salt.length}`);
    }

    console.log("Deriving key from password using Argon2ID13...");

    const passwordBytes =
      typeof password === "string"
        ? this.sodium.from_string(password)
        : password;

    // Use Argon2ID with parameters matching the backend constants
    // Backend: Argon2MemLimit = 67108864 (64 MB), Argon2OpsLimit = 4
    const derivedKey = this.sodium.crypto_pwhash(
      32, // output length (32 bytes for ChaCha20-Poly1305)
      passwordBytes,
      salt,
      4, // Argon2OpsLimit = 4 (matches backend)
      67108864, // Argon2MemLimit = 64 MB (matches backend)
      this.sodium.crypto_pwhash_ALG_ARGON2ID13,
    );

    return derivedKey;
  }

  // ChaCha20-Poly1305 encryption (LibSodium native)
  encrypt(message, key) {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }

    const nonce = this.generateRandomBytes(
      this.sodium.crypto_secretbox_NONCEBYTES,
    );
    const ciphertext = this.sodium.crypto_secretbox_easy(message, nonce, key);

    // Combine nonce + ciphertext (as expected by backend)
    const result = new Uint8Array(nonce.length + ciphertext.length);
    result.set(nonce);
    result.set(ciphertext, nonce.length);

    return result;
  }

  // ChaCha20-Poly1305 decryption (LibSodium native)
  decrypt(encryptedData, key) {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }

    const nonceLength = this.sodium.crypto_secretbox_NONCEBYTES;
    const nonce = encryptedData.slice(0, nonceLength);
    const ciphertext = encryptedData.slice(nonceLength);

    const decrypted = this.sodium.crypto_secretbox_open_easy(
      ciphertext,
      nonce,
      key,
    );
    return decrypted;
  }

  // Base64 URL encode (LibSodium native)
  base64UrlEncode(data) {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }
    return this.sodium.to_base64(
      data,
      this.sodium.base64_variants.URLSAFE_NO_PADDING,
    );
  }

  // Base64 URL decode (LibSodium native)
  base64UrlDecode(data) {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }
    return this.sodium.from_base64(
      data,
      this.sodium.base64_variants.URLSAFE_NO_PADDING,
    );
  }

  // Generate verification ID that matches backend expectations
  generateVerificationID(publicKey) {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }

    console.log("Generating verification ID from public key...");

    // Hash the public key using Blake2b (or SHA-256 if Blake2b not available)
    let hash;
    if (this.sodium.crypto_generichash) {
      // Use Blake2b (preferred)
      hash = this.sodium.crypto_generichash(16, publicKey);
    } else {
      // Fallback to SHA-256
      const fullHash = this.sodium.crypto_hash_sha256(publicKey);
      hash = fullHash.slice(0, 16);
    }

    return this.base64UrlEncode(hash);
  }

  // Generate registration crypto data (matching backend expectations)
  async generateRegistrationCrypto(password) {
    try {
      console.log("Generating registration crypto with LibSodium...");
      await this.ensureReady();

      // Generate all the required keys and data
      const salt = this.generateSalt(); // Now generates 16 bytes
      const masterKey = this.generateMasterKey();
      const recoveryKey = this.generateRecoveryKey();
      const keyPair = this.generateKeyPair();

      console.log("Generated keys:", {
        saltLength: salt.length,
        masterKeyLength: masterKey.length,
        recoveryKeyLength: recoveryKey.length,
        publicKeyLength: keyPair.publicKey.length,
        privateKeyLength: keyPair.privateKey.length,
      });

      // Derive key encryption key using Argon2ID
      const kek = await this.deriveKeyFromPassword(password, salt);
      console.log("Key encryption key derived, length:", kek.length);

      // Encrypt keys using ChaCha20-Poly1305
      const encryptedMasterKey = this.encrypt(masterKey, kek);
      const encryptedPrivateKey = this.encrypt(keyPair.privateKey, masterKey);
      const encryptedRecoveryKey = this.encrypt(recoveryKey, masterKey);
      const masterKeyEncryptedWithRecoveryKey = this.encrypt(
        masterKey,
        recoveryKey,
      );

      console.log("Keys encrypted successfully");

      // Generate verification ID that matches backend expectations
      const verificationID = this.generateVerificationID(keyPair.publicKey);
      console.log("Verification ID generated:", verificationID);

      // Return data in the format expected by the backend
      const result = {
        salt: this.base64UrlEncode(salt),
        publicKey: this.base64UrlEncode(keyPair.publicKey),
        encryptedMasterKey: this.base64UrlEncode(encryptedMasterKey),
        encryptedPrivateKey: this.base64UrlEncode(encryptedPrivateKey),
        encryptedRecoveryKey: this.base64UrlEncode(encryptedRecoveryKey),
        masterKeyEncryptedWithRecoveryKey: this.base64UrlEncode(
          masterKeyEncryptedWithRecoveryKey,
        ),
        verificationID: verificationID,
        // Store raw keys for session use (these won't be sent to server)
        _masterKey: this.base64UrlEncode(masterKey),
        _recoveryKey: this.base64UrlEncode(recoveryKey),
        _privateKey: this.base64UrlEncode(keyPair.privateKey),
      };

      console.log("Registration crypto data generated successfully");
      return result;
    } catch (error) {
      console.error("❌ LibSodium registration crypto error:", error);
      throw new Error(`Failed to generate encryption data: ${error.message}`);
    }
  }

  // Derive key with parameters (for login compatibility)
  async deriveKeyWithParams(password, salt, kdfParams) {
    await this.ensureReady();

    // For now, use the same derivation as registration
    // In the future, this could be extended to handle different KDF parameters
    return await this.deriveKeyFromPassword(password, salt);
  }

  // Process login step 2 (decrypt user data)
  async processLoginStep2(password, loginData) {
    try {
      await this.ensureReady();

      const salt = this.base64UrlDecode(loginData.salt);
      const encryptedMasterKey = this.base64UrlDecode(
        loginData.encryptedMasterKey,
      );
      const encryptedPrivateKey = this.base64UrlDecode(
        loginData.encryptedPrivateKey,
      );
      const publicKey = this.base64UrlDecode(loginData.publicKey);

      // Derive key encryption key
      const kek = await this.deriveKeyWithParams(
        password,
        salt,
        loginData.kdf_params,
      );

      // Decrypt master key
      const masterKey = this.decrypt(encryptedMasterKey, kek);

      // Decrypt private key
      const privateKey = this.decrypt(encryptedPrivateKey, masterKey);

      // Handle encrypted challenge if present
      let decryptedChallenge;
      if (loginData.encryptedChallenge) {
        const encryptedChallengeBytes = this.base64UrlDecode(
          loginData.encryptedChallenge,
        );
        // For challenge decryption, we might need to use box_open with ephemeral key
        // This is a simplified version - full implementation would need the ephemeral public key
        decryptedChallenge = this.generateRandomBytes(32); // Placeholder
      } else {
        decryptedChallenge = this.generateRandomBytes(32);
      }

      return {
        masterKey: this.base64UrlEncode(masterKey),
        privateKey: this.base64UrlEncode(privateKey),
        publicKey: this.base64UrlEncode(publicKey),
        decryptedChallenge: this.base64UrlEncode(decryptedChallenge),
        challengeId: loginData.challengeId,
      };
    } catch (error) {
      console.error("Login step 2 error:", error);
      throw new Error(
        "Failed to decrypt login data. Please check your password.",
      );
    }
  }

  // Token decryption using box_open
  async decryptTokens(encryptedTokens, tokenNonce, privateKey) {
    try {
      await this.ensureReady();

      const encryptedData = this.base64UrlDecode(encryptedTokens);
      const nonce = this.base64UrlDecode(tokenNonce);
      const privateKeyBytes = this.base64UrlDecode(privateKey);

      // For token decryption, we would need the server's public key
      // This is a simplified version
      const key = privateKeyBytes.slice(0, 32); // Use first 32 bytes as key
      const decryptedData = this.decrypt(encryptedData, key);

      return JSON.parse(this.sodium.to_string(decryptedData));
    } catch (error) {
      console.error("Token decryption error:", error);
      throw new Error("Failed to decrypt authentication tokens");
    }
  }

  // Utility method to convert string to bytes
  stringToBytes(str) {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }
    return this.sodium.from_string(str);
  }

  // Utility method to convert bytes to string
  bytesToString(bytes) {
    if (!this.sodium) {
      throw new Error("LibSodium not initialized");
    }
    return this.sodium.to_string(bytes);
  }
}

export default CryptoService;
