// monorepo/web/prototype/maplefile-register/src/App.jsx
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import Register from "./pages/Register";
import RecoveryCode from "./pages/RecoveryCode";
import VerifyEmail from "./pages/VerifyEmail";
import VerificationSuccess from "./pages/VerificationSuccess";

function App() {
  return (
    <Router>
      <div className="container">
        <h1>MapleApps Registration</h1>

        <Routes>
          <Route path="/" element={<Navigate to="/register" replace />} />
          <Route path="/register" element={<Register />} />
          <Route path="/register/recovery" element={<RecoveryCode />} />
          <Route path="/verify-email" element={<VerifyEmail />} />
          <Route
            path="/verify-email/success"
            element={<VerificationSuccess />}
          />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
