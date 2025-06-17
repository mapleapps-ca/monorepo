// src/App.jsx

import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";

// Import your existing pages
import IndexPage from "./pages/anonymous/Index/page";
import RegisterPage from "./pages/anonymous/Register/page";
import EmailVerificationPage from "./pages/anonymous/VerifyEmail/page";
import RequestOTTPage from "./pages/anonymous/Login/requestOTTPage";
import VerifyOTTPage from "./pages/anonymous/Login/verifyOTTPage";
import CompleteLoginPage from "./pages/anonymous/Login/completeLoginPage";
import UserDashboardPage from "./pages/user/Dashboard/page";

// Import the new providers
import { DIProvider } from "./contexts/DIProvider.jsx";
import { AuthProvider } from "./contexts/AuthContext.jsx";
import { useAuth } from "./contexts/AuthContext.jsx";

// Protected route component (now this actually works!)
function ProtectedRoute({ children }) {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

// Navigation with authentication status
function Navigation() {
  const { isAuthenticated, logout, user } = useAuth();

  return (
    <nav style={{ padding: "1rem", borderBottom: "1px solid #ccc" }}>
      <div style={{ display: "flex", gap: "1rem", alignItems: "center" }}>
        <a href="/">MapleFile</a>

        {!isAuthenticated ? (
          <>
            <a href="/register">Register</a>
            <a href="/login">Login</a>
          </>
        ) : (
          <>
            <a href="/dashboard">Dashboard</a>
            <span>Welcome, {user?.name}!</span>
            <button onClick={logout}>Logout</button>
          </>
        )}
      </div>
    </nav>
  );
}

// Main App component content
function AppContent() {
  const { isLoading } = useAuth();

  if (isLoading) {
    return <div>Loading authentication...</div>;
  }

  return (
    <div>
      <Navigation />
      <main style={{ padding: "2rem" }}>
        <Routes>
          <Route path="/" element={<IndexPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/verify-email" element={<EmailVerificationPage />} />
          <Route path="/login" element={<RequestOTTPage />} />
          <Route path="/verify-ott" element={<VerifyOTTPage />} />
          <Route path="/complete-login" element={<CompleteLoginPage />} />
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <UserDashboardPage />
              </ProtectedRoute>
            }
          />
        </Routes>
      </main>
    </div>
  );
}

// Main App component with all providers
function App() {
  return (
    <Router>
      {/* DIProvider makes dependency injection available everywhere */}
      <DIProvider>
        {/* AuthProvider uses dependency injection to get services */}
        <AuthProvider>
          <AppContent />
        </AuthProvider>
      </DIProvider>
    </Router>
  );
}

export default App;
