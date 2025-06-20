import Navigation from "./Navigation";

/**
 * Layout component that wraps all pages
 * Follows Single Responsibility Principle - only handles page layout
 */
const Layout = ({ children }) => {
  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {children}
      </main>
    </div>
  );
};

export default Layout;
