// File: monorepo/web/maplefile-frontend/src/service/AuthService.js
// Authentication Service for API calls - Updated to store encrypted user data
import LocalStorageService from "./LocalStorageService.js";
import CryptoService from "./Crypto/CryptoService.js";
import WorkerManager from "./WorkerManager.js";

const API_BASE_URL = "/iam/api/v1"; // Using proxy from vite config
class AuthService {
  constructor() {
    this.isInitialized = false;
    this.workerDisabled = false;
    this.initializationPromise = null;
  }

  async _doInitialize() {
    try {
      console.log("[AuthService] Initializing background worker...");
      await WorkerManager.initialize();
      this.isInitialized = true;
      console.log("[AuthService] Background worker initialized successfully");

      // Start monitoring if authenticated
      if (this.isAuthenticated()) {
        console.log("[AuthService] Starting token monitoring");
        WorkerManager.startMonitoring();
      }
    } catch (error) {
      console.error("[AuthService] Failed to initialize worker:", error);
      console.warn(
        "[AuthService] Continuing without background worker - token refresh will be manual only",
      );

      // Set a flag to indicate we're running without worker
      this.isInitialized = false;
      this.workerDisabled = true;
    }
  }

  // Initialize the background worker
  async initializeWorker() {
    if (this.isInitialized) return;

    // Prevent multiple simultaneous initializations
    if (this.initializationPromise) {
      return this.initializationPromise;
    }

    this.initializationPromise = this._doInitialize();
    return this.initializationPromise;
  }

  // Helper method to make API requests
  async makeRequest(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;

    const defaultOptions = {
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    };

    const requestOptions = {
      ...defaultOptions,
      ...options,
    };

    try {
      console.log(`Making ${requestOptions.method || "GET"} request to:`, url);
      const response = await fetch(url, requestOptions);

      const data = await response.json();

      if (!response.ok) {
        console.error("API Error:", data);
        throw new Error(
          data.details
            ? Object.values(data.details)[0]
            : data.error || "Request failed",
        );
      }

      console.log("API Response:", data);
      return data;
    } catch (error) {
      console.error("Request failed:", error);
      throw error;
    }
  }

  // Step 1: Request One-Time Token (OTT)
  async requestOTT(email) {
    try {
      const response = await this.makeRequest("/request-ott", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
        }),
      });

      // Store email for next steps
      LocalStorageService.setUserEmail(email);

      return response;
    } catch (error) {
      throw new Error(`Failed to request OTT: ${error.message}`);
    }
  }

  // Step 2: Verify One-Time Token
  async verifyOTT(email, ott) {
    try {
      const response = await this.makeRequest("/verify-ott", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
          ott: ott.trim(),
        }),
      });

      // Store verification data for the final step
      LocalStorageService.setLoginSessionData("verify_response", response);

      // IMPORTANT: Store the user's encrypted data for future password-based decryption
      if (
        response.salt &&
        response.encryptedMasterKey &&
        response.encryptedPrivateKey
      ) {
        console.log(
          "[AuthService] Storing user's encrypted data for future use",
        );
        LocalStorageService.storeUserEncryptedData(
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
          LocalStorageService.storeUserEncryptedData(
            salt,
            encryptedMasterKey,
            encryptedPrivateKey,
            publicKey,
          );
        }
      }

      return response;
    } catch (error) {
      throw new Error(`Failed to verify OTT: ${error.message}`);
    }
  }

  // Step 3: Complete Login with Token Decryption and Storage
  async completeLogin(email, challengeId, decryptedChallenge) {
    try {
      const response = await this.makeRequest("/complete-login", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
          challengeId: challengeId,
          decryptedData: decryptedChallenge,
        }),
      });

      console.log("[AuthService] Complete login response:", response);

      // Log the actual token data without truncation
      if (response.encrypted_access_token) {
        console.log(
          "[AuthService] Full encrypted_access_token:",
          response.encrypted_access_token,
        );
        console.log(
          "[AuthService] encrypted_access_token first 50 chars:",
          response.encrypted_access_token.substring(0, 50),
        );
        console.log(
          "[AuthService] encrypted_access_token last 50 chars:",
          response.encrypted_access_token.substring(
            response.encrypted_access_token.length - 50,
          ),
        );
      }

      if (response.encrypted_refresh_token) {
        console.log(
          "[AuthService] Full encrypted_refresh_token:",
          response.encrypted_refresh_token,
        );
      }

      if (response.token_nonce) {
        console.log("[AuthService] Full token_nonce:", response.token_nonce);
      }

      // Handle encrypted tokens from backend - decrypt and store unencrypted
      if (
        (response.encrypted_access_token &&
          response.encrypted_refresh_token &&
          response.token_nonce) ||
        (response.encrypted_tokens && response.token_nonce)
      ) {
        console.log("[AuthService] Received encrypted tokens - decrypting...");
        console.log(
          "[AuthService] Encrypted access token length:",
          response.encrypted_access_token?.length,
        );
        console.log(
          "[AuthService] Encrypted refresh token length:",
          response.encrypted_refresh_token?.length,
        );
        console.log(
          "[AuthService] Token nonce length:",
          response.token_nonce?.length,
        );

        // Check if we have session keys for decryption
        if (!LocalStorageService.hasSessionKeys()) {
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

          if (
            response.encrypted_access_token.includes("…") ||
            response.encrypted_refresh_token.includes("…")
          ) {
            console.error(
              "[AuthService] ERROR: Encrypted tokens appear to be truncated!",
            );
            console.error(
              "This might be a console display issue or actual data truncation",
            );
            throw new Error("Encrypted tokens appear to be truncated");
          }

          // Handle separate encrypted tokens
          console.log(
            "[AuthService] Decrypting separate access and refresh tokens",
          );

          const accessTokenData =
            await LocalStorageService.decryptTokensFromLogin(
              response.encrypted_access_token,
              response.token_nonce,
            );

          const refreshTokenData =
            await LocalStorageService.decryptTokensFromLogin(
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

          decryptedTokens = await LocalStorageService.decryptTokensFromLogin(
            response.encrypted_tokens,
            response.token_nonce,
          );
        }

        console.log("[AuthService] Token decryption successful");
        console.log(
          "[AuthService] Decrypted token keys:",
          Object.keys(decryptedTokens),
        );

        // Store unencrypted tokens in localStorage
        LocalStorageService.setTokens(
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
          LocalStorageService.setTokens(
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
        LocalStorageService.setUserEmail(response.username);
      }

      // Clear login session data but NOT session keys
      LocalStorageService.clearAllLoginSessionData();

      // IMPORTANT: Clear session keys after successful login
      // They were only needed temporarily for token decryption
      console.log(
        "[AuthService] Clearing session keys after login - they will be re-derived from password when needed",
      );
      LocalStorageService.clearSessionKeys();

      // Clean up any old encrypted token data
      LocalStorageService.cleanupEncryptedTokenData();

      // Start background monitoring after successful login
      if (this.isInitialized) {
        WorkerManager.startMonitoring();
      }

      console.log(
        "[AuthService] Login completed successfully with unencrypted tokens",
      );
      return response;
    } catch (error) {
      // Clear session keys on error
      LocalStorageService.clearSessionKeys();
      throw new Error(`Failed to complete login: ${error.message}`);
    }
  }

  // Real challenge decryption using CryptoService
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

      // We need to re-derive the keys to cache them
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

      // Cache the keys in LocalStorageService for token decryption during login
      LocalStorageService.setSessionKeys(
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

  // Token Refresh using unencrypted tokens
  async refreshToken() {
    try {
      console.log(
        "[AuthService] Starting token refresh with unencrypted tokens",
      );

      const refreshToken = LocalStorageService.getRefreshToken();
      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      // Use the new API endpoint format
      const response = await this.makeRequest("/token/refresh", {
        method: "POST",
        body: JSON.stringify({
          value: refreshToken,
        }),
      });

      console.log("[AuthService] Token refresh successful:", response);

      // Handle refreshed tokens - they might be encrypted or unencrypted
      if (response.encrypted_tokens && response.token_nonce) {
        console.log(
          "[AuthService] Received encrypted tokens from refresh - this shouldn't happen",
        );
        console.warn(
          "[AuthService] Backend should return unencrypted tokens after initial login",
        );
        throw new Error(
          "Unexpected encrypted tokens in refresh response - backend configuration issue",
        );
      } else if (response.access_token && response.refresh_token) {
        // Handle unencrypted tokens (expected)
        console.log("[AuthService] Received unencrypted tokens from refresh");

        LocalStorageService.setTokens(
          response.access_token,
          response.refresh_token,
          response.access_token_expiry_date,
          response.refresh_token_expiry_date,
        );

        // Update user email if provided
        if (response.username) {
          LocalStorageService.setUserEmail(response.username);
        }

        console.log("[AuthService] Refreshed tokens stored successfully");
        return response;
      } else {
        console.error("[AuthService] No valid tokens in refresh response");
        throw new Error("Token refresh failed: No valid tokens received");
      }
    } catch (error) {
      console.error("[AuthService] Token refresh failed:", error);

      // Clear tokens on refresh failure
      LocalStorageService.clearAuthData();

      // Stop monitoring
      if (this.isInitialized) {
        WorkerManager.stopMonitoring();
      }

      throw new Error(`Failed to refresh tokens: ${error.message}`);
    }
  }

  // Refresh tokens using background worker
  async refreshTokenViaWorker() {
    if (!this.isInitialized && !this.workerDisabled) {
      await this.initializeWorker();
    }

    if (this.workerDisabled) {
      console.warn("[AuthService] Worker disabled, using direct refresh");
      return this.refreshToken();
    }

    try {
      console.log("[AuthService] Requesting token refresh via worker");
      const result = await WorkerManager.manualRefresh();
      console.log("[AuthService] Worker token refresh successful");
      return result;
    } catch (error) {
      console.error("[AuthService] Worker token refresh failed:", error);
      throw new Error(`Failed to refresh tokens via worker: ${error.message}`);
    }
  }

  // Force a token check (useful for testing)
  forceTokenCheck() {
    if (this.isInitialized && !this.workerDisabled) {
      WorkerManager.forceTokenCheck();
    }
  }

  // Get worker status for debugging
  async getWorkerStatus() {
    if (this.workerDisabled) {
      return {
        isInitialized: false,
        workerDisabled: true,
        error: "Worker initialization failed - running in manual mode",
      };
    }

    if (!this.isInitialized) {
      return {
        isInitialized: false,
        workerDisabled: false,
        status: "not_initialized",
      };
    }

    try {
      return await WorkerManager.getWorkerStatus();
    } catch (error) {
      return {
        isInitialized: this.isInitialized,
        error: error.message,
      };
    }
  }

  // Logout and stop background monitoring
  logout() {
    console.log("[AuthService] Logging out user");

    // Stop background monitoring
    if (this.isInitialized) {
      WorkerManager.stopMonitoring();
    }

    // Clear all authentication data
    LocalStorageService.clearAuthData();
    LocalStorageService.clearAllLoginSessionData();
    LocalStorageService.clearSessionKeys();
  }

  // Check if user is authenticated
  isAuthenticated() {
    return LocalStorageService.isAuthenticated();
  }

  // Get current user email
  getCurrentUserEmail() {
    return LocalStorageService.getUserEmail();
  }

  // Get access token for API calls
  getAccessToken() {
    return LocalStorageService.getAccessToken();
  }

  // Check if tokens need refresh
  shouldRefreshTokens() {
    return LocalStorageService.isAccessTokenExpiringSoon(5); // 5 minutes threshold
  }

  // Check if we can make authenticated requests
  canMakeAuthenticatedRequests() {
    return LocalStorageService.hasValidTokens();
  }

  // Get session key status for debugging (only used during login)
  getSessionKeyStatus() {
    return {
      hasSessionKeys: LocalStorageService.hasSessionKeys(),
      hasUserEncryptedData: LocalStorageService.hasUserEncryptedData(),
      isAuthenticated: this.isAuthenticated(),
      canMakeRequests: this.canMakeAuthenticatedRequests(),
    };
  }

  // Registration method
  async registerUser(userData) {
    try {
      const url = `${API_BASE_URL}/register`;
      console.log("Making registration request to:", url);

      const response = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(userData),
      });

      console.log("Registration response status:", response.status);

      const result = await response.json();

      if (!response.ok) {
        console.error("Registration failed with status:", response.status);
        console.error("Error details:", result);
        throw new Error(
          result.details
            ? JSON.stringify(result.details)
            : result.error || "Registration failed",
        );
      }

      return result;
    } catch (error) {
      console.error("Registration error:", error);
      throw error;
    }
  }

  // Email verification method
  async verifyEmail(verificationCode) {
    try {
      const url = `${API_BASE_URL}/verify-email-code`;
      console.log("Making email verification request to:", url);

      const response = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          code: verificationCode.trim(),
        }),
      });

      const result = await response.json();

      if (!response.ok) {
        console.error(
          "Email verification failed with status:",
          response.status,
        );
        console.error("Error details:", result);
        throw new Error(
          result.details?.code || result.error || "Email verification failed",
        );
      }

      return result;
    } catch (error) {
      console.error("Email verification error:", error);
      throw error;
    }
  }

  // Generate E2EE data for registration
  async generateE2EEData(password) {
    return await CryptoService.generateE2EEData(password);
  }

  // Utility methods for debugging
  generateVerificationID(publicKey) {
    return CryptoService.generateVerificationID(publicKey);
  }

  validateMnemonic(mnemonic) {
    return CryptoService.validateMnemonic(mnemonic);
  }
}

// Export singleton instance
export default new AuthService();
