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
  EnvelopeIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  UserGroupIcon,
  PencilIcon,
  EyeIcon,
  ChevronDownIcon,
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
  const [showPermissionDropdown, setShowPermissionDropdown] = useState(false);

  const permissionLevels = [
    {
      value: "read_only",
      label: "View Only",
      description: "Can view and download files",
      icon: EyeIcon,
      color: "text-blue-600",
    },
    {
      value: "read_write",
      label: "Can Edit",
      description: "Can add, edit, and delete files",
      icon: PencilIcon,
      color: "text-green-600",
    },
    {
      value: "admin",
      label: "Full Access",
      description: "Can manage sharing and permissions",
      icon: ShieldCheckIcon,
      color: "text-purple-600",
    },
  ];

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
      const members = await shareCollectionManager.getCollectionMembers(
        collectionId,
        forceRefresh,
      );
      if (!Array.isArray(members)) {
        console.warn("[CollectionShare] Members is not an array:", members);
        setCollectionMembers([]);
        return;
      }
      setCollectionMembers(members);
      if (members.length === 0 && success) {
        setError(
          "Collection shared successfully, but member list may take a moment to update. Try refreshing in a few seconds.",
        );
      }
    } catch (err) {
      console.error("[CollectionShare] Failed to load members:", err);
      setError("Could not load folder members: " + err.message);
      try {
        const localShares =
          shareCollectionManager.getSharedCollectionsByCollectionId(
            collectionId,
          );
        if (localShares && localShares.length > 0) {
          const membersFromLocal = localShares.map((share) => ({
            recipient_id: share.recipient_id,
            recipient_email: share.recipient_email,
            permission_level: share.permission_level,
            shared_at: share.shared_at || share.locally_stored_at,
          }));
          setCollectionMembers(membersFromLocal);
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

    const currentUserEmail = authManager.getCurrentUserEmail();
    if (
      recipientEmail.trim().toLowerCase() === currentUserEmail?.toLowerCase()
    ) {
      setError("You cannot share a folder with yourself");
      return;
    }

    setRecipientVerified(false);
    setRecipientId("");
    setRecipientInfo(null);
    setError("");
    setSuccess("");
    setIsVerifying(true);

    try {
      const userInfo = await userLookupManager.getUserPublicKey(
        recipientEmail.trim(),
      );
      if (userInfo && userInfo.userId) {
        setRecipientId(userInfo.userId);
        setRecipientInfo(userInfo);
        setRecipientVerified(true);
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

      const newMember = {
        recipient_id: recipientId,
        recipient_email: recipientEmail.trim(),
        permission_level: permissionLevel,
        shared_at: new Date().toISOString(),
        _isLocal: true,
      };

      setCollectionMembers((prevMembers) => {
        const filtered = prevMembers.filter(
          (m) => m.recipient_id !== recipientId,
        );
        return [...filtered, newMember];
      });

      setRecipientEmail("");
      setRecipientId("");
      setRecipientVerified(false);
      setRecipientInfo(null);
      setPermissionLevel("read_write");

      setTimeout(() => {
        loadCollectionMembers(true);
      }, 1000);
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
      setCollectionMembers((prevMembers) =>
        prevMembers.filter((m) => m.recipient_id !== memberId),
      );

      await shareCollectionManager.removeMember(collectionId, memberId, true);
      setSuccess(`${memberEmail} removed from folder`);

      setTimeout(() => {
        loadCollectionMembers(true);
      }, 1000);
    } catch (err) {
      console.error("[CollectionShare] Failed to remove member:", err);
      setError("Could not remove member: " + err.message);
      loadCollectionMembers(true);
    } finally {
      setIsLoading(false);
    }
  };

  const handleEmailChange = (e) => {
    const newEmail = e.target.value;
    setRecipientEmail(newEmail);

    if (recipientVerified) {
      setRecipientVerified(false);
      setRecipientId("");
      setRecipientInfo(null);
      setSuccess("");
    }
    if (error) {
      setError("");
    }
    const currentUserEmail = authManager.getCurrentUserEmail();
    if (
      newEmail.trim().toLowerCase() === currentUserEmail?.toLowerCase() &&
      newEmail.trim()
    ) {
      setError("You cannot share a folder with yourself");
    }
  };

  const formatTimeAgo = (dateString) => {
    if (!dateString) return "";
    const date = new Date(dateString);
    const now = new Date();
    const diffInSeconds = Math.floor((now - date) / 1000);

    if (diffInSeconds < 60) return "Just now";
    const diffInMinutes = Math.floor(diffInSeconds / 60);
    if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) return `${diffInHours}h ago`;
    const diffInDays = Math.floor(diffInHours / 24);
    if (diffInDays === 1) return "Yesterday";
    if (diffInDays < 7) return `${diffInDays} days ago`;
    return date.toLocaleDateString();
  };

  const currentPermission = permissionLevels.find(
    (p) => p.value === permissionLevel,
  );
  const currentUserEmail = authManager?.getCurrentUserEmail();

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8 animate-fade-in-down">
          <button
            onClick={() =>
              navigate(`/file-manager/collections/${collectionId}`)
            }
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors duration-200"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to {collection?.name || "Folder"}
          </button>
          <div className="flex items-center">
            <div className="h-12 w-12 bg-gradient-to-br from-red-600 to-red-800 rounded-xl flex items-center justify-center mr-4">
              <ShareIcon className="h-6 w-6 text-white" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">Share Folder</h1>
              {collection && (
                <p className="text-gray-600 mt-1">
                  Share "{collection.name}" with end-to-end encryption
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Messages */}
        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 flex items-start animate-fade-in">
            <ExclamationTriangleIcon className="h-5 w-5 text-red-400 mt-0.5 mr-3 flex-shrink-0" />
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}
        {success && (
          <div className="mb-6 p-4 rounded-lg bg-green-50 border border-green-200 flex items-start animate-fade-in">
            <CheckIcon className="h-5 w-5 text-green-400 mt-0.5 mr-3 flex-shrink-0" />
            <p className="text-sm text-green-700">{success}</p>
          </div>
        )}

        {/* Add People Section */}
        <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 mb-6 animate-fade-in-up">
          <div className="flex items-center mb-6">
            <UserPlusIcon className="h-6 w-6 text-gray-700 mr-3" />
            <h2 className="text-lg font-semibold text-gray-900">Add People</h2>
          </div>

          <div className="space-y-4">
            {/* Email Input */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Email Address
              </label>
              <div className="flex space-x-3">
                <div className="flex-1 relative">
                  <EnvelopeIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                  <input
                    type="email"
                    value={recipientEmail}
                    onChange={handleEmailChange}
                    placeholder="Enter email address"
                    className={`w-full pl-10 pr-10 py-2 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-colors ${
                      recipientVerified
                        ? "border-green-300 bg-green-50"
                        : "border-gray-300"
                    }`}
                    disabled={isVerifying || isLoading}
                  />
                  {recipientVerified && (
                    <CheckIcon className="absolute right-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-green-600" />
                  )}
                </div>
                <button
                  onClick={handleVerifyRecipient}
                  disabled={
                    isVerifying ||
                    isLoading ||
                    !recipientEmail.trim() ||
                    recipientEmail.trim().toLowerCase() ===
                      currentUserEmail?.toLowerCase()
                  }
                  className="px-4 py-2 bg-gray-100 text-gray-800 rounded-lg hover:bg-gray-200 disabled:bg-gray-300 disabled:text-gray-500 disabled:cursor-not-allowed transition-colors flex items-center"
                >
                  {isVerifying ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2"></div>
                      Verifying...
                    </>
                  ) : (
                    <>
                      <MagnifyingGlassIcon className="h-4 w-4 mr-2" />
                      Verify
                    </>
                  )}
                </button>
              </div>
              {recipientVerified && (
                <div className="mt-2 p-3 bg-green-50 border border-green-200 rounded-lg animate-fade-in">
                  <p className="text-sm text-green-800 flex items-center">
                    <CheckIcon className="h-4 w-4 mr-2" />
                    User verified and ready to share
                  </p>
                </div>
              )}
            </div>

            {/* Permission Level */}
            {currentPermission && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Permission Level
                </label>
                <div className="relative">
                  <button
                    onClick={() =>
                      setShowPermissionDropdown(!showPermissionDropdown)
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 flex items-center justify-between text-left"
                    disabled={isLoading}
                  >
                    <div className="flex items-center">
                      <currentPermission.icon
                        className={`h-5 w-5 mr-3 ${currentPermission.color}`}
                      />
                      <div>
                        <p className="font-medium text-gray-900">
                          {currentPermission.label}
                        </p>
                        <p className="text-xs text-gray-500">
                          {currentPermission.description}
                        </p>
                      </div>
                    </div>
                    <ChevronDownIcon
                      className={`h-5 w-5 text-gray-400 transition-transform duration-200 ${
                        showPermissionDropdown ? "rotate-180" : ""
                      }`}
                    />
                  </button>

                  {showPermissionDropdown && (
                    <>
                      <div
                        className="fixed inset-0 z-10"
                        onClick={() => setShowPermissionDropdown(false)}
                      ></div>
                      <div className="absolute top-full mt-2 w-full bg-white rounded-lg shadow-xl border border-gray-200 py-2 z-20 animate-fade-in-down">
                        {permissionLevels.map((level) => (
                          <button
                            key={level.value}
                            onClick={() => {
                              setPermissionLevel(level.value);
                              setShowPermissionDropdown(false);
                            }}
                            className={`w-full px-4 py-3 flex items-center hover:bg-gray-50 transition-colors duration-200 ${
                              permissionLevel === level.value ? "bg-red-50" : ""
                            }`}
                          >
                            <level.icon
                              className={`h-5 w-5 mr-3 ${level.color}`}
                            />
                            <div className="text-left flex-1">
                              <p
                                className={`font-medium ${
                                  permissionLevel === level.value
                                    ? "text-red-900"
                                    : "text-gray-900"
                                }`}
                              >
                                {level.label}
                              </p>
                              <p className="text-xs text-gray-500">
                                {level.description}
                              </p>
                            </div>
                            {permissionLevel === level.value && (
                              <CheckIcon className="h-5 w-5 text-red-600" />
                            )}
                          </button>
                        ))}
                      </div>
                    </>
                  )}
                </div>
              </div>
            )}

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
              className="w-full py-3 px-4 rounded-lg text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors font-medium flex items-center justify-center text-base"
            >
              {isLoading && !isVerifying ? (
                <>
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2"></div>
                  Sharing Folder...
                </>
              ) : (
                <>
                  <ShareIcon className="h-5 w-5 mr-2" />
                  Share Folder Securely
                </>
              )}
            </button>
          </div>
        </div>

        {/* Current Members */}
        <div
          className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 animate-fade-in-up"
          style={{ animationDelay: "100ms" }}
        >
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center">
              <UserGroupIcon className="h-6 w-6 text-gray-700 mr-3" />
              <h2 className="text-lg font-semibold text-gray-900">
                People with Access ({collectionMembers.length + 1})
              </h2>
            </div>
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
          ) : (
            <div className="space-y-3">
              {/* Owner */}
              <div className="p-4 bg-gradient-to-r from-red-50 to-pink-50 border border-red-200 rounded-lg">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    <div className="h-10 w-10 bg-gradient-to-br from-red-600 to-red-800 rounded-full flex items-center justify-center text-white font-semibold">
                      YOU
                    </div>
                    <div>
                      <div className="flex items-center space-x-2">
                        <p className="font-medium text-gray-900">You</p>
                        <span className="text-xs font-medium px-2 py-0.5 bg-red-100 text-red-800 rounded-full">
                          Owner
                        </span>
                      </div>
                      <p className="text-sm text-gray-500">
                        {currentUserEmail}
                      </p>
                    </div>
                  </div>
                  <div className="text-sm text-gray-500">Full control</div>
                </div>
              </div>

              {collectionMembers.length > 0
                ? collectionMembers.map((member) => {
                    const memberPermission =
                      permissionLevels.find(
                        (p) => p.value === member.permission_level,
                      ) || permissionLevels[0];
                    return (
                      <div
                        key={member.recipient_id}
                        className={`p-4 bg-white border rounded-lg hover:shadow-sm transition-all duration-200 ${
                          member._isLocal
                            ? "border-green-300 bg-green-50"
                            : "border-gray-200"
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center space-x-3">
                            <div className="h-10 w-10 bg-gray-200 rounded-full flex items-center justify-center text-gray-700 font-semibold">
                              {member.recipient_email.charAt(0).toUpperCase()}
                            </div>
                            <div>
                              <div className="flex items-center space-x-2">
                                <p className="font-medium text-gray-900">
                                  {member.recipient_email}
                                </p>
                                {member._isLocal && (
                                  <span className="text-xs font-medium px-2 py-0.5 bg-green-100 text-green-800 rounded-full">
                                    Just added
                                  </span>
                                )}
                              </div>
                              <p className="text-sm text-gray-500">
                                Shared {formatTimeAgo(member.shared_at)}
                              </p>
                            </div>
                          </div>
                          <div className="flex items-center space-x-3">
                            <div className="text-right">
                              <div className="flex items-center justify-end text-sm">
                                <memberPermission.icon
                                  className={`h-4 w-4 mr-1 ${memberPermission.color}`}
                                />
                                <span className="text-gray-700">
                                  {memberPermission.label}
                                </span>
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
                              className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all duration-200"
                              title="Remove access"
                            >
                              <TrashIcon className="h-4 w-4" />
                            </button>
                          </div>
                        </div>
                      </div>
                    );
                  })
                : !isLoading && (
                    <div className="text-center py-8 border-2 border-dashed border-gray-200 rounded-lg">
                      <UserGroupIcon className="h-12 w-12 text-gray-400 mx-auto mb-3" />
                      <p className="text-gray-500 text-lg font-medium">
                        No one else has access yet
                      </p>
                      <p className="text-gray-400 text-sm">
                        Share this folder to start collaborating
                      </p>
                    </div>
                  )}
            </div>
          )}
        </div>

        {/* Security Notice */}
        <div
          className="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200 animate-fade-in-up"
          style={{ animationDelay: "200ms" }}
        >
          <div className="flex">
            <LockClosedIcon className="h-5 w-5 text-blue-600 flex-shrink-0" />
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">
                End-to-End Encryption
              </h3>
              <p className="text-sm text-blue-700 mt-1">
                When you share this folder, the encryption key is securely
                shared with the recipient using their public key. Only they can
                decrypt and access the files.
              </p>
            </div>
          </div>
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
      </div>
    </div>
  );
};

export default withPasswordProtection(CollectionShare);
