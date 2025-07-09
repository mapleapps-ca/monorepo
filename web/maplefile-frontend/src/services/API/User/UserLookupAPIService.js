// File: monorepo/web/maplefile-frontend/src/services/API/User/UserLookupAPIService.js
// User Lookup API Service - Handles user public key lookup API calls with proper error handling

class UserLookupAPIService {
  constructor() {
    this._apiClient = null;
    console.log(
      "[UserLookupAPIService] API service initialized for public user lookups",
    );
  }

  // Import ApiClient for public requests
  async getApiClient() {
    if (!this._apiClient) {
      const { default: ApiClient } = await import("../ApiClient.js");
      this._apiClient = ApiClient;
    }
    return this._apiClient;
  }

  // Validate email format
  validateEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  // Sanitize email (lowercase and trim)
  sanitizeEmail(email) {
    return email.toLowerCase().trim();
  }

  // Lookup user public key by email with proper error handling
  async lookupUser(email) {
    try {
      console.log("[UserLookupAPIService] Looking up user by email:", email);

      if (!email) {
        throw new Error("Email address is required");
      }

      // Sanitize email
      const sanitizedEmail = this.sanitizeEmail(email);

      // Validate email format
      if (!this.validateEmail(sanitizedEmail)) {
        throw new Error("Invalid email format");
      }

      // Make direct fetch call to handle errors properly
      const response = await fetch(
        `/iam/api/v1/users/lookup?email=${encodeURIComponent(sanitizedEmail)}`,
      );

      if (!response.ok) {
        // Try to get the error message from the response
        let errorMessage = `Lookup failed (${response.status})`;

        try {
          const errorData = await response.json();
          if (errorData.email) {
            errorMessage = errorData.email; // "Email address does not exist: ..."
          } else if (errorData.error) {
            errorMessage = errorData.error;
          } else if (errorData.details) {
            errorMessage = Object.values(errorData.details)[0];
          }
        } catch (parseError) {
          // If we can't parse the error, use status-based messages
          if (response.status === 400) {
            errorMessage = `User not found: ${sanitizedEmail}`;
          } else if (response.status === 500) {
            errorMessage = "Server error - please try again later";
          }
        }

        throw new Error(errorMessage);
      }

      const user = await response.json();

      console.log("[UserLookupAPIService] User lookup successful:", {
        email: user.email,
        userId: user.user_id,
        name: user.name,
        hasPublicKey: !!user.public_key_in_base64,
        hasVerificationId: !!user.verification_id,
      });

      return user;
    } catch (error) {
      console.error("[UserLookupAPIService] User lookup failed:", error);
      throw error; // Re-throw with the original message
    }
  }

  // Check if user exists (lightweight check)
  async userExists(email) {
    try {
      console.log("[UserLookupAPIService] Checking if user exists:", email);
      await this.lookupUser(email);
      return true;
    } catch (error) {
      if (
        error.message.includes("not found") ||
        error.message.includes("does not exist")
      ) {
        return false;
      }
      // Re-throw other errors (network, validation, etc.)
      throw error;
    }
  }

  // Get debug information
  getDebugInfo() {
    return {
      serviceName: "UserLookupAPIService",
      type: "public_api_service",
      authRequired: false,
      supportedOperations: ["lookupUser", "userExists"],
      endpoints: ["/iam/api/v1/users/lookup"],
    };
  }
}

export default UserLookupAPIService;
