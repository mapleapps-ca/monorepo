// src/services/AuthService.js
class AuthService {
  constructor() {
    // In a real app, this would connect to a backend API
    // For this example, we'll use localStorage to simulate a database
    this.users = JSON.parse(localStorage.getItem("users")) || [];
    this.currentUser = JSON.parse(localStorage.getItem("currentUser")) || null;
  }

  // Register a new user
  register(email, password, name) {
    // Check if user already exists
    const existingUser = this.users.find((user) => user.email === email);
    if (existingUser) {
      throw new Error("User already exists with this email");
    }

    // Create new user
    const newUser = {
      id: Date.now(), // Simple ID generation
      email,
      password, // In real app, this should be hashed!
      name,
      createdAt: new Date().toISOString(),
    };

    // Add to users array
    this.users.push(newUser);

    // Save to localStorage
    localStorage.setItem("users", JSON.stringify(this.users));

    // Log the user in automatically
    this.currentUser = newUser;
    localStorage.setItem("currentUser", JSON.stringify(newUser));

    return newUser;
  }

  // Login an existing user
  login(email, password) {
    // Find user by email and password
    const user = this.users.find(
      (u) => u.email === email && u.password === password,
    );

    if (!user) {
      throw new Error("Invalid email or password");
    }

    // Set current user
    this.currentUser = user;
    localStorage.setItem("currentUser", JSON.stringify(user));

    return user;
  }

  // Logout the current user
  logout() {
    this.currentUser = null;
    localStorage.removeItem("currentUser");
  }

  // Get the current logged in user
  getCurrentUser() {
    return this.currentUser;
  }

  // Check if a user is logged in
  isAuthenticated() {
    return this.currentUser !== null;
  }
}

export default AuthService;
