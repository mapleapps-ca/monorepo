// File: src/pages/User/FileManager/FileUploadRedirect.jsx
import React from "react";
import { useNavigate } from "react-router";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const FileUploadRedirect = () => {
  const navigate = useNavigate();

  return (
    <div
      style={{
        padding: "20px",
        maxWidth: "600px",
        margin: "0 auto",
        textAlign: "center",
      }}
    >
      <h1>Upload File</h1>

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

        <h3>Please select a folder first</h3>
        <p style={{ color: "#666", marginBottom: "30px" }}>
          Files must be uploaded to a specific folder. Navigate to the folder
          where you want to upload your files.
        </p>

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
      </div>

      <div style={{ marginTop: "40px", color: "#666" }}>
        <p>
          <strong>Tip:</strong> You can create a new folder in the File Manager
          if you need to organize your files.
        </p>
      </div>
    </div>
  );
};

const ProtectedFileUploadRedirect = withPasswordProtection(FileUploadRedirect);
export default ProtectedFileUploadRedirect;
