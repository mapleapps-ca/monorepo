// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/VerifyOTT.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router";
import { useServices } from "../../../services/Services";
import {
  ArrowRightIcon,
  ArrowLeftIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  ServerIcon,
  EyeSlashIcon,
  ClockIcon,
  EnvelopeIcon,
  KeyIcon,
} from "@heroicons/react/24/outline";

const VerifyOTT = () => {
  const navigate = useNavigate();
  const { authManager, localStorageService } = useServices();
  const [ott, setOtt] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [resendLoading, setResendLoading] = useState(false);
  const [resendSuccess, setResendSuccess] = useState(false);
  const [timeRemaining, setTimeRemaining] = useState(600); // 10 minutes in seconds

  useEffect(() => {
    // FIXED: Try multiple sources for email
    let storedEmail = null;

    // Try 1: Get from authManager
    if (authManager && typeof authManager.getCurrentUserEmail === "function") {
      try {
        storedEmail = authManager.getCurrentUserEmail();
        if (storedEmail) {
          console.log("[VerifyOTT] Using email from authManager:", storedEmail);
        }
      } catch (err) {
        console.warn("[VerifyOTT] Could not get email from authManager:", err);
      }
    }

    // Try 2: Get from localStorageService
    if (!storedEmail && localStorageService) {
      try {
        storedEmail = localStorageService.getUserEmail();
        if (storedEmail) {
          console.log(
            "[VerifyOTT] Using email from localStorageService:",
            storedEmail,
          );
        }
      } catch (err) {
        console.warn(
          "[VerifyOTT] Could not get email from localStorageService:",
          err,
        );
      }
    }

    // Try 3: Get from sessionStorage (fallback)
    if (!storedEmail) {
      storedEmail = sessionStorage.getItem("loginEmail");
      if (storedEmail) {
        console.log(
          "[VerifyOTT] Using email from sessionStorage:",
          storedEmail,
        );
      }
    }

    if (storedEmail) {
      setEmail(storedEmail);
    } else {
      // If no email found anywhere, redirect to first step
      console.log(
        "[VerifyOTT] No stored email found in any location, redirecting to start",
      );
      navigate("/login/request-ott");
    }
  }, [navigate, authManager, localStorageService]);

  // Countdown timer effect
  useEffect(() => {
    if (timeRemaining <= 0) return;

    const timer = setInterval(() => {
      setTimeRemaining((prev) => {
        if (prev <= 1) {
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [timeRemaining]);

  // Format time as MM:SS
  const formatTime = (seconds) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      // Validate services are available
      if (!authManager) {
        throw new Error(
          "Authentication service not available. Please refresh the page.",
        );
      }

      // Validate OTT
      if (!ott) {
        throw new Error("Verification code is required");
      }

      if (ott.length !== 6 || !/^\d{6}$/.test(ott)) {
        throw new Error("Verification code must be 6 digits");
      }

      if (!email || !email.trim()) {
        throw new Error("Email address is required");
      }

      if (!email.includes("@")) {
        throw new Error("Please enter a valid email address");
      }

      console.log(
        "[VerifyOTT] Verifying OTT via AuthManager for:",
        email.trim(),
      );
      console.log(
        "[VerifyOTT] Available authManager methods:",
        Object.getOwnPropertyNames(Object.getPrototypeOf(authManager)),
      );

      const trimmedEmail = email.trim().toLowerCase();

      // FIXED: Try different possible method names and validate responses
      let response;

      if (typeof authManager.verifyOTP === "function") {
        console.log("[VerifyOTT] Using verifyOTP method");
        response = await authManager.verifyOTP({
          email: trimmedEmail,
          otp_code: ott,
        });
      } else if (typeof authManager.verifyOTT === "function") {
        console.log("[VerifyOTT] Using verifyOTT method");
        response = await authManager.verifyOTT(trimmedEmail, ott);
      } else if (typeof authManager.verify === "function") {
        console.log("[VerifyOTT] Using verify method");
        response = await authManager.verify({
          email: trimmedEmail,
          code: ott,
        });
      } else {
        // List available methods for debugging
        const methods = Object.getOwnPropertyNames(
          Object.getPrototypeOf(authManager),
        ).filter(
          (name) =>
            typeof authManager[name] === "function" && !name.startsWith("_"),
        );

        console.error("[VerifyOTT] Available authManager methods:", methods);
        throw new Error(
          `OTP verification method not found. Available methods: ${methods.join(", ")}`,
        );
      }

      // Validate that we got a proper response
      if (!response) {
        throw new Error("Verification method returned no response");
      }

      console.log("[VerifyOTT] Raw verification response:", response);

      // Check if response indicates success
      if (response.error) {
        throw new Error(response.error);
      }

      // If response is just a success message, we might need additional data
      if (
        typeof response === "string" ||
        (response.success && !response.challengeId)
      ) {
        console.warn(
          "[VerifyOTT] Response may not contain challenge data:",
          response,
        );
        // Try to get additional session data if available
        if (
          localStorageService &&
          typeof localStorageService.getLoginSessionData === "function"
        ) {
          try {
            const additionalData =
              localStorageService.getLoginSessionData("challenge_data");
            if (additionalData) {
              response = { ...response, ...additionalData };
              console.log("[VerifyOTT] Merged with additional session data");
            }
          } catch (err) {
            console.warn(
              "[VerifyOTT] Could not get additional session data:",
              err,
            );
          }
        }
      }

      console.log(
        "[VerifyOTT] Verification successful via AuthManager, received encrypted keys",
      );
      console.log("[VerifyOTT] Verification response:", response);

      // FIXED: Store the verification response data for CompleteLogin to use
      try {
        // Store in sessionStorage as primary method
        sessionStorage.setItem(
          "otpVerificationResult",
          JSON.stringify(response),
        );
        console.log(
          "[VerifyOTT] Verification response stored in sessionStorage",
        );

        // Also try to store via localStorageService if available
        if (
          localStorageService &&
          typeof localStorageService.setLoginSessionData === "function"
        ) {
          try {
            localStorageService.setLoginSessionData(
              "verify_response",
              response,
            );
            console.log(
              "[VerifyOTT] Verification response stored via localStorageService",
            );
          } catch (storageError) {
            console.warn(
              "[VerifyOTT] Could not store via localStorageService:",
              storageError,
            );
          }
        }

        // Store email as well to ensure it's available
        sessionStorage.setItem("loginEmail", trimmedEmail);

        // Also try to store email via authManager if available
        if (
          authManager &&
          typeof authManager.setCurrentUserEmail === "function"
        ) {
          try {
            authManager.setCurrentUserEmail(trimmedEmail);
            console.log("[VerifyOTT] Email stored via authManager");
          } catch (emailError) {
            console.warn(
              "[VerifyOTT] Could not store email via authManager:",
              emailError,
            );
          }
        }
      } catch (storageError) {
        console.error(
          "[VerifyOTT] Failed to store verification data:",
          storageError,
        );
        throw new Error(
          "Verification successful but failed to store data. Please try again.",
        );
      }

      // Navigate to complete login step
      navigate("/login/complete");
    } catch (error) {
      console.error("[VerifyOTT] Verification failed via AuthManager:", error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleResendCode = async () => {
    if (!authManager) {
      setError(
        "Authentication service not available. Please refresh the page.",
      );
      return;
    }

    if (!email || !email.trim()) {
      setError("Email address is required to resend code");
      return;
    }

    setResendLoading(true);
    setError("");
    setResendSuccess(false);

    try {
      console.log("[VerifyOTT] Resending OTT code...");

      const trimmedEmail = email.trim().toLowerCase();

      // Try different possible method names for requesting OTT
      if (typeof authManager.requestOTT === "function") {
        await authManager.requestOTT(trimmedEmail);
      } else if (typeof authManager.requestOTP === "function") {
        await authManager.requestOTP({ email: trimmedEmail });
      } else {
        throw new Error("Request OTT method not found on authManager");
      }

      console.log("[VerifyOTT] Code resent successfully");
      setResendSuccess(true);
      setOtt(""); // Clear the current code
      setTimeRemaining(600); // Reset timer to 10 minutes

      // Clear success message after 3 seconds
      setTimeout(() => {
        setResendSuccess(false);
      }, 3000);
    } catch (err) {
      console.error("[VerifyOTT] Failed to resend code:", err);
      setError(err.message || "Failed to resend code. Please try again.");
    } finally {
      setResendLoading(false);
    }
  };

  const handleBackToEmail = () => {
    navigate("/login/request-ott");
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50 flex flex-col">
      {/* Navigation */}
      <nav className="bg-white/95 backdrop-blur-sm border-b border-gray-100">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <Link to="/" className="flex items-center group">
              <div className="flex items-center justify-center h-10 w-10 bg-gradient-to-br from-red-800 to-red-900 rounded-lg mr-3 group-hover:scale-105 transition-transform duration-200">
                <LockClosedIcon className="h-6 w-6 text-white" />
              </div>
              <span className="text-2xl font-bold bg-gradient-to-r from-gray-900 to-red-800 bg-clip-text text-transparent">
                MapleFile
              </span>
            </Link>
            <div className="flex items-center space-x-6">
              <Link
                to="/register"
                className="text-base font-medium text-gray-700 hover:text-red-800 transition-colors duration-200"
              >
                Need an account?
              </Link>
              <Link
                to="/recovery"
                className="text-base font-medium text-gray-700 hover:text-red-800 transition-colors duration-200"
              >
                Forgot password?
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <div className="flex-1 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          {/* Progress Indicator */}
          <div className="flex items-center justify-center mb-8">
            <div className="flex items-center space-x-4">
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-green-500 rounded-full text-white text-sm font-bold">
                  <CheckIcon className="h-4 w-4" />
                </div>
                <span className="ml-2 text-sm font-semibold text-green-600">
                  Email
                </span>
              </div>
              <div className="w-12 h-0.5 bg-green-500"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gradient-to-r from-red-800 to-red-900 rounded-full text-white text-sm font-bold">
                  2
                </div>
                <span className="ml-2 text-sm font-semibold text-gray-900">
                  Verify
                </span>
              </div>
              <div className="w-12 h-0.5 bg-gray-300"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gray-300 rounded-full text-gray-500 text-sm font-bold">
                  3
                </div>
                <span className="ml-2 text-sm text-gray-500">Access</span>
              </div>
            </div>
          </div>

          {/* Header */}
          <div className="text-center animate-fade-in-up">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="flex items-center justify-center h-16 w-16 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-lg animate-pulse">
                  <KeyIcon className="h-8 w-8 text-white" />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
              </div>
            </div>
            <h2 className="text-3xl font-black text-gray-900 mb-2">
              Verify Your Code
            </h2>
            <p className="text-gray-600 mb-2">
              Enter the 6-digit verification code sent to your email
            </p>
            <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
              <ClockIcon className="h-4 w-4 text-amber-600" />
              <span
                className={
                  timeRemaining <= 60 ? "text-red-600 font-semibold" : ""
                }
              >
                {timeRemaining > 0
                  ? `Code expires in ${formatTime(timeRemaining)}`
                  : "Code has expired"}
              </span>
            </div>
          </div>

          {/* Form Card */}
          <div className="bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 animate-fade-in-up-delay">
            {/* Error Message */}
            {error && (
              <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 animate-fade-in">
                <div className="flex items-center">
                  <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-red-800">
                      Verification Error
                    </h3>
                    <p className="text-sm text-red-700 mt-1">{error}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Resend Success Message */}
            {resendSuccess && (
              <div className="mb-6 p-4 rounded-lg bg-green-50 border border-green-200 animate-fade-in">
                <div className="flex items-center">
                  <CheckIcon className="h-5 w-5 text-green-500 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-green-800">
                      Code Resent Successfully
                    </h3>
                    <p className="text-sm text-green-700 mt-1">
                      A new verification code has been sent to your email.
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Email Display */}
            <div className="mb-6 p-4 bg-gray-50 rounded-lg border">
              <div className="flex items-center">
                <EnvelopeIcon className="h-5 w-5 text-gray-500 mr-3" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-gray-700">
                    Sending to:
                  </p>
                  <p className="text-sm text-gray-900 font-mono">{email}</p>
                </div>
                <button
                  onClick={handleBackToEmail}
                  disabled={loading}
                  className="text-sm text-red-600 hover:text-red-700 font-medium hover:underline transition-colors duration-200"
                >
                  Change
                </button>
              </div>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              <div>
                <label
                  htmlFor="ott"
                  className="block text-sm font-semibold text-gray-700 mb-2"
                >
                  Verification Code
                </label>
                <div className="relative">
                  <input
                    type="text"
                    id="ott"
                    value={ott}
                    onChange={(e) =>
                      setOtt(e.target.value.replace(/\D/g, "").slice(0, 6))
                    }
                    placeholder="000000"
                    maxLength={6}
                    required
                    disabled={loading}
                    className={`w-full px-4 py-4 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-400 text-center text-2xl font-mono tracking-widest ${
                      ott.length === 6
                        ? "border-green-300 bg-green-50"
                        : "border-gray-300"
                    }`}
                  />
                </div>
                <p className="mt-2 text-xs text-gray-500 text-center">
                  Enter the 6-digit code from your email
                </p>
              </div>

              <div className="space-y-3">
                <button
                  type="submit"
                  disabled={
                    loading ||
                    ott.length !== 6 ||
                    !email.trim() ||
                    !authManager ||
                    timeRemaining <= 0
                  }
                  className="group w-full flex justify-center items-center py-3 px-4 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
                >
                  {loading ? (
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
                      Verifying Code...
                    </>
                  ) : timeRemaining <= 0 ? (
                    "Code Expired - Resend Required"
                  ) : (
                    <>
                      Verify & Continue
                      <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
                    </>
                  )}
                </button>

                <div className="flex space-x-3">
                  <button
                    type="button"
                    onClick={handleResendCode}
                    disabled={resendLoading || !authManager}
                    className="flex-1 flex justify-center items-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:bg-gray-50 disabled:cursor-not-allowed transition-all duration-200"
                  >
                    {resendLoading ? (
                      <>
                        <svg
                          className="animate-spin -ml-1 mr-2 h-4 w-4 text-gray-500"
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
                        Sending...
                      </>
                    ) : (
                      "Resend Code"
                    )}
                  </button>

                  <button
                    type="button"
                    onClick={handleBackToEmail}
                    disabled={loading}
                    className="flex-1 flex justify-center items-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 disabled:bg-gray-50 disabled:cursor-not-allowed transition-all duration-200"
                  >
                    <ArrowLeftIcon className="mr-2 h-4 w-4" />
                    Change Email
                  </button>
                </div>
              </div>
            </form>

            {/* Help Section */}
            <div className="mt-6 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
              <h3 className="text-sm font-semibold text-blue-900 mb-3 flex items-center">
                <InformationCircleIcon className="h-4 w-4 mr-2" />
                Didn't receive the code?
              </h3>
              <ul className="text-sm text-blue-800 space-y-2">
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  Check your spam/junk folder
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  Make sure the email address is correct
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  The code expires in 10 minutes
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  Use "Resend Code" to get a new verification code
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-100 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center text-sm text-gray-500">
            <p>&copy; 2025 MapleFile Inc. All rights reserved.</p>
            <div className="mt-2 space-x-4">
              <Link
                to="#"
                className="hover:text-gray-700 transition-colors duration-200"
              >
                Privacy Policy
              </Link>
              <Link
                to="#"
                className="hover:text-gray-700 transition-colors duration-200"
              >
                Terms of Service
              </Link>
              <Link
                to="#"
                className="hover:text-gray-700 transition-colors duration-200"
              >
                Support
              </Link>
            </div>
          </div>
        </div>
      </footer>

      <style jsx>{`
        @keyframes fade-in {
          from {
            opacity: 0;
            transform: translateY(10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        @keyframes fade-in-up {
          from {
            opacity: 0;
            transform: translateY(30px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        .animate-fade-in {
          animation: fade-in 0.4s ease-out;
        }

        .animate-fade-in-up {
          animation: fade-in-up 0.6s ease-out;
        }

        .animate-fade-in-up-delay {
          animation: fade-in-up 0.6s ease-out 0.2s both;
        }
      `}</style>
    </div>
  );
};

export default VerifyOTT;
