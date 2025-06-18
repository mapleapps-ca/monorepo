// src/services/CryptoService.js - Web Crypto API Fallback
class CryptoService {
  constructor() {
    this.isReady = true; // Web Crypto is always ready in modern browsers
    console.log("Using Web Crypto API fallback");
  }

  async init() {
    // Web Crypto doesn't need initialization
    return Promise.resolve();
  }

  async ensureReady() {
    // Always ready
    return Promise.resolve();
  }

  // Generate random bytes using Web Crypto
  generateRandomBytes(length) {
    const array = new Uint8Array(length);
    crypto.getRandomValues(array);
    return array;
  }

  // Generate salt (32 bytes for Web Crypto)
  generateSalt() {
    return this.generateRandomBytes(32);
  }

  // Generate master key
  generateMasterKey() {
    return this.generateRandomBytes(32);
  }

  // Generate recovery key
  generateRecoveryKey() {
    return this.generateRandomBytes(32);
  }

  // Generate keypair using Web Crypto
  async generateKeyPair() {
    const keyPair = await crypto.subtle.generateKey(
      {
        name: "ECDH",
        namedCurve: "P-256",
      },
      true,
      ["deriveKey"],
    );

    const publicKey = await crypto.subtle.exportKey("raw", keyPair.publicKey);
    const privateKey = await crypto.subtle.exportKey(
      "pkcs8",
      keyPair.privateKey,
    );

    return {
      publicKey: new Uint8Array(publicKey),
      privateKey: new Uint8Array(privateKey),
    };
  }

  // PBKDF2 key derivation (Web Crypto native)
  async deriveKeyFromPassword(password, salt) {
    const passwordBytes =
      typeof password === "string"
        ? new TextEncoder().encode(password)
        : password;

    const keyMaterial = await crypto.subtle.importKey(
      "raw",
      passwordBytes,
      { name: "PBKDF2" },
      false,
      ["deriveKey"],
    );

    const key = await crypto.subtle.deriveKey(
      {
        name: "PBKDF2",
        salt: salt,
        iterations: 100000, // 100k iterations
        hash: "SHA-256",
      },
      keyMaterial,
      { name: "AES-GCM", length: 256 },
      true,
      ["encrypt", "decrypt"],
    );

    return new Uint8Array(await crypto.subtle.exportKey("raw", key));
  }

  // AES-GCM encryption (Web Crypto native)
  async encrypt(message, key) {
    const iv = this.generateRandomBytes(12); // 96-bit IV for GCM

    const cryptoKey = await crypto.subtle.importKey(
      "raw",
      key,
      { name: "AES-GCM" },
      false,
      ["encrypt"],
    );

    const encrypted = await crypto.subtle.encrypt(
      {
        name: "AES-GCM",
        iv: iv,
      },
      cryptoKey,
      message,
    );

    // Combine IV + ciphertext
    const result = new Uint8Array(iv.length + encrypted.byteLength);
    result.set(iv);
    result.set(new Uint8Array(encrypted), iv.length);

    return result;
  }

  // AES-GCM decryption (Web Crypto native)
  async decrypt(encryptedData, key) {
    const iv = encryptedData.slice(0, 12);
    const ciphertext = encryptedData.slice(12);

    const cryptoKey = await crypto.subtle.importKey(
      "raw",
      key,
      { name: "AES-GCM" },
      false,
      ["decrypt"],
    );

    const decrypted = await crypto.subtle.decrypt(
      {
        name: "AES-GCM",
        iv: iv,
      },
      cryptoKey,
      ciphertext,
    );

    return new Uint8Array(decrypted);
  }

  // Base64 URL encode
  base64UrlEncode(data) {
    const base64 = btoa(String.fromCharCode(...data));
    return base64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
  }

  // Base64 URL decode
  base64UrlDecode(data) {
    const base64 = data.replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64 + "===".slice((base64.length + 3) % 4);
    const binary = atob(padded);
    return new Uint8Array([...binary].map((char) => char.charCodeAt(0)));
  }

  // Generate registration crypto data
  async generateRegistrationCrypto(password) {
    try {
      console.log("Generating registration crypto with Web Crypto API...");

      const salt = this.generateSalt();
      const masterKey = this.generateMasterKey();
      const recoveryKey = this.generateRecoveryKey();
      const keyPair = await this.generateKeyPair();

      // Derive key encryption key
      const kek = await this.deriveKeyFromPassword(password, salt);

      // Encrypt keys
      const encryptedMasterKey = await this.encrypt(masterKey, kek);
      const encryptedPrivateKey = await this.encrypt(
        keyPair.privateKey,
        masterKey,
      );
      const encryptedRecoveryKey = await this.encrypt(recoveryKey, masterKey);
      const masterKeyEncryptedWithRecoveryKey = await this.encrypt(
        masterKey,
        recoveryKey,
      );

      // Generate verification ID
      const publicKeyHash = await crypto.subtle.digest(
        "SHA-256",
        keyPair.publicKey,
      );
      const verificationID = this.base64UrlEncode(
        new Uint8Array(publicKeyHash.slice(0, 16)),
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
        _masterKey: this.base64UrlEncode(masterKey),
        _recoveryKey: this.base64UrlEncode(recoveryKey),
        _privateKey: this.base64UrlEncode(keyPair.privateKey),
      };
    } catch (error) {
      console.error("Web Crypto registration error:", error);
      throw new Error(`Failed to generate encryption data: ${error.message}`);
    }
  }

  // Derive key with parameters (for login)
  async deriveKeyWithParams(password, salt, kdfParams) {
    // Use same method as registration
    return await this.deriveKeyFromPassword(password, salt);
  }

  // Process login step 2
  async processLoginStep2(password, loginData) {
    try {
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
      const masterKey = await this.decrypt(encryptedMasterKey, kek);

      // Decrypt private key
      const privateKey = await this.decrypt(encryptedPrivateKey, masterKey);

      // For challenge decryption, we'll need a different approach
      // This is a simplified version - real implementation would need ECDH
      const decryptedChallenge = this.generateRandomBytes(32); // Placeholder

      return {
        masterKey: this.base64UrlEncode(masterKey),
        privateKey: this.base64UrlEncode(privateKey),
        publicKey: this.base64UrlEncode(publicKey),
        decryptedChallenge: this.base64UrlEncode(decryptedChallenge),
        challengeId: loginData.challengeId,
      };
    } catch (error) {
      throw new Error(
        "Failed to decrypt login data. Please check your password.",
      );
    }
  }

  // Token decryption (simplified)
  async decryptTokens(encryptedTokens, tokenNonce, privateKey) {
    try {
      const encryptedData = this.base64UrlDecode(encryptedTokens);
      const key = this.base64UrlDecode(privateKey).slice(0, 32); // Use first 32 bytes as key

      const decryptedData = await this.decrypt(encryptedData, key);
      return JSON.parse(new TextDecoder().decode(decryptedData));
    } catch (error) {
      throw new Error("Failed to decrypt authentication tokens");
    }
  }
}

export default CryptoService;
