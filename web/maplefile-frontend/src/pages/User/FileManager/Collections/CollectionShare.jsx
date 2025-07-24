// File: src/pages/User/FileManager/Collections/CollectionShare.jsx
import React, { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import { useFiles, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  ArrowLeftIcon,
  ShareIcon,
  UserPlusIcon,
  TrashIcon,
  CheckIcon,
} from "@heroicons/react/24/outline";

const CollectionShare = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();
  const { shareCollectionManager, getCollectionManager } = useFiles();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [collection, setCollection] = useState(null);
  const [collectionMembers, setCollectionMembers] = useState([]);
  const [recipientEmail, setRecipientEmail] = useState("");
  const [recipientId, setRecipientId] = useState("");
  const [permissionLevel, setPermissionLevel] = useState("read_write");

  useEffect(() => {
    if (collectionId && getCollectionManager && shareCollectionManager) {
      loadCollectionData();
      loadCollectionMembers();
    }
  }, [collectionId, getCollectionManager, shareCollectionManager]);

  const loadCollectionData = async () => {
    try {
      const result = await getCollectionManager.getCollection(collectionId);
      if (result.collection) {
        setCollection(result.collection);
      }
    } catch (err) {
      setError("Could not load folder details");
    }
  };

  const loadCollectionMembers = async () => {
    setIsLoading(true);
    try {
      const members = await shareCollectionManager.getCollectionMembers(
        collectionId,
        false,
      );
      setCollectionMembers(Array.isArray(members) ? members : []);
    } catch (err) {
      console.error("Failed to load members:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleVerifyRecipient = async () => {
    if (!recipientEmail.trim()) {
      setError("Please enter an email address");
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
        setSuccess("User found! You can now share with them.");
      }
    } catch (err) {
      setError("Could not find user. Please check the email.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleShareCollection = async () => {
    if (!recipientId || !recipientEmail) {
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
          recipient_email: recipientEmail,
          permission_level: permissionLevel,
          share_with_descendants: true,
        },
        null,
      );

      setSuccess("Folder shared successfully!");
      setRecipientEmail("");
      setRecipientId("");
      await loadCollectionMembers();
    } catch (err) {
      setError("Could not share folder. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleRemoveMember = async (memberId, memberEmail) => {
    if (!confirm(`Remove ${memberEmail} from this folder?`)) return;

    setIsLoading(true);
    try {
      await shareCollectionManager.removeMember(collectionId, memberId, true);
      setSuccess("Member removed");
      await loadCollectionMembers();
    } catch (err) {
      setError("Could not remove member");
    } finally {
      setIsLoading(false);
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
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to {collection?.name || "Folder"}
          </button>
          <h1 className="text-2xl font-semibold text-gray-900">Share Folder</h1>
        </div>

        {/* Messages */}
        {error && (
          <div className="mb-6 p-3 rounded-lg bg-red-50 border border-red-200">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}
        {success && (
          <div className="mb-6 p-3 rounded-lg bg-green-50 border border-green-200">
            <p className="text-sm text-green-700">{success}</p>
          </div>
        )}

        {/* Add People */}
        <div className="bg-white rounded-lg border border-gray-200 p-6 mb-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
            <UserPlusIcon className="h-5 w-5 mr-2" />
            Add People
          </h2>

          <div className="space-y-4">
            {/* Email Input */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Email Address
              </label>
              <div className="flex space-x-3">
                <input
                  type="email"
                  value={recipientEmail}
                  onChange={(e) => setRecipientEmail(e.target.value)}
                  placeholder="Enter email address"
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
                />
                <button
                  onClick={handleVerifyRecipient}
                  disabled={isLoading || !recipientEmail.trim()}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400"
                >
                  {isLoading ? "Checking..." : "Verify"}
                </button>
              </div>
            </div>

            {/* Permission Level */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Permission
              </label>
              <select
                value={permissionLevel}
                onChange={(e) => setPermissionLevel(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
              >
                <option value="read_only">View Only</option>
                <option value="read_write">Can Edit</option>
                <option value="admin">Full Access</option>
              </select>
            </div>

            {/* Share Button */}
            <button
              onClick={handleShareCollection}
              disabled={isLoading || !recipientId || !recipientEmail}
              className="w-full py-2 px-4 rounded-lg text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400"
            >
              {isLoading ? "Sharing..." : "Share Folder"}
            </button>
          </div>
        </div>

        {/* Current Members */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">
            People with Access ({collectionMembers.length})
          </h2>

          {collectionMembers.length === 0 ? (
            <p className="text-center text-gray-500 py-8">
              No one else has access to this folder yet
            </p>
          ) : (
            <div className="space-y-3">
              {collectionMembers.map((member, index) => (
                <div
                  key={`${member.recipient_id}-${index}`}
                  className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                >
                  <div>
                    <p className="font-medium text-gray-900">
                      {member.recipient_email}
                    </p>
                    <p className="text-sm text-gray-500 capitalize">
                      {member.permission_level.replace("_", " ")}
                    </p>
                  </div>
                  <button
                    onClick={() =>
                      handleRemoveMember(
                        member.recipient_id,
                        member.recipient_email,
                      )
                    }
                    disabled={isLoading}
                    className="text-gray-400 hover:text-red-600"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default withPasswordProtection(CollectionShare);
