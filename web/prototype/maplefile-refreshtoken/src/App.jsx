import React, { useState, useEffect } from "react";
import TokenRefreshWorkerManager from "./services/tokenRefreshWorkerManager.js";
import TokenStorageService from "./services/tokenStorageService.js";
import "./App.css";

function App() {
  const [workerStatus, setWorkerStatus] = useState({ isInitialized: false });
  const [tokenInfo, setTokenInfo] = useState({});
  const [isLoading, setIsLoading] = useState(true);
  const [logs, setLogs] = useState([]);
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Add log entry
  const addLog = (message, type = "info") => {
    const timestamp = new Date().toLocaleTimeString();
    setLogs((prev) => [...prev.slice(-19), { timestamp, message, type }]);
  };

  // Initialize worker and token monitoring
  useEffect(() => {
    const initializeSystem = async () => {
      try {
        addLog("Initializing token refresh system...", "info");

        // Initialize the worker
        await TokenRefreshWorkerManager.initialize();
        addLog("Worker initialized successfully", "success");

        // Get initial worker status
        const status = await TokenRefreshWorkerManager.getWorkerStatus();
        setWorkerStatus(status);

        // Get initial token info
        updateTokenInfo();

        // Set up worker event listeners
        TokenRefreshWorkerManager.addStateChangeListener(handleWorkerMessage);

        // Start monitoring if we have tokens
        const hasTokens = TokenStorageService.hasEncryptedTokens();
        if (hasTokens) {
          addLog("Starting token monitoring (tokens found)", "info");
          TokenRefreshWorkerManager.startMonitoring();
        } else {
          addLog("No tokens found - monitoring not started", "warning");
        }

        setIsLoading(false);
      } catch (error) {
        console.error("Failed to initialize system:", error);
        addLog(`Failed to initialize: ${error.message}`, "error");
        setIsLoading(false);
      }

      TokenRefreshWorkerManager.addAuthStateChangeListener((type, data) => {
        console.log(`[App] Worker event: ${type}`, data);

        switch (type) {
          case "token_refreshed":
          case "token_refresh_success":
            checkTokens();
            setWorkerStatus((prev) => ({ ...prev, isRefreshing: false }));
            break;
          case "token_status_update":
            checkTokens();
            if (data && typeof data.isRefreshing !== "undefined") {
              setWorkerStatus((prev) => ({
                ...prev,
                isRefreshing: data.isRefreshing,
              }));
            }
            break;
          case "token_refresh_failed":
            console.error("[App] Token refresh failed:", data);
            checkTokens();
            setWorkerStatus((prev) => ({ ...prev, isRefreshing: false }));
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
    };

    initializeSystem();

    // Cleanup
    return () => {
      TokenRefreshWorkerManager.removeStateChangeListener(handleWorkerMessage);
    };
  }, []);

  // Handle worker messages
  const handleWorkerMessage = (type, data) => {
    console.log(`[App] Received worker message: ${type}`, data);

    switch (type) {
      case "token_refresh_success":
        addLog("✅ Token refresh successful!", "success");
        updateTokenInfo();
        setIsRefreshing(false);
        break;

      case "token_refresh_failed":
        addLog(`❌ Token refresh failed: ${data.error}`, "error");
        updateTokenInfo();
        setIsRefreshing(false);
        break;

      case "token_status_update":
        addLog("📊 Token status updated", "info");
        updateTokenInfo();
        break;

      case "force_logout":
        addLog(`🚪 Force logout: ${data.reason}`, "warning");
        updateTokenInfo();
        break;

      case "worker_error":
        addLog(`❌ Worker error: ${data.error}`, "error");
        break;

      default:
        addLog(`📨 Worker message: ${type}`, "info");
        break;
    }
  };

  // Update token information
  const updateTokenInfo = () => {
    const info = TokenStorageService.getTokenInfo();
    setTokenInfo(info);
    console.log("[App] Token info updated:", info);
  };

  // Test functions
  const createDemoTokens = () => {
    const tokenData = TokenStorageService.createDemoTokens();
    addLog("🎭 Demo tokens created", "success");
    updateTokenInfo();

    // Start monitoring
    TokenRefreshWorkerManager.startMonitoring();
    addLog("🔄 Token monitoring started", "info");
  };

  const createExpiringSoonTokens = () => {
    const tokenData = TokenStorageService.createExpiringSoonTokens();
    addLog(
      "⏰ Expiring soon tokens created (will refresh in ~2 minutes)",
      "warning",
    );
    updateTokenInfo();

    // Start monitoring
    TokenRefreshWorkerManager.startMonitoring();
    addLog("🔄 Token monitoring started", "info");
  };

  const clearAllTokens = () => {
    TokenStorageService.clearAllTokens();
    addLog("🗑️ All tokens cleared", "info");
    updateTokenInfo();

    // Stop monitoring
    TokenRefreshWorkerManager.stopMonitoring();
    addLog("⏹️ Token monitoring stopped", "info");
  };

  const manualRefresh = async () => {
    if (!TokenStorageService.hasEncryptedTokens()) {
      addLog("❌ No tokens to refresh", "error");
      return;
    }

    setIsRefreshing(true);
    addLog("🔄 Starting manual token refresh...", "info");

    try {
      await TokenRefreshWorkerManager.manualRefresh();
      addLog("✅ Manual refresh completed", "success");
    } catch (error) {
      addLog(`❌ Manual refresh failed: ${error.message}`, "error");
    } finally {
      setIsRefreshing(false);
    }
  };

  const forceTokenCheck = () => {
    TokenRefreshWorkerManager.forceTokenCheck();
    addLog("🔍 Force token check triggered", "info");
  };

  const refreshWorkerStatus = async () => {
    try {
      const status = await TokenRefreshWorkerManager.getWorkerStatus();
      setWorkerStatus(status);
      addLog("📊 Worker status refreshed", "info");
    } catch (error) {
      addLog(`❌ Failed to get worker status: ${error.message}`, "error");
    }
  };

  // Format time remaining
  const formatTimeRemaining = (expiryTime) => {
    if (!expiryTime) return "Unknown";

    const now = new Date();
    const expiry = new Date(expiryTime);
    const diff = expiry.getTime() - now.getTime();

    if (diff <= 0) return "Expired";

    const minutes = Math.floor(diff / (1000 * 60));
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ${hours % 24}h`;
    if (hours > 0) return `${hours}h ${minutes % 60}m`;
    return `${minutes}m`;
  };

  if (isLoading) {
    return (
      <div className="app">
        <div className="loading">
          <div className="spinner"></div>
          <p>Initializing Token Refresh System...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="app">
      <header className="header">
        <h1>🔄 MapleApps Token Refresh</h1>
        <p>Background Token Refresh System Prototype</p>
      </header>

      <main className="main">
        <div className="container">
          {/* Worker Status */}
          <section className="section">
            <h2>🛠️ Worker Status</h2>
            <div className="status-grid">
              <div className="status-item">
                <span className="label">Worker Initialized:</span>
                <span
                  className={`status ${workerStatus.isInitialized ? "success" : "error"}`}
                >
                  {workerStatus.isInitialized ? "✅ YES" : "❌ NO"}
                </span>
              </div>
              <div className="status-item">
                <span className="label">Worker Refreshing:</span>
                <span
                  className={`status ${workerStatus.isRefreshing ? "warning" : "success"}`}
                >
                  {workerStatus.isRefreshing ? "🔄 YES" : "⭕ NO"}
                </span>
              </div>
              <div className="status-item">
                <span className="label">Last Check:</span>
                <span className="value">
                  {workerStatus.lastCheck
                    ? new Date(workerStatus.lastCheck).toLocaleTimeString()
                    : "Never"}
                </span>
              </div>
              <div className="status-item">
                <span className="label">Check Interval:</span>
                <span className="value">
                  {workerStatus.checkInterval / 1000}s
                </span>
              </div>
            </div>
            <button onClick={refreshWorkerStatus} className="btn btn-secondary">
              🔄 Refresh Status
            </button>
          </section>

          {/* Token Information */}
          <section className="section">
            <h2>🎫 Token Information</h2>
            <div className="status-grid">
              <div className="status-item">
                <span className="label">Has Encrypted Tokens:</span>
                <span
                  className={`status ${tokenInfo.hasEncryptedTokens ? "success" : "error"}`}
                >
                  {tokenInfo.hasEncryptedTokens ? "✅ YES" : "❌ NO"}
                </span>
              </div>
              <div className="status-item">
                <span className="label">Is Authenticated:</span>
                <span
                  className={`status ${tokenInfo.isAuthenticated ? "success" : "error"}`}
                >
                  {tokenInfo.isAuthenticated ? "✅ YES" : "❌ NO"}
                </span>
              </div>
              <div className="status-item">
                <span className="label">Token Format:</span>
                <span className="value">{tokenInfo.tokenFormat || "none"}</span>
              </div>
              <div className="status-item">
                <span className="label">User Email:</span>
                <span className="value">
                  {tokenInfo.userEmail || "Not set"}
                </span>
              </div>
              <div className="status-item">
                <span className="label">Access Token:</span>
                <span
                  className={`status ${tokenInfo.accessTokenExpired ? "error" : "success"}`}
                >
                  {tokenInfo.accessTokenExpired ? "❌ Expired" : "✅ Valid"}
                  {tokenInfo.accessTokenExpiry && (
                    <span className="time-remaining">
                      ({formatTimeRemaining(tokenInfo.accessTokenExpiry)})
                    </span>
                  )}
                </span>
              </div>
              <div className="status-item">
                <span className="label">Refresh Token:</span>
                <span
                  className={`status ${tokenInfo.refreshTokenExpired ? "error" : "success"}`}
                >
                  {tokenInfo.refreshTokenExpired ? "❌ Expired" : "✅ Valid"}
                  {tokenInfo.refreshTokenExpiry && (
                    <span className="time-remaining">
                      ({formatTimeRemaining(tokenInfo.refreshTokenExpiry)})
                    </span>
                  )}
                </span>
              </div>
            </div>
          </section>

          {/* Controls */}
          <section className="section">
            <h2>🎮 Controls</h2>
            <div className="controls">
              <div className="control-group">
                <h3>Token Management</h3>
                <div className="button-group">
                  <button
                    onClick={createDemoTokens}
                    className="btn btn-primary"
                  >
                    🎭 Create Demo Tokens
                  </button>
                  <button
                    onClick={createExpiringSoonTokens}
                    className="btn btn-warning"
                  >
                    ⏰ Create Expiring Soon Tokens
                  </button>
                  <button onClick={clearAllTokens} className="btn btn-danger">
                    🗑️ Clear All Tokens
                  </button>
                </div>
              </div>

              <div className="control-group">
                <h3>Refresh Operations</h3>
                <div className="button-group">
                  <button
                    onClick={manualRefresh}
                    disabled={isRefreshing || !tokenInfo.hasEncryptedTokens}
                    className="btn btn-primary"
                  >
                    {isRefreshing ? "🔄 Refreshing..." : "🔄 Manual Refresh"}
                  </button>
                  <button
                    onClick={forceTokenCheck}
                    className="btn btn-secondary"
                  >
                    🔍 Force Token Check
                  </button>
                </div>
              </div>
            </div>
          </section>

          {/* Activity Log */}
          <section className="section">
            <h2>📝 Activity Log</h2>
            <div className="log-container">
              {logs.length === 0 ? (
                <div className="log-empty">No activity yet...</div>
              ) : (
                logs.map((log, index) => (
                  <div key={index} className={`log-entry log-${log.type}`}>
                    <span className="log-time">{log.timestamp}</span>
                    <span className="log-message">{log.message}</span>
                  </div>
                ))
              )}
            </div>
          </section>

          {/* Information */}
          <section className="section">
            <h2>ℹ️ How It Works</h2>
            <div className="info-content">
              <div className="info-box">
                <h3>🔄 Automatic Token Refresh</h3>
                <ul>
                  <li>
                    Worker checks tokens every <strong>30 seconds</strong>
                  </li>
                  <li>
                    Refreshes tokens <strong>5 minutes before</strong> they
                    expire
                  </li>
                  <li>Uses encrypted tokens stored in localStorage</li>
                  <li>
                    Supports both separate access/refresh tokens and legacy
                    single token format
                  </li>
                </ul>
              </div>

              <div className="info-box">
                <h3>🛠️ API Endpoint</h3>
                <ul>
                  <li>
                    Endpoint: <code>POST /iam/api/v1/token/refresh</code>
                  </li>
                  <li>
                    Request body:{" "}
                    <code>{'{ "value": "encrypted_refresh_token" }'}</code>
                  </li>
                  <li>Response: New encrypted tokens with expiry times</li>
                </ul>
              </div>

              <div className="info-box">
                <h3>🔒 Security Features</h3>
                <ul>
                  <li>All tokens are stored encrypted</li>
                  <li>Cross-tab communication via BroadcastChannel</li>
                  <li>Automatic cleanup on refresh failure</li>
                  <li>Force logout when refresh token expires</li>
                </ul>
              </div>
            </div>
          </section>
        </div>
      </main>

      <footer className="footer">
        <p>&copy; 2025 MapleApps Token Refresh Prototype</p>
      </footer>
    </div>
  );
}

export default App;
