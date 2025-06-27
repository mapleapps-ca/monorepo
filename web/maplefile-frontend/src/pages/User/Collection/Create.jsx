// Updated pages/User/Collection/Create.jsx
import React, { useState } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";

const CollectionCreate = () => {
  const navigate = useNavigate();
  const { collectionService, cryptoService, passwordStorageService } =
    useServices();
  const { isAuthenticated, user } = useAuth();

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    collection_type: "folder",
    parent_id: null,
    description: "",
  });

  const [password, setPassword] = useState("");
  const [showPasswordPrompt, setShowPasswordPrompt] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Handle form changes
  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    // Clear error when user makes changes
    if (error) {
      setError("");
    }
  };

  // Handle initial form submission (show password prompt)
  const handleSubmit = async () => {
    // Validate
    if (!formData.name.trim()) {
      setError("Collection name is required");
      return;
    }

    if (!isAuthenticated) {
      setError("You must be logged in to create collections");
      navigate("/login");
      return;
    }

    // Show password prompt
    setShowPasswordPrompt(true);
  };

  // Handle password submission and create collection
  const handlePasswordSubmit = async () => {
    if (!password) {
      setError("Password is required");
      return;
    }

    setLoading(true);
    setError("");

    try {
      console.log("[CollectionCreate] Creating collection with password");

      const fileId = await cryptoService.generateUUID();
      formData.id = fileId;

      // Create collection with password
      const newCollection =
        await collectionService.createCollectionWithPassword(
          {
            id: fileId,
            name: formData.name.trim(),
            collection_type: formData.collection_type,
            parent_id: formData.parent_id,
          },
          password,
        );

      console.log(
        "[CollectionCreate] Collection created successfully:",
        newCollection,
      );

      // Navigate to collections list or the new collection
      navigate("/collections", {
        state: {
          message: `Collection "${newCollection.name}" created successfully!`,
          newCollectionId: newCollection.id,
        },
      });
    } catch (err) {
      console.error("[CollectionCreate] Failed to create collection:", err);

      // Handle specific error cases
      if (
        err.message.includes("Invalid password") ||
        err.message.includes("Decryption failed")
      ) {
        setError("Invalid password. Please try again.");
        setPassword(""); // Clear password field
      } else if (err.message.includes("not initialized")) {
        setError("Encryption service not ready. Please wait and try again.");
      } else {
        setError(err.message || "Failed to create collection");
      }
    } finally {
      setLoading(false);
    }
  };

  // Handle cancel
  const handleCancel = () => {
    if (showPasswordPrompt) {
      setShowPasswordPrompt(false);
      setPassword("");
      setError("");
    } else {
      navigate("/collections");
    }
  };

  // Handle enter key press
  const handleKeyPress = (e, action) => {
    if (e.key === "Enter") {
      action();
    }
  };

  // Render password prompt
  if (showPasswordPrompt) {
    return (
      <div style={styles.container}>
        <h1>Confirm Password</h1>

        <div style={styles.card}>
          <p style={styles.info}>
            Please enter your password to create the encrypted collection.
          </p>

          {error && (
            <div style={styles.error}>
              <strong>Error:</strong> {error}
            </div>
          )}

          <div style={styles.formGroup}>
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              onKeyPress={(e) => handleKeyPress(e, handlePasswordSubmit)}
              placeholder="Enter your password"
              required
              disabled={loading}
              autoFocus
              style={styles.input}
            />
          </div>

          <div style={styles.securityInfo}>
            <h3>🔐 Why is this needed?</h3>
            <ul>
              <li>Your password is needed to decrypt your encryption keys</li>
              <li>Collections are encrypted end-to-end for your security</li>
              <li>Your password is never stored or transmitted</li>
              <li>Each collection has its own unique encryption key</li>
            </ul>
          </div>

          <div style={styles.buttonGroup}>
            <button
              onClick={handlePasswordSubmit}
              disabled={loading || !password}
              style={{
                ...styles.primaryButton,
                opacity: loading || !password ? 0.6 : 1,
              }}
            >
              {loading ? "Creating..." : "Create Collection"}
            </button>

            <button
              onClick={handleCancel}
              disabled={loading}
              style={{ ...styles.secondaryButton, opacity: loading ? 0.6 : 1 }}
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Render main form
  return (
    <div style={styles.container}>
      <h1>Create New Collection</h1>

      <div style={styles.navigation}>
        <button
          onClick={() => navigate("/collections")}
          style={styles.backButton}
        >
          ← Back to Collections
        </button>
      </div>

      <div style={styles.card}>
        <div style={styles.formSection}>
          <h2>Collection Details</h2>

          {error && (
            <div style={styles.error}>
              <strong>Error:</strong> {error}
            </div>
          )}

          <div style={styles.formGroup}>
            <label htmlFor="name">Collection Name *</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              onKeyPress={(e) =>
                e.key === "Enter" && formData.name.trim() && handleSubmit()
              }
              placeholder="Enter collection name"
              required
              disabled={loading}
              autoFocus
              style={styles.input}
            />
            <small style={styles.helpText}>
              This name will be encrypted before storage
            </small>
          </div>

          <div style={styles.formGroup}>
            <label htmlFor="collection_type">Collection Type *</label>
            <select
              id="collection_type"
              name="collection_type"
              value={formData.collection_type}
              onChange={handleChange}
              required
              disabled={loading}
              style={styles.select}
            >
              <option value="folder">Folder</option>
              <option value="album">Album</option>
            </select>
            <small style={styles.helpText}>
              Folders are for general files, Albums are optimized for
              photos/media
            </small>
          </div>

          <div style={styles.formGroup}>
            <label htmlFor="description">Description (Optional)</label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleChange}
              placeholder="Enter collection description"
              rows={3}
              disabled={loading}
              style={styles.textarea}
            />
            <small style={styles.helpText}>
              Note: Description is currently not encrypted
            </small>
          </div>
        </div>

        <div style={styles.securityInfo}>
          <h3>🔐 Security Information</h3>
          <ul>
            <li>Collection names are encrypted using ChaCha20-Poly1305</li>
            <li>Each collection has its own unique encryption key</li>
            <li>Collection keys are encrypted with your master key</li>
            <li>Only you can decrypt your collection names</li>
            <li>Your password will be required to create the collection</li>
          </ul>
        </div>

        <div style={styles.buttonGroup}>
          <button
            onClick={handleSubmit}
            disabled={loading || !formData.name.trim()}
            style={{
              ...styles.primaryButton,
              opacity: loading || !formData.name.trim() ? 0.6 : 1,
            }}
          >
            {loading ? "Creating..." : "Continue"}
          </button>

          <button
            onClick={handleCancel}
            disabled={loading}
            style={{ ...styles.secondaryButton, opacity: loading ? 0.6 : 1 }}
          >
            Cancel
          </button>
        </div>
      </div>

      {/* Debug info in development */}
      {import.meta.env.DEV && (
        <details style={styles.debug}>
          <summary>🔍 Debug Information</summary>
          <div style={styles.debugContent}>
            <h4>Form Data:</h4>
            <pre>{JSON.stringify(formData, null, 2)}</pre>

            <h4>Auth Status:</h4>
            <pre>{JSON.stringify({ isAuthenticated, user }, null, 2)}</pre>
          </div>
        </details>
      )}
    </div>
  );
};

const styles = {
  container: {
    padding: "20px",
    maxWidth: "800px",
    margin: "0 auto",
  },
  navigation: {
    marginBottom: "20px",
  },
  backButton: {
    padding: "8px 16px",
    background: "#f0f0f0",
    border: "1px solid #ddd",
    borderRadius: "4px",
    cursor: "pointer",
  },
  card: {
    background: "white",
    border: "1px solid #e0e0e0",
    borderRadius: "8px",
    padding: "30px",
    boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
  },
  formSection: {
    marginBottom: "30px",
  },
  formGroup: {
    marginBottom: "20px",
  },
  input: {
    width: "100%",
    padding: "10px",
    fontSize: "16px",
    border: "1px solid #ddd",
    borderRadius: "4px",
    marginTop: "5px",
    display: "block",
  },
  select: {
    width: "100%",
    padding: "10px",
    fontSize: "16px",
    border: "1px solid #ddd",
    borderRadius: "4px",
    marginTop: "5px",
    background: "white",
    display: "block",
  },
  textarea: {
    width: "100%",
    padding: "10px",
    fontSize: "16px",
    border: "1px solid #ddd",
    borderRadius: "4px",
    marginTop: "5px",
    resize: "vertical",
    display: "block",
  },
  helpText: {
    display: "block",
    marginTop: "5px",
    fontSize: "14px",
    color: "#666",
  },
  error: {
    background: "#fee",
    border: "1px solid #fcc",
    color: "#c00",
    padding: "10px",
    borderRadius: "4px",
    marginBottom: "20px",
  },
  info: {
    marginBottom: "20px",
    fontSize: "16px",
    color: "#333",
  },
  securityInfo: {
    background: "#f8f9fa",
    border: "1px solid #e9ecef",
    borderRadius: "4px",
    padding: "20px",
    marginBottom: "30px",
  },
  buttonGroup: {
    display: "flex",
    gap: "10px",
    justifyContent: "flex-end",
  },
  primaryButton: {
    padding: "10px 20px",
    background: "#007bff",
    color: "white",
    border: "none",
    borderRadius: "4px",
    fontSize: "16px",
    cursor: "pointer",
  },
  secondaryButton: {
    padding: "10px 20px",
    background: "#6c757d",
    color: "white",
    border: "none",
    borderRadius: "4px",
    fontSize: "16px",
    cursor: "pointer",
  },
  debug: {
    marginTop: "40px",
    padding: "20px",
    background: "#f8f9fa",
    border: "1px solid #dee2e6",
    borderRadius: "4px",
  },
  debugContent: {
    marginTop: "10px",
  },
};

export default CollectionCreate;
