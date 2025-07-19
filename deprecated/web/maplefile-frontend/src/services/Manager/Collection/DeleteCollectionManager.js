// File: monorepo/web/maplefile-frontend/src/services/Manager/Collection/DeleteCollectionManager.js
// Delete Collection Manager - Orchestrates API, Storage, and Crypto services for collection deletion

import DeleteCollectionAPIService from "../../API/Collection/DeleteCollectionAPIService.js";
import DeleteCollectionStorageService from "../../Storage/Collection/DeleteCollectionStorageService.js";
import CollectionCryptoService from "../../Crypto/CollectionCryptoService.js";

class DeleteCollectionManager {
  constructor(authManager) {
    // DeleteCollectionManager depends on AuthManager and orchestrates API, Storage, and Crypto services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new DeleteCollectionAPIService(authManager);
    this.storageService = new DeleteCollectionStorageService();
    this.cryptoService = CollectionCryptoService; // Use singleton instance

    // Event listeners for collection deletion events
    this.collectionDeletionListeners = new Set();

    console.log(
      "[DeleteCollectionManager] Collection manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log(
        "[DeleteCollectionManager] Initializing collection manager...",
      );

      // Initialize crypto service
      await this.cryptoService.initialize();

      console.log(
        "[DeleteCollectionManager] Collection manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Failed to initialize collection manager:",
        error,
      );
    }
  }

  // === Collection Deletion ===

  // Delete collection (soft delete)
  async deleteCollection(collectionId, password = null) {
    try {
      this.isLoading = true;
      console.log("[DeleteCollectionManager] Starting collection deletion");
      console.log("[DeleteCollectionManager] Collection ID:", collectionId);

      // Validate input early
      if (!collectionId) {
        throw new Error("Collection ID is required");
      }

      // First, try to get the collection details before deletion (for logging and storage)
      let collectionToDelete = null;
      try {
        collectionToDelete = await this.getCurrentCollection(collectionId);
        console.log("[DeleteCollectionManager] Collection details retrieved:", {
          id: collectionToDelete.id,
          type: collectionToDelete.collection_type,
          hasEncryptedName: !!collectionToDelete.encrypted_name,
        });

        // Try to decrypt collection name for better logging
        if (collectionToDelete && collectionToDelete.encrypted_name) {
          try {
            const decryptedCollection = await this.decryptCollection(
              collectionToDelete,
              password,
            );
            collectionToDelete.name = decryptedCollection.name;
          } catch (decryptError) {
            console.warn(
              "[DeleteCollectionManager] Could not decrypt collection name:",
              decryptError.message,
            );
            collectionToDelete.name = "[Encrypted]";
          }
        }
      } catch (getError) {
        console.warn(
          "[DeleteCollectionManager] Could not retrieve collection details before deletion:",
          getError.message,
        );
        // Continue with deletion even if we can't get details
        collectionToDelete = {
          id: collectionId,
          name: "[Unknown]",
          collection_type: "unknown",
        };
      }

      console.log("[DeleteCollectionManager] Sending deletion to API");

      // Delete collection via API
      const deletionResult =
        await this.apiService.deleteCollection(collectionId);

      console.log(
        "[DeleteCollectionManager] Collection deleted via API successfully",
      );

      // Store in local storage as deleted
      this.storageService.storeDeletedCollection(collectionToDelete);

      // Notify listeners
      this.notifyCollectionDeletionListeners("collection_deleted", {
        collectionId,
        name: collectionToDelete.name,
        type: collectionToDelete.collection_type,
      });

      console.log(
        "[DeleteCollectionManager] Collection deleted successfully:",
        collectionId,
      );

      return {
        collection: collectionToDelete,
        collectionId,
        success: true,
        result: deletionResult,
      };
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Collection deletion failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionDeletionListeners("collection_deletion_failed", {
        collectionId,
        error: error.message,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Delete multiple collections (batch operation)
  async deleteCollections(collectionIds, password = null) {
    try {
      this.isLoading = true;
      console.log(
        "[DeleteCollectionManager] Starting batch collection deletion:",
        collectionIds.length,
      );

      if (!Array.isArray(collectionIds) || collectionIds.length === 0) {
        throw new Error("Collection IDs array is required");
      }

      const results = [];
      const errors = [];

      for (const collectionId of collectionIds) {
        try {
          const result = await this.deleteCollection(collectionId, password);
          results.push(result);
        } catch (error) {
          errors.push({ collectionId, error: error.message });
          console.warn(
            `[DeleteCollectionManager] Failed to delete collection ${collectionId}:`,
            error.message,
          );
        }
      }

      // Notify listeners
      this.notifyCollectionDeletionListeners("multiple_collections_deleted", {
        requestedCount: collectionIds.length,
        successCount: results.length,
        errorCount: errors.length,
      });

      console.log(
        `[DeleteCollectionManager] Batch deletion completed: ${results.length} successful, ${errors.length} failed`,
      );

      return {
        successful: results,
        errors: errors,
        successCount: results.length,
        errorCount: errors.length,
        success: true,
      };
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Batch collection deletion failed:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Restore collection from soft delete
  async restoreCollection(collectionId) {
    try {
      this.isLoading = true;
      console.log("[DeleteCollectionManager] Starting collection restoration");
      console.log("[DeleteCollectionManager] Collection ID:", collectionId);

      // Get deleted collection details
      const deletedCollection =
        this.storageService.getDeletedCollectionById(collectionId);
      if (!deletedCollection) {
        throw new Error("Deleted collection not found in local storage");
      }

      console.log("[DeleteCollectionManager] Sending restoration to API");

      // Restore collection via API
      const restorationResult =
        await this.apiService.restoreCollection(collectionId);

      console.log(
        "[DeleteCollectionManager] Collection restored via API successfully",
      );

      // Remove from deleted collections storage
      this.storageService.restoreDeletedCollection(collectionId);

      // Notify listeners
      this.notifyCollectionDeletionListeners("collection_restored", {
        collectionId,
        name: deletedCollection.name || "[Encrypted]",
        type: deletedCollection.collection_type,
      });

      console.log(
        "[DeleteCollectionManager] Collection restored successfully:",
        collectionId,
      );

      return {
        collection: deletedCollection,
        collectionId,
        success: true,
        result: restorationResult,
      };
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Collection restoration failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionDeletionListeners("collection_restoration_failed", {
        collectionId,
        error: error.message,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Helper Methods ===

  // Get current collection data (try cache first, then API)
  async getCurrentCollection(collectionId) {
    try {
      console.log(
        "[DeleteCollectionManager] Attempting to get current collection:",
        collectionId,
      );

      // Try to get from GetCollectionManager if available
      const getCollectionManager = await this.getCollectionManager();
      if (getCollectionManager) {
        try {
          console.log("[DeleteCollectionManager] Using GetCollectionManager");
          const result = await getCollectionManager.getCollection(collectionId);
          if (result && result.collection) {
            console.log(
              "[DeleteCollectionManager] Collection retrieved via GetCollectionManager",
            );
            return result.collection;
          }
        } catch (getError) {
          console.warn(
            "[DeleteCollectionManager] GetCollectionManager failed, trying direct API:",
            getError.message,
          );
        }
      }

      // Fallback: direct API call
      console.log("[DeleteCollectionManager] Using direct API call");
      const { default: GetCollectionAPIService } = await import(
        "../../API/Collection/GetCollectionAPIService.js"
      );
      const apiService = new GetCollectionAPIService(this.authManager);
      const collection = await apiService.getCollection(collectionId);

      console.log(
        "[DeleteCollectionManager] Collection retrieved via direct API",
      );
      return collection;
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Failed to get current collection:",
        error,
      );
      throw new Error(`Failed to get current collection: ${error.message}`);
    }
  }

  // Get GetCollectionManager instance if available
  async getCollectionManager() {
    try {
      // Try to get from service context
      if (window.mapleAppsServices?.getCollectionManager) {
        return window.mapleAppsServices.getCollectionManager;
      }
      return null;
    } catch (error) {
      console.warn(
        "[DeleteCollectionManager] GetCollectionManager not available:",
        error,
      );
      return null;
    }
  }

  // === Collection Decryption ===

  // Decrypt collection data (for displaying deleted collections)
  async decryptCollection(encryptedCollection, password = null) {
    try {
      console.log(
        "[DeleteCollectionManager] Decrypting collection:",
        encryptedCollection.id,
      );

      // Get collection key from storage
      let collectionKey = this.storageService.getCollectionKey(
        encryptedCollection.id,
      );

      // Use CollectionCryptoService to decrypt the collection
      const decryptedCollection =
        await this.cryptoService.decryptCollectionFromAPI(
          encryptedCollection,
          collectionKey,
        );

      // If we successfully decrypted and didn't have the key cached, cache it now
      if (decryptedCollection._isDecrypted && !collectionKey) {
        try {
          // The crypto service already cached the key during decryption
          console.log(
            "[DeleteCollectionManager] Collection key was cached during decryption",
          );
        } catch (keyError) {
          console.warn(
            "[DeleteCollectionManager] Could not verify collection key cache:",
            keyError,
          );
        }
      }

      return decryptedCollection;
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Failed to decrypt collection:",
        error,
      );
      throw error;
    }
  }

  // === Password Management ===

  // Get user password from password storage service
  async getUserPassword() {
    try {
      const { default: passwordStorageService } = await import(
        "../../PasswordStorageService.js"
      );
      return passwordStorageService.getPassword();
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Failed to get user password:",
        error,
      );
      return null;
    }
  }

  // === Collection Management ===

  // Get all deleted collections
  getDeletedCollections() {
    return this.storageService.getDeletedCollections();
  }

  // Get deleted collection by ID
  getDeletedCollectionById(collectionId) {
    return this.storageService.getDeletedCollectionById(collectionId);
  }

  // Get deletion history for a collection
  getDeletionHistory(collectionId = null) {
    return this.storageService.getDeletionHistory(collectionId);
  }

  // Search deleted collections
  searchDeletedCollections(searchTerm) {
    const collections = this.getDeletedCollections();
    return this.storageService.searchDeletedCollections(
      searchTerm,
      collections,
    );
  }

  // Permanently remove deleted collection from local storage
  async permanentlyRemoveCollection(collectionId) {
    try {
      console.log(
        "[DeleteCollectionManager] Permanently removing deleted collection:",
        collectionId,
      );

      const removed =
        this.storageService.permanentlyRemoveCollection(collectionId);

      if (removed) {
        this.notifyCollectionDeletionListeners(
          "deleted_collection_permanently_removed",
          {
            collectionId,
          },
        );
      }

      return removed;
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Failed to permanently remove deleted collection:",
        error,
      );
      throw error;
    }
  }

  // Clear all deleted collections
  async clearAllDeletedCollections() {
    try {
      console.log("[DeleteCollectionManager] Clearing all deleted collections");

      this.storageService.clearAllDeletedCollections();

      this.notifyCollectionDeletionListeners(
        "all_deleted_collections_cleared",
        {},
      );

      console.log("[DeleteCollectionManager] All deleted collections cleared");
    } catch (error) {
      console.error(
        "[DeleteCollectionManager] Failed to clear deleted collections:",
        error,
      );
      throw error;
    }
  }

  // === Event Management ===

  // Add collection deletion listener
  addCollectionDeletionListener(callback) {
    if (typeof callback === "function") {
      this.collectionDeletionListeners.add(callback);
      console.log(
        "[DeleteCollectionManager] Collection deletion listener added. Total listeners:",
        this.collectionDeletionListeners.size,
      );
    }
  }

  // Remove collection deletion listener
  removeCollectionDeletionListener(callback) {
    this.collectionDeletionListeners.delete(callback);
    console.log(
      "[DeleteCollectionManager] Collection deletion listener removed. Total listeners:",
      this.collectionDeletionListeners.size,
    );
  }

  // Notify collection deletion listeners
  notifyCollectionDeletionListeners(eventType, eventData) {
    console.log(
      `[DeleteCollectionManager] Notifying ${this.collectionDeletionListeners.size} listeners of ${eventType}`,
    );

    this.collectionDeletionListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[DeleteCollectionManager] Error in collection deletion listener:",
          error,
        );
      }
    });
  }

  // === Manager Status ===

  // Get manager status and information
  getManagerStatus() {
    const storageInfo = this.storageService.getStorageInfo();
    const cryptoStatus = this.cryptoService.getStatus();

    return {
      isAuthenticated: this.authManager.isAuthenticated(),
      isLoading: this.isLoading,
      canDeleteCollections: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      crypto: cryptoStatus,
      listenerCount: this.collectionDeletionListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "DeleteCollectionManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      cryptoService: this.cryptoService.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default DeleteCollectionManager;
