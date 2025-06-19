import React from "react";
import { useNavigate } from "react-router";
import Layout from "../components/Layout.jsx";
import AuthService from "../services/authService.jsx";
import LocalStorageService from "../services/localStorageService.jsx";

const Dashboard = () => {
  const navigate = useNavigate();
  const userEmail = LocalStorageService.getUserEmail();

  const handleLogout = () => {
    AuthService.logout();
    navigate("/");
  };

  return (
    <Layout
      title="Welcome to MapleApps!"
      subtitle="You have successfully completed the secure E2EE login process"
    >
      <div className="dashboard-container">
        {/* Welcome Section */}
        <div className="welcome-section">
          <div className="success-icon">
            <svg
              width="64"
              height="64"
              viewBox="0 0 24 24"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <circle cx="12" cy="12" r="10" fill="#4CAF50" />
              <path
                d="M9 12l2 2 4-4"
                stroke="white"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
          </div>
          <h2>Login Successful!</h2>
          <p className="user-info">
            Welcome back, <strong>{userEmail}</strong>
          </p>
          <p className="security-note">
            Your session is protected with end-to-end encryption using
            ChaCha20-Poly1305 and X25519 key exchange. All cryptographic
            operations were performed locally in your browser.
          </p>
        </div>

        {/* Features Section */}
        <div className="features-section">
          <h3>What's Next?</h3>
          <div className="feature-grid">
            <div className="feature-card">
              <div className="feature-icon">📁</div>
              <h4>Secure File Storage</h4>
              <p>Upload and encrypt your files with client-side encryption</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">🔐</div>
              <h4>Password Manager</h4>
              <p>
                Store your passwords securely with zero-knowledge encryption
              </p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">💬</div>
              <h4>Encrypted Messaging</h4>
              <p>Send messages that only you and your recipient can read</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">🔑</div>
              <h4>Key Management</h4>
              <p>Manage your cryptographic keys and recovery phrases</p>
            </div>
          </div>
        </div>

        {/* Security Summary */}
        <div className="security-summary">
          <h3>Your Security Details</h3>
          <div className="security-stats">
            <div className="stat-item">
              <span className="stat-label">Encryption:</span>
              <span className="stat-value">ChaCha20-Poly1305</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">Key Exchange:</span>
              <span className="stat-value">X25519 ECDH</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">Key Derivation:</span>
              <span className="stat-value">PBKDF2 SHA-256</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">Authentication:</span>
              <span className="stat-value">3-Step E2EE Process</span>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="dashboard-actions">
          <button onClick={handleLogout} className="btn btn-secondary">
            Logout
          </button>
          <button
            onClick={() => window.location.reload()}
            className="btn btn-primary"
          >
            Refresh Page
          </button>
        </div>

        {/* Footer Note */}
        <div className="dashboard-footer">
          <p>
            🎉 <strong>Congratulations!</strong> You've successfully implemented
            production-grade end-to-end encryption in your React application.
          </p>
          <p className="tech-note">
            <small>
              This dashboard is a placeholder. You can now build your secure
              application knowing that user authentication and encryption are
              working perfectly.
            </small>
          </p>
        </div>
      </div>
    </Layout>
  );
};

export default Dashboard;
