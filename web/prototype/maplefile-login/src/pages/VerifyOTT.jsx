// monorepo/web/prototype/maplefile-login/src/pages/VerifyOTT.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import Layout from "../components/Layout.jsx";
import AuthService from "../services/authService.jsx";
import LocalStorageService from "../services/localStorageService.jsx";

const VerifyOTT = () => {
  const [ott, setOtt] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    // Get email from previous step
    const storedEmail = LocalStorageService.getUserEmail();
    if (storedEmail) {
      setEmail(storedEmail);
    } else {
      // If no email stored, redirect to first step
      navigate("/");
    }
  }, [navigate]);

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

      const response = await AuthService.verifyOTT(email, ott);

      console.log("Verification successful, received encrypted keys");

      // Navigate to complete login step
      navigate("/complete-login");
    } catch (error) {
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleBackToEmail = () => {
    navigate("/");
  };

  return (
    <Layout
      title="Step 2: Verify Code"
      subtitle={`Enter the 6-digit code sent to ${email}`}
    >
      <div className="form-container">
        <form onSubmit={handleSubmit} className="auth-form">
          <div className="form-group">
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
              className="code-input"
            />
            <small>Enter the 6-digit code sent to your email</small>
          </div>

          {error && <div className="error-message">{error}</div>}

          <div className="form-actions">
            <button
              type="submit"
              className={`btn btn-primary ${loading ? "loading" : ""}`}
              disabled={loading || ott.length !== 6}
            >
              {loading ? "Verifying..." : "Verify Code"}
            </button>

            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleBackToEmail}
              disabled={loading}
            >
              Change Email
            </button>
          </div>
        </form>

        <div className="info-box">
          <h3>Didn't receive the code?</h3>
          <ul>
            <li>Check your spam/junk folder</li>
            <li>The code expires in 10 minutes</li>
            <li>
              Use the "Change Email" button to go back and request a new code
            </li>
          </ul>
        </div>
      </div>
    </Layout>
  );
};

export default VerifyOTT;
