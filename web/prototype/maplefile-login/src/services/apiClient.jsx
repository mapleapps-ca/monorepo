// API Client with encrypted token support and automatic token refresh
import authService from "./authService.js";
import localStorageService from "./localStorageService.js";

const API_BASE_URL = "/iam/api/v1"; // Using proxy from vite config

class ApiClient {
  constructor() {
    this.isRefreshing = false;
    this.failedQueue = [];
  }

  // Process failed requests queue after token refresh
  processQueue(error = null, token = null) {
    this.failedQueue.forEach(({ resolve, reject }) => {
      if (error) {
        reject(error);
      } else {
        resolve(token);
      }
    });

    this.failedQueue = [];
  }

  // Get authorization header for encrypted token system
  async getAuthorizationHeader() {
    try {
      // Check if we have encrypted tokens
      const encryptedTokens = localStorageService.getEncryptedTokens();
      const tokenNonce = localStorageService.getTokenNonce();

      if (encryptedTokens && tokenNonce) {
        // For encrypted tokens, we need to implement token decryption
        // For now, we'll use a placeholder approach
        console.log("[ApiClient] Using encrypted token system");

        // TODO: Implement actual token decryption here
        // This would require:
        // 1. User's private key (derived from password or stored securely)
        // 2. Decryption of the encrypted tokens using the nonce
        // 3. Extraction of the actual JWT access token

        // For now, return null to indicate we have encrypted tokens but can't use them yet
        // In a real implementation, you would decrypt and return the JWT token
        return null;
      }

      // Fallback to legacy token system
      const accessToken = localStorageService.getAccessToken();
      if (accessToken && !accessToken.startsWith("encrypted_")) {
        return `JWT ${accessToken}`;
      }

      return null;
    } catch (error) {
      console.error("[ApiClient] Error getting authorization header:", error);
      return null;
    }
  }

  // Make authenticated API request with automatic token refresh
  async request(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;

    // Prepare headers
    const headers = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    // Add authorization header if we have a decrypted token
    const authHeader = await this.getAuthorizationHeader();
    if (authHeader) {
      headers["Authorization"] = authHeader;
    }

    const requestOptions = {
      ...options,
      headers,
    };

    try {
      console.log(`Making ${requestOptions.method || "GET"} request to:`, url);

      // Make the API request
      const response = await fetch(url, requestOptions);

      // Handle authentication errors
      if (response.status === 401) {
        return this.handleUnauthorized(url, requestOptions);
      }

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
      // Handle network errors
      if (error.name === "TypeError" && error.message.includes("fetch")) {
        throw new Error("Network error - please check your connection");
      }
      throw error;
    }
  }

  // Handle 401 Unauthorized responses with token refresh
  async handleUnauthorized(url, originalRequestOptions) {
    // If already refreshing, queue this request
    if (this.isRefreshing) {
      return new Promise((resolve, reject) => {
        this.failedQueue.push({ resolve, reject });
      }).then(async () => {
        // Retry original request with new token
        const authHeader = await this.getAuthorizationHeader();
        if (authHeader) {
          originalRequestOptions.headers["Authorization"] = authHeader;
        }
        const response = await fetch(url, originalRequestOptions);
        return response.json();
      });
    }

    this.isRefreshing = true;

    try {
      console.log("[ApiClient] Access token expired, attempting refresh...");

      // Check if we have encrypted tokens to refresh
      const encryptedTokens = localStorageService.getEncryptedTokens();
      if (!encryptedTokens) {
        throw new Error("No refresh tokens available");
      }

      // Try to refresh the token using the auth service
      await authService.refreshToken();

      // Update authorization header with new token
      const newAuthHeader = await this.getAuthorizationHeader();
      if (newAuthHeader) {
        originalRequestOptions.headers["Authorization"] = newAuthHeader;
      }

      this.isRefreshing = false;
      this.processQueue(null, newAuthHeader);

      // Retry the original request
      const response = await fetch(url, originalRequestOptions);
      const data = await response.json();

      if (!response.ok) {
        throw new Error(
          data.details
            ? Object.values(data.details)[0]
            : data.error || "Request failed",
        );
      }

      return data;
    } catch (refreshError) {
      console.error("[ApiClient] Token refresh failed:", refreshError);
      this.isRefreshing = false;
      this.processQueue(refreshError, null);

      // Clear authentication data and redirect to login
      authService.logout();

      // Redirect to login page
      if (window.location.pathname !== "/") {
        window.location.href = "/";
      }

      throw new Error("Session expired. Please log in again.");
    }
  }

  // Check if we can make authenticated requests
  canMakeAuthenticatedRequests() {
    // Check if we have any form of valid tokens
    return localStorageService.hasValidTokens();
  }

  // Make request for public endpoints (no auth required)
  async requestPublic(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;

    const requestOptions = {
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
      ...options,
    };

    try {
      console.log(
        `Making public ${requestOptions.method || "GET"} request to:`,
        url,
      );

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
      if (error.name === "TypeError" && error.message.includes("fetch")) {
        throw new Error("Network error - please check your connection");
      }
      throw error;
    }
  }

  // Convenience methods for different HTTP verbs
  async get(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: "GET" });
  }

  async post(endpoint, data = null, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: "POST",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put(endpoint, data = null, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: "PUT",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async patch(endpoint, data = null, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: "PATCH",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: "DELETE" });
  }

  // Public endpoint methods (no authentication)
  async getPublic(endpoint, options = {}) {
    return this.requestPublic(endpoint, { ...options, method: "GET" });
  }

  async postPublic(endpoint, data = null, options = {}) {
    return this.requestPublic(endpoint, {
      ...options,
      method: "POST",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // Method to check token status
  getTokenStatus() {
    const encryptedTokens = localStorageService.getEncryptedTokens();
    const tokenNonce = localStorageService.getTokenNonce();
    const tokenInfo = localStorageService.getTokenExpiryInfo();

    return {
      hasEncryptedTokens: !!(encryptedTokens && tokenNonce),
      ...tokenInfo,
      canMakeAuthenticatedRequests: this.canMakeAuthenticatedRequests(),
      tokenDecryptionImplemented: false, // Set to true when decryption is implemented
    };
  }

  // Utility method to clear failed request queue
  clearFailedQueue() {
    this.failedQueue = [];
  }
}

// Export singleton instance
export default new ApiClient();
