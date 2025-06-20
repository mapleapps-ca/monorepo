import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import RequestOTT from "./pages/RequestOTT.jsx";
import VerifyOTT from "./pages/VerifyOTT.jsx";
import CompleteLogin from "./pages/CompleteLogin.jsx";
import Dashboard from "./pages/Dashboard.jsx";
import LocalStorageService from "./services/localStorageService.jsx";
import "./styles/App.css";

// Protected Route component
const ProtectedRoute = ({ children }) => {
  const isAuthenticated = LocalStorageService.isAuthenticated();

  return isAuthenticated ? children : <Navigate to="/" replace />;
};

// Redirect authenticated users away from login pages
const LoginRoute = ({ children }) => {
  const isAuthenticated = LocalStorageService.isAuthenticated();

  return isAuthenticated ? <Navigate to="/dashboard" replace /> : children;
};

function App() {
  return (
    <Router>
      <div className="app">
        <Routes>
          {/* Login flow routes */}
          <Route
            path="/"
            element={
              <LoginRoute>
                <RequestOTT />
              </LoginRoute>
            }
          />

          <Route
            path="/verify-ott"
            element={
              <LoginRoute>
                <VerifyOTT />
              </LoginRoute>
            }
          />

          <Route
            path="/complete-login"
            element={
              <LoginRoute>
                <CompleteLogin />
              </LoginRoute>
            }
          />

          {/* Protected routes */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <Dashboard />
              </ProtectedRoute>
            }
          />

          {/* Catch all route - redirect to home */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
