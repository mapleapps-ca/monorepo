import React, { useState, useEffect, useCallback } from "react";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";

const CollectionsList = () => {
  const { collectionService } = useServices();
  const { isAuthenticated } = useAuth();

  // State management
  const [ownedCollections, setOwnedCollections] = useState([]);
  const [sharedCollections, setSharedCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [viewMode, setViewMode] = useState("grid"); // 'grid' or 'list'

  // Client-side pagination (API endpoints don't support server-side pagination for rich data)
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);

  // Filter state - only what's supported by GET /collections/filtered
  const [includeOwned, setIncludeOwned] = useState(true);
  const [includeShared, setIncludeShared] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [typeFilter, setTypeFilter] = useState("all"); // 'all', 'folder', 'album'

  // Load collections using GET /collections/filtered
  const loadCollections = useCallback(async () => {
    if (!isAuthenticated) return;

    try {
      setLoading(true);
      setError(null);

      // API: GET /collections/filtered?include_owned={bool}&include_shared={bool}
      const response = await collectionService.getFilteredCollections(
        includeOwned,
        includeShared,
      );

      console.log("Collections API response:", response);

      // Validate response structure
      const ownedData = response.owned_collections || [];
      const sharedData = response.shared_collections || [];

      // Log any collections missing IDs for debugging
      ownedData.forEach((collection, index) => {
        if (!collection || !collection.id) {
          console.warn(
            `Owned collection at index ${index} missing ID:`,
            collection,
          );
        }
      });

      sharedData.forEach((collection, index) => {
        if (!collection || !collection.id) {
          console.warn(
            `Shared collection at index ${index} missing ID:`,
            collection,
          );
        }
      });

      setOwnedCollections(ownedData);
      setSharedCollections(sharedData);
    } catch (err) {
      console.error("Failed to load collections:", err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [collectionService, isAuthenticated, includeOwned, includeShared]);

  // Refresh collections
  const refresh = useCallback(() => {
    setCurrentPage(1);
    loadCollections();
  }, [loadCollections]);

  // Initial load and when filters change
  useEffect(() => {
    if (isAuthenticated) {
      refresh();
    }
  }, [isAuthenticated, includeOwned, includeShared, refresh]);

  // Combine and filter collections (client-side since API doesn't support filtering)
  const allCollections = [...ownedCollections, ...sharedCollections].filter(
    (collection) => collection && collection.id, // Only include collections with valid IDs
  );

  const filteredCollections = allCollections.filter((collection) => {
    // Safety check
    if (!collection || !collection.id) {
      return false;
    }

    // Collection type filter (API supports 'folder' and 'album')
    if (typeFilter !== "all" && collection.collection_type !== typeFilter) {
      return false;
    }

    // Search filter (client-side since API doesn't support search)
    if (
      searchTerm &&
      collection.encrypted_name &&
      !collection.encrypted_name
        .toLowerCase()
        .includes(searchTerm.toLowerCase())
    ) {
      return false;
    }

    return true;
  });

  // Client-side pagination
  const totalPages = Math.ceil(filteredCollections.length / pageSize);
  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const paginatedCollections = filteredCollections.slice(startIndex, endIndex);

  // Collection actions - using exact API endpoints
  const handleDelete = async (collectionId) => {
    if (!collectionId) {
      console.error("Cannot delete collection: missing ID");
      return;
    }

    if (
      !window.confirm(
        "Are you sure you want to delete this collection? This action will soft-delete the collection, making it recoverable for 30 days.",
      )
    ) {
      return;
    }

    try {
      setLoading(true);
      // API: DELETE /collections/{collection_id}
      await collectionService.deleteCollection(collectionId);
      refresh(); // Reload to reflect changes
    } catch (err) {
      console.error("Failed to delete collection:", err);
      alert(`Failed to delete collection: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleArchive = async (collectionId) => {
    if (!collectionId) {
      console.error("Cannot archive collection: missing ID");
      return;
    }

    try {
      setLoading(true);
      // API: POST /collections/{collection_id}/archive
      await collectionService.archiveCollection(collectionId);
      refresh(); // Reload to reflect changes
    } catch (err) {
      console.error("Failed to archive collection:", err);
      alert(`Failed to archive collection: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleRestore = async (collectionId) => {
    if (!collectionId) {
      console.error("Cannot restore collection: missing ID");
      return;
    }

    try {
      setLoading(true);
      // API: POST /collections/{collection_id}/restore
      await collectionService.restoreCollection(collectionId);
      refresh(); // Reload to reflect changes
    } catch (err) {
      console.error("Failed to restore collection:", err);
      alert(`Failed to restore collection: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  // Get icon for collection type (API documentation specifies 'folder' and 'album')
  const getCollectionIcon = (collection) => {
    switch (collection.collection_type) {
      case "album":
        return "🖼️"; // Album for photo/media collections
      case "folder":
        return "📁"; // Folder for organizing files
      default:
        return "📄"; // Default for unknown types
    }
  };

  // Determine collection ownership
  const isOwnedCollection = (collection) => {
    if (!collection || !collection.id) return false;
    return ownedCollections.some(
      (owned) => owned && owned.id === collection.id,
    );
  };

  // Get ownership badge
  const getOwnershipBadge = (collection) => {
    return isOwnedCollection(collection) ? (
      <span style={{ color: "blue", fontSize: "11px" }}>👤 Owned</span>
    ) : (
      <span style={{ color: "purple", fontSize: "11px" }}>🤝 Shared</span>
    );
  };

  // Handle page changes
  const goToPage = (page) => {
    setCurrentPage(Math.max(1, Math.min(page, totalPages)));
  };

  if (!isAuthenticated) {
    return (
      <div>
        <h2>Collections</h2>
        <p>Please log in to view your collections.</p>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px" }}>
      <div>
        <h1>My Collections</h1>

        {/* Controls - strictly following API capabilities */}
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            border: "1px solid #ddd",
            borderRadius: "5px",
          }}
        >
          <div
            style={{
              display: "flex",
              gap: "15px",
              alignItems: "center",
              flexWrap: "wrap",
              marginBottom: "15px",
            }}
          >
            {/* View mode toggle */}
            <div>
              <label>
                <strong>View: </strong>
              </label>
              <select
                value={viewMode}
                onChange={(e) => setViewMode(e.target.value)}
              >
                <option value="grid">📋 Grid</option>
                <option value="list">📄 List</option>
              </select>
            </div>

            {/* Collection source - API: GET /collections/filtered parameters */}
            <div>
              <label>
                <strong>Include: </strong>
              </label>
              <label style={{ marginLeft: "5px" }}>
                <input
                  type="checkbox"
                  checked={includeOwned}
                  onChange={(e) => setIncludeOwned(e.target.checked)}
                />
                Owned Collections
              </label>
              <label style={{ marginLeft: "10px" }}>
                <input
                  type="checkbox"
                  checked={includeShared}
                  onChange={(e) => setIncludeShared(e.target.checked)}
                />
                Shared Collections
              </label>
            </div>

            {/* Collection type filter - API supports 'folder' and 'album' */}
            <div>
              <label>
                <strong>Type: </strong>
              </label>
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
              >
                <option value="all">All Types</option>
                <option value="folder">📁 Folders</option>
                <option value="album">🖼️ Albums</option>
              </select>
            </div>

            {/* Page size */}
            <div>
              <label>
                <strong>Per page: </strong>
              </label>
              <select
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value));
                  setCurrentPage(1);
                }}
              >
                <option value="10">10</option>
                <option value="25">25</option>
                <option value="50">50</option>
                <option value="100">100</option>
              </select>
            </div>
          </div>

          <div
            style={{
              display: "flex",
              gap: "15px",
              alignItems: "center",
              flexWrap: "wrap",
            }}
          >
            {/* Client-side search (API doesn't provide search capability) */}
            <div>
              <label>
                <strong>Search: </strong>
              </label>
              <input
                type="text"
                placeholder="Search collection names..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                style={{ width: "200px", padding: "5px" }}
              />
              <small style={{ marginLeft: "5px", color: "#666" }}>
                (client-side filter)
              </small>
            </div>

            {/* Refresh button */}
            <button onClick={refresh} disabled={loading}>
              {loading ? "⏳ Loading..." : "🔄 Refresh"}
            </button>
          </div>
        </div>

        {/* Error display */}
        {error && (
          <div
            style={{
              backgroundColor: "#fee",
              border: "1px solid #fcc",
              padding: "15px",
              marginBottom: "20px",
              borderRadius: "5px",
            }}
          >
            <strong>❌ Error:</strong> {error}
            <button onClick={refresh} style={{ marginLeft: "10px" }}>
              🔄 Retry
            </button>
          </div>
        )}

        {/* Collection count and pagination info */}
        {filteredCollections.length > 0 && (
          <div
            style={{ marginBottom: "15px", fontSize: "14px", color: "#666" }}
          >
            Showing {startIndex + 1}-
            {Math.min(endIndex, filteredCollections.length)} of{" "}
            {filteredCollections.length} collections
            {totalPages > 1 && (
              <span>
                {" "}
                • Page {currentPage} of {totalPages}
              </span>
            )}
            {searchTerm && <span> • Filtered by: "{searchTerm}"</span>}
          </div>
        )}

        {/* Collections display */}
        {viewMode === "grid" ? (
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(250px, 1fr))",
              gap: "20px",
              marginBottom: "20px",
            }}
          >
            {paginatedCollections
              .map((collection, index) => {
                // Safety check for collection data
                if (!collection || !collection.id) {
                  console.warn(
                    "Skipping invalid collection at index",
                    index,
                    ":",
                    collection,
                  );
                  return null;
                }

                const collectionId = collection.id;

                return (
                  <div
                    key={collectionId}
                    style={{
                      border: "1px solid #ddd",
                      padding: "15px",
                      borderRadius: "8px",
                      backgroundColor: "#fafafa",
                    }}
                  >
                    <div
                      style={{
                        fontSize: "48px",
                        textAlign: "center",
                        marginBottom: "10px",
                      }}
                    >
                      {getCollectionIcon(collection)}
                    </div>
                    <div
                      style={{
                        fontSize: "16px",
                        fontWeight: "bold",
                        marginBottom: "8px",
                        textAlign: "center",
                      }}
                    >
                      {collection.encrypted_name || "Untitled Collection"}
                    </div>
                    <div style={{ fontSize: "12px", marginBottom: "5px" }}>
                      <strong>Type:</strong>{" "}
                      {collection.collection_type || "unknown"}
                    </div>
                    <div style={{ fontSize: "12px", marginBottom: "5px" }}>
                      {getOwnershipBadge(collection)}
                    </div>
                    {collection.created_at && (
                      <div
                        style={{
                          fontSize: "11px",
                          color: "#666",
                          marginBottom: "5px",
                        }}
                      >
                        <strong>Created:</strong>{" "}
                        {new Date(collection.created_at).toLocaleDateString()}
                      </div>
                    )}
                    {collection.modified_at && (
                      <div
                        style={{
                          fontSize: "11px",
                          color: "#666",
                          marginBottom: "10px",
                        }}
                      >
                        <strong>Modified:</strong>{" "}
                        {new Date(collection.modified_at).toLocaleDateString()}
                      </div>
                    )}
                    {collection.members && collection.members.length > 0 && (
                      <div
                        style={{
                          fontSize: "11px",
                          color: "#666",
                          marginBottom: "10px",
                        }}
                      >
                        👥 {collection.members.length} member
                        {collection.members.length !== 1 ? "s" : ""}
                      </div>
                    )}
                    {collection.parent_id && (
                      <div
                        style={{
                          fontSize: "11px",
                          color: "#666",
                          marginBottom: "10px",
                        }}
                      >
                        📂 Has parent folder
                      </div>
                    )}

                    {/* Actions - API-defined endpoints */}
                    <div
                      style={{
                        display: "flex",
                        gap: "5px",
                        flexWrap: "wrap",
                        justifyContent: "center",
                      }}
                    >
                      <button
                        onClick={() => handleArchive(collectionId)}
                        style={{ fontSize: "11px", padding: "4px 8px" }}
                        title="Archive collection (make read-only)"
                        disabled={loading}
                      >
                        📦 Archive
                      </button>
                      <button
                        onClick={() => handleRestore(collectionId)}
                        style={{ fontSize: "11px", padding: "4px 8px" }}
                        title="Restore archived collection to active state"
                        disabled={loading}
                      >
                        🔄 Restore
                      </button>
                      <button
                        onClick={() => handleDelete(collectionId)}
                        style={{
                          fontSize: "11px",
                          padding: "4px 8px",
                          color: "red",
                        }}
                        title="Soft delete (recoverable for 30 days)"
                        disabled={loading}
                      >
                        🗑️ Delete
                      </button>
                    </div>
                  </div>
                );
              })
              .filter(Boolean)}
          </div>
        ) : (
          <div style={{ marginBottom: "20px" }}>
            <table
              style={{
                width: "100%",
                borderCollapse: "collapse",
                border: "1px solid #ddd",
              }}
            >
              <thead>
                <tr
                  style={{
                    backgroundColor: "#f5f5f5",
                    borderBottom: "2px solid #ddd",
                  }}
                >
                  <th
                    style={{ textAlign: "left", padding: "12px", width: "30%" }}
                  >
                    Name
                  </th>
                  <th
                    style={{ textAlign: "left", padding: "12px", width: "10%" }}
                  >
                    Type
                  </th>
                  <th
                    style={{ textAlign: "left", padding: "12px", width: "10%" }}
                  >
                    Ownership
                  </th>
                  <th
                    style={{ textAlign: "left", padding: "12px", width: "8%" }}
                  >
                    Members
                  </th>
                  <th
                    style={{ textAlign: "left", padding: "12px", width: "12%" }}
                  >
                    Created
                  </th>
                  <th
                    style={{ textAlign: "left", padding: "12px", width: "12%" }}
                  >
                    Modified
                  </th>
                  <th
                    style={{
                      textAlign: "center",
                      padding: "12px",
                      width: "18%",
                    }}
                  >
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {paginatedCollections
                  .map((collection, index) => {
                    // Safety check for collection data
                    if (!collection || !collection.id) {
                      console.warn(
                        "Skipping invalid collection in table at index",
                        index,
                        ":",
                        collection,
                      );
                      return null;
                    }

                    const collectionId = collection.id;

                    return (
                      <tr
                        key={collectionId}
                        style={{ borderBottom: "1px solid #eee" }}
                      >
                        <td
                          style={{ padding: "12px", verticalAlign: "middle" }}
                        >
                          <div
                            style={{ display: "flex", alignItems: "center" }}
                          >
                            <span
                              style={{ marginRight: "10px", fontSize: "20px" }}
                            >
                              {getCollectionIcon(collection)}
                            </span>
                            <div>
                              <div style={{ fontWeight: "bold" }}>
                                {collection.encrypted_name ||
                                  "Untitled Collection"}
                              </div>
                              {collection.parent_id && (
                                <div
                                  style={{ fontSize: "10px", color: "#999" }}
                                >
                                  📂 Nested collection
                                </div>
                              )}
                            </div>
                          </div>
                        </td>
                        <td
                          style={{ padding: "12px", verticalAlign: "middle" }}
                        >
                          {collection.collection_type || "unknown"}
                        </td>
                        <td
                          style={{ padding: "12px", verticalAlign: "middle" }}
                        >
                          {getOwnershipBadge(collection)}
                        </td>
                        <td
                          style={{ padding: "12px", verticalAlign: "middle" }}
                        >
                          {collection.members ? collection.members.length : 0}
                        </td>
                        <td
                          style={{ padding: "12px", verticalAlign: "middle" }}
                        >
                          {collection.created_at
                            ? new Date(
                                collection.created_at,
                              ).toLocaleDateString()
                            : "—"}
                        </td>
                        <td
                          style={{ padding: "12px", verticalAlign: "middle" }}
                        >
                          {collection.modified_at
                            ? new Date(
                                collection.modified_at,
                              ).toLocaleDateString()
                            : "—"}
                        </td>
                        <td
                          style={{
                            padding: "12px",
                            textAlign: "center",
                            verticalAlign: "middle",
                          }}
                        >
                          <div
                            style={{
                              display: "flex",
                              gap: "5px",
                              justifyContent: "center",
                            }}
                          >
                            <button
                              onClick={() => handleArchive(collectionId)}
                              style={{ fontSize: "11px", padding: "4px 6px" }}
                              title="Archive"
                              disabled={loading}
                            >
                              📦
                            </button>
                            <button
                              onClick={() => handleRestore(collectionId)}
                              style={{ fontSize: "11px", padding: "4px 6px" }}
                              title="Restore"
                              disabled={loading}
                            >
                              🔄
                            </button>
                            <button
                              onClick={() => handleDelete(collectionId)}
                              style={{
                                fontSize: "11px",
                                padding: "4px 6px",
                                color: "red",
                              }}
                              title="Delete"
                              disabled={loading}
                            >
                              🗑️
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })
                  .filter(Boolean)}
              </tbody>
            </table>
          </div>
        )}

        {/* Loading indicator */}
        {loading && (
          <div style={{ textAlign: "center", padding: "40px" }}>
            <div style={{ fontSize: "24px", marginBottom: "10px" }}>⏳</div>
            <div>Loading collections...</div>
          </div>
        )}

        {/* Empty state */}
        {!loading && filteredCollections.length === 0 && (
          <div style={{ textAlign: "center", padding: "60px" }}>
            <div style={{ fontSize: "64px", marginBottom: "20px" }}>📁</div>
            <h3>No collections found</h3>
            {!includeOwned && !includeShared ? (
              <p>
                Please select at least one collection source (Owned or Shared
                Collections).
              </p>
            ) : searchTerm || typeFilter !== "all" ? (
              <p>No collections match your current filters.</p>
            ) : (
              <p>You don't have any collections yet.</p>
            )}
            {(searchTerm || typeFilter !== "all") && (
              <button
                onClick={() => {
                  setSearchTerm("");
                  setTypeFilter("all");
                  setCurrentPage(1);
                }}
              >
                Clear filters
              </button>
            )}
          </div>
        )}

        {/* Client-side pagination controls */}
        {!loading && totalPages > 1 && (
          <div
            style={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              gap: "10px",
              padding: "20px",
              borderTop: "1px solid #eee",
            }}
          >
            <button
              onClick={() => goToPage(1)}
              disabled={currentPage === 1}
              title="First page"
            >
              ⏮️
            </button>
            <button
              onClick={() => goToPage(currentPage - 1)}
              disabled={currentPage === 1}
              title="Previous page"
            >
              ⬅️ Prev
            </button>

            <div style={{ display: "flex", gap: "5px", alignItems: "center" }}>
              {/* Show page numbers */}
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum;
                if (totalPages <= 5) {
                  pageNum = i + 1;
                } else if (currentPage <= 3) {
                  pageNum = i + 1;
                } else if (currentPage >= totalPages - 2) {
                  pageNum = totalPages - 4 + i;
                } else {
                  pageNum = currentPage - 2 + i;
                }

                return (
                  <button
                    key={pageNum}
                    onClick={() => goToPage(pageNum)}
                    style={{
                      backgroundColor:
                        currentPage === pageNum ? "#007bff" : "white",
                      color: currentPage === pageNum ? "white" : "black",
                      border: "1px solid #ddd",
                      padding: "8px 12px",
                      fontSize: "14px",
                    }}
                  >
                    {pageNum}
                  </button>
                );
              })}
            </div>

            <button
              onClick={() => goToPage(currentPage + 1)}
              disabled={currentPage === totalPages}
              title="Next page"
            >
              Next ➡️
            </button>
            <button
              onClick={() => goToPage(totalPages)}
              disabled={currentPage === totalPages}
              title="Last page"
            >
              ⏭️
            </button>

            <div
              style={{ marginLeft: "15px", fontSize: "14px", color: "#666" }}
            >
              Page {currentPage} of {totalPages}
            </div>
          </div>
        )}

        {/* API Compliance Footer */}
        <div
          style={{
            marginTop: "20px",
            padding: "15px",
            backgroundColor: "#f8f9fa",
            borderRadius: "5px",
            border: "1px solid #e9ecef",
          }}
        >
          <div style={{ fontSize: "12px", color: "#666" }}>
            <strong>📡 API Endpoints Used:</strong>
            <ul style={{ margin: "5px 0", paddingLeft: "20px" }}>
              <li>
                <strong>GET /collections/filtered</strong> - Main data source
                with include_owned/include_shared parameters
              </li>
              <li>
                <strong>DELETE /collections/{"{collection_id}"}</strong> - Soft
                delete (30-day recovery)
              </li>
              <li>
                <strong>POST /collections/{"{collection_id}"}/archive</strong> -
                Make collection read-only
              </li>
              <li>
                <strong>POST /collections/{"{collection_id}"}/restore</strong> -
                Restore archived collection to active state
              </li>
            </ul>
            <div style={{ marginTop: "10px" }}>
              <strong>Data Available:</strong> id, owner_id, encrypted_name,
              collection_type, parent_id, ancestor_ids, created_at, modified_at,
              members
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CollectionsList;
