// File: monorepo/web/maplefile-frontend/src/pages/Developer/Recovery/VerifyRecovery.jsx
// Step 2: Verify Recovery - Enter recovery phrase and decrypt challenge
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../services/Services";

const DeveloperVerifyRecovery = () => {
  const navigate = useNavigate();
  const { recoveryManager } = useServices();
  const [recoveryPhrase, setRecoveryPhrase] = useState("");
  const [loading, setLoading] = useState(false);
  const [decrypting, setDecrypting] = useState(false);
  const [error, setError] = useState("");
  const [email, setEmail] = useState("");
  const [sessionInfo, setSessionInfo] = useState(null);

  useEffect(() => {
    // Check if we have an active recovery session
    const recoveryEmail = recoveryManager.getRecoveryEmail();
    const hasSession = recoveryManager.hasActiveRecoverySession();

    if (!recoveryEmail || !hasSession) {
      console.log("[VerifyRecovery] No active recovery session, redirecting");
      navigate("/developer/recovery/initiate");
      return;
    }

    setEmail(recoveryEmail);

    // Get session info for display
    const status = recoveryManager.getRecoveryStatus();
    setSessionInfo(status.sessionData);
  }, [navigate, recoveryManager]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setDecrypting(true);
    setError("");

    try {
      // Validate recovery phrase
      const words = recoveryPhrase.trim().toLowerCase().split(/\s+/);
      if (words.length !== 12) {
        throw new Error("Recovery phrase must be exactly 12 words");
      }

      // Join words with single space
      const normalizedPhrase = words.join(" ");

      console.log("[VerifyRecovery] Decrypting challenge with recovery phrase");

      // Decrypt the challenge using recovery phrase
      const decryptedChallenge =
        await recoveryManager.decryptChallengeWithRecoveryPhrase(
          normalizedPhrase,
        );

      setDecrypting(false);
      console.log(
        "[VerifyRecovery] Challenge decrypted, verifying with server",
      );

      // Verify the decrypted challenge
      const response = await recoveryManager.verifyRecovery(decryptedChallenge);

      console.log("[VerifyRecovery] Recovery verified successfully");

      // Navigate to completion step
      navigate("/developer/recovery/complete");
    } catch (error) {
      console.error("[VerifyRecovery] Recovery verification failed:", error);
      setError(error.message);
      setDecrypting(false);
    } finally {
      setLoading(false);
    }
  };

  const handleBackToInitiate = () => {
    recoveryManager.clearRecoverySession();
    navigate("/developer/recovery/initiate");
  };

  const handlePasteRecoveryPhrase = async () => {
    try {
      const text = await navigator.clipboard.readText();
      setRecoveryPhrase(text.trim());
    } catch (error) {
      console.error("Failed to read clipboard:", error);
      alert("Failed to paste from clipboard. Please paste manually.");
    }
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
      <h2>Account Recovery - Step 2</h2>
      <p>Enter your 12-word recovery phrase for {email}</p>

      <div>
        <form onSubmit={handleSubmit}>
          <div>
            <label htmlFor="recoveryPhrase">Recovery Phrase</label>
            <textarea
              id="recoveryPhrase"
              value={recoveryPhrase}
              onChange={(e) => setRecoveryPhrase(e.target.value)}
              placeholder="Enter your 12-word recovery phrase separated by spaces"
              rows={4}
              required
              disabled={loading}
              style={{ width: "100%", fontFamily: "monospace" }}
            />
            <div>
              <button
                type="button"
                onClick={handlePasteRecoveryPhrase}
                disabled={loading}
              >
                📋 Paste from Clipboard
              </button>
            </div>
            <div>
              Enter all 12 words of your recovery phrase in the correct order,
              separated by spaces
            </div>
          </div>

          {error && <div style={{ color: "red" }}>{error}</div>}

          {decrypting && (
            <div style={{ color: "blue" }}>
              Decrypting challenge with your recovery key...
            </div>
          )}

          <div>
            <button type="submit" disabled={loading}>
              {loading
                ? decrypting
                  ? "Decrypting..."
                  : "Verifying..."
                : "Verify Recovery Phrase"}
            </button>

            <button
              type="button"
              onClick={handleBackToInitiate}
              disabled={loading}
            >
              Start Over
            </button>
          </div>
        </form>

        <div>
          <h3>Recovery Phrase Format</h3>
          <ul>
            <li>Your recovery phrase consists of exactly 12 words</li>
            <li>Words should be separated by spaces</li>
            <li>
              The order of words matters - enter them exactly as you saved them
            </li>
            <li>Words are case-insensitive (will be converted to lowercase)</li>
          </ul>
        </div>

        <div>
          <h3>Security Information</h3>
          <ul>
            <li>Your recovery phrase is never sent to our servers</li>
            <li>It's used locally to derive your recovery key</li>
            <li>The recovery key decrypts a challenge to prove ownership</li>
            <li>Session expires in 10 minutes for security</li>
          </ul>
        </div>

        {sessionInfo && (
          <div style={{ fontSize: "12px", color: "#666", marginTop: "20px" }}>
            <h4>Session Information</h4>
            <p>Session ID: {sessionInfo.sessionId ? "Active" : "None"}</p>
            <p>Challenge ID: {sessionInfo.challengeId ? "Present" : "None"}</p>
            <p>
              Encrypted Challenge:{" "}
              {sessionInfo.encryptedChallenge ? "Ready" : "Missing"}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default DeveloperVerifyRecovery;
