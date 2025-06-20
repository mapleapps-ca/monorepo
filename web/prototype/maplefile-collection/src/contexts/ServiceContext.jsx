import React, { createContext, useContext, useEffect } from "react";
import ApiService from "../services/ApiService";
import CollectionService from "../services/CollectionService";
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
  // Initialize all services following dependency injection principles
  const cryptoService = new CryptoService();
  const apiService = new ApiService();
  const collectionService = new CollectionService(apiService);

  // Initialize the cryptoService early
  useEffect(() => {
    cryptoService.init().catch(console.error);
  }, [cryptoService]);

  const services = {
    apiService,
    collectionService,
    cryptoService,
  };

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
};
