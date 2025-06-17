// src/pages/anonymous/VerifyEmail/page.jsx

import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router";
import { useService } from "../../../hooks/useService.js";
import { TYPES } from "../../../di/types.js";

/**
 * EmailVerificationPage Component
 *
 * This page handles the email verification step that comes after registration.
 * It demonstrates several key concepts:
 * 1. Using React Router's location state to receive data from previous pages
 * 2. Input handling for verification codes with automatic formatting
 * 3. Integration with backend services for verification
 * 4. Progressive user interface states (loading, success, error)
 * 5. Automatic navigation after successful verification
 */
function EmailVerificationPage() {
  // Get our services from dependency injection
  const registrationService = useService(TYPES.RegistrationService);
  const logger = useService(TYPES.LoggerService);
  const navigate = useNavigate();
  const location = useLocation();

  // Extract email and message from navigation state (passed from registration page)
  const email = location.state?.email || "";
  const initialMessage =
    location.state?.message ||
    "Please check your email for a verification code";

  // State management for the verification form
  const [verificationCode, setVerificationCode] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [message, setMessage] = useState(initialMessage);
  const [error, setError] = useState(null);
  const [isSuccess, setIsSuccess] = useState(false);

  // Log when the component mounts
  useEffect(() => {
    logger.log(`EmailVerificationPage: Loaded for email: ${email}`);

    // If no email was provided, redirect back to registration
    if (!email) {
      logger.log(
        "EmailVerificationPage: No email provided, redirecting to registration",
      );
      navigate("/register", {
        state: { message: "Please complete registration first" },
      });
    }
  }, [email, logger, navigate]);

  /**
   * Handle verification code input with automatic formatting
   * This creates a better user experience by formatting the code as they type
   */
  const handleCodeChange = (value) => {
    // Remove any non-numeric characters and limit to 6 digits
    const numericValue = value.replace(/\D/g, "").slice(0, 6);
    setVerificationCode(numericValue);

    // Clear errors when user starts typing
    if (error) {
      setError(null);
    }
  };

  /**
   * Handle form submission for email verification
   */
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validate the verification code
    if (verificationCode.length !== 6) {
      setError("Please enter a complete 6-digit verification code");
      return;
    }

    setIsSubmitting(true);
    setError(null);

    logger.log(
      `EmailVerificationPage: Submitting verification code: ${verificationCode}`,
    );

    try {
      // Call the registration service to verify the email
      const result = await registrationService.verifyEmail(verificationCode);

      if (result.success) {
        // Verification successful!
        setIsSuccess(true);
        setMessage("Email verified successfully! Redirecting to login...");

        logger.log("EmailVerificationPage: Email verification successful");

        // Navigate to login page after a brief delay
        setTimeout(() => {
          navigate("/login", {
            state: {
              email: email,
              message: "Registration complete! Please sign in to your account.",
            },
          });
        }, 2000);
      } else {
        // Verification failed
        setError(result.message);
        setVerificationCode(""); // Clear the input for retry
      }
    } catch (error) {
      logger.log(`EmailVerificationPage: Unexpected error: ${error.message}`);
      setError("An unexpected error occurred. Please try again.");
      setVerificationCode("");
    } finally {
      setIsSubmitting(false);
    }
  };

  /**
   * Handle request for a new verification code
   * In a real application, this would trigger sending a new email
   */
  const handleResendCode = async () => {
    setMessage("A new verification code has been sent to your email");
    setError(null);
    setVerificationCode("");

    // In a real app, you would call an API to resend the verification email
    logger.log("EmailVerificationPage: Resend verification code requested");
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Verify your email address
        </h2>
        <p className="mt-2 text-center text-sm text-gray-600">
          We sent a verification code to
        </p>
        <p className="text-center text-sm font-medium text-gray-900">{email}</p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          {/* Success Message */}
          {isSuccess && (
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

          {/* Information Message */}
          {!isSuccess && !error && (
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

          {/* Verification Form */}
          {!isSuccess && (
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
                    id="verificationCode"
                    type="text"
                    inputMode="numeric"
                    pattern="[0-9]*"
                    maxLength="6"
                    value={verificationCode}
                    onChange={(e) => handleCodeChange(e.target.value)}
                    placeholder="123456"
                    className="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm text-center text-2xl font-mono tracking-widest"
                    disabled={isSubmitting}
                  />
                </div>
                <p className="mt-2 text-sm text-gray-500">
                  Enter the 6-digit code sent to your email
                </p>
              </div>

              <div>
                <button
                  type="submit"
                  disabled={isSubmitting || verificationCode.length !== 6}
                  className={`w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white ${
                    isSubmitting || verificationCode.length !== 6
                      ? "bg-gray-400 cursor-not-allowed"
                      : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                  }`}
                >
                  {isSubmitting ? "Verifying..." : "Verify Email"}
                </button>
              </div>
            </form>
          )}

          {/* Additional Actions */}
          {!isSuccess && (
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

              <div className="mt-6 space-y-4">
                <button
                  type="button"
                  onClick={handleResendCode}
                  className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                  disabled={isSubmitting}
                >
                  Resend verification code
                </button>

                <button
                  type="button"
                  onClick={() => navigate("/register")}
                  className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Back to registration
                </button>
              </div>
            </div>
          )}

          {/* Help Information */}
          <div className="mt-6 text-center">
            <p className="text-xs text-gray-500">
              Having trouble? Check your spam folder or contact support for
              assistance.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default EmailVerificationPage;
