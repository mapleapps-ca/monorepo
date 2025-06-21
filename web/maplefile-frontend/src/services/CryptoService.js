// web/maplefile-frontend/src/services/CryptoService.js
// src/services/CryptoService.js - LibSodium Implementation
class CryptoService {
  constructor() {
    this.sodium = null;
    this.isReady = false;
    console.log("Initializing CryptoService with libsodium-wrappers-sumo");
  }

  async init() {
    if (this.isReady && this.sodium) {
      return;
    }

    try {
      console.log("Loading libsodium-wrappers-sumo...");

      // Dynamic import with proper error handling
      const sodiumModule = await import("libsodium-wrappers-sumo");
      const sodium = sodiumModule.default || sodiumModule;

      console.log("LibSodium module loaded, waiting for ready...");
      await sodium.ready;

      this.sodium = sodium;
      this.isReady = true;

      console.log("✅ LibSodium initialized successfully");
      console.log("Available methods:", {
        randombytes_buf: typeof sodium.randombytes_buf,
        crypto_box_keypair: typeof sodium.crypto_box_keypair,
        crypto_pwhash: typeof sodium.crypto_pwhash,
        crypto_secretbox_easy: typeof sodium.crypto_secretbox_easy,
        crypto_box_easy: typeof sodium.crypto_box_easy,
        crypto_pwhash_ALG_ARGON2ID: typeof sodium.crypto_pwhash_ALG_ARGON2ID, // Add this line to log
      });
    } catch (error) {
      console.error("❌ Failed to initialize LibSodium:", error);
      throw new Error(`LibSodium initialization failed: ${error.message}`);
    }
  }

  async ensureReady() {
    if (!this.isReady) {
      await this.init();
    }

    if (!this.sodium) {
      throw new Error("LibSodium not available");
    }
  }
}

export default CryptoService;
