// File: src/pages/FileManager/FileManagerIndex.jsx
import React from "react";
import { Link, useNavigate } from "react-router";
import Navigation from "../../../components/Navigation";
import {
  FolderIcon,
  PlusIcon,
  CloudArrowUpIcon,
  DocumentIcon,
  MagnifyingGlassIcon,
  TrashIcon,
  ArrowRightIcon,
  SparklesIcon,
  ViewfinderCircleIcon,
  PencilIcon,
  ShareIcon,
  LockClosedIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";

const FileManagerIndex = () => {
  const navigate = useNavigate();

  const prototypePages = [
    {
      category: "Collections",
      icon: FolderIcon,
      color: "blue",
      pages: [
        {
          name: "Collections View",
          description:
            "Main file browser with grid/list views, filtering, and search",
          path: "/file-manager/collections",
          icon: FolderIcon,
        },
        {
          name: "Create Collection",
          description:
            "Create new folders/albums with privacy settings and sharing",
          path: "/file-manager/collections/create",
          icon: PlusIcon,
        },
        {
          name: "Collection Details",
          description:
            "View collection contents, sub-folders, and manage files",
          path: "/file-manager/collections/1",
          icon: ViewfinderCircleIcon,
        },
        {
          name: "Edit Collection",
          description: "Edit collection settings, permissions, and sharing",
          path: "/file-manager/collections/1/edit",
          icon: PencilIcon,
        },
      ],
    },
    {
      category: "Files",
      icon: DocumentIcon,
      color: "green",
      pages: [
        {
          name: "File Upload",
          description: "Drag & drop file upload with encryption options",
          path: "/file-manager/upload",
          icon: CloudArrowUpIcon,
        },
        {
          name: "File Details",
          description: "View file preview, versions, activity, and sharing",
          path: "/file-manager/files/1",
          icon: DocumentIcon,
        },
      ],
    },
    {
      category: "Search & Organization",
      icon: MagnifyingGlassIcon,
      color: "purple",
      pages: [
        {
          name: "Search Results",
          description: "Search across all files with filters and highlighting",
          path: "/file-manager/search?q=Q4",
          icon: MagnifyingGlassIcon,
        },
        {
          name: "Trash",
          description: "View and restore deleted items, auto-delete tracking",
          path: "/file-manager/trash",
          icon: TrashIcon,
        },
      ],
    },
  ];

  const features = [
    "End-to-end encryption with ChaCha20-Poly1305",
    "File and folder organization with collections",
    "Secure file sharing with permissions",
    "Version history and activity tracking",
    "Advanced search with filters",
    "30-day trash with restore capability",
    "Canadian data residency",
    "Zero-knowledge architecture",
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-red-50">
      <Navigation />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="text-center mb-12">
          <div className="flex justify-center mb-6">
            <div className="relative">
              <div className="flex items-center justify-center h-20 w-20 bg-gradient-to-br from-red-800 to-red-900 rounded-2xl shadow-lg">
                <FolderIcon className="h-10 w-10 text-white" />
              </div>
              <div className="absolute -inset-2 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl blur opacity-20 animate-pulse"></div>
            </div>
          </div>

          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            MapleFile Manager Prototype
          </h1>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto mb-2">
            Explore the file management interface with end-to-end encryption
          </p>
          <div className="flex items-center justify-center space-x-2 text-sm text-gray-500">
            <ShieldCheckIcon className="h-4 w-4 text-green-600" />
            <span>
              All pages are UI prototypes without backend functionality
            </span>
          </div>
        </div>

        {/* Quick Start */}
        <div className="bg-white rounded-2xl shadow-lg border border-gray-200 p-8 mb-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center">
            <SparklesIcon className="h-6 w-6 mr-3 text-yellow-500" />
            Quick Start
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <button
              onClick={() => navigate("/file-manager/collections")}
              className="group flex items-center p-4 bg-gradient-to-r from-blue-50 to-blue-100 rounded-xl hover:from-blue-100 hover:to-blue-200 transition-all duration-200"
            >
              <FolderIcon className="h-8 w-8 text-blue-600 mr-4" />
              <div className="text-left">
                <h3 className="font-semibold text-gray-900 group-hover:text-blue-700">
                  Browse Files
                </h3>
                <p className="text-sm text-gray-600">View your collections</p>
              </div>
              <ArrowRightIcon className="h-5 w-5 text-blue-600 ml-auto group-hover:translate-x-1 transition-transform" />
            </button>

            <button
              onClick={() => navigate("/file-manager/upload")}
              className="group flex items-center p-4 bg-gradient-to-r from-green-50 to-green-100 rounded-xl hover:from-green-100 hover:to-green-200 transition-all duration-200"
            >
              <CloudArrowUpIcon className="h-8 w-8 text-green-600 mr-4" />
              <div className="text-left">
                <h3 className="font-semibold text-gray-900 group-hover:text-green-700">
                  Upload Files
                </h3>
                <p className="text-sm text-gray-600">Drag & drop upload</p>
              </div>
              <ArrowRightIcon className="h-5 w-5 text-green-600 ml-auto group-hover:translate-x-1 transition-transform" />
            </button>

            <button
              onClick={() => navigate("/file-manager/collections/create")}
              className="group flex items-center p-4 bg-gradient-to-r from-purple-50 to-purple-100 rounded-xl hover:from-purple-100 hover:to-purple-200 transition-all duration-200"
            >
              <PlusIcon className="h-8 w-8 text-purple-600 mr-4" />
              <div className="text-left">
                <h3 className="font-semibold text-gray-900 group-hover:text-purple-700">
                  New Collection
                </h3>
                <p className="text-sm text-gray-600">Create folder/album</p>
              </div>
              <ArrowRightIcon className="h-5 w-5 text-purple-600 ml-auto group-hover:translate-x-1 transition-transform" />
            </button>
          </div>
        </div>

        {/* All Prototype Pages */}
        <div className="space-y-8">
          {prototypePages.map((category) => (
            <div
              key={category.category}
              className="bg-white rounded-2xl shadow-lg border border-gray-200 p-8"
            >
              <h2 className="text-xl font-bold text-gray-900 mb-6 flex items-center">
                <category.icon
                  className={`h-6 w-6 mr-3 text-${category.color}-600`}
                />
                {category.category}
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {category.pages.map((page) => (
                  <Link
                    key={page.path}
                    to={page.path}
                    className="group flex items-start p-6 bg-gray-50 rounded-xl hover:bg-gray-100 transition-all duration-200"
                  >
                    <div
                      className={`flex items-center justify-center h-12 w-12 bg-${category.color}-100 text-${category.color}-600 rounded-lg mr-4 flex-shrink-0 group-hover:scale-110 transition-transform`}
                    >
                      <page.icon className="h-6 w-6" />
                    </div>
                    <div className="flex-1">
                      <h3 className="font-semibold text-gray-900 mb-1 group-hover:text-red-600">
                        {page.name}
                      </h3>
                      <p className="text-sm text-gray-600">
                        {page.description}
                      </p>
                    </div>
                    <ArrowRightIcon className="h-5 w-5 text-gray-400 group-hover:text-red-600 group-hover:translate-x-1 transition-all" />
                  </Link>
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* Features Section */}
        <div className="mt-12 bg-gradient-to-r from-red-800 to-red-900 rounded-2xl shadow-xl p-8 text-white">
          <h2 className="text-2xl font-bold mb-6 flex items-center">
            <ShieldCheckIcon className="h-6 w-6 mr-3" />
            Security Features in the Prototype
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {features.map((feature, index) => (
              <div key={index} className="flex items-center">
                <LockClosedIcon className="h-4 w-4 mr-3 text-red-200 flex-shrink-0" />
                <span className="text-sm">{feature}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Navigation Tips */}
        <div className="mt-8 bg-blue-50 border border-blue-200 rounded-xl p-6">
          <h3 className="font-semibold text-blue-900 mb-3 flex items-center">
            <SparklesIcon className="h-5 w-5 mr-2" />
            Navigation Tips
          </h3>
          <ul className="space-y-2 text-sm text-blue-800">
            <li>• Click on any collection or file to view its details</li>
            <li>• Use the breadcrumb navigation to go back</li>
            <li>
              • Try different view modes (grid/list) in the collections view
            </li>
            <li>• Search functionality shows mock results for "Q4"</li>
            <li>• All actions show visual feedback but don't persist data</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default FileManagerIndex;
