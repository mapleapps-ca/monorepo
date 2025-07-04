// File: monorepo/web/maplefile-frontend/src/services/ApiClient.js
// Enhanced API Client with automatic token refresh interceptor
import LocalStorageService from "./LocalStorageService.js";
import WorkerManager from "./WorkerManager.js";

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

  // Decrypt encrypted tokens using password and user's private key
  async decryptTokensFromRefresh(
    encryptedAccessToken,
    encryptedRefreshToken,
    tokenNonce,
  ) {
    try {
      console.log("[ApiClient] Starting token decryption process");

      // Get password from PasswordStorageService
      const { default: passwordStorageService } = await import(
        "./PasswordStorageService.js"
      );
      const password = passwordStorageService.getPassword();

      if (!password) {
        throw new Error(
          "No password available for token decryption - please log in again",
        );
      }

      // Get user's encrypted data
      const userEncryptedData = LocalStorageService.getUserEncryptedData();
      if (
        !userEncryptedData.salt ||
        !userEncryptedData.encryptedMasterKey ||
        !userEncryptedData.encryptedPrivateKey
      ) {
        throw new Error(
          "Missing user encrypted data for token decryption - please log in again",
        );
      }

      // Import and initialize CryptoService
      const { default: CryptoService } = await import(
        "./Crypto/CryptoService.js"
      );
      await CryptoService.initialize();

      console.log("[ApiClient] Deriving keys from password...");

      // Decode user's encrypted data
      const salt = CryptoService.tryDecodeBase64(userEncryptedData.salt);
      const encryptedMasterKey = CryptoService.tryDecodeBase64(
        userEncryptedData.encryptedMasterKey,
      );
      const encryptedPrivateKey = CryptoService.tryDecodeBase64(
        userEncryptedData.encryptedPrivateKey,
      );

      // Derive key encryption key from password
      const keyEncryptionKey = await CryptoService.deriveKeyFromPassword(
        password,
        salt,
      );

      // Decrypt master key
      const masterKey = CryptoService.decryptWithSecretBox(
        encryptedMasterKey,
        keyEncryptionKey,
      );

      // Decrypt private key
      const privateKey = CryptoService.decryptWithSecretBox(
        encryptedPrivateKey,
        masterKey,
      );

      // Derive public key
      let publicKey;
      if (userEncryptedData.publicKey) {
        publicKey = CryptoService.tryDecodeBase64(userEncryptedData.publicKey);
      } else {
        publicKey = CryptoService.sodium.crypto_scalarmult_base(privateKey);
      }

      console.log(
        "[ApiClient] Keys derived successfully, decrypting tokens...",
      );

      // Decode encrypted tokens
      const encryptedAccessBytes =
        CryptoService.tryDecodeBase64(encryptedAccessToken);
      const encryptedRefreshBytes = CryptoService.tryDecodeBase64(
        encryptedRefreshToken,
      );

      // Decrypt tokens using sealed box (anonymous encryption)
      const decryptedAccessBytes = CryptoService.sodium.crypto_box_seal_open(
        encryptedAccessBytes,
        publicKey,
        privateKey,
      );
      const decryptedRefreshBytes = CryptoService.sodium.crypto_box_seal_open(
        encryptedRefreshBytes,
        publicKey,
        privateKey,
      );

      // Convert to strings
      const decryptedAccessToken = new TextDecoder().decode(
        decryptedAccessBytes,
      );
      const decryptedRefreshToken = new TextDecoder().decode(
        decryptedRefreshBytes,
      );

      // Validate JWT format
      if (
        decryptedAccessToken.split(".").length !== 3 ||
        decryptedRefreshToken.split(".").length !== 3
      ) {
        throw new Error("Decrypted tokens are not valid JWT format");
      }

      console.log("[ApiClient] Tokens decrypted successfully");

      return {
        access_token: decryptedAccessToken,
        refresh_token: decryptedRefreshToken,
      };
    } catch (error) {
      console.error("[ApiClient] Token decryption failed:", error);
      throw new Error(`Token decryption failed: ${error.message}`);
    }
  }

  // Enhanced token refresh with decryption
  async refreshTokens() {
    try {
      console.log("[ApiClient] Starting token refresh process");

      const refreshToken = LocalStorageService.getRefreshToken();
      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      // Call refresh endpoint
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
      console.log("[ApiClient] Token refresh response received");

      // Handle different response formats
      let decryptedTokens;

      if (
        result.encrypted_access_token &&
        result.encrypted_refresh_token &&
        result.token_nonce
      ) {
        // Encrypted tokens - need to decrypt
        console.log("[ApiClient] Received encrypted tokens, decrypting...");
        decryptedTokens = await this.decryptTokensFromRefresh(
          result.encrypted_access_token,
          result.encrypted_refresh_token,
          result.token_nonce,
        );
      } else if (result.access_token && result.refresh_token) {
        // Already unencrypted tokens
        console.log("[ApiClient] Received unencrypted tokens");
        decryptedTokens = {
          access_token: result.access_token,
          refresh_token: result.refresh_token,
        };
      } else {
        throw new Error("Invalid token refresh response format");
      }

      // Store the new unencrypted tokens
      LocalStorageService.setTokens(
        decryptedTokens.access_token,
        decryptedTokens.refresh_token,
        result.access_token_expiry_date,
        result.refresh_token_expiry_date,
      );

      // Update user email if provided
      if (result.username) {
        LocalStorageService.setUserEmail(result.username);
      }

      console.log("[ApiClient] Tokens refreshed and stored successfully");

      // Notify listeners of successful refresh
      WorkerManager.notifyAuthStateChange("token_refresh_success", {
        accessTokenExpiry: result.access_token_expiry_date,
        refreshTokenExpiry: result.refresh_token_expiry_date,
        username: result.username,
      });

      return result;
    } catch (error) {
      console.error("[ApiClient] Token refresh failed:", error);

      // Notify listeners of failed refresh
      WorkerManager.notifyAuthStateChange("token_refresh_failed", {
        error: error.message,
      });

      throw error;
    }
  }

  // Handle 401 Unauthorized responses with automatic token refresh
  async handleUnauthorizedResponse(
    url,
    originalRequestOptions,
    isMapleFileAPI = false,
  ) {
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

        if (isMapleFileAPI && response.status === 204) {
          return null;
        }

        return response.json();
      });
    }

    this.isRefreshing = true;

    try {
      console.log("[ApiClient] Access token expired, refreshing tokens...");

      // Refresh the tokens
      await this.refreshTokens();

      // Update authorization header with new token
      const newAuthHeader = this.getAuthorizationHeader();
      if (newAuthHeader) {
        originalRequestOptions.headers["Authorization"] = newAuthHeader;
      }

      this.isRefreshing = false;
      this.processQueue(null, newAuthHeader);

      // Retry the original request
      const response = await fetch(url, originalRequestOptions);

      if (isMapleFileAPI && response.status === 204) {
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
      console.error("[ApiClient] Token refresh failed:", refreshError);
      this.isRefreshing = false;
      this.processQueue(refreshError, null);

      // Clear authentication data and redirect to login
      this.clearAuthData();
      this.redirectToLogin("Session expired - please log in again");

      throw new Error("Session expired. Please log in again.");
    }
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

      // Handle authentication errors with automatic refresh
      if (response.status === 401) {
        return this.handleUnauthorizedResponse(url, requestOptions, false);
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

      // Handle authentication errors with automatic refresh
      if (response.status === 401) {
        return this.handleUnauthorizedResponse(url, requestOptions, true);
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
    // Add debug logging
    console.log("[ApiClient] PUT request to MapleFile:", {
      endpoint,
      method: "PUT",
      dataKeys: data ? Object.keys(data) : null,
      hasId: data?.id,
      version: data?.version,
    });

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
