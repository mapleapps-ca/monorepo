// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Register/VerifyEmail.jsx
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
  EnvelopeIcon,
  ClockIcon,
  KeyIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";

const VerifyEmail = () => {
  const navigate = useNavigate();
  const { authManager } = useServices();
  const [email, setEmail] = useState("");
  const [verificationCode, setVerificationCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [resendLoading, setResendLoading] = useState(false);
  const [resendSuccess, setResendSuccess] = useState(false);
  const [timeRemaining, setTimeRemaining] = useState(4320); // 72 hours in minutes

  useEffect(() => {
    // Get email from sessionStorage
    const registeredEmail = sessionStorage.getItem("registeredEmail");

    if (!registeredEmail) {
      // Redirect back to registration if no email found
      navigate("/register");
      return;
    }

    setEmail(registeredEmail);
  }, [navigate]);

  // Countdown timer effect (simplified for 72 hours)
  useEffect(() => {
    const timer = setInterval(() => {
      setTimeRemaining((prev) => {
        if (prev <= 0) return 0;
        return prev - 1;
      });
    }, 60000); // Update every minute

    return () => clearInterval(timer);
  }, []);

  // Format time as hours
  const formatTime = (minutes) => {
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    if (hours > 24) {
      const days = Math.floor(hours / 24);
      const remainingHours = hours % 24;
      return `${days} day${days > 1 ? "s" : ""}, ${remainingHours} hour${remainingHours !== 1 ? "s" : ""}`;
    }
    return `${hours} hour${hours !== 1 ? "s" : ""}, ${remainingMinutes} minute${remainingMinutes !== 1 ? "s" : ""}`;
  };

  const handleInputChange = (e) => {
    const value = e.target.value.replace(/\D/g, ""); // Only allow digits
    if (value.length <= 6) {
      setVerificationCode(value);
      if (error) setError("");
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (verificationCode.length !== 6) {
      setError("Please enter a 6-digit verification code");
      return;
    }

    setLoading(true);
    setError("");

    try {
      console.log("Verifying email with code:", verificationCode);
      const result = await authManager.verifyEmail(verificationCode);

      console.log("Email verification successful:", result);

      // Store user role for success page
      sessionStorage.setItem("userRole", result.user_role.toString());

      // Navigate to success page
      navigate("/register/verify-success");
    } catch (error) {
      console.error("Email verification failed:", error);
      setError(
        error.message ||
          "Verification failed. Please check your code and try again.",
      );
    } finally {
      setLoading(false);
    }
  };

  const handleResendCode = async () => {
    setResendLoading(true);
    setError("");
    setResendSuccess(false);

    try {
      // In a real app, you'd implement a resend verification email endpoint
      // For now, we'll simulate it
      await new Promise((resolve) => setTimeout(resolve, 2000));
      setResendSuccess(true);
      setVerificationCode(""); // Clear the current code

      // Clear success message after 5 seconds
      setTimeout(() => {
        setResendSuccess(false);
      }, 5000);
    } catch (err) {
      setError("Failed to resend code. Please try again.");
    } finally {
      setResendLoading(false);
    }
  };

  const handleBackToRecovery = () => {
    navigate("/register/recovery");
  };

  const handleBackToRegistration = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    navigate("/register");
  };

  if (!email) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Loading...</h2>
          <p className="text-gray-600">Loading verification page...</p>
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
              <span className="text-base font-medium text-gray-500">
                Step 3 of 3
              </span>
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
                  Register
                </span>
              </div>
              <div className="w-12 h-0.5 bg-green-500"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-green-500 rounded-full text-white text-sm font-bold">
                  <CheckIcon className="h-4 w-4" />
                </div>
                <span className="ml-2 text-sm font-semibold text-green-600">
                  Recovery
                </span>
              </div>
              <div className="w-12 h-0.5 bg-green-500"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gradient-to-r from-red-800 to-red-900 rounded-full text-white text-sm font-bold">
                  3
                </div>
                <span className="ml-2 text-sm font-semibold text-gray-900">
                  Verify
                </span>
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
              Verify Your Email
            </h2>
            <p className="text-gray-600 mb-2">
              Enter the 6-digit code sent to your email
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
                    Verification email sent to:
                  </p>
                  <p className="text-sm text-gray-900 font-mono">{email}</p>
                </div>
              </div>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              <div>
                <label
                  htmlFor="verification_code"
                  className="block text-sm font-semibold text-gray-700 mb-2"
                >
                  Verification Code
                </label>
                <div className="relative">
                  <input
                    type="text"
                    id="verification_code"
                    name="verification_code"
                    value={verificationCode}
                    onChange={handleInputChange}
                    placeholder="000000"
                    maxLength={6}
                    required
                    disabled={loading}
                    className={`w-full px-4 py-4 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-400 text-center text-2xl font-mono tracking-widest ${
                      verificationCode.length === 6
                        ? "border-green-300 bg-green-50"
                        : "border-gray-300"
                    }`}
                  />
                </div>
                <p className="mt-2 text-xs text-gray-500 text-center">
                  Check your email inbox and spam folder
                </p>
              </div>

              <div className="space-y-3">
                <button
                  type="submit"
                  disabled={
                    loading ||
                    verificationCode.length !== 6 ||
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
                      Verifying...
                    </>
                  ) : timeRemaining <= 0 ? (
                    "Code Expired - Resend Required"
                  ) : (
                    <>
                      <CheckIcon className="mr-2 h-4 w-4" />
                      Verify Email & Complete
                      <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
                    </>
                  )}
                </button>

                <div className="flex space-x-3">
                  <button
                    type="button"
                    onClick={handleResendCode}
                    disabled={resendLoading}
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
                      <>
                        <ArrowPathIcon className="mr-2 h-4 w-4" />
                        Resend Code
                      </>
                    )}
                  </button>

                  <button
                    type="button"
                    onClick={handleBackToRecovery}
                    disabled={loading}
                    className="flex-1 flex justify-center items-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 disabled:bg-gray-50 disabled:cursor-not-allowed transition-all duration-200"
                  >
                    <ArrowLeftIcon className="mr-2 h-4 w-4" />
                    Back to Recovery
                  </button>
                </div>
              </div>
            </form>

            {/* Start Over Link */}
            <div className="mt-6 text-center">
              <button
                onClick={handleBackToRegistration}
                className="text-sm text-red-600 hover:text-red-700 font-medium hover:underline transition-colors duration-200"
              >
                Start registration over
              </button>
            </div>
          </div>

          {/* Help Section */}
          <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100 p-6 animate-fade-in-up-delay-2">
            <h3 className="text-sm font-semibold text-blue-900 mb-3 flex items-center">
              <InformationCircleIcon className="h-4 w-4 mr-2" />
              Having trouble?
            </h3>
            <ul className="text-sm text-blue-800 space-y-2">
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                Check your spam/junk folder for the email
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                Make sure you're checking {email}
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                The verification code expires after 72 hours
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                Click "Resend Code" to get a new verification email
              </li>
            </ul>
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
                to="/privacy"
                className="hover:text-gray-700 transition-colors duration-200"
              >
                Privacy Policy
              </Link>
              <Link
                to="/terms"
                className="hover:text-gray-700 transition-colors duration-200"
              >
                Terms of Service
              </Link>
              <Link
                to="/support"
                className="hover:text-gray-700 transition-colors duration-200"
              >
                Support
              </Link>
            </div>
          </div>
        </div>
      </footer>

      {/* CSS Animations */}
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

        .animate-fade-in-up-delay-2 {
          animation: fade-in-up 0.6s ease-out 0.4s both;
        }
      `}</style>
    </div>
  );
};

export default VerifyEmail;
