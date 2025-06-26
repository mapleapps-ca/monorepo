// Updated CollectionService.js with password-based collection creation
import CollectionCryptoService from "./CollectionCryptoService.js";

class CollectionService {
  constructor() {
    this._apiClient = null;
    this.cache = new Map(); // Simple cache for collections
    this.isLoading = false;
  }

  // Import ApiClient for authenticated requests
  async getApiClient() {
    if (!this._apiClient) {
      const { default: ApiClient } = await import("./ApiClient.js");
      this._apiClient = ApiClient;
    }
    return this._apiClient;
  }

  // 1. Create Collection with Password (NEW METHOD)
  async createCollectionWithPassword(collectionData, password) {
    try {
      this.isLoading = true;
      console.log(
        "[CollectionService] Creating new collection with password-based encryption",
      );

      // Prepare encrypted data for API using password
      const { apiData, collectionKey, collectionId } =
        await CollectionCryptoService.prepareCollectionForAPIWithPassword(
          collectionData,
          password,
        );

      console.log(
        "[CollectionService] Collection data encrypted with password, sending to API",
      );

      // Clean up the API data - remove null values that Go doesn't handle well
      if (apiData.parent_id === null || apiData.parent_id === undefined) {
        delete apiData.parent_id;
      }
      if (!apiData.ancestor_ids || apiData.ancestor_ids.length === 0) {
        delete apiData.ancestor_ids;
      }

      const apiClient = await this.getApiClient();
      const encryptedCollection = await apiClient.postMapleFile(
        "/collections",
        apiData,
      );

      // Decrypt the response for local use (pass password for decryption)
      const decryptedCollection =
        await CollectionCryptoService.decryptCollectionFromAPI(
          encryptedCollection,
          password,
        );

      // Cache the collection key using the generated ID
      CollectionCryptoService.cacheCollectionKey(collectionId, collectionKey);

      // Cache the decrypted collection
      this.cache.set(collectionId, decryptedCollection);
      console.log(
        "[CollectionService] Collection created with password:",
        collectionId,
      );

      return decryptedCollection;
    } catch (error) {
      console.error(
        "[CollectionService] Failed to create collection with password:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 1a. Create Collection (original method - requires session keys)
  async createCollection(collectionData) {
    try {
      this.isLoading = true;
      console.log(
        "[CollectionService] Creating new collection with session keys",
      );

      // Check if we have session keys
      const { default: LocalStorageService } = await import(
        "./LocalStorageService.js"
      );
      if (!LocalStorageService.hasSessionKeys()) {
        throw new Error(
          "Session keys not available. Please use createCollectionWithPassword method.",
        );
      }

      // Prepare encrypted data for API
      const { apiData, collectionKey, collectionId } =
        await CollectionCryptoService.prepareCollectionForAPI(collectionData);

      console.log(
        "[CollectionService] Collection data encrypted, sending to API",
      );

      // Clean up the API data - remove null values that Go doesn't handle well
      if (apiData.parent_id === null || apiData.parent_id === undefined) {
        delete apiData.parent_id;
      }
      if (!apiData.ancestor_ids || apiData.ancestor_ids.length === 0) {
        delete apiData.ancestor_ids;
      }

      const apiClient = await this.getApiClient();
      const encryptedCollection = await apiClient.postMapleFile(
        "/collections",
        apiData,
      );

      // Decrypt the response for local use
      const decryptedCollection =
        await CollectionCryptoService.decryptCollectionFromAPI(
          encryptedCollection,
        );

      // Cache the collection key using the generated ID
      CollectionCryptoService.cacheCollectionKey(collectionId, collectionKey);

      // Cache the decrypted collection
      this.cache.set(collectionId, decryptedCollection);
      console.log("[CollectionService] Collection created:", collectionId);

      return decryptedCollection;
    } catch (error) {
      console.error("[CollectionService] Failed to create collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 2. Get Collection by ID (with optional password for decryption)
  async getCollection(collectionId, password = null) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Getting collection:", collectionId);

      // Check cache first
      if (this.cache.has(collectionId) && !password) {
        console.log("[CollectionService] Collection found in cache");
        return this.cache.get(collectionId);
      }

      const apiClient = await this.getApiClient();
      const encryptedCollection = await apiClient.getMapleFile(
        `/collections/${collectionId}`,
      );

      // Decrypt the collection
      const decryptedCollection =
        await CollectionCryptoService.decryptCollectionFromAPI(
          encryptedCollection,
          password,
        );

      // Cache the decrypted collection
      this.cache.set(collectionId, decryptedCollection);
      console.log(
        "[CollectionService] Collection retrieved and decrypted:",
        collectionId,
      );

      return decryptedCollection;
    } catch (error) {
      console.error("[CollectionService] Failed to get collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 3. Update Collection (with encryption)
  async updateCollection(collectionId, updateData) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Updating collection:", collectionId);

      // Get the cached collection to retrieve its key
      const cachedCollection = this.cache.get(collectionId);
      let collectionKey =
        cachedCollection?.collection_key ||
        CollectionCryptoService.getCachedCollectionKey(collectionId);

      if (!collectionKey && cachedCollection?._encrypted_collection_key) {
        // Try to decrypt the collection key - this will require session keys
        const userKeys = await CollectionCryptoService.getUserKeys();
        collectionKey = await CollectionCryptoService.decryptCollectionKey(
          cachedCollection._encrypted_collection_key,
          userKeys.masterKey,
        );
      }

      if (!collectionKey) {
        throw new Error("Collection key not found for update");
      }

      // Encrypt the updated name if provided
      const apiData = {
        id: collectionId,
        version: updateData.version || cachedCollection?.version || 1,
      };

      if (updateData.name) {
        apiData.encrypted_name = CollectionCryptoService.encryptCollectionName(
          updateData.name,
          collectionKey,
        );
      }

      if (updateData.collection_type) {
        apiData.collection_type = updateData.collection_type;
      }

      const apiClient = await this.getApiClient();
      const encryptedCollection = await apiClient.putMapleFile(
        `/collections/${collectionId}`,
        apiData,
      );

      // Decrypt the response
      const decryptedCollection =
        await CollectionCryptoService.decryptCollectionFromAPI(
          encryptedCollection,
        );

      // Update cache
      this.cache.set(collectionId, decryptedCollection);
      console.log("[CollectionService] Collection updated:", collectionId);

      return decryptedCollection;
    } catch (error) {
      console.error("[CollectionService] Failed to update collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 7. List User Collections (with optional password for decryption)
  async listUserCollections(password = null) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Listing user collections");

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections");

      // Decrypt all collections
      const decryptedCollections =
        await CollectionCryptoService.decryptCollections(
          response.collections || [],
          password,
        );

      // Cache collections
      decryptedCollections.forEach((collection) => {
        this.cache.set(collection.id, collection);
      });

      console.log(
        "[CollectionService] User collections retrieved and decrypted:",
        decryptedCollections.length,
      );
      return decryptedCollections;
    } catch (error) {
      console.error(
        "[CollectionService] Failed to list user collections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 8. List Shared Collections (with optional password for decryption)
  async listSharedCollections(password = null) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Listing shared collections");

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections/shared");

      // Decrypt all collections
      const decryptedCollections =
        await CollectionCryptoService.decryptCollections(
          response.collections || [],
          password,
        );

      // Cache collections
      decryptedCollections.forEach((collection) => {
        this.cache.set(collection.id, collection);
      });

      console.log(
        "[CollectionService] Shared collections retrieved and decrypted:",
        decryptedCollections.length,
      );
      return decryptedCollections;
    } catch (error) {
      console.error(
        "[CollectionService] Failed to list shared collections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 9. Get Filtered Collections (with optional password for decryption)
  async getFilteredCollections(
    includeOwned = true,
    includeShared = false,
    password = null,
  ) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Getting filtered collections", {
        includeOwned,
        includeShared,
      });

      const apiClient = await this.getApiClient();
      const params = new URLSearchParams({
        include_owned: includeOwned.toString(),
        include_shared: includeShared.toString(),
      });

      const response = await apiClient.getMapleFile(
        `/collections/filtered?${params}`,
      );

      // Decrypt all collections
      const decryptedOwned = await CollectionCryptoService.decryptCollections(
        response.owned_collections || [],
        password,
      );
      const decryptedShared = await CollectionCryptoService.decryptCollections(
        response.shared_collections || [],
        password,
      );

      // Cache all collections
      [...decryptedOwned, ...decryptedShared].forEach((collection) => {
        this.cache.set(collection.id, collection);
      });

      console.log(
        "[CollectionService] Filtered collections retrieved and decrypted:",
        {
          owned: decryptedOwned.length,
          shared: decryptedShared.length,
          total: response.total_count || 0,
        },
      );

      return {
        owned_collections: decryptedOwned,
        shared_collections: decryptedShared,
        total_count: response.total_count || 0,
      };
    } catch (error) {
      console.error(
        "[CollectionService] Failed to get filtered collections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 10. Find Root Collections (with optional password for decryption)
  async findRootCollections(password = null) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Finding root collections");

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections/root");

      // Decrypt all collections
      const decryptedCollections =
        await CollectionCryptoService.decryptCollections(
          response.collections || [],
          password,
        );

      // Cache collections
      decryptedCollections.forEach((collection) => {
        this.cache.set(collection.id, collection);
      });

      console.log(
        "[CollectionService] Root collections found and decrypted:",
        decryptedCollections.length,
      );
      return decryptedCollections;
    } catch (error) {
      console.error(
        "[CollectionService] Failed to find root collections:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 11. Find Collections by Parent (with optional password for decryption)
  async findCollectionsByParent(parentId, password = null) {
    try {
      this.isLoading = true;
      console.log(
        "[CollectionService] Finding collections by parent:",
        parentId,
      );

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile(
        `/collections-by-parent/${parentId}`,
      );

      // Decrypt all collections
      const decryptedCollections =
        await CollectionCryptoService.decryptCollections(
          response.collections || [],
          password,
        );

      // Cache collections
      decryptedCollections.forEach((collection) => {
        this.cache.set(collection.id, collection);
      });

      console.log(
        "[CollectionService] Child collections found and decrypted:",
        decryptedCollections.length,
      );
      return decryptedCollections;
    } catch (error) {
      console.error(
        "[CollectionService] Failed to find collections by parent:",
        error,
      );
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 12. Share Collection (with key encryption for recipient)
  async shareCollection(collectionId, shareData) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Sharing collection:", collectionId);

      // Get the collection key
      const collection = await this.getCollection(collectionId);
      const collectionKey =
        collection.collection_key ||
        CollectionCryptoService.getCachedCollectionKey(collectionId);

      if (!collectionKey) {
        throw new Error("Collection key not found for sharing");
      }

      // Encrypt collection key for recipient
      const encryptedKeyForRecipient =
        await CollectionCryptoService.encryptCollectionKeyForRecipient(
          collectionKey,
          shareData.recipient_public_key,
        );

      // Prepare request data
      const requestData = {
        collection_id: collectionId,
        recipient_id: shareData.recipient_id,
        recipient_email: shareData.recipient_email,
        permission_level: shareData.permission_level || "read_only",
        encrypted_collection_key: encryptedKeyForRecipient,
        share_with_descendants: shareData.share_with_descendants || false,
      };

      const apiClient = await this.getApiClient();
      const result = await apiClient.postMapleFile(
        `/collections/${collectionId}/share`,
        requestData,
      );

      // Invalidate cache for this collection since members changed
      this.cache.delete(collectionId);

      console.log(
        "[CollectionService] Collection shared successfully:",
        result,
      );
      return result;
    } catch (error) {
      console.error("[CollectionService] Failed to share collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Keep all other methods unchanged but ensure they handle decrypted data
  // 4. Delete Collection (Soft Delete)
  async deleteCollection(collectionId) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Deleting collection:", collectionId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.deleteMapleFile(
        `/collections/${collectionId}`,
      );

      // Remove from cache
      this.cache.delete(collectionId);
      CollectionCryptoService.cacheCollectionKey(collectionId, null);

      console.log("[CollectionService] Collection deleted:", collectionId);
      return result;
    } catch (error) {
      console.error("[CollectionService] Failed to delete collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 5. Archive Collection
  async archiveCollection(collectionId) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Archiving collection:", collectionId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.postMapleFile(
        `/collections/${collectionId}/archive`,
      );

      // Update cache if we have the collection
      if (this.cache.has(collectionId)) {
        const collection = this.cache.get(collectionId);
        collection.state = "archived";
        this.cache.set(collectionId, collection);
      }

      console.log("[CollectionService] Collection archived:", collectionId);
      return result;
    } catch (error) {
      console.error("[CollectionService] Failed to archive collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 6. Restore Collection
  async restoreCollection(collectionId) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Restoring collection:", collectionId);

      const apiClient = await this.getApiClient();
      const result = await apiClient.postMapleFile(
        `/collections/${collectionId}/restore`,
      );

      // Update cache if we have the collection
      if (this.cache.has(collectionId)) {
        const collection = this.cache.get(collectionId);
        collection.state = "active";
        this.cache.set(collectionId, collection);
      }

      console.log("[CollectionService] Collection restored:", collectionId);
      return result;
    } catch (error) {
      console.error("[CollectionService] Failed to restore collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 13. Remove Member
  async removeMember(collectionId, recipientId, removeFromDescendants = true) {
    try {
      this.isLoading = true;
      console.log(
        "[CollectionService] Removing member from collection:",
        collectionId,
      );

      const requestData = {
        collection_id: collectionId,
        recipient_id: recipientId,
        remove_from_descendants: removeFromDescendants,
      };

      const apiClient = await this.getApiClient();
      const result = await apiClient.deleteMapleFile(
        `/collections/${collectionId}/members`,
        { body: JSON.stringify(requestData) },
      );

      // Invalidate cache for this collection since members changed
      this.cache.delete(collectionId);

      console.log("[CollectionService] Member removed successfully");
      return result;
    } catch (error) {
      console.error("[CollectionService] Failed to remove member:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 14. Move Collection
  async moveCollection(collectionId, moveData) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Moving collection:", collectionId);

      const requestData = {
        ...moveData,
        collection_id: collectionId,
      };

      const apiClient = await this.getApiClient();
      const result = await apiClient.postMapleFile(
        `/collections/${collectionId}/move`,
        requestData,
      );

      // Invalidate cache for this collection and its parent
      this.cache.delete(collectionId);
      if (moveData.new_parent_id) {
        this.cache.delete(moveData.new_parent_id);
      }

      console.log("[CollectionService] Collection moved successfully");
      return result;
    } catch (error) {
      console.error("[CollectionService] Failed to move collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 15. Sync Collections
  async syncCollections(cursor = null, limit = 1000) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Syncing collections", { cursor, limit });

      const apiClient = await this.getApiClient();
      const params = new URLSearchParams({ limit: limit.toString() });

      if (cursor) {
        params.append("cursor", cursor);
      }

      const response = await apiClient.getMapleFile(
        `/sync/collections?${params}`,
      );

      console.log("[CollectionService] Collections synced:", {
        count: response.collections?.length || 0,
        hasMore: response.has_more || false,
      });

      return response;
    } catch (error) {
      console.error("[CollectionService] Failed to sync collections:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Utility methods remain the same
  async getAllCollections(password = null) {
    try {
      const result = await this.getFilteredCollections(true, true, password);
      return [
        ...(result.owned_collections || []),
        ...(result.shared_collections || []),
      ];
    } catch (error) {
      console.error(
        "[CollectionService] Failed to get all collections:",
        error,
      );
      throw error;
    }
  }

  async getCollectionHierarchy(collectionId, password = null) {
    try {
      const collection = await this.getCollection(collectionId, password);
      const hierarchy = [collection];

      if (collection.ancestor_ids && collection.ancestor_ids.length > 0) {
        for (const ancestorId of collection.ancestor_ids.reverse()) {
          const ancestor = await this.getCollection(ancestorId, password);
          hierarchy.unshift(ancestor);
        }
      }

      return hierarchy;
    } catch (error) {
      console.error(
        "[CollectionService] Failed to get collection hierarchy:",
        error,
      );
      throw error;
    }
  }

  async getCollectionTree(parentId = null, password = null) {
    try {
      const collections = parentId
        ? await this.findCollectionsByParent(parentId, password)
        : await this.findRootCollections(password);

      const tree = await Promise.all(
        collections.map(async (collection) => {
          const children = await this.getCollectionTree(
            collection.id,
            password,
          );
          return {
            ...collection,
            children,
          };
        }),
      );

      return tree;
    } catch (error) {
      console.error(
        "[CollectionService] Failed to get collection tree:",
        error,
      );
      throw error;
    }
  }

  async batchCreateCollections(collectionsData) {
    try {
      console.log(
        "[CollectionService] Batch creating collections:",
        collectionsData.length,
      );

      const results = await Promise.all(
        collectionsData.map((data) => this.createCollection(data)),
      );

      console.log("[CollectionService] Batch creation completed");
      return results;
    } catch (error) {
      console.error("[CollectionService] Batch creation failed:", error);
      throw error;
    }
  }

  clearCache() {
    this.cache.clear();
    CollectionCryptoService.clearCollectionKeyCache();
    console.log("[CollectionService] Cache cleared");
  }

  getCachedCollection(collectionId) {
    return this.cache.get(collectionId) || null;
  }

  getCacheSize() {
    return this.cache.size;
  }

  isLoadingData() {
    return this.isLoading;
  }

  getDebugInfo() {
    return {
      cacheSize: this.cache.size,
      cachedCollectionIds: Array.from(this.cache.keys()),
      isLoading: this.isLoading,
    };
  }
}

// Export singleton instance
export default new CollectionService();
