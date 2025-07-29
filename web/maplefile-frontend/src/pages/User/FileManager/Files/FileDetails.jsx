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
  ArrowPathIcon,
  XMarkIcon,
  LockClosedIcon,
  ServerIcon,
  ShieldCheckIcon,
  InformationCircleIcon,
  ChartBarIcon,
  SparklesIcon,
  DocumentDuplicateIcon,
  FolderIcon,
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

  const [isLoadingVersions, setIsLoadingVersions] = useState(false);
  const [versionError, setVersionError] = useState("");
  const [hasLoadedVersions, setHasLoadedVersions] = useState(false);

  const loadFileDetails = useCallback(
    async (forceRefresh = false) => {
      if (!getFileManager || !fileId) return;

      setIsLoading(true);
      setError("");

      try {
        const file = await getFileManager.getFileById(fileId, forceRefresh);
        setFileDetails(file);
      } catch (err) {
        setError("Could not load file details");
        console.error("[FileDetails] Failed to load file details:", err);
      } finally {
        setIsLoading(false);
      }
    },
    [getFileManager, fileId],
  );

  const loadVersionHistory = useCallback(
    async (forceRefresh = false) => {
      if (!fileDetails) {
        return;
      }

      setIsLoadingVersions(true);
      setVersionError("");

      try {
        // Since version history endpoint doesn't exist in backend, create mock from current file
        const mockVersion = {
          id: fileDetails.id,
          version: fileDetails.version || 1,
          name: fileDetails.name,
          state: fileDetails.state || "active",
          size: fileDetails.size,
          encrypted_file_size_in_bytes:
            fileDetails.encrypted_file_size_in_bytes,
          modified_at: fileDetails.modified_at,
          created_at: fileDetails.created_at,
          _isDecrypted: fileDetails._isDecrypted,
          mime_type: fileDetails.mime_type,
          modified_by: "You", // Mock data
          isCurrent: true, // Mock data
        };

        setVersionHistory([mockVersion]);
        setHasLoadedVersions(true);

        setVersionError(
          "Version history is not yet implemented in the backend. Showing current file version only.",
        );
      } catch (err) {
        setVersionError(
          "Could not load version history. This feature is not yet available.",
        );
        setVersionHistory([]);
      } finally {
        setIsLoadingVersions(false);
      }
    },
    [fileDetails, fileId],
  );

  useEffect(() => {
    if (getFileManager && fileId && authManager?.isAuthenticated()) {
      loadFileDetails();
    }
  }, [getFileManager, fileId, authManager, loadFileDetails]);

  useEffect(() => {
    if (
      activeTab === "versions" &&
      !hasLoadedVersions &&
      getFileManager &&
      fileId
    ) {
      loadVersionHistory();
    }
  }, [
    activeTab,
    hasLoadedVersions,
    loadVersionHistory,
    getFileManager,
    fileId,
  ]);

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
      console.error("[FileDetails] Failed to download file:", err);
    } finally {
      setIsDownloading(false);
    }
  };

  const handleDeleteFile = async () => {
    if (!deleteFileManager || !fileId) return;

    try {
      await deleteFileManager.deleteFile(fileId, false); // Move to trash
      navigate(`/file-manager/collections/${fileDetails?.collection_id}`);
    } catch (err) {
      setError("Could not delete file");
      console.error("[FileDetails] Failed to delete file:", err);
    }
    setShowDeleteConfirm(false);
  };

  const getFileIcon = () => {
    const iconClass = "h-8 w-8";
    if (!fileDetails) return <DocumentIcon className={iconClass} />;

    const mimeType = fileDetails.mime_type || "";
    if (mimeType.startsWith("image/"))
      return <PhotoIcon className={`${iconClass} text-purple-600`} />;
    if (mimeType.startsWith("video/"))
      return <PlayIcon className={`${iconClass} text-pink-600`} />;
    if (mimeType.startsWith("audio/"))
      return <SpeakerWaveIcon className={`${iconClass} text-green-600`} />;
    if (mimeType.includes("pdf"))
      return <DocumentIcon className={`${iconClass} text-red-600`} />;
    return <DocumentIcon className={`${iconClass} text-blue-600`} />;
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  const formatDate = (dateString) => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "Unknown";
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatTimeAgo = (dateString) => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z")
      return "a while ago";
    const date = new Date(dateString);
    const now = new Date();
    const diffInMinutes = Math.floor((now - date) / (1000 * 60));

    if (diffInMinutes < 1) return "just now";
    if (diffInMinutes < 60) return `${diffInMinutes} minutes ago`;
    if (diffInMinutes < 1440)
      return `${Math.floor(diffInMinutes / 60)} hours ago`;
    if (diffInMinutes < 2880) return "Yesterday";
    return `${Math.floor(diffInMinutes / 1440)} days ago`;
  };

  const getStateLabel = (state) => {
    return state ? state.charAt(0).toUpperCase() + state.slice(1) : "Unknown";
  };

  const tabs = [
    { id: "details", label: "Details", icon: InformationCircleIcon },
    { id: "versions", label: "Version History", icon: ClockIcon },
    { id: "activity", label: "Activity", icon: ChartBarIcon },
  ];

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-subtle">
        <Navigation />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="h-8 w-8 spinner border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading file details...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error && !fileDetails) {
    return (
      <div className="min-h-screen bg-gradient-subtle">
        <Navigation />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center py-12">
            <ExclamationTriangleIcon className="h-12 w-12 text-red-500 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              File not found or could not be loaded
            </h3>
            <p className="text-sm text-gray-600 mb-4">{error}</p>
            <button
              onClick={() => navigate("/file-manager")}
              className="btn-secondary"
            >
              Back to Files
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!fileDetails) return null; // Should be covered by error state, but for safety

  return (
    <div className="min-h-screen bg-gradient-subtle">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8 animate-fade-in-down">
          <button
            onClick={() =>
              navigate(`/file-manager/collections/${fileDetails.collection_id}`)
            }
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors duration-200"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to Collection
          </button>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className="h-14 w-14 bg-gradient-to-br from-gray-100 to-gray-200 rounded-xl flex items-center justify-center">
                {getFileIcon()}
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900 flex items-center">
                  {fileDetails.name || "Encrypted File"}
                  {fileDetails._isDecrypted && (
                    <CheckIcon className="h-5 w-5 text-green-600 ml-2" />
                  )}
                </h1>
                <div className="flex items-center space-x-4 mt-1">
                  <span className="text-sm text-gray-600">
                    {formatFileSize(
                      fileDetails.size ||
                        fileDetails.encrypted_file_size_in_bytes,
                    )}
                  </span>
                  <span className="text-sm text-gray-400">•</span>
                  <span className="text-sm text-gray-600">
                    Modified {formatTimeAgo(fileDetails.modified_at)}
                  </span>
                  <span className="text-sm text-gray-400">•</span>
                  <span className="text-sm text-gray-600">
                    v{fileDetails.version || 1}
                  </span>
                </div>
              </div>
            </div>

            <div className="flex items-center space-x-3">
              <button
                onClick={() =>
                  navigate(
                    `/file-manager/collections/${fileDetails.collection_id}/share`,
                  )
                }
                className="btn-secondary"
              >
                <ShareIcon className="h-4 w-4 mr-2" />
                Share
              </button>
              <button
                onClick={handleDownloadFile}
                disabled={isDownloading || !fileDetails._isDecrypted}
                className="btn-primary"
              >
                {isDownloading ? (
                  <>
                    <div className="h-4 w-4 spinner border-white mr-2"></div>
                    Downloading...
                  </>
                ) : (
                  <>
                    <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
                    Download
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2">
            <div className="card animate-fade-in-up">
              <div className="border-b border-gray-200">
                <nav className="-mb-px flex">
                  {tabs.map((tab) => (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id)}
                      className={`px-6 py-3 border-b-2 text-sm font-medium flex items-center transition-all duration-200 ${
                        activeTab === tab.id
                          ? "border-red-700 text-red-800"
                          : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                      }`}
                    >
                      <tab.icon className="h-4 w-4 mr-2" />
                      {tab.label}
                    </button>
                  ))}
                </nav>
              </div>

              <div className="p-6">
                {activeTab === "details" && (
                  <div className="space-y-6 animate-fade-in">
                    <h3 className="text-sm font-semibold text-gray-900 mb-4 flex items-center">
                      <DocumentIcon className="h-5 w-5 mr-2 text-gray-600" />
                      File Information
                    </h3>
                    <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div className="p-3 bg-gray-50 rounded-lg">
                        <dt className="text-sm font-medium text-gray-600">
                          Type
                        </dt>
                        <dd className="text-sm text-gray-900 mt-1">
                          {fileDetails.mime_type || "Unknown"}
                        </dd>
                      </div>
                      <div className="p-3 bg-gray-50 rounded-lg">
                        <dt className="text-sm font-medium text-gray-600">
                          Size
                        </dt>
                        <dd className="text-sm text-gray-900 mt-1">
                          {formatFileSize(fileDetails.size)}
                        </dd>
                      </div>
                      <div className="p-3 bg-gray-50 rounded-lg">
                        <dt className="text-sm font-medium text-gray-600">
                          Created
                        </dt>
                        <dd className="text-sm text-gray-900 mt-1">
                          {formatDate(fileDetails.created_at)}
                        </dd>
                      </div>
                      <div className="p-3 bg-gray-50 rounded-lg">
                        <dt className="text-sm font-medium text-gray-600">
                          Modified
                        </dt>
                        <dd className="text-sm text-gray-900 mt-1">
                          {formatDate(fileDetails.modified_at)}
                        </dd>
                      </div>
                    </dl>

                    <h3 className="text-sm font-semibold text-gray-900 mb-4 flex items-center">
                      <LockClosedIcon className="h-5 w-5 mr-2 text-gray-600" />
                      Encryption Status
                    </h3>
                    {fileDetails._isDecrypted ? (
                      <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                        <div className="flex items-start">
                          <CheckIcon className="h-5 w-5 text-green-600 mr-3 flex-shrink-0 mt-0.5" />
                          <div className="flex-1">
                            <h4 className="text-sm font-medium text-green-900">
                              File is Encrypted and Decrypted
                            </h4>
                            <p className="text-sm text-green-700 mt-1">
                              This file is protected with end-to-end encryption
                              and has been successfully decrypted for viewing.
                            </p>
                          </div>
                        </div>
                      </div>
                    ) : (
                      <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                        <div className="flex items-start">
                          <ExclamationTriangleIcon className="h-5 w-5 text-yellow-600 mr-3 flex-shrink-0 mt-0.5" />
                          <div className="flex-1">
                            <h4 className="text-sm font-medium text-yellow-900">
                              File Not Decrypted
                            </h4>
                            <p className="text-sm text-yellow-700 mt-1">
                              File content is currently encrypted. Decryption is
                              required to download or view.
                            </p>
                            {fileDetails._decryptionError && (
                              <p className="text-xs text-red-700 mt-2 font-medium">
                                <strong>Error:</strong>{" "}
                                {fileDetails._decryptionError}
                              </p>
                            )}
                          </div>
                        </div>
                      </div>
                    )}

                    <h3 className="text-sm font-semibold text-gray-900 mb-4 flex items-center">
                      <ServerIcon className="h-5 w-5 mr-2 text-gray-600" />
                      Technical Details
                    </h3>
                    <dl className="space-y-2">
                      <div className="flex justify-between py-2 border-b border-gray-100">
                        <dt className="text-sm font-medium text-gray-600">
                          File ID
                        </dt>
                        <dd className="text-sm text-gray-900 font-mono break-all">
                          {fileDetails.id}
                        </dd>
                      </div>
                      <div className="flex justify-between py-2 border-b border-gray-100">
                        <dt className="text-sm font-medium text-gray-600">
                          Collection ID
                        </dt>
                        <dd className="text-sm text-gray-900 font-mono break-all">
                          {fileDetails.collection_id}
                        </dd>
                      </div>
                      <div className="flex justify-between py-2 border-b border-gray-100">
                        <dt className="text-sm font-medium text-gray-600">
                          Version
                        </dt>
                        <dd className="text-sm text-gray-900">
                          v{fileDetails.version || 1}
                        </dd>
                      </div>
                      <div className="flex justify-between py-2">
                        <dt className="text-sm font-medium text-gray-600">
                          State
                        </dt>
                        <dd>
                          <span className="badge-success capitalize">
                            {getStateLabel(fileDetails.state)}
                          </span>
                        </dd>
                      </div>
                    </dl>
                  </div>
                )}

                {activeTab === "versions" && (
                  <div className="animate-fade-in">
                    {isLoadingVersions ? (
                      <div className="text-center py-12">
                        <div className="h-8 w-8 spinner border-red-800 mx-auto mb-4"></div>
                        <p className="text-gray-600">Loading versions...</p>
                      </div>
                    ) : (
                      <>
                        <div className="space-y-3">
                          {versionHistory.map((version) => (
                            <div
                              key={version.id}
                              className={`p-4 rounded-lg border transition-all duration-200 ${
                                version.isCurrent
                                  ? "border-red-200 bg-red-50"
                                  : "border-gray-200 hover:shadow-sm"
                              }`}
                            >
                              <div className="flex items-center justify-between">
                                <div className="flex items-center space-x-4">
                                  <div className="flex-shrink-0">
                                    <div
                                      className={`h-10 w-10 rounded-lg flex items-center justify-center ${
                                        version.isCurrent
                                          ? "bg-gradient-to-br from-red-600 to-red-700"
                                          : "bg-gray-200"
                                      }`}
                                    >
                                      <span
                                        className={`text-sm font-semibold ${
                                          version.isCurrent
                                            ? "text-white"
                                            : "text-gray-700"
                                        }`}
                                      >
                                        v{version.version}
                                      </span>
                                    </div>
                                  </div>
                                  <div>
                                    <div className="flex items-center space-x-2">
                                      <h4 className="text-sm font-medium text-gray-900">
                                        {version.name || "[Encrypted]"}
                                      </h4>
                                      {version.isCurrent && (
                                        <span className="badge-success">
                                          Current
                                        </span>
                                      )}
                                    </div>
                                    <div className="flex items-center space-x-4 mt-1 text-xs text-gray-500">
                                      <span>
                                        {formatFileSize(version.size)}
                                      </span>
                                      <span>•</span>
                                      <span>
                                        Modified by {version.modified_by}
                                      </span>
                                      <span>•</span>
                                      <span>
                                        {formatTimeAgo(version.modified_at)}
                                      </span>
                                    </div>
                                  </div>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                        {versionError && (
                          <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                            <div className="flex">
                              <InformationCircleIcon className="h-5 w-5 text-blue-600 flex-shrink-0" />
                              <div className="ml-3">
                                <h4 className="text-sm font-medium text-blue-800">
                                  Version Control
                                </h4>
                                <p className="text-sm text-blue-700 mt-1">
                                  {versionError}
                                </p>
                              </div>
                            </div>
                          </div>
                        )}
                      </>
                    )}
                  </div>
                )}

                {activeTab === "activity" && (
                  <div className="animate-fade-in">
                    <div className="text-center py-12">
                      <ChartBarIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                      <h3 className="text-base font-medium text-gray-900 mb-2">
                        Activity Log Coming Soon
                      </h3>
                      <p className="text-sm text-gray-600">
                        Track all actions performed on this file
                      </p>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            <div
              className="card p-6 animate-fade-in-up"
              style={{ animationDelay: "100ms" }}
            >
              <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
                Actions
                <SparklesIcon className="h-5 w-5 text-yellow-500 ml-2" />
              </h3>
              <div className="space-y-3">
                <button
                  onClick={handleDownloadFile}
                  disabled={isDownloading || !fileDetails._isDecrypted}
                  className="w-full btn-primary"
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
                  className="w-full btn-secondary"
                >
                  <ShareIcon className="h-4 w-4 mr-2" />
                  Share Collection
                </button>
                <button className="w-full btn-secondary" disabled>
                  <DocumentDuplicateIcon className="h-4 w-4 mr-2" />
                  Make a Copy
                </button>
                <hr className="my-3" />
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="w-full btn-danger"
                >
                  <TrashIcon className="h-4 w-4 mr-2" />
                  Delete File
                </button>
              </div>
            </div>

            <div
              className="card p-6 animate-fade-in-up"
              style={{ animationDelay: "200ms" }}
            >
              <h3 className="text-sm font-semibold text-gray-900 mb-4">
                Quick Info
              </h3>
              <dl className="space-y-3">
                <div>
                  <dt className="text-xs text-gray-600">Location</dt>
                  <dd className="text-sm text-gray-900 mt-0.5 flex items-center">
                    <FolderIcon className="h-4 w-4 mr-1 text-blue-600" />
                    Collection
                  </dd>
                </div>
                <div>
                  <dt className="text-xs text-gray-600">Encryption</dt>
                  <dd className="text-sm text-gray-900 mt-0.5 flex items-center">
                    <ShieldCheckIcon className="h-4 w-4 mr-1 text-green-600" />
                    Protected
                  </dd>
                </div>
                <div>
                  <dt className="text-xs text-gray-600">Encrypted Size</dt>
                  <dd className="text-sm text-gray-900 mt-0.5">
                    {formatFileSize(
                      fileDetails.encrypted_file_size_in_bytes || 0,
                    )}
                  </dd>
                </div>
              </dl>
            </div>
          </div>
        </div>
      </div>

      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6 animate-scale-in">
            <div className="text-center">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
                <ExclamationTriangleIcon className="h-6 w-6 text-red-600" />
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                Delete File
              </h3>
              <p className="text-sm text-gray-600 mb-6">
                Are you sure you want to delete "
                {fileDetails.name || "this file"}"? The file will be moved to
                trash and permanently deleted after 30 days.
              </p>
            </div>
            <div className="flex justify-end space-x-3">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="btn-secondary"
              >
                Cancel
              </button>
              <button onClick={handleDeleteFile} className="btn-danger">
                Delete File
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default withPasswordProtection(FileDetails);
