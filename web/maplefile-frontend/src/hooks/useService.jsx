// File: src/hooks/useService.jsx
import { useServices as useUnifiedServices } from "../services/Services";

/**
 * Legacy useServices hook for backward compatibility
 * This wraps the new unified service system
 */
export function useServices() {
  // Import all services from the new unified system
  const services = useUnifiedServices();

  // Return services with any legacy aliases needed
  return {
    ...services,
    // Add any backward compatibility aliases here
    authService: services.authManager, // Legacy alias
    meService: services.meManager, // Legacy alias
    tokenService: services.tokenManager, // Legacy alias
  };
}

// Default export for backward compatibility
export default useServices;
