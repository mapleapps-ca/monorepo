// File: src/pages/User/FileManager/FileManagerIndex.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useFiles, useCrypto, useAuth } from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";
import Navigation from "../../../components/Navigation";
import {
  FolderIcon,
  FolderOpenIcon,
  PhotoIcon,
  DocumentIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  AdjustmentsHorizontalIcon,
  ArrowUpTrayIcon,
  EllipsisVerticalIcon,
  ShareIcon,
  PencilIcon,
  TrashIcon,
  ArrowDownTrayIcon,
  ViewColumnsIcon,
  Squares2X2Icon,
  ListBulletIcon,
  ChevronRightIcon,
  HomeIcon,
  StarIcon,
  ClockIcon,
  CloudArrowUpIcon,
  ChevronDownIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
  InformationCircleIcon,
  LockClosedIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import { StarIcon as StarIconSolid } from "@heroicons/react/24/solid";

const FileManagerIndex = () => {
  const navigate = useNavigate();

  // Get services from context
  const { listCollectionManager, getCollectionManager } = useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [collections, setCollections] = useState([]);
  const [viewMode, setViewMode] = useState("grid"); // grid, list, columns
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [showDropdown, setShowDropdown] = useState(null);
  const [sortBy, setSortBy] = useState("name");
  const [filterType, setFilterType] = useState("all");
  const [searchQuery, setSearchQuery] = useState("");

  // Load collections with proper decryption
  const loadCollections = useCallback(
    async (forceRefresh = false) => {
      if (!listCollectionManager) return;

      setIsLoading(true);
      setError("");

      try {
        console.log("[FileManager] === Loading Collections ===");
        console.log("[FileManager] Force refresh:", forceRefresh);

        // Step 1: Get collections from ListCollectionManager
        const result =
          await listCollectionManager.listCollections(forceRefresh);

        console.log("[FileManager] Raw collections loaded:", {
          count: result.collections?.length || 0,
          hasResults: !!result.collections,
        });

        if (!result.collections) {
          setCollections([]);
          return;
        }

        // Step 2: Process collections for proper decryption
        const processedCollections = await processCollections(
          result.collections,
        );

        setCollections(processedCollections);
        console.log(
          "[FileManager] Collections processed successfully:",
          processedCollections.length,
        );
      } catch (err) {
        console.error("[FileManager] Failed to load collections:", err);
        setError(err.message);
        setCollections([]);
      } finally {
        setIsLoading(false);
      }
    },
    [listCollectionManager, getCollectionManager, CollectionCryptoService],
  );

  // Process collections for decryption
  const processCollections = async (rawCollections) => {
    console.log("[FileManager] === Processing Collections for Decryption ===");

    if (!Array.isArray(rawCollections) || rawCollections.length === 0) {
      return [];
    }

    const processedCollections = [];

    for (const collection of rawCollections) {
      try {
        let processedCollection = { ...collection };

        // Check if collection needs decryption
        if (collection.encrypted_name || collection.encrypted_description) {
          console.log("[FileManager] Decrypting collection:", collection.id);

          // Try to get collection key from cache first
          let collectionKey = CollectionCryptoService.getCachedCollectionKey(
            collection.id,
          );

          // If not cached, load the full collection to get the key
          if (!collectionKey && getCollectionManager) {
            try {
              const fullCollection = await getCollectionManager.getCollection(
                collection.id,
              );
              if (fullCollection.collection_key) {
                CollectionCryptoService.cacheCollectionKey(
                  collection.id,
                  fullCollection.collection_key,
                );
                collectionKey = fullCollection.collection_key;
              }
            } catch (keyError) {
              console.warn(
                `[FileManager] Could not get collection key for ${collection.id}:`,
                keyError,
              );
            }
          }

          // Decrypt if we have the key
          if (collectionKey) {
            try {
              // Import CollectionCryptoService for decryption
              const { default: CollectionCryptoServiceClass } = await import(
                "../../../services/Crypto/CollectionCryptoService.js"
              );

              const decryptedCollection =
                await CollectionCryptoServiceClass.decryptCollectionFromAPI(
                  collection,
                  collectionKey,
                );

              processedCollection = decryptedCollection;
              console.log(
                `[FileManager] ✅ Collection decrypted: ${decryptedCollection.name}`,
              );
            } catch (decryptError) {
              console.error(
                `[FileManager] Failed to decrypt collection ${collection.id}:`,
                decryptError,
              );
              processedCollection.name = "[Decryption failed]";
              processedCollection.description = "[Decryption failed]";
              processedCollection._isDecrypted = false;
              processedCollection._decryptionError = decryptError.message;
            }
          } else {
            console.warn(
              `[FileManager] No collection key available for: ${collection.id}`,
            );
            processedCollection.name = "[Collection key unavailable]";
            processedCollection.description = "[Collection key unavailable]";
            processedCollection._isDecrypted = false;
            processedCollection._decryptionError =
              "Collection key not available";
          }
        } else {
          // Collection is not encrypted or already decrypted
          processedCollection._isDecrypted = true;
        }

        // Add UI-specific properties
        processedCollection.type = "folder"; // or determine based on collection type
        processedCollection.items = collection.file_count || 0;
        processedCollection.modified =
          collection.updated_at || collection.created_at;
        processedCollection.size = formatCollectionSize(collection.total_size);
        processedCollection.starred = false; // TODO: implement starring functionality

        processedCollections.push(processedCollection);
      } catch (error) {
        console.error(
          `[FileManager] Failed to process collection ${collection.id}:`,
          error,
        );

        // Add error collection to list
        processedCollections.push({
          ...collection,
          name: `[Error: ${error.message.substring(0, 30)}...]`,
          description: "Failed to process collection",
          type: "folder",
          items: 0,
          modified: collection.updated_at || collection.created_at,
          size: "Unknown",
          starred: false,
          _isDecrypted: false,
          _processingError: error.message,
        });
      }
    }

    return processedCollections;
  };

  // Format collection size
  const formatCollectionSize = (bytes) => {
    if (!bytes || bytes === 0) return "0 B";

    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));

    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  // Format time ago
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

  // Load collections when component mounts
  useEffect(() => {
    if (listCollectionManager && authManager?.isAuthenticated()) {
      loadCollections();
    }
  }, [listCollectionManager, authManager, loadCollections]);

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
    await loadCollections(true);
  };

  // Get collection icon
  const getIcon = (type, isOpen = false) => {
    switch (type) {
      case "folder":
        return isOpen ? (
          <FolderOpenIcon className="h-5 w-5" />
        ) : (
          <FolderIcon className="h-5 w-5" />
        );
      case "album":
        return <PhotoIcon className="h-5 w-5" />;
      default:
        return <DocumentIcon className="h-5 w-5" />;
    }
  };

  // Filter and sort collections
  const filteredCollections = collections
    .filter((collection) => {
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        return (
          collection.name?.toLowerCase().includes(query) ||
          collection.description?.toLowerCase().includes(query)
        );
      }
      return true;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case "name":
          return (a.name || "").localeCompare(b.name || "");
        case "modified":
          return new Date(b.modified || 0) - new Date(a.modified || 0);
        case "size":
          return (b.total_size || 0) - (a.total_size || 0);
        default:
          return 0;
      }
    });

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Breadcrumb */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 mb-6">
          <HomeIcon className="h-4 w-4" />
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900">My Files</span>
        </div>

        {/* Header with Actions */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 mb-2">My Files</h1>
            <p className="text-gray-600">
              {isLoading ? "Loading..." : `${collections.length} collections`}
            </p>
          </div>

          <div className="flex items-center space-x-3">
            <button
              onClick={handleRefresh}
              disabled={isLoading}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
            >
              <ArrowPathIcon className="h-4 w-4 mr-2" />
              {isLoading ? "Refreshing..." : "Refresh"}
            </button>
            <button
              onClick={() => navigate("/file-manager/upload")}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
            >
              <ArrowUpTrayIcon className="h-4 w-4 mr-2" />
              Upload
            </button>
            <button
              onClick={() => navigate("/file-manager/collections/create")}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
            >
              <PlusIcon className="h-4 w-4 mr-2" />
              New Collection
            </button>
          </div>
        </div>

        {/* Search and Filter Bar */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-4 lg:space-y-0">
            {/* Search */}
            <div className="flex-1 max-w-lg">
              <div className="relative">
                <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search collections..."
                  className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                />
              </div>
            </div>

            {/* Filters and View Options */}
            <div className="flex items-center space-x-3">
              {/* Sort Dropdown */}
              <div className="relative">
                <select
                  value={sortBy}
                  onChange={(e) => setSortBy(e.target.value)}
                  className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:ring-2 focus:ring-red-500 focus:border-red-500"
                >
                  <option value="name">Sort by Name</option>
                  <option value="modified">Sort by Modified</option>
                  <option value="size">Sort by Size</option>
                </select>
              </div>

              {/* View Mode Toggle */}
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
        </div>

        {/* Loading State */}
        {isLoading && (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-600 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading collections...</p>
            </div>
          </div>
        )}

        {/* Error State */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
            <div className="flex">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
              <div>
                <h3 className="text-sm font-medium text-red-800">
                  Error loading collections
                </h3>
                <p className="text-sm text-red-700 mt-1">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Collections Section */}
        {!isLoading && (
          <div>
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
                All collections are encrypted on your device before storage
              </p>
            </div>

            {/* Grid View */}
            {viewMode === "grid" && (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {filteredCollections.map((collection) => (
                  <div
                    key={collection.id}
                    className="bg-white rounded-xl border border-gray-200 p-6 hover:shadow-lg transition-all duration-200 cursor-pointer group relative"
                    onDoubleClick={() =>
                      navigate(`/file-manager/collections/${collection.id}`)
                    }
                  >
                    {/* Selection Checkbox */}
                    <div className="absolute top-4 left-4">
                      <input
                        type="checkbox"
                        checked={selectedItems.has(collection.id)}
                        onChange={() => handleSelectItem(collection.id)}
                        className="h-4 w-4 text-red-600 rounded border-gray-300 focus:ring-red-500 opacity-0 group-hover:opacity-100 transition-opacity duration-200"
                      />
                    </div>

                    {/* Dropdown Menu */}
                    <div className="absolute top-4 right-4">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setShowDropdown(
                            showDropdown === collection.id
                              ? null
                              : collection.id,
                          );
                        }}
                        className="p-1 rounded hover:bg-gray-100 opacity-0 group-hover:opacity-100 transition-opacity duration-200"
                      >
                        <EllipsisVerticalIcon className="h-5 w-5 text-gray-500" />
                      </button>
                    </div>

                    {/* Collection Icon */}
                    <div
                      className={`inline-flex items-center justify-center h-12 w-12 rounded-lg mb-4 ${
                        collection.type === "album"
                          ? "bg-pink-100 text-pink-600"
                          : "bg-blue-100 text-blue-600"
                      } ${!collection._isDecrypted ? "opacity-50" : ""}`}
                    >
                      {getIcon(collection.type)}
                      {!collection._isDecrypted && (
                        <LockClosedIcon className="h-3 w-3 absolute" />
                      )}
                    </div>

                    {/* Collection Info */}
                    <h3 className="font-semibold text-gray-900 mb-1 truncate">
                      {collection.name || "[Encrypted]"}
                    </h3>
                    {collection._decryptionError && (
                      <p className="text-xs text-red-500 mb-1">
                        {collection._decryptionError}
                      </p>
                    )}
                    <p className="text-sm text-gray-500 mb-2">
                      {collection.items} items
                    </p>
                    <div className="flex items-center justify-between text-xs text-gray-400">
                      <span>{collection.size}</span>
                      <span>{getTimeAgo(collection.modified)}</span>
                    </div>

                    {/* Star Icon */}
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        // TODO: Toggle star
                      }}
                      className="absolute bottom-4 right-4"
                    >
                      {collection.starred ? (
                        <StarIconSolid className="h-4 w-4 text-yellow-400" />
                      ) : (
                        <StarIcon className="h-4 w-4 text-gray-300 hover:text-yellow-400 transition-colors duration-200" />
                      )}
                    </button>
                  </div>
                ))}

                {/* Create New Collection Card */}
                <div
                  onClick={() => navigate("/file-manager/collections/create")}
                  className="bg-white rounded-xl border-2 border-dashed border-gray-300 p-6 hover:border-red-400 transition-all duration-200 cursor-pointer group flex flex-col items-center justify-center min-h-[200px]"
                >
                  <div className="inline-flex items-center justify-center h-12 w-12 rounded-lg bg-gray-100 group-hover:bg-red-100 mb-4 transition-colors duration-200">
                    <PlusIcon className="h-6 w-6 text-gray-400 group-hover:text-red-600 transition-colors duration-200" />
                  </div>
                  <span className="font-medium text-gray-600 group-hover:text-red-600 transition-colors duration-200">
                    Create Collection
                  </span>
                </div>
              </div>
            )}

            {/* List View */}
            {viewMode === "list" && (
              <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
                <table className="min-w-full">
                  <thead className="bg-gray-50 border-b border-gray-200">
                    <tr>
                      <th className="px-6 py-3 text-left">
                        <input
                          type="checkbox"
                          className="h-4 w-4 text-red-600 rounded border-gray-300 focus:ring-red-500"
                        />
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Name
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Type
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Items
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Size
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Modified
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {filteredCollections.map((collection) => (
                      <tr
                        key={collection.id}
                        className="hover:bg-gray-50 cursor-pointer"
                        onClick={() =>
                          navigate(`/file-manager/collections/${collection.id}`)
                        }
                      >
                        <td className="px-6 py-4">
                          <input
                            type="checkbox"
                            checked={selectedItems.has(collection.id)}
                            onChange={() => handleSelectItem(collection.id)}
                            onClick={(e) => e.stopPropagation()}
                            className="h-4 w-4 text-red-600 rounded border-gray-300 focus:ring-red-500"
                          />
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center">
                            <div
                              className={`flex-shrink-0 h-8 w-8 rounded-lg flex items-center justify-center mr-3 ${
                                collection.type === "album"
                                  ? "bg-pink-100 text-pink-600"
                                  : "bg-blue-100 text-blue-600"
                              } ${!collection._isDecrypted ? "opacity-50" : ""}`}
                            >
                              {getIcon(collection.type)}
                              {!collection._isDecrypted && (
                                <LockClosedIcon className="h-3 w-3 absolute" />
                              )}
                            </div>
                            <div className="flex items-center">
                              <span className="font-medium text-gray-900">
                                {collection.name || "[Encrypted]"}
                              </span>
                              {collection.starred && (
                                <StarIconSolid className="h-4 w-4 text-yellow-400 ml-2" />
                              )}
                              {collection._decryptionError && (
                                <ExclamationTriangleIcon
                                  className="h-4 w-4 text-red-500 ml-2"
                                  title={collection._decryptionError}
                                />
                              )}
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500 capitalize">
                          {collection.type}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">
                          {collection.items}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">
                          {collection.size}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">
                          {getTimeAgo(collection.modified)}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">
                          <div className="flex items-center space-x-2">
                            <button
                              className="text-gray-400 hover:text-gray-600"
                              onClick={(e) => e.stopPropagation()}
                            >
                              <ShareIcon className="h-4 w-4" />
                            </button>
                            <button
                              className="text-gray-400 hover:text-gray-600"
                              onClick={(e) => {
                                e.stopPropagation();
                                navigate(
                                  `/file-manager/collections/${collection.id}/edit`,
                                );
                              }}
                            >
                              <PencilIcon className="h-4 w-4" />
                            </button>
                            <button
                              className="text-gray-400 hover:text-red-600"
                              onClick={(e) => e.stopPropagation()}
                            >
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

            {/* Empty State */}
            {!isLoading && filteredCollections.length === 0 && (
              <div className="text-center py-12">
                <FolderIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
                <h3 className="text-sm font-medium text-gray-900 mb-2">
                  {searchQuery ? "No collections found" : "No collections yet"}
                </h3>
                <p className="text-sm text-gray-500 mb-4">
                  {searchQuery
                    ? `No collections match "${searchQuery}"`
                    : "Create your first collection to organize your files"}
                </p>
                {!searchQuery && (
                  <button
                    onClick={() => navigate("/file-manager/collections/create")}
                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
                  >
                    <PlusIcon className="h-4 w-4 mr-2" />
                    Create Collection
                  </button>
                )}
              </div>
            )}
          </div>
        )}

        {/* Selection Actions Bar */}
        {selectedItems.size > 0 && (
          <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 bg-gray-900 text-white rounded-lg shadow-lg px-6 py-3 flex items-center space-x-4 animate-fade-in-up">
            <span className="text-sm font-medium">
              {selectedItems.size} selected
            </span>
            <div className="h-4 w-px bg-gray-600"></div>
            <button className="text-sm hover:text-red-400 transition-colors duration-200">
              Share
            </button>
            <button className="text-sm hover:text-red-400 transition-colors duration-200">
              Download
            </button>
            <button className="text-sm hover:text-red-400 transition-colors duration-200">
              Move
            </button>
            <button className="text-sm hover:text-red-400 transition-colors duration-200">
              Delete
            </button>
          </div>
        )}
      </div>

      {/* Upload Floating Action Button */}
      <button
        onClick={() => navigate("/file-manager/upload")}
        className="fixed bottom-6 right-6 inline-flex items-center justify-center h-14 w-14 bg-gradient-to-r from-red-800 to-red-900 text-white rounded-full shadow-lg hover:shadow-xl transform hover:scale-110 transition-all duration-200"
      >
        <CloudArrowUpIcon className="h-6 w-6" />
      </button>

      <style jsx>{`
        @keyframes fade-in-up {
          from {
            opacity: 0;
            transform: translate(-50%, 20px);
          }
          to {
            opacity: 1;
            transform: translate(-50%, 0);
          }
        }
        .animate-fade-in-up {
          animation: fade-in-up 0.3s ease-out;
        }
      `}</style>
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(FileManagerIndex);
