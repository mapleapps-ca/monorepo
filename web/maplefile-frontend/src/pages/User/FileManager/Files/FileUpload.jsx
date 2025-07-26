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
  CheckIcon,
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
  const [success, setSuccess] = useState("");

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

  useEffect(() => {
    if (createCollectionManager && listCollectionManager) {
      loadCollections();
    }
  }, [createCollectionManager, listCollectionManager]);

  const loadCollections = async () => {
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
  };

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
      id: Math.random().toString(36).substr(2, 9),
      file,
      name: file.name,
      size: file.size,
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

    for (const fileObj of files) {
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

    if (successCount === files.length) {
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
        }, 1000);
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
        }, 1000);
      }
    } else {
      setSuccess(`${successCount} of ${files.length} files uploaded`);

      // Even for partial success, trigger dashboard refresh
      if (successCount > 0) {
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

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => navigate(getBackUrl())}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to {preSelectedCollectionInfo?.name || "My Files"}
          </button>
          <h1 className="text-2xl font-semibold text-gray-900">Upload Files</h1>
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

        {/* Upload Area */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
            {/* Drop Zone */}
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => !isUploading && fileInputRef.current?.click()}
              className={`bg-white rounded-lg border-2 border-dashed p-8 text-center cursor-pointer transition-all ${
                isDragging
                  ? "border-red-800 bg-red-50"
                  : "border-gray-300 hover:border-gray-400"
              } ${isUploading ? "opacity-50 cursor-not-allowed" : ""}`}
            >
              <CloudArrowUpIcon
                className={`mx-auto h-12 w-12 mb-3 ${
                  isDragging ? "text-red-800" : "text-gray-400"
                }`}
              />
              <h3 className="text-sm font-medium text-gray-900 mb-1">
                {isDragging ? "Drop files here" : "Click or drag files"}
              </h3>
              <p className="text-xs text-gray-500">Max 5GB per file</p>
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
              <div className="mt-6 bg-white rounded-lg border border-gray-200">
                <div className="p-4 border-b border-gray-200">
                  <h3 className="font-medium text-gray-900">
                    {files.length} file{files.length !== 1 ? "s" : ""} selected
                  </h3>
                </div>
                <div className="divide-y divide-gray-200 max-h-64 overflow-y-auto">
                  {files.map((file) => (
                    <div
                      key={file.id}
                      className="p-4 flex items-center justify-between"
                    >
                      <div className="flex items-center space-x-3 flex-1 min-w-0">
                        <DocumentIcon className="h-5 w-5 text-gray-400 flex-shrink-0" />
                        <div className="min-w-0">
                          <p className="text-sm font-medium text-gray-900 truncate">
                            {file.name}
                          </p>
                          <p className="text-xs text-gray-500">
                            {formatFileSize(file.size)}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center">
                        {file.status === "pending" && (
                          <button
                            onClick={() => removeFile(file.id)}
                            disabled={isUploading}
                            className="text-gray-400 hover:text-red-600"
                          >
                            <XMarkIcon className="h-5 w-5" />
                          </button>
                        )}
                        {file.status === "uploading" && (
                          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-red-800"></div>
                        )}
                        {file.status === "complete" && (
                          <CheckIcon className="h-5 w-5 text-green-600" />
                        )}
                        {file.status === "error" && (
                          <XMarkIcon className="h-5 w-5 text-red-600" />
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Sidebar */}
          <div>
            {/* Folder Selection */}
            <div className="bg-white rounded-lg border border-gray-200 p-4 mb-4">
              <h3 className="font-medium text-gray-900 mb-3">Upload to</h3>

              {preSelectedCollectionId ? (
                <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
                  <div className="flex items-center space-x-2">
                    <FolderIcon className="h-5 w-5 text-blue-600" />
                    <span className="text-sm font-medium text-blue-900">
                      {preSelectedCollectionInfo?.name || "Selected Folder"}
                    </span>
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

            {/* Upload Button */}
            <button
              onClick={startUpload}
              disabled={
                !selectedCollection ||
                !fileManager ||
                files.length === 0 ||
                isUploading
              }
              className="w-full py-2 px-4 rounded-lg text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400"
            >
              {isUploading ? "Uploading..." : "Upload Files"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(FileUpload);
