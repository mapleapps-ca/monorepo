// File: monorepo/web/maplefile-frontend/src/services/Manager/RecoveryManager.js
// Recovery Manager - Orchestrates API, Storage and Crypto services for account recovery

import RecoveryAPIService from "../API/RecoveryAPIService.js";
import RecoveryStorageService from "../Storage/RecoveryStorageService.js";
import CryptoService from "../Crypto/CryptoService.js";

class RecoveryManager {
  constructor() {
    this.isInitialized = false;

    // Initialize dependent services
    this.apiService = new RecoveryAPIService();
    this.storageService = new RecoveryStorageService();

    console.log("[RecoveryManager] Recovery manager initialized");
  }

  // Initialize the manager
  async initialize() {
    if (this.isInitialized) return;

    try {
      console.log("[RecoveryManager] Initializing recovery manager...");

      // Initialize CryptoService
      await CryptoService.initialize();

      this.isInitialized = true;
      console.log(
        "[RecoveryManager] Recovery manager initialized successfully",
      );
    } catch (error) {
      console.error("[RecoveryManager] Failed to initialize:", error);
      this.isInitialized = true; // Continue anyway
    }
  }

  // === Recovery Flow Methods ===

  // Step 1: Initiate Recovery
  async initiateRecovery(email) {
    try {
      console.log(
        "[RecoveryManager] Orchestrating recovery initiation for:",
        email,
      );

      // Store email for later steps
      this.storageService.storeRecoveryEmail(email);

      // Call API to initiate recovery
      const response = await this.apiService.initiateRecovery(email);

      // Store session data
      this.storageService.storeRecoverySession(response);

      console.log("[RecoveryManager] Recovery initiation successful");
      return response;
    } catch (error) {
      console.error("[RecoveryManager] Recovery initiation failed:", error);
      throw error;
    }
  }

  // Step 2: Decrypt challenge with recovery phrase
  async decryptChallengeWithRecoveryPhrase(recoveryPhrase) {
    try {
      console.log(
        "[RecoveryManager] Decrypting challenge with recovery phrase",
      );

      // Validate recovery phrase
      if (!CryptoService.validateMnemonic(recoveryPhrase)) {
        throw new Error("Invalid recovery phrase format");
      }

      // Get stored encrypted challenge
      const encryptedChallenge = this.storageService.getEncryptedChallenge();
      if (!encryptedChallenge) {
        throw new Error(
          "No encrypted challenge found - please restart recovery",
        );
      }

      // Convert recovery phrase to recovery key
      console.log(
        "[RecoveryManager] Converting recovery phrase to recovery key...",
      );
      const recoveryKey =
        await CryptoService.mnemonicToRecoveryKey(recoveryPhrase);

      // Decode encrypted challenge
      const encryptedChallengeBytes =
        CryptoService.tryDecodeBase64(encryptedChallenge);

      // Decrypt challenge using recovery key
      console.log(
        "[RecoveryManager] Decrypting challenge with recovery key...",
      );

      // The challenge is encrypted using sealed box with user's public key
      // But we don't have the private key yet (it's encrypted with master key)
      // So the backend should encrypt the challenge in a way that can be decrypted with recovery key

      // According to the flow, the challenge should be decryptable with the recovery key
      // This might be a custom implementation where the challenge is encrypted with recovery key

      let decryptedChallenge;

      try {
        // First, let's check if this is encrypted data format (nonce + ciphertext)
        // The backend documentation mentions ChaCha20-Poly1305 encryption

        if (encryptedChallengeBytes.length > 40) {
          // At least nonce (12/24) + some data + MAC (16)
          console.log(
            "[RecoveryManager] Attempting to decrypt challenge as ChaCha20-Poly1305",
          );

          // Try with 12-byte nonce (ChaCha20-Poly1305 IETF)
          try {
            const nonce = encryptedChallengeBytes.slice(0, 12);
            const ciphertext = encryptedChallengeBytes.slice(12);

            decryptedChallenge =
              CryptoService.sodium.crypto_aead_chacha20poly1305_ietf_decrypt(
                null,
                ciphertext,
                null,
                nonce,
                recoveryKey,
              );
            console.log(
              "[RecoveryManager] Challenge decrypted using ChaCha20-Poly1305 IETF",
            );
          } catch (e1) {
            // Try with 24-byte nonce (secretbox)
            console.log(
              "[RecoveryManager] ChaCha20-Poly1305 IETF failed, trying secretbox",
            );
            decryptedChallenge = CryptoService.decryptWithSecretBox(
              encryptedChallengeBytes,
              recoveryKey,
            );
            console.log(
              "[RecoveryManager] Challenge decrypted using secretbox",
            );
          }
        } else {
          throw new Error("Encrypted challenge too short");
        }
      } catch (error) {
        console.log(
          "[RecoveryManager] Failed to decrypt challenge:",
          error.message,
        );
        throw new Error(
          "Failed to decrypt challenge with recovery key. Please check your recovery phrase.",
        );
      }

      // Convert to base64 for API
      const decryptedChallengeBase64 =
        CryptoService.uint8ArrayToBase64(decryptedChallenge);

      console.log("[RecoveryManager] Challenge decrypted successfully");
      return decryptedChallengeBase64;
    } catch (error) {
      console.error("[RecoveryManager] Challenge decryption failed:", error);
      throw error;
    }
  }

  // Step 2: Verify Recovery
  async verifyRecovery(decryptedChallenge) {
    try {
      console.log("[RecoveryManager] Orchestrating recovery verification");

      const sessionId = this.storageService.getRecoverySessionId();
      if (!sessionId) {
        throw new Error("No recovery session found - please restart recovery");
      }

      // Call API to verify recovery
      const response = await this.apiService.verifyRecovery(
        sessionId,
        decryptedChallenge,
      );

      // Store verification response
      this.storageService.storeVerificationResponse(response);

      console.log("[RecoveryManager] Recovery verification successful");
      return response;
    } catch (error) {
      console.error("[RecoveryManager] Recovery verification failed:", error);
      throw error;
    }
  }

  // Complete recovery with recovery phrase and new password (combined method)
  async completeRecoveryWithPhrase(recoveryPhrase, newPassword) {
    try {
      console.log(
        "[RecoveryManager] Completing recovery with phrase and new password",
      );

      // Get verification response
      const verifyResponse = this.storageService.getVerificationResponse();
      if (!verifyResponse || !verifyResponse.recovery_token) {
        throw new Error(
          "No verification data found - please complete verification first",
        );
      }

      // Validate recovery phrase
      if (!CryptoService.validateMnemonic(recoveryPhrase)) {
        throw new Error("Invalid recovery phrase format");
      }

      // Convert recovery phrase to recovery key
      const recoveryKey =
        await CryptoService.mnemonicToRecoveryKey(recoveryPhrase);

      // Get master key encrypted with recovery key
      const masterKeyWithRecovery =
        verifyResponse.master_key_encrypted_with_recovery_key;
      if (!masterKeyWithRecovery) {
        throw new Error(
          "No encrypted master key found in verification response",
        );
      }

      // Decrypt master key using recovery key
      const encryptedMasterKeyBytes = CryptoService.tryDecodeBase64(
        masterKeyWithRecovery,
      );
      const masterKey = CryptoService.decryptWithSecretBox(
        encryptedMasterKeyBytes,
        recoveryKey,
      );

      // Generate new salt for password
      const newSalt = CryptoService.sodium.randombytes_buf(16);

      // Derive new key encryption key from new password
      const newKEK = await CryptoService.deriveKeyFromPassword(
        newPassword,
        newSalt,
      );

      // Generate new key pair
      const newKeyPair = CryptoService.sodium.crypto_box_keypair();

      // Encrypt master key with new KEK
      const newMasterKeyNonce = CryptoService.sodium.randombytes_buf(
        CryptoService.sodium.crypto_secretbox_NONCEBYTES,
      );
      const newEncryptedMasterKeyData =
        CryptoService.sodium.crypto_secretbox_easy(
          masterKey,
          newMasterKeyNonce,
          newKEK,
        );
      const newEncryptedMasterKey = new Uint8Array(
        newMasterKeyNonce.length + newEncryptedMasterKeyData.length,
      );
      newEncryptedMasterKey.set(newMasterKeyNonce, 0);
      newEncryptedMasterKey.set(
        newEncryptedMasterKeyData,
        newMasterKeyNonce.length,
      );

      // Encrypt new private key with master key
      const newPrivateKeyNonce = CryptoService.sodium.randombytes_buf(
        CryptoService.sodium.crypto_secretbox_NONCEBYTES,
      );
      const newEncryptedPrivateKeyData =
        CryptoService.sodium.crypto_secretbox_easy(
          newKeyPair.privateKey,
          newPrivateKeyNonce,
          masterKey,
        );
      const newEncryptedPrivateKey = new Uint8Array(
        newPrivateKeyNonce.length + newEncryptedPrivateKeyData.length,
      );
      newEncryptedPrivateKey.set(newPrivateKeyNonce, 0);
      newEncryptedPrivateKey.set(
        newEncryptedPrivateKeyData,
        newPrivateKeyNonce.length,
      );

      // Encrypt recovery key with master key
      const newRecoveryKeyNonce = CryptoService.sodium.randombytes_buf(
        CryptoService.sodium.crypto_secretbox_NONCEBYTES,
      );
      const newEncryptedRecoveryKeyData =
        CryptoService.sodium.crypto_secretbox_easy(
          recoveryKey,
          newRecoveryKeyNonce,
          masterKey,
        );
      const newEncryptedRecoveryKey = new Uint8Array(
        newRecoveryKeyNonce.length + newEncryptedRecoveryKeyData.length,
      );
      newEncryptedRecoveryKey.set(newRecoveryKeyNonce, 0);
      newEncryptedRecoveryKey.set(
        newEncryptedRecoveryKeyData,
        newRecoveryKeyNonce.length,
      );

      // Encrypt master key with recovery key (for future recovery)
      const newMasterWithRecoveryNonce = CryptoService.sodium.randombytes_buf(
        CryptoService.sodium.crypto_secretbox_NONCEBYTES,
      );
      const newMasterWithRecoveryData =
        CryptoService.sodium.crypto_secretbox_easy(
          masterKey,
          newMasterWithRecoveryNonce,
          recoveryKey,
        );
      const newMasterKeyWithRecoveryKey = new Uint8Array(
        newMasterWithRecoveryNonce.length + newMasterWithRecoveryData.length,
      );
      newMasterKeyWithRecoveryKey.set(newMasterWithRecoveryNonce, 0);
      newMasterKeyWithRecoveryKey.set(
        newMasterWithRecoveryData,
        newMasterWithRecoveryNonce.length,
      );

      // Prepare recovery completion data
      const recoveryData = {
        recovery_token: verifyResponse.recovery_token,
        new_salt: CryptoService.uint8ArrayToBase64(newSalt),
        new_encrypted_master_key: CryptoService.uint8ArrayToBase64(
          newEncryptedMasterKey,
        ),
        new_encrypted_private_key: CryptoService.uint8ArrayToBase64(
          newEncryptedPrivateKey,
        ),
        new_encrypted_recovery_key: CryptoService.uint8ArrayToBase64(
          newEncryptedRecoveryKey,
        ),
        new_master_key_encrypted_with_recovery_key:
          CryptoService.uint8ArrayToBase64(newMasterKeyWithRecoveryKey),
      };

      // Call API to complete recovery
      const response = await this.apiService.completeRecovery(recoveryData);

      // Clear recovery session data
      this.storageService.clearRecoverySession();

      console.log("[RecoveryManager] Recovery completed successfully");
      return response;
    } catch (error) {
      console.error("[RecoveryManager] Recovery completion failed:", error);
      throw error;
    }
  }

  // === Recovery State Management ===

  // Check if recovery session is active
  hasActiveRecoverySession() {
    return this.storageService.hasValidRecoverySession();
  }

  // Check if verification is complete
  isVerificationComplete() {
    return this.storageService.hasVerificationResponse();
  }

  // Get recovery email
  getRecoveryEmail() {
    return this.storageService.getRecoveryEmail();
  }

  // Clear recovery session
  clearRecoverySession() {
    this.storageService.clearRecoverySession();
    console.log("[RecoveryManager] Recovery session cleared");
  }

  // === Debug and Information ===

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "RecoveryManager",
      role: "orchestrator",
      isInitialized: this.isInitialized,
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      recoveryState: {
        hasActiveSession: this.hasActiveRecoverySession(),
        isVerificationComplete: this.isVerificationComplete(),
        recoveryEmail: this.getRecoveryEmail(),
      },
    };
  }

  // Get recovery status
  getRecoveryStatus() {
    return {
      hasActiveSession: this.hasActiveRecoverySession(),
      isVerificationComplete: this.isVerificationComplete(),
      recoveryEmail: this.getRecoveryEmail(),
      sessionData: this.storageService.getStorageInfo(),
    };
  }
}

export default RecoveryManager;
