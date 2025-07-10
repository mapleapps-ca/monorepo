// File: monorepo/web/maplefile-frontend/src/hocs/withPasswordProtection.jsx
// Enhanced HOC with password expiry event handling - UPDATED for unified services
import React, { useEffect, useState } from "react";
import { useNavigate, useLocation } from "react-router";
import passwordStorageService from "../services/PasswordStorageService.js";

/**
 * HOC that protects components by checking if password is available
 * Enhanced with password expiry event handling - Works with unified services architecture
 */
const withPasswordProtection = (WrappedComponent, options = {}) => {
  const {
    redirectTo = "/login",
    showLoadingWhileChecking = true,
    checkInterval = null,
    customMessage = "Password required. Redirecting to login...",
    onPasswordExpired = null, // Callback when password expires
  } = options;

  const PasswordProtectedComponent = (props) => {
    const navigate = useNavigate();
    const location = useLocation();
    const [isChecking, setIsChecking] = useState(true);
    const [hasPassword, setHasPassword] = useState(false);
    const [expiryMessage, setExpiryMessage] = useState("");

    useEffect(() => {
      let mounted = true;
      let intervalId = null;

      // ENHANCED: Listen for password expiry events
      const handlePasswordExpired = (event) => {
        console.log(
          "[withPasswordProtection] Password expired event received:",
          event.detail,
        );

        if (!mounted) return;

        const reason = event.detail?.reason || "unknown";
        let message = "Session expired - please log in again";

        switch (reason) {
          case "inactivity_timeout":
            message = "Session expired due to inactivity - please log in again";
            break;
          case "manual_clear":
            message = "Session cleared - please log in again";
            break;
          default:
            message = "Session expired - please log in again";
        }

        setExpiryMessage(message);
        setHasPassword(false);

        // Call custom callback if provided
        if (onPasswordExpired) {
          try {
            onPasswordExpired(reason, message);
          } catch (error) {
            console.error(
              "[withPasswordProtection] Error in onPasswordExpired callback:",
              error,
            );
          }
        }

        // Redirect to login after a brief delay to show the message
        setTimeout(() => {
          if (mounted) {
            console.log(
              `[withPasswordProtection] Redirecting due to password expiry: ${message}`,
            );
            navigate(redirectTo, {
              state: {
                from: location,
                message: message,
                reason: reason,
              },
              replace: true,
            });
          }
        }, 2000); // 2 second delay to show the message
      };

      // Add password expiry event listener
      window.addEventListener("passwordExpired", handlePasswordExpired);

      const checkPasswordAsync = async () => {
        console.log(
          `[withPasswordProtection] Starting password check for ${location.pathname}`,
        );

        // UPDATED: Check if unified services are available
        let authManager = null;
        if (window.mapleAppsServices?.authManager) {
          authManager = window.mapleAppsServices.authManager;
          console.log(
            "[withPasswordProtection] Found authManager from unified services",
          );
        }

        // Check authentication first if authManager is available
        if (authManager && typeof authManager.isAuthenticated === "function") {
          try {
            const isAuthenticated = authManager.isAuthenticated();
            console.log(
              "[withPasswordProtection] Authentication status:",
              isAuthenticated,
            );

            if (!isAuthenticated) {
              console.log(
                "[withPasswordProtection] User not authenticated, redirecting to login",
              );
              if (mounted) {
                navigate(redirectTo, {
                  state: {
                    from: location,
                    message: "Please log in to access this page",
                  },
                  replace: true,
                });
              }
              return false;
            }
          } catch (error) {
            console.warn(
              "[withPasswordProtection] Error checking authentication:",
              error,
            );
          }
        }

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

        if (!passwordAvailable && !expiryMessage) {
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
            if (!stillHasPassword && mounted && !expiryMessage) {
              console.log(
                "[withPasswordProtection] Password lost during periodic check, redirecting...",
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
        window.removeEventListener("passwordExpired", handlePasswordExpired);
        if (intervalId) {
          clearInterval(intervalId);
        }
      };
    }, [
      navigate,
      location,
      redirectTo,
      checkInterval,
      customMessage,
      onPasswordExpired,
      expiryMessage,
    ]);

    // Show expiry message if password expired
    if (expiryMessage) {
      return (
        <div
          style={{
            padding: "20px",
            textAlign: "center",
            backgroundColor: "#fff3cd",
            border: "1px solid #ffeaa7",
            borderRadius: "4px",
            margin: "20px",
          }}
        >
          <h3 style={{ color: "#856404", margin: "0 0 10px 0" }}>
            ⚠️ Session Expired
          </h3>
          <p style={{ color: "#856404", margin: "0 0 15px 0" }}>
            {expiryMessage}
          </p>
          <p style={{ fontSize: "14px", color: "#666" }}>
            Redirecting to login...
          </p>
        </div>
      );
    }

    // Show loading while checking
    if (isChecking && showLoadingWhileChecking) {
      return (
        <div style={{ padding: "20px", textAlign: "center" }}>
          <p>Checking authentication...</p>
          {import.meta.env.DEV && (
            <div style={{ fontSize: "12px", color: "#666", marginTop: "10px" }}>
              <p>
                Dev mode: Attempting to restore password from localStorage...
              </p>
              <p>
                Unified Services: Using window.mapleAppsServices for authManager
                access
              </p>
              <p>
                Token refresh: Automatic via ApiClient interceptors (no workers)
              </p>
              <p>
                Password timeout: Extended on user activity and API calls
                (logging reduced)
              </p>
            </div>
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

// ENHANCED: Additional debug helper with password service status
export const debugPasswordProtection = () => {
  const service = passwordStorageService;
  const info = service.getStorageInfo();

  console.log("Enhanced Password Protection Debug Info:");
  console.log("Service initialized:", service.isInitialized);
  console.log("Storage mode:", info.mode);
  console.log("Is development:", info.isDevelopment);
  console.log("Has password in memory:", service.password !== null);
  console.log("Unified Services available:", !!window.mapleAppsServices);
  console.log(
    "AuthManager available:",
    !!window.mapleAppsServices?.authManager,
  );
  console.log("Token refresh method: ApiClient interceptors (no workers)");
  console.log("Password timeout minutes:", info.timeoutMinutes);
  console.log("Activity detected:", info.activityDetected);
  console.log("Logging: Reduced verbosity for activity tracking");
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
    timeoutMinutes: info.timeoutMinutes,
    activityDetected: info.activityDetected,
    loggingMode: "reduced_verbosity",
    unifiedServicesAvailable: !!window.mapleAppsServices,
    authManagerAvailable: !!window.mapleAppsServices?.authManager,
  };
};

// ENHANCED: Utility to manually extend password timeout
export const extendPasswordTimeout = () => {
  if (passwordStorageService.hasPassword()) {
    passwordStorageService.resetTimeout();
    console.log("[withPasswordProtection] Password timeout manually extended");
    return true;
  } else {
    console.warn(
      "[withPasswordProtection] No password available to extend timeout",
    );
    return false;
  }
};

// ENHANCED: Utility to get password service status
export const getPasswordServiceStatus = () => {
  return {
    hasPassword: passwordStorageService.hasPassword(),
    isInitialized: passwordStorageService.isInitialized,
    storageInfo: passwordStorageService.getStorageInfo(),
    detailedStatus: passwordStorageService.getDetailedStatus
      ? passwordStorageService.getDetailedStatus()
      : null,
    unifiedServicesIntegration: {
      available: !!window.mapleAppsServices,
      authManager: !!window.mapleAppsServices?.authManager,
      passwordStorageService:
        !!window.mapleAppsServices?.passwordStorageService,
    },
  };
};
