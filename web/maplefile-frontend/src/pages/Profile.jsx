// web/prototyping/maplefile-frontend/src/pages/Profile.jsx
import { useState, useEffect } from "react";
import { userAPI } from "../services/api";

function Profile() {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const response = await userAPI.getProfile();
        setProfile(response.data);
        console.log("Profile data:", response.data);
      } catch (err) {
        console.error("Error fetching profile:", err);
        setError(
          err.response?.data?.message ||
            err.message ||
            "Failed to load profile",
        );
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, []);

  // Helper function to format the role
  const formatRole = (role) => {
    switch (role) {
      case 1:
        return "Root";
      case 2:
        return "Company";
      case 3:
        return "Individual";
      default:
        return "Unknown";
    }
  };

  // Helper function to format the status
  const formatStatus = (status) => {
    switch (status) {
      case 1:
        return "Active";
      case 50:
        return "Locked";
      case 100:
        return "Archived";
      default:
        return "Unknown";
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-lg font-semibold">Loading profile...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-800 rounded-lg p-4 mt-4">
        <div className="font-bold">Error</div>
        <div>{error}</div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 text-yellow-800 rounded-lg p-4 mt-4">
        No profile data available
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">User Profile</h1>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        {/* Header with user name and email */}
        <div className="bg-blue-600 text-white p-6">
          <div className="flex items-center space-x-4">
            <div className="bg-white rounded-full p-3">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-8 w-8 text-blue-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                />
              </svg>
            </div>
            <div>
              <h2 className="text-xl font-semibold">{profile.name}</h2>
              <p className="text-blue-100">{profile.email}</p>
              <div className="mt-1">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  {formatRole(profile.role)}
                </span>
                <span className="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                  {formatStatus(profile.status)}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Personal Information */}
        <div className="p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            Personal Information
          </h3>
          <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-gray-500">First Name</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.first_name || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Last Name</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.last_name || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Phone</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.phone || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">User ID</dt>
              <dd className="mt-1 text-sm text-gray-900 truncate">
                {profile.id}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">
                Account Created
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.created_at
                  ? new Date(profile.created_at).toLocaleDateString()
                  : "Unknown"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">
                Email Verified
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.was_email_verified ? "Yes" : "No"}
              </dd>
            </div>
          </dl>
        </div>

        {/* Location */}
        <div className="bg-gray-50 p-6 border-t border-gray-200">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Location</h3>
          <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt className="text-sm font-medium text-gray-500">Country</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.country || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">
                Region/State
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.region || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">City</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.city || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Postal Code</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.postal_code || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">
                Address Line 1
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.address_line1 || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">
                Address Line 2
              </dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.address_line2 || "Not provided"}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Timezone</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {profile.timezone || "Not provided"}
              </dd>
            </div>
          </dl>
        </div>

        {/* Preferences */}
        <div className="p-6 border-t border-gray-200">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            Preferences
          </h3>
          <dl className="grid grid-cols-1 gap-y-4">
            <div>
              <dt className="text-sm font-medium text-gray-500">
                Receive Promotional Emails
              </dt>
              <dd className="mt-1">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${profile.agree_promotions ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"}`}
                >
                  {profile.agree_promotions ? "Yes" : "No"}
                </span>
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">
                Allow Tracking Across Third-Party Apps
              </dt>
              <dd className="mt-1">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${profile.agree_to_tracking_across_third_party_apps_and_services ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"}`}
                >
                  {profile.agree_to_tracking_across_third_party_apps_and_services
                    ? "Yes"
                    : "No"}
                </span>
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  );
}

export default Profile;
