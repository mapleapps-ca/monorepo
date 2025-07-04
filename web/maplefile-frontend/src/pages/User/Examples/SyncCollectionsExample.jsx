// SyncCollectionsExample.jsx
// Example component demonstrating how to use the SyncCollectionsService

import React, { useState } from "react";
import { useServices } from "../../../hooks/useService.jsx";

const SyncCollectionsExample = () => {
  const { syncCollectionsService } = useServices();
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [nextCursor, setNextCursor] = useState(null);
  const [hasMore, setHasMore] = useState(false);
  const [pageCount, setPageCount] = useState(0);

  // Example 1: Simple sync with default parameters
  const handleSimpleSync = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await syncCollectionsService.syncCollections();
      setCollections(response.collections || []);
      setNextCursor(response.next_cursor);
      setHasMore(response.has_more);
      setPageCount(1);
      console.log("Simple sync completed:", response);
    } catch (err) {
      setError(err.message);
      console.error("Simple sync failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Example 2: Sync with custom limit
  const handleSyncWithLimit = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await syncCollectionsService.syncCollections({
        limit: 10,
      });
      setCollections(response.collections || []);
      setNextCursor(response.next_cursor);
      setHasMore(response.has_more);
      setPageCount(1);
      console.log("Sync with limit completed:", response);
    } catch (err) {
      setError(err.message);
      console.error("Sync with limit failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Example 3: Load next page using cursor
  const handleLoadNextPage = async () => {
    if (!nextCursor) return;

    setLoading(true);
    setError(null);

    try {
      const response = await syncCollectionsService.syncCollections({
        limit: 10,
        cursor: nextCursor,
      });

      // Append new collections to existing ones
      setCollections((prev) => [...prev, ...(response.collections || [])]);
      setNextCursor(response.next_cursor);
      setHasMore(response.has_more);
      setPageCount((prev) => prev + 1);
      console.log("Next page loaded:", response);
    } catch (err) {
      setError(err.message);
      console.error("Load next page failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Example 4: Sync all collections (automatic pagination)
  const handleSyncAll = async () => {
    setLoading(true);
    setError(null);
    setCollections([]);
    setPageCount(0);

    try {
      const allCollections = await syncCollectionsService.syncAllCollections({
        limit: 10, // Small limit for demonstration
        onPageReceived: (pageCollections, pageNum, response) => {
          console.log(
            `Received page ${pageNum} with ${pageCollections.length} collections`,
          );
          setPageCount(pageNum);
          // Optionally update UI with each page as it arrives
          setCollections((prev) => [...prev, ...pageCollections]);
        },
      });

      console.log("All collections synced:", allCollections);
      setHasMore(false);
      setNextCursor(null);
    } catch (err) {
      setError(err.message);
      console.error("Sync all failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Example 5: Sync collections since a specific time
  const handleSyncSince = async () => {
    setLoading(true);
    setError(null);

    try {
      // Sync collections modified in the last 24 hours
      const since = new Date();
      since.setHours(since.getHours() - 24);

      const recentCollections =
        await syncCollectionsService.syncCollectionsSince(since.toISOString(), {
          limit: 10,
        });

      setCollections(recentCollections);
      setHasMore(false);
      setNextCursor(null);
      setPageCount(1);
      console.log("Recent collections synced:", recentCollections);
    } catch (err) {
      setError(err.message);
      console.error("Sync since failed:", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>🔄 Sync Collections Service Examples</h2>

      {/* Action Buttons */}
      <div
        style={{
          marginBottom: "20px",
          display: "flex",
          gap: "10px",
          flexWrap: "wrap",
        }}
      >
        <button
          onClick={handleSimpleSync}
          disabled={loading}
          style={{
            padding: "8px 16px",
            backgroundColor: "#007bff",
            color: "white",
            border: "none",
            borderRadius: "4px",
          }}
        >
          Simple Sync (Default)
        </button>

        <button
          onClick={handleSyncWithLimit}
          disabled={loading}
          style={{
            padding: "8px 16px",
            backgroundColor: "#28a745",
            color: "white",
            border: "none",
            borderRadius: "4px",
          }}
        >
          Sync with Limit (10)
        </button>

        <button
          onClick={handleLoadNextPage}
          disabled={loading || !hasMore}
          style={{
            padding: "8px 16px",
            backgroundColor: "#ffc107",
            color: "black",
            border: "none",
            borderRadius: "4px",
          }}
        >
          Load Next Page
        </button>

        <button
          onClick={handleSyncAll}
          disabled={loading}
          style={{
            padding: "8px 16px",
            backgroundColor: "#dc3545",
            color: "white",
            border: "none",
            borderRadius: "4px",
          }}
        >
          Sync All (Auto-paginate)
        </button>

        <button
          onClick={handleSyncSince}
          disabled={loading}
          style={{
            padding: "8px 16px",
            backgroundColor: "#6c757d",
            color: "white",
            border: "none",
            borderRadius: "4px",
          }}
        >
          Sync Last 24h
        </button>
      </div>

      {/* Status Display */}
      <div
        style={{
          marginBottom: "20px",
          padding: "10px",
          backgroundColor: "#f8f9fa",
          borderRadius: "4px",
        }}
      >
        <p>
          <strong>Status:</strong> {loading ? "Loading..." : "Ready"}
        </p>
        <p>
          <strong>Collections Count:</strong> {collections.length}
        </p>
        <p>
          <strong>Pages Loaded:</strong> {pageCount}
        </p>
        <p>
          <strong>Has More:</strong> {hasMore ? "Yes" : "No"}
        </p>
        <p>
          <strong>Next Cursor:</strong>{" "}
          {nextCursor ? `${nextCursor.substring(0, 20)}...` : "None"}
        </p>
      </div>

      {/* Error Display */}
      {error && (
        <div
          style={{
            marginBottom: "20px",
            padding: "10px",
            backgroundColor: "#f8d7da",
            borderRadius: "4px",
            color: "#721c24",
          }}
        >
          <strong>Error:</strong> {error}
        </div>
      )}

      {/* Collections Display */}
      <div>
        <h3>Collections ({collections.length})</h3>
        {collections.length === 0 ? (
          <p>
            No collections found. Try clicking one of the sync buttons above.
          </p>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {collections.map((collection, index) => (
              <div
                key={`${collection.id}-${index}`}
                style={{
                  padding: "10px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                  backgroundColor:
                    collection.state === "active" ? "#f8f9fa" : "#fff3cd",
                }}
              >
                <p>
                  <strong>ID:</strong> {collection.id}
                </p>
                <p>
                  <strong>State:</strong> {collection.state}
                </p>
                <p>
                  <strong>Version:</strong> {collection.version}
                </p>
                <p>
                  <strong>Modified:</strong>{" "}
                  {new Date(collection.modified_at).toLocaleString()}
                </p>
                {collection.parent_id && (
                  <p>
                    <strong>Parent ID:</strong> {collection.parent_id}
                  </p>
                )}
                {collection.tombstone_version > 0 && (
                  <p>
                    <strong>Tombstone Version:</strong>{" "}
                    {collection.tombstone_version}
                  </p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default SyncCollectionsExample;
