// src/services/userApi.js
import { mapleFileApi } from "./apiConfig";

/**
 * User profile API endpoints
 */
export const userAPI = {
  // Get the current user's profile
  getProfile: () => {
    return mapleFileApi.get("/me");
  },

  // Update the current user's profile
  updateProfile: (profileData) => {
    return mapleFileApi.put("/me", profileData);
  },
};

export default userAPI;
