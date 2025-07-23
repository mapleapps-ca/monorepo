// File: src/pages/FileManager/Search/SearchResults.jsx
import React, { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import Navigation from "../../../../components/Navigation";
import {
  MagnifyingGlassIcon,
  FolderIcon,
  DocumentIcon,
  PhotoIcon,
  AdjustmentsHorizontalIcon,
  FunnelIcon,
  XMarkIcon,
  CalendarIcon,
  ChevronRightIcon,
  StarIcon,
  ArrowDownTrayIcon,
  ShareIcon,
  CheckIcon,
  ClockIcon,
  ChevronDownIcon,
  UserIcon,
  ServerIcon,
} from "@heroicons/react/24/outline";
import { StarIcon as StarIconSolid } from "@heroicons/react/24/solid";

const SearchResults = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const initialQuery = searchParams.get("q") || "";

  const [searchQuery, setSearchQuery] = useState(initialQuery);
  const [showFilters, setShowFilters] = useState(false);
  const [selectedFilters, setSelectedFilters] = useState({
    type: "all",
    dateRange: "all",
    size: "all",
    collection: "all",
    sharedStatus: "all",
  });
  const [sortBy, setSortBy] = useState("relevance");
  const [selectedItems, setSelectedItems] = useState(new Set());

  // Mock search results
  const mockResults = {
    query: searchQuery,
    totalResults: 47,
    results: [
      {
        id: 1,
        type: "file",
        name: "Q4 Financial Report 2024.pdf",
        fileType: "pdf",
        size: "2.4 MB",
        modified: "2 hours ago",
        collection: { id: "1", name: "Work Documents" },
        matchedIn: ["filename", "content"],
        snippet:
          "...quarterly financial results show a 15% increase in revenue...",
        starred: true,
      },
      {
        id: 2,
        type: "collection",
        name: "Financial Reports",
        collectionType: "folder",
        itemCount: 24,
        modified: "1 day ago",
        matchedIn: ["name"],
        starred: false,
      },
      {
        id: 3,
        type: "file",
        name: "Budget Planning 2024.xlsx",
        fileType: "spreadsheet",
        size: "1.8 MB",
        modified: "3 days ago",
        collection: { id: "1", name: "Work Documents" },
        matchedIn: ["filename"],
        snippet: "...annual budget allocation for Q4 operations...",
        starred: false,
      },
      {
        id: 4,
        type: "file",
        name: "Team Meeting Notes - Q4 Planning.docx",
        fileType: "document",
        size: "156 KB",
        modified: "1 week ago",
        collection: { id: "3", name: "Meeting Notes" },
        matchedIn: ["content"],
        snippet: "...discussed Q4 targets and financial projections...",
        starred: true,
      },
      {
        id: 5,
        type: "collection",
        name: "Q4 Marketing Campaign",
        collectionType: "album",
        itemCount: 89,
        modified: "2 weeks ago",
        matchedIn: ["name", "description"],
        starred: false,
      },
    ],
    suggestions: [
      "Q4 report",
      "financial statements",
      "quarterly results",
      "2024 budget",
    ],
  };

  const getIcon = (item) => {
    if (item.type === "collection") {
      return item.collectionType === "album" ? (
        <PhotoIcon className="h-5 w-5" />
      ) : (
        <FolderIcon className="h-5 w-5" />
      );
    }

    switch (item.fileType) {
      case "pdf":
        return <DocumentIcon className="h-5 w-5 text-red-600" />;
      case "spreadsheet":
        return <DocumentIcon className="h-5 w-5 text-green-600" />;
      case "document":
        return <DocumentIcon className="h-5 w-5 text-blue-600" />;
      case "image":
        return <PhotoIcon className="h-5 w-5 text-pink-600" />;
      default:
        return <DocumentIcon className="h-5 w-5" />;
    }
  };

  const getIconBackground = (item) => {
    if (item.type === "collection") {
      return item.collectionType === "album"
        ? "bg-pink-100 text-pink-600"
        : "bg-blue-100 text-blue-600";
    }

    switch (item.fileType) {
      case "pdf":
        return "bg-red-100";
      case "spreadsheet":
        return "bg-green-100";
      case "document":
        return "bg-blue-100";
      case "image":
        return "bg-pink-100";
      default:
        return "bg-gray-100";
    }
  };

  const handleSearch = (e) => {
    e.preventDefault();
    navigate(`/file-manager/search?q=${encodeURIComponent(searchQuery)}`);
  };

  const highlightMatch = (text, query) => {
    if (!query) return text;
    const parts = text.split(new RegExp(`(${query})`, "gi"));
    return parts.map((part, i) =>
      part.toLowerCase() === query.toLowerCase() ? (
        <mark key={i} className="bg-yellow-200 text-gray-900 font-semibold">
          {part}
        </mark>
      ) : (
        part
      ),
    );
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Search Header */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <form onSubmit={handleSearch} className="max-w-3xl mx-auto">
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-4 top-1/2 transform -translate-y-1/2 h-6 w-6 text-gray-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search files and collections..."
                className="w-full pl-12 pr-4 py-4 text-lg border border-gray-300 rounded-xl focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                autoFocus
              />
              <button
                type="submit"
                className="absolute right-2 top-1/2 transform -translate-y-1/2 px-6 py-2 bg-gradient-to-r from-red-800 to-red-900 text-white rounded-lg hover:from-red-900 hover:to-red-950 transition-all duration-200"
              >
                Search
              </button>
            </div>
          </form>

          {/* Search Suggestions */}
          {mockResults.suggestions.length > 0 && (
            <div className="mt-4 flex items-center justify-center space-x-2">
              <span className="text-sm text-gray-500">Try:</span>
              {mockResults.suggestions.map((suggestion, index) => (
                <button
                  key={index}
                  onClick={() => {
                    setSearchQuery(suggestion);
                    navigate(
                      `/file-manager/search?q=${encodeURIComponent(suggestion)}`,
                    );
                  }}
                  className="text-sm text-blue-600 hover:text-blue-700 hover:underline"
                >
                  {suggestion}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Results Header */}
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">
              {mockResults.totalResults} results for "{initialQuery}"
            </h1>
            <p className="text-sm text-gray-500 mt-1">
              Found in files, collections, and shared items
            </p>
          </div>

          <div className="flex items-center space-x-3">
            {/* Sort Dropdown */}
            <div className="relative">
              <button className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                Sort by {sortBy}
                <ChevronDownIcon className="h-4 w-4 ml-2" />
              </button>
            </div>

            {/* Filter Button */}
            <button
              onClick={() => setShowFilters(!showFilters)}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
            >
              <FunnelIcon className="h-4 w-4 mr-2" />
              Filters
              {Object.values(selectedFilters).some((v) => v !== "all") && (
                <span className="ml-2 inline-flex items-center justify-center h-5 w-5 bg-red-600 text-white text-xs rounded-full">
                  {
                    Object.values(selectedFilters).filter((v) => v !== "all")
                      .length
                  }
                </span>
              )}
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Filters Sidebar */}
          {showFilters && (
            <div className="lg:col-span-1">
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="font-semibold text-gray-900">Filters</h3>
                  <button
                    onClick={() =>
                      setSelectedFilters({
                        type: "all",
                        dateRange: "all",
                        size: "all",
                        collection: "all",
                        sharedStatus: "all",
                      })
                    }
                    className="text-sm text-red-600 hover:text-red-700"
                  >
                    Clear all
                  </button>
                </div>

                <div className="space-y-6">
                  {/* Type Filter */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-3">
                      Type
                    </h4>
                    <div className="space-y-2">
                      {[
                        "all",
                        "files",
                        "collections",
                        "images",
                        "documents",
                      ].map((type) => (
                        <label key={type} className="flex items-center">
                          <input
                            type="radio"
                            name="type"
                            value={type}
                            checked={selectedFilters.type === type}
                            onChange={(e) =>
                              setSelectedFilters({
                                ...selectedFilters,
                                type: e.target.value,
                              })
                            }
                            className="h-4 w-4 text-red-600 border-gray-300 focus:ring-red-500"
                          />
                          <span className="ml-2 text-sm text-gray-700 capitalize">
                            {type === "all" ? "All types" : type}
                          </span>
                        </label>
                      ))}
                    </div>
                  </div>

                  {/* Date Range Filter */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-3">
                      Modified
                    </h4>
                    <div className="space-y-2">
                      {[
                        { value: "all", label: "Any time" },
                        { value: "today", label: "Today" },
                        { value: "week", label: "Past week" },
                        { value: "month", label: "Past month" },
                        { value: "year", label: "Past year" },
                      ].map((option) => (
                        <label key={option.value} className="flex items-center">
                          <input
                            type="radio"
                            name="dateRange"
                            value={option.value}
                            checked={selectedFilters.dateRange === option.value}
                            onChange={(e) =>
                              setSelectedFilters({
                                ...selectedFilters,
                                dateRange: e.target.value,
                              })
                            }
                            className="h-4 w-4 text-red-600 border-gray-300 focus:ring-red-500"
                          />
                          <span className="ml-2 text-sm text-gray-700">
                            {option.label}
                          </span>
                        </label>
                      ))}
                    </div>
                  </div>

                  {/* Size Filter */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-3">
                      Size
                    </h4>
                    <div className="space-y-2">
                      {[
                        { value: "all", label: "Any size" },
                        { value: "small", label: "Less than 10 MB" },
                        { value: "medium", label: "10 MB - 100 MB" },
                        { value: "large", label: "More than 100 MB" },
                      ].map((option) => (
                        <label key={option.value} className="flex items-center">
                          <input
                            type="radio"
                            name="size"
                            value={option.value}
                            checked={selectedFilters.size === option.value}
                            onChange={(e) =>
                              setSelectedFilters({
                                ...selectedFilters,
                                size: e.target.value,
                              })
                            }
                            className="h-4 w-4 text-red-600 border-gray-300 focus:ring-red-500"
                          />
                          <span className="ml-2 text-sm text-gray-700">
                            {option.label}
                          </span>
                        </label>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Search Results */}
          <div className={showFilters ? "lg:col-span-3" : "lg:col-span-4"}>
            <div className="space-y-4">
              {mockResults.results.map((result) => (
                <div
                  key={result.id}
                  className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 hover:shadow-md transition-all duration-200 cursor-pointer"
                  onClick={() => {
                    if (result.type === "collection") {
                      navigate(`/file-manager/collections/${result.id}`);
                    } else {
                      navigate(`/file-manager/files/${result.id}`);
                    }
                  }}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-4 flex-1">
                      {/* Icon */}
                      <div
                        className={`flex items-center justify-center h-12 w-12 rounded-lg flex-shrink-0 ${getIconBackground(result)}`}
                      >
                        {getIcon(result)}
                      </div>

                      {/* Content */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center">
                          <h3 className="text-lg font-semibold text-gray-900">
                            {highlightMatch(result.name, initialQuery)}
                          </h3>
                          {result.starred && (
                            <StarIconSolid className="h-4 w-4 text-yellow-400 ml-2 flex-shrink-0" />
                          )}
                        </div>

                        {/* Metadata */}
                        <div className="flex items-center space-x-4 mt-1 text-sm text-gray-500">
                          {result.type === "collection" ? (
                            <>
                              <span className="capitalize">
                                {result.collectionType}
                              </span>
                              <span>•</span>
                              <span>{result.itemCount} items</span>
                            </>
                          ) : (
                            <>
                              <span>{result.size}</span>
                              <span>•</span>
                              <span className="flex items-center">
                                <FolderIcon className="h-4 w-4 mr-1" />
                                {result.collection.name}
                              </span>
                            </>
                          )}
                          <span>•</span>
                          <span className="flex items-center">
                            <ClockIcon className="h-4 w-4 mr-1" />
                            {result.modified}
                          </span>
                        </div>

                        {/* Snippet */}
                        {result.snippet && (
                          <p className="mt-2 text-sm text-gray-600 line-clamp-2">
                            {highlightMatch(result.snippet, initialQuery)}
                          </p>
                        )}

                        {/* Match info */}
                        <div className="flex items-center space-x-2 mt-2">
                          {result.matchedIn.map((location, index) => (
                            <span
                              key={index}
                              className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800"
                            >
                              Match in {location}
                            </span>
                          ))}
                        </div>
                      </div>
                    </div>

                    {/* Actions */}
                    <div className="flex items-center space-x-2 ml-4">
                      {result.type === "file" && (
                        <>
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              // Download action
                            }}
                            className="p-2 text-gray-400 hover:text-gray-600"
                          >
                            <ArrowDownTrayIcon className="h-5 w-5" />
                          </button>
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              // Share action
                            }}
                            className="p-2 text-gray-400 hover:text-gray-600"
                          >
                            <ShareIcon className="h-5 w-5" />
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* No Results */}
            {mockResults.results.length === 0 && (
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12 text-center">
                <MagnifyingGlassIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 mb-2">
                  No results found
                </h3>
                <p className="text-gray-500">
                  Try adjusting your search terms or filters
                </p>
              </div>
            )}

            {/* Load More */}
            {mockResults.results.length > 0 && (
              <div className="mt-8 text-center">
                <button className="inline-flex items-center px-6 py-3 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
                  Load More Results
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SearchResults;
