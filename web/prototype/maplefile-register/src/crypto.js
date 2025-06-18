// Crypto utility functions for E2EE using libsodium-wrappers-sumo
import sodium from "libsodium-wrappers-sumo";

// Simple PBKDF2 implementation for key derivation (since we're avoiding argon2-browser)
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
      iterations: 100000,
      hash: "SHA-256",
    },
    cryptoKey,
    { name: "AES-GCM", length: 256 },
    true,
    ["encrypt", "decrypt"],
  );

  const keyBuffer = await crypto.subtle.exportKey("raw", derivedKey);
  return new Uint8Array(keyBuffer);
};

// Generate E2EE data for registration
export const generateE2EEData = async (password) => {
  try {
    console.log("Initializing libsodium...");
    await sodium.ready;
    console.log("Libsodium ready!");

    // 1. Generate salt for password key derivation
    const salt = sodium.randombytes_buf(16);
    console.log("Generated salt");

    // 2. Derive key encryption key from password using PBKDF2
    const keyEncryptionKey = await deriveKeyFromPassword(password, salt);
    console.log("Derived key encryption key");

    // 3. Generate X25519 key pair
    const keyPair = sodium.crypto_box_keypair();
    const publicKey = keyPair.publicKey;
    const privateKey = keyPair.privateKey;
    console.log("Generated key pair");

    // 4. Generate master key and recovery key
    const masterKey = sodium.randombytes_buf(32);
    const recoveryKey = sodium.randombytes_buf(32);
    console.log("Generated master and recovery keys");

    // 5. Encrypt master key with KEK using ChaCha20-Poly1305
    const masterKeyNonce = sodium.randombytes_buf(
      sodium.crypto_secretbox_NONCEBYTES,
    );
    const encryptedMasterKeyData = sodium.crypto_secretbox_easy(
      masterKey,
      masterKeyNonce,
      keyEncryptionKey,
    );
    const encryptedMasterKey = new Uint8Array(
      masterKeyNonce.length + encryptedMasterKeyData.length,
    );
    encryptedMasterKey.set(masterKeyNonce, 0);
    encryptedMasterKey.set(encryptedMasterKeyData, masterKeyNonce.length);

    // 6. Encrypt private key with master key
    const privateKeyNonce = sodium.randombytes_buf(
      sodium.crypto_secretbox_NONCEBYTES,
    );
    const encryptedPrivateKeyData = sodium.crypto_secretbox_easy(
      privateKey,
      privateKeyNonce,
      masterKey,
    );
    const encryptedPrivateKey = new Uint8Array(
      privateKeyNonce.length + encryptedPrivateKeyData.length,
    );
    encryptedPrivateKey.set(privateKeyNonce, 0);
    encryptedPrivateKey.set(encryptedPrivateKeyData, privateKeyNonce.length);

    // 7. Encrypt recovery key with master key
    const recoveryKeyNonce = sodium.randombytes_buf(
      sodium.crypto_secretbox_NONCEBYTES,
    );
    const encryptedRecoveryKeyData = sodium.crypto_secretbox_easy(
      recoveryKey,
      recoveryKeyNonce,
      masterKey,
    );
    const encryptedRecoveryKeyResult = new Uint8Array(
      recoveryKeyNonce.length + encryptedRecoveryKeyData.length,
    );
    encryptedRecoveryKeyResult.set(recoveryKeyNonce, 0);
    encryptedRecoveryKeyResult.set(
      encryptedRecoveryKeyData,
      recoveryKeyNonce.length,
    );

    // 8. Encrypt master key with recovery key
    const masterWithRecoveryNonce = sodium.randombytes_buf(
      sodium.crypto_secretbox_NONCEBYTES,
    );
    const masterWithRecoveryData = sodium.crypto_secretbox_easy(
      masterKey,
      masterWithRecoveryNonce,
      recoveryKey,
    );
    const masterKeyEncryptedWithRecoveryKey = new Uint8Array(
      masterWithRecoveryNonce.length + masterWithRecoveryData.length,
    );
    masterKeyEncryptedWithRecoveryKey.set(masterWithRecoveryNonce, 0);
    masterKeyEncryptedWithRecoveryKey.set(
      masterWithRecoveryData,
      masterWithRecoveryNonce.length,
    );

    console.log("Encrypted all keys");

    // 9. Let the server generate the verification ID from the public key
    // The backend will generate the proper BIP39 mnemonic from the public key's SHA256 hash
    console.log("Server will generate verification ID from public key");

    // 10. Encode everything to base64 URL
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
      verificationID: "", // Let server generate the proper BIP39 mnemonic
    };

    console.log("E2EE data generated successfully");
    console.log(
      "Public key (first 16 chars):",
      result.publicKey.substring(0, 16) + "...",
    );
    console.log("Verification ID will be generated server-side");
    return result;
  } catch (error) {
    console.error("Error generating E2EE data:", error);
    throw new Error("Failed to generate encryption data: " + error.message);
  }
};
