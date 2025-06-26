// Updated pages/User/Collection/List.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";

const CollectionList = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { collectionService, localStorageService } = useServices();
  const { isAuthenticated } = useAuth();

  // State
  const [collections, setCollections] = useState([]);
  const [sharedCollections, setSharedCollections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [filter, setFilter] = useState("all"); // all, owned, shared

  // Password prompt state
  const [showPasswordPrompt, setShowPasswordPrompt] = useState(false);
  const [password, setPassword] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [decryptionAttempted, setDecryptionAttempted] = useState(false);

  // Check if we need password on mount
  useEffect(() => {
    if (isAuthenticated) {
      // Check if we have session keys or encrypted user data
      if (
        !localStorageService.hasSessionKeys() &&
        localStorageService.hasUserEncryptedData()
      ) {
        // We have encrypted data but no session keys - need password
        setShowPasswordPrompt(true);
        setLoading(false);
      } else {
        // Try to load collections
        loadCollections();
      }
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
  const loadCollections = async (passwordParam = null) => {
    try {
      setLoading(true);
      setError("");

      console.log("[CollectionList] Loading collections...");

      const result = await collectionService.getFilteredCollections(
        true,
        true,
        passwordParam,
      );

      setCollections(result.owned_collections || []);
      setSharedCollections(result.shared_collections || []);

      console.log("[CollectionList] Collections loaded:", {
        owned: result.owned_collections?.length || 0,
        shared: result.shared_collections?.length || 0,
      });

      // Check if any collections failed to decrypt
      const failedDecryptions = [
        ...result.owned_collections,
        ...result.shared_collections,
      ].filter((c) => c.decrypt_error);

      if (
        failedDecryptions.length > 0 &&
        !decryptionAttempted &&
        !passwordParam
      ) {
        console.log(
          "[CollectionList] Some collections failed to decrypt, may need password",
        );
        setShowPasswordPrompt(true);
      }
    } catch (err) {
      console.error("[CollectionList] Failed to load collections:", err);

      // Handle specific error cases
      if (
        err.message.includes("Password required") ||
        err.message.includes("session keys not available")
      ) {
        setShowPasswordPrompt(true);
        setError("");
      } else if (err.message.includes("User encryption keys not available")) {
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

  // Handle password submission
  const handlePasswordSubmit = async () => {
    if (!password) {
      setPasswordError("Password is required");
      return;
    }

    setPasswordError("");
    setDecryptionAttempted(true);

    try {
      console.log("[CollectionList] Loading collections with password");
      await loadCollections(password);

      // Success - hide password prompt
      setShowPasswordPrompt(false);
      setPassword("");
    } catch (err) {
      console.error("[CollectionList] Failed to decrypt with password:", err);
      setPasswordError("Invalid password. Please try again.");
      setPassword("");
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

  // Handle skip password
  const handleSkipPassword = () => {
    setShowPasswordPrompt(false);
    setPasswordError("");
    setPassword("");
    // Try to load without password - will show encrypted collections
    loadCollections();
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
          onClick={() => !hasDecryptError && handleCollectionClick(collection)}
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
          {isOwned && !hasDecryptError && (
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

  // Show password prompt if needed
  if (showPasswordPrompt && !loading) {
    return (
      <div style={styles.container}>
        <h1>Enter Password to Decrypt Collections</h1>

        <div style={styles.passwordCard}>
          <p style={styles.info}>
            Your collections are encrypted. Please enter your password to
            decrypt them.
          </p>

          {passwordError && (
            <div style={styles.errorMessage}>❌ {passwordError}</div>
          )}

          <div style={styles.passwordForm}>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              onKeyPress={(e) => e.key === "Enter" && handlePasswordSubmit()}
              placeholder="Enter your password"
              autoFocus
              style={styles.passwordInput}
            />

            <div style={styles.passwordButtons}>
              <button
                onClick={handlePasswordSubmit}
                disabled={!password}
                style={{ ...styles.submitButton, opacity: !password ? 0.6 : 1 }}
              >
                Decrypt Collections
              </button>

              <button onClick={handleSkipPassword} style={styles.skipButton}>
                Skip (View Encrypted)
              </button>
            </div>
          </div>

          <div style={styles.securityNote}>
            <p>
              🔐 Your password is never stored and is only used to decrypt your
              data.
            </p>
          </div>
        </div>
      </div>
    );
  }

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
        {filteredCollections.some((c) => c.decrypt_error) && (
          <p style={styles.warningText}>
            Some collections could not be decrypted.
            <button
              onClick={() => setShowPasswordPrompt(true)}
              style={styles.linkButton}
            >
              Enter password to decrypt
            </button>
          </p>
        )}
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
                  hasSessionKeys:
                    localStorageService?.hasSessionKeys?.() || false,
                  hasUserEncryptedData:
                    localStorageService?.hasUserEncryptedData?.() || false,
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
  passwordCard: {
    background: "white",
    border: "1px solid #e0e0e0",
    borderRadius: "8px",
    padding: "30px",
    maxWidth: "500px",
    margin: "40px auto",
    boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
  },
  passwordForm: {
    marginTop: "20px",
  },
  passwordInput: {
    width: "100%",
    padding: "12px",
    fontSize: "16px",
    border: "1px solid #ddd",
    borderRadius: "4px",
    marginBottom: "20px",
  },
  passwordButtons: {
    display: "flex",
    gap: "10px",
    justifyContent: "space-between",
  },
  submitButton: {
    flex: 1,
    padding: "12px 20px",
    background: "#007bff",
    color: "white",
    border: "none",
    borderRadius: "4px",
    fontSize: "16px",
    cursor: "pointer",
  },
  skipButton: {
    flex: 1,
    padding: "12px 20px",
    background: "#6c757d",
    color: "white",
    border: "none",
    borderRadius: "4px",
    fontSize: "16px",
    cursor: "pointer",
  },
  securityNote: {
    marginTop: "20px",
    padding: "15px",
    background: "#f8f9fa",
    borderRadius: "4px",
    fontSize: "14px",
    color: "#666",
  },
  warningText: {
    marginTop: "10px",
    color: "#856404",
  },
  linkButton: {
    background: "none",
    border: "none",
    color: "#007bff",
    textDecoration: "underline",
    cursor: "pointer",
    padding: "0",
    marginLeft: "5px",
  },
};

export default CollectionList;
