// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Register/Register.jsx
import React, { useState } from "react";
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
  UserIcon,
  EnvelopeIcon,
  PhoneIcon,
  GlobeAltIcon,
  ClockIcon,
  KeyIcon,
  EyeIcon,
  EyeSlashIcon,
  SparklesIcon,
  ServerIcon,
  EyeSlashIcon as PrivacyIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";

const Register = () => {
  const navigate = useNavigate();
  const { authManager } = useServices();

  const [formData, setFormData] = useState({
    beta_access_code: "",
    first_name: "",
    last_name: "",
    email: "",
    phone: "",
    country: "Canada",
    timezone: "America/Toronto",
    agree_terms_of_service: false,
    agree_promotions: false,
    agree_to_tracking_across_third_party_apps_and_services: false,
    module: 1, // 1 = MapleFile, 2 = PaperCloud
    password: "", // This won't be sent to server, used for E2EE key generation
    confirmPassword: "", // Added for better UX
  });

  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState({});
  const [generalError, setGeneralError] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [keyGenerationProgress, setKeyGenerationProgress] = useState(false);

  const countries = [
    "Canada",
    "United States",
    "United Kingdom",
    "Australia",
    "Germany",
    "France",
    "Japan",
    "South Korea",
    "Brazil",
    "Mexico",
    "India",
    "Other",
  ];

  const timezones = [
    { value: "America/Toronto", label: "Toronto (EST/EDT)" },
    { value: "America/Vancouver", label: "Vancouver (PST/PDT)" },
    { value: "America/New_York", label: "New York (EST/EDT)" },
    { value: "America/Los_Angeles", label: "Los Angeles (PST/PDT)" },
    { value: "America/Chicago", label: "Chicago (CST/CDT)" },
    { value: "America/Denver", label: "Denver (MST/MDT)" },
    { value: "Europe/London", label: "London (GMT/BST)" },
    { value: "Europe/Paris", label: "Paris (CET/CEST)" },
    { value: "Europe/Berlin", label: "Berlin (CET/CEST)" },
    { value: "Asia/Tokyo", label: "Tokyo (JST)" },
    { value: "Asia/Seoul", label: "Seoul (KST)" },
    { value: "Australia/Sydney", label: "Sydney (AEST/AEDT)" },
    { value: "UTC", label: "UTC" },
  ];

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value,
    }));

    // Clear field-specific error when user starts typing
    if (errors[name]) {
      setErrors((prev) => ({
        ...prev,
        [name]: "",
      }));
    }

    if (generalError) {
      setGeneralError("");
    }
  };

  const validateForm = () => {
    const newErrors = {};

    if (!formData.beta_access_code.trim()) {
      newErrors.beta_access_code = "Beta access code is required";
    }

    if (!formData.first_name.trim()) {
      newErrors.first_name = "First name is required";
    }

    if (!formData.last_name.trim()) {
      newErrors.last_name = "Last name is required";
    }

    if (!formData.email.trim()) {
      newErrors.email = "Email is required";
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = "Please enter a valid email address";
    }

    if (!formData.phone.trim()) {
      newErrors.phone = "Phone number is required";
    }

    if (!formData.password.trim()) {
      newErrors.password = "Master password is required for encryption";
    } else if (formData.password.length < 8) {
      newErrors.password = "Password must be at least 8 characters long";
    }

    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = "Passwords do not match";
    }

    if (!formData.agree_terms_of_service) {
      newErrors.agree_terms_of_service =
        "You must agree to the terms of service";
    }

    return newErrors;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    const validationErrors = validateForm();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    setLoading(true);
    setErrors({});
    setGeneralError("");
    setKeyGenerationProgress(true);

    try {
      // Generate E2EE data using the crypto service via AuthManager
      console.log("[Register] Generating encryption data via AuthManager...");
      const e2eeData = await authManager.generateE2EEData(formData.password);
      console.log(
        "[Register] Encryption data generated successfully via AuthManager!",
      );

      // Extract the mnemonic and remove it from the data sent to server
      const { recoveryMnemonic, ...e2eeDataForServer } = e2eeData;

      // Prepare registration data
      const registrationData = {
        beta_access_code: formData.beta_access_code.trim(),
        first_name: formData.first_name.trim(),
        last_name: formData.last_name.trim(),
        email: formData.email.trim().toLowerCase(),
        phone: formData.phone.trim(),
        country: formData.country,
        timezone: formData.timezone,
        agree_terms_of_service: formData.agree_terms_of_service,
        agree_promotions: formData.agree_promotions,
        agree_to_tracking_across_third_party_apps_and_services:
          formData.agree_to_tracking_across_third_party_apps_and_services,
        module: formData.module,
        ...e2eeDataForServer,
      };

      setKeyGenerationProgress(false);
      console.log("[Register] Sending registration request via AuthManager...");
      const result = await authManager.registerUser(registrationData);

      console.log(
        "[Register] Registration successful via AuthManager:",
        result,
      );

      // Store data for next page - include the recovery mnemonic
      const registrationWithMnemonic = {
        ...result,
        recoveryMnemonic: recoveryMnemonic,
      };

      sessionStorage.setItem(
        "registrationResult",
        JSON.stringify(registrationWithMnemonic),
      );
      sessionStorage.setItem("registeredEmail", formData.email);

      // Navigate to recovery code page
      navigate("/register/recovery");
    } catch (error) {
      console.error("[Register] Registration failed via AuthManager:", error);

      // Parse error details if they exist
      try {
        const errorDetails = JSON.parse(error.message);
        setErrors(errorDetails);
      } catch {
        setGeneralError(
          error.message || "Registration failed. Please try again.",
        );
      }
    } finally {
      setLoading(false);
      setKeyGenerationProgress(false);
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
                to="/login"
                className="text-base font-medium text-gray-700 hover:text-red-800 transition-colors duration-200"
              >
                Already have an account?
              </Link>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <div className="flex-1 py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-2xl mx-auto">
          {/* Header */}
          <div className="text-center mb-8 animate-fade-in-up">
            <div className="flex justify-center mb-6">
              <div className="relative">
                <div className="flex items-center justify-center h-16 w-16 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-lg animate-pulse">
                  <UserIcon className="h-8 w-8 text-white" />
                </div>
                <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
              </div>
            </div>
            <h2 className="text-3xl font-black text-gray-900 mb-2">
              Create Your Secure Account
            </h2>
            <p className="text-gray-600 mb-2">
              Join MapleFile and protect your files with end-to-end encryption
            </p>
            <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
              <ShieldCheckIcon className="h-4 w-4 text-green-600" />
              <span>
                Zero-knowledge architecture with ChaCha20-Poly1305 encryption
              </span>
            </div>
          </div>

          {/* Form Card */}
          <div className="bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 animate-fade-in-up-delay">
            {/* General Error Message */}
            {generalError && (
              <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 animate-fade-in">
                <div className="flex items-center">
                  <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
                  <div>
                    <h3 className="text-sm font-semibold text-red-800">
                      Registration Error
                    </h3>
                    <p className="text-sm text-red-700 mt-1">{generalError}</p>
                  </div>
                </div>
              </div>
            )}

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Beta Access Code */}
              <div>
                <label
                  htmlFor="beta_access_code"
                  className="block text-sm font-semibold text-gray-700 mb-2"
                >
                  Beta Access Code
                </label>
                <div className="relative">
                  <input
                    type="text"
                    id="beta_access_code"
                    name="beta_access_code"
                    value={formData.beta_access_code}
                    onChange={handleInputChange}
                    placeholder="Enter your beta access code"
                    className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 pl-12 ${
                      errors.beta_access_code
                        ? "border-red-300"
                        : "border-gray-300"
                    }`}
                  />
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <TicketIcon className="h-5 w-5 text-gray-400" />
                  </div>
                </div>
                {errors.beta_access_code && (
                  <p className="mt-1 text-xs text-red-600">
                    {errors.beta_access_code}
                  </p>
                )}
              </div>

              {/* Name Fields */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label
                    htmlFor="first_name"
                    className="block text-sm font-semibold text-gray-700 mb-2"
                  >
                    First Name
                  </label>
                  <input
                    type="text"
                    id="first_name"
                    name="first_name"
                    value={formData.first_name}
                    onChange={handleInputChange}
                    placeholder="John"
                    className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 ${
                      errors.first_name ? "border-red-300" : "border-gray-300"
                    }`}
                  />
                  {errors.first_name && (
                    <p className="mt-1 text-xs text-red-600">
                      {errors.first_name}
                    </p>
                  )}
                </div>
                <div>
                  <label
                    htmlFor="last_name"
                    className="block text-sm font-semibold text-gray-700 mb-2"
                  >
                    Last Name
                  </label>
                  <input
                    type="text"
                    id="last_name"
                    name="last_name"
                    value={formData.last_name}
                    onChange={handleInputChange}
                    placeholder="Doe"
                    className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 ${
                      errors.last_name ? "border-red-300" : "border-gray-300"
                    }`}
                  />
                  {errors.last_name && (
                    <p className="mt-1 text-xs text-red-600">
                      {errors.last_name}
                    </p>
                  )}
                </div>
              </div>

              {/* Email */}
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
                    name="email"
                    value={formData.email}
                    onChange={handleInputChange}
                    placeholder="john.doe@example.com"
                    className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 pl-12 ${
                      errors.email
                        ? "border-red-300"
                        : formData.email &&
                            formData.email.includes("@") &&
                            formData.email.includes(".")
                          ? "border-green-300 bg-green-50"
                          : "border-gray-300"
                    }`}
                  />
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <EnvelopeIcon className="h-5 w-5 text-gray-400" />
                  </div>
                  {formData.email &&
                    formData.email.includes("@") &&
                    formData.email.includes(".") && (
                      <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                        <CheckIcon className="h-5 w-5 text-green-500" />
                      </div>
                    )}
                </div>
                {errors.email && (
                  <p className="mt-1 text-xs text-red-600">{errors.email}</p>
                )}
              </div>

              {/* Phone */}
              <div>
                <label
                  htmlFor="phone"
                  className="block text-sm font-semibold text-gray-700 mb-2"
                >
                  Phone Number
                </label>
                <div className="relative">
                  <input
                    type="tel"
                    id="phone"
                    name="phone"
                    value={formData.phone}
                    onChange={handleInputChange}
                    placeholder="+1 (555) 123-4567"
                    className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 pl-12 ${
                      errors.phone ? "border-red-300" : "border-gray-300"
                    }`}
                  />
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <PhoneIcon className="h-5 w-5 text-gray-400" />
                  </div>
                </div>
                {errors.phone && (
                  <p className="mt-1 text-xs text-red-600">{errors.phone}</p>
                )}
              </div>

              {/* Location Fields */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label
                    htmlFor="country"
                    className="block text-sm font-semibold text-gray-700 mb-2"
                  >
                    Country
                  </label>
                  <div className="relative">
                    <select
                      id="country"
                      name="country"
                      value={formData.country}
                      onChange={handleInputChange}
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 pl-12 appearance-none"
                    >
                      {countries.map((country) => (
                        <option key={country} value={country}>
                          {country}
                        </option>
                      ))}
                    </select>
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <GlobeAltIcon className="h-5 w-5 text-gray-400" />
                    </div>
                  </div>
                </div>
                <div>
                  <label
                    htmlFor="timezone"
                    className="block text-sm font-semibold text-gray-700 mb-2"
                  >
                    Timezone
                  </label>
                  <div className="relative">
                    <select
                      id="timezone"
                      name="timezone"
                      value={formData.timezone}
                      onChange={handleInputChange}
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 pl-12 appearance-none"
                    >
                      {timezones.map((tz) => (
                        <option key={tz.value} value={tz.value}>
                          {tz.label}
                        </option>
                      ))}
                    </select>
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <ClockIcon className="h-5 w-5 text-gray-400" />
                    </div>
                  </div>
                </div>
              </div>

              {/* Service Selection */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-3">
                  Select Your Service
                </label>
                <div className="grid grid-cols-2 gap-4">
                  <label
                    className={`relative flex cursor-pointer rounded-lg border p-4 focus:outline-none transition-all duration-200 ${
                      formData.module === 1
                        ? "border-red-800 bg-red-50 ring-2 ring-red-800"
                        : "border-gray-300 hover:border-gray-400"
                    }`}
                  >
                    <input
                      type="radio"
                      name="module"
                      value={1}
                      checked={formData.module === 1}
                      onChange={(e) =>
                        handleInputChange({
                          target: {
                            name: "module",
                            value: parseInt(e.target.value),
                          },
                        })
                      }
                      className="sr-only"
                    />
                    <div className="flex flex-col">
                      <span className="block text-sm font-semibold text-gray-900">
                        MapleFile
                      </span>
                      <span className="mt-1 text-xs text-gray-500">
                        Secure file storage & sharing
                      </span>
                    </div>
                    {formData.module === 1 && (
                      <CheckIcon className="absolute right-3 top-3 h-5 w-5 text-red-800" />
                    )}
                  </label>
                  <label
                    className={`relative flex cursor-pointer rounded-lg border p-4 focus:outline-none transition-all duration-200 ${
                      formData.module === 2
                        ? "border-red-800 bg-red-50 ring-2 ring-red-800"
                        : "border-gray-300 hover:border-gray-400"
                    }`}
                  >
                    <input
                      type="radio"
                      name="module"
                      value={2}
                      checked={formData.module === 2}
                      onChange={(e) =>
                        handleInputChange({
                          target: {
                            name: "module",
                            value: parseInt(e.target.value),
                          },
                        })
                      }
                      className="sr-only"
                    />
                    <div className="flex flex-col">
                      <span className="block text-sm font-semibold text-gray-900">
                        PaperCloud
                      </span>
                      <span className="mt-1 text-xs text-gray-500">
                        Document management
                      </span>
                    </div>
                    {formData.module === 2 && (
                      <CheckIcon className="absolute right-3 top-3 h-5 w-5 text-red-800" />
                    )}
                  </label>
                </div>
              </div>

              {/* Master Password Section */}
              <div className="space-y-4 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
                <h3 className="text-sm font-semibold text-blue-900 flex items-center">
                  <KeyIcon className="h-4 w-4 mr-2" />
                  Master Password for Encryption
                </h3>
                <p className="text-xs text-blue-800">
                  This password generates your encryption keys locally and is
                  never sent to our servers
                </p>

                {/* Password Field */}
                <div>
                  <label
                    htmlFor="password"
                    className="block text-sm font-semibold text-gray-700 mb-2"
                  >
                    Master Password
                  </label>
                  <div className="relative">
                    <input
                      type={showPassword ? "text" : "password"}
                      id="password"
                      name="password"
                      value={formData.password}
                      onChange={handleInputChange}
                      placeholder="Enter a strong password"
                      className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 pr-12 ${
                        errors.password ? "border-red-300" : "border-gray-300"
                      }`}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    >
                      {showPassword ? (
                        <EyeSlashIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                      ) : (
                        <EyeIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                      )}
                    </button>
                  </div>
                  {errors.password && (
                    <p className="mt-1 text-xs text-red-600">
                      {errors.password}
                    </p>
                  )}
                </div>

                {/* Confirm Password Field */}
                <div>
                  <label
                    htmlFor="confirmPassword"
                    className="block text-sm font-semibold text-gray-700 mb-2"
                  >
                    Confirm Master Password
                  </label>
                  <div className="relative">
                    <input
                      type={showConfirmPassword ? "text" : "password"}
                      id="confirmPassword"
                      name="confirmPassword"
                      value={formData.confirmPassword}
                      onChange={handleInputChange}
                      placeholder="Confirm your password"
                      className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 text-gray-900 placeholder-gray-500 pr-12 ${
                        errors.confirmPassword
                          ? "border-red-300"
                          : formData.confirmPassword &&
                              formData.password === formData.confirmPassword
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
                  {errors.confirmPassword && (
                    <p className="mt-1 text-xs text-red-600">
                      {errors.confirmPassword}
                    </p>
                  )}
                  {formData.confirmPassword &&
                    formData.password === formData.confirmPassword && (
                      <p className="mt-1 text-xs text-green-600 flex items-center">
                        <CheckIcon className="h-3 w-3 mr-1" />
                        Passwords match
                      </p>
                    )}
                </div>
              </div>

              {/* Agreements */}
              <div className="space-y-3">
                {/* Terms of Service */}
                <div className="relative flex items-start">
                  <div className="flex items-center h-5">
                    <input
                      type="checkbox"
                      id="agree_terms_of_service"
                      name="agree_terms_of_service"
                      checked={formData.agree_terms_of_service}
                      onChange={handleInputChange}
                      className="h-4 w-4 text-red-800 border-gray-300 rounded focus:ring-red-500"
                    />
                  </div>
                  <div className="ml-3 text-sm">
                    <label
                      htmlFor="agree_terms_of_service"
                      className="font-medium text-gray-700"
                    >
                      I agree to the{" "}
                      <Link
                        to="/terms"
                        className="text-red-600 hover:text-red-700 underline"
                      >
                        Terms of Service
                      </Link>{" "}
                      and{" "}
                      <Link
                        to="/privacy"
                        className="text-red-600 hover:text-red-700 underline"
                      >
                        Privacy Policy
                      </Link>{" "}
                      *
                    </label>
                  </div>
                </div>
                {errors.agree_terms_of_service && (
                  <p className="ml-7 text-xs text-red-600">
                    {errors.agree_terms_of_service}
                  </p>
                )}

                {/* Promotional Communications */}
                <div className="relative flex items-start">
                  <div className="flex items-center h-5">
                    <input
                      type="checkbox"
                      id="agree_promotions"
                      name="agree_promotions"
                      checked={formData.agree_promotions}
                      onChange={handleInputChange}
                      className="h-4 w-4 text-red-800 border-gray-300 rounded focus:ring-red-500"
                    />
                  </div>
                  <div className="ml-3 text-sm">
                    <label
                      htmlFor="agree_promotions"
                      className="font-medium text-gray-700"
                    >
                      Send me product updates and security alerts
                    </label>
                  </div>
                </div>

                {/* Tracking */}
                <div className="relative flex items-start">
                  <div className="flex items-center h-5">
                    <input
                      type="checkbox"
                      id="agree_to_tracking_across_third_party_apps_and_services"
                      name="agree_to_tracking_across_third_party_apps_and_services"
                      checked={
                        formData.agree_to_tracking_across_third_party_apps_and_services
                      }
                      onChange={handleInputChange}
                      className="h-4 w-4 text-red-800 border-gray-300 rounded focus:ring-red-500"
                    />
                  </div>
                  <div className="ml-3 text-sm">
                    <label
                      htmlFor="agree_to_tracking_across_third_party_apps_and_services"
                      className="font-medium text-gray-700"
                    >
                      Allow analytics to improve our service
                    </label>
                  </div>
                </div>
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                disabled={loading}
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
                    {keyGenerationProgress
                      ? "Generating encryption keys..."
                      : "Creating your secure account..."}
                  </>
                ) : (
                  <>
                    <ShieldCheckIcon className="mr-2 h-5 w-5" />
                    Create Secure Account
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
                    <PrivacyIcon className="h-4 w-4 text-purple-600" />
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
                  We'll generate your encryption keys locally in your browser
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  You'll receive a recovery phrase to backup your account
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  We'll send a verification email to confirm your address
                </li>
                <li className="flex items-start">
                  <span className="text-blue-500 mr-2 mt-0.5">•</span>
                  Your files will be encrypted before leaving your device
                </li>
              </ul>
            </div>

            {/* Alternative Actions */}
            <div className="mt-6 text-center text-sm">
              <span className="text-gray-600">Already have an account? </span>
              <Link
                to="/login"
                className="text-red-600 hover:text-red-700 font-medium hover:underline transition-colors duration-200"
              >
                Sign in
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

export default Register;
