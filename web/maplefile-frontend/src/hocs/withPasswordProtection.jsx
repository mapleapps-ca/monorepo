// File: src/hoc/withPasswordProtection.jsx
import React, { useEffect, useState } from "react";
import { useNavigate, useLocation } from "react-router";
import passwordStorageService from "../services/PasswordStorageService.js";

/**
 * HOC that protects components by checking if password is available
 * Redirects to login if no password is stored
 */
const withPasswordProtection = (WrappedComponent, options = {}) => {
  const {
    redirectTo = "/login",
    showLoadingWhileChecking = true,
    checkInterval = null, // Optional: check password periodically
    customMessage = "Password required. Redirecting to login...",
  } = options;

  const PasswordProtectedComponent = (props) => {
    const navigate = useNavigate();
    const location = useLocation();
    const [isChecking, setIsChecking] = useState(true);
    const [hasPassword, setHasPassword] = useState(false);

    const checkPassword = () => {
      const passwordAvailable = passwordStorageService.hasPassword();
      console.log(
        `[withPasswordProtection] Password check for ${location.pathname}:`,
        passwordAvailable,
      );

      setHasPassword(passwordAvailable);

      if (!passwordAvailable) {
        console.log(
          `[withPasswordProtection] No password found, redirecting to ${redirectTo}`,
        );
        // Store the attempted path for redirect after login
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

    useEffect(() => {
      const isPasswordAvailable = checkPassword();
      setIsChecking(false);

      // Set up periodic checking if specified
      let intervalId;
      if (checkInterval && isPasswordAvailable) {
        intervalId = setInterval(checkPassword, checkInterval);
      }

      return () => {
        if (intervalId) {
          clearInterval(intervalId);
        }
      };
    }, [navigate, location]);

    // Show loading while checking
    if (isChecking && showLoadingWhileChecking) {
      return (
        <div style={{ padding: "20px", textAlign: "center" }}>
          <p>Checking authentication...</p>
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
  PasswordProtectedComponent.displayName = `withPasswordProtection(${WrappedComponent.displayName || WrappedComponent.name})`;

  return PasswordProtectedComponent;
};

export default withPasswordProtection;

// ================================================================
// USAGE - Just wrap your component exports:

// In Collection/Create.jsx:
// import withPasswordProtection from '../../../hoc/withPasswordProtection.jsx';
// export default withPasswordProtection(CollectionCreate);

// In Collection/List.jsx:
// import withPasswordProtection from '../../../hoc/withPasswordProtection.jsx';
// export default withPasswordProtection(CollectionList);

// With custom options:
// export default withPasswordProtection(SomePage, {
//   checkInterval: 60000, // Check every minute for password expiration
//   customMessage: 'Session expired. Redirecting...'
// });
