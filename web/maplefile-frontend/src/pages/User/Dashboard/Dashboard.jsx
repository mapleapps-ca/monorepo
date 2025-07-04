// File: src/pages/User/Dashboard/Dashboard.jsx
// Updated to showcase the consolidated file manager
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import useAuth from "../../../hooks/useAuth.js";
import useCollections from "../../../hooks/useCollections.js";
import useFiles from "../../../hooks/useFiles.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";
import FileSyncManager from "../../../components/FileSyncManager.jsx";

const Dashboard = () => {
  const navigate = useNavigate();
  const { user, logout, getDebugInfo } = useAuth();

  // Get some basic stats for the dashboard
  const {
    collections,
    sharedCollections,
    isLoading: collectionsLoading,
    loadAllCollections,
  } = useCollections();

  const [showQuickSync, setShowQuickSync] = useState(false);
  const [showDebugInfo, setShowDebugInfo] = useState(false);
  const [dashboardStats, setDashboardStats] = useState({
    totalFolders: 0,
    sharedFolders: 0,
    recentActivity: [],
  });

  // Load collections for dashboard stats
  useEffect(() => {
    const loadDashboardData = async () => {
      try {
        await loadAllCollections();
      } catch (error) {
        console.warn(
          "[Dashboard] Could not load collections for stats:",
          error,
        );
      }
    };

    loadDashboardData();
  }, [loadAllCollections]);

  // Update stats when collections change
  useEffect(() => {
    setDashboardStats({
      totalFolders: collections.length,
      sharedFolders: sharedCollections.length,
      recentActivity: [
        ...collections.slice(0, 3).map((c) => ({
          type: "folder",
          name: c.name || "[Encrypted]",
          id: c.id,
          date: c.modified_at || c.created_at,
        })),
        ...sharedCollections.slice(0, 2).map((c) => ({
          type: "shared_folder",
          name: c.name || "[Encrypted]",
          id: c.id,
          date: c.modified_at || c.created_at,
        })),
      ]
        .sort((a, b) => new Date(b.date) - new Date(a.date))
        .slice(0, 5),
    });
  }, [collections, sharedCollections]);

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  const handleSyncComplete = (result) => {
    console.log("[Dashboard] Sync completed:", result);
    // Reload collections to reflect any changes
    loadAllCollections();
  };

  const formatDate = (dateString) => {
    if (!dateString) return "Unknown";
    try {
      return new Date(dateString).toLocaleDateString();
    } catch {
      return "Unknown";
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <div style={{ marginBottom: "30px" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "20px",
          }}
        >
          <div>
            <h1 style={{ margin: 0 }}>🏠 Dashboard</h1>
            <p style={{ margin: "5px 0 0 0", color: "#666" }}>
              Welcome back, <strong>{user?.email || "User"}</strong>!
            </p>
          </div>

          <div style={{ display: "flex", gap: "10px" }}>
            <button
              onClick={() => setShowQuickSync(!showQuickSync)}
              style={{
                padding: "8px 16px",
                backgroundColor: showQuickSync ? "#dc3545" : "#17a2b8",
                color: "white",
                border: "none",
                borderRadius: "4px",
              }}
            >
              {showQuickSync ? "Hide Sync" : "🔄 Quick Sync"}
            </button>

            {import.meta.env.DEV && (
              <button
                onClick={() => setShowDebugInfo(!showDebugInfo)}
                style={{
                  padding: "8px 16px",
                  backgroundColor: "#6c757d",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                }}
              >
                {showDebugInfo ? "Hide Debug" : "🔍 Debug"}
              </button>
            )}
          </div>
        </div>

        {/* Quick stats */}
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: "15px",
            marginBottom: "30px",
          }}
        >
          <div
            style={{
              padding: "20px",
              backgroundColor: "#e3f2fd",
              borderRadius: "8px",
              textAlign: "center",
            }}
          >
            <div style={{ fontSize: "32px", marginBottom: "10px" }}>📁</div>
            <div style={{ fontSize: "24px", fontWeight: "bold" }}>
              {dashboardStats.totalFolders}
            </div>
            <div style={{ color: "#666" }}>My Folders</div>
          </div>

          <div
            style={{
              padding: "20px",
              backgroundColor: "#e8f5e8",
              borderRadius: "8px",
              textAlign: "center",
            }}
          >
            <div style={{ fontSize: "32px", marginBottom: "10px" }}>🤝</div>
            <div style={{ fontSize: "24px", fontWeight: "bold" }}>
              {dashboardStats.sharedFolders}
            </div>
            <div style={{ color: "#666" }}>Shared Folders</div>
          </div>

          <div
            style={{
              padding: "20px",
              backgroundColor: "#fff3e0",
              borderRadius: "8px",
              textAlign: "center",
            }}
          >
            <div style={{ fontSize: "32px", marginBottom: "10px" }}>🔐</div>
            <div style={{ fontSize: "24px", fontWeight: "bold" }}>E2EE</div>
            <div style={{ color: "#666" }}>End-to-End Encrypted</div>
          </div>

          <div
            style={{
              padding: "20px",
              backgroundColor: "#f3e5f5",
              borderRadius: "8px",
              textAlign: "center",
            }}
          >
            <div style={{ fontSize: "32px", marginBottom: "10px" }}>⚡</div>
            <div style={{ fontSize: "24px", fontWeight: "bold" }}>
              {dashboardStats.recentActivity.length}
            </div>
            <div style={{ color: "#666" }}>Recent Items</div>
          </div>
        </div>

        {/* Quick Sync Panel */}
        {showQuickSync && (
          <div style={{ marginBottom: "30px" }}>
            <FileSyncManager
              onSyncComplete={handleSyncComplete}
              showDetailed={false}
            />
          </div>
        )}
      </div>

      {/* Main action cards */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(300px, 1fr))",
          gap: "20px",
          marginBottom: "40px",
        }}
      >
        {/* Primary File Manager Card */}
        <div
          style={{
            border: "2px solid #007bff",
            borderRadius: "8px",
            padding: "25px",
            cursor: "pointer",
            transition: "all 0.3s",
            backgroundColor: "white",
          }}
          onClick={() => navigate("/files")}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow = "0 8px 16px rgba(0,123,255,0.2)";
            e.currentTarget.style.transform = "translateY(-2px)";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "none";
            e.currentTarget.style.transform = "translateY(0)";
          }}
        >
          <div style={{ textAlign: "center" }}>
            <div style={{ fontSize: "48px", marginBottom: "15px" }}>📁</div>
            <h3 style={{ marginTop: 0, color: "#007bff" }}>File Manager</h3>
            <p style={{ color: "#666", marginBottom: "20px" }}>
              Browse, upload, and manage your encrypted files and folders in a
              unified interface
            </p>
            <div
              style={{
                backgroundColor: "#e3f2fd",
                padding: "10px",
                borderRadius: "4px",
                marginBottom: "15px",
              }}
            >
              <strong>New Features:</strong>
              <ul
                style={{
                  margin: "5px 0 0 0",
                  paddingLeft: "20px",
                  textAlign: "left",
                  fontSize: "14px",
                }}
              >
                <li>Drag & drop file uploads</li>
                <li>Inline folder creation</li>
                <li>File state management (active/archived/deleted)</li>
                <li>Enhanced navigation & breadcrumbs</li>
              </ul>
            </div>
            <button
              style={{
                padding: "12px 24px",
                backgroundColor: "#007bff",
                color: "white",
                border: "none",
                borderRadius: "4px",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              Open File Manager →
            </button>
          </div>
        </div>

        {/* Collections Card (Legacy) */}
        <div
          style={{
            border: "1px solid #ddd",
            borderRadius: "8px",
            padding: "25px",
            cursor: "pointer",
            transition: "all 0.3s",
            backgroundColor: "white",
          }}
          onClick={() => navigate("/collections")}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow = "0 4px 8px rgba(0,0,0,0.1)";
            e.currentTarget.style.transform = "translateY(-1px)";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "none";
            e.currentTarget.style.transform = "translateY(0)";
          }}
        >
          <div style={{ textAlign: "center" }}>
            <div style={{ fontSize: "48px", marginBottom: "15px" }}>📚</div>
            <h3 style={{ marginTop: 0, color: "#6c757d" }}>
              Collections (Legacy)
            </h3>
            <p style={{ color: "#666", marginBottom: "20px" }}>
              Advanced collection management with detailed controls and sharing
              options
            </p>
            <div
              style={{
                backgroundColor: "#f8f9fa",
                padding: "10px",
                borderRadius: "4px",
                marginBottom: "15px",
              }}
            >
              <strong>Advanced Features:</strong>
              <ul
                style={{
                  margin: "5px 0 0 0",
                  paddingLeft: "20px",
                  textAlign: "left",
                  fontSize: "14px",
                }}
              >
                <li>Detailed collection sharing</li>
                <li>Permission management</li>
                <li>Collection hierarchy tools</li>
                <li>Advanced metadata viewing</li>
              </ul>
            </div>
            <button
              style={{
                padding: "12px 24px",
                backgroundColor: "#6c757d",
                color: "white",
                border: "none",
                borderRadius: "4px",
                fontSize: "16px",
              }}
            >
              View Collections →
            </button>
          </div>
        </div>

        {/* Profile Card */}
        <div
          style={{
            border: "1px solid #ddd",
            borderRadius: "8px",
            padding: "25px",
            cursor: "pointer",
            transition: "all 0.3s",
            backgroundColor: "white",
          }}
          onClick={() => navigate("/profile")}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow = "0 4px 8px rgba(0,0,0,0.1)";
            e.currentTarget.style.transform = "translateY(-1px)";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "none";
            e.currentTarget.style.transform = "translateY(0)";
          }}
        >
          <div style={{ textAlign: "center" }}>
            <div style={{ fontSize: "48px", marginBottom: "15px" }}>👤</div>
            <h3 style={{ marginTop: 0, color: "#28a745" }}>My Profile</h3>
            <p style={{ color: "#666", marginBottom: "20px" }}>
              View and manage your account settings, security, and encryption
              keys
            </p>
            <button
              style={{
                padding: "12px 24px",
                backgroundColor: "#28a745",
                color: "white",
                border: "none",
                borderRadius: "4px",
                fontSize: "16px",
              }}
            >
              View Profile →
            </button>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      {dashboardStats.recentActivity.length > 0 && (
        <div
          style={{
            backgroundColor: "#f8f9fa",
            padding: "20px",
            borderRadius: "8px",
            marginBottom: "30px",
          }}
        >
          <h3 style={{ marginTop: 0 }}>📋 Recent Activity</h3>
          <div style={{ display: "grid", gap: "10px" }}>
            {dashboardStats.recentActivity.map((item, index) => (
              <div
                key={index}
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  padding: "10px",
                  backgroundColor: "white",
                  borderRadius: "4px",
                  cursor: "pointer",
                }}
                onClick={() => navigate(`/files/${item.id}`)}
              >
                <div
                  style={{ display: "flex", alignItems: "center", gap: "10px" }}
                >
                  <span style={{ fontSize: "20px" }}>
                    {item.type === "shared_folder" ? "🤝" : "📁"}
                  </span>
                  <div>
                    <div style={{ fontWeight: "bold" }}>{item.name}</div>
                    <div style={{ fontSize: "12px", color: "#666" }}>
                      {item.type === "shared_folder"
                        ? "Shared Folder"
                        : "My Folder"}
                    </div>
                  </div>
                </div>
                <div style={{ fontSize: "12px", color: "#999" }}>
                  {formatDate(item.date)}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Quick Actions */}
      <div
        style={{
          backgroundColor: "#e9ecef",
          padding: "20px",
          borderRadius: "8px",
          marginBottom: "30px",
        }}
      >
        <h3 style={{ marginTop: 0 }}>⚡ Quick Actions</h3>
        <div
          style={{
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
          }}
        >
          <button
            onClick={() => navigate("/files")}
            style={{
              padding: "10px 20px",
              backgroundColor: "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            📁 Browse Files
          </button>

          <button
            onClick={() => navigate("/collections/create")}
            style={{
              padding: "10px 20px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            ➕ Create Folder
          </button>

          <button
            onClick={() => setShowQuickSync(true)}
            style={{
              padding: "10px 20px",
              backgroundColor: "#17a2b8",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            🔄 Sync Files
          </button>

          <button
            onClick={() => navigate("/profile")}
            style={{
              padding: "10px 20px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            ⚙️ Settings
          </button>
        </div>
      </div>

      {/* Debug Info */}
      {showDebugInfo && import.meta.env.DEV && (
        <div
          style={{
            backgroundColor: "#fff3cd",
            border: "1px solid #ffeaa7",
            padding: "15px",
            borderRadius: "4px",
            marginBottom: "20px",
          }}
        >
          <h4 style={{ margin: "0 0 10px 0" }}>🔍 Debug Information</h4>
          <details>
            <summary style={{ cursor: "pointer", marginBottom: "10px" }}>
              Authentication Debug Info
            </summary>
            <pre
              style={{
                backgroundColor: "#f8f9fa",
                padding: "10px",
                borderRadius: "4px",
                fontSize: "12px",
                overflow: "auto",
              }}
            >
              {JSON.stringify(getDebugInfo(), null, 2)}
            </pre>
          </details>

          <details>
            <summary style={{ cursor: "pointer", marginBottom: "10px" }}>
              Collections Debug Info
            </summary>
            <pre
              style={{
                backgroundColor: "#f8f9fa",
                padding: "10px",
                borderRadius: "4px",
                fontSize: "12px",
                overflow: "auto",
              }}
            >
              {JSON.stringify(
                {
                  totalCollections: collections.length,
                  totalSharedCollections: sharedCollections.length,
                  collectionsLoading,
                  firstCollection: collections[0]
                    ? {
                        id: collections[0].id,
                        name: collections[0].name || "[Encrypted]",
                        type: collections[0].collection_type,
                      }
                    : null,
                },
                null,
                2,
              )}
            </pre>
          </details>
        </div>
      )}

      {/* Security Information */}
      <div
        style={{
          backgroundColor: "#f8f9fa",
          borderLeft: "4px solid #28a745",
          padding: "20px",
          borderRadius: "4px",
          marginBottom: "30px",
        }}
      >
        <h4 style={{ marginTop: 0 }}>🔐 Security Information</h4>
        <ul style={{ marginBottom: 0 }}>
          <li>
            All your files are end-to-end encrypted using ChaCha20-Poly1305
          </li>
          <li>Your password never leaves your device</li>
          <li>We cannot access your encrypted data</li>
          <li>Files support versioning and soft deletion with recovery</li>
          <li>Remember to keep your recovery key safe</li>
        </ul>
      </div>

      {/* Logout */}
      <div
        style={{
          borderTop: "1px solid #ddd",
          paddingTop: "20px",
          textAlign: "center",
        }}
      >
        <button
          onClick={handleLogout}
          style={{
            padding: "12px 40px",
            backgroundColor: "#dc3545",
            color: "white",
            border: "none",
            borderRadius: "4px",
            fontSize: "16px",
            cursor: "pointer",
          }}
        >
          🚪 Logout
        </button>
      </div>
    </div>
  );
};

const ProtectedDashboard = withPasswordProtection(Dashboard);
export default ProtectedDashboard;
