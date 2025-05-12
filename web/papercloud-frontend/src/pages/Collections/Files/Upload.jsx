// src/pages/Collections/Files/Upload.jsx
import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router";
import { collectionsAPI } from "../../../services/collectionApi";
import { fileAPI } from "../../../services/fileApi";
import { useAuth } from "../../../contexts/AuthContext"; // Import useAuth

function FileUploadPage() {
  const { collectionId } = useParams();
  const navigate = useNavigate();
  const { masterKey, sodium } = useAuth(); // Get masterKey and sodium from AuthContext

  const [collection, setCollection] = useState(null);
  const [decryptedCollectionKey, setDecryptedCollectionKey] = useState(null);
  const [file, setFile] = useState(null);
  const [fileName, setFileName] = useState("");
  const [loading, setLoading] = useState(true); // For fetching collection
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState(null);
  const [progress, setProgress] = useState(0);

  // Fetch collection details and decrypt its key
  useEffect(() => {
    const fetchAndPrepareCollection = async () => {
      if (!collectionId || !masterKey || !sodium) {
        if (!masterKey)
          setError(
            "Master key not available. Please ensure you are logged in correctly.",
          );
        if (!sodium) setError("Encryption library not ready.");
        setLoading(false);
        return;
      }
      try {
        setLoading(true);
        const collectionData = await collectionsAPI.getCollection(
          collectionId,
          masterKey,
        );
        setCollection(collectionData);
        if (collectionData && collectionData.decryptedCollectionKey) {
          setDecryptedCollectionKey(collectionData.decryptedCollectionKey);
          console.log("Collection key decrypted for upload page.");
        } else if (collectionData && collectionData.decryptionError) {
          setError(
            `Failed to prepare collection for upload: ${collectionData.decryptionError}`,
          );
        } else if (!collectionData) {
          setError("Collection not found.");
        } else {
          setError("Failed to decrypt collection key. Cannot upload files.");
        }
      } catch (err) {
        console.error("Error fetching/preparing collection:", err);
        setError("Failed to load collection details for upload.");
      } finally {
        setLoading(false);
      }
    };

    fetchAndPrepareCollection();
  }, [collectionId, masterKey, sodium]);

  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0];
    if (selectedFile) {
      setFile(selectedFile);
      setFileName(selectedFile.name);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!file) {
      setError("Please select a file to upload");
      return;
    }
    if (!decryptedCollectionKey) {
      setError("Collection key not available. Cannot encrypt file.");
      return;
    }
    if (!sodium) {
      setError("Encryption library not initialized");
      return;
    }

    try {
      setUploading(true);
      setError(null);
      setProgress(10);

      console.log("Starting file upload with decrypted collection key...");

      // The fileAPI.uploadFile now expects the decryptedCollectionKey
      await fileAPI.uploadFile(
        file,
        collectionId,
        decryptedCollectionKey, // Pass the decrypted collection key
        (p) => setProgress(p), // Progress callback
      );

      console.log("Upload complete!");
      setTimeout(() => {
        navigate(`/collections/${collectionId}/files`);
      }, 500);
    } catch (err) {
      console.error("Error uploading file:", err);
      setError(err.message || "Failed to upload file");
    } finally {
      setUploading(false);
      setProgress(0); // Reset progress
    }
  };

  if (loading) return <div>Loading collection details...</div>;

  if (error && !uploading) {
    return (
      <div>
        <div style={{ color: "red", marginBottom: "15px" }}>{error}</div>
        <button onClick={() => navigate(`/collections/${collectionId}/files`)}>
          Back to Files
        </button>
      </div>
    );
  }

  if (!collection) return <div>Collection not found or access denied.</div>;
  if (!decryptedCollectionKey && !loading)
    return <div>Error: Collection key could not be prepared for upload.</div>;

  return (
    <div>
      <h1>Upload File to {collection?.name || "Collection"}</h1>
      {/* ... rest of the form is similar, ensure disabled states use !decryptedCollectionKey ... */}
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "20px" }}>
          <label
            htmlFor="file"
            style={{ display: "block", marginBottom: "5px" }}
          >
            Select File:
          </label>
          <input
            type="file"
            id="file"
            onChange={handleFileChange}
            disabled={uploading || !decryptedCollectionKey}
            style={{ display: "block", marginBottom: "10px" }}
          />
          {fileName && (
            <div style={{ fontSize: "0.9rem", marginTop: "5px" }}>
              Selected: {fileName}
            </div>
          )}
        </div>

        {uploading && (
          <div style={{ marginBottom: "15px" }}>
            <div
              style={{
                height: "10px",
                background: "#eee",
                borderRadius: "5px",
              }}
            >
              <div
                style={{
                  height: "100%",
                  width: `${progress}%`,
                  background: "#4CAF50",
                  borderRadius: "5px",
                  transition: "width 0.3s ease-in-out",
                }}
              />
            </div>
            <div style={{ textAlign: "center", marginTop: "5px" }}>
              {progress}% -{" "}
              {progress < 100 ? "Encrypting and uploading..." : "Finalizing..."}
            </div>
          </div>
        )}

        <div style={{ display: "flex", gap: "10px" }}>
          <button
            type="submit"
            disabled={!file || uploading || !decryptedCollectionKey || !sodium}
            style={{
              padding: "8px 16px",
              background:
                !file || uploading || !decryptedCollectionKey || !sodium
                  ? "#cccccc"
                  : "#4CAF50",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor:
                !file || uploading || !decryptedCollectionKey || !sodium
                  ? "not-allowed"
                  : "pointer",
            }}
          >
            {uploading ? "Uploading..." : "Upload File"}
          </button>

          <button
            type="button"
            onClick={() => navigate(`/collections/${collectionId}/files`)}
            disabled={uploading}
            style={{
              padding: "8px 16px",
              background: "#f44336",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: uploading ? "not-allowed" : "pointer",
            }}
          >
            Cancel
          </button>
        </div>
      </form>
      <div style={{ marginTop: "20px", fontSize: "0.8rem", color: "#666" }}>
        <p>
          Files are encrypted end-to-end. Only users with access to this
          collection and the necessary keys can decrypt and view the files.
        </p>
      </div>
    </div>
  );
}

export default FileUploadPage;
