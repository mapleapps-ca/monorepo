// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Recovery/InitiateRecovery.jsx
// Step 1: Initiate Account Recovery - Enter email
import React, { useState } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../services/Services";

const InitiateRecovery = () => {
  const navigate = useNavigate();
  const { recoveryManager } = useServices();
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      // Validate email
      if (!email) {
        throw new Error("Email address is required");
      }

      if (!email.includes("@")) {
        throw new Error("Please enter a valid email address");
      }

      console.log("[InitiateRecovery] Starting recovery process for:", email);
      const response = await recoveryManager.initiateRecovery(email);

      console.log("[InitiateRecovery] Recovery initiated successfully");

      // Navigate to verification step
      navigate("/recovery/verify");
    } catch (error) {
      console.error("[InitiateRecovery] Recovery initiation failed:", error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleBackToLogin = () => {
    navigate("/login");
  };

  return (
    <div>
      <h2>Account Recovery - Step 1</h2>
      <p>Enter your email address to begin the account recovery process</p>

      <div>
        <form onSubmit={handleSubmit}>
          <div>
            <label htmlFor="email">Email Address</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Enter your email address"
              required
              disabled={loading}
              autoComplete="email"
            />
            <div>
              We'll use this to verify your identity and send recovery
              instructions
            </div>
          </div>

          {error && <div style={{ color: "red" }}>{error}</div>}

          <div>
            <button type="submit" disabled={loading}>
              {loading ? "Starting Recovery..." : "Start Recovery"}
            </button>

            <button
              type="button"
              onClick={handleBackToLogin}
              disabled={loading}
            >
              Back to Login
            </button>
          </div>
        </form>

        <div>
          <h3>Important Information</h3>
          <ul>
            <li>
              You'll need your 12-word recovery phrase to complete this process
            </li>
            <li>
              Make sure you have your recovery phrase ready before proceeding
            </li>
            <li>The recovery process will allow you to set a new password</li>
            <li>
              All your encrypted data will remain accessible after recovery
            </li>
          </ul>
        </div>

        <div>
          <h3>Security Notes</h3>
          <ul>
            <li>Recovery sessions expire after 10 minutes for security</li>
            <li>Maximum 5 recovery attempts allowed within 15 minutes</li>
            <li>Never share your recovery phrase with anyone</li>
            <li>MapleApps support will never ask for your recovery phrase</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default InitiateRecovery;
