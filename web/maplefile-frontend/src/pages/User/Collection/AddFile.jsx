// File: src/pages/User/Collection/AddFile.jsx
import React, { useState, useRef, useCallback } from "react";
import { useParams, useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const AddFile = () => {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const fileInputRef = useRef(null);

  const {
    fileService,
    collectionService,
    collectionCryptoService,
    cryptoService,
    passwordStorageService,
  } = useServices();

  const [isDragging, setIsDragging] = useState(false);
  const [selectedFile, setSelectedFile] = useState(null);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState("");
  const [collection, setCollection] = useState(null);

  const loadCollection = useCallback(async () => {
    try {
      const password = passwordStorageService.getPassword();
      const collectionData = await collectionService.getCollection(
        collectionId,
        password,
      );
      setCollection(collectionData);
    } catch (err) {
      console.error("Failed to load collection:", err);
      setError("Failed to load collection");
    }
  }, [collectionId, collectionService, passwordStorageService]);

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

  // Load collection info on mount
  React.useEffect(() => {
    loadCollection();
  }, [collectionId, loadCollection]); // Added loadCollection to dependency array

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

  const handleDrop = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragging(false);

      const files = e.dataTransfer.files;
      if (files && files.length > 0) {
        handleFileSelection(files[0]);
      }
    },
    [handleFileSelection],
  ); // Added handleFileSelection to dependency array

  const handleFileInputChange = useCallback(
    (e) => {
      const files = e.target.files;
      if (files && files.length > 0) {
        handleFileSelection(files[0]);
      }
    },
    [handleFileSelection],
  );

  // Encrypt and upload file
  const handleUpload = useCallback(async () => {
    if (!selectedFile || !collection) return;

    setIsUploading(true);
    setError("");
    setUploadProgress(10);

    try {
      console.log("[AddFile] Starting file upload process");

      // Step 1: Generate file encryption key
      const fileKey = cryptoService.generateRandomKey();
      console.log("[AddFile] Generated file encryption key");
      setUploadProgress(20);

      // Step 2: Read file content
      const fileContent = await readFileAsArrayBuffer(selectedFile);
      console.log("[AddFile] Read file content, size:", fileContent.byteLength);
      setUploadProgress(30);

      // Step 3: Encrypt file content
      const encryptedContent = await cryptoService.encryptWithKey(
        new Uint8Array(fileContent),
        fileKey,
      );
      console.log("[AddFile] Encrypted file content");
      setUploadProgress(40);

      // Step 4: Generate file hash
      const fileHash = await cryptoService.hashData(
        new Uint8Array(fileContent),
      );
      const encryptedHash = cryptoService.uint8ArrayToBase64(fileHash);
      setUploadProgress(50);

      // Step 5: Get collection key (should be cached from collection load)
      const collectionKey =
        collection.collection_key ||
        collectionCryptoService.getCachedCollectionKey(collectionId);

      if (!collectionKey) {
        throw new Error("Collection key not found. Please reload the page.");
      }

      // Step 6: Encrypt file key with collection key
      const encryptedFileKey = await cryptoService.encryptFileKey(
        fileKey,
        collectionKey,
      );
      console.log("[AddFile] Encrypted file key with collection key");
      setUploadProgress(60);

      // Step 7: Prepare file metadata
      const metadata = {
        name: selectedFile.name,
        mime_type: selectedFile.type || "application/octet-stream",
        size: selectedFile.size,
        created_at: new Date().toISOString(),
      };

      // Step 8: Encrypt metadata
      const encryptedMetadata = await cryptoService.encryptWithKey(
        JSON.stringify(metadata),
        fileKey,
      );
      console.log("[AddFile] Encrypted metadata");
      setUploadProgress(70);

      // Step 9: Convert encrypted content to proper format
      // The encryptedContent is already base64, we need the raw bytes for upload
      const encryptedBytes = cryptoService.tryDecodeBase64(encryptedContent);

      // Step 10: Prepare file data for API with correct size
      const fileData = {
        id: cryptoService.generateUUID(),
        collection_id: collectionId,
        encrypted_metadata: encryptedMetadata,
        encrypted_file_key: {
          ciphertext: cryptoService.uint8ArrayToBase64(
            encryptedFileKey.ciphertext,
          ),
          nonce: cryptoService.uint8ArrayToBase64(encryptedFileKey.nonce),
          key_version: 1,
        },
        encryption_version: "v1.0",
        encrypted_hash: encryptedHash,
        expected_file_size_in_bytes: encryptedBytes.length, // Use actual encrypted size
      };

      console.log("[AddFile] Prepared file data for upload");
      setUploadProgress(80);

      // Step 11: Create Blob from encrypted bytes
      const encryptedBlob = new Blob([encryptedBytes], {
        type: "application/octet-stream",
      });

      // Step 12: Upload file using FileService
      const uploadedFile = await fileService.uploadFile(
        fileData,
        encryptedBlob,
      );

      console.log("[AddFile] File uploaded successfully:", uploadedFile.id);
      setUploadProgress(100);

      // Navigate back to collection after successful upload
      setTimeout(() => {
        navigate(`/collections/${collectionId}/files`);
      }, 1000);
    } catch (err) {
      console.error("[AddFile] Upload failed:", err);

      // Better error messages for common issues
      if (err.message?.includes("CORS")) {
        setError(
          "Upload failed: CORS error. Please contact support to configure S3 bucket CORS settings.",
        );
      } else if (err.message?.includes("Failed to fetch")) {
        setError(
          "Upload failed: Network error. Please check your connection and try again.",
        );
      } else {
        setError(err.message || "Failed to upload file");
      }

      setUploadProgress(0);
    } finally {
      setIsUploading(false);
    }
  }, [
    selectedFile,
    collection,
    collectionId,
    cryptoService,
    fileService,
    navigate,
    collectionCryptoService,
    passwordStorageService,
  ]);

  // Helper function to read file as ArrayBuffer
  const readFileAsArrayBuffer = useCallback((file) => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = (e) => resolve(e.target.result);
      reader.onerror = (e) => reject(new Error("Failed to read file"));
      reader.readAsArrayBuffer(file);
    });
  }, []);

  // Format file size
  const formatFileSize = useCallback((bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  }, []);

  return (
    <div style={{ padding: "20px", maxWidth: "800px", margin: "0 auto" }}>
      {/* Header */}
      <div style={{ marginBottom: "30px" }}>
        <button
          onClick={() => navigate(`/collections/${collectionId}/files`)}
          disabled={isUploading}
          style={{ marginBottom: "20px" }}
        >
          ← Back to Files
        </button>

        <h1>Add File to Collection</h1>
        {collection && (
          <p style={{ color: "#666" }}>
            Adding to: <strong>{collection.name || "[Encrypted]"}</strong>
          </p>
        )}
      </div>

      {/* Error Display */}
      {error && (
        <div
          style={{
            backgroundColor: "#fee",
            color: "#c00",
            padding: "10px",
            marginBottom: "20px",
            borderRadius: "4px",
          }}
        >
          {error}
        </div>
      )}

      {/* Drag and Drop Zone */}
      <div
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        onClick={() => !isUploading && fileInputRef.current?.click()}
        style={{
          border: `2px dashed ${isDragging ? "#007bff" : "#ccc"}`,
          borderRadius: "8px",
          padding: "60px 20px",
          textAlign: "center",
          backgroundColor: isDragging ? "#f0f8ff" : "#fafafa",
          cursor: isUploading ? "not-allowed" : "pointer",
          transition: "all 0.3s ease",
          marginBottom: "30px",
        }}
      >
        <div style={{ fontSize: "48px", marginBottom: "20px" }}>📁</div>

        {!selectedFile ? (
          <>
            <h3>Drag and drop a file here</h3>
            <p style={{ color: "#666", marginTop: "10px" }}>
              or click to select a file from your computer
            </p>
            <p style={{ color: "#999", fontSize: "14px", marginTop: "10px" }}>
              Maximum file size: 100MB
            </p>
          </>
        ) : (
          <div>
            <h3>Selected File:</h3>
            <p style={{ fontSize: "18px", marginTop: "10px" }}>
              📄 {selectedFile.name}
            </p>
            <p style={{ color: "#666", marginTop: "5px" }}>
              Size: {formatFileSize(selectedFile.size)}
            </p>
            <p style={{ color: "#666" }}>
              Type: {selectedFile.type || "Unknown"}
            </p>

            {!isUploading && (
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
          disabled={isUploading}
        />
      </div>

      {/* Upload Progress */}
      {isUploading && (
        <div style={{ marginBottom: "30px" }}>
          <div
            style={{
              backgroundColor: "#e0e0e0",
              borderRadius: "4px",
              overflow: "hidden",
              height: "30px",
              position: "relative",
            }}
          >
            <div
              style={{
                backgroundColor: "#007bff",
                height: "100%",
                width: `${uploadProgress}%`,
                transition: "width 0.3s ease",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                color: "white",
                fontWeight: "bold",
              }}
            >
              {uploadProgress}%
            </div>
          </div>
          <p style={{ textAlign: "center", marginTop: "10px", color: "#666" }}>
            Encrypting and uploading file...
          </p>
        </div>
      )}

      {/* Action Buttons */}
      <div style={{ textAlign: "center" }}>
        <button
          onClick={handleUpload}
          disabled={!selectedFile || isUploading || !!error}
          style={{
            padding: "12px 40px",
            fontSize: "16px",
            backgroundColor: selectedFile && !error ? "#28a745" : "#ccc",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor:
              selectedFile && !error && !isUploading
                ? "pointer"
                : "not-allowed",
            marginRight: "15px",
          }}
        >
          {isUploading ? "Uploading..." : "Upload File"}
        </button>

        <button
          onClick={() => navigate(`/collections/${collectionId}/files`)}
          disabled={isUploading}
          style={{
            padding: "12px 40px",
            fontSize: "16px",
            backgroundColor: "#6c757d",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: isUploading ? "not-allowed" : "pointer",
          }}
        >
          Cancel
        </button>
      </div>

      {/* Info Box */}
      <div
        style={{
          marginTop: "40px",
          padding: "20px",
          backgroundColor: "#f8f9fa",
          borderRadius: "4px",
          borderLeft: "4px solid #17a2b8",
        }}
      >
        <h4 style={{ marginTop: 0 }}>🔒 End-to-End Encryption</h4>
        <p style={{ marginBottom: "10px" }}>
          Your file will be encrypted on your device before upload:
        </p>
        <ul style={{ marginLeft: "20px", color: "#666" }}>
          <li>File content is encrypted with a unique file key</li>
          <li>File key is encrypted with the collection key</li>
          <li>Only encrypted data is sent to our servers</li>
          <li>Your password never leaves your device</li>
        </ul>
      </div>

      {/* CORS Configuration Note */}
      {error && error.includes("CORS") && (
        <div
          style={{
            marginTop: "20px",
            padding: "20px",
            backgroundColor: "#fff3cd",
            borderRadius: "4px",
            borderLeft: "4px solid #ffc107",
          }}
        >
          <h4 style={{ marginTop: 0 }}>⚠️ CORS Configuration Required</h4>
          <p style={{ marginBottom: "10px" }}>
            To enable file uploads from your browser, the S3/Spaces bucket needs
            proper CORS configuration.
          </p>
          <details>
            <summary style={{ cursor: "pointer", color: "#0066cc" }}>
              View CORS Configuration
            </summary>
            <pre
              style={{
                backgroundColor: "#f5f5f5",
                padding: "10px",
                marginTop: "10px",
                overflow: "auto",
                fontSize: "12px",
              }}
            >
              {`<?xml version="1.0" encoding="UTF-8"?>
<CORSConfiguration>
  <CORSRule>
    <AllowedOrigin>http://localhost:5173</AllowedOrigin>
    <AllowedOrigin>http://localhost:3000</AllowedOrigin>
    <AllowedMethod>GET</AllowedMethod>
    <AllowedMethod>PUT</AllowedMethod>
    <AllowedMethod>POST</AllowedMethod>
    <AllowedMethod>DELETE</AllowedMethod>
    <AllowedHeader>*</AllowedHeader>
    <MaxAgeSeconds>3000</MaxAgeSeconds>
  </CORSRule>
</CORSConfiguration>`}
            </pre>
          </details>
        </div>
      )}
    </div>
  );
};

const AddFileWithPasswordProtection = withPasswordProtection(AddFile);
export default AddFileWithPasswordProtection;
