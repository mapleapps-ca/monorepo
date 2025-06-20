// Token Refresh Worker - Simplified for maplefile-refreshtoken prototype
// This worker runs independently and handles automatic token refresh

const STORAGE_KEYS = {
  ENCRYPTED_ACCESS_TOKEN: "mapleapps_encrypted_access_token",
  ENCRYPTED_REFRESH_TOKEN: "mapleapps_encrypted_refresh_token",
  TOKEN_NONCE: "mapleapps_token_nonce",
  ACCESS_TOKEN_EXPIRY: "mapleapps_access_token_expiry",
  REFRESH_TOKEN_EXPIRY: "mapleapps_refresh_token_expiry",
  USER_EMAIL: "mapleapps_user_email",
  // Legacy single encrypted tokens field
  ENCRYPTED_TOKENS: "mapleapps_encrypted_tokens",
};

const API_BASE_URL = "/iam/api/v1";
const CHECK_INTERVAL = 30000; // Check every 30 seconds
const REFRESH_THRESHOLD = 5 * 60 * 1000; // Refresh 5 minutes before expiry

let checkInterval = null;
let isRefreshing = false;

// Broadcast channel for cross-tab communication
let broadcastChannel = null;
try {
  broadcastChannel = new BroadcastChannel("token_refresh_worker");
  console.log("[TokenRefreshWorker] BroadcastChannel initialized successfully");
} catch (error) {
  console.warn(
    "[TokenRefreshWorker] BroadcastChannel not supported, falling back to postMessage:",
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

  console.log(`[TokenRefreshWorker] Broadcasting message: ${type}`, data);

  // Always try postMessage first (most reliable)
  try {
    self.postMessage(message);
    console.log(`[TokenRefreshWorker] Message sent via postMessage: ${type}`);
  } catch (error) {
    console.error(
      `[TokenRefreshWorker] Failed to send message via postMessage: ${type}`,
      error,
    );
  }

  // Use BroadcastChannel if available (for cross-tab communication)
  if (broadcastChannel) {
    try {
      broadcastChannel.postMessage(message);
      console.log(
        `[TokenRefreshWorker] Message sent via BroadcastChannel: ${type}`,
      );
    } catch (error) {
      console.error(
        `[TokenRefreshWorker] Failed to send message via BroadcastChannel: ${type}`,
        error,
      );
    }
  }
}

// Storage update functions
function setStorageItem(key, value) {
  broadcastMessage("storage_update", { key, value });
}

function removeStorageItem(key) {
  broadcastMessage("storage_remove", { key });
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
  // Check for separate encrypted tokens first
  const encryptedAccessToken = storageData[STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN];
  const encryptedRefreshToken =
    storageData[STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN];
  const tokenNonce = storageData[STORAGE_KEYS.TOKEN_NONCE];

  // Fallback to legacy single encrypted_tokens field
  const legacyEncryptedTokens = storageData[STORAGE_KEYS.ENCRYPTED_TOKENS];

  const accessTokenExpiry = storageData[STORAGE_KEYS.ACCESS_TOKEN_EXPIRY];
  const refreshTokenExpiry = storageData[STORAGE_KEYS.REFRESH_TOKEN_EXPIRY];

  const accessTokenExpired = isTokenExpired(accessTokenExpiry);
  const refreshTokenExpired = isTokenExpired(refreshTokenExpiry);
  const accessTokenExpiringSoon = isTokenExpiringSoon(accessTokenExpiry);

  // Determine if we have valid encrypted tokens
  const hasEncryptedTokens = !!(
    (encryptedAccessToken && encryptedRefreshToken && tokenNonce) ||
    (legacyEncryptedTokens && tokenNonce)
  );

  return {
    hasEncryptedTokens,
    hasRefreshToken: !!(encryptedRefreshToken || legacyEncryptedTokens),
    accessTokenExpired,
    refreshTokenExpired,
    accessTokenExpiringSoon,
    accessTokenExpiry,
    refreshTokenExpiry,
    isAuthenticated: hasEncryptedTokens && !refreshTokenExpired,
    tokenFormat:
      encryptedAccessToken && encryptedRefreshToken ? "separate" : "legacy",
  };
}

// Make API request for token refresh
async function refreshTokens(refreshTokenValue, storageData) {
  const url = `${API_BASE_URL}/token/refresh`;

  try {
    console.log("[TokenRefreshWorker] Attempting token refresh...");

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
      console.log("[TokenRefreshWorker] Token refresh successful");

      // Handle separate encrypted tokens (new format)
      if (
        result.encrypted_access_token &&
        result.encrypted_refresh_token &&
        result.token_nonce
      ) {
        console.log(
          "[TokenRefreshWorker] Received separate encrypted access and refresh tokens",
        );

        setStorageItem(
          STORAGE_KEYS.ENCRYPTED_ACCESS_TOKEN,
          result.encrypted_access_token,
        );
        setStorageItem(
          STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN,
          result.encrypted_refresh_token,
        );
        setStorageItem(STORAGE_KEYS.TOKEN_NONCE, result.token_nonce);

        // Update expiry times
        if (result.access_token_expiry_date) {
          setStorageItem(
            STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
            result.access_token_expiry_date,
          );
        }
        if (result.refresh_token_expiry_date) {
          setStorageItem(
            STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
            result.refresh_token_expiry_date,
          );
        }

        // Update user email if provided
        if (result.username) {
          setStorageItem(STORAGE_KEYS.USER_EMAIL, result.username);
        }

        broadcastMessage("token_refresh_success", {
          accessTokenExpiry: result.access_token_expiry_date,
          refreshTokenExpiry: result.refresh_token_expiry_date,
          username: result.username,
          tokenFormat: "separate",
        });

        return true;
      } else if (result.encrypted_tokens && result.token_nonce) {
        // Handle legacy single encrypted_tokens field
        console.log(
          "[TokenRefreshWorker] Received legacy single encrypted tokens",
        );

        setStorageItem(STORAGE_KEYS.ENCRYPTED_TOKENS, result.encrypted_tokens);
        setStorageItem(STORAGE_KEYS.TOKEN_NONCE, result.token_nonce);

        // Update expiry times
        if (result.access_token_expiry_date) {
          setStorageItem(
            STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
            result.access_token_expiry_date,
          );
        }
        if (result.refresh_token_expiry_date) {
          setStorageItem(
            STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
            result.refresh_token_expiry_date,
          );
        }

        // Update user email if provided
        if (result.username) {
          setStorageItem(STORAGE_KEYS.USER_EMAIL, result.username);
        }

        broadcastMessage("token_refresh_success", {
          accessTokenExpiry: result.access_token_expiry_date,
          refreshTokenExpiry: result.refresh_token_expiry_date,
          username: result.username,
          tokenFormat: "legacy",
        });

        return true;
      } else {
        console.error(
          "[TokenRefreshWorker] No encrypted tokens in refresh response",
        );
        throw new Error("Token refresh failed: No encrypted tokens received");
      }
    } else {
      const errorData = await response.json();
      throw new Error(errorData.message || "Token refresh failed");
    }
  } catch (error) {
    console.error("[TokenRefreshWorker] Token refresh failed:", error);

    // Clear all tokens on refresh failure
    Object.values(STORAGE_KEYS).forEach((key) => {
      removeStorageItem(key);
    });

    broadcastMessage("token_refresh_failed", {
      error: error.message,
      shouldRedirect: true,
    });

    return false;
  }
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

  // If no encrypted tokens, nothing to do
  if (!tokenInfo.hasEncryptedTokens) {
    console.log("[TokenRefreshWorker] No encrypted tokens available");
    return;
  }

  // If refresh token is expired, logout user
  if (tokenInfo.refreshTokenExpired) {
    console.log("[TokenRefreshWorker] Refresh token expired, logging out user");

    Object.values(STORAGE_KEYS).forEach((key) => {
      removeStorageItem(key);
    });

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

    console.log("[TokenRefreshWorker] Access token needs refresh");

    // Get the refresh token (could be separate or legacy format)
    const refreshToken =
      storageData[STORAGE_KEYS.ENCRYPTED_REFRESH_TOKEN] ||
      storageData[STORAGE_KEYS.ENCRYPTED_TOKENS];

    const success = await refreshTokens(refreshToken, storageData);

    isRefreshing = false;
    workerState.isRefreshing = success;

    if (!success) {
      workerState.isAuthenticated = false;
    }
  }
}

// Start monitoring tokens
function startTokenMonitoring() {
  console.log("[TokenRefreshWorker] Starting token monitoring...");

  if (checkInterval) {
    clearInterval(checkInterval);
  }

  checkInterval = setInterval(async () => {
    if (isRefreshing) {
      console.log("[TokenRefreshWorker] Refresh in progress, skipping check");
      return;
    }

    try {
      // Request current storage data from main thread
      broadcastMessage("request_storage_data", {});
    } catch (error) {
      console.error("[TokenRefreshWorker] Error during token check:", error);
    }
  }, CHECK_INTERVAL);

  // Also check immediately
  broadcastMessage("request_storage_data", {});
}

// Stop monitoring tokens
function stopTokenMonitoring() {
  console.log("[TokenRefreshWorker] Stopping token monitoring...");

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
      console.log("[TokenRefreshWorker] Received start_monitoring command");
      startTokenMonitoring();
      break;

    case "stop_monitoring":
      console.log("[TokenRefreshWorker] Received stop_monitoring command");
      stopTokenMonitoring();
      break;

    case "storage_data_response":
      // Received storage data from main thread
      if (data && !isRefreshing) {
        await checkTokens(data);
      }
      break;

    case "force_token_check":
      console.log("[TokenRefreshWorker] Received force_token_check command");
      if (data && !isRefreshing) {
        await checkTokens(data);
      }
      break;

    case "manual_refresh":
      console.log("[TokenRefreshWorker] Received manual_refresh command");
      if (data && data.refreshToken && !isRefreshing) {
        isRefreshing = true;
        workerState.isRefreshing = true;
        const success = await refreshTokens(
          data.refreshToken,
          data.storageData || {},
        );
        isRefreshing = false;
        workerState.isRefreshing = false;
      }
      break;

    case "get_worker_status":
      // Send current worker status
      broadcastMessage("worker_status_response", {
        ...workerState,
        isRefreshing,
        isInitialized: true, // Worker is initialized if it can respond
        tokenSystem: "encrypted",
        checkInterval: CHECK_INTERVAL,
        refreshThreshold: REFRESH_THRESHOLD,
      });
      break;

    default:
      console.log("[TokenRefreshWorker] Unknown message type:", type);
  }
});

// Handle errors
self.addEventListener("error", (error) => {
  console.error("[TokenRefreshWorker] Worker error:", error);
  broadcastMessage("worker_error", {
    error: error.message,
    filename: error.filename,
    lineno: error.lineno,
  });
});

// Initialize worker
console.log("[TokenRefreshWorker] Token refresh worker initializing...");

// Send ready signal
try {
  console.log("[TokenRefreshWorker] Sending worker ready signal...");
  broadcastMessage("worker_ready", {
    timestamp: Date.now(),
    checkInterval: CHECK_INTERVAL,
    refreshThreshold: REFRESH_THRESHOLD,
    tokenSystem: "encrypted",
  });
  console.log("[TokenRefreshWorker] Worker ready signal sent successfully");
} catch (error) {
  console.error("[TokenRefreshWorker] Failed to send ready signal:", error);
}

console.log(
  "[TokenRefreshWorker] Token refresh worker initialized successfully",
);
