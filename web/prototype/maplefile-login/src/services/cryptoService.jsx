// monorepo/web/prototype/maplefile-login/src/services/cryptoService.jsx

// Production Crypto Service for E2EE operations
import sodium from "libsodium-wrappers-sumo";
import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";

class CryptoService {
  constructor() {
    this.isInitialized = false;
  }

  // Initialize libsodium
  async initialize() {
    if (this.isInitialized) return;

    await sodium.ready;
    this.isInitialized = true;
    console.log("[CryptoService] Libsodium initialized");
  }

  // Derive key from password using PBKDF2 (keeping current working approach)
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
        iterations: 100000, // Same as registration
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
      if (key.length !== sodium.crypto_secretbox_KEYBYTES) {
        throw new Error(
          `Invalid key length: ${key.length}, expected: ${sodium.crypto_secretbox_KEYBYTES}`,
        );
      }

      // encryptedData format: nonce (24 bytes) + ciphertext + mac (16 bytes)
      const nonceLength = sodium.crypto_secretbox_NONCEBYTES; // 24 bytes
      const macLength = sodium.crypto_secretbox_MACBYTES; // 16 bytes

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
      const decrypted = sodium.crypto_secretbox_open_easy(
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
      if (privateKey.length !== sodium.crypto_box_SECRETKEYBYTES) {
        throw new Error(
          `Invalid private key length: ${privateKey.length}, expected: ${sodium.crypto_box_SECRETKEYBYTES}`,
        );
      }

      // For anonymous box (SealAnonymous/OpenAnonymous), the format is different
      // It should just be: ephemeral_public_key + ciphertext (no separate nonce)
      // The total overhead is 48 bytes (32 bytes ephemeral key + 16 bytes MAC)

      const expectedMinLength = sodium.crypto_box_SEALBYTES; // Should be 48
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
        const derivedPublicKey = sodium.crypto_scalarmult_base(privateKey);

        console.log("[CryptoService] Derived public key for sealed box");
        this.hexDump(derivedPublicKey, "Derived Public Key");

        // Try with derived public key first
        let decrypted;
        try {
          decrypted = sodium.crypto_box_seal_open(
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
              decrypted = sodium.crypto_box_seal_open(
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
          const ephemeralPubKeyLength = sodium.crypto_box_PUBLICKEYBYTES; // 32 bytes
          const nonceLength = sodium.crypto_box_NONCEBYTES; // 24 bytes
          const macLength = sodium.crypto_box_MACBYTES; // 16 bytes

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
          const decrypted = sodium.crypto_box_open_easy(
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
    variant = sodium.base64_variants.URLSAFE_NO_PADDING,
  ) {
    return sodium.from_base64(base64String, variant);
  }

  // Convert Uint8Array to base64
  uint8ArrayToBase64(
    uint8Array,
    variant = sodium.base64_variants.URLSAFE_NO_PADDING,
  ) {
    return sodium.to_base64(uint8Array, variant);
  }

  // Try different base64 variants for decoding
  tryDecodeBase64(base64String) {
    const variants = [
      sodium.base64_variants.URLSAFE_NO_PADDING,
      sodium.base64_variants.ORIGINAL,
      sodium.base64_variants.URLSAFE,
      sodium.base64_variants.ORIGINAL_NO_PADDING,
    ];

    for (const variant of variants) {
      try {
        const decoded = sodium.from_base64(base64String, variant);
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
        sodium.crypto_secretbox_NONCEBYTES +
        1 +
        sodium.crypto_secretbox_MACBYTES;
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
      if (privateKey.length !== sodium.crypto_box_SECRETKEYBYTES) {
        throw new Error(
          `Invalid private key length: ${privateKey.length}, expected: ${sodium.crypto_box_SECRETKEYBYTES}`,
        );
      }

      // Derive public key from private key and compare with stored one
      const derivedPublicKey = sodium.crypto_scalarmult_base(privateKey);
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
      publicKey.length !== sodium.crypto_box_PUBLICKEYBYTES
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
}

// Export singleton instance
export default new CryptoService();
