// File: src/pages/FileManager/Collections/CollectionsView.jsx
import React, { useState } from "react";
import { Link, useNavigate } from "react-router";
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
} from "@heroicons/react/24/outline";
import { StarIcon as StarIconSolid } from "@heroicons/react/24/solid";

const CollectionsView = () => {
  const navigate = useNavigate();
  const [viewMode, setViewMode] = useState("grid"); // grid, list, columns
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [showDropdown, setShowDropdown] = useState(null);
  const [sortBy, setSortBy] = useState("name");
  const [filterType, setFilterType] = useState("all");
  const [searchQuery, setSearchQuery] = useState("");

  // Mock data for prototype
  const mockCollections = [
    {
      id: 1,
      name: "Work Documents",
      type: "folder",
      items: 24,
      modified: "2 hours ago",
      size: "1.2 GB",
      starred: true,
    },
    {
      id: 2,
      name: "Vacation Photos",
      type: "album",
      items: 156,
      modified: "3 days ago",
      size: "4.5 GB",
      starred: false,
    },
    {
      id: 3,
      name: "Project Files",
      type: "folder",
      items: 42,
      modified: "1 day ago",
      size: "856 MB",
      starred: false,
    },
    {
      id: 4,
      name: "Family Album",
      type: "album",
      items: 89,
      modified: "1 week ago",
      size: "2.3 GB",
      starred: true,
    },
    {
      id: 5,
      name: "Archived Data",
      type: "folder",
      items: 312,
      modified: "1 month ago",
      size: "8.7 GB",
      starred: false,
    },
    {
      id: 6,
      name: "Design Assets",
      type: "folder",
      items: 67,
      modified: "5 hours ago",
      size: "3.2 GB",
      starred: false,
    },
  ];

  const mockRecentFiles = [
    {
      id: 1,
      name: "Budget_2024.xlsx",
      type: "spreadsheet",
      size: "2.4 MB",
      modified: "10 minutes ago",
    },
    {
      id: 2,
      name: "Presentation_Final.pdf",
      type: "pdf",
      size: "5.8 MB",
      modified: "1 hour ago",
    },
    {
      id: 3,
      name: "Meeting_Notes.docx",
      type: "document",
      size: "124 KB",
      modified: "3 hours ago",
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

  const getFileIcon = (type) => {
    const iconClass = "h-5 w-5";
    switch (type) {
      case "spreadsheet":
        return (
          <div className="text-green-600">
            <DocumentIcon className={iconClass} />
          </div>
        );
      case "pdf":
        return (
          <div className="text-red-600">
            <DocumentIcon className={iconClass} />
          </div>
        );
      case "document":
        return (
          <div className="text-blue-600">
            <DocumentIcon className={iconClass} />
          </div>
        );
      default:
        return <DocumentIcon className={iconClass} />;
    }
  };

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
            <p className="text-gray-600">6 collections • 3 recent files</p>
          </div>

          <div className="flex items-center space-x-3">
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
                  placeholder="Search files and collections..."
                  className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                />
              </div>
            </div>

            {/* Filters and View Options */}
            <div className="flex items-center space-x-3">
              {/* Filter Dropdown */}
              <div className="relative">
                <button className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                  <AdjustmentsHorizontalIcon className="h-4 w-4 mr-2" />
                  {filterType === "all" ? "All Types" : filterType}
                  <ChevronDownIcon className="h-4 w-4 ml-2" />
                </button>
              </div>

              {/* Sort Dropdown */}
              <div className="relative">
                <button className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                  Sort by {sortBy}
                  <ChevronDownIcon className="h-4 w-4 ml-2" />
                </button>
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
                <button
                  onClick={() => setViewMode("columns")}
                  className={`p-1.5 rounded ${viewMode === "columns" ? "bg-white shadow-sm" : ""}`}
                >
                  <ViewColumnsIcon
                    className={`h-4 w-4 ${viewMode === "columns" ? "text-red-600" : "text-gray-600"}`}
                  />
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Quick Access Section */}
        <div className="mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
            <ClockIcon className="h-5 w-5 mr-2 text-gray-500" />
            Recent Files
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {mockRecentFiles.map((file) => (
              <div
                key={file.id}
                className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-all duration-200 cursor-pointer group"
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-center space-x-3">
                    {getFileIcon(file.type)}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {file.name}
                      </p>
                      <p className="text-xs text-gray-500">
                        {file.size} • {file.modified}
                      </p>
                    </div>
                  </div>
                  <button className="opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                    <ArrowDownTrayIcon className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Collections Section */}
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
            <FolderIcon className="h-5 w-5 mr-2 text-gray-500" />
            Collections
          </h2>

          {/* Grid View */}
          {viewMode === "grid" && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
              {mockCollections.map((collection) => (
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
                          showDropdown === collection.id ? null : collection.id,
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
                    }`}
                  >
                    {getIcon(collection.type)}
                  </div>

                  {/* Collection Info */}
                  <h3 className="font-semibold text-gray-900 mb-1 truncate">
                    {collection.name}
                  </h3>
                  <p className="text-sm text-gray-500 mb-2">
                    {collection.items} items
                  </p>
                  <div className="flex items-center justify-between text-xs text-gray-400">
                    <span>{collection.size}</span>
                    <span>{collection.modified}</span>
                  </div>

                  {/* Star Icon */}
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      // Toggle star
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
                className="bg-white rounded-xl border-2 border-dashed border-gray-300 p-6 hover:border-red-400 transition-all duration-200 cursor-pointer group flex flex-col items-center justify-center"
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
                  {mockCollections.map((collection) => (
                    <tr
                      key={collection.id}
                      className="hover:bg-gray-50 cursor-pointer"
                    >
                      <td className="px-6 py-4">
                        <input
                          type="checkbox"
                          checked={selectedItems.has(collection.id)}
                          onChange={() => handleSelectItem(collection.id)}
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
                            }`}
                          >
                            {getIcon(collection.type)}
                          </div>
                          <div className="flex items-center">
                            <span className="font-medium text-gray-900">
                              {collection.name}
                            </span>
                            {collection.starred && (
                              <StarIconSolid className="h-4 w-4 text-yellow-400 ml-2" />
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
                        {collection.modified}
                      </td>
                      <td className="px-6 py-4 text-sm text-gray-500">
                        <div className="flex items-center space-x-2">
                          <button className="text-gray-400 hover:text-gray-600">
                            <ShareIcon className="h-4 w-4" />
                          </button>
                          <button className="text-gray-400 hover:text-gray-600">
                            <PencilIcon className="h-4 w-4" />
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

export default CollectionsView;
