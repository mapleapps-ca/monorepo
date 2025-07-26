// File: src/pages/Anonymous/Login/SessionExpired.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useLocation, Link } from "react-router";
import {
  ClockIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  ArrowRightIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";

const SessionExpired = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [countdown, setCountdown] = useState(30);

  // Get the reason and message from location state
  const { reason, message, from } = location.state || {};

  // Determine the appropriate message and icon based on reason
  const getSessionInfo = () => {
    switch (reason) {
      case "inactivity_timeout":
        return {
          title: "Session Expired",
          subtitle: "Your session expired due to inactivity",
          description:
            "For your security, we automatically log you out after 60 minutes of inactivity. This helps protect your encrypted files from unauthorized access.",
          icon: ClockIcon,
          iconColor: "text-amber-600",
          iconBg: "bg-amber-100",
        };
      case "manual_clear":
        return {
          title: "Session Cleared",
          subtitle: "Your session was manually cleared",
          description:
            "Your session has been cleared. This might have happened if you logged out from another tab or if there was a security-related action.",
          icon: ShieldCheckIcon,
          iconColor: "text-blue-600",
          iconBg: "bg-blue-100",
        };
      default:
        return {
          title: "Session Expired",
          subtitle: "Your session has expired",
          description:
            "For your security, your session has expired. Please sign in again to continue accessing your encrypted files.",
          icon: LockClosedIcon,
          iconColor: "text-red-600",
          iconBg: "bg-red-100",
        };
    }
  };

  const sessionInfo = getSessionInfo();

  // Auto-redirect countdown
  useEffect(() => {
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          navigate("/login", {
            state: {
              from: from,
              autoRedirect: true,
              reason: "session_expired",
            },
          });
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [navigate, from]);

  const handleSignInNow = () => {
    navigate("/login", {
      state: {
        from: from,
        reason: "session_expired",
      },
    });
  };

  const handleGoHome = () => {
    navigate("/");
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
          {/* Header */}
          <div className="text-center animate-fade-in-up">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div
                  className={`flex items-center justify-center h-16 w-16 ${sessionInfo.iconBg} rounded-2xl shadow-lg`}
                >
                  <sessionInfo.icon
                    className={`h-8 w-8 ${sessionInfo.iconColor}`}
                  />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20"></div>
              </div>
            </div>
            <h2 className="text-3xl font-black text-gray-900 mb-2">
              {sessionInfo.title}
            </h2>
            <p className="text-gray-600 mb-4">{sessionInfo.subtitle}</p>
          </div>

          {/* Session Info Card */}
          <div className="bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 animate-fade-in-up-delay">
            {/* Description */}
            <div className="mb-6 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
              <div className="flex items-start">
                <InformationCircleIcon className="h-5 w-5 text-blue-500 mr-3 flex-shrink-0 mt-0.5" />
                <div>
                  <h3 className="text-sm font-semibold text-blue-900 mb-2">
                    What happened?
                  </h3>
                  <p className="text-sm text-blue-800">
                    {sessionInfo.description}
                  </p>
                </div>
              </div>
            </div>

            {/* Security Info */}
            <div className="mb-6 p-4 bg-gradient-to-r from-green-50 to-emerald-50 rounded-lg border border-green-100">
              <div className="flex items-start">
                <CheckCircleIcon className="h-5 w-5 text-green-500 mr-3 flex-shrink-0 mt-0.5" />
                <div>
                  <h3 className="text-sm font-semibold text-green-900 mb-2">
                    Your data is secure
                  </h3>
                  <ul className="text-sm text-green-800 space-y-1">
                    <li>• Your files remain encrypted and protected</li>
                    <li>• No one can access your data without your password</li>
                    <li>
                      • Session expiry is a security feature, not a problem
                    </li>
                  </ul>
                </div>
              </div>
            </div>

            {/* Auto-redirect notice */}
            <div className="mb-6 p-4 bg-gradient-to-r from-amber-50 to-yellow-50 rounded-lg border border-amber-100">
              <div className="flex items-center justify-center">
                <ClockIcon className="h-5 w-5 text-amber-500 mr-3" />
                <p className="text-sm text-amber-800">
                  Automatically redirecting to sign in page in{" "}
                  <span className="font-bold text-amber-900">{countdown}</span>{" "}
                  seconds
                </p>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="space-y-4">
              <button
                onClick={handleSignInNow}
                className="group w-full flex justify-center items-center py-3 px-4 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
              >
                Sign In Now
                <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
              </button>

              <button
                onClick={handleGoHome}
                className="w-full flex justify-center items-center py-3 px-4 border-2 border-gray-300 text-base font-semibold rounded-lg text-gray-700 bg-white hover:bg-gray-50 hover:border-gray-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
              >
                Go to Homepage
              </button>
            </div>

            {/* Alternative Actions */}
            <div className="mt-6 flex flex-col sm:flex-row justify-between items-center space-y-2 sm:space-y-0 text-sm">
              <Link
                to="/register"
                className="text-red-600 hover:text-red-700 font-medium hover:underline transition-colors duration-200"
              >
                Create new account
              </Link>
              <Link
                to="/recovery"
                className="text-gray-600 hover:text-gray-700 font-medium hover:underline transition-colors duration-200"
              >
                Forgot your password?
              </Link>
            </div>
          </div>

          {/* Additional Info */}
          <div className="text-center text-sm text-gray-500 animate-fade-in-up-delay-2">
            <p>Session timeout: 60 minutes of inactivity</p>
            <p className="mt-1">This helps keep your encrypted files secure</p>
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

export default SessionExpired;
