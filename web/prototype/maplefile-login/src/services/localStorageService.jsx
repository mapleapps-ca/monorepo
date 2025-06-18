// Local Storage Service for managing authentication tokens
const LOCAL_STORAGE_KEYS = {
  ACCESS_TOKEN: "mapleapps_access_token",
  REFRESH_TOKEN: "mapleapps_refresh_token",
  ACCESS_TOKEN_EXPIRY: "mapleapps_access_token_expiry",
  REFRESH_TOKEN_EXPIRY: "mapleapps_refresh_token_expiry",
  USER_EMAIL: "mapleapps_user_email",
};

class LocalStorageService {
  // Set access token with expiry
  setAccessToken(token, expiryTime) {
    if (token) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN, token);
      if (expiryTime) {
        localStorage.setItem(
          LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
          expiryTime,
        );
      }
    }
  }

  // Get access token
  getAccessToken() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN);
  }

  // Set refresh token with expiry
  setRefreshToken(token, expiryTime) {
    if (token) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN, token);
      if (expiryTime) {
        localStorage.setItem(
          LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
          expiryTime,
        );
      }
    }
  }

  // Get refresh token
  getRefreshToken() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN);
  }

  // Set user email
  setUserEmail(email) {
    if (email) {
      localStorage.setItem(LOCAL_STORAGE_KEYS.USER_EMAIL, email);
    }
  }

  // Get user email
  getUserEmail() {
    return localStorage.getItem(LOCAL_STORAGE_KEYS.USER_EMAIL);
  }

  // Check if access token is expired
  isAccessTokenExpired() {
    const expiryTime = localStorage.getItem(
      LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
    );
    if (!expiryTime) return true;

    return new Date() >= new Date(expiryTime);
  }

  // Check if access token is expiring soon (within specified minutes)
  isAccessTokenExpiringSoon(minutesBeforeExpiry = 5) {
    const expiryTime = localStorage.getItem(
      LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY,
    );
    if (!expiryTime) return true;

    const expiry = new Date(expiryTime);
    const now = new Date();
    const timeUntilExpiry = expiry.getTime() - now.getTime();
    const warningThreshold = minutesBeforeExpiry * 60 * 1000; // Convert to milliseconds

    return timeUntilExpiry <= warningThreshold;
  }

  // Check if refresh token is expired
  isRefreshTokenExpired() {
    const expiryTime = localStorage.getItem(
      LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY,
    );
    if (!expiryTime) return true;

    return new Date() >= new Date(expiryTime);
  }

  // Check if user is authenticated (has valid tokens)
  isAuthenticated() {
    const accessToken = this.getAccessToken();
    const refreshToken = this.getRefreshToken();

    if (!accessToken || !refreshToken) return false;

    // If access token is valid, user is authenticated
    if (!this.isAccessTokenExpired()) return true;

    // If access token is expired but refresh token is valid, we can refresh
    if (!this.isRefreshTokenExpired()) return true;

    return false;
  }

  // Clear all authentication data
  clearAuthData() {
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.ACCESS_TOKEN_EXPIRY);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.REFRESH_TOKEN_EXPIRY);
    localStorage.removeItem(LOCAL_STORAGE_KEYS.USER_EMAIL);
  }

  // Store login session data (for multi-step login)
  setLoginSessionData(key, data) {
    localStorage.setItem(`login_session_${key}`, JSON.stringify(data));
  }

  // Get login session data
  getLoginSessionData(key) {
    const data = localStorage.getItem(`login_session_${key}`);
    return data ? JSON.parse(data) : null;
  }

  // Clear login session data
  clearLoginSessionData(key) {
    localStorage.removeItem(`login_session_${key}`);
  }

  // Clear all login session data
  clearAllLoginSessionData() {
    const keys = Object.keys(localStorage);
    keys.forEach((key) => {
      if (key.startsWith("login_session_")) {
        localStorage.removeItem(key);
      }
    });
  }
}

// Export singleton instance
export default new LocalStorageService();
