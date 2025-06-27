// File: src/pages/User/Collection/Detail.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useParams, useNavigate, useLocation } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";
import useFiles from "../../../hooks/useFiles.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const CollectionDetail = () => {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const location = useLocation();

  const {
    collectionService,
    cryptoService,
    passwordStorageService,
    localStorageService,
  } = useServices();

  const { isAuthenticated } = useAuth();
  // Don't use useCollections hook to avoid automatic loading without session keys
  const COLLECTION_TYPES = {
    FOLDER: "folder",
    ALBUM: "album",
  };

  const COLLECTION_STATES = {
    ACTIVE: "active",
    DELETED: "deleted",
    ARCHIVED: "archived",
  };

  const PERMISSION_LEVELS = {
    READ_ONLY: "read_only",
    READ_WRITE: "read_write",
    ADMIN: "admin",
  };

  // Load files manually when needed
  const loadFiles = async () => {
    if (!collectionId) return;

    try {
      setFilesLoading(true);
      setFilesError("");

      const { fileService } = await import("../../../services/FileService.js");
      const fileList = await fileService.listFilesByCollection(collectionId);
      setFiles(fileList || []);

      console.log("[CollectionDetail] Files loaded:", fileList?.length || 0);
    } catch (error) {
      console.error("[CollectionDetail] Failed to load files:", error);
      setFilesError(error.message);
    } finally {
      setFilesLoading(false);
    }
  };

  // State management
  const [collection, setCollection] = useState(null);
  const [hierarchy, setHierarchy] = useState([]);
  const [childCollections, setChildCollections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");

  // Password prompt state
  const [showPasswordPrompt, setShowPasswordPrompt] = useState(false);
  const [password, setPassword] = useState("");
  const [passwordError, setPasswordError] = useState("");

  // Edit state
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState({
    name: "",
    collection_type: "folder",
  });

  // Share state
  const [showShareDialog, setShowShareDialog] = useState(false);
  const [shareForm, setShareForm] = useState({
    recipient_email: "",
    permission_level: "read_only",
    share_with_descendants: false,
  });

  // Move state
  const [showMoveDialog, setShowMoveDialog] = useState(false);
  const [moveForm, setMoveForm] = useState({
    new_parent_id: "",
    parent_name: "",
  });

  // Load collection data with proper session management
  const loadCollectionData = useCallback(
    async (passwordParam = null, forceRefresh = false) => {
      if (!collectionId) {
        setError("No collection ID provided");
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        setError("");

        // Check for stored password if no parameter provided
        let passwordToUse = passwordParam;
        if (!passwordToUse) {
          passwordToUse = passwordStorageService.getPassword();
        }

        console.log("[CollectionDetail] Loading collection:", collectionId, {
          hasPasswordParam: !!passwordParam,
          hasStoredPassword: !!passwordToUse,
          hasUserEncryptedData: localStorageService.hasUserEncryptedData(),
          hasSessionKeys: localStorageService.hasSessionKeys(),
        });

        // Prevent competing operations during navigation
        console.log("[CollectionDetail] About to load collection");

        // Load collection with password directly using the same method as Collections List
        const collectionData = await collectionService.getCollection(
          collectionId,
          passwordToUse,
        );

        if (
          collectionData.decrypt_error &&
          !passwordParam &&
          !passwordStorageService.hasPassword()
        ) {
          console.log(
            "[CollectionDetail] Collection failed to decrypt, need password",
          );
          setShowPasswordPrompt(true);
          return;
        }

        setCollection(collectionData);

        // Set edit form with current data
        setEditForm({
          name: collectionData.name || "[Encrypted]",
          collection_type: collectionData.collection_type || "folder",
        });

        // Skip hierarchy and child loading to prevent competing operations for now
        setHierarchy([collectionData]);
        setChildCollections([]);

        // Skip file loading during initial navigation to prevent conflicts
        console.log(
          "[CollectionDetail] Skipping files during navigation to prevent conflicts",
        );

        console.log("[CollectionDetail] Collection data loaded successfully", {
          collectionId: collectionData.id,
          collectionName: collectionData.name,
          hasDecryptError: !!collectionData.decrypt_error,
          collectionState: collectionData.state || "unknown",
        });
      } catch (err) {
        console.error("[CollectionDetail] Failed to load collection:", err);

        if (
          err.message.includes("Password required") ||
          err.message.includes("session keys not available")
        ) {
          setShowPasswordPrompt(true);
          setError("");
        } else if (err.message.includes("403")) {
          setError("You don't have permission to access this collection");
        } else if (err.message.includes("404")) {
          setError("Collection not found");
        } else {
          setError(err.message || "Failed to load collection");
        }
      } finally {
        setLoading(false);
      }
    },
    [collectionId, collectionService, passwordStorageService],
  );

  // Initial load with resilient password handling
  useEffect(() => {
    if (!isAuthenticated) {
      navigate("/login");
      return;
    }

    if (collectionId) {
      // Ensure password persists during navigation
      const storedPassword = passwordStorageService.getPassword();

      if (storedPassword) {
        console.log(
          "[CollectionDetail] Using stored password to load collection",
        );
        // Re-store password to ensure it persists through navigation conflicts
        passwordStorageService.setPassword(storedPassword);

        // Add a small delay to let competing operations settle
        setTimeout(() => {
          loadCollectionData(storedPassword);
        }, 100);
      } else if (localStorageService.hasUserEncryptedData()) {
        // Need password
        setShowPasswordPrompt(true);
        setLoading(false);
      } else {
        loadCollectionData();
      }
    }
  }, [
    collectionId,
    isAuthenticated,
    navigate,
    loadCollectionData,
    passwordStorageService,
    localStorageService,
  ]);

  // Handle success message from navigation state
  useEffect(() => {
    if (location.state?.message) {
      setSuccessMessage(location.state.message);
      const timer = setTimeout(() => setSuccessMessage(""), 5000);
      return () => clearTimeout(timer);
    }
  }, [location.state]);

  // Handle password submission
  const handlePasswordSubmit = async (e) => {
    e.preventDefault();

    if (!password) {
      setPasswordError("Password is required");
      return;
    }

    setPasswordError("");

    try {
      console.log("[CollectionDetail] Loading collection with password");
      await loadCollectionData(password, true);

      // Success - hide password prompt
      setShowPasswordPrompt(false);
      setPassword("");
    } catch (err) {
      console.error("[CollectionDetail] Failed to decrypt with password:", err);
      setPasswordError("Invalid password. Please try again.");
      setPassword("");
    }
  };

  // Handle collection update
  const handleUpdateCollection = async (e) => {
    e.preventDefault();

    if (!editForm.name.trim()) {
      setError("Collection name is required");
      return;
    }

    try {
      setLoading(true);
      setError("");

      const updateData = {
        name: editForm.name.trim(),
        collection_type: editForm.collection_type,
        version: collection.version || 1,
      };

      const updatedCollection = await collectionService.updateCollection(
        collection.id,
        updateData,
      );
      setCollection(updatedCollection);
      setIsEditing(false);
      setSuccessMessage("Collection updated successfully");
    } catch (err) {
      console.error("[CollectionDetail] Failed to update collection:", err);
      setError(err.message || "Failed to update collection");
    } finally {
      setLoading(false);
    }
  };

  // Handle collection deletion
  const handleDeleteCollection = async () => {
    if (
      !window.confirm(
        `Are you sure you want to delete "${collection.name}"? This action cannot be undone.`,
      )
    ) {
      return;
    }

    try {
      setLoading(true);
      await collectionService.deleteCollection(collection.id);
      navigate("/collections", {
        state: {
          message: `Collection "${collection.name}" deleted successfully`,
        },
      });
    } catch (err) {
      console.error("[CollectionDetail] Failed to delete collection:", err);
      setError(err.message || "Failed to delete collection");
      setLoading(false);
    }
  };

  // Handle archive/restore
  const handleArchiveToggle = async () => {
    const isArchived = collection.state === COLLECTION_STATES.ARCHIVED;
    const action = isArchived ? "restore" : "archive";

    if (
      !window.confirm(
        `Are you sure you want to ${action} "${collection.name}"?`,
      )
    ) {
      return;
    }

    try {
      setLoading(true);

      if (isArchived) {
        await collectionService.restoreCollection(collection.id);
        setSuccessMessage("Collection restored successfully");
      } else {
        await collectionService.archiveCollection(collection.id);
        setSuccessMessage("Collection archived successfully");
      }

      // Reload collection data
      await loadCollectionData();
    } catch (err) {
      console.error(`[CollectionDetail] Failed to ${action} collection:`, err);
      setError(err.message || `Failed to ${action} collection`);
    } finally {
      setLoading(false);
    }
  };

  // Handle share collection
  const handleShareCollection = async (e) => {
    e.preventDefault();

    if (!shareForm.recipient_email.trim()) {
      setError("Recipient email is required");
      return;
    }

    try {
      setLoading(true);
      setError("");

      // Note: In a real implementation, you'd need to get the recipient's public key
      // For now, this is a placeholder showing the structure
      const shareData = {
        recipient_email: shareForm.recipient_email.trim(),
        permission_level: shareForm.permission_level,
        share_with_descendants: shareForm.share_with_descendants,
        // recipient_public_key: await getUserPublicKey(shareForm.recipient_email)
      };

      await collectionService.shareCollection(collection.id, shareData);
      setShowShareDialog(false);
      setShareForm({
        recipient_email: "",
        permission_level: "read_only",
        share_with_descendants: false,
      });
      setSuccessMessage("Collection shared successfully");

      // Reload to show new member
      await loadCollectionData();
    } catch (err) {
      console.error("[CollectionDetail] Failed to share collection:", err);
      setError(err.message || "Failed to share collection");
    } finally {
      setLoading(false);
    }
  };

  // Handle remove member
  const handleRemoveMember = async (memberId, memberEmail) => {
    if (!window.confirm(`Remove ${memberEmail} from this collection?`)) {
      return;
    }

    try {
      setLoading(true);
      await collectionService.removeMember(collection.id, memberId, true);
      setSuccessMessage(`${memberEmail} removed from collection`);
      await loadCollectionData();
    } catch (err) {
      console.error("[CollectionDetail] Failed to remove member:", err);
      setError(err.message || "Failed to remove member");
    } finally {
      setLoading(false);
    }
  };

  // Handle navigation
  const handleNavigateToCollection = (targetCollectionId) => {
    navigate(`/collections/${targetCollectionId}`);
  };

  const handleNavigateToParent = () => {
    if (collection.parent_id) {
      navigate(`/collections/${collection.parent_id}`);
    } else {
      navigate("/collections");
    }
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "Unknown";
    return new Date(dateString).toLocaleString();
  };

  // Get permission level display
  const getPermissionDisplay = (level) => {
    switch (level) {
      case PERMISSION_LEVELS.READ_ONLY:
        return "Read Only";
      case PERMISSION_LEVELS.READ_WRITE:
        return "Read & Write";
      case PERMISSION_LEVELS.ADMIN:
        return "Admin";
      default:
        return level;
    }
  };

  // Get collection type icon
  const getCollectionIcon = (type) => {
    return type === COLLECTION_TYPES.ALBUM ? "🖼️" : "📁";
  };

  // Get state badge
  const getStateBadge = (state) => {
    switch (state) {
      case COLLECTION_STATES.ACTIVE:
        return { text: "Active", color: "green" };
      case COLLECTION_STATES.ARCHIVED:
        return { text: "Archived", color: "orange" };
      case COLLECTION_STATES.DELETED:
        return { text: "Deleted", color: "red" };
      default:
        return { text: state, color: "gray" };
    }
  };

  // Check if user can edit (is owner or admin)
  const canEdit =
    collection &&
    (collection.owner_id === collection.current_user_id ||
      collection.members?.some(
        (m) =>
          m.recipient_id === collection.current_user_id &&
          m.permission_level === PERMISSION_LEVELS.ADMIN,
      ));

  // Show password prompt if needed
  if (showPasswordPrompt && !loading) {
    return (
      <div style={{ padding: "20px", maxWidth: "600px" }}>
        <h1>Enter Password to Decrypt Collection</h1>

        <p>
          This collection is encrypted. Please enter your password to decrypt
          it.
        </p>

        {passwordError && (
          <div style={{ color: "red", marginBottom: "10px" }}>
            ❌ {passwordError}
          </div>
        )}

        <form onSubmit={handlePasswordSubmit}>
          <div style={{ marginBottom: "15px" }}>
            <label htmlFor="password">Password:</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              autoFocus
              style={{ marginLeft: "10px", width: "200px" }}
            />
          </div>

          <div>
            <button type="submit" disabled={!password}>
              Decrypt Collection
            </button>
            <button
              type="button"
              onClick={() => navigate("/collections")}
              style={{ marginLeft: "10px" }}
            >
              Back to Collections
            </button>
          </div>
        </form>

        <div
          style={{
            marginTop: "20px",
            padding: "10px",
            backgroundColor: "#f8f9fa",
          }}
        >
          <p>
            🔐 Your password is never stored and is only used to decrypt your
            data.
          </p>
        </div>
      </div>
    );
  }

  // Show loading state
  if (loading) {
    return (
      <div style={{ padding: "20px", maxWidth: "1200px" }}>
        {/* Debug info at top */}
        {import.meta.env.DEV && (
          <div
            style={{
              backgroundColor: "#d4edda",
              padding: "10px",
              marginBottom: "20px",
              fontSize: "12px",
              border: "1px solid #28a745",
            }}
          >
            ✅ DEBUG: Collection Detail Page Loaded Successfully! Collection:{" "}
            {collection?.name || "[Encrypted]"}
          </div>
        )}
        {/* Debug info */}
        {import.meta.env.DEV && (
          <div
            style={{
              backgroundColor: "#fff3cd",
              padding: "10px",
              marginBottom: "20px",
            }}
          >
            🔍 DEBUG: Component is in loading state
          </div>
        )}
        <h1>Loading Collection...</h1>
        <p>Decrypting collection data...</p>
        <div style={{ marginTop: "20px" }}>
          <button onClick={() => navigate("/collections")}>
            ← Back to Collections (Cancel Loading)
          </button>
        </div>
      </div>
    );
  }

  // Show error state
  if (error && !collection) {
    return (
      <div style={{ padding: "20px", maxWidth: "1200px" }}>
        {/* Debug info */}
        {import.meta.env.DEV && (
          <div
            style={{
              backgroundColor: "#f8d7da",
              padding: "10px",
              marginBottom: "20px",
            }}
          >
            🔍 DEBUG: Component has error state: {error}
          </div>
        )}
        <h1>Error Loading Collection</h1>
        <div style={{ color: "red", marginBottom: "20px" }}>❌ {error}</div>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
        <button
          onClick={() => window.location.reload()}
          style={{ marginLeft: "10px" }}
        >
          🔄 Retry
        </button>
      </div>
    );
  }

  // Show collection not found
  if (!collection && !loading && !showPasswordPrompt) {
    return (
      <div style={{ padding: "20px", maxWidth: "1200px" }}>
        {/* Debug info */}
        {import.meta.env.DEV && (
          <div
            style={{
              backgroundColor: "#f8d7da",
              padding: "10px",
              marginBottom: "20px",
            }}
          >
            🔍 DEBUG: No collection found, not loading, not showing password
            prompt
          </div>
        )}
        <h1>Collection Not Found</h1>
        <p>
          The requested collection could not be found or you don't have access
          to it.
        </p>
        <div style={{ marginTop: "20px" }}>
          <button onClick={() => navigate("/collections")}>
            ← Back to Collections
          </button>
          <button
            onClick={() => window.location.reload()}
            style={{ marginLeft: "10px" }}
          >
            🔄 Retry Loading
          </button>
        </div>
      </div>
    );
  }

  const stateBadge = getStateBadge(collection.state);

  // Debug what we're about to render
  console.log("[CollectionDetail] About to render component", {
    loading,
    error,
    hasCollection: !!collection,
    collectionName: collection?.name,
    showPasswordPrompt,
    isAuthenticated,
    collectionId,
  });

  return (
    <div style={{ padding: "20px", maxWidth: "1200px" }}>
      {/* Debug info at top */}
      {import.meta.env.DEV && (
        <div
          style={{
            backgroundColor: "#e3f2fd",
            padding: "10px",
            marginBottom: "20px",
            fontSize: "12px",
            border: "1px solid #2196f3",
          }}
        >
          🔍 DEBUG: loading={loading.toString()}, error={error || "none"},
          hasCollection={!!collection}, showPasswordPrompt=
          {showPasswordPrompt.toString()}
        </div>
      )}
      {/* Header */}
      <div style={{ marginBottom: "20px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <button onClick={handleNavigateToParent}>
            ← Back to{" "}
            {collection.parent_id ? "Parent Collection" : "Collections"}
          </button>
          <div>
            <button
              onClick={() => window.location.reload()}
              style={{ marginRight: "10px" }}
            >
              🔄 Refresh
            </button>
            <button onClick={() => navigate("/collections")}>
              📋 All Collections
            </button>
          </div>
        </div>

        {/* Breadcrumb */}
        {hierarchy.length > 0 && (
          <div
            style={{ fontSize: "14px", color: "#666", marginBottom: "10px" }}
          >
            <span>Path: </span>
            {hierarchy.map((item, index) => (
              <span key={item.id || index}>
                {index > 0 && <span> → </span>}
                {index < hierarchy.length - 1 ? (
                  <button
                    onClick={() => handleNavigateToCollection(item.id)}
                    style={{
                      background: "none",
                      border: "none",
                      color: "#007bff",
                      textDecoration: "underline",
                      cursor: "pointer",
                    }}
                  >
                    {item.name || "[Encrypted]"}
                  </button>
                ) : (
                  <strong>{item.name || "[Encrypted]"}</strong>
                )}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Success/Error Messages */}
      {successMessage && (
        <div style={{ color: "green", marginBottom: "10px" }}>
          ✅ {successMessage}
        </div>
      )}

      {error && (
        <div style={{ color: "red", marginBottom: "10px" }}>❌ {error}</div>
      )}

      {/* Collection Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          marginBottom: "30px",
          padding: "20px",
          backgroundColor: "#f8f9fa",
          border: "1px solid #ddd",
        }}
      >
        <span style={{ fontSize: "48px", marginRight: "20px" }}>
          {getCollectionIcon(collection.collection_type)}
        </span>
        <div style={{ flex: 1 }}>
          {isEditing ? (
            <form
              onSubmit={handleUpdateCollection}
              style={{ display: "flex", alignItems: "center", gap: "10px" }}
            >
              <input
                type="text"
                value={editForm.name}
                onChange={(e) =>
                  setEditForm({ ...editForm, name: e.target.value })
                }
                style={{ fontSize: "24px", padding: "5px", minWidth: "300px" }}
                autoFocus
              />
              <select
                value={editForm.collection_type}
                onChange={(e) =>
                  setEditForm({ ...editForm, collection_type: e.target.value })
                }
                style={{ padding: "5px" }}
              >
                <option value="folder">Folder</option>
                <option value="album">Album</option>
              </select>
              <button type="submit" disabled={loading}>
                💾 Save
              </button>
              <button
                type="button"
                onClick={() => {
                  setIsEditing(false);
                  setEditForm({
                    name: collection.name || "[Encrypted]",
                    collection_type: collection.collection_type || "folder",
                  });
                }}
              >
                ❌ Cancel
              </button>
            </form>
          ) : (
            <>
              <h1 style={{ margin: "0 0 10px 0", fontSize: "32px" }}>
                {collection.name || "[Encrypted]"}
                {collection.decrypt_error && " 🔒"}
              </h1>
              <div style={{ fontSize: "16px", color: "#666" }}>
                {collection.collection_type} • {stateBadge.text} • Modified:{" "}
                {formatDate(collection.modified_at)}
              </div>
              {collection.decrypt_error && (
                <div
                  style={{ color: "red", fontSize: "14px", marginTop: "5px" }}
                >
                  Unable to decrypt: {collection.decrypt_error}
                </div>
              )}
            </>
          )}
        </div>

        {/* Action Buttons */}
        {canEdit && !isEditing && !collection.decrypt_error && (
          <div style={{ display: "flex", gap: "10px" }}>
            <button onClick={() => setIsEditing(true)}>✏️ Edit</button>
            <button onClick={handleArchiveToggle}>
              {collection.state === COLLECTION_STATES.ARCHIVED
                ? "📤 Restore"
                : "📥 Archive"}
            </button>
            <button onClick={() => setShowShareDialog(true)}>🔗 Share</button>
            <button onClick={handleDeleteCollection} style={{ color: "red" }}>
              🗑️ Delete
            </button>
          </div>
        )}
      </div>

      {/* Collection Details */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: "30px",
          marginBottom: "30px",
        }}
      >
        {/* Basic Info */}
        <div>
          <h3>Collection Information</h3>
          <div
            style={{
              backgroundColor: "#f8f9fa",
              padding: "15px",
              border: "1px solid #ddd",
            }}
          >
            <div style={{ marginBottom: "10px" }}>
              <strong>ID:</strong> {collection.id}
            </div>
            <div style={{ marginBottom: "10px" }}>
              <strong>Type:</strong> {collection.collection_type || "Unknown"}
            </div>
            <div style={{ marginBottom: "10px" }}>
              <strong>State:</strong>
              <span
                style={{
                  color: stateBadge.color,
                  fontWeight: "bold",
                  marginLeft: "5px",
                }}
              >
                {stateBadge.text}
              </span>
            </div>
            <div style={{ marginBottom: "10px" }}>
              <strong>Created:</strong> {formatDate(collection.created_at)}
            </div>
            <div style={{ marginBottom: "10px" }}>
              <strong>Modified:</strong> {formatDate(collection.modified_at)}
            </div>
            <div>
              <strong>Files:</strong> {files.length} files
            </div>
          </div>
        </div>

        {/* Members & Sharing */}
        <div>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <h3>Members & Sharing</h3>
            {canEdit && (
              <button onClick={() => setShowShareDialog(true)}>
                + Add Member
              </button>
            )}
          </div>
          <div
            style={{
              backgroundColor: "#f8f9fa",
              padding: "15px",
              border: "1px solid #ddd",
            }}
          >
            {collection.members && collection.members.length > 0 ? (
              collection.members.map((member) => (
                <div
                  key={member.id}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    padding: "8px 0",
                    borderBottom: "1px solid #eee",
                  }}
                >
                  <div>
                    <div style={{ fontWeight: "bold" }}>
                      {member.recipient_email}
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      {getPermissionDisplay(member.permission_level)}
                      {member.is_inherited && " (Inherited)"}
                    </div>
                  </div>
                  {canEdit &&
                    member.permission_level !== PERMISSION_LEVELS.ADMIN && (
                      <button
                        onClick={() =>
                          handleRemoveMember(
                            member.recipient_id,
                            member.recipient_email,
                          )
                        }
                        style={{ color: "red", fontSize: "12px" }}
                      >
                        Remove
                      </button>
                    )}
                </div>
              ))
            ) : (
              <p style={{ color: "#666", fontStyle: "italic" }}>
                No members. This collection is private.
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Child Collections */}
      {childCollections.length > 0 && (
        <div style={{ marginBottom: "30px" }}>
          <h3>Subfolders ({childCollections.length})</h3>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(250px, 1fr))",
              gap: "15px",
            }}
          >
            {childCollections.map((child) => (
              <div
                key={child.id}
                onClick={() => handleNavigateToCollection(child.id)}
                style={{
                  padding: "15px",
                  backgroundColor: "#f8f9fa",
                  border: "1px solid #ddd",
                  cursor: "pointer",
                  borderRadius: "4px",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    marginBottom: "5px",
                  }}
                >
                  <span style={{ fontSize: "20px", marginRight: "10px" }}>
                    {getCollectionIcon(child.collection_type)}
                  </span>
                  <strong>{child.name || "[Encrypted]"}</strong>
                </div>
                <div style={{ fontSize: "12px", color: "#666" }}>
                  {child.collection_type} • Modified:{" "}
                  {formatDate(child.modified_at)}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Files */}
      <div style={{ marginBottom: "30px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <h3>Files ({files.length})</h3>
          <div>
            <button onClick={loadFiles} style={{ marginRight: "10px" }}>
              📁 Load Files
            </button>
            <button
              onClick={() => navigate(`/collections/${collectionId}/upload`)}
            >
              📁 Upload Files
            </button>
          </div>
        </div>

        {filesLoading ? (
          <p>Loading files...</p>
        ) : filesError ? (
          <div style={{ color: "red" }}>
            ❌ Error loading files: {filesError}
          </div>
        ) : files.length > 0 ? (
          <div
            style={{
              backgroundColor: "#f8f9fa",
              padding: "15px",
              border: "1px solid #ddd",
            }}
          >
            {files.map((file) => (
              <div
                key={file.id}
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  padding: "10px 0",
                  borderBottom: "1px solid #eee",
                }}
              >
                <div>
                  <div style={{ fontWeight: "bold" }}>
                    {file.name || "[Encrypted]"}
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    {file.file_type} •{" "}
                    {file.file_size_in_bytes
                      ? `${Math.round(file.file_size_in_bytes / 1024)}KB`
                      : "Unknown size"}{" "}
                    • {formatDate(file.created_at)}
                  </div>
                </div>
                <div>
                  <button onClick={() => navigate(`/files/${file.id}`)}>
                    👁️ View
                  </button>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div
            style={{
              backgroundColor: "#f8f9fa",
              padding: "40px",
              border: "1px solid #ddd",
              textAlign: "center",
              color: "#666",
            }}
          >
            <p>This collection is empty.</p>
            <button
              onClick={() => navigate(`/collections/${collectionId}/upload`)}
            >
              Upload your first file
            </button>
          </div>
        )}
      </div>

      {/* Share Dialog */}
      {showShareDialog && (
        <div
          style={{
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: "rgba(0, 0, 0, 0.5)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: "white",
              padding: "30px",
              maxWidth: "500px",
              width: "90%",
              border: "1px solid #ddd",
            }}
          >
            <h3>Share Collection</h3>
            <form onSubmit={handleShareCollection}>
              <div style={{ marginBottom: "15px" }}>
                <label htmlFor="recipient_email">Email Address:</label>
                <input
                  type="email"
                  id="recipient_email"
                  value={shareForm.recipient_email}
                  onChange={(e) =>
                    setShareForm({
                      ...shareForm,
                      recipient_email: e.target.value,
                    })
                  }
                  placeholder="user@example.com"
                  required
                  style={{ width: "100%", padding: "8px", marginTop: "5px" }}
                />
              </div>

              <div style={{ marginBottom: "15px" }}>
                <label htmlFor="permission_level">Permission Level:</label>
                <select
                  id="permission_level"
                  value={shareForm.permission_level}
                  onChange={(e) =>
                    setShareForm({
                      ...shareForm,
                      permission_level: e.target.value,
                    })
                  }
                  style={{ width: "100%", padding: "8px", marginTop: "5px" }}
                >
                  <option value="read_only">Read Only</option>
                  <option value="read_write">Read & Write</option>
                  <option value="admin">Admin</option>
                </select>
              </div>

              <div style={{ marginBottom: "20px" }}>
                <label>
                  <input
                    type="checkbox"
                    checked={shareForm.share_with_descendants}
                    onChange={(e) =>
                      setShareForm({
                        ...shareForm,
                        share_with_descendants: e.target.checked,
                      })
                    }
                    style={{ marginRight: "8px" }}
                  />
                  Share with subfolders
                </label>
              </div>

              <div style={{ display: "flex", gap: "10px" }}>
                <button type="submit" disabled={loading}>
                  Share Collection
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowShareDialog(false);
                    setShareForm({
                      recipient_email: "",
                      permission_level: "read_only",
                      share_with_descendants: false,
                    });
                  }}
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Debug Info (Development Only) */}
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
            <h4>Component State:</h4>
            <pre style={{ fontSize: "12px", overflow: "auto" }}>
              {JSON.stringify(
                {
                  loading,
                  error,
                  showPasswordPrompt,
                  isAuthenticated,
                  collectionId,
                  hasCollection: !!collection,
                  collectionName: collection?.name,
                  collectionType: collection?.collection_type,
                  hasDecryptError: !!collection?.decrypt_error,
                  filesCount: files.length,
                  hierarchyCount: hierarchy.length,
                  childCollectionsCount: childCollections.length,
                  canEdit,
                  hasPassword: passwordStorageService.hasPassword(),
                  hasUserEncryptedData:
                    localStorageService.hasUserEncryptedData(),
                  hasSessionKeys: localStorageService.hasSessionKeys(),
                },
                null,
                2,
              )}
            </pre>

            <h4>Collection Data:</h4>
            <pre style={{ fontSize: "12px", overflow: "auto" }}>
              {JSON.stringify(collection, null, 2)}
            </pre>
          </div>
        </details>
      )}
    </div>
  );
};

export default withPasswordProtection(CollectionDetail);
