// File: src/pages/User/FileManager/Collections/CollectionCreate.jsx
import React, { useState } from "react";
import { useNavigate } from "react-router";
import { useFiles, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  ArrowLeftIcon,
  FolderIcon,
  PhotoIcon,
  CheckIcon,
} from "@heroicons/react/24/outline";

const CollectionCreate = () => {
  const navigate = useNavigate();
  const { createCollectionManager } = useFiles();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [collectionName, setCollectionName] = useState("");
  const [collectionType, setCollectionType] = useState("folder");

  const handleCreateCollection = async (e) => {
    e.preventDefault();

    if (!collectionName.trim()) {
      setError("Please enter a name");
      return;
    }

    if (!createCollectionManager) {
      setError("Service not available. Please refresh the page.");
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      const result = await createCollectionManager.createCollection(
        {
          name: collectionName.trim(),
          collection_type: collectionType,
        },
        null,
      );

      if (result.success) {
        navigate(`/file-manager/collections/${result.collectionId}`);
      }
    } catch (err) {
      setError("Could not create folder. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => navigate("/file-manager")}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to My Files
          </button>
          <h1 className="text-2xl font-semibold text-gray-900">
            Create New Folder
          </h1>
        </div>

        {/* Form */}
        <form
          onSubmit={handleCreateCollection}
          className="bg-white rounded-lg border border-gray-200 p-6"
        >
          {/* Name Input */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Folder Name
            </label>
            <input
              type="text"
              value={collectionName}
              onChange={(e) => setCollectionName(e.target.value)}
              placeholder="Enter folder name"
              required
              disabled={isLoading}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
            />
          </div>

          {/* Type Selection */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-3">
              Folder Type
            </label>
            <div className="grid grid-cols-2 gap-3">
              <label
                className={`relative flex cursor-pointer rounded-lg border p-4 ${
                  collectionType === "folder"
                    ? "border-red-800 bg-red-50"
                    : "border-gray-300"
                }`}
              >
                <input
                  type="radio"
                  name="type"
                  value="folder"
                  checked={collectionType === "folder"}
                  onChange={(e) => setCollectionType(e.target.value)}
                  className="sr-only"
                />
                <div className="flex items-center">
                  <FolderIcon
                    className={`h-5 w-5 mr-3 ${
                      collectionType === "folder"
                        ? "text-red-800"
                        : "text-gray-400"
                    }`}
                  />
                  <div>
                    <span className="block font-medium text-gray-900">
                      Documents
                    </span>
                    <span className="text-sm text-gray-500">For files</span>
                  </div>
                </div>
                {collectionType === "folder" && (
                  <CheckIcon className="absolute right-3 top-3 h-5 w-5 text-red-800" />
                )}
              </label>

              <label
                className={`relative flex cursor-pointer rounded-lg border p-4 ${
                  collectionType === "album"
                    ? "border-red-800 bg-red-50"
                    : "border-gray-300"
                }`}
              >
                <input
                  type="radio"
                  name="type"
                  value="album"
                  checked={collectionType === "album"}
                  onChange={(e) => setCollectionType(e.target.value)}
                  className="sr-only"
                />
                <div className="flex items-center">
                  <PhotoIcon
                    className={`h-5 w-5 mr-3 ${
                      collectionType === "album"
                        ? "text-red-800"
                        : "text-gray-400"
                    }`}
                  />
                  <div>
                    <span className="block font-medium text-gray-900">
                      Photos
                    </span>
                    <span className="text-sm text-gray-500">For images</span>
                  </div>
                </div>
                {collectionType === "album" && (
                  <CheckIcon className="absolute right-3 top-3 h-5 w-5 text-red-800" />
                )}
              </label>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-6 p-3 rounded-lg bg-red-50 border border-red-200">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}

          {/* Submit Button */}
          <button
            type="submit"
            disabled={isLoading || !collectionName.trim()}
            className="w-full py-2 px-4 rounded-lg text-white bg-red-800 hover:bg-red-900 disabled:bg-gray-400"
          >
            {isLoading ? "Creating..." : "Create Folder"}
          </button>
        </form>
      </div>
    </div>
  );
};

export default withPasswordProtection(CollectionCreate);
