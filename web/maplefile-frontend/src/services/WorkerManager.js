// Worker Manager - Updated for Unencrypted Token System
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

      // Start monitoring
      this.authWorker.postMessage({ type: "start_monitoring" });
    } catch (error) {
      console.error("[WorkerManager] Failed to initialize:", error);
      throw error;
    }
  }

  handleWorkerMessage(event) {
    const { type, data } = event.data;

    switch (type) {
      case "password_request":
        // Worker is requesting password for token refresh
        this.handlePasswordRequest(data);
        break;

      case "token_refresh_success":
        console.log("[WorkerManager] Token refresh successful");
        break;

      case "token_refresh_failed":
        console.log("[WorkerManager] Token refresh failed:", data);
        break;

      case "force_logout":
        console.log("[WorkerManager] Force logout requested:", data);
        // Handle logout logic here
        break;

      default:
        // Handle other worker messages
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

  // Send message to worker
  sendMessage(type, data) {
    if (this.authWorker) {
      this.authWorker.postMessage({ type, data });
    }
  }

  destroy() {
    if (this.authWorker) {
      this.authWorker.terminate();
      this.authWorker = null;
    }
    this.isInitialized = false;
  }
}

// Create singleton
const workerManager = new WorkerManager();
export default workerManager;
