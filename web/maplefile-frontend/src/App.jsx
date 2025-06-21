// monorepo/web/maplefile-frontend/src/App.jsx
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import { ServiceProvider } from "./contexts/ServiceContext";
// import Navigation from "./components/Navigation";
import IndexPage from "./pages/Anonymous/Index/IndexPage";

// Registration pages
import Register from "./pages/Anonymous/Register/Register";
import RecoveryCode from "./pages/Anonymous/Register/RecoveryCode";
import VerifyEmail from "./pages/Anonymous/Register/VerifyEmail";
import VerifySuccess from "./pages/Anonymous/Register/VerifySuccess";

// Login pages
import RequestOTT from "./pages/Anonymous/Login/RequestOTT";
import VerifyOTT from "./pages/Anonymous/Login/VerifyOTT";
import CompleteLogin from "./pages/Anonymous/Login/CompleteLogin";

// User pages
import Dashboard from "./pages/User/Dashboard/Dashboard";
import MeDetail from "./pages/User/Me/Detail";

// import Profile from "./components/Profile";
// import EmailVerification from "./components/EmailVerification";
// import RegistrationSuccess from "./components/RegistrationSuccess";
// import ProtectedRoute from "./components/ProtectedRoute";

// Main App component
function App() {
  return (
    // Wrap entire app with ServiceProvider for dependency injection
    <ServiceProvider>
      <Router>
        <div style={styles.app}>
          {/* Navigation will be shown on all pages */}
          {/*<Navigation />*/}

          {/* Define all routes */}
          <Routes>
            <Route path="/" element={<IndexPage />} />

            {/* Registration routes */}
            <Route path="/register" element={<Register />} />
            <Route path="/register/recovery" element={<RecoveryCode />} />
            <Route path="/register/verify-email" element={<VerifyEmail />} />
            <Route
              path="/register/verify-success"
              element={<VerifySuccess />}
            />

            {/* Login routes */}
            <Route path="/login" element={<RequestOTT />} />
            <Route path="/login/request-ott" element={<RequestOTT />} />
            <Route path="/login/verify-ott" element={<VerifyOTT />} />
            <Route path="/login/complete" element={<CompleteLogin />} />

            {/* User routes */}
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/me" element={<MeDetail />} />
            <Route path="/profile" element={<MeDetail />} />

            {/*
            <Route path="/verify-email" element={<EmailVerification />} />
            <Route
              path="/registration-success"
              element={<RegistrationSuccess />}
            />
            */}

            {/* Protected routes */}
            {/*
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <Profile />
                </ProtectedRoute>
              }
            />
            */}

            {/* Redirect any unknown routes to home */}
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </div>
      </Router>
    </ServiceProvider>
  );
}

const styles = {
  app: {
    minHeight: "100vh",
    backgroundColor: "#f5f5f5",
  },
  home: {
    textAlign: "center",
    padding: "2rem",
    maxWidth: "800px",
    margin: "0 auto",
  },
};

export default App;
