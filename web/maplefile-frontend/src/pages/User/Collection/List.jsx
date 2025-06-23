import React, { useState, useEffect, useCallback } from "react";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";

const CollectionsList = () => {
  const { collectionService } = useServices();
  const { isAuthenticated } = useAuth();

  // State management
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [viewMode, setViewMode] = useState("grid"); // 'grid' or 'list'

  // Server-side pagination using sync endpoint (supports cursor + limit)
  const [cursor, setCursor] = useState(null);
  const [hasMore, setHasMore] = useState(true);
  const [limit, setLimit] = useState(25);
  const [totalLoaded, setTotalLoaded] = useState(0);

  // Filter state - only what's available in API
  const [includeOwned, setIncludeOwned] = useState(true);
  const [includeShared, setIncludeShared] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");

  // Load collections using appropriate API endpoint
  const loadCollections = useCallback(
    async (resetCursor = true) => {
      if (!isAuthenticated) return;

      try {
        setLoading(true);
        setError(null);

        if (resetCursor) {
          setCursor(null);
          setCollections([]);
          setTotalLoaded(0);
        }

        // Use getFilteredCollections for rich data, then sync for pagination if needed
        const response = await collectionService.getFilteredCollections(
          includeOwned,
          includeShared,
        );

        // Combine owned and shared collections
        const allCollections = [
          ...(response.owned_collections || []),
          ...(response.shared_collections || []),
        ];

        setCollections(allCollections);
        setTotalLoaded(allCollections.length);
        setHasMore(false); // getFilteredCollections returns all data at once
      } catch (err) {
        console.error("Failed to load collections:", err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    },
    [collectionService, isAuthenticated, includeOwned, includeShared],
  );

  // Alternative: Load using sync endpoint for true server-side pagination
  const loadCollectionsWithSync = useCallback(
    async (resetCursor = true) => {
      if (!isAuthenticated) return;

      try {
        setLoading(true);
        setError(null);

        const currentCursor = resetCursor ? null : cursor;

        if (resetCursor) {
          setCollections([]);
          setTotalLoaded(0);
        }

        // Use sync endpoint for server-side pagination
        const response = await collectionService.syncCollections(
          currentCursor,
          limit,
        );

        if (resetCursor) {
          setCollections(response.collections || []);
        } else {
          setCollections((prev) => [...prev, ...(response.collections || [])]);
        }

        setCursor(response.next_cursor);
        setHasMore(response.has_more || false);
        setTotalLoaded((prev) =>
          resetCursor
            ? response.collections?.length || 0
            : prev + (response.collections?.length || 0),
        );
      } catch (err) {
        console.error("Failed to load collections with sync:", err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    },
    [collectionService, isAuthenticated, cursor, limit],
  );

  // Refresh collections
  const refresh = useCallback(() => {
    loadCollections(true);
  }, [loadCollections]);

  // Initial load and when filters change
  useEffect(() => {
    if (isAuthenticated) {
      refresh();
    }
  }, [isAuthenticated, includeOwned, includeShared, refresh]);

  // Client-side filtering (since API doesn't support all filter types)
  const filteredCollections = collections.filter((collection) => {
    // Search filter - only if we have the data
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

  // Collection actions - strictly following API documentation
  const handleDelete = async (collectionId) => {
    if (
      !window.confirm(
        "Are you sure you want to delete this collection? This action will soft-delete the collection and make it recoverable for 30 days.",
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

  // Get icon for collection type (API supports 'folder' and 'album')
  const getCollectionIcon = (collection) => {
    switch (collection.collection_type) {
      case "album":
        return "🖼️"; // Album/photo collection
      case "folder":
        return "📁"; // Folder collection
      default:
        return "📄"; // Unknown type
    }
  };

  // Get status badge - only if state is available (from sync endpoint)
  const getStatusBadge = (collection) => {
    if (!collection.state) {
      return <span style={{ color: "gray", fontSize: "12px" }}>● Unknown</span>;
    }

    switch (collection.state) {
      case "active":
        return (
          <span style={{ color: "green", fontSize: "12px" }}>● Active</span>
        );
      case "archived":
        return (
          <span style={{ color: "orange", fontSize: "12px" }}>● Archived</span>
        );
      case "deleted":
        return (
          <span style={{ color: "red", fontSize: "12px" }}>● Deleted</span>
        );
      default:
        return (
          <span style={{ color: "gray", fontSize: "12px" }}>
            ● {collection.state}
          </span>
        );
    }
  };

  // Determine ownership (collections from owned vs shared arrays)
  const getOwnershipInfo = (collection) => {
    // Since we're combining owned and shared collections, we need to track this
    // This is an approximation - ideally we'd track this when loading
    return (
      <span style={{ color: "blue", fontSize: "11px" }}>👤 Collection</span>
    );
  };

  // Load more collections (for sync endpoint pagination)
  const loadMore = () => {
    if (!loading && hasMore) {
      loadCollectionsWithSync(false);
    }
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

        {/* Controls - only API-supported options */}
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

            {/* Collection source - matches API getFilteredCollections parameters */}
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

            {/* Page size for sync endpoint */}
            <div>
              <label>
                <strong>Load size: </strong>
              </label>
              <select
                value={limit}
                onChange={(e) => {
                  setLimit(Number(e.target.value));
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
            {/* Client-side search (since API doesn't support search) */}
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
                (client-side)
              </small>
            </div>

            {/* Refresh button */}
            <button onClick={refresh} disabled={loading}>
              {loading ? "⏳ Loading..." : "🔄 Refresh"}
            </button>

            {/* Switch to sync pagination */}
            <button
              onClick={() => loadCollectionsWithSync(true)}
              disabled={loading}
              title="Use server-side pagination with minimal data"
            >
              📡 Use Sync API
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

        {/* Collection count info */}
        {filteredCollections.length > 0 && (
          <div
            style={{ marginBottom: "15px", fontSize: "14px", color: "#666" }}
          >
            Showing {filteredCollections.length} collections
            {searchTerm && <span> (filtered by search)</span>}
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
            {filteredCollections.map((collection) => (
              <div
                key={collection.id}
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
                {collection.state && (
                  <div style={{ fontSize: "12px", marginBottom: "5px" }}>
                    {getStatusBadge(collection)}
                  </div>
                )}
                <div style={{ fontSize: "12px", marginBottom: "5px" }}>
                  {getOwnershipInfo(collection)}
                </div>
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
                {collection.created_at && (
                  <div
                    style={{
                      fontSize: "11px",
                      color: "#666",
                      marginBottom: "10px",
                    }}
                  >
                    <strong>Created:</strong>{" "}
                    {new Date(collection.created_at).toLocaleDateString()}
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

                {/* Actions - only available API operations */}
                <div
                  style={{
                    display: "flex",
                    gap: "5px",
                    flexWrap: "wrap",
                    justifyContent: "center",
                  }}
                >
                  <button
                    onClick={() => handleArchive(collection.id)}
                    style={{ fontSize: "11px", padding: "4px 8px" }}
                    title="Archive collection (make read-only)"
                    disabled={loading}
                  >
                    📦 Archive
                  </button>
                  <button
                    onClick={() => handleRestore(collection.id)}
                    style={{ fontSize: "11px", padding: "4px 8px" }}
                    title="Restore archived collection"
                    disabled={loading}
                  >
                    🔄 Restore
                  </button>
                  <button
                    onClick={() => handleDelete(collection.id)}
                    style={{
                      fontSize: "11px",
                      padding: "4px 8px",
                      color: "red",
                    }}
                    title="Soft delete collection (30-day recovery)"
                    disabled={loading}
                  >
                    🗑️ Delete
                  </button>
                </div>
              </div>
            ))}
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
                    style={{ textAlign: "left", padding: "12px", width: "35%" }}
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
                    Status
                  </th>
                  <th
                    style={{ textAlign: "left", padding: "12px", width: "10%" }}
                  >
                    Members
                  </th>
                  <th
                    style={{ textAlign: "left", padding: "12px", width: "15%" }}
                  >
                    Modified
                  </th>
                  <th
                    style={{
                      textAlign: "center",
                      padding: "12px",
                      width: "20%",
                    }}
                  >
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {filteredCollections.map((collection) => (
                  <tr
                    key={collection.id}
                    style={{ borderBottom: "1px solid #eee" }}
                  >
                    <td style={{ padding: "12px", verticalAlign: "middle" }}>
                      <div style={{ display: "flex", alignItems: "center" }}>
                        <span style={{ marginRight: "10px", fontSize: "20px" }}>
                          {getCollectionIcon(collection)}
                        </span>
                        <div>
                          <div style={{ fontWeight: "bold" }}>
                            {collection.encrypted_name || "Untitled Collection"}
                          </div>
                          {collection.id && (
                            <div
                              style={{
                                fontSize: "10px",
                                color: "#999",
                                fontFamily: "monospace",
                              }}
                            >
                              ID: {collection.id.substring(0, 8)}...
                            </div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td style={{ padding: "12px", verticalAlign: "middle" }}>
                      {collection.collection_type || "unknown"}
                    </td>
                    <td style={{ padding: "12px", verticalAlign: "middle" }}>
                      {collection.state ? getStatusBadge(collection) : "—"}
                    </td>
                    <td style={{ padding: "12px", verticalAlign: "middle" }}>
                      {collection.members ? collection.members.length : "—"}
                    </td>
                    <td style={{ padding: "12px", verticalAlign: "middle" }}>
                      {collection.modified_at
                        ? new Date(collection.modified_at).toLocaleDateString()
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
                          onClick={() => handleArchive(collection.id)}
                          style={{ fontSize: "11px", padding: "4px 6px" }}
                          title="Archive"
                          disabled={loading}
                        >
                          📦
                        </button>
                        <button
                          onClick={() => handleRestore(collection.id)}
                          style={{ fontSize: "11px", padding: "4px 6px" }}
                          title="Restore"
                          disabled={loading}
                        >
                          🔄
                        </button>
                        <button
                          onClick={() => handleDelete(collection.id)}
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
                ))}
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
            {searchTerm ? (
              <p>No collections match your search criteria.</p>
            ) : !includeOwned && !includeShared ? (
              <p>
                Please select at least one collection source (Owned or Shared).
              </p>
            ) : (
              <p>You don't have any collections yet.</p>
            )}
            {searchTerm && (
              <button onClick={() => setSearchTerm("")}>Clear search</button>
            )}
          </div>
        )}

        {/* Server-side pagination controls (for sync endpoint) */}
        {hasMore && (
          <div
            style={{
              textAlign: "center",
              padding: "20px",
              borderTop: "1px solid #eee",
            }}
          >
            <button
              onClick={loadMore}
              disabled={loading}
              style={{ padding: "10px 20px", fontSize: "14px" }}
            >
              {loading ? "⏳ Loading..." : "📥 Load More Collections"}
            </button>
            <div style={{ marginTop: "10px", fontSize: "12px", color: "#666" }}>
              {totalLoaded} collections loaded so far
            </div>
          </div>
        )}

        {/* API Information Footer */}
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
            <strong>📡 API Usage:</strong>
            <ul style={{ margin: "5px 0", paddingLeft: "20px" }}>
              <li>
                <strong>Data Source:</strong> GET /collections/filtered
                (include_owned, include_shared)
              </li>
              <li>
                <strong>Actions:</strong> DELETE /collections/:id (soft delete),
                POST /collections/:id/archive, POST /collections/:id/restore
              </li>
              <li>
                <strong>Collection Types:</strong> folder, album (per API
                specification)
              </li>
              <li>
                <strong>Pagination:</strong> Client-side filtering + Server-side
                available via GET /sync/collections
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CollectionsList;
