import React, { useState } from "react";
import RegistrationForm from "./components/RegistrationForm";
import EmailVerification from "./components/EmailVerification";

function App() {
  const [currentStep, setCurrentStep] = useState("register"); // 'register', 'verify', 'complete'
  const [registeredEmail, setRegisteredEmail] = useState("");
  const [userRole, setUserRole] = useState(null);

  const handleRegistrationSuccess = (email) => {
    setRegisteredEmail(email);
    setCurrentStep("verify");
  };

  const handleVerificationSuccess = (role) => {
    setUserRole(role);
    setCurrentStep("complete");
  };

  const handleBackToRegistration = () => {
    setCurrentStep("register");
    setRegisteredEmail("");
    setUserRole(null);
  };

  return (
    <div className="container">
      <h1>MapleApps Registration</h1>

      {currentStep === "register" && (
        <RegistrationForm onSuccess={handleRegistrationSuccess} />
      )}

      {currentStep === "verify" && (
        <EmailVerification
          email={registeredEmail}
          onSuccess={handleVerificationSuccess}
          onBack={handleBackToRegistration}
        />
      )}

      {currentStep === "complete" && (
        <div className="step">
          <h2>Registration Complete! 🎉</h2>
          <div className="success">
            <p>
              Welcome to MapleApps! Your email has been verified successfully.
            </p>
            <p>
              <strong>Email:</strong> {registeredEmail}
            </p>
            <p>
              <strong>User Role:</strong> {getUserRoleText(userRole)}
            </p>
            <p>You can now log in to access MapleApps services.</p>
          </div>
          <div className="navigation">
            <button
              className="btn-secondary"
              onClick={handleBackToRegistration}
            >
              Register Another Account
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

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

export default App;
