// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/Collection/ShareCollectionManagerExample.jsx
// Example component demonstrating how to use the useCollectionSharing hook

import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useCollections } from "../../../../services/Services";
import { useAuth } from "../../../../services/Services";

const ShareCollectionManagerExample = () => {
  const navigate = useNavigate();
  const {
    // State
    isLoading,
    error,
    success,
    sharedCollections,
    collectionMembers,
    sharingHistory,
    managerStatus,

    // Operations
    shareCollection,
    removeMember,
    getCollectionMembers,
    shareCollectionReadOnly,
    shareCollectionReadWrite,
    shareCollectionAdmin,
    removeAllSharesForCollection,
    clearAllSharedCollections,

    // Utilities
    getSharedCollectionsByCollectionId,
    searchSharedCollections,
    getCollectionSharingHistory,
    getUserPassword,
    clearMessages,

    // Status
    isAuthenticated,
    canShareCollections,
    totalSharedCollections,
    PERMISSION_LEVELS,
    getRecentShares,
    getCollectionMembersById,
  } = useCollections();

  const { authManager, authService } = useAuth();

  // Form state
  const [collectionId, setCollectionId] = useState("");
  const [recipientId, setRecipientId] = useState("");
  const [recipientEmail, setRecipientEmail] = useState("");
  const [permissionLevel, setPermissionLevel] = useState(
    PERMISSION_LEVELS.READ_WRITE,
  );
  const [shareWithDescendants, setShareWithDescendants] = useState(true);
  const [password, setPassword] = useState("");
  const [searchTerm, setSearchTerm] = useState("");

  // Remove member form state
  const [removeCollectionId, setRemoveCollectionId] = useState("");
  const [removeRecipientId, setRemoveRecipientId] = useState("");
  const [removeFromDescendants, setRemoveFromDescendants] = useState(true);

  // UI state
  const [selectedShare, setSelectedShare] = useState(null);
  const [eventLog, setEventLog] = useState([]);
  const [showHistory, setShowHistory] = useState(false);
  const [showMembers, setShowMembers] = useState(false);
  const [activeCollectionMembers, setActiveCollectionMembers] = useState("");

  // Handle collection sharing
  const handleShareCollection = async () => {
    if (!collectionId.trim()) {
      alert("Collection ID is required");
      return;
    }

    if (!recipientId.trim()) {
      alert("Recipient ID is required");
      return;
    }

    if (!recipientEmail.trim()) {
      alert("Recipient email is required");
      return;
    }

    try {
      const shareData = {
        recipient_id: recipientId.trim(),
        recipient_email: recipientEmail.trim(),
        permission_level: permissionLevel,
        share_with_descendants: shareWithDescendants,
      };

      await shareCollection(collectionId.trim(), shareData, password || null);

      // Clear form on success
      setPassword("");

      // Log the event
      addToEventLog("collection_shared", {
        collectionId,
        recipientEmail,
        permissionLevel,
        shareWithDescendants,
      });
    } catch (err) {
      console.error("Collection sharing failed:", err);
      // Error is handled by the hook
    }
  };

  // Handle member removal
  const handleRemoveMember = async () => {
    if (!removeCollectionId.trim()) {
      alert("Collection ID is required");
      return;
    }

    if (!removeRecipientId.trim()) {
      alert("Recipient ID is required");
      return;
    }

    try {
      await removeMember(
        removeCollectionId.trim(),
        removeRecipientId.trim(),
        removeFromDescendants,
      );

      // Clear form on success
      setRemoveCollectionId("");
      setRemoveRecipientId("");

      // Log the event
      addToEventLog("member_removed", {
        collectionId: removeCollectionId,
        recipientId: removeRecipientId,
        removeFromDescendants,
      });
    } catch (err) {
      console.error("Member removal failed:", err);
      // Error is handled by the hook
    }
  };

  // Handle getting collection members
  const handleGetCollectionMembers = async (
    collectionId,
    forceRefresh = false,
  ) => {
    if (!collectionId.trim()) {
      alert("Collection ID is required");
      return;
    }

    try {
      await getCollectionMembers(collectionId.trim(), forceRefresh);
      setActiveCollectionMembers(collectionId.trim());
      setShowMembers(true);

      addToEventLog("collection_members_retrieved", {
        collectionId,
        forceRefresh,
      });
    } catch (err) {
      console.error("Failed to get collection members:", err);
    }
  };

  // Get password from storage
  const handleGetStoredPassword = async () => {
    try {
      const storedPassword = await getUserPassword();
      if (storedPassword) {
        setPassword(storedPassword);
        addToEventLog("password_loaded", { source: "storage" });
      } else {
        alert("No password found in storage");
      }
    } catch (err) {
      alert(`Failed to get stored password: ${err.message}`);
    }
  };

  // Handle quick share operations
  const handleQuickShareReadOnly = async () => {
    if (!collectionId.trim() || !recipientId.trim() || !recipientEmail.trim()) {
      alert("Collection ID, recipient ID, and email are required");
      return;
    }

    try {
      await shareCollectionReadOnly(
        collectionId.trim(),
        recipientId.trim(),
        recipientEmail.trim(),
        shareWithDescendants,
        password || null,
      );

      setPassword("");
      addToEventLog("quick_share_read_only", {
        collectionId,
        recipientEmail,
      });
    } catch (err) {
      console.error("Quick share read-only failed:", err);
    }
  };

  const handleQuickShareReadWrite = async () => {
    if (!collectionId.trim() || !recipientId.trim() || !recipientEmail.trim()) {
      alert("Collection ID, recipient ID, and email are required");
      return;
    }

    try {
      await shareCollectionReadWrite(
        collectionId.trim(),
        recipientId.trim(),
        recipientEmail.trim(),
        shareWithDescendants,
        password || null,
      );

      setPassword("");
      addToEventLog("quick_share_read_write", {
        collectionId,
        recipientEmail,
      });
    } catch (err) {
      console.error("Quick share read-write failed:", err);
    }
  };

  const handleQuickShareAdmin = async () => {
    if (!collectionId.trim() || !recipientId.trim() || !recipientEmail.trim()) {
      alert("Collection ID, recipient ID, and email are required");
      return;
    }

    try {
      await shareCollectionAdmin(
        collectionId.trim(),
        recipientId.trim(),
        recipientEmail.trim(),
        shareWithDescendants,
        password || null,
      );

      setPassword("");
      addToEventLog("quick_share_admin", {
        collectionId,
        recipientEmail,
      });
    } catch (err) {
      console.error("Quick share admin failed:", err);
    }
  };

  // Handle clear all shared collections
  const handleClearAllSharedCollections = async () => {
    if (
      !confirm(
        "Are you sure you want to clear ALL shared collections? This cannot be undone.",
      )
    )
      return;

    try {
      await clearAllSharedCollections();
      addToEventLog("all_shared_collections_cleared", {});
    } catch (err) {
      console.error("Failed to clear shared collections:", err);
    }
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

  // Search shared collections
  const filteredShares = searchTerm
    ? searchSharedCollections(searchTerm)
    : sharedCollections;

  // Get recent shares
  const recentShares = getRecentShares(24);

  // Auto-clear messages after 5 seconds
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        clearMessages();
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [success, error, clearMessages]);

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <div style={{ marginBottom: "20px" }}>
        <button onClick={() => navigate("/dashboard")}>
          ← Back to Dashboard
        </button>
      </div>

      <h2>🤝 Share Collection Manager Example (with Hooks)</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the <strong>useCollectionSharing</strong> hook
        with E2EE encryption for collection sharing. Collection keys are
        encrypted with recipients' public keys for secure sharing.
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
            <strong>User:</strong> {user?.email || "Not logged in"}
          </div>
          <div>
            <strong>Authenticated:</strong>{" "}
            {isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Can Share Collections:</strong>{" "}
            {canShareCollections ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Loading:</strong> {isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Total Shared:</strong> {totalSharedCollections}
          </div>
          <div>
            <strong>Recent Shares (24h):</strong> {recentShares.length}
          </div>
        </div>
      </div>

      {/* Share Collection Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>🤝 Share Collection:</h4>
        <div style={{ display: "grid", gap: "15px" }}>
          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Collection ID *
            </label>
            <input
              type="text"
              value={collectionId}
              onChange={(e) => setCollectionId(e.target.value)}
              placeholder="Enter collection UUID..."
              style={{
                width: "100%",
                padding: "8px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                fontFamily: "monospace",
                fontSize: "12px",
              }}
            />
          </div>

          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "15px",
            }}
          >
            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                Recipient User ID *
              </label>
              <input
                type="text"
                value={recipientId}
                onChange={(e) => setRecipientId(e.target.value)}
                placeholder="Enter recipient UUID..."
                style={{
                  width: "100%",
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                  fontFamily: "monospace",
                  fontSize: "12px",
                }}
              />
            </div>

            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                Recipient Email *
              </label>
              <input
                type="email"
                value={recipientEmail}
                onChange={(e) => setRecipientEmail(e.target.value)}
                placeholder="Enter recipient email..."
                style={{
                  width: "100%",
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                }}
              />
            </div>
          </div>

          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "15px",
            }}
          >
            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                Permission Level
              </label>
              <select
                value={permissionLevel}
                onChange={(e) => setPermissionLevel(e.target.value)}
                style={{
                  width: "100%",
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                }}
              >
                <option value={PERMISSION_LEVELS.READ_ONLY}>
                  👁️ Read Only
                </option>
                <option value={PERMISSION_LEVELS.READ_WRITE}>
                  ✏️ Read Write
                </option>
                <option value={PERMISSION_LEVELS.ADMIN}>👑 Admin</option>
              </select>
            </div>

            <div>
              <label
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "8px",
                  marginTop: "25px",
                }}
              >
                <input
                  type="checkbox"
                  checked={shareWithDescendants}
                  onChange={(e) => setShareWithDescendants(e.target.checked)}
                />
                <span style={{ fontWeight: "bold" }}>
                  Share with child collections
                </span>
              </label>
            </div>
          </div>

          <div>
            <label
              style={{
                display: "block",
                marginBottom: "5px",
                fontWeight: "bold",
              }}
            >
              Password (for encryption)
            </label>
            <div style={{ display: "flex", gap: "10px" }}>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter password or use stored password..."
                style={{
                  flex: 1,
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                }}
              />
              <button
                onClick={handleGetStoredPassword}
                style={{
                  padding: "8px 15px",
                  backgroundColor: "#6c757d",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                }}
              >
                Use Stored
              </button>
            </div>
            <small style={{ color: "#666" }}>
              Leave empty to use password from PasswordStorageService
            </small>
          </div>

          <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
            <button
              onClick={handleShareCollection}
              disabled={
                isLoading ||
                !collectionId.trim() ||
                !recipientId.trim() ||
                !recipientEmail.trim() ||
                !isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading ||
                  !collectionId.trim() ||
                  !recipientId.trim() ||
                  !recipientEmail.trim() ||
                  !isAuthenticated
                    ? "#6c757d"
                    : "#28a745",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading ||
                  !collectionId.trim() ||
                  !recipientId.trim() ||
                  !recipientEmail.trim() ||
                  !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {isLoading ? "🔄 Sharing..." : "🤝 Share Collection"}
            </button>

            <button
              onClick={handleQuickShareReadOnly}
              disabled={
                isLoading ||
                !collectionId.trim() ||
                !recipientId.trim() ||
                !recipientEmail.trim() ||
                !isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading ||
                  !collectionId.trim() ||
                  !recipientId.trim() ||
                  !recipientEmail.trim() ||
                  !isAuthenticated
                    ? "#6c757d"
                    : "#007bff",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading ||
                  !collectionId.trim() ||
                  !recipientId.trim() ||
                  !recipientEmail.trim() ||
                  !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              👁️ Quick Read-Only
            </button>

            <button
              onClick={handleQuickShareReadWrite}
              disabled={
                isLoading ||
                !collectionId.trim() ||
                !recipientId.trim() ||
                !recipientEmail.trim() ||
                !isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading ||
                  !collectionId.trim() ||
                  !recipientId.trim() ||
                  !recipientEmail.trim() ||
                  !isAuthenticated
                    ? "#6c757d"
                    : "#17a2b8",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading ||
                  !collectionId.trim() ||
                  !recipientId.trim() ||
                  !recipientEmail.trim() ||
                  !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              ✏️ Quick Read-Write
            </button>

            <button
              onClick={handleQuickShareAdmin}
              disabled={
                isLoading ||
                !collectionId.trim() ||
                !recipientId.trim() ||
                !recipientEmail.trim() ||
                !isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading ||
                  !collectionId.trim() ||
                  !recipientId.trim() ||
                  !recipientEmail.trim() ||
                  !isAuthenticated
                    ? "#6c757d"
                    : "#ffc107",
                color: "black",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading ||
                  !collectionId.trim() ||
                  !recipientId.trim() ||
                  !recipientEmail.trim() ||
                  !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "14px",
              }}
            >
              👑 Quick Admin
            </button>
          </div>
        </div>
      </div>

      {/* Remove Member Form */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#fff3cd",
          borderRadius: "6px",
          border: "1px solid #ffeaa7",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>🗑️ Remove Member:</h4>
        <div style={{ display: "grid", gap: "15px" }}>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "15px",
            }}
          >
            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                Collection ID *
              </label>
              <input
                type="text"
                value={removeCollectionId}
                onChange={(e) => setRemoveCollectionId(e.target.value)}
                placeholder="Enter collection UUID..."
                style={{
                  width: "100%",
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                  fontFamily: "monospace",
                  fontSize: "12px",
                }}
              />
            </div>

            <div>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                }}
              >
                Recipient User ID *
              </label>
              <input
                type="text"
                value={removeRecipientId}
                onChange={(e) => setRemoveRecipientId(e.target.value)}
                placeholder="Enter recipient UUID..."
                style={{
                  width: "100%",
                  padding: "8px",
                  border: "1px solid #ddd",
                  borderRadius: "4px",
                  fontFamily: "monospace",
                  fontSize: "12px",
                }}
              />
            </div>
          </div>

          <div>
            <label
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
              }}
            >
              <input
                type="checkbox"
                checked={removeFromDescendants}
                onChange={(e) => setRemoveFromDescendants(e.target.checked)}
              />
              <span style={{ fontWeight: "bold" }}>
                Remove from child collections too
              </span>
            </label>
          </div>

          <div>
            <button
              onClick={handleRemoveMember}
              disabled={
                isLoading ||
                !removeCollectionId.trim() ||
                !removeRecipientId.trim() ||
                !isAuthenticated
              }
              style={{
                padding: "12px 20px",
                backgroundColor:
                  isLoading ||
                  !removeCollectionId.trim() ||
                  !removeRecipientId.trim() ||
                  !isAuthenticated
                    ? "#6c757d"
                    : "#dc3545",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor:
                  isLoading ||
                  !removeCollectionId.trim() ||
                  !removeRecipientId.trim() ||
                  !isAuthenticated
                    ? "not-allowed"
                    : "pointer",
                fontSize: "16px",
                fontWeight: "bold",
              }}
            >
              {isLoading ? "🔄 Removing..." : "🗑️ Remove Member"}
            </button>
          </div>
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
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>✅ {success}</span>
          <button
            onClick={clearMessages}
            style={{
              background: "none",
              border: "none",
              color: "#155724",
              cursor: "pointer",
              fontSize: "16px",
            }}
          >
            ✕
          </button>
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
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>❌ {error}</span>
          <button
            onClick={clearMessages}
            style={{
              background: "none",
              border: "none",
              color: "#721c24",
              cursor: "pointer",
              fontSize: "16px",
            }}
          >
            ✕
          </button>
        </div>
      )}

      {/* Get Collection Members */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#d1ecf1",
          borderRadius: "6px",
          border: "1px solid #bee5eb",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>👥 Collection Members:</h4>
        <div
          style={{
            display: "flex",
            gap: "10px",
            alignItems: "center",
            marginBottom: "15px",
          }}
        >
          <input
            type="text"
            placeholder="Enter collection ID to get members..."
            onChange={(e) => setActiveCollectionMembers(e.target.value)}
            style={{
              flex: 1,
              padding: "8px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              fontFamily: "monospace",
              fontSize: "12px",
            }}
          />
          <button
            onClick={() =>
              handleGetCollectionMembers(activeCollectionMembers, false)
            }
            disabled={isLoading || !activeCollectionMembers.trim()}
            style={{
              padding: "8px 15px",
              backgroundColor:
                isLoading || !activeCollectionMembers.trim()
                  ? "#6c757d"
                  : "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                isLoading || !activeCollectionMembers.trim()
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            Get Members
          </button>
          <button
            onClick={() =>
              handleGetCollectionMembers(activeCollectionMembers, true)
            }
            disabled={isLoading || !activeCollectionMembers.trim()}
            style={{
              padding: "8px 15px",
              backgroundColor:
                isLoading || !activeCollectionMembers.trim()
                  ? "#6c757d"
                  : "#17a2b8",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                isLoading || !activeCollectionMembers.trim()
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            Refresh
          </button>
        </div>

        {showMembers && activeCollectionMembers && (
          <div>
            <h5>Members for {activeCollectionMembers}:</h5>
            <div style={{ maxHeight: "200px", overflow: "auto" }}>
              {getCollectionMembersById(activeCollectionMembers).map(
                (member, index) => (
                  <div
                    key={index}
                    style={{
                      padding: "10px",
                      border: "1px solid #dee2e6",
                      borderRadius: "4px",
                      marginBottom: "5px",
                      backgroundColor: "white",
                    }}
                  >
                    <div>
                      <strong>Email:</strong> {member.recipient_email}
                    </div>
                    <div>
                      <strong>Permission:</strong> {member.permission_level}
                    </div>
                    <div>
                      <strong>ID:</strong> {member.recipient_id}
                    </div>
                  </div>
                ),
              )}
              {getCollectionMembersById(activeCollectionMembers).length ===
                0 && (
                <p style={{ color: "#666" }}>
                  No members found for this collection.
                </p>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Shared Collections List */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#fff3cd",
          borderRadius: "6px",
          border: "1px solid #ffeaa7",
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
          <h4 style={{ margin: 0 }}>
            🤝 Shared Collections ({totalSharedCollections}):
          </h4>
          <div style={{ display: "flex", gap: "10px" }}>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search shared collections..."
              style={{
                padding: "5px 10px",
                border: "1px solid #ddd",
                borderRadius: "4px",
                fontSize: "14px",
              }}
            />
            <button
              onClick={() => setShowHistory(!showHistory)}
              style={{
                padding: "5px 15px",
                backgroundColor: "#007bff",
                color: "white",
                border: "none",
                borderRadius: "4px",
                cursor: "pointer",
              }}
            >
              {showHistory ? "Hide" : "Show"} History
            </button>
            {sharedCollections.length > 0 && (
              <button
                onClick={handleClearAllSharedCollections}
                style={{
                  padding: "5px 15px",
                  backgroundColor: "#dc3545",
                  color: "white",
                  border: "none",
                  borderRadius: "4px",
                  cursor: "pointer",
                }}
              >
                🗑️ Clear All
              </button>
            )}
          </div>
        </div>

        {filteredShares.length === 0 ? (
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
              {sharedCollections.length === 0
                ? "No collections shared yet."
                : "No collections match your search."}
            </p>
            <p style={{ color: "#6c757d" }}>
              {sharedCollections.length === 0
                ? "Share your first collection using the form above."
                : "Try a different search term."}
            </p>
          </div>
        ) : (
          <div style={{ display: "grid", gap: "10px" }}>
            {filteredShares.map((share) => (
              <div
                key={`${share.collection_id}-${share.recipient_id}`}
                style={{
                  padding: "15px",
                  border: "1px solid #dee2e6",
                  borderRadius: "6px",
                  backgroundColor: "white",
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: "bold", marginBottom: "5px" }}>
                    🤝 {share.recipient_email}
                    <span
                      style={{
                        fontSize: "12px",
                        color: "#666",
                        marginLeft: "10px",
                        padding: "2px 6px",
                        backgroundColor: "#e9ecef",
                        borderRadius: "3px",
                      }}
                    >
                      {share.permission_level}
                    </span>
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <strong>Collection:</strong> {share.collection_id}
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <strong>Recipient ID:</strong> {share.recipient_id}
                  </div>
                  <div style={{ fontSize: "12px", color: "#666" }}>
                    <strong>Shared:</strong>{" "}
                    {new Date(
                      share.shared_at || share.locally_stored_at,
                    ).toLocaleString()}
                  </div>
                  {share.share_with_descendants && (
                    <div style={{ fontSize: "12px", color: "#28a745" }}>
                      📂 Includes child collections
                    </div>
                  )}
                  {share.memberships_created > 1 && (
                    <div style={{ fontSize: "12px", color: "#17a2b8" }}>
                      👥 Created {share.memberships_created} memberships
                    </div>
                  )}
                </div>
                <div style={{ display: "flex", gap: "10px" }}>
                  <button
                    onClick={() => setSelectedShare(share)}
                    style={{
                      padding: "5px 10px",
                      backgroundColor: "#007bff",
                      color: "white",
                      border: "none",
                      borderRadius: "4px",
                      cursor: "pointer",
                      fontSize: "12px",
                    }}
                  >
                    🔍 Details
                  </button>
                  <button
                    onClick={() =>
                      handleGetCollectionMembers(share.collection_id)
                    }
                    style={{
                      padding: "5px 10px",
                      backgroundColor: "#17a2b8",
                      color: "white",
                      border: "none",
                      borderRadius: "4px",
                      cursor: "pointer",
                      fontSize: "12px",
                    }}
                  >
                    👥 Members
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Sharing History */}
      {showHistory && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#e2e3e5",
            borderRadius: "6px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>
            📜 Sharing History ({sharingHistory.length}):
          </h4>
          {sharingHistory.length === 0 ? (
            <p style={{ color: "#6c757d" }}>No sharing history available.</p>
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
              {sharingHistory
                .slice()
                .reverse()
                .map((entry, index) => (
                  <div
                    key={`${entry.timestamp}-${index}`}
                    style={{
                      padding: "10px",
                      borderBottom:
                        index < sharingHistory.length - 1
                          ? "1px solid #dee2e6"
                          : "none",
                      fontFamily: "monospace",
                      fontSize: "12px",
                    }}
                  >
                    <div style={{ marginBottom: "5px" }}>
                      <strong style={{ color: "#007bff" }}>
                        {new Date(entry.timestamp).toLocaleString()}
                      </strong>
                      {" - "}
                      <strong style={{ color: "#28a745" }}>
                        {entry.action}
                      </strong>
                      {" - "}
                      <span style={{ color: "#666" }}>
                        {entry.collection_id}
                      </span>
                    </div>
                    {entry.recipient_email && (
                      <div style={{ color: "#666", marginLeft: "20px" }}>
                        Recipient: {entry.recipient_email} (
                        {entry.permission_level})
                      </div>
                    )}
                  </div>
                ))}
            </div>
          )}
        </div>
      )}

      {/* Selected Share Details */}
      {selectedShare && (
        <div
          style={{
            marginBottom: "20px",
            padding: "15px",
            backgroundColor: "#d1ecf1",
            borderRadius: "6px",
            border: "1px solid #bee5eb",
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
            <h4 style={{ margin: 0 }}>🔍 Share Details:</h4>
            <button
              onClick={() => setSelectedShare(null)}
              style={{
                background: "none",
                border: "none",
                color: "#0c5460",
                cursor: "pointer",
                fontSize: "16px",
              }}
            >
              ✕
            </button>
          </div>
          <pre
            style={{
              backgroundColor: "#f8f9fa",
              padding: "10px",
              borderRadius: "4px",
              fontSize: "12px",
              overflow: "auto",
              maxHeight: "300px",
            }}
          >
            {JSON.stringify(selectedShare, null, 2)}
          </pre>
        </div>
      )}

      {/* Event Log */}
      <div>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <h3>📋 Collection Sharing Event Log ({eventLog.length})</h3>
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
              No collection sharing events logged yet.
            </p>
            <p style={{ color: "#6c757d" }}>
              Events will appear here when collections are shared, members are
              removed, or other sharing actions occur.
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

      {/* Quick Test Section */}
      <div
        style={{
          padding: "15px",
          backgroundColor: "#e9ecef",
          borderRadius: "8px",
          marginTop: "20px",
          border: "1px solid #dee2e6",
        }}
      >
        <h5 style={{ margin: "0 0 10px 0" }}>🚀 Quick Test</h5>
        <p style={{ margin: "0 0 10px 0", fontSize: "14px", color: "#666" }}>
          First create a collection in the Create Collection Manager, then use
          its ID here. You'll need valid recipient user IDs and emails for
          sharing.
          <strong>Note:</strong> The system encrypts collection keys with
          recipients' public keys.
        </p>
        <div
          style={{
            display: "flex",
            gap: "10px",
            alignItems: "center",
            flexWrap: "wrap",
          }}
        >
          <button
            onClick={() => {
              setCollectionId("7f558adb-57b6-11f0-8b98-c60a0c48537c");
              setRecipientId("8f558adb-57b6-11f0-8b98-c60a0c48537d");
              setRecipientEmail("recipient@example.com");
            }}
            style={{
              padding: "5px 10px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "3px",
              cursor: "pointer",
              fontSize: "12px",
              fontFamily: "monospace",
            }}
          >
            Use Sample IDs
          </button>
          <span style={{ fontSize: "12px", color: "#666" }}>
            Click to populate form with sample collection and recipient IDs for
            testing!
          </span>
        </div>
      </div>
    </div>
  );
};

export default ShareCollectionManagerExample;
