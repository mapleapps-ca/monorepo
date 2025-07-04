// File: monorepo/web/maplefile-frontend/src/services/PasswordStorageService.js
// Enhanced with development mode localStorage support

class PasswordStorageService {
  constructor() {
    this.password = null;
    this.sessionKey = null;
    this.timeout = null;
    this.inactivityTimeout = 30 * 60 * 1000; // 30 minutes
    this.isInitialized = false;

    // Determine storage mode from environment
    this.storageMode =
      import.meta.env.VITE_PASSWORD_STORAGE_MODE || "sessionStorage";
    this.isDevelopment =
      import.meta.env.VITE_DEV_MODE === "true" || import.meta.env.DEV;

    // Additional safety check - never allow localStorage if not explicitly in dev mode
    if (
      this.storageMode === "localStorage" &&
      !import.meta.env.DEV &&
      import.meta.env.PROD
    ) {
      console.error(
        "[PasswordStorageService] Blocking localStorage in production build!",
      );
      this.storageMode = "sessionStorage";
    }

    // Security warning for development mode
    if (this.storageMode === "localStorage" && this.isDevelopment) {
      console.warn(
        "%c⚠️ WARNING: Using localStorage for password storage! %c\n" +
          "This is INSECURE and should ONLY be used in development.\n" +
          "Never use this in production!",
        "background: #ff0000; color: #ffffff; font-weight: bold; font-size: 16px; padding: 4px;",
        "color: #ff0000; font-weight: bold;",
      );
    } else if (this.storageMode === "localStorage" && !this.isDevelopment) {
      // Force sessionStorage in production even if misconfigured
      console.error(
        "[PasswordStorageService] localStorage requested in production - forcing sessionStorage for security!",
      );
      this.storageMode = "sessionStorage";
    }

    this.storage = this.getStorage();
    this.STORAGE_KEY = "mapleapps_dev_pwd";
    this.STORAGE_METADATA_KEY = "mapleapps_dev_pwd_meta";

    this.setupActivityListeners();
    console.log(
      `[PasswordStorageService] Initialized with ${this.storageMode} mode`,
    );
  }

  // Get the appropriate storage based on mode
  getStorage() {
    if (typeof window === "undefined") return null;

    if (this.storageMode === "localStorage" && this.isDevelopment) {
      return window.localStorage;
    }
    return window.sessionStorage;
  }

  // Store password temporarily
  setPassword(password) {
    this.password = password;
    this.storeEncryptedPassword();
    this.resetTimeout();
    console.log(
      `[PasswordStorageService] Password stored in ${this.storageMode}`,
    );
  }

  // Get stored password
  getPassword() {
    // First check memory
    if (this.password) {
      this.resetTimeout();
      return this.password;
    }

    // Try to restore from storage (useful for dev mode)
    if (this.isDevelopment && this.storageMode === "localStorage") {
      this.restorePasswordFromStorage();
      return this.password;
    }

    return null;
  }

  // Check if password is available
  hasPassword() {
    // Check memory first
    if (this.password !== null) return true;

    // In dev mode with localStorage, try to restore
    if (this.isDevelopment && this.storageMode === "localStorage") {
      this.restorePasswordFromStorage();
      return this.password !== null;
    }

    return false;
  }

  // Clear password from memory and storage
  clearPassword() {
    this.password = null;
    this.sessionKey = null;

    if (this.storage) {
      this.storage.removeItem(this.STORAGE_KEY);
      this.storage.removeItem(this.STORAGE_METADATA_KEY);
    }

    if (this.timeout) {
      clearTimeout(this.timeout);
    }

    console.log("[PasswordStorageService] Password cleared");
  }

  // Store encrypted password in storage
  async storeEncryptedPassword() {
    if (!this.password || !this.storage) return;

    try {
      // Generate session key for encryption
      this.sessionKey = window.crypto.getRandomValues(new Uint8Array(32));

      // Encrypt password
      const iv = window.crypto.getRandomValues(new Uint8Array(12));
      const key = await window.crypto.subtle.importKey(
        "raw",
        this.sessionKey,
        { name: "AES-GCM" },
        false,
        ["encrypt"],
      );

      const encrypted = await window.crypto.subtle.encrypt(
        { name: "AES-GCM", iv },
        key,
        new TextEncoder().encode(this.password),
      );

      // Store encrypted data
      const encryptedData = {
        data: Array.from(new Uint8Array(encrypted)),
        iv: Array.from(iv),
        timestamp: Date.now(),
      };

      this.storage.setItem(this.STORAGE_KEY, JSON.stringify(encryptedData));

      // Store metadata separately for easier checking
      this.storage.setItem(
        this.STORAGE_METADATA_KEY,
        JSON.stringify({
          timestamp: Date.now(),
          mode: this.storageMode,
          isDev: this.isDevelopment,
        }),
      );

      // Store session key in memory only (or in storage for dev mode)
      if (this.isDevelopment && this.storageMode === "localStorage") {
        // In dev mode, we need to persist the key too (INSECURE!)
        this.storage.setItem(
          "mapleapps_dev_session_key",
          JSON.stringify(Array.from(this.sessionKey)),
        );
      }
    } catch (error) {
      console.warn(
        "[PasswordStorageService] Failed to store encrypted password:",
        error,
      );
    }
  }

  // Restore password from storage (enhanced for dev mode)
  async restorePasswordFromStorage() {
    if (!this.storage) return false;

    const stored = this.storage.getItem(this.STORAGE_KEY);
    const metadata = this.storage.getItem(this.STORAGE_METADATA_KEY);

    if (!stored) return false;

    try {
      const { data, iv, timestamp } = JSON.parse(stored);

      // Check if expired
      if (Date.now() - timestamp > this.inactivityTimeout) {
        this.clearPassword();
        return false;
      }

      // Get session key
      if (!this.sessionKey) {
        if (this.isDevelopment && this.storageMode === "localStorage") {
          // In dev mode, restore session key from storage
          const storedKey = this.storage.getItem("mapleapps_dev_session_key");
          if (storedKey) {
            this.sessionKey = new Uint8Array(JSON.parse(storedKey));
          }
        }
      }

      if (!this.sessionKey) {
        console.warn(
          "[PasswordStorageService] No session key available for decryption",
        );
        this.clearPassword();
        return false;
      }

      // Decrypt password
      const key = await window.crypto.subtle.importKey(
        "raw",
        this.sessionKey,
        { name: "AES-GCM" },
        false,
        ["decrypt"],
      );

      const decrypted = await window.crypto.subtle.decrypt(
        { name: "AES-GCM", iv: new Uint8Array(iv) },
        key,
        new Uint8Array(data),
      );

      this.password = new TextDecoder().decode(decrypted);
      this.resetTimeout();

      console.log("[PasswordStorageService] Password restored from storage");
      return true;
    } catch (error) {
      console.warn(
        "[PasswordStorageService] Failed to restore password:",
        error,
      );
      this.clearPassword();
      return false;
    }
  }

  // Original restorePassword method for backward compatibility
  async restorePassword() {
    return this.restorePasswordFromStorage();
  }

  // Reset inactivity timeout
  resetTimeout() {
    if (this.timeout) {
      clearTimeout(this.timeout);
    }

    this.timeout = setTimeout(() => {
      console.log(
        "[PasswordStorageService] Password expired due to inactivity",
      );
      this.clearPassword();
    }, this.inactivityTimeout);
  }

  // Setup activity listeners to reset timeout
  setupActivityListeners() {
    if (typeof document === "undefined") return;

    const events = ["click", "keydown", "mousemove", "scroll"];
    events.forEach((event) => {
      document.addEventListener(
        event,
        () => {
          if (this.hasPassword()) {
            this.resetTimeout();
          }
        },
        { passive: true },
      );
    });
  }

  // Initialize the service
  async initialize() {
    if (this.isInitialized) return;

    try {
      // Try to restore password from storage
      if (this.isDevelopment && this.storageMode === "localStorage") {
        console.log(
          "[PasswordStorageService] Attempting to restore password from localStorage...",
        );
        const restored = await this.restorePasswordFromStorage();
        if (restored) {
          console.log(
            "[PasswordStorageService] Password successfully restored from localStorage",
          );
        } else {
          console.log(
            "[PasswordStorageService] No valid password found in localStorage",
          );
        }
      }

      this.isInitialized = true;
      console.log("[PasswordStorageService] Service initialized successfully");
    } catch (error) {
      console.error("[PasswordStorageService] Initialization failed:", error);
      this.isInitialized = true; // Continue anyway
    }
  }

  // Get storage mode info (useful for debugging)
  getStorageInfo() {
    return {
      mode: this.storageMode,
      isDevelopment: this.isDevelopment,
      hasPassword: this.hasPassword(),
      isInitialized: this.isInitialized,
      storageAvailable: !!this.storage,
    };
  }

  // Set custom inactivity timeout
  setInactivityTimeout(minutes) {
    this.inactivityTimeout = minutes * 60 * 1000;
    this.resetTimeout();
    console.log(
      `[PasswordStorageService] Inactivity timeout set to ${minutes} minutes`,
    );
  }

  // Force clear all stored data (useful for debugging)
  forceClearAll() {
    this.clearPassword();

    if (this.storage) {
      // Clear dev session key if exists
      this.storage.removeItem("mapleapps_dev_session_key");
    }

    console.log("[PasswordStorageService] All data force cleared");
  }
}

// Create singleton instance
const passwordStorageService = new PasswordStorageService();

// In development, add to window for debugging
if (import.meta.env.DEV) {
  window.__passwordService = passwordStorageService;
  console.log(
    "[PasswordStorageService] Debug access available at window.__passwordService",
  );
}

export default passwordStorageService;
