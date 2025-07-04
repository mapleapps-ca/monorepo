// File: monorepo/web/maplefile-frontend/src/services/Storage/AuthStorageService.js
// Authentication Storage Service - Handles all local storage operations for authentication
import LocalStorageService from "../LocalStorageService.js";

class AuthStorageService {
  constructor() {
    console.log("[AuthStorageService] Storage service initialized");
  }

  // === User Email Management ===

  setUserEmail(email) {
    if (email) {
      LocalStorageService.setUserEmail(email);
      console.log("[AuthStorageService] User email stored");
    }
  }

  getUserEmail() {
    return LocalStorageService.getUserEmail();
  }

  // === Token Management ===

  setTokens(accessToken, refreshToken, accessTokenExpiry, refreshTokenExpiry) {
    LocalStorageService.setTokens(
      accessToken,
      refreshToken,
      accessTokenExpiry,
      refreshTokenExpiry,
    );
    console.log("[AuthStorageService] Tokens stored successfully");
  }

  getAccessToken() {
    return LocalStorageService.getAccessToken();
  }

  getRefreshToken() {
    return LocalStorageService.getRefreshToken();
  }

  // === Token Status Checking ===

  isAuthenticated() {
    return LocalStorageService.isAuthenticated();
  }

  hasValidTokens() {
    return LocalStorageService.hasValidTokens();
  }

  isAccessTokenExpired() {
    return LocalStorageService.isAccessTokenExpired();
  }

  isAccessTokenExpiringSoon(minutesBeforeExpiry = 5) {
    return LocalStorageService.isAccessTokenExpiringSoon(minutesBeforeExpiry);
  }

  isRefreshTokenExpired() {
    return LocalStorageService.isRefreshTokenExpired();
  }

  getTokenExpiryInfo() {
    return LocalStorageService.getTokenExpiryInfo();
  }

  // === User Encrypted Data Management ===

  storeUserEncryptedData(
    salt,
    encryptedMasterKey,
    encryptedPrivateKey,
    publicKey,
  ) {
    LocalStorageService.storeUserEncryptedData(
      salt,
      encryptedMasterKey,
      encryptedPrivateKey,
      publicKey,
    );
    console.log(
      "[AuthStorageService] User encrypted data stored for future password-based decryption",
    );
  }

  getUserEncryptedData() {
    return LocalStorageService.getUserEncryptedData();
  }

  hasUserEncryptedData() {
    return LocalStorageService.hasUserEncryptedData();
  }

  // === Session Key Management (Memory Only) ===

  setSessionKeys(masterKey, privateKey, publicKey, keyEncryptionKey) {
    LocalStorageService.setSessionKeys(
      masterKey,
      privateKey,
      publicKey,
      keyEncryptionKey,
    );
    console.log(
      "[AuthStorageService] Session keys cached in memory for token decryption",
    );
  }

  hasSessionKeys() {
    return LocalStorageService.hasSessionKeys();
  }

  getSessionKeys() {
    return LocalStorageService.getSessionKeys();
  }

  clearSessionKeys() {
    LocalStorageService.clearSessionKeys();
    console.log("[AuthStorageService] Session keys cleared from memory");
  }

  // === Login Session Data (Multi-step login) ===

  setLoginSessionData(key, data) {
    LocalStorageService.setLoginSessionData(key, data);
    console.log(`[AuthStorageService] Login session data stored: ${key}`);
  }

  getLoginSessionData(key) {
    return LocalStorageService.getLoginSessionData(key);
  }

  clearLoginSessionData(key) {
    LocalStorageService.clearLoginSessionData(key);
    console.log(`[AuthStorageService] Login session data cleared: ${key}`);
  }

  clearAllLoginSessionData() {
    LocalStorageService.clearAllLoginSessionData();
    console.log("[AuthStorageService] All login session data cleared");
  }

  // === Token Decryption Support ===

  async decryptTokensFromLogin(encryptedTokensData, tokenNonce) {
    try {
      console.log("[AuthStorageService] Decrypting tokens from login response");
      return await LocalStorageService.decryptTokensFromLogin(
        encryptedTokensData,
        tokenNonce,
      );
    } catch (error) {
      console.error("[AuthStorageService] Token decryption failed:", error);
      throw error;
    }
  }

  // === Cleanup and Maintenance ===

  clearAuthData() {
    LocalStorageService.clearAuthData();
    console.log("[AuthStorageService] All authentication data cleared");
  }

  cleanupEncryptedTokenData() {
    const hadData = LocalStorageService.cleanupEncryptedTokenData();
    if (hadData) {
      console.log("[AuthStorageService] Old encrypted token data cleaned up");
    }
    return hadData;
  }

  // === Storage Information ===

  getAllStorageData() {
    return LocalStorageService.getAllStorageData();
  }

  getStorageInfo() {
    const tokenInfo = this.getTokenExpiryInfo();
    const userEmail = this.getUserEmail();
    const hasValidTokens = this.hasValidTokens();
    const hasUserData = this.hasUserEncryptedData();
    const hasSessionKeys = this.hasSessionKeys();

    return {
      // Authentication status
      isAuthenticated: this.isAuthenticated(),
      hasValidTokens,
      userEmail,

      // Token information
      tokenInfo,

      // User data
      hasUserEncryptedData: hasUserData,
      userEncryptedData: hasUserData
        ? {
            hasSalt: !!this.getUserEncryptedData().salt,
            hasEncryptedMasterKey:
              !!this.getUserEncryptedData().encryptedMasterKey,
            hasEncryptedPrivateKey:
              !!this.getUserEncryptedData().encryptedPrivateKey,
            hasPublicKey: !!this.getUserEncryptedData().publicKey,
          }
        : null,

      // Session keys (memory only)
      hasSessionKeys,

      // Storage keys in localStorage
      storageKeys: {
        accessToken: !!this.getAccessToken(),
        refreshToken: !!this.getRefreshToken(),
        userEmail: !!userEmail,
        accessTokenExpiry: !!localStorage.getItem(
          "mapleapps_access_token_expiry",
        ),
        refreshTokenExpiry: !!localStorage.getItem(
          "mapleapps_refresh_token_expiry",
        ),
        userSalt: !!localStorage.getItem("mapleapps_user_salt"),
        userEncryptedMasterKey: !!localStorage.getItem(
          "mapleapps_user_encrypted_master_key",
        ),
        userEncryptedPrivateKey: !!localStorage.getItem(
          "mapleapps_user_encrypted_private_key",
        ),
        userPublicKey: !!localStorage.getItem("mapleapps_user_public_key"),
      },
    };
  }

  // === Utility Methods ===

  canMakeAuthenticatedRequests() {
    return this.hasValidTokens();
  }

  shouldRefreshTokens(minutesBeforeExpiry = 5) {
    return this.isAccessTokenExpiringSoon(minutesBeforeExpiry);
  }

  getSessionKeyStatus() {
    return {
      hasSessionKeys: this.hasSessionKeys(),
      hasUserEncryptedData: this.hasUserEncryptedData(),
      isAuthenticated: this.isAuthenticated(),
      canMakeRequests: this.canMakeAuthenticatedRequests(),
    };
  }

  // === Debug Information ===

  getDebugInfo() {
    return {
      serviceName: "AuthStorageService",
      storageInfo: this.getStorageInfo(),
      sessionKeyStatus: this.getSessionKeyStatus(),
      canMakeAuthenticatedRequests: this.canMakeAuthenticatedRequests(),
      shouldRefreshTokens: this.shouldRefreshTokens(),
    };
  }
}

export default AuthStorageService;
