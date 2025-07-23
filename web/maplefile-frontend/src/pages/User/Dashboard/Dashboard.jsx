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
  ChartBarIcon,
  ArrowDownTrayIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
  InformationCircleIcon,
  CheckIcon,
  ClockIcon,
  ServerIcon,
  EyeSlashIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  SparklesIcon,
  HeartIcon,
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

  // Clear cache and refresh dashboard
  const handleClearCache = async () => {
    if (dashboardManager) {
      dashboardManager.clearAllCaches();
      console.log("[Dashboard] Cache cleared. Forcing data refresh.");
      await loadDashboardData(true);
    }
  };

  // Format storage percentage color
  const getStoragePercentageColor = (percentage) => {
    if (percentage > 90) return "bg-red-500";
    if (percentage > 70) return "bg-yellow-500";
    return "bg-green-500";
  };

  // Get file type icon
  const getFileTypeIcon = (mimeType) => {
    if (!mimeType) return "📄";
    if (mimeType.startsWith("image/")) return "🖼️";
    if (mimeType.startsWith("video/")) return "🎥";
    if (mimeType.startsWith("audio/")) return "🎵";
    if (mimeType.includes("pdf")) return "📄";
    if (mimeType.includes("text")) return "📝";
    return "📄";
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      {/* Navigation */}
      <Navigation />

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8 animate-fade-in-up">
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between">
            <div className="mb-4 lg:mb-0">
              <h1 className="text-4xl font-black text-gray-900 mb-2">
                Welcome Back! 🏠
              </h1>
              <p className="text-xl text-gray-600">
                Manage your secure files with end-to-end encryption
              </p>
              <div className="flex items-center space-x-2 mt-2 text-sm text-gray-500">
                <ShieldCheckIcon className="h-4 w-4 text-green-600" />
                <span>All files encrypted with ChaCha20-Poly1305</span>
              </div>
            </div>
            <div className="flex flex-col sm:flex-row gap-3">
              <button
                onClick={handleClearCache}
                disabled={isLoading}
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
                title="Clear local cache and fetch fresh data from the server"
              >
                <ArrowPathIcon className="h-4 w-4 mr-2" />
                {isLoading ? "Clearing..." : "Clear Cache"}
              </button>
              <button
                onClick={handleRefresh}
                disabled={isLoading}
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
              >
                <ArrowPathIcon className="h-4 w-4 mr-2" />
                {isLoading ? "Refreshing..." : "Refresh"}
              </button>
            </div>
          </div>
        </div>

        {/* Loading State */}
        {isLoading && !dashboardData && (
          <div className="flex items-center justify-center py-12 animate-fade-in">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading your dashboard...</p>
            </div>
          </div>
        )}

        {/* Error State */}
        {error && !dashboardData && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-6 animate-fade-in">
            <div className="flex items-start">
              <ExclamationTriangleIcon className="h-6 w-6 text-red-500 mr-3 flex-shrink-0 mt-1" />
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-red-800 mb-2">
                  Failed to Load Dashboard
                </h3>
                <p className="text-red-700">{error}</p>
                <button
                  onClick={handleRefresh}
                  className="mt-4 inline-flex items-center px-4 py-2 border border-red-300 text-sm font-medium rounded-lg text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
                >
                  <ArrowPathIcon className="h-4 w-4 mr-2" />
                  Try Again
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Dashboard Content */}
        {dashboardData && (
          <>
            {/* Error Banner */}
            {error && (
              <div className="bg-amber-50 border border-amber-200 rounded-xl p-4 mb-6 animate-fade-in">
                <div className="flex items-center">
                  <ExclamationTriangleIcon className="h-5 w-5 text-amber-500 mr-3 flex-shrink-0" />
                  <p className="text-amber-700">{error}</p>
                  <button
                    onClick={() => setError("")}
                    className="ml-auto text-amber-500 hover:text-amber-700"
                  >
                    ✕
                  </button>
                </div>
              </div>
            )}

            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8 animate-fade-in-up-delay">
              {/* Total Files Card */}
              <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-6 transform hover:scale-105 transition-all duration-200 cursor-pointer group">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-4">
                      <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl group-hover:scale-110 transition-transform duration-200">
                        <DocumentIcon className="h-6 w-6 text-white" />
                      </div>
                      <h3 className="text-lg font-semibold text-gray-900">
                        Total Files
                      </h3>
                    </div>
                    <p className="text-3xl font-black text-blue-600">
                      {dashboardData.summary?.total_files || 0}
                    </p>
                    <p className="text-sm text-gray-500 mt-1">
                      Securely encrypted files
                    </p>
                  </div>
                  <SparklesIcon className="h-8 w-8 text-blue-300 opacity-50 group-hover:opacity-100 transition-opacity duration-200" />
                </div>
              </div>

              {/* Total Folders Card */}
              <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-6 transform hover:scale-105 transition-all duration-200 cursor-pointer group">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center space-x-3 mb-4">
                      <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-green-500 to-green-600 rounded-xl group-hover:scale-110 transition-transform duration-200">
                        <FolderIcon className="h-6 w-6 text-white" />
                      </div>
                      <h3 className="text-lg font-semibold text-gray-900">
                        Collections
                      </h3>
                    </div>
                    <p className="text-3xl font-black text-green-600">
                      {dashboardData.summary?.total_folders || 0}
                    </p>
                    <p className="text-sm text-gray-500 mt-1">
                      Organized collections
                    </p>
                  </div>
                  <FolderIcon className="h-8 w-8 text-green-300 opacity-50 group-hover:opacity-100 transition-opacity duration-200" />
                </div>
              </div>

              {/* Storage Usage Card */}
              <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-6 transform hover:scale-105 transition-all duration-200 group">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-xl group-hover:scale-110 transition-transform duration-200">
                      <ChartBarIcon className="h-6 w-6 text-white" />
                    </div>
                    <h3 className="text-lg font-semibold text-gray-900">
                      Storage Usage
                    </h3>
                  </div>
                  <CloudArrowUpIcon className="h-8 w-8 text-purple-300 opacity-50 group-hover:opacity-100 transition-opacity duration-200" />
                </div>

                <p className="text-2xl font-black text-purple-600 mb-1">
                  {dashboardManager?.formatStorageValue(
                    dashboardData.summary?.storage_used,
                  ) || "0 Bytes"}
                </p>
                <p className="text-sm text-gray-500 mb-4">
                  of{" "}
                  {dashboardManager?.formatStorageValue(
                    dashboardData.summary?.storage_limit,
                  ) || "0 Bytes"}{" "}
                  used
                </p>

                {/* Storage Progress Bar */}
                <div className="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
                  <div
                    className={`h-full transition-all duration-500 ${getStoragePercentageColor(
                      dashboardData.summary?.storage_usage_percentage || 0,
                    )}`}
                    style={{
                      width: `${Math.min(dashboardData.summary?.storage_usage_percentage || 0, 100)}%`,
                    }}
                  />
                </div>
                <p className="text-right text-xs text-gray-500 mt-2">
                  {dashboardData.summary?.storage_usage_percentage || 0}% used
                </p>
              </div>
            </div>

            {/* Storage Usage Trend Chart */}
            {dashboardData.storage_usage_trend &&
              dashboardData.storage_usage_trend.data_points &&
              dashboardData.storage_usage_trend.data_points.length > 0 && (
                <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8 mb-8 animate-fade-in-up-delay-2">
                  <div className="flex items-center justify-between mb-6">
                    <div>
                      <h2 className="text-2xl font-bold text-gray-900 mb-2">
                        📈 Storage Usage Trend
                      </h2>
                      <p className="text-gray-600">
                        {dashboardData.storage_usage_trend.period}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-sm text-gray-500">
                        {dashboardData.storage_usage_trend.data_points.length}{" "}
                        data point
                        {dashboardData.storage_usage_trend.data_points
                          .length !== 1
                          ? "s"
                          : ""}
                      </p>
                    </div>
                  </div>

                  {/* Chart Container */}
                  <div className="relative bg-gradient-to-br from-gray-50 to-gray-100 rounded-xl p-6 border">
                    <div className="flex items-end justify-center space-x-4 h-48">
                      {dashboardData.storage_usage_trend.data_points.map(
                        (point, index) => {
                          const maxValue = Math.max(
                            ...dashboardData.storage_usage_trend.data_points.map(
                              (p) => p.usage?.value || 0,
                            ),
                          );

                          let percentage =
                            maxValue > 0
                              ? (point.usage?.value || 0) / maxValue
                              : 0;

                          // For single data point, show it as 70% for better visual
                          if (
                            dashboardData.storage_usage_trend.data_points
                              .length === 1
                          ) {
                            percentage = 0.7;
                          }

                          // Ensure minimum 10% height if there's any value
                          if (percentage > 0 && percentage < 0.1) {
                            percentage = 0.1;
                          }

                          const barHeight = percentage * 160; // 160px max height

                          return (
                            <div
                              key={index}
                              className="flex flex-col items-center group"
                            >
                              {/* Value label */}
                              <div className="mb-2 px-2 py-1 bg-white rounded-lg shadow-sm border text-xs font-semibold text-gray-700 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                                {point.usage?.value
                                  ? `${point.usage.value.toFixed(1)} ${point.usage.unit}`
                                  : "0"}
                              </div>

                              {/* Bar */}
                              <div
                                className="w-12 bg-gradient-to-t from-red-800 to-red-600 rounded-t-lg transition-all duration-300 hover:from-red-900 hover:to-red-700 cursor-pointer"
                                style={{ height: `${barHeight}px` }}
                                title={`${dashboardManager?.formatStorageValue(point.usage) || `${point.usage?.value || 0} ${point.usage?.unit || "KB"}`} on ${new Date(point.date).toLocaleDateString()}`}
                              />

                              {/* Date label */}
                              <div className="mt-2 text-xs font-medium text-gray-600">
                                {new Date(point.date).toLocaleDateString(
                                  "en-US",
                                  {
                                    month: "short",
                                    day: "numeric",
                                  },
                                )}
                              </div>
                            </div>
                          );
                        },
                      )}
                    </div>

                    {/* Chart info */}
                    <div className="mt-6 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100">
                      <div className="flex items-center">
                        <InformationCircleIcon className="h-4 w-4 text-blue-600 mr-2" />
                        <span className="text-sm text-blue-800">
                          {dashboardData.storage_usage_trend.data_points
                            .length === 1 ? (
                            <>
                              <strong>Current Usage:</strong>{" "}
                              {dashboardManager?.formatStorageValue(
                                dashboardData.storage_usage_trend.data_points[0]
                                  .usage,
                              ) ||
                                `${dashboardData.storage_usage_trend.data_points[0].usage?.value || 0} ${dashboardData.storage_usage_trend.data_points[0].usage?.unit || "KB"}`}
                              <br />
                              <em className="text-blue-600">
                                More data points will be available as you use
                                the app over time.
                              </em>
                            </>
                          ) : (
                            <>
                              <strong>Trend:</strong>{" "}
                              {
                                dashboardData.storage_usage_trend.data_points
                                  .length
                              }{" "}
                              data points over{" "}
                              {dashboardData.storage_usage_trend.period?.toLowerCase() ||
                                "time period"}
                            </>
                          )}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              )}

            {/* Recent Files Section */}
            <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8 animate-fade-in-up-delay-3">
              <div className="flex items-center justify-between mb-6">
                <div>
                  <h2 className="text-2xl font-bold text-gray-900 mb-2">
                    🕒 Recent Files
                  </h2>
                  <p className="text-gray-600">
                    Your recently accessed encrypted files
                  </p>
                </div>
                {dashboardData.recent_files &&
                  dashboardData.recent_files.length > 0 && (
                    <button
                      onClick={() =>
                        navigate("/developer/recent-file-manager-example")
                      }
                      className="inline-flex items-center px-4 py-2 border border-red-300 text-sm font-medium rounded-lg text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-all duration-200"
                    >
                      View All Files →
                    </button>
                  )}
              </div>

              {/* Recent Files Table */}
              {dashboardData.recent_files &&
              dashboardData.recent_files.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b-2 border-gray-100">
                        <th className="text-left py-4 px-2 text-sm font-semibold text-gray-600">
                          Name
                        </th>
                        <th className="text-left py-4 px-2 text-sm font-semibold text-gray-600">
                          Type
                        </th>
                        <th className="text-left py-4 px-2 text-sm font-semibold text-gray-600">
                          Size
                        </th>
                        <th className="text-left py-4 px-2 text-sm font-semibold text-gray-600">
                          Modified
                        </th>
                        <th className="text-center py-4 px-2 text-sm font-semibold text-gray-600">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50">
                      {dashboardData.recent_files.map((file) => (
                        <tr
                          key={file.id}
                          className="hover:bg-gray-50 transition-colors duration-150"
                        >
                          <td className="py-4 px-2">
                            <div className="flex items-center space-x-3">
                              <span className="text-xl">
                                {getFileTypeIcon(file.mime_type)}
                              </span>
                              <div>
                                <div className="font-medium text-gray-900">
                                  {file.name || "[Encrypted]"}
                                </div>
                                {file._decryptionError && (
                                  <div className="text-xs text-red-500 mt-1 flex items-center">
                                    <ExclamationTriangleIcon className="h-3 w-3 mr-1" />
                                    Unable to decrypt
                                  </div>
                                )}
                              </div>
                            </div>
                          </td>
                          <td className="py-4 px-2">
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                              {file.mime_type || "Document"}
                            </span>
                          </td>
                          <td className="py-4 px-2 text-sm text-gray-600">
                            {file.size
                              ? dashboardManager?.formatFileSize?.(file.size) ||
                                `${file.size} bytes`
                              : "Unknown"}
                          </td>
                          <td className="py-4 px-2 text-sm text-gray-600">
                            {file.modified_at ? (
                              <div className="flex items-center space-x-1">
                                <ClockIcon className="h-4 w-4 text-gray-400" />
                                <span>
                                  {new Date(
                                    file.modified_at,
                                  ).toLocaleDateString()}
                                </span>
                              </div>
                            ) : (
                              "Unknown"
                            )}
                          </td>
                          <td className="py-4 px-2 text-center">
                            <button
                              onClick={() =>
                                handleDownloadFile(file.id, file.name)
                              }
                              disabled={
                                !file._isDecrypted ||
                                downloadingFiles.has(file.id)
                              }
                              className={`inline-flex items-center px-3 py-1 border text-xs font-medium rounded-lg transition-all duration-200 ${
                                !file._isDecrypted ||
                                downloadingFiles.has(file.id)
                                  ? "border-gray-300 text-gray-400 cursor-not-allowed"
                                  : "border-red-300 text-red-700 bg-white hover:bg-red-50 hover:scale-105"
                              }`}
                              title={
                                !file._isDecrypted
                                  ? "File cannot be decrypted"
                                  : "Download file"
                              }
                            >
                              {downloadingFiles.has(file.id) ? (
                                <div className="animate-spin rounded-full h-3 w-3 border-b border-gray-400"></div>
                              ) : (
                                <ArrowDownTrayIcon className="h-3 w-3" />
                              )}
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                /* Empty State for Recent Files */
                <div className="text-center py-12">
                  <div className="max-w-md mx-auto">
                    <div className="flex justify-center mb-6">
                      <div className="flex items-center justify-center h-20 w-20 bg-gradient-to-br from-gray-100 to-gray-200 rounded-full">
                        <FolderIcon className="h-10 w-10 text-gray-400" />
                      </div>
                    </div>
                    <h3 className="text-xl font-semibold text-gray-900 mb-2">
                      No Recent Files
                    </h3>
                    <p className="text-gray-600 mb-6">
                      Upload your first files to see them here with end-to-end
                      encryption
                    </p>
                    <button
                      onClick={() =>
                        navigate("/developer/create-file-manager-example")
                      }
                      className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transform hover:scale-105 transition-all duration-200 shadow-lg hover:shadow-xl"
                    >
                      <CloudArrowUpIcon className="h-5 w-5 mr-2" />
                      Upload Files
                    </button>
                  </div>
                </div>
              )}
            </div>
          </>
        )}
      </div>

      {/* Trust Badges Footer */}
      <div className="border-t border-gray-100 bg-white/50 backdrop-blur-sm py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-center space-x-8 text-sm">
            <div className="flex items-center space-x-2">
              <LockClosedIcon className="h-4 w-4 text-green-600" />
              <span className="text-gray-600 font-medium">
                ChaCha20-Poly1305 Encryption
              </span>
            </div>
            <div className="flex items-center space-x-2">
              <ServerIcon className="h-4 w-4 text-blue-600" />
              <span className="text-gray-600 font-medium">Canadian Hosted</span>
            </div>
            <div className="flex items-center space-x-2">
              <EyeSlashIcon className="h-4 w-4 text-purple-600" />
              <span className="text-gray-600 font-medium">Zero Knowledge</span>
            </div>
            <div className="flex items-center space-x-2">
              <HeartIcon className="h-4 w-4 text-red-600" />
              <span className="text-gray-600 font-medium">Made in Canada</span>
            </div>
          </div>
        </div>
      </div>

      {/* CSS Animations */}
      <style jsx>{`
        @keyframes fade-in-up {
          from {
            opacity: 0;
            transform: translateY(30px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        .animate-fade-in {
          animation: fade-in-up 0.4s ease-out;
        }

        .animate-fade-in-up {
          animation: fade-in-up 0.6s ease-out;
        }

        .animate-fade-in-up-delay {
          animation: fade-in-up 0.6s ease-out 0.2s both;
        }

        .animate-fade-in-up-delay-2 {
          animation: fade-in-up 0.6s ease-out 0.4s both;
        }

        .animate-fade-in-up-delay-3 {
          animation: fade-in-up 0.6s ease-out 0.6s both;
        }
      `}</style>
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(Dashboard);
