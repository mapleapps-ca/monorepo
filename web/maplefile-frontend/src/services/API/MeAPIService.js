// File: monorepo/web/maplefile-frontend/src/services/API/MeAPIService.js
// Me API Service - Handles all API calls for user management

class MeAPIService {
  constructor(authManager) {
    // MeAPIService depends on AuthManager for authentication
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[MeAPIService] API service initialized with AuthManager dependency",
    );
  }

  // Import ApiClient for authenticated requests
  async getApiClient() {
    if (!this._apiClient) {
      const { default: ApiClient } = await import("../ApiClient.js");
      this._apiClient = ApiClient;
    }
    return this._apiClient;
  }

  // Get current user information from API
  async getCurrentUser() {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      console.log(
        "[MeAPIService] Fetching current user information from API via AuthManager",
      );

      const apiClient = await this.getApiClient();
      const userData = await apiClient.getMapleFile("/me");

      console.log(
        "[MeAPIService] Current user data retrieved from API via AuthManager:",
        userData,
      );
      return userData;
    } catch (error) {
      console.error(
        "[MeAPIService] Failed to get current user from API via AuthManager:",
        error,
      );
      throw error;
    }
  }

  // Update current user information via API
  async updateCurrentUser(updateData) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      console.log(
        "[MeAPIService] Updating current user information via API and AuthManager",
      );

      const apiClient = await this.getApiClient();
      const updatedUser = await apiClient.putMapleFile("/me", updateData);

      console.log(
        "[MeAPIService] User data updated via API and AuthManager:",
        updatedUser,
      );
      return updatedUser;
    } catch (error) {
      console.error(
        "[MeAPIService] Failed to update current user via API and AuthManager:",
        error,
      );
      throw error;
    }
  }

  // Delete current user account via API
  async deleteCurrentUser(password) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    if (!password) {
      throw new Error("Password is required to delete account");
    }

    try {
      console.log(
        "[MeAPIService] Deleting current user account via API and AuthManager",
      );

      const apiClient = await this.getApiClient();
      await apiClient.deleteMapleFile("/me", {
        body: JSON.stringify({
          password: password,
        }),
      });

      console.log(
        "[MeAPIService] User account deleted successfully via API and AuthManager",
      );
    } catch (error) {
      console.error(
        "[MeAPIService] Failed to delete current user via API and AuthManager:",
        error,
      );
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "MeAPIService",
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

export default MeAPIService;
