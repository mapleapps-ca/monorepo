// src/services/RegistrationService.js

import axios from "axios";

/**
 * RegistrationService handles user registration and email verification
 * It coordinates between the frontend form, cryptographic operations, and the backend API
 */
export class RegistrationService {
  constructor(logger, cryptoService) {
    this.logger = logger;
    this.cryptoService = cryptoService;

    // Get API configuration from environment variables
    // In development: http://127.0.0.1:8000
    // In production: https://mapleapps.net
    this.apiBaseUrl = `${import.meta.env.VITE_API_PROTOCOL}://${import.meta.env.VITE_API_DOMAIN}`;

    this.logger.log(
      `RegistrationService: Initialized with API base URL: ${this.apiBaseUrl}`,
    );
  }

  /**
   * Register a new user with end-to-end encryption
   * This is the main registration function that coordinates all the steps
   */
  async registerUser(formData) {
    this.logger.log("RegistrationService: Starting user registration process");

    try {
      // Step 1: Validate the form data before proceeding
      this.validateRegistrationData(formData);

      // Step 2: Wait for crypto service to be ready
      await this.cryptoService.initSodium();

      // Step 3: Generate all cryptographic keys from the user's password
      this.logger.log("RegistrationService: Generating encryption keys...");
      const cryptoKeys = await this.cryptoService.generateRegistrationKeys(
        formData.password,
      );

      // Step 4: Prepare the registration payload according to the API specification
      const registrationPayload = {
        // Required beta access code - in a real app, users would get this from support
        beta_access_code: "BETA2024",

        // Personal information
        first_name: formData.firstName,
        last_name: formData.lastName,
        email: formData.email.toLowerCase().trim(), // API automatically does this, but let's be explicit
        phone: formData.phone,
        country: formData.country,
        timezone: formData.timezone || "America/Toronto", // Default to Toronto timezone

        // Agreement checkboxes - terms of service is required
        agree_terms_of_service: formData.agreeTermsOfService,
        agree_promotions: formData.agreePromotions || false,
        agree_to_tracking_across_third_party_apps_and_services:
          formData.agreeTracking || false,

        // Module selection: 1 = MapleFile, 2 = PaperCloud
        module: formData.module || 1, // Default to MapleFile

        // Cryptographic data - all the encrypted keys we generated
        salt: cryptoKeys.salt,
        publicKey: cryptoKeys.publicKey,
        encryptedMasterKey: cryptoKeys.encryptedMasterKey,
        encryptedPrivateKey: cryptoKeys.encryptedPrivateKey,
        encryptedRecoveryKey: cryptoKeys.encryptedRecoveryKey,
        masterKeyEncryptedWithRecoveryKey:
          cryptoKeys.masterKeyEncryptedWithRecoveryKey,
        verificationID: cryptoKeys.verificationID,
      };

      // Step 5: Submit the registration to the API
      this.logger.log("RegistrationService: Submitting registration to API...");
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/register`,
        registrationPayload,
        {
          headers: {
            "Content-Type": "application/json",
          },
          timeout: 30000, // 30 second timeout for registration
        },
      );

      // Step 6: Handle successful registration
      if (response.status === 201) {
        this.logger.log(
          "RegistrationService: Registration completed successfully",
        );
        return {
          success: true,
          message: response.data.message,
          recoveryKeyInfo: response.data.recovery_key_info,
          email: formData.email,
        };
      } else {
        throw new Error(`Unexpected response status: ${response.status}`);
      }
    } catch (error) {
      return this.handleRegistrationError(error);
    }
  }

  /**
   * Verify email address using the code sent during registration
   */
  async verifyEmail(verificationCode) {
    this.logger.log(
      `RegistrationService: Verifying email with code: ${verificationCode}`,
    );

    try {
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/verify-email-code`,
        {
          code: verificationCode.trim(),
        },
        {
          headers: {
            "Content-Type": "application/json",
          },
          timeout: 10000, // 10 second timeout for verification
        },
      );

      if (response.status === 201) {
        this.logger.log("RegistrationService: Email verification successful");
        return {
          success: true,
          message: response.data.message,
          userRole: response.data.user_role,
        };
      } else {
        throw new Error(`Unexpected response status: ${response.status}`);
      }
    } catch (error) {
      return this.handleVerificationError(error);
    }
  }

  /**
   * Validate registration form data before submitting
   * This catches common errors early and provides helpful feedback
   */
  validateRegistrationData(formData) {
    const requiredFields = [
      "firstName",
      "lastName",
      "email",
      "phone",
      "country",
      "password",
    ];

    for (const field of requiredFields) {
      if (!formData[field] || formData[field].trim() === "") {
        throw new Error(`${field} is required`);
      }
    }

    // Email format validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(formData.email)) {
      throw new Error("Please enter a valid email address");
    }

    // Email length validation (API limit is 255 characters)
    if (formData.email.length > 255) {
      throw new Error("Email address is too long (maximum 255 characters)");
    }

    // Password strength validation
    if (formData.password.length < 8) {
      throw new Error("Password must be at least 8 characters long");
    }

    // Terms of service agreement validation
    if (!formData.agreeTermsOfService) {
      throw new Error("You must agree to the terms of service to register");
    }

    this.logger.log("RegistrationService: Form data validation passed");
  }

  /**
   * Handle registration errors with user-friendly messages
   */
  handleRegistrationError(error) {
    this.logger.log(
      `RegistrationService: Registration error: ${error.message}`,
    );

    if (error.response) {
      // The server responded with an error status
      const status = error.response.status;
      const data = error.response.data;

      if (status === 400 && data.details) {
        // Handle field-specific validation errors from the API
        const fieldErrors = data.details;
        const errorMessages = Object.values(fieldErrors).join(", ");
        return {
          success: false,
          message: `Registration failed: ${errorMessages}`,
          fieldErrors: fieldErrors,
        };
      } else if (status === 400) {
        return {
          success: false,
          message: data.error || "Registration data is invalid",
        };
      } else if (status === 500) {
        return {
          success: false,
          message: "Server error during registration. Please try again later.",
        };
      }
    } else if (error.request) {
      // Network error - no response received
      return {
        success: false,
        message:
          "Unable to connect to the registration server. Please check your internet connection and try again.",
      };
    }

    // Generic error fallback
    return {
      success: false,
      message:
        error.message || "An unexpected error occurred during registration",
    };
  }

  /**
   * Handle email verification errors
   */
  handleVerificationError(error) {
    this.logger.log(
      `RegistrationService: Email verification error: ${error.message}`,
    );

    if (error.response) {
      const status = error.response.status;
      const data = error.response.data;

      if (status === 400) {
        if (data.details && data.details.code) {
          return {
            success: false,
            message:
              "Verification code is invalid or has expired. Please check your email and try again.",
          };
        }
        return {
          success: false,
          message: data.error || "Invalid verification request",
        };
      }
    } else if (error.request) {
      return {
        success: false,
        message:
          "Unable to connect to the verification server. Please check your internet connection and try again.",
      };
    }

    return {
      success: false,
      message:
        error.message ||
        "An unexpected error occurred during email verification",
    };
  }

  /**
   * Get a list of common countries for the registration form
   * In a real application, this might come from an API or a more comprehensive list
   */
  getCountryList() {
    return [
      "Canada",
      "United States",
      "United Kingdom",
      "Australia",
      "Germany",
      "France",
      "Japan",
      "Brazil",
      "Mexico",
      "India",
      "Other",
    ];
  }

  /**
   * Get a list of common timezones
   * This is a simplified list - a real app might use a timezone library
   */
  getTimezoneList() {
    return [
      { value: "America/Toronto", label: "Eastern Time (Toronto)" },
      { value: "America/Vancouver", label: "Pacific Time (Vancouver)" },
      { value: "America/New_York", label: "Eastern Time (New York)" },
      { value: "America/Los_Angeles", label: "Pacific Time (Los Angeles)" },
      { value: "America/Chicago", label: "Central Time (Chicago)" },
      { value: "Europe/London", label: "Greenwich Mean Time (London)" },
      { value: "Europe/Paris", label: "Central European Time (Paris)" },
      { value: "Asia/Tokyo", label: "Japan Standard Time (Tokyo)" },
      { value: "Australia/Sydney", label: "Australian Eastern Time (Sydney)" },
    ];
  }
}
