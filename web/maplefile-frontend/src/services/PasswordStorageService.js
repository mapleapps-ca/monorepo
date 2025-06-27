// File: src/services/PasswordStorageService.js
class PasswordStorageService {
  constructor() {
    this.password = null;
    this.sessionKey = null;
    this.timeout = null;
    this.inactivityTimeout = 30 * 60 * 1000; // 30 minutes
    this.isInitialized = false;

    this.setupActivityListeners();
    console.log("[PasswordStorageService] Service initialized");
  }

  // Store password temporarily
  setPassword(password) {
    this.password = password;
    this.storeEncryptedPassword();
    this.resetTimeout();
    console.log("[PasswordStorageService] Password stored temporarily");
  }

  // Get stored password
  getPassword() {
    if (this.password) {
      this.resetTimeout();
    }
    return this.password;
  }

  // Check if password is available
  hasPassword() {
    return this.password !== null;
  }

  // Clear password from memory and storage
  clearPassword() {
    this.password = null;
    this.sessionKey = null;
    if (typeof sessionStorage !== "undefined") {
      sessionStorage.removeItem("enc_pwd");
    }
    if (this.timeout) {
      clearTimeout(this.timeout);
    }
    console.log("[PasswordStorageService] Password cleared");
  }

  // Store encrypted password in sessionStorage for page refresh survival
  async storeEncryptedPassword() {
    if (!this.password || typeof sessionStorage === "undefined") return;

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

      sessionStorage.setItem(
        "enc_pwd",
        JSON.stringify({
          data: Array.from(new Uint8Array(encrypted)),
          iv: Array.from(iv),
          timestamp: Date.now(),
        }),
      );
    } catch (error) {
      console.warn(
        "[PasswordStorageService] Failed to store encrypted password:",
        error,
      );
    }
  }

  // Restore password from sessionStorage after page refresh
  async restorePassword() {
    if (typeof sessionStorage === "undefined") return false;

    const stored = sessionStorage.getItem("enc_pwd");
    if (!stored || !this.sessionKey) return false;

    try {
      const { data, iv, timestamp } = JSON.parse(stored);

      // Check if expired
      if (Date.now() - timestamp > this.inactivityTimeout) {
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
      this.clearPassword();
      return false;
    }
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

  // Initialize the service (called by ServiceContext)
  async initialize() {
    if (this.isInitialized) return;

    try {
      // Try to restore password from previous session
      await this.restorePassword();
      this.isInitialized = true;
      console.log("[PasswordStorageService] Service initialized successfully");
    } catch (error) {
      console.error("[PasswordStorageService] Initialization failed:", error);
      this.isInitialized = true; // Continue anyway
    }
  }
}

// Create singleton instance
const passwordStorageService = new PasswordStorageService();
export default passwordStorageService;

// ================================================================
// 2. In Collection Create.jsx - MODIFY YOUR EXISTING handleSubmit:

/*
const { passwordStorageService } = useServices(); // Add this line

const handleSubmit = async () => {
  // ... your existing validation

  // CHECK IF PASSWORD IS ALREADY AVAILABLE:
  const storedPassword = passwordStorageService.getPassword();
  if (storedPassword) {
    // Use stored password directly, skip prompt
    setPassword(storedPassword);
    await handlePasswordSubmit(storedPassword);
  } else {
    // Show password prompt as you currently do
    setShowPasswordPrompt(true);
  }
};

// MODIFY handlePasswordSubmit to accept password parameter:
const handlePasswordSubmit = async (passwordToUse = password) => {
  if (!passwordToUse) {
    setError("Password is required");
    return;
  }

  setLoading(true);
  setError("");

  try {
    // ... your existing creation logic using passwordToUse
  } catch (err) {
    // ... your existing error handling
  } finally {
    setLoading(false);
  }
};
*/

// ================================================================
// 3. In Collection List.jsx - MODIFY YOUR EXISTING loadCollections:

/*
const { passwordStorageService } = useServices(); // Add this line

const loadCollections = async (passwordParam = null, forceRefresh = false) => {
  try {
    setLoading(true);
    setError("");

    // CHECK FOR STORED PASSWORD IF NO PARAMETER PROVIDED:
    let passwordToUse = passwordParam;
    if (!passwordToUse) {
      passwordToUse = passwordStorageService.getPassword();
    }

    // ... rest of your existing loadCollections logic
    // Use passwordToUse instead of passwordParam

  } catch (err) {
    // ... your existing error handling
  } finally {
    setLoading(false);
  }
};

// MODIFY your useEffect to check for stored password:
useEffect(() => {
  if (authLoading) return;

  if (isAuthenticated) {
    const cachedData = loadCachedCollections();
    if (cachedData) {
      setCollections(cachedData.owned_collections || []);
      setSharedCollections(cachedData.shared_collections || []);
      setLoading(false);

      const needsDecryption = [...cachedData.owned_collections, ...cachedData.shared_collections]
        .some(c => c.decrypt_error);

      // CHECK FOR STORED PASSWORD BEFORE SHOWING PROMPT:
      if (needsDecryption && !passwordStorageService.hasPassword()) {
        setShowPasswordPrompt(true);
      }
    } else {
      // CHECK FOR STORED PASSWORD:
      if (passwordStorageService.hasPassword()) {
        // We have password, try to load directly
        loadCollections();
      } else if (localStorageService.hasUserEncryptedData()) {
        // Need password
        setShowPasswordPrompt(true);
        setLoading(false);
      } else {
        loadCollections();
      }
    }
  } else {
    navigate("/login");
  }
}, [isAuthenticated, authLoading, navigate]);
*/
