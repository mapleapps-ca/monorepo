// File: src/pages/FileManager/Collections/CollectionEdit.jsx
import React, { useState, useEffect } from "react";
import { Link, useNavigate, useParams } from "react-router";
import Navigation from "../../../components/Navigation";
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
  ChevronRightIcon,
  HomeIcon,
  ExclamationTriangleIcon,
  TrashIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";

const CollectionEdit = () => {
  const navigate = useNavigate();
  const { collectionId } = useParams();

  // Mock existing collection data
  const [collectionData, setCollectionData] = useState({
    id: collectionId || "1",
    name: "Work Documents",
    type: "folder",
    description: "Important work-related documents and presentations",
    parentCollection: null,
    privacyMode: "shared",
    sharedWith: [
      {
        id: 1,
        name: "Alice Johnson",
        email: "alice@example.com",
        avatar: "AJ",
        role: "viewer",
      },
      {
        id: 2,
        name: "Bob Smith",
        email: "bob@example.com",
        avatar: "BS",
        role: "editor",
      },
    ],
    created: "January 15, 2024",
    modified: "2 hours ago",
    itemCount: 24,
    totalSize: "1.2 GB",
  });

  const [formData, setFormData] = useState({
    name: collectionData.name,
    description: collectionData.description,
    parentCollection: collectionData.parentCollection,
    privacyMode: collectionData.privacyMode,
  });

  const [selectedUsers, setSelectedUsers] = useState(collectionData.sharedWith);
  const [showUserSearch, setShowUserSearch] = useState(false);
  const [userSearchQuery, setUserSearchQuery] = useState("");
  const [hasChanges, setHasChanges] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  // Mock users for sharing
  const mockUsers = [
    { id: 3, name: "Carol Williams", email: "carol@example.com", avatar: "CW" },
    { id: 4, name: "David Brown", email: "david@example.com", avatar: "DB" },
    { id: 5, name: "Eva Martinez", email: "eva@example.com", avatar: "EM" },
  ];

  useEffect(() => {
    // Check if any changes were made
    const changed =
      formData.name !== collectionData.name ||
      formData.description !== collectionData.description ||
      formData.parentCollection !== collectionData.parentCollection ||
      formData.privacyMode !== collectionData.privacyMode ||
      JSON.stringify(selectedUsers) !==
        JSON.stringify(collectionData.sharedWith);

    setHasChanges(changed);
  }, [formData, selectedUsers, collectionData]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleAddUser = (user) => {
    if (!selectedUsers.find((u) => u.id === user.id)) {
      setSelectedUsers([...selectedUsers, { ...user, role: "viewer" }]);
    }
    setShowUserSearch(false);
    setUserSearchQuery("");
  };

  const handleRemoveUser = (userId) => {
    setSelectedUsers(selectedUsers.filter((u) => u.id !== userId));
  };

  const handleUserRoleChange = (userId, newRole) => {
    setSelectedUsers(
      selectedUsers.map((u) => (u.id === userId ? { ...u, role: newRole } : u)),
    );
  };

  const handleSave = () => {
    setIsLoading(true);
    // Simulate save
    setTimeout(() => {
      navigate(`/file-manager/collections/${collectionId}`);
    }, 1000);
  };

  const handleDelete = () => {
    // Simulate delete
    navigate("/file-manager/collections");
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
          <Link
            to={`/file-manager/collections/${collectionId}`}
            className="hover:text-gray-900"
          >
            {collectionData.name}
          </Link>
          <ChevronRightIcon className="h-3 w-3" />
          <span className="font-medium text-gray-900">Edit</span>
        </div>

        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() =>
              navigate(`/file-manager/collections/${collectionId}`)
            }
            className="inline-flex items-center text-sm text-gray-600 hover:text-gray-900 mb-4 transition-colors duration-200"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to Collection
          </button>

          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">
                Edit Collection
              </h1>
              <p className="text-gray-600">
                Update collection settings and permissions
              </p>
            </div>

            {hasChanges && (
              <div className="flex items-center text-amber-600 bg-amber-50 px-4 py-2 rounded-lg">
                <ExclamationTriangleIcon className="h-5 w-5 mr-2" />
                <span className="text-sm font-medium">
                  You have unsaved changes
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Main Form */}
        <div className="space-y-6">
          {/* Collection Info Card */}
          <div className="bg-gray-50 rounded-xl border border-gray-200 p-6">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <div
                  className={`flex items-center justify-center h-12 w-12 rounded-lg ${
                    collectionData.type === "album"
                      ? "bg-pink-100 text-pink-600"
                      : "bg-blue-100 text-blue-600"
                  }`}
                >
                  {collectionData.type === "album" ? (
                    <PhotoIcon className="h-6 w-6" />
                  ) : (
                    <FolderIcon className="h-6 w-6" />
                  )}
                </div>
                <div>
                  <p className="text-sm text-gray-500">Collection Type</p>
                  <p className="font-semibold text-gray-900 capitalize">
                    {collectionData.type}
                  </p>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-6 text-center">
                <div>
                  <p className="text-2xl font-bold text-gray-900">
                    {collectionData.itemCount}
                  </p>
                  <p className="text-sm text-gray-500">Items</p>
                </div>
                <div>
                  <p className="text-2xl font-bold text-gray-900">
                    {collectionData.totalSize}
                  </p>
                  <p className="text-sm text-gray-500">Total Size</p>
                </div>
                <div>
                  <p className="text-2xl font-bold text-gray-900">
                    {selectedUsers.length}
                  </p>
                  <p className="text-sm text-gray-500">Shared</p>
                </div>
              </div>
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
                  name="name"
                  value={formData.name}
                  onChange={handleInputChange}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                />
              </div>

              <div>
                <label
                  htmlFor="description"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Description
                </label>
                <textarea
                  id="description"
                  name="description"
                  value={formData.description}
                  onChange={handleInputChange}
                  rows={3}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                />
              </div>

              <div>
                <label
                  htmlFor="parentCollection"
                  className="block text-sm font-medium text-gray-700 mb-2"
                >
                  Parent Collection
                </label>
                <select
                  id="parentCollection"
                  name="parentCollection"
                  value={formData.parentCollection || ""}
                  onChange={handleInputChange}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 transition-all duration-200"
                >
                  <option value="">Root level (no parent)</option>
                  <option value="2">Personal Files</option>
                  <option value="3">Archive</option>
                  <option value="4">Shared Projects</option>
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
                    formData.privacyMode === "private"
                      ? "border-red-500 bg-red-50"
                      : "border-gray-200"
                  }`}
                >
                  <input
                    type="radio"
                    name="privacyMode"
                    value="private"
                    checked={formData.privacyMode === "private"}
                    onChange={handleInputChange}
                    className="sr-only"
                  />
                  <div className="flex items-center">
                    <LockClosedIcon
                      className={`h-5 w-5 mr-2 ${formData.privacyMode === "private" ? "text-red-600" : "text-gray-400"}`}
                    />
                    <div>
                      <p className="font-medium text-gray-900">Private</p>
                      <p className="text-xs text-gray-500">Only you</p>
                    </div>
                  </div>
                  {formData.privacyMode === "private" && (
                    <CheckIcon className="absolute top-3 right-3 h-4 w-4 text-red-600" />
                  )}
                </label>

                <label
                  className={`relative flex cursor-pointer rounded-lg border p-4 hover:border-gray-300 transition-all duration-200 ${
                    formData.privacyMode === "shared"
                      ? "border-red-500 bg-red-50"
                      : "border-gray-200"
                  }`}
                >
                  <input
                    type="radio"
                    name="privacyMode"
                    value="shared"
                    checked={formData.privacyMode === "shared"}
                    onChange={handleInputChange}
                    className="sr-only"
                  />
                  <div className="flex items-center">
                    <UsersIcon
                      className={`h-5 w-5 mr-2 ${formData.privacyMode === "shared" ? "text-red-600" : "text-gray-400"}`}
                    />
                    <div>
                      <p className="font-medium text-gray-900">Shared</p>
                      <p className="text-xs text-gray-500">Specific users</p>
                    </div>
                  </div>
                  {formData.privacyMode === "shared" && (
                    <CheckIcon className="absolute top-3 right-3 h-4 w-4 text-red-600" />
                  )}
                </label>

                <label
                  className={`relative flex cursor-pointer rounded-lg border p-4 hover:border-gray-300 transition-all duration-200 ${
                    formData.privacyMode === "public"
                      ? "border-red-500 bg-red-50"
                      : "border-gray-200"
                  }`}
                >
                  <input
                    type="radio"
                    name="privacyMode"
                    value="public"
                    checked={formData.privacyMode === "public"}
                    onChange={handleInputChange}
                    className="sr-only"
                  />
                  <div className="flex items-center">
                    <GlobeAltIcon
                      className={`h-5 w-5 mr-2 ${formData.privacyMode === "public" ? "text-red-600" : "text-gray-400"}`}
                    />
                    <div>
                      <p className="font-medium text-gray-900">Public</p>
                      <p className="text-xs text-gray-500">Anyone with link</p>
                    </div>
                  </div>
                  {formData.privacyMode === "public" && (
                    <CheckIcon className="absolute top-3 right-3 h-4 w-4 text-red-600" />
                  )}
                </label>
              </div>
            </div>

            {/* Share with Users */}
            {formData.privacyMode === "shared" && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-3">
                  Shared with
                </label>

                {/* Current Shared Users */}
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
                        <div className="flex items-center space-x-2">
                          <select
                            value={user.role}
                            onChange={(e) =>
                              handleUserRoleChange(user.id, e.target.value)
                            }
                            className="text-sm px-3 py-1 border border-gray-300 rounded-lg"
                          >
                            <option value="viewer">Can view</option>
                            <option value="editor">Can edit</option>
                          </select>
                          <button
                            onClick={() => handleRemoveUser(user.id)}
                            className="text-gray-400 hover:text-red-600 transition-colors duration-200"
                          >
                            <XMarkIcon className="h-4 w-4" />
                          </button>
                        </div>
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

          {/* Danger Zone */}
          <div className="bg-red-50 border border-red-200 rounded-xl p-6">
            <h2 className="text-lg font-semibold text-red-900 mb-4 flex items-center">
              <ExclamationTriangleIcon className="h-5 w-5 mr-2" />
              Danger Zone
            </h2>

            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-medium text-gray-900">
                  Delete this collection
                </h3>
                <p className="text-sm text-gray-600 mt-1">
                  Once deleted, this collection and all its contents cannot be
                  recovered.
                </p>
              </div>
              <button
                onClick={() => setShowDeleteConfirm(true)}
                className="px-4 py-2 border border-red-300 rounded-lg text-red-700 hover:bg-red-100 transition-all duration-200"
              >
                Delete Collection
              </button>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center justify-between pt-6 border-t">
            <button
              onClick={() =>
                navigate(`/file-manager/collections/${collectionId}`)
              }
              className="px-6 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-all duration-200"
            >
              Cancel
            </button>

            <div className="flex items-center space-x-3">
              {hasChanges && (
                <button
                  onClick={() => window.location.reload()}
                  className="inline-flex items-center px-4 py-2 text-sm text-gray-600 hover:text-gray-900"
                >
                  <ArrowPathIcon className="h-4 w-4 mr-1" />
                  Reset Changes
                </button>
              )}
              <button
                onClick={handleSave}
                disabled={!hasChanges || !formData.name.trim() || isLoading}
                className="inline-flex items-center px-6 py-2 border border-transparent rounded-lg shadow-sm text-white bg-gradient-to-r from-red-800 to-red-900 hover:from-red-900 hover:to-red-950 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200"
              >
                {isLoading ? (
                  <>
                    <ArrowPathIcon className="h-4 w-4 mr-2 animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <CheckIcon className="h-4 w-4 mr-2" />
                    Save Changes
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">
                Delete Collection
              </h3>
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>

            <div className="mb-6">
              <div className="flex items-center justify-center h-12 w-12 bg-red-100 rounded-lg mb-4">
                <TrashIcon className="h-6 w-6 text-red-600" />
              </div>
              <p className="text-gray-700 mb-2">
                Are you sure you want to delete{" "}
                <strong>{collectionData.name}</strong>?
              </p>
              <p className="text-sm text-gray-600">
                This will permanently delete the collection and all{" "}
                {collectionData.itemCount} items inside it. This action cannot
                be undone.
              </p>
            </div>

            <div className="flex justify-end space-x-3">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
              >
                Delete Collection
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default CollectionEdit;
