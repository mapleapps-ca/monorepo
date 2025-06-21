// Service Context with Dependency Injection for all authentication services
import React, { createContext, useContext, useEffect } from "react";
import AuthService from "../services/AuthService.js";
import MeService from "../services/MeService.js";
import TokenService from "../services/TokenService.js";
import CryptoService from "../services/CryptoService.js";
import LocalStorageService from "../services/LocalStorageService.js";
import ApiClient from "../services/ApiClient.js";
import WorkerManager from "../services/WorkerManager.js";

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

  // 1. Initialize CryptoService first (no dependencies)
  const cryptoService = CryptoService;

  // Initialize the cryptoService early
  useEffect(() => {
    cryptoService.init().catch(console.error); // Handle errors during initialization
  }, [cryptoService]);

  // 2. Initialize LocalStorageService (no dependencies)
  const localStorageService = LocalStorageService;

  // 3. Initialize TokenService (depends on LocalStorageService)
  const tokenService = new TokenService();

  // 4. Initialize ApiClient (depends on LocalStorageService)
  const apiClient = ApiClient;

  // 5. Initialize WorkerManager (depends on LocalStorageService)
  const workerManager = WorkerManager;

  // 6. Initialize AuthService (depends on CryptoService)
  const authService = new AuthService(cryptoService);

  // 7. Initialize MeService (depends on AuthService)
  const meService = new MeService(authService);

  // Create services object
  const services = {
    // Core services
    authService,
    meService,
    tokenService,
    cryptoService,

    // Storage and utility services
    localStorageService,
    apiClient,
    workerManager,
  };

  // Initialize services that need async initialization
  useEffect(() => {
    const initializeServices = async () => {
      try {
        console.log("[ServiceProvider] Initializing services...");

        // Initialize crypto service
        await cryptoService.init();
        console.log("[ServiceProvider] CryptoService initialized");

        // Initialize worker manager (this will also initialize the auth worker)
        try {
          await workerManager.initialize();
          console.log("[ServiceProvider] WorkerManager initialized");
        } catch (error) {
          console.warn(
            "[ServiceProvider] WorkerManager initialization failed, continuing without worker:",
            error,
          );
        }

        // Initialize auth service worker
        try {
          await authService.initializeWorker();
          console.log("[ServiceProvider] AuthService worker initialized");
        } catch (error) {
          console.warn(
            "[ServiceProvider] AuthService worker initialization failed:",
            error,
          );
        }

        console.log("[ServiceProvider] All services initialized successfully");
      } catch (error) {
        console.error(
          "[ServiceProvider] Service initialization failed:",
          error,
        );
      }
    };

    initializeServices();
  }, [cryptoService, authService, workerManager]);

  // Set up error handling for services
  useEffect(() => {
    const handleUnhandledRejection = (event) => {
      console.error(
        "[ServiceProvider] Unhandled promise rejection:",
        event.reason,
      );
      // You can add more sophisticated error handling here
    };

    const handleError = (event) => {
      console.error("[ServiceProvider] Unhandled error:", event.error);
      // You can add more sophisticated error handling here
    };

    window.addEventListener("unhandledrejection", handleUnhandledRejection);
    window.addEventListener("error", handleError);

    return () => {
      window.removeEventListener(
        "unhandledrejection",
        handleUnhandledRejection,
      );
      window.removeEventListener("error", handleError);
    };
  }, []);

  // Provide debug information in development
  useEffect(() => {
    if (import.meta.env.DEV) {
      console.log(
        "[ServiceProvider] Services available:",
        Object.keys(services),
      );
      console.log(
        "[ServiceProvider] CryptoService ready:",
        cryptoService.isInitialized,
      );
      console.log(
        "[ServiceProvider] AuthService authenticated:",
        authService.isAuthenticated(),
      );

      // Add services to window for debugging in development
      window.mapleAppsServices = services;
      console.log(
        "[ServiceProvider] Services added to window.mapleAppsServices for debugging",
      );
    }
  }, [services, cryptoService, authService]);

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
};
