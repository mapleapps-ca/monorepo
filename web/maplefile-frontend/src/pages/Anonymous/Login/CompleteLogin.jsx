// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/CompleteLogin.jsx
import { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router";
import { useServices } from "../../../services/Services";
import {
  ArrowRightIcon,
  ArrowLeftIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  KeyIcon,
  EyeIcon,
  EyeSlashIcon,
  UserIcon,
  ClockIcon,
  ServerIcon,
  GlobeAltIcon,
  HeartIcon,
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

  // Early return if verification data is not available
  if (!verificationData) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Loading...</h2>
          <p className="text-gray-600">Preparing secure login...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Header */}
      <div className="flex-shrink-0">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <Link to="/" className="inline-flex items-center space-x-3 group">
            <div className="relative">
              <div className="absolute inset-0 bg-gradient-to-r from-red-600 to-red-800 rounded-xl opacity-0 group-hover:opacity-100 blur-xl transition-opacity duration-300"></div>
              <div className="relative flex items-center justify-center h-10 w-10 bg-gradient-to-br from-red-700 to-red-800 rounded-xl shadow-md group-hover:shadow-lg transform group-hover:scale-105 transition-all duration-200">
                <LockClosedIcon className="h-5 w-5 text-white" />
              </div>
            </div>
            <span className="text-xl font-bold text-gray-900 group-hover:text-red-800 transition-colors duration-200">
              MapleFile
            </span>
          </Link>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex items-center justify-center px-4 sm:px-6 lg:px-8 py-12">
        <div className="w-full max-w-md space-y-8 animate-fade-in-up">
          {/* Progress Indicator */}
          <div className="flex items-center justify-center">
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
                <div className="flex items-center justify-center w-8 h-8 bg-gradient-to-r from-red-700 to-red-800 rounded-full text-white text-sm font-bold shadow-lg">
                  3
                </div>
                <span className="ml-2 text-sm font-semibold text-gray-900">
                  Access
                </span>
              </div>
            </div>
          </div>

          {/* Header */}
          <div className="text-center">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="absolute inset-0 bg-gradient-to-r from-red-600 to-red-800 rounded-2xl blur-2xl opacity-20 animate-pulse"></div>
                <div className="relative h-16 w-16 bg-gradient-to-br from-red-600 to-red-800 rounded-2xl flex items-center justify-center shadow-xl">
                  <KeyIcon className="h-8 w-8 text-white" />
                </div>
              </div>
            </div>
            <h1 className="text-3xl font-bold text-gray-900">
              Unlock Your Account
            </h1>
            <p className="mt-2 text-gray-600">
              Enter your master password to decrypt your data.
            </p>
          </div>

          {/* Form Card */}
          <div
            className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8"
            style={{ animationDelay: "100ms" }}
          >
            {/* Challenge Expired Message */}
            {challengeExpired && (
              <div className="mb-6 p-4 rounded-lg bg-amber-50 border border-amber-200 animate-fade-in">
                <div className="flex items-start">
                  <ClockIcon className="h-5 w-5 text-amber-600 mr-3 flex-shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <h3 className="text-sm font-semibold text-amber-800">
                      Security Code Expired
                    </h3>
                    <p className="text-sm text-amber-700 mt-1 mb-3">
                      Please go back to get a new one-time code.
                    </p>
                    <button
                      onClick={() => navigate("/login/verify-ott")}
                      className="inline-flex items-center justify-center px-3 py-1.5 bg-amber-600 text-white text-sm font-semibold rounded-md hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 transition-all duration-200"
                    >
                      <ArrowLeftIcon className="mr-2 h-4 w-4" />
                      Get New Code
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Error Message */}
            {error && !challengeExpired && (
              <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 animate-fade-in">
                <div className="flex items-center">
                  <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
                  <p className="text-sm text-red-700">{error}</p>
                </div>
              </div>
            )}

            {/* User Display */}
            <div className="mb-6 p-3 bg-gray-50 rounded-lg border border-gray-200">
              <div className="flex items-center">
                <UserIcon className="h-5 w-5 text-gray-400 mr-3" />
                <div className="flex-1">
                  <p className="text-xs font-medium text-gray-500">
                    Signing in as
                  </p>
                  <p className="text-sm text-gray-800 font-semibold">{email}</p>
                </div>
                <button
                  onClick={handleStartOver}
                  disabled={loading}
                  className="text-sm text-red-600 hover:text-red-700 font-medium hover:underline disabled:text-gray-400 disabled:no-underline"
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
                  className="block text-sm font-medium text-gray-700 mb-1"
                >
                  Master Password
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <KeyIcon className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    type={showPassword ? "text" : "password"}
                    id="password"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      if (error) setError("");
                      if (challengeExpired) setChallengeExpired(false);
                    }}
                    placeholder="Enter your master password"
                    required
                    disabled={loading || challengeExpired}
                    className="block w-full pl-10 pr-10 py-3 border border-gray-300 rounded-lg shadow-sm focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-500"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    disabled={loading || challengeExpired}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                  >
                    {showPassword ? (
                      <EyeSlashIcon className="h-5 w-5" />
                    ) : (
                      <EyeIcon className="h-5 w-5" />
                    )}
                  </button>
                </div>
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
                  className="group w-full flex justify-center items-center py-3 px-4 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-red-700 to-red-800 hover:from-red-800 hover:to-red-900 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
                >
                  {loading ? (
                    <>
                      <div className="h-5 w-5 animate-spin border-white mr-3 border-2 border-r-transparent rounded-full"></div>
                      Decrypting & Signing In...
                    </>
                  ) : (
                    <>
                      Decrypt & Access Account
                      <ArrowRightIcon className="ml-2 h-5 w-5 group-hover:translate-x-1 transition-transform" />
                    </>
                  )}
                </button>
                <button
                  type="button"
                  onClick={handleBackToVerify}
                  disabled={loading}
                  className="w-full flex justify-center items-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-400 disabled:bg-gray-50 disabled:cursor-not-allowed transition-all duration-200"
                >
                  <ArrowLeftIcon className="mr-2 h-4 w-4" />
                  Back to Verification
                </button>
              </div>
            </form>
          </div>

          {/* Security Features */}
          <div
            className="grid grid-cols-2 gap-4 animate-fade-in-up"
            style={{ animationDelay: "200ms" }}
          >
            <div className="bg-white rounded-xl border border-gray-100 p-4 text-center shadow-lg">
              <ShieldCheckIcon className="h-8 w-8 text-green-600 mx-auto mb-2" />
              <h3 className="text-sm font-medium text-gray-900">End-to-End</h3>
              <p className="text-xs text-gray-600">Encrypted</p>
            </div>
            <div className="bg-white rounded-xl border border-gray-100 p-4 text-center shadow-lg">
              <KeyIcon className="h-8 w-8 text-blue-600 mx-auto mb-2" />
              <h3 className="text-sm font-medium text-gray-900">
                Zero Knowledge
              </h3>
              <p className="text-xs text-gray-600">Architecture</p>
            </div>
          </div>

          {/* Help Links */}
          <div
            className="text-center space-x-4 animate-fade-in-up"
            style={{ animationDelay: "300ms" }}
          >
            <Link
              to="/recovery"
              className="text-sm text-red-700 hover:text-red-800 font-medium"
            >
              Forgot your password?
            </Link>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="flex-shrink-0 border-t border-gray-200 bg-white/50 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex flex-col sm:flex-row items-center justify-center space-y-4 sm:space-y-0 sm:space-x-8 text-sm">
            <div className="flex items-center space-x-2">
              <ShieldCheckIcon className="h-4 w-4 text-green-600" />
              <span className="text-gray-600">
                ChaCha20-Poly1305 Encryption
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <ServerIcon className="h-4 w-4 text-blue-600" />
              <span className="text-gray-600">Canadian Hosted</span>
            </div>
            <div className="flex items-center space-x-2">
              <GlobeAltIcon className="h-4 w-4 text-purple-600" />
              <span className="text-gray-600">Privacy First</span>
            </div>
            <div className="flex items-center space-x-2">
              <HeartIcon className="h-4 w-4 text-red-600" />
              <span className="text-gray-600">Made in Canada</span>
            </div>
          </div>
        </div>
      </div>

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
          animation: fade-in 0.4s ease-out forwards;
        }
        .animate-fade-in-up {
          animation: fade-in-up 0.6s ease-out forwards;
        }
      `}</style>
    </div>
  );
};

export default CompleteLogin;
