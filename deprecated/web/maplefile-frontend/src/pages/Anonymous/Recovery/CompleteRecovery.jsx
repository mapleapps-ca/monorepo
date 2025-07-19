// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Recovery/CompleteRecovery.jsx
// Step 3: Complete Recovery - Set new password
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../services/Services";

const CompleteRecovery = () => {
  const navigate = useNavigate();
  const { recoveryManager } = useServices();
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [recoveryPhrase, setRecoveryPhrase] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [email, setEmail] = useState("");
  const [showRecoveryPhrase, setShowRecoveryPhrase] = useState(false);

  useEffect(() => {
    // Check if we have completed verification
    const recoveryEmail = recoveryManager.getRecoveryEmail();
    const isVerified = recoveryManager.isVerificationComplete();

    if (!recoveryEmail || !isVerified) {
      console.log(
        "[CompleteRecovery] No verified recovery session, redirecting",
      );
      navigate("/recovery/initiate");
      return;
    }

    setEmail(recoveryEmail);
  }, [navigate, recoveryManager]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      // Validate passwords
      if (!newPassword) {
        throw new Error("Password is required");
      }

      if (newPassword.length < 8) {
        throw new Error("Password must be at least 8 characters long");
      }

      if (newPassword !== confirmPassword) {
        throw new Error("Passwords do not match");
      }

      // Validate recovery phrase again
      const words = recoveryPhrase.trim().toLowerCase().split(/\s+/);
      if (words.length !== 12) {
        throw new Error("Recovery phrase must be exactly 12 words");
      }

      // Join words with single space
      const normalizedPhrase = words.join(" ");

      console.log("[CompleteRecovery] Completing recovery with new password");

      // Complete recovery with both recovery phrase and new password
      const response = await recoveryManager.completeRecoveryWithPhrase(
        normalizedPhrase,
        newPassword,
      );

      console.log("[CompleteRecovery] Recovery completed successfully");

      // Show success message
      alert(
        "Account recovery completed successfully! You can now log in with your new password.",
      );

      // Navigate to login
      navigate("/login");
    } catch (error) {
      console.error("[CompleteRecovery] Recovery completion failed:", error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleBackToVerify = () => {
    navigate("/recovery/verify");
  };

  const toggleRecoveryPhraseVisibility = () => {
    setShowRecoveryPhrase(!showRecoveryPhrase);
  };

  if (!email) {
    return (
      <div>
        <h2>Loading...</h2>
        <p>Checking recovery session...</p>
      </div>
    );
  }

  return (
    <div>
      <h2>Account Recovery - Step 3</h2>
      <p>Set a new password for {email}</p>

      <div>
        <form onSubmit={handleSubmit}>
          <div>
            <label htmlFor="recoveryPhrase">
              Recovery Phrase (Required Again)
            </label>
            <textarea
              id="recoveryPhrase"
              value={recoveryPhrase}
              onChange={(e) => setRecoveryPhrase(e.target.value)}
              placeholder="Re-enter your 12-word recovery phrase"
              rows={4}
              required
              disabled={loading}
              style={{ width: "100%", fontFamily: "monospace" }}
            />
            <div>
              We need your recovery phrase again to decrypt your master key and
              re-encrypt it with your new password
            </div>
          </div>

          <div>
            <label htmlFor="newPassword">New Password</label>
            <input
              type={showRecoveryPhrase ? "text" : "password"}
              id="newPassword"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="Enter your new password"
              required
              disabled={loading}
              autoComplete="new-password"
            />
            <div>Password must be at least 8 characters long</div>
          </div>

          <div>
            <label htmlFor="confirmPassword">Confirm New Password</label>
            <input
              type={showRecoveryPhrase ? "text" : "password"}
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm your new password"
              required
              disabled={loading}
              autoComplete="new-password"
            />
          </div>

          <div>
            <label>
              <input
                type="checkbox"
                checked={showRecoveryPhrase}
                onChange={toggleRecoveryPhraseVisibility}
              />
              Show passwords
            </label>
          </div>

          {error && <div style={{ color: "red" }}>{error}</div>}

          <div>
            <button type="submit" disabled={loading}>
              {loading ? "Setting New Password..." : "Complete Recovery"}
            </button>

            <button
              type="button"
              onClick={handleBackToVerify}
              disabled={loading}
            >
              Back to Verification
            </button>
          </div>
        </form>

        <div>
          <h3>What Happens Next?</h3>
          <ul>
            <li>Your master key will be decrypted using your recovery key</li>
            <li>New encryption keys will be generated</li>
            <li>All keys will be re-encrypted with your new password</li>
            <li>Your recovery phrase remains the same for future use</li>
            <li>You'll be able to log in immediately with your new password</li>
          </ul>
        </div>

        <div>
          <h3>Security Notes</h3>
          <ul>
            <li>Choose a strong, unique password</li>
            <li>Your new password will be used to encrypt your keys</li>
            <li>Keep your recovery phrase safe - it hasn't changed</li>
            <li>All your encrypted data remains accessible</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default CompleteRecovery;
