// ================================================================
// File: src/pages/User/Collection/List.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useLocation, Link } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const CollectionList = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { collectionService, localStorageService, passwordStorageService } =
    useServices();
  const { isAuthenticated, isLoading: authLoading } = useAuth();

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

  // Cache management
  const CACHE_KEY = "mapleapps_decrypted_collections";
  const CACHE_EXPIRY_KEY = "mapleapps_collections_cache_expiry";
  const CACHE_DURATION = 30 * 60 * 1000; // 30 minutes

  // Load cached collections from localStorage
  const loadCachedCollections = () => {
    try {
      const cached = localStorage.getItem(CACHE_KEY);
      const expiry = localStorage.getItem(CACHE_EXPIRY_KEY);

      if (cached && expiry) {
        const expiryTime = new Date(expiry);
        if (new Date() < expiryTime) {
          const data = JSON.parse(cached);
          console.log("[CollectionList] Loading collections from cache");

          // The new service stores all collections together with ownership markers
          if (data.collections) {
            const ownedCollections = data.collections.filter(
              (c) => c._isOwned === true,
            );
            const sharedCollections = data.collections.filter(
              (c) => c._isOwned === false,
            );

            return {
              owned_collections: ownedCollections,
              shared_collections: sharedCollections,
              cached_at: data.cached_at,
            };
          }

          // Fallback for old cache format
          return data;
        } else {
          localStorage.removeItem(CACHE_KEY);
          localStorage.removeItem(CACHE_EXPIRY_KEY);
        }
      }
    } catch (error) {
      console.error(
        "[CollectionList] Failed to load cached collections:",
        error,
      );
    }
    return null;
  };

  // Save collections to localStorage cache
  const saveCachedCollections = (owned, shared) => {
    try {
      const cacheData = {
        owned_collections: owned,
        shared_collections: shared,
        cached_at: new Date().toISOString(),
      };

      const expiryTime = new Date(Date.now() + CACHE_DURATION);
      localStorage.setItem(CACHE_KEY, JSON.stringify(cacheData));
      localStorage.setItem(CACHE_EXPIRY_KEY, expiryTime.toISOString());

      console.log("[CollectionList] Collections cached until", expiryTime);
    } catch (error) {
      console.error("[CollectionList] Failed to cache collections:", error);
    }
  };

  // Clear cache
  const clearCache = () => {
    localStorage.removeItem(CACHE_KEY);
    localStorage.removeItem(CACHE_EXPIRY_KEY);
    console.log("[CollectionList] Collections cache cleared");
  };

  // Load collections
  const loadCollections = async (
    passwordParam = null,
    forceRefresh = false,
  ) => {
    try {
      setLoading(true);
      setError("");

      // Check for stored password if no parameter provided
      let passwordToUse = passwordParam;
      if (!passwordToUse) {
        passwordToUse = passwordStorageService.getPassword();
      }

      // Check cache first unless forcing refresh
      if (!forceRefresh && !passwordParam) {
        const cachedData = loadCachedCollections();
        if (cachedData) {
          // Ensure proper separation of owned vs shared
          const ownedCollections = cachedData.owned_collections || [];
          const sharedCollections = cachedData.shared_collections || [];

          setCollections(ownedCollections);
          setSharedCollections(sharedCollections);
          setLoading(false);

          console.log("[CollectionList] Loaded from cache:", {
            owned: ownedCollections.length,
            shared: sharedCollections.length,
          });
          return;
        }
      }

      console.log("[CollectionList] Loading collections from server...");

      const result = await collectionService.getFilteredCollections(
        true,
        true,
        passwordToUse,
        forceRefresh,
      );

      // Set collections from the properly deduplicated results
      setCollections(result.owned_collections || []);
      setSharedCollections(result.shared_collections || []);

      // The service now handles caching internally with deduplication
      console.log("[CollectionList] Collections loaded:", {
        owned: result.owned_collections?.length || 0,
        shared: result.shared_collections?.length || 0,
        fromCache: result.from_cache || false,
      });

      // Check if any collections failed to decrypt and we haven't tried password yet
      const allCollections = [
        ...(result.owned_collections || []),
        ...(result.shared_collections || []),
      ];

      const failedDecryptions = allCollections.filter((c) => c.decrypt_error);

      if (
        failedDecryptions.length > 0 &&
        !passwordParam &&
        !passwordStorageService.hasPassword()
      ) {
        console.log(
          "[CollectionList] Some collections failed to decrypt, need password",
        );
        setShowPasswordPrompt(true);
      }
    } catch (err) {
      console.error("[CollectionList] Failed to load collections:", err);

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

  // Initial load
  useEffect(() => {
    if (authLoading) return;

    if (isAuthenticated) {
      console.log(
        "[DEBUG] Password service has password:",
        passwordStorageService.hasPassword(),
      );

      // Check for stored password first
      const storedPassword = passwordStorageService.getPassword();

      if (storedPassword) {
        console.log(
          "[CollectionList] Using stored password to load collections",
        );
        loadCollections(storedPassword);
      } else {
        // Try to load from cache first
        const cachedData = loadCachedCollections();
        if (cachedData) {
          setCollections(cachedData.owned_collections || []);
          setSharedCollections(cachedData.shared_collections || []);
          setLoading(false);

          // Check if any collections need decryption
          const needsDecryption = [
            ...cachedData.owned_collections,
            ...cachedData.shared_collections,
          ].some((c) => c.decrypt_error);

          if (needsDecryption) {
            setShowPasswordPrompt(true);
          }
        } else if (localStorageService.hasUserEncryptedData()) {
          // Need password
          setShowPasswordPrompt(true);
          setLoading(false);
        } else {
          loadCollections();
        }
      }
    } else {
      navigate("/login");
    }
  }, [isAuthenticated, authLoading, navigate]);

  // Handle success message
  useEffect(() => {
    if (location.state?.message) {
      setSuccessMessage(location.state.message);
      const timer = setTimeout(() => setSuccessMessage(""), 5000);
      return () => clearTimeout(timer);
    }
  }, [location.state]);

  // Handle password submission
  const handlePasswordSubmit = async (e) => {
    e.preventDefault();

    if (!password) {
      setPasswordError("Password is required");
      return;
    }

    setPasswordError("");

    try {
      console.log("[CollectionList] Loading collections with password");
      await loadCollections(password, true);

      // Success - hide password prompt
      setShowPasswordPrompt(false);
      setPassword("");
    } catch (err) {
      console.error("[CollectionList] Failed to decrypt with password:", err);
      setPasswordError("Invalid password. Please try again.");
      setPassword("");
    }
  };

  // Force refresh collections
  const handleRefreshCollections = async () => {
    clearCache();
    await loadCollections(null, true);
  };

  // Get filtered collections
  const getFilteredCollections = () => {
    switch (filter) {
      case "owned":
        return collections;
      case "shared":
        return sharedCollections;
      case "all":
      default:
        // The collections are already deduplicated by the service
        return [...collections, ...sharedCollections];
    }
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
      clearCache();
      await loadCollections(null, true);
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
    loadCollections();
  };

  const filteredCollections = getFilteredCollections();

  // Show loading while auth is loading
  if (authLoading) {
    return (
      <div>
        <h1>My Collections</h1>
        <p>Checking authentication...</p>
      </div>
    );
  }

  // Show password prompt if needed
  if (showPasswordPrompt && !loading) {
    return (
      <div>
        <h1>Enter Password to Decrypt Collections</h1>

        <p>
          Your collections are encrypted. Please enter your password to decrypt
          them.
        </p>

        {passwordError && (
          <div style={{ color: "red", marginBottom: "10px" }}>
            ❌ {passwordError}
          </div>
        )}

        <form onSubmit={handlePasswordSubmit}>
          <div>
            <label htmlFor="password">Password:</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              autoFocus
              style={{ marginLeft: "10px", width: "200px" }}
            />
          </div>

          <div style={{ marginTop: "15px" }}>
            <button type="submit" disabled={!password}>
              Decrypt Collections
            </button>
            <button
              type="button"
              onClick={handleSkipPassword}
              style={{ marginLeft: "10px" }}
            >
              Skip (View Encrypted)
            </button>
          </div>
        </form>

        <div
          style={{
            marginTop: "20px",
            padding: "10px",
            backgroundColor: "#f8f9fa",
          }}
        >
          <p>
            🔐 Your password is never stored and is only used to decrypt your
            data.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "20px",
        }}
      >
        <h1>My Collections</h1>
        <button onClick={() => navigate("/dashboard")}>
          ← Back to Dashboard
        </button>
      </div>

      {successMessage && (
        <div style={{ color: "green", marginBottom: "10px" }}>
          ✅ {successMessage}
        </div>
      )}

      {error && (
        <div style={{ color: "red", marginBottom: "10px" }}>❌ {error}</div>
      )}

      <div
        style={{
          marginBottom: "20px",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <div>
          <button
            onClick={() => setFilter("all")}
            style={{
              marginRight: "10px",
              fontWeight: filter === "all" ? "bold" : "normal",
            }}
          >
            All ({collections.length + sharedCollections.length})
          </button>
          <button
            onClick={() => setFilter("owned")}
            style={{
              marginRight: "10px",
              fontWeight: filter === "owned" ? "bold" : "normal",
            }}
          >
            Owned ({collections.length})
          </button>
          <button
            onClick={() => setFilter("shared")}
            style={{ fontWeight: filter === "shared" ? "bold" : "normal" }}
          >
            Shared ({sharedCollections.length})
          </button>
        </div>

        <div>
          <button
            onClick={handleRefreshCollections}
            title="Refresh collections"
            style={{ marginRight: "10px" }}
          >
            🔄 Refresh
          </button>
          <button onClick={() => navigate("/collections/create")}>
            + Create Collection
          </button>
        </div>
      </div>

      <div>
        {loading ? (
          <div>
            <p>Loading collections...</p>
            <p>Decrypting collection names...</p>
          </div>
        ) : filteredCollections.length === 0 ? (
          <div>
            <p>No collections found.</p>
            <button onClick={() => navigate("/collections/create")}>
              Create Your First Collection
            </button>
          </div>
        ) : (
          <div>
            {filteredCollections.map((collection) => {
              const isOwned = collections.some((c) => c.id === collection.id);
              const hasDecryptError = collection.decrypt_error;

              return (
                <div
                  key={collection.id} // Ensure we're using the id as key
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    padding: "15px",
                    marginBottom: "10px",
                    backgroundColor: "#f8f9fa",
                    border: "1px solid #ddd",
                    borderRadius: "4px",
                    position: "relative",
                  }}
                >
                  {/* Main clickable area using React Router Link */}
                  {!hasDecryptError ? (
                    <Link
                      to={`/collections/${collection.id}`}
                      style={{
                        position: "absolute",
                        top: 0,
                        left: 0,
                        right: 0,
                        bottom: 0,
                        textDecoration: "none",
                        color: "inherit",
                        zIndex: 1,
                      }}
                      onMouseEnter={(e) => {
                        e.target.parentElement.style.backgroundColor =
                          "#e9ecef";
                      }}
                      onMouseLeave={(e) => {
                        e.target.parentElement.style.backgroundColor =
                          "#f8f9fa";
                      }}
                    />
                  ) : (
                    <div
                      style={{
                        position: "absolute",
                        top: 0,
                        left: 0,
                        right: 0,
                        bottom: 0,
                        cursor: "not-allowed",
                        zIndex: 1,
                      }}
                    />
                  )}

                  {/* Content area */}
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      position: "relative",
                      zIndex: 2,
                      pointerEvents: "none",
                    }}
                  >
                    <span style={{ fontSize: "24px", marginRight: "15px" }}>
                      {collection.collection_type === "album" ? "🖼️" : "📁"}
                    </span>
                    <div>
                      <h3 style={{ margin: "0 0 5px 0" }}>
                        {collection.name}
                        {hasDecryptError && " 🔒"}
                      </h3>
                      <div style={{ fontSize: "14px", color: "#666" }}>
                        {collection.collection_type} •{" "}
                        {isOwned ? "Owned" : "Shared"} • Modified:{" "}
                        {new Date(collection.modified_at).toLocaleDateString()}
                        {!hasDecryptError && (
                          <span
                            style={{ marginLeft: "10px", color: "#007bff" }}
                          >
                            Click to open →
                          </span>
                        )}
                      </div>
                      {hasDecryptError && (
                        <div
                          style={{
                            fontSize: "12px",
                            color: "red",
                            marginTop: "4px",
                          }}
                        >
                          Unable to decrypt: {collection.decrypt_error}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Action buttons */}
                  <div style={{ position: "relative", zIndex: 3 }}>
                    {isOwned && !hasDecryptError && (
                      <>
                        <button
                          type="button"
                          onClick={(e) => {
                            console.log(
                              "[CollectionList] Delete button clicked",
                            );
                            e.preventDefault();
                            e.stopPropagation();
                            handleDeleteCollection(
                              collection.id,
                              collection.name,
                            );
                          }}
                          style={{
                            color: "red",
                            background: "none",
                            border: "none",
                            cursor: "pointer",
                            padding: "5px",
                            fontSize: "16px",
                          }}
                          title="Delete collection"
                        >
                          🗑️
                        </button>
                        <button
                          type="button"
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            navigate(`/collections/${collection.id}/files`);
                          }}
                          style={{
                            color: "#17a2b8",
                            background: "none",
                            border: "none",
                            cursor: "pointer",
                            padding: "5px",
                            fontSize: "16px",
                            marginRight: "10px",
                          }}
                          title="View files"
                        >
                          📄
                        </button>
                      </>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div
        style={{
          marginTop: "40px",
          padding: "20px",
          backgroundColor: "#f8f9fa",
          border: "1px solid #ddd",
        }}
      >
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
          <p style={{ color: "#856404" }}>
            Some collections could not be decrypted.
            <button
              onClick={() => setShowPasswordPrompt(true)}
              style={{
                background: "none",
                border: "none",
                color: "#007bff",
                textDecoration: "underline",
                cursor: "pointer",
              }}
            >
              Enter password to decrypt
            </button>
          </p>
        )}
        {loadCachedCollections() && (
          <p style={{ color: "#17a2b8" }}>
            📦 Collections loaded from cache.
            <button
              onClick={handleRefreshCollections}
              style={{
                background: "none",
                border: "none",
                color: "#007bff",
                textDecoration: "underline",
                cursor: "pointer",
              }}
            >
              Refresh from server
            </button>
          </p>
        )}
      </div>

      {/* Debug info in development */}
      {import.meta.env.DEV && (
        <details
          style={{
            marginTop: "20px",
            padding: "10px",
            backgroundColor: "#f8f9fa",
          }}
        >
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
                  hasStoredPassword: passwordStorageService.hasPassword(),
                  hasSessionKeys:
                    localStorageService?.hasSessionKeys?.() || false,
                  hasUserEncryptedData:
                    localStorageService?.hasUserEncryptedData?.() || false,
                  hasCachedData: !!loadCachedCollections(),
                  authLoading,
                  isAuthenticated,
                },
                null,
                2,
              )}
            </pre>
            <button onClick={clearCache} style={{ marginTop: "10px" }}>
              Clear Cache
            </button>
          </div>
        </details>
      )}
    </div>
  );
};

const ProtectedCollectionList = withPasswordProtection(CollectionList);

export default ProtectedCollectionList;
