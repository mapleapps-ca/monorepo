import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import Layout from "../components/Layout.jsx";
import AuthService from "../services/authService.jsx";
import LocalStorageService from "../services/localStorageService.jsx";
import useAuth from "../hooks/useAuth.js";

const Dashboard = () => {
  const navigate = useNavigate();
  const {
    user,
    tokenInfo,
    workerStatus,
    logout,
    manualRefresh,
    forceTokenCheck,
    updateAuthState,
    isAuthenticated,
  } = useAuth();

  const [refreshing, setRefreshing] = useState(false);
  const [workerDetails, setWorkerDetails] = useState(null);

  useEffect(() => {
    // Redirect if not authenticated
    if (!isAuthenticated) {
      navigate("/");
      return;
    }
  }, [isAuthenticated, navigate]);

  // Get detailed worker status
  const updateWorkerStatus = async () => {
    try {
      const status = await AuthService.getWorkerStatus();
      setWorkerDetails(status);
    } catch (error) {
      console.error("Failed to get worker status:", error);
    }
  };

  useEffect(() => {
    updateWorkerStatus();
    // Update worker status periodically
    const interval = setInterval(updateWorkerStatus, 10000);
    return () => clearInterval(interval);
  }, []);

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  const handleClearStorage = () => {
    LocalStorageService.clearAuthData();
    navigate("/");
  };

  const handleManualRefresh = async () => {
    setRefreshing(true);
    try {
      await manualRefresh();
      alert("Tokens refreshed successfully via background worker!");
    } catch (error) {
      alert(`Refresh failed: ${error.message}`);
      // If refresh fails completely, redirect to login
      navigate("/");
    } finally {
      setRefreshing(false);
    }
  };

  const handleForceTokenCheck = () => {
    forceTokenCheck();
    alert("Token check triggered in background worker");
  };

  const handleUpdateStatus = () => {
    updateAuthState();
    updateWorkerStatus();
  };

  if (!user) {
    return (
      <Layout title="Loading...">
        <div className="loading-container">
          <p>Loading dashboard...</p>
        </div>
      </Layout>
    );
  }

  return (
    <Layout
      title="Dashboard"
      subtitle="Welcome to your secure dashboard with background worker"
    >
      <div className="dashboard-container">
        <div className="welcome-section">
          <h3>Welcome back!</h3>
          <p>
            You have successfully completed the 3-step secure login process.
          </p>
          <p>
            <strong>Logged in as:</strong> {user.email}
          </p>
          <p className="worker-status">
            <strong>Background Worker:</strong>
            <span
              className={`status ${workerStatus.isInitialized ? "success" : "error"}`}
            >
              {workerStatus.isInitialized ? "✓ Active" : "✗ Inactive"}
            </span>
          </p>
        </div>

        <div className="token-info">
          <h3>Authentication Status</h3>
          <div className="info-grid">
            <div className="info-item">
              <span className="label">Access Token:</span>
              <span
                className={`status ${tokenInfo.hasAccessToken ? "success" : "error"}`}
              >
                {tokenInfo.hasAccessToken ? "✓ Present" : "✗ Missing"}
              </span>
              {tokenInfo.accessTokenExpired && (
                <span className="error">✗ Expired</span>
              )}
              {!tokenInfo.accessTokenExpired &&
                tokenInfo.accessTokenExpiringSoon && (
                  <span className="warning">⚠️ Expires Soon</span>
                )}
            </div>

            <div className="info-item">
              <span className="label">Refresh Token:</span>
              <span
                className={`status ${tokenInfo.hasRefreshToken ? "success" : "error"}`}
              >
                {tokenInfo.hasRefreshToken ? "✓ Present" : "✗ Missing"}
              </span>
              {tokenInfo.refreshTokenExpired && (
                <span className="error">✗ Expired</span>
              )}
            </div>

            <div className="info-item">
              <span className="label">Background Monitor:</span>
              <span className="status success">
                {workerStatus.isInitialized
                  ? "✓ Monitoring Active"
                  : "⚠️ Not Monitoring"}
              </span>
            </div>
          </div>
        </div>

        <div className="worker-info">
          <h3>Background Worker Status</h3>
          <div className="worker-details">
            <div className="worker-stat">
              <span className="label">Worker Initialized:</span>
              <span
                className={workerStatus.isInitialized ? "success" : "error"}
              >
                {workerStatus.isInitialized ? "Yes" : "No"}
              </span>
            </div>

            {workerDetails && (
              <>
                <div className="worker-stat">
                  <span className="label">Currently Refreshing:</span>
                  <span
                    className={
                      workerDetails.isRefreshing ? "warning" : "success"
                    }
                  >
                    {workerDetails.isRefreshing ? "Yes" : "No"}
                  </span>
                </div>

                <div className="worker-stat">
                  <span className="label">Worker Authenticated:</span>
                  <span
                    className={
                      workerDetails.isAuthenticated ? "success" : "error"
                    }
                  >
                    {workerDetails.isAuthenticated ? "Yes" : "No"}
                  </span>
                </div>

                {workerDetails.lastCheck && (
                  <div className="worker-stat">
                    <span className="label">Last Check:</span>
                    <span>
                      {new Date(workerDetails.lastCheck).toLocaleString()}
                    </span>
                  </div>
                )}
              </>
            )}
          </div>
        </div>

        <div className="token-details">
          <h3>Token Details</h3>
          <div className="token-display">
            {tokenInfo.accessToken && (
              <div className="token-item">
                <label>Access Token (truncated):</label>
                <code>
                  {LocalStorageService.getAccessToken()?.substring(0, 20)}...
                </code>
                {tokenInfo.accessTokenExpiry && (
                  <small>
                    Expires:{" "}
                    {new Date(tokenInfo.accessTokenExpiry).toLocaleString()}
                  </small>
                )}
              </div>
            )}

            {tokenInfo.refreshToken && (
              <div className="token-item">
                <label>Refresh Token (truncated):</label>
                <code>
                  {LocalStorageService.getRefreshToken()?.substring(0, 20)}...
                </code>
                {tokenInfo.refreshTokenExpiry && (
                  <small>
                    Expires:{" "}
                    {new Date(tokenInfo.refreshTokenExpiry).toLocaleString()}
                  </small>
                )}
              </div>
            )}
          </div>
        </div>

        <div className="actions">
          <h3>Background Worker Controls</h3>
          <div className="action-buttons">
            <button
              onClick={handleManualRefresh}
              className={`btn btn-primary ${refreshing ? "loading" : ""}`}
              disabled={refreshing || !tokenInfo.hasRefreshToken}
            >
              {refreshing ? "Refreshing..." : "Manual Refresh (Worker)"}
            </button>

            <button
              onClick={handleForceTokenCheck}
              className="btn btn-secondary"
              disabled={!workerStatus.isInitialized}
            >
              Force Token Check
            </button>

            <button onClick={handleUpdateStatus} className="btn btn-secondary">
              Update Status
            </button>
          </div>
        </div>

        <div className="actions">
          <h3>Session Management</h3>
          <div className="action-buttons">
            <button onClick={handleLogout} className="btn btn-primary">
              Logout (Stop Worker)
            </button>

            <button onClick={handleClearStorage} className="btn btn-danger">
              Clear All Storage
            </button>

            <button
              onClick={() => window.location.reload()}
              className="btn btn-secondary"
            >
              Refresh Page
            </button>
          </div>
        </div>

        <div className="security-summary">
          <h3>Authentication Flow Summary</h3>
          <div className="process-steps">
            <div className="step completed">
              <span className="step-number">1</span>
              <span className="step-title">Email Verification</span>
              <span className="step-status">✓ Completed</span>
            </div>

            <div className="step completed">
              <span className="step-number">2</span>
              <span className="step-title">OTT Verification</span>
              <span className="step-status">✓ Completed</span>
            </div>

            <div className="step completed">
              <span className="step-number">3</span>
              <span className="step-title">Challenge Decryption</span>
              <span className="step-status">✓ Completed</span>
            </div>

            <div className="step completed">
              <span className="step-number">4</span>
              <span className="step-title">Background Monitoring</span>
              <span className="step-status">
                {workerStatus.isInitialized
                  ? "✓ Worker Active"
                  : "⚠️ Worker Inactive"}
              </span>
            </div>
          </div>
        </div>

        <div className="api-info">
          <h3>Background Worker Features</h3>
          <div className="api-details">
            <p>
              <strong>Automatic Monitoring:</strong> Checks tokens every 30
              seconds
            </p>
            <p>
              <strong>Proactive Refresh:</strong> Refreshes 5 minutes before
              expiry
            </p>
            <p>
              <strong>Cross-Tab Sync:</strong> Works across multiple browser
              tabs
            </p>
            <p>
              <strong>Auto-Logout:</strong> Redirects to login when refresh
              token expires
            </p>
            <p>
              <strong>Error Recovery:</strong> Handles network failures and API
              errors
            </p>
            <p>
              <strong>Background Processing:</strong> No UI blocking during
              token operations
            </p>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default Dashboard;
