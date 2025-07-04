// Updated ServiceContext.jsx with CollectionCryptoService
import React, { createContext, useEffect } from "react";
import AuthService from "../services/AuthService.js";
import MeService from "../services/MeService.js";
import TokenService from "../services/TokenService.js";
import CryptoService from "../services/CryptoService.js";
import LocalStorageService from "../services/LocalStorageService.js";
import ApiClient from "../services/ApiClient.js";
import WorkerManager from "../services/WorkerManager.js";
import PasswordStorageService from "../services/PasswordStorageService.js";

// Create a context for our services
export const ServiceContext = createContext();

// Create a provider component that will wrap our app
export function ServiceProvider({ children }) {
  // All services are singletons that are already instantiated
  // We just need to import them and provide them via context

  // Initialize MeService with AuthService dependency
  const meService = new MeService(AuthService);

  // Initialize TokenService
  const tokenService = new TokenService();

  // Create services object with all singleton services
  const services = {
    // Core services (singletons)
    authService: AuthService,
    cryptoService: CryptoService,
    passwordStorageService: PasswordStorageService,
    localStorageService: LocalStorageService,
    apiClient: ApiClient,
    workerManager: WorkerManager,

    // Services that need dependency injection
    meService,
    tokenService,
  };

  // Initialize services that need async initialization
  useEffect(() => {
    const initializeServices = async () => {
      try {
        console.log("[ServiceProvider] Initializing services...");

        // Initialize crypto service
        await CryptoService.initialize();
        console.log("[ServiceProvider] CryptoService initialized");

        // Initialize password storage service
        await PasswordStorageService.initialize();
        console.log("[ServiceProvider] PasswordStorageService initialized");

        // Initialize worker manager (this will also initialize the auth worker)
        try {
          await WorkerManager.initialize();
          console.log("[ServiceProvider] WorkerManager initialized");
        } catch (error) {
          console.warn(
            "[ServiceProvider] WorkerManager initialization failed, continuing without worker:",
            error,
          );
        }

        // Initialize auth service worker
        try {
          await AuthService.initializeWorker();
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
  }, []);

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
        CryptoService.isInitialized,
      );
      console.log(
        "[ServiceProvider] AuthService authenticated:",
        AuthService.isAuthenticated(),
      );

      // Add services to window for debugging in development
      window.mapleAppsServices = services;
      console.log(
        "[ServiceProvider] Services added to window.mapleAppsServices for debugging",
      );
    }
  }, [services]);

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
}
