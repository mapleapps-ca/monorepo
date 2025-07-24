// File: src/pages/User/FileManager/Trash/TrashView.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useFiles, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  TrashIcon,
  ArrowLeftIcon,
  ArrowPathIcon,
  DocumentIcon,
  FolderIcon,
  PhotoIcon,
  ExclamationTriangleIcon,
  CheckIcon,
  MagnifyingGlassIcon,
} from "@heroicons/react/24/outline";

const TrashView = () => {
  const navigate = useNavigate();
  const { listFileManager, deleteFileManager, listCollectionManager } =
    useFiles();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [deletedFiles, setDeletedFiles] = useState([]);
  const [deletedCollections, setDeletedCollections] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [showEmptyConfirm, setShowEmptyConfirm] = useState(false);

  // Load deleted items
  const loadDeletedItems = useCallback(
    async (forceRefresh = false) => {
      if (!listFileManager || !listCollectionManager) return;

      setIsLoading(true);
      setError("");

      try {
        // Load deleted files
        const filesResult = await listFileManager.listFiles({
          states: ["deleted"],
          force_refresh: forceRefresh,
        });

        if (filesResult.files) {
          setDeletedFiles(filesResult.files);
        }

        // Load deleted collections
        const collectionsResult =
          await listCollectionManager.listCollections(forceRefresh);

        if (collectionsResult.collections) {
          const deleted = collectionsResult.collections.filter(
            (c) => c.state === "deleted",
          );
          setDeletedCollections(deleted);
        }
      } catch (err) {
        setError("Could not load deleted items");
        console.error("Failed to load trash items:", err);
      } finally {
        setIsLoading(false);
      }
    },
    [listFileManager, listCollectionManager],
  );

  useEffect(() => {
    if (
      listFileManager &&
      listCollectionManager &&
      authManager?.isAuthenticated()
    ) {
      loadDeletedItems();
    }
  }, [listFileManager, listCollectionManager, authManager, loadDeletedItems]);

  // Handle restore
  const handleRestore = async () => {
    if (!deleteFileManager || selectedItems.size === 0) return;

    setError("");

    try {
      const promises = [];

      // Restore files
      selectedItems.forEach((itemId) => {
        if (itemId.startsWith("file-")) {
          const fileId = itemId.replace("file-", "");
          promises.push(deleteFileManager.restoreFile(fileId));
        }
        // Note: Collection restoration would be handled similarly if the API supports it
      });

      await Promise.all(promises);

      setSuccess("Items restored successfully");
      setSelectedItems(new Set());
      await loadDeletedItems(true);
    } catch (err) {
      setError("Failed to restore some items");
      console.error("Restore error:", err);
    }
  };

  // Handle permanent delete
  const handlePermanentDelete = async () => {
    if (!deleteFileManager || selectedItems.size === 0) return;

    setError("");

    try {
      const promises = [];

      // Permanently delete files
      selectedItems.forEach((itemId) => {
        if (itemId.startsWith("file-")) {
          const fileId = itemId.replace("file-", "");
          promises.push(deleteFileManager.deleteFile(fileId, true)); // permanent = true
        }
      });

      await Promise.all(promises);

      setSuccess("Items permanently deleted");
      setSelectedItems(new Set());
      await loadDeletedItems(true);
    } catch (err) {
      setError("Failed to delete some items");
      console.error("Delete error:", err);
    }
  };

  // Handle empty trash
  const handleEmptyTrash = async () => {
    if (!deleteFileManager) return;

    setError("");

    try {
      // Delete all files permanently
      const promises = deletedFiles.map((file) =>
        deleteFileManager.deleteFile(file.id, true),
      );

      await Promise.all(promises);

      setSuccess("Trash emptied successfully");
      setShowEmptyConfirm(false);
      await loadDeletedItems(true);
    } catch (err) {
      setError("Failed to empty trash");
      console.error("Empty trash error:", err);
    }
  };

  // Filter items
  const filteredFiles = deletedFiles.filter((file) =>
    (file.name || "").toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const filteredCollections = deletedCollections.filter((collection) =>
    (collection.name || "").toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const allFilteredItems = [
    ...filteredFiles.map((f) => ({
      ...f,
      itemType: "file",
      itemId: `file-${f.id}`,
    })),
    ...filteredCollections.map((c) => ({
      ...c,
      itemType: "collection",
      itemId: `collection-${c.id}`,
    })),
  ];

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "Unknown";
    const date = new Date(dateString);
    const now = new Date();
    const diffInDays = Math.floor((now - date) / (1000 * 60 * 60 * 24));

    if (diffInDays === 0) return "Today";
    if (diffInDays === 1) return "Yesterday";
    if (diffInDays < 7) return `${diffInDays} days ago`;
    return date.toLocaleDateString();
  };

  // Calculate days until permanent deletion (30 days)
  const getDaysUntilDeletion = (deletedDate) => {
    if (!deletedDate) return 30;
    const deleted = new Date(deletedDate);
    const now = new Date();
    const daysSinceDeletion = Math.floor(
      (now - deleted) / (1000 * 60 * 60 * 24),
    );
    return Math.max(0, 30 - daysSinceDeletion);
  };

  // Toggle selection
  const toggleSelection = (itemId) => {
    const newSelected = new Set(selectedItems);
    if (newSelected.has(itemId)) {
      newSelected.delete(itemId);
    } else {
      newSelected.add(itemId);
    }
    setSelectedItems(newSelected);
  };

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
            <div>
              <h1 className="text-2xl font-semibold text-gray-900 flex items-center">
                <TrashIcon className="h-6 w-6 mr-2" />
                Trash
              </h1>
              <p className="text-sm text-gray-500 mt-1">
                Items in trash will be permanently deleted after 30 days
              </p>
            </div>

            <div className="flex items-center space-x-3">
              <button
                onClick={() => loadDeletedItems(true)}
                disabled={isLoading}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ArrowPathIcon
                  className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
                />
                Refresh
              </button>

              {allFilteredItems.length > 0 && (
                <button
                  onClick={() => setShowEmptyConfirm(true)}
                  className="inline-flex items-center px-4 py-2 border border-red-300 rounded-lg text-sm font-medium text-red-700 bg-white hover:bg-red-50"
                >
                  <TrashIcon className="h-4 w-4 mr-2" />
                  Empty Trash
                </button>
              )}
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
              placeholder="Search in trash..."
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
            />
          </div>
        </div>

        {/* Loading State */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading deleted items...</p>
            </div>
          </div>
        ) : allFilteredItems.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
            <TrashIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {searchQuery ? "No items found" : "Trash is empty"}
            </h3>
            <p className="text-gray-500">
              {searchQuery
                ? "Try adjusting your search"
                : "Items you delete will appear here"}
            </p>
          </div>
        ) : (
          <>
            {/* Items List */}
            <div className="bg-white rounded-lg border border-gray-200">
              <table className="min-w-full">
                <thead className="bg-gray-50 border-b">
                  <tr>
                    <th className="px-6 py-3 text-left">
                      <input
                        type="checkbox"
                        checked={
                          selectedItems.size === allFilteredItems.length &&
                          allFilteredItems.length > 0
                        }
                        onChange={() => {
                          if (selectedItems.size === allFilteredItems.length) {
                            setSelectedItems(new Set());
                          } else {
                            setSelectedItems(
                              new Set(
                                allFilteredItems.map((item) => item.itemId),
                              ),
                            );
                          }
                        }}
                        className="rounded border-gray-300 text-red-800 focus:ring-red-500"
                      />
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Type
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Deleted
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Days Left
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {allFilteredItems.map((item) => {
                    const daysLeft = getDaysUntilDeletion(item.modified_at);
                    return (
                      <tr key={item.itemId} className="hover:bg-gray-50">
                        <td className="px-6 py-4">
                          <input
                            type="checkbox"
                            checked={selectedItems.has(item.itemId)}
                            onChange={() => toggleSelection(item.itemId)}
                            className="rounded border-gray-300 text-red-800 focus:ring-red-500"
                          />
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center">
                            <div className="flex items-center justify-center h-8 w-8 rounded-lg bg-gray-100 mr-3">
                              {item.itemType === "collection" ? (
                                <FolderIcon className="h-5 w-5 text-blue-600" />
                              ) : item.mime_type?.startsWith("image/") ? (
                                <PhotoIcon className="h-5 w-5 text-pink-600" />
                              ) : (
                                <DocumentIcon className="h-5 w-5 text-gray-600" />
                              )}
                            </div>
                            <span className="text-sm font-medium text-gray-900">
                              {item.name || "Encrypted"}
                            </span>
                          </div>
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">
                          {item.itemType === "collection" ? "Folder" : "File"}
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">
                          {formatDate(item.modified_at)}
                        </td>
                        <td className="px-6 py-4">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                              daysLeft <= 7
                                ? "bg-red-100 text-red-800"
                                : daysLeft <= 14
                                  ? "bg-yellow-100 text-yellow-800"
                                  : "bg-gray-100 text-gray-800"
                            }`}
                          >
                            {daysLeft} days
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Action Bar */}
            {selectedItems.size > 0 && (
              <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 bg-gray-900 text-white rounded-lg shadow-lg px-6 py-3 flex items-center space-x-4">
                <span className="text-sm">{selectedItems.size} selected</span>
                <div className="h-4 w-px bg-gray-600"></div>
                <button
                  onClick={handleRestore}
                  className="text-sm hover:text-green-400"
                >
                  <ArrowPathIcon className="h-4 w-4 mr-1 inline" />
                  Restore
                </button>
                <button
                  onClick={handlePermanentDelete}
                  className="text-sm hover:text-red-400"
                >
                  <TrashIcon className="h-4 w-4 mr-1 inline" />
                  Delete Forever
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {/* Empty Trash Confirmation */}
      {showEmptyConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">
              Empty Trash
            </h3>
            <div className="mb-6">
              <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
              <p className="text-gray-700 text-center">
                Are you sure you want to empty the trash? This will permanently
                delete {allFilteredItems.length} items.
              </p>
              <p className="text-sm text-gray-500 text-center mt-2">
                This action cannot be undone.
              </p>
            </div>
            <div className="flex justify-end space-x-3">
              <button
                onClick={() => setShowEmptyConfirm(false)}
                className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleEmptyTrash}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
              >
                Empty Trash
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default withPasswordProtection(TrashView);
