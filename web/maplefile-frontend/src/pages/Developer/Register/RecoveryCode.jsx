// File: monorepo/web/maplefile-frontend/src/pages/Developer/Register/RecoveryCode.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";

const DeveloperRecoveryCode = () => {
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
      navigate("/developer/register");
      return;
    }

    try {
      const result = JSON.parse(registrationResult);
      console.log("Full registration result:", result);

      // Get the recovery mnemonic that was generated on the frontend
      const mnemonic = result.recoveryMnemonic;

      if (!mnemonic) {
        console.error("No recovery mnemonic found in registration result");
        navigate("/developer/register");
        return;
      }

      setRecoveryMnemonic(mnemonic);
      setEmail(registeredEmail);
    } catch (error) {
      console.error("Error parsing registration result:", error);
      navigate("/developer/register");
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
    navigate("/developer/register/verify-email");
  };

  const handleBackToRegistration = () => {
    // Clear session storage
    sessionStorage.removeItem("registrationResult");
    sessionStorage.removeItem("registeredEmail");
    navigate("/developer/register");
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
      <div>
        <h2>Loading...</h2>
      </div>
    );
  }

  // Split mnemonic into words for better display
  const mnemonicWords = recoveryMnemonic.split(" ");

  return (
    <div>
      <h2>🔑 Save Your Recovery Phrase</h2>

      <div>
        <h3>⚠️ CRITICAL: Save This 12-Word Recovery Phrase</h3>
        <p>
          This 12-word recovery phrase is your <strong>only way</strong> to
          recover your account if you forget your password. Write it down
          exactly as shown and store it in a secure location such as a safe or
          safety deposit box.
        </p>
        <p>
          <strong>Never store this digitally or share it with anyone!</strong>
        </p>
      </div>

      <div>
        <h4>Your 12-Word Recovery Phrase</h4>
        <div>
          {mnemonicWords.map((word, index) => (
            <div key={index}>
              <span>{index + 1}.</span> {word}
            </div>
          ))}
        </div>
      </div>

      <div>
        <button type="button" onClick={handleCopyToClipboard}>
          📋 Copy to Clipboard
        </button>
        <button type="button" onClick={handlePrint}>
          🖨️ Print Recovery Phrase
        </button>
      </div>

      <div>
        <div>
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

      <div>
        <h4>Next Steps:</h4>
        <ol>
          <li>Write down your 12-word recovery phrase securely</li>
          <li>Check your email ({email}) for a verification code</li>
          <li>Complete email verification to finish registration</li>
        </ol>
      </div>

      <div>
        <h4>🚨 Security Warnings</h4>
        <ul>
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

      <div>
        <button type="button" onClick={handleBackToRegistration}>
          Back to Registration
        </button>

        <button
          type="button"
          onClick={handleContinue}
          disabled={!savedRecoveryCode}
        >
          Continue to Email Verification
        </button>
      </div>
    </div>
  );
};

export default DeveloperRecoveryCode;
