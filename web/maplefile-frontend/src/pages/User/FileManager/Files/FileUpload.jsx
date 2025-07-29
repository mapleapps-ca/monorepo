// File: src/pages/User/FileManager/Files/FileUpload.jsx
import React, { useState, useCallback, useEffect, useRef } from "react";
import { useNavigate, useLocation, useSearchParams } from "react-router";
import { useServices } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  CloudArrowUpIcon,
  FolderIcon,
  ArrowLeftIcon,
  DocumentIcon,
  XMarkIcon,
  PhotoIcon,
  FilmIcon,
  MusicalNoteIcon,
  DocumentTextIcon,
  ArrowUpTrayIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
} from "@heroicons/react/24/outline";

const FileUpload = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const fileInputRef = useRef(null);

  const {
    createFileManager,
    createCollectionManager,
    listCollectionManager,
    authManager,
    dashboardManager, // 🔧 NEW: Add dashboardManager for cache clearing
  } = useServices();

  const preSelectedCollectionId = searchParams.get("collection");
  const preSelectedCollectionInfo = location.state?.preSelectedCollection;

  const [fileManager, setFileManager] = useState(null);
  const [files, setFiles] = useState([]);
  const [selectedCollection, setSelectedCollection] = useState(
    preSelectedCollectionId || "",
  );
  const [availableCollections, setAvailableCollections] = useState([]);
  const [isLoadingCollections, setIsLoadingCollections] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(""); // Kept for logic, but UI will use inline status

  useEffect(() => {
    const initializeManager = async () => {
      if (!authManager.isAuthenticated()) return;

      try {
        const { default: CreateFileManager } = await import(
          "../../../../services/Manager/File/CreateFileManager.js"
        );

        const manager = new CreateFileManager(authManager);
        await manager.initialize();
        setFileManager(manager);
      } catch (err) {
        setError("Could not initialize upload service");
      }
    };

    initializeManager();
  }, [authManager]);

  const loadCollections = useCallback(async () => {
    setIsLoadingCollections(true);
    try {
      const result = await listCollectionManager.listCollections(false);
      if (result.collections && result.collections.length > 0) {
        setAvailableCollections(result.collections);
      }
    } catch (err) {
      setError("Could not load folders");
    } finally {
      setIsLoadingCollections(false);
    }
  }, [listCollectionManager]);

  useEffect(() => {
    if (createCollectionManager && listCollectionManager) {
      loadCollections();
    }
  }, [createCollectionManager, listCollectionManager, loadCollections]);

  const handleDragOver = useCallback((e) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e) => {
    e.preventDefault();
    setIsDragging(false);
    const droppedFiles = Array.from(e.dataTransfer.files);
    addFiles(droppedFiles);
  }, []);

  const handleFileSelect = (e) => {
    const selectedFiles = Array.from(e.target.files);
    addFiles(selectedFiles);
  };

  const addFiles = (newFiles) => {
    const maxSize = 5 * 1024 * 1024 * 1024; // 5GB
    const validFiles = newFiles.filter((file) => {
      if (file.size > maxSize) {
        setError(`${file.name} is too large (max 5GB)`);
        return false;
      }
      return true;
    });

    const fileObjects = validFiles.map((file) => ({
      id: Math.random().toString(36).substring(2, 11),
      file,
      name: file.name,
      size: file.size,
      type: file.type,
      status: "pending",
    }));

    setFiles((prev) => [...prev, ...fileObjects]);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const removeFile = (fileId) => {
    setFiles(files.filter((f) => f.id !== fileId));
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  const getFileIcon = (file) => {
    const iconClass = "h-5 w-5";

    if (file.type.startsWith("image/")) {
      return <PhotoIcon className={`${iconClass} text-purple-600`} />;
    }
    if (file.type.startsWith("video/")) {
      return <FilmIcon className={`${iconClass} text-pink-600`} />;
    }
    if (file.type.startsWith("audio/")) {
      return <MusicalNoteIcon className={`${iconClass} text-green-600`} />;
    }
    if (file.type.includes("pdf")) {
      return <DocumentIcon className={`${iconClass} text-red-600`} />;
    }
    if (file.type.includes("text") || file.name.endsWith(".txt")) {
      return <DocumentTextIcon className={`${iconClass} text-gray-600`} />;
    }
    return <DocumentIcon className={`${iconClass} text-blue-600`} />;
  };

  // 🔧 ENHANCED: Comprehensive cache clearing including dashboard cache
  const clearAllRelevantCaches = async () => {
    console.log("[FileUpload] Starting comprehensive cache clearing...");

    try {
      // Clear ListCollectionManager cache
      if (listCollectionManager) {
        console.log("[FileUpload] Clearing ListCollectionManager cache");
        listCollectionManager.clearAllCache();
      }

      // 🔧 NEW: Clear dashboard cache
      if (dashboardManager) {
        console.log("[FileUpload] Clearing dashboard cache");
        dashboardManager.clearAllCaches();
      }

      // Clear file-related caches from localStorage
      console.log("[FileUpload] Clearing file caches from localStorage");
      const fileListKeys = Object.keys(localStorage).filter(
        (key) =>
          key.includes("file_list") ||
          key.includes("file_cache") ||
          key.includes("mapleapps_file") ||
          key.includes("dashboard"), // 🔧 NEW: Also clear dashboard keys
      );

      fileListKeys.forEach((key) => {
        localStorage.removeItem(key);
        console.log("[FileUpload] Removed cache key:", key);
      });

      // If we have access to the services, clear their caches too
      const services = window.mapleAppsServices;
      if (services) {
        console.log(
          "[FileUpload] Clearing services caches via window.mapleAppsServices",
        );

        // Clear GetCollectionManager cache if available
        if (services.getCollectionManager) {
          services.getCollectionManager.clearAllCache();
          console.log("[FileUpload] GetCollectionManager cache cleared");
        }

        // Clear any file managers that might be cached
        if (services.getFileManager) {
          services.getFileManager.clearAllCaches();
          console.log("[FileUpload] GetFileManager cache cleared");
        }

        // 🔧 NEW: Clear dashboard manager cache
        if (services.dashboardManager) {
          services.dashboardManager.clearAllCaches();
          console.log("[FileUpload] DashboardManager cache cleared");
        }
      }

      console.log("[FileUpload] ✅ Comprehensive cache clearing completed");
    } catch (error) {
      console.warn("[FileUpload] ⚠️ Some cache clearing failed:", error);
    }
  };

  // 🔧 NEW: Trigger dashboard refresh event
  const triggerDashboardRefresh = () => {
    console.log("[FileUpload] Triggering dashboard refresh event");

    // Dispatch custom event for dashboard refresh
    const refreshEvent = new CustomEvent("dashboardRefresh", {
      detail: { reason: "file_upload_completed", timestamp: Date.now() },
    });
    window.dispatchEvent(refreshEvent);

    // Also store in localStorage to signal refresh across tabs
    const refreshSignal = {
      event: "file_upload_completed",
      timestamp: Date.now(),
      collection: selectedCollection,
      fileCount: files.length,
    };
    localStorage.setItem(
      "mapleapps_upload_refresh_signal",
      JSON.stringify(refreshSignal),
    );

    // Remove the signal after a short delay
    setTimeout(() => {
      localStorage.removeItem("mapleapps_upload_refresh_signal");
    }, 5000);
  };

  const startUpload = async () => {
    if (!fileManager || !selectedCollection || files.length === 0) {
      setError("Please select files and a folder");
      return;
    }

    setIsUploading(true);
    setError("");
    setSuccess("");

    let successCount = 0;
    const filesToUpload = files.filter((f) => f.status === "pending");

    for (const fileObj of filesToUpload) {
      try {
        setFiles((prev) =>
          prev.map((f) =>
            f.id === fileObj.id ? { ...f, status: "uploading" } : f,
          ),
        );

        await fileManager.createAndUploadFileFromFile(
          fileObj.file,
          selectedCollection,
          null,
        );

        setFiles((prev) =>
          prev.map((f) =>
            f.id === fileObj.id ? { ...f, status: "complete" } : f,
          ),
        );

        successCount++;
      } catch (err) {
        setFiles((prev) =>
          prev.map((f) =>
            f.id === fileObj.id
              ? { ...f, status: "error", error: err.message }
              : f,
          ),
        );
      }
    }

    setIsUploading(false);

    if (successCount === filesToUpload.length && successCount > 0) {
      setSuccess("All files uploaded successfully!");

      // 🔧 ENHANCED: Comprehensive cache clearing and improved navigation
      if (preSelectedCollectionId) {
        console.log(
          `[FileUpload] 📁 Upload successful - preparing redirect for ${successCount} files`,
        );

        // Clear all relevant caches immediately
        await clearAllRelevantCaches();

        // Trigger dashboard refresh event
        triggerDashboardRefresh();

        // Show success message briefly before redirect
        setTimeout(async () => {
          console.log(
            `[FileUpload] 🔄 Redirecting to collection with comprehensive refresh state`,
          );

          navigate(`/file-manager/collections/${preSelectedCollectionId}`, {
            state: {
              refresh: true,
              refreshFiles: true,
              forceFileRefresh: true,
              uploadedFileCount: successCount,
              uploadTimestamp: Date.now(),
              cacheCleared: true,
              refreshDashboard: true, // 🔧 NEW: Signal dashboard refresh
            },
            replace: false,
          });
        }, 1500);
      } else {
        // If no pre-selected collection, go to dashboard with refresh
        setTimeout(async () => {
          await clearAllRelevantCaches();
          triggerDashboardRefresh();

          navigate("/dashboard", {
            state: {
              refreshDashboard: true,
              uploadCompleted: true,
              uploadedFileCount: successCount,
              uploadTimestamp: Date.now(),
            },
            replace: false,
          });
        }, 1500);
      }
    } else {
      if (successCount > 0) {
        setSuccess(`${successCount} of ${filesToUpload.length} files uploaded`);
        // Even for partial success, trigger dashboard refresh
        await clearAllRelevantCaches();
        triggerDashboardRefresh();
      }
    }
  };

  const getBackUrl = () => {
    return preSelectedCollectionId
      ? `/file-manager/collections/${preSelectedCollectionId}`
      : "/file-manager";
  };

  const totalSize = files.reduce((sum, file) => sum + file.size, 0);
  const pendingFiles = files.filter((f) => f.status === "pending");
  const uploadingFiles = files.filter((f) => f.status === "uploading");
  const completedFiles = files.filter((f) => f.status === "complete");
  const errorFiles = files.filter((f) => f.status === "error");

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8 animate-fade-in-down">
          <button
            onClick={() => navigate(getBackUrl())}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors duration-200"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to {preSelectedCollectionInfo?.name || "My Files"}
          </button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Upload Files</h1>
            <p className="text-gray-600 mt-1">
              Drag and drop or click to select files
            </p>
          </div>
        </div>

        {/* General Error Message */}
        {error && (
          <div className="mb-6 p-3 rounded-lg bg-red-50 border border-red-200">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Upload Area */}
          <div className="lg:col-span-2 space-y-6">
            {/* Drop Zone */}
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => !isUploading && fileInputRef.current?.click()}
              className={`bg-white rounded-lg border-2 border-dashed p-12 text-center cursor-pointer transition-all duration-300 animate-fade-in-up ${
                isDragging
                  ? "border-red-500 bg-red-50 scale-105 shadow-lg"
                  : "border-gray-300 hover:border-gray-400 hover:shadow-md"
              } ${isUploading ? "opacity-50 cursor-not-allowed" : ""}`}
            >
              <div
                className={`h-16 w-16 rounded-2xl mx-auto mb-4 flex items-center justify-center transition-all duration-300 ${
                  isDragging
                    ? "bg-gradient-to-br from-red-500 to-red-600 scale-110"
                    : "bg-gradient-to-br from-gray-400 to-gray-500"
                }`}
              >
                <CloudArrowUpIcon className="h-8 w-8 text-white" />
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                {isDragging ? "Drop files here" : "Upload your files"}
              </h3>
              <p className="text-sm text-gray-600 mb-4">
                Drag and drop or click to browse
              </p>
              <p className="text-xs text-gray-500">
                Maximum file size: 5GB • All file types supported
              </p>
              <input
                ref={fileInputRef}
                type="file"
                multiple
                onChange={handleFileSelect}
                disabled={isUploading}
                className="sr-only"
              />
            </div>

            {/* Files List */}
            {files.length > 0 && (
              <div
                className="bg-white rounded-lg border border-gray-200 shadow-sm animate-fade-in-up"
                style={{ animationDelay: "100ms" }}
              >
                <div className="p-4 border-b border-gray-100">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-gray-900">
                      {files.length} file{files.length !== 1 ? "s" : ""}{" "}
                      selected
                    </h3>
                    <span className="text-sm text-gray-600">
                      Total: {formatFileSize(totalSize)}
                    </span>
                  </div>
                </div>

                <div className="divide-y divide-gray-100 max-h-96 overflow-y-auto">
                  {files.map((file) => (
                    <div
                      key={file.id}
                      className="p-4 hover:bg-gray-50 transition-colors duration-200"
                    >
                      <div className="flex items-center space-x-4">
                        <div className="flex-shrink-0">
                          <div className="h-10 w-10 rounded-lg bg-gray-50 flex items-center justify-center">
                            {getFileIcon(file)}
                          </div>
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center justify-between mb-1">
                            <h4 className="text-sm font-medium text-gray-900 truncate pr-2">
                              {file.name}
                            </h4>
                            <span className="text-xs text-gray-500 flex-shrink-0">
                              {formatFileSize(file.size)}
                            </span>
                          </div>
                          {file.status === "error" && file.error && (
                            <p className="text-xs text-red-600 mt-1 truncate">
                              {file.error}
                            </p>
                          )}
                        </div>
                        <div className="flex-shrink-0">
                          {file.status === "pending" && (
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                removeFile(file.id);
                              }}
                              disabled={isUploading}
                              className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all duration-200"
                            >
                              <XMarkIcon className="h-5 w-5" />
                            </button>
                          )}
                          {file.status === "uploading" && (
                            <div className="p-2">
                              <div className="animate-spin rounded-full h-5 w-5 border-2 border-red-700 border-t-transparent"></div>
                            </div>
                          )}
                          {file.status === "complete" && (
                            <div className="p-2">
                              <CheckCircleIcon className="h-5 w-5 text-green-600" />
                            </div>
                          )}
                          {file.status === "error" && (
                            <div className="p-2">
                              <ExclamationTriangleIcon className="h-5 w-5 text-red-600" />
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            <div
              className="bg-white rounded-lg border border-gray-200 shadow-sm p-6 animate-fade-in-up"
              style={{ animationDelay: "200ms" }}
            >
              <h3 className="font-semibold text-gray-900 mb-4 flex items-center">
                <FolderIcon className="h-5 w-5 mr-2 text-gray-600" />
                Upload Destination
              </h3>

              {preSelectedCollectionId ? (
                <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
                  <div className="flex items-center space-x-3">
                    <div className="h-10 w-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center">
                      <FolderIcon className="h-5 w-5 text-white" />
                    </div>
                    <div>
                      <p className="text-sm font-medium text-blue-900">
                        {preSelectedCollectionInfo?.name || "Selected Folder"}
                      </p>
                      <p className="text-xs text-blue-700">
                        Files will be uploaded here
                      </p>
                    </div>
                  </div>
                </div>
              ) : (
                <select
                  value={selectedCollection}
                  onChange={(e) => setSelectedCollection(e.target.value)}
                  disabled={isLoadingCollections || isUploading}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                >
                  <option value="">Select a folder...</option>
                  {availableCollections.map((collection) => (
                    <option key={collection.id} value={collection.id}>
                      {collection.name || "Unnamed Folder"}
                    </option>
                  ))}
                </select>
              )}
            </div>

            {files.length > 0 && (
              <div
                className="bg-white rounded-lg border border-gray-200 shadow-sm p-6 animate-fade-in-up"
                style={{ animationDelay: "300ms" }}
              >
                <h3 className="font-semibold text-gray-900 mb-4">
                  Upload Status
                </h3>
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">Pending</span>
                    <span className="text-sm font-medium text-gray-900">
                      {pendingFiles.length}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">Uploading</span>
                    <span className="text-sm font-medium text-blue-600">
                      {uploadingFiles.length}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">Completed</span>
                    <span className="text-sm font-medium text-green-600">
                      {completedFiles.length}
                    </span>
                  </div>
                  {errorFiles.length > 0 && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Errors</span>
                      <span className="text-sm font-medium text-red-600">
                        {errorFiles.length}
                      </span>
                    </div>
                  )}
                </div>
                {completedFiles.length + errorFiles.length === files.length &&
                  files.length > 0 &&
                  !isUploading &&
                  pendingFiles.length === 0 && (
                    <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-lg">
                      <p className="text-sm text-green-800 flex items-center">
                        <CheckCircleIcon className="h-4 w-4 mr-2" />
                        All uploads processed!
                      </p>
                    </div>
                  )}
              </div>
            )}

            <button
              onClick={startUpload}
              disabled={
                !selectedCollection ||
                !fileManager ||
                files.length === 0 ||
                isUploading ||
                pendingFiles.length === 0
              }
              className="w-full flex justify-center items-center py-3 px-4 rounded-lg text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400 text-base font-semibold animate-fade-in-up"
              style={{ animationDelay: "400ms" }}
            >
              {isUploading ? (
                <>
                  <div className="animate-spin rounded-full h-5 w-5 border-2 border-white border-t-transparent mr-2"></div>
                  Uploading {uploadingFiles.length} of {files.length}...
                </>
              ) : (
                <>
                  <ArrowUpTrayIcon className="h-5 w-5 mr-2" />
                  Upload {pendingFiles.length} File
                  {pendingFiles.length !== 1 ? "s" : ""}
                </>
              )}
            </button>

            <div
              className="p-4 bg-blue-50 rounded-lg border border-blue-200 animate-fade-in-up"
              style={{ animationDelay: "500ms" }}
            >
              <h4 className="text-sm font-medium text-blue-900 mb-2 flex items-center">
                <ExclamationTriangleIcon className="h-4 w-4 mr-2" />
                Security Notice
              </h4>
              <p className="text-xs text-blue-800">
                All files are encrypted end-to-end before upload. Only you and
                those you share with can decrypt them.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(FileUpload);
