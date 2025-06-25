// Updated pages/User/Collection/List.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";

const CollectionList = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { collectionService } = useServices();
  const { isAuthenticated } = useAuth();

  // State
  const [collections, setCollections] = useState([]);
  const [sharedCollections, setSharedCollections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [filter, setFilter] = useState("all"); // all, owned, shared

  // Load collections on mount
  useEffect(() => {
    if (isAuthenticated) {
      loadCollections();
    } else {
      navigate("/login");
    }
  }, [isAuthenticated]);

  // Handle success message from navigation state
  useEffect(() => {
    if (location.state?.message) {
      setSuccessMessage(location.state.message);
      // Clear the message after 5 seconds
      const timer = setTimeout(() => {
        setSuccessMessage("");
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [location.state]);

  // Load collections based on filter
  const loadCollections = async () => {
    try {
      setLoading(true);
      setError("");

      console.log("[CollectionList] Loading collections...");

      const result = await collectionService.getFilteredCollections(true, true);

      setCollections(result.owned_collections || []);
      setSharedCollections(result.shared_collections || []);

      console.log("[CollectionList] Collections loaded:", {
        owned: result.owned_collections?.length || 0,
        shared: result.shared_collections?.length || 0,
      });
    } catch (err) {
      console.error("[CollectionList] Failed to load collections:", err);

      // Handle specific error cases
      if (err.message.includes("User encryption keys not available")) {
        setError(
          "Encryption keys not available. Please log out and log in again.",
        );
      } else {
        setError(err.message || "Failed to load collections");
      }
    } finally {
      setLoading(false);
    }
  };

  // Get filtered collections based on current filter
  const getFilteredCollections = () => {
    switch (filter) {
      case "owned":
        return collections;
      case "shared":
        return sharedCollections;
      case "all":
      default:
        return [...collections, ...sharedCollections];
    }
  };

  // Handle collection click
  const handleCollectionClick = (collection) => {
    // Navigate to collection detail page
    console.log("[CollectionList] Opening collection:", collection.id);
    // navigate(`/collections/${collection.id}`);
  };

  // Handle delete collection
  const handleDeleteCollection = async (collectionId, collectionName) => {
    if (
      !window.confirm(`Are you sure you want to delete "${collectionName}"?`)
    ) {
      return;
    }

    try {
      await collectionService.deleteCollection(collectionId);
      setSuccessMessage(`Collection "${collectionName}" deleted successfully`);
      // Reload collections
      await loadCollections();
    } catch (err) {
      console.error("[CollectionList] Failed to delete collection:", err);
      setError(err.message || "Failed to delete collection");
    }
  };

  // Render collection item
  const renderCollectionItem = (collection) => {
    const isOwned =
      collection.owner_id ===
      collections.find((c) => c.id === collection.id)?.owner_id;
    const hasDecryptError = collection.decrypt_error;

    return (
      <div key={collection.id} style={styles.collectionItem}>
        <div
          style={styles.collectionInfo}
          onClick={() => handleCollectionClick(collection)}
        >
          <div style={styles.collectionIcon}>
            {collection.collection_type === "album" ? "🖼️" : "📁"}
          </div>
          <div style={styles.collectionDetails}>
            <h3 style={styles.collectionName}>
              {collection.name}
              {hasDecryptError && " 🔒"}
            </h3>
            <div style={styles.collectionMeta}>
              <span>{collection.collection_type}</span>
              <span> • </span>
              <span>{isOwned ? "Owned" : "Shared"}</span>
              <span> • </span>
              <span>
                Modified:{" "}
                {new Date(collection.modified_at).toLocaleDateString()}
              </span>
            </div>
            {hasDecryptError && (
              <div style={styles.decryptError}>
                Unable to decrypt: {collection.decrypt_error}
              </div>
            )}
          </div>
        </div>
        <div style={styles.collectionActions}>
          {isOwned && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                handleDeleteCollection(collection.id, collection.name);
              }}
              style={styles.deleteButton}
            >
              🗑️
            </button>
          )}
        </div>
      </div>
    );
  };

  const filteredCollections = getFilteredCollections();

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h1>My Collections</h1>
        <button
          onClick={() => navigate("/dashboard")}
          style={styles.backButton}
        >
          ← Back to Dashboard
        </button>
      </div>

      {successMessage && (
        <div style={styles.successMessage}>✅ {successMessage}</div>
      )}

      {error && <div style={styles.errorMessage}>❌ {error}</div>}

      <div style={styles.toolbar}>
        <div style={styles.filterButtons}>
          <button
            onClick={() => setFilter("all")}
            style={{
              ...styles.filterButton,
              ...(filter === "all" ? styles.filterButtonActive : {}),
            }}
          >
            All ({collections.length + sharedCollections.length})
          </button>
          <button
            onClick={() => setFilter("owned")}
            style={{
              ...styles.filterButton,
              ...(filter === "owned" ? styles.filterButtonActive : {}),
            }}
          >
            Owned ({collections.length})
          </button>
          <button
            onClick={() => setFilter("shared")}
            style={{
              ...styles.filterButton,
              ...(filter === "shared" ? styles.filterButtonActive : {}),
            }}
          >
            Shared ({sharedCollections.length})
          </button>
        </div>

        <button
          onClick={() => navigate("/collections/create")}
          style={styles.createButton}
        >
          + Create Collection
        </button>
      </div>

      <div style={styles.content}>
        {loading ? (
          <div style={styles.loading}>
            <p>Loading collections...</p>
            <p>Decrypting collection names...</p>
          </div>
        ) : filteredCollections.length === 0 ? (
          <div style={styles.empty}>
            <p>No collections found.</p>
            <button
              onClick={() => navigate("/collections/create")}
              style={styles.createButton}
            >
              Create Your First Collection
            </button>
          </div>
        ) : (
          <div style={styles.collectionsList}>
            {filteredCollections.map(renderCollectionItem)}
          </div>
        )}
      </div>

      <div style={styles.info}>
        <h3>🔐 Encryption Status</h3>
        <p>
          All collection names are encrypted end-to-end. Only you can decrypt
          your collection names.
          {sharedCollections.length > 0 && (
            <span>
              {" "}
              Shared collections are decrypted using keys shared with you by
              their owners.
            </span>
          )}
        </p>
      </div>

      {/* Debug info in development */}
      {import.meta.env.DEV && (
        <details style={styles.debug}>
          <summary>🔍 Debug Information</summary>
          <div>
            <h4>Collections State:</h4>
            <pre>
              {JSON.stringify(
                {
                  ownedCount: collections.length,
                  sharedCount: sharedCollections.length,
                  filter,
                  loading,
                  error,
                },
                null,
                2,
              )}
            </pre>

            <h4>Sample Collection (if any):</h4>
            <pre>
              {JSON.stringify(collections[0] || sharedCollections[0], null, 2)}
            </pre>
          </div>
        </details>
      )}
    </div>
  );
};

const styles = {
  container: {
    padding: "20px",
    maxWidth: "1200px",
    margin: "0 auto",
  },
  header: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: "20px",
  },
  backButton: {
    padding: "8px 16px",
    background: "#f0f0f0",
    border: "1px solid #ddd",
    borderRadius: "4px",
    cursor: "pointer",
  },
  successMessage: {
    background: "#d4edda",
    border: "1px solid #c3e6cb",
    color: "#155724",
    padding: "12px",
    borderRadius: "4px",
    marginBottom: "20px",
  },
  errorMessage: {
    background: "#f8d7da",
    border: "1px solid #f5c6cb",
    color: "#721c24",
    padding: "12px",
    borderRadius: "4px",
    marginBottom: "20px",
  },
  toolbar: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: "20px",
  },
  filterButtons: {
    display: "flex",
    gap: "10px",
  },
  filterButton: {
    padding: "8px 16px",
    background: "#f0f0f0",
    border: "1px solid #ddd",
    borderRadius: "4px",
    cursor: "pointer",
  },
  filterButtonActive: {
    background: "#007bff",
    color: "white",
    border: "1px solid #007bff",
  },
  createButton: {
    padding: "8px 16px",
    background: "#28a745",
    color: "white",
    border: "none",
    borderRadius: "4px",
    cursor: "pointer",
  },
  content: {
    minHeight: "400px",
  },
  loading: {
    textAlign: "center",
    padding: "40px",
    color: "#666",
  },
  empty: {
    textAlign: "center",
    padding: "40px",
    color: "#666",
  },
  collectionsList: {
    display: "flex",
    flexDirection: "column",
    gap: "10px",
  },
  collectionItem: {
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    padding: "15px",
    background: "#f8f9fa",
    border: "1px solid #dee2e6",
    borderRadius: "4px",
    cursor: "pointer",
    transition: "background 0.2s",
    ":hover": {
      background: "#e9ecef",
    },
  },
  collectionInfo: {
    display: "flex",
    alignItems: "center",
    gap: "15px",
    flex: 1,
  },
  collectionIcon: {
    fontSize: "24px",
  },
  collectionDetails: {
    flex: 1,
  },
  collectionName: {
    margin: 0,
    fontSize: "16px",
    fontWeight: "500",
  },
  collectionMeta: {
    fontSize: "14px",
    color: "#666",
    marginTop: "4px",
  },
  decryptError: {
    fontSize: "12px",
    color: "#dc3545",
    marginTop: "4px",
  },
  collectionActions: {
    display: "flex",
    gap: "10px",
  },
  deleteButton: {
    padding: "4px 8px",
    background: "transparent",
    border: "1px solid #dc3545",
    borderRadius: "4px",
    cursor: "pointer",
    fontSize: "16px",
  },
  info: {
    marginTop: "40px",
    padding: "20px",
    background: "#f8f9fa",
    border: "1px solid #dee2e6",
    borderRadius: "4px",
  },
  debug: {
    marginTop: "20px",
    padding: "10px",
    background: "#f8f9fa",
    border: "1px solid #dee2e6",
    borderRadius: "4px",
  },
};

export default CollectionList;
