// src/components/TokenDisplay.jsx
import React from "react";

const TokenDisplay = ({ tokenInfo }) => {
  const formatDate = (dateString) => {
    if (!dateString) return "Not set";
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  const getTimeRemaining = (expiryDate) => {
    if (!expiryDate) return "Unknown";

    const now = new Date();
    const expiry = new Date(expiryDate);
    const diff = expiry.getTime() - now.getTime();

    if (diff <= 0) return "Expired";

    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  };

  const getStatusColor = (expired, expiringSoon) => {
    if (expired) return "error";
    if (expiringSoon) return "warning";
    return "success";
  };

  return (
    <div className="token-card">
      <h3>🔑 Token Information</h3>

      {tokenInfo.userEmail && (
        <div className="user-info">
          <strong>👤 User:</strong> {tokenInfo.userEmail}
        </div>
      )}

      <div className="token-grid">
        <div className="token-item">
          <h4>Access Token</h4>
          <div className="token-status">
            <span className="label">Status:</span>
            <span
              className={`status ${getStatusColor(tokenInfo.accessTokenExpired, tokenInfo.accessTokenExpiringSoon)}`}
            >
              {tokenInfo.accessTokenExpired
                ? "🔴 Expired"
                : tokenInfo.accessTokenExpiringSoon
                  ? "🟡 Expires Soon"
                  : "🟢 Valid"}
            </span>
          </div>
          <div className="token-detail">
            <span className="label">Expires:</span>
            <span className="value">
              {formatDate(tokenInfo.accessTokenExpiry)}
            </span>
          </div>
          <div className="token-detail">
            <span className="label">Time Remaining:</span>
            <span className="value">
              {getTimeRemaining(tokenInfo.accessTokenExpiry)}
            </span>
          </div>
        </div>

        <div className="token-item">
          <h4>Refresh Token</h4>
          <div className="token-status">
            <span className="label">Status:</span>
            <span
              className={`status ${tokenInfo.refreshTokenExpired ? "error" : "success"}`}
            >
              {tokenInfo.refreshTokenExpired ? "🔴 Expired" : "🟢 Valid"}
            </span>
          </div>
          <div className="token-detail">
            <span className="label">Expires:</span>
            <span className="value">
              {formatDate(tokenInfo.refreshTokenExpiry)}
            </span>
          </div>
          <div className="token-detail">
            <span className="label">Time Remaining:</span>
            <span className="value">
              {getTimeRemaining(tokenInfo.refreshTokenExpiry)}
            </span>
          </div>
        </div>
      </div>

      <div className="token-summary">
        <div className="summary-item">
          <span className="label">Access Token Present:</span>
          <span
            className={`value ${tokenInfo.hasAccessToken ? "success" : "error"}`}
          >
            {tokenInfo.hasAccessToken ? "✅ Yes" : "❌ No"}
          </span>
        </div>
        <div className="summary-item">
          <span className="label">Refresh Token Present:</span>
          <span
            className={`value ${tokenInfo.hasRefreshToken ? "success" : "error"}`}
          >
            {tokenInfo.hasRefreshToken ? "✅ Yes" : "❌ No"}
          </span>
        </div>
      </div>

      {(tokenInfo.accessTokenExpired || tokenInfo.refreshTokenExpired) && (
        <div className="warning-box">
          <h4>⚠️ Token Issues</h4>
          {tokenInfo.refreshTokenExpired && (
            <p className="error">
              Refresh token has expired. Please log in again.
            </p>
          )}
          {tokenInfo.accessTokenExpired && !tokenInfo.refreshTokenExpired && (
            <p className="warning">
              Access token has expired but will be automatically refreshed.
            </p>
          )}
        </div>
      )}
    </div>
  );
};

export default TokenDisplay;
