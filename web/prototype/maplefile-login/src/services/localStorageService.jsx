// monorepo/web/prototype/maplefile-login/src/services/localStorageService.jsx

// Local Storage Service for managing encrypted authentication tokens
const LOCAL_STORAGE_KEYS = {
  ENCRYPTED_ACCESS_TOKEN: "mapleapps_encrypted_access_token",
  ENCRYPTED_REFRESH_TOKEN: "mapleapps_encrypted_refresh_token",
  TOKEN_NONCE: "mapleapps_token_nonce",
  ACCESS_TOKEN_EXPIRY: "mapleapps_access_token_expiry",
  REFRESH_TOKEN_EXPIRY: "mapleapps_refresh_token_expiry",
  USER_EMAIL: "mapleapps_user_email",
  // Legacy keys (will be removed)
  ACCESS_TOKEN: "mapleapps_access_token",
  REFRESH_TOKEN: "mapleapps_refresh_token",
  // Deprecated single encrypted tokens key
  ENCRYPTED_TOKENS: "mapleapps_encrypted_tokens",
};

class LocalStorageService {
  // Store encrypted access token with expiry
  setEncryptedAccessToken(encryptedAccessToken, accessTokenExpiry) {
    if (encryptedAccessToken) {
      localStorage.setItem(
        LOCAL_STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN,
        encryptedAccessToken,
      );

      if (accessTokenExpiry) {
        localStorage.setItem(
          LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
          accessTokenExpiry,
        );
      }

      console.log(
        "[LocalStorageService] Encrypted access token stored successfully",
      );
    }
  }

  // Store encrypted refresh token with expiry
  setEncryptedRefreshToken(encryptedRefreshToken, refreshTokenExpiry) {
    if (encryptedRefreshToken) {
      localStorage.setItem(
        LOCAL_STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN,
        encryptedRefreshToken,
      );

      if (refreshTokenExpiry) {
        localStorage.setItem(
          LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
          refreshTokenExpiry,
        );
      }

      console.log(
        "[LocalStorageService] Encrypted refresh token stored successfully",
      );
    }
  }

  // Store token nonce
  setTokenNonce(tokenNonce) {
    if (tokenNonce) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.TOKEN_NONCE, tokenNonce);
      console.log("[LocalStorageService] Token nonce stored successfully");
    }
  }

  // Store encrypted tokens with expiry times (legacy method for single encrypted_tokens field)
  setEncryptedTokens(
    encryptedTokens,
    tokenNonce,
    accessTokenExpiry,
    refreshTokenExpiry,
  ) {
    if (encryptedTokens && tokenNonce) {
      localStorage.setItem(
        LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS,
        encryptedTokens,
      );
      localStorage.setItem(LOCAL_STORAGE_KEYS.TOKEN_NONCE, tokenNonce);

      if (accessTokenExpiry) {
        localStorage.setItem(
          LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
          accessTokenExpiry,
        );
      }

      if (refreshTokenExpiry) {
        localStorage.setItem(
          LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
          refreshTokenExpiry,
        );
      }

      console.log(
        "[LocalStorageService] Encrypted tokens stored successfully (legacy format)",
      );
    }
  }

  // Get encrypted access token
  getEncryptedAccessToken() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN);
  }

  // Get encrypted refresh token
  getEncryptedRefreshToken() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN);
  }

  // Get encrypted tokens blob (legacy method)
  getEncryptedTokens() {
    // First try the new separate tokens
    const encryptedAccessToken = this.getEncryptedAccessToken();
    const encryptedRefreshToken = this.getEncryptedRefreshToken();

    if (encryptedAccessToken && encryptedRefreshToken) {
      // For refresh API calls, we need to send the refresh token
      return encryptedRefreshToken;
    }

    // Fallback to legacy single encrypted_tokens field
    return localStorage.getItem(LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS);
  }

  // Get token nonce
  getTokenNonce() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.TOKEN_NONCE);
  }

  // Decrypt and get access token (placeholder - would need crypto implementation)
  async getDecryptedAccessToken() {
    const encryptedAccessToken = this.getEncryptedAccessToken();
    const encryptedTokens = this.getEncryptedTokens(); // fallback to legacy
    const tokenNonce = this.getTokenNonce();

    if ((encryptedAccessToken || encryptedTokens) && tokenNonce) {
      try {
        // TODO: Implement actual token decryption here
        // For now, we'll indicate that tokens are encrypted
        console.log(
          "[LocalStorageService] Encrypted tokens available but decryption not yet implemented",
        );

        // Return a placeholder that indicates we have encrypted tokens
        // In a real implementation, you would decrypt the tokens here using the user's private key
        return "encrypted_token_available";
      } catch (error) {
        console.error(
          "[LocalStorageService] Failed to decrypt access token:",
          error,
        );
        return null;
      }
    }

    console.warn("[LocalStorageService] No encrypted tokens available");
    return null;
  }

  // Get refresh token for refresh API calls
  getRefreshToken() {
    // For token refresh, we send the encrypted refresh token
    const encryptedRefreshToken = this.getEncryptedRefreshToken();

    if (encryptedRefreshToken) {
      console.log(
        "[LocalStorageService] Using encrypted refresh token for refresh",
      );
      return encryptedRefreshToken;
    }

    // Fallback: check for legacy single encrypted_tokens field
    const encryptedTokens = localStorage.getItem(
      LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS,
    );
    if (encryptedTokens) {
      console.log(
        "[LocalStorageService] Using legacy encrypted tokens for refresh",
      );
      return encryptedTokens;
    }

    // Last fallback: check for legacy refresh token
    const legacyRefreshToken = localStorage.getItem(
      LOCAL_STORAGE_KEYS.REFRESH_TOKEN,
    );
    if (legacyRefreshToken) {
      console.warn(
        "[LocalStorageService] Using legacy refresh token - this should be migrated",
      );
      return legacyRefreshToken;
    }

    return null;
  }

  // Legacy method for backward compatibility
  getAccessToken() {
    // First try to get decrypted access token
    const encryptedAccessToken = this.getEncryptedAccessToken();
    const encryptedTokens = this.getEncryptedTokens(); // fallback to legacy

    if (encryptedAccessToken || encryptedTokens) {
      // Return indicator that we have encrypted tokens
      return "encrypted_token_available";
    }

    // Fallback to legacy token
    return localStorage.getItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN);
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

  // Check if user is authenticated (has valid encrypted tokens)
  isAuthenticated() {
    // Check for new separate encrypted tokens first
    const encryptedAccessToken = this.getEncryptedAccessToken();
    const encryptedRefreshToken = this.getEncryptedRefreshToken();
    const tokenNonce = this.getTokenNonce();

    if (encryptedAccessToken && encryptedRefreshToken && tokenNonce) {
      // We have separate encrypted tokens, check if refresh token is still valid
      if (!this.isRefreshTokenExpired()) {
        console.log(
          "[LocalStorageService] Authenticated with separate encrypted tokens",
        );
        return true;
      }
    }

    // Fallback: check for legacy single encrypted_tokens field
    const encryptedTokens = localStorage.getItem(
      LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS,
    );
    if (encryptedTokens && tokenNonce) {
      if (!this.isRefreshTokenExpired()) {
        console.log(
          "[LocalStorageService] Authenticated with legacy encrypted tokens",
        );
        return true;
      }
    }

    // Last fallback: check legacy tokens
    const accessToken = localStorage.getItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN);
    const refreshToken = localStorage.getItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN);

    if (!accessToken || !refreshToken) return false;

    // If access token is valid, user is authenticated
    if (!this.isAccessTokenExpired()) return true;

    // If access token is expired but refresh token is valid, we can refresh
    if (!this.isRefreshTokenExpired()) return true;

    return false;
  }

  // Check if we have any form of valid tokens
  hasValidTokens() {
    return this.isAuthenticated();
  }

  // Check if we have encrypted tokens (new format)
  hasEncryptedTokens() {
    const encryptedAccessToken = this.getEncryptedAccessToken();
    const encryptedRefreshToken = this.getEncryptedRefreshToken();
    const encryptedTokens = localStorage.getItem(
      LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS,
    ); // legacy
    const tokenNonce = this.getTokenNonce();

    return !!(
      (encryptedAccessToken && encryptedRefreshToken && tokenNonce) ||
      (encryptedTokens && tokenNonce)
    );
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
    // Clear new separate encrypted tokens
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.TOKEN_NONCE);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.USER_EMAIL);

    // Clear legacy single encrypted tokens
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS);

    // Clear legacy plaintext tokens
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN);

    console.log("[LocalStorageService] All authentication data cleared");
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

  // Migrate legacy tokens to new system (utility method)
  migrateLegacyTokens() {
    const legacyAccessToken = localStorage.getItem(
      LOCAL_STORAGE_KEYS.ACCESS_TOKEN,
    );
    const legacyRefreshToken = localStorage.getItem(
      LOCAL_STORAGE_KEYS.REFRESH_TOKEN,
    );

    if (legacyAccessToken || legacyRefreshToken) {
      console.warn(
        "[LocalStorageService] Legacy tokens detected - user should re-authenticate with new system",
      );

      // Clear legacy tokens to force re-authentication
      localStorage.removeItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN);
      localStorage.removeItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN);

      return true; // Indicates migration occurred
    }

    return false; // No migration needed
  }

  // Get all storage data for worker communication
  getAllStorageData() {
    return {
      // New separate encrypted tokens
      encryptedAccessToken: this.getEncryptedAccessToken(),
      encryptedRefreshToken: this.getEncryptedRefreshToken(),
      tokenNonce: this.getTokenNonce(),
      accessTokenExpiry: localStorage.getItem(
        LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
      ),
      refreshTokenExpiry: localStorage.getItem(
        LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
      ),
      userEmail: this.getUserEmail(),

      // Legacy single encrypted tokens (for backward compatibility)
      encryptedTokens: localStorage.getItem(
        LOCAL_STORAGE_KEYS.ENCRYPTED_TOKENS,
      ),

      // Include token status
      ...this.getTokenExpiryInfo(),
      hasEncryptedTokens: this.hasEncryptedTokens(),
    };
  }

  // Set legacy access token (for compatibility during transition)
  setAccessToken(token, expiryTime) {
    console.warn(
      "[LocalStorageService] setAccessToken called - this method is deprecated in favor of encrypted tokens",
    );
    if (token) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN, token);
      if (expiryTime) {
        localStorage.setItem(
          LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
          expiryTime,
        );
      }
    }
  }

  // Set legacy refresh token (for compatibility during transition)
  setRefreshToken(token, expiryTime) {
    console.warn(
      "[LocalStorageService] setRefreshToken called - this method is deprecated in favor of encrypted tokens",
    );
    if (token) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN, token);
      if (expiryTime) {
        localStorage.setItem(
          LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
          expiryTime,
        );
      }
    }
  }
}

// Export singleton instance
export default new LocalStorageService();
