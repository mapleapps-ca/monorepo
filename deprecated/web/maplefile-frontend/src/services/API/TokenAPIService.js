// File: monorepo/web/maplefile-frontend/src/services/API/TokenAPIService.js
// Token API Service - Handles token-related API calls

class TokenAPIService {
  constructor(authManager) {
    // TokenAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[TokenAPIService] API service initialized with AuthManager dependency",
    );
  }

  // Import ApiClient for authenticated requests
  async getApiClient() {
    if (!this._apiClient) {
      const { default: ApiClient } = await import("./ApiClient.js");
      this._apiClient = ApiClient;
    }
    return this._apiClient;
  }

  // Refresh tokens via API
  async refreshTokens() {
    try {
      console.log("[TokenAPIService] Refreshing tokens via API");

      const apiClient = await this.getApiClient();
      const response = await apiClient.refreshTokens();

      console.log("[TokenAPIService] Token refresh successful");
      return response;
    } catch (error) {
      console.error("[TokenAPIService] Token refresh failed:", error);
      throw error;
    }
  }

  // Validate token with API (if needed in future)
  async validateToken(token) {
    try {
      console.log("[TokenAPIService] Validating token via API");

      // This endpoint doesn't exist yet but could be added
      // const apiClient = await this.getApiClient();
      // const response = await apiClient.post('/token/validate', { token });

      // For now, just return a mock response
      console.log(
        "[TokenAPIService] Token validation not implemented on backend",
      );
      return {
        valid: true,
        message: "Token validation endpoint not implemented",
      };
    } catch (error) {
      console.error("[TokenAPIService] Token validation failed:", error);
      throw error;
    }
  }

  // Revoke tokens (logout)
  async revokeTokens(refreshToken) {
    try {
      console.log("[TokenAPIService] Revoking tokens via API");

      // This endpoint might not exist yet but is a common pattern
      // const apiClient = await this.getApiClient();
      // const response = await apiClient.post('/token/revoke', {
      //   refresh_token: refreshToken
      // });

      console.log(
        "[TokenAPIService] Token revocation not implemented on backend",
      );
      return {
        success: true,
        message: "Token revocation endpoint not implemented",
      };
    } catch (error) {
      console.error("[TokenAPIService] Token revocation failed:", error);
      throw error;
    }
  }

  // Get token status from API
  async getTokenStatus() {
    try {
      const apiClient = await this.getApiClient();
      return apiClient.getTokenStatus();
    } catch (error) {
      console.error("[TokenAPIService] Failed to get token status:", error);
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "TokenAPIService",
      managedBy: "AuthManager",
      isAuthenticated: this.authManager.isAuthenticated(),
      canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default TokenAPIService;
