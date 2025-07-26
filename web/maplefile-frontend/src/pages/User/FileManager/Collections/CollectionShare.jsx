// File: src/pages/User/FileManager/Collections/CollectionShare.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import {
  useFiles,
  useAuth,
  useUsers,
  useCollections,
} from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  ArrowLeftIcon,
  ShareIcon,
  UserPlusIcon,
  TrashIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  MagnifyingGlassIcon,
} from "@heroicons/react/24/outline";

const CollectionShare = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();
  const { getCollectionManager } = useFiles();
  const { shareCollectionManager } = useCollections();
  const { authManager } = useAuth();
  const { userLookupManager } = useUsers();

  const [isLoading, setIsLoading] = useState(false);
  const [isVerifying, setIsVerifying] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [collection, setCollection] = useState(null);
  const [collectionMembers, setCollectionMembers] = useState([]);
  const [recipientEmail, setRecipientEmail] = useState("");
  const [recipientId, setRecipientId] = useState("");
  const [recipientVerified, setRecipientVerified] = useState(false);
  const [recipientInfo, setRecipientInfo] = useState(null);
  const [permissionLevel, setPermissionLevel] = useState("read_write");

  useEffect(() => {
    if (collectionId && getCollectionManager && shareCollectionManager) {
      loadCollectionData();
      loadCollectionMembers();
    }
  }, [collectionId, getCollectionManager, shareCollectionManager]);

  // Clear messages after 5 seconds
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(() => {
        setError("");
        setSuccess("");
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [success, error]);

  const loadCollectionData = async () => {
    try {
      const result = await getCollectionManager.getCollection(collectionId);
      if (result.collection) {
        setCollection(result.collection);
      }
    } catch (err) {
      console.error("Failed to load collection:", err);
      setError("Could not load folder details");
    }
  };

  const loadCollectionMembers = async (forceRefresh = false) => {
    setIsLoading(true);
    try {
      console.log(
        `[CollectionShare] Loading collection members for ${collectionId}, forceRefresh: ${forceRefresh}`,
      );

      const members = await shareCollectionManager.getCollectionMembers(
        collectionId,
        forceRefresh,
      );

      console.log(
        `[CollectionShare] API returned ${members.length} members:`,
        members,
      );

      // Validate the members array
      if (!Array.isArray(members)) {
        console.warn("[CollectionShare] Members is not an array:", members);
        setCollectionMembers([]);
        return;
      }

      setCollectionMembers(members);

      // If no members but we know we just shared, show a warning
      if (members.length === 0 && success) {
        console.warn(
          "[CollectionShare] No members returned from API despite successful sharing",
        );
        setError(
          "Collection shared successfully, but member list may take a moment to update. Try refreshing in a few seconds.",
        );
      }
    } catch (err) {
      console.error("[CollectionShare] Failed to load members:", err);
      setError("Could not load folder members: " + err.message);

      // Try to get from local storage as fallback
      try {
        const localShares =
          shareCollectionManager.getSharedCollectionsByCollectionId(
            collectionId,
          );
        console.log("[CollectionShare] Fallback - local shares:", localShares);

        if (localShares && localShares.length > 0) {
          // Convert local shares to member format
          const membersFromLocal = localShares.map((share) => ({
            recipient_id: share.recipient_id,
            recipient_email: share.recipient_email,
            permission_level: share.permission_level,
            shared_at: share.shared_at || share.locally_stored_at,
          }));
          setCollectionMembers(membersFromLocal);
          console.log("[CollectionShare] Using local shares as member list");
        }
      } catch (localErr) {
        console.error(
          "[CollectionShare] Failed to get local shares:",
          localErr,
        );
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerifyRecipient = async () => {
    if (!recipientEmail.trim()) {
      setError("Please enter an email address");
      return;
    }

    // Check if user is trying to share with themselves
    const currentUserEmail = authManager.getCurrentUserEmail();
    if (
      recipientEmail.trim().toLowerCase() === currentUserEmail?.toLowerCase()
    ) {
      setError("You cannot share a folder with yourself");
      return;
    }

    // Reset verification state
    setRecipientVerified(false);
    setRecipientId("");
    setRecipientInfo(null);
    setError("");
    setSuccess("");
    setIsVerifying(true);

    try {
      // Use userLookupManager to get user public key and info
      const userInfo = await userLookupManager.getUserPublicKey(
        recipientEmail.trim(),
      );

      if (userInfo && userInfo.userId) {
        setRecipientId(userInfo.userId);
        setRecipientInfo(userInfo);
        setRecipientVerified(true);
        setSuccess(`✅ User verified: ${userInfo.name || userInfo.email}`);
      } else {
        throw new Error("User information incomplete");
      }
    } catch (err) {
      console.error("User verification failed:", err);
      setError("User not found. Please check the email address and try again.");
      setRecipientVerified(false);
      setRecipientId("");
      setRecipientInfo(null);
    } finally {
      setIsVerifying(false);
    }
  };

  const handleShareCollection = async () => {
    if (!recipientVerified || !recipientId || !recipientEmail) {
      setError("Please verify the recipient first");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      await shareCollectionManager.shareCollection(
        collectionId,
        {
          recipient_id: recipientId,
          recipient_email: recipientEmail.trim(),
          permission_level: permissionLevel,
          share_with_descendants: true,
        },
        null, // Password will be retrieved automatically from PasswordStorageService
      );

      setSuccess(`Folder shared successfully with ${recipientEmail}!`);

      // Immediately add the new member to the local list (optimistic update)
      const newMember = {
        recipient_id: recipientId,
        recipient_email: recipientEmail,
        permission_level: permissionLevel,
        shared_at: new Date().toISOString(),
        _isLocal: true, // Mark as locally added
      };

      setCollectionMembers((prevMembers) => {
        // Remove any existing member with same recipient_id to avoid duplicates
        const filtered = prevMembers.filter(
          (m) => m.recipient_id !== recipientId,
        );
        return [...filtered, newMember];
      });

      // Reset form
      setRecipientEmail("");
      setRecipientId("");
      setRecipientVerified(false);
      setRecipientInfo(null);
      setPermissionLevel("read_write");

      // Also try to reload from API in the background (don't block UI)
      console.log(
        "[CollectionShare] Sharing successful, attempting background refresh...",
      );
      setTimeout(async () => {
        try {
          await loadCollectionMembers(true);
        } catch (bgErr) {
          console.warn(
            "[CollectionShare] Background refresh failed:",
            bgErr.message,
          );
          // Don't show error to user since we already have the optimistic update
        }
      }, 1000); // Wait 1 second for API to update
    } catch (err) {
      console.error("Failed to share collection:", err);
      setError("Could not share folder. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleRemoveMember = async (memberId, memberEmail) => {
    if (!confirm(`Remove ${memberEmail} from this folder?`)) return;

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      // Optimistically remove from UI immediately
      setCollectionMembers((prevMembers) =>
        prevMembers.filter((m) => m.recipient_id !== memberId),
      );

      await shareCollectionManager.removeMember(collectionId, memberId, true);
      setSuccess(`${memberEmail} removed from folder`);

      // Background refresh to ensure consistency
      console.log(
        "[CollectionShare] Member removal successful, attempting background refresh...",
      );
      setTimeout(async () => {
        try {
          await loadCollectionMembers(true);
        } catch (bgErr) {
          console.warn(
            "[CollectionShare] Background refresh after removal failed:",
            bgErr.message,
          );
        }
      }, 1000);
    } catch (err) {
      console.error("[CollectionShare] Failed to remove member:", err);
      setError("Could not remove member: " + err.message);

      // Revert the optimistic update on error
      await loadCollectionMembers(true);
    } finally {
      setIsLoading(false);
    }
  };

  const handleEmailChange = (e) => {
    const newEmail = e.target.value;
    setRecipientEmail(newEmail);

    // Reset verification when email changes
    if (recipientVerified) {
      setRecipientVerified(false);
      setRecipientId("");
      setRecipientInfo(null);
    }

    // Clear any existing errors when user starts typing
    if (error) {
      setError("");
    }

    // Check if user is trying to enter their own email
    const currentUserEmail = authManager.getCurrentUserEmail();
    if (
      newEmail.trim().toLowerCase() === currentUserEmail?.toLowerCase() &&
      newEmail.trim()
    ) {
      setError("You cannot share a folder with yourself");
    }
  };

  const getPermissionDisplayName = (level) => {
    switch (level) {
      case "read_only":
        return "View Only";
      case "read_write":
        return "Can Edit";
      case "admin":
        return "Full Access";
      default:
        return level;
    }
  };

  const getPermissionIcon = (level) => {
    switch (level) {
      case "read_only":
        return "👁️";
      case "read_write":
        return "✏️";
      case "admin":
        return "👑";
      default:
        return "📝";
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() =>
              navigate(`/file-manager/collections/${collectionId}`)
            }
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to {collection?.name || "Folder"}
          </button>
          <div className="flex items-center">
            <ShareIcon className="h-8 w-8 text-red-800 mr-3" />
            <h1 className="text-2xl font-semibold text-gray-900">
              Share Folder
            </h1>
          </div>
          {collection && (
            <p className="text-gray-600 mt-2">
              Sharing "{collection.name}" with end-to-end encryption
            </p>
          )}
        </div>

        {/* Messages */}
        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 flex items-start">
            <ExclamationTriangleIcon className="h-5 w-5 text-red-400 mt-0.5 mr-3 flex-shrink-0" />
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}
        {success && (
          <div className="mb-6 p-4 rounded-lg bg-green-50 border border-green-200 flex items-start">
            <CheckIcon className="h-5 w-5 text-green-400 mt-0.5 mr-3 flex-shrink-0" />
            <p className="text-sm text-green-700">{success}</p>
          </div>
        )}

        {/* Add People */}
        <div className="bg-white rounded-lg border border-gray-200 p-6 mb-6 shadow-sm">
          <h2 className="text-lg font-medium text-gray-900 mb-6 flex items-center">
            <UserPlusIcon className="h-5 w-5 mr-2" />
            Add People
          </h2>

          <div className="space-y-6">
            {/* Email Input with Verification */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Email Address
              </label>
              <div className="flex space-x-3">
                <div className="flex-1 relative">
                  <input
                    type="email"
                    value={recipientEmail}
                    onChange={handleEmailChange}
                    placeholder="Enter email address to share with"
                    className={`w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 ${
                      recipientVerified
                        ? "border-green-300 bg-green-50"
                        : authManager.getCurrentUserEmail()?.toLowerCase() ===
                              recipientEmail.trim().toLowerCase() &&
                            recipientEmail.trim()
                          ? "border-red-300 bg-red-50"
                          : "border-gray-300"
                    }`}
                    disabled={isVerifying || isLoading}
                  />
                  {recipientVerified && (
                    <div className="absolute right-2 top-2">
                      <CheckIcon className="h-5 w-5 text-green-500" />
                    </div>
                  )}
                </div>
                <button
                  onClick={handleVerifyRecipient}
                  disabled={
                    isVerifying ||
                    isLoading ||
                    !recipientEmail.trim() ||
                    authManager.getCurrentUserEmail()?.toLowerCase() ===
                      recipientEmail.trim().toLowerCase()
                  }
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center"
                >
                  {isVerifying ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                      Verifying...
                    </>
                  ) : (
                    <>
                      <MagnifyingGlassIcon className="h-4 w-4 mr-1" />
                      Verify
                    </>
                  )}
                </button>
              </div>
              {recipientVerified && recipientInfo && (
                <div className="mt-2 p-3 bg-green-50 border border-green-200 rounded-lg">
                  <p className="text-sm text-green-800">
                    <strong>✅ Verified:</strong>{" "}
                    {recipientInfo.name || recipientInfo.email}
                  </p>
                  <div className="text-xs text-green-600 mt-1 space-y-1">
                    <p>
                      <strong>User ID:</strong> {recipientInfo.userId}
                    </p>
                    <p>
                      <strong>Verification ID:</strong>{" "}
                      {recipientInfo.verificationId ||
                        recipientInfo.verification_id ||
                        "N/A"}
                    </p>
                  </div>
                </div>
              )}
              <p className="text-xs text-gray-500 mt-1">
                We'll verify this user exists before sharing. You cannot share
                with yourself.
              </p>
            </div>

            {/* Permission Level */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Permission Level
              </label>
              <select
                value={permissionLevel}
                onChange={(e) => setPermissionLevel(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                disabled={isLoading}
              >
                <option value="read_only">
                  👁️ View Only - Can only view files
                </option>
                <option value="read_write">
                  ✏️ Can Edit - Can add and edit files
                </option>
                <option value="admin">
                  👑 Full Access - Can manage sharing and permissions
                </option>
              </select>
              <p className="text-xs text-gray-500 mt-1">
                Choose what level of access this person should have
              </p>
            </div>

            {/* Share Button */}
            <button
              onClick={handleShareCollection}
              disabled={
                isLoading ||
                isVerifying ||
                !recipientVerified ||
                !recipientId ||
                !recipientEmail.trim()
              }
              className="w-full py-3 px-4 rounded-lg text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors font-medium flex items-center justify-center"
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Sharing Folder...
                </>
              ) : (
                <>
                  <ShareIcon className="h-4 w-4 mr-2" />
                  Share Folder Securely
                </>
              )}
            </button>
            {!recipientVerified && recipientEmail.trim() && (
              <p className="text-xs text-amber-600 text-center">
                Please verify the recipient before sharing
              </p>
            )}
          </div>
        </div>

        {/* Current Members */}
        <div className="bg-white rounded-lg border border-gray-200 p-6 shadow-sm">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-medium text-gray-900">
              People with Access ({collectionMembers.length})
            </h2>
            <button
              onClick={() => loadCollectionMembers(true)}
              disabled={isLoading}
              className="px-3 py-1 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg transition-colors disabled:opacity-50"
              title="Refresh member list"
            >
              {isLoading ? "Refreshing..." : "Refresh"}
            </button>
          </div>

          {isLoading && collectionMembers.length === 0 ? (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-800 mx-auto"></div>
              <p className="text-gray-500 mt-2">Loading members...</p>
            </div>
          ) : collectionMembers.length === 0 ? (
            <div className="text-center py-8 border-2 border-dashed border-gray-200 rounded-lg">
              <UserPlusIcon className="h-12 w-12 text-gray-400 mx-auto mb-3" />
              <p className="text-gray-500 text-lg font-medium">
                No one else has access yet
              </p>
              <p className="text-gray-400 text-sm">
                Share this folder with others to start collaborating
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {collectionMembers.map((member, index) => (
                <div
                  key={`${member.recipient_id}-${index}`}
                  className={`flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 transition-colors ${
                    member._isLocal
                      ? "border-green-200 bg-green-50"
                      : "border-gray-200"
                  }`}
                >
                  <div className="flex items-center space-x-3">
                    <div className="h-10 w-10 bg-red-100 rounded-full flex items-center justify-center">
                      <span className="text-red-800 font-medium">
                        {member.recipient_email.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <div className="flex items-center space-x-2">
                        <p className="font-medium text-gray-900">
                          {member.recipient_email}
                        </p>
                        {member._isLocal && (
                          <span className="text-xs px-2 py-1 bg-green-100 text-green-700 rounded-full">
                            Just shared
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-gray-500 flex items-center">
                        {getPermissionIcon(member.permission_level)}
                        <span className="ml-1">
                          {getPermissionDisplayName(member.permission_level)}
                        </span>
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={() =>
                      handleRemoveMember(
                        member.recipient_id,
                        member.recipient_email,
                      )
                    }
                    disabled={isLoading}
                    className="text-gray-400 hover:text-red-600 transition-colors p-2 rounded-lg hover:bg-red-50"
                    title="Remove access"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Debug Section for Development */}
        {import.meta.env.DEV && (
          <div className="mt-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <h3 className="text-sm font-medium text-yellow-800 mb-2">
              🔧 Debug Info (Dev Mode Only)
            </h3>
            <div className="text-xs text-yellow-700 space-y-1">
              <p>
                <strong>Collection ID:</strong> {collectionId}
              </p>
              <p>
                <strong>Members Array Length:</strong>{" "}
                {collectionMembers.length}
              </p>
              <p>
                <strong>Share Collection Manager:</strong>{" "}
                {shareCollectionManager ? "✅ Available" : "❌ Missing"}
              </p>
              <p>
                <strong>User Lookup Manager:</strong>{" "}
                {userLookupManager ? "✅ Available" : "❌ Missing"}
              </p>
              <p>
                <strong>Last Success Message:</strong> {success || "None"}
              </p>
              <p>
                <strong>Last Error Message:</strong> {error || "None"}
              </p>
              <button
                onClick={() => {
                  console.log("=== COLLECTION SHARE DEBUG ===");
                  console.log("Collection ID:", collectionId);
                  console.log("Collection Members:", collectionMembers);
                  console.log(
                    "Share Collection Manager:",
                    shareCollectionManager,
                  );
                  console.log("Auth Manager:", authManager);
                  console.log("User Lookup Manager:", userLookupManager);

                  // Try to get local shares
                  try {
                    const localShares =
                      shareCollectionManager.getSharedCollectionsByCollectionId(
                        collectionId,
                      );
                    console.log("Local Shares:", localShares);
                  } catch (e) {
                    console.error("Error getting local shares:", e);
                  }

                  console.log("=== END DEBUG ===");
                }}
                className="mt-2 px-2 py-1 bg-yellow-200 hover:bg-yellow-300 text-yellow-800 rounded text-xs"
              >
                Log Debug Info to Console
              </button>
            </div>
          </div>
        )}

        {/* Security Notice */}
        <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="flex items-start">
            <div className="h-5 w-5 text-blue-400 mt-0.5 mr-3 flex-shrink-0">
              🔒
            </div>
            <div>
              <h3 className="text-sm font-medium text-blue-800">
                End-to-End Encryption
              </h3>
              <p className="text-sm text-blue-700 mt-1">
                All files in this folder are encrypted with your keys. When you
                share, the folder key is securely encrypted with the recipient's
                public key. Neither we nor anyone else can access your files.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(CollectionShare);
