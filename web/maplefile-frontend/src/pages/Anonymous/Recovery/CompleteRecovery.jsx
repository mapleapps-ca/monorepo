// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Recovery/CompleteRecovery.jsx
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
  DocumentTextIcon,
  CheckCircleIcon,
  ArrowPathIcon,
  LockOpenIcon,
  ServerIcon,
} from "@heroicons/react/24/outline";

const CompleteRecovery = () => {
  const navigate = useNavigate();
  const { recoveryManager } = useServices();
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [recoveryPhrase, setRecoveryPhrase] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [email, setEmail] = useState("");
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [showRecoveryPhrase, setShowRecoveryPhrase] = useState(false);

  useEffect(() => {
    // Check if we have completed verification
    const recoveryEmail = recoveryManager.getRecoveryEmail();
    const isVerified = recoveryManager.isVerificationComplete();

    if (!recoveryEmail || !isVerified) {
      console.log(
        "[CompleteRecovery] No verified recovery session, redirecting",
      );
      navigate("/recovery/initiate");
      return;
    }

    setEmail(recoveryEmail);
  }, [navigate, recoveryManager]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      // Validate passwords
      if (!newPassword) {
        throw new Error("Password is required");
      }

      if (newPassword.length < 8) {
        throw new Error("Password must be at least 8 characters long");
      }

      if (newPassword !== confirmPassword) {
        throw new Error("Passwords do not match");
      }

      // Validate recovery phrase again
      const words = recoveryPhrase.trim().toLowerCase().split(/\s+/);
      if (words.length !== 12) {
        throw new Error("Recovery phrase must be exactly 12 words");
      }

      // Join words with single space
      const normalizedPhrase = words.join(" ");

      console.log("[CompleteRecovery] Completing recovery with new password");

      // Complete recovery with both recovery phrase and new password
      const response = await recoveryManager.completeRecoveryWithPhrase(
        normalizedPhrase,
        newPassword,
      );

      console.log("[CompleteRecovery] Recovery completed successfully");

      // Show success modal or alert
      alert(
        "Account recovery completed successfully! You can now log in with your new password.",
      );

      // Navigate to login
      navigate("/login");
    } catch (error) {
      console.error("[CompleteRecovery] Recovery completion failed:", error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleBackToVerify = () => {
    navigate("/recovery/verify");
  };

  if (!email) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Loading...</h2>
          <p className="text-gray-600">Checking recovery session...</p>
        </div>
      </div>
    );
  }

  // Count words in recovery phrase
  const wordCount = recoveryPhrase.trim()
    ? recoveryPhrase.trim().split(/\s+/).length
    : 0;

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
        <div className="max-w-2xl w-full space-y-8">
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
                  Reset
                </span>
              </div>
            </div>
          </div>

          {/* Header */}
          <div className="text-center animate-fade-in-up">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="flex items-center justify-center h-16 w-16 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-lg animate-pulse">
                  <LockOpenIcon className="h-8 w-8 text-white" />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
              </div>
            </div>
            <h2 className="text-3xl font-black text-gray-900 mb-2">
              Set Your New Password
            </h2>
            <p className="text-gray-600 mb-2">
              Final step: Create a new password for {email}
            </p>
            <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
              <ArrowPathIcon className="h-4 w-4 text-green-600" />
              <span>
                Your encryption keys will be re-encrypted with the new password
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
                      Recovery Error
                    </h3>
                    <p className="text-sm text-red-700 mt-1">{error}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Security Notice */}
            <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <div className="flex items-start">
                <InformationCircleIcon className="h-5 w-5 text-blue-600 mr-3 flex-shrink-0 mt-0.5" />
                <div className="flex-1">
                  <h3 className="text-sm font-semibold text-blue-800 mb-1">
                    Why enter your recovery phrase again?
                  </h3>
                  <p className="text-sm text-blue-700">
                    We need your recovery phrase to decrypt your master key and
                    re-encrypt it with your new password. This ensures
                    continuous access to your encrypted files.
                  </p>
                </div>
              </div>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Recovery Phrase */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label
                    htmlFor="recoveryPhrase"
                    className="block text-sm font-semibold text-gray-700"
                  >
                    Recovery Phrase (Required Again)
                  </label>
                  <span
                    className={`text-xs font-medium ${
                      wordCount === 12
                        ? "text-green-600"
                        : wordCount > 0
                          ? "text-amber-600"
                          : "text-gray-500"
                    }`}
                  >
                    {wordCount}/12 words
                  </span>
                </div>
                <div className="relative">
                  <textarea
                    id="recoveryPhrase"
                    value={recoveryPhrase}
                    onChange={(e) => setRecoveryPhrase(e.target.value)}
                    placeholder="Re-enter your 12-word recovery phrase"
                    rows={3}
                    required
                    disabled={loading}
                    className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-500 font-mono text-sm leading-relaxed resize-none ${
                      wordCount === 12
                        ? "border-green-300 bg-green-50"
                        : "border-gray-300"
                    }`}
                  />
                  {showRecoveryPhrase && (
                    <button
                      type="button"
                      onClick={() => setShowRecoveryPhrase(!showRecoveryPhrase)}
                      className="absolute top-3 right-3"
                    >
                      <EyeSlashIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                    </button>
                  )}
                </div>
              </div>

              {/* New Password Section */}
              <div className="space-y-4 p-4 bg-gradient-to-r from-green-50 to-emerald-50 rounded-lg border border-green-100">
                <h3 className="text-sm font-semibold text-green-900 flex items-center">
                  <KeyIcon className="h-4 w-4 mr-2" />
                  Create Your New Password
                </h3>

                {/* New Password */}
                <div>
                  <label
                    htmlFor="newPassword"
                    className="block text-sm font-semibold text-gray-700 mb-2"
                  >
                    New Password
                  </label>
                  <div className="relative">
                    <input
                      type={showNewPassword ? "text" : "password"}
                      id="newPassword"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      placeholder="Enter your new password"
                      required
                      disabled={loading}
                      autoComplete="new-password"
                      className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 pr-12 ${
                        error && error.includes("password")
                          ? "border-red-300"
                          : "border-gray-300"
                      }`}
                    />
                    <button
                      type="button"
                      onClick={() => setShowNewPassword(!showNewPassword)}
                      className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    >
                      {showNewPassword ? (
                        <EyeSlashIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                      ) : (
                        <EyeIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                      )}
                    </button>
                  </div>
                  <p className="mt-1 text-xs text-gray-500">
                    Password must be at least 8 characters long
                  </p>
                </div>

                {/* Confirm Password */}
                <div>
                  <label
                    htmlFor="confirmPassword"
                    className="block text-sm font-semibold text-gray-700 mb-2"
                  >
                    Confirm New Password
                  </label>
                  <div className="relative">
                    <input
                      type={showConfirmPassword ? "text" : "password"}
                      id="confirmPassword"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      placeholder="Confirm your new password"
                      required
                      disabled={loading}
                      autoComplete="new-password"
                      className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 pr-12 ${
                        confirmPassword && newPassword === confirmPassword
                          ? "border-green-300 bg-green-50"
                          : "border-gray-300"
                      }`}
                    />
                    <button
                      type="button"
                      onClick={() =>
                        setShowConfirmPassword(!showConfirmPassword)
                      }
                      className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    >
                      {showConfirmPassword ? (
                        <EyeSlashIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                      ) : (
                        <EyeIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                      )}
                    </button>
                  </div>
                  {confirmPassword && newPassword === confirmPassword && (
                    <p className="mt-1 text-xs text-green-600 flex items-center">
                      <CheckIcon className="h-3 w-3 mr-1" />
                      Passwords match
                    </p>
                  )}
                </div>
              </div>

              <div className="space-y-3">
                <button
                  type="submit"
                  disabled={
                    loading ||
                    wordCount !== 12 ||
                    !newPassword ||
                    newPassword !== confirmPassword
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
                      Setting New Password...
                    </>
                  ) : (
                    <>
                      <CheckCircleIcon className="mr-2 h-4 w-4" />
                      Complete Recovery
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
          </div>

          {/* What Happens Next */}
          <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100 p-6 animate-fade-in-up-delay-2">
            <h3 className="text-sm font-semibold text-blue-900 mb-3 flex items-center">
              <InformationCircleIcon className="h-4 w-4 mr-2" />
              What Happens Next?
            </h3>
            <ul className="text-sm text-blue-800 space-y-2">
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                Your master key will be decrypted using your recovery key
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                New encryption keys will be generated
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                All keys will be re-encrypted with your new password
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                Your recovery phrase remains the same for future use
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                You'll be able to log in immediately with your new password
              </li>
            </ul>
          </div>

          {/* Security Notes */}
          <div className="bg-gray-50 rounded-lg border border-gray-200 p-4 animate-fade-in-up-delay-3">
            <h4 className="text-xs font-semibold text-gray-700 mb-2 flex items-center">
              <ShieldCheckIcon className="h-4 w-4 mr-1" />
              Security Notes
            </h4>
            <div className="text-xs text-gray-600 space-y-1">
              <p>• Choose a strong, unique password</p>
              <p>• Your new password will be used to encrypt your keys</p>
              <p>• Keep your recovery phrase safe - it hasn't changed</p>
              <p>• All your encrypted data remains accessible</p>
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

        .animate-fade-in-up-delay-3 {
          animation: fade-in-up 0.6s ease-out 0.6s both;
        }
      `}</style>
    </div>
  );
};

export default CompleteRecovery;
