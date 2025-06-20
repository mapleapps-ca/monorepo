import { useNavigate } from "react-router";
import { useCollectionOperations } from "../hooks/useCollections";
import CollectionForm from "../components/collections/CollectionForm";

/**
 * Create Collection Page
 * Follows Single Responsibility Principle - only handles collection creation
 */
const CreateCollectionPage = () => {
  const navigate = useNavigate();
  const { createCollection, loading, error } = useCollectionOperations();

  const handleSubmit = async (formData) => {
    try {
      const response = await createCollection(formData);
      // Navigate to the new collection's detail page
      navigate(`/collections/${response.id}`, {
        state: { message: "Collection created successfully!" },
      });
    } catch (error) {
      console.error("Failed to create collection:", error);
      // Error is handled by the hook
    }
  };

  const handleCancel = () => {
    navigate("/collections");
  };

  return (
    <div className="max-w-md mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Create Collection</h1>
        <p className="text-gray-600 mt-1">
          Create a new folder or album to organize your files
        </p>
      </div>

      {/* Error Display */}
      {error && (
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
                Error creating collection
              </h3>
              <p className="text-sm text-red-600 mt-1">{error}</p>
            </div>
          </div>
        </div>
      )}

      {/* Form */}
      <div className="bg-white shadow rounded-lg p-6">
        <CollectionForm
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          loading={loading}
        />
      </div>
    </div>
  );
};

export default CreateCollectionPage;
