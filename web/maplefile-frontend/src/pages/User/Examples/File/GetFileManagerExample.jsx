// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/File/GetFileManagerExample.jsx
// Example component demonstrating how to use the GetFileManager

import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../../hooks/useService.jsx";

const GetFileManagerExample = () => {
  const navigate = useNavigate();
  const { authService, getCollectionManager, listCollectionManager } =
    useServices();

  // State management
  const [fileManager, setFileManager] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [selectedFileId, setSelectedFileId] = useState("");
  const [fileDetails, setFileDetails] = useState(null);
  const [versionHistory, setVersionHistory] = useState([]);
  const [filePermissions, setFilePermissions] = useState(null);
  const [fileStats, setFileStats] = useState(null);
  const [completeFileData, setCompleteFileData] = useState(null);
  const [showVersionHistory, setShowVersionHistory] = useState(false);
  const [showPermissions, setShowPermissions] = useState(false);
  const [showStats, setShowStats] = useState(false);
  const [showCompleteData, setShowCompleteData] = useState(false);
  const [eventLog, setEventLog] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Available file IDs for testing (these would come from a file list in real usage)
  const [availableFiles, setAvailableFiles] = useState([]);
  const [isLoadingFiles, setIsLoadingFiles] = useState(false);

  // Initialize file manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authService.isAuthenticated()) return;

      try {
        const { default: GetFileManager } = await import(
          "../../../../services/Manager/File/GetFileManager.js"
        );

        // Pass collection managers to the GetFileManager
        const manager = new GetFileManager(
          authService,
          getCollectionManager,
          listCollectionManager,
        );
        await manager.initialize();

        setFileManager(manager);

        // Set up event listener
        manager.addFileRetrievalListener(handleFileEvent);

        console.log(
          "[Example] GetFileManager initialized with collection managers",
        );
      } catch (err) {
        console.error("[Example] Failed to initialize GetFileManager:", err);
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();

    return () => {
      if (fileManager) {
        fileManager.removeFileRetrievalListener(handleFileEvent);
      }
    };
  }, [authService, getCollectionManager, listCollectionManager]);

  // Load manager status when manager is ready
  useEffect(() => {
    if (fileManager) {
      loadManagerStatus();
    }
  }, [fileManager]);

  // Load some available files for testing
  useEffect(() => {
    if (fileManager && listCollectionManager) {
      loadAvailableFiles();
    }
  }, [fileManager, listCollectionManager]);

  // Load available files from collections
  const loadAvailableFiles = async () => {
    setIsLoadingFiles(true);
    try {
      console.log("[Example] Loading available files for testing...");

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
              ["active"],
              false,
            );

            files.forEach((file) => {
              filesToLoad.push({
                id: file.id,
                name: file.name || "[Encrypted]",
                collectionName: collection.name || "[Encrypted]",
                state: file.state,
                version: file.version,
              });
            });
          } catch (fileError) {
            console.warn(
              `[Example] Failed to load files from collection ${collection.id}:`,
              fileError,
            );
          }
        }

        setAvailableFiles(filesToLoad.slice(0, 10)); // Limit to first 10 files
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

  // Get file details
  const handleGetFileDetails = async (forceRefresh = false) => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Getting file details for:", selectedFileId);

      const file = await fileManager.getFileById(
        selectedFileId,
        forceRefresh,
        false,
      );

      setFileDetails(file);
      setSuccess(`File details loaded: ${file.name || "[Unable to decrypt]"}`);

      addToEventLog("file_details_loaded", {
        fileId: selectedFileId,
        fileName: file.name,
        isDecrypted: file._isDecrypted,
        forceRefresh,
      });
    } catch (err) {
      console.error("[Example] Failed to get file details:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Get file version history
  const handleGetVersionHistory = async (forceRefresh = false) => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      console.log("[Example] Getting version history for:", selectedFileId);

      const versions = await fileManager.getFileVersionHistory(
        selectedFileId,
        forceRefresh,
      );

      setVersionHistory(versions);
      setShowVersionHistory(true);
      setSuccess(`Version history loaded: ${versions.length} versions`);

      addToEventLog("version_history_loaded", {
        fileId: selectedFileId,
        versionCount: versions.length,
        forceRefresh,
      });
    } catch (err) {
      console.error("[Example] Failed to get version history:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Get file permissions
  const handleGetPermissions = async (forceRefresh = false) => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      console.log("[Example] Getting file permissions for:", selectedFileId);

      const permissions = await fileManager.getFilePermissions(
        selectedFileId,
        forceRefresh,
      );

      setFilePermissions(permissions);
      setShowPermissions(true);
      setSuccess(`File permissions loaded`);

      addToEventLog("permissions_loaded", {
        fileId: selectedFileId,
        forceRefresh,
      });
    } catch (err) {
      console.error("[Example] Failed to get file permissions:", err);
      setError(err.message);
      // This is expected to fail if the endpoint doesn't exist
      if (err.message.includes("404") || err.message.includes("Not Found")) {
        setError("File permissions endpoint not implemented in backend");
      }
    } finally {
      setIsLoading(false);
    }
  };

  // Get file statistics
  const handleGetStats = async (forceRefresh = false) => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      console.log("[Example] Getting file statistics for:", selectedFileId);

      const stats = await fileManager.getFileStats(
        selectedFileId,
        forceRefresh,
      );

      setFileStats(stats);
      setShowStats(true);
      setSuccess(`File statistics loaded`);

      addToEventLog("stats_loaded", {
        fileId: selectedFileId,
        forceRefresh,
      });
    } catch (err) {
      console.error("[Example] Failed to get file statistics:", err);
      setError(err.message);
      // This is expected to fail if the endpoint doesn't exist
      if (err.message.includes("404") || err.message.includes("Not Found")) {
        setError("File statistics endpoint not implemented in backend");
      }
    } finally {
      setIsLoading(false);
    }
  };

  // Get complete file data
  const handleGetCompleteData = async (forceRefresh = false) => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      console.log("[Example] Getting complete file data for:", selectedFileId);

      const completeData = await fileManager.getFileComplete(
        selectedFileId,
        forceRefresh,
      );

      setCompleteFileData(completeData);
      setShowCompleteData(true);
      setSuccess(
        `Complete file data loaded with ${completeData.errors.length} errors`,
      );

      addToEventLog("complete_data_loaded", {
        fileId: selectedFileId,
        hasFile: !!completeData.file,
        versionCount: completeData.versionHistory.length,
        hasPermissions: !!completeData.permissions,
        hasStats: !!completeData.stats,
        errorCount: completeData.errors.length,
        forceRefresh,
      });
    } catch (err) {
      console.error("[Example] Failed to get complete file data:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Check file existence
  const handleCheckFileExists = async () => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      console.log("[Example] Checking file existence for:", selectedFileId);

      const existsData = await fileManager.checkFileExists(selectedFileId);

      setSuccess(
        `File exists: ${existsData.exists}, accessible: ${existsData.accessible}`,
      );

      addToEventLog("file_existence_checked", {
        fileId: selectedFileId,
        exists: existsData.exists,
        accessible: existsData.accessible,
      });
    } catch (err) {
      console.error("[Example] Failed to check file existence:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Clear caches
  const handleClearCaches = () => {
    if (!fileManager) return;

    if (selectedFileId) {
      fileManager.clearFileCache(selectedFileId);
      setSuccess("File cache cleared");
    } else {
      fileManager.clearAllCaches();
      setSuccess("All caches cleared");
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

      <h2>📄 Get File Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This example demonstrates the complete file retrieval workflow with E2EE
        decryption.
        <br />
        <strong>Features:</strong> Get file details, version history,
        permissions, statistics, and complete file data
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
            <strong>Can Get Files:</strong>{" "}
            {managerStatus.canGetFiles ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Available Files:</strong> {availableFiles.length}
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
          <div>
            <strong>File Manager:</strong>{" "}
            {fileManager ? "✅ Ready" : "❌ Not Ready"}
          </div>
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
            📄 Select File ({availableFiles.length} available):
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
                : "Select a file to get details..."}
            </option>
            {availableFiles.map((file) => (
              <option key={file.id} value={file.id}>
                {file.name} (v{file.version}, {file.state}) -{" "}
                {file.collectionName} - {file.id.substring(0, 8)}...
              </option>
            ))}
          </select>
        </div>

        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          <button
            onClick={() => handleGetFileDetails(false)}
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
            {isLoading ? "🔄 Loading..." : "📄 Get File Details"}
          </button>
          <button
            onClick={() => handleGetFileDetails(true)}
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
            🔄 Force Refresh
          </button>
          <button
            onClick={() => handleGetVersionHistory(false)}
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
            📋 Version History
          </button>
          <button
            onClick={() => handleGetPermissions(false)}
            disabled={!selectedFileId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6f42c1",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !selectedFileId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            🔒 Permissions
          </button>
          <button
            onClick={() => handleGetStats(false)}
            disabled={!selectedFileId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#fd7e14",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !selectedFileId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            📊 Statistics
          </button>
          <button
            onClick={() => handleGetCompleteData(false)}
            disabled={!selectedFileId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#e83e8c",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !selectedFileId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            🎯 Complete Data
          </button>
          <button
            onClick={() => handleCheckFileExists()}
            disabled={!selectedFileId || isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#20c997",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !selectedFileId || isLoading ? "not-allowed" : "pointer",
            }}
          >
            ❓ Check Exists
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

      {/* File Details */}
      {fileDetails && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>📄 File Details:</h4>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
              gap: "10px",
            }}
          >
            <div>
              <strong>ID:</strong> {fileDetails.id}
            </div>
            <div>
              <strong>Name:</strong> {fileDetails.name || "[Unable to decrypt]"}
            </div>
            <div>
              <strong>MIME Type:</strong> {fileDetails.mime_type || "Unknown"}
            </div>
            <div>
              <strong>Size:</strong>{" "}
              {fileDetails.size ? formatFileSize(fileDetails.size) : "Unknown"}
            </div>
            <div>
              <strong>Encrypted Size:</strong>{" "}
              {fileDetails.encrypted_file_size_in_bytes
                ? formatFileSize(fileDetails.encrypted_file_size_in_bytes)
                : "Unknown"}
            </div>
            <div>
              <strong>State:</strong>
              <span
                style={{
                  marginLeft: "5px",
                  padding: "2px 6px",
                  borderRadius: "3px",
                  backgroundColor: getStateColor(fileDetails.state),
                  color: "white",
                  fontSize: "12px",
                }}
              >
                {fileDetails.state?.toUpperCase()}
              </span>
            </div>
            <div>
              <strong>Version:</strong> v{fileDetails.version}
            </div>
            <div>
              <strong>Collection ID:</strong> {fileDetails.collection_id}
            </div>
            <div>
              <strong>Created:</strong> {formatDateTime(fileDetails.created_at)}
            </div>
            <div>
              <strong>Modified:</strong>{" "}
              {formatDateTime(fileDetails.modified_at)}
            </div>
            <div>
              <strong>Decrypted:</strong>{" "}
              {fileDetails._isDecrypted ? "✅ Yes" : "❌ No"}
            </div>
            <div>
              <strong>Has Tombstone:</strong>{" "}
              {fileDetails._has_tombstone ? "✅ Yes" : "❌ No"}
            </div>
            {fileDetails._has_tombstone && (
              <>
                <div>
                  <strong>Tombstone Version:</strong> v
                  {fileDetails.tombstone_version}
                </div>
                <div>
                  <strong>Tombstone Expiry:</strong>{" "}
                  {formatDateTime(fileDetails.tombstone_expiry)}
                </div>
                <div>
                  <strong>Tombstone Expired:</strong>{" "}
                  {fileDetails._tombstone_expired ? "✅ Yes" : "❌ No"}
                </div>
              </>
            )}
            {fileDetails._decryptionError && (
              <div style={{ color: "#dc3545", gridColumn: "1 / -1" }}>
                <strong>Decryption Error:</strong>{" "}
                {fileDetails._decryptionError}
              </div>
            )}
          </div>
        </div>
      )}

      {/* Version History */}
      {showVersionHistory && versionHistory.length > 0 && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
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
            <h4 style={{ margin: 0 }}>
              📋 Version History ({versionHistory.length} versions):
            </h4>
            <button
              onClick={() => setShowVersionHistory(false)}
              style={{
                padding: "5px 10px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
              }}
            >
              Hide
            </button>
          </div>
          <div style={{ maxHeight: "300px", overflow: "auto" }}>
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
                    Version
                  </th>
                  <th
                    style={{
                      padding: "8px",
                      textAlign: "left",
                      border: "1px solid #dee2e6",
                    }}
                  >
                    Name
                  </th>
                  <th
                    style={{
                      padding: "8px",
                      textAlign: "left",
                      border: "1px solid #dee2e6",
                    }}
                  >
                    State
                  </th>
                  <th
                    style={{
                      padding: "8px",
                      textAlign: "left",
                      border: "1px solid #dee2e6",
                    }}
                  >
                    Size
                  </th>
                  <th
                    style={{
                      padding: "8px",
                      textAlign: "left",
                      border: "1px solid #dee2e6",
                    }}
                  >
                    Modified
                  </th>
                  <th
                    style={{
                      padding: "8px",
                      textAlign: "left",
                      border: "1px solid #dee2e6",
                    }}
                  >
                    Decrypted
                  </th>
                </tr>
              </thead>
              <tbody>
                {versionHistory.map((version, index) => (
                  <tr key={`${version.id}-${version.version}`}>
                    <td style={{ padding: "8px", border: "1px solid #dee2e6" }}>
                      v{version.version}
                    </td>
                    <td style={{ padding: "8px", border: "1px solid #dee2e6" }}>
                      {version.name || "[Unable to decrypt]"}
                    </td>
                    <td style={{ padding: "8px", border: "1px solid #dee2e6" }}>
                      <span
                        style={{
                          padding: "2px 6px",
                          borderRadius: "3px",
                          backgroundColor: getStateColor(version.state),
                          color: "white",
                          fontSize: "11px",
                        }}
                      >
                        {version.state?.toUpperCase()}
                      </span>
                    </td>
                    <td style={{ padding: "8px", border: "1px solid #dee2e6" }}>
                      {version.size ? formatFileSize(version.size) : "Unknown"}
                    </td>
                    <td style={{ padding: "8px", border: "1px solid #dee2e6" }}>
                      {formatDateTime(version.modified_at)}
                    </td>
                    <td style={{ padding: "8px", border: "1px solid #dee2e6" }}>
                      {version._isDecrypted ? "✅" : "❌"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* File Permissions */}
      {showPermissions && filePermissions && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
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
            <h4 style={{ margin: 0 }}>🔒 File Permissions:</h4>
            <button
              onClick={() => setShowPermissions(false)}
              style={{
                padding: "5px 10px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
              }}
            >
              Hide
            </button>
          </div>
          <pre
            style={{
              backgroundColor: "#f8f9fa",
              padding: "10px",
              borderRadius: "4px",
              overflow: "auto",
            }}
          >
            {JSON.stringify(filePermissions, null, 2)}
          </pre>
        </div>
      )}

      {/* File Statistics */}
      {showStats && fileStats && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
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
            <h4 style={{ margin: 0 }}>📊 File Statistics:</h4>
            <button
              onClick={() => setShowStats(false)}
              style={{
                padding: "5px 10px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
              }}
            >
              Hide
            </button>
          </div>
          <pre
            style={{
              backgroundColor: "#f8f9fa",
              padding: "10px",
              borderRadius: "4px",
              overflow: "auto",
            }}
          >
            {JSON.stringify(fileStats, null, 2)}
          </pre>
        </div>
      )}

      {/* Complete File Data */}
      {showCompleteData && completeFileData && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
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
            <h4 style={{ margin: 0 }}>🎯 Complete File Data:</h4>
            <button
              onClick={() => setShowCompleteData(false)}
              style={{
                padding: "5px 10px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
              }}
            >
              Hide
            </button>
          </div>

          <div style={{ marginBottom: "15px" }}>
            <h5>Summary:</h5>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
                gap: "10px",
              }}
            >
              <div>
                <strong>Has File:</strong>{" "}
                {completeFileData.file ? "✅ Yes" : "❌ No"}
              </div>
              <div>
                <strong>Version Count:</strong>{" "}
                {completeFileData.versionHistory.length}
              </div>
              <div>
                <strong>Has Permissions:</strong>{" "}
                {completeFileData.permissions ? "✅ Yes" : "❌ No"}
              </div>
              <div>
                <strong>Has Statistics:</strong>{" "}
                {completeFileData.stats ? "✅ Yes" : "❌ No"}
              </div>
              <div>
                <strong>Error Count:</strong> {completeFileData.errors.length}
              </div>
            </div>
          </div>

          {completeFileData.errors.length > 0 && (
            <div style={{ marginBottom: "15px" }}>
              <h5 style={{ color: "#dc3545" }}>Errors:</h5>
              {completeFileData.errors.map((error, index) => (
                <div
                  key={index}
                  style={{ color: "#dc3545", marginBottom: "5px" }}
                >
                  <strong>{error.type}:</strong> {error.error}
                </div>
              ))}
            </div>
          )}

          <details>
            <summary style={{ cursor: "pointer", fontWeight: "bold" }}>
              View Raw Data
            </summary>
            <pre
              style={{
                backgroundColor: "#f8f9fa",
                padding: "10px",
                borderRadius: "4px",
                overflow: "auto",
                marginTop: "10px",
              }}
            >
              {JSON.stringify(completeFileData, null, 2)}
            </pre>
          </details>
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

export default GetFileManagerExample;
