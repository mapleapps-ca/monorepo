// File: src/pages/FileManager/Files/FileDetails.jsx
import React, { useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import Navigation from "../../../../components/Navigation";
import {
  DocumentIcon,
  ArrowLeftIcon,
  ArrowDownTrayIcon,
  ShareIcon,
  PencilIcon,
  TrashIcon,
  ChevronRightIcon,
  HomeIcon,
  StarIcon,
  LockClosedIcon,
  ClockIcon,
  InformationCircleIcon,
  CheckIcon,
  XMarkIcon,
  ShieldCheckIcon,
  LinkIcon,
  DocumentDuplicateIcon,
  FolderIcon,
  UsersIcon,
  CalendarIcon,
  ServerIcon,
  EyeIcon,
  PhotoIcon,
  PlayIcon,
  SpeakerWaveIcon,
  ChevronLeftIcon,
  ChevronRightIcon as NextIcon,
} from "@heroicons/react/24/outline";
import { StarIcon as StarIconSolid } from "@heroicons/react/24/solid";

const FileDetails = () => {
  const navigate = useNavigate();
  const { fileId } = useParams();
  const [activeTab, setActiveTab] = useState("details");
  const [showShareModal, setShowShareModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showRenameModal, setShowRenameModal] = useState(false);
  const [newFileName, setNewFileName] = useState("");

  // Mock file data
  const mockFile = {
    id: fileId || "1",
    name: "Q4 Financial Report 2024.pdf",
    type: "pdf",
    size: "2.4 MB",
    mimeType: "application/pdf",
    created: "January 20, 2024 at 2:30 PM",
    modified: "2 hours ago",
    owner: "You",
    starred: true,
    encrypted: true,
    collection: { id: "1", name: "Work Documents", type: "folder" },
    sharedWith: [
      {
        id: 1,
        name: "Alice Johnson",
        email: "alice@example.com",
        role: "viewer",
        avatar: "AJ",
      },
      {
        id: 2,
        name: "Bob Smith",
        email: "bob@example.com",
        role: "editor",
        avatar: "BS",
      },
    ],
    versions: [
      {
        id: 1,
        version: "1.2",
        date: "2 hours ago",
        size: "2.4 MB",
        user: "You",
      },
      { id: 2, version: "1.1", date: "Yesterday", size: "2.3 MB", user: "You" },
      {
        id: 3,
        version: "1.0",
        date: "January 20, 2024",
        size: "2.1 MB",
        user: "You",
      },
    ],
    activity: [
      { id: 1, action: "Modified", user: "You", date: "2 hours ago" },
      {
        id: 2,
        action: "Shared with Bob Smith",
        user: "You",
        date: "1 day ago",
      },
      {
        id: 3,
        action: "Shared with Alice Johnson",
        user: "You",
        date: "3 days ago",
      },
      { id: 4, action: "Uploaded", user: "You", date: "January 20, 2024" },
    ],
  };

  // For image preview
  const isImage = mockFile.mimeType?.startsWith("image/");
  const isVideo = mockFile.mimeType?.startsWith("video/");
  const isAudio = mockFile.mimeType?.startsWith("audio/");
  const isPDF = mockFile.mimeType === "application/pdf";

  const getFileIcon = () => {
    if (isImage) return <PhotoIcon className="h-8 w-8" />;
    if (isVideo) return <PlayIcon className="h-8 w-8" />;
    if (isAudio) return <SpeakerWaveIcon className="h-8 w-8" />;
    if (isPDF) return <DocumentIcon className="h-8 w-8 text-red-600" />;
    return <DocumentIcon className="h-8 w-8" />;
  };

  const handleRename = () => {
    setShowRenameModal(false);
    setNewFileName("");
    // Mock rename action
  };

  const handleDelete = () => {
    navigate("/file-manager/collections");
  };

  const handleDownload = () => {
    // Mock download action
    console.log("Downloading file...");
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
          <Link
            to={`/file-manager/collections/${mockFile.collection.id}`}
            className="hover:text-gray-900"
          >
            {mockFile.collection.name}
          </Link>
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900 truncate max-w-xs">
            {mockFile.name}
          </span>
        </div>

        {/* Back Button */}
        <button
          onClick={() =>
            navigate(`/file-manager/collections/${mockFile.collection.id}`)
          }
          className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-6 transition-colors duration-200"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-1" />
          Back to {mockFile.collection.name}
        </button>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content Area */}
          <div className="lg:col-span-2 space-y-6">
            {/* File Preview */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200">
              {/* Preview Header */}
              <div className="px-6 py-4 border-b border-gray-200">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div
                      className={`flex items-center justify-center h-12 w-12 rounded-lg ${
                        isPDF ? "bg-red-100" : "bg-gray-100"
                      }`}
                    >
                      {getFileIcon()}
                    </div>
                    <div>
                      <h1 className="text-lg font-semibold text-gray-900">
                        {mockFile.name}
                      </h1>
                      <p className="text-sm text-gray-500">
                        {mockFile.size} • {mockFile.type.toUpperCase()}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center space-x-2">
                    <button className="p-2 text-gray-400 hover:text-gray-600">
                      <ChevronLeftIcon className="h-5 w-5" />
                    </button>
                    <button className="p-2 text-gray-400 hover:text-gray-600">
                      <NextIcon className="h-5 w-5" />
                    </button>
                  </div>
                </div>
              </div>

              {/* Preview Content */}
              <div className="p-8 bg-gray-50">
                <div className="flex flex-col items-center justify-center h-96 text-center">
                  <div className="mb-6">
                    <div
                      className={`inline-flex items-center justify-center h-24 w-24 rounded-xl ${
                        isPDF ? "bg-red-100" : "bg-gray-200"
                      }`}
                    >
                      {getFileIcon()}
                    </div>
                  </div>
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    File Preview
                  </h3>
                  <p className="text-sm text-gray-500 mb-6">
                    Preview is not available for this file type in the prototype
                  </p>
                  <button
                    onClick={handleDownload}
                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950"
                  >
                    <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
                    Download to View
                  </button>
                </div>
              </div>
            </div>

            {/* Tabs */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200">
              <div className="border-b border-gray-200">
                <nav className="-mb-px flex">
                  <button
                    onClick={() => setActiveTab("details")}
                    className={`px-6 py-3 border-b-2 text-sm font-medium transition-colors duration-200 ${
                      activeTab === "details"
                        ? "border-red-500 text-red-600"
                        : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                    }`}
                  >
                    Details
                  </button>
                  <button
                    onClick={() => setActiveTab("activity")}
                    className={`px-6 py-3 border-b-2 text-sm font-medium transition-colors duration-200 ${
                      activeTab === "activity"
                        ? "border-red-500 text-red-600"
                        : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                    }`}
                  >
                    Activity
                  </button>
                  <button
                    onClick={() => setActiveTab("versions")}
                    className={`px-6 py-3 border-b-2 text-sm font-medium transition-colors duration-200 ${
                      activeTab === "versions"
                        ? "border-red-500 text-red-600"
                        : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                    }`}
                  >
                    Versions
                  </button>
                </nav>
              </div>

              <div className="p-6">
                {/* Details Tab */}
                {activeTab === "details" && (
                  <div className="space-y-6">
                    <div>
                      <h3 className="text-sm font-semibold text-gray-600 mb-3">
                        File Information
                      </h3>
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                          <p className="text-gray-500">Type</p>
                          <p className="font-medium text-gray-900">
                            {mockFile.mimeType}
                          </p>
                        </div>
                        <div>
                          <p className="text-gray-500">Size</p>
                          <p className="font-medium text-gray-900">
                            {mockFile.size}
                          </p>
                        </div>
                        <div>
                          <p className="text-gray-500">Created</p>
                          <p className="font-medium text-gray-900">
                            {mockFile.created}
                          </p>
                        </div>
                        <div>
                          <p className="text-gray-500">Modified</p>
                          <p className="font-medium text-gray-900">
                            {mockFile.modified}
                          </p>
                        </div>
                      </div>
                    </div>

                    <div>
                      <h3 className="text-sm font-semibold text-gray-600 mb-3">
                        Location
                      </h3>
                      <div className="flex items-center space-x-2 text-sm">
                        <FolderIcon className="h-4 w-4 text-gray-400" />
                        <Link
                          to={`/file-manager/collections/${mockFile.collection.id}`}
                          className="text-blue-600 hover:text-blue-700"
                        >
                          {mockFile.collection.name}
                        </Link>
                      </div>
                    </div>

                    <div>
                      <h3 className="text-sm font-semibold text-gray-600 mb-3">
                        Security
                      </h3>
                      <div className="flex items-center space-x-2 text-sm">
                        <ShieldCheckIcon className="h-4 w-4 text-green-600" />
                        <span className="font-medium text-green-600">
                          End-to-End Encrypted
                        </span>
                      </div>
                      <p className="text-xs text-gray-500 mt-1">
                        ChaCha20-Poly1305 encryption
                      </p>
                    </div>
                  </div>
                )}

                {/* Activity Tab */}
                {activeTab === "activity" && (
                  <div className="space-y-4">
                    {mockFile.activity.map((activity) => (
                      <div
                        key={activity.id}
                        className="flex items-start space-x-3"
                      >
                        <div className="flex-shrink-0">
                          <div className="h-8 w-8 bg-gray-100 rounded-full flex items-center justify-center">
                            <ClockIcon className="h-4 w-4 text-gray-500" />
                          </div>
                        </div>
                        <div className="flex-1">
                          <p className="text-sm text-gray-900">
                            <span className="font-medium">{activity.user}</span>{" "}
                            {activity.action}
                          </p>
                          <p className="text-xs text-gray-500">
                            {activity.date}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {/* Versions Tab */}
                {activeTab === "versions" && (
                  <div className="space-y-3">
                    {mockFile.versions.map((version) => (
                      <div
                        key={version.id}
                        className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                      >
                        <div>
                          <p className="font-medium text-gray-900">
                            Version {version.version}
                          </p>
                          <p className="text-sm text-gray-500">
                            {version.size} • {version.date} • by {version.user}
                          </p>
                        </div>
                        <button className="text-sm text-blue-600 hover:text-blue-700">
                          Restore
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Actions */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                Actions
              </h3>
              <div className="space-y-3">
                <button
                  onClick={handleDownload}
                  className="w-full inline-flex items-center justify-center px-4 py-2 border border-transparent rounded-lg shadow-sm text-sm font-medium text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950"
                >
                  <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
                  Download
                </button>

                <button
                  onClick={() => setShowShareModal(true)}
                  className="w-full inline-flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  <ShareIcon className="h-4 w-4 mr-2" />
                  Share
                </button>

                <button
                  onClick={() => {
                    setNewFileName(mockFile.name);
                    setShowRenameModal(true);
                  }}
                  className="w-full inline-flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
                >
                  <PencilIcon className="h-4 w-4 mr-2" />
                  Rename
                </button>

                <button className="w-full inline-flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                  <DocumentDuplicateIcon className="h-4 w-4 mr-2" />
                  Make a Copy
                </button>

                <button className="w-full inline-flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                  <LinkIcon className="h-4 w-4 mr-2" />
                  Copy Link
                </button>

                <hr className="my-3" />

                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="w-full inline-flex items-center justify-center px-4 py-2 border border-red-300 rounded-lg text-sm font-medium text-red-700 bg-white hover:bg-red-50"
                >
                  <TrashIcon className="h-4 w-4 mr-2" />
                  Delete
                </button>
              </div>
            </div>

            {/* Shared With */}
            {mockFile.sharedWith.length > 0 && (
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
                  <UsersIcon className="h-5 w-5 mr-2 text-gray-500" />
                  Shared With
                </h3>
                <div className="space-y-3">
                  {mockFile.sharedWith.map((user) => (
                    <div
                      key={user.id}
                      className="flex items-center justify-between"
                    >
                      <div className="flex items-center">
                        <div className="flex items-center justify-center h-8 w-8 bg-red-100 text-red-600 text-sm font-medium rounded-full mr-3">
                          {user.avatar}
                        </div>
                        <div>
                          <p className="text-sm font-medium text-gray-900">
                            {user.name}
                          </p>
                          <p className="text-xs text-gray-500">{user.role}</p>
                        </div>
                      </div>
                      <button className="text-xs text-gray-400 hover:text-gray-600">
                        Remove
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* File Properties */}
            <div className="bg-blue-50 border border-blue-200 rounded-xl p-4">
              <div className="flex items-start">
                <InformationCircleIcon className="h-5 w-5 text-blue-600 mr-3 flex-shrink-0 mt-0.5" />
                <div className="text-sm text-blue-800">
                  <h4 className="font-semibold mb-1">File Protection</h4>
                  <p className="text-xs">
                    This file is protected with end-to-end encryption. Only
                    users with permission can decrypt and view it.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Share Modal */}
      {showShareModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">
                Share File
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

      {/* Rename Modal */}
      {showRenameModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">
                Rename File
              </h3>
              <button
                onClick={() => setShowRenameModal(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  File name
                </label>
                <input
                  type="text"
                  value={newFileName}
                  onChange={(e) => setNewFileName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500"
                />
              </div>

              <div className="pt-4 flex justify-end space-x-3">
                <button
                  onClick={() => setShowRenameModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRename}
                  className="px-4 py-2 bg-gradient-to-r from-red-800 to-red-900 text-white rounded-lg hover:from-red-900 hover:to-red-950"
                >
                  Rename
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">
                Delete File
              </h3>
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>

            <div className="mb-6">
              <div className="flex items-center justify-center h-12 w-12 bg-red-100 rounded-lg mb-4">
                <TrashIcon className="h-6 w-6 text-red-600" />
              </div>
              <p className="text-gray-700 mb-2">
                Are you sure you want to delete <strong>{mockFile.name}</strong>
                ?
              </p>
              <p className="text-sm text-gray-600">
                This file will be moved to trash where it will be permanently
                deleted after 30 days.
              </p>
            </div>

            <div className="flex justify-end space-x-3">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
              >
                Move to Trash
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default FileDetails;
