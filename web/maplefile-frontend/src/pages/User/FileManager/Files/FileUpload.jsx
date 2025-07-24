// File: src/pages/User/FileManager/Files/FileUpload.jsx
import React, { useState, useCallback, useEffect, useRef } from "react";
import { Link, useNavigate } from "react-router";
import { useServices } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  CloudArrowUpIcon,
  FolderIcon,
  ArrowLeftIcon,
  InformationCircleIcon,
  ShieldCheckIcon,
  XMarkIcon,
  CheckIcon,
  DocumentIcon,
  PhotoIcon,
  VideoCameraIcon,
  MusicalNoteIcon,
  DocumentTextIcon,
  PaperClipIcon,
  SparklesIcon,
  ChevronRightIcon,
  HomeIcon,
  ArrowPathIcon,
  LockClosedIcon,
  ExclamationTriangleIcon,
  KeyIcon,
  EyeIcon,
  EyeSlashIcon,
} from "@heroicons/react/24/outline";

const FileUpload = () => {
  const navigate = useNavigate();
  const fileInputRef = useRef(null);

  // Get services from context
  const {
    createFileManager,
    createCollectionManager,
    listCollectionManager,
    authManager,
  } = useServices();

  // State management
  const [fileManager, setFileManager] = useState(null);
  const [files, setFiles] = useState([]);
  const [selectedCollection, setSelectedCollection] = useState("");
  const [availableCollections, setAvailableCollections] = useState([]);
  const [isLoadingCollections, setIsLoadingCollections] = useState(false);
  const [uploadProgress, setUploadProgress] = useState({});
  const [isUploading, setIsUploading] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [eventLog, setEventLog] = useState([]);

  // Upload options
  const [encryptFiles, setEncryptFiles] = useState(true);
  const [generateThumbnails, setGenerateThumbnails] = useState(false);
  const [skipDuplicates, setSkipDuplicates] = useState(true);

  // Initialize file manager
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

        // Set up event listener
        manager.addFileCreationListener(handleFileEvent);

        console.log("[FileUpload] CreateFileManager initialized");
      } catch (err) {
        console.error(
          "[FileUpload] Failed to initialize CreateFileManager:",
          err,
        );
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();

    return () => {
      if (fileManager) {
        fileManager.removeFileCreationListener(handleFileEvent);
      }
    };
  }, [authManager]);

  // Load collections when managers are ready
  useEffect(() => {
    if (createCollectionManager && listCollectionManager) {
      loadCollections();
    }
  }, [createCollectionManager, listCollectionManager]);

  // Handle file events
  const handleFileEvent = useCallback((eventType, eventData) => {
    console.log("[FileUpload] File event:", eventType, eventData);
    addToEventLog(eventType, eventData);

    // Update file status based on events
    if (eventType === "file_upload_started") {
      setFiles((prev) =>
        prev.map((f) =>
          f.id === eventData.tempId
            ? { ...f, status: "uploading", fileId: eventData.fileId }
            : f,
        ),
      );
    } else if (eventType === "file_upload_completed") {
      setFiles((prev) =>
        prev.map((f) =>
          f.fileId === eventData.fileId
            ? { ...f, status: "complete", progress: 100 }
            : f,
        ),
      );
    } else if (eventType === "file_upload_failed") {
      setFiles((prev) =>
        prev.map((f) =>
          f.fileId === eventData.fileId
            ? { ...f, status: "error", error: eventData.error }
            : f,
        ),
      );
    }
  }, []);

  // Load collections
  const loadCollections = async () => {
    setIsLoadingCollections(true);
    setError("");

    try {
      console.log("[FileUpload] Loading existing collections...");

      // Try to list existing collections first
      const result = await listCollectionManager.listCollections(false);

      if (result.collections && result.collections.length > 0) {
        // Use existing collections
        setAvailableCollections(result.collections);
        console.log(
          "[FileUpload] Loaded collections:",
          result.collections.length,
        );
      } else {
        // Create test collections if none exist
        console.log(
          "[FileUpload] No collections found, creating test collections...",
        );
        await createTestCollections();
      }
    } catch (err) {
      console.error("[FileUpload] Failed to load collections:", err);
      // Try to create test collections as fallback
      try {
        await createTestCollections();
      } catch (createErr) {
        console.error(
          "[FileUpload] Failed to create test collections:",
          createErr,
        );
        setError(`Failed to load or create collections: ${createErr.message}`);
      }
    } finally {
      setIsLoadingCollections(false);
    }
  };

  // Create test collections for the example
  const createTestCollections = async () => {
    try {
      console.log("[FileUpload] Creating test collections...");

      const testCollections = [
        { name: "Work Documents", collection_type: "folder" },
        { name: "Vacation Photos", collection_type: "album" },
        { name: "Project Files", collection_type: "folder" },
        { name: "Family Album", collection_type: "album" },
      ];

      const createdCollections = [];

      for (const collectionData of testCollections) {
        try {
          const result =
            await createCollectionManager.createCollection(collectionData);
          createdCollections.push(result.collection);
          console.log(
            "[FileUpload] Created test collection:",
            result.collection.id,
          );
        } catch (err) {
          console.warn(
            "[FileUpload] Failed to create test collection:",
            err.message,
          );
        }
      }

      if (createdCollections.length > 0) {
        setAvailableCollections(createdCollections);
        console.log("[FileUpload] Test collections created successfully");
      } else {
        throw new Error("Failed to create any test collections");
      }
    } catch (err) {
      console.error("[FileUpload] Failed to create test collections:", err);
      throw err;
    }
  };

  // Add event to log
  const addToEventLog = (eventType, eventData) => {
    setEventLog((prev) => [
      ...prev,
      {
        timestamp: new Date().toISOString(),
        eventType,
        eventData,
      },
    ]);
  };

  // Handle drag events
  const handleDragOver = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const droppedFiles = Array.from(e.dataTransfer.files);
    addFiles(droppedFiles);
  }, []);

  const handleFileSelect = (e) => {
    const selectedFiles = Array.from(e.target.files);
    addFiles(selectedFiles);
  };

  const addFiles = (newFiles) => {
    // Basic validation
    const maxSize = 5 * 1024 * 1024 * 1024; // 5GB limit
    const validFiles = newFiles.filter((file) => {
      if (file.size > maxSize) {
        setError(`File "${file.name}" exceeds 5GB limit`);
        return false;
      }
      return true;
    });

    const fileObjects = validFiles.map((file) => ({
      id: Math.random().toString(36).substr(2, 9),
      file,
      name: file.name,
      size: file.size,
      type: file.type,
      status: "pending",
      progress: 0,
      fileId: null,
      error: null,
    }));

    setFiles((prev) => [...prev, ...fileObjects]);

    // Clear file input
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  const removeFile = (fileId) => {
    setFiles(files.filter((f) => f.id !== fileId));
  };

  const getFileIcon = (type) => {
    if (type.startsWith("image/")) return <PhotoIcon className="h-5 w-5" />;
    if (type.startsWith("video/"))
      return <VideoCameraIcon className="h-5 w-5" />;
    if (type.startsWith("audio/"))
      return <MusicalNoteIcon className="h-5 w-5" />;
    if (type.includes("pdf")) return <DocumentTextIcon className="h-5 w-5" />;
    return <DocumentIcon className="h-5 w-5" />;
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Get password from storage
  const handleGetStoredPassword = async () => {
    try {
      const { default: passwordStorageService } = await import(
        "../../../../services/PasswordStorageService.js"
      );
      const storedPassword = passwordStorageService.getPassword();

      if (storedPassword) {
        setPassword(storedPassword);
        setSuccess("Password loaded from storage");
      } else {
        setError("No password found in storage");
      }
    } catch (err) {
      setError(`Failed to get stored password: ${err.message}`);
    }
  };

  // Start upload process
  const startUpload = async () => {
    if (!fileManager || !selectedCollection || files.length === 0) {
      setError("Please select files and a collection");
      return;
    }

    // Check if the selected collection exists
    const selectedCol = availableCollections.find(
      (col) => col.id === selectedCollection,
    );
    if (!selectedCol) {
      setError("Selected collection not found. Please refresh collections.");
      return;
    }

    setIsUploading(true);
    setError("");
    setSuccess("");

    try {
      console.log("[FileUpload] Starting upload for", files.length, "files");

      // Upload files one by one
      for (const fileObj of files) {
        try {
          // Update file status
          setFiles((prev) =>
            prev.map((f) =>
              f.id === fileObj.id
                ? { ...f, status: "uploading", progress: 0 }
                : f,
            ),
          );

          // Upload file using CreateFileManager
          const result = await fileManager.createAndUploadFileFromFile(
            fileObj.file,
            selectedCollection,
            password || null,
          );

          // Update file with success
          setFiles((prev) =>
            prev.map((f) =>
              f.id === fileObj.id
                ? {
                    ...f,
                    status: "complete",
                    progress: 100,
                    fileId: result.fileId,
                  }
                : f,
            ),
          );

          console.log(
            "[FileUpload] File uploaded successfully:",
            result.fileId,
          );
        } catch (err) {
          console.error("[FileUpload] Failed to upload file:", err);

          // Update file with error
          setFiles((prev) =>
            prev.map((f) =>
              f.id === fileObj.id
                ? { ...f, status: "error", error: err.message }
                : f,
            ),
          );
        }
      }

      const successCount = files.filter((f) => f.status === "complete").length;
      const totalCount = files.length;

      if (successCount === totalCount) {
        setSuccess(`All ${totalCount} files uploaded successfully!`);
      } else {
        setSuccess(
          `${successCount} of ${totalCount} files uploaded successfully`,
        );
      }
    } catch (err) {
      console.error("[FileUpload] Upload process failed:", err);
      setError(err.message);
    } finally {
      setIsUploading(false);
    }
  };

  // Clear messages
  const clearMessages = () => {
    setError("");
    setSuccess("");
  };

  // Auto-clear messages
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(clearMessages, 5000);
      return () => clearTimeout(timer);
    }
  }, [success, error]);

  const totalSize = files.reduce((acc, file) => acc + file.size, 0);
  const completedFiles = files.filter((f) => f.status === "complete").length;
  const errorFiles = files.filter((f) => f.status === "error").length;

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Breadcrumb */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 mb-6">
          <HomeIcon className="h-4 w-4" />
          <ChevronRightIcon className="h-3 w-3" />
          <Link to="/file-manager" className="hover:text-gray-900">
            My Files
          </Link>
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900">Upload Files</span>
        </div>

        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => navigate("/file-manager")}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors duration-200"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to Collections
          </button>

          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Upload Files
          </h1>
          <p className="text-gray-600">
            Securely upload and encrypt your files with end-to-end encryption
          </p>
        </div>

        {/* Error/Success Messages */}
        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 animate-fade-in">
            <div className="flex items-center">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-red-800">Error</h3>
                <p className="text-sm text-red-700 mt-1">{error}</p>
              </div>
              <button
                onClick={clearMessages}
                className="text-red-500 hover:text-red-700 transition-colors duration-200"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}

        {success && (
          <div className="mb-6 p-4 rounded-lg bg-green-50 border border-green-200 animate-fade-in">
            <div className="flex items-center">
              <CheckIcon className="h-5 w-5 text-green-500 mr-3 flex-shrink-0" />
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-green-800">
                  Success
                </h3>
                <p className="text-sm text-green-700 mt-1">{success}</p>
              </div>
              <button
                onClick={clearMessages}
                className="text-green-500 hover:text-green-700 transition-colors duration-200"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}

        {/* Main Content */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Upload Area */}
          <div className="lg:col-span-2 space-y-6">
            {/* Drag and Drop Zone */}
            <div
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
              onClick={() => !isUploading && fileInputRef.current?.click()}
              className={`bg-white rounded-xl border-2 border-dashed transition-all duration-200 cursor-pointer ${
                isDragging
                  ? "border-red-500 bg-red-50"
                  : isUploading
                    ? "border-gray-300 cursor-not-allowed opacity-50"
                    : "border-gray-300 hover:border-red-400 hover:bg-red-50"
              }`}
            >
              <div className="p-12 text-center">
                <CloudArrowUpIcon
                  className={`mx-auto h-16 w-16 mb-4 transition-colors duration-200 ${
                    isDragging
                      ? "text-red-600"
                      : isUploading
                        ? "text-gray-400"
                        : "text-gray-400"
                  }`}
                />
                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                  {isDragging
                    ? "Drop files here"
                    : isUploading
                      ? "Upload in progress..."
                      : "Drag and drop files here"}
                </h3>
                {!isUploading && (
                  <>
                    <p className="text-sm text-gray-500 mb-4">or</p>
                    <div className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-all duration-200">
                      <PaperClipIcon className="h-4 w-4 mr-2" />
                      Browse Files
                    </div>
                  </>
                )}
                <p className="text-xs text-gray-500 mt-4">
                  Maximum file size: 5GB • Supported formats: All file types
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
            </div>

            {/* Files List */}
            {files.length > 0 && (
              <div className="bg-white rounded-xl shadow-sm border border-gray-200">
                <div className="p-4 border-b border-gray-200">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-gray-900">
                      {files.length} file{files.length !== 1 ? "s" : ""}{" "}
                      selected
                    </h3>
                    <div className="text-sm text-gray-500 space-x-4">
                      <span>Total: {formatFileSize(totalSize)}</span>
                      {completedFiles > 0 && (
                        <span className="text-green-600">
                          ✓ {completedFiles} completed
                        </span>
                      )}
                      {errorFiles > 0 && (
                        <span className="text-red-600">
                          ✗ {errorFiles} failed
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                <div className="divide-y divide-gray-200 max-h-96 overflow-y-auto">
                  {files.map((file) => (
                    <div key={file.id} className="p-4 hover:bg-gray-50">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center flex-1 min-w-0">
                          <div
                            className={`flex-shrink-0 h-10 w-10 rounded-lg flex items-center justify-center mr-3 ${
                              file.type.startsWith("image/")
                                ? "bg-pink-100 text-pink-600"
                                : file.type.startsWith("video/")
                                  ? "bg-purple-100 text-purple-600"
                                  : file.type.startsWith("audio/")
                                    ? "bg-blue-100 text-blue-600"
                                    : "bg-gray-100 text-gray-600"
                            }`}
                          >
                            {getFileIcon(file.type)}
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-gray-900 truncate">
                              {file.name}
                            </p>
                            <p className="text-xs text-gray-500">
                              {formatFileSize(file.size)}
                            </p>
                            {file.error && (
                              <p className="text-xs text-red-500 mt-1">
                                Error: {file.error}
                              </p>
                            )}
                          </div>
                        </div>

                        <div className="flex items-center space-x-3">
                          {file.status === "pending" && (
                            <button
                              onClick={() => removeFile(file.id)}
                              disabled={isUploading}
                              className="text-gray-400 hover:text-red-600 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                              <XMarkIcon className="h-5 w-5" />
                            </button>
                          )}

                          {file.status === "uploading" && (
                            <div className="flex items-center">
                              <div className="w-24 bg-gray-200 rounded-full h-2 mr-3">
                                <div className="bg-red-600 h-2 rounded-full transition-all duration-300 animate-pulse" />
                              </div>
                              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-red-600"></div>
                            </div>
                          )}

                          {file.status === "complete" && (
                            <div className="flex items-center text-green-600">
                              <CheckIcon className="h-5 w-5" />
                              <span className="text-sm ml-1">Uploaded</span>
                            </div>
                          )}

                          {file.status === "error" && (
                            <div className="flex items-center text-red-600">
                              <ExclamationTriangleIcon className="h-5 w-5" />
                              <span className="text-sm ml-1">Failed</span>
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
            {/* Collection Selection */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900 flex items-center">
                  <FolderIcon className="h-5 w-5 mr-2 text-gray-500" />
                  Upload to Collection
                </h3>
                <button
                  onClick={loadCollections}
                  disabled={isLoadingCollections}
                  className="text-sm text-blue-600 hover:text-blue-700 disabled:opacity-50"
                >
                  <ArrowPathIcon className="h-4 w-4" />
                </button>
              </div>

              <select
                value={selectedCollection}
                onChange={(e) => setSelectedCollection(e.target.value)}
                disabled={availableCollections.length === 0 || isUploading}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed"
              >
                <option value="">
                  {isLoadingCollections
                    ? "Loading collections..."
                    : availableCollections.length === 0
                      ? "No collections available - click refresh"
                      : "Select a collection..."}
                </option>
                {availableCollections.map((collection) => (
                  <option key={collection.id} value={collection.id}>
                    {collection.name || "[Encrypted]"} (
                    {collection.collection_type})
                  </option>
                ))}
              </select>

              <p className="text-xs text-gray-500 mt-2">
                Files will inherit the collection's encryption settings
              </p>
            </div>

            {/* Password Section */}
            <div className="bg-blue-50 border border-blue-100 rounded-xl p-4">
              <h4 className="text-sm font-semibold text-blue-900 flex items-center mb-3">
                <KeyIcon className="h-4 w-4 mr-2" />
                Encryption Password
              </h4>
              <p className="text-xs text-blue-800 mb-3">
                Used to generate encryption keys locally on your device
              </p>

              <div className="flex space-x-2">
                <div className="relative flex-1">
                  <input
                    type={showPassword ? "text" : "password"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter password or use stored password"
                    className="w-full px-3 py-2 border border-blue-200 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  >
                    {showPassword ? (
                      <EyeSlashIcon className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                    ) : (
                      <EyeIcon className="h-4 w-4 text-gray-400 hover:text-gray-600" />
                    )}
                  </button>
                </div>
                <button
                  type="button"
                  onClick={handleGetStoredPassword}
                  disabled={isUploading}
                  className="px-3 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors duration-200 text-sm"
                >
                  Use Stored
                </button>
              </div>
              <p className="text-xs text-blue-700 mt-1">
                Leave empty to use password from secure storage
              </p>
            </div>

            {/* Upload Options */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
                <SparklesIcon className="h-5 w-5 mr-2 text-gray-500" />
                Upload Options
              </h3>

              <div className="space-y-3">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={encryptFiles}
                    onChange={(e) => setEncryptFiles(e.target.checked)}
                    disabled={isUploading}
                    className="h-4 w-4 text-red-600 rounded border-gray-300 focus:ring-red-500 disabled:opacity-50"
                  />
                  <span className="ml-2 text-sm text-gray-700">
                    Encrypt files before upload
                  </span>
                </label>

                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={generateThumbnails}
                    onChange={(e) => setGenerateThumbnails(e.target.checked)}
                    disabled={isUploading}
                    className="h-4 w-4 text-red-600 rounded border-gray-300 focus:ring-red-500 disabled:opacity-50"
                  />
                  <span className="ml-2 text-sm text-gray-700">
                    Generate thumbnails for images
                  </span>
                </label>

                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={skipDuplicates}
                    onChange={(e) => setSkipDuplicates(e.target.checked)}
                    disabled={isUploading}
                    className="h-4 w-4 text-red-600 rounded border-gray-300 focus:ring-red-500 disabled:opacity-50"
                  />
                  <span className="ml-2 text-sm text-gray-700">
                    Skip duplicate files
                  </span>
                </label>
              </div>
            </div>

            {/* Security Info */}
            <div className="bg-gradient-to-r from-green-50 to-blue-50 rounded-lg border border-green-100 p-4">
              <div className="flex items-start">
                <InformationCircleIcon className="h-5 w-5 text-blue-600 mr-3 flex-shrink-0 mt-0.5" />
                <div className="text-sm text-blue-800">
                  <h4 className="font-semibold mb-1">Secure Upload</h4>
                  <p className="text-xs">
                    Files are encrypted locally with ChaCha20-Poly1305 before
                    upload. Your files remain encrypted at rest and only you can
                    decrypt them.
                  </p>
                </div>
              </div>
            </div>

            {/* Upload Summary */}
            {files.length > 0 && (
              <div className="bg-gray-50 rounded-xl border border-gray-200 p-6">
                <h4 className="font-semibold text-gray-900 mb-3">
                  Upload Summary
                </h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Total files:</span>
                    <span className="font-medium text-gray-900">
                      {files.length}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Total size:</span>
                    <span className="font-medium text-gray-900">
                      {formatFileSize(totalSize)}
                    </span>
                  </div>
                  {(completedFiles > 0 || errorFiles > 0) && (
                    <>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Completed:</span>
                        <span className="font-medium text-green-600">
                          {completedFiles}
                        </span>
                      </div>
                      {errorFiles > 0 && (
                        <div className="flex justify-between">
                          <span className="text-gray-600">Failed:</span>
                          <span className="font-medium text-red-600">
                            {errorFiles}
                          </span>
                        </div>
                      )}
                    </>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Action Buttons */}
        {files.length > 0 && (
          <div className="mt-8 flex items-center justify-between">
            <button
              onClick={() => setFiles([])}
              disabled={isUploading}
              className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Clear All
            </button>

            <div className="flex items-center space-x-3">
              {isUploading ? (
                <button
                  disabled
                  className="inline-flex items-center px-6 py-2 border border-transparent rounded-lg shadow-sm text-white bg-gray-400 cursor-not-allowed"
                >
                  <ArrowPathIcon className="h-4 w-4 mr-2 animate-spin" />
                  Uploading files...
                </button>
              ) : (
                <button
                  onClick={startUpload}
                  disabled={
                    !selectedCollection ||
                    !fileManager ||
                    availableCollections.length === 0
                  }
                  className="inline-flex items-center px-6 py-2 border border-transparent rounded-lg shadow-sm text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                >
                  <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                  Start Secure Upload
                </button>
              )}
            </div>
          </div>
        )}
      </div>

      <style jsx>{`
        @keyframes fade-in {
          from {
            opacity: 0;
            transform: translateY(10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        .animate-fade-in {
          animation: fade-in 0.4s ease-out;
        }
      `}</style>
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(FileUpload);
