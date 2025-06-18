// src/components/EmailVerification.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useLocation, Link } from "react-router";
import { useServices } from "../contexts/ServiceContext";

const EmailVerification = () => {
  const { authService } = useServices();
  const navigate = useNavigate();
  const location = useLocation();

  // Get email from navigation state (passed from registration)
  const email = location.state?.email || "";

  // Form state
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [resendLoading, setResendLoading] = useState(false);

  // Auto-focus on code input
  useEffect(() => {
    // If no email provided, redirect to registration
    if (!email) {
      navigate("/register");
    }
  }, [email, navigate]);

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      // Validate input
      if (!code || code.trim().length !== 6) {
        throw new Error("Please enter a valid 6-digit verification code");
      }

      // Verify the code
      const result = await authService.verifyEmail(code);

      if (result.success) {
        // Redirect to registration success page
        navigate("/registration-success", {
          state: {
            email: email,
            message: result.message,
            userRole: result.userRole,
          },
        });
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle resend verification code (placeholder for future implementation)
  const handleResendCode = async () => {
    setResendLoading(true);
    setError("");

    try {
      // This would require a separate API endpoint to resend the code
      // For now, just show a message
      alert(
        "Resend functionality will be implemented when the API endpoint is available.",
      );
    } catch (err) {
      setError("Failed to resend verification code. Please try again.");
    } finally {
      setResendLoading(false);
    }
  };

  // Handle code input (only allow digits, max 6 characters)
  const handleCodeChange = (e) => {
    const value = e.target.value.replace(/\D/g, ""); // Remove non-digits
    if (value.length <= 6) {
      setCode(value);
    }
  };

  return (
    <div style={styles.container}>
      <div style={styles.formCard}>
        <h2 style={styles.title}>Verify Your Email</h2>

        <div style={styles.instructions}>
          <p>We've sent a 6-digit verification code to:</p>
          <p style={styles.email}>{email}</p>
          <p>Please enter the code below to verify your email address.</p>
        </div>

        {error && <div style={styles.error}>{error}</div>}

        <form onSubmit={handleSubmit} style={styles.form}>
          <div style={styles.formGroup}>
            <label htmlFor="code" style={styles.label}>
              Verification Code
            </label>
            <input
              type="text"
              id="code"
              value={code}
              onChange={handleCodeChange}
              style={styles.codeInput}
              placeholder="000000"
              disabled={loading}
              maxLength={6}
              autoComplete="off"
              autoFocus
            />
            <small style={styles.hint}>
              Enter the 6-digit code from your email
            </small>
          </div>

          <button
            type="submit"
            style={styles.verifyButton}
            disabled={loading || code.length !== 6}
          >
            {loading ? "Verifying..." : "Verify Email"}
          </button>
        </form>

        <div style={styles.resendSection}>
          <p style={styles.resendText}>
            Didn't receive the code?{" "}
            <button
              type="button"
              onClick={handleResendCode}
              style={styles.resendButton}
              disabled={resendLoading}
            >
              {resendLoading ? "Sending..." : "Resend Code"}
            </button>
          </p>
        </div>

        <div style={styles.backSection}>
          <Link to="/register" style={styles.backLink}>
            ← Back to Registration
          </Link>
        </div>
      </div>
    </div>
  );
};

const styles = {
  container: {
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    minHeight: "80vh",
    padding: "1rem",
  },
  formCard: {
    backgroundColor: "white",
    padding: "2rem",
    borderRadius: "8px",
    boxShadow: "0 2px 10px rgba(0, 0, 0, 0.1)",
    width: "100%",
    maxWidth: "500px",
  },
  title: {
    textAlign: "center",
    marginBottom: "1.5rem",
    color: "#333",
  },
  instructions: {
    textAlign: "center",
    marginBottom: "1.5rem",
    color: "#666",
  },
  email: {
    fontWeight: "bold",
    color: "#007bff",
    margin: "0.5rem 0",
  },
  error: {
    backgroundColor: "#f8d7da",
    color: "#721c24",
    padding: "0.75rem",
    borderRadius: "4px",
    marginBottom: "1rem",
  },
  form: {
    display: "flex",
    flexDirection: "column",
  },
  formGroup: {
    marginBottom: "1.5rem",
  },
  label: {
    display: "block",
    marginBottom: "0.5rem",
    color: "#555",
    fontWeight: "bold",
  },
  codeInput: {
    width: "100%",
    padding: "1rem",
    border: "2px solid #ddd",
    borderRadius: "4px",
    fontSize: "1.5rem",
    textAlign: "center",
    letterSpacing: "0.5rem",
    boxSizing: "border-box",
    fontFamily: "monospace",
  },
  hint: {
    display: "block",
    marginTop: "0.5rem",
    color: "#888",
    fontSize: "0.9rem",
  },
  verifyButton: {
    backgroundColor: "#28a745",
    color: "white",
    padding: "0.75rem",
    border: "none",
    borderRadius: "4px",
    fontSize: "1rem",
    fontWeight: "bold",
    cursor: "pointer",
    marginBottom: "1rem",
  },
  resendSection: {
    textAlign: "center",
    marginBottom: "1rem",
    padding: "1rem 0",
    borderTop: "1px solid #eee",
  },
  resendText: {
    color: "#666",
    margin: 0,
  },
  resendButton: {
    background: "none",
    border: "none",
    color: "#007bff",
    textDecoration: "underline",
    cursor: "pointer",
    fontSize: "inherit",
  },
  backSection: {
    textAlign: "center",
  },
  backLink: {
    color: "#007bff",
    textDecoration: "none",
  },
};

export default EmailVerification;
