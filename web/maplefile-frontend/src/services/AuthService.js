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
      // Generate crypto fields
      const cryptoData =
        await this.cryptoService.generateRegistrationCrypto(password);

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
      // Handle different types of errors
      if (error.response) {
        // Server responded with error status
        const errorData = error.response.data;
        if (errorData.details) {
          // Validation errors
          const errorMessages = Object.values(errorData.details).join(", ");
          throw new Error(errorMessages);
        } else {
          throw new Error(errorData.message || "Registration failed");
        }
      } else if (error.request) {
        // Network error
        throw new Error(
          "Network error. Please check your connection and try again.",
        );
      } else {
        // Other error (including crypto generation errors)
        throw new Error(error.message || "Registration failed");
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
  }

  // Get the current logged in user
  getCurrentUser() {
    return this.currentUser;
  }

  // Check if a user is logged in
  isAuthenticated() {
    return this.currentUser !== null;
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
