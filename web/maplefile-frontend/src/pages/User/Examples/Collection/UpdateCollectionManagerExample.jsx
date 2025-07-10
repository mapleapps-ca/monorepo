// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/Collection/UpdateCollectionManagerExample.jsx
// Example component demonstrating how to use the UpdateCollectionManager

import React, { useState, useEffect } from "react";
import { useCollections, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";

const UpdateCollectionManagerExample = () => {
  const { updateCollectionManager, getCollectionManager } = useCollections();
  const { authManager } = useAuth();

  // React state for the component
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [updatedCollections, setUpdatedCollections] = useState([]);
  const [updateHistory, setUpdateHistory] = useState([]);

  // Form state
  const [collectionId, setCollectionId] = useState("");
  const [newName, setNewName] = useState("");
  const [newType, setNewType] = useState("folder");
  const [version, setVersion] = useState(1);
  const [password, setPassword] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [updateOperation, setUpdateOperation] = useState("name"); // name, type, key_rotation, full

  // UI state
  const [selectedCollection, setSelectedCollection] = useState(null);
  const [eventLog, setEventLog] = useState([]);
  const [showHistory, setShowHistory] = useState(false);

  // Get user info
  const user = {
    email: authManager?.getCurrentUserEmail?.() || "Unknown",
    isAuthenticated: authManager?.isAuthenticated?.() || false,
  };

  // Load data on component mount
  useEffect(() => {
    loadUpdatedCollections();
    loadUpdateHistory();
  }, []);

  // Load updated collections from manager
  const loadUpdatedCollections = () => {
    try {
      const collections = updateCollectionManager.getUpdatedCollections();
      setUpdatedCollections(collections);
    } catch (err) {
      console.error("Failed to load updated collections:", err);
      setUpdatedCollections([]);
    }
  };

  // Load update history from manager
  const loadUpdateHistory = () => {
    try {
      const history = updateCollectionManager.getUpdateHistory();
      setUpdateHistory(history);
    } catch (err) {
      console.error("Failed to load update history:", err);
      setUpdateHistory([]);
    }
  };

  // Get recent updates (custom implementation)
  const getRecentUpdates = (hours = 24) => {
    const cutoffTime = Date.now() - hours * 60 * 60 * 1000;
    return updateHistory.filter((entry) => {
      return new Date(entry.timestamp).getTime() > cutoffTime;
    });
  };

  // Handle collection update
  const handleUpdateCollection = async () => {
    if (!collectionId.trim()) {
      setError("Collection ID is required");
      return;
    }

    if (version === null || version === undefined) {
      setError("Version is required for optimistic locking");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      let updateData = { version: parseInt(version) };

      switch (updateOperation) {
        case "name":
          if (!newName.trim()) {
            setError("New name is required for name update");
            return;
          }
          updateData.name = newName.trim();
          break;

        case "type":
          updateData.collection_type = newType;
          break;

        case "key_rotation":
          updateData.rotateCollectionKey = true;
          break;

        case "full":
          if (newName.trim()) {
            updateData.name = newName.trim();
          }
          updateData.collection_type = newType;
          updateData.rotateCollectionKey = true;
          break;

        default:
          setError("Please select an update operation");
          return;
      }

      const result = await updateCollectionManager.updateCollection(
        collectionId.trim(),
        updateData,
        password || null,
      );

      setSuccess(
        `Collection updated successfully! New version: ${result.newVersion}`,
      );

      // Clear form on success
      setNewName("");
      setPassword("");
      setVersion(result.newVersion);

      // Reload data
      loadUpdatedCollections();
      loadUpdateHistory();

      // Log the event
      addToEventLog("collection_updated", {
        id: collectionId,
        operation: updateOperation,
        updateData,
        newVersion: result.newVersion,
      });
    } catch (err) {
      console.error("Collection update failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Get password from storage
  const handleGetStoredPassword = async () => {
    try {
      const storedPassword = await updateCollectionManager.getUserPassword();
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

  // Handle quick update operations
  const handleQuickUpdateName = async () => {
    if (!collectionId.trim() || !newName.trim()) {
      setError("Collection ID and new name are required");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      const result = await updateCollectionManager.updateCollection(
        collectionId.trim(),
        {
          name: newName.trim(),
          version: parseInt(version),
        },
        password || null,
      );

      setSuccess("Collection name updated successfully!");
      setNewName("");
      setPassword("");
      setVersion(result.newVersion);

      loadUpdatedCollections();
      loadUpdateHistory();

      addToEventLog("collection_name_updated", {
        id: collectionId,
        newName: newName.trim(),
      });
    } catch (err) {
      console.error("Collection name update failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleQuickUpdateType = async () => {
    if (!collectionId.trim()) {
      setError("Collection ID is required");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      const result = await updateCollectionManager.updateCollection(
        collectionId.trim(),
        {
          collection_type: newType,
          version: parseInt(version),
        },
        password || null,
      );

      setSuccess("Collection type updated successfully!");
      setPassword("");
      setVersion(result.newVersion);

      loadUpdatedCollections();
      loadUpdateHistory();

      addToEventLog("collection_type_updated", {
        id: collectionId,
        newType,
      });
    } catch (err) {
      console.error("Collection type update failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleQuickRotateKey = async () => {
    if (!collectionId.trim()) {
      setError("Collection ID is required");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      const result = await updateCollectionManager.updateCollection(
        collectionId.trim(),
        {
          rotateCollectionKey: true,
          version: parseInt(version),
        },
        password || null,
      );

      setSuccess("Collection key rotated successfully!");
      setPassword("");
      setVersion(result.newVersion);

      loadUpdatedCollections();
      loadUpdateHistory();

      addToEventLog("collection_key_rotated", {
        id: collectionId,
      });
    } catch (err) {
      console.error("Collection key rotation failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle collection decryption
  const handleDecryptCollection = async (collection) => {
    try {
      setIsLoading(true);
      const decrypted = await updateCollectionManager.decryptCollection(
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
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle collection removal
  const handleRemoveUpdatedCollection = async (collectionId) => {
    if (
      !confirm(
        "Are you sure you want to remove this updated collection from local storage?",
      )
    )
      return;

    try {
      await updateCollectionManager.removeUpdatedCollection(collectionId);

      if (selectedCollection && selectedCollection.id === collectionId) {
        setSelectedCollection(null);
      }

      loadUpdatedCollections();
      addToEventLog("updated_collection_removed", { id: collectionId });
    } catch (err) {
      console.error("Failed to remove updated collection:", err);
      setError(err.message);
    }
  };

  // Handle clear all updated collections
  const handleClearAllUpdatedCollections = async () => {
    if (
      !confirm(
        "Are you sure you want to clear ALL updated collections? This cannot be undone.",
      )
    )
      return;

    try {
      await updateCollectionManager.clearAllUpdatedCollections();
      setSelectedCollection(null);
      loadUpdatedCollections();
      loadUpdateHistory();
      addToEventLog("all_updated_collections_cleared", {});
    } catch (err) {
      console.error("Failed to clear updated collections:", err);
      setError(err.message);
    }
  };

  // Get current collection version automatically
  const handleGetCurrentVersion = async () => {
    if (!collectionId.trim()) {
      setError("Collection ID is required");
      return;
    }

    try {
      setIsLoading(true);

      // Try to get from GetCollectionManager if available
      if (getCollectionManager) {
        const result = await getCollectionManager.getCollection(
          collectionId.trim(),
        );
        if (result.collection && result.collection.version !== undefined) {
          setVersion(result.collection.version);
          addToEventLog("current_version_retrieved", {
            id: collectionId,
            version: result.collection.version,
            source: "GetCollectionManager",
          });
          return;
        }
      }

      // Fallback: try to get from updated collections
      const updatedCollection =
        updateCollectionManager.getUpdatedCollectionById(collectionId.trim());
      if (updatedCollection && updatedCollection.version !== undefined) {
        setVersion(updatedCollection.version);
        addToEventLog("current_version_retrieved", {
          id: collectionId,
          version: updatedCollection.version,
          source: "LocalStorage",
        });
        return;
      }

      setError(
        "Could not retrieve current version. Collection may not exist or you may not have access.",
      );
    } catch (err) {
      console.error("Failed to get current version:", err);
      setError(`Failed to get current version: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
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
    ? updateCollectionManager.searchUpdatedCollections(searchTerm)
    : updatedCollections;

  // Get recent updates
  const recentUpdates = getRecentUpdates(24);

  // Manager status
  const managerStatus = updateCollectionManager.getManagerStatus();

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
      <h2>✏️ Update Collection Manager Example (with Unified Services)</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the <strong>UpdateCollectionManager</strong> with
        E2EE encryption for collection updates, including version management and
        optimistic locking. The backend requires the current
        encrypted_collection_key to be included in all update requests.
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
            {user.isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can Update Collections:</strong>{" "}
            {managerStatus.canUpdateCollections ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Total Updated:</strong> {updatedCollections.length}
          </div>
          <div>
            <strong>Recent Updates (24h):</strong> {recentUpdates.length}
          </div>
        </div>
      </div>

      {/* Update Collection Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>✏️ Update Collection:</h4>
        <div style={{ display: "grid", gap: "15px" }}>
          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Collection ID *
            </label>
            <input
              type="text"
              value={collectionId}
              onChange={(e) => setCollectionId(e.target.value)}
              placeholder="Enter collection UUID..."
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                fontFamily: "monospace",
                fontSize: "12px",
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
              Version * (for optimistic locking)
            </label>
            <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
              <input
                type="number"
                value={version}
                onChange={(e) => setVersion(parseInt(e.target.value) || 0)}
                min="0"
                style={{
                  width: "150px",
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                }}
              />
              <button
                onClick={handleGetCurrentVersion}
                disabled={!collectionId.trim() || isLoading}
                style={{
                  padding: "8px 15px",
                  backgroundColor:
                    !collectionId.trim() || isLoading ? "#6c757d" : "#17a2b8",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor:
                    !collectionId.trim() || isLoading
                      ? "not-allowed"
                      : "pointer",
                  fontSize: "14px",
                }}
              >
                Get Current Version
              </button>
            </div>
            <small style={{ color: "#666" }}>
              Current version of the collection (backend handles conflict
              detection). The system will automatically include the current
              encrypted_collection_key.
            </small>
          </div>

          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Update Operation
            </label>
            <select
              value={updateOperation}
              onChange={(e) => setUpdateOperation(e.target.value)}
              style={{
                width: "300px",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            >
              <option value="name">Update Name Only</option>
              <option value="type">Update Type Only</option>
              <option value="key_rotation">Rotate Collection Key Only</option>
              <option value="full">
                Full Update (Name + Type + Key Rotation)
              </option>
            </select>
          </div>

          {(updateOperation === "name" || updateOperation === "full") && (
            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                New Collection Name
              </label>
              <input
                type="text"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="Enter new collection name..."
                style={{
                  width: "100%",
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                }}
              />
            </div>
          )}

          {(updateOperation === "type" || updateOperation === "full") && (
            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                New Collection Type
              </label>
              <select
                value={newType}
                onChange={(e) => setNewType(e.target.value)}
                style={{
                  width: "200px",
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                }}
              >
                <option value={"folder"}>📁 Folder</option>
                <option value={"album"}>📷 Album</option>
              </select>
            </div>
          )}

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

          <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
            <button
              onClick={handleUpdateCollection}
              disabled={
                isLoading || !collectionId.trim() || !user.isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionId.trim() || !user.isAuthenticated
                    ? "#6c757d"
                    : "#28a745",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionId.trim() || !user.isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {isLoading ? "🔄 Updating..." : "✏️ Update Collection"}
            </button>

            <button
              onClick={handleQuickUpdateName}
              disabled={
                isLoading ||
                !collectionId.trim() ||
                !newName.trim() ||
                !user.isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading ||
                  !collectionId.trim() ||
                  !newName.trim() ||
                  !user.isAuthenticated
                    ? "#6c757d"
                    : "#007bff",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading ||
                  !collectionId.trim() ||
                  !newName.trim() ||
                  !user.isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              📝 Quick Name Update
            </button>

            <button
              onClick={handleQuickUpdateType}
              disabled={
                isLoading || !collectionId.trim() || !user.isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionId.trim() || !user.isAuthenticated
                    ? "#6c757d"
                    : "#17a2b8",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionId.trim() || !user.isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              🔄 Quick Type Update
            </button>

            <button
              onClick={handleQuickRotateKey}
              disabled={
                isLoading || !collectionId.trim() || !user.isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionId.trim() || !user.isAuthenticated
                    ? "#6c757d"
                    : "#ffc107",
                color: "black",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionId.trim() || !user.isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              🔐 Rotate Key
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

      {/* Updated Collections List */}
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
            📚 Updated Collections ({updatedCollections.length}):
          </h4>
          <div style={{ display: "flex", gap: "10px" }}>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search updated collections..."
              style={{
                padding: "5px 10px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                fontSize: "14px",
              }}
            />
            <button
              onClick={() => setShowHistory(!showHistory)}
              style={{
                padding: "5px 15px",
                backgroundColor: "#007bff",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              {showHistory ? "Hide" : "Show"} History
            </button>
            {updatedCollections.length > 0 && (
              <button
                onClick={handleClearAllUpdatedCollections}
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
              {updatedCollections.length === 0
                ? "No collections updated yet."
                : "No collections match your search."}
            </p>
            <p style={{ color: "#6c757d" }}>
              {updatedCollections.length === 0
                ? "Update your first collection using the form above."
                : "Try a different search term."}
            </p>
          </div>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {filteredCollections.map((collection) => {
              const latestUpdate = updateHistory
                .filter((entry) => entry.collectionId === collection.id)
                .sort(
                  (a, b) => new Date(b.timestamp) - new Date(a.timestamp),
                )[0];

              return (
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
                      <span
                        style={{
                          fontSize: "12px",
                          color: "#666",
                          marginLeft: "10px",
                        }}
                      >
                        v{collection.version}
                      </span>
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      <strong>ID:</strong> {collection.id}
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      <strong>Type:</strong> {collection.collection_type} |
                      <strong> Updated:</strong>{" "}
                      {new Date(
                        collection.locally_updated_at || collection.updated_at,
                      ).toLocaleString()}
                    </div>
                    {latestUpdate && (
                      <div style={{ fontSize: "12px", color: "#28a745" }}>
                        🔄 Last action: {latestUpdate.action} at{" "}
                        {new Date(latestUpdate.timestamp).toLocaleTimeString()}
                      </div>
                    )}
                    {collection.encrypted_collection_key && (
                      <div style={{ fontSize: "12px", color: "#17a2b8" }}>
                        🔐 Has encrypted collection key (v
                        {collection.encrypted_collection_key.key_version || 1})
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
                      onClick={() =>
                        handleRemoveUpdatedCollection(collection.id)
                      }
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
              );
            })}
          </div>
        )}
      </div>

      {/* Update History */}
      {showHistory && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#e2e3e5",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>
            📜 Update History ({updateHistory.length}):
          </h4>
          {updateHistory.length === 0 ? (
            <p style={{ color: "#6c757d" }}>No update history available.</p>
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
              {updateHistory
                .slice()
                .reverse()
                .map((entry, index) => (
                  <div
                    key={`${entry.timestamp}-${index}`}
                    style={{
                      padding: "10px",
                      borderBottom:
                        index < updateHistory.length - 1
                          ? "1px solid #dee2e6"
                          : "none",
                      fontFamily: "monospace",
                      fontSize: "12px",
                    }}
                  >
                    <div style={{ marginBottom: "5px" }}>
                      <strong style={{ color: "#007bff" }}>
                        {new Date(entry.timestamp).toLocaleString()}
                      </strong>
                      {" - "}
                      <strong style={{ color: "#28a745" }}>
                        {entry.action}
                      </strong>
                      {" - "}
                      <span style={{ color: "#666" }}>
                        {entry.collectionId}
                      </span>
                    </div>
                    {entry.version && (
                      <div style={{ color: "#666", marginLeft: "20px" }}>
                        Version: {entry.version}
                      </div>
                    )}
                  </div>
                ))}
            </div>
          )}
        </div>
      )}

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
          <h3>📋 Collection Update Event Log ({eventLog.length})</h3>
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
              No collection update events logged yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Events will appear here when collections are updated, decrypted,
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

      {/* Quick Test Section */}
      <div
        style={{
          padding: "15px",
          backgroundColor: "#e9ecef",
          borderRadius: "8px",
          marginTop: "20px",
          border: "1px solid #dee2e6",
        }}
      >
        <h5 style={{ margin: "0 0 10px 0" }}>🚀 Quick Test</h5>
        <p style={{ margin: "0 0 10px 0", fontSize: "14px", color: "#666" }}>
          First create a collection in the Create Collection Manager, then use
          its ID here. Use "Get Current Version" button to automatically fetch
          the latest version.
          <strong>Note:</strong> The backend requires the current
          encrypted_collection_key for all updates.
        </p>
        <div
          style={{
            display: "flex",
            gap: "10px",
            alignItems: "center",
            flexWrap: "wrap",
          }}
        >
          <button
            onClick={() => {
              setCollectionId("7f558adb-57b6-11f0-8b98-c60a0c48537c");
              setVersion(1);
            }}
            style={{
              padding: "5px 10px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "3px",
              cursor: "pointer",
              fontSize: "12px",
              fontFamily: "monospace",
            }}
          >
            Use Sample ID
          </button>
          <span style={{ fontSize: "12px", color: "#666" }}>
            Click "Get Current Version" after entering a collection ID to fetch
            the latest version automatically!
          </span>
        </div>
      </div>
    </div>
  );
};

// Wrap with password protection to ensure user is authenticated
export default withPasswordProtection(UpdateCollectionManagerExample);
