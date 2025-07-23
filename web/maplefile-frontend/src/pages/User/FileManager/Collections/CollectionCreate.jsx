// File: src/pages/FileManager/Collections/CollectionCreate.jsx
import React, { useState } from "react";
import { Link, useNavigate } from "react-router";
import Navigation from "../../../../components/Navigation";
import {
  FolderIcon,
  PhotoIcon,
  ArrowLeftIcon,
  InformationCircleIcon,
  ShieldCheckIcon,
  LockClosedIcon,
  UsersIcon,
  CheckIcon,
  XMarkIcon,
  PlusIcon,
  SparklesIcon,
  DocumentDuplicateIcon,
  GlobeAltIcon,
  EyeIcon,
  ChevronRightIcon,
  HomeIcon,
} from "@heroicons/react/24/outline";

const CollectionCreate = () => {
  const navigate = useNavigate();
  const [collectionType, setCollectionType] = useState("folder");
  const [collectionName, setCollectionName] = useState("");
  const [description, setDescription] = useState("");
  const [parentCollection, setParentCollection] = useState(null);
  const [privacyMode, setPrivacyMode] = useState("private");
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [showUserSearch, setShowUserSearch] = useState(false);
  const [userSearchQuery, setUserSearchQuery] = useState("");

  // Mock users for sharing
  const mockUsers = [
    { id: 1, name: "Alice Johnson", email: "alice@example.com", avatar: "AJ" },
    { id: 2, name: "Bob Smith", email: "bob@example.com", avatar: "BS" },
    { id: 3, name: "Carol Williams", email: "carol@example.com", avatar: "CW" },
  ];

  const handleCreate = () => {
    // Mock creation - just navigate back
    navigate("/file-manager/collections");
  };

  const handleAddUser = (user) => {
    if (!selectedUsers.find((u) => u.id === user.id)) {
      setSelectedUsers([...selectedUsers, user]);
    }
    setShowUserSearch(false);
    setUserSearchQuery("");
  };

  const handleRemoveUser = (userId) => {
    setSelectedUsers(selectedUsers.filter((u) => u.id !== userId));
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Breadcrumb */}
        <div className="flex items-center space-x-2 text-sm text-gray-600 mb-6">
          <HomeIcon className="h-4 w-4" />
          <ChevronRightIcon className="h-3 w-3" />
          <Link to="/file-manager/collections" className="hover:text-gray-900">
            My Files
          </Link>
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900">Create Collection</span>
        </div>

        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => navigate("/file-manager/collections")}
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors duration-200"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to Collections
          </button>

          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Create New Collection
          </h1>
          <p className="text-gray-600">
            Organize your files with encrypted collections
          </p>
        </div>

        {/* Main Form */}
        <div className="space-y-6">
          {/* Collection Type Selection */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
              <SparklesIcon className="h-5 w-5 mr-2 text-gray-500" />
              Collection Type
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <label
                className={`relative flex cursor-pointer rounded-lg border-2 p-6 hover:border-gray-300 transition-all duration-200 ${
                  collectionType === "folder"
                    ? "border-red-500 bg-red-50"
                    : "border-gray-200"
                }`}
              >
                <input
                  type="radio"
                  name="collection-type"
                  value="folder"
                  checked={collectionType === "folder"}
                  onChange={(e) => setCollectionType(e.target.value)}
                  className="sr-only"
                />
                <div className="flex items-center">
                  <div
                    className={`flex items-center justify-center h-12 w-12 rounded-lg mr-4 ${
                      collectionType === "folder"
                        ? "bg-blue-600 text-white"
                        : "bg-blue-100 text-blue-600"
                    }`}
                  >
                    <FolderIcon className="h-6 w-6" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-900">Folder</h3>
                    <p className="text-sm text-gray-500 mt-1">
                      For documents and general files
                    </p>
                  </div>
                </div>
                {collectionType === "folder" && (
                  <CheckIcon className="absolute top-4 right-4 h-5 w-5 text-red-600" />
                )}
              </label>

              <label
                className={`relative flex cursor-pointer rounded-lg border-2 p-6 hover:border-gray-300 transition-all duration-200 ${
                  collectionType === "album"
                    ? "border-red-500 bg-red-50"
                    : "border-gray-200"
                }`}
              >
                <input
                  type="radio"
                  name="collection-type"
                  value="album"
                  checked={collectionType === "album"}
                  onChange={(e) => setCollectionType(e.target.value)}
                  className="sr-only"
                />
                <div className="flex items-center">
                  <div
                    className={`flex items-center justify-center h-12 w-12 rounded-lg mr-4 ${
                      collectionType === "album"
                        ? "bg-pink-600 text-white"
                        : "bg-pink-100 text-pink-600"
                    }`}
                  >
                    <PhotoIcon className="h-6 w-6" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-900">Album</h3>
                    <p className="text-sm text-gray-500 mt-1">
                      For photos and media files
                    </p>
                  </div>
                </div>
                {collectionType === "album" && (
                  <CheckIcon className="absolute top-4 right-4 h-5 w-5 text-red-600" />
                )}
              </label>
            </div>
          </div>

          {/* Basic Information */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
              <DocumentDuplicateIcon className="h-5 w-5 mr-2 text-gray-500" />
              Basic Information
            </h2>

            <div className="space-y-4">
              <div>
                <label
                  htmlFor="name"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Collection Name <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  id="name"
                  value={collectionName}
                  onChange={(e) => setCollectionName(e.target.value)}
                  placeholder={
                    collectionType === "album"
                      ? "e.g., Summer Vacation 2024"
                      : "e.g., Work Documents"
                  }
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                />
              </div>

              <div>
                <label
                  htmlFor="description"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Description (Optional)
                </label>
                <textarea
                  id="description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  rows={3}
                  placeholder="Add a description to help you remember what's in this collection..."
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                />
              </div>

              <div>
                <label
                  htmlFor="parent"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Parent Collection (Optional)
                </label>
                <select
                  id="parent"
                  value={parentCollection || ""}
                  onChange={(e) => setParentCollection(e.target.value)}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                >
                  <option value="">Root level (no parent)</option>
                  <option value="1">Work Documents</option>
                  <option value="2">Personal Files</option>
                  <option value="3">Archive</option>
                </select>
              </div>
            </div>
          </div>

          {/* Privacy & Sharing */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center">
              <UsersIcon className="h-5 w-5 mr-2 text-gray-500" />
              Privacy & Sharing
            </h2>

            {/* Privacy Mode */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-3">
                Privacy Mode
              </label>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                <label
                  className={`relative flex cursor-pointer rounded-lg border p-4 hover:border-gray-300 transition-all duration-200 ${
                    privacyMode === "private"
                      ? "border-red-500 bg-red-50"
                      : "border-gray-200"
                  }`}
                >
                  <input
                    type="radio"
                    name="privacy"
                    value="private"
                    checked={privacyMode === "private"}
                    onChange={(e) => setPrivacyMode(e.target.value)}
                    className="sr-only"
                  />
                  <div className="flex items-center">
                    <LockClosedIcon
                      className={`h-5 w-5 mr-2 ${privacyMode === "private" ? "text-red-600" : "text-gray-400"}`}
                    />
                    <div>
                      <p className="font-medium text-gray-900">Private</p>
                      <p className="text-xs text-gray-500">Only you</p>
                    </div>
                  </div>
                  {privacyMode === "private" && (
                    <CheckIcon className="absolute top-3 right-3 h-4 w-4 text-red-600" />
                  )}
                </label>

                <label
                  className={`relative flex cursor-pointer rounded-lg border p-4 hover:border-gray-300 transition-all duration-200 ${
                    privacyMode === "shared"
                      ? "border-red-500 bg-red-50"
                      : "border-gray-200"
                  }`}
                >
                  <input
                    type="radio"
                    name="privacy"
                    value="shared"
                    checked={privacyMode === "shared"}
                    onChange={(e) => setPrivacyMode(e.target.value)}
                    className="sr-only"
                  />
                  <div className="flex items-center">
                    <UsersIcon
                      className={`h-5 w-5 mr-2 ${privacyMode === "shared" ? "text-red-600" : "text-gray-400"}`}
                    />
                    <div>
                      <p className="font-medium text-gray-900">Shared</p>
                      <p className="text-xs text-gray-500">Specific users</p>
                    </div>
                  </div>
                  {privacyMode === "shared" && (
                    <CheckIcon className="absolute top-3 right-3 h-4 w-4 text-red-600" />
                  )}
                </label>

                <label
                  className={`relative flex cursor-pointer rounded-lg border p-4 hover:border-gray-300 transition-all duration-200 ${
                    privacyMode === "public"
                      ? "border-red-500 bg-red-50"
                      : "border-gray-200"
                  }`}
                >
                  <input
                    type="radio"
                    name="privacy"
                    value="public"
                    checked={privacyMode === "public"}
                    onChange={(e) => setPrivacyMode(e.target.value)}
                    className="sr-only"
                  />
                  <div className="flex items-center">
                    <GlobeAltIcon
                      className={`h-5 w-5 mr-2 ${privacyMode === "public" ? "text-red-600" : "text-gray-400"}`}
                    />
                    <div>
                      <p className="font-medium text-gray-900">Public</p>
                      <p className="text-xs text-gray-500">Anyone with link</p>
                    </div>
                  </div>
                  {privacyMode === "public" && (
                    <CheckIcon className="absolute top-3 right-3 h-4 w-4 text-red-600" />
                  )}
                </label>
              </div>
            </div>

            {/* Share with Users */}
            {privacyMode === "shared" && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-3">
                  Share with Users
                </label>

                {/* Selected Users */}
                {selectedUsers.length > 0 && (
                  <div className="mb-3 space-y-2">
                    {selectedUsers.map((user) => (
                      <div
                        key={user.id}
                        className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                      >
                        <div className="flex items-center">
                          <div className="flex items-center justify-center h-8 w-8 bg-red-100 text-red-600 text-sm font-medium rounded-full mr-3">
                            {user.avatar}
                          </div>
                          <div>
                            <p className="text-sm font-medium text-gray-900">
                              {user.name}
                            </p>
                            <p className="text-xs text-gray-500">
                              {user.email}
                            </p>
                          </div>
                        </div>
                        <button
                          onClick={() => handleRemoveUser(user.id)}
                          className="text-gray-400 hover:text-red-600 transition-colors duration-200"
                        >
                          <XMarkIcon className="h-4 w-4" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                {/* Add User Button */}
                <button
                  onClick={() => setShowUserSearch(!showUserSearch)}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-all duration-200"
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  Add Users
                </button>

                {/* User Search Dropdown */}
                {showUserSearch && (
                  <div className="mt-3 p-4 bg-gray-50 rounded-lg border border-gray-200">
                    <input
                      type="text"
                      value={userSearchQuery}
                      onChange={(e) => setUserSearchQuery(e.target.value)}
                      placeholder="Search by name or email..."
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-red-500 focus:border-red-500"
                    />
                    <div className="mt-3 space-y-2 max-h-48 overflow-y-auto">
                      {mockUsers
                        .filter(
                          (u) => !selectedUsers.find((su) => su.id === u.id),
                        )
                        .map((user) => (
                          <button
                            key={user.id}
                            onClick={() => handleAddUser(user)}
                            className="w-full flex items-center p-2 hover:bg-white rounded-lg transition-colors duration-200"
                          >
                            <div className="flex items-center justify-center h-8 w-8 bg-red-100 text-red-600 text-sm font-medium rounded-full mr-3">
                              {user.avatar}
                            </div>
                            <div className="text-left">
                              <p className="text-sm font-medium text-gray-900">
                                {user.name}
                              </p>
                              <p className="text-xs text-gray-500">
                                {user.email}
                              </p>
                            </div>
                          </button>
                        ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Security Info */}
          <div className="bg-blue-50 border border-blue-200 rounded-xl p-4">
            <div className="flex items-start">
              <InformationCircleIcon className="h-5 w-5 text-blue-600 mr-3 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-blue-800">
                <h3 className="font-semibold mb-1">End-to-End Encryption</h3>
                <p>
                  Your collection will be encrypted with ChaCha20-Poly1305
                  encryption. Files added to this collection will inherit its
                  encryption settings.
                </p>
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center justify-between pt-6 border-t">
            <button
              onClick={() => navigate("/file-manager/collections")}
              className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-all duration-200"
            >
              Cancel
            </button>

            <div className="flex items-center space-x-3">
              <button className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-all duration-200">
                Save as Draft
              </button>
              <button
                onClick={handleCreate}
                disabled={!collectionName.trim()}
                className="inline-flex items-center px-6 py-2 border border-transparent rounded-lg shadow-sm text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
              >
                <CheckIcon className="h-4 w-4 mr-2" />
                Create Collection
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CollectionCreate;
