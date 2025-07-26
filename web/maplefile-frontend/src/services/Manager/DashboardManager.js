// File: src/services/Manager/DashboardManager.js
// Dashboard Manager - Orchestrates API and Storage services for dashboard data
// Enhanced with better cache management and event handling

import DashboardAPIService from "../API/DashboardAPIService.js";
import DashboardStorageService from "../Storage/DashboardStorageService.js";

class DashboardManager {
  constructor(authManager) {
    // DashboardManager depends on AuthManager
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new DashboardAPIService(authManager);
    this.storageService = new DashboardStorageService();

    // Event listeners for dashboard events
    this.dashboardListeners = new Set();

    // 🔧 NEW: Track last refresh for preventing rapid refreshes
    this.lastRefreshTime = 0;
    this.minimumRefreshInterval = 2000; // 2 seconds

    console.log(
      "[DashboardManager] Manager initialized with AuthManager dependency",
    );

    // 🔧 NEW: Set up global event listeners for cache invalidation
    this.setupCacheInvalidationListeners();
  }

  // 🔧 NEW: Set up listeners for events that should invalidate dashboard cache
  setupCacheInvalidationListeners() {
    // Listen for file upload completion events
    if (typeof window !== "undefined") {
      window.addEventListener(
        "dashboardRefresh",
        this.handleDashboardRefreshEvent.bind(this),
      );
      window.addEventListener("storage", this.handleStorageEvent.bind(this));
    }
  }

  // 🔧 NEW: Handle dashboard refresh events
  handleDashboardRefreshEvent(event) {
    console.log(
      "[DashboardManager] Dashboard refresh event received:",
      event.detail,
    );
    this.invalidateCacheAndRefresh("external_event");
  }

  // 🔧 NEW: Handle storage events (for cross-tab communication)
  handleStorageEvent(event) {
    if (event.key === "mapleapps_upload_refresh_signal") {
      console.log("[DashboardManager] Upload refresh signal detected");
      this.invalidateCacheAndRefresh("cross_tab_upload");
    }
  }

  // 🔧 NEW: Invalidate cache and trigger refresh with throttling
  async invalidateCacheAndRefresh(reason = "unknown") {
    const now = Date.now();

    // Throttle refreshes to prevent rapid successive calls
    if (now - this.lastRefreshTime < this.minimumRefreshInterval) {
      console.log(
        "[DashboardManager] Refresh throttled, too soon since last refresh",
      );
      return;
    }

    this.lastRefreshTime = now;

    console.log(
      "[DashboardManager] Invalidating cache and refreshing due to:",
      reason,
    );

    // Clear all caches
    this.clearAllCaches();

    // Notify listeners that cache was invalidated
    this.notifyDashboardListeners("cache_invalidated", {
      reason,
      timestamp: now,
    });

    // If we're authenticated, trigger a refresh
    if (this.authManager.isAuthenticated()) {
      try {
        await this.getDashboardData(true);
        this.notifyDashboardListeners("auto_refresh_completed", {
          reason,
          timestamp: Date.now(),
        });
      } catch (error) {
        console.error("[DashboardManager] Auto-refresh failed:", error);
        this.notifyDashboardListeners("auto_refresh_failed", {
          reason,
          error: error.message,
          timestamp: Date.now(),
        });
      }
    }
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[DashboardManager] Initializing manager...");

      // Initialize crypto services for decrypting recent files
      const { default: FileCryptoService } = await import(
        "../Crypto/FileCryptoService.js"
      );
      await FileCryptoService.initialize();
      this.fileCryptoService = FileCryptoService;

      // Initialize collection crypto service for collection keys
      const { default: CollectionCryptoService } = await import(
        "../Crypto/CollectionCryptoService.js"
      );
      await CollectionCryptoService.initialize();
      this.collectionCryptoService = CollectionCryptoService;

      console.log("[DashboardManager] Manager initialized successfully");
    } catch (error) {
      console.error("[DashboardManager] Failed to initialize manager:", error);
    }
  }

  // === Core Dashboard Methods ===

  // Get dashboard data with caching
  async getDashboardData(forceRefresh = false) {
    try {
      this.isLoading = true;
      console.log("[DashboardManager] === Getting Dashboard Data ===");
      console.log("[DashboardManager] Force refresh:", forceRefresh);

      // 🔧 ENHANCED: Check if cache should be bypassed due to recent invalidation
      const cacheAge = this.storageService.getCacheAge();
      const shouldBypassCache =
        forceRefresh || (cacheAge !== null && cacheAge < 1); // Less than 1 minute old but force refresh

      // STEP 1: Check cache first unless forcing refresh
      if (!shouldBypassCache) {
        const cachedDashboard = this.storageService.getDashboardData();
        if (cachedDashboard) {
          console.log("[DashboardManager] Found cached dashboard data");

          this.notifyDashboardListeners("dashboard_loaded_from_cache", {
            fromCache: true,
            summary: cachedDashboard.dashboard?.summary,
            cacheAge: cacheAge,
          });

          console.log("[DashboardManager] Returning cached dashboard data");
          return cachedDashboard.dashboard;
        }
      }

      // STEP 2: Fetch from API
      console.log("[DashboardManager] Fetching dashboard data from API");
      const response = await this.apiService.getDashboardData();

      const dashboardData = response.dashboard;

      if (!dashboardData) {
        throw new Error("Invalid dashboard response format");
      }

      console.log("[DashboardManager] Dashboard data fetched from API");

      // STEP 3: Process recent files (decrypt metadata)
      if (dashboardData.recentFiles && dashboardData.recentFiles.length > 0) {
        console.log(
          "[DashboardManager] Processing recent files for decryption...",
        );
        dashboardData.recentFiles = await this.processRecentFiles(
          dashboardData.recentFiles,
        );
      }

      // STEP 4: Store in cache with enhanced metadata
      this.storageService.storeDashboardData(dashboardData, {
        fetched_at: new Date().toISOString(),
        source: "api",
        force_refresh: forceRefresh,
        file_count: dashboardData.recentFiles?.length || 0,
        storage_usage: dashboardData.summary?.storage_usage_percentage || 0,
      });

      this.notifyDashboardListeners("dashboard_loaded_from_api", {
        fromCache: false,
        summary: dashboardData.summary,
        hasStorageTrend: !!dashboardData.storageUsageTrend,
        recentFilesCount: dashboardData.recentFiles?.length || 0,
        forceRefresh: forceRefresh,
      });

      console.log(
        "[DashboardManager] Dashboard data retrieved and processed successfully",
      );
      return dashboardData;
    } catch (error) {
      console.error("[DashboardManager] Failed to get dashboard data:", error);

      this.notifyDashboardListeners("dashboard_load_failed", {
        error: error.message,
        forceRefresh: forceRefresh,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Process recent files to decrypt metadata
  async processRecentFiles(files) {
    if (!files || files.length === 0) return [];

    console.log("[DashboardManager] Processing", files.length, "recent files");

    const processedFiles = [];

    for (const file of files) {
      try {
        // Normalize the file
        const normalizedFile = this.fileCryptoService.normalizeFile(file);

        // Try to get collection key if available
        const collectionKey =
          this.collectionCryptoService.getCachedCollectionKey(
            file.collection_id,
          );

        let processedFile = normalizedFile;

        // If we have collection key, try to decrypt
        if (collectionKey) {
          processedFile = await this.fileCryptoService.decryptFileFromAPI(
            normalizedFile,
            collectionKey,
          );
        } else {
          // Mark as unable to decrypt
          processedFile = {
            ...normalizedFile,
            name: "[Collection not loaded]",
            _isDecrypted: false,
            _decryptionError: "Collection key not available",
          };
        }

        // Format for dashboard display
        const dashboardFile = {
          ...processedFile,
          fileName: processedFile.name || "[Encrypted]",
          type: this.getFileType(processedFile.name),
          size: processedFile.size
            ? { value: processedFile.size, unit: "Bytes" }
            : null,
          uploaded: this.formatRelativeTime(processedFile.created_at),
          uploadedTimestamp: processedFile.created_at,
        };

        processedFiles.push(dashboardFile);
      } catch (error) {
        console.error(
          "[DashboardManager] Failed to process file:",
          file.id,
          error,
        );
        processedFiles.push({
          ...file,
          fileName: "[Error processing file]",
          type: "Unknown",
          _decryptionError: error.message,
        });
      }
    }

    return processedFiles;
  }

  // Refresh dashboard data
  async refreshDashboardData() {
    return this.getDashboardData(true);
  }

  // 🔧 NEW: Smart refresh that checks if refresh is needed
  async smartRefresh() {
    const cacheAge = this.storageService.getCacheAge();

    // If cache is older than 2 minutes or doesn't exist, force refresh
    if (cacheAge === null || cacheAge > 2) {
      console.log(
        "[DashboardManager] Smart refresh: Cache is old, forcing refresh",
      );
      return this.getDashboardData(true);
    }

    console.log(
      "[DashboardManager] Smart refresh: Cache is recent, using cached data",
    );
    return this.getDashboardData(false);
  }

  // === Utility Methods ===

  // Format storage value for display
  formatStorageValue(storageObj) {
    if (!storageObj || typeof storageObj.value !== "number") {
      return "0 Bytes";
    }

    const value = storageObj.value;
    const unit = storageObj.unit || "Bytes";

    if (value === 0) return "0 Bytes";

    // Format number with appropriate decimal places
    const formatted = value % 1 === 0 ? value.toString() : value.toFixed(1);

    return `${formatted} ${unit}`;
  }

  // Calculate storage usage percentage
  calculateStorageUsagePercentage(used, limit) {
    if (!used || !limit || limit.value === 0) {
      return 0;
    }

    // Convert to same unit for calculation
    let usedBytes = used.value;
    let limitBytes = limit.value;

    const units = {
      Bytes: 1,
      KB: 1024,
      MB: 1024 * 1024,
      GB: 1024 * 1024 * 1024,
      TB: 1024 * 1024 * 1024 * 1024,
    };

    if (used.unit && units[used.unit]) {
      usedBytes = used.value * units[used.unit];
    }

    if (limit.unit && units[limit.unit]) {
      limitBytes = limit.value * units[limit.unit];
    }

    const percentage = (usedBytes / limitBytes) * 100;
    return Math.min(100, Math.max(0, Math.round(percentage)));
  }

  // Format relative time for recent files
  formatRelativeTime(dateString) {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "Unknown";

    try {
      const date = new Date(dateString);
      const now = new Date();
      const diffMs = now - date;
      const diffMins = Math.floor(diffMs / (1000 * 60));
      const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

      if (diffMins < 1) return "Just now";
      if (diffMins < 60) return `${diffMins} minutes ago`;
      if (diffHours < 24) return `${diffHours} hours ago`;
      if (diffDays < 30) return `${diffDays} days ago`;
      return date.toLocaleDateString();
    } catch {
      return "Invalid Date";
    }
  }

  // Format file size
  formatFileSize(bytes) {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }

  // Get file type from filename
  getFileType(filename) {
    if (!filename) return "Document";

    const extension = filename.split(".").pop()?.toLowerCase();
    if (!extension) return "Document";

    const typeMap = {
      // Images
      jpg: "Image",
      jpeg: "Image",
      png: "Image",
      gif: "Image",
      bmp: "Image",
      svg: "Image",
      webp: "Image",

      // Videos
      mp4: "Video",
      avi: "Video",
      mov: "Video",
      wmv: "Video",
      flv: "Video",
      webm: "Video",
      mkv: "Video",

      // Audio
      mp3: "Audio",
      wav: "Audio",
      flac: "Audio",
      aac: "Audio",
      ogg: "Audio",
      m4a: "Audio",

      // Documents
      pdf: "PDF",
      doc: "Word Document",
      docx: "Word Document",
      xls: "Spreadsheet",
      xlsx: "Spreadsheet",
      ppt: "Presentation",
      pptx: "Presentation",
      txt: "Text",

      // Archives
      zip: "Archive",
      rar: "Archive",
      "7z": "Archive",
      tar: "Archive",
      gz: "Archive",
    };

    return typeMap[extension] || "Document";
  }

  // === Cache Management ===

  // 🔧 ENHANCED: Clear all dashboard caches with reason logging
  clearAllCaches(reason = "manual") {
    console.log("[DashboardManager] Clearing all caches, reason:", reason);
    this.storageService.clearAllCaches();

    // Notify listeners about cache clearing
    this.notifyDashboardListeners("caches_cleared", {
      reason,
      timestamp: Date.now(),
    });
  }

  // Clear expired caches
  clearExpiredCaches() {
    return this.storageService.clearExpiredCaches();
  }

  // 🔧 NEW: Force cache refresh (clear and reload)
  async forceCacheRefresh(reason = "manual") {
    console.log(
      "[DashboardManager] Force cache refresh requested, reason:",
      reason,
    );
    this.clearAllCaches(reason);
    return this.getDashboardData(true);
  }

  // === Event Management ===

  // Add dashboard listener
  addDashboardListener(callback) {
    if (typeof callback === "function") {
      this.dashboardListeners.add(callback);
      console.log(
        "[DashboardManager] Dashboard listener added. Total listeners:",
        this.dashboardListeners.size,
      );
    }
  }

  // Remove dashboard listener
  removeDashboardListener(callback) {
    this.dashboardListeners.delete(callback);
    console.log(
      "[DashboardManager] Dashboard listener removed. Total listeners:",
      this.dashboardListeners.size,
    );
  }

  // Notify dashboard listeners
  notifyDashboardListeners(eventType, eventData) {
    console.log(
      `[DashboardManager] Notifying ${this.dashboardListeners.size} listeners of ${eventType}`,
    );

    this.dashboardListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error("[DashboardManager] Error in dashboard listener:", error);
      }
    });
  }

  // === Manager Status ===

  // Get manager status and information
  getManagerStatus() {
    const storageInfo = this.storageService.getStorageInfo();

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canGetDashboard: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      listenerCount: this.dashboardListeners.size,
      lastRefreshTime: this.lastRefreshTime,
      cacheAge: this.storageService.getCacheAge(),
      hasValidCache: this.storageService.hasValidDashboardCache(),
    };
  }

  // 🔧 NEW: Cleanup method for removing event listeners
  cleanup() {
    if (typeof window !== "undefined") {
      window.removeEventListener(
        "dashboardRefresh",
        this.handleDashboardRefreshEvent.bind(this),
      );
      window.removeEventListener("storage", this.handleStorageEvent.bind(this));
    }

    // Clear all listeners
    this.dashboardListeners.clear();

    console.log("[DashboardManager] Cleanup completed");
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "DashboardManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      fileCryptoService: this.fileCryptoService?.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
      cacheInvalidation: {
        lastRefreshTime: this.lastRefreshTime,
        minimumRefreshInterval: this.minimumRefreshInterval,
        hasEventListeners: true,
      },
    };
  }
}

export default DashboardManager;
