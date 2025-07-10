// File: src/pages/User/Examples/User/UserLookupExample.jsx
// Simplified example component demonstrating user lookup functionality

import React, { useState } from "react";
import { useUsers } from "../../../../services/Services";

const UserLookupExample = () => {
  const { userLookupManager } = useUsers();

  // State management
  const [email, setEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [result, setResult] = useState(null);

  // Handle user lookup
  const handleLookupUser = async () => {
    if (!email.trim()) {
      setError("Email address is required");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      const lookupResult = await userLookupManager.lookupUser(email.trim());
      setResult(lookupResult);
      setSuccess(`User found: ${lookupResult.user.email}`);
    } catch (err) {
      console.error("Lookup failed:", err);
      setError(err.message || "Failed to lookup user");
    } finally {
      setIsLoading(false);
    }
  };

  // Handle check if user exists
  const handleCheckExists = async () => {
    if (!email.trim()) {
      setError("Email address is required");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      const exists = await userLookupManager.userExists(email.trim());
      setSuccess(`User ${exists ? "exists" : "does not exist"} in the system`);
    } catch (err) {
      console.error("Existence check failed:", err);
      setError(err.message || "Failed to check user existence");
    } finally {
      setIsLoading(false);
    }
  };

  // Handle get public key
  const handleGetPublicKey = async () => {
    if (!email.trim()) {
      setError("Email address is required");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      const publicKeyResult = await userLookupManager.getUserPublicKey(
        email.trim(),
      );
      setResult({
        user: publicKeyResult,
        publicKeyLength: publicKeyResult.publicKey?.length || 0,
      });
      setSuccess(`Public key retrieved for: ${publicKeyResult.email}`);
    } catch (err) {
      console.error("Public key retrieval failed:", err);
      setError(err.message || "Failed to get public key");
    } finally {
      setIsLoading(false);
    }
  };

  // Clear results
  const handleClear = () => {
    setResult(null);
    setEmail("");
    setError("");
    setSuccess("");
  };

  return (
    <div style={{ padding: "20px", maxWidth: "800px", margin: "0 auto" }}>
      <h2>👥 User Public Key Lookup Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates user public key lookup for end-to-end encryption.
        <br />
        <strong>Note:</strong> This is a public API - no authentication
        required.
      </p>

      {/* User Lookup Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>🔍 Lookup User:</h4>
        <div style={{ display: "grid", gap: "15px" }}>
          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Email Address *
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Enter email address..."
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                fontSize: "14px",
              }}
              disabled={isLoading}
            />
          </div>

          <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
            <button
              onClick={handleLookupUser}
              disabled={isLoading || !email.trim()}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !email.trim() ? "#6c757d" : "#28a745",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: isLoading || !email.trim() ? "not-allowed" : "pointer",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {isLoading ? "🔄 Looking up..." : "🔍 Lookup User"}
            </button>

            <button
              onClick={handleCheckExists}
              disabled={isLoading || !email.trim()}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !email.trim() ? "#6c757d" : "#17a2b8",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: isLoading || !email.trim() ? "not-allowed" : "pointer",
                fontSize: "14px",
              }}
            >
              ❓ Check Exists
            </button>

            <button
              onClick={handleGetPublicKey}
              disabled={isLoading || !email.trim()}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !email.trim() ? "#6c757d" : "#6f42c1",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: isLoading || !email.trim() ? "not-allowed" : "pointer",
                fontSize: "14px",
              }}
            >
              🔑 Get Public Key
            </button>

            <button
              onClick={handleClear}
              disabled={isLoading}
              style={{
                padding: "12px 20px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: isLoading ? "not-allowed" : "pointer",
                fontSize: "14px",
              }}
            >
              🗑️ Clear
            </button>
          </div>
        </div>
      </div>

      {/* Success Message */}
      {success && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#d4edda",
            borderRadius: "6px",
            color: "#155724",
            border: "1px solid #c3e6cb",
          }}
        >
          ✅ {success}
        </div>
      )}

      {/* Error Message */}
      {error && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#f8d7da",
            borderRadius: "6px",
            color: "#721c24",
            border: "1px solid #f5c6cb",
          }}
        >
          ❌ {error}
        </div>
      )}

      {/* Result Display */}
      {result && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff3cd",
            borderRadius: "6px",
            border: "1px solid #ffeaa7",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>📋 Result:</h4>

          {result.user && (
            <div style={{ display: "grid", gap: "10px" }}>
              <div>
                <strong>👤 Name:</strong> {result.user.name || "N/A"}
              </div>
              <div>
                <strong>📧 Email:</strong> {result.user.email || "N/A"}
              </div>
              <div>
                <strong>🆔 User ID:</strong>{" "}
                {result.user.userId || result.user.user_id || "N/A"}
              </div>
              <div>
                <strong>🔑 Verification ID:</strong>{" "}
                {result.user.verificationId ||
                  result.user.verification_id ||
                  "N/A"}
              </div>
              {result.publicKeyLength > 0 && (
                <div>
                  <strong>🔐 Public Key Length:</strong>{" "}
                  {result.publicKeyLength} bytes
                </div>
              )}
            </div>
          )}

          <details style={{ marginTop: "15px" }}>
            <summary style={{ cursor: "pointer", fontWeight: "bold" }}>
              🔍 View Raw Data
            </summary>
            <pre
              style={{
                backgroundColor: "#f8f9fa",
                padding: "10px",
                borderRadius: "4px",
                fontSize: "11px",
                overflow: "auto",
                maxHeight: "200px",
                fontFamily: "monospace",
                marginTop: "10px",
              }}
            >
              {JSON.stringify(result, null, 2)}
            </pre>
          </details>
        </div>
      )}

      {/* Quick Test Section */}
      <div
        style={{
          padding: "15px",
          backgroundColor: "#e9ecef",
          borderRadius: "8px",
          border: "1px solid #dee2e6",
        }}
      >
        <h5 style={{ margin: "0 0 10px 0" }}>🚀 Quick Test</h5>
        <p style={{ margin: "0 0 10px 0", fontSize: "14px", color: "#666" }}>
          Test user lookup with sample email addresses:
        </p>
        <div
          style={{
            display: "flex",
            gap: "10px",
            alignItems: "center",
            flexWrap: "wrap",
          }}
        >
          <button
            onClick={() => setEmail("john.doe@example.com")}
            style={{
              padding: "5px 10px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "3px",
              cursor: "pointer",
              fontSize: "12px",
              fontFamily: "monospace",
            }}
          >
            john.doe@example.com
          </button>
          <button
            onClick={() => setEmail("alice@example.com")}
            style={{
              padding: "5px 10px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "3px",
              cursor: "pointer",
              fontSize: "12px",
              fontFamily: "monospace",
            }}
          >
            alice@example.com
          </button>
          <span style={{ fontSize: "12px", color: "#666" }}>
            Click to use sample emails for testing
          </span>
        </div>
      </div>
    </div>
  );
};

export default UserLookupExample;
