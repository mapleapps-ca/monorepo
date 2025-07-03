// Authentication Worker - Updated for Encrypted Token System
// This worker runs independently and communicates with all tabs

const STORAGE_KEYS = {
  ACCESS_TOKEN: "mapleapps_access_token",
  REFRESH_TOKEN: "mapleapps_refresh_token",
  ACCESS_TOKEN_EXPIRY: "mapleapps_access_token_expiry",
  REFRESH_TOKEN_EXPIRY: "mapleapps_refresh_token_expiry",
  USER_EMAIL: "mapleapps_user_email",
};

const API_BASE_URL = "/iam/api/v1";
const CHECK_INTERVAL = 30000; // Check every 30 seconds
const REFRESH_THRESHOLD = 5 * 60 * 1000; // Refresh 5 minutes before expiry

let checkInterval = null;
let isRefreshing = false;
let isMonitoring = false;

// Broadcast channel for cross-tab communication
let broadcastChannel = null;
try {
  broadcastChannel = new BroadcastChannel("auth_worker");
  console.log("[AuthWorker] BroadcastChannel initialized successfully");
} catch (error) {
  console.warn(
    "[AuthWorker] BroadcastChannel not supported, falling back to postMessage:",
    error,
  );
}

// Worker state
let workerState = {
  isAuthenticated: false,
  isRefreshing: false,
  lastCheck: null,
  tokenInfo: {},
};

// Broadcast message to all tabs
function broadcastMessage(type, data) {
  const message = {
    type,
    data,
    timestamp: Date.now(),
  };

  console.log(`[AuthWorker] Broadcasting message: ${type}`, data);

  // Always try postMessage first (most reliable)
  try {
    self.postMessage(message);
  } catch (error) {
    console.error(
      `[AuthWorker] Failed to send message via postMessage: ${type}`,
      error,
    );
  }

  // Use BroadcastChannel if available (for cross-tab communication)
  if (broadcastChannel) {
    try {
      broadcastChannel.postMessage(message);
    } catch (error) {
      console.error(
        `[AuthWorker] Failed to send message via BroadcastChannel: ${type}`,
        error,
      );
    }
  }
}

// Check if tokens are expired
function isTokenExpired(expiryTime) {
  if (!expiryTime) return true;
  return new Date() >= new Date(expiryTime);
}

// Check if token expires soon
function isTokenExpiringSoon(expiryTime, thresholdMs = REFRESH_THRESHOLD) {
  if (!expiryTime) return true;
  const expiry = new Date(expiryTime);
  const now = new Date();
  return expiry.getTime() - now.getTime() <= thresholdMs;
}

// Get current token information
function getTokenInfo(storageData) {
  const accessToken = storageData[STORAGE_KEYS.ACCESS_TOKEN];
  const refreshToken = storageData[STORAGE_KEYS.REFRESH_TOKEN];
  const accessTokenExpiry = storageData[STORAGE_KEYS.ACCESS_TOKEN_EXPIRY];
  const refreshTokenExpiry = storageData[STORAGE_KEYS.REFRESH_TOKEN_EXPIRY];

  const accessTokenExpired = isTokenExpired(accessTokenExpiry);
  const refreshTokenExpired = isTokenExpired(refreshTokenExpiry);
  const accessTokenExpiringSoon = isTokenExpiringSoon(accessTokenExpiry);

  return {
    hasTokens: !!(accessToken && refreshToken),
    hasRefreshToken: !!refreshToken,
    accessTokenExpired,
    refreshTokenExpired,
    accessTokenExpiringSoon,
    accessTokenExpiry,
    refreshTokenExpiry,
    isAuthenticated: !!(accessToken && refreshToken) && !refreshTokenExpired,
  };
}

// Make API request for token refresh
async function refreshTokens(refreshTokenValue, storageData, requestId = null) {
  const url = `${API_BASE_URL}/token/refresh`;

  try {
    console.log("[AuthWorker] Attempting token refresh...");

    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        value: refreshTokenValue,
      }),
    });

    if (response.status === 201) {
      const result = await response.json();
      console.log("[AuthWorker] Token refresh successful");

      // Backend returns ENCRYPTED tokens - we need to request decryption
      if (
        result.encrypted_access_token &&
        result.encrypted_refresh_token &&
        result.token_nonce
      ) {
        console.log("[AuthWorker] Received encrypted tokens from refresh");

        // Request main thread to decrypt tokens
        // We'll send the encrypted data and wait for decrypted response
        broadcastMessage("decrypt_tokens_request", {
          encryptedAccessToken: result.encrypted_access_token,
          encryptedRefreshToken: result.encrypted_refresh_token,
          tokenNonce: result.token_nonce,
          accessTokenExpiry: result.access_token_expiry_date,
          refreshTokenExpiry: result.refresh_token_expiry_date,
          username: result.username,
          requestId: requestId,
        });

        // Don't mark as success yet - wait for decryption to complete
        return "pending_decryption";
      } else {
        console.error(
          "[AuthWorker] Unexpected response format - no encrypted tokens",
        );
        throw new Error("Invalid token refresh response format");
      }
    } else {
      const errorData = await response.json();
      throw new Error(errorData.message || "Token refresh failed");
    }
  } catch (error) {
    console.error("[AuthWorker] Token refresh failed:", error);

    // Clear all tokens on refresh failure
    Object.values(STORAGE_KEYS).forEach((key) => {
      broadcastMessage("storage_remove", { key });
    });

    // Broadcast failure
    broadcastMessage("token_refresh_failed", {
      error: error.message,
      shouldRedirect: true,
      requestId: requestId,
    });

    return false;
  }
}

// Handle decrypted tokens response
function handleDecryptedTokens(data) {
  console.log("[AuthWorker] Handling decrypted tokens");

  const {
    accessToken,
    refreshToken,
    accessTokenExpiry,
    refreshTokenExpiry,
    username,
    requestId,
  } = data;

  // Store decrypted tokens
  broadcastMessage("storage_update", {
    key: STORAGE_KEYS.ACCESS_TOKEN,
    value: accessToken,
  });
  broadcastMessage("storage_update", {
    key: STORAGE_KEYS.REFRESH_TOKEN,
    value: refreshToken,
  });

  // Update expiry times
  if (accessTokenExpiry) {
    broadcastMessage("storage_update", {
      key: STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
      value: accessTokenExpiry,
    });
  }
  if (refreshTokenExpiry) {
    broadcastMessage("storage_update", {
      key: STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
      value: refreshTokenExpiry,
    });
  }

  // Update user email if provided
  if (username) {
    broadcastMessage("storage_update", {
      key: STORAGE_KEYS.USER_EMAIL,
      value: username,
    });
  }

  // Broadcast success
  broadcastMessage("token_refresh_success", {
    accessTokenExpiry: accessTokenExpiry,
    refreshTokenExpiry: refreshTokenExpiry,
    username: username,
    requestId: requestId,
  });
}

// Main token checking logic
async function checkTokens(storageData) {
  const tokenInfo = getTokenInfo(storageData);

  // Update worker state
  workerState.tokenInfo = tokenInfo;
  workerState.lastCheck = new Date().toISOString();
  workerState.isAuthenticated = tokenInfo.isAuthenticated;

  // Broadcast current status
  broadcastMessage("token_status_update", {
    tokenInfo,
    lastCheck: workerState.lastCheck,
    isAuthenticated: workerState.isAuthenticated,
  });

  // If no tokens, nothing to do
  if (!tokenInfo.hasTokens) {
    console.log("[AuthWorker] No tokens available");
    return;
  }

  // If refresh token is expired, logout user
  if (tokenInfo.refreshTokenExpired) {
    console.log("[AuthWorker] Refresh token expired, logging out user");

    // Clear all tokens
    Object.values(STORAGE_KEYS).forEach((key) => {
      broadcastMessage("storage_remove", { key });
    });

    // Broadcast logout
    broadcastMessage("force_logout", {
      reason: "refresh_token_expired",
      shouldRedirect: true,
    });

    return;
  }

  // If access token is expired or expiring soon, refresh
  if (
    (tokenInfo.accessTokenExpired || tokenInfo.accessTokenExpiringSoon) &&
    !isRefreshing
  ) {
    isRefreshing = true;
    workerState.isRefreshing = true;

    console.log("[AuthWorker] Access token needs refresh");

    const refreshToken = storageData[STORAGE_KEYS.REFRESH_TOKEN];
    const result = await refreshTokens(refreshToken, storageData);

    // If result is "pending_decryption", we'll wait for the main thread to decrypt
    if (result !== "pending_decryption") {
      isRefreshing = false;
      workerState.isRefreshing = false;

      if (!result) {
        workerState.isAuthenticated = false;
      }
    }
    // If pending, we'll clear the flag when we receive the decrypted tokens
  }
}

// Start monitoring tokens
function startTokenMonitoring() {
  console.log("[AuthWorker] Starting token monitoring...");
  isMonitoring = true;

  if (checkInterval) {
    clearInterval(checkInterval);
  }

  // Initial check
  broadcastMessage("request_storage_data", {});

  // Set up interval
  checkInterval = setInterval(async () => {
    if (isRefreshing || !isMonitoring) {
      console.log(
        "[AuthWorker] Skipping check - refresh in progress or monitoring stopped",
      );
      return;
    }

    try {
      broadcastMessage("request_storage_data", {});
    } catch (error) {
      console.error("[AuthWorker] Error during token check:", error);
    }
  }, CHECK_INTERVAL);
}

// Stop monitoring tokens
function stopTokenMonitoring() {
  console.log("[AuthWorker] Stopping token monitoring...");
  isMonitoring = false;

  if (checkInterval) {
    clearInterval(checkInterval);
    checkInterval = null;
  }

  isRefreshing = false;
  workerState.isRefreshing = false;
  workerState.isAuthenticated = false;
}

// Handle messages from main thread
self.addEventListener("message", async (event) => {
  const { type, data } = event.data;

  switch (type) {
    case "start_monitoring":
      console.log("[AuthWorker] Received start_monitoring command");
      startTokenMonitoring();
      break;

    case "stop_monitoring":
      console.log("[AuthWorker] Received stop_monitoring command");
      stopTokenMonitoring();
      break;

    case "storage_data_response":
      // Received storage data from main thread
      if (data && !isRefreshing && isMonitoring) {
        await checkTokens(data);
      }
      break;

    case "force_token_check":
      console.log("[AuthWorker] Received force_token_check command");
      if (data && !isRefreshing) {
        await checkTokens(data);
      }
      break;

    case "manual_refresh":
      console.log("[AuthWorker] Received manual_refresh command");
      if (data && data.refreshToken && !isRefreshing) {
        isRefreshing = true;
        workerState.isRefreshing = true;
        const result = await refreshTokens(
          data.refreshToken,
          data.storageData || {},
          data.requestId,
        );

        // Only clear refreshing flag if not pending decryption
        if (result !== "pending_decryption") {
          isRefreshing = false;
          workerState.isRefreshing = false;

          if (!result && data.requestId) {
            // Make sure failure is reported for manual refresh
            broadcastMessage("token_refresh_failed", {
              error: "Manual refresh failed",
              requestId: data.requestId,
            });
          }
        }
      }
      break;

    case "decrypted_tokens_response":
      console.log("[AuthWorker] Received decrypted tokens");
      handleDecryptedTokens(data);
      // Clear refreshing flag after successful decryption
      isRefreshing = false;
      workerState.isRefreshing = false;
      break;

    case "decrypt_tokens_failed":
      console.log("[AuthWorker] Token decryption failed");
      isRefreshing = false;
      workerState.isRefreshing = false;

      // Clear all tokens on decryption failure
      Object.values(STORAGE_KEYS).forEach((key) => {
        broadcastMessage("storage_remove", { key });
      });

      // Broadcast failure
      broadcastMessage("token_refresh_failed", {
        error: data.error || "Token decryption failed",
        shouldRedirect: true,
        requestId: data.requestId,
      });
      break;

    case "get_worker_status":
      // Send current worker status
      broadcastMessage("worker_status_response", {
        ...workerState,
        isRefreshing,
        isInitialized: true,
        isMonitoring,
        tokenSystem: "encrypted",
        checkInterval: CHECK_INTERVAL,
        refreshThreshold: REFRESH_THRESHOLD,
      });
      break;

    default:
      console.log("[AuthWorker] Unknown message type:", type);
  }
});

// Handle errors
self.addEventListener("error", (error) => {
  console.error("[AuthWorker] Worker error:", error);
  broadcastMessage("worker_error", {
    error: error.message,
    filename: error.filename,
    lineno: error.lineno,
  });
});

// Initialize worker
console.log("[AuthWorker] Authentication worker initializing...");

// Send ready signal
try {
  console.log("[AuthWorker] Sending worker ready signal...");
  broadcastMessage("worker_ready", {
    timestamp: Date.now(),
    checkInterval: CHECK_INTERVAL,
    refreshThreshold: REFRESH_THRESHOLD,
    tokenSystem: "encrypted",
  });
  console.log("[AuthWorker] Worker ready signal sent successfully");
} catch (error) {
  console.error("[AuthWorker] Failed to send ready signal:", error);
}

console.log(
  "[AuthWorker] Authentication worker initialized with encrypted token support",
);
