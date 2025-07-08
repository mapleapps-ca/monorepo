// File: monorepo/web/maplefile-frontend/src/hooks/useService.jsx
// Updated to include DeleteFileManager and RecoveryManager
import { useContext } from "react";
import { ServiceContext } from "../contexts/ServiceContext.jsx";

// Hook to access all services
export const useServices = () => {
  const services = useContext(ServiceContext);
  if (!services) {
    throw new Error("useServices must be used within a ServiceProvider");
  }
  return services;
};

// Hook to access authentication services
export const useAuth = () => {
  const services = useServices();
  return {
    authManager: services.authManager,
    authService: services.authService, // Backward compatibility
  };
};

// Hook to access recovery services
export const useRecovery = () => {
  const services = useServices();
  return {
    recoveryManager: services.recoveryManager,
  };
};

// Hook to access crypto services
export const useCrypto = () => {
  const services = useServices();
  return {
    cryptoService: services.cryptoService,
    CollectionCryptoService: services.CollectionCryptoService,
  };
};

// Hook to access storage services
export const useStorage = () => {
  const services = useServices();
  return {
    localStorageService: services.localStorageService,
    passwordStorageService: services.passwordStorageService,
  };
};

// Hook to access API client
export const useApiClient = () => {
  const services = useServices();
  return services.apiClient;
};

// Hook to access collection services
export const useCollections = () => {
  const services = useServices();
  return {
    syncCollectionAPIService: services.syncCollectionAPIService,
    syncCollectionStorageService: services.syncCollectionStorageService,
    syncCollectionManager: services.syncCollectionManager,
    createCollectionManager: services.createCollectionManager,
    getCollectionManager: services.getCollectionManager,
    updateCollectionManager: services.updateCollectionManager,
    deleteCollectionManager: services.deleteCollectionManager,
    listCollectionManager: services.listCollectionManager,
  };
};

// Hook to access file services
export const useFiles = () => {
  const services = useServices();
  return {
    syncFileAPIService: services.syncFileAPIService,
    syncFileStorageService: services.syncFileStorageService,
    syncFileManager: services.syncFileManager,
    createFileManager: services.createFileManager,
    getFileManager: services.getFileManager,
    downloadFileManager: services.downloadFileManager,
    deleteFileManager: services.deleteFileManager, // NEW
  };
};

// Hook to access user/profile services
export const useProfile = () => {
  const services = useServices();
  return {
    meManager: services.meManager,
    meService: services.meService, // Backward compatibility
  };
};

// Hook to access token services
export const useTokens = () => {
  const services = useServices();
  return {
    tokenManager: services.tokenManager,
    tokenService: services.tokenService, // Backward compatibility
  };
};

// Hook specifically for download file operations
export const useFileDownloads = () => {
  const services = useServices();
  return {
    downloadFileManager: services.downloadFileManager,
    getFileManager: services.getFileManager, // Often needed for file metadata
    getCollectionManager: services.getCollectionManager, // For collection keys
  };
};

// Hook specifically for delete file operations (NEW)
export const useFileDeletions = () => {
  const services = useServices();
  return {
    deleteFileManager: services.deleteFileManager,
    getFileManager: services.getFileManager, // Often needed for file metadata
    getCollectionManager: services.getCollectionManager, // For collection keys
    listCollectionManager: services.listCollectionManager, // For collection listing
  };
};

// Hook for comprehensive file operations (NEW)
export const useFileOperations = () => {
  const services = useServices();
  return {
    createFileManager: services.createFileManager,
    getFileManager: services.getFileManager,
    downloadFileManager: services.downloadFileManager,
    deleteFileManager: services.deleteFileManager,
    syncFileManager: services.syncFileManager,
    // Collection managers often needed for file operations
    getCollectionManager: services.getCollectionManager,
    listCollectionManager: services.listCollectionManager,
  };
};

// Default export for backward compatibility
export default useServices;
