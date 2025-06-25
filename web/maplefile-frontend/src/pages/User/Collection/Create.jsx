// Updated pages/User/Collection/Create.jsx
import React, { useState } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";

const CollectionCreate = () => {
  const navigate = useNavigate();
  const { collectionService } = useServices();
  const { isAuthenticated, user } = useAuth();

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    collection_type: "folder",
    parent_id: null,
    description: "", // Note: description would also need encryption in production
  });

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

    setLoading(true);
    setError("");

    try {
      console.log("[CollectionCreate] Creating collection:", formData);

      // Create collection with encryption handled by service
      const newCollection = await collectionService.createCollection({
        name: formData.name.trim(),
        collection_type: formData.collection_type,
        parent_id: formData.parent_id,
        // Add any additional metadata that should be encrypted
        // description: formData.description, // Would need separate encryption
      });

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
      if (err.message.includes("User encryption keys not available")) {
        setError(
          "Encryption keys not available. Please log out and log in again.",
        );
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
    navigate("/collections");
  };

  return (
    <div>
      <h1>Create New Collection</h1>

      <div>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
      </div>

      <div>
        <form onSubmit={handleSubmit}>
          <div>
            <h2>Collection Details</h2>

            {error && (
              <div style={{ color: "red", marginBottom: "1rem" }}>
                <strong>Error:</strong> {error}
              </div>
            )}

            <div>
              <label htmlFor="name">Collection Name *</label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleChange}
                placeholder="Enter collection name"
                required
                disabled={loading}
                autoFocus
              />
              <small>This name will be encrypted before storage</small>
            </div>

            <div>
              <label htmlFor="collection_type">Collection Type *</label>
              <select
                id="collection_type"
                name="collection_type"
                value={formData.collection_type}
                onChange={handleChange}
                required
                disabled={loading}
              >
                <option value="folder">Folder</option>
                <option value="album">Album</option>
              </select>
              <small>
                Folders are for general files, Albums are optimized for
                photos/media
              </small>
            </div>

            {/* TODO: Add parent collection selector */}
            {/*
            <div>
              <label htmlFor="parent_id">
                Parent Collection (Optional)
              </label>
              <select
                id="parent_id"
                name="parent_id"
                value={formData.parent_id || ""}
                onChange={handleChange}
                disabled={loading}
              >
                <option value="">Root Level</option>
                {parentCollections.map(collection => (
                  <option key={collection.id} value={collection.id}>
                    {collection.name}
                  </option>
                ))}
              </select>
            </div>
            */}

            <div>
              <label htmlFor="description">Description (Optional)</label>
              <textarea
                id="description"
                name="description"
                value={formData.description}
                onChange={handleChange}
                placeholder="Enter collection description"
                rows={3}
                disabled={loading}
              />
              <small>Note: Description is currently not encrypted</small>
            </div>
          </div>

          <div>
            <h3>🔐 Security Information</h3>
            <ul>
              <li>Collection names are encrypted using ChaCha20-Poly1305</li>
              <li>Each collection has its own unique encryption key</li>
              <li>Collection keys are encrypted with your master key</li>
              <li>Only you can decrypt your collection names</li>
              <li>
                Sharing a collection requires re-encrypting the key for
                recipients
              </li>
            </ul>
          </div>

          <div>
            <button type="submit" disabled={loading || !formData.name.trim()}>
              {loading ? "Creating..." : "Create Collection"}
            </button>

            <button type="button" onClick={handleCancel} disabled={loading}>
              Cancel
            </button>
          </div>
        </form>
      </div>

      {/* Debug info in development */}
      {import.meta.env.DEV && (
        <details style={{ marginTop: "2rem" }}>
          <summary>🔍 Debug Information</summary>
          <div>
            <h4>Form Data:</h4>
            <pre>{JSON.stringify(formData, null, 2)}</pre>

            <h4>Auth Status:</h4>
            <pre>{JSON.stringify({ isAuthenticated, user }, null, 2)}</pre>

            <h4>Collection Service:</h4>
            <pre>
              {JSON.stringify(collectionService.getDebugInfo(), null, 2)}
            </pre>
          </div>
        </details>
      )}
    </div>
  );
};

export default CollectionCreate;
