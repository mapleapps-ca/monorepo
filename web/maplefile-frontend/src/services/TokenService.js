// Token Service for managing authentication tokens
import LocalStorageService from "./LocalStorageService.js";
import WorkerManager from "./WorkerManager.js";

class TokenService {
  constructor() {
    this.ACCESS_TOKEN_KEY = "access_token";
    this.REFRESH_TOKEN_KEY = "refresh_token";

    // New encrypted token keys
    this.ENCRYPTED_ACCESS_TOKEN_KEY = "mapleapps_encrypted_access_token";
    this.ENCRYPTED_REFRESH_TOKEN_KEY = "mapleapps_encrypted_refresh_token";
    this.TOKEN_NONCE_KEY = "mapleapps_token_nonce";
    this.ACCESS_TOKEN_EXPIRY_KEY = "mapleapps_access_token_expiry";
    this.REFRESH_TOKEN_EXPIRY_KEY = "mapleapps_refresh_token_expiry";
    this.USER_EMAIL_KEY = "mapleapps_user_email";
  }

  // Store encrypted access token
  setEncryptedAccessToken(token, expiry) {
    return LocalStorageService.setEncryptedAccessToken(token, expiry);
  }

  // Store encrypted refresh token
  setEncryptedRefreshToken(token, expiry) {
    return LocalStorageService.setEncryptedRefreshToken(token, expiry);
  }

  // Store token nonce
  setTokenNonce(nonce) {
    return LocalStorageService.setTokenNonce(nonce);
  }

  // Store encrypted tokens (legacy method)
  setEncryptedTokens(tokens, nonce, accessExpiry, refreshExpiry) {
    return LocalStorageService.setEncryptedTokens(
      tokens,
      nonce,
      accessExpiry,
      refreshExpiry,
    );
  }

  // Get encrypted access token
  getEncryptedAccessToken() {
    return LocalStorageService.getEncryptedAccessToken();
  }

  // Get encrypted refresh token
  getEncryptedRefreshToken() {
    return LocalStorageService.getEncryptedRefreshToken();
  }

  // Get encrypted tokens (for refresh calls)
  getEncryptedTokens() {
    return LocalStorageService.getEncryptedTokens();
  }

  // Get token nonce
  getTokenNonce() {
    return LocalStorageService.getTokenNonce();
  }

  // Get refresh token for API calls
  getRefreshToken() {
    return LocalStorageService.getRefreshToken();
  }

  // Get access token (legacy compatibility)
  getAccessToken() {
    return LocalStorageService.getAccessToken();
  }

  // Check if access token is expired
  isAccessTokenExpired() {
    return LocalStorageService.isAccessTokenExpired();
  }

  // Check if access token is expiring soon
  isAccessTokenExpiringSoon(minutesBeforeExpiry = 5) {
    return LocalStorageService.isAccessTokenExpiringSoon(minutesBeforeExpiry);
  }

  // Check if refresh token is expired
  isRefreshTokenExpired() {
    return LocalStorageService.isRefreshTokenExpired();
  }

  // Check if user has valid tokens
  hasValidTokens() {
    return LocalStorageService.hasValidTokens();
  }

  // Check if user has encrypted tokens
  hasEncryptedTokens() {
    return LocalStorageService.hasEncryptedTokens();
  }

  // Get token expiry information
  getTokenExpiryInfo() {
    return LocalStorageService.getTokenExpiryInfo();
  }

  // Set user email
  setUserEmail(email) {
    return LocalStorageService.setUserEmail(email);
  }

  // Get user email
  getUserEmail() {
    return LocalStorageService.getUserEmail();
  }

  // Clear all tokens and auth data
  clearTokens() {
    return LocalStorageService.clearAuthData();
  }

  // Clear all authentication data (alias)
  clearAuthData() {
    return LocalStorageService.clearAuthData();
  }

  // Store login session data
  setLoginSessionData(key, data) {
    return LocalStorageService.setLoginSessionData(key, data);
  }

  // Get login session data
  getLoginSessionData(key) {
    return LocalStorageService.getLoginSessionData(key);
  }

  // Clear login session data
  clearLoginSessionData(key) {
    return LocalStorageService.clearLoginSessionData(key);
  }

  // Clear all login session data
  clearAllLoginSessionData() {
    return LocalStorageService.clearAllLoginSessionData();
  }

  // Migrate legacy tokens
  migrateLegacyTokens() {
    return LocalStorageService.migrateLegacyTokens();
  }

  // Get all storage data for worker communication
  getAllStorageData() {
    return LocalStorageService.getAllStorageData();
  }

  // Check if user is authenticated
  isAuthenticated() {
    return LocalStorageService.isAuthenticated();
  }

  // Get token health status
  getTokenHealth() {
    const tokenInfo = this.getTokenExpiryInfo();
    const hasEncrypted = this.hasEncryptedTokens();
    const hasLegacy = !!(
      localStorage.getItem("mapleapps_access_token") ||
      localStorage.getItem("mapleapps_refresh_token")
    );

    const health = {
      status: "unknown",
      recommendations: [],
      canRefresh: false,
      needsReauth: false,
    };

    if (!hasEncrypted && hasLegacy) {
      health.status = "legacy_migration_needed";
      health.recommendations.push(
        "Migrate to encrypted token system by re-authenticating",
      );
      health.needsReauth = true;
    } else if (hasEncrypted) {
      if (tokenInfo.refreshTokenExpired) {
        health.status = "expired";
        health.recommendations.push(
          "Refresh token expired - re-authentication required",
        );
        health.needsReauth = true;
      } else if (tokenInfo.accessTokenExpired) {
        health.status = "needs_refresh";
        health.recommendations.push(
          "Access token expired - refresh recommended",
        );
        health.canRefresh = true;
      } else if (tokenInfo.accessTokenExpiringSoon) {
        health.status = "expiring_soon";
        health.recommendations.push(
          "Access token expiring soon - refresh recommended",
        );
        health.canRefresh = true;
      } else {
        health.status = "healthy";
        health.recommendations.push("Tokens are valid and healthy");
      }
    } else {
      health.status = "no_tokens";
      health.recommendations.push("No authentication tokens found");
      health.needsReauth = true;
    }

    return health;
  }

  // Refresh tokens via worker
  async refreshTokensViaWorker() {
    try {
      const result = await WorkerManager.manualRefresh();
      return result;
    } catch (error) {
      console.error("[TokenService] Worker token refresh failed:", error);
      throw error;
    }
  }

  // Force token check via worker
  forceTokenCheck() {
    WorkerManager.forceTokenCheck();
  }

  // Get debug information
  getDebugInfo() {
    return {
      hasEncryptedTokens: this.hasEncryptedTokens(),
      hasValidTokens: this.hasValidTokens(),
      isAuthenticated: this.isAuthenticated(),
      tokenHealth: this.getTokenHealth(),
      tokenInfo: this.getTokenExpiryInfo(),
      userEmail: this.getUserEmail(),
      hasSessionKeys: LocalStorageService.hasSessionKeys(),
      canDecryptTokens: LocalStorageService.hasSessionKeys(),
      storageKeys: {
        encryptedAccessToken: !!this.getEncryptedAccessToken(),
        encryptedRefreshToken: !!this.getEncryptedRefreshToken(),
        tokenNonce: !!this.getTokenNonce(),
        accessTokenExpiry: !!localStorage.getItem(this.ACCESS_TOKEN_EXPIRY_KEY),
        refreshTokenExpiry: !!localStorage.getItem(
          this.REFRESH_TOKEN_EXPIRY_KEY,
        ),
        userEmail: !!this.getUserEmail(),
      },
    };
  }

  // Check if we have session keys for token decryption
  hasSessionKeys() {
    return LocalStorageService.hasSessionKeys();
  }

  // Get decrypted access token
  async getDecryptedAccessToken() {
    return await LocalStorageService.getDecryptedAccessToken();
  }

  // Check if we can make authenticated API calls
  canMakeAuthenticatedRequests() {
    return this.hasSessionKeys() && this.hasEncryptedTokens();
  }

  // Clear session keys (for logout)
  clearSessionKeys() {
    return LocalStorageService.clearSessionKeys();
  }

  // Legacy methods for backward compatibility
  setAccessToken(token, expiryTime) {
    return LocalStorageService.setAccessToken(token, expiryTime);
  }

  setRefreshToken(token, expiryTime) {
    return LocalStorageService.setRefreshToken(token, expiryTime);
  }

  // Auto-refresh tokens if needed
  async autoRefreshIfNeeded() {
    const tokenHealth = this.getTokenHealth();

    if (tokenHealth.needsReauth) {
      throw new Error("Re-authentication required");
    }

    if (
      tokenHealth.canRefresh &&
      (tokenHealth.status === "needs_refresh" ||
        tokenHealth.status === "expiring_soon")
    ) {
      console.log("[TokenService] Auto-refreshing tokens...");
      try {
        await this.refreshTokensViaWorker();
        console.log("[TokenService] Auto-refresh successful");
        return true;
      } catch (error) {
        console.error("[TokenService] Auto-refresh failed:", error);
        throw error;
      }
    }

    return false; // No refresh needed
  }
}

export default TokenService;
