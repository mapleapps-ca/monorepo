// web/prototyping/papercloud-frontend/src/services/TokenManager.js
import { authAPI } from "./api";

class TokenManager {
  constructor() {
    this.refreshTimeoutId = null;
    this.refreshIntervalMs = 60000; // Check every minute
    this.refreshBeforeExpiryMs = 60000; // Refresh 1 minute before expiry
    this.listeners = [];
    this.isRefreshing = false;
  }

  /**
   * Initialize the token manager
   * Should be called when the application starts
   */
  initialize() {
    console.log("TokenManager: Initializing");

    // Clear any existing timers
    if (this.refreshTimeoutId) {
      clearTimeout(this.refreshTimeoutId);
    }

    // Start the token refresh checker
    this.scheduleTokenRefresh();

    // Return the current access token if it exists
    return this.getAccessToken();
  }

  /**
   * Schedule the next token refresh check
   */
  scheduleTokenRefresh() {
    // Clear any existing timer
    if (this.refreshTimeoutId) {
      clearTimeout(this.refreshTimeoutId);
    }

    this.refreshTimeoutId = setTimeout(() => {
      this.checkTokenExpiration();
    }, this.refreshIntervalMs);
  }

  /**
   * Check if the access token is expired or about to expire
   */
  async checkTokenExpiration() {
    console.log("TokenManager: Checking token expiration");

    try {
      // Get the current tokens and their expiry times
      const accessToken = localStorage.getItem("accessToken");
      const accessTokenExpiryStr = localStorage.getItem("accessTokenExpiry");
      const refreshToken = localStorage.getItem("refreshToken");

      // If no tokens, we're not logged in
      if (!accessToken || !refreshToken) {
        console.log("TokenManager: No tokens found");
        this.scheduleTokenRefresh();
        return;
      }

      // Convert expiry string to Date
      const accessTokenExpiry = accessTokenExpiryStr
        ? new Date(accessTokenExpiryStr)
        : null;

      // Calculate time until expiry
      const now = new Date();
      const timeUntilExpiry = accessTokenExpiry
        ? accessTokenExpiry.getTime() - now.getTime()
        : -1;

      console.log(
        `TokenManager: Access token expires in ${timeUntilExpiry / 1000} seconds`,
      );

      // If token is expired or about to expire, refresh it
      if (timeUntilExpiry < this.refreshBeforeExpiryMs) {
        await this.refreshTokens();
      }
    } catch (error) {
      console.error("TokenManager: Error checking token expiration", error);
    }

    // Schedule the next check
    this.scheduleTokenRefresh();
  }

  /**
   * Refresh the tokens using the refresh token
   */
  async refreshTokens() {
    // If already refreshing, don't start another refresh
    if (this.isRefreshing) {
      return;
    }

    this.isRefreshing = true;

    try {
      console.log("TokenManager: Refreshing tokens");

      const refreshToken = localStorage.getItem("refreshToken");

      if (!refreshToken) {
        throw new Error("No refresh token available");
      }

      // Call the refresh token API
      const response = await authAPI.refreshToken(refreshToken);

      // Update tokens in localStorage
      this.updateTokens(
        response.data.access_token,
        response.data.access_token_expiry_date,
        response.data.refresh_token,
        response.data.refresh_token_expiry_date,
      );

      console.log("TokenManager: Tokens refreshed successfully");
      this.notifyListeners("refresh", true);
    } catch (error) {
      console.error("TokenManager: Error refreshing tokens", error);

      // If refresh fails, redirect to login
      this.clearTokens();
      this.redirectToLogin();
      this.notifyListeners("refresh", false, error);
    } finally {
      this.isRefreshing = false;
    }
  }

  /**
   * Update tokens in localStorage
   * Safely handles different date formats
   */
  updateTokens(
    accessToken,
    accessTokenExpiry,
    refreshToken,
    refreshTokenExpiry,
  ) {
    if (!accessToken || !refreshToken) {
      console.error("TokenManager: Missing token values");
      return;
    }

    // Store the tokens
    localStorage.setItem("accessToken", accessToken);
    localStorage.setItem("refreshToken", refreshToken);

    // Handle the expiry dates safely
    try {
      // Convert string dates to Date objects if they're not already
      let accessExpiryDate = accessTokenExpiry;
      let refreshExpiryDate = refreshTokenExpiry;

      // If it's a string, try to parse it as a date
      if (typeof accessTokenExpiry === "string") {
        // Try various date parsing approaches
        try {
          accessExpiryDate = new Date(accessTokenExpiry);
        } catch (e) {
          console.warn("Failed to parse access token expiry date", e);
          // Default to 5 minutes from now if parsing fails
          accessExpiryDate = new Date(Date.now() + 5 * 60 * 1000);
        }
      }

      if (typeof refreshTokenExpiry === "string") {
        try {
          refreshExpiryDate = new Date(refreshTokenExpiry);
        } catch (e) {
          console.warn("Failed to parse refresh token expiry date", e);
          // Default to 24 hours from now if parsing fails
          refreshExpiryDate = new Date(Date.now() + 24 * 60 * 60 * 1000);
        }
      }

      // Check if dates are valid before storing
      if (isNaN(accessExpiryDate.getTime())) {
        console.warn("Invalid access token expiry date, using default");
        accessExpiryDate = new Date(Date.now() + 5 * 60 * 1000);
      }

      if (isNaN(refreshExpiryDate.getTime())) {
        console.warn("Invalid refresh token expiry date, using default");
        refreshExpiryDate = new Date(Date.now() + 24 * 60 * 60 * 1000);
      }

      // Store the expiry dates as ISO strings
      localStorage.setItem("accessTokenExpiry", accessExpiryDate.toISOString());
      localStorage.setItem(
        "refreshTokenExpiry",
        refreshExpiryDate.toISOString(),
      );

      console.log("TokenManager: Tokens updated successfully", {
        accessTokenExpiry: accessExpiryDate,
        refreshTokenExpiry: refreshExpiryDate,
      });
    } catch (err) {
      console.error("TokenManager: Error updating token expiry dates", err);
      // Still store the tokens even if expiry dates fail
    }
  }

  /**
   * Clear all tokens from localStorage
   */
  clearTokens() {
    localStorage.removeItem("accessToken");
    localStorage.removeItem("accessTokenExpiry");
    localStorage.removeItem("refreshToken");
    localStorage.removeItem("refreshTokenExpiry");
  }

  /**
   * Get the current access token
   * @returns {string|null} The access token or null if not available
   */
  getAccessToken() {
    return localStorage.getItem("accessToken");
  }

  /**
   * Check if the user is logged in
   * @returns {boolean} True if the user is logged in
   */
  isLoggedIn() {
    return !!this.getAccessToken();
  }

  /**
   * Redirect to the login page
   */
  redirectToLogin() {
    // Use window.location for hard redirect to ensure a clean login state
    window.location.href = "/login";
  }

  /**
   * Register a listener for token events
   * @param {function} listener - Callback function for token events
   */
  addListener(listener) {
    this.listeners.push(listener);
    return () => this.removeListener(listener);
  }

  /**
   * Remove a token event listener
   * @param {function} listener - The listener to remove
   */
  removeListener(listener) {
    this.listeners = this.listeners.filter((l) => l !== listener);
  }

  /**
   * Notify all listeners of a token event
   * @param {string} event - The event type
   * @param {boolean} success - Whether the event was successful
   * @param {Error} error - Error object if the event failed
   */
  notifyListeners(event, success, error = null) {
    this.listeners.forEach((listener) => {
      try {
        listener(event, success, error);
      } catch (e) {
        console.error("TokenManager: Error in listener", e);
      }
    });
  }

  /**
   * Clean up the token manager
   * Should be called when the application is unmounted
   */
  cleanup() {
    if (this.refreshTimeoutId) {
      clearTimeout(this.refreshTimeoutId);
      this.refreshTimeoutId = null;
    }
    this.listeners = [];
  }
}

// Create a singleton instance
const tokenManager = new TokenManager();

export default tokenManager;
