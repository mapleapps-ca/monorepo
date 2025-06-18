// src/contexts/ServiceContext.jsx
import React, { createContext, useContext } from "react";
import AuthService from "../services/AuthService";

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
  const services = {
    authService: new AuthService(),
  };

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
};
