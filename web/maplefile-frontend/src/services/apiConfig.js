// src/services/apiConfig.js
import axios from "axios";
import tokenManager from "./TokenManager";

export const createApiInstance = (baseURL) => {
  const instance = axios.create({
    baseURL,
    // Don't set default Content-Type here
  });

  // Add request interceptor to include auth token and set appropriate content type
  instance.interceptors.request.use(
    (config) => {
      const token = tokenManager.getAccessToken();
      if (token) {
        config.headers["Authorization"] = `JWT ${token}`;
      }

      // Only set Content-Type for non-FormData requests
      if (!(config.data instanceof FormData)) {
        config.headers["Content-Type"] = "application/json";
      }

      return config;
    },
    (error) => {
      return Promise.reject(error);
    },
  );

  // Add response interceptor to handle 401 errors
  instance.interceptors.response.use(
    (response) => {
      return response;
    },
    async (error) => {
      // If the server returned a 401 error, redirect to login
      if (error.response && error.response.status === 401) {
        console.error("Unauthorized API call - redirecting to login", error);
        tokenManager.clearTokens();
        tokenManager.redirectToLogin();
      }
      return Promise.reject(error);
    },
  );

  return instance;
};

// Create instances for different API services
export const iamApi = createApiInstance("/iam/api/v1");
export const mapleFileApi = createApiInstance("/maplefile/api/v1");

export default {
  iamApi,
  mapleFileApi,
};
