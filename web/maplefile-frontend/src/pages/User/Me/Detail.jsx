// File: src/pages/User/Me/Detail.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";
import Navigation from "../../../components/Navigation";
import {
  UserIcon,
  KeyIcon,
  ShieldCheckIcon,
  LockClosedIcon,
  ClipboardDocumentIcon,
  EyeIcon,
  EyeSlashIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  ArrowPathIcon,
  DocumentTextIcon,
  CalendarIcon,
  EnvelopeIcon,
  PhoneIcon,
  GlobeAltIcon,
  ClockIcon,
  ServerIcon,
  EyeSlashIcon as PrivacyIcon,
  HeartIcon,
  SparklesIcon,
} from "@heroicons/react/24/outline";

const MeDetail = () => {
  const navigate = useNavigate();
  const { authManager, meManager } = useAuth();

  const [userProfile, setUserProfile] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [showKeys, setShowKeys] = useState(false);
  const [copiedKey, setCopiedKey] = useState("");

  // Load user profile data
  const loadUserProfile = async () => {
    if (!meManager) return;

    setIsLoading(true);
    setError("");

    try {
      console.log("[Profile] Loading user profile...");
      const profile = await meManager.getMe();
      setUserProfile(profile);
      console.log("[Profile] Profile loaded successfully");
    } catch (err) {
      console.error("[Profile] Failed to load profile:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (meManager && authManager?.isAuthenticated()) {
      loadUserProfile();
    }
  }, [meManager, authManager]);

  // Handle key copy
  const handleCopyKey = async (keyType, keyValue) => {
    try {
      await navigator.clipboard.writeText(keyValue);
      setCopiedKey(keyType);
      setTimeout(() => setCopiedKey(""), 3000);
    } catch (error) {
      console.error("Failed to copy to clipboard:", error);
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = keyValue;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand("copy");
      document.body.removeChild(textArea);
      setCopiedKey(keyType);
      setTimeout(() => setCopiedKey(""), 3000);
    }
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "Unknown";
    return new Date(dateString).toLocaleString();
  };

  // Get role display
  const getRoleInfo = (role) => {
    switch (role) {
      case 1:
        return {
          text: "Root User",
          color: "text-purple-700",
          bg: "bg-purple-100",
        };
      case 2:
        return {
          text: "Company User",
          color: "text-blue-700",
          bg: "bg-blue-100",
        };
      case 3:
        return {
          text: "Individual User",
          color: "text-green-700",
          bg: "bg-green-100",
        };
      default:
        return { text: "Unknown", color: "text-gray-700", bg: "bg-gray-100" };
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      {/* Navigation */}
      <Navigation />

      {/* Main Content */}
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8 animate-fade-in-up">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-4xl font-black text-gray-900 mb-2">
                My Profile 👤
              </h1>
              <p className="text-xl text-gray-600">
                Manage your account settings and encryption keys
              </p>
              <div className="flex items-center space-x-2 mt-2 text-sm text-gray-500">
                <ShieldCheckIcon className="h-4 w-4 text-green-600" />
                <span>Your keys never leave your device</span>
              </div>
            </div>
            <button
              onClick={loadUserProfile}
              disabled={isLoading}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
            >
              <ArrowPathIcon className="h-4 w-4 mr-2" />
              {isLoading ? "Refreshing..." : "Refresh"}
            </button>
          </div>
        </div>

        {/* Loading State */}
        {isLoading && !userProfile && (
          <div className="flex items-center justify-center py-12 animate-fade-in">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading your profile...</p>
            </div>
          </div>
        )}

        {/* Error State */}
        {error && !userProfile && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-6 animate-fade-in">
            <div className="flex items-start">
              <ExclamationTriangleIcon className="h-6 w-6 text-red-500 mr-3 flex-shrink-0 mt-1" />
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-red-800 mb-2">
                  Failed to Load Profile
                </h3>
                <p className="text-red-700">{error}</p>
                <button
                  onClick={loadUserProfile}
                  className="mt-4 inline-flex items-center px-4 py-2 border border-red-300 text-sm font-medium rounded-lg text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
                >
                  <ArrowPathIcon className="h-4 w-4 mr-2" />
                  Try Again
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Profile Content */}
        {userProfile && (
          <>
            {/* Error Banner */}
            {error && (
              <div className="bg-amber-50 border border-amber-200 rounded-xl p-4 mb-6 animate-fade-in">
                <div className="flex items-center">
                  <ExclamationTriangleIcon className="h-5 w-5 text-amber-500 mr-3 flex-shrink-0" />
                  <p className="text-amber-700">{error}</p>
                  <button
                    onClick={() => setError("")}
                    className="ml-auto text-amber-500 hover:text-amber-700"
                  >
                    ✕
                  </button>
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
              {/* Personal Information */}
              <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8 animate-fade-in-up-delay">
                <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl">
                      <UserIcon className="h-6 w-6 text-white" />
                    </div>
                    <h2 className="text-2xl font-bold text-gray-900">
                      Personal Information
                    </h2>
                  </div>
                  <SparklesIcon className="h-6 w-6 text-blue-300" />
                </div>

                <div className="space-y-6">
                  {/* Name */}
                  <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center space-x-3">
                      <UserIcon className="h-5 w-5 text-gray-500" />
                      <div>
                        <p className="text-sm font-medium text-gray-600">
                          Full Name
                        </p>
                        <p className="text-lg font-semibold text-gray-900">
                          {userProfile.first_name} {userProfile.last_name}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Email */}
                  <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center space-x-3">
                      <EnvelopeIcon className="h-5 w-5 text-gray-500" />
                      <div>
                        <p className="text-sm font-medium text-gray-600">
                          Email
                        </p>
                        <p className="text-lg font-semibold text-gray-900 font-mono">
                          {userProfile.email}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Phone */}
                  {userProfile.phone && (
                    <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                      <div className="flex items-center space-x-3">
                        <PhoneIcon className="h-5 w-5 text-gray-500" />
                        <div>
                          <p className="text-sm font-medium text-gray-600">
                            Phone
                          </p>
                          <p className="text-lg font-semibold text-gray-900">
                            {userProfile.phone}
                          </p>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Country */}
                  {userProfile.country && (
                    <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                      <div className="flex items-center space-x-3">
                        <GlobeAltIcon className="h-5 w-5 text-gray-500" />
                        <div>
                          <p className="text-sm font-medium text-gray-600">
                            Country
                          </p>
                          <p className="text-lg font-semibold text-gray-900">
                            {userProfile.country}
                          </p>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Timezone */}
                  {userProfile.timezone && (
                    <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                      <div className="flex items-center space-x-3">
                        <ClockIcon className="h-5 w-5 text-gray-500" />
                        <div>
                          <p className="text-sm font-medium text-gray-600">
                            Timezone
                          </p>
                          <p className="text-lg font-semibold text-gray-900">
                            {userProfile.timezone}
                          </p>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* Account Information */}
              <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8 animate-fade-in-up-delay-2">
                <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-green-500 to-green-600 rounded-xl">
                      <ShieldCheckIcon className="h-6 w-6 text-white" />
                    </div>
                    <h2 className="text-2xl font-bold text-gray-900">
                      Account Details
                    </h2>
                  </div>
                  <CheckIcon className="h-6 w-6 text-green-500" />
                </div>

                <div className="space-y-6">
                  {/* User ID */}
                  <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center space-x-3">
                      <DocumentTextIcon className="h-5 w-5 text-gray-500" />
                      <div>
                        <p className="text-sm font-medium text-gray-600">
                          User ID
                        </p>
                        <p className="text-lg font-semibold text-gray-900 font-mono">
                          {userProfile.id}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* User Role */}
                  <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center space-x-3">
                      <ShieldCheckIcon className="h-5 w-5 text-gray-500" />
                      <div>
                        <p className="text-sm font-medium text-gray-600">
                          Role
                        </p>
                        <span
                          className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-semibold ${getRoleInfo(userProfile.user_role).bg} ${getRoleInfo(userProfile.user_role).color}`}
                        >
                          {getRoleInfo(userProfile.user_role).text}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Registration Date */}
                  <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center space-x-3">
                      <CalendarIcon className="h-5 w-5 text-gray-500" />
                      <div>
                        <p className="text-sm font-medium text-gray-600">
                          Member Since
                        </p>
                        <p className="text-lg font-semibold text-gray-900">
                          {formatDate(userProfile.created_at)}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Last Updated */}
                  <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center space-x-3">
                      <ClockIcon className="h-5 w-5 text-gray-500" />
                      <div>
                        <p className="text-sm font-medium text-gray-600">
                          Last Updated
                        </p>
                        <p className="text-lg font-semibold text-gray-900">
                          {formatDate(userProfile.updated_at)}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Module */}
                  {userProfile.module && (
                    <div className="flex items-center justify-between p-4 bg-gray-50 rounded-xl">
                      <div className="flex items-center space-x-3">
                        <DocumentTextIcon className="h-5 w-5 text-gray-500" />
                        <div>
                          <p className="text-sm font-medium text-gray-600">
                            Service
                          </p>
                          <p className="text-lg font-semibold text-gray-900">
                            {userProfile.module === 1
                              ? "MapleFile"
                              : userProfile.module === 2
                                ? "PaperCloud"
                                : "Unknown"}
                          </p>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Encryption Keys Section */}
            <div className="mt-8 bg-white rounded-2xl shadow-xl border border-gray-100 p-8 animate-fade-in-up-delay-3">
              <div className="flex items-center justify-between mb-6">
                <div className="flex items-center space-x-3">
                  <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-red-800 to-red-900 rounded-xl">
                    <KeyIcon className="h-6 w-6 text-white" />
                  </div>
                  <div>
                    <h2 className="text-2xl font-bold text-gray-900">
                      Encryption Keys
                    </h2>
                    <p className="text-gray-600">
                      Your cryptographic keys for secure file encryption
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => setShowKeys(!showKeys)}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
                >
                  {showKeys ? (
                    <>
                      <EyeSlashIcon className="h-4 w-4 mr-2" />
                      Hide Keys
                    </>
                  ) : (
                    <>
                      <EyeIcon className="h-4 w-4 mr-2" />
                      Show Keys
                    </>
                  )}
                </button>
              </div>

              {/* Security Warning */}
              <div className="bg-red-50 border border-red-200 rounded-xl p-4 mb-6">
                <div className="flex items-start">
                  <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0 mt-0.5" />
                  <div>
                    <h3 className="text-sm font-semibold text-red-800 mb-1">
                      Security Notice
                    </h3>
                    <p className="text-sm text-red-700">
                      Your encryption keys are sensitive information. Never
                      share them with anyone. These keys are used to encrypt and
                      decrypt your files locally.
                    </p>
                  </div>
                </div>
              </div>

              {/* Keys Display */}
              <div className="space-y-4">
                {/* Public Key */}
                {userProfile.public_key && (
                  <div className="p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <KeyIcon className="h-4 w-4 text-gray-500" />
                        <span className="text-sm font-semibold text-gray-700">
                          Public Key
                        </span>
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                          Safe to share
                        </span>
                      </div>
                      <button
                        onClick={() =>
                          handleCopyKey("public", userProfile.public_key)
                        }
                        className="inline-flex items-center px-2 py-1 border border-gray-300 text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
                      >
                        {copiedKey === "public" ? (
                          <>
                            <CheckIcon className="h-3 w-3 mr-1 text-green-500" />
                            Copied!
                          </>
                        ) : (
                          <>
                            <ClipboardDocumentIcon className="h-3 w-3 mr-1" />
                            Copy
                          </>
                        )}
                      </button>
                    </div>
                    <div
                      className={`font-mono text-sm break-all ${showKeys ? "text-gray-900" : "text-gray-400 select-none"} bg-white p-3 rounded border`}
                    >
                      {showKeys ? userProfile.public_key : "•".repeat(64)}
                    </div>
                  </div>
                )}

                {/* Encrypted Private Key */}
                {userProfile.encrypted_private_key && (
                  <div className="p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <LockClosedIcon className="h-4 w-4 text-gray-500" />
                        <span className="text-sm font-semibold text-gray-700">
                          Encrypted Private Key
                        </span>
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800">
                          Keep secret
                        </span>
                      </div>
                      <button
                        onClick={() =>
                          handleCopyKey(
                            "private",
                            userProfile.encrypted_private_key,
                          )
                        }
                        className="inline-flex items-center px-2 py-1 border border-gray-300 text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
                      >
                        {copiedKey === "private" ? (
                          <>
                            <CheckIcon className="h-3 w-3 mr-1 text-green-500" />
                            Copied!
                          </>
                        ) : (
                          <>
                            <ClipboardDocumentIcon className="h-3 w-3 mr-1" />
                            Copy
                          </>
                        )}
                      </button>
                    </div>
                    <div
                      className={`font-mono text-sm break-all ${showKeys ? "text-gray-900" : "text-gray-400 select-none"} bg-white p-3 rounded border`}
                    >
                      {showKeys
                        ? userProfile.encrypted_private_key
                        : "•".repeat(96)}
                    </div>
                  </div>
                )}

                {/* Key Derivation Salt */}
                {userProfile.key_derivation_salt && (
                  <div className="p-4 bg-gray-50 rounded-xl">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <ShieldCheckIcon className="h-4 w-4 text-gray-500" />
                        <span className="text-sm font-semibold text-gray-700">
                          Key Derivation Salt
                        </span>
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                          For recovery
                        </span>
                      </div>
                      <button
                        onClick={() =>
                          handleCopyKey("salt", userProfile.key_derivation_salt)
                        }
                        className="inline-flex items-center px-2 py-1 border border-gray-300 text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
                      >
                        {copiedKey === "salt" ? (
                          <>
                            <CheckIcon className="h-3 w-3 mr-1 text-green-500" />
                            Copied!
                          </>
                        ) : (
                          <>
                            <ClipboardDocumentIcon className="h-3 w-3 mr-1" />
                            Copy
                          </>
                        )}
                      </button>
                    </div>
                    <div
                      className={`font-mono text-sm break-all ${showKeys ? "text-gray-900" : "text-gray-400 select-none"} bg-white p-3 rounded border`}
                    >
                      {showKeys
                        ? userProfile.key_derivation_salt
                        : "•".repeat(64)}
                    </div>
                  </div>
                )}
              </div>

              {/* Key Info */}
              <div className="mt-6 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
                <div className="flex items-start">
                  <InformationCircleIcon className="h-5 w-5 text-blue-600 mr-3 flex-shrink-0 mt-0.5" />
                  <div className="text-sm text-blue-800">
                    <p className="font-semibold mb-2">
                      About Your Encryption Keys:
                    </p>
                    <ul className="space-y-1 text-xs">
                      <li>
                        • <strong>Public Key:</strong> Used by others to encrypt
                        files shared with you
                      </li>
                      <li>
                        • <strong>Private Key:</strong> Encrypted with your
                        password, used to decrypt your files
                      </li>
                      <li>
                        • <strong>Salt:</strong> Used with your recovery phrase
                        to restore your keys
                      </li>
                      <li>
                        • All cryptographic operations happen locally in your
                        browser
                      </li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Trust Badges Footer */}
      <div className="border-t border-gray-100 bg-white/50 backdrop-blur-sm py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center space-x-8 text-sm">
            <div className="flex items-center space-x-2">
              <LockClosedIcon className="h-4 w-4 text-green-600" />
              <span className="text-gray-600 font-medium">
                ChaCha20-Poly1305 Encryption
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <ServerIcon className="h-4 w-4 text-blue-600" />
              <span className="text-gray-600 font-medium">Canadian Hosted</span>
            </div>
            <div className="flex items-center space-x-2">
              <PrivacyIcon className="h-4 w-4 text-purple-600" />
              <span className="text-gray-600 font-medium">Zero Knowledge</span>
            </div>
            <div className="flex items-center space-x-2">
              <HeartIcon className="h-4 w-4 text-red-600" />
              <span className="text-gray-600 font-medium">Made in Canada</span>
            </div>
          </div>
        </div>
      </div>

      {/* CSS Animations */}
      <style jsx>{`
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
          animation: fade-in-up 0.4s ease-out;
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

// Export with password protection
export default withPasswordProtection(MeDetail);
