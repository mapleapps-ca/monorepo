// src/services/CryptoService.js
import sodium from "libsodium-wrappers-sumo";

class CryptoService {
  constructor() {
    this.isReady = false;
    this.initPromise = this.init();
  }

  // Initialize libsodium
  async init() {
    if (this.isReady) return;
    await sodium.ready;
    this.isReady = true;
  }

  // Ensure libsodium is ready before any crypto operations
  async ensureReady() {
    await this.initPromise;
  }

  // Generate a random salt for key derivation (Argon2ID requires specific length)
  generateSalt() {
    return sodium.randombytes_buf(sodium.crypto_pwhash_SALTBYTES);
  }

  // Generate a keypair for encryption
  generateKeyPair() {
    return sodium.crypto_box_keypair();
  }

  // Generate a random master key
  generateMasterKey() {
    return sodium.randombytes_buf(32);
  }

  // Generate a random recovery key
  generateRecoveryKey() {
    return sodium.randombytes_buf(32);
  }

  // Derive key from password using Argon2ID
  async deriveKeyFromPassword(password, salt) {
    await this.ensureReady();

    // Convert password to Uint8Array if it's a string
    const passwordBytes =
      typeof password === "string"
        ? new TextEncoder().encode(password)
        : password;

    // Use recommended Argon2ID parameters
    const opsLimit = sodium.crypto_pwhash_OPSLIMIT_INTERACTIVE;
    const memLimit = sodium.crypto_pwhash_MEMLIMIT_INTERACTIVE;

    try {
      return sodium.crypto_pwhash(
        32, // key length
        passwordBytes,
        salt,
        opsLimit,
        memLimit,
        sodium.crypto_pwhash_ALG_ARGON2ID,
      );
    } catch (error) {
      console.error("Key derivation error:", error);
      throw new Error("Failed to derive encryption key from password");
    }
  }

  // Encrypt data using ChaCha20-Poly1305
  encrypt(message, key) {
    const nonce = sodium.randombytes_buf(
      sodium.crypto_aead_chacha20poly1305_NPUBBYTES,
    );
    const ciphertext = sodium.crypto_aead_chacha20poly1305_encrypt(
      message,
      null, // additional data
      null, // nsec (not used)
      nonce,
      key,
    );

    // Return nonce + ciphertext concatenated
    const result = new Uint8Array(nonce.length + ciphertext.length);
    result.set(nonce);
    result.set(ciphertext, nonce.length);

    return result;
  }

  // Decrypt data using ChaCha20-Poly1305
  decrypt(encryptedData, key) {
    const nonceLength = sodium.crypto_aead_chacha20poly1305_NPUBBYTES;
    const nonce = encryptedData.slice(0, nonceLength);
    const ciphertext = encryptedData.slice(nonceLength);

    return sodium.crypto_aead_chacha20poly1305_decrypt(
      null, // nsec (not used)
      ciphertext,
      null, // additional data
      nonce,
      key,
    );
  }

  // Base64 URL encode (as required by API)
  base64UrlEncode(data) {
    return sodium.to_base64(data, sodium.base64_variants.URLSAFE_NO_PADDING);
  }

  // Base64 URL decode
  base64UrlDecode(data) {
    return sodium.from_base64(data, sodium.base64_variants.URLSAFE_NO_PADDING);
  }

  // Decrypt challenge using X25519 box (used in login step 3)
  decryptWithPrivateKey(encryptedData, privateKey, publicKey) {
    const nonceLength = sodium.crypto_box_NONCEBYTES;
    const nonce = encryptedData.slice(0, nonceLength);
    const ciphertext = encryptedData.slice(nonceLength);

    return sodium.crypto_box_open_easy(
      ciphertext,
      nonce,
      publicKey,
      privateKey,
    );
  }

  // Derive key from password with specific KDF parameters
  async deriveKeyWithParams(password, salt, kdfParams) {
    await this.ensureReady();

    // Convert password to Uint8Array if it's a string
    const passwordBytes =
      typeof password === "string"
        ? new TextEncoder().encode(password)
        : password;

    // Use provided KDF parameters or defaults
    const opsLimit =
      kdfParams?.iterations || sodium.crypto_pwhash_OPSLIMIT_INTERACTIVE;
    const memLimit =
      kdfParams?.memory || sodium.crypto_pwhash_MEMLIMIT_INTERACTIVE;
    const keyLength = kdfParams?.key_length || 32;

    try {
      return sodium.crypto_pwhash(
        keyLength,
        passwordBytes,
        salt,
        opsLimit,
        memLimit,
        sodium.crypto_pwhash_ALG_ARGON2ID,
      );
    } catch (error) {
      console.error("Key derivation with params error:", error);
      throw new Error(
        "Failed to derive encryption key with provided parameters",
      );
    }
  }

  // Process login step 2 response - decrypt keys and challenge
  async processLoginStep2(password, loginData) {
    await this.ensureReady();

    try {
      // Decode base64 data
      const salt = this.base64UrlDecode(loginData.salt);
      const encryptedMasterKey = this.base64UrlDecode(
        loginData.encryptedMasterKey,
      );
      const encryptedPrivateKey = this.base64UrlDecode(
        loginData.encryptedPrivateKey,
      );
      const encryptedChallenge = this.base64UrlDecode(
        loginData.encryptedChallenge,
      );
      const publicKey = this.base64UrlDecode(loginData.publicKey);

      // Derive key encryption key from password using provided KDF params
      const kek = await this.deriveKeyWithParams(
        password,
        salt,
        loginData.kdf_params,
      );

      // Decrypt master key
      const masterKey = this.decrypt(encryptedMasterKey, kek);

      // Decrypt private key using master key
      const privateKey = this.decrypt(encryptedPrivateKey, masterKey);

      // Decrypt challenge using private key and server's public key
      // Note: For X25519 box, we need to handle the nonce differently
      const decryptedChallenge = this.decryptWithPrivateKey(
        encryptedChallenge,
        privateKey,
        publicKey, // This should be the server's public key for the challenge
      );

      return {
        masterKey: this.base64UrlEncode(masterKey),
        privateKey: this.base64UrlEncode(privateKey),
        publicKey: this.base64UrlEncode(publicKey),
        decryptedChallenge: this.base64UrlEncode(decryptedChallenge),
        challengeId: loginData.challengeId,
      };
    } catch (error) {
      console.error("Decryption error:", error);
      throw new Error(
        "Failed to decrypt login data. Please check your password.",
      );
    }
  }

  // Decrypt tokens received from login step 3
  async decryptTokens(encryptedTokens, tokenNonce, privateKey) {
    await this.ensureReady();

    try {
      const encryptedData = this.base64UrlDecode(encryptedTokens);
      const nonce = this.base64UrlDecode(tokenNonce);
      const privateKeyBytes = this.base64UrlDecode(privateKey);

      // For token decryption, we use ChaCha20-Poly1305
      const decryptedData = sodium.crypto_aead_chacha20poly1305_decrypt(
        null, // nsec (not used)
        encryptedData,
        null, // additional data
        nonce,
        privateKeyBytes,
      );

      // Parse the decrypted JSON
      const tokenData = JSON.parse(new TextDecoder().decode(decryptedData));
      return tokenData;
    } catch (error) {
      console.error("Token decryption error:", error);
      throw new Error("Failed to decrypt authentication tokens");
    }
  }

  // Generate all encryption fields needed for registration
  async generateRegistrationCrypto(password) {
    await this.ensureReady();

    try {
      // Generate random values
      const salt = this.generateSalt();
      const masterKey = this.generateMasterKey();
      const recoveryKey = this.generateRecoveryKey();
      const keyPair = this.generateKeyPair();

      console.log("Generated salt length:", salt.length);
      console.log("Expected salt length:", sodium.crypto_pwhash_SALTBYTES);

      // Derive key encryption key from password
      const kek = await this.deriveKeyFromPassword(password, salt);

      // Encrypt master key with KEK
      const encryptedMasterKey = this.encrypt(masterKey, kek);

      // Encrypt private key with master key
      const encryptedPrivateKey = this.encrypt(keyPair.privateKey, masterKey);

      // Encrypt recovery key with master key
      const encryptedRecoveryKey = this.encrypt(recoveryKey, masterKey);

      // Encrypt master key with recovery key
      const masterKeyEncryptedWithRecoveryKey = this.encrypt(
        masterKey,
        recoveryKey,
      );

      // Generate verification ID (can be derived from public key)
      const verificationID = this.base64UrlEncode(
        sodium.crypto_generichash(16, keyPair.publicKey),
      );

      return {
        salt: this.base64UrlEncode(salt),
        publicKey: this.base64UrlEncode(keyPair.publicKey),
        encryptedMasterKey: this.base64UrlEncode(encryptedMasterKey),
        encryptedPrivateKey: this.base64UrlEncode(encryptedPrivateKey),
        encryptedRecoveryKey: this.base64UrlEncode(encryptedRecoveryKey),
        masterKeyEncryptedWithRecoveryKey: this.base64UrlEncode(
          masterKeyEncryptedWithRecoveryKey,
        ),
        verificationID,
        // Keep these for potential future use (not sent to API)
        _masterKey: this.base64UrlEncode(masterKey),
        _recoveryKey: this.base64UrlEncode(recoveryKey),
        _privateKey: this.base64UrlEncode(keyPair.privateKey),
      };
    } catch (error) {
      console.error("Registration crypto generation error:", error);
      throw new Error(`Failed to generate encryption data: ${error.message}`);
    }
  }
}

export default CryptoService;
