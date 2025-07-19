// File: monorepo/web/maplefile-frontend/src/services/Storage/MeStorageService.js
// Me Storage Service - Handles user data caching and storage operations

class MeStorageService {
  constructor() {
    // In-memory cache for user data
    this.currentUser = null;
    this.lastFetched = null;
    this.cacheTimeout = 5 * 60 * 1000; // 5 minutes cache timeout

    // Storage keys for localStorage (if needed for persistence)
    this.STORAGE_KEYS = {
      USER_CACHE: "maplefile_user_cache",
      USER_CACHE_TIMESTAMP: "maplefile_user_cache_timestamp",
    };

    console.log("[MeStorageService] Storage service initialized");
  }

  // === User Data Caching ===

  // Cache user data in memory
  cacheUser(userData) {
    if (!userData) {
      console.warn(
        "[MeStorageService] Attempted to cache null/undefined user data",
      );
      return;
    }

    this.currentUser = userData;
    this.lastFetched = Date.now();
    console.log(
      "[MeStorageService] User data cached in memory:",
      userData.email || "unknown",
    );
  }

  // Get cached user data
  getCachedUser() {
    return this.currentUser;
  }

  // Check if user data is cached
  hasUserData() {
    return !!this.currentUser;
  }

  // Check if cached data is fresh (within timeout)
  isCacheFresh() {
    if (!this.lastFetched || !this.currentUser) {
      return false;
    }

    const age = Date.now() - this.lastFetched;
    return age < this.cacheTimeout;
  }

  // Clear cached user data
  clearUserData() {
    this.currentUser = null;
    this.lastFetched = null;
    console.log("[MeStorageService] User data cleared from cache");
  }

  // === Persistent Storage (Optional - for long-term caching) ===

  // Save user data to localStorage (optional for persistence)
  persistUserData(userData) {
    try {
      if (!userData) return;

      localStorage.setItem(
        this.STORAGE_KEYS.USER_CACHE,
        JSON.stringify(userData),
      );
      localStorage.setItem(
        this.STORAGE_KEYS.USER_CACHE_TIMESTAMP,
        Date.now().toString(),
      );
      console.log("[MeStorageService] User data persisted to localStorage");
    } catch (error) {
      console.warn("[MeStorageService] Failed to persist user data:", error);
    }
  }

  // Load user data from localStorage
  loadPersistedUserData() {
    try {
      const userData = localStorage.getItem(this.STORAGE_KEYS.USER_CACHE);
      const timestamp = localStorage.getItem(
        this.STORAGE_KEYS.USER_CACHE_TIMESTAMP,
      );

      if (!userData || !timestamp) {
        return null;
      }

      // Check if persisted data is still fresh
      const age = Date.now() - parseInt(timestamp);
      if (age > this.cacheTimeout) {
        console.log(
          "[MeStorageService] Persisted user data is stale, ignoring",
        );
        this.clearPersistedUserData();
        return null;
      }

      const parsedData = JSON.parse(userData);
      console.log("[MeStorageService] User data loaded from localStorage");
      return parsedData;
    } catch (error) {
      console.warn(
        "[MeStorageService] Failed to load persisted user data:",
        error,
      );
      this.clearPersistedUserData();
      return null;
    }
  }

  // Clear persisted user data
  clearPersistedUserData() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.USER_CACHE);
      localStorage.removeItem(this.STORAGE_KEYS.USER_CACHE_TIMESTAMP);
      console.log("[MeStorageService] Persisted user data cleared");
    } catch (error) {
      console.warn(
        "[MeStorageService] Failed to clear persisted user data:",
        error,
      );
    }
  }

  // === User Email Storage (from auth) ===

  // Get user email from cached data
  getUserEmail() {
    if (this.currentUser && this.currentUser.email) {
      return this.currentUser.email;
    }
    return null;
  }

  // === User Display Name ===

  // Get user's display name from cached data
  getUserDisplayName() {
    if (!this.currentUser) {
      return null;
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

  // === User Role and Permissions ===

  // Get user role from cached data
  getUserRole() {
    if (!this.currentUser) {
      return null;
    }
    return this.currentUser.role || this.currentUser.user_role || null;
  }

  // Check if user is admin based on cached data
  isAdmin() {
    const role = this.getUserRole();
    return role === 1; // Root role
  }

  // Get user permissions from cached data
  getUserPermissions() {
    if (!this.currentUser) {
      return [];
    }

    if (
      this.currentUser.permissions &&
      Array.isArray(this.currentUser.permissions)
    ) {
      return this.currentUser.permissions;
    }

    return [];
  }

  // === Subscription Information ===

  // Get user's subscription info from cached data
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

  // === Cache Management ===

  // Set cache timeout
  setCacheTimeout(timeoutMs) {
    this.cacheTimeout = timeoutMs;
    console.log(`[MeStorageService] Cache timeout set to ${timeoutMs}ms`);
  }

  // Force cache refresh by clearing cached data
  invalidateCache() {
    this.clearUserData();
    this.clearPersistedUserData();
    console.log("[MeStorageService] Cache invalidated");
  }

  // Initialize storage (load persisted data if available)
  async initialize() {
    try {
      // Try to load persisted user data
      const persistedData = this.loadPersistedUserData();
      if (persistedData) {
        this.cacheUser(persistedData);
        console.log("[MeStorageService] Initialized with persisted user data");
      } else {
        console.log("[MeStorageService] Initialized with no persisted data");
      }
    } catch (error) {
      console.warn("[MeStorageService] Failed to initialize:", error);
    }
  }

  // === Storage Information ===

  // Get storage info for debugging
  getStorageInfo() {
    return {
      hasUserData: this.hasUserData(),
      isCacheFresh: this.isCacheFresh(),
      lastFetched: this.lastFetched,
      cacheTimeout: this.cacheTimeout,
      userEmail: this.getUserEmail(),
      displayName: this.getUserDisplayName(),
      userRole: this.getUserRole(),
      isAdmin: this.isAdmin(),
      subscriptionInfo: this.getSubscriptionInfo(),
      cacheAge: this.lastFetched ? Date.now() - this.lastFetched : null,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "MeStorageService",
      storageInfo: this.getStorageInfo(),
      currentUser: this.currentUser,
      storageKeys: {
        userCache: !!localStorage.getItem(this.STORAGE_KEYS.USER_CACHE),
        userCacheTimestamp: !!localStorage.getItem(
          this.STORAGE_KEYS.USER_CACHE_TIMESTAMP,
        ),
      },
    };
  }
}

export default MeStorageService;
