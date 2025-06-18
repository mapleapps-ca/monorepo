// src/services/MeService.js
class MeService {
  constructor(authService) {
    // MeService depends on AuthService to get the current user
    this.authService = authService;
  }

  // Get the current user's profile
  getProfile() {
    const currentUser = this.authService.getCurrentUser();
    if (!currentUser) {
      throw new Error("No user logged in");
    }
    return currentUser;
  }

  // Update the current user's profile
  updateProfile(updates) {
    const currentUser = this.authService.getCurrentUser();
    if (!currentUser) {
      throw new Error("No user logged in");
    }

    // Validate updates
    if (updates.name !== undefined && !updates.name.trim()) {
      throw new Error("Name cannot be empty");
    }

    // Get all users from localStorage
    const users = JSON.parse(localStorage.getItem("users")) || [];

    // Find and update the current user
    const userIndex = users.findIndex((u) => u.id === currentUser.id);
    if (userIndex === -1) {
      throw new Error("User not found");
    }

    // Merge updates with current user data
    const updatedUser = {
      ...currentUser,
      ...updates,
      updatedAt: new Date().toISOString(),
    };

    // Update in users array
    users[userIndex] = updatedUser;

    // Save to localStorage
    localStorage.setItem("users", JSON.stringify(users));
    localStorage.setItem("currentUser", JSON.stringify(updatedUser));

    // Update the auth service's current user reference
    this.authService.currentUser = updatedUser;

    return updatedUser;
  }

  // Get user statistics
  getUserStats() {
    const profile = this.getProfile();

    return {
      userId: profile.id,
      accountType: "Standard User",
      status: "Active",
      memberSince: new Date(profile.createdAt).toLocaleDateString(),
      lastUpdated: profile.updatedAt
        ? new Date(profile.updatedAt).toLocaleDateString()
        : "Never",
    };
  }

  // Update specific profile fields
  updateName(name) {
    return this.updateProfile({ name });
  }

  // Add additional profile fields in the future
  updateEmail(email) {
    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      throw new Error("Invalid email format");
    }

    // Check if email is already taken by another user
    const users = JSON.parse(localStorage.getItem("users")) || [];
    const currentUser = this.authService.getCurrentUser();
    const emailTaken = users.some(
      (u) => u.email === email && u.id !== currentUser.id,
    );

    if (emailTaken) {
      throw new Error("Email is already in use");
    }

    return this.updateProfile({ email });
  }

  // Future method: Update avatar
  updateAvatar(avatarUrl) {
    return this.updateProfile({ avatar: avatarUrl });
  }

  // Future method: Update preferences
  updatePreferences(preferences) {
    const currentUser = this.getProfile();
    const updatedPreferences = {
      ...currentUser.preferences,
      ...preferences,
    };
    return this.updateProfile({ preferences: updatedPreferences });
  }

  // Delete account (soft delete)
  deleteAccount() {
    const currentUser = this.authService.getCurrentUser();
    if (!currentUser) {
      throw new Error("No user logged in");
    }

    // Mark account as deleted
    const deletedUser = this.updateProfile({
      status: "deleted",
      deletedAt: new Date().toISOString(),
    });

    // Log the user out
    this.authService.logout();

    return deletedUser;
  }
}

export default MeService;
