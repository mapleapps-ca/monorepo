// monorepo/web/maplefile-frontend/src/pages/User/Dashboard/Dashboard.jsx
import React, { useState } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";

// Simple inline debug component
const TokenDebugComponent = () => {
  const { localStorageService, authService } = useServices();
  const [tokenInfo, setTokenInfo] = useState({});
  const [workerStatus, setWorkerStatus] = useState({});
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showDebug, setShowDebug] = useState(false);

  const refreshTokenInfo = () => {
    const info = {
      // Encrypted token system
      encryptedTokens: localStorageService.getEncryptedTokens(),
      tokenNonce: localStorageService.getTokenNonce(),
      hasEncryptedTokens: !!(
        localStorageService.getEncryptedTokens() &&
        localStorageService.getTokenNonce()
      ),

      // Token expiry info
      ...localStorageService.getTokenExpiryInfo(),

      // Authentication status
      isAuthenticated: localStorageService.isAuthenticated(),
      userEmail: localStorageService.getUserEmail(),

      // Legacy tokens (should be empty)
      legacyAccessToken: localStorage.getItem("mapleapps_access_token"),
      legacyRefreshToken: localStorage.getItem("mapleapps_refresh_token"),
    };
    setTokenInfo(info);
  };

  const getWorkerStatus = async () => {
    try {
      const status = await authService.getWorkerStatus();
      setWorkerStatus(status);
    } catch (error) {
      setWorkerStatus({ error: error.message });
    }
  };

  const testTokenRefresh = async () => {
    setIsRefreshing(true);
    try {
      console.log("[Debug] Testing token refresh...");
      await authService.refreshToken();
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
      <div>
        <button onClick={() => setShowDebug(true)}>
          🔍 Show Token Debug Info
        </button>
      </div>
    );
  }

  return (
    <div>
      <div>
        <h3>🔍 Token System Debug Info</h3>
        <button onClick={() => setShowDebug(false)}>✕ Hide</button>
      </div>

      {/* Worker Status Alert */}
      {workerStatus.workerDisabled && (
        <div>
          ⚠️ <strong>Worker Disabled:</strong> {workerStatus.error}
        </div>
      )}

      <div>
        <h4>Encrypted Token Status</h4>
        <div>
          <div>
            <strong>Has Encrypted Tokens:</strong>
          </div>
          <div>{tokenInfo.hasEncryptedTokens ? "✅ YES" : "❌ NO"}</div>

          <div>
            <strong>Is Authenticated:</strong>
          </div>
          <div>{tokenInfo.isAuthenticated ? "✅ YES" : "❌ NO"}</div>

          <div>
            <strong>Worker Status:</strong>
          </div>
          <div>
            {workerStatus.isInitialized ? "✅ Initialized" : "❌ Failed"}
          </div>

          <div>
            <strong>Access Token Expired:</strong>
          </div>
          <div>{tokenInfo.accessTokenExpired ? "❌ YES" : "✅ NO"}</div>

          <div>
            <strong>Refresh Token Expired:</strong>
          </div>
          <div>{tokenInfo.refreshTokenExpired ? "❌ YES" : "✅ NO"}</div>
        </div>
      </div>

      <div>
        <h4>Worker Status Details</h4>
        <pre>{JSON.stringify(workerStatus, null, 2)}</pre>
      </div>

      <div>
        <button onClick={refreshTokenInfo}>🔄 Refresh Info</button>

        <button
          onClick={testTokenRefresh}
          disabled={isRefreshing || !tokenInfo.hasEncryptedTokens}
        >
          {isRefreshing ? "⏳ Refreshing..." : "🔄 Test Token Refresh"}
        </button>

        <button onClick={() => authService.forceTokenCheck()}>
          🔍 Force Token Check
        </button>
      </div>

      <div>
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
    <div>
      <span>
        <strong>{label}:</strong>
      </span>
      <span>
        {success ? "✅ PASS" : "❌ FAIL"}
        {error && ` (${error})`}
      </span>
    </div>
  );

  if (!showTest) {
    return (
      <div>
        <button onClick={() => setShowTest(true)}>
          🧪 Run Worker Diagnostic Test
        </button>
      </div>
    );
  }

  return (
    <div>
      <div>
        <h3>🧪 Worker Diagnostic Test</h3>
        <button onClick={() => setShowTest(false)}>✕ Hide</button>
      </div>

      <p>
        This test will help diagnose why the authentication worker is not
        initializing.
      </p>

      <button onClick={runWorkerTests} disabled={testing}>
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
            <div>
              <strong>Worker Message Received:</strong>
              <pre>
                {JSON.stringify(testResults.workerMessageData, null, 2)}
              </pre>
            </div>
          )}

          <div>
            <strong>🔧 Troubleshooting Tips:</strong>
            <ul>
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
  const { authService, localStorageService } = useServices();
  const { logout } = useAuth();
  const userEmail = localStorageService.getUserEmail();

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  return (
    <div>
      <h2>Welcome to MapleApps!</h2>
      <p>You have successfully completed the secure E2EE login process</p>

      <div>
        {/* Welcome Section */}
        <div>
          <div>
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
          <p>
            Welcome back, <strong>{userEmail}</strong>
          </p>
          <p>
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
        <div>
          <h3>What's Next?</h3>
          <div>
            <div>
              <div>🔐</div>
              <h4>Encrypted Token System</h4>
              <p>
                Your authentication tokens are now encrypted end-to-end for
                maximum security
              </p>
            </div>
            <div>
              <div>📁</div>
              <h4>Secure File Storage</h4>
              <p>Upload and encrypt your files with client-side encryption</p>
            </div>
            <div>
              <div>🔄</div>
              <h4>Background Token Refresh</h4>
              <p>
                Tokens are automatically refreshed in the background using the
                new API
              </p>
            </div>
            <div>
              <div>🛡️</div>
              <h4>Enhanced Security</h4>
              <p>No plaintext tokens are ever stored on your device</p>
            </div>
          </div>
        </div>

        {/* Security Summary */}
        <div>
          <h3>Your Security Details</h3>
          <div>
            <div>
              <span>Encryption:</span>
              <span>ChaCha20-Poly1305</span>
            </div>
            <div>
              <span>Key Exchange:</span>
              <span>X25519 ECDH</span>
            </div>
            <div>
              <span>Token System:</span>
              <span>Encrypted E2EE</span>
            </div>
            <div>
              <span>Authentication:</span>
              <span>3-Step E2EE Process</span>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div>
          <button onClick={handleLogout}>Logout</button>
          <button onClick={() => window.location.reload()}>Refresh Page</button>
        </div>

        {/* Footer Note */}
        <div>
          <p>
            🎉 <strong>Congratulations!</strong> You've successfully implemented
            the new encrypted token system with production-grade end-to-end
            encryption.
          </p>
          <p>
            <small>
              Your application now uses encrypted authentication tokens that are
              automatically refreshed in the background using the new API
              endpoints. Check the debug section above to verify the token
              system is working correctly.
            </small>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
