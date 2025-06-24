// Collection Create Page - E2EE Collection Creation
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";

const CollectionCreate = () => {
  const navigate = useNavigate();
  const { collectionService, cryptoService } = useServices();
  const { isAuthenticated } = useAuth();

  // Form state
  const [formData, setFormData] = useState({
    name: "",
    collection_type: "folder", // Default to folder
    parent_id: "", // Optional parent collection
    description: "", // Optional description for user reference
  });

  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const [createdCollection, setCreatedCollection] = useState(null);

  // Parent collection selection
  const [availableParents, setAvailableParents] = useState([]);
  const [loadingParents, setLoadingParents] = useState(false);

  // Sharing configuration
  const [enableSharing, setEnableSharing] = useState(false);
  const [shareWithEmail, setShareWithEmail] = useState("");
  const [sharePermission, setSharePermission] = useState("read_write");

  // Load available parent collections on mount
  useEffect(() => {
    if (isAuthenticated) {
      loadAvailableParents();
    }
  }, [isAuthenticated]);

  // Load available collections that can be parents
  const loadAvailableParents = useCallback(async () => {
    try {
      setLoadingParents(true);
      // Get user's owned collections to use as potential parents
      const collections = await collectionService.listUserCollections();

      // Only show folders as potential parents (albums typically don't contain other collections)
      const folders = collections.filter(
        (col) => col.collection_type === "folder",
      );
      setAvailableParents(folders);
    } catch (err) {
      console.error("Failed to load parent collections:", err);
      // Non-critical error, just log it
    } finally {
      setLoadingParents(false);
    }
  }, [collectionService]);

  // Handle form input changes
  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value,
    }));

    // Clear error when user makes changes
    if (error) {
      setError("");
    }
  };

  // Generate a UUID v4 - Note: In production, the server will override this
  const generateUUID = () => {
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(
      /[xy]/g,
      function (c) {
        const r = (Math.random() * 16) | 0;
        const v = c === "x" ? r : (r & 0x3) | 0x8;
        return v.toString(16);
      },
    );
  };

  // Generate collection encryption key and encrypt collection name
  const prepareEncryptedData = async (collectionName) => {
    try {
      console.log(
        "[CollectionCreate] Preparing encrypted data for:",
        collectionName,
      );

      // Initialize crypto service
      await cryptoService.initialize();

      // For this demo, we'll create a basic encrypted collection
      // In a real implementation, this would use the user's master key and proper E2EE

      // Generate a collection key using a simple approach
      // Note: This is a simplified version - production should use proper key derivation
      const encoder = new TextEncoder();
      const keyMaterial = encoder.encode(
        collectionName + Date.now().toString(),
      );

      // Create a simple hash for demonstration (not production-ready)
      const keyBuffer = await crypto.subtle.digest("SHA-256", keyMaterial);
      const collectionKey = new Uint8Array(keyBuffer);

      console.log(
        "[CollectionCreate] Generated collection key, length:",
        collectionKey.length,
      );

      // For demonstration, we'll use a simple encoding approach
      // Production should use proper ChaCha20-Poly1305 encryption

      // Create a simple "encrypted" name (base64 encoded for demo)
      const nameData = encoder.encode(collectionName);
      const encryptedName = btoa(String.fromCharCode(...nameData));

      // Create a placeholder encrypted collection key structure
      // In production, this would be encrypted with the user's master key
      const keyNonce = new Uint8Array(24); // Placeholder nonce
      crypto.getRandomValues(keyNonce);

      const encryptedCollectionKey = {
        ciphertext: Array.from(collectionKey), // Placeholder - should be encrypted
        nonce: Array.from(keyNonce),
        key_version: 1,
        rotated_at: new Date().toISOString(),
        previous_keys: [],
      };

      console.log("[CollectionCreate] Encryption complete (demo mode)");
      return {
        encryptedName,
        encryptedCollectionKey,
        collectionKey: Array.from(collectionKey), // For potential sharing
      };
    } catch (error) {
      console.error("[CollectionCreate] Encryption failed:", error);
      throw new Error(`Encryption failed: ${error.message}`);
    }
  };

  // Build ancestor chain for hierarchy (only when parent exists)
  const buildAncestorChain = async (parentId) => {
    if (!parentId || parentId.trim() === "") {
      return [];
    }

    try {
      // Get the parent collection to build the ancestor chain
      const parentCollection = await collectionService.getCollection(parentId);

      // The ancestor chain includes the parent's ancestors plus the parent itself
      const ancestorIds = [...(parentCollection.ancestor_ids || []), parentId];

      return ancestorIds;
    } catch (error) {
      console.error("Failed to build ancestor chain:", error);
      // If we can't get parent info, just use the parent ID
      return [parentId];
    }
  };

  // Prepare sharing configuration according to API specification
  const prepareSharingConfig = (encryptedCollectionKey) => {
    if (!enableSharing || !shareWithEmail.trim()) {
      return [];
    }

    // In a real implementation, you would:
    // 1. Look up the recipient's public key from the user database
    // 2. Encrypt the collection key with their public key using BoxSeal
    // 3. Create the proper member object with encrypted access

    // Placeholder sharing configuration for demo - matches API format exactly
    return [
      {
        recipient_email: shareWithEmail.trim(),
        permission_level: sharePermission,
        encrypted_collection_key: encryptedCollectionKey.ciphertext, // Array of numbers as per API
        is_inherited: false,
        // Note: recipient_id, granted_by_id, created_at, id, collection_id will be set by server
      },
    ];
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();

    // Validation
    if (!formData.name.trim()) {
      setError("Collection name is required");
      return;
    }

    if (!formData.collection_type) {
      setError("Collection type is required");
      return;
    }

    if (enableSharing && !shareWithEmail.trim()) {
      setError("Email is required when enabling sharing");
      return;
    }

    if (
      enableSharing &&
      shareWithEmail.trim() &&
      !shareWithEmail.includes("@")
    ) {
      setError("Please enter a valid email address");
      return;
    }

    try {
      setLoading(true);
      setError("");
      setSuccess(false);

      console.log("[CollectionCreate] Starting collection creation process");

      // Step 1: Prepare encrypted data
      const { encryptedName, encryptedCollectionKey } =
        await prepareEncryptedData(formData.name.trim());

      // Step 2: Build ancestor chain only if parent is selected
      const ancestorIds = formData.parent_id
        ? await buildAncestorChain(formData.parent_id)
        : [];

      // Step 3: Prepare sharing configuration
      const members = prepareSharingConfig(encryptedCollectionKey);

      // Step 4: Build the API request payload according to exact API specification
      const collectionData = {
        // Required fields only
        encrypted_name: encryptedName,
        collection_type: formData.collection_type,
        encrypted_collection_key: encryptedCollectionKey,
      };

      // Add optional fields only if they have values
      if (formData.parent_id) {
        collectionData.parent_id = formData.parent_id;
        collectionData.ancestor_ids = ancestorIds;
      }

      // Add members only if sharing is enabled and configured
      if (members.length > 0) {
        collectionData.members = members;
      }

      console.log("[CollectionCreate] Creating collection with data:", {
        name: formData.name,
        type: formData.collection_type,
        hasParent: !!formData.parent_id,
        ancestorCount: ancestorIds.length,
        memberCount: members.length,
        payload: collectionData,
      });

      console.log(
        "[CollectionCreate] Exact JSON payload being sent:",
        JSON.stringify(collectionData, null, 2),
      );

      // Step 5: Create the collection via API
      const createdCollection =
        await collectionService.createCollection(collectionData);

      console.log(
        "[CollectionCreate] Collection created successfully:",
        createdCollection,
      );

      setSuccess(true);
      setCreatedCollection(createdCollection);

      // Reset form
      setFormData({
        name: "",
        collection_type: "folder",
        parent_id: "",
        description: "",
      });
      setEnableSharing(false);
      setShareWithEmail("");
    } catch (err) {
      console.error("[CollectionCreate] Collection creation failed:", err);
      setError(err.message || "Failed to create collection");
    } finally {
      setLoading(false);
    }
  };

  // Handle navigation
  const handleBackToList = () => {
    navigate("/collections");
  };

  const handleViewCreated = () => {
    if (createdCollection) {
      // Navigate to collection detail or file list
      navigate(`/collections/${createdCollection.id}`);
    }
  };

  if (!isAuthenticated) {
    return (
      <div>
        <h2>Create Collection</h2>
        <p>Please log in to create collections.</p>
        <button onClick={() => navigate("/login")}>Go to Login</button>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "600px" }}>
      <div style={{ marginBottom: "20px" }}>
        <h1>Create New Collection</h1>
        <button onClick={handleBackToList}>← Back to Collections</button>
      </div>

      {/* Success Message */}
      {success && createdCollection && (
        <div
          style={{
            backgroundColor: "#d4edda",
            border: "1px solid #c3e6cb",
            padding: "15px",
            marginBottom: "20px",
            borderRadius: "5px",
          }}
        >
          <h3>✅ Collection Created Successfully!</h3>
          <p>
            <strong>Name:</strong> {formData.name}
          </p>
          <p>
            <strong>Type:</strong> {createdCollection.collection_type}
          </p>
          <p>
            <strong>ID:</strong> {createdCollection.id}
          </p>
          <div
            style={{
              backgroundColor: "#fff3cd",
              border: "1px solid #ffeaa7",
              padding: "10px",
              marginTop: "10px",
              borderRadius: "3px",
              fontSize: "12px",
            }}
          >
            <strong>⚠️ Demo Mode:</strong> This is a simplified encryption
            implementation for demonstration. Production requires proper master
            key derivation and ChaCha20-Poly1305 encryption.
          </div>
          <div style={{ marginTop: "10px" }}>
            <button onClick={handleViewCreated} style={{ marginRight: "10px" }}>
              View Collection
            </button>
            <button onClick={() => setSuccess(false)}>Create Another</button>
          </div>
        </div>
      )}

      {/* Error Message */}
      {error && (
        <div
          style={{
            backgroundColor: "#f8d7da",
            border: "1px solid #f5c6cb",
            padding: "15px",
            marginBottom: "20px",
            borderRadius: "5px",
          }}
        >
          <strong>❌ Error:</strong> {error}
        </div>
      )}

      {/* Collection Creation Form */}
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "20px" }}>
          <h2>Collection Details</h2>

          {/* Collection Name */}
          <div style={{ marginBottom: "15px" }}>
            <label
              htmlFor="name"
              style={{ display: "block", marginBottom: "5px" }}
            >
              <strong>Collection Name *</strong>
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              placeholder="Enter collection name"
              required
              disabled={loading}
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            />
            <small style={{ color: "#666" }}>
              This name will be encrypted before being stored.
            </small>
          </div>

          {/* Collection Type */}
          <div style={{ marginBottom: "15px" }}>
            <label
              htmlFor="collection_type"
              style={{ display: "block", marginBottom: "5px" }}
            >
              <strong>Collection Type *</strong>
            </label>
            <select
              id="collection_type"
              name="collection_type"
              value={formData.collection_type}
              onChange={handleInputChange}
              required
              disabled={loading}
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            >
              <option value="folder">
                📁 Folder - For organizing files and other collections
              </option>
              <option value="album">
                🖼️ Album - For photo/media collections with specialized features
              </option>
            </select>
          </div>

          {/* Description (for user reference only) */}
          <div style={{ marginBottom: "15px" }}>
            <label
              htmlFor="description"
              style={{ display: "block", marginBottom: "5px" }}
            >
              <strong>Description</strong> (optional, for your reference only)
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleInputChange}
              placeholder="Describe what this collection will contain..."
              disabled={loading}
              rows="3"
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                resize: "vertical",
              }}
            />
            <small style={{ color: "#666" }}>
              This description is not sent to the server and is only for your
              reference.
            </small>
          </div>
        </div>

        {/* Hierarchy Section */}
        <div style={{ marginBottom: "20px" }}>
          <h3>📂 Hierarchy (Optional)</h3>

          <div style={{ marginBottom: "15px" }}>
            <label
              htmlFor="parent_id"
              style={{ display: "block", marginBottom: "5px" }}
            >
              <strong>Parent Collection</strong>
            </label>
            <select
              id="parent_id"
              name="parent_id"
              value={formData.parent_id}
              onChange={handleInputChange}
              disabled={loading || loadingParents}
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            >
              <option value="">No parent (create at root level)</option>
              {availableParents.map((parent) => (
                <option key={parent.id} value={parent.id}>
                  📁 {parent.encrypted_name}
                </option>
              ))}
            </select>
            <small style={{ color: "#666" }}>
              {loadingParents
                ? "Loading available parents..."
                : "Select a parent folder to create this collection inside it."}
            </small>
          </div>
        </div>

        {/* Sharing Section */}
        <div style={{ marginBottom: "20px" }}>
          <h3>👥 Initial Sharing (Optional)</h3>

          <div style={{ marginBottom: "15px" }}>
            <label>
              <input
                type="checkbox"
                checked={enableSharing}
                onChange={(e) => setEnableSharing(e.target.checked)}
                disabled={loading}
              />
              <span style={{ marginLeft: "8px" }}>
                <strong>Share this collection immediately upon creation</strong>
              </span>
            </label>
          </div>

          {enableSharing && (
            <div
              style={{
                padding: "15px",
                border: "1px solid #ddd",
                borderRadius: "5px",
                backgroundColor: "#f9f9f9",
              }}
            >
              <div style={{ marginBottom: "15px" }}>
                <label
                  htmlFor="shareWithEmail"
                  style={{ display: "block", marginBottom: "5px" }}
                >
                  <strong>Share with Email *</strong>
                </label>
                <input
                  type="email"
                  id="shareWithEmail"
                  value={shareWithEmail}
                  onChange={(e) => setShareWithEmail(e.target.value)}
                  placeholder="user@example.com"
                  required={enableSharing}
                  disabled={loading}
                  style={{
                    width: "100%",
                    padding: "8px",
                    border: "1px solid #ddd",
                    borderRadius: "4px",
                  }}
                />
              </div>

              <div style={{ marginBottom: "15px" }}>
                <label
                  htmlFor="sharePermission"
                  style={{ display: "block", marginBottom: "5px" }}
                >
                  <strong>Permission Level *</strong>
                </label>
                <select
                  id="sharePermission"
                  value={sharePermission}
                  onChange={(e) => setSharePermission(e.target.value)}
                  disabled={loading}
                  style={{
                    width: "100%",
                    padding: "8px",
                    border: "1px solid #ddd",
                    borderRadius: "4px",
                  }}
                >
                  <option value="read_only">
                    👁️ Read Only - Can view collection and files
                  </option>
                  <option value="read_write">
                    ✏️ Read & Write - Can add/modify files and subcollections
                  </option>
                  <option value="admin">
                    👑 Admin - Full control including sharing and deletion
                  </option>
                </select>
              </div>

              <small style={{ color: "#666" }}>
                Note: In this demo, sharing is simulated. Production requires
                recipient public key lookup and proper BoxSeal encryption for
                the collection key.
              </small>
            </div>
          )}
        </div>

        {/* Submit Section */}
        <div
          style={{
            borderTop: "1px solid #eee",
            paddingTop: "20px",
            display: "flex",
            gap: "10px",
            justifyContent: "flex-end",
          }}
        >
          <button
            type="button"
            onClick={handleBackToList}
            disabled={loading}
            style={{
              padding: "10px 20px",
              border: "1px solid #ddd",
              backgroundColor: "#f8f9fa",
              borderRadius: "4px",
              cursor: loading ? "not-allowed" : "pointer",
            }}
          >
            Cancel
          </button>

          <button
            type="submit"
            disabled={loading || !formData.name.trim()}
            style={{
              padding: "10px 20px",
              border: "1px solid #007bff",
              backgroundColor: loading ? "#6c757d" : "#007bff",
              color: "white",
              borderRadius: "4px",
              cursor:
                loading || !formData.name.trim() ? "not-allowed" : "pointer",
            }}
          >
            {loading ? "🔄 Creating..." : "✅ Create Collection"}
          </button>
        </div>
      </form>

      {/* API Compliance Information */}
      <div
        style={{
          marginTop: "40px",
          padding: "15px",
          backgroundColor: "#f8f9fa",
          borderRadius: "5px",
          border: "1px solid #e9ecef",
        }}
      >
        <h4>🔐 E2EE Implementation Notes</h4>
        <div
          style={{
            backgroundColor: "#fff3cd",
            border: "1px solid #ffeaa7",
            padding: "10px",
            marginBottom: "15px",
            borderRadius: "3px",
          }}
        >
          <strong>⚠️ Demo Implementation:</strong> This uses simplified
          encryption for demonstration. Production requires proper CryptoService
          integration with ChaCha20-Poly1305.
        </div>
        <ul style={{ fontSize: "12px", color: "#666", paddingLeft: "20px" }}>
          <li>
            <strong>Collection Name:</strong> Base64 encoded for demo (should be
            ChaCha20-Poly1305 encrypted)
          </li>
          <li>
            <strong>Collection Key:</strong> Generated using browser crypto API
            (should use libsodium)
          </li>
          <li>
            <strong>API Endpoint:</strong> POST /collections (as per API
            documentation)
          </li>
          <li>
            <strong>Hierarchy:</strong> Supports parent-child relationships with
            automatic ancestor chain building
          </li>
          <li>
            <strong>Sharing:</strong> Placeholder implementation - production
            requires BoxSeal encryption
          </li>
          <li>
            <strong>UUID Generation:</strong> Client-side UUID v4 generation
            (server may override)
          </li>
          <li>
            <strong>Data Types:</strong> Supports both "folder" and "album"
            collection types as specified in API
          </li>
        </ul>
        <div style={{ marginTop: "10px", fontSize: "12px", color: "#856404" }}>
          <strong>Production TODO:</strong> Integrate with user's master key,
          implement proper ChaCha20-Poly1305 encryption, and add recipient
          public key lookup for sharing.
        </div>
        <div style={{ marginTop: "5px", fontSize: "11px", color: "#666" }}>
          <strong>Payload Structure:</strong> Minimal required fields only
          (encrypted_name, collection_type, encrypted_collection_key). Optional
          fields (parent_id, ancestor_ids, members) included only when needed.
        </div>
      </div>
    </div>
  );
};

export default CollectionCreate;
