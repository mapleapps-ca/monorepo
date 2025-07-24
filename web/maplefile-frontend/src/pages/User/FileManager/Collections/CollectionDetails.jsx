// File: src/pages/User/FileManager/Collections/CollectionDetails.jsx
import React, { useState, useEffect, useCallback } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { useFiles, useCrypto, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  FolderIcon,
  PhotoIcon,
  DocumentIcon,
  ArrowLeftIcon,
  ArrowUpTrayIcon,
  ShareIcon,
  PencilIcon,
  TrashIcon,
  ArrowDownTrayIcon,
  MagnifyingGlassIcon,
  Squares2X2Icon,
  ListBulletIcon,
  EllipsisVerticalIcon,
  ChevronRightIcon,
  HomeIcon,
  StarIcon,
  UsersIcon,
  LockClosedIcon,
  ChevronDownIcon,
  ClockIcon,
  InformationCircleIcon,
  CheckIcon,
  PlusIcon,
  XMarkIcon,
  DocumentDuplicateIcon,
  ArrowsRightLeftIcon,
  FolderPlusIcon,
  ShieldCheckIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
  CloudArrowUpIcon,
} from "@heroicons/react/24/outline";
import { StarIcon as StarIconSolid } from "@heroicons/react/24/solid";

const CollectionDetails = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();

  // Get services from context
  const { getCollectionManager, listFileManager, downloadFileManager } =
    useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(true);
  const [isFilesLoading, setIsFilesLoading] = useState(false);
  const [error, setError] = useState("");
  const [collection, setCollection] = useState(null);
  const [files, setFiles] = useState([]);
  const [subCollections, setSubCollections] = useState([]);
  const [viewMode, setViewMode] = useState("grid");
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [searchQuery, setSearchQuery] = useState("");
  const [showInfo, setShowInfo] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());

  // Load collection details
  const loadCollection = useCallback(
    async (forceRefresh = false) => {
      if (!getCollectionManager || !collectionId) return;

      setIsLoading(true);
      setError("");

      try {
        console.log("[CollectionDetails] === Loading Collection ===");
        console.log("[CollectionDetails] Collection ID:", collectionId);
        console.log("[CollectionDetails] Force refresh:", forceRefresh);

        // Step 1: Get collection from GetCollectionManager
        const result = await getCollectionManager.getCollection(
          collectionId,
          forceRefresh,
        );

        console.log("[CollectionDetails] Collection loaded:", {
          id: result.collection?.id,
          name: result.collection?.name,
          hasCollectionKey: !!result.collection?.collection_key,
          source: result.source,
        });

        if (!result.collection) {
          throw new Error("Collection not found or access denied");
        }

        // Step 2: Process collection for proper display
        const processedCollection = await processCollection(result.collection);
        setCollection(processedCollection);

        // Step 3: Load files in this collection
        await loadCollectionFiles(collectionId, forceRefresh);

        console.log(
          "[CollectionDetails] Collection details loaded successfully",
        );
      } catch (err) {
        console.error("[CollectionDetails] Failed to load collection:", err);
        setError(err.message);
        setCollection(null);
      } finally {
        setIsLoading(false);
      }
    },
    [getCollectionManager, collectionId],
  );

  // Process collection for display
  const processCollection = async (rawCollection) => {
    console.log("[CollectionDetails] === Processing Collection ===");

    let processedCollection = { ...rawCollection };

    // Check if collection needs decryption
    if (rawCollection.encrypted_name || rawCollection.encrypted_description) {
      console.log(
        "[CollectionDetails] Collection is encrypted, attempting decryption",
      );

      try {
        // Get collection key (should be available from getCollection call)
        let collectionKey = rawCollection.collection_key;

        if (!collectionKey) {
          // Try to get from cache
          collectionKey = CollectionCryptoService.getCachedCollectionKey(
            rawCollection.id,
          );
        }

        if (collectionKey) {
          // Import CollectionCryptoService for decryption
          const { default: CollectionCryptoServiceClass } = await import(
            "../../../../services/Crypto/CollectionCryptoService.js"
          );

          const decryptedCollection =
            await CollectionCryptoServiceClass.decryptCollectionFromAPI(
              rawCollection,
              collectionKey,
            );

          processedCollection = decryptedCollection;
          console.log(
            `[CollectionDetails] ✅ Collection decrypted: ${decryptedCollection.name}`,
          );
        } else {
          console.warn(
            "[CollectionDetails] No collection key available for decryption",
          );
          processedCollection.name = "[Collection key unavailable]";
          processedCollection.description = "[Collection key unavailable]";
          processedCollection._isDecrypted = false;
          processedCollection._decryptionError = "Collection key not available";
        }
      } catch (decryptError) {
        console.error(
          "[CollectionDetails] Failed to decrypt collection:",
          decryptError,
        );
        processedCollection.name = "[Decryption failed]";
        processedCollection.description = "[Decryption failed]";
        processedCollection._isDecrypted = false;
        processedCollection._decryptionError = decryptError.message;
      }
    } else {
      // Collection is not encrypted or already decrypted
      processedCollection._isDecrypted = true;
    }

    // Add UI-specific properties
    processedCollection.type = processedCollection.collection_type || "folder";
    processedCollection.totalSize = formatCollectionSize(
      processedCollection.total_size,
    );
    processedCollection.itemCount =
      (processedCollection.file_count || 0) +
      (processedCollection.sub_collection_count || 0);
    processedCollection.starred = false; // TODO: implement starring functionality

    return processedCollection;
  };

  // Load files in collection
  const loadCollectionFiles = useCallback(
    async (collectionId, forceRefresh = false) => {
      if (!listFileManager) {
        console.warn("[CollectionDetails] ListFileManager not available");
        return;
      }

      setIsFilesLoading(true);

      try {
        console.log("[CollectionDetails] === Loading Collection Files ===");
        console.log("[CollectionDetails] Collection ID:", collectionId);

        // Get files in this collection
        const result = await listFileManager.listFiles({
          collection_id: collectionId,
          force_refresh: forceRefresh,
        });

        console.log("[CollectionDetails] Files loaded:", {
          count: result.files?.length || 0,
          hasResults: !!result.files,
        });

        if (result.files) {
          // Process files for proper decryption
          const processedFiles = await processFiles(result.files, collectionId);
          setFiles(processedFiles);
        } else {
          setFiles([]);
        }

        // TODO: Load sub-collections if the API supports it
        // For now, we'll assume no sub-collections
        setSubCollections([]);
      } catch (err) {
        console.error(
          "[CollectionDetails] Failed to load collection files:",
          err,
        );
        // Don't set error here as it's secondary to the main collection loading
        setFiles([]);
        setSubCollections([]);
      } finally {
        setIsFilesLoading(false);
      }
    },
    [listFileManager],
  );

  // Process files for decryption
  const processFiles = async (rawFiles, collectionId) => {
    console.log("[CollectionDetails] === Processing Files for Decryption ===");

    if (!Array.isArray(rawFiles) || rawFiles.length === 0) {
      return [];
    }

    const processedFiles = [];

    // Get collection key for file decryption
    const collectionKey =
      CollectionCryptoService.getCachedCollectionKey(collectionId);

    if (!collectionKey) {
      console.warn(
        "[CollectionDetails] No collection key available for file decryption",
      );
      // Return files as-is but mark as not decrypted
      return rawFiles.map((file) => ({
        ...file,
        name: "[Collection key unavailable]",
        _isDecrypted: false,
        _decryptionError: "Collection key not available",
      }));
    }

    // Import FileCryptoService for decryption
    const { default: FileCryptoService } = await import(
      "../../../../services/Crypto/FileCryptoService.js"
    );

    for (const file of rawFiles) {
      try {
        // Decrypt file with collection key
        const decryptedFile = await FileCryptoService.decryptFileFromAPI(
          file,
          collectionKey,
        );

        // Add UI-specific properties
        decryptedFile.type = getFileTypeFromMime(decryptedFile.mime_type);
        decryptedFile.size = formatFileSize(decryptedFile.size);
        decryptedFile.modified = getTimeAgo(
          decryptedFile.updated_at || decryptedFile.created_at,
        );
        decryptedFile.starred = false; // TODO: implement starring

        processedFiles.push(decryptedFile);

        if (decryptedFile._isDecrypted) {
          console.log(
            `[CollectionDetails] ✅ File decrypted: ${decryptedFile.name}`,
          );
        } else {
          console.log(
            `[CollectionDetails] ❌ File decryption failed: ${decryptedFile._decryptionError}`,
          );
        }
      } catch (error) {
        console.error(
          `[CollectionDetails] Failed to decrypt file ${file.id}:`,
          error,
        );
        processedFiles.push({
          ...file,
          name: `[Decrypt failed: ${error.message.substring(0, 30)}...]`,
          _isDecrypted: false,
          _decryptionError: error.message,
          type: getFileTypeFromMime(file.mime_type),
          size: formatFileSize(file.size),
          modified: getTimeAgo(file.updated_at || file.created_at),
          starred: false,
        });
      }
    }

    return processedFiles;
  };

  // Utility functions
  const formatCollectionSize = (bytes) => {
    if (!bytes || bytes === 0) return "0 B";

    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));

    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  const formatFileSize = (bytes) => {
    if (!bytes || bytes === 0) return "0 B";

    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));

    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  const getFileTypeFromMime = (mimeType) => {
    if (!mimeType) return "document";
    if (mimeType.includes("pdf")) return "pdf";
    if (mimeType.includes("sheet") || mimeType.includes("excel"))
      return "spreadsheet";
    if (mimeType.includes("document") || mimeType.includes("word"))
      return "document";
    if (mimeType.includes("presentation") || mimeType.includes("powerpoint"))
      return "presentation";
    if (mimeType.startsWith("image/")) return "image";
    if (mimeType.startsWith("video/")) return "video";
    if (mimeType.startsWith("audio/")) return "audio";
    return "document";
  };

  const getTimeAgo = (dateString) => {
    if (!dateString) return "Unknown";

    const now = new Date();
    const date = new Date(dateString);
    const diffInMinutes = Math.floor((now - date) / (1000 * 60));

    if (diffInMinutes < 1) return "Just now";
    if (diffInMinutes < 60) return `${diffInMinutes} minutes ago`;
    if (diffInMinutes < 1440)
      return `${Math.floor(diffInMinutes / 60)} hours ago`;
    if (diffInMinutes < 10080)
      return `${Math.floor(diffInMinutes / 1440)} days ago`;
    return date.toLocaleDateString();
  };

  // Load collection when component mounts or collectionId changes
  useEffect(() => {
    if (
      getCollectionManager &&
      collectionId &&
      authManager?.isAuthenticated()
    ) {
      loadCollection();
    }
  }, [getCollectionManager, collectionId, authManager, loadCollection]);

  // Handle item selection
  const handleSelectItem = (id) => {
    const newSelected = new Set(selectedItems);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedItems(newSelected);
  };

  // Handle refresh
  const handleRefresh = async () => {
    await loadCollection(true);
  };

  // Handle file download
  const handleDownloadFile = async (fileId, fileName) => {
    if (!downloadFileManager) return;

    try {
      setDownloadingFiles((prev) => new Set(prev).add(fileId));

      await downloadFileManager.downloadFile(fileId, {
        saveToDisk: true,
      });

      console.log(
        "[CollectionDetails] File downloaded successfully:",
        fileName,
      );
    } catch (err) {
      console.error("[CollectionDetails] Failed to download file:", err);
      setError(`Failed to download file: ${err.message}`);
    } finally {
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  // Get file icon component
  const getFileIcon = (type) => {
    const iconClass = "h-5 w-5";
    switch (type) {
      case "pdf":
        return (
          <div className="text-red-600">
            <DocumentIcon className={iconClass} />
          </div>
        );
      case "spreadsheet":
        return (
          <div className="text-green-600">
            <DocumentIcon className={iconClass} />
          </div>
        );
      case "document":
        return (
          <div className="text-blue-600">
            <DocumentIcon className={iconClass} />
          </div>
        );
      case "presentation":
        return (
          <div className="text-orange-600">
            <DocumentIcon className={iconClass} />
          </div>
        );
      case "image":
        return (
          <div className="text-purple-600">
            <PhotoIcon className={iconClass} />
          </div>
        );
      default:
        return <DocumentIcon className={iconClass} />;
    }
  };

  // Filter files based on search
  const filteredFiles = files.filter((file) =>
    (file.name || "").toLowerCase().includes(searchQuery.toLowerCase()),
  );

  // Show loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
        <Navigation />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-600 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading collection...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Show error state
  if (error && !collection) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
        <Navigation />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
              <div>
                <h3 className="text-sm font-medium text-red-800">
                  Error loading collection
                </h3>
                <p className="text-sm text-red-700 mt-1">{error}</p>
                <div className="mt-4">
                  <button
                    onClick={() => navigate("/file-manager")}
                    className="text-sm bg-red-100 text-red-800 px-3 py-1 rounded hover:bg-red-200"
                  >
                    ← Back to Collections
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!collection) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Breadcrumb */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 mb-6">
          <HomeIcon className="h-4 w-4" />
          <ChevronRightIcon className="h-3 w-3" />
          <Link to="/file-manager" className="hover:text-gray-900">
            My Files
          </Link>
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900">
            {collection.name || "[Encrypted]"}
          </span>
        </div>

        {/* Collection Header */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex items-start justify-between">
            <div className="flex items-start space-x-4">
              <div
                className={`flex items-center justify-center h-14 w-14 rounded-lg ${
                  collection.type === "album"
                    ? "bg-pink-100 text-pink-600"
                    : "bg-blue-100 text-blue-600"
                } ${!collection._isDecrypted ? "opacity-50" : ""}`}
              >
                {collection.type === "album" ? (
                  <PhotoIcon className="h-7 w-7" />
                ) : (
                  <FolderIcon className="h-7 w-7" />
                )}
                {!collection._isDecrypted && (
                  <LockClosedIcon className="h-4 w-4 absolute" />
                )}
              </div>
              <div>
                <div className="flex items-center space-x-3">
                  <h1 className="text-2xl font-bold text-gray-900">
                    {collection.name || "[Encrypted]"}
                  </h1>
                  <button>
                    {collection.starred ? (
                      <StarIconSolid className="h-5 w-5 text-yellow-400" />
                    ) : (
                      <StarIcon className="h-5 w-5 text-gray-300 hover:text-yellow-400" />
                    )}
                  </button>
                </div>
                {collection._decryptionError && (
                  <p className="text-red-600 text-sm mt-1">
                    ⚠️ {collection._decryptionError}
                  </p>
                )}
                <p className="text-gray-600 mt-1">
                  {collection.description || "No description"}
                </p>
                <div className="flex items-center space-x-4 mt-3 text-sm text-gray-500">
                  <span className="flex items-center">
                    <DocumentIcon className="h-4 w-4 mr-1" />
                    {collection.itemCount || 0} items
                  </span>
                  <span>•</span>
                  <span>{collection.totalSize}</span>
                  <span>•</span>
                  <span className="flex items-center">
                    <ClockIcon className="h-4 w-4 mr-1" />
                    Modified{" "}
                    {getTimeAgo(collection.updated_at || collection.created_at)}
                  </span>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex items-center space-x-2">
              <button
                onClick={handleRefresh}
                disabled={isLoading}
                className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ArrowPathIcon className="h-4 w-4 mr-2" />
                {isLoading ? "Refreshing..." : "Refresh"}
              </button>
              <button
                onClick={() => setShowShareModal(true)}
                className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ShareIcon className="h-4 w-4 mr-2" />
                Share
              </button>
              <button
                onClick={() =>
                  navigate(`/file-manager/collections/${collection.id}/edit`)
                }
                className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <PencilIcon className="h-4 w-4 mr-2" />
                Edit
              </button>
              <button
                onClick={() => setShowInfo(!showInfo)}
                className="p-2 text-gray-400 hover:text-gray-600"
              >
                <InformationCircleIcon className="h-5 w-5" />
              </button>
            </div>
          </div>

          {/* Collection Info Panel */}
          {showInfo && (
            <div className="mt-6 pt-6 border-t border-gray-200">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <h3 className="text-sm font-semibold text-gray-600 mb-2">
                    Details
                  </h3>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-500">Type:</span>
                      <span className="font-medium capitalize">
                        {collection.type}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Created:</span>
                      <span className="font-medium">
                        {collection.created_at
                          ? new Date(collection.created_at).toLocaleDateString()
                          : "Unknown"}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Version:</span>
                      <span className="font-medium">
                        {collection.version || "1"}
                      </span>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-semibold text-gray-600 mb-2">
                    Encryption
                  </h3>
                  <div className="space-y-2 text-sm">
                    <div
                      className={`flex items-center ${collection._isDecrypted ? "text-green-600" : "text-red-600"}`}
                    >
                      <ShieldCheckIcon className="h-4 w-4 mr-2" />
                      <span className="font-medium">
                        {collection._isDecrypted
                          ? "Decrypted Successfully"
                          : "Decryption Failed"}
                      </span>
                    </div>
                    <div className="text-gray-500">
                      ChaCha20-Poly1305 encryption
                    </div>
                    <div className="text-gray-500">
                      Zero-knowledge architecture
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-semibold text-gray-600 mb-2">
                    Files
                  </h3>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-500">Total Files:</span>
                      <span className="font-medium">{files.length}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Decrypted:</span>
                      <span className="font-medium text-green-600">
                        {files.filter((f) => f._isDecrypted).length}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Failed:</span>
                      <span className="font-medium text-red-600">
                        {files.filter((f) => !f._isDecrypted).length}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Security Notice */}
        <div className="bg-gradient-to-r from-green-50 to-blue-50 rounded-lg border border-green-100 p-4 mb-6">
          <div className="flex items-center justify-center mb-2">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-1">
                <LockClosedIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-semibold text-green-800">
                  End-to-End Encrypted
                </span>
              </div>
              <div className="flex items-center space-x-1">
                <ShieldCheckIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-semibold text-blue-800">
                  Zero-Knowledge Architecture
                </span>
              </div>
            </div>
          </div>
          <p className="text-center text-xs text-gray-600">
            All files in this collection are encrypted on your device before
            storage
          </p>
        </div>

        {/* Toolbar */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <button
                onClick={() => navigate("/file-manager/upload")}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ArrowUpTrayIcon className="h-4 w-4 mr-2" />
                Upload
              </button>
              <button className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                <FolderPlusIcon className="h-4 w-4 mr-2" />
                New Folder
              </button>

              <div className="h-6 w-px bg-gray-300"></div>

              {/* Search */}
              <div className="relative flex-1 max-w-md">
                <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search in this collection..."
                  className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                />
              </div>
            </div>

            {/* View Toggle */}
            <div className="flex items-center bg-gray-100 rounded-lg p-1">
              <button
                onClick={() => setViewMode("grid")}
                className={`p-1.5 rounded ${viewMode === "grid" ? "bg-white shadow-sm" : ""}`}
              >
                <Squares2X2Icon
                  className={`h-4 w-4 ${viewMode === "grid" ? "text-red-600" : "text-gray-600"}`}
                />
              </button>
              <button
                onClick={() => setViewMode("list")}
                className={`p-1.5 rounded ${viewMode === "list" ? "bg-white shadow-sm" : ""}`}
              >
                <ListBulletIcon
                  className={`h-4 w-4 ${viewMode === "list" ? "text-red-600" : "text-gray-600"}`}
                />
              </button>
            </div>
          </div>
        </div>

        {/* Sub-collections */}
        {subCollections.length > 0 && (
          <div className="mb-6">
            <h2 className="text-sm font-semibold text-gray-600 mb-3">
              Folders
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
              {subCollections.map((subCollection) => (
                <div
                  key={subCollection.id}
                  className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-all duration-200 cursor-pointer"
                  onDoubleClick={() =>
                    navigate(`/file-manager/collections/${subCollection.id}`)
                  }
                >
                  <div className="flex items-center">
                    <div className="flex items-center justify-center h-10 w-10 bg-blue-100 text-blue-600 rounded-lg mr-3">
                      <FolderIcon className="h-5 w-5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-gray-900 truncate">
                        {subCollection.name}
                      </p>
                      <p className="text-xs text-gray-500">
                        {subCollection.items} items • {subCollection.modified}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Files Loading State */}
        {isFilesLoading && (
          <div className="flex items-center justify-center py-8">
            <div className="text-center">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-red-600 mx-auto mb-2"></div>
              <p className="text-sm text-gray-600">Loading files...</p>
            </div>
          </div>
        )}

        {/* Files */}
        {!isFilesLoading && (
          <div>
            <h2 className="text-sm font-semibold text-gray-600 mb-3">
              Files ({filteredFiles.length})
            </h2>

            {filteredFiles.length === 0 ? (
              <div className="text-center py-12">
                <DocumentIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
                <h3 className="text-sm font-medium text-gray-900 mb-2">
                  {searchQuery ? "No files found" : "No files yet"}
                </h3>
                <p className="text-sm text-gray-500 mb-4">
                  {searchQuery
                    ? `No files match "${searchQuery}"`
                    : "Upload your first files to this collection"}
                </p>
                {!searchQuery && (
                  <button
                    onClick={() => navigate("/file-manager/upload")}
                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
                  >
                    <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                    Upload Files
                  </button>
                )}
              </div>
            ) : viewMode === "grid" ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {filteredFiles.map((file) => (
                  <div
                    key={file.id}
                    className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-all duration-200 cursor-pointer group"
                  >
                    <div className="absolute top-2 left-2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <input
                        type="checkbox"
                        checked={selectedItems.has(file.id)}
                        onChange={() => handleSelectItem(file.id)}
                        className="h-4 w-4 text-red-600 rounded border-gray-300"
                      />
                    </div>

                    <div className="flex flex-col items-center text-center">
                      <div
                        className={`flex items-center justify-center h-12 w-12 rounded-lg mb-3 relative ${
                          file.type === "pdf"
                            ? "bg-red-100"
                            : file.type === "spreadsheet"
                              ? "bg-green-100"
                              : file.type === "document"
                                ? "bg-blue-100"
                                : file.type === "image"
                                  ? "bg-purple-100"
                                  : "bg-orange-100"
                        } ${!file._isDecrypted ? "opacity-50" : ""}`}
                      >
                        {getFileIcon(file.type)}
                        {!file._isDecrypted && (
                          <LockClosedIcon className="h-3 w-3 absolute" />
                        )}
                      </div>
                      <p className="font-medium text-gray-900 truncate w-full">
                        {file.name || "[Encrypted]"}
                      </p>
                      {file._decryptionError && (
                        <p className="text-xs text-red-500 mt-1">
                          Decryption failed
                        </p>
                      )}
                      <p className="text-xs text-gray-500 mt-1">{file.size}</p>
                      <p className="text-xs text-gray-400">{file.modified}</p>
                    </div>

                    <div className="flex items-center justify-between mt-3 pt-3 border-t">
                      <button
                        className="text-gray-400 hover:text-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
                        disabled={
                          !file._isDecrypted || downloadingFiles.has(file.id)
                        }
                        onClick={() => handleDownloadFile(file.id, file.name)}
                      >
                        {downloadingFiles.has(file.id) ? (
                          <div className="animate-spin rounded-full h-4 w-4 border-b border-gray-400"></div>
                        ) : (
                          <ArrowDownTrayIcon className="h-4 w-4" />
                        )}
                      </button>
                      <button>
                        {file.starred ? (
                          <StarIconSolid className="h-4 w-4 text-yellow-400" />
                        ) : (
                          <StarIcon className="h-4 w-4 text-gray-300 hover:text-yellow-400" />
                        )}
                      </button>
                      <button className="text-gray-400 hover:text-gray-600">
                        <EllipsisVerticalIcon className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
                <table className="min-w-full">
                  <thead className="bg-gray-50 border-b">
                    <tr>
                      <th className="px-4 py-3 text-left">
                        <input
                          type="checkbox"
                          className="h-4 w-4 text-red-600 rounded"
                        />
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Name
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Size
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Modified
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {filteredFiles.map((file) => (
                      <tr key={file.id} className="hover:bg-gray-50">
                        <td className="px-4 py-3">
                          <input
                            type="checkbox"
                            checked={selectedItems.has(file.id)}
                            onChange={() => handleSelectItem(file.id)}
                            className="h-4 w-4 text-red-600 rounded"
                          />
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center">
                            <div className="relative">
                              {getFileIcon(file.type)}
                              {!file._isDecrypted && (
                                <LockClosedIcon className="h-3 w-3 absolute -top-1 -right-1 text-red-500" />
                              )}
                            </div>
                            <div className="ml-3">
                              <span className="font-medium text-gray-900">
                                {file.name || "[Encrypted]"}
                              </span>
                              {file.starred && (
                                <StarIconSolid className="h-4 w-4 text-yellow-400 ml-2 inline" />
                              )}
                              {file._decryptionError && (
                                <div className="text-xs text-red-500 mt-1">
                                  {file._decryptionError}
                                </div>
                              )}
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-sm text-gray-500">
                          {file.size}
                        </td>
                        <td className="px-4 py-3 text-sm text-gray-500">
                          {file.modified}
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center space-x-2">
                            <button
                              className="text-gray-400 hover:text-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
                              disabled={
                                !file._isDecrypted ||
                                downloadingFiles.has(file.id)
                              }
                              onClick={() =>
                                handleDownloadFile(file.id, file.name)
                              }
                            >
                              {downloadingFiles.has(file.id) ? (
                                <div className="animate-spin rounded-full h-4 w-4 border-b border-gray-400"></div>
                              ) : (
                                <ArrowDownTrayIcon className="h-4 w-4" />
                              )}
                            </button>
                            <button className="text-gray-400 hover:text-gray-600">
                              <ShareIcon className="h-4 w-4" />
                            </button>
                            <button className="text-gray-400 hover:text-red-600">
                              <TrashIcon className="h-4 w-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* Selection Actions Bar */}
        {selectedItems.size > 0 && (
          <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 bg-gray-900 text-white rounded-lg shadow-lg px-6 py-3 flex items-center space-x-4">
            <span className="text-sm font-medium">
              {selectedItems.size} selected
            </span>
            <div className="h-4 w-px bg-gray-600"></div>
            <button className="text-sm hover:text-red-400">Download</button>
            <button className="text-sm hover:text-red-400">Share</button>
            <button className="text-sm hover:text-red-400">Move</button>
            <button className="text-sm hover:text-red-400">Delete</button>
          </div>
        )}
      </div>

      {/* Share Modal */}
      {showShareModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">
                Share Collection
              </h3>
              <button
                onClick={() => setShowShareModal(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Add people
                </label>
                <input
                  type="email"
                  placeholder="Enter email addresses..."
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Permission
                </label>
                <select className="w-full px-3 py-2 border border-gray-300 rounded-lg">
                  <option>Can view</option>
                  <option>Can edit</option>
                </select>
              </div>

              <div className="pt-4 flex justify-end space-x-3">
                <button
                  onClick={() => setShowShareModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button className="px-4 py-2 bg-gradient-to-r from-red-800 to-red-900 text-white rounded-lg hover:from-red-900 hover:to-red-950">
                  Share
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(CollectionDetails);
