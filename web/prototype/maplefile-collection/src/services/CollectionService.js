/**
 * Collection service implementing CRUD operations
 * Follows Single Responsibility Principle - only handles collection operations
 * Depends on ApiService abstraction (Dependency Inversion Principle)
 */
class CollectionService {
  constructor(apiService) {
    this.apiService = apiService;
  }

  /**
   * Create a new collection
   * @param {object} collectionData
   * @returns {Promise<object>}
   */
  async createCollection(collectionData) {
    const payload = {
      encrypted_name: collectionData.encrypted_name,
      collection_type: collectionData.collection_type || "folder",
      encrypted_collection_key: collectionData.encrypted_collection_key,
      parent_id: collectionData.parent_id || null,
      ancestor_ids: collectionData.ancestor_ids || [],
      members: collectionData.members || [],
    };

    return await this.apiService.post("/collections", payload);
  }

  /**
   * Get a specific collection by ID
   * @param {string} collectionId
   * @returns {Promise<object>}
   */
  async getCollection(collectionId) {
    return await this.apiService.get(`/collections/${collectionId}`);
  }

  /**
   * Update an existing collection
   * @param {string} collectionId
   * @param {object} updateData
   * @returns {Promise<object>}
   */
  async updateCollection(collectionId, updateData) {
    const payload = {
      id: collectionId,
      encrypted_name: updateData.encrypted_name,
      collection_type: updateData.collection_type,
      encrypted_collection_key: updateData.encrypted_collection_key,
      version: updateData.version,
    };

    return await this.apiService.put(`/collections/${collectionId}`, payload);
  }

  /**
   * Soft delete a collection
   * @param {string} collectionId
   * @returns {Promise<object>}
   */
  async deleteCollection(collectionId) {
    return await this.apiService.delete(`/collections/${collectionId}`);
  }

  /**
   * Get all collections owned by the user
   * @returns {Promise<object>}
   */
  async getUserCollections() {
    return await this.apiService.get("/collections");
  }

  /**
   * Get all collections shared with the user
   * @returns {Promise<object>}
   */
  async getSharedCollections() {
    return await this.apiService.get("/collections/shared");
  }

  /**
   * Get filtered collections based on ownership and sharing
   * @param {object} filters
   * @returns {Promise<object>}
   */
  async getFilteredCollections(filters = {}) {
    const params = {
      include_owned: filters.include_owned ?? true,
      include_shared: filters.include_shared ?? false,
    };

    return await this.apiService.get("/collections/filtered", params);
  }

  /**
   * Get root-level collections (no parent)
   * @returns {Promise<object>}
   */
  async getRootCollections() {
    return await this.apiService.get("/collections/root");
  }

  /**
   * Get child collections of a parent
   * @param {string} parentId
   * @returns {Promise<object>}
   */
  async getCollectionsByParent(parentId) {
    return await this.apiService.get(`/collections-by-parent/${parentId}`);
  }

  /**
   * Share a collection with another user
   * @param {string} collectionId
   * @param {object} shareData
   * @returns {Promise<object>}
   */
  async shareCollection(collectionId, shareData) {
    const payload = {
      collection_id: collectionId,
      recipient_id: shareData.recipient_id,
      recipient_email: shareData.recipient_email,
      permission_level: shareData.permission_level,
      encrypted_collection_key: shareData.encrypted_collection_key,
      share_with_descendants: shareData.share_with_descendants ?? true,
    };

    return await this.apiService.post(
      `/collections/${collectionId}/share`,
      payload,
    );
  }

  /**
   * Remove a member from a collection
   * @param {string} collectionId
   * @param {object} memberData
   * @returns {Promise<object>}
   */
  async removeMember(collectionId, memberData) {
    const payload = {
      collection_id: collectionId,
      recipient_id: memberData.recipient_id,
      remove_from_descendants: memberData.remove_from_descendants ?? true,
    };

    return await this.apiService.delete(
      `/collections/${collectionId}/members`,
      payload,
    );
  }

  /**
   * Move a collection to a new parent
   * @param {string} collectionId
   * @param {object} moveData
   * @returns {Promise<object>}
   */
  async moveCollection(collectionId, moveData) {
    const payload = {
      collection_id: collectionId,
      new_parent_id: moveData.new_parent_id,
      updated_ancestors: moveData.updated_ancestors,
      updated_path_segments: moveData.updated_path_segments,
    };

    return await this.apiService.post(
      `/collections/${collectionId}/move`,
      payload,
    );
  }

  /**
   * Archive a collection
   * @param {string} collectionId
   * @returns {Promise<object>}
   */
  async archiveCollection(collectionId) {
    return await this.apiService.post(`/collections/${collectionId}/archive`);
  }

  /**
   * Restore an archived collection
   * @param {string} collectionId
   * @returns {Promise<object>}
   */
  async restoreCollection(collectionId) {
    return await this.apiService.post(`/collections/${collectionId}/restore`);
  }

  /**
   * Sync collections for offline clients
   * @param {object} syncParams
   * @returns {Promise<object>}
   */
  async syncCollections(syncParams = {}) {
    const params = {
      limit: syncParams.limit || 1000,
      cursor: syncParams.cursor || null,
    };

    return await this.apiService.get("/sync/collections", params);
  }
}

export default CollectionService;
