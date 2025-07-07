// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/File/CreateFileManagerExample.jsx
// Example component demonstrating how to use the CreateFileManager

import React, { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../../hooks/useService.jsx";

const CreateFileManagerExample = () => {
  const navigate = useNavigate();
  const fileInputRef = useRef(null);
  const { authService } = useServices();

  // State management
  const [fileManager, setFileManager] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [selectedFile, setSelectedFile] = useState(null);
  const [selectedCollectionId, setSelectedCollectionId] = useState("");
  const [password, setPassword] = useState("");
  const [pendingFiles, setPendingFiles] = useState([]);
  const [uploadQueue, setUploadQueue] = useState({});
  const [managerStatus, setManagerStatus] = useState({});
  const [eventLog, setEventLog] = useState([]);
  const [isDragging, setIsDragging] = useState(false);

  // Test collections for demo
  const testCollections = [
    { id: "test-collection-1", name: "Test Collection 1" },
    { id: "test-collection-2", name: "Test Collection 2" },
  ];

  // Initialize file manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authService.isAuthenticated()) return;

      try {
        const { default: CreateFileManager } = await import(
          "../../../../services/Manager/File/CreateFileManager.js"
        );

        const manager = new CreateFileManager(authService);
        await manager.initialize();

        setFileManager(manager);

        // Set up event listener
        manager.addFileCreationListener(handleFileEvent);

        console.log("[Example] CreateFileManager initialized");
      } catch (err) {
        console.error("[Example] Failed to initialize CreateFileManager:", err);
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();

    return () => {
      if (fileManager) {
        fileManager.removeFileCreationListener(handleFileEvent);
      }
    };
  }, [authService]);

  // Load data when manager is ready
  useEffect(() => {
    if (fileManager) {
      loadPendingFiles();
      loadManagerStatus();
    }
  }, [fileManager]);

  // Handle file events
  const handleFileEvent = useCallback((eventType, eventData) => {
    console.log("[Example] File event:", eventType, eventData);
    addToEventLog(eventType, eventData);

    // Reload data on certain events
    if (
      [
        "pending_file_created",
        "pending_file_removed",
        "all_pending_files_cleared",
      ].includes(eventType)
    ) {
      loadPendingFiles();
      loadManagerStatus();
    }
  }, []);

  // Load pending files
  const loadPendingFiles = useCallback(() => {
    if (!fileManager) return;

    const files = fileManager.getPendingFiles();
    const queue = fileManager.getUploadQueue();

    setPendingFiles(files);
    setUploadQueue(queue);

    console.log("[Example] Loaded pending files:", files.length);
  }, [fileManager]);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!fileManager) return;

    const status = fileManager.getManagerStatus();
    setManagerStatus(status);
    console.log("[Example] Manager status:", status);
  }, [fileManager]);

  // Handle drag events
  const handleDragEnter = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleDragOver = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDrop = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      handleFileSelection(files[0]);
    }
  }, []);

  // Handle file selection
  const handleFileSelection = useCallback((file) => {
    if (!file) return;

    // Basic validation
    const maxSize = 100 * 1024 * 1024; // 100MB limit
    if (file.size > maxSize) {
      setError("File size exceeds 100MB limit");
      return;
    }

    setSelectedFile(file);
    setError("");
  }, []);

  // Handle file input change
  const handleFileInputChange = useCallback(
    (e) => {
      const files = e.target.files;
      if (files && files.length > 0) {
        handleFileSelection(files[0]);
      }
    },
    [handleFileSelection],
  );

  // Create pending file
  const handleCreatePendingFile = async () => {
    if (!fileManager || !selectedFile || !selectedCollectionId) {
      setError("Please select a file and collection");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Creating pending file...");

      const result = await fileManager.createPendingFileFromFile(
        selectedFile,
        selectedCollectionId,
        password || null,
      );

      setSuccess(
        `Pending file created successfully! File ID: ${result.fileId}`,
      );
      setSelectedFile(null);

      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }

      addToEventLog("manual_file_created", {
        fileId: result.fileId,
        name: selectedFile.name,
        uploadUrl: result.uploadUrl,
      });
    } catch (err) {
      console.error("[Example] Failed to create pending file:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Remove pending file
  const handleRemovePendingFile = async (fileId) => {
    if (!fileManager) return;

    try {
      await fileManager.removePendingFile(fileId);
      setSuccess("Pending file removed successfully!");
    } catch (err) {
      setError(`Failed to remove file: ${err.message}`);
    }
  };

  // Clear all pending files
  const handleClearAllPendingFiles = async () => {
    if (!fileManager) return;

    if (!window.confirm("Are you sure you want to clear all pending files?")) {
      return;
    }

    try {
      await fileManager.clearAllPendingFiles();
      setSuccess("All pending files cleared!");
    } catch (err) {
      setError(`Failed to clear pending files: ${err.message}`);
    }
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

  // Clear event log
  const handleClearLog = () => {
    setEventLog([]);
  };

  // Format file size
  const formatFileSize = (bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  // Clear messages after 5 seconds
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        setSuccess("");
        setError("");
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [success, error]);

  if (!authService.isAuthenticated()) {
    return (
      <div style={{ padding: "20px", textAlign: "center" }}>
        <p>Please log in to access this example.</p>
        <button onClick={() => navigate("/login")}>Go to Login</button>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <div style={{ marginBottom: "20px" }}>
        <button onClick={() => navigate("/dashboard")}>
          ← Back to Dashboard
        </button>
      </div>

      <h2>📄 Create File Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This example demonstrates creating pending files with E2EE encryption.
      </p>

      {/* Manager Status */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#f8f9fa",
          borderRadius: "6px",
          border: "1px solid #dee2e6",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>📊 Manager Status:</h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: "10px",
          }}
        >
          <div>
            <strong>Authenticated:</strong>{" "}
            {managerStatus.isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can Create Files:</strong>{" "}
            {managerStatus.canCreateFiles ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Pending Files:</strong> {pendingFiles.length}
          </div>
          <div>
            <strong>Upload Queue:</strong> {Object.keys(uploadQueue).length}
          </div>
        </div>
      </div>

      {/* File Upload Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>📤 Create Pending File:</h4>

        {/* Collection Selection */}
        <div style={{ marginBottom: "15px" }}>
          <label
            style={{
              display: "block",
              marginBottom: "5px",
              fontWeight: "bold",
            }}
          >
            Collection *
          </label>
          <select
            value={selectedCollectionId}
            onChange={(e) => setSelectedCollectionId(e.target.value)}
            style={{
              width: "100%",
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
            }}
          >
            <option value="">Select a collection...</option>
            {testCollections.map((col) => (
              <option key={col.id} value={col.id}>
                {col.name} ({col.id})
              </option>
            ))}
          </select>
        </div>

        {/* Password Input */}
        <div style={{ marginBottom: "15px" }}>
          <label
            style={{
              display: "block",
              marginBottom: "5px",
              fontWeight: "bold",
            }}
          >
            Password (for encryption)
          </label>
          <div style={{ display: "flex", gap: "10px" }}>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter password or use stored password..."
              style={{
                flex: 1,
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
              }}
            />
            <button
              onClick={handleGetStoredPassword}
              style={{
                padding: "8px 15px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              Use Stored
            </button>
          </div>
          <small style={{ color: "#666" }}>
            Leave empty to use password from PasswordStorageService
          </small>
        </div>

        {/* File Drop Zone */}
        <div
          onDragEnter={handleDragEnter}
          onDragLeave={handleDragLeave}
          onDragOver={handleDragOver}
          onDrop={handleDrop}
          onClick={() => !isLoading && fileInputRef.current?.click()}
          style={{
            border: `2px dashed ${isDragging ? "#007bff" : "#ccc"}`,
            borderRadius: "8px",
            padding: "40px 20px",
            textAlign: "center",
            backgroundColor: isDragging ? "#f0f8ff" : "#fafafa",
            cursor: isLoading ? "not-allowed" : "pointer",
            transition: "all 0.3s ease",
            marginBottom: "20px",
          }}
        >
          <div style={{ fontSize: "36px", marginBottom: "15px" }}>
            {isLoading ? "⏳" : "📁"}
          </div>

          {!selectedFile ? (
            <>
              <h4>Drag and drop a file here</h4>
              <p style={{ color: "#666", marginTop: "10px" }}>
                or click to select a file from your computer
              </p>
              <p style={{ color: "#999", fontSize: "14px", marginTop: "10px" }}>
                Maximum file size: 100MB
              </p>
            </>
          ) : (
            <div>
              <h4>Selected File:</h4>
              <p style={{ fontSize: "16px", marginTop: "10px" }}>
                📄 {selectedFile.name}
              </p>
              <p style={{ color: "#666", marginTop: "5px" }}>
                Size: {formatFileSize(selectedFile.size)}
              </p>
              <p style={{ color: "#666" }}>
                Type: {selectedFile.type || "Unknown"}
              </p>

              {!isLoading && (
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setSelectedFile(null);
                    setError("");
                  }}
                  style={{
                    marginTop: "15px",
                    padding: "5px 15px",
                    backgroundColor: "#dc3545",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor: "pointer",
                  }}
                >
                  Remove
                </button>
              )}
            </div>
          )}

          <input
            ref={fileInputRef}
            type="file"
            onChange={handleFileInputChange}
            style={{ display: "none" }}
            disabled={isLoading}
          />
        </div>

        {/* Create Button */}
        <button
          onClick={handleCreatePendingFile}
          disabled={isLoading || !selectedFile || !selectedCollectionId}
          style={{
            width: "100%",
            padding: "12px 20px",
            backgroundColor:
              isLoading || !selectedFile || !selectedCollectionId
                ? "#6c757d"
                : "#28a745",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor:
              isLoading || !selectedFile || !selectedCollectionId
                ? "not-allowed"
                : "pointer",
            fontSize: "16px",
            fontWeight: "bold",
          }}
        >
          {isLoading ? "🔄 Creating..." : "📤 Create Pending File"}
        </button>
      </div>

      {/* Success/Error Messages */}
      {success && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#d4edda",
            borderRadius: "6px",
            color: "#155724",
            border: "1px solid #c3e6cb",
          }}
        >
          ✅ {success}
        </div>
      )}

      {error && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#f8d7da",
            borderRadius: "6px",
            color: "#721c24",
            border: "1px solid #f5c6cb",
          }}
        >
          ❌ {error}
        </div>
      )}

      {/* Pending Files List */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#fff3cd",
          borderRadius: "6px",
          border: "1px solid #ffeaa7",
        }}
      >
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "15px",
          }}
        >
          <h4 style={{ margin: 0 }}>
            📋 Pending Files ({pendingFiles.length}):
          </h4>
          {pendingFiles.length > 0 && (
            <button
              onClick={handleClearAllPendingFiles}
              style={{
                padding: "5px 15px",
                backgroundColor: "#dc3545",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              🗑️ Clear All
            </button>
          )}
        </div>

        {pendingFiles.length === 0 ? (
          <p style={{ color: "#666" }}>No pending files yet.</p>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {pendingFiles.map((fileInfo) => (
              <div
                key={fileInfo.file.id}
                style={{
                  padding: "15px",
                  border: "1px solid #dee2e6",
                  borderRadius: "6px",
                  backgroundColor: "white",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "start",
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div style={{ fontWeight: "bold", marginBottom: "5px" }}>
                      📄 File ID: {fileInfo.file.id}
                    </div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      <strong>State:</strong> {fileInfo.file.state} |
                      <strong> Version:</strong> {fileInfo.file.version} |
                      <strong> Created:</strong>{" "}
                      {new Date(fileInfo.stored_at).toLocaleString()}
                    </div>
                    <div
                      style={{
                        fontSize: "12px",
                        color: "#666",
                        marginTop: "5px",
                      }}
                    >
                      <strong>Upload URL Expires:</strong>{" "}
                      {new Date(
                        fileInfo.upload_url_expiration_time,
                      ).toLocaleString()}
                    </div>
                    {uploadQueue[fileInfo.file.id] && (
                      <div
                        style={{
                          fontSize: "12px",
                          color: "#17a2b8",
                          marginTop: "5px",
                        }}
                      >
                        📤 In upload queue - Status:{" "}
                        {uploadQueue[fileInfo.file.id].status}
                      </div>
                    )}
                  </div>
                  <button
                    onClick={() => handleRemovePendingFile(fileInfo.file.id)}
                    style={{
                      padding: "5px 10px",
                      backgroundColor: "#dc3545",
                      color: "white",
                      border: "none",
                      borderRadius: "4px",
                      cursor: "pointer",
                      fontSize: "12px",
                    }}
                  >
                    🗑️ Remove
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Event Log */}
      <div>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <h3>📋 File Event Log ({eventLog.length})</h3>
          <button
            onClick={handleClearLog}
            disabled={eventLog.length === 0}
            style={{
              padding: "5px 15px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: eventLog.length === 0 ? "not-allowed" : "pointer",
              fontSize: "14px",
            }}
          >
            Clear Log
          </button>
        </div>

        {eventLog.length === 0 ? (
          <div
            style={{
              padding: "40px",
              textAlign: "center",
              backgroundColor: "#f8f9fa",
              borderRadius: "6px",
              border: "2px dashed #dee2e6",
            }}
          >
            <p style={{ fontSize: "18px", color: "#6c757d" }}>
              No file events logged yet.
            </p>
          </div>
        ) : (
          <div
            style={{
              maxHeight: "300px",
              overflow: "auto",
              border: "1px solid #dee2e6",
              borderRadius: "6px",
              backgroundColor: "#f8f9fa",
            }}
          >
            {eventLog
              .slice()
              .reverse()
              .map((event, index) => (
                <div
                  key={`${event.timestamp}-${index}`}
                  style={{
                    padding: "10px",
                    borderBottom:
                      index < eventLog.length - 1
                        ? "1px solid #dee2e6"
                        : "none",
                    fontFamily: "monospace",
                    fontSize: "12px",
                  }}
                >
                  <div style={{ marginBottom: "5px" }}>
                    <strong style={{ color: "#007bff" }}>
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </strong>
                    {" - "}
                    <strong style={{ color: "#28a745" }}>
                      {event.eventType}
                    </strong>
                  </div>
                  <div style={{ color: "#666", marginLeft: "20px" }}>
                    {JSON.stringify(event.eventData, null, 2)}
                  </div>
                </div>
              ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default CreateFileManagerExample;
