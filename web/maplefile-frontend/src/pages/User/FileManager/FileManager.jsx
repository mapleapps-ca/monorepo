// File: src/pages/User/FileManager/FileManager.jsx
// Enhanced consolidated file manager with inline upload and better file/folder management
import React, { useState, useEffect, useCallback, useRef } from "react";
import { useNavigate, useParams, useLocation } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";
import useCollections from "../../../hooks/useCollections.js";
import useFiles from "../../../hooks/useFiles.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const FileManager = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { folderId } = useParams(); // Current folder (collection) ID from URL
  const fileInputRef = useRef(null);

  const {
    collectionService,
    fileService,
    passwordStorageService,
    cryptoService,
    collectionCryptoService,
  } = useServices();
  const { isAuthenticated } = useAuth();

  // Hooks
  const {
    collections,
    sharedCollections,
    isLoading: collectionsLoading,
    error: collectionsError,
    loadAllCollections,
    deleteCollection,
    createCollection,
  } = useCollections();

  const {
    files,
    isLoading: filesLoading,
    error: filesError,
    loadFilesByCollection,
    downloadAndSaveFile,
    deleteFile,
    archiveFile,
    restoreFile,
    reloadFiles,
    getActiveFiles,
    getArchivedFiles,
    getDeletedFiles,
    getFileStats,
    canDownloadFile,
    canRestoreFile,
    FILE_STATES,
  } = useFiles(folderId);

  // Local state
  const [currentFolder, setCurrentFolder] = useState(null);
  const [breadcrumbs, setBreadcrumbs] = useState([]);
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [viewMode, setViewMode] = useState("list"); // list or grid
  const [sortBy, setSortBy] = useState("name"); // name, date, type
  const [filterType, setFilterType] = useState("all"); // all, folders, files
  const [fileStateFilter, setFileStateFilter] = useState("active"); // active, archived, deleted, all
  const [showUploadArea, setShowUploadArea] = useState(false);
  const [uploadingFiles, setUploadingFiles] = useState(new Map());
  const [showCreateFolder, setShowCreateFolder] = useState(false);
  const [newFolderName, setNewFolderName] = useState("");
  const [isCreatingFolder, setIsCreatingFolder] = useState(false);
  const [needsPassword, setNeedsPassword] = useState(false);
  const [loadAttempted, setLoadAttempted] = useState(false);

  const hasCollections = collections.length > 0 || sharedCollections.length > 0;

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
  }, [loadAttempted, loadAllCollections]);

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

  // Show success messages from navigation state
  useEffect(() => {
    if (location.state?.message) {
      // You could show a toast notification here
      console.log("[FileManager] Success:", location.state.message);
      // Clear the message from state
      navigate(location.pathname, { replace: true });
    }
  }, [location.state, navigate]);

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
      // Load files when entering a folder with state filtering
      const statesToInclude = getStatesToInclude(fileStateFilter);
      loadFilesByCollection(folderId, true, statesToInclude);
    } else {
      setCurrentFolder(null);
      setBreadcrumbs([{ id: null, name: "My Files", path: "/files" }]);
    }
  }, [folderId, loadCurrentFolder, loadFilesByCollection, fileStateFilter]);

  // Get states to include based on filter
  const getStatesToInclude = (filter) => {
    switch (filter) {
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
    const collectionMap = new Map();
    collections.forEach((c) => collectionMap.set(c.id, c));
    sharedCollections.forEach((c) => collectionMap.set(c.id, c));
    const allColls = Array.from(collectionMap.values());

    let result;
    if (!folderId) {
      result = allColls.filter((c) => isRootLevel(c.parent_id));
    } else {
      result = allColls.filter((c) => c.parent_id === folderId);
    }

    return result;
  }, [collections, sharedCollections, folderId, isRootLevel]);

  // Get files to display based on state filter
  const getFilesToDisplay = useCallback(() => {
    switch (fileStateFilter) {
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
  }, [
    files,
    fileStateFilter,
    getActiveFiles,
    getArchivedFiles,
    getDeletedFiles,
    FILE_STATES,
  ]);

  // Get items to display (folders + files)
  const getDisplayItems = useCallback(() => {
    const subfolders = getSubfolders();
    const filesToShow = getFilesToDisplay();

    const folders = subfolders.map((folder) => ({
      ...folder,
      itemType: "folder",
      displayName: folder.name || "[Encrypted]",
      displaySize: "-",
      displayDate: folder.modified_at || folder.created_at,
    }));

    const displayFiles = filesToShow.map((file) => ({
      ...file,
      itemType: "file",
      displayName: file.name || "[Encrypted]",
      displaySize: file.size || file.encrypted_file_size_in_bytes || 0,
      displayDate: file.modified_at || file.created_at,
    }));

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
  }, [getSubfolders, getFilesToDisplay, filterType, sortBy]);

  // Handle item click (navigation for folders, selection for files)
  const handleItemClick = (item) => {
    if (item.itemType === "folder") {
      navigate(`/files/${item.id}`);
    } else {
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

  // Create new folder
  const handleCreateFolder = async () => {
    if (!newFolderName.trim()) return;

    setIsCreatingFolder(true);
    try {
      const folderData = {
        name: newFolderName.trim(),
        collection_type: "folder",
        parent_id: folderId || null,
      };

      const password = passwordStorageService.getPassword();
      if (!password) {
        throw new Error("Password required to create folder");
      }

      await collectionService.createCollectionWithPassword(
        folderData,
        password,
      );

      // Reload collections to show the new folder
      await loadAllCollections();

      // Clear form
      setNewFolderName("");
      setShowCreateFolder(false);

      console.log("[FileManager] Folder created successfully");
    } catch (error) {
      console.error("[FileManager] Failed to create folder:", error);
      alert("Failed to create folder: " + error.message);
    } finally {
      setIsCreatingFolder(false);
    }
  };

  // Handle file upload
  const handleFileUpload = async (file) => {
    if (!folderId) {
      alert("Please select a folder first");
      return;
    }

    const uploadId = Date.now().toString();
    setUploadingFiles((prev) =>
      new Map(prev).set(uploadId, {
        id: uploadId,
        name: file.name,
        progress: 0,
        status: "starting",
      }),
    );

    try {
      // Update progress
      setUploadingFiles((prev) =>
        new Map(prev).set(uploadId, {
          ...prev.get(uploadId),
          status: "encrypting",
          progress: 10,
        }),
      );

      // Get password and collection key
      const password = passwordStorageService.getPassword();
      if (!password) {
        throw new Error("Password required to upload files");
      }

      // Ensure collection is loaded
      const collection = await collectionService.getCollection(
        folderId,
        password,
      );
      let collectionKey =
        collection.collection_key ||
        collectionCryptoService.getCachedCollectionKey(folderId);

      if (!collectionKey) {
        throw new Error("Collection key not available");
      }

      // Generate file encryption key
      const fileKey = cryptoService.generateRandomKey();

      // Read file content
      const fileContent = await file.arrayBuffer();

      setUploadingFiles((prev) =>
        new Map(prev).set(uploadId, {
          ...prev.get(uploadId),
          progress: 30,
        }),
      );

      // Encrypt file content
      const encryptedContent = await cryptoService.encryptWithKey(
        new Uint8Array(fileContent),
        fileKey,
      );

      // Generate file hash
      const fileHash = await cryptoService.hashData(
        new Uint8Array(fileContent),
      );
      const encryptedHash = cryptoService.uint8ArrayToBase64(fileHash);

      setUploadingFiles((prev) =>
        new Map(prev).set(uploadId, {
          ...prev.get(uploadId),
          progress: 50,
        }),
      );

      // Prepare file metadata
      const metadata = {
        name: file.name,
        mime_type: file.type || "application/octet-stream",
        size: file.size,
        created_at: new Date().toISOString(),
        uploaded_at: new Date().toISOString(),
      };

      // Encrypt metadata
      const encryptedMetadata = await cryptoService.encryptWithKey(
        JSON.stringify(metadata),
        fileKey,
      );

      // Encrypt file key with collection key
      const encryptedFileKeyData = await cryptoService.encryptFileKey(
        fileKey,
        collectionKey,
      );

      setUploadingFiles((prev) =>
        new Map(prev).set(uploadId, {
          ...prev.get(uploadId),
          progress: 70,
        }),
      );

      // Prepare file data for API
      const fileData = {
        id: cryptoService.generateUUID(),
        collection_id: folderId,
        encrypted_metadata: encryptedMetadata,
        encrypted_file_key: {
          ciphertext: cryptoService.uint8ArrayToBase64(
            encryptedFileKeyData.ciphertext,
          ),
          nonce: cryptoService.uint8ArrayToBase64(encryptedFileKeyData.nonce),
          key_version: 1,
        },
        encryption_version: "v1.0",
        encrypted_hash: encryptedHash,
        expected_file_size_in_bytes:
          cryptoService.tryDecodeBase64(encryptedContent).length,
        version: 1,
        state: FILE_STATES.PENDING,
        tombstone_version: 0,
        tombstone_expiry: "0001-01-01T00:00:00Z",
      };

      setUploadingFiles((prev) =>
        new Map(prev).set(uploadId, {
          ...prev.get(uploadId),
          status: "uploading",
          progress: 80,
        }),
      );

      // Convert encrypted content to blob
      const encryptedBytes = cryptoService.tryDecodeBase64(encryptedContent);
      const encryptedBlob = new Blob([encryptedBytes], {
        type: "application/octet-stream",
      });

      // Upload file
      await fileService.uploadFile(fileData, encryptedBlob);

      setUploadingFiles((prev) =>
        new Map(prev).set(uploadId, {
          ...prev.get(uploadId),
          status: "completed",
          progress: 100,
        }),
      );

      // Reload files to show the new file
      const statesToInclude = getStatesToInclude(fileStateFilter);
      await loadFilesByCollection(folderId, true, statesToInclude);

      // Remove from upload queue after a delay
      setTimeout(() => {
        setUploadingFiles((prev) => {
          const next = new Map(prev);
          next.delete(uploadId);
          return next;
        });
      }, 2000);
    } catch (error) {
      console.error("[FileManager] File upload failed:", error);
      setUploadingFiles((prev) =>
        new Map(prev).set(uploadId, {
          ...prev.get(uploadId),
          status: "error",
          error: error.message,
        }),
      );

      // Remove from upload queue after delay even on error
      setTimeout(() => {
        setUploadingFiles((prev) => {
          const next = new Map(prev);
          next.delete(uploadId);
          return next;
        });
      }, 5000);
    }
  };

  // Handle file input change
  const handleFileInputChange = (e) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      Array.from(files).forEach((file) => {
        handleFileUpload(file);
      });
    }
    // Reset input
    e.target.value = "";
  };

  // Handle drag and drop
  const handleDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      Array.from(files).forEach((file) => {
        handleFileUpload(file);
      });
    }
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    e.stopPropagation();
  };

  // File operations
  const handleDownloadFile = async (fileId) => {
    try {
      await downloadAndSaveFile(fileId);
    } catch (error) {
      alert("Failed to download file: " + error.message);
    }
  };

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

  const handleArchiveFile = async (fileId) => {
    if (!window.confirm("Are you sure you want to archive this file?")) return;
    try {
      await archiveFile(fileId);
      setSelectedItems((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    } catch (error) {
      alert("Failed to archive file: " + error.message);
    }
  };

  const handleRestoreFile = async (fileId) => {
    if (!window.confirm("Are you sure you want to restore this file?")) return;
    try {
      await restoreFile(fileId);
      setSelectedItems((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    } catch (error) {
      alert("Failed to restore file: " + error.message);
    }
  };

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
  const fileStats = getFileStats();

  // If we need password and don't have it, show prompt
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

  return (
    <div
      style={{ padding: "20px", maxWidth: "1400px", margin: "0 auto" }}
      onDrop={handleDrop}
      onDragOver={handleDragOver}
    >
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
          <h1 style={{ margin: 0 }}>📁 File Manager</h1>
          <button onClick={() => navigate("/dashboard")}>
            ← Back to Dashboard
          </button>
        </div>

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

        {/* Action buttons */}
        <div
          style={{
            display: "flex",
            gap: "10px",
            marginBottom: "20px",
            flexWrap: "wrap",
          }}
        >
          <button
            onClick={() => setShowCreateFolder(true)}
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
            onClick={() => fileInputRef.current?.click()}
            disabled={!folderId}
            style={{
              padding: "8px 16px",
              backgroundColor: folderId ? "#28a745" : "#ccc",
              color: "white",
              border: "none",
              borderRadius: "4px",
            }}
            title={!folderId ? "Select a folder first" : "Upload files"}
          >
            📄 Upload Files
          </button>

          <button
            onClick={() => setShowUploadArea(!showUploadArea)}
            disabled={!folderId}
            style={{
              padding: "8px 16px",
              backgroundColor: folderId ? "#17a2b8" : "#ccc",
              color: "white",
              border: "none",
              borderRadius: "4px",
            }}
          >
            {showUploadArea ? "Hide" : "Show"} Upload Area
          </button>

          {/* File state filter */}
          <select
            value={fileStateFilter}
            onChange={(e) => setFileStateFilter(e.target.value)}
            style={{ padding: "4px 8px" }}
          >
            <option value="active">Active Files ({fileStats.active})</option>
            <option value="archived">
              Archived Files ({fileStats.archived})
            </option>
            <option value="deleted">Deleted Files ({fileStats.deleted})</option>
            <option value="pending">Pending Files ({fileStats.pending})</option>
            <option value="all">All Files ({fileStats.total})</option>
          </select>

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

        {/* Upload progress */}
        {uploadingFiles.size > 0 && (
          <div
            style={{
              backgroundColor: "#f8f9fa",
              border: "1px solid #dee2e6",
              borderRadius: "4px",
              padding: "15px",
              marginBottom: "20px",
            }}
          >
            <h4 style={{ margin: "0 0 10px 0" }}>
              Uploading Files ({uploadingFiles.size})
            </h4>
            {Array.from(uploadingFiles.values()).map((upload) => (
              <div key={upload.id} style={{ marginBottom: "10px" }}>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                  }}
                >
                  <span>{upload.name}</span>
                  <span
                    style={{
                      color: upload.status === "error" ? "#dc3545" : "#666",
                    }}
                  >
                    {upload.status === "error"
                      ? "Failed"
                      : `${upload.progress}%`}
                  </span>
                </div>
                <div
                  style={{
                    backgroundColor: "#e0e0e0",
                    borderRadius: "4px",
                    height: "8px",
                    marginTop: "4px",
                  }}
                >
                  <div
                    style={{
                      backgroundColor:
                        upload.status === "error"
                          ? "#dc3545"
                          : upload.status === "completed"
                            ? "#28a745"
                            : "#007bff",
                      height: "100%",
                      borderRadius: "4px",
                      width: `${upload.progress}%`,
                      transition: "width 0.3s ease",
                    }}
                  />
                </div>
                {upload.status === "error" && (
                  <div
                    style={{
                      color: "#dc3545",
                      fontSize: "12px",
                      marginTop: "2px",
                    }}
                  >
                    {upload.error}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Upload area */}
        {showUploadArea && folderId && (
          <div
            style={{
              border: "2px dashed #ccc",
              borderRadius: "8px",
              padding: "40px 20px",
              textAlign: "center",
              backgroundColor: "#fafafa",
              marginBottom: "20px",
            }}
          >
            <div style={{ fontSize: "48px", marginBottom: "20px" }}>📁</div>
            <h3>Drag and drop files here</h3>
            <p style={{ color: "#666", marginBottom: "20px" }}>
              or click the "Upload Files" button to select files
            </p>
            <p style={{ color: "#999", fontSize: "14px" }}>
              Files will be encrypted before upload
            </p>
          </div>
        )}

        {/* Create folder form */}
        {showCreateFolder && (
          <div
            style={{
              backgroundColor: "#f8f9fa",
              border: "1px solid #dee2e6",
              borderRadius: "4px",
              padding: "15px",
              marginBottom: "20px",
            }}
          >
            <h4 style={{ margin: "0 0 10px 0" }}>Create New Folder</h4>
            <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
              <input
                type="text"
                value={newFolderName}
                onChange={(e) => setNewFolderName(e.target.value)}
                placeholder="Folder name"
                onKeyPress={(e) => e.key === "Enter" && handleCreateFolder()}
                disabled={isCreatingFolder}
                style={{ padding: "8px", flex: 1 }}
              />
              <button
                onClick={handleCreateFolder}
                disabled={!newFolderName.trim() || isCreatingFolder}
                style={{
                  padding: "8px 16px",
                  backgroundColor: "#28a745",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                }}
              >
                {isCreatingFolder ? "Creating..." : "Create"}
              </button>
              <button
                onClick={() => {
                  setShowCreateFolder(false);
                  setNewFolderName("");
                }}
                disabled={isCreatingFolder}
                style={{
                  padding: "8px 16px",
                  backgroundColor: "#6c757d",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                }}
              >
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Hidden file input */}
      <input
        ref={fileInputRef}
        type="file"
        onChange={handleFileInputChange}
        style={{ display: "none" }}
        multiple
      />

      {/* Loading state */}
      {isLoading && (
        <div style={{ textAlign: "center", padding: "40px" }}>
          <div style={{ fontSize: "32px", marginBottom: "10px" }}>⏳</div>
          <p>Loading folders and files...</p>
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
            <button
              onClick={() => setShowCreateFolder(true)}
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
              Create New Folder
            </button>
            {folderId && (
              <button
                onClick={() => fileInputRef.current?.click()}
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
                Upload Files
              </button>
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
                      width: "100px",
                    }}
                  >
                    Type
                  </th>
                  {fileStateFilter !== "active" && (
                    <th
                      style={{
                        padding: "12px",
                        textAlign: "left",
                        width: "80px",
                      }}
                    >
                      State
                    </th>
                  )}
                  <th
                    style={{
                      padding: "12px",
                      textAlign: "center",
                      width: "150px",
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
                        {item._decrypt_error && (
                          <span style={{ color: "#ff6b6b", fontSize: "12px" }}>
                            (Decrypt failed)
                          </span>
                        )}
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
                    {fileStateFilter !== "active" && (
                      <td style={{ padding: "12px" }}>
                        {item.itemType === "file" && (
                          <span
                            style={{
                              padding: "2px 6px",
                              borderRadius: "3px",
                              fontSize: "11px",
                              backgroundColor:
                                item.state === "active"
                                  ? "#d4edda"
                                  : item.state === "archived"
                                    ? "#e2e3e5"
                                    : item.state === "deleted"
                                      ? "#f8d7da"
                                      : "#fff3cd",
                              color:
                                item.state === "active"
                                  ? "#155724"
                                  : item.state === "archived"
                                    ? "#383d41"
                                    : item.state === "deleted"
                                      ? "#721c24"
                                      : "#856404",
                            }}
                          >
                            {item.state?.toUpperCase()}
                          </span>
                        )}
                      </td>
                    )}
                    <td
                      style={{ padding: "12px", textAlign: "center" }}
                      onClick={(e) => e.stopPropagation()}
                    >
                      <div
                        style={{
                          display: "flex",
                          gap: "4px",
                          justifyContent: "center",
                        }}
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
                          <>
                            {canDownloadFile(item) && (
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
                            )}
                            {item.state === "active" && (
                              <button
                                onClick={() => handleArchiveFile(item.id)}
                                style={{
                                  padding: "4px 8px",
                                  fontSize: "12px",
                                  backgroundColor: "#6c757d",
                                  color: "white",
                                  border: "none",
                                  borderRadius: "3px",
                                }}
                              >
                                Archive
                              </button>
                            )}
                            {canRestoreFile(item) && (
                              <button
                                onClick={() => handleRestoreFile(item.id)}
                                style={{
                                  padding: "4px 8px",
                                  fontSize: "12px",
                                  backgroundColor: "#28a745",
                                  color: "white",
                                  border: "none",
                                  borderRadius: "3px",
                                }}
                              >
                                Restore
                              </button>
                            )}
                            {!item._is_deleted && (
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
                            )}
                          </>
                        )}
                      </div>
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
                  {fileStateFilter !== "active" && item.itemType === "file" && (
                    <div
                      style={{
                        fontSize: "10px",
                        color: "#999",
                        marginTop: "2px",
                      }}
                    >
                      {item.state?.toUpperCase()}
                    </div>
                  )}
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
          <li>Drag and drop files to upload them to the current folder</li>
          <li>Use the breadcrumbs to navigate back to parent folders</li>
          <li>All files and folder names are end-to-end encrypted</li>
          <li>
            Files can be in different states: Active, Archived, Deleted, or
            Pending
          </li>
          <li>Use the state filter to view files in different states</li>
        </ul>
      </div>
    </div>
  );
};

const ProtectedFileManager = withPasswordProtection(FileManager);
export default ProtectedFileManager;
