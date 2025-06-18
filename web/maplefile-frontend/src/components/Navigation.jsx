// src/components/Navigation.jsx
import React from "react";
import { Link, useNavigate } from "react-router";
import { useServices } from "../contexts/ServiceContext";

const Navigation = () => {
  const { authService } = useServices();
  const navigate = useNavigate();
  const isAuthenticated = authService.isAuthenticated();
  const currentUser = authService.getCurrentUser();

  const handleLogout = () => {
    authService.logout();
    navigate("/login");
    // Force re-render by updating the component
    window.location.reload();
  };

  return (
    <nav style={styles.nav}>
      <div style={styles.navContent}>
        <Link to="/" style={styles.logo}>
          My Auth App
        </Link>

        <div style={styles.links}>
          {isAuthenticated ? (
            <>
              <span style={styles.welcome}>Welcome, {currentUser.name}!</span>
              <Link to="/profile" style={styles.link}>
                Profile
              </Link>
              <button onClick={handleLogout} style={styles.logoutButton}>
                Logout
              </button>
            </>
          ) : (
            <>
              <Link to="/" style={styles.logo}>
                MapleFile
              </Link>
              <Link to="/login" style={styles.link}>
                Login
              </Link>
              <Link to="/register" style={styles.link}>
                Register
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

const styles = {
  nav: {
    backgroundColor: "#333",
    padding: "1rem 0",
    marginBottom: "2rem",
  },
  navContent: {
    maxWidth: "1200px",
    margin: "0 auto",
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    padding: "0 1rem",
  },
  logo: {
    color: "white",
    textDecoration: "none",
    fontSize: "1.5rem",
    fontWeight: "bold",
  },
  links: {
    display: "flex",
    alignItems: "center",
    gap: "1rem",
  },
  link: {
    color: "white",
    textDecoration: "none",
    padding: "0.5rem 1rem",
    borderRadius: "4px",
    transition: "background-color 0.3s",
  },
  welcome: {
    color: "white",
    marginRight: "1rem",
  },
  logoutButton: {
    backgroundColor: "#dc3545",
    color: "white",
    border: "none",
    padding: "0.5rem 1rem",
    borderRadius: "4px",
    cursor: "pointer",
    fontSize: "1rem",
  },
};

export default Navigation;
