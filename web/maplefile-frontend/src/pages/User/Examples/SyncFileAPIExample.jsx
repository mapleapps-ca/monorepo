// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/SyncFileAPIExample.jsx
// Example component demonstrating how to use the SyncFileAPIService

import React, { useState } from "react";
import { useServices } from "../../../services/Services";

const SyncFileAPIExample = () => {
  const { syncFileAPIService } = useServices();
  const [syncFiles, setSyncFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [selectedCollectionId, setSelectedCollectionId] = useState("");

  // The ONLY function you need - sync all files from API
  const handleSyncAll = async () => {
    setLoading(true);
    setError(null);
    setSyncFiles([]);

    try {
      console.log("🔄 Starting complete API sync...");

      // This ONE function call gets ALL sync files from API automatically
      const allSyncFiles = await syncFileAPIService.syncAllFiles();

      setSyncFiles(allSyncFiles);
      console.log("✅ API sync complete:", allSyncFiles.length, "sync files");
    } catch (err) {
      setError(err.message);
      console.error("❌ API sync failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Sync files for a specific collection
  const handleSyncByCollection = async () => {
    if (!selectedCollectionId) {
      setError("Please enter a collection ID");
      return;
    }

    setLoading(true);
    setError(null);
    setSyncFiles([]);

    try {
      console.log("🔄 Starting collection sync for:", selectedCollectionId);

      const collectionFiles =
        await syncFileAPIService.syncFilesByCollection(selectedCollectionId);

      setSyncFiles(collectionFiles);
      console.log(
        "✅ Collection sync complete:",
        collectionFiles.length,
        "sync files",
      );
    } catch (err) {
      setError(err.message);
      console.error("❌ Collection sync failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Group files by collection for display
  const getFilesByCollection = () => {
    const grouped = {};
    syncFiles.forEach((file) => {
      const collectionId = file.collection_id || "no_collection";
      if (!grouped[collectionId]) {
        grouped[collectionId] = [];
      }
      grouped[collectionId].push(file);
    });
    return grouped;
  };

  // Calculate total file size
  const getTotalFileSize = () => {
    return syncFiles.reduce((sum, file) => sum + (file.file_size || 0), 0);
  };

  // Format file size for display
  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
  };

  const groupedFiles = getFilesByCollection();

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>🔄 Sync All Files from API (Simplified)</h2>

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
            marginRight: "10px",
          }}
        >
          {loading ? "🔄 Syncing from API..." : "🚀 Sync All Files from API"}
        </button>

        <div style={{ display: "inline-block", marginLeft: "20px" }}>
          <input
            type="text"
            placeholder="Collection ID (optional)"
            value={selectedCollectionId}
            onChange={(e) => setSelectedCollectionId(e.target.value)}
            style={{
              padding: "10px",
              marginRight: "10px",
              borderRadius: "4px",
              border: "1px solid #ddd",
            }}
          />
          <button
            onClick={handleSyncByCollection}
            disabled={loading}
            style={{
              padding: "10px 20px",
              backgroundColor: loading ? "#6c757d" : "#28a745",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: loading ? "not-allowed" : "pointer",
            }}
          >
            Sync Collection
          </button>
        </div>
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
          <strong>Sync Files Found:</strong> {syncFiles.length}
        </p>
        <p style={{ margin: "5px 0" }}>
          <strong>Total File Size:</strong> {formatFileSize(getTotalFileSize())}
        </p>
        <p style={{ margin: "5px 0" }}>
          <strong>Collections:</strong> {Object.keys(groupedFiles).length}
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

      {/* Sync Files Display */}
      <div>
        <h3>📁 Sync Files from API ({syncFiles.length})</h3>
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
              No sync files loaded yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Click "Sync All Files from API" to load all your sync files.
            </p>
          </div>
        ) : (
          <div>
            {Object.entries(groupedFiles).map(([collectionId, files]) => (
              <div key={collectionId} style={{ marginBottom: "20px" }}>
                <h4 style={{ marginBottom: "10px", color: "#495057" }}>
                  Collection: {collectionId} ({files.length} files)
                </h4>
                <div
                  style={{
                    display: "grid",
                    gap: "10px",
                    maxHeight: "300px",
                    overflow: "auto",
                    border: "1px solid #dee2e6",
                    borderRadius: "6px",
                    padding: "10px",
                    backgroundColor: "#f8f9fa",
                  }}
                >
                  {files.map((syncFile, index) => (
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
                          <p
                            style={{ margin: "0 0 5px 0", fontWeight: "bold" }}
                          >
                            {syncFile.file_name || `File ${syncFile.id}`}
                          </p>
                          <p
                            style={{
                              margin: "0",
                              fontSize: "14px",
                              color: "#666",
                            }}
                          >
                            ID: {syncFile.id} | State: {syncFile.state} |
                            Version: {syncFile.version}
                          </p>
                          <p
                            style={{
                              margin: "5px 0 0 0",
                              fontSize: "12px",
                              color: "#666",
                            }}
                          >
                            Size: {formatFileSize(syncFile.file_size)} | Type:{" "}
                            {syncFile.mime_type || "unknown"}
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
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default SyncFileAPIExample;
