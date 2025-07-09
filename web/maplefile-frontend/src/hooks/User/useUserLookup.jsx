// File: monorepo/web/maplefile-frontend/src/hooks/User/useUserLookup.jsx
// Simplified hook for user public key lookup

import { useState } from "react";
import { useUsers } from "../useService.jsx";

/**
 * Simplified hook for user public key lookup
 * @returns {Object} User lookup API
 */
const useUserLookup = () => {
  const { userLookupManager } = useUsers();

  // Simple state management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  // Lookup user with basic error handling
  const lookupUser = async (email) => {
    if (!userLookupManager) {
      throw new Error("User lookup manager not available");
    }

    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      console.log("[useUserLookup] Looking up user:", email);

      const result = await userLookupManager.lookupUser(email);

      setSuccess(`User lookup successful!`);
      console.log("[useUserLookup] User lookup successful:", result);
      return result;
    } catch (err) {
      console.error("[useUserLookup] User lookup failed:", err);
      setError(err.message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  // Check if user exists
  const userExists = async (email) => {
    if (!userLookupManager) {
      throw new Error("User lookup manager not available");
    }

    setIsLoading(true);
    setError(null);

    try {
      const exists = await userLookupManager.userExists(email);
      console.log("[useUserLookup] User existence check:", exists);
      return exists;
    } catch (err) {
      console.error("[useUserLookup] User existence check failed:", err);
      setError(err.message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  // Get user's public key for encryption
  const getUserPublicKey = async (email) => {
    if (!userLookupManager) {
      throw new Error("User lookup manager not available");
    }

    setIsLoading(true);
    setError(null);

    try {
      const result = await userLookupManager.getUserPublicKey(email);
      setSuccess(`Public key retrieved for ${result.name}!`);
      console.log("[useUserLookup] User public key retrieved:", result);
      return result;
    } catch (err) {
      console.error("[useUserLookup] Failed to get user public key:", err);
      setError(err.message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  // Clear messages
  const clearMessages = () => {
    setError(null);
    setSuccess(null);
  };

  // Validate email format
  const validateEmail = (email) => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  // Sanitize email
  const sanitizeEmail = (email) => {
    return email.toLowerCase().trim();
  };

  return {
    // State
    isLoading,
    error,
    success,

    // Core operations
    lookupUser,
    userExists,
    getUserPublicKey,

    // Utilities
    clearMessages,
    validateEmail,
    sanitizeEmail,

    // Status
    isAvailable: !!userLookupManager,
  };
};

export default useUserLookup;
