// File: src/pages/User/FileManager/Files/FileDetails.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useParams } from "react-router";
import { useFiles, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  DocumentIcon,
  ArrowLeftIcon,
  ArrowDownTrayIcon,
  ShareIcon,
  TrashIcon,
  ClockIcon,
  CheckIcon,
  PhotoIcon,
  PlayIcon,
  SpeakerWaveIcon,
  ExclamationTriangleIcon,
} from "@heroicons/react/24/outline";

const FileDetails = () => {
  const navigate = useNavigate();
  const { fileId } = useParams();

  const { getFileManager, downloadFileManager, deleteFileManager } = useFiles();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [fileDetails, setFileDetails] = useState(null);
  const [versionHistory, setVersionHistory] = useState([]);
  const [activeTab, setActiveTab] = useState("details");
  const [isDownloading, setIsDownloading] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  // Load file details
  const loadFileDetails = useCallback(
    async (forceRefresh = false) => {
      if (!getFileManager || !fileId) return;

      setIsLoading(true);
      setError("");

      try {
        // Get file details
        const file = await getFileManager.getFileById(fileId, forceRefresh);
        setFileDetails(file);

        // Get version history
        try {
          const versions = await getFileManager.getFileVersionHistory(
            fileId,
            forceRefresh,
          );
          setVersionHistory(versions);
        } catch (err) {
          console.warn("Failed to load version history:", err);
          setVersionHistory([]);
        }
      } catch (err) {
        setError("Could not load file details");
        console.error("Failed to load file details:", err);
      } finally {
        setIsLoading(false);
      }
    },
    [getFileManager, fileId],
  );

  useEffect(() => {
    if (getFileManager && fileId && authManager?.isAuthenticated()) {
      loadFileDetails();
    }
  }, [getFileManager, fileId, authManager, loadFileDetails]);

  // Handle file download
  const handleDownloadFile = async () => {
    if (!downloadFileManager || !fileDetails) return;

    setIsDownloading(true);
    setError("");

    try {
      await downloadFileManager.downloadFile(fileId, {
        saveToDisk: true,
      });
    } catch (err) {
      setError("Could not download file");
      console.error("Failed to download file:", err);
    } finally {
      setIsDownloading(false);
    }
  };

  // Handle file deletion
  const handleDeleteFile = async () => {
    if (!deleteFileManager || !fileId) return;

    try {
      await deleteFileManager.deleteFile(fileId, false); // Move to trash
      navigate(`/file-manager/collections/${fileDetails?.collection_id}`);
    } catch (err) {
      setError("Could not delete file");
      console.error("Failed to delete file:", err);
    }
    setShowDeleteConfirm(false);
  };

  // Get file icon
  const getFileIcon = () => {
    if (!fileDetails) return <DocumentIcon className="h-8 w-8" />;

    const mimeType = fileDetails.mime_type || "";
    if (mimeType.startsWith("image/"))
      return <PhotoIcon className="h-8 w-8 text-pink-600" />;
    if (mimeType.startsWith("video/"))
      return <PlayIcon className="h-8 w-8 text-purple-600" />;
    if (mimeType.startsWith("audio/"))
      return <SpeakerWaveIcon className="h-8 w-8 text-green-600" />;
    if (mimeType.includes("pdf"))
      return <DocumentIcon className="h-8 w-8 text-red-600" />;
    return <DocumentIcon className="h-8 w-8 text-gray-600" />;
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  // Format date
  const formatDate = (dateString) => {
    if (!dateString) return "Unknown";
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return "Invalid Date";
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Navigation />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading file details...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!fileDetails) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Navigation />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center py-12">
            <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              File not found
            </h3>
            <button
              onClick={() => navigate("/file-manager")}
              className="text-red-800 hover:text-red-900"
            >
              Back to Files
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-6">
          <button
            onClick={() =>
              navigate(`/file-manager/collections/${fileDetails.collection_id}`)
            }
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to Collection
          </button>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="flex items-center justify-center h-12 w-12 rounded-lg bg-gray-100">
                {getFileIcon()}
              </div>
              <div>
                <h1 className="text-2xl font-semibold text-gray-900">
                  {fileDetails.name || "Encrypted File"}
                </h1>
                <p className="text-sm text-gray-500 mt-1">
                  {formatFileSize(
                    fileDetails.size ||
                      fileDetails.encrypted_file_size_in_bytes,
                  )}{" "}
                  • Modified {formatDate(fileDetails.modified_at)}
                </p>
              </div>
            </div>

            <div className="flex items-center space-x-3">
              <button
                onClick={() =>
                  navigate(
                    `/file-manager/collections/${fileDetails.collection_id}/share`,
                  )
                }
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ShareIcon className="h-4 w-4 mr-2" />
                Share
              </button>
              <button
                onClick={handleDownloadFile}
                disabled={isDownloading || !fileDetails._isDecrypted}
                className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400"
              >
                <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
                {isDownloading ? "Downloading..." : "Download"}
              </button>
            </div>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-3 rounded-lg bg-red-50 border border-red-200">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        {/* Main Content */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* File Preview/Details */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg border border-gray-200">
              {/* Tabs */}
              <div className="border-b border-gray-200">
                <nav className="-mb-px flex">
                  <button
                    onClick={() => setActiveTab("details")}
                    className={`px-6 py-3 border-b-2 text-sm font-medium ${
                      activeTab === "details"
                        ? "border-red-800 text-red-800"
                        : "border-transparent text-gray-500 hover:text-gray-700"
                    }`}
                  >
                    Details
                  </button>
                  <button
                    onClick={() => setActiveTab("versions")}
                    className={`px-6 py-3 border-b-2 text-sm font-medium ${
                      activeTab === "versions"
                        ? "border-red-800 text-red-800"
                        : "border-transparent text-gray-500 hover:text-gray-700"
                    }`}
                  >
                    Version History
                  </button>
                </nav>
              </div>

              <div className="p-6">
                {/* Details Tab */}
                {activeTab === "details" && (
                  <div className="space-y-4">
                    <div>
                      <h3 className="text-sm font-medium text-gray-900 mb-2">
                        File Information
                      </h3>
                      <dl className="space-y-2">
                        <div className="flex justify-between">
                          <dt className="text-sm text-gray-500">Type</dt>
                          <dd className="text-sm text-gray-900">
                            {fileDetails.mime_type || "Unknown"}
                          </dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-sm text-gray-500">Size</dt>
                          <dd className="text-sm text-gray-900">
                            {formatFileSize(
                              fileDetails.size ||
                                fileDetails.encrypted_file_size_in_bytes,
                            )}
                          </dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-sm text-gray-500">Created</dt>
                          <dd className="text-sm text-gray-900">
                            {formatDate(fileDetails.created_at)}
                          </dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-sm text-gray-500">Modified</dt>
                          <dd className="text-sm text-gray-900">
                            {formatDate(fileDetails.modified_at)}
                          </dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-sm text-gray-500">Version</dt>
                          <dd className="text-sm text-gray-900">
                            v{fileDetails.version || 1}
                          </dd>
                        </div>
                        <div className="flex justify-between">
                          <dt className="text-sm text-gray-500">State</dt>
                          <dd className="text-sm text-gray-900 capitalize">
                            {fileDetails.state || "active"}
                          </dd>
                        </div>
                      </dl>
                    </div>

                    {fileDetails._decryptionError && (
                      <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                        <p className="text-sm text-yellow-800">
                          This file could not be decrypted. You may not have
                          access.
                        </p>
                      </div>
                    )}
                  </div>
                )}

                {/* Versions Tab */}
                {activeTab === "versions" && (
                  <div className="space-y-3">
                    {versionHistory.length === 0 ? (
                      <p className="text-center text-gray-500 py-8">
                        No version history available
                      </p>
                    ) : (
                      versionHistory.map((version, index) => (
                        <div
                          key={`${version.id}-${version.version}`}
                          className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                        >
                          <div>
                            <p className="font-medium text-gray-900">
                              Version {version.version}
                            </p>
                            <p className="text-sm text-gray-500">
                              {formatFileSize(
                                version.size ||
                                  version.encrypted_file_size_in_bytes,
                              )}{" "}
                              •{formatDate(version.modified_at)}
                            </p>
                          </div>
                          {index === 0 && (
                            <span className="text-sm text-green-600 font-medium">
                              Current
                            </span>
                          )}
                        </div>
                      ))
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div>
            <div className="bg-white rounded-lg border border-gray-200 p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                Actions
              </h3>
              <div className="space-y-3">
                <button
                  onClick={handleDownloadFile}
                  disabled={isDownloading || !fileDetails._isDecrypted}
                  className="w-full inline-flex items-center justify-center px-4 py-2 rounded-lg text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400"
                >
                  <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
                  {isDownloading ? "Downloading..." : "Download"}
                </button>

                <button
                  onClick={() =>
                    navigate(
                      `/file-manager/collections/${fileDetails.collection_id}/share`,
                    )
                  }
                  className="w-full inline-flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 bg-white hover:bg-gray-50"
                >
                  <ShareIcon className="h-4 w-4 mr-2" />
                  Share Collection
                </button>

                <hr className="my-3" />

                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="w-full inline-flex items-center justify-center px-4 py-2 border border-red-300 rounded-lg text-red-700 bg-white hover:bg-red-50"
                >
                  <TrashIcon className="h-4 w-4 mr-2" />
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">
              Delete File
            </h3>
            <p className="text-gray-700 mb-6">
              Are you sure you want to delete "{fileDetails.name}"? The file
              will be moved to trash and permanently deleted after 30 days.
            </p>
            <div className="flex justify-end space-x-3">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteFile}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default withPasswordProtection(FileDetails);
