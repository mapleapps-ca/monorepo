// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/SyncCollectionAPIExample.jsx
// Example component demonstrating how to use the SyncCollectionAPIService

import React, { useState } from "react";
import { useServices } from "../../../hooks/useService.jsx";

const SyncCollectionAPIExample = () => {
  const { syncCollectionAPIService } = useServices();
  const [syncCollections, setSyncCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // The ONLY function you need - sync all collections from API
  const handleSyncAll = async () => {
    setLoading(true);
    setError(null);
    setSyncCollections([]);

    try {
      console.log("🔄 Starting complete API sync...");

      // This ONE function call gets ALL sync collections from API automatically
      const allSyncCollections =
        await syncCollectionAPIService.syncAllCollections();

      setSyncCollections(allSyncCollections);
      console.log(
        "✅ API sync complete:",
        allSyncCollections.length,
        "sync collections",
      );
    } catch (err) {
      setError(err.message);
      console.error("❌ API sync failed:", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>🔄 Sync All Collections from API (Simplified)</h2>

      <div style={{ marginBottom: "30px" }}>
        <button
          onClick={handleSyncAll}
          disabled={loading}
          style={{
            padding: "12px 24px",
            fontSize: "16px",
            backgroundColor: loading ? "#6c757d" : "#007bff",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: loading ? "not-allowed" : "pointer",
            minWidth: "200px",
          }}
        >
          {loading
            ? "🔄 Syncing from API..."
            : "🚀 Sync All Collections from API"}
        </button>
      </div>

      {/* Status Display */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#f8f9fa",
          borderRadius: "6px",
          border: "1px solid #dee2e6",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>Status:</h4>
        <p style={{ margin: "5px 0" }}>
          <strong>Sync Collections Found:</strong> {syncCollections.length}
        </p>
        <p style={{ margin: "5px 0" }}>
          <strong>Status:</strong>{" "}
          {loading ? "🔄 Loading from API..." : "✅ Ready"}
        </p>
      </div>

      {/* Error Display */}
      {error && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#f8d7da",
            borderRadius: "6px",
            color: "#721c24",
            border: "1px solid #f5c6cb",
          }}
        >
          <h4 style={{ margin: "0 0 10px 0" }}>❌ Error:</h4>
          <p style={{ margin: 0 }}>{error}</p>
        </div>
      )}

      {/* Sync Collections Display */}
      <div>
        <h3>📁 Sync Collections from API ({syncCollections.length})</h3>
        {syncCollections.length === 0 ? (
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
              No sync collections loaded yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Click "Sync All Collections from API" to load all your sync
              collections.
            </p>
          </div>
        ) : (
          <div
            style={{
              display: "grid",
              gap: "10px",
              maxHeight: "400px",
              overflow: "auto",
            }}
          >
            {syncCollections.map((syncCollection, index) => (
              <div
                key={`${syncCollection.id}-${index}`}
                style={{
                  padding: "12px",
                  border: "1px solid #dee2e6",
                  borderRadius: "6px",
                  backgroundColor:
                    syncCollection.state === "active" ? "#d1ecf1" : "#fff3cd",
                  borderLeft: `4px solid ${syncCollection.state === "active" ? "#0c5460" : "#856404"}`,
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                  }}
                >
                  <div>
                    <p style={{ margin: "0 0 5px 0", fontWeight: "bold" }}>
                      ID: {syncCollection.id}
                    </p>
                    <p style={{ margin: "0", fontSize: "14px", color: "#666" }}>
                      State: {syncCollection.state} | Version:{" "}
                      {syncCollection.version}
                    </p>
                  </div>
                  <div
                    style={{
                      fontSize: "12px",
                      color: "#999",
                      textAlign: "right",
                    }}
                  >
                    Modified:{" "}
                    {new Date(syncCollection.modified_at).toLocaleDateString()}
                  </div>
                </div>

                {syncCollection.parent_id && (
                  <p
                    style={{
                      margin: "5px 0 0 0",
                      fontSize: "12px",
                      color: "#666",
                    }}
                  >
                    Parent: {syncCollection.parent_id}
                  </p>
                )}

                {syncCollection.tombstone_version > 0 && (
                  <p
                    style={{
                      margin: "5px 0 0 0",
                      fontSize: "12px",
                      color: "#d63384",
                    }}
                  >
                    🪦 Tombstone Version: {syncCollection.tombstone_version}
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

export default SyncCollectionAPIExample;
