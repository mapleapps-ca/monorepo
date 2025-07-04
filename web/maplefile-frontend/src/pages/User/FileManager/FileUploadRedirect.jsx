// File: src/pages/User/FileManager/FileUploadRedirect.jsx
// Simplified redirect component since FileManager now handles uploads directly
import React, { useEffect } from "react";
import { useNavigate } from "react-router";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const FileUploadRedirect = () => {
  const navigate = useNavigate();

  // Auto-redirect to file manager after a short delay
  useEffect(() => {
    const timer = setTimeout(() => {
      navigate("/files");
    }, 3000);

    return () => clearTimeout(timer);
  }, [navigate]);

  return (
    <div
      style={{
        padding: "20px",
        maxWidth: "600px",
        margin: "0 auto",
        textAlign: "center",
      }}
    >
      <h1>File Upload</h1>

      <div
        style={{
          backgroundColor: "#f8f9fa",
          padding: "40px",
          borderRadius: "8px",
          marginTop: "30px",
          border: "2px dashed #dee2e6",
        }}
      >
        <div style={{ fontSize: "48px", marginBottom: "20px" }}>📁</div>

        <h3>Uploads are now handled in the File Manager</h3>
        <p style={{ color: "#666", marginBottom: "30px" }}>
          The file upload feature has been integrated directly into the File
          Manager. You can now upload files by:
        </p>

        <div
          style={{
            textAlign: "left",
            backgroundColor: "white",
            padding: "20px",
            borderRadius: "4px",
            marginBottom: "30px",
          }}
        >
          <ul style={{ margin: 0, paddingLeft: "20px" }}>
            <li>
              <strong>Navigate to a folder</strong> in the File Manager
            </li>
            <li>
              <strong>Click "Upload Files"</strong> button to select files
            </li>
            <li>
              <strong>Drag and drop files</strong> directly into the upload area
            </li>
            <li>
              <strong>Use the "Show Upload Area"</strong> for easier drag & drop
            </li>
          </ul>
        </div>

        <div style={{ display: "flex", gap: "10px", justifyContent: "center" }}>
          <button
            onClick={() => navigate("/files")}
            style={{
              padding: "12px 30px",
              backgroundColor: "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              fontSize: "16px",
              cursor: "pointer",
            }}
          >
            Go to File Manager
          </button>

          <button
            onClick={() => navigate("/dashboard")}
            style={{
              padding: "12px 30px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              fontSize: "16px",
              cursor: "pointer",
            }}
          >
            Back to Dashboard
          </button>
        </div>

        <p
          style={{
            fontSize: "12px",
            color: "#999",
            marginTop: "20px",
            fontStyle: "italic",
          }}
        >
          Redirecting to File Manager in 3 seconds...
        </p>
      </div>

      <div style={{ marginTop: "40px", color: "#666" }}>
        <h4 style={{ marginTop: 0 }}>🔐 Security Features</h4>
        <ul style={{ textAlign: "left", marginBottom: 0 }}>
          <li>All files are encrypted on your device before upload</li>
          <li>File names and content are end-to-end encrypted</li>
          <li>Your password never leaves your device</li>
          <li>Files support versioning and soft deletion with recovery</li>
        </ul>
      </div>
    </div>
  );
};

const ProtectedFileUploadRedirect = withPasswordProtection(FileUploadRedirect);
export default ProtectedFileUploadRedirect;
