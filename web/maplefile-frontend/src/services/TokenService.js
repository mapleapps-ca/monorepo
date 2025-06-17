// src/services/TokenService.js

import axios from "axios";

/**
 * TokenService provides a clean interface for token management
 * This service wraps token storage, refresh, and lifecycle management
 * It demonstrates several important concepts:
 * 1. Automatic background token refresh
 * 2. Event-driven architecture for token state changes
 * 3. Secure token storage and handling
 * 4. Integration with API authentication headers
 */
export class TokenService {
  constructor(logger) {
    this.logger = logger;

    // Build API URL from environment variables
    this.apiBaseUrl = `${import.meta.env.VITE_API_PROTOCOL}://${import.meta.env.VITE_API_DOMAIN}`;

    // Token refresh configuration
    this.refreshTimeoutId = null;
    this.refreshIntervalMs = 60000; // Check every minute
    this.refreshBeforeExpiryMs = 60000; // Refresh 1 minute before expiry
    this.isRefreshing = false;

    // Event listeners for token state changes
    this.listeners = [];

    this.logger.log("TokenService: Initialized");
  }

  /**
   * Initialize the token service and start automatic refresh
   * This should be called when the application starts
   */
  async initialize() {
    this.logger.log("TokenService: Starting initialization");

    // Clear any existing refresh timers
    if (this.refreshTimeoutId) {
      clearTimeout(this.refreshTimeoutId);
    }

    // Start the automatic token refresh cycle
    this.scheduleTokenRefresh();

    // Return current authentication status
    const isAuthenticated = this.isAuthenticated();
    this.logger.log(
      `TokenService: Initialization complete, authenticated: ${isAuthenticated}`,
    );

    return isAuthenticated;
  }

  /**
   * Store authentication tokens securely
   * This method handles both token storage and automatic refresh scheduling
   */
  storeTokens(tokens) {
    this.logger.log("TokenService: Storing authentication tokens");

    try {
      // Store the tokens in localStorage
      localStorage.setItem("accessToken", tokens.accessToken);
      localStorage.setItem("refreshToken", tokens.refreshToken);

      // Handle token expiry dates safely
      const now = new Date();
      let accessExpiry, refreshExpiry;

      // Parse or generate expiry times
      if (tokens.accessTokenExpiry) {
        accessExpiry = new Date(tokens.accessTokenExpiry);
      } else {
        // Default to 30 minutes if no expiry provided
        accessExpiry = new Date(now.getTime() + 30 * 60 * 1000);
      }

      if (tokens.refreshTokenExpiry) {
        refreshExpiry = new Date(tokens.refreshTokenExpiry);
      } else {
        // Default to 14 days if no expiry provided
        refreshExpiry = new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000);
      }

      // Validate the dates before storing
      if (isNaN(accessExpiry.getTime())) {
        this.logger.log(
          "TokenService: Invalid access token expiry, using default",
        );
        accessExpiry = new Date(now.getTime() + 30 * 60 * 1000);
      }

      if (isNaN(refreshExpiry.getTime())) {
        this.logger.log(
          "TokenService: Invalid refresh token expiry, using default",
        );
        refreshExpiry = new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000);
      }

      // Store expiry times as ISO strings
      localStorage.setItem("accessTokenExpiry", accessExpiry.toISOString());
      localStorage.setItem("refreshTokenExpiry", refreshExpiry.toISOString());

      this.logger.log("TokenService: Tokens stored successfully");

      // Restart the refresh cycle with new tokens
      this.scheduleTokenRefresh();

      // Notify listeners that tokens were updated
      this.notifyListeners("tokensStored", true);
    } catch (error) {
      this.logger.log(`TokenService: Error storing tokens: ${error.message}`);
      throw error;
    }
  }

  /**
   * Check if the user is currently authenticated
   */
  isAuthenticated() {
    const accessToken = localStorage.getItem("accessToken");
    const refreshToken = localStorage.getItem("refreshToken");

    // We need both tokens to consider the user authenticated
    return !!(accessToken && refreshToken);
  }

  /**
   * Get the current access token
   * This is used by API calls to authenticate requests
   */
  getAccessToken() {
    return localStorage.getItem("accessToken");
  }

  /**
   * Get authentication headers for API requests
   * This provides a convenient way to add auth headers to axios requests
   */
  getAuthHeaders() {
    const accessToken = this.getAccessToken();
    if (accessToken) {
      return {
        Authorization: `Bearer ${accessToken}`,
      };
    }
    return {};
  }

  /**
   * Schedule the next token refresh check
   * This implements the automatic background refresh functionality
   */
  scheduleTokenRefresh() {
    // Clear any existing timer first
    if (this.refreshTimeoutId) {
      clearTimeout(this.refreshTimeoutId);
    }

    // Schedule the next check
    this.refreshTimeoutId = setTimeout(() => {
      this.checkTokenExpiration();
    }, this.refreshIntervalMs);
  }

  /**
   * Check if tokens need to be refreshed
   * This is called periodically by the scheduler
   */
  async checkTokenExpiration() {
    this.logger.log("TokenService: Checking token expiration");

    try {
      const accessToken = localStorage.getItem("accessToken");
      const refreshToken = localStorage.getItem("refreshToken");
      const accessTokenExpiryStr = localStorage.getItem("accessTokenExpiry");

      // If no tokens, nothing to check
      if (!accessToken || !refreshToken) {
        this.logger.log(
          "TokenService: No tokens found, skipping refresh check",
        );
        this.scheduleTokenRefresh();
        return;
      }

      // Check if access token is expired or about to expire
      const now = new Date();
      const accessTokenExpiry = accessTokenExpiryStr
        ? new Date(accessTokenExpiryStr)
        : null;

      if (accessTokenExpiry) {
        const timeUntilExpiry = accessTokenExpiry.getTime() - now.getTime();
        this.logger.log(
          `TokenService: Access token expires in ${Math.round(timeUntilExpiry / 1000)} seconds`,
        );

        // If token expires within our threshold, refresh it
        if (timeUntilExpiry < this.refreshBeforeExpiryMs) {
          await this.refreshTokens();
        }
      }
    } catch (error) {
      this.logger.log(
        `TokenService: Error checking token expiration: ${error.message}`,
      );
    }

    // Schedule the next check regardless of success/failure
    this.scheduleTokenRefresh();
  }

  /**
   * Refresh the authentication tokens
   * This is called automatically when tokens are about to expire
   */
  async refreshTokens() {
    // Prevent multiple simultaneous refresh attempts
    if (this.isRefreshing) {
      this.logger.log("TokenService: Refresh already in progress, skipping");
      return;
    }

    this.isRefreshing = true;

    try {
      this.logger.log("TokenService: Starting token refresh");

      const refreshToken = localStorage.getItem("refreshToken");
      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      // Call the refresh token API
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/refresh-token`,
        {
          refresh_token: refreshToken,
        },
        {
          headers: { "Content-Type": "application/json" },
          timeout: 10000,
        },
      );

      if (response.status === 200) {
        const data = response.data;

        // Store the new tokens
        this.storeTokens({
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
          accessTokenExpiry: data.access_token_expiry_time,
          refreshTokenExpiry: data.refresh_token_expiry_time,
        });

        this.logger.log("TokenService: Tokens refreshed successfully");
        this.notifyListeners("tokensRefreshed", true);
      } else {
        throw new Error(`Token refresh failed with status: ${response.status}`);
      }
    } catch (error) {
      this.logger.log(`TokenService: Token refresh failed: ${error.message}`);

      // If refresh fails, the user needs to log in again
      this.clearTokens();
      this.notifyListeners("tokensRefreshed", false, error);

      // Optionally redirect to login
      this.handleAuthenticationFailure();
    } finally {
      this.isRefreshing = false;
    }
  }

  /**
   * Clear all stored tokens
   * This is called during logout or when authentication fails
   */
  clearTokens() {
    this.logger.log("TokenService: Clearing all tokens");

    // Remove tokens from storage
    localStorage.removeItem("accessToken");
    localStorage.removeItem("accessTokenExpiry");
    localStorage.removeItem("refreshToken");
    localStorage.removeItem("refreshTokenExpiry");

    // Clear any scheduled refresh
    if (this.refreshTimeoutId) {
      clearTimeout(this.refreshTimeoutId);
      this.refreshTimeoutId = null;
    }

    // Notify listeners that tokens were cleared
    this.notifyListeners("tokensCleared", true);
  }

  /**
   * Handle authentication failure
   * This determines what happens when token refresh fails
   */
  handleAuthenticationFailure() {
    this.logger.log("TokenService: Handling authentication failure");

    // In a real application, you might want to show a notification
    // before redirecting, or allow the user to save their work

    // For now, we'll just notify listeners and let them handle the redirect
    this.notifyListeners("authenticationFailed", false);
  }

  /**
   * Add a listener for token events
   * This allows other parts of the application to react to token changes
   */
  addListener(listener) {
    this.listeners.push(listener);

    // Return a function to remove the listener
    return () => this.removeListener(listener);
  }

  /**
   * Remove a token event listener
   */
  removeListener(listener) {
    this.listeners = this.listeners.filter((l) => l !== listener);
  }

  /**
   * Notify all listeners of a token event
   * This implements the observer pattern for token state changes
   */
  notifyListeners(event, success, error = null) {
    this.logger.log(
      `TokenService: Notifying listeners of event: ${event}, success: ${success}`,
    );

    this.listeners.forEach((listener) => {
      try {
        listener(event, success, error);
      } catch (e) {
        this.logger.log(`TokenService: Error in listener: ${e.message}`);
      }
    });
  }

  /**
   * Get current token information for debugging
   * This is useful for development and troubleshooting
   */
  getTokenInfo() {
    const accessToken = localStorage.getItem("accessToken");
    const refreshToken = localStorage.getItem("refreshToken");
    const accessExpiry = localStorage.getItem("accessTokenExpiry");
    const refreshExpiry = localStorage.getItem("refreshTokenExpiry");

    return {
      hasAccessToken: !!accessToken,
      hasRefreshToken: !!refreshToken,
      accessTokenExpiry: accessExpiry ? new Date(accessExpiry) : null,
      refreshTokenExpiry: refreshExpiry ? new Date(refreshExpiry) : null,
      isRefreshing: this.isRefreshing,
    };
  }

  /**
   * Cleanup the token service
   * This should be called when the application is shutting down
   */
  cleanup() {
    this.logger.log("TokenService: Cleaning up");

    if (this.refreshTimeoutId) {
      clearTimeout(this.refreshTimeoutId);
      this.refreshTimeoutId = null;
    }

    this.listeners = [];
    this.isRefreshing = false;
  }
}
