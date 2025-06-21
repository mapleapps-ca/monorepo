// monorepo/web/prototype/maplefile-register/src/crypto.jsx

// Crypto utility functions for E2EE using libsodium-wrappers-sumo
import sodium from "libsodium-wrappers-sumo";
import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";

// Simple PBKDF2 implementation for key derivation using Web Crypto API
const deriveKeyFromPassword = async (password, salt) => {
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
      iterations: 100000, // Recommended iterations
      hash: "SHA-256",
    },
    cryptoKey,
    { name: "AES-GCM", length: 256 }, // Derive a 256-bit key for AES-GCM
    true, // Exportable
    ["encrypt", "decrypt"],
  );

  const keyBuffer = await crypto.subtle.exportKey("raw", derivedKey);
  return new Uint8Array(keyBuffer); // Return as Uint8Array
};

// Convert BIP39 mnemonic to recovery key
const mnemonicToRecoveryKey = async (mnemonic) => {
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
};

// Generate VerificationID from public key (deterministic)
const generateVerificationID = async (publicKey) => {
  if (
    !(publicKey instanceof Uint8Array) ||
    publicKey.length !== sodium.crypto_box_PUBLICKEYBYTES
  ) {
    throw new Error("Invalid public key for verification ID generation.");
  }

  // 1. Hash the public key with SHA256 using Web Crypto API
  const hashBuffer = await crypto.subtle.digest("SHA-256", publicKey);
  const hash = new Uint8Array(hashBuffer); // SHA256 produces 32 bytes

  // 2. Use the hash as entropy for BIP39 (32 bytes entropy generates 24 words)
  // entropyToMnemonic expects 16, 20, 24, 28, or 32 bytes. 32 bytes = 24 words.
  const mnemonic = bip39.entropyToMnemonic(hash, wordlist);

  return mnemonic;
};

// Generate E2EE data for registration
export const generateE2EEData = async (password) => {
  try {
    console.log("Initializing libsodium...");
    await sodium.ready;
    console.log("Libsodium ready!");

    // 1. Generate BIP39 mnemonic (12 words) for account recovery
    const mnemonicEntropy = sodium.randombytes_buf(16); // 128 bits = 12 words
    const recoveryMnemonic = bip39.entropyToMnemonic(mnemonicEntropy, wordlist);
    console.log("Generated 12-word recovery mnemonic");

    // 2. Convert recovery mnemonic to recovery key
    const recoveryKey = await mnemonicToRecoveryKey(recoveryMnemonic);
    console.log("Derived recovery key from mnemonic");

    // 3. Generate salt for password key derivation
    const salt = sodium.randombytes_buf(16);
    console.log("Generated salt");

    // 4. Derive key encryption key from password using PBKDF2 (AES-GCM key)
    const keyEncryptionKey = await deriveKeyFromPassword(password, salt);
    console.log("Derived key encryption key from password");

    // 5. Generate X25519 key pair
    const keyPair = sodium.crypto_box_keypair();
    const publicKey = keyPair.publicKey; // 32 bytes
    const privateKey = keyPair.privateKey; // 32 bytes
    console.log("Generated X25519 key pair");

    // 6. Generate deterministic verification ID (24-word mnemonic) from public key
    const verificationID = await generateVerificationID(publicKey);
    console.log("Generated verification ID (24-word mnemonic) from public key");

    // 7. Generate master key (used for symmetric encryption of sensitive keys)
    const masterKey = sodium.randombytes_buf(32); // 32-byte secretbox key
    console.log("Generated master key");

    // 8. Encrypt master key with KEK using ChaCha20-Poly1305 (libsodium secretbox)
    const masterKeyNonce = sodium.randombytes_buf(
      sodium.crypto_secretbox_NONCEBYTES,
    ); // 24 bytes nonce
    const encryptedMasterKeyData = sodium.crypto_secretbox_easy(
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
    const privateKeyNonce = sodium.randombytes_buf(
      sodium.crypto_secretbox_NONCEBYTES,
    ); // 24 bytes nonce
    const encryptedPrivateKeyData = sodium.crypto_secretbox_easy(
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
    const recoveryKeyNonce = sodium.randombytes_buf(
      sodium.crypto_secretbox_NONCEBYTES,
    ); // 24 bytes nonce
    const encryptedRecoveryKeyData = sodium.crypto_secretbox_easy(
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
    const masterWithRecoveryNonce = sodium.randombytes_buf(
      sodium.crypto_secretbox_NONCEBYTES,
    ); // 24 bytes nonce
    const masterWithRecoveryData = sodium.crypto_secretbox_easy(
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
      salt: sodium.to_base64(salt, sodium.base64_variants.URLSAFE_NO_PADDING),
      publicKey: sodium.to_base64(
        publicKey,
        sodium.base64_variants.URLSAFE_NO_PADDING,
      ),
      encryptedMasterKey: sodium.to_base64(
        encryptedMasterKey,
        sodium.base64_variants.URLSAFE_NO_PADDING,
      ),
      encryptedPrivateKey: sodium.to_base64(
        encryptedPrivateKey,
        sodium.base64_variants.URLSAFE_NO_PADDING,
      ),
      encryptedRecoveryKey: sodium.to_base64(
        encryptedRecoveryKeyResult,
        sodium.base64_variants.URLSAFE_NO_PADDING,
      ),
      masterKeyEncryptedWithRecoveryKey: sodium.to_base64(
        masterKeyEncryptedWithRecoveryKey,
        sodium.base64_variants.URLSAFE_NO_PADDING,
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
};
