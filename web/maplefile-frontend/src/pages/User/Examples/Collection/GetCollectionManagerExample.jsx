// File: monorepo/web/maplefile-frontend/src/pages/User/Examples/Collection/GetCollectionManagerExample.jsx
// Simplified example for getting and displaying collection data

import React, { useState } from "react";
import useAuth from "../../../../hooks/useAuth.js";
import { useCollections } from "../../../../hooks/useService.jsx";

const GetCollectionManagerExample = () => {
  const { user } = useAuth();
  const { getCollectionManager } = useCollections();

  // Simple state
  const [collectionId, setCollectionId] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);

  // Simple collection lookup
  const handleGetCollection = async () => {
    if (!collectionId.trim()) {
      alert("Please enter a Collection ID");
      return;
    }

    setIsLoading(true);
    setError(null);
    setResult(null);

    try {
      console.log("[GetCollectionExample] Getting collection:", collectionId);

      // Get collection using the manager
      const response = await getCollectionManager.getCollection(
        collectionId.trim(),
      );

      console.log("[GetCollectionExample] Collection retrieved:", response);
      setResult(response);
    } catch (err) {
      console.error("[GetCollectionExample] Error:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Clear results
  const handleClear = () => {
    setResult(null);
    setError(null);
    setCollectionId("");
  };

  return (
    <div style={{ padding: "20px", maxWidth: "1000px", margin: "0 auto" }}>
      <h2>🔍 Simple Collection Lookup</h2>
      <p style={{ color: "#666", marginBottom: "30px" }}>
        Enter a Collection ID to retrieve and display the collection data.
        <br />
        <strong>User:</strong> {user?.email || "Not logged in"}
      </p>

      {/* Input Form */}
      <div
        style={{
          padding: "20px",
          backgroundColor: "#f8f9fa",
          borderRadius: "8px",
          marginBottom: "20px",
          border: "1px solid #dee2e6",
        }}
      >
        <h4 style={{ margin: "0 0 15px 0" }}>Get Collection</h4>

        <div style={{ marginBottom: "15px" }}>
          <label
            style={{
              display: "block",
              marginBottom: "5px",
              fontWeight: "bold",
            }}
          >
            Collection ID:
          </label>
          <input
            type="text"
            value={collectionId}
            onChange={(e) => setCollectionId(e.target.value)}
            placeholder="Enter collection UUID (e.g., 7f558adb-57b6-11f0-8b98-c60a0c48537c)"
            style={{
              width: "100%",
              padding: "10px",
              border: "1px solid #ddd",
              borderRadius: "4px",
              fontFamily: "monospace",
              fontSize: "14px",
            }}
          />
        </div>

        <div style={{ display: "flex", gap: "10px" }}>
          <button
            onClick={handleGetCollection}
            disabled={isLoading || !collectionId.trim()}
            style={{
              padding: "12px 24px",
              backgroundColor:
                isLoading || !collectionId.trim() ? "#6c757d" : "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                isLoading || !collectionId.trim() ? "not-allowed" : "pointer",
              fontSize: "16px",
              fontWeight: "bold",
            }}
          >
            {isLoading ? "🔄 Loading..." : "🔍 Get Collection"}
          </button>

          <button
            onClick={handleClear}
            disabled={isLoading}
            style={{
              padding: "12px 24px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: isLoading ? "not-allowed" : "pointer",
              fontSize: "16px",
            }}
          >
            🗑️ Clear
          </button>
        </div>
      </div>

      {/* Error Display */}
      {error && (
        <div
          style={{
            padding: "20px",
            backgroundColor: "#f8d7da",
            borderRadius: "8px",
            marginBottom: "20px",
            border: "1px solid #f5c6cb",
            color: "#721c24",
          }}
        >
          <h4 style={{ margin: "0 0 10px 0", color: "#721c24" }}>❌ Error</h4>
          <p style={{ margin: 0, fontFamily: "monospace" }}>{error}</p>
        </div>
      )}

      {/* Success Display */}
      {result && (
        <div
          style={{
            padding: "20px",
            backgroundColor: "#d4edda",
            borderRadius: "8px",
            marginBottom: "20px",
            border: "1px solid #c3e6cb",
            color: "#155724",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0", color: "#155724" }}>
            ✅ Collection Retrieved Successfully
          </h4>

          {/* Collection Summary */}
          <div
            style={{
              backgroundColor: "white",
              padding: "15px",
              borderRadius: "6px",
              marginBottom: "15px",
              border: "1px solid #c3e6cb",
            }}
          >
            <h5 style={{ margin: "0 0 10px 0", color: "#333" }}>
              📄 Collection Summary
            </h5>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <tbody>
                <tr>
                  <td
                    style={{
                      padding: "5px",
                      fontWeight: "bold",
                      width: "150px",
                    }}
                  >
                    ID:
                  </td>
                  <td
                    style={{
                      padding: "5px",
                      fontFamily: "monospace",
                      fontSize: "12px",
                    }}
                  >
                    {result.collection?.id || "N/A"}
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: "5px", fontWeight: "bold" }}>Name:</td>
                  <td style={{ padding: "5px" }}>
                    <span
                      style={{
                        backgroundColor: result.collection?.name?.startsWith(
                          "[",
                        )
                          ? "#fff3cd"
                          : "#d1ecf1",
                        padding: "2px 6px",
                        borderRadius: "3px",
                        fontWeight: "bold",
                      }}
                    >
                      {result.collection?.name || "N/A"}
                    </span>
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: "5px", fontWeight: "bold" }}>Type:</td>
                  <td style={{ padding: "5px" }}>
                    {result.collection?.collection_type === "folder"
                      ? "📁"
                      : "📷"}{" "}
                    {result.collection?.collection_type || "N/A"}
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: "5px", fontWeight: "bold" }}>
                    Source:
                  </td>
                  <td style={{ padding: "5px" }}>
                    <span
                      style={{
                        backgroundColor:
                          result.source === "cache" ? "#fff3cd" : "#e2e3e5",
                        padding: "2px 6px",
                        borderRadius: "3px",
                        fontSize: "12px",
                      }}
                    >
                      {result.source || "unknown"}
                    </span>
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: "5px", fontWeight: "bold" }}>
                    Decrypted:
                  </td>
                  <td style={{ padding: "5px" }}>
                    {result.collection?._isDecrypted ? (
                      <span style={{ color: "#28a745", fontWeight: "bold" }}>
                        ✅ Yes
                      </span>
                    ) : (
                      <span style={{ color: "#dc3545", fontWeight: "bold" }}>
                        ❌ No
                      </span>
                    )}
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: "5px", fontWeight: "bold" }}>
                    Created:
                  </td>
                  <td style={{ padding: "5px", fontSize: "12px" }}>
                    {result.collection?.created_at
                      ? new Date(result.collection.created_at).toLocaleString()
                      : "N/A"}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          {/* Decryption Status */}
          {result.collection?._decryptionError && (
            <div
              style={{
                backgroundColor: "#f8d7da",
                padding: "10px",
                borderRadius: "4px",
                marginBottom: "15px",
                border: "1px solid #f5c6cb",
              }}
            >
              <strong style={{ color: "#721c24" }}>🔓 Decryption Error:</strong>
              <br />
              <span
                style={{
                  fontFamily: "monospace",
                  fontSize: "12px",
                  color: "#721c24",
                }}
              >
                {result.collection._decryptionError}
              </span>
            </div>
          )}

          {/* Members Info */}
          {result.collection?.members &&
            result.collection.members.length > 0 && (
              <div
                style={{
                  backgroundColor: "white",
                  padding: "10px",
                  borderRadius: "4px",
                  marginBottom: "15px",
                  border: "1px solid #c3e6cb",
                }}
              >
                <strong>👥 Members:</strong> {result.collection.members.length}
              </div>
            )}
        </div>
      )}

      {/* Raw Data Display (for debugging) */}
      {result && (
        <div
          style={{
            padding: "20px",
            backgroundColor: "#f8f9fa",
            borderRadius: "8px",
            border: "1px solid #dee2e6",
          }}
        >
          <h4 style={{ margin: "0 0 15px 0" }}>🔍 Raw Data (Debug)</h4>
          <details>
            <summary
              style={{
                cursor: "pointer",
                fontWeight: "bold",
                marginBottom: "10px",
              }}
            >
              Click to show raw collection data
            </summary>
            <pre
              style={{
                backgroundColor: "white",
                padding: "15px",
                borderRadius: "4px",
                fontSize: "11px",
                overflow: "auto",
                maxHeight: "400px",
                border: "1px solid #dee2e6",
                fontFamily: "monospace",
              }}
            >
              {JSON.stringify(result, null, 2)}
            </pre>
          </details>
        </div>
      )}

      {/* Quick Test IDs */}
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
          Try entering a collection ID that you created previously:
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
            onClick={() =>
              setCollectionId("7f558adb-57b6-11f0-8b98-c60a0c48537c")
            }
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
            Use Sample ID
          </button>
          <span style={{ fontSize: "12px", color: "#666" }}>
            or enter your own collection ID above
          </span>
        </div>
      </div>
    </div>
  );
};

export default GetCollectionManagerExample;
