// SyncCollectionStorageExample.jsx - UPDATED
// Example page to test SyncCollectionStorageService and SyncCollectionAPIService

import React, { useState, useEffect } from "react";
import { useServices } from "../../../hooks/useService.jsx";

const SyncCollectionStorageExample = () => {
  const { syncCollectionStorageService, syncCollectionAPIService } =
    useServices();
  const [syncCollections, setSyncCollections] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [storageInfo, setStorageInfo] = useState({});

  // Load storage info on component mount
  useEffect(() => {
    updateStorageInfo();
  }, []);

  // Update storage information
  const updateStorageInfo = () => {
    const info = syncCollectionStorageService.getStorageInfo();
    setStorageInfo(info);
    console.log("[SyncCollectionStorageExample] Storage info updated:", info);
  };

  // Load sync collections from localStorage
  const handleLoadFromStorage = () => {
    try {
      setError(null);
      console.log("📂 Loading sync collections from localStorage...");

      const storedSyncCollections =
        syncCollectionStorageService.getSyncCollections();
      setSyncCollections(storedSyncCollections);
      updateStorageInfo();

      console.log(
        "✅ Loaded from localStorage:",
        storedSyncCollections.length,
        "sync collections",
      );
    } catch (err) {
      setError(err.message);
      console.error("❌ Failed to load from localStorage:", err);
    }
  };

  // Save current sync collections to localStorage
  const handleSaveToStorage = () => {
    try {
      setError(null);
      console.log("💾 Saving sync collections to localStorage...");

      const success =
        syncCollectionStorageService.saveSyncCollections(syncCollections);
      if (success) {
        updateStorageInfo();
        console.log(
          "✅ Saved to localStorage:",
          syncCollections.length,
          "sync collections",
        );
      } else {
        throw new Error("Failed to save sync collections");
      }
    } catch (err) {
      setError(err.message);
      console.error("❌ Failed to save to localStorage:", err);
    }
  };

  // Sync collections from API and save to localStorage
  const handleSyncAndSave = async () => {
    setLoading(true);
    setError(null);

    try {
      console.log("🔄 Syncing collections from API...");

      // First, sync all collections from the API
      const syncedCollections =
        await syncCollectionAPIService.syncAllCollections();
      setSyncCollections(syncedCollections);

      // Then, save them to localStorage
      const success =
        syncCollectionStorageService.saveSyncCollections(syncedCollections);
      if (success) {
        updateStorageInfo();
        console.log(
          "✅ Synced from API and saved to localStorage:",
          syncedCollections.length,
          "sync collections",
        );
      } else {
        throw new Error("Failed to save synced collections");
      }
    } catch (err) {
      setError(err.message);
      console.error("❌ Sync and save failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Clear all stored sync collections
  const handleClearStorage = () => {
    try {
      setError(null);
      console.log("🗑️ Clearing stored sync collections...");

      const success = syncCollectionStorageService.clearSyncCollections();
      if (success) {
        setSyncCollections([]);
        updateStorageInfo();
        console.log("✅ Storage cleared");
      } else {
        throw new Error("Failed to clear storage");
      }
    } catch (err) {
      setError(err.message);
      console.error("❌ Failed to clear storage:", err);
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>💾 Sync Collection Storage Service Test</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates both the{" "}
        <strong>SyncCollectionAPIService</strong> (for API calls) and the{" "}
        <strong>SyncCollectionStorageService</strong> (for localStorage).
      </p>

      {/* Action Buttons */}
      <div
        style={{
          marginBottom: "30px",
          display: "flex",
          gap: "10px",
          flexWrap: "wrap",
        }}
      >
        <button
          onClick={handleLoadFromStorage}
          disabled={loading}
          style={{
            padding: "10px 20px",
            backgroundColor: "#28a745",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: loading ? "not-allowed" : "pointer",
          }}
        >
          📂 Load from Storage
        </button>

        <button
          onClick={handleSaveToStorage}
          disabled={loading || syncCollections.length === 0}
          style={{
            padding: "10px 20px",
            backgroundColor: "#007bff",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor:
              loading || syncCollections.length === 0
                ? "not-allowed"
                : "pointer",
          }}
        >
          💾 Save to Storage
        </button>

        <button
          onClick={handleSyncAndSave}
          disabled={loading}
          style={{
            padding: "10px 20px",
            backgroundColor: "#17a2b8",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: loading ? "not-allowed" : "pointer",
          }}
        >
          {loading ? "🔄 Syncing..." : "🔄 Sync from API & Save"}
        </button>

        <button
          onClick={handleClearStorage}
          disabled={loading}
          style={{
            padding: "10px 20px",
            backgroundColor: "#dc3545",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: loading ? "not-allowed" : "pointer",
          }}
        >
          🗑️ Clear Storage
        </button>
      </div>

      {/* Storage Info */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#f8f9fa",
          borderRadius: "6px",
          border: "1px solid #dee2e6",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>📊 Storage Information:</h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: "10px",
          }}
        >
          <div>
            <strong>Has Stored Sync Collections:</strong>{" "}
            {storageInfo.hasSyncCollections ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Stored Count:</strong>{" "}
            {storageInfo.syncCollectionsCount || 0}
          </div>
          <div>
            <strong>Current Count:</strong> {syncCollections.length}
          </div>
          <div>
            <strong>Last Saved:</strong>{" "}
            {storageInfo.metadata?.savedAt
              ? new Date(storageInfo.metadata.savedAt).toLocaleString()
              : "Never"}
          </div>
        </div>
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
        <h3>📁 Current Sync Collections ({syncCollections.length})</h3>
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
              No sync collections loaded.
            </p>
            <p style={{ color: "#6c757d" }}>
              Click "Load from Storage" to load saved sync collections, or "Sync
              from API & Save" to fetch from API.
            </p>
          </div>
        ) : (
          <div
            style={{
              display: "grid",
              gap: "10px",
              maxHeight: "400px",
              overflow: "auto",
              border: "1px solid #dee2e6",
              borderRadius: "6px",
              padding: "10px",
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

      {/* Debug Info */}
      <details style={{ marginTop: "20px" }}>
        <summary style={{ cursor: "pointer", fontWeight: "bold" }}>
          🔧 Debug Info
        </summary>
        <pre
          style={{
            backgroundColor: "#f8f9fa",
            padding: "10px",
            borderRadius: "4px",
            fontSize: "12px",
            overflow: "auto",
            marginTop: "10px",
          }}
        >
          {JSON.stringify(storageInfo, null, 2)}
        </pre>
      </details>
    </div>
  );
};

export default SyncCollectionStorageExample;
