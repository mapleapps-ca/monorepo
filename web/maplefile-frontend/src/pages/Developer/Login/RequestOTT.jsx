// File: monorepo/web/maplefile-frontend/src/pages/Developer/Login/RequestOTT.jsx
import React, { useState } from "react";
import { useNavigate, Link } from "react-router";
import { useServices } from "../../../services/Services";

const DeveloperRequestOTT = () => {
  const navigate = useNavigate();
  const { authManager } = useServices();
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
      // Validate services
      if (!authManager) {
        throw new Error(
          "Authentication service not available. Please refresh the page.",
        );
      }

      // Validate email
      if (!email) {
        throw new Error("Email address is required");
      }

      if (!email.includes("@")) {
        throw new Error("Please enter a valid email address");
      }

      const trimmedEmail = email.trim().toLowerCase();

      console.log(
        "[RequestOTT] Requesting OTT via AuthManager for:",
        trimmedEmail,
      );

      // FIXED: Store email BEFORE making the request
      // Store in sessionStorage as a fallback
      sessionStorage.setItem("loginEmail", trimmedEmail);
      console.log("[RequestOTT] Email stored in sessionStorage:", trimmedEmail);

      // Make the OTT request
      let response;
      if (typeof authManager.requestOTT === "function") {
        response = await authManager.requestOTT(trimmedEmail);
      } else if (typeof authManager.requestOTP === "function") {
        response = await authManager.requestOTP({ email: trimmedEmail });
      } else {
        throw new Error("OTT request method not found on authManager");
      }

      setMessage(response.message || "Verification code sent successfully!");
      console.log("[RequestOTT] OTT request successful via AuthManager");

      // FIXED: Store email in localStorage via authManager if available
      try {
        if (
          authManager.setCurrentUserEmail &&
          typeof authManager.setCurrentUserEmail === "function"
        ) {
          authManager.setCurrentUserEmail(trimmedEmail);
          console.log("[RequestOTT] Email stored via authManager");
        }
      } catch (storageError) {
        console.warn(
          "[RequestOTT] Could not store email via authManager:",
          storageError,
        );
        // Continue anyway as we have sessionStorage fallback
      }

      // Wait a moment to show the success message, then navigate
      setTimeout(() => {
        navigate("/developer/login/verify-ott");
      }, 2000);
    } catch (error) {
      console.error("[RequestOTT] OTT request failed via AuthManager:", error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "400px", margin: "0 auto" }}>
      <h2>Step 1: Request Verification Code</h2>
      <p>Enter your email address to receive a verification code</p>

      {error && (
        <div
          style={{
            color: "#d32f2f",
            backgroundColor: "#ffebee",
            padding: "10px",
            borderRadius: "4px",
            marginBottom: "15px",
          }}
        >
          {error}
        </div>
      )}

      {message && (
        <div
          style={{
            color: "#2e7d32",
            backgroundColor: "#e8f5e8",
            padding: "10px",
            borderRadius: "4px",
            marginBottom: "15px",
          }}
        >
          {message}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "20px" }}>
          <label htmlFor="email">Email Address</label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Enter your email address"
            required
            disabled={loading}
            style={{
              width: "100%",
              padding: "8px",
              marginTop: "5px",
              border: "1px solid #ccc",
              borderRadius: "4px",
            }}
          />
        </div>

        <button
          type="submit"
          disabled={loading || !authManager}
          style={{
            width: "100%",
            padding: "10px",
            backgroundColor: loading || !authManager ? "#ccc" : "#1976d2",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: loading || !authManager ? "not-allowed" : "pointer",
          }}
        >
          {loading ? "Sending..." : "Send Verification Code"}
        </button>
        <br />
        <Link to="/developer">Back</Link>
      </form>

      <div style={{ marginTop: "20px", fontSize: "14px", color: "#666" }}>
        <h3>What happens next?</h3>
        <ul>
          <li>We'll send a 6-digit verification code to your email</li>
          <li>Check your inbox (and spam folder) for the code</li>
          <li>Enter the code on the next page to continue</li>
          <li>AuthManager will orchestrate the authentication flow</li>
        </ul>
      </div>

      {/* Debug Info (only in development) */}
      {import.meta.env.DEV && (
        <div
          style={{
            marginTop: "20px",
            padding: "10px",
            backgroundColor: "#f5f5f5",
            borderRadius: "4px",
            fontSize: "12px",
            color: "#666",
          }}
        >
          <strong>Debug Info:</strong>
          <br />
          AuthManager Available: {authManager ? "Yes" : "No"}
          <br />
          Email: {email}
          <br />
          SessionStorage Email: {sessionStorage.getItem("loginEmail")}
          <br />
          {authManager && (
            <>
              Has requestOTT:{" "}
              {typeof authManager.requestOTT === "function" ? "Yes" : "No"}
              <br />
              Has requestOTP:{" "}
              {typeof authManager.requestOTP === "function" ? "Yes" : "No"}
              <br />
              Has setCurrentUserEmail:{" "}
              {typeof authManager.setCurrentUserEmail === "function"
                ? "Yes"
                : "No"}
              <br />
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default DeveloperRequestOTT;
