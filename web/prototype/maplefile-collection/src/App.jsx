import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router";
import { ServiceProvider } from "./contexts/ServiceContext";
import Layout from "./components/layout/Layout";
import DevTools from "./components/dev/DevTools";
import CollectionsPage from "./pages/CollectionsPage";
import SharedCollectionsPage from "./pages/SharedCollectionsPage";
import CreateCollectionPage from "./pages/CreateCollectionPage";
import EditCollectionPage from "./pages/EditCollectionPage";
import CollectionDetailPage from "./pages/CollectionDetailPage";
import { ROUTES } from "./constants";

/**
 * Main App component with routing setup
 * Follows dependency injection pattern using ServiceProvider
 */
function App() {
  // Debug log to verify we're in development mode
  console.log(
    "App running in:",
    import.meta.env.DEV ? "DEVELOPMENT" : "PRODUCTION",
  );

  return (
    <ServiceProvider>
      <Router>
        <Layout>
          <Routes>
            {/* Redirect root to collections */}
            <Route
              path="/"
              element={<Navigate to={ROUTES.COLLECTIONS} replace />}
            />

            {/* Collections routes */}
            <Route path={ROUTES.COLLECTIONS} element={<CollectionsPage />} />
            <Route
              path={ROUTES.CREATE_COLLECTION}
              element={<CreateCollectionPage />}
            />
            <Route
              path={ROUTES.COLLECTION_DETAIL}
              element={<CollectionDetailPage />}
            />
            <Route
              path={ROUTES.EDIT_COLLECTION}
              element={<EditCollectionPage />}
            />

            {/* Shared collections */}
            <Route
              path={ROUTES.SHARED_COLLECTIONS}
              element={<SharedCollectionsPage />}
            />

            {/* 404 fallback */}
            <Route
              path="*"
              element={
                <div className="text-center py-12">
                  <h1 className="text-2xl font-bold text-gray-900 mb-4">
                    Page Not Found
                  </h1>
                  <p className="text-gray-600 mb-6">
                    The page you're looking for doesn't exist.
                  </p>
                  <a
                    href="/collections"
                    className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
                  >
                    Go to Collections
                  </a>
                </div>
              }
            />
          </Routes>
        </Layout>

        {/* Development tools - only shown in development */}
        {import.meta.env.DEV && (
          <div>
            {console.log("Rendering DevTools component...")}
            <DevTools />
          </div>
        )}
      </Router>
    </ServiceProvider>
  );
}

export default App;
