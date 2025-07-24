// File: src/pages/User/FileManager/Collections/CollectionDetails.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useParams } from "react-router";
import { useFiles, useCrypto, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  FolderIcon,
  PhotoIcon,
  DocumentIcon,
  ArrowLeftIcon,
  CloudArrowUpIcon,
  ShareIcon,
  ArrowDownTrayIcon,
  MagnifyingGlassIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";

const CollectionDetails = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();

  const { getCollectionManager, listFileManager, downloadFileManager } =
    useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [collection, setCollection] = useState(null);
  const [files, setFiles] = useState([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());

  const loadCollection = useCallback(
    async (forceRefresh = false) => {
      if (!getCollectionManager || !collectionId) return;

      setIsLoading(true);
      setError("");

      try {
        const result = await getCollectionManager.getCollection(
          collectionId,
          forceRefresh,
        );

        if (result.collection) {
          const processedCollection = await processCollection(
            result.collection,
          );
          setCollection(processedCollection);
          await loadCollectionFiles(collectionId, forceRefresh);
        }
      } catch (err) {
        setError("Could not load folder. Please try again.");
      } finally {
        setIsLoading(false);
      }
    },
    [getCollectionManager, collectionId],
  );

  const processCollection = async (rawCollection) => {
    let processedCollection = { ...rawCollection };

    if (rawCollection.encrypted_name || rawCollection.encrypted_description) {
      try {
        let collectionKey =
          rawCollection.collection_key ||
          CollectionCryptoService.getCachedCollectionKey(rawCollection.id);

        if (collectionKey) {
          const { default: CollectionCryptoServiceClass } = await import(
            "../../../../services/Crypto/CollectionCryptoService.js"
          );

          const decryptedCollection =
            await CollectionCryptoServiceClass.decryptCollectionFromAPI(
              rawCollection,
              collectionKey,
            );
          processedCollection = decryptedCollection;
        } else {
          processedCollection.name = "Locked Folder";
          processedCollection._isDecrypted = false;
        }
      } catch (decryptError) {
        processedCollection.name = "Locked Folder";
        processedCollection._isDecrypted = false;
      }
    } else {
      processedCollection._isDecrypted = true;
    }

    processedCollection.type = processedCollection.collection_type || "folder";
    return processedCollection;
  };

  const loadCollectionFiles = useCallback(
    async (collectionId, forceRefresh = false) => {
      if (!listFileManager) return;

      try {
        const result = await listFileManager.listFiles({
          collection_id: collectionId,
          force_refresh: forceRefresh,
        });

        if (result.files) {
          const processedFiles = await processFiles(result.files, collectionId);
          setFiles(processedFiles);
        }
      } catch (err) {
        console.error("Failed to load files:", err);
        setFiles([]);
      }
    },
    [listFileManager],
  );

  const processFiles = async (rawFiles, collectionId) => {
    if (!Array.isArray(rawFiles) || rawFiles.length === 0) return [];

    const collectionKey =
      CollectionCryptoService.getCachedCollectionKey(collectionId);
    if (!collectionKey) {
      return rawFiles.map((file) => ({
        ...file,
        name: "Locked File",
        _isDecrypted: false,
      }));
    }

    const { default: FileCryptoService } = await import(
      "../../../../services/Crypto/FileCryptoService.js"
    );

    const processedFiles = [];
    for (const file of rawFiles) {
      try {
        const decryptedFile = await FileCryptoService.decryptFileFromAPI(
          file,
          collectionKey,
        );
        processedFiles.push({
          ...decryptedFile,
          sizeFormatted: formatFileSize(decryptedFile.size),
        });
      } catch (error) {
        processedFiles.push({
          ...file,
          name: "Locked File",
          _isDecrypted: false,
          sizeFormatted: formatFileSize(file.size),
        });
      }
    }

    return processedFiles;
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  useEffect(() => {
    if (
      getCollectionManager &&
      collectionId &&
      authManager?.isAuthenticated()
    ) {
      loadCollection();
    }
  }, [getCollectionManager, collectionId, authManager, loadCollection]);

  const handleDownloadFile = async (fileId, fileName) => {
    if (!downloadFileManager) return;

    try {
      setDownloadingFiles((prev) => new Set(prev).add(fileId));
      await downloadFileManager.downloadFile(fileId, { saveToDisk: true });
    } catch (err) {
      setError("Could not download file. Please try again.");
    } finally {
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  const handleUploadToCollection = () => {
    navigate(`/file-manager/upload?collection=${collectionId}`, {
      state: {
        preSelectedCollection: {
          id: collectionId,
          name: collection?.name || "Folder",
          type: collection?.type || "folder",
        },
      },
    });
  };

  const filteredFiles = files.filter((file) =>
    (file.name || "").toLowerCase().includes(searchQuery.toLowerCase()),
  );

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Navigation />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading folder...</p>
            </div>
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
            onClick={() => navigate("/file-manager")}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to My Files
          </button>

          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div
                className={`flex items-center justify-center h-10 w-10 rounded-lg ${
                  collection?.type === "album"
                    ? "bg-pink-100 text-pink-600"
                    : "bg-blue-100 text-blue-600"
                }`}
              >
                {collection?.type === "album" ? (
                  <PhotoIcon className="h-6 w-6" />
                ) : (
                  <FolderIcon className="h-6 w-6" />
                )}
              </div>
              <h1 className="text-2xl font-semibold text-gray-900">
                {collection?.name || "Locked Folder"}
              </h1>
            </div>

            <div className="flex items-center space-x-3">
              <button
                onClick={() => loadCollection(true)}
                disabled={isLoading}
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ArrowPathIcon
                  className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
                />
                Refresh
              </button>
              <button
                onClick={() =>
                  navigate(`/file-manager/collections/${collection?.id}/share`)
                }
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <ShareIcon className="h-4 w-4 mr-2" />
                Share
              </button>
              <button
                onClick={handleUploadToCollection}
                className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900"
              >
                <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                Upload
              </button>
            </div>
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="mb-6 p-3 rounded-lg bg-red-50 border border-red-200">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        {/* Search */}
        <div className="mb-6">
          <div className="relative max-w-md">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search files..."
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
            />
          </div>
        </div>

        {/* Files */}
        <div className="bg-white rounded-lg border border-gray-200">
          {filteredFiles.length === 0 ? (
            <div className="p-8 text-center">
              <DocumentIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <h3 className="text-sm font-medium text-gray-900 mb-2">
                {searchQuery ? "No files found" : "No files yet"}
              </h3>
              <p className="text-sm text-gray-500 mb-4">
                {searchQuery
                  ? `No files match "${searchQuery}"`
                  : "Upload files to this folder"}
              </p>
              {!searchQuery && (
                <button
                  onClick={handleUploadToCollection}
                  className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900"
                >
                  <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                  Upload Files
                </button>
              )}
            </div>
          ) : (
            <div className="divide-y divide-gray-200">
              {filteredFiles.map((file) => (
                <div
                  key={file.id}
                  className="p-4 hover:bg-gray-50 flex items-center justify-between"
                >
                  <div className="flex items-center space-x-3">
                    <DocumentIcon className="h-5 w-5 text-gray-400" />
                    <div>
                      <div className="text-sm font-medium text-gray-900">
                        {file.name || "Locked File"}
                      </div>
                      <div className="text-xs text-gray-500">
                        {file.sizeFormatted}
                      </div>
                    </div>
                  </div>
                  <button
                    onClick={() => handleDownloadFile(file.id, file.name)}
                    disabled={
                      !file._isDecrypted || downloadingFiles.has(file.id)
                    }
                    className="p-2 text-gray-400 hover:text-gray-600 disabled:opacity-50"
                  >
                    {downloadingFiles.has(file.id) ? (
                      <div className="animate-spin rounded-full h-4 w-4 border-b border-gray-400"></div>
                    ) : (
                      <ArrowDownTrayIcon className="h-4 w-4" />
                    )}
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(CollectionDetails);
