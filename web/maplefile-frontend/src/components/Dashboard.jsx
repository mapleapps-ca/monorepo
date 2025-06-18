// src/components/Dashboard.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router";
import { useServices } from "../contexts/ServiceContext";

const Dashboard = () => {
  const { authService } = useServices();
  const navigate = useNavigate();

  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  // Load user data on component mount
  useEffect(() => {
    if (!authService.isAuthenticated()) {
      navigate("/login");
      return;
    }

    try {
      const currentUser = authService.getCurrentUser();
      setUser(currentUser);
    } catch (error) {
      console.error("Error loading user data:", error);
      authService.logout();
      navigate("/login");
    } finally {
      setLoading(false);
    }
  }, [authService, navigate]);

  // Quick actions
  const handleUploadFile = () => {
    alert("File upload functionality coming soon!");
  };

  const handleCreateFolder = () => {
    alert("Create folder functionality coming soon!");
  };

  const handleViewFiles = () => {
    alert("File browser coming soon!");
  };

  // Show loading state
  if (loading) {
    return <div style={styles.loading}>Loading dashboard...</div>;
  }

  // Show error state if no user
  if (!user) {
    return <div style={styles.error}>Failed to load user data</div>;
  }

  return (
    <div style={styles.container}>
      {/* Welcome Header */}
      <div style={styles.header}>
        <h1 style={styles.welcomeTitle}>
          Welcome back{user.email ? `, ${user.email.split("@")[0]}` : ""}!
        </h1>
        <p style={styles.welcomeSubtitle}>
          Your secure, encrypted file storage dashboard
        </p>
      </div>

      {/* Quick Stats */}
      <div style={styles.statsGrid}>
        <div style={styles.statCard}>
          <div style={styles.statIcon}>📁</div>
          <div style={styles.statContent}>
            <h3 style={styles.statNumber}>0</h3>
            <p style={styles.statLabel}>Files</p>
          </div>
        </div>

        <div style={styles.statCard}>
          <div style={styles.statIcon}>📊</div>
          <div style={styles.statContent}>
            <h3 style={styles.statNumber}>0 MB</h3>
            <p style={styles.statLabel}>Used</p>
          </div>
        </div>

        <div style={styles.statCard}>
          <div style={styles.statIcon}>🔒</div>
          <div style={styles.statContent}>
            <h3 style={styles.statNumber}>E2EE</h3>
            <p style={styles.statLabel}>Encrypted</p>
          </div>
        </div>

        <div style={styles.statCard}>
          <div style={styles.statIcon}>⚡</div>
          <div style={styles.statContent}>
            <h3 style={styles.statNumber}>Active</h3>
            <p style={styles.statLabel}>Status</p>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div style={styles.section}>
        <h2 style={styles.sectionTitle}>Quick Actions</h2>
        <div style={styles.actionGrid}>
          <button
            onClick={handleUploadFile}
            style={{ ...styles.actionCard, ...styles.primaryAction }}
          >
            <div style={styles.actionIcon}>⬆️</div>
            <h3 style={styles.actionTitle}>Upload Files</h3>
            <p style={styles.actionDescription}>
              Securely upload and encrypt your files
            </p>
          </button>

          <button onClick={handleCreateFolder} style={styles.actionCard}>
            <div style={styles.actionIcon}>📁</div>
            <h3 style={styles.actionTitle}>Create Folder</h3>
            <p style={styles.actionDescription}>
              Organize your files with folders
            </p>
          </button>

          <button onClick={handleViewFiles} style={styles.actionCard}>
            <div style={styles.actionIcon}>📋</div>
            <h3 style={styles.actionTitle}>Browse Files</h3>
            <p style={styles.actionDescription}>View and manage your files</p>
          </button>

          <Link to="/profile" style={styles.actionCard}>
            <div style={styles.actionIcon}>⚙️</div>
            <h3 style={styles.actionTitle}>Account Settings</h3>
            <p style={styles.actionDescription}>
              Manage your profile and security
            </p>
          </Link>
        </div>
      </div>

      {/* Security Info */}
      <div style={styles.section}>
        <h2 style={styles.sectionTitle}>Security Information</h2>
        <div style={styles.securityGrid}>
          <div style={styles.securityCard}>
            <h4 style={styles.securityTitle}>🔐 End-to-End Encryption</h4>
            <p style={styles.securityText}>
              All your files are encrypted on your device before being uploaded.
              Only you have the keys to decrypt them.
            </p>
          </div>

          <div style={styles.securityCard}>
            <h4 style={styles.securityTitle}>🛡️ Zero-Knowledge Architecture</h4>
            <p style={styles.securityText}>
              MapleFile cannot access your files or passwords. Your privacy is
              guaranteed by design.
            </p>
          </div>

          <div style={styles.securityCard}>
            <h4 style={styles.securityTitle}>🔑 Secure Key Management</h4>
            <p style={styles.securityText}>
              Your encryption keys are derived from your password and never
              stored on our servers.
            </p>
          </div>
        </div>
      </div>

      {/* Recent Activity Placeholder */}
      <div style={styles.section}>
        <h2 style={styles.sectionTitle}>Recent Activity</h2>
        <div style={styles.emptyState}>
          <div style={styles.emptyIcon}>📭</div>
          <h3 style={styles.emptyTitle}>No recent activity</h3>
          <p style={styles.emptyText}>Upload your first file to get started!</p>
        </div>
      </div>
    </div>
  );
};

const styles = {
  container: {
    maxWidth: "1200px",
    margin: "0 auto",
    padding: "2rem 1rem",
  },
  loading: {
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    height: "60vh",
    fontSize: "1.2rem",
    color: "#666",
  },
  error: {
    display: "flex",
    justifyContent: "center",
    alignItems: "center",
    height: "60vh",
    fontSize: "1.2rem",
    color: "#dc3545",
  },
  header: {
    textAlign: "center",
    marginBottom: "3rem",
  },
  welcomeTitle: {
    color: "#333",
    marginBottom: "0.5rem",
    fontSize: "2.5rem",
  },
  welcomeSubtitle: {
    color: "#666",
    fontSize: "1.1rem",
    margin: 0,
  },
  statsGrid: {
    display: "grid",
    gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
    gap: "1rem",
    marginBottom: "3rem",
  },
  statCard: {
    backgroundColor: "white",
    padding: "1.5rem",
    borderRadius: "8px",
    boxShadow: "0 2px 4px rgba(0, 0, 0, 0.1)",
    display: "flex",
    alignItems: "center",
    gap: "1rem",
  },
  statIcon: {
    fontSize: "2rem",
  },
  statContent: {
    flex: 1,
  },
  statNumber: {
    margin: 0,
    fontSize: "1.5rem",
    color: "#333",
  },
  statLabel: {
    margin: 0,
    color: "#666",
    fontSize: "0.9rem",
  },
  section: {
    marginBottom: "3rem",
  },
  sectionTitle: {
    color: "#333",
    marginBottom: "1rem",
    fontSize: "1.5rem",
  },
  actionGrid: {
    display: "grid",
    gridTemplateColumns: "repeat(auto-fit, minmax(280px, 1fr))",
    gap: "1rem",
  },
  actionCard: {
    backgroundColor: "white",
    padding: "1.5rem",
    borderRadius: "8px",
    boxShadow: "0 2px 4px rgba(0, 0, 0, 0.1)",
    border: "1px solid #e9ecef",
    textAlign: "center",
    cursor: "pointer",
    textDecoration: "none",
    color: "inherit",
    transition: "transform 0.2s, box-shadow 0.2s",
    ":hover": {
      transform: "translateY(-2px)",
      boxShadow: "0 4px 8px rgba(0, 0, 0, 0.15)",
    },
  },
  primaryAction: {
    backgroundColor: "#007bff",
    color: "white",
  },
  actionIcon: {
    fontSize: "2.5rem",
    marginBottom: "0.5rem",
  },
  actionTitle: {
    margin: "0 0 0.5rem 0",
    fontSize: "1.2rem",
  },
  actionDescription: {
    margin: 0,
    fontSize: "0.9rem",
    opacity: 0.8,
  },
  securityGrid: {
    display: "grid",
    gridTemplateColumns: "repeat(auto-fit, minmax(300px, 1fr))",
    gap: "1rem",
  },
  securityCard: {
    backgroundColor: "#f8f9fa",
    padding: "1.5rem",
    borderRadius: "8px",
    border: "1px solid #e9ecef",
  },
  securityTitle: {
    margin: "0 0 0.5rem 0",
    color: "#333",
    fontSize: "1.1rem",
  },
  securityText: {
    margin: 0,
    color: "#666",
    fontSize: "0.9rem",
    lineHeight: "1.4",
  },
  emptyState: {
    textAlign: "center",
    padding: "3rem 1rem",
    backgroundColor: "#f8f9fa",
    borderRadius: "8px",
    border: "1px solid #e9ecef",
  },
  emptyIcon: {
    fontSize: "3rem",
    marginBottom: "1rem",
  },
  emptyTitle: {
    color: "#333",
    marginBottom: "0.5rem",
  },
  emptyText: {
    color: "#666",
    margin: 0,
  },
};

export default Dashboard;
