import { useParams, useNavigate } from "react-router";
import {
  useCollection,
  useCollectionOperations,
} from "../hooks/useCollections";
import CollectionForm from "../components/collections/CollectionForm";
import Loading from "../components/ui/Loading";
import Error from "../components/ui/Error";

/**
 * Edit Collection Page
 * Follows Single Responsibility Principle - only handles collection editing
 */
const EditCollectionPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const {
    collection,
    loading: fetchLoading,
    error: fetchError,
    refreshCollection,
  } = useCollection(id);
  const {
    updateCollection,
    loading: updateLoading,
    error: updateError,
  } = useCollectionOperations();

  const handleSubmit = async (formData) => {
    try {
      await updateCollection(id, formData);
      // Navigate back to collection detail page
      navigate(`/collections/${id}`, {
        state: { message: "Collection updated successfully!" },
      });
    } catch (error) {
      console.error("Failed to update collection:", error);
      // Error is handled by the hook
    }
  };

  const handleCancel = () => {
    navigate(`/collections/${id}`);
  };

  if (fetchLoading) {
    return <Loading message="Loading collection..." />;
  }

  if (fetchError) {
    return (
      <div className="max-w-md mx-auto">
        <Error message={fetchError} onRetry={refreshCollection} />
      </div>
    );
  }

  if (!collection) {
    return (
      <div className="max-w-md mx-auto">
        <Error message="Collection not found" />
      </div>
    );
  }

  return (
    <div className="max-w-md mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Edit Collection</h1>
        <p className="text-gray-600 mt-1">Update your collection details</p>
      </div>

      {/* Error Display */}
      {updateError && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex">
            <div className="text-red-500">
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
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">
                Error updating collection
              </h3>
              <p className="text-sm text-red-600 mt-1">{updateError}</p>
            </div>
          </div>
        </div>
      )}

      {/* Form */}
      <div className="bg-white shadow rounded-lg p-6">
        <CollectionForm
          collection={collection}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          loading={updateLoading}
        />
      </div>
    </div>
  );
};

export default EditCollectionPage;
