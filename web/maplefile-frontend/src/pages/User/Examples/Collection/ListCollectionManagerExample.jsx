// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/Collection/ListCollectionManagerExample.jsx
// Enhanced example component demonstrating how to use the useCollectionListing hook

import React, { useState, useEffect } from "react";
import { useCollections } from "../../../../services/Services";
import { useAuth } from "../../../../services/Services";

const ListCollectionManagerExample = () => {
  const {
    listCollectionManager,
    // State
    isLoading,
    error: hookError,
    success: hookSuccess,
    collections,
    filteredCollections,
    rootCollections,
    collectionsByParent,
    managerStatus,

    // Core operations
    listCollections,
    listFilteredCollections,
    listRootCollections,
    listCollectionsByParent,

    // Cache operations
    clearAllCache,
    clearSpecificCache,

    // Utilities
    searchCollections,
    getUserPassword,
    clearMessages,

    // Status
    isAuthenticated,
    canListCollections,
    totalCollections,
    totalFilteredCollections,
    totalRootCollections,
  } = useCollections();

  const { user } = useAuth();

  // Local component state
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  // UI state
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedListType, setSelectedListType] = useState("user");
  const [parentId, setParentId] = useState("");
  const [includeOwned, setIncludeOwned] = useState(true);
  const [includeShared, setIncludeShared] = useState(false);
  const [eventLog, setEventLog] = useState([]);
  const [showDetails, setShowDetails] = useState({});

  // Handle list collections
  const handleListCollections = async (forceRefresh = false) => {
    console.log("[ListCollectionExample] Listing collections...");
    console.log("[ListCollectionExample] List type:", selectedListType);
    console.log("[ListCollectionExample] Force refresh:", forceRefresh);
    console.log(
      "[ListCollectionExample] ListCollectionManager available:",
      !!listCollectionManager,
    );
    console.log("[ListCollectionExample] Is authenticated:", isAuthenticated);

    if (!listCollectionManager) {
      alert(
        "ListCollectionManager is not available. Please check the service initialization.",
      );
      return;
    }

    if (!isAuthenticated) {
      alert("You must be authenticated to list collections.");
      return;
    }

    setError(null);
    setSuccess(null);

    try {
      let result;

      switch (selectedListType) {
        case "user":
          console.log("[ListCollectionExample] Listing user collections");
          result = await listCollections(forceRefresh);
          addToEventLog("user_collections_listed", {
            totalCount: result.totalCount,
            source: result.source,
            forceRefresh,
          });
          break;

        case "filtered":
          console.log("[ListCollectionExample] Listing filtered collections:", {
            includeOwned,
            includeShared,
          });
          result = await listFilteredCollections(
            includeOwned,
            includeShared,
            forceRefresh,
          );
          addToEventLog("filtered_collections_listed", {
            ownedCount: result.owned_collections.length,
            sharedCount: result.shared_collections.length,
            totalCount: result.total_count,
            source: result.source,
            forceRefresh,
          });
          break;

        case "root":
          console.log("[ListCollectionExample] Listing root collections");
          result = await listRootCollections(forceRefresh);
          addToEventLog("root_collections_listed", {
            totalCount: result.totalCount,
            source: result.source,
            forceRefresh,
          });
          break;

        case "byParent":
          if (!parentId.trim()) {
            alert("Parent ID is required for listing by parent");
            return;
          }
          console.log(
            "[ListCollectionExample] Listing collections by parent:",
            parentId,
          );
          result = await listCollectionsByParent(parentId.trim(), forceRefresh);
          addToEventLog("collections_by_parent_listed", {
            parentId: parentId.trim(),
            totalCount: result.totalCount,
            source: result.source,
            forceRefresh,
          });
          break;

        default:
          throw new Error(`Unknown list type: ${selectedListType}`);
      }

      console.log(
        "[ListCollectionExample] Listing completed successfully:",
        result,
      );
    } catch (err) {
      console.error("[ListCollectionExample] Collection listing failed:", err);
      addToEventLog("listing_failed", {
        listType: selectedListType,
        error: err.message,
        forceRefresh,
      });
      alert(`Collection listing failed: ${err.message}`);
    }
  };

  // Handle cache operations
  const handleClearAllCache = async () => {
    if (!confirm("Clear ALL collection list cache? This cannot be undone."))
      return;

    try {
      await clearAllCache();
      addToEventLog("all_cache_cleared", {});
    } catch (err) {
      console.error("Failed to clear all cache:", err);
    }
  };

  const handleClearSpecificCache = async (cacheType) => {
    if (!confirm(`Clear ${cacheType} cache? This cannot be undone.`)) return;

    try {
      await clearSpecificCache(cacheType);
      addToEventLog("specific_cache_cleared", { cacheType });
    } catch (err) {
      console.error(`Failed to clear ${cacheType} cache:`, err);
    }
  };

  // Get password from storage
  const handleGetStoredPassword = async () => {
    try {
      const storedPassword = await getUserPassword();
      if (storedPassword) {
        addToEventLog("password_loaded", { source: "storage" });
        alert("Password loaded from storage successfully");
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

  // Toggle collection details
  const toggleDetails = (collectionId) => {
    setShowDetails((prev) => ({
      ...prev,
      [collectionId]: !prev[collectionId],
    }));
  };

  // This function seems to be missing from the useCollections hook, so we define it locally.
  const filterCollectionsByType = (collectionsList, type) => {
    if (!Array.isArray(collectionsList)) return [];
    return collectionsList.filter(
      (collection) => collection.collection_type === type,
    );
  };

  // Ensure collections is an array for safe operations
  const safeCollections = collections || [];

  // Search filtered collections
  const searchResults = searchTerm
    ? searchCollections(searchTerm, safeCollections)
    : safeCollections;

  // Filter results by type
  const folders = filterCollectionsByType(safeCollections, "folder");
  const albums = filterCollectionsByType(safeCollections, "album");

  // Get cached data
  const cachedData = listCollectionManager?.getCachedCollections() || {
    collections: [],
    isExpired: true,
  };
  const cachedFilteredData =
    listCollectionManager?.getCachedFilteredCollections() || {
      total_count: 0,
      isExpired: true,
    };

  // Debug: Log when component loads
  useEffect(() => {
    console.log("[ListCollectionExample] Component mounted");
    console.log(
      "[ListCollectionExample] listCollectionManager from useCollections:",
      listCollectionManager,
    );
    console.log("[ListCollectionExample] useCollectionListing hook data:", {
      isLoading,
      error,
      success,
      isAuthenticated,
      canListCollections,
      listCollectionsFunction: typeof listCollections,
    });
  }, []);

  // Debug: Log when listCollectionManager changes
  useEffect(() => {
    console.log(
      "[ListCollectionExample] listCollectionManager changed:",
      !!listCollectionManager,
    );
    if (listCollectionManager) {
      console.log(
        "[ListCollectionExample] listCollectionManager methods:",
        Object.getOwnPropertyNames(listCollectionManager),
      );
    }
  }, [listCollectionManager]);

  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        clearMessages();
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [success, error, clearMessages]);

  return (
    <div style={{ padding: "20px", maxWidth: "1600px", margin: "0 auto" }}>
      <h2>📂 Enhanced List Collection Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the <strong>useCollectionListing</strong> hook
        with multiple list types, caching, and E2EE decryption.
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
            <strong>ListCollectionManager Available:</strong>{" "}
            {listCollectionManager ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>UseCollectionListing Hook:</strong>{" "}
            {typeof useCollections === "function"
              ? "✅ Loaded"
              : "❌ Not loaded"}
          </div>
          <div>
            <strong>List Collections Function:</strong>{" "}
            {typeof listCollections === "function"
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
            <strong>Selected List Type:</strong> {selectedListType}
          </div>
        </div>
        <button
          onClick={() => {
            console.log("=== DEBUG INFO ===");
            console.log("listCollectionManager:", listCollectionManager);
            console.log("useCollectionListing functions:", {
              listCollections: typeof listCollections,
              listFilteredCollections: typeof listFilteredCollections,
              listRootCollections: typeof listRootCollections,
            });
            console.log("Hook state:", {
              isLoading,
              hookError,
              hookSuccess,
              isAuthenticated,
              canListCollections,
            });
            console.log("Local state:", {
              error,
              success,
              selectedListType,
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
            console.log("=== DIRECT SERVICE TEST ===");
            try {
              if (listCollectionManager) {
                console.log(
                  "Calling listCollectionManager.listCollections directly...",
                );
                const result =
                  await listCollectionManager.listCollections(true);
                console.log("Direct call result:", result);
                alert(
                  `Direct call successful! Listed ${result.totalCount} collections from ${result.source}`,
                );
              } else {
                console.log("listCollectionManager is not available");
                alert("ListCollectionManager service is not available");
              }
            } catch (err) {
              console.error("Direct call error:", err);
              alert(`Direct call failed: ${err.message}`);
            }
            console.log("========================");
          }}
          style={{
            marginTop: "10px",
            padding: "5px 10px",
            backgroundColor: "#28a745",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
          }}
        >
          Test Direct Service Call
        </button>
      </div>

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
            <strong>Authenticated:</strong>{" "}
            {isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can List Collections:</strong>{" "}
            {canListCollections ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Total User Collections:</strong> {totalCollections}
          </div>
          <div>
            <strong>Total Filtered Collections:</strong>{" "}
            {totalFilteredCollections}
          </div>
          <div>
            <strong>Total Root Collections:</strong> {totalRootCollections}
          </div>
        </div>
      </div>

      {/* List Collections Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>📂 List Collections:</h4>
        <div style={{ display: "grid", gap: "15px" }}>
          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              List Type
            </label>
            <select
              value={selectedListType}
              onChange={(e) => setSelectedListType(e.target.value)}
              style={{
                width: "300px",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            >
              <option value="user">👤 User Collections (All owned)</option>
              <option value="filtered">
                🔍 Filtered Collections (Owned/Shared)
              </option>
              <option value="root">🏠 Root Collections (No parent)</option>
              <option value="byParent">👨‍👩‍👧‍👦 Collections by Parent</option>
            </select>
          </div>

          {selectedListType === "filtered" && (
            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                Filter Options
              </label>
              <div style={{ display: "flex", gap: "20px" }}>
                <label style={{ display: "flex", alignItems: "center" }}>
                  <input
                    type="checkbox"
                    checked={includeOwned}
                    onChange={(e) => setIncludeOwned(e.target.checked)}
                    style={{ marginRight: "5px" }}
                  />
                  Include Owned Collections
                </label>
                <label style={{ display: "flex", alignItems: "center" }}>
                  <input
                    type="checkbox"
                    checked={includeShared}
                    onChange={(e) => setIncludeShared(e.target.checked)}
                    style={{ marginRight: "5px" }}
                  />
                  Include Shared Collections
                </label>
              </div>
            </div>
          )}

          {selectedListType === "byParent" && (
            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                Parent Collection ID *
              </label>
              <input
                type="text"
                value={parentId}
                onChange={(e) => setParentId(e.target.value)}
                placeholder="Enter parent collection UUID..."
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
          )}

          <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
            <button
              onClick={() => handleListCollections(false)}
              disabled={isLoading || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !isAuthenticated ? "#6c757d" : "#28a745",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !isAuthenticated ? "not-allowed" : "pointer",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {isLoading ? "🔄 Listing..." : "📂 List Collections"}
            </button>

            <button
              onClick={() => handleListCollections(true)}
              disabled={isLoading || !isAuthenticated}
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading || !isAuthenticated ? "#6c757d" : "#007bff",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading || !isAuthenticated ? "not-allowed" : "pointer",
                fontSize: "14px",
              }}
            >
              🔄 Force Refresh
            </button>

            <button
              onClick={handleGetStoredPassword}
              style={{
                padding: "12px 20px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: "pointer",
                fontSize: "14px",
              }}
            >
              🔑 Get Stored Password
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

      {/* Collections Display */}
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
            📋 Collections ({totalCollections} total):
          </h4>
          <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
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
            <span style={{ fontSize: "12px", color: "#666" }}>
              📁 {folders.length} folders | 📷 {albums.length} albums
            </span>
          </div>
        </div>

        {searchResults.length === 0 ? (
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
              {totalCollections === 0
                ? "No collections found."
                : "No collections match your search."}
            </p>
            <p style={{ color: "#6c757d" }}>
              Use the form above to list collections.
            </p>
          </div>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {searchResults.map((collection) => (
              <div
                key={collection.id}
                style={{
                  padding: "15px",
                  border: "1px solid #dee2e6",
                  borderRadius: "6px",
                  backgroundColor: "white",
                }}
              >
                <div
                  style={{
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
                        v{collection.version || "?"}
                      </span>
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      <strong>ID:</strong> {collection.id}
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      <strong>Type:</strong> {collection.collection_type}
                      {" | "}
                      <strong>Owner:</strong> {collection.owner_id}
                      {" | "}
                      <strong>Created:</strong>{" "}
                      {new Date(collection.created_at).toLocaleString()}
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      <strong>Decrypted:</strong>{" "}
                      {collection._isDecrypted ? "✅ Yes" : "❌ No"}
                      {collection._decryptionError && (
                        <span style={{ color: "#dc3545", marginLeft: "10px" }}>
                          Error: {collection._decryptionError}
                        </span>
                      )}
                    </div>
                  </div>
                  <div style={{ display: "flex", gap: "10px" }}>
                    <button
                      onClick={() => toggleDetails(collection.id)}
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
                      {showDetails[collection.id] ? "Hide" : "Show"} Details
                    </button>
                  </div>
                </div>

                {showDetails[collection.id] && (
                  <div style={{ marginTop: "10px" }}>
                    <pre
                      style={{
                        backgroundColor: "#f8f9fa",
                        padding: "10px",
                        borderRadius: "4px",
                        fontSize: "11px",
                        overflow: "auto",
                        maxHeight: "200px",
                        fontFamily: "monospace",
                      }}
                    >
                      {JSON.stringify(collection, null, 2)}
                    </pre>
                  </div>
                )}
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
          <h4 style={{ margin: 0 }}>💾 Cache Management:</h4>
          <div style={{ display: "flex", gap: "10px" }}>
            <button
              onClick={() => handleClearSpecificCache("listed")}
              style={{
                padding: "5px 15px",
                backgroundColor: "#ffc107",
                color: "black",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              Clear User Cache
            </button>
            <button
              onClick={() => handleClearSpecificCache("filtered")}
              style={{
                padding: "5px 15px",
                backgroundColor: "#ffc107",
                color: "black",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              Clear Filtered Cache
            </button>
            <button
              onClick={() => handleClearSpecificCache("root")}
              style={{
                padding: "5px 15px",
                backgroundColor: "#ffc107",
                color: "black",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              Clear Root Cache
            </button>
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
          </div>
        </div>

        <div style={{ fontSize: "12px", color: "#666" }}>
          <div>
            <strong>Cached User Collections:</strong>{" "}
            {cachedData.collections?.length || 0} (Expired:{" "}
            {cachedData.isExpired ? "Yes" : "No"})
          </div>
          <div>
            <strong>Cached Filtered Collections:</strong>{" "}
            {cachedFilteredData.total_count} (Expired:{" "}
            {cachedFilteredData.isExpired ? "Yes" : "No"})
          </div>
          <div>
            <strong>Cache Status:</strong>{" "}
            {managerStatus?.storage?.hasListedCollections
              ? "Has cached data"
              : "No cached data"}
          </div>
        </div>
      </div>

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
          <h3>📋 Collection Listing Event Log ({eventLog.length})</h3>
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
              No listing events logged yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Events will appear here when collections are listed or operations
              occur.
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
          Try different list types: User Collections (all owned), Filtered
          Collections (owned/shared), Root Collections (no parent), Collections
          by Parent.
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
              setSelectedListType("user");
              handleListCollections(false);
            }}
            style={{
              padding: "5px 10px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "3px",
              cursor: "pointer",
              fontSize: "12px",
            }}
          >
            List User Collections
          </button>
          <button
            onClick={() => {
              setSelectedListType("root");
              handleListCollections(false);
            }}
            style={{
              padding: "5px 10px",
              backgroundColor: "#007bff",
              color: "white",
              border: "none",
              borderRadius: "3px",
              cursor: "pointer",
              fontSize: "12px",
            }}
          >
            List Root Collections
          </button>
          <span style={{ fontSize: "12px", color: "#666" }}>
            Use cache first, then refresh to see API vs cache behavior
          </span>
        </div>
      </div>
    </div>
  );
};

export default ListCollectionManagerExample;
