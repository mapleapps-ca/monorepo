// File: monorepo/web/maplefile-frontend/src/hooks/useService.jsx
// Updated to include ListCollectionManager
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
    listCollectionManager: services.listCollectionManager, // NEW
  };
};

// Hook to access file services
export const useFiles = () => {
  const services = useServices();
  return {
    syncFileAPIService: services.syncFileAPIService,
    syncFileStorageService: services.syncFileStorageService,
    syncFileManager: services.syncFileManager,
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

// Default export for backward compatibility
export default useServices;
