// monorepo/web/prototype/maplefile-login/src/pages/RequestOTT.jsx
import React, { useState } from "react";
import { useNavigate } from "react-router";
import Layout from "../components/Layout.jsx";
import AuthService from "../services/authService.jsx";

const RequestOTT = () => {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

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

      const response = await AuthService.requestOTT(email);

      setMessage(response.message);

      // Wait a moment to show the success message, then navigate
      setTimeout(() => {
        navigate("/verify-ott");
      }, 2000);
    } catch (error) {
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout
      title="Step 1: Request Verification Code"
      subtitle="Enter your email address to receive a verification code"
    >
      <div className="form-container">
        <form onSubmit={handleSubmit} className="auth-form">
          <div className="form-group">
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

          {error && <div className="error-message">{error}</div>}
          {message && <div className="success-message">{message}</div>}

          <button
            type="submit"
            className={`btn btn-primary ${loading ? "loading" : ""}`}
            disabled={loading}
          >
            {loading ? "Sending..." : "Send Verification Code"}
          </button>
        </form>

        <div className="info-box">
          <h3>What happens next?</h3>
          <ul>
            <li>We'll send a 6-digit verification code to your email</li>
            <li>Check your inbox (and spam folder) for the code</li>
            <li>Enter the code on the next page to continue</li>
          </ul>
        </div>
      </div>
    </Layout>
  );
};

export default RequestOTT;
