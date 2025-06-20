import { useState } from "react";
import { Link } from "react-router";
import { useServices } from "../../contexts/ServiceContext";
import { formatDate } from "../../utils";
import { COLLECTION_TYPE_LABELS } from "../../constants";
import Button from "../ui/Button";

/**
 * Collection Card component for displaying collection information
 * Follows Single Responsibility Principle - only handles collection card display
 */
const CollectionCard = ({ collection, onDelete, onEdit }) => {
  const { cryptoService } = useServices();
  const [showActions, setShowActions] = useState(false);

  // Decrypt collection name for display
  const getDisplayName = () => {
    try {
      return cryptoService.decrypt(collection.encrypted_name);
    } catch (error) {
      return "Encrypted Collection";
    }
  };

  const getTypeIcon = () => {
    if (collection.collection_type === "album") {
      return (
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
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
      );
    }
    return (
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
          d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2V7"
        />
      </svg>
    );
  };

  const handleDelete = (e) => {
    e.preventDefault();
    e.stopPropagation();

    if (confirm(`Are you sure you want to delete "${getDisplayName()}"?`)) {
      onDelete(collection.id);
    }
  };

  const handleEdit = (e) => {
    e.preventDefault();
    e.stopPropagation();
    onEdit(collection.id);
  };

  return (
    <div
      className="group relative bg-white border border-gray-200 rounded-lg p-4 hover:shadow-lg transition-all duration-200 cursor-pointer"
      onMouseEnter={() => setShowActions(true)}
      onMouseLeave={() => setShowActions(false)}
    >
      <Link to={`/collections/${collection.id}`} className="block">
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center space-x-2">
            <div className="text-blue-600">{getTypeIcon()}</div>
            <div className="flex-1 min-w-0">
              <h3 className="text-lg font-semibold text-gray-900 truncate">
                {getDisplayName()}
              </h3>
              <p className="text-sm text-gray-500">
                {COLLECTION_TYPE_LABELS[collection.collection_type]}
              </p>
            </div>
          </div>

          {/* Actions */}
          {showActions && (
            <div className="flex space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <button
                onClick={handleEdit}
                className="p-1 text-gray-400 hover:text-blue-600 transition-colors"
                title="Edit collection"
              >
                <svg
                  className="w-4 h-4"
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
              </button>
              <button
                onClick={handleDelete}
                className="p-1 text-gray-400 hover:text-red-600 transition-colors"
                title="Delete collection"
              >
                <svg
                  className="w-4 h-4"
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
              </button>
            </div>
          )}
        </div>

        {/* Metadata */}
        <div className="space-y-2 text-sm text-gray-600">
          <div className="flex justify-between">
            <span>Created:</span>
            <span>{formatDate(collection.created_at)}</span>
          </div>
          <div className="flex justify-between">
            <span>Modified:</span>
            <span>{formatDate(collection.modified_at)}</span>
          </div>
          {collection.members && collection.members.length > 1 && (
            <div className="flex justify-between">
              <span>Shared with:</span>
              <span>{collection.members.length - 1} people</span>
            </div>
          )}
        </div>

        {/* Status indicators */}
        <div className="mt-3 flex items-center space-x-2">
          {collection.parent_id && (
            <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-gray-100 text-gray-800 rounded-full">
              Nested
            </span>
          )}
          {collection.members && collection.members.length > 1 && (
            <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
              Shared
            </span>
          )}
        </div>
      </Link>
    </div>
  );
};

export default CollectionCard;
