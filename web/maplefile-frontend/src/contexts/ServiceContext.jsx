// File: monorepo/web/maplefile-frontend/src/contexts/ServiceContext.jsx
// ServiceContext - Updated to use both MeManager and SyncCollectionManager
import React, { createContext, useEffect } from "react";
import AuthManager from "../services/Manager/AuthManager.js";
import MeManager from "../services/Manager/MeManager.js";
import SyncCollectionManager from "../services/Manager/SyncCollectionManager.js";
import TokenService from "../services/TokenService.js";
import CryptoService from "../services/Crypto/CryptoService.js";
import LocalStorageService from "../services/Storage/LocalStorageService.js";
import ApiClient from "../services/API/ApiClient.js";
import WorkerManager from "../services/WorkerManager.js";
import PasswordStorageService from "../services/PasswordStorageService.js";
import SyncCollectionAPIService from "../services/API/SyncCollectionAPIService.js";
import SyncCollectionStorageService from "../services/Storage/SyncCollectionStorageService.js";

// Create a context for our services
export const ServiceContext = createContext();

// Create a provider component that will wrap our app
export function ServiceProvider({ children }) {
  // Initialize AuthManager (singleton)
  const authManager = new AuthManager();

  // Initialize MeManager with AuthManager dependency (replaces MeService)
  const meManager = new MeManager(authManager);

  // Initialize TokenService
  const tokenService = new TokenService();

  // Initialize SyncCollectionAPIService with AuthManager dependency
  const syncCollectionAPIService = new SyncCollectionAPIService(authManager);

  // Initialize SyncCollectionStorageService (no dependencies needed)
  const syncCollectionStorageService = new SyncCollectionStorageService();

  // Initialize SyncCollectionManager with AuthManager dependency
  const syncCollectionManager = new SyncCollectionManager(authManager);

  // Create services object with all services
  const services = {
    // Core services (singletons)
    authManager: authManager, // New: AuthManager for orchestration
    authService: authManager, // Backward compatibility alias
    cryptoService: CryptoService,
    passwordStorageService: PasswordStorageService,
    localStorageService: LocalStorageService,
    apiClient: ApiClient,
    workerManager: WorkerManager, // Simplified version without web workers

    // Services that need dependency injection
    meManager, // New: MeManager (replaces meService)
    meService: meManager, // Backward compatibility alias
    tokenService,
    syncCollectionAPIService,
    syncCollectionStorageService,
    syncCollectionManager, // New: SyncCollectionManager (replaces syncCollectionService)
    syncCollectionService: syncCollectionManager, // Backward compatibility alias
  };

  // Initialize services that need async initialization
  useEffect(() => {
    const initializeServices = async () => {
      try {
        console.log(
          "[ServiceProvider] Initializing services with AuthManager...",
        );

        // Initialize crypto service
        await CryptoService.initialize();
        console.log("[ServiceProvider] CryptoService initialized");

        // Initialize password storage service
        await PasswordStorageService.initialize();
        console.log("[ServiceProvider] PasswordStorageService initialized");

        // Initialize simplified worker manager (no actual workers)
        try {
          await WorkerManager.initialize();
          console.log(
            "[ServiceProvider] WorkerManager initialized (simplified)",
          );
        } catch (error) {
          console.warn(
            "[ServiceProvider] WorkerManager initialization failed, continuing:",
            error,
          );
        }

        // Initialize auth manager (simplified)
        try {
          await authManager.initializeWorker();
          console.log("[ServiceProvider] AuthManager initialized (no workers)");
        } catch (error) {
          console.warn(
            "[ServiceProvider] AuthManager initialization failed:",
            error,
          );
        }

        // Initialize me manager
        try {
          await meManager.initialize();
          console.log("[ServiceProvider] MeManager initialized");
        } catch (error) {
          console.warn(
            "[ServiceProvider] MeManager initialization failed:",
            error,
          );
        }

        // Initialize sync collection manager
        try {
          await syncCollectionManager.initialize();
          console.log("[ServiceProvider] SyncCollectionManager initialized");
        } catch (error) {
          console.warn(
            "[ServiceProvider] SyncCollectionManager initialization failed:",
            error,
          );
        }

        console.log(
          "[ServiceProvider] All services initialized successfully with AuthManager",
        );
      } catch (error) {
        console.error(
          "[ServiceProvider] Service initialization failed:",
          error,
        );
      }
    };

    initializeServices();
  }, [authManager, meManager, syncCollectionManager]);

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
        "[ServiceProvider] AuthManager authenticated:",
        authManager.isAuthenticated(),
      );
      console.log(
        "[ServiceProvider] Architecture: Manager/API/Storage pattern with AuthManager orchestrator",
      );

      // Add services to window for debugging in development
      window.mapleAppsServices = services;
      console.log(
        "[ServiceProvider] Services added to window.mapleAppsServices for debugging",
      );
    }
  }, [services, authManager]);

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
}
