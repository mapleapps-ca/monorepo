// src/services/authApi.js
import { iamApi } from "./apiConfig";

/**
 * Authentication API endpoints
 */
export const authAPI = {
  // Register a new user
  register: (userData) => {
    return iamApi.post("/register", userData);
  },

  // Request a one-time token for login
  requestOTT: (email) => {
    return iamApi.post("/request-ott", { email });
  },

  // Verify the one-time token
  verifyOTT: (email, ott) => {
    return iamApi.post("/verify-ott", { email, ott });
  },

  // Complete login with the decrypted challenge
  completeLogin: (email, challengeId, decryptedData) => {
    return iamApi.post("/complete-login", {
      email,
      challengeId,
      decryptedData,
    });
  },

  // Log out the current user
  logout: () => {
    return iamApi.post("/logout");
  },

  // Refresh the access token
  refreshToken: (refreshToken) => {
    return iamApi.post("/token/refresh", { value: refreshToken });
  },
};

export default authAPI;
