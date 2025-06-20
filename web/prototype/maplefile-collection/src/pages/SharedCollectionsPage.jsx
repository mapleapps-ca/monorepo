import { useCollections } from "../hooks/useCollections";
import CollectionCard from "../components/collections/CollectionCard";
import Loading from "../components/ui/Loading";
import Error from "../components/ui/Error";

/**
 * Shared Collections Page - displays collections shared with the user
 * Follows Single Responsibility Principle - only handles shared collections display
 */
const SharedCollectionsPage = () => {
  const { collections, loading, error, refreshCollections } = useCollections({
    include_owned: false,
    include_shared: true,
  });

  // For shared collections, users typically can't delete them
  // Only the owner can delete, so we'll just show a message
  const handleDeleteCollection = () => {
    alert(
      "You cannot delete shared collections. Contact the owner to remove access.",
    );
  };

  // Users can't edit shared collections unless they have admin permission
  const handleEditCollection = () => {
    alert(
      "You cannot edit this shared collection. Contact the owner for changes.",
    );
  };

  if (loading) {
    return <Loading message="Loading shared collections..." />;
  }

  if (error) {
    return <Error message={error} onRetry={refreshCollections} />;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Shared with Me</h1>
        <p className="text-gray-600 mt-1">
          Collections that others have shared with you
        </p>
      </div>

      {/* Collections Grid */}
      {collections.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-gray-400 mb-4">
            <svg
              className="w-16 h-16 mx-auto"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1}
                d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"
              />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            No shared collections
          </h3>
          <p className="text-gray-600">
            When someone shares a collection with you, it will appear here.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {collections.map((collection) => (
            <CollectionCard
              key={collection.id}
              collection={collection}
              onDelete={handleDeleteCollection}
              onEdit={handleEditCollection}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default SharedCollectionsPage;
