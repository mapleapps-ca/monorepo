// src/services/AuthService.js
import axios from "axios";

class AuthService {
  constructor(cryptoService) {
    this.cryptoService = cryptoService;
    // In a real app, this would connect to a backend API
    // For now, we'll use localStorage for login but real API for registration
    this.users = JSON.parse(localStorage.getItem("users")) || [];
    this.currentUser = JSON.parse(localStorage.getItem("currentUser")) || null;

    // Get API configuration from environment
    this.apiBaseUrl = `${import.meta.env.VITE_API_PROTOCOL}://${import.meta.env.VITE_API_DOMAIN}`;
  }

  // Register a new user using real API
  async register(
    email,
    password,
    name,
    phone,
    country,
    timezone = "America/Toronto",
  ) {
    try {
      console.log("Starting registration for:", email);

      // Generate crypto fields
      console.log("Generating crypto data...");
      const cryptoData =
        await this.cryptoService.generateRegistrationCrypto(password);
      console.log("Crypto data generated successfully");

      // Prepare registration payload
      const registrationData = {
        beta_access_code: "BETA2024", // You may want to make this configurable
        first_name: name.split(" ")[0] || name,
        last_name: name.split(" ").slice(1).join(" ") || "",
        email: email.toLowerCase().trim(),
        phone: phone || "+1234567890", // Default phone if not provided
        country: country || "Canada",
        timezone,
        agree_terms_of_service: true,
        agree_promotions: false,
        agree_to_tracking_across_third_party_apps_and_services: false,
        module: 1, // 1 for MapleFile, 2 for PaperCloud
        salt: cryptoData.salt,
        publicKey: cryptoData.publicKey,
        encryptedMasterKey: cryptoData.encryptedMasterKey,
        encryptedPrivateKey: cryptoData.encryptedPrivateKey,
        encryptedRecoveryKey: cryptoData.encryptedRecoveryKey,
        masterKeyEncryptedWithRecoveryKey:
          cryptoData.masterKeyEncryptedWithRecoveryKey,
        verificationID: cryptoData.verificationID,
      };

      console.log("Sending registration request...");

      // Make API call
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/register`,
        registrationData,
        {
          headers: {
            "Content-Type": "application/json",
          },
        },
      );

      if (response.status === 201) {
        console.log("Registration successful");
        // Registration successful, return the response
        return {
          success: true,
          message: response.data.message,
          recoveryKeyInfo: response.data.recovery_key_info,
          email: email,
          name: name,
          // Store crypto data for potential future use
          cryptoData: {
            masterKey: cryptoData._masterKey,
            recoveryKey: cryptoData._recoveryKey,
            privateKey: cryptoData._privateKey,
          },
        };
      } else {
        throw new Error("Registration failed");
      }
    } catch (error) {
      console.error("Registration error details:", error);

      // Handle different types of errors
      if (error.response) {
        // Server responded with error status
        console.error("Server response:", error.response.data);
        const errorData = error.response.data;
        if (errorData.details) {
          // Validation errors
          const errorMessages = Object.values(errorData.details).join(", ");
          throw new Error(errorMessages);
        } else {
          throw new Error(
            errorData.message || errorData.error || "Registration failed",
          );
        }
      } else if (error.request) {
        // Network error
        console.error("Network error:", error.request);
        throw new Error(
          "Network error. Please check your connection and try again.",
        );
      } else {
        // Other error (including crypto generation errors)
        console.error("General error:", error.message);
        throw new Error(error.message || "Registration failed");
      }
    }
  }

  // Step 1: Request One-Time Token
  async requestOTT(email) {
    try {
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/request-ott`,
        {
          email: email.toLowerCase().trim(),
        },
        {
          headers: {
            "Content-Type": "application/json",
          },
        },
      );

      if (response.status === 200) {
        return {
          success: true,
          message: response.data.message,
        };
      } else {
        throw new Error("Failed to request OTT");
      }
    } catch (error) {
      if (error.response) {
        const errorData = error.response.data;
        if (errorData.details) {
          const errorMessage =
            errorData.details.email || "Failed to request OTT";
          throw new Error(errorMessage);
        } else {
          throw new Error(errorData.error || "Failed to request OTT");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your connection and try again.",
        );
      } else {
        throw new Error(error.message || "Failed to request OTT");
      }
    }
  }

  // Step 2: Verify One-Time Token
  async verifyOTT(email, ott) {
    try {
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/verify-ott`,
        {
          email: email.toLowerCase().trim(),
          ott: ott.trim(),
        },
        {
          headers: {
            "Content-Type": "application/json",
          },
        },
      );

      if (response.status === 200) {
        return {
          success: true,
          data: response.data,
        };
      } else {
        throw new Error("Failed to verify OTT");
      }
    } catch (error) {
      if (error.response) {
        const errorData = error.response.data;
        if (errorData.details) {
          const errorMessage =
            errorData.details.ott ||
            errorData.details.email ||
            "Failed to verify OTT";
          throw new Error(errorMessage);
        } else {
          throw new Error(errorData.error || "Failed to verify OTT");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your connection and try again.",
        );
      } else {
        throw new Error(error.message || "Failed to verify OTT");
      }
    }
  }

  // Step 3: Complete Login
  async completeLogin(email, password, ottData) {
    try {
      // Use CryptoService to decrypt keys and challenge
      const decryptedData = await this.cryptoService.processLoginStep2(
        password,
        ottData,
      );

      // Send the decrypted challenge to complete login
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/complete-login`,
        {
          email: email.toLowerCase().trim(),
          challengeId: decryptedData.challengeId,
          decryptedData: decryptedData.decryptedChallenge,
        },
        {
          headers: {
            "Content-Type": "application/json",
          },
        },
      );

      if (response.status === 200) {
        const loginResult = response.data;

        // Handle tokens
        let accessToken, refreshToken;

        if (loginResult.encrypted_tokens && loginResult.token_nonce) {
          // Decrypt tokens using private key
          try {
            const tokens = await this.cryptoService.decryptTokens(
              loginResult.encrypted_tokens,
              loginResult.token_nonce,
              decryptedData.privateKey,
            );
            accessToken = tokens.access_token;
            refreshToken = tokens.refresh_token;
          } catch (decryptError) {
            console.warn("Failed to decrypt tokens, using fallback");
            accessToken = loginResult.access_token;
            refreshToken = loginResult.refresh_token;
          }
        } else {
          // Use plaintext tokens as fallback
          accessToken = loginResult.access_token;
          refreshToken = loginResult.refresh_token;
        }

        // Store tokens and user data
        if (accessToken) {
          localStorage.setItem("access_token", accessToken);
          localStorage.setItem("refresh_token", refreshToken);
          localStorage.setItem(
            "access_token_expiry",
            loginResult.access_token_expiry_time,
          );
          localStorage.setItem(
            "refresh_token_expiry",
            loginResult.refresh_token_expiry_time,
          );
        }

        // Store user crypto data for session
        const userData = {
          email: email,
          masterKey: decryptedData.masterKey,
          privateKey: decryptedData.privateKey,
          publicKey: decryptedData.publicKey,
          loginTime: new Date().toISOString(),
          lastPasswordChange: ottData.last_password_change,
          keyVersion: ottData.current_key_version,
        };

        this.currentUser = userData;
        localStorage.setItem("currentUser", JSON.stringify(userData));

        return {
          success: true,
          user: userData,
          accessToken,
          refreshToken,
        };
      } else {
        throw new Error("Failed to complete login");
      }
    } catch (error) {
      if (error.response) {
        const errorData = error.response.data;
        if (errorData.details) {
          const errorMessage = Object.values(errorData.details).join(", ");
          throw new Error(errorMessage);
        } else {
          throw new Error(errorData.error || "Failed to complete login");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your connection and try again.",
        );
      } else {
        throw new Error(error.message || "Failed to complete login");
      }
    }
  }

  // Login an existing user (keeping existing localStorage logic for now)
  login(email, password) {
    // Find user by email and password
    const user = this.users.find(
      (u) => u.email === email && u.password === password,
    );

    if (!user) {
      throw new Error("Invalid email or password");
    }

    // Set current user
    this.currentUser = user;
    localStorage.setItem("currentUser", JSON.stringify(user));

    return user;
  }

  // Logout the current user
  logout() {
    this.currentUser = null;
    localStorage.removeItem("currentUser");
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("access_token_expiry");
    localStorage.removeItem("refresh_token_expiry");
  }

  // Get the current logged in user
  getCurrentUser() {
    // Try to get from memory first
    if (this.currentUser) {
      return this.currentUser;
    }

    // Try to get from localStorage
    const storedUser = localStorage.getItem("currentUser");
    if (storedUser) {
      try {
        this.currentUser = JSON.parse(storedUser);
        return this.currentUser;
      } catch (error) {
        console.error("Failed to parse stored user data");
        this.logout(); // Clear invalid data
        return null;
      }
    }

    return null;
  }

  // Check if a user is logged in
  isAuthenticated() {
    const user = this.getCurrentUser();
    const accessToken = localStorage.getItem("access_token");
    const tokenExpiry = localStorage.getItem("access_token_expiry");

    if (!user || !accessToken) {
      return false;
    }

    // Check if token is expired
    if (tokenExpiry) {
      const expiryDate = new Date(tokenExpiry);
      const now = new Date();
      if (now >= expiryDate) {
        // Token is expired, logout
        this.logout();
        return false;
      }
    }

    return true;
  }

  // Get access token
  getAccessToken() {
    if (!this.isAuthenticated()) {
      return null;
    }
    return localStorage.getItem("access_token");
  }

  // Check if refresh token is available and valid
  canRefreshToken() {
    const refreshToken = localStorage.getItem("refresh_token");
    const refreshTokenExpiry = localStorage.getItem("refresh_token_expiry");

    if (!refreshToken) {
      return false;
    }

    if (refreshTokenExpiry) {
      const expiryDate = new Date(refreshTokenExpiry);
      const now = new Date();
      if (now >= expiryDate) {
        return false;
      }
    }

    return true;
  }

  // Verify email with code
  async verifyEmail(code) {
    try {
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/verify-email-code`,
        {
          code: code.trim(),
        },
        {
          headers: {
            "Content-Type": "application/json",
          },
        },
      );

      if (response.status === 201) {
        return {
          success: true,
          message: response.data.message,
          userRole: response.data.user_role,
        };
      } else {
        throw new Error("Verification failed");
      }
    } catch (error) {
      if (error.response) {
        const errorData = error.response.data;
        if (errorData.details) {
          const errorMessage = errorData.details.code || "Verification failed";
          throw new Error(errorMessage);
        } else {
          throw new Error(errorData.error || "Verification failed");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your connection and try again.",
        );
      } else {
        throw new Error(error.message || "Verification failed");
      }
    }
  }
}

export default AuthService;
