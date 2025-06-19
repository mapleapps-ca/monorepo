// src/components/TokenMonitor.jsx
import React from "react";

const TokenMonitor = ({
  isMonitoring,
  workerStatus,
  onStartMonitoring,
  onStopMonitoring,
  onForceRefresh,
  onClearTokens,
}) => {
  return (
    <div className="monitor-card">
      <h3>🔍 Token Monitor</h3>

      <div className="monitor-status">
        <div className="status-item">
          <span className="label">Monitoring Status:</span>
          <span className={`status ${isMonitoring ? "active" : "inactive"}`}>
            {isMonitoring ? "🟢 Active" : "🔴 Inactive"}
          </span>
        </div>

        <div className="status-item">
          <span className="label">Worker Status:</span>
          <span
            className={`status ${workerStatus.isInitialized ? "active" : "inactive"}`}
          >
            {workerStatus.isInitialized ? "🟢 Ready" : "🔴 Not Ready"}
          </span>
        </div>

        {workerStatus.lastCheck && (
          <div className="status-item">
            <span className="label">Last Check:</span>
            <span className="timestamp">
              {new Date(workerStatus.lastCheck).toLocaleTimeString()}
            </span>
          </div>
        )}

        {workerStatus.isRefreshing && (
          <div className="status-item">
            <span className="label">Current Status:</span>
            <span className="status refreshing">🔄 Refreshing tokens...</span>
          </div>
        )}
      </div>

      <div className="monitor-controls">
        {!isMonitoring ? (
          <button
            className="btn btn-primary"
            onClick={onStartMonitoring}
            disabled={!workerStatus.isInitialized}
          >
            ▶️ Start Monitoring
          </button>
        ) : (
          <button className="btn btn-secondary" onClick={onStopMonitoring}>
            ⏸️ Stop Monitoring
          </button>
        )}

        <button
          className="btn btn-action"
          onClick={onForceRefresh}
          disabled={!workerStatus.isInitialized || !isMonitoring}
        >
          🔄 Force Refresh
        </button>

        <button className="btn btn-danger" onClick={onClearTokens}>
          🗑️ Clear Tokens
        </button>
      </div>

      <div className="monitor-info">
        <h4>ℹ️ How it works:</h4>
        <ul>
          <li>Worker checks tokens every 30 seconds</li>
          <li>Auto-refreshes tokens 5 minutes before expiry</li>
          <li>Handles token refresh failures gracefully</li>
          <li>Synchronizes across multiple browser tabs</li>
          <li>Uses secure encrypted token storage</li>
        </ul>
      </div>
    </div>
  );
};

export default TokenMonitor;
