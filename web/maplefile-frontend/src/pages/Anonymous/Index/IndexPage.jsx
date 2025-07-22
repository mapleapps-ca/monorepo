// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Index/IndexPage.jsx
import { useState, useEffect } from "react";
import { Link } from "react-router";
import {
  ArrowRightIcon,
  PlayIcon,
  ShieldCheckIcon,
  LockClosedIcon,
  ServerIcon,
  KeyIcon,
  CloudArrowUpIcon,
  EyeSlashIcon,
  BoltIcon,
  CheckIcon,
  SparklesIcon,
  GlobeAltIcon,
  HeartIcon,
} from "@heroicons/react/24/outline";

function IndexPage() {
  const [authMessage, setAuthMessage] = useState("");

  useEffect(() => {
    // Check for auth redirect message
    const message = sessionStorage.getItem("auth_redirect_message");
    if (message) {
      setAuthMessage(message);
      sessionStorage.removeItem("auth_redirect_message");

      // Clear message after 10 seconds
      const timer = setTimeout(() => {
        setAuthMessage("");
      }, 10000);

      return () => clearTimeout(timer);
    }
  }, []);

  const securityFeatures = [
    {
      icon: LockClosedIcon,
      title: "End-to-End Encryption",
      description:
        "ChaCha20-Poly1305 encryption with X25519 key exchange ensures your files are always protected.",
    },
    {
      icon: KeyIcon,
      title: "Client-Side Cryptography",
      description:
        "All encryption happens in your browser. Your keys never leave your device.",
    },
    {
      icon: EyeSlashIcon,
      title: "Zero-Knowledge Architecture",
      description:
        "We can't see your files even if we wanted to. True privacy by design.",
    },
    {
      icon: ServerIcon,
      title: "Canadian Data Residency",
      description:
        "Your encrypted data stays in Canada, protected by strong privacy laws.",
    },
  ];

  const technicalFeatures = [
    {
      icon: BoltIcon,
      feature: "ChaCha20-Poly1305 Encryption",
    },
    {
      icon: KeyIcon,
      feature: "X25519 Key Exchange",
    },
    {
      icon: ShieldCheckIcon,
      feature: "Encrypted Auth Tokens",
    },
    {
      icon: CloudArrowUpIcon,
      feature: "Secure File Sharing",
    },
    {
      icon: SparklesIcon,
      feature: "Auto Token Refresh",
    },
    {
      icon: LockClosedIcon,
      feature: "Browser-Only Crypto",
    },
  ];

  return (
    <div className="min-h-screen bg-white overflow-hidden">
      {/* Show auth message if present */}
      {authMessage && (
        <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 fixed top-0 left-0 right-0 z-50">
          <div className="flex">
            <div className="ml-3">
              <p className="text-sm text-yellow-700">
                <strong>Authentication Required:</strong> {authMessage}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <nav className="relative z-50 bg-white/95 backdrop-blur-sm border-b border-gray-100">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center group">
              <div className="flex items-center justify-center h-10 w-10 bg-gradient-to-br from-red-800 to-red-900 rounded-lg mr-3 group-hover:scale-105 transition-transform duration-200">
                <LockClosedIcon className="h-6 w-6 text-white animate-pulse" />
              </div>
              <span className="text-2xl font-bold bg-gradient-to-r from-gray-900 to-red-800 bg-clip-text text-transparent">
                MapleFile
              </span>
            </div>
            <div className="flex items-center space-x-6">
              <Link
                to="/login"
                className="text-base font-medium text-gray-700 hover:text-red-800 transition-colors duration-200"
              >
                Sign in
              </Link>
              <Link
                to="/register"
                className="group inline-flex items-center px-6 py-2.5 border border-transparent text-base font-medium rounded-lg shadow-sm text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 transform hover:scale-105 transition-all duration-200"
              >
                Get Started Free
                <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <div className="relative bg-gradient-to-br from-gray-50 via-white to-red-50 overflow-hidden">
        <div className="absolute inset-0 bg-grid-pattern opacity-5"></div>
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-20 pb-32">
          <div className="text-center">
            <div className="flex justify-center mb-8">
              <div className="relative">
                <div className="flex items-center justify-center h-20 w-20 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-xl animate-bounce">
                  <LockClosedIcon className="h-12 w-12 text-white" />
                </div>
                <div className="absolute -inset-2 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
              </div>
            </div>

            <h1 className="text-5xl md:text-7xl tracking-tight font-black text-gray-900 mb-6">
              <span className="block animate-fade-in-up">
                Secure File Storage
              </span>
              <span className="block bg-gradient-to-r from-red-800 via-red-700 to-red-900 bg-clip-text text-transparent animate-fade-in-up-delay">
                with End-to-End Encryption
              </span>
            </h1>

            <p className="mt-6 max-w-3xl mx-auto text-xl text-gray-600 leading-relaxed animate-fade-in-up-delay-2">
              Store, sync, and share your files with
              <span className="font-semibold text-red-800">
                {" "}
                military-grade encryption
              </span>
              . Your data is protected by ChaCha20-Poly1305 encryption and stays
              in Canada under strong privacy laws.
            </p>

            <div className="mt-10 grid grid-cols-1 sm:grid-cols-2 gap-3 justify-center items-center animate-fade-in-up-delay-3 max-w-md mx-auto">
              <Link
                to="/register"
                className="group inline-flex items-center justify-center px-6 py-3 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 shadow-lg hover:shadow-xl transform hover:scale-105 transition-all duration-300"
              >
                <PlayIcon className="mr-2 h-4 w-4 group-hover:scale-110 transition-transform duration-200" />
                Start Free Trial
                <ArrowRightIcon className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform duration-200" />
              </Link>
              <Link
                to="/login"
                className="inline-flex items-center justify-center px-6 py-3 border-2 border-red-800 text-base font-semibold rounded-lg text-red-800 bg-white hover:bg-red-50 shadow-lg hover:shadow-xl transform hover:scale-105 transition-all duration-300"
              >
                Sign In
              </Link>
            </div>

            <div className="mt-8 text-sm text-gray-500">
              Already have an account?{" "}
              <Link
                to="/recovery"
                className="text-red-800 hover:text-red-900 font-medium"
              >
                Recover your account
              </Link>
            </div>
          </div>
        </div>
      </div>

      {/* Stats Section */}
      <div className="bg-gradient-to-r from-red-800 to-red-900 py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
            <div className="text-white">
              <div className="text-4xl font-black mb-2">256-bit</div>
              <div className="text-red-100 font-medium">
                Encryption Strength
              </div>
            </div>
            <div className="text-white">
              <div className="text-4xl font-black mb-2">100%</div>
              <div className="text-red-100 font-medium">Canadian Hosted</div>
            </div>
            <div className="text-white">
              <div className="text-4xl font-black mb-2">Zero</div>
              <div className="text-red-100 font-medium">Knowledge Access</div>
            </div>
            <div className="text-white">
              <div className="text-4xl font-black mb-2">99.9%</div>
              <div className="text-red-100 font-medium">Uptime SLA</div>
            </div>
          </div>
        </div>
      </div>

      {/* Security Features Section */}
      <div className="bg-gradient-to-br from-gray-50 to-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-black text-gray-900 mb-4">
              Security By Design
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Your files are protected with the same encryption used by
              intelligence agencies
            </p>
          </div>

          <div className="grid grid-cols-2 gap-12 max-w-4xl mx-auto">
            {securityFeatures.map((feature, index) => (
              <div key={index} className="group relative">
                <div
                  className="p-8 rounded-xl transform hover:scale-105 transition-all duration-300 h-full"
                  style={{
                    backgroundColor: "#8a1622",
                    boxShadow:
                      "0 25px 50px -12px rgba(0, 0, 0, 0.4), 0 25px 25px -12px rgba(0, 0, 0, 0.3)",
                    filter: "drop-shadow(0 20px 25px rgba(0, 0, 0, 0.15))",
                  }}
                >
                  <div className="text-center">
                    <div
                      className="inline-flex items-center justify-center h-16 w-16 rounded-xl mb-6 group-hover:scale-110 transition-transform duration-200"
                      style={{ backgroundColor: "#f6f6f6" }}
                    >
                      <feature.icon
                        className="h-8 w-8"
                        style={{ color: "#8a1622" }}
                      />
                    </div>
                    <h3
                      className="text-xl font-bold mb-4"
                      style={{ color: "#f6f6f6" }}
                    >
                      {feature.title}
                    </h3>
                    <p className="leading-relaxed" style={{ color: "#f6f6f6" }}>
                      {feature.description}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Technical Features Section */}
      <div className="bg-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-black text-gray-900 mb-4">
              Technical Excellence
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Built with cutting-edge cryptography and security best practices
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 gap-8 max-w-4xl mx-auto">
            {technicalFeatures.map((item, index) => (
              <div key={index} className="group">
                <div
                  className="p-6 rounded-xl transform hover:scale-105 transition-all duration-300 text-center h-full"
                  style={{
                    backgroundColor: "#f6f6f6",
                    boxShadow:
                      "0 25px 50px -12px rgba(0, 0, 0, 0.4), 0 25px 25px -12px rgba(0, 0, 0, 0.3)",
                    filter: "drop-shadow(0 20px 25px rgba(0, 0, 0, 0.15))",
                  }}
                >
                  <div
                    className="inline-flex items-center justify-center h-12 w-12 rounded-lg mb-4 group-hover:scale-110 transition-transform duration-200"
                    style={{ backgroundColor: "#8a1622" }}
                  >
                    <item.icon
                      className="h-6 w-6"
                      style={{ color: "#f6f6f6" }}
                    />
                  </div>
                  <h3
                    className="text-sm font-bold"
                    style={{ color: "#222222" }}
                  >
                    {item.feature}
                  </h3>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Pricing Section */}
      <div className="bg-gradient-to-br from-gray-50 to-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-black text-gray-900 mb-4">
              Simple, Transparent Pricing
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Choose the plan that fits your storage needs
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
            {/* Free Plan */}
            <div className="relative">
              <div className="bg-white border-2 border-gray-200 rounded-3xl p-8 shadow-lg h-full">
                <div className="text-center">
                  <h3 className="text-2xl font-black text-gray-900 mb-2">
                    Personal
                  </h3>
                  <p className="text-gray-600 mb-8">
                    Perfect for getting started
                  </p>

                  <div className="mb-8">
                    <span className="text-5xl font-black text-gray-900">
                      Free
                    </span>
                    <span className="text-lg text-gray-600 ml-2">forever</span>
                  </div>

                  <div className="grid grid-cols-1 gap-4 mb-8 text-left">
                    {[
                      "10 GB encrypted storage",
                      "End-to-end encryption",
                      "Canadian data residency",
                      "Basic file sharing",
                      "Mobile & desktop sync",
                      "Community support",
                    ].map((feature, idx) => (
                      <div key={idx} className="flex items-start">
                        <CheckIcon className="h-5 w-5 text-green-500 mr-3 mt-0.5 flex-shrink-0" />
                        <span className="text-gray-700 text-sm font-medium">
                          {feature}
                        </span>
                      </div>
                    ))}
                  </div>

                  <Link
                    to="/register"
                    className="group w-full inline-flex items-center justify-center px-6 py-3 border-2 border-gray-300 text-gray-700 text-base font-bold rounded-lg hover:border-gray-400 hover:bg-gray-50 shadow-lg hover:shadow-xl transform hover:scale-105 transition-all duration-300"
                  >
                    Get Started Free
                  </Link>
                </div>
              </div>
            </div>

            {/* Pro Plan */}
            <div className="relative">
              <div className="absolute -inset-4 bg-gradient-to-r from-red-800 to-red-900 rounded-3xl blur opacity-20"></div>
              <div className="relative bg-gradient-to-br from-white to-gray-50 border-2 border-red-800 rounded-3xl p-10 shadow-2xl">
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2 bg-gradient-to-r from-red-800 to-red-900 text-white text-sm font-bold py-2 px-6 rounded-full shadow-lg">
                  MOST POPULAR
                </div>

                <div className="text-center">
                  <h3 className="text-3xl font-black text-gray-900 mb-2">
                    Pro
                  </h3>
                  <p className="text-gray-600 mb-8">
                    For power users and professionals
                  </p>

                  <div className="mb-8">
                    <span className="text-6xl font-black text-gray-900">
                      $9.99
                    </span>
                    <span className="text-xl text-gray-600 ml-2">
                      CAD/month
                    </span>
                  </div>

                  <div className="grid grid-cols-1 gap-4 mb-8 text-left">
                    {[
                      "1 TB encrypted storage",
                      "All Personal features",
                      "Advanced file sharing",
                      "Version history",
                      "Priority support",
                      "Mobile & desktop apps",
                    ].map((feature, idx) => (
                      <div key={idx} className="flex items-start">
                        <CheckIcon className="h-6 w-6 text-green-500 mr-3 mt-0.5 flex-shrink-0" />
                        <span className="text-gray-700 font-medium">
                          {feature}
                        </span>
                      </div>
                    ))}
                  </div>

                  <Link
                    to="/register"
                    className="group w-full inline-flex items-center justify-center px-8 py-4 bg-gradient-to-r from-red-800 to-red-900 text-white text-lg font-bold rounded-xl hover:from-red-900 hover:to-red-950 shadow-xl hover:shadow-2xl transform hover:scale-105 transition-all duration-300"
                  >
                    Start Free Trial
                    <span className="ml-2 transform group-hover:translate-x-1 transition-transform duration-200">
                      →
                    </span>
                  </Link>
                  <p className="mt-4 text-sm text-gray-500">
                    No credit card required
                  </p>
                </div>
              </div>
            </div>

            {/* Team Plan */}
            <div className="relative">
              <div className="bg-white border-2 border-gray-200 rounded-3xl p-8 shadow-lg h-full">
                <div className="text-center">
                  <h3 className="text-2xl font-black text-gray-900 mb-2">
                    Team
                  </h3>
                  <p className="text-gray-600 mb-8">
                    For teams and organizations
                  </p>

                  <div className="mb-8">
                    <span className="text-5xl font-black text-gray-900">
                      $49.99
                    </span>
                    <span className="text-lg text-gray-600 ml-2">
                      CAD/month
                    </span>
                    <p className="text-sm text-gray-500 mt-1">Up to 10 users</p>
                  </div>

                  <div className="grid grid-cols-1 gap-4 mb-8 text-left">
                    {[
                      "10 TB shared storage",
                      "All Pro features",
                      "Team collaboration",
                      "Admin controls",
                      "Advanced sharing",
                      "Dedicated support",
                    ].map((feature, idx) => (
                      <div key={idx} className="flex items-start">
                        <CheckIcon className="h-5 w-5 text-green-500 mr-3 mt-0.5 flex-shrink-0" />
                        <span className="text-gray-700 text-sm font-medium">
                          {feature}
                        </span>
                      </div>
                    ))}
                  </div>

                  <Link
                    to="/register"
                    className="group w-full inline-flex items-center justify-center px-6 py-3 border-2 text-base font-bold rounded-lg shadow-lg hover:shadow-xl transform hover:scale-105 transition-all duration-300"
                    style={{
                      borderColor: "#8a1622",
                      color: "#8a1622",
                      backgroundColor: "white",
                    }}
                  >
                    Contact Sales
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Trust Section */}
      <div className="bg-gradient-to-br from-red-50 to-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-black text-gray-900 mb-4">
              Built with Canadian Values
            </h2>
          </div>

          <div className="bg-white rounded-3xl shadow-2xl border border-gray-100 overflow-hidden">
            <div className="grid md:grid-cols-2 gap-0">
              <div className="p-12">
                <h3 className="text-3xl font-black text-gray-900 mb-6">
                  Privacy You Can Trust
                </h3>
                <p className="text-gray-600 mb-6 text-lg leading-relaxed">
                  Your files are encrypted on your device before they ever
                  leave. We use the same cryptographic standards trusted by
                  governments and security professionals worldwide.
                </p>
                <p className="text-gray-600 mb-8 text-lg leading-relaxed">
                  <strong>Session Security:</strong> For maximum security,
                  you'll need to re-enter your password if you refresh the page.
                  This ensures your encryption keys are never permanently
                  stored.
                </p>
                <div className="space-y-4">
                  {[
                    {
                      icon: "🍁",
                      text: "100% Canadian infrastructure",
                    },
                    {
                      icon: ShieldCheckIcon,
                      text: "PIPEDA compliant data handling",
                    },
                    {
                      icon: LockClosedIcon,
                      text: "Zero-knowledge architecture",
                    },
                  ].map((item, idx) => (
                    <div key={idx} className="flex items-center group">
                      {typeof item.icon === "string" ? (
                        <div className="flex items-center justify-center h-8 w-8 bg-red-100 rounded-lg mr-4 group-hover:scale-110 transition-transform duration-200">
                          <span className="text-lg">{item.icon}</span>
                        </div>
                      ) : (
                        <div className="flex items-center justify-center h-8 w-8 bg-red-100 rounded-lg mr-4 group-hover:scale-110 transition-transform duration-200">
                          <item.icon className="h-5 w-5 text-red-800" />
                        </div>
                      )}
                      <span className="font-semibold text-gray-900">
                        {item.text}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
              <div className="bg-gradient-to-br from-red-800 to-red-900 p-12 flex items-center justify-center">
                <div className="text-white text-center">
                  <div className="flex justify-center mb-6">
                    <LockClosedIcon className="h-20 w-20 animate-pulse" />
                  </div>
                  <p className="text-3xl font-black mb-2">Secure by Design</p>
                  <p className="text-xl text-red-100">Privacy First, Always</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8 mb-12">
            <div className="lg:col-span-2">
              <div className="flex items-center mb-6">
                <div className="flex items-center justify-center h-10 w-10 bg-gradient-to-br from-red-800 to-red-900 rounded-lg mr-3">
                  <LockClosedIcon className="h-6 w-6 text-white" />
                </div>
                <span className="text-2xl font-bold text-white">MapleFile</span>
              </div>
              <p className="text-gray-400 mb-6 max-w-md">
                Secure, encrypted file storage with Canadian privacy protection.
                Your data stays private with military-grade encryption.
              </p>
            </div>

            {[
              {
                title: "Product",
                links: ["Security", "Privacy", "Features", "Pricing"],
              },
              {
                title: "Support",
                links: [
                  "Help Center",
                  "Privacy Policy",
                  "Terms of Service",
                  "Contact",
                ],
              },
            ].map((section, idx) => (
              <div key={idx}>
                <h3 className="text-sm font-bold text-gray-300 tracking-wider uppercase mb-4">
                  {section.title}
                </h3>
                <ul className="space-y-3">
                  {section.links.map((link, linkIdx) => (
                    <li key={linkIdx}>
                      <a
                        href="#"
                        className="text-gray-400 hover:text-white transition-colors duration-200"
                      >
                        {link}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>

          <div className="border-t border-gray-800 pt-8">
            <div className="md:flex md:items-center md:justify-between">
              <p className="text-gray-400">
                &copy; 2025 MapleFile Inc. All rights reserved. Made with ❤️ in
                Canada.
              </p>
              <div className="mt-4 md:mt-0 flex space-x-6">
                <Link
                  to="/login"
                  className="text-gray-400 hover:text-white transition-colors duration-200"
                >
                  Sign In
                </Link>
                <Link
                  to="/register"
                  className="text-gray-400 hover:text-white transition-colors duration-200"
                >
                  Register
                </Link>
                <Link
                  to="/recovery"
                  className="text-gray-400 hover:text-white transition-colors duration-200"
                >
                  Recovery
                </Link>
              </div>
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

        .animate-fade-in-up-delay-3 {
          animation: fade-in-up 0.6s ease-out 0.6s both;
        }

        .bg-grid-pattern {
          background-image:
            linear-gradient(rgba(139, 22, 34, 0.1) 1px, transparent 1px),
            linear-gradient(90deg, rgba(139, 22, 34, 0.1) 1px, transparent 1px);
          background-size: 20px 20px;
        }
      `}</style>
    </div>
  );
}

export default IndexPage;
