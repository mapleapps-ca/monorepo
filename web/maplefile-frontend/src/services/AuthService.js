// src/services/AuthService.js
import axios from "axios";

class AuthService {
  constructor(cryptoService) {
    this.cryptoService = cryptoService;

    // Use proxy paths in development, full URLs in production
    const isDevelopment = import.meta.env.DEV;

    if (isDevelopment) {
      // Use Vite proxy paths (no protocol/domain needed)
      this.apiBaseUrl = "";
      console.log("Using Vite dev proxy for API calls");
    } else {
      // Use full URLs in production
      this.apiBaseUrl = `${import.meta.env.VITE_API_PROTOCOL}://${import.meta.env.VITE_API_DOMAIN}`;
      console.log("Using full API URL:", this.apiBaseUrl);
    }

    // Configure axios timeout
    this.axiosConfig = {
      timeout: 30000,
      headers: {
        "Content-Type": "application/json",
      },
    };
  }
}

export default AuthService;
