// File: src/App.jsx
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import { ServiceProvider } from "./services/Services";

// Front-facing pages
import IndexPage from "./pages/Anonymous/Index/IndexPage"; //TODO
import DeveloperIndexPage from "./pages/Developer/Index/IndexPage";

// Registration pages
import Register from "./pages/Anonymous/Register/Register";
import RecoveryCode from "./pages/Anonymous/Register/RecoveryCode";
import VerifyEmail from "./pages/Anonymous/Register/VerifyEmail";
import VerifySuccess from "./pages/Anonymous/Register/VerifySuccess";
import DeveloperRegister from "./pages/Developer/Register/Register";
import DeveloperRecoveryCode from "./pages/Developer/Register/RecoveryCode";
import DeveloperVerifyEmail from "./pages/Developer/Register/VerifyEmail";
import DeveloperVerifySuccess from "./pages/Developer/Register/VerifySuccess";

// Login pages
import RequestOTT from "./pages/Anonymous/Login/RequestOTT";
import VerifyOTT from "./pages/Anonymous/Login/VerifyOTT";
import CompleteLogin from "./pages/Anonymous/Login/CompleteLogin";
import DeveloperRequestOTT from "./pages/Developer/Login/RequestOTT";
import DeveloperVerifyOTT from "./pages/Developer/Login/VerifyOTT";
import DeveloperCompleteLogin from "./pages/Developer/Login/CompleteLogin";

// Recovery pages
import InitiateRecovery from "./pages/Anonymous/Recovery/InitiateRecovery";
import VerifyRecovery from "./pages/Anonymous/Recovery/VerifyRecovery";
import CompleteRecovery from "./pages/Anonymous/Recovery/CompleteRecovery";
import DeveloperInitiateRecovery from "./pages/Developer/Recovery/InitiateRecovery";
import DeveloperVerifyRecovery from "./pages/Developer/Recovery/VerifyRecovery";
import DeveloperCompleteRecovery from "./pages/Developer/Recovery/CompleteRecovery";

// User pages
import Dashboard from "./pages/User/Dashboard/Dashboard";
import DeveloperDashboard from "./pages/Developer/Dashboard/Dashboard";
import MeDetail from "./pages/User/Me/Detail";
import DeveloperMeDetail from "./pages/Developer/Me/Detail";

// Example Pages
import TokenManagerExample from "./pages/Developer/Examples/TokenManagerExample";
import DashboardManagerExample from "./pages/Developer/Examples/DashboardManagerExample";
import SyncCollectionAPIExample from "./pages/Developer/Examples/SyncCollectionAPIExample";
import SyncCollectionStorageExample from "./pages/Developer/Examples/SyncCollectionStorageExample";
import SyncCollectionManagerExample from "./pages/Developer/Examples/SyncCollectionManagerExample.jsx";

import SyncFileAPIExample from "./pages/Developer/Examples/SyncFileAPIExample";
import SyncFileStorageExample from "./pages/Developer/Examples/SyncFileStorageExample";
import SyncFileManagerExample from "./pages/Developer/Examples/SyncFileManagerExample.jsx";

import CreateCollectionManagerExample from "./pages/Developer/Examples/Collection/CreateCollectionManagerExample.jsx";
import GetCollectionManagerExample from "./pages/Developer/Examples/Collection/GetCollectionManagerExample.jsx";
import UpdateCollectionManagerExample from "./pages/Developer/Examples/Collection/UpdateCollectionManagerExample.jsx";
import DeleteCollectionManagerExample from "./pages/Developer/Examples/Collection/DeleteCollectionManagerExample.jsx";
import ListCollectionManagerExample from "./pages/Developer/Examples/Collection/ListCollectionManagerExample.jsx";
import ShareCollectionManagerExample from "./pages/Developer/Examples/Collection/ShareCollectionManagerExample.jsx";
import UserLookupExample from "./pages/Developer/Examples/User/UserLookupExample.jsx";

import CreateFileManagerExample from "./pages/Developer/Examples/File/CreateFileManagerExample.jsx";
import GetFileManagerExample from "./pages/Developer/Examples/File/GetFileManagerExample.jsx";
import DownloadFileManagerExample from "./pages/Developer/Examples/File/DownloadFileManagerExample.jsx";
import DeleteFileManagerExample from "./pages/Developer/Examples/File/DeleteFileManagerExample.jsx";
import ListFileManagerExample from "./pages/Developer/Examples/File/ListFileManagerExample.jsx";
import RecentFileManagerExample from "./pages/Developer/Examples/File/RecentFileManagerExample.jsx";

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
            {/* Front-facing pages */}
            <Route path="/" element={<IndexPage />} />
            <Route path="/developer" element={<DeveloperIndexPage />} />

            {/* Registration routes */}
            <Route path="/developer/register" element={<DeveloperRegister />} />
            <Route
              path="/developer/register/recovery"
              element={<DeveloperRecoveryCode />}
            />
            <Route
              path="/developer/register/verify-email"
              element={<DeveloperVerifyEmail />}
            />
            <Route
              path="/developer/register/verify-success"
              element={<DeveloperVerifySuccess />}
            />

            {/* Login routes */}
            <Route path="/developer/login" element={<DeveloperRequestOTT />} />
            <Route
              path="/developer/login/request-ott"
              element={<DeveloperRequestOTT />}
            />
            <Route
              path="/developer/login/verify-ott"
              element={<DeveloperVerifyOTT />}
            />
            <Route
              path="/developer/login/complete"
              element={<DeveloperCompleteLogin />}
            />
            {/* Recovery routes */}
            <Route
              path="/developer/recovery"
              element={<DeveloperInitiateRecovery />}
            />
            <Route
              path="/developer/recovery/initiate"
              element={<DeveloperInitiateRecovery />}
            />
            <Route
              path="/developer/recovery/verify"
              element={<DeveloperVerifyRecovery />}
            />
            <Route
              path="/developer/recovery/complete"
              element={<DeveloperCompleteRecovery />}
            />

            {/* User routes */}

            <Route path="/dashboard" element={<Dashboard />} />
            <Route
              path="/developer/dashboard"
              element={<DeveloperDashboard />}
            />
            <Route path="/developer/me" element={<DeveloperMeDetail />} />
            <Route path="/me" element={<MeDetail />} />
            <Route path="/developer/profile" element={<DeveloperMeDetail />} />
            <Route path="/profile" element={<MeDetail />} />

            {/* Example routes */}
            <Route
              path="/developer/dashboard-example"
              element={<DashboardManagerExample />}
            />
            <Route
              path="/developer/token-manager-example"
              element={<TokenManagerExample />}
            />
            <Route
              path="/developer/sync-collection-api-example"
              element={<SyncCollectionAPIExample />}
            />
            <Route
              path="/developer/sync-collection-storage-example"
              element={<SyncCollectionStorageExample />}
            />
            <Route
              path="/developer/sync-collection-manager-example"
              element={<SyncCollectionManagerExample />}
            />
            <Route
              path="/developer/create-collection-manager-example"
              element={<CreateCollectionManagerExample />}
            />
            <Route
              path="/developer/get-collection-manager-example"
              element={<GetCollectionManagerExample />}
            />
            <Route
              path="/developer/update-collection-manager-example"
              element={<UpdateCollectionManagerExample />}
            />
            <Route
              path="/developer/delete-collection-manager-example"
              element={<DeleteCollectionManagerExample />}
            />
            <Route
              path="/developer/list-collection-manager-example"
              element={<ListCollectionManagerExample />}
            />
            <Route
              path="/developer/create-file-manager-example"
              element={<CreateFileManagerExample />}
            />
            <Route
              path="/developer/get-file-manager-example"
              element={<GetFileManagerExample />}
            />
            <Route
              path="/developer/download-file-manager-example"
              element={<DownloadFileManagerExample />}
            />
            <Route
              path="/developer/delete-file-manager-example"
              element={<DeleteFileManagerExample />}
            />
            <Route
              path="/developer/list-file-manager-example"
              element={<ListFileManagerExample />}
            />
            <Route
              path="/developer/recent-file-manager-example"
              element={<RecentFileManagerExample />}
            />
            <Route
              path="/developer/sync-file-api-example"
              element={<SyncFileAPIExample />}
            />
            <Route
              path="/developer/sync-file-storage-example"
              element={<SyncFileStorageExample />}
            />
            <Route
              path="/developer/sync-file-manager-example"
              element={<SyncFileManagerExample />}
            />
            <Route
              path="/developer/user-lookup-manager-example"
              element={<UserLookupExample />}
            />
            <Route
              path="/developer/share-collection-manager-example"
              element={<ShareCollectionManagerExample />}
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
