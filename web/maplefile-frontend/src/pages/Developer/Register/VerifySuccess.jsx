// File: monorepo/web/maplefile-frontend/src/pages/Developer/Register/VerifySuccess.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";

const DeveloperVerifySuccess = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [userRole, setUserRole] = useState(null);

  useEffect(() => {
    // Get data from sessionStorage
    const registeredEmail = sessionStorage.getItem("registeredEmail");
    const storedUserRole = sessionStorage.getItem("userRole");

    if (!registeredEmail || !storedUserRole) {
      // Redirect back to registration if no data found
      navigate("/developer/register");
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
    navigate("/developer/register");
  };

  const handleGoToLogin = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    sessionStorage.removeItem("userRole");
    // Navigate to login page
    navigate("/developer/");
  };

  if (!email || userRole === null) {
    return (
      <div>
        <h2>Loading...</h2>
      </div>
    );
  }

  return (
    <div>
      <h2>Registration Complete! 🎉</h2>

      <div>
        <h3>✅ Welcome to MapleApps!</h3>
        <p>
          Your email has been verified successfully and your account is now
          active.
        </p>

        <div>
          <h4>Account Details:</h4>
          <p>
            <strong>Email:</strong> {email}
          </p>
          <p>
            <strong>User Role:</strong> {getUserRoleText(userRole)}
          </p>
          <p>
            <strong>Status:</strong> Active
          </p>
        </div>
      </div>

      <div>
        <h4>What's Next?</h4>
        <ul>
          <li>You can now log in to access MapleApps services</li>
          <li>
            Keep your recovery code safe - you'll need it if you forget your
            password
          </li>
          <li>Check your email for additional setup instructions</li>
          <li>Start uploading and organizing your files with MapleFile</li>
        </ul>
      </div>

      <div>
        <h4>🔐 Security Reminder</h4>
        <p>
          Your account uses end-to-end encryption. Your recovery code is the
          only way to recover your data if you forget your password. Make sure
          it's stored securely!
        </p>
      </div>

      <div>
        <button type="button" onClick={handleGoToLogin}>
          Go to Login
        </button>

        <button type="button" onClick={handleRegisterAnother}>
          Register Another Account
        </button>
      </div>
    </div>
  );
};

export default DeveloperVerifySuccess;
