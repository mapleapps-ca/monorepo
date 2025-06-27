// File: src/pages/User/Collection/Detail.jsx
import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const CollectionDetail = () => {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const { collectionService, passwordStorageService } = useServices();

  const [collection, setCollection] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    loadCollection();
  }, [collectionId]);

  const loadCollection = async () => {
    if (!collectionId) {
      setError("No collection ID provided");
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError("");

      // Get stored password
      const password = passwordStorageService.getPassword();

      // Load collection with password
      const collectionData = await collectionService.getCollection(
        collectionId,
        password,
      );

      setCollection(collectionData);
      console.log("Collection loaded:", collectionData);
    } catch (err) {
      console.error("Failed to load collection:", err);
      setError(err.message || "Failed to load collection");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Loading Collection...</h1>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Error</h1>
        <p style={{ color: "red" }}>{error}</p>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
      </div>
    );
  }

  if (!collection) {
    return (
      <div style={{ padding: "20px" }}>
        <h1>Collection Not Found</h1>
        <button onClick={() => navigate("/collections")}>
          ← Back to Collections
        </button>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "800px" }}>
      {/* Header */}
      <div style={{ marginBottom: "30px" }}>
        <button
          onClick={() => navigate("/collections")}
          style={{ marginBottom: "20px" }}
        >
          ← Back to Collections
        </button>

        <div
          style={{
            display: "flex",
            alignItems: "center",
            marginBottom: "10px",
          }}
        >
          <span style={{ fontSize: "32px", marginRight: "15px" }}>
            {collection.collection_type === "album" ? "🖼️" : "📁"}
          </span>
          <h1 style={{ margin: 0 }}>{collection.name || "[Encrypted]"}</h1>
        </div>

        {collection.decrypt_error && (
          <p style={{ color: "red" }}>
            ⚠️ Decryption Error: {collection.decrypt_error}
          </p>
        )}
      </div>

      {/* Collection Details */}
      <div
        style={{
          backgroundColor: "#f8f9fa",
          padding: "20px",
          marginBottom: "20px",
        }}
      >
        <h3>Collection Information</h3>
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <tbody>
            <tr>
              <td
                style={{ padding: "8px", fontWeight: "bold", width: "200px" }}
              >
                ID:
              </td>
              <td style={{ padding: "8px" }}>{collection.id}</td>
            </tr>
            <tr>
              <td style={{ padding: "8px", fontWeight: "bold" }}>Name:</td>
              <td style={{ padding: "8px" }}>
                {collection.name || "[Encrypted]"}
              </td>
            </tr>
            <tr>
              <td style={{ padding: "8px", fontWeight: "bold" }}>Type:</td>
              <td style={{ padding: "8px" }}>{collection.collection_type}</td>
            </tr>
            <tr>
              <td style={{ padding: "8px", fontWeight: "bold" }}>Owner ID:</td>
              <td style={{ padding: "8px" }}>{collection.owner_id}</td>
            </tr>
            <tr>
              <td style={{ padding: "8px", fontWeight: "bold" }}>Created:</td>
              <td style={{ padding: "8px" }}>
                {collection.created_at
                  ? new Date(collection.created_at).toLocaleString()
                  : "Unknown"}
              </td>
            </tr>
            <tr>
              <td style={{ padding: "8px", fontWeight: "bold" }}>Modified:</td>
              <td style={{ padding: "8px" }}>
                {collection.modified_at
                  ? new Date(collection.modified_at).toLocaleString()
                  : "Unknown"}
              </td>
            </tr>
            <tr>
              <td style={{ padding: "8px", fontWeight: "bold" }}>Parent ID:</td>
              <td style={{ padding: "8px" }}>
                {collection.parent_id || "None (Root Level)"}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      {/* Members */}
      {collection.members && collection.members.length > 0 && (
        <div
          style={{
            backgroundColor: "#f8f9fa",
            padding: "20px",
            marginBottom: "20px",
          }}
        >
          <h3>Members ({collection.members.length})</h3>
          {collection.members.map((member, index) => (
            <div
              key={index}
              style={{
                padding: "10px",
                borderBottom: "1px solid #ddd",
                display: "flex",
                justifyContent: "space-between",
              }}
            >
              <div>
                <strong>{member.recipient_email}</strong>
                <div style={{ fontSize: "12px", color: "#666" }}>
                  Permission: {member.permission_level}
                  {member.is_inherited && " (Inherited)"}
                </div>
              </div>
              <div style={{ fontSize: "12px", color: "#666" }}>
                Added:{" "}
                {member.created_at
                  ? new Date(member.created_at).toLocaleDateString()
                  : "Unknown"}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Raw Data (for debugging) */}
      <details style={{ marginTop: "40px" }}>
        <summary>🔍 Raw Collection Data (Debug)</summary>
        <pre
          style={{
            backgroundColor: "#f8f9fa",
            padding: "15px",
            overflow: "auto",
            fontSize: "12px",
            marginTop: "10px",
          }}
        >
          {JSON.stringify(collection, null, 2)}
        </pre>
      </details>
    </div>
  );
};

export default withPasswordProtection(CollectionDetail);
