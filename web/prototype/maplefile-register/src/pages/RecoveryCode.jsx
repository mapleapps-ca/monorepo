// monorepo/web/prototype/maplefile-register/src/pages/RecoveryCode.js
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";

const RecoveryCode = () => {
  const navigate = useNavigate();
  const [recoveryMnemonic, setRecoveryMnemonic] = useState("");
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

      // Get the recovery mnemonic that was generated on the frontend
      const mnemonic = result.recoveryMnemonic;

      if (!mnemonic) {
        console.error("No recovery mnemonic found in registration result");
        navigate("/register");
        return;
      }

      setRecoveryMnemonic(mnemonic);
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
      await navigator.clipboard.writeText(recoveryMnemonic);
      alert("Recovery phrase copied to clipboard!");
    } catch (error) {
      console.error("Failed to copy to clipboard:", error);
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = recoveryMnemonic;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand("copy");
      document.body.removeChild(textArea);
      alert("Recovery phrase copied to clipboard!");
    }
  };

  const handlePrint = () => {
    const printWindow = window.open("", "_blank");
    printWindow.document.write(`
      <html>
        <head>
          <title>MapleApps Recovery Phrase</title>
          <style>
            body {
              font-family: Arial, sans-serif;
              padding: 20px;
              line-height: 1.6;
            }
            .header {
              text-align: center;
              margin-bottom: 30px;
            }
            .warning {
              background: #fff3cd;
              border: 1px solid #ffeaa7;
              padding: 15px;
              margin: 20px 0;
              border-radius: 4px;
            }
            .mnemonic {
              background: #f8f9fa;
              border: 2px solid #dee2e6;
              padding: 20px;
              margin: 20px 0;
              border-radius: 4px;
              font-family: monospace;
              font-size: 16px;
              text-align: center;
              line-height: 2;
            }
            .word {
              display: inline-block;
              margin: 5px;
              padding: 5px 10px;
              background: white;
              border: 1px solid #ccc;
              border-radius: 3px;
            }
            .footer {
              margin-top: 30px;
              font-size: 12px;
              color: #666;
            }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>MapleApps Recovery Phrase</h1>
            <p><strong>Account:</strong> ${email}</p>
            <p><strong>Generated:</strong> ${new Date().toLocaleString()}</p>
          </div>

          <div class="warning">
            <h3>⚠️ IMPORTANT SECURITY NOTICE</h3>
            <p>This recovery phrase is the ONLY way to recover your account if you forget your password. Keep it safe and never share it with anyone.</p>
          </div>

          <div class="mnemonic">
            ${recoveryMnemonic
              .split(" ")
              .map(
                (word, index) =>
                  `<span class="word">${index + 1}. ${word}</span>`,
              )
              .join("")}
          </div>

          <div class="footer">
            <p><strong>Security Tips:</strong></p>
            <ul>
              <li>Store this phrase in a secure location (safe, safety deposit box)</li>
              <li>Consider making multiple copies and storing them separately</li>
              <li>Never store this digitally (computer files, cloud storage, photos)</li>
              <li>Never share this phrase with anyone, including MapleApps support</li>
              <li>Write clearly and double-check each word</li>
            </ul>
          </div>
        </body>
      </html>
    `);
    printWindow.document.close();
    printWindow.print();
  };

  if (!recoveryMnemonic) {
    return (
      <div className="step">
        <h2>Loading...</h2>
      </div>
    );
  }

  // Split mnemonic into words for better display
  const mnemonicWords = recoveryMnemonic.split(" ");

  return (
    <div className="step">
      <h2>🔑 Save Your Recovery Phrase</h2>

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
          ⚠️ CRITICAL: Save This 12-Word Recovery Phrase
        </h3>
        <p style={{ margin: "0 0 15px 0", color: "#856404" }}>
          This 12-word recovery phrase is your <strong>only way</strong> to
          recover your account if you forget your password. Write it down
          exactly as shown and store it in a secure location such as a safe or
          safety deposit box.
        </p>
        <p style={{ margin: "0", color: "#856404", fontSize: "14px" }}>
          <strong>Never store this digitally or share it with anyone!</strong>
        </p>
      </div>

      <div
        style={{
          backgroundColor: "#f8f9fa",
          border: "2px solid #dee2e6",
          borderRadius: "4px",
          padding: "20px",
          marginBottom: "20px",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0", textAlign: "center" }}>
          Your 12-Word Recovery Phrase
        </h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))",
            gap: "10px",
            fontFamily: "monospace",
            fontSize: "14px",
          }}
        >
          {mnemonicWords.map((word, index) => (
            <div
              key={index}
              style={{
                backgroundColor: "white",
                border: "1px solid #ccc",
                borderRadius: "4px",
                padding: "8px 12px",
                textAlign: "center",
                fontWeight: "bold",
              }}
            >
              <span style={{ color: "#666", fontSize: "12px" }}>
                {index + 1}.
              </span>{" "}
              {word}
            </div>
          ))}
        </div>
      </div>

      <div style={{ marginBottom: "20px", textAlign: "center" }}>
        <button
          type="button"
          onClick={handleCopyToClipboard}
          style={{
            marginRight: "10px",
            backgroundColor: "#6c757d",
            color: "white",
          }}
        >
          📋 Copy to Clipboard
        </button>
        <button
          type="button"
          onClick={handlePrint}
          style={{
            backgroundColor: "#6c757d",
            color: "white",
          }}
        >
          🖨️ Print Recovery Phrase
        </button>
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
            I have written down my 12-word recovery phrase and stored it
            securely *
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
          <li>Write down your 12-word recovery phrase securely</li>
          <li>Check your email ({email}) for a verification code</li>
          <li>Complete email verification to finish registration</li>
        </ol>
      </div>

      <div
        style={{
          backgroundColor: "#f8d7da",
          border: "1px solid #f5c6cb",
          borderRadius: "4px",
          padding: "15px",
          marginBottom: "20px",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0", color: "#721c24" }}>
          🚨 Security Warnings
        </h4>
        <ul
          style={{
            margin: 0,
            paddingLeft: "20px",
            color: "#721c24",
            fontSize: "14px",
          }}
        >
          <li>Never take a screenshot or photo of your recovery phrase</li>
          <li>
            Never save it in a computer file, cloud storage, or password manager
          </li>
          <li>
            Never share it via email, text message, or any digital communication
          </li>
          <li>MapleApps will never ask for your recovery phrase</li>
          <li>Anyone with this phrase can access your account and data</li>
        </ul>
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
