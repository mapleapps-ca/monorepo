// File: src/services/passwordStorage.js

class PasswordStorageService {
  constructor(options = {}) {
    this.password = null;
    this.sessionKey = null;
    this.timeout = null;
    this.inactivityTimeout = options.inactivityTimeout || 30 * 60 * 1000; // 30 minutes
    this.setupActivityListeners();
  }

  // Store password temporarily
  setPassword(password) {
    this.password = password;
    this.storeEncryptedPassword();
    this.resetTimeout();
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
      console.warn("Failed to store encrypted password:", error);
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
}

// Create singleton instance
const passwordService = new PasswordStorageService();

export default passwordService;
