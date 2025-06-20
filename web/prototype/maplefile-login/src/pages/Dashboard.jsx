import React, { useState } from "react";
import { useNavigate } from "react-router";
import Layout from "../components/Layout.jsx";
import AuthService from "../services/authService.jsx";
import LocalStorageService from "../services/localStorageService.jsx";

// Simple inline debug component
const TokenDebugComponent = () => {
  const [tokenInfo, setTokenInfo] = useState({});
  const [workerStatus, setWorkerStatus] = useState({});
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showDebug, setShowDebug] = useState(false);

  const refreshTokenInfo = () => {
    const info = {
      // Encrypted token system
      encryptedTokens: LocalStorageService.getEncryptedTokens(),
      tokenNonce: LocalStorageService.getTokenNonce(),
      hasEncryptedTokens: !!(
        LocalStorageService.getEncryptedTokens() &&
        LocalStorageService.getTokenNonce()
      ),

      // Token expiry info
      ...LocalStorageService.getTokenExpiryInfo(),

      // Authentication status
      isAuthenticated: LocalStorageService.isAuthenticated(),
      userEmail: LocalStorageService.getUserEmail(),

      // Legacy tokens (should be empty)
      legacyAccessToken: localStorage.getItem("mapleapps_access_token"),
      legacyRefreshToken: localStorage.getItem("mapleapps_refresh_token"),
    };
    setTokenInfo(info);
  };

  const getWorkerStatus = async () => {
    try {
      const status = await AuthService.getWorkerStatus();
      setWorkerStatus(status);
    } catch (error) {
      setWorkerStatus({ error: error.message });
    }
  };

  const testTokenRefresh = async () => {
    setIsRefreshing(true);
    try {
      console.log("[Debug] Testing token refresh...");
      await AuthService.refreshToken();
      console.log("[Debug] Token refresh successful");
      refreshTokenInfo();
      alert("Token refresh successful! Check console for details.");
    } catch (error) {
      console.error("[Debug] Token refresh failed:", error);
      alert(`Token refresh failed: ${error.message}`);
    } finally {
      setIsRefreshing(false);
    }
  };

  React.useEffect(() => {
    if (showDebug) {
      refreshTokenInfo();
      getWorkerStatus();
    }
  }, [showDebug]);

  if (!showDebug) {
    return (
      <div style={{ textAlign: "center", marginTop: "20px" }}>
        <button
          onClick={() => setShowDebug(true)}
          className="btn btn-secondary"
        >
          🔍 Show Token Debug Info
        </button>
      </div>
    );
  }

  return (
    <div
      style={{
        backgroundColor: "#f8f9fa",
        padding: "20px",
        borderRadius: "8px",
        marginTop: "20px",
        border: "1px solid #e5e5e7",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "15px",
        }}
      >
        <h3>🔍 Token System Debug Info</h3>
        <button
          onClick={() => setShowDebug(false)}
          style={{
            padding: "4px 8px",
            backgroundColor: "#6c757d",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
            fontSize: "12px",
          }}
        >
          ✕ Hide
        </button>
      </div>

      {/* Worker Status Alert */}
      {workerStatus.workerDisabled && (
        <div
          style={{
            backgroundColor: "#fff3cd",
            color: "#856404",
            padding: "10px",
            borderRadius: "4px",
            marginBottom: "15px",
            border: "1px solid #ffeaa7",
          }}
        >
          ⚠️ <strong>Worker Disabled:</strong> {workerStatus.error}
        </div>
      )}

      <div style={{ marginBottom: "20px" }}>
        <h4>Encrypted Token Status</h4>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "200px 1fr",
            gap: "8px",
            fontSize: "14px",
          }}
        >
          <div style={{ fontWeight: "bold" }}>Has Encrypted Tokens:</div>
          <div
            style={{
              color: tokenInfo.hasEncryptedTokens ? "green" : "red",
              fontWeight: "bold",
            }}
          >
            {tokenInfo.hasEncryptedTokens ? "✅ YES" : "❌ NO"}
          </div>

          <div style={{ fontWeight: "bold" }}>Is Authenticated:</div>
          <div
            style={{
              color: tokenInfo.isAuthenticated ? "green" : "red",
              fontWeight: "bold",
            }}
          >
            {tokenInfo.isAuthenticated ? "✅ YES" : "❌ NO"}
          </div>

          <div style={{ fontWeight: "bold" }}>Worker Status:</div>
          <div
            style={{
              color: workerStatus.isInitialized ? "green" : "red",
              fontWeight: "bold",
            }}
          >
            {workerStatus.isInitialized ? "✅ Initialized" : "❌ Failed"}
          </div>

          <div style={{ fontWeight: "bold" }}>Access Token Expired:</div>
          <div
            style={{ color: tokenInfo.accessTokenExpired ? "red" : "green" }}
          >
            {tokenInfo.accessTokenExpired ? "❌ YES" : "✅ NO"}
          </div>

          <div style={{ fontWeight: "bold" }}>Refresh Token Expired:</div>
          <div
            style={{ color: tokenInfo.refreshTokenExpired ? "red" : "green" }}
          >
            {tokenInfo.refreshTokenExpired ? "❌ YES" : "✅ NO"}
          </div>
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <h4>Worker Status Details</h4>
        <div
          style={{
            fontFamily: "monospace",
            fontSize: "12px",
            backgroundColor: "#fff",
            padding: "10px",
            borderRadius: "4px",
          }}
        >
          <pre>{JSON.stringify(workerStatus, null, 2)}</pre>
        </div>
      </div>

      <div style={{ display: "flex", gap: "10px", flexWrap: "wrap" }}>
        <button
          onClick={refreshTokenInfo}
          className="btn btn-primary"
          style={{ fontSize: "12px", padding: "6px 12px" }}
        >
          🔄 Refresh Info
        </button>

        <button
          onClick={testTokenRefresh}
          disabled={isRefreshing || !tokenInfo.hasEncryptedTokens}
          className="btn btn-secondary"
          style={{ fontSize: "12px", padding: "6px 12px" }}
        >
          {isRefreshing ? "⏳ Refreshing..." : "🔄 Test Token Refresh"}
        </button>

        <button
          onClick={() => AuthService.forceTokenCheck()}
          className="btn btn-secondary"
          style={{ fontSize: "12px", padding: "6px 12px" }}
        >
          🔍 Force Token Check
        </button>
      </div>

      <div
        style={{
          marginTop: "15px",
          padding: "10px",
          backgroundColor: "#e9ecef",
          borderRadius: "4px",
          fontSize: "12px",
        }}
      >
        <strong>🔧 Debug Tips:</strong> Check browser console for detailed logs.
        Look for [AuthWorker], [AuthService], and [LocalStorageService]
        messages.
      </div>
    </div>
  );
};

// Worker Test Component
const WorkerTestComponent = () => {
  const [testResults, setTestResults] = useState({});
  const [testing, setTesting] = useState(false);
  const [showTest, setShowTest] = useState(false);

  const runWorkerTests = async () => {
    setTesting(true);
    const results = {};

    // Test 1: Check if worker file exists
    try {
      const response = await fetch("/auth-worker.js");
      results.workerFileExists = response.ok;
      results.workerFileStatus = response.status;
      if (!response.ok) {
        results.workerFileError = `HTTP ${response.status}`;
      }
    } catch (error) {
      results.workerFileExists = false;
      results.workerFileError = error.message;
    }

    // Test 2: Try to create a simple worker
    try {
      const simpleWorker = new Worker("/auth-worker.js");
      results.workerCreation = true;

      // Test 3: Listen for messages from worker
      const messagePromise = new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error("Worker message timeout"));
        }, 3000);

        simpleWorker.addEventListener("message", (event) => {
          clearTimeout(timeout);
          resolve(event.data);
        });

        simpleWorker.addEventListener("error", (error) => {
          clearTimeout(timeout);
          reject(error);
        });
      });

      try {
        const workerMessage = await messagePromise;
        results.workerMessage = true;
        results.workerMessageData = workerMessage;
      } catch (msgError) {
        results.workerMessage = false;
        results.workerMessageError = msgError.message;
      }

      simpleWorker.terminate();
    } catch (creationError) {
      results.workerCreation = false;
      results.workerCreationError = creationError.message;
    }

    // Test 4: Check browser support
    results.workerSupport = typeof Worker !== "undefined";
    results.broadcastChannelSupport = typeof BroadcastChannel !== "undefined";

    setTestResults(results);
    setTesting(false);
  };

  const renderTestResult = (label, success, error = null) => (
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        padding: "8px",
        backgroundColor: success ? "#d4edda" : "#f8d7da",
        borderRadius: "4px",
        marginBottom: "8px",
      }}
    >
      <span style={{ fontWeight: "bold" }}>{label}:</span>
      <span style={{ color: success ? "green" : "red" }}>
        {success ? "✅ PASS" : "❌ FAIL"}
        {error && ` (${error})`}
      </span>
    </div>
  );

  if (!showTest) {
    return (
      <div style={{ textAlign: "center", marginTop: "20px" }}>
        <button onClick={() => setShowTest(true)} className="btn btn-danger">
          🧪 Run Worker Diagnostic Test
        </button>
      </div>
    );
  }

  return (
    <div
      style={{
        backgroundColor: "#fff3cd",
        padding: "20px",
        borderRadius: "8px",
        marginTop: "20px",
        border: "1px solid #ffeaa7",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "15px",
        }}
      >
        <h3>🧪 Worker Diagnostic Test</h3>
        <button
          onClick={() => setShowTest(false)}
          style={{
            padding: "4px 8px",
            backgroundColor: "#6c757d",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
            fontSize: "12px",
          }}
        >
          ✕ Hide
        </button>
      </div>

      <p>
        This test will help diagnose why the authentication worker is not
        initializing.
      </p>

      <button
        onClick={runWorkerTests}
        disabled={testing}
        style={{
          padding: "10px 20px",
          backgroundColor: "#007bff",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: testing ? "not-allowed" : "pointer",
          marginBottom: "20px",
        }}
      >
        {testing ? "🔄 Testing..." : "🧪 Run Worker Tests"}
      </button>

      {Object.keys(testResults).length > 0 && (
        <div>
          <h4>Test Results:</h4>

          {renderTestResult(
            "Browser Worker Support",
            testResults.workerSupport,
          )}

          {renderTestResult(
            "BroadcastChannel Support",
            testResults.broadcastChannelSupport,
          )}

          {renderTestResult(
            "Worker File Accessible",
            testResults.workerFileExists,
            testResults.workerFileError,
          )}

          {renderTestResult(
            "Worker Creation",
            testResults.workerCreation,
            testResults.workerCreationError,
          )}

          {renderTestResult(
            "Worker Communication",
            testResults.workerMessage,
            testResults.workerMessageError,
          )}

          {testResults.workerMessageData && (
            <div
              style={{
                marginTop: "15px",
                padding: "10px",
                backgroundColor: "#f8f9fa",
                borderRadius: "4px",
                border: "1px solid #e9ecef",
              }}
            >
              <strong>Worker Message Received:</strong>
              <pre style={{ fontSize: "12px", marginTop: "5px" }}>
                {JSON.stringify(testResults.workerMessageData, null, 2)}
              </pre>
            </div>
          )}

          <div
            style={{
              marginTop: "15px",
              padding: "10px",
              backgroundColor: "#e9ecef",
              borderRadius: "4px",
              fontSize: "14px",
            }}
          >
            <strong>🔧 Troubleshooting Tips:</strong>
            <ul style={{ margin: "5px 0", paddingLeft: "20px" }}>
              <li>
                Make sure <code>auth-worker.js</code> exists in the{" "}
                <code>public</code> folder
              </li>
              <li>Check browser console for JavaScript errors in the worker</li>
              <li>
                Verify your development server is serving files from the public
                folder
              </li>
              <li>Try disabling browser extensions that might block workers</li>
              <li>
                Check if your browser supports Web Workers and BroadcastChannel
              </li>
            </ul>
          </div>
        </div>
      )}
    </div>
  );
};

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
            operations were performed locally in your browser.{" "}
            <strong>Your tokens are now stored encrypted!</strong>
          </p>
        </div>

        {/* Token Debug Component */}
        <TokenDebugComponent />

        {/* Worker Test Component */}
        <WorkerTestComponent />

        {/* Features Section */}
        <div className="features-section">
          <h3>What's Next?</h3>
          <div className="feature-grid">
            <div className="feature-card">
              <div className="feature-icon">🔐</div>
              <h4>Encrypted Token System</h4>
              <p>
                Your authentication tokens are now encrypted end-to-end for
                maximum security
              </p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">📁</div>
              <h4>Secure File Storage</h4>
              <p>Upload and encrypt your files with client-side encryption</p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">🔄</div>
              <h4>Background Token Refresh</h4>
              <p>
                Tokens are automatically refreshed in the background using the
                new API
              </p>
            </div>
            <div className="feature-card">
              <div className="feature-icon">🛡️</div>
              <h4>Enhanced Security</h4>
              <p>No plaintext tokens are ever stored on your device</p>
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
              <span className="stat-label">Token System:</span>
              <span className="stat-value">Encrypted E2EE</span>
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
            the new encrypted token system with production-grade end-to-end
            encryption.
          </p>
          <p className="tech-note">
            <small>
              Your application now uses encrypted authentication tokens that are
              automatically refreshed in the background using the new API
              endpoints. Check the debug section above to verify the token
              system is working correctly.
            </small>
          </p>
        </div>
      </div>
    </Layout>
  );
};

export default Dashboard;
