import { createContext, useContext, useState, useEffect } from "react";
import { useService } from "../hooks/useService.js";
import { TYPES } from "../di/types.js";

// Create the authentication context
const AuthContext = createContext(null);

// Provider component that gives auth functionality to your whole app
export const AuthProvider = ({ children }) => {
  // Get the auth service using our custom hook
  const authService = useService(TYPES.AuthService);

  // React state to track authentication status
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  // Check authentication status when the component loads
  useEffect(() => {
    setIsAuthenticated(authService.isUserAuthenticated());
    setUser(authService.getCurrentUser());
  }, [authService]);

  // Function to request an OTT
  const requestOTT = async (email) => {
    setIsLoading(true);
    try {
      const result = await authService.requestOTT(email);
      setIsLoading(false);
      return result;
    } catch (error) {
      setIsLoading(false);
      throw error;
    }
  };

  // Function to verify an OTT
  const verifyOTT = async (email, token) => {
    setIsLoading(true);
    try {
      const result = await authService.verifyOTT(email, token);
      setIsLoading(false);
      return result;
    } catch (error) {
      setIsLoading(false);
      throw error;
    }
  };

  // Function to complete login
  const completeLogin = async (sessionToken) => {
    setIsLoading(true);
    try {
      const result = await authService.completeLogin(sessionToken);
      if (result.success) {
        setIsAuthenticated(true);
        setUser(result.user);
      }
      setIsLoading(false);
      return result;
    } catch (error) {
      setIsLoading(false);
      throw error;
    }
  };

  // Function to log out
  const logout = () => {
    authService.logout();
    setIsAuthenticated(false);
    setUser(null);
  };

  // Provide all auth functions to child components
  const value = {
    isAuthenticated,
    user,
    isLoading,
    requestOTT,
    verifyOTT,
    completeLogin,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

// Hook for components to use auth functionality
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};
