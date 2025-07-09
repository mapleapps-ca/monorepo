// File: monorepo/web/maplefile-frontend/src/hooks/useService.jsx
// Updated to include ShareCollectionManager (fixed naming conflict)
import { useContext } from "react";
import { ServiceContext } from "../contexts/ServiceContext.jsx";

// Main hook to get all services
export const useService = () => {
  const context = useContext(ServiceContext);
  if (!context) {
    throw new Error("useService must be used within a ServiceProvider");
  }
  return context;
};

// Hook for authentication services (renamed to avoid conflict with existing useAuth)
export const useAuthServices = () => {
  const services = useService();
  return {
    authManager: services.authManager,
    authService: services.authService, // Backward compatibility
    meManager: services.meManager,
    meService: services.meService, // Backward compatibility
    tokenManager: services.tokenManager,
    tokenService: services.tokenService, // Backward compatibility
    recoveryManager: services.recoveryManager,
  };
};

// Alternative export name for backward compatibility
export const useServices = useAuthServices;

// Hook for crypto services
export const useCrypto = () => {
  const services = useService();
  return {
    cryptoService: services.cryptoService,
    CollectionCryptoService: services.CollectionCryptoService,
  };
};

// Hook for storage services
export const useStorage = () => {
  const services = useService();
  return {
    localStorageService: services.localStorageService,
    passwordStorageService: services.passwordStorageService,
  };
};

// Hook for API services
export const useApi = () => {
  const services = useService();
  return {
    apiClient: services.apiClient,
  };
};

// Hook for collection services
export const useCollections = () => {
  const services = useService();
  return {
    // Sync services
    syncCollectionAPIService: services.syncCollectionAPIService,
    syncCollectionStorageService: services.syncCollectionStorageService,
    syncCollectionManager: services.syncCollectionManager,

    // CRUD services
    createCollectionManager: services.createCollectionManager,
    getCollectionManager: services.getCollectionManager,
    updateCollectionManager: services.updateCollectionManager,
    deleteCollectionManager: services.deleteCollectionManager,
    listCollectionManager: services.listCollectionManager,
    shareCollectionManager: services.shareCollectionManager,
  };
};

// Hook for file services
export const useFiles = () => {
  const services = useService();
  return {
    // Sync services
    syncFileAPIService: services.syncFileAPIService,
    syncFileStorageService: services.syncFileStorageService,
    syncFileManager: services.syncFileManager,

    // CRUD services
    createFileManager: services.createFileManager,
    getFileManager: services.getFileManager,
    downloadFileManager: services.downloadFileManager,
    deleteFileManager: services.deleteFileManager,
  };
};

// Default export
export default useService;
