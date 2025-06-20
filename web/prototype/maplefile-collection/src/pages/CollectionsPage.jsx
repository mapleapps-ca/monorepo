import { useState } from "react";
import { useNavigate } from "react-router";
import {
  useCollections,
  useCollectionOperations,
} from "../hooks/useCollections";
import CollectionCard from "../components/collections/CollectionCard";
import Loading from "../components/ui/Loading";
import Error from "../components/ui/Error";
import Button from "../components/ui/Button";

/**
 * Collections Page - displays user's collections
 * Follows Single Responsibility Principle - only handles collections list display
 */
const CollectionsPage = () => {
  const navigate = useNavigate();
  const { collections, loading, error, refreshCollections } = useCollections({
    include_owned: true,
    include_shared: false,
  });
  const { deleteCollection, loading: operationLoading } =
    useCollectionOperations();
  const [filter, setFilter] = useState("all");

  const handleDeleteCollection = async (collectionId) => {
    try {
      await deleteCollection(collectionId);
      refreshCollections();
    } catch (error) {
      console.error("Failed to delete collection:", error);
      // Error is handled by the hook
    }
  };

  const handleEditCollection = (collectionId) => {
    navigate(`/collections/${collectionId}/edit`);
  };

  const filteredCollections = collections.filter((collection) => {
    if (filter === "folders") return collection.collection_type === "folder";
    if (filter === "albums") return collection.collection_type === "album";
    return true;
  });

  if (loading) {
    return <Loading message="Loading your collections..." />;
  }

  if (error) {
    return <Error message={error} onRetry={refreshCollections} />;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">My Collections</h1>
          <p className="text-gray-600 mt-1">Manage your folders and albums</p>
        </div>

        <Button onClick={() => navigate("/collections/new")}>
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
              d="M12 4v16m8-8H4"
            />
          </svg>
          New Collection
        </Button>
      </div>

      {/* Filters */}
      <div className="flex space-x-2">
        <button
          onClick={() => setFilter("all")}
          className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
            filter === "all"
              ? "bg-blue-100 text-blue-700"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
        >
          All ({collections.length})
        </button>
        <button
          onClick={() => setFilter("folders")}
          className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
            filter === "folders"
              ? "bg-blue-100 text-blue-700"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
        >
          Folders (
          {collections.filter((c) => c.collection_type === "folder").length})
        </button>
        <button
          onClick={() => setFilter("albums")}
          className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
            filter === "albums"
              ? "bg-blue-100 text-blue-700"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
        >
          Albums (
          {collections.filter((c) => c.collection_type === "album").length})
        </button>
      </div>

      {/* Collections Grid */}
      {filteredCollections.length === 0 ? (
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
                d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2V7"
              />
            </svg>
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {filter === "all" ? "No collections yet" : `No ${filter} found`}
          </h3>
          <p className="text-gray-600 mb-6">
            {filter === "all"
              ? "Create your first collection to get started organizing your files."
              : `You don't have any ${filter} yet.`}
          </p>
          <Button onClick={() => navigate("/collections/new")}>
            Create Collection
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {filteredCollections.map((collection) => (
            <CollectionCard
              key={collection.id}
              collection={collection}
              onDelete={handleDeleteCollection}
              onEdit={handleEditCollection}
            />
          ))}
        </div>
      )}

      {/* Loading overlay for operations */}
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

export default CollectionsPage;
