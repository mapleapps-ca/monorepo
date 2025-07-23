// File: src/pages/FileManager/Collections/CollectionDetails.jsx
import React, { useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import Navigation from "../../../components/Navigation";
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
} from "@heroicons/react/24/outline";
import { StarIcon as StarIconSolid } from "@heroicons/react/24/solid";

const CollectionDetails = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();
  const [viewMode, setViewMode] = useState("grid");
  const [selectedItems, setSelectedItems] = useState(new Set());
  const [searchQuery, setSearchQuery] = useState("");
  const [showInfo, setShowInfo] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);

  // Mock collection data
  const mockCollection = {
    id: collectionId || "1",
    name: "Work Documents",
    type: "folder",
    description: "Important work-related documents and presentations",
    created: "January 15, 2024",
    modified: "2 hours ago",
    owner: "You",
    starred: true,
    totalSize: "1.2 GB",
    itemCount: 24,
    sharedWith: [
      {
        id: 1,
        name: "Alice Johnson",
        email: "alice@example.com",
        role: "viewer",
      },
      { id: 2, name: "Bob Smith", email: "bob@example.com", role: "editor" },
    ],
  };

  // Mock files in collection
  const mockFiles = [
    {
      id: 1,
      name: "Q4 Report.pdf",
      type: "pdf",
      size: "2.4 MB",
      modified: "1 hour ago",
      starred: false,
    },
    {
      id: 2,
      name: "Budget 2024.xlsx",
      type: "spreadsheet",
      size: "1.8 MB",
      modified: "2 hours ago",
      starred: true,
    },
    {
      id: 3,
      name: "Team Meeting Notes.docx",
      type: "document",
      size: "156 KB",
      modified: "1 day ago",
      starred: false,
    },
    {
      id: 4,
      name: "Product Roadmap.pptx",
      type: "presentation",
      size: "4.2 MB",
      modified: "3 days ago",
      starred: false,
    },
    {
      id: 5,
      name: "Contract Draft.pdf",
      type: "pdf",
      size: "892 KB",
      modified: "1 week ago",
      starred: false,
    },
    {
      id: 6,
      name: "Marketing Strategy.docx",
      type: "document",
      size: "2.1 MB",
      modified: "2 weeks ago",
      starred: true,
    },
  ];

  // Mock sub-collections
  const mockSubCollections = [
    {
      id: 10,
      name: "2023 Archive",
      type: "folder",
      items: 45,
      modified: "1 month ago",
    },
    {
      id: 11,
      name: "Templates",
      type: "folder",
      items: 12,
      modified: "2 weeks ago",
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
      default:
        return <DocumentIcon className={iconClass} />;
    }
  };

  const filteredFiles = mockFiles.filter((file) =>
    file.name.toLowerCase().includes(searchQuery.toLowerCase()),
  );

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
          <span className="font-medium text-gray-900">
            {mockCollection.name}
          </span>
        </div>

        {/* Collection Header */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex items-start justify-between">
            <div className="flex items-start space-x-4">
              <div
                className={`flex items-center justify-center h-14 w-14 rounded-lg ${
                  mockCollection.type === "album"
                    ? "bg-pink-100 text-pink-600"
                    : "bg-blue-100 text-blue-600"
                }`}
              >
                {mockCollection.type === "album" ? (
                  <PhotoIcon className="h-7 w-7" />
                ) : (
                  <FolderIcon className="h-7 w-7" />
                )}
              </div>
              <div>
                <div className="flex items-center space-x-3">
                  <h1 className="text-2xl font-bold text-gray-900">
                    {mockCollection.name}
                  </h1>
                  <button>
                    {mockCollection.starred ? (
                      <StarIconSolid className="h-5 w-5 text-yellow-400" />
                    ) : (
                      <StarIcon className="h-5 w-5 text-gray-300 hover:text-yellow-400" />
                    )}
                  </button>
                </div>
                <p className="text-gray-600 mt-1">
                  {mockCollection.description}
                </p>
                <div className="flex items-center space-x-4 mt-3 text-sm text-gray-500">
                  <span className="flex items-center">
                    <DocumentIcon className="h-4 w-4 mr-1" />
                    {mockCollection.itemCount} items
                  </span>
                  <span>•</span>
                  <span>{mockCollection.totalSize}</span>
                  <span>•</span>
                  <span className="flex items-center">
                    <ClockIcon className="h-4 w-4 mr-1" />
                    Modified {mockCollection.modified}
                  </span>
                  {mockCollection.sharedWith.length > 0 && (
                    <>
                      <span>•</span>
                      <span className="flex items-center text-blue-600">
                        <UsersIcon className="h-4 w-4 mr-1" />
                        Shared with {mockCollection.sharedWith.length}
                      </span>
                    </>
                  )}
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setShowShareModal(true)}
                className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ShareIcon className="h-4 w-4 mr-2" />
                Share
              </button>
              <button className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
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
                        {mockCollection.type}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Created:</span>
                      <span className="font-medium">
                        {mockCollection.created}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Owner:</span>
                      <span className="font-medium">
                        {mockCollection.owner}
                      </span>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-semibold text-gray-600 mb-2">
                    Encryption
                  </h3>
                  <div className="space-y-2 text-sm">
                    <div className="flex items-center text-green-600">
                      <ShieldCheckIcon className="h-4 w-4 mr-2" />
                      <span className="font-medium">End-to-End Encrypted</span>
                    </div>
                    <div className="text-gray-500">
                      ChaCha20-Poly1305 encryption
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-sm font-semibold text-gray-600 mb-2">
                    Shared With
                  </h3>
                  <div className="space-y-2">
                    {mockCollection.sharedWith.map((user) => (
                      <div key={user.id} className="flex items-center text-sm">
                        <div className="flex items-center justify-center h-6 w-6 bg-red-100 text-red-600 text-xs font-medium rounded-full mr-2">
                          {user.name
                            .split(" ")
                            .map((n) => n[0])
                            .join("")}
                        </div>
                        <span className="text-gray-700">{user.name}</span>
                        <span className="ml-auto text-gray-500">
                          {user.role}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}
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
        {mockSubCollections.length > 0 && (
          <div className="mb-6">
            <h2 className="text-sm font-semibold text-gray-600 mb-3">
              Folders
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3">
              {mockSubCollections.map((subCollection) => (
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

        {/* Files */}
        <div>
          <h2 className="text-sm font-semibold text-gray-600 mb-3">Files</h2>

          {viewMode === "grid" ? (
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
                      className={`flex items-center justify-center h-12 w-12 rounded-lg mb-3 ${
                        file.type === "pdf"
                          ? "bg-red-100"
                          : file.type === "spreadsheet"
                            ? "bg-green-100"
                            : file.type === "document"
                              ? "bg-blue-100"
                              : "bg-orange-100"
                      }`}
                    >
                      {getFileIcon(file.type)}
                    </div>
                    <p className="font-medium text-gray-900 truncate w-full">
                      {file.name}
                    </p>
                    <p className="text-xs text-gray-500 mt-1">{file.size}</p>
                    <p className="text-xs text-gray-400">{file.modified}</p>
                  </div>

                  <div className="flex items-center justify-between mt-3 pt-3 border-t">
                    <button className="text-gray-400 hover:text-gray-600">
                      <ArrowDownTrayIcon className="h-4 w-4" />
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
                          {getFileIcon(file.type)}
                          <span className="ml-3 font-medium text-gray-900">
                            {file.name}
                          </span>
                          {file.starred && (
                            <StarIconSolid className="h-4 w-4 text-yellow-400 ml-2" />
                          )}
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
                          <button className="text-gray-400 hover:text-gray-600">
                            <ArrowDownTrayIcon className="h-4 w-4" />
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

export default CollectionDetails;
