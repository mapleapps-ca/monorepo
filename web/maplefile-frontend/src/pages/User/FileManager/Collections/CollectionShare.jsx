// File: src/pages/User/FileManager/Collections/CollectionShare.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import { useFiles, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  ArrowLeftIcon,
  ShareIcon,
  UsersIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  ChevronRightIcon,
  HomeIcon,
  ExclamationTriangleIcon,
  CheckIcon,
  TrashIcon,
  PencilIcon,
  EyeIcon,
  UserPlusIcon,
  ClockIcon,
  InformationCircleIcon,
  XMarkIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";

// Permission levels constant
const PERMISSION_LEVELS = {
  READ_ONLY: "read_only",
  READ_WRITE: "read_write",
  ADMIN: "admin",
};

const CollectionShare = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();
  const { shareCollectionManager, getCollectionManager } = useFiles();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [collection, setCollection] = useState(null);
  const [collectionMembers, setCollectionMembers] = useState([]);
  const [sharedCollections, setSharedCollections] = useState([]);

  // Form state
  const [recipientId, setRecipientId] = useState("");
  const [recipientEmail, setRecipientEmail] = useState("");
  const [permissionLevel, setPermissionLevel] = useState(
    PERMISSION_LEVELS.READ_WRITE,
  );
  const [shareWithDescendants, setShareWithDescendants] = useState(true);
  const [password, setPassword] = useState("");

  // UI state
  const [showAdvanced, setShowAdvanced] = useState(false);

  // Load collection and sharing data
  useEffect(() => {
    if (collectionId && getCollectionManager && shareCollectionManager) {
      loadCollectionData();
      loadSharingData();
    }
  }, [collectionId, getCollectionManager, shareCollectionManager]);

  // Load collection details
  const loadCollectionData = async () => {
    try {
      const result = await getCollectionManager.getCollection(collectionId);
      if (result.collection) {
        setCollection(result.collection);
      }
    } catch (err) {
      console.error("Failed to load collection:", err);
      setError("Failed to load collection details");
    }
  };

  // Load sharing data
  const loadSharingData = async () => {
    try {
      // Load shared collections
      const shared = shareCollectionManager.getSharedCollections();
      setSharedCollections(Array.isArray(shared) ? shared : []);

      // Load collection members
      if (collectionId) {
        await loadCollectionMembers();
      }
    } catch (err) {
      console.error("Failed to load sharing data:", err);
    }
  };

  // Load collection members
  const loadCollectionMembers = async (forceRefresh = false) => {
    setIsLoading(true);
    try {
      const members = await shareCollectionManager.getCollectionMembers(
        collectionId,
        forceRefresh,
      );
      setCollectionMembers(Array.isArray(members) ? members : []);
    } catch (err) {
      console.error("Failed to load collection members:", err);
      setCollectionMembers([]);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle user lookup
  const handleVerifyRecipient = async () => {
    if (!recipientEmail.trim()) {
      setError("Enter recipient email first");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      const userLookupManager =
        shareCollectionManager._userLookupManager ||
        (await shareCollectionManager.getUserLookupManager());

      const userInfo = await userLookupManager.getUserPublicKey(
        recipientEmail.trim(),
      );

      if (userInfo.userId) {
        setRecipientId(userInfo.userId);
        setSuccess(`✅ User found: ${userInfo.name || userInfo.email}`);
      }
    } catch (err) {
      setError(`User lookup failed: ${err.message}`);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle collection sharing
  const handleShareCollection = async () => {
    if (!recipientId.trim() || !recipientEmail.trim()) {
      setError("Recipient ID and email are required");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      const shareData = {
        recipient_id: recipientId.trim(),
        recipient_email: recipientEmail.trim(),
        permission_level: permissionLevel,
        share_with_descendants: shareWithDescendants,
      };

      await shareCollectionManager.shareCollection(
        collectionId,
        shareData,
        password || null,
      );

      setSuccess(`Collection shared successfully with ${recipientEmail}`);

      // Clear form
      setRecipientId("");
      setRecipientEmail("");
      setPassword("");

      // Reload data
      await loadSharingData();
      await loadCollectionMembers(true);
    } catch (err) {
      console.error("Collection sharing failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle member removal
  const handleRemoveMember = async (memberId, memberEmail) => {
    if (!confirm(`Remove ${memberEmail} from this collection?`)) return;

    setIsLoading(true);
    setError("");

    try {
      await shareCollectionManager.removeMember(
        collectionId,
        memberId,
        shareWithDescendants,
      );

      setSuccess(`${memberEmail} removed from collection`);
      await loadCollectionMembers(true);
      await loadSharingData();
    } catch (err) {
      console.error("Member removal failed:", err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  // Get password from storage
  const handleGetStoredPassword = async () => {
    try {
      const storedPassword = await shareCollectionManager.getUserPassword();
      if (storedPassword) {
        setPassword(storedPassword);
        setSuccess("Password loaded from storage");
      } else {
        setError("No password found in storage");
      }
    } catch (err) {
      setError(`Failed to get stored password: ${err.message}`);
    }
  };

  // Clear messages
  const clearMessages = () => {
    setError("");
    setSuccess("");
  };

  // Auto-clear messages
  useEffect(() => {
    if (success || error) {
      const timer = setTimeout(clearMessages, 5000);
      return () => clearTimeout(timer);
    }
  }, [success, error]);

  // Get permission badge style
  const getPermissionBadge = (permission) => {
    switch (permission) {
      case PERMISSION_LEVELS.READ_ONLY:
        return "bg-blue-100 text-blue-800";
      case PERMISSION_LEVELS.READ_WRITE:
        return "bg-green-100 text-green-800";
      case PERMISSION_LEVELS.ADMIN:
        return "bg-red-100 text-red-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // Get permission icon
  const getPermissionIcon = (permission) => {
    switch (permission) {
      case PERMISSION_LEVELS.READ_ONLY:
        return <EyeIcon className="h-4 w-4" />;
      case PERMISSION_LEVELS.READ_WRITE:
        return <PencilIcon className="h-4 w-4" />;
      case PERMISSION_LEVELS.ADMIN:
        return <ShieldCheckIcon className="h-4 w-4" />;
      default:
        return <EyeIcon className="h-4 w-4" />;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Breadcrumb */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 mb-6">
          <HomeIcon className="h-4 w-4" />
          <ChevronRightIcon className="h-3 w-3" />
          <button
            onClick={() => navigate("/file-manager")}
            className="hover:text-gray-900"
          >
            My Files
          </button>
          <ChevronRightIcon className="h-3 w-3" />
          <button
            onClick={() =>
              navigate(`/file-manager/collections/${collectionId}`)
            }
            className="hover:text-gray-900"
          >
            {collection?.name || "Collection"}
          </button>
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900">Share</span>
        </div>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center space-x-4">
            <button
              onClick={() =>
                navigate(`/file-manager/collections/${collectionId}`)
              }
              className="p-2 text-gray-400 hover:text-gray-600 transition-colors duration-200"
            >
              <ArrowLeftIcon className="h-5 w-5" />
            </button>
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">
                Share Collection
              </h1>
              <p className="text-gray-600">
                {collection?.name
                  ? `Sharing "${collection.name}"`
                  : "Managing collection access"}
              </p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <div className="flex items-center space-x-1 px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm">
              <LockClosedIcon className="h-4 w-4" />
              <span>End-to-End Encrypted</span>
            </div>
          </div>
        </div>

        {/* Error/Success Messages */}
        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200">
            <div className="flex items-center">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
              <div>
                <h3 className="text-sm font-medium text-red-800">Error</h3>
                <p className="text-sm text-red-700 mt-1">{error}</p>
              </div>
              <button
                onClick={clearMessages}
                className="ml-auto text-red-500 hover:text-red-700"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}

        {success && (
          <div className="mb-6 p-4 rounded-lg bg-green-50 border border-green-200">
            <div className="flex items-center">
              <CheckIcon className="h-5 w-5 text-green-500 mr-3 flex-shrink-0" />
              <div>
                <h3 className="text-sm font-medium text-green-800">Success</h3>
                <p className="text-sm text-green-700 mt-1">{success}</p>
              </div>
              <button
                onClick={clearMessages}
                className="ml-auto text-green-500 hover:text-green-700"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}

        {/* Share Form */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex items-center space-x-3 mb-6">
            <div className="flex items-center justify-center h-10 w-10 bg-red-100 rounded-lg">
              <UserPlusIcon className="h-5 w-5 text-red-600" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900">
                Add People
              </h2>
              <p className="text-sm text-gray-600">
                Grant access to this encrypted collection
              </p>
            </div>
          </div>

          <div className="space-y-6">
            {/* Recipient Email */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Email Address
              </label>
              <div className="flex space-x-3">
                <input
                  type="email"
                  value={recipientEmail}
                  onChange={(e) => setRecipientEmail(e.target.value)}
                  placeholder="Enter recipient's email address"
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                />
                <button
                  onClick={handleVerifyRecipient}
                  disabled={isLoading || !recipientEmail.trim()}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors duration-200"
                >
                  {isLoading ? "Verifying..." : "Verify"}
                </button>
              </div>
              <p className="text-xs text-gray-500 mt-1">
                We'll verify this user exists and can receive encrypted data
              </p>
            </div>

            {/* Permission Level */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Permission Level
              </label>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                {[
                  {
                    value: PERMISSION_LEVELS.READ_ONLY,
                    label: "View Only",
                    desc: "Can view files but not edit",
                  },
                  {
                    value: PERMISSION_LEVELS.READ_WRITE,
                    label: "Can Edit",
                    desc: "Can view and modify files",
                  },
                  {
                    value: PERMISSION_LEVELS.ADMIN,
                    label: "Full Access",
                    desc: "Can manage sharing and settings",
                  },
                ].map((option) => (
                  <label
                    key={option.value}
                    className={`relative flex cursor-pointer rounded-lg border p-4 focus:outline-none transition-all duration-200 ${
                      permissionLevel === option.value
                        ? "border-red-600 bg-red-50 ring-2 ring-red-600"
                        : "border-gray-300 hover:border-gray-400"
                    }`}
                  >
                    <input
                      type="radio"
                      name="permission"
                      value={option.value}
                      checked={permissionLevel === option.value}
                      onChange={(e) => setPermissionLevel(e.target.value)}
                      className="sr-only"
                    />
                    <div className="flex flex-col">
                      <div className="flex items-center">
                        {getPermissionIcon(option.value)}
                        <span className="ml-2 font-medium text-gray-900">
                          {option.label}
                        </span>
                      </div>
                      <span className="mt-1 text-xs text-gray-500">
                        {option.desc}
                      </span>
                    </div>
                    {permissionLevel === option.value && (
                      <CheckIcon className="absolute right-3 top-3 h-5 w-5 text-red-600" />
                    )}
                  </label>
                ))}
              </div>
            </div>

            {/* Advanced Options */}
            <div>
              <button
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="flex items-center text-sm text-gray-600 hover:text-gray-900"
              >
                <span>Advanced Options</span>
                <ChevronRightIcon
                  className={`h-4 w-4 ml-1 transition-transform duration-200 ${showAdvanced ? "rotate-90" : ""}`}
                />
              </button>

              {showAdvanced && (
                <div className="mt-4 space-y-4 p-4 bg-gray-50 rounded-lg">
                  {/* Share with descendants */}
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={shareWithDescendants}
                      onChange={(e) =>
                        setShareWithDescendants(e.target.checked)
                      }
                      className="h-4 w-4 text-red-600 rounded border-gray-300 focus:ring-red-500"
                    />
                    <span className="ml-2 text-sm text-gray-700">
                      Also share sub-collections (folders within this
                      collection)
                    </span>
                  </label>

                  {/* Password field */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Encryption Password
                    </label>
                    <div className="flex space-x-2">
                      <input
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        placeholder="Leave empty to use stored password"
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                      />
                      <button
                        onClick={handleGetStoredPassword}
                        className="px-3 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors duration-200"
                      >
                        Use Stored
                      </button>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">
                      Used to encrypt the collection key for the recipient
                    </p>
                  </div>
                </div>
              )}
            </div>

            {/* Share Button */}
            <div className="flex justify-end">
              <button
                onClick={handleShareCollection}
                disabled={
                  isLoading || !recipientId.trim() || !recipientEmail.trim()
                }
                className="inline-flex items-center px-6 py-3 border border-transparent rounded-lg shadow-sm text-base font-medium text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed transition-all duration-200"
              >
                {isLoading ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    Sharing...
                  </>
                ) : (
                  <>
                    <ShareIcon className="h-4 w-4 mr-2" />
                    Share Collection
                  </>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Current Members */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center space-x-3">
              <div className="flex items-center justify-center h-10 w-10 bg-blue-100 rounded-lg">
                <UsersIcon className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-gray-900">
                  Current Access
                </h2>
                <p className="text-sm text-gray-600">
                  {collectionMembers.length}{" "}
                  {collectionMembers.length === 1
                    ? "person has"
                    : "people have"}{" "}
                  access
                </p>
              </div>
            </div>
            <button
              onClick={() => loadCollectionMembers(true)}
              disabled={isLoading}
              className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
            >
              <ArrowPathIcon className="h-4 w-4 mr-2" />
              {isLoading ? "Refreshing..." : "Refresh"}
            </button>
          </div>

          {collectionMembers.length === 0 ? (
            <div className="text-center py-8">
              <UsersIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <h3 className="text-sm font-medium text-gray-900 mb-2">
                No shared access yet
              </h3>
              <p className="text-sm text-gray-500">
                Share this collection with others to see them listed here
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {collectionMembers.map((member, index) => (
                <div
                  key={`${member.recipient_id}-${index}`}
                  className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors duration-200"
                >
                  <div className="flex items-center space-x-4">
                    <div className="flex items-center justify-center h-10 w-10 bg-gray-100 rounded-full">
                      <span className="text-sm font-medium text-gray-600">
                        {member.recipient_email
                          ? member.recipient_email.charAt(0).toUpperCase()
                          : "?"}
                      </span>
                    </div>
                    <div>
                      <p className="font-medium text-gray-900">
                        {member.recipient_email}
                      </p>
                      <p className="text-sm text-gray-500">
                        ID: {member.recipient_id}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-3">
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getPermissionBadge(member.permission_level)}`}
                    >
                      {getPermissionIcon(member.permission_level)}
                      <span className="ml-1 capitalize">
                        {member.permission_level.replace("_", " ")}
                      </span>
                    </span>
                    <button
                      onClick={() =>
                        handleRemoveMember(
                          member.recipient_id,
                          member.recipient_email,
                        )
                      }
                      disabled={isLoading}
                      className="text-gray-400 hover:text-red-600 transition-colors duration-200 disabled:opacity-50"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Security Info */}
        <div className="mt-6 bg-gradient-to-r from-green-50 to-blue-50 rounded-lg border border-green-100 p-4">
          <div className="flex items-start space-x-3">
            <InformationCircleIcon className="h-5 w-5 text-blue-600 mt-0.5" />
            <div>
              <h3 className="text-sm font-semibold text-blue-900 mb-2">
                How Sharing Works
              </h3>
              <ul className="text-sm text-blue-800 space-y-1">
                <li>
                  • Collection keys are encrypted with each recipient's public
                  key
                </li>
                <li>
                  • Recipients can only decrypt files if they have access to the
                  collection
                </li>
                <li>• All sharing operations use end-to-end encryption</li>
                <li>• You can revoke access at any time by removing members</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(CollectionShare);
