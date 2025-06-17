// src/pages/anonymous/Login/requestOTTPage.jsx

import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router";
import { useService } from "../../../hooks/useService.js";
import { TYPES } from "../../../di/types.js";

/**
 * RequestOTTPage - Step 1 of the Login Process
 *
 * This component demonstrates several important React and UX concepts:
 * 1. Form state management with validation
 * 2. Error handling and user feedback
 * 3. Navigation state management between pages
 * 4. Integration with dependency injection services
 * 5. Progressive disclosure (showing information as needed)
 *
 * The page serves as the entry point to the secure login process.
 * By requesting email verification first, we can:
 * - Verify the user owns the email account
 * - Prepare the cryptographic challenge before password entry
 * - Provide better error messages if the account doesn't exist
 * - Implement rate limiting at the email level
 */
function RequestOTTPage() {
  // Get our authentication service from dependency injection
  // This gives us access to the three-step login process
  const authService = useService(TYPES.AuthService);
  const logger = useService(TYPES.LoggerService);
  const navigate = useNavigate();
  const location = useLocation();

  // Extract any messages passed from other pages (like registration completion)
  const initialMessage = location.state?.message || null;
  const prefillEmail = location.state?.email || "";

  // Form state management
  // We separate the form data from UI state for cleaner organization
  const [formData, setFormData] = useState({
    email: prefillEmail,
  });

  // UI state tracking - this manages the user's current experience
  const [uiState, setUiState] = useState({
    isSubmitting: false,
    showPassword: false,
  });

  // Error and success message handling
  const [message, setMessage] = useState(initialMessage);
  const [error, setError] = useState(null);

  // Log component initialization for debugging
  useEffect(() => {
    logger.log(
      `RequestOTTPage: Component mounted${prefillEmail ? " with prefilled email" : ""}`,
    );

    // Check if we're already in the middle of a login process
    const loginState = authService.getLoginState();
    if (loginState.step > 0) {
      logger.log(
        `RequestOTTPage: Login already in progress at step ${loginState.step}`,
      );

      // If we're already past this step, redirect appropriately
      if (loginState.step === 1) {
        navigate("/verify-ott", {
          state: {
            email: loginState.email,
            message: "Please enter the verification code from your email",
          },
        });
      } else if (loginState.step === 2) {
        navigate("/complete-login", {
          state: {
            email: loginState.email,
            message: "Please enter your password to complete login",
          },
        });
      }
    }
  }, [authService, logger, navigate, prefillEmail]);

  /**
   * Handle input changes with automatic error clearing
   * This pattern provides immediate feedback as users correct their input
   */
  const handleInputChange = (field, value) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear errors when user starts typing - this provides a responsive feel
    if (error) {
      setError(null);
    }

    // Clear success messages when user modifies the form
    if (message) {
      setMessage(null);
    }
  };

  /**
   * Validate the email before submission
   * This prevents unnecessary network requests and provides immediate feedback
   */
  const validateForm = () => {
    if (!formData.email.trim()) {
      setError("Please enter your email address");
      return false;
    }

    // Basic email format validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(formData.email)) {
      setError("Please enter a valid email address");
      return false;
    }

    return true;
  };

  /**
   * Handle form submission - Step 1 of the login process
   * This initiates the secure authentication flow
   */
  const handleSubmit = async (e) => {
    e.preventDefault();

    logger.log("RequestOTTPage: Form submission started");

    // Clear any previous messages
    setError(null);
    setMessage(null);

    // Validate the form before proceeding
    if (!validateForm()) {
      return;
    }

    // Show loading state to provide user feedback
    setUiState((prev) => ({ ...prev, isSubmitting: true }));

    try {
      // Call the authentication service to request the OTT
      const result = await authService.requestOTT(formData.email);

      if (result.success) {
        logger.log(
          "RequestOTTPage: OTT request successful, navigating to verification",
        );

        // Show success message briefly before navigation
        setMessage(result.message);

        // Navigate to the next step after a brief delay
        // This gives users time to read the success message
        setTimeout(() => {
          navigate("/verify-ott", {
            state: {
              email: result.email,
              message: "Please check your email for a verification code",
            },
          });
        }, 1500);
      } else {
        // Handle API errors gracefully
        setError(result.message);
      }
    } catch (error) {
      // Handle unexpected errors
      logger.log(`RequestOTTPage: Unexpected error: ${error.message}`);
      setError("An unexpected error occurred. Please try again.");
    } finally {
      // Always clear the loading state when done
      setUiState((prev) => ({ ...prev, isSubmitting: false }));
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Sign in to your account
        </h2>
        <p className="mt-2 text-center text-sm text-gray-600">
          Enter your email to begin secure authentication
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          {/* Success Message Display */}
          {message && (
            <div className="mb-4 bg-green-50 border border-green-200 rounded-md p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg
                    className="h-5 w-5 text-green-400"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm font-medium text-green-800">
                    {message}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Error Message Display */}
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

          {/* Login Form */}
          <form className="space-y-6" onSubmit={handleSubmit}>
            {/* Email Input Field */}
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-gray-700"
              >
                Email address
              </label>
              <div className="mt-1">
                <input
                  id="email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  required
                  value={formData.email}
                  onChange={(e) => handleInputChange("email", e.target.value)}
                  className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                  placeholder="Enter your email address"
                  disabled={uiState.isSubmitting}
                />
              </div>
              <p className="mt-2 text-sm text-gray-500">
                We'll send a verification code to this email address
              </p>
            </div>

            {/* Submit Button */}
            <div>
              <button
                type="submit"
                disabled={uiState.isSubmitting || !formData.email.trim()}
                className={`w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white ${
                  uiState.isSubmitting || !formData.email.trim()
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
                    Sending verification code...
                  </>
                ) : (
                  "Continue with Email"
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
                  Why do we need your email?
                </span>
              </div>
            </div>

            <div className="mt-6 text-center">
              <div className="text-sm text-gray-600 space-y-2">
                <p>🔒 Your password never leaves your device</p>
                <p>📧 Email verification proves account ownership</p>
                <p>🔐 End-to-end encryption protects your data</p>
              </div>
            </div>
          </div>

          {/* Alternative Actions */}
          <div className="mt-6">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">
                  Need an account?
                </span>
              </div>
            </div>

            <div className="mt-6">
              <a
                href="/register"
                className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Create a new account
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default RequestOTTPage;
