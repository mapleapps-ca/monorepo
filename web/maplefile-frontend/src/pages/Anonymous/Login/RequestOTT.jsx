// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/RequestOTT.jsx
import React, { useState } from "react";
import { useNavigate, Link } from "react-router";
import { useServices } from "../../../services/Services";
import {
  ArrowRightIcon,
  EnvelopeIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  ServerIcon,
  EyeSlashIcon,
} from "@heroicons/react/24/outline";

const RequestOTT = () => {
  const navigate = useNavigate();
  const { authManager } = useServices();
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    setMessage("");

    try {
      // Validate services
      if (!authManager) {
        throw new Error(
          "Authentication service not available. Please refresh the page.",
        );
      }

      // Validate email
      if (!email) {
        throw new Error("Email address is required");
      }

      if (!email.includes("@")) {
        throw new Error("Please enter a valid email address");
      }

      const trimmedEmail = email.trim().toLowerCase();

      console.log(
        "[RequestOTT] Requesting OTT via AuthManager for:",
        trimmedEmail,
      );

      // FIXED: Store email BEFORE making the request
      // Store in sessionStorage as a fallback
      sessionStorage.setItem("loginEmail", trimmedEmail);
      console.log("[RequestOTT] Email stored in sessionStorage:", trimmedEmail);

      // Make the OTT request
      let response;
      if (typeof authManager.requestOTT === "function") {
        response = await authManager.requestOTT(trimmedEmail);
      } else if (typeof authManager.requestOTP === "function") {
        response = await authManager.requestOTP({ email: trimmedEmail });
      } else {
        throw new Error("OTT request method not found on authManager");
      }

      setMessage(response.message || "Verification code sent successfully!");
      console.log("[RequestOTT] OTT request successful via AuthManager");

      // FIXED: Store email in localStorage via authManager if available
      try {
        if (
          authManager.setCurrentUserEmail &&
          typeof authManager.setCurrentUserEmail === "function"
        ) {
          authManager.setCurrentUserEmail(trimmedEmail);
          console.log("[RequestOTT] Email stored via authManager");
        }
      } catch (storageError) {
        console.warn(
          "[RequestOTT] Could not store email via authManager:",
          storageError,
        );
        // Continue anyway as we have sessionStorage fallback
      }

      // Wait a moment to show the success message, then navigate
      setTimeout(() => {
        navigate("/login/verify-ott");
      }, 2000);
    } catch (error) {
      console.error("[RequestOTT] OTT request failed via AuthManager:", error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
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
                <div className="flex items-center justify-center w-8 h-8 bg-gradient-to-r from-red-800 to-red-900 rounded-full text-white text-sm font-bold">
                  1
                </div>
                <span className="ml-2 text-sm font-semibold text-gray-900">
                  Email
                </span>
              </div>
              <div className="w-12 h-0.5 bg-gray-300"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gray-300 rounded-full text-gray-500 text-sm font-bold">
                  2
                </div>
                <span className="ml-2 text-sm text-gray-500">Verify</span>
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
                  <EnvelopeIcon className="h-8 w-8 text-white" />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
              </div>
            </div>
            <h2 className="text-3xl font-black text-gray-900 mb-2">
              Secure Sign In
            </h2>
            <p className="text-gray-600 mb-2">
              Enter your email to receive an encrypted verification code
            </p>
            <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
              <LockClosedIcon className="h-4 w-4 text-green-600" />
              <span>End-to-end encrypted authentication</span>
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
                      Authentication Error
                    </h3>
                    <p className="text-sm text-red-700 mt-1">{error}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Success Message */}
            {message && (
              <div className="mb-6 p-4 rounded-lg bg-green-50 border border-green-200 animate-fade-in">
                <div className="flex items-center">
                  <CheckIcon className="h-5 w-5 text-green-500 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-green-800">
                      Code Sent Successfully
                    </h3>
                    <p className="text-sm text-green-700 mt-1">{message}</p>
                    <p className="text-sm text-green-600 mt-2">
                      Redirecting to verification page...
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              <div>
                <label
                  htmlFor="email"
                  className="block text-sm font-semibold text-gray-700 mb-2"
                >
                  Email Address
                </label>
                <div className="relative">
                  <input
                    type="email"
                    id="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="Enter your email address"
                    required
                    disabled={loading}
                    className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-500 pr-12 ${
                      email && email.includes("@") && email.includes(".")
                        ? "border-green-300 bg-green-50"
                        : email && email.length > 0
                          ? "border-gray-300"
                          : "border-gray-300"
                    }`}
                  />
                  <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                    {email && email.includes("@") && email.includes(".") ? (
                      <CheckIcon className="h-5 w-5 text-green-500" />
                    ) : (
                      <EnvelopeIcon className="h-5 w-5 text-gray-400" />
                    )}
                  </div>
                </div>
                {email && email.length > 0 && !email.includes("@") && (
                  <p className="mt-1 text-xs text-gray-500">
                    Please enter a valid email address
                  </p>
                )}
              </div>

              <button
                type="submit"
                disabled={loading || !authManager || !email.includes("@")}
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
                    Sending Encrypted Code...
                  </>
                ) : (
                  <>
                    Send Secure Verification Code
                    <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
                  </>
                )}
              </button>
            </form>

            {/* Security Trust Section */}
            <div className="mt-8 p-4 bg-gradient-to-r from-green-50 to-blue-50 rounded-lg border border-green-100">
              <div className="flex items-center justify-center mb-3">
                <div className="flex items-center space-x-4">
                  <div className="flex items-center space-x-1">
                    <LockClosedIcon className="h-4 w-4 text-green-600" />
                    <span className="text-xs font-semibold text-green-800">
                      ChaCha20-Poly1305
                    </span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <ServerIcon className="h-4 w-4 text-blue-600" />
                    <span className="text-xs font-semibold text-blue-800">
                      Canadian Hosted
                    </span>
                  </div>
                  <div className="flex items-center space-x-1">
                    <EyeSlashIcon className="h-4 w-4 text-purple-600" />
                    <span className="text-xs font-semibold text-purple-800">
                      Zero Knowledge
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Info Section */}
            <div className="mt-6 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
              <h3 className="text-sm font-semibold text-blue-900 mb-3 flex items-center">
                <InformationCircleIcon className="h-4 w-4 mr-2" />
                What happens next?
              </h3>
              <ul className="text-sm text-blue-800 space-y-2">
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  We'll send an encrypted 6-digit code to your email
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  Check your inbox (and spam folder) for the code
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  Enter the code to decrypt your authentication token
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  Complete secure login with your master password
                </li>
              </ul>
            </div>

            {/* Alternative Actions */}
            <div className="mt-6 flex flex-col sm:flex-row justify-between items-center space-y-2 sm:space-y-0 text-sm">
              <Link
                to="/register"
                className="text-red-600 hover:text-red-700 font-medium hover:underline transition-colors duration-200"
              >
                Don't have an account? Sign up
              </Link>
              <Link
                to="/recovery"
                className="text-gray-600 hover:text-gray-700 font-medium hover:underline transition-colors duration-200"
              >
                Forgot your password?
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

export default RequestOTT;
