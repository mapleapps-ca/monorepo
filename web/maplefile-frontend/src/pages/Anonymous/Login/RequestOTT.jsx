// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/RequestOTT.jsx
import React, { useState } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";

const RequestOTT = () => {
  const navigate = useNavigate();
  const { authService } = useServices();
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    setMessage("");

    try {
      // Validate email
      if (!email) {
        throw new Error("Email address is required");
      }

      if (!email.includes("@")) {
        throw new Error("Please enter a valid email address");
      }

      const response = await authService.requestOTT(email);

      setMessage(response.message || "Verification code sent successfully!");

      // Wait a moment to show the success message, then navigate
      setTimeout(() => {
        navigate("/login/verify-ott");
      }, 2000);
    } catch (error) {
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2>Step 1: Request Verification Code</h2>
      <p>Enter your email address to receive a verification code</p>

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
            />
          </div>

          {error && <div>{error}</div>}
          {message && <div>{message}</div>}

          <button type="submit" disabled={loading}>
            {loading ? "Sending..." : "Send Verification Code"}
          </button>
        </form>

        <div>
          <h3>What happens next?</h3>
          <ul>
            <li>We'll send a 6-digit verification code to your email</li>
            <li>Check your inbox (and spam folder) for the code</li>
            <li>Enter the code on the next page to continue</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default RequestOTT;
