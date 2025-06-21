// monorepo/web/prototype/maplefile-login/src/services/authService.jsx

// Authentication Service for API calls - Updated for New Token System
import LocalStorageService from "./localStorageService.jsx";
import CryptoService from "./cryptoService.jsx";
import workerManager from "./workerManager.jsx";

const API_BASE_URL = "/iam/api/v1"; // Using proxy from vite config

class AuthService {
  constructor() {
    this.isInitialized = false;
    this.workerDisabled = false;
    this.initializeWorker();
  }

  // Initialize the background worker
  async initializeWorker() {
    if (this.isInitialized) return;

    try {
      console.log("[AuthService] Initializing background worker...");
      await workerManager.initialize();
      this.isInitialized = true;
      console.log("[AuthService] Background worker initialized successfully");

      // Start monitoring if authenticated
      if (this.isAuthenticated()) {
        console.log("[AuthService] Starting token monitoring");
        workerManager.startMonitoring();
      }
    } catch (error) {
      console.error("[AuthService] Failed to initialize worker:", error);
      console.warn(
        "[AuthService] Continuing without background worker - token refresh will be manual only",
      );

      // Set a flag to indicate we're running without worker
      this.isInitialized = false;
      this.workerDisabled = true;

      // Don't throw the error - continue without worker
    }
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

      return response;
    } catch (error) {
      throw new Error(`Failed to verify OTT: ${error.message}`);
    }
  }

  // Step 3: Complete Login with Real Decryption
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

      // Handle encrypted tokens (new system)
      if (
        response.encrypted_access_token &&
        response.encrypted_refresh_token &&
        response.token_nonce
      ) {
        console.log(
          "[AuthService] Received encrypted access and refresh tokens",
        );

        // Store encrypted tokens separately as they come from backend
        LocalStorageService.setEncryptedAccessToken(
          response.encrypted_access_token,
          response.access_token_expiry_date,
        );

        LocalStorageService.setEncryptedRefreshToken(
          response.encrypted_refresh_token,
          response.refresh_token_expiry_date,
        );

        LocalStorageService.setTokenNonce(response.token_nonce);

        console.log("[AuthService] Encrypted tokens stored successfully");
      } else if (response.encrypted_tokens && response.token_nonce) {
        // Fallback: handle single encrypted_tokens field (if backend changes)
        console.log("[AuthService] Received single encrypted tokens field");

        LocalStorageService.setEncryptedTokens(
          response.encrypted_tokens,
          response.token_nonce,
          response.access_token_expiry_date,
          response.refresh_token_expiry_date,
        );

        console.log(
          "[AuthService] Single encrypted tokens stored successfully",
        );
      } else {
        // This should not happen with the new system, but handle gracefully
        console.warn(
          "[AuthService] No encrypted tokens received - missing fields:",
        );
        console.warn(
          "- encrypted_access_token:",
          !!response.encrypted_access_token,
        );
        console.warn(
          "- encrypted_refresh_token:",
          !!response.encrypted_refresh_token,
        );
        console.warn("- token_nonce:", !!response.token_nonce);
        console.warn("- encrypted_tokens:", !!response.encrypted_tokens);
        throw new Error(
          "Authentication system error: No encrypted tokens received",
        );
      }

      // Store username/email
      if (response.username) {
        LocalStorageService.setUserEmail(response.username);
      }

      // Clear login session data
      LocalStorageService.clearAllLoginSessionData();

      // Start background monitoring after successful login
      if (this.isInitialized) {
        workerManager.startMonitoring();
      }

      return response;
    } catch (error) {
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
      console.log("[AuthService] Verify data:", verifyData);

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
          // publicKey is optional
          console.error(
            `[AuthService] Could not find field for ${expectedField}`,
          );
          console.error(
            `[AuthService] Looked for: ${possibleNames.join(", ")}`,
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
        console.error("[AuthService] Missing required fields:", missingFields);
        console.error(
          "[AuthService] Available verify data:",
          Object.keys(verifyData),
        );
        console.error(
          "[AuthService] Mapped challenge data:",
          Object.keys(challengeData),
        );
        throw new Error(
          `Missing required encrypted data: ${missingFields.join(", ")}`,
        );
      }

      console.log("[AuthService] Successfully mapped all required fields");
      if (challengeData.publicKey) {
        console.log("[AuthService] Public key also available for verification");
      }

      // Use CryptoService to decrypt the challenge
      const decryptedChallenge = await CryptoService.decryptLoginChallenge(
        password,
        challengeData,
      );

      console.log("[AuthService] Challenge decryption successful");
      return decryptedChallenge;
    } catch (error) {
      console.error("[AuthService] Challenge decryption failed:", error);
      throw new Error(`Challenge decryption failed: ${error.message}`);
    }
  }

  // New Token Refresh using the updated API
  async refreshToken() {
    try {
      console.log("[AuthService] Starting token refresh with new API");

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

      // Handle the new encrypted token response
      if (response.encrypted_tokens && response.token_nonce) {
        console.log("[AuthService] Received new encrypted tokens");

        // Store the new encrypted tokens
        LocalStorageService.setEncryptedTokens(
          response.encrypted_tokens,
          response.token_nonce,
          response.access_token_expiry_date,
          response.refresh_token_expiry_date,
        );

        // Update user email if provided
        if (response.username) {
          LocalStorageService.setUserEmail(response.username);
        }

        console.log("[AuthService] New encrypted tokens stored successfully");
        return response;
      } else {
        // This should not happen with the new system
        console.error("[AuthService] No encrypted tokens in refresh response");
        throw new Error("Token refresh failed: No encrypted tokens received");
      }
    } catch (error) {
      console.error("[AuthService] Token refresh failed:", error);

      // Clear tokens on refresh failure
      LocalStorageService.clearAuthData();

      // Stop monitoring
      if (this.isInitialized) {
        workerManager.stopMonitoring();
      }

      throw new Error(`Failed to refresh tokens: ${error.message}`);
    }
  }

  // Refresh tokens using background worker (updated for new system)
  async refreshTokenViaWorker() {
    if (!this.isInitialized) {
      await this.initializeWorker();
    }

    try {
      console.log("[AuthService] Requesting token refresh via worker");
      const result = await workerManager.manualRefresh();
      console.log("[AuthService] Worker token refresh successful");
      return result;
    } catch (error) {
      console.error("[AuthService] Worker token refresh failed:", error);
      throw new Error(`Failed to refresh tokens via worker: ${error.message}`);
    }
  }

  // Force a token check (useful for testing)
  forceTokenCheck() {
    if (this.isInitialized) {
      workerManager.forceTokenCheck();
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
      return await workerManager.getWorkerStatus();
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
      workerManager.stopMonitoring();
    }

    // Clear all authentication data
    LocalStorageService.clearAuthData();
    LocalStorageService.clearAllLoginSessionData();
  }

  // Check if user is authenticated
  isAuthenticated() {
    return LocalStorageService.isAuthenticated();
  }

  // Get current user email
  getCurrentUserEmail() {
    return LocalStorageService.getUserEmail();
  }

  // Generate verification ID from public key (utility method)
  async generateVerificationID(publicKey) {
    return await CryptoService.generateVerificationID(publicKey);
  }

  // Validate BIP39 mnemonic (utility method)
  validateMnemonic(mnemonic) {
    return CryptoService.validateMnemonic(mnemonic);
  }

  // Decrypt access token for API calls (when needed)
  async getDecryptedAccessToken() {
    return await LocalStorageService.getDecryptedAccessToken();
  }

  // Check if tokens need refresh
  shouldRefreshTokens() {
    return LocalStorageService.isAccessTokenExpiringSoon(5); // 5 minutes threshold
  }
}

// Export singleton instance
export default new AuthService();
