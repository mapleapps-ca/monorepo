import { useState, useEffect } from "react";
import { useParams, useNavigate, useLocation } from "react-router";
import {
  useCollection,
  useCollectionOperations,
} from "../hooks/useCollections";
import { useServices } from "../contexts/ServiceContext";
import { formatDate, hasPermission } from "../utils";
import { COLLECTION_TYPE_LABELS, PERMISSION_LABELS } from "../constants";
import ShareCollectionForm from "../components/collections/ShareCollectionForm";
import Modal from "../components/ui/Modal";
import Button from "../components/ui/Button";
import Loading from "../components/ui/Loading";
import Error from "../components/ui/Error";

/**
 * Collection Detail Page
 * Follows Single Responsibility Principle - only handles collection detail display
 */
const CollectionDetailPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const { cryptoService } = useServices();

  const { collection, loading, error, refreshCollection } = useCollection(id);
  const {
    shareCollection,
    deleteCollection,
    loading: operationLoading,
  } = useCollectionOperations();

  const [showShareModal, setShowShareModal] = useState(false);
  const [successMessage, setSuccessMessage] = useState("");

  // Show success message from navigation state
  useEffect(() => {
    if (location.state?.message) {
      setSuccessMessage(location.state.message);
      // Clear the message from state
      window.history.replaceState({}, document.title);
    }
  }, [location.state]);

  // Auto-hide success message
  useEffect(() => {
    if (successMessage) {
      const timer = setTimeout(() => setSuccessMessage(""), 5000);
      return () => clearTimeout(timer);
    }
  }, [successMessage]);

  const getDisplayName = () => {
    if (!collection) return "";
    try {
      return cryptoService.decrypt(collection.encrypted_name);
    } catch (error) {
      return "Encrypted Collection";
    }
  };

  const getUserPermission = () => {
    // In a real app, you'd get this from authentication context
    // For now, assume owner has admin permission
    return "admin";
  };

  const handleShare = async (shareData) => {
    try {
      await shareCollection(id, shareData);
      setShowShareModal(false);
      refreshCollection();
      setSuccessMessage("Collection shared successfully!");
    } catch (error) {
      console.error("Failed to share collection:", error);
    }
  };

  const handleDelete = async () => {
    if (confirm(`Are you sure you want to delete "${getDisplayName()}"?`)) {
      try {
        await deleteCollection(id);
        navigate("/collections", {
          state: { message: "Collection deleted successfully!" },
        });
      } catch (error) {
        console.error("Failed to delete collection:", error);
      }
    }
  };

  if (loading) {
    return <Loading message="Loading collection..." />;
  }

  if (error) {
    return <Error message={error} onRetry={refreshCollection} />;
  }

  if (!collection) {
    return <Error message="Collection not found" />;
  }

  const userPermission = getUserPermission();
  const canEdit = hasPermission(userPermission, "read_write");
  const canShare = hasPermission(userPermission, "admin");
  const canDelete = hasPermission(userPermission, "admin");

  return (
    <div className="space-y-6">
      {/* Success Message */}
      {successMessage && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
          <div className="flex">
            <div className="text-green-500">
              <svg
                className="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-green-600">{successMessage}</p>
            </div>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center space-x-3 mb-2">
              <div className="text-blue-600">
                {collection.collection_type === "album" ? (
                  <svg
                    className="w-8 h-8"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                  </svg>
                ) : (
                  <svg
                    className="w-8 h-8"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2V7"
                    />
                  </svg>
                )}
              </div>
              <div>
                <h1 className="text-3xl font-bold text-gray-900">
                  {getDisplayName()}
                </h1>
                <p className="text-gray-600">
                  {COLLECTION_TYPE_LABELS[collection.collection_type]}
                </p>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
              <div>
                <p className="text-sm text-gray-500">Created</p>
                <p className="font-medium">
                  {formatDate(collection.created_at)}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Modified</p>
                <p className="font-medium">
                  {formatDate(collection.modified_at)}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Type</p>
                <p className="font-medium">
                  {COLLECTION_TYPE_LABELS[collection.collection_type]}
                </p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Your Permission</p>
                <p className="font-medium">
                  {PERMISSION_LABELS[userPermission]}
                </p>
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="flex space-x-2 ml-4">
            {canEdit && (
              <Button
                variant="outline"
                onClick={() => navigate(`/collections/${id}/edit`)}
              >
                <svg
                  className="w-4 h-4 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
                Edit
              </Button>
            )}

            {canShare && (
              <Button onClick={() => setShowShareModal(true)}>
                <svg
                  className="w-4 h-4 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z"
                  />
                </svg>
                Share
              </Button>
            )}

            {canDelete && (
              <Button variant="danger" onClick={handleDelete}>
                <svg
                  className="w-4 h-4 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
                Delete
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Members */}
      {collection.members && collection.members.length > 0 && (
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">
            Shared With
          </h2>
          <div className="space-y-3">
            {collection.members.map((member) => (
              <div
                key={member.id}
                className="flex items-center justify-between py-2"
              >
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
                    <span className="text-sm font-medium text-gray-700">
                      {member.recipient_email.charAt(0).toUpperCase()}
                    </span>
                  </div>
                  <div>
                    <p className="font-medium">{member.recipient_email}</p>
                    <p className="text-sm text-gray-500">
                      Added {formatDate(member.created_at)}
                    </p>
                  </div>
                </div>
                <span className="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
                  {PERMISSION_LABELS[member.permission_level]}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Files section would go here in a real app */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Files</h2>
        <div className="text-center py-8 text-gray-500">
          <svg
            className="w-12 h-12 mx-auto mb-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1}
              d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
            />
          </svg>
          <p>File management functionality would be implemented here</p>
        </div>
      </div>

      {/* Share Modal */}
      <Modal
        isOpen={showShareModal}
        onClose={() => setShowShareModal(false)}
        title="Share Collection"
        size="medium"
      >
        <ShareCollectionForm
          onSubmit={handleShare}
          onCancel={() => setShowShareModal(false)}
          loading={operationLoading}
        />
      </Modal>

      {/* Loading overlay */}
      {operationLoading && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-4">
            <Loading message="Processing..." />
          </div>
        </div>
      )}
    </div>
  );
};

export default CollectionDetailPage;
