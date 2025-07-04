// SyncCollectionService.js
class SyncCollectionService {
  constructor(
    apiService, // Inject SyncCollectionAPIService
    storageService, // Inject SyncCollectionStorageService
  ) {
    if (!apiService || !storageService) {
      throw new Error(
        "SyncCollectionService requires an apiService and storageService",
      );
    }
    this.apiService = apiService;
    this.storageService = storageService;
    this.isLoading = false;
  }

  async getSyncCollections() {
    try {
      this.isLoading = true;
      // 1. Try to get SyncCollections from local storage
      const syncCollections = await this.storageService.getSyncCollections();
      if (syncCollections) {
        console.log(
          "[SyncCollectionService] Retrieved SyncCollections from local storage",
        );
        return syncCollections;
      }

      // 2. If not in local storage, fetch from API and save to local storage.
      const refreshedSyncCollections = await this.refreshSyncCollections();
      return refreshedSyncCollections;
    } catch (error) {
      console.error(
        "[SyncCollectionService] Error getting SyncCollections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  async refreshSyncCollections() {
    try {
      this.isLoading = true;
      // 1. Fetch SyncCollections from API
      const syncCollections = await this.apiService.fetchSyncCollections();

      // 2. Save SyncCollections to local storage
      await this.storageService.saveSyncCollections(syncCollections);

      console.log(
        "[SyncCollectionService] Refreshed and saved SyncCollections from API",
      );
      return syncCollections;
    } catch (error) {
      console.error(
        "[SyncCollectionService] Error refreshing SyncCollections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  async forceRefreshSyncCollections() {
    try {
      this.isLoading = true;
      // 1. Fetch SyncCollections from API
      const syncCollections = await this.apiService.syncAllCollections();

      // 2. Save SyncCollections to local storage
      await this.storageService.saveSyncCollections(syncCollections);

      console.log(
        "[SyncCollectionService] Force refreshed and saved SyncCollections from API",
      );
      return syncCollections;
    } catch (error) {
      console.error(
        "[SyncCollectionService] Error force refreshing SyncCollections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  getIsLoading() {
    return this.isLoading;
  }
}

export default SyncCollectionService;
