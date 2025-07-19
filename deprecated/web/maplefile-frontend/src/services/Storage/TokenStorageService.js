// File: monorepo/web/maplefile-frontend/src/services/Storage/TokenStorageService.js
// Token Storage Service - Handles all token storage operations
import LocalStorageService from "./LocalStorageService.js";

class TokenStorageService {
  constructor() {
    console.log("[TokenStorageService] Storage service initialized");
  }

  // === Token Storage Operations ===

  // Store unencrypted access token with expiry
  setAccessToken(token, expiryTime) {
    LocalStorageService.setAccessToken(token, expiryTime);
    console.log("[TokenStorageService] Access token stored");
  }

  // Store unencrypted refresh token with expiry
  setRefreshToken(token, expiryTime) {
    LocalStorageService.setRefreshToken(token, expiryTime);
    console.log("[TokenStorageService] Refresh token stored");
  }

  // Store both tokens with expiry times
  setTokens(accessToken, refreshToken, accessTokenExpiry, refreshTokenExpiry) {
    LocalStorageService.setTokens(
      accessToken,
      refreshToken,
      accessTokenExpiry,
      refreshTokenExpiry,
    );
    console.log("[TokenStorageService] Both tokens stored successfully");
  }

  // === Token Retrieval Operations ===

  // Get unencrypted access token
  getAccessToken() {
    return LocalStorageService.getAccessToken();
  }

  // Get unencrypted refresh token
  getRefreshToken() {
    return LocalStorageService.getRefreshToken();
  }

  // === Token Status Checking ===

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

  // === User Email Management ===

  // Set user email
  setUserEmail(email) {
    LocalStorageService.setUserEmail(email);
    console.log("[TokenStorageService] User email stored");
  }

  // Get user email
  getUserEmail() {
    return LocalStorageService.getUserEmail();
  }

  // === Authentication Status ===

  // Check if user is authenticated
  isAuthenticated() {
    return LocalStorageService.isAuthenticated();
  }

  // Check if we can make authenticated API calls
  canMakeAuthenticatedRequests() {
    return this.hasValidTokens();
  }

  // === Token Cleanup ===

  // Clear all tokens and auth data
  clearTokens() {
    LocalStorageService.clearAuthData();
    console.log("[TokenStorageService] All tokens cleared");
  }

  // Clear all authentication data (alias)
  clearAuthData() {
    LocalStorageService.clearAuthData();
    console.log("[TokenStorageService] All authentication data cleared");
  }

  // === Login Session Data ===

  // Store login session data
  setLoginSessionData(key, data) {
    LocalStorageService.setLoginSessionData(key, data);
    console.log(`[TokenStorageService] Login session data stored: ${key}`);
  }

  // Get login session data
  getLoginSessionData(key) {
    return LocalStorageService.getLoginSessionData(key);
  }

  // Clear login session data
  clearLoginSessionData(key) {
    LocalStorageService.clearLoginSessionData(key);
    console.log(`[TokenStorageService] Login session data cleared: ${key}`);
  }

  // Clear all login session data
  clearAllLoginSessionData() {
    LocalStorageService.clearAllLoginSessionData();
    console.log("[TokenStorageService] All login session data cleared");
  }

  // === Storage Information ===

  // Get all storage data for debugging
  getAllStorageData() {
    return LocalStorageService.getAllStorageData();
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

  // Get storage info
  getStorageInfo() {
    const tokenInfo = this.getTokenExpiryInfo();
    const userEmail = this.getUserEmail();
    const hasValidTokens = this.hasValidTokens();

    return {
      // Authentication status
      isAuthenticated: this.isAuthenticated(),
      hasValidTokens,
      userEmail,

      // Token information
      tokenInfo,

      // Token health
      tokenHealth: this.getTokenHealth(),

      // Storage keys in localStorage
      storageKeys: {
        accessToken: !!this.getAccessToken(),
        refreshToken: !!this.getRefreshToken(),
        userEmail: !!userEmail,
        accessTokenExpiry: !!localStorage.getItem(
          "mapleapps_access_token_expiry",
        ),
        refreshTokenExpiry: !!localStorage.getItem(
          "mapleapps_refresh_token_expiry",
        ),
      },
    };
  }

  // === Debug Information ===

  getDebugInfo() {
    return {
      serviceName: "TokenStorageService",
      storageInfo: this.getStorageInfo(),
      canMakeAuthenticatedRequests: this.canMakeAuthenticatedRequests(),
      refreshMethod: "api_interceptor",
    };
  }
}

export default TokenStorageService;
