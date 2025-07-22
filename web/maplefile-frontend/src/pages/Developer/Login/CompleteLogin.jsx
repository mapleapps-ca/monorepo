// File: monorepo/web/maplefile-frontend/src/pages/Developer/Login/CompleteLogin.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../services/Services";

const DeveloperCompleteLogin = () => {
  const navigate = useNavigate();

  // Get services from the unified service system
  const services = useServices();
  const authManager = services?.authManager;
  const localStorageService = services?.localStorageService;
  const passwordStorageService = services?.passwordStorageService;

  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [decrypting, setDecrypting] = useState(false);
  const [error, setError] = useState("");
  const [verifyData, setVerifyData] = useState(null);
  const [decryptionProgress, setDecryptionProgress] = useState("");
  const [servicesReady, setServicesReady] = useState(false);
  const [debugInfo, setDebugInfo] = useState({});

  // FIXED: Stable service debugging that won't cause infinite loops
  const checkServices = useCallback(() => {
    const debug = {
      authManagerExists: !!authManager,
      passwordStorageServiceExists: !!passwordStorageService,
      localStorageServiceExists: !!localStorageService,
      windowPasswordService: !!window.__passwordService,
      windowMapleServices: !!window.mapleAppsServices?.passwordStorageService,
      allServicesKeys: services ? Object.keys(services) : [],
      timestamp: new Date().toISOString(),
    };

    // Only update if something actually changed
    setDebugInfo((prevDebug) => {
      const changed = JSON.stringify(prevDebug) !== JSON.stringify(debug);
      if (changed) {
        console.log("[CompleteLogin] Service status changed:", debug);
      }
      return changed ? debug : prevDebug;
    });

    const ready = !!(authManager && passwordStorageService);
    setServicesReady((prevReady) => {
      if (prevReady !== ready) {
        console.log("[CompleteLogin] Services ready status changed:", ready);
      }
      return ready;
    });

    return debug;
  }, [authManager, passwordStorageService, localStorageService, services]);

  // FIXED: Stable effect that only runs when actual services change
  useEffect(() => {
    checkServices();
  }, [checkServices]);

  // FIXED: Initial service check with delayed retries
  useEffect(() => {
    console.log("[CompleteLogin] Initial service check");

    // Log what we got from useServices immediately
    console.log("[CompleteLogin] Services from useServices:", {
      total: services ? Object.keys(services).length : 0,
      authManager: !!authManager,
      passwordStorageService: !!passwordStorageService,
      localStorageService: !!localStorageService,
    });

    // Check alternative sources
    if (!passwordStorageService) {
      console.log(
        "[CompleteLogin] Looking for alternative passwordStorageService sources...",
      );

      if (window.__passwordService) {
        console.log("[CompleteLogin] Found window.__passwordService");
      }

      if (window.mapleAppsServices?.passwordStorageService) {
        console.log(
          "[CompleteLogin] Found window.mapleAppsServices.passwordStorageService",
        );
      }
    }

    // Retry after delays for services that might still be initializing
    const timer1 = setTimeout(() => {
      console.log("[CompleteLogin] Retry 1 - checking services again...");
      checkServices();
    }, 1000);

    const timer2 = setTimeout(() => {
      console.log("[CompleteLogin] Retry 2 - final check...");
      checkServices();
    }, 3000);

    return () => {
      clearTimeout(timer1);
      clearTimeout(timer2);
    };
  }, []); // Empty dependency array - only run once on mount

  // Load email and verify data (same as before but stable dependencies)
  useEffect(() => {
    let storedEmail = null;
    let storedVerifyData = null;

    // Try to get email from multiple sources
    if (authManager && typeof authManager.getCurrentUserEmail === "function") {
      try {
        storedEmail = authManager.getCurrentUserEmail();
        if (storedEmail) {
          console.log(
            "[CompleteLogin] Using email from authManager:",
            storedEmail,
          );
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] Could not get email from authManager:",
          err,
        );
      }
    }

    if (
      !storedEmail &&
      localStorageService &&
      typeof localStorageService.getUserEmail === "function"
    ) {
      try {
        storedEmail = localStorageService.getUserEmail();
        if (storedEmail) {
          console.log(
            "[CompleteLogin] Using email from localStorageService:",
            storedEmail,
          );
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] Could not get email from localStorageService:",
          err,
        );
      }
    }

    if (!storedEmail) {
      storedEmail = sessionStorage.getItem("loginEmail");
      if (storedEmail) {
        console.log(
          "[CompleteLogin] Using email from sessionStorage:",
          storedEmail,
        );
      }
    }

    // Try to get verify data
    if (
      localStorageService &&
      typeof localStorageService.getLoginSessionData === "function"
    ) {
      try {
        storedVerifyData =
          localStorageService.getLoginSessionData("verify_response");
        if (storedVerifyData) {
          console.log(
            "[CompleteLogin] Using verify data from localStorageService",
          );
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] Could not get verify data from localStorageService:",
          err,
        );
      }
    }

    if (!storedVerifyData) {
      try {
        const sessionVerifyData = sessionStorage.getItem(
          "otpVerificationResult",
        );
        if (sessionVerifyData) {
          storedVerifyData = JSON.parse(sessionVerifyData);
          console.log("[CompleteLogin] Using verify data from sessionStorage");
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] Could not parse verify data from sessionStorage:",
          err,
        );
      }
    }

    if (storedEmail && storedVerifyData) {
      setEmail(storedEmail);
      setVerifyData(storedVerifyData);
    } else {
      console.error(
        "[CompleteLogin] Missing email or verify data, redirecting to start",
      );
      navigate("/developer/login/request-ott");
    }
  }, [navigate, authManager, localStorageService]); // Stable dependencies

  // FIXED: Better token validation function
  const validateTokensAfterLogin = async () => {
    console.log("[CompleteLogin] Starting token validation...");

    // Method 1: Use AuthManager authentication check (preferred)
    if (authManager && typeof authManager.isAuthenticated === "function") {
      try {
        const isAuthenticated = authManager.isAuthenticated();
        console.log(
          "[CompleteLogin] AuthManager.isAuthenticated():",
          isAuthenticated,
        );

        if (isAuthenticated) {
          console.log(
            "[CompleteLogin] ✅ Authentication confirmed via AuthManager",
          );
          return true;
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] AuthManager.isAuthenticated() failed:",
          err,
        );
      }
    }

    // Method 2: Check if AuthManager can make requests
    if (
      authManager &&
      typeof authManager.canMakeAuthenticatedRequests === "function"
    ) {
      try {
        const canMakeRequests = authManager.canMakeAuthenticatedRequests();
        console.log(
          "[CompleteLogin] AuthManager.canMakeAuthenticatedRequests():",
          canMakeRequests,
        );

        if (canMakeRequests) {
          console.log(
            "[CompleteLogin] ✅ Can make authenticated requests via AuthManager",
          );
          return true;
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] AuthManager.canMakeAuthenticatedRequests() failed:",
          err,
        );
      }
    }

    // Method 3: Direct localStorage check via localStorageService
    if (localStorageService) {
      try {
        const accessToken = localStorageService.getAccessToken?.();
        const refreshToken = localStorageService.getRefreshToken?.();
        console.log("[CompleteLogin] Direct token check:", {
          hasAccessToken: !!accessToken,
          hasRefreshToken: !!refreshToken,
          accessTokenLength: accessToken ? accessToken.length : 0,
          refreshTokenLength: refreshToken ? refreshToken.length : 0,
        });

        if (accessToken && refreshToken) {
          console.log(
            "[CompleteLogin] ✅ Tokens found via localStorageService",
          );
          return true;
        }
      } catch (err) {
        console.warn(
          "[CompleteLogin] LocalStorageService token check failed:",
          err,
        );
      }
    }

    // Method 4: Direct localStorage check (fallback)
    try {
      const accessToken = localStorage.getItem("mapleapps_access_token");
      const refreshToken = localStorage.getItem("mapleapps_refresh_token");
      console.log("[CompleteLogin] Direct localStorage check:", {
        hasAccessToken: !!accessToken,
        hasRefreshToken: !!refreshToken,
        accessTokenLength: accessToken ? accessToken.length : 0,
        refreshTokenLength: refreshToken ? refreshToken.length : 0,
      });

      if (accessToken && refreshToken) {
        console.log(
          "[CompleteLogin] ✅ Tokens found in direct localStorage check",
        );
        return true;
      }
    } catch (err) {
      console.warn("[CompleteLogin] Direct localStorage check failed:", err);
    }

    console.error("[CompleteLogin] ❌ No valid tokens found via any method");
    return false;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    // Find working password service
    let workingPasswordService =
      passwordStorageService ||
      window.mapleAppsServices?.passwordStorageService ||
      window.__passwordService;

    if (!workingPasswordService) {
      setError(
        "Password storage service is not available. Please refresh the page and try again.",
      );
      return;
    }

    if (!authManager) {
      setError(
        "Authentication service not available. Please refresh the page.",
      );
      return;
    }

    setLoading(true);
    setDecrypting(true);
    setError("");

    try {
      if (!password) {
        throw new Error("Password is required");
      }

      if (!verifyData || !verifyData.challengeId) {
        throw new Error(
          "Missing challenge data. Please start the login process again.",
        );
      }

      console.log("[CompleteLogin] Starting login completion...");

      setDecryptionProgress("Initializing cryptographic libraries...");
      await new Promise((resolve) => setTimeout(resolve, 500));

      setDecryptionProgress("Deriving encryption key from password...");
      await new Promise((resolve) => setTimeout(resolve, 300));

      setDecryptionProgress("Decrypting challenge data...");

      const decryptedChallenge = await authManager.decryptChallenge(
        password,
        verifyData,
      );

      setDecryptionProgress("Completing authentication...");
      setDecrypting(false);

      console.log("[CompleteLogin] Calling authManager.completeLogin...");
      const response = await authManager.completeLogin(
        email,
        verifyData.challengeId,
        decryptedChallenge,
      );
      console.log("[CompleteLogin] Login completed successfully!", response);

      // FIXED: Better token validation with retry mechanism
      setDecryptionProgress("Validating authentication tokens...");

      let tokenValidationAttempts = 0;
      const maxAttempts = 5;
      let tokensValid = false;

      while (tokenValidationAttempts < maxAttempts && !tokensValid) {
        tokenValidationAttempts++;
        console.log(
          `[CompleteLogin] Token validation attempt ${tokenValidationAttempts}/${maxAttempts}`,
        );

        // Wait a bit longer for tokens to be fully stored
        if (tokenValidationAttempts > 1) {
          await new Promise((resolve) =>
            setTimeout(resolve, 200 * tokenValidationAttempts),
          );
        }

        tokensValid = await validateTokensAfterLogin();

        if (!tokensValid && tokenValidationAttempts < maxAttempts) {
          console.log(
            `[CompleteLogin] Tokens not ready, retrying in ${200 * tokenValidationAttempts}ms...`,
          );
        }
      }

      if (tokensValid) {
        console.log(
          "[CompleteLogin] ✅ Authentication successful - storing password and navigating",
        );
        workingPasswordService.setPassword(password);
        console.log(
          "[CompleteLogin] Password stored in PasswordStorageService",
        );
        navigate("/developer/dashboard", { replace: true });
      } else {
        console.error(
          "[CompleteLogin] ❌ Token validation failed after all attempts",
        );

        // Additional debugging - check what's actually in localStorage
        const allKeys = Object.keys(localStorage).filter((key) =>
          key.includes("mapleapps"),
        );
        console.log(
          "[CompleteLogin] All MapleApps localStorage keys:",
          allKeys,
        );

        allKeys.forEach((key) => {
          const value = localStorage.getItem(key);
          console.log(
            `[CompleteLogin] ${key}:`,
            value ? `${value.substring(0, 50)}...` : "null",
          );
        });

        throw new Error(
          "Authentication completed but tokens could not be validated. " +
            "This may be a temporary issue - please try logging in again.",
        );
      }
    } catch (error) {
      console.error("[CompleteLogin] Login failed:", error);
      setError(error.message);
      setDecrypting(false);
      setDecryptionProgress("");
    } finally {
      setLoading(false);
    }
  };

  const handleBackToVerify = () => {
    navigate("/developer/login/verify-ott");
  };

  if (!verifyData) {
    return (
      <div style={{ padding: "20px", maxWidth: "600px", margin: "0 auto" }}>
        <h2>Loading...</h2>
        <p>Loading verification data...</p>
      </div>
    );
  }

  // Check all possible password service sources
  const hasPasswordService = !!(
    passwordStorageService ||
    window.mapleAppsServices?.passwordStorageService ||
    window.__passwordService
  );

  const canSubmit = !loading && password && authManager && hasPasswordService;

  return (
    <div style={{ padding: "20px", maxWidth: "600px", margin: "0 auto" }}>
      <h2>Step 3: Complete Login</h2>
      <p>
        Enter your password to complete login for <strong>{email}</strong>
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

      {decrypting && decryptionProgress && (
        <div
          style={{
            color: "#1976d2",
            backgroundColor: "#e3f2fd",
            padding: "10px",
            borderRadius: "4px",
            marginBottom: "15px",
          }}
        >
          🔐 {decryptionProgress}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "20px" }}>
          <label htmlFor="password">Password</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Enter your password"
            required
            disabled={loading}
            autoComplete="current-password"
            style={{
              width: "100%",
              padding: "8px",
              marginTop: "5px",
              border: "1px solid #ccc",
              borderRadius: "4px",
            }}
          />
          <div style={{ fontSize: "12px", color: "#666", marginTop: "5px" }}>
            Your password will be used to decrypt your secure keys locally
          </div>
        </div>

        {/* ENHANCED Service status display */}
        <div
          style={{
            fontSize: "12px",
            color: "#666",
            marginBottom: "15px",
            padding: "8px",
            backgroundColor: "#f5f5f5",
            borderRadius: "4px",
          }}
        >
          <strong>Service Status:</strong>
          <br />
          Password Length: {password.length}
          <br />
          AuthManager: {authManager ? "✅ Available" : "❌ Missing"}
          <br />
          LocalStorageService:{" "}
          {localStorageService ? "✅ Available" : "❌ Missing"}
          <br />
          PasswordStorageService (useServices):{" "}
          {passwordStorageService ? "✅ Available" : "❌ Missing"}
          <br />
          PasswordStorageService (window.__passwordService):{" "}
          {window.__passwordService ? "✅ Available" : "❌ Missing"}
          <br />
          PasswordStorageService (window.mapleAppsServices):{" "}
          {window.mapleAppsServices?.passwordStorageService
            ? "✅ Available"
            : "❌ Missing"}
          <br />
          Any Password Service:{" "}
          {hasPasswordService ? "✅ Available" : "❌ Missing"}
          <br />
          Can Submit: {canSubmit ? "✅ Yes" : "❌ No"}
        </div>

        <div style={{ display: "flex", gap: "10px", marginBottom: "20px" }}>
          <button
            type="submit"
            disabled={!canSubmit}
            style={{
              flex: 1,
              padding: "10px",
              backgroundColor: canSubmit ? "#1976d2" : "#ccc",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: canSubmit ? "pointer" : "not-allowed",
            }}
          >
            {loading
              ? decrypting
                ? "Decrypting..."
                : "Completing Login..."
              : "Complete Login"}
          </button>

          <button
            type="button"
            onClick={handleBackToVerify}
            disabled={loading}
            style={{
              padding: "10px 15px",
              backgroundColor: "transparent",
              color: loading ? "#ccc" : "#666",
              border: `1px solid ${loading ? "#ccc" : "#666"}`,
              borderRadius: "4px",
              cursor: loading ? "not-allowed" : "pointer",
            }}
          >
            Back
          </button>
        </div>

        {/* Warning if services missing */}
        {!hasPasswordService && (
          <div
            style={{
              padding: "10px",
              backgroundColor: "#fff3cd",
              border: "1px solid #ffeaa7",
              borderRadius: "4px",
              marginBottom: "15px",
              fontSize: "14px",
            }}
          >
            <strong>⚠️ Password storage service not found</strong>
            <p>
              The password storage service could not be found in any of the
              expected locations. Please refresh the page to reinitialize
              services.
            </p>
          </div>
        )}
      </form>

      {/* Debug info (development only) */}
      {import.meta.env.DEV && (
        <div
          style={{
            marginTop: "20px",
            padding: "10px",
            backgroundColor: "#f0f0f0",
            borderRadius: "4px",
            fontSize: "11px",
            fontFamily: "monospace",
          }}
        >
          <strong>Debug Info:</strong>
          <br />
          Services Available: {debugInfo.allServicesKeys?.join(", ")}
          <br />
          AuthManager Exists: {debugInfo.authManagerExists ? "Yes" : "No"}
          <br />
          LocalStorageService Exists:{" "}
          {debugInfo.localStorageServiceExists ? "Yes" : "No"}
          <br />
          PasswordStorageService Exists:{" "}
          {debugInfo.passwordStorageServiceExists ? "Yes" : "No"}
          <br />
          Window.__passwordService:{" "}
          {debugInfo.windowPasswordService ? "Yes" : "No"}
          <br />
          Window.mapleAppsServices:{" "}
          {debugInfo.windowMapleServices ? "Yes" : "No"}
          <br />
          Last Check: {debugInfo.timestamp}
          <br />
          {/* ENHANCED: Show current localStorage state */}
          <br />
          <strong>Current Token State:</strong>
          <br />
          Access Token in localStorage:{" "}
          {localStorage.getItem("mapleapps_access_token") ? "Yes" : "No"}
          <br />
          Refresh Token in localStorage:{" "}
          {localStorage.getItem("mapleapps_refresh_token") ? "Yes" : "No"}
          <br />
          User Email in localStorage:{" "}
          {localStorage.getItem("mapleapps_user_email") ? "Yes" : "No"}
        </div>
      )}
    </div>
  );
};

export default DeveloperCompleteLogin;
