// File: src/hooks/useAuth.js
// UPDATED to work with unified services architecture via window.mapleAppsServices
/**
 * Legacy useAuth hook for backward compatibility
 * Uses direct access to unified services via window.mapleAppsServices
 */
function useAuth() {
  // Access services from the unified service system via window
  const authManager = window.mapleAppsServices?.authManager;
  const meManager = window.mapleAppsServices?.meManager;

  // Create user object from authManager
  const user = {
    email: authManager?.getCurrentUserEmail?.() || null,
    isAuthenticated: authManager?.isAuthenticated?.() || false,
  };

  // Create logout function
  const logout = () => {
    if (authManager?.logout) {
      authManager.logout();
    }
    // Also clear any session storage
    sessionStorage.clear();
    localStorage.removeItem("mapleapps_access_token");
    localStorage.removeItem("mapleapps_refresh_token");
    localStorage.removeItem("mapleapps_user_email");
  };

  return {
    user,
    logout,
    authManager,
    authService: authManager, // alias for backward compatibility
    meManager,
    meService: meManager, // alias for backward compatibility
    isAuthenticated: user.isAuthenticated,
  };
}

export default useAuth;
