// src/pages/anonymous/Login/completeLoginPage.jsx

import { useState, useEffect, useRef } from "react";
import { useNavigate, useLocation } from "react-router";
import { useService } from "../../../hooks/useService.js";
import { TYPES } from "../../../di/types.js";

/**
 * CompleteLoginPage - Step 3 of the Login Process (Password Verification)
 *
 * This component demonstrates the most sophisticated part of our authentication system.
 * It's where we prove knowledge of the password through cryptographic challenge-response,
 * without ever transmitting the password itself. This page showcases several advanced concepts:
 *
 * 1. Zero-knowledge password verification through cryptographic challenges
 * 2. Client-side key derivation and decryption operations
 * 3. Progressive feedback during computationally intensive operations
 * 4. Secure session establishment with encrypted token handling
 * 5. Error handling for cryptographic failures
 *
 * The cryptographic process works like this:
 * - User enters password → derive encryption key using Argon2ID
 * - Use derived key → decrypt master key → decrypt private key
 * - Use private key → decrypt the challenge from the server
 * - Send decrypted challenge → server verifies → issues auth tokens
 *
 * This approach ensures that even if network traffic is intercepted, the password
 * remains secure because it never leaves the user's device. The server verifies
 * authentication by checking if the user could decrypt the challenge, which proves
 * they had the correct password and encryption keys.
 */
function CompleteLoginPage() {
  // Get our services from dependency injection
  const authService = useService(TYPES.AuthService);
  const logger = useService(TYPES.LoggerService);
  const navigate = useNavigate();
  const location = useLocation();

  // Extract navigation state from previous steps
  const email = location.state?.email || "";
  const initialMessage =
    location.state?.message || "Please enter your password to complete login";

  // Form state management
  const [formData, setFormData] = useState({
    password: "",
  });

  // UI state for password visibility and loading states
  const [uiState, setUiState] = useState({
    showPassword: false,
    isSubmitting: false,
    cryptoStage: null, // Tracks which crypto operation is happening
  });

  // Message and error handling
  const [message, setMessage] = useState(initialMessage);
  const [error, setError] = useState(null);

  // Ref for password input focus management
  const passwordInputRef = useRef(null);

  // Component initialization and flow validation
  useEffect(() => {
    logger.log(`CompleteLoginPage: Component mounted for email: ${email}`);

    // Validate that we're in the correct step of the login process
    const loginState = authService.getLoginState();

    if (!email) {
      // No email means we shouldn't be here
      logger.log(
        "CompleteLoginPage: No email provided, redirecting to login start",
      );
      navigate("/login", {
        state: {
          message: "Please start the login process by entering your email",
        },
      });
      return;
    }

    if (loginState.step < 2) {
      // If we haven't completed email verification, redirect appropriately
      logger.log(
        `CompleteLoginPage: Login process not at correct step (${loginState.step}), redirecting`,
      );

      if (loginState.step === 0) {
        navigate("/login", {
          state: {
            email: email,
            message: "Please request a verification code to continue",
          },
        });
      } else if (loginState.step === 1) {
        navigate("/verify-ott", {
          state: {
            email: email,
            message: "Please enter the verification code from your email",
          },
        });
      }
      return;
    }

    // Focus the password input for better user experience
    if (passwordInputRef.current) {
      passwordInputRef.current.focus();
    }
  }, [email, authService, logger, navigate]);

  /**
   * Handle input changes with real-time feedback
   * This provides immediate response as users type their password
   */
  const handleInputChange = (field, value) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear errors when user starts typing
    // This creates a responsive feel and encourages users to try again
    if (error) {
      setError(null);
    }

    // Clear informational messages when user starts interacting
    if (message) {
      setMessage(null);
    }
  };

  /**
   * Toggle password visibility
   * This improves usability for users with complex passwords
   */
  const togglePasswordVisibility = () => {
    setUiState((prev) => ({
      ...prev,
      showPassword: !prev.showPassword,
    }));
  };

  /**
   * Validate the password before submission
   * This prevents unnecessary crypto operations with empty passwords
   */
  const validateForm = () => {
    if (!formData.password) {
      setError("Please enter your password");
      return false;
    }

    // We don't validate password strength here because this is login, not registration
    // The cryptographic operations will fail if the password is incorrect
    return true;
  };

  /**
   * Handle form submission - the final step of authentication
   * This is where the complex cryptographic operations happen
   */
  const handleSubmit = async (e) => {
    e.preventDefault();

    logger.log("CompleteLoginPage: Password submission started");

    // Clear any previous messages
    setError(null);
    setMessage(null);

    // Validate the form before proceeding
    if (!validateForm()) {
      return;
    }

    // Set loading state and prepare for crypto operations
    setUiState((prev) => ({
      ...prev,
      isSubmitting: true,
      cryptoStage: "initializing",
    }));

    try {
      // Stage 1: Key Derivation
      // This is computationally intensive and may take 1-3 seconds
      setUiState((prev) => ({ ...prev, cryptoStage: "deriving-keys" }));
      setMessage("Deriving encryption keys from your password...");

      // Give the UI a moment to update before starting intensive operations
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Stage 2: Decryption Operations
      setUiState((prev) => ({ ...prev, cryptoStage: "decrypting" }));
      setMessage("Decrypting your secure data...");

      // Another brief pause for UI responsiveness
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Stage 3: Challenge Response
      setUiState((prev) => ({ ...prev, cryptoStage: "authenticating" }));
      setMessage("Completing secure authentication...");

      // Call the authentication service to complete the login
      // This performs all the cryptographic operations:
      // 1. Derives key from password using Argon2ID
      // 2. Decrypts master key using derived key
      // 3. Decrypts private key using master key
      // 4. Decrypts authentication challenge using private key
      // 5. Submits decrypted challenge to server
      // 6. Receives and handles authentication tokens
      const result = await authService.completeLogin(formData.password);

      if (result.success) {
        // Authentication successful!
        logger.log("CompleteLoginPage: Login completed successfully");

        setUiState((prev) => ({ ...prev, cryptoStage: "success" }));
        setMessage("Login successful! Redirecting to your dashboard...");

        // Brief delay to show success message, then redirect
        setTimeout(() => {
          navigate("/dashboard", {
            replace: true, // Replace history so back button doesn't return to login
          });
        }, 1500);
      } else {
        // Authentication failed
        logger.log(`CompleteLoginPage: Login failed: ${result.message}`);
        setError(result.message);

        // Clear the password field for security and retry
        setFormData((prev) => ({ ...prev, password: "" }));

        // Refocus the password input
        if (passwordInputRef.current) {
          passwordInputRef.current.focus();
        }
      }
    } catch (error) {
      // Handle unexpected errors
      logger.log(`CompleteLoginPage: Unexpected error: ${error.message}`);

      // Provide user-friendly error messages for common issues
      let userMessage = "An unexpected error occurred. Please try again.";

      if (error.message.includes("incorrect password")) {
        userMessage = "Incorrect password. Please try again.";
      } else if (error.message.includes("corrupted")) {
        userMessage = "Account data appears corrupted. Please contact support.";
      } else if (error.message.includes("network")) {
        userMessage =
          "Network error. Please check your connection and try again.";
      }

      setError(userMessage);
      setFormData((prev) => ({ ...prev, password: "" }));

      if (passwordInputRef.current) {
        passwordInputRef.current.focus();
      }
    } finally {
      // Always reset the loading state
      setUiState((prev) => ({
        ...prev,
        isSubmitting: false,
        cryptoStage: null,
      }));
    }
  };

  /**
   * Get user-friendly description of current crypto stage
   * This helps users understand what's happening during the waiting period
   */
  const getCryptoStageDescription = () => {
    switch (uiState.cryptoStage) {
      case "deriving-keys":
        return "This process is intentionally slow for security - please wait...";
      case "decrypting":
        return "Unlocking your encrypted data with your password...";
      case "authenticating":
        return "Proving your identity to the server securely...";
      case "success":
        return "Authentication complete! Preparing your secure session...";
      default:
        return "";
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Complete your login
        </h2>
        <p className="mt-2 text-center text-sm text-gray-600">
          Enter your password for
        </p>
        <p className="text-center text-sm font-medium text-gray-900">{email}</p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          {/* Progress Message */}
          {message && !error && (
            <div className="mb-4 bg-blue-50 border border-blue-200 rounded-md p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  {uiState.isSubmitting ? (
                    <svg
                      className="animate-spin h-5 w-5 text-blue-400"
                      xmlns="http://www.w3.org/2000/svg"
                      fill="none"
                      viewBox="0 0 24 24"
                    >
                      <circle
                        className="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        strokeWidth="4"
                      ></circle>
                      <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                      ></path>
                    </svg>
                  ) : (
                    <svg
                      className="h-5 w-5 text-blue-400"
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path
                        fillRule="evenodd"
                        d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                        clipRule="evenodd"
                      />
                    </svg>
                  )}
                </div>
                <div className="ml-3">
                  <p className="text-sm font-medium text-blue-800">{message}</p>
                  {uiState.cryptoStage && (
                    <p className="mt-1 text-sm text-blue-700">
                      {getCryptoStageDescription()}
                    </p>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="mb-4 bg-red-50 border border-red-200 rounded-md p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg
                    className="h-5 w-5 text-red-400"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                      clipRule="evenodd"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm font-medium text-red-800">{error}</p>
                </div>
              </div>
            </div>
          )}

          {/* Password Form */}
          <form className="space-y-6" onSubmit={handleSubmit}>
            {/* Password Input Field */}
            <div>
              <label
                htmlFor="password"
                className="block text-sm font-medium text-gray-700"
              >
                Password
              </label>
              <div className="mt-1 relative">
                <input
                  ref={passwordInputRef}
                  id="password"
                  name="password"
                  type={uiState.showPassword ? "text" : "password"}
                  autoComplete="current-password"
                  required
                  value={formData.password}
                  onChange={(e) =>
                    handleInputChange("password", e.target.value)
                  }
                  className="appearance-none block w-full px-3 py-2 pr-10 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                  placeholder="Enter your password"
                  disabled={uiState.isSubmitting}
                />

                {/* Password visibility toggle */}
                <button
                  type="button"
                  className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  onClick={togglePasswordVisibility}
                  disabled={uiState.isSubmitting}
                >
                  <span className="text-sm text-gray-500 hover:text-gray-700">
                    {uiState.showPassword ? "Hide" : "Show"}
                  </span>
                </button>
              </div>
              <p className="mt-2 text-sm text-gray-500">
                Your password is used to decrypt your data locally - it never
                leaves your device
              </p>
            </div>

            {/* Submit Button */}
            <div>
              <button
                type="submit"
                disabled={uiState.isSubmitting || !formData.password.trim()}
                className={`w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white ${
                  uiState.isSubmitting || !formData.password.trim()
                    ? "bg-gray-400 cursor-not-allowed"
                    : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                }`}
              >
                {uiState.isSubmitting ? (
                  <>
                    <svg
                      className="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
                      xmlns="http://www.w3.org/2000/svg"
                      fill="none"
                      viewBox="0 0 24 24"
                    >
                      <circle
                        className="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        strokeWidth="4"
                      ></circle>
                      <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                      ></path>
                    </svg>
                    Authenticating...
                  </>
                ) : (
                  "Complete Login"
                )}
              </button>
            </div>
          </form>

          {/* Security Information */}
          <div className="mt-6">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">
                  How this works
                </span>
              </div>
            </div>

            <div className="mt-6 text-center">
              <div className="text-sm text-gray-600 space-y-2">
                <p>🔐 Your password decrypts your data locally</p>
                <p>🔒 Authentication happens through cryptographic proof</p>
                <p>🛡️ Zero-knowledge verification protects your privacy</p>
              </div>
            </div>
          </div>

          {/* Recovery Options */}
          <div className="mt-6">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">
                  Having trouble?
                </span>
              </div>
            </div>

            <div className="mt-6 space-y-3">
              <button
                type="button"
                onClick={() =>
                  navigate("/verify-ott", {
                    state: {
                      email: email,
                      message: "Let's verify your email again",
                    },
                  })
                }
                className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                disabled={uiState.isSubmitting}
              >
                Start over with email verification
              </button>

              <button
                type="button"
                onClick={() => navigate("/login")}
                className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                disabled={uiState.isSubmitting}
              >
                Use a different email address
              </button>
            </div>
          </div>

          {/* Help Information */}
          <div className="mt-6 text-center">
            <p className="text-xs text-gray-500">
              Forgot your password? Contact support for account recovery options
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default CompleteLoginPage;
