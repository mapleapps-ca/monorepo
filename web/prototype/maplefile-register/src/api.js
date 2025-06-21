// monorepo/web/prototype/maplefile-register/src/api.js

// API utility functions for MapleApps backend

// Use relative URLs in development (goes through Vite proxy) and full URLs in production
const isDevelopment = import.meta.env.DEV;
const API_BASE_URL = isDevelopment ? "" : "https://api.mapleapps.ca";

console.log(
  "API Mode:",
  isDevelopment ? "Development (using Vite proxy)" : "Production",
);
console.log("API Base URL:", API_BASE_URL || "Relative URLs via proxy");

export const registerUser = async (userData) => {
  try {
    const url = `${API_BASE_URL}/iam/api/v1/register`;
    console.log("Making registration request to:", url);

    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(userData),
    });

    console.log("Registration response status:", response.status);
    console.log(
      "Registration response headers:",
      Object.fromEntries(response.headers),
    );

    const result = await response.json();

    if (!response.ok) {
      console.error("Registration failed with status:", response.status);
      console.error("Error details:", result);
      throw new Error(
        result.details
          ? JSON.stringify(result.details)
          : result.error || "Registration failed",
      );
    }

    return result;
  } catch (error) {
    console.error("Registration error:", error);
    throw error;
  }
};

export const verifyEmail = async (verificationCode) => {
  try {
    const url = `${API_BASE_URL}/iam/api/v1/verify-email-code`;
    console.log("Making email verification request to:", url);

    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        code: verificationCode.trim(),
      }),
    });

    console.log("Email verification response status:", response.status);

    const result = await response.json();

    if (!response.ok) {
      console.error("Email verification failed with status:", response.status);
      console.error("Error details:", result);
      throw new Error(
        result.details?.code || result.error || "Email verification failed",
      );
    }

    return result;
  } catch (error) {
    console.error("Email verification error:", error);
    throw error;
  }
};
