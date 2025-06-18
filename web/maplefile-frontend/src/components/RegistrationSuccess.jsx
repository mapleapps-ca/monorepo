// src/components/RegistrationSuccess.jsx
import React, { useEffect } from "react";
import { useNavigate, useLocation, Link } from "react-router";

const RegistrationSuccess = () => {
  const navigate = useNavigate();
  const location = useLocation();

  // Get data from navigation state (passed from email verification)
  const email = location.state?.email || "";
  const message =
    location.state?.message || "Registration completed successfully!";
  const userRole = location.state?.userRole;

  // Redirect to home if no email provided (direct access without proper flow)
  useEffect(() => {
    if (!email) {
      navigate("/");
    }
  }, [email, navigate]);

  // Handle login button click
  const handleGoToLogin = () => {
    navigate("/login", {
      state: {
        email: email,
        fromRegistration: true,
      },
    });
  };

  // Get user role description
  const getUserRoleDescription = (role) => {
    switch (role) {
      case 1:
        return "Root Administrator";
      case 2:
        return "Company Account";
      case 3:
        return "Individual Account";
      default:
        return "Standard User";
    }
  };

  return (
    <div style={styles.container}>
      <div style={styles.successCard}>
        {/* Success Icon */}
        <div style={styles.iconContainer}>
          <div style={styles.successIcon}>✓</div>
        </div>

        {/* Main Title */}
        <h1 style={styles.title}>Registration Successful!</h1>

        {/* Success Message */}
        <div style={styles.messageContainer}>
          <p style={styles.message}>{message}</p>
        </div>

        {/* Account Details */}
        <div style={styles.accountDetails}>
          <h3 style={styles.detailsTitle}>Account Information</h3>
          <div style={styles.detailItem}>
            <span style={styles.detailLabel}>Email:</span>
            <span style={styles.detailValue}>{email}</span>
          </div>
          {userRole && (
            <div style={styles.detailItem}>
              <span style={styles.detailLabel}>Account Type:</span>
              <span style={styles.detailValue}>
                {getUserRoleDescription(userRole)}
              </span>
            </div>
          )}
        </div>

        {/* Security Notice */}
        <div style={styles.securityNotice}>
          <h3 style={styles.securityTitle}>🔒 Security Notice</h3>
          <p style={styles.securityText}>
            For your security, you'll need to log in with your credentials to
            access your account. This ensures that only you can access your
            encrypted data.
          </p>
        </div>

        {/* Action Buttons */}
        <div style={styles.actionContainer}>
          <button onClick={handleGoToLogin} style={styles.loginButton}>
            Continue to Login
          </button>

          <Link to="/" style={styles.homeLink}>
            Go to Homepage
          </Link>
        </div>

        {/* Recovery Key Reminder */}
        <div style={styles.recoveryReminder}>
          <h4 style={styles.reminderTitle}>📝 Important Reminder</h4>
          <p style={styles.reminderText}>
            Make sure you've saved your recovery key in a secure location.
            You'll need it to recover your account if you forget your password.
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
    backgroundColor: "#f8f9fa",
  },
  successCard: {
    backgroundColor: "white",
    padding: "3rem 2rem",
    borderRadius: "12px",
    boxShadow: "0 4px 20px rgba(0, 0, 0, 0.1)",
    width: "100%",
    maxWidth: "600px",
    textAlign: "center",
  },
  iconContainer: {
    marginBottom: "1.5rem",
  },
  successIcon: {
    display: "inline-block",
    width: "80px",
    height: "80px",
    backgroundColor: "#28a745",
    color: "white",
    borderRadius: "50%",
    fontSize: "3rem",
    lineHeight: "80px",
    fontWeight: "bold",
  },
  title: {
    color: "#28a745",
    marginBottom: "1rem",
    fontSize: "2rem",
  },
  messageContainer: {
    backgroundColor: "#d4edda",
    border: "1px solid #c3e6cb",
    borderRadius: "8px",
    padding: "1rem",
    marginBottom: "2rem",
  },
  message: {
    color: "#155724",
    margin: 0,
    fontSize: "1.1rem",
  },
  accountDetails: {
    backgroundColor: "#f8f9fa",
    borderRadius: "8px",
    padding: "1.5rem",
    marginBottom: "2rem",
    textAlign: "left",
  },
  detailsTitle: {
    color: "#333",
    marginBottom: "1rem",
    marginTop: 0,
    textAlign: "center",
  },
  detailItem: {
    display: "flex",
    justifyContent: "space-between",
    marginBottom: "0.5rem",
    padding: "0.5rem 0",
    borderBottom: "1px solid #e9ecef",
  },
  detailLabel: {
    fontWeight: "bold",
    color: "#555",
  },
  detailValue: {
    color: "#333",
  },
  securityNotice: {
    backgroundColor: "#fff3cd",
    border: "1px solid #ffeaa7",
    borderRadius: "8px",
    padding: "1.5rem",
    marginBottom: "2rem",
  },
  securityTitle: {
    color: "#856404",
    marginBottom: "0.5rem",
    marginTop: 0,
  },
  securityText: {
    color: "#856404",
    margin: 0,
    lineHeight: "1.5",
  },
  actionContainer: {
    marginBottom: "2rem",
  },
  loginButton: {
    backgroundColor: "#007bff",
    color: "white",
    padding: "0.75rem 2rem",
    border: "none",
    borderRadius: "6px",
    fontSize: "1.1rem",
    fontWeight: "bold",
    cursor: "pointer",
    marginBottom: "1rem",
    width: "100%",
    maxWidth: "300px",
  },
  homeLink: {
    display: "block",
    color: "#6c757d",
    textDecoration: "none",
    fontSize: "0.9rem",
  },
  recoveryReminder: {
    backgroundColor: "#e7f3ff",
    border: "1px solid #b3d9ff",
    borderRadius: "8px",
    padding: "1rem",
    textAlign: "left",
  },
  reminderTitle: {
    color: "#004085",
    marginBottom: "0.5rem",
    marginTop: 0,
  },
  reminderText: {
    color: "#004085",
    margin: 0,
    fontSize: "0.9rem",
    lineHeight: "1.4",
  },
};

export default RegistrationSuccess;
