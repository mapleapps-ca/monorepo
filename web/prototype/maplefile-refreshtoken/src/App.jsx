// src/App.jsx
import React, { useState, useEffect } from "react";
import TokenMonitor from "./components/TokenMonitor";
import TokenDisplay from "./components/TokenDisplay";
import workerManager from "./services/workerManager";
import localStorageService from "./services/localStorageService";
import "./App.css";

function App() {
  const [hasTokens, setHasTokens] = useState(false);
  const [isMonitoring, setIsMonitoring] = useState(false);
  const [tokenInfo, setTokenInfo] = useState({});
  const [workerStatus, setWorkerStatus] = useState({ isInitialized: false });

  useEffect(() => {
    // Check for existing tokens on startup
    checkTokens();

    // Initialize worker manager
    initializeWorker();

    // Set up storage event listener for cross-tab synchronization
    const handleStorageChange = (event) => {
      if (event.key && event.key.startsWith("mapleapps_")) {
        console.log("[App] Storage change detected:", event.key);
        checkTokens();
      }
    };

    window.addEventListener("storage", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
      workerManager.destroy();
    };
  }, []);

  const checkTokens = () => {
    const authenticated = localStorageService.isAuthenticated();
    const accessToken = localStorageService.getAccessToken();
    const refreshToken = localStorageService.getRefreshToken();
    const accessTokenExpired = localStorageService.isAccessTokenExpired();
    const refreshTokenExpired = localStorageService.isRefreshTokenExpired();
    const accessTokenExpiringSoon =
      localStorageService.isAccessTokenExpiringSoon(5);

    setHasTokens(authenticated);
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
      userEmail: localStorageService.getUserEmail(),
    });
  };

  const initializeWorker = async () => {
    try {
      await workerManager.initialize();

      // Add message listener for worker events
      workerManager.addAuthStateChangeListener((type, data) => {
        console.log(`[App] Worker event: ${type}`, data);

        switch (type) {
          case "token_refreshed":
          case "token_status_update":
            checkTokens();
            break;
          case "token_refresh_failed":
            console.error("[App] Token refresh failed:", data);
            checkTokens();
            break;
          case "force_logout":
            console.log("[App] Force logout received:", data.reason);
            setHasTokens(false);
            setIsMonitoring(false);
            break;
          case "worker_error":
            console.error("[App] Worker error:", data);
            break;
        }
      });

      const status = await workerManager.getWorkerStatus();
      setWorkerStatus(status);

      // If we have tokens, start monitoring
      if (localStorageService.isAuthenticated()) {
        startMonitoring();
      }
    } catch (error) {
      console.error("[App] Failed to initialize worker:", error);
    }
  };

  const startMonitoring = () => {
    console.log("[App] Starting token monitoring...");
    workerManager.startMonitoring();
    setIsMonitoring(true);
  };

  const stopMonitoring = () => {
    console.log("[App] Stopping token monitoring...");
    workerManager.stopMonitoring();
    setIsMonitoring(false);
  };

  const clearTokens = () => {
    localStorageService.clearAuthData();
    setHasTokens(false);
    setIsMonitoring(false);
    checkTokens();
  };

  const forceRefresh = async () => {
    try {
      await workerManager.manualRefresh();
      console.log("[App] Manual refresh successful");
    } catch (error) {
      console.error("[App] Manual refresh failed:", error);
    }
  };

  return (
    <div className="app">
      <header className="header">
        <h1>🔄 MapleApps Token Refresh Service</h1>
        <p>Background token monitoring and refresh</p>
      </header>

      <main className="main">
        <div className="container">
          {hasTokens ? (
            <>
              <div className="status-card">
                <h2>✅ Tokens Found</h2>
                <p>Background token refresh service is available</p>
              </div>

              <TokenDisplay tokenInfo={tokenInfo} />

              <TokenMonitor
                isMonitoring={isMonitoring}
                workerStatus={workerStatus}
                onStartMonitoring={startMonitoring}
                onStopMonitoring={stopMonitoring}
                onForceRefresh={forceRefresh}
                onClearTokens={clearTokens}
              />
            </>
          ) : (
            <div className="status-card no-tokens">
              <h2>❌ No Tokens Found</h2>
              <p>No authentication tokens found in localStorage.</p>
              <p>Please log in through the main application first.</p>

              <div className="expected-keys">
                <h3>Expected localStorage keys:</h3>
                <ul>
                  <li>
                    <code>mapleapps_access_token</code>
                  </li>
                  <li>
                    <code>mapleapps_refresh_token</code>
                  </li>
                  <li>
                    <code>mapleapps_access_token_expiry</code>
                  </li>
                  <li>
                    <code>mapleapps_refresh_token_expiry</code>
                  </li>
                  <li>
                    <code>mapleapps_user_email</code>
                  </li>
                </ul>
              </div>
            </div>
          )}
        </div>
      </main>

      <footer className="footer">
        <p>&copy; 2025 MapleApps Token Refresh Service</p>
      </footer>
    </div>
  );
}

export default App;
