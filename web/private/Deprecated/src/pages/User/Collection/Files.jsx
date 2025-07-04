// File: src/pages/User/Collection/Files.jsx
// Updated to support version, state, tombstone_version, and tombstone_expiry fields
import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";
import useFiles from "../../../hooks/useFiles.js";

const CollectionFiles = () => {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const { collectionService, passwordStorageService, collectionCryptoService } =
    useServices();

  const {
    files,
    isLoading,
    error: filesError,
    loadFilesByCollection,
    getActiveFiles,
    getArchivedFiles,
    getDeletedFiles,
    getTombstoneFiles,
    getRestorableFiles,
    getPermanentlyDeletableFiles,
    getFileStats,
    deleteFile,
    archiveFile,
    restoreFile,
    downloadAndSaveFile,
    reloadFiles,
    canDownloadFile,
    canEditFile,
    canRestoreFile,
    canPermanentlyDeleteFile,
    getFileVersionInfo,
    FILE_STATES,
  } = useFiles(collectionId);

  const [collection, setCollection] = useState(null);
  const [selectedFiles, setSelectedFiles] = useState(new Set());
  const [error, setError] = useState("");
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());
  const [viewMode, setViewMode] = useState("active"); // active, archived, deleted, all
  const [showDetails, setShowDetails] = useState(false);

  const loadCollectionAndFiles = async (includeStates = null) => {
    try {
      console.log("[Files] === Loading Collection and Files ===");

      // Load collection info FIRST and ensure it's decrypted
      const password = passwordStorageService.getPassword();
      console.log("[Files] Password available:", !!password);

      const collectionData = await collectionService.getCollection(
        collectionId,
        password,
      );

      console.log("[Files] Collection loaded:", {
        id: collectionData.id,
        name: collectionData.name,
        hasCollectionKey: !!collectionData.collection_key,
        collectionKeyLength: collectionData.collection_key?.length,
      });

      // Verify collection key is cached
      const cachedKey =
        collectionCryptoService.getCachedCollectionKey(collectionId);
      console.log("[Files] Collection key cached:", !!cachedKey);

      // CRITICAL: Check if the cached key matches the collection's key
      if (collectionData.collection_key && cachedKey) {
        const keysMatch = collectionData.collection_key.every(
          (byte, index) => byte === cachedKey[index],
        );
        console.log("[Files] Collection keys match:", keysMatch);

        if (!keysMatch) {
          console.error("[Files] Collection key mismatch! Re-caching...");
          collectionCryptoService.cacheCollectionKey(
            collectionId,
            collectionData.collection_key,
          );
        }
      }

      setCollection(collectionData);

      // Now load files with optional state filtering
      console.log("[Files] Loading files with states:", includeStates);
      await reloadFiles(true, includeStates);
    } catch (err) {
      console.error("[Files] Failed to load collection and files:", err);
      setError("Failed to load collection: " + err.message);
    }
  };

  useEffect(() => {
    const statesToInclude = getStatesToInclude(viewMode);
    loadCollectionAndFiles(statesToInclude);
  }, [collectionId, viewMode]);

  // Get states to include based on view mode
  const getStatesToInclude = (mode) => {
    switch (mode) {
      case "active":
        return [FILE_STATES.ACTIVE];
      case "archived":
        return [FILE_STATES.ARCHIVED];
      case "deleted":
        return [FILE_STATES.DELETED];
      case "pending":
        return [FILE_STATES.PENDING];
      case "all":
        return Object.values(FILE_STATES);
      default:
        return [FILE_STATES.ACTIVE];
    }
  };

  // Get files for current view mode
  const getCurrentFiles = () => {
    switch (viewMode) {
      case "active":
        return getActiveFiles();
      case "archived":
        return getArchivedFiles();
      case "deleted":
        return getDeletedFiles();
      case "pending":
        return files.filter((f) => f.state === FILE_STATES.PENDING);
      case "all":
        return files;
      default:
        return getActiveFiles();
    }
  };

  const handleFileSelect = (fileId) => {
    const newSelected = new Set(selectedFiles);
    if (newSelected.has(fileId)) {
      newSelected.delete(fileId);
    } else {
      newSelected.add(fileId);
    }
    setSelectedFiles(newSelected);
  };

  const handleDeleteFile = async (fileId) => {
    if (!window.confirm("Are you sure you want to delete this file?")) return;

    try {
      await deleteFile(fileId);
      setSelectedFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    } catch (err) {
      console.error("Failed to delete file:", err);
      alert("Failed to delete file: " + err.message);
    }
  };

  const handleArchiveFile = async (fileId) => {
    if (!window.confirm("Are you sure you want to archive this file?")) return;

    try {
      await archiveFile(fileId);
      setSelectedFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    } catch (err) {
      console.error("Failed to archive file:", err);
      alert("Failed to archive file: " + err.message);
    }
  };

  const handleRestoreFile = async (fileId) => {
    if (!window.confirm("Are you sure you want to restore this file?")) return;

    try {
      await restoreFile(fileId);
      setSelectedFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    } catch (err) {
      console.error("Failed to restore file:", err);
      alert("Failed to restore file: " + err.message);
    }
  };

  const handleDownloadFile = async (fileId, fileName) => {
    const file = files.find((f) => f.id === fileId);
    if (file && !canDownloadFile(file)) {
      alert("This file cannot be downloaded in its current state.");
      return;
    }

    try {
      console.log("[Files] Starting download for:", fileId, fileName);

      // Track this file as downloading
      setDownloadingFiles((prev) => new Set(prev).add(fileId));

      await downloadAndSaveFile(fileId);

      console.log("[Files] File download completed successfully");
    } catch (err) {
      console.error("[Files] Failed to download file:", err);
      alert("Failed to download file: " + err.message);
    } finally {
      // Remove from downloading set
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const formatDate = (dateString) => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "N/A";
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return "Invalid Date";
    }
  };

  const formatDateTime = (dateString) => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "N/A";
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return "Invalid Date";
    }
  };

  const getStateColor = (state) => {
    switch (state) {
      case FILE_STATES.ACTIVE:
        return "#28a745";
      case FILE_STATES.ARCHIVED:
        return "#6c757d";
      case FILE_STATES.DELETED:
        return "#dc3545";
      case FILE_STATES.PENDING:
        return "#ffc107";
      default:
        return "#6c757d";
    }
  };

  const getStateIcon = (state) => {
    switch (state) {
      case FILE_STATES.ACTIVE:
        return "✅";
      case FILE_STATES.ARCHIVED:
        return "📦";
      case FILE_STATES.DELETED:
        return "🗑️";
      case FILE_STATES.PENDING:
        return "⏳";
      default:
        return "❓";
    }
  };

  const getFileActions = (file) => {
    const actions = [];

    // Download action (if possible)
    if (canDownloadFile(file)) {
      actions.push({
        label: downloadingFiles.has(file.id) ? "Downloading..." : "Download",
        action: () =>
          handleDownloadFile(file.id, file.name || "downloaded_file"),
        disabled: !file._file_key || downloadingFiles.has(file.id),
        style: {
          backgroundColor:
            file._file_key && !downloadingFiles.has(file.id)
              ? "#007bff"
              : "#ccc",
          color: "white",
        },
        title: !file._file_key
          ? "File key not available - refresh page"
          : "Download file",
      });
    }

    // Archive/Unarchive action
    if (file._is_active) {
      actions.push({
        label: "Archive",
        action: () => handleArchiveFile(file.id),
        style: { backgroundColor: "#6c757d", color: "white" },
      });
    }

    // Restore action
    if (canRestoreFile(file)) {
      actions.push({
        label: "Restore",
        action: () => handleRestoreFile(file.id),
        style: { backgroundColor: "#28a745", color: "white" },
      });
    }

    // Delete action (if not already deleted)
    if (!file._is_deleted) {
      actions.push({
        label: "Delete",
        action: () => handleDeleteFile(file.id),
        style: { backgroundColor: "#dc3545", color: "white" },
      });
    }

    return actions;
  };

  const currentFiles = getCurrentFiles();
  const fileStats = getFileStats();

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      {/* Header */}
      <div style={{ marginBottom: "30px" }}>
        <button
          onClick={() => navigate(`/collections/${collectionId}`)}
          style={{ marginBottom: "20px" }}
        >
          ← Back to Collection
        </button>

        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "20px",
          }}
        >
          <div>
            <h1 style={{ margin: 0 }}>Files in Collection</h1>
            {collection && (
              <p style={{ color: "#666", marginTop: "5px" }}>
                Collection: <strong>{collection.name || "[Encrypted]"}</strong>
              </p>
            )}
          </div>

          <button
            onClick={() => navigate(`/collections/${collectionId}/add-file`)}
            style={{
              padding: "10px 20px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
              fontSize: "16px",
            }}
          >
            ➕ Add File
          </button>
        </div>

        {/* File Statistics */}
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
            <strong>Total:</strong> {fileStats.total}
          </div>
          <div>
            <strong>Active:</strong>{" "}
            <span style={{ color: getStateColor(FILE_STATES.ACTIVE) }}>
              {fileStats.active}
            </span>
          </div>
          <div>
            <strong>Archived:</strong>{" "}
            <span style={{ color: getStateColor(FILE_STATES.ARCHIVED) }}>
              {fileStats.archived}
            </span>
          </div>
          <div>
            <strong>Deleted:</strong>{" "}
            <span style={{ color: getStateColor(FILE_STATES.DELETED) }}>
              {fileStats.deleted}
            </span>
          </div>
          <div>
            <strong>Pending:</strong>{" "}
            <span style={{ color: getStateColor(FILE_STATES.PENDING) }}>
              {fileStats.pending}
            </span>
          </div>
          {fileStats.withTombstones > 0 && (
            <div>
              <strong>With Tombstones:</strong> {fileStats.withTombstones}
            </div>
          )}
          {fileStats.restorable > 0 && (
            <div>
              <strong>Restorable:</strong> {fileStats.restorable}
            </div>
          )}
        </div>

        {/* View Mode Toggle */}
        <div
          style={{
            display: "flex",
            gap: "10px",
            marginBottom: "20px",
            flexWrap: "wrap",
          }}
        >
          {[
            { key: "active", label: "Active", count: fileStats.active },
            { key: "archived", label: "Archived", count: fileStats.archived },
            { key: "deleted", label: "Deleted", count: fileStats.deleted },
            { key: "pending", label: "Pending", count: fileStats.pending },
            { key: "all", label: "All", count: fileStats.total },
          ].map(({ key, label, count }) => (
            <button
              key={key}
              onClick={() => setViewMode(key)}
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

        {/* Show Details Toggle */}
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
      </div>

      {/* Loading State */}
      {isLoading && (
        <div style={{ textAlign: "center", padding: "40px" }}>
          <p>Loading files...</p>
        </div>
      )}

      {/* Error State */}
      {(error || filesError) && (
        <div
          style={{
            backgroundColor: "#fee",
            color: "#c00",
            padding: "15px",
            marginBottom: "20px",
            borderRadius: "4px",
          }}
        >
          Error: {error || filesError}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && currentFiles.length === 0 && (
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
          {viewMode === "active" ? (
            <>
              <p style={{ color: "#666", marginBottom: "20px" }}>
                Start by adding your first file to this collection
              </p>
              <button
                onClick={() =>
                  navigate(`/collections/${collectionId}/add-file`)
                }
                style={{
                  padding: "12px 30px",
                  backgroundColor: "#007bff",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                  fontSize: "16px",
                }}
              >
                Add Your First File
              </button>
            </>
          ) : (
            <p style={{ color: "#666" }}>
              Switch to "Active" view to see available files, or use "All" to
              see files in all states.
            </p>
          )}
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
              <div style={{ display: "flex", gap: "10px" }}>
                <button
                  onClick={() => {
                    alert("Bulk operations coming soon");
                  }}
                  style={{
                    padding: "8px 16px",
                    backgroundColor: "#dc3545",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor: "pointer",
                  }}
                >
                  Bulk Actions
                </button>
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
                const versionInfo = getFileVersionInfo(file);
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
                            {file._decrypt_error && (
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
                          v{versionInfo.currentVersion}
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
                        {getFileActions(file).map((action, index) => (
                          <button
                            key={index}
                            onClick={action.action}
                            disabled={action.disabled}
                            style={{
                              padding: "4px 8px",
                              border: "none",
                              borderRadius: "4px",
                              cursor: action.disabled
                                ? "not-allowed"
                                : "pointer",
                              fontSize: "12px",
                              ...action.style,
                            }}
                            title={action.title}
                          >
                            {action.label}
                          </button>
                        ))}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Info Box */}
      <div
        style={{
          marginTop: "40px",
          padding: "20px",
          backgroundColor: "#f8f9fa",
          borderRadius: "4px",
          borderLeft: "4px solid #17a2b8",
        }}
      >
        <h4 style={{ marginTop: 0 }}>ℹ️ About File States & Versions</h4>
        <div style={{ color: "#666", lineHeight: "1.6" }}>
          <p style={{ marginBottom: "10px" }}>
            <strong>File States:</strong>
          </p>
          <ul style={{ marginLeft: "20px", marginBottom: "15px" }}>
            <li>
              <strong>Active:</strong> Files that are fully uploaded and
              available for use
            </li>
            <li>
              <strong>Pending:</strong> Files that are being uploaded or
              processing
            </li>
            <li>
              <strong>Archived:</strong> Files that are stored but not actively
              used
            </li>
            <li>
              <strong>Deleted:</strong> Files that have been soft-deleted (can
              be restored)
            </li>
          </ul>
          <p style={{ marginBottom: "10px" }}>
            <strong>Versioning & Tombstones:</strong>
          </p>
          <ul style={{ marginLeft: "20px", marginBottom: "0" }}>
            <li>
              Each file change increments the version number for conflict
              resolution
            </li>
            <li>
              Deleted files create "tombstones" that expire after a period
            </li>
            <li>
              Files can be restored from deletion before their tombstone expires
            </li>
            <li>
              All files are end-to-end encrypted on your device before upload
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
};

const ProtectedCollectionFiles = withPasswordProtection(CollectionFiles);
export default ProtectedCollectionFiles;
