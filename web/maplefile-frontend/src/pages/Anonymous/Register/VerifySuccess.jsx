// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Register/VerifySuccess.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router";
import {
  ArrowRightIcon,
  CheckCircleIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  UserIcon,
  DocumentCheckIcon,
  KeyIcon,
  CloudArrowUpIcon,
  SparklesIcon,
  InformationCircleIcon,
  ServerIcon,
  EyeSlashIcon,
  GlobeAltIcon,
  HeartIcon,
} from "@heroicons/react/24/outline";

const VerifySuccess = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [userRole, setUserRole] = useState(null);
  const [countdown, setCountdown] = useState(10);

  useEffect(() => {
    // Get data from sessionStorage
    const registeredEmail = sessionStorage.getItem("registeredEmail");
    const storedUserRole = sessionStorage.getItem("userRole");

    if (!registeredEmail || !storedUserRole) {
      // Redirect back to registration if no data found
      navigate("/register");
      return;
    }

    setEmail(registeredEmail);
    setUserRole(parseInt(storedUserRole));
  }, [navigate]);

  // Auto-redirect countdown
  useEffect(() => {
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          navigate("/login");
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [navigate]);

  const getUserRoleText = (role) => {
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

  const handleRegisterAnother = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    sessionStorage.removeItem("userRole");
    navigate("/register");
  };

  const handleGoToLogin = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    sessionStorage.removeItem("userRole");
    // Navigate to login page
    navigate("/login");
  };

  if (!email || userRole === null) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Loading...</h2>
          <p className="text-gray-600">Loading success page...</p>
        </div>
      </div>
    );
  }

  const roleInfo = getUserRoleText(userRole);

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
              <div className="flex items-center space-x-2 text-sm text-green-600">
                <CheckCircleIcon className="h-5 w-5" />
                <span className="font-semibold">Registration Complete</span>
              </div>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <div className="flex-1 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-2xl w-full space-y-8">
          {/* Success Animation */}
          <div className="text-center animate-fade-in-up">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="flex items-center justify-center h-20 w-20 bg-gradient-to-br from-green-500 to-green-600 rounded-full shadow-lg animate-bounce-once">
                  <CheckCircleIcon className="h-12 w-12 text-white" />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-green-500 to-green-600 rounded-full blur opacity-30 animate-pulse"></div>
                <div className="absolute -top-1 -right-1">
                  <SparklesIcon className="h-6 w-6 text-yellow-500 animate-pulse" />
                </div>
              </div>
            </div>
            <h2 className="text-4xl font-black text-gray-900 mb-2">
              Welcome to MapleFile! 🎉
            </h2>
            <p className="text-xl text-gray-600">
              Your account has been successfully created
            </p>
          </div>

          {/* Account Details Card */}
          <div className="bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 animate-fade-in-up-delay">
            <div className="flex items-center mb-6">
              <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-red-800 to-red-900 rounded-xl mr-4">
                <UserIcon className="h-6 w-6 text-white" />
              </div>
              <h3 className="text-xl font-semibold text-gray-900">
                Account Details
              </h3>
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <span className="text-sm font-medium text-gray-600">
                  Email Address
                </span>
                <span className="text-sm font-mono text-gray-900">{email}</span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <span className="text-sm font-medium text-gray-600">
                  User Role
                </span>
                <span
                  className={`text-sm font-semibold px-3 py-1 rounded-full ${roleInfo.bg} ${roleInfo.color}`}
                >
                  {roleInfo.text}
                </span>
              </div>
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <span className="text-sm font-medium text-gray-600">
                  Account Status
                </span>
                <span className="text-sm font-semibold text-green-700 flex items-center">
                  <CheckCircleIcon className="h-4 w-4 mr-1" />
                  Active
                </span>
              </div>
            </div>
          </div>

          {/* What's Next Section */}
          <div className="bg-gradient-to-br from-blue-50 to-indigo-50 rounded-2xl border border-blue-100 p-8 animate-fade-in-up-delay-2">
            <h3 className="text-lg font-semibold text-blue-900 mb-4 flex items-center">
              <InformationCircleIcon className="h-5 w-5 mr-2" />
              What's Next?
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-start">
                <div className="flex items-center justify-center h-8 w-8 bg-blue-100 rounded-lg mr-3 flex-shrink-0">
                  <KeyIcon className="h-4 w-4 text-blue-600" />
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-gray-900">
                    Secure Your Recovery Phrase
                  </h4>
                  <p className="text-xs text-gray-600 mt-1">
                    Keep your 12-word recovery phrase in a safe place
                  </p>
                </div>
              </div>
              <div className="flex items-start">
                <div className="flex items-center justify-center h-8 w-8 bg-blue-100 rounded-lg mr-3 flex-shrink-0">
                  <CloudArrowUpIcon className="h-4 w-4 text-blue-600" />
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-gray-900">
                    Start Uploading Files
                  </h4>
                  <p className="text-xs text-gray-600 mt-1">
                    Your files are encrypted before leaving your device
                  </p>
                </div>
              </div>
              <div className="flex items-start">
                <div className="flex items-center justify-center h-8 w-8 bg-blue-100 rounded-lg mr-3 flex-shrink-0">
                  <DocumentCheckIcon className="h-4 w-4 text-blue-600" />
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-gray-900">
                    Organize Your Documents
                  </h4>
                  <p className="text-xs text-gray-600 mt-1">
                    Create collections to organize your files
                  </p>
                </div>
              </div>
              <div className="flex items-start">
                <div className="flex items-center justify-center h-8 w-8 bg-blue-100 rounded-lg mr-3 flex-shrink-0">
                  <ShieldCheckIcon className="h-4 w-4 text-blue-600" />
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-gray-900">
                    Share Securely
                  </h4>
                  <p className="text-xs text-gray-600 mt-1">
                    Share files with end-to-end encryption
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Security Reminder */}
          <div className="bg-amber-50 border border-amber-200 rounded-xl p-6 animate-fade-in-up-delay-3">
            <div className="flex items-start">
              <ShieldCheckIcon className="h-6 w-6 text-amber-600 mr-3 flex-shrink-0 mt-1" />
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-amber-800 mb-2">
                  Security Reminder
                </h3>
                <p className="text-sm text-amber-700">
                  Your account uses end-to-end encryption. Your recovery phrase
                  is the <strong>only way</strong> to recover your data if you
                  forget your password. Make sure it's stored securely in a
                  physical location!
                </p>
              </div>
            </div>
          </div>

          {/* Auto-redirect Notice */}
          <div className="text-center p-4 bg-gray-50 rounded-lg animate-fade-in-up-delay-3">
            <p className="text-sm text-gray-600">
              Redirecting to login page in{" "}
              <span className="font-semibold text-gray-900">{countdown}</span>{" "}
              seconds...
            </p>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 animate-fade-in-up-delay-3">
            <button
              type="button"
              onClick={handleRegisterAnother}
              className="flex-1 inline-flex items-center justify-center py-3 px-4 border border-gray-300 text-base font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 transition-all duration-200"
            >
              Register Another Account
            </button>

            <button
              type="button"
              onClick={handleGoToLogin}
              className="group flex-1 inline-flex items-center justify-center py-3 px-4 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
            >
              <LockClosedIcon className="mr-2 h-5 w-5" />
              Sign In Now
              <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
            </button>
          </div>

          {/* Trust Badges */}
          <div className="flex items-center justify-center space-x-6 text-xs text-gray-500 animate-fade-in-up-delay-3">
            <div className="flex items-center space-x-1">
              <LockClosedIcon className="h-4 w-4 text-green-600" />
              <span>ChaCha20-Poly1305</span>
            </div>
            <div className="flex items-center space-x-1">
              <ServerIcon className="h-4 w-4 text-blue-600" />
              <span>Canadian Hosted</span>
            </div>
            <div className="flex items-center space-x-1">
              <EyeSlashIcon className="h-4 w-4 text-purple-600" />
              <span>Zero Knowledge</span>
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-100 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center text-sm text-gray-500">
            <p className="flex items-center justify-center">
              &copy; 2025 MapleFile Inc. All rights reserved. Made with
              <HeartIcon className="h-4 w-4 mx-1 text-red-500" />
              in Canada
            </p>
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

        @keyframes bounce-once {
          0%,
          100% {
            transform: translateY(0);
          }
          50% {
            transform: translateY(-20px);
          }
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

        .animate-bounce-once {
          animation: bounce-once 0.8s ease-out;
        }
      `}</style>
    </div>
  );
};

export default VerifySuccess;
