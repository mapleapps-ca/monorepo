// monorepo/web/prototyping/maplefile-cli/src/App.jsx
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  Navigate,
} from "react-router";
import IndexPage from "./pages/anonymous/Index/page";
import RegisterPage from "./pages/anonymous/Register/page";
import RequestOTTPage from "./pages/anonymous/Login/requestOTTPage";
import VerifyOTTPage from "./pages/anonymous/Login/verifyOTTPage";
import CompleteLoginPage from "./pages/anonymous/Login/completeLoginPage";
import UserDashboardPage from "./pages/user/Dashboard/page";

// // Protected route component
// function ProtectedRoute({ children }) {
//   const { isAuthenticated, isLoading } = useAuth();

//   if (isLoading) {
//     return <div>Loading...</div>;
//   }

//   if (!isAuthenticated) {
//     return <Navigate to="/login" replace />;
//   }

//   return children;
// }

// Navigation with authentication status
function Navigation() {
  return <>test</>;
  // const { isAuthenticated, logout } = useAuth();

  // return (
  //   <nav>
  //     <ul>
  //       <li>
  //         <Link to="/">Home</Link>
  //       </li>
  //       {!isAuthenticated ? (
  //         <>
  //           <li>
  //             <Link to="/register">Register</Link>
  //           </li>
  //           <li>
  //             <Link to="/login">Login</Link>
  //           </li>
  //         </>
  //       ) : (
  //         <>
  //           <li>
  //             <Link to="/collections">Collections</Link>{" "}
  //           </li>
  //           <li>
  //             <Link to="/profile">Profile</Link>{" "}
  //           </li>
  //           <li>
  //             <button onClick={logout}>Logout</button>
  //           </li>
  //         </>
  //       )}
  //     </ul>
  //   </nav>
  // );
}

// Main App component
function AppContent() {
  // const { isLoading } = useAuth();

  // if (isLoading) {
  //   return <div>Loading authentication...</div>;
  // }

  return (
    <div>
      <Navigation />

      <Routes>
        <Route
          path="/"
          element={
            // <ProtectedRoute>
            <IndexPage />
            // </ProtectedRoute>
          }
        />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/login" element={<RequestOTTPage />} />
        <Route path="/verify-ott" element={<VerifyOTTPage />} />
        <Route path="/complete-login" element={<CompleteLoginPage />} />

        <Route path="/dashboard" element={<UserDashboardPage />} />
        {/*
        <Route
          path="/profile"
          element={
            <ProtectedRoute>
              <Profile />
            </ProtectedRoute>
          }
        />
        <Route
          path="/collections"
          element={
            <ProtectedRoute>
              <CollectionListPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/collections/:collectionId/files"
          element={
            <ProtectedRoute>
              <CollectionFileListPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="/collections/:collectionId/upload"
          element={
            <ProtectedRoute>
              <FileUploadPage />
            </ProtectedRoute>
          }
        />

        */}
      </Routes>
    </div>
  );
}

// Wrap everything with the auth provider
function App() {
  return (
    <Router>
      {/* <AuthProvider> */}
      <AppContent />
      {/* </AuthProvider> */}
    </Router>
  );
}

export default App;
