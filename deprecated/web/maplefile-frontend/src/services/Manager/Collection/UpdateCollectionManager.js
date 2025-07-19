// File: monorepo/web/maplefile-frontend/src/services/Manager/Collection/UpdateCollectionManager.js
// Update Collection Manager - Orchestrates API, Storage, and Crypto services for collection updates

import UpdateCollectionAPIService from "../../API/Collection/UpdateCollectionAPIService.js";
import UpdateCollectionStorageService from "../../Storage/Collection/UpdateCollectionStorageService.js";
import CollectionCryptoService from "../../Crypto/CollectionCryptoService.js";

class UpdateCollectionManager {
  constructor(authManager) {
    // UpdateCollectionManager depends on AuthManager and orchestrates API, Storage, and Crypto services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new UpdateCollectionAPIService(authManager);
    this.storageService = new UpdateCollectionStorageService();
    this.cryptoService = CollectionCryptoService; // Use singleton instance

    // Event listeners for collection update events
    this.collectionUpdateListeners = new Set();

    console.log(
      "[UpdateCollectionManager] Collection manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log(
        "[UpdateCollectionManager] Initializing collection manager...",
      );

      // Initialize crypto service
      await this.cryptoService.initialize();

      console.log(
        "[UpdateCollectionManager] Collection manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[UpdateCollectionManager] Failed to initialize collection manager:",
        error,
      );
    }
  }

  // === Collection Update with Encryption ===

  // Update collection with full E2EE encryption
  async updateCollection(collectionId, updateData, password = null) {
    try {
      this.isLoading = true;
      console.log("[UpdateCollectionManager] Starting collection update");
      console.log("[UpdateCollectionManager] Collection ID:", collectionId);
      console.log("[UpdateCollectionManager] Update data:", {
        hasNewName: !!updateData.name,
        hasNewType: !!updateData.collection_type,
        version: updateData.version,
        rotateKey: !!updateData.rotateCollectionKey,
      });

      // Validate input early
      if (!collectionId) {
        throw new Error("Collection ID is required");
      }

      if (updateData.version === undefined || updateData.version === null) {
        throw new Error(
          "Collection version is required for optimistic locking",
        );
      }

      // Get password for encryption if needed
      let userPassword = null;
      if (updateData.name || updateData.rotateCollectionKey) {
        userPassword = password || (await this.getUserPassword());
        if (!userPassword) {
          throw new Error(
            "Password required for collection encryption operations",
          );
        }
      }

      // Get current collection to get current data (required for encrypted_collection_key)
      let currentCollection = null;
      try {
        currentCollection = await this.getCurrentCollection(collectionId);
        console.log("[UpdateCollectionManager] Current collection retrieved:", {
          id: currentCollection.id,
          currentVersion: currentCollection.version,
          requestedVersion: updateData.version,
          collection_type: currentCollection.collection_type,
          hasEncryptedKey: !!currentCollection.encrypted_collection_key,
        });
      } catch (error) {
        throw new Error(
          `Cannot update collection: Unable to retrieve current collection data. ${error.message}`,
        );
      }

      if (!currentCollection || !currentCollection.encrypted_collection_key) {
        throw new Error(
          "Cannot update collection: Current collection or encrypted collection key not found. The collection may not exist or you may not have access.",
        );
      }

      // Prepare API update data - always include current encrypted_collection_key (backend requirement)
      const apiUpdateData = {
        id: collectionId,
        version: updateData.version,
        encrypted_collection_key: currentCollection.encrypted_collection_key, // Always required by backend
      };

      // Handle name encryption if new name provided
      if (updateData.name && updateData.name.trim()) {
        console.log("[UpdateCollectionManager] Encrypting new collection name");

        let collectionKey = this.storageService.getCollectionKey(collectionId);

        // If no cached key, decrypt from current collection data
        if (
          !collectionKey &&
          userPassword &&
          currentCollection.encrypted_collection_key
        ) {
          const userKeys = await this.cryptoService.getUserKeys();
          collectionKey = await this.cryptoService.decryptCollectionKey(
            currentCollection.encrypted_collection_key,
            userKeys.masterKey,
          );
          this.storageService.storeCollectionKey(collectionId, collectionKey);
        }

        if (!collectionKey) {
          throw new Error(
            "Collection key not available for name encryption. Password may be required.",
          );
        }

        // Encrypt new name with existing collection key
        const encryptedName = this.cryptoService.encryptCollectionName(
          updateData.name.trim(),
          collectionKey,
        );

        apiUpdateData.encrypted_name = encryptedName;
        console.log("[UpdateCollectionManager] Collection name encrypted");
      }

      // Handle collection type change
      if (updateData.collection_type) {
        const validTypes = ["folder", "album"];
        if (!validTypes.includes(updateData.collection_type)) {
          throw new Error(
            `Invalid collection type: ${updateData.collection_type}`,
          );
        }
        apiUpdateData.collection_type = updateData.collection_type;
        console.log(
          "[UpdateCollectionManager] Collection type updated:",
          updateData.collection_type,
        );
      }

      // Handle collection key rotation if requested (replaces the encrypted_collection_key)
      if (updateData.rotateCollectionKey && userPassword) {
        console.log("[UpdateCollectionManager] Rotating collection key");

        const userKeys = await this.cryptoService.getUserKeys();

        // Generate new collection key
        const newCollectionKey = this.cryptoService.generateCollectionKey();

        // Encrypt new collection key with user's master key
        const encryptedNewCollectionKey =
          await this.cryptoService.encryptCollectionKey(
            newCollectionKey,
            userKeys.masterKey,
          );

        // Include previous key for backwards compatibility
        encryptedNewCollectionKey.previous_keys = [
          currentCollection.encrypted_collection_key,
        ];
        encryptedNewCollectionKey.key_version =
          (currentCollection.encrypted_collection_key?.key_version || 1) + 1;

        // Replace the encrypted_collection_key with the new one
        apiUpdateData.encrypted_collection_key = encryptedNewCollectionKey;

        // Cache new collection key
        this.storageService.storeCollectionKey(collectionId, newCollectionKey);

        console.log(
          "[UpdateCollectionManager] Collection key rotated successfully",
        );
      }

      console.log("[UpdateCollectionManager] Sending update to API");

      // Update collection via API
      const updatedCollection = await this.apiService.updateCollection(
        collectionId,
        apiUpdateData,
      );

      console.log(
        "[UpdateCollectionManager] Collection updated via API successfully",
      );

      // Prepare decrypted collection for local storage
      const decryptedCollection = {
        ...updatedCollection,
        name:
          updateData.name ||
          (currentCollection ? currentCollection.name : "[Unknown]"), // Use new name or keep current
        _originalEncryptedName:
          updatedCollection.encrypted_name ||
          (currentCollection ? currentCollection.encrypted_name : null),
        _hasCollectionKey: true,
        _isDecrypted: true,
      };

      // Store in local storage
      this.storageService.storeUpdatedCollection(decryptedCollection);

      // Notify listeners
      this.notifyCollectionUpdateListeners("collection_updated", {
        collectionId,
        name: decryptedCollection.name,
        type: decryptedCollection.collection_type,
        version: updatedCollection.version,
        changes: {
          nameChanged: !!updateData.name,
          typeChanged: !!updateData.collection_type,
          keyRotated: !!updateData.rotateCollectionKey,
        },
      });

      console.log(
        "[UpdateCollectionManager] Collection updated successfully:",
        collectionId,
      );

      return {
        collection: decryptedCollection,
        collectionId,
        previousVersion: updateData.version,
        newVersion: updatedCollection.version,
        success: true,
      };
    } catch (error) {
      console.error(
        "[UpdateCollectionManager] Collection update failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionUpdateListeners("collection_update_failed", {
        collectionId,
        error: error.message,
        updateData,
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
        "[UpdateCollectionManager] Attempting to get current collection:",
        collectionId,
      );

      // Try to get from GetCollectionManager if available
      const getCollectionManager = await this.getCollectionManager();
      if (getCollectionManager) {
        try {
          console.log("[UpdateCollectionManager] Using GetCollectionManager");
          const result = await getCollectionManager.getCollection(collectionId);
          if (result && result.collection) {
            console.log(
              "[UpdateCollectionManager] Collection retrieved via GetCollectionManager",
            );
            return result.collection;
          }
        } catch (getError) {
          console.warn(
            "[UpdateCollectionManager] GetCollectionManager failed, trying direct API:",
            getError.message,
          );
        }
      }

      // Fallback: direct API call
      console.log("[UpdateCollectionManager] Using direct API call");
      const { default: GetCollectionAPIService } = await import(
        "../../API/Collection/GetCollectionAPIService.js"
      );
      const apiService = new GetCollectionAPIService(this.authManager);
      const collection = await apiService.getCollection(collectionId);

      console.log(
        "[UpdateCollectionManager] Collection retrieved via direct API",
      );
      return collection;
    } catch (error) {
      console.error(
        "[UpdateCollectionManager] Failed to get current collection:",
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
        "[UpdateCollectionManager] GetCollectionManager not available:",
        error,
      );
      return null;
    }
  }

  // === Collection Decryption ===

  // Decrypt collection data (for displaying updated collections)
  async decryptCollection(encryptedCollection, password = null) {
    try {
      console.log(
        "[UpdateCollectionManager] Decrypting collection:",
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
            "[UpdateCollectionManager] Collection key was cached during decryption",
          );
        } catch (keyError) {
          console.warn(
            "[UpdateCollectionManager] Could not verify collection key cache:",
            keyError,
          );
        }
      }

      return decryptedCollection;
    } catch (error) {
      console.error(
        "[UpdateCollectionManager] Failed to decrypt collection:",
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
        "[UpdateCollectionManager] Failed to get user password:",
        error,
      );
      return null;
    }
  }

  // === Collection Management ===

  // Get all updated collections
  getUpdatedCollections() {
    return this.storageService.getUpdatedCollections();
  }

  // Get updated collection by ID
  getUpdatedCollectionById(collectionId) {
    return this.storageService.getUpdatedCollectionById(collectionId);
  }

  // Get update history for a collection
  getUpdateHistory(collectionId = null) {
    return this.storageService.getUpdateHistory(collectionId);
  }

  // Search updated collections
  searchUpdatedCollections(searchTerm) {
    const collections = this.getUpdatedCollections();
    return this.storageService.searchUpdatedCollections(
      searchTerm,
      collections,
    );
  }

  // Remove updated collection
  async removeUpdatedCollection(collectionId) {
    try {
      console.log(
        "[UpdateCollectionManager] Removing updated collection:",
        collectionId,
      );

      const removed = this.storageService.removeUpdatedCollection(collectionId);

      if (removed) {
        this.notifyCollectionUpdateListeners("updated_collection_removed", {
          collectionId,
        });
      }

      return removed;
    } catch (error) {
      console.error(
        "[UpdateCollectionManager] Failed to remove updated collection:",
        error,
      );
      throw error;
    }
  }

  // Clear all updated collections
  async clearAllUpdatedCollections() {
    try {
      console.log("[UpdateCollectionManager] Clearing all updated collections");

      this.storageService.clearAllUpdatedCollections();

      this.notifyCollectionUpdateListeners(
        "all_updated_collections_cleared",
        {},
      );

      console.log("[UpdateCollectionManager] All updated collections cleared");
    } catch (error) {
      console.error(
        "[UpdateCollectionManager] Failed to clear updated collections:",
        error,
      );
      throw error;
    }
  }

  // === Event Management ===

  // Add collection update listener
  addCollectionUpdateListener(callback) {
    if (typeof callback === "function") {
      this.collectionUpdateListeners.add(callback);
      console.log(
        "[UpdateCollectionManager] Collection update listener added. Total listeners:",
        this.collectionUpdateListeners.size,
      );
    }
  }

  // Remove collection update listener
  removeCollectionUpdateListener(callback) {
    this.collectionUpdateListeners.delete(callback);
    console.log(
      "[UpdateCollectionManager] Collection update listener removed. Total listeners:",
      this.collectionUpdateListeners.size,
    );
  }

  // Notify collection update listeners
  notifyCollectionUpdateListeners(eventType, eventData) {
    console.log(
      `[UpdateCollectionManager] Notifying ${this.collectionUpdateListeners.size} listeners of ${eventType}`,
    );

    this.collectionUpdateListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[UpdateCollectionManager] Error in collection update listener:",
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
      canUpdateCollections: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      crypto: cryptoStatus,
      listenerCount: this.collectionUpdateListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "UpdateCollectionManager",
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

export default UpdateCollectionManager;
