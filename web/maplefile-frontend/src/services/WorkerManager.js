// File: src/services/WorkerManager.js - ENHANCED VERSION
import LocalStorageService from "./LocalStorageService.js";
import passwordStorageService from "./PasswordStorageService.js";

class WorkerManager {
  constructor() {
    this.authWorker = null;
    this.isInitialized = false;
  }

  async initialize() {
    if (this.isInitialized) return;

    try {
      // Initialize the auth worker
      this.authWorker = new Worker("/auth-worker.js");

      // Set up message handling
      this.authWorker.onmessage = (event) => {
        this.handleWorkerMessage(event);
      };

      // Set up error handling
      this.authWorker.onerror = (error) => {
        console.error("[WorkerManager] Auth worker error:", error);
      };

      this.isInitialized = true;
      console.log("[WorkerManager] Initialized successfully");

      // Start monitoring (the worker will auto-start, but this ensures it's active)
      this.startMonitoring();
    } catch (error) {
      console.error("[WorkerManager] Failed to initialize:", error);
      throw error;
    }
  }

  startMonitoring() {
    console.log("[WorkerManager] Starting token monitoring");
    if (this.authWorker) {
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

  forceTokenCheck(storageData = null) {
    console.log("[WorkerManager] Forcing token check");
    if (this.authWorker) {
      this.authWorker.postMessage({
        type: "force_token_check",
        data: storageData || this.getCurrentStorageData(),
      });
    }
  }

  manualRefresh(refreshToken, storageData = null) {
    console.log("[WorkerManager] Manual token refresh");
    if (this.authWorker) {
      this.authWorker.postMessage({
        type: "manual_refresh",
        data: {
          refreshToken,
          storageData: storageData || this.getCurrentStorageData(),
        },
      });
    }
  }

  getCurrentStorageData() {
    try {
      return {
        mapleapps_access_token:
          LocalStorageService.getAccessToken?.() ||
          localStorage.getItem("mapleapps_access_token"),
        mapleapps_refresh_token:
          LocalStorageService.getRefreshToken?.() ||
          localStorage.getItem("mapleapps_refresh_token"),
        mapleapps_access_token_expiry:
          LocalStorageService.getAccessTokenExpiry?.() ||
          localStorage.getItem("mapleapps_access_token_expiry"),
        mapleapps_refresh_token_expiry:
          LocalStorageService.getRefreshTokenExpiry?.() ||
          localStorage.getItem("mapleapps_refresh_token_expiry"),
        mapleapps_user_email:
          LocalStorageService.getUserEmail?.() ||
          localStorage.getItem("mapleapps_user_email"),
      };
    } catch (error) {
      console.error("[WorkerManager] Failed to get storage data:", error);
      return {};
    }
  }

  handleWorkerMessage(event) {
    const { type, data } = event.data;

    switch (type) {
      case "password_request":
        // Worker is requesting password for token refresh
        this.handlePasswordRequest(data);
        break;

      case "request_storage_data":
        // Worker is requesting current storage data
        this.handleStorageDataRequest();
        break;

      case "storage_update":
        // Worker wants to update storage
        this.handleStorageUpdate(data);
        break;

      case "storage_remove":
        // Worker wants to remove storage item
        this.handleStorageRemove(data);
        break;

      case "token_refresh_success":
        console.log("[WorkerManager] Token refresh successful");
        // Notify auth state change listeners
        this.notifyAuthStateChange({
          isAuthenticated: true,
          userEmail: localStorage.getItem("mapleapps_user_email") || null,
        });
        break;

      case "token_refresh_failed":
        console.log("[WorkerManager] Token refresh failed:", data);
        // Notify auth state change listeners if this affects auth status
        if (data.shouldRedirect) {
          this.notifyAuthStateChange({
            isAuthenticated: false,
            userEmail: null,
          });
        }
        break;

      case "force_logout":
        console.log("[WorkerManager] Force logout requested:", data);
        this.handleForceLogout(data);
        // Notify auth state change listeners
        this.notifyAuthStateChange({
          isAuthenticated: false,
          userEmail: null,
        });
        break;

      case "worker_ready":
        console.log("[WorkerManager] Worker ready:", data);
        break;

      case "token_status_update":
        console.log("[WorkerManager] Token status update:", data);
        break;

      default:
        console.log("[WorkerManager] Worker message:", type, data);
    }
  }

  handlePasswordRequest(data) {
    // Get password from password service
    const password = passwordStorageService.getPassword();

    // Send response back to worker
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

  // ADD THIS MISSING METHOD
  handleStorageDataRequest() {
    // Send current storage data to worker
    const storageData = this.getCurrentStorageData();
    this.authWorker.postMessage({
      type: "storage_data_response",
      data: storageData,
    });
    console.log("[WorkerManager] Storage data sent to worker");
  }

  // ADD THIS MISSING METHOD
  handleStorageUpdate(data) {
    try {
      const { key, value } = data;
      localStorage.setItem(key, value);
      console.log(`[WorkerManager] Storage updated: ${key}`);
    } catch (error) {
      console.error("[WorkerManager] Failed to update storage:", error);
    }
  }

  // ADD THIS MISSING METHOD
  handleStorageRemove(data) {
    try {
      const { key } = data;
      localStorage.removeItem(key);
      console.log(`[WorkerManager] Storage removed: ${key}`);
    } catch (error) {
      console.error("[WorkerManager] Failed to remove storage:", error);
    }
  }

  // ADD THIS MISSING METHOD
  handleForceLogout(data) {
    console.log("[WorkerManager] Handling force logout:", data);

    // Clear all local storage
    const keysToRemove = [
      "mapleapps_access_token",
      "mapleapps_refresh_token",
      "mapleapps_access_token_expiry",
      "mapleapps_refresh_token_expiry",
      "mapleapps_user_email",
    ];

    keysToRemove.forEach((key) => {
      localStorage.removeItem(key);
    });

    // Clear password service
    passwordStorageService.clearPassword();

    // Redirect to login if specified
    if (data.shouldRedirect) {
      window.location.href = "/login";
    }
  }

  // Send message to worker
  sendMessage(type, data) {
    if (this.authWorker) {
      this.authWorker.postMessage({ type, data });
    } else {
      console.warn(
        "[WorkerManager] Cannot send message - worker not initialized",
      );
    }
  }

  // ADD THIS UTILITY METHOD
  getWorkerStatus() {
    return new Promise((resolve) => {
      if (!this.authWorker) {
        resolve({ error: "Worker not initialized" });
        return;
      }

      const timeout = setTimeout(() => {
        resolve({ error: "Worker status request timed out" });
      }, 5000);

      const handleResponse = (event) => {
        if (event.data.type === "worker_status_response") {
          this.authWorker.removeEventListener("message", handleResponse);
          clearTimeout(timeout);
          resolve(event.data.data);
        }
      };

      this.authWorker.addEventListener("message", handleResponse);
      this.authWorker.postMessage({ type: "get_worker_status" });
    });
  }

  // ADD THESE MISSING AUTH STATE LISTENER METHODS
  addAuthStateChangeListener(callback) {
    console.log("[WorkerManager] Adding auth state change listener");
    // For now, just store the callback - you can enhance this later
    this.authStateCallback = callback;

    // Immediately call with current state if we have it
    if (this.authStateCallback) {
      // You can get current auth state from your auth worker or localStorage
      const currentState = {
        isAuthenticated: !!localStorage.getItem("mapleapps_access_token"),
        userEmail: localStorage.getItem("mapleapps_user_email") || null,
      };
      this.authStateCallback(currentState);
    }
  }

  // ADD THIS MISSING METHOD
  removeAuthStateChangeListener() {
    console.log("[WorkerManager] Removing auth state change listener");
    this.authStateCallback = null;
  }

  // ADD THIS HELPER METHOD TO NOTIFY AUTH STATE CHANGES
  notifyAuthStateChange(authState) {
    if (this.authStateCallback) {
      console.log("[WorkerManager] Notifying auth state change:", authState);
      this.authStateCallback(authState);
    }
  }

  destroy() {
    if (this.authWorker) {
      this.stopMonitoring();
      this.authWorker.terminate();
      this.authWorker = null;
    }
    this.authStateCallback = null;
    this.isInitialized = false;
    console.log("[WorkerManager] Destroyed");
  }
}

// Create singleton
const workerManager = new WorkerManager();
export default workerManager;
