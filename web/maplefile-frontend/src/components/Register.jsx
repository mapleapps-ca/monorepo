// src/components/Register.jsx
import React, { useState } from "react";
import { useNavigate, Link } from "react-router";
import { useServices } from "../contexts/ServiceContext";

const Register = () => {
  const { authService } = useServices();
  const navigate = useNavigate();

  // Form state
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [country, setCountry] = useState("Canada");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [agreeTerms, setAgreeTerms] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  // Common countries list
  const countries = [
    "Canada",
    "United States",
    "United Kingdom",
    "Australia",
    "Germany",
    "France",
    "Italy",
    "Spain",
    "Netherlands",
    "Sweden",
    "Norway",
    "Denmark",
    "Other",
  ];

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      // Validate inputs
      if (!name || !email || !phone || !password || !confirmPassword) {
        throw new Error("Please fill in all required fields");
      }

      if (!agreeTerms) {
        throw new Error("You must agree to the terms of service to continue");
      }

      // Check if passwords match
      if (password !== confirmPassword) {
        throw new Error("Passwords do not match");
      }

      // Check password length
      if (password.length < 8) {
        throw new Error("Password must be at least 8 characters long");
      }

      // Validate email format
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) {
        throw new Error("Please enter a valid email address");
      }

      // Validate phone format (basic validation)
      const phoneRegex = /^[+]?[\d\s\-\(\)]{10,}$/;
      if (!phoneRegex.test(phone)) {
        throw new Error("Please enter a valid phone number");
      }

      // Attempt to register
      const result = await authService.register(
        email.trim(),
        password,
        name.trim(),
        phone.trim(),
        country,
        "America/Toronto", // Default timezone
      );

      if (result.success) {
        // Show recovery key info if provided
        if (result.recoveryKeyInfo) {
          alert(result.recoveryKeyInfo);
        }

        // Redirect to email verification page
        navigate("/verify-email", {
          state: {
            email: email.trim(),
            name: name.trim(),
          },
        });
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={styles.container}>
      <div style={styles.formCard}>
        <h2 style={styles.title}>Create Your Account</h2>

        {error && <div style={styles.error}>{error}</div>}

        <form onSubmit={handleSubmit} style={styles.form}>
          <div style={styles.formGroup}>
            <label htmlFor="name" style={styles.label}>
              Full Name *
            </label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              style={styles.input}
              placeholder="Enter your full name"
              disabled={loading}
              required
            />
          </div>

          <div style={styles.formGroup}>
            <label htmlFor="email" style={styles.label}>
              Email Address *
            </label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              style={styles.input}
              placeholder="Enter your email address"
              disabled={loading}
              required
            />
          </div>

          <div style={styles.formGroup}>
            <label htmlFor="phone" style={styles.label}>
              Phone Number *
            </label>
            <input
              type="tel"
              id="phone"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              style={styles.input}
              placeholder="Enter your phone number"
              disabled={loading}
              required
            />
            <small style={styles.hint}>
              Include country code (e.g., +1 for North America)
            </small>
          </div>

          <div style={styles.formGroup}>
            <label htmlFor="country" style={styles.label}>
              Country *
            </label>
            <select
              id="country"
              value={country}
              onChange={(e) => setCountry(e.target.value)}
              style={styles.select}
              disabled={loading}
              required
            >
              {countries.map((countryOption) => (
                <option key={countryOption} value={countryOption}>
                  {countryOption}
                </option>
              ))}
            </select>
          </div>

          <div style={styles.formGroup}>
            <label htmlFor="password" style={styles.label}>
              Password *
            </label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              style={styles.input}
              placeholder="Enter your password (min 8 characters)"
              disabled={loading}
              required
            />
          </div>

          <div style={styles.formGroup}>
            <label htmlFor="confirmPassword" style={styles.label}>
              Confirm Password *
            </label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              style={styles.input}
              placeholder="Confirm your password"
              disabled={loading}
              required
            />
          </div>

          <div style={styles.checkboxGroup}>
            <label style={styles.checkboxLabel}>
              <input
                type="checkbox"
                checked={agreeTerms}
                onChange={(e) => setAgreeTerms(e.target.checked)}
                style={styles.checkbox}
                disabled={loading}
                required
              />
              <span style={styles.checkboxText}>
                I agree to the{" "}
                <a
                  href="#"
                  style={styles.link}
                  onClick={(e) => e.preventDefault()}
                >
                  Terms of Service
                </a>{" "}
                and{" "}
                <a
                  href="#"
                  style={styles.link}
                  onClick={(e) => e.preventDefault()}
                >
                  Privacy Policy
                </a>
              </span>
            </label>
          </div>

          <button
            type="submit"
            style={styles.button}
            disabled={loading || !agreeTerms}
          >
            {loading ? "Creating Account..." : "Create Account"}
          </button>
        </form>

        <p style={styles.switchText}>
          Already have an account?{" "}
          <Link to="/login" style={styles.link}>
            Login here
          </Link>
        </p>
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
  select: {
    width: "100%",
    padding: "0.75rem",
    border: "1px solid #ddd",
    borderRadius: "4px",
    fontSize: "1rem",
    boxSizing: "border-box",
    backgroundColor: "white",
  },
  hint: {
    display: "block",
    marginTop: "0.25rem",
    color: "#888",
    fontSize: "0.85rem",
  },
  checkboxGroup: {
    marginBottom: "1.5rem",
  },
  checkboxLabel: {
    display: "flex",
    alignItems: "flex-start",
    cursor: "pointer",
  },
  checkbox: {
    marginRight: "0.5rem",
    marginTop: "0.1rem",
  },
  checkboxText: {
    color: "#555",
    fontSize: "0.9rem",
    lineHeight: "1.4",
  },
  button: {
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
  switchText: {
    textAlign: "center",
    marginTop: "1rem",
    color: "#666",
  },
  link: {
    color: "#007bff",
    textDecoration: "none",
  },
};

export default Register;
