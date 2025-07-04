// File: src/components/FileOperations.jsx
// Reusable component for common file operations (download, delete, archive, restore)
import React, { useState } from "react";
import { useServices } from "../hooks/useService.jsx";

const FileOperations = ({
  file,
  onOperationComplete = null,
  showLabels = false,
  buttonSize = "small", // small, medium, large
  layout = "horizontal", // horizontal, vertical, dropdown
}) => {
  const { fileService } = useServices();
  const [isOperating, setIsOperating] = useState(false);
  const [operationStatus, setOperationStatus] = useState(null);

  // Check file capabilities
  const canDownload = file && (!file._is_deleted || canRestoreFile(file));
  const canEdit = file && (file._is_active || file._is_archived);
  const canArchive = file && file._is_active;
  const canRestore = file && canRestoreFile(file);
  const canDelete = file && !file._is_deleted;

  // File state checks
  function canRestoreFile(file) {
    return file._has_tombstone && !file._tombstone_expired && file._is_deleted;
  }

  // Operation handlers
  const handleDownload = async () => {
    if (!file || isOperating) return;

    setIsOperating(true);
    setOperationStatus("downloading");

    try {
      await fileService.downloadAndSaveFile(file.id);
      setOperationStatus("download_success");

      if (onOperationComplete) {
        onOperationComplete({ type: "download", file, success: true });
      }
    } catch (error) {
      console.error("[FileOperations] Download failed:", error);
      setOperationStatus("download_error");

      if (onOperationComplete) {
        onOperationComplete({
          type: "download",
          file,
          success: false,
          error: error.message,
        });
      }
    } finally {
      setIsOperating(false);
      setTimeout(() => setOperationStatus(null), 3000);
    }
  };

  const handleArchive = async () => {
    if (!file || isOperating) return;

    if (!window.confirm(`Archive "${file.name || file.displayName}"?`)) return;

    setIsOperating(true);
    setOperationStatus("archiving");

    try {
      await fileService.archiveFile(file.id);
      setOperationStatus("archive_success");

      if (onOperationComplete) {
        onOperationComplete({ type: "archive", file, success: true });
      }
    } catch (error) {
      console.error("[FileOperations] Archive failed:", error);
      setOperationStatus("archive_error");

      if (onOperationComplete) {
        onOperationComplete({
          type: "archive",
          file,
          success: false,
          error: error.message,
        });
      }
    } finally {
      setIsOperating(false);
      setTimeout(() => setOperationStatus(null), 3000);
    }
  };

  const handleRestore = async () => {
    if (!file || isOperating) return;

    if (!window.confirm(`Restore "${file.name || file.displayName}"?`)) return;

    setIsOperating(true);
    setOperationStatus("restoring");

    try {
      await fileService.restoreFile(file.id);
      setOperationStatus("restore_success");

      if (onOperationComplete) {
        onOperationComplete({ type: "restore", file, success: true });
      }
    } catch (error) {
      console.error("[FileOperations] Restore failed:", error);
      setOperationStatus("restore_error");

      if (onOperationComplete) {
        onOperationComplete({
          type: "restore",
          file,
          success: false,
          error: error.message,
        });
      }
    } finally {
      setIsOperating(false);
      setTimeout(() => setOperationStatus(null), 3000);
    }
  };

  const handleDelete = async () => {
    if (!file || isOperating) return;

    if (
      !window.confirm(
        `Delete "${file.name || file.displayName}"? This will move it to deleted files where it can be restored.`,
      )
    )
      return;

    setIsOperating(true);
    setOperationStatus("deleting");

    try {
      await fileService.deleteFile(file.id);
      setOperationStatus("delete_success");

      if (onOperationComplete) {
        onOperationComplete({ type: "delete", file, success: true });
      }
    } catch (error) {
      console.error("[FileOperations] Delete failed:", error);
      setOperationStatus("delete_error");

      if (onOperationComplete) {
        onOperationComplete({
          type: "delete",
          file,
          success: false,
          error: error.message,
        });
      }
    } finally {
      setIsOperating(false);
      setTimeout(() => setOperationStatus(null), 3000);
    }
  };

  // Get button size styles
  const getButtonStyles = (type, enabled = true) => {
    const sizes = {
      small: { padding: "4px 8px", fontSize: "12px" },
      medium: { padding: "6px 12px", fontSize: "14px" },
      large: { padding: "8px 16px", fontSize: "16px" },
    };

    const colors = {
      download: {
        backgroundColor: enabled ? "#007bff" : "#ccc",
        color: "white",
      },
      archive: {
        backgroundColor: enabled ? "#6c757d" : "#ccc",
        color: "white",
      },
      restore: {
        backgroundColor: enabled ? "#28a745" : "#ccc",
        color: "white",
      },
      delete: { backgroundColor: enabled ? "#dc3545" : "#ccc", color: "white" },
    };

    return {
      ...sizes[buttonSize],
      ...colors[type],
      border: "none",
      borderRadius: "3px",
      cursor: enabled && !isOperating ? "pointer" : "not-allowed",
      opacity: enabled ? 1 : 0.6,
    };
  };

  // Get operation status message
  const getStatusMessage = () => {
    const messages = {
      downloading: "Downloading...",
      download_success: "Downloaded!",
      download_error: "Download failed",
      archiving: "Archiving...",
      archive_success: "Archived!",
      archive_error: "Archive failed",
      restoring: "Restoring...",
      restore_success: "Restored!",
      restore_error: "Restore failed",
      deleting: "Deleting...",
      delete_success: "Deleted!",
      delete_error: "Delete failed",
    };

    return messages[operationStatus] || null;
  };

  // Get available operations
  const operations = [
    {
      key: "download",
      label: "Download",
      icon: "⬇️",
      enabled: canDownload,
      handler: handleDownload,
      visible: true,
    },
    {
      key: "archive",
      label: "Archive",
      icon: "📦",
      enabled: canArchive,
      handler: handleArchive,
      visible: canArchive,
    },
    {
      key: "restore",
      label: "Restore",
      icon: "♻️",
      enabled: canRestore,
      handler: handleRestore,
      visible: canRestore,
    },
    {
      key: "delete",
      label: "Delete",
      icon: "🗑️",
      enabled: canDelete,
      handler: handleDelete,
      visible: canDelete,
    },
  ].filter((op) => op.visible);

  if (!file || operations.length === 0) {
    return null;
  }

  // Render operations based on layout
  if (layout === "dropdown") {
    return (
      <div style={{ position: "relative", display: "inline-block" }}>
        <select
          onChange={(e) => {
            const operation = operations.find(
              (op) => op.key === e.target.value,
            );
            if (operation && operation.enabled) {
              operation.handler();
            }
            e.target.value = ""; // Reset
          }}
          disabled={isOperating}
          style={{
            ...getButtonStyles("download"),
            backgroundColor: "#6c757d",
          }}
        >
          <option value="">Actions...</option>
          {operations.map((op) => (
            <option key={op.key} value={op.key} disabled={!op.enabled}>
              {op.icon} {op.label}
            </option>
          ))}
        </select>

        {operationStatus && (
          <div
            style={{
              position: "absolute",
              top: "100%",
              left: "0",
              backgroundColor: "#333",
              color: "white",
              padding: "4px 8px",
              borderRadius: "3px",
              fontSize: "12px",
              whiteSpace: "nowrap",
              zIndex: 1000,
            }}
          >
            {getStatusMessage()}
          </div>
        )}
      </div>
    );
  }

  return (
    <div
      style={{
        display: "flex",
        gap: "4px",
        flexDirection: layout === "vertical" ? "column" : "row",
        alignItems: layout === "vertical" ? "stretch" : "center",
      }}
    >
      {operations.map((op) => (
        <button
          key={op.key}
          onClick={op.handler}
          disabled={!op.enabled || isOperating}
          style={getButtonStyles(op.key, op.enabled)}
          title={`${op.label} ${file.name || file.displayName || "file"}`}
        >
          {op.icon}
          {showLabels && <span style={{ marginLeft: "4px" }}>{op.label}</span>}
        </button>
      ))}

      {operationStatus && (
        <div
          style={{
            padding: "2px 6px",
            fontSize: "11px",
            backgroundColor: operationStatus.includes("error")
              ? "#dc3545"
              : operationStatus.includes("success")
                ? "#28a745"
                : "#17a2b8",
            color: "white",
            borderRadius: "3px",
            whiteSpace: "nowrap",
          }}
        >
          {getStatusMessage()}
        </div>
      )}
    </div>
  );
};

// Higher-order component for bulk operations
export const BulkFileOperations = ({
  selectedFiles = [],
  onOperationComplete = null,
  onClearSelection = null,
}) => {
  const { fileService } = useServices();
  const [isOperating, setBulkOperating] = useState(false);
  const [operationStatus, setOperationStatus] = useState(null);

  const handleBulkArchive = async () => {
    if (selectedFiles.length === 0 || isOperating) return;

    if (!window.confirm(`Archive ${selectedFiles.length} file(s)?`)) return;

    setBulkOperating(true);
    setOperationStatus("archiving");

    try {
      const fileIds = selectedFiles.map((f) => f.id);
      const result = await fileService.batchArchiveFiles(fileIds);
      setOperationStatus("archive_success");

      if (onOperationComplete) {
        onOperationComplete({ type: "bulk_archive", result });
      }

      if (onClearSelection) {
        onClearSelection();
      }
    } catch (error) {
      console.error("[BulkFileOperations] Bulk archive failed:", error);
      setOperationStatus("archive_error");
    } finally {
      setBulkOperating(false);
      setTimeout(() => setOperationStatus(null), 3000);
    }
  };

  const handleBulkRestore = async () => {
    if (selectedFiles.length === 0 || isOperating) return;

    if (!window.confirm(`Restore ${selectedFiles.length} file(s)?`)) return;

    setBulkOperating(true);
    setOperationStatus("restoring");

    try {
      const fileIds = selectedFiles.map((f) => f.id);
      const result = await fileService.batchRestoreFiles(fileIds);
      setOperationStatus("restore_success");

      if (onOperationComplete) {
        onOperationComplete({ type: "bulk_restore", result });
      }

      if (onClearSelection) {
        onClearSelection();
      }
    } catch (error) {
      console.error("[BulkFileOperations] Bulk restore failed:", error);
      setOperationStatus("restore_error");
    } finally {
      setBulkOperating(false);
      setTimeout(() => setOperationStatus(null), 3000);
    }
  };

  const handleBulkDelete = async () => {
    if (selectedFiles.length === 0 || isOperating) return;

    if (
      !window.confirm(
        `Delete ${selectedFiles.length} file(s)? They will be moved to deleted files where they can be restored.`,
      )
    )
      return;

    setBulkOperating(true);
    setOperationStatus("deleting");

    try {
      const fileIds = selectedFiles.map((f) => f.id);
      const result = await fileService.deleteMultipleFiles(fileIds);
      setOperationStatus("delete_success");

      if (onOperationComplete) {
        onOperationComplete({ type: "bulk_delete", result });
      }

      if (onClearSelection) {
        onClearSelection();
      }
    } catch (error) {
      console.error("[BulkFileOperations] Bulk delete failed:", error);
      setOperationStatus("delete_error");
    } finally {
      setBulkOperating(false);
      setTimeout(() => setOperationStatus(null), 3000);
    }
  };

  if (selectedFiles.length === 0) {
    return null;
  }

  const canArchive = selectedFiles.some((f) => f._is_active);
  const canRestore = selectedFiles.some(
    (f) => f._is_deleted && !f._tombstone_expired,
  );
  const canDelete = selectedFiles.some((f) => !f._is_deleted);

  return (
    <div
      style={{
        position: "fixed",
        bottom: "20px",
        left: "50%",
        transform: "translateX(-50%)",
        backgroundColor: "#333",
        color: "white",
        padding: "15px 30px",
        borderRadius: "30px",
        boxShadow: "0 2px 10px rgba(0,0,0,0.3)",
        display: "flex",
        alignItems: "center",
        gap: "15px",
      }}
    >
      <span>{selectedFiles.length} file(s) selected</span>

      {canArchive && (
        <button
          onClick={handleBulkArchive}
          disabled={isOperating}
          style={{
            padding: "5px 15px",
            backgroundColor: "#6c757d",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: isOperating ? "not-allowed" : "pointer",
          }}
        >
          📦 Archive
        </button>
      )}

      {canRestore && (
        <button
          onClick={handleBulkRestore}
          disabled={isOperating}
          style={{
            padding: "5px 15px",
            backgroundColor: "#28a745",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: isOperating ? "not-allowed" : "pointer",
          }}
        >
          ♻️ Restore
        </button>
      )}

      {canDelete && (
        <button
          onClick={handleBulkDelete}
          disabled={isOperating}
          style={{
            padding: "5px 15px",
            backgroundColor: "#dc3545",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: isOperating ? "not-allowed" : "pointer",
          }}
        >
          🗑️ Delete
        </button>
      )}

      <button
        onClick={onClearSelection}
        disabled={isOperating}
        style={{
          padding: "5px 15px",
          backgroundColor: "#666",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: isOperating ? "not-allowed" : "pointer",
        }}
      >
        Clear
      </button>

      {operationStatus && (
        <span
          style={{
            padding: "4px 8px",
            borderRadius: "4px",
            fontSize: "12px",
            backgroundColor: operationStatus.includes("error")
              ? "#dc3545"
              : operationStatus.includes("success")
                ? "#28a745"
                : "#17a2b8",
          }}
        >
          {operationStatus
            .replace(/_/g, " ")
            .replace(/\b\w/g, (l) => l.toUpperCase())}
        </span>
      )}
    </div>
  );
};

export default FileOperations;
