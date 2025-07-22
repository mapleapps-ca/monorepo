// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/File/RecentFileManagerExample.jsx
// Example component demonstrating how to use the RecentFileManager

import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useFiles } from "../../../../services/Services";

const RecentFileManagerExample = () => {
  const navigate = useNavigate();
  const { authService, getCollectionManager, listCollectionManager } =
    useFiles();

  // State management
  const [fileManager, setFileManager] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [recentFiles, setRecentFiles] = useState([]);
  const [selectedFiles, setSelectedFiles] = useState(new Set());
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());
  const [eventLog, setEventLog] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Pagination state
  const [paginationState, setPaginationState] = useState({
    currentCursor: null,
    hasMore: false,
    totalLoaded: 0,
  });

  // Settings
  const [limit, setLimit] = useState(30);
  const [showDetails, setShowDetails] = useState(false);

  // Initialize file manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authService.isAuthenticated()) return;

      try {
        const { default: RecentFileManager } = await import(
          "../../../../services/Manager/File/RecentFileManager.js"
        );

        // Pass collection managers to the RecentFileManager
        const manager = new RecentFileManager(
          authService,
          getCollectionManager,
          listCollectionManager,
        );
        await manager.initialize();

        setFileManager(manager);

        // Set up event listener
        manager.addRecentFileListener(handleFileEvent);

        console.log(
          "[Example] RecentFileManager initialized with collection managers",
        );
      } catch (err) {
        console.error("[Example] Failed to initialize RecentFileManager:", err);
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();

    return () => {
      if (fileManager) {
        fileManager.removeRecentFileListener(handleFileEvent);
      }
    };
  }, [authService, getCollectionManager, listCollectionManager]);

  // Load manager status when manager is ready
  useEffect(() => {
    if (fileManager) {
      loadManagerStatus();
    }
  }, [fileManager]);

  // Handle file events
  const handleFileEvent = useCallback((eventType, eventData) => {
    console.log("[Example] Recent file event:", eventType, eventData);
    addToEventLog(eventType, eventData);

    // Update pagination state from events
    if (eventData.hasMore !== undefined || eventData.nextCursor !== undefined) {
      setPaginationState((prev) => ({
        ...prev,
        hasMore: eventData.hasMore ?? prev.hasMore,
        currentCursor: eventData.nextCursor ?? prev.currentCursor,
        totalLoaded: eventData.totalLoaded ?? prev.totalLoaded,
      }));
    }
  }, []);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!fileManager) return;

    const status = fileManager.getManagerStatus();
    setManagerStatus(status);

    // Update pagination state
    const paginationInfo = fileManager.getPaginationState();
    setPaginationState(paginationInfo);

    console.log("[Example] Manager status:", status);
  }, [fileManager]);

  // Load recent files
  const handleLoadRecentFiles = async (forceRefresh = false) => {
    if (!fileManager) {
      setError("File manager not initialized");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Loading recent files:", { limit, forceRefresh });

      const loadedFiles = await fileManager.getRecentFiles(limit, forceRefresh);

      setRecentFiles(loadedFiles);
      setSuccess(`Loaded ${loadedFiles.length} recent files`);

      // Update pagination state
      const paginationInfo = fileManager.getPaginationState();
      setPaginationState(paginationInfo);

      addToEventLog("recent_files_loaded", {
        count: loadedFiles.length,
        forceRefresh,
        hasMore: paginationInfo.hasMore,
      });
    } catch (err) {
      console.error("[Example] Failed to load recent files:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Load more recent files (pagination)
  const handleLoadMoreFiles = async () => {
    if (!fileManager || !paginationState.hasMore) {
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      console.log("[Example] Loading more recent files");

      const moreFiles = await fileManager.loadMoreRecentFiles(limit);

      // Append to existing files
      setRecentFiles((prev) => [...prev, ...moreFiles]);
      setSuccess(`Loaded ${moreFiles.length} more files`);

      // Update pagination state
      const paginationInfo = fileManager.getPaginationState();
      setPaginationState(paginationInfo);

      addToEventLog("more_recent_files_loaded", {
        count: moreFiles.length,
        totalLoaded: paginationInfo.totalLoaded,
      });
    } catch (err) {
      console.error("[Example] Failed to load more files:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle file selection
  const handleFileSelect = (fileId) => {
    const newSelected = new Set(selectedFiles);
    if (newSelected.has(fileId)) {
      newSelected.delete(fileId);
    } else {
      newSelected.add(fileId);
    }
    setSelectedFiles(newSelected);
  };

  // Handle file download
  const handleDownloadFile = async (fileId, fileName) => {
    if (!fileManager) return;

    try {
      console.log("[Example] Starting download for:", fileId, fileName);

      // Track this file as downloading
      setDownloadingFiles((prev) => new Set(prev).add(fileId));

      setSuccess(`Preparing download for "${fileName}"...`);

      // Download and save file
      await fileManager.downloadAndSaveFile(fileId);

      setSuccess(`File "${fileName}" downloaded successfully! 🎉`);

      addToEventLog("file_downloaded", {
        fileId,
        fileName,
        timestamp: new Date().toISOString(),
      });

      console.log("[Example] File download completed successfully");
    } catch (err) {
      console.error("[Example] Failed to download file:", err);

      // Enhanced error messaging
      let errorMessage = `Failed to download file: ${err.message}`;

      if (err.message.includes("File key not available")) {
        errorMessage = `Download failed: File key missing. This may happen with shared files. Try refreshing the page.`;
      } else if (err.message.includes("Collection key not available")) {
        errorMessage = `Download failed: Collection access issue. Try refreshing the page.`;
      } else if (err.message.includes("timeout")) {
        errorMessage = `Download failed: Request timed out. Please try again.`;
      }

      setError(errorMessage);

      addToEventLog("file_download_failed", {
        fileId,
        fileName,
        error: err.message,
        timestamp: new Date().toISOString(),
      });
    } finally {
      // Remove from downloading set
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  // Clear caches
  const handleClearCaches = () => {
    if (!fileManager) return;

    fileManager.clearAllCaches();
    fileManager.resetPaginationState();
    setPaginationState({
      currentCursor: null,
      hasMore: false,
      totalLoaded: 0,
    });
    setRecentFiles([]);
    setSuccess("All caches cleared and pagination reset");

    addToEventLog("caches_cleared", {
      timestamp: new Date().toISOString(),
    });
  };

  // Add event to log
  const addToEventLog = (eventType, eventData) => {
    setEventLog((prev) => [
      ...prev,
      {
        timestamp: new Date().toISOString(),
        eventType,
        eventData,
      },
    ]);
  };

  // Clear event log
  const handleClearLog = () => {
    setEventLog([]);
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Format date time
  const formatDateTime = (dateString) => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "N/A";
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return "Invalid Date";
    }
  };

  // Get relative time
  const getRelativeTime = (dateString) => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "Unknown";

    try {
      const date = new Date(dateString);
      const now = new Date();
      const diffMs = now - date;
      const diffMins = Math.floor(diffMs / (1000 * 60));
      const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

      if (diffMins < 1) return "Just now";
      if (diffMins < 60) return `${diffMins} minutes ago`;
      if (diffHours < 24) return `${diffHours} hours ago`;
      if (diffDays < 30) return `${diffDays} days ago`;
      return formatDateTime(dateString);
    } catch {
      return "Invalid Date";
    }
  };

  // Clear messages after 5 seconds
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        setSuccess("");
        setError("");
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [success, error]);

  if (!authService.isAuthenticated()) {
    return (
      <div style={{ padding: "20px", textAlign: "center" }}>
        <p>Please log in to access this example.</p>
        <button onClick={() => navigate("/login")}>Go to Login</button>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "1400px", margin: "0 auto" }}>
      <div style={{ marginBottom: "20px" }}>
        <button onClick={() => navigate("/dashboard")}>
          ← Back to Dashboard
        </button>
      </div>

      <h2>📋 Recent File Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This example demonstrates listing recent files across all collections
        with E2EE decryption.
        <br />
        <strong>Features:</strong> Recent files, pagination, download files,
        cache management
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
            {managerStatus.isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can List Files:</strong>{" "}
            {managerStatus.canListFiles ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Recent Files:</strong> {recentFiles.length}
          </div>
          <div>
            <strong>Total Loaded:</strong> {paginationState.totalLoaded}
          </div>
          <div>
            <strong>Has More:</strong>{" "}
            {paginationState.hasMore ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Listeners:</strong> {managerStatus.listenerCount || 0}
          </div>
          <div>
            <strong>Cache Valid:</strong>{" "}
            {managerStatus.storage?.isValid ? "✅ Yes" : "❌ No"}
          </div>
        </div>
      </div>

      {/* Controls */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e3f2fd",
          borderRadius: "6px",
          border: "1px solid #bbdefb",
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
          <h4 style={{ margin: 0 }}>🔧 Controls:</h4>
          <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
            <label
              style={{ display: "flex", alignItems: "center", gap: "5px" }}
            >
              Limit:
              <select
                value={limit}
                onChange={(e) => setLimit(Number(e.target.value))}
                style={{ padding: "4px", borderRadius: "4px" }}
              >
                <option value={10}>10</option>
                <option value={30}>30</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
              </select>
            </label>
            <label
              style={{ display: "flex", alignItems: "center", gap: "8px" }}
            >
              <input
                type="checkbox"
                checked={showDetails}
                onChange={(e) => setShowDetails(e.target.checked)}
              />
              Show Details
            </label>
          </div>
        </div>

        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          <button
            onClick={() => handleLoadRecentFiles(false)}
            disabled={isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isLoading ? "not-allowed" : "pointer",
            }}
          >
            {isLoading ? "🔄 Loading..." : "📋 Load Recent Files"}
          </button>
          <button
            onClick={() => handleLoadRecentFiles(true)}
            disabled={isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#ffc107",
              color: "#212529",
              border: "none",
              borderRadius: "4px",
              cursor: isLoading ? "not-allowed" : "pointer",
            }}
          >
            🔄 Force Refresh
          </button>
          <button
            onClick={handleLoadMoreFiles}
            disabled={isLoading || !paginationState.hasMore}
            style={{
              padding: "8px 16px",
              backgroundColor: "#17a2b8",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                isLoading || !paginationState.hasMore
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            📄 Load More ({paginationState.hasMore ? "Available" : "None"})
          </button>
          <button
            onClick={handleClearCaches}
            disabled={!fileManager}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !fileManager ? "not-allowed" : "pointer",
            }}
          >
            🗑️ Clear Cache
          </button>
        </div>
      </div>

      {/* Pagination Info */}
      {paginationState.totalLoaded > 0 && (
        <div
          style={{
            marginBottom: "20px",
            padding: "10px 15px",
            backgroundColor: "#fff3cd",
            borderRadius: "6px",
            border: "1px solid #ffeaa7",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>
            📊 Loaded: {paginationState.totalLoaded} files
            {paginationState.hasMore && " (more available)"}
          </span>
          <span>🔄 Current page: {recentFiles.length} files</span>
        </div>
      )}

      {/* Success/Error Messages */}
      {success && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#d4edda",
            borderRadius: "6px",
            color: "#155724",
            border: "1px solid #c3e6cb",
          }}
        >
          ✅ {success}
        </div>
      )}

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
          ❌ {error}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && recentFiles.length === 0 && (
        <div
          style={{
            textAlign: "center",
            padding: "60px 20px",
            backgroundColor: "#f8f9fa",
            borderRadius: "8px",
            border: "2px dashed #dee2e6",
          }}
        >
          <div style={{ fontSize: "64px", marginBottom: "20px" }}>📁</div>
          <h3>No recent files found</h3>
          <p style={{ color: "#666" }}>
            No recent files are available. Try creating or modifying some files
            first.
          </p>
        </div>
      )}

      {/* Files List */}
      {!isLoading && recentFiles.length > 0 && (
        <div>
          {/* Bulk Actions */}
          {selectedFiles.size > 0 && (
            <div
              style={{
                backgroundColor: "#e9ecef",
                padding: "15px",
                marginBottom: "20px",
                borderRadius: "4px",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
              }}
            >
              <span>{selectedFiles.size} file(s) selected</span>
              <button
                onClick={() => setSelectedFiles(new Set())}
                style={{
                  padding: "8px 16px",
                  backgroundColor: "#6c757d",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                }}
              >
                Clear Selection
              </button>
            </div>
          )}

          {/* Files Table */}
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr
                style={{
                  backgroundColor: "#f8f9fa",
                  borderBottom: "2px solid #dee2e6",
                }}
              >
                <th
                  style={{ padding: "12px", textAlign: "left", width: "40px" }}
                >
                  <input
                    type="checkbox"
                    checked={
                      selectedFiles.size === recentFiles.length &&
                      recentFiles.length > 0
                    }
                    onChange={(e) => {
                      if (e.target.checked) {
                        setSelectedFiles(new Set(recentFiles.map((f) => f.id)));
                      } else {
                        setSelectedFiles(new Set());
                      }
                    }}
                  />
                </th>
                <th style={{ padding: "12px", textAlign: "left" }}>File</th>
                <th style={{ padding: "12px", textAlign: "left" }}>Size</th>
                <th style={{ padding: "12px", textAlign: "left" }}>
                  Collection
                </th>
                <th style={{ padding: "12px", textAlign: "left" }}>Modified</th>
                {showDetails && (
                  <>
                    <th style={{ padding: "12px", textAlign: "left" }}>
                      Version
                    </th>
                    <th style={{ padding: "12px", textAlign: "left" }}>
                      State
                    </th>
                  </>
                )}
                <th style={{ padding: "12px", textAlign: "center" }}>
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {recentFiles.map((file) => (
                <tr key={file.id} style={{ borderBottom: "1px solid #dee2e6" }}>
                  <td style={{ padding: "12px" }}>
                    <input
                      type="checkbox"
                      checked={selectedFiles.has(file.id)}
                      onChange={() => handleFileSelect(file.id)}
                    />
                  </td>
                  <td style={{ padding: "12px" }}>
                    <div
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: "10px",
                      }}
                    >
                      <span style={{ fontSize: "24px" }}>📄</span>
                      <div>
                        <div>{file.name || "[Encrypted]"}</div>
                        <div style={{ fontSize: "12px", color: "#666" }}>
                          ID: {file.id.substring(0, 8)}...
                          {file._decryptionError && (
                            <span style={{ color: "#ff6b6b" }}>
                              {" "}
                              • Decrypt failed
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td style={{ padding: "12px" }}>
                    {file.size
                      ? formatFileSize(file.size)
                      : file.encrypted_file_size_in_bytes
                        ? `${formatFileSize(file.encrypted_file_size_in_bytes)} (encrypted)`
                        : "Unknown"}
                  </td>
                  <td style={{ padding: "12px" }}>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      {file.collection_id.substring(0, 8)}...
                    </div>
                  </td>
                  <td style={{ padding: "12px" }}>
                    <div>{getRelativeTime(file.modified_at)}</div>
                    {showDetails && (
                      <div style={{ fontSize: "10px", color: "#999" }}>
                        {formatDateTime(file.modified_at)}
                      </div>
                    )}
                  </td>
                  {showDetails && (
                    <>
                      <td style={{ padding: "12px" }}>v{file.version || 1}</td>
                      <td style={{ padding: "12px" }}>
                        <span
                          style={{
                            padding: "2px 6px",
                            borderRadius: "4px",
                            fontSize: "10px",
                            backgroundColor:
                              file.state === "active" ? "#28a745" : "#6c757d",
                            color: "white",
                          }}
                        >
                          {(file.state || "active").toUpperCase()}
                        </span>
                      </td>
                    </>
                  )}
                  <td style={{ padding: "12px", textAlign: "center" }}>
                    <button
                      onClick={() =>
                        handleDownloadFile(
                          file.id,
                          file.name || "downloaded_file",
                        )
                      }
                      disabled={
                        !file._file_key || downloadingFiles.has(file.id)
                      }
                      style={{
                        padding: "4px 8px",
                        border: "none",
                        borderRadius: "4px",
                        cursor:
                          !file._file_key || downloadingFiles.has(file.id)
                            ? "not-allowed"
                            : "pointer",
                        fontSize: "12px",
                        backgroundColor:
                          file._file_key && !downloadingFiles.has(file.id)
                            ? "#007bff"
                            : "#ccc",
                        color: "white",
                      }}
                      title={
                        !file._file_key
                          ? "File key not available - try refreshing"
                          : "Download file"
                      }
                    >
                      {downloadingFiles.has(file.id)
                        ? "Downloading..."
                        : "Download"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Event Log */}
      <div style={{ marginTop: "40px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <h3>📋 Recent File Event Log ({eventLog.length})</h3>
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
              No recent file events logged yet.
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
            {eventLog
              .slice()
              .reverse()
              .map((event, index) => (
                <div
                  key={`${event.timestamp}-${index}`}
                  style={{
                    padding: "10px",
                    borderBottom:
                      index < eventLog.length - 1
                        ? "1px solid #dee2e6"
                        : "none",
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
    </div>
  );
};

export default RecentFileManagerExample;
