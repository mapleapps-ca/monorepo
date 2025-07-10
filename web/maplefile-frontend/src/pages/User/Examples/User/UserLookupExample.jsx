// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/User/UserLookupExample.jsx
// Simplified example component demonstrating user lookup functionality

import React, { useState } from "react";
import { useUsers } from "../../../../services/Services";

const UserLookupExample = () => {
  const {
    // State
    isLoading,
    error,
    success,

    // Operations
    lookupUser,
    userExists,
    getUserPublicKey,

    // Utilities
    clearMessages,
    validateEmail,
    sanitizeEmail,
    isAvailable,
  } = useUsers();

  // Form state
  const [email, setEmail] = useState("");
  const [result, setResult] = useState(null);

  // Handle user lookup
  const handleLookupUser = async () => {
    if (!email.trim()) {
      alert("Email address is required");
      return;
    }

    const sanitizedEmail = sanitizeEmail(email.trim());

    if (!validateEmail(sanitizedEmail)) {
      alert("Please enter a valid email address");
      return;
    }

    try {
      const lookupResult = await lookupUser(sanitizedEmail);
      setResult(lookupResult);
    } catch (err) {
      console.error("Lookup failed:", err);
      // Error is already set by the hook
    }
  };

  // Handle check if user exists
  const handleCheckExists = async () => {
    if (!email.trim()) {
      alert("Email address is required");
      return;
    }

    const sanitizedEmail = sanitizeEmail(email.trim());

    if (!validateEmail(sanitizedEmail)) {
      alert("Please enter a valid email address");
      return;
    }

    try {
      const exists = await userExists(sanitizedEmail);
      alert(`User ${exists ? "exists" : "does not exist"} in the system`);
    } catch (err) {
      console.error("Existence check failed:", err);
      // Error is already set by the hook
    }
  };

  // Handle get public key
  const handleGetPublicKey = async () => {
    if (!email.trim()) {
      alert("Email address is required");
      return;
    }

    const sanitizedEmail = sanitizeEmail(email.trim());

    if (!validateEmail(sanitizedEmail)) {
      alert("Please enter a valid email address");
      return;
    }

    try {
      const publicKeyResult = await getUserPublicKey(sanitizedEmail);
      setResult({
        user: {
          ...publicKeyResult,
          email: publicKeyResult.email,
          user_id: publicKeyResult.userId,
          name: publicKeyResult.name,
          verification_id: publicKeyResult.verificationId,
        },
        source: publicKeyResult.source,
        publicKeyLength: publicKeyResult.publicKey.length,
      });
    } catch (err) {
      console.error("Public key retrieval failed:", err);
      // Error is already set by the hook
    }
  };

  // Clear results
  const handleClear = () => {
    setResult(null);
    setEmail("");
    clearMessages();
  };

  return (
    <div style={{ padding: "20px", maxWidth: "800px", margin: "0 auto" }}>
      <h2>👥 User Public Key Lookup Example (Simplified)</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates user public key lookup for end-to-end encryption.
        <br />
        <strong>Note:</strong> This is a public API - no authentication
        required.
      </p>

      {/* Service Status */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: isAvailable ? "#d4edda" : "#f8d7da",
          borderRadius: "6px",
          border: `1px solid ${isAvailable ? "#c3e6cb" : "#f5c6cb"}`,
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>🔧 Service Status:</h4>
        <div style={{ fontSize: "14px" }}>
          <div>
            <strong>UserLookupManager Available:</strong>{" "}
            {isAvailable ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Is Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Current Email:</strong> {email || "Empty"}
          </div>
          <div>
            <strong>Valid Email:</strong>{" "}
            {email
              ? validateEmail(sanitizeEmail(email))
                ? "✅"
                : "❌"
              : "N/A"}
          </div>
        </div>
      </div>

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
              disabled={isLoading || !email.trim() || !isAvailable}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !email.trim() || !isAvailable
                    ? "#6c757d"
                    : "#28a745",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !email.trim() || !isAvailable
                    ? "not-allowed"
                    : "pointer",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {isLoading ? "🔄 Looking up..." : "🔍 Lookup User"}
            </button>

            <button
              onClick={handleCheckExists}
              disabled={isLoading || !email.trim() || !isAvailable}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !email.trim() || !isAvailable
                    ? "#6c757d"
                    : "#17a2b8",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !email.trim() || !isAvailable
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              ❓ Check Exists
            </button>

            <button
              onClick={handleGetPublicKey}
              disabled={isLoading || !email.trim() || !isAvailable}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !email.trim() || !isAvailable
                    ? "#6c757d"
                    : "#6f42c1",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !email.trim() || !isAvailable
                    ? "not-allowed"
                    : "pointer",
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
                backgroundColor: isLoading ? "#6c757d" : "#6c757d",
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
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>✅ {success}</span>
          <button
            onClick={clearMessages}
            style={{
              background: "none",
              border: "none",
              color: "#155724",
              cursor: "pointer",
              fontSize: "16px",
            }}
          >
            ✕
          </button>
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
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>❌ {error}</span>
          <button
            onClick={clearMessages}
            style={{
              background: "none",
              border: "none",
              color: "#721c24",
              cursor: "pointer",
              fontSize: "16px",
            }}
          >
            ✕
          </button>
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
                <strong>🆔 User ID:</strong> {result.user.user_id || "N/A"}
              </div>
              <div>
                <strong>🔑 Verification ID:</strong>{" "}
                {result.user.verification_id || "N/A"}
              </div>
              <div>
                <strong>📡 Source:</strong>{" "}
                <span
                  style={{
                    backgroundColor:
                      result.source === "cache" ? "#fff3cd" : "#d1ecf1",
                    padding: "2px 6px",
                    borderRadius: "3px",
                  }}
                >
                  {result.source || "unknown"}
                </span>
              </div>
              {result.publicKeyLength && (
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
