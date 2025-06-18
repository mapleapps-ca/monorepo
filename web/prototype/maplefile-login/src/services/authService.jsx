// Authentication Service for API calls
import LocalStorageService from "./localStorageService.jsx";
import workerManager from "./workerManager.jsx";

const API_BASE_URL = "/iam/api/v1"; // Using proxy from vite config

class AuthService {
  constructor() {
    this.isInitialized = false;
    this.initializeWorker();
  }

  // Initialize the background worker
  async initializeWorker() {
    if (this.isInitialized) return;

    try {
      await workerManager.initialize();
      this.isInitialized = true;
      console.log("[AuthService] Background worker initialized");

      // Start monitoring if authenticated
      if (this.isAuthenticated()) {
        workerManager.startMonitoring();
      }
    } catch (error) {
      console.error("[AuthService] Failed to initialize worker:", error);
      // Fallback to manual token management if worker fails
    }
  }
  // Helper method to make API requests
  async makeRequest(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;

    const defaultOptions = {
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    };

    const requestOptions = {
      ...defaultOptions,
      ...options,
    };

    try {
      console.log(`Making ${requestOptions.method || "GET"} request to:`, url);
      const response = await fetch(url, requestOptions);

      const data = await response.json();

      if (!response.ok) {
        console.error("API Error:", data);
        throw new Error(
          data.details
            ? Object.values(data.details)[0]
            : data.error || "Request failed",
        );
      }

      console.log("API Response:", data);
      return data;
    } catch (error) {
      console.error("Request failed:", error);
      throw error;
    }
  }

  // Step 1: Request One-Time Token (OTT)
  async requestOTT(email) {
    try {
      const response = await this.makeRequest("/request-ott", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
        }),
      });

      // Store email for next steps
      LocalStorageService.setUserEmail(email);

      return response;
    } catch (error) {
      throw new Error(`Failed to request OTT: ${error.message}`);
    }
  }

  // Step 2: Verify One-Time Token
  async verifyOTT(email, ott) {
    try {
      const response = await this.makeRequest("/verify-ott", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
          ott: ott.trim(),
        }),
      });

      // Store verification data for the final step
      LocalStorageService.setLoginSessionData("verify_response", response);

      return response;
    } catch (error) {
      throw new Error(`Failed to verify OTT: ${error.message}`);
    }
  }

  // Step 3: Complete Login
  async completeLogin(email, challengeId, decryptedData) {
    try {
      const response = await this.makeRequest("/complete-login", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
          challengeId: challengeId,
          decryptedData: decryptedData,
        }),
      });

      // Store tokens in local storage
      if (response.access_token) {
        LocalStorageService.setAccessToken(
          response.access_token,
          response.access_token_expiry_time,
        );
      }

      if (response.refresh_token) {
        LocalStorageService.setRefreshToken(
          response.refresh_token,
          response.refresh_token_expiry_time,
        );
      }

      // Handle encrypted tokens if provided
      if (response.encrypted_tokens) {
        console.log("Received encrypted tokens:", response.encrypted_tokens);
        // Note: In a real implementation, you would decrypt these tokens client-side
        // For this demo, we'll just log them
      }

      // Clear login session data
      LocalStorageService.clearAllLoginSessionData();

      // Start background monitoring after successful login
      if (this.isInitialized) {
        workerManager.startMonitoring();
      }

      return response;
    } catch (error) {
      throw new Error(`Failed to complete login: ${error.message}`);
    }
  }

  // Refresh tokens using background worker
  async refreshToken() {
    if (!this.isInitialized) {
      await this.initializeWorker();
    }

    try {
      console.log("[AuthService] Requesting manual token refresh via worker");
      const result = await workerManager.manualRefresh();
      console.log("[AuthService] Manual token refresh successful");
      return result;
    } catch (error) {
      console.error("[AuthService] Manual token refresh failed:", error);
      throw new Error(`Failed to refresh tokens: ${error.message}`);
    }
  }

  // Force a token check (useful for testing)
  forceTokenCheck() {
    if (this.isInitialized) {
      workerManager.forceTokenCheck();
    }
  }

  // Get worker status for debugging
  async getWorkerStatus() {
    if (!this.isInitialized) {
      return { isInitialized: false };
    }
    return await workerManager.getWorkerStatus();
  }

  // Logout and stop background monitoring
  logout() {
    console.log("[AuthService] Logging out user");

    // Stop background monitoring
    if (this.isInitialized) {
      workerManager.stopMonitoring();
    }

    // Clear all authentication data
    LocalStorageService.clearAuthData();
    LocalStorageService.clearAllLoginSessionData();
  }

  // Check if user is authenticated
  isAuthenticated() {
    return LocalStorageService.isAuthenticated();
  }

  // Get current user email
  getCurrentUserEmail() {
    return LocalStorageService.getUserEmail();
  }

  // Simple challenge decryption simulation
  // Note: In a real implementation, this would involve proper cryptographic operations
  // with libsodium or similar libraries for ChaCha20-Poly1305 decryption
  simulateDecryption(encryptedChallenge) {
    console.log("Simulating challenge decryption...");
    console.log("Encrypted challenge:", encryptedChallenge);

    // For demo purposes, we'll simulate the decryption process
    // In reality, this would involve:
    // 1. Deriving key from password using Argon2ID
    // 2. Decrypting master key with derived key
    // 3. Decrypting private key with master key
    // 4. Decrypting challenge with private key
    // 5. Return base64 encoded decrypted challenge

    // Simulate some processing time
    return new Promise((resolve) => {
      setTimeout(() => {
        // Return a simulated decrypted challenge
        // This is just for demo - replace with actual cryptographic implementation
        const simulatedDecrypted = btoa(
          "simulated_decrypted_challenge_" + Date.now(),
        );
        console.log("Simulated decrypted challenge:", simulatedDecrypted);
        resolve(simulatedDecrypted);
      }, 1000);
    });
  }
}

// Export singleton instance
export default new AuthService();
