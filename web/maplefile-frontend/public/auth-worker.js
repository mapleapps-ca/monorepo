// Authentication Worker - Updated for Unencrypted Token System
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

// Utility functions for localStorage access in worker
function getStorageItem(key) {
  try {
    // In worker context, we need to simulate localStorage access
    // We'll get this data from the main thread
    return self.localStorage?.[key] || null;
  } catch (error) {
    return null;
  }
}

function setStorageItem(key, value) {
  try {
    if (self.localStorage) {
      self.localStorage[key] = value;
    }
    // Also broadcast to main thread
    broadcastMessage("storage_update", { key, value });
  } catch (error) {
    // Fallback - send to main thread
    broadcastMessage("storage_update", { key, value });
  }
}

function removeStorageItem(key) {
  try {
    if (self.localStorage) {
      delete self.localStorage[key];
    }
    broadcastMessage("storage_remove", { key });
  } catch (error) {
    broadcastMessage("storage_remove", { key });
  }
}

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
    console.log(`[AuthWorker] Message sent via postMessage: ${type}`);
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
      console.log(`[AuthWorker] Message sent via BroadcastChannel: ${type}`);
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

// Get current token information for unencrypted tokens
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

// Make API request for token refresh using unencrypted tokens
async function refreshTokens(refreshTokenValue, storageData) {
  const url = `${API_BASE_URL}/token/refresh`;

  try {
    console.log(
      "[AuthWorker] Attempting token refresh with unencrypted token...",
    );

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

      // Handle unencrypted tokens (expected after initial login)
      if (result.access_token && result.refresh_token) {
        console.log(
          "[AuthWorker] Received unencrypted access and refresh tokens",
        );

        // Store new unencrypted tokens
        setStorageItem(STORAGE_KEYS.ACCESS_TOKEN, result.access_token);
        setStorageItem(STORAGE_KEYS.REFRESH_TOKEN, result.refresh_token);

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

        // Broadcast success
        broadcastMessage("token_refresh_success", {
          accessTokenExpiry: result.access_token_expiry_date,
          refreshTokenExpiry: result.refresh_token_expiry_date,
          username: result.username,
        });

        return true;
      } else if (result.encrypted_tokens && result.token_nonce) {
        // This shouldn't happen after initial login
        console.error(
          "[AuthWorker] Received encrypted tokens in refresh - this is unexpected",
        );
        throw new Error("Unexpected encrypted tokens in refresh response");
      } else {
        console.error("[AuthWorker] No valid tokens in refresh response");
        console.error(
          "[AuthWorker] Available response fields:",
          Object.keys(result),
        );
        throw new Error("Token refresh failed: No valid tokens received");
      }
    } else {
      const errorData = await response.json();
      throw new Error(errorData.message || "Token refresh failed");
    }
  } catch (error) {
    console.error("[AuthWorker] Token refresh failed:", error);

    // Clear all tokens on refresh failure
    Object.values(STORAGE_KEYS).forEach((key) => {
      removeStorageItem(key);
    });

    // Broadcast failure
    broadcastMessage("token_refresh_failed", {
      error: error.message,
      shouldRedirect: true,
    });

    return false;
  }
}

// Main token checking logic for unencrypted tokens
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
      removeStorageItem(key);
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
    const success = await refreshTokens(refreshToken, storageData);

    isRefreshing = false;
    workerState.isRefreshing = false;

    if (!success) {
      // Refresh failed, user will be redirected by the failed handler
      workerState.isAuthenticated = false;
    }
  }
}

// Start monitoring tokens
function startTokenMonitoring() {
  console.log("[AuthWorker] Starting unencrypted token monitoring...");

  if (checkInterval) {
    clearInterval(checkInterval);
  }

  checkInterval = setInterval(async () => {
    if (isRefreshing) {
      console.log("[AuthWorker] Refresh in progress, skipping check");
      return;
    }

    try {
      // Request current storage data from main thread
      broadcastMessage("request_storage_data", {});
    } catch (error) {
      console.error("[AuthWorker] Error during token check:", error);
    }
  }, CHECK_INTERVAL);

  // Also check immediately
  broadcastMessage("request_storage_data", {});
}

// Stop monitoring tokens
function stopTokenMonitoring() {
  console.log("[AuthWorker] Stopping token monitoring...");

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
      if (data && !isRefreshing) {
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
        isInitialized: true,
        tokenSystem: "unencrypted",
      });
      break;

    case "password_response":
      // This is handled by the requestPasswordFromMainThread promise
      // No additional action needed here - just don't throw unknown message error
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

// Send ready signal immediately
try {
  console.log("[AuthWorker] Sending worker ready signal...");
  broadcastMessage("worker_ready", {
    timestamp: Date.now(),
    checkInterval: CHECK_INTERVAL,
    refreshThreshold: REFRESH_THRESHOLD,
    tokenSystem: "unencrypted",
  });
  console.log("[AuthWorker] Worker ready signal sent successfully");
} catch (error) {
  console.error("[AuthWorker] Failed to send ready signal:", error);
  // Try to send via postMessage as fallback
  try {
    self.postMessage({
      type: "worker_ready",
      data: {
        timestamp: Date.now(),
        checkInterval: CHECK_INTERVAL,
        refreshThreshold: REFRESH_THRESHOLD,
        tokenSystem: "unencrypted",
      },
      timestamp: Date.now(),
    });
    console.log(
      "[AuthWorker] Worker ready signal sent via postMessage fallback",
    );
  } catch (fallbackError) {
    console.error(
      "[AuthWorker] Failed to send ready signal via fallback:",
      fallbackError,
    );
  }
}

console.log(
  "[AuthWorker] Authentication worker initialized with unencrypted token support",
);

async function refreshTokensWithPassword(refreshTokenValue, storageData) {
  const url = `${API_BASE_URL}/token/refresh`;

  try {
    console.log("[AuthWorker] Attempting token refresh...");

    // Check if we need password for this refresh
    const needsPassword =
      storageData.encrypted_tokens || storageData.requires_decryption;

    let password = null;
    if (needsPassword) {
      // Request password from main thread
      password = await requestPasswordFromMainThread();

      if (!password) {
        throw new Error(
          "Password required for token refresh but not available",
        );
      }
    }

    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        value: refreshTokenValue,
        // Include password if needed for E2EE token decryption
        ...(password && { password: password }),
      }),
    });

    if (response.status === 201) {
      const result = await response.json();
      console.log("[AuthWorker] Token refresh successful");

      // Handle the response based on your existing logic
      if (result.access_token && result.refresh_token) {
        // Store new tokens
        setStorageItem(STORAGE_KEYS.ACCESS_TOKEN, result.access_token);
        setStorageItem(STORAGE_KEYS.REFRESH_TOKEN, result.refresh_token);

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

        // Broadcast success
        broadcastMessage("token_refresh_success", {
          accessTokenExpiry: result.access_token_expiry_date,
          refreshTokenExpiry: result.refresh_token_expiry_date,
        });

        return true;
      } else {
        throw new Error("No valid tokens in refresh response");
      }
    } else {
      const errorData = await response.json();
      throw new Error(errorData.message || "Token refresh failed");
    }
  } catch (error) {
    console.error("[AuthWorker] Token refresh failed:", error);

    // Broadcast failure
    broadcastMessage("token_refresh_failed", {
      error: error.message,
      shouldRedirect: error.message.includes("Password required"),
    });

    return false;
  }
}

// Function to request password from main thread
function requestPasswordFromMainThread() {
  return new Promise((resolve) => {
    const requestId = Date.now() + Math.random();

    const handlePasswordResponse = (event) => {
      if (
        event.data.type === "password_response" &&
        event.data.requestId === requestId
      ) {
        self.removeEventListener("message", handlePasswordResponse);
        resolve(event.data.password);
      }
    };

    self.addEventListener("message", handlePasswordResponse);

    // Send password request
    broadcastMessage("password_request", { requestId });

    // Timeout after 10 seconds
    setTimeout(() => {
      self.removeEventListener("message", handlePasswordResponse);
      resolve(null);
    }, 10000);
  });
}
