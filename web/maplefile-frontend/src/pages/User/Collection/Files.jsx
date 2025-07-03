// File: src/pages/User/Collection/Files.jsx
import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";
import useFiles from "../../../hooks/useFiles.js";

const CollectionFiles = () => {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const { collectionService, passwordStorageService } = useServices();

  const {
    files,
    isLoading,
    error,
    loadFilesByCollection,
    getActiveFiles,
    deleteFile,
    downloadAndSaveFile,
  } = useFiles(collectionId);

  const [collection, setCollection] = useState(null);
  const [selectedFiles, setSelectedFiles] = useState(new Set());

  useEffect(() => {
    loadCollectionAndFiles();
  }, [collectionId]);

  const loadCollectionAndFiles = async () => {
    try {
      // Load collection info FIRST and ensure it's decrypted
      const password = passwordStorageService.getPassword();
      console.log("[Files] Loading collection with password...");

      const collectionData = await collectionService.getCollection(
        collectionId,
        password,
      );
      setCollection(collectionData);

      console.log(
        "[Files] Collection loaded, collection key cached:",
        !!collectionData.collection_key,
      );

      // Now load files - the collection key should be cached
      console.log("[Files] Loading files for collection...");
      await reloadFiles(true);
    } catch (err) {
      console.error("Failed to load collection:", err);
      setError("Failed to load collection: " + err.message);
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
      // Refresh files list
      await loadFilesByCollection(collectionId);
    } catch (err) {
      console.error("Failed to delete file:", err);
      alert("Failed to delete file: " + err.message);
    }
  };

  const handleDownloadFile = async (fileId, fileName) => {
    try {
      console.log("[Files] Starting download for:", fileId, fileName);

      // Don't call setIsLoading here - it's managed by the useFiles hook
      await downloadAndSaveFile(fileId);

      console.log("[Files] File download completed successfully");
    } catch (err) {
      console.error("[Files] Failed to download file:", err);
      alert("Failed to download file: " + err.message);
    }
    // No finally block needed - useFiles hook manages its own loading state
  };

  const activeFiles = getActiveFiles();

  return (
    <div style={{ padding: "20px", maxWidth: "1000px", margin: "0 auto" }}>
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
      </div>

      {/* Loading State */}
      {isLoading && (
        <div style={{ textAlign: "center", padding: "40px" }}>
          <p>Loading files...</p>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div
          style={{
            backgroundColor: "#fee",
            color: "#c00",
            padding: "15px",
            marginBottom: "20px",
            borderRadius: "4px",
          }}
        >
          Error: {error}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && activeFiles.length === 0 && (
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
          <h3>No files in this collection yet</h3>
          <p style={{ color: "#666", marginBottom: "20px" }}>
            Start by adding your first file to this collection
          </p>
          <button
            onClick={() => navigate(`/collections/${collectionId}/add-file`)}
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
        </div>
      )}

      {/* Files List */}
      {!isLoading && activeFiles.length > 0 && (
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
                    // Implement bulk delete
                    alert("Bulk delete coming soon");
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
                  Delete Selected
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
                      selectedFiles.size === activeFiles.length &&
                      activeFiles.length > 0
                    }
                    onChange={(e) => {
                      if (e.target.checked) {
                        setSelectedFiles(new Set(activeFiles.map((f) => f.id)));
                      } else {
                        setSelectedFiles(new Set());
                      }
                    }}
                  />
                </th>
                <th style={{ padding: "12px", textAlign: "left" }}>File</th>
                <th style={{ padding: "12px", textAlign: "left" }}>Size</th>
                <th style={{ padding: "12px", textAlign: "left" }}>Created</th>
                <th style={{ padding: "12px", textAlign: "center" }}>
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {activeFiles.map((file) => (
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
                    {file.encrypted_file_size_in_bytes
                      ? `${(file.encrypted_file_size_in_bytes / 1024).toFixed(2)} KB`
                      : "Unknown"}
                  </td>
                  <td style={{ padding: "12px" }}>
                    {file.created_at
                      ? new Date(file.created_at).toLocaleDateString()
                      : "Unknown"}
                  </td>
                  <td style={{ padding: "12px", textAlign: "center" }}>
                    <button
                      onClick={() =>
                        handleDownloadFile(
                          file.id,
                          file.name || "downloaded_file",
                        )
                      }
                      disabled={!file._file_key || isLoading}
                      style={{
                        padding: "6px 12px",
                        marginRight: "8px",
                        backgroundColor:
                          file._file_key && !isLoading ? "#007bff" : "#ccc",
                        color: "white",
                        border: "none",
                        borderRadius: "4px",
                        cursor:
                          file._file_key && !isLoading
                            ? "pointer"
                            : "not-allowed",
                        fontSize: "14px",
                      }}
                      title={
                        !file._file_key
                          ? "File key not available - refresh page"
                          : "Download file"
                      }
                    >
                      ⬇️ {isLoading ? "Downloading..." : "Download"}
                    </button>
                    <button
                      onClick={() => handleDeleteFile(file.id)}
                      style={{
                        padding: "6px 12px",
                        backgroundColor: "#dc3545",
                        color: "white",
                        border: "none",
                        borderRadius: "4px",
                        cursor: "pointer",
                        fontSize: "14px",
                      }}
                    >
                      🗑️ Delete
                    </button>
                  </td>
                </tr>
              ))}
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
        <h4 style={{ marginTop: 0 }}>ℹ️ About File Encryption</h4>
        <p style={{ marginBottom: 0, color: "#666" }}>
          All files in this collection are end-to-end encrypted. File names and
          content are encrypted on your device before upload, ensuring only you
          and authorized members can access them.
        </p>
      </div>
    </div>
  );
};
const ProtectedCollectionFiles = withPasswordProtection(CollectionFiles);

export default ProtectedCollectionFiles;
