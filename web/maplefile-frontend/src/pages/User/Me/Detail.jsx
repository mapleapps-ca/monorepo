// File: src/pages/User/Me/Detail.jsx
import React, { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { useAuth } from "../../../services/Services";
import withPasswordProtection from "../../../hocs/withPasswordProtection";
import Navigation from "../../../components/Navigation";
import {
  UserIcon,
  KeyIcon,
  ShieldCheckIcon,
  LockClosedIcon,
  ClipboardDocumentIcon,
  EyeIcon,
  EyeSlashIcon,
  CheckIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
  DocumentTextIcon,
  CalendarIcon,
  EnvelopeIcon,
  PhoneIcon,
  GlobeAltIcon,
  ClockIcon,
  PencilIcon,
  TrashIcon,
  XMarkIcon,
  CameraIcon,
  SparklesIcon,
} from "@heroicons/react/24/outline";

const MeDetail = () => {
  const navigate = useNavigate();
  const { authManager, meManager } = useAuth();

  // Page state
  const [userProfile, setUserProfile] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [activeTab, setActiveTab] = useState("profile");

  // Edit mode states
  const [isEditing, setIsEditing] = useState(false);
  const [editLoading, setEditLoading] = useState(false);
  const [editError, setEditError] = useState("");
  const [formData, setFormData] = useState({
    email: "",
    first_name: "",
    last_name: "",
    phone: "",
    country: "",
    timezone: "",
  });

  // Keys state
  const [showKeys, setShowKeys] = useState(false);
  const [copiedKey, setCopiedKey] = useState("");

  // Delete account states
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deletePassword, setDeletePassword] = useState("");
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState("");

  // Countries and timezones lists
  const countries = [
    "Canada",
    "United States",
    "United Kingdom",
    "Australia",
    "Germany",
    "France",
    "Japan",
    "South Korea",
    "Brazil",
    "Mexico",
    "India",
    "Other",
  ];

  const timezones = [
    { value: "America/Toronto", label: "Toronto (EST/EDT)" },
    { value: "America/Vancouver", label: "Vancouver (PST/PDT)" },
    { value: "America/New_York", label: "New York (EST/EDT)" },
    { value: "America/Los_Angeles", label: "Los Angeles (PST/PDT)" },
    { value: "America/Chicago", label: "Chicago (CST/CDT)" },
    { value: "America/Denver", label: "Denver (MST/MDT)" },
    { value: "Europe/London", label: "London (GMT/BST)" },
    { value: "Europe/Paris", label: "Paris (CET/CEST)" },
    { value: "Europe/Berlin", label: "Berlin (CET/CEST)" },
    { value: "Asia/Tokyo", label: "Tokyo (JST)" },
    { value: "Asia/Seoul", label: "Seoul (KST)" },
    { value: "Australia/Sydney", label: "Sydney (AEST/AEDT)" },
    { value: "UTC", label: "UTC" },
  ];

  const tabs = [
    { id: "profile", label: "Profile", icon: UserIcon },
    { id: "security", label: "Security", icon: ShieldCheckIcon },
    { id: "keys", label: "Encryption Keys", icon: KeyIcon },
  ];

  const loadUserProfile = useCallback(async () => {
    if (!meManager) return;

    setIsLoading(true);
    setError("");

    try {
      console.log("[Profile] Loading user profile...");
      const profile = await meManager.getCurrentUser();
      setUserProfile(profile);

      setFormData({
        email: profile.email || "",
        first_name: profile.first_name || "",
        last_name: profile.last_name || "",
        phone: profile.phone || "",
        country: profile.country || "Canada",
        timezone: profile.timezone || "America/Toronto",
      });

      console.log("[Profile] Profile loaded successfully");
    } catch (err) {
      console.error("[Profile] Failed to load profile:", err);
      setError(err.message || "Failed to load profile.");
    } finally {
      setIsLoading(false);
    }
  }, [meManager]);

  useEffect(() => {
    if (meManager && authManager?.isAuthenticated()) {
      loadUserProfile();
    }
  }, [meManager, authManager, loadUserProfile]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
    if (editError) setEditError("");
  };

  const handleSaveChanges = async (e) => {
    e.preventDefault();
    setEditLoading(true);
    setEditError("");

    try {
      const updatedProfile = await meManager.updateCurrentUser(formData);
      setUserProfile(updatedProfile);
      setIsEditing(false);
      console.log("Profile updated successfully");
    } catch (err) {
      console.error("Failed to update profile:", err);
      setEditError(err.message || "Failed to update profile");
    } finally {
      setEditLoading(false);
    }
  };

  const handleCancelEdit = () => {
    if (userProfile) {
      setFormData({
        email: userProfile.email || "",
        first_name: userProfile.first_name || "",
        last_name: userProfile.last_name || "",
        phone: userProfile.phone || "",
        country: userProfile.country || "Canada",
        timezone: userProfile.timezone || "America/Toronto",
      });
    }
    setIsEditing(false);
    setEditError("");
  };

  const handleDeleteAccount = async (e) => {
    e.preventDefault();
    if (!deletePassword) {
      setDeleteError("Password is required to delete your account.");
      return;
    }

    setDeleteLoading(true);
    setDeleteError("");

    try {
      await meManager.deleteCurrentUser(deletePassword);
      if (authManager?.logout) {
        authManager.logout();
      }
      sessionStorage.clear();
      localStorage.removeItem("mapleapps_access_token");
      localStorage.removeItem("mapleapps_refresh_token");
      localStorage.removeItem("mapleapps_user_email");
      navigate("/");
    } catch (err) {
      console.error("Failed to delete account:", err);
      setDeleteError(
        err.message || "Failed to delete account. Please check your password.",
      );
    } finally {
      setDeleteLoading(false);
    }
  };

  const handleCopyKey = async (keyType, keyValue) => {
    try {
      await navigator.clipboard.writeText(keyValue);
      setCopiedKey(keyType);
      setTimeout(() => setCopiedKey(""), 3000);
    } catch (error) {
      console.error("Failed to copy:", error);
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return "N/A";
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const getRoleInfoText = (role) => {
    switch (role) {
      case 1:
        return "Root User";
      case 2:
        return "Company User";
      case 3:
        return "Individual User";
      default:
        return "Unknown";
    }
  };

  // Style definitions for buttons and badges, as the new design uses abstract classes
  const btn_primary =
    "inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg text-white bg-red-700 hover:bg-red-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200";
  const btn_secondary =
    "inline-flex items-center justify-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200";
  const btn_danger =
    "inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200";
  const badge_success =
    "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800";
  const badge_danger =
    "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800";
  const badge_warning =
    "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800";
  const input_style =
    "block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-red-500 focus:border-red-500 sm:text-sm disabled:opacity-50";
  const label_style = "block text-sm font-medium text-gray-700 mb-1";

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />

      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8 animate-fade-in-down">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 flex items-center">
                My Account
                <SparklesIcon className="h-8 w-8 text-yellow-500 ml-2" />
              </h1>
              <p className="text-gray-600 mt-1">
                Manage your profile and security settings
              </p>
            </div>
            <button
              onClick={loadUserProfile}
              disabled={isLoading}
              className={btn_secondary}
            >
              <ArrowPathIcon
                className={`h-4 w-4 mr-2 ${isLoading ? "animate-spin" : ""}`}
              />
              {isLoading ? "Refreshing..." : "Refresh"}
            </button>
          </div>
        </div>

        {/* Loading State */}
        {isLoading && !userProfile && (
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-red-800 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading your profile...</p>
            </div>
          </div>
        )}

        {/* Error State */}
        {error && !userProfile && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-6">
            <div className="flex items-start">
              <ExclamationTriangleIcon className="h-6 w-6 text-red-500 mr-3 flex-shrink-0 mt-1" />
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-red-800 mb-2">
                  Failed to Load Profile
                </h3>
                <p className="text-red-700">{error}</p>
                <button
                  onClick={loadUserProfile}
                  className={`mt-4 ${btn_secondary}`}
                >
                  <ArrowPathIcon className="h-4 w-4 mr-2" />
                  Try Again
                </button>
              </div>
            </div>
          </div>
        )}

        {userProfile && (
          <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
            {/* Sidebar */}
            <div className="lg:col-span-1">
              <div className="bg-white rounded-xl shadow-lg border border-gray-100/50 p-6 mb-6 text-center animate-fade-in-up">
                <div className="relative inline-block mb-4">
                  <div className="h-24 w-24 bg-gradient-to-br from-red-600 to-red-800 rounded-2xl flex items-center justify-center text-white text-3xl font-bold mx-auto">
                    {userProfile.first_name?.charAt(0)}
                    {userProfile.last_name?.charAt(0)}
                  </div>
                  <button className="absolute bottom-0 right-0 h-8 w-8 bg-white rounded-full shadow-lg flex items-center justify-center hover:bg-gray-50 transition-colors duration-200">
                    <CameraIcon className="h-4 w-4 text-gray-600" />
                  </button>
                </div>
                <h2 className="text-lg font-semibold text-gray-900">
                  {userProfile.first_name} {userProfile.last_name}
                </h2>
                <p className="text-sm text-gray-600">{userProfile.email}</p>
                <div className={`mt-4 ${badge_success}`}>
                  <CheckIcon className="h-3 w-3 mr-1" />
                  Verified User
                </div>
              </div>

              <nav
                className="space-y-1 animate-fade-in-up"
                style={{ animationDelay: "100ms" }}
              >
                {tabs.map((tab) => (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id)}
                    className={`w-full flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-all duration-200 ${
                      activeTab === tab.id
                        ? "bg-gradient-to-r from-red-700 to-red-800 text-white shadow-md"
                        : "text-gray-700 hover:bg-gray-100"
                    }`}
                  >
                    <tab.icon className="h-5 w-5 mr-3" />
                    {tab.label}
                  </button>
                ))}
              </nav>
            </div>

            {/* Main Content */}
            <div className="lg:col-span-3">
              {/* Profile Tab */}
              {activeTab === "profile" && (
                <div className="space-y-6 animate-fade-in-up">
                  <div className="bg-white rounded-xl shadow-lg border border-gray-100/50 p-6">
                    <div className="flex items-center justify-between mb-6">
                      <h3 className="text-lg font-semibold text-gray-900">
                        Personal Information
                      </h3>
                      {!isEditing && (
                        <button
                          onClick={() => setIsEditing(true)}
                          className={btn_secondary}
                        >
                          <PencilIcon className="h-4 w-4 mr-2" />
                          Edit Profile
                        </button>
                      )}
                    </div>

                    {isEditing ? (
                      <form onSubmit={handleSaveChanges} className="space-y-4">
                        {editError && (
                          <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                            <p className="text-sm text-red-700">{editError}</p>
                          </div>
                        )}
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                          <div>
                            <label htmlFor="first_name" className={label_style}>
                              First Name
                            </label>
                            <input
                              id="first_name"
                              name="first_name"
                              type="text"
                              value={formData.first_name}
                              onChange={handleInputChange}
                              className={input_style}
                              disabled={editLoading}
                              required
                            />
                          </div>
                          <div>
                            <label htmlFor="last_name" className={label_style}>
                              Last Name
                            </label>
                            <input
                              id="last_name"
                              name="last_name"
                              type="text"
                              value={formData.last_name}
                              onChange={handleInputChange}
                              className={input_style}
                              disabled={editLoading}
                              required
                            />
                          </div>
                          <div>
                            <label htmlFor="email" className={label_style}>
                              Email
                            </label>
                            <input
                              id="email"
                              name="email"
                              type="email"
                              value={formData.email}
                              onChange={handleInputChange}
                              className={input_style}
                              disabled={editLoading}
                              required
                            />
                          </div>
                          <div>
                            <label htmlFor="phone" className={label_style}>
                              Phone
                            </label>
                            <input
                              id="phone"
                              name="phone"
                              type="tel"
                              value={formData.phone}
                              onChange={handleInputChange}
                              className={input_style}
                              disabled={editLoading}
                            />
                          </div>
                          <div>
                            <label htmlFor="country" className={label_style}>
                              Country
                            </label>
                            <select
                              id="country"
                              name="country"
                              value={formData.country}
                              onChange={handleInputChange}
                              className={input_style}
                              disabled={editLoading}
                              required
                            >
                              {countries.map((c) => (
                                <option key={c} value={c}>
                                  {c}
                                </option>
                              ))}
                            </select>
                          </div>
                          <div>
                            <label htmlFor="timezone" className={label_style}>
                              Timezone
                            </label>
                            <select
                              id="timezone"
                              name="timezone"
                              value={formData.timezone}
                              onChange={handleInputChange}
                              className={input_style}
                              disabled={editLoading}
                              required
                            >
                              {timezones.map((tz) => (
                                <option key={tz.value} value={tz.value}>
                                  {tz.label}
                                </option>
                              ))}
                            </select>
                          </div>
                        </div>

                        <div className="flex justify-end space-x-3 pt-4">
                          <button
                            type="button"
                            onClick={handleCancelEdit}
                            className={btn_secondary}
                            disabled={editLoading}
                          >
                            Cancel
                          </button>
                          <button
                            type="submit"
                            className={btn_primary}
                            disabled={editLoading}
                          >
                            {editLoading ? "Saving..." : "Save Changes"}
                          </button>
                        </div>
                      </form>
                    ) : (
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="flex items-start space-x-3">
                          <UserIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                          <div>
                            <p className="text-sm text-gray-600">Full Name</p>
                            <p className="font-medium text-gray-900">
                              {userProfile.first_name} {userProfile.last_name}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-start space-x-3">
                          <EnvelopeIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                          <div>
                            <p className="text-sm text-gray-600">Email</p>
                            <p className="font-medium text-gray-900">
                              {userProfile.email}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-start space-x-3">
                          <PhoneIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                          <div>
                            <p className="text-sm text-gray-600">Phone</p>
                            <p className="font-medium text-gray-900">
                              {userProfile.phone || "N/A"}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-start space-x-3">
                          <GlobeAltIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                          <div>
                            <p className="text-sm text-gray-600">Country</p>
                            <p className="font-medium text-gray-900">
                              {userProfile.country}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-start space-x-3">
                          <ClockIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                          <div>
                            <p className="text-sm text-gray-600">Timezone</p>
                            <p className="font-medium text-gray-900">
                              {userProfile.timezone}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-start space-x-3">
                          <CalendarIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                          <div>
                            <p className="text-sm text-gray-600">
                              Member Since
                            </p>
                            <p className="font-medium text-gray-900">
                              {formatDate(userProfile.created_at)}
                            </p>
                          </div>
                        </div>
                      </div>
                    )}
                  </div>

                  <div className="bg-white rounded-xl shadow-lg border border-gray-100/50 p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-4">
                      Account Information
                    </h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div className="flex items-start space-x-3">
                        <DocumentTextIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                        <div>
                          <p className="text-sm text-gray-600">User ID</p>
                          <p className="font-mono text-sm text-gray-900">
                            {userProfile.id}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-start space-x-3">
                        <ShieldCheckIcon className="h-5 w-5 text-gray-400 mt-0.5" />
                        <div>
                          <p className="text-sm text-gray-600">Account Type</p>
                          <p className="font-medium text-gray-900">
                            {getRoleInfoText(
                              userProfile.user_role || userProfile.role,
                            )}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Security Tab */}
              {activeTab === "security" && (
                <div className="space-y-6 animate-fade-in-up">
                  <div className="bg-white rounded-xl shadow-lg border border-gray-100/50 p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-6">
                      Security Settings
                    </h3>
                    <div className="space-y-4">
                      <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
                        <div className="flex items-start">
                          <CheckIcon className="h-5 w-5 text-green-600 mr-3 flex-shrink-0 mt-0.5" />
                          <div>
                            <h4 className="text-sm font-medium text-green-900">
                              Two-Factor Authentication
                            </h4>
                            <p className="text-sm text-green-700 mt-1">
                              Your account is protected with 2FA (managed at the
                              identity provider level).
                            </p>
                          </div>
                        </div>
                      </div>
                      <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
                        <div className="flex items-start">
                          <ShieldCheckIcon className="h-5 w-5 text-blue-600 mr-3 flex-shrink-0 mt-0.5" />
                          <div>
                            <h4 className="text-sm font-medium text-blue-900">
                              End-to-End Encryption
                            </h4>
                            <p className="text-sm text-blue-700 mt-1">
                              All your files are encrypted with your keys before
                              they leave your device.
                            </p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="bg-white rounded-xl shadow-lg border-2 border-red-100 p-6">
                    <h3 className="text-lg font-semibold text-red-900 mb-2">
                      Danger Zone
                    </h3>
                    <p className="text-sm text-gray-600 mb-4">
                      Once you delete your account, there is no going back.
                      Please be certain.
                    </p>
                    <button
                      onClick={() => setShowDeleteModal(true)}
                      className={btn_danger}
                    >
                      <TrashIcon className="h-4 w-4 mr-2" />
                      Delete Account
                    </button>
                  </div>
                </div>
              )}

              {/* Encryption Keys Tab */}
              {activeTab === "keys" && (
                <div className="space-y-6 animate-fade-in-up">
                  <div className="bg-white rounded-xl shadow-lg border border-gray-100/50 p-6">
                    <div className="flex items-center justify-between mb-6">
                      <h3 className="text-lg font-semibold text-gray-900">
                        Encryption Keys
                      </h3>
                      <button
                        onClick={() => setShowKeys(!showKeys)}
                        className={btn_secondary}
                      >
                        {showKeys ? (
                          <>
                            <EyeSlashIcon className="h-4 w-4 mr-2" />
                            Hide Keys
                          </>
                        ) : (
                          <>
                            <EyeIcon className="h-4 w-4 mr-2" />
                            Show Keys
                          </>
                        )}
                      </button>
                    </div>

                    <div className="p-4 bg-amber-50 border border-amber-200 rounded-lg mb-6">
                      <div className="flex">
                        <ExclamationTriangleIcon className="h-5 w-5 text-amber-600 flex-shrink-0" />
                        <div className="ml-3">
                          <h4 className="text-sm font-medium text-amber-800">
                            Security Notice
                          </h4>
                          <p className="text-sm text-amber-700 mt-1">
                            Never share your private key with anyone. It's used
                            to decrypt your files.
                          </p>
                        </div>
                      </div>
                    </div>

                    <div className="space-y-4">
                      {userProfile.public_key && (
                        <div className="p-4 bg-gray-50 rounded-lg">
                          <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center space-x-2">
                              <KeyIcon className="h-5 w-5 text-gray-600" />
                              <span className="font-medium text-gray-900">
                                Public Key
                              </span>
                              <span className={badge_success}>
                                Safe to share
                              </span>
                            </div>
                            <button
                              onClick={() =>
                                handleCopyKey("public", userProfile.public_key)
                              }
                              className={`${btn_secondary} py-1 px-3 text-xs`}
                            >
                              {copiedKey === "public" ? (
                                <>
                                  <CheckIcon className="h-4 w-4 mr-1 text-green-600" />
                                  Copied!
                                </>
                              ) : (
                                <>
                                  <ClipboardDocumentIcon className="h-4 w-4 mr-1" />
                                  Copy
                                </>
                              )}
                            </button>
                          </div>
                          <div
                            className={`font-mono text-sm break-all bg-white p-3 rounded border border-gray-200 ${!showKeys && "select-none"}`}
                          >
                            {showKeys ? userProfile.public_key : "•".repeat(64)}
                          </div>
                        </div>
                      )}
                      {userProfile.encrypted_private_key && (
                        <div className="p-4 bg-gray-50 rounded-lg">
                          <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center space-x-2">
                              <LockClosedIcon className="h-5 w-5 text-gray-600" />
                              <span className="font-medium text-gray-900">
                                Encrypted Private Key
                              </span>
                              <span className={badge_danger}>Keep secret</span>
                            </div>
                            <button
                              onClick={() =>
                                handleCopyKey(
                                  "private",
                                  userProfile.encrypted_private_key,
                                )
                              }
                              className={`${btn_secondary} py-1 px-3 text-xs`}
                            >
                              {copiedKey === "private" ? (
                                <>
                                  <CheckIcon className="h-4 w-4 mr-1 text-green-600" />
                                  Copied!
                                </>
                              ) : (
                                <>
                                  <ClipboardDocumentIcon className="h-4 w-4 mr-1" />
                                  Copy
                                </>
                              )}
                            </button>
                          </div>
                          <div
                            className={`font-mono text-sm break-all bg-white p-3 rounded border border-gray-200 ${!showKeys && "select-none"}`}
                          >
                            {showKeys
                              ? userProfile.encrypted_private_key
                              : "•".repeat(96)}
                          </div>
                        </div>
                      )}
                      {userProfile.key_derivation_salt && (
                        <div className="p-4 bg-gray-50 rounded-lg">
                          <div className="flex items-center justify-between mb-2">
                            <div className="flex items-center space-x-2">
                              <ShieldCheckIcon className="h-5 w-5 text-gray-600" />
                              <span className="font-medium text-gray-900">
                                Key Derivation Salt
                              </span>
                              <span className={badge_warning}>
                                For recovery
                              </span>
                            </div>
                            <button
                              onClick={() =>
                                handleCopyKey(
                                  "salt",
                                  userProfile.key_derivation_salt,
                                )
                              }
                              className={`${btn_secondary} py-1 px-3 text-xs`}
                            >
                              {copiedKey === "salt" ? (
                                <>
                                  <CheckIcon className="h-4 w-4 mr-1 text-green-600" />
                                  Copied!
                                </>
                              ) : (
                                <>
                                  <ClipboardDocumentIcon className="h-4 w-4 mr-1" />
                                  Copy
                                </>
                              )}
                            </button>
                          </div>
                          <div
                            className={`font-mono text-sm break-all bg-white p-3 rounded border border-gray-200 ${!showKeys && "select-none"}`}
                          >
                            {showKeys
                              ? userProfile.key_derivation_salt
                              : "•".repeat(64)}
                          </div>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Delete Account Modal */}
      {showDeleteModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50 animate-fade-in-fast">
          <div className="bg-white rounded-xl shadow-xl max-w-md w-full p-6 animate-scale-in">
            <form onSubmit={handleDeleteAccount}>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900">
                  Delete Account
                </h3>
                <button
                  type="button"
                  onClick={() => setShowDeleteModal(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <XMarkIcon className="h-5 w-5" />
                </button>
              </div>

              <div className="mb-4 text-center">
                <div className="flex items-center justify-center h-12 w-12 bg-red-100 rounded-full mx-auto mb-4">
                  <ExclamationTriangleIcon className="h-6 w-6 text-red-600" />
                </div>
                <p className="text-gray-700 mb-4">
                  This action cannot be undone. All your data will be
                  permanently deleted.
                </p>
              </div>

              {deleteError && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                  <p className="text-sm text-red-700">{deleteError}</p>
                </div>
              )}

              <div className="mb-4">
                <label htmlFor="delete_password" className={label_style}>
                  Enter your password to confirm
                </label>
                <input
                  id="delete_password"
                  type="password"
                  placeholder="Your password"
                  value={deletePassword}
                  onChange={(e) => {
                    setDeletePassword(e.target.value);
                    if (deleteError) setDeleteError("");
                  }}
                  className={input_style}
                  required
                  disabled={deleteLoading}
                />
              </div>

              <div className="flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => setShowDeleteModal(false)}
                  className={btn_secondary}
                  disabled={deleteLoading}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className={btn_danger}
                  disabled={deleteLoading || !deletePassword}
                >
                  {deleteLoading ? "Deleting..." : "Delete My Account"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* CSS Animations */}
      <style jsx>{`
        @keyframes fade-in-down {
          from {
            opacity: 0;
            transform: translateY(-20px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        .animate-fade-in-down {
          animation: fade-in-down 0.5s ease-out both;
        }

        @keyframes fade-in-up {
          from {
            opacity: 0;
            transform: translateY(20px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        .animate-fade-in-up {
          animation: fade-in-up 0.5s ease-out both;
        }

        @keyframes fade-in-fast {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }
        .animate-fade-in-fast {
          animation: fade-in-fast 0.2s ease-in-out both;
        }

        @keyframes scale-in {
          from {
            opacity: 0;
            transform: scale(0.95);
          }
          to {
            opacity: 1;
            transform: scale(1);
          }
        }
        .animate-scale-in {
          animation: scale-in 0.3s ease-out both;
        }
      `}</style>
    </div>
  );
};

export default withPasswordProtection(MeDetail);
