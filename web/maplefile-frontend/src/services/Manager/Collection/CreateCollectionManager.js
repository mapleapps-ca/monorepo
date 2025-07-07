// File: monorepo/web/maplefile-frontend/src/services/Manager/Colleciton/ CreateCollectionManager.js
// Create Collection Manager - Orchestrates API and Storage services for collection creation

import CreateCollectionAPIService from "../../API/Collection/CreateCollectionAPIService.js";
import CreateCollectionStorageService from "../../Storage/Collection/CreateCollectionStorageService.js";

class CreateCollectionManager {
  constructor(authManager) {
    // CreateCollectionManager depends on AuthManager and orchestrates API and Storage services
    this.authManager = authManager;
    this.isLoading = false;

    // Initialize dependent services
    this.apiService = new CreateCollectionAPIService(authManager);
    this.storageService = new CreateCollectionStorageService();

    // Event listeners for collection creation events
    this.collectionCreationListeners = new Set();

    console.log(
      "[CreateCollectionManager] Collection manager initialized with AuthManager dependency",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log(
        "[CreateCollectionManager] Initializing collection manager...",
      );

      // Initialize crypto service
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );
      await CryptoService.initialize();

      console.log(
        "[CreateCollectionManager] Collection manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to initialize collection manager:",
        error,
      );
    }
  }

  // === Collection Creation with Encryption ===

  // Create collection with full E2EE encryption
  async createCollection(collectionData, password = null) {
    try {
      this.isLoading = true;
      console.log("[CreateCollectionManager] Starting collection creation");
      console.log("[CreateCollectionManager] Collection data:", {
        name: collectionData.name,
        type: collectionData.collection_type || "folder",
        hasParent: !!collectionData.parent_id,
      });

      // Validate input
      if (!collectionData.name) {
        throw new Error("Collection name is required");
      }

      // Get password for encryption
      const userPassword = password || (await this.getUserPassword());
      if (!userPassword) {
        throw new Error("Password required for collection encryption");
      }

      // Generate collection ID
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );
      const collectionId = CryptoService.generateUUID();

      console.log(
        "[CreateCollectionManager] Generated collection ID:",
        collectionId,
      );

      // Step 1: Generate collection key (32 bytes)
      const collectionKey = CryptoService.generateRandomKey();
      console.log("[CreateCollectionManager] Generated collection key");

      // Step 2: Encrypt collection name with collection key
      const encryptedName = await this.encryptCollectionName(
        collectionData.name,
        collectionKey,
      );
      console.log("[CreateCollectionManager] Encrypted collection name");

      // Step 3: Get user's master key and encrypt collection key
      const encryptedCollectionKey =
        await this.encryptCollectionKeyWithPassword(
          collectionKey,
          userPassword,
        );
      console.log("[CreateCollectionManager] Encrypted collection key");

      // Step 4: Prepare API data
      const apiData = {
        id: collectionId,
        encrypted_name: encryptedName,
        collection_type: collectionData.collection_type || "folder",
        encrypted_collection_key: encryptedCollectionKey,
      };

      // Add optional fields
      if (collectionData.parent_id) {
        apiData.parent_id = collectionData.parent_id;
      }
      if (
        collectionData.ancestor_ids &&
        collectionData.ancestor_ids.length > 0
      ) {
        apiData.ancestor_ids = collectionData.ancestor_ids;
      }

      // Step 5: Create collection via API
      console.log("[CreateCollectionManager] Sending to API");
      const createdCollection = await this.apiService.createCollection(apiData);

      // Step 6: Store collection key in memory for future use
      this.storageService.storeCollectionKey(collectionId, collectionKey);

      // Step 7: Prepare decrypted collection for local storage
      const decryptedCollection = {
        ...createdCollection,
        name: collectionData.name, // Store decrypted name
        _originalEncryptedName: encryptedName,
        _hasCollectionKey: true,
      };

      // Step 8: Store in local storage
      this.storageService.storeCreatedCollection(decryptedCollection);

      // Step 9: Notify listeners
      this.notifyCollectionCreationListeners("collection_created", {
        collectionId,
        name: collectionData.name,
        type: collectionData.collection_type || "folder",
      });

      console.log(
        "[CreateCollectionManager] Collection created successfully:",
        collectionId,
      );

      return {
        collection: decryptedCollection,
        collectionId,
        success: true,
      };
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Collection creation failed:",
        error,
      );

      // Notify listeners of failure
      this.notifyCollectionCreationListeners("collection_creation_failed", {
        error: error.message,
        collectionData,
      });

      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  // === Encryption Helper Methods ===

  // Encrypt collection name with collection key
  async encryptCollectionName(name, collectionKey) {
    try {
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );

      // Use the crypto service to encrypt the name
      const encryptedName = await CryptoService.encryptWithKey(
        name,
        collectionKey,
      );

      console.log(
        "[CreateCollectionManager] Collection name encrypted successfully",
      );
      return encryptedName;
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to encrypt collection name:",
        error,
      );
      throw new Error(`Name encryption failed: ${error.message}`);
    }
  }

  // Encrypt collection key with user's master key (derived from password)
  async encryptCollectionKeyWithPassword(collectionKey, password) {
    try {
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );
      const { default: LocalStorageService } = await import(
        "../../Storage/LocalStorageService.js"
      );

      // Get user's encrypted data
      const userEncryptedData = LocalStorageService.getUserEncryptedData();
      if (!userEncryptedData.salt || !userEncryptedData.encryptedMasterKey) {
        throw new Error("Missing user encrypted data. Please log in again.");
      }

      console.log("[CreateCollectionManager] Decrypting user's master key...");

      // Decode encrypted data
      const salt = CryptoService.tryDecodeBase64(userEncryptedData.salt);
      const encryptedMasterKey = CryptoService.tryDecodeBase64(
        userEncryptedData.encryptedMasterKey,
      );

      // Derive key encryption key from password
      const keyEncryptionKey = await CryptoService.deriveKeyFromPassword(
        password,
        salt,
      );

      // Decrypt master key
      const masterKey = CryptoService.decryptWithSecretBox(
        encryptedMasterKey,
        keyEncryptionKey,
      );

      console.log(
        "[CreateCollectionManager] Master key decrypted, encrypting collection key...",
      );

      // Encrypt collection key with master key using the same pattern as file keys
      const encryptedKey = await CryptoService.encryptFileKey(
        collectionKey,
        masterKey,
      );

      console.log(
        "[CreateCollectionManager] Collection key encrypted successfully",
      );
      return encryptedKey;
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to encrypt collection key:",
        error,
      );
      throw new Error(`Collection key encryption failed: ${error.message}`);
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
        "[CreateCollectionManager] Failed to get user password:",
        error,
      );
      return null;
    }
  }

  // === Collection Decryption ===

  // Decrypt collection data (for displaying stored collections)
  async decryptCollection(encryptedCollection, password = null) {
    try {
      console.log(
        "[CreateCollectionManager] Decrypting collection:",
        encryptedCollection.id,
      );

      // Get collection key
      let collectionKey = this.storageService.getCollectionKey(
        encryptedCollection.id,
      );

      if (!collectionKey) {
        // Try to decrypt collection key if we have the password
        const userPassword = password || (await this.getUserPassword());
        if (!userPassword) {
          throw new Error("Password required to decrypt collection");
        }

        collectionKey = await this.decryptCollectionKey(
          encryptedCollection.encrypted_collection_key,
          userPassword,
        );

        // Cache the decrypted key
        this.storageService.storeCollectionKey(
          encryptedCollection.id,
          collectionKey,
        );
      }

      // Decrypt collection name
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );
      const decryptedNameBytes = await CryptoService.decryptWithKey(
        encryptedCollection.encrypted_name,
        collectionKey,
      );

      const name = new TextDecoder().decode(decryptedNameBytes);

      return {
        ...encryptedCollection,
        name,
        _isDecrypted: true,
        _hasCollectionKey: true,
      };
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to decrypt collection:",
        error,
      );

      // Return collection with error marker
      return {
        ...encryptedCollection,
        name: "[Decryption Failed]",
        _isDecrypted: false,
        _decryptionError: error.message,
      };
    }
  }

  // Decrypt collection key with password
  async decryptCollectionKey(encryptedCollectionKey, password) {
    try {
      const { default: CryptoService } = await import(
        "../../Crypto/CryptoService.js"
      );
      const { default: LocalStorageService } = await import(
        "../../Storage/LocalStorageService.js"
      );

      // Get user's encrypted data
      const userEncryptedData = LocalStorageService.getUserEncryptedData();
      if (!userEncryptedData.salt || !userEncryptedData.encryptedMasterKey) {
        throw new Error("Missing user encrypted data");
      }

      // Decrypt user's master key
      const salt = CryptoService.tryDecodeBase64(userEncryptedData.salt);
      const encryptedMasterKey = CryptoService.tryDecodeBase64(
        userEncryptedData.encryptedMasterKey,
      );
      const keyEncryptionKey = await CryptoService.deriveKeyFromPassword(
        password,
        salt,
      );
      const masterKey = CryptoService.decryptWithSecretBox(
        encryptedMasterKey,
        keyEncryptionKey,
      );

      // Decrypt collection key with master key
      const collectionKey = await CryptoService.decryptFileKey(
        encryptedCollectionKey,
        masterKey,
      );

      return collectionKey;
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to decrypt collection key:",
        error,
      );
      throw error;
    }
  }

  // === Collection Management ===

  // Get all created collections
  getCreatedCollections() {
    return this.storageService.getCreatedCollections();
  }

  // Get collection by ID
  getCollectionById(collectionId) {
    return this.storageService.getCollectionById(collectionId);
  }

  // Search collections
  searchCollections(searchTerm) {
    const collections = this.getCreatedCollections();
    return this.storageService.searchCollections(searchTerm, collections);
  }

  // Remove collection
  async removeCollection(collectionId) {
    try {
      console.log(
        "[CreateCollectionManager] Removing collection:",
        collectionId,
      );

      const removed = this.storageService.removeCollection(collectionId);

      if (removed) {
        this.notifyCollectionCreationListeners("collection_removed", {
          collectionId,
        });
      }

      return removed;
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to remove collection:",
        error,
      );
      throw error;
    }
  }

  // Clear all collections
  async clearAllCollections() {
    try {
      console.log("[CreateCollectionManager] Clearing all collections");

      this.storageService.clearAllCollections();

      this.notifyCollectionCreationListeners("all_collections_cleared", {});

      console.log("[CreateCollectionManager] All collections cleared");
    } catch (error) {
      console.error(
        "[CreateCollectionManager] Failed to clear collections:",
        error,
      );
      throw error;
    }
  }

  // === Event Management ===

  // Add collection creation listener
  addCollectionCreationListener(callback) {
    if (typeof callback === "function") {
      this.collectionCreationListeners.add(callback);
      console.log(
        "[CreateCollectionManager] Collection creation listener added. Total listeners:",
        this.collectionCreationListeners.size,
      );
    }
  }

  // Remove collection creation listener
  removeCollectionCreationListener(callback) {
    this.collectionCreationListeners.delete(callback);
    console.log(
      "[CreateCollectionManager] Collection creation listener removed. Total listeners:",
      this.collectionCreationListeners.size,
    );
  }

  // Notify collection creation listeners
  notifyCollectionCreationListeners(eventType, eventData) {
    console.log(
      `[CreateCollectionManager] Notifying ${this.collectionCreationListeners.size} listeners of ${eventType}`,
    );

    this.collectionCreationListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error(
          "[CreateCollectionManager] Error in collection creation listener:",
          error,
        );
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
      canCreateCollections: this.authManager.canMakeAuthenticatedRequests(),
      storage: storageInfo,
      listenerCount: this.collectionCreationListeners.size,
      hasPasswordService: !!this.getUserPassword,
    };
  }

  // === Debug Information ===

  // Get comprehensive debug information
  getDebugInfo() {
    return {
      serviceName: "CreateCollectionManager",
      role: "orchestrator",
      isAuthenticated: this.authManager.isAuthenticated(),
      apiService: this.apiService.getDebugInfo(),
      storageService: this.storageService.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
      authManagerStatus: {
        userEmail: this.authManager.getCurrentUserEmail(),
        canMakeRequests: this.authManager.canMakeAuthenticatedRequests(),
        sessionKeyStatus: this.authManager.getSessionKeyStatus(),
      },
    };
  }
}

export default CreateCollectionManager;
