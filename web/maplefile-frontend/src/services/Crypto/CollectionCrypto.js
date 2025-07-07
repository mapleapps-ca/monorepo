// File: monorepo/web/maplefile-frontend/src/services/Crypto/CollectionCrypto.js
// Collection-specific encryption operations following E2EE architecture

class CollectionCrypto {
  constructor() {
    this.isInitialized = false;
    console.log("[CollectionCrypto] Collection crypto service initialized");
  }

  // Initialize the crypto service
  async initialize() {
    if (this.isInitialized) return;

    try {
      // Initialize the main crypto service
      const { default: CryptoService } = await import("./CryptoService.js");
      await CryptoService.initialize();

      this.cryptoService = CryptoService;
      this.isInitialized = true;

      console.log(
        "[CollectionCrypto] Collection crypto service initialized successfully",
      );
    } catch (error) {
      console.error("[CollectionCrypto] Failed to initialize:", error);
      throw new Error(
        `Failed to initialize CollectionCrypto: ${error.message}`,
      );
    }
  }

  // === Collection Key Generation ===

  // Generate a new 32-byte collection key
  generateCollectionKey() {
    if (!this.isInitialized) {
      throw new Error("CollectionCrypto not initialized");
    }

    return this.cryptoService.generateRandomKey();
  }

  // === Collection Name Encryption ===

  // Encrypt collection name with collection key
  async encryptCollectionName(name, collectionKey) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCrypto not initialized");
      }

      if (!name || !collectionKey) {
        throw new Error("Name and collection key are required");
      }

      console.log("[CollectionCrypto] Encrypting collection name");

      // Use the crypto service to encrypt the name with the collection key
      const encryptedName = await this.cryptoService.encryptWithKey(
        name,
        collectionKey,
      );

      console.log("[CollectionCrypto] Collection name encrypted successfully");
      return encryptedName;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to encrypt collection name:",
        error,
      );
      throw new Error(`Name encryption failed: ${error.message}`);
    }
  }

  // Decrypt collection name with collection key
  async decryptCollectionName(encryptedName, collectionKey) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCrypto not initialized");
      }

      if (!encryptedName || !collectionKey) {
        console.warn(
          "[CollectionCrypto] Missing encrypted name or collection key",
        );
        return "[Unable to decrypt]";
      }

      console.log("[CollectionCrypto] Decrypting collection name");

      // Decrypt the name with the collection key
      const decryptedNameBytes = await this.cryptoService.decryptWithKey(
        encryptedName,
        collectionKey,
      );

      const name = new TextDecoder().decode(decryptedNameBytes);

      console.log("[CollectionCrypto] Collection name decrypted successfully");
      return name;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to decrypt collection name:",
        error,
      );
      return "[Decryption Failed]";
    }
  }

  // === Collection Key Encryption ===

  // Encrypt collection key with user's master key (derived from password)
  async encryptCollectionKeyWithPassword(collectionKey, password) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCrypto not initialized");
      }

      if (!collectionKey || !password) {
        throw new Error("Collection key and password are required");
      }

      console.log(
        "[CollectionCrypto] Encrypting collection key with user's master key",
      );

      // Get user's master key
      const masterKey = await this.getUserMasterKeyFromPassword(password);

      // Encrypt collection key with master key using the same pattern as file keys
      const encryptedKey = await this.cryptoService.encryptFileKey(
        collectionKey,
        masterKey,
      );

      console.log("[CollectionCrypto] Collection key encrypted successfully");
      return encryptedKey;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to encrypt collection key:",
        error,
      );
      throw new Error(`Collection key encryption failed: ${error.message}`);
    }
  }

  // Decrypt collection key with user's master key (derived from password)
  async decryptCollectionKeyWithPassword(encryptedCollectionKey, password) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCrypto not initialized");
      }

      if (!encryptedCollectionKey || !password) {
        throw new Error("Encrypted collection key and password are required");
      }

      console.log(
        "[CollectionCrypto] Decrypting collection key with user's master key",
      );

      // Get user's master key
      const masterKey = await this.getUserMasterKeyFromPassword(password);

      // Decrypt collection key with master key
      const collectionKey = await this.cryptoService.decryptFileKey(
        encryptedCollectionKey,
        masterKey,
      );

      console.log("[CollectionCrypto] Collection key decrypted successfully");
      return collectionKey;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to decrypt collection key:",
        error,
      );
      throw error;
    }
  }

  // === User Master Key Derivation ===

  // Get user's master key from password (private helper method)
  async getUserMasterKeyFromPassword(password) {
    try {
      const { default: LocalStorageService } = await import(
        "../Storage/LocalStorageService.js"
      );

      // Get user's encrypted data
      const userEncryptedData = LocalStorageService.getUserEncryptedData();
      if (!userEncryptedData.salt || !userEncryptedData.encryptedMasterKey) {
        throw new Error("Missing user encrypted data. Please log in again.");
      }

      console.log(
        "[CollectionCrypto] Deriving user's master key from password",
      );

      // Decode encrypted data
      const salt = this.cryptoService.tryDecodeBase64(userEncryptedData.salt);
      const encryptedMasterKey = this.cryptoService.tryDecodeBase64(
        userEncryptedData.encryptedMasterKey,
      );

      // Derive key encryption key from password
      const keyEncryptionKey = await this.cryptoService.deriveKeyFromPassword(
        password,
        salt,
      );

      // Decrypt master key
      const masterKey = this.cryptoService.decryptWithSecretBox(
        encryptedMasterKey,
        keyEncryptionKey,
      );

      console.log("[CollectionCrypto] User's master key derived successfully");
      return masterKey;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to derive user's master key:",
        error,
      );
      throw new Error(`Master key derivation failed: ${error.message}`);
    }
  }

  // === Collection Data Encryption/Decryption ===

  // Encrypt complete collection data for API
  async encryptCollectionForAPI(collectionData, password) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCrypto not initialized");
      }

      console.log("[CollectionCrypto] Encrypting collection data for API");

      // Validate input
      if (!collectionData.name) {
        throw new Error("Collection name is required");
      }

      // Generate collection ID and key
      const collectionId = this.cryptoService.generateUUID();
      const collectionKey = this.generateCollectionKey();

      // Encrypt collection name with collection key
      const encryptedName = await this.encryptCollectionName(
        collectionData.name,
        collectionKey,
      );

      // Encrypt collection key with user's master key
      const encryptedCollectionKey =
        await this.encryptCollectionKeyWithPassword(collectionKey, password);

      // Prepare API data
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

      console.log(
        "[CollectionCrypto] Collection data encrypted for API successfully",
      );

      return {
        apiData,
        collectionKey,
        collectionId,
      };
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to encrypt collection for API:",
        error,
      );
      throw error;
    }
  }

  // Decrypt collection data from API response
  async decryptCollectionFromAPI(
    encryptedCollection,
    collectionKey = null,
    password = null,
  ) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCrypto not initialized");
      }

      console.log("[CollectionCrypto] Decrypting collection data from API");

      let workingCollectionKey = collectionKey;

      // If no collection key provided, try to decrypt it
      if (!workingCollectionKey && password) {
        workingCollectionKey = await this.decryptCollectionKeyWithPassword(
          encryptedCollection.encrypted_collection_key,
          password,
        );
      }

      if (!workingCollectionKey) {
        throw new Error("Collection key not available for decryption");
      }

      // Decrypt collection name
      const name = await this.decryptCollectionName(
        encryptedCollection.encrypted_name,
        workingCollectionKey,
      );

      // Return decrypted collection
      const decryptedCollection = {
        ...encryptedCollection,
        name,
        _isDecrypted: true,
        _hasCollectionKey: true,
        _originalEncryptedName: encryptedCollection.encrypted_name,
      };

      console.log(
        "[CollectionCrypto] Collection data decrypted from API successfully",
      );
      return decryptedCollection;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to decrypt collection from API:",
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

  // === Collection Key Sharing (for future use) ===

  // Encrypt collection key for sharing with another user (uses their public key)
  async encryptCollectionKeyForRecipient(collectionKey, recipientPublicKey) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCrypto not initialized");
      }

      if (!collectionKey || !recipientPublicKey) {
        throw new Error("Collection key and recipient public key are required");
      }

      console.log("[CollectionCrypto] Encrypting collection key for recipient");

      // Use sealed box (anonymous encryption) for sharing
      const encrypted = this.cryptoService.sodium.crypto_box_seal(
        collectionKey,
        recipientPublicKey,
      );

      // Return base64 string as API expects
      const encryptedBase64 = this.cryptoService.uint8ArrayToBase64(encrypted);

      console.log(
        "[CollectionCrypto] Collection key encrypted for recipient successfully",
      );
      return encryptedBase64;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to encrypt collection key for recipient:",
        error,
      );
      throw error;
    }
  }

  // Decrypt collection key shared with us (uses our private key)
  async decryptSharedCollectionKey(
    encryptedKey,
    userPrivateKey,
    userPublicKey,
  ) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCrypto not initialized");
      }

      if (!encryptedKey || !userPrivateKey || !userPublicKey) {
        throw new Error("Encrypted key and user keypair are required");
      }

      console.log("[CollectionCrypto] Decrypting shared collection key");

      // Decode from base64
      const encryptedData =
        typeof encryptedKey === "string"
          ? this.cryptoService.tryDecodeBase64(encryptedKey)
          : new Uint8Array(encryptedKey);

      // Decrypt with our private key
      const decrypted = this.cryptoService.sodium.crypto_box_seal_open(
        encryptedData,
        userPublicKey,
        userPrivateKey,
      );

      console.log(
        "[CollectionCrypto] Shared collection key decrypted successfully",
      );
      return decrypted;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to decrypt shared collection key:",
        error,
      );
      throw error;
    }
  }

  // === Utility Methods ===

  // Check if service is initialized
  isReady() {
    return this.isInitialized;
  }

  // Get service status
  getStatus() {
    return {
      isInitialized: this.isInitialized,
      hasCryptoService: !!this.cryptoService,
      cryptoServiceReady: this.cryptoService?.isInitialized || false,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "CollectionCrypto",
      status: this.getStatus(),
      capabilities: [
        "generateCollectionKey",
        "encryptCollectionName",
        "decryptCollectionName",
        "encryptCollectionKeyWithPassword",
        "decryptCollectionKeyWithPassword",
        "encryptCollectionForAPI",
        "decryptCollectionFromAPI",
        "encryptCollectionKeyForRecipient",
        "decryptSharedCollectionKey",
      ],
    };
  }
}

// Export singleton instance
export default new CollectionCrypto();
