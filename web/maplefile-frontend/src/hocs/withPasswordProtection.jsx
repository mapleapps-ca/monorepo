// File: monorepo/web/maplefile-frontend/src/hocs/withPasswordProtection.jsx
// HOC that protects components by checking if password is available - Updated without worker dependencies
import React, { useEffect, useState } from "react";
import { useNavigate, useLocation } from "react-router";
import passwordStorageService from "../services/PasswordStorageService.js";

/**
 * HOC that protects components by checking if password is available
 * Updated: Simplified without worker dependencies
 */
const withPasswordProtection = (WrappedComponent, options = {}) => {
  const {
    redirectTo = "/login",
    showLoadingWhileChecking = true,
    checkInterval = null,
    customMessage = "Password required. Redirecting to login...",
  } = options;

  const PasswordProtectedComponent = (props) => {
    const navigate = useNavigate();
    const location = useLocation();
    const [isChecking, setIsChecking] = useState(true);
    const [hasPassword, setHasPassword] = useState(false);

    useEffect(() => {
      let mounted = true;
      let intervalId = null;

      const checkPasswordAsync = async () => {
        console.log(
          `[withPasswordProtection] Starting password check for ${location.pathname}`,
        );

        // Wait a bit for service initialization if needed
        if (!passwordStorageService.isInitialized) {
          console.log(
            "[withPasswordProtection] Waiting for service initialization...",
          );
          await new Promise((resolve) => setTimeout(resolve, 100));
        }

        // In development with localStorage, explicitly try to restore
        const storageInfo = passwordStorageService.getStorageInfo();
        if (storageInfo.isDevelopment && storageInfo.mode === "localStorage") {
          console.log(
            "[withPasswordProtection] Dev mode detected, attempting restore...",
          );

          // Give the service a chance to restore from localStorage
          if (!passwordStorageService.password) {
            try {
              const restored =
                await passwordStorageService.restorePasswordFromStorage();
              console.log(
                "[withPasswordProtection] Restore attempt result:",
                restored,
              );
            } catch (error) {
              console.error("[withPasswordProtection] Restore error:", error);
            }
          }
        }

        // Now check for password
        const passwordAvailable = passwordStorageService.hasPassword();
        console.log(
          `[withPasswordProtection] Password check result for ${location.pathname}:`,
          passwordAvailable,
        );

        if (!mounted) return false;

        setHasPassword(passwordAvailable);

        if (!passwordAvailable) {
          console.log(
            `[withPasswordProtection] No password found, redirecting to ${redirectTo}`,
          );
          navigate(redirectTo, {
            state: {
              from: location,
              message: customMessage,
            },
            replace: true,
          });
          return false;
        }

        return true;
      };

      // Initial check
      checkPasswordAsync().then((hasPass) => {
        if (!mounted) return;

        setIsChecking(false);

        // Set up periodic checking if specified and password exists
        if (checkInterval && hasPass) {
          intervalId = setInterval(() => {
            const stillHasPassword = passwordStorageService.hasPassword();
            if (!stillHasPassword && mounted) {
              console.log(
                "[withPasswordProtection] Password lost, redirecting...",
              );
              navigate(redirectTo, {
                state: { from: location, message: "Session expired" },
                replace: true,
              });
            }
          }, checkInterval);
        }
      });

      return () => {
        mounted = false;
        if (intervalId) {
          clearInterval(intervalId);
        }
      };
    }, [navigate, location, redirectTo, checkInterval, customMessage]);

    // Show loading while checking
    if (isChecking && showLoadingWhileChecking) {
      return (
        <div style={{ padding: "20px", textAlign: "center" }}>
          <p>Checking authentication...</p>
          {import.meta.env.DEV && (
            <p style={{ fontSize: "12px", color: "#666", marginTop: "10px" }}>
              Dev mode: Attempting to restore password from localStorage...
              <br />
              Token refresh: Automatic via ApiClient interceptors (no workers)
            </p>
          )}
        </div>
      );
    }

    // If no password, don't render the component (redirect is happening)
    if (!hasPassword) {
      return (
        <div style={{ padding: "20px", textAlign: "center" }}>
          <p>{customMessage}</p>
        </div>
      );
    }

    // Render the protected component
    return <WrappedComponent {...props} />;
  };

  // Set display name for debugging
  PasswordProtectedComponent.displayName = `withPasswordProtection(${
    WrappedComponent.displayName || WrappedComponent.name
  })`;

  return PasswordProtectedComponent;
};

export default withPasswordProtection;

// Additional debug helper - updated without worker references
export const debugPasswordProtection = () => {
  const service = passwordStorageService;
  const info = service.getStorageInfo();

  console.log("Password Protection Debug Info:");
  console.log("Service initialized:", service.isInitialized);
  console.log("Storage mode:", info.mode);
  console.log("Is development:", info.isDevelopment);
  console.log("Has password in memory:", service.password !== null);
  console.log("Token refresh method: ApiClient interceptors (no workers)");
  console.log(
    "Storage type:",
    service.storage === localStorage
      ? "localStorage"
      : service.storage === sessionStorage
        ? "sessionStorage"
        : "unknown",
  );

  // Check localStorage directly
  const keys = Object.keys(localStorage).filter(
    (k) => k.includes("pwd") || k.includes("session"),
  );
  console.log("Password-related localStorage keys:", keys);

  return {
    serviceInfo: info,
    hasPasswordInMemory: service.password !== null,
    localStorageKeys: keys,
    refreshMethod: "api_interceptor",
    hasWorkers: false,
  };
};
