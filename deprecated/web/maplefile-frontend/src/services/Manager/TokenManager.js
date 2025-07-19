// File: monorepo/web/maplefile-frontend/src/services/Manager/TokenManager.js
// Token Manager - Orchestrates API and Storage services for token management

import TokenAPIService from "../API/TokenAPIService.js";
import TokenStorageService from "../Storage/TokenStorageService.js";

class TokenManager {
  constructor(authManager) {
    // TokenManager depends on AuthManager and orchestrates API and Storage services
    this.authManager = authManager;
    this.isLoading = false;
    this.refreshPromise = null;

    // Initialize dependent services
    this.apiService = new TokenAPIService(authManager);
    this.storageService = new TokenStorageService();

    // Event listeners (replacing WorkerManager functionality)
    this.authStateListeners = new Set();

    console.log(
      "[TokenManager] Token manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[TokenManager] Initializing token manager...");
      console.log("[TokenManager] Token manager initialized successfully");
    } catch (error) {
      console.error(
        "[TokenManager] Failed to initialize token manager:",
        error,
      );
    }
  }

  // === Token Storage Operations (via Storage Service) ===

  // Store tokens
  setTokens(accessToken, refreshToken, accessTokenExpiry, refreshTokenExpiry) {
    console.log("[TokenManager] Orchestrating token storage");
    this.storageService.setTokens(
      accessToken,
      refreshToken,
      accessTokenExpiry,
      refreshTokenExpiry,
    );

    // Notify listeners of token update
    this.notifyAuthStateChange("tokens_updated", {
      hasTokens: true,
      accessTokenExpiry,
      refreshTokenExpiry,
    });
  }

  // Get access token
  getAccessToken() {
    return this.storageService.getAccessToken();
  }

  // Get refresh token
  getRefreshToken() {
    return this.storageService.getRefreshToken();
  }

  // === Token Refresh Operations (via API Service) ===

  // Refresh tokens with deduplication
  async refreshTokens() {
    // If already refreshing, return the existing promise
    if (this.refreshPromise) {
      console.log(
        "[TokenManager] Token refresh already in progress, waiting...",
      );
      return this.refreshPromise;
    }

    try {
      this.isLoading = true;
      console.log("[TokenManager] Orchestrating token refresh");

      // Create refresh promise for deduplication
      this.refreshPromise = this.apiService.refreshTokens();

      const response = await this.refreshPromise;

      // The actual token storage is handled by ApiClient/AuthManager
      // We just need to notify listeners
      this.notifyAuthStateChange("token_refresh_success", {
        accessTokenExpiry: response.access_token_expiry_date,
        refreshTokenExpiry: response.refresh_token_expiry_date,
      });

      console.log("[TokenManager] Token refresh orchestrated successfully");
      return response;
    } catch (error) {
      console.error(
        "[TokenManager] Token refresh orchestration failed:",
        error,
      );

      // Notify listeners of refresh failure
      this.notifyAuthStateChange("token_refresh_failed", {
        error: error.message,
      });

      throw error;
    } finally {
      this.isLoading = false;
      this.refreshPromise = null;
    }
  }

  // Force token check (checks if refresh is needed)
  forceTokenCheck() {
    console.log("[TokenManager] Performing token health check");

    const tokenHealth = this.getTokenHealth();

    if (tokenHealth.needsReauth) {
      this.notifyAuthStateChange("force_logout", {
        reason: "tokens_expired",
      });
    } else if (tokenHealth.canRefresh) {
      console.log(
        "[TokenManager] Tokens can be refreshed - handled by ApiClient interceptors",
      );
    }

    return tokenHealth;
  }

  // === Token Status Operations ===

  // Check if user is authenticated
  isAuthenticated() {
    return this.storageService.isAuthenticated();
  }

  // Check if tokens are valid
  hasValidTokens() {
    return this.storageService.hasValidTokens();
  }

  // Check if access token is expired
  isAccessTokenExpired() {
    return this.storageService.isAccessTokenExpired();
  }

  // Check if access token is expiring soon
  isAccessTokenExpiringSoon(minutesBeforeExpiry = 5) {
    return this.storageService.isAccessTokenExpiringSoon(minutesBeforeExpiry);
  }

  // Check if refresh token is expired
  isRefreshTokenExpired() {
    return this.storageService.isRefreshTokenExpired();
  }

  // Get token expiry information
  getTokenExpiryInfo() {
    return this.storageService.getTokenExpiryInfo();
  }

  // Get token health status
  getTokenHealth() {
    return this.storageService.getTokenHealth();
  }

  // Check if we can make authenticated requests
  canMakeAuthenticatedRequests() {
    return this.storageService.canMakeAuthenticatedRequests();
  }

  // === User Management ===

  // Set user email
  setUserEmail(email) {
    this.storageService.setUserEmail(email);
  }

  // Get user email
  getUserEmail() {
    return this.storageService.getUserEmail();
  }

  // === Token Cleanup ===

  // Clear all tokens
  clearTokens() {
    console.log("[TokenManager] Orchestrating token cleanup");
    this.storageService.clearTokens();

    // Notify listeners
    this.notifyAuthStateChange("tokens_cleared", {
      reason: "manual_clear",
    });
  }

  // Clear all authentication data
  clearAuthData() {
    console.log("[TokenManager] Orchestrating auth data cleanup");
    this.storageService.clearAuthData();

    // Notify listeners
    this.notifyAuthStateChange("auth_cleared", {
      reason: "manual_clear",
    });
  }

  // === Login Session Management ===

  // Store login session data
  setLoginSessionData(key, data) {
    this.storageService.setLoginSessionData(key, data);
  }

  // Get login session data
  getLoginSessionData(key) {
    return this.storageService.getLoginSessionData(key);
  }

  // Clear login session data
  clearLoginSessionData(key) {
    this.storageService.clearLoginSessionData(key);
  }

  // Clear all login session data
  clearAllLoginSessionData() {
    this.storageService.clearAllLoginSessionData();
  }

  // === Event Management (Replacing WorkerManager) ===

  // Add auth state change listener
  addAuthStateChangeListener(callback) {
    if (typeof callback === "function") {
      this.authStateListeners.add(callback);
      console.log(
        "[TokenManager] Auth state listener added. Total listeners:",
        this.authStateListeners.size,
      );
    }
  }

  // Remove auth state change listener
  removeAuthStateChangeListener(callback) {
    this.authStateListeners.delete(callback);
    console.log(
      "[TokenManager] Auth state listener removed. Total listeners:",
      this.authStateListeners.size,
    );
  }

  // Notify auth state change
  notifyAuthStateChange(eventType, eventData) {
    console.log(
      `[TokenManager] Notifying ${this.authStateListeners.size} listeners of ${eventType}`,
    );

    this.authStateListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error("[TokenManager] Error in auth state listener:", error);
      }
    });
  }

  // === Auto-refresh Logic ===

  // Check if tokens should be refreshed
  shouldRefreshTokens(minutesBeforeExpiry = 5) {
    return this.isAccessTokenExpiringSoon(minutesBeforeExpiry);
  }

  // Auto-refresh tokens if needed
  async autoRefreshIfNeeded() {
    const tokenHealth = this.getTokenHealth();

    if (tokenHealth.needsReauth) {
      throw new Error("Re-authentication required");
    }

    if (tokenHealth.canRefresh) {
      console.log(
        "[TokenManager] Auto-refresh is handled automatically by ApiClient interceptors",
      );
      return false; // No manual refresh needed - handled automatically
    }

    return false; // No refresh needed
  }

  // === State Management ===

  // Get loading state
  isLoadingTokens() {
    return this.isLoading;
  }

  // === Manager Status ===

  // Get manager status and information
  getManagerStatus() {
    const storageInfo = this.storageService.getStorageInfo();
    const tokenHealth = this.getTokenHealth();

    return {
      isAuthenticated: this.isAuthenticated(),
      isLoading: this.isLoading,
      canMakeRequests: this.canMakeAuthenticatedRequests(),
      tokenHealth,
      storage: storageInfo,
      listenerCount: this.authStateListeners.size,
      refreshMethod: "api_interceptor",
      hasActiveRefresh: !!this.refreshPromise,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "TokenManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }

  // Get refresh method info
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

export default TokenManager;
