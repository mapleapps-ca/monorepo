// File: src/components/Navigation.jsx
import React, { useState, useEffect } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import { useAuth } from "../services/Services";
import {
  LockClosedIcon,
  HomeIcon,
  FolderIcon,
  UserIcon,
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  XMarkIcon,
  ChevronDownIcon,
  BellIcon,
  Cog6ToothIcon,
} from "@heroicons/react/24/outline";

const Navigation = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { authManager } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isProfileMenuOpen, setIsProfileMenuOpen] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 10);
    };

    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

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

  const isActive = (path) => {
    if (path === "/file-manager") {
      return location.pathname.startsWith("/file-manager");
    }
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
  const userInitial = userEmail.charAt(0).toUpperCase();

  return (
    <>
      <nav
        className={`sticky top-0 z-50 transition-all duration-300 ${
          isScrolled
            ? "bg-white/95 backdrop-blur-md shadow-lg border-b border-gray-200/50"
            : "bg-white border-b border-gray-200"
        }`}
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            {/* Logo */}
            <Link to="/dashboard" className="flex items-center space-x-3 group">
              <div className="relative">
                <div className="absolute inset-0 bg-gradient-to-r from-red-600 to-red-800 rounded-xl opacity-0 group-hover:opacity-100 blur-xl transition-opacity duration-300"></div>
                <div className="relative flex items-center justify-center h-10 w-10 bg-gradient-to-br from-red-700 to-red-800 rounded-xl shadow-md group-hover:shadow-lg transform group-hover:scale-105 transition-all duration-200">
                  <LockClosedIcon className="h-5 w-5 text-white" />
                </div>
              </div>
              <div>
                <span className="text-xl font-bold text-gray-900 group-hover:text-red-800 transition-colors duration-200">
                  MapleFile
                </span>
                <span className="hidden sm:block text-xs text-gray-500">
                  Secure Storage
                </span>
              </div>
            </Link>

            {/* Desktop Navigation */}
            <div className="hidden lg:flex items-center space-x-1">
              {mainNavItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`relative px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 group ${
                    isActive(item.path)
                      ? "text-white"
                      : "text-gray-700 hover:text-gray-900 hover:bg-gray-100"
                  }`}
                >
                  {isActive(item.path) && (
                    <div className="absolute inset-0 bg-gradient-to-r from-red-700 to-red-800 rounded-lg"></div>
                  )}
                  <div className="relative flex items-center space-x-2">
                    <item.icon
                      className={`h-4 w-4 ${
                        isActive(item.path)
                          ? "text-white"
                          : "text-gray-500 group-hover:text-gray-700"
                      }`}
                    />
                    <span>{item.name}</span>
                  </div>
                  {!isActive(item.path) && (
                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-red-700 transform scale-x-0 group-hover:scale-x-100 transition-transform duration-200"></div>
                  )}
                </Link>
              ))}
            </div>

            {/* User Menu */}
            <div className="hidden lg:flex items-center space-x-4">
              {/* Notifications */}
              <button className="relative p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-all duration-200">
                <BellIcon className="h-5 w-5" />
                <span className="absolute top-1.5 right-1.5 h-2 w-2 bg-red-600 rounded-full animate-pulse"></span>
              </button>

              {/* Profile Dropdown */}
              <div className="relative">
                <button
                  onClick={() => setIsProfileMenuOpen(!isProfileMenuOpen)}
                  className="flex items-center space-x-3 p-2 hover:bg-gray-100 rounded-lg transition-all duration-200"
                >
                  <div className="h-8 w-8 bg-gradient-to-br from-red-600 to-red-800 rounded-lg flex items-center justify-center text-white font-semibold text-sm shadow-sm">
                    {userInitial}
                  </div>
                  <div className="hidden md:block text-left">
                    <p className="text-sm font-medium text-gray-900">
                      {userEmail.split("@")[0]}
                    </p>
                    <p className="text-xs text-gray-500">Free Plan</p>
                  </div>
                  <ChevronDownIcon
                    className={`h-4 w-4 text-gray-500 transition-transform duration-200 ${
                      isProfileMenuOpen ? "rotate-180" : ""
                    }`}
                  />
                </button>

                {/* Dropdown Menu */}
                {isProfileMenuOpen && (
                  <>
                    <div
                      className="fixed inset-0 z-10"
                      onClick={() => setIsProfileMenuOpen(false)}
                    ></div>
                    <div className="absolute right-0 mt-2 w-56 bg-white rounded-xl shadow-xl border border-gray-200 py-2 z-20 animate-fade-in-down">
                      <div className="px-4 py-3 border-b border-gray-100">
                        <p className="text-sm font-medium text-gray-900">
                          {userEmail}
                        </p>
                        <p className="text-xs text-gray-500 mt-0.5">
                          Free Plan • 5GB Storage
                        </p>
                      </div>

                      <Link
                        to="/me"
                        onClick={() => setIsProfileMenuOpen(false)}
                        className="flex items-center px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors duration-200"
                      >
                        <UserIcon className="h-4 w-4 mr-3 text-gray-500" />
                        My Profile
                      </Link>

                      <Link
                        to="/settings"
                        onClick={() => setIsProfileMenuOpen(false)}
                        className="flex items-center px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition-colors duration-200"
                      >
                        <Cog6ToothIcon className="h-4 w-4 mr-3 text-gray-500" />
                        Settings
                      </Link>

                      <div className="border-t border-gray-100 mt-2 pt-2">
                        <button
                          onClick={() => {
                            setIsProfileMenuOpen(false);
                            handleLogout();
                          }}
                          className="w-full flex items-center px-4 py-2.5 text-sm text-red-700 hover:bg-red-50 transition-colors duration-200"
                        >
                          <ArrowRightOnRectangleIcon className="h-4 w-4 mr-3 text-red-600" />
                          Sign Out
                        </button>
                      </div>
                    </div>
                  </>
                )}
              </div>
            </div>

            {/* Mobile Menu Button */}
            <button
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              className="lg:hidden p-2 rounded-lg text-gray-700 hover:text-gray-900 hover:bg-gray-100 transition-all duration-200"
            >
              {isMobileMenuOpen ? (
                <XMarkIcon className="h-6 w-6" />
              ) : (
                <Bars3Icon className="h-6 w-6" />
              )}
            </button>
          </div>
        </div>
      </nav>

      {/* Mobile Menu */}
      <div
        className={`lg:hidden fixed inset-0 z-40 transition-opacity duration-300 ${
          isMobileMenuOpen
            ? "opacity-100 pointer-events-auto"
            : "opacity-0 pointer-events-none"
        }`}
      >
        <div
          className="absolute inset-0 bg-gray-900/50 backdrop-blur-sm"
          onClick={() => setIsMobileMenuOpen(false)}
        ></div>

        <div
          className={`absolute right-0 top-0 h-full w-72 bg-white shadow-2xl transform transition-transform duration-300 ${
            isMobileMenuOpen ? "translate-x-0" : "translate-x-full"
          }`}
        >
          <div className="p-6">
            <div className="flex items-center justify-between mb-8">
              <h2 className="text-lg font-semibold text-gray-900">Menu</h2>
              <button
                onClick={() => setIsMobileMenuOpen(false)}
                className="p-2 rounded-lg text-gray-500 hover:text-gray-700 hover:bg-gray-100 transition-all duration-200"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>

            {/* Mobile User Info */}
            <div className="flex items-center space-x-3 p-4 bg-gray-50 rounded-xl mb-6">
              <div className="h-10 w-10 bg-gradient-to-br from-red-600 to-red-800 rounded-lg flex items-center justify-center text-white font-semibold">
                {userInitial}
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900">
                  {userEmail.split("@")[0]}
                </p>
                <p className="text-xs text-gray-500">{userEmail}</p>
              </div>
            </div>

            {/* Mobile Navigation Items */}
            <div className="space-y-1">
              {mainNavItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className={`flex items-center space-x-3 px-4 py-3 rounded-lg text-base font-medium transition-all duration-200 ${
                    isActive(item.path)
                      ? "bg-gradient-to-r from-red-700 to-red-800 text-white"
                      : "text-gray-700 hover:bg-gray-100 hover:text-gray-900"
                  }`}
                >
                  <item.icon className="h-5 w-5" />
                  <span>{item.name}</span>
                </Link>
              ))}

              <Link
                to="/me"
                onClick={() => setIsMobileMenuOpen(false)}
                className="flex items-center space-x-3 px-4 py-3 rounded-lg text-base font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 transition-all duration-200"
              >
                <UserIcon className="h-5 w-5" />
                <span>Profile</span>
              </Link>
            </div>

            <div className="absolute bottom-0 left-0 right-0 p-6 border-t border-gray-200">
              <button
                onClick={() => {
                  setIsMobileMenuOpen(false);
                  handleLogout();
                }}
                className="w-full flex items-center justify-center space-x-3 px-4 py-3 rounded-lg text-base font-medium text-white bg-gradient-to-r from-red-700 to-red-800 hover:from-red-800 hover:to-red-900 transition-all duration-200"
              >
                <ArrowRightOnRectangleIcon className="h-5 w-5" />
                <span>Sign Out</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default Navigation;
