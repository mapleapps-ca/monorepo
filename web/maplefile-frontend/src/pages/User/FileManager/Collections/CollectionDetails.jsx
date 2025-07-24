// File: src/pages/User/FileManager/Collections/CollectionDetails.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useParams } from "react-router";
import { useFiles, useCrypto, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  FolderIcon,
  PhotoIcon,
  DocumentIcon,
  ArrowLeftIcon,
  CloudArrowUpIcon,
  ShareIcon,
  ArrowDownTrayIcon,
  MagnifyingGlassIcon,
  ArrowPathIcon,
  PlusIcon,
  CheckIcon,
} from "@heroicons/react/24/outline";

const CollectionDetails = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();

  const {
    getCollectionManager,
    listCollectionManager,
    downloadFileManager,
    authService,
  } = useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [collection, setCollection] = useState(null);
  const [files, setFiles] = useState([]);
  const [subCollections, setSubCollections] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());
  const [fileManager, setFileManager] = useState(null);
  const [viewMode, setViewMode] = useState("active"); // active, archived, deleted, all

  // Initialize file manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authService?.isAuthenticated()) return;

      try {
        const { default: ListFileManager } = await import(
          "../../../../services/Manager/File/ListFileManager.js"
        );

        const manager = new ListFileManager(
          authService,
          getCollectionManager,
          listCollectionManager,
        );
        await manager.initialize();
        setFileManager(manager);

        console.log("[CollectionDetails] ListFileManager initialized");
      } catch (err) {
        console.error(
          "[CollectionDetails] Failed to initialize ListFileManager:",
          err,
        );
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();
  }, [authService, getCollectionManager, listCollectionManager]);

  // Load collection and its contents
  const loadCollection = useCallback(
    async (forceRefresh = false) => {
      if (!getCollectionManager || !collectionId || !fileManager) return;

      setIsLoading(true);
      setError("");

      try {
        console.log("[CollectionDetails] Loading collection:", collectionId);

        // Load collection details
        const result = await getCollectionManager.getCollection(
          collectionId,
          forceRefresh,
        );

        if (result.collection) {
          const processedCollection = await processCollection(
            result.collection,
          );
          setCollection(processedCollection);

          // Cache collection key if available
          if (result.collection.collection_key) {
            CollectionCryptoService.cacheCollectionKey(
              collectionId,
              result.collection.collection_key,
            );
          }
        }

        // Load files in this collection
        await loadCollectionFiles(collectionId, forceRefresh);

        // Load sub-collections
        await loadSubCollections(collectionId, forceRefresh);
      } catch (err) {
        console.error("[CollectionDetails] Failed to load collection:", err);
        setError("Could not load folder. Please try again.");
      } finally {
        setIsLoading(false);
      }
    },
    [getCollectionManager, collectionId, fileManager, CollectionCryptoService],
  );

  // Process collection for decryption
  const processCollection = async (rawCollection) => {
    let processedCollection = { ...rawCollection };

    if (rawCollection.encrypted_name || rawCollection.encrypted_description) {
      try {
        let collectionKey =
          rawCollection.collection_key ||
          CollectionCryptoService.getCachedCollectionKey(rawCollection.id);

        if (collectionKey) {
          const { default: CollectionCryptoServiceClass } = await import(
            "../../../../services/Crypto/CollectionCryptoService.js"
          );

          const decryptedCollection =
            await CollectionCryptoServiceClass.decryptCollectionFromAPI(
              rawCollection,
              collectionKey,
            );
          processedCollection = decryptedCollection;
        } else {
          processedCollection.name = "Locked Folder";
          processedCollection._isDecrypted = false;
        }
      } catch (decryptError) {
        console.error("[CollectionDetails] Decryption error:", decryptError);
        processedCollection.name = "Locked Folder";
        processedCollection._isDecrypted = false;
      }
    } else {
      processedCollection._isDecrypted = true;
    }

    processedCollection.type = processedCollection.collection_type || "folder";
    return processedCollection;
  };

  // Load files using the correct method from ListFileManager
  const loadCollectionFiles = useCallback(
    async (collectionId, forceRefresh = false) => {
      if (!fileManager) return;

      try {
        console.log(
          "[CollectionDetails] Loading files for collection:",
          collectionId,
        );

        // Determine states to include based on view mode
        const statesToInclude = getStatesToInclude(viewMode);

        // Use the correct method: listFilesByCollection
        const loadedFiles = await fileManager.listFilesByCollection(
          collectionId,
          statesToInclude,
          forceRefresh,
        );

        console.log("[CollectionDetails] Files loaded:", loadedFiles.length);
        setFiles(loadedFiles);

        if (loadedFiles.length === 0 && viewMode === "active") {
          console.log(
            "[CollectionDetails] No active files found in collection",
          );
        }
      } catch (err) {
        console.error("[CollectionDetails] Failed to load files:", err);
        setFiles([]);
      }
    },
    [fileManager, viewMode],
  );

  // Load sub-collections
  const loadSubCollections = useCallback(
    async (parentCollectionId, forceRefresh = false) => {
      if (!listCollectionManager) return;

      try {
        console.log(
          "[CollectionDetails] Loading sub-collections for:",
          parentCollectionId,
        );

        // Get all collections
        const result =
          await listCollectionManager.listCollections(forceRefresh);

        if (result.collections) {
          // Filter for collections that have this collection as parent
          const subCollections = result.collections.filter(
            (col) => col.parent_collection_id === parentCollectionId,
          );

          // Process sub-collections for decryption
          const processedSubCollections = [];
          for (const subCol of subCollections) {
            const processed = await processCollection(subCol);
            processedSubCollections.push(processed);
          }

          console.log(
            "[CollectionDetails] Sub-collections found:",
            processedSubCollections.length,
          );
          setSubCollections(processedSubCollections);
        }
      } catch (err) {
        console.error(
          "[CollectionDetails] Failed to load sub-collections:",
          err,
        );
        setSubCollections([]);
      }
    },
    [listCollectionManager],
  );

  // Get states to include based on view mode
  const getStatesToInclude = (mode) => {
    if (!fileManager) return ["active"];

    switch (mode) {
      case "active":
        return [fileManager.FILE_STATES.ACTIVE];
      case "archived":
        return [fileManager.FILE_STATES.ARCHIVED];
      case "deleted":
        return [fileManager.FILE_STATES.DELETED];
      case "pending":
        return [fileManager.FILE_STATES.PENDING];
      case "all":
        return Object.values(fileManager.FILE_STATES);
      default:
        return [fileManager.FILE_STATES.ACTIVE];
    }
  };

  // Handle view mode change
  const handleViewModeChange = (newMode) => {
    setViewMode(newMode);
    // Reload files with new view mode
    if (collectionId && fileManager) {
      loadCollectionFiles(collectionId, false);
    }
  };

  // Handle file download
  const handleDownloadFile = async (fileId, fileName) => {
    if (!downloadFileManager) return;

    const file = files.find((f) => f.id === fileId);
    if (file && fileManager && !fileManager.canDownloadFile(file)) {
      setError("This file cannot be downloaded in its current state.");
      return;
    }

    try {
      console.log(
        "[CollectionDetails] Starting download for:",
        fileId,
        fileName,
      );
      setDownloadingFiles((prev) => new Set(prev).add(fileId));
      setSuccess(`Preparing download for "${fileName}"...`);

      await downloadFileManager.downloadFile(fileId, { saveToDisk: true });

      setSuccess(`File "${fileName}" downloaded successfully!`);
    } catch (err) {
      console.error("[CollectionDetails] Failed to download file:", err);
      setError("Could not download file. Please try again.");
    } finally {
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  // Get time ago
  const getTimeAgo = (dateString) => {
    if (!dateString) return "Recently";
    const now = new Date();
    const date = new Date(dateString);
    const diffInMinutes = Math.floor((now - date) / (1000 * 60));

    if (diffInMinutes < 60) return "Just now";
    if (diffInMinutes < 1440) return "Today";
    if (diffInMinutes < 2880) return "Yesterday";
    return "This week";
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

  // Get file statistics
  const getFileStats = () => {
    if (!fileManager || !collectionId) return {};
    return fileManager.getFileStats(collectionId);
  };

  // Load collection when ready
  useEffect(() => {
    if (
      getCollectionManager &&
      fileManager &&
      collectionId &&
      authManager?.isAuthenticated()
    ) {
      loadCollection();
    }
  }, [
    getCollectionManager,
    fileManager,
    collectionId,
    authManager,
    loadCollection,
  ]);

  // Handle upload to collection
  const handleUploadToCollection = () => {
    navigate(`/file-manager/upload?collection=${collectionId}`, {
      state: {
        preSelectedCollection: {
          id: collectionId,
          name: collection?.name || "Folder",
          type: collection?.type || "folder",
        },
      },
    });
  };

  // Filter files and collections
  const filteredFiles = files.filter((file) =>
    (file.name || "").toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const filteredSubCollections = subCollections.filter((col) =>
    (col.name || "").toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const fileStats = getFileStats();

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

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Navigation />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading folder...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-6">
          <button
            onClick={() => navigate("/file-manager")}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to My Files
          </button>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div
                className={`flex items-center justify-center h-10 w-10 rounded-lg ${
                  collection?.type === "album"
                    ? "bg-pink-100 text-pink-600"
                    : "bg-blue-100 text-blue-600"
                }`}
              >
                {collection?.type === "album" ? (
                  <PhotoIcon className="h-6 w-6" />
                ) : (
                  <FolderIcon className="h-6 w-6" />
                )}
              </div>
              <h1 className="text-2xl font-semibold text-gray-900">
                {collection?.name || "Locked Folder"}
              </h1>
            </div>

            <div className="flex items-center space-x-3">
              <button
                onClick={() => loadCollection(true)}
                disabled={isLoading}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ArrowPathIcon
                  className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
                />
                Refresh
              </button>
              <button
                onClick={() =>
                  navigate(`/file-manager/collections/${collection?.id}/share`)
                }
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ShareIcon className="h-4 w-4 mr-2" />
                Share
              </button>
              <button
                onClick={handleUploadToCollection}
                className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900"
              >
                <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                Upload
              </button>
            </div>
          </div>
        </div>

        {/* Messages */}
        {error && (
          <div className="mb-6 p-3 rounded-lg bg-red-50 border border-red-200">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        {success && (
          <div className="mb-6 p-3 rounded-lg bg-green-50 border border-green-200">
            <p className="text-sm text-green-700">{success}</p>
          </div>
        )}

        {/* Search */}
        <div className="mb-6">
          <div className="relative max-w-md">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search in this folder..."
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
            />
          </div>
        </div>

        {/* File Statistics */}
        {fileStats && Object.keys(fileStats).length > 0 && (
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
              <strong>Total:</strong> {fileStats.total || 0}
            </div>
            <div>
              <strong>Active:</strong>{" "}
              <span style={{ color: getStateColor("active") }}>
                {fileStats.active || 0}
              </span>
            </div>
            <div>
              <strong>Archived:</strong>{" "}
              <span style={{ color: getStateColor("archived") }}>
                {fileStats.archived || 0}
              </span>
            </div>
            <div>
              <strong>Deleted:</strong>{" "}
              <span style={{ color: getStateColor("deleted") }}>
                {fileStats.deleted || 0}
              </span>
            </div>
            <div>
              <strong>Pending:</strong>{" "}
              <span style={{ color: getStateColor("pending") }}>
                {fileStats.pending || 0}
              </span>
            </div>
          </div>
        )}

        {/* View Mode Toggle */}
        {(files.length > 0 || fileStats.total > 0) && (
          <div
            style={{
              display: "flex",
              gap: "10px",
              marginBottom: "20px",
              flexWrap: "wrap",
            }}
          >
            {[
              { key: "active", label: "Active", count: fileStats.active || 0 },
              {
                key: "archived",
                label: "Archived",
                count: fileStats.archived || 0,
              },
              {
                key: "deleted",
                label: "Deleted",
                count: fileStats.deleted || 0,
              },
              {
                key: "pending",
                label: "Pending",
                count: fileStats.pending || 0,
              },
              { key: "all", label: "All", count: fileStats.total || 0 },
            ].map(({ key, label, count }) => (
              <button
                key={key}
                onClick={() => handleViewModeChange(key)}
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
        )}

        {/* Sub-Collections */}
        {filteredSubCollections.length > 0 && (
          <div className="mb-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Folders</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {filteredSubCollections.map((subCol) => (
                <div
                  key={subCol.id}
                  className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md hover:border-gray-300 cursor-pointer transition-all"
                  onClick={() =>
                    navigate(`/file-manager/collections/${subCol.id}`)
                  }
                >
                  <div className="flex items-center space-x-3">
                    <div
                      className={`flex items-center justify-center h-10 w-10 rounded-lg ${
                        subCol.type === "album"
                          ? "bg-pink-100 text-pink-600"
                          : "bg-blue-100 text-blue-600"
                      }`}
                    >
                      {subCol.type === "album" ? (
                        <PhotoIcon className="h-6 w-6" />
                      ) : (
                        <FolderIcon className="h-6 w-6" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="font-medium text-gray-900 truncate">
                        {subCol.name || "Locked Folder"}
                      </h3>
                      <p className="text-sm text-gray-500">
                        {subCol.file_count || 0} items
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Files */}
        <div className="bg-white rounded-lg border border-gray-200">
          {filteredFiles.length === 0 && filteredSubCollections.length === 0 ? (
            <div className="p-8 text-center">
              <DocumentIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <h3 className="text-sm font-medium text-gray-900 mb-2">
                {searchQuery ? "No items found" : `No ${viewMode} items yet`}
              </h3>
              <p className="text-sm text-gray-500 mb-4">
                {searchQuery
                  ? `No items match "${searchQuery}"`
                  : viewMode === "active"
                    ? "Upload files or create folders"
                    : `Switch to "Active" view to see available items`}
              </p>
              {!searchQuery && viewMode === "active" && (
                <div className="flex justify-center space-x-3">
                  <button
                    onClick={handleUploadToCollection}
                    className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900"
                  >
                    <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                    Upload Files
                  </button>
                  <button
                    onClick={() => navigate("/file-manager/collections/create")}
                    className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-gray-700 bg-white border border-gray-300 hover:bg-gray-50"
                  >
                    <PlusIcon className="h-4 w-4 mr-2" />
                    New Folder
                  </button>
                </div>
              )}
            </div>
          ) : filteredFiles.length > 0 ? (
            <div className="divide-y divide-gray-200">
              {filteredFiles.map((file) => (
                <div
                  key={file.id}
                  className="p-4 hover:bg-gray-50 flex items-center justify-between"
                >
                  <div
                    className="flex items-center space-x-3 flex-1 cursor-pointer"
                    onClick={() => navigate(`/file-manager/files/${file.id}`)}
                  >
                    <DocumentIcon className="h-5 w-5 text-gray-400" />
                    <div className="flex-1">
                      <div className="text-sm font-medium text-gray-900">
                        {file.name || "Locked File"}
                      </div>
                      <div className="text-xs text-gray-500">
                        {formatFileSize(
                          file.size || file.encrypted_file_size_in_bytes,
                        )}{" "}
                        •{getTimeAgo(file.modified_at || file.created_at)}
                        {file.state !== "active" && (
                          <span
                            style={{
                              marginLeft: "8px",
                              padding: "2px 6px",
                              borderRadius: "3px",
                              backgroundColor: getStateColor(file.state),
                              color: "white",
                              fontSize: "10px",
                            }}
                          >
                            {file.state?.toUpperCase()}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDownloadFile(file.id, file.name);
                    }}
                    disabled={
                      !file._isDecrypted ||
                      downloadingFiles.has(file.id) ||
                      (fileManager && !fileManager.canDownloadFile(file))
                    }
                    className="p-2 text-gray-400 hover:text-gray-600 disabled:opacity-50"
                    title={
                      !file._isDecrypted
                        ? "File cannot be decrypted"
                        : fileManager && !fileManager.canDownloadFile(file)
                          ? "File cannot be downloaded in its current state"
                          : "Download file"
                    }
                  >
                    {downloadingFiles.has(file.id) ? (
                      <div className="animate-spin rounded-full h-4 w-4 border-b border-gray-400"></div>
                    ) : (
                      <ArrowDownTrayIcon className="h-4 w-4" />
                    )}
                  </button>
                </div>
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(CollectionDetails);
