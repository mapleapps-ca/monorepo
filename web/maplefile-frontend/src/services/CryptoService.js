// Production Crypto Service for E2EE operations
import sodium from "libsodium-wrappers-sumo";
import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";

class CryptoService {
  constructor() {
    this.isInitialized = false;
    this.sodium = null;
    console.log("Initializing CryptoService with libsodium-wrappers-sumo");
  }

  // Initialize libsodium
  async initialize() {
    if (this.isInitialized) return;

    await sodium.ready;
    this.sodium = sodium;
    this.isInitialized = true;
    console.log("[CryptoService] Libsodium initialized");
  }

  // Legacy compatibility method
  async init() {
    return this.initialize();
  }

  // Derive key from password using PBKDF2
  async deriveKeyFromPassword(password, salt) {
    console.log(
      `[CryptoService] PBKDF2 - password length: ${password.length}, salt length: ${salt.length}`,
    );
    this.hexDump(salt, "PBKDF2 Salt");

    const encoder = new TextEncoder();
    const passwordBuffer = encoder.encode(password);

    const cryptoKey = await crypto.subtle.importKey(
      "raw",
      passwordBuffer,
      { name: "PBKDF2" },
      false,
      ["deriveKey"],
    );

    const derivedKey = await crypto.subtle.deriveKey(
      {
        name: "PBKDF2",
        salt: salt,
        iterations: 100000,
        hash: "SHA-256",
      },
      cryptoKey,
      { name: "AES-GCM", length: 256 },
      true,
      ["encrypt", "decrypt"],
    );

    const keyBuffer = await crypto.subtle.exportKey("raw", derivedKey);
    const result = new Uint8Array(keyBuffer);

    console.log(`[CryptoService] PBKDF2 derived key length: ${result.length}`);
    this.hexDump(result, "Derived Key");

    return result;
  }

  // Decrypt data using ChaCha20-Poly1305 (libsodium secretbox)
  decryptWithSecretBox(encryptedData, key) {
    if (!this.isInitialized) {
      throw new Error("CryptoService not initialized");
    }

    try {
      console.log(
        `[CryptoService] SecretBox decrypt - encrypted data length: ${encryptedData.length}, key length: ${key.length}`,
      );

      // Validate key length
      if (key.length !== this.sodium.crypto_secretbox_KEYBYTES) {
        throw new Error(
          `Invalid key length: ${key.length}, expected: ${this.sodium.crypto_secretbox_KEYBYTES}`,
        );
      }

      // encryptedData format: nonce (24 bytes) + ciphertext + mac (16 bytes)
      const nonceLength = this.sodium.crypto_secretbox_NONCEBYTES; // 24 bytes
      const macLength = this.sodium.crypto_secretbox_MACBYTES; // 16 bytes

      if (encryptedData.length <= nonceLength + macLength) {
        throw new Error(
          `Invalid encrypted data length: ${encryptedData.length}, minimum required: ${nonceLength + macLength + 1}`,
        );
      }

      // Extract nonce and ciphertext
      const nonce = encryptedData.slice(0, nonceLength);
      const ciphertext = encryptedData.slice(nonceLength);

      console.log(
        `[CryptoService] Extracted nonce length: ${nonce.length}, ciphertext length: ${ciphertext.length}`,
      );
      this.hexDump(nonce, "Nonce");
      this.hexDump(key, "Key");
      this.hexDump(ciphertext, "Ciphertext", 48);

      // Decrypt using libsodium
      const decrypted = this.sodium.crypto_secretbox_open_easy(
        ciphertext,
        nonce,
        key,
      );

      console.log(
        `[CryptoService] SecretBox decryption successful, result length: ${decrypted.length}`,
      );
      return decrypted;
    } catch (error) {
      console.error("[CryptoService] SecretBox decryption failed:", error);
      console.error(
        "[CryptoService] Encrypted data (first 50 bytes):",
        encryptedData.slice(0, 50),
      );
      console.error("[CryptoService] Key (first 10 bytes):", key.slice(0, 10));
      throw new Error(`SecretBox decryption failed: ${error.message}`);
    }
  }

  // Decrypt challenge using anonymous box (matching backend)
  async decryptChallenge(
    encryptedChallenge,
    privateKey,
    storedPublicKey = null,
  ) {
    if (!this.isInitialized) {
      throw new Error("CryptoService not initialized");
    }

    try {
      console.log(
        `[CryptoService] Challenge decrypt - encrypted length: ${encryptedChallenge.length}, private key length: ${privateKey.length}`,
      );

      // Validate private key length
      if (privateKey.length !== this.sodium.crypto_box_SECRETKEYBYTES) {
        throw new Error(
          `Invalid private key length: ${privateKey.length}, expected: ${this.sodium.crypto_box_SECRETKEYBYTES}`,
        );
      }

      // For anonymous box (SealAnonymous/OpenAnonymous), the format is different
      // It should just be: ephemeral_public_key + ciphertext (no separate nonce)
      // The total overhead is 48 bytes (32 bytes ephemeral key + 16 bytes MAC)

      const expectedMinLength = this.sodium.crypto_box_SEALBYTES; // Should be 48
      if (encryptedChallenge.length < expectedMinLength) {
        throw new Error(
          `Invalid encrypted challenge length: ${encryptedChallenge.length}, minimum required: ${expectedMinLength}`,
        );
      }

      console.log(
        "[CryptoService] Attempting anonymous box decryption (box.seal format)",
      );
      this.hexDump(encryptedChallenge, "Encrypted Challenge", 80);
      this.hexDump(privateKey, "Private Key");

      try {
        // Try anonymous box decryption first (matches backend's SealAnonymous)
        // For sealed box, we need to derive the public key from the private key
        // Using the correct libsodium method
        const derivedPublicKey = this.sodium.crypto_scalarmult_base(privateKey);

        console.log("[CryptoService] Derived public key for sealed box");
        this.hexDump(derivedPublicKey, "Derived Public Key");

        // Try with derived public key first
        let decrypted;
        try {
          decrypted = this.sodium.crypto_box_seal_open(
            encryptedChallenge,
            derivedPublicKey,
            privateKey,
          );
          console.log(
            `[CryptoService] Anonymous box decryption successful with derived key, result length: ${decrypted.length}`,
          );
          return decrypted;
        } catch (derivedKeyError) {
          console.log(
            "[CryptoService] Failed with derived public key:",
            derivedKeyError.message,
          );

          // If we have a stored public key, try that too
          if (storedPublicKey) {
            try {
              decrypted = this.sodium.crypto_box_seal_open(
                encryptedChallenge,
                storedPublicKey,
                privateKey,
              );
              console.log(
                `[CryptoService] Anonymous box decryption successful with stored key, result length: ${decrypted.length}`,
              );
              return decrypted;
            } catch (storedKeyError) {
              console.log(
                "[CryptoService] Failed with stored public key:",
                storedKeyError.message,
              );
              throw derivedKeyError; // Throw the original error
            }
          } else {
            throw derivedKeyError;
          }
        }
      } catch (anonymousError) {
        console.log(
          "[CryptoService] Anonymous box failed:",
          anonymousError.message,
        );

        // Let's also try with the stored public key from verification data
        // The public key might be available in the verification response
        console.log(
          "[CryptoService] Trying with stored public key if available...",
        );

        try {
          // We can get the public key from the auth service or pass it as a parameter
          // For now, let's try regular box format
          throw new Error("Switching to regular box format");
        } catch (storedKeyError) {
          console.log("[CryptoService] Trying regular box format...");

          // Fallback: try regular box format
          // Format: ephemeral_public_key (32 bytes) + nonce (24 bytes) + ciphertext + mac
          const ephemeralPubKeyLength = this.sodium.crypto_box_PUBLICKEYBYTES; // 32 bytes
          const nonceLength = this.sodium.crypto_box_NONCEBYTES; // 24 bytes
          const macLength = this.sodium.crypto_box_MACBYTES; // 16 bytes

          const minLength = ephemeralPubKeyLength + nonceLength + macLength + 1;
          if (encryptedChallenge.length < minLength) {
            throw new Error(
              `Invalid encrypted challenge length for regular box: ${encryptedChallenge.length}, minimum required: ${minLength}`,
            );
          }

          const ephemeralPublicKey = encryptedChallenge.slice(
            0,
            ephemeralPubKeyLength,
          );
          const nonce = encryptedChallenge.slice(
            ephemeralPubKeyLength,
            ephemeralPubKeyLength + nonceLength,
          );
          const ciphertext = encryptedChallenge.slice(
            ephemeralPubKeyLength + nonceLength,
          );

          console.log("[CryptoService] Regular box challenge components:");
          console.log(
            "Ephemeral public key length:",
            ephemeralPublicKey.length,
          );
          console.log("Nonce length:", nonce.length);
          console.log("Ciphertext length:", ciphertext.length);
          this.hexDump(ephemeralPublicKey, "Ephemeral Public Key");

          // Validate component lengths
          if (ephemeralPublicKey.length !== ephemeralPubKeyLength) {
            throw new Error(
              `Invalid ephemeral public key length: ${ephemeralPublicKey.length}`,
            );
          }
          if (nonce.length !== nonceLength) {
            throw new Error(`Invalid nonce length: ${nonce.length}`);
          }
          if (ciphertext.length < macLength + 1) {
            throw new Error(
              `Ciphertext too short: ${ciphertext.length}, minimum: ${macLength + 1}`,
            );
          }

          console.log(
            "[CryptoService] Decrypting challenge with regular X25519 + ChaCha20-Poly1305",
          );

          // Decrypt using regular box
          const decrypted = this.sodium.crypto_box_open_easy(
            ciphertext,
            nonce,
            ephemeralPublicKey,
            privateKey,
          );

          console.log(
            `[CryptoService] Regular box decryption successful, result length: ${decrypted.length}`,
          );
          return decrypted;
        }
      }
    } catch (error) {
      console.error("[CryptoService] Challenge decryption failed:", error);
      console.error(
        "[CryptoService] Encrypted challenge (first 50 bytes):",
        encryptedChallenge.slice(0, 50),
      );
      console.error(
        "[CryptoService] Private key (first 10 bytes):",
        privateKey.slice(0, 10),
      );
      throw new Error(`Challenge decryption failed: ${error.message}`);
    }
  }

  // Utility method for hex dump debugging
  hexDump(data, label = "Data", maxBytes = 32) {
    const bytes = data.slice(0, maxBytes);
    const hex = Array.from(bytes)
      .map((b) => b.toString(16).padStart(2, "0"))
      .join(" ");
    console.log(
      `[CryptoService] ${label} (${data.length} bytes): ${hex}${data.length > maxBytes ? "..." : ""}`,
    );
  }

  // Convert base64 to Uint8Array
  base64ToUint8Array(
    base64String,
    variant = this.sodium?.base64_variants?.URLSAFE_NO_PADDING,
  ) {
    if (!this.sodium) {
      throw new Error("CryptoService not initialized");
    }
    return this.sodium.from_base64(base64String, variant);
  }

  // Convert Uint8Array to base64
  uint8ArrayToBase64(
    uint8Array,
    variant = this.sodium?.base64_variants?.URLSAFE_NO_PADDING,
  ) {
    if (!this.sodium) {
      throw new Error("CryptoService not initialized");
    }
    return this.sodium.to_base64(uint8Array, variant);
  }

  // Try different base64 variants for decoding
  tryDecodeBase64(base64String) {
    if (!this.sodium) {
      throw new Error("CryptoService not initialized");
    }

    const variants = [
      this.sodium.base64_variants.URLSAFE_NO_PADDING,
      this.sodium.base64_variants.ORIGINAL,
      this.sodium.base64_variants.URLSAFE,
      this.sodium.base64_variants.ORIGINAL_NO_PADDING,
    ];

    for (const variant of variants) {
      try {
        const decoded = this.sodium.from_base64(base64String, variant);
        console.log(
          `[CryptoService] Successfully decoded with variant ${variant}, length: ${decoded.length}`,
        );
        return decoded;
      } catch (error) {
        continue;
      }
    }

    // If all variants fail, try native atob as fallback
    try {
      const binaryString = atob(base64String);
      const bytes = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }
      console.log(
        `[CryptoService] Successfully decoded with native atob, length: ${bytes.length}`,
      );
      return bytes;
    } catch (error) {
      throw new Error(
        `Failed to decode base64 string with any variant: ${base64String.substring(0, 50)}...`,
      );
    }
  }

  // Main decryption flow for login challenge
  async decryptLoginChallenge(password, verifyData) {
    try {
      await this.initialize();

      console.log("[CryptoService] Starting login challenge decryption");
      console.log("[CryptoService] Verify data keys:", Object.keys(verifyData));

      // 1. Parse the verification data and validate
      const requiredFields = [
        "salt",
        "encryptedMasterKey",
        "encryptedPrivateKey",
        "encryptedChallenge",
      ];
      for (const field of requiredFields) {
        if (!verifyData[field]) {
          throw new Error(`Missing required field: ${field}`);
        }
      }

      // 2. Convert base64 data to Uint8Array with better error handling
      console.log("[CryptoService] Decoding base64 data...");

      let salt,
        encryptedMasterKey,
        encryptedPrivateKey,
        encryptedChallenge,
        storedPublicKey;

      try {
        console.log("[CryptoService] Decoding salt...");
        salt = this.tryDecodeBase64(verifyData.salt);
        console.log(`[CryptoService] Salt decoded, length: ${salt.length}`);
        this.hexDump(salt, "Salt");
      } catch (error) {
        throw new Error(`Failed to decode salt: ${error.message}`);
      }

      try {
        console.log("[CryptoService] Decoding encrypted master key...");
        encryptedMasterKey = this.tryDecodeBase64(
          verifyData.encryptedMasterKey,
        );
        console.log(
          `[CryptoService] Encrypted master key decoded, length: ${encryptedMasterKey.length}`,
        );
        this.hexDump(encryptedMasterKey, "Encrypted Master Key");
      } catch (error) {
        throw new Error(
          `Failed to decode encrypted master key: ${error.message}`,
        );
      }

      try {
        console.log("[CryptoService] Decoding encrypted private key...");
        encryptedPrivateKey = this.tryDecodeBase64(
          verifyData.encryptedPrivateKey,
        );
        console.log(
          `[CryptoService] Encrypted private key decoded, length: ${encryptedPrivateKey.length}`,
        );
        this.hexDump(encryptedPrivateKey, "Encrypted Private Key");
      } catch (error) {
        throw new Error(
          `Failed to decode encrypted private key: ${error.message}`,
        );
      }

      try {
        console.log("[CryptoService] Decoding encrypted challenge...");
        encryptedChallenge = this.tryDecodeBase64(
          verifyData.encryptedChallenge,
        );
        console.log(
          `[CryptoService] Encrypted challenge decoded, length: ${encryptedChallenge.length}`,
        );
        this.hexDump(encryptedChallenge, "Encrypted Challenge");
      } catch (error) {
        throw new Error(
          `Failed to decode encrypted challenge: ${error.message}`,
        );
      }

      // Also try to get the stored public key for comparison
      if (verifyData.publicKey) {
        try {
          storedPublicKey = this.tryDecodeBase64(verifyData.publicKey);
          console.log(
            `[CryptoService] Stored public key decoded, length: ${storedPublicKey.length}`,
          );
          this.hexDump(storedPublicKey, "Stored Public Key");
        } catch (error) {
          console.log(
            "[CryptoService] Could not decode stored public key:",
            error.message,
          );
        }
      }

      // 3. Validate data lengths
      console.log("[CryptoService] Validating data lengths...");

      // Salt should be 16 bytes for PBKDF2
      if (salt.length !== 16) {
        console.warn(
          `[CryptoService] Unexpected salt length: ${salt.length}, expected 16`,
        );
      }

      // Encrypted data should be at least nonce + some ciphertext + MAC
      const minSecretBoxLength =
        this.sodium.crypto_secretbox_NONCEBYTES +
        1 +
        this.sodium.crypto_secretbox_MACBYTES;
      if (encryptedMasterKey.length < minSecretBoxLength) {
        throw new Error(
          `Encrypted master key too short: ${encryptedMasterKey.length}, minimum: ${minSecretBoxLength}`,
        );
      }
      if (encryptedPrivateKey.length < minSecretBoxLength) {
        throw new Error(
          `Encrypted private key too short: ${encryptedPrivateKey.length}, minimum: ${minSecretBoxLength}`,
        );
      }

      // 4. Derive key encryption key from password
      console.log("[CryptoService] Deriving key from password...");
      const keyEncryptionKey = await this.deriveKeyFromPassword(password, salt);
      console.log(
        `[CryptoService] Key derived successfully, length: ${keyEncryptionKey.length}`,
      );

      // 5. Decrypt master key with KEK
      console.log("[CryptoService] Decrypting master key...");
      const masterKey = this.decryptWithSecretBox(
        encryptedMasterKey,
        keyEncryptionKey,
      );
      console.log(
        `[CryptoService] Master key decrypted successfully, length: ${masterKey.length}`,
      );

      // 6. Decrypt private key with master key
      console.log("[CryptoService] Decrypting private key...");
      const privateKey = this.decryptWithSecretBox(
        encryptedPrivateKey,
        masterKey,
      );
      console.log(
        `[CryptoService] Private key decrypted successfully, length: ${privateKey.length}`,
      );

      // Validate private key length
      if (privateKey.length !== this.sodium.crypto_box_SECRETKEYBYTES) {
        throw new Error(
          `Invalid private key length: ${privateKey.length}, expected: ${this.sodium.crypto_box_SECRETKEYBYTES}`,
        );
      }

      // Derive public key from private key and compare with stored one
      const derivedPublicKey = this.sodium.crypto_scalarmult_base(privateKey);
      console.log(
        `[CryptoService] Derived public key from private key, length: ${derivedPublicKey.length}`,
      );
      this.hexDump(derivedPublicKey, "Derived Public Key");

      if (storedPublicKey) {
        const publicKeysMatch = derivedPublicKey.every(
          (byte, index) => byte === storedPublicKey[index],
        );
        console.log(
          `[CryptoService] Public key comparison: ${publicKeysMatch ? "MATCH" : "MISMATCH"}`,
        );

        if (!publicKeysMatch) {
          console.error(
            "[CryptoService] Public key mismatch! This indicates wrong password or corrupted keys.",
          );
          throw new Error(
            "Public key mismatch - incorrect password or corrupted keys",
          );
        }
      }

      // 7. Decrypt challenge with private key
      console.log("[CryptoService] Decrypting challenge...");
      const decryptedChallenge = await this.decryptChallenge(
        encryptedChallenge,
        privateKey,
        storedPublicKey,
      );
      console.log(
        `[CryptoService] Challenge decrypted successfully, length: ${decryptedChallenge.length}`,
      );

      // 8. Return base64 encoded decrypted challenge
      const result = this.uint8ArrayToBase64(decryptedChallenge);
      console.log("[CryptoService] Decryption complete");

      return result;
    } catch (error) {
      console.error("[CryptoService] Decryption failed:", error);
      throw new Error(`Decryption failed: ${error.message}`);
    }
  }

  // Generate verification ID from public key (for reference)
  async generateVerificationID(publicKey) {
    await this.initialize();

    if (
      !(publicKey instanceof Uint8Array) ||
      publicKey.length !== this.sodium.crypto_box_PUBLICKEYBYTES
    ) {
      throw new Error("Invalid public key for verification ID generation");
    }

    // Hash the public key with SHA256
    const hashBuffer = await crypto.subtle.digest("SHA-256", publicKey);
    const hash = new Uint8Array(hashBuffer);

    // Convert hash to BIP39 mnemonic (24 words)
    const mnemonic = bip39.entropyToMnemonic(hash, wordlist);

    return mnemonic;
  }

  // Validate BIP39 mnemonic
  validateMnemonic(mnemonic) {
    return bip39.validateMnemonic(mnemonic, wordlist);
  }

  // Convert BIP39 mnemonic to seed (for recovery)
  mnemonicToSeed(mnemonic) {
    if (!this.validateMnemonic(mnemonic)) {
      throw new Error("Invalid mnemonic");
    }
    return bip39.mnemonicToSeedSync(mnemonic);
  }

  // Convert BIP39 mnemonic to recovery key (from registration prototype)
  async mnemonicToRecoveryKey(mnemonic) {
    // Validate the mnemonic
    if (!bip39.validateMnemonic(mnemonic, wordlist)) {
      throw new Error("Invalid mnemonic");
    }

    // Convert mnemonic to seed (this gives us 64 bytes)
    // This uses the BIP39 standard seed derivation (HMAC-SHA512)
    const seed = bip39.mnemonicToSeedSync(mnemonic);

    // Use first 32 bytes as our recovery key (sodium expects 32 bytes)
    // This is a non-standard derivation, but matches the original logic
    return seed.slice(0, 32);
  }

  // Generate E2EE data for registration (from registration prototype)
  async generateE2EEData(password) {
    try {
      console.log("Initializing libsodium...");
      await this.initialize();
      console.log("Libsodium ready!");

      // 1. Generate BIP39 mnemonic (12 words) for account recovery
      const mnemonicEntropy = this.sodium.randombytes_buf(16); // 128 bits = 12 words
      const recoveryMnemonic = bip39.entropyToMnemonic(
        mnemonicEntropy,
        wordlist,
      );
      console.log("Generated 12-word recovery mnemonic");

      // 2. Convert recovery mnemonic to recovery key
      const recoveryKey = await this.mnemonicToRecoveryKey(recoveryMnemonic);
      console.log("Derived recovery key from mnemonic");

      // 3. Generate salt for password key derivation
      const salt = this.sodium.randombytes_buf(16);
      console.log("Generated salt");

      // 4. Derive key encryption key from password using PBKDF2 (AES-GCM key)
      const keyEncryptionKey = await this.deriveKeyFromPassword(password, salt);
      console.log("Derived key encryption key from password");

      // 5. Generate X25519 key pair
      const keyPair = this.sodium.crypto_box_keypair();
      const publicKey = keyPair.publicKey; // 32 bytes
      const privateKey = keyPair.privateKey; // 32 bytes
      console.log("Generated X25519 key pair");

      // 6. Generate deterministic verification ID (24-word mnemonic) from public key
      const verificationID = await this.generateVerificationID(publicKey);
      console.log(
        "Generated verification ID (24-word mnemonic) from public key",
      );

      // 7. Generate master key (used for symmetric encryption of sensitive keys)
      const masterKey = this.sodium.randombytes_buf(32); // 32-byte secretbox key
      console.log("Generated master key");

      // 8. Encrypt master key with KEK using ChaCha20-Poly1305 (libsodium secretbox)
      const masterKeyNonce = this.sodium.randombytes_buf(
        this.sodium.crypto_secretbox_NONCEBYTES,
      ); // 24 bytes nonce
      const encryptedMasterKeyData = this.sodium.crypto_secretbox_easy(
        masterKey,
        masterKeyNonce,
        keyEncryptionKey, // 32 bytes KEK (from PBKDF2 AES-GCM derivation)
      ); // Result is ciphertext + 16 bytes MAC
      const encryptedMasterKey = new Uint8Array(
        masterKeyNonce.length + encryptedMasterKeyData.length,
      );
      encryptedMasterKey.set(masterKeyNonce, 0);
      encryptedMasterKey.set(encryptedMasterKeyData, masterKeyNonce.length);
      console.log("Encrypted master key with KEK");

      // 9. Encrypt private key with master key (ChaCha20-Poly1305)
      const privateKeyNonce = this.sodium.randombytes_buf(
        this.sodium.crypto_secretbox_NONCEBYTES,
      ); // 24 bytes nonce
      const encryptedPrivateKeyData = this.sodium.crypto_secretbox_easy(
        privateKey, // 32 bytes private key
        privateKeyNonce,
        masterKey, // 32 bytes master key
      ); // Result is ciphertext + 16 bytes MAC
      const encryptedPrivateKey = new Uint8Array(
        privateKeyNonce.length + encryptedPrivateKeyData.length,
      );
      encryptedPrivateKey.set(privateKeyNonce, 0);
      encryptedPrivateKey.set(encryptedPrivateKeyData, privateKeyNonce.length);
      console.log("Encrypted private key with master key");

      // 10. Encrypt recovery key with master key (ChaCha20-Poly1305)
      const recoveryKeyNonce = this.sodium.randombytes_buf(
        this.sodium.crypto_secretbox_NONCEBYTES,
      ); // 24 bytes nonce
      const encryptedRecoveryKeyData = this.sodium.crypto_secretbox_easy(
        recoveryKey, // 32 bytes recovery key
        recoveryKeyNonce,
        masterKey, // 32 bytes master key
      ); // Result is ciphertext + 16 bytes MAC
      const encryptedRecoveryKeyResult = new Uint8Array(
        recoveryKeyNonce.length + encryptedRecoveryKeyData.length,
      );
      encryptedRecoveryKeyResult.set(recoveryKeyNonce, 0);
      encryptedRecoveryKeyResult.set(
        encryptedRecoveryKeyData,
        recoveryKeyNonce.length,
      );
      console.log("Encrypted recovery key with master key");

      // 11. Encrypt master key with recovery key (ChaCha20-Poly1305) - for recovery path
      const masterWithRecoveryNonce = this.sodium.randombytes_buf(
        this.sodium.crypto_secretbox_NONCEBYTES,
      ); // 24 bytes nonce
      const masterWithRecoveryData = this.sodium.crypto_secretbox_easy(
        masterKey, // 32 bytes master key
        masterWithRecoveryNonce,
        recoveryKey, // 32 bytes recovery key
      ); // Result is ciphertext + 16 bytes MAC
      const masterKeyEncryptedWithRecoveryKey = new Uint8Array(
        masterWithRecoveryNonce.length + masterWithRecoveryData.length,
      );
      masterKeyEncryptedWithRecoveryKey.set(masterWithRecoveryNonce, 0);
      masterKeyEncryptedWithRecoveryKey.set(
        masterWithRecoveryData,
        masterWithRecoveryNonce.length,
      );
      console.log("Encrypted master key with recovery key");

      // 12. Encode everything to base64 URL safe without padding
      const result = {
        salt: this.sodium.to_base64(
          salt,
          this.sodium.base64_variants.URLSAFE_NO_PADDING,
        ),
        publicKey: this.sodium.to_base64(
          publicKey,
          this.sodium.base64_variants.URLSAFE_NO_PADDING,
        ),
        encryptedMasterKey: this.sodium.to_base64(
          encryptedMasterKey,
          this.sodium.base64_variants.URLSAFE_NO_PADDING,
        ),
        encryptedPrivateKey: this.sodium.to_base64(
          encryptedPrivateKey,
          this.sodium.base64_variants.URLSAFE_NO_PADDING,
        ),
        encryptedRecoveryKey: this.sodium.to_base64(
          encryptedRecoveryKeyResult,
          this.sodium.base64_variants.URLSAFE_NO_PADDING,
        ),
        masterKeyEncryptedWithRecoveryKey: this.sodium.to_base64(
          masterKeyEncryptedWithRecoveryKey,
          this.sodium.base64_variants.URLSAFE_NO_PADDING,
        ),
        verificationID: verificationID, // Now generated from public key
        // Include the recovery mnemonic so it can be displayed to the user
        recoveryMnemonic: recoveryMnemonic,
      };

      console.log("E2EE data generation complete.");
      console.log(
        "Recovery Mnemonic preview:",
        recoveryMnemonic.split(" ").slice(0, 3).join(" ") + "...",
      );
      console.log(
        "Verification ID preview:",
        verificationID.split(" ").slice(0, 3).join(" ") + "...",
      );
      return result;
    } catch (error) {
      console.error("Error generating E2EE data:", error);
      throw new Error("Failed to generate encryption data: " + error.message);
    }
  }
}

// Export singleton instance
export default new CryptoService();
