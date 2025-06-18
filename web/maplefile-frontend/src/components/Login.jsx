// src/components/Login.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, Link, useLocation } from "react-router";
import { useServices } from "../contexts/ServiceContext";

const Login = () => {
  const { authService } = useServices();
  const navigate = useNavigate();
  const location = useLocation();

  // Get email from navigation state (if coming from registration)
  const prefilledEmail = location.state?.email || "";
  const fromRegistration = location.state?.fromRegistration || false;

  // Login step state (1: email, 2: OTT, 3: password)
  const [currentStep, setCurrentStep] = useState(1);

  // Form state
  const [email, setEmail] = useState(prefilledEmail);
  const [ott, setOtt] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  // Store OTT verification data for step 3
  const [ottData, setOttData] = useState(null);

  // Check if already authenticated
  useEffect(() => {
    if (authService.isAuthenticated()) {
      navigate("/dashboard");
    }
  }, [authService, navigate]);

  // Show registration success message if coming from registration
  useEffect(() => {
    if (fromRegistration) {
      // Could show a success toast or message here
    }
  }, [fromRegistration]);

  // Step 1: Request OTT
  const handleRequestOTT = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (!email || !email.trim()) {
        throw new Error("Please enter your email address");
      }

      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) {
        throw new Error("Please enter a valid email address");
      }

      const result = await authService.requestOTT(email);

      if (result.success) {
        setCurrentStep(2);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Step 2: Verify OTT
  const handleVerifyOTT = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (!ott || ott.trim().length !== 6) {
        throw new Error("Please enter the 6-digit verification code");
      }

      const result = await authService.verifyOTT(email, ott);

      if (result.success) {
        setOttData(result.data);
        setCurrentStep(3);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Step 3: Complete Login
  const handleCompleteLogin = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (!password || password.trim().length === 0) {
        throw new Error("Please enter your password");
      }

      if (!ottData) {
        throw new Error("Invalid login session. Please start over.");
      }

      const result = await authService.completeLogin(email, password, ottData);

      if (result.success) {
        // Redirect to dashboard on successful login
        navigate("/dashboard");
      }
    } catch (err) {
      setError(err.message);
      // If decryption fails, it might be wrong password
      if (err.message.includes("decrypt") || err.message.includes("password")) {
        setError("Incorrect password. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  // Handle OTT input (only digits, max 6)
  const handleOttChange = (e) => {
    const value = e.target.value.replace(/\D/g, "");
    if (value.length <= 6) {
      setOtt(value);
    }
  };

  // Go back to previous step
  const handleGoBack = () => {
    setError("");
    if (currentStep === 2) {
      setCurrentStep(1);
      setOtt("");
    } else if (currentStep === 3) {
      setCurrentStep(2);
      setPassword("");
    }
  };

  // Start over (go to step 1)
  const handleStartOver = () => {
    setCurrentStep(1);
    setEmail(prefilledEmail);
    setOtt("");
    setPassword("");
    setError("");
    setOttData(null);
  };

  // Render step 1: Email input
  const renderStep1 = () => (
    <>
      <h2 style={styles.title}>Sign In to MapleFile</h2>
      <p style={styles.subtitle}>
        Enter your email address to begin secure login
      </p>

      {fromRegistration && (
        <div style={styles.successNotice}>
          <p style={styles.successText}>
            ✓ Registration completed! Please sign in with your credentials.
          </p>
        </div>
      )}

      {error && <div style={styles.error}>{error}</div>}

      <form onSubmit={handleRequestOTT} style={styles.form}>
        <div style={styles.formGroup}>
          <label htmlFor="email" style={styles.label}>
            Email Address
          </label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            style={styles.input}
            placeholder="Enter your email address"
            disabled={loading}
            autoFocus
            required
          />
        </div>

        <button type="submit" style={styles.button} disabled={loading}>
          {loading ? "Sending code..." : "Continue"}
        </button>
      </form>
    </>
  );

  // Render step 2: OTT verification
  const renderStep2 = () => (
    <>
      <h2 style={styles.title}>Verify Your Email</h2>
      <p style={styles.subtitle}>
        We've sent a 6-digit code to <strong>{email}</strong>
      </p>

      {error && <div style={styles.error}>{error}</div>}

      <form onSubmit={handleVerifyOTT} style={styles.form}>
        <div style={styles.formGroup}>
          <label htmlFor="ott" style={styles.label}>
            Verification Code
          </label>
          <input
            type="text"
            id="ott"
            value={ott}
            onChange={handleOttChange}
            style={styles.codeInput}
            placeholder="000000"
            disabled={loading}
            maxLength={6}
            autoComplete="off"
            autoFocus
            required
          />
          <small style={styles.hint}>
            Enter the 6-digit code from your email
          </small>
        </div>

        <button
          type="submit"
          style={styles.button}
          disabled={loading || ott.length !== 6}
        >
          {loading ? "Verifying..." : "Verify Code"}
        </button>
      </form>

      <button
        type="button"
        onClick={handleGoBack}
        style={styles.backButton}
        disabled={loading}
      >
        ← Change Email
      </button>
    </>
  );

  // Render step 3: Password input
  const renderStep3 = () => (
    <>
      <h2 style={styles.title}>Enter Your Password</h2>
      <p style={styles.subtitle}>
        Please enter your password to decrypt your data
      </p>

      {error && <div style={styles.error}>{error}</div>}

      <form onSubmit={handleCompleteLogin} style={styles.form}>
        <div style={styles.formGroup}>
          <label htmlFor="password" style={styles.label}>
            Password
          </label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            style={styles.input}
            placeholder="Enter your password"
            disabled={loading}
            autoFocus
            required
          />
        </div>

        <button type="submit" style={styles.button} disabled={loading}>
          {loading ? "Signing in..." : "Sign In"}
        </button>
      </form>

      <button
        type="button"
        onClick={handleGoBack}
        style={styles.backButton}
        disabled={loading}
      >
        ← Go Back
      </button>
    </>
  );

  return (
    <div style={styles.container}>
      <div style={styles.formCard}>
        {/* Progress indicator */}
        <div style={styles.progressContainer}>
          <div style={styles.progressBar}>
            <div
              style={{
                ...styles.progressFill,
                width: `${(currentStep / 3) * 100}%`,
              }}
            />
          </div>
          <p style={styles.progressText}>Step {currentStep} of 3</p>
        </div>

        {/* Render current step */}
        {currentStep === 1 && renderStep1()}
        {currentStep === 2 && renderStep2()}
        {currentStep === 3 && renderStep3()}

        {/* Footer links */}
        <div style={styles.footer}>
          {currentStep !== 1 && (
            <button
              type="button"
              onClick={handleStartOver}
              style={styles.startOverButton}
            >
              Start Over
            </button>
          )}

          <p style={styles.switchText}>
            Don't have an account?{" "}
            <Link to="/register" style={styles.link}>
              Register here
            </Link>
          </p>
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
    maxWidth: "450px",
  },
  progressContainer: {
    marginBottom: "2rem",
  },
  progressBar: {
    width: "100%",
    height: "4px",
    backgroundColor: "#e9ecef",
    borderRadius: "2px",
    overflow: "hidden",
    marginBottom: "0.5rem",
  },
  progressFill: {
    height: "100%",
    backgroundColor: "#007bff",
    transition: "width 0.3s ease",
  },
  progressText: {
    textAlign: "center",
    color: "#666",
    fontSize: "0.9rem",
    margin: 0,
  },
  title: {
    textAlign: "center",
    marginBottom: "0.5rem",
    color: "#333",
  },
  subtitle: {
    textAlign: "center",
    marginBottom: "1.5rem",
    color: "#666",
    fontSize: "0.95rem",
  },
  successNotice: {
    backgroundColor: "#d4edda",
    border: "1px solid #c3e6cb",
    borderRadius: "4px",
    padding: "0.75rem",
    marginBottom: "1rem",
  },
  successText: {
    color: "#155724",
    margin: 0,
    fontSize: "0.9rem",
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
    marginBottom: "1rem",
  },
  formGroup: {
    marginBottom: "1rem",
  },
  label: {
    display: "block",
    marginBottom: "0.5rem",
    color: "#555",
    fontWeight: "bold",
  },
  input: {
    width: "100%",
    padding: "0.75rem",
    border: "1px solid #ddd",
    borderRadius: "4px",
    fontSize: "1rem",
    boxSizing: "border-box",
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
  button: {
    backgroundColor: "#007bff",
    color: "white",
    padding: "0.75rem",
    border: "none",
    borderRadius: "4px",
    fontSize: "1rem",
    fontWeight: "bold",
    cursor: "pointer",
    marginBottom: "0.5rem",
  },
  backButton: {
    backgroundColor: "transparent",
    color: "#6c757d",
    padding: "0.5rem",
    border: "1px solid #6c757d",
    borderRadius: "4px",
    fontSize: "0.9rem",
    cursor: "pointer",
    width: "100%",
  },
  footer: {
    marginTop: "1rem",
    textAlign: "center",
  },
  startOverButton: {
    backgroundColor: "transparent",
    color: "#dc3545",
    border: "none",
    fontSize: "0.9rem",
    cursor: "pointer",
    marginBottom: "1rem",
    textDecoration: "underline",
  },
  switchText: {
    color: "#666",
    margin: 0,
  },
  link: {
    color: "#007bff",
    textDecoration: "none",
  },
};

export default Login;
