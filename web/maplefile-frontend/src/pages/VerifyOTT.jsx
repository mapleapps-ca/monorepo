// monorepo/web/prototyping/maplefile-cli/src/pages/VerifyOTT.jsx
import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router";
import { authAPI } from "../services/api";

function VerifyOTT() {
  const navigate = useNavigate();
  const location = useLocation();

  // Try to get email from location state, otherwise use empty string
  const [email, setEmail] = useState(location.state?.email || "");
  const [ott, setOtt] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // If no email is provided through state, ask the user to enter it
  const [needsEmail, setNeedsEmail] = useState(!email);

  useEffect(() => {
    if (!email) {
      setNeedsEmail(true);
    }
  }, [email]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // Use the API service instead of direct axios call
      const response = await authAPI.verifyOTT(email, ott);

      // Response contains encrypted challenge and keys needed for password-based decryption
      const authData = response.data;

      // Navigate to complete login with the auth data
      navigate("/complete-login", {
        state: {
          email,
          authData,
        },
      });
    } catch (err) {
      console.error("Error verifying OTT:", err);
      setError(
        err.response?.data?.message ||
          err.message ||
          "Invalid verification code",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1>Verify Login Code</h1>
      {error && <p>{error}</p>}

      <form onSubmit={handleSubmit}>
        {needsEmail && (
          <div>
            <label htmlFor="email">Email:</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
        )}

        <div>
          <label htmlFor="ott">Verification Code:</label>
          <input
            type="text"
            id="ott"
            value={ott}
            onChange={(e) => setOtt(e.target.value)}
            required
            placeholder="Enter 6-digit code"
            maxLength={6}
          />
        </div>

        <button type="submit" disabled={loading}>
          {loading ? "Verifying..." : "Verify Code"}
        </button>
      </form>
    </div>
  );
}

export default VerifyOTT;
