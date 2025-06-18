// web/maplefile-frontend/src/contexts/ServiceContext.jsx
// src/contexts/ServiceContext.js
import React, { createContext, useContext, useEffect } from "react"; // Added useEffect
import AuthService from "../services/AuthService";
import MeService from "../services/MeService";
import TokenService from "../services/TokenService";
import CryptoService from "../services/CryptoService";

// Create a context for our services
const ServiceContext = createContext();

// Create a custom hook to use our services
export const useServices = () => {
  const context = useContext(ServiceContext);
  if (!context) {
    throw new Error("useServices must be used within a ServiceProvider");
  }
  return context;
};

// Create a provider component that will wrap our app
export const ServiceProvider = ({ children }) => {
  // Initialize all services here following dependency injection principles
  const cryptoService = new CryptoService();

  // Initialize the cryptoService early
  useEffect(() => {
    cryptoService.init().catch(console.error); // Handle errors during initialization
  }, []);

  const authService = new AuthService(cryptoService);
  const meService = new MeService(authService);
  const tokenService = new TokenService();

  const services = {
    authService,
    meService,
    tokenService,
    cryptoService,
  };

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
};
