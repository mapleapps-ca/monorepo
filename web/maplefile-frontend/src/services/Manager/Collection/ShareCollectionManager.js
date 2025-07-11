// File: monorepo/web/maplefile-frontend/src/services/Manager/Collection/ShareCollectionManager.js
// Share Collection Manager - Orchestrates API, Storage, and Crypto services for collection sharing

import ShareCollectionAPIService from "../../API/Collection/ShareCollectionAPIService.js";
import ShareCollectionStorageService from "../../Storage/Collection/ShareCollectionStorageService.js";
import CollectionCryptoService from "../../Crypto/CollectionCryptoService.js";

class ShareCollectionManager {
  constructor(authManager) {
    // ShareCollectionManager depends on AuthManager and orchestrates API, Storage, and Crypto services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new ShareCollectionAPIService(authManager);
    this.storageService = new ShareCollectionStorageService();
    this.cryptoService = CollectionCryptoService; // Use singleton instance

    // Event listeners for collection sharing events
    this.collectionSharingListeners = new Set();

    console.log(
      "[ShareCollectionManager] Collection sharing manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log(
        "[ShareCollectionManager] Initializing collection sharing manager...",
      );

      // Initialize crypto service
      await this.cryptoService.initialize();

      console.log(
        "[ShareCollectionManager] Collection sharing manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[ShareCollectionManager] Failed to initialize collection sharing manager:",
        error,
      );
    }
  }

  // === Collection Sharing with Encryption ===

  // Share collection with another user with full E2EE encryption
  async shareCollection(collectionId, shareData, password = null) {
    try {
      this.isLoading = true;
      console.log("[ShareCollectionManager] Starting collection sharing");
      console.log("[ShareCollectionManager] Collection ID:", collectionId);
      console.log("[ShareCollectionManager] Share data:", {
        recipient_id: shareData.recipient_id,
        recipient_email: shareData.recipient_email,
        permission_level: shareData.permission_level,
        share_with_descendants: shareData.share_with_descendants,
      });

      // Validate input early
      if (!collectionId) {
        throw new Error("Collection ID is required");
      }

      if (!shareData.recipient_id) {
        throw new Error("Recipient user ID is required");
      }

      if (!shareData.recipient_email) {
        throw new Error("Recipient email is required");
      }

      if (!shareData.permission_level) {
        throw new Error("Permission level is required");
      }

      // Validate permission level
      const validPermissions = ["read_only", "read_write", "admin"];
      if (!validPermissions.includes(shareData.permission_level)) {
        throw new Error(
          `Invalid permission level: ${shareData.permission_level}`,
        );
      }

      // Get current collection to get collection key
      const currentCollection = await this.getCurrentCollection(collectionId);
      if (!currentCollection || !currentCollection.encrypted_collection_key) {
        throw new Error(
          "Cannot share collection: Current collection or encrypted collection key not found. The collection may not exist or you may not have access.",
        );
      }

      // Get recipient's public key (this would need to be fetched from user directory/API)
      const recipientPublicKey = await this.getRecipientPublicKey(
        shareData.recipient_id,
      );
      if (!recipientPublicKey) {
        throw new Error(
          "Cannot get recipient's public key. User may not exist or may not have set up encryption.",
        );
      }

      // Get collection key for encryption
      let collectionKey = this.storageService.getCollectionKey(collectionId);

      // If no cached key, decrypt it from current collection data
      if (!collectionKey) {
        console.log(
          "[ShareCollectionManager] Decrypting collection key for sharing",
        );

        const userKeys = await this.cryptoService.getUserKeys();
        collectionKey = await this.cryptoService.decryptCollectionKey(
          currentCollection.encrypted_collection_key,
          userKeys.masterKey,
        );

        // Cache it for future use
        this.storageService.storeCollectionKey(collectionId, collectionKey);
      }

      if (!collectionKey) {
        throw new Error(
          "Collection key not available for sharing. Password may be required.",
        );
      }

      // Encrypt collection key for recipient using their public key
      console.log(
        "[ShareCollectionManager] Encrypting collection key for recipient",
      );
      const encryptedCollectionKeyForRecipient =
        await this.cryptoService.encryptCollectionKeyForRecipient(
          collectionKey,
          recipientPublicKey,
        );

      console.log(
        "[ShareCollectionManager] Encrypted collection key type:",
        typeof encryptedCollectionKeyForRecipient,
      );
      console.log(
        "[ShareCollectionManager] Encrypted collection key length:",
        encryptedCollectionKeyForRecipient.length,
      );

      // Convert base64 string to byte array for API
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );
      const encryptedKeyBytes = CryptoService.tryDecodeBase64(
        encryptedCollectionKeyForRecipient,
      );
      const encryptedKeyArray = Array.from(encryptedKeyBytes);

      console.log(
        "[ShareCollectionManager] Encrypted key array length:",
        encryptedKeyArray.length,
      );

      // Prepare API share data with correct format
      const apiShareData = {
        collection_id: collectionId,
        recipient_id: shareData.recipient_id,
        recipient_email: shareData.recipient_email,
        permission_level: shareData.permission_level,
        encrypted_collection_key: encryptedKeyArray, // Send as byte array, not base64 string!
        share_with_descendants: shareData.share_with_descendants !== false, // Default to true
      };

      console.log("[ShareCollectionManager] Final API payload:", {
        collection_id: apiShareData.collection_id,
        recipient_id: apiShareData.recipient_id,
        recipient_email: apiShareData.recipient_email,
        permission_level: apiShareData.permission_level,
        encrypted_collection_key_length:
          apiShareData.encrypted_collection_key.length,
        encrypted_collection_key_type:
          typeof apiShareData.encrypted_collection_key,
        share_with_descendants: apiShareData.share_with_descendants,
      });

      console.log("[ShareCollectionManager] Sending share request to API");

      // Share collection via API
      const shareResponse = await this.apiService.shareCollection(
        collectionId,
        apiShareData,
      );

      console.log(
        "[ShareCollectionManager] Collection shared via API successfully",
        shareResponse,
      );

      // Store sharing information locally
      const shareInfo = {
        recipient_id: shareData.recipient_id,
        recipient_email: shareData.recipient_email,
        permission_level: shareData.permission_level,
        share_with_descendants: shareData.share_with_descendants,
        memberships_created: shareResponse.memberships_created,
        encrypted_key_for_recipient: encryptedCollectionKeyForRecipient,
      };

      this.storageService.storeSharedCollection(collectionId, shareInfo);

      // Notify listeners
      this.notifyCollectionSharingListeners("collection_shared", {
        collectionId,
        recipientId: shareData.recipient_id,
        recipientEmail: shareData.recipient_email,
        permissionLevel: shareData.permission_level,
        shareWithDescendants: shareData.share_with_descendants,
        membershipsCreated: shareResponse.memberships_created,
      });

      console.log(
        "[ShareCollectionManager] Collection shared successfully:",
        collectionId,
        "->",
        shareData.recipient_email,
      );

      return {
        collection_id: collectionId,
        recipient_id: shareData.recipient_id,
        recipient_email: shareData.recipient_email,
        permission_level: shareData.permission_level,
        share_with_descendants: shareData.share_with_descendants,
        memberships_created: shareResponse.memberships_created,
        success: true,
        message: shareResponse.message,
      };
    } catch (error) {
      console.error(
        "[ShareCollectionManager] Collection sharing failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionSharingListeners("collection_sharing_failed", {
        collectionId,
        recipientId: shareData.recipient_id,
        error: error.message,
        shareData,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Remove member from collection
  async removeMember(collectionId, recipientId, removeFromDescendants = true) {
    try {
      this.isLoading = true;
      console.log("[ShareCollectionManager] Starting member removal");
      console.log("[ShareCollectionManager] Collection ID:", collectionId);
      console.log("[ShareCollectionManager] Recipient ID:", recipientId);

      // Validate input
      if (!collectionId) {
        throw new Error("Collection ID is required");
      }

      if (!recipientId) {
        throw new Error("Recipient ID is required");
      }

      // Prepare API remove data
      const apiRemoveData = {
        collection_id: collectionId,
        recipient_id: recipientId,
        remove_from_descendants: removeFromDescendants,
      };

      console.log(
        "[ShareCollectionManager] Sending remove member request to API",
      );

      // Remove member via API
      const removeResponse = await this.apiService.removeMember(
        collectionId,
        apiRemoveData,
      );

      console.log(
        "[ShareCollectionManager] Member removed via API successfully",
        removeResponse,
      );

      // Remove from local storage
      this.storageService.removeSharedCollection(collectionId, recipientId);

      // Notify listeners
      this.notifyCollectionSharingListeners("member_removed", {
        collectionId,
        recipientId,
        removeFromDescendants,
      });

      console.log(
        "[ShareCollectionManager] Member removed successfully:",
        collectionId,
        "->",
        recipientId,
      );

      return {
        collection_id: collectionId,
        recipient_id: recipientId,
        remove_from_descendants: removeFromDescendants,
        success: true,
        message: removeResponse.message,
      };
    } catch (error) {
      console.error("[ShareCollectionManager] Member removal failed:", error);

      // Notify listeners of failure
      this.notifyCollectionSharingListeners("member_removal_failed", {
        collectionId,
        recipientId,
        error: error.message,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // Get collection members
  async getCollectionMembers(collectionId, forceRefresh = false) {
    try {
      console.log(
        "[ShareCollectionManager] Getting collection members:",
        collectionId,
      );

      // Check cache first unless force refresh
      if (!forceRefresh) {
        const cachedMembers =
          this.storageService.getCollectionMembers(collectionId);
        if (cachedMembers.length > 0) {
          console.log("[ShareCollectionManager] Using cached members");
          return cachedMembers;
        }
      }

      // Get from API
      const members = await this.apiService.getCollectionMembers(collectionId);

      // Store in cache
      this.storageService.storeCollectionMembers(collectionId, members);

      console.log(
        "[ShareCollectionManager] Collection members retrieved:",
        members.length,
      );
      return members;
    } catch (error) {
      console.error(
        "[ShareCollectionManager] Failed to get collection members:",
        error,
      );

      // Return cached data on API failure
      const cachedMembers =
        this.storageService.getCollectionMembers(collectionId);
      if (cachedMembers.length > 0) {
        console.log(
          "[ShareCollectionManager] Returning cached members due to API failure",
        );
        return cachedMembers;
      }

      throw error;
    }
  }

  // === Helper Methods ===

  // Get current collection data
  async getCurrentCollection(collectionId) {
    try {
      console.log(
        "[ShareCollectionManager] Attempting to get current collection:",
        collectionId,
      );

      // Try to get from GetCollectionManager if available
      const getCollectionManager = await this.getCollectionManager();
      if (getCollectionManager) {
        try {
          console.log("[ShareCollectionManager] Using GetCollectionManager");
          const result = await getCollectionManager.getCollection(collectionId);
          if (result && result.collection) {
            console.log(
              "[ShareCollectionManager] Collection retrieved via GetCollectionManager",
            );
            return result.collection;
          }
        } catch (getError) {
          console.warn(
            "[ShareCollectionManager] GetCollectionManager failed, trying direct API:",
            getError.message,
          );
        }
      }

      // Fallback: direct API call
      console.log("[ShareCollectionManager] Using direct API call");
      const { default: GetCollectionAPIService } = await import(
        "../../API/Collection/GetCollectionAPIService.js"
      );
      const apiService = new GetCollectionAPIService(this.authManager);
      const collection = await apiService.getCollection(collectionId);

      console.log(
        "[ShareCollectionManager] Collection retrieved via direct API",
      );
      return collection;
    } catch (error) {
      console.error(
        "[ShareCollectionManager] Failed to get current collection:",
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
        "[ShareCollectionManager] GetCollectionManager not available:",
        error,
      );
      return null;
    }
  }

  // Get recipient's public key (placeholder - would need to be implemented based on user directory API)
  async getRecipientPublicKey(recipientId) {
    try {
      console.log(
        "[ShareCollectionManager] Getting recipient public key:",
        recipientId,
      );

      // TODO: Implement actual user directory lookup
      // For now, this is a placeholder that would need to call:
      // - User directory API to find user by ID
      // - Get user's public key from their profile
      // - Return the public key as Uint8Array

      // Placeholder implementation - in real app this would be an API call
      console.warn(
        "[ShareCollectionManager] PLACEHOLDER: getRecipientPublicKey needs implementation",
      );

      // For testing, generate a fake public key
      // In real implementation, this would be: await userDirectoryAPI.getUserPublicKey(recipientId)
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );
      await CryptoService.initialize();

      // Generate a fake public key for testing (32 bytes)
      const fakePublicKey = CryptoService.sodium.randombytes_buf(32);

      console.warn(
        "[ShareCollectionManager] USING FAKE PUBLIC KEY FOR TESTING - recipient:",
        recipientId,
      );

      return fakePublicKey;
    } catch (error) {
      console.error(
        "[ShareCollectionManager] Failed to get recipient public key:",
        error,
      );
      throw new Error(`Failed to get recipient public key: ${error.message}`);
    }
  }

  // === Collection Key Management ===

  // Store collection key in memory cache
  storeCollectionKey(collectionId, collectionKey) {
    this.storageService.storeCollectionKey(collectionId, collectionKey);
  }

  // Get collection key from memory cache
  getCollectionKey(collectionId) {
    return this.storageService.getCollectionKey(collectionId);
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
        "[ShareCollectionManager] Failed to get user password:",
        error,
      );
      return null;
    }
  }

  // === Collection Sharing Management ===

  // Get all shared collections
  getSharedCollections() {
    return this.storageService.getSharedCollections();
  }

  // Get shared collections by collection ID
  getSharedCollectionsByCollectionId(collectionId) {
    return this.storageService.getSharedCollectionsByCollectionId(collectionId);
  }

  // Get shared collections by recipient
  getSharedCollectionsByRecipient(recipientId) {
    return this.storageService.getSharedCollectionsByRecipient(recipientId);
  }

  // Get sharing history for a collection
  getSharingHistory(collectionId = null) {
    return this.storageService.getSharingHistory(collectionId);
  }

  // Search shared collections
  searchSharedCollections(searchTerm) {
    const shares = this.getSharedCollections();
    return this.storageService.searchSharedCollections(searchTerm, shares);
  }

  // Remove all shares for a collection
  async removeAllSharesForCollection(collectionId) {
    try {
      console.log(
        "[ShareCollectionManager] Removing all shares for collection:",
        collectionId,
      );

      const removedCount =
        this.storageService.removeAllSharedForCollection(collectionId);

      this.notifyCollectionSharingListeners(
        "all_shares_removed_for_collection",
        {
          collectionId,
          removedCount,
        },
      );

      return removedCount;
    } catch (error) {
      console.error(
        "[ShareCollectionManager] Failed to remove all shares for collection:",
        error,
      );
      throw error;
    }
  }

  // Clear all shared collections
  async clearAllSharedCollections() {
    try {
      console.log("[ShareCollectionManager] Clearing all shared collections");

      this.storageService.clearAllSharedCollections();

      this.notifyCollectionSharingListeners(
        "all_shared_collections_cleared",
        {},
      );

      console.log("[ShareCollectionManager] All shared collections cleared");
    } catch (error) {
      console.error(
        "[ShareCollectionManager] Failed to clear shared collections:",
        error,
      );
      throw error;
    }
  }

  // === Event Management ===

  // Add collection sharing listener
  addCollectionSharingListener(callback) {
    if (typeof callback === "function") {
      this.collectionSharingListeners.add(callback);
      console.log(
        "[ShareCollectionManager] Collection sharing listener added. Total listeners:",
        this.collectionSharingListeners.size,
      );
    }
  }

  // Remove collection sharing listener
  removeCollectionSharingListener(callback) {
    this.collectionSharingListeners.delete(callback);
    console.log(
      "[ShareCollectionManager] Collection sharing listener removed. Total listeners:",
      this.collectionSharingListeners.size,
    );
  }

  // Notify collection sharing listeners
  notifyCollectionSharingListeners(eventType, eventData) {
    console.log(
      `[ShareCollectionManager] Notifying ${this.collectionSharingListeners.size} listeners of ${eventType}`,
    );

    this.collectionSharingListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[ShareCollectionManager] Error in collection sharing listener:",
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
      canShareCollections: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      crypto: cryptoStatus,
      listenerCount: this.collectionSharingListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "ShareCollectionManager",
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

export default ShareCollectionManager;
