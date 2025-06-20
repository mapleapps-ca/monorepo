// Token Storage Service - Simplified for maplefile-refreshtoken prototype
// Handles encrypted token storage and management

const STORAGE_KEYS = {
  ENCRYPTED_ACCESS_TOKEN: "mapleapps_encrypted_access_token",
  ENCRYPTED_REFRESH_TOKEN: "mapleapps_encrypted_refresh_token",
  TOKEN_NONCE: "mapleapps_token_nonce",
  ACCESS_TOKEN_EXPIRY: "mapleapps_access_token_expiry",
  REFRESH_TOKEN_EXPIRY: "mapleapps_refresh_token_expiry",
  USER_EMAIL: "mapleapps_user_email",
  // Legacy single encrypted tokens field
  ENCRYPTED_TOKENS: "mapleapps_encrypted_tokens",
};

class TokenStorageService {
  // Store encrypted access token with expiry
  setEncryptedAccessToken(encryptedAccessToken, accessTokenExpiry) {
    if (encryptedAccessToken) {
      localStorage.setItem(
        STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN,
        encryptedAccessToken,
      );

      if (accessTokenExpiry) {
        localStorage.setItem(
          STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
          accessTokenExpiry,
        );
      }

      console.log(
        "[TokenStorageService] Encrypted access token stored successfully",
      );
    }
  }

  // Store encrypted refresh token with expiry
  setEncryptedRefreshToken(encryptedRefreshToken, refreshTokenExpiry) {
    if (encryptedRefreshToken) {
      localStorage.setItem(
        STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN,
        encryptedRefreshToken,
      );

      if (refreshTokenExpiry) {
        localStorage.setItem(
          STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
          refreshTokenExpiry,
        );
      }

      console.log(
        "[TokenStorageService] Encrypted refresh token stored successfully",
      );
    }
  }

  // Store token nonce
  setTokenNonce(tokenNonce) {
    if (tokenNonce) {
      localStorage.setItem(STORAGE_KEYS.TOKEN_NONCE, tokenNonce);
      console.log("[TokenStorageService] Token nonce stored successfully");
    }
  }

  // Store encrypted tokens (legacy single field format)
  setEncryptedTokens(
    encryptedTokens,
    tokenNonce,
    accessTokenExpiry,
    refreshTokenExpiry,
  ) {
    if (encryptedTokens && tokenNonce) {
      localStorage.setItem(STORAGE_KEYS.ENCRYPTED_TOKENS, encryptedTokens);
      localStorage.setItem(STORAGE_KEYS.TOKEN_NONCE, tokenNonce);

      if (accessTokenExpiry) {
        localStorage.setItem(
          STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
          accessTokenExpiry,
        );
      }

      if (refreshTokenExpiry) {
        localStorage.setItem(
          STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
          refreshTokenExpiry,
        );
      }

      console.log(
        "[TokenStorageService] Encrypted tokens stored successfully (legacy format)",
      );
    }
  }

  // Store user email
  setUserEmail(email) {
    if (email) {
      localStorage.setItem(STORAGE_KEYS.USER_EMAIL, email);
    }
  }

  // Get encrypted access token
  getEncryptedAccessToken() {
    return localStorage.getItem(STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN);
  }

  // Get encrypted refresh token
  getEncryptedRefreshToken() {
    return localStorage.getItem(STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN);
  }

  // Get token nonce
  getTokenNonce() {
    return localStorage.getItem(STORAGE_KEYS.TOKEN_NONCE);
  }

  // Get user email
  getUserEmail() {
    return localStorage.getItem(STORAGE_KEYS.USER_EMAIL);
  }

  // Get encrypted tokens (for refresh API calls)
  getEncryptedTokensForRefresh() {
    // First try the new separate refresh token
    const encryptedRefreshToken = this.getEncryptedRefreshToken();
    if (encryptedRefreshToken) {
      return encryptedRefreshToken;
    }

    // Fallback to legacy single encrypted_tokens field
    return localStorage.getItem(STORAGE_KEYS.ENCRYPTED_TOKENS);
  }

  // Check if access token is expired
  isAccessTokenExpired() {
    const expiryTime = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN_EXPIRY);
    if (!expiryTime) return true;
    return new Date() >= new Date(expiryTime);
  }

  // Check if access token is expiring soon
  isAccessTokenExpiringSoon(minutesBeforeExpiry = 5) {
    const expiryTime = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN_EXPIRY);
    if (!expiryTime) return true;

    const expiry = new Date(expiryTime);
    const now = new Date();
    const timeUntilExpiry = expiry.getTime() - now.getTime();
    const warningThreshold = minutesBeforeExpiry * 60 * 1000;

    return timeUntilExpiry <= warningThreshold;
  }

  // Check if refresh token is expired
  isRefreshTokenExpired() {
    const expiryTime = localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRY);
    if (!expiryTime) return true;
    return new Date() >= new Date(expiryTime);
  }

  // Check if user has encrypted tokens
  hasEncryptedTokens() {
    const encryptedAccessToken = this.getEncryptedAccessToken();
    const encryptedRefreshToken = this.getEncryptedRefreshToken();
    const encryptedTokens = localStorage.getItem(STORAGE_KEYS.ENCRYPTED_TOKENS);
    const tokenNonce = this.getTokenNonce();

    return !!(
      (encryptedAccessToken && encryptedRefreshToken && tokenNonce) ||
      (encryptedTokens && tokenNonce)
    );
  }

  // Check if user is authenticated
  isAuthenticated() {
    if (!this.hasEncryptedTokens()) {
      return false;
    }

    // User is authenticated if refresh token is still valid
    return !this.isRefreshTokenExpired();
  }

  // Get token expiry information
  getTokenExpiryInfo() {
    return {
      accessTokenExpiry: localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN_EXPIRY),
      refreshTokenExpiry: localStorage.getItem(
        STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
      ),
      accessTokenExpired: this.isAccessTokenExpired(),
      refreshTokenExpired: this.isRefreshTokenExpired(),
      accessTokenExpiringSoon: this.isAccessTokenExpiringSoon(5),
    };
  }

  // Get token format (separate vs legacy)
  getTokenFormat() {
    const encryptedAccessToken = this.getEncryptedAccessToken();
    const encryptedRefreshToken = this.getEncryptedRefreshToken();

    if (encryptedAccessToken && encryptedRefreshToken) {
      return "separate";
    }

    const legacyEncryptedTokens = localStorage.getItem(
      STORAGE_KEYS.ENCRYPTED_TOKENS,
    );
    if (legacyEncryptedTokens) {
      return "legacy";
    }

    return "none";
  }

  // Get all token information for debugging
  getTokenInfo() {
    return {
      hasEncryptedTokens: this.hasEncryptedTokens(),
      isAuthenticated: this.isAuthenticated(),
      tokenFormat: this.getTokenFormat(),
      userEmail: this.getUserEmail(),
      ...this.getTokenExpiryInfo(),
    };
  }

  // Clear all authentication data
  clearAllTokens() {
    Object.values(STORAGE_KEYS).forEach((key) => {
      localStorage.removeItem(key);
    });
    console.log("[TokenStorageService] All tokens cleared");
  }

  // Get all storage data (for worker communication)
  getAllStorageData() {
    return {
      [STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN]: this.getEncryptedAccessToken(),
      [STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN]: this.getEncryptedRefreshToken(),
      [STORAGE_KEYS.TOKEN_NONCE]: this.getTokenNonce(),
      [STORAGE_KEYS.ACCESS_TOKEN_EXPIRY]: localStorage.getItem(
        STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
      ),
      [STORAGE_KEYS.REFRESH_TOKEN_EXPIRY]: localStorage.getItem(
        STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
      ),
      [STORAGE_KEYS.USER_EMAIL]: this.getUserEmail(),
      [STORAGE_KEYS.ENCRYPTED_TOKENS]: localStorage.getItem(
        STORAGE_KEYS.ENCRYPTED_TOKENS,
      ),
    };
  }

  // Helper method to create demo tokens for testing
  createDemoTokens() {
    const now = new Date();
    const accessTokenExpiry = new Date(now.getTime() + 30 * 60 * 1000); // 30 minutes
    const refreshTokenExpiry = new Date(
      now.getTime() + 14 * 24 * 60 * 60 * 1000,
    ); // 14 days

    // Create fake encrypted tokens for demo purposes
    const fakeEncryptedAccessToken =
      "demo_encrypted_access_token_" + Math.random().toString(36).substring(7);
    const fakeEncryptedRefreshToken =
      "demo_encrypted_refresh_token_" + Math.random().toString(36).substring(7);
    const fakeTokenNonce =
      "demo_token_nonce_" + Math.random().toString(36).substring(7);

    this.setEncryptedAccessToken(
      fakeEncryptedAccessToken,
      accessTokenExpiry.toISOString(),
    );
    this.setEncryptedRefreshToken(
      fakeEncryptedRefreshToken,
      refreshTokenExpiry.toISOString(),
    );
    this.setTokenNonce(fakeTokenNonce);
    this.setUserEmail("demo@example.com");

    console.log("[TokenStorageService] Demo tokens created for testing");
    return {
      accessTokenExpiry: accessTokenExpiry.toISOString(),
      refreshTokenExpiry: refreshTokenExpiry.toISOString(),
    };
  }

  // Helper method to create demo tokens that expire soon (for testing refresh)
  createExpiringSoonTokens() {
    const now = new Date();
    const accessTokenExpiry = new Date(now.getTime() + 3 * 60 * 1000); // 3 minutes (will trigger refresh)
    const refreshTokenExpiry = new Date(
      now.getTime() + 14 * 24 * 60 * 60 * 1000,
    ); // 14 days

    const fakeEncryptedAccessToken =
      "demo_expiring_access_token_" + Math.random().toString(36).substring(7);
    const fakeEncryptedRefreshToken =
      "demo_refresh_token_" + Math.random().toString(36).substring(7);
    const fakeTokenNonce =
      "demo_token_nonce_" + Math.random().toString(36).substring(7);

    this.setEncryptedAccessToken(
      fakeEncryptedAccessToken,
      accessTokenExpiry.toISOString(),
    );
    this.setEncryptedRefreshToken(
      fakeEncryptedRefreshToken,
      refreshTokenExpiry.toISOString(),
    );
    this.setTokenNonce(fakeTokenNonce);
    this.setUserEmail("demo@example.com");

    console.log(
      "[TokenStorageService] Expiring soon demo tokens created for testing",
    );
    return {
      accessTokenExpiry: accessTokenExpiry.toISOString(),
      refreshTokenExpiry: refreshTokenExpiry.toISOString(),
    };
  }
}

// Export singleton instance
export default new TokenStorageService();
