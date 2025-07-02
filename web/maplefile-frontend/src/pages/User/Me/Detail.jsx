// monorepo/web/maplefile-frontend/src/pages/User/Me/Detail.jsx
import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import { useServices } from "../../../hooks/useService.jsx";
import useAuth from "../../../hooks/useAuth.js";
import withPasswordProtection from "../../../hocs/withPasswordProtection.jsx";

const MeDetail = () => {
  const navigate = useNavigate();
  const { meService, authService } = useServices();
  const { logout } = useAuth();

  // State for user data
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // State for editing
  const [isEditing, setIsEditing] = useState(false);
  const [editLoading, setEditLoading] = useState(false);
  const [editError, setEditError] = useState("");

  // State for form data
  const [formData, setFormData] = useState({
    email: "",
    first_name: "",
    last_name: "",
    phone: "",
    country: "",
    region: "",
    timezone: "",
    agree_promotions: false,
    agree_to_tracking_across_third_party_apps_and_services: false,
  });

  // State for account deletion
  const [showDeleteSection, setShowDeleteSection] = useState(false);
  const [deletePassword, setDeletePassword] = useState("");
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState("");

  // Countries list
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

  // Timezones list
  const timezones = [
    "America/Toronto",
    "America/New_York",
    "America/Los_Angeles",
    "America/Chicago",
    "America/Denver",
    "America/Vancouver",
    "Europe/London",
    "Europe/Paris",
    "Europe/Berlin",
    "Asia/Tokyo",
    "Asia/Seoul",
    "Australia/Sydney",
    "UTC",
  ];

  // Load user data on component mount
  useEffect(() => {
    loadUserData();
  }, []);

  const loadUserData = async () => {
    try {
      setLoading(true);
      setError("");

      const userData = await meService.getCurrentUser();
      setUser(userData);

      // Populate form data for editing
      setFormData({
        email: userData.email || "",
        first_name: userData.first_name || "",
        last_name: userData.last_name || "",
        phone: userData.phone || "",
        country: userData.country || "Canada",
        region: userData.region || "",
        timezone: userData.timezone || "America/Toronto",
        agree_promotions: userData.agree_promotions || false,
        agree_to_tracking_across_third_party_apps_and_services:
          userData.agree_to_tracking_across_third_party_apps_and_services ||
          false,
      });
    } catch (err) {
      console.error("Failed to load user data:", err);
      setError(err.message || "Failed to load user information");

      // If it's an authentication error, redirect to login
      if (
        err.message.includes("Session expired") ||
        err.message.includes("Authentication required") ||
        err.message.includes("please log in again") ||
        err.message.includes("User not authenticated")
      ) {
        setTimeout(() => {
          navigate("/login");
        }, 2000);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value,
    }));

    // Clear edit error when user makes changes
    if (editError) {
      setEditError("");
    }
  };

  const handleSaveChanges = async (e) => {
    e.preventDefault();

    try {
      setEditLoading(true);
      setEditError("");

      const updateData = {
        email: formData.email,
        first_name: formData.first_name,
        last_name: formData.last_name,
        phone: formData.phone,
        country: formData.country,
        region: formData.region,
        timezone: formData.timezone,
        agree_promotions: formData.agree_promotions,
        agree_to_tracking_across_third_party_apps_and_services:
          formData.agree_to_tracking_across_third_party_apps_and_services,
      };

      const updatedUser = await meService.updateCurrentUser(updateData);
      setUser(updatedUser);
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
    // Reset form data to original user data
    if (user) {
      setFormData({
        email: user.email || "",
        first_name: user.first_name || "",
        last_name: user.last_name || "",
        phone: user.phone || "",
        country: user.country || "Canada",
        region: user.region || "",
        timezone: user.timezone || "America/Toronto",
        agree_promotions: user.agree_promotions || false,
        agree_to_tracking_across_third_party_apps_and_services:
          user.agree_to_tracking_across_third_party_apps_and_services || false,
      });
    }
    setIsEditing(false);
    setEditError("");
  };

  const handleDeleteAccount = async (e) => {
    e.preventDefault();

    if (!deletePassword) {
      setDeleteError("Password is required to delete your account");
      return;
    }

    if (
      !window.confirm(
        "Are you sure you want to delete your account? This action cannot be undone.",
      )
    ) {
      return;
    }

    try {
      setDeleteLoading(true);
      setDeleteError("");

      // Use the MeService to delete the account
      await meService.deleteCurrentUser(deletePassword);

      // Account deleted successfully, logout and redirect
      logout();
      navigate("/");
    } catch (err) {
      console.error("Failed to delete account:", err);
      setDeleteError(err.message || "Failed to delete account");
    } finally {
      setDeleteLoading(false);
    }
  };

  const getUserRoleText = (role) => {
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

  const getUserStatusText = (status) => {
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
      <div>
        <h2>Loading Profile...</h2>
        <p>Please wait while we load your profile information.</p>
      </div>
    );
  }

  if (error) {
    return (
      <div>
        <h2>Error Loading Profile</h2>
        <p>{error}</p>
        <button onClick={loadUserData}>Try Again</button>
        <button onClick={() => navigate("/dashboard")}>
          Back to Dashboard
        </button>
      </div>
    );
  }

  if (!user) {
    return (
      <div>
        <h2>Profile Not Found</h2>
        <p>Unable to load your profile information.</p>
        <button onClick={() => navigate("/dashboard")}>
          Back to Dashboard
        </button>
      </div>
    );
  }

  return (
    <div>
      <h1>My Profile</h1>

      {/* Navigation */}
      <div>
        <button onClick={() => navigate("/dashboard")}>
          ← Back to Dashboard
        </button>
        <button onClick={loadUserData}>🔄 Refresh</button>
      </div>

      {/* Profile Information */}
      <div>
        <h2>Profile Information</h2>

        {!isEditing ? (
          /* View Mode */
          <div>
            <div>
              <h3>Basic Information</h3>
              <p>
                <strong>Name:</strong> {user.name}
              </p>
              <p>
                <strong>Email:</strong> {user.email}
              </p>
              <p>
                <strong>Phone:</strong> {user.phone || "Not provided"}
              </p>
              <p>
                <strong>Role:</strong> {getUserRoleText(user.role)}
              </p>
              <p>
                <strong>Status:</strong> {getUserStatusText(user.status)}
              </p>
              <p>
                <strong>Member Since:</strong>{" "}
                {new Date(user.created_at).toLocaleDateString()}
              </p>
            </div>

            <div>
              <h3>Location Information</h3>
              <p>
                <strong>Country:</strong> {user.country || "Not provided"}
              </p>
              <p>
                <strong>Region:</strong> {user.region || "Not provided"}
              </p>
              <p>
                <strong>City:</strong> {user.city || "Not provided"}
              </p>
              <p>
                <strong>Postal Code:</strong>{" "}
                {user.postal_code || "Not provided"}
              </p>
              <p>
                <strong>Address Line 1:</strong>{" "}
                {user.address_line1 || "Not provided"}
              </p>
              <p>
                <strong>Address Line 2:</strong>{" "}
                {user.address_line2 || "Not provided"}
              </p>
              <p>
                <strong>Timezone:</strong> {user.timezone}
              </p>
            </div>

            <div>
              <h3>Preferences</h3>
              <p>
                <strong>Promotional Communications:</strong>{" "}
                {user.agree_promotions ? "Yes" : "No"}
              </p>
              <p>
                <strong>Third-party Tracking:</strong>{" "}
                {user.agree_to_tracking_across_third_party_apps_and_services
                  ? "Yes"
                  : "No"}
              </p>
            </div>

            <div>
              <button onClick={() => setIsEditing(true)}>
                ✏️ Edit Profile
              </button>
            </div>
          </div>
        ) : (
          /* Edit Mode */
          <div>
            {editError && (
              <div>
                <p>{editError}</p>
              </div>
            )}

            <form onSubmit={handleSaveChanges}>
              <div>
                <h3>Basic Information</h3>

                <div>
                  <label htmlFor="first_name">First Name *</label>
                  <input
                    type="text"
                    id="first_name"
                    name="first_name"
                    value={formData.first_name}
                    onChange={handleInputChange}
                    required
                    disabled={editLoading}
                  />
                </div>

                <div>
                  <label htmlFor="last_name">Last Name *</label>
                  <input
                    type="text"
                    id="last_name"
                    name="last_name"
                    value={formData.last_name}
                    onChange={handleInputChange}
                    required
                    disabled={editLoading}
                  />
                </div>

                <div>
                  <label htmlFor="email">Email *</label>
                  <input
                    type="email"
                    id="email"
                    name="email"
                    value={formData.email}
                    onChange={handleInputChange}
                    required
                    disabled={editLoading}
                  />
                </div>

                <div>
                  <label htmlFor="phone">Phone *</label>
                  <input
                    type="tel"
                    id="phone"
                    name="phone"
                    value={formData.phone}
                    onChange={handleInputChange}
                    required
                    disabled={editLoading}
                  />
                </div>
              </div>

              <div>
                <h3>Location</h3>

                <div>
                  <label htmlFor="country">Country *</label>
                  <select
                    id="country"
                    name="country"
                    value={formData.country}
                    onChange={handleInputChange}
                    required
                    disabled={editLoading}
                  >
                    {countries.map((country) => (
                      <option key={country} value={country}>
                        {country}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label htmlFor="region">Region</label>
                  <input
                    type="text"
                    id="region"
                    name="region"
                    value={formData.region}
                    onChange={handleInputChange}
                    disabled={editLoading}
                  />
                </div>

                <div>
                  <label htmlFor="timezone">Timezone *</label>
                  <select
                    id="timezone"
                    name="timezone"
                    value={formData.timezone}
                    onChange={handleInputChange}
                    required
                    disabled={editLoading}
                  >
                    {timezones.map((tz) => (
                      <option key={tz} value={tz}>
                        {tz}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div>
                <h3>Preferences</h3>

                <div>
                  <input
                    type="checkbox"
                    id="agree_promotions"
                    name="agree_promotions"
                    checked={formData.agree_promotions}
                    onChange={handleInputChange}
                    disabled={editLoading}
                  />
                  <label htmlFor="agree_promotions">
                    I agree to receive promotional communications
                  </label>
                </div>

                <div>
                  <input
                    type="checkbox"
                    id="agree_to_tracking_across_third_party_apps_and_services"
                    name="agree_to_tracking_across_third_party_apps_and_services"
                    checked={
                      formData.agree_to_tracking_across_third_party_apps_and_services
                    }
                    onChange={handleInputChange}
                    disabled={editLoading}
                  />
                  <label htmlFor="agree_to_tracking_across_third_party_apps_and_services">
                    I agree to tracking across third-party apps and services
                  </label>
                </div>
              </div>

              <div>
                <button type="submit" disabled={editLoading}>
                  {editLoading ? "Saving..." : "💾 Save Changes"}
                </button>

                <button
                  type="button"
                  onClick={handleCancelEdit}
                  disabled={editLoading}
                >
                  ❌ Cancel
                </button>
              </div>
            </form>
          </div>
        )}
      </div>

      {/* Account Actions */}
      <div>
        <h2>Account Actions</h2>

        <div>
          <button onClick={() => navigate("/dashboard")}>🏠 Dashboard</button>

          <button onClick={logout}>🚪 Logout</button>
        </div>
      </div>

      {/* Danger Zone - Account Deletion */}
      <div>
        <h2>🚨 Danger Zone</h2>

        <div>
          <h3>Delete Account</h3>
          <p>
            <strong>Warning:</strong> Deleting your account is permanent and
            cannot be undone. All your data will be permanently removed from our
            servers.
          </p>

          {!showDeleteSection ? (
            <button onClick={() => setShowDeleteSection(true)}>
              🗑️ Delete My Account
            </button>
          ) : (
            <div>
              {deleteError && (
                <div>
                  <p>{deleteError}</p>
                </div>
              )}

              <form onSubmit={handleDeleteAccount}>
                <div>
                  <label htmlFor="delete_password">
                    Enter your password to confirm deletion *
                  </label>
                  <input
                    type="password"
                    id="delete_password"
                    value={deletePassword}
                    onChange={(e) => setDeletePassword(e.target.value)}
                    placeholder="Enter your password"
                    required
                    disabled={deleteLoading}
                  />
                </div>

                <div>
                  <button
                    type="submit"
                    disabled={deleteLoading || !deletePassword}
                  >
                    {deleteLoading
                      ? "Deleting..."
                      : "🗑️ Permanently Delete Account"}
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      setShowDeleteSection(false);
                      setDeletePassword("");
                      setDeleteError("");
                    }}
                    disabled={deleteLoading}
                  >
                    Cancel
                  </button>
                </div>
              </form>

              <div>
                <h4>What happens when you delete your account:</h4>
                <ul>
                  <li>
                    Your profile and account information will be permanently
                    deleted
                  </li>
                  <li>All your files and data will be permanently removed</li>
                  <li>You will be immediately logged out</li>
                  <li>This action cannot be undone</li>
                  <li>
                    Your email address will become available for new
                    registrations
                  </li>
                </ul>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Debug Information (in development only) */}
      {import.meta.env.DEV && (
        <details>
          <summary>🔍 Debug Information</summary>
          <div>
            <h4>User Object:</h4>
            <pre>{JSON.stringify(user, null, 2)}</pre>

            <h4>Form Data:</h4>
            <pre>{JSON.stringify(formData, null, 2)}</pre>

            <h4>Service Info:</h4>
            <pre>{JSON.stringify(meService.getDebugInfo(), null, 2)}</pre>
          </div>
        </details>
      )}
    </div>
  );
};

const ProtectedMeDetail = withPasswordProtection(MeDetail);
export default ProtectedMeDetail;
