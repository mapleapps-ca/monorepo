// File: src/pages/User/FileManager/FileManager.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useParams } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";
import useCollections from "../../../hooks/useCollections.js";
import useFiles from "../../../hooks/useFiles.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const FileManager = () => {
  const navigate = useNavigate();
  const { folderId } = useParams(); // Current folder (collection) ID from URL
  const { collectionService, fileService, passwordStorageService } =
    useServices();
  const { isAuthenticated } = useAuth();

  // Hooks
  const {
    collections,
    sharedCollections,
    isLoading: collectionsLoading,
    error: collectionsError,
    loadAllCollections,
    deleteCollection,
  } = useCollections();

  const {
    files,
    isLoading: filesLoading,
    error: filesError,
    loadFilesByCollection,
    downloadAndSaveFile,
    deleteFile,
    archiveFile,
  } = useFiles(folderId);

  // State
  const [currentFolder, setCurrentFolder] = useState(null);
  const [breadcrumbs, setBreadcrumbs] = useState([]);
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [viewMode, setViewMode] = useState("list"); // list or grid
  const [sortBy, setSortBy] = useState("name"); // name, date, type
  const [filterType, setFilterType] = useState("all"); // all, folders, files
  const [needsPassword, setNeedsPassword] = useState(false);
  const [loadAttempted, setLoadAttempted] = useState(false);
  const [, forceUpdate] = useState({});

  const hasCollections = collections.length > 0 || sharedCollections.length > 0;

  // Debug - log collections when they change
  useEffect(() => {
    console.log("[FileManager] Collections updated:", {
      collectionsCount: collections.length,
      sharedCount: sharedCollections.length,
      firstCollection: collections[0],
      isLoading: collectionsLoading,
      loadAttempted,
      hasCollections,
    });

    // Force a re-render when collections change
    if (collections.length > 0 || sharedCollections.length > 0) {
      forceUpdate({});
    }
  }, [
    collections,
    sharedCollections,
    collectionsLoading,
    loadAttempted,
    hasCollections,
  ]);

  // Load collections on mount
  useEffect(() => {
    const loadCollectionsAsync = async () => {
      try {
        console.log("[FileManager] Component mounted, loading collections...");
        setNeedsPassword(false);
        setLoadAttempted(false);

        const result = await loadAllCollections();

        console.log("[FileManager] Collections loaded successfully", result);
        setLoadAttempted(true);
      } catch (err) {
        console.error("[FileManager] Failed to load collections:", err);
        setLoadAttempted(true);

        // Check if it's a password-related error
        if (
          err.message?.includes("Password required") ||
          err.message?.includes("session keys not available")
        ) {
          setNeedsPassword(true);
        }
      }
    };

    if (!loadAttempted) {
      loadCollectionsAsync();
    }
  }, [loadAttempted, loadAllCollections]); // Include loadAttempted to prevent re-runs

  // Check for decrypt errors when collections change
  useEffect(() => {
    const allCollections = [...collections, ...sharedCollections];
    const hasDecryptErrors = allCollections.some((c) => c.decrypt_error);
    if (hasDecryptErrors && !passwordStorageService.hasPassword()) {
      setNeedsPassword(true);
    } else {
      setNeedsPassword(false);
    }
  }, [collections, sharedCollections, passwordStorageService]);

  // Load current folder details and build breadcrumbs
  const loadCurrentFolder = useCallback(async () => {
    if (!folderId) {
      // Root level
      setCurrentFolder(null);
      setBreadcrumbs([{ id: null, name: "My Files", path: "/files" }]);
      return;
    }

    try {
      const password = passwordStorageService.getPassword();
      const folder = await collectionService.getCollection(folderId, password);
      setCurrentFolder(folder);

      // Build breadcrumbs
      const crumbs = [{ id: null, name: "My Files", path: "/files" }];

      // Get hierarchy if available
      if (folder.ancestor_ids && folder.ancestor_ids.length > 0) {
        // Load ancestors
        for (const ancestorId of folder.ancestor_ids) {
          try {
            const ancestor = await collectionService.getCollection(
              ancestorId,
              password,
            );
            crumbs.push({
              id: ancestor.id,
              name: ancestor.name || "[Encrypted]",
              path: `/files/${ancestor.id}`,
            });
          } catch (ancestorError) {
            console.warn("Failed to load ancestor:", ancestorId, ancestorError);
            crumbs.push({
              id: ancestorId,
              name: "[Encrypted]",
              path: `/files/${ancestorId}`,
            });
          }
        }
      }

      // Add current folder
      crumbs.push({
        id: folder.id,
        name: folder.name || "[Encrypted]",
        path: `/files/${folder.id}`,
      });

      setBreadcrumbs(crumbs);
    } catch (error) {
      console.error("Failed to load folder:", error);
      // Still set breadcrumbs to allow navigation back
      setBreadcrumbs([
        { id: null, name: "My Files", path: "/files" },
        {
          id: folderId,
          name: "[Unable to decrypt]",
          path: `/files/${folderId}`,
        },
      ]);
    }
  }, [folderId, collectionService, passwordStorageService]);

  // Load current folder and its contents when folderId changes
  useEffect(() => {
    if (folderId) {
      loadCurrentFolder();
      // Load files when entering a folder
      loadFilesByCollection(folderId, true);
    } else {
      // At root level, just set the breadcrumbs
      setCurrentFolder(null);
      setBreadcrumbs([{ id: null, name: "My Files", path: "/files" }]);
    }
  }, [folderId, loadCurrentFolder, loadFilesByCollection]);

  // Check if a collection is at root level
  const isRootLevel = useCallback((parentId) => {
    return (
      !parentId ||
      parentId === null ||
      parentId === "00000000-0000-0000-0000-000000000000"
    );
  }, []);

  // Get child collections (subfolders) of current folder
  const getSubfolders = useCallback(() => {
    // Combine owned and shared collections, removing duplicates
    const collectionMap = new Map();

    // Add owned collections
    collections.forEach((c) => collectionMap.set(c.id, c));

    // Add shared collections (will overwrite if duplicate)
    sharedCollections.forEach((c) => collectionMap.set(c.id, c));

    const allColls = Array.from(collectionMap.values());

    let result;
    if (!folderId) {
      // Root level - show collections that are at root
      result = allColls.filter((c) => isRootLevel(c.parent_id));
    } else {
      // Show collections with current folder as parent
      result = allColls.filter((c) => c.parent_id === folderId);
    }

    // Debug logging
    if (import.meta.env.DEV) {
      console.log("[FileManager] getSubfolders:", {
        folderId: folderId || "root",
        totalCollections: allColls.length,
        resultCount: result.length,
        sampleParentIds: allColls.slice(0, 3).map((c) => c.parent_id),
      });
    }

    return result;
  }, [collections, sharedCollections, folderId, isRootLevel]);

  // Get items to display (folders + files)
  const getDisplayItems = useCallback(() => {
    const subfolders = getSubfolders();

    const folders = subfolders.map((folder) => ({
      ...folder,
      itemType: "folder",
      displayName: folder.name || "[Encrypted]",
      displaySize: "-",
      displayDate: folder.modified_at || folder.created_at,
    }));

    const displayFiles = files.map((file) => ({
      ...file,
      itemType: "file",
      displayName: file.name || "[Encrypted]",
      displaySize: file.size || file.encrypted_file_size_in_bytes || 0,
      displayDate: file.modified_at || file.created_at,
    }));

    // Debug logging
    if (import.meta.env.DEV) {
      console.log("[FileManager] Display items debug:", {
        folderId,
        subfoldersCount: subfolders.length,
        foldersCount: folders.length,
        filesCount: displayFiles.length,
        folders: folders.map((f) => ({
          id: f.id,
          name: f.displayName,
          parent_id: f.parent_id,
        })),
        totalCollections: collections.length + sharedCollections.length,
        collectionsLoading,
        filesLoading,
      });
    }

    let items = [];

    switch (filterType) {
      case "folders":
        items = folders;
        break;
      case "files":
        items = displayFiles;
        break;
      default:
        items = [...folders, ...displayFiles];
    }

    // Sort items
    items.sort((a, b) => {
      // Folders first, then files
      if (a.itemType !== b.itemType) {
        return a.itemType === "folder" ? -1 : 1;
      }

      switch (sortBy) {
        case "date":
          return new Date(b.displayDate) - new Date(a.displayDate);
        case "type":
          if (a.itemType === "file" && b.itemType === "file") {
            return (a.mime_type || "").localeCompare(b.mime_type || "");
          }
          return 0;
        default: // name
          return a.displayName.localeCompare(b.displayName);
      }
    });

    return items;
  }, [
    files,
    getSubfolders,
    filterType,
    sortBy,
    collections,
    sharedCollections,
    folderId,
    collectionsLoading,
    filesLoading,
  ]);

  // Handle item click
  const handleItemClick = (item) => {
    if (item.itemType === "folder") {
      // Navigate into folder
      navigate(`/files/${item.id}`);
    } else {
      // For files, toggle selection
      handleItemSelect(item.id);
    }
  };

  // Handle item selection
  const handleItemSelect = (itemId) => {
    const newSelected = new Set(selectedItems);
    if (newSelected.has(itemId)) {
      newSelected.delete(itemId);
    } else {
      newSelected.add(itemId);
    }
    setSelectedItems(newSelected);
  };

  // Handle file download
  const handleDownloadFile = async (fileId) => {
    try {
      await downloadAndSaveFile(fileId);
    } catch (error) {
      alert("Failed to download file: " + error.message);
    }
  };

  // Handle file deletion
  const handleDeleteFile = async (fileId) => {
    if (!window.confirm("Are you sure you want to delete this file?")) return;

    try {
      await deleteFile(fileId);
      setSelectedItems((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    } catch (error) {
      alert("Failed to delete file: " + error.message);
    }
  };

  // Handle folder deletion
  const handleDeleteFolder = async (folderId, folderName) => {
    if (
      !window.confirm(
        `Are you sure you want to delete the folder "${folderName}"?`,
      )
    )
      return;

    try {
      await deleteCollection(folderId);
    } catch (error) {
      alert("Failed to delete folder: " + error.message);
    }
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (bytes === "-" || !bytes) return "-";
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "-";
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return "-";
    }
  };

  const displayItems = getDisplayItems();
  const isLoading = collectionsLoading || filesLoading;

  // If we need password and don't have it, prompt
  if (needsPassword && !passwordStorageService.hasPassword()) {
    return (
      <div style={{ padding: "20px", maxWidth: "600px", margin: "0 auto" }}>
        <h1>Password Required</h1>
        <div
          style={{
            backgroundColor: "#f8f9fa",
            padding: "20px",
            borderRadius: "4px",
            marginBottom: "20px",
          }}
        >
          <p>
            Your folders are encrypted. Please enter your password to decrypt
            them.
          </p>
          <p style={{ marginBottom: 0, fontSize: "14px", color: "#666" }}>
            Note: You may need to navigate to{" "}
            <a href="/collections" style={{ color: "#007bff" }}>
              Collections
            </a>{" "}
            to enter your password first.
          </p>
        </div>
        <button
          onClick={() => navigate("/collections")}
          style={{ padding: "10px 20px" }}
        >
          Go to Collections to Enter Password
        </button>
      </div>
    );
  }

  // If collections haven't loaded yet and we're not loading, show a message
  if (!hasCollections && !isLoading && !folderId) {
    return (
      <div style={{ padding: "20px", maxWidth: "1400px", margin: "0 auto" }}>
        {/* Header */}
        <div style={{ marginBottom: "20px" }}>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: "20px",
            }}
          >
            <h1 style={{ margin: 0 }}>File Manager</h1>
            <button onClick={() => navigate("/dashboard")}>
              ← Back to Dashboard
            </button>
          </div>
        </div>

        <div
          style={{
            textAlign: "center",
            padding: "60px",
            backgroundColor: "#f8f9fa",
            borderRadius: "8px",
          }}
        >
          <div style={{ fontSize: "48px", marginBottom: "20px" }}>📂</div>
          <h3>No collections found</h3>
          <p style={{ color: "#666", marginBottom: "20px" }}>
            Your collections may not have loaded yet, or you may need to create
            your first collection.
          </p>
          <div
            style={{ display: "flex", gap: "10px", justifyContent: "center" }}
          >
            <button
              onClick={() => loadAllCollections()}
              style={{
                padding: "12px 30px",
                backgroundColor: "#007bff",
                color: "white",
                border: "none",
                borderRadius: "4px",
                fontSize: "16px",
                cursor: "pointer",
              }}
            >
              Reload Collections
            </button>
            <button
              onClick={() => navigate("/collections/create")}
              style={{
                padding: "12px 30px",
                backgroundColor: "#28a745",
                color: "white",
                border: "none",
                borderRadius: "4px",
                fontSize: "16px",
                cursor: "pointer",
              }}
            >
              Create First Collection
            </button>
            <button
              onClick={() => navigate("/collections")}
              style={{
                padding: "12px 30px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
                fontSize: "16px",
                cursor: "pointer",
              }}
            >
              Use Legacy View
            </button>
          </div>

          {collectionsError && (
            <div style={{ marginTop: "20px", color: "#c00" }}>
              <strong>Error loading collections:</strong> {collectionsError}
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "1400px", margin: "0 auto" }}>
      {/* Header */}
      <div style={{ marginBottom: "20px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "20px",
          }}
        >
          <h1 style={{ margin: 0 }}>File Manager</h1>
          <button onClick={() => navigate("/dashboard")}>
            ← Back to Dashboard
          </button>
        </div>

        {/* Action buttons */}
        <div style={{ display: "flex", gap: "10px", marginBottom: "20px" }}>
          <button
            onClick={() =>
              navigate(
                folderId
                  ? `/collections/create?parent=${folderId}`
                  : "/collections/create",
              )
            }
            style={{
              padding: "8px 16px",
              backgroundColor: "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
            }}
          >
            📁 New Folder
          </button>
          <button
            onClick={() =>
              navigate(
                folderId
                  ? `/collections/${folderId}/add-file`
                  : "/files/upload",
              )
            }
            style={{
              padding: "8px 16px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
            }}
            disabled={!folderId}
            title={
              !folderId
                ? "Select a folder first"
                : "Upload file to current folder"
            }
          >
            📄 Upload File
          </button>

          {/* Manual refresh button for debugging */}
          {import.meta.env.DEV && (
            <button
              onClick={async () => {
                console.log("[FileManager] Manual refresh triggered");
                console.log("[FileManager] Current collections:", {
                  collections,
                  sharedCollections,
                });
                try {
                  const result = await loadAllCollections();
                  console.log("[FileManager] Manual refresh result:", result);
                  console.log("[FileManager] After refresh collections:", {
                    collections,
                    sharedCollections,
                  });
                  // Force update
                  forceUpdate({});
                } catch (err) {
                  console.error("[FileManager] Manual refresh failed:", err);
                }
              }}
              style={{
                padding: "8px 16px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
              }}
            >
              🔄 Refresh Collections (Debug)
            </button>
          )}

          <div style={{ marginLeft: "auto", display: "flex", gap: "10px" }}>
            {/* Filter */}
            <select
              value={filterType}
              onChange={(e) => setFilterType(e.target.value)}
              style={{ padding: "4px 8px" }}
            >
              <option value="all">All Items</option>
              <option value="folders">Folders Only</option>
              <option value="files">Files Only</option>
            </select>

            {/* Sort */}
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              style={{ padding: "4px 8px" }}
            >
              <option value="name">Sort by Name</option>
              <option value="date">Sort by Date</option>
              <option value="type">Sort by Type</option>
            </select>

            {/* View mode */}
            <button
              onClick={() => setViewMode(viewMode === "list" ? "grid" : "list")}
              style={{ padding: "4px 12px" }}
            >
              {viewMode === "list" ? "📋" : "⚏"}{" "}
              {viewMode === "list" ? "List" : "Grid"}
            </button>
          </div>
        </div>

        {/* Status message */}
        {!isLoading && hasCollections && (
          <div
            style={{ fontSize: "12px", color: "#666", marginBottom: "10px" }}
          >
            {collections.length} owned collections, {sharedCollections.length}{" "}
            shared collections loaded
          </div>
        )}

        {/* Breadcrumbs */}
        <div
          style={{
            padding: "10px",
            backgroundColor: "#f8f9fa",
            borderRadius: "4px",
            marginBottom: "20px",
          }}
        >
          {breadcrumbs.map((crumb, index) => (
            <span key={crumb.id || "root"}>
              {index > 0 && <span style={{ margin: "0 5px" }}>/</span>}
              {index === breadcrumbs.length - 1 ? (
                <strong>{crumb.name}</strong>
              ) : (
                <a
                  href="#"
                  onClick={(e) => {
                    e.preventDefault();
                    navigate(crumb.path);
                  }}
                  style={{ color: "#007bff", textDecoration: "none" }}
                >
                  {crumb.name}
                </a>
              )}
            </span>
          ))}
        </div>

        {/* Current folder info if in a subfolder */}
        {currentFolder &&
          currentFolder.members &&
          currentFolder.members.length > 1 && (
            <div
              style={{
                backgroundColor: "#d4edda",
                padding: "10px",
                borderRadius: "4px",
                marginBottom: "10px",
              }}
            >
              <small>
                Shared with {currentFolder.members.length - 1} user(s)
              </small>
            </div>
          )}
      </div>

      {/* Loading state */}
      {isLoading && (
        <div style={{ textAlign: "center", padding: "40px" }}>
          <div style={{ fontSize: "32px", marginBottom: "10px" }}>⏳</div>
          <p>Loading folders and files...</p>
          <p style={{ fontSize: "12px", color: "#666", marginTop: "10px" }}>
            {collectionsLoading && "Loading collections... "}
            {filesLoading && "Loading files... "}
            {!collectionsLoading && !filesLoading && "Decrypting data..."}
          </p>
          {!loadAttempted && (
            <p style={{ fontSize: "11px", color: "#999", marginTop: "5px" }}>
              Initializing collection service...
            </p>
          )}
        </div>
      )}

      {/* Error state */}
      {(collectionsError || filesError) && (
        <div
          style={{
            backgroundColor: "#fee",
            color: "#c00",
            padding: "15px",
            marginBottom: "20px",
            borderRadius: "4px",
          }}
        >
          <strong>Error:</strong> {collectionsError || filesError}
          {collectionsError && (
            <div style={{ marginTop: "10px" }}>
              <button
                onClick={() => loadAllCollections()}
                style={{
                  padding: "5px 10px",
                  backgroundColor: "#c00",
                  color: "white",
                  border: "none",
                  borderRadius: "3px",
                }}
              >
                Retry Loading Collections
              </button>
            </div>
          )}
        </div>
      )}

      {/* Empty state */}
      {!isLoading && displayItems.length === 0 && (
        <div
          style={{
            textAlign: "center",
            padding: "60px",
            backgroundColor: "#f8f9fa",
            borderRadius: "8px",
          }}
        >
          <div style={{ fontSize: "48px", marginBottom: "20px" }}>
            {!folderId ? "📂" : "📁"}
          </div>
          <h3>
            {!folderId
              ? hasCollections
                ? "No items in root folder"
                : "No collections found"
              : "This folder is empty"}
          </h3>
          <p style={{ color: "#666", marginBottom: "20px" }}>
            {!folderId
              ? hasCollections
                ? "Your collections may be in subfolders, or you can create a new folder"
                : "Create your first collection to get started"
              : "Create a new folder or upload files to get started"}
          </p>
          <div
            style={{ display: "flex", gap: "10px", justifyContent: "center" }}
          >
            {folderId ? (
              <button
                onClick={() => navigate(`/collections/${folderId}/add-file`)}
                style={{
                  padding: "12px 30px",
                  backgroundColor: "#28a745",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  fontSize: "16px",
                  cursor: "pointer",
                }}
              >
                Upload Your First File
              </button>
            ) : (
              <>
                <button
                  onClick={() => navigate("/collections/create")}
                  style={{
                    padding: "12px 30px",
                    backgroundColor: "#28a745",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    fontSize: "16px",
                    cursor: "pointer",
                  }}
                >
                  Create New Folder
                </button>
                {!hasCollections && (
                  <button
                    onClick={() => loadAllCollections()}
                    style={{
                      padding: "12px 30px",
                      backgroundColor: "#007bff",
                      color: "white",
                      border: "none",
                      borderRadius: "4px",
                      fontSize: "16px",
                      cursor: "pointer",
                    }}
                  >
                    Reload Collections
                  </button>
                )}
              </>
            )}
          </div>
        </div>
      )}

      {/* Items list/grid */}
      {!isLoading && displayItems.length > 0 && (
        <div>
          {viewMode === "list" ? (
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr
                  style={{
                    backgroundColor: "#f8f9fa",
                    borderBottom: "2px solid #dee2e6",
                  }}
                >
                  <th
                    style={{
                      padding: "12px",
                      textAlign: "left",
                      width: "40px",
                    }}
                  >
                    <input
                      type="checkbox"
                      checked={
                        selectedItems.size === displayItems.length &&
                        displayItems.length > 0
                      }
                      onChange={(e) => {
                        if (e.target.checked) {
                          setSelectedItems(
                            new Set(displayItems.map((item) => item.id)),
                          );
                        } else {
                          setSelectedItems(new Set());
                        }
                      }}
                    />
                  </th>
                  <th style={{ padding: "12px", textAlign: "left" }}>Name</th>
                  <th
                    style={{
                      padding: "12px",
                      textAlign: "left",
                      width: "100px",
                    }}
                  >
                    Size
                  </th>
                  <th
                    style={{
                      padding: "12px",
                      textAlign: "left",
                      width: "150px",
                    }}
                  >
                    Modified
                  </th>
                  <th
                    style={{
                      padding: "12px",
                      textAlign: "left",
                      width: "150px",
                    }}
                  >
                    Type
                  </th>
                  <th
                    style={{
                      padding: "12px",
                      textAlign: "center",
                      width: "100px",
                    }}
                  >
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {displayItems.map((item) => (
                  <tr
                    key={item.id}
                    style={{
                      borderBottom: "1px solid #dee2e6",
                      cursor: "pointer",
                    }}
                    onClick={() => handleItemClick(item)}
                  >
                    <td
                      style={{ padding: "12px" }}
                      onClick={(e) => e.stopPropagation()}
                    >
                      <input
                        type="checkbox"
                        checked={selectedItems.has(item.id)}
                        onChange={() => handleItemSelect(item.id)}
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
                        <span style={{ fontSize: "20px" }}>
                          {item.itemType === "folder" ? "📁" : "📄"}
                        </span>
                        <span>{item.displayName}</span>
                      </div>
                    </td>
                    <td style={{ padding: "12px" }}>
                      {formatFileSize(item.displaySize)}
                    </td>
                    <td style={{ padding: "12px" }}>
                      {formatDate(item.displayDate)}
                    </td>
                    <td style={{ padding: "12px" }}>
                      {item.itemType === "folder"
                        ? "Folder"
                        : item.mime_type || "File"}
                    </td>
                    <td
                      style={{ padding: "12px", textAlign: "center" }}
                      onClick={(e) => e.stopPropagation()}
                    >
                      {item.itemType === "folder" ? (
                        <button
                          onClick={() =>
                            handleDeleteFolder(item.id, item.displayName)
                          }
                          style={{
                            padding: "4px 8px",
                            fontSize: "12px",
                            backgroundColor: "#dc3545",
                            color: "white",
                            border: "none",
                            borderRadius: "3px",
                          }}
                        >
                          Delete
                        </button>
                      ) : (
                        <div
                          style={{
                            display: "flex",
                            gap: "4px",
                            justifyContent: "center",
                          }}
                        >
                          <button
                            onClick={() => handleDownloadFile(item.id)}
                            style={{
                              padding: "4px 8px",
                              fontSize: "12px",
                              backgroundColor: "#007bff",
                              color: "white",
                              border: "none",
                              borderRadius: "3px",
                            }}
                          >
                            Download
                          </button>
                          <button
                            onClick={() => handleDeleteFile(item.id)}
                            style={{
                              padding: "4px 8px",
                              fontSize: "12px",
                              backgroundColor: "#dc3545",
                              color: "white",
                              border: "none",
                              borderRadius: "3px",
                            }}
                          >
                            Delete
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            // Grid view
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fill, minmax(150px, 1fr))",
                gap: "20px",
              }}
            >
              {displayItems.map((item) => (
                <div
                  key={item.id}
                  style={{
                    border: "1px solid #ddd",
                    borderRadius: "8px",
                    padding: "20px",
                    textAlign: "center",
                    cursor: "pointer",
                    backgroundColor: selectedItems.has(item.id)
                      ? "#e9ecef"
                      : "white",
                  }}
                  onClick={() => {
                    if (item.itemType === "folder") {
                      navigate(`/files/${item.id}`);
                    } else {
                      handleItemSelect(item.id);
                    }
                  }}
                >
                  <div style={{ fontSize: "48px", marginBottom: "10px" }}>
                    {item.itemType === "folder" ? "📁" : "📄"}
                  </div>
                  <div style={{ fontSize: "14px", wordBreak: "break-word" }}>
                    {item.displayName}
                  </div>
                  <div
                    style={{
                      fontSize: "12px",
                      color: "#666",
                      marginTop: "5px",
                    }}
                  >
                    {item.itemType === "folder"
                      ? "Folder"
                      : formatFileSize(item.displaySize)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Selected items actions */}
      {selectedItems.size > 0 && (
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
          }}
        >
          <span>{selectedItems.size} item(s) selected</span>
          <button
            onClick={() => setSelectedItems(new Set())}
            style={{
              marginLeft: "20px",
              padding: "5px 15px",
              backgroundColor: "#666",
              color: "white",
              border: "none",
              borderRadius: "4px",
            }}
          >
            Clear
          </button>
        </div>
      )}

      {/* Info box */}
      <div
        style={{
          marginTop: "40px",
          padding: "20px",
          backgroundColor: "#f8f9fa",
          borderRadius: "4px",
          fontSize: "14px",
          color: "#666",
        }}
      >
        <p style={{ margin: "0 0 10px 0" }}>
          <strong>💡 Tips:</strong>
        </p>
        <ul style={{ marginBottom: 0, paddingLeft: "20px" }}>
          <li>Click on folders to navigate into them</li>
          <li>Use the breadcrumbs to navigate back to parent folders</li>
          <li>Files must be uploaded to a specific folder</li>
          <li>All files and folder names are end-to-end encrypted</li>
          <li>
            Access the{" "}
            <a href="/collections" style={{ color: "#007bff" }}>
              legacy collections view
            </a>{" "}
            for advanced features
          </li>
        </ul>
      </div>

      {/* Debug info in development */}
      {import.meta.env.DEV && (
        <details
          style={{
            marginTop: "20px",
            padding: "10px",
            backgroundColor: "#f8f9fa",
          }}
        >
          <summary>🔍 Debug Information</summary>
          <div>
            <h4>Collections State:</h4>
            <pre>
              {JSON.stringify(
                {
                  collectionsCount: collections.length,
                  sharedCollectionsCount: sharedCollections.length,
                  currentFolderId: folderId || "root",
                  displayItemsCount: displayItems.length,
                  isLoading: collectionsLoading || filesLoading,
                  collectionsLoading,
                  filesLoading,
                  subfoldersInCurrentFolder: getSubfolders().length,
                  filesInCurrentFolder: files.length,
                  hasPassword: passwordStorageService.hasPassword(),
                  needsPassword,
                },
                null,
                2,
              )}
            </pre>
            <h4>Raw collections data:</h4>
            <pre>
              {JSON.stringify(
                {
                  owned: collections.length,
                  shared: sharedCollections.length,
                  firstOwned: collections[0]
                    ? {
                        id: collections[0].id,
                        parent_id: collections[0].parent_id,
                        name: collections[0].name || "[Not decrypted]",
                        decrypt_error: collections[0].decrypt_error,
                      }
                    : null,
                  firstShared: sharedCollections[0]
                    ? {
                        id: sharedCollections[0].id,
                        parent_id: sharedCollections[0].parent_id,
                        name: sharedCollections[0].name || "[Not decrypted]",
                      }
                    : null,
                },
                null,
                2,
              )}
            </pre>
            <h4>First few collections:</h4>
            <pre>
              {JSON.stringify(
                {
                  collectionsSample: collections.slice(0, 2).map((c) => ({
                    id: c.id,
                    name: c.name || "[Not decrypted]",
                    parent_id: c.parent_id,
                    parent_id_type: typeof c.parent_id,
                    isNull: c.parent_id === null,
                    isZeroUUID:
                      c.parent_id === "00000000-0000-0000-0000-000000000000",
                    isRootLevel: isRootLevel(c.parent_id),
                  })),
                },
                null,
                2,
              )}
            </pre>
            <h4>GetSubfolders debug:</h4>
            <pre>
              {JSON.stringify(
                {
                  currentFolderId: folderId || "root",
                  getSubfoldersDebug: (() => {
                    const allColls = [
                      ...new Map(
                        [...collections, ...sharedCollections].map((c) => [
                          c.id,
                          c,
                        ]),
                      ).values(),
                    ];
                    const rootColls = allColls.filter((c) =>
                      isRootLevel(c.parent_id),
                    );
                    return {
                      totalUnique: allColls.length,
                      rootLevel: rootColls.length,
                      rootCollections: rootColls.map((c) => ({
                        id: c.id,
                        name: c.name || "[Not decrypted]",
                        parent_id: c.parent_id,
                      })),
                    };
                  })(),
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

const ProtectedFileManager = withPasswordProtection(FileManager);
export default ProtectedFileManager;
