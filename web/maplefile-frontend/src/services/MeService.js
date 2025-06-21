// Me Service for managing current user information
import ApiClient from "./ApiClient.js";

class MeService {
  constructor(authService) {
    // MeService depends on AuthService to get the current user
    this.authService = authService;
    this.apiClient = ApiClient;
    this.currentUser = null;
    this.isLoading = false;
  }

  // Get current user information
  async getCurrentUser() {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    try {
      this.isLoading = true;
      console.log("[MeService] Fetching current user information");

      // Make authenticated request to get user profile
      const userData = await this.apiClient.get("/me");

      this.currentUser = userData;
      console.log("[MeService] Current user data retrieved:", userData);

      return userData;
    } catch (error) {
      console.error("[MeService] Failed to get current user:", error);

      // If unauthorized, clear auth data
      if (
        error.message.includes("401") ||
        error.message.includes("Unauthorized")
      ) {
        this.authService.logout();
      }

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Update current user information
  async updateCurrentUser(updateData) {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    try {
      this.isLoading = true;
      console.log("[MeService] Updating current user information");

      // Make authenticated request to update user profile
      const updatedUser = await this.apiClient.patch("/me", updateData);

      this.currentUser = updatedUser;
      console.log("[MeService] User data updated:", updatedUser);

      return updatedUser;
    } catch (error) {
      console.error("[MeService] Failed to update current user:", error);

      // If unauthorized, clear auth data
      if (
        error.message.includes("401") ||
        error.message.includes("Unauthorized")
      ) {
        this.authService.logout();
      }

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Get user profile from cache or fetch if needed
  async getUserProfile(forceRefresh = false) {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    // Return cached data if available and not forcing refresh
    if (this.currentUser && !forceRefresh) {
      return this.currentUser;
    }

    return await this.getCurrentUser();
  }

  // Get user's email from auth service or profile
  getUserEmail() {
    // First try to get from auth service (from token storage)
    const emailFromAuth = this.authService.getCurrentUserEmail();
    if (emailFromAuth) {
      return emailFromAuth;
    }

    // Fallback to cached user profile
    if (this.currentUser && this.currentUser.email) {
      return this.currentUser.email;
    }

    return null;
  }

  // Get user's display name
  getUserDisplayName() {
    if (!this.currentUser) {
      return this.getUserEmail(); // Fallback to email
    }

    // Try various name combinations
    if (this.currentUser.display_name) {
      return this.currentUser.display_name;
    }

    if (this.currentUser.first_name && this.currentUser.last_name) {
      return `${this.currentUser.first_name} ${this.currentUser.last_name}`;
    }

    if (this.currentUser.first_name) {
      return this.currentUser.first_name;
    }

    if (this.currentUser.name) {
      return this.currentUser.name;
    }

    return this.getUserEmail(); // Fallback to email
  }

  // Check if user has specific permissions or roles
  hasPermission(permission) {
    if (!this.currentUser) {
      return false;
    }

    // Check user role or permissions
    if (
      this.currentUser.permissions &&
      Array.isArray(this.currentUser.permissions)
    ) {
      return this.currentUser.permissions.includes(permission);
    }

    if (this.currentUser.role) {
      // Simple role-based permission check
      switch (this.currentUser.role) {
        case "admin":
        case "root":
          return true; // Admin has all permissions
        case "user":
          return ["read", "write", "upload", "download"].includes(permission);
        default:
          return false;
      }
    }

    return false;
  }

  // Check if user is admin
  isAdmin() {
    if (!this.currentUser) {
      return false;
    }

    return (
      this.currentUser.role === "admin" ||
      this.currentUser.role === "root" ||
      this.currentUser.user_role === 1
    ); // Root role from API
  }

  // Get user's subscription or plan information
  getSubscriptionInfo() {
    if (!this.currentUser) {
      return null;
    }

    return {
      plan: this.currentUser.plan || "free",
      subscription_status: this.currentUser.subscription_status || "inactive",
      storage_quota: this.currentUser.storage_quota || null,
      storage_used: this.currentUser.storage_used || 0,
    };
  }

  // Clear cached user data
  clearUserData() {
    this.currentUser = null;
    console.log("[MeService] User data cleared");
  }

  // Get loading state
  isLoadingUser() {
    return this.isLoading;
  }

  // Get cached user data without making API call
  getCachedUser() {
    return this.currentUser;
  }

  // Check if user data is cached
  hasUserData() {
    return !!this.currentUser;
  }

  // Refresh user data
  async refreshUserData() {
    return await this.getCurrentUser();
  }

  // Update user settings
  async updateUserSettings(settings) {
    return await this.updateCurrentUser({ settings });
  }

  // Change user password (if supported by API)
  async changePassword(currentPassword, newPassword) {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    try {
      console.log("[MeService] Changing user password");

      const result = await this.apiClient.post("/me/change-password", {
        current_password: currentPassword,
        new_password: newPassword,
      });

      console.log("[MeService] Password changed successfully");
      return result;
    } catch (error) {
      console.error("[MeService] Failed to change password:", error);
      throw error;
    }
  }

  // Get user activity or session information
  async getUserActivity() {
    if (!this.authService.isAuthenticated()) {
      throw new Error("User not authenticated");
    }

    try {
      console.log("[MeService] Fetching user activity");

      const activity = await this.apiClient.get("/me/activity");

      console.log("[MeService] User activity retrieved");
      return activity;
    } catch (error) {
      console.error("[MeService] Failed to get user activity:", error);
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      isAuthenticated: this.authService.isAuthenticated(),
      currentUser: this.currentUser,
      userEmail: this.getUserEmail(),
      displayName: this.getUserDisplayName(),
      isAdmin: this.isAdmin(),
      subscriptionInfo: this.getSubscriptionInfo(),
      isLoading: this.isLoading,
      hasUserData: this.hasUserData(),
    };
  }
}

export default MeService;
