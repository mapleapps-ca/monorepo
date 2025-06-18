// src/components/EmailVerification.jsx
import React, { useState, useEffect, useRef } from "react";
import { useNavigate, useLocation, Link } from "react-router";
import { useServices } from "../contexts/ServiceContext";

const EmailVerification = () => {
  const { authService } = useServices();
  const navigate = useNavigate();
  const location = useLocation();
  const inputRef = useRef(null);

  // Get email from navigation state (passed from registration)
  const email = location.state?.email || "";

  // Form state
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [resendLoading, setResendLoading] = useState(false);
  const [resendCooldown, setResendCooldown] = useState(0);
  const [attempts, setAttempts] = useState(0);
  const [isBlocked, setIsBlocked] = useState(false);

  const MAX_ATTEMPTS = 5;
  const COOLDOWN_SECONDS = 60;
  const BLOCK_DURATION = 300; // 5 minutes

  // Auto-focus on code input and check for pending registration
  useEffect(() => {
    // If no email provided, check for pending registration
    if (!email) {
      const pending = authService.getPendingRegistration();
      if (pending && pending.email) {
        // Redirect with the pending email
        navigate("/verify-email", {
          state: { email: pending.email },
          replace: true,
        });
        return;
      } else {
        // No email and no pending registration, redirect to register
        navigate("/register");
        return;
      }
    }

    // Focus the input
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, [email, navigate, authService]);

  // Cooldown timer effect
  useEffect(() => {
    let timer;
    if (resendCooldown > 0) {
      timer = setTimeout(() => {
        setResendCooldown(resendCooldown - 1);
      }, 1000);
    }
    return () => clearTimeout(timer);
  }, [resendCooldown]);

  // Block timer effect
  useEffect(() => {
    let timer;
    if (isBlocked && resendCooldown > 0) {
      timer = setTimeout(() => {
        if (resendCooldown <= 1) {
          setIsBlocked(false);
          setAttempts(0);
        }
      }, 1000);
    }
    return () => clearTimeout(timer);
  }, [isBlocked, resendCooldown]);

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();

    if (isBlocked) {
      setError("Too many failed attempts. Please wait before trying again.");
      return;
    }

    setError("");
    setLoading(true);

    try {
      // Validate input
      if (!code || code.trim().length !== 6) {
        throw new Error("Please enter a valid 6-digit verification code");
      }

      if (!/^\d{6}$/.test(code.trim())) {
        throw new Error("Verification code must contain only digits");
      }

      // Verify the code
      const result = await authService.verifyEmail(code.trim());

      if (result.success) {
        // Reset attempts on success
        setAttempts(0);

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
      const newAttempts = attempts + 1;
      setAttempts(newAttempts);

      // Check if user should be blocked
      if (newAttempts >= MAX_ATTEMPTS) {
        setIsBlocked(true);
        setResendCooldown(BLOCK_DURATION);
        setError(
          `Too many failed attempts. Please wait ${Math.floor(BLOCK_DURATION / 60)} minutes before trying again.`,
        );
      } else {
        setError(err.message);

        // Show remaining attempts
        const remainingAttempts = MAX_ATTEMPTS - newAttempts;
        if (remainingAttempts <= 2) {
          setError(`${err.message} (${remainingAttempts} attempts remaining)`);
        }
      }
    } finally {
      setLoading(false);
    }
  };

  // Handle resend verification code
  const handleResendCode = async () => {
    if (resendCooldown > 0 || isBlocked) {
      return;
    }

    setResendLoading(true);
    setError("");

    try {
      // Use the resend method if available, otherwise show message
      try {
        const result = await authService.resendVerificationCode(email);
        if (result.success) {
          // Start cooldown
          setResendCooldown(COOLDOWN_SECONDS);
          // Show success message briefly
          const successMsg = result.message;
          setError(""); // Clear any existing errors

          // Show success state (you could use a success state instead)
          const originalPlaceholder = inputRef.current?.placeholder;
          if (inputRef.current) {
            inputRef.current.placeholder = "Code sent! Check your email";
            setTimeout(() => {
              if (inputRef.current) {
                inputRef.current.placeholder = originalPlaceholder;
              }
            }, 3000);
          }
        }
      } catch (apiError) {
        // If resend API is not available, show a helpful message
        if (
          apiError.message.includes("Network error") ||
          apiError.response?.status === 404
        ) {
          setError(
            "Resend feature is temporarily unavailable. Please check your email for the original code or contact support if needed.",
          );
          setResendCooldown(COOLDOWN_SECONDS); // Still apply cooldown
        } else {
          throw apiError;
        }
      }
    } catch (err) {
      setError(`Failed to resend code: ${err.message}`);
    } finally {
      setResendLoading(false);
    }
  };

  // Handle code input (only allow digits, max 6 characters)
  const handleCodeChange = (e) => {
    const value = e.target.value.replace(/\D/g, ""); // Remove non-digits
    if (value.length <= 6) {
      setCode(value);
      // Clear error when user starts typing
      if (error && !isBlocked) {
        setError("");
      }
    }
  };

  // Handle paste event
  const handlePaste = (e) => {
    e.preventDefault();
    const paste = (e.clipboardData || window.clipboardData).getData("text");
    const digits = paste.replace(/\D/g, "").slice(0, 6);
    setCode(digits);
    if (error && !isBlocked) {
      setError("");
    }
  };

  // Format time remaining
  const formatTimeRemaining = (seconds) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    if (minutes > 0) {
      return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
    }
    return `${remainingSeconds}s`;
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

        {error && (
          <div
            style={{
              ...styles.error,
              ...(isBlocked ? styles.blockedError : {}),
            }}
          >
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} style={styles.form}>
          <div style={styles.formGroup}>
            <label htmlFor="code" style={styles.label}>
              Verification Code
            </label>
            <input
              ref={inputRef}
              type="text"
              id="code"
              value={code}
              onChange={handleCodeChange}
              onPaste={handlePaste}
              style={{
                ...styles.codeInput,
                ...(isBlocked ? styles.disabledInput : {}),
              }}
              placeholder="000000"
              disabled={loading || isBlocked}
              maxLength={6}
              autoComplete="off"
              inputMode="numeric"
              pattern="[0-9]*"
            />
            <small style={styles.hint}>
              Enter the 6-digit code from your email
              {attempts > 0 && !isBlocked && (
                <span style={styles.attemptCounter}>
                  {" "}
                  • {MAX_ATTEMPTS - attempts} attempts remaining
                </span>
              )}
            </small>
          </div>

          <button
            type="submit"
            style={{
              ...styles.verifyButton,
              ...(isBlocked ? styles.disabledButton : {}),
              ...(loading ? styles.loadingButton : {}),
            }}
            disabled={loading || code.length !== 6 || isBlocked}
          >
            {loading ? (
              <span style={styles.buttonContent}>
                <span style={styles.spinner}></span>
                Verifying...
              </span>
            ) : isBlocked ? (
              `Blocked for ${formatTimeRemaining(resendCooldown)}`
            ) : (
              "Verify Email"
            )}
          </button>
        </form>

        <div style={styles.resendSection}>
          <p style={styles.resendText}>
            Didn't receive the code?{" "}
            <button
              type="button"
              onClick={handleResendCode}
              style={{
                ...styles.resendButton,
                ...(resendCooldown > 0 || resendLoading || isBlocked
                  ? styles.disabledResendButton
                  : {}),
              }}
              disabled={resendLoading || resendCooldown > 0 || isBlocked}
            >
              {resendLoading
                ? "Sending..."
                : resendCooldown > 0
                  ? `Resend in ${formatTimeRemaining(resendCooldown)}`
                  : "Resend Code"}
            </button>
          </p>
        </div>

        <div style={styles.helpSection}>
          <details style={styles.helpDetails}>
            <summary style={styles.helpSummary}>Need help?</summary>
            <div style={styles.helpContent}>
              <ul style={styles.helpList}>
                <li>Check your spam/junk folder</li>
                <li>Make sure you entered the correct email address</li>
                <li>The code expires after 10 minutes</li>
                <li>If you're still having trouble, contact support</li>
              </ul>
            </div>
          </details>
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
    wordBreak: "break-word",
  },
  error: {
    backgroundColor: "#f8d7da",
    color: "#721c24",
    padding: "0.75rem",
    borderRadius: "4px",
    marginBottom: "1rem",
    border: "1px solid #f5c6cb",
  },
  blockedError: {
    backgroundColor: "#f8d7da",
    borderColor: "#dc3545",
    fontWeight: "bold",
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
    transition: "border-color 0.3s",
  },
  disabledInput: {
    backgroundColor: "#f8f9fa",
    borderColor: "#e9ecef",
    color: "#6c757d",
  },
  hint: {
    display: "block",
    marginTop: "0.5rem",
    color: "#888",
    fontSize: "0.9rem",
  },
  attemptCounter: {
    color: "#dc3545",
    fontWeight: "bold",
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
    transition: "background-color 0.3s",
    minHeight: "48px",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
  },
  disabledButton: {
    backgroundColor: "#6c757d",
    cursor: "not-allowed",
  },
  loadingButton: {
    backgroundColor: "#17a2b8",
  },
  buttonContent: {
    display: "flex",
    alignItems: "center",
    gap: "0.5rem",
  },
  spinner: {
    width: "16px",
    height: "16px",
    border: "2px solid #ffffff40",
    borderTop: "2px solid #ffffff",
    borderRadius: "50%",
    animation: "spin 1s linear infinite",
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
  disabledResendButton: {
    color: "#6c757d",
    cursor: "not-allowed",
    textDecoration: "none",
  },
  helpSection: {
    marginBottom: "1rem",
  },
  helpDetails: {
    backgroundColor: "#f8f9fa",
    border: "1px solid #e9ecef",
    borderRadius: "4px",
    padding: "0.5rem",
  },
  helpSummary: {
    cursor: "pointer",
    fontWeight: "bold",
    color: "#007bff",
    outline: "none",
  },
  helpContent: {
    marginTop: "0.5rem",
    paddingTop: "0.5rem",
    borderTop: "1px solid #e9ecef",
  },
  helpList: {
    color: "#666",
    fontSize: "0.9rem",
    lineHeight: "1.4",
    margin: 0,
    paddingLeft: "1.5rem",
  },
  backSection: {
    textAlign: "center",
  },
  backLink: {
    color: "#007bff",
    textDecoration: "none",
  },
};

// Add CSS keyframes for spinner animation
if (typeof document !== "undefined") {
  const styleSheet = document.createElement("style");
  styleSheet.textContent = `
    @keyframes spin {
      0% { transform: rotate(0deg); }
      100% { transform: rotate(360deg); }
    }
  `;
  if (!document.head.querySelector("style[data-email-verification]")) {
    styleSheet.setAttribute("data-email-verification", "true");
    document.head.appendChild(styleSheet);
  }
}

export default EmailVerification;
