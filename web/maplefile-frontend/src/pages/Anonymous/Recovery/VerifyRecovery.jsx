// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Recovery/VerifyRecovery.jsx
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
  ClipboardDocumentIcon,
  DocumentTextIcon,
  ExclamationCircleIcon,
  ClockIcon,
  CpuChipIcon,
} from "@heroicons/react/24/outline";

const VerifyRecovery = () => {
  const navigate = useNavigate();
  const { recoveryManager } = useServices();
  const [recoveryPhrase, setRecoveryPhrase] = useState("");
  const [loading, setLoading] = useState(false);
  const [decrypting, setDecrypting] = useState(false);
  const [error, setError] = useState("");
  const [email, setEmail] = useState("");
  const [sessionInfo, setSessionInfo] = useState(null);
  const [pasted, setPasted] = useState(false);

  useEffect(() => {
    // Check if we have an active recovery session
    const recoveryEmail = recoveryManager.getRecoveryEmail();
    const hasSession = recoveryManager.hasActiveRecoverySession();

    if (!recoveryEmail || !hasSession) {
      console.log("[VerifyRecovery] No active recovery session, redirecting");
      navigate("/recovery/initiate");
      return;
    }

    setEmail(recoveryEmail);

    // Get session info for display
    const status = recoveryManager.getRecoveryStatus();
    setSessionInfo(status.sessionData);
  }, [navigate, recoveryManager]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setDecrypting(true);
    setError("");

    try {
      // Validate recovery phrase
      const words = recoveryPhrase.trim().toLowerCase().split(/\s+/);
      if (words.length !== 12) {
        throw new Error("Recovery phrase must be exactly 12 words");
      }

      // Join words with single space
      const normalizedPhrase = words.join(" ");

      console.log("[VerifyRecovery] Decrypting challenge with recovery phrase");

      // Decrypt the challenge using recovery phrase
      const decryptedChallenge =
        await recoveryManager.decryptChallengeWithRecoveryPhrase(
          normalizedPhrase,
        );

      setDecrypting(false);
      console.log(
        "[VerifyRecovery] Challenge decrypted, verifying with server",
      );

      // Verify the decrypted challenge
      const response = await recoveryManager.verifyRecovery(decryptedChallenge);

      console.log("[VerifyRecovery] Recovery verified successfully");

      // Navigate to completion step
      navigate("/recovery/complete");
    } catch (error) {
      console.error("[VerifyRecovery] Recovery verification failed:", error);
      setError(error.message);
      setDecrypting(false);
    } finally {
      setLoading(false);
    }
  };

  const handleBackToInitiate = () => {
    recoveryManager.clearRecoverySession();
    navigate("/recovery/initiate");
  };

  const handlePasteRecoveryPhrase = async () => {
    try {
      const text = await navigator.clipboard.readText();
      setRecoveryPhrase(text.trim());
      setPasted(true);
      setTimeout(() => setPasted(false), 3000);
    } catch (error) {
      console.error("Failed to read clipboard:", error);
      alert("Failed to paste from clipboard. Please paste manually.");
    }
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
                Step 2 of 3
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
                <span className="ml-2 text-sm text-gray-500">Reset</span>
              </div>
            </div>
          </div>

          {/* Header */}
          <div className="text-center animate-fade-in-up">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="flex items-center justify-center h-16 w-16 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-lg animate-pulse">
                  <DocumentTextIcon className="h-8 w-8 text-white" />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
              </div>
            </div>
            <h2 className="text-3xl font-black text-gray-900 mb-2">
              Verify Your Recovery Phrase
            </h2>
            <p className="text-gray-600 mb-2">
              Enter your 12-word recovery phrase for {email}
            </p>
            <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
              <CpuChipIcon className="h-4 w-4 text-green-600" />
              <span>Local decryption using your recovery key</span>
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

            {/* Decrypting Message */}
            {decrypting && (
              <div className="mb-6 p-4 rounded-lg bg-blue-50 border border-blue-200 animate-fade-in">
                <div className="flex items-center">
                  <svg
                    className="animate-spin h-5 w-5 text-blue-500 mr-3"
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
                  <span className="text-sm text-blue-700">
                    Decrypting challenge with your recovery key...
                  </span>
                </div>
              </div>
            )}

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label
                    htmlFor="recoveryPhrase"
                    className="block text-sm font-semibold text-gray-700"
                  >
                    Recovery Phrase
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
                    placeholder="Enter your 12-word recovery phrase separated by spaces"
                    rows={4}
                    required
                    disabled={loading}
                    className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-500 font-mono text-sm leading-relaxed resize-none ${
                      wordCount === 12
                        ? "border-green-300 bg-green-50"
                        : "border-gray-300"
                    }`}
                  />
                </div>
                <div className="mt-2 flex items-center justify-between">
                  <p className="text-xs text-gray-500">
                    Enter all 12 words in the correct order, separated by spaces
                  </p>
                  <button
                    type="button"
                    onClick={handlePasteRecoveryPhrase}
                    disabled={loading}
                    className={`inline-flex items-center px-3 py-1 text-xs font-medium rounded-md transition-all duration-200 ${
                      pasted
                        ? "bg-green-100 text-green-700 border border-green-300"
                        : "bg-gray-100 text-gray-700 border border-gray-300 hover:bg-gray-200"
                    }`}
                  >
                    {pasted ? (
                      <>
                        <CheckIcon className="h-3 w-3 mr-1" />
                        Pasted!
                      </>
                    ) : (
                      <>
                        <ClipboardDocumentIcon className="h-3 w-3 mr-1" />
                        Paste
                      </>
                    )}
                  </button>
                </div>
              </div>

              <div className="space-y-3">
                <button
                  type="submit"
                  disabled={loading || wordCount !== 12}
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
                      {decrypting ? "Decrypting..." : "Verifying..."}
                    </>
                  ) : (
                    <>
                      <KeyIcon className="mr-2 h-4 w-4" />
                      Verify Recovery Phrase
                      <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
                    </>
                  )}
                </button>

                <button
                  type="button"
                  onClick={handleBackToInitiate}
                  disabled={loading}
                  className="w-full flex justify-center items-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 disabled:bg-gray-50 disabled:cursor-not-allowed transition-all duration-200"
                >
                  <ArrowLeftIcon className="mr-2 h-4 w-4" />
                  Start Over
                </button>
              </div>
            </form>
          </div>

          {/* Recovery Phrase Format */}
          <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100 p-6 animate-fade-in-up-delay-2">
            <h3 className="text-sm font-semibold text-blue-900 mb-3 flex items-center">
              <InformationCircleIcon className="h-4 w-4 mr-2" />
              Recovery Phrase Format
            </h3>
            <ul className="text-sm text-blue-800 space-y-2">
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                Your recovery phrase consists of exactly 12 words
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                Words should be separated by spaces
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                The order of words matters - enter them exactly as saved
              </li>
              <li className="flex items-start">
                <span className="text-blue-500 mr-2 mt-0.5">•</span>
                Words are case-insensitive (converted to lowercase)
              </li>
            </ul>
          </div>

          {/* Security Information */}
          <div className="bg-gray-50 rounded-lg border border-gray-200 p-4 animate-fade-in-up-delay-3">
            <h4 className="text-xs font-semibold text-gray-700 mb-2 flex items-center">
              <ShieldCheckIcon className="h-4 w-4 mr-1" />
              Security Information
            </h4>
            <div className="text-xs text-gray-600 space-y-1">
              <p>• Your recovery phrase is never sent to our servers</p>
              <p>• It's used locally to derive your recovery key</p>
              <p>• The recovery key decrypts a challenge to prove ownership</p>
              <p>• Session expires in 10 minutes for security</p>
            </div>
          </div>

          {/* Session Info (Debug) */}
          {sessionInfo && import.meta.env.DEV && (
            <div className="bg-gray-100 rounded-lg p-3 text-xs text-gray-600 animate-fade-in-up-delay-3">
              <h4 className="font-semibold mb-1">Session Information (Dev)</h4>
              <p>Session ID: {sessionInfo.sessionId ? "Active" : "None"}</p>
              <p>
                Challenge ID: {sessionInfo.challengeId ? "Present" : "None"}
              </p>
              <p>
                Encrypted Challenge:{" "}
                {sessionInfo.encryptedChallenge ? "Ready" : "Missing"}
              </p>
            </div>
          )}
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

export default VerifyRecovery;
