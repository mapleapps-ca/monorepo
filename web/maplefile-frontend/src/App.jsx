// src/App.js
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import { ServiceProvider } from "./contexts/ServiceContext";
import Navigation from "./components/Navigation";
import Login from "./components/Login";
import Register from "./components/Register";
import Profile from "./components/Profile";
import EmailVerification from "./components/EmailVerification";
import RegistrationSuccess from "./components/RegistrationSuccess";

// Home component - simple landing page
const Home = () => {
  return (
    <div style={styles.home}>
      <h1>Welcome to MapleFile</h1>
      <p>Secure file storage and sharing with end-to-end encryption.</p>
      <p>Please login or register to access your account.</p>
    </div>
  );
};

// Main App component
function App() {
  return (
    // Wrap entire app with ServiceProvider for dependency injection
    <ServiceProvider>
      <Router>
        <div style={styles.app}>
          {/* Navigation will be shown on all pages */}
          <Navigation />

          {/* Define all routes */}
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/verify-email" element={<EmailVerification />} />
            <Route
              path="/registration-success"
              element={<RegistrationSuccess />}
            />
            <Route path="/profile" element={<Profile />} />
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
