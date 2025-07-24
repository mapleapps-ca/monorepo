// File: src/components/Navigation.jsx
import React, { useState } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import { useAuth } from "../services/Services";
import {
  LockClosedIcon,
  HomeIcon,
  FolderIcon,
  DocumentIcon,
  UserIcon,
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  XMarkIcon,
  ShieldCheckIcon,
  CloudArrowUpIcon,
} from "@heroicons/react/24/outline";

const Navigation = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { authManager } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const handleLogout = () => {
    if (authManager?.logout) {
      authManager.logout();
    }
    sessionStorage.clear();
    localStorage.removeItem("mapleapps_access_token");
    localStorage.removeItem("mapleapps_refresh_token");
    localStorage.removeItem("mapleapps_user_email");
    navigate("/");
  };

  // Updated isActive function to handle FileManager pages
  const isActive = (path) => {
    // For My Files, check if current path starts with /file-manager
    if (path === "/file-manager") {
      return location.pathname.startsWith("/file-manager");
    }
    // For other paths, use exact matching
    return location.pathname === path;
  };

  const mainNavItems = [
    {
      name: "Dashboard",
      path: "/dashboard",
      icon: HomeIcon,
      description: "Overview",
    },
    {
      name: "My Files",
      path: "/file-manager",
      icon: FolderIcon,
      description: "Your files",
    },
  ];

  const userEmail = authManager?.getCurrentUserEmail?.() || "User";

  return (
    <nav className="bg-white border-b border-gray-200 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo */}
          <Link to="/dashboard" className="flex items-center">
            <div className="flex items-center justify-center h-9 w-9 bg-red-800 rounded-lg mr-3">
              <LockClosedIcon className="h-5 w-5 text-white" />
            </div>
            <span className="text-xl font-bold text-gray-900">MapleFile</span>
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden lg:flex items-center space-x-8">
            {mainNavItems.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isActive(item.path)
                    ? "bg-red-800 text-white"
                    : "text-gray-700 hover:text-red-800 hover:bg-red-50"
                }`}
              >
                <item.icon className="h-4 w-4" />
                <span>{item.name}</span>
              </Link>
            ))}
          </div>

          {/* User Menu */}
          <div className="hidden lg:flex items-center space-x-4">
            <span className="text-sm text-gray-700">{userEmail}</span>
            <Link
              to="/me"
              className={`p-2 rounded-md transition-colors ${
                isActive("/me")
                  ? "bg-red-800 text-white"
                  : "text-gray-700 hover:text-red-800 hover:bg-red-50"
              }`}
            >
              <UserIcon className="h-5 w-5" />
            </Link>
            <button
              onClick={handleLogout}
              className="p-2 rounded-md text-gray-700 hover:text-red-800 hover:bg-red-50 transition-colors"
            >
              <ArrowRightOnRectangleIcon className="h-5 w-5" />
            </button>
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className="lg:hidden p-2 rounded-md text-gray-700 hover:text-red-800 hover:bg-red-50"
          >
            {isMobileMenuOpen ? (
              <XMarkIcon className="h-6 w-6" />
            ) : (
              <Bars3Icon className="h-6 w-6" />
            )}
          </button>
        </div>
      </div>

      {/* Mobile Menu */}
      {isMobileMenuOpen && (
        <div className="lg:hidden bg-white border-t border-gray-200">
          <div className="px-2 pt-2 pb-3 space-y-1">
            {mainNavItems.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                onClick={() => setIsMobileMenuOpen(false)}
                className={`flex items-center space-x-3 px-3 py-2 rounded-md text-base font-medium ${
                  isActive(item.path)
                    ? "bg-red-800 text-white"
                    : "text-gray-700 hover:bg-red-50 hover:text-red-800"
                }`}
              >
                <item.icon className="h-5 w-5" />
                <span>{item.name}</span>
              </Link>
            ))}

            <div className="border-t border-gray-200 pt-2 mt-2">
              <div className="px-3 py-2 text-sm text-gray-700">{userEmail}</div>
              <Link
                to="/me"
                onClick={() => setIsMobileMenuOpen(false)}
                className="flex items-center space-x-3 px-3 py-2 rounded-md text-base font-medium text-gray-700 hover:bg-red-50 hover:text-red-800"
              >
                <UserIcon className="h-5 w-5" />
                <span>Profile</span>
              </Link>
              <button
                onClick={() => {
                  setIsMobileMenuOpen(false);
                  handleLogout();
                }}
                className="w-full flex items-center space-x-3 px-3 py-2 rounded-md text-base font-medium text-gray-700 hover:bg-red-50 hover:text-red-800"
              >
                <ArrowRightOnRectangleIcon className="h-5 w-5" />
                <span>Sign Out</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </nav>
  );
};

export default Navigation;
