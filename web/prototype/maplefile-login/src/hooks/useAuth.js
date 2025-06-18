import { useState, useEffect, useCallback } from "react";
import AuthService from "../services/authService.jsx";
import LocalStorageService from "../services/localStorageService.jsx";
import workerManager from "../services/workerManager.jsx";

// Custom hook for authentication management with background worker
const useAuth = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState(null);
  const [tokenInfo, setTokenInfo] = useState({});
  const [workerStatus, setWorkerStatus] = useState({ isInitialized: false });

  // Update authentication state
  const updateAuthState = useCallback(() => {
    const authenticated = AuthService.isAuthenticated();
    const userEmail = AuthService.getCurrentUserEmail();

    setIsAuthenticated(authenticated);
    setUser(userEmail ? { email: userEmail } : null);

    // Update token information
    const accessToken = LocalStorageService.getAccessToken();
    const refreshToken = LocalStorageService.getRefreshToken();
    const accessTokenExpired = LocalStorageService.isAccessTokenExpired();
    const refreshTokenExpired = LocalStorageService.isRefreshTokenExpired();
    const accessTokenExpiringSoon =
      LocalStorageService.isAccessTokenExpiringSoon(5);

    setTokenInfo({
      hasAccessToken: !!accessToken,
      hasRefreshToken: !!refreshToken,
      accessTokenExpired,
      refreshTokenExpired,
      accessTokenExpiringSoon,
      accessTokenExpiry: localStorage.getItem("mapleapps_access_token_expiry"),
      refreshTokenExpiry: localStorage.getItem(
        "mapleapps_refresh_token_expiry",
      ),
    });
  }, []);

  // Handle worker messages
  const handleWorkerMessage = useCallback(
    (type, data) => {
      console.log(`[useAuth] Received worker message: ${type}`, data);

      switch (type) {
        case "token_refreshed":
        case "token_status_update":
          // Update auth state when tokens change
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
      await AuthService.refreshToken();
      updateAuthState();
      return true;
    } catch (error) {
      console.error("[useAuth] Manual token refresh failed:", error);
      setIsAuthenticated(false);
      setUser(null);
      throw error;
    }
  }, [updateAuthState]);

  // Force token check
  const forceTokenCheck = useCallback(() => {
    AuthService.forceTokenCheck();
  }, []);

  // Logout function
  const logout = useCallback(() => {
    AuthService.logout();
    setIsAuthenticated(false);
    setUser(null);
    setTokenInfo({});
  }, []);

  // Initialize authentication and worker
  useEffect(() => {
    const initAuth = async () => {
      setIsLoading(true);

      try {
        // Initialize the worker
        await AuthService.initializeWorker();

        // Update initial authentication state
        updateAuthState();

        // Get worker status
        const status = await AuthService.getWorkerStatus();
        setWorkerStatus(status);
      } catch (error) {
        console.error("[useAuth] Failed to initialize auth:", error);
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

    // Also listen for our custom storage events
    window.addEventListener("storage", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, [updateAuthState]);

  // API call wrapper with worker-based token management
  const apiCall = useCallback(
    async (apiFunction) => {
      try {
        // The worker handles token refresh automatically,
        // so we just need to make the API call
        return await apiFunction();
      } catch (error) {
        // If the API call fails with auth error, the worker should handle it
        // but we can also try a manual refresh as fallback
        if (
          error.message?.includes("401") ||
          error.message?.includes("Unauthorized")
        ) {
          try {
            await manualRefresh();
            return await apiFunction();
          } catch (refreshError) {
            // If refresh fails, logout user
            logout();
            throw new Error("Session expired. Please log in again.");
          }
        }
        throw error;
      }
    },
    [manualRefresh, logout],
  );

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
    isAccessTokenExpired: tokenInfo.accessTokenExpired,
    isRefreshTokenExpired: tokenInfo.refreshTokenExpired,
    isAccessTokenExpiringSoon: tokenInfo.accessTokenExpiringSoon,
  };
};

export default useAuth;
