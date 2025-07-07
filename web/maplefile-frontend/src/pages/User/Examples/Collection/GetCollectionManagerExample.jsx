// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/Collection/GetCollectionManagerExample.jsx
// Enhanced example component demonstrating how to use the useCollectionRetrieval hook

import React, { useState, useEffect } from "react";
import useCollectionRetrieval from "../../../../hooks/Collection/useCollectionRetrieval.jsx";
import { useCollections } from "../../../../hooks/useService.jsx";
import useAuth from "../../../../hooks/useAuth.js";

const GetCollectionManagerExample = () => {
  const { getCollectionManager } = useCollections();
  const {
    // State
    isLoading,
    error: hookError,
    success: hookSuccess,
    cachedCollections,
    managerStatus,

    // Core operations
    getCollection,
    getCachedCollection,
    refreshCollection,
    getCollections,
    collectionExists,

    // Cache operations
    getCollectionCacheStatus,
    removeFromCache,
    clearAllCache,
    clearExpiredCollections,

    // Utilities
    searchCachedCollections,
    getUserPassword,
    clearMessages,

    // Status
    isAuthenticated,
    canGetCollections,
    totalCachedCollections,
    isCached,
    isExpired,
  } = useCollectionRetrieval();

  const { user, logout } = useAuth();

  // Local component state
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  // Form state
  const [collectionId, setCollectionId] = useState("");
  const [password, setPassword] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [operationType, setOperationType] = useState("auto"); // auto, cache_only, force_api

  // UI state
  const [retrievalResults, setRetrievalResults] = useState([]);
  const [selectedCollection, setSelectedCollection] = useState(null);
  const [eventLog, setEventLog] = useState([]);
  const [showCacheDetails, setShowCacheDetails] = useState(false);

  // Handle collection retrieval with different options
  const handleGetCollection = async () => {
    if (!collectionId.trim()) {
      alert("Collection ID is required");
      return;
    }

    console.log("[GetCollectionExample] Starting collection retrieval...");
    console.log("[GetCollectionExample] Collection ID:", collectionId.trim());
    console.log("[GetCollectionExample] Operation Type:", operationType);
    console.log(
      "[GetCollectionExample] GetCollectionManager available:",
      !!getCollectionManager,
    );
    console.log("[GetCollectionExample] Is authenticated:", isAuthenticated);

    if (!getCollectionManager) {
      alert(
        "GetCollectionManager is not available. Please check the service initialization.",
      );
      return;
    }

    if (!isAuthenticated) {
      alert("You must be authenticated to retrieve collections.");
      return;
    }

    setError(null);
    setSuccess(null);

    try {
      let result;
      const id = collectionId.trim();

      console.log(
        `[GetCollectionExample] Starting ${operationType} operation for:`,
        id,
      );

      switch (operationType) {
        case "cache_only":
          console.log("[GetCollectionExample] Attempting cache-only retrieval");
          result = await getCachedCollection(id);
          addToEventLog("cache_only_retrieval", {
            id,
            success: !!result,
            source: result?.source || "cache_miss",
          });
          break;

        case "force_api":
          console.log(
            "[GetCollectionExample] Forcing API call (bypassing cache)",
          );
          result = await refreshCollection(id);
          addToEventLog("force_api_retrieval", {
            id,
            success: !!result,
            source: result?.source || "api",
          });
          break;

        case "auto":
        default:
          console.log(
            "[GetCollectionExample] Auto retrieval (cache first, then API)",
          );
          result = await getCollection(id, false);
          addToEventLog("auto_retrieval", {
            id,
            success: !!result,
            source: result?.source || "unknown",
          });
          break;
      }

      console.log("[GetCollectionExample] Retrieval result:", result);

      if (result && result.collection) {
        // Add to results list
        setRetrievalResults((prev) => [
          {
            ...result,
            retrievedAt: new Date().toISOString(),
            operationType,
            id: Date.now(), // For React key
          },
          ...prev.slice(0, 9), // Keep last 10 results
        ]);

        setSelectedCollection(result.collection);
        console.log(
          "[GetCollectionExample] Collection retrieved successfully:",
          result,
        );
      } else {
        console.warn(
          "[GetCollectionExample] No collection data in result:",
          result,
        );
        alert("No collection data returned. Check console for details.");
      }
    } catch (err) {
      console.error("[GetCollectionExample] Collection retrieval failed:", err);
      addToEventLog("retrieval_failed", {
        id: collectionId,
        operationType,
        error: err.message,
      });
      alert(`Collection retrieval failed: ${err.message}`);
    }
  };

  // Handle multiple collection retrieval
  const handleGetMultipleCollections = async () => {
    const ids = collectionId
      .split(",")
      .map((id) => id.trim())
      .filter((id) => id);

    if (ids.length === 0) {
      alert("Enter one or more collection IDs (comma-separated)");
      return;
    }

    try {
      console.log("[GetCollectionExample] Getting multiple collections:", ids);
      const forceRefresh = operationType === "force_api";

      const result = await getCollections(ids, forceRefresh);

      addToEventLog("multiple_collections_retrieved", {
        requestedIds: ids,
        successCount: result.successCount,
        errorCount: result.errorCount,
        forceRefresh,
      });

      // Add successful results to the list
      result.collections.forEach((collectionResult) => {
        setRetrievalResults((prev) => [
          {
            ...collectionResult,
            retrievedAt: new Date().toISOString(),
            operationType: `batch_${operationType}`,
            id: Date.now() + Math.random(), // For React key
          },
          ...prev.slice(0, 9),
        ]);
      });

      console.log(
        "[GetCollectionExample] Multiple collections result:",
        result,
      );
    } catch (err) {
      console.error(
        "[GetCollectionExample] Multiple collections retrieval failed:",
        err,
      );
      addToEventLog("multiple_retrieval_failed", {
        requestedIds: ids,
        error: err.message,
      });
    }
  };

  // Check if collection exists
  const handleCheckExists = async () => {
    if (!collectionId.trim()) {
      alert("Collection ID is required");
      return;
    }

    try {
      console.log(
        "[GetCollectionExample] Checking collection existence:",
        collectionId,
      );
      const exists = await collectionExists(collectionId.trim());

      addToEventLog("existence_check", {
        id: collectionId,
        exists,
      });

      alert(
        `Collection ${exists ? "exists" : "does not exist or you don't have access"}`,
      );
    } catch (err) {
      console.error("[GetCollectionExample] Existence check failed:", err);
      addToEventLog("existence_check_failed", {
        id: collectionId,
        error: err.message,
      });
    }
  };

  // Get cache status
  const handleGetCacheStatus = () => {
    if (!collectionId.trim()) {
      alert("Collection ID is required");
      return;
    }

    const status = getCollectionCacheStatus(collectionId.trim());

    addToEventLog("cache_status_checked", {
      id: collectionId,
      status,
    });

    alert(`Cache Status:\n${JSON.stringify(status, null, 2)}`);
  };

  // Clear specific collection from cache
  const handleRemoveFromCache = async (id) => {
    if (!confirm(`Remove collection ${id} from cache?`)) return;

    try {
      await removeFromCache(id);
      addToEventLog("collection_removed_from_cache", { id });
    } catch (err) {
      console.error("Failed to remove from cache:", err);
    }
  };

  // Clear all cache
  const handleClearAllCache = async () => {
    if (!confirm("Clear ALL cached collections? This cannot be undone."))
      return;

    try {
      await clearAllCache();
      setRetrievalResults([]);
      setSelectedCollection(null);
      addToEventLog("all_cache_cleared", {});
    } catch (err) {
      console.error("Failed to clear all cache:", err);
    }
  };

  // Clear expired collections
  const handleClearExpired = async () => {
    try {
      const count = await clearExpiredCollections();
      addToEventLog("expired_collections_cleared", { count });
      alert(`Cleared ${count} expired collections from cache`);
    } catch (err) {
      console.error("Failed to clear expired collections:", err);
    }
  };

  // Get password from storage
  const handleGetStoredPassword = async () => {
    try {
      const storedPassword = await getUserPassword();
      if (storedPassword) {
        setPassword(storedPassword);
        addToEventLog("password_loaded", { source: "storage" });
      } else {
        alert("No password found in storage");
      }
    } catch (err) {
      alert(`Failed to get stored password: ${err.message}`);
    }
  };

  // Add event to log
  const addToEventLog = (eventType, eventData) => {
    setEventLog((prev) => [
      {
        timestamp: new Date().toISOString(),
        eventType,
        eventData,
      },
      ...prev.slice(0, 49), // Keep last 50 events
    ]);
  };

  // Clear event log
  const handleClearLog = () => {
    setEventLog([]);
  };

  // Search cached collections
  const filteredCachedCollections = searchTerm
    ? searchCachedCollections(searchTerm)
    : cachedCollections;

  // Debug: Log when component loads
  useEffect(() => {
    console.log("[GetCollectionExample] Component mounted");
    console.log(
      "[GetCollectionExample] getCollectionManager from useCollections:",
      getCollectionManager,
    );
    console.log("[GetCollectionExample] useCollectionRetrieval hook data:", {
      isLoading,
      error,
      success,
      isAuthenticated,
      canGetCollections,
      getCollectionFunction: typeof getCollection,
    });
  }, []);

  // Debug: Log when getCollectionManager changes
  useEffect(() => {
    console.log(
      "[GetCollectionExample] getCollectionManager changed:",
      !!getCollectionManager,
    );
    if (getCollectionManager) {
      console.log(
        "[GetCollectionExample] getCollectionManager methods:",
        Object.getOwnPropertyNames(getCollectionManager),
      );
    }
  }, [getCollectionManager]);
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        clearMessages();
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [success, error, clearMessages]);

  return (
    <div style={{ padding: "20px", maxWidth: "1400px", margin: "0 auto" }}>
      <h2>🔍 Enhanced Get Collection Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the <strong>useCollectionRetrieval</strong> hook
        with cache management, API calls, and E2EE decryption.
        <br />
        <strong>User:</strong> {user?.email || "Not logged in"}
      </p>

      {/* Debug Section */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#fff3cd",
          borderRadius: "6px",
          border: "1px solid #ffeaa7",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>🐛 Debug Information:</h4>
        <div style={{ fontSize: "14px", fontFamily: "monospace" }}>
          <div>
            <strong>GetCollectionManager Available:</strong>{" "}
            {getCollectionManager ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>UseCollectionRetrieval Hook:</strong>{" "}
            {typeof useCollectionRetrieval === "function"
              ? "✅ Loaded"
              : "❌ Not loaded"}
          </div>
          <div>
            <strong>Get Collection Function:</strong>{" "}
            {typeof getCollection === "function"
              ? "✅ Available"
              : "❌ Not available"}
          </div>
          <div>
            <strong>Is Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Local Error:</strong> {error || "None"}
          </div>
          <div>
            <strong>Hook Error:</strong> {hookError || "None"}
          </div>
          <div>
            <strong>Local Success:</strong> {success || "None"}
          </div>
          <div>
            <strong>Hook Success:</strong> {hookSuccess || "None"}
          </div>
          <div>
            <strong>Collection ID:</strong> {collectionId || "Empty"}
          </div>
          <div>
            <strong>Operation Type:</strong> {operationType}
          </div>
        </div>
        <button
          onClick={() => {
            console.log("=== DEBUG INFO ===");
            console.log("getCollectionManager:", getCollectionManager);
            console.log("useCollectionRetrieval functions:", {
              getCollection: typeof getCollection,
              getCachedCollection: typeof getCachedCollection,
              refreshCollection: typeof refreshCollection,
            });
            console.log("Hook state:", {
              isLoading,
              hookError,
              hookSuccess,
              isAuthenticated,
              canGetCollections,
            });
            console.log("Local state:", {
              error,
              success,
              collectionId,
              operationType,
            });
            console.log("Manager status:", managerStatus);
            console.log("==================");
          }}
          style={{
            marginTop: "10px",
            padding: "5px 10px",
            backgroundColor: "#007bff",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
            marginRight: "10px",
          }}
        >
          Log Debug Info to Console
        </button>
        <button
          onClick={async () => {
            if (!collectionId.trim()) {
              alert("Enter a Collection ID first");
              return;
            }

            console.log("=== DIRECT SERVICE TEST ===");
            try {
              if (getCollectionManager) {
                console.log(
                  "Calling getCollectionManager.getCollection directly...",
                );
                const result = await getCollectionManager.getCollection(
                  collectionId.trim(),
                  true,
                );
                console.log("Direct call result:", result);
                alert(`Direct call successful! Source: ${result.source}`);
              } else {
                console.log("getCollectionManager is not available");
                alert("GetCollectionManager service is not available");
              }
            } catch (err) {
              console.error("Direct call error:", err);
              alert(`Direct call failed: ${err.message}`);
            }
            console.log("========================");
          }}
          disabled={!collectionId.trim()}
          style={{
            marginTop: "10px",
            padding: "5px 10px",
            backgroundColor: !collectionId.trim() ? "#6c757d" : "#28a745",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: !collectionId.trim() ? "not-allowed" : "pointer",
          }}
        >
          Test Direct Service Call
        </button>
      </div>
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
            <strong>Authenticated:</strong>{" "}
            {isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can Get Collections:</strong>{" "}
            {canGetCollections ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Total Cached:</strong> {totalCachedCollections}
          </div>
          <div>
            <strong>Cache Hit Rate:</strong>{" "}
            {managerStatus.cache?.hitRate || "N/A"}
          </div>
          <div>
            <strong>Expired in Cache:</strong>{" "}
            {managerStatus.cache?.expired || 0}
          </div>
        </div>
      </div>

      {/* Get Collection Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>🔍 Retrieve Collection(s):</h4>
        <div style={{ display: "grid", gap: "15px" }}>
          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Collection ID(s) *
            </label>
            <input
              type="text"
              value={collectionId}
              onChange={(e) => setCollectionId(e.target.value)}
              placeholder="Enter collection UUID or comma-separated UUIDs..."
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
              Single ID for single retrieval, comma-separated for batch
              retrieval
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
              Retrieval Method
            </label>
            <select
              value={operationType}
              onChange={(e) => setOperationType(e.target.value)}
              style={{
                width: "300px",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            >
              <option value="auto">🤖 Auto (Cache first, then API)</option>
              <option value="cache_only">💾 Cache Only (No API call)</option>
              <option value="force_api">
                🌐 Force API Call (Bypass cache)
              </option>
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
              Password (for decryption)
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
              onClick={handleGetCollection}
              disabled={isLoading || !collectionId.trim() || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionId.trim() || !isAuthenticated
                    ? "#6c757d"
                    : "#28a745",
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
              {isLoading ? "🔄 Getting..." : "🔍 Get Collection"}
            </button>

            <button
              onClick={handleGetMultipleCollections}
              disabled={isLoading || !collectionId.trim() || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionId.trim() || !isAuthenticated
                    ? "#6c757d"
                    : "#007bff",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionId.trim() || !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              📦 Batch Get
            </button>

            <button
              onClick={handleCheckExists}
              disabled={isLoading || !collectionId.trim() || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !collectionId.trim() || !isAuthenticated
                    ? "#6c757d"
                    : "#17a2b8",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !collectionId.trim() || !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              ❓ Check Exists
            </button>

            <button
              onClick={handleGetCacheStatus}
              disabled={!collectionId.trim()}
              style={{
                padding: "12px 20px",
                backgroundColor: !collectionId.trim() ? "#6c757d" : "#ffc107",
                color: "black",
                border: "none",
                borderRadius: "6px",
                cursor: !collectionId.trim() ? "not-allowed" : "pointer",
                fontSize: "14px",
              }}
            >
              📋 Cache Status
            </button>
          </div>
        </div>
      </div>

      {/* Success/Error Messages */}
      {(success || hookSuccess) && (
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
          <span>✅ {success || hookSuccess}</span>
          <button
            onClick={() => {
              setSuccess(null);
              clearMessages();
            }}
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

      {(error || hookError) && (
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
          <span>❌ {error || hookError}</span>
          <button
            onClick={() => {
              setError(null);
              clearMessages();
            }}
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

      {/* Retrieval Results */}
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
            📋 Retrieval Results ({retrievalResults.length}):
          </h4>
          <button
            onClick={() => setRetrievalResults([])}
            disabled={retrievalResults.length === 0}
            style={{
              padding: "5px 15px",
              backgroundColor:
                retrievalResults.length === 0 ? "#6c757d" : "#dc3545",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: retrievalResults.length === 0 ? "not-allowed" : "pointer",
            }}
          >
            🗑️ Clear Results
          </button>
        </div>

        {retrievalResults.length === 0 ? (
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
              No retrieval results yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Use the form above to retrieve collections.
            </p>
          </div>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {retrievalResults.map((result) => (
              <div
                key={result.id}
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
                    {result.collection?.collection_type === "folder"
                      ? "📁"
                      : "📷"}{" "}
                    {result.collection?.name || "[Encrypted]"}
                    <span
                      style={{
                        fontSize: "12px",
                        color: "#666",
                        marginLeft: "10px",
                      }}
                    >
                      v{result.collection?.version || "?"}
                    </span>
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <strong>ID:</strong> {result.collection?.id || "N/A"}
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <strong>Source:</strong>
                    <span
                      style={{
                        backgroundColor:
                          result.source === "cache"
                            ? "#fff3cd"
                            : result.source === "api"
                              ? "#d1ecf1"
                              : "#e2e3e5",
                        padding: "2px 6px",
                        borderRadius: "3px",
                        marginLeft: "5px",
                      }}
                    >
                      {result.source || "unknown"}
                    </span>
                    {" | "}
                    <strong>Operation:</strong> {result.operationType}
                    {" | "}
                    <strong>Retrieved:</strong>{" "}
                    {new Date(result.retrievedAt).toLocaleTimeString()}
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <strong>Decrypted:</strong>{" "}
                    {result.collection?._isDecrypted ? "✅ Yes" : "❌ No"}
                    {result.collection?._decryptionError && (
                      <span style={{ color: "#dc3545", marginLeft: "10px" }}>
                        Error: {result.collection._decryptionError}
                      </span>
                    )}
                  </div>
                </div>
                <div style={{ display: "flex", gap: "10px" }}>
                  <button
                    onClick={() => setSelectedCollection(result.collection)}
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
                    🔍 View Details
                  </button>
                  {result.collection?.id && isCached(result.collection.id) && (
                    <button
                      onClick={() =>
                        handleRemoveFromCache(result.collection.id)
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
                      🗑️ Remove Cache
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Cache Management */}
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
          <h4 style={{ margin: 0 }}>
            💾 Cache Management ({totalCachedCollections} cached):
          </h4>
          <div style={{ display: "flex", gap: "10px" }}>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search cached collections..."
              style={{
                padding: "5px 10px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                fontSize: "14px",
              }}
            />
            <button
              onClick={() => setShowCacheDetails(!showCacheDetails)}
              style={{
                padding: "5px 15px",
                backgroundColor: "#007bff",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              {showCacheDetails ? "Hide" : "Show"} Cache Details
            </button>
            <button
              onClick={handleClearExpired}
              style={{
                padding: "5px 15px",
                backgroundColor: "#ffc107",
                color: "black",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              🧹 Clear Expired
            </button>
            {totalCachedCollections > 0 && (
              <button
                onClick={handleClearAllCache}
                style={{
                  padding: "5px 15px",
                  backgroundColor: "#dc3545",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                }}
              >
                🗑️ Clear All Cache
              </button>
            )}
          </div>
        </div>

        {showCacheDetails && (
          <div style={{ marginTop: "15px" }}>
            {filteredCachedCollections.length === 0 ? (
              <p style={{ color: "#6c757d" }}>No cached collections found.</p>
            ) : (
              <div style={{ display: "grid", gap: "8px" }}>
                {filteredCachedCollections.map((collection) => {
                  const cacheStatus = getCollectionCacheStatus(collection.id);
                  return (
                    <div
                      key={collection.id}
                      style={{
                        padding: "10px",
                        border: "1px solid #dee2e6",
                        borderRadius: "4px",
                        backgroundColor: cacheStatus.expired
                          ? "#f8d7da"
                          : "white",
                        fontSize: "12px",
                      }}
                    >
                      <div style={{ fontWeight: "bold" }}>
                        {collection.collection_type === "folder" ? "📁" : "📷"}{" "}
                        {collection.name || "[Encrypted]"}
                        {cacheStatus.expired && (
                          <span
                            style={{ color: "#dc3545", marginLeft: "10px" }}
                          >
                            ⚠️ EXPIRED
                          </span>
                        )}
                      </div>
                      <div style={{ color: "#666" }}>
                        <strong>ID:</strong> {collection.id}
                      </div>
                      <div style={{ color: "#666" }}>
                        <strong>Cached:</strong>{" "}
                        {new Date(cacheStatus.cachedAt).toLocaleString()}
                        {" | "}
                        <strong>Expires:</strong>{" "}
                        {new Date(cacheStatus.expiresAt).toLocaleString()}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Selected Collection Details */}
      {selectedCollection && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#e2e3e5",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
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
                color: "#6c757d",
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
              fontSize: "11px",
              overflow: "auto",
              maxHeight: "300px",
              fontFamily: "monospace",
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
          <h3>📋 Retrieval Event Log ({eventLog.length})</h3>
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
              No retrieval events logged yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Events will appear here when collections are retrieved or
              operations occur.
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
            {eventLog.map((event, index) => (
              <div
                key={`${event.timestamp}-${index}`}
                style={{
                  padding: "10px",
                  borderBottom:
                    index < eventLog.length - 1 ? "1px solid #dee2e6" : "none",
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
          First create a collection in the Create Collection Manager, then test
          different retrieval methods:
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
              setCollectionId("7f558adb-57b6-11f0-8b98-c60a0c48537c")
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
            Try different retrieval methods: Auto (cache+API), Cache Only, Force
            API
          </span>
        </div>
      </div>
    </div>
  );
};

export default GetCollectionManagerExample;
