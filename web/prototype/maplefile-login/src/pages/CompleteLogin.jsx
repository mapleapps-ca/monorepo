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
  const navigate = useNavigate();

  useEffect(() => {
    // Get email and verification data from previous steps
    const storedEmail = LocalStorageService.getUserEmail();
    const storedVerifyData =
      LocalStorageService.getLoginSessionData("verify_response");

    if (storedEmail && storedVerifyData) {
      setEmail(storedEmail);
      setVerifyData(storedVerifyData);
    } else {
      // If no data from previous steps, redirect to start
      navigate("/");
    }
  }, [navigate]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setDecrypting(true);
    setError("");

    try {
      // Validate password
      if (!password) {
        throw new Error("Password is required");
      }

      if (
        !verifyData ||
        !verifyData.challengeId ||
        !verifyData.encryptedChallenge
      ) {
        throw new Error(
          "Missing challenge data. Please start the login process again.",
        );
      }

      // Simulate the decryption process
      // In a real implementation, this would involve:
      // 1. Deriving key from password using Argon2ID with provided salt and KDF params
      // 2. Decrypting the master key with the derived key
      // 3. Decrypting the private key with the master key
      // 4. Decrypting the challenge with the private key using X25519 and ChaCha20-Poly1305

      console.log("Starting challenge decryption process...");
      console.log("Challenge ID:", verifyData.challengeId);
      console.log("Encrypted Challenge:", verifyData.encryptedChallenge);
      console.log("KDF Params:", verifyData.kdf_params);

      setDecrypting(true);
      const decryptedChallenge = await AuthService.simulateDecryption(
        verifyData.encryptedChallenge,
      );
      setDecrypting(false);

      // Complete the login with the decrypted challenge
      const response = await AuthService.completeLogin(
        email,
        verifyData.challengeId,
        decryptedChallenge,
      );

      console.log("Login completed successfully!");

      // Navigate to dashboard
      navigate("/dashboard");
    } catch (error) {
      setError(error.message);
      setDecrypting(false);
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
            />
            <small>
              Your password will be used to decrypt your secure keys
            </small>
          </div>

          {error && <div className="error-message">{error}</div>}

          {decrypting && (
            <div className="info-message">
              <div className="spinner"></div>
              Decrypting challenge data...
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
            <p>
              <strong>Current Key Version:</strong>{" "}
              {verifyData.current_key_version}
            </p>
            <p>
              <strong>Encryption:</strong> End-to-end encrypted with
              ChaCha20-Poly1305
            </p>
            <p>
              <strong>Key Derivation:</strong> Argon2ID with{" "}
              {verifyData.kdf_params?.iterations} iterations
            </p>
            {verifyData.kdf_params_need_upgrade && (
              <p className="warning">
                ⚠️ Your encryption parameters will be upgraded after login
              </p>
            )}
          </div>
        </div>

        <div className="info-box">
          <h3>What's happening?</h3>
          <ul>
            <li>Your password is used to derive an encryption key</li>
            <li>This key decrypts your master key stored securely</li>
            <li>
              The challenge is decrypted to prove you have the correct password
            </li>
            <li>
              No passwords are sent to the server - everything is client-side
            </li>
          </ul>
        </div>
      </div>
    </Layout>
  );
};

export default CompleteLogin;
