// File: monorepo/web/maplefile-frontend/src/services/API/ApiClient.js
// Enhanced API Client with password service integration
import LocalStorageService from "../Storage/LocalStorageService.js";

const API_BASE_URL = "/iam/api/v1"; // Using proxy from vite config

class ApiClient {
  constructor() {
    this.isRefreshing = false;
    this.failedQueue = [];
    this._authManager = null; // Will be set by setAuthManager
    this._passwordService = null; // Will be set during initialization
  }

  // Set AuthManager instance for event notifications
  setAuthManager(authManager) {
    this._authManager = authManager;
    console.log("[ApiClient] AuthManager set for event notifications");
  }

  // ENHANCED: Initialize password service integration
  async initializePasswordService() {
    if (!this._passwordService) {
      try {
        const { default: passwordStorageService } = await import(
          "../PasswordStorageService.js"
        );
        this._passwordService = passwordStorageService;
        console.log("[ApiClient] Password service integration initialized");
      } catch (error) {
        console.warn(
          "[ApiClient] Failed to initialize password service:",
          error,
        );
      }
    }
  }

  // ENHANCED: Notify password service of successful API activity
  notifyApiSuccess() {
    if (this._passwordService && this._passwordService.recordApiActivity) {
      this._passwordService.recordApiActivity();
    }
  }

  // ENHANCED: Notify password service of successful token refresh
  notifyTokenRefreshSuccess() {
    if (this._passwordService && this._passwordService.recordTokenRefresh) {
      this._passwordService.recordTokenRefresh();
    }
  }

  // ENHANCED: Notify password service of successful auth operations
  notifyAuthSuccess() {
    if (this._passwordService && this._passwordService.recordAuthSuccess) {
      this._passwordService.recordAuthSuccess();
    }
  }

  // Notify auth state change via AuthManager
  notifyAuthStateChange(eventType, eventData) {
    if (this._authManager && this._authManager.notifyAuthStateChange) {
      this._authManager.notifyAuthStateChange(eventType, eventData);
    } else {
      console.log(`[ApiClient] Auth state change: ${eventType}`, eventData);
    }
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

      // Initialize password service integration
      await this.initializePasswordService();

      // Import and initialize CryptoService
      const { default: CryptoService } = await import(
        "../Crypto/CryptoService.js"
      );
      await CryptoService.initialize();

      // Try to get cached public key first
      let publicKey = LocalStorageService.getDerivedPublicKey();
      let privateKey = null;

      if (publicKey) {
        console.log("[ApiClient] Using cached public key for token decryption");
      } else {
        console.log(
          "[ApiClient] No cached public key, will derive from password",
        );
      }

      // Get password from PasswordStorageService
      const { default: passwordStorageService } = await import(
        "../PasswordStorageService.js"
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

      console.log("[ApiClient] Deriving private key from password...");

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
      privateKey = CryptoService.decryptWithSecretBox(
        encryptedPrivateKey,
        masterKey,
      );

      // If we didn't have a cached public key, derive it now and cache it
      if (!publicKey) {
        publicKey = CryptoService.sodium.crypto_scalarmult_base(privateKey);

        // Cache the derived public key for future use
        LocalStorageService.storeDerivedPublicKey(publicKey);
        console.log("[ApiClient] Derived and cached public key for future use");
      }

      console.log("[ApiClient] Keys ready, decrypting tokens...");

      // Decode encrypted tokens
      const encryptedAccessBytes =
        CryptoService.tryDecodeBase64(encryptedAccessToken);
      const encryptedRefreshBytes = CryptoService.tryDecodeBase64(
        encryptedRefreshToken,
      );

      console.log(
        "[ApiClient] Encrypted access token length:",
        encryptedAccessBytes.length,
      );
      console.log(
        "[ApiClient] Encrypted refresh token length:",
        encryptedRefreshBytes.length,
      );

      // Decrypt tokens using regular NaCl box (NOT sealed box)
      // Format: [32 bytes ephemeral pubkey][24 bytes nonce][encrypted data]
      let decryptedAccessBytes, decryptedRefreshBytes;

      try {
        console.log("[ApiClient] Decrypting access token using NaCl box...");

        // Extract components for access token
        const ephemeralPubKeyLength = 32;
        const nonceLength = 24;

        if (encryptedAccessBytes.length < ephemeralPubKeyLength + nonceLength) {
          throw new Error("Encrypted access token too short");
        }

        const accessEphemeralPubKey = encryptedAccessBytes.slice(
          0,
          ephemeralPubKeyLength,
        );
        const accessNonce = encryptedAccessBytes.slice(
          ephemeralPubKeyLength,
          ephemeralPubKeyLength + nonceLength,
        );
        const accessCiphertext = encryptedAccessBytes.slice(
          ephemeralPubKeyLength + nonceLength,
        );

        console.log("[ApiClient] Access token components:");
        console.log(
          "- Ephemeral public key length:",
          accessEphemeralPubKey.length,
        );
        console.log("- Nonce length:", accessNonce.length);
        console.log("- Ciphertext length:", accessCiphertext.length);

        // Decrypt using regular box
        decryptedAccessBytes = CryptoService.sodium.crypto_box_open_easy(
          accessCiphertext,
          accessNonce,
          accessEphemeralPubKey,
          privateKey,
        );

        console.log("[ApiClient] Access token decrypted successfully");
      } catch (error) {
        console.error("[ApiClient] Access token decryption failed:", error);
        throw new Error(`Access token decryption failed: ${error.message}`);
      }

      try {
        console.log("[ApiClient] Decrypting refresh token using NaCl box...");

        // Extract components for refresh token
        const ephemeralPubKeyLength = 32;
        const nonceLength = 24;

        if (
          encryptedRefreshBytes.length <
          ephemeralPubKeyLength + nonceLength
        ) {
          throw new Error("Encrypted refresh token too short");
        }

        const refreshEphemeralPubKey = encryptedRefreshBytes.slice(
          0,
          ephemeralPubKeyLength,
        );
        const refreshNonce = encryptedRefreshBytes.slice(
          ephemeralPubKeyLength,
          ephemeralPubKeyLength + nonceLength,
        );
        const refreshCiphertext = encryptedRefreshBytes.slice(
          ephemeralPubKeyLength + nonceLength,
        );

        console.log("[ApiClient] Refresh token components:");
        console.log(
          "- Ephemeral public key length:",
          refreshEphemeralPubKey.length,
        );
        console.log("- Nonce length:", refreshNonce.length);
        console.log("- Ciphertext length:", refreshCiphertext.length);

        // Decrypt using regular box
        decryptedRefreshBytes = CryptoService.sodium.crypto_box_open_easy(
          refreshCiphertext,
          refreshNonce,
          refreshEphemeralPubKey,
          privateKey,
        );

        console.log("[ApiClient] Refresh token decrypted successfully");
      } catch (error) {
        console.error("[ApiClient] Refresh token decryption failed:", error);
        throw new Error(`Refresh token decryption failed: ${error.message}`);
      }

      // Convert to strings
      const decryptedAccessToken = new TextDecoder().decode(
        decryptedAccessBytes,
      );
      const decryptedRefreshToken = new TextDecoder().decode(
        decryptedRefreshBytes,
      );

      // Validate JWT format
      const accessParts = decryptedAccessToken.split(".");
      const refreshParts = decryptedRefreshToken.split(".");

      if (accessParts.length !== 3) {
        throw new Error(`Invalid access token format`);
      }

      if (refreshParts.length !== 3) {
        throw new Error(`Invalid refresh token format`);
      }

      console.log("[ApiClient] Tokens decrypted successfully");

      return {
        access_token: decryptedAccessToken,
        refresh_token: decryptedRefreshToken,
      };
    } catch (error) {
      console.error("[ApiClient] Token decryption failed:", error);

      // Clear cached public key if decryption failed
      if (LocalStorageService.getDerivedPublicKey()) {
        console.log(
          "[ApiClient] Clearing cached public key due to decryption failure",
        );
        localStorage.removeItem("mapleapps_user_derived_public_key");
      }

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

      // ENHANCED: Notify password service of successful token refresh
      this.notifyTokenRefreshSuccess();

      // Notify listeners of successful refresh
      this.notifyAuthStateChange("token_refresh_success", {
        accessTokenExpiry: result.access_token_expiry_date,
        refreshTokenExpiry: result.refresh_token_expiry_date,
        username: result.username,
      });

      return result;
    } catch (error) {
      console.error("[ApiClient] Token refresh failed:", error);

      // Notify listeners of failed refresh
      this.notifyAuthStateChange("token_refresh_failed", {
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

        // ENHANCED: Notify of successful API activity
        if (response.ok) {
          this.notifyApiSuccess();
        }

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

      // ENHANCED: Notify of successful API activity
      if (response.ok) {
        this.notifyApiSuccess();
      }

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

      // ENHANCED: Notify of successful API activity
      if (response.ok) {
        this.notifyApiSuccess();
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

      // ENHANCED: Notify of successful API activity
      if (response.ok) {
        this.notifyApiSuccess();
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
const apiClient = new ApiClient();

// Helper function to set AuthManager after initialization
export const setApiClientAuthManager = (authManager) => {
  apiClient.setAuthManager(authManager);

  // Initialize password service integration
  apiClient.initializePasswordService().catch((error) => {
    console.warn("[ApiClient] Failed to initialize password service:", error);
  });
};

export default apiClient;
