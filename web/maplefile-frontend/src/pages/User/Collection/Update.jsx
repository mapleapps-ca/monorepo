// File: src/pages/User/Collection/Update.jsx
import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const CollectionUpdate = () => {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const { collectionService, passwordStorageService } = useServices();
  const { isAuthenticated } = useAuth();

  // State for collection data
  const [collection, setCollection] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [loadError, setLoadError] = useState("");

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    collection_type: "folder",
  });

  // Password prompt state
  const [showPasswordPrompt, setShowPasswordPrompt] = useState(false);
  const [password, setPassword] = useState("");
  const [saving, setSaving] = useState(false);

  // Load collection on mount
  useEffect(() => {
    loadCollection();
  }, [collectionId]);

  const loadCollection = async () => {
    if (!collectionId) {
      setLoadError("No collection ID provided");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setLoadError("");

      // Get stored password
      const storedPassword = passwordStorageService.getPassword();

      // Load collection with password
      const collectionData = await collectionService.getCollection(
        collectionId,
        storedPassword,
      );

      setCollection(collectionData);

      // Set form data with current values
      setFormData({
        name: collectionData.name || "",
        collection_type: collectionData.collection_type || "folder",
      });

      console.log("[CollectionUpdate] Collection loaded:", collectionData);
    } catch (err) {
      console.error("[CollectionUpdate] Failed to load collection:", err);
      setLoadError(err.message || "Failed to load collection");
    } finally {
      setLoading(false);
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
      setError("You must be logged in to update collections");
      navigate("/login");
      return;
    }

    // Check if data has changed
    const hasChanges =
      formData.name !== collection.name ||
      formData.collection_type !== collection.collection_type;

    if (!hasChanges) {
      setError("No changes to save");
      return;
    }

    // Check if we have a stored password
    const storedPassword = passwordStorageService.getPassword();
    if (storedPassword) {
      console.log("[CollectionUpdate] Using stored password");
      await updateCollection(storedPassword);
    } else {
      console.log("[CollectionUpdate] No stored password, prompting user");
      setShowPasswordPrompt(true);
    }
  };

  // Update collection with password
  const updateCollection = async (passwordToUse) => {
    setSaving(true);
    setError("");

    try {
      console.log("[CollectionUpdate] Updating collection with password");

      // Prepare update data
      const updateData = {
        name: formData.name.trim(),
        collection_type: formData.collection_type,
        version: collection.version || 1, // Include version for optimistic locking
      };

      // Update via service (handles encryption internally)
      const updatedCollection = await collectionService.updateCollection(
        collectionId,
        updateData,
      );

      console.log(
        "[CollectionUpdate] Collection updated successfully:",
        updatedCollection,
      );

      // Navigate back to collection detail
      navigate(`/collections/${collectionId}`, {
        state: {
          message: `Collection "${updatedCollection.name}" updated successfully!`,
        },
      });
    } catch (err) {
      console.error("[CollectionUpdate] Failed to update collection:", err);

      if (err.message.includes("version conflict")) {
        setError(
          "Collection was modified by another user. Please refresh and try again.",
        );
      } else if (
        err.message.includes("Invalid password") ||
        err.message.includes("Decryption failed")
      ) {
        setError("Invalid password. Please try again.");
        setPassword("");
      } else if (err.message.includes("not initialized")) {
        setError("Encryption service not ready. Please wait and try again.");
      } else if (err.message.includes("admin permission")) {
        setError("You don't have permission to update this collection.");
      } else {
        setError(err.message || "Failed to update collection");
      }
    } finally {
      setSaving(false);
    }
  };

  // Handle password submission
  const handlePasswordSubmit = async (e) => {
    e.preventDefault();
    if (!password) {
      setError("Password is required");
      return;
    }
    await updateCollection(password);
  };

  // Handle cancel
  const handleCancel = () => {
    if (showPasswordPrompt) {
      setShowPasswordPrompt(false);
      setPassword("");
      setError("");
    } else {
      navigate(`/collections/${collectionId}`);
    }
  };

  // Loading state
  if (loading) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Loading Collection...</h1>
        <p>Please wait while we load the collection details...</p>
      </div>
    );
  }

  // Load error state
  if (loadError) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Error</h1>
        <p style={{ color: "red" }}>{loadError}</p>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
      </div>
    );
  }

  // Collection not found
  if (!collection) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Collection Not Found</h1>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
      </div>
    );
  }

  // Show password prompt if needed
  if (showPasswordPrompt) {
    return (
      <div style={{ padding: "20px", maxWidth: "600px" }}>
        <h1>Confirm Password</h1>

        <p>Please enter your password to update the encrypted collection.</p>

        {error && (
          <div style={{ color: "red", marginBottom: "10px" }}>
            Error: {error}
          </div>
        )}

        <form onSubmit={handlePasswordSubmit}>
          <div style={{ marginBottom: "15px" }}>
            <label htmlFor="password">Password:</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              required
              disabled={saving}
              autoFocus
              style={{ marginLeft: "10px", width: "250px" }}
            />
          </div>

          <div
            style={{
              marginTop: "20px",
              padding: "15px",
              backgroundColor: "#f8f9fa",
              border: "1px solid #ddd",
            }}
          >
            <h3>Why is this needed?</h3>
            <ul>
              <li>Your password is needed to decrypt your encryption keys</li>
              <li>Collection names are encrypted for your security</li>
              <li>Your password is never stored or transmitted</li>
            </ul>
          </div>

          <div style={{ marginTop: "20px" }}>
            <button type="submit" disabled={saving || !password}>
              {saving ? "Updating..." : "Update Collection"}
            </button>
            <button
              type="button"
              onClick={handleCancel}
              disabled={saving}
              style={{ marginLeft: "10px" }}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    );
  }

  // Main update form
  return (
    <div style={{ padding: "20px", maxWidth: "800px" }}>
      <h1>Update Collection</h1>

      <div style={{ marginBottom: "20px" }}>
        <button onClick={() => navigate(`/collections/${collectionId}`)}>
          ← Back to Collection
        </button>
      </div>

      {/* Current collection info */}
      <div
        style={{
          backgroundColor: "#f8f9fa",
          padding: "15px",
          marginBottom: "20px",
          border: "1px solid #ddd",
        }}
      >
        <h3>Current Collection Information</h3>
        <p>
          <strong>ID:</strong> {collection.id}
        </p>
        <p>
          <strong>Current Name:</strong> {collection.name || "[Encrypted]"}
        </p>
        <p>
          <strong>Current Type:</strong> {collection.collection_type}
        </p>
        <p>
          <strong>Created:</strong>{" "}
          {collection.created_at
            ? new Date(collection.created_at).toLocaleString()
            : "Unknown"}
        </p>
      </div>

      {error && (
        <div style={{ color: "red", marginBottom: "10px" }}>Error: {error}</div>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "15px" }}>
          <label htmlFor="name">Collection Name *</label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleChange}
            placeholder="Enter collection name"
            required
            disabled={saving}
            autoFocus
            style={{ width: "300px", marginLeft: "10px" }}
          />
          <small style={{ display: "block", color: "#666", marginTop: "5px" }}>
            This name will be encrypted before storage
          </small>
        </div>

        <div style={{ marginBottom: "15px" }}>
          <label htmlFor="collection_type">Collection Type *</label>
          <select
            id="collection_type"
            name="collection_type"
            value={formData.collection_type}
            onChange={handleChange}
            required
            disabled={saving}
            style={{ marginLeft: "10px" }}
          >
            <option value="folder">Folder</option>
            <option value="album">Album</option>
          </select>
          <small style={{ display: "block", color: "#666", marginTop: "5px" }}>
            Folders are for general files, Albums are optimized for photos/media
          </small>
        </div>

        {/* Warning for shared collections */}
        {collection.members && collection.members.length > 1 && (
          <div
            style={{
              marginTop: "20px",
              padding: "15px",
              backgroundColor: "#fff3cd",
              border: "1px solid #ffeaa7",
            }}
          >
            <h4>⚠️ Shared Collection Warning</h4>
            <p>
              This collection is shared with {collection.members.length - 1}{" "}
              other user(s). Changes will be visible to all members.
            </p>
          </div>
        )}

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
            <li>Collection names are encrypted using ChaCha20-Poly1305</li>
            <li>Only you can decrypt your collection names</li>
            <li>Your password will be required to save changes</li>
            <li>Version control prevents conflicting updates</li>
          </ul>
        </div>

        <div style={{ marginTop: "20px" }}>
          <button type="submit" disabled={saving || !formData.name.trim()}>
            {saving ? "Saving..." : "Save Changes"}
          </button>
          <button
            type="button"
            onClick={handleCancel}
            disabled={saving}
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
            <h4>Collection Data:</h4>
            <pre>
              {JSON.stringify(
                {
                  id: collection.id,
                  name: collection.name,
                  collection_type: collection.collection_type,
                  version: collection.version,
                  owner_id: collection.owner_id,
                  members_count: collection.members?.length || 0,
                },
                null,
                2,
              )}
            </pre>
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

export default withPasswordProtection(CollectionUpdate);
