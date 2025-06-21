// monorepo/web/prototype/maplefile-login/src/components/Layout.jsx
import React from "react";

const Layout = ({ children, title, subtitle }) => {
  return (
    <div className="layout">
      <header className="header">
        <h1>MapleApps</h1>
        <p>Secure Login System</p>
      </header>

      <main className="main">
        <div className="container">
          {title && <h2 className="page-title">{title}</h2>}
          {subtitle && <p className="page-subtitle">{subtitle}</p>}
          {children}
        </div>
      </main>

      <footer className="footer">
        <p>&copy; 2025 MapleApps. All rights reserved.</p>
      </footer>
    </div>
  );
};

export default Layout;
