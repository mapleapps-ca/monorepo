// src/services/LoginService.js

import axios from "axios";

/**
 * LoginService handles the three-step login process with end-to-end encryption
 *
 * This service demonstrates several important concepts:
 * 1. Multi-step authentication flows
 * 2. Client-side cryptographic operations for login
 * 3. Secure token handling and storage
 * 4. Integration with background token refresh systems
 *
 * The three-step process ensures that:
 * - Passwords never leave the user's device
 * - Email ownership is verified before password checking
 * - Authentication challenges prove knowledge of both email and password
 * - Tokens are securely encrypted before transmission
 */
export class LoginService {
  constructor(logger, cryptoService) {
    this.logger = logger;
    this.cryptoService = cryptoService;

    // Build API URL from environment variables
    this.apiBaseUrl = `${import.meta.env.VITE_API_PROTOCOL}://${import.meta.env.VITE_API_DOMAIN}`;

    // Store intermediate login state between steps
    this.loginState = {
      email: null,
      challengeId: null,
      userKeys: null,
      step: 0,
    };

    this.logger.log(
      `LoginService: Initialized with API base URL: ${this.apiBaseUrl}`,
    );
  }

  /**
   * Step 1: Request One-Time Token
   * This initiates the login process by sending a verification code to the user's email
   *
   * @param {string} email - The user's email address
   * @returns {Promise<Object>} Result with success status and message
   */
  async requestOTT(email) {
    this.logger.log(`LoginService: Requesting OTT for email: ${email}`);

    try {
      // Clear any previous login state when starting fresh
      this.clearLoginState();

      // Validate email format before making API call
      if (!this.isValidEmail(email)) {
        throw new Error("Please enter a valid email address");
      }

      // Make API call to request the one-time token
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/request-ott`,
        {
          email: email.toLowerCase().trim(),
        },
        {
          headers: { "Content-Type": "application/json" },
          timeout: 10000, // 10 second timeout
        },
      );

      if (response.status === 200) {
        // Store email for subsequent steps
        this.loginState.email = email.toLowerCase().trim();
        this.loginState.step = 1;

        this.logger.log("LoginService: OTT requested successfully");
        return {
          success: true,
          message:
            response.data.message || "Verification code sent to your email",
          email: this.loginState.email,
        };
      } else {
        throw new Error(`Unexpected response status: ${response.status}`);
      }
    } catch (error) {
      return this.handleLoginError("requestOTT", error);
    }
  }

  /**
   * Step 2: Verify One-Time Token and Prepare Cryptographic Challenge
   * This step verifies email ownership and retrieves encrypted user data
   *
   * @param {string} email - The user's email address (must match step 1)
   * @param {string} ott - The 6-digit verification code from email
   * @returns {Promise<Object>} Result with encrypted user data for password verification
   */
  async verifyOTT(email, ott) {
    this.logger.log(`LoginService: Verifying OTT for email: ${email}`);

    try {
      // Ensure we're in the correct step of the login process
      if (this.loginState.step !== 1) {
        throw new Error("Please request a verification code first");
      }

      // Ensure email matches the one from step 1
      if (this.loginState.email !== email.toLowerCase().trim()) {
        throw new Error(
          "Email address must match the one used to request the verification code",
        );
      }

      // Validate OTT format
      if (!ott || ott.trim().length !== 6 || !/^\d{6}$/.test(ott.trim())) {
        throw new Error("Please enter a valid 6-digit verification code");
      }

      // Make API call to verify the OTT
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/verify-ott`,
        {
          email: email.toLowerCase().trim(),
          ott: ott.trim(),
        },
        {
          headers: { "Content-Type": "application/json" },
          timeout: 15000, // 15 second timeout (crypto operations can be slow)
        },
      );

      if (response.status === 200) {
        const data = response.data;

        // Store the encrypted user data and challenge for step 3
        this.loginState.challengeId = data.challengeId;
        this.loginState.userKeys = {
          salt: data.salt,
          kdfParams: data.kdf_params,
          publicKey: data.publicKey,
          encryptedMasterKey: data.encryptedMasterKey,
          encryptedPrivateKey: data.encryptedPrivateKey,
          encryptedChallenge: data.encryptedChallenge,
        };
        this.loginState.step = 2;

        this.logger.log(
          "LoginService: OTT verified successfully, received encrypted user data",
        );
        return {
          success: true,
          message: "Email verified. Please enter your password to continue.",
          requiresPassword: true,
          challengeId: data.challengeId,
          // Include additional metadata that might be useful for the UI
          lastPasswordChange: data.last_password_change,
          kdfParamsNeedUpgrade: data.kdf_params_need_upgrade,
          currentKeyVersion: data.current_key_version,
        };
      } else {
        throw new Error(`Unexpected response status: ${response.status}`);
      }
    } catch (error) {
      return this.handleLoginError("verifyOTT", error);
    }
  }

  /**
   * Step 3: Complete Login with Password Verification
   * This step verifies the password by decrypting the challenge and completes authentication
   *
   * @param {string} password - The user's password
   * @returns {Promise<Object>} Result with authentication tokens
   */
  async completeLogin(password) {
    this.logger.log(
      "LoginService: Completing login with password verification",
    );

    try {
      // Ensure we're in the correct step
      if (this.loginState.step !== 2) {
        throw new Error("Please complete email verification first");
      }

      // Ensure crypto service is ready
      await this.cryptoService.initSodium();

      // Step 3a: Derive the encryption key from the user's password
      this.logger.log("LoginService: Deriving encryption key from password...");
      const salt = this.cryptoService.fromBase64Url(
        this.loginState.userKeys.salt,
      );
      const keyEncryptionKey = await this.cryptoService.deriveKeyFromPassword(
        password,
        salt,
      );

      // Step 3b: Decrypt the master key using the password-derived key
      this.logger.log("LoginService: Decrypting master key...");
      const encryptedMasterKey = this.cryptoService.fromBase64Url(
        this.loginState.userKeys.encryptedMasterKey,
      );
      const masterKey = this.decryptWithNonce(
        encryptedMasterKey,
        keyEncryptionKey,
      );

      // Step 3c: Decrypt the private key using the master key
      this.logger.log("LoginService: Decrypting private key...");
      const encryptedPrivateKey = this.cryptoService.fromBase64Url(
        this.loginState.userKeys.encryptedPrivateKey,
      );
      const privateKey = this.decryptWithNonce(encryptedPrivateKey, masterKey);

      // Step 3d: Decrypt the authentication challenge using the private key
      this.logger.log("LoginService: Decrypting authentication challenge...");
      const encryptedChallenge = this.cryptoService.fromBase64Url(
        this.loginState.userKeys.encryptedChallenge,
      );
      const decryptedChallenge = this.decryptAsymmetric(
        encryptedChallenge,
        privateKey,
      );

      // Step 3e: Submit the decrypted challenge to complete authentication
      this.logger.log("LoginService: Submitting decrypted challenge...");
      const response = await axios.post(
        `${this.apiBaseUrl}/iam/api/v1/complete-login`,
        {
          email: this.loginState.email,
          challengeId: this.loginState.challengeId,
          decryptedData: this.cryptoService.toBase64Url(decryptedChallenge),
        },
        {
          headers: { "Content-Type": "application/json" },
          timeout: 15000,
        },
      );

      if (response.status === 200) {
        const tokenData = response.data;

        // Step 3f: Handle the authentication tokens
        let accessToken, refreshToken;

        if (tokenData.encrypted_tokens) {
          // Preferred method: decrypt the encrypted tokens
          this.logger.log("LoginService: Decrypting authentication tokens...");
          const encryptedTokens = this.cryptoService.fromBase64Url(
            tokenData.encrypted_tokens,
          );
          const decryptedTokensJson = this.decryptAsymmetric(
            encryptedTokens,
            privateKey,
          );
          const tokens = JSON.parse(
            new TextDecoder().decode(decryptedTokensJson),
          );

          accessToken = tokens.access_token;
          refreshToken = tokens.refresh_token;
        } else {
          // Fallback method: use plaintext tokens (less secure)
          this.logger.log(
            "LoginService: Using plaintext tokens (fallback mode)",
          );
          accessToken = tokenData.access_token;
          refreshToken = tokenData.refresh_token;
        }

        // Store the user's decrypted keys for this session
        const userSession = {
          email: this.loginState.email,
          masterKey: this.cryptoService.toBase64Url(masterKey),
          privateKey: this.cryptoService.toBase64Url(privateKey),
          publicKey: this.loginState.userKeys.publicKey,
        };

        // Clear the login state since we're done
        this.clearLoginState();

        this.logger.log("LoginService: Login completed successfully");
        return {
          success: true,
          message: "Login successful",
          tokens: {
            accessToken,
            refreshToken,
            accessTokenExpiry: tokenData.access_token_expiry_time,
            refreshTokenExpiry: tokenData.refresh_token_expiry_time,
          },
          userSession,
        };
      } else {
        throw new Error(`Unexpected response status: ${response.status}`);
      }
    } catch (error) {
      return this.handleLoginError("completeLogin", error);
    }
  }

  /**
   * Decrypt data that was encrypted with nonce prefix (ChaCha20-Poly1305)
   * This is used for symmetric encryption where nonce + ciphertext are concatenated
   */
  decryptWithNonce(encryptedData, key) {
    this.cryptoService.ensureReady();

    // Extract nonce and ciphertext
    const sodium = window.sodium;
    const nonceLength = sodium.crypto_secretbox_NONCEBYTES;
    const nonce = encryptedData.slice(0, nonceLength);
    const ciphertext = encryptedData.slice(nonceLength);

    // Decrypt the data
    const decrypted = sodium.crypto_secretbox_open_easy(ciphertext, nonce, key);
    if (!decrypted) {
      throw new Error(
        "Failed to decrypt data - incorrect password or corrupted data",
      );
    }

    return decrypted;
  }

  /**
   * Decrypt data using asymmetric encryption (X25519 + ChaCha20-Poly1305)
   * This is used for data encrypted with the user's public key
   */
  decryptAsymmetric(encryptedData, privateKey) {
    this.cryptoService.ensureReady();

    const sodium = window.sodium;

    // For box encryption, we need to extract the ephemeral public key and nonce
    const ephemeralPublicKeyLength = 32; // X25519 public key size
    const nonceLength = sodium.crypto_box_NONCEBYTES;

    const ephemeralPublicKey = encryptedData.slice(0, ephemeralPublicKeyLength);
    const nonce = encryptedData.slice(
      ephemeralPublicKeyLength,
      ephemeralPublicKeyLength + nonceLength,
    );
    const ciphertext = encryptedData.slice(
      ephemeralPublicKeyLength + nonceLength,
    );

    // Decrypt using the ephemeral public key and our private key
    const decrypted = sodium.crypto_box_open_easy(
      ciphertext,
      nonce,
      ephemeralPublicKey,
      privateKey,
    );
    if (!decrypted) {
      throw new Error(
        "Failed to decrypt asymmetric data - incorrect private key or corrupted data",
      );
    }

    return decrypted;
  }

  /**
   * Get the current login state
   * Useful for UI components to understand which step the user is on
   */
  getLoginState() {
    return {
      step: this.loginState.step,
      email: this.loginState.email,
      hasChallenge: !!this.loginState.challengeId,
    };
  }

  /**
   * Clear the current login state
   * This should be called when starting a new login or on errors
   */
  clearLoginState() {
    this.loginState = {
      email: null,
      challengeId: null,
      userKeys: null,
      step: 0,
    };
    this.logger.log("LoginService: Login state cleared");
  }

  /**
   * Validate email format
   */
  isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  /**
   * Handle login errors with appropriate user messages
   */
  handleLoginError(step, error) {
    this.logger.log(`LoginService: Error in ${step}: ${error.message}`);

    if (error.response) {
      const status = error.response.status;
      const data = error.response.data;

      if (status === 400 && data.details) {
        // Handle field-specific errors from the API
        const fieldErrors = data.details;
        const errorMessage = Object.values(fieldErrors)[0] || "Invalid request";

        return {
          success: false,
          message: errorMessage,
          fieldErrors: fieldErrors,
          step: this.loginState.step,
        };
      } else if (status === 400) {
        return {
          success: false,
          message: data.error || "Invalid request",
          step: this.loginState.step,
        };
      } else if (status === 429) {
        return {
          success: false,
          message:
            "Too many login attempts. Please wait a moment and try again.",
          step: this.loginState.step,
        };
      } else if (status >= 500) {
        return {
          success: false,
          message: "Server error. Please try again later.",
          step: this.loginState.step,
        };
      }
    } else if (error.request) {
      return {
        success: false,
        message:
          "Unable to connect to the server. Please check your internet connection.",
        step: this.loginState.step,
      };
    }

    // Handle client-side errors (like decryption failures)
    let userMessage = error.message;
    if (userMessage.includes("incorrect password")) {
      userMessage = "Incorrect password. Please try again.";
    } else if (userMessage.includes("corrupted data")) {
      userMessage = "Account data appears corrupted. Please contact support.";
    }

    return {
      success: false,
      message: userMessage,
      step: this.loginState.step,
    };
  }
}
