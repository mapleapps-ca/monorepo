// File: src/pages/User/FileManager/Search/SearchResults.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { useFiles, useCrypto, useAuth } from "../../../../services/Services";
import withPasswordProtection from "../../../../hocs/withPasswordProtection";
import Navigation from "../../../../components/Navigation";
import {
  MagnifyingGlassIcon,
  FolderIcon,
  DocumentIcon,
  PhotoIcon,
  ArrowDownTrayIcon,
  ArrowLeftIcon,
} from "@heroicons/react/24/outline";

const SearchResults = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const query = searchParams.get("q") || "";

  const { listFileManager, listCollectionManager, downloadFileManager } =
    useFiles();
  const { CollectionCryptoService } = useCrypto();
  const { authManager } = useAuth();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [searchQuery, setSearchQuery] = useState(query);
  const [results, setResults] = useState({ files: [], collections: [] });
  const [downloadingFiles, setDownloadingFiles] = useState(new Set());

  // Search function
  const performSearch = useCallback(
    async (searchTerm) => {
      if (!searchTerm.trim() || !listFileManager || !listCollectionManager)
        return;

      setIsLoading(true);
      setError("");

      try {
        // Search in collections
        const collectionsResult =
          await listCollectionManager.listCollections(false);
        let matchedCollections = [];

        if (collectionsResult.collections) {
          matchedCollections = collectionsResult.collections.filter(
            (collection) => {
              const name = collection.name || "";
              return name.toLowerCase().includes(searchTerm.toLowerCase());
            },
          );
        }

        // Search in files across all collections
        const filesResult = await listFileManager.listFiles({
          states: ["active"],
          force_refresh: false,
        });

        let matchedFiles = [];
        if (filesResult.files) {
          matchedFiles = filesResult.files.filter((file) => {
            const name = file.name || "";
            return name.toLowerCase().includes(searchTerm.toLowerCase());
          });
        }

        setResults({
          collections: matchedCollections,
          files: matchedFiles,
        });
      } catch (err) {
        setError("Search failed. Please try again.");
        console.error("Search error:", err);
      } finally {
        setIsLoading(false);
      }
    },
    [listFileManager, listCollectionManager],
  );

  // Perform initial search
  useEffect(() => {
    if (
      query &&
      listFileManager &&
      listCollectionManager &&
      authManager?.isAuthenticated()
    ) {
      performSearch(query);
    }
  }, [
    query,
    listFileManager,
    listCollectionManager,
    authManager,
    performSearch,
  ]);

  // Handle new search
  const handleSearch = (e) => {
    e.preventDefault();
    navigate(`/file-manager/search?q=${encodeURIComponent(searchQuery)}`);
    performSearch(searchQuery);
  };

  // Handle file download
  const handleDownloadFile = async (fileId, fileName) => {
    if (!downloadFileManager) return;

    try {
      setDownloadingFiles((prev) => new Set(prev).add(fileId));
      await downloadFileManager.downloadFile(fileId, { saveToDisk: true });
    } catch (err) {
      setError("Could not download file");
    } finally {
      setDownloadingFiles((prev) => {
        const next = new Set(prev);
        next.delete(fileId);
        return next;
      });
    }
  };

  // Get file icon
  const getFileIcon = (file) => {
    const mimeType = file.mime_type || "";
    if (mimeType.startsWith("image/"))
      return <PhotoIcon className="h-5 w-5 text-pink-600" />;
    return <DocumentIcon className="h-5 w-5 text-gray-600" />;
  };

  const totalResults = results.collections.length + results.files.length;

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-6">
          <button
            onClick={() => navigate("/file-manager")}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to My Files
          </button>

          <h1 className="text-2xl font-semibold text-gray-900">
            Search Results
          </h1>
        </div>

        {/* Search Bar */}
        <form onSubmit={handleSearch} className="mb-8">
          <div className="relative max-w-2xl">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search files and folders..."
              className="w-full pl-10 pr-20 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500"
              autoFocus
            />
            <button
              type="submit"
              disabled={isLoading}
              className="absolute right-2 top-1/2 transform -translate-y-1/2 px-4 py-1.5 bg-red-800 text-white rounded-md hover:bg-red-900 disabled:bg-gray-400"
            >
              Search
            </button>
          </div>
        </form>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-3 rounded-lg bg-red-50 border border-red-200">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        {/* Results */}
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Searching...</p>
            </div>
          </div>
        ) : query && totalResults === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
            <MagnifyingGlassIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              No results found
            </h3>
            <p className="text-gray-500">
              Try different keywords or check your spelling
            </p>
          </div>
        ) : query ? (
          <div className="space-y-6">
            {/* Results Summary */}
            <p className="text-sm text-gray-600">
              Found {totalResults} result{totalResults !== 1 ? "s" : ""} for "
              {query}"
            </p>

            {/* Collections */}
            {results.collections.length > 0 && (
              <div>
                <h2 className="text-lg font-medium text-gray-900 mb-4">
                  Folders
                </h2>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {results.collections.map((collection) => (
                    <div
                      key={collection.id}
                      onClick={() =>
                        navigate(`/file-manager/collections/${collection.id}`)
                      }
                      className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md hover:border-gray-300 cursor-pointer transition-all"
                    >
                      <div className="flex items-center space-x-3">
                        <div className="flex items-center justify-center h-10 w-10 rounded-lg bg-blue-100">
                          <FolderIcon className="h-6 w-6 text-blue-600" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <h3 className="font-medium text-gray-900 truncate">
                            {collection.name || "Encrypted Folder"}
                          </h3>
                          <p className="text-sm text-gray-500">
                            {collection.file_count || 0} files
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Files */}
            {results.files.length > 0 && (
              <div>
                <h2 className="text-lg font-medium text-gray-900 mb-4">
                  Files
                </h2>
                <div className="bg-white rounded-lg border border-gray-200 divide-y divide-gray-200">
                  {results.files.map((file) => (
                    <div
                      key={file.id}
                      className="p-4 hover:bg-gray-50 flex items-center justify-between"
                    >
                      <div
                        className="flex items-center space-x-3 flex-1 cursor-pointer"
                        onClick={() =>
                          navigate(`/file-manager/files/${file.id}`)
                        }
                      >
                        <div className="flex items-center justify-center h-10 w-10 rounded-lg bg-gray-100">
                          {getFileIcon(file)}
                        </div>
                        <div>
                          <h3 className="text-sm font-medium text-gray-900">
                            {file.name || "Encrypted File"}
                          </h3>
                          <p className="text-xs text-gray-500">
                            {file.size
                              ? `${(file.size / 1024 / 1024).toFixed(1)} MB`
                              : "Unknown size"}
                          </p>
                        </div>
                      </div>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDownloadFile(file.id, file.name);
                        }}
                        disabled={
                          !file._isDecrypted || downloadingFiles.has(file.id)
                        }
                        className="p-2 text-gray-400 hover:text-gray-600 disabled:opacity-50"
                      >
                        {downloadingFiles.has(file.id) ? (
                          <div className="animate-spin rounded-full h-4 w-4 border-b border-gray-400"></div>
                        ) : (
                          <ArrowDownTrayIcon className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
            <MagnifyingGlassIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              Start searching
            </h3>
            <p className="text-gray-500">
              Enter keywords to search your files and folders
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default withPasswordProtection(SearchResults);
