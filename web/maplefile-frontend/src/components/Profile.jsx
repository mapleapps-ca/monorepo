// src/components/Profile.js
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../contexts/ServiceContext";

const Profile = () => {
  const { authService, meService } = useServices();
  const navigate = useNavigate();

  // State
  const [user, setUser] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isEditingEmail, setIsEditingEmail] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  // Check authentication and load user data on component mount
  useEffect(() => {
    if (!authService.isAuthenticated()) {
      // Redirect to login if not authenticated
      navigate("/login");
    } else {
      try {
        const currentUser = authService.getCurrentUser();
        if (currentUser) {
          // For the new E2EE system, we store different user data
          const profile = {
            email: currentUser.email,
            name: currentUser.email?.split("@")[0] || "User", // Use email prefix as default name
            createdAt: currentUser.loginTime || new Date().toISOString(),
          };
          setUser(profile);
          setName(profile.name);
          setEmail(profile.email);
        } else {
          throw new Error("No user data found");
        }
      } catch (err) {
        console.error("Error loading profile:", err);
        navigate("/login");
      }
    }
  }, [authService, navigate]);

  // Handle profile update
  const handleUpdate = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setLoading(true);

    try {
      // Validate input
      if (!name.trim()) {
        throw new Error("Name cannot be empty");
      }

      // Update profile
      const updatedUser = meService.updateName(name);
      setUser(updatedUser);
      setSuccess("Profile updated successfully!");
      setIsEditing(false);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle email update
  const handleUpdateEmail = async (e) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setLoading(true);

    try {
      // Update email
      const updatedUser = meService.updateEmail(email);
      setUser(updatedUser);
      setSuccess("Email updated successfully!");
      setIsEditingEmail(false);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle cancel editing
  const handleCancel = () => {
    setName(user.name);
    setEmail(user.email);
    setIsEditing(false);
    setIsEditingEmail(false);
    setError("");
  };

  // Show loading state while checking authentication
  if (!user) {
    return <div style={styles.loading}>Loading...</div>;
  }

  return (
    <div style={styles.container}>
      <div style={styles.profileCard}>
        <h2 style={styles.title}>My Profile</h2>

        {error && <div style={styles.error}>{error}</div>}

        {success && <div style={styles.success}>{success}</div>}

        <div style={styles.profileInfo}>
          <div style={styles.infoRow}>
            <span style={styles.label}>Email:</span>
            {isEditingEmail ? (
              <form onSubmit={handleUpdateEmail} style={styles.editForm}>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  style={styles.input}
                  disabled={loading}
                />
                <button
                  type="submit"
                  style={styles.saveButton}
                  disabled={loading}
                >
                  {loading ? "Saving..." : "Save"}
                </button>
                <button
                  type="button"
                  onClick={handleCancel}
                  style={styles.cancelButton}
                  disabled={loading}
                >
                  Cancel
                </button>
              </form>
            ) : (
              <>
                <span style={styles.value}>{user.email}</span>
                <button
                  onClick={() => setIsEditingEmail(true)}
                  style={styles.editButton}
                >
                  Edit
                </button>
              </>
            )}
          </div>

          <div style={styles.infoRow}>
            <span style={styles.label}>Name:</span>
            {isEditing ? (
              <form onSubmit={handleUpdate} style={styles.editForm}>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  style={styles.input}
                  disabled={loading}
                />
                <button
                  type="submit"
                  style={styles.saveButton}
                  disabled={loading}
                >
                  {loading ? "Saving..." : "Save"}
                </button>
                <button
                  type="button"
                  onClick={handleCancel}
                  style={styles.cancelButton}
                  disabled={loading}
                >
                  Cancel
                </button>
              </form>
            ) : (
              <>
                <span style={styles.value}>{user.name}</span>
                <button
                  onClick={() => setIsEditing(true)}
                  style={styles.editButton}
                >
                  Edit
                </button>
              </>
            )}
          </div>

          <div style={styles.infoRow}>
            <span style={styles.label}>Member Since:</span>
            <span style={styles.value}>
              {new Date(user.createdAt).toLocaleDateString()}
            </span>
          </div>
        </div>

        <div style={styles.stats}>
          <h3 style={styles.statsTitle}>Account Statistics</h3>
          {(() => {
            const stats = meService.getUserStats();
            return (
              <>
                <div style={styles.statItem}>
                  <span style={styles.statLabel}>User ID:</span>
                  <span style={styles.statValue}>#{stats.userId}</span>
                </div>
                <div style={styles.statItem}>
                  <span style={styles.statLabel}>Account Type:</span>
                  <span style={styles.statValue}>{stats.accountType}</span>
                </div>
                <div style={styles.statItem}>
                  <span style={styles.statLabel}>Status:</span>
                  <span style={styles.statValue}>{stats.status}</span>
                </div>
                <div style={styles.statItem}>
                  <span style={styles.statLabel}>Member Since:</span>
                  <span style={styles.statValue}>{stats.memberSince}</span>
                </div>
                <div style={styles.statItem}>
                  <span style={styles.statLabel}>Last Updated:</span>
                  <span style={styles.statValue}>{stats.lastUpdated}</span>
                </div>
              </>
            );
          })()}
        </div>
      </div>
    </div>
  );
};

const styles = {
  container: {
    display: "flex",
    justifyContent: "center",
    alignItems: "flex-start",
    minHeight: "80vh",
    padding: "1rem",
  },
  profileCard: {
    backgroundColor: "white",
    padding: "2rem",
    borderRadius: "8px",
    boxShadow: "0 2px 10px rgba(0, 0, 0, 0.1)",
    width: "100%",
    maxWidth: "600px",
  },
  title: {
    textAlign: "center",
    marginBottom: "1.5rem",
    color: "#333",
  },
  loading: {
    textAlign: "center",
    padding: "2rem",
    fontSize: "1.2rem",
    color: "#666",
  },
  error: {
    backgroundColor: "#f8d7da",
    color: "#721c24",
    padding: "0.75rem",
    borderRadius: "4px",
    marginBottom: "1rem",
  },
  success: {
    backgroundColor: "#d4edda",
    color: "#155724",
    padding: "0.75rem",
    borderRadius: "4px",
    marginBottom: "1rem",
  },
  profileInfo: {
    marginBottom: "2rem",
  },
  infoRow: {
    display: "flex",
    alignItems: "center",
    marginBottom: "1rem",
    padding: "0.75rem",
    backgroundColor: "#f8f9fa",
    borderRadius: "4px",
  },
  label: {
    fontWeight: "bold",
    marginRight: "1rem",
    minWidth: "120px",
    color: "#555",
  },
  value: {
    flex: 1,
    color: "#333",
  },
  editForm: {
    display: "flex",
    gap: "0.5rem",
    flex: 1,
  },
  input: {
    flex: 1,
    padding: "0.5rem",
    border: "1px solid #ddd",
    borderRadius: "4px",
    fontSize: "1rem",
  },
  editButton: {
    marginLeft: "auto",
    backgroundColor: "#007bff",
    color: "white",
    border: "none",
    padding: "0.25rem 1rem",
    borderRadius: "4px",
    cursor: "pointer",
    fontSize: "0.9rem",
  },
  saveButton: {
    backgroundColor: "#28a745",
    color: "white",
    border: "none",
    padding: "0.5rem 1rem",
    borderRadius: "4px",
    cursor: "pointer",
    fontSize: "0.9rem",
  },
  cancelButton: {
    backgroundColor: "#6c757d",
    color: "white",
    border: "none",
    padding: "0.5rem 1rem",
    borderRadius: "4px",
    cursor: "pointer",
    fontSize: "0.9rem",
  },
  stats: {
    borderTop: "1px solid #dee2e6",
    paddingTop: "1.5rem",
  },
  statsTitle: {
    marginBottom: "1rem",
    color: "#333",
  },
  statItem: {
    display: "flex",
    justifyContent: "space-between",
    marginBottom: "0.5rem",
    padding: "0.5rem",
    backgroundColor: "#f8f9fa",
    borderRadius: "4px",
  },
  statLabel: {
    color: "#666",
  },
  statValue: {
    fontWeight: "bold",
    color: "#333",
  },
};

export default Profile;
