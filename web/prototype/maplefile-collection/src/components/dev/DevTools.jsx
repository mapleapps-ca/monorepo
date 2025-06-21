import { useState, useEffect } from "react";
import { useServices } from "../../contexts/ServiceContext";

/**
 * Development tools component for debugging API connection
 * Only shown in development mode
 */
const DevTools = () => {
  const { apiService } = useServices();
  const [apiStatus, setApiStatus] = useState({
    status: "checking",
    message: "Checking API connection...",
  });
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const checkApiStatus = async () => {
      try {
        const health = await apiService.healthCheck();
        if (health.ok) {
          setApiStatus({
            status: "success",
            message: `API connected successfully (${health.status})`,
            details: `Base URL: ${health.baseURL}`,
          });
        } else {
          setApiStatus({
            status: "error",
            message: `API connection failed (${health.status})`,
            details: `Base URL: ${health.baseURL}\nError: ${health.error || "Unknown error"}`,
          });
        }
      } catch (error) {
        setApiStatus({
          status: "error",
          message: "API connection failed",
          details: `Error: ${error.message}`,
        });
      }
    };

    // Only run in development
    if (import.meta.env.DEV) {
      checkApiStatus();
    }
  }, [apiService]);

  // Don't render in production
  if (!import.meta.env.DEV) {
    return null;
  }

  const statusColors = {
    checking: "bg-yellow-100 border-yellow-300 text-yellow-800",
    success: "bg-green-100 border-green-300 text-green-800",
    error: "bg-red-100 border-red-300 text-red-800",
  };

  const statusIcons = {
    checking: (
      <div className="w-4 h-4 border-2 border-yellow-600 border-t-transparent rounded-full animate-spin"></div>
    ),
    success: (
      <svg
        className="w-4 h-4"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    ),
    error: (
      <svg
        className="w-4 h-4"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    ),
  };

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <div
        className={`border rounded-lg p-3 shadow-lg cursor-pointer transition-all ${statusColors[apiStatus.status]}`}
        onClick={() => setIsVisible(!isVisible)}
      >
        <div className="flex items-center space-x-2">
          {statusIcons[apiStatus.status]}
          <span className="text-sm font-medium">
            {isVisible ? "API Status" : "API"}
          </span>
        </div>

        {isVisible && (
          <div className="mt-2 pt-2 border-t border-current/20">
            <p className="text-sm">{apiStatus.message}</p>
            {apiStatus.details && (
              <p className="text-xs mt-1 font-mono whitespace-pre-line">
                {apiStatus.details}
              </p>
            )}
            {apiStatus.status === "error" && (
              <div className="mt-2 text-xs">
                <p className="font-semibold">Troubleshooting:</p>
                <ul className="list-disc list-inside space-y-1">
                  <li>Ensure backend is running on localhost:8000</li>
                  <li>Check that CORS is configured</li>
                  <li>Verify /maplefile/api/v1 endpoint exists</li>
                  <li>Check browser network tab for request details</li>
                </ul>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default DevTools;
