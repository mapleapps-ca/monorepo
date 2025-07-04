// File: monorepo/web/maplefile-frontend/src/services/AuthService.js
// Main Authentication Service - Orchestrates API and Storage services
import AuthAPIService from "./API/AuthAPIService.js";
import AuthStorageService from "./Storage/AuthStorageService.js";
import CryptoService from "./Crypto/CryptoService.js";
import WorkerManager from "./WorkerManager.js";

class AuthService {
  constructor() {
    this.isInitialized = false;

    // Initialize dependent services
    this.apiService = new AuthAPIService();
    this.storageService = new AuthStorageService();

    console.log("[AuthService] Main authentication service initialized");
  }

  // Initialize the service (simplified without workers)
  async initializeWorker() {
    if (this.isInitialized) return;

    try {
      console.log("[AuthService] Initializing auth service (no workers)...");
      await WorkerManager.initialize();
      this.isInitialized = true;
      console.log("[AuthService] Auth service initialized successfully");
    } catch (error) {
      console.error("[AuthService] Failed to initialize:", error);
      this.isInitialized = true; // Continue anyway
    }
  }

  // === Authentication Flow Methods ===

  // Step 1: Request One-Time Token (OTT)
  async requestOTT(email) {
    try {
      console.log("[AuthService] Requesting OTT for:", email);

      const response = await this.apiService.requestOTT(email);

      // Store email for next steps
      this.storageService.setUserEmail(email);

      console.log("[AuthService] OTT request completed successfully");
      return response;
    } catch (error) {
      console.error("[AuthService] OTT request failed:", error);
      throw error;
    }
  }

  // Step 2: Verify One-Time Token
  async verifyOTT(email, ott) {
    try {
      console.log("[AuthService] Verifying OTT for:", email);

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
          "[AuthService] Storing user's encrypted data for future use",
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
            "[AuthService] Storing user's encrypted data (alternative field names)",
          );
          this.storageService.storeUserEncryptedData(
            salt,
            encryptedMasterKey,
            encryptedPrivateKey,
            publicKey,
          );
        }
      }

      console.log("[AuthService] OTT verification completed successfully");
      return response;
    } catch (error) {
      console.error("[AuthService] OTT verification failed:", error);
      throw error;
    }
  }

  // Step 3: Complete Login with Token Decryption and Storage
  async completeLogin(email, challengeId, decryptedChallenge) {
    try {
      console.log("[AuthService] Completing login for:", email);

      const response = await this.apiService.completeLogin(
        email,
        challengeId,
        decryptedChallenge,
      );

      console.log(
        "[AuthService] Login API call successful, processing tokens...",
      );

      // Handle encrypted tokens from backend - decrypt and store unencrypted
      if (
        (response.encrypted_access_token &&
          response.encrypted_refresh_token &&
          response.token_nonce) ||
        (response.encrypted_tokens && response.token_nonce)
      ) {
        console.log("[AuthService] Received encrypted tokens - decrypting...");

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
            "[AuthService] Decrypting separate access and refresh tokens",
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
          console.log("[AuthService] Decrypting combined token blob");

          decryptedTokens = await this.storageService.decryptTokensFromLogin(
            response.encrypted_tokens,
            response.token_nonce,
          );
        }

        console.log("[AuthService] Token decryption successful");

        // Store unencrypted tokens
        this.storageService.setTokens(
          decryptedTokens.access_token,
          decryptedTokens.refresh_token,
          response.access_token_expiry_date,
          response.refresh_token_expiry_date,
        );

        console.log("[AuthService] Unencrypted tokens stored successfully");
      } else {
        // Fallback: handle already unencrypted tokens (for testing/legacy)
        console.log("[AuthService] Received unencrypted tokens");

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
        "[AuthService] Clearing session keys after login - they will be re-derived from password when needed",
      );
      this.storageService.clearSessionKeys();

      // Clean up any old encrypted token data
      this.storageService.cleanupEncryptedTokenData();

      console.log(
        "[AuthService] Login completed successfully with unencrypted tokens",
      );
      return response;
    } catch (error) {
      console.error("[AuthService] Login completion failed:", error);
      // Clear session keys on error
      this.storageService.clearSessionKeys();
      throw error;
    }
  }

  // === Challenge Decryption ===

  async decryptChallenge(password, verifyData) {
    try {
      console.log("[AuthService] Starting challenge decryption");
      console.log(
        "[AuthService] Available verify data fields:",
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
              `[AuthService] Mapped ${expectedField} -> ${possibleName}`,
            );
            found = true;
            break;
          }
        }
        if (!found && expectedField !== "publicKey") {
          console.error(
            `[AuthService] Could not find field for ${expectedField}`,
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

      console.log("[AuthService] Successfully mapped all required fields");

      // Use CryptoService to decrypt the challenge
      const decryptedChallenge = await CryptoService.decryptLoginChallenge(
        password,
        challengeData,
      );

      // After successful decryption, cache the keys for token decryption
      console.log("[AuthService] Caching session keys for token decryption");

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

      console.log(
        "[AuthService] Challenge decryption successful and session keys cached for token decryption",
      );
      return decryptedChallenge;
    } catch (error) {
      console.error("[AuthService] Challenge decryption failed:", error);
      throw new Error(`Decryption failed: ${error.message}`);
    }
  }

  // === Token Management ===

  // Token refresh is now handled by ApiClient automatically
  async refreshToken() {
    try {
      console.log("[AuthService] Delegating token refresh to ApiClient");
      // Import ApiClient to use its refresh functionality
      const { default: ApiClient } = await import("./ApiClient.js");
      return await ApiClient.refreshTokens();
    } catch (error) {
      console.error("[AuthService] Token refresh failed:", error);
      throw error;
    }
  }

  // Manual token refresh (delegated to ApiClient)
  async refreshTokenViaWorker() {
    console.log("[AuthService] Manual refresh delegated to ApiClient");
    return await this.refreshToken();
  }

  // Force a token check (no-op since handled by ApiClient interceptors)
  forceTokenCheck() {
    console.log(
      "[AuthService] Force token check - handled by ApiClient interceptors",
    );
  }

  // === Registration and Email Verification ===

  async registerUser(userData) {
    try {
      console.log("[AuthService] Registering user");
      return await this.apiService.registerUser(userData);
    } catch (error) {
      console.error("[AuthService] Registration failed:", error);
      throw error;
    }
  }

  async verifyEmail(verificationCode) {
    try {
      console.log("[AuthService] Verifying email");
      return await this.apiService.verifyEmail(verificationCode);
    } catch (error) {
      console.error("[AuthService] Email verification failed:", error);
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
    console.log("[AuthService] Logging out user");

    // Clear all authentication data
    this.storageService.clearAuthData();

    // Notify listeners of logout
    WorkerManager.notifyAuthStateChange("force_logout", {
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
      serviceName: "AuthService",
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
    };
  }
}

// Export singleton instance
export default new AuthService();
