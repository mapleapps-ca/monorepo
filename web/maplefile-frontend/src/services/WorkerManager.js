// File: monorepo/web/maplefile-frontend/src/services/WorkerManager.js
// Simplified WorkerManager without web workers - just event management
import LocalStorageService from "./LocalStorageService.js";

class WorkerManager {
  constructor() {
    this.isInitialized = false;
    this.authStateListeners = new Set();
  }

  async initialize() {
    if (this.isInitialized) return;

    try {
      console.log(
        "[WorkerManager] Initializing simplified worker manager (no web workers)...",
      );
      this.isInitialized = true;
      console.log(
        "[WorkerManager] Initialized successfully without web workers",
      );
    } catch (error) {
      console.error("[WorkerManager] Failed to initialize:", error);
      this.isInitialized = false;
      throw error;
    }
  }

  // These methods are now no-ops since we don't have workers
  startMonitoring() {
    console.log(
      "[WorkerManager] Token monitoring is now handled by ApiClient interceptors",
    );
  }

  stopMonitoring() {
    console.log("[WorkerManager] Token monitoring stopped");
  }

  forceTokenCheck() {
    console.log("[WorkerManager] Force token check - handled by ApiClient now");
  }

  // Manual refresh is now handled by ApiClient directly
  async manualRefresh() {
    console.log("[WorkerManager] Manual refresh delegated to ApiClient");
    throw new Error("Manual refresh should be called through ApiClient now");
  }

  // Event listener management remains the same
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
    return {
      isInitialized: this.isInitialized,
      tokenSystem: "unencrypted",
      method: "api_interceptor",
      hasWorker: false,
    };
  }

  destroy() {
    this.authStateListeners.clear();
    this.isInitialized = false;
    console.log("[WorkerManager] Destroyed (simplified)");
  }
}

// Create singleton
const workerManager = new WorkerManager();
export default workerManager;
