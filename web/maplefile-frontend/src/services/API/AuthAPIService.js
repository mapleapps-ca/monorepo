// File: monorepo/web/maplefile-frontend/src/services/API/AuthAPIService.js
// Authentication API Service - Handles all API calls for authentication

const API_BASE_URL = "/iam/api/v1"; // Using proxy from vite config

class AuthAPIService {
  constructor() {
    console.log("[AuthAPIService] API service initialized");
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
        `[AuthAPIService] Making ${requestOptions.method || "GET"} request to:`,
        url,
      );
      const response = await fetch(url, requestOptions);

      const data = await response.json();

      if (!response.ok) {
        console.error("[AuthAPIService] API Error:", data);
        throw new Error(
          data.details
            ? Object.values(data.details)[0]
            : data.error || "Request failed",
        );
      }

      console.log("[AuthAPIService] API Response:", data);
      return data;
    } catch (error) {
      console.error("[AuthAPIService] Request failed:", error);
      throw error;
    }
  }

  // Step 1: Request One-Time Token (OTT)
  async requestOTT(email) {
    try {
      console.log("[AuthAPIService] Requesting OTT for:", email);

      const response = await this.makeRequest("/request-ott", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
        }),
      });

      console.log("[AuthAPIService] OTT request successful");
      return response;
    } catch (error) {
      console.error("[AuthAPIService] OTT request failed:", error);
      throw new Error(`Failed to request OTT: ${error.message}`);
    }
  }

  // Step 2: Verify One-Time Token
  async verifyOTT(email, ott) {
    try {
      console.log("[AuthAPIService] Verifying OTT for:", email);

      const response = await this.makeRequest("/verify-ott", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
          ott: ott.trim(),
        }),
      });

      console.log("[AuthAPIService] OTT verification successful");
      return response;
    } catch (error) {
      console.error("[AuthAPIService] OTT verification failed:", error);
      throw new Error(`Failed to verify OTT: ${error.message}`);
    }
  }

  // Step 3: Complete Login
  async completeLogin(email, challengeId, decryptedChallenge) {
    try {
      console.log("[AuthAPIService] Completing login for:", email);

      const response = await this.makeRequest("/complete-login", {
        method: "POST",
        body: JSON.stringify({
          email: email.toLowerCase().trim(),
          challengeId: challengeId,
          decryptedData: decryptedChallenge,
        }),
      });

      console.log("[AuthAPIService] Login completion successful");
      return response;
    } catch (error) {
      console.error("[AuthAPIService] Login completion failed:", error);
      throw new Error(`Failed to complete login: ${error.message}`);
    }
  }

  // Token refresh endpoint
  async refreshTokens(refreshTokenValue) {
    try {
      console.log("[AuthAPIService] Refreshing tokens");

      const response = await this.makeRequest("/token/refresh", {
        method: "POST",
        body: JSON.stringify({
          value: refreshTokenValue,
        }),
      });

      console.log("[AuthAPIService] Token refresh successful");
      return response;
    } catch (error) {
      console.error("[AuthAPIService] Token refresh failed:", error);
      throw new Error(`Failed to refresh tokens: ${error.message}`);
    }
  }

  // User registration
  async registerUser(userData) {
    try {
      console.log("[AuthAPIService] Registering user");

      const response = await fetch(`${API_BASE_URL}/register`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(userData),
      });

      console.log(
        "[AuthAPIService] Registration response status:",
        response.status,
      );

      const result = await response.json();

      if (!response.ok) {
        console.error(
          "[AuthAPIService] Registration failed with status:",
          response.status,
        );
        console.error("[AuthAPIService] Error details:", result);
        throw new Error(
          result.details
            ? JSON.stringify(result.details)
            : result.error || "Registration failed",
        );
      }

      console.log("[AuthAPIService] Registration successful");
      return result;
    } catch (error) {
      console.error("[AuthAPIService] Registration error:", error);
      throw error;
    }
  }

  // Email verification
  async verifyEmail(verificationCode) {
    try {
      console.log("[AuthAPIService] Verifying email with code");

      const response = await fetch(`${API_BASE_URL}/verify-email-code`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          code: verificationCode.trim(),
        }),
      });

      const result = await response.json();

      if (!response.ok) {
        console.error(
          "[AuthAPIService] Email verification failed with status:",
          response.status,
        );
        console.error("[AuthAPIService] Error details:", result);
        throw new Error(
          result.details?.code || result.error || "Email verification failed",
        );
      }

      console.log("[AuthAPIService] Email verification successful");
      return result;
    } catch (error) {
      console.error("[AuthAPIService] Email verification error:", error);
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "AuthAPIService",
      baseUrl: API_BASE_URL,
      availableEndpoints: [
        "POST /request-ott",
        "POST /verify-ott",
        "POST /complete-login",
        "POST /token/refresh",
        "POST /register",
        "POST /verify-email-code",
      ],
    };
  }
}

export default AuthAPIService;
