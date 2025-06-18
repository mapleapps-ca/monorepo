// API Client with automatic token refresh and authentication headers
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

  // Make authenticated API request with automatic token refresh
  async request(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;

    // Prepare headers
    const headers = {
      "Content-Type": "application/json",
      ...options.headers,
    };

    // Add authorization header if access token exists
    const accessToken = localStorageService.getAccessToken();
    if (accessToken && !accessToken.startsWith("encrypted_")) {
      headers["Authorization"] = `JWT ${accessToken}`;
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
      }).then(() => {
        // Retry original request with new token
        const accessToken = localStorageService.getAccessToken();
        if (accessToken && !accessToken.startsWith("encrypted_")) {
          originalRequestOptions.headers["Authorization"] =
            `JWT ${accessToken}`;
        }
        return fetch(url, originalRequestOptions).then((res) => res.json());
      });
    }

    this.isRefreshing = true;

    try {
      console.log("Access token expired, attempting refresh...");

      // Try to refresh the token
      await authService.refreshToken();

      // Update authorization header with new token
      const newAccessToken = localStorageService.getAccessToken();
      if (newAccessToken && !newAccessToken.startsWith("encrypted_")) {
        originalRequestOptions.headers["Authorization"] =
          `JWT ${newAccessToken}`;
      }

      this.isRefreshing = false;
      this.processQueue(null, newAccessToken);

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
      console.error("Token refresh failed:", refreshError);
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
}

// Export singleton instance
export default new ApiClient();
