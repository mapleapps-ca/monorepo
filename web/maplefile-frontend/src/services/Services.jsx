// File: src/services/Services.jsx
// Complete single-file service boundary interface - TOTAL REPLACEMENT

import React, { createContext, useContext, useEffect } from "react";

// ========================================
// SERVICE CLASS IMPORTS
// ========================================

// Core Services
import AuthManager from "./Manager/AuthManager.js";
import MeManager from "./Manager/MeManager.js";
import TokenManager from "./Manager/TokenManager.js";
import RecoveryManager from "./Manager/RecoveryManager.js";

// Crypto Services
import CryptoService from "./Crypto/CryptoService.js";
import CollectionCryptoService from "./Crypto/CollectionCryptoService.js";

// Storage Services
import LocalStorageService from "./Storage/LocalStorageService.js";
import passwordStorageService from "./PasswordStorageService.js";

// API Services
import ApiClient, { setApiClientAuthManager } from "./API/ApiClient.js";

// Collection Services
import SyncCollectionAPIService from "./API/SyncCollectionAPIService.js";
import SyncCollectionStorageService from "./Storage/SyncCollectionStorageService.js";
import SyncCollectionManager from "./Manager/SyncCollectionManager.js";
import CreateCollectionManager from "./Manager/Collection/CreateCollectionManager.js";
import GetCollectionManager from "./Manager/Collection/GetCollectionManager.js";
import UpdateCollectionManager from "./Manager/Collection/UpdateCollectionManager.js";
import DeleteCollectionManager from "./Manager/Collection/DeleteCollectionManager.js";
import ListCollectionManager from "./Manager/Collection/ListCollectionManager.js";
import ShareCollectionManager from "./Manager/Collection/ShareCollectionManager.js";

// File Services
import SyncFileAPIService from "./API/SyncFileAPIService.js";
import SyncFileStorageService from "./Storage/SyncFileStorageService.js";
import SyncFileManager from "./Manager/SyncFileManager.js";
import CreateFileManager from "./Manager/File/CreateFileManager.js";
import GetFileManager from "./Manager/File/GetFileManager.js";
import DownloadFileManager from "./Manager/File/DownloadFileManager.js";
import DeleteFileManager from "./Manager/File/DeleteFileManager.js";

// User Services
import UserLookupManager from "./Manager/User/UserLookupManager.js";

// ========================================
// SERVICE CREATION & DEPENDENCY INJECTION
// ========================================

function createServices() {
  console.log(
    "[Services] 🚀 Creating service instances with dependency injection...",
  );

  // ========================================
  // 1. CORE SERVICES (No Dependencies)
  // ========================================

  const authManager = new AuthManager();
  console.log("[Services] ✓ AuthManager created");

  // ========================================
  // 2. AUTH-DEPENDENT SERVICES
  // ========================================

  const meManager = new MeManager(authManager);
  const tokenManager = new TokenManager(authManager);
  const recoveryManager = new RecoveryManager();
  console.log("[Services] ✓ Auth-dependent managers created");

  // ========================================
  // 3. COLLECTION SERVICES
  // ========================================

  const syncCollectionAPIService = new SyncCollectionAPIService(authManager);
  const syncCollectionStorageService = new SyncCollectionStorageService();
  const syncCollectionManager = new SyncCollectionManager(authManager);
  const createCollectionManager = new CreateCollectionManager(authManager);
  const getCollectionManager = new GetCollectionManager(authManager);
  const updateCollectionManager = new UpdateCollectionManager(authManager);
  const deleteCollectionManager = new DeleteCollectionManager(authManager);
  const listCollectionManager = new ListCollectionManager(authManager);
  const shareCollectionManager = new ShareCollectionManager(authManager);
  console.log("[Services] ✓ Collection services created");

  // ========================================
  // 4. FILE SERVICES (Complex Dependencies)
  // ========================================

  const syncFileAPIService = new SyncFileAPIService(authManager);
  const syncFileStorageService = new SyncFileStorageService();
  const syncFileManager = new SyncFileManager(authManager);
  const createFileManager = new CreateFileManager(authManager);

  // File services with collection dependencies
  const getFileManager = new GetFileManager(
    authManager,
    getCollectionManager,
    listCollectionManager,
  );

  const downloadFileManager = new DownloadFileManager(
    authManager,
    getFileManager,
    getCollectionManager,
  );

  const deleteFileManager = new DeleteFileManager(
    authManager,
    getCollectionManager,
    listCollectionManager,
  );
  console.log("[Services] ✓ File services created with dependencies");

  // ========================================
  // 5. USER SERVICES
  // ========================================

  const userLookupManager = new UserLookupManager();
  console.log("[Services] ✓ User services created");

  // ========================================
  // 6. API CLIENT SETUP
  // ========================================

  setApiClientAuthManager(authManager);
  console.log("[Services] ✓ ApiClient configured with AuthManager");

  // ========================================
  // 7. SERVICE REGISTRY
  // ========================================

  const services = {
    // Core services (singletons)
    authManager,
    authService: authManager, // Backward compatibility alias
    cryptoService: CryptoService,
    CollectionCryptoService,
    passwordStorageService,
    localStorageService: LocalStorageService,
    apiClient: ApiClient,

    // Auth & User Management
    meManager,
    meService: meManager, // Backward compatibility alias
    tokenManager,
    tokenService: tokenManager, // Backward compatibility alias
    recoveryManager,

    // Collection Services
    syncCollectionAPIService,
    syncCollectionStorageService,
    syncCollectionManager,
    createCollectionManager,
    getCollectionManager,
    updateCollectionManager,
    deleteCollectionManager,
    listCollectionManager,
    shareCollectionManager,

    // File Services
    createFileManager,
    getFileManager,
    downloadFileManager,
    deleteFileManager,
    syncFileAPIService,
    syncFileStorageService,
    syncFileManager,

    // User Services
    userLookupManager,
  };

  console.log(
    "[Services] ✅ Service registry created with",
    Object.keys(services).length,
    "services",
  );

  return services;
}

// ========================================
// REACT CONTEXT & PROVIDER
// ========================================

const ServiceContext = createContext();

export function ServiceProvider({ children }) {
  const services = createServices();

  // ========================================
  // SERVICE INITIALIZATION
  // ========================================

  useEffect(() => {
    const initializeServices = async () => {
      try {
        console.log("[Services] 🚀 Starting service initialization...");

        // ========================================
        // 1. CORE CRYPTO & STORAGE SERVICES
        // ========================================

        await CryptoService.initialize();
        console.log("[Services] ✓ CryptoService initialized");

        await CollectionCryptoService.initialize();
        console.log("[Services] ✓ CollectionCryptoService initialized");

        try {
          await passwordStorageService.initialize();
          console.log("[Services] ✓ PasswordStorageService initialized");
        } catch (error) {
          console.warn(
            "[Services] ⚠️ PasswordStorageService initialization failed:",
            error,
          );
        }

        // ========================================
        // 2. AUTH MANAGER
        // ========================================

        try {
          await services.authManager.initializeWorker();
          console.log("[Services] ✓ AuthManager initialized");
        } catch (error) {
          console.warn(
            "[Services] ⚠️ AuthManager initialization failed:",
            error,
          );
        }

        // ========================================
        // 3. MANAGER INITIALIZATION
        // ========================================

        const managersToInitialize = [
          { manager: services.meManager, name: "MeManager" },
          { manager: services.tokenManager, name: "TokenManager" },
          { manager: services.recoveryManager, name: "RecoveryManager" },
          {
            manager: services.syncCollectionManager,
            name: "SyncCollectionManager",
          },
          {
            manager: services.createCollectionManager,
            name: "CreateCollectionManager",
          },
          {
            manager: services.getCollectionManager,
            name: "GetCollectionManager",
          },
          {
            manager: services.updateCollectionManager,
            name: "UpdateCollectionManager",
          },
          {
            manager: services.deleteCollectionManager,
            name: "DeleteCollectionManager",
          },
          {
            manager: services.listCollectionManager,
            name: "ListCollectionManager",
          },
          {
            manager: services.shareCollectionManager,
            name: "ShareCollectionManager",
          },
          { manager: services.createFileManager, name: "CreateFileManager" },
          { manager: services.getFileManager, name: "GetFileManager" },
          {
            manager: services.downloadFileManager,
            name: "DownloadFileManager",
          },
          { manager: services.deleteFileManager, name: "DeleteFileManager" },
          { manager: services.syncFileManager, name: "SyncFileManager" },
          { manager: services.userLookupManager, name: "UserLookupManager" },
        ];

        for (const { manager, name } of managersToInitialize) {
          try {
            await manager.initialize();
            console.log(`[Services] ✓ ${name} initialized`);
          } catch (error) {
            console.warn(`[Services] ⚠️ ${name} initialization failed:`, error);
          }
        }

        console.log("[Services] 🎉 All services initialized successfully!");
      } catch (error) {
        console.error(
          "[Services] ❌ Critical service initialization failure:",
          error,
        );
      }
    };

    initializeServices();
  }, [services]);

  // ========================================
  // ERROR HANDLING
  // ========================================

  useEffect(() => {
    const handleUnhandledRejection = (event) => {
      console.error("[Services] 🚨 Unhandled promise rejection:", event.reason);
    };

    const handleError = (event) => {
      console.error("[Services] 🚨 Unhandled error:", event.error);
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

  // ========================================
  // DEVELOPMENT DEBUG INFO
  // ========================================

  useEffect(() => {
    if (import.meta.env.DEV) {
      console.log("[Services] 🔧 Development mode - adding debug info");
      console.log("[Services] Available services:", Object.keys(services));
      console.log(
        "[Services] CryptoService ready:",
        CryptoService.isInitialized,
      );
      console.log(
        "[Services] CollectionCryptoService ready:",
        CollectionCryptoService.isReady(),
      );
      console.log(
        "[Services] PasswordStorageService ready:",
        passwordStorageService.isInitialized,
      );
      console.log(
        "[Services] AuthManager authenticated:",
        services.authManager.isAuthenticated(),
      );
      console.log(
        "[Services] UserLookupManager ready:",
        services.userLookupManager ? "✅" : "❌",
      );
      console.log(
        "[Services] 🏗️ Architecture: Single-file service boundary with full DI",
      );

      // Add services to window for debugging
      window.mapleAppsServices = services;
      console.log(
        "[Services] 🪟 Services available at window.mapleAppsServices for debugging",
      );
    }
  }, [services]);

  return (
    <ServiceContext.Provider value={services}>
      {children}
    </ServiceContext.Provider>
  );
}

// ========================================
// SERVICE HOOKS - BOUNDARY INTERFACE
// ========================================

/**
 * Main service hook - returns ALL services
 * This is your primary interface to the service layer
 */
export function useServices() {
  const context = useContext(ServiceContext);
  if (!context) {
    throw new Error("useServices must be used within a ServiceProvider");
  }
  return context;
}

/**
 * Authentication & User Management Services
 */
export function useAuth() {
  const {
    authManager,
    authService,
    meManager,
    meService,
    tokenManager,
    tokenService,
    recoveryManager,
  } = useServices();
  return {
    authManager,
    authService, // alias
    meManager,
    meService, // alias
    tokenManager,
    tokenService, // alias
    recoveryManager,
  };
}

/**
 * Collection Management Services
 */
export function useCollections() {
  const {
    syncCollectionAPIService,
    syncCollectionStorageService,
    syncCollectionManager,
    createCollectionManager,
    getCollectionManager,
    updateCollectionManager,
    deleteCollectionManager,
    listCollectionManager,
    shareCollectionManager,
  } = useServices();

  return {
    syncCollectionAPIService,
    syncCollectionStorageService,
    syncCollectionManager,
    createCollectionManager,
    getCollectionManager,
    updateCollectionManager,
    deleteCollectionManager,
    listCollectionManager,
    shareCollectionManager,
  };
}

/**
 * File Management Services
 * 🔧 FIXED: Now includes auth and collection dependencies that file managers need
 */
export function useFiles() {
  const {
    // File services
    syncFileAPIService,
    syncFileStorageService,
    syncFileManager,
    createFileManager,
    getFileManager,
    downloadFileManager,
    deleteFileManager,

    // 🔧 ADDED: Dependencies that file managers need
    authManager,
    authService, // alias for authManager
    getCollectionManager,
    listCollectionManager,
  } = useServices();

  return {
    // File services
    syncFileAPIService,
    syncFileStorageService,
    syncFileManager,
    createFileManager,
    getFileManager,
    downloadFileManager,
    deleteFileManager,

    // 🔧 ADDED: Dependencies that file managers need
    authService: authManager, // Use authManager as authService
    authManager, // Also provide direct access
    getCollectionManager,
    listCollectionManager,
  };
}

/**
 * Cryptography Services
 */
export function useCrypto() {
  const { cryptoService, CollectionCryptoService } = useServices();
  return {
    cryptoService,
    CollectionCryptoService,
  };
}

/**
 * Storage Services
 */
export function useStorage() {
  const { localStorageService, passwordStorageService } = useServices();
  return {
    localStorageService,
    passwordStorageService,
  };
}

/**
 * API Services
 */
export function useApi() {
  const { apiClient } = useServices();
  return {
    apiClient,
  };
}

/**
 * User Services
 */
export function useUsers() {
  const { userLookupManager } = useServices();
  return {
    userLookupManager,
  };
}
