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
} from "@heroicons/react/24/outline";

const FileDetails = () => {
  const navigate = useNavigate();
  const { fileId } = useParams();

  const {
    getFileManager,
    downloadFileManager,
    deleteFileManager,
    listCollectionManager,
  } = useFiles();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [fileDetails, setFileDetails] = useState(null);
  const [versionHistory, setVersionHistory] = useState([]);
  const [activeTab, setActiveTab] = useState("details");
  const [isDownloading, setIsDownloading] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  // NEW: Version history specific states
  const [isLoadingVersions, setIsLoadingVersions] = useState(false);
  const [versionError, setVersionError] = useState("");
  const [hasLoadedVersions, setHasLoadedVersions] = useState(false);

  // Load file details
  const loadFileDetails = useCallback(
    async (forceRefresh = false) => {
      if (!getFileManager || !fileId) return;

      setIsLoading(true);
      setError("");

      try {
        console.log("[FileDetails] Loading file details for:", fileId);
        // Get file details
        const file = await getFileManager.getFileById(fileId, forceRefresh);
        setFileDetails(file);
        console.log(
          "[FileDetails] File details loaded:",
          file.name || "[Encrypted]",
        );
      } catch (err) {
        setError("Could not load file details");
        console.error("[FileDetails] Failed to load file details:", err);
      } finally {
        setIsLoading(false);
      }
    },
    [getFileManager, fileId],
  );

  // NEW: Load version history separately with better error handling
  const loadVersionHistory = useCallback(
    async (forceRefresh = false) => {
      if (!fileDetails) {
        console.log(
          "[FileDetails] No file details available for version history",
        );
        return;
      }

      setIsLoadingVersions(true);
      setVersionError("");

      try {
        console.log("[FileDetails] Creating mock version history for:", fileId);

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
        };

        setVersionHistory([mockVersion]);
        setHasLoadedVersions(true);

        // Show informational message about version history not being implemented
        setVersionError(
          "Version history is not yet implemented in the backend. Showing current file version only.",
        );

        console.log("[FileDetails] Mock version history created successfully");
      } catch (err) {
        console.warn(
          "[FileDetails] Failed to create mock version history:",
          err,
        );
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

  // NEW: Load version history when switching to versions tab
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

  // Handle file download
  const handleDownloadFile = async () => {
    if (!downloadFileManager || !fileDetails) return;

    setIsDownloading(true);
    setError("");

    try {
      console.log("[FileDetails] Starting download for:", fileId);
      await downloadFileManager.downloadFile(fileId, {
        saveToDisk: true,
      });
      console.log("[FileDetails] Download completed successfully");
    } catch (err) {
      setError("Could not download file");
      console.error("[FileDetails] Failed to download file:", err);
    } finally {
      setIsDownloading(false);
    }
  };

  // Handle file deletion
  const handleDeleteFile = async () => {
    if (!deleteFileManager || !fileId) return;

    try {
      console.log("[FileDetails] Deleting file:", fileId);
      await deleteFileManager.deleteFile(fileId, false); // Move to trash
      navigate(`/file-manager/collections/${fileDetails?.collection_id}`);
    } catch (err) {
      setError("Could not delete file");
      console.error("[FileDetails] Failed to delete file:", err);
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
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "Unknown";
    try {
      return new Date(dateString).toLocaleString();
    } catch {
      return "Invalid Date";
    }
  };

  // NEW: Get state color for version display
  const getStateColor = (state) => {
    switch (state) {
      case "active":
        return "bg-green-500";
      case "archived":
        return "bg-gray-500";
      case "deleted":
        return "bg-red-500";
      case "pending":
        return "bg-yellow-500";
      default:
        return "bg-gray-500";
    }
  };

  // NEW: Get state label
  const getStateLabel = (state) => {
    switch (state) {
      case "active":
        return "Active";
      case "archived":
        return "Archived";
      case "deleted":
        return "Deleted";
      case "pending":
        return "Pending";
      default:
        return state
          ? state.charAt(0).toUpperCase() + state.slice(1)
          : "Unknown";
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
                <div className="flex items-center space-x-4 mt-1">
                  <p className="text-sm text-gray-500">
                    {formatFileSize(
                      fileDetails.size ||
                        fileDetails.encrypted_file_size_in_bytes,
                    )}{" "}
                    • Modified {formatDate(fileDetails.modified_at)}
                  </p>
                  {/* NEW: Show decryption status */}
                  {fileDetails._isDecrypted ? (
                    <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                      <CheckIcon className="h-3 w-3 mr-1" />
                      Decrypted
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                      <ExclamationTriangleIcon className="h-3 w-3 mr-1" />
                      Encrypted
                    </span>
                  )}
                  {/* NEW: Show file state */}
                  <span
                    className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium text-white ${getStateColor(fileDetails.state)}`}
                  >
                    {getStateLabel(fileDetails.state)}
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
                    {versionHistory.length > 0 && (
                      <span className="ml-2 px-2 py-1 text-xs bg-gray-100 text-gray-600 rounded-full">
                        {versionHistory.length}
                      </span>
                    )}
                  </button>
                </nav>
              </div>

              <div className="p-6">
                {/* Details Tab */}
                {activeTab === "details" && (
                  <div className="space-y-6">
                    <div>
                      <h3 className="text-sm font-medium text-gray-900 mb-4">
                        File Information
                      </h3>
                      <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        <div>
                          <dt className="text-sm font-medium text-gray-500">
                            Type
                          </dt>
                          <dd className="text-sm text-gray-900 mt-1">
                            {fileDetails.mime_type || "Unknown"}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500">
                            Size
                          </dt>
                          <dd className="text-sm text-gray-900 mt-1">
                            {formatFileSize(
                              fileDetails.size ||
                                fileDetails.encrypted_file_size_in_bytes,
                            )}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500">
                            Created
                          </dt>
                          <dd className="text-sm text-gray-900 mt-1">
                            {formatDate(fileDetails.created_at)}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500">
                            Modified
                          </dt>
                          <dd className="text-sm text-gray-900 mt-1">
                            {formatDate(fileDetails.modified_at)}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500">
                            Version
                          </dt>
                          <dd className="text-sm text-gray-900 mt-1">
                            v{fileDetails.version || 1}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500">
                            State
                          </dt>
                          <dd className="text-sm text-gray-900 mt-1 capitalize">
                            {getStateLabel(fileDetails.state)}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500">
                            File ID
                          </dt>
                          <dd className="text-sm text-gray-900 mt-1 font-mono">
                            {fileDetails.id}
                          </dd>
                        </div>
                        <div>
                          <dt className="text-sm font-medium text-gray-500">
                            Collection ID
                          </dt>
                          <dd className="text-sm text-gray-900 mt-1 font-mono">
                            {fileDetails.collection_id}
                          </dd>
                        </div>
                      </dl>
                    </div>

                    {/* NEW: Enhanced encryption status */}
                    <div>
                      <h3 className="text-sm font-medium text-gray-900 mb-4">
                        Encryption Status
                      </h3>
                      <div className="space-y-3">
                        <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                          <span className="text-sm font-medium text-gray-700">
                            Decrypted
                          </span>
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                              fileDetails._isDecrypted
                                ? "bg-green-100 text-green-800"
                                : "bg-red-100 text-red-800"
                            }`}
                          >
                            {fileDetails._isDecrypted ? (
                              <>
                                <CheckIcon className="h-3 w-3 mr-1" />
                                Yes
                              </>
                            ) : (
                              <>
                                <XMarkIcon className="h-3 w-3 mr-1" />
                                No
                              </>
                            )}
                          </span>
                        </div>

                        <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                          <span className="text-sm font-medium text-gray-700">
                            File Key Available
                          </span>
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                              fileDetails._hasFileKey
                                ? "bg-green-100 text-green-800"
                                : "bg-yellow-100 text-yellow-800"
                            }`}
                          >
                            {fileDetails._hasFileKey ? (
                              <>
                                <CheckIcon className="h-3 w-3 mr-1" />
                                Yes
                              </>
                            ) : (
                              <>
                                <ClockIcon className="h-3 w-3 mr-1" />
                                Cached
                              </>
                            )}
                          </span>
                        </div>

                        {fileDetails._decryptionError && (
                          <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                            <p className="text-sm text-yellow-800">
                              <strong>Decryption Error:</strong>{" "}
                              {fileDetails._decryptionError}
                            </p>
                            <p className="text-xs text-yellow-700 mt-1">
                              You may not have access to this file or the
                              encryption keys may be missing.
                            </p>
                          </div>
                        )}
                      </div>
                    </div>

                    {/* NEW: Tombstone information if applicable */}
                    {fileDetails._has_tombstone && (
                      <div>
                        <h3 className="text-sm font-medium text-gray-900 mb-4">
                          Tombstone Information
                        </h3>
                        <div className="p-3 bg-orange-50 border border-orange-200 rounded-lg">
                          <dl className="space-y-2">
                            <div className="flex justify-between">
                              <dt className="text-sm font-medium text-orange-800">
                                Tombstone Version
                              </dt>
                              <dd className="text-sm text-orange-900">
                                v{fileDetails.tombstone_version}
                              </dd>
                            </div>
                            <div className="flex justify-between">
                              <dt className="text-sm font-medium text-orange-800">
                                Expiry Date
                              </dt>
                              <dd className="text-sm text-orange-900">
                                {formatDate(fileDetails.tombstone_expiry)}
                              </dd>
                            </div>
                            <div className="flex justify-between">
                              <dt className="text-sm font-medium text-orange-800">
                                Status
                              </dt>
                              <dd className="text-sm text-orange-900">
                                {fileDetails._tombstone_expired
                                  ? "Expired"
                                  : "Active"}
                              </dd>
                            </div>
                          </dl>
                        </div>
                      </div>
                    )}
                  </div>
                )}

                {/* NEW: Enhanced Versions Tab */}
                {activeTab === "versions" && (
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <h3 className="text-sm font-medium text-gray-900">
                        Version History
                      </h3>
                      <button
                        onClick={() => loadVersionHistory(true)}
                        disabled={isLoadingVersions}
                        className="inline-flex items-center px-3 py-1 border border-gray-300 rounded text-sm text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                      >
                        <ArrowPathIcon
                          className={`h-4 w-4 mr-1 ${isLoadingVersions ? "animate-spin" : ""}`}
                        />
                        Refresh
                      </button>
                    </div>

                    {/* Version History Information */}
                    {versionError && (
                      <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
                        <div className="flex">
                          <div className="h-5 w-5 text-blue-400 mr-2 flex-shrink-0 mt-0.5">
                            ℹ️
                          </div>
                          <div>
                            <h4 className="text-sm font-medium text-blue-800">
                              Version History
                            </h4>
                            <p className="text-sm text-blue-700 mt-1">
                              {versionError}
                            </p>
                            <p className="text-xs text-blue-600 mt-2">
                              💡 The backend currently doesn't support the{" "}
                              <code className="bg-blue-100 px-1 rounded">
                                GET /files/{`{id}`}/versions
                              </code>{" "}
                              endpoint. Only the current version is shown.
                            </p>
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Loading State */}
                    {isLoadingVersions && (
                      <div className="flex items-center justify-center py-8">
                        <div className="text-center">
                          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-red-800 mx-auto mb-2"></div>
                          <p className="text-sm text-gray-600">
                            Loading version history...
                          </p>
                        </div>
                      </div>
                    )}

                    {/* Version History Table */}
                    {!isLoadingVersions &&
                      versionHistory.length === 0 &&
                      !versionError && (
                        <div className="text-center py-8">
                          <ClockIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
                          <p className="text-gray-500">
                            No version history available
                          </p>
                          <p className="text-sm text-gray-400 mt-1">
                            This file may only have one version or version
                            history is not enabled.
                          </p>
                        </div>
                      )}

                    {!isLoadingVersions && versionHistory.length > 0 && (
                      <div>
                        {versionError && versionHistory.length === 1 && (
                          <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
                            <p className="text-sm text-blue-800">
                              ℹ️ Showing current file information as a single
                              version entry.
                            </p>
                          </div>
                        )}

                        <div className="overflow-hidden border border-gray-200 rounded-lg">
                          <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                              <tr>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Version
                                </th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Name
                                </th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  State
                                </th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Size
                                </th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Modified
                                </th>
                                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                  Status
                                </th>
                              </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                              {versionHistory.map((version, index) => (
                                <tr
                                  key={`${version.id}-${version.version}`}
                                  className={index === 0 ? "bg-green-50" : ""}
                                >
                                  <td className="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">
                                    v{version.version}
                                    {index === 0 && (
                                      <span className="ml-2 inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                        Current
                                      </span>
                                    )}
                                  </td>
                                  <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                                    {version.name || "[Unable to decrypt]"}
                                  </td>
                                  <td className="px-4 py-3 whitespace-nowrap">
                                    <span
                                      className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium text-white ${getStateColor(version.state)}`}
                                    >
                                      {getStateLabel(version.state)}
                                    </span>
                                  </td>
                                  <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                                    {formatFileSize(
                                      version.size ||
                                        version.encrypted_file_size_in_bytes,
                                    )}
                                  </td>
                                  <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                                    {formatDate(version.modified_at)}
                                  </td>
                                  <td className="px-4 py-3 whitespace-nowrap">
                                    <span
                                      className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                                        version._isDecrypted
                                          ? "bg-green-100 text-green-800"
                                          : "bg-red-100 text-red-800"
                                      }`}
                                    >
                                      {version._isDecrypted ? (
                                        <>
                                          <CheckIcon className="h-3 w-3 mr-1" />
                                          Decrypted
                                        </>
                                      ) : (
                                        <>
                                          <XMarkIcon className="h-3 w-3 mr-1" />
                                          Encrypted
                                        </>
                                      )}
                                    </span>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
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
                  className="w-full inline-flex items-center justify-center px-4 py-2 rounded-lg text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400 disabled:cursor-not-allowed"
                >
                  <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
                  {isDownloading ? "Downloading..." : "Download"}
                </button>

                {!fileDetails._isDecrypted && (
                  <p className="text-xs text-gray-500 text-center">
                    File must be decrypted to download
                  </p>
                )}

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

              {/* NEW: File metadata in sidebar */}
              <div className="mt-6 pt-6 border-t border-gray-200">
                <h4 className="text-sm font-medium text-gray-900 mb-3">
                  Quick Info
                </h4>
                <dl className="space-y-2">
                  <div>
                    <dt className="text-xs text-gray-500">File ID</dt>
                    <dd className="text-xs text-gray-900 font-mono break-all">
                      {fileDetails.id}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-xs text-gray-500">Encrypted Size</dt>
                    <dd className="text-xs text-gray-900">
                      {formatFileSize(
                        fileDetails.encrypted_file_size_in_bytes || 0,
                      )}
                    </dd>
                  </div>
                  <div>
                    <dt className="text-xs text-gray-500">Created</dt>
                    <dd className="text-xs text-gray-900">
                      {formatDate(fileDetails.created_at)}
                    </dd>
                  </div>
                </dl>
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
              Are you sure you want to delete "{fileDetails.name || "this file"}
              "? The file will be moved to trash and permanently deleted after
              30 days.
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
