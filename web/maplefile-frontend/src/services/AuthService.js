// src/services/AuthService.js (Updated)

/**
 * AuthService - Orchestrates authentication state management
 *
 * This updated service demonstrates how to coordinate multiple specialized services
 * to create a cohesive authentication system. It acts as the main interface between
 * the user interface and the underlying authentication infrastructure.
 *
 * Key responsibilities:
 * 1. Coordinate the three-step login process
 * 2. Manage authentication state and user sessions
 * 3. Handle token lifecycle and automatic refresh
 * 4. Provide a clean API for React components
 * 5. Maintain user data and encryption keys during the session
 */
export class AuthService {
  constructor(logger, loginService, tokenService) {
    this.logger = logger;
    this.loginService = loginService;
    this.tokenService = tokenService;

    // Authentication state management
    this.isAuthenticated = false;
    this.user = null;
    this.userSession = null; // Contains decrypted keys and user data

    // State change listeners for React components
    this.stateChangeListeners = [];

    // Initialize token event handling
    this.setupTokenEventHandling();

    this.logger.log("AuthService: Initialized with login and token services");
  }

  /**
   * Initialize the authentication service
   * This should be called when the application starts
   */
  async initialize() {
    this.logger.log("AuthService: Starting initialization");

    try {
      // Initialize the token service and check for existing authentication
      const hasValidTokens = await this.tokenService.initialize();

      if (hasValidTokens) {
        // If we have valid tokens, restore the authentication state
        // Note: In a real application, you might want to validate the tokens
        // with the server or load user profile information
        this.isAuthenticated = true;
        this.user = this.loadUserFromStorage();

        this.logger.log(
          "AuthService: Restored authentication state from stored tokens",
        );
        this.notifyStateChange();
      } else {
        this.logger.log(
          "AuthService: No valid tokens found, user needs to log in",
        );
      }

      return this.isAuthenticated;
    } catch (error) {
      this.logger.log(
        `AuthService: Error during initialization: ${error.message}`,
      );
      return false;
    }
  }

  /**
   * Setup event handling for token lifecycle events
   * This allows the AuthService to react to token refresh, expiry, etc.
   */
  setupTokenEventHandling() {
    this.tokenService.addListener((event, success, error) => {
      this.logger.log(
        `AuthService: Token event received: ${event}, success: ${success}`,
      );

      switch (event) {
        case "tokensRefreshed":
          if (success) {
            // Tokens were refreshed successfully, nothing special needed
            this.logger.log("AuthService: Tokens refreshed automatically");
          } else {
            // Token refresh failed, user needs to log in again
            this.logger.log(
              "AuthService: Token refresh failed, clearing authentication",
            );
            this.handleAuthenticationFailure();
          }
          break;

        case "authenticationFailed":
          this.handleAuthenticationFailure();
          break;

        case "tokensCleared":
          // Tokens were cleared, ensure our state is also cleared
          this.clearAuthenticationState();
          break;
      }
    });
  }

  /**
   * Step 1: Request OTT (One-Time Token)
   * Initiates the login process by requesting a verification code
   */
  async requestOTT(email) {
    this.logger.log(`AuthService: Requesting OTT for email: ${email}`);

    try {
      const result = await this.loginService.requestOTT(email);

      if (result.success) {
        this.logger.log("AuthService: OTT request successful");
      }

      return result;
    } catch (error) {
      this.logger.log(`AuthService: Error requesting OTT: ${error.message}`);
      throw error;
    }
  }

  /**
   * Step 2: Verify OTT and prepare for password entry
   * Verifies the email token and retrieves encrypted user data
   */
  async verifyOTT(email, ott) {
    this.logger.log(`AuthService: Verifying OTT for email: ${email}`);

    try {
      const result = await this.loginService.verifyOTT(email, ott);

      if (result.success) {
        this.logger.log(
          "AuthService: OTT verification successful, ready for password",
        );
      }

      return result;
    } catch (error) {
      this.logger.log(`AuthService: Error verifying OTT: ${error.message}`);
      throw error;
    }
  }

  /**
   * Step 3: Complete login with password
   * Finalizes the authentication process and establishes the user session
   */
  async completeLogin(password) {
    this.logger.log("AuthService: Completing login process");

    try {
      const result = await this.loginService.completeLogin(password);

      if (result.success) {
        // Store the authentication tokens
        await this.tokenService.storeTokens(result.tokens);

        // Set up the user session with decrypted keys
        this.userSession = result.userSession;
        this.user = {
          email: result.userSession.email,
          name: this.extractNameFromEmail(result.userSession.email),
        };
        this.isAuthenticated = true;

        // Store user data for session restoration
        this.saveUserToStorage(this.user);

        this.logger.log("AuthService: Login completed successfully");
        this.notifyStateChange();
      }

      return result;
    } catch (error) {
      this.logger.log(`AuthService: Error completing login: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get current login state for UI components
   * This helps components understand which step of login the user is on
   */
  getLoginState() {
    return this.loginService.getLoginState();
  }

  /**
   * Check if user is currently authenticated
   */
  isUserAuthenticated() {
    return this.isAuthenticated;
  }

  /**
   * Get current user information
   */
  getCurrentUser() {
    return this.user;
  }

  /**
   * Get user session data (includes decrypted keys)
   * This is used by other services that need access to user's encryption keys
   */
  getUserSession() {
    return this.userSession;
  }

  /**
   * Get authentication headers for API requests
   * This provides authenticated HTTP headers for axios requests
   */
  getAuthHeaders() {
    return this.tokenService.getAuthHeaders();
  }

  /**
   * Log out the user
   * This clears all authentication state and tokens
   */
  async logout() {
    this.logger.log("AuthService: Logging out user");

    try {
      // Clear tokens from storage and stop refresh cycle
      this.tokenService.clearTokens();

      // Clear login state in login service
      this.loginService.clearLoginState();

      // Clear our authentication state
      this.clearAuthenticationState();

      // Clear user data from storage
      this.clearUserFromStorage();

      this.logger.log("AuthService: Logout completed");
      this.notifyStateChange();
    } catch (error) {
      this.logger.log(`AuthService: Error during logout: ${error.message}`);
      // Even if there's an error, we should clear local state
      this.clearAuthenticationState();
      this.notifyStateChange();
    }
  }

  /**
   * Handle authentication failure
   * This is called when tokens can't be refreshed or other auth errors occur
   */
  handleAuthenticationFailure() {
    this.logger.log("AuthService: Handling authentication failure");

    // Clear all authentication state
    this.clearAuthenticationState();
    this.clearUserFromStorage();

    // Notify listeners that authentication failed
    this.notifyStateChange();

    // Optionally, you could emit a specific event for authentication failure
    // that components could listen to for showing error messages
  }

  /**
   * Clear authentication state
   * This resets all in-memory authentication data
   */
  clearAuthenticationState() {
    this.isAuthenticated = false;
    this.user = null;
    this.userSession = null;

    // Clear any sensitive data from memory
    if (this.userSession) {
      // In a real application, you might want to securely wipe memory
      this.userSession = null;
    }
  }

  /**
   * Add a listener for authentication state changes
   * This allows React components to react to login/logout events
   */
  addStateChangeListener(listener) {
    this.stateChangeListeners.push(listener);

    // Return a function to remove the listener
    return () => this.removeStateChangeListener(listener);
  }

  /**
   * Remove a state change listener
   */
  removeStateChangeListener(listener) {
    this.stateChangeListeners = this.stateChangeListeners.filter(
      (l) => l !== listener,
    );
  }

  /**
   * Notify all listeners of authentication state changes
   * This implements the observer pattern for state management
   */
  notifyStateChange() {
    const currentState = {
      isAuthenticated: this.isAuthenticated,
      user: this.user,
      loginState: this.getLoginState(),
    };

    this.stateChangeListeners.forEach((listener) => {
      try {
        listener(currentState);
      } catch (error) {
        this.logger.log(
          `AuthService: Error in state change listener: ${error.message}`,
        );
      }
    });
  }

  /**
   * Save user data to storage for session restoration
   * This allows the app to remember the user between page refreshes
   */
  saveUserToStorage(user) {
    try {
      localStorage.setItem("userData", JSON.stringify(user));
    } catch (error) {
      this.logger.log(`AuthService: Error saving user data: ${error.message}`);
    }
  }

  /**
   * Load user data from storage
   * This is used during app initialization to restore user state
   */
  loadUserFromStorage() {
    try {
      const userData = localStorage.getItem("userData");
      if (userData) {
        return JSON.parse(userData);
      }
    } catch (error) {
      this.logger.log(`AuthService: Error loading user data: ${error.message}`);
    }
    return null;
  }

  /**
   * Clear user data from storage
   */
  clearUserFromStorage() {
    try {
      localStorage.removeItem("userData");
    } catch (error) {
      this.logger.log(
        `AuthService: Error clearing user data: ${error.message}`,
      );
    }
  }

  /**
   * Extract a display name from email address
   * This is a simple utility for creating user-friendly names
   */
  extractNameFromEmail(email) {
    if (!email) return "User";

    const localPart = email.split("@")[0];

    // Convert common email patterns to readable names
    return localPart
      .replace(/[._-]/g, " ")
      .replace(/\b\w/g, (char) => char.toUpperCase());
  }

  /**
   * Get token information for debugging
   * This is useful during development and troubleshooting
   */
  getTokenInfo() {
    return this.tokenService.getTokenInfo();
  }

  /**
   * Cleanup the authentication service
   * This should be called when the application is shutting down
   */
  cleanup() {
    this.logger.log("AuthService: Cleaning up");

    // Cleanup token service
    this.tokenService.cleanup();

    // Clear listeners
    this.stateChangeListeners = [];

    // Clear state
    this.clearAuthenticationState();
  }
}
