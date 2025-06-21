// monorepo/web/maplefile-frontend/src/pages/Anonymous/Index/IndexPage.jsx
import { useState, useEffect } from "react";
import { Link } from "react-router";

function IndexPage() {
  const [authMessage, setAuthMessage] = useState("");

  useEffect(() => {
    // Check for auth redirect message
    const message = sessionStorage.getItem("auth_redirect_message");
    if (message) {
      setAuthMessage(message);
      sessionStorage.removeItem("auth_redirect_message");

      // Clear message after 10 seconds
      const timer = setTimeout(() => {
        setAuthMessage("");
      }, 10000);

      return () => clearTimeout(timer);
    }
  }, []);

  return (
    <>
      <h1>MapleApps</h1>

      {/* Show auth message if present */}
      {authMessage && (
        <div
          style={{
            background: "#fff3cd",
            border: "1px solid #ffeaa7",
            color: "#856404",
            padding: "12px",
            borderRadius: "4px",
            marginBottom: "20px",
          }}
        >
          <strong>Authentication Required:</strong> {authMessage}
        </div>
      )}

      <div>
        <p>Welcome to MapleApps - Secure End-to-End Encrypted File Storage</p>
        <div>
          <h3>Get Started</h3>
          <p>Choose an option below to access your secure account:</p>

          <div>
            <Link to="/login">
              <button>Login</button>
            </Link>
            <span> or </span>
            <Link to="/register">
              <button>Register</button>
            </Link>
          </div>

          <div>
            <h4>Features:</h4>
            <ul>
              <li>🔐 End-to-end encryption with ChaCha20-Poly1305</li>
              <li>🔑 X25519 elliptic curve key exchange</li>
              <li>🛡️ Encrypted authentication tokens</li>
              <li>📁 Secure file storage and sharing</li>
              <li>🔄 Automatic background token refresh</li>
              <li>🔒 Client-side cryptographic operations</li>
            </ul>
          </div>

          <div>
            <h4>Security Information:</h4>
            <p>
              Your data is protected with military-grade encryption. All
              cryptographic operations happen locally in your browser, ensuring
              your passwords and private keys never leave your device.
            </p>
            <p>
              <strong>Session Security:</strong> For maximum security, you'll
              need to re-enter your password if you refresh the page or navigate
              away and come back. This ensures your encryption keys are never
              permanently stored.
            </p>
          </div>
        </div>
      </div>
    </>
  );
}

export default IndexPage;
