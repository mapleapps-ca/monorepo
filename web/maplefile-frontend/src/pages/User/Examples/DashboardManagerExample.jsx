// File: monorepo/web/maplefile-frontend/src/pages/User/Dashboard/DashboardManagerExample.jsx

import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useDashboard, useFiles, useCrypto } from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";

const DashboardManagerExample = () => {
  const navigate = useNavigate();

  // Get services from context
  const { dashboardManager } = useDashboard();
  const {
    authService,
    getCollectionManager,
    listCollectionManager,
    downloadFileManager,
  } = useFiles();
  const { CollectionCryptoService } = useCrypto();

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

  // Ensure collection keys are loaded (similar to RecentFileManagerExample)
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
    if (dashboardManager && authService.isAuthenticated()) {
      loadDashboardData();
    }
  }, [dashboardManager, authService, loadDashboardData]);

  // Handle logout
  const handleLogout = () => {
    if (authService?.logout) {
      authService.logout();
    }
    navigate("/");
  };

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

  // Navigate to file manager
  const handleViewAllFiles = () => {
    navigate("/recent-file-manager-example");
  };

  // Navigate to collection manager
  const handleViewAllCollections = () => {
    navigate("/list-collection-manager-example");
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
    if (percentage > 90) return "#dc3545"; // Red
    if (percentage > 70) return "#ffc107"; // Yellow
    return "#28a745"; // Green
  };

  // Loading state
  if (isLoading && !dashboardData) {
    return (
      <div style={{ padding: "20px", textAlign: "center" }}>
        <div style={{ fontSize: "24px", marginBottom: "10px" }}>⏳</div>
        <p>Loading dashboard...</p>
      </div>
    );
  }

  // Error state
  if (error && !dashboardData) {
    return (
      <div style={{ padding: "20px", textAlign: "center" }}>
        <div
          style={{ fontSize: "24px", marginBottom: "10px", color: "#dc3545" }}
        >
          ❌
        </div>
        <p style={{ color: "#dc3545" }}>{error}</p>
        <button
          onClick={handleRefresh}
          style={{
            marginTop: "10px",
            padding: "8px 16px",
            backgroundColor: "#007bff",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
          }}
        >
          Try Again
        </button>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "30px",
        }}
      >
        <div>
          <h1 style={{ margin: 0 }}>🏠 Dashboard</h1>
          <p style={{ margin: "5px 0 0 0", color: "#666" }}>
            Welcome back,{" "}
            <strong>{authService.getCurrentUserEmail() || "User"}</strong>!
          </p>
        </div>
        <div style={{ display: "flex", gap: "10px" }}>
          <button
            onClick={handleClearCache}
            disabled={isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isLoading ? "not-allowed" : "pointer",
              opacity: isLoading ? 0.6 : 1,
            }}
            title="Clear local cache and fetch fresh data from the server"
          >
            {isLoading ? "..." : "🧹 Clear Cache"}
          </button>
          <button
            onClick={handleRefresh}
            disabled={isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isLoading ? "not-allowed" : "pointer",
              opacity: isLoading ? 0.6 : 1,
            }}
            title="Force a refresh from the server, bypassing the cache"
          >
            {isLoading ? "Refreshing..." : "🔄 Refresh"}
          </button>
          <button
            onClick={handleLogout}
            style={{
              padding: "8px 16px",
              backgroundColor: "#dc3545",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            🚪 Logout
          </button>
        </div>
      </div>

      {/* Error Message */}
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

      {dashboardData && (
        <>
          {/* Summary Cards */}
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
              gap: "20px",
              marginBottom: "30px",
            }}
          >
            {/* Total Files Card */}
            <div
              style={{
                backgroundColor: "#fff",
                padding: "20px",
                borderRadius: "8px",
                boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
                cursor: "pointer",
                transition: "transform 0.2s",
              }}
              onClick={handleViewAllFiles}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = "translateY(-2px)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = "translateY(0)";
              }}
            >
              <div
                style={{ display: "flex", alignItems: "center", gap: "15px" }}
              >
                <div
                  style={{
                    width: "50px",
                    height: "50px",
                    backgroundColor: "#e3f2fd",
                    borderRadius: "50%",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    fontSize: "24px",
                  }}
                >
                  📄
                </div>
                <div>
                  <h3 style={{ margin: 0, fontSize: "28px", color: "#007bff" }}>
                    {dashboardData.summary?.total_files || 0}
                  </h3>
                  <p style={{ margin: 0, color: "#666", fontSize: "14px" }}>
                    Total Files
                  </p>
                </div>
              </div>
            </div>

            {/* Total Folders Card */}
            <div
              style={{
                backgroundColor: "#fff",
                padding: "20px",
                borderRadius: "8px",
                boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
                cursor: "pointer",
                transition: "transform 0.2s",
              }}
              onClick={handleViewAllCollections}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = "translateY(-2px)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = "translateY(0)";
              }}
            >
              <div
                style={{ display: "flex", alignItems: "center", gap: "15px" }}
              >
                <div
                  style={{
                    width: "50px",
                    height: "50px",
                    backgroundColor: "#e8f5e9",
                    borderRadius: "50%",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    fontSize: "24px",
                  }}
                >
                  📁
                </div>
                <div>
                  <h3 style={{ margin: 0, fontSize: "28px", color: "#28a745" }}>
                    {dashboardData.summary?.total_folders || 0}
                  </h3>
                  <p style={{ margin: 0, color: "#666", fontSize: "14px" }}>
                    Total Folders
                  </p>
                </div>
              </div>
            </div>

            {/* Storage Usage Card */}
            <div
              style={{
                backgroundColor: "#fff",
                padding: "20px",
                borderRadius: "8px",
                boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
              }}
            >
              <div
                style={{ display: "flex", alignItems: "center", gap: "15px" }}
              >
                <div
                  style={{
                    width: "50px",
                    height: "50px",
                    backgroundColor: "#fff3cd",
                    borderRadius: "50%",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    fontSize: "24px",
                  }}
                >
                  💾
                </div>
                <div style={{ flex: 1 }}>
                  <h3 style={{ margin: 0, fontSize: "20px", color: "#333" }}>
                    {dashboardManager?.formatStorageValue(
                      dashboardData.summary?.storage_used,
                    ) || "0 Bytes"}
                  </h3>
                  <p style={{ margin: 0, color: "#666", fontSize: "14px" }}>
                    of{" "}
                    {dashboardManager?.formatStorageValue(
                      dashboardData.summary?.storage_limit,
                    ) || "0 Bytes"}{" "}
                    used
                  </p>
                </div>
              </div>

              {/* Storage Progress Bar */}
              <div style={{ marginTop: "15px" }}>
                <div
                  style={{
                    width: "100%",
                    height: "10px",
                    backgroundColor: "#e0e0e0",
                    borderRadius: "5px",
                    overflow: "hidden",
                  }}
                >
                  <div
                    style={{
                      width: `${dashboardData.summary?.storage_usage_percentage || 0}%`,
                      height: "100%",
                      backgroundColor: getStoragePercentageColor(
                        dashboardData.summary?.storage_usage_percentage || 0,
                      ),
                      transition: "width 0.3s ease",
                    }}
                  />
                </div>
                <p
                  style={{
                    margin: "5px 0 0 0",
                    textAlign: "right",
                    fontSize: "12px",
                    color: "#666",
                  }}
                >
                  {dashboardData.summary?.storage_usage_percentage || 0}% used
                </p>
              </div>
            </div>
          </div>

          {/* Storage Trend */}
          {dashboardData.storage_usage_trend &&
            dashboardData.storage_usage_trend.data_points &&
            dashboardData.storage_usage_trend.data_points.length > 0 && (
              <div
                style={{
                  backgroundColor: "#fff",
                  padding: "20px",
                  borderRadius: "8px",
                  boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
                  marginBottom: "30px",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    marginBottom: "20px",
                  }}
                >
                  <h2 style={{ margin: 0, fontSize: "20px" }}>
                    📈 Storage Usage Trend
                  </h2>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    {dashboardData.storage_usage_trend.data_points.length} data
                    point
                    {dashboardData.storage_usage_trend.data_points.length !== 1
                      ? "s"
                      : ""}
                  </div>
                </div>

                <p style={{ margin: "0 0 15px 0", color: "#666" }}>
                  {dashboardData.storage_usage_trend.period}
                </p>

                {/* Chart Container */}
                <div
                  style={{
                    position: "relative",
                    height: "200px",
                    backgroundColor: "#f8f9fa",
                    borderRadius: "8px",
                    padding: "20px",
                    border: "1px solid #e9ecef",
                  }}
                >
                  {/* Y-axis labels */}
                  <div
                    style={{
                      position: "absolute",
                      left: "5px",
                      top: "20px",
                      bottom: "40px",
                      width: "50px",
                      display: "flex",
                      flexDirection: "column",
                      justifyContent: "space-between",
                      fontSize: "10px",
                      color: "#666",
                    }}
                  >
                    {(() => {
                      const maxValue = Math.max(
                        ...dashboardData.storage_usage_trend.data_points.map(
                          (p) => p.usage?.value || 0,
                        ),
                      );
                      const maxUnit =
                        dashboardData.storage_usage_trend.data_points.find(
                          (p) => p.usage?.value === maxValue,
                        )?.usage?.unit || "KB";

                      // Create 5 evenly spaced labels from max to 0
                      const labels = [];
                      for (let i = 4; i >= 0; i--) {
                        const value = (maxValue * i) / 4;
                        labels.push(
                          <div key={i} style={{ textAlign: "right" }}>
                            {value.toFixed(0)} {maxUnit}
                          </div>,
                        );
                      }
                      return labels;
                    })()}
                  </div>

                  {/* Chart Area */}
                  <div
                    style={{
                      marginLeft: "60px",
                      height: "140px",
                      display: "flex",
                      alignItems: "flex-end",
                      justifyContent:
                        dashboardData.storage_usage_trend.data_points.length ===
                        1
                          ? "center"
                          : "space-between",
                      borderBottom: "2px solid #dee2e6",
                      borderLeft: "2px solid #dee2e6",
                      position: "relative",
                    }}
                  >
                    {/* Grid lines */}
                    {[1, 2, 3, 4].map((i) => (
                      <div
                        key={`grid-${i}`}
                        style={{
                          position: "absolute",
                          left: 0,
                          right: 0,
                          bottom: `${i * 25}%`,
                          height: "1px",
                          backgroundColor: "#e9ecef",
                          zIndex: 1,
                        }}
                      />
                    ))}

                    {/* Data bars */}
                    {dashboardData.storage_usage_trend.data_points.map(
                      (point, index) => {
                        const maxValue = Math.max(
                          ...dashboardData.storage_usage_trend.data_points.map(
                            (p) => p.usage?.value || 0,
                          ),
                        );

                        // Calculate percentage with minimum height for visibility
                        let percentage =
                          maxValue > 0
                            ? (point.usage?.value || 0) / maxValue
                            : 0;

                        // For single data point, show it as 80% of max height for better visual
                        if (
                          dashboardData.storage_usage_trend.data_points
                            .length === 1
                        ) {
                          percentage = 0.8; // 80% height
                        }

                        // Ensure minimum 5% height if there's any value
                        if (percentage > 0 && percentage < 0.05) {
                          percentage = 0.05;
                        }

                        const barHeight = Math.max(percentage * 100, 0);
                        const barWidth =
                          dashboardData.storage_usage_trend.data_points
                            .length === 1
                            ? "80px"
                            : "40px";

                        return (
                          <div
                            key={index}
                            style={{
                              display: "flex",
                              flexDirection: "column",
                              alignItems: "center",
                              gap: "8px",
                              position: "relative",
                              zIndex: 2,
                            }}
                            onMouseEnter={(e) => {
                              e.currentTarget.querySelector(
                                ".chart-bar",
                              ).style.backgroundColor = "#0056b3";
                              e.currentTarget.querySelector(
                                ".chart-bar",
                              ).style.transform = "scaleY(1.05)";
                            }}
                            onMouseLeave={(e) => {
                              e.currentTarget.querySelector(
                                ".chart-bar",
                              ).style.backgroundColor = "#007bff";
                              e.currentTarget.querySelector(
                                ".chart-bar",
                              ).style.transform = "scaleY(1)";
                            }}
                          >
                            {/* Value label on top of bar */}
                            <div
                              style={{
                                fontSize: "10px",
                                color: "#495057",
                                fontWeight: "bold",
                                textAlign: "center",
                                minHeight: "12px",
                                display: "flex",
                                alignItems: "center",
                              }}
                            >
                              {point.usage?.value
                                ? `${point.usage.value.toFixed(1)} ${point.usage.unit}`
                                : ""}
                            </div>

                            {/* Bar */}
                            <div
                              className="chart-bar"
                              style={{
                                width: barWidth,
                                height: `${barHeight}%`,
                                backgroundColor: "#007bff",
                                borderRadius: "4px 4px 0 0",
                                transition: "all 0.3s ease",
                                cursor: "pointer",
                                boxShadow: "0 2px 4px rgba(0,123,255,0.3)",
                                position: "relative",
                              }}
                              title={`${dashboardManager?.formatStorageValue(point.usage) || `${point.usage?.value || 0} ${point.usage?.unit || "KB"}`} on ${new Date(point.date).toLocaleDateString()}`}
                            >
                              {/* Gradient effect */}
                              <div
                                style={{
                                  position: "absolute",
                                  top: 0,
                                  left: 0,
                                  right: 0,
                                  height: "30%",
                                  background:
                                    "linear-gradient(to bottom, rgba(255,255,255,0.3), transparent)",
                                  borderRadius: "4px 4px 0 0",
                                }}
                              />
                            </div>
                          </div>
                        );
                      },
                    )}
                  </div>

                  {/* X-axis labels */}
                  <div
                    style={{
                      marginLeft: "60px",
                      marginTop: "10px",
                      display: "flex",
                      justifyContent:
                        dashboardData.storage_usage_trend.data_points.length ===
                        1
                          ? "center"
                          : "space-between",
                    }}
                  >
                    {dashboardData.storage_usage_trend.data_points.map(
                      (point, index) => (
                        <div
                          key={`date-${index}`}
                          style={{
                            fontSize: "11px",
                            color: "#666",
                            textAlign: "center",
                            fontWeight: "500",
                          }}
                        >
                          {new Date(point.date).toLocaleDateString("en-US", {
                            month: "short",
                            day: "numeric",
                          })}
                        </div>
                      ),
                    )}
                  </div>
                </div>

                {/* Legend/Summary */}
                <div
                  style={{
                    marginTop: "15px",
                    padding: "10px",
                    backgroundColor: "#f8f9fa",
                    borderRadius: "4px",
                    fontSize: "12px",
                    color: "#666",
                  }}
                >
                  {dashboardData.storage_usage_trend.data_points.length ===
                  1 ? (
                    <div>
                      <strong>Current Usage:</strong>{" "}
                      {dashboardManager?.formatStorageValue(
                        dashboardData.storage_usage_trend.data_points[0].usage,
                      ) ||
                        `${dashboardData.storage_usage_trend.data_points[0].usage?.value || 0} ${dashboardData.storage_usage_trend.data_points[0].usage?.unit || "KB"}`}
                      <br />
                      <em>
                        Note: More data points will be available as you use the
                        app over time.
                      </em>
                    </div>
                  ) : (
                    <div>
                      <strong>Trend:</strong>{" "}
                      {dashboardData.storage_usage_trend.data_points.length}{" "}
                      data points over{" "}
                      {dashboardData.storage_usage_trend.period?.toLowerCase() ||
                        "time period"}
                    </div>
                  )}
                </div>
              </div>
            )}

          {/* Recent Files */}
          {dashboardData.recent_files &&
            dashboardData.recent_files.length > 0 && (
              <div
                style={{
                  backgroundColor: "#fff",
                  padding: "20px",
                  borderRadius: "8px",
                  boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    marginBottom: "20px",
                  }}
                >
                  <h2 style={{ margin: 0, fontSize: "20px" }}>
                    🕒 Recent Files
                  </h2>
                  <button
                    onClick={handleViewAllFiles}
                    style={{
                      padding: "6px 12px",
                      backgroundColor: "transparent",
                      color: "#007bff",
                      border: "1px solid #007bff",
                      borderRadius: "4px",
                      cursor: "pointer",
                      fontSize: "14px",
                    }}
                  >
                    View All Files →
                  </button>
                </div>

                <div style={{ overflowX: "auto" }}>
                  <table style={{ width: "100%", borderCollapse: "collapse" }}>
                    <thead>
                      <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
                        <th
                          style={{
                            padding: "10px",
                            textAlign: "left",
                            color: "#666",
                            fontWeight: "normal",
                          }}
                        >
                          Name
                        </th>
                        <th
                          style={{
                            padding: "10px",
                            textAlign: "left",
                            color: "#666",
                            fontWeight: "normal",
                          }}
                        >
                          Type
                        </th>
                        <th
                          style={{
                            padding: "10px",
                            textAlign: "left",
                            color: "#666",
                            fontWeight: "normal",
                          }}
                        >
                          Size
                        </th>
                        <th
                          style={{
                            padding: "10px",
                            textAlign: "left",
                            color: "#666",
                            fontWeight: "normal",
                          }}
                        >
                          Modified
                        </th>
                        <th
                          style={{
                            padding: "10px",
                            textAlign: "center",
                            color: "#666",
                            fontWeight: "normal",
                          }}
                        >
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {dashboardData.recent_files.map((file) => (
                        <tr
                          key={file.id}
                          style={{ borderBottom: "1px solid #f0f0f0" }}
                        >
                          <td style={{ padding: "12px" }}>
                            <div
                              style={{
                                display: "flex",
                                alignItems: "center",
                                gap: "10px",
                              }}
                            >
                              <span style={{ fontSize: "20px" }}>📄</span>
                              <div>
                                <div>{file.name || "[Encrypted]"}</div>
                                {file._decryptionError && (
                                  <div
                                    style={{
                                      fontSize: "11px",
                                      color: "#dc3545",
                                      marginTop: "2px",
                                    }}
                                  >
                                    Unable to decrypt
                                  </div>
                                )}
                              </div>
                            </div>
                          </td>
                          <td style={{ padding: "12px", color: "#666" }}>
                            {file.mime_type || "Document"}
                          </td>
                          <td style={{ padding: "12px", color: "#666" }}>
                            {file.size
                              ? dashboardManager?.formatFileSize?.(file.size) ||
                                `${file.size} bytes`
                              : "Unknown"}
                          </td>
                          <td style={{ padding: "12px", color: "#666" }}>
                            {file.modified_at
                              ? new Date(file.modified_at).toLocaleString()
                              : "Unknown"}
                          </td>
                          <td style={{ padding: "12px", textAlign: "center" }}>
                            <button
                              onClick={() =>
                                handleDownloadFile(file.id, file.name)
                              }
                              disabled={
                                !file._isDecrypted ||
                                downloadingFiles.has(file.id)
                              }
                              style={{
                                padding: "4px 12px",
                                backgroundColor:
                                  !file._isDecrypted ||
                                  downloadingFiles.has(file.id)
                                    ? "#ccc"
                                    : "#007bff",
                                color: "white",
                                border: "none",
                                borderRadius: "4px",
                                cursor:
                                  !file._isDecrypted ||
                                  downloadingFiles.has(file.id)
                                    ? "not-allowed"
                                    : "pointer",
                                fontSize: "12px",
                              }}
                              title={
                                !file._isDecrypted
                                  ? "File cannot be decrypted"
                                  : "Download file"
                              }
                            >
                              {downloadingFiles.has(file.id) ? "..." : "↓"}
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

          {/* Empty State for Recent Files */}
          {(!dashboardData.recent_files ||
            dashboardData.recent_files.length === 0) && (
            <div
              style={{
                backgroundColor: "#fff",
                padding: "40px",
                borderRadius: "8px",
                boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
                textAlign: "center",
              }}
            >
              <div style={{ fontSize: "48px", marginBottom: "10px" }}>📁</div>
              <h3 style={{ margin: "0 0 10px 0", color: "#333" }}>
                No Recent Files
              </h3>
              <p style={{ margin: 0, color: "#666" }}>
                Upload some files to see them here
              </p>
            </div>
          )}
        </>
      )}
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(DashboardManagerExample);
