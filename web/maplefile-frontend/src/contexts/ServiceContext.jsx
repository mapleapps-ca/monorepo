// src/contexts/ServiceContext.js
import React, { createContext, useContext } from "react";
import AuthService from "../services/AuthService";
import MeService from "../services/MeService";

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
  // Initialize all services here
  // Note: MeService depends on AuthService, so we pass it as a dependency
  const authService = new AuthService();
  const meService = new MeService(authService);

  const services = {
    authService,
    meService,
  };

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
};
