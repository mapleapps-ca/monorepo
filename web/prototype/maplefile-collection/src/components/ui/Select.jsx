/**
 * Reusable Select component
 * Follows Single Responsibility Principle - only handles select rendering
 */
const Select = ({
  label,
  options = [],
  error,
  required = false,
  placeholder = "Select an option",
  className = "",
  ...props
}) => {
  const selectClasses = `
    w-full px-3 py-2 border rounded-lg
    focus:outline-none focus:ring-2 focus:ring-blue-500
    ${error ? "border-red-500" : "border-gray-300"}
    ${className}
  `;

  return (
    <div className="w-full">
      {label && (
        <label className="block text-sm font-medium text-gray-700 mb-1">
          {label}
          {required && <span className="text-red-500 ml-1">*</span>}
        </label>
      )}
      <select className={selectClasses} {...props}>
        <option value="">{placeholder}</option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {error && <p className="mt-1 text-sm text-red-600">{error}</p>}
    </div>
  );
};

export default Select;
