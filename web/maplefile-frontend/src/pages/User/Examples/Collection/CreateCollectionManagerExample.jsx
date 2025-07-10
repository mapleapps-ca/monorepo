// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/Collection/CreateCollectionManagerExample.jsx
// Fixed example component demonstrating how to use the collection creation services

import React, { useState, useEffect } from "react";
import { useCollections, useAuth } from "../../../../services/Services";

const CreateCollectionManagerExample = () => {
  // Get service managers
  const { createCollectionManager } = useCollections();
  const { authManager } = useAuth();

  // Create user object from authManager
  const user = {
    email: authManager?.getCurrentUserEmail?.() || null,
    isAuthenticated: authManager?.isAuthenticated?.() || false,
  };

  // Component state
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [collections, setCollections] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Form state
  const [collectionName, setCollectionName] = useState("");
  const [collectionType, setCollectionType] = useState("folder");
  const [password, setPassword] = useState("");
  const [searchTerm, setSearchTerm] = useState("");

  // UI state
  const [selectedCollection, setSelectedCollection] = useState(null);
  const [eventLog, setEventLog] = useState([]);

  // Computed values
  const isAuthenticated = user.isAuthenticated;
  const canCreateCollections = isAuthenticated && !isLoading;
  const totalCollections = collections.length;
  const collectionsByType = collections.reduce((acc, col) => {
    const type = col.collection_type || "unknown";
    acc[type] = (acc[type] || 0) + 1;
    return acc;
  }, {});

  // Load collections and manager status
  useEffect(() => {
    const loadData = async () => {
      try {
        if (createCollectionManager) {
          // Load created collections
          const createdCollections =
            createCollectionManager.getCreatedCollections();
          setCollections(createdCollections);

          // Get manager status
          const status = createCollectionManager.getManagerStatus();
          setManagerStatus(status);
        }
      } catch (err) {
        console.error("Failed to load collection data:", err);
      }
    };

    loadData();
  }, [createCollectionManager, isAuthenticated]);

  // Handle collection creation
  const handleCreateCollection = async () => {
    if (!collectionName.trim()) {
      setError("Collection name is required");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      setSuccess(null);

      const result = await createCollectionManager.createCollection(
        {
          name: collectionName.trim(),
          collection_type: collectionType,
        },
        password || null,
      );

      if (result.success) {
        setSuccess(`Collection "${collectionName}" created successfully!`);

        // Refresh collections list
        const updatedCollections =
          createCollectionManager.getCreatedCollections();
        setCollections(updatedCollections);

        // Clear form
        setCollectionName("");
        setPassword("");

        // Log the event
        addToEventLog("collection_created", {
          name: collectionName,
          type: collectionType,
          id: result.collectionId,
        });
      }
    } catch (err) {
      console.error("Collection creation failed:", err);
      setError(err.message || "Failed to create collection");
    } finally {
      setIsLoading(false);
    }
  };

  // Handle quick folder creation
  const handleCreateFolder = async () => {
    if (!collectionName.trim()) {
      setError("Collection name is required");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      setSuccess(null);

      const result = await createCollectionManager.createCollection(
        {
          name: collectionName.trim(),
          collection_type: "folder",
        },
        password || null,
      );

      if (result.success) {
        setSuccess(`Folder "${collectionName}" created successfully!`);

        // Refresh collections list
        const updatedCollections =
          createCollectionManager.getCreatedCollections();
        setCollections(updatedCollections);

        // Clear form
        setCollectionName("");
        setPassword("");

        // Log the event
        addToEventLog("folder_created", {
          name: collectionName,
          id: result.collectionId,
        });
      }
    } catch (err) {
      console.error("Folder creation failed:", err);
      setError(err.message || "Failed to create folder");
    } finally {
      setIsLoading(false);
    }
  };

  // Handle quick album creation
  const handleCreateAlbum = async () => {
    if (!collectionName.trim()) {
      setError("Collection name is required");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      setSuccess(null);

      const result = await createCollectionManager.createCollection(
        {
          name: collectionName.trim(),
          collection_type: "album",
        },
        password || null,
      );

      if (result.success) {
        setSuccess(`Album "${collectionName}" created successfully!`);

        // Refresh collections list
        const updatedCollections =
          createCollectionManager.getCreatedCollections();
        setCollections(updatedCollections);

        // Clear form
        setCollectionName("");
        setPassword("");

        // Log the event
        addToEventLog("album_created", {
          name: collectionName,
          id: result.collectionId,
        });
      }
    } catch (err) {
      console.error("Album creation failed:", err);
      setError(err.message || "Failed to create album");
    } finally {
      setIsLoading(false);
    }
  };

  // Handle collection decryption
  const handleDecryptCollection = async (collection) => {
    try {
      setIsLoading(true);
      const decrypted = await createCollectionManager.decryptCollection(
        collection,
        password || null,
      );
      setSelectedCollection(decrypted);
      addToEventLog("collection_decrypted", {
        id: collection.id,
        name: decrypted.name,
      });
    } catch (err) {
      console.error("Decryption failed:", err);
      setError(err.message || "Failed to decrypt collection");
    } finally {
      setIsLoading(false);
    }
  };

  // Handle collection removal
  const handleRemoveCollection = async (collectionId) => {
    if (!confirm("Are you sure you want to remove this collection?")) return;

    try {
      await createCollectionManager.removeCollection(collectionId);

      // Refresh collections list
      const updatedCollections =
        createCollectionManager.getCreatedCollections();
      setCollections(updatedCollections);

      if (selectedCollection && selectedCollection.id === collectionId) {
        setSelectedCollection(null);
      }

      addToEventLog("collection_removed", { id: collectionId });
      setSuccess("Collection removed successfully");
    } catch (err) {
      console.error("Failed to remove collection:", err);
      setError(err.message || "Failed to remove collection");
    }
  };

  // Handle clear all collections
  const handleClearAllCollections = async () => {
    if (
      !confirm(
        "Are you sure you want to clear ALL collections? This cannot be undone.",
      )
    )
      return;

    try {
      await createCollectionManager.clearAllCollections();
      setCollections([]);
      setSelectedCollection(null);
      addToEventLog("all_collections_cleared", {});
      setSuccess("All collections cleared successfully");
    } catch (err) {
      console.error("Failed to clear collections:", err);
      setError(err.message || "Failed to clear collections");
    }
  };

  // Get password from storage
  const handleGetStoredPassword = async () => {
    try {
      const storedPassword = await createCollectionManager.getUserPassword();
      if (storedPassword) {
        setPassword(storedPassword);
        addToEventLog("password_loaded", { source: "storage" });
      } else {
        setError("No password found in storage");
      }
    } catch (err) {
      setError(`Failed to get stored password: ${err.message}`);
    }
  };

  // Search collections
  const searchCollections = (searchTerm) => {
    return createCollectionManager.searchCollections(searchTerm);
  };

  // Add event to log
  const addToEventLog = (eventType, eventData) => {
    setEventLog((prev) => [
      ...prev,
      {
        timestamp: new Date().toISOString(),
        eventType,
        eventData,
      },
    ]);
  };

  // Clear event log
  const handleClearLog = () => {
    setEventLog([]);
  };

  // Clear messages
  const clearMessages = () => {
    setError(null);
    setSuccess(null);
  };

  // Search collections
  const filteredCollections = searchTerm
    ? searchCollections(searchTerm)
    : collections;

  // Auto-clear messages after 5 seconds
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        clearMessages();
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [success, error]);

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>📁 Create Collection Manager Example (with Hooks)</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the <strong>createCollectionManager</strong>{" "}
        service with E2EE encryption for collection creation.
      </p>

      {/* Manager Status */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#f8f9fa",
          borderRadius: "6px",
          border: "1px solid #dee2e6",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>📊 Manager Status:</h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: "10px",
          }}
        >
          <div>
            <strong>User:</strong> {user?.email || "Not logged in"}
          </div>
          <div>
            <strong>Authenticated:</strong>{" "}
            {isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can Create Collections:</strong>{" "}
            {canCreateCollections ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Total Collections:</strong> {totalCollections}
          </div>
          <div>
            <strong>By Type:</strong>{" "}
            {Object.entries(collectionsByType)
              .map(([type, count]) => `${type}: ${count}`)
              .join(", ") || "None"}
          </div>
        </div>
        {import.meta.env.DEV && (
          <div style={{ marginTop: "10px", fontSize: "12px", color: "#666" }}>
            <strong>Debug Info:</strong> AuthManager available:{" "}
            {authManager ? "✅" : "❌"}, CreateCollectionManager available:{" "}
            {createCollectionManager ? "✅" : "❌"}
          </div>
        )}
      </div>

      {/* Create Collection Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>➕ Create New Collection:</h4>
        <div style={{ display: "grid", gap: "15px" }}>
          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Collection Name *
            </label>
            <input
              type="text"
              value={collectionName}
              onChange={(e) => setCollectionName(e.target.value)}
              placeholder="Enter collection name..."
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            />
          </div>

          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Collection Type
            </label>
            <select
              value={collectionType}
              onChange={(e) => setCollectionType(e.target.value)}
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            >
              <option value={"folder"}>📁 Folder</option>
              <option value={"album"}>📷 Album</option>
            </select>
          </div>

          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Password (for encryption)
            </label>
            <div style={{ display: "flex", gap: "10px" }}>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter password or use stored password..."
                style={{
                  flex: 1,
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                }}
              />
              <button
                onClick={handleGetStoredPassword}
                style={{
                  padding: "8px 15px",
                  backgroundColor: "#6c757d",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                }}
              >
                Use Stored
              </button>
            </div>
            <small style={{ color: "#666" }}>
              Leave empty to use password from PasswordStorageService
            </small>
          </div>

          <div style={{ display: "flex", gap: "10px" }}>
            <button
              onClick={handleCreateCollection}
              disabled={isLoading || !collectionName.trim() || !isAuthenticated}
              style={{
                flex: 1,
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionName.trim() || !isAuthenticated
                    ? "#6c757d"
                    : "#28a745",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionName.trim() || !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {isLoading ? "🔄 Creating..." : "➕ Create Collection"}
            </button>

            <button
              onClick={handleCreateFolder}
              disabled={isLoading || !collectionName.trim() || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionName.trim() || !isAuthenticated
                    ? "#6c757d"
                    : "#007bff",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionName.trim() || !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              📁 Quick Folder
            </button>

            <button
              onClick={handleCreateAlbum}
              disabled={isLoading || !collectionName.trim() || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionName.trim() || !isAuthenticated
                    ? "#6c757d"
                    : "#17a2b8",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionName.trim() || !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              📷 Quick Album
            </button>
          </div>
        </div>
      </div>

      {/* Success/Error Messages */}
      {success && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#d4edda",
            borderRadius: "6px",
            color: "#155724",
            border: "1px solid #c3e6cb",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>✅ {success}</span>
          <button
            onClick={clearMessages}
            style={{
              background: "none",
              border: "none",
              color: "#155724",
              cursor: "pointer",
              fontSize: "16px",
            }}
          >
            ✕
          </button>
        </div>
      )}

      {error && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#f8d7da",
            borderRadius: "6px",
            color: "#721c24",
            border: "1px solid #f5c6cb",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>❌ {error}</span>
          <button
            onClick={clearMessages}
            style={{
              background: "none",
              border: "none",
              color: "#721c24",
              cursor: "pointer",
              fontSize: "16px",
            }}
          >
            ✕
          </button>
        </div>
      )}

      {/* Collections List */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#fff3cd",
          borderRadius: "6px",
          border: "1px solid #ffeaa7",
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "15px",
          }}
        >
          <h4 style={{ margin: 0 }}>
            📚 Created Collections ({totalCollections}):
          </h4>
          <div style={{ display: "flex", gap: "10px" }}>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search collections..."
              style={{
                padding: "5px 10px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                fontSize: "14px",
              }}
            />
            {collections.length > 0 && (
              <button
                onClick={handleClearAllCollections}
                style={{
                  padding: "5px 15px",
                  backgroundColor: "#dc3545",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                }}
              >
                🗑️ Clear All
              </button>
            )}
          </div>
        </div>

        {filteredCollections.length === 0 ? (
          <div
            style={{
              padding: "40px",
              textAlign: "center",
              backgroundColor: "#f8f9fa",
              borderRadius: "6px",
              border: "2px dashed #dee2e6",
            }}
          >
            <p style={{ fontSize: "18px", color: "#6c757d" }}>
              {collections.length === 0
                ? "No collections created yet."
                : "No collections match your search."}
            </p>
            <p style={{ color: "#6c757d" }}>
              {collections.length === 0
                ? "Create your first collection using the form above."
                : "Try a different search term."}
            </p>
          </div>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {filteredCollections.map((collection) => (
              <div
                key={collection.id}
                style={{
                  padding: "15px",
                  border: "1px solid #dee2e6",
                  borderRadius: "6px",
                  backgroundColor: "white",
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: "bold", marginBottom: "5px" }}>
                    {collection.collection_type === "folder" ? "📁" : "📷"}{" "}
                    {collection.name || "[Encrypted]"}
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <strong>ID:</strong> {collection.id}
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <strong>Type:</strong> {collection.collection_type} |
                    <strong> Created:</strong>{" "}
                    {new Date(collection.created_at).toLocaleString()}
                  </div>
                  {collection._hasCollectionKey && (
                    <div style={{ fontSize: "12px", color: "#28a745" }}>
                      🔑 Collection key available in memory
                    </div>
                  )}
                </div>
                <div style={{ display: "flex", gap: "10px" }}>
                  <button
                    onClick={() => handleDecryptCollection(collection)}
                    disabled={isLoading}
                    style={{
                      padding: "5px 10px",
                      backgroundColor: "#007bff",
                      color: "white",
                      border: "none",
                      borderRadius: "4px",
                      cursor: isLoading ? "not-allowed" : "pointer",
                      fontSize: "12px",
                    }}
                  >
                    🔓 Decrypt
                  </button>
                  <button
                    onClick={() => handleRemoveCollection(collection.id)}
                    style={{
                      padding: "5px 10px",
                      backgroundColor: "#dc3545",
                      color: "white",
                      border: "none",
                      borderRadius: "4px",
                      cursor: "pointer",
                      fontSize: "12px",
                    }}
                  >
                    🗑️ Remove
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Selected Collection Details */}
      {selectedCollection && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#d1ecf1",
            borderRadius: "6px",
            border: "1px solid #bee5eb",
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: "15px",
            }}
          >
            <h4 style={{ margin: 0 }}>🔍 Collection Details:</h4>
            <button
              onClick={() => setSelectedCollection(null)}
              style={{
                background: "none",
                border: "none",
                color: "#0c5460",
                cursor: "pointer",
                fontSize: "16px",
              }}
            >
              ✕
            </button>
          </div>
          <pre
            style={{
              backgroundColor: "#f8f9fa",
              padding: "10px",
              borderRadius: "4px",
              fontSize: "12px",
              overflow: "auto",
              maxHeight: "300px",
            }}
          >
            {JSON.stringify(selectedCollection, null, 2)}
          </pre>
        </div>
      )}

      {/* Event Log */}
      <div>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <h3>📋 Collection Event Log ({eventLog.length})</h3>
          <button
            onClick={handleClearLog}
            disabled={eventLog.length === 0}
            style={{
              padding: "5px 15px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: eventLog.length === 0 ? "not-allowed" : "pointer",
              fontSize: "14px",
            }}
          >
            Clear Log
          </button>
        </div>

        {eventLog.length === 0 ? (
          <div
            style={{
              padding: "40px",
              textAlign: "center",
              backgroundColor: "#f8f9fa",
              borderRadius: "6px",
              border: "2px dashed #dee2e6",
            }}
          >
            <p style={{ fontSize: "18px", color: "#6c757d" }}>
              No collection events logged yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Events will appear here when collections are created, decrypted,
              or other actions occur.
            </p>
          </div>
        ) : (
          <div
            style={{
              maxHeight: "300px",
              overflow: "auto",
              border: "1px solid #dee2e6",
              borderRadius: "6px",
              backgroundColor: "#f8f9fa",
            }}
          >
            {eventLog
              .slice()
              .reverse()
              .map((event, index) => (
                <div
                  key={`${event.timestamp}-${index}`}
                  style={{
                    padding: "10px",
                    borderBottom:
                      index < eventLog.length - 1
                        ? "1px solid #dee2e6"
                        : "none",
                    fontFamily: "monospace",
                    fontSize: "12px",
                  }}
                >
                  <div style={{ marginBottom: "5px" }}>
                    <strong style={{ color: "#007bff" }}>
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </strong>
                    {" - "}
                    <strong style={{ color: "#28a745" }}>
                      {event.eventType}
                    </strong>
                  </div>
                  <div style={{ color: "#666", marginLeft: "20px" }}>
                    {JSON.stringify(event.eventData, null, 2)}
                  </div>
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default CreateCollectionManagerExample;
