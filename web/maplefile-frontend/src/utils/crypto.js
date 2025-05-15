// src/utils/crypto.js
import _sodium_original_module_name_to_avoid_global_conflicts from "libsodium-wrappers-sumo";

let sodiumInstance = null;
let sodiumReadyPromise = null; // To ensure .ready is called only once

/**
 * Initializes the sodium library if it hasn't been already.
 * Returns a promise that resolves with the initialized sodium instance.
 */
export const initSodium = async () => {
  if (sodiumInstance) {
    return sodiumInstance;
  }

  if (!sodiumReadyPromise) {
    console.log("Sodium: Starting initialization (_sodium.ready)...");
    sodiumReadyPromise = (async () => {
      try {
        await _sodium_original_module_name_to_avoid_global_conflicts.ready;
        // Check if the actual library is under a .default property
        sodiumInstance =
          _sodium_original_module_name_to_avoid_global_conflicts.default ||
          _sodium_original_module_name_to_avoid_global_conflicts;
        console.log(
          "Sodium: Initialization complete. sodiumInstance is now set.",
        );
        if (typeof sodiumInstance.crypto_pwhash !== "function") {
          console.error(
            "Sodium critical error: crypto_pwhash is NOT a function even after .ready!",
          );
          // Log available keys for debugging if crypto_pwhash is missing
          // console.log("Available keys on sodiumInstance:", Object.keys(sodiumInstance));
        }
        return sodiumInstance;
      } catch (err) {
        console.error("Sodium: Initialization FAILED", err);
        sodiumReadyPromise = null; // Reset promise to allow re-attempt if it failed
        throw err; // Re-throw to propagate the error
      }
    })();
  }
  return sodiumReadyPromise;
};

// Helper to ensure sodium is ready before any crypto operation
const ensureSodium = async () => {
  if (!sodiumInstance) {
    // This will either return the existing promise or start a new initialization
    return await initSodium();
  }
  return sodiumInstance;
};

export const cryptoUtils = {
  generateRandomBytes: async (length) => {
    const sodium = await ensureSodium();
    return sodium.randombytes_buf(length);
  },

  deriveKeyFromPassword: async (password, salt) => {
    const sodium = await ensureSodium();

    console.log(
      "In deriveKeyFromPassword - sodium object available:",
      !!sodium,
    );
    console.log("typeof sodium.crypto_pwhash:", typeof sodium.crypto_pwhash);
    console.log(
      "Constants check: KEYBYTES",
      sodium.crypto_secretbox_KEYBYTES,
      "SALTBYTES",
      sodium.crypto_pwhash_SALTBYTES,
      "OPSLIMIT",
      sodium.crypto_pwhash_OPSLIMIT_INTERACTIVE,
      "MEMLIMIT",
      sodium.crypto_pwhash_MEMLIMIT_INTERACTIVE,
      "ALG_DEFAULT",
      sodium.crypto_pwhash_ALG_DEFAULT,
    );

    if (typeof password !== "string" || password.length === 0) {
      throw new Error("Password must be a non-empty string.");
    }
    if (
      !(salt instanceof Uint8Array) ||
      salt.length !== sodium.crypto_pwhash_SALTBYTES
    ) {
      const saltBytesLen = sodium.crypto_pwhash_SALTBYTES || "NOT_DEFINED";
      throw new Error(
        `Salt must be a Uint8Array of length ${saltBytesLen}. Got length ${salt?.length}`,
      );
    }

    if (typeof sodium.crypto_pwhash !== "function") {
      console.error(
        "CRITICAL: sodium.crypto_pwhash is not a function inside deriveKeyFromPassword even after ensureSodium.",
      );
      // Attempt to log available properties for diagnosis
      // console.error('Available sodium properties:', Object.keys(sodium));
      throw new Error(
        "sodium.crypto_pwhash is not a function. Sodium library may not be initialized correctly.",
      );
    }

    return sodium.crypto_pwhash(
      sodium.crypto_secretbox_KEYBYTES,
      password, // libsodium-wrappers' crypto_pwhash can take string directly
      salt,
      sodium.crypto_pwhash_OPSLIMIT_INTERACTIVE,
      sodium.crypto_pwhash_MEMLIMIT_INTERACTIVE,
      sodium.crypto_pwhash_ALG_DEFAULT,
    );
  },

  encryptWithKey: async (data /* Uint8Array */, key /* Uint8Array */) => {
    const sodium = await ensureSodium();
    if (!(data instanceof Uint8Array))
      throw new Error("Data to encrypt must be Uint8Array");
    if (
      !(key instanceof Uint8Array) ||
      key.length !== sodium.crypto_secretbox_KEYBYTES
    ) {
      throw new Error(
        `Encryption key must be a Uint8Array of length ${sodium.crypto_secretbox_KEYBYTES}. Got length ${key?.length}`,
      );
    }
    const nonce = sodium.randombytes_buf(sodium.crypto_secretbox_NONCEBYTES);
    const ciphertext = sodium.crypto_secretbox_easy(data, nonce, key);
    return { nonce, ciphertext };
  },

  decryptWithKey: async (
    ciphertext /* Uint8Array */,
    nonce /* Uint8Array */,
    key /* Uint8Array */,
  ) => {
    const sodium = await ensureSodium();
    if (!(ciphertext instanceof Uint8Array))
      throw new Error("Ciphertext must be Uint8Array");
    if (
      !(nonce instanceof Uint8Array) ||
      nonce.length !== sodium.crypto_secretbox_NONCEBYTES
    ) {
      throw new Error(
        `Nonce must be a Uint8Array of length ${sodium.crypto_secretbox_NONCEBYTES}. Got length ${nonce?.length}`,
      );
    }
    if (
      !(key instanceof Uint8Array) ||
      key.length !== sodium.crypto_secretbox_KEYBYTES
    ) {
      throw new Error(
        `Decryption key must be a Uint8Array of length ${sodium.crypto_secretbox_KEYBYTES}. Got length ${key?.length}`,
      );
    }
    const decrypted = sodium.crypto_secretbox_open_easy(ciphertext, nonce, key);
    if (!decrypted)
      throw new Error(
        "Decryption failed (crypto_secretbox_open_easy). Invalid key, nonce, or ciphertext.",
      );
    return decrypted;
  },

  encryptWithBoxSeal: async (
    data /* Uint8Array */,
    publicKey /* Uint8Array */,
  ) => {
    const sodium = await ensureSodium();
    if (!(data instanceof Uint8Array))
      throw new Error("Data for box_seal must be Uint8Array");
    if (
      !(publicKey instanceof Uint8Array) ||
      publicKey.length !== sodium.crypto_box_PUBLICKEYBYTES
    ) {
      throw new Error(
        `Public key for box_seal must be Uint8Array of length ${sodium.crypto_box_PUBLICKEYBYTES}`,
      );
    }
    return sodium.crypto_box_seal(data, publicKey);
  },

  decryptWithBoxSealOpen: async (
    sealedCiphertext /* Uint8Array */,
    publicKey /* Uint8Array */,
    privateKey /* Uint8Array */,
  ) => {
    const sodium = await ensureSodium();
    if (!(sealedCiphertext instanceof Uint8Array))
      throw new Error("Sealed ciphertext must be Uint8Array");
    if (
      !(publicKey instanceof Uint8Array) ||
      publicKey.length !== sodium.crypto_box_PUBLICKEYBYTES
    ) {
      throw new Error(
        `Public key for box_seal_open must be Uint8Array of length ${sodium.crypto_box_PUBLICKEYBYTES}`,
      );
    }
    if (
      !(privateKey instanceof Uint8Array) ||
      privateKey.length !== sodium.crypto_box_SECRETKEYBYTES
    ) {
      throw new Error(
        `Private key for box_seal_open must be Uint8Array of length ${sodium.crypto_box_SECRETKEYBYTES}`,
      );
    }
    const decrypted = sodium.crypto_box_seal_open(
      sealedCiphertext,
      publicKey,
      privateKey,
    );
    if (!decrypted) throw new Error("Box seal open decryption failed.");
    return decrypted;
  },

  toBase64: async (
    bytes,
    variant = _sodium_original_module_name_to_avoid_global_conflicts
      .base64_variants.URLSAFE_NO_PADDING,
  ) => {
    const sodium = await ensureSodium();
    if (!(bytes instanceof Uint8Array))
      throw new Error("Input must be Uint8Array for toBase64.");
    return sodium.to_base64(bytes, variant);
  },

  fromBase64: async (
    base64Str,
    variant = _sodium_original_module_name_to_avoid_global_conflicts
      .base64_variants.URLSAFE_NO_PADDING,
  ) => {
    const sodium = await ensureSodium();
    if (typeof base64Str !== "string")
      throw new Error("Input must be a string for fromBase64.");
    try {
      return sodium.from_base64(base64Str, variant);
    } catch (e) {
      if (variant !== sodium.base64_variants.ORIGINAL) {
        try {
          console.warn(
            "fromBase64: Retrying with ORIGINAL variant for:",
            base64Str.substring(0, 10) + "...",
          );
          return sodium.from_base64(base64Str, sodium.base64_variants.ORIGINAL);
        } catch (e2) {
          throw new Error(
            `Failed to decode base64 string with ${variant} and ORIGINAL variants: ${e2.message}`,
          );
        }
      }
      throw e;
    }
  },

  stringToBytes: async (str) => {
    const sodium = await ensureSodium();
    if (typeof str !== "string")
      throw new Error("Input must be a string for stringToBytes.");
    return sodium.from_string(str);
  },

  bytesToString: async (bytes) => {
    const sodium = await ensureSodium();
    if (!(bytes instanceof Uint8Array))
      throw new Error("Input must be Uint8Array for bytesToString.");
    return sodium.to_string(bytes);
  },

  generateKeyPair: async () => {
    const sodium = await ensureSodium();
    return sodium.crypto_box_keypair();
  },

  combineNonceAndCiphertext: (nonce, ciphertext) => {
    if (!(nonce instanceof Uint8Array) || !(ciphertext instanceof Uint8Array)) {
      throw new Error("Nonce and Ciphertext must be Uint8Array.");
    }
    const combined = new Uint8Array(nonce.length + ciphertext.length);
    combined.set(nonce, 0);
    combined.set(ciphertext, nonce.length);
    return combined;
  },

  splitNonceAndCiphertext: async (combined) => {
    const sodium = await ensureSodium(); // For sodium.crypto_secretbox_NONCEBYTES
    const nonceLength = sodium.crypto_secretbox_NONCEBYTES;
    if (!(combined instanceof Uint8Array) || combined.length < nonceLength) {
      throw new Error(
        `Combined data is invalid or too short for nonce splitting. Min length: ${nonceLength}, Got: ${combined?.length}`,
      );
    }
    const nonce = combined.slice(0, nonceLength);
    const ciphertext = combined.slice(nonceLength);
    return { nonce, ciphertext };
  },
};

export default cryptoUtils;
