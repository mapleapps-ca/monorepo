// src/services/CryptoService.js

import sodium from "libsodium-wrappers-sumo";

/**
 * CryptoService handles all cryptographic operations for the application
 * This includes key generation, encryption, and password-based key derivation
 */
export class CryptoService {
  constructor(logger) {
    this.logger = logger;
    this.isReady = false;
    this.initSodium();
  }

  /**
   * Initialize libsodium - this is async and must complete before any crypto operations
   */
  async initSodium() {
    try {
      await sodium.ready;
      this.isReady = true;
      this.logger.log("CryptoService: Libsodium initialized successfully");
    } catch (error) {
      this.logger.log(
        `CryptoService: Failed to initialize libsodium: ${error.message}`,
      );
      throw error;
    }
  }

  /**
   * Ensure libsodium is ready before performing crypto operations
   */
  ensureReady() {
    if (!this.isReady) {
      throw new Error("CryptoService not ready. Call initSodium() first.");
    }
  }

  /**
   * Generate a cryptographically secure random salt
   * Salt is used in password-based key derivation to prevent rainbow table attacks
   */
  generateSalt() {
    this.ensureReady();
    // Generate 32 bytes of random data for the salt
    const salt = sodium.randombytes_buf(32);
    this.logger.log("CryptoService: Generated new salt");
    return salt;
  }

  /**
   * Generate a master key for encrypting user data
   * This is the main encryption key that will encrypt all user files
   */
  generateMasterKey() {
    this.ensureReady();
    // Generate 32 bytes (256 bits) for ChaCha20-Poly1305
    const masterKey = sodium.randombytes_buf(32);
    this.logger.log("CryptoService: Generated new master key");
    return masterKey;
  }

  /**
   * Generate a recovery key that can be used to recover the master key
   * This provides a backup method to access encrypted data if password is lost
   */
  generateRecoveryKey() {
    this.ensureReady();
    const recoveryKey = sodium.randombytes_buf(32);
    this.logger.log("CryptoService: Generated new recovery key");
    return recoveryKey;
  }

  /**
   * Generate an X25519 key pair for asymmetric encryption
   * Public key can be shared, private key must be kept secret
   */
  generateKeyPair() {
    this.ensureReady();
    const keyPair = sodium.crypto_box_keypair();
    this.logger.log("CryptoService: Generated new X25519 key pair");
    return {
      publicKey: keyPair.publicKey,
      privateKey: keyPair.privateKey,
    };
  }

  /**
   * Derive an encryption key from a password using Argon2ID
   * This is computationally expensive to make brute force attacks difficult
   */
  async deriveKeyFromPassword(password, salt) {
    this.ensureReady();
    this.logger.log(
      "CryptoService: Starting password-based key derivation (this may take a moment)",
    );

    try {
      // Use Argon2ID for password hashing - it's memory-hard and resistant to attacks
      const derivedKey = sodium.crypto_pwhash(
        32, // Output length (32 bytes for ChaCha20-Poly1305)
        password, // The user's password
        salt, // The random salt
        sodium.crypto_pwhash_OPSLIMIT_INTERACTIVE, // CPU cost (interactive = fast enough for login)
        sodium.crypto_pwhash_MEMLIMIT_INTERACTIVE, // Memory cost
        sodium.crypto_pwhash_ALG_ARGON2ID, // Algorithm (Argon2ID is the most secure)
      );

      this.logger.log("CryptoService: Key derivation completed successfully");
      return derivedKey;
    } catch (error) {
      this.logger.log(`CryptoService: Key derivation failed: ${error.message}`);
      throw error;
    }
  }

  /**
   * Encrypt data using ChaCha20-Poly1305 authenticated encryption
   * This provides both confidentiality and authenticity
   */
  encrypt(plaintext, key) {
    this.ensureReady();

    // Convert string to Uint8Array if needed
    const plaintextBytes =
      typeof plaintext === "string" ? sodium.from_string(plaintext) : plaintext;

    // Generate a random nonce (number used once)
    const nonce = sodium.randombytes_buf(sodium.crypto_secretbox_NONCEBYTES);

    // Encrypt the data
    const ciphertext = sodium.crypto_secretbox_easy(plaintextBytes, nonce, key);

    // Concatenate nonce + ciphertext as required by the API
    const result = new Uint8Array(nonce.length + ciphertext.length);
    result.set(nonce, 0);
    result.set(ciphertext, nonce.length);

    this.logger.log("CryptoService: Data encrypted successfully");
    return result;
  }

  /**
   * Convert bytes to Base64 URL encoding as required by the API
   * Base64 URL is web-safe (no +, /, or = characters that could cause issues in URLs)
   */
  toBase64Url(bytes) {
    this.ensureReady();
    return sodium.to_base64(bytes, sodium.base64_variants.URLSAFE_NO_PADDING);
  }

  /**
   * Convert Base64 URL back to bytes
   */
  fromBase64Url(base64String) {
    this.ensureReady();
    return sodium.from_base64(
      base64String,
      sodium.base64_variants.URLSAFE_NO_PADDING,
    );
  }

  /**
   * Generate all the cryptographic materials needed for user registration
   * This is the main function that coordinates all the key generation and encryption
   */
  async generateRegistrationKeys(password) {
    this.ensureReady();
    this.logger.log(
      "CryptoService: Starting registration key generation process",
    );

    try {
      // Step 1: Generate all the raw cryptographic materials
      const salt = this.generateSalt();
      const masterKey = this.generateMasterKey();
      const recoveryKey = this.generateRecoveryKey();
      const keyPair = this.generateKeyPair();

      // Step 2: Derive the Key Encryption Key (KEK) from the user's password
      // This key will encrypt the master key so only the password can unlock it
      const keyEncryptionKey = await this.deriveKeyFromPassword(password, salt);

      // Step 3: Encrypt all the keys that need to be stored securely
      const encryptedMasterKey = this.encrypt(masterKey, keyEncryptionKey);
      const encryptedPrivateKey = this.encrypt(keyPair.privateKey, masterKey);
      const encryptedRecoveryKey = this.encrypt(recoveryKey, masterKey);
      const masterKeyEncryptedWithRecoveryKey = this.encrypt(
        masterKey,
        recoveryKey,
      );

      // Step 4: Encode everything in Base64 URL format for API transmission
      const result = {
        salt: this.toBase64Url(salt),
        publicKey: this.toBase64Url(keyPair.publicKey),
        encryptedMasterKey: this.toBase64Url(encryptedMasterKey),
        encryptedPrivateKey: this.toBase64Url(encryptedPrivateKey),
        encryptedRecoveryKey: this.toBase64Url(encryptedRecoveryKey),
        masterKeyEncryptedWithRecoveryKey: this.toBase64Url(
          masterKeyEncryptedWithRecoveryKey,
        ),
        // Generate a verification ID for the public key
        verificationID: this.toBase64Url(
          sodium.crypto_generichash(16, keyPair.publicKey),
        ),
      };

      this.logger.log(
        "CryptoService: Registration keys generated successfully",
      );
      return result;
    } catch (error) {
      this.logger.log(`CryptoService: Key generation failed: ${error.message}`);
      throw error;
    }
  }
}
