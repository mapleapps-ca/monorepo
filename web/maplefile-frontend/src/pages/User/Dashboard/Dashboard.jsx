// File: src/pages/User/Dashboard/Dashboard.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import {
  useDashboard,
  useFiles,
  useCrypto,
  useAuth,
} from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";
import Navigation from "../../../components/Navigation";
import {
  CloudArrowUpIcon,
  FolderIcon,
  DocumentIcon,
  ArrowDownTrayIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";

const Dashboard = () => {
  const navigate = useNavigate();
  const { dashboardManager } = useDashboard();
  const { getCollectionManager, downloadFileManager } = useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [dashboardData, setDashboardData] = useState(null);
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());

  const loadDashboardData = useCallback(
    async (forceRefresh = false) => {
      if (!dashboardManager) return;

      setIsLoading(true);
      setError("");

      try {
        const data = await dashboardManager.getDashboardData(forceRefresh);

        if (data.recent_files && data.recent_files.length > 0) {
          const filesByCollection = {};
          data.recent_files.forEach((file) => {
            const collectionId = file.collection_id;
            if (!filesByCollection[collectionId]) {
              filesByCollection[collectionId] = [];
            }
            filesByCollection[collectionId].push(file);
          });

          const collectionIds = Object.keys(filesByCollection);
          await ensureCollectionKeysLoaded(collectionIds);
          const reDecryptedFiles = await reDecryptRecentFiles(
            data.recent_files,
          );
          data.recent_files = reDecryptedFiles;
        }

        setDashboardData(data);
      } catch (err) {
        setError("Could not load your files. Please try again.");
      } finally {
        setIsLoading(false);
      }
    },
    [dashboardManager, getCollectionManager, CollectionCryptoService],
  );

  const ensureCollectionKeysLoaded = async (collectionIds) => {
    if (!getCollectionManager) return;

    const loadPromises = collectionIds.map(async (collectionId) => {
      try {
        let cachedKey =
          CollectionCryptoService.getCachedCollectionKey(collectionId);
        if (cachedKey) return;

        const collection =
          await getCollectionManager.getCollection(collectionId);
        if (collection.collection_key) {
          CollectionCryptoService.cacheCollectionKey(
            collectionId,
            collection.collection_key,
          );
        }
      } catch (error) {
        console.error(`Failed to load collection ${collectionId}:`, error);
      }
    });

    await Promise.allSettled(loadPromises);
  };

  const reDecryptRecentFiles = async (files) => {
    if (!files || files.length === 0) return [];

    const { default: FileCryptoService } = await import(
      "../../../services/Crypto/FileCryptoService.js"
    );
    const decryptedFiles = [];

    for (const file of files) {
      try {
        const collectionKey = CollectionCryptoService.getCachedCollectionKey(
          file.collection_id,
        );
        if (!collectionKey) {
          decryptedFiles.push({
            ...file,
            name: "Locked File",
            _isDecrypted: false,
          });
          continue;
        }

        const decryptedFile = await FileCryptoService.decryptFileFromAPI(
          file,
          collectionKey,
        );
        decryptedFiles.push(decryptedFile);
      } catch (error) {
        decryptedFiles.push({
          ...file,
          name: "Locked File",
          _isDecrypted: false,
        });
      }
    }

    return decryptedFiles;
  };

  useEffect(() => {
    if (dashboardManager && authManager?.isAuthenticated()) {
      loadDashboardData();
    }
  }, [dashboardManager, authManager, loadDashboardData]);

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

  const getTimeAgo = (dateString) => {
    if (!dateString) return "Recently";
    const now = new Date();
    const date = new Date(dateString);
    const diffInMinutes = Math.floor((now - date) / (1000 * 60));

    if (diffInMinutes < 60) return "Just now";
    if (diffInMinutes < 1440) return "Today";
    if (diffInMinutes < 2880) return "Yesterday";
    return "This week";
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return "Small";
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(0)} ${sizes[i]}`;
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-semibold text-gray-900">Welcome Back</h1>
          <div className="flex space-x-3">
            <button
              onClick={() => loadDashboardData(true)}
              disabled={isLoading}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
            >
              <ArrowPathIcon
                className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
              />
              Refresh
            </button>
            <button
              onClick={() => navigate("/file-manager/upload")}
              className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900"
            >
              <CloudArrowUpIcon className="h-4 w-4 mr-2" />
              Upload
            </button>
          </div>
        </div>

        {/* Quick Stats */}
        {dashboardData && (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <div className="bg-white rounded-lg border border-gray-200 p-4">
              <div className="text-2xl font-semibold text-gray-900">
                {dashboardData.summary?.total_files || 0}
              </div>
              <div className="text-sm text-gray-600">Files</div>
            </div>
            <div className="bg-white rounded-lg border border-gray-200 p-4">
              <div className="text-2xl font-semibold text-gray-900">
                {dashboardData.summary?.total_folders || 0}
              </div>
              <div className="text-sm text-gray-600">Folders</div>
            </div>
            <div className="bg-white rounded-lg border border-gray-200 p-4">
              <div className="text-2xl font-semibold text-gray-900">
                {dashboardManager?.formatStorageValue(
                  dashboardData.summary?.storage_used,
                ) || "0 GB"}
              </div>
              <div className="text-sm text-gray-600">Used</div>
            </div>
            <div className="bg-white rounded-lg border border-gray-200 p-4">
              <div className="text-2xl font-semibold text-gray-900">
                {dashboardManager?.formatStorageValue(
                  dashboardData.summary?.storage_limit,
                ) || "0 GB"}
              </div>
              <div className="text-sm text-gray-600">Total</div>
            </div>
          </div>
        )}

        {/* Recent Files */}
        <div className="bg-white rounded-lg border border-gray-200">
          <div className="p-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Recent Files</h2>
          </div>

          {isLoading ? (
            <div className="p-8 text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading your files...</p>
            </div>
          ) : error ? (
            <div className="p-8 text-center">
              <p className="text-red-600">{error}</p>
            </div>
          ) : dashboardData?.recent_files?.length > 0 ? (
            <div className="divide-y divide-gray-200">
              {dashboardData.recent_files.slice(0, 10).map((file) => (
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
                        {formatFileSize(file.size)} •{" "}
                        {getTimeAgo(file.created_at)}
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
          ) : (
            <div className="p-8 text-center">
              <FolderIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <h3 className="text-sm font-medium text-gray-900 mb-2">
                No files yet
              </h3>
              <p className="text-sm text-gray-500 mb-4">
                Upload your first file to get started
              </p>
              <button
                onClick={() => navigate("/file-manager/upload")}
                className="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium text-white bg-red-800 hover:bg-red-900"
              >
                <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                Upload Files
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(Dashboard);
