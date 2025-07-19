// File: src/services/API/DashboardAPIService.js
// Dashboard API Service - Handles API calls for dashboard data

class DashboardAPIService {
  constructor(authManager) {
    // DashboardAPIService depends on AuthManager for authentication context
    this.authManager = authManager;
    this._apiClient = null;
    console.log(
      "[DashboardAPIService] API service initialized with AuthManager dependency",
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

  // Get dashboard data
  async getDashboardData() {
    try {
      console.log("[DashboardAPIService] Fetching dashboard data");

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/dashboard");

      console.log("[DashboardAPIService] Dashboard data retrieved:", {
        totalFiles: response.dashboard?.summary?.totalFiles || 0,
        totalFolders: response.dashboard?.summary?.totalFolders || 0,
        storageUsed: response.dashboard?.summary?.storageUsed?.value || 0,
        recentFilesCount: response.dashboard?.recentFiles?.length || 0,
        storageTrendDataPoints:
          response.dashboard?.storageUsageTrend?.dataPoints?.length || 0,
      });

      return response;
    } catch (error) {
      console.error(
        "[DashboardAPIService] Failed to fetch dashboard data:",
        error,
      );
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "DashboardAPIService",
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

export default DashboardAPIService;
