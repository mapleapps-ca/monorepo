// File: src/pages/User/Dashboard/Dashboard.jsx
// A simple navigation page for the user dashboard.
import { useNavigate } from "react-router";
import useAuth from "../../../hooks/useAuth.js";

const Dashboard = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  const navItems = [
    {
      path: "/files",
      title: "File Manager",
      description: "Browse, upload, and manage your encrypted files.",
      icon: "📁",
    },
    {
      path: "/sync-collections-example",
      title: "Sync Collections Example",
      description: "Developers page to test sync collections.",
      icon: "📁",
    },
    {
      path: "/profile",
      title: "My Profile",
      description: "Manage account settings and security keys.",
      icon: "👤",
    },
  ];

  return (
    <div style={{ padding: "20px", maxWidth: "800px", margin: "0 auto" }}>
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
            Welcome back, <strong>{user?.email || "User"}</strong>!
          </p>
        </div>
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

      <div style={{ display: "grid", gap: "15px" }}>
        {navItems.map((item) => (
          <div
            key={item.path}
            onClick={() => navigate(item.path)}
            style={{
              padding: "20px",
              border: "1px solid #ddd",
              borderRadius: "8px",
              cursor: "pointer",
              transition: "all 0.2s",
              display: "flex",
              alignItems: "center",
              gap: "20px",
              backgroundColor: "white",
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.boxShadow = "0 4px 8px rgba(0,0,0,0.1)";
              e.currentTarget.style.transform = "translateY(-2px)";
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.boxShadow = "none";
              e.currentTarget.style.transform = "translateY(0)";
            }}
          >
            <span style={{ fontSize: "2.5rem" }}>{item.icon}</span>
            <div>
              <h3 style={{ margin: "0 0 5px 0", color: "#333" }}>
                {item.title}
              </h3>
              <p style={{ margin: 0, color: "#666" }}>{item.description}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Dashboard;
