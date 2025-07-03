// File: src/services/WorkerManager.js - FIXED VERSION with Token Decryption
import LocalStorageService from "./LocalStorageService.js";
import passwordStorageService from "./PasswordStorageService.js";

class WorkerManager {
  constructor() {
    this.authWorker = null;
    this.isInitialized = false;
    this.authStateListeners = new Set();
    this.workerReadyPromise = null;
    this.workerReadyResolve = null;
    this.refreshPromises = new Map();
  }

  async initialize() {
    if (this.isInitialized) return;

    try {
      console.log("[WorkerManager] Starting initialization...");

      // Create a promise that resolves when worker is ready
      this.workerReadyPromise = new Promise((resolve) => {
        this.workerReadyResolve = resolve;
      });

      // Initialize the auth worker
      this.authWorker = new Worker("/auth-worker.js");

      // Set up message handling
      this.authWorker.onmessage = (event) => {
        this.handleWorkerMessage(event);
      };

      // Set up error handling
      this.authWorker.onerror = (error) => {
        console.error("[WorkerManager] Auth worker error:", error);
        // Reject any pending refresh promises
        this.refreshPromises.forEach((promise) => {
          promise.reject(new Error("Worker error: " + error.message));
        });
        this.refreshPromises.clear();
      };

      // Wait for worker ready signal
      await this.workerReadyPromise;

      this.isInitialized = true;
      console.log("[WorkerManager] Initialized successfully");

      // Start monitoring if we have tokens
      if (LocalStorageService.hasValidTokens()) {
        this.startMonitoring();
      }
    } catch (error) {
      console.error("[WorkerManager] Failed to initialize:", error);
      this.isInitialized = false;
      throw error;
    }
  }

  startMonitoring() {
    console.log("[WorkerManager] Starting token monitoring");
    if (this.authWorker && this.isInitialized) {
      this.authWorker.postMessage({ type: "start_monitoring" });
    } else {
      console.warn(
        "[WorkerManager] Cannot start monitoring - worker not initialized",
      );
    }
  }

  stopMonitoring() {
    console.log("[WorkerManager] Stopping token monitoring");
    if (this.authWorker) {
      this.authWorker.postMessage({ type: "stop_monitoring" });
    }
  }

  forceTokenCheck() {
    console.log("[WorkerManager] Forcing token check");
    if (this.authWorker && this.isInitialized) {
      const storageData = this.getCurrentStorageData();
      this.authWorker.postMessage({
        type: "force_token_check",
        data: storageData,
      });
    }
  }

  async manualRefresh() {
    console.log("[WorkerManager] Manual token refresh requested");

    if (!this.authWorker || !this.isInitialized) {
      throw new Error("Worker not initialized");
    }

    const refreshToken = LocalStorageService.getRefreshToken();
    if (!refreshToken) {
      throw new Error("No refresh token available");
    }

    // Create a promise for this refresh request
    const requestId = Date.now() + Math.random();

    return new Promise((resolve, reject) => {
      // Store the promise handlers
      this.refreshPromises.set(requestId, { resolve, reject });

      // Set a timeout
      const timeout = setTimeout(() => {
        this.refreshPromises.delete(requestId);
        reject(new Error("Token refresh timeout"));
      }, 30000); // 30 second timeout

      // Update promise handlers to clear timeout
      const originalResolve = resolve;
      const originalReject = reject;

      this.refreshPromises.set(requestId, {
        resolve: (result) => {
          clearTimeout(timeout);
          this.refreshPromises.delete(requestId);
          originalResolve(result);
        },
        reject: (error) => {
          clearTimeout(timeout);
          this.refreshPromises.delete(requestId);
          originalReject(error);
        },
      });

      // Send refresh request to worker
      this.authWorker.postMessage({
        type: "manual_refresh",
        data: {
          refreshToken: refreshToken,
          storageData: this.getCurrentStorageData(),
          requestId: requestId,
        },
      });
    });
  }

  getCurrentStorageData() {
    try {
      return {
        mapleapps_access_token: LocalStorageService.getAccessToken(),
        mapleapps_refresh_token: LocalStorageService.getRefreshToken(),
        mapleapps_access_token_expiry: localStorage.getItem(
          "mapleapps_access_token_expiry",
        ),
        mapleapps_refresh_token_expiry: localStorage.getItem(
          "mapleapps_refresh_token_expiry",
        ),
        mapleapps_user_email: LocalStorageService.getUserEmail(),
      };
    } catch (error) {
      console.error("[WorkerManager] Failed to get storage data:", error);
      return {};
    }
  }

  async handleDecryptTokensRequest(data) {
    console.log("[WorkerManager] Received decrypt tokens request from worker");

    try {
      const password = passwordStorageService.getPassword();
      if (!password) {
        throw new Error("No password available for token decryption");
      }

      // Check if we have user's encrypted data
      const userEncryptedData = LocalStorageService.getUserEncryptedData();
      if (
        !userEncryptedData.salt ||
        !userEncryptedData.encryptedMasterKey ||
        !userEncryptedData.encryptedPrivateKey
      ) {
        throw new Error("Missing user encrypted data for token decryption");
      }

      // Decrypt tokens using password
      const { default: CryptoService } = await import("./CryptoService.js");
      await CryptoService.initialize();

      // First, derive keys from password
      const salt = CryptoService.tryDecodeBase64(userEncryptedData.salt);
      const encryptedMasterKey = CryptoService.tryDecodeBase64(
        userEncryptedData.encryptedMasterKey,
      );
      const encryptedPrivateKey = CryptoService.tryDecodeBase64(
        userEncryptedData.encryptedPrivateKey,
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

      // Decrypt private key
      const privateKey = CryptoService.decryptWithSecretBox(
        encryptedPrivateKey,
        masterKey,
      );

      // Now decrypt the tokens
      // Try different decryption methods as tokens might be encrypted with different keys
      const encryptedAccessToken = CryptoService.tryDecodeBase64(
        data.encryptedAccessToken,
      );
      const encryptedRefreshToken = CryptoService.tryDecodeBase64(
        data.encryptedRefreshToken,
      );
      const nonce = CryptoService.tryDecodeBase64(data.tokenNonce);

      let decryptedAccessToken, decryptedRefreshToken;

      // Try sealed box first (anonymous encryption with private key)
      try {
        console.log("[WorkerManager] Trying sealed box decryption for tokens");
        const decryptedAccess = await CryptoService.decryptChallenge(
          encryptedAccessToken,
          privateKey,
        );
        const decryptedRefresh = await CryptoService.decryptChallenge(
          encryptedRefreshToken,
          privateKey,
        );

        decryptedAccessToken = new TextDecoder().decode(decryptedAccess);
        decryptedRefreshToken = new TextDecoder().decode(decryptedRefresh);
        console.log("[WorkerManager] Tokens decrypted using sealed box");
      } catch (sealError) {
        console.log(
          "[WorkerManager] Sealed box failed, trying secretbox with master key",
        );

        // Try secretbox with master key
        try {
          const accessTokenData = new Uint8Array(
            nonce.length + encryptedAccessToken.length,
          );
          accessTokenData.set(nonce, 0);
          accessTokenData.set(encryptedAccessToken, nonce.length);

          const refreshTokenData = new Uint8Array(
            nonce.length + encryptedRefreshToken.length,
          );
          refreshTokenData.set(nonce, 0);
          refreshTokenData.set(encryptedRefreshToken, nonce.length);

          const decryptedAccess = CryptoService.decryptWithSecretBox(
            accessTokenData,
            masterKey,
          );
          const decryptedRefresh = CryptoService.decryptWithSecretBox(
            refreshTokenData,
            masterKey,
          );

          decryptedAccessToken = new TextDecoder().decode(decryptedAccess);
          decryptedRefreshToken = new TextDecoder().decode(decryptedRefresh);
          console.log("[WorkerManager] Tokens decrypted using secretbox");
        } catch (error) {
          throw new Error("Failed to decrypt tokens with available keys");
        }
      }

      // Store decrypted tokens
      LocalStorageService.setTokens(
        decryptedAccessToken,
        decryptedRefreshToken,
        data.accessTokenExpiry,
        data.refreshTokenExpiry,
      );

      if (data.username) {
        LocalStorageService.setUserEmail(data.username);
      }

      // Send success response to worker
      this.authWorker.postMessage({
        type: "decrypted_tokens_response",
        data: {
          accessToken: decryptedAccessToken,
          refreshToken: decryptedRefreshToken,
          accessTokenExpiry: data.accessTokenExpiry,
          refreshTokenExpiry: data.refreshTokenExpiry,
          username: data.username,
          requestId: data.requestId,
        },
      });

      console.log("[WorkerManager] Tokens decrypted and stored successfully");
    } catch (error) {
      console.error("[WorkerManager] Failed to decrypt tokens:", error);

      // Send failure response to worker
      this.authWorker.postMessage({
        type: "decrypt_tokens_failed",
        data: {
          error: error.message,
          requestId: data.requestId,
        },
      });
    }
  }

  handleWorkerMessage(event) {
    const { type, data } = event.data;

    switch (type) {
      case "worker_ready":
        console.log("[WorkerManager] Worker ready signal received");
        if (this.workerReadyResolve) {
          this.workerReadyResolve();
          this.workerReadyResolve = null;
        }
        break;

      case "decrypt_tokens_request":
        // Worker needs tokens decrypted
        this.handleDecryptTokensRequest(data);
        break;

      case "password_request":
        this.handlePasswordRequest(data);
        break;

      case "request_storage_data":
        this.handleStorageDataRequest();
        break;

      case "storage_update":
        this.handleStorageUpdate(data);
        break;

      case "storage_remove":
        this.handleStorageRemove(data);
        break;

      case "token_refresh_success":
        console.log("[WorkerManager] Token refresh successful");
        // Handle manual refresh promise if exists
        if (
          data &&
          data.requestId &&
          this.refreshPromises.has(data.requestId)
        ) {
          const promise = this.refreshPromises.get(data.requestId);
          promise.resolve(data);
        }
        // Notify all listeners
        this.notifyAuthStateChange("token_refresh_success", data);
        break;

      case "token_refresh_failed":
        console.log("[WorkerManager] Token refresh failed:", data);
        // Handle manual refresh promise if exists
        if (
          data &&
          data.requestId &&
          this.refreshPromises.has(data.requestId)
        ) {
          const promise = this.refreshPromises.get(data.requestId);
          promise.reject(new Error(data.error || "Token refresh failed"));
        }
        // Notify all listeners
        this.notifyAuthStateChange("token_refresh_failed", data);
        break;

      case "force_logout":
        console.log("[WorkerManager] Force logout requested:", data);
        this.handleForceLogout(data);
        break;

      case "token_status_update":
        console.log("[WorkerManager] Token status update:", data);
        this.notifyAuthStateChange("token_status_update", data);
        break;

      case "worker_error":
        console.error("[WorkerManager] Worker error:", data);
        // Reject all pending refresh promises
        this.refreshPromises.forEach((promise) => {
          promise.reject(new Error(data.error || "Worker error"));
        });
        this.refreshPromises.clear();
        break;

      case "worker_status_response":
        // Handled by getWorkerStatus promise
        break;

      default:
        console.log("[WorkerManager] Unknown worker message:", type, data);
    }
  }

  handlePasswordRequest(data) {
    const password = passwordStorageService.getPassword();

    this.authWorker.postMessage({
      type: "password_response",
      requestId: data.requestId,
      password: password,
    });

    console.log("[WorkerManager] Password request handled:", {
      requestId: data.requestId,
      hasPassword: !!password,
    });
  }

  handleStorageDataRequest() {
    const storageData = this.getCurrentStorageData();
    this.authWorker.postMessage({
      type: "storage_data_response",
      data: storageData,
    });
    console.log("[WorkerManager] Storage data sent to worker");
  }

  handleStorageUpdate(data) {
    try {
      const { key, value } = data;
      if (key && value !== undefined) {
        localStorage.setItem(key, value);
        console.log(`[WorkerManager] Storage updated: ${key}`);
      }
    } catch (error) {
      console.error("[WorkerManager] Failed to update storage:", error);
    }
  }

  handleStorageRemove(data) {
    try {
      const { key } = data;
      if (key) {
        localStorage.removeItem(key);
        console.log(`[WorkerManager] Storage removed: ${key}`);
      }
    } catch (error) {
      console.error("[WorkerManager] Failed to remove storage:", error);
    }
  }

  handleForceLogout(data) {
    console.log("[WorkerManager] Handling force logout:", data);

    // Clear all authentication data
    LocalStorageService.clearAuthData();
    passwordStorageService.clearPassword();

    // Notify all listeners
    this.notifyAuthStateChange("force_logout", data);

    // Redirect to login if specified
    if (data.shouldRedirect && window.location.pathname !== "/") {
      setTimeout(() => {
        window.location.href = "/";
      }, 100);
    }
  }

  addAuthStateChangeListener(callback) {
    if (typeof callback === "function") {
      this.authStateListeners.add(callback);
      console.log(
        "[WorkerManager] Auth state listener added. Total listeners:",
        this.authStateListeners.size,
      );
    }
  }

  removeAuthStateChangeListener(callback) {
    this.authStateListeners.delete(callback);
    console.log(
      "[WorkerManager] Auth state listener removed. Total listeners:",
      this.authStateListeners.size,
    );
  }

  notifyAuthStateChange(eventType, eventData) {
    console.log(
      `[WorkerManager] Notifying ${this.authStateListeners.size} listeners of ${eventType}`,
    );

    this.authStateListeners.forEach((callback) => {
      try {
        callback(eventType, eventData);
      } catch (error) {
        console.error("[WorkerManager] Error in auth state listener:", error);
      }
    });
  }

  async getWorkerStatus() {
    if (!this.authWorker || !this.isInitialized) {
      return {
        isInitialized: false,
        error: "Worker not initialized",
        tokenSystem: "encrypted",
      };
    }

    return new Promise((resolve) => {
      const timeout = setTimeout(() => {
        resolve({ error: "Worker status request timed out" });
      }, 5000);

      const handleResponse = (event) => {
        if (event.data.type === "worker_status_response") {
          clearTimeout(timeout);
          resolve(event.data.data);
        }
      };

      // Use a one-time listener
      const originalHandler = this.authWorker.onmessage;
      this.authWorker.onmessage = (event) => {
        if (event.data.type === "worker_status_response") {
          handleResponse(event);
          this.authWorker.onmessage = originalHandler;
        } else {
          originalHandler(event);
        }
      };

      this.authWorker.postMessage({ type: "get_worker_status" });
    });
  }

  destroy() {
    if (this.authWorker) {
      this.stopMonitoring();
      this.authWorker.terminate();
      this.authWorker = null;
    }

    // Reject all pending refresh promises
    this.refreshPromises.forEach((promise) => {
      promise.reject(new Error("Worker manager destroyed"));
    });
    this.refreshPromises.clear();

    this.authStateListeners.clear();
    this.isInitialized = false;
    this.workerReadyPromise = null;
    this.workerReadyResolve = null;

    console.log("[WorkerManager] Destroyed");
  }
}

// Create singleton
const workerManager = new WorkerManager();
export default workerManager;
