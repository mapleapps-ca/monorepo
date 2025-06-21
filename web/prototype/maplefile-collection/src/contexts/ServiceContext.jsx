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
  console.log("🔧 ServiceProvider initializing...");
  console.log("🔧 Environment variables:");
  console.log("  - DEV mode:", import.meta.env.DEV);
  console.log("  - VITE_API_BASE_URL:", import.meta.env.VITE_API_BASE_URL);
  console.log("  - All env vars:", import.meta.env);

  // Initialize all services following dependency injection principles
  const cryptoService = new CryptoService();

  // Initialize ApiService with NO parameters to use default logic
  console.log("🔧 Initializing ApiService with default constructor...");
  const apiService = new ApiService(); // No parameters!

  console.log("🔧 Initializing CollectionService...");
  const collectionService = new CollectionService(apiService);

  // Initialize services and run health checks
  useEffect(() => {
    const initializeServices = async () => {
      try {
        // Initialize crypto service
        await cryptoService.init();
        console.log("✅ CryptoService initialized successfully");

        // Run API health check in development
        if (import.meta.env.DEV) {
          console.log("🔍 Running API health check...");
          const health = await apiService.healthCheck();
          console.log("🔍 API Health Check result:", health);

          if (!health.ok) {
            console.warn(
              "⚠️ API health check failed. Make sure your backend server is running on http://localhost:8000",
            );
            console.warn("Expected backend endpoints:");
            console.warn(
              "- GET http://localhost:8000/maplefile/api/v1/collections",
            );
            console.warn(
              "- Ensure CORS is properly configured on your backend",
            );
          } else {
            console.log("✅ API connection successful");
          }
        }
      } catch (error) {
        console.error("❌ Service initialization error:", error);
      }
    };

    initializeServices();
  }, [cryptoService, apiService]);

  const services = {
    apiService,
    collectionService,
    cryptoService,
  };

  console.log("🔧 ServiceProvider services created:", services);

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
};
