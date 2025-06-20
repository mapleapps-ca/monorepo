/**
 * Error component for displaying error states
 * Follows Single Responsibility Principle - only handles error display
 */
const Error = ({ message, onRetry, className = "" }) => {
  return (
    <div
      className={`flex flex-col items-center justify-center p-6 bg-red-50 border border-red-200 rounded-lg ${className}`}
    >
      <div className="text-red-600 mb-2">
        <svg
          className="w-8 h-8"
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
      </div>
      <h3 className="text-lg font-semibold text-red-800 mb-2">
        Something went wrong
      </h3>
      <p className="text-red-600 text-center mb-4">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition-colors"
        >
          Try Again
        </button>
      )}
    </div>
  );
};

export default Error;
