// File: src/pages/User/Dashboard/Dashboard.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useLocation, Link } from "react-router";
import {
  useDashboard,
  useFiles,
  useCrypto,
  useAuth,
} from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";
import Navigation from "../../../components/Navigation";
import {
  AreaChart,
  Area,
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
  ArrowPathIcon,
  ExclamationTriangleIcon,
  ArrowTrendingUpIcon,
  ClockIcon,
  ChartBarIcon,
  SparklesIcon,
} from "@heroicons/react/24/outline";

const Dashboard = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { dashboardManager } = useDashboard();
  const { getCollectionManager, downloadFileManager, createFileManager } =
    useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [dashboardData, setDashboardData] = useState(null);
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());

  const ensureCollectionKeysLoaded = useCallback(
    async (collectionIds) => {
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
    },
    [getCollectionManager, CollectionCryptoService],
  );

  const reDecryptRecentFiles = useCallback(
    async (files) => {
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
    },
    [CollectionCryptoService],
  );

  const loadDashboardData = useCallback(
    async (forceRefresh = false) => {
      if (!dashboardManager) return;

      setIsLoading(true);
      setError("");

      try {
        console.log("[Dashboard] Loading dashboard data...", { forceRefresh });
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
        console.log("[Dashboard] Dashboard loaded successfully");
      } catch (err) {
        console.error("[Dashboard] Failed to load dashboard:", err);
        setError("Could not load your files. Please try again.");
      } finally {
        setIsLoading(false);
      }
    },
    [dashboardManager, ensureCollectionKeysLoaded, reDecryptRecentFiles],
  );

  const handleFileUploadEvents = useCallback(
    (eventType, eventData) => {
      console.log(
        "[Dashboard] File upload event received:",
        eventType,
        eventData,
      );

      if (eventType === "file_upload_completed") {
        console.log(
          "[Dashboard] File upload completed, refreshing dashboard...",
        );
        loadDashboardData(true);
      }
    },
    [loadDashboardData],
  );

  const clearDashboardCache = useCallback(() => {
    if (dashboardManager) {
      console.log("[Dashboard] Clearing dashboard cache");
      dashboardManager.clearAllCaches();
    }
  }, [dashboardManager]);

  useEffect(() => {
    if (dashboardManager && authManager?.isAuthenticated()) {
      loadDashboardData();
    }
  }, [dashboardManager, authManager, loadDashboardData]);

  useEffect(() => {
    if (createFileManager) {
      console.log("[Dashboard] Adding file upload event listener");
      createFileManager.addFileCreationListener(handleFileUploadEvents);

      return () => {
        console.log("[Dashboard] Removing file upload event listener");
        createFileManager.removeFileCreationListener(handleFileUploadEvents);
      };
    }
  }, [createFileManager, handleFileUploadEvents]);

  useEffect(() => {
    if (location.state?.refreshDashboard || location.state?.uploadCompleted) {
      console.log("[Dashboard] Detected upload completion, forcing refresh");
      clearDashboardCache();
      loadDashboardData(true);

      navigate(location.pathname, { replace: true, state: {} });
    }
  }, [
    location.state,
    loadDashboardData,
    clearDashboardCache,
    navigate,
    location.pathname,
  ]);

  useEffect(() => {
    const handleStorageChange = (e) => {
      if (e.key && e.key.includes("file") && e.key.includes("upload")) {
        console.log(
          "[Dashboard] Storage change detected, refreshing dashboard",
        );
        loadDashboardData(true);
      }
    };

    window.addEventListener("storage", handleStorageChange);
    return () => window.removeEventListener("storage", handleStorageChange);
  }, [loadDashboardData]);

  useEffect(() => {
    const handleDashboardRefresh = () => {
      console.log("[Dashboard] Dashboard refresh event received");
      clearDashboardCache();
      loadDashboardData(true);
    };

    window.addEventListener("dashboardRefresh", handleDashboardRefresh);
    return () =>
      window.removeEventListener("dashboardRefresh", handleDashboardRefresh);
  }, [loadDashboardData, clearDashboardCache]);

  const handleDownloadFile = async (fileId) => {
    if (!downloadFileManager) return;

    try {
      setDownloadingFiles((prev) => new Set(prev).add(fileId));
      await downloadFileManager.downloadFile(fileId, { saveToDisk: true });
    } catch (err) {
      console.error("[Dashboard] Failed to download file:", err);
      setError("Could not download file. Please try again.");
    } finally {
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  const handleManualRefresh = async () => {
    console.log("[Dashboard] Manual refresh triggered");
    clearDashboardCache();
    await loadDashboardData(true);
  };

  const formatFileSize = (bytes) => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / 1024 ** i).toFixed(1)} ${sizes[i]}`;
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

  const getFileIcon = (fileName) => {
    if (!fileName || fileName === "Locked File") {
      return <DocumentIcon className="h-5 w-5 text-gray-600" />;
    }
    const extension = fileName.split(".").pop().toLowerCase();
    const iconClass = "h-5 w-5";

    if (["pdf"].includes(extension)) {
      return <DocumentIcon className={`${iconClass} text-red-600`} />;
    }
    if (["doc", "docx"].includes(extension)) {
      return <DocumentIcon className={`${iconClass} text-blue-600`} />;
    }
    if (["fig", "sketch"].includes(extension)) {
      return <DocumentIcon className={`${iconClass} text-purple-600`} />;
    }
    return <DocumentIcon className={`${iconClass} text-gray-600`} />;
  };

  const stats = dashboardData
    ? [
        {
          label: "Files",
          value: dashboardData.summary?.total_files || 0,
          icon: DocumentIcon,
          color: "from-blue-600 to-blue-700",
          bgColor: "bg-blue-50",
          textColor: "text-blue-700",
        },
        {
          label: "Folders",
          value: dashboardData.summary?.total_folders || 0,
          icon: FolderIcon,
          color: "from-purple-600 to-purple-700",
          bgColor: "bg-purple-50",
          textColor: "text-purple-700",
        },
        {
          label: "Storage Used",
          value:
            dashboardManager?.formatStorageValue(
              dashboardData.summary?.storage_used,
            ) || "0 GB",
          icon: ChartBarIcon,
          color: "from-green-600 to-green-700",
          bgColor: "bg-green-50",
          textColor: "text-green-700",
        },
        {
          label: "Total Storage",
          value:
            dashboardManager?.formatStorageValue(
              dashboardData.summary?.storage_limit,
            ) || "0 GB",
          icon: CloudArrowUpIcon,
          color: "from-red-600 to-red-700",
          bgColor: "bg-red-50",
          textColor: "text-red-700",
        },
      ]
    : [];

  return (
    <div className="min-h-screen bg-gradient-subtle">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8 animate-fade-in-down">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 flex items-center">
                Welcome Back
                <SparklesIcon className="h-8 w-8 text-yellow-500 ml-2" />
              </h1>
              <p className="text-gray-600 mt-1">
                Here's what's happening with your files today
              </p>
            </div>
            <div className="flex items-center space-x-3">
              <button
                onClick={handleManualRefresh}
                disabled={isLoading}
                className="btn-secondary"
              >
                <ArrowPathIcon
                  className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
                />
                Refresh
              </button>
              <button
                onClick={() => navigate("/file-manager/upload")}
                className="btn-primary"
              >
                <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                Upload
              </button>
            </div>
          </div>
        </div>

        {/* Loading State */}
        {isLoading && !dashboardData && (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-600 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading your dashboard...</p>
            </div>
          </div>
        )}

        {/* Error State */}
        {error && !dashboardData && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-8">
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
            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              {stats.map((stat, index) => (
                <div
                  key={stat.label}
                  className="card hover:shadow-lg transform hover:-translate-y-1 transition-all duration-300 animate-fade-in-up"
                  style={{ animationDelay: `${index * 100}ms` }}
                >
                  <div className="p-6">
                    <div className="flex items-center justify-between mb-4">
                      <div
                        className={`h-12 w-12 ${stat.bgColor} rounded-xl flex items-center justify-center`}
                      >
                        <stat.icon className={`h-6 w-6 ${stat.textColor}`} />
                      </div>
                      <ArrowTrendingUpIcon className="h-5 w-5 text-green-500" />
                    </div>
                    <h3 className="text-3xl font-bold text-gray-900">
                      {stat.value}
                    </h3>
                    <p className="text-sm text-gray-600 mt-1">{stat.label}</p>
                  </div>
                  <div
                    className={`h-1 bg-gradient-to-r ${stat.color} rounded-b-xl`}
                  ></div>
                </div>
              ))}
            </div>

            {/* Charts Section */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
              {/* Storage Usage Chart */}
              <div className="lg:col-span-2 card animate-fade-in-up">
                <div className="p-6">
                  <div className="flex items-center justify-between mb-6">
                    <div>
                      <h2 className="text-lg font-semibold text-gray-900">
                        Storage Usage Trend
                      </h2>
                      <p className="text-sm text-gray-600">Last 7 days</p>
                    </div>
                    <div className="flex items-center space-x-2">
                      <div className="h-3 w-3 bg-red-600 rounded-full"></div>
                      <span className="text-sm text-gray-600">
                        Storage (GB)
                      </span>
                    </div>
                  </div>

                  <div className="h-64">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart
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
                        <defs>
                          <linearGradient
                            id="colorUsage"
                            x1="0"
                            y1="0"
                            x2="0"
                            y2="1"
                          >
                            <stop
                              offset="5%"
                              stopColor="#dc2626"
                              stopOpacity={0.3}
                            />
                            <stop
                              offset="95%"
                              stopColor="#dc2626"
                              stopOpacity={0}
                            />
                          </linearGradient>
                        </defs>
                        <CartesianGrid
                          strokeDasharray="3 3"
                          vertical={false}
                          stroke="#f3f4f6"
                        />
                        <XAxis dataKey="name" stroke="#9ca3af" fontSize={12} />
                        <YAxis
                          unit={` ${dashboardData.storage_usage_trend.data_points[0]?.usage?.unit || "GB"}`}
                          stroke="#9ca3af"
                          fontSize={12}
                        />
                        <Tooltip
                          contentStyle={{
                            backgroundColor: "#ffffff",
                            border: "1px solid #e5e7eb",
                            borderRadius: "0.5rem",
                            boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1)",
                          }}
                          formatter={(value, name, props) => [
                            `${value} ${props.payload?.unit || "GB"}`,
                            "Storage Used",
                          ]}
                        />
                        <Area
                          type="monotone"
                          dataKey="usage"
                          stroke="#dc2626"
                          strokeWidth={2}
                          fill="url(#colorUsage)"
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </div>
              </div>

              {/* Storage Summary */}
              <div
                className="card animate-fade-in-up"
                style={{ animationDelay: "200ms" }}
              >
                <div className="p-6">
                  <h2 className="text-lg font-semibold text-gray-900 mb-6">
                    Storage Overview
                  </h2>

                  <div className="space-y-6">
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-gray-700">
                          Used Storage
                        </span>
                        <span className="text-sm font-semibold text-gray-900">
                          {dashboardData.summary?.storage_usage_percentage || 0}
                          %
                        </span>
                      </div>
                      <div className="relative">
                        <div className="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
                          <div
                            className="bg-gradient-to-r from-red-600 to-red-700 h-3 rounded-full transition-all duration-1000 ease-out"
                            style={{
                              width: `${Math.min(dashboardData.summary?.storage_usage_percentage || 0, 100)}%`,
                            }}
                          ></div>
                        </div>
                      </div>
                      <div className="flex items-center justify-between mt-2">
                        <span className="text-xs text-gray-500">
                          {dashboardManager?.formatStorageValue(
                            dashboardData.summary?.storage_used,
                          ) || "0 GB"}{" "}
                          used
                        </span>
                        <span className="text-xs text-gray-500">
                          {dashboardManager?.formatStorageValue(
                            dashboardData.summary?.storage_limit,
                          ) || "0 GB"}{" "}
                          total
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Recent Files */}
            <div
              className="card animate-fade-in-up"
              style={{ animationDelay: "300ms" }}
            >
              <div className="p-6 border-b border-gray-100">
                <div className="flex items-center justify-between">
                  <h2 className="text-lg font-semibold text-gray-900">
                    Recent Files
                  </h2>
                  <Link
                    to="/file-manager"
                    className="text-sm font-medium text-red-700 hover:text-red-800 transition-colors duration-200"
                  >
                    View all →
                  </Link>
                </div>
              </div>

              {dashboardData?.recent_files?.length > 0 ? (
                <div className="divide-y divide-gray-100">
                  {dashboardData.recent_files.slice(0, 5).map((file) => (
                    <div
                      key={file.id}
                      className="p-4 hover:bg-gray-50 transition-colors duration-200 group"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-4">
                          <div className="h-10 w-10 bg-gray-50 rounded-lg flex items-center justify-center group-hover:bg-gray-100 transition-colors duration-200">
                            {getFileIcon(file.name)}
                          </div>
                          <div>
                            <h3 className="text-sm font-medium text-gray-900">
                              {file.name || "Locked File"}
                            </h3>
                            <div className="flex items-center space-x-4 mt-1">
                              <span className="text-xs text-gray-500">
                                {formatFileSize(file.size)}
                              </span>
                              <span className="text-xs text-gray-400">•</span>
                              <span className="text-xs text-gray-500 flex items-center">
                                <ClockIcon className="h-3 w-3 mr-1" />
                                {getTimeAgo(file.created_at)}
                              </span>
                            </div>
                          </div>
                        </div>
                        <button
                          onClick={() => handleDownloadFile(file.id)}
                          disabled={
                            !file._isDecrypted || downloadingFiles.has(file.id)
                          }
                          className="opacity-0 group-hover:opacity-100 p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-all duration-200 disabled:opacity-50"
                        >
                          {downloadingFiles.has(file.id) ? (
                            <ArrowPathIcon className="h-4 w-4 animate-spin" />
                          ) : (
                            <ArrowDownTrayIcon className="h-4 w-4" />
                          )}
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="p-12 text-center">
                  <div className="h-16 w-16 bg-gray-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
                    <FolderIcon className="h-8 w-8 text-gray-400" />
                  </div>
                  <h3 className="text-base font-medium text-gray-900 mb-2">
                    No files yet
                  </h3>
                  <p className="text-sm text-gray-500 mb-6">
                    Upload your first file to get started
                  </p>
                  <button
                    onClick={() => navigate("/file-manager/upload")}
                    className="btn-primary"
                  >
                    <CloudArrowUpIcon className="h-4 w-4 mr-2" />
                    Upload Files
                  </button>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default withPasswordProtection(Dashboard);
