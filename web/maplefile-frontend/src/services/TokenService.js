// File: monorepo/web/maplefile-frontend/src/services/TokenService.js
// Token Service - Updated without worker dependencies, delegates to ApiClient
// TODO: Migrate to `Manager/TokenManager.js`
import LocalStorageService from "./Storage/LocalStorageService.js";

class TokenService {
  constructor() {
    // Token storage keys
    this.ACCESS_TOKEN_KEY = "mapleapps_access_token";
    this.REFRESH_TOKEN_KEY = "mapleapps_refresh_token";
    this.ACCESS_TOKEN_EXPIRY_KEY = "mapleapps_access_token_expiry";
    this.REFRESH_TOKEN_EXPIRY_KEY = "mapleapps_refresh_token_expiry";
    this.USER_EMAIL_KEY = "mapleapps_user_email";
  }

  // Store unencrypted access token with expiry
  setAccessToken(token, expiryTime) {
    return LocalStorageService.setAccessToken(token, expiryTime);
  }

  // Store unencrypted refresh token with expiry
  setRefreshToken(token, expiryTime) {
    return LocalStorageService.setRefreshToken(token, expiryTime);
  }

  // Store both tokens with expiry times
  setTokens(accessToken, refreshToken, accessTokenExpiry, refreshTokenExpiry) {
    return LocalStorageService.setTokens(
      accessToken,
      refreshToken,
      accessTokenExpiry,
      refreshTokenExpiry,
    );
  }

  // Get unencrypted access token
  getAccessToken() {
    return LocalStorageService.getAccessToken();
  }

  // Get unencrypted refresh token
  getRefreshToken() {
    return LocalStorageService.getRefreshToken();
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

  // Get all storage data for debugging
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
    const hasTokens = this.hasValidTokens();

    const health = {
      status: "unknown",
      recommendations: [],
      canRefresh: false,
      needsReauth: false,
      refreshMethod: "api_interceptor",
    };

    if (!hasTokens) {
      health.status = "no_tokens";
      health.recommendations.push("No authentication tokens found");
      health.needsReauth = true;
    } else if (tokenInfo.refreshTokenExpired) {
      health.status = "expired";
      health.recommendations.push(
        "Refresh token expired - re-authentication required",
      );
      health.needsReauth = true;
    } else if (tokenInfo.accessTokenExpired) {
      health.status = "needs_refresh";
      health.recommendations.push(
        "Access token expired - refresh handled automatically by ApiClient",
      );
      health.canRefresh = true;
    } else if (tokenInfo.accessTokenExpiringSoon) {
      health.status = "expiring_soon";
      health.recommendations.push(
        "Access token expiring soon - refresh handled automatically by ApiClient",
      );
      health.canRefresh = true;
    } else {
      health.status = "healthy";
      health.recommendations.push("Tokens are valid and healthy");
    }

    return health;
  }

  // Refresh tokens via ApiClient (replaces worker refresh)
  async refreshTokens() {
    try {
      console.log("[TokenService] Delegating token refresh to ApiClient");
      const { default: ApiClient } = await import("./API/ApiClient.js");
      return await ApiClient.refreshTokens();
    } catch (error) {
      console.error("[TokenService] Token refresh failed:", error);
      throw error;
    }
  }

  // Force token check (no-op since handled by ApiClient interceptors)
  forceTokenCheck() {
    console.log(
      "[TokenService] Force token check - handled automatically by ApiClient interceptors",
    );
  }

  // Legacy method names for backward compatibility
  async refreshTokensViaWorker() {
    console.log(
      "[TokenService] refreshTokensViaWorker() is now handled by ApiClient",
    );
    return await this.refreshTokens();
  }

  // Get debug information
  getDebugInfo() {
    return {
      hasValidTokens: this.hasValidTokens(),
      isAuthenticated: this.isAuthenticated(),
      tokenHealth: this.getTokenHealth(),
      tokenInfo: this.getTokenExpiryInfo(),
      userEmail: this.getUserEmail(),
      refreshMethod: "api_interceptor",
      hasWorker: false,
      storageKeys: {
        accessToken: !!this.getAccessToken(),
        refreshToken: !!this.getRefreshToken(),
        accessTokenExpiry: !!localStorage.getItem(this.ACCESS_TOKEN_EXPIRY_KEY),
        refreshTokenExpiry: !!localStorage.getItem(
          this.REFRESH_TOKEN_EXPIRY_KEY,
        ),
        userEmail: !!this.getUserEmail(),
      },
    };
  }

  // Check if we can make authenticated API calls
  canMakeAuthenticatedRequests() {
    return this.hasValidTokens();
  }

  // Auto-refresh tokens if needed (delegated to ApiClient interceptors)
  async autoRefreshIfNeeded() {
    const tokenHealth = this.getTokenHealth();

    if (tokenHealth.needsReauth) {
      throw new Error("Re-authentication required");
    }

    if (tokenHealth.canRefresh) {
      console.log(
        "[TokenService] Auto-refresh is handled automatically by ApiClient interceptors",
      );
      return false; // No manual refresh needed - handled automatically
    }

    return false; // No refresh needed
  }

  // Get method info
  getRefreshMethod() {
    return {
      method: "api_interceptor",
      description:
        "Automatic token refresh via ApiClient interceptors on 401 responses",
      hasWorker: false,
      isAutomatic: true,
    };
  }
}

export default TokenService;
