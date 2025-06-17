// src/pages/anonymous/Login/verifyOTTPage.jsx

import { useState, useEffect, useRef } from "react";
import { useNavigate, useLocation } from "react-router";
import { useService } from "../../../hooks/useService.js";
import { TYPES } from "../../../di/types.js";

/**
 * VerifyOTTPage - Step 2 of the Login Process
 *
 * This component demonstrates several advanced React patterns and UX concepts:
 * 1. State validation and flow control between multi-step processes
 * 2. Input formatting and validation for verification codes
 * 3. Automatic form submission when complete input is detected
 * 4. Error recovery and navigation fallbacks
 * 5. Loading states that communicate what's happening during crypto operations
 *
 * This step serves as the bridge between email verification and password entry.
 * It's where the server prepares the cryptographic challenge that will prove
 * the user knows their password without ever transmitting the password itself.
 *
 * The UX is designed to feel fast and fluid - users can simply type their code
 * and the form submits automatically, reducing friction in the authentication flow.
 */
function VerifyOTTPage() {
  // Get our services from dependency injection
  const authService = useService(TYPES.AuthService);
  const logger = useService(TYPES.LoggerService);
  const navigate = useNavigate();
  const location = useLocation();

  // Extract navigation state from the previous page
  const email = location.state?.email || "";
  const initialMessage =
    location.state?.message ||
    "Please enter the verification code from your email";

  // Form state for the verification code
  const [verificationCode, setVerificationCode] = useState("");

  // UI state management
  const [uiState, setUiState] = useState({
    isSubmitting: false,
    isValidating: false,
  });

  // Message and error handling
  const [message, setMessage] = useState(initialMessage);
  const [error, setError] = useState(null);

  // Ref for the input field to manage focus
  const codeInputRef = useRef(null);

  // Component initialization and state validation
  useEffect(() => {
    logger.log(`VerifyOTTPage: Component mounted for email: ${email}`);

    // Validate that we're in the correct step of the login process
    const loginState = authService.getLoginState();

    if (!email) {
      // If no email provided, redirect back to start
      logger.log(
        "VerifyOTTPage: No email provided, redirecting to login start",
      );
      navigate("/login", {
        state: {
          message: "Please start the login process by entering your email",
        },
      });
      return;
    }

    if (loginState.step === 0) {
      // If no login process started, redirect to start with the email
      logger.log(
        "VerifyOTTPage: No login process in progress, redirecting to start",
      );
      navigate("/login", {
        state: {
          email: email,
          message: "Please request a new verification code",
        },
      });
      return;
    }

    if (loginState.step === 2) {
      // If already verified, skip to password entry
      logger.log("VerifyOTTPage: Already verified, proceeding to password");
      navigate("/complete-login", {
        state: {
          email: email,
          message: "Please enter your password to complete login",
        },
      });
      return;
    }

    // Focus the input field for better UX
    if (codeInputRef.current) {
      codeInputRef.current.focus();
    }
  }, [email, authService, logger, navigate]);

  /**
   * Handle verification code input with automatic formatting and submission
   * This creates a smooth user experience where the form submits automatically
   * when a complete code is entered, reducing the need for manual submission
   */
  const handleCodeChange = (value) => {
    // Remove any non-numeric characters and limit to 6 digits
    const numericValue = value.replace(/\D/g, "").slice(0, 6);
    setVerificationCode(numericValue);

    // Clear any existing errors when user starts typing
    if (error) {
      setError(null);
    }

    // If we have a complete 6-digit code, automatically submit
    // This reduces friction and makes the experience feel more responsive
    if (numericValue.length === 6 && !uiState.isSubmitting) {
      logger.log("VerifyOTTPage: Complete code entered, auto-submitting");
      setTimeout(() => handleSubmit(null, numericValue), 100);
    }
  };

  /**
   * Handle form submission for OTT verification
   * This can be called either by manual form submission or automatically
   * when a complete code is entered
   */
  const handleSubmit = async (e, autoCode = null) => {
    // Prevent default form submission if called from form event
    if (e) {
      e.preventDefault();
    }

    const codeToVerify = autoCode || verificationCode;

    logger.log(`VerifyOTTPage: Submitting verification code: ${codeToVerify}`);

    // Validate the code before submission
    if (codeToVerify.length !== 6) {
      setError("Please enter a complete 6-digit verification code");
      return;
    }

    // Clear any previous messages and set loading state
    setError(null);
    setMessage(null);
    setUiState((prev) => ({ ...prev, isSubmitting: true }));

    try {
      // Call the authentication service to verify the OTT
      // This step retrieves the encrypted user data and prepares the challenge
      const result = await authService.verifyOTT(email, codeToVerify);

      if (result.success) {
        logger.log(
          "VerifyOTTPage: OTT verification successful, proceeding to password",
        );

        // Show success message briefly
        setMessage("Email verified! Preparing secure login...");

        // Navigate to password entry after a brief delay
        setTimeout(() => {
          navigate("/complete-login", {
            state: {
              email: email,
              message: "Please enter your password to complete login",
              challengeId: result.challengeId,
            },
          });
        }, 1000);
      } else {
        // Handle verification failure
        setError(result.message);
        setVerificationCode(""); // Clear the code for retry

        // Refocus the input field
        if (codeInputRef.current) {
          codeInputRef.current.focus();
        }
      }
    } catch (error) {
      // Handle unexpected errors
      logger.log(`VerifyOTTPage: Unexpected error: ${error.message}`);
      setError("An unexpected error occurred. Please try again.");
      setVerificationCode("");

      if (codeInputRef.current) {
        codeInputRef.current.focus();
      }
    } finally {
      // Always clear the loading state
      setUiState((prev) => ({ ...prev, isSubmitting: false }));
    }
  };

  /**
   * Handle request for a new verification code
   * This provides a recovery path if the user didn't receive the email
   */
  const handleResendCode = () => {
    logger.log("VerifyOTTPage: Resend code requested");

    // Navigate back to the first step with the current email
    navigate("/login", {
      state: {
        email: email,
        message: "Please request a new verification code",
      },
    });
  };

  /**
   * Format the verification code display with spaces for better readability
   * This makes it easier for users to verify they've entered the code correctly
   */
  const formatCodeDisplay = (code) => {
    // Add a space every 3 digits: "123 456"
    return code.replace(/(\d{3})(\d{1,3})/, "$1 $2");
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Verify your email
        </h2>
        <p className="mt-2 text-center text-sm text-gray-600">
          We sent a verification code to
        </p>
        <p className="text-center text-sm font-medium text-gray-900">{email}</p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          {/* Success Message */}
          {message && !error && (
            <div className="mb-4 bg-blue-50 border border-blue-200 rounded-md p-4">
              <div className="flex">
                <div className="flex-shrink-0">
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
                </div>
                <div className="ml-3">
                  <p className="text-sm font-medium text-blue-800">{message}</p>
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

          {/* Verification Code Form */}
          <form className="space-y-6" onSubmit={handleSubmit}>
            <div>
              <label
                htmlFor="verificationCode"
                className="block text-sm font-medium text-gray-700"
              >
                Verification Code
              </label>
              <div className="mt-1">
                <input
                  ref={codeInputRef}
                  id="verificationCode"
                  type="text"
                  inputMode="numeric"
                  pattern="[0-9]*"
                  maxLength="6"
                  value={formatCodeDisplay(verificationCode)}
                  onChange={(e) => handleCodeChange(e.target.value)}
                  placeholder="000 000"
                  className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm text-center text-2xl font-mono tracking-widest"
                  disabled={uiState.isSubmitting}
                  autoComplete="one-time-code"
                />
              </div>
              <p className="mt-2 text-sm text-gray-500">
                Enter the 6-digit code sent to your email
              </p>
            </div>

            {/* Manual Submit Button (hidden when auto-submit is available) */}
            {verificationCode.length < 6 && (
              <div>
                <button
                  type="submit"
                  disabled={
                    uiState.isSubmitting || verificationCode.length !== 6
                  }
                  className={`w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white ${
                    uiState.isSubmitting || verificationCode.length !== 6
                      ? "bg-gray-400 cursor-not-allowed"
                      : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                  }`}
                >
                  {uiState.isSubmitting ? "Verifying..." : "Verify Code"}
                </button>
              </div>
            )}

            {/* Auto-submit status */}
            {verificationCode.length === 6 && uiState.isSubmitting && (
              <div className="text-center">
                <div className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-indigo-700 bg-indigo-100">
                  <svg
                    className="animate-spin -ml-1 mr-2 h-4 w-4 text-indigo-700"
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
                  Verifying code...
                </div>
              </div>
            )}
          </form>

          {/* Recovery Options */}
          <div className="mt-6">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">
                  Didn't receive the code?
                </span>
              </div>
            </div>

            <div className="mt-6 space-y-3">
              <button
                type="button"
                onClick={handleResendCode}
                className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                disabled={uiState.isSubmitting}
              >
                Request a new code
              </button>

              <button
                type="button"
                onClick={() => navigate("/login")}
                className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Use a different email
              </button>
            </div>
          </div>

          {/* Help Information */}
          <div className="mt-6 text-center">
            <p className="text-xs text-gray-500">
              Check your spam folder or contact support if you continue to have
              trouble
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default VerifyOTTPage;
