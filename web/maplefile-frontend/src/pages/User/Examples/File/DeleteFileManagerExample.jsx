// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/File/DeleteFileManagerExample.jsx
// Example component demonstrating how to use the DeleteFileManager

import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useFiles } from "../../../../services/Services";

const DeleteFileManagerExample = () => {
  const navigate = useNavigate();
  const { authService, getCollectionManager, listCollectionManager } =
    useFiles();

  // State management
  const [fileManager, setFileManager] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [selectedFileId, setSelectedFileId] = useState("");
  const [reason, setReason] = useState("");
  const [deletionResult, setDeletionResult] = useState(null);
  const [tombstoneData, setTombstoneData] = useState(null);
  const [deletionHistory, setDeletionHistory] = useState(null);
  const [extensionDays, setExtensionDays] = useState(30);
  const [eventLog, setEventLog] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Available file IDs for testing
  const [availableFiles, setAvailableFiles] = useState([]);
  const [isLoadingFiles, setIsLoadingFiles] = useState(false);

  // Operation flags
  const [operations, setOperations] = useState({
    deleting: false,
    restoring: false,
    archiving: false,
    unarchiving: false,
    permanentlyDeleting: false,
    extendingTombstone: false,
  });

  // Add event to log
  const addToEventLog = useCallback((eventType, eventData) => {
    setEventLog((prev) => [
      ...prev,
      {
        timestamp: new Date().toISOString(),
        eventType,
        eventData,
      },
    ]);
  }, []);

  // Handle file events
  const handleFileEvent = useCallback(
    (eventType, eventData) => {
      console.log("[Example] File event:", eventType, eventData);
      addToEventLog(eventType, eventData);
    },
    [addToEventLog],
  );

  // Load available files from collections
  const loadAvailableFiles = useCallback(
    async (manager = fileManager) => {
      setIsLoadingFiles(true);
      try {
        console.log("[Example] Loading available files for testing...");
        console.log("[Example] Using manager:", !!manager);

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
                ["active", "archived", "deleted"], // Include all relevant states
                false,
              );

              files.forEach((file) => {
                // Normalize file to ensure computed properties exist
                const normalizedFile = manager
                  ? manager.fileCryptoService?.normalizeFile(file) || file
                  : file;

                // Simple, direct capability checking
                const canDelete =
                  normalizedFile.state === "active" ||
                  normalizedFile.state === "archived";
                const canRestore = normalizedFile.state === "deleted";
                const canArchive = normalizedFile.state === "active";
                const canUnarchive = normalizedFile.state === "archived";
                const canPermanentlyDelete = false; // Only for expired tombstones

                console.log(`[Example] File ${file.id} capabilities:`, {
                  canDelete,
                  canRestore,
                  canArchive,
                  canUnarchive,
                  canPermanentlyDelete,
                  state: normalizedFile.state,
                });

                filesToLoad.push({
                  id: normalizedFile.id,
                  name: normalizedFile.name || "[Encrypted]",
                  collectionName: collection.name || "[Encrypted]",
                  state: normalizedFile.state,
                  version: normalizedFile.version,
                  canDelete,
                  canRestore,
                  canArchive,
                  canUnarchive,
                  canPermanentlyDelete,
                  hasTombstone:
                    normalizedFile._has_tombstone ||
                    normalizedFile.tombstone_version > 0,
                  tombstoneExpiry: normalizedFile.tombstone_expiry,
                  // Add computed properties for debugging
                  _is_active:
                    normalizedFile._is_active ||
                    normalizedFile.state === "active",
                  _is_archived:
                    normalizedFile._is_archived ||
                    normalizedFile.state === "archived",
                  _is_deleted:
                    normalizedFile._is_deleted ||
                    normalizedFile.state === "deleted",
                  _has_tombstone:
                    normalizedFile._has_tombstone ||
                    normalizedFile.tombstone_version > 0,
                  _tombstone_expired:
                    normalizedFile._tombstone_expired || false,
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
          console.log("[Example] Sample file capabilities:", filesToLoad[0]);
          console.log("[Example] DeleteFileManager ready:", !!manager);
          console.log(
            "[Example] DeleteFileManager has crypto service:",
            !!manager?.fileCryptoService,
          );
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
    },
    [
      authService,
      fileManager,
      getCollectionManager,
      listCollectionManager,
      setError,
    ],
  );

  // Initialize delete file manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authService || !authService.isAuthenticated()) return;

      try {
        const { default: DeleteFileManager } = await import(
          "../../../../services/Manager/File/DeleteFileManager.js"
        );

        // Pass collection managers to the DeleteFileManager
        const manager = new DeleteFileManager(
          authService,
          getCollectionManager,
          listCollectionManager,
        );
        await manager.initialize();

        setFileManager(manager);

        // Set up event listener
        manager.addFileDeletionListener(handleFileEvent);

        console.log(
          "[Example] DeleteFileManager initialized with collection managers",
        );
        console.log(
          "[Example] Manager has fileCryptoService:",
          !!manager.fileCryptoService,
        );

        // Load files after manager is fully ready
        if (listCollectionManager) {
          console.log(
            "[Example] Auto-loading files after manager initialization",
          );
          await loadAvailableFiles(manager);
        }
      } catch (err) {
        console.error("[Example] Failed to initialize DeleteFileManager:", err);
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();

    return () => {
      if (fileManager) {
        fileManager.removeFileDeletionListener(handleFileEvent);
      }
    };
  }, [
    authService,
    getCollectionManager,
    listCollectionManager,
    handleFileEvent,
    loadAvailableFiles,
    fileManager,
  ]);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!fileManager) return;

    const status = fileManager.getManagerStatus();
    setManagerStatus(status);
    console.log("[Example] Manager status:", status);
  }, [fileManager]);

  // Load manager status when manager is ready
  useEffect(() => {
    if (fileManager) {
      loadManagerStatus();
    }
  }, [fileManager, loadManagerStatus]);

  // Load some available files for testing - this will be called from the initialization effect
  // We keep this effect as a fallback in case the manager becomes ready later
  useEffect(() => {
    if (fileManager && listCollectionManager && availableFiles.length === 0) {
      console.log(
        "[Example] FileManager ready but no files loaded, loading files...",
      );
      loadAvailableFiles();
    }
  }, [
    fileManager,
    listCollectionManager,
    availableFiles.length,
    loadAvailableFiles,
  ]);

  // Soft delete file
  const handleDeleteFile = async (forceRefresh = false) => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setOperations((prev) => ({ ...prev, deleting: true }));
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Deleting file:", selectedFileId);

      const result = await fileManager.deleteFile(
        selectedFileId,
        reason || null,
        forceRefresh,
      );

      setDeletionResult(result);
      setSuccess(
        `File deleted successfully: ${result.name || "[Unable to decrypt]"}`,
      );

      addToEventLog("file_deleted", {
        fileId: selectedFileId,
        fileName: result.name,
        newState: result.state,
        newVersion: result.version,
        reason,
        forceRefresh,
      });

      // Clear selection since file is now deleted
      setSelectedFileId("");

      // Refresh available files
      await loadAvailableFiles();
    } catch (err) {
      console.error("[Example] Failed to delete file:", err);
      setError(err.message);
    } finally {
      setOperations((prev) => ({ ...prev, deleting: false }));
    }
  };

  // Restore file
  const handleRestoreFile = async (forceRefresh = false) => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setOperations((prev) => ({ ...prev, restoring: true }));
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Restoring file:", selectedFileId);

      const result = await fileManager.restoreFile(
        selectedFileId,
        reason || null,
        forceRefresh,
      );

      setDeletionResult(result);
      setSuccess(
        `File restored successfully: ${result.name || "[Unable to decrypt]"}`,
      );

      addToEventLog("file_restored", {
        fileId: selectedFileId,
        fileName: result.name,
        newState: result.state,
        newVersion: result.version,
        reason,
        forceRefresh,
      });

      // Clear selection since file state changed
      setSelectedFileId("");

      // Refresh available files
      await loadAvailableFiles();
    } catch (err) {
      console.error("[Example] Failed to restore file:", err);
      setError(err.message);
    } finally {
      setOperations((prev) => ({ ...prev, restoring: false }));
    }
  };

  // Archive file
  const handleArchiveFile = async () => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setOperations((prev) => ({ ...prev, archiving: true }));
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Archiving file:", selectedFileId);

      const result = await fileManager.archiveFile(
        selectedFileId,
        reason || null,
      );

      setDeletionResult(result);
      setSuccess(
        `File archived successfully: ${result.name || "[Unable to decrypt]"}`,
      );

      addToEventLog("file_archived", {
        fileId: selectedFileId,
        fileName: result.name,
        newState: result.state,
        newVersion: result.version,
        reason,
      });

      // Clear selection since file state changed
      setSelectedFileId("");

      // Refresh available files
      await loadAvailableFiles();
    } catch (err) {
      console.error("[Example] Failed to archive file:", err);
      setError(err.message);
    } finally {
      setOperations((prev) => ({ ...prev, archiving: false }));
    }
  };

  // Unarchive file
  const handleUnarchiveFile = async () => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setOperations((prev) => ({ ...prev, unarchiving: true }));
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Unarchiving file:", selectedFileId);

      const result = await fileManager.unarchiveFile(
        selectedFileId,
        reason || null,
      );

      setDeletionResult(result);
      setSuccess(
        `File unarchived successfully: ${result.name || "[Unable to decrypt]"}`,
      );

      addToEventLog("file_unarchived", {
        fileId: selectedFileId,
        fileName: result.name,
        newState: result.state,
        newVersion: result.version,
        reason,
      });

      // Clear selection since file state changed
      setSelectedFileId("");

      // Refresh available files
      await loadAvailableFiles();
    } catch (err) {
      console.error("[Example] Failed to unarchive file:", err);
      setError(err.message);
    } finally {
      setOperations((prev) => ({ ...prev, unarchiving: false }));
    }
  };

  // Permanently delete file
  const handlePermanentlyDeleteFile = async () => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    if (
      !window.confirm(
        "Are you sure you want to permanently delete this file? This action cannot be undone!",
      )
    ) {
      return;
    }

    setOperations((prev) => ({ ...prev, permanentlyDeleting: true }));
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Permanently deleting file:", selectedFileId);

      await fileManager.permanentlyDeleteFile(selectedFileId, reason || null);

      setDeletionResult(null);
      setSuccess(`File permanently deleted: ${selectedFileId}`);

      addToEventLog("file_permanently_deleted", {
        fileId: selectedFileId,
        reason,
      });

      // Clear selection and refresh available files
      setSelectedFileId("");
      await loadAvailableFiles();
    } catch (err) {
      console.error("[Example] Failed to permanently delete file:", err);
      setError(err.message);
    } finally {
      setOperations((prev) => ({ ...prev, permanentlyDeleting: false }));
    }
  };

  // Get deletion history
  const handleGetDeletionHistory = async (forceRefresh = false) => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      console.log("[Example] Getting deletion history for:", selectedFileId);

      const history = await fileManager.getFileDeletionHistory(
        selectedFileId,
        forceRefresh,
      );

      setDeletionHistory(history);
      setSuccess(`Deletion history loaded for file: ${selectedFileId}`);

      addToEventLog("deletion_history_loaded", {
        fileId: selectedFileId,
        forceRefresh,
      });
    } catch (err) {
      console.error("[Example] Failed to get deletion history:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Extend tombstone expiry
  const handleExtendTombstoneExpiry = async () => {
    if (!fileManager || !selectedFileId) {
      setError("Please select a file");
      return;
    }

    setOperations((prev) => ({ ...prev, extendingTombstone: true }));
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Extending tombstone expiry for:", selectedFileId);

      const result = await fileManager.extendTombstoneExpiry(
        selectedFileId,
        parseInt(extensionDays, 10),
      );

      setSuccess(`Tombstone expiry extended by ${extensionDays} days`);

      addToEventLog("tombstone_extended", {
        fileId: selectedFileId,
        extensionDays: parseInt(extensionDays, 10),
        newExpiry: result.file?.tombstone_expiry,
      });

      // Refresh available files
      await loadAvailableFiles();
    } catch (err) {
      console.error("[Example] Failed to extend tombstone expiry:", err);
      setError(err.message);
    } finally {
      setOperations((prev) => ({ ...prev, extendingTombstone: false }));
    }
  };

  // Get expired tombstones
  const handleGetExpiredTombstones = () => {
    if (!fileManager) return;

    const expired = fileManager.getExpiredTombstones();
    setTombstoneData({ type: "expired", data: expired });
    setSuccess(`Found ${expired.length} expired tombstones`);

    addToEventLog("expired_tombstones_retrieved", {
      count: expired.length,
    });
  };

  // Get restorable files
  const handleGetRestorableFiles = () => {
    if (!fileManager) return;

    const restorable = fileManager.getRestorableFiles();
    setTombstoneData({ type: "restorable", data: restorable });
    setSuccess(`Found ${restorable.length} restorable files`);

    addToEventLog("restorable_files_retrieved", {
      count: restorable.length,
    });
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

  // Get file capabilities for selected file
  const getFileCapabilities = () => {
    if (!fileManager || !selectedFileId) return null;

    const file = availableFiles.find((f) => f.id === selectedFileId);
    if (!file) return null;

    // Return the pre-computed capabilities from the file object
    return {
      canDelete: file.canDelete,
      canRestore: file.canRestore,
      canArchive: file.canArchive,
      canUnarchive: file.canUnarchive,
      canPermanentlyDelete: file.canPermanentlyDelete,
      hasTombstone: file.hasTombstone,
      tombstoneVersion: file.version,
      tombstoneExpiry: file.tombstoneExpiry,
      isExpired: file._tombstone_expired,
      state: file.state,
      version: file.version,
    };
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

  const fileCapabilities = getFileCapabilities();

  if (!authService || !authService.isAuthenticated()) {
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

      <h2>🗑️ Delete File Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This example demonstrates the complete file deletion workflow with
        tombstone management.
        <br />
        <strong>Features:</strong> Soft delete, restore, archive, tombstone
        management, and permanent deletion
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
            <strong>Can Delete Files:</strong>{" "}
            {managerStatus.canDeleteFiles ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Available Files:</strong> {availableFiles.length}
          </div>
          <div>
            <strong>Deletable Files:</strong>{" "}
            {availableFiles.filter((f) => f.canDelete).length}
          </div>
          <div>
            <strong>Restorable Files:</strong>{" "}
            {availableFiles.filter((f) => f.canRestore).length}
          </div>
          <div>
            <strong>Expired Tombstones:</strong>{" "}
            {managerStatus.expiredTombstones || 0}
          </div>
          <div>
            <strong>Listeners:</strong> {managerStatus.listenerCount || 0}
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
            🗂️ Select File ({availableFiles.length} available):
          </h4>
          <button
            onClick={() => loadAvailableFiles()}
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
                : "Select a file for operations..."}
            </option>
            {availableFiles.map((file) => (
              <option key={file.id} value={file.id}>
                {file.name} (v{file.version}, {file.state}) -{" "}
                {file.collectionName} - {file.id.substring(0, 8)}...
                {file.hasTombstone && " [Tombstone]"}
              </option>
            ))}
          </select>
        </div>

        {/* File Capabilities */}
        {fileCapabilities && (
          <div
            style={{ marginBottom: "15px", fontSize: "14px", color: "#666" }}
          >
            <strong>File Capabilities:</strong>
            <span style={{ marginLeft: "10px" }}>
              {fileCapabilities.canDelete && "🗑️ Can Delete "}
              {fileCapabilities.canRestore && "♻️ Can Restore "}
              {fileCapabilities.canArchive && "📦 Can Archive "}
              {fileCapabilities.canUnarchive && "📤 Can Unarchive "}
              {fileCapabilities.canPermanentlyDelete &&
                "💀 Can Permanently Delete "}
              {fileCapabilities.hasTombstone &&
                `🪦 Has Tombstone (expires: ${formatDate(fileCapabilities.tombstoneExpiry)}) `}
            </span>
          </div>
        )}

        {/* Reason Input */}
        <div style={{ marginBottom: "15px" }}>
          <label style={{ display: "block", marginBottom: "5px" }}>
            Reason (optional):
          </label>
          <input
            type="text"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Enter reason for operation..."
            style={{
              width: "100%",
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
            }}
          />
        </div>

        {/* Single File Operations */}
        <div
          style={{
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
            marginBottom: "15px",
          }}
        >
          <button
            onClick={() => handleDeleteFile(false)}
            disabled={
              !selectedFileId ||
              operations.deleting ||
              !fileCapabilities?.canDelete
            }
            style={{
              padding: "8px 16px",
              backgroundColor: fileCapabilities?.canDelete
                ? "#dc3545"
                : "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !selectedFileId ||
                operations.deleting ||
                !fileCapabilities?.canDelete
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            {operations.deleting ? "🔄 Deleting..." : "🗑️ Delete File"}
          </button>
          <button
            onClick={() => handleRestoreFile(false)}
            disabled={
              !selectedFileId ||
              operations.restoring ||
              !fileCapabilities?.canRestore
            }
            style={{
              padding: "8px 16px",
              backgroundColor: fileCapabilities?.canRestore
                ? "#28a745"
                : "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !selectedFileId ||
                operations.restoring ||
                !fileCapabilities?.canRestore
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            {operations.restoring ? "🔄 Restoring..." : "♻️ Restore File"}
          </button>
          <button
            onClick={handleArchiveFile}
            disabled={
              !selectedFileId ||
              operations.archiving ||
              !fileCapabilities?.canArchive
            }
            style={{
              padding: "8px 16px",
              backgroundColor: fileCapabilities?.canArchive
                ? "#6c757d"
                : "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !selectedFileId ||
                operations.archiving ||
                !fileCapabilities?.canArchive
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            {operations.archiving ? "🔄 Archiving..." : "📦 Archive File"}
          </button>
          <button
            onClick={handleUnarchiveFile}
            disabled={
              !selectedFileId ||
              operations.unarchiving ||
              !fileCapabilities?.canUnarchive
            }
            style={{
              padding: "8px 16px",
              backgroundColor: fileCapabilities?.canUnarchive
                ? "#17a2b8"
                : "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !selectedFileId ||
                operations.unarchiving ||
                !fileCapabilities?.canUnarchive
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            {operations.unarchiving ? "🔄 Unarchiving..." : "📤 Unarchive File"}
          </button>
          <button
            onClick={handlePermanentlyDeleteFile}
            disabled={
              !selectedFileId ||
              operations.permanentlyDeleting ||
              !fileCapabilities?.canPermanentlyDelete
            }
            style={{
              padding: "8px 16px",
              backgroundColor: fileCapabilities?.canPermanentlyDelete
                ? "#8b0000"
                : "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !selectedFileId ||
                operations.permanentlyDeleting ||
                !fileCapabilities?.canPermanentlyDelete
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            {operations.permanentlyDeleting
              ? "🔄 Permanently Deleting..."
              : "💀 Permanently Delete"}
          </button>
        </div>

        {/* History and Tombstone Operations */}
        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          <button
            onClick={() => handleGetDeletionHistory(false)}
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
            📋 Get History
          </button>
          <div style={{ display: "flex", alignItems: "center", gap: "5px" }}>
            <input
              type="number"
              value={extensionDays}
              onChange={(e) => setExtensionDays(e.target.value)}
              min="1"
              max="365"
              style={{ width: "60px", padding: "4px" }}
            />
            <button
              onClick={handleExtendTombstoneExpiry}
              disabled={
                !selectedFileId ||
                operations.extendingTombstone ||
                !fileCapabilities?.hasTombstone
              }
              style={{
                padding: "8px 16px",
                backgroundColor: fileCapabilities?.hasTombstone
                  ? "#e83e8c"
                  : "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor:
                  !selectedFileId ||
                  operations.extendingTombstone ||
                  !fileCapabilities?.hasTombstone
                    ? "not-allowed"
                    : "pointer",
              }}
            >
              {operations.extendingTombstone
                ? "🔄 Extending..."
                : "⏰ Extend Tombstone"}
            </button>
          </div>
          <button
            onClick={handleGetExpiredTombstones}
            disabled={!fileManager}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6f42c1",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !fileManager ? "not-allowed" : "pointer",
            }}
          >
            💀 Get Expired
          </button>
          <button
            onClick={handleGetRestorableFiles}
            disabled={!fileManager}
            style={{
              padding: "8px 16px",
              backgroundColor: "#20c997",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !fileManager ? "not-allowed" : "pointer",
            }}
          >
            ♻️ Get Restorable
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

      {/* Operation Results */}
      {deletionResult && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>📋 Operation Result:</h4>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
              gap: "10px",
            }}
          >
            <div>
              <strong>File ID:</strong> {deletionResult.id}
            </div>
            <div>
              <strong>Name:</strong>{" "}
              {deletionResult.name || "[Unable to decrypt]"}
            </div>
            <div>
              <strong>State:</strong>
              <span
                style={{
                  marginLeft: "5px",
                  padding: "2px 6px",
                  borderRadius: "3px",
                  backgroundColor: getStateColor(deletionResult.state),
                  color: "white",
                  fontSize: "12px",
                }}
              >
                {deletionResult.state?.toUpperCase()}
              </span>
            </div>
            <div>
              <strong>Version:</strong> v{deletionResult.version}
            </div>
            {deletionResult.tombstone_version && (
              <>
                <div>
                  <strong>Tombstone Version:</strong> v
                  {deletionResult.tombstone_version}
                </div>
                <div>
                  <strong>Tombstone Expiry:</strong>{" "}
                  {formatDateTime(deletionResult.tombstone_expiry)}
                </div>
              </>
            )}
            <div>
              <strong>Modified:</strong>{" "}
              {formatDateTime(deletionResult.modified_at)}
            </div>
          </div>
        </div>
      )}

      {/* Tombstone Data */}
      {tombstoneData && (
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
            🪦{" "}
            {tombstoneData.type === "expired"
              ? "Expired Tombstones"
              : "Restorable Files"}{" "}
            ({tombstoneData.data.length}):
          </h4>
          <div style={{ maxHeight: "300px", overflow: "auto" }}>
            {tombstoneData.data.length === 0 ? (
              <p>No {tombstoneData.type} files found.</p>
            ) : (
              <table
                style={{
                  width: "100%",
                  borderCollapse: "collapse",
                  fontSize: "12px",
                }}
              >
                <thead>
                  <tr style={{ backgroundColor: "#f8f9fa" }}>
                    <th
                      style={{
                        padding: "5px",
                        border: "1px solid #dee2e6",
                        textAlign: "left",
                      }}
                    >
                      File ID
                    </th>
                    <th
                      style={{
                        padding: "5px",
                        border: "1px solid #dee2e6",
                        textAlign: "left",
                      }}
                    >
                      State
                    </th>
                    <th
                      style={{
                        padding: "5px",
                        border: "1px solid #dee2e6",
                        textAlign: "left",
                      }}
                    >
                      Tombstone Version
                    </th>
                    <th
                      style={{
                        padding: "5px",
                        border: "1px solid #dee2e6",
                        textAlign: "left",
                      }}
                    >
                      Expiry
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {tombstoneData.data.map((item, index) => (
                    <tr key={index}>
                      <td
                        style={{ padding: "5px", border: "1px solid #dee2e6" }}
                      >
                        {item.fileId || item.tombstoneData?.fileId || "Unknown"}
                      </td>
                      <td
                        style={{ padding: "5px", border: "1px solid #dee2e6" }}
                      >
                        {item.tombstoneData?.state || "Unknown"}
                      </td>
                      <td
                        style={{ padding: "5px", border: "1px solid #dee2e6" }}
                      >
                        v{item.tombstoneData?.tombstone_version || "Unknown"}
                      </td>
                      <td
                        style={{ padding: "5px", border: "1px solid #dee2e6" }}
                      >
                        {formatDateTime(item.tombstoneData?.tombstone_expiry)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* Deletion History */}
      {deletionHistory && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#fff",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>📜 Deletion History:</h4>
          <details>
            <summary style={{ cursor: "pointer", fontWeight: "bold" }}>
              View Raw History Data
            </summary>
            <pre
              style={{
                backgroundColor: "#f8f9fa",
                padding: "10px",
                borderRadius: "4px",
                overflow: "auto",
                marginTop: "10px",
                fontSize: "12px",
              }}
            >
              {JSON.stringify(deletionHistory, null, 2)}
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
          <h3>📋 Deletion Event Log ({eventLog.length})</h3>
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
              No deletion events logged yet.
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

export default DeleteFileManagerExample;
