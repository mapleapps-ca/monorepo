// Token Refresh Worker Manager - Simplified for maplefile-refreshtoken prototype

class TokenRefreshWorkerManager {
  constructor() {
    this.worker = null;
    this.broadcastChannel = null;
    this.isInitialized = false;
    this.messageHandlers = new Map();
    this.eventListeners = [];

    // Initialize BroadcastChannel for cross-tab communication
    this.initBroadcastChannel();
  }

  // Initialize BroadcastChannel
  initBroadcastChannel() {
    try {
      this.broadcastChannel = new BroadcastChannel("token_refresh_worker");
      this.broadcastChannel.addEventListener("message", (event) => {
        this.handleWorkerMessage(event.data);
      });
      console.log("[TokenRefreshWorkerManager] BroadcastChannel initialized");
    } catch (error) {
      console.warn(
        "[TokenRefreshWorkerManager] BroadcastChannel not supported:",
        error,
      );
    }
  }

  // Initialize the worker
  async initialize() {
    if (this.isInitialized) {
      return;
    }

    try {
      console.log(
        "[TokenRefreshWorkerManager] Starting worker initialization...",
      );

      // Create the worker
      this.worker = new Worker("/token-refresh-worker.js");
      console.log("[TokenRefreshWorkerManager] Worker created successfully");

      // Set up worker message handling
      this.worker.addEventListener("message", (event) => {
        console.log(
          "[TokenRefreshWorkerManager] Received message from worker:",
          event.data.type,
        );
        this.handleWorkerMessage(event.data);
      });

      this.worker.addEventListener("error", (error) => {
        console.error("[TokenRefreshWorkerManager] Worker error:", error);
        this.handleWorkerError(error);
      });

      console.log(
        "[TokenRefreshWorkerManager] Worker event listeners attached",
      );

      // Wait for worker ready signal
      console.log(
        "[TokenRefreshWorkerManager] Waiting for worker ready signal...",
      );
      await this.waitForWorkerReady();

      this.isInitialized = true;
      console.log(
        "[TokenRefreshWorkerManager] Worker initialized successfully",
      );
    } catch (error) {
      console.error(
        "[TokenRefreshWorkerManager] Failed to initialize worker:",
        error,
      );
      throw error;
    }
  }

  // Wait for worker ready signal
  waitForWorkerReady(timeout = 10000) {
    return new Promise((resolve, reject) => {
      console.log(
        `[TokenRefreshWorkerManager] Setting up worker ready timeout: ${timeout}ms`,
      );

      const timer = setTimeout(() => {
        this.removeMessageHandler("worker_ready", handleReady);
        console.error(
          "[TokenRefreshWorkerManager] Worker ready timeout reached",
        );
        reject(new Error("Worker ready timeout"));
      }, timeout);

      const handleReady = (data, message) => {
        console.log(
          "[TokenRefreshWorkerManager] Received worker ready signal:",
          data,
        );
        clearTimeout(timer);
        this.removeMessageHandler("worker_ready", handleReady);
        resolve(data);
      };

      this.addMessageHandler("worker_ready", handleReady);
    });
  }

  // Handle messages from worker
  handleWorkerMessage(message) {
    const { type, data, timestamp } = message;

    console.log(`[TokenRefreshWorkerManager] Received message: ${type}`, data);

    // Call registered handlers
    const handlers = this.messageHandlers.get(type) || [];
    handlers.forEach((handler) => {
      try {
        handler(data, message);
      } catch (error) {
        console.error(
          `[TokenRefreshWorkerManager] Error in ${type} handler:`,
          error,
        );
      }
    });

    // Handle specific message types
    switch (type) {
      case "request_storage_data":
        this.sendStorageDataToWorker();
        break;

      case "storage_update":
        if (data.key && data.value !== undefined) {
          localStorage.setItem(data.key, data.value);
          this.notifyStorageChange(data.key, data.value);
        }
        break;

      case "storage_remove":
        if (data.key) {
          localStorage.removeItem(data.key);
          this.notifyStorageChange(data.key, null);
        }
        break;

      case "token_refresh_success":
        console.log("[TokenRefreshWorkerManager] Token refresh successful");
        this.notifyStateChange("token_refresh_success", data);
        break;

      case "token_refresh_failed":
        console.error("[TokenRefreshWorkerManager] Token refresh failed");
        this.notifyStateChange("token_refresh_failed", data);
        break;

      case "force_logout":
        console.log("[TokenRefreshWorkerManager] Force logout received");
        this.handleForceLogout(data);
        break;

      case "token_status_update":
        this.notifyStateChange("token_status_update", data);
        break;

      case "worker_error":
        console.error(
          "[TokenRefreshWorkerManager] Worker reported error:",
          data,
        );
        break;

      default:
        // Custom message types handled by registered handlers
        break;
    }
  }

  // Send current storage data to worker
  sendStorageDataToWorker() {
    const storageData = {
      // Encrypted token keys
      mapleapps_encrypted_access_token: localStorage.getItem(
        "mapleapps_encrypted_access_token",
      ),
      mapleapps_encrypted_refresh_token: localStorage.getItem(
        "mapleapps_encrypted_refresh_token",
      ),
      mapleapps_token_nonce: localStorage.getItem("mapleapps_token_nonce"),
      mapleapps_access_token_expiry: localStorage.getItem(
        "mapleapps_access_token_expiry",
      ),
      mapleapps_refresh_token_expiry: localStorage.getItem(
        "mapleapps_refresh_token_expiry",
      ),
      mapleapps_user_email: localStorage.getItem("mapleapps_user_email"),

      // Legacy single encrypted tokens field
      mapleapps_encrypted_tokens: localStorage.getItem(
        "mapleapps_encrypted_tokens",
      ),
    };

    console.log("[TokenRefreshWorkerManager] Sending storage data to worker:", {
      hasEncryptedTokens:
        !!(
          (storageData.mapleapps_encrypted_access_token &&
            storageData.mapleapps_encrypted_refresh_token) ||
          storageData.mapleapps_encrypted_tokens
        ) && storageData.mapleapps_token_nonce,
      userEmail: storageData.mapleapps_user_email,
    });

    this.sendToWorker("storage_data_response", storageData);
  }

  // Handle force logout
  handleForceLogout(data) {
    console.log(
      "[TokenRefreshWorkerManager] Handling force logout:",
      data.reason,
    );

    // Clear localStorage
    this.clearAllTokens();

    // Notify listeners
    this.notifyStateChange("force_logout", data);
  }

  // Clear all tokens
  clearAllTokens() {
    const tokenKeys = [
      "mapleapps_encrypted_access_token",
      "mapleapps_encrypted_refresh_token",
      "mapleapps_token_nonce",
      "mapleapps_access_token_expiry",
      "mapleapps_refresh_token_expiry",
      "mapleapps_user_email",
      "mapleapps_encrypted_tokens",
    ];

    tokenKeys.forEach((key) => {
      localStorage.removeItem(key);
    });

    console.log(
      "[TokenRefreshWorkerManager] All tokens cleared from localStorage",
    );
  }

  // Send message to worker
  sendToWorker(type, data = {}) {
    if (!this.worker) {
      console.warn("[TokenRefreshWorkerManager] Worker not initialized");
      return;
    }

    const message = {
      type,
      data,
      timestamp: Date.now(),
    };

    try {
      this.worker.postMessage(message);
      console.log(
        `[TokenRefreshWorkerManager] Sent message to worker: ${type}`,
      );
    } catch (error) {
      console.error(
        "[TokenRefreshWorkerManager] Failed to send message to worker:",
        error,
      );
    }
  }

  // Start token monitoring
  startMonitoring() {
    if (!this.isInitialized) {
      console.warn(
        "[TokenRefreshWorkerManager] Worker not initialized, cannot start monitoring",
      );
      return;
    }

    console.log("[TokenRefreshWorkerManager] Starting token monitoring");
    this.sendToWorker("start_monitoring");
  }

  // Stop token monitoring
  stopMonitoring() {
    if (!this.isInitialized) {
      return;
    }

    console.log("[TokenRefreshWorkerManager] Stopping token monitoring");
    this.sendToWorker("stop_monitoring");
  }

  // Force token check
  forceTokenCheck() {
    if (!this.isInitialized) {
      return;
    }

    console.log("[TokenRefreshWorkerManager] Forcing token check");
    this.sendStorageDataToWorker();
    this.sendToWorker("force_token_check");
  }

  // Manual token refresh
  async manualRefresh() {
    if (!this.isInitialized) {
      throw new Error("Worker not initialized");
    }

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.removeMessageHandler("token_refresh_success", successHandler);
        this.removeMessageHandler("token_refresh_failed", failHandler);
        reject(new Error("Manual refresh timeout"));
      }, 30000); // 30 second timeout

      const successHandler = (data) => {
        clearTimeout(timeout);
        this.removeMessageHandler("token_refresh_success", successHandler);
        this.removeMessageHandler("token_refresh_failed", failHandler);
        resolve(data);
      };

      const failHandler = (data) => {
        clearTimeout(timeout);
        this.removeMessageHandler("token_refresh_success", successHandler);
        this.removeMessageHandler("token_refresh_failed", failHandler);
        reject(new Error(data.error || "Token refresh failed"));
      };

      this.addMessageHandler("token_refresh_success", successHandler);
      this.addMessageHandler("token_refresh_failed", failHandler);

      // Get refresh token from storage
      const refreshToken =
        localStorage.getItem("mapleapps_encrypted_refresh_token") ||
        localStorage.getItem("mapleapps_encrypted_tokens");

      this.sendToWorker("manual_refresh", {
        refreshToken: refreshToken,
        storageData: this.getCurrentStorageData(),
      });
    });
  }

  // Get current storage data
  getCurrentStorageData() {
    return {
      mapleapps_encrypted_access_token: localStorage.getItem(
        "mapleapps_encrypted_access_token",
      ),
      mapleapps_encrypted_refresh_token: localStorage.getItem(
        "mapleapps_encrypted_refresh_token",
      ),
      mapleapps_token_nonce: localStorage.getItem("mapleapps_token_nonce"),
      mapleapps_access_token_expiry: localStorage.getItem(
        "mapleapps_access_token_expiry",
      ),
      mapleapps_refresh_token_expiry: localStorage.getItem(
        "mapleapps_refresh_token_expiry",
      ),
      mapleapps_user_email: localStorage.getItem("mapleapps_user_email"),
      mapleapps_encrypted_tokens: localStorage.getItem(
        "mapleapps_encrypted_tokens",
      ),
    };
  }

  // Get worker status
  async getWorkerStatus() {
    if (!this.isInitialized) {
      return { isInitialized: false };
    }

    return new Promise((resolve) => {
      const timeout = setTimeout(() => {
        resolve({ error: "Status request timeout" });
      }, 5000);

      const handler = (data) => {
        clearTimeout(timeout);
        this.removeMessageHandler("worker_status_response", handler);
        resolve(data);
      };

      this.addMessageHandler("worker_status_response", handler);
      this.sendToWorker("get_worker_status");
    });
  }

  // Add message handler
  addMessageHandler(type, handler) {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, []);
    }
    this.messageHandlers.get(type).push(handler);
  }

  // Remove message handler
  removeMessageHandler(type, handler) {
    const handlers = this.messageHandlers.get(type);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  // Add state change listener
  addStateChangeListener(listener) {
    this.eventListeners.push(listener);
  }

  // Remove state change listener
  removeStateChangeListener(listener) {
    const index = this.eventListeners.indexOf(listener);
    if (index > -1) {
      this.eventListeners.splice(index, 1);
    }
  }

  // Notify state change
  notifyStateChange(type, data) {
    this.eventListeners.forEach((listener) => {
      try {
        listener(type, data);
      } catch (error) {
        console.error(
          "[TokenRefreshWorkerManager] Error in state listener:",
          error,
        );
      }
    });
  }

  // Notify storage change
  notifyStorageChange(key, value) {
    const event = new CustomEvent("storage", {
      detail: { key, newValue: value },
    });
    window.dispatchEvent(event);
  }

  // Handle worker errors
  handleWorkerError(error) {
    console.error("[TokenRefreshWorkerManager] Worker error occurred:", error);
    this.notifyStateChange("worker_error", { error: error.message });
  }

  // Cleanup
  destroy() {
    console.log("[TokenRefreshWorkerManager] Destroying worker manager");

    this.stopMonitoring();

    if (this.worker) {
      this.worker.terminate();
      this.worker = null;
    }

    if (this.broadcastChannel) {
      this.broadcastChannel.close();
      this.broadcastChannel = null;
    }

    this.messageHandlers.clear();
    this.eventListeners = [];
    this.isInitialized = false;
  }
}

// Export singleton instance
export default new TokenRefreshWorkerManager();
