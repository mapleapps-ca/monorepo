// monorepo/web/maplefile-frontend/src/pages/Anonymous/Index/IndexPage.jsx
import { useState } from "react";
import { Link } from "react-router";
// import "./App.css";

function IndexPage() {
  return (
    <>
      <h1>MapleApps</h1>
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
          </div>
        </div>
      </div>
    </>
  );
}

export default IndexPage;
