// File: src/hooks/useService.jsx
import { useServices } from "../services/Services";

/**
 * Legacy useService hook for backward compatibility
 * This wraps the new unified service system
 */
export function useService() {
  return useServices();
}

// Default export for backward compatibility
export default useService;
