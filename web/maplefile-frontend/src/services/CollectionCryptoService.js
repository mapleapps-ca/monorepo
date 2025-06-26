// CollectionCryptoService.js - Updated for password-based decryption
import CryptoService from "./CryptoService.js";
import LocalStorageService from "./LocalStorageService.js";
import sodium from "libsodium-wrappers-sumo";

class CollectionCryptoService {
  constructor() {
    this.isInitialized = false;
    this.sodium = null;
  }

  async initialize() {
    if (this.isInitialized) return;

    await CryptoService.initialize();
    await sodium.ready;
    this.sodium = sodium;
    this.isInitialized = true;
    console.log("[CollectionCryptoService] Initialized with libsodium");
  }

  // Generate a new collection key
  generateCollectionKey() {
    if (!this.sodium) {
      throw new Error("CollectionCryptoService not initialized");
    }

    // Generate a 32-byte key for the collection
    return this.sodium.randombytes_buf(32);
  }

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
    return CryptoService.uint8ArrayToBase64(combined);
  }

  // Decrypt collection name with collection key
  decryptCollectionName(encryptedName, collectionKey) {
    if (!encryptedName || !collectionKey) {
      return ""; // Return empty string if can't decrypt
    }

    if (!this.sodium) {
      console.error("[CollectionCrypto] Service not initialized");
      return "[Not initialized]";
    }

    try {
      // Decode from base64
      const combined = CryptoService.tryDecodeBase64(encryptedName);

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
      return decoder.decode(decrypted);
    } catch (error) {
      console.error("[CollectionCrypto] Failed to decrypt name:", error);
      return "[Encrypted]";
    }
  }

  // Decrypt user keys with password (on-demand)
  async decryptUserKeysWithPassword(password) {
    try {
      await this.initialize();

      // Get stored encrypted user data
      const encryptedData = LocalStorageService.getUserEncryptedData();

      if (
        !encryptedData.salt ||
        !encryptedData.encryptedMasterKey ||
        !encryptedData.encryptedPrivateKey
      ) {
        throw new Error("Missing encrypted user data. Please log in again.");
      }

      console.log("[CollectionCrypto] Decrypting user keys with password");

      // Decode the encrypted data
      const salt = CryptoService.tryDecodeBase64(encryptedData.salt);
      const encryptedMasterKey = CryptoService.tryDecodeBase64(
        encryptedData.encryptedMasterKey,
      );
      const encryptedPrivateKey = CryptoService.tryDecodeBase64(
        encryptedData.encryptedPrivateKey,
      );
      const publicKey = encryptedData.publicKey
        ? CryptoService.tryDecodeBase64(encryptedData.publicKey)
        : null;

      // Derive key encryption key from password
      const keyEncryptionKey = await CryptoService.deriveKeyFromPassword(
        password,
        salt,
      );

      // Decrypt master key with KEK
      const masterKey = CryptoService.decryptWithSecretBox(
        encryptedMasterKey,
        keyEncryptionKey,
      );

      // Decrypt private key with master key
      const privateKey = CryptoService.decryptWithSecretBox(
        encryptedPrivateKey,
        masterKey,
      );

      // Derive public key if not provided
      const derivedPublicKey =
        publicKey || CryptoService.sodium.crypto_scalarmult_base(privateKey);

      console.log("[CollectionCrypto] User keys decrypted successfully");

      return {
        masterKey,
        privateKey,
        publicKey: derivedPublicKey,
        keyEncryptionKey,
      };
    } catch (error) {
      console.error("[CollectionCrypto] Failed to decrypt user keys:", error);
      throw new Error(
        `Failed to decrypt keys: ${error.message}. Please check your password.`,
      );
    }
  }

  // Get user's encryption keys from session or decrypt with password
  async getUserKeys(password = null) {
    // First check if we have session keys in memory
    const sessionKeys = LocalStorageService.getSessionKeys();

    if (sessionKeys.masterKey && sessionKeys.publicKey) {
      console.log("[CollectionCrypto] Using session keys from memory");
      return {
        masterKey: sessionKeys.masterKey,
        publicKey: sessionKeys.publicKey,
        privateKey: sessionKeys.privateKey,
      };
    }

    // If no session keys and no password provided, throw error
    if (!password) {
      throw new Error("Password required to decrypt collection keys");
    }

    // Decrypt keys with password
    return await this.decryptUserKeysWithPassword(password);
  }

  // Encrypt collection key with user's master key
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

    // Combine nonce + ciphertext for storage
    const combined = new Uint8Array(nonce.length + encrypted.length);
    combined.set(nonce, 0);
    combined.set(encrypted, nonce.length);

    // Return structure expected by API with base64 strings
    return {
      ciphertext: CryptoService.uint8ArrayToBase64(combined), // Base64 string, not array!
      nonce: CryptoService.uint8ArrayToBase64(nonce), // Base64 string for separate storage
      key_version: 1,
      rotated_at: new Date().toISOString(),
      previous_keys: [],
    };
  }

  // Decrypt collection key with user's master key
  async decryptCollectionKey(encryptedKeyData, userMasterKey) {
    if (!encryptedKeyData || !userMasterKey) {
      throw new Error("Encrypted key data and user master key are required");
    }

    if (!this.sodium) {
      throw new Error("CollectionCryptoService not initialized");
    }

    try {
      // Decode from base64
      let combined;

      // Handle different formats - some APIs store nonce+ciphertext together
      if (typeof encryptedKeyData.ciphertext === "string") {
        combined = CryptoService.tryDecodeBase64(encryptedKeyData.ciphertext);
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

      // Decrypt with master key
      const decrypted = this.sodium.crypto_secretbox_open_easy(
        ciphertext,
        nonce,
        userMasterKey,
      );

      return decrypted;
    } catch (error) {
      console.error(
        "[CollectionCrypto] Failed to decrypt collection key:",
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
          } catch (e) {
            continue;
          }
        }
      }

      throw error;
    }
  }

  // Encrypt collection key for sharing with another user (uses box_seal)
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
    return CryptoService.uint8ArrayToBase64(encrypted);
  }

  // Decrypt collection key shared with us
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
          ? CryptoService.tryDecodeBase64(encryptedKey)
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
        "[CollectionCrypto] Failed to decrypt shared collection key:",
        error,
      );
      throw error;
    }
  }

  // Prepare collection data for API with password
  async prepareCollectionForAPIWithPassword(collectionData, password) {
    await this.initialize();

    // Get user's encryption keys by decrypting with password
    const userKeys = await this.getUserKeys(password);

    // Generate collection key
    const collectionKey = this.generateCollectionKey();

    // Generate collection ID
    const collectionId = CryptoService.generateUUID();

    // Encrypt collection name with collection key
    const encryptedName = this.encryptCollectionName(
      collectionData.name || "Untitled Collection",
      collectionKey,
    );

    // Encrypt collection key with user's master key
    const encryptedCollectionKey = await this.encryptCollectionKey(
      collectionKey,
      userKeys.masterKey,
    );

    // Prepare the API request
    const apiData = {
      id: collectionId,
      encrypted_name: encryptedName,
      collection_type: collectionData.collection_type || "folder",
      encrypted_collection_key: encryptedCollectionKey,
    };

    // Only add optional fields if they have values
    if (collectionData.parent_id) {
      apiData.parent_id = collectionData.parent_id;
    }
    if (collectionData.ancestor_ids && collectionData.ancestor_ids.length > 0) {
      apiData.ancestor_ids = collectionData.ancestor_ids;
    }

    // If sharing during creation, add members
    if (collectionData.members && collectionData.members.length > 0) {
      apiData.members = await Promise.all(
        collectionData.members.map(async (member) => {
          // Encrypt collection key for each recipient
          const encryptedKey = await this.encryptCollectionKeyForRecipient(
            collectionKey,
            member.recipient_public_key,
          );

          return {
            recipient_id: member.recipient_id,
            recipient_email: member.recipient_email,
            encrypted_collection_key: encryptedKey, // Now a base64 string
            permission_level: member.permission_level || "read_only",
          };
        }),
      );
    }

    // Clear the decrypted keys from memory if we decrypted them
    if (password) {
      // We decrypted keys on-demand, so don't store them
      console.log(
        "[CollectionCrypto] Clearing temporary decrypted keys from memory",
      );
    }

    return { apiData, collectionKey, collectionId };
  }

  // Original method updated to require password if no session keys
  async prepareCollectionForAPI(collectionData) {
    // Check if we have session keys
    const sessionKeys = LocalStorageService.getSessionKeys();
    if (!sessionKeys.masterKey) {
      throw new Error(
        "Password required to create collection - session keys not available",
      );
    }

    // Use the password-based method with null password (will use session keys)
    return this.prepareCollectionForAPIWithPassword(collectionData, null);
  }

  // Decrypt collection data from API (with optional password)
  async decryptCollectionFromAPI(encryptedCollection, password = null) {
    if (!encryptedCollection) return null;

    await this.initialize();

    try {
      // Get user keys - either from session or by decrypting with password
      const userKeys = await this.getUserKeys(password);

      // Determine if this is our collection or shared with us
      const isOwnCollection =
        !encryptedCollection.members ||
        encryptedCollection.members.some(
          (m) =>
            m.recipient_id === userKeys.userId &&
            m.permission_level === "admin",
        );

      let collectionKey;

      if (isOwnCollection || encryptedCollection.encrypted_collection_key) {
        // Our collection - decrypt with master key
        collectionKey = await this.decryptCollectionKey(
          encryptedCollection.encrypted_collection_key,
          userKeys.masterKey,
        );
      } else {
        // Shared collection - find our encrypted key in members
        const ourMembership = encryptedCollection.members?.find(
          (m) => m.recipient_id === userKeys.userId,
        );

        if (ourMembership && ourMembership.encrypted_collection_key) {
          collectionKey = await this.decryptSharedCollectionKey(
            ourMembership.encrypted_collection_key,
            userKeys.privateKey,
            userKeys.publicKey,
          );
        } else {
          throw new Error("No collection key available for decryption");
        }
      }

      // Cache the collection key
      this.cacheCollectionKey(encryptedCollection.id, collectionKey);

      // Decrypt collection name
      const name = this.decryptCollectionName(
        encryptedCollection.encrypted_name,
        collectionKey,
      );

      // Return decrypted collection
      return {
        ...encryptedCollection,
        name, // Add decrypted name
        collection_key: collectionKey, // Store for future use (in memory only!)
        // Keep encrypted versions for reference
        _encrypted_name: encryptedCollection.encrypted_name,
        _encrypted_collection_key: encryptedCollection.encrypted_collection_key,
      };
    } catch (error) {
      console.error("[CollectionCrypto] Failed to decrypt collection:", error);

      // Return collection with placeholder name
      return {
        ...encryptedCollection,
        name: "[Unable to decrypt]",
        decrypt_error: error.message,
      };
    }
  }

  // Batch decrypt collections (with optional password)
  async decryptCollections(encryptedCollections, password = null) {
    if (!encryptedCollections || !Array.isArray(encryptedCollections)) {
      return [];
    }

    // If password provided, decrypt keys once and use for all collections
    let userKeys = null;
    if (password) {
      userKeys = await this.getUserKeys(password);
      // Temporarily store in session for batch operation
      LocalStorageService.setSessionKeys(
        userKeys.masterKey,
        userKeys.privateKey,
        userKeys.publicKey,
        userKeys.keyEncryptionKey,
      );
    }

    try {
      const results = await Promise.all(
        encryptedCollections.map((collection) =>
          this.decryptCollectionFromAPI(collection),
        ),
      );

      return results;
    } finally {
      // Clear temporary session keys if we set them
      if (password) {
        LocalStorageService.clearSessionKeys();
      }
    }
  }

  // Store collection keys in memory (not localStorage!)
  _collectionKeyCache = new Map();

  cacheCollectionKey(collectionId, collectionKey) {
    this._collectionKeyCache.set(collectionId, collectionKey);
  }

  getCachedCollectionKey(collectionId) {
    return this._collectionKeyCache.get(collectionId);
  }

  clearCollectionKeyCache() {
    this._collectionKeyCache.clear();
  }

  // Update collection with existing key
  async updateCollectionForAPI(collectionId, updateData, currentVersion) {
    await this.initialize();

    // Get cached collection key
    const collectionKey = this.getCachedCollectionKey(collectionId);
    if (!collectionKey) {
      throw new Error(
        "Collection key not found. Please reload the collection.",
      );
    }

    const apiData = {
      id: collectionId,
      version: currentVersion,
    };

    // Encrypt updated name if provided
    if (updateData.name !== undefined) {
      apiData.encrypted_name = this.encryptCollectionName(
        updateData.name,
        collectionKey,
      );
    }

    if (updateData.collection_type !== undefined) {
      apiData.collection_type = updateData.collection_type;
    }

    return apiData;
  }
}

// Export singleton instance
export default new CollectionCryptoService();
