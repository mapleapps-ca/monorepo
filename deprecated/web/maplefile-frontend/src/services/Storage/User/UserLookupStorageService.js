// File: monorepo/web/maplefile-frontend/src/services/Storage/User/UserLookupStorageService.js
// User Lookup Storage Service - Handles localStorage operations for user lookup results

class UserLookupStorageService {
  constructor() {
    this.STORAGE_KEYS = {
      USER_LOOKUPS: "mapleapps_user_lookups",
      USER_LOOKUP_METADATA: "mapleapps_user_lookup_metadata",
    };

    // Cache expiry time (1 hour for public keys)
    this.CACHE_EXPIRY_MS = 60 * 60 * 1000;

    console.log("[UserLookupStorageService] Storage service initialized");
  }

  // === User Lookup Caching Operations ===

  // Store user lookup result in cache
  storeUserLookup(userLookup) {
    try {
      const existingLookups = this.getUserLookups();

      // Remove any existing lookup for the same email
      const filteredLookups = existingLookups.filter(
        (lookup) => lookup.email !== userLookup.email,
      );

      // Add the new lookup with cache metadata
      const cachedLookup = {
        ...userLookup,
        cached_at: new Date().toISOString(),
        cache_expiry: new Date(Date.now() + this.CACHE_EXPIRY_MS).toISOString(),
      };

      filteredLookups.push(cachedLookup);

      localStorage.setItem(
        this.STORAGE_KEYS.USER_LOOKUPS,
        JSON.stringify(filteredLookups),
      );

      console.log(
        "[UserLookupStorageService] User lookup cached:",
        userLookup.email,
      );

      // Update cache metadata
      this.updateCacheMetadata();
    } catch (error) {
      console.error(
        "[UserLookupStorageService] Failed to cache user lookup:",
        error,
      );
    }
  }

  // Get all user lookups from cache
  getUserLookups(includeExpired = false) {
    try {
      const stored = localStorage.getItem(this.STORAGE_KEYS.USER_LOOKUPS);
      const lookups = stored ? JSON.parse(stored) : [];

      if (!includeExpired) {
        // Filter out expired lookups
        const now = new Date();
        return lookups.filter((lookup) => {
          if (!lookup.cache_expiry) return true; // Keep lookups without expiry
          return new Date(lookup.cache_expiry) > now;
        });
      }

      return lookups;
    } catch (error) {
      console.error(
        "[UserLookupStorageService] Failed to get cached user lookups:",
        error,
      );
      return [];
    }
  }

  // Get user lookup by email from cache
  getCachedUserLookup(email) {
    const sanitizedEmail = email.toLowerCase().trim();
    const lookups = this.getUserLookups();
    const lookup = lookups.find((l) => l.email === sanitizedEmail);

    if (lookup) {
      console.log(
        `[UserLookupStorageService] User lookup found in cache: ${sanitizedEmail}`,
      );
      return lookup;
    }

    console.log(
      `[UserLookupStorageService] User lookup not found in cache: ${sanitizedEmail}`,
    );
    return null;
  }

  // Check if user lookup is cached and not expired
  isUserLookupCached(email) {
    const lookup = this.getCachedUserLookup(email);
    return !!lookup;
  }

  // Get user lookup cache status
  getUserLookupCacheStatus(email) {
    const sanitizedEmail = email.toLowerCase().trim();
    const lookups = this.getUserLookups(true); // Include expired
    const lookup = lookups.find((l) => l.email === sanitizedEmail);

    if (!lookup) {
      return {
        cached: false,
        expired: false,
        cachedAt: null,
        expiresAt: null,
      };
    }

    const now = new Date();
    const expiryDate = new Date(lookup.cache_expiry);
    const expired = expiryDate <= now;

    return {
      cached: true,
      expired: expired,
      cachedAt: lookup.cached_at,
      expiresAt: lookup.cache_expiry,
      timeUntilExpiry: expired ? 0 : expiryDate.getTime() - now.getTime(),
    };
  }

  // === Cache Management ===

  // Clear expired user lookups from cache
  clearExpiredUserLookups() {
    try {
      const allLookups = this.getUserLookups(true); // Include expired
      const validLookups = this.getUserLookups(false); // Exclude expired
      const expiredCount = allLookups.length - validLookups.length;

      if (expiredCount > 0) {
        localStorage.setItem(
          this.STORAGE_KEYS.USER_LOOKUPS,
          JSON.stringify(validLookups),
        );

        console.log(
          `[UserLookupStorageService] Cleared ${expiredCount} expired user lookups`,
        );
        this.updateCacheMetadata();
      }

      return expiredCount;
    } catch (error) {
      console.error(
        "[UserLookupStorageService] Failed to clear expired user lookups:",
        error,
      );
      return 0;
    }
  }

  // Remove specific user lookup from cache
  removeFromCache(email) {
    try {
      const sanitizedEmail = email.toLowerCase().trim();
      const lookups = this.getUserLookups(true); // Include expired
      const filteredLookups = lookups.filter((l) => l.email !== sanitizedEmail);

      localStorage.setItem(
        this.STORAGE_KEYS.USER_LOOKUPS,
        JSON.stringify(filteredLookups),
      );

      console.log(
        "[UserLookupStorageService] User lookup removed from cache:",
        sanitizedEmail,
      );
      this.updateCacheMetadata();
      return true;
    } catch (error) {
      console.error(
        "[UserLookupStorageService] Failed to remove user lookup from cache:",
        error,
      );
      return false;
    }
  }

  // Clear all cached user lookups
  clearAllCachedUserLookups() {
    try {
      localStorage.removeItem(this.STORAGE_KEYS.USER_LOOKUPS);
      localStorage.removeItem(this.STORAGE_KEYS.USER_LOOKUP_METADATA);

      console.log("[UserLookupStorageService] All cached user lookups cleared");
    } catch (error) {
      console.error(
        "[UserLookupStorageService] Failed to clear cached user lookups:",
        error,
      );
    }
  }

  // === Cache Metadata ===

  // Update cache metadata
  updateCacheMetadata() {
    try {
      const lookups = this.getUserLookups(true); // Include expired
      const validLookups = this.getUserLookups(false); // Exclude expired

      const metadata = {
        totalCached: lookups.length,
        validCached: validLookups.length,
        expiredCached: lookups.length - validLookups.length,
        lastUpdated: new Date().toISOString(),
        cacheExpiryMs: this.CACHE_EXPIRY_MS,
      };

      localStorage.setItem(
        this.STORAGE_KEYS.USER_LOOKUP_METADATA,
        JSON.stringify(metadata),
      );
    } catch (error) {
      console.error(
        "[UserLookupStorageService] Failed to update cache metadata:",
        error,
      );
    }
  }

  // Get cache metadata
  getCacheMetadata() {
    try {
      const stored = localStorage.getItem(
        this.STORAGE_KEYS.USER_LOOKUP_METADATA,
      );
      return stored ? JSON.parse(stored) : null;
    } catch (error) {
      console.error(
        "[UserLookupStorageService] Failed to get cache metadata:",
        error,
      );
      return null;
    }
  }

  // === User Lookup Search ===

  // Search cached user lookups
  searchCachedUserLookups(searchTerm) {
    if (!searchTerm) return this.getUserLookups();

    const lookups = this.getUserLookups();
    const term = searchTerm.toLowerCase();

    return lookups.filter((lookup) => {
      // Search in email
      if (lookup.email && lookup.email.toLowerCase().includes(term)) {
        return true;
      }

      // Search in name
      if (lookup.name && lookup.name.toLowerCase().includes(term)) {
        return true;
      }

      // Search in user ID (partial)
      if (lookup.user_id && lookup.user_id.toLowerCase().includes(term)) {
        return true;
      }

      return false;
    });
  }

  // === User Lookup Statistics ===

  // Get cache statistics
  getCacheStats() {
    const allLookups = this.getUserLookups(true);
    const validLookups = this.getUserLookups(false);
    const metadata = this.getCacheMetadata();

    const stats = {
      total: allLookups.length,
      valid: validLookups.length,
      expired: allLookups.length - validLookups.length,
      cacheExpiryMinutes: this.CACHE_EXPIRY_MS / (60 * 1000),
      lastUpdated: metadata?.lastUpdated || null,
    };

    return stats;
  }

  // === Storage Information ===

  // Get storage information
  getStorageInfo() {
    const stats = this.getCacheStats();

    return {
      cachedUserLookupsCount: stats.valid,
      expiredUserLookupsCount: stats.expired,
      stats,
      storageKeys: Object.keys(this.STORAGE_KEYS),
      hasCachedUserLookups: stats.valid > 0,
      cacheExpiryMs: this.CACHE_EXPIRY_MS,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "UserLookupStorageService",
      storageInfo: this.getStorageInfo(),
      cacheMetadata: this.getCacheMetadata(),
    };
  }

  // === User Lookup Validation ===

  // Validate user lookup data structure
  validateUserLookupStructure(userLookup) {
    const requiredFields = [
      "user_id",
      "email",
      "name",
      "public_key_in_base64",
      "verification_id",
    ];
    const errors = [];

    requiredFields.forEach((field) => {
      if (!userLookup[field]) {
        errors.push(`Missing required field: ${field}`);
      }
    });

    return {
      isValid: errors.length === 0,
      errors,
    };
  }
}

export default UserLookupStorageService;
