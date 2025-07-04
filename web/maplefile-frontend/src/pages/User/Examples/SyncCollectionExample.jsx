// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/SyncCollectionExample.jsx
// Updated to use SyncCollectionManager instead of SyncCollectionService
import React, { useState, useEffect } from "react";
import { useServices } from "../../../hooks/useService.jsx";
import useSyncCollectionManager from "../../../hooks/useSyncCollectionManager.js";

const SyncCollectionExample = () => {
  // Option 1: Use the new hook (recommended)
  const {
    syncCollections,
    loading,
    error,
    getSyncCollections,
    refreshSyncCollections,
    clearSyncCollections,
    statistics,
    managerStatus,
    debugInfo,
  } = useSyncCollectionManager();

  // Option 2: Direct service access (for comparison)
  const { syncCollectionManager } = useServices();
  const [directSyncCollections, setDirectSyncCollections] = useState([]);
  const [directLoading, setDirectLoading] = useState(false);
  const [directError, setDirectError] = useState(null);

  useEffect(() => {
    loadSyncCollectionsViaHook();
  }, []);

  // Using the hook (recommended approach)
  const loadSyncCollectionsViaHook = async () => {
    try {
      await getSyncCollections();
      console.log("Loaded sync collections via hook");
    } catch (error) {
      console.error("Failed to load sync collections via hook:", error);
    }
  };

  const refreshViaHook = async () => {
    try {
      await refreshSyncCollections();
      console.log("Refreshed sync collections via hook");
    } catch (error) {
      console.error("Failed to refresh sync collections via hook:", error);
    }
  };

  const clearViaHook = () => {
    try {
      clearSyncCollections();
      console.log("Cleared sync collections via hook");
    } catch (error) {
      console.error("Failed to clear sync collections via hook:", error);
    }
  };

  // Using direct service access (alternative approach)
  const loadSyncCollectionsDirectly = async () => {
    try {
      setDirectLoading(true);
      setDirectError(null);
      const collections = await syncCollectionManager.getSyncCollections();
      setDirectSyncCollections(collections);
      console.log("Loaded sync collections directly:", collections);
    } catch (error) {
      setDirectError(error.message || "Failed to load sync collections");
      console.error("Failed to load sync collections directly:", error);
    } finally {
      setDirectLoading(false);
    }
  };

  const refreshDirectly = async () => {
    try {
      setDirectLoading(true);
      setDirectError(null);
      const collections =
        await syncCollectionManager.forceRefreshSyncCollections();
      setDirectSyncCollections(collections);
      console.log("Refreshed sync collections directly:", collections);
    } catch (error) {
      setDirectError(error.message || "Failed to refresh sync collections");
      console.error("Failed to refresh sync collections directly:", error);
    } finally {
      setDirectLoading(false);
    }
  };

  const clearDirectly = () => {
    try {
      const success = syncCollectionManager.clearSyncCollections();
      if (success) {
        setDirectSyncCollections([]);
        console.log("Cleared sync collections directly");
      }
    } catch (error) {
      setDirectError(error.message || "Failed to clear sync collections");
      console.error("Failed to clear sync collections directly:", error);
    }
  };

  if (loading) {
    return (
      <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
        <h2>🔄 Sync Collection Manager Example</h2>
        <p>Loading sync collections...</p>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <h2>🔄 Sync Collection Manager Example</h2>
      <p style={{ color: "#666", marginBottom: "20px" }}>
        This page demonstrates the new <strong>SyncCollectionManager</strong>{" "}
        with both hook-based and direct service access patterns.
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
            <strong>Authenticated:</strong>{" "}
            {managerStatus?.isAuthenticated ? "✅ Yes" : "❌ No"}
          </div>
          <div>
            <strong>Manager Loading:</strong>{" "}
            {managerStatus?.isLoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>API Loading:</strong>{" "}
            {managerStatus?.isAPILoading ? "🔄 Yes" : "✅ No"}
          </div>
          <div>
            <strong>Storage Count:</strong>{" "}
            {managerStatus?.storage?.syncCollectionsCount || 0}
          </div>
          <div>
            <strong>Last Sync:</strong>{" "}
            {managerStatus?.lastSync
              ? new Date(managerStatus.lastSync).toLocaleString()
              : "Never"}
          </div>
        </div>
      </div>

      {/* Statistics */}
      <div
        style={{
          marginBottom: "20px",
          padding: "15px",
          backgroundColor: "#e8f5e8",
          borderRadius: "6px",
          border: "1px solid #c3e6cb",
        }}
      >
        <h4 style={{ margin: "0 0 10px 0" }}>📈 Statistics:</h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(150px, 1fr))",
            gap: "10px",
          }}
        >
          <div>
            <strong>Total:</strong> {statistics?.total || 0}
          </div>
          <div>
            <strong>Active:</strong> {statistics?.active || 0}
          </div>
          <div>
            <strong>Deleted:</strong> {statistics?.deleted || 0}
          </div>
          <div>
            <strong>Archived:</strong> {statistics?.archived || 0}
          </div>
        </div>
      </div>

      {/* Hook-based Approach */}
      <div style={{ marginBottom: "30px" }}>
        <h3>🎣 Hook-based Approach (Recommended)</h3>
        <div
          style={{
            marginBottom: "20px",
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
          }}
        >
          <button
            onClick={loadSyncCollectionsViaHook}
            disabled={loading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#28a745",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: loading ? "not-allowed" : "pointer",
            }}
          >
            {loading ? "🔄 Loading..." : "📂 Load via Hook"}
          </button>

          <button
            onClick={refreshViaHook}
            disabled={loading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#17a2b8",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: loading ? "not-allowed" : "pointer",
            }}
          >
            {loading ? "🔄 Refreshing..." : "🔄 Refresh via Hook"}
          </button>

          <button
            onClick={clearViaHook}
            disabled={loading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#dc3545",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: loading ? "not-allowed" : "pointer",
            }}
          >
            🗑️ Clear via Hook
          </button>
        </div>

        {error && (
          <div
            style={{
              marginBottom: "20px",
              padding: "15px",
              backgroundColor: "#f8d7da",
              borderRadius: "6px",
              color: "#721c24",
              border: "1px solid #f5c6cb",
            }}
          >
            <h4 style={{ margin: "0 0 10px 0" }}>❌ Hook Error:</h4>
            <p style={{ margin: 0 }}>{error}</p>
          </div>
        )}

        <div>
          <h4>📁 Sync Collections via Hook ({syncCollections.length})</h4>
          {syncCollections.length === 0 ? (
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
                No sync collections loaded via hook.
              </p>
            </div>
          ) : (
            <div
              style={{
                display: "grid",
                gap: "10px",
                maxHeight: "300px",
                overflow: "auto",
                border: "1px solid #dee2e6",
                borderRadius: "6px",
                padding: "10px",
              }}
            >
              {syncCollections.map((collection, index) => (
                <div
                  key={`hook-${collection.id}-${index}`}
                  style={{
                    padding: "12px",
                    border: "1px solid #dee2e6",
                    borderRadius: "6px",
                    backgroundColor:
                      collection.state === "active" ? "#d1ecf1" : "#fff3cd",
                    borderLeft: `4px solid ${collection.state === "active" ? "#0c5460" : "#856404"}`,
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                    }}
                  >
                    <div>
                      <p style={{ margin: "0 0 5px 0", fontWeight: "bold" }}>
                        ID: {collection.id}
                      </p>
                      <p
                        style={{ margin: "0", fontSize: "14px", color: "#666" }}
                      >
                        State: {collection.state} | Version:{" "}
                        {collection.version}
                      </p>
                    </div>
                    <div
                      style={{
                        fontSize: "12px",
                        color: "#999",
                        textAlign: "right",
                      }}
                    >
                      Modified:{" "}
                      {new Date(collection.modified_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Direct Service Approach */}
      <div style={{ marginBottom: "30px" }}>
        <h3>🔧 Direct Service Approach (Alternative)</h3>
        <div
          style={{
            marginBottom: "20px",
            display: "flex",
            gap: "10px",
            flexWrap: "wrap",
          }}
        >
          <button
            onClick={loadSyncCollectionsDirectly}
            disabled={directLoading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#6f42c1",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: directLoading ? "not-allowed" : "pointer",
            }}
          >
            {directLoading ? "🔄 Loading..." : "📂 Load Directly"}
          </button>

          <button
            onClick={refreshDirectly}
            disabled={directLoading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#fd7e14",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: directLoading ? "not-allowed" : "pointer",
            }}
          >
            {directLoading ? "🔄 Refreshing..." : "🔄 Refresh Directly"}
          </button>

          <button
            onClick={clearDirectly}
            disabled={directLoading}
            style={{
              padding: "10px 20px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: directLoading ? "not-allowed" : "pointer",
            }}
          >
            🗑️ Clear Directly
          </button>
        </div>

        {directError && (
          <div
            style={{
              marginBottom: "20px",
              padding: "15px",
              backgroundColor: "#f8d7da",
              borderRadius: "6px",
              color: "#721c24",
              border: "1px solid #f5c6cb",
            }}
          >
            <h4 style={{ margin: "0 0 10px 0" }}>❌ Direct Error:</h4>
            <p style={{ margin: 0 }}>{directError}</p>
          </div>
        )}

        <div>
          <h4>
            📁 Sync Collections via Direct Service (
            {directSyncCollections.length})
          </h4>
          {directSyncCollections.length === 0 ? (
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
                No sync collections loaded directly.
              </p>
            </div>
          ) : (
            <div
              style={{
                display: "grid",
                gap: "10px",
                maxHeight: "300px",
                overflow: "auto",
                border: "1px solid #dee2e6",
                borderRadius: "6px",
                padding: "10px",
              }}
            >
              {directSyncCollections.map((collection, index) => (
                <div
                  key={`direct-${collection.id}-${index}`}
                  style={{
                    padding: "12px",
                    border: "1px solid #dee2e6",
                    borderRadius: "6px",
                    backgroundColor:
                      collection.state === "active" ? "#d4edda" : "#fff3cd",
                    borderLeft: `4px solid ${collection.state === "active" ? "#155724" : "#856404"}`,
                  }}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                    }}
                  >
                    <div>
                      <p style={{ margin: "0 0 5px 0", fontWeight: "bold" }}>
                        ID: {collection.id}
                      </p>
                      <p
                        style={{ margin: "0", fontSize: "14px", color: "#666" }}
                      >
                        State: {collection.state} | Version:{" "}
                        {collection.version}
                      </p>
                    </div>
                    <div
                      style={{
                        fontSize: "12px",
                        color: "#999",
                        textAlign: "right",
                      }}
                    >
                      Modified:{" "}
                      {new Date(collection.modified_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Debug Info */}
      <details style={{ marginTop: "20px" }}>
        <summary style={{ cursor: "pointer", fontWeight: "bold" }}>
          🔧 Debug Info
        </summary>
        <pre
          style={{
            backgroundColor: "#f8f9fa",
            padding: "10px",
            borderRadius: "4px",
            fontSize: "12px",
            overflow: "auto",
            marginTop: "10px",
          }}
        >
          {JSON.stringify(debugInfo, null, 2)}
        </pre>
      </details>
    </div>
  );
};

export default SyncCollectionExample;
