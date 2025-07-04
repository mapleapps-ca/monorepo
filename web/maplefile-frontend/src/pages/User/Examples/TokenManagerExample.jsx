// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/TokenManagerExample.jsx
// Example component demonstrating how to use the TokenManager

import React, { useState, useEffect } from "react";
import { useServices } from "../../../hooks/useService.jsx";

const TokenManagerExample = () => {
  const { tokenManager } = useServices();
  const [tokenInfo, setTokenInfo] = useState({});
  const [tokenHealth, setTokenHealth] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [managerStatus, setManagerStatus] = useState({});
  const [refreshLog, setRefreshLog] = useState([]);

  // Load token information on mount
  useEffect(() => {
    loadTokenInfo();

    // Add auth state listener
    const handleAuthStateChange = (eventType, eventData) => {
      console.log(
        "[TokenManagerExample] Auth state change:",
        eventType,
        eventData,
      );
      setRefreshLog((prev) => [
        ...prev,
        {
          timestamp: new Date().toISOString(),
          eventType,
          eventData,
        },
      ]);

      // Reload token info on relevant events
      if (
        ["tokens_updated", "token_refresh_success", "tokens_cleared"].includes(
          eventType,
        )
      ) {
        loadTokenInfo();
      }
    };

    tokenManager.addAuthStateChangeListener(handleAuthStateChange);

    // Cleanup
    return () => {
      tokenManager.removeAuthStateChangeListener(handleAuthStateChange);
    };
  }, [tokenManager]);

  // Load token information
  const loadTokenInfo = () => {
    try {
      setError(null);

      // Get token expiry info
      const expiryInfo = tokenManager.getTokenExpiryInfo();
      setTokenInfo(expiryInfo);

      // Get token health
      const health = tokenManager.getTokenHealth();
      setTokenHealth(health);

      // Get manager status
      const status = tokenManager.getManagerStatus();
      setManagerStatus(status);

      console.log("[TokenManagerExample] Token info loaded:", {
        expiryInfo,
        health,
        status,
      });
    } catch (err) {
      setError(err.message);
      console.error("[TokenManagerExample] Failed to load token info:", err);
    }
  };

  // Manual token refresh
  const handleManualRefresh = async () => {
    setLoading(true);
    setError(null);

    try {
      console.log("[TokenManagerExample] Initiating manual token refresh...");

      const response = await tokenManager.refreshTokens();

      console.log("[TokenManagerExample] Token refresh successful:", response);

      // Reload token info
      loadTokenInfo();

      setRefreshLog((prev) => [
        ...prev,
        {
          timestamp: new Date().toISOString(),
          eventType: "manual_refresh_success",
          eventData: { success: true },
        },
      ]);
    } catch (err) {
      setError(err.message);
      console.error("[TokenManagerExample] Token refresh failed:", err);

      setRefreshLog((prev) => [
        ...prev,
        {
          timestamp: new Date().toISOString(),
          eventType: "manual_refresh_failed",
          eventData: { error: err.message },
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  // Force token check
  const handleForceCheck = () => {
    try {
      setError(null);
      console.log("[TokenManagerExample] Forcing token check...");

      const health = tokenManager.forceTokenCheck();
      setTokenHealth(health);

      console.log("[TokenManagerExample] Token check result:", health);
    } catch (err) {
      setError(err.message);
      console.error("[TokenManagerExample] Token check failed:", err);
    }
  };

  // Clear tokens
  const handleClearTokens = () => {
    try {
      setError(null);
      console.log("[TokenManagerExample] Clearing tokens...");

      tokenManager.clearTokens();

      console.log("[TokenManagerExample] Tokens cleared");
      loadTokenInfo();
    } catch (err) {
      setError(err.message);
      console.error("[TokenManagerExample] Failed to clear tokens:", err);
    }
  };

  // Clear refresh log
  const handleClearLog = () => {
    setRefreshLog([]);
  };

  // Format date for display
  const formatDate = (dateString) => {
    if (!dateString) return "N/A";
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  // Format time until expiry
  const getTimeUntilExpiry = (expiryDate) => {
    if (!expiryDate) return "N/A";
    const now = new Date();
    const expiry = new Date(expiryDate);
    const diff = expiry - now;

    if (diff <= 0) return "Expired";

    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days} days`;
    if (hours > 0) return `${hours} hours`;
    return `${minutes} minutes`;
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>🔑 Token Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the <strong>TokenManager</strong> with
        orchestrated API and Storage services.
      </p>

      {/* Manager Status */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#f8f9fa",
          borderRadius: "6px",
          border: "1px solid #dee2e6",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>📊 Manager Status:</h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: "10px",
          }}
        >
          <div>
            <strong>Authenticated:</strong>{" "}
            {managerStatus.isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can Make Requests:</strong>{" "}
            {managerStatus.canMakeRequests ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Manager Loading:</strong>{" "}
            {managerStatus.isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Active Refresh:</strong>{" "}
            {managerStatus.hasActiveRefresh ? "🔄 Yes" : "❌ No"}
          </div>
          <div>
            <strong>Refresh Method:</strong>{" "}
            {managerStatus.refreshMethod || "Unknown"}
          </div>
          <div>
            <strong>Listeners:</strong> {managerStatus.listenerCount || 0}
          </div>
        </div>
      </div>

      {/* Token Information */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>🎫 Token Information:</h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(300px, 1fr))",
            gap: "15px",
          }}
        >
          <div>
            <h5 style={{ margin: "0 0 5px 0" }}>Access Token:</h5>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Exists:</strong>{" "}
              {tokenManager.getAccessToken() ? "✅ Yes" : "❌ No"}
            </p>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Expired:</strong>{" "}
              {tokenInfo.accessTokenExpired ? "❌ Yes" : "✅ No"}
            </p>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Expiring Soon:</strong>{" "}
              {tokenInfo.accessTokenExpiringSoon ? "⚠️ Yes" : "✅ No"}
            </p>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Expires:</strong>{" "}
              {formatDate(tokenInfo.accessTokenExpiry)}
            </p>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Time Until Expiry:</strong>{" "}
              {getTimeUntilExpiry(tokenInfo.accessTokenExpiry)}
            </p>
          </div>
          <div>
            <h5 style={{ margin: "0 0 5px 0" }}>Refresh Token:</h5>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Exists:</strong>{" "}
              {tokenManager.getRefreshToken() ? "✅ Yes" : "❌ No"}
            </p>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Expired:</strong>{" "}
              {tokenInfo.refreshTokenExpired ? "❌ Yes" : "✅ No"}
            </p>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Expires:</strong>{" "}
              {formatDate(tokenInfo.refreshTokenExpiry)}
            </p>
            <p style={{ margin: "2px 0", fontSize: "14px" }}>
              <strong>Time Until Expiry:</strong>{" "}
              {getTimeUntilExpiry(tokenInfo.refreshTokenExpiry)}
            </p>
          </div>
        </div>
      </div>

      {/* Token Health */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor:
            tokenHealth.status === "healthy"
              ? "#d4edda"
              : tokenHealth.status === "needs_refresh" ||
                  tokenHealth.status === "expiring_soon"
                ? "#fff3cd"
                : "#f8d7da",
          borderRadius: "6px",
          border: `1px solid ${
            tokenHealth.status === "healthy"
              ? "#c3e6cb"
              : tokenHealth.status === "needs_refresh" ||
                  tokenHealth.status === "expiring_soon"
                ? "#ffeaa7"
                : "#f5c6cb"
          }`,
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>
          {tokenHealth.status === "healthy"
            ? "✅"
            : tokenHealth.status === "needs_refresh" ||
                tokenHealth.status === "expiring_soon"
              ? "⚠️"
              : "❌"}{" "}
          Token Health:
        </h4>
        <p style={{ margin: "5px 0" }}>
          <strong>Status:</strong> {tokenHealth.status || "Unknown"}
        </p>
        <p style={{ margin: "5px 0" }}>
          <strong>Can Refresh:</strong>{" "}
          {tokenHealth.canRefresh ? "✅ Yes" : "❌ No"}
        </p>
        <p style={{ margin: "5px 0" }}>
          <strong>Needs Re-auth:</strong>{" "}
          {tokenHealth.needsReauth ? "❌ Yes" : "✅ No"}
        </p>
        {tokenHealth.recommendations &&
          tokenHealth.recommendations.length > 0 && (
            <div style={{ marginTop: "10px" }}>
              <strong>Recommendations:</strong>
              <ul style={{ margin: "5px 0 0 20px", padding: 0 }}>
                {tokenHealth.recommendations.map((rec, index) => (
                  <li key={index} style={{ fontSize: "14px" }}>
                    {rec}
                  </li>
                ))}
              </ul>
            </div>
          )}
      </div>

      {/* Action Buttons */}
      <div
        style={{
          marginBottom: "30px",
          display: "flex",
          gap: "10px",
          flexWrap: "wrap",
        }}
      >
        <button
          onClick={loadTokenInfo}
          style={{
            padding: "10px 20px",
            backgroundColor: "#28a745",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
          }}
        >
          🔄 Reload Info
        </button>

        <button
          onClick={handleManualRefresh}
          disabled={loading || !tokenManager.hasValidTokens()}
          style={{
            padding: "10px 20px",
            backgroundColor: loading ? "#6c757d" : "#17a2b8",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor:
              loading || !tokenManager.hasValidTokens()
                ? "not-allowed"
                : "pointer",
          }}
        >
          {loading ? "🔄 Refreshing..." : "🔄 Manual Refresh"}
        </button>

        <button
          onClick={handleForceCheck}
          style={{
            padding: "10px 20px",
            backgroundColor: "#ffc107",
            color: "#212529",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
          }}
        >
          🔍 Force Check
        </button>

        <button
          onClick={handleClearTokens}
          style={{
            padding: "10px 20px",
            backgroundColor: "#dc3545",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
          }}
        >
          🗑️ Clear Tokens
        </button>
      </div>

      {/* Error Display */}
      {error && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#f8d7da",
            borderRadius: "6px",
            color: "#721c24",
            border: "1px solid #f5c6cb",
          }}
        >
          <h4 style={{ margin: "0 0 10px 0" }}>❌ Error:</h4>
          <p style={{ margin: 0 }}>{error}</p>
        </div>
      )}

      {/* Event Log */}
      <div>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <h3>📋 Token Event Log ({refreshLog.length})</h3>
          <button
            onClick={handleClearLog}
            disabled={refreshLog.length === 0}
            style={{
              padding: "5px 15px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: refreshLog.length === 0 ? "not-allowed" : "pointer",
              fontSize: "14px",
            }}
          >
            Clear Log
          </button>
        </div>

        {refreshLog.length === 0 ? (
          <div
            style={{
              padding: "40px",
              textAlign: "center",
              backgroundColor: "#f8f9fa",
              borderRadius: "6px",
              border: "2px dashed #dee2e6",
            }}
          >
            <p style={{ fontSize: "18px", color: "#6c757d" }}>
              No token events logged yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Token events will appear here when tokens are updated, refreshed,
              or cleared.
            </p>
          </div>
        ) : (
          <div
            style={{
              maxHeight: "300px",
              overflow: "auto",
              border: "1px solid #dee2e6",
              borderRadius: "6px",
              backgroundColor: "#f8f9fa",
            }}
          >
            {refreshLog
              .slice()
              .reverse()
              .map((event, index) => (
                <div
                  key={`${event.timestamp}-${index}`}
                  style={{
                    padding: "10px",
                    borderBottom:
                      index < refreshLog.length - 1
                        ? "1px solid #dee2e6"
                        : "none",
                    fontFamily: "monospace",
                    fontSize: "12px",
                  }}
                >
                  <div style={{ marginBottom: "5px" }}>
                    <strong style={{ color: "#007bff" }}>
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </strong>
                    {" - "}
                    <strong style={{ color: "#28a745" }}>
                      {event.eventType}
                    </strong>
                  </div>
                  <div style={{ color: "#666", marginLeft: "20px" }}>
                    {JSON.stringify(event.eventData, null, 2)}
                  </div>
                </div>
              ))}
          </div>
        )}
      </div>

      {/* Debug Info */}
      <details style={{ marginTop: "20px" }}>
        <summary style={{ cursor: "pointer", fontWeight: "bold" }}>
          🔧 Debug Info
        </summary>
        <pre
          style={{
            backgroundColor: "#f8f9fa",
            padding: "10px",
            borderRadius: "4px",
            fontSize: "12px",
            overflow: "auto",
            marginTop: "10px",
          }}
        >
          {JSON.stringify(tokenManager.getDebugInfo(), null, 2)}
        </pre>
      </details>
    </div>
  );
};

export default TokenManagerExample;
