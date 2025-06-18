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

  // Validate registration input data
  _validateRegistrationData(email, password, name, phone, country) {
    const errors = [];

    // Email validation
    if (!email || !email.trim()) {
      errors.push("Email is required");
    } else {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email.trim())) {
        errors.push("Please enter a valid email address");
      }
    }

    // Password validation
    if (!password || password.length < 8) {
      errors.push("Password must be at least 8 characters long");
    }

    // Name validation
    if (!name || !name.trim()) {
      errors.push("Name is required");
    } else if (name.trim().length < 2) {
      errors.push("Name must be at least 2 characters long");
    }

    // Phone validation
    if (!phone || !phone.trim()) {
      errors.push("Phone number is required");
    } else {
      const phoneRegex = /^[+]?[\d\s\-\(\)]{10,}$/;
      if (!phoneRegex.test(phone.trim())) {
        errors.push("Please enter a valid phone number");
      }
    }

    // Country validation
    if (!country || !country.trim()) {
      errors.push("Country is required");
    }

    if (errors.length > 0) {
      throw new Error(errors.join("; "));
    }
  }

  // Register a new user using real API
  async register(
    email,
    password,
    name,
    phone,
    country,
    timezone = "America/Toronto",
    betaAccessCode,
  ) {
    try {
      console.log("Starting registration process for:", email);

      // Validate input data
      this._validateRegistrationData(email, password, name, phone, country);

      // Ensure crypto service is ready
      await this.cryptoService.ensureReady();

      // Generate crypto fields
      console.log("Generating cryptographic data...");
      const cryptoData =
        await this.cryptoService.generateRegistrationCrypto(password);
      console.log("Cryptographic data generated successfully");

      // Split name into first and last name
      const nameParts = name.trim().split(/\s+/);
      const firstName = nameParts[0] || name.trim();
      const lastName = nameParts.slice(1).join(" ") || "";

      // Prepare registration payload
      const registrationData = {
        beta_access_code: betaAccessCode,
        first_name: firstName,
        last_name: lastName,
        email: email.toLowerCase().trim(),
        phone: phone.trim(),
        country: country.trim(),
        timezone: timezone,
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

      console.log(
        "Sending registration request to:",
        `${this.apiBaseUrl}/iam/api/v1/register`,
      );
      console.log(
        "Registration payload prepared (crypto fields omitted for security)",
      );

      // Make API call
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/register`,
        registrationData,
        this.axiosConfig,
      );

      console.log("Registration API response status:", response.status);

      if (response.status === 201) {
        console.log("Registration successful");

        // Store registration details temporarily for verification process
        const registrationSession = {
          email: email.toLowerCase().trim(),
          name: name.trim(),
          timestamp: new Date().toISOString(),
          cryptoData: {
            masterKey: cryptoData._masterKey,
            recoveryKey: cryptoData._recoveryKey,
            privateKey: cryptoData._privateKey,
          },
        };

        // Store in sessionStorage for the verification process
        sessionStorage.setItem(
          "pendingRegistration",
          JSON.stringify(registrationSession),
        );

        return {
          success: true,
          message:
            response.data.message || "Registration completed successfully!",
          recoveryKeyInfo: response.data.recovery_key_info,
          email: email.toLowerCase().trim(),
          name: name.trim(),
          verificationRequired: true,
        };
      } else {
        throw new Error(`Registration failed with status: ${response.status}`);
      }
    } catch (error) {
      console.error("Registration error details:", error);

      // Handle different types of errors
      if (error.response) {
        // Server responded with error status
        console.error("Server response:", error.response.data);
        const errorData = error.response.data;

        if (errorData.details && typeof errorData.details === "object") {
          // Validation errors - combine all field errors
          const errorMessages = Object.entries(errorData.details)
            .map(([field, message]) => `${field}: ${message}`)
            .join("; ");
          throw new Error(errorMessages);
        } else if (errorData.message) {
          throw new Error(errorData.message);
        } else if (errorData.error) {
          throw new Error(errorData.error);
        } else {
          throw new Error("Registration failed due to server error");
        }
      } else if (error.request) {
        // Network error
        console.error("Network error:", error.request);
        throw new Error(
          "Network error. Please check your internet connection and try again.",
        );
      } else if (error.message) {
        // Validation error or crypto error
        throw error;
      } else {
        // Other error
        console.error("Unknown error:", error);
        throw new Error("Registration failed due to an unexpected error");
      }
    }
  }

  // Verify email with code
  async verifyEmail(code) {
    try {
      console.log("Starting email verification process");

      // Validate input
      if (!code || typeof code !== "string") {
        throw new Error("Verification code is required");
      }

      const trimmedCode = code.trim();
      if (trimmedCode.length !== 6) {
        throw new Error("Verification code must be exactly 6 digits");
      }

      if (!/^\d{6}$/.test(trimmedCode)) {
        throw new Error("Verification code must contain only digits");
      }

      console.log("Sending email verification request");

      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/verify-email-code`,
        {
          code: trimmedCode,
        },
        this.axiosConfig,
      );

      console.log("Email verification API response status:", response.status);

      if (response.status === 201) {
        console.log("Email verification successful");

        // Clean up pending registration data
        sessionStorage.removeItem("pendingRegistration");

        return {
          success: true,
          message: response.data.message || "Email verified successfully!",
          userRole: response.data.user_role,
          verified: true,
        };
      } else {
        throw new Error(
          `Email verification failed with status: ${response.status}`,
        );
      }
    } catch (error) {
      console.error("Email verification error:", error);

      if (error.response) {
        const errorData = error.response.data;
        console.error("Server response:", errorData);

        if (errorData.details && errorData.details.code) {
          throw new Error(errorData.details.code);
        } else if (errorData.message) {
          throw new Error(errorData.message);
        } else if (errorData.error) {
          throw new Error(errorData.error);
        } else {
          throw new Error("Email verification failed");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your internet connection and try again.",
        );
      } else if (error.message) {
        throw error;
      } else {
        throw new Error("Email verification failed due to an unexpected error");
      }
    }
  }

  // Resend email verification code (for future implementation)
  async resendVerificationCode(email) {
    try {
      console.log("Requesting new verification code for:", email);

      if (!email || !email.trim()) {
        throw new Error("Email is required");
      }

      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email.trim())) {
        throw new Error("Please enter a valid email address");
      }

      // This endpoint might not exist yet in your API
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/resend-verification-code`,
        {
          email: email.toLowerCase().trim(),
        },
        this.axiosConfig,
      );

      if (response.status === 200) {
        return {
          success: true,
          message:
            response.data.message || "Verification code sent successfully!",
        };
      } else {
        throw new Error("Failed to resend verification code");
      }
    } catch (error) {
      console.error("Resend verification code error:", error);

      if (error.response) {
        const errorData = error.response.data;
        if (errorData.message) {
          throw new Error(errorData.message);
        } else if (errorData.error) {
          throw new Error(errorData.error);
        } else {
          throw new Error("Failed to resend verification code");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your internet connection and try again.",
        );
      } else {
        throw new Error(error.message || "Failed to resend verification code");
      }
    }
  }

  // Get pending registration data
  getPendingRegistration() {
    try {
      const pending = sessionStorage.getItem("pendingRegistration");
      return pending ? JSON.parse(pending) : null;
    } catch (error) {
      console.error("Error getting pending registration:", error);
      return null;
    }
  }

  // Step 1: Request One-Time Token
  async requestOTT(email) {
    try {
      if (!email || !email.trim()) {
        throw new Error("Email is required");
      }

      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email.trim())) {
        throw new Error("Please enter a valid email address");
      }

      console.log("Requesting OTT for:", email);

      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/request-ott`,
        {
          email: email.toLowerCase().trim(),
        },
        this.axiosConfig,
      );

      if (response.status === 200) {
        console.log("OTT request successful");
        return {
          success: true,
          message:
            response.data.message || "Verification code sent to your email",
        };
      } else {
        throw new Error("Failed to request verification code");
      }
    } catch (error) {
      console.error("OTT request error:", error);

      if (error.response) {
        const errorData = error.response.data;
        if (errorData.details && errorData.details.email) {
          throw new Error(errorData.details.email);
        } else if (errorData.message) {
          throw new Error(errorData.message);
        } else if (errorData.error) {
          throw new Error(errorData.error);
        } else {
          throw new Error("Failed to request verification code");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your internet connection and try again.",
        );
      } else {
        throw new Error(error.message || "Failed to request verification code");
      }
    }
  }

  // Step 2: Verify One-Time Token
  async verifyOTT(email, ott) {
    try {
      if (!email || !email.trim()) {
        throw new Error("Email is required");
      }

      if (!ott || ott.trim().length !== 6) {
        throw new Error("Please enter the 6-digit verification code");
      }

      if (!/^\d{6}$/.test(ott.trim())) {
        throw new Error("Verification code must contain only digits");
      }

      console.log("Verifying OTT for:", email);

      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/verify-ott`,
        {
          email: email.toLowerCase().trim(),
          ott: ott.trim(),
        },
        this.axiosConfig,
      );

      if (response.status === 200) {
        console.log("OTT verification successful");
        return {
          success: true,
          data: response.data,
        };
      } else {
        throw new Error("Failed to verify code");
      }
    } catch (error) {
      console.error("OTT verification error:", error);

      if (error.response) {
        const errorData = error.response.data;
        if (errorData.details) {
          const errorMessage =
            errorData.details.ott ||
            errorData.details.email ||
            "Failed to verify code";
          throw new Error(errorMessage);
        } else if (errorData.message) {
          throw new Error(errorData.message);
        } else if (errorData.error) {
          throw new Error(errorData.error);
        } else {
          throw new Error("Failed to verify code");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your internet connection and try again.",
        );
      } else {
        throw new Error(error.message || "Failed to verify code");
      }
    }
  }

  // Step 3: Complete Login
  async completeLogin(email, password, ottData) {
    try {
      if (!email || !password || !ottData) {
        throw new Error("Email, password, and verification data are required");
      }

      console.log("Completing login for:", email);

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
        this.axiosConfig,
      );

      if (response.status === 200) {
        const loginResult = response.data;
        console.log("Login completion successful");

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
            console.log("Tokens decrypted successfully");
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
          if (refreshToken) {
            localStorage.setItem("refresh_token", refreshToken);
          }
          if (loginResult.access_token_expiry_time) {
            localStorage.setItem(
              "access_token_expiry",
              loginResult.access_token_expiry_time,
            );
          }
          if (loginResult.refresh_token_expiry_time) {
            localStorage.setItem(
              "refresh_token_expiry",
              loginResult.refresh_token_expiry_time,
            );
          }
        }

        // Store user crypto data for session
        const userData = {
          email: email.toLowerCase().trim(),
          masterKey: decryptedData.masterKey,
          privateKey: decryptedData.privateKey,
          publicKey: decryptedData.publicKey,
          loginTime: new Date().toISOString(),
          lastPasswordChange: ottData.last_password_change,
          keyVersion: ottData.current_key_version,
        };

        this.currentUser = userData;
        localStorage.setItem("currentUser", JSON.stringify(userData));

        console.log("Login process completed successfully");

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
      console.error("Login completion error:", error);

      if (error.response) {
        const errorData = error.response.data;
        if (errorData.details) {
          const errorMessage = Object.values(errorData.details).join("; ");
          throw new Error(errorMessage);
        } else if (errorData.message) {
          throw new Error(errorData.message);
        } else if (errorData.error) {
          throw new Error(errorData.error);
        } else {
          throw new Error("Failed to complete login");
        }
      } else if (error.request) {
        throw new Error(
          "Network error. Please check your internet connection and try again.",
        );
      } else {
        throw new Error(error.message || "Failed to complete login");
      }
    }
  }

  // Login an existing user (keeping existing localStorage logic for backward compatibility)
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
    console.log("Logging out user");
    this.currentUser = null;
    localStorage.removeItem("currentUser");
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("access_token_expiry");
    localStorage.removeItem("refresh_token_expiry");
    sessionStorage.removeItem("pendingRegistration");
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
        console.error("Failed to parse stored user data:", error);
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
      try {
        const expiryDate = new Date(tokenExpiry);
        const now = new Date();
        if (now >= expiryDate) {
          console.log("Access token expired, logging out");
          this.logout();
          return false;
        }
      } catch (error) {
        console.error("Error parsing token expiry:", error);
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
      try {
        const expiryDate = new Date(refreshTokenExpiry);
        const now = new Date();
        if (now >= expiryDate) {
          return false;
        }
      } catch (error) {
        console.error("Error parsing refresh token expiry:", error);
        return false;
      }
    }

    return true;
  }

  // Get registration status
  getRegistrationStatus() {
    const pending = this.getPendingRegistration();
    return {
      hasPendingRegistration: !!pending,
      pendingEmail: pending?.email,
      registrationTime: pending?.timestamp,
    };
  }
}

export default AuthService;
