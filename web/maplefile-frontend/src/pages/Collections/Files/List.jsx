// src/pages/Collections/Files/List.jsx
import { useState, useEffect } from "react";
import { useParams, Link, useNavigate } from "react-router";
import { collectionsAPI } from "../../../services/collectionApi";
import { fileAPI } from "../../../services/fileApi";
import { useAuth } from "../../../contexts/AuthContext"; // Import useAuth

function CollectionFileListPage() {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const { masterKey, sodium, isAuthenticated } = useAuth(); // Get masterKey and sodium

  const [collection, setCollection] = useState(null);
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [downloadingFile, setDownloadingFile] = useState(null); // Stores ID of file being downloaded

  useEffect(() => {
    if (!isAuthenticated || !masterKey || !sodium) {
      setError("Authentication or encryption context not ready.");
      setLoading(false);
      return;
    }
    const fetchCollectionAndFiles = async () => {
      try {
        setLoading(true);
        // Fetch collection details (this will also decrypt collectionKey if masterKey is passed)
        const collectionData = await collectionsAPI.getCollection(
          collectionId,
          masterKey,
        );
        setCollection(collectionData);

        if (collectionData && collectionData.decryptionError) {
          setError(`Error with collection: ${collectionData.decryptionError}`);
          setLoading(false);
          return;
        }
        if (!collectionData) {
          setError("Collection not found.");
          setLoading(false);
          return;
        }

        // Fetch files in the collection
        const filesData = await collectionsAPI.listFiles(collectionId);
        setFiles(filesData.files || []);
      } catch (err) {
        console.error("Error fetching collection data:", err);
        setError("Failed to load collection data. " + err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchCollectionAndFiles();
  }, [collectionId, masterKey, sodium, isAuthenticated]);

  const handleDeleteFile = async (fileId) => {
    if (
      !confirm(
        "Are you sure you want to delete this file? This action cannot be undone.",
      )
    ) {
      return;
    }
    try {
      await fileAPI.deleteFile(fileId);
      // Refresh file list
      const filesData = await collectionsAPI.listFiles(collectionId);
      setFiles(filesData.files || []);
    } catch (err) {
      console.error("Error deleting file:", err);
      alert("Failed to delete file: " + err.message);
    }
  };

  const handleDownloadFile = async (fileId) => {
    if (!masterKey) {
      alert("Cannot download: Master key is not available.");
      return;
    }
    setDownloadingFile(fileId);
    try {
      console.log(`Attempting to download and decrypt file: ${fileId}`);
      // The fileAPI.downloadFile now handles all E2EE decryption steps
      const downloadResult = await fileAPI.downloadFile(fileId, masterKey);

      if (downloadResult.success) {
        alert(`File "${downloadResult.fileName}" downloaded successfully.`);
      } else {
        // This case might not be reached if downloadFile throws on failure
        alert(
          "File download completed, but there might have been an issue during decryption or saving.",
        );
      }
    } catch (err) {
      console.error("Error downloading file:", err);
      alert(`Failed to download file: ${err.message}`);
    } finally {
      setDownloadingFile(null);
    }
  };

  if (loading) return <div>Loading collection files...</div>;
  if (error) return <div>Error: {error}</div>;
  if (!collection && !loading)
    return <div>Collection not found or access denied.</div>;

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
        <h1>{collection?.name || "Collection"} Files</h1>
        <div>
          <button
            onClick={() => navigate(`/collections/${collectionId}/upload`)}
            disabled={!collection || collection.decryptionError} // Disable if collection key couldn't be decrypted
            style={{
              padding: "10px 15px",
              background:
                !collection || collection.decryptionError ? "#ccc" : "#4CAF50",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !collection || collection.decryptionError
                  ? "not-allowed"
                  : "pointer",
              display: "flex",
              alignItems: "center",
              gap: "5px",
            }}
          >
            <span>+</span> Upload File
          </button>
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <Link
          to="/collections"
          style={{
            color: "#666",
            textDecoration: "none",
            display: "inline-flex",
            alignItems: "center",
            gap: "5px",
          }}
        >
          ← Back to Collections
        </Link>
      </div>

      {files.length === 0 ? (
        <div
          style={{
            padding: "40px 20px",
            textAlign: "center",
            background: "#f9f9f9",
            borderRadius: "8px",
          }}
        >
          <p style={{ fontSize: "1.1rem", marginBottom: "15px" }}>
            No files in this collection yet.
          </p>
          <button
            onClick={() => navigate(`/collections/${collectionId}/upload`)}
            disabled={!collection || collection.decryptionError}
            style={{
              padding: "10px 15px",
              background:
                !collection || collection.decryptionError ? "#ccc" : "#4CAF50",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !collection || collection.decryptionError
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            Upload Your First File
          </button>
        </div>
      ) : (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fill, minmax(250px, 1fr))",
            gap: "20px",
          }}
          className="files-grid"
        >
          {files.map((file) => (
            <div
              key={file.id} // Assuming file object from API has `id`
              style={{
                border: "1px solid #ddd",
                borderRadius: "8px",
                padding: "15px",
                background: "white",
                boxShadow: "0 2px 4px rgba(0,0,0,0.05)",
              }}
              className="file-card"
            >
              {/* ... (file card rendering, ensure to use file.id for keys and actions) ... */}
              <div
                style={{
                  height: "120px",
                  background: "#f5f5f5",
                  borderRadius: "4px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  marginBottom: "10px",
                  position: "relative",
                }}
              >
                <svg
                  width="48"
                  height="48"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="#666"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                  <polyline points="14 2 14 8 20 8"></polyline>
                </svg>
                {downloadingFile === file.id && (
                  <div
                    style={{
                      position: "absolute",
                      top: 0,
                      left: 0,
                      right: 0,
                      bottom: 0,
                      background: "rgba(255, 255, 255, 0.8)",
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      borderRadius: "4px",
                    }}
                  >
                    <div className="loading-spinner">
                      <div
                        style={{
                          border: "3px solid #f3f3f3",
                          borderTop: "3px solid #4285F4",
                          borderRadius: "50%",
                          width: "24px",
                          height: "24px",
                          animation: "spin 1s linear infinite",
                        }}
                      ></div>
                    </div>
                  </div>
                )}
              </div>
              <h3
                style={{
                  margin: "0 0 5px 0",
                  fontSize: "1rem",
                  wordBreak: "break-word",
                }}
              >
                {/* We can't display original name until metadata is decrypted during download */}
                File ID: ...
                {file.file_id ? file.file_id.slice(-10) : file.id.slice(-6)}
              </h3>
              <p
                style={{
                  margin: "0 0 10px 0",
                  color: "#666",
                  fontSize: "0.9rem",
                }}
              >
                Size: {(file.encrypted_size / 1024).toFixed(1)} KB (encrypted)
              </p>
              <div
                style={{ display: "flex", gap: "10px" }}
                className="file-actions"
              >
                <button
                  onClick={() => handleDownloadFile(file.id)}
                  disabled={downloadingFile === file.id || !masterKey}
                  style={{
                    flex: "1",
                    padding: "8px",
                    background:
                      downloadingFile === file.id || !masterKey
                        ? "#ccc"
                        : "#4285F4",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor:
                      downloadingFile === file.id || !masterKey
                        ? "not-allowed"
                        : "pointer",
                    fontSize: "0.9rem",
                  }}
                >
                  {downloadingFile === file.id ? "Downloading..." : "Download"}
                </button>
                <button
                  onClick={() => handleDeleteFile(file.id)}
                  disabled={downloadingFile === file.id}
                  style={{
                    flex: "1",
                    padding: "8px",
                    background:
                      downloadingFile === file.id ? "#ccc" : "#f44336",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor:
                      downloadingFile === file.id ? "not-allowed" : "pointer",
                    fontSize: "0.9rem",
                  }}
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
      <style jsx global>{`
        @keyframes spin {
          0% {
            transform: rotate(0deg);
          }
          100% {
            transform: rotate(360deg);
          }
        }
      `}</style>
    </div>
  );
}

export default CollectionFileListPage;
