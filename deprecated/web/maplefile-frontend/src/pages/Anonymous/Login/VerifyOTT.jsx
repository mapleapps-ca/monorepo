// File: monorepo/web/maplefile-frontend/src/pages/Anonymous/Login/VerifyOTT.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../services/Services";

const VerifyOTT = () => {
  const navigate = useNavigate();
  const { authManager, localStorageService } = useServices();
  const [ott, setOtt] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    // FIXED: Try multiple sources for email
    let storedEmail = null;

    // Try 1: Get from authManager
    if (authManager && typeof authManager.getCurrentUserEmail === "function") {
      try {
        storedEmail = authManager.getCurrentUserEmail();
        if (storedEmail) {
          console.log("[VerifyOTT] Using email from authManager:", storedEmail);
        }
      } catch (err) {
        console.warn("[VerifyOTT] Could not get email from authManager:", err);
      }
    }

    // Try 2: Get from localStorageService
    if (!storedEmail && localStorageService) {
      try {
        storedEmail = localStorageService.getUserEmail();
        if (storedEmail) {
          console.log(
            "[VerifyOTT] Using email from localStorageService:",
            storedEmail,
          );
        }
      } catch (err) {
        console.warn(
          "[VerifyOTT] Could not get email from localStorageService:",
          err,
        );
      }
    }

    // Try 3: Get from sessionStorage (fallback)
    if (!storedEmail) {
      storedEmail = sessionStorage.getItem("loginEmail");
      if (storedEmail) {
        console.log(
          "[VerifyOTT] Using email from sessionStorage:",
          storedEmail,
        );
      }
    }

    if (storedEmail) {
      setEmail(storedEmail);
    } else {
      // If no email found anywhere, redirect to first step
      console.log(
        "[VerifyOTT] No stored email found in any location, redirecting to start",
      );
      navigate("/login/request-ott");
    }
  }, [navigate, authManager, localStorageService]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      // Validate services are available
      if (!authManager) {
        throw new Error(
          "Authentication service not available. Please refresh the page.",
        );
      }

      // Validate OTT
      if (!ott) {
        throw new Error("Verification code is required");
      }

      if (ott.length !== 6 || !/^\d{6}$/.test(ott)) {
        throw new Error("Verification code must be 6 digits");
      }

      if (!email || !email.trim()) {
        throw new Error("Email address is required");
      }

      if (!email.includes("@")) {
        throw new Error("Please enter a valid email address");
      }

      console.log(
        "[VerifyOTT] Verifying OTT via AuthManager for:",
        email.trim(),
      );
      console.log(
        "[VerifyOTT] Available authManager methods:",
        Object.getOwnPropertyNames(Object.getPrototypeOf(authManager)),
      );

      const trimmedEmail = email.trim().toLowerCase();

      // FIXED: Try different possible method names and validate responses
      let response;

      if (typeof authManager.verifyOTP === "function") {
        console.log("[VerifyOTT] Using verifyOTP method");
        response = await authManager.verifyOTP({
          email: trimmedEmail,
          otp_code: ott,
        });
      } else if (typeof authManager.verifyOTT === "function") {
        console.log("[VerifyOTT] Using verifyOTT method");
        response = await authManager.verifyOTT(trimmedEmail, ott);
      } else if (typeof authManager.verify === "function") {
        console.log("[VerifyOTT] Using verify method");
        response = await authManager.verify({
          email: trimmedEmail,
          code: ott,
        });
      } else {
        // List available methods for debugging
        const methods = Object.getOwnPropertyNames(
          Object.getPrototypeOf(authManager),
        ).filter(
          (name) =>
            typeof authManager[name] === "function" && !name.startsWith("_"),
        );

        console.error("[VerifyOTT] Available authManager methods:", methods);
        throw new Error(
          `OTP verification method not found. Available methods: ${methods.join(", ")}`,
        );
      }

      // Validate that we got a proper response
      if (!response) {
        throw new Error("Verification method returned no response");
      }

      console.log("[VerifyOTT] Raw verification response:", response);

      // Check if response indicates success
      if (response.error) {
        throw new Error(response.error);
      }

      // If response is just a success message, we might need additional data
      if (
        typeof response === "string" ||
        (response.success && !response.challengeId)
      ) {
        console.warn(
          "[VerifyOTT] Response may not contain challenge data:",
          response,
        );
        // Try to get additional session data if available
        if (
          localStorageService &&
          typeof localStorageService.getLoginSessionData === "function"
        ) {
          try {
            const additionalData =
              localStorageService.getLoginSessionData("challenge_data");
            if (additionalData) {
              response = { ...response, ...additionalData };
              console.log("[VerifyOTT] Merged with additional session data");
            }
          } catch (err) {
            console.warn(
              "[VerifyOTT] Could not get additional session data:",
              err,
            );
          }
        }
      }

      console.log(
        "[VerifyOTT] Verification successful via AuthManager, received encrypted keys",
      );
      console.log("[VerifyOTT] Verification response:", response);

      // FIXED: Store the verification response data for CompleteLogin to use
      try {
        // Store in sessionStorage as primary method
        sessionStorage.setItem(
          "otpVerificationResult",
          JSON.stringify(response),
        );
        console.log(
          "[VerifyOTT] Verification response stored in sessionStorage",
        );

        // Also try to store via localStorageService if available
        if (
          localStorageService &&
          typeof localStorageService.setLoginSessionData === "function"
        ) {
          try {
            localStorageService.setLoginSessionData(
              "verify_response",
              response,
            );
            console.log(
              "[VerifyOTT] Verification response stored via localStorageService",
            );
          } catch (storageError) {
            console.warn(
              "[VerifyOTT] Could not store via localStorageService:",
              storageError,
            );
          }
        }

        // Store email as well to ensure it's available
        sessionStorage.setItem("loginEmail", trimmedEmail);

        // Also try to store email via authManager if available
        if (
          authManager &&
          typeof authManager.setCurrentUserEmail === "function"
        ) {
          try {
            authManager.setCurrentUserEmail(trimmedEmail);
            console.log("[VerifyOTT] Email stored via authManager");
          } catch (emailError) {
            console.warn(
              "[VerifyOTT] Could not store email via authManager:",
              emailError,
            );
          }
        }
      } catch (storageError) {
        console.error(
          "[VerifyOTT] Failed to store verification data:",
          storageError,
        );
        throw new Error(
          "Verification successful but failed to store data. Please try again.",
        );
      }

      // Navigate to complete login step
      navigate("/login/complete");
    } catch (error) {
      console.error("[VerifyOTT] Verification failed via AuthManager:", error);
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleResendCode = async () => {
    if (!authManager) {
      setError(
        "Authentication service not available. Please refresh the page.",
      );
      return;
    }

    if (!email || !email.trim()) {
      setError("Email address is required to resend code");
      return;
    }

    setLoading(true);
    setError("");

    try {
      console.log("[VerifyOTT] Resending OTT code...");

      const trimmedEmail = email.trim().toLowerCase();

      // Try different possible method names for requesting OTT
      if (typeof authManager.requestOTT === "function") {
        await authManager.requestOTT(trimmedEmail);
      } else if (typeof authManager.requestOTP === "function") {
        await authManager.requestOTP({ email: trimmedEmail });
      } else {
        throw new Error("Request OTT method not found on authManager");
      }

      console.log("[VerifyOTT] Code resent successfully");
      setError(""); // Clear any previous errors
      // You might want to show a success message here
    } catch (err) {
      console.error("[VerifyOTT] Failed to resend code:", err);
      setError(err.message || "Failed to resend code. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleBackToEmail = () => {
    navigate("/login/request-ott");
  };

  return (
    <div style={{ padding: "20px", maxWidth: "400px", margin: "0 auto" }}>
      <h2>Step 2: Verify Code</h2>
      <p>
        Enter your email address and the 6-digit verification code sent to your
        email.
      </p>

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

      <form onSubmit={handleSubmit}>
        {/* Email field - editable if not auto-populated */}
        <div style={{ marginBottom: "15px" }}>
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
          <div style={{ fontSize: "12px", color: "#666", marginTop: "5px" }}>
            {email
              ? "Verification code will be sent to this email"
              : "Enter the email you used in step 1"}
          </div>
        </div>

        <div style={{ marginBottom: "15px" }}>
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
            style={{
              width: "100%",
              padding: "8px",
              marginTop: "5px",
              border: "1px solid #ccc",
              borderRadius: "4px",
              fontSize: "18px",
              textAlign: "center",
              letterSpacing: "2px",
            }}
          />
          <div style={{ fontSize: "12px", color: "#666", marginTop: "5px" }}>
            Enter the 6-digit code sent to your email
          </div>
        </div>

        <div style={{ display: "flex", gap: "10px", marginBottom: "15px" }}>
          <button
            type="submit"
            disabled={
              loading || ott.length !== 6 || !email.trim() || !authManager
            }
            style={{
              flex: 1,
              padding: "10px",
              backgroundColor:
                loading || ott.length !== 6 || !email.trim() || !authManager
                  ? "#ccc"
                  : "#1976d2",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                loading || ott.length !== 6 || !email.trim() || !authManager
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            {loading ? "Verifying..." : "Verify Code"}
          </button>

          <button
            type="button"
            onClick={handleResendCode}
            disabled={loading || !authManager}
            style={{
              padding: "10px 15px",
              backgroundColor: "transparent",
              color: loading || !authManager ? "#ccc" : "#1976d2",
              border: `1px solid ${loading || !authManager ? "#ccc" : "#1976d2"}`,
              borderRadius: "4px",
              cursor: loading || !authManager ? "not-allowed" : "pointer",
            }}
          >
            Resend
          </button>
        </div>

        <button
          type="button"
          onClick={handleBackToEmail}
          disabled={loading}
          style={{
            width: "100%",
            padding: "10px",
            backgroundColor: "transparent",
            color: "#666",
            border: "1px solid #ccc",
            borderRadius: "4px",
            cursor: loading ? "not-allowed" : "pointer",
          }}
        >
          ← Change Email
        </button>
      </form>

      <div style={{ marginTop: "20px", fontSize: "14px", color: "#666" }}>
        <h3>Didn't receive the code?</h3>
        <ul>
          <li>Check your spam/junk folder</li>
          <li>Make sure the email address above is correct</li>
          <li>The code expires in 10 minutes</li>
          <li>Use the "Resend" button to get a new code</li>
          <li>Use the "Change Email" button to go back and start over</li>
          <li>AuthManager will handle the verification orchestration</li>
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
          LocalStorageService Available: {localStorageService ? "Yes" : "No"}
          <br />
          Email State: {email}
          <br />
          SessionStorage Email: {sessionStorage.getItem("loginEmail")}
          <br />
          SessionStorage Verify Data:{" "}
          {sessionStorage.getItem("otpVerificationResult") ? "Yes" : "No"}
          <br />
          OTT Length: {ott.length}
          <br />
          {authManager && (
            <>
              AuthManager Email:{" "}
              {(() => {
                try {
                  return authManager.getCurrentUserEmail
                    ? authManager.getCurrentUserEmail()
                    : "No method";
                } catch (e) {
                  return "Error: " + e.message;
                }
              })()}
              <br />
            </>
          )}
          {localStorageService && (
            <>
              LocalStorage Email:{" "}
              {(() => {
                try {
                  return localStorageService.getUserEmail
                    ? localStorageService.getUserEmail()
                    : "No method";
                } catch (e) {
                  return "Error: " + e.message;
                }
              })()}
              <br />
            </>
          )}
          {authManager && (
            <>
              AuthManager Methods:{" "}
              {Object.getOwnPropertyNames(Object.getPrototypeOf(authManager))
                .filter(
                  (name) =>
                    typeof authManager[name] === "function" &&
                    !name.startsWith("_"),
                )
                .join(", ")}
              <br />
              Has verifyOTP:{" "}
              {typeof authManager.verifyOTP === "function" ? "Yes" : "No"}
              <br />
              Has verifyOTT:{" "}
              {typeof authManager.verifyOTT === "function" ? "Yes" : "No"}
              <br />
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default VerifyOTT;
