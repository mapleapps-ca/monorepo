// File: monorepo/web/maplefile-frontend/src/hooks/useAuth.js
// Authentication hook that integrates with ServiceContext
import { useState, useEffect, useCallback } from "react";
import { useServices } from "./useService.jsx";

/**
 * Hook for authentication operations
 * @returns {Object} Authentication API
 */
const useAuth = () => {
  const { authManager, tokenManager, meManager } = useServices();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [user, setUser] = useState(null);
  const [authStatus, setAuthStatus] = useState({});

  // Load user data
  const loadUser = useCallback(async () => {
    if (!authManager || !meManager) return;

    try {
      if (authManager.isAuthenticated()) {
        // Try to get user data from meManager
        const userData = await meManager.getCurrentUser();
        setUser(userData);
        console.log("[useAuth] User data loaded:", userData);
      } else {
        setUser(null);
      }
    } catch (err) {
      console.error("[useAuth] Failed to load user:", err);
      setUser(null);
    }
  }, [authManager, meManager]);

  // Load auth status
  const loadAuthStatus = useCallback(() => {
    if (!authManager || !tokenManager) return;

    try {
      const status = {
        isAuthenticated: authManager.isAuthenticated(),
        canMakeRequests: authManager.canMakeAuthenticatedRequests(),
        tokenHealth: tokenManager.getTokenHealth(),
        userEmail: authManager.getCurrentUserEmail(),
      };

      setAuthStatus(status);
      console.log("[useAuth] Auth status loaded:", status);
    } catch (err) {
      console.error("[useAuth] Failed to load auth status:", err);
    }
  }, [authManager, tokenManager]);

  // Check if user is authenticated
  const isAuthenticated = useCallback(() => {
    return authManager?.isAuthenticated() || false;
  }, [authManager]);

  // Logout user
  const logout = useCallback(async () => {
    if (!authManager) {
      console.error("[useAuth] AuthManager not available for logout");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      console.log("[useAuth] Starting logout process");

      // Clear authentication data
      await authManager.logout();

      // Clear user state
      setUser(null);
      setAuthStatus({});

      console.log("[useAuth] Logout completed successfully");
    } catch (err) {
      console.error("[useAuth] Logout failed:", err);
      setError(`Logout failed: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  }, [authManager]);

  // Refresh tokens
  const refreshTokens = useCallback(async () => {
    if (!tokenManager) {
      throw new Error("TokenManager not available");
    }

    setIsLoading(true);
    setError(null);

    try {
      console.log("[useAuth] Refreshing tokens");

      const result = await tokenManager.refreshTokens();

      // Reload auth status
      loadAuthStatus();

      console.log("[useAuth] Tokens refreshed successfully");
      return result;
    } catch (err) {
      console.error("[useAuth] Token refresh failed:", err);
      setError(`Token refresh failed: ${err.message}`);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [tokenManager, loadAuthStatus]);

  // Get current user email
  const getCurrentUserEmail = useCallback(() => {
    return authManager?.getCurrentUserEmail() || user?.email || null;
  }, [authManager, user]);

  // Check if user can make authenticated requests
  const canMakeAuthenticatedRequests = useCallback(() => {
    return authManager?.canMakeAuthenticatedRequests() || false;
  }, [authManager]);

  // Get token information
  const getTokenInfo = useCallback(() => {
    if (!tokenManager) return null;

    return tokenManager.getTokenExpiryInfo();
  }, [tokenManager]);

  // Get token health
  const getTokenHealth = useCallback(() => {
    if (!tokenManager) return null;

    return tokenManager.getTokenHealth();
  }, [tokenManager]);

  // Clear error
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  // Force reload all auth data
  const reloadAuthData = useCallback(async () => {
    loadAuthStatus();
    await loadUser();
  }, [loadAuthStatus, loadUser]);

  // Load data on mount and when managers change
  useEffect(() => {
    if (authManager && tokenManager && meManager) {
      loadAuthStatus();
      loadUser();
    }
  }, [authManager, tokenManager, meManager, loadAuthStatus, loadUser]);

  // Set up event listeners for auth state changes
  useEffect(() => {
    if (!authManager) return;

    const handleAuthStateChange = (eventType, eventData) => {
      console.log("[useAuth] Auth state change:", eventType, eventData);

      // Reload data on relevant events
      if (
        ["tokens_updated", "token_refresh_success", "login_completed"].includes(
          eventType,
        )
      ) {
        reloadAuthData();
      } else if (
        ["tokens_cleared", "logout_completed", "force_logout"].includes(
          eventType,
        )
      ) {
        setUser(null);
        setAuthStatus({});
      }
    };

    // Add listener if the method exists
    if (authManager.addAuthStateChangeListener) {
      authManager.addAuthStateChangeListener(handleAuthStateChange);

      return () => {
        if (authManager.removeAuthStateChangeListener) {
          authManager.removeAuthStateChangeListener(handleAuthStateChange);
        }
      };
    }
  }, [authManager, reloadAuthData]);

  return {
    // State
    isLoading,
    error,
    user,
    authStatus,

    // Core operations
    logout,
    refreshTokens,

    // Status checks
    isAuthenticated: isAuthenticated(),
    canMakeAuthenticatedRequests: canMakeAuthenticatedRequests(),

    // User information
    getCurrentUserEmail,
    userEmail: getCurrentUserEmail(),

    // Token information
    getTokenInfo,
    getTokenHealth,
    tokenInfo: getTokenInfo(),
    tokenHealth: getTokenHealth(),

    // Utility operations
    loadUser,
    loadAuthStatus,
    reloadAuthData,
    clearError,

    // Managers (for advanced usage)
    authManager,
    tokenManager,
    meManager,
  };
};

export default useAuth;
