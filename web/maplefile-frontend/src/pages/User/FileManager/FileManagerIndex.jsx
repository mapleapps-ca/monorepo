// File: src/pages/User/FileManager/FileManagerIndex.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useLocation } from "react-router";
import { useFiles, useCrypto, useAuth } from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";
import Navigation from "../../../components/Navigation";
import {
  FolderIcon,
  PhotoIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  CloudArrowUpIcon,
  ArrowPathIcon,
  FunnelIcon,
  UserIcon,
  UsersIcon,
  RectangleStackIcon,
} from "@heroicons/react/24/outline";

const FileManagerIndex = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { listCollectionManager, getCollectionManager } = useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [collections, setCollections] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");

  // NEW: Filter state
  const [filterType, setFilterType] = useState("owned"); // 'owned', 'shared', 'all'
  const [showFilterMenu, setShowFilterMenu] = useState(false);

  const loadCollections = useCallback(
    async (forceRefresh = false, currentFilterType = null) => {
      if (!listCollectionManager) return;

      setIsLoading(true);
      setError("");

      try {
        const activeFilterType = currentFilterType || filterType;
        let result;

        console.log(
          "[FileManagerIndex] Loading collections with filter:",
          activeFilterType,
        );

        // Use appropriate API call based on filter type
        switch (activeFilterType) {
          case "owned":
            result = await listCollectionManager.listCollections(forceRefresh);
            break;
          case "shared":
            result =
              await listCollectionManager.listSharedCollections(forceRefresh);
            break;
          case "all":
            result = await listCollectionManager.listFilteredCollections(
              true, // includeOwned
              true, // includeShared
              forceRefresh,
            );
            // For filtered collections, we need to combine the arrays
            if (result.owned_collections || result.shared_collections) {
              result.collections = [
                ...(result.owned_collections || []),
                ...(result.shared_collections || []),
              ];
            }
            break;
          default:
            result = await listCollectionManager.listCollections(forceRefresh);
        }

        if (result.collections) {
          // Filter to only show root-level collections (no parent)
          const rootCollections = result.collections.filter(
            (collection) =>
              !collection.parent_id ||
              collection.parent_id === "00000000-0000-0000-0000-000000000000",
          );

          console.log(
            `[FileManagerIndex] Found ${rootCollections.length} root collections for filter: ${activeFilterType}`,
          );

          const processedCollections =
            await processCollections(rootCollections);
          setCollections(processedCollections);
        }
      } catch (err) {
        console.error("[FileManagerIndex] Failed to load collections:", err);
        setError("Could not load your folders. Please try again.");
      } finally {
        setIsLoading(false);
      }
    },
    [
      listCollectionManager,
      getCollectionManager,
      CollectionCryptoService,
      filterType,
    ],
  );

  const processCollections = async (rawCollections) => {
    if (!Array.isArray(rawCollections) || rawCollections.length === 0)
      return [];

    const processedCollections = [];

    for (const collection of rawCollections) {
      try {
        let processedCollection = { ...collection };

        if (collection.encrypted_name || collection.encrypted_description) {
          let collectionKey = CollectionCryptoService.getCachedCollectionKey(
            collection.id,
          );

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
                `Could not get collection key for ${collection.id}:`,
                keyError,
              );
            }
          }

          if (collectionKey) {
            try {
              const { default: CollectionCryptoServiceClass } = await import(
                "../../../services/Crypto/CollectionCryptoService.js"
              );

              const decryptedCollection =
                await CollectionCryptoServiceClass.decryptCollectionFromAPI(
                  collection,
                  collectionKey,
                );
              processedCollection = decryptedCollection;
            } catch (decryptError) {
              processedCollection.name = "Locked Folder";
              processedCollection._isDecrypted = false;
            }
          } else {
            processedCollection.name = "Locked Folder";
            processedCollection._isDecrypted = false;
          }
        } else {
          processedCollection._isDecrypted = true;
        }

        processedCollection.type = collection.collection_type || "folder";
        processedCollection.fileCount = collection.file_count || 0;
        processedCollection.modified =
          collection.updated_at || collection.created_at;

        // NEW: Add sharing information
        processedCollection.isShared = !!(
          collection.members && collection.members.length > 0
        );
        processedCollection.isOwned =
          collection.owner_id === authManager?.getCurrentUserEmail();

        processedCollections.push(processedCollection);
      } catch (error) {
        processedCollections.push({
          ...collection,
          name: "Locked Folder",
          type: "folder",
          fileCount: 0,
          _isDecrypted: false,
          isShared: false,
          isOwned: true,
        });
      }
    }

    return processedCollections;
  };

  // NEW: Handle filter change
  const handleFilterChange = (newFilter) => {
    console.log("[FileManagerIndex] Filter changed to:", newFilter);
    setFilterType(newFilter);
    setShowFilterMenu(false);
    // Load collections with the new filter immediately
    loadCollections(false, newFilter);
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

  // Load collections when dependencies are ready
  useEffect(() => {
    if (listCollectionManager && authManager?.isAuthenticated()) {
      loadCollections();
    }
  }, [listCollectionManager, authManager, loadCollections]);

  // Listen for collection creation events and refresh
  useEffect(() => {
    const handleCollectionCreated = () => {
      console.log(
        "[FileManagerIndex] Collection created event received, refreshing collections",
      );
      if (listCollectionManager && authManager?.isAuthenticated()) {
        loadCollections(true);
      }
    };

    const handleRootCollectionCreated = (event) => {
      console.log(
        "[FileManagerIndex] Root collection created event received:",
        event.detail,
      );
      if (listCollectionManager && authManager?.isAuthenticated()) {
        loadCollections(true);
      }
    };

    window.addEventListener("collectionCreated", handleCollectionCreated);
    window.addEventListener(
      "rootCollectionCreated",
      handleRootCollectionCreated,
    );

    return () => {
      window.removeEventListener("collectionCreated", handleCollectionCreated);
      window.removeEventListener(
        "rootCollectionCreated",
        handleRootCollectionCreated,
      );
    };
  }, [listCollectionManager, authManager, loadCollections]);

  // Handle refresh when returning from collection creation
  useEffect(() => {
    if (location.state?.refresh && location.state?.newRootCollectionCreated) {
      console.log(
        "[FileManagerIndex] Forcing refresh due to new root collection creation",
      );
      if (listCollectionManager && authManager?.isAuthenticated()) {
        loadCollections(true);
      }

      navigate(location.pathname, { replace: true, state: {} });
    }
  }, [
    location.state,
    listCollectionManager,
    authManager,
    loadCollections,
    navigate,
    location.pathname,
  ]);

  const filteredCollections = collections.filter((collection) => {
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      return collection.name?.toLowerCase().includes(query);
    }
    return true;
  });

  // NEW: Filter type configurations
  const filterTypes = [
    {
      key: "owned",
      label: "My Folders",
      icon: UserIcon,
      description: "Folders you own",
    },
    {
      key: "shared",
      label: "Shared with Me",
      icon: UsersIcon,
      description: "Folders shared with you",
    },
    {
      key: "all",
      label: "All Folders",
      icon: RectangleStackIcon,
      description: "All folders you can access",
    },
  ];

  const currentFilter = filterTypes.find((f) => f.key === filterType);

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center space-x-4">
            <h1 className="text-2xl font-semibold text-gray-900">My Files</h1>

            {/* NEW: Filter Badge */}
            <div className="flex items-center text-sm text-gray-500 bg-gray-100 px-3 py-1 rounded-full">
              <currentFilter.icon className="h-4 w-4 mr-1" />
              {currentFilter.label}
            </div>
          </div>

          <div className="flex items-center space-x-3">
            {/* NEW: Filter Menu */}
            <div className="relative">
              <button
                onClick={() => setShowFilterMenu(!showFilterMenu)}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <FunnelIcon className="h-4 w-4 mr-2" />
                Filter
              </button>

              {showFilterMenu && (
                <div className="absolute right-0 mt-2 w-56 bg-white rounded-lg shadow-lg border border-gray-200 z-10">
                  <div className="py-1">
                    {filterTypes.map((filter) => (
                      <button
                        key={filter.key}
                        onClick={() => handleFilterChange(filter.key)}
                        className={`w-full text-left px-4 py-3 text-sm hover:bg-gray-50 flex items-center ${
                          filterType === filter.key
                            ? "bg-red-50 text-red-800"
                            : "text-gray-700"
                        }`}
                      >
                        <filter.icon
                          className={`h-4 w-4 mr-3 ${
                            filterType === filter.key
                              ? "text-red-600"
                              : "text-gray-400"
                          }`}
                        />
                        <div>
                          <div className="font-medium">{filter.label}</div>
                          <div className="text-xs text-gray-500">
                            {filter.description}
                          </div>
                        </div>
                        {filterType === filter.key && (
                          <div className="ml-auto w-2 h-2 bg-red-600 rounded-full"></div>
                        )}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <button
              onClick={() => loadCollections(true)}
              disabled={isLoading}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
            >
              <ArrowPathIcon
                className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
              />
              Refresh
            </button>

            <button
              onClick={() => navigate("/file-manager/upload")}
              className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900"
            >
              <CloudArrowUpIcon className="h-4 w-4 mr-2" />
              Upload
            </button>
          </div>
        </div>

        {/* Search and Results Info */}
        <div className="mb-6">
          <div className="flex items-center justify-between">
            <div className="relative max-w-md">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search folders..."
                className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
              />
            </div>

            {/* Results Count */}
            {!isLoading && (
              <div className="text-sm text-gray-500">
                {filteredCollections.length === 1
                  ? "1 folder"
                  : `${filteredCollections.length} folders`}
                {searchQuery && ` matching "${searchQuery}"`}
              </div>
            )}
          </div>
        </div>

        {/* Collections Grid */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading folders...</p>
            </div>
          </div>
        ) : error ? (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <p className="text-red-700">{error}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {filteredCollections.map((collection) => (
              <div
                key={collection.id}
                className="bg-white rounded-lg border border-gray-200 p-6 hover:shadow-md hover:border-gray-300 cursor-pointer transition-all"
                onClick={() =>
                  navigate(`/file-manager/collections/${collection.id}`)
                }
              >
                <div className="flex items-start space-x-4">
                  <div
                    className={`flex items-center justify-center h-12 w-12 rounded-lg ${
                      collection.type === "album"
                        ? "bg-pink-100 text-pink-600"
                        : "bg-blue-100 text-blue-600"
                    }`}
                  >
                    {collection.type === "album" ? (
                      <PhotoIcon className="h-6 w-6" />
                    ) : (
                      <FolderIcon className="h-6 w-6" />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2">
                      <h3 className="font-semibold text-gray-900 truncate">
                        {collection.name || "Locked Folder"}
                      </h3>
                      {/* NEW: Sharing indicator */}
                      {collection.isShared && !collection.isOwned && (
                        <UsersIcon
                          className="h-4 w-4 text-blue-500 flex-shrink-0"
                          title="Shared with you"
                        />
                      )}
                      {collection.isShared && collection.isOwned && (
                        <UsersIcon
                          className="h-4 w-4 text-green-500 flex-shrink-0"
                          title="Shared by you"
                        />
                      )}
                    </div>
                    <p className="text-sm text-gray-500">
                      {collection.fileCount > 0
                        ? `${collection.fileCount} files`
                        : "Empty"}
                    </p>
                    <p className="text-xs text-gray-400 mt-1">
                      {getTimeAgo(collection.modified)}
                    </p>
                  </div>
                </div>
              </div>
            ))}

            {/* Create New Collection Card - Only show for owned/all filters */}
            {(filterType === "owned" || filterType === "all") && (
              <div
                onClick={() => navigate("/file-manager/collections/create")}
                className="bg-white rounded-lg border-2 border-dashed border-gray-300 p-6 hover:border-red-400 cursor-pointer transition-all flex flex-col items-center justify-center"
              >
                <div className="h-12 w-12 rounded-lg bg-gray-100 flex items-center justify-center mb-3">
                  <PlusIcon className="h-6 w-6 text-gray-400" />
                </div>
                <h3 className="font-semibold text-gray-700">New Folder</h3>
                <p className="text-sm text-gray-500 mt-1">Create a folder</p>
              </div>
            )}
          </div>
        )}

        {/* Empty State */}
        {!isLoading && filteredCollections.length === 0 && !searchQuery && (
          <div className="text-center py-12">
            <currentFilter.icon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {filterType === "shared"
                ? "No shared folders"
                : filterType === "all"
                  ? "No folders found"
                  : "No folders yet"}
            </h3>
            <p className="text-gray-500 mb-4">
              {filterType === "shared"
                ? "No one has shared any folders with you yet"
                : filterType === "all"
                  ? "You don't have access to any folders yet"
                  : "Create your first folder to organize files"}
            </p>
            {filterType !== "shared" && (
              <button
                onClick={() => navigate("/file-manager/collections/create")}
                className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900"
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                Create Folder
              </button>
            )}
          </div>
        )}

        {/* No Search Results */}
        {!isLoading && filteredCollections.length === 0 && searchQuery && (
          <div className="text-center py-12">
            <MagnifyingGlassIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              No folders found
            </h3>
            <p className="text-gray-500 mb-4">
              No folders match your search for "{searchQuery}"
            </p>
            <button
              onClick={() => setSearchQuery("")}
              className="text-red-600 hover:text-red-800 font-medium"
            >
              Clear search
            </button>
          </div>
        )}

        {/* Click outside to close filter menu */}
        {showFilterMenu && (
          <div
            className="fixed inset-0 z-0"
            onClick={() => setShowFilterMenu(false)}
          />
        )}
      </div>
    </div>
  );
};

export default withPasswordProtection(FileManagerIndex);
