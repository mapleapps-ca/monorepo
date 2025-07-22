// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/Collection/DeleteCollectionManagerExample.jsx
// Example component demonstrating how to use the DeleteCollectionManager

import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useCollections, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";

const DeleteCollectionManagerExample = () => {
  const navigate = useNavigate();
  const { deleteCollectionManager } = useCollections();
  const { authManager } = useAuth();

  // Component state
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [managerStatus, setManagerStatus] = useState(null);

  // Form state
  const [collectionId, setCollectionId] = useState("");
  const [collectionIds, setCollectionIds] = useState("");
  const [password, setPassword] = useState("");
  const [searchTerm, setSearchTerm] = useState("");

  // Data state
  const [deletedCollections, setDeletedCollections] = useState([]);
  const [deletionHistory, setDeletionHistory] = useState([]);

  // UI state
  const [selectedCollection, setSelectedCollection] = useState(null);
  const [eventLog, setEventLog] = useState([]);
  const [showHistory, setShowHistory] = useState(false);
  const [availableCollections, setAvailableCollections] = useState([]);

  // Load data on mount and setup listeners
  useEffect(() => {
    loadData();
    loadAvailableCollections();
    loadManagerStatus();

    // Setup deletion event listener
    const handleDeletionEvent = (eventType, eventData) => {
      addToEventLog(eventType, eventData);
      loadData(); // Refresh data after any deletion event
    };

    deleteCollectionManager.addCollectionDeletionListener(handleDeletionEvent);

    return () => {
      deleteCollectionManager.removeCollectionDeletionListener(
        handleDeletionEvent,
      );
    };
  }, [deleteCollectionManager]);

  // Load deleted collections and history
  const loadData = () => {
    try {
      const collections = deleteCollectionManager.getDeletedCollections();
      const history = deleteCollectionManager.getDeletionHistory();

      setDeletedCollections(collections);
      setDeletionHistory(history);
    } catch (err) {
      console.error("Failed to load deletion data:", err);
    }
  };

  // Load manager status
  const loadManagerStatus = () => {
    try {
      const status = deleteCollectionManager.getManagerStatus();
      setManagerStatus(status);
    } catch (err) {
      console.error("Failed to load manager status:", err);
    }
  };

  // Load available collections for deletion (mock data for testing)
  const loadAvailableCollections = async () => {
    try {
      // In a real app, you'd get collections from ListCollectionManager or GetCollectionManager
      // For this example, we'll simulate some collections
      const mockCollections = [
        {
          id: "550e8400-e29b-41d4-a716-446655440001",
          name: "Sample Folder 1",
          collection_type: "folder",
          created_at: new Date(Date.now() - 86400000).toISOString(),
        },
        {
          id: "550e8400-e29b-41d4-a716-446655440002",
          name: "Sample Album 1",
          collection_type: "album",
          created_at: new Date(Date.now() - 172800000).toISOString(),
        },
      ];
      setAvailableCollections(mockCollections);
    } catch (error) {
      console.error("Failed to load available collections:", error);
    }
  };

  // Get recent deletions (helper function)
  const getRecentDeletions = (hours = 24) => {
    const cutoffTime = Date.now() - hours * 60 * 60 * 1000;
    return deletionHistory.filter((entry) => {
      const entryTime = new Date(entry.timestamp).getTime();
      return entryTime > cutoffTime;
    });
  };

  // Get latest deletion for a collection
  const getLatestDeletionForCollection = (collectionId) => {
    const collectionHistory = deletionHistory.filter(
      (entry) => entry.collectionId === collectionId,
    );
    return collectionHistory.length > 0
      ? collectionHistory[collectionHistory.length - 1]
      : null;
  };

  // Clear success/error messages
  const clearMessages = () => {
    setError(null);
    setSuccess(null);
  };

  // Handle collection deletion
  const handleDeleteCollection = async () => {
    if (!collectionId.trim()) {
      setError("Collection ID is required");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      setSuccess(null);

      const result = await deleteCollectionManager.deleteCollection(
        collectionId.trim(),
        password || null,
      );

      setSuccess(
        `Collection deleted successfully: ${result.collection.name || result.collection.id}`,
      );

      // Clear form
      setCollectionId("");
      setPassword("");

      // Refresh data
      loadData();
      loadManagerStatus();

      // Log the event
      addToEventLog("collection_deleted", {
        id: result.collection.id,
        name: result.collection.name,
      });
    } catch (err) {
      console.error("Collection deletion failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle batch collection deletion
  const handleDeleteCollections = async () => {
    const ids = collectionIds
      .split(",")
      .map((id) => id.trim())
      .filter((id) => id);

    if (ids.length === 0) {
      setError("Enter one or more collection IDs (comma-separated)");
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      setSuccess(null);

      const result = await deleteCollectionManager.deleteCollections(
        ids,
        password || null,
      );

      setSuccess(
        `Batch deletion completed: ${result.successCount} successful, ${result.errorCount} failed`,
      );

      // Clear form
      setCollectionIds("");
      setPassword("");

      // Refresh data
      loadData();
      loadManagerStatus();

      // Log the event
      addToEventLog("multiple_collections_deleted", {
        ids,
        successCount: result.successCount,
        errorCount: result.errorCount,
      });
    } catch (err) {
      console.error("Batch collection deletion failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle collection restoration
  const handleRestoreCollection = async (collectionToRestore) => {
    if (
      !confirm(
        `Are you sure you want to restore collection "${collectionToRestore.name || collectionToRestore.id}"?`,
      )
    )
      return;

    try {
      setIsLoading(true);
      setError(null);

      const result = await deleteCollectionManager.restoreCollection(
        collectionToRestore.id,
      );

      setSuccess(
        `Collection restored successfully: ${collectionToRestore.name || collectionToRestore.id}`,
      );

      // Refresh data
      loadData();
      loadManagerStatus();

      addToEventLog("collection_restored", {
        id: collectionToRestore.id,
        name: collectionToRestore.name,
      });
    } catch (err) {
      console.error("Failed to restore collection:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle collection decryption
  const handleDecryptCollection = async (collection) => {
    try {
      setIsLoading(true);
      setError(null);

      const decrypted = await deleteCollectionManager.decryptCollection(
        collection,
        password || null,
      );

      setSelectedCollection(decrypted);
      setSuccess(`Collection decrypted successfully: ${decrypted.name}`);

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

  // Handle permanent removal
  const handlePermanentlyRemove = async (collectionToRemove) => {
    if (
      !confirm(
        `Are you sure you want to PERMANENTLY remove collection "${collectionToRemove.name || collectionToRemove.id}" from local storage? This cannot be undone.`,
      )
    )
      return;

    try {
      setIsLoading(true);

      await deleteCollectionManager.permanentlyRemoveCollection(
        collectionToRemove.id,
      );

      if (
        selectedCollection &&
        selectedCollection.id === collectionToRemove.id
      ) {
        setSelectedCollection(null);
      }

      setSuccess("Collection permanently removed from local storage");

      // Refresh data
      loadData();

      addToEventLog("collection_permanently_removed", {
        id: collectionToRemove.id,
      });
    } catch (err) {
      console.error("Failed to permanently remove collection:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle clear all deleted collections
  const handleClearAllDeletedCollections = async () => {
    if (
      !confirm(
        "Are you sure you want to clear ALL deleted collections? This cannot be undone.",
      )
    )
      return;

    try {
      setIsLoading(true);

      await deleteCollectionManager.clearAllDeletedCollections();

      setSelectedCollection(null);
      setSuccess("All deleted collections cleared");

      // Refresh data
      loadData();

      addToEventLog("all_deleted_collections_cleared", {});
    } catch (err) {
      console.error("Failed to clear deleted collections:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Get password from storage
  const handleGetStoredPassword = async () => {
    try {
      const storedPassword = await deleteCollectionManager.getUserPassword();
      if (storedPassword) {
        setPassword(storedPassword);
        setSuccess("Password loaded from storage");
        addToEventLog("password_loaded", { source: "storage" });
      } else {
        setError("No password found in storage");
      }
    } catch (err) {
      setError(`Failed to get stored password: ${err.message}`);
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

  // Search collections
  const filteredDeletedCollections = searchTerm
    ? deleteCollectionManager.searchDeletedCollections(searchTerm)
    : deletedCollections;

  // Get recent deletions
  const recentDeletions = getRecentDeletions(24);

  // Auto-clear messages after 5 seconds
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        clearMessages();
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [success, error]);

  // Check authentication status
  const isAuthenticated = authManager.isAuthenticated();
  const canDeleteCollections = authManager.canMakeAuthenticatedRequests();
  const userEmail = authManager.getCurrentUserEmail();

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <div style={{ marginBottom: "20px" }}>
        <button onClick={() => navigate("/dashboard")}>
          ← Back to Dashboard
        </button>
      </div>
      <h2>🗑️ Delete Collection Manager Example (with Manager)</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the <strong>DeleteCollectionManager</strong> for
        collection soft deletion and restoration functionality.
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
            <strong>User:</strong> {userEmail || "Not logged in"}
          </div>
          <div>
            <strong>Authenticated:</strong>{" "}
            {isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can Delete Collections:</strong>{" "}
            {canDeleteCollections ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Total Deleted:</strong> {deletedCollections.length}
          </div>
          <div>
            <strong>Recent Deletions (24h):</strong> {recentDeletions.length}
          </div>
        </div>
        {managerStatus && (
          <div style={{ marginTop: "10px", fontSize: "12px", color: "#666" }}>
            <strong>Manager Info:</strong> Listeners:{" "}
            {managerStatus.listenerCount}, Storage:{" "}
            {managerStatus.storage?.deletedCollectionsCount || 0} deleted
            collections
          </div>
        )}
      </div>

      {/* Delete Collection Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#f8d7da",
          borderRadius: "6px",
          border: "1px solid #f5c6cb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>🗑️ Delete Collection:</h4>
        <div style={{ display: "grid", gap: "15px" }}>
          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Collection ID * (Single)
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
              Collection IDs (Batch)
            </label>
            <input
              type="text"
              value={collectionIds}
              onChange={(e) => setCollectionIds(e.target.value)}
              placeholder="Enter comma-separated collection UUIDs..."
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                fontFamily: "monospace",
                fontSize: "12px",
              }}
            />
            <small style={{ color: "#666" }}>
              For batch deletion: id1,id2,id3
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
              Password (for decryption if needed)
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
              onClick={handleDeleteCollection}
              disabled={isLoading || !collectionId.trim() || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionId.trim() || !isAuthenticated
                    ? "#6c757d"
                    : "#dc3545",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionId.trim() || !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {isLoading ? "🔄 Deleting..." : "🗑️ Delete Collection"}
            </button>

            <button
              onClick={handleDeleteCollections}
              disabled={isLoading || !collectionIds.trim() || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionIds.trim() || !isAuthenticated
                    ? "#6c757d"
                    : "#dc3545",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionIds.trim() || !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              🗑️ Batch Delete
            </button>
          </div>
        </div>
      </div>

      {/* Available Collections */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#d1ecf1",
          borderRadius: "6px",
          border: "1px solid #bee5eb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>
          📁 Available Collections (for testing):
        </h4>
        {availableCollections.length === 0 ? (
          <p style={{ color: "#6c757d" }}>
            No collections available. Create some collections first using the
            Create Collection Manager.
          </p>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {availableCollections.map((collection) => (
              <div
                key={collection.id}
                style={{
                  padding: "10px",
                  border: "1px solid #dee2e6",
                  borderRadius: "4px",
                  backgroundColor: "white",
                  fontSize: "12px",
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <div style={{ fontWeight: "bold" }}>
                    {collection.collection_type === "folder" ? "📁" : "📷"}{" "}
                    {collection.name}
                  </div>
                  <div style={{ color: "#666" }}>
                    <strong>ID:</strong> {collection.id}
                  </div>
                </div>
                <button
                  onClick={() => setCollectionId(collection.id)}
                  style={{
                    padding: "5px 10px",
                    backgroundColor: "#007bff",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor: "pointer",
                    fontSize: "12px",
                  }}
                >
                  Use ID
                </button>
              </div>
            ))}
          </div>
        )}
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

      {/* Deleted Collections List */}
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
            🗑️ Deleted Collections ({deletedCollections.length}):
          </h4>
          <div style={{ display: "flex", gap: "10px" }}>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search deleted collections..."
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
            {deletedCollections.length > 0 && (
              <button
                onClick={handleClearAllDeletedCollections}
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

        {filteredDeletedCollections.length === 0 ? (
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
              {deletedCollections.length === 0
                ? "No collections deleted yet."
                : "No collections match your search."}
            </p>
            <p style={{ color: "#6c757d" }}>
              {deletedCollections.length === 0
                ? "Delete collections using the form above."
                : "Try a different search term."}
            </p>
          </div>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {filteredDeletedCollections.map((collection) => {
              const latestDeletion = getLatestDeletionForCollection(
                collection.id,
              );
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
                        🗑️ DELETED
                      </span>
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      <strong>ID:</strong> {collection.id}
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      <strong>Type:</strong> {collection.collection_type} |
                      <strong> Deleted:</strong>{" "}
                      {new Date(
                        collection.locally_deleted_at || collection.deleted_at,
                      ).toLocaleString()}
                    </div>
                    {latestDeletion && (
                      <div style={{ fontSize: "12px", color: "#dc3545" }}>
                        🗑️ Last action: {latestDeletion.action} at{" "}
                        {new Date(
                          latestDeletion.timestamp,
                        ).toLocaleTimeString()}
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
                      onClick={() => handleRestoreCollection(collection)}
                      disabled={isLoading}
                      style={{
                        padding: "5px 10px",
                        backgroundColor: "#28a745",
                        color: "white",
                        border: "none",
                        borderRadius: "4px",
                        cursor: isLoading ? "not-allowed" : "pointer",
                        fontSize: "12px",
                      }}
                    >
                      ↩️ Restore
                    </button>
                    <button
                      onClick={() => handlePermanentlyRemove(collection)}
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
                      ❌ Remove
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Deletion History */}
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
            📜 Deletion History ({deletionHistory.length}):
          </h4>
          {deletionHistory.length === 0 ? (
            <p style={{ color: "#6c757d" }}>No deletion history available.</p>
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
              {deletionHistory
                .slice()
                .reverse()
                .map((entry, index) => (
                  <div
                    key={`${entry.timestamp}-${index}`}
                    style={{
                      padding: "10px",
                      borderBottom:
                        index < deletionHistory.length - 1
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
                      <strong style={{ color: "#dc3545" }}>
                        {entry.action}
                      </strong>
                      {" - "}
                      <span style={{ color: "#666" }}>
                        {entry.collectionId}
                      </span>
                    </div>
                    {entry.collection_name && (
                      <div style={{ color: "#666", marginLeft: "20px" }}>
                        Name: {entry.collection_name}
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
          <h3>📋 Collection Deletion Event Log ({eventLog.length})</h3>
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
              No collection deletion events logged yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Events will appear here when collections are deleted, restored, or
              other actions occur.
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
                    <strong style={{ color: "#dc3545" }}>
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
          its ID here to test deletion and restoration.
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
            onClick={() =>
              setCollectionId("550e8400-e29b-41d4-a716-446655440001")
            }
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
            ⚠️ Soft delete - collections are marked as deleted but can be
            restored!
          </span>
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(DeleteCollectionManagerExample);
