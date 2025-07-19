// File: src/services/Storage/DashboardStorageService.js
// Dashboard Storage Service - Handles localStorage operations for dashboard data

class DashboardStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      DASHBOARD_DATA: "mapleapps_dashboard_data",
      DASHBOARD_METADATA: "mapleapps_dashboard_metadata",
    };

    // Cache configuration - shorter duration for dashboard since it's dynamic data
    this.CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

    console.log("[DashboardStorageService] Storage service initialized");
  }

  // === Dashboard Data Storage Operations ===

  // Store dashboard data with metadata
  storeDashboardData(dashboardData, metadata = {}) {
    try {
      const cachedData = {
        dashboard: dashboardData,
        metadata: {
          ...metadata,
          cached_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + this.CACHE_DURATION).toISOString(),
        },
      };

      localStorage.setItem(
        this.STORAGE_KEYS.DASHBOARD_DATA,
        JSON.stringify(cachedData),
      );

      console.log(
        "[DashboardStorageService] Dashboard data stored successfully",
      );
    } catch (error) {
      console.error(
        "[DashboardStorageService] Failed to store dashboard data:",
        error,
      );
    }
  }

  // Get stored dashboard data
  getDashboardData() {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.DASHBOARD_DATA);
      if (!stored) {
        return null;
      }

      const cachedData = JSON.parse(stored);

      // Check if cache has expired
      const expiresAt = new Date(cachedData.metadata.expires_at);
      if (new Date() > expiresAt) {
        console.log("[DashboardStorageService] Dashboard cache expired");
        this.clearDashboardData();
        return null;
      }

      console.log(
        "[DashboardStorageService] Dashboard data retrieved from cache",
      );
      return cachedData;
    } catch (error) {
      console.error(
        "[DashboardStorageService] Failed to get dashboard data:",
        error,
      );
      return null;
    }
  }

  // Clear dashboard data cache
  clearDashboardData() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.DASHBOARD_DATA);
      localStorage.removeItem(this.STORAGE_KEYS.DASHBOARD_METADATA);
      console.log("[DashboardStorageService] Dashboard cache cleared");
    } catch (error) {
      console.error(
        "[DashboardStorageService] Failed to clear dashboard data:",
        error,
      );
    }
  }

  // === Cache Validation ===

  // Check if dashboard cache is valid
  hasValidDashboardCache() {
    const dashboardData = this.getDashboardData();
    return dashboardData !== null;
  }

  // Get cache age in minutes
  getCacheAge() {
    const dashboardData = this.getDashboardData();
    if (!dashboardData) {
      return null;
    }

    const cachedAt = new Date(dashboardData.metadata.cached_at);
    const now = new Date();
    return Math.floor((now - cachedAt) / (1000 * 60));
  }

  // === Cache Management ===

  // Clear expired caches
  clearExpiredCaches() {
    try {
      const now = new Date();
      let clearedCount = 0;

      // Clear expired dashboard data
      const dashboardData = this.getDashboardData();
      if (dashboardData) {
        const expiresAt = new Date(dashboardData.metadata.expires_at);
        if (now > expiresAt) {
          this.clearDashboardData();
          clearedCount++;
        }
      }

      console.log(
        "[DashboardStorageService] Cleared",
        clearedCount,
        "expired cache entries",
      );
      return clearedCount;
    } catch (error) {
      console.error(
        "[DashboardStorageService] Failed to clear expired caches:",
        error,
      );
      return 0;
    }
  }

  // Clear all caches
  clearAllCaches() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.DASHBOARD_DATA);
      localStorage.removeItem(this.STORAGE_KEYS.DASHBOARD_METADATA);
      console.log("[DashboardStorageService] All caches cleared");
    } catch (error) {
      console.error(
        "[DashboardStorageService] Failed to clear all caches:",
        error,
      );
    }
  }

  // === Configuration ===

  // Set cache duration
  setCacheDuration(duration) {
    this.CACHE_DURATION = duration;
    console.log(
      "[DashboardStorageService] Cache duration set to:",
      duration,
      "ms",
    );
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const dashboardData = this.getDashboardData();

    return {
      hasDashboardCache: !!dashboardData,
      cacheDuration: this.CACHE_DURATION,
      cacheAge: this.getCacheAge(),
      isValid: this.hasValidDashboardCache(),
      lastCached: dashboardData?.metadata?.cached_at || null,
      cacheExpiry: dashboardData?.metadata?.expires_at || null,
    };
  }

  // Get debug information
  getDebugInfo() {
    const dashboardData = this.getDashboardData();

    return {
      serviceName: "DashboardStorageService",
      storageInfo: this.getStorageInfo(),
      lastCached: dashboardData?.metadata?.cached_at || null,
      cacheExpiry: dashboardData?.metadata?.expires_at || null,
      dashboardDataAvailable: !!dashboardData?.dashboard,
    };
  }
}

export default DashboardStorageService;
