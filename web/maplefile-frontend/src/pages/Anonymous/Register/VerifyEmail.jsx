// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Register/VerifyEmail.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";

const VerifyEmail = () => {
  const navigate = useNavigate();
  const { authService } = useServices();
  const [email, setEmail] = useState("");
  const [verificationCode, setVerificationCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    // Get email from sessionStorage
    const registeredEmail = sessionStorage.getItem("registeredEmail");

    if (!registeredEmail) {
      // Redirect back to registration if no email found
      navigate("/register");
      return;
    }

    setEmail(registeredEmail);
  }, [navigate]);

  const handleInputChange = (e) => {
    const value = e.target.value.replace(/\D/g, ""); // Only allow digits
    if (value.length <= 6) {
      setVerificationCode(value);
      if (error) setError("");
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (verificationCode.length !== 6) {
      setError("Please enter a 6-digit verification code");
      return;
    }

    setLoading(true);
    setError("");

    try {
      console.log("Verifying email with code:", verificationCode);
      const result = await authService.verifyEmail(verificationCode);

      console.log("Email verification successful:", result);

      // Store user role for success page
      sessionStorage.setItem("userRole", result.user_role.toString());

      // Navigate to success page
      navigate("/register/verify-success");
    } catch (error) {
      console.error("Email verification failed:", error);
      setError(
        error.message ||
          "Verification failed. Please check your code and try again.",
      );
    } finally {
      setLoading(false);
    }
  };

  const handleResendCode = () => {
    // In a real app, you'd implement a resend verification email endpoint
    alert(
      "Resend functionality would be implemented here.\nFor now, please check your email for the original verification code.",
    );
  };

  const handleBackToRecovery = () => {
    navigate("/register/recovery");
  };

  const handleBackToRegistration = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    navigate("/register");
  };

  if (!email) {
    return (
      <div>
        <h2>Loading...</h2>
      </div>
    );
  }

  return (
    <div>
      <h2>Verify Your Email</h2>

      <p>
        We've sent a 6-digit verification code to <strong>{email}</strong>.
        Please enter it below to complete your registration.
      </p>

      {error && <div>{error}</div>}

      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="verification_code">Verification Code</label>
          <input
            type="text"
            id="verification_code"
            name="verification_code"
            value={verificationCode}
            onChange={handleInputChange}
            placeholder="Enter 6-digit code"
            maxLength="6"
          />
          <div>
            Check your email for the verification code. It may take a few
            minutes to arrive.
          </div>
        </div>

        <button
          type="submit"
          disabled={loading || verificationCode.length !== 6}
        >
          {loading ? "Verifying..." : "Verify Email"}
        </button>
      </form>

      <div>
        <button type="button" onClick={handleBackToRecovery} disabled={loading}>
          Back to Recovery Code
        </button>

        <button
          type="button"
          onClick={handleBackToRegistration}
          disabled={loading}
        >
          Start Over
        </button>

        <button type="button" onClick={handleResendCode} disabled={loading}>
          Resend Code
        </button>
      </div>

      <div>
        <h4>Troubleshooting:</h4>
        <ul>
          <li>Check your spam/junk folder</li>
          <li>Make sure you entered the correct email address</li>
          <li>The code expires after 72 hours</li>
          <li>If you still don't receive the code, try registering again</li>
        </ul>
      </div>
    </div>
  );
};

export default VerifyEmail;
