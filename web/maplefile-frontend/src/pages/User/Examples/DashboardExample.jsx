// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/DashboardExample.jsx
// Main Dashboard page - Shows summary, storage trend, and recent files

import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useFiles } from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";

const DashboardExample = () => {
  const navigate = useNavigate();
  const {
    authService,
    downloadFileManager,
    getCollectionManager,
    listCollectionManager,
  } = useFiles();

  // State management
  const [dashboardManager, setDashboardManager] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [dashboardData, setDashboardData] = useState(null);
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());

  // Initialize dashboard manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authService.isAuthenticated()) return;

      try {
        const { default: DashboardManager } = await import(
          "../../../services/Manager/DashboardManager.js"
        );

        const manager = new DashboardManager(
          authService,
          getCollectionManager,
          listCollectionManager,
        );
        await manager.initialize();

        setDashboardManager(manager);
        console.log("[Dashboard] DashboardManager initialized");
      } catch (err) {
        console.error(
          "[Dashboard] Failed to initialize DashboardManager:",
          err,
        );
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();
  }, [authService, getCollectionManager, listCollectionManager]);

  // Load dashboard data
  const loadDashboardData = useCallback(
    async (forceRefresh = false) => {
      if (!dashboardManager) return;

      setIsLoading(true);
      setError("");

      try {
        const data = await dashboardManager.getDashboardData(forceRefresh);
        setDashboardData(data);
      } catch (err) {
        console.error("[Dashboard] Failed to load dashboard:", err);
        setError(err.message);
      } finally {
        setIsLoading(false);
      }
    },
    [dashboardManager],
  );

  // Load dashboard data when manager is ready
  useEffect(() => {
    if (dashboardManager) {
      loadDashboardData();
    }
  }, [dashboardManager, loadDashboardData]);

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
          <h1 style={{ margin: 0 }}>Dashboard</h1>
          <p style={{ margin: "5px 0 0 0", color: "#666" }}>
            Welcome back, {authService.getCurrentUserEmail() || "User"}
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
            {isLoading ? "Refreshing..." : "🔄 Force Refresh"}
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
                <h2 style={{ margin: "0 0 20px 0", fontSize: "20px" }}>
                  📈 Storage Usage Trend
                </h2>
                <p style={{ margin: "0 0 15px 0", color: "#666" }}>
                  {dashboardData.storage_usage_trend.period}
                </p>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "flex-end",
                    height: "150px",
                    borderBottom: "2px solid #e0e0e0",
                    paddingBottom: "10px",
                  }}
                >
                  {dashboardData.storage_usage_trend.data_points.map(
                    (point, index) => {
                      const maxValue = Math.max(
                        ...dashboardData.storage_usage_trend.data_points.map(
                          (p) => p.usage?.value || 0,
                        ),
                      );
                      const percentage =
                        maxValue > 0 ? (point.usage?.value || 0) / maxValue : 0;

                      return (
                        <div
                          key={index}
                          style={{
                            flex: 1,
                            display: "flex",
                            flexDirection: "column",
                            alignItems: "center",
                            gap: "5px",
                          }}
                        >
                          <div
                            style={{
                              width: "40px",
                              height: `${percentage * 100}%`,
                              backgroundColor: "#007bff",
                              borderRadius: "4px 4px 0 0",
                              transition: "height 0.3s ease",
                            }}
                            title={dashboardManager?.formatStorageValue(
                              point.usage,
                            )}
                          />
                          <div
                            style={{
                              fontSize: "10px",
                              color: "#666",
                              textAlign: "center",
                            }}
                          >
                            {new Date(point.date).toLocaleDateString("en-US", {
                              month: "short",
                              day: "numeric",
                            })}
                          </div>
                        </div>
                      );
                    },
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
                            {file.type || "Document"}
                          </td>
                          <td style={{ padding: "12px", color: "#666" }}>
                            {file.size
                              ? dashboardManager?.formatStorageValue(file.size)
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
export default withPasswordProtection(DashboardExample);
