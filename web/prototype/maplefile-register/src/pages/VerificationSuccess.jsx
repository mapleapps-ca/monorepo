import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";

const VerificationSuccess = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [userRole, setUserRole] = useState(null);

  useEffect(() => {
    // Get data from sessionStorage
    const registeredEmail = sessionStorage.getItem("registeredEmail");
    const storedUserRole = sessionStorage.getItem("userRole");

    if (!registeredEmail || !storedUserRole) {
      // Redirect back to registration if no data found
      navigate("/register");
      return;
    }

    setEmail(registeredEmail);
    setUserRole(parseInt(storedUserRole));
  }, [navigate]);

  const getUserRoleText = (role) => {
    switch (role) {
      case 1:
        return "Root User";
      case 2:
        return "Company User";
      case 3:
        return "Individual User";
      default:
        return "Unknown";
    }
  };

  const handleRegisterAnother = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    sessionStorage.removeItem("userRole");
    navigate("/register");
  };

  const handleGoToLogin = () => {
    // In a real app, this would redirect to the login page
    alert("This would redirect to the MapleApps login page");
  };

  if (!email || userRole === null) {
    return (
      <div className="step">
        <h2>Loading...</h2>
      </div>
    );
  }

  return (
    <div className="step">
      <h2>Registration Complete! 🎉</h2>

      <div
        style={{
          backgroundColor: "#d4edda",
          border: "1px solid #c3e6cb",
          borderRadius: "4px",
          padding: "20px",
          marginBottom: "20px",
        }}
      >
        <h3 style={{ margin: "0 0 15px 0", color: "#155724" }}>
          ✅ Welcome to MapleApps!
        </h3>
        <p style={{ margin: "0 0 10px 0", color: "#155724" }}>
          Your email has been verified successfully and your account is now
          active.
        </p>

        <div
          style={{
            backgroundColor: "#ffffff",
            border: "1px solid #c3e6cb",
            borderRadius: "4px",
            padding: "15px",
            marginTop: "15px",
          }}
        >
          <h4 style={{ margin: "0 0 10px 0", color: "#155724" }}>
            Account Details:
          </h4>
          <p style={{ margin: "5px 0", color: "#155724" }}>
            <strong>Email:</strong> {email}
          </p>
          <p style={{ margin: "5px 0", color: "#155724" }}>
            <strong>User Role:</strong> {getUserRoleText(userRole)}
          </p>
          <p style={{ margin: "5px 0", color: "#155724" }}>
            <strong>Status:</strong> Active
          </p>
        </div>
      </div>

      <div
        style={{
          backgroundColor: "#d1ecf1",
          border: "1px solid #bee5eb",
          borderRadius: "4px",
          padding: "15px",
          marginBottom: "20px",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0", color: "#0c5460" }}>What's Next?</h4>
        <ul style={{ margin: 0, paddingLeft: "20px", color: "#0c5460" }}>
          <li>You can now log in to access MapleApps services</li>
          <li>
            Keep your recovery code safe - you'll need it if you forget your
            password
          </li>
          <li>Check your email for additional setup instructions</li>
          <li>Start uploading and organizing your files with MapleFile</li>
        </ul>
      </div>

      <div
        style={{
          backgroundColor: "#fff3cd",
          border: "1px solid #ffeaa7",
          borderRadius: "4px",
          padding: "15px",
          marginBottom: "20px",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0", color: "#856404" }}>
          🔐 Security Reminder
        </h4>
        <p style={{ margin: 0, color: "#856404", fontSize: "14px" }}>
          Your account uses end-to-end encryption. Your recovery code is the
          only way to recover your data if you forget your password. Make sure
          it's stored securely!
        </p>
      </div>

      <div className="navigation">
        <button
          type="button"
          onClick={handleGoToLogin}
          style={{
            backgroundColor: "#007bff",
            color: "white",
            marginRight: "10px",
          }}
        >
          Go to Login
        </button>

        <button
          type="button"
          className="btn-secondary"
          onClick={handleRegisterAnother}
        >
          Register Another Account
        </button>
      </div>
    </div>
  );
};

export default VerificationSuccess;
