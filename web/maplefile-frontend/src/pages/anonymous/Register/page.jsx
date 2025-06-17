// src/pages/anonymous/Register/page.jsx

import { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useService } from "../../../hooks/useService.js";
import { TYPES } from "../../../di/types.js";

/**
 * RegistrationPage Component
 *
 * This is a comprehensive registration form that demonstrates several important concepts:
 * 1. Complex form state management with React hooks
 * 2. Integration with dependency injection services
 * 3. Client-side cryptographic key generation
 * 4. Progressive enhancement with loading states
 * 5. Comprehensive error handling and user feedback
 * 6. Responsive design with Tailwind CSS
 */
function RegisterPage() {
  // Get our services from the dependency injection container
  // This is much cleaner than importing and instantiating them directly
  const registrationService = useService(TYPES.RegistrationService);
  const logger = useService(TYPES.LoggerService);
  const navigate = useNavigate();

  // State management for the registration form
  // We're using multiple state objects to keep related data together
  const [formData, setFormData] = useState({
    firstName: "",
    lastName: "",
    email: "",
    phone: "",
    country: "",
    timezone: "",
    password: "",
    confirmPassword: "",
    agreeTermsOfService: false,
    agreePromotions: false,
    agreeTracking: false,
    module: 1, // Default to MapleFile
  });

  // UI state management - this tracks what the user is currently experiencing
  const [uiState, setUiState] = useState({
    isSubmitting: false,
    isGeneratingKeys: false,
    showPassword: false,
    showConfirmPassword: false,
  });

  // Error handling state - we track both general errors and field-specific ones
  const [errors, setErrors] = useState({
    general: null,
    fields: {},
  });

  // Success state for showing confirmation messages
  const [successMessage, setSuccessMessage] = useState(null);

  // Load countries and timezones when the component mounts
  // This demonstrates the useEffect hook for side effects
  const [countries, setCountries] = useState([]);
  const [timezones, setTimezones] = useState([]);

  useEffect(() => {
    // Load the dropdown options from our service
    setCountries(registrationService.getCountryList());
    setTimezones(registrationService.getTimezoneList());
    logger.log("RegisterPage: Component mounted, loaded dropdown options");
  }, [registrationService, logger]);

  /**
   * Handle input changes in a reusable way
   * This function updates the form state and clears any related errors
   */
  const handleInputChange = (field, value) => {
    // Update the form data
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));

    // Clear any existing errors for this field as the user types
    if (errors.fields[field]) {
      setErrors((prev) => ({
        ...prev,
        fields: {
          ...prev.fields,
          [field]: null,
        },
      }));
    }

    // Clear the general error when user starts making changes
    if (errors.general) {
      setErrors((prev) => ({
        ...prev,
        general: null,
      }));
    }
  };

  /**
   * Validate the form before submitting
   * This provides immediate feedback to users about any issues
   */
  const validateForm = () => {
    const newFieldErrors = {};

    // Check all required fields
    if (!formData.firstName.trim())
      newFieldErrors.firstName = "First name is required";
    if (!formData.lastName.trim())
      newFieldErrors.lastName = "Last name is required";
    if (!formData.email.trim()) newFieldErrors.email = "Email is required";
    if (!formData.phone.trim())
      newFieldErrors.phone = "Phone number is required";
    if (!formData.country)
      newFieldErrors.country = "Please select your country";
    if (!formData.password) newFieldErrors.password = "Password is required";
    if (!formData.confirmPassword)
      newFieldErrors.confirmPassword = "Please confirm your password";

    // Email format validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (formData.email && !emailRegex.test(formData.email)) {
      newFieldErrors.email = "Please enter a valid email address";
    }

    // Password strength validation
    if (formData.password && formData.password.length < 8) {
      newFieldErrors.password = "Password must be at least 8 characters long";
    }

    // Password confirmation validation
    if (
      formData.password &&
      formData.confirmPassword &&
      formData.password !== formData.confirmPassword
    ) {
      newFieldErrors.confirmPassword = "Passwords do not match";
    }

    // Terms of service validation
    if (!formData.agreeTermsOfService) {
      newFieldErrors.agreeTermsOfService =
        "You must agree to the terms of service";
    }

    // Update the error state
    setErrors((prev) => ({
      ...prev,
      fields: newFieldErrors,
    }));

    // Return true if there are no errors
    return Object.keys(newFieldErrors).length === 0;
  };

  /**
   * Handle form submission
   * This is where all the magic happens - form validation, key generation, and API calls
   */
  const handleSubmit = async (e) => {
    e.preventDefault(); // Prevent the default form submission

    logger.log("RegisterPage: Form submission started");

    // Clear any previous errors or success messages
    setErrors({ general: null, fields: {} });
    setSuccessMessage(null);

    // Validate the form first
    if (!validateForm()) {
      logger.log("RegisterPage: Form validation failed");
      return;
    }

    try {
      // Step 1: Show loading state - key generation can take a moment
      setUiState((prev) => ({ ...prev, isGeneratingKeys: true }));
      logger.log("RegisterPage: Starting cryptographic key generation...");

      // Give the UI a moment to update before starting intensive crypto operations
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Step 2: Show submission state
      setUiState((prev) => ({
        ...prev,
        isGeneratingKeys: false,
        isSubmitting: true,
      }));

      // Step 3: Call the registration service to handle the complex registration process
      const result = await registrationService.registerUser(formData);

      if (result.success) {
        // Registration succeeded!
        logger.log("RegisterPage: Registration completed successfully");

        setSuccessMessage(result.message);

        // Show the recovery key warning to the user
        if (result.recoveryKeyInfo) {
          alert(result.recoveryKeyInfo);
        }

        // Navigate to email verification after a brief delay
        setTimeout(() => {
          navigate("/verify-email", {
            state: {
              email: result.email,
              message: "Please check your email for a verification code",
            },
          });
        }, 2000);
      } else {
        // Registration failed - show the error to the user
        setErrors({
          general: result.message,
          fields: result.fieldErrors || {},
        });
      }
    } catch (error) {
      // Handle unexpected errors gracefully
      logger.log(
        `RegisterPage: Unexpected error during registration: ${error.message}`,
      );
      setErrors({
        general: "An unexpected error occurred. Please try again.",
        fields: {},
      });
    } finally {
      // Always reset the loading states when done
      setUiState({
        isSubmitting: false,
        isGeneratingKeys: false,
        showPassword: false,
        showConfirmPassword: false,
      });
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Create your MapleFile account
        </h2>
        <p className="mt-2 text-center text-sm text-gray-600">
          Secure file storage with end-to-end encryption
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
          {/* Show success message if registration completed */}
          {successMessage && (
            <div className="mb-4 bg-green-50 border border-green-200 rounded-md p-4">
              <div className="flex">
                <div className="ml-3">
                  <p className="text-sm font-medium text-green-800">
                    {successMessage}
                  </p>
                  <p className="mt-1 text-sm text-green-700">
                    Redirecting to email verification...
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Show general error message */}
          {errors.general && (
            <div className="mb-4 bg-red-50 border border-red-200 rounded-md p-4">
              <div className="flex">
                <div className="ml-3">
                  <p className="text-sm font-medium text-red-800">
                    {errors.general}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* The main registration form */}
          <form className="space-y-6" onSubmit={handleSubmit}>
            {/* Personal Information Section */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900">
                Personal Information
              </h3>

              {/* First and Last Name Row */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label
                    htmlFor="firstName"
                    className="block text-sm font-medium text-gray-700"
                  >
                    First Name *
                  </label>
                  <input
                    id="firstName"
                    type="text"
                    value={formData.firstName}
                    onChange={(e) =>
                      handleInputChange("firstName", e.target.value)
                    }
                    className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm ${
                      errors.fields.firstName
                        ? "border-red-300"
                        : "border-gray-300"
                    }`}
                    placeholder="John"
                  />
                  {errors.fields.firstName && (
                    <p className="mt-1 text-sm text-red-600">
                      {errors.fields.firstName}
                    </p>
                  )}
                </div>

                <div>
                  <label
                    htmlFor="lastName"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Last Name *
                  </label>
                  <input
                    id="lastName"
                    type="text"
                    value={formData.lastName}
                    onChange={(e) =>
                      handleInputChange("lastName", e.target.value)
                    }
                    className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm ${
                      errors.fields.lastName
                        ? "border-red-300"
                        : "border-gray-300"
                    }`}
                    placeholder="Doe"
                  />
                  {errors.fields.lastName && (
                    <p className="mt-1 text-sm text-red-600">
                      {errors.fields.lastName}
                    </p>
                  )}
                </div>
              </div>

              {/* Email Field */}
              <div>
                <label
                  htmlFor="email"
                  className="block text-sm font-medium text-gray-700"
                >
                  Email Address *
                </label>
                <input
                  id="email"
                  type="email"
                  value={formData.email}
                  onChange={(e) => handleInputChange("email", e.target.value)}
                  className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm ${
                    errors.fields.email ? "border-red-300" : "border-gray-300"
                  }`}
                  placeholder="john@example.com"
                />
                {errors.fields.email && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.fields.email}
                  </p>
                )}
              </div>

              {/* Phone Field */}
              <div>
                <label
                  htmlFor="phone"
                  className="block text-sm font-medium text-gray-700"
                >
                  Phone Number *
                </label>
                <input
                  id="phone"
                  type="tel"
                  value={formData.phone}
                  onChange={(e) => handleInputChange("phone", e.target.value)}
                  className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm ${
                    errors.fields.phone ? "border-red-300" : "border-gray-300"
                  }`}
                  placeholder="+1 (555) 123-4567"
                />
                {errors.fields.phone && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.fields.phone}
                  </p>
                )}
              </div>

              {/* Country and Timezone Row */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label
                    htmlFor="country"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Country *
                  </label>
                  <select
                    id="country"
                    value={formData.country}
                    onChange={(e) =>
                      handleInputChange("country", e.target.value)
                    }
                    className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm ${
                      errors.fields.country
                        ? "border-red-300"
                        : "border-gray-300"
                    }`}
                  >
                    <option value="">Select your country</option>
                    {countries.map((country) => (
                      <option key={country} value={country}>
                        {country}
                      </option>
                    ))}
                  </select>
                  {errors.fields.country && (
                    <p className="mt-1 text-sm text-red-600">
                      {errors.fields.country}
                    </p>
                  )}
                </div>

                <div>
                  <label
                    htmlFor="timezone"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Timezone
                  </label>
                  <select
                    id="timezone"
                    value={formData.timezone}
                    onChange={(e) =>
                      handleInputChange("timezone", e.target.value)
                    }
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                  >
                    <option value="">Auto-detect timezone</option>
                    {timezones.map((tz) => (
                      <option key={tz.value} value={tz.value}>
                        {tz.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            </div>

            {/* Security Section */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900">Security</h3>

              {/* Password Field */}
              <div>
                <label
                  htmlFor="password"
                  className="block text-sm font-medium text-gray-700"
                >
                  Password *
                </label>
                <div className="mt-1 relative">
                  <input
                    id="password"
                    type={uiState.showPassword ? "text" : "password"}
                    value={formData.password}
                    onChange={(e) =>
                      handleInputChange("password", e.target.value)
                    }
                    className={`block w-full px-3 py-2 pr-10 border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm ${
                      errors.fields.password
                        ? "border-red-300"
                        : "border-gray-300"
                    }`}
                    placeholder="Choose a strong password"
                  />
                  <button
                    type="button"
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    onClick={() =>
                      setUiState((prev) => ({
                        ...prev,
                        showPassword: !prev.showPassword,
                      }))
                    }
                  >
                    <span className="text-sm text-gray-500">
                      {uiState.showPassword ? "Hide" : "Show"}
                    </span>
                  </button>
                </div>
                {errors.fields.password && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.fields.password}
                  </p>
                )}
                <p className="mt-1 text-sm text-gray-500">
                  Minimum 8 characters. This will be used to encrypt your data.
                </p>
              </div>

              {/* Confirm Password Field */}
              <div>
                <label
                  htmlFor="confirmPassword"
                  className="block text-sm font-medium text-gray-700"
                >
                  Confirm Password *
                </label>
                <div className="mt-1 relative">
                  <input
                    id="confirmPassword"
                    type={uiState.showConfirmPassword ? "text" : "password"}
                    value={formData.confirmPassword}
                    onChange={(e) =>
                      handleInputChange("confirmPassword", e.target.value)
                    }
                    className={`block w-full px-3 py-2 pr-10 border rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm ${
                      errors.fields.confirmPassword
                        ? "border-red-300"
                        : "border-gray-300"
                    }`}
                    placeholder="Confirm your password"
                  />
                  <button
                    type="button"
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                    onClick={() =>
                      setUiState((prev) => ({
                        ...prev,
                        showConfirmPassword: !prev.showConfirmPassword,
                      }))
                    }
                  >
                    <span className="text-sm text-gray-500">
                      {uiState.showConfirmPassword ? "Hide" : "Show"}
                    </span>
                  </button>
                </div>
                {errors.fields.confirmPassword && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.fields.confirmPassword}
                  </p>
                )}
              </div>
            </div>

            {/* Service Selection */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900">Service</h3>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="module"
                    value="1"
                    checked={formData.module === 1}
                    onChange={() => handleInputChange("module", 1)}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300"
                  />
                  <span className="ml-3 text-sm font-medium text-gray-700">
                    MapleFile - Secure file storage and sharing
                  </span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    name="module"
                    value="2"
                    checked={formData.module === 2}
                    onChange={() => handleInputChange("module", 2)}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300"
                  />
                  <span className="ml-3 text-sm font-medium text-gray-700">
                    PaperCloud - Document management system
                  </span>
                </label>
              </div>
            </div>

            {/* Agreements Section */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900">Agreements</h3>

              <div className="space-y-3">
                {/* Terms of Service - Required */}
                <label className="flex items-start">
                  <input
                    type="checkbox"
                    checked={formData.agreeTermsOfService}
                    onChange={(e) =>
                      handleInputChange("agreeTermsOfService", e.target.checked)
                    }
                    className={`mt-1 h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded ${
                      errors.fields.agreeTermsOfService ? "border-red-300" : ""
                    }`}
                  />
                  <span className="ml-3 text-sm text-gray-700">
                    I agree to the{" "}
                    <a
                      href="#"
                      className="text-indigo-600 hover:text-indigo-500"
                    >
                      Terms of Service
                    </a>{" "}
                    *
                  </span>
                </label>
                {errors.fields.agreeTermsOfService && (
                  <p className="ml-7 text-sm text-red-600">
                    {errors.fields.agreeTermsOfService}
                  </p>
                )}

                {/* Promotional Emails - Optional */}
                <label className="flex items-start">
                  <input
                    type="checkbox"
                    checked={formData.agreePromotions}
                    onChange={(e) =>
                      handleInputChange("agreePromotions", e.target.checked)
                    }
                    className="mt-1 h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                  <span className="ml-3 text-sm text-gray-700">
                    I would like to receive promotional emails about new
                    features and updates
                  </span>
                </label>

                {/* Analytics Tracking - Optional */}
                <label className="flex items-start">
                  <input
                    type="checkbox"
                    checked={formData.agreeTracking}
                    onChange={(e) =>
                      handleInputChange("agreeTracking", e.target.checked)
                    }
                    className="mt-1 h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                  <span className="ml-3 text-sm text-gray-700">
                    I agree to cross-platform analytics to help improve the
                    service
                  </span>
                </label>
              </div>
            </div>

            {/* Submit Button */}
            <div>
              <button
                type="submit"
                disabled={uiState.isSubmitting || uiState.isGeneratingKeys}
                className={`w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white ${
                  uiState.isSubmitting || uiState.isGeneratingKeys
                    ? "bg-gray-400 cursor-not-allowed"
                    : "bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                }`}
              >
                {uiState.isGeneratingKeys
                  ? "Generating encryption keys..."
                  : uiState.isSubmitting
                    ? "Creating your account..."
                    : "Create Account"}
              </button>
            </div>
          </form>

          {/* Additional Information */}
          <div className="mt-6">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-300" />
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-gray-500">
                  Already have an account?
                </span>
              </div>
            </div>

            <div className="mt-6">
              <a
                href="/login"
                className="w-full flex justify-center py-2 px-4 border border-gray-300 rounded-md shadow-sm bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                Sign in to your account
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default RegisterPage;
