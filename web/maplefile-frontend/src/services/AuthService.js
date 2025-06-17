// This service handles all authentication-related operations
// Notice how it receives a logger instead of creating one itself
export class AuthService {
  constructor(logger) {
    this.logger = logger; // InversifyJS will provide this
    this.isAuthenticated = false;
    this.user = null;
  }

  // Simulate requesting a one-time token
  async requestOTT(email) {
    this.logger.log(`AuthService: Requesting OTT for ${email}`);

    // Here you would normally make an API call
    // For now, we'll just simulate it
    await new Promise((resolve) => setTimeout(resolve, 1000));

    this.logger.log("AuthService: OTT request sent successfully");
    return { success: true, message: "OTT sent to your email" };
  }

  // Simulate verifying the OTT
  async verifyOTT(email, token) {
    this.logger.log(`AuthService: Verifying OTT for ${email}`);

    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // For demo purposes, any 6-digit token works
    if (token && token.length === 6) {
      this.logger.log("AuthService: OTT verified successfully");
      return { success: true, sessionToken: "demo-session-token" };
    } else {
      this.logger.log("AuthService: OTT verification failed");
      return { success: false, message: "Invalid token" };
    }
  }

  // Complete the login process
  async completeLogin(sessionToken) {
    this.logger.log("AuthService: Completing login");

    // Simulate final login step
    await new Promise((resolve) => setTimeout(resolve, 500));

    this.isAuthenticated = true;
    this.user = { email: "demo@example.com", name: "Demo User" };

    this.logger.log("AuthService: Login completed successfully");
    return { success: true, user: this.user };
  }

  // Check if user is currently authenticated
  isUserAuthenticated() {
    return this.isAuthenticated;
  }

  // Get current user info
  getCurrentUser() {
    return this.user;
  }

  // Log out the user
  logout() {
    this.logger.log("AuthService: User logging out");
    this.isAuthenticated = false;
    this.user = null;
  }
}
