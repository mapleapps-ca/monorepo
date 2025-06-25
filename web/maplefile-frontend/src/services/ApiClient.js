// API Client with unencrypted token support
import LocalStorageService from "./LocalStorageService.js";

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

  // Get authorization header for unencrypted tokens
  getAuthorizationHeader() {
    const accessToken = LocalStorageService.getAccessToken();

    if (accessToken) {
      console.log("[ApiClient] Using unencrypted access token");
      return `JWT ${accessToken}`;
    }

    console.log("[ApiClient] No access token available");
    return null;
  }

  // Check if we can make authenticated requests
  canMakeAuthenticatedRequests() {
    return LocalStorageService.hasValidTokens();
  }

  // Make authenticated API request with automatic token refresh
  async request(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;

    // Check if we can make authenticated requests BEFORE making the request
    if (!this.canMakeAuthenticatedRequests()) {
      console.error(
        "[ApiClient] Cannot make authenticated requests - no valid tokens",
      );
      throw new Error("Authentication required - please log in again");
    }

    // Prepare headers
    const headers = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    // Add authorization header
    const authHeader = this.getAuthorizationHeader();
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

  // Make request for MapleFile API endpoints
  async requestMapleFile(endpoint, options = {}) {
    const url = `/maplefile/api/v1${endpoint}`;

    // Check if we can make authenticated requests BEFORE making the request
    if (!this.canMakeAuthenticatedRequests()) {
      console.error(
        "[ApiClient] Cannot make authenticated requests to MapleFile - no valid tokens",
      );
      throw new Error("Authentication required - please log in again");
    }

    // Prepare headers
    const headers = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    // Add authorization header
    const authHeader = this.getAuthorizationHeader();
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
        return this.handleUnauthorizedMapleFile(url, requestOptions);
      }

      // Handle 204 No Content responses
      if (response.status === 204) {
        return null;
      }

      const data = await response.json();

      if (!response.ok) {
        console.error("MapleFile API Error:", data);
        throw new Error(
          data.details
            ? Object.values(data.details)[0]
            : data.error || "Request failed",
        );
      }

      console.log("MapleFile API Response:", data);
      return data;
    } catch (error) {
      // Handle network errors
      if (error.name === "TypeError" && error.message.includes("fetch")) {
        throw new Error("Network error - please check your connection");
      }
      throw error;
    }
  }

  // Handle 401 Unauthorized responses with token refresh for IAM API
  async handleUnauthorized(url, originalRequestOptions) {
    // If already refreshing, queue this request
    if (this.isRefreshing) {
      return new Promise((resolve, reject) => {
        this.failedQueue.push({ resolve, reject });
      }).then(async () => {
        // Retry original request with new token
        const authHeader = this.getAuthorizationHeader();
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

      // Check if we have a refresh token
      const refreshToken = LocalStorageService.getRefreshToken();
      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      // Try to refresh the token directly using the API
      await this.refreshTokenDirect();

      // Update authorization header with new token
      const newAuthHeader = this.getAuthorizationHeader();
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
      this.clearAuthData();
      this.redirectToLogin("Session expired - please log in again");

      throw new Error("Session expired. Please log in again.");
    }
  }

  // Handle 401 Unauthorized responses for MapleFile API
  async handleUnauthorizedMapleFile(url, originalRequestOptions) {
    // If already refreshing, queue this request
    if (this.isRefreshing) {
      return new Promise((resolve, reject) => {
        this.failedQueue.push({ resolve, reject });
      }).then(async () => {
        // Retry original request with new token
        const authHeader = this.getAuthorizationHeader();
        if (authHeader) {
          originalRequestOptions.headers["Authorization"] = authHeader;
        }
        const response = await fetch(url, originalRequestOptions);

        if (response.status === 204) {
          return null;
        }

        return response.json();
      });
    }

    this.isRefreshing = true;

    try {
      console.log(
        "[ApiClient] MapleFile access token expired, attempting refresh...",
      );

      // Check if we have a refresh token
      const refreshToken = LocalStorageService.getRefreshToken();
      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      // Try to refresh the token directly using the IAM API
      await this.refreshTokenDirect();

      // Update authorization header with new token
      const newAuthHeader = this.getAuthorizationHeader();
      if (newAuthHeader) {
        originalRequestOptions.headers["Authorization"] = newAuthHeader;
      }

      this.isRefreshing = false;
      this.processQueue(null, newAuthHeader);

      // Retry the original request
      const response = await fetch(url, originalRequestOptions);

      if (response.status === 204) {
        return null;
      }

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
      console.error(
        "[ApiClient] MapleFile token refresh failed:",
        refreshError,
      );
      this.isRefreshing = false;
      this.processQueue(refreshError, null);

      // Clear authentication data and redirect to login
      this.clearAuthData();
      this.redirectToLogin("Session expired - please log in again");

      throw new Error("Session expired. Please log in again.");
    }
  }

  // Redirect to login page with message
  redirectToLogin(message) {
    console.log(`[ApiClient] Redirecting to login: ${message}`);

    // Use setTimeout to ensure current execution completes
    setTimeout(() => {
      if (window.location.pathname !== "/") {
        // Store the redirect message for display
        sessionStorage.setItem("auth_redirect_message", message);
        window.location.href = "/";
      }
    }, 100);
  }

  // Direct token refresh method
  async refreshTokenDirect() {
    try {
      console.log("[ApiClient] Starting direct token refresh");

      const refreshToken = LocalStorageService.getRefreshToken();
      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      // Use the token refresh API endpoint
      const response = await fetch(`${API_BASE_URL}/token/refresh`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          value: refreshToken,
        }),
      });

      if (response.status !== 201) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Token refresh failed");
      }

      const result = await response.json();
      console.log("[ApiClient] Direct token refresh successful");

      // Handle refreshed tokens - should be unencrypted
      if (result.access_token && result.refresh_token) {
        console.log("[ApiClient] Received unencrypted tokens from refresh");

        // Store the new unencrypted tokens
        LocalStorageService.setTokens(
          result.access_token,
          result.refresh_token,
          result.access_token_expiry_date,
          result.refresh_token_expiry_date,
        );

        // Update user email if provided
        if (result.username) {
          LocalStorageService.setUserEmail(result.username);
        }

        console.log("[ApiClient] New tokens stored successfully");
        return result;
      } else if (result.encrypted_tokens && result.token_nonce) {
        // This shouldn't happen after initial login
        console.error(
          "[ApiClient] Received encrypted tokens in refresh - this is unexpected",
        );
        throw new Error(
          "Backend returned encrypted tokens in refresh response - configuration issue",
        );
      } else {
        console.error("[ApiClient] No valid tokens in refresh response");
        console.error(
          "[ApiClient] Available response fields:",
          Object.keys(result),
        );
        throw new Error("Token refresh failed: No valid tokens received");
      }
    } catch (error) {
      console.error("[ApiClient] Direct token refresh failed:", error);
      throw error;
    }
  }

  // Clear authentication data
  clearAuthData() {
    LocalStorageService.clearAuthData();
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

  // MapleFile API methods
  async getMapleFile(endpoint, options = {}) {
    return this.requestMapleFile(endpoint, { ...options, method: "GET" });
  }

  async postMapleFile(endpoint, data = null, options = {}) {
    return this.requestMapleFile(endpoint, {
      ...options,
      method: "POST",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async putMapleFile(endpoint, data = null, options = {}) {
    return this.requestMapleFile(endpoint, {
      ...options,
      method: "PUT",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async deleteMapleFile(endpoint, options = {}) {
    return this.requestMapleFile(endpoint, { ...options, method: "DELETE" });
  }

  // Method to check token status
  getTokenStatus() {
    const tokenInfo = LocalStorageService.getTokenExpiryInfo();

    return {
      hasValidTokens: LocalStorageService.hasValidTokens(),
      canMakeAuthenticatedRequests: this.canMakeAuthenticatedRequests(),
      ...tokenInfo,
    };
  }

  // Utility method to clear failed request queue
  clearFailedQueue() {
    this.failedQueue = [];
  }
}

// Export singleton instance
export default new ApiClient();
