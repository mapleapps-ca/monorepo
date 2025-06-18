// src/contexts/ServiceContext.js
import React, { createContext, useContext } from "react";
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
  // CryptoService has no dependencies
  const cryptoService = new CryptoService();

  // AuthService depends on CryptoService
  const authService = new AuthService(cryptoService);

  // MeService depends on AuthService
  const meService = new MeService(authService);

  // TokenService has no dependencies
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
