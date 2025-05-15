// monorepo/web/prototyping/maplefile-cli/src/pages/Register.jsx
import { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import axios from "axios";
import _sodium from "libsodium-wrappers-sumo";
import { authAPI } from "../services/api";
import { cryptoUtils } from "../utils/crypto";

function Register() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    beta_access_code: "",
    first_name: "",
    last_name: "",
    email: "",
    phone: "",
    country: "",
    timezone: "America/Toronto",
    password: "",
    confirmPassword: "",
    agree_terms_of_service: false,
    agree_promotions: false,
    agree_to_tracking_across_third_party_apps_and_services: false,
    module: 1, // MAPLEFILE
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [sodium, setSodium] = useState(null);

  // Initialize sodium when component mounts
  useEffect(() => {
    const initSodium = async () => {
      try {
        await _sodium.ready;
        console.log("Sodium initialized successfully");
        setSodium(_sodium);
      } catch (err) {
        console.error("Failed to initialize sodium:", err);
        setError(
          "Failed to initialize encryption library. Please try again later.",
        );
      }
    };

    initSodium();
  }, []);

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!sodium) {
      setError(
        "Encryption library is not ready. Please wait or refresh the page.",
      );
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // Password validation
      if (formData.password !== formData.confirmPassword) {
        throw new Error("Passwords don't match");
      }

      if (formData.password.length < 8) {
        throw new Error("Password must be at least 8 characters");
      }

      console.log("Starting encryption process...");

      // Generate random salt (16 bytes for Argon2)
      const salt = sodium.randombytes_buf(16);
      console.log("Salt generated");

      // Derive key from password - for JavaScript we'll use a simpler approach with BLAKE2b
      const passwordBytes = sodium.from_string(formData.password);
      const keyEncryptionKey = await cryptoUtils.deriveKeyFromPassword(
        formData.password, // Pass the string password
        salt, // Pass the Uint8Array salt
      );
      console.log("Key encryption key derived (using crypto_pwhash)");

      // Generate master key
      const masterKey = sodium.randombytes_buf(32);
      console.log("Master key generated");

      // Generate key pair for asymmetric encryption
      const keyPair = sodium.crypto_box_keypair();
      const publicKey = keyPair.publicKey;
      const privateKey = keyPair.privateKey;
      console.log("Key pair generated");

      // Generate recovery key
      const recoveryKey = sodium.randombytes_buf(32);
      console.log("Recovery key generated");

      // Encrypt master key with key encryption key
      const encryptedMasterKeyNonce = sodium.randombytes_buf(
        sodium.crypto_secretbox_NONCEBYTES,
      );
      const encryptedMasterKey = sodium.crypto_secretbox_easy(
        masterKey,
        encryptedMasterKeyNonce,
        keyEncryptionKey,
      );
      console.log("Master key encrypted");

      // Encrypt private key with master key
      const encryptedPrivateKeyNonce = sodium.randombytes_buf(
        sodium.crypto_secretbox_NONCEBYTES,
      );
      const encryptedPrivateKey = sodium.crypto_secretbox_easy(
        privateKey,
        encryptedPrivateKeyNonce,
        masterKey,
      );
      console.log("Private key encrypted");

      // Encrypt recovery key with master key
      const encryptedRecoveryKeyNonce = sodium.randombytes_buf(
        sodium.crypto_secretbox_NONCEBYTES,
      );
      const encryptedRecoveryKey = sodium.crypto_secretbox_easy(
        recoveryKey,
        encryptedRecoveryKeyNonce,
        masterKey,
      );
      console.log("Recovery key encrypted");

      // Encrypt master key with recovery key
      const masterKeyEncryptedWithRecoveryKeyNonce = sodium.randombytes_buf(
        sodium.crypto_secretbox_NONCEBYTES,
      );
      const masterKeyEncryptedWithRecoveryKeyCiphertext =
        sodium.crypto_secretbox_easy(
          masterKey,
          masterKeyEncryptedWithRecoveryKeyNonce,
          recoveryKey,
        );
      console.log("Master key encrypted with recovery key");

      // Generate verification ID from public key (simple hash approach)
      const publicKeyHash = sodium.crypto_generichash(32, publicKey);
      const verificationID = sodium.to_base64(publicKeyHash).slice(0, 12);
      console.log("Verification ID generated");

      // Create registration payload
      const registrationData = {
        ...formData,
        salt: sodium.to_base64(salt),
        publicKey: sodium.to_base64(publicKey),
        encryptedMasterKey: sodium.to_base64(
          new Uint8Array([...encryptedMasterKeyNonce, ...encryptedMasterKey]),
        ),
        encryptedPrivateKey: sodium.to_base64(
          new Uint8Array([...encryptedPrivateKeyNonce, ...encryptedPrivateKey]),
        ),
        encryptedRecoveryKey: sodium.to_base64(
          new Uint8Array([
            ...encryptedRecoveryKeyNonce,
            ...encryptedRecoveryKey,
          ]),
        ),
        masterKeyEncryptedWithRecoveryKey: sodium.to_base64(
          new Uint8Array([
            ...masterKeyEncryptedWithRecoveryKeyNonce,
            ...masterKeyEncryptedWithRecoveryKeyCiphertext,
          ]),
        ),
        verificationID: verificationID,
      };

      // Remove confirmPassword as it's not needed in the API
      delete registrationData.confirmPassword;

      console.log("Registration payload created, sending to API...");

      // Send registration request to API using our service
      await authAPI.register(registrationData);

      alert(
        "Registration successful! Please check your email to verify your account.",
      );
      navigate("/login");
    } catch (err) {
      console.error("Registration error:", err);
      setError(
        err.response?.data?.message || err.message || "Registration failed",
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1>Register</h1>
      {error && <p>{error}</p>}
      {!sodium && <p>Initializing security... Please wait.</p>}

      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="beta_access_code">Beta Access Code:</label>
          <input
            type="text"
            id="beta_access_code"
            name="beta_access_code"
            value={formData.beta_access_code}
            onChange={handleChange}
            required
          />
        </div>

        <div>
          <label htmlFor="first_name">First Name:</label>
          <input
            type="text"
            id="first_name"
            name="first_name"
            value={formData.first_name}
            onChange={handleChange}
            required
          />
        </div>

        <div>
          <label htmlFor="last_name">Last Name:</label>
          <input
            type="text"
            id="last_name"
            name="last_name"
            value={formData.last_name}
            onChange={handleChange}
            required
          />
        </div>

        <div>
          <label htmlFor="email">Email:</label>
          <input
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            required
          />
        </div>

        <div>
          <label htmlFor="phone">Phone:</label>
          <input
            type="tel"
            id="phone"
            name="phone"
            value={formData.phone}
            onChange={handleChange}
            required
          />
        </div>

        <div>
          <label htmlFor="country">Country:</label>
          <input
            type="text"
            id="country"
            name="country"
            value={formData.country}
            onChange={handleChange}
            required
          />
        </div>

        <div>
          <label htmlFor="timezone">Timezone:</label>
          <input
            type="text"
            id="timezone"
            name="timezone"
            value={formData.timezone}
            onChange={handleChange}
            required
          />
        </div>

        <div>
          <label htmlFor="password">Password:</label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            required
            minLength={8}
          />
        </div>

        <div>
          <label htmlFor="confirmPassword">Confirm Password:</label>
          <input
            type="password"
            id="confirmPassword"
            name="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            required
            minLength={8}
          />
        </div>

        <div>
          <label>
            <input
              type="checkbox"
              name="agree_terms_of_service"
              checked={formData.agree_terms_of_service}
              onChange={handleChange}
              required
            />
            I agree to the Terms of Service
          </label>
        </div>

        <div>
          <label>
            <input
              type="checkbox"
              name="agree_promotions"
              checked={formData.agree_promotions}
              onChange={handleChange}
            />
            I agree to receive promotional emails
          </label>
        </div>

        <div>
          <label>
            <input
              type="checkbox"
              name="agree_to_tracking_across_third_party_apps_and_services"
              checked={
                formData.agree_to_tracking_across_third_party_apps_and_services
              }
              onChange={handleChange}
            />
            I agree to tracking across third party apps and services
          </label>
        </div>

        <button type="submit" disabled={loading || !sodium}>
          {loading ? "Registering..." : "Register"}
        </button>
      </form>
    </div>
  );
}

export default Register;
