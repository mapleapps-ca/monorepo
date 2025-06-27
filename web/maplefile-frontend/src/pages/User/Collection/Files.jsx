// File: src/pages/User/Collection/Files.jsx
import React, { useState, useEffect } from "react";
import { useParams, useNavigate, Link } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useFiles from "../../../hooks/useFiles.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const CollectionFiles = () => {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const { collectionService, passwordStorageService } = useServices();

  // Use the files hook
  const {
    files,
    isLoading: filesLoading,
    error: filesError,
    loadFilesByCollection,
    getActiveFiles,
    getArchivedFiles,
    getDeletedFiles,
  } = useFiles(collectionId);

  const [collection, setCollection] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filter, setFilter] = useState("active"); // active, archived, deleted, all
  const [currentPage, setCurrentPage] = useState(1);
  const [filesPerPage] = useState(20);

  // Load collection info
  useEffect(() => {
    loadCollection();
  }, [collectionId]);

  const loadCollection = async () => {
    if (!collectionId) {
      setError("No collection ID provided");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError("");

      // Get stored password
      const password = passwordStorageService.getPassword();

      // Load collection with password
      const collectionData = await collectionService.getCollection(
        collectionId,
        password,
      );

      setCollection(collectionData);
      console.log("Collection loaded:", collectionData);
    } catch (err) {
      console.error("Failed to load collection:", err);
      setError(err.message || "Failed to load collection");
    } finally {
      setLoading(false);
    }
  };

  // Get filtered files based on state
  const getFilteredFiles = () => {
    switch (filter) {
      case "active":
        return getActiveFiles();
      case "archived":
        return getArchivedFiles();
      case "deleted":
        return getDeletedFiles();
      case "all":
      default:
        return files;
    }
  };

  // Pagination logic
  const filteredFiles = getFilteredFiles();
  const totalPages = Math.ceil(filteredFiles.length / filesPerPage);
  const indexOfLastFile = currentPage * filesPerPage;
  const indexOfFirstFile = indexOfLastFile - filesPerPage;
  const currentFiles = filteredFiles.slice(indexOfFirstFile, indexOfLastFile);

  // Handle page change
  const handlePageChange = (pageNumber) => {
    setCurrentPage(pageNumber);
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (!bytes) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "Unknown";
    return new Date(dateString).toLocaleString();
  };

  if (loading || filesLoading) {
    return (
      <div>
        <h1>Loading Files...</h1>
        <p>Please wait while we load the files in this collection...</p>
        <button onClick={() => navigate(`/collections/${collectionId}`)}>
          ← Back to Collection
        </button>
      </div>
    );
  }

  if (error || filesError) {
    return (
      <div>
        <h1>Error</h1>
        <p style={{ color: "red" }}>{error || filesError}</p>
        <button onClick={() => navigate(`/collections/${collectionId}`)}>
          ← Back to Collection
        </button>
      </div>
    );
  }

  if (!collection) {
    return (
      <div>
        <h1>Collection Not Found</h1>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
      </div>
    );
  }

  return (
    <div>
      {/* Header with breadcrumb */}
      <div>
        <div>
          <Link to="/collections">Collections</Link>
          {" > "}
          <Link to={`/collections/${collectionId}`}>
            {collection.name || "[Encrypted]"}
          </Link>
          {" > "}
          <span>Files</span>
        </div>

        <h1>Files in "{collection.name || "[Encrypted]"}"</h1>

        <div>
          <span>{collection.collection_type === "album" ? "🖼️" : "📁"}</span>
          <span> Collection ID: {collection.id}</span>
        </div>
      </div>

      {/* Collection hierarchy info */}
      {collection.parent_id && (
        <div>
          <p>
            Parent Collection:
            <Link to={`/collections/${collection.parent_id}/files`}>
              {collection.parent_id}
            </Link>
          </p>
        </div>
      )}

      {/* Action buttons */}
      <div>
        <button onClick={() => navigate(`/collections/${collectionId}`)}>
          ← Back to Collection Details
        </button>
        <button onClick={() => navigate(`/collections/${collectionId}/upload`)}>
          + Upload Files
        </button>
      </div>

      {/* Filter buttons */}
      <div>
        <h3>Filter Files:</h3>
        <button onClick={() => setFilter("all")} disabled={filter === "all"}>
          All ({files.length})
        </button>
        <button
          onClick={() => setFilter("active")}
          disabled={filter === "active"}
        >
          Active ({getActiveFiles().length})
        </button>
        <button
          onClick={() => setFilter("archived")}
          disabled={filter === "archived"}
        >
          Archived ({getArchivedFiles().length})
        </button>
        <button
          onClick={() => setFilter("deleted")}
          disabled={filter === "deleted"}
        >
          Deleted ({getDeletedFiles().length})
        </button>
      </div>

      {/* Files list */}
      <div>
        <h3>Files ({filteredFiles.length} total)</h3>

        {currentFiles.length === 0 ? (
          <div>
            <p>No files found in this collection.</p>
            <button
              onClick={() => navigate(`/collections/${collectionId}/upload`)}
            >
              Upload your first file
            </button>
          </div>
        ) : (
          <div>
            {/* File items */}
            {currentFiles.map((file) => (
              <div
                key={file.id}
                style={{
                  border: "1px solid #ccc",
                  padding: "10px",
                  marginBottom: "10px",
                }}
              >
                <div>
                  <strong>File ID:</strong> {file.id}
                </div>

                {/* Encrypted metadata */}
                <div>
                  <strong>Encrypted Metadata:</strong>
                  {file.encrypted_metadata
                    ? " [Encrypted - Need to decrypt]"
                    : " [No metadata]"}
                </div>

                {/* File details */}
                <div>
                  <strong>State:</strong> {file.state}
                </div>
                <div>
                  <strong>Created:</strong> {formatDate(file.created_at)}
                </div>
                <div>
                  <strong>Modified:</strong> {formatDate(file.modified_at)}
                </div>

                {/* Size information */}
                {file.encrypted_file_size_in_bytes > 0 && (
                  <div>
                    <strong>Encrypted Size:</strong>{" "}
                    {formatFileSize(file.encrypted_file_size_in_bytes)}
                  </div>
                )}

                {file.encrypted_thumbnail_size_in_bytes > 0 && (
                  <div>
                    <strong>Has Thumbnail:</strong> Yes (
                    {formatFileSize(file.encrypted_thumbnail_size_in_bytes)})
                  </div>
                )}

                {/* Encryption info */}
                <div>
                  <strong>Encryption Version:</strong>{" "}
                  {file.encryption_version || "Unknown"}
                </div>

                {/* Action buttons */}
                <div style={{ marginTop: "10px" }}>
                  <button onClick={() => navigate(`/files/${file.id}`)}>
                    View Details
                  </button>
                  <button onClick={() => console.log("Download:", file.id)}>
                    Download
                  </button>
                  {file.state === "active" && (
                    <button onClick={() => console.log("Archive:", file.id)}>
                      Archive
                    </button>
                  )}
                  {file.state === "archived" && (
                    <button onClick={() => console.log("Restore:", file.id)}>
                      Restore
                    </button>
                  )}
                  {file.state !== "deleted" && (
                    <button
                      onClick={() => console.log("Delete:", file.id)}
                      style={{ color: "red" }}
                    >
                      Delete
                    </button>
                  )}
                </div>
              </div>
            ))}

            {/* Pagination */}
            {totalPages > 1 && (
              <div style={{ marginTop: "20px" }}>
                <h4>Pages:</h4>
                <div>
                  <button
                    onClick={() => handlePageChange(1)}
                    disabled={currentPage === 1}
                  >
                    First
                  </button>
                  <button
                    onClick={() => handlePageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                  >
                    Previous
                  </button>

                  <span style={{ margin: "0 10px" }}>
                    Page {currentPage} of {totalPages}
                  </span>

                  <button
                    onClick={() => handlePageChange(currentPage + 1)}
                    disabled={currentPage === totalPages}
                  >
                    Next
                  </button>
                  <button
                    onClick={() => handlePageChange(totalPages)}
                    disabled={currentPage === totalPages}
                  >
                    Last
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Debug info */}
      {import.meta.env.DEV && (
        <details style={{ marginTop: "40px" }}>
          <summary>🔍 Debug Information</summary>
          <pre>
            {JSON.stringify(
              {
                collectionId,
                collectionName: collection.name,
                collectionType: collection.collection_type,
                parentId: collection.parent_id,
                totalFiles: files.length,
                activeFiles: getActiveFiles().length,
                archivedFiles: getArchivedFiles().length,
                deletedFiles: getDeletedFiles().length,
                currentFilter: filter,
                currentPage,
                totalPages,
                filesPerPage,
              },
              null,
              2,
            )}
          </pre>
        </details>
      )}
    </div>
  );
};

export default withPasswordProtection(CollectionFiles);
