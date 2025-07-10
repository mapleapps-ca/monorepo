// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/File/DownloadFileManagerExample.jsx
// Example component demonstrating how to use the DownloadFileManager

import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useFiles } from "../../../../services/Services";

const DownloadFileManagerExample = () => {
  const navigate = useNavigate();
  const { authService, getCollectionManager, listCollectionManager } =
    useFiles();

  // State management
  const [downloadManager, setDownloadManager] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [selectedFileId, setSelectedFileId] = useState("");
  const [downloadHistory, setDownloadHistory] = useState([]);
  const [activeDownloads, setActiveDownloads] = useState([]);
  const [downloadStats, setDownloadStats] = useState({});
  const [eventLog, setEventLog] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Download options
  const [downloadOptions, setDownloadOptions] = useState({
    saveToDisk: true,
    forceRefresh: false,
    urlDuration: null,
  });

  // Available file IDs for testing
  const [availableFiles, setAvailableFiles] = useState([]);
  const [isLoadingFiles, setIsLoadingFiles] = useState(false);

  // Batch download
  const [selectedFilesForBatch, setSelectedFilesForBatch] = useState([]);
  const [batchDownloadProgress, setBatchDownloadProgress] = useState(null);

  // Initialize download manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authService.isAuthenticated()) return;

      try {
        const { default: DownloadFileManager } = await import(
          "../../../../services/Manager/File/DownloadFileManager.js"
        );

        // Pass collection managers to the DownloadFileManager
        const manager = new DownloadFileManager(
          authService,
          null, // getFileManager - we'll use API directly for this example
          getCollectionManager,
        );
        await manager.initialize();

        setDownloadManager(manager);

        // Set up event listener
        manager.addDownloadListener(handleDownloadEvent);

        console.log(
          "[Example] DownloadFileManager initialized with collection managers",
        );
      } catch (err) {
        console.error(
          "[Example] Failed to initialize DownloadFileManager:",
          err,
        );
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();

    return () => {
      if (downloadManager) {
        downloadManager.removeDownloadListener(handleDownloadEvent);
      }
    };
  }, [authService, getCollectionManager]);

  // Load manager status when manager is ready
  useEffect(() => {
    if (downloadManager) {
      loadManagerStatus();
      loadDownloadHistory();
      loadDownloadStats();
    }
  }, [downloadManager]);

  // Update active downloads periodically
  useEffect(() => {
    if (downloadManager) {
      const interval = setInterval(() => {
        const active = downloadManager.getActiveDownloads();
        setActiveDownloads(active);
      }, 1000);

      return () => clearInterval(interval);
    }
  }, [downloadManager]);

  // Load some available files for testing
  useEffect(() => {
    if (downloadManager && listCollectionManager) {
      loadAvailableFiles();
    }
  }, [downloadManager, listCollectionManager]);

  // Load available files from collections
  const loadAvailableFiles = async () => {
    setIsLoadingFiles(true);
    try {
      console.log("[Example] Loading available files for download testing...");

      // Get collections first
      const result = await listCollectionManager.listCollections(false);

      if (result.collections && result.collections.length > 0) {
        // Load files from the first few collections
        const filesToLoad = [];

        for (let i = 0; i < Math.min(3, result.collections.length); i++) {
          const collection = result.collections[i];
          try {
            // Use the ListFileManager to get files for this collection
            const { default: ListFileManager } = await import(
              "../../../../services/Manager/File/ListFileManager.js"
            );

            const listManager = new ListFileManager(
              authService,
              getCollectionManager,
              listCollectionManager,
            );
            await listManager.initialize();

            const files = await listManager.listFilesByCollection(
              collection.id,
              ["active", "archived"], // Include active and archived files
              false,
            );

            files.forEach((file) => {
              filesToLoad.push({
                id: file.id,
                name: file.name || "[Encrypted]",
                collectionName: collection.name || "[Encrypted]",
                state: file.state,
                version: file.version,
                size: file.size || file.encrypted_file_size_in_bytes || 0,
                mimeType: file.mime_type,
              });
            });
          } catch (fileError) {
            console.warn(
              `[Example] Failed to load files from collection ${collection.id}:`,
              fileError,
            );
          }
        }

        setAvailableFiles(filesToLoad.slice(0, 15)); // Limit to first 15 files
        console.log("[Example] Available files loaded:", filesToLoad.length);
      } else {
        console.log("[Example] No collections found");
        setError(
          "No collections found. Please create collections and files first.",
        );
      }
    } catch (err) {
      console.error("[Example] Failed to load available files:", err);
      setError(`Failed to load available files: ${err.message}`);
    } finally {
      setIsLoadingFiles(false);
    }
  };

  // Handle download events
  const handleDownloadEvent = useCallback(
    (eventType, eventData) => {
      console.log("[Example] Download event:", eventType, eventData);
      addToEventLog(eventType, eventData);

      // Update active downloads when events occur
      if (downloadManager) {
        const active = downloadManager.getActiveDownloads();
        setActiveDownloads(active);
      }
    },
    [downloadManager],
  );

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!downloadManager) return;

    const status = downloadManager.getManagerStatus();
    setManagerStatus(status);
    console.log("[Example] Manager status:", status);
  }, [downloadManager]);

  // Load download history
  const loadDownloadHistory = useCallback(() => {
    if (!downloadManager) return;

    const history = downloadManager.getDownloadHistory(10);
    setDownloadHistory(history);
    console.log("[Example] Download history:", history);
  }, [downloadManager]);

  // Load download statistics
  const loadDownloadStats = useCallback(() => {
    if (!downloadManager) return;

    const stats = downloadManager.getDownloadStats();
    setDownloadStats(stats);
    console.log("[Example] Download stats:", stats);
  }, [downloadManager]);

  // Download single file
  const handleDownloadFile = async (forceRefresh = false) => {
    if (!downloadManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Starting file download:", selectedFileId);

      const result = await downloadManager.downloadFile(selectedFileId, {
        forceRefresh,
        saveToDisk: downloadOptions.saveToDisk,
        urlDuration: downloadOptions.urlDuration,
        onProgress: (progressData) => {
          console.log("[Example] Download progress:", progressData);
        },
      });

      setSuccess(`File downloaded successfully: ${result.fileName}`);

      // Refresh history and stats
      loadDownloadHistory();
      loadDownloadStats();

      addToEventLog("download_completed_ui", {
        fileId: selectedFileId,
        fileName: result.fileName,
        downloadId: result.downloadId,
      });
    } catch (err) {
      console.error("[Example] Failed to download file:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Download thumbnail only
  const handleDownloadThumbnail = async () => {
    if (!downloadManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      console.log("[Example] Downloading thumbnail:", selectedFileId);

      const result = await downloadManager.downloadThumbnail(selectedFileId, {
        forceRefresh: downloadOptions.forceRefresh,
      });

      // Create object URL and display thumbnail
      const thumbnailUrl = URL.createObjectURL(result.thumbnail);

      // Open thumbnail in new window/tab
      window.open(thumbnailUrl, "_blank");

      setSuccess("Thumbnail downloaded and opened");

      addToEventLog("thumbnail_downloaded", {
        fileId: selectedFileId,
        thumbnailSize: result.thumbnail.size,
      });
    } catch (err) {
      console.error("[Example] Failed to download thumbnail:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Batch download files
  const handleBatchDownload = async () => {
    if (!downloadManager || selectedFilesForBatch.length === 0) {
      setError("Please select files for batch download");
      return;
    }

    setIsLoading(true);
    setError("");
    setBatchDownloadProgress({
      total: selectedFilesForBatch.length,
      completed: 0,
      current: null,
      files: [],
    });

    try {
      console.log("[Example] Starting batch download:", selectedFilesForBatch);

      const result = await downloadManager.downloadMultipleFiles(
        selectedFilesForBatch,
        {
          concurrent: 2, // Download 2 files at a time
          saveToDisk: downloadOptions.saveToDisk,
          onProgress: (progressData) => {
            console.log("[Example] Batch progress:", progressData);
            if (progressData.batchProgress) {
              setBatchDownloadProgress((prev) => ({
                ...prev,
                completed: progressData.batchProgress.completed,
                current: progressData.batchProgress.current,
              }));
            }
          },
          onFileComplete: (fileResult) => {
            console.log("[Example] File completed:", fileResult);
            setBatchDownloadProgress((prev) => ({
              ...prev,
              files: [...prev.files, fileResult],
            }));
          },
        },
      );

      setSuccess(
        `Batch download completed: ${result.successful.length} successful, ${result.failed.length} failed`,
      );

      // Refresh history and stats
      loadDownloadHistory();
      loadDownloadStats();

      addToEventLog("batch_download_completed", {
        successful: result.successful.length,
        failed: result.failed.length,
        total: result.total,
      });
    } catch (err) {
      console.error("[Example] Batch download failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
      setBatchDownloadProgress(null);
      setSelectedFilesForBatch([]);
    }
  };

  // Clear caches
  const handleClearCaches = () => {
    if (!downloadManager) return;

    if (selectedFileId) {
      downloadManager.clearFileDownloadCache(selectedFileId);
      setSuccess("File download cache cleared");
    } else {
      downloadManager.clearAllDownloadCaches();
      setSuccess("All download caches cleared");
    }

    addToEventLog("caches_cleared", {
      fileId: selectedFileId || "all",
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

  // Format date time
  const formatDateTime = (dateString) => {
    if (!dateString) return "N/A";
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return "Invalid Date";
    }
  };

  // Get state color
  const getStateColor = (state) => {
    switch (state) {
      case "active":
        return "#28a745";
      case "archived":
        return "#6c757d";
      case "deleted":
        return "#dc3545";
      case "pending":
        return "#ffc107";
      default:
        return "#6c757d";
    }
  };

  // Get download state color
  const getDownloadStateColor = (state) => {
    if (!downloadManager) return "#6c757d";

    switch (state) {
      case downloadManager.DOWNLOAD_STATES.PREPARING:
        return "#17a2b8";
      case downloadManager.DOWNLOAD_STATES.DOWNLOADING:
        return "#007bff";
      case downloadManager.DOWNLOAD_STATES.DECRYPTING:
        return "#6f42c1";
      case downloadManager.DOWNLOAD_STATES.COMPLETED:
        return "#28a745";
      case downloadManager.DOWNLOAD_STATES.FAILED:
        return "#dc3545";
      case downloadManager.DOWNLOAD_STATES.CANCELLED:
        return "#6c757d";
      default:
        return "#6c757d";
    }
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

  // Toggle file selection for batch download
  const toggleFileForBatch = (fileId) => {
    setSelectedFilesForBatch((prev) => {
      if (prev.includes(fileId)) {
        return prev.filter((id) => id !== fileId);
      } else {
        return [...prev, fileId];
      }
    });
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

      <h2>📥 Download File Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This example demonstrates the complete file download workflow with E2EE
        decryption.
        <br />
        <strong>Features:</strong> Single file download, batch download,
        thumbnail download, progress tracking, and caching
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
            <strong>Can Download:</strong>{" "}
            {managerStatus.canDownloadFiles ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Available Files:</strong> {availableFiles.length}
          </div>
          <div>
            <strong>Active Downloads:</strong> {activeDownloads.length}
          </div>
          <div>
            <strong>Listeners:</strong> {managerStatus.listenerCount || 0}
          </div>
          <div>
            <strong>Downloads Today:</strong>{" "}
            {downloadStats.totalDownloadsToday || 0}
          </div>
          <div>
            <strong>Manager Ready:</strong>{" "}
            {downloadManager ? "✅ Yes" : "❌ No"}
          </div>
        </div>
      </div>

      {/* Download Options */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f4f8",
          borderRadius: "6px",
          border: "1px solid #b3d9e8",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>⚙️ Download Options:</h4>
        <div
          style={{
            display: "flex",
            gap: "20px",
            flexWrap: "wrap",
            alignItems: "center",
          }}
        >
          <label style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <input
              type="checkbox"
              checked={downloadOptions.saveToDisk}
              onChange={(e) =>
                setDownloadOptions((prev) => ({
                  ...prev,
                  saveToDisk: e.target.checked,
                }))
              }
            />
            Save to Disk
          </label>
          <label style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <input
              type="checkbox"
              checked={downloadOptions.forceRefresh}
              onChange={(e) =>
                setDownloadOptions((prev) => ({
                  ...prev,
                  forceRefresh: e.target.checked,
                }))
              }
            />
            Force Refresh URLs
          </label>
          <label style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            URL Duration (seconds):
            <input
              type="number"
              value={downloadOptions.urlDuration || ""}
              onChange={(e) =>
                setDownloadOptions((prev) => ({
                  ...prev,
                  urlDuration: e.target.value ? parseInt(e.target.value) : null,
                }))
              }
              placeholder="Default"
              style={{ width: "100px", padding: "4px" }}
            />
          </label>
        </div>
      </div>

      {/* File Selection */}
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
            📄 Select File for Download ({availableFiles.length} available):
          </h4>
          <button
            onClick={loadAvailableFiles}
            disabled={isLoadingFiles}
            style={{
              padding: "5px 15px",
              backgroundColor: "#1976d2",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isLoadingFiles ? "not-allowed" : "pointer",
            }}
          >
            {isLoadingFiles ? "🔄 Loading..." : "🔄 Refresh Files"}
          </button>
        </div>

        <div style={{ marginBottom: "15px" }}>
          <select
            value={selectedFileId}
            onChange={(e) => setSelectedFileId(e.target.value)}
            disabled={availableFiles.length === 0}
            style={{
              width: "100%",
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              backgroundColor:
                availableFiles.length === 0 ? "#f5f5f5" : "white",
            }}
          >
            <option value="">
              {availableFiles.length === 0
                ? "No files available - click refresh"
                : "Select a file to download..."}
            </option>
            {availableFiles.map((file) => (
              <option key={file.id} value={file.id}>
                {file.name} ({formatFileSize(file.size)}, {file.state}) -{" "}
                {file.collectionName} - {file.id.substring(0, 8)}...
              </option>
            ))}
          </select>
        </div>

        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          <button
            onClick={() => handleDownloadFile(false)}
            disabled={!selectedFileId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !selectedFileId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            {isLoading ? "🔄 Downloading..." : "📥 Download File"}
          </button>
          <button
            onClick={() => handleDownloadFile(true)}
            disabled={!selectedFileId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#ffc107",
              color: "#212529",
              border: "none",
              borderRadius: "4px",
              cursor: !selectedFileId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            🔄 Force Refresh & Download
          </button>
          <button
            onClick={handleDownloadThumbnail}
            disabled={!selectedFileId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#17a2b8",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !selectedFileId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            🖼️ Download Thumbnail
          </button>
          <button
            onClick={handleClearCaches}
            disabled={!downloadManager}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !downloadManager ? "not-allowed" : "pointer",
            }}
          >
            🗑️ Clear Cache
          </button>
        </div>
      </div>

      {/* Batch Download Section */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#fff3cd",
          borderRadius: "6px",
          border: "1px solid #ffecb5",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>
          📦 Batch Download ({selectedFilesForBatch.length} selected):
        </h4>

        <div
          style={{ marginBottom: "15px", maxHeight: "200px", overflow: "auto" }}
        >
          {availableFiles.map((file) => (
            <label
              key={file.id}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                padding: "4px 0",
                cursor: "pointer",
              }}
            >
              <input
                type="checkbox"
                checked={selectedFilesForBatch.includes(file.id)}
                onChange={() => toggleFileForBatch(file.id)}
              />
              <span style={{ flex: 1 }}>
                {file.name} ({formatFileSize(file.size)})
              </span>
              <span
                style={{
                  padding: "2px 6px",
                  borderRadius: "3px",
                  backgroundColor: getStateColor(file.state),
                  color: "white",
                  fontSize: "11px",
                }}
              >
                {file.state?.toUpperCase()}
              </span>
            </label>
          ))}
        </div>

        <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
          <button
            onClick={handleBatchDownload}
            disabled={selectedFilesForBatch.length === 0 || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#e83e8c",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                selectedFilesForBatch.length === 0 || isLoading
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            📦 Download Selected ({selectedFilesForBatch.length})
          </button>
          <button
            onClick={() => setSelectedFilesForBatch([])}
            disabled={selectedFilesForBatch.length === 0}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                selectedFilesForBatch.length === 0 ? "not-allowed" : "pointer",
            }}
          >
            Clear Selection
          </button>
          <button
            onClick={() =>
              setSelectedFilesForBatch(availableFiles.map((f) => f.id))
            }
            disabled={availableFiles.length === 0}
            style={{
              padding: "8px 16px",
              backgroundColor: "#fd7e14",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: availableFiles.length === 0 ? "not-allowed" : "pointer",
            }}
          >
            Select All
          </button>
        </div>

        {/* Batch Download Progress */}
        {batchDownloadProgress && (
          <div
            style={{
              marginTop: "15px",
              padding: "10px",
              backgroundColor: "#f8f9fa",
              borderRadius: "4px",
            }}
          >
            <div style={{ marginBottom: "5px" }}>
              <strong>Batch Progress:</strong> {batchDownloadProgress.completed}{" "}
              / {batchDownloadProgress.total} completed
            </div>
            {batchDownloadProgress.current && (
              <div style={{ fontSize: "14px", color: "#666" }}>
                Currently downloading:{" "}
                {batchDownloadProgress.current.substring(0, 8)}...
              </div>
            )}
            <div
              style={{
                backgroundColor: "#e0e0e0",
                borderRadius: "4px",
                height: "8px",
                marginTop: "8px",
              }}
            >
              <div
                style={{
                  backgroundColor: "#007bff",
                  height: "100%",
                  borderRadius: "4px",
                  width: `${(batchDownloadProgress.completed / batchDownloadProgress.total) * 100}%`,
                  transition: "width 0.3s ease",
                }}
              />
            </div>
          </div>
        )}
      </div>

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

      {/* Active Downloads */}
      {activeDownloads.length > 0 && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>
            🔄 Active Downloads ({activeDownloads.length}):
          </h4>
          {activeDownloads.map((download) => (
            <div
              key={download.fileId}
              style={{
                marginBottom: "10px",
                padding: "10px",
                backgroundColor: "#f8f9fa",
                borderRadius: "4px",
                border: "1px solid #e9ecef",
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  marginBottom: "5px",
                }}
              >
                <span>
                  <strong>File:</strong> {download.fileId.substring(0, 12)}...
                </span>
                <span
                  style={{
                    padding: "2px 8px",
                    borderRadius: "3px",
                    backgroundColor: getDownloadStateColor(download.state),
                    color: "white",
                    fontSize: "12px",
                  }}
                >
                  {download.state?.toUpperCase()}
                </span>
              </div>
              <div
                style={{
                  backgroundColor: "#e0e0e0",
                  borderRadius: "4px",
                  height: "8px",
                  marginTop: "5px",
                }}
              >
                <div
                  style={{
                    backgroundColor: getDownloadStateColor(download.state),
                    height: "100%",
                    borderRadius: "4px",
                    width: `${download.progress}%`,
                    transition: "width 0.3s ease",
                  }}
                />
              </div>
              <div
                style={{ fontSize: "12px", color: "#666", marginTop: "2px" }}
              >
                Progress: {download.progress?.toFixed(1)}%
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Download History */}
      {downloadHistory.length > 0 && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>
            📋 Recent Downloads ({downloadHistory.length}):
          </h4>
          <div style={{ maxHeight: "200px", overflow: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr style={{ backgroundColor: "#f8f9fa" }}>
                  <th
                    style={{
                      padding: "8px",
                      textAlign: "left",
                      border: "1px solid #dee2e6",
                    }}
                  >
                    File Name
                  </th>
                  <th
                    style={{
                      padding: "8px",
                      textAlign: "left",
                      border: "1px solid #dee2e6",
                    }}
                  >
                    Downloaded At
                  </th>
                  <th
                    style={{
                      padding: "8px",
                      textAlign: "left",
                      border: "1px solid #dee2e6",
                    }}
                  >
                    File ID
                  </th>
                </tr>
              </thead>
              <tbody>
                {downloadHistory.map((record, index) => (
                  <tr key={`${record.fileId}-${index}`}>
                    <td style={{ padding: "8px", border: "1px solid #dee2e6" }}>
                      {record.fileName}
                    </td>
                    <td style={{ padding: "8px", border: "1px solid #dee2e6" }}>
                      {formatDateTime(record.downloadedAt)}
                    </td>
                    <td
                      style={{
                        padding: "8px",
                        border: "1px solid #dee2e6",
                        fontFamily: "monospace",
                      }}
                    >
                      {record.fileId.substring(0, 8)}...
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
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
          <h3>📋 Download Event Log ({eventLog.length})</h3>
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
              No download events logged yet.
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

export default DownloadFileManagerExample;
