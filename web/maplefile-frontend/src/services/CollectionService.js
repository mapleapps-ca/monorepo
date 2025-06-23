// CollectionService for managing MapleFile collections (folders and albums)
// Supports all collection operations with end-to-end encryption

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

  // 1. Create Collection
  async createCollection(collectionData) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Creating new collection");

      const apiClient = await this.getApiClient();
      const collection = await apiClient.postMapleFile(
        "/collections",
        collectionData,
      );

      // Cache the new collection
      this.cache.set(collection.id, collection);
      console.log("[CollectionService] Collection created:", collection.id);

      return collection;
    } catch (error) {
      console.error("[CollectionService] Failed to create collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 2. Get Collection by ID
  async getCollection(collectionId) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Getting collection:", collectionId);

      // Check cache first
      if (this.cache.has(collectionId)) {
        console.log("[CollectionService] Collection found in cache");
        return this.cache.get(collectionId);
      }

      const apiClient = await this.getApiClient();
      const collection = await apiClient.getMapleFile(
        `/collections/${collectionId}`,
      );

      // Cache the collection
      this.cache.set(collectionId, collection);
      console.log("[CollectionService] Collection retrieved:", collectionId);

      return collection;
    } catch (error) {
      console.error("[CollectionService] Failed to get collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // 3. Update Collection
  async updateCollection(collectionId, updateData) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Updating collection:", collectionId);

      // Ensure the ID matches
      if (updateData.id && updateData.id !== collectionId) {
        throw new Error("Collection ID mismatch");
      }

      const apiClient = await this.getApiClient();
      const updatedCollection = await apiClient.putMapleFile(
        `/collections/${collectionId}`,
        { ...updateData, id: collectionId },
      );

      // Update cache
      this.cache.set(collectionId, updatedCollection);
      console.log("[CollectionService] Collection updated:", collectionId);

      return updatedCollection;
    } catch (error) {
      console.error("[CollectionService] Failed to update collection:", error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

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

  // 7. List User Collections (owned by authenticated user)
  async listUserCollections() {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Listing user collections");

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections");

      // Cache collections
      if (response.collections) {
        response.collections.forEach((collection) => {
          this.cache.set(collection.id, collection);
        });
      }

      console.log(
        "[CollectionService] User collections retrieved:",
        response.collections?.length || 0,
      );
      return response.collections || [];
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

  // 8. List Shared Collections
  async listSharedCollections() {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Listing shared collections");

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections/shared");

      // Cache collections
      if (response.collections) {
        response.collections.forEach((collection) => {
          this.cache.set(collection.id, collection);
        });
      }

      console.log(
        "[CollectionService] Shared collections retrieved:",
        response.collections?.length || 0,
      );
      return response.collections || [];
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

  // 9. Get Filtered Collections
  async getFilteredCollections(includeOwned = true, includeShared = false) {
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

      // Cache all collections
      const allCollections = [
        ...(response.owned_collections || []),
        ...(response.shared_collections || []),
      ];

      allCollections.forEach((collection) => {
        this.cache.set(collection.id, collection);
      });

      console.log("[CollectionService] Filtered collections retrieved:", {
        owned: response.owned_collections?.length || 0,
        shared: response.shared_collections?.length || 0,
        total: response.total_count || 0,
      });

      return response;
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

  // 10. Find Root Collections
  async findRootCollections() {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Finding root collections");

      const apiClient = await this.getApiClient();
      const response = await apiClient.getMapleFile("/collections/root");

      // Cache collections
      if (response.collections) {
        response.collections.forEach((collection) => {
          this.cache.set(collection.id, collection);
        });
      }

      console.log(
        "[CollectionService] Root collections found:",
        response.collections?.length || 0,
      );
      return response.collections || [];
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

  // 11. Find Collections by Parent
  async findCollectionsByParent(parentId) {
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

      // Cache collections
      if (response.collections) {
        response.collections.forEach((collection) => {
          this.cache.set(collection.id, collection);
        });
      }

      console.log(
        "[CollectionService] Child collections found:",
        response.collections?.length || 0,
      );
      return response.collections || [];
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

  // 12. Share Collection
  async shareCollection(collectionId, shareData) {
    try {
      this.isLoading = true;
      console.log("[CollectionService] Sharing collection:", collectionId);

      // Ensure collection_id is set
      const requestData = {
        ...shareData,
        collection_id: collectionId,
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

  // 13. Remove Member from Collection
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

  // 15. Sync Collections (for offline clients)
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

  // Utility method: Get all collections (owned + shared)
  async getAllCollections() {
    try {
      const result = await this.getFilteredCollections(true, true);
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

  // Utility method: Get collection hierarchy
  async getCollectionHierarchy(collectionId) {
    try {
      const collection = await this.getCollection(collectionId);
      const hierarchy = [collection];

      // Walk up the ancestor chain
      if (collection.ancestor_ids && collection.ancestor_ids.length > 0) {
        for (const ancestorId of collection.ancestor_ids.reverse()) {
          const ancestor = await this.getCollection(ancestorId);
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

  // Utility method: Get collection tree (with children)
  async getCollectionTree(parentId = null) {
    try {
      const collections = parentId
        ? await this.findCollectionsByParent(parentId)
        : await this.findRootCollections();

      // Recursively build tree
      const tree = await Promise.all(
        collections.map(async (collection) => {
          const children = await this.getCollectionTree(collection.id);
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

  // Utility method: Batch create collections
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

  // Cache management methods
  clearCache() {
    this.cache.clear();
    console.log("[CollectionService] Cache cleared");
  }

  getCachedCollection(collectionId) {
    return this.cache.get(collectionId) || null;
  }

  getCacheSize() {
    return this.cache.size;
  }

  // Check if service is loading
  isLoadingData() {
    return this.isLoading;
  }

  // Get debug information
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
