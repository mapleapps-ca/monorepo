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
  SparklesIcon,
  ChevronRightIcon,
  LockClosedIcon,
  ShareIcon,
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
  const [filterType, setFilterType] = useState("owned");
  const [showFilterMenu, setShowFilterMenu] = useState(false);
  const [hoveredCollection, setHoveredCollection] = useState(null);
  const [collectionCounts, setCollectionCounts] = useState({
    owned: 0,
    shared: 0,
    all: 0,
  });

  const processCollections = useCallback(
    async (rawCollections) => {
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
    },
    [CollectionCryptoService, getCollectionManager, authManager],
  );

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
    [listCollectionManager, filterType, processCollections],
  );

  useEffect(() => {
    const fetchCounts = async () => {
      if (!listCollectionManager) return;
      try {
        const result = await listCollectionManager.listFilteredCollections(
          true,
          true,
          false,
        );
        const owned =
          result.owned_collections?.filter(
            (c) =>
              !c.parent_id ||
              c.parent_id === "00000000-0000-0000-0000-000000000000",
          ) || [];
        const shared =
          result.shared_collections?.filter(
            (c) =>
              !c.parent_id ||
              c.parent_id === "00000000-0000-0000-0000-000000000000",
          ) || [];
        setCollectionCounts({
          owned: owned.length,
          shared: shared.length,
          all: owned.length + shared.length,
        });
      } catch (err) {
        console.error(
          "[FileManagerIndex] Failed to fetch collection counts:",
          err,
        );
      }
    };
    if (authManager?.isAuthenticated()) {
      fetchCounts();
    }
  }, [listCollectionManager, authManager]);

  const handleFilterChange = (newFilter) => {
    console.log("[FileManagerIndex] Filter changed to:", newFilter);
    setFilterType(newFilter);
    setShowFilterMenu(false);
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

  useEffect(() => {
    if (listCollectionManager && authManager?.isAuthenticated()) {
      loadCollections();
    }
  }, [listCollectionManager, authManager, loadCollections]);

  useEffect(() => {
    const handleCollectionEvent = () => {
      console.log(
        "[FileManagerIndex] Collection event received, refreshing collections",
      );
      if (listCollectionManager && authManager?.isAuthenticated()) {
        loadCollections(true);
      }
    };
    window.addEventListener("collectionCreated", handleCollectionEvent);
    window.addEventListener("rootCollectionCreated", handleCollectionEvent);
    return () => {
      window.removeEventListener("collectionCreated", handleCollectionEvent);
      window.removeEventListener(
        "rootCollectionCreated",
        handleCollectionEvent,
      );
    };
  }, [listCollectionManager, authManager, loadCollections]);

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
      return (collection.name || "Locked Folder").toLowerCase().includes(query);
    }
    return true;
  });

  const filterTypes = [
    {
      key: "owned",
      label: "My Folders",
      icon: UserIcon,
      description: "Folders you own",
      count: collectionCounts.owned,
    },
    {
      key: "shared",
      label: "Shared with Me",
      icon: UsersIcon,
      description: "Folders shared with you",
      count: collectionCounts.shared,
    },
    {
      key: "all",
      label: "All Folders",
      icon: RectangleStackIcon,
      description: "All folders you can access",
      count: collectionCounts.all,
    },
  ];

  const currentFilter = filterTypes.find((f) => f.key === filterType);

  return (
    <div className="min-h-screen bg-gradient-subtle">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8 animate-fade-in-down">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 flex items-center">
                My Files
                <SparklesIcon className="h-8 w-8 text-yellow-500 ml-2" />
              </h1>
              <p className="text-gray-600 mt-1">
                Organize and manage your encrypted files
              </p>
            </div>

            <div className="flex items-center space-x-3">
              {/* Filter Menu */}
              <div className="relative">
                <button
                  onClick={() => setShowFilterMenu(!showFilterMenu)}
                  className="btn-secondary flex items-center"
                >
                  <FunnelIcon className="h-4 w-4 mr-2" />
                  {currentFilter.label}
                  <ChevronRightIcon
                    className={`h-4 w-4 ml-2 transition-transform duration-200 ${
                      showFilterMenu ? "rotate-90" : ""
                    }`}
                  />
                </button>

                {showFilterMenu && (
                  <>
                    <div
                      className="fixed inset-0 z-10"
                      onClick={() => setShowFilterMenu(false)}
                    ></div>
                    <div className="absolute right-0 mt-2 w-64 bg-white rounded-xl shadow-xl border border-gray-200 py-2 z-20 animate-fade-in-down">
                      {filterTypes.map((filter) => (
                        <button
                          key={filter.key}
                          onClick={() => handleFilterChange(filter.key)}
                          className={`w-full text-left px-4 py-3 hover:bg-gray-50 transition-colors duration-200 ${
                            filterType === filter.key ? "bg-red-50" : ""
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex items-center space-x-3">
                              <filter.icon
                                className={`h-5 w-5 ${
                                  filterType === filter.key
                                    ? "text-red-700"
                                    : "text-gray-500"
                                }`}
                              />
                              <div>
                                <p
                                  className={`text-sm font-medium ${
                                    filterType === filter.key
                                      ? "text-red-900"
                                      : "text-gray-900"
                                  }`}
                                >
                                  {filter.label}
                                </p>
                                <p className="text-xs text-gray-500">
                                  {filter.description}
                                </p>
                              </div>
                            </div>
                            <span
                              className={`text-sm font-medium ${
                                filterType === filter.key
                                  ? "text-red-700"
                                  : "text-gray-500"
                              }`}
                            >
                              {filter.count}
                            </span>
                          </div>
                        </button>
                      ))}
                    </div>
                  </>
                )}
              </div>

              <button
                onClick={() => loadCollections(true)}
                disabled={isLoading}
                className="btn-secondary"
              >
                <ArrowPathIcon
                  className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
                />
                Refresh
              </button>

              <button
                onClick={() => navigate("/file-manager/upload")}
                className="btn-primary"
              >
                <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                Upload
              </button>
            </div>
          </div>
        </div>

        {/* Search Bar */}
        <div className="mb-8 animate-fade-in-up">
          <div className="relative max-w-2xl">
            <MagnifyingGlassIcon className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search folders..."
              className="input pl-12 h-12 text-base"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery("")}
                className="absolute right-4 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
              >
                ✕
              </button>
            )}
          </div>
        </div>

        {/* Collections Grid */}
        {isLoading ? (
          <div className="flex items-center justify-center py-24">
            <div className="text-center">
              <div className="h-12 w-12 spinner border-red-700 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading folders...</p>
            </div>
          </div>
        ) : error ? (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-center">
            <p className="text-red-700">{error}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredCollections.map((collection, index) => (
              <div
                key={collection.id}
                className="card card-hover cursor-pointer animate-fade-in-up group"
                style={{ animationDelay: `${index * 50}ms` }}
                onClick={() =>
                  navigate(`/file-manager/collections/${collection.id}`)
                }
                onMouseEnter={() => setHoveredCollection(collection.id)}
                onMouseLeave={() => setHoveredCollection(null)}
              >
                <div className="p-6">
                  <div className="mb-4">
                    <div
                      className={`h-14 w-14 rounded-xl flex items-center justify-center transition-all duration-200 ${
                        collection.type === "album"
                          ? "bg-gradient-to-br from-pink-500 to-pink-600 group-hover:from-pink-600 group-hover:to-pink-700"
                          : "bg-gradient-to-br from-blue-500 to-blue-600 group-hover:from-blue-600 group-hover:to-blue-700"
                      }`}
                    >
                      {collection.type === "album" ? (
                        <PhotoIcon className="h-7 w-7 text-white" />
                      ) : (
                        <FolderIcon className="h-7 w-7 text-white" />
                      )}
                    </div>
                  </div>

                  <div className="space-y-3">
                    <div>
                      <h3 className="font-semibold text-gray-900 text-lg flex items-center justify-between">
                        <span className="truncate">
                          {collection.name || "Locked Folder"}
                        </span>
                        {!collection._isDecrypted && (
                          <LockClosedIcon className="h-4 w-4 text-gray-400 flex-shrink-0 ml-2" />
                        )}
                      </h3>
                      <p className="text-sm text-gray-600 mt-1">
                        {collection.fileCount > 0
                          ? `${collection.fileCount} files`
                          : "Empty folder"}
                      </p>
                    </div>

                    <div className="flex items-center justify-between pt-3 border-t border-gray-100">
                      <span className="text-xs text-gray-500">
                        {getTimeAgo(collection.modified)}
                      </span>
                      <div className="flex items-center space-x-2">
                        {collection.isShared && (
                          <div
                            className={`flex items-center space-x-1 px-2 py-1 rounded-full text-xs font-medium ${
                              collection.isOwned
                                ? "bg-green-100 text-green-700"
                                : "bg-blue-100 text-blue-700"
                            }`}
                          >
                            <ShareIcon className="h-3 w-3" />
                            <span>
                              {collection.isOwned ? "Shared" : "With me"}
                            </span>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>

                <div
                  className={`absolute inset-0 rounded-xl border-2 border-transparent transition-all duration-200 pointer-events-none ${
                    hoveredCollection === collection.id ? "border-red-200" : ""
                  }`}
                ></div>
              </div>
            ))}

            {(filterType === "owned" || filterType === "all") && (
              <div
                onClick={() => navigate("/file-manager/collections/create")}
                className="group card border-2 border-dashed border-gray-300 hover:border-red-400 cursor-pointer transition-all duration-300 animate-fade-in-up"
                style={{
                  animationDelay: `${filteredCollections.length * 50}ms`,
                }}
              >
                <div className="p-6 text-center">
                  <div className="h-14 w-14 rounded-xl bg-gray-100 group-hover:bg-red-50 flex items-center justify-center mx-auto mb-4 transition-colors duration-200">
                    <PlusIcon className="h-7 w-7 text-gray-400 group-hover:text-red-600 transition-colors duration-200" />
                  </div>
                  <h3 className="font-semibold text-gray-700 group-hover:text-gray-900 text-lg transition-colors duration-200">
                    New Folder
                  </h3>
                  <p className="text-sm text-gray-500 mt-1">
                    Create a new encrypted folder
                  </p>
                </div>
              </div>
            )}
          </div>
        )}

        {!isLoading && filteredCollections.length === 0 && !searchQuery && (
          <div className="text-center py-24 animate-fade-in">
            <div className="h-20 w-20 bg-gray-100 rounded-2xl flex items-center justify-center mx-auto mb-6">
              <currentFilter.icon className="h-10 w-10 text-gray-400" />
            </div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              {filterType === "shared"
                ? "No shared folders"
                : filterType === "all"
                  ? "No folders found"
                  : "No folders yet"}
            </h3>
            <p className="text-gray-600 mb-8 max-w-md mx-auto">
              {filterType === "shared"
                ? "When someone shares a folder with you, it will appear here"
                : filterType === "all"
                  ? "Create your first folder or wait for someone to share with you"
                  : "Create your first encrypted folder to start organizing your files"}
            </p>
            {filterType !== "shared" && (
              <button
                onClick={() => navigate("/file-manager/collections/create")}
                className="btn-primary"
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                Create Your First Folder
              </button>
            )}
          </div>
        )}

        {!isLoading && filteredCollections.length === 0 && searchQuery && (
          <div className="text-center py-24 animate-fade-in">
            <div className="h-20 w-20 bg-gray-100 rounded-2xl flex items-center justify-center mx-auto mb-6">
              <MagnifyingGlassIcon className="h-10 w-10 text-gray-400" />
            </div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              No folders found
            </h3>
            <p className="text-gray-600 mb-8">
              No folders match your search for "{searchQuery}"
            </p>
            <button
              onClick={() => setSearchQuery("")}
              className="text-red-700 hover:text-red-800 font-medium"
            >
              Clear search
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default withPasswordProtection(FileManagerIndex);
