// File: src/pages/User/FileManager/Collections/CollectionCreate.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useFiles, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  ArrowLeftIcon,
  FolderIcon,
  PhotoIcon,
  PlusIcon,
  LockClosedIcon,
  ShieldCheckIcon,
  ChevronRightIcon,
  HomeIcon,
  ExclamationTriangleIcon,
  CheckIcon,
  InformationCircleIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";

const CollectionCreate = () => {
  const navigate = useNavigate();
  const { createCollectionManager } = useFiles();
  const { authManager } = useAuth();

  // State management
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  // Form state
  const [collectionName, setCollectionName] = useState("");
  const [collectionType, setCollectionType] = useState("folder");

  // Computed values
  const isAuthenticated = authManager?.isAuthenticated() || false;
  const canCreateCollections = isAuthenticated && !isLoading;

  // Handle collection creation
  const handleCreateCollection = async () => {
    if (!collectionName.trim()) {
      setError("Collection name is required");
      return;
    }

    if (!createCollectionManager) {
      setError("Collection service not available. Please refresh the page.");
      return;
    }

    setIsLoading(true);
    setError("");
    setSuccess("");

    try {
      console.log("[CollectionCreate] Creating collection:", {
        name: collectionName.trim(),
        type: collectionType,
      });

      const result = await createCollectionManager.createCollection(
        {
          name: collectionName.trim(),
          collection_type: collectionType,
        },
        null, // Password handled by withPasswordProtection
      );

      if (result.success) {
        setSuccess(
          `${collectionType === "folder" ? "Folder" : "Album"} "${collectionName}" created successfully!`,
        );

        console.log("[CollectionCreate] Collection created successfully:", {
          id: result.collectionId,
          name: collectionName,
          type: collectionType,
        });

        // Navigate to the new collection after a brief delay
        setTimeout(() => {
          navigate(`/file-manager/collections/${result.collectionId}`);
        }, 1500);
      }
    } catch (err) {
      console.error("[CollectionCreate] Collection creation failed:", err);
      setError(err.message || "Failed to create collection");
    } finally {
      setIsLoading(false);
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

  // Handle form submission
  const handleSubmit = (e) => {
    e.preventDefault();
    handleCreateCollection();
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 mb-6">
          <HomeIcon className="h-4 w-4" />
          <ChevronRightIcon className="h-3 w-3" />
          <button
            onClick={() => navigate("/file-manager")}
            className="hover:text-gray-900 transition-colors duration-200"
          >
            My Files
          </button>
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900">Create Collection</span>
        </div>

        {/* Header */}
        <div className="text-center mb-8 animate-fade-in-up">
          <div className="flex justify-center mb-6">
            <div className="relative">
              <div className="flex items-center justify-center h-16 w-16 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-lg animate-pulse">
                <PlusIcon className="h-8 w-8 text-white" />
              </div>
              <div className="absolute -inset-1 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
            </div>
          </div>
          <h1 className="text-3xl font-black text-gray-900 mb-2">
            Create New Collection
          </h1>
          <p className="text-gray-600 mb-2">
            Organize your files with end-to-end encrypted folders and albums
          </p>
          <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
            <LockClosedIcon className="h-4 w-4 text-green-600" />
            <span>Automatically encrypted with your master password</span>
          </div>
        </div>

        {/* Error/Success Messages */}
        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-50 border border-red-200 animate-fade-in">
            <div className="flex items-center">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-3 flex-shrink-0" />
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-red-800">Error</h3>
                <p className="text-sm text-red-700 mt-1">{error}</p>
              </div>
              <button
                onClick={clearMessages}
                className="text-red-500 hover:text-red-700 transition-colors duration-200"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}

        {success && (
          <div className="mb-6 p-4 rounded-lg bg-green-50 border border-green-200 animate-fade-in">
            <div className="flex items-center">
              <CheckIcon className="h-5 w-5 text-green-500 mr-3 flex-shrink-0" />
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-green-800">
                  Success
                </h3>
                <p className="text-sm text-green-700 mt-1">{success}</p>
                <p className="text-xs text-green-600 mt-1">
                  Redirecting to your new collection...
                </p>
              </div>
              <button
                onClick={clearMessages}
                className="text-green-500 hover:text-green-700 transition-colors duration-200"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}

        {/* Main Form Card */}
        <div className="bg-white rounded-2xl shadow-2xl border border-gray-100 p-8 animate-fade-in-up-delay">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Collection Name */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-2">
                Collection Name
              </label>
              <input
                type="text"
                value={collectionName}
                onChange={(e) => setCollectionName(e.target.value)}
                placeholder="Enter a name for your collection"
                required
                disabled={isLoading}
                className={`w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200 disabled:bg-gray-50 disabled:cursor-not-allowed text-gray-900 placeholder-gray-500 ${
                  collectionName.length > 0
                    ? "border-green-300 bg-green-50"
                    : "border-gray-300"
                }`}
              />
              {collectionName.length > 0 && (
                <p className="mt-1 text-xs text-green-600 flex items-center">
                  <CheckIcon className="h-3 w-3 mr-1" />
                  Ready to be encrypted and stored securely
                </p>
              )}
            </div>

            {/* Collection Type */}
            <div>
              <label className="block text-sm font-semibold text-gray-700 mb-3">
                Collection Type
              </label>
              <div className="grid grid-cols-2 gap-4">
                {/* Folder Option */}
                <label
                  className={`relative flex cursor-pointer rounded-xl border p-6 focus:outline-none transition-all duration-200 ${
                    collectionType === "folder"
                      ? "border-red-600 bg-red-50 ring-2 ring-red-600"
                      : "border-gray-300 hover:border-gray-400 hover:bg-gray-50"
                  }`}
                >
                  <input
                    type="radio"
                    name="collection_type"
                    value="folder"
                    checked={collectionType === "folder"}
                    onChange={(e) => setCollectionType(e.target.value)}
                    className="sr-only"
                  />
                  <div className="flex flex-col items-center text-center w-full">
                    <div
                      className={`flex items-center justify-center h-16 w-16 rounded-xl mb-4 transition-all duration-200 ${
                        collectionType === "folder"
                          ? "bg-blue-100 text-blue-600 scale-110"
                          : "bg-gray-100 text-gray-400"
                      }`}
                    >
                      <FolderIcon className="h-8 w-8" />
                    </div>
                    <span className="block text-lg font-bold text-gray-900 mb-1">
                      📁 Folder
                    </span>
                    <span className="text-sm text-gray-500">
                      For documents, files, and general storage
                    </span>
                  </div>
                  {collectionType === "folder" && (
                    <CheckIcon className="absolute right-3 top-3 h-6 w-6 text-red-600" />
                  )}
                </label>

                {/* Album Option */}
                <label
                  className={`relative flex cursor-pointer rounded-xl border p-6 focus:outline-none transition-all duration-200 ${
                    collectionType === "album"
                      ? "border-red-600 bg-red-50 ring-2 ring-red-600"
                      : "border-gray-300 hover:border-gray-400 hover:bg-gray-50"
                  }`}
                >
                  <input
                    type="radio"
                    name="collection_type"
                    value="album"
                    checked={collectionType === "album"}
                    onChange={(e) => setCollectionType(e.target.value)}
                    className="sr-only"
                  />
                  <div className="flex flex-col items-center text-center w-full">
                    <div
                      className={`flex items-center justify-center h-16 w-16 rounded-xl mb-4 transition-all duration-200 ${
                        collectionType === "album"
                          ? "bg-pink-100 text-pink-600 scale-110"
                          : "bg-gray-100 text-gray-400"
                      }`}
                    >
                      <PhotoIcon className="h-8 w-8" />
                    </div>
                    <span className="block text-lg font-bold text-gray-900 mb-1">
                      📷 Album
                    </span>
                    <span className="text-sm text-gray-500">
                      For photos, videos, and media files
                    </span>
                  </div>
                  {collectionType === "album" && (
                    <CheckIcon className="absolute right-3 top-3 h-6 w-6 text-red-600" />
                  )}
                </label>
              </div>
            </div>

            {/* Create Button */}
            <button
              type="submit"
              disabled={isLoading || !collectionName.trim() || !isAuthenticated}
              className="group w-full flex justify-center items-center py-3 px-4 border border-transparent text-base font-semibold rounded-lg text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed transform hover:scale-[1.02] transition-all duration-200 shadow-lg hover:shadow-xl"
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-3"></div>
                  Creating {collectionType === "folder" ? "Folder" : "Album"}...
                </>
              ) : (
                <>
                  <PlusIcon className="h-5 w-5 mr-2 group-hover:scale-110 transition-transform duration-200" />
                  Create {collectionType === "folder" ? "Folder" : "Album"}
                </>
              )}
            </button>

            {/* Back to Files */}
            <div className="text-center">
              <button
                type="button"
                onClick={() => navigate("/file-manager")}
                className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 font-medium transition-colors duration-200"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-1" />
                Back to My Files
              </button>
            </div>
          </form>
        </div>

        {/* Security Trust Section */}
        <div className="mt-8 p-4 bg-gradient-to-r from-green-50 to-blue-50 rounded-lg border border-green-100">
          <div className="flex items-center justify-center mb-3">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-1">
                <LockClosedIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-semibold text-green-800">
                  ChaCha20-Poly1305 Encrypted
                </span>
              </div>
              <div className="flex items-center space-x-1">
                <ShieldCheckIcon className="h-4 w-4 text-blue-600" />
                <span className="text-xs font-semibold text-blue-800">
                  Zero-Knowledge Architecture
                </span>
              </div>
            </div>
          </div>
          <p className="text-center text-xs text-gray-600">
            Collections are encrypted on your device before storage using your
            master password
          </p>
        </div>

        {/* Info Section */}
        <div className="mt-6 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg border border-blue-100 p-4">
          <div className="flex items-start space-x-3">
            <InformationCircleIcon className="h-5 w-5 text-blue-600 mt-0.5" />
            <div>
              <h3 className="text-sm font-semibold text-blue-900 mb-2">
                What happens next?
              </h3>
              <ul className="text-sm text-blue-800 space-y-1">
                <li>
                  • Your collection will be encrypted automatically using your
                  master password
                </li>
                <li>
                  • A unique encryption key will be generated for this
                  collection
                </li>
                <li>
                  • You'll be redirected to your new collection to start adding
                  files
                </li>
                <li>
                  • All files added will be end-to-end encrypted before upload
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <style jsx>{`
        @keyframes fade-in-up {
          from {
            opacity: 0;
            transform: translateY(30px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        .animate-fade-in-up {
          animation: fade-in-up 0.6s ease-out;
        }

        .animate-fade-in-up-delay {
          animation: fade-in-up 0.6s ease-out 0.2s both;
        }

        @keyframes fade-in {
          from {
            opacity: 0;
            transform: translateY(10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        .animate-fade-in {
          animation: fade-in 0.4s ease-out;
        }
      `}</style>
    </div>
  );
};

// Export with password protection
export default withPasswordProtection(CollectionCreate);
