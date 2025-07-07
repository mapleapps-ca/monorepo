// File: monorepo/web/maplefile-frontend/src/services/Crypto/CollectionCryptoService.js
// Collection-specific encryption operations following E2EE architecture - FIXED with PasswordStorageService

class CollectionCryptoService {
  constructor() {
    this.isInitialized = false;
    this.sodium = null;

    // In-memory cache for collection keys (NEVER stored in localStorage)
    this._collectionKeyCache = new Map();

    console.log(
      "[CollectionCryptoService] Collection crypto service initialized",
    );
  }

  // Initialize the crypto service
  async initialize() {
    if (this.isInitialized) return;

    try {
      // Initialize the main crypto service
      const { default: CryptoService } = await import("./CryptoService.js");
      await CryptoService.initialize();

      // Initialize libsodium directly
      const sodium = await import("libsodium-wrappers-sumo");
      await sodium.ready;
      this.sodium = sodium.default;

      this.cryptoService = CryptoService;
      this.isInitialized = true;

      console.log(
        "[CollectionCryptoService] Collection crypto service initialized successfully",
      );
    } catch (error) {
      console.error("[CollectionCryptoService] Failed to initialize:", error);
      throw new Error(
        `Failed to initialize CollectionCryptoService: ${error.message}`,
      );
    }
  }

  // === Password Management via PasswordStorageService ===

  // Get user password from PasswordStorageService
  async getUserPassword() {
    try {
      const { default: passwordStorageService } = await import(
        "../PasswordStorageService.js"
      );
      const password = passwordStorageService.getPassword();

      if (!password) {
        throw new Error(
          "No password available in PasswordStorageService. Please log in again.",
        );
      }

      console.log(
        "[CollectionCryptoService] Retrieved password from PasswordStorageService",
      );
      return password;
    } catch (error) {
      console.error(
        "[CollectionCryptoService] Failed to get password from PasswordStorageService:",
        error,
      );
      throw error;
    }
  }

  // === Collection Key Generation ===

  // Generate a new 32-byte collection key
  generateCollectionKey() {
    if (!this.isInitialized || !this.sodium) {
      throw new Error("CollectionCryptoService not initialized");
    }

    return this.sodium.randombytes_buf(32);
  }

  // === Collection Name Encryption/Decryption ===

  // Encrypt collection name with collection key
  encryptCollectionName(name, collectionKey) {
    if (!name || !collectionKey) {
      throw new Error("Name and collection key are required");
    }

    if (!this.sodium) {
      throw new Error("CollectionCryptoService not initialized");
    }

    const encoder = new TextEncoder();
    const nameBytes = encoder.encode(name);

    // Generate nonce
    const nonce = this.sodium.randombytes_buf(
      this.sodium.crypto_secretbox_NONCEBYTES,
    );

    // Encrypt name
    const encrypted = this.sodium.crypto_secretbox_easy(
      nameBytes,
      nonce,
      collectionKey,
    );

    // Combine nonce + encrypted data
    const combined = new Uint8Array(nonce.length + encrypted.length);
    combined.set(nonce, 0);
    combined.set(encrypted, nonce.length);

    // Return base64 encoded
    return this.cryptoService.uint8ArrayToBase64(combined);
  }

  // Decrypt collection name with collection key
  decryptCollectionName(encryptedName, collectionKey) {
    if (!encryptedName || !collectionKey) {
      console.warn(
        "[CollectionCryptoService] Missing encrypted name or collection key",
      );
      return "[Unable to decrypt]";
    }

    if (!this.sodium) {
      console.error("[CollectionCryptoService] Service not initialized");
      return "[Not initialized]";
    }

    try {
      // Decode from base64
      const combined = this.cryptoService.tryDecodeBase64(encryptedName);

      // Extract nonce and ciphertext
      const nonceLength = this.sodium.crypto_secretbox_NONCEBYTES;
      const nonce = combined.slice(0, nonceLength);
      const ciphertext = combined.slice(nonceLength);

      // Decrypt
      const decrypted = this.sodium.crypto_secretbox_open_easy(
        ciphertext,
        nonce,
        collectionKey,
      );

      // Convert to string
      const decoder = new TextDecoder();
      const name = decoder.decode(decrypted);

      console.log(
        "[CollectionCryptoService] Collection name decrypted successfully:",
        name,
      );
      return name;
    } catch (error) {
      console.error(
        "[CollectionCryptoService] Failed to decrypt collection name:",
        error,
      );
      return "[Encrypted]";
    }
  }

  // === User Key Management ===

  // Decrypt user keys with password from PasswordStorageService
  async decryptUserKeysWithPassword() {
    try {
      await this.initialize();

      // Get password from PasswordStorageService automatically
      const password = await this.getUserPassword();

      // Get stored encrypted user data
      const { default: LocalStorageService } = await import(
        "../Storage/LocalStorageService.js"
      );

      const encryptedData = LocalStorageService.getUserEncryptedData();

      if (
        !encryptedData.salt ||
        !encryptedData.encryptedMasterKey ||
        !encryptedData.encryptedPrivateKey
      ) {
        throw new Error("Missing encrypted user data. Please log in again.");
      }

      console.log(
        "[CollectionCryptoService] Decrypting user keys with password from PasswordStorageService",
      );

      // Decode the encrypted data
      const salt = this.cryptoService.tryDecodeBase64(encryptedData.salt);
      const encryptedMasterKey = this.cryptoService.tryDecodeBase64(
        encryptedData.encryptedMasterKey,
      );
      const encryptedPrivateKey = this.cryptoService.tryDecodeBase64(
        encryptedData.encryptedPrivateKey,
      );
      const publicKey = encryptedData.publicKey
        ? this.cryptoService.tryDecodeBase64(encryptedData.publicKey)
        : null;

      // Derive key encryption key from password
      const keyEncryptionKey = await this.cryptoService.deriveKeyFromPassword(
        password,
        salt,
      );

      // Decrypt master key with KEK
      const masterKey = this.cryptoService.decryptWithSecretBox(
        encryptedMasterKey,
        keyEncryptionKey,
      );

      // Decrypt private key with master key
      const privateKey = this.cryptoService.decryptWithSecretBox(
        encryptedPrivateKey,
        masterKey,
      );

      // Derive public key if not provided
      const derivedPublicKey =
        publicKey || this.sodium.crypto_scalarmult_base(privateKey);

      console.log(
        "[CollectionCryptoService] User keys decrypted successfully using PasswordStorageService",
      );

      return {
        masterKey,
        privateKey,
        publicKey: derivedPublicKey,
        keyEncryptionKey,
      };
    } catch (error) {
      console.error(
        "[CollectionCryptoService] Failed to decrypt user keys:",
        error,
      );
      throw new Error(
        `Failed to decrypt keys: ${error.message}. Please check your password or log in again.`,
      );
    }
  }

  // Get user's encryption keys from session or decrypt with PasswordStorageService
  async getUserKeys() {
    // First check if we have session keys in memory
    const { default: LocalStorageService } = await import(
      "../Storage/LocalStorageService.js"
    );

    const sessionKeys = LocalStorageService.getSessionKeys();

    if (sessionKeys.masterKey && sessionKeys.publicKey) {
      console.log("[CollectionCryptoService] Using session keys from memory");
      return {
        masterKey: sessionKeys.masterKey,
        publicKey: sessionKeys.publicKey,
        privateKey: sessionKeys.privateKey,
      };
    }

    // If no session keys, decrypt keys using PasswordStorageService
    console.log(
      "[CollectionCryptoService] No session keys found, decrypting with PasswordStorageService",
    );
    return await this.decryptUserKeysWithPassword();
  }

  // === Collection Key Encryption/Decryption ===

  // Encrypt collection key with user's master key (matching deprecated format)
  async encryptCollectionKey(collectionKey, userMasterKey) {
    if (!collectionKey || !userMasterKey) {
      throw new Error("Collection key and user master key are required");
    }

    if (!this.sodium) {
      throw new Error("CollectionCryptoService not initialized");
    }

    // Generate nonce
    const nonce = this.sodium.randombytes_buf(
      this.sodium.crypto_secretbox_NONCEBYTES,
    );

    // Encrypt collection key with master key (ChaCha20-Poly1305)
    const encrypted = this.sodium.crypto_secretbox_easy(
      collectionKey,
      nonce,
      userMasterKey,
    );

    // Combine nonce + ciphertext for storage (matching deprecated format)
    const combined = new Uint8Array(nonce.length + encrypted.length);
    combined.set(nonce, 0);
    combined.set(encrypted, nonce.length);

    // Return structure expected by API with base64 strings (matching deprecated format)
    return {
      ciphertext: this.cryptoService.uint8ArrayToBase64(combined), // Base64 string, not array!
      nonce: this.cryptoService.uint8ArrayToBase64(nonce), // Base64 string for separate storage
      key_version: 1,
      rotated_at: new Date().toISOString(),
      previous_keys: [],
    };
  }

  // Decrypt collection key with user's master key (matching deprecated format)
  async decryptCollectionKey(encryptedKeyData, userMasterKey) {
    if (!encryptedKeyData || !userMasterKey) {
      throw new Error("Encrypted key data and user master key are required");
    }

    if (!this.sodium) {
      throw new Error("CollectionCryptoService not initialized");
    }

    try {
      // Decode from base64 (matching deprecated format logic)
      let combined;

      // Handle different formats - some APIs store nonce+ciphertext together
      if (typeof encryptedKeyData.ciphertext === "string") {
        combined = this.cryptoService.tryDecodeBase64(
          encryptedKeyData.ciphertext,
        );
      } else if (Array.isArray(encryptedKeyData.ciphertext)) {
        // Legacy format - convert array to Uint8Array
        const ciphertext = new Uint8Array(encryptedKeyData.ciphertext);
        const nonce = new Uint8Array(encryptedKeyData.nonce);
        combined = new Uint8Array(nonce.length + ciphertext.length);
        combined.set(nonce, 0);
        combined.set(ciphertext, nonce.length);
      } else {
        throw new Error("Invalid encrypted key format");
      }

      // Extract nonce and ciphertext
      const nonceLength = this.sodium.crypto_secretbox_NONCEBYTES;
      const nonce = combined.slice(0, nonceLength);
      const ciphertext = combined.slice(nonceLength);

      console.log(
        `[CollectionCryptoService] Decrypting collection key - nonce: ${nonce.length}, ciphertext: ${ciphertext.length}`,
      );

      // Decrypt with master key
      const decrypted = this.sodium.crypto_secretbox_open_easy(
        ciphertext,
        nonce,
        userMasterKey,
      );

      console.log(
        `[CollectionCryptoService] Collection key decrypted successfully, length: ${decrypted.length}`,
      );
      return decrypted;
    } catch (error) {
      console.error(
        "[CollectionCryptoService] Failed to decrypt collection key:",
        error,
      );

      // Try previous keys if available
      if (
        encryptedKeyData.previous_keys &&
        encryptedKeyData.previous_keys.length > 0
      ) {
        for (const prevKey of encryptedKeyData.previous_keys) {
          try {
            return await this.decryptCollectionKey(prevKey, userMasterKey);
          } catch {
            continue;
          }
        }
      }

      throw error;
    }
  }

  // === Collection Key Sharing (for future use) ===

  // Encrypt collection key for sharing with another user (uses their public key)
  async encryptCollectionKeyForRecipient(collectionKey, recipientPublicKey) {
    if (!collectionKey || !recipientPublicKey) {
      throw new Error("Collection key and recipient public key are required");
    }

    if (!this.sodium) {
      throw new Error("CollectionCryptoService not initialized");
    }

    // Ensure public key is Uint8Array
    const publicKey =
      recipientPublicKey instanceof Uint8Array
        ? recipientPublicKey
        : new Uint8Array(recipientPublicKey);

    // Use sealed box (anonymous encryption) for sharing
    const encrypted = this.sodium.crypto_box_seal(collectionKey, publicKey);

    // Return base64 string as API expects
    return this.cryptoService.uint8ArrayToBase64(encrypted);
  }

  // Decrypt collection key shared with us (uses our private key)
  async decryptSharedCollectionKey(
    encryptedKey,
    userPrivateKey,
    userPublicKey,
  ) {
    if (!encryptedKey || !userPrivateKey || !userPublicKey) {
      throw new Error("Encrypted key and user keypair are required");
    }

    if (!this.sodium) {
      throw new Error("CollectionCryptoService not initialized");
    }

    try {
      // Decode from base64
      const encryptedData =
        typeof encryptedKey === "string"
          ? this.cryptoService.tryDecodeBase64(encryptedKey)
          : new Uint8Array(encryptedKey);

      // Decrypt with our private key
      const decrypted = this.sodium.crypto_box_seal_open(
        encryptedData,
        userPublicKey,
        userPrivateKey,
      );

      return decrypted;
    } catch (error) {
      console.error(
        "[CollectionCryptoService] Failed to decrypt shared collection key:",
        error,
      );
      throw error;
    }
  }

  // === Collection Data Encryption/Decryption ===

  // Encrypt complete collection data for API (uses PasswordStorageService automatically)
  async encryptCollectionForAPI(collectionData) {
    try {
      if (!this.isInitialized) {
        throw new Error("CollectionCryptoService not initialized");
      }

      console.log(
        "[CollectionCryptoService] Encrypting collection data for API using PasswordStorageService",
      );

      // Validate input
      if (!collectionData.name) {
        throw new Error("Collection name is required");
      }

      // Get user's encryption keys using PasswordStorageService
      const userKeys = await this.getUserKeys();

      // Generate collection key
      const collectionKey = this.generateCollectionKey();

      // Generate collection ID
      const collectionId = this.cryptoService.generateUUID();

      // Encrypt collection name with collection key
      const encryptedName = this.encryptCollectionName(
        collectionData.name,
        collectionKey,
      );

      // Encrypt collection key with user's master key
      const encryptedCollectionKey = await this.encryptCollectionKey(
        collectionKey,
        userKeys.masterKey,
      );

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
        "[CollectionCryptoService] Collection data encrypted for API successfully using PasswordStorageService",
      );

      return {
        apiData,
        collectionKey,
        collectionId,
      };
    } catch (error) {
      console.error(
        "[CollectionCryptoService] Failed to encrypt collection for API:",
        error,
      );
      throw error;
    }
  }

  // Decrypt collection data from API response (uses PasswordStorageService automatically)
  async decryptCollectionFromAPI(encryptedCollection, collectionKey = null) {
    if (!encryptedCollection) return null;

    await this.initialize();

    try {
      console.log(
        "[CollectionCryptoService] Decrypting collection data from API using PasswordStorageService:",
        encryptedCollection.id,
      );

      let workingCollectionKey = collectionKey;

      // If no collection key provided, we need to decrypt it
      if (!workingCollectionKey) {
        // Get user keys using PasswordStorageService automatically
        const userKeys = await this.getUserKeys();

        // Check if this is our collection or shared with us
        if (encryptedCollection.encrypted_collection_key) {
          // Our collection - decrypt with master key
          workingCollectionKey = await this.decryptCollectionKey(
            encryptedCollection.encrypted_collection_key,
            userKeys.masterKey,
          );
        } else {
          // Shared collection - find our encrypted key in members
          const ourMembership = encryptedCollection.members?.find(
            (m) => m.recipient_id === userKeys.userId,
          );

          if (ourMembership && ourMembership.encrypted_collection_key) {
            workingCollectionKey = await this.decryptSharedCollectionKey(
              ourMembership.encrypted_collection_key,
              userKeys.privateKey,
              userKeys.publicKey,
            );
          } else {
            throw new Error("No collection key available for decryption");
          }
        }

        // Cache the collection key
        this.cacheCollectionKey(encryptedCollection.id, workingCollectionKey);
      }

      if (!workingCollectionKey) {
        throw new Error("Collection key not available for decryption");
      }

      // Decrypt collection name
      const name = this.decryptCollectionName(
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
        collection_key: workingCollectionKey, // Store for future use (in memory only!)
      };

      console.log(
        "[CollectionCryptoService] Collection data decrypted from API successfully using PasswordStorageService:",
        name,
      );
      return decryptedCollection;
    } catch (error) {
      console.error(
        "[CollectionCryptoService] Failed to decrypt collection from API:",
        error,
      );

      // Return collection with error marker
      return {
        ...encryptedCollection,
        name: "[Unable to decrypt]",
        _isDecrypted: false,
        _decryptionError: error.message,
        decrypt_error: error.message,
      };
    }
  }

  // === Collection Key Cache Management (In-Memory Only) ===

  // Store collection keys in memory (not localStorage!)
  cacheCollectionKey(collectionId, collectionKey) {
    this._collectionKeyCache.set(collectionId, collectionKey);
    console.log(
      `[CollectionCryptoService] Collection key cached in memory for: ${collectionId}`,
    );
  }

  getCachedCollectionKey(collectionId) {
    const cached = this._collectionKeyCache.get(collectionId);
    if (cached) {
      console.log(
        `[CollectionCryptoService] Retrieved collection key from memory for: ${collectionId}`,
      );
    }
    return cached;
  }

  clearCollectionKeyCache() {
    this._collectionKeyCache.clear();
    console.log("[CollectionCryptoService] Collection key cache cleared");
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
      hasSodium: !!this.sodium,
      cryptoServiceReady: this.cryptoService?.isInitialized || false,
      collectionKeyCacheSize: this._collectionKeyCache.size,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "CollectionCryptoService",
      status: this.getStatus(),
      capabilities: [
        "generateCollectionKey",
        "encryptCollectionName",
        "decryptCollectionName",
        "encryptCollectionKey",
        "decryptCollectionKey",
        "encryptCollectionForAPI",
        "decryptCollectionFromAPI",
        "encryptCollectionKeyForRecipient",
        "decryptSharedCollectionKey",
      ],
    };
  }
}

// Export singleton instance
export default new CollectionCryptoService();
