// Example usage of FileService in React components

// Example 1: Basic file upload component
import React, { useState } from "react";
import { useServices } from "../hooks/useService.jsx";

const FileUploadExample = ({ collectionId }) => {
  const { fileService, cryptoService } = useServices();
  const [uploading, setUploading] = useState(false);

  const handleFileUpload = async (file) => {
    try {
      setUploading(true);

      // Step 1: Read file content
      const fileBuffer = await file.arrayBuffer();
      const fileData = new Uint8Array(fileBuffer);

      // Step 2: Generate file encryption key
      const fileKey = cryptoService.generateRandomKey(); // You'll need to implement this

      // Step 3: Encrypt file content
      const encryptedContent = await cryptoService.encryptWithKey(
        fileData,
        fileKey,
      );

      // Step 4: Encrypt file metadata
      const metadata = {
        filename: file.name,
        mimeType: file.type,
        size: file.size,
        lastModified: file.lastModified,
      };
      const encryptedMetadata = await cryptoService.encryptWithKey(
        JSON.stringify(metadata),
        fileKey,
      );

      // Step 5: Encrypt the file key with collection key
      const encryptedFileKey = await cryptoService.encryptFileKey(fileKey);

      // Step 6: Create file hash for integrity
      const fileHash = await cryptoService.hashData(encryptedContent);
      const encryptedHash = await cryptoService.encryptWithKey(
        fileHash,
        fileKey,
      );

      // Step 7: Generate a unique file ID
      const fileId = cryptoService.generateUUID(); // You'll need to implement this

      // Step 8: Prepare file data for API
      const fileUploadData = {
        id: fileId,
        collection_id: collectionId,
        encrypted_metadata: encryptedMetadata,
        encrypted_file_key: {
          ciphertext: Array.from(encryptedFileKey.ciphertext),
          nonce: Array.from(encryptedFileKey.nonce),
          key_version: 1,
          rotated_at: new Date().toISOString(),
        },
        encryption_version: "v1.0",
        encrypted_hash: encryptedHash,
        expected_file_size_in_bytes: encryptedContent.length,
      };

      // Step 9: Upload file using FileService
      const uploadedFile = await fileService.uploadFile(
        fileUploadData,
        encryptedContent,
      );

      console.log("File uploaded successfully:", uploadedFile);
      return uploadedFile;
    } catch (error) {
      console.error("File upload failed:", error);
      throw error;
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input
        type="file"
        onChange={(e) =>
          e.target.files[0] && handleFileUpload(e.target.files[0])
        }
        disabled={uploading}
      />
      {uploading && <p>Uploading...</p>}
    </div>
  );
};

// Example 2: File download component
const FileDownloadExample = ({ fileId }) => {
  const { fileService, cryptoService } = useServices();
  const [downloading, setDownloading] = useState(false);

  const handleFileDownload = async () => {
    try {
      setDownloading(true);

      // Step 1: Download encrypted file
      const downloadResult = await fileService.downloadFile(fileId);
      const { metadata, encryptedContent } = downloadResult;

      // Step 2: Get the collection to access its key
      // You would get this from the collection service
      const collectionKey = await getCollectionKey(metadata.collection_id);

      // Step 3: Decrypt the file key
      const fileKey = await cryptoService.decryptFileKey(
        metadata.encrypted_file_key,
        collectionKey,
      );

      // Step 4: Decrypt file content
      const decryptedContent = await cryptoService.decryptWithKey(
        encryptedContent,
        fileKey,
      );

      // Step 5: Decrypt metadata
      const decryptedMetadata = JSON.parse(
        await cryptoService.decryptWithKey(
          metadata.encrypted_metadata,
          fileKey,
        ),
      );

      // Step 6: Create download blob
      const blob = new Blob([decryptedContent], {
        type: decryptedMetadata.mimeType,
      });

      // Step 7: Trigger download
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = decryptedMetadata.filename;
      a.click();
      URL.revokeObjectURL(url);

      console.log("File downloaded successfully");
    } catch (error) {
      console.error("File download failed:", error);
    } finally {
      setDownloading(false);
    }
  };

  return (
    <button onClick={handleFileDownload} disabled={downloading}>
      {downloading ? "Downloading..." : "Download File"}
    </button>
  );
};

// Example 3: File list component
const FileListExample = ({ collectionId }) => {
  const { fileService } = useServices();
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(false);

  React.useEffect(() => {
    loadFiles();
  }, [collectionId]);

  const loadFiles = async () => {
    try {
      setLoading(true);
      const fileList = await fileService.listFilesByCollection(collectionId);
      setFiles(fileList);
    } catch (error) {
      console.error("Failed to load files:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleArchiveFile = async (fileId) => {
    try {
      await fileService.archiveFile(fileId);
      await loadFiles(); // Reload list
    } catch (error) {
      console.error("Failed to archive file:", error);
    }
  };

  const handleDeleteFile = async (fileId) => {
    if (window.confirm("Are you sure you want to delete this file?")) {
      try {
        await fileService.deleteFile(fileId);
        await loadFiles(); // Reload list
      } catch (error) {
        console.error("Failed to delete file:", error);
      }
    }
  };

  if (loading) return <p>Loading files...</p>;

  return (
    <div>
      <h3>Files in Collection</h3>
      <ul>
        {files.map((file) => (
          <li key={file.id}>
            {file.encrypted_metadata} ({file.encrypted_file_size_in_bytes}{" "}
            bytes)
            <button onClick={() => handleArchiveFile(file.id)}>Archive</button>
            <button onClick={() => handleDeleteFile(file.id)}>Delete</button>
          </li>
        ))}
      </ul>
    </div>
  );
};

// Example 4: Batch operations
const BatchFileOperationsExample = () => {
  const { fileService } = useServices();
  const [selectedFiles, setSelectedFiles] = useState([]);

  const handleBatchDelete = async () => {
    if (selectedFiles.length === 0) {
      alert("No files selected");
      return;
    }

    if (window.confirm(`Delete ${selectedFiles.length} files?`)) {
      try {
        const result = await fileService.deleteMultipleFiles(selectedFiles);
        console.log("Batch delete result:", result);
        setSelectedFiles([]);
      } catch (error) {
        console.error("Batch delete failed:", error);
      }
    }
  };

  const handleBatchArchive = async () => {
    if (selectedFiles.length === 0) {
      alert("No files selected");
      return;
    }

    try {
      const result = await fileService.batchArchiveFiles(selectedFiles);
      console.log("Batch archive result:", result);
      setSelectedFiles([]);
    } catch (error) {
      console.error("Batch archive failed:", error);
    }
  };

  return (
    <div>
      <p>Selected files: {selectedFiles.length}</p>
      <button onClick={handleBatchDelete}>Delete Selected</button>
      <button onClick={handleBatchArchive}>Archive Selected</button>
    </div>
  );
};

// Example 5: File sync for offline support
const FileSyncExample = () => {
  const { fileService } = useServices();
  const [syncing, setSyncing] = useState(false);
  const [syncProgress, setSyncProgress] = useState(null);

  const handleSync = async () => {
    try {
      setSyncing(true);
      setSyncProgress({ synced: 0, hasMore: true });

      let cursor = null;
      let totalSynced = 0;

      while (true) {
        const result = await fileService.syncFiles(cursor, 1000);

        totalSynced += result.files?.length || 0;
        setSyncProgress({
          synced: totalSynced,
          hasMore: result.has_more,
        });

        if (!result.has_more) break;
        cursor = result.next_cursor;
      }

      console.log(`Sync complete: ${totalSynced} files synced`);
    } catch (error) {
      console.error("Sync failed:", error);
    } finally {
      setSyncing(false);
    }
  };

  return (
    <div>
      <button onClick={handleSync} disabled={syncing}>
        {syncing ? "Syncing..." : "Sync Files"}
      </button>
      {syncProgress && (
        <p>
          Synced: {syncProgress.synced} files
          {syncProgress.hasMore && " (more available)"}
        </p>
      )}
    </div>
  );
};

// Example 6: Advanced file management with custom hook
const useFiles = (collectionId) => {
  const { fileService } = useServices();
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const loadFiles = React.useCallback(async () => {
    if (!collectionId) return;

    try {
      setLoading(true);
      setError(null);
      const fileList = await fileService.listFilesByCollection(collectionId);
      setFiles(fileList);
    } catch (err) {
      setError(err.message);
      setFiles([]);
    } finally {
      setLoading(false);
    }
  }, [collectionId, fileService]);

  React.useEffect(() => {
    loadFiles();
  }, [loadFiles]);

  const uploadFile = React.useCallback(
    async (fileData, encryptedContent, encryptedThumbnail) => {
      try {
        const uploadedFile = await fileService.uploadFile(
          fileData,
          encryptedContent,
          encryptedThumbnail,
        );
        await loadFiles(); // Reload the list
        return uploadedFile;
      } catch (err) {
        setError(err.message);
        throw err;
      }
    },
    [fileService, loadFiles],
  );

  const downloadFile = React.useCallback(
    async (fileId) => {
      try {
        return await fileService.downloadFile(fileId);
      } catch (err) {
        setError(err.message);
        throw err;
      }
    },
    [fileService],
  );

  const deleteFile = React.useCallback(
    async (fileId) => {
      try {
        await fileService.deleteFile(fileId);
        await loadFiles(); // Reload the list
      } catch (err) {
        setError(err.message);
        throw err;
      }
    },
    [fileService, loadFiles],
  );

  return {
    files,
    loading,
    error,
    loadFiles,
    uploadFile,
    downloadFile,
    deleteFile,
    // Additional methods
    archiveFile: (fileId) => fileService.archiveFile(fileId).then(loadFiles),
    restoreFile: (fileId) => fileService.restoreFile(fileId).then(loadFiles),
    updateFile: (fileId, data) =>
      fileService.updateFile(fileId, data).then(loadFiles),
  };
};

export {
  FileUploadExample,
  FileDownloadExample,
  FileListExample,
  BatchFileOperationsExample,
  FileSyncExample,
  useFiles,
};
