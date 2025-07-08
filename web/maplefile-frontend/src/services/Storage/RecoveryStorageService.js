// File: monorepo/web/maplefile-frontend/src/services/Storage/RecoveryStorageService.js
// Recovery Storage Service - Handles storage operations for account recovery

class RecoveryStorageService {
  constructor() {
    // Storage keys for recovery session data
    this.STORAGE_KEYS = {
      RECOVERY_SESSION_ID: "mapleapps_recovery_session_id",
      RECOVERY_CHALLENGE_ID: "mapleapps_recovery_challenge_id",
      RECOVERY_EMAIL: "mapleapps_recovery_email",
      RECOVERY_SESSION_DATA: "mapleapps_recovery_session_data",
      RECOVERY_VERIFY_RESPONSE: "mapleapps_recovery_verify_response",
      RECOVERY_ENCRYPTED_CHALLENGE: "mapleapps_recovery_encrypted_challenge",
      RECOVERY_MASTER_KEY_WITH_RECOVERY:
        "mapleapps_recovery_master_key_with_recovery",
    };

    console.log("[RecoveryStorageService] Storage service initialized");
  }

  // === Session Management ===

  // Store recovery session data from initiate step
  storeRecoverySession(sessionData) {
    try {
      if (sessionData.session_id) {
        sessionStorage.setItem(
          this.STORAGE_KEYS.RECOVERY_SESSION_ID,
          sessionData.session_id,
        );
      }

      if (sessionData.challenge_id) {
        sessionStorage.setItem(
          this.STORAGE_KEYS.RECOVERY_CHALLENGE_ID,
          sessionData.challenge_id,
        );
      }

      if (sessionData.encrypted_challenge) {
        sessionStorage.setItem(
          this.STORAGE_KEYS.RECOVERY_ENCRYPTED_CHALLENGE,
          sessionData.encrypted_challenge,
        );
      }

      // Store full session data
      sessionStorage.setItem(
        this.STORAGE_KEYS.RECOVERY_SESSION_DATA,
        JSON.stringify(sessionData),
      );

      console.log("[RecoveryStorageService] Recovery session data stored");
    } catch (error) {
      console.error(
        "[RecoveryStorageService] Failed to store recovery session:",
        error,
      );
    }
  }

  // Get recovery session data
  getRecoverySession() {
    const sessionData = sessionStorage.getItem(
      this.STORAGE_KEYS.RECOVERY_SESSION_DATA,
    );
    return sessionData ? JSON.parse(sessionData) : null;
  }

  // Get recovery session ID
  getRecoverySessionId() {
    return sessionStorage.getItem(this.STORAGE_KEYS.RECOVERY_SESSION_ID);
  }

  // Get recovery challenge ID
  getRecoveryChallengeId() {
    return sessionStorage.getItem(this.STORAGE_KEYS.RECOVERY_CHALLENGE_ID);
  }

  // Get encrypted challenge
  getEncryptedChallenge() {
    return sessionStorage.getItem(
      this.STORAGE_KEYS.RECOVERY_ENCRYPTED_CHALLENGE,
    );
  }

  // === Email Management ===

  // Store recovery email
  storeRecoveryEmail(email) {
    if (email) {
      sessionStorage.setItem(this.STORAGE_KEYS.RECOVERY_EMAIL, email);
      console.log("[RecoveryStorageService] Recovery email stored");
    }
  }

  // Get recovery email
  getRecoveryEmail() {
    return sessionStorage.getItem(this.STORAGE_KEYS.RECOVERY_EMAIL);
  }

  // === Verification Response ===

  // Store verification response data
  storeVerificationResponse(verifyData) {
    try {
      sessionStorage.setItem(
        this.STORAGE_KEYS.RECOVERY_VERIFY_RESPONSE,
        JSON.stringify(verifyData),
      );

      // Store master key encrypted with recovery key separately for easy access
      if (verifyData.master_key_encrypted_with_recovery_key) {
        sessionStorage.setItem(
          this.STORAGE_KEYS.RECOVERY_MASTER_KEY_WITH_RECOVERY,
          verifyData.master_key_encrypted_with_recovery_key,
        );
      }

      console.log("[RecoveryStorageService] Verification response stored");
    } catch (error) {
      console.error(
        "[RecoveryStorageService] Failed to store verification response:",
        error,
      );
    }
  }

  // Get verification response data
  getVerificationResponse() {
    const data = sessionStorage.getItem(
      this.STORAGE_KEYS.RECOVERY_VERIFY_RESPONSE,
    );
    return data ? JSON.parse(data) : null;
  }

  // Get master key encrypted with recovery key
  getMasterKeyWithRecoveryKey() {
    return sessionStorage.getItem(
      this.STORAGE_KEYS.RECOVERY_MASTER_KEY_WITH_RECOVERY,
    );
  }

  // === Session Validation ===

  // Check if recovery session is valid
  hasValidRecoverySession() {
    const sessionId = this.getRecoverySessionId();
    const challengeId = this.getRecoveryChallengeId();
    const encryptedChallenge = this.getEncryptedChallenge();
    const email = this.getRecoveryEmail();

    return !!(sessionId && challengeId && encryptedChallenge && email);
  }

  // Check if verification is complete
  hasVerificationResponse() {
    const verifyResponse = this.getVerificationResponse();
    return !!(verifyResponse && verifyResponse.recovery_token);
  }

  // === Cleanup ===

  // Clear all recovery session data
  clearRecoverySession() {
    Object.values(this.STORAGE_KEYS).forEach((key) => {
      sessionStorage.removeItem(key);
    });
    console.log("[RecoveryStorageService] Recovery session cleared");
  }

  // Clear specific recovery data
  clearRecoveryData(key) {
    if (this.STORAGE_KEYS[key]) {
      sessionStorage.removeItem(this.STORAGE_KEYS[key]);
    }
  }

  // === Storage Information ===

  // Get storage info for debugging
  getStorageInfo() {
    return {
      hasRecoverySession: this.hasValidRecoverySession(),
      hasVerificationResponse: this.hasVerificationResponse(),
      recoveryEmail: this.getRecoveryEmail(),
      sessionId: !!this.getRecoverySessionId(),
      challengeId: !!this.getRecoveryChallengeId(),
      encryptedChallenge: !!this.getEncryptedChallenge(),
      verifyResponse: !!this.getVerificationResponse(),
      masterKeyWithRecovery: !!this.getMasterKeyWithRecoveryKey(),
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "RecoveryStorageService",
      storageInfo: this.getStorageInfo(),
      storageKeys: Object.keys(this.STORAGE_KEYS),
    };
  }
}

export default RecoveryStorageService;
