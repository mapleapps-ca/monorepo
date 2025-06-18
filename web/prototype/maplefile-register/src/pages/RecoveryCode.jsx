import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";

const RecoveryCode = () => {
  const navigate = useNavigate();
  const [recoveryInfo, setRecoveryInfo] = useState("");
  const [email, setEmail] = useState("");
  const [savedRecoveryCode, setSavedRecoveryCode] = useState(false);

  useEffect(() => {
    // Get registration result from sessionStorage
    const registrationResult = sessionStorage.getItem("registrationResult");
    const registeredEmail = sessionStorage.getItem("registeredEmail");

    if (!registrationResult || !registeredEmail) {
      // Redirect back to registration if no data found
      navigate("/register");
      return;
    }

    try {
      const result = JSON.parse(registrationResult);
      console.log("Full registration result:", result);
      console.log("recovery_key_info field:", result.recovery_key_info);
      console.log("Available fields:", Object.keys(result));

      // Try different possible field names
      const recoveryCode =
        result.recovery_key_info ||
        result.recovery_key ||
        result.recoveryKey ||
        result.recovery_code ||
        result.recoveryCode ||
        result.mnemonic ||
        result.backup_phrase ||
        "Recovery code not found in API response";

      setRecoveryInfo(recoveryCode);
      setEmail(registeredEmail);
    } catch (error) {
      console.error("Error parsing registration result:", error);
      navigate("/register");
    }
  }, [navigate]);

  const handleContinue = () => {
    if (!savedRecoveryCode) {
      alert(
        "Please confirm that you have saved your recovery code before continuing!",
      );
      return;
    }

    // Navigate to email verification
    navigate("/verify-email");
  };

  const handleBackToRegistration = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    navigate("/register");
  };

  const handleCopyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(recoveryInfo);
      alert("Recovery code copied to clipboard!");
    } catch (error) {
      console.error("Failed to copy to clipboard:", error);
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = recoveryInfo;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand("copy");
      document.body.removeChild(textArea);
      alert("Recovery code copied to clipboard!");
    }
  };

  if (!recoveryInfo) {
    return (
      <div className="step">
        <h2>Loading...</h2>
      </div>
    );
  }

  return (
    <div className="step">
      <h2>🔑 Save Your Recovery Code</h2>

      <div
        style={{
          backgroundColor: "#fff3cd",
          border: "1px solid #ffeaa7",
          borderRadius: "4px",
          padding: "20px",
          marginBottom: "20px",
        }}
      >
        <h3 style={{ margin: "0 0 15px 0", color: "#856404" }}>
          ⚠️ IMPORTANT: Save This Recovery Code
        </h3>
        <p style={{ margin: "0 0 15px 0", color: "#856404" }}>
          This recovery code is your <strong>only way</strong> to recover your
          account if you forget your password. Store it in a secure location
          such as a password manager or write it down and keep it safe.
        </p>
        <div
          style={{
            backgroundColor: "#f8f9fa",
            border: "2px solid #dee2e6",
            borderRadius: "4px",
            padding: "15px",
            fontFamily: "monospace",
            fontSize: "14px",
            wordBreak: "break-all",
            lineHeight: "1.5",
          }}
        >
          {recoveryInfo}
        </div>
        <div style={{ marginTop: "15px" }}>
          <button
            type="button"
            onClick={handleCopyToClipboard}
            style={{ marginRight: "10px" }}
          >
            📋 Copy to Clipboard
          </button>
        </div>
      </div>

      <div className="form-group">
        <div className="checkbox-group">
          <input
            type="checkbox"
            id="saved_recovery_code"
            checked={savedRecoveryCode}
            onChange={(e) => setSavedRecoveryCode(e.target.checked)}
          />
          <label htmlFor="saved_recovery_code">
            I have saved my recovery code in a secure location *
          </label>
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
        <h4 style={{ margin: "0 0 10px 0", color: "#0c5460" }}>Next Steps:</h4>
        <ol style={{ margin: 0, paddingLeft: "20px", color: "#0c5460" }}>
          <li>Save your recovery code securely</li>
          <li>Check your email ({email}) for a verification code</li>
          <li>Complete email verification to finish registration</li>
        </ol>
      </div>

      <div className="navigation">
        <button
          type="button"
          className="btn-secondary"
          onClick={handleBackToRegistration}
        >
          Back to Registration
        </button>

        <button
          type="button"
          onClick={handleContinue}
          disabled={!savedRecoveryCode}
          style={{
            marginLeft: "10px",
            opacity: savedRecoveryCode ? 1 : 0.6,
          }}
        >
          Continue to Email Verification
        </button>
      </div>
    </div>
  );
};

export default RecoveryCode;
