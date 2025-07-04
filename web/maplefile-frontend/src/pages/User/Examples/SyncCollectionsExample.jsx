// SyncCollectionsExample.jsx - SIMPLIFIED VERSION
// Just one button that syncs all collections

import React, { useState } from "react";
import { useServices } from "../../../hooks/useService.jsx";

const SyncCollectionsExample = () => {
  const { syncCollectionsService } = useServices();
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // The ONLY function you need - sync all collections
  const handleSyncAll = async () => {
    setLoading(true);
    setError(null);
    setCollections([]);

    try {
      console.log("🔄 Starting complete sync...");

      // This ONE function call gets ALL collections automatically
      const allCollections = await syncCollectionsService.syncAllCollections();

      setCollections(allCollections);
      console.log("✅ Sync complete:", allCollections.length, "collections");
    } catch (err) {
      setError(err.message);
      console.error("❌ Sync failed:", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>🔄 Sync All Collections (Simplified)</h2>

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
          {loading ? "🔄 Syncing..." : "🚀 Sync All Collections"}
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
          <strong>Collections Found:</strong> {collections.length}
        </p>
        <p style={{ margin: "5px 0" }}>
          <strong>Status:</strong> {loading ? "🔄 Loading..." : "✅ Ready"}
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

      {/* Collections Display */}
      <div>
        <h3>📁 Collections ({collections.length})</h3>
        {collections.length === 0 ? (
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
              No collections loaded yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Click "Sync All Collections" to load all your collections.
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
            {collections.map((collection, index) => (
              <div
                key={`${collection.id}-${index}`}
                style={{
                  padding: "12px",
                  border: "1px solid #dee2e6",
                  borderRadius: "6px",
                  backgroundColor:
                    collection.state === "active" ? "#d1ecf1" : "#fff3cd",
                  borderLeft: `4px solid ${collection.state === "active" ? "#0c5460" : "#856404"}`,
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
                      ID: {collection.id}
                    </p>
                    <p style={{ margin: "0", fontSize: "14px", color: "#666" }}>
                      State: {collection.state} | Version: {collection.version}
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
                    {new Date(collection.modified_at).toLocaleDateString()}
                  </div>
                </div>

                {collection.parent_id && (
                  <p
                    style={{
                      margin: "5px 0 0 0",
                      fontSize: "12px",
                      color: "#666",
                    }}
                  >
                    Parent: {collection.parent_id}
                  </p>
                )}

                {collection.tombstone_version > 0 && (
                  <p
                    style={{
                      margin: "5px 0 0 0",
                      fontSize: "12px",
                      color: "#d63384",
                    }}
                  >
                    🪦 Tombstone Version: {collection.tombstone_version}
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
