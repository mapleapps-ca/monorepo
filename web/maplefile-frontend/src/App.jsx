// File: src/App.jsx
import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import { ServiceProvider } from "./services/Services"; // NEW: Single service import
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

// Recovery pages
import InitiateRecovery from "./pages/Anonymous/Recovery/InitiateRecovery";
import VerifyRecovery from "./pages/Anonymous/Recovery/VerifyRecovery";
import CompleteRecovery from "./pages/Anonymous/Recovery/CompleteRecovery";

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
import GetCollectionManagerExample from "./pages/User/Examples/Collection/GetCollectionManagerExample.jsx";
import UpdateCollectionManagerExample from "./pages/User/Examples/Collection/UpdateCollectionManagerExample.jsx";
import DeleteCollectionManagerExample from "./pages/User/Examples/Collection/DeleteCollectionManagerExample.jsx";
import ListCollectionManagerExample from "./pages/User/Examples/Collection/ListCollectionManagerExample.jsx";
import ShareCollectionManagerExample from "./pages/User/Examples/Collection/ShareCollectionManagerExample.jsx";
import UserLookupExample from "./pages/User/Examples/User/UserLookupExample.jsx";

import CreateFileManagerExample from "./pages/User/Examples/File/CreateFileManagerExample.jsx";
import GetFileManagerExample from "./pages/User/Examples/File/GetFileManagerExample.jsx";
import DownloadFileManagerExample from "./pages/User/Examples/File/DownloadFileManagerExample.jsx";
import DeleteFileManagerExample from "./pages/User/Examples/File/DeleteFileManagerExample.jsx";
import ListFileManagerExample from "./pages/User/Examples/File/ListFileManagerExample.jsx";

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
            {/* Recovery routes */}
            <Route path="/recovery" element={<InitiateRecovery />} />
            <Route path="/recovery/initiate" element={<InitiateRecovery />} />
            <Route path="/recovery/verify" element={<VerifyRecovery />} />
            <Route path="/recovery/complete" element={<CompleteRecovery />} />
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
              path="/get-collection-manager-example"
              element={<GetCollectionManagerExample />}
            />
            <Route
              path="/update-collection-manager-example"
              element={<UpdateCollectionManagerExample />}
            />
            <Route
              path="/delete-collection-manager-example"
              element={<DeleteCollectionManagerExample />}
            />
            <Route
              path="/list-collection-manager-example"
              element={<ListCollectionManagerExample />}
            />
            <Route
              path="/create-file-manager-example"
              element={<CreateFileManagerExample />}
            />
            <Route
              path="/get-file-manager-example"
              element={<GetFileManagerExample />}
            />
            <Route
              path="/download-file-manager-example"
              element={<DownloadFileManagerExample />}
            />
            <Route
              path="/delete-file-manager-example"
              element={<DeleteFileManagerExample />}
            />
            <Route
              path="/list-file-manager-example"
              element={<ListFileManagerExample />}
            />
            <Route
              path="/sync-file-api-example"
              element={<SyncFileAPIExample />}
            />
            <Route
              path="/sync-file-storage-example"
              element={<SyncFileStorageExample />}
            />
            <Route
              path="/sync-file-manager-example"
              element={<SyncFileManagerExample />}
            />
            <Route
              path="/user-lookup-manager-example"
              element={<UserLookupExample />}
            />
            <Route
              path="/share-collection-manager-example"
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
