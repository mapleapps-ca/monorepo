// monorepo/web/prototyping/maplefile-cli/src/pages/RequestOTT.jsx
import { useState } from "react";
import { useNavigate } from "react-router";
import { authAPI } from "../services/api";

function RequestOTT() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      console.log("Sending request to request OTT API...");

      // Use the API service instead of direct axios call
      const response = await authAPI.requestOTT(email);

      console.log("OTT request successful:", response);

      setSuccess(true);
      // Navigate to verify OTT page after successful request
      navigate("/verify-ott", { state: { email } });
    } catch (err) {
      console.error("Error requesting OTT:", err);

      // More detailed error logging
      if (err.response) {
        // The request was made and the server responded with a status code
        // that falls out of the range of 2xx
        console.error("Response error data:", err.response.data);
        console.error("Response error status:", err.response.status);
        console.error("Response error headers:", err.response.headers);
      } else if (err.request) {
        // The request was made but no response was received
        console.error("No response received:", err.request);
      } else {
        // Something happened in setting up the request that triggered an Error
        console.error("Request setup error:", err.message);
      }

      setError(
        err.response?.data?.message ||
          err.message ||
          "Failed to request verification code. Please try again.",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1>Login</h1>
      <p>Enter your email to receive a one-time verification code</p>

      {error && (
        <div
          style={{
            color: "red",
            padding: "10px",
            margin: "10px 0",
            border: "1px solid red",
            borderRadius: "4px",
          }}
        >
          {error}
        </div>
      )}

      {success && (
        <div
          style={{
            color: "green",
            padding: "10px",
            margin: "10px 0",
            border: "1px solid green",
            borderRadius: "4px",
          }}
        >
          Verification code sent! Please check your email.
        </div>
      )}

      <form onSubmit={handleSubmit}>
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

        <button type="submit" disabled={loading}>
          {loading ? "Sending..." : "Send Verification Code"}
        </button>
      </form>
    </div>
  );
}

export default RequestOTT;
