// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/VerifyOTT.jsx
// Verify OTT Page
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";

const VerifyOTT = () => {
  const navigate = useNavigate();
  const { authManager, localStorageService } = useServices();
  const [ott, setOtt] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    // Get email from previous step
    const storedEmail = localStorageService.getUserEmail();
    if (storedEmail) {
      setEmail(storedEmail);
      console.log(
        "[VerifyOTT] Using stored email via AuthManager:",
        storedEmail,
      );
    } else {
      // If no email stored, redirect to first step
      console.log("[VerifyOTT] No stored email found, redirecting to start");
      navigate("/login/request-ott");
    }
  }, [navigate, localStorageService]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      // Validate OTT
      if (!ott) {
        throw new Error("Verification code is required");
      }

      if (ott.length !== 6 || !/^\d{6}$/.test(ott)) {
        throw new Error("Verification code must be 6 digits");
      }

      console.log("[VerifyOTT] Verifying OTT via AuthManager for:", email);
      const response = await authManager.verifyOTT(email, ott);

      console.log(
        "[VerifyOTT] Verification successful via AuthManager, received encrypted keys",
      );

      // Navigate to complete login step
      navigate("/login/complete");
    } catch (error) {
      console.error("[VerifyOTT] Verification failed via AuthManager:", error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleBackToEmail = () => {
    navigate("/login/request-ott");
  };

  return (
    <div>
      <h2>Step 2: Verify Code</h2>
      <p>Enter the 6-digit code sent to {email}</p>

      <div>
        <form onSubmit={handleSubmit}>
          <div>
            <label htmlFor="ott">Verification Code</label>
            <input
              type="text"
              id="ott"
              value={ott}
              onChange={(e) =>
                setOtt(e.target.value.replace(/\D/g, "").slice(0, 6))
              }
              placeholder="Enter 6-digit code"
              maxLength={6}
              required
              disabled={loading}
            />
            <div>Enter the 6-digit code sent to your email</div>
          </div>

          {error && <div>{error}</div>}

          <div>
            <button type="submit" disabled={loading || ott.length !== 6}>
              {loading ? "Verifying..." : "Verify Code"}
            </button>

            <button
              type="button"
              onClick={handleBackToEmail}
              disabled={loading}
            >
              Change Email
            </button>
          </div>
        </form>

        <div>
          <h3>Didn't receive the code?</h3>
          <ul>
            <li>Check your spam/junk folder</li>
            <li>The code expires in 10 minutes</li>
            <li>
              Use the "Change Email" button to go back and request a new code
            </li>
            <li>AuthManager will handle the verification orchestration</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default VerifyOTT;
