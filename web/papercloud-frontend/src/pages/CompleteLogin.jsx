// monorepo/web/prototyping/papercloud-cli/src/pages/CompleteLogin.jsx
import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router";
import { authAPI } from "../services/api";
import { useAuth } from "../contexts/AuthContext";
import { cryptoUtils } from "../utils/crypto"; // Import cryptoUtils

function CompleteLogin() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, sodium, isLoading: authLoading, authError } = useAuth(); // Get sodium from context

  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [debug, setDebug] = useState({});

  const email = location.state?.email;
  const authData = location.state?.authData; // This comes from VerifyOTT.jsx state

  useEffect(() => {
    if (!email || !authData) {
      setError(
        "Missing authentication data. Please start the login process again.",
      );
      navigate("/login");
      return;
    }
    if (authLoading) {
      // Wait for AuthContext to finish loading sodium
      console.log("CompleteLogin: AuthContext is loading...");
      return;
    }
    if (!sodium && !authLoading) {
      // If AuthContext done loading and sodium still not set
      setError(
        "Encryption library failed to initialize via AuthContext. Please refresh.",
      );
      return;
    }
    console.log("CompleteLogin: Auth data received:", authData);
    setDebug(authData);
  }, [email, authData, navigate, sodium, authLoading]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    if (!sodium) {
      setError(
        "Encryption library not ready. Please wait or refresh the page.",
      );
      setLoading(false);
      return;
    }

    try {
      console.log(
        "CompleteLogin: Attempting to derive keys and decrypt challenge with password.",
      );
      // Keys and salt are base64 encoded in authData from the server
      const userSaltBytes = await cryptoUtils.fromBase64(authData.salt);
      const userPublicKeyBytes = await cryptoUtils.fromBase64(
        authData.publicKey,
      );

      const encryptedMasterKeyCombined = await cryptoUtils.fromBase64(
        authData.encryptedMasterKey,
      );
      const encryptedPrivateKeyCombined = await cryptoUtils.fromBase64(
        authData.encryptedPrivateKey,
      );
      const encryptedChallengeSealed = await cryptoUtils.fromBase64(
        authData.encryptedChallenge,
      ); // This is crypto_box_seal output
      const challengeId = authData.challengeId;

      // 1. Derive Key Encryption Key (KEK) from password and salt
      const kek = await cryptoUtils.deriveKeyFromPassword(
        password,
        userSaltBytes,
      );
      console.log("CompleteLogin: KEK derived.");

      // 2. Decrypt Master Key using KEK
      const { nonce: mkNonce, ciphertext: mkCiphertext } =
        await cryptoUtils.splitNonceAndCiphertext(encryptedMasterKeyCombined);
      const decryptedMasterKey = await cryptoUtils.decryptWithKey(
        mkCiphertext,
        mkNonce,
        kek,
      );
      if (!decryptedMasterKey)
        throw new Error(
          "Failed to decrypt master key. Likely incorrect password.",
        );
      console.log("CompleteLogin: Master Key decrypted.");

      // 3. Decrypt Private Key using Master Key
      const { nonce: pkNonce, ciphertext: pkCiphertext } =
        await cryptoUtils.splitNonceAndCiphertext(encryptedPrivateKeyCombined);
      const decryptedPrivateKey = await cryptoUtils.decryptWithKey(
        pkCiphertext,
        pkNonce,
        decryptedMasterKey,
      );
      if (!decryptedPrivateKey)
        throw new Error("Failed to decrypt private key with master key.");
      console.log("CompleteLogin: Private Key decrypted.");

      // 4. Decrypt Challenge
      // The challenge was encrypted using crypto_box_seal (server public key, anonymous)
      // It needs to be opened with the user's keypair (publicKeyBytes, decryptedPrivateKey)
      const decryptedChallengeBytes = await cryptoUtils.decryptWithBoxSealOpen(
        encryptedChallengeSealed,
        userPublicKeyBytes,
        decryptedPrivateKey,
      );
      if (!decryptedChallengeBytes)
        throw new Error("Failed to decrypt server challenge.");

      // ADD THESE LOGS:
      console.log("CompleteLogin: Raw decryptedChallengeBytes:", decryptedChallengeBytes);
      // If you suspect it might be text, you can try logging it as a string
      try {
        const decryptedChallengeString = await cryptoUtils.bytesToString(decryptedChallengeBytes);
        console.log("CompleteLogin: Decrypted challenge as string:", decryptedChallengeString);
      } catch (e) {
        console.warn("CompleteLogin: Could not convert decrypted challenge to string", e);
      }
      // Also log its length
      console.log("CompleteLogin: Decrypted challenge byte length:", decryptedChallengeBytes.length);
      // END OF ADDED LOGS

      console.log("CompleteLogin: Server challenge decrypted.");

      const decryptedChallengeB64 = await cryptoUtils.toBase64(
        decryptedChallengeBytes,
      );
      // Log the base64 version that will be sent
      console.log("CompleteLogin: Decrypted challenge to be sent (Base64):", decryptedChallengeB64);

      // 5. Send to server to complete login
      const response = await authAPI.completeLogin(
        email,
        challengeId,
        decryptedChallengeB64,
      );
      if (!response.data || !response.data.access_token) {
        throw new Error("Server returned invalid data after completing login.");
      }
      console.log("CompleteLogin: API call to /complete-login successful.");

      // 6. Call AuthContext's login function with all necessary data
      login(
        response.data.access_token,
        new Date(response.data.access_token_expiry_time),
        response.data.refresh_token,
        new Date(response.data.refresh_token_expiry_time),
        decryptedMasterKey, // Uint8Array
        decryptedPrivateKey, // Uint8Array
        userPublicKeyBytes, // Uint8Array
        userSaltBytes, // Uint8Array
        email,
      );

      console.log("CompleteLogin: Navigating to home page...");
      navigate("/");
    } catch (err) {
      console.error("CompleteLogin handleSubmit error:", err);
      setError(
        err.message ||
          "Login completion failed. Please check your password or restart the login process.",
      );
    } finally {
      setLoading(false);
    }
  };

  const handleRestartLogin = () => {
    navigate("/login");
  };

  if (authLoading) return <div>Loading authentication context...</div>;
  if (authError)
    return (
      <div>
        Context Error: {authError}{" "}
        <button onClick={() => window.location.reload()}>Try Refreshing</button>
      </div>
    );
  if (!sodium)
    return (
      <div>
        Initializing security components... If this persists, please refresh.
      </div>
    );
  if (!email || !authData) {
    return (
      <div>
        Missing authentication data. Please <a href="/login">login again</a>.
      </div>
    );
  }

  return (
    <div>
      <h1>Enter Your Password</h1>
      <p>Please enter your password to complete login for: {email}</p>

      {error && (
        <div
          style={{
            color: "red",
            marginBottom: "15px",
            padding: "10px",
            border: "1px solid red",
            borderRadius: "4px",
          }}
        >
          <p>{error}</p>
          <button onClick={handleRestartLogin} style={{ marginTop: "10px" }}>
            Restart Login Process
          </button>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="password">Password:</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        <button type="submit" disabled={loading || !sodium}>
          {loading ? "Logging in..." : "Log In"}
        </button>
      </form>

      {process.env.NODE_ENV !== "production" &&
        Object.keys(debug).length > 0 && (
          <div style={{ marginTop: "20px", textAlign: "left" }}>
            <details>
              <summary>Debug Info (Data from VerifyOTT)</summary>
              <pre>{JSON.stringify(debug, null, 2)}</pre>
            </details>
          </div>
        )}
    </div>
  );
}

export default CompleteLogin;
