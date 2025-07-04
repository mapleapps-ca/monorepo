// File: src/components/FileSyncManager.jsx
// Enhanced FileSyncManager that integrates better with the consolidated FileManager
import React, { useState, useEffect, useCallback } from "react";
import { useServices } from "../hooks/useService.jsx";

const FileSyncManager = ({
  onSyncComplete = null,
  showDetailed = false,
  currentFolderId = null, // Add current folder context
  onRefreshNeeded = null, // Callback to refresh current view
}) => {
  const { authService } = useServices();
  const [syncService, setSyncService] = useState(null);
  const [syncStatus, setSyncStatus] = useState(null);
  const [isSyncing, setIsSyncing] = useState(false);
  const [syncProgress, setSyncProgress] = useState(null);
  const [conflicts, setConflicts] = useState([]);
  const [error, setError] = useState("");
  const [showConflicts, setShowConflicts] = useState(false);
  const [lastSyncResult, setLastSyncResult] = useState(null);

  // Initialize sync service
  useEffect(() => {
    const initSyncService = async () => {
      try {
        const { default: FileSyncService } = await import(
          "../services/FileSyncService.js"
        );
        await FileSyncService.initialize();
        setSyncService(FileSyncService);
        updateSyncStatus(FileSyncService);
      } catch (err) {
        console.error(
          "[FileSyncManager] Failed to initialize sync service:",
          err,
        );
        setError("Failed to initialize sync service");
      }
    };

    if (authService.isAuthenticated()) {
      initSyncService();
    }
  }, [authService]);

  // Update sync status
  const updateSyncStatus = useCallback(
    (service = syncService) => {
      if (!service) return;

      const status = service.getSyncStatus();
      const conflicts = service.getConflictedFiles();

      setSyncStatus(status);
      setConflicts(conflicts);
      setIsSyncing(status.isSyncing);
    },
    [syncService],
  );

  // Refresh status periodically
  useEffect(() => {
    if (!syncService) return;

    const interval = setInterval(() => {
      updateSyncStatus();
    }, 5000); // Update every 5 seconds

    return () => clearInterval(interval);
  }, [syncService, updateSyncStatus]);

  // Handle sync progress
  const handleSyncProgress = useCallback((progress) => {
    setSyncProgress(progress);
    console.log("[FileSyncManager] Sync progress:", progress);
  }, []);

  // Perform full sync with FileManager integration
  const handleFullSync = useCallback(async () => {
    if (!syncService || isSyncing) return;

    try {
      setError("");
      setSyncProgress({ phase: "starting", processedFiles: 0 });

      const stats = await syncService.performFullSync(handleSyncProgress);

      console.log("[FileSyncManager] Full sync completed:", stats);
      updateSyncStatus();
      setLastSyncResult({ type: "full", stats, completedAt: new Date() });

      // Notify parent component to refresh if needed
      if (onRefreshNeeded) {
        onRefreshNeeded();
      }

      if (onSyncComplete) {
        onSyncComplete({ type: "full", stats });
      }
    } catch (err) {
      console.error("[FileSyncManager] Full sync failed:", err);
      setError("Full sync failed: " + err.message);
    } finally {
      setSyncProgress(null);
    }
  }, [
    syncService,
    isSyncing,
    handleSyncProgress,
    updateSyncStatus,
    onSyncComplete,
    onRefreshNeeded,
  ]);

  // Perform incremental sync with FileManager integration
  const handleIncrementalSync = useCallback(async () => {
    if (!syncService || isSyncing) return;

    try {
      setError("");
      setSyncProgress({ phase: "starting", processedFiles: 0 });

      const stats =
        await syncService.performIncrementalSync(handleSyncProgress);

      console.log("[FileSyncManager] Incremental sync completed:", stats);
      updateSyncStatus();
      setLastSyncResult({
        type: "incremental",
        stats,
        completedAt: new Date(),
      });

      // Notify parent component to refresh if needed
      if (onRefreshNeeded) {
        onRefreshNeeded();
      }

      if (onSyncComplete) {
        onSyncComplete({ type: "incremental", stats });
      }
    } catch (err) {
      console.error("[FileSyncManager] Incremental sync failed:", err);
      setError("Incremental sync failed: " + err.message);
    } finally {
      setSyncProgress(null);
    }
  }, [
    syncService,
    isSyncing,
    handleSyncProgress,
    updateSyncStatus,
    onSyncComplete,
    onRefreshNeeded,
  ]);

  // Quick sync for current folder only
  const handleQuickFolderSync = useCallback(async () => {
    if (!syncService || isSyncing || !currentFolderId) return;

    try {
      setError("");
      setSyncProgress({ phase: "starting", processedFiles: 0 });

      // This would be a folder-specific sync if the backend supports it
      // For now, we'll do an incremental sync and filter results
      const stats =
        await syncService.performIncrementalSync(handleSyncProgress);

      console.log("[FileSyncManager] Quick folder sync completed:", stats);
      updateSyncStatus();
      setLastSyncResult({
        type: "folder",
        folderId: currentFolderId,
        stats,
        completedAt: new Date(),
      });

      // Always notify for refresh on folder sync
      if (onRefreshNeeded) {
        onRefreshNeeded();
      }

      if (onSyncComplete) {
        onSyncComplete({ type: "folder", folderId: currentFolderId, stats });
      }
    } catch (err) {
      console.error("[FileSyncManager] Quick folder sync failed:", err);
      setError("Quick folder sync failed: " + err.message);
    } finally {
      setSyncProgress(null);
    }
  }, [
    syncService,
    isSyncing,
    currentFolderId,
    handleSyncProgress,
    updateSyncStatus,
    onSyncComplete,
    onRefreshNeeded,
  ]);

  // Resolve conflict
  const handleResolveConflict = useCallback(
    async (conflictId, resolution) => {
      if (!syncService) return;

      try {
        setError("");
        await syncService.resolveConflict(conflictId, resolution);
        updateSyncStatus();

        // Refresh the current view after conflict resolution
        if (onRefreshNeeded) {
          onRefreshNeeded();
        }

        console.log(
          "[FileSyncManager] Conflict resolved:",
          conflictId,
          resolution,
        );
      } catch (err) {
        console.error("[FileSyncManager] Failed to resolve conflict:", err);
        setError("Failed to resolve conflict: " + err.message);
      }
    },
    [syncService, updateSyncStatus, onRefreshNeeded],
  );

  // Clear sync data
  const handleClearSyncData = useCallback(() => {
    if (!syncService) return;

    if (
      window.confirm(
        "Are you sure you want to clear all sync data? This will require a full resync.",
      )
    ) {
      syncService.clearSyncData();
      updateSyncStatus();
      setLastSyncResult(null);
    }
  }, [syncService, updateSyncStatus]);

  // Auto-sync on mount if needed
  useEffect(() => {
    if (!syncService || isSyncing) return;

    const recommendation = syncService.getRecommendedSyncAction();

    // Auto-sync if it's been a while and we're not showing detailed view
    if (!showDetailed && recommendation === "incremental_sync") {
      const timeSinceLastSync = syncStatus?.lastSyncTime
        ? Date.now() - new Date(syncStatus.lastSyncTime).getTime()
        : Infinity;

      // Auto-sync if more than 10 minutes since last sync
      if (timeSinceLastSync > 10 * 60 * 1000) {
        console.log("[FileSyncManager] Auto-triggering incremental sync");
        handleIncrementalSync();
      }
    }
  }, [syncService, isSyncing, showDetailed, syncStatus, handleIncrementalSync]);

  // Format time duration
  const formatDuration = (ms) => {
    if (!ms) return "N/A";
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`;
    } else {
      return `${seconds}s`;
    }
  };

  // Format file count
  const formatFileCount = (count) => {
    if (count === 0) return "0";
    if (count < 1000) return count.toString();
    if (count < 1000000) return `${(count / 1000).toFixed(1)}K`;
    return `${(count / 1000000).toFixed(1)}M`;
  };

  // Get sync recommendation
  const getSyncRecommendation = () => {
    if (!syncService) return null;
    return syncService.getRecommendedSyncAction();
  };

  // Get status color
  const getStatusColor = (action) => {
    switch (action) {
      case "no_sync_needed":
        return "#28a745";
      case "incremental_sync":
        return "#ffc107";
      case "full_sync":
        return "#17a2b8";
      case "resolve_conflicts":
        return "#dc3545";
      case "wait":
        return "#6c757d";
      default:
        return "#6c757d";
    }
  };

  const getStatusIcon = (action) => {
    switch (action) {
      case "no_sync_needed":
        return "✅";
      case "incremental_sync":
        return "🔄";
      case "full_sync":
        return "📥";
      case "resolve_conflicts":
        return "⚠️";
      case "wait":
        return "⏳";
      default:
        return "❓";
    }
  };

  if (!authService.isAuthenticated()) {
    return (
      <div style={{ padding: "20px", textAlign: "center", color: "#666" }}>
        Please log in to access file synchronization.
      </div>
    );
  }

  if (!syncService) {
    return (
      <div style={{ padding: "20px", textAlign: "center" }}>
        <p>Initializing sync service...</p>
      </div>
    );
  }

  const recommendation = getSyncRecommendation();
  const summary = syncService?.getSyncSummary();

  return (
    <div
      style={{
        padding: "20px",
        border: "1px solid #dee2e6",
        borderRadius: "8px",
        backgroundColor: "#f8f9fa",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "20px",
        }}
      >
        <h3 style={{ margin: 0 }}>📡 File Synchronization</h3>
        <div style={{ display: "flex", gap: "10px" }}>
          {conflicts.length > 0 && showDetailed && (
            <button
              onClick={() => setShowConflicts(!showConflicts)}
              style={{
                padding: "6px 12px",
                fontSize: "12px",
                backgroundColor: "#dc3545",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              {conflicts.length} Conflicts
            </button>
          )}

          {/* Last sync result indicator */}
          {lastSyncResult && (
            <span
              style={{
                padding: "6px 12px",
                fontSize: "12px",
                backgroundColor: "#28a745",
                color: "white",
                borderRadius: "4px",
              }}
            >
              Last: {lastSyncResult.type} sync
            </span>
          )}
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div
          style={{
            backgroundColor: "#fee",
            color: "#c00",
            padding: "10px",
            marginBottom: "15px",
            borderRadius: "4px",
            border: "1px solid #fcc",
          }}
        >
          {error}
        </div>
      )}

      {/* Sync Status */}
      <div style={{ marginBottom: "20px" }}>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "10px",
            marginBottom: "10px",
          }}
        >
          <span style={{ fontSize: "20px" }}>
            {getStatusIcon(recommendation)}
          </span>
          <span
            style={{
              fontWeight: "bold",
              color: getStatusColor(recommendation),
            }}
          >
            {summary?.summary || "Loading..."}
          </span>
        </div>

        {syncStatus?.lastSyncTime && (
          <div style={{ fontSize: "12px", color: "#666" }}>
            Last sync: {new Date(syncStatus.lastSyncTime).toLocaleString()}
          </div>
        )}
      </div>

      {/* Sync Progress */}
      {syncProgress && (
        <div style={{ marginBottom: "20px" }}>
          <div
            style={{
              backgroundColor: "#e0e0e0",
              borderRadius: "4px",
              height: "20px",
              marginBottom: "5px",
              overflow: "hidden",
            }}
          >
            <div
              style={{
                backgroundColor: "#007bff",
                height: "100%",
                width: syncProgress.processedFiles ? "50%" : "10%",
                transition: "width 0.3s ease",
              }}
            />
          </div>
          <div style={{ fontSize: "12px", color: "#666" }}>
            {syncProgress.phase === "starting" && "Starting synchronization..."}
            {syncProgress.phase === "syncing" &&
              `Processing files... (${formatFileCount(syncProgress.processedFiles)} processed)`}
            {syncProgress.phase === "completed" &&
              `Sync completed in ${formatDuration(syncProgress.duration)}`}
          </div>
        </div>
      )}

      {/* Sync Statistics */}
      {syncStatus?.stats && showDetailed && (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(120px, 1fr))",
            gap: "10px",
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "white",
            borderRadius: "4px",
            border: "1px solid #dee2e6",
          }}
        >
          <div style={{ textAlign: "center" }}>
            <div
              style={{ fontSize: "18px", fontWeight: "bold", color: "#007bff" }}
            >
              {formatFileCount(syncStatus.stats.totalFiles)}
            </div>
            <div style={{ fontSize: "12px", color: "#666" }}>Total</div>
          </div>
          <div style={{ textAlign: "center" }}>
            <div
              style={{ fontSize: "18px", fontWeight: "bold", color: "#28a745" }}
            >
              {formatFileCount(syncStatus.stats.newFiles)}
            </div>
            <div style={{ fontSize: "12px", color: "#666" }}>New</div>
          </div>
          <div style={{ textAlign: "center" }}>
            <div
              style={{ fontSize: "18px", fontWeight: "bold", color: "#ffc107" }}
            >
              {formatFileCount(syncStatus.stats.updatedFiles)}
            </div>
            <div style={{ fontSize: "12px", color: "#666" }}>Updated</div>
          </div>
          <div style={{ textAlign: "center" }}>
            <div
              style={{ fontSize: "18px", fontWeight: "bold", color: "#dc3545" }}
            >
              {formatFileCount(syncStatus.stats.deletedFiles)}
            </div>
            <div style={{ fontSize: "12px", color: "#666" }}>Deleted</div>
          </div>
          {syncStatus.stats.conflictedFiles > 0 && (
            <div style={{ textAlign: "center" }}>
              <div
                style={{
                  fontSize: "18px",
                  fontWeight: "bold",
                  color: "#fd7e14",
                }}
              >
                {formatFileCount(syncStatus.stats.conflictedFiles)}
              </div>
              <div style={{ fontSize: "12px", color: "#666" }}>Conflicts</div>
            </div>
          )}
        </div>
      )}

      {/* Action Buttons */}
      <div
        style={{
          display: "flex",
          gap: "10px",
          flexWrap: "wrap",
          marginBottom: "15px",
        }}
      >
        <button
          onClick={handleFullSync}
          disabled={isSyncing}
          style={{
            padding: "8px 16px",
            backgroundColor: "#17a2b8",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: isSyncing ? "not-allowed" : "pointer",
            fontSize: "14px",
          }}
        >
          {isSyncing ? "Syncing..." : "Full Sync"}
        </button>

        <button
          onClick={handleIncrementalSync}
          disabled={isSyncing}
          style={{
            padding: "8px 16px",
            backgroundColor: "#ffc107",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: isSyncing ? "not-allowed" : "pointer",
            fontSize: "14px",
          }}
        >
          Quick Sync
        </button>

        {currentFolderId && (
          <button
            onClick={handleQuickFolderSync}
            disabled={isSyncing}
            style={{
              padding: "8px 16px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isSyncing ? "not-allowed" : "pointer",
              fontSize: "14px",
            }}
          >
            Sync Folder
          </button>
        )}

        {showDetailed && (
          <button
            onClick={handleClearSyncData}
            disabled={isSyncing}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isSyncing ? "not-allowed" : "pointer",
              fontSize: "14px",
            }}
          >
            Reset Sync
          </button>
        )}
      </div>

      {/* Conflicts Section */}
      {showConflicts && conflicts.length > 0 && (
        <div
          style={{
            marginTop: "20px",
            padding: "15px",
            backgroundColor: "#fff3cd",
            border: "1px solid #ffeaa7",
            borderRadius: "4px",
          }}
        >
          <h4 style={{ margin: "0 0 10px 0", color: "#856404" }}>
            ⚠️ File Conflicts ({conflicts.length})
          </h4>

          {conflicts.map((conflict) => (
            <div
              key={conflict.id}
              style={{
                padding: "10px",
                backgroundColor: "white",
                border: "1px solid #dee2e6",
                borderRadius: "4px",
                marginBottom: "10px",
              }}
            >
              <div style={{ fontWeight: "bold", marginBottom: "5px" }}>
                File ID: {conflict.fileId.substring(0, 8)}...
              </div>
              <div
                style={{
                  fontSize: "12px",
                  color: "#666",
                  marginBottom: "10px",
                }}
              >
                Local v{conflict.localVersion} ({conflict.localState}) vs Remote
                v{conflict.remoteVersion} ({conflict.remoteState})
              </div>

              <div style={{ display: "flex", gap: "5px" }}>
                <button
                  onClick={() =>
                    handleResolveConflict(conflict.id, "use_local")
                  }
                  style={{
                    padding: "4px 8px",
                    fontSize: "12px",
                    backgroundColor: "#28a745",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor: "pointer",
                  }}
                >
                  Keep Local
                </button>
                <button
                  onClick={() =>
                    handleResolveConflict(conflict.id, "use_remote")
                  }
                  style={{
                    padding: "4px 8px",
                    fontSize: "12px",
                    backgroundColor: "#dc3545",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor: "pointer",
                  }}
                >
                  Use Remote
                </button>
                <button
                  onClick={() =>
                    handleResolveConflict(conflict.id, "create_copy")
                  }
                  style={{
                    padding: "4px 8px",
                    fontSize: "12px",
                    backgroundColor: "#6c757d",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor: "pointer",
                  }}
                >
                  Create Copy
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Info */}
      <div style={{ fontSize: "12px", color: "#666", marginTop: "15px" }}>
        💡 Synchronization keeps your files up to date across devices. Version
        conflicts occur when the same file is modified on multiple devices.
        {currentFolderId &&
          " Use 'Sync Folder' for quick folder-specific sync."}
      </div>
    </div>
  );
};

export default FileSyncManager;
