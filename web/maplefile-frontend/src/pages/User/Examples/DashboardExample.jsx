// File: src/pages/User/Examples/DashboardExample.jsx
// Example component demonstrating how to use the DashboardManager

import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "../../../services/Services";

const DashboardExample = () => {
  const navigate = useNavigate();
  const { authManager } = useAuth();

  // State management
  const [dashboardManager, setDashboardManager] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [dashboardData, setDashboardData] = useState(null);
  const [eventLog, setEventLog] = useState([]);
  const [managerStatus, setManagerStatus] = useState({});

  // Settings
  const [showDetails, setShowDetails] = useState(false);

  // Initialize dashboard manager
  useEffect(() => {
    const initializeManager = async () => {
      if (!authManager.isAuthenticated()) return;

      try {
        const { default: DashboardManager } = await import(
          "../../../services/Manager/DashboardManager.js"
        );

        const manager = new DashboardManager(authManager);
        await manager.initialize();

        setDashboardManager(manager);

        // Set up event listener
        manager.addDashboardListener(handleDashboardEvent);

        console.log("[Example] DashboardManager initialized");
      } catch (err) {
        console.error("[Example] Failed to initialize DashboardManager:", err);
        setError(`Failed to initialize: ${err.message}`);
      }
    };

    initializeManager();

    return () => {
      if (dashboardManager) {
        dashboardManager.removeDashboardListener(handleDashboardEvent);
      }
    };
  }, [authManager]);

  // Load manager status when manager is ready
  useEffect(() => {
    if (dashboardManager) {
      loadManagerStatus();
    }
  }, [dashboardManager]);

  // Handle dashboard events
  const handleDashboardEvent = useCallback((eventType, eventData) => {
    console.log("[Example] Dashboard event:", eventType, eventData);
    addToEventLog(eventType, eventData);
  }, []);

  // Load manager status
  const loadManagerStatus = useCallback(() => {
    if (!dashboardManager) return;

    const status = dashboardManager.getManagerStatus();
    setManagerStatus(status);

    console.log("[Example] Manager status:", status);
  }, [dashboardManager]);

  // Load dashboard data
  const handleLoadDashboard = async (forceRefresh = false) => {
    if (!dashboardManager) {
      setError("Dashboard manager not initialized");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      console.log("[Example] Loading dashboard data:", { forceRefresh });

      const data = await dashboardManager.getDashboardData(forceRefresh);

      setDashboardData(data);
      setSuccess(`Dashboard data loaded successfully`);

      addToEventLog("dashboard_data_loaded", {
        totalFiles: data.summary?.totalFiles || 0,
        totalFolders: data.summary?.totalFolders || 0,
        storageUsed: data.summary?.storageUsed?.value || 0,
        recentFilesCount: data.recentFiles?.length || 0,
        forceRefresh,
      });

      // Update manager status
      loadManagerStatus();
    } catch (err) {
      console.error("[Example] Failed to load dashboard:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Clear caches
  const handleClearCaches = () => {
    if (!dashboardManager) return;

    dashboardManager.clearAllCaches();
    setDashboardData(null);
    setSuccess("All caches cleared");

    addToEventLog("caches_cleared", {
      timestamp: new Date().toISOString(),
    });

    loadManagerStatus();
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

  if (!authManager.isAuthenticated()) {
    return (
      <div style={{ padding: "20px", textAlign: "center" }}>
        <p>Please log in to access this example.</p>
        <button onClick={() => navigate("/login")}>Go to Login</button>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "1400px", margin: "0 auto" }}>
      <div style={{ marginBottom: "20px" }}>
        <button onClick={() => navigate("/dashboard")}>
          ← Back to Dashboard
        </button>
      </div>

      <h2>📊 Dashboard Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This example demonstrates fetching and displaying dashboard data with
        caching.
        <br />
        <strong>Features:</strong> Dashboard summary, storage usage, recent
        files, cache management
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
            <strong>Can Get Dashboard:</strong>{" "}
            {managerStatus.canGetDashboard ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Has Data:</strong> {dashboardData ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Listeners:</strong> {managerStatus.listenerCount || 0}
          </div>
          <div>
            <strong>Cache Valid:</strong>{" "}
            {managerStatus.storage?.isValid ? "✅ Yes" : "❌ No"}
          </div>
        </div>
      </div>

      {/* Controls */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e3f2fd",
          borderRadius: "6px",
          border: "1px solid #bbdefb",
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
          <h4 style={{ margin: 0 }}>🔧 Controls:</h4>
          <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
            <label
              style={{ display: "flex", alignItems: "center", gap: "8px" }}
            >
              <input
                type="checkbox"
                checked={showDetails}
                onChange={(e) => setShowDetails(e.target.checked)}
              />
              Show Details
            </label>
          </div>
        </div>

        <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
          <button
            onClick={() => handleLoadDashboard(false)}
            disabled={isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isLoading ? "not-allowed" : "pointer",
            }}
          >
            {isLoading ? "🔄 Loading..." : "📊 Load Dashboard"}
          </button>
          <button
            onClick={() => handleLoadDashboard(true)}
            disabled={isLoading}
            style={{
              padding: "8px 16px",
              backgroundColor: "#ffc107",
              color: "#212529",
              border: "none",
              borderRadius: "4px",
              cursor: isLoading ? "not-allowed" : "pointer",
            }}
          >
            🔄 Force Refresh
          </button>
          <button
            onClick={handleClearCaches}
            disabled={!dashboardManager}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: !dashboardManager ? "not-allowed" : "pointer",
            }}
          >
            🗑️ Clear Cache
          </button>
        </div>
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

      {/* Dashboard Data Display */}
      {dashboardData && (
        <div>
          {/* Summary Section */}
          <div
            style={{
              marginBottom: "30px",
              padding: "20px",
              backgroundColor: "#fff",
              borderRadius: "8px",
              border: "1px solid #dee2e6",
            }}
          >
            <h3 style={{ margin: "0 0 20px 0" }}>📊 Dashboard Summary</h3>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
                gap: "20px",
              }}
            >
              <div style={{ textAlign: "center" }}>
                <div
                  style={{
                    fontSize: "32px",
                    fontWeight: "bold",
                    color: "#007bff",
                  }}
                >
                  {dashboardData.summary?.totalFiles || 0}
                </div>
                <div style={{ color: "#666" }}>Total Files</div>
              </div>
              <div style={{ textAlign: "center" }}>
                <div
                  style={{
                    fontSize: "32px",
                    fontWeight: "bold",
                    color: "#28a745",
                  }}
                >
                  {dashboardData.summary?.totalFolders || 0}
                </div>
                <div style={{ color: "#666" }}>Total Folders</div>
              </div>
              <div style={{ textAlign: "center" }}>
                <div
                  style={{
                    fontSize: "24px",
                    fontWeight: "bold",
                    color: "#ffc107",
                  }}
                >
                  {dashboardManager?.formatStorageValue(
                    dashboardData.summary?.storageUsed,
                  ) || "0 Bytes"}
                </div>
                <div style={{ color: "#666" }}>Storage Used</div>
              </div>
              <div style={{ textAlign: "center" }}>
                <div
                  style={{
                    fontSize: "24px",
                    fontWeight: "bold",
                    color: "#6c757d",
                  }}
                >
                  {dashboardManager?.formatStorageValue(
                    dashboardData.summary?.storageLimit,
                  ) || "0 Bytes"}
                </div>
                <div style={{ color: "#666" }}>Storage Limit</div>
              </div>
            </div>

            {/* Storage Usage Bar */}
            {dashboardData.summary?.storageUsagePercentage !== undefined && (
              <div style={{ marginTop: "20px" }}>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    marginBottom: "5px",
                  }}
                >
                  <span>Storage Usage</span>
                  <span>{dashboardData.summary.storageUsagePercentage}%</span>
                </div>
                <div
                  style={{
                    width: "100%",
                    height: "20px",
                    backgroundColor: "#e0e0e0",
                    borderRadius: "10px",
                    overflow: "hidden",
                  }}
                >
                  <div
                    style={{
                      width: `${dashboardData.summary.storageUsagePercentage}%`,
                      height: "100%",
                      backgroundColor:
                        dashboardData.summary.storageUsagePercentage > 80
                          ? "#dc3545"
                          : dashboardData.summary.storageUsagePercentage > 60
                            ? "#ffc107"
                            : "#28a745",
                      transition: "width 0.3s ease",
                    }}
                  />
                </div>
              </div>
            )}
          </div>

          {/* Storage Trend Section */}
          {dashboardData.storageUsageTrend && (
            <div
              style={{
                marginBottom: "30px",
                padding: "20px",
                backgroundColor: "#fff",
                borderRadius: "8px",
                border: "1px solid #dee2e6",
              }}
            >
              <h3 style={{ margin: "0 0 15px 0" }}>📈 Storage Usage Trend</h3>
              <p style={{ margin: "0 0 15px 0", color: "#666" }}>
                Period: {dashboardData.storageUsageTrend.period}
              </p>
              {dashboardData.storageUsageTrend.dataPoints &&
              dashboardData.storageUsageTrend.dataPoints.length > 0 ? (
                <div>
                  <table style={{ width: "100%", borderCollapse: "collapse" }}>
                    <thead>
                      <tr style={{ backgroundColor: "#f8f9fa" }}>
                        <th
                          style={{
                            padding: "10px",
                            textAlign: "left",
                            border: "1px solid #dee2e6",
                          }}
                        >
                          Date
                        </th>
                        <th
                          style={{
                            padding: "10px",
                            textAlign: "left",
                            border: "1px solid #dee2e6",
                          }}
                        >
                          Usage
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {dashboardData.storageUsageTrend.dataPoints.map(
                        (point, index) => (
                          <tr key={index}>
                            <td
                              style={{
                                padding: "10px",
                                border: "1px solid #dee2e6",
                              }}
                            >
                              {new Date(point.date).toLocaleDateString()}
                            </td>
                            <td
                              style={{
                                padding: "10px",
                                border: "1px solid #dee2e6",
                              }}
                            >
                              {dashboardManager?.formatStorageValue(
                                point.usage,
                              ) || "N/A"}
                            </td>
                          </tr>
                        ),
                      )}
                    </tbody>
                  </table>
                </div>
              ) : (
                <p style={{ color: "#666", fontStyle: "italic" }}>
                  No trend data available
                </p>
              )}
            </div>
          )}

          {/* Recent Files Section */}
          {dashboardData.recentFiles &&
            dashboardData.recentFiles.length > 0 && (
              <div
                style={{
                  marginBottom: "30px",
                  padding: "20px",
                  backgroundColor: "#fff",
                  borderRadius: "8px",
                  border: "1px solid #dee2e6",
                }}
              >
                <h3 style={{ margin: "0 0 15px 0" }}>
                  🕒 Recent Files ({dashboardData.recentFiles.length})
                </h3>
                <table style={{ width: "100%", borderCollapse: "collapse" }}>
                  <thead>
                    <tr style={{ backgroundColor: "#f8f9fa" }}>
                      <th
                        style={{
                          padding: "10px",
                          textAlign: "left",
                          border: "1px solid #dee2e6",
                        }}
                      >
                        File
                      </th>
                      <th
                        style={{
                          padding: "10px",
                          textAlign: "left",
                          border: "1px solid #dee2e6",
                        }}
                      >
                        Type
                      </th>
                      <th
                        style={{
                          padding: "10px",
                          textAlign: "left",
                          border: "1px solid #dee2e6",
                        }}
                      >
                        Size
                      </th>
                      <th
                        style={{
                          padding: "10px",
                          textAlign: "left",
                          border: "1px solid #dee2e6",
                        }}
                      >
                        Uploaded
                      </th>
                      {showDetails && (
                        <>
                          <th
                            style={{
                              padding: "10px",
                              textAlign: "left",
                              border: "1px solid #dee2e6",
                            }}
                          >
                            Timestamp
                          </th>
                        </>
                      )}
                    </tr>
                  </thead>
                  <tbody>
                    {dashboardData.recentFiles.map((file, index) => (
                      <tr key={index}>
                        <td
                          style={{
                            padding: "10px",
                            border: "1px solid #dee2e6",
                          }}
                        >
                          <div
                            style={{
                              display: "flex",
                              alignItems: "center",
                              gap: "10px",
                            }}
                          >
                            <span style={{ fontSize: "18px" }}>📄</span>
                            <span>{file.fileName || "[Unknown]"}</span>
                          </div>
                        </td>
                        <td
                          style={{
                            padding: "10px",
                            border: "1px solid #dee2e6",
                          }}
                        >
                          {dashboardManager?.getFileType(file.fileName) ||
                            file.type ||
                            "Document"}
                        </td>
                        <td
                          style={{
                            padding: "10px",
                            border: "1px solid #dee2e6",
                          }}
                        >
                          {file.size
                            ? dashboardManager?.formatStorageValue(file.size)
                            : "Unknown"}
                        </td>
                        <td
                          style={{
                            padding: "10px",
                            border: "1px solid #dee2e6",
                          }}
                        >
                          {file.uploaded || "Unknown"}
                        </td>
                        {showDetails && (
                          <td
                            style={{
                              padding: "10px",
                              border: "1px solid #dee2e6",
                              fontSize: "12px",
                              color: "#666",
                            }}
                          >
                            {file.uploadedTimestamp || "N/A"}
                          </td>
                        )}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

          {/* Debug Information */}
          {showDetails && (
            <div
              style={{
                marginBottom: "30px",
                padding: "20px",
                backgroundColor: "#f8f9fa",
                borderRadius: "8px",
                border: "1px solid #dee2e6",
              }}
            >
              <h3 style={{ margin: "0 0 15px 0" }}>🔧 Debug Information</h3>
              <pre
                style={{ fontSize: "12px", whiteSpace: "pre-wrap", margin: 0 }}
              >
                {JSON.stringify(dashboardData, null, 2)}
              </pre>
            </div>
          )}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !dashboardData && (
        <div
          style={{
            textAlign: "center",
            padding: "60px 20px",
            backgroundColor: "#f8f9fa",
            borderRadius: "8px",
            border: "2px dashed #dee2e6",
          }}
        >
          <div style={{ fontSize: "64px", marginBottom: "20px" }}>📊</div>
          <h3>No dashboard data loaded</h3>
          <p style={{ color: "#666" }}>
            Click "Load Dashboard" to fetch your dashboard data.
          </p>
        </div>
      )}

      {/* Event Log */}
      <div style={{ marginTop: "40px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <h3>📋 Dashboard Event Log ({eventLog.length})</h3>
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
              No dashboard events logged yet.
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

export default DashboardExample;
