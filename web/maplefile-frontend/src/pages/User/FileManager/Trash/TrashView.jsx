// File: src/pages/FileManager/Trash/TrashView.jsx
import React, { useState } from "react";
import { Link, useNavigate } from "react-router";
import Navigation from "../../../../components/Navigation";
import {
  TrashIcon,
  ArrowLeftIcon,
  ArrowPathIcon,
  XMarkIcon,
  DocumentIcon,
  FolderIcon,
  PhotoIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  CheckIcon,
  ClockIcon,
  MagnifyingGlassIcon,
  Squares2X2Icon,
  ListBulletIcon,
  ChevronRightIcon,
  HomeIcon,
  SparklesIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";

const TrashView = () => {
  const navigate = useNavigate();
  const [viewMode, setViewMode] = useState("list");
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [searchQuery, setSearchQuery] = useState("");
  const [showEmptyConfirm, setShowEmptyConfirm] = useState(false);
  const [showRestoreSuccess, setShowRestoreSuccess] = useState(false);
  const [showDeleteSuccess, setShowDeleteSuccess] = useState(false);

  // Mock deleted items
  const mockDeletedItems = [
    {
      id: 1,
      name: "Old Presentation.pptx",
      type: "file",
      fileType: "presentation",
      size: "3.2 MB",
      deletedDate: "2 days ago",
      deletedBy: "You",
      originalLocation: "Work Documents",
      daysUntilPermanentDelete: 28,
    },
    {
      id: 2,
      name: "Archived Projects",
      type: "collection",
      collectionType: "folder",
      itemCount: 23,
      deletedDate: "1 week ago",
      deletedBy: "You",
      originalLocation: "Root",
      daysUntilPermanentDelete: 23,
    },
    {
      id: 3,
      name: "draft_notes.txt",
      type: "file",
      fileType: "text",
      size: "45 KB",
      deletedDate: "2 weeks ago",
      deletedBy: "You",
      originalLocation: "Personal Files",
      daysUntilPermanentDelete: 16,
    },
    {
      id: 4,
      name: "vacation_photo_123.jpg",
      type: "file",
      fileType: "image",
      size: "2.8 MB",
      deletedDate: "3 weeks ago",
      deletedBy: "You",
      originalLocation: "Vacation Photos",
      daysUntilPermanentDelete: 9,
    },
    {
      id: 5,
      name: "Old Receipts",
      type: "collection",
      collectionType: "folder",
      itemCount: 67,
      deletedDate: "25 days ago",
      deletedBy: "You",
      originalLocation: "Finance",
      daysUntilPermanentDelete: 5,
    },
  ];

  const handleSelectItem = (id) => {
    const newSelected = new Set(selectedItems);
    if (newSelected.has(id)) {
      newSelected.delete(id);
    } else {
      newSelected.add(id);
    }
    setSelectedItems(newSelected);
  };

  const handleSelectAll = () => {
    if (selectedItems.size === filteredItems.length) {
      setSelectedItems(new Set());
    } else {
      setSelectedItems(new Set(filteredItems.map((item) => item.id)));
    }
  };

  const getIcon = (item) => {
    if (item.type === "collection") {
      return item.collectionType === "album" ? (
        <PhotoIcon className="h-5 w-5" />
      ) : (
        <FolderIcon className="h-5 w-5" />
      );
    }

    switch (item.fileType) {
      case "presentation":
        return <DocumentIcon className="h-5 w-5 text-orange-600" />;
      case "text":
        return <DocumentIcon className="h-5 w-5 text-gray-600" />;
      case "image":
        return <PhotoIcon className="h-5 w-5 text-pink-600" />;
      default:
        return <DocumentIcon className="h-5 w-5" />;
    }
  };

  const getDaysUntilDeleteColor = (days) => {
    if (days <= 7) return "text-red-600 bg-red-50";
    if (days <= 14) return "text-amber-600 bg-amber-50";
    return "text-gray-600 bg-gray-50";
  };

  const filteredItems = mockDeletedItems.filter((item) =>
    item.name.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const handleRestore = () => {
    setShowRestoreSuccess(true);
    setSelectedItems(new Set());
    setTimeout(() => setShowRestoreSuccess(false), 3000);
  };

  const handlePermanentDelete = () => {
    setShowDeleteSuccess(true);
    setSelectedItems(new Set());
    setTimeout(() => setShowDeleteSuccess(false), 3000);
  };

  const handleEmptyTrash = () => {
    setShowEmptyConfirm(false);
    setShowDeleteSuccess(true);
    setTimeout(() => setShowDeleteSuccess(false), 3000);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Breadcrumb */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 mb-6">
          <HomeIcon className="h-4 w-4" />
          <ChevronRightIcon className="h-3 w-3" />
          <Link to="/file-manager/collections" className="hover:text-gray-900">
            My Files
          </Link>
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900">Trash</span>
        </div>

        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => navigate("/file-manager/collections")}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors duration-200"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to Files
          </button>

          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2 flex items-center">
                <TrashIcon className="h-8 w-8 mr-3 text-gray-700" />
                Trash
              </h1>
              <p className="text-gray-600">
                Items in trash will be permanently deleted after 30 days
              </p>
            </div>

            {filteredItems.length > 0 && (
              <button
                onClick={() => setShowEmptyConfirm(true)}
                className="inline-flex items-center px-4 py-2 border border-red-300 rounded-lg text-sm font-medium text-red-700 bg-white hover:bg-red-50 transition-all duration-200"
              >
                <TrashIcon className="h-4 w-4 mr-2" />
                Empty Trash
              </button>
            )}
          </div>
        </div>

        {/* Info Banner */}
        <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 mb-6">
          <div className="flex items-start">
            <InformationCircleIcon className="h-5 w-5 text-blue-600 mr-3 flex-shrink-0 mt-0.5" />
            <div className="text-sm text-blue-800">
              <h3 className="font-semibold mb-1">About Trash</h3>
              <ul className="space-y-1 text-xs">
                <li>• Items remain encrypted while in trash</li>
                <li>• Files are permanently deleted after 30 days</li>
                <li>• Restore items to their original location anytime</li>
                <li>
                  • Emptying trash permanently deletes all items immediately
                </li>
              </ul>
            </div>
          </div>
        </div>

        {/* Success Messages */}
        {showRestoreSuccess && (
          <div className="mb-6 bg-green-50 border border-green-200 rounded-xl p-4 animate-fade-in">
            <div className="flex items-center">
              <CheckIcon className="h-5 w-5 text-green-600 mr-3" />
              <p className="text-green-800">
                Items successfully restored to their original locations
              </p>
            </div>
          </div>
        )}

        {showDeleteSuccess && (
          <div className="mb-6 bg-green-50 border border-green-200 rounded-xl p-4 animate-fade-in">
            <div className="flex items-center">
              <CheckIcon className="h-5 w-5 text-green-600 mr-3" />
              <p className="text-green-800">Items permanently deleted</p>
            </div>
          </div>
        )}

        {/* Search and View Options */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3 flex-1">
              {/* Search */}
              <div className="relative flex-1 max-w-md">
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

        {/* Empty State */}
        {filteredItems.length === 0 && (
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
            <div className="flex items-center justify-center h-16 w-16 bg-gray-100 rounded-full mx-auto mb-4">
              <TrashIcon className="h-8 w-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              {searchQuery ? "No items found" : "Trash is empty"}
            </h3>
            <p className="text-gray-500">
              {searchQuery
                ? "Try adjusting your search"
                : "Items you delete will appear here"}
            </p>
          </div>
        )}

        {/* Items List */}
        {filteredItems.length > 0 && viewMode === "list" && (
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
            <table className="min-w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left">
                    <input
                      type="checkbox"
                      checked={selectedItems.size === filteredItems.length}
                      onChange={handleSelectAll}
                      className="h-4 w-4 text-red-600 rounded border-gray-300"
                    />
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Original Location
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Deleted
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Auto-Delete
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                    Size
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredItems.map((item) => (
                  <tr key={item.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <input
                        type="checkbox"
                        checked={selectedItems.has(item.id)}
                        onChange={() => handleSelectItem(item.id)}
                        className="h-4 w-4 text-red-600 rounded border-gray-300"
                      />
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center">
                        <div
                          className={`flex items-center justify-center h-8 w-8 rounded-lg mr-3 ${
                            item.type === "collection"
                              ? item.collectionType === "album"
                                ? "bg-pink-100"
                                : "bg-blue-100"
                              : item.fileType === "image"
                                ? "bg-pink-100"
                                : item.fileType === "presentation"
                                  ? "bg-orange-100"
                                  : "bg-gray-100"
                          }`}
                        >
                          {getIcon(item)}
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">
                            {item.name}
                          </p>
                          {item.type === "collection" && (
                            <p className="text-xs text-gray-500">
                              {item.itemCount} items
                            </p>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {item.originalLocation}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {item.deletedDate}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getDaysUntilDeleteColor(item.daysUntilPermanentDelete)}`}
                      >
                        {item.daysUntilPermanentDelete} days
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {item.type === "file" ? item.size : "-"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Items Grid */}
        {filteredItems.length > 0 && viewMode === "grid" && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {filteredItems.map((item) => (
              <div
                key={item.id}
                className="bg-white rounded-xl border border-gray-200 p-6 hover:shadow-md transition-all duration-200"
              >
                <div className="flex items-start justify-between mb-4">
                  <input
                    type="checkbox"
                    checked={selectedItems.has(item.id)}
                    onChange={() => handleSelectItem(item.id)}
                    className="h-4 w-4 text-red-600 rounded border-gray-300"
                  />
                  <span
                    className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getDaysUntilDeleteColor(item.daysUntilPermanentDelete)}`}
                  >
                    {item.daysUntilPermanentDelete}d
                  </span>
                </div>

                <div className="text-center">
                  <div
                    className={`inline-flex items-center justify-center h-12 w-12 rounded-lg mb-3 ${
                      item.type === "collection"
                        ? item.collectionType === "album"
                          ? "bg-pink-100 text-pink-600"
                          : "bg-blue-100 text-blue-600"
                        : item.fileType === "image"
                          ? "bg-pink-100"
                          : item.fileType === "presentation"
                            ? "bg-orange-100"
                            : "bg-gray-100"
                    }`}
                  >
                    {getIcon(item)}
                  </div>
                  <h3 className="font-medium text-gray-900 truncate">
                    {item.name}
                  </h3>
                  <p className="text-sm text-gray-500 mt-1">
                    {item.type === "collection"
                      ? `${item.itemCount} items`
                      : item.size}
                  </p>
                  <p className="text-xs text-gray-400 mt-2">
                    Deleted {item.deletedDate}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Selection Actions Bar */}
        {selectedItems.size > 0 && (
          <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 bg-gray-900 text-white rounded-lg shadow-lg px-6 py-3 flex items-center space-x-4 animate-fade-in-up">
            <span className="text-sm font-medium">
              {selectedItems.size} selected
            </span>
            <div className="h-4 w-px bg-gray-600"></div>
            <button
              onClick={handleRestore}
              className="inline-flex items-center text-sm hover:text-green-400 transition-colors duration-200"
            >
              <ArrowPathIcon className="h-4 w-4 mr-1" />
              Restore
            </button>
            <button
              onClick={handlePermanentDelete}
              className="inline-flex items-center text-sm hover:text-red-400 transition-colors duration-200"
            >
              <XMarkIcon className="h-4 w-4 mr-1" />
              Delete Forever
            </button>
          </div>
        )}
      </div>

      {/* Empty Trash Confirmation Modal */}
      {showEmptyConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">
                Empty Trash
              </h3>
              <button
                onClick={() => setShowEmptyConfirm(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>

            <div className="mb-6">
              <div className="flex items-center justify-center h-12 w-12 bg-red-100 rounded-lg mb-4">
                <ExclamationTriangleIcon className="h-6 w-6 text-red-600" />
              </div>
              <p className="text-gray-700 mb-2">
                Are you sure you want to empty the trash?
              </p>
              <p className="text-sm text-gray-600">
                This will permanently delete{" "}
                <strong>{filteredItems.length} items</strong>. This action
                cannot be undone.
              </p>

              <div className="mt-4 p-3 bg-amber-50 rounded-lg border border-amber-200">
                <div className="flex items-start">
                  <ShieldCheckIcon className="h-4 w-4 text-amber-600 mr-2 flex-shrink-0 mt-0.5" />
                  <p className="text-xs text-amber-800">
                    Items will be securely deleted and cannot be recovered, even
                    by MapleFile.
                  </p>
                </div>
              </div>
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

        @keyframes fade-in {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }

        .animate-fade-in-up {
          animation: fade-in-up 0.3s ease-out;
        }

        .animate-fade-in {
          animation: fade-in 0.3s ease-out;
        }
      `}</style>
    </div>
  );
};

export default TrashView;
