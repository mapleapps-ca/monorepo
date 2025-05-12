// src/services/userApi.js
import { paperCloudApi } from "./apiConfig";

/**
 * User profile API endpoints
 */
export const userAPI = {
  // Get the current user's profile
  getProfile: () => {
    return paperCloudApi.get("/me");
  },

  // Update the current user's profile
  updateProfile: (profileData) => {
    return paperCloudApi.put("/me", profileData);
  },
};

export default userAPI;
