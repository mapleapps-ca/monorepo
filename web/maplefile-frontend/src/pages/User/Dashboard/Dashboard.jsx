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
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import {
  CloudArrowUpIcon,
  FolderIcon,
  DocumentIcon,
  ArrowDownTrayIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
  ShareIcon,
} from "@heroicons/react/24/outline";

const Dashboard = () => {
  const navigate = useNavigate();

  // Get services from context
  const { dashboardManager } = useDashboard();
  const { getCollectionManager, downloadFileManager } = useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [dashboardData, setDashboardData] = useState(null);
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());

  // Load dashboard data with collection key pre-loading
  const loadDashboardData = useCallback(
    async (forceRefresh = false) => {
      if (!dashboardManager) return;

      setIsLoading(true);
      setError("");

      try {
        console.log("[Dashboard] === Loading Dashboard Data ===");
        console.log("[Dashboard] Force refresh:", forceRefresh);

        // Step 1: Get dashboard data from DashboardManager
        const data = await dashboardManager.getDashboardData(forceRefresh);

        console.log("[Dashboard] Dashboard data loaded:", {
          hasRecentFiles: !!(data.recent_files && data.recent_files.length > 0),
          recentFilesCount: data.recent_files?.length || 0,
        });

        // Step 2: If there are recent files, ensure collection keys are loaded
        if (data.recent_files && data.recent_files.length > 0) {
          console.log("[Dashboard] Processing recent files for decryption...");

          // Group files by collection ID
          const filesByCollection = {};
          data.recent_files.forEach((file) => {
            const collectionId = file.collection_id;
            if (!filesByCollection[collectionId]) {
              filesByCollection[collectionId] = [];
            }
            filesByCollection[collectionId].push(file);
          });

          const collectionIds = Object.keys(filesByCollection);
          console.log("[Dashboard] Collections needed:", collectionIds.length);

          // Step 3: Load collection keys for all collections
          await ensureCollectionKeysLoaded(collectionIds);

          // Step 4: Re-decrypt files with available collection keys
          const reDecryptedFiles = await reDecryptRecentFiles(
            data.recent_files,
          );

          // Update dashboard data with re-decrypted files
          data.recent_files = reDecryptedFiles;

          const decryptedCount = reDecryptedFiles.filter(
            (f) => f._isDecrypted,
          ).length;
          const errorCount = reDecryptedFiles.filter(
            (f) => f._decryptionError,
          ).length;

          console.log("[Dashboard] Re-decryption results:", {
            total: reDecryptedFiles.length,
            decrypted: decryptedCount,
            errors: errorCount,
          });
        }

        setDashboardData(data);
        console.log("[Dashboard] Dashboard loaded successfully");
      } catch (err) {
        console.error("[Dashboard] Failed to load dashboard:", err);
        setError(err.message);
      } finally {
        setIsLoading(false);
      }
    },
    [dashboardManager, getCollectionManager, CollectionCryptoService],
  );

  // Ensure collection keys are loaded
  const ensureCollectionKeysLoaded = async (collectionIds) => {
    console.log("[Dashboard] === Loading Collection Keys ===");
    console.log("[Dashboard] Collections needed:", collectionIds.length);

    if (!getCollectionManager) {
      throw new Error("GetCollectionManager not available");
    }

    const loadPromises = collectionIds.map(async (collectionId) => {
      try {
        // Check if we already have the collection key cached
        let cachedKey =
          CollectionCryptoService.getCachedCollectionKey(collectionId);
        if (cachedKey) {
          console.log(
            "[Dashboard] Collection key already cached:",
            collectionId,
          );
          return;
        }

        // Load collection using collection manager
        console.log("[Dashboard] Loading collection to get key:", collectionId);
        const collection =
          await getCollectionManager.getCollection(collectionId);

        console.log("[Dashboard] Collection loaded:", {
          id: collection.id,
          name: collection.name,
          hasCollectionKey: !!collection.collection_key,
        });

        // Verify collection key is available
        if (!collection.collection_key) {
          throw new Error(
            "Collection key not available after loading collection",
          );
        }

        // Cache the collection key
        CollectionCryptoService.cacheCollectionKey(
          collectionId,
          collection.collection_key,
        );

        console.log(
          "[Dashboard] Collection key cached successfully:",
          collectionId,
        );
      } catch (error) {
        console.error(
          `[Dashboard] Failed to load collection ${collectionId}:`,
          error,
        );
        // Continue with other collections even if one fails
      }
    });

    // Wait for all collection keys to be loaded
    await Promise.allSettled(loadPromises);
    console.log("[Dashboard] Collection key loading completed");
  };

  // Re-decrypt recent files with loaded collection keys
  const reDecryptRecentFiles = async (files) => {
    if (!files || files.length === 0) return [];

    console.log("[Dashboard] === Re-decrypting Recent Files ===");

    // Import FileCryptoService for decryption
    const { default: FileCryptoService } = await import(
      "../../../services/Crypto/FileCryptoService.js"
    );

    const decryptedFiles = [];

    for (const file of files) {
      try {
        // Get collection key
        const collectionKey = CollectionCryptoService.getCachedCollectionKey(
          file.collection_id,
        );

        if (!collectionKey) {
          console.warn(
            `[Dashboard] No collection key available for: ${file.collection_id}`,
          );
          decryptedFiles.push({
            ...file,
            name: "[Collection key unavailable]",
            _isDecrypted: false,
            _decryptionError: "Collection key not available",
          });
          continue;
        }

        // Decrypt file with collection key
        const decryptedFile = await FileCryptoService.decryptFileFromAPI(
          file,
          collectionKey,
        );

        decryptedFiles.push(decryptedFile);

        if (decryptedFile._isDecrypted) {
          console.log(`[Dashboard] ✅ File decrypted: ${decryptedFile.name}`);
        } else {
          console.log(
            `[Dashboard] ❌ File decryption failed: ${decryptedFile._decryptionError}`,
          );
        }
      } catch (error) {
        console.error(`[Dashboard] Failed to decrypt file ${file.id}:`, error);
        decryptedFiles.push({
          ...file,
          name: `[Decrypt failed: ${error.message.substring(0, 50)}...]`,
          _isDecrypted: false,
          _decryptionError: error.message,
        });
      }
    }

    return decryptedFiles;
  };

  // Load dashboard data when component mounts
  useEffect(() => {
    if (dashboardManager && authManager?.isAuthenticated()) {
      loadDashboardData();
    }
  }, [dashboardManager, authManager, loadDashboardData]);

  // Handle file download
  const handleDownloadFile = async (fileId, fileName) => {
    if (!downloadFileManager) return;

    try {
      setDownloadingFiles((prev) => new Set(prev).add(fileId));

      await downloadFileManager.downloadFile(fileId, {
        saveToDisk: true,
      });

      console.log("[Dashboard] File downloaded successfully:", fileName);
    } catch (err) {
      console.error("[Dashboard] Failed to download file:", err);
      setError(`Failed to download file: ${err.message}`);
    } finally {
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  // Refresh dashboard with forced API call
  const handleRefresh = async () => {
    await loadDashboardData(true);
  };

  // Get file type from mime type
  const getFileType = (mimeType) => {
    if (!mimeType) return "Document";
    if (mimeType.startsWith("image/")) return "Image";
    if (mimeType.startsWith("video/")) return "Video";
    if (mimeType.startsWith("audio/")) return "Audio";
    if (mimeType.includes("pdf")) return "PDF";
    if (mimeType.includes("sheet") || mimeType.includes("excel"))
      return "Spreadsheet";
    if (mimeType.includes("document") || mimeType.includes("word"))
      return "Document";
    if (mimeType.includes("presentation") || mimeType.includes("powerpoint"))
      return "Presentation";
    if (mimeType.includes("text")) return "Text";
    return "Document";
  };

  // Format time ago
  const getTimeAgo = (dateString) => {
    if (!dateString) return "Unknown";

    const now = new Date();
    const date = new Date(dateString);
    const diffInMinutes = Math.floor((now - date) / (1000 * 60));

    if (diffInMinutes < 60) return `${diffInMinutes} minutes ago`;
    if (diffInMinutes < 1440)
      return `${Math.floor(diffInMinutes / 60)} hours ago`;
    if (diffInMinutes < 10080)
      return `${Math.floor(diffInMinutes / 1440)} days ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <Navigation />

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-semibold text-gray-900">Dashboard</h1>
          <div className="flex space-x-3">
            <button
              onClick={handleRefresh}
              disabled={isLoading}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <ArrowPathIcon className="h-4 w-4 mr-2" />
              {isLoading ? "Refreshing..." : "Refresh"}
            </button>
            <button
              onClick={() => navigate("/developer/create-file-manager-example")}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
            >
              <CloudArrowUpIcon className="h-4 w-4 mr-2" />
              Upload Files
            </button>
          </div>
        </div>

        {/* Loading State */}
        {isLoading && !dashboardData && (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-600 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading dashboard...</p>
            </div>
          </div>
        )}

        {/* Error State */}
        {error && !dashboardData && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
              <div>
                <h3 className="text-sm font-medium text-red-800">
                  Error loading dashboard
                </h3>
                <p className="text-sm text-red-700 mt-1">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Dashboard Content */}
        {dashboardData && (
          <>
            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              {/* Total Files */}
              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="text-sm font-medium text-gray-500 mb-1">
                  Total Files
                </div>
                <div className="text-3xl font-semibold text-gray-900">
                  {dashboardData.summary?.total_files || 0}
                </div>
              </div>

              {/* Total Folders */}
              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="text-sm font-medium text-gray-500 mb-1">
                  Total Folders
                </div>
                <div className="text-3xl font-semibold text-gray-900">
                  {dashboardData.summary?.total_folders || 0}
                </div>
              </div>

              {/* Storage Used */}
              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="text-sm font-medium text-gray-500 mb-1">
                  Storage Used
                </div>
                <div className="text-3xl font-semibold text-gray-900">
                  {dashboardManager?.formatStorageValue(
                    dashboardData.summary?.storage_used,
                  ) || "0 GB"}
                </div>
              </div>

              {/* Storage Limit */}
              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="text-sm font-medium text-gray-500 mb-1">
                  Storage Limit
                </div>
                <div className="text-3xl font-semibold text-gray-900">
                  {dashboardManager?.formatStorageValue(
                    dashboardData.summary?.storage_limit,
                  ) || "0 GB"}
                </div>
              </div>
            </div>

            {/* Storage Usage Trend Chart - FIXED WITH RECHARTS */}
            {dashboardData.storage_usage_trend &&
              dashboardData.storage_usage_trend.data_points &&
              dashboardData.storage_usage_trend.data_points.length > 0 && (
                <div className="bg-white rounded-lg border border-gray-200 p-6 mb-8">
                  <div className="flex items-center justify-between mb-6">
                    <h2 className="text-lg font-medium text-gray-900">
                      Storage Usage Trend
                    </h2>
                    <span className="text-sm text-gray-500">Last 7 days</span>
                  </div>

                  {/* Chart Container with Recharts */}
                  <div className="h-64">
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart
                        data={dashboardData.storage_usage_trend.data_points.map(
                          (point) => ({
                            name: new Date(point.date).toLocaleDateString(
                              "en-US",
                              {
                                month: "short",
                                day: "numeric",
                              },
                            ),
                            usage: point.usage?.value || 0,
                            unit: point.usage?.unit || "GB",
                          }),
                        )}
                        margin={{ top: 5, right: 20, bottom: 5, left: 0 }}
                      >
                        <CartesianGrid strokeDasharray="3 3" vertical={false} />
                        <XAxis dataKey="name" />
                        <YAxis
                          unit={` ${dashboardData.storage_usage_trend.data_points[0]?.usage?.unit || "GB"}`}
                        />
                        <Tooltip
                          formatter={(value, name, props) => [
                            `${value} ${props.payload?.unit || "GB"}`,
                            "Storage Used",
                          ]}
                        />
                        <Line
                          type="monotone"
                          dataKey="usage"
                          stroke="#dc2626"
                          strokeWidth={2}
                          dot={{ r: 4, strokeWidth: 2 }}
                          activeDot={{ r: 6, strokeWidth: 2 }}
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  </div>
                </div>
              )}

            {/* Recent Files */}
            <div className="bg-white rounded-lg border border-gray-200">
              <div className="flex items-center justify-between p-6 border-b border-gray-200">
                <h2 className="text-lg font-medium text-gray-900">
                  Recent Files
                </h2>
                <button
                  onClick={() =>
                    navigate("/developer/recent-file-manager-example")
                  }
                  className="text-sm text-red-600 hover:text-red-700 font-medium"
                >
                  View All
                </button>
              </div>

              {dashboardData.recent_files &&
              dashboardData.recent_files.length > 0 ? (
                <div className="overflow-hidden">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          File Name
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Uploaded
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Type
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Size
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {dashboardData.recent_files.slice(0, 10).map((file) => (
                        <tr key={file.id} className="hover:bg-gray-50">
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="flex items-center">
                              <DocumentIcon className="h-5 w-5 text-gray-400 mr-3" />
                              <div className="text-sm font-medium text-gray-900">
                                {file.name || "[Encrypted]"}
                              </div>
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {getTimeAgo(file.created_at || file.modified_at)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {getFileType(file.mime_type)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {file.size
                              ? dashboardManager?.formatFileSize?.(file.size) ||
                                `${file.size} bytes`
                              : "Unknown"}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                            <div className="flex space-x-2">
                              <button
                                onClick={() =>
                                  handleDownloadFile(file.id, file.name)
                                }
                                disabled={
                                  !file._isDecrypted ||
                                  downloadingFiles.has(file.id)
                                }
                                className="text-gray-400 hover:text-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
                                title="Download"
                              >
                                {downloadingFiles.has(file.id) ? (
                                  <div className="animate-spin rounded-full h-4 w-4 border-b border-gray-400"></div>
                                ) : (
                                  <ArrowDownTrayIcon className="h-4 w-4" />
                                )}
                              </button>
                              <button
                                className="text-gray-400 hover:text-gray-600"
                                title="Share"
                              >
                                <ShareIcon className="h-4 w-4" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-12">
                  <FolderIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
                  <h3 className="text-sm font-medium text-gray-900 mb-2">
                    No recent files
                  </h3>
                  <p className="text-sm text-gray-500 mb-4">
                    Upload your first files to get started
                  </p>
                  <button
                    onClick={() =>
                      navigate("/developer/create-file-manager-example")
                    }
                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
                  >
                    <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                    Upload Files
                  </button>
                </div>
              )}
            </div>

            {/* Storage Usage Bar */}
            <div className="mt-8 bg-white rounded-lg border border-gray-200 p-6">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm font-medium text-gray-900">
                  Storage Usage:{" "}
                  {dashboardData.summary?.storage_usage_percentage || 0}%
                </span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-red-600 h-2 rounded-full transition-all duration-500"
                  style={{
                    width: `${Math.min(dashboardData.summary?.storage_usage_percentage || 0, 100)}%`,
                  }}
                ></div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(Dashboard);
