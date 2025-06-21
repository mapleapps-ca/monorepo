// Example component demonstrating usage of authentication services
import React, { useState, useEffect } from "react";
import { useServices } from "../contexts/ServiceContext.jsx";
import useAuth from "../hooks/useAuth.js";

const AuthExample = () => {
  const { authService, cryptoService, tokenService, meService } = useServices();
  const {
    isAuthenticated,
    isLoading,
    user,
    tokenInfo,
    workerStatus,
    logout,
    manualRefresh,
    forceTokenCheck,
    getTokenHealth,
    getDebugInfo,
    hasSessionKeys,
    canMakeAuthenticatedRequests,
    canDecryptTokens,
  } = useAuth();

  // Local state for forms
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [ott, setOtt] = useState("");
  const [verifyData, setVerifyData] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState("email"); // email, ott, password
  const [userProfile, setUserProfile] = useState(null);

  // Registration form state
  const [registrationData, setRegistrationData] = useState({
    firstName: "",
    lastName: "",
    email: "",
    phone: "",
    country: "",
    password: "",
    agreeTerms: false,
  });

  // Clear error when changing steps
  useEffect(() => {
    setError("");
  }, [step]);

  // Handle email submission (Step 1)
  const handleEmailSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      console.log("Requesting OTT for:", email);
      await authService.requestOTT(email);
      setStep("ott");
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle OTT verification (Step 2)
  const handleOttSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      console.log("Verifying OTT:", ott);
      const response = await authService.verifyOTT(email, ott);
      setVerifyData(response);
      setStep("password");
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Handle password and complete login (Step 3)
  const handlePasswordSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      console.log("Decrypting challenge and completing login");

      // Decrypt the challenge
      const decryptedChallenge = await authService.decryptChallenge(
        password,
        verifyData,
      );

      // Complete login
      await authService.completeLogin(
        email,
        verifyData.challengeId,
        decryptedChallenge,
      );

      console.log("Login successful!");
      setStep("email"); // Reset form
      setEmail("");
      setPassword("");
      setOtt("");
      setVerifyData(null);
    } catch (err) {
      setError(err.message);
      console.error("Login failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Handle registration
  const handleRegistration = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      console.log("Generating E2EE data...");
      const e2eeData = await cryptoService.generateE2EEData(
        registrationData.password,
      );

      console.log("Registering user...");
      const userData = {
        beta_access_code: "BETA2024", // You might want to make this configurable
        first_name: registrationData.firstName,
        last_name: registrationData.lastName,
        email: registrationData.email,
        phone: registrationData.phone,
        country: registrationData.country,
        timezone: "America/Toronto", // You might want to detect this
        agree_terms_of_service: registrationData.agreeTerms,
        agree_promotions: false,
        agree_to_tracking_across_third_party_apps_and_services: false,
        module: 1, // 1 for MapleFile, 2 for PaperCloud
        salt: e2eeData.salt,
        publicKey: e2eeData.publicKey,
        encryptedMasterKey: e2eeData.encryptedMasterKey,
        encryptedPrivateKey: e2eeData.encryptedPrivateKey,
        encryptedRecoveryKey: e2eeData.encryptedRecoveryKey,
        masterKeyEncryptedWithRecoveryKey:
          e2eeData.masterKeyEncryptedWithRecoveryKey,
        verificationID: e2eeData.verificationID,
      };

      const result = await authService.registerUser(userData);
      console.log("Registration successful:", result);

      // Show recovery mnemonic to user
      alert(
        `Registration successful! Please save your recovery phrase: ${e2eeData.recoveryMnemonic}`,
      );

      // Reset form
      setRegistrationData({
        firstName: "",
        lastName: "",
        email: "",
        phone: "",
        country: "",
        password: "",
        agreeTerms: false,
      });
    } catch (err) {
      setError(err.message);
      console.error("Registration failed:", err);
    } finally {
      setLoading(false);
    }
  };

  // Handle logout
  const handleLogout = () => {
    logout();
    setUserProfile(null);
  };

  // Test token decryption
  const testTokenDecryption = async () => {
    try {
      const decryptedToken = await tokenService.getDecryptedAccessToken();
      if (decryptedToken) {
        alert(
          `Token decryption successful! Token preview: ${decryptedToken.substring(0, 50)}...`,
        );
      } else {
        alert("No decrypted token available (may need session keys)");
      }
    } catch (err) {
      alert(`Token decryption failed: ${err.message}`);
    }
  };

  // Load user profile
  const loadUserProfile = async () => {
    try {
      const profile = await meService.getCurrentUser();
      setUserProfile(profile);
    } catch (err) {
      console.error("Failed to load user profile:", err);
    }
  };

  // Token health info
  const tokenHealth = getTokenHealth();
  const debugInfo = getDebugInfo();

  if (isLoading) {
    return <div className="p-4">Loading authentication...</div>;
  }

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      <h1 className="text-3xl font-bold mb-6">MapleApps Authentication Demo</h1>

      {/* Authentication Status */}
      <div className="bg-white p-6 rounded-lg shadow">
        <h2 className="text-xl font-semibold mb-4">Authentication Status</h2>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <strong>Authenticated:</strong> {isAuthenticated ? "Yes" : "No"}
          </div>
          <div>
            <strong>User:</strong> {user?.email || "None"}
          </div>
          <div>
            <strong>Token System:</strong> {tokenInfo.tokenSystem || "None"}
          </div>
          <div>
            <strong>Has Encrypted Tokens:</strong>{" "}
            {tokenInfo.hasEncryptedTokens ? "Yes" : "No"}
          </div>
          <div>
            <strong>Has Session Keys:</strong> {hasSessionKeys ? "Yes" : "No"}
          </div>
          <div>
            <strong>Can Decrypt Tokens:</strong>{" "}
            {canDecryptTokens ? "Yes" : "No"}
          </div>
          <div>
            <strong>Can Make API Calls:</strong>{" "}
            {canMakeAuthenticatedRequests ? "Yes" : "No"}
          </div>
          <div>
            <strong>Access Token Expired:</strong>{" "}
            {tokenInfo.accessTokenExpired ? "Yes" : "No"}
          </div>
          <div>
            <strong>Worker Status:</strong>{" "}
            {workerStatus.isInitialized ? "Ready" : "Not Ready"}
          </div>
        </div>

        {isAuthenticated && (
          <div className="mt-4 space-x-2">
            <button
              onClick={handleLogout}
              className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
            >
              Logout
            </button>
            <button
              onClick={manualRefresh}
              className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
            >
              Refresh Tokens
            </button>
            <button
              onClick={forceTokenCheck}
              className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
            >
              Force Token Check
            </button>
            <button
              onClick={loadUserProfile}
              className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
            >
              Load Profile
            </button>
            <button
              onClick={testTokenDecryption}
              className="bg-purple-500 text-white px-4 py-2 rounded hover:bg-purple-600"
            >
              Test Token Decryption
            </button>
          </div>
        )}
      </div>

      {/* Token Health */}
      <div className="bg-white p-6 rounded-lg shadow">
        <h2 className="text-xl font-semibold mb-4">Token Health</h2>
        <div className="space-y-2">
          <div>
            <strong>Status:</strong> {tokenHealth.status}
          </div>
          <div>
            <strong>Can Refresh:</strong>{" "}
            {tokenHealth.canRefresh ? "Yes" : "No"}
          </div>
          <div>
            <strong>Needs Reauth:</strong>{" "}
            {tokenHealth.needsReauth ? "Yes" : "No"}
          </div>
          <div>
            <strong>Recommendations:</strong>
          </div>
          <ul className="list-disc list-inside ml-4">
            {tokenHealth.recommendations.map((rec, index) => (
              <li key={index}>{rec}</li>
            ))}
          </ul>
        </div>
      </div>

      {!isAuthenticated ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Login Form */}
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Login</h2>

            {error && (
              <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
                {error}
              </div>
            )}

            {step === "email" && (
              <form onSubmit={handleEmailSubmit}>
                <div className="mb-4">
                  <label className="block text-sm font-medium mb-2">
                    Email
                  </label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full p-2 border rounded"
                    required
                  />
                </div>
                <button
                  type="submit"
                  disabled={loading}
                  className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600 disabled:opacity-50"
                >
                  {loading ? "Sending..." : "Request Login Code"}
                </button>
              </form>
            )}

            {step === "ott" && (
              <form onSubmit={handleOttSubmit}>
                <div className="mb-4">
                  <label className="block text-sm font-medium mb-2">
                    Verification Code
                  </label>
                  <input
                    type="text"
                    value={ott}
                    onChange={(e) => setOtt(e.target.value)}
                    className="w-full p-2 border rounded"
                    placeholder="Enter 6-digit code"
                    required
                  />
                </div>
                <button
                  type="submit"
                  disabled={loading}
                  className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600 disabled:opacity-50"
                >
                  {loading ? "Verifying..." : "Verify Code"}
                </button>
                <button
                  type="button"
                  onClick={() => setStep("email")}
                  className="w-full mt-2 bg-gray-500 text-white p-2 rounded hover:bg-gray-600"
                >
                  Back to Email
                </button>
              </form>
            )}

            {step === "password" && (
              <form onSubmit={handlePasswordSubmit}>
                <div className="mb-4">
                  <label className="block text-sm font-medium mb-2">
                    Password
                  </label>
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full p-2 border rounded"
                    required
                  />
                </div>
                <button
                  type="submit"
                  disabled={loading}
                  className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600 disabled:opacity-50"
                >
                  {loading ? "Logging in..." : "Complete Login"}
                </button>
                <button
                  type="button"
                  onClick={() => setStep("ott")}
                  className="w-full mt-2 bg-gray-500 text-white p-2 rounded hover:bg-gray-600"
                >
                  Back to Code
                </button>
              </form>
            )}
          </div>

          {/* Registration Form */}
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Register</h2>

            <form onSubmit={handleRegistration}>
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium mb-2">
                      First Name
                    </label>
                    <input
                      type="text"
                      value={registrationData.firstName}
                      onChange={(e) =>
                        setRegistrationData({
                          ...registrationData,
                          firstName: e.target.value,
                        })
                      }
                      className="w-full p-2 border rounded"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Last Name
                    </label>
                    <input
                      type="text"
                      value={registrationData.lastName}
                      onChange={(e) =>
                        setRegistrationData({
                          ...registrationData,
                          lastName: e.target.value,
                        })
                      }
                      className="w-full p-2 border rounded"
                      required
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Email
                  </label>
                  <input
                    type="email"
                    value={registrationData.email}
                    onChange={(e) =>
                      setRegistrationData({
                        ...registrationData,
                        email: e.target.value,
                      })
                    }
                    className="w-full p-2 border rounded"
                    required
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Phone
                  </label>
                  <input
                    type="tel"
                    value={registrationData.phone}
                    onChange={(e) =>
                      setRegistrationData({
                        ...registrationData,
                        phone: e.target.value,
                      })
                    }
                    className="w-full p-2 border rounded"
                    required
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Country
                  </label>
                  <input
                    type="text"
                    value={registrationData.country}
                    onChange={(e) =>
                      setRegistrationData({
                        ...registrationData,
                        country: e.target.value,
                      })
                    }
                    className="w-full p-2 border rounded"
                    required
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">
                    Password
                  </label>
                  <input
                    type="password"
                    value={registrationData.password}
                    onChange={(e) =>
                      setRegistrationData({
                        ...registrationData,
                        password: e.target.value,
                      })
                    }
                    className="w-full p-2 border rounded"
                    required
                  />
                </div>

                <div className="flex items-center">
                  <input
                    type="checkbox"
                    checked={registrationData.agreeTerms}
                    onChange={(e) =>
                      setRegistrationData({
                        ...registrationData,
                        agreeTerms: e.target.checked,
                      })
                    }
                    className="mr-2"
                    required
                  />
                  <label className="text-sm">
                    I agree to the terms of service
                  </label>
                </div>
              </div>

              <button
                type="submit"
                disabled={loading || !registrationData.agreeTerms}
                className="w-full mt-4 bg-green-500 text-white p-2 rounded hover:bg-green-600 disabled:opacity-50"
              >
                {loading ? "Registering..." : "Register"}
              </button>
            </form>
          </div>
        </div>
      ) : (
        /* Authenticated User Panel */
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">User Profile</h2>
          {userProfile ? (
            <div className="space-y-2">
              <div>
                <strong>Email:</strong> {userProfile.email}
              </div>
              <div>
                <strong>Name:</strong> {userProfile.first_name}{" "}
                {userProfile.last_name}
              </div>
              <div>
                <strong>Role:</strong>{" "}
                {userProfile.role || userProfile.user_role}
              </div>
              <div>
                <strong>Created:</strong> {userProfile.created_at}
              </div>
            </div>
          ) : (
            <p>Click "Load Profile" to fetch user data</p>
          )}
        </div>
      )}

      {/* Debug Information */}
      <details className="bg-gray-100 p-4 rounded-lg">
        <summary className="font-semibold cursor-pointer">
          Debug Information
        </summary>
        <pre className="mt-4 text-sm overflow-auto bg-gray-50 p-4 rounded">
          {JSON.stringify(debugInfo, null, 2)}
        </pre>
      </details>
    </div>
  );
};

export default AuthExample;
