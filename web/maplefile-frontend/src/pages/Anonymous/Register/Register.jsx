// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Register/Register.jsx
import React, { useState } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";

const Register = () => {
  const navigate = useNavigate();
  const { authService, cryptoService } = useServices();

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
  });

  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState({});
  const [generalError, setGeneralError] = useState("");

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
    "America/Toronto",
    "America/New_York",
    "America/Los_Angeles",
    "America/Chicago",
    "America/Denver",
    "America/Vancouver",
    "Europe/London",
    "Europe/Paris",
    "Europe/Berlin",
    "Asia/Tokyo",
    "Asia/Seoul",
    "Australia/Sydney",
    "UTC",
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
      newErrors.password = "Password is required for encryption key generation";
    } else if (formData.password.length < 8) {
      newErrors.password = "Password must be at least 8 characters long";
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

    try {
      // Generate E2EE data using the crypto service
      console.log("Generating encryption data...");
      const e2eeData = await cryptoService.generateE2EEData(formData.password);
      console.log("Encryption data generated successfully!");

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

      console.log("Sending registration request...");
      const result = await authService.registerUser(registrationData);

      console.log("Registration successful:", result);

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
      console.error("Registration failed:", error);

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
    }
  };

  return (
    <div>
      <h2>Create Your Account</h2>

      {generalError && <div>{generalError}</div>}

      <form onSubmit={handleSubmit}>
        {/* Beta Access Code */}
        <div>
          <label htmlFor="beta_access_code">Beta Access Code *</label>
          <input
            type="text"
            id="beta_access_code"
            name="beta_access_code"
            value={formData.beta_access_code}
            onChange={handleInputChange}
            placeholder="Enter your beta access code"
          />
          {errors.beta_access_code && <div>{errors.beta_access_code}</div>}
        </div>

        {/* Personal Information */}
        <div>
          <label htmlFor="first_name">First Name *</label>
          <input
            type="text"
            id="first_name"
            name="first_name"
            value={formData.first_name}
            onChange={handleInputChange}
            placeholder="Enter your first name"
          />
          {errors.first_name && <div>{errors.first_name}</div>}
        </div>

        <div>
          <label htmlFor="last_name">Last Name *</label>
          <input
            type="text"
            id="last_name"
            name="last_name"
            value={formData.last_name}
            onChange={handleInputChange}
            placeholder="Enter your last name"
          />
          {errors.last_name && <div>{errors.last_name}</div>}
        </div>

        <div>
          <label htmlFor="email">Email Address *</label>
          <input
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleInputChange}
            placeholder="Enter your email address"
          />
          {errors.email && <div>{errors.email}</div>}
        </div>

        <div>
          <label htmlFor="phone">Phone Number *</label>
          <input
            type="tel"
            id="phone"
            name="phone"
            value={formData.phone}
            onChange={handleInputChange}
            placeholder="Enter your phone number"
          />
          {errors.phone && <div>{errors.phone}</div>}
        </div>

        <div>
          <label htmlFor="country">Country *</label>
          <select
            id="country"
            name="country"
            value={formData.country}
            onChange={handleInputChange}
          >
            {countries.map((country) => (
              <option key={country} value={country}>
                {country}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label htmlFor="timezone">Timezone *</label>
          <select
            id="timezone"
            name="timezone"
            value={formData.timezone}
            onChange={handleInputChange}
          >
            {timezones.map((tz) => (
              <option key={tz} value={tz}>
                {tz}
              </option>
            ))}
          </select>
        </div>

        {/* Password for E2EE */}
        <div>
          <label htmlFor="password">Password (for encryption) *</label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleInputChange}
            placeholder="Enter a strong password"
          />
          <div>
            This password is used to generate your encryption keys and is never
            sent to the server. Key generation may take a few seconds on first
            use.
          </div>
          {errors.password && <div>{errors.password}</div>}
        </div>

        {/* Service Selection */}
        <div>
          <label htmlFor="module">Service *</label>
          <select
            id="module"
            name="module"
            value={formData.module}
            onChange={handleInputChange}
          >
            <option value={1}>MapleFile (File Storage & Sharing)</option>
            <option value={2}>PaperCloud (Document Management)</option>
          </select>
        </div>

        {/* Agreements */}
        <div>
          <div>
            <input
              type="checkbox"
              id="agree_terms_of_service"
              name="agree_terms_of_service"
              checked={formData.agree_terms_of_service}
              onChange={handleInputChange}
            />
            <label htmlFor="agree_terms_of_service">
              I agree to the Terms of Service *
            </label>
          </div>
          {errors.agree_terms_of_service && (
            <div>{errors.agree_terms_of_service}</div>
          )}
        </div>

        <div>
          <div>
            <input
              type="checkbox"
              id="agree_promotions"
              name="agree_promotions"
              checked={formData.agree_promotions}
              onChange={handleInputChange}
            />
            <label htmlFor="agree_promotions">
              I agree to receive promotional communications
            </label>
          </div>
        </div>

        <div>
          <div>
            <input
              type="checkbox"
              id="agree_to_tracking_across_third_party_apps_and_services"
              name="agree_to_tracking_across_third_party_apps_and_services"
              checked={
                formData.agree_to_tracking_across_third_party_apps_and_services
              }
              onChange={handleInputChange}
            />
            <label htmlFor="agree_to_tracking_across_third_party_apps_and_services">
              I agree to tracking across third-party apps and services
            </label>
          </div>
        </div>

        <button type="submit" disabled={loading}>
          {loading
            ? "Generating encryption keys and creating account..."
            : "Create Account"}
        </button>
      </form>
    </div>
  );
};

export default Register;
