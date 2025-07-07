// File: monorepo/web/maplefile-frontend/src/contexts/ServiceContext.jsx
// ServiceContext - Updated to include CollectionCryptoService
import React, { createContext, useEffect } from "react";
import AuthManager from "../services/Manager/AuthManager.js";
import MeManager from "../services/Manager/MeManager.js";
import TokenManager from "../services/Manager/TokenManager.js";
import CryptoService from "../services/Crypto/CryptoService.js";
import CollectionCryptoService from "../services/Crypto/CollectionCryptoService.js";
import LocalStorageService from "../services/Storage/LocalStorageService.js";
import ApiClient, {
  setApiClientAuthManager,
} from "../services/API/ApiClient.js";
import PasswordStorageService from "../services/PasswordStorageService.js";
import SyncCollectionAPIService from "../services/API/SyncCollectionAPIService.js";
import SyncCollectionStorageService from "../services/Storage/SyncCollectionStorageService.js";
import SyncCollectionManager from "../services/Manager/SyncCollectionManager.js";

import SyncFileAPIService from "../services/API/SyncFileAPIService.js";
import SyncFileStorageService from "../services/Storage/SyncFileStorageService.js";
import SyncFileManager from "../services/Manager/SyncFileManager.js";

import CreateCollectionManager from "../services/Manager/Collection/CreateCollectionManager.js";

// Create a context for our services
export const ServiceContext = createContext();

// Create a provider component that will wrap our app
export function ServiceProvider({ children }) {
  // Initialize AuthManager (singleton)
  const authManager = new AuthManager();

  // Initialize MeManager with AuthManager dependency (replaces MeService)
  const meManager = new MeManager(authManager);

  // Initialize TokenManager with AuthManager dependency (replaces TokenService)
  const tokenManager = new TokenManager(authManager);

  // Initialize SyncCollectionAPIService with AuthManager dependency
  const syncCollectionAPIService = new SyncCollectionAPIService(authManager);

  // Initialize SyncCollectionStorageService (no dependencies needed)
  const syncCollectionStorageService = new SyncCollectionStorageService();

  // Initialize SyncCollectionManager with AuthManager dependency
  const syncCollectionManager = new SyncCollectionManager(authManager);

  // Initialize CreateCollectionManager with AuthManager dependency
  const createCollectionManager = new CreateCollectionManager(authManager);

  // Initialize SyncFileAPIService with AuthManager dependency
  const syncFileAPIService = new SyncFileAPIService(authManager);

  // Initialize SyncFileStorageService (no dependencies needed)
  const syncFileStorageService = new SyncFileStorageService();

  // Initialize SyncFileManager with AuthManager dependency
  const syncFileManager = new SyncFileManager(authManager);

  // Set AuthManager on ApiClient for event notifications
  setApiClientAuthManager(authManager);

  // Create services object with all services
  const services = {
    // Core services (singletons)
    authManager: authManager, // New: AuthManager for orchestration
    authService: authManager, // Backward compatibility alias
    cryptoService: CryptoService,
    CollectionCryptoService: CollectionCryptoService, // New: Collection-specific crypto operations
    passwordStorageService: PasswordStorageService,
    localStorageService: LocalStorageService,
    apiClient: ApiClient,

    // Services that need dependency injection
    meManager, // New: MeManager (replaces meService)
    meService: meManager, // Backward compatibility alias
    tokenManager, // New: TokenManager (replaces tokenService)
    tokenService: tokenManager, // Backward compatibility alias

    // Collection services
    syncCollectionAPIService,
    syncCollectionStorageService,
    syncCollectionManager, // New: SyncCollectionManager (replaces syncCollectionService)
    createCollectionManager, // New: CreateCollectionManager for collection creation

    // File services
    syncFileAPIService,
    syncFileStorageService,
    syncFileManager, // New: SyncFileManager (replaces syncFileService)
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

        // Initialize collection crypto service
        await CollectionCryptoService.initialize();
        console.log("[ServiceProvider] CollectionCryptoService initialized");

        // Initialize password storage service
        await PasswordStorageService.initialize();
        console.log("[ServiceProvider] PasswordStorageService initialized");

        // Initialize auth manager
        try {
          await authManager.initializeWorker();
          console.log("[ServiceProvider] AuthManager initialized");
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

        // Initialize token manager
        try {
          await tokenManager.initialize();
          console.log("[ServiceProvider] TokenManager initialized");
        } catch (error) {
          console.warn(
            "[ServiceProvider] TokenManager initialization failed:",
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

        // Initialize create collection manager
        try {
          await createCollectionManager.initialize();
          console.log("[ServiceProvider] CreateCollectionManager initialized");
        } catch (error) {
          console.warn(
            "[ServiceProvider] CreateCollectionManager initialization failed:",
            error,
          );
        }

        // Initialize sync file manager
        try {
          await syncFileManager.initialize();
          console.log("[ServiceProvider] SyncFileManager initialized");
        } catch (error) {
          console.warn(
            "[ServiceProvider] SyncFileManager initialization failed:",
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
  }, [
    authManager,
    meManager,
    tokenManager,
    syncCollectionManager,
    createCollectionManager,
    syncFileManager,
  ]);

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
        "[ServiceProvider] CollectionCryptoService ready:",
        CollectionCryptoService.isReady(),
      );
      console.log(
        "[ServiceProvider] AuthManager authenticated:",
        authManager.isAuthenticated(),
      );
      console.log(
        "[ServiceProvider] Architecture: Manager/API/Storage/Crypto pattern with AuthManager orchestrator",
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
