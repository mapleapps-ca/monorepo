// Local Storage Service for managing unencrypted authentication tokens
const LOCAL_STORAGE_KEYS = {
  ACCESS_TOKEN: "mapleapps_access_token",
  REFRESH_TOKEN: "mapleapps_refresh_token",
  ACCESS_TOKEN_EXPIRY: "mapleapps_access_token_expiry",
  REFRESH_TOKEN_EXPIRY: "mapleapps_refresh_token_expiry",
  USER_EMAIL: "mapleapps_user_email",

  // Deprecated encrypted token keys (for cleanup)
  ENCRYPTED_ACCESS_TOKEN: "mapleapps_encrypted_access_token",
  ENCRYPTED_REFRESH_TOKEN: "mapleapps_encrypted_refresh_token",
  TOKEN_NONCE: "mapleapps_token_nonce",
  ENCRYPTED_TOKENS: "mapleapps_encrypted_tokens",
};

class LocalStorageService {
  // Store unencrypted access token with expiry
  setAccessToken(accessToken, accessTokenExpiry) {
    if (accessToken) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN, accessToken);
      console.log("[LocalStorageService] Access token stored");
    }

    if (accessTokenExpiry) {
      localStorage.setItem(
        LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
        accessTokenExpiry,
      );
    }
  }

  // Store unencrypted refresh token with expiry
  setRefreshToken(refreshToken, refreshTokenExpiry) {
    if (refreshToken) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN, refreshToken);
      console.log("[LocalStorageService] Refresh token stored");
    }

    if (refreshTokenExpiry) {
      localStorage.setItem(
        LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
        refreshTokenExpiry,
      );
    }
  }

  // Store both tokens with expiry times
  setTokens(accessToken, refreshToken, accessTokenExpiry, refreshTokenExpiry) {
    this.setAccessToken(accessToken, accessTokenExpiry);
    this.setRefreshToken(refreshToken, refreshTokenExpiry);
    console.log("[LocalStorageService] Both tokens stored successfully");
  }

  // Get unencrypted access token
  getAccessToken() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN);
  }

  // Get unencrypted refresh token
  getRefreshToken() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN);
  }

  // Set user email
  setUserEmail(email) {
    if (email) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.USER_EMAIL, email);
    }
  }

  // Get user email
  getUserEmail() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.USER_EMAIL);
  }

  // Check if access token is expired
  isAccessTokenExpired() {
    const expiryTime = localStorage.getItem(
      LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
    );
    if (!expiryTime) return true;

    return new Date() >= new Date(expiryTime);
  }

  // Check if access token is expiring soon (within specified minutes)
  isAccessTokenExpiringSoon(minutesBeforeExpiry = 5) {
    const expiryTime = localStorage.getItem(
      LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
    );
    if (!expiryTime) return true;

    const expiry = new Date(expiryTime);
    const now = new Date();
    const timeUntilExpiry = expiry.getTime() - now.getTime();
    const warningThreshold = minutesBeforeExpiry * 60 * 1000; // Convert to milliseconds

    return timeUntilExpiry <= warningThreshold;
  }

  // Check if refresh token is expired
  isRefreshTokenExpired() {
    const expiryTime = localStorage.getItem(
      LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
    );
    if (!expiryTime) return true;

    return new Date() >= new Date(expiryTime);
  }

  // Check if user is authenticated (has valid tokens)
  isAuthenticated() {
    const accessToken = this.getAccessToken();
    const refreshToken = this.getRefreshToken();

    if (!accessToken || !refreshToken) return false;

    // If access token is valid, user is authenticated
    if (!this.isAccessTokenExpired()) return true;

    // If access token is expired but refresh token is valid, we can refresh
    if (!this.isRefreshTokenExpired()) return true;

    return false;
  }

  // Check if we have valid tokens
  hasValidTokens() {
    return this.isAuthenticated();
  }

  // Get token expiry information
  getTokenExpiryInfo() {
    return {
      accessTokenExpiry: localStorage.getItem(
        LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
      ),
      refreshTokenExpiry: localStorage.getItem(
        LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
      ),
      accessTokenExpired: this.isAccessTokenExpired(),
      refreshTokenExpired: this.isRefreshTokenExpired(),
      accessTokenExpiringSoon: this.isAccessTokenExpiringSoon(5),
    };
  }

  // Clear all authentication data
  clearAuthData() {
    // Clear unencrypted tokens
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.USER_EMAIL);

    // Also clear any leftover encrypted token data
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.TOKEN_NONCE);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS);

    // Clear session keys
    this.clearSessionKeys();

    console.log("[LocalStorageService] All authentication data cleared");
  }

  // Session-based key storage for login process (in memory only)
  _sessionKeys = {
    masterKey: null,
    privateKey: null,
    publicKey: null,
    keyEncryptionKey: null,
  };

  // Store session keys during login process (in memory only)
  setSessionKeys(masterKey, privateKey, publicKey, keyEncryptionKey) {
    this._sessionKeys = {
      masterKey,
      privateKey,
      publicKey,
      keyEncryptionKey,
    };
    console.log("[LocalStorageService] Session keys cached for login process");
  }

  // Clear session keys after login complete
  clearSessionKeys() {
    this._sessionKeys = {
      masterKey: null,
      privateKey: null,
      publicKey: null,
      keyEncryptionKey: null,
    };
    console.log("[LocalStorageService] Session keys cleared");
  }

  // Check if we have session keys (only needed during login)
  hasSessionKeys() {
    return !!(this._sessionKeys.masterKey && this._sessionKeys.privateKey);
  }

  // Get session keys (only used during login to decrypt tokens)
  getSessionKeys() {
    return this._sessionKeys;
  }

  // Store login session data (for multi-step login)
  setLoginSessionData(key, data) {
    localStorage.setItem(`login_session_${key}`, JSON.stringify(data));
  }

  // Get login session data
  getLoginSessionData(key) {
    const data = localStorage.getItem(`login_session_${key}`);
    return data ? JSON.parse(data) : null;
  }

  // Clear login session data
  clearLoginSessionData(key) {
    localStorage.removeItem(`login_session_${key}`);
  }

  // Clear all login session data
  clearAllLoginSessionData() {
    const keys = Object.keys(localStorage);
    keys.forEach((key) => {
      if (key.startsWith("login_session_")) {
        localStorage.removeItem(key);
      }
    });
  }

  // Get all storage data for worker communication
  getAllStorageData() {
    return {
      accessToken: this.getAccessToken(),
      refreshToken: this.getRefreshToken(),
      accessTokenExpiry: localStorage.getItem(
        LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
      ),
      refreshTokenExpiry: localStorage.getItem(
        LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
      ),
      userEmail: this.getUserEmail(),
      ...this.getTokenExpiryInfo(),
      hasValidTokens: this.hasValidTokens(),
    };
  }

  // Clean up any old encrypted token data (migration helper)
  cleanupEncryptedTokenData() {
    const hadEncryptedData = !!(
      localStorage.getItem(LOCAL_STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN) ||
      localStorage.getItem(LOCAL_STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN) ||
      localStorage.getItem(LOCAL_STORAGE_KEYS.TOKEN_NONCE) ||
      localStorage.getItem(LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS)
    );

    if (hadEncryptedData) {
      console.log("[LocalStorageService] Cleaning up old encrypted token data");
      localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN);
      localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN);
      localStorage.removeItem(LOCAL_STORAGE_KEYS.TOKEN_NONCE);
      localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS);
      return true;
    }

    return false;
  }

  // Decrypt encrypted tokens using session keys (used during login)
  async decryptTokensFromLogin(encryptedTokensData, tokenNonce) {
    if (!this.hasSessionKeys()) {
      throw new Error("No session keys available for token decryption");
    }

    try {
      console.log(
        "[LocalStorageService] Decrypting tokens from login response",
      );
      console.log(
        "[LocalStorageService] Encrypted tokens length:",
        encryptedTokensData?.length,
      );
      console.log(
        "[LocalStorageService] Token nonce length:",
        tokenNonce?.length,
      );

      // Import the crypto service for decryption
      const { default: CryptoService } = await import("./CryptoService.js");
      await CryptoService.initialize();

      // Decode the nonce and encrypted data using tryDecodeBase64 to handle different encoding variants
      const nonce = CryptoService.tryDecodeBase64(tokenNonce);
      const encryptedData = CryptoService.tryDecodeBase64(encryptedTokensData);

      console.log("[LocalStorageService] Decoded nonce length:", nonce.length);
      console.log(
        "[LocalStorageService] Decoded encrypted data length:",
        encryptedData.length,
      );
      console.log("[LocalStorageService] Attempting token decryption...");

      // Try sealed box decryption first (anonymous encryption)
      let decryptedTokenData;
      try {
        console.log(
          "[LocalStorageService] Trying sealed box decryption with private key",
        );
        decryptedTokenData = await CryptoService.decryptChallenge(
          encryptedData,
          this._sessionKeys.privateKey,
          this._sessionKeys.publicKey,
        );
        console.log("[LocalStorageService] Tokens decrypted using sealed box");
      } catch (sealError) {
        console.log(
          "[LocalStorageService] Sealed box failed:",
          sealError.message,
        );
        console.log(
          "[LocalStorageService] Trying secretbox with master key...",
        );

        // Fallback: try secretbox decryption with master key
        try {
          decryptedTokenData = CryptoService.decryptWithSecretBox(
            encryptedData,
            this._sessionKeys.masterKey,
          );
          console.log(
            "[LocalStorageService] Tokens decrypted using secretbox with master key",
          );
        } catch (secretError) {
          console.log(
            "[LocalStorageService] Secretbox with master key failed:",
            secretError.message,
          );

          // Try with KEK as another option
          if (this._sessionKeys.keyEncryptionKey) {
            try {
              console.log("[LocalStorageService] Trying secretbox with KEK...");
              decryptedTokenData = CryptoService.decryptWithSecretBox(
                encryptedData,
                this._sessionKeys.keyEncryptionKey,
              );
              console.log(
                "[LocalStorageService] Tokens decrypted using secretbox with KEK",
              );
            } catch (kekError) {
              console.error(
                "[LocalStorageService] All decryption methods failed",
              );
              console.error("Sealed box error:", sealError.message);
              console.error("Master key error:", secretError.message);
              console.error("KEK error:", kekError.message);
              throw new Error(
                "Failed to decrypt tokens with any available keys",
              );
            }
          } else {
            console.error(
              "[LocalStorageService] Both decryption methods failed",
            );
            throw new Error("Failed to decrypt tokens with available keys");
          }
        }
      }

      // Convert decrypted data to string
      const tokenString = new TextDecoder().decode(decryptedTokenData);

      // Try to parse as JSON, but if it fails, assume it's a plain token string
      let result;
      try {
        result = JSON.parse(tokenString);
        console.log(
          "[LocalStorageService] Token decryption successful (JSON format)",
        );
        console.log(
          "[LocalStorageService] Decrypted token object keys:",
          Object.keys(result),
        );
      } catch (parseError) {
        // If JSON parsing fails, assume it's a plain token string
        console.log(
          "[LocalStorageService] Token decryption successful (plain string format)",
        );
        result = tokenString;
      }

      return result;
    } catch (error) {
      console.error("[LocalStorageService] Failed to decrypt tokens:", error);
      throw error;
    }
  }
}

// Export singleton instance
export default new LocalStorageService();
