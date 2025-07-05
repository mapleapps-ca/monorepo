// File: monorepo/web/maplefile-frontend/src/services/Manager/AuthManager.js
// Authentication Manager - Orchestrates API and Storage services for authentication flows
import AuthAPIService from "../API/AuthAPIService.js";
import AuthStorageService from "../Storage/AuthStorageService.js";
import CryptoService from "../Crypto/CryptoService.js";

class AuthManager {
  constructor() {
    this.isInitialized = false;

    // Initialize dependent services
    this.apiService = new AuthAPIService();
    this.storageService = new AuthStorageService();

    // Event listener management (replacing WorkerManager)
    this.authStateListeners = new Set();

    console.log("[AuthManager] Authentication manager initialized");
  }

  // Initialize the manager (simplified without workers)
  async initializeWorker() {
    if (this.isInitialized) return;

    try {
      console.log("[AuthManager] Initializing auth manager (no workers)...");
      this.isInitialized = true;
      console.log("[AuthManager] Auth manager initialized successfully");
    } catch (error) {
      console.error("[AuthManager] Failed to initialize:", error);
      this.isInitialized = true; // Continue anyway
    }
  }

  // === Event Management (Replacing WorkerManager) ===

  // Add auth state change listener
  addAuthStateChangeListener(callback) {
    if (typeof callback === "function") {
      this.authStateListeners.add(callback);
      console.log(
        "[AuthManager] Auth state listener added. Total listeners:",
        this.authStateListeners.size,
      );
    }
  }

  // Remove auth state change listener
  removeAuthStateChangeListener(callback) {
    this.authStateListeners.delete(callback);
    console.log(
      "[AuthManager] Auth state listener removed. Total listeners:",
      this.authStateListeners.size,
    );
  }

  // Notify auth state change
  notifyAuthStateChange(eventType, eventData) {
    console.log(
      `[AuthManager] Notifying ${this.authStateListeners.size} listeners of ${eventType}`,
    );

    this.authStateListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error("[AuthManager] Error in auth state listener:", error);
      }
    });
  }

  // === Authentication Flow Methods ===

  // Step 1: Request One-Time Token (OTT)
  async requestOTT(email) {
    try {
      console.log("[AuthManager] Orchestrating OTT request for:", email);

      const response = await this.apiService.requestOTT(email);

      // Store email for next steps
      this.storageService.setUserEmail(email);

      console.log("[AuthManager] OTT request flow completed successfully");
      return response;
    } catch (error) {
      console.error("[AuthManager] OTT request flow failed:", error);
      throw error;
    }
  }

  // Step 2: Verify One-Time Token
  async verifyOTT(email, ott) {
    try {
      console.log("[AuthManager] Orchestrating OTT verification for:", email);

      const response = await this.apiService.verifyOTT(email, ott);

      // Store verification data for the final step
      this.storageService.setLoginSessionData("verify_response", response);

      // Store the user's encrypted data for future password-based decryption
      if (
        response.salt &&
        response.encryptedMasterKey &&
        response.encryptedPrivateKey
      ) {
        console.log(
          "[AuthManager] Storing user's encrypted data for future use",
        );
        this.storageService.storeUserEncryptedData(
          response.salt,
          response.encryptedMasterKey,
          response.encryptedPrivateKey,
          response.publicKey || response.userPublicKey,
        );
      } else {
        // Try alternative field names
        const salt = response.salt || response.Salt || response.password_salt;
        const encryptedMasterKey =
          response.encryptedMasterKey ||
          response.encrypted_master_key ||
          response.masterKey ||
          response.master_key;
        const encryptedPrivateKey =
          response.encryptedPrivateKey ||
          response.encrypted_private_key ||
          response.privateKey ||
          response.private_key;
        const publicKey =
          response.publicKey ||
          response.public_key ||
          response.userPublicKey ||
          response.user_public_key;

        if (salt && encryptedMasterKey && encryptedPrivateKey) {
          console.log(
            "[AuthManager] Storing user's encrypted data (alternative field names)",
          );
          this.storageService.storeUserEncryptedData(
            salt,
            encryptedMasterKey,
            encryptedPrivateKey,
            publicKey,
          );
        }
      }

      console.log("[AuthManager] OTT verification flow completed successfully");
      return response;
    } catch (error) {
      console.error("[AuthManager] OTT verification flow failed:", error);
      throw error;
    }
  }

  // Step 3: Complete Login with Token Decryption and Storage
  async completeLogin(email, challengeId, decryptedChallenge) {
    try {
      console.log("[AuthManager] Orchestrating login completion for:", email);

      const response = await this.apiService.completeLogin(
        email,
        challengeId,
        decryptedChallenge,
      );

      console.log(
        "[AuthManager] Login API call successful, processing tokens...",
      );

      // Handle encrypted tokens from backend - decrypt and store unencrypted
      if (
        (response.encrypted_access_token &&
          response.encrypted_refresh_token &&
          response.token_nonce) ||
        (response.encrypted_tokens && response.token_nonce)
      ) {
        console.log(
          "[AuthManager] Received encrypted tokens - orchestrating decryption...",
        );

        // Check if we have session keys for decryption
        if (!this.storageService.hasSessionKeys()) {
          throw new Error("No session keys available for token decryption");
        }

        let decryptedTokens;

        if (
          response.encrypted_access_token &&
          response.encrypted_refresh_token
        ) {
          // Validate encrypted tokens before attempting decryption
          if (
            !response.encrypted_access_token ||
            !response.encrypted_refresh_token
          ) {
            throw new Error("Missing encrypted tokens in response");
          }

          // Handle separate encrypted tokens
          console.log(
            "[AuthManager] Orchestrating separate access and refresh token decryption",
          );

          const accessTokenData =
            await this.storageService.decryptTokensFromLogin(
              response.encrypted_access_token,
              response.token_nonce,
            );

          const refreshTokenData =
            await this.storageService.decryptTokensFromLogin(
              response.encrypted_refresh_token,
              response.token_nonce,
            );

          // Handle both string and object responses
          decryptedTokens = {
            access_token:
              typeof accessTokenData === "string"
                ? accessTokenData
                : accessTokenData.access_token || accessTokenData,
            refresh_token:
              typeof refreshTokenData === "string"
                ? refreshTokenData
                : refreshTokenData.refresh_token || refreshTokenData,
          };
        } else {
          // Handle single encrypted_tokens field containing both tokens
          console.log(
            "[AuthManager] Orchestrating combined token blob decryption",
          );

          decryptedTokens = await this.storageService.decryptTokensFromLogin(
            response.encrypted_tokens,
            response.token_nonce,
          );
        }

        console.log("[AuthManager] Token decryption successful");

        // Store unencrypted tokens
        this.storageService.setTokens(
          decryptedTokens.access_token,
          decryptedTokens.refresh_token,
          response.access_token_expiry_date,
          response.refresh_token_expiry_date,
        );

        console.log("[AuthManager] Unencrypted tokens stored successfully");
      } else {
        // Fallback: handle already unencrypted tokens (for testing/legacy)
        console.log("[AuthManager] Received unencrypted tokens");

        if (response.access_token && response.refresh_token) {
          this.storageService.setTokens(
            response.access_token,
            response.refresh_token,
            response.access_token_expiry_date,
            response.refresh_token_expiry_date,
          );
        } else {
          throw new Error("No valid tokens received from login response");
        }
      }

      // Store username/email
      if (response.username) {
        this.storageService.setUserEmail(response.username);
      }

      // Clear login session data but NOT session keys
      this.storageService.clearAllLoginSessionData();

      // Clear session keys after successful login
      console.log(
        "[AuthManager] Clearing session keys after login - they will be re-derived from password when needed",
      );
      this.storageService.clearSessionKeys();

      // Clean up any old encrypted token data
      this.storageService.cleanupEncryptedTokenData();

      console.log(
        "[AuthManager] Login flow completed successfully with unencrypted tokens",
      );
      return response;
    } catch (error) {
      console.error("[AuthManager] Login completion flow failed:", error);
      // Clear session keys on error
      this.storageService.clearSessionKeys();
      throw error;
    }
  }

  // === Challenge Decryption ===

  async decryptChallenge(password, verifyData) {
    try {
      console.log("[AuthManager] Orchestrating challenge decryption");
      console.log(
        "[AuthManager] Available verify data fields:",
        Object.keys(verifyData),
      );

      // Validate required data
      if (!verifyData) {
        throw new Error("No verification data provided");
      }

      // Common field name variations to check
      const fieldMappings = {
        salt: ["salt", "Salt", "password_salt"],
        encryptedMasterKey: [
          "encryptedMasterKey",
          "encrypted_master_key",
          "masterKey",
          "master_key",
        ],
        encryptedPrivateKey: [
          "encryptedPrivateKey",
          "encrypted_private_key",
          "privateKey",
          "private_key",
        ],
        encryptedChallenge: [
          "encryptedChallenge",
          "encrypted_challenge",
          "challenge",
        ],
        publicKey: [
          "publicKey",
          "public_key",
          "userPublicKey",
          "user_public_key",
        ],
      };

      const challengeData = {};

      // Map fields to their actual names in the response
      for (const [expectedField, possibleNames] of Object.entries(
        fieldMappings,
      )) {
        let found = false;
        for (const possibleName of possibleNames) {
          if (verifyData[possibleName]) {
            challengeData[expectedField] = verifyData[possibleName];
            console.log(
              `[AuthManager] Mapped ${expectedField} -> ${possibleName}`,
            );
            found = true;
            break;
          }
        }
        if (!found && expectedField !== "publicKey") {
          console.error(
            `[AuthManager] Could not find field for ${expectedField}`,
          );
        }
      }

      // Check if we have all required data
      const missingFields = [];
      const requiredFields = [
        "salt",
        "encryptedMasterKey",
        "encryptedPrivateKey",
        "encryptedChallenge",
      ];

      for (const field of requiredFields) {
        if (!challengeData[field]) {
          missingFields.push(field);
        }
      }

      if (missingFields.length > 0) {
        throw new Error(
          `Missing required encrypted data: ${missingFields.join(", ")}`,
        );
      }

      console.log("[AuthManager] Successfully mapped all required fields");

      // Use CryptoService to decrypt the challenge
      const decryptedChallenge = await CryptoService.decryptLoginChallenge(
        password,
        challengeData,
      );

      // After successful decryption, cache the keys for token decryption
      console.log(
        "[AuthManager] Orchestrating session key caching for token decryption",
      );

      // Re-derive the keys to cache them
      await CryptoService.initialize();

      // Decode the encrypted data
      const salt = CryptoService.tryDecodeBase64(challengeData.salt);
      const encryptedMasterKey = CryptoService.tryDecodeBase64(
        challengeData.encryptedMasterKey,
      );
      const encryptedPrivateKey = CryptoService.tryDecodeBase64(
        challengeData.encryptedPrivateKey,
      );
      const publicKey = challengeData.publicKey
        ? CryptoService.tryDecodeBase64(challengeData.publicKey)
        : null;

      // Derive the key encryption key
      const keyEncryptionKey = await CryptoService.deriveKeyFromPassword(
        password,
        salt,
      );

      // Decrypt the master key
      const masterKey = CryptoService.decryptWithSecretBox(
        encryptedMasterKey,
        keyEncryptionKey,
      );

      // Decrypt the private key
      const privateKey = CryptoService.decryptWithSecretBox(
        encryptedPrivateKey,
        masterKey,
      );

      // Derive public key if not provided
      const derivedPublicKey =
        publicKey || CryptoService.sodium.crypto_scalarmult_base(privateKey);

      // Cache the keys in storage service for token decryption during login
      this.storageService.setSessionKeys(
        masterKey, // decrypted master key
        privateKey, // decrypted private key
        derivedPublicKey, // derived public key
        keyEncryptionKey, // derived from password
      );

      // NEW: Also cache the derived public key separately for future token refreshes
      // This is safe because the public key is not sensitive
      this.storageService.storeDerivedPublicKey(derivedPublicKey);

      console.log(
        "[AuthManager] Challenge decryption orchestrated successfully and keys cached",
      );
      return decryptedChallenge;
    } catch (error) {
      console.error(
        "[AuthManager] Challenge decryption orchestration failed:",
        error,
      );
      throw new Error(`Decryption failed: ${error.message}`);
    }
  }

  // === Token Management ===

  // Token refresh is now handled by ApiClient automatically
  async refreshToken() {
    try {
      console.log("[AuthManager] Delegating token refresh to ApiClient");
      // Import ApiClient to use its refresh functionality
      const { default: ApiClient } = await import("../API/ApiClient.js");
      return await ApiClient.refreshTokens();
    } catch (error) {
      console.error("[AuthManager] Token refresh delegation failed:", error);
      throw error;
    }
  }

  // Manual token refresh (delegated to ApiClient)
  async refreshTokenViaWorker() {
    console.log("[AuthManager] Manual refresh delegated to ApiClient");
    return await this.refreshToken();
  }

  // Force a token check (no-op since handled by ApiClient interceptors)
  forceTokenCheck() {
    console.log(
      "[AuthManager] Force token check - handled by ApiClient interceptors",
    );
  }

  // === Registration and Email Verification ===

  async registerUser(userData) {
    try {
      console.log("[AuthManager] Orchestrating user registration");
      return await this.apiService.registerUser(userData);
    } catch (error) {
      console.error("[AuthManager] Registration orchestration failed:", error);
      throw error;
    }
  }

  async verifyEmail(verificationCode) {
    try {
      console.log("[AuthManager] Orchestrating email verification");
      return await this.apiService.verifyEmail(verificationCode);
    } catch (error) {
      console.error(
        "[AuthManager] Email verification orchestration failed:",
        error,
      );
      throw error;
    }
  }

  // === Authentication State Methods ===

  isAuthenticated() {
    return this.storageService.isAuthenticated();
  }

  getCurrentUserEmail() {
    return this.storageService.getUserEmail();
  }

  getAccessToken() {
    return this.storageService.getAccessToken();
  }

  shouldRefreshTokens() {
    return this.storageService.shouldRefreshTokens(5); // 5 minutes threshold
  }

  canMakeAuthenticatedRequests() {
    return this.storageService.canMakeAuthenticatedRequests();
  }

  getSessionKeyStatus() {
    return this.storageService.getSessionKeyStatus();
  }

  // === Logout and Cleanup ===

  logout() {
    console.log("[AuthManager] Orchestrating user logout");

    // Clear all authentication data
    this.storageService.clearAuthData();

    // Notify listeners of logout
    this.notifyAuthStateChange("force_logout", {
      reason: "manual_logout",
    });
  }

  // === Utility Methods ===

  async generateE2EEData(password) {
    return await CryptoService.generateE2EEData(password);
  }

  generateVerificationID(publicKey) {
    return CryptoService.generateVerificationID(publicKey);
  }

  validateMnemonic(mnemonic) {
    return CryptoService.validateMnemonic(mnemonic);
  }

  // === Status and Debug Methods ===

  async getWorkerStatus() {
    return {
      isInitialized: this.isInitialized,
      method: "api_interceptor",
      hasWorker: false,
    };
  }

  getDebugInfo() {
    return {
      serviceName: "AuthManager",
      role: "orchestrator",
      isInitialized: this.isInitialized,
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      authenticationState: {
        isAuthenticated: this.isAuthenticated(),
        userEmail: this.getCurrentUserEmail(),
        canMakeRequests: this.canMakeAuthenticatedRequests(),
        shouldRefreshTokens: this.shouldRefreshTokens(),
        sessionKeyStatus: this.getSessionKeyStatus(),
      },
      eventListeners: {
        count: this.authStateListeners.size,
      },
    };
  }
}

export default AuthManager;
