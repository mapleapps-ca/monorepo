// src/services/TokenService.js
class TokenService {
  constructor() {
    this.ACCESS_TOKEN_KEY = "access_token";
    this.REFRESH_TOKEN_KEY = "refresh_token";
  }

  // Get access token
  getAccessToken() {
    return localStorage.getItem(this.ACCESS_TOKEN_KEY);
  }

  // Get refresh token
  getRefreshToken() {
    return localStorage.getItem(this.REFRESH_TOKEN_KEY);
  }

  // Save tokens
  setTokens(accessToken, refreshToken) {
    localStorage.setItem(this.ACCESS_TOKEN_KEY, accessToken);
    localStorage.setItem(this.REFRESH_TOKEN_KEY, refreshToken);
  }

  // Update only access token (used after refresh)
  setAccessToken(accessToken) {
    localStorage.setItem(this.ACCESS_TOKEN_KEY, accessToken);
  }

  // Clear all tokens
  clearTokens() {
    localStorage.removeItem(this.ACCESS_TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_TOKEN_KEY);
  }

  // Check if we have valid tokens
  hasTokens() {
    return this.getAccessToken() !== null && this.getRefreshToken() !== null;
  }

  // Decode JWT token to get payload (without verification)
  decodeToken(token) {
    try {
      const base64Url = token.split(".")[1];
      const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split("")
          .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
          .join(""),
      );
      return JSON.parse(jsonPayload);
    } catch (error) {
      console.error("Error decoding token:", error);
      return null;
    }
  }

  // Check if access token is expired
  isAccessTokenExpired() {
    const token = this.getAccessToken();
    if (!token) return true;

    const decoded = this.decodeToken(token);
    if (!decoded || !decoded.exp) return true;

    // Check if token is expired (with 30 second buffer)
    const expirationTime = decoded.exp * 1000; // Convert to milliseconds
    const currentTime = Date.now();
    const bufferTime = 30 * 1000; // 30 seconds buffer

    return currentTime > expirationTime - bufferTime;
  }

  // Get user info from token
  getUserFromToken() {
    const token = this.getAccessToken();
    if (!token) return null;

    const decoded = this.decodeToken(token);
    if (!decoded) return null;

    // Adjust these fields based on your JWT payload structure
    return {
      id: decoded.sub || decoded.userId,
      email: decoded.email,
      name: decoded.name,
      // Add other fields as needed
    };
  }
}

export default TokenService;
