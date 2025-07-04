// File: monorepo/web/maplefile-frontend/src/hooks/useAuth.js
// Custom hook for authentication management with automatic token refresh via ApiClient interceptors
import { useState, useEffect, useCallback } from "react";
import AuthService from "../services/AuthService.js";
import LocalStorageService from "../services/LocalStorageService.js";
import WorkerManager from "../services/WorkerManager.js";

// Custom hook for authentication management
const useAuth = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState(null);
  const [tokenInfo, setTokenInfo] = useState({});
  const [authStatus, setAuthStatus] = useState({ isInitialized: false });

  // Update authentication state
  const updateAuthState = useCallback(() => {
    try {
      const authenticated = AuthService.isAuthenticated();
      const userEmail = AuthService.getCurrentUserEmail();

      setIsAuthenticated(authenticated);
      setUser(userEmail ? { email: userEmail } : null);

      // Update token information
      const tokenExpiryInfo = LocalStorageService.getTokenExpiryInfo();
      const hasTokens = !!(
        LocalStorageService.getAccessToken() &&
        LocalStorageService.getRefreshToken()
      );

      setTokenInfo({
        hasTokens,
        tokenSystem: "unencrypted",
        refreshMethod: "api_interceptor",
        ...tokenExpiryInfo,
      });

      console.log("[useAuth] Auth state updated:", {
        authenticated,
        userEmail,
        hasTokens,
        tokenExpiryInfo,
      });
    } catch (error) {
      console.error("[useAuth] Error updating auth state:", error);
      // Don't clear state on error, just log it
    }
  }, []);

  // Handle worker/auth manager events
  const handleAuthMessage = useCallback(
    (type, data) => {
      console.log(`[useAuth] Received auth message: ${type}`, data);

      switch (type) {
        case "token_refresh_success":
          console.log("[useAuth] Tokens refreshed automatically");
          updateAuthState();
          break;

        case "token_refresh_failed":
          console.error("[useAuth] Token refresh failed:", data);
          // Auth state will be updated by the logout handling
          break;

        case "force_logout":
          console.log("[useAuth] Force logout received:", data.reason);
          setIsAuthenticated(false);
          setUser(null);
          setTokenInfo({});
          // Redirect to login
          if (data.shouldRedirect !== false) {
            setTimeout(() => {
              if (window.location.pathname !== "/") {
                window.location.href = "/";
              }
            }, 1000);
          }
          break;

        default:
          break;
      }
    },
    [updateAuthState],
  );

  // Manual token refresh (now delegated to ApiClient)
  const manualRefresh = useCallback(async () => {
    try {
      console.log("[useAuth] Initiating manual token refresh");

      // Check if we have a refresh token
      if (!LocalStorageService.getRefreshToken()) {
        throw new Error("No refresh token available for refresh");
      }

      // Use AuthService which delegates to ApiClient
      await AuthService.refreshToken();
      updateAuthState();
      console.log("[useAuth] Manual token refresh successful");
      return true;
    } catch (error) {
      console.error("[useAuth] Manual token refresh failed:", error);
      setIsAuthenticated(false);
      setUser(null);
      setTokenInfo({});
      throw error;
    }
  }, [updateAuthState]);

  // Force token check (no-op since handled automatically)
  const forceTokenCheck = useCallback(() => {
    console.log(
      "[useAuth] Force token check - handled automatically by ApiClient",
    );
    updateAuthState();
  }, [updateAuthState]);

  // Logout function
  const logout = useCallback(() => {
    console.log("[useAuth] Logging out user");
    AuthService.logout();
    setIsAuthenticated(false);
    setUser(null);
    setTokenInfo({});
  }, []);

  // Check token health and suggest actions
  const getTokenHealth = useCallback(() => {
    const health = {
      status: "unknown",
      recommendations: [],
      canRefresh: false,
      needsReauth: false,
    };

    if (!tokenInfo.hasTokens) {
      health.status = "no_tokens";
      health.recommendations.push("No authentication tokens found");
      health.needsReauth = true;
    } else if (tokenInfo.refreshTokenExpired) {
      health.status = "expired";
      health.recommendations.push(
        "Refresh token expired - re-authentication required",
      );
      health.needsReauth = true;
    } else if (tokenInfo.accessTokenExpired) {
      health.status = "needs_refresh";
      health.recommendations.push(
        "Access token expired - refresh will happen automatically",
      );
      health.canRefresh = true;
    } else if (tokenInfo.accessTokenExpiringSoon) {
      health.status = "expiring_soon";
      health.recommendations.push(
        "Access token expiring soon - refresh will happen automatically",
      );
      health.canRefresh = true;
    } else {
      health.status = "healthy";
      health.recommendations.push("Tokens are valid and healthy");
    }

    return health;
  }, [tokenInfo]);

  // Initialize authentication
  useEffect(() => {
    const initAuth = async () => {
      setIsLoading(true);

      try {
        console.log("[useAuth] Initializing authentication system");

        // Clean up any old encrypted token data
        LocalStorageService.cleanupEncryptedTokenData();

        // Initialize the auth service (simplified)
        await AuthService.initializeWorker();

        // Update initial authentication state
        updateAuthState();

        // Get auth status
        const status = await AuthService.getWorkerStatus();
        setAuthStatus(status);

        console.log("[useAuth] Authentication system initialized successfully");
      } catch (error) {
        console.error("[useAuth] Failed to initialize auth:", error);
        // Set safe defaults on initialization failure
        setIsAuthenticated(false);
        setUser(null);
        setTokenInfo({});
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, [updateAuthState]);

  // Set up auth message listener
  useEffect(() => {
    // Add auth message listener
    WorkerManager.addAuthStateChangeListener(handleAuthMessage);

    // Cleanup listener on unmount
    return () => {
      WorkerManager.removeAuthStateChangeListener(handleAuthMessage);
    };
  }, [handleAuthMessage]);

  // Listen for localStorage changes (cross-tab synchronization)
  useEffect(() => {
    const handleStorageChange = (event) => {
      if (event.key && event.key.startsWith("mapleapps_")) {
        console.log("[useAuth] Storage change detected:", event.key);
        updateAuthState();
      }
    };

    // Listen for storage events from other tabs
    window.addEventListener("storage", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, [updateAuthState]);

  // API call wrapper with automatic token management
  const apiCall = useCallback(
    async (apiFunction) => {
      try {
        // Check token health before making API calls
        const tokenHealth = getTokenHealth();

        if (tokenHealth.needsReauth) {
          logout();
          throw new Error("Authentication required. Please log in again.");
        }

        // No need to manually refresh - ApiClient handles this automatically
        // Make the API call
        return await apiFunction();
      } catch (error) {
        // Handle authentication errors
        if (
          error.message?.includes("401") ||
          error.message?.includes("Unauthorized") ||
          error.message?.includes("expired") ||
          error.message?.includes("Session expired")
        ) {
          console.log("[useAuth] Authentication error detected, logging out");
          logout();
          throw new Error("Session expired. Please log in again.");
        }
        throw error;
      }
    },
    [getTokenHealth, logout],
  );

  // Get debug information
  const getDebugInfo = useCallback(() => {
    return {
      isAuthenticated,
      user,
      tokenInfo,
      authStatus,
      tokenHealth: getTokenHealth(),
      canMakeAuthenticatedRequests: AuthService.canMakeAuthenticatedRequests(),
      storageKeys: {
        hasAccessToken: !!LocalStorageService.getAccessToken(),
        hasRefreshToken: !!LocalStorageService.getRefreshToken(),
        hasUserEmail: !!LocalStorageService.getUserEmail(),
      },
      refreshMethod: "api_interceptor",
    };
  }, [isAuthenticated, user, tokenInfo, authStatus, getTokenHealth]);

  return {
    // State
    isAuthenticated,
    isLoading,
    user,
    tokenInfo,
    authStatus,

    // Actions
    logout,
    manualRefresh,
    forceTokenCheck,
    updateAuthState,
    apiCall,

    // Utilities
    getTokenHealth,
    getDebugInfo,
    isAccessTokenExpired: tokenInfo.accessTokenExpired,
    isRefreshTokenExpired: tokenInfo.refreshTokenExpired,
    isAccessTokenExpiringSoon: tokenInfo.accessTokenExpiringSoon,
    hasTokens: tokenInfo.hasTokens,
    tokenSystem: tokenInfo.tokenSystem || "unencrypted",
    refreshMethod: "api_interceptor",

    // Simplified capabilities
    canMakeAuthenticatedRequests: AuthService.canMakeAuthenticatedRequests(),
  };
};

export default useAuth;
