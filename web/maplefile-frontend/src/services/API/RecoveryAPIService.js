// File: monorepo/web/maplefile-frontend/src/services/API/RecoveryAPIService.js
// Recovery API Service - Handles all API calls for account recovery

const API_BASE_URL = "/iam/api/v1"; // Using proxy from vite config

class RecoveryAPIService {
  constructor() {
    console.log("[RecoveryAPIService] API service initialized");
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
      console.log(
        `[RecoveryAPIService] Making ${requestOptions.method || "GET"} request to:`,
        url,
      );
      const response = await fetch(url, requestOptions);

      const data = await response.json();

      if (!response.ok) {
        console.error("[RecoveryAPIService] API Error:", data);
        throw new Error(
          data.details
            ? Object.values(data.details)[0]
            : data.error || "Request failed",
        );
      }

      console.log("[RecoveryAPIService] API Response:", data);
      return data;
    } catch (error) {
      console.error("[RecoveryAPIService] Request failed:", error);
      throw error;
    }
  }

  // Step 1: Initiate Account Recovery
  async initiateRecovery(email, method = "recovery_key") {
    try {
      console.log("[RecoveryAPIService] Initiating recovery for:", email);

      const response = await this.makeRequest("/recovery/initiate", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
          method: method,
        }),
      });

      console.log("[RecoveryAPIService] Recovery initiation successful");
      return response;
    } catch (error) {
      console.error("[RecoveryAPIService] Recovery initiation failed:", error);
      throw new Error(`Failed to initiate recovery: ${error.message}`);
    }
  }

  // Step 2: Verify Recovery Challenge
  async verifyRecovery(sessionId, decryptedChallenge) {
    try {
      console.log("[RecoveryAPIService] Verifying recovery challenge");

      const response = await this.makeRequest("/recovery/verify", {
        method: "POST",
        body: JSON.stringify({
          session_id: sessionId,
          decrypted_challenge: decryptedChallenge,
        }),
      });

      console.log("[RecoveryAPIService] Recovery verification successful");
      return response;
    } catch (error) {
      console.error(
        "[RecoveryAPIService] Recovery verification failed:",
        error,
      );
      throw new Error(`Failed to verify recovery: ${error.message}`);
    }
  }

  // Step 3: Complete Account Recovery
  async completeRecovery(recoveryData) {
    try {
      console.log("[RecoveryAPIService] Completing account recovery");

      const response = await this.makeRequest("/recovery/complete", {
        method: "POST",
        body: JSON.stringify(recoveryData),
      });

      console.log("[RecoveryAPIService] Recovery completion successful");
      return response;
    } catch (error) {
      console.error("[RecoveryAPIService] Recovery completion failed:", error);
      throw new Error(`Failed to complete recovery: ${error.message}`);
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "RecoveryAPIService",
      baseUrl: API_BASE_URL,
      availableEndpoints: [
        "POST /recovery/initiate",
        "POST /recovery/verify",
        "POST /recovery/complete",
      ],
    };
  }
}

export default RecoveryAPIService;
