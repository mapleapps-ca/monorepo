// File: monorepo/web/maplefile-frontend/src/services/Manager/User/UserLookupManager.js
// Simplified User Lookup Manager - Orchestrates API service for user public key lookups

import UserLookupAPIService from "../../API/User/UserLookupAPIService.js";

class UserLookupManager {
  constructor() {
    // UserLookupManager is a public service - no auth required
    this.isLoading = false;
    this.isInitialized = false;

    // Initialize API service
    this.apiService = new UserLookupAPIService();

    console.log(
      "[UserLookupManager] User lookup manager initialized for public key lookups",
    );
  }

  // Initialize the manager
  async initialize() {
    try {
      console.log("[UserLookupManager] Initializing user lookup manager...");
      this.isInitialized = true;
      console.log(
        "[UserLookupManager] User lookup manager initialized successfully",
      );
    } catch (error) {
      console.error(
        "[UserLookupManager] Failed to initialize user lookup manager:",
        error,
      );
    }
  }

  // === Basic User Lookup ===

  // Lookup user using API service
  async lookupUser(email) {
    try {
      this.isLoading = true;
      console.log("[UserLookupManager] Looking up user:", email);

      // Use API service which handles all the error logic
      const user = await this.apiService.lookupUser(email);

      console.log("[UserLookupManager] User lookup successful:", user.email);

      return {
        user: user,
        source: "api",
        success: true,
      };
    } catch (error) {
      console.error("[UserLookupManager] User lookup failed:", error);
      throw error; // Pass through the detailed error message from API service
    } finally {
      this.isLoading = false;
    }
  }

  // Check if user exists
  async userExists(email) {
    try {
      console.log("[UserLookupManager] Checking if user exists:", email);
      return await this.apiService.userExists(email);
    } catch (error) {
      console.error("[UserLookupManager] User existence check failed:", error);
      throw error;
    }
  }

  // Get user's public key for encryption
  async getUserPublicKey(email) {
    try {
      console.log("[UserLookupManager] Getting user public key:", email);

      const result = await this.lookupUser(email);

      // Decode public key from base64
      const publicKeyBase64 = result.user.public_key_in_base64;
      if (!publicKeyBase64) {
        throw new Error("No public key available for user");
      }

      const binaryString = atob(publicKeyBase64);
      const bytes = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }

      console.log(
        `[UserLookupManager] Public key retrieved for ${email}, length: ${bytes.length} bytes`,
      );

      return {
        email: result.user.email,
        userId: result.user.user_id,
        name: result.user.name,
        publicKey: bytes,
        verificationId: result.user.verification_id,
        source: result.source,
      };
    } catch (error) {
      console.error(
        "[UserLookupManager] Failed to get user public key:",
        error,
      );
      throw error;
    }
  }

  // === Manager Status ===

  getManagerStatus() {
    return {
      isLoading: this.isLoading,
      isInitialized: this.isInitialized,
      isPublicService: true,
      authRequired: false,
    };
  }

  // === Debug Information ===

  getDebugInfo() {
    return {
      serviceName: "UserLookupManager",
      role: "orchestrator",
      type: "public_service",
      apiService: this.apiService.getDebugInfo(),
      managerStatus: this.getManagerStatus(),
    };
  }
}

export default UserLookupManager;
