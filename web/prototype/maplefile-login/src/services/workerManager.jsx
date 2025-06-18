// Worker Manager - Handles communication with the authentication worker
import LocalStorageService from "./localStorageService.jsx";

class WorkerManager {
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
      this.broadcastChannel = new BroadcastChannel("auth_worker");
      this.broadcastChannel.addEventListener("message", (event) => {
        this.handleWorkerMessage(event.data);
      });
      console.log("[WorkerManager] BroadcastChannel initialized");
    } catch (error) {
      console.warn("[WorkerManager] BroadcastChannel not supported:", error);
    }
  }

  // Initialize the worker
  async initialize() {
    if (this.isInitialized) {
      return;
    }

    try {
      // Create the worker
      this.worker = new Worker("/auth-worker.js");

      // Set up worker message handling
      this.worker.addEventListener("message", (event) => {
        this.handleWorkerMessage(event.data);
      });

      this.worker.addEventListener("error", (error) => {
        console.error("[WorkerManager] Worker error:", error);
        this.handleWorkerError(error);
      });

      this.isInitialized = true;
      console.log("[WorkerManager] Worker initialized successfully");

      // Wait for worker ready signal
      await this.waitForWorkerReady();

      // Start monitoring if user is authenticated
      if (LocalStorageService.isAuthenticated()) {
        this.startMonitoring();
      }
    } catch (error) {
      console.error("[WorkerManager] Failed to initialize worker:", error);
      throw error;
    }
  }

  // Wait for worker ready signal
  waitForWorkerReady(timeout = 5000) {
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        reject(new Error("Worker ready timeout"));
      }, timeout);

      const handleReady = (data) => {
        if (data.type === "worker_ready") {
          clearTimeout(timer);
          this.removeMessageHandler("worker_ready", handleReady);
          console.log("[WorkerManager] Worker is ready");
          resolve(data.data);
        }
      };

      this.addMessageHandler("worker_ready", handleReady);
    });
  }

  // Handle messages from worker
  handleWorkerMessage(message) {
    const { type, data, timestamp } = message;

    console.log(`[WorkerManager] Received message: ${type}`, data);

    // Call registered handlers
    const handlers = this.messageHandlers.get(type) || [];
    handlers.forEach((handler) => {
      try {
        handler(data, message);
      } catch (error) {
        console.error(`[WorkerManager] Error in ${type} handler:`, error);
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
        console.log("[WorkerManager] Token refresh successful");
        this.notifyAuthStateChange("token_refreshed", data);
        break;

      case "token_refresh_failed":
        console.error("[WorkerManager] Token refresh failed");
        this.notifyAuthStateChange("token_refresh_failed", data);
        if (data.shouldRedirect) {
          this.redirectToLogin();
        }
        break;

      case "force_logout":
        console.log("[WorkerManager] Force logout received");
        this.handleForceLogout(data);
        break;

      case "token_status_update":
        this.notifyAuthStateChange("token_status_update", data);
        break;

      case "worker_error":
        console.error("[WorkerManager] Worker reported error:", data);
        break;

      default:
        // Custom message types handled by registered handlers
        break;
    }
  }

  // Send current storage data to worker
  sendStorageDataToWorker() {
    const storageData = {
      [LocalStorageService.constructor.STORAGE_KEYS?.ACCESS_TOKEN ||
      "mapleapps_access_token"]: localStorage.getItem("mapleapps_access_token"),
      [LocalStorageService.constructor.STORAGE_KEYS?.REFRESH_TOKEN ||
      "mapleapps_refresh_token"]: localStorage.getItem(
        "mapleapps_refresh_token",
      ),
      [LocalStorageService.constructor.STORAGE_KEYS?.ACCESS_TOKEN_EXPIRY ||
      "mapleapps_access_token_expiry"]: localStorage.getItem(
        "mapleapps_access_token_expiry",
      ),
      [LocalStorageService.constructor.STORAGE_KEYS?.REFRESH_TOKEN_EXPIRY ||
      "mapleapps_refresh_token_expiry"]: localStorage.getItem(
        "mapleapps_refresh_token_expiry",
      ),
      [LocalStorageService.constructor.STORAGE_KEYS?.USER_EMAIL ||
      "mapleapps_user_email"]: localStorage.getItem("mapleapps_user_email"),
    };

    this.sendToWorker("storage_data_response", storageData);
  }

  // Handle force logout
  handleForceLogout(data) {
    console.log("[WorkerManager] Handling force logout:", data.reason);

    // Clear local storage
    LocalStorageService.clearAuthData();

    // Notify listeners
    this.notifyAuthStateChange("force_logout", data);

    // Redirect to login if required
    if (data.shouldRedirect) {
      this.redirectToLogin();
    }
  }

  // Redirect to login page
  redirectToLogin() {
    console.log("[WorkerManager] Redirecting to login page");

    // Use setTimeout to ensure current execution completes
    setTimeout(() => {
      if (window.location.pathname !== "/") {
        window.location.href = "/";
      }
    }, 100);
  }

  // Send message to worker
  sendToWorker(type, data = {}) {
    if (!this.worker) {
      console.warn("[WorkerManager] Worker not initialized");
      return;
    }

    const message = {
      type,
      data,
      timestamp: Date.now(),
    };

    try {
      this.worker.postMessage(message);
      console.log(`[WorkerManager] Sent message to worker: ${type}`);
    } catch (error) {
      console.error("[WorkerManager] Failed to send message to worker:", error);
    }
  }

  // Start token monitoring
  startMonitoring() {
    if (!this.isInitialized) {
      console.warn(
        "[WorkerManager] Worker not initialized, cannot start monitoring",
      );
      return;
    }

    console.log("[WorkerManager] Starting token monitoring");
    this.sendToWorker("start_monitoring");
  }

  // Stop token monitoring
  stopMonitoring() {
    if (!this.isInitialized) {
      return;
    }

    console.log("[WorkerManager] Stopping token monitoring");
    this.sendToWorker("stop_monitoring");
  }

  // Force token check
  forceTokenCheck() {
    if (!this.isInitialized) {
      return;
    }

    console.log("[WorkerManager] Forcing token check");
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

      // Send refresh request
      this.sendToWorker("manual_refresh", {
        refreshToken: LocalStorageService.getRefreshToken(),
        storageData: this.getCurrentStorageData(),
      });
    });
  }

  // Get current storage data
  getCurrentStorageData() {
    return {
      mapleapps_access_token: localStorage.getItem("mapleapps_access_token"),
      mapleapps_refresh_token: localStorage.getItem("mapleapps_refresh_token"),
      mapleapps_access_token_expiry: localStorage.getItem(
        "mapleapps_access_token_expiry",
      ),
      mapleapps_refresh_token_expiry: localStorage.getItem(
        "mapleapps_refresh_token_expiry",
      ),
      mapleapps_user_email: localStorage.getItem("mapleapps_user_email"),
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

  // Add auth state change listener
  addAuthStateChangeListener(listener) {
    this.eventListeners.push(listener);
  }

  // Remove auth state change listener
  removeAuthStateChangeListener(listener) {
    const index = this.eventListeners.indexOf(listener);
    if (index > -1) {
      this.eventListeners.splice(index, 1);
    }
  }

  // Notify auth state change
  notifyAuthStateChange(type, data) {
    this.eventListeners.forEach((listener) => {
      try {
        listener(type, data);
      } catch (error) {
        console.error("[WorkerManager] Error in auth state listener:", error);
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
    console.error("[WorkerManager] Worker error occurred:", error);
    this.notifyAuthStateChange("worker_error", { error: error.message });
  }

  // Cleanup
  destroy() {
    console.log("[WorkerManager] Destroying worker manager");

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
export default new WorkerManager();
