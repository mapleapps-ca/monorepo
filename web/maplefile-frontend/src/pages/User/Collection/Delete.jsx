// File: src/pages/User/Collection/Delete.jsx
import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const CollectionDelete = () => {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const { collectionService, fileService, passwordStorageService } =
    useServices();
  const { isAuthenticated } = useAuth();

  // State
  const [collection, setCollection] = useState(null);
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [loadingFiles, setLoadingFiles] = useState(false);
  const [error, setError] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [confirmText, setConfirmText] = useState("");
  const [showFinalConfirm, setShowFinalConfirm] = useState(false);

  // Load collection and files on mount
  useEffect(() => {
    loadCollectionData();
  }, [collectionId]);

  const loadCollectionData = async () => {
    if (!collectionId) {
      setError("No collection ID provided");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError("");

      // Get stored password
      const storedPassword = passwordStorageService.getPassword();

      // Load collection with password
      const collectionData = await collectionService.getCollection(
        collectionId,
        storedPassword,
      );

      setCollection(collectionData);

      // Check if user is the owner
      const userEmail = localStorage.getItem("mapleapps_user_email");
      const isOwner = collectionData.members?.some(
        (member) =>
          member.recipient_email === userEmail &&
          member.permission_level === "admin" &&
          member.recipient_id === collectionData.owner_id,
      );

      if (!isOwner) {
        setError("Only the collection owner can delete this collection");
        setLoading(false);
        return;
      }

      console.log("[CollectionDelete] Collection loaded:", collectionData);

      // Try to load files count
      await loadFilesCount();
    } catch (err) {
      console.error("[CollectionDelete] Failed to load collection:", err);
      setError(err.message || "Failed to load collection");
    } finally {
      setLoading(false);
    }
  };

  const loadFilesCount = async () => {
    try {
      setLoadingFiles(true);
      const fileList = await fileService.listFilesByCollection(collectionId);
      setFiles(fileList);
      console.log(`[CollectionDelete] Found ${fileList.length} files`);
    } catch (err) {
      console.error("[CollectionDelete] Failed to load files:", err);
      // Don't set error - files loading is not critical
    } finally {
      setLoadingFiles(false);
    }
  };

  // Handle confirmation text change
  const handleConfirmTextChange = (e) => {
    setConfirmText(e.target.value);
    if (error) setError("");
  };

  // Handle initial delete button click
  const handleInitialDelete = () => {
    if (!isAuthenticated) {
      setError("You must be logged in to delete collections");
      navigate("/login");
      return;
    }

    const expectedText = `DELETE ${collection.name}`;
    if (confirmText !== expectedText) {
      setError(`Please type "${expectedText}" to confirm deletion`);
      return;
    }

    // Show final confirmation
    setShowFinalConfirm(true);
  };

  // Handle final delete confirmation
  const handleFinalDelete = async () => {
    setDeleting(true);
    setError("");

    try {
      console.log("[CollectionDelete] Deleting collection:", collectionId);

      await collectionService.deleteCollection(collectionId);

      console.log(
        "[CollectionDelete] Collection deleted successfully:",
        collectionId,
      );

      // Navigate to collections list with success message
      navigate("/collections", {
        state: {
          message: `Collection "${collection.name}" and all its contents have been deleted successfully.`,
        },
      });
    } catch (err) {
      console.error("[CollectionDelete] Failed to delete collection:", err);

      if (err.message.includes("permission")) {
        setError("You don't have permission to delete this collection.");
      } else {
        setError(err.message || "Failed to delete collection");
      }
      setShowFinalConfirm(false);
    } finally {
      setDeleting(false);
    }
  };

  // Handle cancel
  const handleCancel = () => {
    navigate(`/collections/${collectionId}`);
  };

  // Loading state
  if (loading) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Loading Collection...</h1>
        <p>Please wait while we load the collection details...</p>
      </div>
    );
  }

  // Error state
  if (!collection) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Collection Not Found</h1>
        <p style={{ color: "red" }}>{error || "Collection not found"}</p>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
      </div>
    );
  }

  // Permission error
  if (error && error.includes("owner")) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Permission Denied</h1>
        <p style={{ color: "red" }}>{error}</p>
        <button onClick={() => navigate(`/collections/${collectionId}`)}>
          ← Back to Collection
        </button>
      </div>
    );
  }

  // Final confirmation dialog
  if (showFinalConfirm) {
    return (
      <div style={{ padding: "20px", maxWidth: "600px" }}>
        <h1>⚠️ Final Confirmation</h1>

        <div
          style={{
            backgroundColor: "#fff3cd",
            border: "1px solid #ffeaa7",
            padding: "20px",
            marginBottom: "20px",
          }}
        >
          <h2 style={{ color: "#856404", marginTop: 0 }}>
            Are you absolutely sure?
          </h2>
          <p style={{ fontSize: "16px", marginBottom: "10px" }}>
            This action <strong>CANNOT</strong> be undone. This will permanently
            delete:
          </p>
          <ul style={{ fontSize: "16px" }}>
            <li>
              The collection <strong>"{collection.name}"</strong>
            </li>
            <li>
              All <strong>{files.length}</strong> files in this collection
            </li>
            <li>All subcollections and their contents</li>
            <li>All sharing permissions</li>
          </ul>
          <p style={{ fontSize: "16px", marginTop: "15px", color: "#721c24" }}>
            The collection will be soft-deleted and recoverable for 30 days, but
            this should not be relied upon.
          </p>
        </div>

        {error && (
          <div style={{ color: "red", marginBottom: "10px" }}>
            Error: {error}
          </div>
        )}

        <div style={{ marginTop: "20px" }}>
          <button
            onClick={handleFinalDelete}
            disabled={deleting}
            style={{
              backgroundColor: "#dc3545",
              color: "white",
              border: "none",
              padding: "10px 20px",
              fontSize: "16px",
              cursor: deleting ? "not-allowed" : "pointer",
            }}
          >
            {deleting ? "Deleting..." : "Yes, Delete Everything"}
          </button>
          <button
            onClick={() => setShowFinalConfirm(false)}
            disabled={deleting}
            style={{
              marginLeft: "10px",
              padding: "10px 20px",
              fontSize: "16px",
            }}
          >
            Cancel
          </button>
        </div>
      </div>
    );
  }

  // Main delete form
  return (
    <div style={{ padding: "20px", maxWidth: "800px" }}>
      <h1>Delete Collection</h1>

      <div style={{ marginBottom: "20px" }}>
        <button onClick={() => navigate(`/collections/${collectionId}`)}>
          ← Back to Collection
        </button>
      </div>

      {/* Collection info */}
      <div
        style={{
          backgroundColor: "#f8f9fa",
          padding: "20px",
          marginBottom: "20px",
          border: "1px solid #ddd",
        }}
      >
        <h3 style={{ marginTop: 0 }}>Collection Information</h3>
        <div style={{ fontSize: "16px" }}>
          <p>
            <strong>Name:</strong> {collection.name || "[Encrypted]"}
          </p>
          <p>
            <strong>Type:</strong>{" "}
            {collection.collection_type === "album" ? "Album 🖼️" : "Folder 📁"}
          </p>
          <p>
            <strong>Created:</strong>{" "}
            {collection.created_at
              ? new Date(collection.created_at).toLocaleString()
              : "Unknown"}
          </p>
          <p>
            <strong>Files:</strong>{" "}
            {loadingFiles ? "Loading..." : `${files.length} files`}
          </p>
          {collection.members && collection.members.length > 1 && (
            <p>
              <strong>Shared with:</strong> {collection.members.length - 1}{" "}
              user(s)
            </p>
          )}
        </div>
      </div>

      {/* Warning box */}
      <div
        style={{
          backgroundColor: "#f8d7da",
          border: "1px solid #f5c6cb",
          padding: "20px",
          marginBottom: "20px",
          borderRadius: "4px",
        }}
      >
        <h3 style={{ color: "#721c24", marginTop: 0 }}>
          ⚠️ Warning: This action cannot be undone!
        </h3>
        <p style={{ color: "#721c24", fontSize: "16px" }}>
          Deleting this collection will:
        </p>
        <ul style={{ color: "#721c24", fontSize: "16px" }}>
          <li>Permanently delete the collection and all its files</li>
          <li>Delete all subcollections and their contents</li>
          <li>Remove access for all shared users</li>
          <li>Free up the storage space used by encrypted files</li>
        </ul>
        {files.length > 0 && (
          <p style={{ color: "#721c24", fontSize: "16px", fontWeight: "bold" }}>
            This will delete {files.length} file(s) in this collection!
          </p>
        )}
      </div>

      {/* Confirmation form */}
      <div style={{ marginTop: "30px" }}>
        <h3>Confirm Deletion</h3>
        <p>
          To confirm deletion, please type{" "}
          <strong style={{ backgroundColor: "#f8f9fa", padding: "2px 5px" }}>
            DELETE {collection.name}
          </strong>{" "}
          in the box below:
        </p>

        {error && (
          <div style={{ color: "red", marginBottom: "10px" }}>
            Error: {error}
          </div>
        )}

        <input
          type="text"
          value={confirmText}
          onChange={handleConfirmTextChange}
          placeholder={`Type "DELETE ${collection.name}" to confirm`}
          disabled={deleting}
          style={{
            width: "100%",
            padding: "10px",
            fontSize: "16px",
            marginBottom: "20px",
            border: error ? "1px solid red" : "1px solid #ccc",
          }}
        />

        <div>
          <button
            onClick={handleInitialDelete}
            disabled={deleting || confirmText !== `DELETE ${collection.name}`}
            style={{
              backgroundColor:
                confirmText === `DELETE ${collection.name}`
                  ? "#dc3545"
                  : "#6c757d",
              color: "white",
              border: "none",
              padding: "10px 20px",
              fontSize: "16px",
              cursor:
                deleting || confirmText !== `DELETE ${collection.name}`
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            Delete Collection
          </button>
          <button
            onClick={handleCancel}
            disabled={deleting}
            style={{
              marginLeft: "10px",
              padding: "10px 20px",
              fontSize: "16px",
            }}
          >
            Cancel
          </button>
        </div>
      </div>

      {/* Additional info */}
      <div
        style={{
          marginTop: "40px",
          padding: "20px",
          backgroundColor: "#e9ecef",
          border: "1px solid #dee2e6",
          borderRadius: "4px",
        }}
      >
        <h4>📌 Important Notes:</h4>
        <ul>
          <li>
            Collections are soft-deleted and can be recovered within 30 days
          </li>
          <li>However, you should not rely on this recovery mechanism</li>
          <li>Encrypted file data will be permanently removed from storage</li>
          <li>This action will be logged for security purposes</li>
        </ul>
      </div>

      {/* Debug info in development */}
      {import.meta.env.DEV && (
        <details
          style={{
            marginTop: "40px",
            padding: "10px",
            backgroundColor: "#f8f9fa",
          }}
        >
          <summary>🔍 Debug Information</summary>
          <div>
            <h4>Collection Data:</h4>
            <pre>
              {JSON.stringify(
                {
                  id: collection.id,
                  name: collection.name,
                  collection_type: collection.collection_type,
                  owner_id: collection.owner_id,
                  members_count: collection.members?.length || 0,
                  files_count: files.length,
                },
                null,
                2,
              )}
            </pre>
            <h4>Delete State:</h4>
            <pre>
              {JSON.stringify(
                {
                  confirmText,
                  expectedText: `DELETE ${collection.name}`,
                  canDelete: confirmText === `DELETE ${collection.name}`,
                  deleting,
                  showFinalConfirm,
                  isAuthenticated,
                },
                null,
                2,
              )}
            </pre>
          </div>
        </details>
      )}
    </div>
  );
};

export default withPasswordProtection(CollectionDelete);
