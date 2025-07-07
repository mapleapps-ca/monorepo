// File: monorepo/web/maplefile-frontend/src/hooks/useService.jsx
// Custom hook to access services from ServiceContext
import { useContext } from "react";
import { ServiceContext } from "../contexts/ServiceContext.jsx";

/**
 * Hook to access all services from ServiceContext
 * @returns {Object} All available services
 */
export const useServices = () => {
  const context = useContext(ServiceContext);

  if (!context) {
    throw new Error("useServices must be used within a ServiceProvider");
  }

  return context;
};

/**
 * Hook to access a specific service from ServiceContext
 * @param {string} serviceName - Name of the service to retrieve
 * @returns {Object} The requested service
 */
export const useService = (serviceName) => {
  const services = useServices();

  if (!services[serviceName]) {
    throw new Error(`Service "${serviceName}" not found in ServiceContext`);
  }

  return services[serviceName];
};

/**
 * Hook to access authentication services
 * @returns {Object} Authentication related services
 */
export const useAuth = () => {
  const services = useServices();

  return {
    authManager: services.authManager,
    authService: services.authService, // Backward compatibility
    tokenManager: services.tokenManager,
    tokenService: services.tokenService, // Backward compatibility
    meManager: services.meManager,
    meService: services.meService, // Backward compatibility
  };
};

/**
 * Hook to access collection services
 * @returns {Object} Collection related services
 */
export const useCollections = () => {
  const services = useServices();

  return {
    syncCollectionManager: services.syncCollectionManager,
    createCollectionManager: services.createCollectionManager,
    syncCollectionAPIService: services.syncCollectionAPIService,
    syncCollectionStorageService: services.syncCollectionStorageService,
  };
};

/**
 * Hook to access file services
 * @returns {Object} File related services
 */
export const useFiles = () => {
  const services = useServices();

  return {
    syncFileManager: services.syncFileManager,
    syncFileAPIService: services.syncFileAPIService,
    syncFileStorageService: services.syncFileStorageService,
  };
};

/**
 * Hook to access crypto services
 * @returns {Object} Crypto related services
 */
export const useCrypto = () => {
  const services = useServices();

  return {
    cryptoService: services.cryptoService,
    passwordStorageService: services.passwordStorageService,
  };
};

/**
 * Hook to access storage services
 * @returns {Object} Storage related services
 */
export const useStorage = () => {
  const services = useServices();

  return {
    localStorageService: services.localStorageService,
    syncCollectionStorageService: services.syncCollectionStorageService,
    syncFileStorageService: services.syncFileStorageService,
  };
};

/**
 * Hook to access API services
 * @returns {Object} API related services
 */
export const useAPI = () => {
  const services = useServices();

  return {
    apiClient: services.apiClient,
    syncCollectionAPIService: services.syncCollectionAPIService,
    syncFileAPIService: services.syncFileAPIService,
  };
};

// Default export for backward compatibility
export default useServices;
