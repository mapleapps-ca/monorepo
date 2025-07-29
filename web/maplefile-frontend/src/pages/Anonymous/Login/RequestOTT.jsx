// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/RequestOTT.jsx
import { useState } from "react";
import { useNavigate, Link } from "react-router";
import { useServices } from "../../../services/Services";
import {
  ArrowRightIcon,
  EnvelopeIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  ServerIcon,
  KeyIcon,
  SparklesIcon,
  GlobeAltIcon,
  HeartIcon,
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
    } catch (err) {
      console.error("[RequestOTT] OTT request failed via AuthManager:", err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === "Enter" && email) {
      handleSubmit(e);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Header */}
      <div className="flex-shrink-0">
        <nav className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex justify-between items-center">
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
        </nav>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex items-center justify-center px-4 sm:px-6 lg:px-8 py-12">
        <div className="w-full max-w-md space-y-8 animate-fade-in-up">
          {/* Progress Indicator */}
          <div className="flex items-center justify-center">
            <div className="flex items-center space-x-4">
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gradient-to-r from-red-700 to-red-800 rounded-full text-white text-sm font-bold shadow-md">
                  1
                </div>
                <span className="ml-2 text-sm font-semibold text-gray-900">
                  Email
                </span>
              </div>
              <div className="w-12 h-0.5 bg-gray-300"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gray-200 rounded-full text-gray-500 text-sm font-bold">
                  2
                </div>
                <span className="ml-2 text-sm text-gray-500">Verify</span>
              </div>
              <div className="w-12 h-0.5 bg-gray-300"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gray-200 rounded-full text-gray-500 text-sm font-bold">
                  3
                </div>
                <span className="ml-2 text-sm text-gray-500">Access</span>
              </div>
            </div>
          </div>

          {/* Welcome Message */}
          <div className="text-center">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="absolute inset-0 bg-gradient-to-r from-red-600 to-red-800 rounded-2xl blur-2xl opacity-20"></div>
                <div className="relative h-16 w-16 bg-gradient-to-br from-red-600 to-red-800 rounded-2xl flex items-center justify-center shadow-xl">
                  <EnvelopeIcon className="h-8 w-8 text-white" />
                </div>
              </div>
            </div>
            <h1 className="text-3xl font-bold text-gray-900">Secure Sign In</h1>
            <p className="mt-2 text-gray-600">
              Enter your email to receive a login code
            </p>
          </div>

          {/* Login Form */}
          <div
            className="bg-white border border-gray-100 rounded-2xl p-8 shadow-xl animate-fade-in-up"
            style={{ animationDelay: "100ms" }}
          >
            <form onSubmit={handleSubmit} className="space-y-6">
              {error && (
                <div className="p-4 rounded-lg bg-red-50 border border-red-200 animate-fade-in">
                  <div className="flex items-start">
                    <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
                    <p className="text-sm text-red-700">{error}</p>
                  </div>
                </div>
              )}

              {message && (
                <div className="p-4 rounded-lg bg-green-50 border border-green-200 animate-fade-in">
                  <div className="flex items-start">
                    <CheckIcon className="h-5 w-5 text-green-500 mr-3 flex-shrink-0" />
                    <div>
                      <p className="text-sm text-green-700">{message}</p>
                      <p className="text-sm text-green-600 mt-1">
                        Redirecting to verification...
                      </p>
                    </div>
                  </div>
                </div>
              )}

              <div>
                <label
                  htmlFor="email"
                  className="block text-sm font-medium text-gray-700 mb-1.5"
                >
                  Email Address
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <EnvelopeIcon className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="email"
                    name="email"
                    type="email"
                    autoComplete="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    onKeyPress={handleKeyPress}
                    className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-500"
                    placeholder="you@example.com"
                  />
                </div>
              </div>

              <div>
                <button
                  type="submit"
                  disabled={loading || !authManager || !email.includes("@")}
                  className="w-full flex justify-center items-center py-3 px-4 border border-transparent text-base font-medium rounded-lg text-white bg-gradient-to-r from-red-700 to-red-800 hover:from-red-800 hover:to-red-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
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
                      Sending login code...
                    </>
                  ) : (
                    <>
                      Continue with Email
                      <ArrowRightIcon className="h-5 w-5 ml-2" />
                    </>
                  )}
                </button>
              </div>

              <div className="relative pt-2">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-200"></div>
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-4 bg-white text-gray-500">
                    New to MapleFile?
                  </span>
                </div>
              </div>

              <div>
                <Link
                  to="/register"
                  className="w-full flex items-center justify-center py-3 text-base font-medium border border-gray-300 rounded-lg text-gray-800 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-400 transition-all duration-200"
                >
                  Create an Account
                  <SparklesIcon className="h-5 w-5 ml-2 text-red-600" />
                </Link>
              </div>
            </form>
          </div>

          {/* Security Features */}
          <div
            className="grid grid-cols-2 gap-4 animate-fade-in-up"
            style={{ animationDelay: "200ms" }}
          >
            <div className="bg-white border border-gray-100 rounded-xl p-4 text-center shadow-lg">
              <ShieldCheckIcon className="h-8 w-8 text-green-600 mx-auto mb-2" />
              <h3 className="text-sm font-medium text-gray-900">End-to-End</h3>
              <p className="text-xs text-gray-600">Encrypted</p>
            </div>
            <div className="bg-white border border-gray-100 rounded-xl p-4 text-center shadow-lg">
              <KeyIcon className="h-8 w-8 text-blue-600 mx-auto mb-2" />
              <h3 className="text-sm font-medium text-gray-900">
                Zero Knowledge
              </h3>
              <p className="text-xs text-gray-600">Architecture</p>
            </div>
          </div>

          {/* Help Links */}
          <div
            className="text-center space-y-2 animate-fade-in-up"
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
      <div className="flex-shrink-0 border-t border-gray-200/75 bg-gray-50/75 backdrop-blur-sm">
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
    </div>
  );
};

export default RequestOTT;
