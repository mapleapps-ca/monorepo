// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Register/RecoveryCode.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router";
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
  PrinterIcon,
  DocumentDuplicateIcon,
  ExclamationCircleIcon,
  BookmarkIcon,
  ServerIcon,
  EyeSlashIcon,
} from "@heroicons/react/24/outline";

const RecoveryCode = () => {
  const navigate = useNavigate();
  const [recoveryMnemonic, setRecoveryMnemonic] = useState("");
  const [email, setEmail] = useState("");
  const [savedRecoveryCode, setSavedRecoveryCode] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    // Get registration result from sessionStorage
    const registrationResult = sessionStorage.getItem("registrationResult");
    const registeredEmail = sessionStorage.getItem("registeredEmail");

    if (!registrationResult || !registeredEmail) {
      // Redirect back to registration if no data found
      navigate("/register");
      return;
    }

    try {
      const result = JSON.parse(registrationResult);
      console.log("Full registration result:", result);

      // Get the recovery mnemonic that was generated on the frontend
      const mnemonic = result.recoveryMnemonic;

      if (!mnemonic) {
        console.error("No recovery mnemonic found in registration result");
        navigate("/register");
        return;
      }

      setRecoveryMnemonic(mnemonic);
      setEmail(registeredEmail);
    } catch (error) {
      console.error("Error parsing registration result:", error);
      navigate("/register");
    }
  }, [navigate]);

  const handleContinue = () => {
    if (!savedRecoveryCode) {
      alert(
        "Please confirm that you have saved your recovery code before continuing!",
      );
      return;
    }

    // Navigate to email verification
    navigate("/register/verify-email");
  };

  const handleBackToRegistration = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    navigate("/register");
  };

  const handleCopyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(recoveryMnemonic);
      setCopied(true);
      setTimeout(() => setCopied(false), 3000);
    } catch (error) {
      console.error("Failed to copy to clipboard:", error);
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = recoveryMnemonic;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand("copy");
      document.body.removeChild(textArea);
      setCopied(true);
      setTimeout(() => setCopied(false), 3000);
    }
  };

  const handlePrint = () => {
    const printWindow = window.open("", "_blank");
    printWindow.document.write(`
      <html>
        <head>
          <title>MapleFile Recovery Phrase</title>
          <style>
            body {
              font-family: Arial, sans-serif;
              padding: 20px;
              line-height: 1.6;
            }
            .header {
              text-align: center;
              margin-bottom: 30px;
            }
            .warning {
              background: #fff3cd;
              border: 1px solid #ffeaa7;
              padding: 15px;
              margin: 20px 0;
              border-radius: 4px;
            }
            .mnemonic {
              background: #f8f9fa;
              border: 2px solid #dee2e6;
              padding: 20px;
              margin: 20px 0;
              border-radius: 4px;
              font-family: monospace;
              font-size: 16px;
              text-align: center;
              line-height: 2;
            }
            .word {
              display: inline-block;
              margin: 5px;
              padding: 5px 10px;
              background: white;
              border: 1px solid #ccc;
              border-radius: 3px;
            }
            .footer {
              margin-top: 30px;
              font-size: 12px;
              color: #666;
            }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>MapleFile Recovery Phrase</h1>
            <p><strong>Account:</strong> ${email}</p>
            <p><strong>Generated:</strong> ${new Date().toLocaleString()}</p>
          </div>

          <div class="warning">
            <h3>⚠️ IMPORTANT SECURITY NOTICE</h3>
            <p>This recovery phrase is the ONLY way to recover your account if you forget your password. Keep it safe and never share it with anyone.</p>
          </div>

          <div class="mnemonic">
            ${recoveryMnemonic
              .split(" ")
              .map(
                (word, index) =>
                  `<span class="word">${index + 1}. ${word}</span>`,
              )
              .join("")}
          </div>

          <div class="footer">
            <p><strong>Security Tips:</strong></p>
            <ul>
              <li>Store this phrase in a secure location (safe, safety deposit box)</li>
              <li>Consider making multiple copies and storing them separately</li>
              <li>Never store this digitally (computer files, cloud storage, photos)</li>
              <li>Never share this phrase with anyone, including MapleFile support</li>
              <li>Write clearly and double-check each word</li>
            </ul>
          </div>
        </body>
      </html>
    `);
    printWindow.document.close();
    printWindow.print();
  };

  if (!recoveryMnemonic) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Loading...</h2>
          <p className="text-gray-600">Loading your recovery phrase...</p>
        </div>
      </div>
    );
  }

  // Split mnemonic into words for better display
  const mnemonicWords = recoveryMnemonic.split(" ");

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
                  Register
                </span>
              </div>
              <div className="w-12 h-0.5 bg-green-500"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gradient-to-r from-red-800 to-red-900 rounded-full text-white text-sm font-bold">
                  2
                </div>
                <span className="ml-2 text-sm font-semibold text-gray-900">
                  Recovery
                </span>
              </div>
              <div className="w-12 h-0.5 bg-gray-300"></div>
              <div className="flex items-center">
                <div className="flex items-center justify-center w-8 h-8 bg-gray-300 rounded-full text-gray-500 text-sm font-bold">
                  3
                </div>
                <span className="ml-2 text-sm text-gray-500">Verify</span>
              </div>
            </div>
          </div>

          {/* Header */}
          <div className="text-center animate-fade-in-up">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="flex items-center justify-center h-16 w-16 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-lg animate-pulse">
                  <KeyIcon className="h-8 w-8 text-white" />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
              </div>
            </div>
            <h2 className="text-3xl font-black text-gray-900 mb-2">
              Save Your Recovery Phrase
            </h2>
            <p className="text-gray-600 mb-2">
              Write down these 12 words in the exact order shown
            </p>
            <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
              <ExclamationCircleIcon className="h-4 w-4 text-amber-600" />
              <span className="text-amber-700 font-semibold">
                This is your ONLY way to recover your account
              </span>
            </div>
          </div>

          {/* Warning Card */}
          <div className="bg-amber-50 border-2 border-amber-200 rounded-xl p-6 animate-fade-in-up-delay">
            <div className="flex items-start">
              <ExclamationTriangleIcon className="h-6 w-6 text-amber-600 mr-3 flex-shrink-0 mt-1" />
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-amber-800 mb-2">
                  Critical Security Notice
                </h3>
                <ul className="text-sm text-amber-700 space-y-1">
                  <li className="flex items-start">
                    <span className="text-amber-600 mr-2">•</span>
                    This recovery phrase is the <strong>only way</strong> to
                    recover your account
                  </li>
                  <li className="flex items-start">
                    <span className="text-amber-600 mr-2">•</span>
                    We cannot recover your account without it
                  </li>
                  <li className="flex items-start">
                    <span className="text-amber-600 mr-2">•</span>
                    Never share this phrase with anyone, including MapleFile
                    support
                  </li>
                  <li className="flex items-start">
                    <span className="text-amber-600 mr-2">•</span>
                    Store it in a secure physical location
                  </li>
                </ul>
              </div>
            </div>
          </div>

          {/* Recovery Phrase Card */}
          <div className="bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 animate-fade-in-up-delay-2">
            <h3 className="text-lg font-semibold text-gray-900 mb-6 flex items-center justify-center">
              <BookmarkIcon className="h-5 w-5 mr-2 text-red-600" />
              Your 12-Word Recovery Phrase
            </h3>

            <div className="grid grid-cols-3 gap-4 mb-8">
              {mnemonicWords.map((word, index) => (
                <div
                  key={index}
                  className="bg-gray-50 border border-gray-200 rounded-lg p-3 text-center transform hover:scale-105 transition-transform duration-200"
                >
                  <span className="text-xs text-gray-500 block mb-1">
                    {index + 1}
                  </span>
                  <span className="text-base font-mono font-semibold text-gray-900">
                    {word}
                  </span>
                </div>
              ))}
            </div>

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-3 mb-8">
              <button
                type="button"
                onClick={handleCopyToClipboard}
                className={`flex-1 inline-flex items-center justify-center px-4 py-3 border rounded-lg font-medium transition-all duration-200 ${
                  copied
                    ? "bg-green-50 border-green-300 text-green-700"
                    : "bg-white border-gray-300 text-gray-700 hover:bg-gray-50 hover:border-gray-400"
                }`}
              >
                {copied ? (
                  <>
                    <CheckIcon className="h-5 w-5 mr-2" />
                    Copied to Clipboard!
                  </>
                ) : (
                  <>
                    <ClipboardDocumentIcon className="h-5 w-5 mr-2" />
                    Copy to Clipboard
                  </>
                )}
              </button>
              <button
                type="button"
                onClick={handlePrint}
                className="flex-1 inline-flex items-center justify-center px-4 py-3 bg-white border border-gray-300 rounded-lg text-gray-700 font-medium hover:bg-gray-50 hover:border-gray-400 transition-all duration-200"
              >
                <PrinterIcon className="h-5 w-5 mr-2" />
                Print Recovery Phrase
              </button>
            </div>

            {/* Confirmation Checkbox */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex items-start">
                <div className="flex items-center h-5">
                  <input
                    type="checkbox"
                    id="saved_recovery_code"
                    checked={savedRecoveryCode}
                    onChange={(e) => setSavedRecoveryCode(e.target.checked)}
                    className="h-4 w-4 text-red-800 border-gray-300 rounded focus:ring-red-500"
                  />
                </div>
                <div className="ml-3">
                  <label
                    htmlFor="saved_recovery_code"
                    className="text-sm font-semibold text-blue-900"
                  >
                    I have written down my recovery phrase and stored it
                    securely
                  </label>
                  <p className="text-xs text-blue-700 mt-1">
                    Please confirm you've saved your recovery phrase before
                    continuing
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Security Tips */}
          <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-6 animate-fade-in-up-delay-3">
            <h3 className="text-sm font-semibold text-red-900 mb-4 flex items-center">
              <ExclamationCircleIcon className="h-4 w-4 mr-2" />
              Security Best Practices
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <h4 className="font-semibold text-green-700 mb-2">✅ DO:</h4>
                <ul className="space-y-1 text-gray-600">
                  <li className="flex items-start">
                    <span className="text-green-500 mr-2">•</span>
                    Write it on paper with a pen
                  </li>
                  <li className="flex items-start">
                    <span className="text-green-500 mr-2">•</span>
                    Store in a safe or safety deposit box
                  </li>
                  <li className="flex items-start">
                    <span className="text-green-500 mr-2">•</span>
                    Make multiple physical copies
                  </li>
                  <li className="flex items-start">
                    <span className="text-green-500 mr-2">•</span>
                    Store copies in separate locations
                  </li>
                </ul>
              </div>
              <div>
                <h4 className="font-semibold text-red-700 mb-2">❌ DON'T:</h4>
                <ul className="space-y-1 text-gray-600">
                  <li className="flex items-start">
                    <span className="text-red-500 mr-2">•</span>
                    Take a screenshot or photo
                  </li>
                  <li className="flex items-start">
                    <span className="text-red-500 mr-2">•</span>
                    Save in a computer file
                  </li>
                  <li className="flex items-start">
                    <span className="text-red-500 mr-2">•</span>
                    Store in cloud storage
                  </li>
                  <li className="flex items-start">
                    <span className="text-red-500 mr-2">•</span>
                    Share via email or messages
                  </li>
                </ul>
              </div>
            </div>
          </div>

          {/* Next Steps */}
          <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100 p-6 animate-fade-in-up-delay-3">
            <h3 className="text-sm font-semibold text-blue-900 mb-3 flex items-center">
              <InformationCircleIcon className="h-4 w-4 mr-2" />
              What happens next?
            </h3>
            <ol className="text-sm text-blue-800 space-y-2">
              <li className="flex items-start">
                <span className="font-semibold mr-2">1.</span>
                Write down your recovery phrase on paper
              </li>
              <li className="flex items-start">
                <span className="font-semibold mr-2">2.</span>
                Check your email ({email}) for a verification code
              </li>
              <li className="flex items-start">
                <span className="font-semibold mr-2">3.</span>
                Enter the code to complete registration
              </li>
              <li className="flex items-start">
                <span className="font-semibold mr-2">4.</span>
                Start using MapleFile with end-to-end encryption
              </li>
            </ol>
          </div>

          {/* Action Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 animate-fade-in-up-delay-3">
            <button
              type="button"
              onClick={handleBackToRegistration}
              className="flex-1 inline-flex items-center justify-center py-3 px-4 border border-gray-300 text-base font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 transition-all duration-200"
            >
              <ArrowLeftIcon className="mr-2 h-4 w-4" />
              Back to Registration
            </button>

            <button
              type="button"
              onClick={handleContinue}
              disabled={!savedRecoveryCode}
              className="group flex-1 inline-flex items-center justify-center py-3 px-4 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
            >
              Continue to Email Verification
              <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
            </button>
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

        .animate-fade-in-up-delay-3 {
          animation: fade-in-up 0.6s ease-out 0.6s both;
        }
      `}</style>
    </div>
  );
};

export default RecoveryCode;
