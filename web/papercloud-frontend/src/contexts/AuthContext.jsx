// web/prototyping/papercloud-frontend/src/contexts/AuthContext.jsx
import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
} from "react";
import { useNavigate } from "react-router";
import tokenManager from "../services/TokenManager";
import { initSodium, cryptoUtils } from "../utils/crypto"; // Import initSodium and cryptoUtils

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [userEmail, setUserEmail] = useState(null);
  const [masterKey, setMasterKey] = useState(null);
  const [privateKey, setPrivateKey] = useState(null);
  const [publicKey, setPublicKey] = useState(null);
  const [salt, setSalt] = useState(null);
  const [sodium, setSodiumInstance] = useState(null); // Renamed to avoid conflict with imported _sodium
  const [authError, setAuthError] = useState(null); // For displaying errors from this context

  const navigate = useNavigate();

  useEffect(() => {
    const initializeApp = async () => {
      setIsLoading(true);
      setAuthError(null);
      try {
        const S = await initSodium(); // Initialize sodium through cryptoUtils
        setSodiumInstance(S); // Store the fully initialized instance
        console.log("Sodium instance set in AuthContext state");

        await tokenManager.initialize();
        const loggedIn = tokenManager.isLoggedIn();
        setIsAuthenticated(loggedIn);
        if (loggedIn) {
          setUserEmail(localStorage.getItem("userEmail"));
          // Attempt to load E2EE keys from localStorage
          console.log("AuthContext: User is logged in, attempting to load E2EE keys from localStorage.");
          const masterKeyB64 = localStorage.getItem("masterKey_e2ee");
          const privateKeyB64 = localStorage.getItem("privateKey_e2ee");
          const publicKeyB64 = localStorage.getItem("publicKey_e2ee");
          const saltB64 = localStorage.getItem("salt_e2ee");

          if (masterKeyB64 && privateKeyB64 && publicKeyB64 && saltB64) {
            try {
              // Ensure sodium is ready before using cryptoUtils for decoding
              await initSodium(); 
              
              const loadedMasterKey = await cryptoUtils.fromBase64(masterKeyB64);
              const loadedPrivateKey = await cryptoUtils.fromBase64(privateKeyB64);
              const loadedPublicKey = await cryptoUtils.fromBase64(publicKeyB64);
              const loadedSalt = await cryptoUtils.fromBase64(saltB64);

              setMasterKey(loadedMasterKey);
              setPrivateKey(loadedPrivateKey);
              setPublicKey(loadedPublicKey);
              setSalt(loadedSalt);
              console.log("AuthContext: E2EE Keys successfully loaded from localStorage and set in state.");
            } catch (keyLoadError) {
              console.error("AuthContext: Failed to decode/set E2EE keys from localStorage. Clearing them.", keyLoadError);
              localStorage.removeItem("masterKey_e2ee");
              localStorage.removeItem("privateKey_e2ee");
              localStorage.removeItem("publicKey_e2ee");
              localStorage.removeItem("salt_e2ee");
              // Optionally, trigger logout or set specific error state if keys are corrupt/unusable
            }
          } else {
            console.log("AuthContext: Some or all E2EE Keys not found in localStorage.");
          }
        }
      } catch (error) {
        console.error("AuthContext: Error during app initialization:", error);
        setAuthError("Initialization failed. Please refresh.");
      } finally {
        setIsLoading(false);
      }
    };

    initializeApp();

    const unsubscribe = tokenManager.addListener(
      (event, success, listenerError) => {
        if (event === "refresh") {
          if (!success) {
            setIsAuthenticated(false);
            clearE2EEKeys();
            navigate("/login", {
              replace: true,
              state: {
                from: window.location.pathname,
                error: `Your session has expired or refresh failed: ${listenerError?.message || "Unknown reason"}. Please log in again.`,
              },
            });
          } else {
            setIsAuthenticated(true);
          }
        }
      },
    );

    return () => {
      unsubscribe();
      tokenManager.cleanup();
    };
  }, [navigate]); // Removed sodium from deps as it's initialized within

  const clearE2EEKeys = () => {
    setMasterKey(null);
    setPrivateKey(null);
    setPublicKey(null);
    setSalt(null);
    // Clear from localStorage
    localStorage.removeItem("masterKey_e2ee");
    localStorage.removeItem("privateKey_e2ee");
    localStorage.removeItem("publicKey_e2ee");
    localStorage.removeItem("salt_e2ee");
    console.log("AuthContext: E2EE Keys cleared from memory and localStorage.");
  };

  const logout = useCallback(() => {
    tokenManager.clearTokens();
    localStorage.removeItem("userEmail");
    setIsAuthenticated(false);
    setUserEmail(null);
    clearE2EEKeys();
    setAuthError(null); // Clear errors on logout
    navigate("/login");
  }, [navigate]);

  const login = useCallback(
    async ( // Made async to use await for cryptoUtils
      accessToken,
      accessTokenExpiry,
      refreshToken,
      refreshTokenExpiry,
      decryptedMasterKey,
      decryptedPrivateKey,
      userPublicKeyBytes, // Expecting Uint8Array
      userSaltBytes, // Expecting Uint8Array
      emailForLogin,
    ) => {
      try {
        tokenManager.updateTokens(
          accessToken,
          accessTokenExpiry,
          refreshToken,
          refreshTokenExpiry,
        );
        setIsAuthenticated(true);
        setUserEmail(emailForLogin);
        localStorage.setItem("userEmail", emailForLogin);

        setMasterKey(decryptedMasterKey);
        setPrivateKey(decryptedPrivateKey);
        setPublicKey(userPublicKeyBytes);
        setSalt(userSaltBytes);

        // Persist E2EE keys to localStorage
        // Ensure cryptoUtils are available (sodium should be initialized by this point via initSodium in useEffect or CompleteLogin)
        if (decryptedMasterKey) localStorage.setItem("masterKey_e2ee", await cryptoUtils.toBase64(decryptedMasterKey));
        if (decryptedPrivateKey) localStorage.setItem("privateKey_e2ee", await cryptoUtils.toBase64(decryptedPrivateKey));
        if (userPublicKeyBytes) localStorage.setItem("publicKey_e2ee", await cryptoUtils.toBase64(userPublicKeyBytes));
        if (userSaltBytes) localStorage.setItem("salt_e2ee", await cryptoUtils.toBase64(userSaltBytes));
        
        console.log(
          "AuthContext: E2EE Keys and user email stored in memory and localStorage.",
        );
        setAuthError(null); // Clear any previous auth errors
      } catch (error) {
        console.error("Error in AuthContext login function:", error);
        // clearE2EEKeys (called by logout) will handle removing from localStorage
        logout(); // Clear all state on login processing error
        setAuthError(`Login processing failed: ${error.message}`);
      }
    },
    [logout], // Added logout as a dependency because it's called
  );

  const getAccessToken = useCallback(() => {
    return tokenManager.getAccessToken();
  }, []);

  const value = {
    isAuthenticated,
    isLoading,
    userEmail,
    login,
    logout,
    getAccessToken,
    masterKey,
    privateKey,
    publicKey,
    salt,
    sodium, // Provide the initialized sodium instance
    authError, // Provide error state
  };

  // Display loading or error state from AuthContext itself if critical
  if (isLoading) return <div>Loading application authentication...</div>;
  if (authError && !isAuthenticated)
    return (
      <div>
        Critical Error: {authError}{" "}
        <button onClick={() => window.location.reload()}>Refresh</button>
      </div>
    );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === null) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
