// File: src/pages/User/Dashboard/Dashboard.jsx
import React from "react";
import { useNavigate } from "react-router";
import useAuth from "../../../hooks/useAuth.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const Dashboard = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  return (
    <div style={{ padding: "20px", maxWidth: "800px", margin: "0 auto" }}>
      <h1>Dashboard</h1>

      <div style={{ marginBottom: "30px" }}>
        <p>
          Welcome back, <strong>{user?.email || "User"}</strong>!
        </p>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
          gap: "20px",
          marginBottom: "40px",
        }}
      >
        {/* File Manager Card */}
        <div
          style={{
            border: "1px solid #ddd",
            borderRadius: "8px",
            padding: "20px",
            cursor: "pointer",
            transition: "box-shadow 0.3s",
          }}
          onClick={() => navigate("/files")}
          onMouseEnter={(e) =>
            (e.currentTarget.style.boxShadow = "0 4px 8px rgba(0,0,0,0.1)")
          }
          onMouseLeave={(e) => (e.currentTarget.style.boxShadow = "none")}
        >
          <h3 style={{ marginTop: 0 }}>📁 File Manager</h3>
          <p style={{ color: "#666" }}>
            Browse, upload, and manage your encrypted files and folders
          </p>
          <button
            style={{
              marginTop: "10px",
              padding: "8px 16px",
              backgroundColor: "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
            }}
          >
            Open File Manager →
          </button>
        </div>

        {/* Collections Card (Legacy) */}
        <div
          style={{
            border: "1px solid #ddd",
            borderRadius: "8px",
            padding: "20px",
            cursor: "pointer",
            transition: "box-shadow 0.3s",
          }}
          onClick={() => navigate("/collections")}
          onMouseEnter={(e) =>
            (e.currentTarget.style.boxShadow = "0 4px 8px rgba(0,0,0,0.1)")
          }
          onMouseLeave={(e) => (e.currentTarget.style.boxShadow = "none")}
        >
          <h3 style={{ marginTop: 0 }}>📚 Collections (Legacy)</h3>
          <p style={{ color: "#666" }}>
            View and manage collections using the traditional interface
          </p>
          <button
            style={{
              marginTop: "10px",
              padding: "8px 16px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
            }}
          >
            View Collections →
          </button>
        </div>

        {/* Profile Card */}
        <div
          style={{
            border: "1px solid #ddd",
            borderRadius: "8px",
            padding: "20px",
            cursor: "pointer",
            transition: "box-shadow 0.3s",
          }}
          onClick={() => navigate("/profile")}
          onMouseEnter={(e) =>
            (e.currentTarget.style.boxShadow = "0 4px 8px rgba(0,0,0,0.1)")
          }
          onMouseLeave={(e) => (e.currentTarget.style.boxShadow = "none")}
        >
          <h3 style={{ marginTop: 0 }}>👤 My Profile</h3>
          <p style={{ color: "#666" }}>View and manage your account settings</p>
          <button
            style={{
              marginTop: "10px",
              padding: "8px 16px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "4px",
            }}
          >
            View Profile →
          </button>
        </div>
      </div>

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
            padding: "10px 30px",
            backgroundColor: "#dc3545",
            color: "white",
            border: "none",
            borderRadius: "4px",
            fontSize: "16px",
            cursor: "pointer",
          }}
        >
          Logout
        </button>
      </div>

      <div
        style={{
          marginTop: "40px",
          padding: "20px",
          backgroundColor: "#f8f9fa",
          borderRadius: "4px",
        }}
      >
        <h4 style={{ marginTop: 0 }}>🔐 Security Information</h4>
        <ul style={{ marginBottom: 0 }}>
          <li>All your files are end-to-end encrypted</li>
          <li>Your password never leaves your device</li>
          <li>We cannot access your encrypted data</li>
          <li>Remember to keep your recovery key safe</li>
        </ul>
      </div>
    </div>
  );
};

const ProtectedDashboard = withPasswordProtection(Dashboard);
export default ProtectedDashboard;
