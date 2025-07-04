// File: src/pages/User/Collection/Create.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const CollectionCreate = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { collectionService, cryptoService, passwordStorageService } =
    useServices();
  const { isAuthenticated } = useAuth();

  // Get parent ID from query string
  const searchParams = new URLSearchParams(location.search);
  const parentId = searchParams.get("parent");

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    collection_type: "folder",
    parent_id: parentId || null,
    description: "",
  });

  const [parentCollection, setParentCollection] = useState(null);
  const [password, setPassword] = useState("");
  const [showPasswordPrompt, setShowPasswordPrompt] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Load parent collection info if parentId is provided
  useEffect(() => {
    if (parentId) {
      loadParentCollection();
    }
  }, [parentId]);

  const loadParentCollection = async () => {
    try {
      const storedPassword = passwordStorageService.getPassword();
      if (storedPassword) {
        const parent = await collectionService.getCollection(
          parentId,
          storedPassword,
        );
        setParentCollection(parent);
      }
    } catch (err) {
      console.warn("Could not load parent collection:", err);
    }
  };

  // Handle form changes
  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
    if (error) setError("");
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();

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

    console.log(
      "[DEBUG] Password service has password:",
      passwordStorageService.hasPassword(),
    );

    // Check if we already have a stored password
    const storedPassword = passwordStorageService.getPassword();
    if (storedPassword) {
      console.log("[CollectionCreate] Using stored password");
      await createCollection(storedPassword);
    } else {
      console.log("[CollectionCreate] No stored password, prompting user");
      setShowPasswordPrompt(true);
    }
  };

  // Create collection with password
  const createCollection = async (passwordToUse) => {
    setLoading(true);
    setError("");

    try {
      console.log("[CollectionCreate] Creating collection with password");

      const fileId = await cryptoService.generateUUID();
      const collectionData = {
        id: fileId,
        name: formData.name.trim(),
        collection_type: formData.collection_type,
        parent_id: formData.parent_id,
      };

      const newCollection =
        await collectionService.createCollectionWithPassword(
          collectionData,
          passwordToUse,
        );

      console.log(
        "[CollectionCreate] Collection created successfully:",
        newCollection,
      );

      // Navigate to file manager at the parent folder
      if (parentId) {
        navigate(`/files/${parentId}`, {
          state: {
            message: `Folder "${newCollection.name}" created successfully!`,
            newCollectionId: newCollection.id,
          },
        });
      } else {
        navigate("/files", {
          state: {
            message: `Folder "${newCollection.name}" created successfully!`,
            newCollectionId: newCollection.id,
          },
        });
      }
    } catch (err) {
      console.error("[CollectionCreate] Failed to create collection:", err);

      if (
        err.message.includes("Invalid password") ||
        err.message.includes("Decryption failed")
      ) {
        setError("Invalid password. Please try again.");
        setPassword("");
      } else if (err.message.includes("not initialized")) {
        setError("Encryption service not ready. Please wait and try again.");
      } else {
        setError(err.message || "Failed to create collection");
      }
    } finally {
      setLoading(false);
    }
  };

  // Handle password submission
  const handlePasswordSubmit = async (e) => {
    e.preventDefault();
    if (!password) {
      setError("Password is required");
      return;
    }
    await createCollection(password);
  };

  // Handle cancel
  const handleCancel = () => {
    if (showPasswordPrompt) {
      setShowPasswordPrompt(false);
      setPassword("");
      setError("");
    } else {
      // Navigate back to file manager
      if (parentId) {
        navigate(`/files/${parentId}`);
      } else {
        navigate("/files");
      }
    }
  };

  // Show password prompt if needed
  if (showPasswordPrompt) {
    return (
      <div>
        <h1>Confirm Password</h1>

        <p>Please enter your password to create the encrypted collection.</p>

        {error && (
          <div style={{ color: "red", marginBottom: "10px" }}>
            Error: {error}
          </div>
        )}

        <form onSubmit={handlePasswordSubmit}>
          <div>
            <label htmlFor="password">Password:</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              required
              disabled={loading}
              autoFocus
            />
          </div>

          <div style={{ marginTop: "20px" }}>
            <h3>Why is this needed?</h3>
            <ul>
              <li>Your password is needed to decrypt your encryption keys</li>
              <li>Collections are encrypted end-to-end for your security</li>
              <li>Your password is never stored or transmitted</li>
              <li>Each collection has its own unique encryption key</li>
            </ul>
          </div>

          <div style={{ marginTop: "20px" }}>
            <button type="submit" disabled={loading || !password}>
              {loading ? "Creating..." : "Create Collection"}
            </button>
            <button
              type="button"
              onClick={handleCancel}
              disabled={loading}
              style={{ marginLeft: "10px" }}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    );
  }

  // Main form
  return (
    <div>
      <h1>Create New Folder</h1>

      <div style={{ marginBottom: "20px" }}>
        <button onClick={handleCancel}>← Back to File Manager</button>
      </div>

      {parentCollection && (
        <div
          style={{
            backgroundColor: "#f8f9fa",
            padding: "15px",
            marginBottom: "20px",
            borderRadius: "4px",
          }}
        >
          <strong>Creating folder in:</strong>{" "}
          {parentCollection.name || "[Encrypted]"}
        </div>
      )}

      {error && (
        <div style={{ color: "red", marginBottom: "10px" }}>Error: {error}</div>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "15px" }}>
          <label htmlFor="name">Folder Name *</label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleChange}
            placeholder="Enter folder name"
            required
            disabled={loading}
            autoFocus
            style={{ width: "300px", marginLeft: "10px" }}
          />
          <small style={{ display: "block", color: "#666" }}>
            This name will be encrypted before storage
          </small>
        </div>

        <div style={{ marginBottom: "15px" }}>
          <label htmlFor="collection_type">Folder Type *</label>
          <select
            id="collection_type"
            name="collection_type"
            value={formData.collection_type}
            onChange={handleChange}
            required
            disabled={loading}
            style={{ marginLeft: "10px" }}
          >
            <option value="folder">Standard Folder</option>
            <option value="album">Photo Album</option>
          </select>
          <small style={{ display: "block", color: "#666" }}>
            Standard folders are for general files, Photo albums are optimized
            for images
          </small>
        </div>

        <div style={{ marginBottom: "15px" }}>
          <label htmlFor="description">Description (Optional)</label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleChange}
            placeholder="Enter folder description"
            rows={3}
            disabled={loading}
            style={{ width: "300px", marginLeft: "10px", display: "block" }}
          />
          <small style={{ color: "#666" }}>
            Note: Description is currently not encrypted
          </small>
        </div>

        <div
          style={{
            marginTop: "20px",
            padding: "15px",
            backgroundColor: "#f8f9fa",
            border: "1px solid #ddd",
          }}
        >
          <h3>🔐 Security Information</h3>
          <ul>
            <li>Folder names are encrypted using ChaCha20-Poly1305</li>
            <li>Each folder has its own unique encryption key</li>
            <li>Folder keys are encrypted with your master key</li>
            <li>Only you can decrypt your folder names</li>
            <li>Your password will be required to create the folder</li>
          </ul>
        </div>

        <div style={{ marginTop: "20px" }}>
          <button type="submit" disabled={loading || !formData.name.trim()}>
            {loading ? "Creating..." : "Continue"}
          </button>
          <button
            type="button"
            onClick={handleCancel}
            disabled={loading}
            style={{ marginLeft: "10px" }}
          >
            Cancel
          </button>
        </div>
      </form>

      {/* Debug info in development */}
      {import.meta.env.DEV && (
        <details
          style={{
            marginTop: "40px",
            padding: "10px",
            backgroundColor: "#f8f9fa",
          }}
        >
          <summary>🔍 Debug Information</summary>
          <div>
            <h4>Form Data:</h4>
            <pre>{JSON.stringify(formData, null, 2)}</pre>
            <h4>Password Service Status:</h4>
            <pre>
              {JSON.stringify(
                {
                  hasPassword: passwordStorageService.hasPassword(),
                  isAuthenticated,
                },
                null,
                2,
              )}
            </pre>
          </div>
        </details>
      )}
    </div>
  );
};

const ProtectedCollectionCreate = withPasswordProtection(CollectionCreate);

export default ProtectedCollectionCreate;
