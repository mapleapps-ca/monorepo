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

  // Generate a random salt for key derivation
  generateSalt() {
    return sodium.randombytes_buf(32);
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
    // Using Argon2ID with recommended parameters
    const opsLimit = sodium.crypto_pwhash_OPSLIMIT_INTERACTIVE;
    const memLimit = sodium.crypto_pwhash_MEMLIMIT_INTERACTIVE;

    return sodium.crypto_pwhash(
      32, // key length
      password,
      salt,
      opsLimit,
      memLimit,
      sodium.crypto_pwhash_ALG_ARGON2ID,
    );
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

  // Generate all encryption fields needed for registration
  async generateRegistrationCrypto(password) {
    await this.ensureReady();

    // Generate random values
    const salt = this.generateSalt();
    const masterKey = this.generateMasterKey();
    const recoveryKey = this.generateRecoveryKey();
    const keyPair = this.generateKeyPair();

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
  }
}

export default CryptoService;
