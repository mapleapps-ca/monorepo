// File: src/pages/User/FileManager/Collections/CollectionDetails.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useParams, useLocation } from "react-router";
import { useFiles, useCrypto, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  FolderIcon,
  PhotoIcon,
  DocumentIcon,
  CloudArrowUpIcon,
  ShareIcon,
  ArrowDownTrayIcon,
  MagnifyingGlassIcon,
  ArrowPathIcon,
  PlusIcon,
  HomeIcon,
  ChevronRightIcon,
  ViewColumnsIcon,
  ListBulletIcon,
  FunnelIcon,
  EllipsisVerticalIcon,
  TrashIcon,
  StarIcon,
  ClockIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";
import { StarIcon as StarIconSolid } from "@heroicons/react/24/solid";

const CollectionDetails = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();
  const location = useLocation();

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
  const [allFiles, setAllFiles] = useState([]);
  const [files, setFiles] = useState([]);
  const [subCollections, setSubCollections] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());
  const [fileManager, setFileManager] = useState(null);

  // Breadcrumb navigation state
  const [breadcrumbs, setBreadcrumbs] = useState([]);
  const [isLoadingBreadcrumbs, setIsLoadingBreadcrumbs] = useState(false);

  // State to track pending refresh operations
  const [pendingRefresh, setPendingRefresh] = useState(false);
  const [refreshReason, setRefreshReason] = useState(null);

  // UI State from new design
  const [viewMode, setViewMode] = useState("grid"); // grid or list
  const [selectedFiles, setSelectedFiles] = useState(new Set());
  const [showDropdown, setShowDropdown] = useState(null);

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

  // Process collection for decryption
  const processCollection = useCallback(
    async (rawCollection) => {
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

      processedCollection.type =
        processedCollection.collection_type || "folder";
      return processedCollection;
    },
    [CollectionCryptoService],
  );

  // Build breadcrumb trail
  const loadBreadcrumbs = useCallback(
    async (currentCollection) => {
      if (!currentCollection || !getCollectionManager) return;

      setIsLoadingBreadcrumbs(true);
      try {
        const breadcrumbItems = [];
        const collectionChain = [];
        let current = currentCollection;

        collectionChain.unshift({
          id: current.id,
          name: current.name || "Locked Folder",
          isCurrent: true,
        });

        while (current.parent_id) {
          const result = await getCollectionManager.getCollection(
            current.parent_id,
          );
          if (result.collection) {
            const processedParent = await processCollection(result.collection);
            collectionChain.unshift({
              id: processedParent.id,
              name: processedParent.name || "Locked Folder",
              isCurrent: false,
            });
            current = processedParent;
          } else {
            break;
          }
        }

        breadcrumbItems.push({
          id: "root",
          name: "My Files",
          path: "/file-manager",
          isRoot: true,
        });

        collectionChain.forEach((collection, index) => {
          const isLast = index === collectionChain.length - 1;
          breadcrumbItems.push({
            id: collection.id,
            name: collection.name,
            path: isLast ? null : `/file-manager/collections/${collection.id}`,
            isRoot: false,
            isCurrent: isLast,
          });
        });

        setBreadcrumbs(breadcrumbItems);
      } catch (error) {
        console.error(
          "[CollectionDetails] Failed to build breadcrumbs:",
          error,
        );
        setBreadcrumbs([
          {
            id: "root",
            name: "My Files",
            path: "/file-manager",
            isRoot: true,
          },
          {
            id: currentCollection.id,
            name: currentCollection.name || "Current Folder",
            path: null,
            isRoot: false,
            isCurrent: true,
          },
        ]);
      } finally {
        setIsLoadingBreadcrumbs(false);
      }
    },
    [getCollectionManager, processCollection],
  );

  // Comprehensive cache clearing function
  const performComprehensiveCacheClearing = useCallback(() => {
    console.log(
      "[CollectionDetails] 🧹 Starting comprehensive cache clearing...",
    );
    if (listCollectionManager) listCollectionManager.clearAllCache();
    if (getCollectionManager) getCollectionManager.clearAllCache();
    if (fileManager) fileManager.clearAllCaches();
    // Omitting other localStorage logic for brevity
  }, [listCollectionManager, getCollectionManager, fileManager]);

  // Filter files based on view mode
  const filterFilesByViewMode = useCallback(
    (filesToFilter, mode = "active") => {
      const filtered = filesToFilter.filter(
        (f) => f.state === "active" || (!f.state && f._is_active),
      );
      setFiles(filtered);
    },
    [],
  );

  // Load ALL files and store them for client-side filtering
  const loadAllCollectionFiles = useCallback(
    async (collectionId, forceRefresh = false) => {
      if (!fileManager) return;

      try {
        const allStates = Object.values(fileManager.FILE_STATES);
        const loadedFiles = await fileManager.listFilesByCollection(
          collectionId,
          allStates,
          forceRefresh,
        );
        setAllFiles(loadedFiles);
        filterFilesByViewMode(loadedFiles);
      } catch (err) {
        console.error("[CollectionDetails] Failed to load files:", err);
        setAllFiles([]);
        setFiles([]);
      }
    },
    [fileManager, filterFilesByViewMode],
  );

  // Load sub-collections
  const loadSubCollections = useCallback(
    async (parentCollectionId, forceRefresh = false) => {
      if (!listCollectionManager) return;
      try {
        const result =
          await listCollectionManager.listCollections(forceRefresh);
        if (result.collections) {
          const subCollections = result.collections.filter(
            (col) => col.parent_id === parentCollectionId,
          );
          const processedSubCollections = await Promise.all(
            subCollections.map((subCol) => processCollection(subCol)),
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
    [listCollectionManager, processCollection],
  );

  // Main load collection function
  const loadCollection = useCallback(
    async (forceRefresh = false) => {
      if (!getCollectionManager || !collectionId || !fileManager) return;
      setIsLoading(true);
      setError("");
      try {
        if (forceRefresh) {
          performComprehensiveCacheClearing();
        }
        const result = await getCollectionManager.getCollection(
          collectionId,
          forceRefresh,
        );
        if (result.collection) {
          const processedCollection = await processCollection(
            result.collection,
          );
          setCollection(processedCollection);
          if (result.collection.collection_key) {
            CollectionCryptoService.cacheCollectionKey(
              collectionId,
              result.collection.collection_key,
            );
          }
          await loadBreadcrumbs(processedCollection);
        }
        await loadAllCollectionFiles(collectionId, forceRefresh);
        await loadSubCollections(collectionId, forceRefresh);
      } catch (err) {
        console.error("[CollectionDetails] Failed to load collection:", err);
        setError("Could not load folder. Please try again.");
      } finally {
        setIsLoading(false);
      }
    },
    [
      getCollectionManager,
      collectionId,
      fileManager,
      CollectionCryptoService,
      performComprehensiveCacheClearing,
      processCollection,
      loadBreadcrumbs,
      loadAllCollectionFiles,
      loadSubCollections,
    ],
  );

  // Refresh detection from navigation state
  useEffect(() => {
    const shouldRefresh = location.state?.refresh;
    if (shouldRefresh) {
      if (location.state?.newCollectionCreated) {
        setRefreshReason("collection_creation");
      }
      if (location.state?.refreshFiles || location.state?.forceFileRefresh) {
        setRefreshReason("file_upload");
        if (location.state?.uploadedFileCount) {
          setSuccess(
            `${location.state.uploadedFileCount} file(s) uploaded successfully!`,
          );
        }
      }
      setPendingRefresh(true);
      navigate(location.pathname, { replace: true, state: {} });
    }
  }, [location.state, navigate, location.pathname]);

  // Execute pending refresh
  useEffect(() => {
    if (
      pendingRefresh &&
      getCollectionManager &&
      fileManager &&
      collectionId &&
      authManager?.isAuthenticated()
    ) {
      loadCollection(true);
      setPendingRefresh(false);
      setRefreshReason(null);
    }
  }, [
    pendingRefresh,
    getCollectionManager,
    fileManager,
    collectionId,
    authManager,
    loadCollection,
  ]);

  // Initial load
  useEffect(() => {
    if (
      getCollectionManager &&
      fileManager &&
      collectionId &&
      authManager?.isAuthenticated() &&
      !pendingRefresh
    ) {
      loadCollection();
    }
  }, [
    getCollectionManager,
    fileManager,
    collectionId,
    authManager,
    loadCollection,
    pendingRefresh,
  ]);

  // Handlers
  const handleCreateSubCollection = () => {
    navigate("/file-manager/collections/create", {
      state: {
        parentCollectionId: collectionId,
        parentCollectionName: collection?.name || "Parent Folder",
      },
    });
  };

  const handleDownloadFile = async (fileId, fileName) => {
    if (!downloadFileManager) return;
    const file = files.find((f) => f.id === fileId);
    if (file && fileManager && !fileManager.canDownloadFile(file)) {
      setError("This file cannot be downloaded in its current state.");
      return;
    }
    try {
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

  const handleUploadToCollection = () => {
    navigate(`/file-manager/upload?collection=${collectionId}`);
  };

  const handleManualRefresh = async () => {
    await loadCollection(true);
  };

  // UI Helpers
  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

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

  const getFileIcon = (mimeType, sizeClass = "h-10 w-10") => {
    if (!mimeType)
      return <DocumentIcon className={`${sizeClass} text-gray-600`} />;
    if (mimeType.includes("pdf"))
      return <DocumentIcon className={`${sizeClass} text-red-600`} />;
    if (mimeType.includes("word"))
      return <DocumentIcon className={`${sizeClass} text-blue-600`} />;
    if (mimeType.includes("sheet"))
      return <DocumentIcon className={`${sizeClass} text-green-600`} />;
    if (mimeType.startsWith("image/"))
      return <PhotoIcon className={`${sizeClass} text-purple-600`} />;
    return <DocumentIcon className={`${sizeClass} text-gray-600`} />;
  };

  const toggleFileSelection = (fileId) => {
    const newSelection = new Set(selectedFiles);
    if (newSelection.has(fileId)) newSelection.delete(fileId);
    else newSelection.add(fileId);
    setSelectedFiles(newSelection);
  };

  // Filtered data for rendering
  const filteredFiles = files.filter((file) =>
    (file.name || "").toLowerCase().includes(searchQuery.toLowerCase()),
  );
  const filteredSubCollections = subCollections.filter((col) =>
    (col.name || "").toLowerCase().includes(searchQuery.toLowerCase()),
  );

  // Clear messages
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
              <p className="text-gray-600">
                Loading folder...
                {pendingRefresh && (
                  <span className="block text-sm text-blue-600 mt-2">
                    🔄 Refreshing with latest data
                  </span>
                )}
              </p>
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
        {/* Breadcrumbs */}
        <nav className="flex items-center space-x-2 text-sm mb-6 animate-fade-in-down">
          {breadcrumbs.map((crumb, index) => (
            <React.Fragment key={crumb.id}>
              {index > 0 && (
                <ChevronRightIcon className="h-4 w-4 text-gray-400" />
              )}
              {crumb.isCurrent ? (
                <span className="font-medium text-gray-900">{crumb.name}</span>
              ) : (
                <button
                  onClick={() => navigate(crumb.path)}
                  className="text-gray-600 hover:text-gray-900 transition-colors duration-200 flex items-center space-x-1"
                >
                  {crumb.isRoot && <HomeIcon className="h-4 w-4" />}
                  <span>{crumb.name}</span>
                </button>
              )}
            </React.Fragment>
          ))}
        </nav>

        {/* Header */}
        <div className="mb-8 animate-fade-in-up">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div
                className={`h-14 w-14 rounded-xl flex items-center justify-center ${
                  collection?.type === "album"
                    ? "bg-gradient-to-br from-pink-500 to-pink-600"
                    : "bg-gradient-to-br from-blue-500 to-blue-600"
                }`}
              >
                {collection?.type === "album" ? (
                  <PhotoIcon className="h-7 w-7 text-white" />
                ) : (
                  <FolderIcon className="h-7 w-7 text-white" />
                )}
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">
                  {collection?.name || "Locked Folder"}
                </h1>
                <p className="text-sm text-gray-600 mt-1">
                  {filteredSubCollections.length + filteredFiles.length} items
                </p>
              </div>
            </div>

            <div className="flex items-center space-x-3">
              <button onClick={handleManualRefresh} className="btn-secondary">
                <ArrowPathIcon className="h-4 w-4 mr-2" />
                Refresh
              </button>
              <button
                onClick={() =>
                  navigate(`/file-manager/collections/${collection?.id}/share`)
                }
                className="btn-secondary"
              >
                <ShareIcon className="h-4 w-4 mr-2" />
                Share
              </button>
              <button
                onClick={handleUploadToCollection}
                className="btn-primary"
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

        {/* Toolbar */}
        <div
          className="mb-6 animate-fade-in-up"
          style={{ animationDelay: "100ms" }}
        >
          <div className="bg-white p-4 rounded-lg border border-gray-200">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4 flex-1">
                <div className="relative flex-1 max-w-md">
                  <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search in this folder..."
                    className="w-full pl-10 pr-3 py-2 h-10 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                  />
                </div>
                <div className="flex items-center bg-gray-100 rounded-lg p-1">
                  <button
                    onClick={() => setViewMode("grid")}
                    className={`p-2 rounded ${
                      viewMode === "grid"
                        ? "bg-white text-gray-900 shadow-sm"
                        : "text-gray-600 hover:text-gray-900"
                    } transition-all duration-200`}
                  >
                    <ViewColumnsIcon className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => setViewMode("list")}
                    className={`p-2 rounded ${
                      viewMode === "list"
                        ? "bg-white text-gray-900 shadow-sm"
                        : "text-gray-600 hover:text-gray-900"
                    } transition-all duration-200`}
                  >
                    <ListBulletIcon className="h-4 w-4" />
                  </button>
                </div>
              </div>
              <div className="flex items-center space-x-3">
                <button className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-all duration-200">
                  <FunnelIcon className="h-5 w-5" />
                </button>
                <button
                  onClick={handleCreateSubCollection}
                  className="btn-secondary"
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  New Folder
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Selected Items Bar */}
        {selectedFiles.size > 0 && (
          <div className="mb-6 animate-fade-in-down">
            <div className="bg-blue-50 border border-blue-200 p-4 rounded-lg">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-4">
                  <CheckCircleIcon className="h-5 w-5 text-blue-600" />
                  <span className="text-sm font-medium text-blue-900">
                    {selectedFiles.size} item{selectedFiles.size > 1 ? "s" : ""}{" "}
                    selected
                  </span>
                </div>
                <div className="flex items-center space-x-3">
                  <button className="text-sm text-blue-700 hover:text-blue-800 font-medium">
                    Download
                  </button>
                  <button className="text-sm text-blue-700 hover:text-blue-800 font-medium">
                    Share
                  </button>
                  <button className="text-sm text-red-700 hover:text-red-800 font-medium">
                    Delete
                  </button>
                  <button
                    onClick={() => setSelectedFiles(new Set())}
                    className="text-sm text-gray-600 hover:text-gray-700"
                  >
                    Clear
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Content */}
        <div className="space-y-6">
          {/* Sub-Collections */}
          {filteredSubCollections.length > 0 && (
            <div
              className="animate-fade-in-up"
              style={{ animationDelay: "200ms" }}
            >
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Folders
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {filteredSubCollections.map((subCol, index) => (
                  <div
                    key={subCol.id}
                    className="bg-white rounded-lg border border-gray-200 hover:shadow-md hover:border-gray-300 cursor-pointer group animate-fade-in-up"
                    style={{ animationDelay: `${(index + 2) * 50}ms` }}
                    onClick={() =>
                      navigate(`/file-manager/collections/${subCol.id}`)
                    }
                  >
                    <div className="p-4">
                      <div className="flex items-center space-x-3">
                        <div
                          className={`h-12 w-12 rounded-xl flex items-center justify-center transition-all duration-200 ${
                            subCol.type === "album"
                              ? "bg-gradient-to-br from-pink-500 to-pink-600 group-hover:from-pink-600 group-hover:to-pink-700"
                              : "bg-gradient-to-br from-blue-500 to-blue-600 group-hover:from-blue-600 group-hover:to-blue-700"
                          }`}
                        >
                          {subCol.type === "album" ? (
                            <PhotoIcon className="h-6 w-6 text-white" />
                          ) : (
                            <FolderIcon className="h-6 w-6 text-white" />
                          )}
                        </div>
                        <div className="flex-1 min-w-0">
                          <h3 className="font-medium text-gray-900 truncate">
                            {subCol.name}
                          </h3>
                          <p className="text-sm text-gray-500">Folder</p>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Files */}
          {filteredFiles.length > 0 && (
            <div
              className="animate-fade-in-up"
              style={{ animationDelay: "300ms" }}
            >
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Files
              </h2>
              {viewMode === "grid" ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  {filteredFiles.map((file, index) => (
                    <div
                      key={file.id}
                      className="bg-white rounded-lg border border-gray-200 hover:shadow-md group animate-fade-in-up relative"
                      style={{ animationDelay: `${(index + 4) * 50}ms` }}
                    >
                      <div
                        className="p-4 cursor-pointer"
                        onClick={() =>
                          navigate(`/file-manager/files/${file.id}`)
                        }
                      >
                        <div
                          className="absolute top-4 left-4"
                          onClick={(e) => e.stopPropagation()}
                        >
                          <input
                            type="checkbox"
                            checked={selectedFiles.has(file.id)}
                            onChange={() => toggleFileSelection(file.id)}
                            className="h-4 w-4 text-red-600 border-gray-300 rounded focus:ring-red-500"
                          />
                        </div>
                        <div className="mb-4 mt-8">
                          <div className="h-32 bg-gray-50 rounded-lg flex items-center justify-center">
                            {getFileIcon(file.mime_type, "h-12 w-12")}
                          </div>
                        </div>
                        <div className="space-y-2">
                          <div className="flex items-start justify-between">
                            <h3 className="text-sm font-medium text-gray-900 truncate flex-1 pr-2">
                              {file.name}
                            </h3>
                            <button
                              onClick={(e) => e.stopPropagation()}
                              className="flex-shrink-0"
                            >
                              {file.isFavorite ? (
                                <StarIconSolid className="h-4 w-4 text-yellow-500" />
                              ) : (
                                <StarIcon className="h-4 w-4 text-gray-400 hover:text-yellow-500" />
                              )}
                            </button>
                          </div>
                          <div className="flex items-center justify-between text-xs text-gray-500">
                            <span>
                              {formatFileSize(
                                file.size || file.encrypted_file_size_in_bytes,
                              )}
                            </span>
                            <span className="flex items-center">
                              <ClockIcon className="h-3 w-3 mr-1" />
                              {getTimeAgo(file.modified_at || file.created_at)}
                            </span>
                          </div>
                        </div>
                        <div
                          className="mt-3 pt-3 border-t border-gray-100 flex items-center justify-center space-x-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200"
                          onClick={(e) => e.stopPropagation()}
                        >
                          <button
                            onClick={() =>
                              handleDownloadFile(file.id, file.name)
                            }
                            disabled={downloadingFiles.has(file.id)}
                            className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-all duration-200 disabled:opacity-50"
                          >
                            {downloadingFiles.has(file.id) ? (
                              <div className="animate-spin rounded-full h-4 w-4 border-b border-gray-400"></div>
                            ) : (
                              <ArrowDownTrayIcon className="h-4 w-4" />
                            )}
                          </button>
                          <button className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-all duration-200">
                            <ShareIcon className="h-4 w-4" />
                          </button>
                          <button className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-all duration-200">
                            <TrashIcon className="h-4 w-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
                  <table className="w-full">
                    <thead className="bg-gray-50 border-b border-gray-200">
                      <tr>
                        <th className="px-4 py-3 text-left">
                          <input
                            type="checkbox"
                            checked={
                              selectedFiles.size === filteredFiles.length &&
                              filteredFiles.length > 0
                            }
                            onChange={() => {
                              if (selectedFiles.size === filteredFiles.length) {
                                setSelectedFiles(new Set());
                              } else {
                                setSelectedFiles(
                                  new Set(filteredFiles.map((f) => f.id)),
                                );
                              }
                            }}
                            className="h-4 w-4 text-red-600 border-gray-300 rounded focus:ring-red-500"
                          />
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Name
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Size
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Modified
                        </th>
                        <th className="px-4 py-3"></th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                      {filteredFiles.map((file) => (
                        <tr
                          key={file.id}
                          className="hover:bg-gray-50 transition-colors duration-200 cursor-pointer"
                          onClick={() =>
                            navigate(`/file-manager/files/${file.id}`)
                          }
                        >
                          <td
                            className="px-4 py-3"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <input
                              type="checkbox"
                              checked={selectedFiles.has(file.id)}
                              onChange={() => toggleFileSelection(file.id)}
                              className="h-4 w-4 text-red-600 border-gray-300 rounded focus:ring-red-500"
                            />
                          </td>
                          <td className="px-4 py-3">
                            <div className="flex items-center space-x-3">
                              {getFileIcon(file.mime_type, "h-5 w-5")}
                              <span className="text-sm font-medium text-gray-900">
                                {file.name}
                              </span>
                              {file.isFavorite && (
                                <StarIconSolid className="h-4 w-4 text-yellow-500" />
                              )}
                            </div>
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-600">
                            {formatFileSize(
                              file.size || file.encrypted_file_size_in_bytes,
                            )}
                          </td>
                          <td className="px-4 py-3 text-sm text-gray-600">
                            {getTimeAgo(file.modified_at || file.created_at)}
                          </td>
                          <td
                            className="px-4 py-3 text-right"
                            onClick={(e) => e.stopPropagation()}
                          >
                            <button
                              onClick={() =>
                                handleDownloadFile(file.id, file.name)
                              }
                              disabled={downloadingFiles.has(file.id)}
                              className="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-all duration-200 disabled:opacity-50"
                            >
                              {downloadingFiles.has(file.id) ? (
                                <div className="animate-spin rounded-full h-4 w-4 border-b border-gray-400"></div>
                              ) : (
                                <ArrowDownTrayIcon className="h-4 w-4" />
                              )}
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {/* Empty State */}
          {filteredFiles.length === 0 &&
            filteredSubCollections.length === 0 && (
              <div className="text-center py-24 animate-fade-in">
                <div className="h-20 w-20 bg-gray-100 rounded-2xl flex items-center justify-center mx-auto mb-6">
                  <DocumentIcon className="h-10 w-10 text-gray-400" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-2">
                  {searchQuery ? "No items found" : "This folder is empty"}
                </h3>
                <p className="text-gray-600 mb-8 max-w-md mx-auto">
                  {searchQuery
                    ? `No items match "${searchQuery}"`
                    : "Upload files or create folders to get started"}
                </p>
                {!searchQuery && (
                  <div className="flex justify-center space-x-3">
                    <button
                      onClick={handleUploadToCollection}
                      className="btn-primary"
                    >
                      <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                      Upload Files
                    </button>
                    <button
                      onClick={handleCreateSubCollection}
                      className="btn-secondary"
                    >
                      <PlusIcon className="h-4 w-4 mr-2" />
                      New Folder
                    </button>
                  </div>
                )}
              </div>
            )}
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(CollectionDetails);
