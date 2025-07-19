// File: monorepo/web/maplefile-frontend/src/services/Crypto/FileCryptoService.js
// File-specific encryption operations following E2EE architecture

class FileCryptoService {
  constructor() {
    this.isInitialized = false;
    this.sodium = null;

    // In-memory cache for file keys (NEVER stored in localStorage)
    this._fileKeyCache = new Map();

    console.log("[FileCryptoService] File crypto service initialized");
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
        "[FileCryptoService] File crypto service initialized successfully",
      );
    } catch (error) {
      console.error("[FileCryptoService] Failed to initialize:", error);
      throw new Error(
        `Failed to initialize FileCryptoService: ${error.message}`,
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
        "[FileCryptoService] Retrieved password from PasswordStorageService",
      );
      return password;
    } catch (error) {
      console.error("[FileCryptoService] Failed to get password:", error);
      throw error;
    }
  }

  // === User Key Management ===

  // Get user's encryption keys from session or decrypt with PasswordStorageService
  async getUserKeys() {
    // First check if we have session keys in memory
    const { default: LocalStorageService } = await import(
      "../Storage/LocalStorageService.js"
    );

    const sessionKeys = LocalStorageService.getSessionKeys();

    if (sessionKeys.masterKey && sessionKeys.publicKey) {
      console.log("[FileCryptoService] Using session keys from memory");
      return {
        masterKey: sessionKeys.masterKey,
        publicKey: sessionKeys.publicKey,
        privateKey: sessionKeys.privateKey,
      };
    }

    // If no session keys, decrypt keys using PasswordStorageService
    console.log(
      "[FileCryptoService] No session keys found, decrypting with PasswordStorageService",
    );
    return await this.decryptUserKeysWithPassword();
  }

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
        "[FileCryptoService] Decrypting user keys with password from PasswordStorageService",
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
        "[FileCryptoService] User keys decrypted successfully using PasswordStorageService",
      );

      return {
        masterKey,
        privateKey,
        publicKey: derivedPublicKey,
        keyEncryptionKey,
      };
    } catch (error) {
      console.error("[FileCryptoService] Failed to decrypt user keys:", error);
      throw new Error(
        `Failed to decrypt keys: ${error.message}. Please check your password or log in again.`,
      );
    }
  }

  // === File Key Decryption ===

  // Decrypt file key with collection key
  async decryptFileKey(encryptedFileKey, collectionKey) {
    if (!this.sodium) {
      throw new Error("FileCryptoService not initialized");
    }

    if (!encryptedFileKey || !collectionKey) {
      throw new Error("Encrypted file key and collection key are required");
    }

    try {
      console.log("[FileCryptoService] === Decrypting File Key ===");
      console.log(
        "[FileCryptoService] Encrypted file key structure:",
        JSON.stringify(encryptedFileKey),
      );
      console.log(
        "[FileCryptoService] Collection key length:",
        collectionKey.length,
      );

      let ciphertext, nonce;

      if (encryptedFileKey.ciphertext && encryptedFileKey.nonce) {
        // Check if it's base64 strings (from API) or Uint8Array
        if (typeof encryptedFileKey.ciphertext === "string") {
          // From API - base64 strings
          console.log(
            "[FileCryptoService] Decoding base64 ciphertext and nonce",
          );
          ciphertext = this.cryptoService.tryDecodeBase64(
            encryptedFileKey.ciphertext,
          );
          nonce = this.cryptoService.tryDecodeBase64(encryptedFileKey.nonce);
        } else {
          // From encryption - Uint8Array
          console.log(
            "[FileCryptoService] Using Uint8Array ciphertext and nonce",
          );
          ciphertext = new Uint8Array(encryptedFileKey.ciphertext);
          nonce = new Uint8Array(encryptedFileKey.nonce);
        }
      } else {
        throw new Error(
          "Invalid encrypted file key format - missing ciphertext or nonce",
        );
      }

      console.log(
        `[FileCryptoService] Decrypting file key - nonce: ${nonce.length}, ciphertext: ${ciphertext.length}`,
      );

      // Decrypt file key
      const fileKey = this.sodium.crypto_secretbox_open_easy(
        ciphertext,
        nonce,
        collectionKey,
      );

      console.log(
        `[FileCryptoService] ✅ File key decrypted successfully, length: ${fileKey.length}`,
      );
      return fileKey;
    } catch (error) {
      console.error(
        "[FileCryptoService] ❌ File key decryption failed:",
        error,
      );
      console.error("[FileCryptoService] Error details:", {
        message: error.message,
        encryptedFileKeyStructure: JSON.stringify(encryptedFileKey),
        collectionKeyLength: collectionKey?.length,
      });
      throw new Error(`File key decryption failed: ${error.message}`);
    }
  }

  // === File Metadata Decryption ===

  // Decrypt file metadata with file key
  async decryptFileMetadata(encryptedMetadata, fileKey) {
    if (!this.cryptoService) {
      throw new Error("FileCryptoService not initialized");
    }

    try {
      console.log("[FileCryptoService] Decrypting file metadata");

      // Decrypt metadata
      const decryptedMetadataBytes = await this.cryptoService.decryptWithKey(
        encryptedMetadata,
        fileKey,
      );

      // Parse the metadata JSON
      const metadataString = new TextDecoder().decode(decryptedMetadataBytes);
      const metadata = JSON.parse(metadataString);

      console.log("[FileCryptoService] File metadata decrypted successfully");
      return metadata;
    } catch (error) {
      console.error(
        "[FileCryptoService] File metadata decryption failed:",
        error,
      );
      throw new Error(`Metadata decryption failed: ${error.message}`);
    }
  }

  // === Complete File Decryption ===

  // Decrypt complete file data from API response
  async decryptFileFromAPI(encryptedFile, collectionKey = null) {
    if (!encryptedFile) return null;

    await this.initialize();

    try {
      console.log("[FileCryptoService] === Decrypting file from API ===");
      console.log("[FileCryptoService] File ID:", encryptedFile.id);
      console.log(
        "[FileCryptoService] Collection ID:",
        encryptedFile.collection_id,
      );
      console.log(
        "[FileCryptoService] Has encrypted_file_key:",
        !!encryptedFile.encrypted_file_key,
      );
      console.log(
        "[FileCryptoService] Has encrypted_metadata:",
        !!encryptedFile.encrypted_metadata,
      );

      let workingCollectionKey = collectionKey;

      // If no collection key provided, get it from cache or collection crypto service
      if (!workingCollectionKey) {
        const { default: CollectionCryptoService } = await import(
          "./CollectionCryptoService.js"
        );

        workingCollectionKey = CollectionCryptoService.getCachedCollectionKey(
          encryptedFile.collection_id,
        );

        if (!workingCollectionKey) {
          throw new Error(
            "Collection key not found. Please ensure the collection is loaded first.",
          );
        }
      }

      if (!workingCollectionKey) {
        throw new Error("Collection key not available for file decryption");
      }

      console.log(
        "[FileCryptoService] Using collection key, length:",
        workingCollectionKey.length,
      );

      // ✅ ENHANCEMENT: Check if file key is already cached in memory
      let fileKey = this.getCachedFileKey(encryptedFile.id);

      if (fileKey) {
        console.log(
          "[FileCryptoService] Using cached file key, length:",
          fileKey.length,
        );
      } else {
        // Step 1: Decrypt file key
        console.log("[FileCryptoService] Step 1: Decrypting file key");
        fileKey = await this.decryptFileKey(
          encryptedFile.encrypted_file_key,
          workingCollectionKey,
        );

        console.log(
          "[FileCryptoService] File key decrypted successfully, length:",
          fileKey.length,
        );

        // Cache the file key in memory
        this.cacheFileKey(encryptedFile.id, fileKey);
      }

      // Step 2: Decrypt metadata if available
      let metadata = null;
      let name = "[Unable to decrypt]";
      let mimeType = "application/octet-stream";
      let size = 0;

      if (encryptedFile.encrypted_metadata) {
        try {
          console.log("[FileCryptoService] Step 2: Decrypting file metadata");
          metadata = await this.decryptFileMetadata(
            encryptedFile.encrypted_metadata,
            fileKey,
          );
          name = metadata.name || "[Unknown]";
          mimeType = metadata.mime_type || "application/octet-stream";
          size = metadata.size || 0;

          console.log("[FileCryptoService] Metadata decrypted successfully:");
          console.log("[FileCryptoService] - Name:", name);
          console.log("[FileCryptoService] - MIME type:", mimeType);
          console.log("[FileCryptoService] - Size:", size);
        } catch (metadataError) {
          console.warn(
            "[FileCryptoService] Metadata decryption failed:",
            metadataError.message,
          );
        }
      } else {
        console.log("[FileCryptoService] No encrypted metadata available");
      }

      // Return decrypted file
      const decryptedFile = {
        ...encryptedFile,
        name,
        mime_type: mimeType,
        size,
        _isDecrypted: true,
        _hasFileKey: true,
        _originalEncryptedMetadata: encryptedFile.encrypted_metadata,
        _file_key: fileKey, // ✅ Store for future use (in memory only!)
        _decrypted_metadata: metadata,
        _fileKeyCached: true, // ✅ Flag to indicate file key is cached
      };

      console.log("[FileCryptoService] ✅ File decrypted successfully:", name);
      return decryptedFile;
    } catch (error) {
      console.error(
        "[FileCryptoService] ❌ Failed to decrypt file from API:",
        error,
      );

      // Return file with error marker
      return {
        ...encryptedFile,
        name: "[Unable to decrypt]",
        _isDecrypted: false,
        _decryptionError: error.message,
        decrypt_error: error.message,
      };
    }
  }

  // ✅ ENHANCEMENT: Get file key with automatic re-decryption if needed
  async getFileKeyForDownload(fileId, collectionId, encryptedFileKey) {
    // First check in-memory cache
    let fileKey = this.getCachedFileKey(fileId);

    if (fileKey) {
      console.log(
        "[FileCryptoService] File key found in memory cache for:",
        fileId,
      );
      return fileKey;
    }

    // If not in cache, need to decrypt it
    console.log(
      "[FileCryptoService] File key not in cache, decrypting for:",
      fileId,
    );

    // Get collection key
    const { default: CollectionCryptoService } = await import(
      "./CollectionCryptoService.js"
    );

    const collectionKey =
      CollectionCryptoService.getCachedCollectionKey(collectionId);

    if (!collectionKey) {
      throw new Error(
        "Collection key not available. Please ensure the collection is loaded first.",
      );
    }

    if (!encryptedFileKey) {
      throw new Error("Encrypted file key not provided for decryption.");
    }

    // Decrypt the file key
    fileKey = await this.decryptFileKey(encryptedFileKey, collectionKey);

    // Cache it for future use
    this.cacheFileKey(fileId, fileKey);

    console.log(
      "[FileCryptoService] File key decrypted and cached for:",
      fileId,
    );

    return fileKey;
  }

  // ✅ ENHANCEMENT: Check if file has cached key
  hasFileKey(fileId) {
    return this._fileKeyCache.has(fileId);
  }

  // ✅ ENHANCEMENT: Get file key cache statistics
  getFileKeyCacheStats() {
    return {
      totalKeys: this._fileKeyCache.size,
      fileIds: Array.from(this._fileKeyCache.keys()).slice(0, 10), // Show first 10
    };
  }

  // ✅ ENHANCEMENT: Cleanup expired file keys (if needed in the future)
  cleanupFileKeyCache(maxAge = 60 * 60 * 1000) {
    // 1 hour default
    // For now, just clear all since we don't track timestamps
    // In the future, we could add timestamps to track key age
    console.log(
      "[FileCryptoService] File key cache cleanup - clearing all keys for security",
    );
    this.clearFileKeyCache();
  }

  // Decrypt multiple files
  async decryptFilesFromAPI(encryptedFiles, collectionKey = null) {
    if (!encryptedFiles || !Array.isArray(encryptedFiles)) {
      return [];
    }

    console.log("[FileCryptoService] === Decrypting multiple files ===");
    console.log("[FileCryptoService] File count:", encryptedFiles.length);
    console.log(
      "[FileCryptoService] Collection key available:",
      !!collectionKey,
    );

    const decryptedFiles = [];

    for (let i = 0; i < encryptedFiles.length; i++) {
      const file = encryptedFiles[i];
      try {
        console.log(
          `[FileCryptoService] Decrypting file ${i + 1}/${encryptedFiles.length}: ${file.id}`,
        );
        const decryptedFile = await this.decryptFileFromAPI(
          file,
          collectionKey,
        );
        decryptedFiles.push(decryptedFile);

        if (decryptedFile._isDecrypted) {
          console.log(
            `[FileCryptoService] ✅ File ${i + 1} decrypted: ${decryptedFile.name}`,
          );
        } else {
          console.log(
            `[FileCryptoService] ❌ File ${i + 1} decryption failed: ${decryptedFile._decryptionError}`,
          );
        }
      } catch (fileError) {
        console.error(
          `[FileCryptoService] ❌ Failed to decrypt file ${file.id}:`,
          fileError.message,
        );
        // Add the file with error info
        decryptedFiles.push({
          ...file,
          name: `[Decrypt failed: ${fileError.message.substring(0, 50)}...]`,
          _isDecrypted: false,
          _decryptionError: fileError.message,
        });
      }
    }

    const successCount = decryptedFiles.filter((f) => f._isDecrypted).length;
    const errorCount = decryptedFiles.filter((f) => f._decryptionError).length;

    console.log(`[FileCryptoService] === Decryption Summary ===`);
    console.log(`[FileCryptoService] Total files: ${decryptedFiles.length}`);
    console.log(`[FileCryptoService] Successfully decrypted: ${successCount}`);
    console.log(`[FileCryptoService] Decryption errors: ${errorCount}`);

    return decryptedFiles;
  }

  // === File Key Cache Management (In-Memory Only) ===

  // Store file keys in memory (not localStorage!)
  cacheFileKey(fileId, fileKey) {
    this._fileKeyCache.set(fileId, fileKey);
    console.log(`[FileCryptoService] File key cached in memory for: ${fileId}`);
  }

  getCachedFileKey(fileId) {
    const cached = this._fileKeyCache.get(fileId);
    if (cached) {
      console.log(
        `[FileCryptoService] Retrieved file key from memory for: ${fileId}`,
      );
    }
    return cached;
  }

  clearFileKeyCache() {
    this._fileKeyCache.clear();
    console.log("[FileCryptoService] File key cache cleared");
  }

  // === File Data Encryption/Decryption ===

  // Decrypt file content with file key
  async decryptFileContent(encryptedContent, fileKey) {
    if (!this.cryptoService) {
      throw new Error("FileCryptoService not initialized");
    }

    try {
      console.log("[FileCryptoService] Decrypting file content");

      // Convert blob to array buffer if needed
      let contentBytes;
      if (encryptedContent instanceof Blob) {
        const arrayBuffer = await encryptedContent.arrayBuffer();
        contentBytes = new Uint8Array(arrayBuffer);
      } else if (encryptedContent instanceof Uint8Array) {
        contentBytes = encryptedContent;
      } else {
        throw new Error("Invalid encrypted content format");
      }

      // Convert to base64 for decryption
      const encryptedBase64 =
        this.cryptoService.uint8ArrayToBase64(contentBytes);

      // Decrypt content
      const decryptedBytes = await this.cryptoService.decryptWithKey(
        encryptedBase64,
        fileKey,
      );

      console.log("[FileCryptoService] File content decrypted successfully");
      return decryptedBytes;
    } catch (error) {
      console.error(
        "[FileCryptoService] File content decryption failed:",
        error,
      );
      throw new Error(`Content decryption failed: ${error.message}`);
    }
  }

  // === Utility Methods ===

  // Normalize file object with computed properties
  normalizeFile(file) {
    return {
      ...file,
      // Ensure version is present (default to 1 for new files)
      version: file.version || 1,
      // Ensure state is present (default to active)
      state: file.state || "active",
      // Ensure tombstone fields are present
      tombstone_version: file.tombstone_version || 0,
      tombstone_expiry: file.tombstone_expiry || "0001-01-01T00:00:00Z",
      // Add computed properties for easier checking
      _is_deleted: file.state === "deleted",
      _is_archived: file.state === "archived",
      _is_pending: file.state === "pending",
      _is_active: file.state === "active",
      _has_tombstone: (file.tombstone_version || 0) > 0,
      _tombstone_expired: file.tombstone_expiry
        ? new Date(file.tombstone_expiry) < new Date() &&
          file.tombstone_expiry !== "0001-01-01T00:00:00Z"
        : false,
    };
  }

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
      fileKeyCacheSize: this._fileKeyCache.size,
    };
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "FileCryptoService",
      status: this.getStatus(),
      capabilities: [
        "decryptFileKey",
        "decryptFileMetadata",
        "decryptFileFromAPI",
        "decryptFilesFromAPI",
        "decryptFileContent",
        "normalizeFile",
      ],
    };
  }
}

// Export singleton instance
export default new FileCryptoService();
