// File: src/pages/User/Dashboard/Dashboard.jsx
// Updated to showcase enhanced file management with version control and sync capabilities
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useCollections from "../../../hooks/useCollections.js";
import useAuth from "../../../hooks/useAuth.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";
import FileSyncManager from "../../../components/FileSyncManager.jsx";

const Dashboard = () => {
  const navigate = useNavigate();
  const { authService, fileService } = useServices();
  const { user } = useAuth();
  const {
    collections,
    allCollections,
    isLoading: collectionsLoading,
    loadAllCollections,
    getCollectionStats,
  } = useCollections();

  const [fileStats, setFileStats] = useState(null);
  const [recentActivity, setRecentActivity] = useState([]);
  const [isLoadingFiles, setIsLoadingFiles] = useState(false);
  const [showSyncManager, setShowSyncManager] = useState(false);
  const [dashboardStats, setDashboardStats] = useState({
    totalFiles: 0,
    activeFiles: 0,
    archivedFiles: 0,
    deletedFiles: 0,
    pendingFiles: 0,
    filesWithTombstones: 0,
    storageUsed: 0,
  });

  // Load dashboard data
  useEffect(() => {
    if (authService.isAuthenticated()) {
      loadDashboardData();
    }
  }, [authService]);

  const loadDashboardData = async () => {
    try {
      // Load collections
      await loadAllCollections();

      // Load file statistics
      await loadFileStatistics();

      // Load recent activity
      await loadRecentActivity();
    } catch (error) {
      console.error("[Dashboard] Failed to load dashboard data:", error);
    }
  };

  const loadFileStatistics = async () => {
    setIsLoadingFiles(true);
    try {
      // Get overall file statistics by syncing recent data
      const syncResponse = await fileService.syncFiles(null, 1000);

      if (syncResponse.files) {
        const stats = calculateFileStats(syncResponse.files);
        setFileStats(stats);
        setDashboardStats((prev) => ({ ...prev, ...stats }));
      }
    } catch (error) {
      console.error("[Dashboard] Failed to load file statistics:", error);
    } finally {
      setIsLoadingFiles(false);
    }
  };

  const calculateFileStats = (files) => {
    const stats = {
      totalFiles: files.length,
      activeFiles: 0,
      archivedFiles: 0,
      deletedFiles: 0,
      pendingFiles: 0,
      filesWithTombstones: 0,
      storageUsed: 0,
      recentFiles: 0,
    };

    const oneWeekAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);

    files.forEach((file) => {
      // Normalize file state
      const state = file.state || "active";

      switch (state) {
        case "active":
          stats.activeFiles++;
          break;
        case "archived":
          stats.archivedFiles++;
          break;
        case "deleted":
          stats.deletedFiles++;
          break;
        case "pending":
          stats.pendingFiles++;
          break;
      }

      if ((file.tombstone_version || 0) > 0) {
        stats.filesWithTombstones++;
      }

      // Calculate storage usage
      const fileSize = file.encrypted_file_size_in_bytes || 0;
      stats.storageUsed += fileSize;

      // Count recent files
      if (file.created_at && new Date(file.created_at) > oneWeekAgo) {
        stats.recentFiles++;
      }
    });

    return stats;
  };

  const loadRecentActivity = async () => {
    try {
      // Get recent files activity (last 50 files)
      const syncResponse = await fileService.syncFiles(null, 50);

      if (syncResponse.files) {
        const activity = syncResponse.files
          .filter((file) => file.modified_at || file.created_at)
          .sort((a, b) => {
            const aTime = new Date(a.modified_at || a.created_at);
            const bTime = new Date(b.modified_at || b.created_at);
            return bTime - aTime;
          })
          .slice(0, 10)
          .map((file) => ({
            id: file.id,
            type: getActivityType(file),
            timestamp: file.modified_at || file.created_at,
            fileId: file.id,
            fileName: "File", // Will show encrypted since we don't have collection keys here
            collectionId: file.collection_id,
            state: file.state,
            version: file.version,
          }));

        setRecentActivity(activity);
      }
    } catch (error) {
      console.error("[Dashboard] Failed to load recent activity:", error);
    }
  };

  const getActivityType = (file) => {
    if (file.state === "deleted") return "deleted";
    if (file.state === "archived") return "archived";
    if (file.state === "pending") return "uploading";
    if ((file.version || 1) > 1) return "updated";
    return "created";
  };

  const getActivityIcon = (type) => {
    switch (type) {
      case "created":
        return "➕";
      case "updated":
        return "📝";
      case "deleted":
        return "🗑️";
      case "archived":
        return "📦";
      case "uploading":
        return "⏳";
      default:
        return "📄";
    }
  };

  const getActivityColor = (type) => {
    switch (type) {
      case "created":
        return "#28a745";
      case "updated":
        return "#17a2b8";
      case "deleted":
        return "#dc3545";
      case "archived":
        return "#6c757d";
      case "uploading":
        return "#ffc107";
      default:
        return "#6c757d";
    }
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const formatTimeAgo = (timestamp) => {
    if (!timestamp) return "Unknown";
    const now = new Date();
    const time = new Date(timestamp);
    const diffMs = now - time;
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return time.toLocaleDateString();
  };

  const handleSyncComplete = (result) => {
    console.log("[Dashboard] Sync completed:", result);
    // Reload dashboard data after sync
    loadDashboardData();
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      {/* Header */}
      <div style={{ marginBottom: "30px" }}>
        <h1 style={{ margin: "0 0 10px 0" }}>Dashboard</h1>
        <p style={{ color: "#666", margin: 0 }}>
          Welcome back, {user?.email || "User"}! Here's your file management
          overview.
        </p>
      </div>

      {/* Stats Grid */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
          gap: "20px",
          marginBottom: "30px",
        }}
      >
        {/* Collections Stats */}
        <div
          style={{
            padding: "20px",
            backgroundColor: "white",
            borderRadius: "8px",
            border: "1px solid #dee2e6",
            textAlign: "center",
          }}
        >
          <div
            style={{ fontSize: "32px", color: "#007bff", marginBottom: "10px" }}
          >
            📁
          </div>
          <div
            style={{
              fontSize: "24px",
              fontWeight: "bold",
              marginBottom: "5px",
            }}
          >
            {allCollections.length}
          </div>
          <div style={{ fontSize: "14px", color: "#666" }}>Collections</div>
        </div>

        {/* Active Files */}
        <div
          style={{
            padding: "20px",
            backgroundColor: "white",
            borderRadius: "8px",
            border: "1px solid #dee2e6",
            textAlign: "center",
          }}
        >
          <div
            style={{ fontSize: "32px", color: "#28a745", marginBottom: "10px" }}
          >
            📄
          </div>
          <div
            style={{
              fontSize: "24px",
              fontWeight: "bold",
              marginBottom: "5px",
            }}
          >
            {dashboardStats.activeFiles}
          </div>
          <div style={{ fontSize: "14px", color: "#666" }}>Active Files</div>
        </div>

        {/* Storage Used */}
        <div
          style={{
            padding: "20px",
            backgroundColor: "white",
            borderRadius: "8px",
            border: "1px solid #dee2e6",
            textAlign: "center",
          }}
        >
          <div
            style={{ fontSize: "32px", color: "#17a2b8", marginBottom: "10px" }}
          >
            💾
          </div>
          <div
            style={{
              fontSize: "24px",
              fontWeight: "bold",
              marginBottom: "5px",
            }}
          >
            {formatFileSize(dashboardStats.storageUsed)}
          </div>
          <div style={{ fontSize: "14px", color: "#666" }}>Storage Used</div>
        </div>

        {/* Recent Activity */}
        <div
          style={{
            padding: "20px",
            backgroundColor: "white",
            borderRadius: "8px",
            border: "1px solid #dee2e6",
            textAlign: "center",
          }}
        >
          <div
            style={{ fontSize: "32px", color: "#ffc107", marginBottom: "10px" }}
          >
            ⚡
          </div>
          <div
            style={{
              fontSize: "24px",
              fontWeight: "bold",
              marginBottom: "5px",
            }}
          >
            {fileStats?.recentFiles || 0}
          </div>
          <div style={{ fontSize: "14px", color: "#666" }}>Files This Week</div>
        </div>
      </div>

      {/* File States Overview */}
      {fileStats && (
        <div
          style={{
            padding: "20px",
            backgroundColor: "white",
            borderRadius: "8px",
            border: "1px solid #dee2e6",
            marginBottom: "30px",
          }}
        >
          <h3 style={{ margin: "0 0 15px 0" }}>📊 File States Overview</h3>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(120px, 1fr))",
              gap: "15px",
            }}
          >
            <div style={{ textAlign: "center" }}>
              <div
                style={{
                  fontSize: "20px",
                  fontWeight: "bold",
                  color: "#28a745",
                }}
              >
                {dashboardStats.activeFiles}
              </div>
              <div style={{ fontSize: "12px", color: "#666" }}>Active</div>
            </div>
            <div style={{ textAlign: "center" }}>
              <div
                style={{
                  fontSize: "20px",
                  fontWeight: "bold",
                  color: "#6c757d",
                }}
              >
                {dashboardStats.archivedFiles}
              </div>
              <div style={{ fontSize: "12px", color: "#666" }}>Archived</div>
            </div>
            <div style={{ textAlign: "center" }}>
              <div
                style={{
                  fontSize: "20px",
                  fontWeight: "bold",
                  color: "#dc3545",
                }}
              >
                {dashboardStats.deletedFiles}
              </div>
              <div style={{ fontSize: "12px", color: "#666" }}>Deleted</div>
            </div>
            <div style={{ textAlign: "center" }}>
              <div
                style={{
                  fontSize: "20px",
                  fontWeight: "bold",
                  color: "#ffc107",
                }}
              >
                {dashboardStats.pendingFiles}
              </div>
              <div style={{ fontSize: "12px", color: "#666" }}>Pending</div>
            </div>
            <div style={{ textAlign: "center" }}>
              <div
                style={{
                  fontSize: "20px",
                  fontWeight: "bold",
                  color: "#fd7e14",
                }}
              >
                {dashboardStats.filesWithTombstones}
              </div>
              <div style={{ fontSize: "12px", color: "#666" }}>Tombstones</div>
            </div>
          </div>
        </div>
      )}

      {/* File Sync Manager */}
      <div style={{ marginBottom: "30px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "15px",
          }}
        >
          <h3 style={{ margin: 0 }}>File Synchronization</h3>
          <button
            onClick={() => setShowSyncManager(!showSyncManager)}
            style={{
              padding: "6px 12px",
              fontSize: "14px",
              backgroundColor: showSyncManager ? "#dc3545" : "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            {showSyncManager ? "Hide Details" : "Show Details"}
          </button>
        </div>

        <FileSyncManager
          onSyncComplete={handleSyncComplete}
          showDetailed={showSyncManager}
        />
      </div>

      {/* Recent Activity */}
      <div
        style={{
          padding: "20px",
          backgroundColor: "white",
          borderRadius: "8px",
          border: "1px solid #dee2e6",
          marginBottom: "30px",
        }}
      >
        <h3 style={{ margin: "0 0 15px 0" }}>🕒 Recent Activity</h3>

        {recentActivity.length === 0 ? (
          <p style={{ color: "#666", textAlign: "center", padding: "20px" }}>
            No recent activity found.
          </p>
        ) : (
          <div
            style={{ display: "flex", flexDirection: "column", gap: "10px" }}
          >
            {recentActivity.map((activity) => (
              <div
                key={`${activity.fileId}_${activity.timestamp}`}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "12px",
                  padding: "10px",
                  borderRadius: "4px",
                  backgroundColor: "#f8f9fa",
                  border: "1px solid #e9ecef",
                }}
              >
                <span style={{ fontSize: "20px" }}>
                  {getActivityIcon(activity.type)}
                </span>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: "bold", marginBottom: "2px" }}>
                    {activity.fileName}
                    {activity.version > 1 && (
                      <span
                        style={{
                          fontSize: "12px",
                          color: "#666",
                          marginLeft: "8px",
                        }}
                      >
                        v{activity.version}
                      </span>
                    )}
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <span
                      style={{
                        color: getActivityColor(activity.type),
                        fontWeight: "bold",
                      }}
                    >
                      {activity.type.charAt(0).toUpperCase() +
                        activity.type.slice(1)}
                    </span>
                    {" • "}
                    ID: {activity.fileId.substring(0, 8)}...
                    {" • "}
                    State: {activity.state}
                  </div>
                </div>
                <div
                  style={{
                    fontSize: "12px",
                    color: "#999",
                    textAlign: "right",
                  }}
                >
                  {formatTimeAgo(activity.timestamp)}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Quick Actions */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
          gap: "20px",
          marginBottom: "30px",
        }}
      >
        <div
          style={{
            padding: "20px",
            backgroundColor: "white",
            borderRadius: "8px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>📁 Collections</h4>
          <p style={{ fontSize: "14px", color: "#666", marginBottom: "15px" }}>
            Manage your encrypted file collections
          </p>
          <div style={{ display: "flex", gap: "10px" }}>
            <button
              onClick={() => navigate("/collections")}
              style={{
                padding: "8px 16px",
                backgroundColor: "#007bff",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
                fontSize: "14px",
              }}
            >
              View All
            </button>
            <button
              onClick={() => navigate("/collections/create")}
              style={{
                padding: "8px 16px",
                backgroundColor: "#28a745",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
                fontSize: "14px",
              }}
            >
              Create New
            </button>
          </div>
        </div>

        <div
          style={{
            padding: "20px",
            backgroundColor: "white",
            borderRadius: "8px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>⚙️ Account Settings</h4>
          <p style={{ fontSize: "14px", color: "#666", marginBottom: "15px" }}>
            Manage your profile and security settings
          </p>
          <button
            onClick={() => navigate("/me")}
            style={{
              padding: "8px 16px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
              fontSize: "14px",
            }}
          >
            Manage Account
          </button>
        </div>
      </div>

      {/* Info Box */}
      <div
        style={{
          padding: "20px",
          backgroundColor: "#e8f4fd",
          borderRadius: "8px",
          border: "1px solid #bee5eb",
          borderLeft: "4px solid #17a2b8",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0", color: "#0c5460" }}>
          🔒 Enhanced File Management Features
        </h4>
        <div style={{ color: "#0c5460", lineHeight: "1.6" }}>
          <p style={{ marginBottom: "10px" }}>
            Your files now support advanced version control and lifecycle
            management:
          </p>
          <ul style={{ marginLeft: "20px", marginBottom: "15px" }}>
            <li>
              <strong>Version Control:</strong> Each file change is tracked with
              version numbers
            </li>
            <li>
              <strong>State Management:</strong> Files can be active, archived,
              deleted, or pending
            </li>
            <li>
              <strong>Soft Deletion:</strong> Deleted files create tombstones
              for recovery
            </li>
            <li>
              <strong>Conflict Resolution:</strong> Version conflicts are
              detected and manageable
            </li>
            <li>
              <strong>Sync Status:</strong> Real-time synchronization across
              devices
            </li>
          </ul>
          <p style={{ marginBottom: 0 }}>
            All files remain end-to-end encrypted with your password-based keys.
          </p>
        </div>
      </div>
    </div>
  );
};

const ProtectedDashboard = withPasswordProtection(Dashboard, {
  showLoadingWhileChecking: true,
  checkInterval: 30000, // Check every 30 seconds
});

export default ProtectedDashboard;
