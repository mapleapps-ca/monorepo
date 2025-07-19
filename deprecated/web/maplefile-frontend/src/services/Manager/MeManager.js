// File: monorepo/web/maplefile-frontend/src/services/Manager/MeManager.js
// Me Manager - Orchestrates API and Storage services for user management

import MeAPIService from "../API/MeAPIService.js";
import MeStorageService from "../Storage/MeStorageService.js";

class MeManager {
  constructor(authManager) {
    // MeManager depends on AuthManager and orchestrates API and Storage services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new MeAPIService(authManager);
    this.storageService = new MeStorageService();

    console.log(
      "[MeManager] User manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[MeManager] Initializing user manager...");
      await this.storageService.initialize();
      console.log("[MeManager] User manager initialized successfully");
    } catch (error) {
      console.error("[MeManager] Failed to initialize user manager:", error);
    }
  }

  // === User Data Management ===

  // Get current user information (with smart caching)
  async getCurrentUser() {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log(
        "[MeManager] Orchestrating current user information retrieval",
      );

      // Check if we have fresh cached data
      if (this.storageService.isCacheFresh()) {
        console.log("[MeManager] Returning cached user data");
        return this.storageService.getCachedUser();
      }

      // Fetch from API if cache is stale or empty
      console.log("[MeManager] Fetching fresh user data from API");
      const userData = await this.apiService.getCurrentUser();

      // Cache the fresh data
      this.storageService.cacheUser(userData);
      this.storageService.persistUserData(userData); // Optional persistence

      console.log("[MeManager] User data retrieved and cached successfully");
      return userData;
    } catch (error) {
      console.error("[MeManager] Failed to get current user:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Update current user information
  async updateCurrentUser(updateData) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    try {
      this.isLoading = true;
      console.log("[MeManager] Orchestrating current user information update");

      // Update via API
      const updatedUser = await this.apiService.updateCurrentUser(updateData);

      // Update cache with fresh data
      this.storageService.cacheUser(updatedUser);
      this.storageService.persistUserData(updatedUser); // Optional persistence

      console.log("[MeManager] User data updated and cached successfully");
      return updatedUser;
    } catch (error) {
      console.error("[MeManager] Failed to update current user:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Delete current user account
  async deleteCurrentUser(password) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    if (!password) {
      throw new Error("Password is required to delete account");
    }

    try {
      this.isLoading = true;
      console.log("[MeManager] Orchestrating current user account deletion");

      // Delete via API
      await this.apiService.deleteCurrentUser(password);

      // Clear cached user data
      this.storageService.clearUserData();
      this.storageService.clearPersistedUserData();

      console.log(
        "[MeManager] User account deleted and cache cleared successfully",
      );
    } catch (error) {
      console.error("[MeManager] Failed to delete current user:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === User Profile Management ===

  // Get user profile from cache or fetch if needed
  async getUserProfile(forceRefresh = false) {
    if (!this.authManager.isAuthenticated()) {
      throw new Error("User not authenticated via AuthManager");
    }

    // Return cached data if available and not forcing refresh
    if (!forceRefresh && this.storageService.isCacheFresh()) {
      return this.storageService.getCachedUser();
    }

    return await this.getCurrentUser();
  }

  // Refresh user data (force fetch from API)
  async refreshUserData() {
    console.log("[MeManager] Forcing user data refresh");
    this.storageService.invalidateCache();
    return await this.getCurrentUser();
  }

  // === User Information Utilities ===

  // Get user's email (from cache or auth manager)
  getUserEmail() {
    // First try to get from storage cache
    const emailFromCache = this.storageService.getUserEmail();
    if (emailFromCache) {
      return emailFromCache;
    }

    // Fallback to auth manager (from token storage)
    const emailFromAuth = this.authManager.getCurrentUserEmail();
    if (emailFromAuth) {
      return emailFromAuth;
    }

    return null;
  }

  // Get user's display name
  getUserDisplayName() {
    const displayName = this.storageService.getUserDisplayName();
    return displayName || this.getUserEmail(); // Fallback to email
  }

  // === Permissions and Roles ===

  // Check if user has specific permissions or roles
  hasPermission(permission) {
    if (!this.storageService.hasUserData()) {
      return false;
    }

    // Check user permissions array
    const permissions = this.storageService.getUserPermissions();
    if (permissions.includes(permission)) {
      return true;
    }

    // Check role-based permissions
    const role = this.storageService.getUserRole();
    if (role) {
      switch (role) {
        case 1: // Root
          return true; // Root has all permissions
        case 2: // Company
          return ["read", "write", "upload", "download", "manage"].includes(
            permission,
          );
        case 3: // Individual
          return ["read", "write", "upload", "download"].includes(permission);
        default:
          return false;
      }
    }

    return false;
  }

  // Check if user is admin
  isAdmin() {
    return this.storageService.isAdmin();
  }

  // Get user role
  getUserRole() {
    return this.storageService.getUserRole();
  }

  // === Subscription Information ===

  // Get user's subscription or plan information
  getSubscriptionInfo() {
    return this.storageService.getSubscriptionInfo();
  }

  // === Settings and Preferences ===

  // Update user settings
  async updateUserSettings(settings) {
    return await this.updateCurrentUser({ settings });
  }

  // Change user password (not implemented in backend yet)
  async changePassword(currentPassword, newPassword) {
    throw new Error(
      "Password change functionality is not yet implemented in the backend",
    );
  }

  // === Cache Management ===

  // Clear cached user data
  clearUserData() {
    this.storageService.clearUserData();
    this.storageService.clearPersistedUserData();
    console.log("[MeManager] User data cleared from cache");
  }

  // Check if user data is cached
  hasUserData() {
    return this.storageService.hasUserData();
  }

  // Get cached user data without making API call
  getCachedUser() {
    return this.storageService.getCachedUser();
  }

  // Set cache timeout
  setCacheTimeout(timeoutMs) {
    this.storageService.setCacheTimeout(timeoutMs);
  }

  // === State Management ===

  // Get loading state
  isLoadingUser() {
    return this.isLoading;
  }

  // Check if cache is fresh
  isCacheFresh() {
    return this.storageService.isCacheFresh();
  }

  // === Debug and Information ===

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "MeManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      userState: {
        userEmail: this.getUserEmail(),
        displayName: this.getUserDisplayName(),
        isAdmin: this.isAdmin(),
        userRole: this.getUserRole(),
        subscriptionInfo: this.getSubscriptionInfo(),
        isLoading: this.isLoading,
        hasUserData: this.hasUserData(),
        isCacheFresh: this.isCacheFresh(),
      },
      authManagerStatus: {
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        hasSessionKeys: this.authManager.getSessionKeyStatus(),
      },
    };
  }

  // Get user status summary
  getUserStatus() {
    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      hasUserData: this.hasUserData(),
      isCacheFresh: this.isCacheFresh(),
      isLoading: this.isLoading,
      userEmail: this.getUserEmail(),
      displayName: this.getUserDisplayName(),
      role: this.getUserRole(),
      isAdmin: this.isAdmin(),
      canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
    };
  }
}

export default MeManager;
