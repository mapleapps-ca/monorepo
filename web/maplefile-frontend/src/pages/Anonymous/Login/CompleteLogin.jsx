// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/CompleteLogin.jsx
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
  KeyIcon,
  EyeIcon,
  EyeSlashIcon,
  UserIcon,
  ClockIcon,
} from "@heroicons/react/24/outline";

const CompleteLogin = () => {
  const navigate = useNavigate();

  // Get services from the unified service system (like developer version)
  const services = useServices();
  const authManager = services?.authManager;
  const localStorageService = services?.localStorageService;
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [challengeExpired, setChallengeExpired] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [verificationData, setVerificationData] = useState(null);

  useEffect(() => {
    // Get email and verification data from previous steps
    let storedEmail = null;
    let storedVerificationData = null;

    // Try to get email from multiple sources
    if (authManager && typeof authManager.getCurrentUserEmail === "function") {
      try {
        storedEmail = authManager.getCurrentUserEmail();
        if (storedEmail) {
          console.log(
            "[CompleteLogin] Using email from authManager:",
            storedEmail,
            typeof storedEmail,
          );
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] Could not get email from authManager:",
          err,
        );
      }
    }

    if (!storedEmail && localStorageService) {
      try {
        storedEmail = localStorageService.getUserEmail();
        if (storedEmail) {
          console.log(
            "[CompleteLogin] Using email from localStorageService:",
            storedEmail,
            typeof storedEmail,
          );
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] Could not get email from localStorageService:",
          err,
        );
      }
    }

    if (!storedEmail) {
      storedEmail = sessionStorage.getItem("loginEmail");
      if (storedEmail) {
        console.log(
          "[CompleteLogin] Using email from sessionStorage:",
          storedEmail,
          typeof storedEmail,
        );
      }
    }

    // Try to get verification data
    try {
      const sessionData = sessionStorage.getItem("otpVerificationResult");
      if (sessionData) {
        storedVerificationData = JSON.parse(sessionData);
        console.log(
          "[CompleteLogin] Using verification data from sessionStorage",
        );
      }
    } catch (err) {
      console.warn(
        "[CompleteLogin] Could not parse verification data from sessionStorage:",
        err,
      );
    }

    if (!storedVerificationData && localStorageService) {
      try {
        storedVerificationData =
          localStorageService.getLoginSessionData("verify_response");
        if (storedVerificationData) {
          console.log(
            "[CompleteLogin] Using verification data from localStorageService",
          );
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] Could not get verification data from localStorageService:",
          err,
        );
      }
    }

    if (storedEmail && storedVerificationData) {
      // Ensure email is a string and set both values together
      setEmail(String(storedEmail));
      setVerificationData(storedVerificationData);
    } else {
      console.error(
        "[CompleteLogin] Missing email or verify data, redirecting to start",
        { email: storedEmail, verifyData: storedVerificationData },
      );
      navigate("/login/request-ott");
    }
  }, [navigate, authManager, localStorageService]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    setChallengeExpired(false);

    try {
      // Validate services
      if (!authManager) {
        throw new Error(
          "Authentication service not available. Please refresh the page.",
        );
      }

      // Validate inputs
      if (!password) {
        throw new Error("Master password is required");
      }

      // Ensure email is a string and validate
      if (!email || typeof email !== "string" || !email.trim()) {
        console.error("[CompleteLogin] Invalid email data:", {
          email,
          type: typeof email,
        });
        throw new Error("Valid email address is required");
      }

      if (!verificationData || !verificationData.challengeId) {
        throw new Error(
          "Missing challenge data. Please start the login process again.",
        );
      }

      console.log(
        "[CompleteLogin] Completing login via AuthManager for:",
        email,
      );

      // Use the working method signature from developer version
      console.log("[CompleteLogin] Decrypting challenge data...");

      let decryptedChallenge;
      try {
        decryptedChallenge = await authManager.decryptChallenge(
          password,
          verificationData,
        );
      } catch (decryptError) {
        console.error(
          "[CompleteLogin] Challenge decryption failed:",
          decryptError,
        );
        throw new Error(
          "Failed to decrypt challenge data. Please check your password.",
        );
      }

      console.log("[CompleteLogin] Calling authManager.completeLogin...");
      let response;
      try {
        response = await authManager.completeLogin(
          email, // Direct email string (no trimming/lowercasing here)
          verificationData.challengeId,
          decryptedChallenge,
        );
      } catch (loginError) {
        console.error(
          "[CompleteLogin] AuthManager.completeLogin failed:",
          loginError,
        );
        throw new Error("Failed to complete login: " + loginError.message);
      }

      // Validate response
      if (!response) {
        throw new Error("Login completion returned no response");
      }

      console.log("[CompleteLogin] Login completed successfully!", response);

      // Store password in password storage service (like developer version)
      if (services?.passwordStorageService) {
        services.passwordStorageService.setPassword(password);
        console.log(
          "[CompleteLogin] Password stored in PasswordStorageService",
        );
      } else {
        console.warn(
          "[CompleteLogin] PasswordStorageService not available, password not stored",
        );
      }

      // Clear stored session data
      try {
        sessionStorage.removeItem("loginEmail");
        sessionStorage.removeItem("otpVerificationResult");

        if (
          localStorageService &&
          typeof localStorageService.clearLoginSessionData === "function"
        ) {
          localStorageService.clearLoginSessionData();
        }
      } catch (cleanupError) {
        console.warn(
          "[CompleteLogin] Could not clear session data:",
          cleanupError,
        );
      }

      console.log(
        "[CompleteLogin] Login completed successfully, redirecting to dashboard",
      );

      // Navigate to dashboard
      navigate("/dashboard");
    } catch (error) {
      console.error("[CompleteLogin] Login completion failed:", error);

      // Handle specific error types with user-friendly messages
      const errorMessage = error.message || error.toString();

      // Check for expired challenge - this is a common, recoverable error
      if (
        errorMessage.includes("Invalid or expired challenge") ||
        errorMessage.includes("challenge") ||
        errorMessage.includes("expired")
      ) {
        console.log(
          "[CompleteLogin] Challenge expired - directing user to get new code",
        );
        setChallengeExpired(true);
        setError("");
      }
      // Check for wrong password
      else if (
        errorMessage.includes("password") ||
        errorMessage.includes("decrypt") ||
        errorMessage.includes("authentication failed")
      ) {
        setError(
          "Incorrect password. Please check your master password and try again.",
        );
        setChallengeExpired(false);
      }
      // Check for network/server errors
      else if (
        errorMessage.includes("fetch") ||
        errorMessage.includes("network") ||
        errorMessage.includes("500") ||
        errorMessage.includes("Request failed")
      ) {
        setError(
          "Connection error. Please check your internet connection and try again.",
        );
        setChallengeExpired(false);
      }
      // Generic error with helpful context
      else {
        setError(
          `Login failed: ${errorMessage}. Please try again or contact support if the issue persists.`,
        );
        setChallengeExpired(false);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleBackToVerify = () => {
    // Clear all error states before navigating
    setError("");
    setChallengeExpired(false);
    navigate("/login/verify-ott");
  };

  const handleStartOver = () => {
    // Clear all session data and error states
    setError("");
    setChallengeExpired(false);

    try {
      sessionStorage.removeItem("loginEmail");
      sessionStorage.removeItem("otpVerificationResult");

      if (
        localStorageService &&
        typeof localStorageService.clearLoginSessionData === "function"
      ) {
        localStorageService.clearLoginSessionData();
      }
    } catch (cleanupError) {
      console.warn(
        "[CompleteLogin] Could not clear session data:",
        cleanupError,
      );
    }

    navigate("/login/request-ott");
  };

  // Early return if verification data is not available (like developer version)
  if (!verificationData) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Loading...</h2>
          <p className="text-gray-600">Loading verification data...</p>
        </div>
      </div>
    );
  }

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
                <div className="flex items-center justify-center w-8 h-8 bg-green-500 rounded-full text-white text-sm font-bold">
                  <CheckIcon className="h-4 w-4" />
                </div>
                <span className="ml-2 text-sm font-semibold text-green-600">
                  Verify
                </span>
              </div>
              <div className="w-12 h-0.5 bg-green-500"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gradient-to-r from-red-800 to-red-900 rounded-full text-white text-sm font-bold">
                  3
                </div>
                <span className="ml-2 text-sm font-semibold text-gray-900">
                  Access
                </span>
              </div>
            </div>
          </div>

          {/* Header */}
          <div className="text-center animate-fade-in-up">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="flex items-center justify-center h-16 w-16 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-lg animate-pulse">
                  <UserIcon className="h-8 w-8 text-white" />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
              </div>
            </div>
            <h2 className="text-3xl font-black text-gray-900 mb-2">
              Complete Secure Login
            </h2>
            <p className="text-gray-600 mb-2">
              Enter your master password to decrypt your files and complete
              login
            </p>
            <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
              <ShieldCheckIcon className="h-4 w-4 text-green-600" />
              <span>All encryption happens locally in your browser</span>
            </div>
          </div>

          {/* Form Card */}
          <div className="bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 animate-fade-in-up-delay">
            {/* Challenge Expired Message */}
            {challengeExpired && (
              <div className="mb-6 p-6 rounded-lg bg-amber-50 border border-amber-200 animate-fade-in">
                <div className="flex items-start">
                  <ClockIcon className="h-6 w-6 text-amber-600 mr-3 flex-shrink-0 mt-1" />
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-amber-800 mb-2">
                      Security Code Expired
                    </h3>
                    <p className="text-sm text-amber-700 mb-4">
                      Your verification code has expired for security reasons.
                      This is normal and helps protect your account.
                    </p>
                    <div className="flex flex-col sm:flex-row gap-3">
                      <button
                        onClick={() => navigate("/login/verify-ott")}
                        className="flex-1 inline-flex items-center justify-center px-4 py-2 bg-amber-600 text-white text-sm font-semibold rounded-lg hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 transition-all duration-200"
                      >
                        <ArrowLeftIcon className="mr-2 h-4 w-4" />
                        Get New Code
                      </button>
                      <button
                        onClick={handleStartOver}
                        className="flex-1 inline-flex items-center justify-center px-4 py-2 border border-amber-300 text-amber-700 text-sm font-medium rounded-lg hover:bg-amber-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 transition-all duration-200"
                      >
                        Start Over
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Error Message */}
            {error && (
              <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 animate-fade-in">
                <div className="flex items-center">
                  <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-red-800">
                      Authentication Error
                    </h3>
                    <p className="text-sm text-red-700 mt-1">{error}</p>
                  </div>
                </div>
              </div>
            )}

            {/* User Display */}
            <div className="mb-6 p-4 bg-gray-50 rounded-lg border">
              <div className="flex items-center">
                <UserIcon className="h-5 w-5 text-gray-500 mr-3" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-gray-700">
                    Signing in as:
                  </p>
                  <p className="text-sm text-gray-900 font-mono">{email}</p>
                </div>
                <button
                  onClick={handleStartOver}
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
                  htmlFor="password"
                  className="block text-sm font-semibold text-gray-700 mb-2"
                >
                  Master Password
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? "text" : "password"}
                    id="password"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      // Clear error states when user starts typing
                      if (error) setError("");
                      if (challengeExpired) setChallengeExpired(false);
                    }}
                    placeholder="Enter your master password"
                    required
                    disabled={loading || challengeExpired}
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-500 pr-12"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    disabled={loading}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  >
                    {showPassword ? (
                      <EyeSlashIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                    ) : (
                      <EyeIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                    )}
                  </button>
                </div>
                <p className="mt-2 text-xs text-gray-500">
                  This password decrypts your encryption keys locally
                </p>
              </div>

              <div className="space-y-3">
                <button
                  type="submit"
                  disabled={
                    loading ||
                    !password.trim() ||
                    !authManager ||
                    !email ||
                    typeof email !== "string" ||
                    !verificationData ||
                    challengeExpired
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
                      Decrypting & Signing In...
                    </>
                  ) : (
                    <>
                      <KeyIcon className="mr-2 h-4 w-4" />
                      Decrypt & Access Files
                      <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
                    </>
                  )}
                </button>

                <button
                  type="button"
                  onClick={handleBackToVerify}
                  disabled={loading}
                  className="w-full flex justify-center items-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 disabled:bg-gray-50 disabled:cursor-not-allowed transition-all duration-200"
                >
                  <ArrowLeftIcon className="mr-2 h-4 w-4" />
                  Back to Verification
                </button>
              </div>
            </form>

            {/* Security Info Section */}
            <div className="mt-8 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
              <h3 className="text-sm font-semibold text-blue-900 mb-3 flex items-center">
                <InformationCircleIcon className="h-4 w-4 mr-2" />
                Security Information
              </h3>
              <ul className="text-sm text-blue-800 space-y-2">
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  Your master password never leaves your device
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  All decryption happens locally in your browser
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  We cannot recover your password if you forget it
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  This ensures maximum privacy and security
                </li>
              </ul>
            </div>

            {/* Forgot Password Link */}
            <div className="mt-6 text-center">
              <Link
                to="/recovery"
                className="text-sm text-red-600 hover:text-red-700 font-medium hover:underline transition-colors duration-200"
              >
                Forgot your master password? Use account recovery
              </Link>
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

      {/* CSS Animations */}
      <style>{`
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

export default CompleteLogin;
