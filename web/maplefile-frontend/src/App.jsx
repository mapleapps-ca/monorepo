// File: monorepo/web/maplefile-frontend/src/App.jsx
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import { ServiceProvider } from "./contexts/ServiceContext";
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

// Example Pages
import TokenManagerExample from "./pages/User/Examples/TokenManagerExample";
import SyncCollectionAPIExample from "./pages/User/Examples/SyncCollectionAPIExample";
import SyncCollectionStorageExample from "./pages/User/Examples/SyncCollectionStorageExample";
import SyncCollectionManagerExample from "./pages/User/Examples/SyncCollectionManagerExample.jsx";

import SyncFileAPIExample from "./pages/User/Examples/SyncFileAPIExample";
import SyncFileStorageExample from "./pages/User/Examples/SyncFileStorageExample";
import SyncFileManagerExample from "./pages/User/Examples/SyncFileManagerExample.jsx";

import CreateCollectionManagerExample from "./pages/User/Examples/Collection/CreateCollectionManagerExample.jsx";

// Styles
const styles = {
  app: {
    minHeight: "100vh",
    backgroundColor: "#f5f5f5",
  },
};

// Main App component
function App() {
  return (
    <ServiceProvider>
      <Router>
        <div style={styles.app}>
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

            {/* Example routes */}
            <Route
              path="/token-manager-example"
              element={<TokenManagerExample />}
            />
            <Route
              path="/sync-collection-api-example"
              element={<SyncCollectionAPIExample />}
            />
            <Route
              path="/sync-collection-storage-example"
              element={<SyncCollectionStorageExample />}
            />
            <Route
              path="/sync-collection-manager-example"
              element={<SyncCollectionManagerExample />}
            />
            <Route
              path="/create-collection-manager-example"
              element={<CreateCollectionManagerExample />}
            />

            <Route
              path="/sync-File-api-example"
              element={<SyncFileAPIExample />}
            />
            <Route
              path="/sync-File-storage-example"
              element={<SyncFileStorageExample />}
            />
            <Route
              path="/sync-File-manager-example"
              element={<SyncFileManagerExample />}
            />

            {/* Redirect any unknown routes to home */}
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </div>
      </Router>
    </ServiceProvider>
  );
}

export default App;
