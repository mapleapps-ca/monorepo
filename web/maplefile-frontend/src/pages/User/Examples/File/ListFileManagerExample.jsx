// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/File/ListFileManagerExample.jsx
// Example component demonstrating how to use the ListFileManager

import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useFiles } from "../../../../services/Services";

const ListFileManagerExample = () => {
  const navigate = useNavigate();
  const { authService, getCollectionManager, listCollectionManager } =
    useFiles();

  // State management
  const [fileManager, setFileManager] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [selectedCollectionId, setSelectedCollectionId] = useState("");
  const [files, setFiles] = useState([]);
  const [selectedFiles, setSelectedFiles] = useState(new Set());
  const [viewMode, setViewMode] = useState("active"); // active, archived, deleted, all
  const [showDetails, setShowDetails] = useState(false);
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());
  const [eventLog, setEventLog] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Collections for the example
  const [availableCollections, setAvailableCollections] = useState([]);
  const [isLoadingCollections, setIsLoadingCollections] = useState(false);

  // Initialize file manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authService.isAuthenticated()) return;

      try {
        const { default: ListFileManager } = await import(
          "../../../../services/Manager/File/ListFileManager.js"
        );

        // Pass collection managers to the ListFileManager
        const manager = new ListFileManager(
          authService,
          getCollectionManager,
          listCollectionManager,
        );
        await manager.initialize();

        setFileManager(manager);

        // Set up event listener
        manager.addFileListingListener(handleFileEvent);

        console.log(
          "[Example] ListFileManager initialized with collection managers",
        );
      } catch (err) {
        console.error("[Example] Failed to initialize ListFileManager:", err);
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();

    return () => {
      if (fileManager) {
        fileManager.removeFileListingListener(handleFileEvent);
      }
    };
  }, [authService, getCollectionManager, listCollectionManager]);

  // Load collections when managers are ready
  useEffect(() => {
    if (getCollectionManager && listCollectionManager) {
      loadCollections();
    }
  }, [getCollectionManager, listCollectionManager]);

  // Load manager status when manager is ready
  useEffect(() => {
    if (fileManager) {
      loadManagerStatus();
    }
  }, [fileManager]);

  // Load collections for selection
  const loadCollections = async () => {
    setIsLoadingCollections(true);
    try {
      console.log("[Example] Loading collections...");

      const result = await listCollectionManager.listCollections(false);

      if (result.collections && result.collections.length > 0) {
        setAvailableCollections(result.collections.slice(0, 10)); // Limit to first 10
        console.log("[Example] Collections loaded:", result.collections.length);
      } else {
        console.log("[Example] No collections found");
        setError("No collections found. Please create a collection first.");
      }
    } catch (err) {
      console.error("[Example] Failed to load collections:", err);
      setError(`Failed to load collections: ${err.message}`);
    } finally {
      setIsLoadingCollections(false);
    }
  };

  // Handle file events
  const handleFileEvent = useCallback((eventType, eventData) => {
    console.log("[Example] File event:", eventType, eventData);
    addToEventLog(eventType, eventData);
  }, []);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!fileManager) return;

    const status = fileManager.getManagerStatus();
    setManagerStatus(status);
    console.log("[Example] Manager status:", status);
  }, [fileManager]);

  // Load files for selected collection
  const handleLoadFiles = async (forceRefresh = false) => {
    if (!fileManager || !selectedCollectionId) {
      setError("Please select a collection");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      console.log(
        "[Example] Loading files for collection:",
        selectedCollectionId,
      );

      // Determine states to include based on view mode
      const statesToInclude = getStatesToInclude(viewMode);

      const loadedFiles = await fileManager.listFilesByCollection(
        selectedCollectionId,
        statesToInclude,
        forceRefresh,
      );

      setFiles(loadedFiles);
      setSuccess(`Loaded ${loadedFiles.length} files in ${viewMode} state(s)`);

      addToEventLog("files_loaded", {
        collectionId: selectedCollectionId,
        count: loadedFiles.length,
        viewMode,
        forceRefresh,
      });
    } catch (err) {
      console.error("[Example] Failed to load files:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Get states to include based on view mode
  const getStatesToInclude = (mode) => {
    if (!fileManager) return null;

    switch (mode) {
      case "active":
        return [fileManager.FILE_STATES.ACTIVE];
      case "archived":
        return [fileManager.FILE_STATES.ARCHIVED];
      case "deleted":
        return [fileManager.FILE_STATES.DELETED];
      case "pending":
        return [fileManager.FILE_STATES.PENDING];
      case "all":
        return Object.values(fileManager.FILE_STATES);
      default:
        return [fileManager.FILE_STATES.ACTIVE];
    }
  };

  // Get current files based on view mode
  const getCurrentFiles = () => {
    if (!fileManager) return files;

    switch (viewMode) {
      case "active":
        return files.filter((f) => f._is_active);
      case "archived":
        return files.filter((f) => f._is_archived);
      case "deleted":
        return files.filter((f) => f._is_deleted);
      case "pending":
        return files.filter((f) => f._is_pending);
      case "all":
        return files;
      default:
        return files.filter((f) => f._is_active);
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

    const file = files.find((f) => f.id === fileId);
    if (file && !fileManager.canDownloadFile(file)) {
      setError("This file cannot be downloaded in its current state.");
      return;
    }

    try {
      console.log("[Example] Starting download for:", fileId, fileName);

      // Track this file as downloading
      setDownloadingFiles((prev) => new Set(prev).add(fileId));

      await fileManager.downloadAndSaveFile(fileId);

      setSuccess(`File "${fileName}" downloaded successfully`);

      addToEventLog("file_downloaded", {
        fileId,
        fileName,
      });

      console.log("[Example] File download completed successfully");
    } catch (err) {
      console.error("[Example] Failed to download file:", err);
      setError(`Failed to download file: ${err.message}`);
    } finally {
      // Remove from downloading set
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  // Handle view mode change
  const handleViewModeChange = (newMode) => {
    setViewMode(newMode);
    setSelectedFiles(new Set());

    // Auto-reload if we have a selected collection
    if (selectedCollectionId && fileManager) {
      handleLoadFiles(false);
    }
  };

  // Clear caches
  const handleClearCaches = () => {
    if (!fileManager) return;

    if (selectedCollectionId) {
      fileManager.clearCollectionCache(selectedCollectionId);
      setSuccess("Collection cache cleared");
    } else {
      fileManager.clearAllCaches();
      setSuccess("All caches cleared");
    }

    addToEventLog("caches_cleared", {
      collectionId: selectedCollectionId || "all",
    });
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "N/A";
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return "Invalid Date";
    }
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

  // Get state color
  const getStateColor = (state) => {
    if (!fileManager) return "#6c757d";

    switch (state) {
      case fileManager.FILE_STATES.ACTIVE:
        return "#28a745";
      case fileManager.FILE_STATES.ARCHIVED:
        return "#6c757d";
      case fileManager.FILE_STATES.DELETED:
        return "#dc3545";
      case fileManager.FILE_STATES.PENDING:
        return "#ffc107";
      default:
        return "#6c757d";
    }
  };

  // Get state icon
  const getStateIcon = (state) => {
    if (!fileManager) return "❓";

    switch (state) {
      case fileManager.FILE_STATES.ACTIVE:
        return "✅";
      case fileManager.FILE_STATES.ARCHIVED:
        return "📦";
      case fileManager.FILE_STATES.DELETED:
        return "🗑️";
      case fileManager.FILE_STATES.PENDING:
        return "⏳";
      default:
        return "❓";
    }
  };

  // Get file statistics
  const getFileStats = () => {
    if (!fileManager || !selectedCollectionId) return {};
    return fileManager.getFileStats(selectedCollectionId);
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

  const currentFiles = getCurrentFiles();
  const fileStats = getFileStats();

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

      <h2>📋 List File Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This example demonstrates the complete file listing workflow with E2EE
        decryption.
        <br />
        <strong>Features:</strong> List files by collection, view different
        states, download files, cache management
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
            <strong>Available Collections:</strong>{" "}
            {availableCollections.length}
          </div>
          <div>
            <strong>Loaded Files:</strong> {files.length}
          </div>
          <div>
            <strong>Listeners:</strong> {managerStatus.listenerCount || 0}
          </div>
          <div>
            <strong>Collection Manager:</strong>{" "}
            {getCollectionManager ? "✅ Available" : "❌ Missing"}
          </div>
          <div>
            <strong>List Collection Manager:</strong>{" "}
            {listCollectionManager ? "✅ Available" : "❌ Missing"}
          </div>
        </div>
      </div>

      {/* Collection Selection */}
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
            marginBottom: "10px",
          }}
        >
          <h4 style={{ margin: 0 }}>
            📂 Select Collection ({availableCollections.length} available):
          </h4>
          <button
            onClick={loadCollections}
            disabled={isLoadingCollections}
            style={{
              padding: "5px 15px",
              backgroundColor: "#1976d2",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isLoadingCollections ? "not-allowed" : "pointer",
            }}
          >
            {isLoadingCollections ? "🔄 Loading..." : "🔄 Refresh"}
          </button>
        </div>

        <div style={{ marginBottom: "15px" }}>
          <select
            value={selectedCollectionId}
            onChange={(e) => setSelectedCollectionId(e.target.value)}
            disabled={availableCollections.length === 0}
            style={{
              width: "100%",
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              backgroundColor:
                availableCollections.length === 0 ? "#f5f5f5" : "white",
            }}
          >
            <option value="">
              {availableCollections.length === 0
                ? "No collections available - click refresh"
                : "Select a collection to list files..."}
            </option>
            {availableCollections.map((col) => (
              <option key={col.id} value={col.id}>
                {col.name || "[Encrypted]"} ({col.collection_type}) -{" "}
                {col.id.substring(0, 8)}...
              </option>
            ))}
          </select>
        </div>

        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          <button
            onClick={() => handleLoadFiles(false)}
            disabled={!selectedCollectionId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !selectedCollectionId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            {isLoading ? "🔄 Loading..." : "📋 Load Files"}
          </button>
          <button
            onClick={() => handleLoadFiles(true)}
            disabled={!selectedCollectionId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#ffc107",
              color: "#212529",
              border: "none",
              borderRadius: "4px",
              cursor:
                !selectedCollectionId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            🔄 Force Refresh
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

      {/* File Statistics */}
      {selectedCollectionId &&
        fileStats &&
        Object.keys(fileStats).length > 0 && (
          <div
            style={{
              display: "flex",
              gap: "15px",
              marginBottom: "20px",
              padding: "15px",
              backgroundColor: "#f8f9fa",
              borderRadius: "8px",
              flexWrap: "wrap",
            }}
          >
            <div>
              <strong>Total:</strong> {fileStats.total || 0}
            </div>
            <div>
              <strong>Active:</strong>{" "}
              <span style={{ color: getStateColor("active") }}>
                {fileStats.active || 0}
              </span>
            </div>
            <div>
              <strong>Archived:</strong>{" "}
              <span style={{ color: getStateColor("archived") }}>
                {fileStats.archived || 0}
              </span>
            </div>
            <div>
              <strong>Deleted:</strong>{" "}
              <span style={{ color: getStateColor("deleted") }}>
                {fileStats.deleted || 0}
              </span>
            </div>
            <div>
              <strong>Pending:</strong>{" "}
              <span style={{ color: getStateColor("pending") }}>
                {fileStats.pending || 0}
              </span>
            </div>
          </div>
        )}

      {/* View Mode Toggle */}
      {files.length > 0 && (
        <div
          style={{
            display: "flex",
            gap: "10px",
            marginBottom: "20px",
            flexWrap: "wrap",
          }}
        >
          {[
            { key: "active", label: "Active", count: fileStats.active || 0 },
            {
              key: "archived",
              label: "Archived",
              count: fileStats.archived || 0,
            },
            { key: "deleted", label: "Deleted", count: fileStats.deleted || 0 },
            { key: "pending", label: "Pending", count: fileStats.pending || 0 },
            { key: "all", label: "All", count: fileStats.total || 0 },
          ].map(({ key, label, count }) => (
            <button
              key={key}
              onClick={() => handleViewModeChange(key)}
              style={{
                padding: "8px 16px",
                backgroundColor: viewMode === key ? "#007bff" : "#e9ecef",
                color: viewMode === key ? "white" : "#495057",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
                fontSize: "14px",
              }}
            >
              {label} ({count})
            </button>
          ))}
        </div>
      )}

      {/* Show Details Toggle */}
      {files.length > 0 && (
        <div style={{ marginBottom: "20px" }}>
          <label style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <input
              type="checkbox"
              checked={showDetails}
              onChange={(e) => setShowDetails(e.target.checked)}
            />
            Show detailed information (version, tombstone data)
          </label>
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
      {!isLoading && currentFiles.length === 0 && selectedCollectionId && (
        <div
          style={{
            textAlign: "center",
            padding: "60px 20px",
            backgroundColor: "#f8f9fa",
            borderRadius: "8px",
            border: "2px dashed #dee2e6",
          }}
        >
          <div style={{ fontSize: "64px", marginBottom: "20px" }}>
            {viewMode === "active" ? "📁" : getStateIcon(viewMode)}
          </div>
          <h3>No {viewMode} files in this collection</h3>
          <p style={{ color: "#666" }}>
            {viewMode === "active"
              ? "This collection doesn't have any active files yet."
              : `Switch to "Active" view to see available files, or use "All" to see files in all states.`}
          </p>
        </div>
      )}

      {/* Files List */}
      {!isLoading && currentFiles.length > 0 && (
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
                      selectedFiles.size === currentFiles.length &&
                      currentFiles.length > 0
                    }
                    onChange={(e) => {
                      if (e.target.checked) {
                        setSelectedFiles(
                          new Set(currentFiles.map((f) => f.id)),
                        );
                      } else {
                        setSelectedFiles(new Set());
                      }
                    }}
                  />
                </th>
                <th style={{ padding: "12px", textAlign: "left" }}>File</th>
                <th style={{ padding: "12px", textAlign: "left" }}>Size</th>
                <th style={{ padding: "12px", textAlign: "left" }}>State</th>
                {showDetails && (
                  <>
                    <th style={{ padding: "12px", textAlign: "left" }}>
                      Version
                    </th>
                    <th style={{ padding: "12px", textAlign: "left" }}>
                      Modified
                    </th>
                    <th style={{ padding: "12px", textAlign: "left" }}>
                      Tombstone
                    </th>
                  </>
                )}
                <th style={{ padding: "12px", textAlign: "center" }}>
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {currentFiles.map((file) => {
                const versionInfo = fileManager?.getFileVersionInfo(file) || {};
                return (
                  <tr
                    key={file.id}
                    style={{ borderBottom: "1px solid #dee2e6" }}
                  >
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
                      <span
                        style={{
                          display: "inline-flex",
                          alignItems: "center",
                          gap: "4px",
                          padding: "4px 8px",
                          borderRadius: "4px",
                          backgroundColor: getStateColor(file.state),
                          color: "white",
                          fontSize: "12px",
                          fontWeight: "bold",
                        }}
                      >
                        {getStateIcon(file.state)} {file.state.toUpperCase()}
                      </span>
                    </td>
                    {showDetails && (
                      <>
                        <td style={{ padding: "12px" }}>
                          v{versionInfo.currentVersion || file.version || 1}
                          {versionInfo.hasTombstone && (
                            <div style={{ fontSize: "10px", color: "#999" }}>
                              Tombstone: v{versionInfo.tombstoneVersion}
                            </div>
                          )}
                        </td>
                        <td style={{ padding: "12px" }}>
                          {file.modified_at
                            ? formatDateTime(file.modified_at)
                            : file.created_at
                              ? formatDate(file.created_at)
                              : "Unknown"}
                        </td>
                        <td style={{ padding: "12px" }}>
                          {versionInfo.hasTombstone ? (
                            <div style={{ fontSize: "11px" }}>
                              <div>
                                Expires:{" "}
                                {formatDateTime(versionInfo.tombstoneExpiry)}
                              </div>
                              {versionInfo.isExpired && (
                                <div style={{ color: "#dc3545" }}>
                                  ⚠️ Expired
                                </div>
                              )}
                              {versionInfo.canRestore && (
                                <div style={{ color: "#28a745" }}>
                                  ✅ Restorable
                                </div>
                              )}
                            </div>
                          ) : (
                            <span style={{ color: "#999" }}>N/A</span>
                          )}
                        </td>
                      </>
                    )}
                    <td style={{ padding: "12px", textAlign: "center" }}>
                      <div
                        style={{
                          display: "flex",
                          gap: "4px",
                          flexWrap: "wrap",
                          justifyContent: "center",
                        }}
                      >
                        {fileManager?.canDownloadFile(file) && (
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
                                ? "File key not available - refresh page"
                                : "Download file"
                            }
                          >
                            {downloadingFiles.has(file.id)
                              ? "Downloading..."
                              : "Download"}
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })}
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
          <h3>📋 File Event Log ({eventLog.length})</h3>
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
              No file events logged yet.
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

export default ListFileManagerExample;
