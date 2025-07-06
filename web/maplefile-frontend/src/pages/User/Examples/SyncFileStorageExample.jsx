// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/SyncFileStorageExample.jsx
// Example page to test SyncFileStorageService and SyncFileAPIService

import React, { useState, useEffect } from "react";
import { useServices } from "../../../hooks/useService.jsx";

const SyncFileStorageExample = () => {
  const { syncFileStorageService, syncFileAPIService } = useServices();
  const [syncFiles, setSyncFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [storageInfo, setStorageInfo] = useState({});
  const [selectedCollectionId, setSelectedCollectionId] = useState("");

  // Load storage info on component mount
  useEffect(() => {
    updateStorageInfo();
  }, []);

  // Update storage information
  const updateStorageInfo = () => {
    const info = syncFileStorageService.getStorageInfo();
    setStorageInfo(info);
    console.log("[SyncFileStorageExample] Storage info updated:", info);
  };

  // Load sync files from localStorage
  const handleLoadFromStorage = () => {
    try {
      setError(null);
      console.log("📂 Loading sync files from localStorage...");

      const storedSyncFiles = syncFileStorageService.getSyncFiles();
      setSyncFiles(storedSyncFiles);
      updateStorageInfo();

      console.log(
        "✅ Loaded from localStorage:",
        storedSyncFiles.length,
        "sync files",
      );
    } catch (err) {
      setError(err.message);
      console.error("❌ Failed to load from localStorage:", err);
    }
  };

  // Save current sync files to localStorage
  const handleSaveToStorage = () => {
    try {
      setError(null);
      console.log("💾 Saving sync files to localStorage...");

      const success = syncFileStorageService.saveSyncFiles(syncFiles);
      if (success) {
        updateStorageInfo();
        console.log(
          "✅ Saved to localStorage:",
          syncFiles.length,
          "sync files",
        );
      } else {
        throw new Error("Failed to save sync files");
      }
    } catch (err) {
      setError(err.message);
      console.error("❌ Failed to save to localStorage:", err);
    }
  };

  // Sync files from API and save to localStorage
  const handleSyncAndSave = async () => {
    setLoading(true);
    setError(null);

    try {
      console.log("🔄 Syncing files from API...");

      // First, sync all files from the API
      const syncedFiles = await syncFileAPIService.syncAllFiles();
      setSyncFiles(syncedFiles);

      // Then, save them to localStorage
      const success = syncFileStorageService.saveSyncFiles(syncedFiles);
      if (success) {
        updateStorageInfo();
        console.log(
          "✅ Synced from API and saved to localStorage:",
          syncedFiles.length,
          "sync files",
        );
      } else {
        throw new Error("Failed to save synced files");
      }
    } catch (err) {
      setError(err.message);
      console.error("❌ Sync and save failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Load files by collection
  const handleLoadByCollection = () => {
    if (!selectedCollectionId) {
      setError("Please enter a collection ID");
      return;
    }

    try {
      setError(null);
      console.log("📂 Loading files for collection:", selectedCollectionId);

      const collectionFiles =
        syncFileStorageService.getSyncFilesByCollection(selectedCollectionId);
      setSyncFiles(collectionFiles);

      console.log("✅ Loaded", collectionFiles.length, "files for collection");
    } catch (err) {
      setError(err.message);
      console.error("❌ Failed to load by collection:", err);
    }
  };

  // Clear all stored sync files
  const handleClearStorage = () => {
    try {
      setError(null);
      console.log("🗑️ Clearing stored sync files...");

      const success = syncFileStorageService.clearSyncFiles();
      if (success) {
        setSyncFiles([]);
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

  // Format file size for display
  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
  };

  // Calculate total file size
  const getTotalFileSize = () => {
    return syncFiles.reduce((sum, file) => sum + (file.file_size || 0), 0);
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>💾 Sync File Storage Service Test</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates both the <strong>SyncFileAPIService</strong> (for
        API calls) and the <strong>SyncFileStorageService</strong> (for
        localStorage).
      </p>

      {/* Action Buttons */}
      <div
        style={{
          marginBottom: "30px",
          display: "flex",
          gap: "10px",
          flexWrap: "wrap",
          alignItems: "center",
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
          disabled={loading || syncFiles.length === 0}
          style={{
            padding: "10px 20px",
            backgroundColor: "#007bff",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor:
              loading || syncFiles.length === 0 ? "not-allowed" : "pointer",
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

        <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
          <input
            type="text"
            placeholder="Collection ID"
            value={selectedCollectionId}
            onChange={(e) => setSelectedCollectionId(e.target.value)}
            style={{
              padding: "8px",
              borderRadius: "4px",
              border: "1px solid #ddd",
            }}
          />
          <button
            onClick={handleLoadByCollection}
            disabled={loading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#6f42c1",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: loading ? "not-allowed" : "pointer",
            }}
          >
            📁 Load Collection
          </button>
        </div>
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
            <strong>Has Stored Sync Files:</strong>{" "}
            {storageInfo.hasSyncFiles ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Stored Count:</strong> {storageInfo.syncFilesCount || 0}
          </div>
          <div>
            <strong>Current Count:</strong> {syncFiles.length}
          </div>
          <div>
            <strong>Total Size:</strong> {formatFileSize(getTotalFileSize())}
          </div>
          <div>
            <strong>Last Saved:</strong>{" "}
            {storageInfo.metadata?.savedAt
              ? new Date(storageInfo.metadata.savedAt).toLocaleString()
              : "Never"}
          </div>
          <div>
            <strong>Collections:</strong>{" "}
            {storageInfo.collectionBreakdown
              ? Object.keys(storageInfo.collectionBreakdown).length
              : 0}
          </div>
        </div>

        {/* Collection Breakdown */}
        {storageInfo.collectionBreakdown &&
          Object.keys(storageInfo.collectionBreakdown).length > 0 && (
            <div style={{ marginTop: "15px" }}>
              <h5 style={{ margin: "0 0 10px 0" }}>Collection Breakdown:</h5>
              <div style={{ fontSize: "14px" }}>
                {Object.entries(storageInfo.collectionBreakdown).map(
                  ([collectionId, info]) => (
                    <div key={collectionId} style={{ marginBottom: "5px" }}>
                      <strong>{collectionId}:</strong> {info.count} files,{" "}
                      {formatFileSize(info.totalSize)}
                      {" ("}Active: {info.states.active}, Deleted:{" "}
                      {info.states.deleted}, Archived: {info.states.archived}
                      {")"}
                    </div>
                  ),
                )}
              </div>
            </div>
          )}
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

      {/* Sync Files Display */}
      <div>
        <h3>📁 Current Sync Files ({syncFiles.length})</h3>
        {syncFiles.length === 0 ? (
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
              No sync files loaded.
            </p>
            <p style={{ color: "#6c757d" }}>
              Click "Load from Storage" to load saved sync files, or "Sync from
              API & Save" to fetch from API.
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
            {syncFiles.map((syncFile, index) => (
              <div
                key={`${syncFile.id}-${index}`}
                style={{
                  padding: "12px",
                  border: "1px solid #dee2e6",
                  borderRadius: "6px",
                  backgroundColor:
                    syncFile.state === "active"
                      ? "#d1ecf1"
                      : syncFile.state === "deleted"
                        ? "#f8d7da"
                        : "#fff3cd",
                  borderLeft: `4px solid ${
                    syncFile.state === "active"
                      ? "#0c5460"
                      : syncFile.state === "deleted"
                        ? "#721c24"
                        : "#856404"
                  }`,
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
                      {syncFile.file_name || `File ${syncFile.id}`}
                    </p>
                    <p style={{ margin: "0", fontSize: "14px", color: "#666" }}>
                      ID: {syncFile.id} | State: {syncFile.state} | Version:{" "}
                      {syncFile.version}
                    </p>
                    <p
                      style={{
                        margin: "5px 0 0 0",
                        fontSize: "12px",
                        color: "#666",
                      }}
                    >
                      Size: {formatFileSize(syncFile.file_size)} | Type:{" "}
                      {syncFile.mime_type || "unknown"} | Collection:{" "}
                      {syncFile.collection_id || "none"}
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
                    {new Date(syncFile.modified_at).toLocaleDateString()}
                  </div>
                </div>

                {syncFile.tombstone_version > 0 && (
                  <p
                    style={{
                      margin: "5px 0 0 0",
                      fontSize: "12px",
                      color: "#d63384",
                    }}
                  >
                    🪦 Tombstone Version: {syncFile.tombstone_version}
                  </p>
                )}

                {syncFile.encrypted_file_key && (
                  <p
                    style={{
                      margin: "5px 0 0 0",
                      fontSize: "12px",
                      color: "#28a745",
                    }}
                  >
                    🔐 Encrypted
                  </p>
                )}

                {syncFile.file_hash && (
                  <p
                    style={{
                      margin: "5px 0 0 0",
                      fontSize: "11px",
                      color: "#999",
                      fontFamily: "monospace",
                    }}
                  >
                    Hash: {syncFile.file_hash.substring(0, 16)}...
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

export default SyncFileStorageExample;
