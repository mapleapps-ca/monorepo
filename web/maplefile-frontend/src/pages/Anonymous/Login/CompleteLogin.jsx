// monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/CompleteLogin.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";

const CompleteLogin = () => {
  const navigate = useNavigate();
  const { authService, localStorageService, passwordStorageService } =
    useServices();
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [decrypting, setDecrypting] = useState(false);
  const [error, setError] = useState("");
  const [verifyData, setVerifyData] = useState(null);
  const [decryptionProgress, setDecryptionProgress] = useState("");

  useEffect(() => {
    // Get email and verification data from previous steps
    const storedEmail = localStorageService.getUserEmail();
    const storedVerifyData =
      localStorageService.getLoginSessionData("verify_response");

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
      navigate("/login/request-ott");
    }
  }, [navigate, localStorageService]);

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
      const decryptedChallenge = await authService.decryptChallenge(
        password,
        verifyData,
      );

      setDecryptionProgress("Completing authentication...");
      setDecrypting(false);

      // Complete the login with the decrypted challenge
      const response = await authService.completeLogin(
        email,
        verifyData.challengeId,
        decryptedChallenge,
      );

      console.log("[CompleteLogin] Login completed successfully!");
      console.log("[CompleteLogin] Response:", response);

      // Wait a moment for tokens to be fully stored
      await new Promise((resolve) => setTimeout(resolve, 100));

      // Check if we have unencrypted tokens
      const accessToken = localStorageService.getAccessToken();
      const refreshToken = localStorageService.getRefreshToken();

      console.log("[CompleteLogin] Stored tokens check:");
      console.log("- Access token exists:", !!accessToken);
      console.log("- Refresh token exists:", !!refreshToken);

      if (accessToken) {
        console.log(
          "- Access token preview:",
          accessToken.substring(0, 30) + "...",
        );
      }

      // For unencrypted tokens, verify both tokens exist
      const hasTokens = !!(accessToken && refreshToken);

      if (hasTokens) {
        console.log(
          "[CompleteLogin] Saving password to session storage for better convenience...",
        );

        // AFTER successful login, store the password
        passwordStorageService.setPassword(password);
        console.log("[CompleteLogin] Password stored for session");
        console.log(
          "[CompleteLogin] Password stored in",
          passwordStorageService.getStorageInfo().mode,
        );

        console.log(
          "[CompleteLogin] Unencrypted tokens found, navigating to dashboard...",
        );
        navigate("/dashboard", { replace: true });
      } else {
        console.error(
          "[CompleteLogin] No authentication tokens found after login",
        );
        setError(
          "Login completed but no authentication tokens were received. Please try again.",
        );
      }
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
    navigate("/login/verify-ott");
  };

  if (!verifyData) {
    return (
      <div>
        <h2>Loading...</h2>
        <p>Loading verification data...</p>
      </div>
    );
  }

  return (
    <div>
      <h2>Step 3: Complete Login</h2>
      <p>Enter your password to complete login for {email}</p>

      <div>
        <form onSubmit={handleSubmit}>
          <div>
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
            <div>
              Your password will be used to decrypt your secure keys locally
            </div>
          </div>

          {error && <div>{error}</div>}

          {decrypting && <div>{decryptionProgress || "Processing..."}</div>}

          <div>
            <button type="submit" disabled={loading}>
              {loading
                ? decrypting
                  ? "Decrypting..."
                  : "Completing Login..."
                : "Complete Login"}
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
          <h3>Security Information</h3>
          <div>
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
              <p>⚠️ Your encryption parameters will be upgraded after login</p>
            )}
            <p>
              <strong>Token System:</strong> Unencrypted tokens stored locally
            </p>
          </div>
        </div>

        <div>
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
            <li>
              Authentication tokens are decrypted during login and stored
              unencrypted locally
            </li>
            <li>Automatic background token refresh maintains your session</li>
          </ul>
        </div>

        {verifyData.verificationID && (
          <div>
            <h3>Verification ID (BIP39)</h3>
            <p>
              <strong>Your Verification ID:</strong>
            </p>
            <div>
              {verifyData.verificationID.split(" ").slice(0, 6).join(" ")}...
            </div>
            <div>
              This 24-word BIP39 mnemonic is derived from your public key using
              SHA256. It serves as a human-readable verification of your
              cryptographic identity.
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default CompleteLogin;
