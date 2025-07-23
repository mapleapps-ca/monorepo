// File: src/pages/User/Collection/List.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useCollections, useAuth } from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";
import Navigation from "../../../components/Navigation";
import {
  FolderIcon,
  FolderOpenIcon,
  PhotoIcon,
  MagnifyingGlassIcon,
  ArrowPathIcon,
  PlusIcon,
  ShareIcon,
  TrashIcon,
  EyeIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  CheckCircleIcon,
  ClockIcon,
  UserIcon,
  UsersIcon,
  HomeIcon,
  FunnelIcon,
  ChevronDownIcon,
  DocumentIcon,
  LockClosedIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";

const CollectionList = () => {
  const navigate = useNavigate();

  // Get services from unified service architecture
  const { listCollectionManager } = useCollections();
  const { authManager, user } = useAuth();

  // Local component state - managed by component, not hook
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [collections, setCollections] = useState([]);
  const [sharedCollections, setSharedCollections] = useState([]);
  const [filteredCollections, setFilteredCollections] = useState({
    owned_collections: [],
    shared_collections: [],
    total_count: 0,
  });
  const [rootCollections, setRootCollections] = useState([]);
  const [collectionsByParent, setCollectionsByParent] = useState([]);

  // UI state
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedListType, setSelectedListType] = useState("user");
  const [parentId, setParentId] = useState("");
  const [includeOwned, setIncludeOwned] = useState(true);
  const [includeShared, setIncludeShared] = useState(false);
  const [showDetails, setShowDetails] = useState({});
  const [showFilters, setShowFilters] = useState(false);

  // Computed properties from authManager
  const isAuthenticated = authManager?.isAuthenticated() || false;
  const canListCollections =
    authManager?.canMakeAuthenticatedRequests() || false;

  // Total counts
  const totalCollections = collections.length;
  const totalSharedCollections = sharedCollections.length;
  const totalFilteredCollections = filteredCollections.total_count;
  const totalRootCollections = rootCollections.length;

  // Add event to log (simplified for production)
  const addToEventLog = (eventType, eventData) => {
    console.log(`[CollectionList] ${eventType}:`, eventData);
  };

  // Handle list collections
  const handleListCollections = useCallback(
    async (forceRefresh = false) => {
      console.log("[CollectionList] Listing collections...");
      console.log("[CollectionList] List type:", selectedListType);
      console.log("[CollectionList] Force refresh:", forceRefresh);

      if (!listCollectionManager) {
        setError(
          "Collection service is not available. Please refresh the page.",
        );
        return;
      }

      if (!isAuthenticated) {
        setError("You must be authenticated to view collections.");
        return;
      }

      setIsLoading(true);
      setError(null);
      setSuccess(null);

      try {
        let result;

        switch (selectedListType) {
          case "user":
            console.log("[CollectionList] Listing user collections");
            result = await listCollectionManager.listCollections(forceRefresh);
            setCollections(result.collections || []);
            addToEventLog("user_collections_listed", {
              totalCount: result.totalCount,
              source: result.source,
              forceRefresh,
            });
            setSuccess(`Found ${result.totalCount} collections`);
            break;

          case "shared":
            console.log("[CollectionList] Listing shared collections");
            result =
              await listCollectionManager.listSharedCollections(forceRefresh);
            setSharedCollections(result.collections || []);
            addToEventLog("shared_collections_listed", {
              totalCount: result.totalCount,
              source: result.source,
              forceRefresh,
            });
            setSuccess(`Found ${result.totalCount} shared collections`);
            break;

          case "filtered":
            console.log("[CollectionList] Listing filtered collections:", {
              includeOwned,
              includeShared,
            });
            result = await listCollectionManager.listFilteredCollections(
              includeOwned,
              includeShared,
              forceRefresh,
            );
            setFilteredCollections({
              owned_collections: result.owned_collections || [],
              shared_collections: result.shared_collections || [],
              total_count: result.total_count || 0,
            });
            addToEventLog("filtered_collections_listed", {
              ownedCount: result.owned_collections?.length || 0,
              sharedCount: result.shared_collections?.length || 0,
              totalCount: result.total_count,
              source: result.source,
              forceRefresh,
            });
            setSuccess(`Found ${result.total_count} filtered collections`);
            break;

          case "root":
            console.log("[CollectionList] Listing root collections");
            result =
              await listCollectionManager.listRootCollections(forceRefresh);
            setRootCollections(result.collections || []);
            addToEventLog("root_collections_listed", {
              totalCount: result.totalCount,
              source: result.source,
              forceRefresh,
            });
            setSuccess(`Found ${result.totalCount} root collections`);
            break;

          case "byParent":
            if (!parentId.trim()) {
              setError("Parent ID is required for listing by parent");
              return;
            }
            console.log(
              "[CollectionList] Listing collections by parent:",
              parentId,
            );
            result = await listCollectionManager.listCollectionsByParent(
              parentId.trim(),
              forceRefresh,
            );
            setCollectionsByParent(result.collections || []);
            addToEventLog("collections_by_parent_listed", {
              parentId: parentId.trim(),
              totalCount: result.totalCount,
              source: result.source,
              forceRefresh,
            });
            setSuccess(`Found ${result.totalCount} collections in parent`);
            break;

          default:
            throw new Error(`Unknown list type: ${selectedListType}`);
        }

        console.log("[CollectionList] Listing completed successfully:", result);
      } catch (err) {
        console.error("[CollectionList] Collection listing failed:", err);
        setError(`Failed to load collections: ${err.message}`);
        addToEventLog("listing_failed", {
          listType: selectedListType,
          error: err.message,
          forceRefresh,
        });
      } finally {
        setIsLoading(false);
      }
    },
    [
      listCollectionManager,
      isAuthenticated,
      selectedListType,
      includeOwned,
      includeShared,
      parentId,
    ],
  );

  // Handle cache operations
  const handleClearAllCache = async () => {
    if (!window.confirm("Clear all collection cache? This cannot be undone."))
      return;

    try {
      await listCollectionManager.clearAllCache();
      setSuccess("All cache cleared successfully");
    } catch (err) {
      console.error("Failed to clear all cache:", err);
      setError(`Failed to clear cache: ${err.message}`);
    }
  };

  // Clear messages
  const clearMessages = () => {
    setError(null);
    setSuccess(null);
  };

  // Toggle collection details
  const toggleDetails = (collectionId) => {
    setShowDetails((prev) => ({
      ...prev,
      [collectionId]: !prev[collectionId],
    }));
  };

  // Get current collections based on selected list type
  const getCurrentCollections = () => {
    switch (selectedListType) {
      case "user":
        return collections;
      case "shared":
        return sharedCollections;
      case "filtered":
        return [
          ...filteredCollections.owned_collections,
          ...filteredCollections.shared_collections,
        ];
      case "root":
        return rootCollections;
      case "byParent":
        return collectionsByParent;
      default:
        return [];
    }
  };

  const currentCollections = getCurrentCollections();

  // Search filtered collections
  const searchResults = searchTerm
    ? listCollectionManager?.searchCollections?.(
        searchTerm,
        currentCollections,
      ) ||
      currentCollections.filter(
        (c) =>
          c.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
          c.id?.toLowerCase().includes(searchTerm.toLowerCase()),
      )
    : currentCollections;

  // Filter results by type
  const folders = searchResults.filter((c) => c.collection_type === "folder");
  const albums = searchResults.filter((c) => c.collection_type === "album");

  // Auto-clear messages after 5 seconds
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        clearMessages();
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [success, error]);

  // Load collections on mount
  useEffect(() => {
    if (listCollectionManager && isAuthenticated) {
      handleListCollections();
    }
  }, [listCollectionManager, isAuthenticated, handleListCollections]);

  // Get collection icon
  const getCollectionIcon = (collection) => {
    if (collection.collection_type === "album") {
      return <PhotoIcon className="h-5 w-5" />;
    }
    return <FolderIcon className="h-5 w-5" />;
  };

  // Get collection type label
  const getTypeLabel = (type) => {
    switch (type) {
      case "user":
        return "My Collections";
      case "shared":
        return "Shared with Me";
      case "filtered":
        return "Filtered View";
      case "root":
        return "Root Collections";
      case "byParent":
        return "Child Collections";
      default:
        return "Collections";
    }
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "Unknown";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      {/* Navigation */}
      <Navigation />

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-3xl font-black text-gray-900 mb-2">
                {getTypeLabel(selectedListType)}
              </h1>
              <p className="text-gray-600">
                Manage your encrypted collections and folders
              </p>
            </div>
            <div className="flex items-center space-x-3">
              <button
                onClick={() => setShowFilters(!showFilters)}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
              >
                <FunnelIcon className="h-4 w-4 mr-2" />
                Filters
                <ChevronDownIcon
                  className={`h-4 w-4 ml-2 transform transition-transform duration-200 ${showFilters ? "rotate-180" : ""}`}
                />
              </button>
              <button
                onClick={() => handleListCollections(true)}
                disabled={isLoading}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
              >
                <ArrowPathIcon
                  className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
                />
                {isLoading ? "Refreshing..." : "Refresh"}
              </button>
              <button
                onClick={() =>
                  navigate("/developer/create-collection-manager-example")
                }
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transform hover:scale-105 transition-all duration-200"
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                New Collection
              </button>
            </div>
          </div>

          {/* Success/Error Messages */}
          {success && (
            <div className="mb-6 p-4 rounded-lg bg-green-50 border border-green-200 animate-fade-in">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <CheckCircleIcon className="h-5 w-5 text-green-500 mr-3" />
                  <span className="text-sm font-medium text-green-800">
                    {success}
                  </span>
                </div>
                <button
                  onClick={clearMessages}
                  className="text-green-500 hover:text-green-700"
                >
                  ×
                </button>
              </div>
            </div>
          )}

          {error && (
            <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 animate-fade-in">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3" />
                  <span className="text-sm font-medium text-red-800">
                    {error}
                  </span>
                </div>
                <button
                  onClick={clearMessages}
                  className="text-red-500 hover:text-red-700"
                >
                  ×
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Filters Section */}
        {showFilters && (
          <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-6 mb-8 animate-fade-in">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Collection Type */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-3">
                  Collection Type
                </label>
                <select
                  value={selectedListType}
                  onChange={(e) => setSelectedListType(e.target.value)}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                >
                  <option value="user">👤 My Collections</option>
                  <option value="shared">🤝 Shared with Me</option>
                  <option value="filtered">🔍 Filtered View</option>
                  <option value="root">🏠 Root Collections</option>
                  <option value="byParent">👨‍👩‍👧‍👦 By Parent</option>
                </select>
              </div>

              {/* Filter Options for Filtered Type */}
              {selectedListType === "filtered" && (
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-3">
                    Include
                  </label>
                  <div className="space-y-3">
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={includeOwned}
                        onChange={(e) => setIncludeOwned(e.target.checked)}
                        className="h-4 w-4 text-red-800 border-gray-300 rounded focus:ring-red-500"
                      />
                      <span className="ml-3 text-sm text-gray-700">
                        Owned Collections
                      </span>
                    </label>
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={includeShared}
                        onChange={(e) => setIncludeShared(e.target.checked)}
                        className="h-4 w-4 text-red-800 border-gray-300 rounded focus:ring-red-500"
                      />
                      <span className="ml-3 text-sm text-gray-700">
                        Shared Collections
                      </span>
                    </label>
                  </div>
                </div>
              )}

              {/* Parent ID for ByParent Type */}
              {selectedListType === "byParent" && (
                <div>
                  <label className="block text-sm font-semibold text-gray-700 mb-3">
                    Parent Collection ID
                  </label>
                  <input
                    type="text"
                    value={parentId}
                    onChange={(e) => setParentId(e.target.value)}
                    placeholder="Enter parent collection UUID..."
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 font-mono text-sm"
                  />
                </div>
              )}

              {/* Search */}
              <div>
                <label className="block text-sm font-semibold text-gray-700 mb-3">
                  Search Collections
                </label>
                <div className="relative">
                  <input
                    type="text"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    placeholder="Search by name or ID..."
                    className="w-full px-4 py-3 pl-10 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                  />
                  <MagnifyingGlassIcon className="h-5 w-5 text-gray-400 absolute left-3 top-1/2 transform -translate-y-1/2" />
                </div>
              </div>
            </div>

            {/* Apply Filters Button */}
            <div className="mt-6 flex justify-end">
              <button
                onClick={() => handleListCollections(false)}
                disabled={isLoading || !isAuthenticated}
                className="inline-flex items-center px-6 py-3 border border-transparent rounded-lg shadow-sm text-sm font-semibold text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transform hover:scale-105 transition-all duration-200"
              >
                {isLoading ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b border-white mr-2"></div>
                    Loading...
                  </>
                ) : (
                  <>
                    <MagnifyingGlassIcon className="h-4 w-4 mr-2" />
                    Apply Filters
                  </>
                )}
              </button>
            </div>
          </div>
        )}

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
            <div className="flex items-center">
              <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg">
                <UserIcon className="h-6 w-6 text-white" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">
                  My Collections
                </p>
                <p className="text-2xl font-black text-gray-900">
                  {totalCollections}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
            <div className="flex items-center">
              <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-green-500 to-green-600 rounded-lg">
                <UsersIcon className="h-6 w-6 text-white" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Shared</p>
                <p className="text-2xl font-black text-gray-900">
                  {totalSharedCollections}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
            <div className="flex items-center">
              <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg">
                <FolderIcon className="h-6 w-6 text-white" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Folders</p>
                <p className="text-2xl font-black text-gray-900">
                  {folders.length}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
            <div className="flex items-center">
              <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-pink-500 to-pink-600 rounded-lg">
                <PhotoIcon className="h-6 w-6 text-white" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Albums</p>
                <p className="text-2xl font-black text-gray-900">
                  {albums.length}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Collections Grid */}
        <div className="bg-white rounded-2xl shadow-xl border border-gray-100">
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900">
                Collections ({searchResults.length})
              </h2>
              <div className="flex items-center space-x-2 text-sm text-gray-500">
                <span>📁 {folders.length} folders</span>
                <span>•</span>
                <span>📷 {albums.length} albums</span>
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

          {/* Empty State */}
          {!isLoading && searchResults.length === 0 && (
            <div className="text-center py-12">
              <div className="flex justify-center mb-4">
                <div className="flex items-center justify-center h-16 w-16 bg-gradient-to-br from-gray-400 to-gray-500 rounded-2xl">
                  <FolderIcon className="h-8 w-8 text-white" />
                </div>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                No collections found
              </h3>
              <p className="text-gray-600 mb-6">
                {currentCollections.length === 0
                  ? "Create your first collection to get started"
                  : "No collections match your search criteria"}
              </p>
              <button
                onClick={() =>
                  navigate("/developer/create-collection-manager-example")
                }
                className="inline-flex items-center px-6 py-3 border border-transparent rounded-lg shadow-sm text-sm font-semibold text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transform hover:scale-105 transition-all duration-200"
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                Create Collection
              </button>
            </div>
          )}

          {/* Collections List */}
          {!isLoading && searchResults.length > 0 && (
            <div className="divide-y divide-gray-200">
              {searchResults.map((collection) => (
                <div
                  key={collection.id}
                  className="p-6 hover:bg-gray-50 transition-colors duration-200"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center flex-1">
                      <div
                        className={`flex items-center justify-center h-12 w-12 rounded-lg mr-4 ${
                          collection.collection_type === "album"
                            ? "bg-gradient-to-br from-pink-500 to-pink-600"
                            : "bg-gradient-to-br from-blue-500 to-blue-600"
                        }`}
                      >
                        {getCollectionIcon(collection)}
                        <span className="text-white">
                          {getCollectionIcon(collection)}
                        </span>
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center">
                          <h3 className="text-lg font-semibold text-gray-900">
                            {collection.name || "[Encrypted]"}
                          </h3>
                          <span className="ml-2 text-xs text-gray-500">
                            v{collection.version || "?"}
                          </span>
                          {collection._isDecrypted ? (
                            <ShieldCheckIcon
                              className="h-4 w-4 text-green-500 ml-2"
                              title="Decrypted"
                            />
                          ) : (
                            <LockClosedIcon
                              className="h-4 w-4 text-red-500 ml-2"
                              title="Encrypted"
                            />
                          )}
                        </div>
                        <div className="text-sm text-gray-600 mt-1">
                          <span className="capitalize">
                            {collection.collection_type}
                          </span>
                          {" • "}
                          <span>
                            Created {formatDate(collection.created_at)}
                          </span>
                          {collection.members &&
                            collection.members.length > 0 && (
                              <>
                                {" • "}
                                <span className="text-blue-600">
                                  Shared with {collection.members.length}{" "}
                                  user(s)
                                </span>
                              </>
                            )}
                        </div>
                        <div className="text-xs text-gray-500 mt-1 font-mono">
                          ID: {collection.id}
                        </div>
                        {collection._decryptionError && (
                          <div className="text-xs text-red-600 mt-1">
                            Decryption Error: {collection._decryptionError}
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <button
                        onClick={() => toggleDetails(collection.id)}
                        className="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100 transition-colors duration-200"
                        title="View Details"
                      >
                        <EyeIcon className="h-4 w-4" />
                      </button>
                      <button
                        className="p-2 text-gray-400 hover:text-blue-600 rounded-lg hover:bg-blue-50 transition-colors duration-200"
                        title="Share"
                      >
                        <ShareIcon className="h-4 w-4" />
                      </button>
                      <button
                        className="p-2 text-gray-400 hover:text-red-600 rounded-lg hover:bg-red-50 transition-colors duration-200"
                        title="Delete"
                      >
                        <TrashIcon className="h-4 w-4" />
                      </button>
                    </div>
                  </div>

                  {/* Collection Details */}
                  {showDetails[collection.id] && (
                    <div className="mt-4 p-4 bg-gray-50 rounded-lg border">
                      <h4 className="text-sm font-semibold text-gray-700 mb-3">
                        Collection Details
                      </h4>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                        <div>
                          <span className="font-medium text-gray-600">
                            Type:
                          </span>
                          <span className="ml-2 text-gray-900 capitalize">
                            {collection.collection_type}
                          </span>
                        </div>
                        <div>
                          <span className="font-medium text-gray-600">
                            Owner:
                          </span>
                          <span className="ml-2 text-gray-900 font-mono text-xs">
                            {collection.owner_id}
                          </span>
                        </div>
                        <div>
                          <span className="font-medium text-gray-600">
                            Created:
                          </span>
                          <span className="ml-2 text-gray-900">
                            {formatDate(collection.created_at)}
                          </span>
                        </div>
                        <div>
                          <span className="font-medium text-gray-600">
                            Modified:
                          </span>
                          <span className="ml-2 text-gray-900">
                            {formatDate(collection.modified_at)}
                          </span>
                        </div>
                        <div>
                          <span className="font-medium text-gray-600">
                            Version:
                          </span>
                          <span className="ml-2 text-gray-900">
                            {collection.version || "Unknown"}
                          </span>
                        </div>
                        <div>
                          <span className="font-medium text-gray-600">
                            Decrypted:
                          </span>
                          <span
                            className={`ml-2 font-medium ${collection._isDecrypted ? "text-green-600" : "text-red-600"}`}
                          >
                            {collection._isDecrypted ? "Yes" : "No"}
                          </span>
                        </div>
                      </div>
                      {collection.members && collection.members.length > 0 && (
                        <div className="mt-3">
                          <span className="font-medium text-gray-600">
                            Shared with:
                          </span>
                          <div className="ml-2 mt-1">
                            {collection.members.map((member, index) => (
                              <span
                                key={index}
                                className="inline-block bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded mr-2 mb-1"
                              >
                                {member.email || member.user_id}
                              </span>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Cache Management */}
        <div className="mt-8 bg-white rounded-xl shadow-lg border border-gray-100 p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">
              Cache Management
            </h3>
            <button
              onClick={handleClearAllCache}
              className="inline-flex items-center px-4 py-2 border border-red-300 rounded-lg text-sm font-medium text-red-700 bg-red-50 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
            >
              <TrashIcon className="h-4 w-4 mr-2" />
              Clear All Cache
            </button>
          </div>
          <p className="text-sm text-gray-600">
            Clear cached collection data to force refresh from server. This
            helps resolve sync issues.
          </p>
        </div>
      </div>

      {/* CSS Animations */}
      <style jsx>{`
        @keyframes fade-in {
          from {
            opacity: 0;
            transform: translateY(10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        .animate-fade-in {
          animation: fade-in 0.4s ease-out;
        }
      `}</style>
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(CollectionList, {
  redirectTo: "/login",
  showLoadingWhileChecking: true,
  customMessage: "Please log in to access your collections",
});
