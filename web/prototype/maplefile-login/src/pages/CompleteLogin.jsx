import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import Layout from "../components/Layout.jsx";
import AuthService from "../services/authService.jsx";
import LocalStorageService from "../services/localStorageService.jsx";

const CompleteLogin = () => {
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [decrypting, setDecrypting] = useState(false);
  const [error, setError] = useState("");
  const [verifyData, setVerifyData] = useState(null);
  const [decryptionProgress, setDecryptionProgress] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    // Get email and verification data from previous steps
    const storedEmail = LocalStorageService.getUserEmail();
    const storedVerifyData =
      LocalStorageService.getLoginSessionData("verify_response");

    if (storedEmail && storedVerifyData) {
      setEmail(storedEmail);
      setVerifyData(storedVerifyData);

      // Log the available data for debugging
      console.log(
        "[CompleteLogin] Available verify data:",
        Object.keys(storedVerifyData),
      );
      console.log("[CompleteLogin] Verify data:", storedVerifyData);
    } else {
      // If no data from previous steps, redirect to start
      console.error(
        "[CompleteLogin] Missing email or verify data, redirecting to start",
      );
      navigate("/");
    }
  }, [navigate]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setDecrypting(true);
    setError("");
    setDecryptionProgress("");

    try {
      // Validate password
      if (!password) {
        throw new Error("Password is required");
      }

      if (!verifyData || !verifyData.challengeId) {
        throw new Error(
          "Missing challenge data. Please start the login process again.",
        );
      }

      console.log("[CompleteLogin] Starting challenge decryption process...");
      console.log("Challenge ID:", verifyData.challengeId);

      // Update progress
      setDecryptionProgress("Initializing cryptographic libraries...");

      // Small delay to show progress
      await new Promise((resolve) => setTimeout(resolve, 500));

      setDecryptionProgress("Deriving encryption key from password...");
      await new Promise((resolve) => setTimeout(resolve, 300));

      setDecryptionProgress("Decrypting master key...");
      await new Promise((resolve) => setTimeout(resolve, 300));

      setDecryptionProgress("Decrypting private key...");
      await new Promise((resolve) => setTimeout(resolve, 300));

      setDecryptionProgress("Decrypting challenge data...");

      // Perform the actual decryption
      const decryptedChallenge = await AuthService.decryptChallenge(
        password,
        verifyData,
      );

      setDecryptionProgress("Completing authentication...");
      setDecrypting(false);

      // Complete the login with the decrypted challenge
      const response = await AuthService.completeLogin(
        email,
        verifyData.challengeId,
        decryptedChallenge,
      );

      console.log("[CompleteLogin] Login completed successfully!");

      // Navigate to dashboard
      navigate("/dashboard");
    } catch (error) {
      console.error("[CompleteLogin] Login failed:", error);
      setError(error.message);
      setDecrypting(false);
      setDecryptionProgress("");
    } finally {
      setLoading(false);
    }
  };

  const handleBackToVerify = () => {
    navigate("/verify-ott");
  };

  if (!verifyData) {
    return (
      <Layout title="Loading...">
        <div className="loading-container">
          <div className="spinner"></div>
          <p>Loading verification data...</p>
        </div>
      </Layout>
    );
  }

  return (
    <Layout
      title="Step 3: Complete Login"
      subtitle={`Enter your password to complete login for ${email}`}
    >
      <div className="form-container">
        <form onSubmit={handleSubmit} className="auth-form">
          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              required
              disabled={loading}
              autoComplete="current-password"
            />
            <small>
              Your password will be used to decrypt your secure keys locally
            </small>
          </div>

          {error && <div className="error-message">{error}</div>}

          {decrypting && (
            <div className="info-message">
              <div className="spinner"></div>
              {decryptionProgress || "Processing..."}
            </div>
          )}

          <div className="form-actions">
            <button
              type="submit"
              className={`btn btn-primary ${loading ? "loading" : ""}`}
              disabled={loading}
            >
              {loading
                ? decrypting
                  ? "Decrypting..."
                  : "Completing Login..."
                : "Complete Login"}
            </button>

            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleBackToVerify}
              disabled={loading}
            >
              Back to Verification
            </button>
          </div>
        </form>

        <div className="security-info">
          <h3>Security Information</h3>
          <div className="security-details">
            {verifyData.current_key_version && (
              <p>
                <strong>Current Key Version:</strong>{" "}
                {verifyData.current_key_version}
              </p>
            )}
            <p>
              <strong>Encryption:</strong> End-to-end encrypted with
              ChaCha20-Poly1305
            </p>
            <p>
              <strong>Key Exchange:</strong> X25519 Elliptic Curve
              Diffie-Hellman
            </p>
            <p>
              <strong>Key Derivation:</strong> PBKDF2 with 100,000 iterations
            </p>
            {verifyData.kdf_params?.iterations && (
              <p>
                <strong>KDF Iterations:</strong>{" "}
                {verifyData.kdf_params.iterations}
              </p>
            )}
            {verifyData.kdf_params_need_upgrade && (
              <p className="warning">
                ⚠️ Your encryption parameters will be upgraded after login
              </p>
            )}
          </div>
        </div>

        <div className="info-box">
          <h3>End-to-End Encryption Process</h3>
          <ul>
            <li>Your password derives an encryption key using PBKDF2</li>
            <li>
              This key decrypts your master key stored securely on the server
            </li>
            <li>The master key decrypts your private X25519 key</li>
            <li>
              Your private key decrypts the challenge using ECDH +
              ChaCha20-Poly1305
            </li>
            <li>No passwords or private keys are ever sent to the server</li>
            <li>All decryption happens locally in your browser</li>
          </ul>
        </div>

        {verifyData.verificationID && (
          <div className="info-box">
            <h3>Verification ID (BIP39)</h3>
            <p>
              <strong>Your Verification ID:</strong>
            </p>
            <code className="verification-id">
              {verifyData.verificationID.split(" ").slice(0, 6).join(" ")}...
            </code>
            <small>
              This 24-word BIP39 mnemonic is derived from your public key using
              SHA256. It serves as a human-readable verification of your
              cryptographic identity.
            </small>
          </div>
        )}
      </div>
    </Layout>
  );
};

export default CompleteLogin;
