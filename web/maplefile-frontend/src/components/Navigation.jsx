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
  Cog6ToothIcon,
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  XMarkIcon,
  ShieldCheckIcon,
  CloudArrowUpIcon,
  ClockIcon,
  ChartBarIcon,
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
    // Also clear any session storage
    sessionStorage.clear();
    localStorage.removeItem("mapleapps_access_token");
    localStorage.removeItem("mapleapps_refresh_token");
    localStorage.removeItem("mapleapps_user_email");
    navigate("/");
  };

  const isActive = (path) => {
    return location.pathname === path;
  };

  const mainNavItems = [
    {
      name: "Dashboard",
      path: "/dashboard",
      icon: HomeIcon,
      description: "Overview and recent files",
    },
    {
      name: "My Files",
      path: "/file-manager",
      icon: FolderIcon,
      description: "Organize your files",
    },
    {
      name: "All Files",
      path: "/developer/recent-file-manager-example",
      icon: DocumentIcon,
      description: "Browse all files",
    },
    {
      name: "Upload",
      path: "/developer/create-file-manager-example",
      icon: CloudArrowUpIcon,
      description: "Add new files",
    },
  ];

  const accountItems = [
    {
      name: "Profile",
      path: "/me",
      icon: UserIcon,
      description: "Account settings",
    },
  ];

  // Get current user email
  const userEmail = authManager?.getCurrentUserEmail?.() || "User";

  return (
    <nav className="bg-white/95 backdrop-blur-sm border-b border-gray-100 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-4">
          {/* Logo */}
          <Link to="/dashboard" className="flex items-center group">
            <div className="flex items-center justify-center h-10 w-10 bg-gradient-to-br from-red-800 to-red-900 rounded-lg mr-3 group-hover:scale-105 transition-transform duration-200">
              <LockClosedIcon className="h-6 w-6 text-white" />
            </div>
            <span className="text-2xl font-bold bg-gradient-to-r from-gray-900 to-red-800 bg-clip-text text-transparent">
              MapleFile
            </span>
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden lg:flex items-center space-x-1">
            {mainNavItems.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                className={`group relative px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                  isActive(item.path)
                    ? "bg-gradient-to-r from-red-800 to-red-900 text-white shadow-lg"
                    : "text-gray-700 hover:text-red-800 hover:bg-red-50"
                }`}
              >
                <div className="flex items-center space-x-2">
                  <item.icon className="h-4 w-4" />
                  <span>{item.name}</span>
                </div>

                {/* Tooltip */}
                <div className="absolute top-full mt-2 left-1/2 transform -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none">
                  <div className="bg-gray-900 text-white text-xs rounded-lg py-2 px-3 whitespace-nowrap">
                    {item.description}
                    <div className="absolute bottom-full left-1/2 transform -translate-x-1/2 border-4 border-transparent border-b-gray-900"></div>
                  </div>
                </div>
              </Link>
            ))}
          </div>

          {/* User Menu */}
          <div className="hidden lg:flex items-center space-x-4">
            {/* User Info */}
            <div className="flex items-center space-x-3 px-3 py-2 bg-gray-50 rounded-lg">
              <div className="flex items-center justify-center h-8 w-8 bg-gradient-to-br from-gray-400 to-gray-500 rounded-full">
                <UserIcon className="h-4 w-4 text-white" />
              </div>
              <div className="text-sm">
                <p className="font-medium text-gray-900">{userEmail}</p>
                <div className="flex items-center space-x-1 text-xs text-gray-500">
                  <ShieldCheckIcon className="h-3 w-3 text-green-500" />
                  <span>Encrypted</span>
                </div>
              </div>
            </div>

            {/* Account Menu */}
            <div className="flex items-center space-x-1">
              {accountItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`group relative p-2 rounded-lg transition-all duration-200 ${
                    isActive(item.path)
                      ? "bg-gradient-to-r from-red-800 to-red-900 text-white"
                      : "text-gray-700 hover:text-red-800 hover:bg-red-50"
                  }`}
                  title={item.description}
                >
                  <item.icon className="h-5 w-5" />
                </Link>
              ))}

              {/* Logout Button */}
              <button
                onClick={handleLogout}
                className="group relative p-2 rounded-lg text-gray-700 hover:text-red-800 hover:bg-red-50 transition-all duration-200"
                title="Sign out"
              >
                <ArrowRightOnRectangleIcon className="h-5 w-5" />
              </button>
            </div>
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className="lg:hidden p-2 rounded-lg text-gray-700 hover:text-red-800 hover:bg-red-50 transition-all duration-200"
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
        <div className="lg:hidden bg-white border-t border-gray-100 animate-fade-in">
          <div className="px-4 py-6 space-y-6">
            {/* User Info */}
            <div className="flex items-center space-x-3 p-4 bg-gray-50 rounded-xl">
              <div className="flex items-center justify-center h-12 w-12 bg-gradient-to-br from-gray-400 to-gray-500 rounded-full">
                <UserIcon className="h-6 w-6 text-white" />
              </div>
              <div>
                <p className="font-semibold text-gray-900">{userEmail}</p>
                <div className="flex items-center space-x-1 text-sm text-gray-500">
                  <ShieldCheckIcon className="h-4 w-4 text-green-500" />
                  <span>End-to-end encrypted</span>
                </div>
              </div>
            </div>

            {/* Main Navigation */}
            <div className="space-y-2">
              <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-2">
                Navigation
              </h3>
              {mainNavItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className={`flex items-center space-x-3 px-4 py-3 rounded-xl transition-all duration-200 ${
                    isActive(item.path)
                      ? "bg-gradient-to-r from-red-800 to-red-900 text-white shadow-lg"
                      : "text-gray-700 hover:bg-red-50 hover:text-red-800"
                  }`}
                >
                  <item.icon className="h-5 w-5" />
                  <div>
                    <div className="font-medium">{item.name}</div>
                    <div
                      className={`text-xs ${isActive(item.path) ? "text-red-100" : "text-gray-500"}`}
                    >
                      {item.description}
                    </div>
                  </div>
                </Link>
              ))}
            </div>

            {/* Account Section */}
            <div className="space-y-2 border-t border-gray-100 pt-6">
              <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-2">
                Account
              </h3>
              {accountItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className={`flex items-center space-x-3 px-4 py-3 rounded-xl transition-all duration-200 ${
                    isActive(item.path)
                      ? "bg-gradient-to-r from-red-800 to-red-900 text-white shadow-lg"
                      : "text-gray-700 hover:bg-red-50 hover:text-red-800"
                  }`}
                >
                  <item.icon className="h-5 w-5" />
                  <div>
                    <div className="font-medium">{item.name}</div>
                    <div
                      className={`text-xs ${isActive(item.path) ? "text-red-100" : "text-gray-500"}`}
                    >
                      {item.description}
                    </div>
                  </div>
                </Link>
              ))}

              {/* Logout */}
              <button
                onClick={() => {
                  setIsMobileMenuOpen(false);
                  handleLogout();
                }}
                className="w-full flex items-center space-x-3 px-4 py-3 rounded-xl text-gray-700 hover:bg-red-50 hover:text-red-800 transition-all duration-200"
              >
                <ArrowRightOnRectangleIcon className="h-5 w-5" />
                <div>
                  <div className="font-medium">Sign Out</div>
                  <div className="text-xs text-gray-500">
                    End your secure session
                  </div>
                </div>
              </button>
            </div>

            {/* Security Badge */}
            <div className="bg-gradient-to-r from-green-50 to-blue-50 rounded-xl p-4 border border-green-100">
              <div className="flex items-center justify-center space-x-4 text-xs">
                <div className="flex items-center space-x-1">
                  <LockClosedIcon className="h-3 w-3 text-green-600" />
                  <span className="text-green-800 font-medium">
                    E2E Encrypted
                  </span>
                </div>
                <div className="flex items-center space-x-1">
                  <ShieldCheckIcon className="h-3 w-3 text-blue-600" />
                  <span className="text-blue-800 font-medium">
                    Canadian Hosted
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* CSS for animations */}
      <style jsx>{`
        @keyframes fade-in {
          from {
            opacity: 0;
            transform: translateY(-10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }

        .animate-fade-in {
          animation: fade-in 0.2s ease-out;
        }
      `}</style>
    </nav>
  );
};

export default Navigation;
