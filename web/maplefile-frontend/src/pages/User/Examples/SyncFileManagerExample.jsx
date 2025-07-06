// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/SyncFileManagerExample.jsx
// Updated to use SyncFileManager instead of SyncFileService
import React, { useState, useEffect } from "react";
import { useServices } from "../../../hooks/useService.jsx";
import useSyncFileManager from "../../../hooks/useSyncFileManager.js";

const SyncFileManagerExample = () => {
  // Option 1: Use the new hook (recommended)
  const {
    syncFiles,
    loading,
    error,
    getSyncFiles,
    getSyncFilesByCollection,
    getSyncFile,
    refreshSyncFiles,
    clearSyncFiles,
    statistics,
    sizeStatistics,
    managerStatus,
    debugInfo,
  } = useSyncFileManager();

  // Option 2: Direct service access (for comparison)
  const { syncFileManager } = useServices();
  const [directSyncFiles, setDirectSyncFiles] = useState([]);
  const [directLoading, setDirectLoading] = useState(false);
  const [directError, setDirectError] = useState(null);
  const [selectedCollectionId, setSelectedCollectionId] = useState("");
  const [selectedFileId, setSelectedFileId] = useState("");

  useEffect(() => {
    loadSyncFilesViaHook();
  }, []);

  // Using the hook (recommended approach)
  const loadSyncFilesViaHook = async () => {
    try {
      await getSyncFiles();
      console.log("Loaded sync files via hook");
    } catch (error) {
      console.error("Failed to load sync files via hook:", error);
    }
  };

  const refreshViaHook = async () => {
    try {
      await refreshSyncFiles();
      console.log("Refreshed sync files via hook");
    } catch (error) {
      console.error("Failed to refresh sync files via hook:", error);
    }
  };

  const clearViaHook = () => {
    try {
      clearSyncFiles();
      console.log("Cleared sync files via hook");
    } catch (error) {
      console.error("Failed to clear sync files via hook:", error);
    }
  };

  const loadCollectionViaHook = async () => {
    if (!selectedCollectionId) {
      alert("Please enter a collection ID");
      return;
    }
    try {
      await getSyncFilesByCollection(selectedCollectionId);
      console.log("Loaded collection files via hook");
    } catch (error) {
      console.error("Failed to load collection files via hook:", error);
    }
  };

  const loadFileViaHook = async () => {
    if (!selectedFileId) {
      alert("Please enter a file ID");
      return;
    }
    try {
      const file = await getSyncFile(selectedFileId);
      console.log("Loaded sync file via hook:", file);
      alert(`Found file: ${file.file_name || file.id}`);
    } catch (error) {
      console.error("Failed to load sync file via hook:", error);
    }
  };

  // Using direct service access (alternative approach)
  const loadSyncFilesDirectly = async () => {
    try {
      setDirectLoading(true);
      setDirectError(null);
      const files = await syncFileManager.getSyncFiles();
      setDirectSyncFiles(files);
      console.log("Loaded sync files directly:", files);
    } catch (error) {
      setDirectError(error.message || "Failed to load sync files");
      console.error("Failed to load sync files directly:", error);
    } finally {
      setDirectLoading(false);
    }
  };

  const refreshDirectly = async () => {
    try {
      setDirectLoading(true);
      setDirectError(null);
      const files = await syncFileManager.forceRefreshSyncFiles();
      setDirectSyncFiles(files);
      console.log("Refreshed sync files directly:", files);
    } catch (error) {
      setDirectError(error.message || "Failed to refresh sync files");
      console.error("Failed to refresh sync files directly:", error);
    } finally {
      setDirectLoading(false);
    }
  };

  const clearDirectly = () => {
    try {
      const success = syncFileManager.clearSyncFiles();
      if (success) {
        setDirectSyncFiles([]);
        console.log("Cleared sync files directly");
      }
    } catch (error) {
      setDirectError(error.message || "Failed to clear sync files");
      console.error("Failed to clear sync files directly:", error);
    }
  };

  // Format file size for display
  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
  };

  if (loading) {
    return (
      <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
        <h2>🔄 Sync File Manager Example</h2>
        <p>Loading sync files...</p>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>🔄 Sync File Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the new <strong>SyncFileManager</strong> with
        both hook-based and direct service access patterns.
      </p>

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
            {managerStatus?.isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Manager Loading:</strong>{" "}
            {managerStatus?.isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>API Loading:</strong>{" "}
            {managerStatus?.isAPILoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Storage Count:</strong>{" "}
            {managerStatus?.storage?.syncFilesCount || 0}
          </div>
          <div>
            <strong>Last Sync:</strong>{" "}
            {managerStatus?.lastSync
              ? new Date(managerStatus.lastSync).toLocaleString()
              : "Never"}
          </div>
          <div>
            <strong>Collections:</strong>{" "}
            {managerStatus?.storage?.collectionBreakdown
              ? Object.keys(managerStatus.storage.collectionBreakdown).length
              : 0}
          </div>
        </div>
      </div>

      {/* Statistics */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>📈 Statistics:</h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))",
            gap: "10px",
          }}
        >
          <div>
            <strong>Total:</strong> {statistics?.total || 0}
          </div>
          <div>
            <strong>Active:</strong> {statistics?.active || 0}
          </div>
          <div>
            <strong>Deleted:</strong> {statistics?.deleted || 0}
          </div>
          <div>
            <strong>Archived:</strong> {statistics?.archived || 0}
          </div>
          <div>
            <strong>Total Size:</strong>{" "}
            {formatFileSize(statistics?.totalSize || 0)}
          </div>
          <div>
            <strong>Collections:</strong> {statistics?.collections || 0}
          </div>
        </div>

        {sizeStatistics?.largestFile && (
          <div style={{ marginTop: "10px", fontSize: "14px" }}>
            <p style={{ margin: "5px 0" }}>
              <strong>Largest File:</strong>{" "}
              {sizeStatistics.largestFile.file_name ||
                sizeStatistics.largestFile.id}
              ({formatFileSize(sizeStatistics.largestFile.file_size)})
            </p>
            <p style={{ margin: "5px 0" }}>
              <strong>Average Size:</strong>{" "}
              {formatFileSize(sizeStatistics.averageSize)}
            </p>
          </div>
        )}
      </div>

      {/* Hook-based Approach */}
      <div style={{ marginBottom: "30px" }}>
        <h3>🎣 Hook-based Approach (Recommended)</h3>
        <div
          style={{
            marginBottom: "20px",
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
            alignItems: "center",
          }}
        >
          <button
            onClick={loadSyncFilesViaHook}
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
            {loading ? "🔄 Loading..." : "📂 Load via Hook"}
          </button>

          <button
            onClick={refreshViaHook}
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
            {loading ? "🔄 Refreshing..." : "🔄 Refresh via Hook"}
          </button>

          <button
            onClick={clearViaHook}
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
            🗑️ Clear via Hook
          </button>

          <div style={{ display: "flex", gap: "5px", alignItems: "center" }}>
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
              onClick={loadCollectionViaHook}
              disabled={loading}
              style={{
                padding: "10px 15px",
                backgroundColor: "#6f42c1",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: loading ? "not-allowed" : "pointer",
              }}
            >
              📁 Load
            </button>
          </div>

          <div style={{ display: "flex", gap: "5px", alignItems: "center" }}>
            <input
              type="text"
              placeholder="File ID"
              value={selectedFileId}
              onChange={(e) => setSelectedFileId(e.target.value)}
              style={{
                padding: "8px",
                borderRadius: "4px",
                border: "1px solid #ddd",
              }}
            />
            <button
              onClick={loadFileViaHook}
              disabled={loading}
              style={{
                padding: "10px 15px",
                backgroundColor: "#fd7e14",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: loading ? "not-allowed" : "pointer",
              }}
            >
              📄 Get
            </button>
          </div>
        </div>

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
            <h4 style={{ margin: "0 0 10px 0" }}>❌ Hook Error:</h4>
            <p style={{ margin: 0 }}>{error}</p>
          </div>
        )}

        <div>
          <h4>📁 Sync Files via Hook ({syncFiles.length})</h4>
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
                No sync files loaded via hook.
              </p>
            </div>
          ) : (
            <div
              style={{
                display: "grid",
                gap: "10px",
                maxHeight: "300px",
                overflow: "auto",
                border: "1px solid #dee2e6",
                borderRadius: "6px",
                padding: "10px",
              }}
            >
              {syncFiles.map((file, index) => (
                <div
                  key={`hook-${file.id}-${index}`}
                  style={{
                    padding: "12px",
                    border: "1px solid #dee2e6",
                    borderRadius: "6px",
                    backgroundColor:
                      file.state === "active"
                        ? "#d1ecf1"
                        : file.state === "deleted"
                          ? "#f8d7da"
                          : "#fff3cd",
                    borderLeft: `4px solid ${
                      file.state === "active"
                        ? "#0c5460"
                        : file.state === "deleted"
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
                        {file.file_name || `File ${file.id}`}
                      </p>
                      <p
                        style={{ margin: "0", fontSize: "14px", color: "#666" }}
                      >
                        ID: {file.id} | State: {file.state} | Version:{" "}
                        {file.version}
                      </p>
                      <p
                        style={{
                          margin: "5px 0 0 0",
                          fontSize: "12px",
                          color: "#666",
                        }}
                      >
                        Size: {formatFileSize(file.file_size)} | Type:{" "}
                        {file.mime_type || "unknown"} | Collection:{" "}
                        {file.collection_id || "none"}
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
                      {new Date(file.modified_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Direct Service Approach */}
      <div style={{ marginBottom: "30px" }}>
        <h3>🔧 Direct Service Approach (Alternative)</h3>
        <div
          style={{
            marginBottom: "20px",
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
          }}
        >
          <button
            onClick={loadSyncFilesDirectly}
            disabled={directLoading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#6f42c1",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: directLoading ? "not-allowed" : "pointer",
            }}
          >
            {directLoading ? "🔄 Loading..." : "📂 Load Directly"}
          </button>

          <button
            onClick={refreshDirectly}
            disabled={directLoading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#fd7e14",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: directLoading ? "not-allowed" : "pointer",
            }}
          >
            {directLoading ? "🔄 Refreshing..." : "🔄 Refresh Directly"}
          </button>

          <button
            onClick={clearDirectly}
            disabled={directLoading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: directLoading ? "not-allowed" : "pointer",
            }}
          >
            🗑️ Clear Directly
          </button>
        </div>

        {directError && (
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
            <h4 style={{ margin: "0 0 10px 0" }}>❌ Direct Error:</h4>
            <p style={{ margin: 0 }}>{directError}</p>
          </div>
        )}

        <div>
          <h4>📁 Sync Files via Direct Service ({directSyncFiles.length})</h4>
          {directSyncFiles.length === 0 ? (
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
                No sync files loaded directly.
              </p>
            </div>
          ) : (
            <div
              style={{
                display: "grid",
                gap: "10px",
                maxHeight: "300px",
                overflow: "auto",
                border: "1px solid #dee2e6",
                borderRadius: "6px",
                padding: "10px",
              }}
            >
              {directSyncFiles.map((file, index) => (
                <div
                  key={`direct-${file.id}-${index}`}
                  style={{
                    padding: "12px",
                    border: "1px solid #dee2e6",
                    borderRadius: "6px",
                    backgroundColor:
                      file.state === "active"
                        ? "#d4edda"
                        : file.state === "deleted"
                          ? "#f8d7da"
                          : "#fff3cd",
                    borderLeft: `4px solid ${
                      file.state === "active"
                        ? "#155724"
                        : file.state === "deleted"
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
                        {file.file_name || `File ${file.id}`}
                      </p>
                      <p
                        style={{ margin: "0", fontSize: "14px", color: "#666" }}
                      >
                        ID: {file.id} | State: {file.state} | Version:{" "}
                        {file.version}
                      </p>
                      <p
                        style={{
                          margin: "5px 0 0 0",
                          fontSize: "12px",
                          color: "#666",
                        }}
                      >
                        Size: {formatFileSize(file.file_size)} | Type:{" "}
                        {file.mime_type || "unknown"}
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
                      {new Date(file.modified_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
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
          {JSON.stringify(debugInfo, null, 2)}
        </pre>
      </details>
    </div>
  );
};

export default SyncFileManagerExample;
