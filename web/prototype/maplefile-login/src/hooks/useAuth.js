// monorepo/web/prototype/maplefile-login/src/hooks/useAuth.jsx
import { useState, useEffect, useCallback } from "react";
import AuthService from "../services/authService.jsx";
import LocalStorageService from "../services/localStorageService.jsx";
import workerManager from "../services/workerManager.jsx";

// Custom hook for authentication management with encrypted token system
const useAuth = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState(null);
  const [tokenInfo, setTokenInfo] = useState({});
  const [workerStatus, setWorkerStatus] = useState({ isInitialized: false });

  // Update authentication state for encrypted token system
  const updateAuthState = useCallback(() => {
    const authenticated = AuthService.isAuthenticated();
    const userEmail = AuthService.getCurrentUserEmail();

    setIsAuthenticated(authenticated);
    setUser(userEmail ? { email: userEmail } : null);

    // Update token information for encrypted tokens
    const encryptedTokens = LocalStorageService.getEncryptedTokens();
    const tokenNonce = LocalStorageService.getTokenNonce();
    const tokenExpiryInfo = LocalStorageService.getTokenExpiryInfo();

    setTokenInfo({
      hasEncryptedTokens: !!(encryptedTokens && tokenNonce),
      tokenSystem: "encrypted",
      ...tokenExpiryInfo,
      // Legacy support during transition
      hasLegacyTokens: !!(
        localStorage.getItem("mapleapps_access_token") ||
        localStorage.getItem("mapleapps_refresh_token")
      ),
    });

    console.log("[useAuth] Auth state updated:", {
      authenticated,
      userEmail,
      hasEncryptedTokens: !!(encryptedTokens && tokenNonce),
      tokenExpiryInfo,
    });
  }, []);

  // Handle worker messages
  const handleWorkerMessage = useCallback(
    (type, data) => {
      console.log(`[useAuth] Received worker message: ${type}`, data);

      switch (type) {
        case "token_refreshed":
        case "token_refresh_success":
        case "token_status_update":
          // Update auth state when tokens change
          console.log("[useAuth] Tokens updated, refreshing auth state");
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
          // Worker manager will handle redirect
          break;

        case "legacy_tokens_migrated":
          console.log("[useAuth] Legacy tokens migrated, updating state");
          updateAuthState();
          if (data.shouldRedirect) {
            // User needs to re-authenticate
            setIsAuthenticated(false);
            setUser(null);
            setTokenInfo({});
          }
          break;

        case "worker_error":
          console.error("[useAuth] Worker error:", data);
          break;

        default:
          break;
      }
    },
    [updateAuthState],
  );

  // Manual token refresh via worker
  const manualRefresh = useCallback(async () => {
    try {
      console.log("[useAuth] Initiating manual token refresh");

      // Check if we have encrypted tokens to refresh
      if (!LocalStorageService.getEncryptedTokens()) {
        throw new Error("No encrypted tokens available for refresh");
      }

      await AuthService.refreshTokenViaWorker();
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

  // Force token check
  const forceTokenCheck = useCallback(() => {
    console.log("[useAuth] Forcing token check");
    AuthService.forceTokenCheck();
  }, []);

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

    if (!tokenInfo.hasEncryptedTokens && tokenInfo.hasLegacyTokens) {
      health.status = "legacy_migration_needed";
      health.recommendations.push(
        "Migrate to encrypted token system by re-authenticating",
      );
      health.needsReauth = true;
    } else if (tokenInfo.hasEncryptedTokens) {
      if (tokenInfo.refreshTokenExpired) {
        health.status = "expired";
        health.recommendations.push(
          "Refresh token expired - re-authentication required",
        );
        health.needsReauth = true;
      } else if (tokenInfo.accessTokenExpired) {
        health.status = "needs_refresh";
        health.recommendations.push(
          "Access token expired - refresh recommended",
        );
        health.canRefresh = true;
      } else if (tokenInfo.accessTokenExpiringSoon) {
        health.status = "expiring_soon";
        health.recommendations.push(
          "Access token expiring soon - refresh recommended",
        );
        health.canRefresh = true;
      } else {
        health.status = "healthy";
        health.recommendations.push("Tokens are valid and healthy");
      }
    } else {
      health.status = "no_tokens";
      health.recommendations.push("No authentication tokens found");
      health.needsReauth = true;
    }

    return health;
  }, [tokenInfo]);

  // Initialize authentication and worker
  useEffect(() => {
    const initAuth = async () => {
      setIsLoading(true);

      try {
        console.log("[useAuth] Initializing authentication system");

        // Check for legacy token migration
        const migrated = LocalStorageService.migrateLegacyTokens();
        if (migrated) {
          console.log(
            "[useAuth] Legacy tokens migrated, user needs to re-authenticate",
          );
        }

        // Initialize the worker
        await AuthService.initializeWorker();

        // Update initial authentication state
        updateAuthState();

        // Get worker status
        const status = await AuthService.getWorkerStatus();
        setWorkerStatus(status);

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

  // Set up worker message listener
  useEffect(() => {
    // Add worker message listener
    workerManager.addAuthStateChangeListener(handleWorkerMessage);

    // Cleanup listener on unmount
    return () => {
      workerManager.removeAuthStateChangeListener(handleWorkerMessage);
    };
  }, [handleWorkerMessage]);

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

  // API call wrapper with encrypted token management
  const apiCall = useCallback(
    async (apiFunction) => {
      try {
        // Check token health before making API calls
        const tokenHealth = getTokenHealth();

        if (tokenHealth.needsReauth) {
          logout();
          throw new Error("Authentication required. Please log in again.");
        }

        if (tokenHealth.canRefresh && tokenHealth.status === "needs_refresh") {
          console.log(
            "[useAuth] Auto-refreshing expired tokens before API call",
          );
          try {
            await manualRefresh();
          } catch (refreshError) {
            console.error("[useAuth] Auto-refresh failed:", refreshError);
            logout();
            throw new Error("Session expired. Please log in again.");
          }
        }

        // Make the API call
        return await apiFunction();
      } catch (error) {
        // Handle authentication errors
        if (
          error.message?.includes("401") ||
          error.message?.includes("Unauthorized") ||
          error.message?.includes("expired")
        ) {
          try {
            console.log(
              "[useAuth] API call failed with auth error, attempting refresh",
            );
            await manualRefresh();
            return await apiFunction();
          } catch (refreshError) {
            console.error(
              "[useAuth] Refresh after API error failed:",
              refreshError,
            );
            logout();
            throw new Error("Session expired. Please log in again.");
          }
        }
        throw error;
      }
    },
    [getTokenHealth, logout, manualRefresh],
  );

  // Get debug information
  const getDebugInfo = useCallback(() => {
    return {
      isAuthenticated,
      user,
      tokenInfo,
      workerStatus,
      tokenHealth: getTokenHealth(),
      storageKeys: {
        hasEncryptedTokens: !!LocalStorageService.getEncryptedTokens(),
        hasTokenNonce: !!LocalStorageService.getTokenNonce(),
        hasUserEmail: !!LocalStorageService.getUserEmail(),
      },
    };
  }, [isAuthenticated, user, tokenInfo, workerStatus, getTokenHealth]);

  return {
    // State
    isAuthenticated,
    isLoading,
    user,
    tokenInfo,
    workerStatus,

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
    hasEncryptedTokens: tokenInfo.hasEncryptedTokens,
    hasLegacyTokens: tokenInfo.hasLegacyTokens,
    tokenSystem: tokenInfo.tokenSystem || "encrypted",
  };
};

export default useAuth;
